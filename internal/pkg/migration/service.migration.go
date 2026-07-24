package migration

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	database "vibe-ddd-golang/internal/pkg/db"

	"ariga.io/atlas/sql/migrate"
	"ariga.io/atlas/sql/mysql"
	"ariga.io/atlas/sql/postgres"
	"ariga.io/atlas/sql/schema"
	"gorm.io/gorm"
)

type SchemaMigrations struct {
	Version       string    `gorm:"column:version;primaryKey;type:varchar(255);not null"`
	Description   string    `gorm:"column:description;type:varchar(255);not null"`
	Applied       int64     `gorm:"column:applied;type:bigint;not null;default:0"`
	Total         int64     `gorm:"column:total;type:bigint;not null;default:0"`
	ExecutedAt    time.Time `gorm:"column:executed_at;type:timestamp;not null"`
	ExecutionTime int64     `gorm:"column:execution_time;type:bigint;not null"`
}

func (SchemaMigrations) TableName() string {
	return "schema_migrations"
}

type MigrationPlan struct {
	Filename    string
	Description string
	Changes     int
	DownChanges int
	UpSQL       string
	DownSQL     string
}

type Service struct {
	config     *Config
	db         *gorm.DB
	sqlDriver  migrate.Driver
	loadSchema func(ctx context.Context, config *Config) (*schema.Realm, error)
	tableList  []string
}

func NewService(
	db *gorm.DB, config *Config, tableList []string,
	loadSchemaFunc func(ctx context.Context, config *Config) (*schema.Realm, error),
) (*Service, error) {
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying DB: %w", err)
	}

	var driver migrate.Driver
	switch config.DbConfig.Driver {
	case database.POSTGRES:
		driver, err = postgres.Open(sqlDB)
		if err != nil {
			return nil, fmt.Errorf("failed to create PostgreSQL driver: %w", err)
		}
	case database.MYSQL:
		driver, err = mysql.Open(sqlDB)
		if err != nil {
			return nil, fmt.Errorf("failed to create MySQL driver: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported migration driver: %s", config.DbConfig.Driver)
	}

	if err := os.MkdirAll(config.MigrationsDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create migrations directory: %w", err)
	}

	return &Service{
		config:     config,
		db:         db,
		sqlDriver:  driver,
		loadSchema: loadSchemaFunc,
		tableList:  tableList,
	}, nil
}

func (s *Service) EnsureSchemaRevisionsTable() error {
	if !s.db.Migrator().HasTable(&SchemaMigrations{}) {
		createSQL := `CREATE TABLE IF NOT EXISTS schema_migrations (
			version varchar(255) NOT NULL,
			description varchar(255) NOT NULL,
			applied bigint NOT NULL DEFAULT 0,
			total bigint NOT NULL DEFAULT 0,
			executed_at timestamp NOT NULL,
			execution_time bigint NOT NULL,
			PRIMARY KEY (version)
		)`

		if err := s.db.Exec(createSQL).Error; err != nil {
			return fmt.Errorf("failed to create schema_migrations table: %w", err)
		}

		if s.config.IsDebug {
			log.Printf("Successfully created schema_migrations table")
		}
	}
	return nil
}

func (s *Service) RollbackToVersion(ctx context.Context, targetVersion string) error {
	var migrations []SchemaMigrations
	err := s.db.Where("version > ?", targetVersion).
		Order("version DESC").
		Find(&migrations).Error

	if err != nil {
		return fmt.Errorf("failed to get migrations to rollback: %w", err)
	}

	if len(migrations) == 0 {
		return fmt.Errorf("no migrations found after version %s", targetVersion)
	}

	for _, migration := range migrations {
		downFile := fmt.Sprintf("%s.down.sql", migration.Version)
		downFilePath := filepath.Join(s.config.MigrationsDir, downFile)

		if _, err := os.Stat(downFilePath); os.IsNotExist(err) {
			downFile = fmt.Sprintf("%s_%s.down.sql", migration.Version, migration.Description)
			downFilePath = filepath.Join(s.config.MigrationsDir, downFile)

			if _, err := os.Stat(downFilePath); os.IsNotExist(err) {
				return fmt.Errorf("down migration file not found for version %s", migration.Version)
			}
		}

		content, err := os.ReadFile(downFilePath)
		if err != nil {
			return fmt.Errorf("failed to read down migration file: %w", err)
		}

		log.Printf("Rolling back migration %s...", migration.Version)

		tx := s.db.Begin()

		statements := splitStatements(string(content))
		for i, stmt := range statements {
			if strings.TrimSpace(stmt) == "" {
				continue
			}

			if strings.HasPrefix(strings.TrimSpace(stmt), "--") {
				continue
			}

			if s.config.IsDebug {
				log.Printf("Executing rollback statement %d: %s", i+1, stmt)
			}

			if err := tx.WithContext(ctx).Exec(stmt).Error; err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to execute rollback statement %d for version %s: %w", i+1, migration.Version, err)
			}
		}

		if err := tx.Delete(&migration).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to delete migration record for version %s: %w", migration.Version, err)
		}

		if err := tx.Commit().Error; err != nil {
			return fmt.Errorf("failed to commit rollback transaction for version %s: %w", migration.Version, err)
		}

		log.Printf("Successfully rolled back migration %s", migration.Version)
	}

	log.Printf("Successfully rolled back to version %s", targetVersion)
	return nil
}

