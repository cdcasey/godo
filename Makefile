.PHONY: help build test migrate-up migrate-down docker-build docker-up docker-down lint clean run

# Variables
BINARY_NAME=godo
DOCKER_IMAGE=godo:latest
MIGRATIONS_PATH=./migrations
DATABASE_URL?=./data/todos.db

help: ## Show this help message
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "make %-15s %s\n", $$1, $$2}'

build: ## Build the application binary
	@echo "Building $(BINARY_NAME)..."
	@go build -o bin/$(BINARY_NAME) ./cmd/api

run: ## Run the application (requires .env file)
	@echo "Running $(BINARY_NAME)..."
	@go run ./cmd/api

test: ## Run tests with coverage
	@echo "Running tests..."
	@go test -v -cover ./...

test-coverage: ## Run tests with detailed coverage report
	@echo "Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

migrate-up: ## Run database migrations up
	@echo "Running migrations up..."
	@migrate -path $(MIGRATIONS_PATH) -database sqlite3://$(DATABASE_URL) up

migrate-down: ## Rollback last migration
	@echo "Rolling back last migration..."
	@migrate -path $(MIGRATIONS_PATH) -database sqlite3://$(DATABASE_URL) down 1

migrate-force: ## Force migration version (use: make migrate-force VERSION=1)
	@migrate -path $(MIGRATIONS_PATH) -database sqlite3://$(DATABASE_URL) force $(VERSION)

docker-build: ## Build Docker image
	@echo "Building Docker image..."
	@docker build -t $(DOCKER_IMAGE) .

docker-up: ## Start services with docker-compose
	@echo "Starting services..."
	@docker-compose up -d

docker-down: ## Stop services
	@echo "Stopping services..."
	@docker-compose down

lint: ## Run golangci-lint
	@echo "Running linter..."
	@golangci-lint run

clean: ## Clean build artifacts and test coverage
	@echo "Cleaning..."
	@rm -rf bin/
	@rm -f coverage.out coverage.html

.DEFAULT_GOAL := help
