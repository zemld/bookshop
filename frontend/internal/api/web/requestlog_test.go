package web

import (
	"bytes"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"bookshop/frontend/internal/adapters/httpcore"
	"bookshop/frontend/internal/requestlog"
)

type synchronizedBuffer struct {
	sync.Mutex
	bytes.Buffer
}

func (b *synchronizedBuffer) Write(p []byte) (int, error) {
	b.Lock()
	defer b.Unlock()
	return b.Buffer.Write(p)
}

func (b *synchronizedBuffer) readRecords(t *testing.T) []map[string]any {
	t.Helper()
	b.Lock()
	defer b.Unlock()
	var records []map[string]any
	for _, line := range bytes.Split(bytes.TrimSpace(b.Bytes()), []byte("\n")) {
		var record map[string]any
		require.NoError(t, json.Unmarshal(line, &record))
		records = append(records, record)
	}
	return records
}

func TestLogRequestIDsAndRoutes(t *testing.T) {
	var output synchronizedBuffer
	original := slog.Default()
	slog.SetDefault(slog.New(requestlog.NewHandler(slog.NewJSONHandler(&output, nil).WithAttrs([]slog.Attr{slog.String("service", "bookshop-web")}))))
	t.Cleanup(func() { slog.SetDefault(original) })
	mux := http.NewServeMux()
	mux.HandleFunc("GET /books/{id}", func(w http.ResponseWriter, r *http.Request) {
		slog.InfoContext(r.Context(), "nested", "event", "test.nested")
		w.WriteHeader(http.StatusConflict)
	})
	handler := requestlog.Wrap(mux, func(r *http.Request) string {
		_, pattern := mux.Handler(r)
		return pattern
	})
	for range 2 {
		req := httptest.NewRequest("GET", "/books/secret?token=private", nil)
		req.Header.Set("X-Request-ID", uuid.NewString())
		res := httptest.NewRecorder()
		handler.ServeHTTP(res, req)
		id, err := uuid.Parse(res.Header().Get("X-Request-ID"))
		require.NoError(t, err)
		require.Equal(t, uuid.Version(4), id.Version())
		require.NotEqual(t, req.Header.Get("X-Request-ID"), id.String())
		require.Equal(t, http.StatusConflict, res.Code)
	}
	records := output.readRecords(t)
	require.Len(t, records, 4)
	require.NotEqual(t, records[0]["request_id"], records[2]["request_id"])
	for i := 0; i < 4; i += 2 {
		require.Equal(t, records[i]["request_id"], records[i+1]["request_id"])
		require.Equal(t, "GET /books/{id}", records[i+1]["route"])
		require.Equal(t, "bookshop-web", records[i+1]["service"])
		require.Equal(t, float64(409), records[i+1]["status_code"])
		require.NotContains(t, output.String(), "private")
		require.NotContains(t, output.String(), "secret")
	}
}

func TestLogUnexpectedErrorOnly(t *testing.T) {
	var output synchronizedBuffer
	original := slog.Default()
	slog.SetDefault(slog.New(requestlog.NewHandler(slog.NewJSONHandler(&output, nil))))
	t.Cleanup(func() { slog.SetDefault(original) })
	for _, cause := range []error{errInvalidForm, errors.New("unavailable")} {
		req := httptest.NewRequest("GET", "/books", nil)
		req = req.WithContext(requestlog.StoreRequest(req.Context(), uuid.NewString(), "GET", "GET /books"))
		renderErrorPage(httptest.NewRecorder(), req, cause)
	}
	records := output.readRecords(t)
	require.Len(t, records, 1)
	require.Equal(t, "http.request_failed", records[0]["event"])
	require.Equal(t, "GET /books", records[0]["route"])
	require.Equal(t, float64(502), records[0]["status_code"])
}

func TestForwardRequestIDToAPI(t *testing.T) {
	var output synchronizedBuffer
	original := slog.Default()
	slog.SetDefault(slog.New(requestlog.NewHandler(slog.NewJSONHandler(&output, nil))))
	t.Cleanup(func() { slog.SetDefault(original) })
	received := make(chan string, 2)
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received <- r.Header.Get("X-Request-ID")
		w.WriteHeader(http.StatusNoContent)
	}))
	defer api.Close()
	client := httpcore.New(api.URL)
	mux := http.NewServeMux()
	mux.HandleFunc("GET /books", func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, client.Request(r.Context(), "GET", "/books", nil, nil))
		w.WriteHeader(http.StatusNoContent)
	})
	handler := requestlog.Wrap(mux, func(r *http.Request) string {
		_, route := mux.Handler(r)
		return route
	})
	for range 2 {
		res := httptest.NewRecorder()
		handler.ServeHTTP(res, httptest.NewRequest("GET", "/books", nil))
		require.Equal(t, res.Header().Get("X-Request-ID"), <-received)
	}
}

func TestOmitSuccessfulAssetsAndLogMissingAssets(t *testing.T) {
	var output synchronizedBuffer
	original := slog.Default()
	slog.SetDefault(slog.New(requestlog.NewHandler(slog.NewJSONHandler(&output, nil))))
	t.Cleanup(func() { slog.SetDefault(original) })
	handler := (Server{}).CreateHandler()
	for _, tc := range []struct {
		path   string
		status int
	}{
		{"/assets/site.css", 200},
		{"/assets/missing-private.css", 404},
	} {
		res := httptest.NewRecorder()
		handler.ServeHTTP(res, httptest.NewRequest("GET", tc.path, nil))
		require.Equal(t, tc.status, res.Code)
		require.NotEmpty(t, res.Header().Get("X-Request-ID"))
	}
	records := output.readRecords(t)
	require.Len(t, records, 1)
	require.Equal(t, "GET /assets/", records[0]["route"])
	require.Equal(t, float64(404), records[0]["status_code"])
	require.NotContains(t, output.String(), "missing-private")
}

func TestRejectMalformedFormAndQueryWithoutLoggingInput(t *testing.T) {
	var output synchronizedBuffer
	original := slog.Default()
	slog.SetDefault(slog.New(requestlog.NewHandler(slog.NewJSONHandler(&output, nil))))
	t.Cleanup(func() { slog.SetDefault(original) })
	server := httptest.NewServer((Server{}).CreateHandler())
	defer server.Close()

	for _, route := range []string{"/books", "/publishers"} {
		for _, input := range []struct {
			path string
			body string
		}{
			{route, "name=%PRIVATE_FORM_MARKER"},
			{route + "?private=%PRIVATE_QUERY_MARKER", "name=Safe"},
		} {
			req, err := http.NewRequest(http.MethodPost, server.URL+input.path, strings.NewReader(input.body))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			res, err := server.Client().Do(req)
			require.NoError(t, err)
			require.Equal(t, http.StatusBadRequest, res.StatusCode)
			require.NotEmpty(t, res.Header.Get("X-Request-ID"))
			require.NoError(t, res.Body.Close())
		}
	}

	records := output.readRecords(t)
	require.Len(t, records, 4)
	for _, record := range records {
		require.Equal(t, "http.request_completed", record["event"])
		require.Equal(t, float64(http.StatusBadRequest), record["status_code"])
	}
	require.NotContains(t, output.String(), "PRIVATE")
	require.NotContains(t, output.String(), "%PR")
	require.NotContains(t, output.String(), "http.request_failed")
}
