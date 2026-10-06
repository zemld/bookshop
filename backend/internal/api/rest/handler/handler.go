package handler

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"bookshop/backend/internal/api/rest/ogen"
	"bookshop/backend/internal/domain/books"
	bookentities "bookshop/backend/internal/domain/books/entities"
	"bookshop/backend/internal/domain/publishers"
	publisherentities "bookshop/backend/internal/domain/publishers/entities"
	"bookshop/backend/internal/domain/shared"
)

type Handler struct {
	Books      books.Books
	Publishers publishers.Publishers
}

var _ ogen.Handler = Handler{}

const deletedStatus = "deleted"

func classifyError(err error) (int, ogen.ProblemError) {
	switch {
	case errors.Is(err, shared.ErrInvalid):
		return http.StatusBadRequest, ogen.ProblemErrorInvalidInput
	case errors.Is(err, shared.ErrConflict):
		return http.StatusConflict, ogen.ProblemErrorConflict
	case errors.Is(err, shared.ErrNotFound):
		return http.StatusNotFound, ogen.ProblemErrorNotFound
	default:
		return http.StatusInternalServerError, ogen.ProblemErrorInternalError
	}
}

var errorCodes = map[error]ogen.ProblemError{
	bookentities.ErrAuthorRequired:            ogen.ProblemErrorBookAuthorRequired,
	bookentities.ErrNameRequired:              ogen.ProblemErrorBookNameRequired,
	bookentities.ErrYearInvalid:               ogen.ProblemErrorBookYearInvalid,
	bookentities.ErrPriceNegative:             ogen.ProblemErrorBookPriceNegative,
	bookentities.ErrPublisherRequired:         ogen.ProblemErrorBookPublisherRequired,
	bookentities.ErrPublicationYearInvalid:    ogen.ProblemErrorBookPublicationYearInvalid,
	bookentities.ErrPublicationYearBeforeYear: ogen.ProblemErrorBookPublicationYearBeforeYear,
	bookentities.ErrQuantityNegative:          ogen.ProblemErrorBookQuantityNegative,
	publisherentities.ErrNameRequired:         ogen.ProblemErrorPublisherNameRequired,
	publisherentities.ErrNameDuplicate:        ogen.ProblemErrorPublisherNameDuplicate,
	bookentities.ErrDuplicate:                 ogen.ProblemErrorBookDuplicate,
	bookentities.ErrPublisherNotFound:         ogen.ProblemErrorBookPublisherNotFound,
	publisherentities.ErrReferenced:           ogen.ProblemErrorPublisherReferenced,
}

func identifyError(err error, category ogen.ProblemError) ogen.ProblemError {
	for cause, code := range errorCodes {
		if errors.Is(err, cause) {
			return code
		}
	}

	return category
}

func (Handler) NewError(_ context.Context, err error) *ogen.ErrorStatusCode {
	code, category := classifyError(err)
	if code == http.StatusInternalServerError {
		slog.Error("request failed", "error", err)
	}

	if code != http.StatusInternalServerError {
		category = identifyError(err, category)
	}

	return &ogen.ErrorStatusCode{StatusCode: code, Response: ogen.Problem{Error: category}}
}
