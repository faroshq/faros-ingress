package storesql

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/mjudeikis/portal/pkg/config"
	"github.com/mjudeikis/portal/pkg/models"
	"gorm.io/gorm"
	klog "k8s.io/klog/v2"
)

func (s *Store) migrate(ctx context.Context, c *config.Database) error {
	logger := klog.FromContext(ctx)
	err := s.db.AutoMigrate(
		&models.User{},
		&models.Connection{},
	)
	if err != nil {
		return err
	}

	// When using sqlite we emulate pub sub with a table. This is only for very
	// low traffic use cases. For high traffic use cases we recommend using
	// postgres.
	if s.db.Dialector.Name() == DatabaseTypeSqlite {
		logger.Info("Creating pubsub table")
		err := s.db.AutoMigrate(
			&models.Event{},
		)
		if err != nil {
			return err
		}
	}

	if c.Type == DatabaseTypePostgres {
		//if err := createFK(s.db, models.Workspace{}, models.User{}, "user_id", "id", "CASCADE", "CASCADE"); err != nil {
		//	logger.Info("failed to add DB FK: %s", err)
		//}
	}
	return nil
}

func createFK(db *gorm.DB, src, dst interface{}, fk, pk string, onDelete, onUpdate string) error {
	srcTableName := db.NamingStrategy.TableName(reflect.TypeOf(src).Name())
	dstTableName := db.NamingStrategy.TableName(reflect.TypeOf(dst).Name())

	constraintName := "fk_" + srcTableName + "_" + dstTableName

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	if !db.Migrator().HasConstraint(src, constraintName) {
		err := db.WithContext(ctx).Exec(fmt.Sprintf("ALTER TABLE %s ADD CONSTRAINT %s FOREIGN KEY (%s) REFERENCES %s(%s) ON DELETE %s ON UPDATE %s",
			srcTableName,
			constraintName,
			fk,
			dstTableName,
			pk,
			onDelete,
			onUpdate)).Error
		if err != nil && !strings.Contains(err.Error(), "already exists") {
			return err
		}
	}
	return nil
}
