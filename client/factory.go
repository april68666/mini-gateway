package client

import (
	"context"
	"crypto/tls"
	"golang.org/x/net/http2"
	"mini-gateway/config"
	"mini-gateway/selector"
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

		for _, uri := range endpoint.Uris {
			parse, err := url.Parse(uri)
			if err != nil {
				slog.Error(err.Error())
				continue
			}

			c := defaultHttpClient
			if endpoint.Protocol == "grpc" {
				c = defaultHttp2Client
			}

			node := selector.NewNode(parse.Scheme, parse.Host, endpoint.Protocol, c)
			nodes = append(nodes, node)
		}

		rotationSelector := selector.NewRotationSelector()
		rotationSelector.Update(nodes)

		return newClient(rotationSelector), nil
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
