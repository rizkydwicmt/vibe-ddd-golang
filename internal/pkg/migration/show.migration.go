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
    --dev <string>      Database connection string for comparing with GORM models

  UTILITY FLAGS:
    --dry-run           Preview changes without creating files
    --force             Force operation even if files exist
                        When used with --apply: continue on errors and track failed migrations
                        When used with --baseline: mark as baseline even if migrations exist
    --verbose           Enable verbose logging
    --help              Show this help message

EXAMPLES:

  1. Generate migration from schema changes:
     go run main.go --diff --dev="user:pass@tcp(127.0.0.1:3306)/db"

  2. Generate and apply migration in one command:
     go run main.go --diff --dev="user:pass@tcp(127.0.0.1:3306)/db" --apply

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
     go run main.go --diff --dev="user:pass@tcp(127.0.0.1:3306)/db" --dry-run

  9. Create migration with custom name:
      go run main.go --diff --dev="user:pass@tcp(127.0.0.1:3306)/db" --name=add_user_table

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

  General Environment Variables
  CONFIG_TYPE 			Configuration type (default: env) (example: secret, env)
  MIGRATIONS_DIR 		Migrations directory (default: migrations)
  DEBUG 				Enable debug logging (default: false)

  Configuration By Environment:
  DB_HOST              	Database host (default: 127.0.0.1)
  DB_PORT              	Database port (default: 3306)
  DB_USER              	Database user (default: user)
  DB_PASSWORD          	Database password (default: password)
  DB_NAME              	Database name (default: db_staging)

  Configuration By Secret:
  PROJECT_ID           	Google Cloud project ID (example: project-name)
  PROJECT_NUMBER       	Google Cloud project number (example: 12345)
  APP_ENV              	Application environment (example: local, development, production)
  APP_NAME             	Application name (example: summary)
  APP_PORT             	Application port (example: 8002)
  APP_TYPE             	Application type (service, web, mobile)
  GOOGLE_APPLICATION_CREDENTIALS  Path to Google Cloud credentials file
  GOOGLE_PRIVATE_KEY_ID           Google service account private key ID
  GOOGLE_PRIVATE_KEY              Google service account private key
  GOOGLE_CLIENT_EMAIL             Google service account email
  GOOGLE_CLIENT_ID                Google service account client ID

NOTES:
  - Migrations are stored in the 'migrations' directory
  - Each migration consists of .up.sql and .down.sql files
  - Migration version format: YYYYMMDDHHMMSS
  - All migrations are versioned and tracked in the 'schema_migrations' table
  - Use --baseline when you have an existing database and want to start tracking migrations
  - Use --dev to specify a connection string for comparing schema between database and GORM models

For more information, visit the documentation.
`)
}
