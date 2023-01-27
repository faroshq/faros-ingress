package database

import (
	"context"
	"os"
	"testing"

	"github.com/mjudeikis/portal/pkg/config"
	"github.com/mjudeikis/portal/pkg/store"
	storesql "github.com/mjudeikis/portal/pkg/store/sql"
	"k8s.io/klog/v2"
)

// NewSQLLiteTestingStore creates a new sqllite database
func NewSQLLiteTestingStore(t *testing.T) (store.Store, error) {
	ctx := klog.NewContext(context.Background(), klog.NewKlogr())
	t.Log("using sqllite database")

	// Setting defaults if nothing is set so it works
	// with local postgres created by docker-compose
	if os.Getenv("FAROS_DATABASE_TYPE") == "" {
		os.Setenv("FAROS_DATABASE_TYPE", "sqllite")
		os.Setenv("FAROS_DATABASE_SQLITE_URI", "file::memory:?cache=shared")
	}

	var store store.Store
	var err error

	cfg, err := config.LoadAPI()
	if err != nil {
		return nil, err
	}

	// connecting to test DB
	store, err = storesql.NewStore(ctx, &cfg.Database)
	if err != nil {
		return nil, err
	}

	return store, nil
}
