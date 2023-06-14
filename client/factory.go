package client

import (
	"context"
	"crypto/tls"
	"errors"
	"golang.org/x/net/http2"
	"mini-gateway/config"
	"mini-gateway/selector"
	"mini-gateway/selector/rotation"
	"mini-gateway/slog"
	"net"
	"net/http"
	"net/url"
	"time"
)

type Factory func(endpoint *config.Endpoint) (http.RoundTripper, error)

func NewFactory() Factory {
	return func(endpoint *config.Endpoint) (http.RoundTripper, error) {
		nodes := make([]*selector.Node, 0)
		for _, target := range endpoint.Targets {
			parse, err := url.Parse(target.Uri)
			if err != nil {
				return nil, errors.New(err.Error())
			}

			c := defaultHttpClient
			if endpoint.Protocol == "grpc" {
				c = defaultHttp2Client
			}

			node := selector.NewNode(parse.Scheme, parse.Host, endpoint.Protocol, target.Color, target.Weight, c)
			nodes = append(nodes, node)
		}

		f, ok := selector.Get(endpoint.LoadBalance)
		if !ok {
			slog.Warn("could not find load balancer selector %s,rotation is used by default", endpoint.LoadBalance)
			f = rotation.Factor
		}
		s := f()
		s.Update(nodes)
		return newClient(s), nil
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
