package web

import (
	"context"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestHTTPServerDrain(t *testing.T) {
	t.Parallel()
	entered, release := make(chan struct{}), make(chan struct{})
	s := NewHTTPServer("127.0.0.1:0", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(200)
		w.(http.Flusher).Flush()
		close(entered)
		<-release
		_, _ = w.Write([]byte("done"))
	}))
	require.NoError(t, s.Start(context.Background()))
	resp, err := http.Get("http://" + s.listener.Addr().String())
	require.NoError(t, err)
	defer func() { require.NoError(t, resp.Body.Close()) }()
	<-entered
	finished := make(chan error, 1)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		finished <- s.Stop(ctx)
	}()
	close(release)
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, "done", string(body))
	require.NoError(t, <-finished)
}

func TestHTTPServerDrainTimeout(t *testing.T) {
	t.Parallel()
	entered, release := make(chan struct{}), make(chan struct{})
	s := NewHTTPServer("127.0.0.1:0", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(200)
		w.(http.Flusher).Flush()
		close(entered)
		<-release
	}))
	require.NoError(t, s.Start(context.Background()))
	resp, err := http.Get("http://" + s.listener.Addr().String())
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	<-entered
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	require.ErrorIs(t, s.Stop(ctx), context.DeadlineExceeded)
	close(release)
}
