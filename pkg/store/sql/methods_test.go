package storesql_test

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/mjudeikis/portal/pkg/models"
	databasetest "github.com/mjudeikis/portal/test/util/database"
)

// TestCascade tests if records deletes cascades
func TestCascade(t *testing.T) {
	if os.Getenv("CI_ONLY") == "" {
		t.Skip("skipping postgres tests in non-CI environment")
		return
	}

	db, err := databasetest.NewPostgresTestingStore(t)
	require.NoError(t, err)

	ctx := context.Background()

	user, err := db.CreateUser(ctx, models.User{
		Email: "foo@foo.lt",
	})
	require.NoError(t, err)

	agent, err := db.CreateConnection(ctx, models.Connection{
		UserID: user.ID,
	})
	require.NoError(t, err)

	err = db.DeleteUser(ctx, *user)
	require.NoError(t, err)

	_, err = db.GetConnection(ctx, models.Connection{
		ID: agent.ID,
	})
	require.Error(t, err)
}
