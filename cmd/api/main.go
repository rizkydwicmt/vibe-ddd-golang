package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"vibe-ddd-golang/internal/config"
	"vibe-ddd-golang/internal/pkg/database"
	"vibe-ddd-golang/internal/pkg/logger"
	"vibe-ddd-golang/internal/server/api"

	"go.uber.org/fx"
)

// @title           Vibe DDD Golang API
// @version         1.0
// @description     A production-ready Go boilerplate following Domain-Driven Design (DDD) principles with NestJS-like architecture patterns.
// @description     Built with modern Go practices, microservice architecture, and comprehensive background job processing.
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  MIT
// @license.url   https://opensource.org/licenses/MIT

// @host      localhost:8080
// @BasePath  /api/v1

// @securityDefinitions.basic  BasicAuth

// @externalDocs.description  OpenAPI
// @externalDocs.url          https://swagger.io/resources/open-api/

func main() {
	app := fx.New(
		fx.Provide(
			config.NewConfig,
			logger.NewLogger,
			database.NewDatabase,
		),
		api.Module,
		fx.Invoke(Run),
		fx.StartTimeout(config.DefaultStartTimeout),
		fx.StopTimeout(config.DefaultStopTimeout),
	)

	ctx := context.Background()
	if err := app.Start(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start application: %v\n", err)
		os.Exit(1)
	}

	// Setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Wait for shutdown signal
	<-sigChan
	fmt.Println("\nReceived shutdown signal, stopping application gracefully...")

	if err := app.Stop(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to stop application gracefully: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Application stopped successfully")
}
