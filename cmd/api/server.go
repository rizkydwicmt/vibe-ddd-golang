package main

import (
	"context"
	"fmt"
	"net/http"

	"vibe-ddd-golang/internal/config"
	serverapi "vibe-ddd-golang/internal/server/api"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// Run wires the API server onto the injected gin engine and manages its lifecycle
// via fx hooks (graceful start/stop). Signal handling lives in main via the
// signal-aware root context and fx.App.Run.
func Run(
	lc fx.Lifecycle,
	cfg *config.Config,
	engine *gin.Engine,
	apiServer *serverapi.Server,
	logger *zap.Logger,
) {
	apiServer.SetupRoutes(engine)

	srv := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:      engine,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			go func() {
				logger.Info("starting HTTP API", zap.String("addr", srv.Addr))
				if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					logger.Fatal("failed to start API server", zap.Error(err))
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			logger.Info("stopping HTTP API")
			return srv.Shutdown(ctx)
		},
	})
}
