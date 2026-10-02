package web

import (
	"errors"
	"net/url"
	"strconv"

	"bookshop/frontend/internal/domain/books/entities"

	"github.com/google/uuid"
)

var errInvalidForm = errors.New("invalid form")

// parseBookForm checks wire syntax only; backend decides business validity.
func parseBookForm(f url.Values) (entities.Book, error) {
	var b entities.Book
	var err error
	b.Author = f.Get("author")
	b.Name = f.Get("name")
	y, e := strconv.Atoi(f.Get("year"))
	if e != nil {
		return b, errInvalidForm
	}
	b.Year = entities.Year(y)
	y, e = strconv.Atoi(f.Get("publicationYear"))
	if e != nil {
		return b, errInvalidForm
	}
	b.PublicationYear = entities.Year(y)
	b.Price, err = strconv.ParseInt(f.Get("price"), 10, 64)
	if err != nil {
		return b, errInvalidForm
	}
	b.Quantity, err = strconv.ParseInt(f.Get("quantity"), 10, 64)
	if err != nil {
		return b, errInvalidForm
	}
	b.PublisherID, err = uuid.Parse(f.Get("publisherId"))
	if err != nil {
		return b, errInvalidForm
	}
	return b, nil
}

func routeID(rid string) (uuid.UUID, error) {
	id, err := uuid.Parse(rid)
	if err != nil {
		return uuid.Nil, errInvalidForm
	}
	return id, nil
}
