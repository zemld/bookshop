package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	core "bookshop/frontend/internal/adapters/httpcore"
	"bookshop/frontend/internal/domain/books/entities"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestBookMethods(t *testing.T) {
	t.Parallel()
	id := uuid.MustParse("4e643615-2a28-40a5-8e7d-fbe6dd2d498d")
	for _, tt := range []struct {
		name, method, path string
		fail               bool
	}{
		{"list", "GET", "/books", false},
		{"get", "GET", "/books/" + id.String(), false},
		{"create", "POST", "/books", false},
		{"update", "PUT", "/books/" + id.String(), false},
		{"delete", "DELETE", "/books/" + id.String(), false},
		{"backend rejection", "POST", "/books", true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, tt.method, r.Method)
				require.Equal(t, tt.path, r.URL.Path)
				if tt.method == "POST" || tt.method == "PUT" {
					var sent entities.Book
					require.NoError(t, json.NewDecoder(r.Body).Decode(&sent))
					require.Equal(t, entities.Year(2000), sent.Year)
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
			b := &Client{Core: c}
			ctx := context.Background()
			var err error
			var book entities.Book
			switch tt.name {
			case "list":
				var books []entities.Book
				books, err = b.ListBooks(ctx)
				if err == nil {
					require.Equal(t, id, books[0].ID)
				}
			case "get":
				book, err = b.GetBook(ctx, id)
			case "create", "backend rejection":
				book, err = b.CreateBook(ctx, entities.Book{Year: 2000})
			case "update":
				book, err = b.UpdateBook(ctx, id, entities.Book{Year: 2000})
			case "delete":
				err = b.DeleteBook(ctx, id)
			}
			if tt.fail {
				require.EqualError(t, err, "API 400: invalid")
			} else {
				require.NoError(t, err)
				if tt.name != "list" && tt.name != "delete" {
					require.Equal(t, id, book.ID)
				}
			}
		})
	}
}
