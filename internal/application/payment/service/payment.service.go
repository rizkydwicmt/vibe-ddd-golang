package service

import (
	"errors"
	"time"

	"vibe-ddd-golang/internal/application/payment/dto"
	"vibe-ddd-golang/internal/application/payment/entity"
	"vibe-ddd-golang/internal/application/payment/repository"
	"vibe-ddd-golang/internal/application/user/service"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type PaymentService interface {
	CreatePayment(req *dto.CreatePaymentRequest) (*dto.PaymentResponse, error)
	GetPaymentByID(id uint) (*dto.PaymentResponse, error)
	GetPayments(filter *dto.PaymentFilter) (*dto.PaymentListResponse, error)
	UpdatePayment(id uint, req *dto.UpdatePaymentRequest) (*dto.PaymentResponse, error)
	DeletePayment(id uint) error
	GetPaymentsByUser(userID uint) ([]dto.PaymentResponse, error)
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

func (s *paymentService) CreatePayment(req *dto.CreatePaymentRequest) (*dto.PaymentResponse, error) {
	// Validate that user exists before creating payment
	_, err := s.userService.GetUserByID(req.UserID)
	if err != nil {
		s.logger.Error("User not found for payment creation", zap.Uint("user_id", req.UserID), zap.Error(err))
		return nil, errors.New("user not found")
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

	err = s.repo.Create(payment)
	if err != nil {
		s.logger.Error("Failed to create payment", zap.Error(err))
		return nil, err
	}

	return s.entityToResponse(payment), nil
}

func (s *paymentService) GetPaymentByID(id uint) (*dto.PaymentResponse, error) {
	payment, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("payment not found")
		}
		return nil, err
	}

	return s.entityToResponse(payment), nil
}

func (s *paymentService) GetPayments(filter *dto.PaymentFilter) (*dto.PaymentListResponse, error) {
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 10
	}

	payments, totalCount, err := s.repo.GetAll(filter)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.PaymentResponse, 0, len(payments))
	for _, payment := range payments {
		responses = append(responses, *s.entityToResponse(&payment))
	}

	return &dto.PaymentListResponse{
		Data:       responses,
		TotalCount: totalCount,
		Page:       filter.Page,
		PageSize:   filter.PageSize,
	}, nil
}

func (s *paymentService) UpdatePayment(id uint, req *dto.UpdatePaymentRequest) (*dto.PaymentResponse, error) {
	payment, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("payment not found")
		}
		return nil, err
	}

	status := entity.PaymentStatus(req.Status)
	if !status.IsValid() {
		return nil, errors.New("invalid payment status")
	}

	payment.Status = status
	if req.Description != "" {
		payment.Description = req.Description
	}
	payment.UpdatedAt = time.Now()

	err = s.repo.Update(payment)
	if err != nil {
		s.logger.Error("Failed to update payment", zap.Error(err))
		return nil, err
	}

	return s.entityToResponse(payment), nil
}

func (s *paymentService) DeletePayment(id uint) error {
	_, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("payment not found")
		}
		return err
	}

	return s.repo.Delete(id)
}

func (s *paymentService) GetPaymentsByUser(userID uint) ([]dto.PaymentResponse, error) {
	payments, err := s.repo.GetByUserID(userID)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.PaymentResponse, 0, len(payments))
	for _, payment := range payments {
		responses = append(responses, *s.entityToResponse(&payment))
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
