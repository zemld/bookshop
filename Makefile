.PHONY: format gen test build
format:
	$(MAKE) -C backend format
	$(MAKE) -C frontend format

gen:
	$(MAKE) -C backend gen
	$(MAKE) -C frontend gen

test:
	cd backend && go test ./...
	cd frontend && go test ./...

build:
	cd backend && go build -buildvcs=false ./...
	cd frontend && go build -buildvcs=false ./...
