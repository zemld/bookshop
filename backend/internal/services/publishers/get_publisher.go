package publishers

import (
	"context"
	"fmt"

	"bookshop/backend/internal/domain/publishers/entities"

	"github.com/google/uuid"
)

func (s Service) GetPublisher(ctx context.Context, id uuid.UUID) (entities.Publisher, error) {
	publisher, err := s.Repository.GetPublisher(ctx, id)
	if err != nil {
		return publisher, fmt.Errorf("get publisher: %w", err)
	}

	return publisher, nil
}
