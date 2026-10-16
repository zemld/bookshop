package handler

import (
	"context"
	"fmt"

	"bookshop/backend/internal/api/rest/convert"
	"bookshop/backend/internal/api/rest/ogen"
)

func (h Handler) CreateBook(ctx context.Context, req *ogen.BookInput) (*ogen.BookHeaders, error) {
	b, err := h.Books.CreateBook(ctx, convert.ConvertToDomainBook(req))
	if err != nil {
		return nil, fmt.Errorf("create book: %w", err)
	}

	return &ogen.BookHeaders{Response: convert.ConvertToOgenBook(b)}, nil
}
