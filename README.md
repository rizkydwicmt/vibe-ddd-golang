# Vibe DDD Golang

A production-ready Go boilerplate following **Domain-Driven Design (DDD)** principles with **NestJS-like architecture patterns**. Built with modern Go practices, microservice architecture, and comprehensive background job processing.

## 🏗️ Architecture Overview

This project implements a **clean, domain-driven architecture** where each domain is a complete vertical slice owning its entire lifecycle - from HTTP endpoints to background processing.

### Core Principles

- **Domain-Driven Design**: Each domain is self-contained and independently deployable
- **Clean Architecture**: Clear separation of concerns with dependency inversion
- **NestJS-like Modules**: Familiar patterns for developers coming from Node.js/NestJS
- **Microservice Ready**: Separate deployable API and Worker servers
- **Production Ready**: Comprehensive logging, graceful shutdown, and error handling

## 📁 Project Structure

```
vibe-ddd-golang/
├── cmd/                                  # Application entry points
│   ├── api/main.go                       # API server startup
│   ├── worker/main.go                    # Worker server startup
│   ├── migration/main.go                 # Database migration server
│   └── grpc/main.go                      # gRPC server startup
├── internal/                             # Private application code
│   ├── application/                      # Domain layer (DDD)
│   │   ├── payment/                      # Payment domain
│   │   │   ├── dto/                      # Data Transfer Objects
│   │   │   │   └── payment.dto.go        # Request/response models
│   │   │   ├── entity/                   # Domain entities
│   │   │   │   └── payment.entity.go     # Database models
│   │   │   ├── repository/               # Data access layer
│   │   │   │   └── payment.repo.go       # Repository implementation
│   │   │   ├── service/                  # Business logic layer
│   │   │   │   └── payment.service.go    # Domain services
│   │   │   ├── handler/                  # HTTP layer
│   │   │   │   └── payment.handler.go    # REST endpoints + routes
│   │   │   ├── worker/                   # Background processing
│   │   │   │   ├── handler.go            # Job handlers
│   │   │   │   └── tasks.go              # Job definitions
│   │   │   └── module.go                 # Domain DI configuration
│   │   └── user/                         # User domain
│   │       ├── dto/user.dto.go           # User DTOs
│   │       ├── entity/user.entity.go     # User entity
│   │       ├── repository/user.repo.go   # User repository
│   │       ├── service/user.service.go   # User services
│   │       ├── handler/user.handler.go   # User endpoints
│   │       └── module.go                 # User DI config
│   ├── server/                           # Server implementations
│   │   ├── api/                          # HTTP API server
│   │   │   ├── module.go                 # Route registration & setup
│   │   │   └── providers.go              # API server DI providers
│   │   ├── worker/                       # Background worker server
│   │   │   ├── module.go                 # Worker handler registration
│   │   │   └── providers.go              # Worker server DI providers
│   │   ├── migration/                    # Database migration server
│   │   │   ├── module.go                 # Migration operations
│   │   │   └── providers.go              # Migration DI providers
│   │   └── grpc/                         # gRPC server
│   │       ├── module.go                 # gRPC service registration
│   │       └── providers.go              # gRPC server DI providers
│   ├── middleware/                       # HTTP middleware
│   │   └── middleware.go                 # Logging, CORS, recovery
│   ├── config/                           # Configuration
│   │   └── config.go                     # App configuration
│   └── pkg/                              # Internal packages
│       ├── database/database.go          # DB connection
│       ├── logger/logger.go              # Structured logging
│       ├── queue/                        # Job queue infrastructure
│       │   ├── client.go                 # Redis queue client
│       │   ├── server.go                 # Worker server
│       │   └── logger.go                 # Queue logging
│       └── testutil/                     # Test utilities
│           ├── database.go               # Test database setup
│           ├── fixtures.go               # Test data fixtures
│           ├── logger.go                 # Test logger setup
│           └── mocks.go                  # Mock implementations
├── config.yaml                           # Configuration file
├── Makefile                              # Build automation
├── Dockerfile                            # Container image
├── go.mod                                # Go modules
└── README.md                             # This file
```

## 🎯 Domain-Driven Design Implementation

### Domain Structure Pattern

Each domain follows the same consistent pattern:

```
internal/application/{domain}/
├── dto/              # Data Transfer Objects
├── entity/           # Domain entities (database models)
├── repository/       # Data access interfaces & implementations
├── service/          # Business logic & domain services
├── handler/          # HTTP handlers & route registration
├── worker/           # Background job processing (optional)
└── module.go         # Dependency injection configuration
```

### Key Design Principles

