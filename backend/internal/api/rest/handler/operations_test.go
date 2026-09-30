package handler

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	bookmocks "bookshop/backend/internal/adapters/books/mocks"
	publishermocks "bookshop/backend/internal/adapters/publishers/mocks"
	"bookshop/backend/internal/api/rest/server"
	bookentities "bookshop/backend/internal/domain/books/entities"
	publisherentities "bookshop/backend/internal/domain/publishers/entities"
	"bookshop/backend/internal/domain/shared"
	bookservice "bookshop/backend/internal/services/books"
	publisherservice "bookshop/backend/internal/services/publishers"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestOperationsHTTP(t *testing.T) {
	t.Parallel()

	bookID := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	publisherID := uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")
	book := bookentities.Book{
		ID: bookID, Author: "Alice", Name: "A book", Year: 2020, PublicationYear: 2021,
		PublisherID: publisherID, Price: 100, Quantity: 3,
	}
	bookInput := `{"author":"Alice","name":"A book","year":2020,"publicationYear":2021,"publisherId":"bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb","price":100,"quantity":3}`
	bookJSON := `{"id":"aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa","author":"Alice","name":"A book","year":2020,"publicationYear":2021,"publisherId":"bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb","price":100,"quantity":3}`
	publisher := publisherentities.Publisher{ID: publisherID, Name: "Publisher"}
	publisherJSON := `{"id":"bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb","name":"Publisher"}`

	tests := []struct {
		name, method, path, body, wantJSON string
		wantStatus                         int
		setup                              func(*bookmocks.MockBooks, *publishermocks.MockPublishers)
	}{
		{
			name: "list books", method: http.MethodGet, path: "/books", wantStatus: http.StatusOK,
			wantJSON: "[" + bookJSON + "]",
			setup: func(b *bookmocks.MockBooks, _ *publishermocks.MockPublishers) {
				b.EXPECT().ListBooks(mock.Anything).Return([]bookentities.Book{book}, nil)
			},
		},
		{
			name: "list empty books", method: http.MethodGet, path: "/books", wantStatus: http.StatusOK,
			wantJSON: "[]",
			setup: func(b *bookmocks.MockBooks, _ *publishermocks.MockPublishers) {
				b.EXPECT().ListBooks(mock.Anything).Return(nil, nil)
			},
		},
		{
			name: "get book", method: http.MethodGet, path: "/books/" + bookID.String(), wantStatus: http.StatusOK,
			wantJSON: bookJSON,
			setup: func(b *bookmocks.MockBooks, _ *publishermocks.MockPublishers) {
				b.EXPECT().GetBook(mock.Anything, bookID).Return(book, nil)
			},
		},
		{
			name: "create book", method: http.MethodPost, path: "/books", body: bookInput, wantStatus: http.StatusOK,
			wantJSON: bookJSON,
			setup: func(b *bookmocks.MockBooks, _ *publishermocks.MockPublishers) {
				input := book
				input.ID = uuid.Nil
				b.EXPECT().CreateBook(mock.Anything, input).Return(book, nil)
			},
		},
		{
			name: "update book", method: http.MethodPut, path: "/books/" + bookID.String(), body: bookInput, wantStatus: http.StatusOK,
			wantJSON: bookJSON,
			setup: func(b *bookmocks.MockBooks, _ *publishermocks.MockPublishers) {
				b.EXPECT().UpdateBook(mock.Anything, book).Return(book, nil)
			},
		},
		{
			name: "delete book", method: http.MethodDelete, path: "/books/" + bookID.String(), wantStatus: http.StatusOK,
			wantJSON: `{"status":"deleted"}`,
			setup: func(b *bookmocks.MockBooks, _ *publishermocks.MockPublishers) {
				b.EXPECT().DeleteBook(mock.Anything, bookID).Return(nil)
			},
		},
		{
			name: "missing book", method: http.MethodGet, path: "/books/" + bookID.String(), wantStatus: http.StatusNotFound,
			wantJSON: `{"error":"record not found"}`,
			setup: func(b *bookmocks.MockBooks, _ *publishermocks.MockPublishers) {
				b.EXPECT().GetBook(mock.Anything, bookID).Return(bookentities.Book{}, shared.ErrNotFound)
			},
		},
		{
			name: "invalid book", method: http.MethodPost, path: "/books", body: bookInput,
			wantStatus: http.StatusBadRequest, wantJSON: `{"error":"invalid input"}`,
			setup: func(b *bookmocks.MockBooks, _ *publishermocks.MockPublishers) {
				input := book
				input.ID = uuid.Nil
				b.EXPECT().CreateBook(mock.Anything, input).Return(bookentities.Book{}, shared.ErrInvalid)
			},
		},
		{
			name: "list publishers", method: http.MethodGet, path: "/publishers", wantStatus: http.StatusOK,
			wantJSON: "[" + publisherJSON + "]",
			setup: func(_ *bookmocks.MockBooks, p *publishermocks.MockPublishers) {
				p.EXPECT().ListPublishers(mock.Anything).Return([]publisherentities.Publisher{publisher}, nil)
			},
		},
		{
			name: "list empty publishers", method: http.MethodGet, path: "/publishers", wantStatus: http.StatusOK,
			wantJSON: "[]",
			setup: func(_ *bookmocks.MockBooks, p *publishermocks.MockPublishers) {
				p.EXPECT().ListPublishers(mock.Anything).Return(nil, nil)
			},
		},
		{
			name: "get publisher", method: http.MethodGet, path: "/publishers/" + publisherID.String(), wantStatus: http.StatusOK,
			wantJSON: publisherJSON,
			setup: func(_ *bookmocks.MockBooks, p *publishermocks.MockPublishers) {
				p.EXPECT().GetPublisher(mock.Anything, publisherID).Return(publisher, nil)
			},
		},
		{
			name: "create publisher", method: http.MethodPost, path: "/publishers", body: `{"name":"Publisher"}`,
			wantStatus: http.StatusOK, wantJSON: publisherJSON,
			setup: func(_ *bookmocks.MockBooks, p *publishermocks.MockPublishers) {
				p.EXPECT().CreatePublisher(mock.Anything, publisherentities.Publisher{Name: publisher.Name}).Return(publisher, nil)
			},
		},
		{
			name: "update publisher", method: http.MethodPut, path: "/publishers/" + publisherID.String(),
			body: `{"name":"Publisher"}`, wantStatus: http.StatusOK, wantJSON: publisherJSON,
			setup: func(_ *bookmocks.MockBooks, p *publishermocks.MockPublishers) {
				p.EXPECT().UpdatePublisher(mock.Anything, publisher).Return(publisher, nil)
			},
		},
		{
			name: "delete publisher", method: http.MethodDelete, path: "/publishers/" + publisherID.String(),
			wantStatus: http.StatusOK, wantJSON: `{"status":"deleted"}`,
			setup: func(_ *bookmocks.MockBooks, p *publishermocks.MockPublishers) {
				p.EXPECT().DeletePublisher(mock.Anything, publisherID).Return(nil)
			},
		},
		{
			name: "publisher conflict", method: http.MethodPost, path: "/publishers", body: `{"name":"Publisher"}`,
			wantStatus: http.StatusConflict, wantJSON: `{"error":"duplicate or referenced record"}`,
			setup: func(_ *bookmocks.MockBooks, p *publishermocks.MockPublishers) {
				p.EXPECT().CreatePublisher(mock.Anything, publisherentities.Publisher{Name: publisher.Name}).
					Return(publisherentities.Publisher{}, shared.ErrConflict)
			},
		},
		{
			name: "unexpected publisher error", method: http.MethodGet, path: "/publishers/" + publisherID.String(),
			wantStatus: http.StatusInternalServerError, wantJSON: `{"error":"internal server error"}`,
			setup: func(_ *bookmocks.MockBooks, p *publishermocks.MockPublishers) {
				p.EXPECT().GetPublisher(mock.Anything, publisherID).
					Return(publisherentities.Publisher{}, errors.New("private database error"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			books := bookmocks.NewMockBooks(t)
			publishers := publishermocks.NewMockPublishers(t)
			tt.setup(books, publishers)
			api := server.New(Handler{
				Books: bookservice.Service{Repository: books}, Publishers: publisherservice.Service{Repository: publishers},
			})

			req := httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
			if tt.body != "" {
				req.Header.Set("Content-Type", "application/json")
			}

			res := httptest.NewRecorder()
			api.ServeHTTP(res, req)

			require.Equal(t, tt.wantStatus, res.Code)
			require.JSONEq(t, tt.wantJSON, res.Body.String())
		})
	}
}
