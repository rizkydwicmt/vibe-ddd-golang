package payment

import (
	"vibe-ddd-golang/internal/application/payment/handler"
	"vibe-ddd-golang/internal/application/payment/repository"
	"vibe-ddd-golang/internal/application/payment/service"
	"vibe-ddd-golang/internal/application/payment/worker"
	"vibe-ddd-golang/internal/pkg/queue"

	"go.uber.org/fx"
)

// Module provides all payment domain dependencies
var Module = fx.Options(
	fx.Provide(
		repository.NewPaymentRepository,
		service.NewPaymentService,
		handler.NewPaymentHandler,
		worker.NewPaymentWorker,
	),
)

// WorkerModule provides only worker dependencies for worker api
var WorkerModule = fx.Options(
	fx.Provide(
		repository.NewPaymentRepository,
		service.NewPaymentService,
		// Provide the queue client as AsynqClient interface
		func(client *queue.Client) worker.AsynqClient {
			return client
		},
		worker.NewPaymentWorker,
	),
)
