package publishers

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

func (s Service) DeletePublisher(ctx context.Context, id uuid.UUID) error {
	if err := s.Repository.DeletePublisher(ctx, id); err != nil {
		return fmt.Errorf("delete publisher: %w", err)
	}

	return nil
}
