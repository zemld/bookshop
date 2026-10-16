package web

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestStop(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name            string
		timeout         time.Duration
		completeRequest bool
	}{
		{name: "successful drain", timeout: time.Second, completeRequest: true},
		{name: "deadline timeout", timeout: 20 * time.Millisecond},
	} {
		t.Run(tc.name, func(t *testing.T) {
			entered, release := make(chan struct{}), make(chan struct{})
			s := NewHTTPServer("127.0.0.1:0", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(200)
				w.(http.Flusher).Flush()
				close(entered)
				<-release
				if tc.completeRequest {
					_, _ = w.Write([]byte("done"))
				}
			}), slog.New(slog.NewTextHandler(io.Discard, nil)))
			require.NoError(t, s.Start(context.Background()))
			resp, err := http.Get("http://" + s.listener.Addr().String())
			require.NoError(t, err)
			if tc.completeRequest {
				defer func() { require.NoError(t, resp.Body.Close()) }()
			} else {
				defer func() { _ = resp.Body.Close() }()
			}
			<-entered
			ctx, cancel := context.WithTimeout(context.Background(), tc.timeout)
			defer cancel()
			if tc.completeRequest {
				finished := make(chan error, 1)
				go func() { finished <- s.Stop(ctx) }()
				close(release)
				body, err := io.ReadAll(resp.Body)
				require.NoError(t, err)
				require.Equal(t, "done", string(body))
				require.NoError(t, <-finished)
				return
			}
			require.ErrorIs(t, s.Stop(ctx), context.DeadlineExceeded)
			close(release)
		})
	}
}

func TestStart(t *testing.T) {
	for _, tc := range []struct {
		name     string
		occupied bool
	}{
		{name: "occupied port", occupied: true},
		{name: "unexpected listener close"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var output bytes.Buffer
			addr := "127.0.0.1:0"
			if tc.occupied {
				listener, err := net.Listen("tcp", addr)
				require.NoError(t, err)
				defer func() { require.NoError(t, listener.Close()) }()
				addr = listener.Addr().String()
			}
			s := NewHTTPServer(addr, http.NotFoundHandler(), slog.New(slog.NewJSONHandler(&output, nil)))
			err := s.Start(context.Background())
			if tc.occupied {
				require.Error(t, err)
				require.Empty(t, output.String())
				return
			}
			require.NoError(t, err)
			require.NoError(t, s.listener.Close())
			require.Error(t, <-s.done)
			var record map[string]any
			require.NoError(t, json.Unmarshal(output.Bytes(), &record))
			require.Equal(t, "http.listener_failed", record["event"])
			require.Equal(t, s.server.Addr, record["address"])
			require.Equal(t, "ERROR", record["level"])
		})
	}
}
