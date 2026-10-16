package api

import (
	"context"
	"net/http"
)

type requestIDKey struct{}

func StoreRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey{}, id)
}

func GetRequestID(ctx context.Context) string {
	id, _ := ctx.Value(requestIDKey{}).(string)
	return id
}

type requestIDTransport struct {
	next http.RoundTripper
}

func (t requestIDTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	req := r.Clone(r.Context())
	if id := GetRequestID(r.Context()); id != "" {
		req.Header.Set("X-Request-ID", id)
	}
	return t.next.RoundTrip(req)
}
