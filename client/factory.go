package client

import (
	"context"
	"crypto/tls"
	"golang.org/x/net/http2"
	"mini-gateway/config"
	"mini-gateway/discovery"
	"mini-gateway/selector"
	"mini-gateway/selector/weight"
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

		f, ok := selector.Get(endpoint.LoadBalance)
		if !ok {
			slog.Warn("could not find load balancer selector %s,rotation is used by default", endpoint.LoadBalance)
			f = weight.Factor
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
	Transport: &http.Transport{
		MaxIdleConns:        0,
		MaxIdleConnsPerHost: 10000,
		MaxConnsPerHost:     10000,
		DisableCompression:  true,
	},
}

var defaultHttp2Client = &http.Client{
	Transport: &http2.Transport{
		DisableCompression: true,
		AllowHTTP:          true,
		DialTLSContext: func(ctx context.Context, network, addr string, cfg *tls.Config) (net.Conn, error) {
			return net.DialTimeout(network, addr, 300*time.Millisecond)
		},
	},
}
