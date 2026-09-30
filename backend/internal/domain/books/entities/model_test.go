package entities

import (
	"testing"

	"bookshop/backend/internal/domain/shared"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestYearValidate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		year    Year
		wantErr error
	}{
		{name: "first valid year", year: 1},
		{name: "last valid year", year: 9999},
		{name: "missing year", year: 0, wantErr: shared.ErrInvalid},
		{name: "year too high", year: 10000, wantErr: shared.ErrInvalid},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.year.Validate()

			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestBookValidate(t *testing.T) {
	t.Parallel()

	base := Book{Author: " Anna ", Name: " Story ", Year: 2000, PublicationYear: 2001, PublisherID: uuid.New()}

	for _, tt := range []struct {
		name    string
		change  func(*Book)
		wantErr error
	}{
		{"normalized", func(*Book) {}, nil},
		{"missing author", func(b *Book) { b.Author = " " }, shared.ErrInvalid},
		{"missing title", func(b *Book) { b.Name = "" }, shared.ErrInvalid},
		{"missing publisher", func(b *Book) { b.PublisherID = uuid.Nil }, shared.ErrInvalid},
		{"missing year", func(b *Book) { b.Year = 0 }, shared.ErrInvalid},
		{"year too high", func(b *Book) { b.Year = 10000 }, shared.ErrInvalid},
		{"publication before year", func(b *Book) { b.PublicationYear = 1999 }, shared.ErrInvalid},
		{"publication too high", func(b *Book) { b.PublicationYear = 10000 }, shared.ErrInvalid},
		{"negative price", func(b *Book) { b.Price = -1 }, shared.ErrInvalid},
		{"negative quantity", func(b *Book) { b.Quantity = -1 }, shared.ErrInvalid},
		{"first year", func(b *Book) { b.Year, b.PublicationYear = 1, 1 }, nil},
		{"last year", func(b *Book) { b.Year, b.PublicationYear = 9999, 9999 }, nil},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			b := base
			tt.change(&b)
			err := b.Validate()
			require.ErrorIs(t, err, tt.wantErr)

			if tt.wantErr == nil {
				require.Equal(t, "Anna", b.Author)
				require.Equal(t, "Story", b.Name)
			}
		})
	}
}
