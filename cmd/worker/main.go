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
	"vibe-ddd-golang/internal/pkg/queue"
	"vibe-ddd-golang/internal/server/worker"

	"go.uber.org/fx"
)

func main() {
	app := fx.New(
		fx.Provide(
			config.NewConfig,
			logger.NewLogger,
			database.NewDatabase,
			queue.NewClient,
			queue.NewServer,
		),
		worker.Module,
		fx.Invoke(runWorker),
		fx.StartTimeout(config.DefaultStartTimeout),
		fx.StopTimeout(config.DefaultStopTimeout),
	)

	ctx := context.Background()
	if err := app.Start(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start worker application: %v\n", err)
		os.Exit(1)
	}

	// Setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Wait for shutdown signal
	<-sigChan
	fmt.Println("\nReceived shutdown signal, stopping worker gracefully...")

	if err := app.Stop(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to stop worker application gracefully: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Worker stopped successfully")
}

func runWorker(lifecycle fx.Lifecycle, workerServer *worker.Server, queueServer *queue.Server) {
	// Register worker handlers
	workerServer.RegisterHandlers()

	// Start the queue api (it manages its own lifecycle)
	queueServer.Start(lifecycle)
}
