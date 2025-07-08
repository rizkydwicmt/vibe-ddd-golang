.PHONY: build run test clean lint format deps docker-build docker-run

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
BINARY_NAME=vibe-ddd-golang
BINARY_PATH=./bin/$(BINARY_NAME)

# Build the API api
build:
	$(GOBUILD) -o $(BINARY_PATH) -v ./cmd/api

# Build the worker api
build-worker:
	$(GOBUILD) -o ./bin/worker -v ./cmd/worker

# Build the migration api
build-migration:
	$(GOBUILD) -o ./bin/migration -v ./cmd/migration

# Build the gRPC api
build-grpc:
	$(GOBUILD) -o ./bin/grpc -v ./cmd/grpc

# Build all servers
build-all: build build-worker build-migration build-grpc

# Proto generation commands
proto-gen:
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative --go-grpc_opt=require_unimplemented_servers=false api/proto/user/user.proto
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative --go-grpc_opt=require_unimplemented_servers=false api/proto/payment/payment.proto

# Clean generated proto files
proto-clean:
	rm -f api/proto/user/user.pb.go api/proto/user/user_grpc.pb.go
	rm -f api/proto/payment/payment.pb.go api/proto/payment/payment_grpc.pb.go

# Install proto tools
proto-tools:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.3.0

# Swagger/OpenAPI commands
swagger-gen:
	swag init -g cmd/api/main.go -o docs

# Clean generated swagger files
swagger-clean:
	rm -rf docs/

# Install swagger tools
swagger-tools:
	go install github.com/swaggo/swag/cmd/swag@latest

# Run the API api
run:
	$(GOCMD) run ./cmd/api

# Run the worker api
run-worker:
	$(GOCMD) run ./cmd/worker

# Run database migrations
run-migration:
	$(GOCMD) run ./cmd/migration -action=migrate

# Run database seeding
run-seed:
	$(GOCMD) run ./cmd/migration -action=seed

# Drop database tables
run-drop:
	$(GOCMD) run ./cmd/migration -action=drop

# Run the gRPC api
run-grpc:
	$(GOCMD) run ./cmd/grpc -port=9090

# Run all tests
test:
	$(GOTEST) -v -race -timeout 30s ./...

# Run tests with coverage
test-coverage:
	$(GOTEST) -v -race -coverprofile=coverage.out -covermode=atomic ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run unit tests only (internal packages)
test-unit:
	$(GOTEST) -v -timeout 30s ./internal/...

# Run integration tests only
test-integration:
	$(GOTEST) -v -race -timeout 30s ./test/...

# Run tests for specific layers
test-repo:
	$(GOTEST) -v -race -timeout 30s ./internal/application/*/repository/...

test-service:
	$(GOTEST) -v -race -timeout 30s ./internal/application/*/service/...

