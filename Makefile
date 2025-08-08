# Makefile for tmpltr

# Build variables
BINARY_NAME=tmpltr
BUILD_DIR=build
GO_FILES=$(shell find . -name "*.go" -not -path "./vendor/*")

# Go commands
GO=go
GOBUILD=$(GO) build
GOCLEAN=$(GO) clean
GOTEST=$(GO) test
GOGET=$(GO) get
GOMOD=$(GO) mod

# Build flags
LDFLAGS=-ldflags "-s -w"
BUILD_FLAGS=-v $(LDFLAGS)

.PHONY: all build clean test test-unit test-integration test-coverage deps tidy fmt lint help

# Default target
all: clean fmt test build

# Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) .
	@echo "Built $(BUILD_DIR)/$(BINARY_NAME)"

# Clean build artifacts
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	@rm -rf $(BUILD_DIR)
	@rm -f $(BINARY_NAME)
	@rm -f $(BINARY_NAME)-test
	@echo "Cleaned"

# Run all tests
test: test-unit test-integration

# Run unit tests
test-unit:
	@echo "Running unit tests..."
	$(GOTEST) -v ./cmd/...
	$(GOTEST) -v ./internal/...
	@echo "Unit tests completed"

# Run integration tests
test-integration: build
	@echo "Running integration tests..."
	$(GOTEST) -v -tags=integration ./...
	@echo "Integration tests completed"

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	@mkdir -p $(BUILD_DIR)
	$(GOTEST) -v -coverprofile=$(BUILD_DIR)/coverage.out ./...
	$(GO) tool cover -html=$(BUILD_DIR)/coverage.out -o $(BUILD_DIR)/coverage.html
	@echo "Coverage report generated at $(BUILD_DIR)/coverage.html"

# Run tests with race detection
test-race:
	@echo "Running tests with race detection..."
	$(GOTEST) -v -race ./...
	@echo "Race detection tests completed"

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	$(GOGET) -d ./...
	@echo "Dependencies downloaded"

# Tidy up go.mod
tidy:
	@echo "Tidying go.mod..."
	$(GOMOD) tidy
	@echo "go.mod tidied"

# Format code
fmt:
	@echo "Formatting code..."
	$(GO) fmt ./...
	@echo "Code formatted"

# Run linting (requires golangci-lint)
lint:
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed. Install with: curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b \$$(go env GOPATH)/bin v1.54.2"; \
	fi

# Vet code
vet:
	@echo "Vetting code..."
	$(GO) vet ./...
	@echo "Code vetted"

# Run security scan (requires gosec)
security:
	@echo "Running security scan..."
	@if command -v gosec >/dev/null 2>&1; then \
		gosec ./...; \
	else \
		echo "gosec not installed. Install with: go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest"; \
	fi

# Install the binary
install: build
	@echo "Installing $(BINARY_NAME)..."
	cp $(BUILD_DIR)/$(BINARY_NAME) $(GOPATH)/bin/$(BINARY_NAME)
	@echo "Installed $(BINARY_NAME) to $(GOPATH)/bin/$(BINARY_NAME)"

# Create release builds for multiple platforms
release:
	@echo "Creating release builds..."
	@mkdir -p $(BUILD_DIR)/release
	
	# Linux AMD64
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/release/$(BINARY_NAME)-linux-amd64 .
	
	# Linux ARM64
	GOOS=linux GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/release/$(BINARY_NAME)-linux-arm64 .
	
	# macOS AMD64
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/release/$(BINARY_NAME)-darwin-amd64 .
	
	# macOS ARM64 (Apple Silicon)
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/release/$(BINARY_NAME)-darwin-arm64 .
	
	# Windows AMD64
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/release/$(BINARY_NAME)-windows-amd64.exe .
	
	@echo "Release builds created in $(BUILD_DIR)/release/"

# Quick development workflow
dev: fmt vet test-unit build

# Full CI workflow
ci: clean fmt vet lint test-coverage security build

# Benchmark tests
benchmark:
	@echo "Running benchmark tests..."
	$(GOTEST) -bench=. -benchmem ./...

# Watch for changes and run tests (requires entr)
watch:
	@echo "Watching for changes..."
	@find . -name "*.go" | entr -c make test-unit

# Generate completion scripts
completions: build
	@echo "Generating completion scripts..."
	@mkdir -p $(BUILD_DIR)/completions
	$(BUILD_DIR)/$(BINARY_NAME) completion bash > $(BUILD_DIR)/completions/$(BINARY_NAME).bash
	$(BUILD_DIR)/$(BINARY_NAME) completion zsh > $(BUILD_DIR)/completions/$(BINARY_NAME).zsh
	$(BUILD_DIR)/$(BINARY_NAME) completion fish > $(BUILD_DIR)/completions/$(BINARY_NAME).fish
	$(BUILD_DIR)/$(BINARY_NAME) completion powershell > $(BUILD_DIR)/completions/$(BINARY_NAME).ps1
	$(BUILD_DIR)/$(BINARY_NAME) completion nushell > $(BUILD_DIR)/completions/$(BINARY_NAME).nu
	@echo "Completion scripts generated in $(BUILD_DIR)/completions/"

# Help
help:
	@echo "Available targets:"
	@echo "  all            - Clean, format, test, and build"
	@echo "  build          - Build the binary"
	@echo "  clean          - Clean build artifacts"
	@echo "  test           - Run all tests"
	@echo "  test-unit      - Run unit tests only"
	@echo "  test-integration - Run integration tests only"
	@echo "  test-coverage  - Run tests with coverage report"
	@echo "  test-race      - Run tests with race detection"
	@echo "  deps           - Download dependencies"
	@echo "  tidy           - Tidy go.mod"
	@echo "  fmt            - Format code"
	@echo "  lint           - Run linter (requires golangci-lint)"
	@echo "  vet            - Vet code"
	@echo "  security       - Run security scan (requires gosec)"
	@echo "  install        - Install binary to GOPATH/bin"
	@echo "  release        - Create release builds for multiple platforms"
	@echo "  dev            - Quick development workflow (fmt, vet, test-unit, build)"
	@echo "  ci             - Full CI workflow (clean, fmt, vet, lint, test-coverage, security, build)"
	@echo "  benchmark      - Run benchmark tests"
	@echo "  watch          - Watch for changes and run tests (requires entr)"
	@echo "  completions    - Generate completion scripts"
	@echo "  help           - Show this help"