package storesql

import (
	"context"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/mjudeikis/portal/pkg/models"
	"github.com/mjudeikis/portal/pkg/store"
)

// GetUser gets full user based on args user
// Search: ID or Name and Namespace must be provided
func (s *Store) GetUser(ctx context.Context, p models.User) (*models.User, error) {
	switch {
	case p.ID != "":
		// OK, getting by ID
	case p.Email != "":
		// OK, getting by Email
	default:
		return nil, store.ErrFailToQuery
	}

	result := models.User{}
	if err := s.db.WithContext(ctx).Where(&p).First(&result).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, store.ErrRecordNotFound
		}
		return nil, err
	}

	return &result, nil
}

// CreateUser creates user and assigns unique ID
func (s *Store) CreateUser(ctx context.Context, p models.User) (*models.User, error) {
	p.ID = uuid.New().String()

	err := s.db.WithContext(ctx).Create(&p).Error
	if err != nil {
		return nil, err
	}

	s.notifyUpdatedUser(ctx, p.ID, models.EventCreated)

	return s.GetUser(ctx, models.User{ID: p.ID})
}

// UpdateUser updates user based on user ID
func (s *Store) UpdateUser(ctx context.Context, p models.User) (*models.User, error) {
	switch {
	case p.ID != "":
		// OK, getting by ID
	default:
		return nil, store.ErrFailToQuery
	}

	query := models.User{ID: p.ID}
	err := s.db.WithContext(ctx).Model(&models.User{}).Where(&query).Save(&p).Error
	if err != nil {
		return nil, err
	}

	s.notifyUpdatedUser(ctx, p.ID, models.EventUpdated)

	return s.GetUser(ctx, models.User{ID: p.ID})
}

// DeleteUser deletes user based on user ID
func (s *Store) DeleteUser(ctx context.Context, p models.User) error {
	switch {
	case p.ID != "":
		// OK, getting by ID
	default:
		return store.ErrFailToQuery
	}

	return s.db.WithContext(ctx).Delete(&p).Error
}

func (s *Store) ListUsers(ctx context.Context, p models.User) ([]models.User, error) {
	results := []models.User{}
	if err := s.db.WithContext(ctx).Where(&p).Find(&results).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, store.ErrRecordNotFound
		}
		return nil, err
	}

	s.notifyUpdatedUser(ctx, p.ID, models.EventDeleted)

	return results, nil
}
