package publishers

import (
	"context"

	"bookshop/frontend/internal/domain/publishers/entities"

	"github.com/google/uuid"
)

type CreatePublisher interface {
	CreatePublisher(context.Context, entities.Publisher) (entities.Publisher, error)
}

type UpdatePublisher interface {
	UpdatePublisher(context.Context, uuid.UUID, entities.Publisher) (entities.Publisher, error)
}
