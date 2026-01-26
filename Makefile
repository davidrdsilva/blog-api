.PHONY: help run build down logs test clean

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

run: ## Start all services with docker-compose
	docker compose up --build

run-detached: ## Start services in background
	docker compose up -d --build

build: ## Build the Go application
	go build -o bin/api ./cmd/api

down: ## Stop all services
	docker compose down

logs: ## View logs from all services
	docker compose logs -f

logs-api: ## View API logs only
	docker compose logs -f api

clean: ## Remove containers, volumes, and built binaries
	docker compose down -v
	rm -rf bin/

test: ## Run tests (placeholder for now)
	go test ./... -v

dev: ## Run API locally (requires postgres and minio running)
	go run ./cmd/api/main.go
