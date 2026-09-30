package postgreserrors

import (
	"errors"
	"fmt"

	"bookshop/backend/internal/domain/shared"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func Map(err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, pgx.ErrNoRows) {
		return shared.ErrNotFound
	}

	var pg *pgconn.PgError

	if errors.As(err, &pg) {
		switch pg.Code {
		case "23505", "23503":
			return shared.ErrConflict
		case "23514", "23502", "22P02":
			return shared.ErrInvalid
		}
	}

	return fmt.Errorf("database: %w", err)
}
