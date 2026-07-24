# CLAUDE.md

Guidance for Claude Code (claude.ai/code) working in this repository.

## What this is

A general-purpose Go DDD boilerplate for service APIs: an fx composition root, a
standardized response envelope, a reusable `internal/pkg` library layer, and Atlas migrations.
Encryption/auth are intentionally **not** included here — this is the plaintext
starting template. Two example domains (`user`, `payment`) demonstrate the full pattern.

## Common commands

Run inside the repo root:

| Target | Purpose |
|--------|---------|
| `make run` | Run the API (`go run ./cmd/api`), default port `8080` |
| `make build` | Build `bin/vibe-ddd-golang` (CGO disabled) |
| `make test` | `go test -race -cover ./...` |
| `make test-unit` / `test-integration` | scoped test runs |
| `make lint` / `lint-fix` | golangci-lint |
| `make format` | gofmt + gofumpt + goimports |
| `make swagger-gen` | Regenerate Swagger into `internal/server/api/docs/` (`swag init`) |
| `make proto` | Regenerate gRPC stubs under `internal/server/grpc/proto` |
| `make deps` | `go mod download && go mod tidy` |

Single test: `go test -race -run TestName ./internal/application/<domain>/...`

### Migrations (Atlas, `cmd/migration`)

| Target | Purpose |
|--------|---------|
| `make migrate-status` | show migration state |
| `make migrate-apply` | apply pending migrations |
| `make migrate-rollback` (`VERSION=`) | roll back last (or to a version) |
| `make migrate-init NAME=init_schema` | generate the initial migration |
| `make migrate-diff NAME=add_x DEV_DSN='postgres://...'` | diff entities → new migration |

Entities owned by the service are listed in
[`internal/server/migration/migration.server.go`](internal/server/migration/migration.server.go)
(`entities()`), which drives both dev auto-migrate and Atlas diffing.

## Architecture

fx-based layered shape. The repo map below is the lookup index — consult it before
grepping; open source files only for the symbol you actually need.

<!-- REPO-MAP:START — format: path | purpose | key symbols. Regenerate by scanning `find cmd infrastructure internal -type d` after structural changes. -->
```
cmd/api                    | HTTP+gRPC bootstrap: signals + fx graph + shutdown | main.go, server.go
cmd/migration              | Atlas migration CLI                                | config/
infrastructure             | composition root: fx providers                     | database.go (named "main_db"), redis.go, rabbitmq.go, lifecycle.go, common.go
internal/config            | Viper config structs — ALL runtime knobs           | Config, IsProduction()
internal/common/enum       | stable result codes clients branch on              | result_code.enum.go
internal/common/params     | fx DI bundle injected into repositories            | Params{Ctx, MainDB, Redis*, RabbitMQ*, Publisher*} (*nilable)
internal/common/type       | envelope DTOs                                      | Response, ResponseAPI
internal/pkg/response      | envelope render + business errors                  | Send, Render, Validate, New*Exception(msg).WithCause(err)
internal/pkg/reqbind       | body/query binding seam (future encryption point)  | Bind, BindQuery, Query
internal/pkg/reqctx        | request-scoped accessors                           | RequestID, StartTime, Requested/SelectedVersion
internal/pkg/middleware    | global chain, order matters                        | CorsMiddleware → RequestInit → ResponseInit → Recovery
internal/pkg/db            | gorm wrapper + pagination/jsonraw/cache            | Database (embeds *gorm.DB), WithContext, paginate
internal/pkg/redis         | optional Redis client + pagination                 | connect, action, paginate
internal/pkg/rabbitmq      | optional AMQP events pub/sub                       | Publisher, Subscriber, Event, Message
internal/pkg/logger        | zap setup                                          | InitializeLogger
internal/pkg/validation    | validator setup + map validators                   | Setup, Validate
internal/pkg/migration     | Atlas engine driving cmd/migration                 | service.migration.go
internal/pkg/grpcstatus    | AppError → gRPC status mapping                     | grpcstatus.go
internal/pkg/helper/*      | small utils                                        | apperror, convert, crypto, env, ginctx, hmac, httpclient, idgen, pointer, response (BuildResponse), time
internal/pkg/testutil      | test harness                                       | router.go, database.go, mocks.go, fixtures.go
internal/application/user  | REFERENCE domain — mirror this                     | entity, dto, repository, service, handler (+gRPC), module.go, *_test.go
internal/application/payment | 2nd example; cross-domain dep on user            | same shape
internal/server/api        | route composition, swagger, health                 | providers.go (register domain Modules), module.go (Server, RegisterRoutes)
internal/server/grpc       | in-process gRPC server + stubs                     | proto/{user,payment}
internal/server/migration  | entity registry — single source of truth           | entities(), ListTableEntity()
docs/specs                 | feature requirements, API contracts, design, tasks | README.md, 0000 template package
docs/adr                   | ADRs (protocol: writing-adr skill)                 | README.md index, 0000 template
docs/diagram               | Mermaid diagrams (protocol: writing-diagram skill) | README.md index, 0000 template
```
<!-- REPO-MAP:END -->