func (s *Service) checkMigrationExists() (bool, error) {
	var count int64
	err := s.db.Model(&SchemaMigrations{}).
		Count(&count).Error
	if err != nil {
		// Fail safe: an unreadable schema_migrations table must not trigger a
		// re-init, so report "exists".
		return true, nil //nolint:nilerr // deliberate fail-safe, see comment
	}
	return count > 0, nil
}

func (s *Service) checkExistingMigrations() (bool, error) {
	entries, err := os.ReadDir(s.config.MigrationsDir)
	if err != nil {
		return false, fmt.Errorf("failed to read migrations directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() && (strings.HasSuffix(entry.Name(), ".up.sql") || strings.HasSuffix(entry.Name(), ".down.sql")) {
			return true, nil
		}
	}

	return false, nil
}

func (s *Service) writeMigrationFiles(baseFilename string, upPlan, downPlan *migrate.Plan) error {
	upFilepath := filepath.Join(s.config.MigrationsDir, baseFilename+".up.sql")
	if err := s.writeMigrationFile(upFilepath, upPlan); err != nil {
		return fmt.Errorf("failed to write up migration: %w", err)
	}

	downFilepath := filepath.Join(s.config.MigrationsDir, baseFilename+".down.sql")
	if err := s.writeMigrationFile(downFilepath, downPlan); err != nil {
		return fmt.Errorf("failed to write down migration: %w", err)
	}

	return nil
}

func (s *Service) writeMigrationFile(filepath string, plan *migrate.Plan) error {
	file, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("failed to create migration file: %w", err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	defer writer.Flush()

	fmt.Fprintf(writer, "-- Auto-generated migration script\n")
	fmt.Fprintf(writer, "-- Generated at: %s\n\n", time.Now().UTC().Format(time.RFC3339))

	for _, change := range plan.Changes {
		fmt.Fprintf(writer, "-- %s\n", change.Comment)
		fmt.Fprintf(writer, "%s;\n\n", change.Cmd)
	}

	return nil
}

func (s *Service) removeDBNameFromSQL(sql string) string {
	var dbPrefix string
	if s.config.DbConfig.Driver == database.MYSQL {
		dbPrefix = fmt.Sprintf("`%s`.", s.config.DbConfig.Database)
	} else {
		dbPrefix = fmt.Sprintf("\"%s\".", s.config.DbConfig.Database)
		// Also check for public schema prefix which is common in Postgres
		sql = strings.ReplaceAll(sql, "\"public\".", "")
	}
	return strings.ReplaceAll(sql, dbPrefix, "")
}

func (s *Service) normalizePrimaryKeyStatement(sql string) string {
	if !strings.Contains(sql, "PRIMARY KEY") {
		return sql
	}

	if strings.Contains(sql, "DROP PRIMARY KEY") && strings.Contains(sql, "ADD PRIMARY KEY") {
		re := regexp.MustCompile(`ALTER TABLE \` + "`" + `([^` + "`" + `]+)\` + "`" + ``)
		matches := re.FindStringSubmatch(sql)
		if len(matches) < 2 {
			return sql
		}
		tableName := matches[1]

		re = regexp.MustCompile(`ADD PRIMARY KEY \(([^)]+)\)`)
		matches = re.FindStringSubmatch(sql)
		if len(matches) < 2 {
			return sql
		}

		columnsStr := matches[1]
		columns := strings.Split(columnsStr, ",")

		for i, col := range columns {
			col = strings.TrimSpace(col)
			col = strings.ReplaceAll(col, "`", "")
			columns[i] = col
		}

		sort.Strings(columns)

		var normalizedColumns []string
		for _, col := range columns {
			normalizedColumns = append(normalizedColumns, "`"+col+"`")
		}

		return fmt.Sprintf("ALTER TABLE `%s` DROP PRIMARY KEY, ADD PRIMARY KEY (%s)",
			tableName, strings.Join(normalizedColumns, ", "))
	}

	return sql
}

