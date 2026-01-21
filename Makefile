.PHONY: build run test fmt lint tidy

BINARY ?= kxl-api

build:
	go build -o $(BINARY) ./cmd/api

run:
	go run ./cmd/api

test:
	go test ./...

fmt:
	go fmt ./...

lint:
	go vet ./...

tidy:
	go mod tidy

