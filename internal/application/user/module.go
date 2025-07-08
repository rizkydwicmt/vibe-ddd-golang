package user

import (
	"vibe-ddd-golang/internal/application/user/handler"
	"vibe-ddd-golang/internal/application/user/repository"
	"vibe-ddd-golang/internal/application/user/service"

	"go.uber.org/fx"
)

// Module provides all user domain dependencies
var Module = fx.Options(
	fx.Provide(
		repository.NewUserRepository,
		service.NewUserService,
		handler.NewUserHandler,
	),
)

// WorkerModule provides only worker dependencies for worker api
var WorkerModule = fx.Options(
	fx.Provide(
		repository.NewUserRepository,
		service.NewUserService,
	),
)
