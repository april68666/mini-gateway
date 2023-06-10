package main

import (
	"context"
	"encoding/json"
	"flag"
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

	_ "mini-gateway/middleware/cors"
	_ "mini-gateway/middleware/forwarding"
	_ "mini-gateway/middleware/jwt"
	_ "mini-gateway/middleware/logging"
	_ "mini-gateway/middleware/stripprefix"
	_ "mini-gateway/selector/rotation"
	_ "mini-gateway/selector/weight"
)

var (
	configFile string
)

func init() {
	flag.StringVar(&configFile, "conf", "config.json", "config path, eg: -conf config.json")
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
	err = json.Unmarshal(fileBytes, &c)
	if err != nil {
		slog.Error("Error parsing JSON:", err)
		return
	}

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", c.Port))
	if err != nil {
		slog.Fatal(err.Error())
		return
	}
	slog.Info(" Listening and serving HTTP on %s", listener.Addr().String())
	p := proxy.NewProxy(client.NewFactory(), router.NewDefaultRouter())
	p.LoadOrUpdateEndpoints(c)
	serv := server.NewHttpServer(p)
	go func() {
		err := serv.Run(listener)
		if err != nil && err != http.ErrServerClosed {
			slog.Fatal(err.Error())
			return
		}
	}()

	exit := make(chan struct{})
	go func() {
		fileInfo, err := os.Stat(configFile)
		if err != nil {
			slog.Error(err.Error())
			return
		}
		lastModifiedTime := fileInfo.ModTime()

		for {
			time.Sleep(1 * time.Second)
			select {
			case <-exit:
				return
			default:
				fileInfo, err := os.Stat(configFile)
				if err != nil {
					slog.Error(err.Error())
					continue
				}
				if fileInfo.ModTime() != lastModifiedTime {
					slog.Info("File %s has been modified", configFile)
					lastModifiedTime = fileInfo.ModTime()

					fileBytes, err := os.ReadFile(configFile)
					if err != nil {
						slog.Error(err.Error())
						continue
					}
					c = &config.Gateway{}
					err = json.Unmarshal(fileBytes, &c)
					if err != nil {
						slog.Error("Error parsing JSON:", err)
						continue
					}
					p.LoadOrUpdateEndpoints(c)
				}
			}
		}
	}()

	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)
	<-quit
	exit <- struct{}{}
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
