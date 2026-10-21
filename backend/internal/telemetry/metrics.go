package telemetry

import (
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	otelexport "go.opentelemetry.io/otel/exporters/prometheus"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
)

type Metrics struct {
	Provider *sdkmetric.MeterProvider
	Handler  http.Handler
}

func New() (*Metrics, error) {
	registry := prometheus.NewRegistry()
	if err := registry.Register(collectors.NewGoCollector()); err != nil {
		return nil, fmt.Errorf("register Go collector: %w", err)
	}

	if err := registry.Register(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{})); err != nil {
		return nil, fmt.Errorf("register process collector: %w", err)
	}

	exporter, err := otelexport.New(otelexport.WithRegisterer(registry))
	if err != nil {
		return nil, fmt.Errorf("initialize Prometheus exporter: %w", err)
	}

	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(exporter))
	export := promhttp.HandlerFor(registry, promhttp.HandlerOpts{EnableOpenMetrics: true})
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		request := r.Clone(r.Context())
		request.Header.Set("Accept", "application/openmetrics-text; version=1.0.0")
		export.ServeHTTP(&openMetricsWriter{ResponseWriter: w}, request)
	})

	return &Metrics{Provider: provider, Handler: handler}, nil
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
