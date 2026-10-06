package httpcore

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"bookshop/frontend/internal/domain/publishers/entities"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequest(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		method     string
		path       string
		input      any
		status     int
		response   string
		wantError  string
		wantResult entities.Publisher
	}{
		{
			name:       "JSON response",
			method:     http.MethodGet,
			path:       "/publishers/1",
			response:   `{"name":"Издательство"}`,
			wantResult: entities.Publisher{Name: "Издательство"},
		},
		{
			name:       "JSON request and response",
			method:     http.MethodPost,
			path:       "/publishers",
			input:      entities.Publisher{Name: "Издательство"},
			response:   `{"name":"Издательство"}`,
			wantResult: entities.Publisher{Name: "Издательство"},
		},
		{
			name:      "backend error",
			method:    http.MethodGet,
			path:      "/publishers/1",
			status:    http.StatusNotFound,
			response:  `{"error":"not_found"}`,
			wantError: "API 404: not_found",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, tt.method, r.Method)
				assert.Equal(t, tt.path, r.URL.Path)
				if tt.input != nil {
					assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
					var sent entities.Publisher
					if assert.NoError(t, json.NewDecoder(r.Body).Decode(&sent)) {
						assert.Equal(t, tt.input, sent)
					}
				}
				if tt.status != 0 {
					w.WriteHeader(tt.status)
				}
				_, _ = w.Write([]byte(tt.response))
			}))
			defer api.Close()
			client := New(api.URL + "/")
			defer client.CloseIdleConnections()
			var result entities.Publisher

			err := client.Request(context.Background(), tt.method, tt.path, tt.input, &result)

			if tt.wantError != "" {
				require.EqualError(t, err, tt.wantError)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.wantResult, result)
		})
	}
}
