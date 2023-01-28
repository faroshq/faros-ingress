package config

import (
	"time"

	"k8s.io/client-go/rest"
)

const (
	ConfigFileName = "config.yaml"
)

type Config struct {
	// APIAddr is the address to bind the controller to.
	APIAddr string `envconfig:"FAROS_API_ADDR" required:"true" default:":8443"`
	// GatewayAddr is the address to bind the gateway to.
	GatewayAddr string `envconfig:"FAROS_GATEWAY_ADDR" required:"true" default:":8444"`
	// ExternalAPIURL is the URL that the api is externally accessible at.
	ExternalAPIURL string `envconfig:"FAROS_API_EXTERNAL_URL" required:"true" default:"https://faros.dev.faros.sh"`
	// ExternalGatewayURL is the URL that the gateway is externally accessible at.
	ExternalGatewayURL string `envconfig:"FAROS_GATEWAY_EXTERNAL_URL" required:"true" default:"https://gateway.faros.sh"`
	// DefaultGateway is the default gateway to use for the access.
	DefaultGateway string `envconfig:"FAROS_DEFAULT_GATEWAY" required:"true" default:"https://gateway.dev.faros.sh"`
	// InternalGatewayURL is the URL that the gateway is internally accessible at.
	InternalGatewayURL string `envconfig:"FAROS_GATEWAY_INTERNAL_GATEWAY_URL" required:"true" default:"https://localhost:8444"`
	// HostnameSuffix is the suffix of the hostname to use for the access.
	HostnameSuffix string `envconfig:"FAROS_HOSTNAME_SUFFIX" required:"true" default:"apps.faros.sh"`
	// ClusterKubeConfigPath
	ClusterKubeConfigPath string `envconfig:"FAROS_CLUSTER_KUBECONFIG"`
	ClusterRestConfig     *rest.Config

	// Both api and gateway use the same TLS configuration. Use common certs or change the code :)
	// In prod we use auto-certs so this is not an issue.
	TLSKeyFile  string `envconfig:"FAROS_TLS_KEY_FILE" default:""`
	TLSCertFile string `envconfig:"FAROS_TLS_CERT_FILE" default:""`

	// AutoCertAPIDomains is the list of domains to use for auto-cert.
	AutoCertAPIDomains []string `envconfig:"FAROS_API_AUTO_DNS_DOMAIN" required:"true" default:"api.faros.sh"`
	// AutoCertGatewayDomains is the list of domains to use for auto-cert.
	AutoCertGatewayDomains []string `envconfig:"FAROS_GATEWAY_AUTO_DNS_DOMAIN" required:"true" default:"gateway.faros.sh"`

	AutoCertCacheDir      string `envconfig:"FAROS_AUTO_CERT_CACHE_DIR" required:"true" default:"/faros/cache"`
	AutoCertLEEmail       string `envconfig:"FAROS_AUTO_CERT_LE_EMAIL" default:""`
	AutoCertCloudFlareKey string `envconfig:"FAROS_AUTO_CERT_CLOUDFLARE_KEY" default:""`
	AutoCertUseStaging    bool   `envconfig:"FAROS_AUTO_CERT_USE_STAGING" default:"true"`

	Database Database   `yaml:"database,omitempty"`
	OIDC     OIDCConfig `yaml:"oidcConfig,omitempty"`

	// Quota is the quota to use for the connections per user. 0 means no quota.
	ConnectionQuota int `envconfig:"FAROS_CONNECTIONS_QUOTA" default:"0"`
}

