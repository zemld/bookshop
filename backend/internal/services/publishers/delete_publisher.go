package publishers

import (
	"context"
	"errors"
	"fmt"

	"bookshop/backend/internal/domain/publishers/entities"
	"bookshop/backend/internal/domain/shared"

	"github.com/google/uuid"
)

func (s Service) DeletePublisher(ctx context.Context, id uuid.UUID) error {
	if err := s.Repository.DeletePublisher(ctx, id); err != nil {
		if errors.Is(err, shared.ErrConflict) {
			err = entities.ErrReferenced
		}

		return fmt.Errorf("delete publisher: %w", err)
	}

	return nil
}
