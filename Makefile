.PHONY: all clean fmt vet lint test build test-unit test-integration test-uat test-coverage build-client build-server dev-server proto

# Build variables
CLIENT_BINARY_NAME=drifter-client
SERVER_BINARY_NAME=drifter-server
BUILD_DIR=bin
PROTO_DIR=proto
GENERATED_DIR=generated

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=gofmt
GOLINT=golangci-lint

# Test parameters
UNIT_TEST_PACKAGES=./tests/unit/...
INTEGRATION_TEST_PACKAGES=./tests/integration/...
UAT_TEST_PACKAGES=./tests/uat/...
COVERAGE_FILE=$(BUILD_DIR)/coverage.out
COVERAGE_HTML=$(BUILD_DIR)/coverage.html

# Default target
all: clean fmt vet lint test build

# Clean build artifacts
clean:
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)
	rm -f $(COVERAGE_FILE) $(COVERAGE_HTML)

fmt:
	$(GOFMT) -s -w .

vet:
	$(GOCMD) vet $(TEST_PACKAGES)

lint:
	$(GOLINT) run

test-integration:
	$(GOTEST) -v -tags=integration $(INTEGRATION_TEST_PACKAGES)

test-uat:
	$(GOTEST) -v -tags=uat $(UAT_TEST_PACKAGES)

test-unit:
	$(GOTEST) -v $(UNIT_TEST_PACKAGES)

test: deps test-unit test-integration test-uat

test-coverage:
	$(GOTEST) -v -coverprofile=$(COVERAGE_FILE) $(UNIT_TEST_PACKAGES)
	$(GOCMD) tool cover -html=$(COVERAGE_FILE) -o $(COVERAGE_HTML)

build: build-server build-client

build-server: deps proto
	mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/$(SERVER_BINARY_NAME) ./cmd/server

build-client: deps proto
	mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/$(CLIENT_BINARY_NAME) ./cmd/client

proto: install-tools
	rm -rf $(GENERATED_DIR)
	mkdir $(GENERATED_DIR)
	protoc --go_out=$(GENERATED_DIR) --go_opt=paths=source_relative \
		--go-grpc_out=$(GENERATED_DIR) --go-grpc_opt=paths=source_relative \
		$(PROTO_DIR)/*.proto

deps:
	$(GOMOD) download
	$(GOMOD) tidy

# Install development tools
install-tools:
	$(GOGET) github.com/golangci/golangci-lint/cmd/golangci-lint
	$(GOGET) github.com/onsi/ginkgo/v2/ginkgo
	$(GOGET) google.golang.org/protobuf/cmd/protoc-gen-go
	$(GOGET) google.golang.org/grpc/cmd/protoc-gen-go-grpc

# Development server with hot reload (requires air)
dev-server:
	air -c .air.toml

# Help
help:
	@echo "Available targets:"
	@echo "  all          - Clean, format, vet, lint, test, and build"
	@echo "  clean        - Remove build artifacts"
	@echo "  fmt          - Format code"
	@echo "  vet          - Run go vet"
	@echo "  lint         - Run linter"
	@echo "  test         - Run tests"
	@echo "  test-coverage- Run tests with coverage report"
	@echo "  test-integration - Run integration tests"
	@echo "  test-uat     - Run user acceptance tests"
	@echo "  build        - Build all binaries"
	@echo "  build-server - Build server binary"
	@echo "  build-client - Build client binary"
	@echo "  proto        - Generate protobuf files"
	@echo "  deps         - Download and tidy dependencies"
	@echo "  install-tools- Install development tools"
	@echo "  dev-server   - Run development server with hot reload"

