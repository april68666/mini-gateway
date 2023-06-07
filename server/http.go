package server

import (
	"context"
	"net"
	"net/http"
)

func NewHttpServer(proxy http.Handler) *HttpServer {
	return &HttpServer{handler: proxy}
}

type HttpServer struct {
	handler http.Handler
	h1s     *http.Server
}

func (s *HttpServer) Run(l net.Listener) error {
	return s.run(l)
}

func (s *HttpServer) run(l net.Listener) error {
	s.h1s = &http.Server{
		Handler: s.handler,
	}
	return s.h1s.Serve(l)
}

func (s *HttpServer) Shutdown(ctx context.Context) error {
	return s.h1s.Shutdown(ctx)
}
