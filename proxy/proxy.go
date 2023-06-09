package proxy

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"golang.org/x/net/http/httpguts"
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
	"net/http/httptrace"
	"net/textproto"
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
	cancelCtx context.CancelFunc
	route     *route.Route
	endpoint  *config.Endpoint
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			errorHandler(w, r, err.(error))
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
			route:     r,
			endpoint:  e,
			cancelCtx: cancel,
		}
		rs = append(rs, r)
	}
	p.router.RegisterOrUpdateRoutes(rs)
	// 通知所有ctx取消
	for _, info := range p.routeInfo {
		info.cancelCtx()
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
	p.routeInfo[e.ID].cancelCtx()
	// 替换
	p.routeInfo[e.ID] = &routeInfo{
		cancelCtx: cancel,
		route:     r,
		endpoint:  e,
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
	p.routeInfo[endpointID].cancelCtx()
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

	// https://github.com/golang/go/blob/98617fd23fa799173c33741987d41ee64cbb2a4f/src/net/http/httputil/reverseproxy.go#L332
	return http.Handler(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		ctx := reqcontext.WithEndpoint(req.Context(), endpoint)
		if endpoint.Timeout > 0 {
			_ctx, cancel := context.WithTimeout(ctx, time.Millisecond*time.Duration(endpoint.Timeout))
			defer cancel()
			ctx = _ctx
		}
		outReq := req.Clone(ctx)

		if req.ContentLength == 0 {
			outReq.Body = nil
		}
		if outReq.Body != nil {
			defer outReq.Body.Close()
		}

		outReq.Close = false

		reqUpType := upgradeType(outReq.Header)
		if !IsPrint(reqUpType) {
			err := fmt.Errorf("client tried to switch to invalid protocol %q", reqUpType)
			slog.Error(err.Error())
			errorHandler(rw, req, err)
			return
		}
		removeHopByHopHeaders(outReq.Header)

		// 兼容 grpc 处理 https://github.com/golang/go/issues/21096
		if httpguts.HeaderValuesContainsToken(req.Header["Te"], "trailers") {
			outReq.Header.Set("Te", "trailers")
		}

		// 剥离逐跳头后添加协议升级信息比如 websockets
		if reqUpType != "" {
			outReq.Header.Set("Connection", "Upgrade")
			outReq.Header.Set("Upgrade", reqUpType)
		}

		p.setXForwarded(outReq)

		if _, ok := outReq.Header["User-Agent"]; !ok {
			outReq.Header.Set("User-Agent", "")
		}

		if outReq.Body != nil {
			body, err := io.ReadAll(outReq.Body)
			if err != nil {
				slog.Error(err.Error())
				errorHandler(rw, req, err)
				return
			}

			outReq.GetBody = func() (io.ReadCloser, error) {
				reader := bytes.NewReader(body)
				return io.NopCloser(reader), nil
			}

			reader := bytes.NewReader(body)
			outReq.Body = io.NopCloser(reader)
		}

		trace := &httptrace.ClientTrace{
			Got1xxResponse: func(code int, header textproto.MIMEHeader) error {
				h := rw.Header()
				copyHeader(h, http.Header(header))
				rw.WriteHeader(code)

				// Clear headers, it's not automatically done by ResponseWriter.WriteHeader() for 1xx responses
				for k := range h {
					delete(h, k)
				}

				return nil
			},
		}
		outReq = outReq.WithContext(httptrace.WithClientTrace(outReq.Context(), trace))

		res, err := tripper.RoundTrip(outReq)
		if err != nil {
			slog.Error("Endpoint id:%s,error:%s", endpoint.ID, err.Error())
			errorHandler(rw, req, err)
			return
		}

		if res.StatusCode == http.StatusSwitchingProtocols {
			p.handleUpgradeResponse(rw, outReq, res)
			return
		}

		removeHopByHopHeaders(res.Header)

		copyHeader(rw.Header(), res.Header)

		announcedTrailers := len(res.Trailer)
		if announcedTrailers > 0 {
			trailerKeys := make([]string, 0, len(res.Trailer))
			for k := range res.Trailer {
				trailerKeys = append(trailerKeys, k)
			}
			rw.Header().Add("Trailer", strings.Join(trailerKeys, ", "))
		}

		rw.WriteHeader(res.StatusCode)

		if res.Body != nil {
			_, err = io.Copy(rw, res.Body)
			if err != nil {
				defer res.Body.Close()
				slog.Error(err.Error())
				errorHandler(rw, req, err)
				return
			}
			res.Body.Close()
		}

		if len(res.Trailer) > 0 {
			err := http.NewResponseController(rw).Flush()
			if err != nil {
				slog.Error(err.Error())
				errorHandler(rw, req, err)
				return
			}
		}

		if len(res.Trailer) == announcedTrailers {
			copyHeader(rw.Header(), res.Trailer)
			return
		}

		for k, vv := range res.Trailer {
			k = http.TrailerPrefix + k
			for _, v := range vv {
				rw.Header().Add(k, v)
			}
		}

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
	req.Header.Set("X-Forwarded-Host", req.Host)
	if req.TLS == nil {
		req.Header.Set("X-Forwarded-Proto", "http")
	} else {
		req.Header.Set("X-Forwarded-Proto", "https")
	}
}

func (p *Proxy) handleUpgradeResponse(rw http.ResponseWriter, req *http.Request, res *http.Response) {
	reqUpType := upgradeType(req.Header)
	resUpType := upgradeType(res.Header)
	if !IsPrint(resUpType) {
		err := fmt.Errorf("backend tried to switch to invalid protocol %q", resUpType)
		slog.Error(err.Error())
		errorHandler(rw, req, err)
	}
	if !EqualFold(reqUpType, resUpType) {
		err := fmt.Errorf("backend tried to switch protocol %q when %q was requested", resUpType, reqUpType)
		slog.Error(err.Error())
		errorHandler(rw, req, err)
		return
	}

	hj, ok := rw.(http.Hijacker)
	if !ok {
		err := fmt.Errorf("can't switch protocols using non-Hijacker ResponseWriter type %T", rw)
		slog.Error(err.Error())
		errorHandler(rw, req, err)
		return
	}
	backConn, ok := res.Body.(io.ReadWriteCloser)
	if !ok {
		err := fmt.Errorf("internal error: 101 switching protocols response with non-writable body")
		slog.Error(err.Error())
		errorHandler(rw, req, err)
		return
	}

	backConnCloseCh := make(chan bool)
	go func() {
		// Ensure that the cancellation of a request closes the backend.
		// See issue https://golang.org/issue/35559.
		select {
		// 如果req设置了超时时间会被取消ws连接
		//case <-req.Context().Done():
		case <-backConnCloseCh:
		}
		backConn.Close()
	}()

	defer close(backConnCloseCh)

	conn, brw, err := hj.Hijack()
	if err != nil {
		slog.Error(err.Error())
		errorHandler(rw, req, err)
		return
	}
	defer conn.Close()

	copyHeader(rw.Header(), res.Header)

	res.Header = rw.Header()
	res.Body = nil
	if err := res.Write(brw); err != nil {
		slog.Error(err.Error())
		errorHandler(rw, req, err)
		return
	}
	if err := brw.Flush(); err != nil {
		slog.Error(err.Error())
		errorHandler(rw, req, err)
		return
	}
	errc := make(chan error, 1)
	spc := switchProtocolCopier{user: conn, backend: backConn}
	go spc.copyToBackend(errc)
	go spc.copyFromBackend(errc)
	<-errc
}

// switchProtocolCopier exists so goroutines proxying data back and
// forth have nice names in stacks.
type switchProtocolCopier struct {
	user, backend io.ReadWriter
}

func (c switchProtocolCopier) copyFromBackend(errc chan<- error) {
	_, err := io.Copy(c.user, c.backend)
	errc <- err
}

func (c switchProtocolCopier) copyToBackend(errc chan<- error) {
	_, err := io.Copy(c.backend, c.user)
	errc <- err
}

func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

// Hop-by-hop headers. These are removed when sent to the backend.
// As of RFC 7230, hop-by-hop headers are required to appear in the
// Connection header field. These are the headers defined by the
// obsoleted RFC 2616 (section 13.5.1) and are used for backward
// compatibility.
var hopHeaders = []string{
	"Connection",
	"Proxy-Connection", // non-standard but still sent by libcurl and rejected by e.g. google
	"Keep-Alive",
	"Proxy-Authenticate",
	"Proxy-Authorization",
	"Te",      // canonicalized version of "TE"
	"Trailer", // not Trailers per URL above; https://www.rfc-editor.org/errata_search.php?eid=4522
	"Transfer-Encoding",
	"Upgrade",
}

func removeHopByHopHeaders(h http.Header) {
	for _, f := range h["Connection"] {
		for _, sf := range strings.Split(f, ",") {
			if sf = textproto.TrimString(sf); sf != "" {
				h.Del(sf)
			}
		}
	}

	for _, f := range hopHeaders {
		h.Del(f)
	}
}

func upgradeType(h http.Header) string {
	if !httpguts.HeaderValuesContainsToken(h["Connection"], "Upgrade") {
		return ""
	}
	return h.Get("Upgrade")
}

// IsPrint returns whether s is ASCII and printable according to
// https://tools.ietf.org/html/rfc20#section-4.2.
func IsPrint(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] < ' ' || s[i] > '~' {
			return false
		}
	}
	return true
}

// EqualFold is strings.EqualFold, ASCII only. It reports whether s and t
// are equal, ASCII-case-insensitively.
func EqualFold(s, t string) bool {
	if len(s) != len(t) {
		return false
	}
	for i := 0; i < len(s); i++ {
		if lower(s[i]) != lower(t[i]) {
			return false
		}
	}
	return true
}

// lower returns the ASCII lowercase version of b.
func lower(b byte) byte {
	if 'A' <= b && b <= 'Z' {
		return b + ('a' - 'A')
	}
	return b
}

func errorHandler(rw http.ResponseWriter, req *http.Request, err error) {
	httpStatus := http.StatusBadGateway
	switch {
	case errors.Is(err, context.Canceled):
		httpStatus = 499
	case errors.Is(err, context.DeadlineExceeded):
		httpStatus = http.StatusGatewayTimeout
	}
	http.Error(rw, http.StatusText(httpStatus), httpStatus)
}
