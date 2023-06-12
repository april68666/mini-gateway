package reqcontext

import (
	"context"
	"mini-gateway/config"
)

type contextKey string

func WithEndpoint(ctx context.Context, v *config.Endpoint) context.Context {
	return context.WithValue(ctx, contextKey("endpoint"), v)
}

func Endpoint(ctx context.Context) (*config.Endpoint, bool) {
	endpoint, b := ctx.Value(contextKey("endpoint")).(*config.Endpoint)
	return endpoint, b
}

func WithColor(ctx context.Context, color string) context.Context {
	return context.WithValue(ctx, contextKey("color"), color)
}

func Color(ctx context.Context) (string, bool) {
	color, b := ctx.Value(contextKey("color")).(string)
	return color, b
}

func WithParams(ctx context.Context, params map[string]string) context.Context {
	return context.WithValue(ctx, contextKey("params"), params)
}

func Params(ctx context.Context) (map[string]string, bool) {
	params, b := ctx.Value(contextKey("params")).(map[string]string)
	return params, b
}
