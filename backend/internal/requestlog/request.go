package requestlog

import (
	"context"
)

const Header = "X-Request-ID"

type contextKey struct{}

type Details struct {
	ID     string
	Method string
	Route  string
}

func WithDetails(ctx context.Context, details Details) context.Context {
	return context.WithValue(ctx, contextKey{}, details)
}

func GetDetails(ctx context.Context) Details {
	details, _ := ctx.Value(contextKey{}).(Details)
	return details
}

func BuildAttributes(ctx context.Context) []any {
	details := GetDetails(ctx)
	return []any{"request_id", details.ID, "method", details.Method, "route", details.Route}
}
