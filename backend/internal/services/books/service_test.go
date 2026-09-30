package books

import (
	"context"
	"errors"
	"testing"

	"bookshop/backend/internal/adapters/books/mocks"
	"bookshop/backend/internal/domain/books/entities"
	"bookshop/backend/internal/domain/shared"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCreateBook(t *testing.T) {
	t.Parallel()

	valid := entities.Book{Author: " A ", Name: " B ", Year: 1, PublicationYear: 1, PublisherID: uuid.New()}
	normalized := valid
	normalized.Author, normalized.Name = "A", "B"
	missingPublisher := valid
	missingPublisher.PublisherID = uuid.Nil
	missingAuthor := valid
	missingAuthor.Author = " "
	unavailable := errors.New("unavailable")

	tests := []struct {
		name             string
		input, want      entities.Book
		repoErr, wantErr error
	}{
		{name: "normalizes input", input: valid, want: normalized},
		{name: "rejects missing publisher", input: missingPublisher, want: entities.Book{Author: "A", Name: "B", Year: 1, PublicationYear: 1}, wantErr: shared.ErrInvalid},
		{name: "rejects blank author", input: missingAuthor, want: entities.Book{Author: "", Name: "B", Year: 1, PublicationYear: 1, PublisherID: valid.PublisherID}, wantErr: shared.ErrInvalid},
		{name: "returns conflict", input: valid, repoErr: shared.ErrConflict, wantErr: shared.ErrConflict},
		{name: "returns repository error", input: valid, repoErr: unavailable, wantErr: unavailable},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			repo := mocks.NewMockBooks(t)
			if tt.wantErr != shared.ErrInvalid {
				repo.EXPECT().CreateBook(mock.Anything, normalized).Return(tt.want, tt.repoErr).Once()
			}

			s := Service{Repository: repo}

			// Act
			got, err := s.CreateBook(context.Background(), tt.input)

			// Assert
			if tt.wantErr == nil {
				require.NoError(t, err)
			} else {
				require.ErrorIs(t, err, tt.wantErr)
			}

			require.Equal(t, tt.want, got)

			if tt.wantErr == shared.ErrInvalid {
				repo.AssertNotCalled(t, "CreateBook", mock.Anything, mock.Anything)
			}
		})
	}
}

func TestUpdateBook(t *testing.T) {
	t.Parallel()

	valid := entities.Book{ID: uuid.New(), Author: " A ", Name: " B ", Year: 1, PublicationYear: 1, PublisherID: uuid.New()}
	normalized := valid
	normalized.Author, normalized.Name = "A", "B"
	missingPublisher := valid
	missingPublisher.PublisherID = uuid.Nil
	invalidYear := valid
	invalidYear.Year = 0
	unavailable := errors.New("unavailable")

	tests := []struct {
		name             string
		input, want      entities.Book
		repoErr, wantErr error
	}{
		{name: "normalizes input", input: valid, want: normalized},
		{name: "rejects missing publisher", input: missingPublisher, want: entities.Book{ID: valid.ID, Author: "A", Name: "B", Year: 1, PublicationYear: 1}, wantErr: shared.ErrInvalid},
		{name: "rejects invalid year", input: invalidYear, want: entities.Book{ID: valid.ID, Author: "A", Name: "B", PublicationYear: 1, PublisherID: valid.PublisherID}, wantErr: shared.ErrInvalid},
		{name: "returns missing", input: valid, repoErr: shared.ErrNotFound, wantErr: shared.ErrNotFound},
		{name: "returns conflict", input: valid, repoErr: shared.ErrConflict, wantErr: shared.ErrConflict},
		{name: "returns repository error", input: valid, repoErr: unavailable, wantErr: unavailable},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			repo := mocks.NewMockBooks(t)
			if tt.wantErr != shared.ErrInvalid {
				repo.EXPECT().UpdateBook(mock.Anything, normalized).Return(tt.want, tt.repoErr).Once()
			}

			s := Service{Repository: repo}

			// Act
			got, err := s.UpdateBook(context.Background(), tt.input)

			// Assert
			if tt.wantErr == nil {
				require.NoError(t, err)
			} else {
				require.ErrorIs(t, err, tt.wantErr)
			}

			require.Equal(t, tt.want, got)

			if tt.wantErr == shared.ErrInvalid {
				repo.AssertNotCalled(t, "UpdateBook", mock.Anything, mock.Anything)
			}
		})
	}
}

func TestListBooks(t *testing.T) {
	t.Parallel()

	books := []entities.Book{{ID: uuid.New(), Name: "B"}}
	unavailable := errors.New("unavailable")

	tests := []struct {
		name    string
		want    []entities.Book
		wantErr error
	}{
		{name: "returns books", want: books},
		{name: "returns empty list", want: []entities.Book{}},
		{name: "returns repository error", wantErr: unavailable},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			repo := mocks.NewMockBooks(t)
			repo.EXPECT().ListBooks(mock.Anything).Return(tt.want, tt.wantErr).Once()
			s := Service{Repository: repo}

			// Act
			got, err := s.ListBooks(context.Background())

			// Assert
			if tt.wantErr == nil {
				require.NoError(t, err)
			} else {
				require.ErrorIs(t, err, tt.wantErr)
			}

			require.Equal(t, tt.want, got)
		})
	}
}

func TestGetBook(t *testing.T) {
	t.Parallel()

	id := uuid.New()
	book := entities.Book{ID: id, Name: "B"}
	unavailable := errors.New("unavailable")

	tests := []struct {
		name    string
		want    entities.Book
		wantErr error
	}{
		{name: "returns book", want: book},
		{name: "returns missing", wantErr: shared.ErrNotFound},
		{name: "returns repository error", wantErr: unavailable},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			repo := mocks.NewMockBooks(t)
			repo.EXPECT().GetBook(mock.Anything, id).Return(tt.want, tt.wantErr).Once()
			s := Service{Repository: repo}

			// Act
			got, err := s.GetBook(context.Background(), id)

			// Assert
			if tt.wantErr == nil {
				require.NoError(t, err)
			} else {
				require.ErrorIs(t, err, tt.wantErr)
			}

			require.Equal(t, tt.want, got)
		})
	}
}

func TestDeleteBook(t *testing.T) {
	t.Parallel()

	id := uuid.New()
	unavailable := errors.New("unavailable")

	tests := []struct {
		name    string
		wantErr error
	}{
		{name: "deletes book"},
		{name: "returns missing", wantErr: shared.ErrNotFound},
		{name: "returns conflict", wantErr: shared.ErrConflict},
		{name: "returns repository error", wantErr: unavailable},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			repo := mocks.NewMockBooks(t)
			repo.EXPECT().DeleteBook(mock.Anything, id).Return(tt.wantErr).Once()
			s := Service{Repository: repo}

			// Act
			err := s.DeleteBook(context.Background(), id)

			// Assert
			if tt.wantErr == nil {
				require.NoError(t, err)
			} else {
				require.ErrorIs(t, err, tt.wantErr)
			}
		})
	}
}
