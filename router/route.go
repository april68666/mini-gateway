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
	for i := 0; i < len(r.predicates.Headers); i++ {
		head := r.predicates.Headers[i]
		v := header.Get(head.Key)
		if v == "" || v != head.Value {
			return false
		}
	}
	return true
}
