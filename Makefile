.DEFAULT_GOAL := help

# Variables
APP_NAME = codi
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME = $(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS = -ldflags="-w -s -X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME)"

# Check dev tools
GOLANGCI_LINT := $(shell command -v golangci-lint 2>/dev/null)

.PHONY: help
help: ## Show help message
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-15s %s\n", $$1, $$2}'

## ────────────────────────────────────────────────────────────
## Setup
## ────────────────────────────────────────────────────────────

.PHONY: setup
setup: ## Setup development environment
	@echo "Downloading dependencies..."
	go mod download
	go mod verify
	@echo "Installing dev tools..."
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.15.0
	@echo "Setup complete"

## ────────────────────────────────────────────────────────────
## Build
## ────────────────────────────────────────────────────────────

.PHONY: build
build: ## Build codi binary
	@echo "Building codi..."
	CGO_ENABLED=0 go build $(LDFLAGS) -o bin/codi ./cmd/codi

## ────────────────────────────────────────────────────────────
## Test
## ────────────────────────────────────────────────────────────

.PHONY: test
test: ## Run tests
	@echo "Running tests..."
	go test -v -race ./...

.PHONY: coverage
coverage: ## Generate test coverage
	@echo "Running test coverage..."
	go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

## ────────────────────────────────────────────────────────────
## Code Quality
## ────────────────────────────────────────────────────────────

.PHONY: lint
lint: ## Run all checks and formatting
ifndef GOLANGCI_LINT
	$(error golangci-lint not installed, run: make setup)
endif
	@echo "Tidying Go modules..."
	go mod tidy
	@echo "Formatting code..."
	gofmt -s -w .
	@echo "Running linter..."
	golangci-lint run --fix --timeout=5m

## ────────────────────────────────────────────────────────────
## Clean
## ────────────────────────────────────────────────────────────

.PHONY: clean
clean: ## Clean build artifacts
	@echo "Cleaning..."
	rm -rf bin/
	rm -f coverage.out coverage.html