func (s *Service) removeDuplicateActions(upPlan, downPlan *migrate.Plan) (*migrate.Plan, *migrate.Plan) {
	var upChanges, downChanges migrate.Plan
	for i := range upPlan.Changes {
		if i >= len(downPlan.Changes) {
			break
		}

		upCmd := s.removeDBNameFromSQL(upPlan.Changes[i].Cmd)
		downCmd := s.removeDBNameFromSQL(downPlan.Changes[i].Cmd)

		upCmd = s.normalizePrimaryKeyStatement(upCmd)
		downCmd = s.normalizePrimaryKeyStatement(downCmd)

		if upCmd == downCmd {
			continue
		}

		upChanges.Changes = append(upChanges.Changes, &migrate.Change{
			Cmd:     upCmd,
			Comment: upPlan.Changes[i].Comment,
		})

		downChanges.Changes = append(downChanges.Changes, &migrate.Change{
			Cmd:     downCmd,
			Comment: downPlan.Changes[i].Comment,
		})
	}

	if len(upPlan.Changes) > len(downPlan.Changes) {
		for i := len(downPlan.Changes); i < len(upPlan.Changes); i++ {
			upCmd := s.removeDBNameFromSQL(upPlan.Changes[i].Cmd)
			upChanges.Changes = append(upChanges.Changes, &migrate.Change{
				Cmd:     upCmd,
				Comment: upPlan.Changes[i].Comment,
			})
		}
	} else if len(downPlan.Changes) > len(upPlan.Changes) {
		for i := len(upPlan.Changes); i < len(downPlan.Changes); i++ {
			downCmd := s.removeDBNameFromSQL(downPlan.Changes[i].Cmd)
			downChanges.Changes = append(downChanges.Changes, &migrate.Change{
				Cmd:     downCmd,
				Comment: downPlan.Changes[i].Comment,
			})
		}
	}

	return &upChanges, &downChanges
}

