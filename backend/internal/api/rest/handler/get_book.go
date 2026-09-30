package handler

import (
	"context"
	"fmt"

	"bookshop/backend/internal/api/rest/convert"
	"bookshop/backend/internal/api/rest/ogen"
)

func (h Handler) GetBook(ctx context.Context, params ogen.GetBookParams) (*ogen.Book, error) {
	b, err := h.Books.GetBook(ctx, params.ID)
	if err != nil {
		return nil, fmt.Errorf("get book: %w", err)
	}

	value := convert.ToOgenBook(b)

	return &value, nil
}
