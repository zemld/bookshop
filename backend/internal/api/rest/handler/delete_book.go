package handler

import (
	"context"
	"fmt"

	"bookshop/backend/internal/api/rest/ogen"
)

func (h Handler) DeleteBook(ctx context.Context, params ogen.DeleteBookParams) (*ogen.DeletedHeaders, error) {
	if err := h.Books.DeleteBook(ctx, params.ID); err != nil {
		return nil, fmt.Errorf("delete book: %w", err)
	}

	return &ogen.DeletedHeaders{Response: ogen.Deleted{Status: _deletedStatus}}, nil
}
