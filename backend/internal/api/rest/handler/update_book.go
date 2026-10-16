package handler

import (
	"context"
	"fmt"

	"bookshop/backend/internal/api/rest/convert"
	"bookshop/backend/internal/api/rest/ogen"
)

func (h Handler) UpdateBook(ctx context.Context, req *ogen.BookInput, params ogen.UpdateBookParams) (*ogen.BookHeaders, error) {
	b := convert.ConvertToDomainBook(req)
	b.ID = params.ID

	updated, err := h.Books.UpdateBook(ctx, b)
	if err != nil {
		return nil, fmt.Errorf("update book: %w", err)
	}

	return &ogen.BookHeaders{Response: convert.ConvertToOgenBook(updated)}, nil
}
