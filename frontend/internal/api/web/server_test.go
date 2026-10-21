package web

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"bookshop/frontend/internal/telemetry"

	"github.com/stretchr/testify/require"
)

func TestHTTPServerLogsListenerFailure(t *testing.T) {
	for _, tc := range []struct {
		name, address, event string
	}{
		{name: "unexpected listener closure", address: "127.0.0.1:0", event: "http.listener_failed"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var output bytes.Buffer
			s := NewHTTPServer(tc.address, "127.0.0.1:0", http.NotFoundHandler(), http.NotFoundHandler())
			s.SetLogger(slog.New(slog.NewJSONHandler(&output, nil)))
			require.NoError(t, s.Start(context.Background()))
			t.Cleanup(func() {
				require.NoError(t, s.metrics.Shutdown(context.Background()))
				<-s.metricsDone
			})
			require.NoError(t, s.listener.Close())
			require.Error(t, <-s.done)
			require.Contains(t, output.String(), `"event":"`+tc.event+`"`)
			require.Contains(t, output.String(), `"address":"`+tc.address+`"`)
		})
	}
}

func TestHTTPServerStart(t *testing.T) {
	for _, tc := range []struct {
		name, accept string
		occupied     bool
	}{
		{name: "without Accept"},
		{name: "Prometheus Accept", accept: "text/plain; version=0.0.4"},
		{name: "OTLP Accept", accept: "application/x-protobuf"},
		{name: "occupied metrics address rolls back web listener", occupied: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if tc.occupied {
				first, err := net.Listen("tcp", "127.0.0.1:0")
				require.NoError(t, err)
				firstAddress := first.Addr().String()
				require.NoError(t, first.Close())
				occupied, err := net.Listen("tcp", "127.0.0.1:0")
				require.NoError(t, err)
				defer func() { require.NoError(t, occupied.Close()) }()
				s := NewHTTPServer(firstAddress, occupied.Addr().String(), http.NotFoundHandler(), http.NotFoundHandler())
				require.ErrorContains(t, s.Start(context.Background()), "listen for metrics")
				rebound, err := net.Listen("tcp", firstAddress)
				require.NoError(t, err)
				require.NoError(t, rebound.Close())
				return
			}

			metrics, err := telemetry.New()
			require.NoError(t, err)
			t.Cleanup(func() { require.NoError(t, metrics.Provider.Shutdown(context.Background())) })
			s := NewHTTPServer("127.0.0.1:0", "127.0.0.1:0", http.NotFoundHandler(), metrics.Handler)
			require.NoError(t, s.Start(context.Background()))
			t.Cleanup(func() { require.NoError(t, s.Stop(context.Background())) })
			req, err := http.NewRequest(http.MethodGet, "http://"+s.metricsListener.Addr().String()+"/metrics", nil)
			require.NoError(t, err)
			req.Header.Set("Accept", tc.accept)
			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer func() { require.NoError(t, resp.Body.Close()) }()
			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, resp.StatusCode)
			require.Equal(t, "application/openmetrics-text; version=1.0.0; charset=utf-8", resp.Header.Get("Content-Type"))
			require.Contains(t, string(body), "go_goroutines")
			require.Contains(t, string(body), "process_cpu_seconds_total")
			require.True(t, strings.HasSuffix(string(body), "# EOF\n"))
			main, err := http.Get("http://" + s.listener.Addr().String() + "/metrics")
			require.NoError(t, err)
			defer func() { require.NoError(t, main.Body.Close()) }()
			require.Equal(t, http.StatusNotFound, main.StatusCode)
		})
	}
}

func TestHTTPServerStop(t *testing.T) {
	for _, tc := range []struct {
		name string
		run  func(*testing.T)
	}{
		{name: "drains active web request", run: func(t *testing.T) {
			entered, release := make(chan struct{}), make(chan struct{})
			s := NewHTTPServer("127.0.0.1:0", "127.0.0.1:0", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(200)
				w.(http.Flusher).Flush()
				close(entered)
				<-release
				_, _ = w.Write([]byte("done"))
			}), http.NotFoundHandler())
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
		}},
		{name: "forces closure at drain timeout", run: func(t *testing.T) {
			entered, release := make(chan struct{}), make(chan struct{})
			s := NewHTTPServer("127.0.0.1:0", "127.0.0.1:0", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(200)
				w.(http.Flusher).Flush()
				close(entered)
				<-release
			}), http.NotFoundHandler())
			require.NoError(t, s.Start(context.Background()))
			resp, err := http.Get("http://" + s.listener.Addr().String())
			require.NoError(t, err)
			defer func() { _ = resp.Body.Close() }()
			<-entered
			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
			defer cancel()
			require.ErrorIs(t, s.Stop(ctx), context.DeadlineExceeded)
			close(release)
		}},
		{name: "closes both listeners", run: func(t *testing.T) {
			s := NewHTTPServer("127.0.0.1:0", "127.0.0.1:0", http.NotFoundHandler(), http.NotFoundHandler())
			require.NoError(t, s.Start(context.Background()))
			addresses := []string{s.listener.Addr().String(), s.metricsListener.Addr().String()}
			require.NoError(t, s.Stop(context.Background()))
			for _, addr := range addresses {
				listener, err := net.Listen("tcp", addr)
				require.NoError(t, err)
				require.NoError(t, listener.Close())
			}
		}},
		{name: "slow metrics does not delay web shutdown or drain", run: func(t *testing.T) {
			mainEntered, release := make(chan struct{}), make(chan struct{})
			s := NewHTTPServer("127.0.0.1:0", "127.0.0.1:0", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.(http.Flusher).Flush()
				close(mainEntered)
				<-release
				_, _ = w.Write([]byte("drained"))
			}), http.NotFoundHandler())
			metricsEntered := make(chan struct{})
			metricsHandler := s.metrics.Handler
			s.metrics.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				close(metricsEntered)
				metricsHandler.ServeHTTP(w, r)
			})
			require.NoError(t, s.Start(context.Background()))

			main, err := http.Get("http://" + s.listener.Addr().String() + "/")
			require.NoError(t, err)
			defer func() { require.NoError(t, main.Body.Close()) }()
			<-mainEntered
			t.Cleanup(func() {
				select {
				case <-release:
				default:
					close(release)
				}
			})

			slow, err := net.Dial("tcp", s.metricsListener.Addr().String())
			require.NoError(t, err)
			defer func() { require.NoError(t, slow.Close()) }()
			_, err = io.WriteString(slow, "GET /metrics HTTP/1.1\r\nHost: localhost\r\nContent-Length: 1\r\n\r\n")
			require.NoError(t, err)
			<-metricsEntered

			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			finished := make(chan error, 1)
			go func() { finished <- s.Stop(ctx) }()

			require.Eventually(t, func() bool {
				conn, err := net.DialTimeout("tcp", s.listener.Addr().String(), 20*time.Millisecond)
				if err == nil {
					_ = conn.Close()
				}
				return err != nil
			}, 500*time.Millisecond, 10*time.Millisecond)
			close(release)
			body, err := io.ReadAll(main.Body)
			require.NoError(t, err)
			require.Equal(t, "drained", string(body))
			require.ErrorIs(t, <-finished, context.DeadlineExceeded)
		}},
	} {
		t.Run(tc.name, tc.run)
	}
}
