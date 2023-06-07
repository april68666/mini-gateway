package proxy

import (
	"bytes"
	"context"
	"io"
	"mini-gateway/client"
	"mini-gateway/config"
	"mini-gateway/middleware"
	_ "mini-gateway/middleware/logging"
	"mini-gateway/router"
	"mini-gateway/slog"
	"net"
	"net/http"
	"strings"
)

func NewProxy(factory client.Factory, router router.Router, ms []*config.Middleware) *Proxy {
	return &Proxy{
		router:     router,
		factory:    factory,
		middleware: ms,
	}
}

type Proxy struct {
	router     router.Router
	factory    client.Factory
	middleware []*config.Middleware
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p.router.ServeHTTP(w, r)
}

func (p *Proxy) buildEndpoints(endpoint *config.Endpoint, ms []*config.Middleware) (http.Handler, error) {
	factory, err := p.factory(endpoint)
	if err != nil {
		return nil, err
	}
	tripper, err := middleware.BuildMiddleware(endpoint.Middlewares, factory)
	if err != nil {
		return nil, err
	}
	tripper, err = middleware.BuildMiddleware(ms, tripper)
	if err != nil {
		return nil, err
	}

	return http.Handler(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		p.setXForwarded(req)
		body, err := io.ReadAll(req.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadGateway)
			slog.Error(err.Error())
			return
		}
		req.GetBody = func() (io.ReadCloser, error) {
			reader := bytes.NewReader(body)
			return io.NopCloser(reader), nil
		}
		reader := bytes.NewReader(body)
		req.Body = io.NopCloser(reader)
		resp, err := tripper.RoundTrip(req.Clone(context.TODO()))
		if err != nil {
			w.WriteHeader(http.StatusBadGateway)
			slog.Error(err.Error())
			return
		}
		headers := w.Header()
		for k, v := range resp.Header {
			headers[k] = v
		}
		w.WriteHeader(resp.StatusCode)

		func() {
			if resp.Body == nil {
				return
			}
			defer func(Body io.ReadCloser) {
				_ = Body.Close()
			}(resp.Body)
			_, err = io.Copy(w, resp.Body)
			if err != nil {
				return
			}
			//	https://pkg.go.dev/net/http#ResponseWriter
			//	https://pkg.go.dev/net/http#example-ResponseWriter-Trailers
			for k, v := range resp.Trailer {
				headers[http.TrailerPrefix+k] = v
			}
		}()
	})), nil
}

func (p *Proxy) LoadOrUpdateEndpoints(endpoints []*config.Endpoint) {
	routes := make([]*router.Route, 0)
	for _, endpoint := range endpoints {
		handler, err := p.buildEndpoints(endpoint, p.middleware)
		if err != nil {
			return
		}
		routes = append(routes, router.NewRoute(endpoint.Predicates, handler))
	}
	p.router.LoadOrUpdateRoutes(routes)
}

func (p *Proxy) setXForwarded(req *http.Request) {
	clientIP, _, err := net.SplitHostPort(req.RemoteAddr)
	if err == nil {
		prior := req.Header["X-Forwarded-For"]
		if len(prior) > 0 {
			clientIP = strings.Join(prior, ", ") + ", " + clientIP
		}
		req.Header.Set("X-Forwarded-For", clientIP)
	} else {
		req.Header.Del("X-Forwarded-For")
	}
}