func (s *Service) createInitialMigration(ctx context.Context, name string, dryRun bool) (*MigrationPlan, error) {
	startTime := time.Now().UTC()

	tables := s.tableList

	schemaName := s.config.DbConfig.Database
	if s.config.DbConfig.Driver == database.POSTGRES {
		schemaName = "public"
	}

	currentSchema, err := s.sqlDriver.InspectSchema(ctx, schemaName, &schema.InspectOptions{
		Mode:   schema.InspectTables,
		Tables: tables,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to inspect schema: %w", err)
	}

	emptySchema := schema.New(s.config.DbConfig.Database)
	currentSchema.Name = s.config.DbConfig.Database

	if s.config.DbConfig.Driver == database.MYSQL {
		emptySchema.Attrs = []schema.Attr{
			&schema.Charset{V: "utf8mb4"},
			&schema.Collation{V: "utf8mb4_unicode_ci"},
		}
		currentSchema.Attrs = []schema.Attr{
			&schema.Charset{V: "utf8mb4"},
			&schema.Collation{V: "utf8mb4_unicode_ci"},
		}
	}

	emptyRealm := &schema.Realm{Schemas: []*schema.Schema{emptySchema}}
	currentRealm := &schema.Realm{Schemas: []*schema.Schema{currentSchema}}

	upChanges, err := s.sqlDriver.RealmDiff(emptyRealm, currentRealm, schema.DiffNormalized())
	if err != nil {
		return nil, fmt.Errorf("failed to generate UP migration: %w", err)
	}

	downChanges, err := s.sqlDriver.RealmDiff(currentRealm, emptyRealm, schema.DiffNormalized())
	if err != nil {
		return nil, fmt.Errorf("failed to generate DOWN migration: %w", err)
	}

	upPlan, err := s.sqlDriver.PlanChanges(ctx, "up", upChanges)
	if err != nil {
		return nil, fmt.Errorf("failed to generate UP migration plan: %w", err)
	}

	downPlan, err := s.sqlDriver.PlanChanges(ctx, "down", downChanges)
	if err != nil {
		return nil, fmt.Errorf("failed to generate DOWN migration plan: %w", err)
	}

	if len(upPlan.Changes) == 0 {
		return nil, fmt.Errorf("no changes detected for initial schema")
	}

	timestamp := time.Now().UTC().Format("20060102150405")
	baseName := "init_schema"
	if name != "" {
		baseName = name
	}
	baseFilename := fmt.Sprintf("%s_%s", timestamp, baseName)

	upPlan, downPlan = s.removeDuplicateActions(upPlan, downPlan)
	if len(upPlan.Changes) == 0 || len(downPlan.Changes) == 0 {
		log.Println("No changes detected")
		return nil, nil //nolint:nilnil // nil plan without error means "no changes to migrate"
	}

	var upSQLContent strings.Builder
	for _, change := range upPlan.Changes {
		upSQLContent.WriteString(fmt.Sprintf("-- %s\n", change.Comment))
		upSQLContent.WriteString(fmt.Sprintf("%s;\n\n", change.Cmd))
	}

	var downSQLContent strings.Builder
	for _, change := range downPlan.Changes {
		downSQLContent.WriteString(fmt.Sprintf("-- %s\n", change.Comment))
		downSQLContent.WriteString(fmt.Sprintf("%s;\n\n", change.Cmd))
	}

	migrationPlan := &MigrationPlan{
		Filename:    baseFilename,
		Description: baseName,
		Changes:     len(upPlan.Changes),
		DownChanges: len(downPlan.Changes),
		UpSQL:       upSQLContent.String(),
		DownSQL:     downSQLContent.String(),
	}

	if dryRun {
		return migrationPlan, nil
	}

	if err := s.writeMigrationFiles(baseFilename, upPlan, downPlan); err != nil {
		return nil, fmt.Errorf("failed to write migration files: %w", err)
	}

	if s.config.IsDebug {
		log.Printf("Initial migration files generated successfully in %v", time.Since(startTime))
	}

	return migrationPlan, nil
}

func (s *Service) ShowMigrationStatus() error {
	var appliedMigrations []SchemaMigrations
	err := s.db.Order("version ASC").Find(&appliedMigrations).Error
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	appliedMap := make(map[string]SchemaMigrations)
	for _, migration := range appliedMigrations {
		appliedMap[migration.Version] = migration
	}

	entries, err := os.ReadDir(s.config.MigrationsDir)
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	type MigrationStatus struct {
		Version       string
		Description   string
		Status        string
		Applied       time.Time
		ExecutionTime int64
	}

	var allMigrations []MigrationStatus
	var pendingCount, appliedCount int

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".up.sql") {
			parts := strings.Split(entry.Name(), "_")
			if len(parts) < 2 {
				continue
			}

			version := parts[0]
			description := strings.TrimSuffix(strings.Join(parts[1:], "_"), ".up.sql")

			var status string
			var appliedTime time.Time
			var executionTime int64

			if migration, exists := appliedMap[version]; exists {
				status = "DONE"
				appliedTime = migration.ExecutedAt
				executionTime = migration.ExecutionTime
				appliedCount++
			} else {
				status = "PENDING"
				pendingCount++
			}

			allMigrations = append(allMigrations, MigrationStatus{
				Version:       version,
				Description:   description,
				Status:        status,
				Applied:       appliedTime,
				ExecutionTime: executionTime,
			})
		}
	}

	sort.Slice(allMigrations, func(i, j int) bool {
		return allMigrations[i].Version < allMigrations[j].Version
	})

	fmt.Println("\n=== Migration Status ===")
	fmt.Printf("Total: %d | Applied: %d | Pending: %d\n\n", len(allMigrations), appliedCount, pendingCount)

	fmt.Printf("%-15s %-30s %-10s %-25s %-12s\n", "VERSION", "DESCRIPTION", "STATUS", "APPLIED AT", "DURATION")
	fmt.Println(strings.Repeat("-", 90))

	for _, migration := range allMigrations {
		var appliedStr string
		var durationStr string

		if migration.Status == "DONE" {
			appliedStr = migration.Applied.Format("2006-01-02 15:04:05")
			durationStr = fmt.Sprintf("%d ms", migration.ExecutionTime)
		} else {
			appliedStr = "-"
			durationStr = "-"
		}

		fmt.Printf("%-15s %-30s %-10s %-25s %-12s\n",
			migration.Version,
			migration.Description,
			migration.Status,
			appliedStr,
			durationStr,
		)
	}

	fmt.Println()

	if appliedCount > 0 {
		var lastMigration SchemaMigrations
		err := s.db.Order("version DESC").First(&lastMigration).Error
		if err == nil {
			fmt.Printf("Last applied migration: %s (%s) at %s\n",
				lastMigration.Version,
				lastMigration.Description,
				lastMigration.ExecutedAt.Format("2006-01-02 15:04:05"),
			)
		}
	}

	if pendingCount > 0 {
		fmt.Printf("\nYou have %d pending migrations. Run with --apply to apply them.\n", pendingCount)
	} else {
		fmt.Println("\nAll migrations are up to date.")
	}

	return nil
}

