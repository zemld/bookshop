package web

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	bookentities "bookshop/frontend/internal/domain/books/entities"
	publisherentities "bookshop/frontend/internal/domain/publishers/entities"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/html"
)

func parseHTML(t *testing.T, body string) *html.Node {
	t.Helper()
	node, err := html.Parse(strings.NewReader(body))
	require.NoError(t, err)
	return node
}

func findElements(root *html.Node, tag string) []*html.Node {
	var found []*html.Node
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == tag {
			found = append(found, n)
		}
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(root)
	return found
}

func getAttribute(n *html.Node, key string) string {
	for _, a := range n.Attr {
		if a.Key == key {
			return a.Val
		}
	}
	return ""
}

func TestRenderedBookForms(t *testing.T) {
	t.Parallel()
	id := uuid.MustParse(testID)
	publisher := publisherentities.Publisher{ID: id, Name: `Дом "Книги"`}
	for _, tt := range []struct {
		name, action, value string
		newBook             bool
	}{
		{"new", "/books", "", true},
		{"existing", "/books/" + testID, "2020", false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			selectedPublisher := uuid.Nil
			if !tt.newBook {
				selectedPublisher = id
			}
			res := httptest.NewRecorder()
			renderPage(res, 200, page{Kind: bookPage, Title: "Книга", New: tt.newBook,
				Book: bookentities.Book{ID: id, Name: `История "1"`, Year: 2020, PublisherID: selectedPublisher}, Publishers: []publisherentities.Publisher{publisher}})

			require.Equal(t, 200, res.Code)
			doc := parseHTML(t, res.Body.String())
			require.Equal(t, "ru", getAttribute(findElements(doc, "html")[0], "lang"))
			require.Equal(t, tt.action, getAttribute(findElements(doc, "form")[0], "action"))
			require.Equal(t, "false", getAttribute(findElements(doc, "form")[0], "hx-push-url"))
			fields := map[string]string{}
			for _, input := range findElements(doc, "input") {
				fields[getAttribute(input, "name")] = getAttribute(input, "value")
			}
			for _, key := range []string{"author", "name", "year", "price", "publicationYear", "quantity"} {
				require.Contains(t, fields, key)
			}
			require.Equal(t, `История "1"`, fields["name"])
			require.Equal(t, tt.value, fields["year"])
			require.Equal(t, "publisherId", getAttribute(findElements(doc, "select")[0], "name"))
			selected := false
			for _, a := range findElements(doc, "option")[1].Attr {
				if a.Key == "selected" {
					selected = true
				}
			}
			require.Equal(t, !tt.newBook, selected)
			require.Equal(t, `Дом "Книги"`, findElements(doc, "option")[1].FirstChild.Data)
		})
	}
}

func TestRenderedEscapingAndNavigation(t *testing.T) {
	t.Parallel()
	res := httptest.NewRecorder()
	renderPage(res, 200, page{Kind: booksPage, Title: "Книги", Books: []bookentities.Book{
		{ID: uuid.MustParse(testID), Name: `<script>alert("bad")</script>`, Author: `Иван & сын`},
	}})

	require.Equal(t, 200, res.Code)
	require.NotContains(t, res.Body.String(), `<script>alert("bad")</script>`)
	require.NotContains(t, res.Body.String(), "<script>alert")
	doc := parseHTML(t, res.Body.String())
	require.Equal(t, "page", getAttribute(findElements(doc, "div")[0], "id"))
	require.Equal(t, "true", getAttribute(findElements(doc, "div")[0], "hx-boost"))
	require.Equal(t, "#page", getAttribute(findElements(doc, "div")[0], "hx-select"))
	require.Equal(t, "#page", getAttribute(findElements(doc, "div")[0], "hx-target"))
	require.Equal(t, "outerHTML", getAttribute(findElements(doc, "div")[0], "hx-swap"))
	var links []string
	for _, a := range findElements(doc, "a") {
		links = append(links, getAttribute(a, "href"))
	}
	require.Contains(t, links, "/books/"+testID)
}