1. **Domain Ownership**: Each domain owns its complete vertical slice
2. **Dependency Inversion**: High-level modules don't depend on low-level modules
3. **Single Responsibility**: Each layer has a clear, focused responsibility
4. **Interface Segregation**: Small, focused interfaces
5. **Separation of Concerns**: HTTP, business logic, and data access are separated

### Layer Responsibilities

| Layer | Responsibility | Example |
|-------|---------------|---------|
| **Handler** | HTTP concerns, routing, request/response | `payment.handler.go` |
| **Service** | Business logic, validation, orchestration | `payment.service.go` |
| **Repository** | Data access, database operations | `payment.repo.go` |
| **Entity** | Domain models, business rules | `payment.entity.go` |
| **DTO** | Data transfer, validation, serialization | `payment.dto.go` |
| **Worker** | Background processing, async jobs | `payment/worker/` |

## 🚀 Quick Start

### Prerequisites

- Go 1.21+
- Redis 6.0+ (for background jobs)
- PostgreSQL 12+ (for database)

### Installation

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd vibe-ddd-golang
   ```

2. **Install dependencies**
   ```bash
   make deps
   ```

3. **Setup configuration**
   ```bash
   cp config.sample.yaml config.yaml
   # Edit config.yaml with your settings
   ```

4. **Start required services**
   ```bash
   # Start Redis (for background jobs)
   redis-api
   
   # Start PostgreSQL (for database)
   # Use your preferred method
   ```

### Running the Application

#### Development Mode

```bash
# Terminal 1: Start API api
make run

# Terminal 2: Start worker api (for background jobs)
make run-worker
```

#### Production Mode

```bash
# Build both servers
make build
make build-worker

# Run API api
./bin/vibe-ddd-golang

# Run worker api
./bin/worker
```

### Docker Deployment

```bash
# Build and run with Docker
make docker-build
make docker-run
```

## 🛠️ Available Commands

For a complete list of all available commands, run:
```bash
make help
```

### Build Commands

```bash
make build            # Build API api
make build-worker     # Build worker api
make build-migration  # Build migration api
make build-grpc       # Build gRPC api
make build-all        # Build all servers
```

### Run Commands

```bash
make run              # Run API api
make run-worker       # Run worker api
make run-grpc         # Run gRPC api
make run-migration    # Run database migrations
make run-seed         # Seed database with initial data
make run-drop         # Drop all database tables
```

### Test Commands

```bash
make test             # Run all tests
make test-coverage    # Run tests with coverage report
make test-unit        # Run unit tests only
make test-integration # Run integration tests only
make test-repo        # Run repository layer tests
make test-service     # Run service layer tests
make test-handler     # Run handler layer tests
make test-worker      # Run worker layer tests
make test-user        # Run user domain tests
make test-payment     # Run payment domain tests
make test-verbose     # Run tests with verbose output
```

### Code Quality & Linting

```bash
make lint             # Run golangci-lint (includes nil detection)
make lint-fix         # Run golangci-lint with auto-fix
make lint-verbose     # Run golangci-lint with verbose output
make lint-new         # Lint only new/changed files
make lint-linter      # Run specific linter (LINTER=name)
make lint-nil-info    # Show enabled nil detection linters
make format           # Format code with go fmt
make format-strict    # Format with stricter rules (gofumpt + goimports)
```

### Development Tools

```bash
make tools            # Install development tools (golangci-lint, etc.)
make dev-setup        # Setup complete development environment
make deps             # Install and tidy dependencies
make clean            # Clean build artifacts
```

### Quality & CI

```bash
make quality          # Run comprehensive quality checks
make pre-commit       # Run pre-commit checks
make install-hooks    # Install pre-commit hooks
make ci               # Run CI checks (linting, tests, build)
```

### Proto Generation

```bash
make proto-gen        # Generate gRPC code from proto files
make proto-clean      # Clean generated proto files
make proto-tools      # Install proto generation tools
```

### Docker

```bash
make docker-build     # Build Docker image
make docker-run       # Run Docker container
```

## 📚 API Documentation

The API is fully documented using **Swagger/OpenAPI 2.0** with interactive documentation available at runtime.

### Swagger UI

When the server is running, you can access the interactive API documentation at:

- **Swagger UI**: `http://localhost:8080/swagger/index.html`
- **Redirect endpoint**: `http://localhost:8080/docs` (redirects to Swagger UI)
- **OpenAPI JSON**: `http://localhost:8080/swagger/doc.json`

### Generating Documentation

```bash
# Generate swagger documentation from code annotations
make swagger-gen

# Clean generated swagger files
make swagger-clean

# Install swagger tools
make swagger-tools
```

