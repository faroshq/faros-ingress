package client

import (
	"bytes"
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
	ListConnections(ctx context.Context) (*api.ConnectionList, error)
	GetConnection(ctx context.Context, agent api.Connection) (*api.Connection, error)
	DeleteConnection(ctx context.Context, agent api.Connection) error
	CreateConnection(ctx context.Context, agent api.Connection) (*api.Connection, error)
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

func (c *client) ListConnections(ctx context.Context) (*api.ConnectionList, error) {
	var result api.ConnectionList
	err := c.get(ctx, &result, "connections")
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *client) CreateConnection(ctx context.Context, conn api.Connection) (*api.Connection, error) {
	var result api.Connection
	err := c.post(ctx, conn, &result, "connections")
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *client) UpdateConnection(ctx context.Context, conn api.Connection) (*api.Connection, error) {
	var result api.Connection
	err := c.put(ctx, conn, &result, "connections", conn.ID)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *client) DeleteConnection(ctx context.Context, conn api.Connection) error {
	var result api.Connection
	err := c.delete(ctx, conn, &result, "connections", conn.ID)
	if err != nil {
		return err
	}

	return nil
}

func (c *client) GetConnection(ctx context.Context, conn api.Connection) (*api.Connection, error) {
	var result api.Connection
	err := c.get(ctx, &result, "connections", conn.ID)
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
	req, err := httputil.NewConnectionRequest(ctx, http.MethodGet, getURL(c.url, s...), nil)
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

func (c *client) put(ctx context.Context, in, out interface{}, s ...string) error {
	bytes, err := c.putB(ctx, in, s...)
	if err != nil {
		return err
	}
	if len(bytes) == 0 {
		return nil
	}
	return json.Unmarshal(bytes, &out)
}

func (c *client) putB(ctx context.Context, in interface{}, s ...string) ([]byte, error) {
	reqBytes, err := json.Marshal(in)
	if err != nil {
		return nil, err
	}
	reader := bytes.NewReader(reqBytes)

	req, err := httputil.NewConnectionRequest(ctx, http.MethodPut, getURL(c.url, s...), reader)
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

func (c *client) post(ctx context.Context, in, out interface{}, s ...string) error {
	reqBytes, err := json.Marshal(in)
	if err != nil {
		return err
	}

	bytes, err := c.postB(ctx, in, reqBytes, s...)
	if err != nil {
		return err
	}
	if len(bytes) == 0 {
		return nil
	}
	return json.Unmarshal(bytes, &out)
}

// PostB posts bytes json payload
func (c *client) postB(ctx context.Context, in interface{}, reqBytes []byte, s ...string) ([]byte, error) {
	reader := bytes.NewReader(reqBytes)

	req, err := httputil.NewConnectionRequest(ctx, http.MethodPost, getURL(c.url, s...), reader)
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

func (c *client) delete(ctx context.Context, in, out interface{}, s ...string) error {
	bytes, err := c.deleteB(ctx, in, s...)
	if err != nil {
		return err
	}
	if len(bytes) == 0 {
		return nil
	}
	return json.Unmarshal(bytes, &out)
}

func (c *client) deleteB(ctx context.Context, in interface{}, s ...string) ([]byte, error) {
	req, err := httputil.NewConnectionRequest(ctx, http.MethodDelete, getURL(c.url, s...), nil)
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
