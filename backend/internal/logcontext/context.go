package logcontext

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

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

func NewHandler(next slog.Handler) *Handler { return &Handler{next: next} }

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