### Base URL
```
http://localhost:8080/api/v1
```

### Health Check
```http
GET /health        # Server health status
GET /health/ready  # Server readiness check
```

### User Management
```http
POST   /users                    # Create user
GET    /users                    # List users (with pagination & filtering)
GET    /users/:id                # Get user by ID
PUT    /users/:id                # Update user
DELETE /users/:id                # Delete user
PUT    /users/:id/password       # Update password
```

### Payment Management
```http
POST   /payments                 # Create payment
GET    /payments                 # List payments (with filtering & pagination)
GET    /payments/:id             # Get payment by ID
PUT    /payments/:id             # Update payment
DELETE /payments/:id             # Delete payment
GET    /users/:user_id/payments  # Get user payments
```

### API Features

- **OpenAPI/Swagger Documentation**: Interactive API docs with try-it-out functionality
- **Request/Response Validation**: Automatic validation using struct tags
- **Error Handling**: Consistent error responses across all endpoints
- **Filtering & Pagination**: Query parameter support for list endpoints
- **Content Negotiation**: JSON request/response format
- **Status Codes**: RESTful HTTP status codes

### Request Examples

#### Create User
```bash
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Doe",
    "email": "john@example.com",
    "password": "securepassword"
  }'
```

#### Create Payment
```bash
curl -X POST http://localhost:8080/api/v1/payments \
  -H "Content-Type: application/json" \
  -d '{
    "amount": 100.00,
    "currency": "USD",
    "description": "Test payment",
    "user_id": 1
  }'
```

## 🚀 Server Architecture

The application follows a **multi-server architecture** where different concerns are separated into independent, deployable servers:

### Available Servers

| Server | Purpose | Entry Point | Default Port |
|--------|---------|-------------|--------------|
| **API Server** | HTTP REST API | `main.go` | 8080 |
| **Worker Server** | Background job processing | `cmd/worker/main.go` | - |
| **Migration Server** | Database operations | `cmd/migration/main.go` | - |
| **gRPC Server** | gRPC services (User & Payment) | `cmd/grpc/main.go` | 9090 |

### Building & Running Servers

```bash
# Build all servers
make build-all

# Or build individually
make build          # API api
make build-worker   # Worker api
make build-migration # Migration api
make build-grpc     # gRPC api

# Run servers
make run            # Start API api
make run-worker     # Start worker api
make run-grpc       # Start gRPC api

# Database operations
make run-migration  # Run migrations
make run-seed      # Seed initial data
make run-drop      # Drop all tables

# Proto generation
make proto-gen     # Generate gRPC code from proto files
make proto-clean   # Clean generated proto files
make proto-tools   # Install proto generation tools

# Swagger/OpenAPI documentation
make swagger-gen   # Generate Swagger/OpenAPI documentation
make swagger-clean # Clean generated swagger files
make swagger-tools # Install swagger generation tools
```

### Code Quality & Linting

The project includes comprehensive linting and code quality tools:

#### Available Linting Commands

```bash
# Primary linting with golangci-lint
make lint           # Run golangci-lint
make lint-fix       # Run golangci-lint with auto-fix
make lint-verbose   # Run golangci-lint with verbose output
make lint-new       # Lint only new/changed files

# Specialized linting
make lint-linter LINTER=name  # Run specific linter (e.g., LINTER=errcheck)
make lint-security  # Security analysis with gosec
make lint-style     # Format and style checking
make lint-misspell  # Check for spelling errors

# Code formatting
make format         # Format code with go fmt
make format-strict  # Stricter formatting with gofumpt

# Development tools
make tools          # Install all linting tools
```

#### Pre-commit Hooks

Set up automated code quality checks before commits using golangci-lint:

```bash
# Install pre-commit hooks
make install-hooks

# Run pre-commit checks manually (golangci-lint + tests)
make pre-commit

# Run comprehensive quality checks
make quality
```

#### Linting Tools Included

- **golangci-lint**: Primary comprehensive linter with multiple checkers
  - **gofmt**: Standard Go code formatting
  - **goimports**: Import organization and formatting  
  - **misspell**: Spelling error detection
  - **whitespace**: Whitespace formatting
  - **gocyclo**: Cyclomatic complexity analysis
  - **funlen**: Function length checking
  - **lll**: Line length checking
- **gosec**: Security vulnerability scanner
- **ineffassign**: Inefficient assignment detection
- **staticcheck**: Advanced static analysis (via tools target)

#### CI/CD Integration

```bash
# Complete CI pipeline
make ci             # Runs linting, tests, and builds
```

### gRPC Services

The gRPC server provides efficient, type-safe APIs for both User and Payment services:

