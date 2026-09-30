package postgres

import (
	"context"
	"fmt"

	"bookshop/backend/internal/domain/publishers/entities"
	"bookshop/backend/internal/domain/shared"
	"bookshop/backend/internal/ports/publishers"
	"bookshop/backend/internal/utils/postgreserrors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct{ DB *pgxpool.Pool }

var _ publishers.Repository = Store{}

func (s Store) ListPublishers(ctx context.Context) ([]entities.Publisher, error) {
	rows, err := s.DB.Query(ctx, `SELECT id, name FROM publishers ORDER BY name, id`)
	if err != nil {
		return nil, fmt.Errorf("query publishers: %w", postgreserrors.Map(err))
	}
	defer rows.Close()

	items := make([]entities.Publisher, 0)

	for rows.Next() {
		var p entities.Publisher
		if err := rows.Scan(&p.ID, &p.Name); err != nil {
			return nil, fmt.Errorf("scan publisher: %w", postgreserrors.Map(err))
		}

		items = append(items, p)
	}

	if err := rows.Err(); err != nil {
		return items, fmt.Errorf("iterate publishers: %w", postgreserrors.Map(err))
	}

	return items, nil
}

func (s Store) GetPublisher(ctx context.Context, id uuid.UUID) (entities.Publisher, error) {
	var p entities.Publisher

	err := s.DB.QueryRow(ctx, `SELECT id, name FROM publishers WHERE id=$1`, id).Scan(&p.ID, &p.Name)
	if err != nil {
		return p, fmt.Errorf("get publisher: %w", postgreserrors.Map(err))
	}

	return p, nil
}

func (s Store) CreatePublisher(ctx context.Context, p entities.Publisher) (entities.Publisher, error) {
	p.ID = uuid.New()

	err := s.DB.QueryRow(ctx, `INSERT INTO publishers (id,name) VALUES ($1,$2) RETURNING id`, p.ID, p.Name).Scan(&p.ID)
	if err != nil {
		return p, fmt.Errorf("create publisher: %w", postgreserrors.Map(err))
	}

	return p, nil
}

func (s Store) UpdatePublisher(ctx context.Context, p entities.Publisher) (entities.Publisher, error) {
	var id uuid.UUID

	err := s.DB.QueryRow(ctx, `UPDATE publishers SET name=$2 WHERE id=$1 RETURNING id`, p.ID, p.Name).Scan(&id)
	if err != nil {
		return p, fmt.Errorf("update publisher: %w", postgreserrors.Map(err))
	}

	return p, nil
}

func (s Store) DeletePublisher(ctx context.Context, id uuid.UUID) error {
	tag, err := s.DB.Exec(ctx, `DELETE FROM publishers WHERE id=$1`, id)
	if err != nil {
		return fmt.Errorf("delete publisher: %w", postgreserrors.Map(err))
	}

	if tag.RowsAffected() == 0 {
		return shared.ErrNotFound
	}

	return nil
}
