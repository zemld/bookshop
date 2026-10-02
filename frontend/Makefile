.PHONY: gen format
gen:
	go run github.com/a-h/templ/cmd/templ@v0.3.960 generate

format: gen
	golangci-lint fmt
	golangci-lint run --fix ./...
