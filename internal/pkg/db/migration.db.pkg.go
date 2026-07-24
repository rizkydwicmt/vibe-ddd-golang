package database

import (
	"fmt"
	"strings"

	"gorm.io/gorm"

	"vibe-ddd-golang/internal/common/enum"
)

// guardAutoMigrate refuses to run GORM AutoMigrate when APP_ENV is
// production or staging. Atlas (make migrate-diff) is the only allowed
// path for non-dev schema changes.
func guardAutoMigrate(environment enum.EnvEnum) error {
	env := strings.ToLower(strings.TrimSpace(environment.ToString()))
	if env == enum.PRODUCTION.ToString() || env == enum.STAGING.ToString() {
		return fmt.Errorf("AutoMigrate forbidden in APP_ENV=%s — use Atlas (make migrate-diff)", env)
	}
	return nil
}

func (db *Database) Migrate(entities ...any) error {
	if err := guardAutoMigrate(db.Config.Environment); err != nil {
		return err
	}
	err := db.Migrator().AutoMigrate(entities...)
	if err != nil {
		return fmt.Errorf("error migrating database: %w", err)
	}

	db.DropUnusedColumns(entities...)
	return nil
}

func (db *Database) MigrateWithEnums(enums []EnumDefinition, entities ...any) error {
	if err := guardAutoMigrate(db.Config.Environment); err != nil {
		return err
	}
	// Create enum types first
	err := db.createEnumTypes(enums)
	if err != nil {
		return fmt.Errorf("error creating enum types: %w", err)
	}

	err = db.Migrator().AutoMigrate(entities...)
	if err != nil {
		return fmt.Errorf("error migrating database: %w", err)
	}

	db.DropUnusedColumns(entities...)
	return nil
}

func (db *Database) DropUnusedColumns(entities ...any) {
	for _, entity := range entities {
		stmt := &gorm.Statement{DB: db.DB}
		stmt.Parse(entity)
		fields := stmt.Schema.Fields
		columns, _ := db.Debug().Migrator().ColumnTypes(entity)

		for i := range columns {
			found := false
			for j := range fields {
				if columns[i].Name() == fields[j].DBName {
					found = true
					break
				}
			}
			if !found {
				db.Migrator().DropColumn(entity, columns[i].Name())
			}
		}
	}
}

// EnumDefinition represents a PostgreSQL enum type definition
type EnumDefinition struct {
	Name   string
	Values []string
}

// CreateEnumStatements generates SQL CREATE TYPE statements for PostgreSQL enum types
func CreateEnumStatements(enums []EnumDefinition) ([]string, error) {
	var statements []string

	for _, enumDef := range enums {
		// Validate enum definition
		if enumDef.Name == "" {
			return nil, fmt.Errorf("enum name cannot be empty")
		}
		if len(enumDef.Values) == 0 {
			return nil, fmt.Errorf("enum %s must have at least one value", enumDef.Name)
		}

		// Extract type name (strip schema if present, use current search_path)
		typeName := enumDef.Name
		if strings.Contains(enumDef.Name, ".") {
			parts := strings.SplitN(enumDef.Name, ".", 2)
			typeName = parts[1]
		}

		// Escape single quotes in enum values
		escapedValues := make([]string, len(enumDef.Values))
		for i, val := range enumDef.Values {
			escapedValues[i] = strings.ReplaceAll(val, "'", "''")
		}

		// Generate idempotent CREATE ENUM statement using DO block
		// Uses current schema from search_path
		valuesList := "'" + strings.Join(escapedValues, "', '") + "'"
		sql := fmt.Sprintf(`DO $$ BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_type t
        JOIN pg_namespace n ON n.oid = t.typnamespace
        WHERE n.nspname = current_schema() AND t.typname = '%s'
    ) THEN
        CREATE TYPE %s AS ENUM (%s);
    END IF;
END $$;`, typeName, typeName, valuesList)
		statements = append(statements, sql)
	}

	return statements, nil
}

// createEnumTypes creates PostgreSQL enum types from provided definitions
func (db *Database) createEnumTypes(enums []EnumDefinition) error {
	for _, enumDef := range enums {
		err := db.createEnumIfNotExists(enumDef)
		if err != nil {
			return err
		}
	}

	return nil
}

