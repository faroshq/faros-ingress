package gateway

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/url"

	"github.com/faroshq/faros-ingress/pkg/api"
	"github.com/faroshq/faros-ingress/pkg/models"
	"github.com/faroshq/faros-ingress/pkg/util/responsewriter"
	"k8s.io/klog/v2"
)

type contextKey int

const (
	contextKeyConnection contextKey = iota
	contextKeyClient
	contextKeyResponse
)

func (s *Service) director(req *http.Request) {
	ctx := req.Context()
	conn := ctx.Value(contextKeyConnection).(*models.Connection)

	if conn == nil {
		klog.Errorf("no connection found in context")
		s.error(req, http.StatusInternalServerError, nil)
		return
	}

	prefix := "/api/v1alpha1/proxy/proxy/" + conn.Identity + "/" + req.URL.Path

	gw, err := url.Parse(s.config.InternalGatewayURL)
	if err != nil {
		s.error(req, http.StatusInternalServerError, err)
		return
	}

	req.Header.Add("X-Forwarded-Host", req.Host)
	req.Header.Add("X-Origin-Host", req.Host)
	req.URL.Scheme = "https"
	req.URL.Host = gw.Host
	req.URL.Path = prefix
	req.RequestURI = ""

	// drop auth headers only if ours are used
	if conn.Secure {
		req.Header.Del("Authorization")
	}
	// Once we are in proxy mode with request, drop the client header

	req.Header.Add(api.ConnectionClientHeader, api.ConnectionClientValue)

	cli := s.clientCache.Get(conn.Identity)
	if cli == nil {
		var err error
		cli, err = s.cli(ctx, gw, conn)
		if err != nil {
			s.error(req, http.StatusInternalServerError, err)
			return
		}

		s.clientCache.Put(conn.Identity, cli)
	}

	*req = *req.WithContext(context.WithValue(ctx, contextKeyClient, cli))

}

func (s *Service) roundTripper(r *http.Request) (*http.Response, error) {
	if resp, ok := r.Context().Value(contextKeyResponse).(*http.Response); ok {
		return resp, nil
	}

	cli := r.Context().Value(contextKeyClient).(*http.Client)
	if cli == nil {
		return nil, fmt.Errorf("no client")
	}
	return cli.Do(r)
}

func (s *Service) cli(ctx context.Context, gw *url.URL, conn *models.Connection) (*http.Client, error) {
	extGW, err := url.Parse(s.config.ExternalURL)
	if err != nil {
		return nil, err
	}
	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
				ServerName:         extGW.Hostname(),
			},
			DialContext: func(ctx context.Context, network, address string) (net.Conn, error) {
				return net.Dial(network, ":"+gw.Port())
			},
		},
	}, nil
}

func (s *Service) error(r *http.Request, statusCode int, err error) {
	if err != nil {
		klog.Error(err)
	}

	w := responsewriter.New(r)
	http.Error(w, http.StatusText(statusCode), statusCode)

	*r = *r.WithContext(context.WithValue(r.Context(), contextKeyResponse, w.Response()))
}
