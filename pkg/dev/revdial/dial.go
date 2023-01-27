package revdial

// Based on https://github.com/aojea/h2rev2

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"k8s.io/klog/v2"
)

// The Dialer can create new connections back to the origin.
// A Dialer can have multiple clients.
type Dialer struct {
	id           string
	conn         net.Conn      // control plane connection
	incomingConn chan net.Conn // data plane connections
	connReady    chan bool
	pickupFailed chan error
	donec        chan struct{}
	closeOnce    sync.Once
	revClient    *http.Client
}

// NewDialer returns the side of the connection which will initiate
// new connections over the already established reverse connections.
func NewDialer(id string, conn net.Conn) *Dialer {
	d := &Dialer{
		id:           id,
		conn:         conn,
		donec:        make(chan struct{}),
		connReady:    make(chan bool),
		pickupFailed: make(chan error),
		incomingConn: make(chan net.Conn),
	}
	go d.serve()
	return d
}

// serve blocks and runs the control message loop, keeping the peer
// alive and notifying the peer when new connections are available.
func (d *Dialer) serve() error {
	defer d.Close()
	go func() {
		defer d.Close()
		br := bufio.NewReader(d.conn)
		for {
			line, err := br.ReadSlice('\n')
			if err != nil {
				return
			}
			select {
			case <-d.donec:
				return
			default:
			}
			var msg controlMsg
			if err := json.Unmarshal(line, &msg); err != nil {
				log.Printf("revdial.Dialer read invalid JSON: %q: %v", line, err)
				return
			}
			switch msg.Command {
			case "pickup-failed":
				err := fmt.Errorf("revdial listener failed to pick up connection: %v", msg.Err)
				select {
				case d.pickupFailed <- err:
				case <-d.donec:
					return
				}
			}
		}
	}()
	for {
		select {
		case <-d.connReady:
			if err := d.sendMessage(controlMsg{
				Command:  "conn-ready",
				ConnPath: "",
			}); err != nil {
				return err
			}
		case <-d.donec:
			return errors.New("revdial.Dialer closed")
		}
	}
}

func (d *Dialer) sendMessage(m controlMsg) error {
	j, _ := json.Marshal(m)
	d.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	j = append(j, '\n')
	_, err := d.conn.Write(j)
	d.conn.SetWriteDeadline(time.Time{})
	return err
}

// Done returns a channel which is closed when d is closed (either by
// this process on purpose, by a local error, or close or error from
// the peer).
func (d *Dialer) Done() <-chan struct{} { return d.donec }

// Close closes the Dialer.
func (d *Dialer) Close() error {
	d.closeOnce.Do(d.close)
	return nil
}

func (d *Dialer) close() {
	d.conn.Close()
	close(d.donec)
}

// reverseClient caches the reverse http client
func (d *Dialer) reverseClient() *http.Client {
	if d.revClient == nil {
		// create the http.client for the reverse connections
		tr := &http.Transport{
			Proxy:               nil,    // no proxies
			DialContext:         d.Dial, // use a reverse connection
			ForceAttemptHTTP2:   false,  // this is a tunneled connection
			DisableKeepAlives:   true,   // one connection per reverse connection
			MaxIdleConnsPerHost: -1,
		}

		client := http.Client{
			Transport: tr,
		}
		d.revClient = &client
	}
	return d.revClient

}

// Dial creates a new connection back to the Listener.
func (d *Dialer) Dial(ctx context.Context, network string, address string) (net.Conn, error) {
	now := time.Now()
	defer klog.V(5).Infof("dial to %s took %v", address, time.Since(now))
	// First, tell serve that we want a connection:
	select {
	case d.connReady <- true:
	case <-d.donec:
		return nil, errors.New("revdial.Dialer closed")
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	// Then pick it up:
	select {
	case c := <-d.incomingConn:
		return c, nil
	case err := <-d.pickupFailed:
		return nil, err
	case <-d.donec:
		return nil, errors.New("revdial.Dialer closed")
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}
