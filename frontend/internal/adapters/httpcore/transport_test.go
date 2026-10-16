package httpcore

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRequestIDTransportRoundTrip(t *testing.T) {
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "51aa5ea3-0653-49ad-a3b5-e4c3b7ade70c", r.Header.Get("X-Request-ID"))
		w.WriteHeader(http.StatusNoContent)
	}))
	defer api.Close()
	transport := requestIDTransport{next: http.DefaultTransport.(*http.Transport).Clone()}
	defer transport.CloseIdleConnections()
	ctx := WithRequestID(context.Background(), "51aa5ea3-0653-49ad-a3b5-e4c3b7ade70c")
	req, err := http.NewRequestWithContext(ctx, "GET", api.URL, nil)
	require.NoError(t, err)
	req.Header.Set("X-Request-ID", "untrusted")
	client := http.Client{Transport: transport}
	resp, err := client.Do(req)
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())
	require.Equal(t, "untrusted", req.Header.Get("X-Request-ID"))
}
