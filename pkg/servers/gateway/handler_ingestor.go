package gateway

import (
	"context"
	"net/http"

	"github.com/faroshq/faros-ingress/pkg/models"
	"k8s.io/klog/v2"
)

func (s *Service) serveIngestor(w http.ResponseWriter, r *http.Request) {
	defer func() {
		host := r.Host

		// Incase we are behind a proxy, we need to use the X-Forwarded-Host header
		if r.Header.Get("X-Forwarded-Host") != "" {
			host = r.Header.Get("X-Forwarded-Host")
		}

		conn, err := s.authenticator.getConnection(host)
		if err != nil || conn == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// If connection was found, bump last used time for housekeeping
		// We gonna clean up connections that are not used for a while
		go func(conn *models.Connection) {
			conn.LastUsedAt = s.clock.Now()
			_, err := s.store.UpdateConnection(context.Background(), *conn)
			if err != nil {
				klog.Errorf("failed to update connection: %s", err)
			}
		}(conn)

		var authenticated bool
		if conn.Secure {
			username, password, ok := r.BasicAuth()
			if !ok {
				w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			authenticated, conn, err = s.authenticator.authenticate(host, username, password)
			if err != nil {
				w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			if !authenticated {
				w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
		}

		// Set conn into context
		*r = *r.WithContext(context.WithValue(r.Context(), contextKeyConnection, conn))

		s.reverseProxy.ServeHTTP(w, r)
	}()
}