func (s *Service) GenerateMigrationDiff(ctx context.Context, name string, dryRun bool) (*MigrationPlan, error) {
	desiredRealm, err := s.loadSchema(ctx, s.config)
	if err != nil {
		return nil, fmt.Errorf("failed to load GORM schema: %w", err)
	}

	schemaName := s.config.DbConfig.Database
	if s.config.DbConfig.Driver == database.POSTGRES {
		schemaName = "public"
	}

	currentSchema, err := s.sqlDriver.InspectSchema(ctx, schemaName, &schema.InspectOptions{
		Mode:   schema.InspectTables,
		Tables: s.tableList,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to inspect current schema: %w", err)
	}

	currentRealm := schema.NewRealm(currentSchema)

	if len(desiredRealm.Schemas) > 0 && len(currentRealm.Schemas) > 0 {
		desiredRealm.Schemas[0].Name = schemaName
		currentRealm.Schemas[0].Name = schemaName
	}

	upChanges, err := s.sqlDriver.RealmDiff(currentRealm, desiredRealm, schema.DiffNormalized())
	if err != nil {
		return nil, fmt.Errorf("failed to calculate up diff: %w", err)
	}

	downChanges, err := s.sqlDriver.RealmDiff(desiredRealm, currentRealm, schema.DiffNormalized())
	if err != nil {
		return nil, fmt.Errorf("failed to calculate down diff: %w", err)
	}

	if len(upChanges) == 0 {
		log.Println("No changes detected")
		return nil, nil //nolint:nilnil // nil plan without error means "no changes to migrate"
	}

	description := "schema_update"
	if name != "" {
		description = name
	}

	upPlan, err := s.sqlDriver.PlanChanges(ctx, description, upChanges)
	if err != nil {
		return nil, fmt.Errorf("failed to plan up changes: %w", err)
	}

	downPlan, err := s.sqlDriver.PlanChanges(ctx, description, downChanges)
	if err != nil {
		return nil, fmt.Errorf("failed to plan down changes: %w", err)
	}

	timestamp := time.Now().UTC().Format("20060102150405")
	baseFilename := fmt.Sprintf("%s_%s", timestamp, description)

	upPlan, downPlan = s.removeDuplicateActions(upPlan, downPlan)
	if len(upPlan.Changes) == 0 || len(downPlan.Changes) == 0 {
		log.Println("No changes detected after removing duplicates")
		return nil, nil //nolint:nilnil // nil plan without error means "no changes to migrate"
	}

	var upSQLContent strings.Builder
	for _, change := range upPlan.Changes {
		upSQLContent.WriteString(fmt.Sprintf("-- %s\n", change.Comment))
		upSQLContent.WriteString(fmt.Sprintf("%s;\n\n", change.Cmd))
	}

	var downSQLContent strings.Builder
	for _, change := range downPlan.Changes {
		downSQLContent.WriteString(fmt.Sprintf("-- %s\n", change.Comment))
		downSQLContent.WriteString(fmt.Sprintf("%s;\n\n", change.Cmd))
	}

	migrationPlan := &MigrationPlan{
		Filename:    baseFilename,
		Description: description,
		Changes:     len(upPlan.Changes),
		DownChanges: len(downPlan.Changes),
		UpSQL:       upSQLContent.String(),
		DownSQL:     downSQLContent.String(),
	}

	if dryRun {
		return migrationPlan, nil
	}

	if err := s.writeMigrationFiles(baseFilename, upPlan, downPlan); err != nil {
		return nil, fmt.Errorf("failed to write migration files: %w", err)
	}

	return migrationPlan, nil
}

func (s *Service) GenerateInitialMigration(ctx context.Context, name string, dryRun bool, force bool) (*MigrationPlan, error) {
	hasMigrations, err := s.checkExistingMigrations()
	if err != nil {
		return nil, fmt.Errorf("failed to check existing migrations: %w", err)
	}

	if !hasMigrations || force {
		existsMigrate, err := s.checkMigrationExists()
		if err != nil {
			return nil, err
		}

		if !existsMigrate || force {
			return s.createInitialMigration(ctx, name, dryRun)
		}
		return nil, fmt.Errorf("database already has migrations. Use --force to override")
	}

	return nil, fmt.Errorf("migrations directory already has files. Use --force to override")
}

func (s *Service) ApplyPendingMigrations(ctx context.Context, count int, force bool) error {
	var lastRevision SchemaMigrations
	err := s.db.Where("1=1").
		Order("version DESC").
		First(&lastRevision).Error

	var lastVersion string
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Println("No migrations found in database, applying all migrations")
			lastVersion = ""
		} else {
			return fmt.Errorf("failed to get last migration: %w", err)
		}
	} else {
		lastVersion = lastRevision.Version
		log.Printf("Last applied migration: %s", lastVersion)
	}

	entries, err := os.ReadDir(s.config.MigrationsDir)
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	type MigrationFile struct {
		Version     string
		Description string
		Filename    string
		UpFilePath  string
	}

	var migrations []MigrationFile
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".up.sql") {
			parts := strings.Split(entry.Name(), "_")
			if len(parts) < 2 {
				continue
			}

			version := parts[0]
			description := strings.TrimSuffix(strings.Join(parts[1:], "_"), ".up.sql")

			if version > lastVersion {
				migrations = append(migrations, MigrationFile{
					Version:     version,
					Description: description,
					Filename:    entry.Name(),
					UpFilePath:  filepath.Join(s.config.MigrationsDir, entry.Name()),
				})
			}
		}
	}

	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	if len(migrations) == 0 {
		log.Println("No pending migrations to apply")
		return nil
	}

	if count > 0 && count < len(migrations) {
		log.Printf("Found %d pending migrations, but only applying %d as requested", len(migrations), count)
		migrations = migrations[:count]
	} else if count > len(migrations) {
		log.Printf("Requested to apply %d migrations, but only %d are pending", count, len(migrations))
	}

	log.Printf("Applying %d migrations", len(migrations))

	for _, migration := range migrations {
		log.Printf("Applying migration %s (%s)...", migration.Version, migration.Description)

		migrationStartTime := time.Now().UTC()

		content, err := os.ReadFile(migration.UpFilePath)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", migration.Filename, err)
		}

		tx := s.db.Begin()
		defer func() {
			if r := recover(); r != nil {
				tx.Rollback()
				panic(r)
			}
		}()

		statements := splitStatements(string(content))
		statementsExecuted := 0
		statementsTotal := 0

		for i, stmt := range statements {
			stmt = strings.TrimSpace(stmt)
			if stmt == "" {
				continue
			}

			if strings.HasPrefix(stmt, "--") {
				continue
			}

			statementsTotal++

			if s.config.IsDebug {
				log.Printf("Applying statement %d: %s", i+1, stmt)
			}

			err := tx.WithContext(ctx).Exec(stmt).Error
			if err != nil {
				if force {
					log.Printf("Warning: Statement %d failed (continuing with --force): %v", i+1, err)
					continue
				} else {
					tx.Rollback()
					return fmt.Errorf("failed to execute statement %d in migration %s: %w", i+1, migration.Version, err)
				}
			}
			statementsExecuted++
		}

		executionTime := time.Since(migrationStartTime).Milliseconds()

		if statementsExecuted == 0 && force && statementsTotal > 0 {
			log.Printf("Warning: All statements failed for migration %s, but continuing with --force", migration.Version)

			tx.Rollback()

			tx = s.db.Begin()

			revision := SchemaMigrations{
				Version:       migration.Version,
				Description:   migration.Description,
				Applied:       0,
				Total:         int64(statementsTotal),
				ExecutedAt:    time.Now().UTC(),
				ExecutionTime: executionTime,
			}

			if err := tx.Create(&revision).Error; err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to insert migration record for version %s: %w", migration.Version, err)
			}

			if err := tx.Commit().Error; err != nil {
				return fmt.Errorf("failed to commit migration record for version %s: %w", migration.Version, err)
			}

			log.Printf("Migration %s tracked with 0 applied statements (total: %d statements failed)", migration.Version, statementsTotal)
		} else {
			revision := SchemaMigrations{
				Version:       migration.Version,
				Description:   migration.Description,
				Applied:       int64(statementsExecuted),
				Total:         int64(statementsTotal),
				ExecutedAt:    time.Now().UTC(),
				ExecutionTime: executionTime,
			}

			if err := tx.Create(&revision).Error; err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to insert migration record for version %s: %w", migration.Version, err)
			}

			if err := tx.Commit().Error; err != nil {
				return fmt.Errorf("failed to commit migration %s: %w", migration.Version, err)
			}

			if force && statementsExecuted < statementsTotal {
				log.Printf("Migration %s partially applied: %d/%d statements executed in %dms",
					migration.Version, statementsExecuted, statementsTotal, executionTime)
			} else {
				log.Printf("Successfully applied migration %s (%d statements executed in %dms)",
					migration.Version, statementsExecuted, executionTime)
			}
		}
	}

	log.Printf("Successfully applied %d migrations", len(migrations))
	return nil
}

