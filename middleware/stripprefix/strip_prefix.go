package stripprefix

import (
	"mini-gateway/config"
	"mini-gateway/middleware"
	"net/http"
	"strings"
)

const NAME = "strip_prefix"

func init() {
	middleware.Register(NAME, Factory)
}

func Factory(c *config.Middleware) middleware.Middleware {

	call := 0
	if v, ok := c.Args["call"]; ok {
		call = v.(int)
	}

	return func(next http.RoundTripper) http.RoundTripper {
		return &stripPrefix{next: next, call: call}
	}
}

type stripPrefix struct {
	call int
	next http.RoundTripper
}

func (s *stripPrefix) RoundTrip(req *http.Request) (*http.Response, error) {
	path := req.URL.Path
	path = strings.TrimLeft(path, "/")
	ps := strings.Split(path, "/")
	if len(ps) >= s.call {
		req.URL.Path = "/" + strings.Join(ps[s.call:], "/")
	}
	return s.next.RoundTrip(req)
}
