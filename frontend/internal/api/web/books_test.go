package web

import (
	"bytes"
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	apiClient "bookshop/frontend/internal/adapters/api"
	"bookshop/frontend/internal/domain/books/entities"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/metric/noop"
)

func TestRenderBookFormError(t *testing.T) {
	for _, tc := range []struct {
		name, response string
		status         int
	}{
		{name: "publisher reload fails", status: http.StatusInternalServerError, response: `{"error":"internal_error"}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tc.status)
				_, _ = w.Write([]byte(tc.response))
			}))
			defer api.Close()
			var output bytes.Buffer
			logger := slog.New(slog.NewJSONHandler(&output, nil))
			client, err := apiClient.New(api.URL, noop.NewMeterProvider())
			require.NoError(t, err)
			defer client.Stop()
			s := Server{Publishers: client, Logger: logger}
			ctx := attachRequest(context.Background(), testID, "POST", "POST /books")
			req := httptest.NewRequest("POST", "/books", nil).WithContext(ctx)
			res := httptest.NewRecorder()
			s.renderBookFormError(res, req, entities.Book{}, errInvalidForm)
			require.Equal(t, http.StatusBadRequest, res.Code)
			require.Contains(t, output.String(), `"event":"book_form.publishers_load_failed"`)
			require.Contains(t, output.String(), `"request_id":"`+testID+`"`)
			require.Contains(t, output.String(), "failed to reload publishers")
			require.NotContains(t, output.String(), `"event":"http.request_failed"`)
		})
	}
}
