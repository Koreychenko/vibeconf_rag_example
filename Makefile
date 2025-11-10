# Go RAG System Makefile

# Variables
APP_NAME := rag-service
MAIN_PATH := ./cmd/api
BINARY := $(APP_NAME)
DOCKER_COMPOSE_FILE := ./deployments/docker-compose.yml

# Go build flags
GO_BUILD_FLAGS := -v

# Default goal
.PHONY: all
all: build

# Build the application
.PHONY: build
build:
	@echo "Building $(APP_NAME)..."
	@go build $(GO_BUILD_FLAGS) -o $(BINARY) $(MAIN_PATH)

# Run the application locally
.PHONY: run
run: build
	@echo "Running $(APP_NAME)..."
	@./$(BINARY)

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	@rm -f $(BINARY)
	@go clean

# Format Go code
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	@go fmt ./...

# Run tests
.PHONY: test
test:
	@echo "Running tests..."
	@go test -v ./...

# Docker commands
# Start all services with docker compose
.PHONY: docker-up
docker-up:
	@echo "Starting Docker containers..."
	@docker compose --env-file .env -f $(DOCKER_COMPOSE_FILE) up -d

# Stop all services
.PHONY: docker-down
docker-down:
	@echo "Stopping Docker containers..."
	@docker compose --env-file .env -f $(DOCKER_COMPOSE_FILE) down

# Start all services in foreground with logs
.PHONY: docker-logs
docker-logs:
	@echo "Starting Docker containers with logs..."
	@docker compose --env-file .env -f $(DOCKER_COMPOSE_FILE) up

# Rebuild and restart only the app container
.PHONY: docker-rebuild
docker-rebuild:
	@echo "Rebuilding and restarting app container..."
	@docker compose --env-file .env -f $(DOCKER_COMPOSE_FILE) up -d --build app

# Show Docker container status
.PHONY: docker-ps
docker-ps:
	@echo "Docker container status:"
	@docker compose --env-file .env -f $(DOCKER_COMPOSE_FILE) ps

# Build the data loader
.PHONY: build-loader
build-loader:
	@echo "Building data loader..."
	@go build $(GO_BUILD_FLAGS) -o dataloader ./cmd/dataloader

# Load sample data (locally)
.PHONY: load-samples
load-samples: build-loader
	@echo "Loading sample data..."
	@./dataloader -dir ./data/samples

# Load sample data using Docker container
.PHONY: docker-load-samples
docker-load-samples:
	@echo "Loading sample data in Docker container..."
	@docker compose --env-file .env -f $(DOCKER_COMPOSE_FILE) exec app ./dataloader -dir ./data/samples

# All-in-one developer setup
.PHONY: dev-setup
dev-setup: build docker-up
	@echo "Development environment set up!"
	@echo "Run 'make docker-ps' to see container status"

# Help
.PHONY: help
help:
	@echo "Go RAG System Makefile"
	@echo ""
	@echo "Usage:"
	@echo "  make              Build the application"
	@echo "  make build        Build the application"
	@echo "  make run          Run the application locally"
	@echo "  make clean        Clean build artifacts"
	@echo "  make fmt          Format Go code"
	@echo "  make test         Run tests"
	@echo "  make docker-up    Start all containers in the background"
	@echo "  make docker-down  Stop all containers"
	@echo "  make docker-logs  Start all containers with logs in foreground"
	@echo "  make docker-rebuild Rebuild and restart only the app container"
	@echo "  make docker-ps    Show Docker container status"
	@echo "  make dev-setup    Set up the development environment"
	@echo "  make help         Show this help message"