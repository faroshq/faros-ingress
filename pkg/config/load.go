package config

import (
	"bytes"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"

	utilfile "github.com/faroshq/faros-ingress/pkg/util/file"
	utiltls "github.com/faroshq/faros-ingress/pkg/util/tls"
)

// LoadAPI loads the configuration from the environment and flags
// Loading order:
// 1. Load .env file
// 2. Load envconfig from ENV variables and defaults
func LoadConfig() (*Config, error) {
	c := &Config{}
	godotenv.Load()

	err := envconfig.Process("", c)
	if err != nil {
		return c, err
	}

	if c.OIDC.OIDCAuthSessionKey == "" {
		fmt.Println("FAROS_OIDC_AUTH_SESSION_KEY not supplied, generating random one")
		c.OIDC.OIDCAuthSessionKey = uuid.Must(uuid.NewUUID()).String()
	}

	exists, err := utilfile.Exist(c.TLSCertFile)
	if err != nil {
		return nil, fmt.Errorf("failed to check if TLS cert file exists: %w", err)
	}
	if !exists {
		klog.V(2).Infof("TLS cert file does not provided, will use certMagic")
	}
	exists, err = utilfile.Exist(c.TLSKeyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to check if TLS key file exists: %w", err)
	}
	if !exists {
		klog.V(2).Infof("TLS key file does not provided, will use certMagic")
	}

	rest, err := loadKubeConfig(c.ClusterKubeConfigPath)
	if err != nil {
		return nil, err
	}
	c.ClusterRestConfig = rest

	return c, err
}

// loadKubeConfig loads a kubeconfig from disk. This method is
// intended to be common between fixture for servers whose lifecycle
// is test-managed and fixture for servers whose lifecycle is managed
// separately from a test run.
func loadKubeConfig(kubeconfigPath string) (*rest.Config, error) {
	exists, err := utilfile.Exist(kubeconfigPath)
	if err != nil {
		return nil, err
	}
	if !exists {
		config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
		if err != nil {
			return nil, err
		}
		return config, nil
	} else {
		rawConfig, err := clientcmd.LoadFromFile(kubeconfigPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load admin kubeconfig: %w", err)
		}

		return clientcmd.NewNonInteractiveClientConfig(*rawConfig, rawConfig.CurrentContext, nil, nil).ClientConfig()
	}

}

// LoadConnector loads the configuration from the environment and flags
// Loading order:
// 1. Load .env file
// 2. Load envconfig from ENV variables and defaults
func LoadConnector() (*ConnectorConfig, error) {
	c := &ConnectorConfig{}
	godotenv.Load()

	err := envconfig.Process("", c)
	if err != nil {
		return c, err
	}

	err = utilfile.EnsureDirExits(c.StateDir)
	if err != nil {
		return nil, fmt.Errorf("failed to ensure agent data dir exists: %w", err)
	}

	// certificate management. If there is no certificate provided and insecure is true,
	// we generate certificates and store them in the state dir to be used. If certificates
	// is provided insecure is set to false and used.
	if c.TLSServerCertFile != "" && c.TLSServerKeyFile != "" {
		c.TLSServerSkipVerify = false
	} else {
		klog.V(4).Info("No server certificates provided, generating temporary ones")
		name := "server"
		c.TLSServerSkipVerify = true
		c.TLSServerCertFile = fmt.Sprintf("%s/%s.crt", c.StateDir, name)
		c.TLSServerKeyFile = fmt.Sprintf("%s/%s.key", c.StateDir, name)

		key, cert, err := utiltls.GenerateKeyAndCertificate(name, nil, nil, false, false)
		if err != nil {
			return nil, fmt.Errorf("failed to generate temporary server certificates: %w", err)
		}

		err = writeCertificates(name, c.StateDir, key, cert)
		if err != nil {
			return nil, fmt.Errorf("failed to write temporary server certificates: %w", err)
		}
	}

	if c.TLSClientCertFile != "" && c.TLSClientKeyFile != "" {
		c.TLSClientSkipVerify = false
	} else {
		klog.V(4).Info("No client certificates provided, generating temporary ones")
		name := "client"
		c.TLSClientSkipVerify = true
		c.TLSClientCertFile = fmt.Sprintf("%s/%s.crt", c.StateDir, name)
		c.TLSClientKeyFile = fmt.Sprintf("%s/%s.key", c.StateDir, name)
		key, cert, err := utiltls.GenerateKeyAndCertificate(name, nil, nil, false, false)
		if err != nil {
			return nil, fmt.Errorf("failed to generate temporary client certificates: %w", err)
		}

		err = writeCertificates(name, c.StateDir, key, cert)
		if err != nil {
			return nil, fmt.Errorf("failed to write temporary client certificates: %w", err)
		}
	}

	return c, err
}

func writeCertificates(name, dir string, key *rsa.PrivateKey, cert []*x509.Certificate) error {
	// key in der format
	fp := filepath.Join(dir, name+".key")
	err := ioutil.WriteFile(fp, x509.MarshalPKCS1PrivateKey(key), 0600)
	if err != nil {
		return err
	}

	// cert in der format
	fp = filepath.Join(dir, name+".crt")
	err = ioutil.WriteFile(fp, cert[0].Raw, 0666)
	if err != nil {
		return err
	}

	buf := &bytes.Buffer{}
	b, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		return err
	}

	err = pem.Encode(buf, &pem.Block{Type: "PRIVATE KEY", Bytes: b})
	if err != nil {
		return err
	}

	err = pem.Encode(buf, &pem.Block{Type: "CERTIFICATE", Bytes: cert[0].Raw})
	if err != nil {
		return err
	}

	fp = filepath.Join(dir, name+".pem")
	return ioutil.WriteFile(fp, buf.Bytes(), 0600)
}
