package server

import (
	"context"
	"net/http"
	"time"

	"github.com/faroshq/faros-ingress/pkg/dev/revdial"

	"github.com/pkg/errors"
	"k8s.io/klog"
)

var _ Interface = &Service{}

type Interface interface {
	Run(context.Context) error
	Shutdown(context.Context) error
}

type Service struct {
	addr     string
	server   *http.Server
	keyFile  string
	certFile string
}

func New(addr, certFile, keyFile string) (*Service, error) {
	s := &Service{
		addr:     addr,
		keyFile:  keyFile,
		certFile: certFile,
	}

	revPool := revdial.NewReversePool()
	mux := http.NewServeMux()
	mux.Handle("/", revPool)

	server := http.Server{
		Addr:    addr,
		Handler: mux,
	}
	s.server = &server

	return s, nil
}

func (s *Service) Run(ctx context.Context) error {
	klog.V(2).Infof("Starting remote server on %s", s.addr)

	return s.server.ListenAndServeTLS(s.certFile, s.keyFile)
}

func (s *Service) Shutdown(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()
	err := s.server.Shutdown(ctx)
	if err != nil {
		return errors.Wrap(err, "shutdown remote server")
	}
	return nil
}
