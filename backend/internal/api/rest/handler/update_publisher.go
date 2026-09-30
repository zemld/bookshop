package handler

import (
	"context"
	"fmt"

	"bookshop/backend/internal/api/rest/convert"
	"bookshop/backend/internal/api/rest/ogen"
	"bookshop/backend/internal/domain/publishers/entities"
)

func (h Handler) UpdatePublisher(ctx context.Context, req *ogen.PublisherInput, params ogen.UpdatePublisherParams) (*ogen.Publisher, error) {
	p, err := h.Publishers.UpdatePublisher(ctx, entities.Publisher{ID: params.ID, Name: req.Name})
	if err != nil {
		return nil, fmt.Errorf("update publisher: %w", err)
	}

	value := convert.ToOgenPublisher(p)

	return &value, nil
}
