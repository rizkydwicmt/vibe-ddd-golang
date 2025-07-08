package handler

import (
	"context"

	"vibe-ddd-golang/api/proto/payment"
	"vibe-ddd-golang/internal/application/payment/dto"
	"vibe-ddd-golang/internal/application/payment/entity"
	"vibe-ddd-golang/internal/application/payment/service"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type PaymentGrpcHandler struct {
	payment.UnimplementedPaymentServiceServer
	paymentService service.PaymentService
	logger         *zap.Logger
}

func NewPaymentGrpcHandler(paymentService service.PaymentService, logger *zap.Logger) *PaymentGrpcHandler {
	return &PaymentGrpcHandler{
		paymentService: paymentService,
		logger:         logger,
	}
}

func (h *PaymentGrpcHandler) CreatePayment(
	ctx context.Context,
	req *payment.CreatePaymentRequest,
) (*payment.CreatePaymentResponse, error) {
	createReq := &dto.CreatePaymentRequest{
		Amount:      req.Amount,
		Currency:    req.Currency,
		Description: req.Description,
		UserID:      uint(req.UserId),
	}

	paymentResponse, err := h.paymentService.CreatePayment(createReq)
	if err != nil {
		h.logger.Error("Failed to create payment via gRPC", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to create payment: %v", err)
	}

	return &payment.CreatePaymentResponse{
		Payment: h.toProtoPayment(paymentResponse),
	}, nil
}

func (h *PaymentGrpcHandler) GetPayment(
	ctx context.Context,
	req *payment.GetPaymentRequest,
) (*payment.GetPaymentResponse, error) {
	paymentResponse, err := h.paymentService.GetPaymentByID(uint(req.Id))
	if err != nil {
		h.logger.Error("Failed to get payment via gRPC", zap.Uint32("id", req.Id), zap.Error(err))
		return nil, status.Errorf(codes.NotFound, "payment not found: %v", err)
	}

	return &payment.GetPaymentResponse{
		Payment: h.toProtoPayment(paymentResponse),
	}, nil
}

func (h *PaymentGrpcHandler) ListPayments(
	ctx context.Context,
	req *payment.ListPaymentsRequest,
) (*payment.ListPaymentsResponse, error) {
	page := int(req.Page)
	pageSize := int(req.PageSize)

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	filter := &dto.PaymentFilter{
		Page:     page,
		PageSize: pageSize,
	}

	// Add status filter if provided
	if req.Status != payment.PaymentStatus_PAYMENT_STATUS_UNSPECIFIED {
		filter.Status = h.protoStatusToString(req.Status)
	}

	// Add user filter if provided
	if req.UserId > 0 {
		filter.UserID = uint(req.UserId)
	}

	listResponse, err := h.paymentService.GetPayments(filter)
	if err != nil {
		h.logger.Error("Failed to list payments via gRPC", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to list payments: %v", err)
	}

	protoPayments := make([]*payment.Payment, len(listResponse.Data))
	for i, p := range listResponse.Data {
		protoPayments[i] = h.toProtoPayment(&p)
	}

	return &payment.ListPaymentsResponse{
		Payments: protoPayments,
		Total:    listResponse.TotalCount,
		Page:     int32(listResponse.Page),
		PageSize: int32(listResponse.PageSize),
	}, nil
}

func (h *PaymentGrpcHandler) UpdatePayment(
	ctx context.Context,
	req *payment.UpdatePaymentRequest,
) (*payment.UpdatePaymentResponse, error) {
	updateReq := &dto.UpdatePaymentRequest{
		Description: req.Description,
	}

	// Add status if provided
	if req.Status != payment.PaymentStatus_PAYMENT_STATUS_UNSPECIFIED {
		updateReq.Status = h.protoStatusToString(req.Status)
	}

	paymentResponse, err := h.paymentService.UpdatePayment(uint(req.Id), updateReq)
	if err != nil {
		h.logger.Error("Failed to update payment via gRPC", zap.Uint32("id", req.Id), zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to update payment: %v", err)
	}

	return &payment.UpdatePaymentResponse{
		Payment: h.toProtoPayment(paymentResponse),
	}, nil
}

func (h *PaymentGrpcHandler) DeletePayment(
	ctx context.Context,
	req *payment.DeletePaymentRequest,
) (*payment.DeletePaymentResponse, error) {
	err := h.paymentService.DeletePayment(uint(req.Id))
	if err != nil {
		h.logger.Error("Failed to delete payment via gRPC", zap.Uint32("id", req.Id), zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to delete payment: %v", err)
	}

	return &payment.DeletePaymentResponse{
		Success: true,
	}, nil
}

func (h *PaymentGrpcHandler) GetUserPayments(
	ctx context.Context,
	req *payment.GetUserPaymentsRequest,
) (*payment.GetUserPaymentsResponse, error) {
	page := int(req.Page)
	pageSize := int(req.PageSize)

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	filter := &dto.PaymentFilter{
		Page:     page,
		PageSize: pageSize,
		UserID:   uint(req.UserId),
	}

	listResponse, err := h.paymentService.GetPayments(filter)
	if err != nil {
		h.logger.Error("Failed to get user payments via gRPC", zap.Uint32("user_id", req.UserId), zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to get user payments: %v", err)
	}

	protoPayments := make([]*payment.Payment, len(listResponse.Data))
	for i, p := range listResponse.Data {
		protoPayments[i] = h.toProtoPayment(&p)
	}

	return &payment.GetUserPaymentsResponse{
		Payments: protoPayments,
		Total:    listResponse.TotalCount,
		Page:     int32(listResponse.Page),
		PageSize: int32(listResponse.PageSize),
	}, nil
}

func (h *PaymentGrpcHandler) toProtoPayment(p *dto.PaymentResponse) *payment.Payment {
	return &payment.Payment{
		Id:          uint32(p.ID),
		Amount:      p.Amount,
		Currency:    p.Currency,
		Description: p.Description,
		Status:      h.stringStatusToProto(p.Status),
		UserId:      uint32(p.UserID),
		CreatedAt:   timestamppb.New(p.CreatedAt),
		UpdatedAt:   timestamppb.New(p.UpdatedAt),
	}
}

func (h *PaymentGrpcHandler) stringStatusToProto(status string) payment.PaymentStatus {
	switch status {
	case entity.PaymentStatusPending.String():
		return payment.PaymentStatus_PAYMENT_STATUS_PENDING
	case entity.PaymentStatusCompleted.String():
		return payment.PaymentStatus_PAYMENT_STATUS_COMPLETED
	case entity.PaymentStatusFailed.String():
		return payment.PaymentStatus_PAYMENT_STATUS_FAILED
	case entity.PaymentStatusCanceled.String():
		return payment.PaymentStatus_PAYMENT_STATUS_CANCELED
	default:
		return payment.PaymentStatus_PAYMENT_STATUS_UNSPECIFIED
	}
}

func (h *PaymentGrpcHandler) protoStatusToString(status payment.PaymentStatus) string {
	switch status {
	case payment.PaymentStatus_PAYMENT_STATUS_PENDING:
		return entity.PaymentStatusPending.String()
	case payment.PaymentStatus_PAYMENT_STATUS_COMPLETED:
		return entity.PaymentStatusCompleted.String()
	case payment.PaymentStatus_PAYMENT_STATUS_FAILED:
		return entity.PaymentStatusFailed.String()
	case payment.PaymentStatus_PAYMENT_STATUS_CANCELED:
		return entity.PaymentStatusCanceled.String()
	default:
		return entity.PaymentStatusPending.String()
	}
}
