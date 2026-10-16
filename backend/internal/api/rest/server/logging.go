package server

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"bookshop/backend/internal/api/rest/handler/requestcontext"
	"bookshop/backend/internal/api/rest/ogen"

	"github.com/google/uuid"
)

const requestIDHeader = "X-Request-ID"

func logRequests(next *ogen.Server, logger *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		started := time.Now()

		id := r.Header.Get(requestIDHeader)
		if _, err := uuid.Parse(id); err != nil {
			id = uuid.NewString()
		}

		route := "<unmatched>"
		if matched, ok := next.FindPath(r.Method, r.URL); ok {
			route = matched.PathPattern()
		}

		ctx := requestcontext.WithRequest(r.Context(), id, r.Method, route)
		w.Header().Set(requestIDHeader, id)
		recorder := &responseRecorder{ResponseWriter: w}
		next.ServeHTTP(recorder, r.WithContext(ctx))

		if logger != nil {
			logger.InfoContext(ctx, "http request completed",
				append(append([]any{"event", "http.request_completed"}, requestcontext.BuildAttributes(ctx)...),
					"status_code", recorder.getStatusCode(), "duration_ms", float64(time.Since(started))/float64(time.Millisecond))...)
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

	n, err := w.ResponseWriter.Write(data)
	if err != nil {
		return n, fmt.Errorf("write HTTP response: %w", err)
	}

	return n, nil
}

func (w *responseRecorder) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

func (w *responseRecorder) FlushError() error {
	if w.status == 0 {
		w.status = http.StatusOK
	}

	if err := http.NewResponseController(w.ResponseWriter).Flush(); err != nil {
		return fmt.Errorf("flush HTTP response: %w", err)
	}

	return nil
}

func (w *responseRecorder) getStatusCode() int {
	if w.status == 0 {
		return http.StatusOK
	}

	return w.status
}
