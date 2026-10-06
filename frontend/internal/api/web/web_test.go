package web

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	bookhttp "bookshop/frontend/internal/adapters/books/http"
	core "bookshop/frontend/internal/adapters/httpcore"
	publisherhttp "bookshop/frontend/internal/adapters/publishers/http"
	bookservice "bookshop/frontend/internal/services/books"
	publisherservice "bookshop/frontend/internal/services/publishers"

	"github.com/stretchr/testify/require"
)

const testID = "4e643615-2a28-40a5-8e7d-fbe6dd2d498d"

func createTestHandler(apiURL string) http.Handler {
	c := core.New(apiURL)
	b := &bookhttp.Client{Core: c}
	p := &publisherhttp.Client{Core: c}
	bs := &bookservice.Service{Books: b, Publishers: p}
	ps := &publisherservice.Service{Publishers: p}
	return (Server{
		Books: b, DeleteBook: b, Publishers: p, GetPublisher: p, DeletePublisher: p,
		LoadBookForm: bs, CreateBook: b, UpdateBook: b,
		CreatePublisher: ps, UpdatePublisher: ps,
	}).CreateHandler()
}

func TestPublisherForms(t *testing.T) {
	t.Parallel()
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "POST", r.Method)
		require.Equal(t, "/publishers", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"` + testID + `","name":"Имя"}`))
	}))
	defer api.Close()
	req := httptest.NewRequest("POST", "/publishers", strings.NewReader(url.Values{"name": {" Имя "}}.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res := httptest.NewRecorder()
	createTestHandler(api.URL).ServeHTTP(res, req)
	require.Equal(t, http.StatusSeeOther, res.Code)
	require.Equal(t, "/publishers/"+testID, res.Header().Get("Location"))
}

func TestBookInvalidForm(t *testing.T) {
	t.Parallel()
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "GET /publishers", r.Method+" "+r.URL.Path)
		_, _ = w.Write([]byte(`[]`))
	}))
	t.Cleanup(api.Close)
	req := httptest.NewRequest("POST", "/books", strings.NewReader(url.Values{"year": {"0"}}.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res := httptest.NewRecorder()
	createTestHandler(api.URL).ServeHTTP(res, req)
	require.Equal(t, http.StatusBadRequest, res.Code)
	require.Contains(t, res.Body.String(), `lang="ru"`)
}

func TestUnknownPage(t *testing.T) {
	t.Parallel()
	res := httptest.NewRecorder()
	renderPage(res, http.StatusOK, page{Kind: 99, Title: "Книги"})
	require.Equal(t, http.StatusInternalServerError, res.Code)
	require.NotContains(t, res.Body.String(), "<html")
}

func TestBookFormAndRoutes(t *testing.T) {
	t.Parallel()
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method + " " + r.URL.Path {
		case "GET /publishers":
			_, _ = w.Write([]byte(`[{"id":"` + testID + `","name":"Издательство"}]`))
		case "GET /books/" + testID:
			_, _ = w.Write([]byte(`{"id":"` + testID + `","author":"Автор","name":"Книга","year":2000,"price":2,"publisherId":"` + testID + `","publicationYear":2001,"quantity":3}`))
		case "POST /books":
			_, _ = w.Write([]byte(`{"id":"` + testID + `"}`))
		default:
			http.Error(w, "unexpected route", http.StatusNotFound)
		}
	}))
	defer api.Close()
	routes := createTestHandler(api.URL)
	for _, tt := range []struct{ path, want string }{
		{"/books/new", `action="/books"`},
		{"/books/" + testID, `value="` + testID + `" selected`},
	} {
		res := httptest.NewRecorder()
		routes.ServeHTTP(res, httptest.NewRequest("GET", tt.path, nil))
		require.Equal(t, http.StatusOK, res.Code)
		require.Contains(t, res.Body.String(), tt.want)
	}
	values := url.Values{"author": {"Автор"}, "name": {"Книга"}, "year": {"2000"}, "price": {"2"}, "publisherId": {testID}, "publicationYear": {"2001"}, "quantity": {"3"}}
	req := httptest.NewRequest("POST", "/books", strings.NewReader(values.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res := httptest.NewRecorder()
	routes.ServeHTTP(res, req)
	require.Equal(t, http.StatusSeeOther, res.Code)
	require.Equal(t, "/books/"+testID, res.Header().Get("Location"))
}

func TestBookFormSyntaxAndBackendValidation(t *testing.T) {
	t.Parallel()
	values := url.Values{"author": {"Автор"}, "name": {"Книга"}, "year": {"2001"}, "price": {"-2"}, "publisherId": {testID}, "publicationYear": {"2000"}, "quantity": {"3"}}
	for _, tt := range []struct {
		name, key, value string
		wantBackend      bool
	}{
		{name: "backend owns business rejection", wantBackend: true},
		{name: "invalid year syntax", key: "year", value: "bad"},
		{name: "invalid publication year syntax", key: "publicationYear", value: "bad"},
		{name: "invalid price syntax", key: "price", value: "bad"},
		{name: "invalid quantity syntax", key: "quantity", value: "bad"},
		{name: "invalid publisher id syntax", key: "publisherId", value: "bad"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method == "GET" && r.URL.Path == "/publishers" {
					_, _ = w.Write([]byte(`[]`))
					return
				}
				require.True(t, tt.wantBackend, "malformed input must not reach backend")
				require.Equal(t, "POST", r.Method)
				require.Equal(t, "/books", r.URL.Path)
				var sent map[string]any
				require.NoError(t, json.NewDecoder(r.Body).Decode(&sent))
				require.Equal(t, float64(-2), sent["price"])
				require.Equal(t, float64(2000), sent["publicationYear"])
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte(`{"error":"book_price_negative"}`))
			}))
			defer api.Close()
			input := url.Values{}
			for k, v := range values {
				input[k] = append([]string(nil), v...)
			}
			if tt.key != "" {
				input.Set(tt.key, tt.value)
			}
			req := httptest.NewRequest("POST", "/books", strings.NewReader(input.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			res := httptest.NewRecorder()
			createTestHandler(api.URL).ServeHTTP(res, req)
			require.Equal(t, http.StatusBadRequest, res.Code)
			if tt.wantBackend {
				require.Contains(t, res.Body.String(), "Цена не может быть отрицательной.")
				require.NotContains(t, res.Body.String(), "Год выпуска версии не может быть раньше года выхода.")
				require.NotContains(t, res.Body.String(), "book_price_negative")
				require.NotContains(t, res.Body.String(), "API 400:")
			} else {
				require.Contains(t, res.Body.String(), `id="`+tt.key+`-error"`)
			}
		})
	}
}

func TestInvalidRouteID(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct{ method, path string }{
		{"GET", "/books/not-a-uuid"},
		{"POST", "/books/not-a-uuid/delete"},
		{"GET", "/publishers/not-a-uuid"},
		{"POST", "/publishers/not-a-uuid/delete"},
	} {
		req := httptest.NewRequest(tt.method, tt.path, nil)
		res := httptest.NewRecorder()
		createTestHandler("http://invalid").ServeHTTP(res, req)
		require.Equal(t, http.StatusBadRequest, res.Code)
		require.Contains(t, res.Body.String(), "Некорректный идентификатор записи.")
		require.NotContains(t, res.Body.String(), "Некорректный формат запроса.")
	}
}
