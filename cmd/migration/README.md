# Database Migration Tool

This is a comprehensive database migration tool that allows you to manage schema changes in a controlled and versioned manner. It supports generating, applying, and rolling back migrations, as well as displaying the current migration status.

## Project Structure

```
├── cmd
│   └── migration
│       ├── main.go                  # Main entry point for the CLI
│       ├── .env                     # Environment variables for local development
│       └── config                   # Default Configuration
├── internal
│   ├── pkg
│   │   ├── migration                # Migration package for handling migration logic
│   │   │   └── config.go            # Configuration handling
│   │   │   └── show.go              # Show help message display
│   │   │   └── helper.go            # helper functions for migration
│   │   │   └── validator.go         # Command-line flag validation
│   │   │   └── service.go           # Migration service implementation
│   │   │   └── loader.go            # Schema loading functionality
│   │   ├── db
│   │   │   └── db.go                # Database connection handling
│   │   ├── helper
│   │   │   └── godotenv.go          # Environment variable helpers
│   │   ├── logger
│   │   │   └── logger.go            # Logging setup
│   └── server
│       └── migration.go             # Server-side migration handlers
├── migrations                       # Directory for migration files
└── README.md
```

## Features

- Generate initial schema migrations from existing database
- Generate incremental migration diffs between GORM models and database
- Apply migrations with transaction support
- Rollback migrations to previous state or specific version
- Baseline existing databases for migration tracking
- Dry-run functionality to preview changes
- Force mode to handle errors during migration
- Migration status display

## Usage

```bash
# Generate migration from schema changes
go run cmd/migration/main.go --diff --dev="user:pass@tcp(127.0.0.1:3306)/db"

# Generate and apply migration in one command
go run cmd/migration/main.go --diff --dev="user:pass@tcp(127.0.0.1:3306)/db" --apply

# Apply all pending migrations
go run cmd/migration/main.go --apply

# Apply specific number of migrations
go run cmd/migration/main.go --apply --count=2

# Check migration status
go run cmd/migration/main.go --status

# Rollback last migration
go run cmd/migration/main.go --rollback

# Rollback to specific version
go run cmd/migration/main.go --rollback --version=20240425123456

# Preview changes without writing files
go run cmd/migration/main.go --diff --dev="user:pass@tcp(127.0.0.1:3306)/db" --dry-run

# Create migration with custom name
go run cmd/migration/main.go --diff --dev="user:pass@tcp(127.0.0.1:3306)/db" --name=add_user_table

# Apply migrations with force (handle existing tables)
go run cmd/migration/main.go --apply --force

# Generate initial migration (this args only for existing database reengineering with golang)
go run cmd/migration/main.go --init

# Mark existing migrations as baseline (this args only for existing database reengineering with golang)
go run cmd/migration/main.go --baseline

# Mark migrations as baseline up to a specific version
go run cmd/migration/main.go --baseline --version=20240425123456

# Mark migrations as baseline and apply new ones
go run cmd/migration/main.go --baseline --apply
```

## Environment Variables

The tool uses the following environment variables:

### General Environment Variables
- `CONFIG_TYPE`: Configuration type (default: env) (example: secret, env)
- `MIGRATIONS_DIR`: Migrations directory (default: migrations)
- `DEBUG`: Enable debug logging (default: false)

### Configuration By Environment
- `DB_HOST`: Database host (example: 127.0.0.1)
- `DB_PORT`: Database port (example: 3306)
- `DB_USER`: Database user (example: user)
- `DB_PASSWORD`: Database password (example: pass)
- `DB_NAME`: Database name (example: db_staging)

### Configuration By Secret
- `PROJECT_ID`: Google Cloud project ID (example: project-name)
- `PROJECT_NUMBER`: Google Cloud project number (example: 12345)
- `APP_ENV`: Application environment (example: local, development, production)
- `APP_NAME`: Application name (example: summary)
- `APP_PORT`: Application port (example: 8002)
- `APP_TYPE`: Application type (service, web, mobile)
- `GOOGLE_APPLICATION_CREDENTIALS`: Path to Google Cloud credentials file
- `GOOGLE_PRIVATE_KEY_ID`: Google service account private key ID
- `GOOGLE_PRIVATE_KEY`: Google service account private key
- `GOOGLE_CLIENT_EMAIL`: Google service account email
- `GOOGLE_CLIENT_ID`: Google service account client ID

## Migration Files

Each migration consists of two files:
- `<version>_<description>.up.sql`: Contains statements to apply the migration
- `<version>_<description>.down.sql`: Contains statements to revert the migration

The version format is YYYYMMDDHHMMSS.

## Schema Tracking

All migrations are tracked in the `schema_migrations` table with the following structure:

```sql
CREATE TABLE schema_migrations (
    version varchar(255) NOT NULL,
    description varchar(255) NOT NULL,
    applied bigint NOT NULL DEFAULT 0,
    total bigint NOT NULL DEFAULT 0,
    executed_at timestamp NOT NULL,
    execution_time bigint NOT NULL,
    PRIMARY KEY (version)
) CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
```

## Best Practices

1. Always check the status before applying migrations
2. Use dry-run to preview changes before applying them
3. Keep migrations small and focused on specific changes
4. Test migrations in development environments before applying to production
5. Use baseline for existing databases that have not been tracked yet
6. Regularly backup your database before running migrations

## TODO

- [ ] Implement support for multiple database types (PostgreSQL, SQLite, etc.)
- [ ] Change CLI to Cobra for better command management
