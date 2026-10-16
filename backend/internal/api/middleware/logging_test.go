package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"bookshop/backend/internal/api/rest/handler"
	"bookshop/backend/internal/api/rest/ogen"
	"bookshop/backend/internal/domain/shared"

	"github.com/google/uuid"
	ogenmiddleware "github.com/ogen-go/ogen/middleware"
	"github.com/stretchr/testify/require"
)

type loggingHandler struct {
	handler.Handler
	err        error
	observedID *uuid.UUID
}

func (h loggingHandler) ListBooks(ctx context.Context) (*ogen.ListBooksOKHeaders, error) {
	if h.observedID != nil {
		*h.observedID, _ = ctx.Value(requestIDKey{}).(uuid.UUID)
	}

	return &ogen.ListBooksOKHeaders{Response: []ogen.Book{}}, h.err
}

func TestNewRequestLoggingMiddleware(t *testing.T) {
	for _, tc := range []struct {
		name, id, method, path, route string
		status                        int
		err                           error
		direct                        bool
	}{
		{name: "valid v4", id: "31079ea7-48a9-46ee-af99-15e6f8020c38", method: http.MethodGet, path: "/books?private=secret", route: "/books", status: http.StatusOK},
		{name: "valid v1", id: "f47ac10b-58cc-11cf-a447-001000000000", method: http.MethodGet, path: "/books?private=secret", route: "/books", status: http.StatusOK},
		{name: "invalid", id: "invalid", method: http.MethodGet, path: "/books", route: "/books", status: http.StatusOK},
		{name: "absent", method: http.MethodGet, path: "/books", route: "/books", status: http.StatusOK},
		{name: "expected error", id: "f47ac10b-58cc-11cf-a447-001000000000", method: http.MethodGet, path: "/books?private=secret", route: "/books", status: http.StatusNotFound, err: shared.ErrNotFound},
		{name: "unexpected error", id: "f47ac10b-58cc-11cf-a447-001000000000", method: http.MethodGet, path: "/books?private=secret", route: "/books", status: http.StatusInternalServerError, err: errors.New("private dependency failure")},
		{name: "context and typed header", id: "f47ac10b-58cc-11cf-a447-001000000000", method: http.MethodGet, path: "/books", route: "/books", status: http.StatusOK, direct: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var (
				output     bytes.Buffer
				observedID uuid.UUID
				api        *ogen.Server
				err        error
			)

			logger := slog.New(slog.NewJSONHandler(&output, nil)).With("service", "bookshop-api")
			h := loggingHandler{err: tc.err, observedID: &observedID}
			mw := NewRequestLoggingMiddleware(&api, h, logger)

			api, err = ogen.NewServer(h, ogen.WithMiddleware(mw))
			require.NoError(t, err)

			req := httptest.NewRequest(tc.method, tc.path, nil)
			req.Header.Set(RequestIDHeader, tc.id)

			if tc.direct {
				id := uuid.MustParse(tc.id)
				response, err := mw(ogenmiddleware.Request{Context: context.Background(), Raw: req}, func(req ogenmiddleware.Request) (ogenmiddleware.Response, error) {
					require.Equal(t, id, req.Context.Value(requestIDKey{}))
					return ogenmiddleware.Response{Type: &ogen.ListBooksOKHeaders{}}, nil
				})
				require.NoError(t, err)
				require.Equal(t, id, response.Type.(*ogen.ListBooksOKHeaders).XRequestID)

				return
			}

			res := httptest.NewRecorder()
			api.ServeHTTP(res, req)
			responseID := res.Header().Get(RequestIDHeader)
			parsedResponseID, parseErr := uuid.Parse(responseID)
			require.NoError(t, parseErr)
			require.Equal(t, tc.status, res.Code)
			require.Equal(t, parsedResponseID, observedID)

			if tc.id == "" || tc.id == "invalid" {
				require.Equal(t, uuid.Version(4), parsedResponseID.Version())
			} else {
				require.Equal(t, tc.id, responseID)
			}

			lines := strings.Split(strings.TrimSpace(output.String()), "\n")
			if tc.status == http.StatusInternalServerError {
				require.Len(t, lines, 2)
				require.Contains(t, lines[1], `"event":"http.request_failed"`)
				require.Contains(t, lines[1], "private dependency failure")
				require.Contains(t, lines[1], `"request_id":"`+responseID+`"`)
			} else {
				require.Len(t, lines, 1)
			}

			var record map[string]any
			require.NoError(t, json.Unmarshal([]byte(lines[0]), &record))
			require.Equal(t, "http.request_completed", record["event"])
			require.Equal(t, "bookshop-api", record["service"])
			require.Equal(t, responseID, record["request_id"])
			require.Equal(t, tc.route, record["route"])
			require.InDelta(t, float64(tc.status), record["status_code"], 0)
			require.NotContains(t, output.String(), "secret")
		})
	}
}
