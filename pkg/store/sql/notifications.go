package storesql

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/faroshq/faros-ingress/pkg/models"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	klog "k8s.io/klog/v2"
)

const (
	// channelUpdates is the channel name for general updates
	channelUpdates = "faros_updates"
)

// SubscribeChanges subscribes to changes in the databaseSubscribeChanges
func (s *Store) SubscribeChanges(ctx context.Context, callback func(event *models.Event) error) error {
	if s.db.Dialector.Name() == DatabaseTypeSqlite {
		return s.subscribeChangesSQLite(ctx, callback)
	}
	return s.subscribeChangesPostgres(ctx, callback)
}

func (s *Store) subscribeChangesPostgres(ctx context.Context, callback func(event *models.Event) error) error {
	logger := klog.FromContext(ctx)
	if s.pgxPool == nil {
		return fmt.Errorf("pgx pool is nil")
	}

	_, err := s.pgxPool.Exec(ctx, "LISTEN "+channelUpdates)
	if err != nil {
		return fmt.Errorf("failed to start listening: %w", err)
	}

	conn, err := s.pgxPool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("failed to acquire conn: %w", err)
	}
	defer conn.Release()

	logger.Info("started listening for application updates")
	defer logger.Info("stopped listening for application updates")

	for {
		notification, err := conn.Conn().WaitForNotification(ctx)
		if err != nil {
			return err
		}

		var event models.Event
		if err := json.Unmarshal([]byte(notification.Payload), &event); err != nil {
			return err
		}

		err = callback(&event)
		if err != nil {
			logger.Error(err, "callback failed on event")
		}
	}
}

func (s *Store) subscribeChangesSQLite(ctx context.Context, callback func(event *models.Event) error) error {
	for {
		var event models.Event
		err := s.db.WithContext(ctx).First(&event).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				time.Sleep(time.Second)
				continue
			}
			return err
		}

		err = callback(&event)
		if err != nil {
			klog.FromContext(ctx).Error(err, "callback failed on event")
		}

		err = s.db.WithContext(ctx).Delete(&event).Error
		if err != nil {
			return err
		}
	}
}

func (s *Store) notifyUpdatedConnection(ctx context.Context, id string, event models.EventType) {
	s._notify(ctx, id, models.EventResourceConnection, event)
}

func (s *Store) notifyUpdatedUser(ctx context.Context, id string, event models.EventType) {
	s._notify(ctx, id, models.EventResourceUser, event)
}

func (s *Store) _notify(ctx context.Context, id string, resource models.EventResource, event models.EventType) {
	if s.db.Dialector.Name() == DatabaseTypeSqlite {
		s._notifySQLlite(ctx, id, resource, event)
	} else {
		s._notifyPostgres(ctx, id, resource, event)
	}
}

func (s *Store) _notifySQLlite(ctx context.Context, id string, resource models.EventResource, event models.EventType) {
	logger := klog.FromContext(ctx)

	e := &models.Event{
		ID:       uuid.New().String(),
		Type:     event,
		Resource: resource,
		ObjectID: id,
	}

	err := s.db.WithContext(ctx).Create(&e).Error
	if err != nil {
		logger.Error(err, "failed to create event")
	}
}

func (s *Store) _notifyPostgres(ctx context.Context, id string, resource models.EventResource, event models.EventType) {
	logger := klog.FromContext(ctx)
	if s.pgxPool == nil {
		// Nothing to do, not initialized (sqlite)
		return
	}

	bts, err := json.Marshal(&models.Event{
		Type:     event,
		Resource: resource,
		ObjectID: id,
	})
	if err != nil {
		logger.Error(err, "failed to marshal membership notification")
		return
	}

	_, err = s.pgxPool.Exec(ctx, fmt.Sprintf("select pg_notify('%s', $1)", channelUpdates), string(bts))
	if err != nil {
		logger.Error(err, "failed to notify")
	}
}
