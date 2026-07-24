package migration

import (
	database "vibe-ddd-golang/internal/pkg/db"
)

type Config struct {
	MigrationsDir string
	DSN           string
	DbConfig      *database.Config
	Statements    *string
	IsDebug       bool
	TableList     []string
}
