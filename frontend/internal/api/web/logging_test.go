package web

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"bookshop/frontend/internal/requestlog"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestLogRequests(t *testing.T) {
	var output bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&output, nil)).With("service", "bookshop-web")
	mux := http.NewServeMux()
	mux.HandleFunc("GET /books/{id}", func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, r.Header.Get("X-Request-ID"), "supplied")
		require.NoError(t, http.NewResponseController(w).Flush())
	})
	mux.HandleFunc("GET /assets/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/assets/missing" {
			w.WriteHeader(http.StatusNotFound)
		}
	})
	mux.HandleFunc("GET /flush", func(w http.ResponseWriter, _ *http.Request) {
		require.NoError(t, http.NewResponseController(w).Flush())
		w.WriteHeader(http.StatusInternalServerError)
	})
	handler := logRequests(mux, logger)
	for _, tc := range []struct {
		name       string
		path       string
		wantRoute  string
		wantStatus int
		wantLog    bool
		run        func(*testing.T)
	}{
		{name: "matched route", path: "/books/secret?token=private", wantRoute: "GET /books/{id}", wantStatus: 200, wantLog: true},
		{name: "unmatched route", path: "/not-found?token=private", wantRoute: "<unmatched>", wantStatus: 404, wantLog: true},
		{name: "successful asset", path: "/assets/ok", wantRoute: "GET /assets/", wantStatus: 200},
		{name: "missing asset", path: "/assets/missing", wantRoute: "GET /assets/", wantStatus: 404, wantLog: true},
		{name: "flush commits success", path: "/flush", wantRoute: "GET /flush", wantStatus: 200, wantLog: true},
		{name: "concurrent IDs", run: func(t *testing.T) {
			concurrentMux := http.NewServeMux()
			concurrentMux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, w.Header().Get("X-Request-ID"), requestlog.GetDetails(r.Context()).ID)
			})
			concurrentHandler := logRequests(concurrentMux, slog.New(slog.NewJSONHandler(io.Discard, nil)))
			var wg sync.WaitGroup
			ids := make(chan string, 20)
			for range 20 {
				wg.Add(1)
				go func() {
					defer wg.Done()
					res := httptest.NewRecorder()
					concurrentHandler.ServeHTTP(res, httptest.NewRequest("GET", "/", nil))
					ids <- res.Header().Get("X-Request-ID")
				}()
			}
			wg.Wait()
			close(ids)
			seen := make(map[string]bool)
			for id := range ids {
				require.False(t, seen[id])
				seen[id] = true
			}
			require.Len(t, seen, 20)
		}},
		{name: "propagates to backend", run: func(t *testing.T) {
			ids := make(chan string, 2)
			api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				ids <- r.Header.Get("X-Request-ID")
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`[]`))
			}))
			defer api.Close()
			for range 2 {
				res := httptest.NewRecorder()
				req := httptest.NewRequest("GET", "/books", nil)
				req.Header.Set("X-Request-ID", "external")
				createTestHandler(api.URL).ServeHTTP(res, req)
				require.Equal(t, http.StatusOK, res.Code)
				require.Equal(t, res.Header().Get("X-Request-ID"), <-ids)
				require.NotEqual(t, "external", res.Header().Get("X-Request-ID"))
			}
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if tc.run != nil {
				tc.run(t)
				return
			}
			output.Reset()
			req := httptest.NewRequest("GET", tc.path, nil)
			req.Header.Set("X-Request-ID", "supplied")
			res := httptest.NewRecorder()
			handler.ServeHTTP(res, req)
			id, err := uuid.Parse(res.Header().Get("X-Request-ID"))
			require.NoError(t, err)
			require.Equal(t, uuid.Version(4), id.Version())
			require.Equal(t, tc.wantStatus, res.Code)
			if !tc.wantLog {
				require.Empty(t, output.String())
				return
			}
			var record map[string]any
			require.NoError(t, json.Unmarshal(output.Bytes(), &record))
			require.Equal(t, "http.request_completed", record["event"])
			require.Equal(t, "bookshop-web", record["service"])
			require.Equal(t, id.String(), record["request_id"])
			require.Equal(t, tc.wantRoute, record["route"])
			require.Equal(t, float64(tc.wantStatus), record["status_code"])
			require.NotContains(t, output.String(), "private")
			require.NotContains(t, output.String(), "secret")
		})
	}
}

func TestWrapStatusWriter(t *testing.T) {
	recorder := httptest.NewRecorder()
	status := &statusWriter{ResponseWriter: recorder}
	wrapped := wrapStatusWriter(status)
	require.NoError(t, http.NewResponseController(wrapped).Flush())
	wrapped.WriteHeader(http.StatusInternalServerError)
	require.Equal(t, http.StatusOK, recorder.Code)
	require.Equal(t, http.StatusOK, status.status)
}
