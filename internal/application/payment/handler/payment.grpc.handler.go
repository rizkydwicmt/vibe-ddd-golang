package handler

import (
	"context"

	"vibe-ddd-golang/internal/application/payment/dto"
	"vibe-ddd-golang/internal/application/payment/entity"
	"vibe-ddd-golang/internal/application/payment/service"
	"vibe-ddd-golang/internal/pkg/grpcstatus"
	"vibe-ddd-golang/internal/pkg/response"
	"vibe-ddd-golang/internal/pkg/validation"
	paymentv1 "vibe-ddd-golang/internal/server/grpc/proto/payment"

	"google.golang.org/protobuf/types/known/timestamppb"
)

type PaymentGRPCServer struct {
	paymentv1.UnimplementedPaymentServiceServer
	service service.PaymentService
}

func NewPaymentGRPCServer(service service.PaymentService) *PaymentGRPCServer {
	return &PaymentGRPCServer{service: service}
}

func (s *PaymentGRPCServer) CreatePayment(
	ctx context.Context, req *paymentv1.CreatePaymentRequest,
) (*paymentv1.CreatePaymentResponse, error) {
	in := &dto.CreatePaymentRequest{
		Amount:      req.GetAmount(),
		Currency:    req.GetCurrency(),
		Description: req.GetDescription(),
		UserID:      uint(req.GetUserId()),
	}
	if err := validatePaymentGRPC(in); err != nil {
		return nil, err
	}

	payment, err := s.service.CreatePayment(ctx, in)
	if err != nil {
		return nil, grpcstatus.FromError(err)
	}
	return &paymentv1.CreatePaymentResponse{Payment: paymentToProto(payment)}, nil
}

func (s *PaymentGRPCServer) GetPayment(ctx context.Context, req *paymentv1.GetPaymentRequest) (*paymentv1.GetPaymentResponse, error) {
	payment, err := s.service.GetPaymentByID(ctx, uint(req.GetId()))
	if err != nil {
		return nil, grpcstatus.FromError(err)
	}
	return &paymentv1.GetPaymentResponse{Payment: paymentToProto(payment)}, nil
}

func (s *PaymentGRPCServer) ListPayments(ctx context.Context, req *paymentv1.ListPaymentsRequest) (*paymentv1.ListPaymentsResponse, error) {
	payments, err := s.service.GetPayments(ctx, &dto.PaymentFilter{
		Status:   statusFromProto(req.GetStatus()),
		Currency: req.GetCurrency(),
		UserID:   uint(req.GetUserId()),
		Page:     int(req.GetPage()),
		PageSize: int(req.GetPageSize()),
	})
	if err != nil {
		return nil, grpcstatus.FromError(err)
	}

	out := make([]*paymentv1.Payment, 0, len(payments.Data))
	for i := range payments.Data {
		out = append(out, paymentToProto(&payments.Data[i]))
	}

	return &paymentv1.ListPaymentsResponse{
		Payments: out,
		Total:    payments.TotalCount,
		Page:     int32(payments.Page),
		PageSize: int32(payments.PageSize),
	}, nil
}

func (s *PaymentGRPCServer) UpdatePayment(
	ctx context.Context, req *paymentv1.UpdatePaymentRequest,
) (*paymentv1.UpdatePaymentResponse, error) {
	in := &dto.UpdatePaymentRequest{
		Status:      statusFromProto(req.GetStatus()),
		Description: req.GetDescription(),
	}
	if err := validatePaymentGRPC(in); err != nil {
		return nil, err
	}

	payment, err := s.service.UpdatePayment(ctx, uint(req.GetId()), in)
	if err != nil {
		return nil, grpcstatus.FromError(err)
	}
	return &paymentv1.UpdatePaymentResponse{Payment: paymentToProto(payment)}, nil
}

func (s *PaymentGRPCServer) DeletePayment(
	ctx context.Context, req *paymentv1.DeletePaymentRequest,
) (*paymentv1.DeletePaymentResponse, error) {
	if err := s.service.DeletePayment(ctx, uint(req.GetId())); err != nil {
		return nil, grpcstatus.FromError(err)
	}
	return &paymentv1.DeletePaymentResponse{Success: true}, nil
}

func (s *PaymentGRPCServer) GetUserPayments(
	ctx context.Context, req *paymentv1.GetUserPaymentsRequest,
) (*paymentv1.GetUserPaymentsResponse, error) {
	payments, err := s.service.GetPaymentsByUser(ctx, uint(req.GetUserId()))
	if err != nil {
		return nil, grpcstatus.FromError(err)
	}

	out := make([]*paymentv1.Payment, 0, len(payments))
	for i := range payments {
		out = append(out, paymentToProto(&payments[i]))
	}
	return &paymentv1.GetUserPaymentsResponse{Payments: out}, nil
}

func validatePaymentGRPC(in any) error {
	if err := validation.Validate(in); err != nil {
		return grpcstatus.FromError(response.NewBadRequestException(err.Error()))
	}
	return nil
}

func paymentToProto(payment *dto.PaymentResponse) *paymentv1.Payment {
	if payment == nil {
		return nil
	}
	return &paymentv1.Payment{
		Id:          uint32(payment.ID),
		Amount:      payment.Amount,
		Currency:    payment.Currency,
		Description: payment.Description,
		Status:      statusToProto(payment.Status),
		UserId:      uint32(payment.UserID),
		CreatedAt:   timestamppb.New(payment.CreatedAt),
		UpdatedAt:   timestamppb.New(payment.UpdatedAt),
	}
}

func statusFromProto(status paymentv1.PaymentStatus) string {
	switch status {
	case paymentv1.PaymentStatus_PAYMENT_STATUS_PENDING:
		return entity.PaymentStatusPending.String()
	case paymentv1.PaymentStatus_PAYMENT_STATUS_COMPLETED:
		return entity.PaymentStatusCompleted.String()
	case paymentv1.PaymentStatus_PAYMENT_STATUS_FAILED:
		return entity.PaymentStatusFailed.String()
	case paymentv1.PaymentStatus_PAYMENT_STATUS_CANCELED:
		return entity.PaymentStatusCanceled.String()
	default:
		return ""
	}
}

func statusToProto(status string) paymentv1.PaymentStatus {
	switch entity.PaymentStatus(status) {
	case entity.PaymentStatusPending:
		return paymentv1.PaymentStatus_PAYMENT_STATUS_PENDING
	case entity.PaymentStatusCompleted:
		return paymentv1.PaymentStatus_PAYMENT_STATUS_COMPLETED
	case entity.PaymentStatusFailed:
		return paymentv1.PaymentStatus_PAYMENT_STATUS_FAILED
	case entity.PaymentStatusCanceled:
		return paymentv1.PaymentStatus_PAYMENT_STATUS_CANCELED
	default:
		return paymentv1.PaymentStatus_PAYMENT_STATUS_UNSPECIFIED
	}
}