// createEnumIfNotExists creates a PostgreSQL enum type if it doesn't exist, or updates it safely
func (db *Database) createEnumIfNotExists(enumDef EnumDefinition) error {
	// Parse schema and type name
	var schemaName, typeName string
	if strings.Contains(enumDef.Name, ".") {
		parts := strings.SplitN(enumDef.Name, ".", 2)
		schemaName = parts[0]
		typeName = parts[1]
	} else {
		schemaName = "public"
		typeName = enumDef.Name
	}

	// Check if enum already exists in the specified schema
	var exists bool
	err := db.Raw(`
		SELECT EXISTS(
			SELECT 1 FROM pg_type t
			JOIN pg_namespace n ON n.oid = t.typnamespace
			WHERE n.nspname = ? AND t.typname = ?
		)`, schemaName, typeName).Scan(&exists).Error
	if err != nil {
		return fmt.Errorf("error checking enum existence for %s: %w", enumDef.Name, err)
	}

	if !exists {
		// Create enum type
		valuesList := "'" + strings.Join(enumDef.Values, "', '") + "'"
		sql := fmt.Sprintf("CREATE TYPE %s AS ENUM (%s)", enumDef.Name, valuesList)
		err = db.Exec(sql).Error
		if err != nil {
			return fmt.Errorf("error creating enum %s: %w", enumDef.Name, err)
		}
	} else {
		// Enum exists, check if we need to add new values
		err = db.updateEnumValues(enumDef)
		if err != nil {
			return fmt.Errorf("error updating enum %s: %w", enumDef.Name, err)
		}
	}

	return nil
}

// updateEnumValues safely updates enum values by adding new ones (PostgreSQL doesn't support removing enum values safely)
func (db *Database) updateEnumValues(enumDef EnumDefinition) error {
	// Parse schema and type name
	var schemaName, typeName string
	if strings.Contains(enumDef.Name, ".") {
		parts := strings.SplitN(enumDef.Name, ".", 2)
		schemaName = parts[0]
		typeName = parts[1]
	} else {
		schemaName = "public"
		typeName = enumDef.Name
	}

	// Get current enum values using schema-qualified name
	var currentValues []string
	qualifiedName := schemaName + "." + typeName
	err := db.Raw(`
		SELECT unnest(enum_range(NULL::`+qualifiedName+`))::text AS enum_value
	`).Pluck("enum_value", &currentValues).Error
	if err != nil {
		return fmt.Errorf("error getting current enum values for %s: %w", enumDef.Name, err)
	}

	// Find new values to add
	currentValuesMap := make(map[string]bool)
	for _, val := range currentValues {
		currentValuesMap[val] = true
	}

	// Add new values that don't exist
	for _, newValue := range enumDef.Values {
		if !currentValuesMap[newValue] {
			sql := fmt.Sprintf("ALTER TYPE %s ADD VALUE '%s'", qualifiedName, newValue)
			err = db.Exec(sql).Error
			if err != nil {
				return fmt.Errorf("error adding enum value '%s' to %s: %w", newValue, enumDef.Name, err)
			}
		}
	}

	return nil
}

// DropEnumSafely drops an enum type if no tables are using it
func (db *Database) DropEnumSafely(enumName string) error {
	// Check if any columns are using this enum
	var count int64
	err := db.Raw(`
		SELECT COUNT(*)
			FROM information_schema.columns
		WHERE udt_name = ?
	`, enumName).Scan(&count).Error
	if err != nil {
		return fmt.Errorf("error checking enum usage for %s: %w", enumName, err)
	}

	if count == 0 {
		// Safe to drop
		sql := fmt.Sprintf("DROP TYPE IF EXISTS %s", enumName)
		err = db.Exec(sql).Error
		if err != nil {
			return fmt.Errorf("error dropping enum %s: %w", enumName, err)
		}
	}

	return nil
}

// NewEnum creates a new EnumDefinition
func NewEnum(name string, values ...string) EnumDefinition {
	return EnumDefinition{
		Name:   name,
		Values: values,
	}
}

func (db *Database) MigrateEnumValueSafely(tableName, columnName, oldValue, newValue string) error {
	sql := fmt.Sprintf("UPDATE %s SET %s = ? WHERE %s = ?",
		tableName, columnName, columnName)

	err := db.Exec(sql, newValue, oldValue).Error
	if err != nil {
		return fmt.Errorf("error migrating enum value from '%s' to '%s' in %s.%s: %w",
			oldValue, newValue, tableName, columnName, err)
	}

	return nil
}
