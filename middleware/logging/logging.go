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
		return &Logging{next: next}
	}
}

type Logging struct {
	next http.RoundTripper
}

func (l *Logging) RoundTrip(req *http.Request) (*http.Response, error) {
	slog.Info("logging req %s", req.URL)
	return l.next.RoundTrip(req)
}