#### Available Services

**User Service** (`api/proto/user/user.proto`):
- `CreateUser` - Create a new user
- `GetUser` - Get user by ID
- `ListUsers` - List users with pagination
- `UpdateUser` - Update user information
- `DeleteUser` - Delete a user
- `UpdateUserPassword` - Update user password

**Payment Service** (`api/proto/payment/payment.proto`):
- `CreatePayment` - Create a new payment
- `GetPayment` - Get payment by ID
- `ListPayments` - List payments with filtering
- `UpdatePayment` - Update payment information
- `DeletePayment` - Delete a payment
- `GetUserPayments` - Get payments for a specific user

#### Proto Generation

```bash
# Install proto tools
make proto-tools

# Generate gRPC code from proto files
make proto-gen

# Clean generated files
make proto-clean
```

#### gRPC Client Example

```go
import (
    "google.golang.org/grpc"
    "vibe-ddd-golang/api/proto/user"
    "vibe-ddd-golang/api/proto/payment"
)

// Connect to gRPC api
conn, err := grpc.Dial("localhost:9090", grpc.WithInsecure())
defer conn.Close()

// Create clients
userClient := user.NewUserServiceClient(conn)
paymentClient := payment.NewPaymentServiceClient(conn)

// Use the services
userResp, err := userClient.CreateUser(ctx, &user.CreateUserRequest{
    Name:     "John Doe",
    Email:    "john@example.com",
    Password: "securepassword",
})
```

### Server Architecture Benefits

- **Independent Scaling**: Scale API and Worker servers independently based on load
- **Deployment Flexibility**: Deploy servers to different environments (API to web tier, workers to background tier)
- **Technology Diversity**: Each server can use different technologies (HTTP, gRPC, message queues)
- **Fault Isolation**: Issues in one server don't affect others
- **Development Workflow**: Developers can work on specific servers without affecting others
- **Graceful Shutdown**: All servers handle SIGINT/SIGTERM signals for clean shutdown

## ⚙️ Configuration

### Environment Variables

```bash
# Server
SERVER_HOST=localhost
SERVER_PORT=8080

# Database
DATABASE_HOST=localhost
DATABASE_PORT=5432
DATABASE_USER=postgres
DATABASE_PASSWORD=postgres
DATABASE_DB_NAME=vibe_db

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=""
REDIS_DB=0

# Worker
WORKER_CONCURRENCY=10
WORKER_PAYMENT_CHECK_INTERVAL=5m
WORKER_RETRY_MAX_ATTEMPTS=3

# Logging
LOGGER_LEVEL=info
LOGGER_FORMAT=json
```

### Configuration File

Edit `config.yaml` with your settings:

```yaml
server:
  host: localhost
  port: 8080
  read_timeout: 10s
  write_timeout: 10s
  idle_timeout: 60s

database:
  host: localhost
  port: 5432
  user: postgres
  password: postgres
  db_name: vibe_db
  ssl_mode: disable

redis:
  host: localhost
  port: 6379
  password: ""
  db: 0

worker:
  concurrency: 10
  payment_check_interval: 5m
  retry_max_attempts: 3
  retry_delay: 30s

logger:
  level: info
  format: json
  output_path: stdout
```

## 🔄 Background Jobs & Workers

### Job Types

| Job Type | Description | Queue | Retry |
|----------|-------------|-------|-------|
| `payment:check_status` | Check payment status with gateway | `default` | 3x |
| `payment:process` | Process payment transaction | `critical` | 3x |

### Job Queues

- **Critical**: High priority jobs (payment processing)
- **Default**: Normal priority jobs (status checks)
- **Low**: Background maintenance jobs

### Worker Features

- **Automatic Retry**: Failed jobs retry with exponential backoff
- **Graceful Shutdown**: Workers complete current jobs before shutdown
- **Dead Letter Queue**: Failed jobs after max retries
- **Job Monitoring**: Comprehensive logging and metrics

## 🏛️ Architecture Patterns

### Dependency Injection (FX)

```go
// Domain module example
var Module = fx.Options(
    fx.Provide(
        repository.NewPaymentRepository,
        service.NewPaymentService,
        handler.NewPaymentHandler,
        worker.NewPaymentWorker,
    ),
)
```

### Service Layer Pattern

```go
// Service handles business logic
func (s *paymentService) CreatePayment(req *dto.CreatePaymentRequest) (*dto.PaymentResponse, error) {
    // 1. Validate user exists (cross-domain call)
    _, err := s.userService.GetUserByID(req.UserID)
    if err != nil {
        return nil, errors.New("user not found")
    }
    
    // 2. Create payment entity
    payment := &entity.Payment{...}
    
    // 3. Save to database
    err = s.repo.Create(payment)
    
    // 4. Schedule background job
    s.scheduler.SchedulePaymentProcessing(payment.ID)
    
    return s.entityToResponse(payment), nil
}
```

