.PHONY: help build run stop clean test docker-build docker-up docker-down docker-logs docker-ps

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the Go binary
	go build -o server ./cmd/server

run: ## Run the server locally
	go run ./cmd/server

test: ## Run tests
	go test -v ./...

clean: ## Clean build artifacts and logs
	rm -f server
	rm -rf logs/*.log*

docker-build: ## Build Docker image
	docker-compose -f docker/docker-compose.yml build

docker-up: ## Start Docker containers
	docker-compose -f docker/docker-compose.yml up -d

docker-down: ## Stop Docker containers
	docker-compose -f docker/docker-compose.yml down

docker-logs: ## Show Docker logs
	docker-compose -f docker/docker-compose.yml logs -f app

docker-ps: ## Show Docker container status
	docker-compose -f docker/docker-compose.yml ps

docker-restart: ## Restart app container
	docker-compose -f docker/docker-compose.yml restart app

docker-clean: ## Remove all containers, images, and volumes
	docker-compose -f docker/docker-compose.yml down -v --rmi all

dev-up: ## Start development environment
	docker-compose -f docker/docker-compose.yml -f docker/docker-compose.dev.yml up

dev-down: ## Stop development environment
	docker-compose -f docker/docker-compose.yml -f docker/docker-compose.dev.yml down
