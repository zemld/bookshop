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
	api := New(handler.Handler{Logger: logger}, logger)

	for _, tc := range []struct {
		id, method, path, route string
		status                  int
	}{
		{"31079ea7-48a9-46ee-af99-15e6f8020c38", http.MethodGet, "/books/bad?private=secret", "/books/{id}", http.StatusBadRequest},
		{"invalid", http.MethodGet, "/books/bad", "/books/{id}", http.StatusBadRequest},
		{"", http.MethodGet, "/unmatched/secret", "<unmatched>", http.StatusNotFound},
		{"", http.MethodPatch, "/books", "<unmatched>", http.StatusMethodNotAllowed},
	} {
		output.Reset()

		req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(""))
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
		require.Equal(t, tc.status, res.Code)
		require.InDelta(t, float64(tc.status), record["status_code"], 0)
		require.NotContains(t, output.String(), "secret")
	}
}

func TestResponseRecorder(t *testing.T) {
	for _, tc := range []struct {
		name string
		run  func(http.ResponseWriter)
		want int
	}{
		{"empty response", func(http.ResponseWriter) {}, http.StatusOK},
		{"implicit write", func(w http.ResponseWriter) { _, _ = w.Write([]byte("ok")) }, http.StatusOK},
		{"informational then error", func(w http.ResponseWriter) {
			w.WriteHeader(http.StatusEarlyHints)
			w.WriteHeader(http.StatusBadRequest)
		}, http.StatusBadRequest},
		{"explicit error", func(w http.ResponseWriter) { w.WriteHeader(http.StatusInternalServerError) }, http.StatusInternalServerError},
		{"flush then error", func(w http.ResponseWriter) {
			require.NoError(t, http.NewResponseController(w).Flush())
			w.WriteHeader(http.StatusInternalServerError)
		}, http.StatusOK},
	} {
		t.Run(tc.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			status := &responseRecorder{ResponseWriter: recorder}
			tc.run(status)
			require.Equal(t, tc.want, status.getStatusCode())

			if tc.name != "empty response" && tc.name != "informational then error" {
				require.Equal(t, tc.want, recorder.Code)
			}
		})
	}
}
