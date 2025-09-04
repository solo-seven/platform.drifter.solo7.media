.PHONY: test test-verbose test-coverage build clean run-server run-client lint fmt vet

# Build variables
BINARY_NAME=drifter-platform
CLIENT_BINARY_NAME=drifter-client
SERVER_BINARY_NAME=drifter-server
BUILD_DIR=build
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
TEST_PACKAGES=./...
COVERAGE_FILE=coverage.out
COVERAGE_HTML=coverage.html

# Default target
all: clean fmt vet lint test build

# Clean build artifacts
clean:
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)
	rm -f $(COVERAGE_FILE) $(COVERAGE_HTML)

# Format code
fmt:
	$(GOFMT) -s -w .

# Run go vet
vet:
	$(GOCMD) vet $(TEST_PACKAGES)

# Run linter
lint:
	$(GOLINT) run

# Run tests
test:
	$(GOTEST) -v $(TEST_PACKAGES)

# Run tests with coverage
test-coverage:
	$(GOTEST) -v -coverprofile=$(COVERAGE_FILE) $(TEST_PACKAGES)
	$(GOCMD) tool cover -html=$(COVERAGE_FILE) -o $(COVERAGE_HTML)

# Run tests with Ginkgo (BDD style)
test-bdd:
	ginkgo -r -v

# Build all binaries
build: build-server build-client

# Build server
build-server: proto
	mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/$(SERVER_BINARY_NAME) ./cmd/server

# Build client
build-client: proto
	mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/$(CLIENT_BINARY_NAME) ./cmd/client

# Generate protobuf files
proto: $(PROTO_DIR)/*.proto
	protoc --go_out=$(GENERATED_DIR) --go_opt=paths=source_relative \
		--go-grpc_out=$(GENERATED_DIR) --go-grpc_opt=paths=source_relative \
		$(PROTO_DIR)/*.proto

# Run server
run-server: build-server
	./$(BUILD_DIR)/$(SERVER_BINARY_NAME)

# Run client
run-client: build-client
	./$(BUILD_DIR)/$(CLIENT_BINARY_NAME)

# Install dependencies
deps:
	$(GOMOD) download
	$(GOMOD) tidy

# Install development tools
install-tools:
	$(GOGET) github.com/golangci/golangci-lint/cmd/golangci-lint
	$(GOGET) github.com/onsi/ginkgo/v2/ginkgo
	$(GOGET) google.golang.org/protobuf/cmd/protoc-gen-go
	$(GOGET) google.golang.org/grpc/cmd/protoc-gen-go-grpc

# Run integration tests
test-integration:
	$(GOTEST) -v -tags=integration ./tests/integration/...

# Run user acceptance tests
test-uat:
	$(GOTEST) -v -tags=uat ./tests/uat/...

# Development server with hot reload (requires air)
dev-server:
	air -c .air.toml

# Docker build
docker-build:
	docker build -t $(BINARY_NAME) .

# Docker run
docker-run:
	docker run -p 8080:8080 $(BINARY_NAME)

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
	@echo "  test-bdd     - Run BDD tests with Ginkgo"
	@echo "  test-integration - Run integration tests"
	@echo "  test-uat     - Run user acceptance tests"
	@echo "  build        - Build all binaries"
	@echo "  build-server - Build server binary"
	@echo "  build-client - Build client binary"
	@echo "  proto        - Generate protobuf files"
	@echo "  run-server   - Run server"
	@echo "  run-client   - Run client"
	@echo "  deps         - Download and tidy dependencies"
	@echo "  install-tools- Install development tools"
	@echo "  dev-server   - Run development server with hot reload"
	@echo "  docker-build - Build Docker image"
	@echo "  docker-run   - Run Docker container"

