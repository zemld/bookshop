package books

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	bookhttp "bookshop/frontend/internal/adapters/books/http"
	core "bookshop/frontend/internal/adapters/httpcore"
	publisherhttp "bookshop/frontend/internal/adapters/publishers/http"

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
				calls++
				if calls == 1 {
					require.Equal(t, "/publishers", r.URL.Path)
				}
				if r.URL.Path == tt.failAt {
					w.WriteHeader(http.StatusBadRequest)
					_, _ = w.Write([]byte(`{"error":"bad"}`))
					return
				}
				if r.URL.Path == "/publishers" {
					_, _ = w.Write([]byte(`[{"id":"` + id.String() + `","name":"Имя"}]`))
				} else {
					require.Equal(t, "/books/"+id.String(), r.URL.Path)
					_, _ = w.Write([]byte(`{"id":"` + id.String() + `","year":2000}`))
				}
			}))
			defer api.Close()
			c := core.New(api.URL)
			defer c.Stop()
			s := Service{Books: &bookhttp.Client{Core: c}, Publishers: &publisherhttp.Client{Core: c}}
			form, err := s.LoadBookForm(context.Background(), id)
			require.Equal(t, tt.wantCalls, calls)
			if tt.failAt != "" {
				require.EqualError(t, err, "API 400: bad")
				return
			}
			require.NoError(t, err)
			require.Equal(t, id, form.Book.ID)
			require.Equal(t, "Имя", form.Publishers[0].Name)
		})
	}
}
