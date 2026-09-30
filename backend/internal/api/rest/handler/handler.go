package handler

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"bookshop/backend/internal/api/rest/ogen"
	"bookshop/backend/internal/domain/books"
	"bookshop/backend/internal/domain/publishers"
	"bookshop/backend/internal/domain/shared"
)

type Handler struct {
	Books      books.Books
	Publishers publishers.Publishers
}

var _ ogen.Handler = Handler{}

const deletedStatus = "deleted"

func classifyError(err error) (int, string) {
	switch {
	case errors.Is(err, shared.ErrInvalid):
		return http.StatusBadRequest, shared.ErrInvalid.Error()
	case errors.Is(err, shared.ErrConflict):
		return http.StatusConflict, shared.ErrConflict.Error()
	case errors.Is(err, shared.ErrNotFound):
		return http.StatusNotFound, shared.ErrNotFound.Error()
	default:
		return http.StatusInternalServerError, "internal server error"
	}
}

func (Handler) NewError(_ context.Context, err error) *ogen.ErrorStatusCode {
	code, message := classifyError(err)
	if code == http.StatusInternalServerError {
		slog.Error("request failed", "error", err)
	}

	return &ogen.ErrorStatusCode{StatusCode: code, Response: ogen.Problem{Error: message}}
}
