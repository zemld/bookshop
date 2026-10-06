package web

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"bookshop/frontend/internal/adapters/httpcore"

	"github.com/stretchr/testify/require"
)

func TestBookFormReturnsFirstSyntaxError(t *testing.T) {
	t.Parallel()
	values := url.Values{
		"author": {"Автор"}, "name": {"Название"}, "year": {"bad"},
		"publicationYear": {""}, "price": {"1.5"}, "quantity": {"9223372036854775808"},
		"publisherId": {"bad"},
	}
	for _, tc := range []struct{ code, field, value string }{
		{"invalid_year", "year", "2024"},
		{"invalid_price", "price", "100"},
		{"invalid_publisher_id", "publisherId", testID},
		{"invalid_publication_year", "publicationYear", "2024"},
		{"invalid_quantity", "quantity", "1"},
	} {
		book, err := parseBookForm(values)
		require.ErrorIs(t, err, errInvalidForm)
		var formError *formValidationError
		require.ErrorAs(t, err, &formError)
		require.Equal(t, tc.code, formError.Code)
		status, message, fields := classifyPublicError(err)
		require.Equal(t, 400, status)
		require.Equal(t, "Проверьте заполнение полей.", message)
		require.Len(t, fields, 1)
		require.NotEmpty(t, fields[tc.field])
		require.Equal(t, "Автор", book.Author)
		require.Equal(t, "Название", book.Name)
		values.Set(tc.field, tc.value)
	}
	_, err := parseBookForm(values)
	require.NoError(t, err)
}

func TestBookValidationKeepsFormAndValues(t *testing.T) {
	t.Parallel()
	for _, path := range []string{"/books", "/books/" + testID} {
		for _, hx := range []string{"", "true"} {
			t.Run(path+"/htmx="+hx, func(t *testing.T) {
				t.Parallel()
				fields := map[string]string{
					"price": "Цена не может быть отрицательной.",
				}
				api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if r.Method == "GET" && r.URL.Path == "/publishers" {
						_, _ = fmt.Fprintf(w, `[{"id":%q,"name":"Издательство"}]`, testID)
						return
					}
					method := "POST"
					if path != "/books" {
						method = "PUT"
					}
					require.Equal(t, method+" "+path, r.Method+" "+r.URL.Path)
					w.WriteHeader(400)
					require.NoError(t, json.NewEncoder(w).Encode(map[string]any{
						"error": "book_price_negative",
					}))
				}))
				t.Cleanup(api.Close)
				values := url.Values{
					"author": {" Автор "}, "name": {`<script>alert("bad")</script>`},
					"year": {"2025"}, "publicationYear": {"2024"},
					"price": {"-1"}, "quantity": {"-2"}, "publisherId": {testID},
				}
				req := httptest.NewRequest("POST", path, strings.NewReader(values.Encode()))
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
				req.Header.Set("HX-Request", hx)
				res := httptest.NewRecorder()
				createTestHandler(api.URL).ServeHTTP(res, req)

				require.Equal(t, 400, res.Code)
				require.Empty(t, res.Header().Get("Location"))
				require.Empty(t, res.Header().Get("HX-Location"))
				require.NotContains(t, res.Body.String(), "API 400:")
				require.NotContains(t, res.Body.String(), "book_price_negative")
				require.NotContains(t, res.Body.String(), "book_publication_year_before_year")
				doc := parseHTML(t, res.Body.String())
				require.Equal(t, path, getAttribute(findElements(doc, "form")[0], "action"))
				for _, input := range findElements(doc, "input") {
					name := getAttribute(input, "name")
					require.Equal(t, values.Get(name), getAttribute(input, "value"))
					if fields[name] != "" {
						require.Equal(t, "true", getAttribute(input, "aria-invalid"))
						require.Equal(t, name+"-error", getAttribute(input, "aria-describedby"))
					} else {
						require.Equal(t, "false", getAttribute(input, "aria-invalid"))
					}
				}
				for field, message := range fields {
					require.Contains(t, res.Body.String(), `id="`+field+`-error"`)
					require.Contains(t, res.Body.String(), message)
				}
				require.Len(t, findElements(doc, "script"), 2)
				require.Empty(t, findElements(doc, "script")[0].FirstChild)
				require.Contains(t, res.Body.String(), `value="`+testID+`" selected`)
				if path != "/books" {
					require.Contains(t, res.Body.String(), path+"/delete")
				}
			})
		}
	}
}

func TestPublisherValidationAndConflictKeepForm(t *testing.T) {
	t.Parallel()
	for _, path := range []string{"/publishers", "/publishers/" + testID} {
		for _, status := range []int{400, 409} {
			for _, hx := range []string{"", "true"} {
				t.Run(fmt.Sprintf("%s/%d/htmx=%s", path, status, hx), func(t *testing.T) {
					t.Parallel()
					message := "Укажите название издательства."
					code := "publisher_name_required"
					if status == 409 {
						message = "Издательство с таким названием уже существует."
						code = "publisher_name_duplicate"
					}
					api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
						w.WriteHeader(status)
						require.NoError(t, json.NewEncoder(w).Encode(map[string]any{
							"error": code,
						}))
					}))
					t.Cleanup(api.Close)
					value := `  Имя "издательства"  `
					req := httptest.NewRequest("POST", path, strings.NewReader(url.Values{"name": {value}}.Encode()))
					req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
					req.Header.Set("HX-Request", hx)
					res := httptest.NewRecorder()
					createTestHandler(api.URL).ServeHTTP(res, req)

					require.Equal(t, status, res.Code)
					doc := parseHTML(t, res.Body.String())
					require.Equal(t, path, getAttribute(findElements(doc, "form")[0], "action"))
					require.Equal(t, value, getAttribute(findElements(doc, "input")[0], "value"))
					require.Equal(t, "true", getAttribute(findElements(doc, "input")[0], "aria-invalid"))
					require.Contains(t, res.Body.String(), message)
					require.NotContains(t, res.Body.String(), "API ")
					require.NotContains(t, res.Body.String(), code)
				})
			}
		}
	}
}

func TestPublicErrorHidesInternalDetails(t *testing.T) {
	t.Parallel()
	for _, err := range []error{
		errors.New("dial tcp private-host:5432 password=secret"),
		fmt.Errorf("wrapped: %w", &httpcore.APIError{StatusCode: 500, Code: "private SQL"}),
		&httpcore.APIError{StatusCode: 400, Code: "invalid_response"},
	} {
		code, message, fields := classifyPublicError(err)
		require.Equal(t, 502, code)
		require.Equal(t, "Не удалось выполнить запрос к серверу. Попробуйте позже.", message)
		require.Empty(t, fields)
	}
}
