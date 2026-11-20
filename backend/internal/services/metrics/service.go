package metrics

import (
	"fmt"

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

const (
	_buildInfoName  = "bookshop_build_info"
	_buildInfoValue = 1
	_serviceLabel   = "service"
	_versionLabel   = "version"
	_revisionLabel  = "revision"
)

func (s *Service) RegisterBuildInfo(service, version, revision string) error {
	info := prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Name: _buildInfoName,
		Help: "Build information for the running Bookshop service.",
		ConstLabels: prometheus.Labels{
			_serviceLabel: service, _versionLabel: version, _revisionLabel: revision,
		},
	}, func() float64 { return _buildInfoValue })
	if err := s.Registry.Register(info); err != nil {
		return fmt.Errorf("register build info: %w", err)
	}
	return nil
}
