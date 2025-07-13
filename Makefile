# OpenBPL Makefile

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOMOD=$(GOCMD) mod

# Build parameters
BINARY_NAME=openbpl
MAIN_PATH=./cmd/cli
BUILD_DIR=./build

# Colors
GREEN=\033[0;32m
YELLOW=\033[1;33m
NC=\033[0m

.PHONY: help build clean test deps run

# Default target
all: deps build

# Help
help: ## Show available commands
	@echo "$(GREEN)OpenBPL Makefile Commands:$(NC)"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "$(YELLOW)%-15s$(NC) %s\n", $$1, $$2}'

# Build
build: ## Build the CLI binary
	@echo "$(GREEN)Building OpenBPL CLI...$(NC)"
	@mkdir -p $(BUILD_DIR)
	@$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "$(GREEN)Build complete: $(BUILD_DIR)/$(BINARY_NAME)$(NC)"

# Development
dev: ## Run in development mode
	@echo "$(GREEN)Running OpenBPL in development mode...$(NC)"
	@$(GOCMD) run $(MAIN_PATH)

# Dependencies
deps: ## Download dependencies
	@echo "$(GREEN)Downloading dependencies...$(NC)"
	@$(GOMOD) download
	@$(GOMOD) tidy

# Testing
test: ## Run tests
	@echo "$(GREEN)Running tests...$(NC)"
	@$(GOTEST) -v ./...

# Clean
clean: ## Clean build artifacts
	@echo "$(GREEN)Cleaning...$(NC)"
	@$(GOCLEAN)
	@rm -rf $(BUILD_DIR)