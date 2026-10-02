package publishers

import (
	"context"
	"strings"

	"bookshop/frontend/internal/domain/publishers/entities"

	"github.com/google/uuid"
)

func (s Service) UpdatePublisher(ctx context.Context, id uuid.UUID, publisher entities.Publisher) (entities.Publisher, error) {
	publisher.Name = strings.TrimSpace(publisher.Name)
	return s.Publishers.UpdatePublisher(ctx, id, publisher)
}
