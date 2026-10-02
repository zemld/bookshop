package entities

import "github.com/google/uuid"

type Publisher struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}
