package books

import (
	"context"
	"fmt"

	"bookshop/backend/internal/domain/books/entities"

	"github.com/google/uuid"
)

func (s Service) GetBook(ctx context.Context, id uuid.UUID) (entities.Book, error) {
	book, err := s.Repository.GetBook(ctx, id)
	if err != nil {
		return book, fmt.Errorf("get book: %w", err)
	}

	return book, nil
}
