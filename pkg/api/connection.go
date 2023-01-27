package api

import "time"

const ConnectionClientHeader = "Faros-Connection-Client"
const ConnectionClientValue = "connector"
const ConnectionProxyValue = "proxy"

// Connection is an external connection model
type Connection struct {
	ID       string    `json:"id,omitempty"`
	Name     string    `json:"name,omitempty"`
	LastUsed time.Time `json:"lastUsed,omitempty"`

	Identity string `json:"identity,omitempty"`
	Hostname string `json:"hostname,omitempty"`

	Secure   bool   `json:"secure,omitempty"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

type ConnectionList struct {
	Items []Connection `json:"items,omitempty"`
}

type ConnectionGateway struct {
	Hostname string `json:"hostname,omitempty"`
}
