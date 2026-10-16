package server

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"bookshop/backend/internal/api/rest/handler"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestLogRequests(t *testing.T) {
	var output bytes.Buffer

	logger := slog.New(slog.NewJSONHandler(&output, nil)).With("service", "bookshop-api")
	api := newWithLogger(handler.Handler{Logger: logger}, logger)

	for _, tc := range []struct {
		id, path, route string
	}{
		{"31079ea7-48a9-46ee-af99-15e6f8020c38", "/books/bad?private=secret", "/books/{id}"},
		{"invalid", "/books/bad", "/books/{id}"},
		{"", "/unmatched/secret", "<unmatched>"},
	} {
		output.Reset()

		req := httptest.NewRequest("GET", tc.path, strings.NewReader(""))
		req.Header.Set("X-Request-ID", tc.id)

		res := httptest.NewRecorder()
		api.ServeHTTP(res, req)
		id := res.Header().Get("X-Request-ID")
		parsed, err := uuid.Parse(id)
		require.NoError(t, err)

		if tc.id == "" || tc.id == "invalid" {
			require.Equal(t, uuid.Version(4), parsed.Version())
		} else {
			require.Equal(t, tc.id, id)
		}

		var record map[string]any
		require.NoError(t, json.Unmarshal(output.Bytes(), &record))
		require.Equal(t, "http.request_completed", record["event"])
		require.Equal(t, "bookshop-api", record["service"])
		require.Equal(t, id, record["request_id"])
		require.Equal(t, tc.route, record["route"])
		require.InDelta(t, float64(res.Code), record["status_code"], 0)
		require.NotContains(t, output.String(), "secret")
	}
}

func TestWrapStatusWriter(t *testing.T) {
	recorder := httptest.NewRecorder()
	status := &statusWriter{ResponseWriter: recorder}
	wrapped := wrapStatusWriter(status)
	require.NoError(t, http.NewResponseController(wrapped).Flush())
	wrapped.WriteHeader(http.StatusInternalServerError)
	require.Equal(t, http.StatusOK, recorder.Code)
	require.Equal(t, http.StatusOK, status.status)
}
