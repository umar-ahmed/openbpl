# OpenBPL Makefile
# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=$(GOCMD) fmt

# Build parameters
BINARY_NAME=openbpl
BINARY_UNIX=$(BINARY_NAME)_unix
MAIN_PATH=./cmd/server
BUILD_DIR=./build

# Version info (can be overridden)
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT?=$(shell git rev-parse HEAD 2>/dev/null || echo "unknown")
BUILD_TIME?=$(shell date -u '+%Y-%m-%d_%H:%M:%S')

# Linker flags to embed version info
LDFLAGS=-ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.buildTime=$(BUILD_TIME)"

# Docker parameters
DOCKER_IMAGE=openbpl
DOCKER_TAG?=latest

# Colors for pretty output
GREEN=\033[0;32m
YELLOW=\033[1;33m
RED=\033[0;31m
NC=\033[0m # No Color

.PHONY: help build clean test coverage lint fmt vet deps run dev docker-build docker-run install-tools

# Default target
all: clean deps test build

# Help target - shows available commands
help: ## Show this help message
	@echo "$(GREEN)OpenBPL Makefile Commands:$(NC)"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "$(YELLOW)%-20s$(NC) %s\n", $$1, $$2}'

# Development commands
dev: ## Run the application in development mode with hot reload
	@echo "$(GREEN)Starting development server...$(NC)"
	@air || (echo "$(RED)Air not found. Install with: go install github.com/cosmtrek/air@latest$(NC)" && exit 1)

run: ## Run the application
	@echo "$(GREEN)Running OpenBPL...$(NC)"
	@$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@$(BUILD_DIR)/$(BINARY_NAME)

# Build commands
build: ## Build the binary
	@echo "$(GREEN)Building OpenBPL...$(NC)"
	@mkdir -p $(BUILD_DIR)
	@$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "$(GREEN)Build complete: $(BUILD_DIR)/$(BINARY_NAME)$(NC)"

build-linux: ## Build binary for Linux
	@echo "$(GREEN)Building for Linux...$(NC)"
	@mkdir -p $(BUILD_DIR)
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_UNIX) $(MAIN_PATH)
	@echo "$(GREEN)Linux build complete: $(BUILD_DIR)/$(BINARY_UNIX)$(NC)"

build-all: ## Build binaries for multiple platforms
	@echo "$(GREEN)Building for multiple platforms...$(NC)"
	@mkdir -p $(BUILD_DIR)
	# Linux
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)
	# macOS
	@CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)
	@CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_PATH)
	# Windows
	@CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PATH)
	@echo "$(GREEN)Multi-platform build complete!$(NC)"

# Test commands
test: ## Run tests
	@echo "$(GREEN)Running tests...$(NC)"
	@$(GOTEST) -v ./...

test-short: ## Run tests without integration tests
	@echo "$(GREEN)Running short tests...$(NC)"
	@$(GOTEST) -short -v ./...

test-race: ## Run tests with race condition detection
	@echo "$(GREEN)Running tests with race detection...$(NC)"
	@$(GOTEST) -race -v ./...

coverage: ## Run tests with coverage report
	@echo "$(GREEN)Running tests with coverage...$(NC)"
	@$(GOTEST) -coverprofile=coverage.out ./...
	@$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)Coverage report generated: coverage.html$(NC)"

benchmark: ## Run benchmarks
	@echo "$(GREEN)Running benchmarks...$(NC)"
	@$(GOTEST) -bench=. -benchmem ./...

# Code quality commands
fmt: ## Format Go code
	@echo "$(GREEN)Formatting code...$(NC)"
	@$(GOFMT) ./...

vet: ## Run go vet
	@echo "$(GREEN)Running go vet...$(NC)"
	@$(GOCMD) vet ./...

lint: ## Run golangci-lint
	@echo "$(GREEN)Running golangci-lint...$(NC)"
	@golangci-lint run || (echo "$(RED)golangci-lint not found. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest$(NC)" && exit 1)

