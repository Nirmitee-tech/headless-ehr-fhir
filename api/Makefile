.PHONY: help build test test-integration test-coverage lint fmt seed migrate-up migrate-status docker-up docker-down docker-logs clean dev

APP_NAME := ehr-server
BUILD_DIR := ./bin
CMD_DIR := ./cmd/ehr-server

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

build: ## Build the server binary
	go build -o $(BUILD_DIR)/$(APP_NAME) $(CMD_DIR)

test: ## Run unit tests
	go test ./... -race -count=1

test-integration: ## Run integration tests (requires Docker)
	go test ./test/integration/ -v -count=1 -timeout 120s

test-coverage: ## Run tests with coverage report
	go test ./... -coverprofile=coverage.out -covermode=atomic
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

lint: ## Run linter
	golangci-lint run ./...

fmt: ## Format code
	gofmt -s -w .
	goimports -w .

seed: ## Seed reference data
	./scripts/seed.sh

migrate-up: ## Run migrations
	go run $(CMD_DIR) migrate up

migrate-status: ## Show migration status
	go run $(CMD_DIR) migrate status

docker-up: ## Start all services
	docker compose up -d

docker-down: ## Stop all services
	docker compose down

docker-logs: ## Tail service logs
	docker compose logs -f $(APP_NAME)

clean: ## Clean build artifacts
	rm -rf $(BUILD_DIR) coverage.out coverage.html

dev: docker-up migrate-up seed ## Full dev setup

.DEFAULT_GOAL := help
