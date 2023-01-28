package api

import "time"

const ConnectionClientHeader = "Faros-Connection-Client"
const ConnectionClientValue = "connector"
const ConnectionProxyValue = "proxy"

// Connection is an external connection model
type Connection struct {
	ID       string        `json:"id,omitempty" yaml:"id,omitempty"`
	Name     string        `json:"name,omitempty" yaml:"name,omitempty"`
	LastUsed time.Time     `json:"lastUsed,omitempty" yaml:"lastUsed,omitempty"`
	TTL      time.Duration `json:"ttl,omitempty" yaml:"ttl,omitempty"`

	Token    string `json:"token,omitempty" yaml:"token,omitempty"`
	Hostname string `json:"hostname,omitempty" yaml:"hostname,omitempty"`

	Secure   bool   `json:"secure,omitempty" yaml:"secure,omitempty"`
	Username string `json:"username,omitempty" yaml:"username,omitempty"`
	Password string `json:"password,omitempty" yaml:"password,omitempty"`
}

type ConnectionList struct {
	Items []Connection `json:"items,omitempty" yaml:"items,omitempty"`
}

type ConnectionGateway struct {
	Hostname string `json:"hostname,omitempty" yaml:"hostname,omitempty"`
}
