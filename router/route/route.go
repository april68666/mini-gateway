package route

import (
	"mini-gateway/config"
	"net/http"
	"strings"
)

func NewRoute(predicates *config.Predicates, handler http.Handler) *Route {
	return &Route{
		handler:    handler,
		predicates: predicates,
	}
}

type Route struct {
	handler    http.Handler
	predicates *config.Predicates
}

func (r *Route) Handler() http.Handler {
	return r.handler
}

func (r *Route) Path() []string {
	return strings.Split(r.predicates.Path, ",")
}

func (r *Route) Match(req *http.Request) bool {
	if !r.matchMethod(req.Method) {
		return false
	}

	if !r.matchHeader(req.Header) {
		return false
	}

	return true
}

func (r *Route) matchMethod(method string) bool {
	if method == "OPTIONS" {
		return true
	}
	match := false
	ms := strings.Split(strings.TrimSpace(r.predicates.Method), ",")
	if len(ms) > 0 {
		for _, m := range ms {
			if strings.ToUpper(m) == strings.ToUpper(method) {
				match = true
				break
			}
		}

	}
	return match
}

func (r *Route) matchHeader(header http.Header) bool {
	for key, value := range r.predicates.Headers {
		v := header.Get(key)
		if v == "" || v != value {
			return false
		}
	}
	return true
}
