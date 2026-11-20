package metrics

import (
	"fmt"
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	otelexport "go.opentelemetry.io/otel/exporters/prometheus"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
)

type Service struct {
	Provider *sdkmetric.MeterProvider
	Registry *prometheus.Registry
}

func New() (*Service, error) {
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
	return &Service{Provider: provider, Registry: registry}, nil
}

var (
	version  = "dev"
	revision = "unknown"
)

func (s *Service) RegisterBuildInfo(service string, logger *slog.Logger) error {
	info := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "bookshop_build_info",
		Help: "Build information for the running Bookshop service.",
		ConstLabels: prometheus.Labels{
			"service": service, "version": version, "revision": revision,
		},
	})
	info.Set(1)
	if err := s.Registry.Register(info); err != nil {
		return fmt.Errorf("register build info: %w", err)
	}
	logger.Info("loaded service build", "event", "service.build", "version", version, "revision", revision)
	return nil
}
