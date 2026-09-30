package handler

import (
	"context"
	"fmt"

	"bookshop/backend/internal/api/rest/convert"
	"bookshop/backend/internal/api/rest/ogen"
	"bookshop/backend/internal/domain/publishers/entities"
)

func (h Handler) CreatePublisher(ctx context.Context, req *ogen.PublisherInput) (*ogen.Publisher, error) {
	p, err := h.Publishers.CreatePublisher(ctx, entities.Publisher{Name: req.Name})
	if err != nil {
		return nil, fmt.Errorf("create publisher: %w", err)
	}

	value := convert.ToOgenPublisher(p)

	return &value, nil
}
