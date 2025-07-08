package grpc

import (
	"vibe-ddd-golang/internal/application/payment"
	paymentHandler "vibe-ddd-golang/internal/application/payment/handler"
	"vibe-ddd-golang/internal/application/user"
	userHandler "vibe-ddd-golang/internal/application/user/handler"

	"go.uber.org/fx"
)

var Module = fx.Options(
	// Include domain modules
	user.Module,
	payment.Module,

	// gRPC handlers
	fx.Provide(
		userHandler.NewUserGrpcHandler,
		paymentHandler.NewPaymentGrpcHandler,
		NewServer,
	),
)
