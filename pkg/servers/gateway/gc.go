package gateway

import (
	"context"
	"time"
)

func (s *Service) runGC(ctx context.Context) error {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {

		conns, err := s.store.ListAllConnections(ctx)
		if err != nil {
			return err
		}

		for _, conn := range conns {
			// now + ttl < last used at
			if s.clock.Now().Add(conn.TTL).Before(conn.LastUsedAt) {
				err := s.store.DeleteConnection(ctx, conn)
				if err != nil {
					return err
				}
			}
		}

		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
		}
	}
}
