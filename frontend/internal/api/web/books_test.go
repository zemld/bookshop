package web

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	core "bookshop/frontend/internal/adapters/httpcore"
	publisherhttp "bookshop/frontend/internal/adapters/publishers/http"
	"bookshop/frontend/internal/domain/books/entities"
	"bookshop/frontend/internal/logcontext"

	"github.com/stretchr/testify/require"
)

func TestRenderBookFormError(t *testing.T) {
	for _, tc := range []struct {
		name         string
		reloadStatus int
		wantEvent    bool
	}{
		{name: "reload fails independently of expected save failure", reloadStatus: http.StatusInternalServerError, wantEvent: true},
		{name: "reload succeeds", reloadStatus: http.StatusOK},
	} {
		t.Run(tc.name, func(t *testing.T) {
			api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, "GET /publishers", r.Method+" "+r.URL.Path)
				w.WriteHeader(tc.reloadStatus)
				if tc.reloadStatus == http.StatusOK {
					_, _ = w.Write([]byte("[]"))
				} else {
					_, _ = w.Write([]byte(`{"error":"internal_error"}`))
				}
			}))
			defer api.Close()
			var output bytes.Buffer
			server := Server{
				Publishers: &publisherhttp.Client{Core: core.New(api.URL)},
				Logger:     slog.New(logcontext.NewHandler(slog.NewJSONHandler(&output, nil))),
			}
			request := httptest.NewRequest(http.MethodPost, "/books", nil)
			request = request.WithContext(logcontext.StoreRequest(request.Context(), testID, http.MethodPost, "POST /books"))
			recorder := httptest.NewRecorder()
			server.renderBookFormError(recorder, request, entities.Book{}, &core.APIError{StatusCode: http.StatusBadRequest, Code: "invalid_input"})
			require.Equal(t, http.StatusBadRequest, recorder.Code)
			if !tc.wantEvent {
				require.Empty(t, output.String())
				return
			}
			lines := bytes.Split(bytes.TrimSpace(output.Bytes()), []byte("\n"))
			require.Len(t, lines, 1)
			var record map[string]any
			require.NoError(t, json.Unmarshal(lines[0], &record))
			require.Equal(t, "book_form.publishers_load_failed", record["event"])
			require.Equal(t, "ERROR", record["level"])
			require.Equal(t, testID, record["request_id"])
			require.Equal(t, "POST", record["method"])
			require.Equal(t, "POST /books", record["route"])
			require.Contains(t, record["error"], "API 500")
		})
	}
}
