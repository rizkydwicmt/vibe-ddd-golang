package migration

import "fmt"

func ShowHelp() {
	fmt.Printf(`
╔════════════════════════════════════════════════════════════════════════════════╗
║                           Database Migration Tool                              ║
╚════════════════════════════════════════════════════════════════════════════════╝

USAGE:
  go run main.go [flags]

AVAILABLE FLAGS:

  MIGRATION OPERATIONS:
    --init              Generate initial migration files from current database
    --diff              Generate migration diff between database and GORM models
    --apply             Apply pending migrations (can combine with --init/--diff)
    --baseline          Mark existing migrations as applied without running them
    --rollback          Rollback the last applied migration
    --status            Show current migration status (pending & applied)

  MIGRATION OPTIONS:
    --name <string>     Custom name for the migration (default: auto-generated)
    --count <int>       Number of migrations to apply (default: all)
    --version <string>  Specific version to rollback to or baseline up to
    --dev <string>      Optional database URL override for schema comparison

  UTILITY FLAGS:
    --dry-run           Preview changes without creating files
    --force             Force operation even if files exist
                        When used with --apply: continue on errors and track failed migrations
                        When used with --baseline: mark as baseline even if migrations exist
    --verbose           Enable verbose logging
    --help              Show this help message

EXAMPLES:

  1. Generate migration from schema changes:
     go run main.go --diff --name=add_user_table

  2. Override the configured database URL:
     go run main.go --diff --dev="postgres://user:pass@localhost:5432/app?sslmode=disable"

  3. Apply all pending migrations:
     go run main.go --apply

  4. Apply specific number of migrations:
     go run main.go --apply --count=2

  5. Check migration status:
     go run main.go --status

  6. Rollback last migration:
     go run main.go --rollback

  7. Rollback to specific version:
     go run main.go --rollback --version=20240425123456

  8. Preview changes without writing files:
     go run main.go --diff --dry-run

  9. Create migration with custom name:
      go run main.go --diff --name=add_user_table

  10. Apply migrations with force (handle existing tables):
      go run main.go --apply --force

  11. Generate initial migration: (handle existing db to track migrations)
	  go run main.go --init

  12. Mark existing migrations as baseline: (handle existing db to track migrations)
      go run main.go --baseline

  13. Mark migrations as baseline up to a specific version:
      go run main.go --baseline --version=20240425123456

  14. Mark migrations as baseline and apply new ones:
      go run main.go --baseline --apply

ENVIRONMENT VARIABLES:

  APP_NAME              Application name
  APP_ENVIRONMENT       local, development, staging, or production
  DATABASE_DRIVER       postgres or mysql
  DATABASE_HOST         Existing database host
  DATABASE_PORT         Existing database port
  DATABASE_USER         Existing database user
  DATABASE_PASSWORD     Existing database password
  DATABASE_DB_NAME      Existing database name
  DATABASE_SSL_MODE     PostgreSQL SSL mode
  DATABASE_TIMEZONE     Database timezone
  MIGRATION_MIGRATIONS_DIR  Migration file directory
  MIGRATION_DEBUG       Enable migration debug logging

NOTES:
  - Migrations are stored in the 'migrations' directory
  - Each migration consists of .up.sql and .down.sql files
  - Migration version format: YYYYMMDDHHMMSS
  - All migrations are versioned and tracked in the 'schema_migrations' table
  - Use --baseline when you have an existing database and want to start tracking migrations
  - --dev is optional; without it, database.* / DATABASE_* configuration is used
  - Diff creates a temporary PostgreSQL schema or MySQL database for the GORM model,
    compares it with the existing configured database, then drops the temporary namespace
  - Staging and production diff operations require an explicit --dev URL

For more information, visit the documentation.
`)
}
