package requestcontext

import "context"

type key struct{}

type request struct {
	id     string
	method string
	route  string
}

func WithRequest(ctx context.Context, id, method, route string) context.Context {
	return context.WithValue(ctx, key{}, request{id: id, method: method, route: route})
}

func BuildAttributes(ctx context.Context) []any {
	info, _ := ctx.Value(key{}).(request)
	return []any{"request_id", info.id, "method", info.method, "route", info.route}
}
