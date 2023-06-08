package router

import (
	"net/http"
	"sync"
)

type Router interface {
	http.Handler
	LoadOrUpdateRoutes(routes []*Route)
}

type defaultRouter struct {
	mux    sync.RWMutex
	routes []*Route
}

func NewDefaultRouter() Router {
	return &defaultRouter{}
}

func (r *defaultRouter) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mux.RLock()
	for i := 0; i < len(r.routes); i++ {
		if r.routes[i].match(req) {
			r.mux.RUnlock()
			r.routes[i].handler.ServeHTTP(w, req)
			return
		}
	}
	r.mux.RUnlock()
	w.WriteHeader(http.StatusNotFound)
}

func (r *defaultRouter) LoadOrUpdateRoutes(routes []*Route) {
	r.mux.Lock()
	defer r.mux.Unlock()
	r.routes = routes
}
