package router

import (
	"mini-gateway/config"
	"mini-gateway/router/trie"
	"net/http"
	"strings"
)

func NewRoute(predicates *config.Predicates, handler http.Handler) *Route {
	t := trie.NewTrie()
	t.Insert(predicates.Path, handler)
	return &Route{
		trie:       t,
		handler:    handler,
		predicates: predicates,
	}
}

type Route struct {
	trie       *trie.Trie
	handler    http.Handler
	predicates *config.Predicates
}

func (r *Route) match(req *http.Request) bool {
	method := strings.TrimSpace(r.predicates.Method)
	if method != "" {
		if strings.ToUpper(method) != req.Method {
			return false
		}
	}

	if strings.TrimSpace(r.predicates.Path) != "" && !r.matchPath(req.URL.Path) {
		return false
	}

	if !r.matchHeader(req.Header) {
		return false
	}

	return true
}

func (r *Route) matchPath(path string) bool {
	_, _, b := r.trie.Search(path)
	return b
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
