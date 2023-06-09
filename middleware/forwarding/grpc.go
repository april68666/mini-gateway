package forwarding

import (
	"bytes"
	"encoding/binary"
	"io"
	"mini-gateway/config"
	"mini-gateway/middleware"
	"mini-gateway/reqcontext"
	"mini-gateway/slog"
	"net/http"
	"strconv"
	"strings"
)

const NAME = "grpc"

func init() {
	middleware.Register(NAME, Factory)
}

func Factory(c *config.Middleware) middleware.Middleware {
	httpStatus := 400
	errorTemplate := "{\"code\": {status},\"message\": \"{message}\"}"
	args := make(map[string]string)
	for _, arg := range c.Args {
		args[arg.Key] = arg.Value

	}

	if v, ok := args["http_status"]; ok {
		status, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			slog.Error("grpc middleware  http_status parse error:%s", err.Error())
		}
		httpStatus = int(status)
	}
	if v, ok := args["grpc_error_template"]; ok {
		errorTemplate = v
	}
	return func(next http.RoundTripper) http.RoundTripper {
		return &grpc{next: next, errorTemplate: errorTemplate, httpStatus: httpStatus}
	}
}

type grpc struct {
	next          http.RoundTripper
	errorTemplate string
	httpStatus    int
}

func (g *grpc) RoundTrip(req *http.Request) (*http.Response, error) {
	contentType := req.Header.Get("Content-Type")
	endpoint, _ := reqcontext.Endpoint(req.Context())
	if (endpoint != nil && strings.ToLower(endpoint.Protocol) != "grpc") || strings.HasSuffix(contentType, "application/grpc") {
		return g.next.RoundTrip(req)
	}
	bodyByte, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}
	// grpc 业务数据包头5字节,第一个字节是否压缩，后4字节消息长度
	// https://github.com/grpc/grpc/blob/master/doc/PROTOCOL-HTTP2.md
	grpcBodyByte := make([]byte, len(bodyByte)+5)
	binary.BigEndian.PutUint32(grpcBodyByte[1:], uint32(len(bodyByte)))
	copy(grpcBodyByte[5:], bodyByte)

	protocol := strings.TrimLeft(contentType, "application/")
	if index := strings.Index(protocol, ";"); index != -1 {
		protocol = protocol[:index]
	}
	req.Header.Set("Content-Type", "application/grpc+"+protocol)
	req.Header.Del("Content-Length")
	req.ContentLength = int64(len(grpcBodyByte))
	req.Body = io.NopCloser(bytes.NewReader(grpcBodyByte))

	resp, err := g.next.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	resData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	/*
		在 gRPC 中，Trailers 用于在响应中包含一些元数据。与 HTTP Trailers 类似，
		gRPC Trailers 也是在响应主体传输结束后发送的。
		在 gRPC 中，Trailers 通常用于传递状态码、错误信息和其他元数据。
		当服务器向客户端发送响应时，可以使用 grpc-status、grpc-message、grpc-status-details-bin
		元数据键来设置状态码和错误信息。
	*/
	for k, v := range resp.Trailer {
		resp.Header[k] = v
	}
	resp.Trailer = nil
	resp.Header.Set("Content-Type", contentType)

	if grpcStatus := resp.Header.Get("Grpc-Status"); grpcStatus != "0" {
		data := strings.ReplaceAll(g.errorTemplate, "{status}", grpcStatus)
		data = strings.ReplaceAll(data, "{message}", resp.Header.Get("Grpc-Message"))
		return &http.Response{
			Body:          io.NopCloser(bytes.NewReader([]byte(data))),
			Header:        resp.Header,
			StatusCode:    g.httpStatus,
			ContentLength: int64(len(data)),
		}, nil
	}

	resp.Body = io.NopCloser(bytes.NewReader(resData[5:]))
	resp.ContentLength = int64(len(resData) - 5)
	resp.Header.Del("Content-Length")
	return resp, nil
}
