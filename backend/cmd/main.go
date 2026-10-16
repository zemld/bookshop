package main

import (
	"bookshop/backend/internal/app"
)

//go:generate go run github.com/ogen-go/ogen/cmd/ogen --target ../internal/api/rest/ogen --package ogen --clean ../api/openapi.yaml

func main() {
	app.New().Run()
}
