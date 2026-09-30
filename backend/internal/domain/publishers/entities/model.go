package entities

import (
	"strings"

	"bookshop/backend/internal/domain/shared"

	"github.com/google/uuid"
)

type Publisher struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

func (p *Publisher) Validate() error {
	p.Name = strings.TrimSpace(p.Name)
	if p.Name == "" {
		return shared.ErrInvalid
	}

	return nil
}