**Composition root.** `cmd/api/main.go` only supplies the signal-aware `app_context`,
provides `config.NewConfig` + the `infrastructure.*` constructors, invokes
`InitializeLogger`/`InitializeValidation`, mounts `serverapi.Module` plus
`grpcserver.Module`, and invokes `Run`. All singleton wiring lives in `infrastructure/`.
The primary DB is a **named** result (`name:"main_db"`).

**Params bundle.** Repositories inject
[`params.Params`](internal/common/params/params.go) (`fx.In`) and read `p.MainDB`
(a `*db.Database` that embeds `*gorm.DB`). Redis/RabbitMQ/Publisher are also on the
bundle and may be **nil** (both are optional — see Config). Repository methods take
`context.Context` and call `r.db.WithContext(ctx)...` so request ids propagate.

**Domain anatomy.** Handlers stay thin: `reqbind.Bind`/`BindQuery` → `response.Validate`
→ service → `response.Send` (success) or `response.Render` (error). Business rules live in
`service`, which returns `*response.AppError` constructors (`NewNotFoundException`,
`NewConflictException`, `NewBadRequestException`, …) — never raw transport errors.

**Response envelope.** Every handler emits one shape:
`{requestId, code, message, debug{version,error,startTime,endTime,runtimeMs}, data}`.
Clients branch on `code` (stable, in
[`internal/common/enum/result_code.enum.go`](internal/common/enum/result_code.enum.go)),
never on `message`. `X-Request-Id` (`req_` + UUIDv7) is echoed on every response.
The envelope is built by the middleware chain: `RequestInit` (stamps id/startTime) →
`ResponseInit` (installs the `send` fn handlers call) → `Recovery` (panic → 500 envelope),
all mounted globally in [`internal/server/api/module.go`](internal/server/api/module.go).

**Config.** Viper + `config.yaml` (see `config.sample.yaml`); every value overridable by
env. `internal/config/config.go` holds the structs; `cfg.IsProduction()` gates gin mode /
zap encoder. Redis (`redis.host`) and RabbitMQ (`rabbitmq.uri`) are optional — leave them
empty to boot on Postgres alone. `database.sync: true` runs GORM auto-migrate on boot (dev).

**Routing.** Health (`/healthz`, `/readyz`) mounts at the root; domains mount under
`/api/v1`. Swagger UI at `/swagger/index.html`.

**gRPC transport.** `internal/server/grpc` owns one in-process `*grpc.Server` served from
`cmd/api` on `api.grpc_port` (default `9090`). Domain gRPC handlers live beside HTTP
handlers and call the same service layer. Reflection, health, and interceptors are
intentionally omitted for the internal-only be-general pattern. Regenerate stubs with
`make proto`.

**Feature and architecture docs.** Feature specs live in [`docs/specs/`](docs/specs/) as one
package per Jira feature: `requirements.md` (behavior), `api-contract.md` when transport behavior
changes (exact wire contract), `design.md` (repository solution), and `tasks.md` (Jira-linked
execution/evidence). ADRs live in [`docs/adr/`](docs/adr/) and diagrams live in
[`docs/diagram/`](docs/diagram/). Feature designs link durable ADRs and diagrams instead of
duplicating them.

