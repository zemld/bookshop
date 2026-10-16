package requestlog

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTransportRoundTrip(t *testing.T) {
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "51aa5ea3-0653-49ad-a3b5-e4c3b7ade70c", r.Header.Get(Header))
		w.WriteHeader(http.StatusNoContent)
	}))
	defer api.Close()
	transport := Transport{Next: http.DefaultTransport.(*http.Transport).Clone()}
	defer transport.CloseIdleConnections()
	ctx := WithDetails(context.Background(), Details{ID: "51aa5ea3-0653-49ad-a3b5-e4c3b7ade70c"})
	req, err := http.NewRequestWithContext(ctx, "GET", api.URL, nil)
	require.NoError(t, err)
	req.Header.Set(Header, "untrusted")
	client := http.Client{Transport: transport}
	resp, err := client.Do(req)
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())
	require.Equal(t, "untrusted", req.Header.Get(Header))
}
