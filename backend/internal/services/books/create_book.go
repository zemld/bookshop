package books

import (
	"context"
	"fmt"

	"bookshop/backend/internal/domain/books/entities"
)

func (s Service) CreateBook(ctx context.Context, b entities.Book) (entities.Book, error) {
	if err := b.Validate(); err != nil {
		return b, fmt.Errorf("validate book: %w", err)
	}

	created, err := s.Repository.CreateBook(ctx, b)
	if err != nil {
		return created, fmt.Errorf("create book: %w", err)
	}

	return created, nil
}