func TestRenderedEmptyStatesAndErrors(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name string
		page page
		want string
	}{
		{"books", page{Kind: booksPage, Title: "Книги"}, "Книг пока нет."},
		{"publishers", page{Kind: publishersPage, Title: "Издательства"}, "Издательств пока нет."},
		{"error", page{Kind: errorPage, Title: "Ошибка", Error: `<img src=x onerror=alert(1)>`}, `<img src=x onerror=alert(1)>`},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			res := httptest.NewRecorder()
			renderPage(res, http.StatusOK, tt.page)

			require.Equal(t, http.StatusOK, res.Code)
			doc := parseHTML(t, res.Body.String())
			require.NotEmpty(t, findElements(doc, "main"))
			if tt.page.Kind == errorPage {
				require.NotContains(t, res.Body.String(), tt.want)
				require.Empty(t, findElements(doc, "img"))
				require.Equal(t, "alert", getAttribute(findElements(doc, "p")[0], "role"))
			} else {
				require.Contains(t, res.Body.String(), tt.want)
				require.NotEmpty(t, findElements(doc, "a"))
			}
		})
	}
}

func TestAssetsServedLocally(t *testing.T) {
	t.Parallel()
	routes := createTestHandler("http://invalid")
	for _, path := range []string{"/assets/site.css", "/assets/site.js", "/assets/htmx.min.js"} {
		t.Run(path, func(t *testing.T) {
			t.Parallel()
			res := httptest.NewRecorder()
			routes.ServeHTTP(res, httptest.NewRequest(http.MethodGet, path, nil))
			require.Equal(t, http.StatusOK, res.Code)
			require.NotEmpty(t, res.Body.String())
			require.NotContains(t, res.Body.String(), "<html")
		})
	}
}

func TestPublisherSaveHTMXAndFallback(t *testing.T) {
	t.Parallel()
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "POST", r.Method)
		require.Equal(t, "/publishers", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"` + testID + `","name":"Книга"}`))
	}))
	t.Cleanup(api.Close)

	for _, tt := range []struct {
		name, hx string
		code     int
	}{
		{"browser", "", http.StatusSeeOther},
		{"htmx", "true", http.StatusOK},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			req := httptest.NewRequest("POST", "/publishers", strings.NewReader(url.Values{"name": {"Книга"}}.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			req.Header.Set("HX-Request", tt.hx)
			res := httptest.NewRecorder()
			createTestHandler(api.URL).ServeHTTP(res, req)

			require.Equal(t, tt.code, res.Code)
			if tt.hx == "" {
				require.Equal(t, "/publishers/"+testID, res.Header().Get("Location"))
				require.Empty(t, res.Header().Get("HX-Location"))
			} else {
				var location struct{ Path, Target, Select, Swap string }
				require.NoError(t, json.Unmarshal([]byte(res.Header().Get("HX-Location")), &location))
				require.Equal(t, "/publishers/"+testID, location.Path)
				require.Equal(t, "#page", location.Target)
				require.Equal(t, "#page", location.Select)
				require.Equal(t, "outerHTML", location.Swap)
			}
		})
	}
}

func TestHTMXBadRequestRendersForm(t *testing.T) {
	t.Parallel()
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "GET /publishers", r.Method+" "+r.URL.Path)
		_, _ = w.Write([]byte(`[]`))
	}))
	t.Cleanup(api.Close)
	req := httptest.NewRequest("POST", "/books", strings.NewReader("year=bad"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("HX-Request", "true")
	res := httptest.NewRecorder()
	createTestHandler(api.URL).ServeHTTP(res, req)

	require.Equal(t, http.StatusBadRequest, res.Code)
	doc := parseHTML(t, res.Body.String())
	require.Equal(t, "page", getAttribute(findElements(doc, "div")[0], "id"))
	require.Equal(t, "alert", getAttribute(findElements(doc, "p")[0], "role"))
	require.Equal(t, "/books", getAttribute(findElements(doc, "form")[0], "action"))
	require.Contains(t, res.Body.String(), `id="year-error"`)
	require.Empty(t, res.Header().Get("HX-Location"))
}
