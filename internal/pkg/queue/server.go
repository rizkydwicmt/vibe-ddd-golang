package queue

import (
	"context"
	"fmt"

	"vibe-ddd-golang/internal/config"

	"github.com/hibiken/asynq"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Server struct {
	server *asynq.Server
	mux    *asynq.ServeMux
	logger *zap.Logger
	cfg    *config.Config
}

func NewServer(cfg *config.Config, logger *zap.Logger) *Server {
	redisAddr := fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port)

	redisOpt := asynq.RedisClientOpt{
		Addr:     redisAddr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	}

	serverConfig := asynq.Config{
		Concurrency: cfg.Worker.Concurrency,
		Queues: map[string]int{
			"critical": 6,
			"default":  3,
			"low":      1,
		},
		ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
			logger.Error("Task processing failed",
				zap.String("task_type", task.Type()),
				zap.ByteString("payload", task.Payload()),
				zap.Error(err))
		}),
		Logger: NewAsynqLogger(logger),
	}

	server := asynq.NewServer(redisOpt, serverConfig)
	mux := asynq.NewServeMux()

	logger.Info("Queue api initialized",
		zap.String("redis_addr", redisAddr),
		zap.Int("concurrency", cfg.Worker.Concurrency))

	return &Server{
		server: server,
		mux:    mux,
		logger: logger,
		cfg:    cfg,
	}
}

func (s *Server) RegisterHandler(pattern string, handler asynq.Handler) {
	s.mux.Handle(pattern, handler)
}

func (s *Server) Start(lifecycle fx.Lifecycle) {
	lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				s.logger.Info("Starting queue api")
				if err := s.server.Run(s.mux); err != nil {
					s.logger.Fatal("Queue api failed", zap.Error(err))
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			s.logger.Info("Stopping queue api")
			s.server.Shutdown()
			return nil
		},
	})
}
