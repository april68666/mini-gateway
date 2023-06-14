package router

import (
	"mini-gateway/reqcontext"
	"mini-gateway/router/route"
	"mini-gateway/router/trie"
	"net/http"
	"sync/atomic"
)

type Router interface {
	http.Handler
	RegisterOrUpdateRoutes([]*route.Route)
}

type defaultRouter struct {
	trie atomic.Value
	//trie *trie.Trie[*route.Route]
}

func NewDefaultRouter() Router {
	return &defaultRouter{}
}

func (r *defaultRouter) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if params, re, b := r.trie.Load().(*trie.Trie[*route.Route]).Search(req.URL.Path); b {
		if re.Match(req) {
			if params != nil && len(params) > 0 {
				req = req.WithContext(reqcontext.WithParams(req.Context(), params))
			}
			re.Handler().ServeHTTP(w, req)
			return
		}
	}
	w.WriteHeader(http.StatusNotFound)
}

func (r *defaultRouter) RegisterOrUpdateRoutes(routes []*route.Route) {
	t := trie.NewTrie[*route.Route]()
	for _, r := range routes {
		for _, path := range r.Path() {
			t.Insert(path, r)
		}
	}
	r.trie.Store(t)
}
