package handler

import (
	"context"
	"fmt"

	"bookshop/backend/internal/api/rest/ogen"
)

func (h Handler) DeleteBook(ctx context.Context, params ogen.DeleteBookParams) (*ogen.Deleted, error) {
	if err := h.Books.DeleteBook(ctx, params.ID); err != nil {
		return nil, fmt.Errorf("delete book: %w", err)
	}

	return &ogen.Deleted{Status: deletedStatus}, nil
}
