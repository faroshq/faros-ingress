package connector

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"golang.org/x/net/http2"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
	"k8s.io/utils/clock"

	"github.com/faroshq/faros-ingress/pkg/api"
	"github.com/faroshq/faros-ingress/pkg/config"
	"github.com/faroshq/faros-ingress/pkg/connector/client"
	"github.com/faroshq/faros-ingress/pkg/h2rev2"
	utilhttp "github.com/faroshq/faros-ingress/pkg/util/http"
)

const (
	dialerSufix = "/api/v1alpha1/proxy"
)

type Connection struct {
	config         *config.ConnectorConfig
	upstreamClient *http.Client
	tlsConfig      *tls.Config
	apiClient      client.Client

	gatewayURL string
}

func New(config *config.ConnectorConfig) (*Connection, error) {
	serverCertFile, err := ioutil.ReadFile(config.TLSServerCertFile)
	if err != nil {
		return nil, err
	}

	serverClientCert, err := x509.ParseCertificate(serverCertFile)
	if err != nil {
		return nil, err
	}

	pool := x509.NewCertPool()
	pool.AddCert(serverClientCert)

	serverKeyFile, err := ioutil.ReadFile(config.TLSServerKeyFile)
	if err != nil {
		return nil, err
	}

	serverKey, err := x509.ParsePKCS1PrivateKey(serverKeyFile)
	if err != nil {
		return nil, err
	}

	tlsConfig := &tls.Config{
		RootCAs: pool,
		Certificates: []tls.Certificate{
			{
				Certificate: [][]byte{
					serverCertFile,
				},
				PrivateKey: serverKey,
			},
		},
		ServerName:         "gateway.faros.sh", // TODO
		InsecureSkipVerify: true,
	}

	upstreamClient := &http.Client{Transport: transport2()}

	u, err := url.Parse(config.ControllerURL)
	if err != nil {
		return nil, err
	}

	apiClient := client.NewClient(u, config.Token, nil)

	return &Connection{
		apiClient:      apiClient,
		config:         config,
		upstreamClient: upstreamClient,
		tlsConfig:      tlsConfig,
	}, nil
}

func transport2() *http2.Transport {
	return &http2.Transport{
		TLSClientConfig:    tlsConfig(),
		DisableCompression: true,
		AllowHTTP:          false,
	}
}

func tlsConfig() *tls.Config {
	crt, err := ioutil.ReadFile("./dev/server.crt")
	if err != nil {
		log.Fatal(err)
	}

	rootCAs := x509.NewCertPool()
	rootCAs.AppendCertsFromPEM(crt)

	return &tls.Config{
		RootCAs:            rootCAs,
		InsecureSkipVerify: true,
		ServerName:         "gateway.faros.sh", // TODO
	}
}

// startTunnel blocks until the context is cancelled trying to establish a tunnel against the specified target
func (c *Connection) Run(ctx context.Context) {
	// connect to create the reverse tunnels
	var (
		initBackoff   = time.Second
		maxBackoff    = time.Minute
		resetDuration = time.Second * 2
		backoffFactor = 2.0
		jitter        = 1.0
		clock         = &clock.RealClock{}
		sliding       = true
	)

	backoffMgr := wait.NewExponentialBackoffManager(initBackoff, maxBackoff, resetDuration, backoffFactor, jitter, clock)
	logger := klog.FromContext(ctx)

	// get gateway url:
	// call API and ask for agent gateway url

	wait.BackoffUntil(func() {
		logger.V(4).Info("get gateway for tunnel")
		gateway, err := c.apiClient.GetConnectionGateway(ctx, api.Connection{ID: c.config.ConnectionID})
		if err != nil {
			klog.Error(err, "failed to get gateway", "connection", c.config.ConnectionID)
			return
		}

		c.gatewayURL = gateway.Hostname + dialerSufix

		logger.V(4).Info("starting tunnel")
		err = c.startTunneler(ctx)
		if err != nil {
			logger.Error(err, "failed to create tunnel")
		}
	}, backoffMgr, sliding, ctx.Done())
}

func (c *Connection) startTunneler(ctx context.Context) error {
	logger := klog.FromContext(ctx)

	logger = logger.WithValues("to", c.config.DownstreamURL).WithValues("from", c.gatewayURL)
	logger.V(2).Info("connecting to destination URL")

	l, err := h2rev2.NewListener(c.upstreamClient, c.gatewayURL, c.config.Token)
	if err != nil {
		return err
	}

	// client --> local dev instance
	downstreamURL, err := url.Parse(c.config.DownstreamURL)
	if err != nil {
		return err
	}

	// dev-proxy-server --> local dev instance
	proxy := httputil.NewSingleHostReverseProxy(downstreamURL)
	if err != nil {
		return err
	}

	director := proxy.Director
	proxy.Director = func(req *http.Request) {
		logger.V(4).Info("proxying request", "method", req.Method, "path", req.URL.Path)
		req.Host = downstreamURL.Host
		director(req)
	}
	clientDownstream := utilhttp.DefaultInsecureClient // TODO
	proxy.Transport = clientDownstream.Transport       // TODO

	// reverse proxy the request coming from the reverse connection to the apiserver
	server := &http.Server{Handler: proxy}
	defer server.Close()

	logger.V(2).Info("serving on reverse connection")
	errCh := make(chan error)
	go func() {
		errCh <- server.Serve(l)
	}()

	select {
	case err = <-errCh:
	case <-ctx.Done():
		err = server.Close()
	}
	logger.V(2).Info("stop serving on reverse connection")
	return err
}
