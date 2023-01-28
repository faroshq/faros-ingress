package models

import "time"

// User is a model for the User database model storing the user information.
type User struct {
	ID        string    `json:"id" yaml:"id" gorm:"primaryKey"`
	CreatedAt time.Time `json:"createdAt" yaml:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt" yaml:"updatedAt"`
	// Email is the email of the user. Must be unique.
	Email string `json:"email" yaml:"email" gorm:"uniqueIndex"`
}

// Workspace is a model for the Workspace database model storing the workspace information.
type Workspace struct {
	ID        string    `json:"id" yaml:"id" gorm:"primaryKey,uniqueIndex"`
	CreatedAt time.Time `json:"createdAt" yaml:"createdAt" grom:"index"`
	UpdatedAt time.Time `json:"updatedAt" yaml:"updatedAt"`
	UserID    string    `json:"userId" yaml:"userId" gorm:"index"`
	// Name is user facing name of the workspace
	Name string `json:"name" yaml:"name" gorm:"index"`
}

// Connection is a model for the connection database model storing the remote connection information.
type Connection struct {
	ID         string        `json:"id" yaml:"id" gorm:"primaryKey,uniqueIndex"`
	CreatedAt  time.Time     `json:"createdAt" yaml:"createdAt" grom:"index"`
	UpdatedAt  time.Time     `json:"updatedAt" yaml:"updatedAt"`
	LastUsedAt time.Time     `json:"lastUsedAt" yaml:"lastUsedAt" grom:"index"`
	TTL        time.Duration `json:"ttl,omitempty"`

	UserID string `json:"userId" yaml:"userId" gorm:"index"`
	// Name is user facing name of the remote connection
	Name string `json:"name" yaml:"name"`

	// Identity is the identity of the remote connection to be used for authentication remote dialing
	Identity string `json:"identity" yaml:"identity" gorm:"uniqueIndex"`

	// Hostname is the hostname of the remote connection
	Hostname string `json:"hostname" yaml:"hostname" gorm:"uniqueIndex"`

	// Secure is the flag for the stating if we should use basic auth
	Secure bool `json:"secure" yaml:"secure"`
	// BasicAuthHash is the authentication hash of the remote connection
	BasicAuthHash []byte `json:"basicAuthHash" yaml:"basicAuthHash"`

	// GatewayURL is the URL of the remote connection to be used for remote dialing
	GatewayURL string `json:"gatewayUrl" yaml:"gatewayUrl"`
}
