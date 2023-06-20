package logging

import (
	"mini-gateway/config"
	"mini-gateway/middleware"
	"mini-gateway/slog"
	"net/http"
	"time"
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
	start := time.Now()
	trip, err := l.next.RoundTrip(req)
	if err != nil {
		return nil, err
	}
	cost := time.Since(start)
	slog.Info("logging req %s,耗时:%fms", req.URL, cost.Seconds()*1000)
	return trip, err
}
