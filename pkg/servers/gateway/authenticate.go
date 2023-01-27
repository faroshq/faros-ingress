package gateway

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/faroshq/faros-ingress/pkg/models"
	"github.com/faroshq/faros-ingress/pkg/store"
	utilpassword "github.com/faroshq/faros-ingress/pkg/util/password"
	"k8s.io/klog/v2"
)

type auth struct {
	store              store.Store
	mu                 sync.Mutex
	existingConnection map[string]models.Connection // hostname -> agent
}

func newAuthenticator(store store.Store) *auth {
	return &auth{
		store:              store,
		existingConnection: map[string]models.Connection{},
	}
}

func (a *auth) run(ctx context.Context) error {
	// initial build of the agent pool
	conns, err := a.store.ListAllConnections(ctx)
	if err != nil {
		return err
	}
	a.mu.Lock()
	for _, conn := range conns {
		hostname := strings.TrimPrefix(conn.Hostname, "https://")
		a.existingConnection[hostname] = conn
	}
	a.mu.Unlock()

	changesCh := make(chan *models.Event)

	go func() {
		klog.V(2).Info("Subscribing to changes")
		defer klog.V(2).Info("Unsubscribing from changes")
		for {
			select {
			case <-ctx.Done():
				return
			default:
				err := a.store.SubscribeChanges(ctx, func(event *models.Event) error {
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

					klog.V(2).Info("connection create")
					conn, err := a.store.GetConnection(ctx, models.Connection{ID: event.ObjectID})
					if err != nil {
						klog.Error(err, "failed to get connection")
						continue
					}
					hostname := strings.TrimPrefix(conn.Hostname, "https://")
					a.mu.Lock()
					a.existingConnection[hostname] = *conn
					a.mu.Unlock()

				case models.EventDeleted:
					klog.V(2).Info("connection delete")
					a.mu.Lock()
					delete(a.existingConnection, event.ObjectID)
					a.mu.Unlock()
				case models.EventUpdated:
					klog.V(2).Info("connection update")
					conn, err := a.store.GetConnection(ctx, models.Connection{ID: event.ObjectID})
					if err != nil {
						klog.Error(err, "failed to get connection")
						continue
					}
					hostname := strings.TrimPrefix(conn.Hostname, "https://")
					a.mu.Lock()
					a.existingConnection[hostname] = *conn
					a.mu.Unlock()
				}
			}
		}
	}
}

func (a *auth) authenticate(hostname string, username, password string) (bool, *models.Connection, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	connection, ok := a.existingConnection[hostname]
	if !ok {
		return false, nil, nil
	}

	err := utilpassword.ComparePasswordHash([]byte(username+":"+password), connection.BasicAuthHash)
	if err != nil {
		return false, nil, err
	}

	return true, &connection, nil
}

func (a *auth) getConnection(hostname string) (*models.Connection, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	connection, ok := a.existingConnection[hostname]
	if !ok {
		return nil, fmt.Errorf("unauthenticated connection")
	}

	return &connection, nil
}