## Adding a domain

Follow the [`golang-ddd-domain` skill](.claude/skills/golang-ddd-domain/SKILL.md) — the
authoritative protocol (file anatomy, build order, 4-point registration checklist, hard
rules). Mirror `internal/application/user/`. Or run `/deliver-domain <domain> <behavior>`.

## Conventions

`.claude/skills/golang-ddd-domain/SKILL.md` is the authoritative domain convention set.
Architecture sections above are orientation, not a second implementation rubric.

## Agentic tooling (`.claude/`)

Skills, commands, and agents live in `.claude/`. Skills are the authoritative protocols;
commands and agents point at them — don't duplicate their content elsewhere. Codex discovers
the same skills through the per-file symlinks under `.agents/skills/`, while `AGENTS.md` points
directly to this file.

**Lookup order (token discipline):** repo map above → relevant skill → the one source file
you need. Don't tree-walk or read whole packages; `internal/application/user/` is the only
domain worth reading end-to-end (it's the reference).

**Primary commands** (invoke as `/name <args>`):

| Command | Purpose |
|---|---|
| [`/deliver-domain`](.claude/commands/deliver-domain.md) | Create or extend a domain and run all applicable delivery checks |
| [`/write-spec`](.claude/commands/write-spec.md) | Convert Jira/PRD input into gated requirements, API contract, design, and tasks |
| [`/write-docs`](.claude/commands/write-docs.md) | Route architecture documentation to ADR, diagram, or both |
| [`/review-domain`](.claude/commands/review-domain.md) | Domain and approved-contract review of a domain or current diff (spawns `ddd-reviewer`) |

Quick start:

```text
/write-spec PD-1234 add scorecard template management
/deliver-domain payment add a refund endpoint
/write-docs document the payment retry strategy
/review-domain payment
```

**Utility commands** (standalone safety-sensitive workflows):

| Command | Purpose |
|---|---|
| [`/add-migration`](.claude/commands/add-migration.md) | Atlas migration from entity changes (`migrate-diff` workflow) |
| [`/verify-api`](.claude/commands/verify-api.md) | Boot the API + probe envelope shape (local twin of CI smoke job) |

```text
/add-migration add_payment_status
/verify-api payment
```

Their underlying skills are loaded automatically by `/deliver-domain` when applicable.

CI (`.github/workflows/ci.yml`) enforces on push/PR: vet + race tests + build, REPO-MAP
freshness (`scripts/repo-map-check.sh`), agentic registry integrity
(`scripts/agentic-docs-check.sh`), swagger drift, and a booted-API smoke probe.

**Skills** (auto-trigger by topic):

| Skill | Triggers on |
|---|---|
| [`golang-ddd-domain`](.claude/skills/golang-ddd-domain/SKILL.md) | adding/changing domain code (entity, dto, repo, service, handler) |
| [`agentic-delivery`](.claude/skills/agentic-delivery/SKILL.md) | autonomous implementation, validation, review, and repair loops |
| [`atlas-migration`](.claude/skills/atlas-migration/SKILL.md) | generating, reviewing, and validating Atlas migrations |
| [`api-runtime-verification`](.claude/skills/api-runtime-verification/SKILL.md) | safely booting and probing the local API |
| [`writing-spec`](.claude/skills/writing-spec/SKILL.md) | Jira-driven requirements, API contracts, design, task mapping, and evidence |
| [`writing-adr`](.claude/skills/writing-adr/SKILL.md) | documenting architectural decisions |
| [`writing-diagram`](.claude/skills/writing-diagram/SKILL.md) | Mermaid diagrams, sequences, ERDs |

**Agents** (spawn via the Agent/Task tool):

| Agent | Use for |
|---|---|
| [`ddd-reviewer`](.claude/agents/ddd-reviewer.md) | read-only domain and approved-contract compliance review |
| [`docs-writer`](.claude/agents/docs-writer.md) | drafting ADRs/diagrams as part of a larger change |