### Repository Pattern

```go
type PaymentRepository interface {
    Create(payment *entity.Payment) error
    GetByID(id uint) (*entity.Payment, error)
    GetAll(filter *dto.PaymentFilter) ([]entity.Payment, int64, error)
    Update(payment *entity.Payment) error
    Delete(id uint) error
}
```

## 🧪 Testing

### Test Structure

```
internal/application/payment/
├── service/
│   ├── payment.service.go
│   └── payment.service_test.go
├── repository/
│   ├── payment.repo.go
│   └── payment.repo_test.go
└── handler/
    ├── payment.handler.go
    └── payment.handler_test.go
```

### Running Tests

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run specific domain tests
go test ./internal/application/payment/...

# Run with verbose output
go test -v ./...
```

## 🐳 Docker Support

### Dockerfile

Multi-stage build for production optimization:

```dockerfile
# Build stage
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o main .

# Production stage
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/main .
EXPOSE 8080
CMD ["./main"]
```

### Docker Compose (Example)

```yaml
version: '3.8'
services:
  api:
    build: .
    ports:
      - "8080:8080"
    environment:
      - DATABASE_HOST=postgres
      - REDIS_HOST=redis
    depends_on:
      - postgres
      - redis

  worker:
    build: .
    command: ["./worker"]
    environment:
      - DATABASE_HOST=postgres
      - REDIS_HOST=redis
    depends_on:
      - postgres
      - redis

  postgres:
    image: postgres:15
    environment:
      POSTGRES_DB: vibe_db
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres

  redis:
    image: redis:7-alpine
```

## 🎨 Adding New Domains

### 1. Create Domain Structure

```bash
mkdir -p internal/application/order/{dto,entity,repository,service,handler,worker}
```

### 2. Implement Domain Layers

```go
// internal/application/order/module.go
package order

import "go.uber.org/fx"

var Module = fx.Options(
    fx.Provide(
        repository.NewOrderRepository,
        service.NewOrderService,
        handler.NewOrderHandler,
        worker.NewOrderWorker, // optional
    ),
)
```

### 3. Register Domain

```go
// internal/domain/providers.go
var Module = fx.Options(
    user.Module,
    payment.Module,
    order.Module,    // Add new domain
    fx.Provide(NewModuleRegistry),
)
```

### 4. Add Routes

```go
// internal/application/order/handler/order.handler.go
func (h *OrderHandler) RegisterRoutes(api *gin.RouterGroup) {
    orders := api.Group("/orders")
    {
        orders.POST("", h.CreateOrder)
        orders.GET("", h.GetOrders)
        // ... more routes
    }
}
```

## 🔧 Best Practices

### Code Organization

1. **One domain per directory**: Keep related code together
2. **Interface-driven design**: Define interfaces in the domain layer
3. **Dependency injection**: Use fx for clean dependency management
4. **Error handling**: Wrap errors with context
5. **Logging**: Use structured logging throughout

### Database

1. **Migrations**: Use GORM auto-migrate or migration tools
2. **Transactions**: Handle transactions in service layer
3. **Connection pooling**: Configure appropriate pool sizes
4. **Indexing**: Add indexes for frequently queried fields

### Security

1. **Input validation**: Validate all inputs using DTO bindings
2. **Password hashing**: Use bcrypt for password storage
3. **SQL injection**: Use parameterized queries (GORM handles this)
4. **CORS**: Configure CORS headers appropriately

### Performance

1. **Database queries**: Use efficient queries and avoid N+1 problems
2. **Caching**: Implement Redis caching for frequently accessed data
3. **Background jobs**: Use workers for heavy processing
4. **Connection limits**: Configure appropriate timeouts and limits

## 📖 Additional Resources

- [Domain-Driven Design](https://martinfowler.com/tags/domain%20driven%20design.html)
- [Clean Architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
- [Go Project Layout](https://github.com/golang-standards/project-layout)
- [Uber FX Documentation](https://uber-go.github.io/fx/)
- [Asynq Documentation](https://github.com/hibiken/asynq)

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Follow the domain-driven design patterns
4. Add tests for new functionality
5. Commit your changes (`git commit -m 'Add amazing feature'`)
6. Push to the branch (`git push origin feature/amazing-feature`)
7. Open a Pull Request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

**Built with ❤️ using Go, following Domain-Driven Design principles**