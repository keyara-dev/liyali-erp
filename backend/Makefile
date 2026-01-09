.PHONY: help build run test test-unit test-verbose bench clean deps lint fmt

# Variables
APP_NAME=liyali-gateway
PORT?=8080
GO_VERSION=1.21

help:
	@echo "$(APP_NAME) - Liyali Gateway Backend"
	@echo ""
	@echo "Available targets:"
	@echo "  build              - Build the application"
	@echo "  run                - Run the application"
	@echo "  test               - Run all tests"
	@echo "  test-unit          - Run unit tests only"
	@echo "  test-verbose       - Run tests with verbose output"
	@echo "  bench              - Run benchmarks"
	@echo "  coverage           - Generate coverage report"
	@echo "  deps               - Download dependencies"
	@echo "  lint               - Run linter"
	@echo "  fmt                - Format code"
	@echo "  clean              - Clean build artifacts"
	@echo ""
	@echo "Examples:"
	@echo "  make build"
	@echo "  make run"
	@echo "  make test"
	@echo "  PORT=9000 make run"

deps:
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy
	@echo "Dependencies installed successfully"

build:
	@echo "Building $(APP_NAME)..."
	go build -o bin/$(APP_NAME) main.go
	@echo "Build complete: bin/$(APP_NAME)"

run:
	@echo "Running $(APP_NAME) on port $(PORT)..."
	go run main.go

test:
	@echo "Running all tests..."
	go test -v ./...

test-unit:
	@echo "Running unit tests..."
	go test -v -short ./...

test-verbose:
	@echo "Running tests with verbose output..."
	go test -v -race -count=1 ./...

test-coverage:
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

test-services:
	@echo "Running service tests..."
	go test -v ./services/...

test-handlers:
	@echo "Running handler tests..."
	go test -v ./handlers/...

bench:
	@echo "Running benchmarks..."
	go test -bench=. -benchmem -run=^$ ./...

bench-services:
	@echo "Running service benchmarks..."
	go test -bench=. -benchmem -run=^$ ./services/...

lint:
	@echo "Running linter..."
	golangci-lint run

fmt:
	@echo "Formatting code..."
	go fmt ./...
	goimports -w .

clean:
	@echo "Cleaning up..."
	rm -f bin/$(APP_NAME)
	rm -f coverage.out coverage.html
	go clean
	@echo "Clean complete"

# Database targets
db-migrate:
	@echo "Running database migrations..."
	go run main.go -migrate

db-seed:
	@echo "Seeding database..."
	go run main.go -seed

seed-approvals:
	@echo "Seeding approval test data..."
	go run cmd/seed_approvals.go

seed-approvals-cleanup:
	@echo "Cleaning up approval test data..."
	go run cmd/seed_approvals.go cleanup

# Development targets
watch:
	@echo "Watching for changes..."
	reflex -r '\.go$$' -s make run

dev:
	@echo "Starting development environment..."
	make deps
	make build
	make run

# Docker targets
docker-build:
	@echo "Building Docker image..."
	docker build -t $(APP_NAME):latest .

docker-run:
	@echo "Running Docker container..."
	docker run -p $(PORT):$(PORT) $(APP_NAME):latest

docker-clean:
	@echo "Removing Docker image..."
	docker rmi $(APP_NAME):latest

# CI/CD targets
ci:
	@echo "Running CI checks..."
	make deps
	make fmt
	make lint
	make test
	make build

# Code generation
generate:
	@echo "Running code generation..."
	go generate ./...

# Version info
version:
	@echo "Go version: $$(go version)"
	@echo "App: $(APP_NAME)"
	@echo "Port: $(PORT)"

.PHONY: help build run test test-unit test-verbose test-coverage test-services test-handlers \
        bench bench-services lint fmt clean deps db-migrate db-seed watch dev version \
        docker-build docker-run docker-clean ci generate
