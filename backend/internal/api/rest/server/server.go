package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"

	"bookshop/backend/internal/api/rest/ogen"

	"go.uber.org/fx"
)

const drainTimeout = 10 * time.Second

// New wraps the generated API with a stable JSON response for decode errors.
func New(handler ogen.Handler) http.Handler {
	server, err := ogen.NewServer(handler, ogen.WithErrorHandler(func(_ context.Context, w http.ResponseWriter, _ *http.Request, _ error) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"invalid input"}`))
	}))
	if err != nil {
		panic(err)
	}

	return server
}

type Config struct {
	Address string
}

type Server struct {
	http     *http.Server
	listener net.Listener
	done     chan error
}

func NewServer(lifecycle fx.Lifecycle, cfg Config, handler ogen.Handler) *Server {
	s := newServer(cfg.Address, New(handler))
	lifecycle.Append(fx.Hook{OnStart: s.Start, OnStop: s.Stop})

	return s
}

func newServer(addr string, handler http.Handler) *Server {
	return &Server{http: &http.Server{Addr: addr, Handler: handler}}
}

func (s *Server) Start(context.Context) error {
	listener, err := net.Listen("tcp", s.http.Addr)
	if err != nil {
		return fmt.Errorf("listen for API: %w", err)
	}

	s.listener = listener
	s.done = make(chan error, 1)

	go func() {
		err := s.http.Serve(listener)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("api listener stopped", "error", err)
		}

		s.done <- err
	}()

	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	drain, cancel := context.WithTimeout(ctx, drainTimeout)
	defer cancel()

	err := s.http.Shutdown(drain)
	if err != nil {
		_ = s.http.Close()
	}

	<-s.done

	if err != nil {
		return fmt.Errorf("shut down API: %w", err)
	}

	return nil
}
