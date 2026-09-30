package books

import (
	"context"

	"bookshop/backend/internal/domain/books/entities"

	"github.com/google/uuid"
)

type ListBooks interface {
	ListBooks(context.Context) ([]entities.Book, error)
}

type GetBook interface {
	GetBook(context.Context, uuid.UUID) (entities.Book, error)
}

type CreateBook interface {
	CreateBook(context.Context, entities.Book) (entities.Book, error)
}

type UpdateBook interface {
	UpdateBook(context.Context, entities.Book) (entities.Book, error)
}

type DeleteBook interface {
	DeleteBook(context.Context, uuid.UUID) error
}

type Repository interface {
	ListBooks
	GetBook
	CreateBook
	UpdateBook
	DeleteBook
}
