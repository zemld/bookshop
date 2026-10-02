package entities

import "github.com/google/uuid"

// Year is the wire-compatible year value; backend owns its business constraints.
type Year int

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
