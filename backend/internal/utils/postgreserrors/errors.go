package postgreserrors

import (
	"errors"
	"fmt"

	bookentities "bookshop/backend/internal/domain/books/entities"
	publisherentities "bookshop/backend/internal/domain/publishers/entities"
	"bookshop/backend/internal/domain/shared"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func MapDatabaseError(err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, pgx.ErrNoRows) {
		return shared.ErrNotFound
	}

	var pg *pgconn.PgError

	if errors.As(err, &pg) {
		switch pg.Code {
		case "23505":
			switch pg.ConstraintName {
			case "publishers_name_unique", "publishers_normalized_name_unique":
				return publisherentities.ErrNameDuplicate
			case "books_content_unique":
				return bookentities.ErrDuplicate
			}

			return shared.ErrConflict
		case "23503":
			if pg.ConstraintName == "books_publisher_id_fkey" {
				return bookentities.ErrPublisherNotFound
			}

			return shared.ErrConflict
		case "23514", "23502", "22P02":
			return shared.ErrInvalid
		}
	}

	return fmt.Errorf("database: %w", err)
}
