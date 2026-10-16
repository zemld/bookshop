package app

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	bookhttp "bookshop/frontend/internal/adapters/books/http"
	core "bookshop/frontend/internal/adapters/httpcore"
	publisherhttp "bookshop/frontend/internal/adapters/publishers/http"
	"bookshop/frontend/internal/api/web"
	"bookshop/frontend/internal/config"
	bookports "bookshop/frontend/internal/ports/books"
	publisherports "bookshop/frontend/internal/ports/publishers"
	bookservice "bookshop/frontend/internal/services/books"
	publisherservice "bookshop/frontend/internal/services/publishers"

	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
)

const stopTimeout = 12 * time.Second

func New() *fx.App {
	return fx.New(
		fx.StopTimeout(stopTimeout),
		fx.WithLogger(func(logger *slog.Logger) fxevent.Logger {
			return &fxevent.SlogLogger{Logger: logger}
		}),
		fx.Provide(
			func() config.ServiceConfig { return config.ServiceConfig{Name: "bookshop-web"} },
			func(cfg config.ServiceConfig) *slog.Logger {
				return slog.New(slog.NewJSONHandler(os.Stdout, nil)).With("service", cfg.Name)
			},
			func() (*core.Client, error) {
				endpoint := os.Getenv("API_URL")
				if endpoint == "" {
					return nil, errors.New("API_URL is required")
				}
				return core.New(endpoint), nil
			},
			func(c *core.Client) *bookhttp.Client { return &bookhttp.Client{Core: c} },
			func(c *core.Client) *publisherhttp.Client { return &publisherhttp.Client{Core: c} },
			func(c *bookhttp.Client) bookports.Books { return c },
			func(c *publisherhttp.Client) publisherports.Publishers { return c },
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
			func(handler web.Server, logger *slog.Logger) *web.HTTPServer {
				addr := os.Getenv("WEB_ADDR")
				if addr == "" {
					addr = ":8081"
				}
				server := web.NewHTTPServer(addr, handler.CreateHandler())
				server.SetLogger(logger)
				return server
			},
		),
		fx.Invoke(func(lc fx.Lifecycle, server *web.HTTPServer, client *core.Client) {
			lc.Append(fx.Hook{OnStop: func(context.Context) error { client.Stop(); return nil }})
			lc.Append(fx.Hook{OnStart: server.Start, OnStop: server.Stop})
		}),
	)
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
	drain, cancel := context.WithTimeout(context.Background(), stopTimeout)
	defer cancel()
	return application.Stop(drain)
}
