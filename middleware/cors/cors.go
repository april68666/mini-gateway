package cors

import (
	"mini-gateway/config"
	"mini-gateway/middleware"
	"net/http"
	"strconv"
)

const NAME = "cors"

func init() {
	middleware.Register(NAME, Factory)
}

func Factory(c *config.Middleware) middleware.Middleware {
	allowOrigin := "*"
	allowHeaders := "*"
	allowMethod := "*"
	exposeHeaders := "*"
	credentials := "true"

	if v, ok := c.Args["allowOrigin"]; ok {
		allowOrigin = v.(string)
	}

	if v, ok := c.Args["allowHeaders"]; ok {
		allowHeaders = v.(string)
	}

	if v, ok := c.Args["allowMethod"]; ok {
		allowMethod = v.(string)
	}

	if v, ok := c.Args["exposeHeaders"]; ok {
		exposeHeaders = v.(string)
	}

	if v, ok := c.Args["credentials"]; ok {
		credentials = strconv.FormatBool(v.(bool))
	}

	return func(next http.RoundTripper) http.RoundTripper {
		return &cors{
			allowOrigin:   allowOrigin,
			allowHeaders:  allowHeaders,
			allowMethod:   allowMethod,
			exposeHeaders: exposeHeaders,
			credentials:   credentials,
			next:          next,
		}
	}
}

type cors struct {
	allowOrigin   string
	allowHeaders  string
	allowMethod   string
	exposeHeaders string
	credentials   string
	next          http.RoundTripper
}

func (c *cors) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.Method == "OPTIONS" {
		header := http.Header{}
		header.Add("Access-Control-Allow-Origin", c.allowOrigin)
		header.Add("Access-Control-Allow-Headers", c.allowHeaders)
		header.Add("Access-Control-Allow-Methods", c.allowMethod)
		header.Add("Access-Control-Expose-Headers", c.exposeHeaders)
		header.Add("Access-Control-Allow-Credentials", c.credentials)
		return &http.Response{
			Header:     header,
			StatusCode: http.StatusNoContent,
		}, nil
	}
	response, err := c.next.RoundTrip(req)
	if req.Header.Get("Sec-Fetch-Mode") == "cors" {
		response.Header.Set("Access-Control-Allow-Origin", "*")
	}

	return response, err
}
