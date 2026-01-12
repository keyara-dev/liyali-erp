.PHONY: help build run test test-unit test-integration test-verbose bench clean deps lint fmt

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
	@echo "  test-integration   - Run integration tests only"
	@echo "  test-verbose       - Run tests with verbose output"
	@echo "  test-comprehensive - Run comprehensive API test suite (47 endpoints)"
	@echo "  bench              - Run benchmarks"
	@echo "  coverage           - Generate coverage report"
	@echo "  deps               - Download dependencies"
	@echo "  lint               - Run linter"
	@echo "  fmt                - Format code"
	@echo "  clean              - Clean build artifacts"
	@echo "  clean-all          - Deep clean (including caches)"
	@echo "  db-migrate         - Run database migrations"
	@echo "  db-reset           - Reset database with fresh migrations"
	@echo "  dev                - Start development environment"
	@echo "  dev-fresh          - Start fresh development environment"
	@echo "  verify-separation  - Verify multi-tenant data separation"
	@echo "  ci                 - Run full CI checks"
	@echo "  ci-quick           - Run quick CI checks"
	@echo ""
	@echo "Examples:"
	@echo "  make build"
	@echo "  make run"
	@echo "  make test"
	@echo "  make test-comprehensive"
	@echo "  make dev-fresh"
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
	go test -v ./tests/...

test-unit:
	@echo "Running unit tests..."
	go test -v ./tests/unit/...

test-integration:
	@echo "Running integration tests..."
	go test -v ./tests/integration/...

test-verbose:
	@echo "Running tests with verbose output..."
	go test -v -race -count=1 ./tests/...

test-comprehensive:
	@echo "Running comprehensive API test suite..."
	@chmod +x scripts/run_tests.sh
	@./scripts/run_tests.sh

test-coverage:
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./tests/...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

test-services:
	@echo "Running service tests..."
	go test -v ./tests/unit/*service*

test-handlers:
	@echo "Running handler tests..."
	go test -v ./tests/unit/*handler*

bench:
	@echo "Running benchmarks..."
	go test -bench=. -benchmem -run=^$ ./tests/...

bench-services:
	@echo "Running service benchmarks..."
	go test -bench=. -benchmem -run=^$ ./tests/unit/*service*

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
	rm -f *.test
	go clean
	@echo "Clean complete"

clean-all:
	@echo "Deep cleaning..."
	make clean
	go clean -cache
	go clean -modcache
	@echo "Deep clean complete"

# Database targets
db-migrate:
	@echo "Running database migrations..."
	go run database/simple_migration.go

db-reset:
	@echo "Resetting database with fresh migrations..."
	go run database/simple_migration.go database/migrations/000_drop_all_tables.up.sql
	go run database/simple_migration.go database/migrations/001_consolidated_complete_schema.up.sql
	go run database/simple_migration.go database/migrations/002_consolidated_seed_data.up.sql

verify-separation:
	@echo "Verifying multi-tenant data separation..."
	go run cmd/verify-separation/main.go

# Development targets
watch:
	@echo "Watching for changes..."
	reflex -r '\.go$$' -s make run

dev:
	@echo "Starting development environment..."
	make deps
	make db-migrate
	make build
	make run

dev-fresh:
	@echo "Starting fresh development environment..."
	make clean
	make deps
	make db-reset
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
	make test-comprehensive
	make build

ci-quick:
	@echo "Running quick CI checks..."
	make deps
	make fmt
	make lint
	make test-unit
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

.PHONY: help build run test test-unit test-integration test-verbose test-comprehensive test-coverage test-services test-handlers \
        bench bench-services lint fmt clean clean-all deps db-migrate db-reset watch dev dev-fresh version \
        docker-build docker-run docker-clean ci ci-quick generate verify-separation
