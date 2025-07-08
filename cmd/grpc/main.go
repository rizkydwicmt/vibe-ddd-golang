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
	"vibe-ddd-golang/internal/server/grpc"

	"go.uber.org/fx"
)

func main() {
	var (
		port = flag.String("port", "9090", "gRPC api port")
	)
	flag.Parse()

	app := fx.New(
		fx.Provide(
			config.NewConfig,
			logger.NewLogger,
			database.NewDatabase,
		),
		grpc.Module,
		fx.Invoke(func(lifecycle fx.Lifecycle, grpcServer *grpc.Server) {
			runGRPCServer(lifecycle, grpcServer, *port)
		}),
		fx.StartTimeout(config.DefaultStartTimeout),
		fx.StopTimeout(config.DefaultStopTimeout),
	)

	ctx := context.Background()
	if err := app.Start(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start gRPC application: %v\n", err)
		os.Exit(1)
	}

	// Setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Wait for shutdown signal
	<-sigChan
	fmt.Println("\nReceived shutdown signal, stopping gRPC api gracefully...")

	if err := app.Stop(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to stop gRPC application gracefully: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("gRPC api stopped successfully")
}

func runGRPCServer(lifecycle fx.Lifecycle, server *grpc.Server, port string) {
	lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				if err := server.Start(port); err != nil {
					fmt.Fprintf(os.Stderr, "Failed to start gRPC api: %v\n", err)
					os.Exit(1)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			server.Stop()
			return nil
		},
	})
}
