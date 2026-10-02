package publishers

import (
	"context"
	"strings"

	"bookshop/frontend/internal/domain/publishers/entities"
)

func (s Service) CreatePublisher(ctx context.Context, publisher entities.Publisher) (entities.Publisher, error) {
	publisher.Name = strings.TrimSpace(publisher.Name)
	return s.Publishers.CreatePublisher(ctx, publisher)
}
