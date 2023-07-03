package color

import (
	"mini-gateway/config"
	"mini-gateway/middleware"
	"mini-gateway/reqcontext"
	"net/http"
)

const NAME = "color"

func init() {
	middleware.Register(NAME, Factory)
}

func Factory(c *config.Middleware) middleware.Middleware {
	fromHeaderKey := ""
	if v, ok := c.Args["fromHeaderKey"]; ok {
		fromHeaderKey = v.(string)
	}

	return func(next http.RoundTripper) http.RoundTripper {
		return &color{
			headerKey: fromHeaderKey,
			next:      next,
		}
	}
}

type color struct {
	headerKey string
	next      http.RoundTripper
}

func (c *color) RoundTrip(req *http.Request) (*http.Response, error) {
	_color := req.Header.Get(c.headerKey)
	ctx := reqcontext.WithColor(req.Context(), _color)
	return c.next.RoundTrip(req.WithContext(ctx))
}
