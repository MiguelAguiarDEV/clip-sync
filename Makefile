GO ?= go

.PHONY: test run build lint

test:
	$(GO) test ./... -v

run:
	$(GO) -C server run ./cmd/server

build:
	$(GO) -C server build -o ../bin/server ./cmd/server
	$(GO) -C clients/cli build -o ../../bin/cli .

lint:
	$(GO) vet ./...
