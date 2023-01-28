package h2rev2

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/faroshq/faros-ingress/pkg/models"
	"github.com/faroshq/faros-ingress/pkg/store"
	"k8s.io/klog/v2"
)

type controlMsg struct {
	Command  string `json:"command,omitempty"`  // "keep-alive", "conn-ready", "pickup-failed"
	ConnPath string `json:"connPath,omitempty"` // conn pick-up URL path for "conn-url", "pickup-failed"
	Err      string `json:"err,omitempty"`
}

// ReversePool contains a pool of Dialers to create reverse connections
// It exposes an http.Handler to handle the clients.
//
//	pool := h2rev2.NewReversePool()
//	mux := http.NewServeMux()
//	mux.Handle("", pool)
type ReversePool struct {
	mu    sync.Mutex
	store store.Store
	// pool of dialer
	pool map[string]*Dialer
	// authenticated is list of authenticated agents
	authenticated map[string]string
}

// NewReversePool returns a ReversePool
func NewReversePool(store store.Store) *ReversePool {
	return &ReversePool{
		pool:          map[string]*Dialer{},
		authenticated: map[string]string{}, // objectID -> identity for tunnel
		store:         store,
	}
}

func (rp *ReversePool) Run(ctx context.Context) error {
	// initial build of the connection pool
	conns, err := rp.store.ListAllConnections(ctx)
	if err != nil {
		return err
	}
	rp.mu.Lock()
	for _, conn := range conns {
		rp.authenticated[conn.ID] = conn.Token
	}
	rp.mu.Unlock()

	changesCh := make(chan *models.Event)

	go func() {
		klog.V(2).Info("Subscribing to changes")
		defer klog.V(2).Info("Unsubscribing from changes")
		for {
			select {
			case <-ctx.Done():
				return
			default:
				err := rp.store.SubscribeChanges(ctx, func(event *models.Event) error {
					changesCh <- event
					return nil
				})
				if err != nil {
					klog.Error(err, "failed to subscribe to changes")
				}
				// Retry to subscribe
				time.Sleep(time.Second)
			}
		}
	}()

	// Start periodically reschedule applications on all devices or individual
	// ones if only minimal changes are required
	for {
		select {
		case <-ctx.Done():
			return nil
		case event := <-changesCh:
			switch event.Resource {
			case models.EventResourceConnection:
				switch event.Type {
				case models.EventCreated:

					klog.V(2).Info("agent connection created")
					agent, err := rp.store.GetConnection(ctx, models.Connection{ID: event.ObjectID})
					if err != nil {
						klog.Error(err, "failed to get connection")
						continue
					}
					rp.mu.Lock()
					rp.authenticated[agent.ID] = agent.Token
					rp.mu.Unlock()

				case models.EventDeleted:
					klog.V(2).Info("connection delete")
					rp.mu.Lock()
					delete(rp.authenticated, event.ObjectID)
					rp.mu.Unlock()
				case models.EventUpdated:
					klog.V(2).Info("connection update")
					conn, err := rp.store.GetConnection(ctx, models.Connection{ID: event.ObjectID})
					if err != nil {
						klog.Error(err, "failed to get connection")
						continue
					}
					rp.mu.Lock()
					rp.authenticated[conn.ID] = conn.Token
					rp.mu.Unlock()
				}
			}
		}
	}
}

// Close the Reverse pool and all its dialers
func (rp *ReversePool) Close() {
	rp.mu.Lock()
	defer rp.mu.Unlock()
	for _, v := range rp.pool {
		v.Close()
	}
}

// GetDialer returns a reverse dialer for the id
func (rp *ReversePool) GetDialer(id string) *Dialer {
	rp.mu.Lock()
	defer rp.mu.Unlock()
	return rp.pool[id]
}

// CreateDialer creates a reverse dialer with id
// it's a noop if a dialer already exists
func (rp *ReversePool) CreateDialer(id string, conn net.Conn) *Dialer {
	rp.mu.Lock()
	defer rp.mu.Unlock()
	if d, ok := rp.pool[id]; ok {
		return d
	}
	d := NewDialer(id, conn)
	rp.pool[id] = d
	return d

}

// DeleteDialer delete the reverse dialer for the id
func (rp *ReversePool) DeleteDialer(id string) {
	rp.mu.Lock()
	defer rp.mu.Unlock()
	delete(rp.pool, id)
}

