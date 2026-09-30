package entities

import (
	"strings"

	"bookshop/backend/internal/domain/shared"

	"github.com/google/uuid"
)

type Year int

func (y Year) Validate() error {
	if y < 1 || y > 9999 {
		return shared.ErrInvalid
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

	if err := b.Year.Validate(); err != nil {
		return err
	}

	if err := b.PublicationYear.Validate(); err != nil {
		return err
	}

	if b.Author == "" || b.Name == "" || b.PublisherID == uuid.Nil ||
		b.PublicationYear < b.Year || b.Price < 0 || b.Quantity < 0 {
		return shared.ErrInvalid
	}

	return nil
}
