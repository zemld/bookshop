package web

import (
	"log/slog"
	"net/http"
	"time"

	"bookshop/frontend/internal/requestlog"

	"github.com/google/uuid"
)

func logRequests(mux *http.ServeMux, logger *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		started := time.Now()
		id := uuid.NewString()
		_, pattern := mux.Handler(r)
		if pattern == "" {
			pattern = "<unmatched>"
		}
		ctx := requestlog.WithDetails(r.Context(), requestlog.Details{ID: id, Method: r.Method, Route: pattern})
		w.Header().Set(requestlog.Header, id)
		recorder := &statusWriter{ResponseWriter: w}
		mux.ServeHTTP(wrapStatusWriter(recorder), r.WithContext(ctx))
		status := recorder.status
		if status == 0 {
			status = http.StatusOK
		}
		if logger != nil && (pattern != "GET /assets/" || status >= 400) {
			logger.InfoContext(ctx, "http request completed", append(
				[]any{"event", "http.request_completed"},
				append(requestlog.BuildAttributes(ctx), "status_code", status, "duration_ms", float64(time.Since(started).Microseconds())/1000)...,
			)...)
		}
	})
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(status int) {
	if w.status == 0 && status >= 200 {
		w.status = status
	}
	w.ResponseWriter.WriteHeader(status)
}

func (w *statusWriter) Write(data []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	return w.ResponseWriter.Write(data)
}

func (w *statusWriter) Unwrap() http.ResponseWriter { return w.ResponseWriter }

func (w *statusWriter) recordFlush() {
	if w.status == 0 {
		w.status = http.StatusOK
	}
}

type flushingWriter struct {
	*statusWriter
	http.Flusher
}

func (w flushingWriter) Flush() {
	w.recordFlush()
	w.Flusher.Flush()
}

type hijackingWriter struct {
	*statusWriter
	http.Hijacker
}

type flushingHijackingWriter struct {
	*statusWriter
	http.Flusher
	http.Hijacker
}

func (w flushingHijackingWriter) Flush() {
	w.recordFlush()
	w.Flusher.Flush()
}

type pushingWriter struct {
	*statusWriter
	http.Pusher
}

type flushingPushingWriter struct {
	*statusWriter
	http.Flusher
	http.Pusher
}

func (w flushingPushingWriter) Flush() {
	w.recordFlush()
	w.Flusher.Flush()
}

type hijackingPushingWriter struct {
	*statusWriter
	http.Hijacker
	http.Pusher
}

type flushingHijackingPushingWriter struct {
	*statusWriter
	http.Flusher
	http.Hijacker
	http.Pusher
}

func (w flushingHijackingPushingWriter) Flush() {
	w.recordFlush()
	w.Flusher.Flush()
}

func wrapStatusWriter(w *statusWriter) http.ResponseWriter {
	flusher, flushes := w.ResponseWriter.(http.Flusher)
	hijacker, hijacks := w.ResponseWriter.(http.Hijacker)
	pusher, pushes := w.ResponseWriter.(http.Pusher)
	switch {
	case flushes && hijacks && pushes:
		return flushingHijackingPushingWriter{w, flusher, hijacker, pusher}
	case flushes && pushes:
		return flushingPushingWriter{w, flusher, pusher}
	case hijacks && pushes:
		return hijackingPushingWriter{w, hijacker, pusher}
	case pushes:
		return pushingWriter{w, pusher}
	case flushes && hijacks:
		return flushingHijackingWriter{w, flusher, hijacker}
	case flushes:
		return flushingWriter{w, flusher}
	case hijacks:
		return hijackingWriter{w, hijacker}
	default:
		return w
	}
}
