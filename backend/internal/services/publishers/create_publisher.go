package publishers

import (
	"context"
	"fmt"

	"bookshop/backend/internal/domain/publishers/entities"
)

func (s Service) CreatePublisher(ctx context.Context, p entities.Publisher) (entities.Publisher, error) {
	if err := p.Validate(); err != nil {
		return p, fmt.Errorf("validate publisher: %w", err)
	}

	created, err := s.Repository.CreatePublisher(ctx, p)
	if err != nil {
		return created, fmt.Errorf("create publisher: %w", err)
	}

	return created, nil
}
