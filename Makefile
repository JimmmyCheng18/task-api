# Makefile for Task API

# Variables
APP_NAME := task-api
VERSION := $(shell cat VERSION 2>/dev/null || echo "1.0.0")
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
LDFLAGS := -X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT)

# Directories
BUILD_DIR := ./bin
COVERAGE_DIR := ./coverage

# Go related variables
GOCMD := go
GOBUILD := $(GOCMD) build
GOCLEAN := $(GOCMD) clean
GOTEST := $(GOCMD) test
GOGET := $(GOCMD) get
GOMOD := $(GOCMD) mod
GOFMT := gofmt

# Colors for output
RED := \033[0;31m
GREEN := \033[0;32m
YELLOW := \033[1;33m
BLUE := \033[0;34m
NC := \033[0m # No Color

.PHONY: help build clean test test-coverage test-unit \
        run dev lint format security deps check docker docker-build docker-run \
        compose-build compose-up compose-down compose-logs compose-restart compose-status \
        install-tools setup version \
        release deploy-dev deploy-prod benchmark profile docs

# Default target
.DEFAULT_GOAL := help

## help: Show this help message
help:
	@echo "$(GREEN)Task API Makefile$(NC)"
	@echo ""
	@echo "Available targets:"
	@awk '/^[a-zA-Z\-\_0-9]+:/ { \
		helpMessage = match(lastLine, /^## (.*)/); \
		if (helpMessage) { \
			helpCommand = substr($$1, 0, index($$1, ":")-1); \
			helpMessage = substr(lastLine, RSTART + 3, RLENGTH); \
			printf "  $(YELLOW)%-20s$(NC) %s\n", helpCommand, helpMessage; \
		} \
	} \
	{ lastLine = $$0 }' $(MAKEFILE_LIST)

## build: Build the application
build: swagger-generate
	@echo "$(GREEN)🚀 Building $(APP_NAME)...$(NC)"
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(APP_NAME) ./cmd/server
	@echo "$(GREEN)✅ Build completed: $(BUILD_DIR)/$(APP_NAME)$(NC)"

## build-all: Build for all platforms
build-all: swagger-generate
	@echo "$(GREEN)🚀 Building $(APP_NAME) for all platforms...$(NC)"
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 $(GOBUILD) -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(APP_NAME)-linux-amd64 ./cmd/server
	GOOS=linux GOARCH=arm64 $(GOBUILD) -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(APP_NAME)-linux-arm64 ./cmd/server
	GOOS=darwin GOARCH=amd64 $(GOBUILD) -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(APP_NAME)-darwin-amd64 ./cmd/server
	GOOS=darwin GOARCH=arm64 $(GOBUILD) -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(APP_NAME)-darwin-arm64 ./cmd/server
	GOOS=windows GOARCH=amd64 $(GOBUILD) -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(APP_NAME)-windows-amd64.exe ./cmd/server
	@echo "$(GREEN)✅ Multi-platform build completed$(NC)"

## clean: Clean build artifacts
clean:
	@echo "$(BLUE)🧹 Cleaning build artifacts...$(NC)"
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)
	rm -rf $(COVERAGE_DIR)
	@echo "$(GREEN)✅ Clean completed$(NC)"

## test: Run all tests
test:
	@echo "$(BLUE)🧪 Running all tests...$(NC)"
	$(GOTEST) -v -race ./...

## test-coverage: Run tests with coverage
test-coverage:
	@echo "$(BLUE)📊 Running tests with coverage...$(NC)"
	@mkdir -p $(COVERAGE_DIR)
	$(GOTEST) -v -race -coverprofile=$(COVERAGE_DIR)/coverage.out -covermode=atomic ./...
	$(GOCMD) tool cover -html=$(COVERAGE_DIR)/coverage.out -o $(COVERAGE_DIR)/coverage.html
	$(GOCMD) tool cover -func=$(COVERAGE_DIR)/coverage.out | grep total

## test-unit: Run unit tests only
test-unit:
	@echo "$(BLUE)🔬 Running unit tests...$(NC)"
	$(GOTEST) -v -race -short ./...


## benchmark: Run benchmark tests with coverage
benchmark:
	@echo "$(BLUE)⚡ Running benchmark tests with coverage...$(NC)"
	@mkdir -p $(COVERAGE_DIR)
	$(GOTEST) -bench=. -benchmem -coverprofile=$(COVERAGE_DIR)/benchmark-coverage.out -covermode=atomic ./...
	@echo "$(GREEN)📊 Benchmark Coverage Summary:$(NC)"
	$(GOCMD) tool cover -func=$(COVERAGE_DIR)/benchmark-coverage.out | grep total
	$(GOCMD) tool cover -html=$(COVERAGE_DIR)/benchmark-coverage.out -o $(COVERAGE_DIR)/benchmark-coverage.html
	@echo "$(GREEN)✅ Benchmark coverage report: $(COVERAGE_DIR)/benchmark-coverage.html$(NC)"

## run: Run the application (use PORT=xxxx to specify port)
run: build
	@echo "$(GREEN)🚀 Running $(APP_NAME)...$(NC)"
	@if [ -n "$(PORT)" ]; then \
		echo "$(YELLOW)🚀 Starting on port $(PORT)$(NC)"; \
		PORT=$(PORT) ./$(BUILD_DIR)/$(APP_NAME); \
	else \
		echo "$(YELLOW)🚀 Starting on default port 8080 (use PORT=xxxx to specify different port)$(NC)"; \
		./$(BUILD_DIR)/$(APP_NAME); \
	fi

## dev: Run in development mode (use PORT=xxxx to specify port)
dev: swagger-generate
	@echo "$(GREEN)🔧 Running in development mode...$(NC)"
	@if [ -n "$(PORT)" ]; then \
		echo "$(YELLOW)🚀 Starting on port $(PORT)$(NC)"; \
		PORT=$(PORT) GIN_MODE=debug $(GOCMD) run ./cmd/server; \
	else \
		echo "$(YELLOW)🚀 Starting on default port 8080 (use PORT=xxxx to specify different port)$(NC)"; \
		GIN_MODE=debug $(GOCMD) run ./cmd/server; \
	fi

## lint: Run linting
lint:
	@echo "$(BLUE)🔍 Running linting...$(NC)"
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "$(YELLOW)⚠️  golangci-lint not found. Install with: make install-tools$(NC)"; \
	fi

## format: Format code
format:
	@echo "$(BLUE)📝 Formatting code...$(NC)"
	$(GOFMT) -w .
	$(GOMOD) tidy

## format-check: Check if code is formatted
format-check:
	@echo "$(BLUE)📝 Checking code formatting...$(NC)"
	@unformatted=$$($(GOFMT) -l .); \
	if [ -n "$$unformatted" ]; then \
		echo "$(RED)❌ The following files are not formatted:$(NC)"; \
		echo "$$unformatted"; \
		exit 1; \
	else \
		echo "$(GREEN)✅ All files are properly formatted$(NC)"; \
	fi

## security: Run security checks
security:
	@echo "$(BLUE)🔒 Running security checks...$(NC)"
	@if command -v gosec >/dev/null 2>&1; then \
		gosec ./...; \
	else \
		echo "$(YELLOW)⚠️  gosec not found. Install with: make install-tools$(NC)"; \
	fi

## deps: Download and verify dependencies
deps:
	@echo "$(BLUE)📦 Downloading dependencies...$(NC)"
	$(GOMOD) download
	$(GOMOD) verify
	$(GOMOD) tidy

## check: Run all checks (format, lint, security, test)
check: format-check lint security test

## docker-build: Build Docker image
docker-build: swagger-generate
	@echo "$(BLUE)🐳 Building Docker image...$(NC)"
	docker build -t $(APP_NAME):$(VERSION) -f ./Dockerfile .
	docker tag $(APP_NAME):$(VERSION) $(APP_NAME):latest

## docker-run: Run Docker container (use PORT=xxxx to specify host port)
docker-run:
	@echo "$(BLUE)🐳 Running Docker container...$(NC)"
	@if [ -n "$(PORT)" ]; then \
		echo "$(YELLOW)🐳 Mapping host port $(PORT) to container port 8080$(NC)"; \
		docker run --rm -d -p $(PORT):8080 -e PORT=8080 $(APP_NAME):latest; \
	else \
		echo "$(YELLOW)🐳 Using default port mapping 8080:8080 (use PORT=xxxx to specify different host port)$(NC)"; \
		docker run --rm -d -p 8080:8080 -e PORT=8080 $(APP_NAME):latest; \
	fi


## install-tools: Install development tools
install-tools:
	@echo "$(BLUE)🛠️  Installing development tools...$(NC)"
	$(GOCMD) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	$(GOCMD) install github.com/securego/gosec/v2/cmd/gosec@latest
	$(GOCMD) install github.com/swaggo/swag/cmd/swag@latest
	@echo "$(GREEN)✅ Tools installed$(NC)"

## setup: Setup development environment
setup: deps install-tools
	@echo "$(GREEN)✅ Development environment setup completed$(NC)"

## version: Show version information
version:
	@echo "$(GREEN)📋 Version Information$(NC)"
	@echo "App Name: $(APP_NAME)"
	@echo "Version: $(VERSION)"
	@echo "Build Time: $(BUILD_TIME)"
	@echo "Git Commit: $(GIT_COMMIT)"

## profile: Run application with profiling
profile: build
	@echo "$(BLUE)📊 Running with profiling enabled...$(NC)"
	ENABLE_PROFILING=true ./$(BUILD_DIR)/$(APP_NAME)

## swagger-generate: Generate Swagger documentation
swagger-generate:
	@echo "$(BLUE)📚 Generating Swagger documentation...$(NC)"
	@if command -v swag >/dev/null 2>&1; then \
		swag init -g cmd/server/main.go -o ./docs; \
	elif [ -f "$(shell go env GOPATH)/bin/swag" ]; then \
		$(shell go env GOPATH)/bin/swag init -g cmd/server/main.go -o ./docs; \
	else \
		echo "$(YELLOW)⚠️  swag not found. Install with: make install-tools$(NC)"; \
	fi

## docs: Generate documentation (alias for swagger-generate)
docs: swagger-generate
	@echo "$(GREEN)✅ Documentation generated successfully$(NC)"

## release: Create a release build
release: clean format-check lint security test build-all
	@echo "$(GREEN)🎉 Release build completed$(NC)"
	@echo "$(YELLOW)📦 Built binaries:$(NC)"
	@ls -la $(BUILD_DIR)

## deploy-dev: Deploy to development environment
deploy-dev: build docker-build
	@echo "$(BLUE)🚀 Deploying to development environment...$(NC)"
	# Add your development deployment commands here

## deploy-prod: Deploy to production environment
deploy-prod: release docker-build
	@echo "$(BLUE)🚀 Deploying to production environment...$(NC)"
	# Add your production deployment commands here

## compose-build: Build all services with docker-compose
compose-build:
	@echo "$(BLUE)🐳 Building all services with docker-compose...$(NC)"
	docker-compose build

## compose-up: Start all services with docker-compose
compose-up:
	@echo "$(GREEN)🚀 Starting all services...$(NC)"
	@echo "$(YELLOW)Frontend: http://localhost:3666$(NC)"
	@echo "$(YELLOW)Backend API: http://localhost:3333$(NC)"
	docker-compose up -d

## compose-down: Stop all services
compose-down:
	@echo "$(BLUE)🛑 Stopping all services...$(NC)"
	docker-compose down

## compose-logs: View logs from all services
compose-logs:
	@echo "$(BLUE)📋 Viewing logs from all services...$(NC)"
	docker-compose logs -f

## compose-restart: Restart all services
compose-restart: compose-down compose-up

## compose-status: Show status of all services
compose-status:
	@echo "$(BLUE)📊 Service status:$(NC)"
	docker-compose ps

## compose-deploy: Build and deploy all services
compose-deploy: compose-build compose-up
	@echo "$(GREEN)🎉 Full stack deployed successfully!$(NC)"
	@echo "$(YELLOW)Frontend: http://localhost:3666$(NC)"
	@echo "$(YELLOW)Backend API: http://localhost:3333$(NC)"
	@echo "$(YELLOW)API Documentation: http://localhost:3333/swagger/index.html$(NC)"

# Development shortcuts
.PHONY: b r t c d cup cdown clogs cstatus
b: build
r: run
t: test
c: clean
d: dev
cup: compose-up
cdown: compose-down
clogs: compose-logs
cstatus: compose-status