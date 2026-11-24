package web

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
)

func TestReportVersionAndCountCatalogRequests(t *testing.T) {
	requests := prometheus.NewCounter(prometheus.CounterOpts{Name: "catalog_requests_total"})
	registry := prometheus.NewRegistry()
	require.NoError(t, registry.Register(requests))
	handler := NewVersionedHandler(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}), "v2", requests)

	for _, tc := range []struct {
		name   string
		method string
		path   string
		count  float64
	}{
		{"readiness probe", http.MethodGet, "/readyz", 0},
		{"different path", http.MethodGet, "/publishers", 0},
		{"different method", http.MethodPost, "/books", 0},
		{"catalog request", http.MethodGet, "/books", 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			request := httptest.NewRequest(tc.method, tc.path, nil)
			response := httptest.NewRecorder()

			handler.ServeHTTP(response, request)

			require.Equal(t, http.StatusOK, response.Code)
			require.Equal(t, "v2", response.Header().Get("X-Bookshop-Version"))
			metrics, err := registry.Gather()
			require.NoError(t, err)
			require.Len(t, metrics, 1)
			require.Len(t, metrics[0].GetMetric(), 1)
			require.Equal(t, tc.count, metrics[0].GetMetric()[0].GetCounter().GetValue())
		})
	}
}
