package entities

import (
	"fmt"
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
		{name: "missing year", year: 0, wantErr: ErrYearInvalid},
		{name: "year too high", year: 10000, wantErr: ErrYearInvalid},
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
		{"missing author", func(b *Book) { b.Author = " " }, ErrAuthorRequired},
		{"missing title", func(b *Book) { b.Name = "" }, ErrNameRequired},
		{"missing publisher", func(b *Book) { b.PublisherID = uuid.Nil }, ErrPublisherRequired},
		{"missing year", func(b *Book) { b.Year = 0 }, ErrYearInvalid},
		{"year too high", func(b *Book) { b.Year = 10000 }, ErrYearInvalid},
		{"publication before year", func(b *Book) { b.PublicationYear = 1999 }, ErrPublicationYearBeforeYear},
		{"publication too high", func(b *Book) { b.PublicationYear = 10000 }, ErrPublicationYearInvalid},
		{"negative price", func(b *Book) { b.Price = -1 }, ErrPriceNegative},
		{"negative quantity", func(b *Book) { b.Quantity = -1 }, ErrQuantityNegative},
		{"first year", func(b *Book) { b.Year, b.PublicationYear = 1, 1 }, nil},
		{"last year", func(b *Book) { b.Year, b.PublicationYear = 9999, 9999 }, nil},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			b := base
			tt.change(&b)
			err := b.Validate()
			require.ErrorIs(t, err, tt.wantErr)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, shared.ErrInvalid)
			}

			if tt.wantErr == nil {
				require.Equal(t, "Anna", b.Author)
				require.Equal(t, "Story", b.Name)
			}
		})
	}
}

func TestBookValidateStopsAtFirstReason(t *testing.T) {
	b := Book{Author: "  ", Name: "\t", Year: 0, Price: -1, PublicationYear: 10000, Quantity: -2}
	for _, tc := range []struct {
		want error
		fix  func()
	}{
		{ErrAuthorRequired, func() { b.Author = "A" }},
		{ErrNameRequired, func() { b.Name = "B" }},
		{ErrYearInvalid, func() { b.Year = 2000 }},
		{ErrPriceNegative, func() { b.Price = 0 }},
		{ErrPublisherRequired, func() { b.PublisherID = uuid.New() }},
		{ErrPublicationYearInvalid, func() { b.PublicationYear = 1999 }},
		{ErrPublicationYearBeforeYear, func() { b.PublicationYear = 2000 }},
		{ErrQuantityNegative, func() { b.Quantity = 0 }},
	} {
		err := fmt.Errorf("validate book: %w", b.Validate())
		require.ErrorIs(t, err, tc.want)
		require.ErrorIs(t, err, shared.ErrInvalid)
		tc.fix()
	}

	require.NoError(t, b.Validate())
}
