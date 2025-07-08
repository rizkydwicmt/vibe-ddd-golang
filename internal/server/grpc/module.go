package grpc

import (
	"context"
	"net"

	"vibe-ddd-golang/api/proto/payment"
	"vibe-ddd-golang/api/proto/user"
	paymentHandler "vibe-ddd-golang/internal/application/payment/handler"
	userHandler "vibe-ddd-golang/internal/application/user/handler"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type Server struct {
	server         *grpc.Server
	logger         *zap.Logger
	userHandler    *userHandler.UserGrpcHandler
	paymentHandler *paymentHandler.PaymentGrpcHandler
}

func NewServer(
	logger *zap.Logger,
	userHandler *userHandler.UserGrpcHandler,
	paymentHandler *paymentHandler.PaymentGrpcHandler,
) *Server {
	// Create gRPC api with options
	server := grpc.NewServer(
		grpc.UnaryInterceptor(unaryLoggingInterceptor(logger)),
	)

	return &Server{
		server:         server,
		logger:         logger,
		userHandler:    userHandler,
		paymentHandler: paymentHandler,
	}
}

func (s *Server) RegisterServices() {
	s.logger.Info("Registering gRPC services")

	// Register user service
	user.RegisterUserServiceServer(s.server, s.userHandler)
	s.logger.Info("User service registered")

	// Register payment service
	payment.RegisterPaymentServiceServer(s.server, s.paymentHandler)
	s.logger.Info("Payment service registered")

	s.logger.Info("gRPC services registered successfully")
}

func (s *Server) Start(port string) error {
	s.logger.Info("Starting gRPC api", zap.String("port", port))

	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		s.logger.Error("Failed to listen on port", zap.String("port", port), zap.Error(err))
		return err
	}

	s.RegisterServices()

	s.logger.Info("gRPC api listening", zap.String("address", listener.Addr().String()))
	return s.server.Serve(listener)
}

func (s *Server) Stop() {
	s.logger.Info("Stopping gRPC api")
	s.server.GracefulStop()
}

// unaryLoggingInterceptor logs gRPC calls
func unaryLoggingInterceptor(logger *zap.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		logger.Info("gRPC call", zap.String("method", info.FullMethod))
		return handler(ctx, req)
	}
}
