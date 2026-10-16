package requestlog

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLogStatusCommittedByFlush(t *testing.T) {
	var output bytes.Buffer

	original := slog.Default()

	slog.SetDefault(slog.New(NewHandler(slog.NewJSONHandler(&output, nil))))
	t.Cleanup(func() { slog.SetDefault(original) })

	done := make(chan struct{})
	wrapped := Wrap(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.(http.Flusher).Flush()
		w.WriteHeader(http.StatusInternalServerError)
	}), func(*http.Request) string { return "GET /flush" })

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wrapped.ServeHTTP(w, r)
		close(done)
	}))
	defer server.Close()

	response, err := server.Client().Get(server.URL)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, response.StatusCode)
	require.NoError(t, response.Body.Close())
	<-done

	var record map[string]any
	require.NoError(t, json.Unmarshal(bytes.TrimSpace(output.Bytes()), &record))
	require.Equal(t, "http.request_completed", record["event"])
	require.Equal(t, response.StatusCode, int(record["status_code"].(float64)))
}
