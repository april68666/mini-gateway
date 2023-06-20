package main

import (
	"context"
	"flag"
	"fmt"
	"gopkg.in/yaml.v2"
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

	_ "mini-gateway/loadbalance/rotation"
	_ "mini-gateway/loadbalance/weight"
	_ "mini-gateway/middleware/cors"
	_ "mini-gateway/middleware/forwarding"
	_ "mini-gateway/middleware/jwt"
	_ "mini-gateway/middleware/logging"
	_ "mini-gateway/middleware/stripprefix"
)

var (
	configFile string
)

func init() {
	flag.StringVar(&configFile, "conf", "config.yaml", "config path, eg: -conf config.yaml")
}

func main() {
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()
		slog.Flush(ctx)
	}()
	flag.Parse()

	fileBytes, err := os.ReadFile(configFile)
	if err != nil {
		slog.Error(err.Error())
		return
	}
	var c *config.Gateway
	err = yaml.Unmarshal(fileBytes, &c)
	if err != nil {
		slog.Error("Error parsing YAML:", err)
		return
	}

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", c.Http.Port))
	if err != nil {
		slog.Fatal(err.Error())
		return
	}

	p := proxy.NewProxy(client.NewFactory(nil), router.NewDefaultRouter())
	err = p.UpdateEndpoints(c.Http.Middlewares, c.Http.Endpoints)
	if err != nil {
		slog.Fatal(err.Error())
		return
	}

	slog.Info(" Listening and serving HTTP on %s", listener.Addr().String())
	serv := server.NewHttpServer(p)
	go func() {
		err := serv.Run(listener)
		if err != nil && err != http.ErrServerClosed {
			slog.Fatal(err.Error())
			return
		}
	}()

	go func() {
		fileInfo, err := os.Stat(configFile)
		if err != nil {
			slog.Error(err.Error())
			return
		}
		lastModifiedTime := fileInfo.ModTime()
		for {
			time.Sleep(1 * time.Second)
			fileInfo, err := os.Stat(configFile)
			if err != nil {
				slog.Error(err.Error())
				continue
			}
			if fileInfo.ModTime() != lastModifiedTime {
				slog.Info("Config file %s has been modified", configFile)
				lastModifiedTime = fileInfo.ModTime()

				fileBytes, err := os.ReadFile(configFile)
				if err != nil {
					slog.Error(err.Error())
					continue
				}
				c = &config.Gateway{}
				err = yaml.Unmarshal(fileBytes, &c)
				if err != nil {
					slog.Error("Error parsing YAML:", err)
					continue
				}
				err = p.UpdateEndpoints(c.Http.Middlewares, c.Http.Endpoints)
				if err != nil {
					slog.Error(err.Error())
					return
				}
			}
		}
	}()

	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)
	<-quit
	slog.Info("Shutdown Server ...")
	if err := serv.Shutdown(context.Background()); err != nil {
		slog.Fatal("Server Shutdown:", err)
	}

	slog.Info("Server exiting")
}
