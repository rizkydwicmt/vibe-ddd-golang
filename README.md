# Vibe DDD Golang

A production-shaped Go boilerplate following Domain-Driven Design with an fx composition
root, a standardized response envelope, a reusable `internal/pkg` library layer, and Atlas
migrations. It is a plaintext starting template for service APIs: encryption/auth are
intentionally not included. Two example domains
(`user`, `payment`) show the full pattern end to end.

## Stack

- **fx** — dependency injection / composition root
- **Gin** — HTTP framework
- **GORM** + PostgreSQL (MySQL driver also available) — persistence
- **Atlas** — versioned SQL migrations
- **zap** + package `logger` — structured logging with request-id propagation
- **go-playground/validator** — request validation
- **Redis** / **RabbitMQ** — optional (cache/session, eventing); disabled unless configured
- **gRPC** — internal transport served from the API process on `api.grpc_port`
- **Swagger** (swaggo) — API docs generated under `internal/server/api/docs`

## Quick start

```bash
cp config.sample.yaml config.yaml     # then edit database.* for your Postgres
make run                              # HTTP :8080 + gRPC :9090 (Postgres required; redis/rabbit optional)
```

`database.sync: true` auto-migrates the owned entities on boot (dev). For real environments
use the Atlas migration CLI instead (see below).

## Project layout

```
cmd/
  api/            # HTTP API bootstrap (signals + fx + graceful shutdown)
  migration/      # Atlas migration CLI (init/diff/apply/rollback/status)
infrastructure/   # fx providers: gin engine, zap, database, redis, rabbitmq
internal/
  common/
    enum/         # ResultCode + HTTP status mapping, EnvEnum
    type/         # Response / ResponseAPI envelope DTOs
    params/       # Params fx bundle (MainDB, Redis, RabbitMQ, Publisher, Ctx)
  pkg/            # logger, reqctx, helper, db, redis, rabbitmq, migration,
                  #   validation, response, middleware, reqbind
  application/
    user/         # example domain (entity, dto, repository, service, handler, module.go)
    payment/      # example domain (depends on user for cross-domain validation)
  server/
    api/          # route composition + Server + generated Swagger docs
    grpc/         # in-process gRPC transport + generated proto stubs
    migration/    # entity registry for auto-migrate + Atlas diffing
```

## Architecture docs

- Feature specs live in [`docs/specs/`](docs/specs/) as gated `requirements.md`, conditional
  `api-contract.md`, `design.md`, and `tasks.md` packages. Generated Swagger verifies the approved
  contract; it does not replace the feature spec. Start with the
  [`docs/specs` quick start](docs/specs/README.md) before creating, reviewing, or implementing one.
- ADRs live in [`docs/adr/`](docs/adr/); use them to record durable decisions and tradeoffs.
- Diagrams live in [`docs/diagram/`](docs/diagram/); use them for architecture, workflow, and runtime views.
- Keep architecture docs generic to the service template unless documenting an example domain.
- Run `make agentic-check` before review; it catches documentation registration drift,
  mismatched 4-digit ADR/diagram headings, and render-unsafe Mermaid placeholders.

### Feature spec workflow

For human authors and reviewers:

1. Open the [`feature spec quick start`](docs/specs/README.md), then find or add the Jira feature
   in its index.
2. Review `requirements.md` first: scope, business rules, acceptance criteria, and unresolved
   questions must be clear before approving it.
3. Review `api-contract.md` when HTTP, gRPC, or event behavior changes. Confirm auth, fields,
   statuses, result codes, failures, and compatibility; otherwise mark it `Not required`.
4. Review `design.md` only after requirements and the applicable contract are approved. Follow its
   direct links to relevant [ADRs](docs/adr/README.md) and [diagrams](docs/diagram/README.md).
5. Use `tasks.md` to track Jira mappings, implementation slices, verification, and completion
   evidence. Generated Swagger verifies the approved contract; it is not the source requirement.

