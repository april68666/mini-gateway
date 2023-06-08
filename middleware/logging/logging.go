package logging

import (
	"mini-gateway/config"
	"mini-gateway/middleware"
	"mini-gateway/slog"
	"net/http"
)

const NAME = "logging"

func init() {
	middleware.Register(NAME, Factory)
}

func Factory(c *config.Middleware) middleware.Middleware {
	return func(next http.RoundTripper) http.RoundTripper {
		return &logging{next: next}
	}
}

type logging struct {
	next http.RoundTripper
}

func (l *logging) RoundTrip(req *http.Request) (*http.Response, error) {
	slog.Info("logging req %s", req.URL)
	return l.next.RoundTrip(req)
}
