package storesql

import (
	"context"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/faroshq/faros-ingress/pkg/models"
	"github.com/faroshq/faros-ingress/pkg/store"
)

// GetConnection gets remote cluster based on remote cluster ID
func (s *Store) GetConnection(ctx context.Context, p models.Connection) (*models.Connection, error) {
	switch {
	case p.ID != "":
		// OK, getting by ID
	case p.UserID != "" && p.Name != "":
		// OK, getting by UserID and Name
	case p.Hostname != "":
		// OK, getting by Hostname
	default:
		return nil, store.ErrFailToQuery
	}

	result := models.Connection{}
	if err := s.db.WithContext(ctx).Where(&p).First(&result).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, store.ErrRecordNotFound
		}
		return nil, err
	}

	return &result, nil
}

// CreateConnection creates remote cluster object
func (s *Store) CreateConnection(ctx context.Context, p models.Connection) (*models.Connection, error) {
	p.ID = uuid.New().String()

	err := s.db.WithContext(ctx).Create(&p).Error
	if err != nil {
		return nil, err
	}

	s.notifyUpdatedConnection(ctx, p.ID, models.EventCreated)

	return s.GetConnection(ctx, models.Connection{ID: p.ID})
}

// UpdateConnection updates remote cluster based on remote cluster ID
func (s *Store) UpdateConnection(ctx context.Context, p models.Connection) (*models.Connection, error) {
	switch {
	case p.ID != "":
		// OK, getting by ID
	case p.UserID != "" && p.Name != "":
		// OK, getting by UserID and Name
	default:
		return nil, store.ErrFailToQuery
	}

	query := models.Connection{ID: p.ID}
	err := s.db.WithContext(ctx).Model(&models.Workspace{}).Where(&query).Save(&p).Error
	if err != nil {
		return nil, err
	}

	s.notifyUpdatedConnection(ctx, p.ID, models.EventUpdated)

	return s.GetConnection(ctx, models.Connection{ID: p.ID})
}

// DeleteWorkspace deletes remote clusters based on cluster ID
func (s *Store) DeleteConnection(ctx context.Context, p models.Connection) error {
	switch {
	case p.ID != "":
		// OK, getting by ID
	default:
		return store.ErrFailToQuery
	}

	s.notifyUpdatedConnection(ctx, p.ID, models.EventDeleted)

	return s.db.WithContext(ctx).Delete(&p).Error
}

// ListConnections lists clusters
func (s *Store) ListConnections(ctx context.Context, p models.Connection) ([]models.Connection, error) {
	switch {
	case p.UserID != "":
		// OK, listing by UserID
	default:
		return nil, store.ErrFailToQuery
	}

	results := []models.Connection{}
	if err := s.db.WithContext(ctx).Where(&p).Find(&results).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, store.ErrRecordNotFound
		}
		return nil, err
	}

	return results, nil
}

// ListAllConnections lists Connections without filtering
func (s *Store) ListAllConnections(ctx context.Context) ([]models.Connection, error) {
	results := []models.Connection{}
	p := models.Connection{}
	if err := s.db.WithContext(ctx).Where(&p).Find(&results).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, store.ErrRecordNotFound
		}
		return nil, err
	}

	return results, nil
}
