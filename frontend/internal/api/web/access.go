package web

import (
	"bufio"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"time"

	"bookshop/frontend/internal/logcontext"

	"github.com/google/uuid"
)

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(status int) {
	if status >= 100 && status < 200 {
		w.ResponseWriter.WriteHeader(status)
		return
	}
	if w.status != 0 {
		return
	}
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *statusWriter) Write(data []byte) (int, error) {
	if w.status == 0 {
		w.WriteHeader(http.StatusOK)
	}
	n, err := w.ResponseWriter.Write(data)
	if err != nil {
		return n, fmt.Errorf("write HTTP response: %w", err)
	}
	return n, nil
}

func (w *statusWriter) Unwrap() http.ResponseWriter { return w.ResponseWriter }

func (w *statusWriter) Flush() {
	if w.status == 0 {
		w.WriteHeader(http.StatusOK)
	}
	_ = http.NewResponseController(w.ResponseWriter).Flush()
}

func (w *statusWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	conn, rw, err := http.NewResponseController(w.ResponseWriter).Hijack()
	if err != nil {
		return nil, nil, fmt.Errorf("hijack HTTP connection: %w", err)
	}
	return conn, rw, nil
}

func (w *statusWriter) Push(target string, opts *http.PushOptions) error {
	if pusher, ok := w.ResponseWriter.(http.Pusher); ok {
		if err := pusher.Push(target, opts); err != nil {
			return fmt.Errorf("push HTTP resource: %w", err)
		}
		return nil
	}
	return http.ErrNotSupported
}

func wrapAccess(next http.Handler, findRoute func(*http.Request) string, logger *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		id := uuid.NewString()
		route := findRoute(r)
		ctx := logcontext.StoreRequest(r.Context(), id, r.Method, route)
		w.Header().Set("X-Request-ID", id)
		writer := &statusWriter{ResponseWriter: w}
		next.ServeHTTP(writer, r.WithContext(ctx))
		if writer.status == 0 {
			writer.status = http.StatusOK
		}
		if strings.HasPrefix(r.URL.Path, "/assets/") && writer.status < http.StatusBadRequest {
			return
		}
		logger.InfoContext(ctx, "HTTP request completed", "event", "http.request_completed",
			"method", r.Method, "route", route, "status_code", writer.status,
			"duration_ms", float64(time.Since(start).Microseconds())/float64((time.Millisecond/time.Microsecond)))
	})
}
