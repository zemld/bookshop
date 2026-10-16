package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"regexp"
	"strings"
	"time"

	"bookshop/backend/internal/api/rest/ogen"

	"github.com/ogen-go/ogen/ogenerrors"
	"github.com/ogen-go/ogen/validate"
	"go.uber.org/fx"
)

const _drainTimeout = 10 * time.Second

var decodedField = regexp.MustCompile(`decode field "([^"]+)"`)

func decodeProblem(err error) ogen.Problem {
	var (
		body       *ogenerrors.DecodeBodyError
		param      *ogenerrors.DecodeParamError
		validation *validate.Error
	)

	switch {
	case errors.As(err, &param) && param.Name == "id":
		return ogen.Problem{Error: ogen.ProblemErrorInvalidID}
	case errors.As(err, &validation):
		for _, field := range validation.Fields {
			if errors.Is(field.Error, validate.ErrFieldRequired) {
				if code, ok := decodeFieldCode(field.Name); ok {
					return ogen.Problem{Error: code}
				}
			}
		}
	case errors.As(err, &body):
		return ogen.Problem{Error: identifyBodyError(body.Err)}
	}

	return ogen.Problem{Error: ogen.ProblemErrorInvalidRequest}
}

func identifyBodyError(err error) ogen.ProblemError {
	if match := decodedField.FindStringSubmatch(err.Error()); match != nil {
		if code, ok := decodeFieldCode(match[1]); ok {
			return code
		}
	}

	if strings.Contains(err.Error(), "unexpected trailing data") {
		return ogen.ProblemErrorTrailingJSONData
	}

	return ogen.ProblemErrorInvalidJSON
}

func decodeFieldCode(name string) (ogen.ProblemError, bool) {
	switch name {
	case "author":
		return ogen.ProblemErrorInvalidAuthor, true
	case "name":
		return ogen.ProblemErrorInvalidName, true
	case "year":
		return ogen.ProblemErrorInvalidYear, true
	case "price":
		return ogen.ProblemErrorInvalidPrice, true
	case "publisherId":
		return ogen.ProblemErrorInvalidPublisherID, true
	case "publicationYear":
		return ogen.ProblemErrorInvalidPublicationYear, true
	case "quantity":
		return ogen.ProblemErrorInvalidQuantity, true
	default:
		return "", false
	}
}
func New(handler ogen.Handler, logger *slog.Logger) http.Handler {
	var (
		server *ogen.Server
		err    error
	)

	server, err = ogen.NewServer(handler, ogen.WithMiddleware(logRequests(&server, handler, logger)), ogen.WithErrorHandler(func(_ context.Context, w http.ResponseWriter, r *http.Request, decodeErr error) {
		w.Header().Set(_requestIDHeader, getRequestID(r.Header).String())
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusBadRequest)

		problem := decodeProblem(decodeErr)
		_ = json.NewEncoder(w).Encode(&problem)
	}), ogen.WithNotFound(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(_requestIDHeader, getRequestID(r.Header).String())
		http.NotFound(w, r)
	}), ogen.WithMethodNotAllowed(func(w http.ResponseWriter, r *http.Request, allowed string) {
		w.Header().Set(_requestIDHeader, getRequestID(r.Header).String())

		if r.Method == http.MethodOptions {
			w.Header().Set("Access-Control-Allow-Methods", allowed)
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			w.WriteHeader(http.StatusNoContent)
		} else {
			w.Header().Set("Allow", allowed)
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
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
	logger   *slog.Logger
}

func NewServer(lifecycle fx.Lifecycle, cfg Config, handler ogen.Handler, logger *slog.Logger) *Server {
	s := newServer(cfg.Address, New(handler, logger))
	s.logger = logger
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
			if s.logger != nil {
				s.logger.Error("api listener failed", "event", "http.listener_failed", "address", s.http.Addr, "error", err)
			}
		}

		s.done <- err
	}()

	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	drain, cancel := context.WithTimeout(ctx, _drainTimeout)
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
