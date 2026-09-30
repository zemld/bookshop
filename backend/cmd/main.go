package main

import (
	"log/slog"
	"os"

	"bookshop/backend/internal/app"
)

//go:generate go run github.com/ogen-go/ogen/cmd/ogen --target ../internal/api/rest/ogen --package ogen --clean ../api/openapi.yaml

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	app.New().Run()
}
