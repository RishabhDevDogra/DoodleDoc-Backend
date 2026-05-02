APP_NAME := doodledoc-backend
SWAG := $(shell go env GOPATH)/bin/swag
AIR := $(shell go env GOPATH)/bin/air

.PHONY: run dev build test tidy docs help

help:
	@echo "Available commands:"
	@echo "  make run   - Run the API server"
	@echo "  make dev   - Run with auto-reload (Air)"
	@echo "  make build - Build all packages"
	@echo "  make test  - Run all tests"
	@echo "  make tidy  - Clean and sync module deps"
	@echo "  make docs  - Regenerate Swagger docs"

run:
	go run .

dev:
	@if [ ! -x "$(AIR)" ]; then \
		echo "Installing Air for auto-reload..."; \
		go install github.com/air-verse/air@latest; \
	fi
	$(AIR)

build:
	go build ./...

test:
	go test ./...

tidy:
	go mod tidy

docs:
	@if [ ! -x "$(SWAG)" ]; then \
		echo "Installing swag CLI..."; \
		go install github.com/swaggo/swag/cmd/swag@latest; \
	fi
	$(SWAG) init -g main.go
