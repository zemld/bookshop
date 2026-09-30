.PHONY: format gen

format:
	golangci-lint run --fix ./...

gen:
	go generate ./...
	go run github.com/vektra/mockery/v3@v3.8.0
