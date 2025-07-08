package database

import (
	"fmt"

	"vibe-ddd-golang/internal/application/payment/entity"
	userEntity "vibe-ddd-golang/internal/application/user/entity"
	"vibe-ddd-golang/internal/config"

	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func NewDatabase(cfg *config.Config, log *zap.Logger) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s",
		cfg.Database.Host,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.DBName,
		cfg.Database.Port,
		cfg.Database.SSLMode,
	)

	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	}

	db, err := gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		log.Error("Failed to connect to database", zap.Error(err))
		return nil, err
	}

	err = db.AutoMigrate(
		&userEntity.User{},
		&entity.Payment{},
	)
	if err != nil {
		log.Error("Failed to migrate database", zap.Error(err))
		return nil, err
	}

	log.Info("Database connected and migrated successfully")
	return db, nil
}
