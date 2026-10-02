package http

import (
	"context"
	"net/http"

	core "bookshop/frontend/internal/adapters/httpcore"
	"bookshop/frontend/internal/domain/publishers/entities"
	"bookshop/frontend/internal/ports/publishers"

	"github.com/google/uuid"
)

type Client struct{ Core *core.Client }

var _ publishers.Publishers = (*Client)(nil)

func (c *Client) ListPublishers(ctx context.Context) ([]entities.Publisher, error) {
	var publishers []entities.Publisher
	err := c.Core.Request(ctx, http.MethodGet, "/publishers", nil, &publishers)
	return publishers, err
}

func (c *Client) GetPublisher(ctx context.Context, id uuid.UUID) (entities.Publisher, error) {
	var publisher entities.Publisher
	err := c.Core.Request(ctx, http.MethodGet, "/publishers/"+id.String(), nil, &publisher)
	return publisher, err
}

func (c *Client) CreatePublisher(ctx context.Context, publisher entities.Publisher) (entities.Publisher, error) {
	var saved entities.Publisher
	err := c.Core.Request(ctx, http.MethodPost, "/publishers", publisher, &saved)
	return saved, err
}

func (c *Client) UpdatePublisher(ctx context.Context, id uuid.UUID, publisher entities.Publisher) (entities.Publisher, error) {
	var saved entities.Publisher
	err := c.Core.Request(ctx, http.MethodPut, "/publishers/"+id.String(), publisher, &saved)
	return saved, err
}

func (c *Client) DeletePublisher(ctx context.Context, id uuid.UUID) error {
	return c.Core.Request(ctx, http.MethodDelete, "/publishers/"+id.String(), nil, nil)
}
