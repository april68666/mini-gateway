package router

import (
	"mini-gateway/config"
	"net/http"
	"strings"
)

func NewRoute(predicates *config.Predicates, handler http.Handler) *Route {
	pathKv := make(map[int]string)
	ss := strings.Split(strings.Trim(predicates.Path, "/"), "/")
	for i := 0; i < len(ss); i++ {
		pathKv[i] = ss[i]
	}
	return &Route{
		pathKv:     pathKv,
		predicates: predicates,
		handler:    handler,
	}
}

type Route struct {
	pathKv     map[int]string
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

	if !r.matchPath(req.URL.Path) {
		return false
	}

	if !r.mathHeader(req.Header) {
		return false
	}

	return true
}

func (r *Route) matchPath(path string) bool {
	ss := strings.Split(strings.Trim(path, "/"), "/")
	for i := 0; i < len(ss); i++ {
		v, ok := r.pathKv[i]
		if !ok {
			break
		}
		if ok && v == "*" {
			break
		}
		if ok && v != ss[i] {
			return false
		}
	}
	return true
}

func (r *Route) mathHeader(header http.Header) bool {
	for i := 0; i < len(r.predicates.Headers); i++ {
		head := r.predicates.Headers[i]
		v := header.Get(head.Key)
		if v == "" || v != head.Value {
			return false
		}
	}
	return true
}
