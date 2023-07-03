package client

import (
	"mini-gateway/loadbalance"
	"net/http"
	"net/url"
)

func newClient(s loadbalance.Picker, c *http.Client) *client {
	return &client{picker: s, httpClient: c}
}

type client struct {
	picker     loadbalance.Picker
	httpClient *http.Client
}

func (c *client) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	node, err := c.picker.Next(req.Context())
	if err != nil {
		return nil, err
	}

	u, err := url.Parse(node.Uri())
	if err != nil {
		return nil, err
	}
	req.RequestURI = ""
	req.URL.Host = u.Host
	req.URL.Scheme = u.Scheme

	resp, err = c.httpClient.Do(req)
	return
}
