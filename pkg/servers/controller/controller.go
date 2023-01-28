package controller

import (
	"context"
	"time"

	"k8s.io/utils/clock"

	"github.com/faroshq/faros-ingress/pkg/config"
	"github.com/faroshq/faros-ingress/pkg/store"
	storesql "github.com/faroshq/faros-ingress/pkg/store/sql"
)

var _ Interface = &Service{}

type Interface interface {
	Run(ctx context.Context) error
}
type Service struct {
	config *config.Config
	store  store.Store
	clock  clock.Clock
}

func New(ctx context.Context, config *config.Config) (*Service, error) {
	store, err := storesql.NewStore(ctx, &config.Database)
	if err != nil {
		return nil, err
	}

	return &Service{
		config: config,
		store:  store,
	}, nil
}

func (s *Service) Run(ctx context.Context) error {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {

		conns, err := s.store.ListAllConnections(ctx)
		if err != nil {
			return err
		}

		for _, conn := range conns {
			// lastUsed + ttl < now
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
