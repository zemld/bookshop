package books

import (
	"context"
	"fmt"

	"bookshop/backend/internal/domain/books/entities"
)

func (s Service) ListBooks(ctx context.Context) ([]entities.Book, error) {
	books, err := s.Repository.ListBooks(ctx)
	if err != nil {
		return books, fmt.Errorf("list books: %w", err)
	}

	return books, nil
}