test-handler:
	$(GOTEST) -v -race -timeout 30s ./internal/application/*/handler/...

test-worker:
	$(GOTEST) -v -race -timeout 30s ./internal/application/payment/worker/...

# Run tests for specific domains
test-user:
	$(GOTEST) -v -race -timeout 30s ./internal/application/user/...

test-payment:
	$(GOTEST) -v -race -timeout 30s ./internal/application/payment/...

# Run tests with verbose output and no cache
test-verbose:
	$(GOTEST) -v -race -count=1 -timeout 30s ./...

# Clean build artifacts
clean:
	$(GOCLEAN)
	rm -rf ./bin
	rm -f coverage.out coverage.html

# Install dependencies
deps:
	$(GOMOD) download
	$(GOMOD) tidy

# Linting using golangci-lint (includes nil pointer detection)
lint:
	golangci-lint run

# Lint with auto-fix
lint-fix:
	golangci-lint run --fix

# Lint verbose output
lint-verbose:
	golangci-lint run --verbose

# Show which nil detection linters are enabled
lint-nil-info:
	@echo "Nil detection linters enabled:"
	@echo "  - nilerr: Finds code that returns nil even if it checks that error is not nil"
	@echo "  - nilnil: Checks that there is no simultaneous return of nil error and invalid value"
	@echo ""
	@echo "Run 'make lint' to detect potential nil pointer issues"

# Run specific linter
lint-linter:
	@if [ -z "$(LINTER)" ]; then \
		echo "Usage: make lint-linter LINTER=errcheck"; \
		exit 1; \
	fi
	golangci-lint run --disable-all --enable=$(LINTER)

# Lint only new/changed files
lint-new:
	golangci-lint run --new-from-rev=HEAD~1

# Format the code
format:
	$(GOCMD) fmt ./...
	gofumpt -l -w .
	goimports -w .

# Format with gofumpt (stricter formatting)
format-strict:
	gofumpt -l -w .
	goimports -w .
	gci write --skip-generated -s standard -s default -s "prefix(vibe-ddd-golang)" .

# Install development tools
tools:
	@echo "Installing development tools..."
	# Install golangci-lint
	@if ! command -v golangci-lint >/dev/null 2>&1; then \
		echo "Installing golangci-lint..."; \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v1.54.2; \
	else \
		echo "golangci-lint already installed"; \
	fi

# Docker build
docker-build:
	docker build -t $(BINARY_NAME) .

# Docker run
docker-run:
	docker run -p 8080:8080 $(BINARY_NAME)

# Development setup
dev-setup: tools deps
	@echo "Development environment setup complete"

# Comprehensive code quality check
quality:
	@echo "Running comprehensive code quality checks..."
	make format
	make lint
	make test
	@echo "Code quality checks completed!"

# Pre-commit checks
pre-commit:
	@echo "Running pre-commit checks..."
	golangci-lint run --fix
	make test-unit
	@echo "Pre-commit checks passed!"

# Install pre-commit hooks
install-hooks:
	@if [ -f scripts/install-pre-commit.sh ]; then \
		./scripts/install-pre-commit.sh; \
	else \
		echo "Error: scripts/install-pre-commit.sh not found"; \
		exit 1; \
	fi

# CI checks (for continuous integration)
ci:
	@echo "Running CI checks..."
	make lint
	make test-coverage
	make build-all
	@echo "CI checks completed!"

# Help
help:
	@echo "Available targets:"
	@echo ""
	@echo "Build Commands:"
	@echo "  build         - Build the API server"
	@echo "  build-worker  - Build the worker server"
	@echo "  build-migration - Build the migration server"
	@echo "  build-grpc    - Build the gRPC server"
	@echo "  build-all     - Build all servers"
	@echo ""
	@echo "Run Commands:"
	@echo "  run           - Run the API server"
	@echo "  run-worker    - Run the worker server"
	@echo "  run-migration - Run database migrations"
	@echo "  run-seed      - Run database seeding"
	@echo "  run-drop      - Drop database tables"
	@echo "  run-grpc      - Run the gRPC server"
	@echo ""
	@echo "Test Commands:"
	@echo "  test          - Run all tests"
	@echo "  test-coverage - Run tests with coverage"
	@echo "  test-unit     - Run unit tests only"
	@echo "  test-integration - Run integration tests only"
	@echo "  test-repo     - Run repository layer tests"
	@echo "  test-service  - Run service layer tests"
	@echo "  test-handler  - Run handler layer tests"
	@echo "  test-worker   - Run worker layer tests"
	@echo "  test-user     - Run user domain tests"
	@echo "  test-payment  - Run payment domain tests"
	@echo "  test-verbose  - Run tests with verbose output"
	@echo ""
	@echo "Development Commands:"
	@echo "  clean         - Clean build artifacts"
	@echo "  deps          - Install dependencies"
	@echo "  format        - Format the code"
	@echo "  format-strict - Format with stricter rules"
	@echo "  tools         - Install development tools"
	@echo "  dev-setup     - Setup development environment"
	@echo ""
	@echo "Linting Commands:"
	@echo "  lint          - Run golangci-lint (includes nil detection)"
	@echo "  lint-fix      - Run golangci-lint with auto-fix"
	@echo "  lint-verbose  - Run golangci-lint with verbose output"
	@echo "  lint-nil-info - Show enabled nil detection linters"
	@echo "  lint-new      - Lint only new/changed code"
	@echo "  lint-linter   - Run specific linter (LINTER=name)"
	@echo ""
	@echo "Quality Commands:"
	@echo "  quality       - Run comprehensive quality checks"
	@echo "  pre-commit    - Run pre-commit checks"
	@echo "  install-hooks - Install pre-commit hooks"
	@echo "  ci            - Run CI checks"
	@echo ""
	@echo "Proto Commands:"
	@echo "  proto-gen     - Generate gRPC code from proto files"
	@echo "  proto-clean   - Clean generated proto files"
	@echo "  proto-tools   - Install proto generation tools"
	@echo ""
	@echo "Swagger Commands:"
	@echo "  swagger-gen   - Generate Swagger/OpenAPI documentation"
	@echo "  swagger-clean - Clean generated swagger files"
	@echo "  swagger-tools - Install swagger generation tools"
	@echo ""
	@echo "Docker Commands:"
	@echo "  docker-build  - Build Docker image"
	@echo "  docker-run    - Run Docker container"
	@echo ""
	@echo "Other:"
	@echo "  help          - Show this help"