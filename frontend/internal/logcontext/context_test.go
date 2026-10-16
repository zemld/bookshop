package logcontext

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestNewHandler(t *testing.T) {
	var output bytes.Buffer
	logger := slog.New(NewHandler(slog.NewJSONHandler(&output, nil))).
		With("service", "bookshop-web").WithGroup("detail").With("kind", "test")
	var group sync.WaitGroup
	const count = 16
	for range count {
		group.Add(1)
		go func() {
			defer group.Done()
			id := uuid.NewString()
			logger.InfoContext(StoreRequest(context.Background(), id, "GET", "GET /books/{id}"), "completed", "id", id)
		}()
	}
	group.Wait()
	var seen sync.Map
	lines := bytes.Split(bytes.TrimSpace(output.Bytes()), []byte("\n"))
	require.Len(t, lines, count)
	for _, line := range lines {
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
