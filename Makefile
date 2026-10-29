# Makefile for open-mcp-auth-proxy

# Variables
PROJECT_ROOT := $(realpath $(dir $(abspath $(lastword $(MAKEFILE_LIST)))))
BINARY_NAME := asgardeo-mcp
GO := go
GOFMT := gofmt
GOVET := go vet
GOTEST := go test
GOLINT := golangci-lint
GOCOV := go tool cover
BUILD_DIR := build

# Source files
SRC := $(shell find . -name "*.go" -not -path "./vendor/*")
PKGS := $(shell go list ./... | grep -v /vendor/)

# Set build options
BUILD_OPTS := -v

# Set test options
TEST_OPTS := -v -race

.PHONY: all clean test fmt lint vet coverage help

# Default target
all: lint test build-linux build-linux-arm build-darwin

build: clean test build-linux build-linux-arm build-darwin

# Servers to build: asgardeo-mcp, self-service-mcp, or both (default)
SERVER ?= asgardeo-mcp self-service-mcp
LDFLAGS := -ldflags "-X main.version=$(BUILD_VERSION)"

build-linux:
	mkdir -p $(BUILD_DIR)/linux
	$(foreach s,$(SERVER),GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build $(LDFLAGS) \
	-o $(BUILD_DIR)/linux/$(s) $(PROJECT_ROOT)/cmd/$(s) &&) true

build-linux-arm:
	mkdir -p $(BUILD_DIR)/linux-arm
	$(foreach s,$(SERVER),GOOS=linux GOARCH=arm CGO_ENABLED=0 go build $(LDFLAGS) \
	-o $(BUILD_DIR)/linux-arm/$(s) $(PROJECT_ROOT)/cmd/$(s) &&) true

build-darwin:
	mkdir -p $(BUILD_DIR)/darwin
	$(foreach s,$(SERVER),GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build $(LDFLAGS) \
	-o $(BUILD_DIR)/darwin/$(s) $(PROJECT_ROOT)/cmd/$(s) &&) true

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out

# Run tests
test:
	@echo "Running tests..."
	$(GOTEST) $(TEST_OPTS) ./...

# Run tests with coverage report
coverage:
	@echo "Running tests with coverage..."
	@$(GOTEST) -coverprofile=coverage.out ./...
	@$(GOCOV) -func=coverage.out
	@$(GOCOV) -html=coverage.out -o coverage.html
	@echo "Coverage report generated in coverage.html"

# Run gofmt
fmt:
	@echo "Running gofmt..."
	@$(GOFMT) -w -s $(SRC)

# Run go vet
vet:
	@echo "Running go vet..."
	@$(GOVET) ./...

# Show help
help:
	@echo "Available targets:"
	@echo "  all        : Run lint, test, and build"
	@echo "  build      : Build the servers (SERVER=asgardeo-mcp or SERVER=self-service-mcp to build one)"
	@echo "  clean      : Clean build artifacts"
	@echo "  test       : Run tests"
	@echo "  coverage   : Run tests with coverage report"
	@echo "  fmt        : Run gofmt"
	@echo "  vet        : Run go vet"
	@echo "  help       : Show this help message"