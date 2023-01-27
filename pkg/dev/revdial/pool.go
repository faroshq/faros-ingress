package revdial

// Based on https://github.com/aojea/h2rev2

import (
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync"

	"k8s.io/klog/v2"
)

const (
	apiPrefix    = "/api/v1alpha1/proxy"
	pathRevDial  = "revdial"
	pathRevProxy = "proxy"
	urlParamKey  = "id"
	identity     = "faros-dev"
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
	mu   sync.Mutex
	pool map[string]*Dialer
}

// NewReversePool returns a ReversePool
func NewReversePool() *ReversePool {
	return &ReversePool{
		pool: map[string]*Dialer{},
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
	path := strings.Replace(r.URL.Path, apiPrefix, "", 1)
	pathParts := strings.Split(strings.Trim(path, "/"), "/")
	if len(pathParts) == 0 {
		http.Error(w, "", http.StatusNotFound)
		return
	}

	// Forward proxy /base/proxy/id/..proxied path...
	if pathParts[0] == pathRevDial {
		d := rp.GetDialer(identity)
		// First flush response headers
		w.WriteHeader(http.StatusOK)
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
		// first connection to register the dialer and start the control loop
		if d == nil || isClosedChan(d.Done()) {
			conn := newConn(r.Body, flushWriter{w})
			rp.DeleteDialer(identity)
			d = rp.CreateDialer(identity, conn)
			// start control loop
			<-conn.Done()
			klog.V(5).Infof("stopped dialer %s control connection ", identity)
			return
		}
		// create a reverse connection
		klog.V(2).Infof("created reverse connection to %s %s id %s", r.RequestURI, r.RemoteAddr, identity)
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
	} else {

		target, err := url.Parse("http://" + identity)
		if err != nil {
			http.Error(w, "wrong url", http.StatusInternalServerError)
			return
		}

		d := rp.GetDialer(identity)
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
			originalDirector(req)
		}
		proxy.FlushInterval = -1
		proxy.ServeHTTP(w, r)
		klog.V(5).Infof("proxy server closed %v ", err)
	}

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
