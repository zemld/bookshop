package handler

import (
	"context"
	"fmt"

	"bookshop/backend/internal/api/rest/convert"
	"bookshop/backend/internal/api/rest/ogen"
)

func (h Handler) ListPublishers(ctx context.Context) ([]ogen.Publisher, error) {
	items, err := h.Publishers.ListPublishers(ctx)
	if err != nil {
		return nil, fmt.Errorf("list publishers: %w", err)
	}

	result := make([]ogen.Publisher, 0, len(items))
	for _, p := range items {
		result = append(result, convert.ToOgenPublisher(p))
	}

	return result, nil
}
