# Makefile for OpenShift MCP Go

# Build variables
BINARY_NAME=openshift-mcp
CLI_BINARY_NAME=oc-ai
BUILD_DIR=bin
CMD_DIR=cmd/openshift-mcp
CLI_CMD_DIR=cmd/oc-ai
MAIN_FILE=main.go

# Version information
VERSION ?= $(shell git describe --tags --always --dirty)
COMMIT ?= $(shell git rev-parse HEAD)
DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Build flags
LDFLAGS=-ldflags="-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)"

# Go variables
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOCLEAN=$(GOCMD) clean
GOMOD=$(GOCMD) mod
GOVET=$(GOCMD) vet
GOFMT=$(GOCMD) fmt

# Default target
.PHONY: all
all: clean deps test build build-cli

# Build the MCP server
.PHONY: build
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./$(CMD_DIR)

# Build the CLI client
.PHONY: build-cli
build-cli:
	@echo "Building $(CLI_BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(CLI_BINARY_NAME) ./$(CLI_CMD_DIR)

# Build both server and CLI
.PHONY: build-both
build-both: build build-cli

# Build for multiple platforms
.PHONY: build-all
build-all: build-linux build-darwin build-windows

.PHONY: build-linux
build-linux:
	@echo "Building for Linux..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 ./$(CMD_DIR)
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(CLI_BINARY_NAME)-linux-amd64 ./$(CLI_CMD_DIR)

.PHONY: build-darwin
build-darwin:
	@echo "Building for macOS..."
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 ./$(CMD_DIR)
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 ./$(CMD_DIR)
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(CLI_BINARY_NAME)-darwin-amd64 ./$(CLI_CMD_DIR)
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(CLI_BINARY_NAME)-darwin-arm64 ./$(CLI_CMD_DIR)

.PHONY: build-windows
build-windows:
	@echo "Building for Windows..."
	@mkdir -p $(BUILD_DIR)
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe ./$(CMD_DIR)
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(CLI_BINARY_NAME)-windows-amd64.exe ./$(CLI_CMD_DIR)

# Download dependencies
.PHONY: deps
deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

# Run tests
.PHONY: test
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

# Run tests with coverage
.PHONY: test-coverage
test-coverage:
	@echo "Running tests with coverage..."
	$(GOTEST) -cover ./...
	$(GOTEST) -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

# Run tests with race detection
.PHONY: test-race
test-race:
	@echo "Running tests with race detection..."
	$(GOTEST) -race ./...

# Lint code
.PHONY: lint
lint:
	@echo "Linting code..."
	@which golangci-lint > /dev/null || (echo "golangci-lint not installed, please install it: https://golangci-lint.run/usage/install/" && exit 1)
	golangci-lint run

# Format code
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	$(GOFMT) ./...

# Vet code
.PHONY: vet
vet:
	@echo "Vetting code..."
	$(GOVET) ./...

# Security scan
.PHONY: security
security:
	@echo "Running security scan..."
	@which gosec > /dev/null || (echo "gosec not installed, installing..." && go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest)
	gosec ./...

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

# Run the application
.PHONY: run
run: build
	@echo "Running $(BINARY_NAME)..."
	./$(BUILD_DIR)/$(BINARY_NAME) --debug

# Run the application in development mode
.PHONY: dev
dev:
	@echo "Running in development mode..."
	$(GOCMD) run ./$(CMD_DIR) --debug --port 8080

# Install the application
.PHONY: install
install: build
	@echo "Installing $(BINARY_NAME)..."
	sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/

# Uninstall the application
.PHONY: uninstall
uninstall:
	@echo "Uninstalling $(BINARY_NAME)..."
	sudo rm -f /usr/local/bin/$(BINARY_NAME)

# Container targets (using Podman)
.PHONY: container-build
container-build:
	@echo "Building container image with Podman..."
	podman build \
		--build-arg VERSION=$(VERSION) \
		--build-arg COMMIT=$(COMMIT) \
		--build-arg DATE=$(DATE) \
		-t openshift-mcp:$(VERSION) \
		-t openshift-mcp:latest .

.PHONY: container-run
container-run:
	@echo "Running container with Podman..."
	podman run --rm -p 8080:8080 \
		-e GEMINI_API_KEY="$(GEMINI_API_KEY)" \
		-e KUBECONFIG=/tmp/kubeconfig \
		-v $(HOME)/.kube/config:/tmp/kubeconfig:ro \
		openshift-mcp:latest

.PHONY: container-push
container-push:
	@echo "Pushing container image..."
	podman push openshift-mcp:$(VERSION)
	podman push openshift-mcp:latest

# Legacy Docker targets (for compatibility)
.PHONY: docker-build
docker-build: container-build

.PHONY: docker-run
docker-run: container-run

.PHONY: docker-push
docker-push: container-push

# Development setup
.PHONY: setup
setup:
	@echo "Setting up development environment..."
	$(GOMOD) download
	@echo "Installing development tools..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
	@echo "Setup complete!"

# Check all (comprehensive check)
.PHONY: check
check: deps fmt vet lint security test

# Release preparation
.PHONY: release
release: clean check build-all
	@echo "Release $(VERSION) ready in $(BUILD_DIR)/"

# Help
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  build         - Build the application"
	@echo "  build-all     - Build for all platforms"
	@echo "  deps          - Download dependencies"
	@echo "  test          - Run tests"
	@echo "  test-coverage - Run tests with coverage"
	@echo "  test-race     - Run tests with race detection"
	@echo "  lint          - Lint code"
	@echo "  fmt           - Format code"
	@echo "  vet           - Vet code"
	@echo "  security      - Run security scan"
	@echo "  clean         - Clean build artifacts"
	@echo "  run           - Build and run the application"
	@echo "  dev           - Run in development mode"
	@echo "  install       - Install the application"
	@echo "  uninstall     - Uninstall the application"
	@echo "  container-build - Build container image (Podman)"
	@echo "  container-run - Run container (Podman)"
	@echo "  container-push - Push container image (Podman)"
	@echo "  docker-build  - Build Docker image (legacy)"
	@echo "  docker-run    - Run Docker container (legacy)"
	@echo "  docker-push   - Push Docker image (legacy)"
	@echo "  setup         - Setup development environment"
	@echo "  check         - Run all checks (fmt, vet, lint, security, test)"
	@echo "  release       - Prepare release build"
	@echo "  help          - Show this help"
