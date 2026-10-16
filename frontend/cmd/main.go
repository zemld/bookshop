package main

import (
	"log/slog"
	"os"

	"bookshop/frontend/internal/app"
	"bookshop/frontend/internal/requestlog"
)

func main() {
	slog.SetDefault(slog.New(requestlog.NewHandler(slog.NewJSONHandler(os.Stdout, nil).WithAttrs([]slog.Attr{slog.String("service", "bookshop-web")}))))
	if err := app.Run(); err != nil {
		slog.Error("web stopped", "error", err)
		os.Exit(1)
	}
}
