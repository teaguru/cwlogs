# Build variables
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
LDFLAGS := -X main.Version=$(VERSION)
BINARY_NAME := cwlogs

# Build targets
.PHONY: build
build:
	@echo "Building $(BINARY_NAME) version $(VERSION)..."
	go build -ldflags "$(LDFLAGS)" -o $(BINARY_NAME)

.PHONY: build-all
build-all:
	@echo "Building for all platforms..."
	GOOS=darwin GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(BINARY_NAME)-darwin-amd64
	GOOS=darwin GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o $(BINARY_NAME)-darwin-arm64
	GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(BINARY_NAME)-linux-amd64
	GOOS=linux GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o $(BINARY_NAME)-linux-arm64

.PHONY: install
install: build
	@echo "Installing $(BINARY_NAME) to /usr/local/bin..."
	cp $(BINARY_NAME) /usr/local/bin/$(BINARY_NAME)

# Test targets
.PHONY: test
test:
	@echo "Running tests..."
	go test -v -race -coverprofile=coverage.out ./...

.PHONY: test-coverage
test-coverage: test
	@echo "Generating coverage report..."
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# Lint targets
.PHONY: lint
lint:
	@echo "Running linters..."
	go vet ./...
	go fmt ./...
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not installed, skipping..."; \
	fi

.PHONY: fmt
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Release targets
.PHONY: release
release:
	@echo "Creating release archives..."
	@mkdir -p dist
	@$(MAKE) build-all
	tar -czf dist/$(BINARY_NAME)-$(VERSION)-darwin-amd64.tar.gz $(BINARY_NAME)-darwin-amd64
	tar -czf dist/$(BINARY_NAME)-$(VERSION)-darwin-arm64.tar.gz $(BINARY_NAME)-darwin-arm64
	tar -czf dist/$(BINARY_NAME)-$(VERSION)-linux-amd64.tar.gz $(BINARY_NAME)-linux-amd64
	tar -czf dist/$(BINARY_NAME)-$(VERSION)-linux-arm64.tar.gz $(BINARY_NAME)-linux-arm64
	@echo "Computing checksums..."
	cd dist && shasum -a 256 *.tar.gz > checksums.txt
	@echo "Release artifacts in dist/"

# GoReleaser targets
.PHONY: goreleaser-check
goreleaser-check:
	@if ! command -v goreleaser >/dev/null 2>&1; then \
		echo "GoReleaser not installed. Install with:"; \
		echo "  brew install goreleaser/tap/goreleaser"; \
		echo "  or go install github.com/goreleaser/goreleaser@latest"; \
		exit 1; \
	fi

.PHONY: release-dry-run
release-dry-run: goreleaser-check
	@echo "Running GoReleaser dry run..."
	goreleaser release --snapshot --clean --skip=publish

.PHONY: release-snapshot
release-snapshot: goreleaser-check
	@echo "Creating snapshot release..."
	goreleaser release --snapshot --clean

.PHONY: release-local
release-local: goreleaser-check
	@echo "Building local release..."
	goreleaser build --clean

# Clean targets
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_NAME)-*
	rm -rf dist/
	rm -f coverage.out coverage.html

# Development targets
.PHONY: run
run: build
	./$(BINARY_NAME)

.PHONY: deps
deps:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

# Help target
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  build          - Build the binary"
	@echo "  build-all      - Build for all platforms (darwin/linux, amd64/arm64)"
	@echo "  install        - Install binary to /usr/local/bin"
	@echo "  test           - Run tests with race detector"
	@echo "  test-coverage  - Run tests and generate HTML coverage report"
	@echo "  lint           - Run linters (vet, fmt, golangci-lint)"
	@echo "  fmt            - Format code"
	@echo "  release        - Build release archives and checksums (manual)"
	@echo "  release-dry-run - Test GoReleaser configuration"
	@echo "  release-snapshot - Create snapshot release with GoReleaser"
	@echo "  release-local  - Build local release with GoReleaser"
	@echo "  clean          - Remove build artifacts"
	@echo "  run            - Build and run the application"
	@echo "  deps           - Download and tidy dependencies"
	@echo "  help           - Show this help message"
	@echo ""
	@echo "Variables:"
	@echo "  VERSION        - Version string (default: git describe or 'dev')"

.DEFAULT_GOAL := build
