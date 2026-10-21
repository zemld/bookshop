package app

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"bookshop/frontend/internal/adapters/api"
	"bookshop/frontend/internal/api/web"
	"bookshop/frontend/internal/config"
	bookports "bookshop/frontend/internal/ports/books"
	publisherports "bookshop/frontend/internal/ports/publishers"
	bookservice "bookshop/frontend/internal/services/books"
	"bookshop/frontend/internal/services/metrics"
	publisherservice "bookshop/frontend/internal/services/publishers"

	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
)

const _stopTimeout = 12 * time.Second

func New() *fx.App {
	return fx.New(
		fx.StopTimeout(_stopTimeout),
		fx.WithLogger(func(logger *slog.Logger) fxevent.Logger {
			return &fxevent.SlogLogger{Logger: logger}
		}),
		fx.Provide(
			provideMetrics,
			func() (config.Config, error) { return config.Load("config/config.yaml") },
			func(cfg config.Config) config.ServiceConfig { return cfg.Service },
			func(cfg config.ServiceConfig) *slog.Logger {
				return slog.New(slog.NewJSONHandler(os.Stdout, nil)).With("service", cfg.Name)
			},
			func(cfg config.Config, metrics *metrics.Service) (*api.Client, error) {
				return api.New(cfg.API.URL, metrics.Provider)
			},
			func(c *api.Client) bookports.Books { return c },
			func(c *api.Client) publisherports.Publishers { return c },
			func(b bookports.Books, p publisherports.Publishers) *bookservice.Service {
				return &bookservice.Service{Books: b, Publishers: p}
			},
			func(p publisherports.Publishers) *publisherservice.Service {
				return &publisherservice.Service{Publishers: p}
			},
			func(b bookports.Books, p publisherports.Publishers, bs *bookservice.Service, ps *publisherservice.Service, logger *slog.Logger) web.Server {
				return web.Server{
					Books: b, DeleteBook: b, Publishers: p, GetPublisher: p, DeletePublisher: p,
					LoadBookForm: bs, CreateBook: b, UpdateBook: b,
					CreatePublisher: ps, UpdatePublisher: ps,
					Logger: logger,
				}
			},
			func(cfg config.Config, handler web.Server, logger *slog.Logger, metrics *metrics.Service) *web.HTTPServer {
				server := web.NewHTTPServer(cfg.Server.Address, cfg.Metrics.Address, handler.CreateHandler(), web.CreateMetricsHandler(metrics.Registry))
				server.SetLogger(logger)
				return server
			},
		),
		fx.Invoke(func(lc fx.Lifecycle, server *web.HTTPServer, client *api.Client) {
			lc.Append(fx.Hook{OnStop: func(context.Context) error { client.Stop(); return nil }})
			lc.Append(fx.Hook{OnStart: server.Start, OnStop: server.Stop})
		}),
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

func Run() error {
	application := New()
	if err := application.Err(); err != nil {
		return err
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := application.Start(ctx); err != nil {
		return err
	}
	<-ctx.Done()
	drain, cancel := context.WithTimeout(context.Background(), _stopTimeout)
	defer cancel()
	return application.Stop(drain)
}
