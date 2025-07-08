package queue

import (
	"go.uber.org/zap"
)

type AsynqLogger struct {
	logger *zap.Logger
}

func NewAsynqLogger(logger *zap.Logger) *AsynqLogger {
	return &AsynqLogger{logger: logger}
}

func (l *AsynqLogger) Debug(args ...interface{}) {
	l.logger.Sugar().Debug(args...)
}

func (l *AsynqLogger) Info(args ...interface{}) {
	l.logger.Sugar().Info(args...)
}

func (l *AsynqLogger) Warn(args ...interface{}) {
	l.logger.Sugar().Warn(args...)
}

func (l *AsynqLogger) Error(args ...interface{}) {
	l.logger.Sugar().Error(args...)
}

func (l *AsynqLogger) Fatal(args ...interface{}) {
	l.logger.Sugar().Fatal(args...)
}
