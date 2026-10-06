package entities

import (
	"strings"

	"github.com/google/uuid"
)

type Year int

func (y Year) Validate() error {
	if y < 1 || y > 9999 {
		return ErrYearInvalid
	}

	return nil
}

type Book struct {
	ID              uuid.UUID `json:"id"`
	Author          string    `json:"author"`
	Name            string    `json:"name"`
	Year            Year      `json:"year"`
	Price           int64     `json:"price"`
	PublisherID     uuid.UUID `json:"publisherId"`
	PublicationYear Year      `json:"publicationYear"`
	Quantity        int64     `json:"quantity"`
}

func (b *Book) Validate() error {
	b.Author = strings.TrimSpace(b.Author)
	b.Name = strings.TrimSpace(b.Name)

	if b.Author == "" {
		return ErrAuthorRequired
	}

	if b.Name == "" {
		return ErrNameRequired
	}

	if b.Year.Validate() != nil {
		return ErrYearInvalid
	}

	if b.Price < 0 {
		return ErrPriceNegative
	}

	if b.PublisherID == uuid.Nil {
		return ErrPublisherRequired
	}

	if b.PublicationYear.Validate() != nil {
		return ErrPublicationYearInvalid
	}

	if b.PublicationYear < b.Year {
		return ErrPublicationYearBeforeYear
	}

	if b.Quantity < 0 {
		return ErrQuantityNegative
	}

	return nil
}
