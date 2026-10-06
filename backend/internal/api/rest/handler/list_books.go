package handler

import (
	"context"
	"fmt"

	"bookshop/backend/internal/api/rest/convert"
	"bookshop/backend/internal/api/rest/ogen"
)

func (h Handler) ListBooks(ctx context.Context) ([]ogen.Book, error) {
	items, err := h.Books.ListBooks(ctx)
	if err != nil {
		return nil, fmt.Errorf("list books: %w", err)
	}

	result := make([]ogen.Book, 0, len(items))
	for _, b := range items {
		result = append(result, convert.ConvertToOgenBook(b))
	}

	return result, nil
}
