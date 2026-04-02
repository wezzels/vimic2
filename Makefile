# Vimic2 Makefile

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Binary names
BINARY_NAME=vimic2
BINARY_UNIX=$(BINARY_NAME)_unix
BINARY_WINDOWS=$(BINARY_NAME).exe
BINARY_DARWIN=$(BINARY_NAME)_darwin

# Build parameters
VERSION=$(shell git describe --tags --always --dirty)
BUILD_TIME=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME)"

# Directories
BUILD_DIR=build
COVERAGE_DIR=coverage

# Platforms
PLATFORMS=darwin/amd64 darwin/arm64 linux/amd64 linux/arm64 windows/amd64

.PHONY: all build clean test coverage lint fmt vet help install dep

all: build

## build: Build the binary for current platform
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/vimic2
	@echo "Built: $(BUILD_DIR)/$(BINARY_NAME)"

## build-all: Build for all platforms
build-all:
	@echo "Building for all platforms..."
	@mkdir -p $(BUILD_DIR)
	@for platform in $(PLATFORMS); do \
		IFS='/' read -r GOOS GOARCH <<< "$$platform"; \
		OUTPUT=$(BUILD_DIR)/$(BINARY_NAME)-$$GOOS-$$GOARCH; \
		if [ "$$GOOS" = "windows" ]; then \
			OUTPUT=$$OUTPUT.exe; \
		fi; \
		echo "Building for $$GOOS/$$GOARCH..."; \
		GOOS=$$GOOS GOARCH=$$GOARCH $(GOBUILD) $(LDFLAGS) -o $$OUTPUT ./cmd/vimic2; \
	done
	@echo "All platforms built in $(BUILD_DIR)/"

## clean: Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@rm -rf $(COVERAGE_DIR)
	$(GOCLEAN)

## test: Run unit tests
test:
	@echo "Running tests..."
	$(GOTEST) -v -tags='!integration' ./...

## test-integration: Run integration tests
test-integration:
	@echo "Running integration tests..."
	$(GOTEST) -v -tags=integration ./...

## test-all: Run all tests
test-all:
	@echo "Running all tests..."
	$(GOTEST) -v ./...

## coverage: Run tests with coverage
coverage:
	@echo "Running tests with coverage..."
	@mkdir -p $(COVERAGE_DIR)
	$(GOTEST) -tags='!integration' -coverprofile=$(COVERAGE_DIR)/coverage.out ./...
	$(GOCMD) tool cover -html=$(COVERAGE_DIR)/coverage.out -o $(COVERAGE_DIR)/coverage.html
	@echo "Coverage report: $(COVERAGE_DIR)/coverage.html"

## coverage-report: Show coverage summary
coverage-report:
	$(GOTEST) -tags='!integration' -coverprofile=$(COVERAGE_DIR)/coverage.out ./...
	$(GOCMD) tool cover -func=$(COVERAGE_DIR)/coverage.out

## lint: Run linters
lint:
	@echo "Running linters..."
	@which golangci-lint > /dev/null || go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	golangci-lint run ./...

## fmt: Format code
fmt:
	@echo "Formatting code..."
	$(GOCMD) fmt ./...

## vet: Run go vet
vet:
	@echo "Running go vet..."
	$(GOCMD) vet ./...

## install: Install binary to $GOPATH/bin
install:
	@echo "Installing $(BINARY_NAME)..."
	$(GOBUILD) $(LDFLAGS) -o $(GOPATH)/bin/$(BINARY_NAME) ./cmd/vimic2

## dep: Download dependencies
dep:
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) verify

## dep-update: Update dependencies
dep-update:
	@echo "Updating dependencies..."
	$(GOMOD) tidy
	$(GOMOD) verify

## docker-build: Build Docker image
docker-build:
	@echo "Building Docker image..."
	docker build -t $(BINARY_NAME):$(VERSION) .

## docker-run: Run Docker container
docker-run:
	@echo "Running Docker container..."
	docker run -it --rm $(BINARY_NAME):$(VERSION)

## release: Create release binaries
release: build-all checksums

## checksums: Create checksums for release binaries
checksums:
	@echo "Creating checksums..."
	@cd $(BUILD_DIR) && sha256sum $(BINARY_NAME)* > checksums.sha256
	@echo "Checksums created: $(BUILD_DIR)/checksums.sha256"

## templates: Create VM templates
templates:
	@echo "Creating VM templates..."
	@./scripts/create-templates.sh all

## templates-base: Create base template only
templates-base:
	@./scripts/create-templates.sh base

## templates-go: Create Go template
templates-go:
	@./scripts/create-templates.sh go

## templates-node: Create Node template
templates-node:
	@./scripts/create-templates.sh node

## templates-docker: Create Docker template
templates-docker:
	@./scripts/create-templates.sh docker

## templates-jenkins: Create Jenkins template
templates-jenkins:
	@./scripts/create-templates.sh jenkins

## templates-verify: Verify templates
templates-verify:
	@./scripts/create-templates.sh verify

## run: Build and run locally
run: build
	@./$(BUILD_DIR)/$(BINARY_NAME)

## dev: Run with hot reload (requires air)
dev:
	@which air > /dev/null || go install github.com/cosmtrek/air@latest
	air

## help: Show this help
help:
	@echo "Vimic2 Makefile"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@sed -n 's/^## //p' $(MAKEFILE_LIST) | column -t -s ':'

# Default target
.DEFAULT_GOAL := build