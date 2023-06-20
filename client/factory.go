package client

import (
	"context"
	"crypto/tls"
	"errors"
	"golang.org/x/net/http2"
	"mini-gateway/config"
	"mini-gateway/discovery"
	"mini-gateway/loadbalance"
	"mini-gateway/loadbalance/rotation"
	"mini-gateway/slog"
	"net"
	"net/http"
	"strings"
	"time"
)

type Factory func(ctx context.Context, endpoint *config.Endpoint) (http.RoundTripper, error)

func NewFactory(resolver discovery.Resolver) Factory {
	return func(ctx context.Context, endpoint *config.Endpoint) (http.RoundTripper, error) {
		c := defaultHttpClient
		if strings.ToLower(endpoint.Protocol) == "grpc" {
			c = defaultHttp2Client
		}

		f, ok := loadbalance.GetPicker(endpoint.LoadBalance)
		if !ok {
			slog.Warn("could not find load balancer picker %s,rotation is used by default", endpoint.LoadBalance)
			f = rotation.Factor
		}
		s := f()

		if resolver != nil && len(strings.TrimSpace(endpoint.Discovery)) > 0 {
			result, err := resolver.Resolve(context.Background(), endpoint.Discovery)
			if err != nil {
				return nil, err
			}
			s.Apply(result.Nodes)

			err = resolver.Watch(ctx, endpoint.Discovery, func(result *discovery.Result) {
				s.Apply(result.Nodes)
			})
			if err != nil {
				return nil, err
			}
		} else {
			ns := make([]discovery.Node, 0)
			for _, target := range endpoint.Targets {
				ns = append(ns, discovery.NewNode(target.Uri, target.Weight, target.Tags))
			}
			s.Apply(ns)
		}
		return newClient(s, c), nil
	}
}

var defaultHttpClient = &http.Client{
	CheckRedirect: defaultCheckRedirect,
	Transport: &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: defaultTransportDialContext(&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}),
		MaxIdleConns:          0,
		MaxIdleConnsPerHost:   10000,
		MaxConnsPerHost:       10000,
		DisableCompression:    true,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	},
}

var defaultHttp2Client = &http.Client{
	CheckRedirect: defaultCheckRedirect,
	Transport: &http2.Transport{
		DialTLSContext: func(ctx context.Context, network, addr string, cfg *tls.Config) (net.Conn, error) {
			return net.DialTimeout(network, addr, 30*time.Millisecond)
		},
		DisableCompression: true,
		AllowHTTP:          true,
	},
}

// https://github.com/golang/go/blob/bc21d6a4fcf2c957a3f279fa8725e16df6586864/src/net/http/client.go#LL690C36-L690C36
func defaultCheckRedirect(req *http.Request, via []*http.Request) error {
	if len(via) >= 10 {
		return errors.New("stopped after 10 redirects")
	}
	return http.ErrUseLastResponse
}

// https://github.com/golang/go/blob/98617fd23fa799173c33741987d41ee64cbb2a4f/src/net/http/transport.go#L43
func defaultTransportDialContext(dialer *net.Dialer) func(context.Context, string, string) (net.Conn, error) {
	return dialer.DialContext
}
