package client

import (
	"mini-gateway/reqcontext"
	"mini-gateway/selector"
	"net/http"
	"net/url"
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

	if req.Form != nil {
		req.URL.RawQuery = cleanQueryParams(req.URL.RawQuery)
	}

	// 防止下游瞎几把弄
	req.Header.Set("Connection", "keep-alive")
	req.Close = false

	resp, err = node.Client().Do(req)
	return
}

func cleanQueryParams(s string) string {
	reEncode := func(s string) string {
		v, _ := url.ParseQuery(s)
		return v.Encode()
	}
	for i := 0; i < len(s); {
		switch s[i] {
		case ';':
			return reEncode(s)
		case '%':
			if i+2 >= len(s) || !isHex(s[i+1]) || !isHex(s[i+2]) {
				return reEncode(s)
			}
			i += 3
		default:
			i++
		}
	}
	return s
}

func isHex(c byte) bool {
	switch {
	case '0' <= c && c <= '9':
		return true
	case 'a' <= c && c <= 'f':
		return true
	case 'A' <= c && c <= 'F':
		return true
	}
	return false
}
