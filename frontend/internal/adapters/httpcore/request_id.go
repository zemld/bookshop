package httpcore

import (
	"fmt"
	"net/http"

	"bookshop/frontend/internal/logcontext"
)

type requestIDTransport struct {
	next http.RoundTripper
}

func (t requestIDTransport) CloseIdleConnections() {
	if transport, ok := t.next.(interface{ CloseIdleConnections() }); ok {
		transport.CloseIdleConnections()
	}
}

func (t requestIDTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	id := logcontext.ReadID(r.Context())
	if id != "" {
		request := r.Clone(r.Context())
		request.Header.Set("X-Request-ID", id)
		r = request
	}
	response, err := t.next.RoundTrip(r)
	if err != nil {
		return nil, fmt.Errorf("send API request: %w", err)
	}
	return response, nil
}
