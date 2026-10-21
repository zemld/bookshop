package web

import (
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func CreateMetricsHandler(registry *prometheus.Registry) http.Handler {
	export := promhttp.HandlerFor(registry, promhttp.HandlerOpts{EnableOpenMetrics: true})
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		request := r.Clone(r.Context())
		request.Header.Set("Accept", "application/openmetrics-text; version=1.0.0")
		export.ServeHTTP(&openMetricsWriter{ResponseWriter: w}, request)
	})
}

type openMetricsWriter struct {
	http.ResponseWriter
}

func (w *openMetricsWriter) WriteHeader(status int) {
	if status == http.StatusOK {
		w.Header().Set("Content-Type", "application/openmetrics-text; version=1.0.0; charset=utf-8")
	}
	w.ResponseWriter.WriteHeader(status)
}

func (w *openMetricsWriter) Write(body []byte) (int, error) {
	w.Header().Set("Content-Type", "application/openmetrics-text; version=1.0.0; charset=utf-8")
	n, err := w.ResponseWriter.Write(body)
	if err != nil {
		return n, fmt.Errorf("write OpenMetrics response: %w", err)
	}
	return n, nil
}
