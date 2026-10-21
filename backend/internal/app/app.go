package app

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	bookstore "bookshop/backend/internal/adapters/books/postgres"
	publisherstore "bookshop/backend/internal/adapters/publishers/postgres"
	"bookshop/backend/internal/api/rest/handler"
	"bookshop/backend/internal/api/rest/ogen"
	"bookshop/backend/internal/api/rest/server"
	"bookshop/backend/internal/config"
	"bookshop/backend/internal/domain/books"
	"bookshop/backend/internal/domain/publishers"
	bookrepository "bookshop/backend/internal/ports/books"
	publisherrepository "bookshop/backend/internal/ports/publishers"
	bookservice "bookshop/backend/internal/services/books"
	"bookshop/backend/internal/services/metrics"
	publisherservice "bookshop/backend/internal/services/publishers"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel/metric"
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
		fx.WithLogger(func(logger *slog.Logger) fxevent.Logger {
			return &fxevent.SlogLogger{Logger: logger}
		}),
		fx.Provide(
			provideDatabase,
			provideMetrics,
			func() (config.Config, error) { return config.Load("config/config.yaml") },
			func(cfg config.Config) config.ServiceConfig { return cfg.Service },
			func(cfg config.ServiceConfig) *slog.Logger {
				return slog.New(slog.NewJSONHandler(os.Stdout, nil)).With("service", cfg.Name)
			},
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
			func(service *metrics.Service) metric.MeterProvider { return service.Provider },
			func(service *metrics.Service) http.Handler { return handler.CreateMetricsHandler(service.Registry) },
			server.NewServer,
		),
		fx.Invoke(func(*server.Server) {}),
	)
}

func provideMetrics(lifecycle fx.Lifecycle) (*metrics.Service, error) {
	service, err := metrics.New()
	if err != nil {
		return nil, fmt.Errorf("initialize metrics: %w", err)
	}

	lifecycle.Append(fx.Hook{OnStop: service.Provider.Shutdown})

	return service, nil
}

func loadServerConfig(cfg config.Config) server.Config {
	return server.Config{Address: cfg.Server.Address, MetricsAddress: cfg.Metrics.Address}
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
