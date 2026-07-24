package testutil

import (
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

// NewTestLogger creates a test logger that outputs to the test
func NewTestLogger(t *testing.T) *zap.Logger {
	return zaptest.NewLogger(t, zaptest.Level(zap.DebugLevel))
}

// NewSilentLogger creates a logger that discards all output
func NewSilentLogger() *zap.Logger {
	return zap.NewNop()
}
