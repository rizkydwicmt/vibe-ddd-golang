package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"vibe-ddd-golang/internal/config"
	"vibe-ddd-golang/internal/server/api"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Server struct {
	router    *gin.Engine
	server    *http.Server
	logger    *zap.Logger
	config    *config.Config
	apiServer *api.Server
}

func NewServer(cfg *config.Config, logger *zap.Logger, apiServer *api.Server) *Server {
	if cfg.Logger.Format == "json" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	return &Server{
		router:    router,
		server:    server,
		logger:    logger,
		config:    cfg,
		apiServer: apiServer,
	}
}

func (s *Server) setupRoutes() {
	s.apiServer.SetupRoutes(s.router)
}

func Run(lifecycle fx.Lifecycle, cfg *config.Config, logger *zap.Logger, apiServer *api.Server) {
	server := NewServer(cfg, logger, apiServer)
	server.setupRoutes()

	lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				logger.Info("Starting HTTP API api",
					zap.String("addr", server.server.Addr))

				if err := server.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					logger.Fatal("Failed to start API api", zap.Error(err))
				}
			}()

			go func() {
				quit := make(chan os.Signal, 1)
				signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
				<-quit
				logger.Info("Shutting down API api...")

				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()

				if err := server.server.Shutdown(ctx); err != nil {
					logger.Fatal("Server forced to shutdown", zap.Error(err))
				}
			}()

			return nil
		},
		OnStop: func(ctx context.Context) error {
			logger.Info("Stopping HTTP API api")
			return server.server.Shutdown(ctx)
		},
	})
}
