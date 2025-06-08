DOCKER_CMD := sudo docker

# Import .env file
include .env
export

# Variables
SHELL := fish

# Install goose
install-goose:
	@echo "Installing goose..."
	@go install github.com/pressly/goose/v3/cmd/goose@latest

# Database migration commands
migrate-up:
	@echo "Running database migrations..."
	@goose up

migrate-down:
	@echo "Rolling back database migrations..."
	@goose down

migrate-status:
	@echo "Checking migration status..."
	@goose status

migrate-create:
	@echo "Creating new migration: $(name)"
	@goose create $(name) sql

# Development commands
dev-setup: install-goose
	@echo "Setting up development environment..."
	@go mod download
	@$(DOCKER_CMD) compose up -d
	@sleep 5
	@make migrate-up

dev-run:
	@echo "Starting development server..."
	@go run cmd/web/main.go

dev-worker:
	@echo "Starting worker..."
	@go run cmd/worker/main.go

# Test commands
test:
	@echo "Running tests..."
	@go test ./...

test-integration:
	@echo "Running integration tests..."
	@$(DOCKER_CMD) compose up -d
	@sleep 5
	@go test ./... -tags=integration
	@$(DOCKER_CMD) compose down

# Clean up
clean:
	@echo "Cleaning up..."
	@$(DOCKER_CMD) compose down
	@go clean -cache

.PHONY: install-goose migrate-up migrate-down migrate-status migrate-create dev-setup dev-run dev-worker test test-integration clean help 