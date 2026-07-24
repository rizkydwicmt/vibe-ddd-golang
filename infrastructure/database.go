package infrastructure

import (
	"context"
	"fmt"

	"vibe-ddd-golang/internal/config"
	database "vibe-ddd-golang/internal/pkg/db"
	"vibe-ddd-golang/internal/pkg/logger"
	mg_server "vibe-ddd-golang/internal/server/migration"

	"go.uber.org/fx"
)

// DatabaseResult names the primary DB so consumers inject it by tag
// (`name:"main_db"`). Add more named results here for read-replicas etc.
type DatabaseResult struct {
	fx.Out

	MainDB *database.Database `name:"main_db"`
}

// NewDatabases opens the primary database, optionally auto-migrates, and registers
// a lifecycle hook that pings on start and closes on stop. A failed connection is
// fatal — the service cannot serve without its system of record.
func NewDatabases(lc fx.Lifecycle, cfg *config.Config) (DatabaseResult, error) {
	var result DatabaseResult

	timezone := cfg.Database.Timezone
	if timezone == "" {
		timezone = cfg.App.Timezone
	}

	db, err := database.Setup(&database.Config{
		Driver:          database.DriverEnum(cfg.Database.Driver),
		Host:            cfg.Database.Host,
		Port:            cfg.Database.Port,
		User:            cfg.Database.User,
		Password:        cfg.Database.Password,
		Database:        cfg.Database.DBName,
		SSLMode:         cfg.Database.SSLMode,
		Timezone:        timezone,
		Cache:           cfg.Database.Cache,
		CacheTime:       cfg.Database.CacheTime,
		MaxOpenConns:    cfg.Database.MaxOpenConns,
		MaxIdleConns:    cfg.Database.MaxIdleConns,
		ConnMaxLifetime: cfg.Database.ConnMaxLifetime,
		ConnMaxIdleTime: cfg.Database.ConnMaxIdleTime,
		Environment:     cfg.App.Environment,
	})
	if err != nil {
		return result, fmt.Errorf("CRITICAL: main DB connection required but failed: %w", err)
	}
	result.MainDB = db

	// Optional auto-migration on boot (dev). Production prefers the out-of-band
	// `cmd/migration` Atlas step over DATABASE sync.
	if cfg.Database.Sync {
		logger.Info.Println("running auto-migration...")
		if err := mg_server.MigrationEntity(db); err != nil {
			_ = db.Close()
			return result, fmt.Errorf("CRITICAL: auto-migration failed: %w", err)
		}
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			checkCtx, cancel := context.WithTimeout(ctx, boundedWaitTimeout(ctx, infrastructureStartupCheckTimeout))
			defer cancel()
			if err := pingDatabase(checkCtx, db); err != nil {
				return fmt.Errorf("main database readiness check failed: %w", err)
			}
			logger.Info.Println("main database readiness check passed")
			return nil
		},
		OnStop: func(context.Context) error { return db.Close() },
	})

	return result, nil
}

func pingDatabase(ctx context.Context, db *database.Database) error {
	if db == nil {
		return fmt.Errorf("database dependency is nil")
	}
	sqlDB, err := db.DB.DB()
	if err != nil {
		return fmt.Errorf("get sql.DB: %w", err)
	}
	return sqlDB.PingContext(ctx)
}
