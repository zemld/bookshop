package handler

import (
	"context"
	"fmt"

	"bookshop/backend/internal/api/rest/convert"
	"bookshop/backend/internal/api/rest/ogen"
)

func (h Handler) GetPublisher(ctx context.Context, params ogen.GetPublisherParams) (*ogen.Publisher, error) {
	p, err := h.Publishers.GetPublisher(ctx, params.ID)
	if err != nil {
		return nil, fmt.Errorf("get publisher: %w", err)
	}

	value := convert.ConvertToOgenPublisher(p)

	return &value, nil
}
