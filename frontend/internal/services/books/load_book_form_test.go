package books

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	apiClient "bookshop/frontend/internal/adapters/api"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestLoadBookForm(t *testing.T) {
	t.Parallel()
	id := uuid.MustParse("4e643615-2a28-40a5-8e7d-fbe6dd2d498d")
	for _, tt := range []struct {
		name      string
		failAt    string
		wantCalls int
	}{
		{name: "publisher and book", wantCalls: 2},
		{name: "publisher error stops book request", failAt: "/publishers", wantCalls: 1},
		{name: "book error", failAt: "/books/" + id.String(), wantCalls: 2},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			calls := 0
			api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("X-Request-ID", "51aa5ea3-0653-49ad-a3b5-e4c3b7ade70c")
				calls++
				if calls == 1 {
					require.Equal(t, "/publishers", r.URL.Path)
				}
				if r.URL.Path == tt.failAt {
					w.WriteHeader(http.StatusBadRequest)
					_, _ = w.Write([]byte(`{"error":"invalid_input"}`))
					return
				}
				if r.URL.Path == "/publishers" {
					_, _ = w.Write([]byte(`[{"id":"` + id.String() + `","name":"Имя"}]`))
				} else {
					require.Equal(t, "/books/"+id.String(), r.URL.Path)
					_, _ = w.Write([]byte(`{"id":"` + id.String() + `","author":"Автор","name":"Книга","year":2000,"price":1,"publisherId":"` + id.String() + `","publicationYear":2001,"quantity":2}`))
				}
			}))
			defer api.Close()
			c, err := apiClient.New(api.URL)
			require.NoError(t, err)
			defer c.Stop()
			s := Service{Books: c, Publishers: c}
			form, err := s.LoadBookForm(context.Background(), id)
			require.Equal(t, tt.wantCalls, calls)
			if tt.failAt != "" {
				require.EqualError(t, err, "API 400: invalid_input")
				return
			}
			require.NoError(t, err)
			require.Equal(t, id, form.Book.ID)
			require.Equal(t, "Имя", form.Publishers[0].Name)
		})
	}
}
