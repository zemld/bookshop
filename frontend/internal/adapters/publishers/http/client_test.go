package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	core "bookshop/frontend/internal/adapters/httpcore"
	"bookshop/frontend/internal/domain/publishers/entities"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestPublisherMethods(t *testing.T) {
	t.Parallel()
	id := uuid.MustParse("4e643615-2a28-40a5-8e7d-fbe6dd2d498d")
	for _, tt := range []struct {
		name, method, path string
		fail               bool
	}{
		{"list", "GET", "/publishers", false},
		{"get", "GET", "/publishers/" + id.String(), false},
		{"create", "POST", "/publishers", false},
		{"update", "PUT", "/publishers/" + id.String(), false},
		{"delete", "DELETE", "/publishers/" + id.String(), false},
		{"backend rejection", "POST", "/publishers", true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, tt.method, r.Method)
				require.Equal(t, tt.path, r.URL.Path)
				if tt.method == "POST" || tt.method == "PUT" {
					var sent entities.Publisher
					require.NoError(t, json.NewDecoder(r.Body).Decode(&sent))
					require.Equal(t, "Имя", sent.Name)
				}
				if tt.fail {
					w.WriteHeader(400)
					_, _ = w.Write([]byte(`{"error":"invalid"}`))
				} else if tt.name == "list" {
					_, _ = w.Write([]byte(`[{"id":"` + id.String() + `"}]`))
				} else if tt.name != "delete" {
					_, _ = w.Write([]byte(`{"id":"` + id.String() + `"}`))
				}
			}))
			defer api.Close()
			c := core.New(api.URL)
			defer c.Stop()
			p := &Client{Core: c}
			ctx := context.Background()
			var err error
			var publisher entities.Publisher
			switch tt.name {
			case "list":
				var publishers []entities.Publisher
				publishers, err = p.ListPublishers(ctx)
				if err == nil {
					require.Equal(t, id, publishers[0].ID)
				}
			case "get":
				publisher, err = p.GetPublisher(ctx, id)
			case "create", "backend rejection":
				publisher, err = p.CreatePublisher(ctx, entities.Publisher{Name: "Имя"})
			case "update":
				publisher, err = p.UpdatePublisher(ctx, id, entities.Publisher{Name: "Имя"})
			case "delete":
				err = p.DeletePublisher(ctx, id)
			}
			if tt.fail {
				require.EqualError(t, err, "API 400: invalid")
			} else {
				require.NoError(t, err)
				if tt.name != "list" && tt.name != "delete" {
					require.Equal(t, id, publisher.ID)
				}
			}
		})
	}
}
