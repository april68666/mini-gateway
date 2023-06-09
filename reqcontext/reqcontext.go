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
