package app

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	bookstore "bookshop/backend/internal/adapters/books/postgres"
	publisherstore "bookshop/backend/internal/adapters/publishers/postgres"
	"bookshop/backend/internal/api/rest/handler"
	"bookshop/backend/internal/api/rest/ogen"
	"bookshop/backend/internal/api/rest/server"
	"bookshop/backend/internal/domain/books"
	"bookshop/backend/internal/domain/publishers"
	bookrepository "bookshop/backend/internal/ports/books"
	publisherrepository "bookshop/backend/internal/ports/publishers"
	bookservice "bookshop/backend/internal/services/books"
	publisherservice "bookshop/backend/internal/services/publishers"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
)

const (
	_startupTimeout = 10 * time.Second
	_stopTimeout    = 12 * time.Second
)

func New() *fx.App {
	return fx.New(
		fx.StopTimeout(_stopTimeout),
		fx.WithLogger(func() fxevent.Logger {
			return &fxevent.SlogLogger{Logger: slog.Default()}
		}),
		fx.Provide(
			provideDatabase,
			loadServerConfig,
			fx.Annotate(func(db *pgxpool.Pool) *bookstore.Store {
				return &bookstore.Store{DB: db}
			}, fx.As(new(bookrepository.Repository))),
			fx.Annotate(func(db *pgxpool.Pool) *publisherstore.Store {
				return &publisherstore.Store{DB: db}
			}, fx.As(new(publisherrepository.Repository))),
			fx.Annotate(func(repo bookrepository.Repository) *bookservice.Service {
				return &bookservice.Service{Repository: repo}
			}, fx.As(new(books.Books))),
			fx.Annotate(func(repo publisherrepository.Repository) *publisherservice.Service {
				return &publisherservice.Service{Repository: repo}
			}, fx.As(new(publishers.Publishers))),
			fx.Annotate(func(books books.Books, publishers publishers.Publishers) handler.Handler {
				return handler.Handler{Books: books, Publishers: publishers}
			}, fx.As(new(ogen.Handler))),
			server.NewServer,
		),
		fx.Invoke(func(*server.Server) {}),
	)
}

func loadServerConfig() server.Config {
	addr := os.Getenv("API_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	return server.Config{Address: addr}
}

func provideDatabase(lifecycle fx.Lifecycle) (*pgxpool.Pool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), _startupTimeout)
	defer cancel()

	db, err := pgxpool.New(ctx, os.Getenv("DATABASE_URL"))
	if err == nil {
		err = db.Ping(ctx)
	}

	if err != nil {
		if db != nil {
			db.Close()
		}

		return nil, fmt.Errorf("connect to database: %w", err)
	}

	lifecycle.Append(fx.Hook{OnStop: func(context.Context) error {
		db.Close()

		return nil
	}})

	return db, nil
}
