package metrics

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Server struct {
	http     *http.Server
	listener net.Listener
	done     chan error
}

func NewServer(address string, registry *prometheus.Registry) *Server {
	export := promhttp.HandlerFor(registry, promhttp.HandlerOpts{EnableOpenMetrics: true})
	mux := http.NewServeMux()
	mux.Handle("/metrics", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		request := r.Clone(r.Context())
		request.Header.Set("Accept", "application/openmetrics-text; version=1.0.0")
		export.ServeHTTP(w, request)
	}))

	return &Server{http: &http.Server{Addr: address, Handler: mux}}
}

func (s *Server) Start(context.Context) error {
	listener, err := net.Listen("tcp", s.http.Addr)
	if err != nil {
		return fmt.Errorf("listen for metrics at %s: %w", s.http.Addr, err)
	}

	s.listener = listener
	s.done = make(chan error, 1)

	go func() { s.done <- s.http.Serve(listener) }()

	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	err := s.http.Shutdown(ctx)
	if err != nil {
		_ = s.http.Close()
	}

	<-s.done

	if err != nil {
		return fmt.Errorf("shut down metrics listener: %w", err)
	}

	return nil
}
