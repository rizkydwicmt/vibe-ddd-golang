.PHONY: build build-migration build-all run run-migration \
        migrate-status migrate-apply migrate-rollback migrate-init migrate-diff \
        test test-coverage test-unit test-integration test-repo test-service test-handler \
        test-user test-payment test-verbose clean deps lint lint-fix lint-verbose lint-new \
        lint-linter format format-strict tools dev-setup quality pre-commit install-hooks ci \
        swagger-gen swagger-clean swagger-tools proto proto-tools agentic-check docker-build docker-run help

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
BINARY_NAME=vibe-ddd-golang
BINARY_PATH=./bin/$(BINARY_NAME)

# Build the API server
build:
	CGO_ENABLED=0 $(GOBUILD) -o $(BINARY_PATH) -v ./cmd/api

# Build the migration CLI
build-migration:
	$(GOBUILD) -o ./bin/migration -v ./cmd/migration

# Build all binaries
build-all: build build-migration

# Run the API server
run:
	$(GOCMD) run ./cmd/api

# ---- Atlas migrations (cmd/migration) ----
# Show migration status
run-migration migrate-status:
	$(GOCMD) run ./cmd/migration -status

# Apply pending migrations
migrate-apply:
	$(GOCMD) run ./cmd/migration -apply

# Rollback the last applied migration (or VERSION=... to a target)
migrate-rollback:
	$(GOCMD) run ./cmd/migration -rollback $(if $(VERSION),-version=$(VERSION),)

# Generate the initial migration: make migrate-init NAME=init_schema
migrate-init:
	$(GOCMD) run ./cmd/migration -init -name=$(NAME)

# Diff entities against configured DB; DEV_DSN optionally overrides the connection.
migrate-diff:
	$(GOCMD) run ./cmd/migration -diff -name=$(NAME) $(if $(DEV_DSN),-dev='$(DEV_DSN)',)

# Swagger/OpenAPI
swagger-gen:
	swag init -g cmd/api/main.go -o internal/server/api/docs

swagger-clean:
	rm -rf internal/server/api/docs/

swagger-tools:
	go install github.com/swaggo/swag/cmd/swag@latest

proto:
	protoc --proto_path=internal/server/grpc/proto \
		--go_out=. --go_opt=module=vibe-ddd-golang \
		--go-grpc_out=. --go-grpc_opt=module=vibe-ddd-golang \
		internal/server/grpc/proto/user/user.proto \
		internal/server/grpc/proto/payment/payment.proto

proto-tools:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Agentic tooling documentation integrity
agentic-check:
	./scripts/agentic-docs-check.sh

# ---- Tests ----
test:
	$(GOTEST) -race -cover -timeout 60s ./...

test-coverage:
	$(GOTEST) -race -coverprofile=coverage.out -covermode=atomic ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

test-unit:
	$(GOTEST) -timeout 60s ./internal/...

test-integration:
	$(GOTEST) -race -timeout 60s ./test/...

test-repo:
	$(GOTEST) -race -timeout 60s ./internal/application/*/repository/...

test-service:
	$(GOTEST) -race -timeout 60s ./internal/application/*/service/...

test-handler:
	$(GOTEST) -race -timeout 60s ./internal/application/*/handler/...

test-user:
	$(GOTEST) -race -timeout 60s ./internal/application/user/...

test-payment:
	$(GOTEST) -race -timeout 60s ./internal/application/payment/...

test-verbose:
	$(GOTEST) -v -race -count=1 -timeout 60s ./...

# ---- Housekeeping ----
clean:
	$(GOCLEAN)
	rm -rf ./bin
	rm -f coverage.out coverage.html

deps:
	$(GOMOD) download
	$(GOMOD) tidy

lint:
	golangci-lint run

lint-fix:
	golangci-lint run --fix

lint-verbose:
	golangci-lint run --verbose

lint-new:
	golangci-lint run --new-from-rev=HEAD~1

lint-linter:
	@if [ -z "$(LINTER)" ]; then echo "Usage: make lint-linter LINTER=errcheck"; exit 1; fi
	golangci-lint run --disable-all --enable=$(LINTER)

format:
	$(GOCMD) fmt ./...
	gofumpt -l -w .
	goimports -w .

format-strict:
	gofumpt -l -w .
	goimports -w .
	gci write --skip-generated -s standard -s default -s "prefix(vibe-ddd-golang)" .

tools:
	@if ! command -v golangci-lint >/dev/null 2>&1; then \
		echo "Installing golangci-lint..."; \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v1.54.2; \
	else echo "golangci-lint already installed"; fi

docker-build:
	docker build -t $(BINARY_NAME) .

docker-run:
	docker run -p 8080:8080 $(BINARY_NAME)

dev-setup: tools deps
	@echo "Development environment setup complete"

quality:
	make format
	make lint
	make test
	@echo "Code quality checks completed!"

pre-commit:
	golangci-lint run --fix
	make test-unit
	@echo "Pre-commit checks passed!"

install-hooks:
	@if [ -f scripts/install-pre-commit.sh ]; then ./scripts/install-pre-commit.sh; \
	else echo "Error: scripts/install-pre-commit.sh not found"; exit 1; fi

ci:
	make lint
	make test-coverage
	make build-all
	@echo "CI checks completed!"

help:
	@echo "Build:      build build-migration build-all"
	@echo "Run:        run"
	@echo "Migrate:    migrate-status migrate-apply migrate-rollback migrate-init(NAME=) migrate-diff(NAME= [DEV_DSN=])"
	@echo "Test:       test test-coverage test-unit test-integration test-repo test-service test-handler test-user test-payment test-verbose"
	@echo "Lint:       lint lint-fix lint-verbose lint-new lint-linter(LINTER=)"
	@echo "Format:     format format-strict"
	@echo "Swagger:    swagger-gen swagger-clean swagger-tools"
	@echo "Proto:      proto proto-tools"
	@echo "Quality:    quality pre-commit ci agentic-check"
	@echo "Docker:     docker-build docker-run"
	@echo "Other:      clean deps tools dev-setup help"
