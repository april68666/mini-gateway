package router

import (
	"net/http"
)

type Router interface {
	http.Handler
	LoadOrUpdateRoutes(routes []*Route)
}

type defaultRouter struct {
	routes []*Route
}

func NewDefaultRouter() Router {
	return &defaultRouter{}
}

func (r *defaultRouter) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	for i := 0; i < len(r.routes); i++ {
		if r.routes[i].match(req) {
			r.routes[i].handler.ServeHTTP(w, req)
			return
		}
	}
	w.WriteHeader(http.StatusNotFound)
}

func (r *defaultRouter) LoadOrUpdateRoutes(routes []*Route) {
	r.routes = routes
}
