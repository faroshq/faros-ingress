package store

import (
	"context"
	"errors"

	"github.com/faroshq/faros-ingress/pkg/models"
)

//go:generate mockgen -source $GOFILE -destination store_mocks.go -package $GOPACKAGE

type Store interface {
	GetConnection(context.Context, models.Connection) (*models.Connection, error)
	ListConnections(context.Context, models.Connection) ([]models.Connection, error)
	ListAllConnections(ctx context.Context) ([]models.Connection, error)
	DeleteConnection(context.Context, models.Connection) error
	CreateConnection(context.Context, models.Connection) (*models.Connection, error)
	UpdateConnection(context.Context, models.Connection) (*models.Connection, error)

	GetUser(context.Context, models.User) (*models.User, error)
	ListUsers(context.Context, models.User) ([]models.User, error)
	DeleteUser(context.Context, models.User) error
	CreateUser(context.Context, models.User) (*models.User, error)
	UpdateUser(context.Context, models.User) (*models.User, error)

	SubscribeChanges(ctx context.Context, callback func(event *models.Event) error) error

	// Status is a health check endpoint
	Status() (interface{}, error)
	RawDB() interface{}
	Close() error
}

var ErrFailToQuery = errors.New("malformed request. failed to query")
var ErrRecordNotFound = errors.New("object not found")
