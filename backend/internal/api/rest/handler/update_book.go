package handler

import (
	"context"
	"fmt"

	"bookshop/backend/internal/api/rest/convert"
	"bookshop/backend/internal/api/rest/ogen"
)

func (h Handler) UpdateBook(ctx context.Context, req *ogen.BookInput, params ogen.UpdateBookParams) (*ogen.Book, error) {
	b := convert.ToDomainBook(req)
	b.ID = params.ID

	updated, err := h.Books.UpdateBook(ctx, b)
	if err != nil {
		return nil, fmt.Errorf("update book: %w", err)
	}

	value := convert.ToOgenBook(updated)

	return &value, nil
}
