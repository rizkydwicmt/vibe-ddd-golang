package worker

import (
	paymentWorker "vibe-ddd-golang/internal/application/payment/worker"
	"vibe-ddd-golang/internal/pkg/queue"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"
)

type Server struct {
	paymentWorker *paymentWorker.PaymentWorker
	queueServer   *queue.Server
	logger        *zap.Logger
}

func NewServer(
	paymentWorker *paymentWorker.PaymentWorker,
	queueServer *queue.Server,
	logger *zap.Logger,
) *Server {
	return &Server{
		paymentWorker: paymentWorker,
		queueServer:   queueServer,
		logger:        logger,
	}
}

func (s *Server) RegisterHandlers() {
	s.logger.Info("Registering worker handlers")

	// Register payment workers
	s.queueServer.RegisterHandler(
		paymentWorker.TypeCheckPaymentStatus,
		asynq.HandlerFunc(s.paymentWorker.HandleCheckPaymentStatus),
	)

	s.queueServer.RegisterHandler(
		paymentWorker.TypeProcessPayment,
		asynq.HandlerFunc(s.paymentWorker.HandleProcessPayment),
	)

	s.logger.Info("Worker handlers registered successfully")
}
