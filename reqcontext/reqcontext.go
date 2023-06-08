package reqcontext

import (
	"context"
	"mini-gateway/config"
)

type contextKey string

func WithEndpoint(ctx context.Context, v *config.Endpoint) context.Context {
	return context.WithValue(ctx, contextKey("Endpoint"), v)
}

func Endpoint(ctx context.Context) *config.Endpoint {
	endpoint, _ := ctx.Value(contextKey("Endpoint")).(*config.Endpoint)
	return endpoint
}
