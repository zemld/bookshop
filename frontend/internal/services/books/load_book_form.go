package books

import (
	"context"

	bookdomain "bookshop/frontend/internal/domain/books"

	"github.com/google/uuid"
)

// LoadBookForm loads the publisher choices before the book, preserving the UI's
// existing error precedence and allowing the selected publisher to render.
func (s Service) LoadBookForm(ctx context.Context, id uuid.UUID) (bookdomain.FormData, error) {
	publishers, err := s.Publishers.ListPublishers(ctx)
	if err != nil {
		return bookdomain.FormData{}, err
	}
	book, err := s.Books.GetBook(ctx, id)
	return bookdomain.FormData{Book: book, Publishers: publishers}, err
}
