package migration

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"ariga.io/atlas/sql/migrate"
	"ariga.io/atlas/sql/mysql"
	"ariga.io/atlas/sql/postgres"
	"ariga.io/atlas/sql/schema"

	database "vibe-ddd-golang/internal/pkg/db"
)

type Loader struct {
	tableList []string
	sts       *string
}

func NewLoader(tableList []string, statements *string) *Loader {
	return &Loader{
		tableList: tableList,
		sts:       statements,
	}
}

func (l *Loader) LoadGORMSchema(ctx context.Context, config *Config) (*schema.Realm, error) {
	driverType := config.DbConfig.Driver
	dbConfig := *config.DbConfig
	dbConfig.Cache = false
	dbConfig.MaxOpenConns = 1
	dbConfig.MaxIdleConns = 1
	dbConfig.ConnMaxLifetime = time.Minute
	dbConfig.ConnMaxIdleTime = time.Minute

	db, err := database.Setup(&dbConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database for schema loading: %w", err)
	}

	sqlDB, err := db.DB.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying DB: %w", err)
	}
	defer sqlDB.Close()

	sqlDB.SetMaxOpenConns(1) // Only need 1 connection for schema loading
	sqlDB.SetMaxIdleConns(1)
	sqlDB.SetConnMaxLifetime(1 * time.Minute)

	var driver migrate.Driver
	switch driverType {
	case database.POSTGRES:
		driver, err = postgres.Open(sqlDB)
	case database.MYSQL:
		driver, err = mysql.Open(sqlDB)
	default:
		return nil, fmt.Errorf("unsupported driver: %s", driverType)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create Atlas driver: %w", err)
	}

	// For PostgreSQL, we use a temporary schema. For MySQL, a temporary database.
	tempName := fmt.Sprintf("atlas_temp_%d", time.Now().UnixNano())

	if driverType == database.POSTGRES {
		// PostgreSQL: Use temporary schema
		if _, err := sqlDB.ExecContext(ctx, fmt.Sprintf("CREATE SCHEMA \"%s\"", tempName)); err != nil {
			return nil, fmt.Errorf("failed to create temporary schema: %w", err)
		}
		defer func() {
			if _, err := sqlDB.ExecContext(ctx, fmt.Sprintf("DROP SCHEMA IF EXISTS \"%s\" CASCADE", tempName)); err != nil {
				log.Printf("Warning: Failed to clean up temporary schema: %v", err)
			}
		}()

		// Set search path to temp schema
		if _, err := sqlDB.ExecContext(ctx, fmt.Sprintf("SET search_path TO \"%s\"", tempName)); err != nil {
			return nil, fmt.Errorf("failed to set search path: %w", err)
		}
	} else {
		// MySQL: Use temporary database
		createStmt := fmt.Sprintf("CREATE DATABASE `%s` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci", tempName)
		if _, err := sqlDB.ExecContext(ctx, createStmt); err != nil {
			return nil, fmt.Errorf("failed to create temporary database: %w", err)
		}
		defer func() {
			if _, err := sqlDB.ExecContext(ctx, fmt.Sprintf("DROP DATABASE IF EXISTS `%s`", tempName)); err != nil {
				log.Printf("Warning: Failed to clean up temporary database: %v", err)
			}
		}()

		if _, err = sqlDB.ExecContext(ctx, fmt.Sprintf("USE `%s`", tempName)); err != nil {
			return nil, fmt.Errorf("failed to switch to temp database: %w", err)
		}
	}

	// Execute GORM-generated statements
	for i, stmt := range splitStatements(*l.sts) {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}
		if config.IsDebug {
			log.Printf("Executing GORM statement %d: %s", i+1, stmt)
		}
		if _, err := sqlDB.ExecContext(ctx, stmt); err != nil {
			return nil, fmt.Errorf("failed to execute GORM statement %d: %w", i+1, err)
		}
	}

	// Inspect the schema
	sc, err := driver.InspectSchema(ctx, tempName, &schema.InspectOptions{
		Mode:   schema.InspectTables,
		Tables: l.tableList,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to inspect temp schema: %w", err)
	}

	// Normalize the schema by erasing the temporary name
	sc.Name = config.DbConfig.Database

	return schema.NewRealm(sc), nil
}
