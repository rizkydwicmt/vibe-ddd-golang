package database

import (
	"fmt"
	"time"

	"vibe-ddd-golang/internal/common/enum"
	helpercrypto "vibe-ddd-golang/internal/pkg/helper/crypto"
	"vibe-ddd-golang/internal/pkg/redis"

	"github.com/go-gorm/caches/v4"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	_logger "gorm.io/gorm/logger"
)

type Config struct {
	Host            string
	Port            int
	User            string
	Password        string
	Database        string
	SSLMode         string
	Timezone        string
	URL             string
	Driver          DriverEnum
	Cache           bool
	Rds             *redis.Client
	CacheTime       time.Duration
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
	Environment     enum.EnvEnum
}

type Database struct {
	*gorm.DB
	Config       *Config
	cursorCrypto *helpercrypto.CursorCrypto
}

func Setup(cfg *Config) (*Database, error) {
	var db *gorm.DB
	var err error
	crypto, err := helpercrypto.NewCursorCrypto(cfg.User + cfg.Password + cfg.Database)
	if err != nil {
		return nil, err
	}

	logLevel := _logger.Info
	if cfg.Environment == enum.PRODUCTION {
		logLevel = _logger.Silent
	}

	gormConfig := &gorm.Config{
		Logger: _logger.Default.LogMode(logLevel),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
		TranslateError: true,
	}

	switch cfg.Driver {
	case POSTGRES:
		dsn := cfg.URL
		if dsn == "" {
			tz := cfg.Timezone
			if tz == "" {
				tz = "UTC"
			}
			dsn = fmt.Sprintf(
				"host=%s user=%s password=%s dbname=%s port=%d timezone=%s sslmode=%s",
				cfg.Host,
				cfg.User,
				cfg.Password,
				cfg.Database,
				cfg.Port,
				tz,
				cfg.SSLMode,
			)
		}
		db, err = gorm.Open(postgres.Open(dsn), gormConfig)

	case MYSQL:
		dsn := cfg.URL
		if dsn == "" {
			dsn = fmt.Sprintf(
				"%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=UTC",
				cfg.User,
				cfg.Password,
				cfg.Host,
				cfg.Port,
				cfg.Database,
			)
		}
		db, err = gorm.Open(mysql.Open(dsn), gormConfig)
	default:
		return nil, fmt.Errorf("unsupported database driver: %s (supported: postgres, mysql)", cfg.Driver)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if cfg.Cache {
		var cachesPlugin *caches.Caches
		if cfg.Rds != nil && cfg.CacheTime > 0 {
			cachesPlugin = &caches.Caches{Conf: &caches.Config{
				Cacher: &redisCacher{
					rdb:       cfg.Rds.Client,
					cacheTime: cfg.CacheTime,
				},
			}}
		} else {
			cachesPlugin = &caches.Caches{Conf: &caches.Config{
				Cacher: &memoryCacher{},
			}}
		}

		_ = db.Use(cachesPlugin)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	maxOpenConns := cfg.MaxOpenConns
	if maxOpenConns == 0 {
		maxOpenConns = 20
	}
	sqlDB.SetMaxOpenConns(maxOpenConns)

	maxIdleConns := cfg.MaxIdleConns
	if maxIdleConns == 0 {
		maxIdleConns = 5
	}
	sqlDB.SetMaxIdleConns(maxIdleConns)

	if cfg.ConnMaxLifetime > 0 {
		sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	} else {
		sqlDB.SetConnMaxLifetime(0)
	}

	if cfg.ConnMaxIdleTime > 0 {
		sqlDB.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)
	} else {
		sqlDB.SetConnMaxIdleTime(30 * time.Minute)
	}

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &Database{
		db,
		cfg,
		crypto,
	}, nil
}

func (db *Database) Close() error {
	sqlDB, err := db.DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get database instance: %w", err)
	}
	return sqlDB.Close()
}
