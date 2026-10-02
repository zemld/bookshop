package main

import (
	"log/slog"
	"os"

	"bookshop/frontend/internal/app"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	if err := app.Run(); err != nil {
		slog.Error("web stopped", "error", err)
		os.Exit(1)
	}
}
