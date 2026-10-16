package web

import (
	"bytes"
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"bookshop/frontend/internal/adapters/httpcore"
	publisherhttp "bookshop/frontend/internal/adapters/publishers/http"
	"bookshop/frontend/internal/domain/books/entities"

	"github.com/stretchr/testify/require"
)

func TestRenderBookFormError(t *testing.T) {
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":"internal_error"}`))
	}))
	defer api.Close()
	var output bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&output, nil))
	s := Server{Publishers: &publisherhttp.Client{Core: httpcore.New(api.URL)}, Logger: logger}
	ctx := withRequest(context.Background(), testID, "POST", "POST /books")
	req := httptest.NewRequest("POST", "/books", nil).WithContext(ctx)
	res := httptest.NewRecorder()
	s.renderBookFormError(res, req, entities.Book{}, errInvalidForm)
	require.Equal(t, http.StatusBadRequest, res.Code)
	require.Contains(t, output.String(), `"event":"book_form.publishers_load_failed"`)
	require.Contains(t, output.String(), `"request_id":"`+testID+`"`)
	require.Contains(t, output.String(), "failed to reload publishers")
	require.NotContains(t, output.String(), `"event":"http.request_failed"`)
}
