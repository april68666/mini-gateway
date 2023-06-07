package router

import (
	"mini-gateway/config"
	"net/http"
	"strings"
)

func NewRoute(predicates *config.Predicates, handler http.Handler) *Route {
	return &Route{
		predicates: predicates,
		handler:    handler,
	}
}

type Route struct {
	predicates *config.Predicates
	handler    http.Handler
}

func (r *Route) match(req *http.Request) bool {
	method := strings.TrimSpace(r.predicates.Method)
	if method != "" {
		if strings.ToUpper(method) != req.Method {
			return false
		}
	}
	if strings.Trim(req.URL.Path, "/") != r.predicates.Path {
		return false
	}
	for i := 0; i < len(r.predicates.Headers); i++ {
		head := r.predicates.Headers[i]
		v := req.Header.Get(head.Key)
		if v == "" || v != head.Value {
			return false
		}
	}
	return true
}
