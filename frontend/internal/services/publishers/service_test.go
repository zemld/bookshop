package publishers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	apiClient "bookshop/frontend/internal/adapters/api"
	"bookshop/frontend/internal/domain/publishers/entities"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestPublisherOperations(t *testing.T) {
	t.Parallel()
	id := uuid.MustParse("4e643615-2a28-40a5-8e7d-fbe6dd2d498d")
	for _, tt := range []struct {
		name, method, path string
		fail               bool
	}{
		{name: "create", method: "POST", path: "/publishers"},
		{name: "update", method: "PUT", path: "/publishers/" + id.String()},
		{name: "create failure", method: "POST", path: "/publishers", fail: true},
		{name: "update failure", method: "PUT", path: "/publishers/" + id.String(), fail: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("X-Request-ID", "51aa5ea3-0653-49ad-a3b5-e4c3b7ade70c")
				require.Equal(t, tt.method, r.Method)
				require.Equal(t, tt.path, r.URL.Path)
				var sent entities.Publisher
				require.NoError(t, json.NewDecoder(r.Body).Decode(&sent))
				require.Equal(t, "Имя", sent.Name)
				if tt.fail {
					w.WriteHeader(400)
					_, _ = w.Write([]byte(`{"error":"invalid_input"}`))
					return
				}
				_, _ = w.Write([]byte(`{"id":"` + id.String() + `","name":"Имя"}`))
			}))
			defer api.Close()
			c, err := apiClient.New(api.URL)
			require.NoError(t, err)
			defer c.Stop()
			s := Service{Publishers: c}
			var saved entities.Publisher
			if tt.method == "POST" {
				saved, err = s.CreatePublisher(context.Background(), entities.Publisher{Name: " Имя "})
			} else {
				saved, err = s.UpdatePublisher(context.Background(), id, entities.Publisher{Name: " Имя "})
			}
			if tt.fail {
				require.EqualError(t, err, "API 400: invalid_input")
				return
			}
			require.NoError(t, err)
			require.Equal(t, id, saved.ID)
		})
	}
}
