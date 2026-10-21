package web

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
)

type HTTPServer struct {
	server          *http.Server
	listener        net.Listener
	done            chan error
	metrics         *http.Server
	metricsListener net.Listener
	metricsDone     chan error
	logger          *slog.Logger
}

func NewHTTPServer(addr, metricsAddr string, handler http.Handler) *HTTPServer {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /metrics", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/openmetrics-text; version=1.0.0; charset=utf-8")
		_, _ = w.Write([]byte("# EOF\n"))
	})
	return &HTTPServer{server: &http.Server{Addr: addr, Handler: handler}, metrics: &http.Server{Addr: metricsAddr, Handler: mux}}
}

func (s *HTTPServer) SetLogger(logger *slog.Logger) {
	s.logger = logger
}

func (s *HTTPServer) Start(context.Context) error {
	listener, err := net.Listen("tcp", s.server.Addr)
	if err != nil {
		return err
	}
	metricsListener, err := net.Listen("tcp", s.metrics.Addr)
	if err != nil {
		_ = listener.Close()
		return fmt.Errorf("listen for metrics at %s: %w", s.metrics.Addr, err)
	}
	s.listener = listener
	s.metricsListener = metricsListener
	s.done = make(chan error, 1)
	s.metricsDone = make(chan error, 1)
	go func() { s.metricsDone <- s.metrics.Serve(metricsListener) }()
	go func() {
		err := s.server.Serve(listener)
		if err != nil && !errors.Is(err, http.ErrServerClosed) && s.logger != nil {
			s.logger.Error("web listener failed", "event", "http.listener_failed", "address", s.server.Addr, "error", err)
		}
		s.done <- err
	}()
	return nil
}

func (s *HTTPServer) Stop(ctx context.Context) error {
	metricsResult := make(chan error, 1)
	go func() {
		metricsErr := s.metrics.Shutdown(ctx)
		if metricsErr != nil {
			_ = s.metrics.Close()
		}
		metricsResult <- metricsErr
	}()
	err := s.server.Shutdown(ctx)
	if err != nil {
		_ = s.server.Close()
	}
	metricsErr := <-metricsResult
	runErr := <-s.done
	<-s.metricsDone
	if err != nil {
		return err
	}
	if metricsErr != nil {
		return fmt.Errorf("shut down metrics: %w", metricsErr)
	}
	if runErr != nil && !errors.Is(runErr, http.ErrServerClosed) {
		return runErr
	}
	return nil
}
