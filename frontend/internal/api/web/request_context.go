package web

import (
	"context"

	"bookshop/frontend/internal/adapters/api"
)

type requestKey struct{}

type requestInfo struct {
	method string
	route  string
}

func attachRequest(ctx context.Context, id, method, route string) context.Context {
	return context.WithValue(api.StoreRequestID(ctx, id), requestKey{}, requestInfo{method: method, route: route})
}

func buildRequestAttributes(ctx context.Context) []any {
	info, _ := ctx.Value(requestKey{}).(requestInfo)
	return []any{"request_id", api.GetRequestID(ctx), "method", info.method, "route", info.route}
}
