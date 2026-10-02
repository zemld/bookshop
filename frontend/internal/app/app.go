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
		fx.WithLogger(func() fxevent.Logger {
			return &fxevent.SlogLogger{Logger: slog.Default()}
		}),
		fx.Provide(
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
			func(b bookports.Books, p publisherports.Publishers, bs *bookservice.Service, ps *publisherservice.Service) web.Server {
				return web.Server{
					Books: b, DeleteBook: b, Publishers: p, GetPublisher: p, DeletePublisher: p,
					LoadBookForm: bs, CreateBook: b, UpdateBook: b,
					CreatePublisher: ps, UpdatePublisher: ps,
				}
			},
			func(handler web.Server) *web.HTTPServer {
				addr := os.Getenv("WEB_ADDR")
				if addr == "" {
					addr = ":8081"
				}
				return web.NewHTTPServer(addr, handler.Routes())
			},
		),
		fx.Invoke(func(lc fx.Lifecycle, server *web.HTTPServer, client *core.Client) {
			lc.Append(fx.Hook{OnStop: func(context.Context) error { client.Stop(); return nil }})
			// Hooks stop in reverse order: drain the server before closing transport.
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
