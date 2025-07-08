# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Common Commands

### Build and Run
- `make build` - Build the API server binary
- `make build-worker` - Build the worker server binary
- `make build-migration` - Build the migration server binary
- `make build-grpc` - Build the gRPC server binary
- `make build-all` - Build all servers
- `make run` - Run the API server directly
- `make run-worker` - Run the worker server directly
- `make run-migration` - Run database migrations
- `make run-seed` - Seed database with initial data
- `make run-drop` - Drop all database tables
- `make run-grpc` - Run the gRPC server
- `go run .` - Alternative way to run API server
- `go run ./cmd/worker` - Alternative way to run worker server

### Testing
- `make test` - Run all tests
- `make test-coverage` - Run tests with coverage report
- `make test-unit` - Run unit tests only
- `make test-integration` - Run integration tests only
- `make test-repo` - Run repository layer tests
- `make test-service` - Run service layer tests
- `make test-handler` - Run handler layer tests
- `make test-worker` - Run worker layer tests
- `make test-user` - Run user domain tests
- `make test-payment` - Run payment domain tests
- `make test-verbose` - Run tests with verbose output

### Code Quality & Linting
- `make lint` - Run golangci-lint (includes nil detection)
- `make lint-fix` - Run golangci-lint with auto-fix
- `make lint-verbose` - Run golangci-lint with verbose output
- `make lint-new` - Lint only new/changed files
- `make lint-linter LINTER=name` - Run specific linter
- `make lint-nil-info` - Show enabled nil detection linters
- `make format` - Format code with go fmt
- `make format-strict` - Format with stricter rules (gofumpt + goimports)

### Dependencies
- `make deps` - Download and tidy dependencies
- `go mod tidy` - Tidy go modules

### Development
- `make dev-setup` - Setup development environment
- `make tools` - Install development tools
- `make quality` - Run comprehensive quality checks
- `make pre-commit` - Run pre-commit checks
- `make install-hooks` - Install pre-commit hooks
- `make ci` - Run CI checks (linting, tests, build)
- `make clean` - Clean build artifacts

### Proto Generation
- `make proto-gen` - Generate gRPC code from proto files
- `make proto-clean` - Clean generated proto files
- `make proto-tools` - Install proto generation tools

### Swagger/OpenAPI Documentation
- `make swagger-gen` - Generate Swagger/OpenAPI documentation
- `make swagger-clean` - Clean generated swagger files
- `make swagger-tools` - Install swagger generation tools

### Docker
- `make docker-build` - Build Docker image
- `make docker-run` - Run Docker container

## Architecture

This is a Go web application following **Domain-Driven Design (DDD)** principles with **NestJS-like architecture patterns**. Built with modern Go practices, microservice architecture, and comprehensive background job processing.

### Multi-Server Architecture

The application follows a **multi-server architecture** where different concerns are separated into independent, deployable servers:

| Server | Purpose | Entry Point | Default Port |
|--------|---------|-------------|--------------|
| **API Server** | HTTP REST API | `main.go` | 8080 |
| **Worker Server** | Background job processing | `cmd/worker/main.go` | - |
| **Migration Server** | Database operations | `cmd/migration/main.go` | - |
| **gRPC Server** | gRPC services (User & Payment) | `cmd/grpc/main.go` | 9090 |

