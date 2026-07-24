package params

import (
	"context"

	database "vibe-ddd-golang/internal/pkg/db"
	"vibe-ddd-golang/internal/pkg/rabbitmq"
	"vibe-ddd-golang/internal/pkg/redis"

	"go.uber.org/fx"
)

type Params struct {
	fx.In

	Ctx       context.Context    `name:"app_context"`
	MainDB    *database.Database `name:"main_db"`
	Redis     *redis.Client
	RabbitMQ  *rabbitmq.ConnectionManager
	Publisher *rabbitmq.Publisher
}
