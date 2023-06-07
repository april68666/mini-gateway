package middleware

import (
	"mini-gateway/config"
	"mini-gateway/slog"
	"net/http"
	"sort"
	"sync"
)

var middlewareFactory = sync.Map{}

type Factory func(c *config.Middleware) Middleware

type Middleware func(http.RoundTripper) http.RoundTripper

func Register(name string, f Factory) {
	middlewareFactory.Store(name, f)
}

func Get(cfg *config.Middleware) (f Factory, ok bool) {
	v, ok := middlewareFactory.Load(cfg.Name)
	if !ok {
		return nil, ok
	}
	return v.(Factory), ok
}

func BuildMiddleware(ms []*config.Middleware, next http.RoundTripper) (http.RoundTripper, error) {
	middlewareSort(ms)
	for i := 0; i < len(ms); i++ {
		factory, ok := Get(ms[i])
		if !ok {
			slog.Error("%s middleware not found", ms[i].Name)
			continue
		}
		middleware := factory(ms[i])
		next = middleware(next)
	}
	return next, nil
}

func middlewareSort(ms []*config.Middleware) {
	sort.Slice(ms, func(i, j int) bool {
		return ms[i].Order > ms[j].Order
	})
}
