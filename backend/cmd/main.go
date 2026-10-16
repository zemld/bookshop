package main

import (
	"log/slog"
	"os"

	"bookshop/backend/internal/app"
	"bookshop/backend/internal/requestlog"
)

//go:generate go run github.com/ogen-go/ogen/cmd/ogen --target ../internal/api/rest/ogen --package ogen --clean ../api/openapi.yaml

func main() {
	slog.SetDefault(slog.New(requestlog.NewHandler(slog.NewJSONHandler(os.Stdout, nil).WithAttrs([]slog.Attr{slog.String("service", "bookshop-api")}))))
	app.New().Run()
}
