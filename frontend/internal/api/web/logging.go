package web

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
)

const _requestIDHeader = "X-Request-ID"

func logRequests(mux *http.ServeMux, logger *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		started := time.Now()
		id := uuid.NewString()
		_, pattern := mux.Handler(r)
		if pattern == "" {
			pattern = "<unmatched>"
		}
		ctx := attachRequest(r.Context(), id, r.Method, pattern)
		w.Header().Set(_requestIDHeader, id)
		recorder := &responseRecorder{ResponseWriter: w}
		mux.ServeHTTP(recorder, r.WithContext(ctx))
		status := recorder.getStatusCode()
		if logger != nil && (pattern != "GET /assets/" || status >= http.StatusBadRequest) {
			logger.InfoContext(ctx, "http request completed", append(
				[]any{"event", "http.request_completed"},
				append(buildRequestAttributes(ctx), "status_code", status, "duration_ms", float64(time.Since(started).Microseconds())/1000)...,
			)...)
		}
	})
}

type responseRecorder struct {
	http.ResponseWriter
	status int
}

func (w *responseRecorder) WriteHeader(status int) {
	if w.status == 0 && status >= http.StatusOK {
		w.status = status
	}
	w.ResponseWriter.WriteHeader(status)
}

func (w *responseRecorder) Write(data []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	return w.ResponseWriter.Write(data)
}

func (w *responseRecorder) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

func (w *responseRecorder) FlushError() error {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	return http.NewResponseController(w.ResponseWriter).Flush()
}

func (w *responseRecorder) getStatusCode() int {
	if w.status == 0 {
		return http.StatusOK
	}
	return w.status
}
