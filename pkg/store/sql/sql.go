package storesql

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	klog "k8s.io/klog/v2"

	"github.com/mjudeikis/portal/pkg/config"
	utilsgorm "github.com/mjudeikis/portal/pkg/util/gorm"
)

type SQL struct{}

// Available DB types
const (
	DatabaseTypePostgres = "postgres"
	DatabaseTypeSqlite   = "sqlite"
)

func connect(ctx context.Context, c *config.Database) (*gorm.DB, *pgxpool.Pool, error) {
	logger := klog.FromContext(ctx)
	for {
		select {
		case <-ctx.Done():
			return nil, nil, fmt.Errorf("sql store startup deadline exceeded")
		default:

			var (
				err       error
				dialector gorm.Dialector
				pgxPool   *pgxpool.Pool
			)

			switch c.Type {
			case DatabaseTypePostgres:
				dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable", c.Host, c.Username, c.Password, c.Name, c.Port)
				dialector = postgres.Open(dsn)

				connConfig, err := pgxpool.ParseConfig(dsn)
				if err != nil {
					return nil, nil, fmt.Errorf("failed to parse postgres config: %w", err)
				}
				connConfig.MaxConnIdleTime = c.MaxConnIdleTime
				connConfig.MaxConnLifetime = c.MaxConnLifeTime
				connConfig.MaxConns = 15

				pgxPool, err = pgxpool.ConnectConfig(ctx, connConfig)
				if err != nil {
					time.Sleep(1 * time.Second)
					logger.V(2).Info("sql store connector can't reach DB, waiting")
					continue
				}

			default:
				dialector = sqlite.Open(c.SqliteURI)
			}
			glogs := utilsgorm.NewLogger(logger)
			db, err := gorm.Open(dialector, &gorm.Config{
				Logger: glogs,
			})
			if err != nil {
				time.Sleep(1 * time.Second)
				klog.V(2).Infof("sql store connector can't reach DB, waiting: %s", err)
				continue
			}

			// TODO: Here we should set db overrides for pool, max connections, etc

			// success
			return db, pgxPool, nil

		}
	}
}
