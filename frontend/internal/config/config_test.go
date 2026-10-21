package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	for _, tc := range []struct {
		name, config, address, apiURL, metrics, wantMetrics, wantError string
	}{
		{"file and environment overrides", "service:\n  name: bookshop-web\nserver:\n  address: ':8081'\nmetrics:\n  address: '127.0.0.1:9091'\napi:\n  url: ''\n", ":8083", "http://backend:8080", "127.0.0.1:9191", "127.0.0.1:9191", ""},
		{"file metrics address", "service:\n  name: bookshop-web\nserver:\n  address: ':8081'\nmetrics:\n  address: '127.0.0.1:9091'\napi:\n  url: ''\n", ":8083", "http://backend:8080", "", "127.0.0.1:9091", ""},
		{"requires metrics address", "service:\n  name: bookshop-web\nserver:\n  address: ':8081'\napi:\n  url: ''\n", ":8083", "http://backend:8080", "", "", "metrics.address"},
		{"requires API URL", "service:\n  name: bookshop-web\nserver:\n  address: ':8081'\nmetrics:\n  address: '127.0.0.1:9091'\n", "", "", "", "", "API_URL"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "config.yaml")
			require.NoError(t, os.WriteFile(path, []byte(tc.config), 0600))
			t.Setenv("WEB_ADDR", tc.address)
			t.Setenv("API_URL", tc.apiURL)
			t.Setenv("METRICS_ADDR", tc.metrics)
			cfg, err := Load(path)
			if tc.wantError != "" {
				require.ErrorContains(t, err, tc.wantError)
				return
			}
			require.NoError(t, err)
			require.Equal(t, "bookshop-web", cfg.Service.Name)
			require.Equal(t, tc.address, cfg.Server.Address)
			require.Equal(t, tc.apiURL, cfg.API.URL)
			require.Equal(t, tc.wantMetrics, cfg.Metrics.Address)
		})
	}
}
