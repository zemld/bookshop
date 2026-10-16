package server

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"bookshop/backend/internal/api/rest/ogen"

	"github.com/google/uuid"
	"github.com/ogen-go/ogen/middleware"
)

const _requestIDHeader = "X-Request-ID"

type requestIDKey struct{}

func logRequests(server **ogen.Server, handler ogen.Handler, logger *slog.Logger) ogen.Middleware {
	return func(req middleware.Request, next middleware.Next) (middleware.Response, error) {
		started := time.Now()

		id := getRequestID(req.Raw.Header)

		req.SetContext(context.WithValue(req.Context, requestIDKey{}, id))
		route, _ := (*server).FindPath(req.Raw.Method, req.Raw.URL)
		attrs := []any{"request_id", id, "method", req.Raw.Method, "route", route.PathPattern()}
		response, err := next(req)

		var (
			status = http.StatusOK
			coded  *ogen.ErrorStatusCodeWithHeaders
		)

		if err != nil {
			var ok bool

			coded, ok = err.(*ogen.ErrorStatusCodeWithHeaders)
			if !ok {
				coded = handler.NewError(req.Context, err)
			}

			coded.XRequestID = id
			status = coded.StatusCode
		} else if headers, ok := response.Type.(interface{ SetXRequestID(uuid.UUID) }); ok {
			headers.SetXRequestID(id)
		}

		if logger != nil {
			logger.InfoContext(req.Context, "http request completed", append(
				append([]any{"event", "http.request_completed"}, attrs...),
				"status_code", status, "duration_ms", float64(time.Since(started))/float64(time.Millisecond),
			)...)

			if status >= http.StatusInternalServerError && err != nil {
				logger.ErrorContext(req.Context, "http request failed", append(
					append([]any{"event", "http.request_failed"}, attrs...),
					"status_code", status, "error", err,
				)...)
			}
		}

		if err != nil {
			return response, coded
		}

		return response, nil
	}
}

func getRequestID(header http.Header) uuid.UUID {
	id, err := uuid.Parse(header.Get(_requestIDHeader))
	if err != nil {
		return uuid.New()
	}

	return id
}
