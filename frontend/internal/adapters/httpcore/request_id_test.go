package httpcore

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"bookshop/frontend/internal/logcontext"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestRequestIDTransportRoundTrip(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name string
		id   string
		want string
		fail bool
	}{
		{name: "context overrides header", id: uuid.NewString(), want: "context"},
		{name: "no context preserves header", want: "original"},
		{name: "transport failure without request ID", fail: true},
		{name: "transport failure with request ID", id: uuid.NewString(), fail: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if tc.fail {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				request, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://example.invalid/private?token=secret", nil)
				require.NoError(t, err)
				if tc.id != "" {
					request = request.WithContext(logcontext.StoreRequest(request.Context(), tc.id, "GET", "GET /books"))
				}
				transport := requestIDTransport{next: http.DefaultTransport}
				_, err = transport.RoundTrip(request)
				require.ErrorIs(t, err, context.Canceled)
				require.True(t, errors.Is(err, context.Canceled))
				require.Contains(t, err.Error(), "send API request")
				require.NotContains(t, err.Error(), "private")
				require.NotContains(t, err.Error(), "secret")
				return
			}
			received := make(chan string, 1)
			api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				received <- r.Header.Get("X-Request-ID")
				w.WriteHeader(http.StatusNoContent)
			}))
			defer api.Close()
			request, err := http.NewRequest(http.MethodGet, api.URL, nil)
			require.NoError(t, err)
			request.Header.Set("X-Request-ID", "original")
			if tc.id != "" {
				request = request.WithContext(logcontext.StoreRequest(request.Context(), tc.id, "GET", "GET /books"))
			}
			client := &http.Client{Transport: requestIDTransport{next: http.DefaultTransport}}
			response, err := client.Do(request)
			require.NoError(t, err)
			require.NoError(t, response.Body.Close())
			require.Equal(t, http.StatusNoContent, response.StatusCode)
			if tc.want == "context" {
				require.Equal(t, tc.id, <-received)
			} else {
				require.Equal(t, tc.want, <-received)
			}
			require.Equal(t, "original", request.Header.Get("X-Request-ID"))
		})
	}
}

func TestRequestIDTransportCloseIdleConnections(t *testing.T) {
	t.Parallel()
	closed := make(chan struct{}, 1)
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	server.Config.ConnState = func(_ net.Conn, state http.ConnState) {
		if state == http.StateClosed {
			closed <- struct{}{}
		}
	}
	server.Start()
	defer server.Close()
	transport := http.DefaultTransport.(*http.Transport).Clone()
	client := &http.Client{Transport: requestIDTransport{next: transport}}
	response, err := client.Get(server.URL)
	require.NoError(t, err)
	require.NoError(t, response.Body.Close())
	client.CloseIdleConnections()
	select {
	case <-closed:
	case <-time.After(2 * time.Second):
		t.Fatal("idle connection remained open")
	}
}
