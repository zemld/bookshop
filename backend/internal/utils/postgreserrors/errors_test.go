package postgreserrors

import (
	"fmt"
	"testing"

	bookentities "bookshop/backend/internal/domain/books/entities"
	publisherentities "bookshop/backend/internal/domain/publishers/entities"
	"bookshop/backend/internal/domain/shared"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/require"
)

func TestKnownConstraints(t *testing.T) {
	for _, tc := range []struct {
		code, constraint string
		reason           error
	}{
		{"23505", "publishers_name_unique", publisherentities.ErrNameDuplicate},
		{"23505", "publishers_normalized_name_unique", publisherentities.ErrNameDuplicate},
		{"23505", "books_content_unique", bookentities.ErrDuplicate},
		{"23503", "books_publisher_id_fkey", bookentities.ErrPublisherNotFound},
	} {
		err := fmt.Errorf("write: %w", MapDatabaseError(&pgconn.PgError{Code: tc.code, ConstraintName: tc.constraint, Detail: "private SQL"}))
		require.ErrorIs(t, err, shared.ErrConflict)
		require.ErrorIs(t, err, tc.reason)
		require.NotContains(t, err.Error(), "private SQL")
	}
}
