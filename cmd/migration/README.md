# Database Migration Tool

The migration CLI generates, reviews, applies, and rolls back versioned Atlas migrations for
the entities registered in `internal/server/migration/migration.server.go`.

## Diff Model

`DEV_DSN` is optional. `make migrate-diff` resolves the existing database in this order:

1. `DEV_DSN`, when explicitly supplied.
2. `database.*` from `config.yaml`, with `DATABASE_*` environment overrides such as
   `DATABASE_HOST` and `DATABASE_DB_NAME`.

The supplied or configured database is the **existing schema**. To build the desired schema,
the loader creates a temporary PostgreSQL schema or MySQL database, executes the registered
GORM entity statements there, asks Atlas to compare both schemas, then drops the temporary
namespace. A diff-only run does not create the `schema_migrations` table.

The database user therefore needs permission to inspect the existing schema and create/drop the
temporary namespace. Implicit config/env fallback is blocked for staging and production; pass an
explicit `DEV_DSN` only after confirming the target is safe.

## Commands

```bash
# Generate a diff using database.* / DATABASE_*
make migrate-diff NAME=add_payment_status

# Override the database connection
make migrate-diff NAME=add_payment_status DEV_DSN='postgres://user:pass@localhost:5432/app?sslmode=disable'

# Preview without writing migration files
go run ./cmd/migration --diff --name=add_payment_status --dry-run

# Generate the initial migration
make migrate-init NAME=init_schema

# Apply, inspect, and roll back
make migrate-apply
make migrate-status
make migrate-rollback
make migrate-rollback VERSION=20260724120000
```

Each generated migration contains matching files:

- `<version>_<description>.up.sql`
- `<version>_<description>.down.sql`

## Environment

Every `config.yaml` path can be overridden by the equivalent uppercase environment variable:

| Config path | Environment variable |
|---|---|
| `app.name` | `APP_NAME` |
| `app.environment` | `APP_ENVIRONMENT` |
| `database.driver` | `DATABASE_DRIVER` |
| `database.host` | `DATABASE_HOST` |
| `database.port` | `DATABASE_PORT` |
| `database.user` | `DATABASE_USER` |
| `database.password` | `DATABASE_PASSWORD` |
| `database.db_name` | `DATABASE_DB_NAME` |
| `database.ssl_mode` | `DATABASE_SSL_MODE` |
| `database.timezone` | `DATABASE_TIMEZONE` |
| `migration.migrations_dir` | `MIGRATION_MIGRATIONS_DIR` |
| `migration.debug` | `MIGRATION_DEBUG` |

`DEV_DSN` is a Make variable used to populate the optional `--dev` CLI flag; never commit it.

## Safety

- Run `--dry-run` and review every generated statement before applying it.
- Treat drops, truncation, narrowing types, and irreversible rewrites as destructive.
- Keep `database.sync` disabled in shared environments; committed migrations are authoritative.
- Never run apply, rollback, baseline, or diff against remote infrastructure without explicit
  approval and a confirmed recovery plan.
