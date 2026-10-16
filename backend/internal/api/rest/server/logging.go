package server

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"bookshop/backend/internal/api/rest/ogen"
	"bookshop/backend/internal/requestlog"

	"github.com/google/uuid"
)

func logRequests(next *ogen.Server, logger *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		started := time.Now()
		id := r.Header.Get(requestlog.Header)

		if _, err := uuid.Parse(id); err != nil {
			id = uuid.NewString()
		}

		route := "<unmatched>"
		if matched, ok := next.FindPath(r.Method, r.URL); ok {
			route = matched.PathPattern()
		}

		ctx := requestlog.WithDetails(r.Context(), requestlog.Details{ID: id, Method: r.Method, Route: route})
		w.Header().Set(requestlog.Header, id)
		recorder := &statusWriter{ResponseWriter: w}
		next.ServeHTTP(wrapStatusWriter(recorder), r.WithContext(ctx))

		status := recorder.status
		if status == 0 {
			status = http.StatusOK
		}

		if logger != nil {
			logger.InfoContext(ctx, "http request completed",
				append(append([]any{"event", "http.request_completed"}, requestlog.BuildAttributes(ctx)...),
					"status_code", status, "duration_ms", float64(time.Since(started))/float64(time.Millisecond))...)
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

	n, err := w.ResponseWriter.Write(data)
	if err != nil {
		return n, fmt.Errorf("write HTTP response: %w", err)
	}

	return n, nil
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

	if flushes {
		if hijacks {
			if pushes {
				return flushingHijackingPushingWriter{w, flusher, hijacker, pusher}
			}

			return flushingHijackingWriter{w, flusher, hijacker}
		}

		if pushes {
			return flushingPushingWriter{w, flusher, pusher}
		}

		return flushingWriter{w, flusher}
	}

	if hijacks {
		if pushes {
			return hijackingPushingWriter{w, hijacker, pusher}
		}

		return hijackingWriter{w, hijacker}
	}

	if pushes {
		return pushingWriter{w, pusher}
	}

	return w
}
