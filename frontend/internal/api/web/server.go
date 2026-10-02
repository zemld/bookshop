package web

import (
	"context"
	"errors"
	"net"
	"net/http"
)

type HTTPServer struct {
	server   *http.Server
	listener net.Listener
	done     chan error
}

func NewHTTPServer(addr string, handler http.Handler) *HTTPServer {
	return &HTTPServer{server: &http.Server{Addr: addr, Handler: handler}}
}

func (s *HTTPServer) Start(context.Context) error {
	listener, err := net.Listen("tcp", s.server.Addr)
	if err != nil {
		return err
	}
	s.listener = listener
	s.done = make(chan error, 1)
	go func() { s.done <- s.server.Serve(listener) }()
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
