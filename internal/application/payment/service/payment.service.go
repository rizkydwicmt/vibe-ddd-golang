package service

import (
	"context"
	"errors"
	"time"

	"vibe-ddd-golang/internal/application/payment/dto"
	"vibe-ddd-golang/internal/application/payment/entity"
	"vibe-ddd-golang/internal/application/payment/repository"
	"vibe-ddd-golang/internal/application/user/service"
	"vibe-ddd-golang/internal/pkg/response"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type PaymentService interface {
	CreatePayment(ctx context.Context, req *dto.CreatePaymentRequest) (*dto.PaymentResponse, error)
	GetPaymentByID(ctx context.Context, id uint) (*dto.PaymentResponse, error)
	GetPayments(ctx context.Context, filter *dto.PaymentFilter) (*dto.PaymentListResponse, error)
	UpdatePayment(ctx context.Context, id uint, req *dto.UpdatePaymentRequest) (*dto.PaymentResponse, error)
	DeletePayment(ctx context.Context, id uint) error
	GetPaymentsByUser(ctx context.Context, userID uint) ([]dto.PaymentResponse, error)
}

type paymentService struct {
	repo        repository.PaymentRepository
	userService service.UserService
	logger      *zap.Logger
}

func NewPaymentService(
	repo repository.PaymentRepository,
	userService service.UserService,
	logger *zap.Logger,
) PaymentService {
	return &paymentService{
		repo:        repo,
		userService: userService,
		logger:      logger,
	}
}

func (s *paymentService) CreatePayment(ctx context.Context, req *dto.CreatePaymentRequest) (*dto.PaymentResponse, error) {
	// Validate that the user exists before creating a payment (cross-domain).
	if _, err := s.userService.GetUserByID(ctx, req.UserID); err != nil {
		return nil, response.NewBadRequestException("user not found")
	}

	payment := &entity.Payment{
		Amount:      req.Amount,
		Currency:    req.Currency,
		Status:      entity.PaymentStatusPending,
		Description: req.Description,
		UserID:      req.UserID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.repo.Create(ctx, payment); err != nil {
		return nil, response.NewInternalServerException("failed to create payment").WithCause(err)
	}

	return s.entityToResponse(payment), nil
}

func (s *paymentService) GetPaymentByID(ctx context.Context, id uint) (*dto.PaymentResponse, error) {
	payment, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, response.NewNotFoundException("payment not found")
		}
		return nil, response.NewInternalServerException("failed to get payment").WithCause(err)
	}

	return s.entityToResponse(payment), nil
}

func (s *paymentService) GetPayments(ctx context.Context, filter *dto.PaymentFilter) (*dto.PaymentListResponse, error) {
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 10
	}

	payments, totalCount, err := s.repo.GetAll(ctx, filter)
	if err != nil {
		return nil, response.NewInternalServerException("failed to list payments").WithCause(err)
	}

	responses := make([]dto.PaymentResponse, 0, len(payments))
	for i := range payments {
		responses = append(responses, *s.entityToResponse(&payments[i]))
	}

	return &dto.PaymentListResponse{
		Data:       responses,
		TotalCount: totalCount,
		Page:       filter.Page,
		PageSize:   filter.PageSize,
	}, nil
}

func (s *paymentService) UpdatePayment(ctx context.Context, id uint, req *dto.UpdatePaymentRequest) (*dto.PaymentResponse, error) {
	payment, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, response.NewNotFoundException("payment not found")
		}
		return nil, response.NewInternalServerException("failed to get payment").WithCause(err)
	}

	status := entity.PaymentStatus(req.Status)
	if !status.IsValid() {
		return nil, response.NewBadRequestException("invalid payment status")
	}

	payment.Status = status
	if req.Description != "" {
		payment.Description = req.Description
	}
	payment.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, payment); err != nil {
		return nil, response.NewInternalServerException("failed to update payment").WithCause(err)
	}

	return s.entityToResponse(payment), nil
}

func (s *paymentService) DeletePayment(ctx context.Context, id uint) error {
	if _, err := s.repo.GetByID(ctx, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return response.NewNotFoundException("payment not found")
		}
		return response.NewInternalServerException("failed to get payment").WithCause(err)
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return response.NewInternalServerException("failed to delete payment").WithCause(err)
	}
	return nil
}

func (s *paymentService) GetPaymentsByUser(ctx context.Context, userID uint) ([]dto.PaymentResponse, error) {
	payments, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, response.NewInternalServerException("failed to get payments").WithCause(err)
	}

	responses := make([]dto.PaymentResponse, 0, len(payments))
	for i := range payments {
		responses = append(responses, *s.entityToResponse(&payments[i]))
	}

	return responses, nil
}

func (s *paymentService) entityToResponse(payment *entity.Payment) *dto.PaymentResponse {
	return &dto.PaymentResponse{
		ID:          payment.ID,
		Amount:      payment.Amount,
		Currency:    payment.Currency,
		Status:      payment.Status.String(),
		Description: payment.Description,
		UserID:      payment.UserID,
		CreatedAt:   payment.CreatedAt,
		UpdatedAt:   payment.UpdatedAt,
	}
}
