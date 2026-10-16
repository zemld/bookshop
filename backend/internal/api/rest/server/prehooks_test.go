package server_test

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"bookshop/backend/internal/api/rest/handler"
	"bookshop/backend/internal/api/rest/server"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	for _, tc := range []struct {
		name, id, method, path string
		status                 int
	}{
		{"invalid path", "31079ea7-48a9-46ee-af99-15e6f8020c38", http.MethodGet, "/books/bad?private=secret", http.StatusBadRequest},
		{"v1 prehook", "f47ac10b-58cc-11cf-a447-001000000000", http.MethodGet, "/books/bad?private=secret", http.StatusBadRequest},
		{"not found", "", http.MethodGet, "/unmatched/secret", http.StatusNotFound},
		{"method not allowed", "", http.MethodPatch, "/books", http.StatusMethodNotAllowed},
		{"options", "", http.MethodOptions, "/books", http.StatusNoContent},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var output bytes.Buffer

			logger := slog.New(slog.NewJSONHandler(&output, nil))
			api := server.New(handler.Handler{}, logger)
			req := httptest.NewRequest(tc.method, tc.path, nil)
			req.Header.Set("X-Request-ID", tc.id)

			res := httptest.NewRecorder()

			api.ServeHTTP(res, req)

			responseID := res.Header().Get("X-Request-ID")
			parsedID, err := uuid.Parse(responseID)
			require.NoError(t, err)
			require.Equal(t, tc.status, res.Code)
			require.Empty(t, output.String())

			if tc.id == "" {
				require.Equal(t, uuid.Version(4), parsedID.Version())
			} else {
				require.Equal(t, tc.id, responseID)
			}

			if tc.status == http.StatusMethodNotAllowed {
				require.Equal(t, "GET,POST", res.Header().Get("Allow"))
			}

			if tc.status == http.StatusNoContent {
				require.Equal(t, "GET,POST", res.Header().Get("Access-Control-Allow-Methods"))
				require.Equal(t, "Content-Type", res.Header().Get("Access-Control-Allow-Headers"))
			}
		})
	}
}
