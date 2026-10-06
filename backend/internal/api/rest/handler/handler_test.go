package handler

import (
	"context"
	"errors"
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"

	"bookshop/backend/internal/api/rest/server"
	bookentities "bookshop/backend/internal/domain/books/entities"
	publisherentities "bookshop/backend/internal/domain/publishers/entities"
	"bookshop/backend/internal/domain/shared"
	"bookshop/backend/internal/services/books"
	"bookshop/backend/internal/services/publishers"
	"bookshop/backend/internal/utils/postgreserrors"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/require"
)

func TestRejectInvalidRequest(t *testing.T) {
	api := server.New(Handler{Books: books.Service{}, Publishers: publishers.Service{}})

	for _, tc := range []struct {
		name, method, path, body, contentType, want string
	}{
		{"malformed ID", "GET", "/books/no-uuid", "", "", `{"error":"invalid_id"}`},
		{"malformed JSON", "POST", "/books", "{", "application/json", `{"error":"invalid_json"}`},
		{"unknown field", "POST", "/publishers", `{"unknown":1}`, "application/json", `{"error":"invalid_name"}`},
		{"trailing JSON", "POST", "/publishers", `{"name":"A"} {}`, "application/json", `{"error":"trailing_json_data"}`},
		{"incorrect content type", "POST", "/publishers", `{"name":"A"}`, "text/plain", `{"error":"invalid_request"}`},
		{"blank publisher", "POST", "/publishers", `{"name":"   "}`, "application/json", `{"error":"publisher_name_required"}`},
		{"absent book years", "POST", "/books", `{"author":"A","name":"B"}`, "application/json", `{"error":"invalid_year"}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
			if tc.contentType != "" {
				req.Header.Set("Content-Type", tc.contentType)
			}

			res := httptest.NewRecorder()
			api.ServeHTTP(res, req)
			require.Equal(t, 400, res.Code)
			require.JSONEq(t, tc.want, res.Body.String())
		})
	}
}

func TestValidationResponseForCreateAndUpdate(t *testing.T) {
	api := server.New(Handler{Books: books.Service{}})

	for _, path := range []struct{ method, path string }{
		{"POST", "/books"}, {"PUT", "/books/aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"},
	} {
		t.Run(path.method, func(t *testing.T) {
			for _, tc := range []struct{ body, want string }{
				{`{"author":" ","name":"","year":0,"price":-1,"publisherId":"00000000-0000-0000-0000-000000000000","publicationYear":10000,"quantity":-1}`, "book_author_required"},
				{`{"author":"A","name":"B","year":2000,"price":-1,"publisherId":"00000000-0000-0000-0000-000000000000","publicationYear":10000,"quantity":-1}`, "book_price_negative"},
			} {
				req := httptest.NewRequest(path.method, path.path, strings.NewReader(tc.body))
				req.Header.Set("Content-Type", "application/json")

				res := httptest.NewRecorder()
				api.ServeHTTP(res, req)
				require.Equal(t, 400, res.Code)
				require.JSONEq(t, `{"error":"`+tc.want+`"}`, res.Body.String())
			}
		})
	}
}

func TestDecodeFieldErrors(t *testing.T) {
	api := server.New(Handler{})

	for _, tc := range []struct{ body, want string }{
		{`{"author":"A","name":"B"}`, "invalid_year"},
		{`{"author":2}`, "invalid_author"},
		{`{"name":2}`, "invalid_name"},
		{`{"year":"bad"}`, "invalid_year"},
		{`{"price":"bad"}`, "invalid_price"},
		{`{"publisherId":"bad"}`, "invalid_publisher_id"},
		{`{"publicationYear":"bad"}`, "invalid_publication_year"},
		{`{"quantity":"bad"}`, "invalid_quantity"},
	} {
		req := httptest.NewRequest("POST", "/books", strings.NewReader(tc.body))
		req.Header.Set("Content-Type", "application/json")

		res := httptest.NewRecorder()
		api.ServeHTTP(res, req)
		require.Equal(t, 400, res.Code)
		require.JSONEq(t, `{"error":"`+tc.want+`"}`, res.Body.String())
	}
}

func TestPublisherValidationResponseForCreateAndUpdate(t *testing.T) {
	api := server.New(Handler{Publishers: publishers.Service{}})

	for _, tc := range []struct{ method, path string }{
		{"POST", "/publishers"},
		{"PUT", "/publishers/aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"},
	} {
		req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(`{"name":"  "}`))
		req.Header.Set("Content-Type", "application/json")

		res := httptest.NewRecorder()
		api.ServeHTTP(res, req)
		require.Equal(t, 400, res.Code)
		require.JSONEq(t, `{"error":"publisher_name_required"}`, res.Body.String())
	}
}

func TestMapDomainAndDatabaseErrorsToProblem(t *testing.T) {
	for _, tc := range []struct {
		name     string
		err      error
		status   int
		wireCode string
	}{
		{"book author", bookentities.ErrAuthorRequired, 400, "book_author_required"},
		{"book name", bookentities.ErrNameRequired, 400, "book_name_required"},
		{"book year", bookentities.ErrYearInvalid, 400, "book_year_invalid"},
		{"book price", bookentities.ErrPriceNegative, 400, "book_price_negative"},
		{"book publisher", bookentities.ErrPublisherRequired, 400, "book_publisher_required"},
		{"book publication year", bookentities.ErrPublicationYearInvalid, 400, "book_publication_year_invalid"},
		{"book publication before year", bookentities.ErrPublicationYearBeforeYear, 400, "book_publication_year_before_year"},
		{"book validation", bookentities.ErrQuantityNegative, 400, "book_quantity_negative"},
		{"publisher name", publisherentities.ErrNameRequired, 400, "publisher_name_required"},
		{"duplicate publisher", postgreserrors.MapDatabaseError(&pgconn.PgError{Code: "23505", ConstraintName: "publishers_normalized_name_unique", Detail: "private SQL"}), 409, "publisher_name_duplicate"},
		{"duplicate book", postgreserrors.MapDatabaseError(&pgconn.PgError{Code: "23505", ConstraintName: "books_content_unique", Detail: "private SQL"}), 409, "book_duplicate"},
		{"publisher missing", postgreserrors.MapDatabaseError(&pgconn.PgError{Code: "23503", ConstraintName: "books_publisher_id_fkey", Detail: "private SQL"}), 409, "book_publisher_not_found"},
		{"referenced publisher", publisherentities.ErrReferenced, 409, "publisher_referenced"},
		{"missing", shared.ErrNotFound, 404, "not_found"},
		{"unrecognized validation", shared.ErrInvalid, 400, "invalid_input"},
		{"unrecognized conflict", shared.ErrConflict, 409, "conflict"},
		{"internal", errors.New("private database error"), 500, "internal_error"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			response := (Handler{}).NewError(context.Background(), fmt.Errorf("operation: %w", tc.err))
			require.Equal(t, tc.status, response.StatusCode)
			require.Equal(t, tc.wireCode, string(response.Response.Error))
		})
	}
}