### Directory Structure
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
│   │   │   ├── entity/                   # Domain entities
│   │   │   ├── repository/               # Data access layer
│   │   │   ├── service/                  # Business logic layer
│   │   │   ├── handler/                  # HTTP layer
│   │   │   ├── worker/                   # Background processing
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
│   │   ├── worker/                       # Background worker server
│   │   ├── migration/                    # Database migration server
│   │   └── grpc/                         # gRPC server
│   ├── middleware/                       # HTTP middleware
│   ├── config/                           # Configuration
│   └── pkg/                              # Internal packages
│       ├── database/database.go          # DB connection
│       ├── logger/logger.go              # Structured logging
│       ├── queue/                        # Job queue infrastructure
│       └── testutil/                     # Test utilities
├── api/                                  # API definitions
│   └── proto/                            # Protocol buffer files
│       ├── user/user.proto               # User service proto
│       └── payment/payment.proto         # Payment service proto
├── test/                                 # Integration tests
├── config.yaml                           # Configuration file
├── Makefile                              # Build automation
├── Dockerfile                            # Container image
├── go.mod                                # Go modules
└── README.md                             # Documentation
```

### Key Technologies
- **Fx**: Dependency injection container (similar to NestJS DI)
- **Gin**: HTTP web framework
- **Asynq**: Background job processing with Redis
- **gRPC**: High-performance RPC framework
- **Protocol Buffers**: Interface definition language
- **Viper**: Configuration management
- **Zap**: Structured logging
- **GORM**: ORM for database operations
- **PostgreSQL**: Primary database
- **Redis**: Message broker for background jobs
- **golangci-lint**: Comprehensive code linting
- **Docker**: Containerization support

### Application Flow
1. `main.go` bootstraps the Fx application with domain modules
2. Dependencies are injected through domain-specific Fx providers
3. `cmd/server/server.go` handles HTTP server lifecycle
4. Domain modules are registered in `internal/server/api/module.go`
5. Each domain follows the pattern: Handler → Service → Repository → Entity
6. Handlers handle HTTP requests, route registration, and delegate to services
7. Services contain business logic and validate DTOs
8. Repositories handle data persistence with GORM
9. Graceful shutdown is implemented via Fx lifecycle hooks
10. Background workers process jobs asynchronously via Asynq
11. gRPC services provide efficient, type-safe APIs
12. Database migrations handle schema changes

### Domain-Driven Design Implementation

#### Domain Structure Pattern

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

#### Key Design Principles

1. **Domain Ownership**: Each domain owns its complete vertical slice
2. **Dependency Inversion**: High-level modules don't depend on low-level modules
3. **Single Responsibility**: Each layer has a clear, focused responsibility
4. **Interface Segregation**: Small, focused interfaces
5. **Separation of Concerns**: HTTP, business logic, and data access are separated

#### Layer Responsibilities

| Layer | Responsibility | Example |
|-------|---------------|---------|  
| **Handler** | HTTP concerns, routing, request/response | `payment.handler.go` |
| **Service** | Business logic, validation, orchestration | `payment.service.go` |
| **Repository** | Data access, database operations | `payment.repo.go` |
| **Entity** | Domain models, business rules | `payment.entity.go` |
| **DTO** | Data transfer, validation, serialization | `payment.dto.go` |
| **Worker** | Background processing, async jobs | `payment/worker/` |

#### Domain Structure Best Practices
- Each domain is self-contained with its own DTO, Entity, Repository, Service, Handler, and Workers
- DTOs handle request/response validation and transformation
- Entities define database models and business rules
- Repositories provide data access abstraction
- Services implement business logic and coordinate between layers
- Handlers handle HTTP concerns, route registration, and delegate to services
- Workers handle background processing for domain-specific tasks
- Each domain has its own module.go for dependency injection configuration
- Domains can be developed and deployed independently

### Configuration
- Uses Viper for configuration management
- Default config in `config.yaml`
- Environment variables override config file values
- Configuration struct in `internal/config/config.go`

### Middleware Stack
- Request logging with structured logs
- Panic recovery with error logging
- CORS headers for cross-origin requests

### Development Notes
- The application uses structured logging with Zap
- Database integration with PostgreSQL using GORM
- Full CRUD operations implemented for User and Payment domains
- Password hashing with bcrypt for user authentication
- Background job processing with Asynq and Redis
- Separate deployable API and Worker servers
- Graceful shutdown handles SIGINT and SIGTERM signals for both servers
- Docker support with multi-stage builds

### Background Jobs & Workers

#### Job Types

| Job Type | Description | Queue | Retry |
|----------|-------------|-------|----- -|
| `payment:check_status` | Check payment status with gateway | `default` | 3x |
| `payment:process` | Process payment transaction | `critical` | 3x |

#### Job Queues

- **Critical**: High priority jobs (payment processing)
- **Default**: Normal priority jobs (status checks)
- **Low**: Background maintenance jobs

#### Worker Features

- **Automatic Retry**: Failed jobs retry with exponential backoff
- **Graceful Shutdown**: Workers complete current jobs before shutdown
- **Dead Letter Queue**: Failed jobs after max retries
- **Job Monitoring**: Comprehensive logging and metrics

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

### Available Endpoints
#### Users
- `POST /api/v1/users` - Create user
- `GET /api/v1/users` - List users (with pagination and filtering)
- `GET /api/v1/users/:id` - Get user by ID
- `PUT /api/v1/users/:id` - Update user
- `DELETE /api/v1/users/:id` - Delete user
- `PUT /api/v1/users/:id/password` - Update user password

#### Payments
- `POST /api/v1/payments` - Create payment
- `GET /api/v1/payments` - List payments (with pagination and filtering)
- `GET /api/v1/payments/:id` - Get payment by ID
- `PUT /api/v1/payments/:id` - Update payment
- `DELETE /api/v1/payments/:id` - Delete payment
- `GET /api/v1/users/:user_id/payments` - Get payments by user

#### Health
- `GET /api/v1/health` - Health check endpoint

## Configuration

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

Edit `config.yaml` or create `config.local.yaml`:

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

## Architecture Patterns

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

## Adding New Domains

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
// internal/api/api/module.go
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

## Best Practices

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