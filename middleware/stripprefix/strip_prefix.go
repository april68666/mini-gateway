package stripprefix

import (
	"mini-gateway/config"
	"mini-gateway/middleware"
	"mini-gateway/slog"
	"net/http"
	"strconv"
	"strings"
)

const NAME = "strip_prefix"

func init() {
	middleware.Register(NAME, Factory)
}

func Factory(c *config.Middleware) middleware.Middleware {

	args := make(map[string]string)
	for _, arg := range c.Args {
		args[arg.Key] = arg.Value

	}
	call := 0
	if v, ok := args["call"]; ok {
		i, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			slog.Error("strip_prefix middleware  call parse error:%s", err.Error())
		}
		call = int(i)
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
