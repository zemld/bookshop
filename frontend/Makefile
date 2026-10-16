.PHONY: gen format
gen:
	go run github.com/ogen-go/ogen/cmd/ogen@v1.14.0 --target internal/adapters/api/ogen --package ogen --clean ../backend/api/openapi.yaml
	go run github.com/a-h/templ/cmd/templ@v0.3.960 generate

format: gen
	golangci-lint fmt
	golangci-lint run --fix ./...
