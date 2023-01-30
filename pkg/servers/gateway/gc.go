package gateway

import (
	"context"
	"time"
)

var gcInterval = time.Minute

func (s *Service) runGC(ctx context.Context) error {
	ticker := time.NewTicker(gcInterval)
	defer ticker.Stop()

	for {

		conns, err := s.store.ListAllConnections(ctx)
		if err != nil {
			return err
		}

		for _, conn := range conns {
			if conn.LastUsedAt.Add(conn.TTL).Before(s.clock.Now()) {
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
