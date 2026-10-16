package server

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

	for _, tt := range []struct {
		name        string
		timeout     time.Duration
		wantTimeout bool
	}{
		{name: "drains active request", timeout: time.Second},
		{name: "forces bounded stop", timeout: 30 * time.Millisecond, wantTimeout: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			release := make(chan struct{})
			started := make(chan struct{})
			handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.(http.Flusher).Flush()
				close(started)
				<-release

				_, _ = w.Write([]byte("done"))
			})
			s := newServer("127.0.0.1:0", handler, slog.New(slog.NewTextHandler(io.Discard, nil)))
			require.NoError(t, s.Start(context.Background()))

			defer func() { _ = s.http.Close() }()

			res, err := http.Get("http://" + s.listener.Addr().String() + "/hold")
			require.NoError(t, err)

			defer func() { require.NoError(t, res.Body.Close()) }()

			<-started

			ctx, cancel := context.WithTimeout(context.Background(), tt.timeout)
			defer cancel()

			stopped := make(chan error, 1)

			go func() { stopped <- s.Stop(ctx) }()

			if !tt.wantTimeout {
				close(release)
			}

			err = <-stopped

			if tt.wantTimeout {
				close(release)
				require.ErrorIs(t, err, context.DeadlineExceeded)

				return
			}

			require.NoError(t, err)
			body, err := io.ReadAll(res.Body)
			require.NoError(t, err)
			require.Equal(t, "done", string(body))
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

			s := newServer(addr, http.NotFoundHandler(), slog.New(slog.NewJSONHandler(&output, nil)))

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
			require.Equal(t, s.http.Addr, record["address"])
			require.Equal(t, "ERROR", record["level"])
		})
	}
}
