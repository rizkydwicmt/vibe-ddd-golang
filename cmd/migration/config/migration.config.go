package config

import (
	"fmt"

	appconfig "vibe-ddd-golang/internal/config"
	database "vibe-ddd-golang/internal/pkg/db"
	"vibe-ddd-golang/internal/pkg/migration"
	migration2 "vibe-ddd-golang/internal/server/migration"
)

func NewConfigEnv(devDSN string, verbose bool) (*migration.Config, error) {
	conf, err := appconfig.NewConfig()
	if err != nil {
		return nil, fmt.Errorf("error getting config: %v", err)
	}

	driver := database.DriverEnum(conf.Database.Driver)

	statements, err := migration2.MigrationServer(driver)
	if err != nil {
		return nil, fmt.Errorf("error creating migration server: %v", err)
	}

	tableList := migration2.ListTableEntity()
	dbConfig := &database.Config{
		Database:        conf.Database.DBName,
		Host:            conf.Database.Host,
		Port:            conf.Database.Port,
		User:            conf.Database.User,
		Password:        conf.Database.Password,
		SSLMode:         conf.Database.SSLMode,
		Timezone:        conf.Database.Timezone,
		Driver:          driver,
		URL:             devDSN,
		Cache:           false,
		MaxOpenConns:    conf.Database.MaxOpenConns,
		MaxIdleConns:    conf.Database.MaxIdleConns,
		ConnMaxLifetime: conf.Database.ConnMaxLifetime,
		ConnMaxIdleTime: conf.Database.ConnMaxIdleTime,
		Environment:     conf.App.Environment,
	}
	if dbConfig.Timezone == "" {
		dbConfig.Timezone = conf.App.Timezone
	}

	return &migration.Config{
		MigrationsDir: conf.Migration.MigrationsDir,
		DbConfig:      dbConfig,
		IsDebug:       conf.Migration.Debug || verbose,
		Statements:    statements,
		TableList:     tableList,
	}, nil
}
