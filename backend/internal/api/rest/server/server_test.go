package server

import (
	"context"
	"io"
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
			s := newServer("127.0.0.1:0", handler)
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

func TestStartRejectsOccupiedPort(t *testing.T) {
	t.Parallel()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	defer func() { require.NoError(t, listener.Close()) }()

	s := newServer(listener.Addr().String(), http.NotFoundHandler())

	err = s.Start(context.Background())

	require.Error(t, err)
}
