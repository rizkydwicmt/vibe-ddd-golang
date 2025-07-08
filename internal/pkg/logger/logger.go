package logger

import (
	"vibe-ddd-golang/internal/config"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func NewLogger(cfg *config.Config) (*zap.Logger, error) {
	var zapConfig zap.Config

	if cfg.Logger.Format == "json" {
		zapConfig = zap.NewProductionConfig()
	} else {
		zapConfig = zap.NewDevelopmentConfig()
	}

	level, err := zapcore.ParseLevel(cfg.Logger.Level)
	if err != nil {
		return nil, err
	}
	zapConfig.Level = zap.NewAtomicLevelAt(level)

	if cfg.Logger.OutputPath != "" && cfg.Logger.OutputPath != "stdout" {
		zapConfig.OutputPaths = []string{cfg.Logger.OutputPath}
	}

	logger, err := zapConfig.Build()
	if err != nil {
		return nil, err
	}

	return logger, nil
}