security: ## Run gosec security scanner
	@echo "$(GREEN)Running security scan...$(NC)"
	@gosec ./... || (echo "$(RED)gosec not found. Install with: go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest$(NC)" && exit 1)

check: fmt vet lint ## Run all code quality checks

# Dependency management
deps: ## Download dependencies
	@echo "$(GREEN)Downloading dependencies...$(NC)"
	@$(GOMOD) download
	@$(GOMOD) verify

deps-upgrade: ## Upgrade all dependencies
	@echo "$(GREEN)Upgrading dependencies...$(NC)"
	@$(GOMOD) get -u all
	@$(GOMOD) tidy

deps-tidy: ## Clean up dependencies
	@echo "$(GREEN)Tidying dependencies...$(NC)"
	@$(GOMOD) tidy

# Docker commands
docker-build: ## Build Docker image
	@echo "$(GREEN)Building Docker image...$(NC)"
	@docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .
	@echo "$(GREEN)Docker image built: $(DOCKER_IMAGE):$(DOCKER_TAG)$(NC)"

docker-run: ## Run Docker container
	@echo "$(GREEN)Running Docker container...$(NC)"
	@docker run -p 8080:8080 --rm $(DOCKER_IMAGE):$(DOCKER_TAG)

docker-compose-up: ## Start services with docker-compose
	@echo "$(GREEN)Starting services with docker-compose...$(NC)"
	@docker-compose up -d

docker-compose-down: ## Stop services with docker-compose
	@echo "$(GREEN)Stopping services with docker-compose...$(NC)"
	@docker-compose down

docker-compose-logs: ## View docker-compose logs
	@docker-compose logs -f

# Database commands (for future use)
db-migrate-up: ## Run database migrations up
	@echo "$(GREEN)Running database migrations up...$(NC)"
	@migrate -path ./migrations -database "$(DATABASE_URL)" up

db-migrate-down: ## Run database migrations down
	@echo "$(GREEN)Running database migrations down...$(NC)"
	@migrate -path ./migrations -database "$(DATABASE_URL)" down

db-migrate-create: ## Create a new migration file (usage: make db-migrate-create NAME=migration_name)
	@echo "$(GREEN)Creating migration: $(NAME)$(NC)"
	@migrate create -ext sql -dir ./migrations $(NAME)

# Cleanup commands
clean: ## Clean build artifacts
	@echo "$(GREEN)Cleaning...$(NC)"
	@$(GOCLEAN)
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html

clean-docker: ## Clean Docker images and containers
	@echo "$(GREEN)Cleaning Docker artifacts...$(NC)"
	@docker system prune -f

# Installation commands
install-tools: ## Install development tools
	@echo "$(GREEN)Installing development tools...$(NC)"
	@go install github.com/cosmtrek/air@latest
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
	@go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	@echo "$(GREEN)Development tools installed!$(NC)"

# Release commands
release-patch: ## Create a patch release
	@echo "$(GREEN)Creating patch release...$(NC)"
	@./scripts/release.sh patch

release-minor: ## Create a minor release
	@echo "$(GREEN)Creating minor release...$(NC)"
	@./scripts/release.sh minor

release-major: ## Create a major release
	@echo "$(GREEN)Creating major release...$(NC)"
	@./scripts/release.sh major

# Info commands
version: ## Show version information
	@echo "Version: $(VERSION)"
	@echo "Commit: $(COMMIT)"
	@echo "Build Time: $(BUILD_TIME)"

info: ## Show project information
	@echo "$(GREEN)OpenBPL Project Information:$(NC)"
	@echo "Go version: $(shell go version)"
	@echo "Project: $(shell pwd)"
	@echo "Git branch: $(shell git branch --show-current 2>/dev/null || echo 'not a git repo')"
	@echo "Git status: $(shell git status --porcelain 2>/dev/null | wc -l | xargs echo) modified files"

# CI/CD simulation
ci: deps check test build ## Run CI pipeline locally