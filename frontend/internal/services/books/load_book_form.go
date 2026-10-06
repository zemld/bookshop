package books

import (
	"context"

	bookdomain "bookshop/frontend/internal/domain/books"

	"github.com/google/uuid"
)

func (s Service) LoadBookForm(ctx context.Context, id uuid.UUID) (bookdomain.FormData, error) {
	publishers, err := s.Publishers.ListPublishers(ctx)
	if err != nil {
		return bookdomain.FormData{}, err
	}
	book, err := s.Books.GetBook(ctx, id)
	return bookdomain.FormData{Book: book, Publishers: publishers}, err
}
