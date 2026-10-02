package http

import (
	"context"
	"net/http"

	core "bookshop/frontend/internal/adapters/httpcore"
	"bookshop/frontend/internal/domain/books/entities"
	"bookshop/frontend/internal/ports/books"

	"github.com/google/uuid"
)

type Client struct{ Core *core.Client }

var _ books.Books = (*Client)(nil)

func (c *Client) ListBooks(ctx context.Context) ([]entities.Book, error) {
	var books []entities.Book
	err := c.Core.Request(ctx, http.MethodGet, "/books", nil, &books)
	return books, err
}

func (c *Client) GetBook(ctx context.Context, id uuid.UUID) (entities.Book, error) {
	var book entities.Book
	err := c.Core.Request(ctx, http.MethodGet, "/books/"+id.String(), nil, &book)
	return book, err
}

func (c *Client) CreateBook(ctx context.Context, book entities.Book) (entities.Book, error) {
	var saved entities.Book
	err := c.Core.Request(ctx, http.MethodPost, "/books", book, &saved)
	return saved, err
}

func (c *Client) UpdateBook(ctx context.Context, id uuid.UUID, book entities.Book) (entities.Book, error) {
	var saved entities.Book
	err := c.Core.Request(ctx, http.MethodPut, "/books/"+id.String(), book, &saved)
	return saved, err
}

func (c *Client) DeleteBook(ctx context.Context, id uuid.UUID) error {
	return c.Core.Request(ctx, http.MethodDelete, "/books/"+id.String(), nil, nil)
}
