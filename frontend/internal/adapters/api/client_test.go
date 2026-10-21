package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	bookentities "bookshop/frontend/internal/domain/books/entities"
	publisherentities "bookshop/frontend/internal/domain/publishers/entities"
	"bookshop/frontend/internal/telemetry"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/metric/noop"
)

func TestExportOgenMetrics(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Request-ID", _testRequestID)
		_, _ = w.Write([]byte("[]"))
	}))
	t.Cleanup(backend.Close)
	metrics, err := telemetry.New()
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, metrics.Provider.Shutdown(context.Background())) })
	client, err := New(backend.URL, metrics.Provider)
	require.NoError(t, err)
	t.Cleanup(client.Stop)
	_, err = client.ListBooks(context.Background())
	require.NoError(t, err)
	export := httptest.NewRecorder()
	metrics.Handler.ServeHTTP(export, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	require.Contains(t, export.Body.String(), "ogen_client_request_count_total")
}

const (
	_testID        = "4e643615-2a28-40a5-8e7d-fbe6dd2d498d"
	_testRequestID = "51aa5ea3-0653-49ad-a3b5-e4c3b7ade70c"
	_testBook      = `{"id":"4e643615-2a28-40a5-8e7d-fbe6dd2d498d","author":"Author","name":"Name","year":2000,"price":1,"publisherId":"4e643615-2a28-40a5-8e7d-fbe6dd2d498d","publicationYear":2001,"quantity":2}`
	_testPublisher = `{"id":"4e643615-2a28-40a5-8e7d-fbe6dd2d498d","name":"Name"}`
)

func TestListBooks(t *testing.T) {
	for _, tc := range []struct {
		name, body, responseID, wantError string
		status                            int
	}{
		{"success", "[" + _testBook + "]", _testRequestID, "", http.StatusOK},
		{"domain error", `{"error":"invalid_input"}`, _testRequestID, "API 400: invalid_input", http.StatusBadRequest},
		{"malformed problem", `{"secret":"private proxy failure"}`, _testRequestID, "API 502: invalid_response", http.StatusBadGateway},
		{"missing success header", `[]`, "", "API 502: invalid_response", http.StatusOK},
		{"missing error header", `{"error":"invalid_input"}`, "", "API 502: invalid_response", http.StatusBadRequest},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodGet, r.Method)
				require.Equal(t, "/books", r.URL.Path)
				require.Equal(t, _testRequestID, r.Header.Get("X-Request-ID"))
				w.Header().Set("Content-Type", "application/json")
				if tc.responseID != "" {
					w.Header().Set("X-Request-ID", tc.responseID)
				}
				w.WriteHeader(tc.status)
				_, _ = w.Write([]byte(tc.body))
			}))
			t.Cleanup(server.Close)
			client, err := New(server.URL, noop.NewMeterProvider())
			require.NoError(t, err)
			t.Cleanup(client.Stop)
			books, err := client.ListBooks(StoreRequestID(context.Background(), _testRequestID))
			if tc.wantError != "" {
				require.EqualError(t, err, tc.wantError)
				require.NotContains(t, err.Error(), "private proxy failure")
				return
			}
			require.NoError(t, err)
			require.Len(t, books, 1)
			require.Equal(t, uuid.MustParse(_testID), books[0].ID)
			require.Equal(t, "Author", books[0].Author)
			require.Equal(t, "Name", books[0].Name)
			require.Equal(t, bookentities.Year(2000), books[0].Year)
			require.Equal(t, int64(1), books[0].Price)
			require.Equal(t, uuid.MustParse(_testID), books[0].PublisherID)
			require.Equal(t, bookentities.Year(2001), books[0].PublicationYear)
			require.Equal(t, int64(2), books[0].Quantity)
		})
	}
}

func TestGetBook(t *testing.T) {
	for _, tc := range []struct{ name, body string }{{"success", _testBook}} {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodGet, r.Method)
				require.Equal(t, "/books/"+_testID, r.URL.Path)
				require.Equal(t, _testRequestID, r.Header.Get("X-Request-ID"))
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("X-Request-ID", _testRequestID)
				_, _ = w.Write([]byte(tc.body))
			}))
			t.Cleanup(server.Close)
			client, err := New(server.URL, noop.NewMeterProvider())
			require.NoError(t, err)
			t.Cleanup(client.Stop)
			book, err := client.GetBook(StoreRequestID(context.Background(), _testRequestID), uuid.MustParse(_testID))
			require.NoError(t, err)
			require.Equal(t, uuid.MustParse(_testID), book.ID)
			require.Equal(t, "Author", book.Author)
		})
	}
}

