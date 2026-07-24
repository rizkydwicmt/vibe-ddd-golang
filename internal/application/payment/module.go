package payment

import (
	"vibe-ddd-golang/internal/application/payment/handler"
	"vibe-ddd-golang/internal/application/payment/repository"
	"vibe-ddd-golang/internal/application/payment/service"

	"go.uber.org/fx"
)

// Module provides all payment domain dependencies
var Module = fx.Options(
	fx.Provide(
		repository.NewPaymentRepository,
		service.NewPaymentService,
		handler.NewPaymentHandler,
		handler.NewPaymentGRPCServer,
	),
)
