package client

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/faroshq/faros-ingress/pkg/api"

	httputil "github.com/faroshq/faros-ingress/pkg/util/http"
)

var (
	_         Client = &client{}
	apiPrefix        = "/api/v1alpha1"
)

type Client interface {
	SetAccessKey(accessKey string)
	GetConnectionGateway(ctx context.Context, agent api.Connection) (*api.ConnectionGateway, error)
}

type client struct {
	url        *url.URL
	httpClient *httputil.Client

	accessKey string
}

func NewClient(url *url.URL, accessKey string, httpClient *httputil.Client) *client {
	if httpClient == nil {
		httpClient = httputil.DefaultInsecureClient // TODO
	}

	return &client{
		url:        url.JoinPath(apiPrefix),
		accessKey:  accessKey,
		httpClient: httpClient,
	}
}

func (c *client) SetAccessKey(accessKey string) {
	c.accessKey = accessKey
}

func (c *client) GetConnectionGateway(ctx context.Context, agent api.Connection) (*api.ConnectionGateway, error) {
	var result api.ConnectionGateway
	err := c.get(ctx, &result, "connection-gateways", agent.ID)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *client) get(ctx context.Context, out interface{}, s ...string) error {
	bytes, err := c.getB(ctx, s...)
	if err != nil {
		return err
	}
	if len(bytes) == 0 {
		return nil
	}
	return json.Unmarshal(bytes, &out)
}

func (c *client) getB(ctx context.Context, s ...string) ([]byte, error) {
	req, err := httputil.NewAgentRequest(ctx, http.MethodGet, getURL(c.url, s...), nil)
	if err != nil {
		return nil, err
	}

	var bearer = "Bearer " + c.accessKey
	req.Header.Add("Authorization", bearer)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return bytes, nil
}

func getURL(url *url.URL, s ...string) string {
	return strings.Join(append([]string{url.String()}, s...), "/")
}
