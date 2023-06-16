package proxy

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"mini-gateway/client"
	"mini-gateway/config"
	"mini-gateway/middleware"
	"mini-gateway/reqcontext"
	"mini-gateway/router"
	"mini-gateway/router/route"
	"mini-gateway/slog"
	"net"
	"net/http"
	"runtime"
	"strings"
	"sync"
	"time"
)

func NewProxy(factory client.Factory, router router.Router) *Proxy {
	return &Proxy{
		router:  router,
		factory: factory,
	}
}

type Proxy struct {
	router    router.Router
	factory   client.Factory
	mux       sync.Mutex
	routeInfo map[string]*routeInfo
}

type routeInfo struct {
	cancel   context.CancelFunc
	route    *route.Route
	endpoint *config.Endpoint
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			w.WriteHeader(http.StatusBadGateway)
			buf := make([]byte, 64<<10)
			n := runtime.Stack(buf, false)
			slog.Error("%s", buf[:n])
		}
	}()
	p.router.ServeHTTP(w, r)
}

// UpdateEndpoints 重新生成所有端点，全局的中间件有更新必须调用此方法重新生成端点否则不生效。
func (p *Proxy) UpdateEndpoints(globalMs []*config.Middleware, es []*config.Endpoint) (err error) {
	p.mux.Lock()
	defer p.mux.Unlock()
	rs := make([]*route.Route, 0)
	ris := make(map[string]*routeInfo)
	for _, e := range es {
		if _, ok := ris[e.ID]; ok {
			return errors.New(fmt.Sprintf("endpoint id cannot be the same,id:%s", e.ID))
		}
		ctx, cancel := context.WithCancel(context.Background())
		handler, err := p.buildEndpoints(ctx, globalMs, e)
		if err != nil {
			cancel()
			return err
		}
		r := route.NewRoute(e.Predicates, handler)
		ris[e.ID] = &routeInfo{
			route:    r,
			endpoint: e,
			cancel:   cancel,
		}
		rs = append(rs, r)
	}
	p.router.RegisterOrUpdateRoutes(rs)
	// 通知所有ctx取消
	for _, info := range p.routeInfo {
		info.cancel()
	}
	// 替换
	p.routeInfo = ris
	return nil
}

// UpdateEndpoint 重新生成单个端点，全局的中间件有更新必须调用 UpdateEndpoints 更新所有端点。
func (p *Proxy) UpdateEndpoint(globalMs []*config.Middleware, e *config.Endpoint) (err error) {
	p.mux.Lock()
	defer p.mux.Unlock()
	ctx, cancel := context.WithCancel(context.Background())
	handler, err := p.buildEndpoints(ctx, globalMs, e)
	if err != nil {
		cancel()
		return err
	}
	r := route.NewRoute(e.Predicates, handler)
	rs := make([]*route.Route, 0)
	for _, r := range p.routeInfo {
		rs = append(rs, r.route)
	}
	p.router.RegisterOrUpdateRoutes(rs)
	// 通知被替换的路由ctx取消
	p.routeInfo[e.ID].cancel()
	// 替换
	p.routeInfo[e.ID] = &routeInfo{
		cancel:   cancel,
		route:    r,
		endpoint: e,
	}
	return nil
}

func (p *Proxy) RemoveEndpoint(endpointID string) {
	p.mux.Lock()
	defer p.mux.Unlock()
	if _, ok := p.routeInfo[endpointID]; !ok {
		return
	}

	// 通知被删除的路由ctx取消
	p.routeInfo[endpointID].cancel()
	delete(p.routeInfo, endpointID)
	rs := make([]*route.Route, 0)
	for _, r := range p.routeInfo {
		rs = append(rs, r.route)
	}
	p.router.RegisterOrUpdateRoutes(rs)
}

func (p *Proxy) buildEndpoints(ctx context.Context, ms []*config.Middleware, endpoint *config.Endpoint) (http.Handler, error) {
	factory, err := p.factory(ctx, endpoint)
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
		ctx := reqcontext.WithEndpoint(req.Context(), endpoint)
		if endpoint.Timeout > 0 {
			_ctx, cancel := context.WithTimeout(ctx, time.Millisecond*time.Duration(endpoint.Timeout))
			defer cancel()
			ctx = _ctx
		}

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
		resp, err := tripper.RoundTrip(req.Clone(ctx))
		if err != nil {
			w.WriteHeader(http.StatusBadGateway)
			slog.Error("Endpoint id:%s,error:%s", endpoint.ID, err.Error())
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
			/*
				https://developer.mozilla.org/zh-CN/docs/Web/HTTP/Headers/Trailer
				https://pkg.go.dev/net/http#ResponseWriter
				https://pkg.go.dev/net/http#example-ResponseWriter-Trailers

				HTTP Trailers 是一种在 HTTP 报文中包含元数据的方式，通常用于在报文主体传输完毕后提供一些附加的信息。
				如果您在做代理时需要处理 HTTP Trailers，可以考虑使用以下步骤：
				等待原始请求的主体传输完毕。
				对接收到的每个数据块进行处理，并检查是否存在 Trailer 标头指定的 Trailer 名称。
				如果 Trailer 存在，则将 Trailer 名称和 Trailer 值存储在缓冲区中。
				在传递响应之前，将缓冲区中的 Trailer 值添加到响应中。
				需要注意的是，HTTP Trailers 的支持在不同的代理服务器上可能会有所不同，请确保您的代理服务器支持处理 HTTP Trailers。
				同时，也要注意 Trailer 的数量和大小，因为过多或过大的 Trailer 可能会对性能产生负面影响。

				示例

				HTTP/1.1 200 OK
				Content-Type: text/plain
				Transfer-Encoding: chunked
				Trailer: Expires

				7\r\n
				Mozilla\r\n
				9\r\n
				Developer\r\n
				7\r\n
				Network\r\n
				0\r\n
				Expires: Wed, 21 Oct 2015 07:28:00 GMT\r\n
				\r\n

			*/
			for k, v := range resp.Trailer {
				headers[http.TrailerPrefix+k] = v
			}
		}()
	})), nil
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
