package web

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"bookshop/frontend/internal/logcontext"

	"github.com/stretchr/testify/require"
)

func TestCreateHandler(t *testing.T) {
	for _, route := range []string{"/books", "/publishers"} {
		t.Run(route, func(t *testing.T) {
			var output bytes.Buffer
			handler := (Server{Logger: slog.New(logcontext.NewHandler(slog.NewJSONHandler(&output, nil)))}).CreateHandler()
			done := make(chan struct{})
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				handler.ServeHTTP(w, r)
				close(done)
			}))
			defer server.Close()
			request, err := http.NewRequest(http.MethodPost, server.URL+route+"?private=%PRIVATE_QUERY_MARKER",
				strings.NewReader("name=%PRIVATE_FORM_MARKER"))
			require.NoError(t, err)
			request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			response, err := server.Client().Do(request)
			require.NoError(t, err)
			require.NoError(t, response.Body.Close())
			<-done
			require.Equal(t, http.StatusBadRequest, response.StatusCode)
			require.NotEmpty(t, response.Header.Get("X-Request-ID"))
			var record map[string]any
			require.NoError(t, json.Unmarshal(output.Bytes(), &record))
			require.Equal(t, "http.request_completed", record["event"])
			require.Equal(t, http.StatusBadRequest, int(record["status_code"].(float64)))
			require.NotContains(t, output.String(), "PRIVATE")
			require.NotContains(t, output.String(), "http.request_failed")
		})
	}
}
