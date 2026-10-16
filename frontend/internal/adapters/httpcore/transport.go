package httpcore

import "net/http"

type requestIDTransport struct {
	next http.RoundTripper
}

func (t requestIDTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	req := r.Clone(r.Context())
	if id := getRequestID(r.Context()); id != "" {
		req.Header.Set("X-Request-ID", id)
	}
	return t.next.RoundTrip(req)
}

func (t requestIDTransport) CloseIdleConnections() {
	if transport, ok := t.next.(interface{ CloseIdleConnections() }); ok {
		transport.CloseIdleConnections()
	}
}
