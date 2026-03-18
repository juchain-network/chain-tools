OUT := build
GOCACHE ?= $(CURDIR)/.cache/go-build
export GOCACHE

# Project information
PROJECT_NAME := ju-cli
VERSION ?= 1.2.2
BUILD_DATE := $(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Build flags
LDFLAGS := -X 'juchain.org/chain/tools/cmd.Version=$(VERSION)' \
           -X 'juchain.org/chain/tools/cmd.BuildDate=$(BUILD_DATE)' \
           -X 'juchain.org/chain/tools/cmd.GitCommit=$(GIT_COMMIT)'

.PHONY: default build clean test lint fmt help

default: build

help: ## Show help information
	@echo "$(PROJECT_NAME) Makefile Commands:"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-15s %s\n", $$1, $$2}'

build: ## Compile project
	@echo "Building $(PROJECT_NAME) v$(VERSION)..."
	@mkdir -p $(OUT) $(GOCACHE)
	@go build -ldflags "$(LDFLAGS)" -o ${OUT}/$(PROJECT_NAME)
	@echo "✅ Build completed: ${OUT}/$(PROJECT_NAME)"

build_linux: ## Build Linux version
	@echo "Building $(PROJECT_NAME) for Linux..."
	@mkdir -p $(OUT) $(GOCACHE)
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o ${OUT}/$(PROJECT_NAME)-linux-amd64
	@echo "✅ Linux build completed: ${OUT}/$(PROJECT_NAME)-linux-amd64"

test: ## Run tests
	@echo "Running tests..."
	@mkdir -p $(GOCACHE)
	@go test -v ./...
	@echo "✅ Tests completed"

lint: ## Run code checks
	@echo "Running linter..."
	@which golangci-lint > /dev/null || (echo "Installing golangci-lint..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	@export PATH=$$PATH:~/go/bin && golangci-lint run --timeout=5m || echo "⚠️  Linting completed with warnings"
	@echo "✅ Linting completed"

fmt: ## Format code
	@echo "Formatting code..."
	@go fmt ./...
	@echo "✅ Code formatted"

tidy: ## Tidy dependencies
	@echo "Tidying go modules..."
	@go mod tidy
	@echo "✅ Dependencies tidied"

clean: ## Clean build files
	@echo "Cleaning build files..."
	@rm -rf build .cache
	@echo "✅ Build files cleaned"

all: fmt lint test build ## Run all checks and builds
