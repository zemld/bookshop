package books

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

func (s Service) DeleteBook(ctx context.Context, id uuid.UUID) error {
	if err := s.Repository.DeleteBook(ctx, id); err != nil {
		return fmt.Errorf("delete book: %w", err)
	}

	return nil
}
