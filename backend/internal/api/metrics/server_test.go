package metrics

import (
	"context"
	"io"
	"net"
	"net/http"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
)

func TestNewServer(t *testing.T) {
	for _, tc := range []struct {
		name       string
		register   bool
		accept     string
		wantMetric bool
	}{
		{name: "empty export"},
		{name: "ignores requested Prometheus format", accept: "text/plain; version=0.0.4"},
		{name: "ignores requested OTLP format", accept: "application/x-protobuf"},
		{name: "registered metric", register: true, wantMetric: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			registry := prometheus.NewRegistry()

			if tc.register {
				gauge := prometheus.NewGauge(prometheus.GaugeOpts{Name: "test_value", Help: "Test value."})
				gauge.Set(3)
				require.NoError(t, registry.Register(gauge))
			}

			server := NewServer("127.0.0.1:0", registry)
			require.NoError(t, server.Start(context.Background()))
			address := server.listener.Addr().String()

			t.Cleanup(func() { require.NoError(t, server.Stop(context.Background())) })

			request, err := http.NewRequest(http.MethodGet, "http://"+address+"/metrics", nil)
			require.NoError(t, err)
			request.Header.Set("Accept", tc.accept)
			response, err := http.DefaultClient.Do(request)
			require.NoError(t, err)

			defer func() { require.NoError(t, response.Body.Close()) }()

			body, err := io.ReadAll(response.Body)
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, response.StatusCode)
			require.Contains(t, response.Header.Get("Content-Type"), "application/openmetrics-text")
			require.True(t, strings.HasSuffix(string(body), "# EOF\n"))
			require.Equal(t, tc.wantMetric, strings.Contains(string(body), "test_value 3"))

			if !tc.wantMetric {
				require.Equal(t, "# EOF\n", string(body))
			}

			other, err := http.Get("http://" + address + "/elsewhere")
			require.NoError(t, err)

			defer func() { require.NoError(t, other.Body.Close()) }()

			require.Equal(t, http.StatusNotFound, other.StatusCode)
		})
	}
}

func TestStart(t *testing.T) {
	occupied, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	defer func() { require.NoError(t, occupied.Close()) }()

	server := NewServer(occupied.Addr().String(), prometheus.NewRegistry())
	require.ErrorContains(t, server.Start(context.Background()), "listen for metrics")
}

func TestStop(t *testing.T) {
	server := NewServer("127.0.0.1:0", prometheus.NewRegistry())
	require.NoError(t, server.Start(context.Background()))
	address := server.listener.Addr().String()
	require.NoError(t, server.Stop(context.Background()))

	listener, err := net.Listen("tcp", address)
	require.NoError(t, err)
	require.NoError(t, listener.Close())
}
