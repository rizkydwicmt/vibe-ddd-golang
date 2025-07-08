package queue

import (
	"fmt"

	"vibe-ddd-golang/internal/config"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"
)

type Client struct {
	client *asynq.Client
	logger *zap.Logger
}

func NewClient(cfg *config.Config, logger *zap.Logger) *Client {
	redisAddr := fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port)

	redisOpt := asynq.RedisClientOpt{
		Addr:     redisAddr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	}

	client := asynq.NewClient(redisOpt)

	logger.Info("Queue client initialized",
		zap.String("redis_addr", redisAddr),
		zap.Int("redis_db", cfg.Redis.DB))

	return &Client{
		client: client,
		logger: logger,
	}
}

func (c *Client) Close() error {
	return c.client.Close()
}

func (c *Client) GetClient() *asynq.Client {
	return c.client
}

// Enqueue implements the AsynqClient interface
func (c *Client) Enqueue(task *asynq.Task, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	return c.client.Enqueue(task, opts...)
}
