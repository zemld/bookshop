package web

import (
	"errors"
	"net/url"
	"strconv"

	"bookshop/frontend/internal/domain/books/entities"

	"github.com/google/uuid"
)

var errInvalidForm = errors.New("invalid form")

type formValidationError struct {
	Code string
}

func (*formValidationError) Error() string {
	return "invalid form"
}
func (*formValidationError) Unwrap() error { return errInvalidForm }

func parseBookForm(f url.Values) (entities.Book, error) {
	var b entities.Book
	b.Author = f.Get("author")
	b.Name = f.Get("name")
	year, err := strconv.Atoi(f.Get("year"))
	if err != nil {
		return b, &formValidationError{Code: "invalid_year"}
	}
	b.Year = entities.Year(year)
	b.Price, err = strconv.ParseInt(f.Get("price"), 10, 64)
	if err != nil {
		return b, &formValidationError{Code: "invalid_price"}
	}
	b.PublisherID, err = uuid.Parse(f.Get("publisherId"))
	if err != nil {
		return b, &formValidationError{Code: "invalid_publisher_id"}
	}
	publicationYear, err := strconv.Atoi(f.Get("publicationYear"))
	if err != nil {
		return b, &formValidationError{Code: "invalid_publication_year"}
	}
	b.PublicationYear = entities.Year(publicationYear)
	b.Quantity, err = strconv.ParseInt(f.Get("quantity"), 10, 64)
	if err != nil {
		return b, &formValidationError{Code: "invalid_quantity"}
	}
	return b, nil
}

func parseRouteID(rid string) (uuid.UUID, error) {
	id, err := uuid.Parse(rid)
	if err != nil {
		return uuid.Nil, &formValidationError{Code: "invalid_id"}
	}
	return id, nil
}
