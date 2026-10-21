package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	for _, tc := range []struct {
		name        string
		config      string
		address     string
		metrics     string
		wantMetrics string
		wantError   bool
	}{
		{name: "file and environment override", config: "service:\n  name: bookshop-api\nserver:\n  address: ':8080'\nmetrics:\n  address: '127.0.0.1:9090'\n", address: ":8082", metrics: "127.0.0.1:9190", wantMetrics: "127.0.0.1:9190"},
		{name: "file metrics address", config: "service:\n  name: bookshop-api\nserver:\n  address: ':8080'\nmetrics:\n  address: '127.0.0.1:9090'\n", address: ":8082", wantMetrics: "127.0.0.1:9090"},
		{name: "missing metrics address", config: "service:\n  name: bookshop-api\nserver:\n  address: ':8080'\n", wantError: true},
		{name: "missing file", wantError: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "config.yaml")
			if tc.config != "" {
				require.NoError(t, os.WriteFile(path, []byte(tc.config), 0600))
			}

			t.Setenv("API_ADDR", tc.address)
			t.Setenv("METRICS_ADDR", tc.metrics)

			cfg, err := Load(path)
			if tc.wantError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, "bookshop-api", cfg.Service.Name)
			require.Equal(t, tc.address, cfg.Server.Address)
			require.Equal(t, tc.wantMetrics, cfg.Metrics.Address)
		})
	}
}
