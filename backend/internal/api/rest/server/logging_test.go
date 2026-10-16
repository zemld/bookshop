package server

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
	"github.com/stretchr/testify/require"
)

type loggingHandler struct {
	handler.Handler
	err  error
	seen chan uuid.UUID
}

func (h loggingHandler) ListBooks(ctx context.Context) (*ogen.ListBooksOKHeaders, error) {
	if h.seen != nil {
		h.seen <- ctx.Value(requestIDKey{}).(uuid.UUID)
	}

	return &ogen.ListBooksOKHeaders{Response: []ogen.Book{}}, h.err
}

func TestLogRequests(t *testing.T) {
	for _, tc := range []struct {
		name, id, method, path, route string
		status                        int
		logged                        bool
		err                           error
	}{
		{"invalid path", "31079ea7-48a9-46ee-af99-15e6f8020c38", http.MethodGet, "/books/bad?private=secret", "/books/{id}", http.StatusBadRequest, false, nil},
		{"valid v4", "31079ea7-48a9-46ee-af99-15e6f8020c38", http.MethodGet, "/books?private=secret", "/books", http.StatusOK, true, nil},
		{"valid v1", "f47ac10b-58cc-11cf-a447-001000000000", http.MethodGet, "/books?private=secret", "/books", http.StatusOK, true, nil},
		{"v1 prehook", "f47ac10b-58cc-11cf-a447-001000000000", http.MethodGet, "/books/bad?private=secret", "/books/{id}", http.StatusBadRequest, false, nil},
		{"invalid", "invalid", http.MethodGet, "/books", "/books", http.StatusOK, true, nil},
		{"absent", "", http.MethodGet, "/books", "/books", http.StatusOK, true, nil},
		{"not found", "", http.MethodGet, "/unmatched/secret", "<unmatched>", http.StatusNotFound, false, nil},
		{"method not allowed", "", http.MethodPatch, "/books", "<unmatched>", http.StatusMethodNotAllowed, false, nil},
		{"options", "", http.MethodOptions, "/books", "<unmatched>", http.StatusNoContent, false, nil},
		{"expected error", "f47ac10b-58cc-11cf-a447-001000000000", http.MethodGet, "/books?private=secret", "/books", http.StatusNotFound, true, shared.ErrNotFound},
		{"unexpected error", "f47ac10b-58cc-11cf-a447-001000000000", http.MethodGet, "/books?private=secret", "/books", http.StatusInternalServerError, true, errors.New("private dependency failure")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var output bytes.Buffer

			logger := slog.New(slog.NewJSONHandler(&output, nil)).With("service", "bookshop-api")
			seen := make(chan uuid.UUID, 1)
			api := New(loggingHandler{err: tc.err, seen: seen}, logger)
			req := httptest.NewRequest(tc.method, tc.path, nil)
			req.Header.Set("X-Request-ID", tc.id)

			res := httptest.NewRecorder()
			api.ServeHTTP(res, req)
			responseID := res.Header().Get("X-Request-ID")
			parsedResponseID, parseErr := uuid.Parse(responseID)
			require.NoError(t, parseErr)
			require.Equal(t, tc.status, res.Code)

			if tc.id == "" || tc.id == "invalid" {
				require.Equal(t, uuid.Version(4), parsedResponseID.Version())
			} else {
				require.Equal(t, tc.id, responseID)
			}

			if !tc.logged {
				require.Empty(t, output.String())

				if tc.status == http.StatusMethodNotAllowed {
					require.Equal(t, "GET,POST", res.Header().Get("Allow"))
				}

				if tc.status == http.StatusNoContent {
					require.Equal(t, "GET,POST", res.Header().Get("Access-Control-Allow-Methods"))
					require.Equal(t, "Content-Type", res.Header().Get("Access-Control-Allow-Headers"))
				}

				return
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
			require.Equal(t, responseID, (<-seen).String())
			require.Equal(t, "bookshop-api", record["service"])
			require.Equal(t, responseID, record["request_id"])
			require.Equal(t, tc.route, record["route"])
			require.InDelta(t, float64(tc.status), record["status_code"], 0)
			require.NotContains(t, output.String(), "secret")
		})
	}
}
