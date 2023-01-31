package gateway

import (
	"context"
	"testing"
	"time"

	"github.com/faroshq/faros-ingress/pkg/models"
	"github.com/faroshq/faros-ingress/pkg/store"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	clock "k8s.io/utils/clock/testing"
)

func TestRunGC(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	gcInterval = time.Millisecond * 100 // GC runs at start and then every 100ms

	conn1 := models.Connection{
		ID:         "conn1",
		LastUsedAt: time.Now().Add(-time.Hour * 2),
		TTL:        time.Hour,
	}

	conn2 := models.Connection{
		ID:         "conn2",
		LastUsedAt: time.Now().Add(-time.Hour * 2),
		TTL:        time.Hour * 24,
	}

	t.Run("should delete expired connections", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond) // stop test after 1ms
		defer cancel()
		store := store.NewMockStore(ctrl)
		clock := clock.NewFakeClock(time.Now())

		s := &Service{
			store: store,
			clock: clock,
		}

		store.EXPECT().ListAllConnections(gomock.Any()).Return([]models.Connection{conn1, conn2}, nil)
		store.EXPECT().DeleteConnection(gomock.Any(), conn1).Return(nil)

		err := s.runGC(ctx)
		assert.NoError(t, err)

	})

	t.Run("should not delete unexpired connections", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond) // stop test after 1ms
		defer cancel()
		store := store.NewMockStore(ctrl)
		clock := clock.NewFakeClock(time.Now())

		s := &Service{
			store: store,
			clock: clock,
		}

		conn1.TTL = 24 * time.Hour
		conn2.TTL = 24 * time.Hour

		store.EXPECT().ListAllConnections(gomock.Any()).Return([]models.Connection{conn1, conn2}, nil)

		err := s.runGC(ctx)
		assert.NoError(t, err)

	})

}
