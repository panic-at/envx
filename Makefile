BINARY  := envx
PKG     := github.com/SEU_USER/envx
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -X $(PKG)/internal/version.Version=$(VERSION)

.PHONY: build test test-race lint cover fmt vet clean

## build: compile the envx binary into ./bin
build:
	go build -ldflags "$(LDFLAGS)" -o bin/$(BINARY) ./cmd/envx

## test: run the test suite
test:
	go test ./...

## test-race: run the test suite with the race detector
test-race:
	go test ./... -race

## lint: check formatting and run go vet
lint: fmt vet

## fmt: fail if any file is not gofmt-formatted
fmt:
	@unformatted=$$(gofmt -l .); \
	if [ -n "$$unformatted" ]; then \
		echo "gofmt needed on:"; echo "$$unformatted"; exit 1; \
	fi

## vet: run go vet across all packages
vet:
	go vet ./...

## cover: run tests with coverage and print a summary
cover:
	go test ./... -coverprofile=coverage.out
	go tool cover -func=coverage.out | tail -n 1

## clean: remove build and coverage artifacts
clean:
	rm -rf bin coverage.out