func TestCreateBook(t *testing.T) {
	for _, tc := range []struct {
		name, body, wantError string
		status                int
	}{
		{"success", _testBook, "", http.StatusOK},
		{"domain error", `{"error":"invalid_input"}`, "API 400: invalid_input", http.StatusBadRequest},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodPost, r.Method)
				require.Equal(t, "/books", r.URL.Path)
				require.Equal(t, _testRequestID, r.Header.Get("X-Request-ID"))
				var input map[string]any
				require.NoError(t, json.NewDecoder(r.Body).Decode(&input))
				require.Equal(t, "Name", input["name"])
				require.Equal(t, _testID, input["publisherId"])
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("X-Request-ID", _testRequestID)
				w.WriteHeader(tc.status)
				_, _ = w.Write([]byte(tc.body))
			}))
			t.Cleanup(server.Close)
			client, err := New(server.URL, noop.NewMeterProvider())
			require.NoError(t, err)
			t.Cleanup(client.Stop)
			book, err := client.CreateBook(StoreRequestID(context.Background(), _testRequestID), bookentities.Book{Name: "Name", PublisherID: uuid.MustParse(_testID)})
			if tc.wantError != "" {
				require.EqualError(t, err, tc.wantError)
				return
			}
			require.NoError(t, err)
			require.Equal(t, uuid.MustParse(_testID), book.ID)
			require.Equal(t, "Author", book.Author)
		})
	}
}

func TestUpdateBook(t *testing.T) {
	for _, tc := range []struct{ name, body string }{{"success", _testBook}} {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodPut, r.Method)
				require.Equal(t, "/books/"+_testID, r.URL.Path)
				require.Equal(t, _testRequestID, r.Header.Get("X-Request-ID"))
				var input map[string]any
				require.NoError(t, json.NewDecoder(r.Body).Decode(&input))
				require.Equal(t, "Name", input["name"])
				require.Equal(t, _testID, input["publisherId"])
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("X-Request-ID", _testRequestID)
				_, _ = w.Write([]byte(tc.body))
			}))
			t.Cleanup(server.Close)
			client, err := New(server.URL, noop.NewMeterProvider())
			require.NoError(t, err)
			t.Cleanup(client.Stop)
			book, err := client.UpdateBook(StoreRequestID(context.Background(), _testRequestID), uuid.MustParse(_testID), bookentities.Book{Name: "Name", PublisherID: uuid.MustParse(_testID)})
			require.NoError(t, err)
			require.Equal(t, uuid.MustParse(_testID), book.ID)
			require.Equal(t, "Author", book.Author)
		})
	}
}

func TestDeleteBook(t *testing.T) {
	for _, tc := range []struct{ name, body string }{{"success", `{"status":"deleted"}`}} {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodDelete, r.Method)
				require.Equal(t, "/books/"+_testID, r.URL.Path)
				require.Equal(t, _testRequestID, r.Header.Get("X-Request-ID"))
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("X-Request-ID", _testRequestID)
				_, _ = w.Write([]byte(tc.body))
			}))
			t.Cleanup(server.Close)
			client, err := New(server.URL, noop.NewMeterProvider())
			require.NoError(t, err)
			t.Cleanup(client.Stop)
			require.NoError(t, client.DeleteBook(StoreRequestID(context.Background(), _testRequestID), uuid.MustParse(_testID)))
		})
	}
}

func TestListPublishers(t *testing.T) {
	for _, tc := range []struct{ name, body string }{{"success", "[" + _testPublisher + "]"}} {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodGet, r.Method)
				require.Equal(t, "/publishers", r.URL.Path)
				require.Equal(t, _testRequestID, r.Header.Get("X-Request-ID"))
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("X-Request-ID", _testRequestID)
				_, _ = w.Write([]byte(tc.body))
			}))
			t.Cleanup(server.Close)
			client, err := New(server.URL, noop.NewMeterProvider())
			require.NoError(t, err)
			t.Cleanup(client.Stop)
			publishers, err := client.ListPublishers(StoreRequestID(context.Background(), _testRequestID))
			require.NoError(t, err)
			require.Len(t, publishers, 1)
			require.Equal(t, uuid.MustParse(_testID), publishers[0].ID)
			require.Equal(t, "Name", publishers[0].Name)
		})
	}
}