func (s *Service) ApplyMigration(ctx context.Context, plan *MigrationPlan, force bool) error {
	migrationStartTime := time.Now().UTC()

	content, err := os.ReadFile(filepath.Join(s.config.MigrationsDir, plan.Filename+".up.sql"))
	if err != nil {
		return fmt.Errorf("failed to read UP migration file: %w", err)
	}

	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	statements := splitStatements(string(content))
	statementsExecuted := 0
	statementsTotal := 0

	for i, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}
		if strings.HasPrefix(stmt, "--") {
			continue
		}

		statementsTotal++

		if s.config.IsDebug {
			log.Printf("Applying statement %d: %s", i+1, stmt)
		}

		err := tx.WithContext(ctx).Exec(stmt).Error
		if err != nil {
			if force {
				log.Printf("Warning: Statement %d failed (continuing with --force): %v", i+1, err)
				continue
			} else {
				tx.Rollback()
				return fmt.Errorf("failed to execute statement %d: %w", i+1, err)
			}
		}
		statementsExecuted++
	}

	executionTime := time.Since(migrationStartTime).Milliseconds()

	if statementsExecuted == 0 && force && statementsTotal > 0 {
		log.Printf("Warning: All statements failed for migration, but continuing with --force")

		tx.Rollback()

		tx = s.db.Begin()

		timestamp := plan.Filename[:14]
		revision := SchemaMigrations{
			Version:       timestamp,
			Description:   plan.Description,
			Applied:       0,
			Total:         int64(statementsTotal),
			ExecutedAt:    time.Now().UTC(),
			ExecutionTime: executionTime,
		}

		if err := tx.Create(&revision).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to insert migration record: %w", err)
		}

		if err := tx.Commit().Error; err != nil {
			return fmt.Errorf("failed to commit migration record: %w", err)
		}

		log.Printf("Migration tracked with 0 applied statements (total: %d statements failed)", statementsTotal)
		return nil
	}

	timestamp := plan.Filename[:14]
	revision := SchemaMigrations{
		Version:       timestamp,
		Description:   plan.Description,
		Applied:       int64(statementsExecuted),
		Total:         int64(statementsTotal),
		ExecutedAt:    time.Now().UTC(),
		ExecutionTime: executionTime,
	}

	if err := tx.Create(&revision).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to insert migration record: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit migration transaction: %w", err)
	}

	if force && statementsExecuted < statementsTotal {
		log.Printf("Migration partially applied: %d/%d statements executed in %dms",
			statementsExecuted, statementsTotal, executionTime)
	} else {
		log.Printf("Migration executed and recorded in %dms", executionTime)
	}
	return nil
}

