package infrastructure

import (
	"context"
	"fmt"
	"time"

	"vibe-ddd-golang/internal/config"
	"vibe-ddd-golang/internal/pkg/logger"
	"vibe-ddd-golang/internal/pkg/redis"

	"go.uber.org/fx"
)

// NewRedis dials Redis and registers a lifecycle hook that pings on start and closes
// on stop. Redis is optional in this boilerplate: when REDIS host is unset the API
// runs on Postgres alone and a nil client is provided (consumers must nil-check).
func NewRedis(lc fx.Lifecycle, cfg *config.Config) (*redis.Client, error) {
	if cfg.Redis.Host == "" {
		logger.Warning.Println("redis host not configured; skipping redis (running without cache/session)")
		return nil, nil //nolint:nilnil // optional dependency: nil client means redis disabled
	}

	rds, err := redis.Setup(context.Background(), &redis.Config{
		Host:     cfg.Redis.Host,
		Port:     cfg.Redis.Port,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
		Prefix:   cfg.Redis.KeyPrefix,
		PoolSize: cfg.Redis.PoolSize,
	})
	if err != nil {
		return nil, fmt.Errorf("CRITICAL: redis connection required but failed: %w", err)
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			checkCtx, cancel := context.WithTimeout(ctx, boundedWaitTimeout(ctx, infrastructureStartupCheckTimeout))
			defer cancel()
			if err := rds.Ping(checkCtx); err != nil {
				return fmt.Errorf("redis readiness check failed: %w", err)
			}
			logger.Info.Println("redis readiness check passed")
			return nil
		},
		OnStop: func(ctx context.Context) error {
			done := make(chan error, 1)
			go func() { done <- rds.Close() }()
			select {
			case <-done:
			case <-time.After(10 * time.Second):
			}
			return nil
		},
	})

	return rds, nil
}

// RedisInterface exposes the concrete Redis client through the interface used by
// consumers. Returns a nil interface (not an interface wrapping a nil pointer) when
// Redis is absent, so `iface == nil` guards behave.
func RedisInterface(rds *redis.Client) redis.IRedis {
	if rds == nil {
		return nil
	}
	return rds
}
