package grpcserver

import (
	"context"
	"errors"
	"fmt"
	"net"

	paymenthandler "vibe-ddd-golang/internal/application/payment/handler"
	userhandler "vibe-ddd-golang/internal/application/user/handler"
	"vibe-ddd-golang/internal/config"
	paymentv1 "vibe-ddd-golang/internal/server/grpc/proto/payment"
	userv1 "vibe-ddd-golang/internal/server/grpc/proto/user"

	"go.uber.org/fx"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func NewServer(
	userGRPC *userhandler.UserGRPCServer,
	paymentGRPC *paymenthandler.PaymentGRPCServer,
) *grpc.Server {
	srv := grpc.NewServer()
	userv1.RegisterUserServiceServer(srv, userGRPC)
	paymentv1.RegisterPaymentServiceServer(srv, paymentGRPC)
	return srv
}

type lifecycleDeps struct {
	fx.In

	Server *grpc.Server
	Config *config.Config
	Logger *zap.Logger
}

func RegisterLifecycle(lc fx.Lifecycle, deps lifecycleDeps) {
	addr := fmt.Sprintf(":%d", deps.Config.Server.GRPCPort)

	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			listener, err := net.Listen("tcp", addr)
			if err != nil {
				return err
			}

			go func() {
				deps.Logger.Info("starting gRPC API", zap.String("addr", addr))
				if err := deps.Server.Serve(listener); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
					deps.Logger.Error("gRPC API stopped with error", zap.Error(err))
				}
			}()

			return nil
		},
		OnStop: func(ctx context.Context) error {
			deps.Logger.Info("stopping gRPC API")
			stopped := make(chan struct{})
			go func() {
				deps.Server.GracefulStop()
				close(stopped)
			}()

			select {
			case <-stopped:
				return nil
			case <-ctx.Done():
				deps.Server.Stop()
				return ctx.Err()
			}
		},
	})
}