type OIDCConfig struct {
	OIDCIssuerURL         string `envconfig:"FAROS_OIDC_ISSUER_URL" yaml:"oidcIssuerURL,omitempty" default:"https://dex.dev.faros.sh"`
	OIDCClientID          string `envconfig:"FAROS_OIDC_CLIENT_ID" yaml:"oidcClientID,omitempty" default:"faros"`
	OIDCClientSecret      string `envconfig:"FAROS_OIDC_CLIENT_SECRET" yaml:"oidcClientSecret,omitempty" default:"faros"`
	OIDCCASecretName      string `envconfig:"FAROS_OIDC_CA_SECRET_NAME" yaml:"oidcCASecretName,omitempty" default:"dex-pki-ca"`
	OIDCCASecretNamespace string `envconfig:"FAROS_OIDC_CA_SECRET_NAMESPACE" yaml:"oidcCASecretNamespace,omitempty" default:"dex"`
	OIDCUsernameClaim     string `envconfig:"FAROS_OIDC_USERNAME_CLAIM" yaml:"oidcFarosUsernameClaim,omitempty" default:"email"`
	OIDCUserPrefix        string `envconfig:"FAROS_OIDC_USER_PREFIX" yaml:"oidcUserPrefix,omitempty" default:"faros-sso"`
	OIDCGroupsPrefix      string `envconfig:"FAROS_OIDC_GROUPS_PREFIX" yaml:"oidcGroupsPrefix,omitempty" default:"faros-sso"`
	OIDCAuthSessionKey    string `envconfig:"FAROS_OIDC_AUTH_SESSION_KEY" yaml:"oidcAuthSessionKey,omitempty" default:""`
}

type Database struct {
	SqliteURI string `envconfig:"FAROS_DATABASE_SQLITE_URI" default:"dev/database.sqlite3"`
	// Name of the database
	Name string `envconfig:"FAROS_DATABASE_NAME" default:"faros"`
	// Type is the type of database to use.
	Type string `envconfig:"FAROS_DATABASE_TYPE" default:"sqlite" `
	// Host is the host of the database
	Host string `envconfig:"FAROS_DATABASE_HOST" default:"localhost"`
	// Port is the port of the database
	Port int `envconfig:"FAROS_DATABASE_PORT" default:"5432"`
	// Password is the password of the database
	Password string `envconfig:"FAROS_DATABASE_PASSWORD" default:""`
	// Username is the username of the database
	Username string `envconfig:"FAROS_DATABASE_USERNAME" default:""`
	// MaxConnIdleTime is the maximum amount of time a database connection can be idle
	MaxConnIdleTime time.Duration `envconfig:"FAROS_DATABASE_MAX_CONN_IDLE_TIME" default:"30s"`
	//MaxConnLifeTime is the maximum amount of time a database connection can be used
	MaxConnLifeTime time.Duration `envconfig:"FAROS_DATABASE_MAX_CONN_LIFE_TIME" default:"1h"`
}

type ConnectorConfig struct {
	// ControllerURL is the URL that the controller is externally reachable at.
	ControllerURL string `envconfig:"FAROS_EXTERNAL_URL" required:"true" default:"https://api.faros.sh"`
	// DownstreamURL is downstream URL for the connector to connect to.
	DownstreamURL string `envconfig:"FAROS_DOWNSTREAM_URL" required:"true" default:"http://localhost:8080"`
	// Token is the token used to authenticate with the API server.
	Token string `envconfig:"FAROS_TOKEN" default:""`
	// ConnectionID is the ID of the connection.
	ConnectionID string `envconfig:"FAROS_CONNECTION_ID" default:""`
	// StateDir is the directory where the connector will store its state.
	StateDir string `envconfig:"FAROS_STATE_DIR" required:"true" default:"/var/tmp/faros/connector"`

	// TLSServerKeyFile is the path to the TLS server key file.
	TLSServerKeyFile string `envconfig:"FAROS_TLS_SERVER_KEY_FILE"`
	// TLSServerCertFile is the path to the TLS server cert file.
	TLSServerCertFile string `envconfig:"FAROS_TLS_SERVER_CERT_FILE"`
	// TLSServerSkipVerify disables TLS verification.
	TLSServerSkipVerify bool

	// TLSClientKeyFile is the path to the TLS client key file.
	TLSClientKeyFile string `envconfig:"FAROS_TLS_CLIENT_KEY_FILE"`
	// TLSClientCertFile is the path to the TLS client cert file.
	TLSClientCertFile string `envconfig:"FAROS_TLS_CLIENT_CERT_FILE"`
	// TLSClientSkipVerify disables TLS verification.
	TLSClientSkipVerify bool
}

func (c *Config) AutoCertEnabled() bool {
	if c.TLSCertFile == "" && c.TLSKeyFile == "" {
		return true
	}
	return false
}
