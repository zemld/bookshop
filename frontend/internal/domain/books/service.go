package books

import (
	"context"

	"bookshop/frontend/internal/domain/books/entities"
	publisherentities "bookshop/frontend/internal/domain/publishers/entities"

	"github.com/google/uuid"
)

type FormData struct {
	Book       entities.Book
	Publishers []publisherentities.Publisher
}

type LoadBookForm interface {
	LoadBookForm(context.Context, uuid.UUID) (FormData, error)
}
