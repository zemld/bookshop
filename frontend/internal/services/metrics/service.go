package metrics

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	otelexport "go.opentelemetry.io/otel/exporters/prometheus"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
)

type Service struct {
	Provider        *sdkmetric.MeterProvider
	Registry        *prometheus.Registry
	CatalogRequests *prometheus.CounterVec
}

func New() (*Service, error) {
	registry := prometheus.NewRegistry()
	if err := registry.Register(collectors.NewGoCollector()); err != nil {
		return nil, fmt.Errorf("register Go collector: %w", err)
	}
	if err := registry.Register(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{})); err != nil {
		return nil, fmt.Errorf("register process collector: %w", err)
	}
	catalogRequests := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "bookshop_frontend_lab_catalog_requests_total",
		Help: "Number of lab GET /books requests received by the frontend.",
	}, []string{"version"})
	if err := registry.Register(catalogRequests); err != nil {
		return nil, fmt.Errorf("register catalog requests: %w", err)
	}
	exporter, err := otelexport.New(otelexport.WithRegisterer(registry))
	if err != nil {
		return nil, fmt.Errorf("initialize Prometheus exporter: %w", err)
	}
	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(exporter))
	return &Service{Provider: provider, Registry: registry, CatalogRequests: catalogRequests}, nil
}
