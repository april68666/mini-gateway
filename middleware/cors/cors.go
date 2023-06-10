package cors

import (
	"mini-gateway/config"
	"mini-gateway/middleware"
	"net/http"
)

const NAME = "cors"

func init() {
	middleware.Register(NAME, Factory)
}

func Factory(c *config.Middleware) middleware.Middleware {
	allowOrigin := "*"
	allowHeaders := "Content-Type,AccessToken,X-CSRF-Token, Authorization, X-Auth-Token"
	allowMethod := "POST, GET, OPTIONS"
	exposeHeaders := "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Content-Type"
	Credentials := "true"

	args := make(map[string]string)
	for _, arg := range c.Args {
		args[arg.Key] = arg.Value

	}

	if v, ok := args["allow-origin"]; ok {
		allowOrigin = v
	}

	if v, ok := args["allow-headers"]; ok {
		allowOrigin = v
	}

	if v, ok := args["allow-method"]; ok {
		allowOrigin = v
	}

	if v, ok := args["expose-headers"]; ok {
		allowOrigin = v
	}

	if v, ok := args["credentials"]; ok {
		allowOrigin = v
	}

	return func(next http.RoundTripper) http.RoundTripper {
		return &cors{
			allowOrigin:   allowOrigin,
			allowHeaders:  allowHeaders,
			allowMethod:   allowMethod,
			exposeHeaders: exposeHeaders,
			Credentials:   Credentials,
			next:          next,
		}
	}
}

type cors struct {
	allowOrigin   string
	allowHeaders  string
	allowMethod   string
	exposeHeaders string
	Credentials   string
	next          http.RoundTripper
}

func (c *cors) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.Method == "OPTIONS" {
		header := http.Header{}
		header.Add("Access-Control-Allow-Origin", c.allowOrigin)
		header.Add("Access-Control-Allow-Headers", c.allowHeaders)
		header.Add("Access-Control-Allow-Methods", c.allowMethod)
		header.Add("Access-Control-Expose-Headers", c.exposeHeaders)
		header.Add("Access-Control-Allow-Credentials", c.Credentials)
		return &http.Response{
			Header:     header,
			StatusCode: http.StatusNoContent,
		}, nil
	}
	return c.next.RoundTrip(req)
}
