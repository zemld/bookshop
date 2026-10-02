package publishers

import (
	"context"

	"bookshop/frontend/internal/domain/publishers/entities"

	"github.com/google/uuid"
)

type ListPublishers interface {
	ListPublishers(context.Context) ([]entities.Publisher, error)
}
type GetPublisher interface {
	GetPublisher(context.Context, uuid.UUID) (entities.Publisher, error)
}
type CreatePublisher interface {
	CreatePublisher(context.Context, entities.Publisher) (entities.Publisher, error)
}

type UpdatePublisher interface {
	UpdatePublisher(context.Context, uuid.UUID, entities.Publisher) (entities.Publisher, error)
}

type DeletePublisher interface {
	DeletePublisher(context.Context, uuid.UUID) error
}

type WritePublisher interface {
	CreatePublisher
	UpdatePublisher
}

type Publishers interface {
	ListPublishers
	GetPublisher
	WritePublisher
	DeletePublisher
}
