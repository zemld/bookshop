package postgres

import (
	"context"
	"fmt"

	"bookshop/backend/internal/domain/books/entities"
	"bookshop/backend/internal/domain/shared"
	"bookshop/backend/internal/ports/books"
	"bookshop/backend/internal/utils/postgreserrors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct{ DB *pgxpool.Pool }

var _ books.Repository = Store{}

const bookColumns = `id, author, name, year, price, publisher_id, publication_year, quantity`

func scanBook(row pgx.Row) (entities.Book, error) {
	var b entities.Book

	err := row.Scan(&b.ID, &b.Author, &b.Name, &b.Year, &b.Price, &b.PublisherID, &b.PublicationYear, &b.Quantity)
	if err != nil {
		return b, fmt.Errorf("scan book: %w", postgreserrors.MapDatabaseError(err))
	}

	return b, nil
}

func (s Store) ListBooks(ctx context.Context) ([]entities.Book, error) {
	rows, err := s.DB.Query(ctx, `SELECT `+bookColumns+` FROM books ORDER BY name,id`)
	if err != nil {
		return nil, fmt.Errorf("query books: %w", postgreserrors.MapDatabaseError(err))
	}
	defer rows.Close()

	items := make([]entities.Book, 0)

	for rows.Next() {
		b, err := scanBook(rows)
		if err != nil {
			return nil, err
		}

		items = append(items, b)
	}

	if err := rows.Err(); err != nil {
		return items, fmt.Errorf("iterate books: %w", postgreserrors.MapDatabaseError(err))
	}

	return items, nil
}

func (s Store) GetBook(ctx context.Context, id uuid.UUID) (entities.Book, error) {
	return scanBook(s.DB.QueryRow(ctx, `SELECT `+bookColumns+` FROM books WHERE id=$1`, id))
}

func (s Store) CreateBook(ctx context.Context, b entities.Book) (entities.Book, error) {
	b.ID = uuid.New()

	return scanBook(s.DB.QueryRow(ctx, `INSERT INTO books (`+bookColumns+`)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8) RETURNING `+bookColumns,
		b.ID, b.Author, b.Name, b.Year, b.Price, b.PublisherID, b.PublicationYear, b.Quantity))
}

func (s Store) UpdateBook(ctx context.Context, b entities.Book) (entities.Book, error) {
	return scanBook(s.DB.QueryRow(ctx, `UPDATE books SET author=$2,name=$3,year=$4,price=$5,
		publisher_id=$6,publication_year=$7,quantity=$8 WHERE id=$1 RETURNING `+bookColumns,
		b.ID, b.Author, b.Name, b.Year, b.Price, b.PublisherID, b.PublicationYear, b.Quantity))
}

func (s Store) DeleteBook(ctx context.Context, id uuid.UUID) error {
	tag, err := s.DB.Exec(ctx, `DELETE FROM books WHERE id=$1`, id)
	if err != nil {
		return fmt.Errorf("delete book: %w", postgreserrors.MapDatabaseError(err))
	}

	if tag.RowsAffected() == 0 {
		return shared.ErrNotFound
	}

	return nil
}
