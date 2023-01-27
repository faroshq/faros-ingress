package storesql

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	klog "k8s.io/klog/v2"

	"github.com/faroshq/faros-ingress/pkg/config"
	"github.com/faroshq/faros-ingress/pkg/store"
)

var _ store.Store = &Store{}

type Store struct {
	db      *gorm.DB
	pgxPool *pgxpool.Pool // used for pubsub if we need one
}

func NewStore(ctx context.Context, c *config.Database) (*Store, error) {
	logger := klog.FromContext(ctx)
	logger = logger.WithValues("database", c.Type)

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Second)
	defer cancel()

	db, pgxPool, err := connect(ctx, c)
	if err != nil {
		return nil, err
	}

	logger.WithValues("dialector", db.Dialector.Name()).Info("Initializing database store")

	if db.Dialector.Name() == sqlite.DriverName {
		err = db.Exec("PRAGMA foreign_keys = ON").Error
		if err != nil {
			return nil, err
		}
	}

	s := &Store{
		db:      db,
		pgxPool: pgxPool,
	}

	err = s.migrate(ctx, c)
	if err != nil {
		return nil, fmt.Errorf("database migration failed: %w", err)
	}

	return s, nil
}

func (s *Store) Status() (interface{}, error) {
	if s.db == nil {
		return nil, fmt.Errorf("database is not initialized")
	}
	db, err := s.db.DB()
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (s *Store) Close() error {
	if s.pgxPool != nil {
		s.pgxPool.Close()
	}
	db, err := s.db.DB()
	if err != nil {
		return nil
	}
	return db.Close()
}

func (s *Store) RawDB() interface{} {
	return s.db
}