func (s *Service) MarkMigrationsAsApplied(versionFlag string, force bool) error {
	log.Println("Marking migrations as applied (baseline mode)")

	entries, err := os.ReadDir(s.config.MigrationsDir)
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	var existingMigrations []SchemaMigrations
	if err := s.db.Find(&existingMigrations).Error; err != nil {
		return fmt.Errorf("failed to get existing migrations: %w", err)
	}

	existingVersions := make(map[string]bool)
	for _, migration := range existingMigrations {
		existingVersions[migration.Version] = true
	}

	type MigrationFile struct {
		Version     string
		Description string
		UpFilePath  string
	}

	var migrationsToBaseline []MigrationFile
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".up.sql") {
			parts := strings.Split(entry.Name(), "_")
			if len(parts) < 2 {
				continue
			}

			version := parts[0]
			description := strings.TrimSuffix(strings.Join(parts[1:], "_"), ".up.sql")

			if existingVersions[version] && !force {
				continue
			}

			if versionFlag != "" && version > versionFlag {
				continue
			}

			migrationsToBaseline = append(migrationsToBaseline, MigrationFile{
				Version:     version,
				Description: description,
				UpFilePath:  filepath.Join(s.config.MigrationsDir, entry.Name()),
			})
		}
	}

	sort.Slice(migrationsToBaseline, func(i, j int) bool {
		return migrationsToBaseline[i].Version < migrationsToBaseline[j].Version
	})

	if len(migrationsToBaseline) == 0 {
		log.Println("No migrations to baseline")
		return nil
	}

	log.Printf("Found %d migrations to mark as baseline", len(migrationsToBaseline))

	tx := s.db.Begin()
	if err := tx.Error; err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	for _, migration := range migrationsToBaseline {
		content, err := os.ReadFile(migration.UpFilePath)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to read migration file %s: %w", migration.UpFilePath, err)
		}

		statements := splitStatements(string(content))
		statementCount := 0
		for _, stmt := range statements {
			if strings.TrimSpace(stmt) != "" && !strings.HasPrefix(strings.TrimSpace(stmt), "--") {
				statementCount++
			}
		}

		if existingVersions[migration.Version] && force {
			if err := tx.Where("version = ?", migration.Version).Delete(&SchemaMigrations{}).Error; err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to delete existing migration record %s: %w", migration.Version, err)
			}
		}

		record := SchemaMigrations{
			Version:       migration.Version,
			Description:   migration.Description,
			Applied:       int64(statementCount),
			Total:         int64(statementCount),
			ExecutedAt:    time.Now().UTC(),
			ExecutionTime: 0,
		}

		if err := tx.Create(&record).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to create migration record for %s: %w", migration.Version, err)
		}

		if s.config.IsDebug {
			log.Printf("Marked migration %s (%s) as applied", migration.Version, migration.Description)
		}
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Printf("Successfully marked %d migrations as applied (baseline)", len(migrationsToBaseline))
	return nil
}

