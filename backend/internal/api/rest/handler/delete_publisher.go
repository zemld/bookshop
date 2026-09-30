package handler

import (
	"context"
	"fmt"

	"bookshop/backend/internal/api/rest/ogen"
)

func (h Handler) DeletePublisher(ctx context.Context, params ogen.DeletePublisherParams) (*ogen.Deleted, error) {
	if err := h.Publishers.DeletePublisher(ctx, params.ID); err != nil {
		return nil, fmt.Errorf("delete publisher: %w", err)
	}

	return &ogen.Deleted{Status: deletedStatus}, nil
}
