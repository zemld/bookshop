package books

import (
	"context"
	"fmt"

	"bookshop/backend/internal/domain/books/entities"
)

func (s Service) UpdateBook(ctx context.Context, b entities.Book) (entities.Book, error) {
	if err := b.Validate(); err != nil {
		return b, fmt.Errorf("validate book: %w", err)
	}

	updated, err := s.Repository.UpdateBook(ctx, b)
	if err != nil {
		return updated, fmt.Errorf("update book: %w", err)
	}

	return updated, nil
}
