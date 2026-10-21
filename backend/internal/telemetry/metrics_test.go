package telemetry

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	for _, tc := range []struct {
		name, accept string
	}{
		{name: "missing Accept"},
		{name: "Prometheus Accept", accept: "text/plain; version=0.0.4"},
		{name: "OTLP Accept", accept: "application/x-protobuf"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			first, err := New()
			require.NoError(t, err)
			t.Cleanup(func() { require.NoError(t, first.Provider.Shutdown(context.Background())) })

			second, err := New()
			require.NoError(t, err)
			t.Cleanup(func() { require.NoError(t, second.Provider.Shutdown(context.Background())) })

			counter, err := first.Provider.Meter("isolation").Int64Counter("bookshop_test_requests")
			require.NoError(t, err)
			counter.Add(context.Background(), 1)

			request := httptest.NewRequest(http.MethodGet, "/metrics", nil)
			request.Header.Set("Accept", tc.accept)

			response := httptest.NewRecorder()
			first.Handler.ServeHTTP(response, request)
			require.Equal(t, http.StatusOK, response.Code)
			require.Equal(t, "application/openmetrics-text; version=1.0.0; charset=utf-8", response.Header().Get("Content-Type"))
			require.Contains(t, response.Body.String(), "bookshop_test_requests_total")
			require.Contains(t, response.Body.String(), "go_goroutines")
			require.True(t, strings.HasSuffix(response.Body.String(), "# EOF\n"))

			other := httptest.NewRecorder()
			second.Handler.ServeHTTP(other, request)
			require.NotContains(t, other.Body.String(), "bookshop_test_requests_total")
		})
	}
}
