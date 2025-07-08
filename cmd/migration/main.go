package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"vibe-ddd-golang/internal/config"
	"vibe-ddd-golang/internal/pkg/database"
	"vibe-ddd-golang/internal/pkg/logger"
	"vibe-ddd-golang/internal/server/migration"

	"go.uber.org/fx"
)

func main() {
	var (
		action = flag.String("action", "migrate", "Action to perform: migrate, seed, drop")
	)
	flag.Parse()

	// Setup graceful shutdown for long-running operations
	ctx, cancel := context.WithCancel(context.Background())
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\nReceived shutdown signal, canceling migration...")
		cancel()
	}()

	app := fx.New(
		fx.Provide(
			config.NewConfig,
			logger.NewLogger,
			database.NewDatabase,
		),
		migration.Module,
		fx.Invoke(func(migrationServer *migration.Server) {
			runMigration(ctx, migrationServer, *action)
		}),
	)

	if err := app.Start(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start migration application: %v\n", err)
		os.Exit(1)
	}

	if err := app.Stop(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to stop migration application gracefully: %v\n", err)
		os.Exit(1)
	}
}

func runMigration(ctx context.Context, server *migration.Server, action string) {
	var err error

	switch action {
	case "migrate":
		fmt.Println("Running database migrations...")
		err = server.RunMigrations()
	case "seed":
		fmt.Println("Seeding database...")
		err = server.SeedData()
	case "drop":
		fmt.Println("Dropping database tables...")
		err = server.DropTables()
	default:
		fmt.Fprintf(os.Stderr, "Unknown action: %s. Available actions: migrate, seed, drop\n", action)
		os.Exit(1)
	}

	// Check if context was canceled
	select {
	case <-ctx.Done():
		fmt.Println("Migration canceled by user")
		os.Exit(1)
	default:
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Migration failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Migration action '%s' completed successfully\n", action)
}
