package infrastructure

import (
	"context"
	"fmt"
	"time"

	"vibe-ddd-golang/internal/config"
	"vibe-ddd-golang/internal/pkg/logger"
	"vibe-ddd-golang/internal/pkg/rabbitmq"

	"go.uber.org/fx"
)

// RabbitMQResult holds the RabbitMQ connection manager and publisher. Both are nil
// when RabbitMQ is not configured (see NewRabbitMQ) — consumers must nil-check.
type RabbitMQResult struct {
	fx.Out

	ConnectionManager *rabbitmq.ConnectionManager
	Publisher         *rabbitmq.Publisher
}

// NewRabbitMQ creates the RabbitMQ connection and publisher. RabbitMQ is optional in
// this boilerplate: when RABBIT uri is unset the API boots without eventing and an
// empty result (nil manager/publisher) is provided.
func NewRabbitMQ(lc fx.Lifecycle, cfg *config.Config) (RabbitMQResult, error) {
	var result RabbitMQResult
	ctx := context.Background()

	if cfg.Rabbit.URI == "" {
		logger.Warning.Println("rabbitmq uri not configured; skipping rabbitmq (running without eventing)")
		return result, nil
	}

	connManager, err := rabbitmq.NewConnectionManager(ctx, &rabbitmq.Config{
		URI: cfg.Rabbit.URI,
	})
	if err != nil {
		return result, fmt.Errorf("CRITICAL: RabbitMQ connection required but failed: %w", err)
	}
	result.ConnectionManager = connManager

	publisher, err := rabbitmq.NewPublisher(ctx, connManager)
	if err != nil {
		_ = connManager.Close()
		return result, fmt.Errorf("CRITICAL: RabbitMQ publisher creation required but failed: %w", err)
	}
	result.Publisher = publisher

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			checkCtx, cancel := context.WithTimeout(ctx, infrastructureStartupCheckTimeout)
			defer cancel()
			if err := verifyRabbitMQInfrastructure(checkCtx, connManager); err != nil {
				return fmt.Errorf("rabbitmq readiness check failed: %w", err)
			}
			logger.Info.Println("rabbitmq readiness check passed")
			return nil
		},
		OnStop: func(ctx context.Context) error {
			timeoutCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
			defer cancel()
			done := make(chan error, 1)
			go func() { done <- connManager.Close() }()
			select {
			case <-done:
			case <-timeoutCtx.Done():
			}
			return nil
		},
	})

	return result, nil
}

func verifyRabbitMQInfrastructure(ctx context.Context, conn *rabbitmq.ConnectionManager) error {
	if conn == nil {
		return fmt.Errorf("rabbitmq dependency is nil")
	}
	if err := conn.WaitForConnection(boundedWaitTimeout(ctx, 5*time.Second)); err != nil {
		return fmt.Errorf("rabbitmq connection not ready: %w", err)
	}
	if !conn.IsHealthy() {
		return fmt.Errorf("rabbitmq connection is unhealthy")
	}
	select {
	case <-ctx.Done():
		return fmt.Errorf("rabbitmq readiness check canceled: %w", ctx.Err())
	default:
		return nil
	}
}
