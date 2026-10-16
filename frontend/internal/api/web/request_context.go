package web

import "context"

type requestKey struct{}

type requestInfo struct {
	id     string
	method string
	route  string
}

func withRequest(ctx context.Context, id, method, route string) context.Context {
	return context.WithValue(ctx, requestKey{}, requestInfo{id: id, method: method, route: route})
}

func buildRequestAttributes(ctx context.Context) []any {
	info, _ := ctx.Value(requestKey{}).(requestInfo)
	return []any{"request_id", info.id, "method", info.method, "route", info.route}
}