Create or update a package with the [`/write-spec` command](.claude/commands/write-spec.md):

```text
/write-spec PD-1234 add scorecard template management
```

The authoritative rules live in the [`writing-spec` skill](.claude/skills/writing-spec/SKILL.md).

## Response envelope

Every endpoint returns one shape:

```json
{
  "requestId": "req_018f...",
  "code": "OK",
  "message": "OK",
  "debug": { "version": "v1.0.0", "error": null, "startTime": "...", "endTime": "...", "runtimeMs": 3 },
  "data": { }
}
```

Clients branch on `code` (stable, machine-readable — `OK`, `CREATED`, `INVALID_PAYLOAD`,
`NOT_FOUND`, `CONFLICT`, `UNAUTHORIZED`, `INTERNAL_ERROR`, …), display `message`, and read
`data`. `X-Request-Id` (`req_` + UUIDv7) is echoed on every response.

## Endpoints

Health at the root; business routes under `/api/v1`:

- `GET /healthz`, `GET /readyz`
- `POST /api/v1/users`, `GET /api/v1/users`, `GET /api/v1/users/:id`,
  `PUT /api/v1/users/:id`, `DELETE /api/v1/users/:id`, `PUT /api/v1/users/:id/password`
- `POST /api/v1/payments`, `GET /api/v1/payments`, `GET /api/v1/payments/:id`,
  `PUT /api/v1/payments/:id`, `DELETE /api/v1/payments/:id`,
  `GET /api/v1/users/:id/payments`
- Swagger UI: `GET /swagger/index.html`

gRPC runs in the same `cmd/api` process on `api.grpc_port` (default `9090`). Reflection is
off; pass the checked-in proto files to tools such as `grpcurl`:

```bash
make proto
grpcurl -plaintext -proto internal/server/grpc/proto/user/user.proto \
  -d '{"id":1}' localhost:9090 vibe.user.v1.UserService/GetUser
```

## Handler pattern

```go
func (h *PaymentHandler) CreatePayment(c *gin.Context) {
    var req dto.CreatePaymentRequest
    if err := reqbind.Bind(c, &req); err != nil {
        response.Render(c, response.NewBadRequestException("Invalid request body").WithCause(err)); return
    }
    if err := response.Validate(c, &req); err != nil { return }

    res, err := h.service.CreatePayment(c.Request.Context(), &req)
    if err != nil { response.Render(c, err); return } // AppError → envelope

    response.Send(c, &types.Response{HTTPStatus: http.StatusCreated, Code: enum.CodeCreated, Message: "Payment created.", Data: res})
}
```

Repositories inject `params.Params` (→ `p.MainDB`) and thread `context.Context`; services
return `*response.AppError`.

## Configuration

`config.yaml` (copy from `config.sample.yaml`); any value is overridable by an env var of the
same UPPER_SNAKE path (e.g. `DATABASE_HOST`, `API_PORT`). Notable knobs: `app.environment`
(gates gin/zap mode), `api.grpc_port`, `database.sync`, `database.max_open_conns`,
`redis.host` (empty → redis disabled), `rabbitmq.uri` (empty → eventing disabled).

## Migrations

```bash
make migrate-init NAME=init_schema                       # first migration from entities
make migrate-diff NAME=add_col DEV_DSN='postgres://...'  # diff entities → new migration
make migrate-apply                                       # apply pending
make migrate-status
make migrate-rollback                                    # last (or VERSION=... to a target)
```

Owned entities live in `internal/server/migration/migration.server.go`.

## Testing

```bash
make test              # race + cover across all packages
make test-integration  # full handler→service→repo flow over in-memory SQLite
make proto             # regenerate gRPC stubs under internal/server/grpc/proto
```

Handler and integration tests run real requests through the envelope middleware and assert
the `code` + `requestId` shape.

## Docker

```bash
make docker-build && make docker-run   # API on :8080
```
