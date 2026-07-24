// Package infrastructure is the composition root for process-wide singletons:
// the gin engine, structured logger, database, redis, and rabbitmq. Constructors
// here are registered with fx in cmd/api/main.go. They read the app's Viper config
// and hand each pkg library its own small Config struct.
package infrastructure

import (
	"io"
	"log"
	"net/http"

	"vibe-ddd-golang/internal/common/enum"
	"vibe-ddd-golang/internal/config"
	"vibe-ddd-golang/internal/pkg/logger"
	"vibe-ddd-golang/internal/pkg/validation"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// InitializeLogger sets up the package-level stdlib/slog loggers.
func InitializeLogger() { logger.Setup() }

// InitializeValidation sets up the shared validator.
func InitializeValidation() error { return validation.Setup() }

// NewGinEngine builds the gin engine, switching mode and recovery behavior by
// environment. In production it discards verbose stdlib logs and installs a quiet
// panic recovery; in dev it keeps the default logger/recovery.
func NewGinEngine(cfg *config.Config) *gin.Engine {
	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
		logger.Warning = log.New(io.Discard, "", 0)
		logger.Debug = log.New(io.Discard, "", 0)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	router := gin.New()
	if cfg.IsProduction() {
		router.Use(gin.CustomRecovery(func(c *gin.Context, recovered any) {
			logger.Error.Printf("panic recovered: %v path=%s", recovered, c.Request.URL.Path)
			c.AbortWithStatus(http.StatusInternalServerError)
		}))
	} else {
		router.Use(gin.Logger())
		router.Use(gin.Recovery())
	}
	router.RedirectTrailingSlash = false
	return router
}

// NewLogger builds the application *zap.Logger. Production emits sampled JSON to
// stdout (GCP/Datadog-friendly); dev emits human-readable console output.
func NewLogger(cfg *config.Config) *zap.Logger {
	var (
		zl  *zap.Logger
		err error
	)
	if cfg.IsProduction() {
		zc := zap.NewProductionConfig()
		zc.OutputPaths = []string{"stdout"}
		zc.ErrorOutputPaths = []string{"stderr"}
		zc.Encoding = "json"

		level := parseLogLevel(cfg.Logger.Level)
		if level == zapcore.InvalidLevel {
			level = zapcore.WarnLevel
		}
		zc.Level = zap.NewAtomicLevelAt(level)
		zc.EncoderConfig.TimeKey = "timestamp"
		zc.EncoderConfig.CallerKey = "caller"
		zc.EncoderConfig.StacktraceKey = "stacktrace"
		zc.EncoderConfig.EncodeTime = zapcore.RFC3339TimeEncoder
		if level <= zapcore.InfoLevel {
			zc.Sampling = &zap.SamplingConfig{Initial: 100, Thereafter: 100}
		}
		zl, err = zc.Build(zap.AddCaller(), zap.AddStacktrace(zap.ErrorLevel))
	} else {
		zl, err = zap.NewDevelopment()
	}
	if err != nil {
		panic(err)
	}
	return zl
}

func parseLogLevel(level enum.LogLevelEnum) zapcore.Level {
	switch level {
	case enum.LogLevelDebug:
		return zapcore.DebugLevel
	case enum.LogLevelInfo:
		return zapcore.InfoLevel
	case enum.LogLevelWarn, enum.LogLevelWarning:
		return zapcore.WarnLevel
	case enum.LogLevelError:
		return zapcore.ErrorLevel
	default:
		return zapcore.InvalidLevel
	}
}
