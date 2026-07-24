package config

import (
	"os"
	"testing"
)

func migrationEnv(t *testing.T) {
	t.Helper()

	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("get cwd: %v", err)
	}
	if err := os.Chdir(t.TempDir()); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(oldDir); err != nil {
			t.Fatalf("restore cwd: %v", err)
		}
	})

	t.Setenv("APP_NAME", "migration-test")
	t.Setenv("APP_ENVIRONMENT", "local")
	t.Setenv("DATABASE_DRIVER", "postgres")
	t.Setenv("DATABASE_HOST", "db.internal")
	t.Setenv("DATABASE_PORT", "5544")
	t.Setenv("DATABASE_USER", "migration-user")
	t.Setenv("DATABASE_PASSWORD", "migration-pass")
	t.Setenv("DATABASE_DB_NAME", "migration-db")
}

func TestNewConfigEnvUsesConfiguredDatabaseWhenDevDSNOmitted(t *testing.T) {
	migrationEnv(t)

	cfg, err := NewConfigEnv("", false)
	if err != nil {
		t.Fatalf("NewConfigEnv() error = %v", err)
	}
	if cfg.DbConfig.URL != "" {
		t.Fatalf("expected empty URL fallback, got %q", cfg.DbConfig.URL)
	}
	if cfg.DbConfig.Host != "db.internal" || cfg.DbConfig.Port != 5544 {
		t.Fatalf("expected configured database, got %s:%d", cfg.DbConfig.Host, cfg.DbConfig.Port)
	}
	if cfg.DbConfig.User != "migration-user" || cfg.DbConfig.Database != "migration-db" {
		t.Fatalf("expected configured credentials/database, got %s/%s", cfg.DbConfig.User, cfg.DbConfig.Database)
	}
}

func TestNewConfigEnvUsesDevDSNOverride(t *testing.T) {
	migrationEnv(t)
	const devDSN = "postgres://override:secret@localhost:5432/override_db?sslmode=disable"

	cfg, err := NewConfigEnv(devDSN, false)
	if err != nil {
		t.Fatalf("NewConfigEnv() error = %v", err)
	}
	if cfg.DbConfig.URL != devDSN {
		t.Fatalf("expected DEV_DSN override, got %q", cfg.DbConfig.URL)
	}
}
