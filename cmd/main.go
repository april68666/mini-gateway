package main

import (
	"context"
	"fmt"
	"mini-gateway/client"
	"mini-gateway/config"
	"mini-gateway/proxy"
	"mini-gateway/router"
	"mini-gateway/server"
	"mini-gateway/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()
		slog.Flush(ctx)
	}()
	c := config.Gateway{
		Port: 8080,
		Middlewares: []*config.Middleware{
			{
				Name:  "logging",
				Order: 0,
				Args:  nil,
			},
		},
		Endpoints: []*config.Endpoint{{
			Uris:     []string{"http://127.0.0.1:8000", "http://127.0.0.1:8001", "http://127.0.0.1:8002"},
			Protocol: "http",
			Timeout:  2000,
			Predicates: &config.Predicates{
				Path:   "ping",
				Method: "GET",
			},
			Middlewares: nil,
		}},
	}
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", c.Port))
	if err != nil {
		slog.Fatal(err.Error())
		return
	}
	slog.Info(" Listening and serving HTTP on %s", listener.Addr().String())
	p := proxy.NewProxy(client.NewFactory(), router.NewDefaultRouter(), c.Middlewares)
	p.LoadOrUpdateEndpoints(c.Endpoints)
	serv := server.NewHttpServer(p)
	go func() {
		err := serv.Run(listener)
		if err != nil && err != http.ErrServerClosed {
			slog.Fatal(err.Error())
			return
		}
	}()

	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)
	<-quit
	slog.Info("Shutdown Server ...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := serv.Shutdown(ctx); err != nil {
		slog.Fatal("Server Shutdown:", err)
	}
	select {
	case <-ctx.Done():
		slog.Info("Timeout of 5 seconds")
	}
	slog.Info("Server exiting")
}
