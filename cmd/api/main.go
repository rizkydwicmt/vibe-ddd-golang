package main

import (
	"context"
	"os/signal"
	"syscall"

	"vibe-ddd-golang/infrastructure"
	"vibe-ddd-golang/internal/config"
	serverapi "vibe-ddd-golang/internal/server/api"
	grpcserver "vibe-ddd-golang/internal/server/grpc"

	"go.uber.org/fx"
)

// @title           Vibe DDD Golang API
// @version         1.0
// @description     A production-ready Go boilerplate following Domain-Driven Design (DDD) principles.
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  MIT
// @license.url   https://opensource.org/licenses/MIT

// @host      localhost:8080
// @BasePath  /api/v1

func main() {
	rootCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	app := fx.New(
		// app_context is a signal-aware context injected into params.Params.
		fx.Provide(fx.Annotate(
			func() context.Context { return rootCtx },
			fx.ResultTags(`name:"app_context"`),
		)),
		fx.Provide(
			config.NewConfig,
			infrastructure.NewGinEngine,
			infrastructure.NewLogger,
			infrastructure.NewDatabases,
			infrastructure.NewRedis,
			infrastructure.RedisInterface,
			infrastructure.NewRabbitMQ,
		),
		// Process init runs before the graph is built.
		fx.Invoke(infrastructure.InitializeLogger),
		fx.Invoke(infrastructure.InitializeValidation),
		serverapi.Module,
		grpcserver.Module,
		fx.Invoke(Run),
		fx.StartTimeout(config.DefaultStartTimeout),
		fx.StopTimeout(config.DefaultStopTimeout),
	)

	app.Run()
}
