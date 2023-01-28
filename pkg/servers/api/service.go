package api

import (
	"context"
	"log"
	"net/http"
	"time"

	health "github.com/InVisionApp/go-health/v2"
	healthhandlers "github.com/InVisionApp/go-health/v2/handlers"
	"github.com/caddyserver/certmagic"
	"github.com/go-acme/lego/v4/challenge/tlsalpn01"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/libdns/cloudflare"
	"go.uber.org/zap"
	"k8s.io/klog/v2"

	"github.com/faroshq/faros-ingress/pkg/config"
	"github.com/faroshq/faros-ingress/pkg/recover"
	"github.com/faroshq/faros-ingress/pkg/servers/api/auth"
	"github.com/faroshq/faros-ingress/pkg/store"
	storesql "github.com/faroshq/faros-ingress/pkg/store/sql"
	utilhttp "github.com/faroshq/faros-ingress/pkg/util/http"
)

var _ Interface = &Service{}

type Interface interface {
	Run(ctx context.Context) error
}
type Service struct {
	config        *config.Config
	authenticator auth.Authenticator
	server        *http.Server
	router        *mux.Router
	health        *health.Health
	store         store.Store
}

func New(ctx context.Context, config *config.Config) (*Service, error) {
	store, err := storesql.NewStore(ctx, &config.Database)
	if err != nil {
		return nil, err
	}

	authenticator, err := auth.NewAuthenticator(
		config,
		store,
		"/api/v1alpha1/oidc/callback",
	)
	if err != nil {
		return nil, err
	}

	s := &Service{
		config:        config,
		health:        health.New(),
		store:         store,
		authenticator: authenticator,
	}

	s.router = setupRouter()

	s.router.HandleFunc("/healthz", healthhandlers.NewJSONHandlerFunc(s.health, nil)) // /healthz

	apiRouter := s.router.PathPrefix("/api/v1alpha1").Subrouter()
	oidcRouter := apiRouter.PathPrefix("/oidc").Subrouter() // /api/v1alpha1/oidc
	oidcRouter.HandleFunc("/login", s.oidcLogin)            // /api/v1alpha1/oidc/login
	oidcRouter.HandleFunc("/callback", s.oidcCallback)      // /api/v1alpha1/oidc/callback

	agentsRouter := apiRouter.PathPrefix("/connections").Subrouter()                        // /api/v1alpha1/connection
	agentsRouter.HandleFunc("", s.listConnections).Methods(http.MethodGet)                  // /api/v1alpha1/connection
	agentsRouter.HandleFunc("/{connection}", s.getConnection).Methods(http.MethodGet)       // /api/v1alpha1/connection/{connection}
	agentsRouter.HandleFunc("/{connection}", s.deleteConnection).Methods(http.MethodDelete) // /api/v1alpha1/connection/{connection}
	agentsRouter.HandleFunc("", s.createConnection).Methods(http.MethodPost)                // /api/v1alpha1/connection
	agentsRouter.HandleFunc("/{connection}", s.updateConnection).Methods(http.MethodPut)    // /api/v1alpha1/connection/{connection}

	agentGateway := apiRouter.PathPrefix("/connection-gateways").Subrouter()                 // /api/v1alpha1/connection-gateway
	agentGateway.HandleFunc("/{connection}", s.getConnectionGateway).Methods(http.MethodGet) // /api/v1alpha1/connection-gateway/{connection}

	s.server = &http.Server{
		Addr:     config.APIAddr,
		ErrorLog: utilhttp.NewServerErrorLog(),
		Handler: handlers.CORS(
			handlers.AllowCredentials(),
			handlers.AllowedHeaders([]string{"Content-Type"}),
			handlers.AllowedMethods([]string{"GET", "POST", "PUT", "PATCH", "DELETE"}),
		)(s),
	}

	return s, nil
}

func (s *Service) Run(ctx context.Context) error {
	klog.Info("Starting API Service")
	go func() {
		defer recover.Panic()
		<-ctx.Done()

		err := s.store.Close()
		if err != nil {
			klog.Errorf("Error closing store: %v", err)
		}

		err = s.health.Stop()
		if err != nil {
			klog.Error(err)
		}

		ctx, cancel := context.WithTimeout(ctx, time.Second*30)
		defer cancel()
		err = s.server.Shutdown(ctx)
		if err != nil {
			klog.Error("api shutdown error", zap.Error(err))
		}
		klog.Info("Stopped API Service")
	}()

	if s.config.AutoCertEnabled() {
		klog.V(2).InfoS("Server will now listen with certMagic", "url", s.config.APIAddr)
		cache := certmagic.NewCache(certmagic.CacheOptions{
			GetConfigForCert: func(cert certmagic.Certificate) (*certmagic.Config, error) {
				return &certmagic.Config{
					Storage: &certmagic.FileStorage{
						Path: s.config.AutoCertCacheDir,
					},
				}, nil
			},
		})

		magic := certmagic.New(cache, certmagic.Config{
			Storage: &certmagic.FileStorage{
				Path: s.config.AutoCertCacheDir,
			},
		})

		ca := certmagic.LetsEncryptProductionCA
		if s.config.AutoCertUseStaging {
			ca = certmagic.LetsEncryptStagingCA
		}

		issuer := certmagic.NewACMEIssuer(magic, certmagic.ACMEIssuer{
			CA:                ca,
			Email:             s.config.AutoCertLEEmail,
			CertObtainTimeout: 5 * time.Minute,
			DNS01Solver: &certmagic.DNS01Solver{
				DNSProvider: &cloudflare.Provider{
					APIToken: s.config.AutoCertCloudFlareKey,
				},
			},
			Agreed:                  true,
			AltHTTPPort:             8080,
			AltTLSALPNPort:          8443,
			DisableHTTPChallenge:    true,
			DisableTLSALPNChallenge: true,
		})

		magic.Issuers = []certmagic.Issuer{issuer}

		// this obtains certificates or renews them if necessary
		err := magic.ManageSync(ctx, s.config.AutoCertAPIDomains)
		if err != nil {
			return err
		}

		s.server.TLSConfig = magic.TLSConfig()
		s.server.TLSConfig.NextProtos = append(s.server.TLSConfig.NextProtos, tlsalpn01.ACMETLS1Protocol)

		log.Printf("Serving https for domains: %+v", s.config.AutoCertAPIDomains)
		go func() {
			for {
				err := http.ListenAndServe(":8080", issuer.HTTPChallengeHandler(nil))
				if err != nil {
					klog.Error("api listen error", zap.Error(err))
				}
				time.Sleep(time.Second * 5)
			}
		}()
		err = s.server.ListenAndServeTLS("", "")
		if err != nil {
			klog.Error("api listen error", zap.Error(err))
		}

	} else {
		// Bring your own certs
		klog.V(2).InfoS("Server will now listen", "url", s.config.APIAddr)
		err := s.server.ListenAndServeTLS(s.config.TLSCertFile, s.config.TLSKeyFile)
		if err != nil {
			klog.Error("api listen error", zap.Error(err))

		}
	}
	return nil

}

func (s *Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

func setupRouter() *mux.Router {
	r := mux.NewRouter()
	r.Use(Panic())
	r.Use(Gzip())
	r.Use(Log())

	r.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
	})

	return r
}