func TestGetPublisher(t *testing.T) {
	for _, tc := range []struct{ name, body string }{{"success", _testPublisher}} {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodGet, r.Method)
				require.Equal(t, "/publishers/"+_testID, r.URL.Path)
				require.Equal(t, _testRequestID, r.Header.Get("X-Request-ID"))
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("X-Request-ID", _testRequestID)
				_, _ = w.Write([]byte(tc.body))
			}))
			t.Cleanup(server.Close)
			client, err := New(server.URL, noop.NewMeterProvider())
			require.NoError(t, err)
			t.Cleanup(client.Stop)
			publisher, err := client.GetPublisher(StoreRequestID(context.Background(), _testRequestID), uuid.MustParse(_testID))
			require.NoError(t, err)
			require.Equal(t, uuid.MustParse(_testID), publisher.ID)
			require.Equal(t, "Name", publisher.Name)
		})
	}
}

func TestCreatePublisher(t *testing.T) {
	for _, tc := range []struct {
		name, body, wantError string
		status                int
	}{
		{"success", _testPublisher, "", http.StatusOK},
		{"domain error", `{"error":"invalid_input"}`, "API 400: invalid_input", http.StatusBadRequest},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodPost, r.Method)
				require.Equal(t, "/publishers", r.URL.Path)
				require.Equal(t, _testRequestID, r.Header.Get("X-Request-ID"))
				var input map[string]any
				require.NoError(t, json.NewDecoder(r.Body).Decode(&input))
				require.Equal(t, "Name", input["name"])
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("X-Request-ID", _testRequestID)
				w.WriteHeader(tc.status)
				_, _ = w.Write([]byte(tc.body))
			}))
			t.Cleanup(server.Close)
			client, err := New(server.URL, noop.NewMeterProvider())
			require.NoError(t, err)
			t.Cleanup(client.Stop)
			publisher, err := client.CreatePublisher(StoreRequestID(context.Background(), _testRequestID), publisherentities.Publisher{Name: "Name"})
			if tc.wantError != "" {
				require.EqualError(t, err, tc.wantError)
				return
			}
			require.NoError(t, err)
			require.Equal(t, uuid.MustParse(_testID), publisher.ID)
			require.Equal(t, "Name", publisher.Name)
		})
	}
}

func TestUpdatePublisher(t *testing.T) {
	for _, tc := range []struct{ name, body string }{{"success", _testPublisher}} {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodPut, r.Method)
				require.Equal(t, "/publishers/"+_testID, r.URL.Path)
				require.Equal(t, _testRequestID, r.Header.Get("X-Request-ID"))
				var input map[string]any
				require.NoError(t, json.NewDecoder(r.Body).Decode(&input))
				require.Equal(t, "Name", input["name"])
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("X-Request-ID", _testRequestID)
				_, _ = w.Write([]byte(tc.body))
			}))
			t.Cleanup(server.Close)
			client, err := New(server.URL, noop.NewMeterProvider())
			require.NoError(t, err)
			t.Cleanup(client.Stop)
			publisher, err := client.UpdatePublisher(StoreRequestID(context.Background(), _testRequestID), uuid.MustParse(_testID), publisherentities.Publisher{Name: "Name"})
			require.NoError(t, err)
			require.Equal(t, uuid.MustParse(_testID), publisher.ID)
			require.Equal(t, "Name", publisher.Name)
		})
	}
}

func TestDeletePublisher(t *testing.T) {
	for _, tc := range []struct{ name, body string }{{"success", `{"status":"deleted"}`}} {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodDelete, r.Method)
				require.Equal(t, "/publishers/"+_testID, r.URL.Path)
				require.Equal(t, _testRequestID, r.Header.Get("X-Request-ID"))
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("X-Request-ID", _testRequestID)
				_, _ = w.Write([]byte(tc.body))
			}))
			t.Cleanup(server.Close)
			client, err := New(server.URL, noop.NewMeterProvider())
			require.NoError(t, err)
			t.Cleanup(client.Stop)
			require.NoError(t, client.DeletePublisher(StoreRequestID(context.Background(), _testRequestID), uuid.MustParse(_testID)))
		})
	}
}
