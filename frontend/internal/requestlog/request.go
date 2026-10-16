package requestlog

import (
	"context"
	"net/http"
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

type Transport struct {
	Next http.RoundTripper
}

func (t Transport) RoundTrip(r *http.Request) (*http.Response, error) {
	req := r.Clone(r.Context())
	if id := GetDetails(r.Context()).ID; id != "" {
		req.Header.Set(Header, id)
	}
	return t.Next.RoundTrip(req)
}

func (t Transport) CloseIdleConnections() {
	if transport, ok := t.Next.(interface{ CloseIdleConnections() }); ok {
		transport.CloseIdleConnections()
	}
}
