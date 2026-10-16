package requestlog

import (
	"bufio"
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

type contextKey struct{}

type requestInfo struct {
	id     string
	method string
	route  string
}

func StoreRequest(ctx context.Context, id, method, route string) context.Context {
	return context.WithValue(ctx, contextKey{}, requestInfo{id, method, route})
}

func ReadID(ctx context.Context) string {
	if info, ok := ctx.Value(contextKey{}).(requestInfo); ok {
		return info.id
	}

	return ""
}

func ReadAttributes(ctx context.Context) []any {
	if info, ok := ctx.Value(contextKey{}).(requestInfo); ok {
		return []any{"method", info.method, "route", info.route}
	}

	return nil
}

type Handler struct {
	next slog.Handler
}

func NewHandler(next slog.Handler) *Handler {
	return &Handler{next: next}
}

func (h *Handler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.next.Enabled(ctx, level)
}

func (h *Handler) Handle(ctx context.Context, record slog.Record) error {
	if id := ReadID(ctx); id != "" {
		record.AddAttrs(slog.String("request_id", id))
	}

	if err := h.next.Handle(ctx, record); err != nil {
		return fmt.Errorf("handle log record: %w", err)
	}

	return nil
}

func (h *Handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return NewHandler(h.next.WithAttrs(attrs))
}

func (h *Handler) WithGroup(name string) slog.Handler {
	return NewHandler(h.next.WithGroup(name))
}

func ResolveID(value string) string {
	id, err := uuid.Parse(value)
	if err == nil && id.Version() == 4 && id.Variant() == uuid.RFC4122 && strings.EqualFold(id.String(), value) {
		return value
	}

	return uuid.NewString()
}

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

func (w *statusWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

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

func Wrap(next http.Handler, findRoute func(*http.Request) string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		id := ResolveID(r.Header.Get("X-Request-ID"))
		route := findRoute(r)
		ctx := StoreRequest(r.Context(), id, r.Method, route)
		w.Header().Set("X-Request-ID", id)
		writer := &statusWriter{ResponseWriter: w}
		next.ServeHTTP(writer, r.WithContext(ctx))

		if writer.status == 0 {
			writer.status = http.StatusOK
		}

		slog.InfoContext(ctx, "HTTP request completed", "event", "http.request_completed",
			"method", r.Method, "route", route, "status_code", writer.status,
			"duration_ms", float64(time.Since(start).Microseconds())/float64((time.Millisecond/time.Microsecond)))
	})
}
