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

	"bookshop/frontend/internal/adapters/api"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestNewRequestLoggingMiddleware(t *testing.T) {
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
	handler := NewRequestLoggingMiddleware(mux, logger)
	assetHandler := (Server{Logger: logger}).CreateHandler()
	for _, tc := range []struct {
		name       string
		path       string
		header     string
		value      string
		asset      bool
		wantRoute  string
		wantStatus int
		wantLog    bool
		run        func(*testing.T)
	}{
		{name: "matched route", path: "/books/secret?token=private", wantRoute: "GET /books/{id}", wantStatus: 200, wantLog: true},
		{name: "unmatched route", path: "/not-found?token=private", wantRoute: "<unmatched>", wantStatus: 404, wantLog: true},
		{name: "successful asset", path: "/assets/ok", wantRoute: "GET /assets/", wantStatus: 200},
		{name: "missing asset", path: "/assets/missing", wantRoute: "GET /assets/", wantStatus: 404, wantLog: true},
		{name: "FileServer existing asset", path: "/assets/site.css", asset: true, wantRoute: "GET /assets/", wantStatus: 200},
		{name: "FileServer conditional asset", path: "/assets/site.css", header: "If-None-Match", value: "*", asset: true, wantRoute: "GET /assets/", wantStatus: 304},
		{name: "FileServer partial asset", path: "/assets/site.css", header: "Range", value: "bytes=0-3", asset: true, wantRoute: "GET /assets/", wantStatus: 206},
		{name: "FileServer missing asset", path: "/assets/missing.css", asset: true, wantRoute: "GET /assets/", wantStatus: 404, wantLog: true},
		{name: "FileServer invalid range", path: "/assets/site.css", header: "Range", value: "bytes=999999999-", asset: true, wantRoute: "GET /assets/", wantStatus: 416, wantLog: true},
		{name: "flush commits success", path: "/flush", wantRoute: "GET /flush", wantStatus: 200, wantLog: true},
		{name: "concurrent IDs", run: func(t *testing.T) {
			concurrentMux := http.NewServeMux()
			concurrentMux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, w.Header().Get("X-Request-ID"), api.GetRequestID(r.Context()))
			})
			concurrentHandler := NewRequestLoggingMiddleware(concurrentMux, slog.New(slog.NewJSONHandler(io.Discard, nil)))
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
				w.Header().Set("X-Request-ID", r.Header.Get("X-Request-ID"))
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
			if tc.header != "" {
				req.Header.Set(tc.header, tc.value)
			}
			res := httptest.NewRecorder()
			if tc.asset {
				assetHandler.ServeHTTP(res, req)
			} else {
				handler.ServeHTTP(res, req)
			}
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

func TestResponseRecorder(t *testing.T) {
	for _, tc := range []struct {
		name string
		run  func(http.ResponseWriter)
		want int
	}{
		{"empty response", func(http.ResponseWriter) {}, http.StatusOK},
		{"implicit write", func(w http.ResponseWriter) { _, _ = w.Write([]byte("ok")) }, http.StatusOK},
		{"informational then error", func(w http.ResponseWriter) {
			w.WriteHeader(http.StatusEarlyHints)
			w.WriteHeader(http.StatusBadRequest)
		}, http.StatusBadRequest},
		{"explicit error", func(w http.ResponseWriter) { w.WriteHeader(http.StatusInternalServerError) }, http.StatusInternalServerError},
		{"flush then error", func(w http.ResponseWriter) {
			require.NoError(t, http.NewResponseController(w).Flush())
			w.WriteHeader(http.StatusInternalServerError)
		}, http.StatusOK},
	} {
		t.Run(tc.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			status := &responseRecorder{ResponseWriter: recorder}
			tc.run(status)
			require.Equal(t, tc.want, status.getStatusCode())
			if tc.name != "empty response" && tc.name != "informational then error" {
				require.Equal(t, tc.want, recorder.Code)
			}
		})
	}
}
