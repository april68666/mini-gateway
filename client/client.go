package client

import (
	"mini-gateway/reqcontext"
	"mini-gateway/selector"
	"net/http"
	"net/url"
)

func newClient(s selector.Selector, c *http.Client) *client {
	return &client{selector: s, httpClient: c}
}

type client struct {
	selector   selector.Selector
	httpClient *http.Client
}

func (c *client) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	node, err := c.selector.Select(req.Context())
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

	if color, b := reqcontext.Color(req.Context()); b {
		req.Header.Add("x-color", color)
	}

	if req.Form != nil {
		req.URL.RawQuery = cleanQueryParams(req.URL.RawQuery)
	}

	// 防止下游瞎几把弄
	req.Header.Set("Connection", "keep-alive")
	req.Close = false

	resp, err = c.httpClient.Do(req)
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
