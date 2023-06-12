package router

import (
	"mini-gateway/config"
	"mini-gateway/router/trie"
	"net/http"
	"strings"
)

func NewRoute(predicates *config.Predicates, handler http.Handler) *Route {
	t := trie.NewTrie[http.Handler]()
	ps := strings.Split(predicates.Path, ",")
	for _, path := range ps {
		t.Insert(path, handler)
	}
	return &Route{
		trie:       t,
		handler:    handler,
		predicates: predicates,
	}
}

type Route struct {
	trie       *trie.Trie[http.Handler]
	handler    http.Handler
	predicates *config.Predicates
}

func (r *Route) match(req *http.Request) bool {
	if !r.matchMethod(req.Method) {
		return false
	}

	if strings.TrimSpace(r.predicates.Path) != "" && !r.matchPath(req.URL.Path) {
		return false
	}

	if !r.matchHeader(req.Header) {
		return false
	}

	return true
}

func (r *Route) matchMethod(method string) bool {
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
