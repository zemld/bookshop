package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"bookshop/frontend/internal/adapters/api/ogen"
	bookentities "bookshop/frontend/internal/domain/books/entities"
	publisherentities "bookshop/frontend/internal/domain/publishers/entities"
	bookports "bookshop/frontend/internal/ports/books"
	publisherports "bookshop/frontend/internal/ports/publishers"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/metric"
)

type Client struct {
	generated *ogen.Client
	http      *http.Client
}

type APIError struct {
	StatusCode int
	Code       string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API %d: %s", e.StatusCode, e.Code)
}

var (
	_ bookports.Books           = (*Client)(nil)
	_ publisherports.Publishers = (*Client)(nil)
)

func New(endpoint string, provider metric.MeterProvider) (*Client, error) {
	httpClient := &http.Client{
		Timeout:   10 * time.Second,
		Transport: requestIDTransport{next: http.DefaultTransport.(*http.Transport).Clone()},
	}
	generated, err := ogen.NewClient(endpoint, ogen.WithClient(httpClient), ogen.WithMeterProvider(provider))
	if err != nil {
		httpClient.CloseIdleConnections()
		return nil, fmt.Errorf("create API client: %w", err)
	}
	return &Client{generated: generated, http: httpClient}, nil
}

func (c *Client) Stop() {
	c.http.CloseIdleConnections()
}

func convertError(err error) error {
	if err == nil {
		return nil
	}
	var problem *ogen.ErrorStatusCodeWithHeaders
	if errors.As(err, &problem) {
		return &APIError{StatusCode: problem.StatusCode, Code: string(problem.Response.Error)}
	}
	return &APIError{StatusCode: http.StatusBadGateway, Code: "invalid_response"}
}

func convertBook(book ogen.Book) bookentities.Book {
	return bookentities.Book{
		ID: book.ID, Author: book.Author, Name: book.Name, Year: bookentities.Year(book.Year),
		Price: book.Price, PublisherID: book.PublisherId, PublicationYear: bookentities.Year(book.PublicationYear),
		Quantity: book.Quantity,
	}
}

func convertBookInput(book bookentities.Book) *ogen.BookInput {
	return &ogen.BookInput{
		Author: book.Author, Name: book.Name, Year: int(book.Year), Price: book.Price,
		PublisherId: book.PublisherID, PublicationYear: int(book.PublicationYear), Quantity: book.Quantity,
	}
}

func convertPublisher(publisher ogen.Publisher) publisherentities.Publisher {
	return publisherentities.Publisher{ID: publisher.ID, Name: publisher.Name}
}

func (c *Client) ListBooks(ctx context.Context) ([]bookentities.Book, error) {
	result, err := c.generated.ListBooks(ctx)
	if err != nil {
		return nil, convertError(err)
	}
	books := make([]bookentities.Book, len(result.Response))
	for index, book := range result.Response {
		books[index] = convertBook(book)
	}
	return books, nil
}

func (c *Client) GetBook(ctx context.Context, id uuid.UUID) (bookentities.Book, error) {
	result, err := c.generated.GetBook(ctx, ogen.GetBookParams{ID: id})
	if err != nil {
		return bookentities.Book{}, convertError(err)
	}
	return convertBook(result.Response), nil
}

func (c *Client) CreateBook(ctx context.Context, book bookentities.Book) (bookentities.Book, error) {
	result, err := c.generated.CreateBook(ctx, convertBookInput(book))
	if err != nil {
		return bookentities.Book{}, convertError(err)
	}
	return convertBook(result.Response), nil
}

func (c *Client) UpdateBook(ctx context.Context, id uuid.UUID, book bookentities.Book) (bookentities.Book, error) {
	result, err := c.generated.UpdateBook(ctx, convertBookInput(book), ogen.UpdateBookParams{ID: id})
	if err != nil {
		return bookentities.Book{}, convertError(err)
	}
	return convertBook(result.Response), nil
}

func (c *Client) DeleteBook(ctx context.Context, id uuid.UUID) error {
	_, err := c.generated.DeleteBook(ctx, ogen.DeleteBookParams{ID: id})
	return convertError(err)
}

func (c *Client) ListPublishers(ctx context.Context) ([]publisherentities.Publisher, error) {
	result, err := c.generated.ListPublishers(ctx)
	if err != nil {
		return nil, convertError(err)
	}
	publishers := make([]publisherentities.Publisher, len(result.Response))
	for index, publisher := range result.Response {
		publishers[index] = convertPublisher(publisher)
	}
	return publishers, nil
}

func (c *Client) GetPublisher(ctx context.Context, id uuid.UUID) (publisherentities.Publisher, error) {
	result, err := c.generated.GetPublisher(ctx, ogen.GetPublisherParams{ID: id})
	if err != nil {
		return publisherentities.Publisher{}, convertError(err)
	}
	return convertPublisher(result.Response), nil
}

func (c *Client) CreatePublisher(ctx context.Context, publisher publisherentities.Publisher) (publisherentities.Publisher, error) {
	result, err := c.generated.CreatePublisher(ctx, &ogen.PublisherInput{Name: publisher.Name})
	if err != nil {
		return publisherentities.Publisher{}, convertError(err)
	}
	return convertPublisher(result.Response), nil
}

func (c *Client) UpdatePublisher(ctx context.Context, id uuid.UUID, publisher publisherentities.Publisher) (publisherentities.Publisher, error) {
	result, err := c.generated.UpdatePublisher(ctx, &ogen.PublisherInput{Name: publisher.Name}, ogen.UpdatePublisherParams{ID: id})
	if err != nil {
		return publisherentities.Publisher{}, convertError(err)
	}
	return convertPublisher(result.Response), nil
}

func (c *Client) DeletePublisher(ctx context.Context, id uuid.UUID) error {
	_, err := c.generated.DeletePublisher(ctx, ogen.DeletePublisherParams{ID: id})
	return convertError(err)
}
