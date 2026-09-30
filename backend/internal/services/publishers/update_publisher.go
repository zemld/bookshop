package publishers

import (
	"context"
	"fmt"

	"bookshop/backend/internal/domain/publishers/entities"
)

func (s Service) UpdatePublisher(ctx context.Context, p entities.Publisher) (entities.Publisher, error) {
	if err := p.Validate(); err != nil {
		return p, fmt.Errorf("validate publisher: %w", err)
	}

	updated, err := s.Repository.UpdatePublisher(ctx, p)
	if err != nil {
		return updated, fmt.Errorf("update publisher: %w", err)
	}

	return updated, nil
}
