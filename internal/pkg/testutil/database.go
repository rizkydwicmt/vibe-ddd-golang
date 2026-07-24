package testutil

import (
	"vibe-ddd-golang/internal/application/payment/entity"
	userEntity "vibe-ddd-golang/internal/application/user/entity"
	database "vibe-ddd-golang/internal/pkg/db"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// SetupTestDB creates an in-memory SQLite database for testing
func SetupTestDB() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, err
	}

	// Auto-migrate all entities
	err = db.AutoMigrate(
		&userEntity.User{},
		&entity.Payment{},
	)
	if err != nil {
		return nil, err
	}

	return db, nil
}

// SetupTestDatabase wraps an in-memory SQLite gorm.DB in the *db.Database type that
// repositories inject via params.Params.MainDB. cursorCrypto stays nil (unused by
// the example repositories).
func SetupTestDatabase() (*database.Database, error) {
	gdb, err := SetupTestDB()
	if err != nil {
		return nil, err
	}
	return &database.Database{DB: gdb}, nil
}

// CleanDB cleans all data from test database
func CleanDB(db *gorm.DB) error {
	// Delete in reverse order of dependencies
	if err := db.Exec("DELETE FROM payments").Error; err != nil {
		return err
	}
	if err := db.Exec("DELETE FROM users").Error; err != nil {
		return err
	}
	return nil
}
