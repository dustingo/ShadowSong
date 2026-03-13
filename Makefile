.PHONY: help build run test clean docker-up docker-down migrate-up migrate-down

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the application
	go build -o bin/server cmd/server/main.go

run: ## Run the application
	go run cmd/server/main.go

test: ## Run tests
	go test -v ./...

clean: ## Clean build artifacts
	rm -rf bin/

docker-up: ## Start Docker services (PostgreSQL and Redis)
	docker-compose up -d

docker-down: ## Stop Docker services
	docker-compose down

deps: ## Download dependencies
	go mod download
	go mod tidy

frontend-install: ## Install frontend dependencies
	cd frontend && pnpm install

frontend-dev: ## Run frontend development server
	cd frontend && pnpm dev

frontend-build: ## Build frontend for production
	cd frontend && pnpm build
