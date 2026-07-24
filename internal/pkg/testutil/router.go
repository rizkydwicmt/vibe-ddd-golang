package testutil

import (
	"sync"

	"vibe-ddd-golang/internal/pkg/logger"
	"vibe-ddd-golang/internal/pkg/middleware"
	"vibe-ddd-golang/internal/pkg/validation"

	"github.com/gin-gonic/gin"
)

var setupOnce sync.Once

// NewTestRouter returns a gin engine wired with the request/response envelope
// middleware so handlers under test can call response.Send/Render. It initializes
// the package logger and validator once per process.
func NewTestRouter() *gin.Engine {
	setupOnce.Do(func() {
		gin.SetMode(gin.TestMode)
		logger.Setup()
		_ = validation.Setup()
	})
	r := gin.New()
	r.Use(middleware.RequestInit(), middleware.ResponseInit(), middleware.Recovery())
	return r
}
