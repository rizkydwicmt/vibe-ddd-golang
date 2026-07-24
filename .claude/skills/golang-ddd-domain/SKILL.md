---
name: golang-ddd-domain
description: Protocol for adding or modifying a domain (entity, dto, repository, service, handler, module) in this Go DDD boilerplate. Use when adding a new domain, endpoint, entity, repository, service, or handler, or when reviewing domain code for boilerplate compliance.
---

# Go DDD Domain Protocol

## Purpose

Every domain in this service follows one exact shape. This skill is the authoritative
checklist for building a new domain or extending an existing one. The reference
implementation is `internal/application/user/` — when in doubt, mirror it.

## File anatomy

Create `internal/application/<domain>/` with:

| File | Responsibility |
|------|----------------|
| `entity/<domain>.entity.go` | GORM model + `TableName()`. No business logic. |
| `dto/<domain>.dto.go` | Request/response/filter structs with `json` + `validate` tags. |
| `repository/<domain>.repo.go` | Interface + GORM implementation. Data access only. |
| `service/<domain>.service.go` | Interface + implementation. All business rules live here. |
| `handler/<domain>.handler.go` | Thin gin handlers + `RegisterRoutes`. Swagger godoc comments on every route. |
| `handler/<domain>.grpc.handler.go` | Optional gRPC handler calling the same service. |
| `module.go` | `var Module = fx.Options(fx.Provide(...))` exposing constructors. |

## Build order

1. **Entity** — GORM model in `entity/`. Lifecycle state = status enum string, never a boolean.
2. **DTO** — request/response/filter structs in `dto/`.
3. **Repository** — interface first, then implementation:
   ```go
   func NewXRepository(p params.Params, logger *zap.Logger) XRepository {
       return &xRepository{db: p.MainDB, logger: logger}
   }
   ```
   Every method takes `context.Context` and uses `r.db.WithContext(ctx)...`.
4. **Service** — constructor takes repository interface + `*zap.Logger`. Returns
   `*response.AppError` constructors only (`NewNotFoundException`, `NewConflictException`,
   `NewBadRequestException`, `NewInternalServerException`, chain `.WithCause(err)`).
5. **Handler** — pattern per route:
   `reqbind.Bind`/`BindQuery` → `response.Validate` → service call →
   `response.Send(c, &types.Response{...})` on success / `response.Render(c, err)` on error.
   Add Swagger godoc comments. Implement `RegisterRoutes(api *gin.RouterGroup)`.
6. **Module** — `module.go` mirrors `internal/application/user/module.go`.
7. **Register** (see checklist below).
8. **Tests** — mirror `*_test.go` files in the user domain (handler, service, repo).

## Registration checklist (all four, exact files)

- [ ] Add `<domain>.Module` to `fx.Options` in `internal/server/api/providers.go`.
- [ ] Add the handler to the `Server` struct, `NewServer` params, and
      `s.<domain>Handler.RegisterRoutes(apiV1)` in `internal/server/api/module.go`.
- [ ] Add the entity to `entities()` in `internal/server/migration/migration.server.go`
      and its `TableName()` to `ListTableEntity()`.
- [ ] gRPC (only if needed): proto under `internal/server/grpc/proto/<domain>/`, handler in
      the domain `handler/` package, `make proto`.

## Hard Rules

- Repository constructor signature is `(p params.Params, logger *zap.Logger)`; the DB is `p.MainDB`.
- Every repository/service method takes `ctx context.Context`; repos call `WithContext(ctx)` so request ids propagate.
- Service returns `*response.AppError` only — never raw `error`, never gorm/transport errors leaking upward.
- Handlers never write `gin.H` or `c.JSON` directly — always the envelope via `response.Send`/`response.Render`.
- Handlers pass `c.Request.Context()` into the service.
- Result codes come from `internal/common/enum/result_code.enum.go`; clients branch on `code`, never `message`.
- `p.Redis`, `p.RabbitMQ`, `p.Publisher` may be **nil** (optional infra) — guard before use.
- New runtime knobs go on `internal/config` (Viper), never `os.Getenv`.
- Status enums over booleans; status transitions over hard deletes.
- Keep request binding on the `reqbind` seam so transport concerns stay isolated from domains.

## Common mistakes

| ❌ Wrong | ✅ Right |
|---|---|
| `return err` from service | `return response.NewInternalServerException("...").WithCause(err)` |
| `c.JSON(200, gin.H{"data": x})` | `response.Send(c, &types.Response{HTTPStatus: http.StatusOK, Code: enum.CodeOK, ...})` |
| `os.Getenv("MY_FLAG")` | field on `internal/config` struct via Viper |
| Entity created but not in `entities()` | register in `internal/server/migration/migration.server.go` |
| `r.db.Create(user)` | `r.db.WithContext(ctx).Create(user)` |
| `IsActive bool` on entity | `Status string` enum |
| Module written but not in `providers.go` | add to `internal/server/api/providers.go` |

## Verify

```bash
make build && make test && make lint
make swagger-gen        # if routes were added/changed
make migrate-diff NAME=add_<domain>   # if an entity was added; DEV_DSN is optional
```
