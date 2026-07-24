---
name: atlas-migration
description: Protocol for generating, reviewing, applying, and validating Atlas migrations from GORM entity changes. Use when entities change, a schema migration is requested, or migration safety must be reviewed.
---

# Atlas Migration Protocol

## Purpose

Generate deterministic schema migrations from the service-owned entity set without touching
shared data or accepting destructive SQL blindly.

## Preconditions

- The intended entity change is complete and reviewed.
- Every changed service-owned entity appears in both `entities()` and `ListTableEntity()` in
  `internal/server/migration/migration.server.go`.
- `DEV_DSN` points to a disposable local/scratch Postgres database, never shared, staging, or
  production infrastructure.
- The migration name is short snake case and describes the schema outcome.

If database ownership or DSN safety is uncertain, stop and ask.

## Flow

1. Inspect the entity diff and migration registrations.
2. Generate against the disposable dev database:
   ```bash
   make migrate-diff NAME=<name> DEV_DSN='<disposable-postgres-dsn>'
   ```
3. Review every generated SQL statement. Confirm table/column names, types, nullability,
   defaults, indexes, constraints, and ordering match the entity intent.
4. Treat drops, truncation, narrowing type changes, new non-null columns without safe defaults,
   and irreversible data rewrites as destructive. Stop for explicit approval before proceeding.
5. Apply only to the disposable local database, then inspect status:
   ```bash
   make migrate-apply
   make migrate-status
   ```
6. Validate rollback when the migration engine supports a safe inverse. Never claim rollback
   safety from command availability alone; inspect the generated down operation.
7. Run `make build` and the tests covering the changed entity/repository behavior.

## Environment rules

- Atlas migrations are authoritative for shared environments; keep `database.sync` disabled
  there.
- Development auto-migrate does not replace a committed migration.
- Never apply migrations to remote infrastructure unless the user explicitly requests it and
  identifies the target environment.
- Never print credentials or commit DSNs.

## Acceptance

- Entity ownership registrations are complete.
- Generated SQL contains only intended changes.
- Destructive operations have explicit approval and a recovery plan.
- Local apply/status succeeds, or the environment failure is reported accurately.
- Relevant build/tests pass.

## Hard rules

- Generate from the registered entity set; never hand-edit around a missing registration.
- Review SQL before apply.
- A rollback command is not proof that rollback is lossless.
- Do not modify shared data as part of autonomous delivery.