func (s *Service) RollbackLastMigration(ctx context.Context) error {
	var lastRevision SchemaMigrations
	err := s.db.Order("version DESC").
		First(&lastRevision).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("no migrations to rollback")
		}
		return fmt.Errorf("failed to get last migration: %w", err)
	}

	downFile := fmt.Sprintf("%s.down.sql", lastRevision.Version)
	downFilePath := filepath.Join(s.config.MigrationsDir, downFile)

	if _, err := os.Stat(downFilePath); os.IsNotExist(err) {
		downFile = fmt.Sprintf("%s_%s.down.sql", lastRevision.Version, lastRevision.Description)
		downFilePath = filepath.Join(s.config.MigrationsDir, downFile)

		if _, err := os.Stat(downFilePath); os.IsNotExist(err) {
			return fmt.Errorf("down migration file not found for version %s", lastRevision.Version)
		}
	}

	content, err := os.ReadFile(downFilePath)
	if err != nil {
		return fmt.Errorf("failed to read down migration file: %w", err)
	}

	log.Printf("Rolling back migration %s...", lastRevision.Version)

	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	statements := splitStatements(string(content))
	for i, stmt := range statements {
		if strings.TrimSpace(stmt) == "" {
			continue
		}
		if strings.HasPrefix(strings.TrimSpace(stmt), "--") {
			continue
		}

		if s.config.IsDebug {
			log.Printf("Executing rollback statement %d: %s", i+1, stmt)
		}

		if err := tx.WithContext(ctx).Exec(stmt).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to execute rollback statement %d: %w", i+1, err)
		}
	}

	if err := tx.Delete(&lastRevision).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete migration record: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit rollback transaction: %w", err)
	}

	log.Printf("Successfully rolled back migration %s", lastRevision.Version)
	return nil
}
