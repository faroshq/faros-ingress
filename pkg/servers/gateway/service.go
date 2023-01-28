package gateway

import (
	"context"
	"log"
	"net/http"
	"net/http/httputil"
	"strings"
	"time"

	"go.uber.org/zap"
	"k8s.io/klog/v2"
	"k8s.io/utils/clock"

	"github.com/caddyserver/certmagic"
	"github.com/davecgh/go-spew/spew"
	"github.com/faroshq/faros-ingress/pkg/config"
	"github.com/faroshq/faros-ingress/pkg/h2rev2"
	"github.com/faroshq/faros-ingress/pkg/recover"
	"github.com/faroshq/faros-ingress/pkg/store"
	storesql "github.com/faroshq/faros-ingress/pkg/store/sql"
	"github.com/faroshq/faros-ingress/pkg/util/clientcache"
	utilhttp "github.com/faroshq/faros-ingress/pkg/util/http"
	"github.com/faroshq/faros-ingress/pkg/util/roundtripper"
	"github.com/go-acme/lego/v4/challenge/tlsalpn01"
	"github.com/libdns/cloudflare"
)

var _ Interface = &Service{}

type Interface interface {
	Run(ctx context.Context) error
}
type Service struct {
	config        *config.GatewayConfig
	server        *http.Server
	store         store.Store
	revPool       *h2rev2.ReversePool
	reverseProxy  *httputil.ReverseProxy
	authenticator *auth
	clientCache   clientcache.ClientCache
	clock         clock.Clock
}

func New(ctx context.Context, config *config.GatewayConfig) (*Service, error) {
	store, err := storesql.NewStore(ctx, &config.Database)
	if err != nil {
		return nil, err
	}

	revPool := h2rev2.NewReversePool(store)
	authenticator := newAuthenticator(store)

	s := &Service{
		config:        config,
		store:         store,
		revPool:       revPool,
		authenticator: authenticator,
		clientCache:   clientcache.New(time.Hour),
		clock:         clock.RealClock{},
	}

	s.server = &http.Server{
		Addr:     config.Addr,
		ErrorLog: utilhttp.NewServerErrorLog(),
		Handler:  s.handler(),
	}

	rp := &httputil.ReverseProxy{
		//ErrorLog:  utilhttp.NewServerErrorLog(),
		Director:  s.director,
		Transport: roundtripper.RoundTripperFunc(s.roundTripper),
	}
	s.reverseProxy = rp

	return s, nil
}

func (s *Service) Run(ctx context.Context) error {
	klog.Info("Starting Gateway Service")
	go func() {
		defer recover.Panic()
		<-ctx.Done()

		err := s.store.Close()
		if err != nil {
			klog.Errorf("Error closing store: %v", err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
		defer cancel()
		err = s.server.Shutdown(ctx)
		if err != nil {
			klog.Error("gateway shutdown error", zap.Error(err))
		}
		klog.Info("Stopped Gateway Service")
	}()

	go s.revPool.Run(ctx)
	go s.authenticator.run(ctx)
	go s.runGC(ctx)

	if s.config.AutoCertEnabled() {
		klog.V(2).InfoS("Server will now listen with certMagic", "url", s.config.Addr)
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
		err := magic.ManageSync(ctx, s.config.AutoCertDomains)
		if err != nil {
			return err
		}

		s.server.TLSConfig = magic.TLSConfig()
		s.server.TLSConfig.NextProtos = append(s.server.TLSConfig.NextProtos, tlsalpn01.ACMETLS1Protocol)

		log.Printf("Serving https for domains: %+v", s.config.AutoCertDomains)
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
		klog.V(2).InfoS("Server will now listen", "url", s.config.Addr)
		err := s.server.ListenAndServeTLS(s.config.TLSCertFile, s.config.TLSKeyFile)
		if err != nil {
			klog.Error("api listen error", zap.Error(err))

		}
	}
	return nil
}

func (s *Service) handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/v1alpha1/proxy/") {
			s.revPool.ServeHTTP(w, r)
		} else {
			spew.Dump(r.URL.Path)
			s.serveIngestor(w, r)
		}
	})
}
