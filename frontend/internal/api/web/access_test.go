package web

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"bookshop/frontend/internal/logcontext"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestWrapAccess(t *testing.T) {
	for _, tc := range []struct {
		name   string
		path   string
		status int
		flush  bool
		logged bool
	}{
		{name: "application request", path: "/books/private?secret=hidden", status: http.StatusConflict, logged: true},
		{name: "successful asset", path: "/assets/site.css", status: http.StatusOK},
		{name: "missing asset", path: "/assets/private.css", status: http.StatusNotFound, logged: true},
		{name: "flush commits OK", path: "/books", status: http.StatusOK, flush: true, logged: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var output bytes.Buffer
			logger := slog.New(logcontext.NewHandler(slog.NewJSONHandler(&output, nil)))
			handler := wrapAccess(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				if tc.flush {
					w.(http.Flusher).Flush()
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				w.WriteHeader(tc.status)
			}), func(*http.Request) string { return "GET /{resource}" }, logger)
			done := make(chan struct{})
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				handler.ServeHTTP(w, r)
				close(done)
			}))
			defer server.Close()
			request, err := http.NewRequest(http.MethodGet, server.URL+tc.path, nil)
			require.NoError(t, err)
			request.Header.Set("X-Request-ID", uuid.NewString())
			response, err := server.Client().Do(request)
			require.NoError(t, err)
			require.NoError(t, response.Body.Close())
			<-done
			require.Equal(t, tc.status, response.StatusCode)
			id := response.Header.Get("X-Request-ID")
			parsed, err := uuid.Parse(id)
			require.NoError(t, err)
			require.Equal(t, uuid.RFC4122, parsed.Variant())
			require.Equal(t, uuid.Version(4), parsed.Version())
			require.NotEqual(t, request.Header.Get("X-Request-ID"), id)
			if !tc.logged {
				require.Empty(t, output.String())
				return
			}
			var record map[string]any
			require.NoError(t, json.Unmarshal(bytes.TrimSpace(output.Bytes()), &record))
			require.Equal(t, id, record["request_id"])
			require.Equal(t, tc.status, int(record["status_code"].(float64)))
			require.Equal(t, "GET /{resource}", record["route"])
			require.NotContains(t, output.String(), "hidden")
			require.NotContains(t, output.String(), "private")
		})
	}
}
