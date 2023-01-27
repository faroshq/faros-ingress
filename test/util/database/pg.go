package database

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"gorm.io/gorm"
	"k8s.io/klog/v2"

	"github.com/mjudeikis/portal/pkg/config"
	"github.com/mjudeikis/portal/pkg/store"
	storesql "github.com/mjudeikis/portal/pkg/store/sql"
)

// NewPostgresTestingStore creates a new, clean test database for the current
// test and drops it on test cleanup.
func NewPostgresTestingStore(t *testing.T) (store.Store, error) {
	ctx := klog.NewContext(context.Background(), klog.NewKlogr())
	t.Log("using postgres database")

	// Setting defaults if nothing is set so it works
	// with local postgres created by docker-compose
	if os.Getenv("FAROS_DATABASE_TYPE") == "" {
		os.Setenv("FAROS_DATABASE_TYPE", "postgres")
		os.Setenv("FAROS_DATABASE_HOST", "localhost")
		os.Setenv("FAROS_DATABASE_PASSWORD", "pgpass")
		os.Setenv("FAROS_DATABASE_USERNAME", "pguser")
	}

	var store store.Store
	var err error

	cfg, err := config.LoadAPI()
	if err != nil {
		return nil, err
	}

	testDatabaseName := fmt.Sprintf("test%d", time.Now().UnixNano())

	t.Cleanup(func() {
		store.Close()

		cfg.Database.Name = "postgres"

		store, err := storesql.NewStore(ctx, &cfg.Database)
		if err != nil {
			t.Log("failed to connect to postgres database, is it running?")
		}
		defer store.Close()

		db := store.RawDB().(*gorm.DB)
		err = db.Exec(fmt.Sprintf("drop database %s", testDatabaseName)).Error
		if err != nil {
			t.Log("failed to drop test database")
		}

	})

	err = createTestDatabase(t, cfg, testDatabaseName)
	if err != nil {
		return nil, err
	}

	// connecting to test DB
	cfg.Database.Name = testDatabaseName
	store, err = storesql.NewStore(ctx, &cfg.Database)
	if err != nil {
		return nil, err
	}

	return store, nil
}

func createTestDatabase(t *testing.T, cfg *config.APIConfig, testDatabaseName string) error {
	ctx := klog.NewContext(context.Background(), klog.NewKlogr())
	cfg.Database.Name = "postgres"

	store, err := storesql.NewStore(ctx, &cfg.Database)
	if err != nil {
		t.Log("failed to connect to postgres database, is it running?")
	}
	defer store.Close()

	db := store.RawDB().(*gorm.DB)
	err = db.Exec(fmt.Sprintf("create database %s", testDatabaseName)).Error
	if err != nil {
		if !strings.Contains(err.Error(), "already exists") {
			return err
		}
	}
	return nil
}
