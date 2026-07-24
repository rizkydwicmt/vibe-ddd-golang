package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	migrationconfig "vibe-ddd-golang/cmd/migration/config"
	"vibe-ddd-golang/internal/common/enum"
	database "vibe-ddd-golang/internal/pkg/db"
	"vibe-ddd-golang/internal/pkg/logger"
	"vibe-ddd-golang/internal/pkg/migration"
	_config "vibe-ddd-golang/internal/pkg/migration"
	"vibe-ddd-golang/internal/pkg/validation"
)

var (
	initFlag       = flag.Bool("init", false, "Generate initial migration files")
	diffFlag       = flag.Bool("diff", false, "Generate migration diff")
	dryRunFlag     = flag.Bool("dry-run", false, "Preview changes without creating files")
	nameFlag       = flag.String("name", "", "Custom name for the migration")
	forceFlag      = flag.Bool("force", false, "Force operation even if files exist")
	verboseFlag    = flag.Bool("verbose", false, "Enable verbose logging")
	applyFlag      = flag.Bool("apply", false, "Apply the migration after creation")
	baselineFlag   = flag.Bool("baseline", false, "Mark existing migrations as applied without running them")
	applyCountFlag = flag.Int("count", 0, "Number of migrations to apply (0 means all)")
	rollbackFlag   = flag.Bool("rollback", false, "Rollback the last applied migration")
	versionFlag    = flag.String("version", "", "Specific version to rollback to")
	devFlag        = flag.String("dev", "", "Optional database connection string override")
	statusFlag     = flag.Bool("status", false, "Show migration status")
	helpFlag       = flag.Bool("help", false, "Show this help message")
)

func main() {
	flag.Parse()
	logger.Setup()

	if flag.NFlag() == 0 || *helpFlag {
		migration.ShowHelp()
		return
	}

	flagOpts := &migration.FlagOptions{
		Init:       *initFlag,
		Diff:       *diffFlag,
		DryRun:     *dryRunFlag,
		Name:       *nameFlag,
		Force:      *forceFlag,
		Verbose:    *verboseFlag,
		Apply:      *applyFlag,
		Baseline:   *baselineFlag,
		ApplyCount: *applyCountFlag,
		Rollback:   *rollbackFlag,
		Version:    *versionFlag,
		DevDSN:     *devFlag,
		Status:     *statusFlag,
		Help:       *helpFlag,
	}

	if err := migration.ValidateFlags(flagOpts); err != nil {
		log.Fatal(err)
		return
	}

	err := validation.Setup()
	if err != nil {
		log.Fatal(err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	var migrationConfig *_config.Config
	migrationConfig, err = migrationconfig.NewConfigEnv(*devFlag, *verboseFlag)
	if err != nil {
		log.Fatal(err)
		return
	}
	if *diffFlag && *devFlag == "" &&
		(migrationConfig.DbConfig.Environment == enum.STAGING ||
			migrationConfig.DbConfig.Environment == enum.PRODUCTION) {
		log.Fatal("--dev is required for diff operations in staging or production environments")
		return
	}

	db, err := database.Setup(migrationConfig.DbConfig)
	if err != nil {
		log.Fatal(err)
		return
	}

	sqlDB, err := db.DB.DB()
	if err != nil {
		log.Fatal(err)
		return
	}
	sqlDB.SetMaxOpenConns(10)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)
	defer sqlDB.Close()

	schemaLoader := migration.NewLoader(migrationConfig.TableList, migrationConfig.Statements)

	migrationService, err := migration.NewService(db.DB, migrationConfig, migrationConfig.TableList, schemaLoader.LoadGORMSchema)
	if err != nil {
		log.Fatalf("Failed to create migration service: %v", err)
		return
	}

	if *applyFlag || *rollbackFlag || *initFlag || *statusFlag || *baselineFlag {
		if err := migrationService.EnsureSchemaRevisionsTable(); err != nil {
			log.Fatalf("Error ensuring schema_migrations table exists: %v", err)
		}
	}

	if *statusFlag {
		if err := migrationService.ShowMigrationStatus(); err != nil {
			log.Fatalf("Error showing migration status: %v", err)
		}
		return
	}

	if *rollbackFlag {
		if *versionFlag != "" {
			if err := migrationService.RollbackToVersion(ctx, *versionFlag); err != nil {
				log.Fatalf("Error rolling back to version %s: %v", *versionFlag, err)
			}
		} else {
			if err := migrationService.RollbackLastMigration(ctx); err != nil {
				log.Fatalf("Error rolling back last migration: %v", err)
			}
		}
		return
	}

	if *baselineFlag && !*initFlag && !*diffFlag {
		if err := migrationService.MarkMigrationsAsApplied(*versionFlag, *forceFlag); err != nil {
			log.Fatalf("Error marking migrations as applied: %v", err)
		}

		if *applyFlag {
			if err := migrationService.ApplyPendingMigrations(ctx, *applyCountFlag, *forceFlag); err != nil {
				log.Fatalf("Error applying pending migrations: %v", err)
			}
		}
		return
	}

	if *applyFlag && !*initFlag && !*diffFlag {
		if err := migrationService.ApplyPendingMigrations(ctx, *applyCountFlag, *forceFlag); err != nil {
			log.Fatalf("Error applying pending migrations: %v", err)
		}
		return
	}

	var migrationPlan *migration.MigrationPlan
	if *initFlag {
		migrationPlan, err = migrationService.GenerateInitialMigration(ctx, *nameFlag, *dryRunFlag, *forceFlag)
		if err != nil {
			log.Fatalf("Error generating initial migration files: %v", err)
		}
	} else if *diffFlag {
		migrationPlan, err = migrationService.GenerateMigrationDiff(ctx, *nameFlag, *dryRunFlag)
		if err != nil {
			log.Fatalf("Error generating migration diff: %v", err)
		}
	}

	if migrationPlan != nil {
		displayMigrationSummary(migrationPlan, migrationConfig.IsDebug)

		if *applyFlag && !*dryRunFlag {
			if err := migrationService.ApplyMigration(ctx, migrationPlan, *forceFlag); err != nil {
				log.Fatalf("Error applying migration: %v", err)
			}
			log.Printf("Migration applied successfully")
		}
	}
}

func displayMigrationSummary(plan *migration.MigrationPlan, isDebug bool) {
	fmt.Printf("\n=== Migration Summary ===\n")
	fmt.Printf("Files: %s.up.sql, %s.down.sql\n", plan.Filename, plan.Filename)
	fmt.Printf("Description: %s\n", plan.Description)
	fmt.Printf("Up Changes: %d\n", plan.Changes)
	fmt.Printf("Down Changes: %d\n", plan.DownChanges)

	if isDebug {
		fmt.Printf("\n=== UP SQL Content ===\n%s\n", plan.UpSQL)
		fmt.Printf("\n=== DOWN SQL Content ===\n%s\n", plan.DownSQL)
	}
}
