package router

import (
	"mini-gateway/router/route"
	"mini-gateway/router/trie"
	"net/http"
)

type Router interface {
	http.Handler
	LoadOrUpdateRoutes(routes []*route.Route)
}

type defaultRouter struct {
	trie *trie.Trie[*route.Route]
}

func NewDefaultRouter() Router {
	return &defaultRouter{}
}

func (r *defaultRouter) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if _, r, b := r.trie.Search(req.URL.Path); b {
		if r.Match(req) {
			r.Handler().ServeHTTP(w, req)
			return
		}
	}
	w.WriteHeader(http.StatusNotFound)
}

func (r *defaultRouter) LoadOrUpdateRoutes(routes []*route.Route) {
	t := trie.NewTrie[*route.Route]()
	for _, r := range routes {
		for _, path := range r.Path() {
			t.Insert(path, r)
		}
	}
	r.trie = t
}
