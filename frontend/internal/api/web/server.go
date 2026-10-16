package web

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
)

type HTTPServer struct {
	server   *http.Server
	listener net.Listener
	done     chan error
	logger   *slog.Logger
}

func NewHTTPServer(addr string, handler http.Handler) *HTTPServer {
	return &HTTPServer{server: &http.Server{Addr: addr, Handler: handler}}
}

func (s *HTTPServer) SetLogger(logger *slog.Logger) {
	s.logger = logger
}

func (s *HTTPServer) Start(context.Context) error {
	listener, err := net.Listen("tcp", s.server.Addr)
	if err != nil {
		return err
	}
	s.listener = listener
	s.done = make(chan error, 1)
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
	err := s.server.Shutdown(ctx)
	if err != nil {
		_ = s.server.Close()
	}
	runErr := <-s.done
	if err != nil {
		return err
	}
	if runErr != nil && !errors.Is(runErr, http.ErrServerClosed) {
		return runErr
	}
	return nil
}