// HTTP Handler that handles reverse connections and reverse proxy requests using 2 different paths:
// path base/revdial?key=id establish reverse connections and queue them so it can be consumed by the dialer
// path base/proxy/id/(path) proxies the (path) through the reverse connection identified by id
func (rp *ReversePool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// recover panic
	defer func() {
		if r := recover(); r != nil {
			var err error
			switch t := r.(type) {
			case string:
				err = errors.New(t)
			case error:
				err = t
			default:
				err = errors.New("unknown error")
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}()

	// process path
	path := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(path) == 0 {
		http.Error(w, "", http.StatusNotFound)
		return
	}

	// route the request
	pos := -1
	for i := len(path) - 1; i >= 0; i-- {
		p := path[i]
		// pathRevDial comes with a param
		if p == pathRevDial {
			if i != len(path)-1 {
				http.Error(w, "revdial: only last element on path allowed", http.StatusInternalServerError)
				return
			}
			pos = i
			break
		}
		// pathRevProxy requires at least the id subpath
		if p == pathRevProxy {
			if i == len(path)-1 {
				http.Error(w, "proxy: reverse path id required", http.StatusInternalServerError)
				return
			}
			pos = i
			break
		}
	}
	if pos < 0 {
		http.Error(w, "revdial: not handler ", http.StatusNotFound)
		return
	}
	// Forward proxy /base/proxy/id/..proxied path...
	if path[pos] == pathRevProxy {

		// authenticate the request
		if !rp.isAuthenticated(path[pos+1]) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		id := path[pos+1]
		target, err := url.Parse("http://" + id)
		if err != nil {
			http.Error(w, "wrong url", http.StatusInternalServerError)
			return
		}
		d := rp.GetDialer(id)
		if d == nil {
			http.Error(w, "not reverse connections for this id available", http.StatusInternalServerError)
			return
		}
		transport := d.reverseClient().Transport
		proxy := httputil.NewSingleHostReverseProxy(target)
		originalDirector := proxy.Director
		proxy.Transport = transport
		proxy.Director = func(req *http.Request) {
			req.Host = target.Host
			req.URL.Path = strings.Join(path[pos+2:], "/")
			originalDirector(req)
		}
		proxy.FlushInterval = -1

		proxy.ServeHTTP(w, r)
		klog.V(5).Infof("proxy server closed %v ", err)
	} else {
		// The caller identify itself by the value of the keu
		// https://server/revdial?id=dialerUniq
		dialerUniq := r.URL.Query().Get(urlParamKey)
		if len(dialerUniq) == 0 {
			http.Error(w, "only reverse connections with id supported", http.StatusInternalServerError)
			return
		}

		d := rp.GetDialer(dialerUniq)
		// First flush response headers
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
		// first connection to register the dialer and start the control loop
		if d == nil || isClosedChan(d.Done()) {
			conn := newConn(r.Body, flushWriter{w})
			rp.DeleteDialer(dialerUniq)
			d = rp.CreateDialer(dialerUniq, conn)
			// start control loop
			<-conn.Done()
			klog.V(5).Infof("stoped dialer %s control connection ", dialerUniq)
			return

		}
		// create a reverse connection
		klog.V(5).Infof("created reverse connection to %s %s id %s", r.RequestURI, r.RemoteAddr, dialerUniq)
		conn := newConn(r.Body, flushWriter{w})
		select {
		case d.incomingConn <- conn:
		case <-d.Done():
			http.Error(w, "Reverse dialer closed", http.StatusInternalServerError)
			return
		}
		// keep the handler alive until the connection is closed
		<-conn.Done()
		klog.V(4).Infof("Connection from %s done", r.RemoteAddr)
	}
}

// isAuthenticated checks if the request is authenticated
func (rp *ReversePool) isAuthenticated(dialerUniq string) bool {
	rp.mu.Lock()
	defer rp.mu.Unlock()
	for _, identity := range rp.authenticated {
		if identity == dialerUniq {
			return true
		}
	}

	return false
}

type flushWriter struct {
	w io.Writer
}

func (fw flushWriter) Write(p []byte) (n int, err error) {
	n, err = fw.w.Write(p)
	if f, ok := fw.w.(http.Flusher); ok {
		f.Flush()
	}
	return
}

func (fw flushWriter) Close() error {
	return nil
}
