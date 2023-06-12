package client

import (
	"mini-gateway/reqcontext"
	"mini-gateway/selector"
	"net/http"
)

func newClient(s selector.Selector) *client {
	return &client{selector: s}
}

type client struct {
	selector selector.Selector
}

func (c *client) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	node, err := c.selector.Select(req.Context())
	if err != nil {
		return nil, err
	}
	req.URL.Scheme = "http"
	req.RequestURI = ""
	req.URL.Host = node.Address()
	if color, b := reqcontext.Color(req.Context()); b {
		req.Header.Add("x-color", color)
	}

	// 防止下游瞎几把弄
	req.Header.Set("Connection", "keep-alive")
	req.Close = false

	resp, err = node.Client().Do(req)
	return
}
