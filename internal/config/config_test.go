package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"vibe-ddd-golang/internal/common/enum"
)

func withTempConfigDir(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("get cwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir temp config dir: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(oldDir); err != nil {
			t.Fatalf("restore cwd: %v", err)
		}
	})
	return dir
}

func TestNewConfigRequiresExplicitRequiredFields(t *testing.T) {
	withTempConfigDir(t)

	_, err := NewConfig()
	if err == nil || !strings.Contains(err.Error(), "required") {
		t.Fatalf("expected required config validation error, got %v", err)
	}
}

func TestNewConfigMapsDefaultsFromStructTags(t *testing.T) {
	dir := withTempConfigDir(t)
	content := `app:
  name: vibe-ddd-golang

database:
  driver: postgres
  host: localhost
  user: postgres
  db_name: vibe_db
`
	if err := os.WriteFile(filepath.Join(dir, "config.yaml"), []byte(content), 0o600); err != nil {
		t.Fatalf("write config.yaml: %v", err)
	}

	cfg, err := NewConfig()
	if err != nil {
		t.Fatalf("NewConfig() error = %v", err)
	}
	if cfg.App.Environment != enum.DEVELOPMENT {
		t.Fatalf("expected default app environment, got %q", cfg.App.Environment)
	}
	if cfg.Server.Port != 8080 {
		t.Fatalf("expected default api port, got %d", cfg.Server.Port)
	}
	if cfg.Database.Port != 5432 {
		t.Fatalf("expected default database port, got %d", cfg.Database.Port)
	}
	if cfg.Database.Driver != DatabaseDriverPostgres {
		t.Fatalf("expected postgres database driver, got %q", cfg.Database.Driver)
	}
	if cfg.Migration.MigrationsDir != "migrations" {
		t.Fatalf("expected default migrations dir, got %q", cfg.Migration.MigrationsDir)
	}
}

func TestNewConfigNormalizesTaggedFields(t *testing.T) {
	dir := withTempConfigDir(t)
	content := `app:
  name: vibe-ddd-golang
  environment: " PRODUCTION "

database:
  driver: postgres
  host: localhost
  user: postgres
  db_name: vibe_db

logger:
  level: " DEBUG "
  format: " JSON "

crypto:
  cipher: " AES-256-CBC "
  dips_aes_method: " AES-256-CBC "
`
	if err := os.WriteFile(filepath.Join(dir, "config.yaml"), []byte(content), 0o600); err != nil {
		t.Fatalf("write config.yaml: %v", err)
	}

	cfg, err := NewConfig()
	if err != nil {
		t.Fatalf("NewConfig() error = %v", err)
	}
	if cfg.App.Environment != enum.PRODUCTION {
		t.Fatalf("expected normalized app environment, got %q", cfg.App.Environment)
	}
	if cfg.Logger.Level != enum.LogLevelDebug {
		t.Fatalf("expected normalized logger level, got %q", cfg.Logger.Level)
	}
	if cfg.Logger.Format != enum.LogFormatJSON {
		t.Fatalf("expected normalized logger format, got %q", cfg.Logger.Format)
	}
	if cfg.Crypto.Cipher != "aes-256-cbc" {
		t.Fatalf("expected normalized cipher, got %q", cfg.Crypto.Cipher)
	}
}

func TestNewConfigRejectsInvalidEnvironmentEnum(t *testing.T) {
	dir := withTempConfigDir(t)
	content := `app:
  name: vibe-ddd-golang
  environment: invalid

database:
  driver: postgres
  host: localhost
  user: postgres
  db_name: vibe_db
`
	if err := os.WriteFile(filepath.Join(dir, "config.yaml"), []byte(content), 0o600); err != nil {
		t.Fatalf("write config.yaml: %v", err)
	}

	_, err := NewConfig()
	if err == nil || !strings.Contains(err.Error(), "enum") {
		t.Fatalf("expected enum validation error, got %v", err)
	}
}

func TestNewConfigAcceptsValidEnvironmentEnum(t *testing.T) {
	dir := withTempConfigDir(t)
	content := `app:
  name: vibe-ddd-golang
  environment: staging

database:
  driver: mysql
  host: localhost
  user: postgres
  db_name: vibe_db
`
	if err := os.WriteFile(filepath.Join(dir, "config.yaml"), []byte(content), 0o600); err != nil {
		t.Fatalf("write config.yaml: %v", err)
	}

	cfg, err := NewConfig()
	if err != nil {
		t.Fatalf("NewConfig() error = %v", err)
	}
	if cfg.App.Environment != enum.STAGING {
		t.Fatalf("expected staging environment, got %q", cfg.App.Environment)
	}
	if cfg.Database.Driver != DatabaseDriverMySQL {
		t.Fatalf("expected mysql database driver, got %q", cfg.Database.Driver)
	}
}

func TestNewConfigRejectsInvalidDatabaseDriverEnum(t *testing.T) {
	dir := withTempConfigDir(t)
	content := `app:
  name: vibe-ddd-golang

database:
  driver: sqlite
  host: localhost
  user: postgres
  db_name: vibe_db
`
	if err := os.WriteFile(filepath.Join(dir, "config.yaml"), []byte(content), 0o600); err != nil {
		t.Fatalf("write config.yaml: %v", err)
	}

	_, err := NewConfig()
	if err == nil || !strings.Contains(err.Error(), "enum") {
		t.Fatalf("expected enum validation error, got %v", err)
	}
}

func TestNewConfigRejectsInvalidLoggerEnums(t *testing.T) {
	dir := withTempConfigDir(t)
	content := `app:
  name: vibe-ddd-golang

database:
  driver: postgres
  host: localhost
  user: postgres
  db_name: vibe_db

logger:
  level: trace
  format: xml
`
	if err := os.WriteFile(filepath.Join(dir, "config.yaml"), []byte(content), 0o600); err != nil {
		t.Fatalf("write config.yaml: %v", err)
	}

	_, err := NewConfig()
	if err == nil || !strings.Contains(err.Error(), "enum") {
		t.Fatalf("expected enum validation error, got %v", err)
	}
}
