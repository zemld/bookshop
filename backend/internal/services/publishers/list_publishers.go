package publishers

import (
	"context"
	"fmt"

	"bookshop/backend/internal/domain/publishers/entities"
)

func (s Service) ListPublishers(ctx context.Context) ([]entities.Publisher, error) {
	publishers, err := s.Repository.ListPublishers(ctx)
	if err != nil {
		return publishers, fmt.Errorf("list publishers: %w", err)
	}

	return publishers, nil
}
