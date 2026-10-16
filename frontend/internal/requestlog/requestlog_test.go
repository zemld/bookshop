package requestlog

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type lockedBuffer struct {
	sync.Mutex
	bytes.Buffer
}

func (b *lockedBuffer) Write(p []byte) (int, error) {
	b.Lock()
	defer b.Unlock()
	return b.Buffer.Write(p)
}

func TestKeepContextAndHandlerAttributesSeparate(t *testing.T) {
	var output lockedBuffer
	logger := slog.New(NewHandler(slog.NewJSONHandler(&output, nil))).With("service", "bookshop-web").WithGroup("detail").With("kind", "test")
	const count = 20
	var group sync.WaitGroup
	for range count {
		group.Add(1)
		go func() {
			defer group.Done()
			id := uuid.NewString()
			logger.InfoContext(StoreRequest(context.Background(), id, "GET", "GET /books/{id}"), "done", "id", id)
		}()
	}
	group.Wait()
	var seen sync.Map
	for _, line := range bytes.Split(bytes.TrimSpace(output.Bytes()), []byte("\n")) {
		var record struct {
			Service string `json:"service"`
			Detail  struct {
				Kind      string `json:"kind"`
				ID        string `json:"id"`
				RequestID string `json:"request_id"`
			} `json:"detail"`
		}
		require.NoError(t, json.Unmarshal(line, &record))
		require.Equal(t, "bookshop-web", record.Service)
		require.Equal(t, "test", record.Detail.Kind)
		require.Equal(t, record.Detail.ID, record.Detail.RequestID)
		_, loaded := seen.LoadOrStore(record.Detail.ID, true)
		require.False(t, loaded)
	}
}

func TestKeepResponseWriterCapabilities(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /flush", func(w http.ResponseWriter, r *http.Request) {
		w.(http.Flusher).Flush()
		_, _ = w.Write([]byte("complete"))
	})
	handler := Wrap(mux, func(r *http.Request) string {
		_, route := mux.Handler(r)
		return route
	})
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, httptest.NewRequest("GET", "/flush", nil))
	require.Equal(t, http.StatusOK, res.Code)
	require.Equal(t, "complete", res.Body.String())
	require.True(t, res.Flushed)
}

func TestLogStatusCommittedByFlush(t *testing.T) {
	var output lockedBuffer
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
