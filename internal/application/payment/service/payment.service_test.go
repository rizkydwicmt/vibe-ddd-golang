package service

import (
	"errors"
	"testing"
	"time"

	"vibe-ddd-golang/internal/application/payment/dto"
	"vibe-ddd-golang/internal/application/payment/entity"
	userDto "vibe-ddd-golang/internal/application/user/dto"
	"vibe-ddd-golang/internal/pkg/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

func TestPaymentService_CreatePayment(t *testing.T) {
	t.Run("should create payment successfully", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockPaymentRepository{}
		mockUserService := &testutil.MockUserService{}
		logger := testutil.NewSilentLogger()
		service := NewPaymentService(mockRepo, mockUserService, logger)

		req := testutil.CreatePaymentRequestFixture()
		userResponse := &userDto.UserResponse{
			ID:    req.UserID,
			Name:  "John Doe",
			Email: "john@example.com",
		}

		// Mock expectations
		mockUserService.On("GetUserByID", req.UserID).Return(userResponse, nil)
		mockRepo.On("Create", mock.AnythingOfType("*entity.Payment")).Return(nil).Run(func(args mock.Arguments) {
			payment := args.Get(0).(*entity.Payment)
			payment.ID = 1
		})

		// When
		response, err := service.CreatePayment(req)

		// Then
		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, uint(1), response.ID)
		assert.Equal(t, req.Amount, response.Amount)
		assert.Equal(t, req.Currency, response.Currency)
		assert.Equal(t, req.Description, response.Description)
		assert.Equal(t, req.UserID, response.UserID)
		assert.Equal(t, entity.PaymentStatusPending.String(), response.Status)
		mockRepo.AssertExpectations(t)
		mockUserService.AssertExpectations(t)
	})

	t.Run("should return error when user not found", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockPaymentRepository{}
		mockUserService := &testutil.MockUserService{}
		logger := testutil.NewSilentLogger()
		service := NewPaymentService(mockRepo, mockUserService, logger)

		req := testutil.CreatePaymentRequestFixture()

		// Mock expectations
		mockUserService.On("GetUserByID", req.UserID).Return(nil, errors.New("user not found"))

		// When
		response, err := service.CreatePayment(req)

		// Then
		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "user not found")
		mockUserService.AssertExpectations(t)
		mockRepo.AssertNotCalled(t, "Create")
	})

	t.Run("should return error when payment creation fails", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockPaymentRepository{}
		mockUserService := &testutil.MockUserService{}
		logger := testutil.NewSilentLogger()
		service := NewPaymentService(mockRepo, mockUserService, logger)

		req := testutil.CreatePaymentRequestFixture()
		userResponse := &userDto.UserResponse{
			ID:    req.UserID,
			Name:  "John Doe",
			Email: "john@example.com",
		}

		// Mock expectations
		mockUserService.On("GetUserByID", req.UserID).Return(userResponse, nil)
		mockRepo.On("Create", mock.AnythingOfType("*entity.Payment")).Return(errors.New("create failed"))

		// When
		response, err := service.CreatePayment(req)

		// Then
		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "create failed")
		mockRepo.AssertExpectations(t)
		mockUserService.AssertExpectations(t)
	})
}

func TestPaymentService_GetPaymentByID(t *testing.T) {
	t.Run("should get payment by ID successfully", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockPaymentRepository{}
		mockUserService := &testutil.MockUserService{}
		logger := testutil.NewSilentLogger()
		service := NewPaymentService(mockRepo, mockUserService, logger)

		paymentID := uint(1)
		payment := testutil.CreatePaymentFixture()
		payment.ID = paymentID

		// Mock expectations
		mockRepo.On("GetByID", paymentID).Return(payment, nil)

		// When
		response, err := service.GetPaymentByID(paymentID)

		// Then
		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, paymentID, response.ID)
		assert.Equal(t, payment.Amount, response.Amount)
		assert.Equal(t, payment.Currency, response.Currency)
		assert.Equal(t, payment.Status.String(), response.Status)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error when payment not found", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockPaymentRepository{}
		mockUserService := &testutil.MockUserService{}
		logger := testutil.NewSilentLogger()
		service := NewPaymentService(mockRepo, mockUserService, logger)

		paymentID := uint(999)

		// Mock expectations
		mockRepo.On("GetByID", paymentID).Return(nil, gorm.ErrRecordNotFound)

		// When
		response, err := service.GetPaymentByID(paymentID)

		// Then
		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "payment not found")
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error when repository fails", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockPaymentRepository{}
		mockUserService := &testutil.MockUserService{}
		logger := testutil.NewSilentLogger()
		service := NewPaymentService(mockRepo, mockUserService, logger)

		paymentID := uint(1)

		// Mock expectations
		mockRepo.On("GetByID", paymentID).Return(nil, errors.New("database error"))

		// When
		response, err := service.GetPaymentByID(paymentID)

		// Then
		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "database error")
		mockRepo.AssertExpectations(t)
	})
}

func TestPaymentService_GetPayments(t *testing.T) {
	t.Run("should get payments with pagination successfully", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockPaymentRepository{}
		mockUserService := &testutil.MockUserService{}
		logger := testutil.NewSilentLogger()
		service := NewPaymentService(mockRepo, mockUserService, logger)

		filter := &dto.PaymentFilter{
			Page:     1,
			PageSize: 10,
		}

		payments := []entity.Payment{
			*testutil.CreatePaymentFixture(),
			*testutil.CreatePaymentFixture(),
		}
		payments[0].ID = 1
		payments[1].ID = 2
		payments[1].Amount = 200.00

		// Mock expectations
		mockRepo.On("GetAll", filter).Return(payments, int64(2), nil)

		// When
		response, err := service.GetPayments(filter)

		// Then
		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Len(t, response.Data, 2)
		assert.Equal(t, int64(2), response.TotalCount)
		assert.Equal(t, 1, response.Page)
		assert.Equal(t, 10, response.PageSize)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should set default pagination values", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockPaymentRepository{}
		mockUserService := &testutil.MockUserService{}
		logger := testutil.NewSilentLogger()
		service := NewPaymentService(mockRepo, mockUserService, logger)

		filter := &dto.PaymentFilter{
			Page:     0,
			PageSize: 0,
		}

		expectedFilter := &dto.PaymentFilter{
			Page:     1,
			PageSize: 10,
		}

		// Mock expectations
		mockRepo.On("GetAll", expectedFilter).Return([]entity.Payment{}, int64(0), nil)

		// When
		response, err := service.GetPayments(filter)

		// Then
		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, 1, response.Page)
		assert.Equal(t, 10, response.PageSize)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error when repository fails", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockPaymentRepository{}
		mockUserService := &testutil.MockUserService{}
		logger := testutil.NewSilentLogger()
		service := NewPaymentService(mockRepo, mockUserService, logger)

		filter := &dto.PaymentFilter{
			Page:     1,
			PageSize: 10,
		}

		// Mock expectations
		mockRepo.On("GetAll", filter).Return(nil, int64(0), errors.New("database error"))

		// When
		response, err := service.GetPayments(filter)

		// Then
		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "database error")
		mockRepo.AssertExpectations(t)
	})
}

func TestPaymentService_UpdatePayment(t *testing.T) {
	t.Run("should update payment successfully", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockPaymentRepository{}
		mockUserService := &testutil.MockUserService{}
		logger := testutil.NewSilentLogger()
		service := NewPaymentService(mockRepo, mockUserService, logger)

		paymentID := uint(1)
		existingPayment := testutil.CreatePaymentFixture()
		existingPayment.ID = paymentID
		existingPayment.Status = entity.PaymentStatusPending

		req := testutil.CreateUpdatePaymentRequestFixture()
		req.Status = entity.PaymentStatusCompleted.String()
		req.Description = "Updated description"

		// Mock expectations
		mockRepo.On("GetByID", paymentID).Return(existingPayment, nil)
		mockRepo.On("Update", mock.AnythingOfType("*entity.Payment")).Return(nil)

		// When
		response, err := service.UpdatePayment(paymentID, req)

		// Then
		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, paymentID, response.ID)
		assert.Equal(t, entity.PaymentStatusCompleted.String(), response.Status)
		assert.Equal(t, req.Description, response.Description)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error when payment not found", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockPaymentRepository{}
		mockUserService := &testutil.MockUserService{}
		logger := testutil.NewSilentLogger()
		service := NewPaymentService(mockRepo, mockUserService, logger)

		paymentID := uint(999)
		req := testutil.CreateUpdatePaymentRequestFixture()

		// Mock expectations
		mockRepo.On("GetByID", paymentID).Return(nil, gorm.ErrRecordNotFound)

		// When
		response, err := service.UpdatePayment(paymentID, req)

		// Then
		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "payment not found")
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error when status is invalid", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockPaymentRepository{}
		mockUserService := &testutil.MockUserService{}
		logger := testutil.NewSilentLogger()
		service := NewPaymentService(mockRepo, mockUserService, logger)

		paymentID := uint(1)
		existingPayment := testutil.CreatePaymentFixture()
		existingPayment.ID = paymentID

		req := testutil.CreateUpdatePaymentRequestFixture()
		req.Status = "invalid_status"

		// Mock expectations
		mockRepo.On("GetByID", paymentID).Return(existingPayment, nil)

		// When
		response, err := service.UpdatePayment(paymentID, req)

		// Then
		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "invalid payment status")
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error when update fails", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockPaymentRepository{}
		mockUserService := &testutil.MockUserService{}
		logger := testutil.NewSilentLogger()
		service := NewPaymentService(mockRepo, mockUserService, logger)

		paymentID := uint(1)
		existingPayment := testutil.CreatePaymentFixture()
		existingPayment.ID = paymentID

		req := testutil.CreateUpdatePaymentRequestFixture()
		req.Status = entity.PaymentStatusCompleted.String()

		// Mock expectations
		mockRepo.On("GetByID", paymentID).Return(existingPayment, nil)
		mockRepo.On("Update", mock.AnythingOfType("*entity.Payment")).Return(errors.New("update failed"))

		// When
		response, err := service.UpdatePayment(paymentID, req)

		// Then
		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "update failed")
		mockRepo.AssertExpectations(t)
	})
}

func TestPaymentService_DeletePayment(t *testing.T) {
	t.Run("should delete payment successfully", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockPaymentRepository{}
		mockUserService := &testutil.MockUserService{}
		logger := testutil.NewSilentLogger()
		service := NewPaymentService(mockRepo, mockUserService, logger)

		paymentID := uint(1)
		payment := testutil.CreatePaymentFixture()
		payment.ID = paymentID

		// Mock expectations
		mockRepo.On("GetByID", paymentID).Return(payment, nil)
		mockRepo.On("Delete", paymentID).Return(nil)

		// When
		err := service.DeletePayment(paymentID)

		// Then
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error when payment not found", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockPaymentRepository{}
		mockUserService := &testutil.MockUserService{}
		logger := testutil.NewSilentLogger()
		service := NewPaymentService(mockRepo, mockUserService, logger)

		paymentID := uint(999)

		// Mock expectations
		mockRepo.On("GetByID", paymentID).Return(nil, gorm.ErrRecordNotFound)

		// When
		err := service.DeletePayment(paymentID)

		// Then
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "payment not found")
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error when delete fails", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockPaymentRepository{}
		mockUserService := &testutil.MockUserService{}
		logger := testutil.NewSilentLogger()
		service := NewPaymentService(mockRepo, mockUserService, logger)

		paymentID := uint(1)
		payment := testutil.CreatePaymentFixture()
		payment.ID = paymentID

		// Mock expectations
		mockRepo.On("GetByID", paymentID).Return(payment, nil)
		mockRepo.On("Delete", paymentID).Return(errors.New("delete failed"))

		// When
		err := service.DeletePayment(paymentID)

		// Then
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "delete failed")
		mockRepo.AssertExpectations(t)
	})
}

func TestPaymentService_GetPaymentsByUser(t *testing.T) {
	t.Run("should get payments by user successfully", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockPaymentRepository{}
		mockUserService := &testutil.MockUserService{}
		logger := testutil.NewSilentLogger()
		service := NewPaymentService(mockRepo, mockUserService, logger)

		userID := uint(1)
		payments := []entity.Payment{
			*testutil.CreatePaymentFixture(),
			*testutil.CreatePaymentFixture(),
		}
		payments[0].ID = 1
		payments[0].UserID = userID
		payments[1].ID = 2
		payments[1].UserID = userID
		payments[1].Amount = 200.00

		// Mock expectations
		mockRepo.On("GetByUserID", userID).Return(payments, nil)

		// When
		response, err := service.GetPaymentsByUser(userID)

		// Then
		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Len(t, response, 2)
		assert.Equal(t, uint(1), response[0].ID)
		assert.Equal(t, uint(2), response[1].ID)
		assert.Equal(t, userID, response[0].UserID)
		assert.Equal(t, userID, response[1].UserID)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return empty slice when no payments found", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockPaymentRepository{}
		mockUserService := &testutil.MockUserService{}
		logger := testutil.NewSilentLogger()
		service := NewPaymentService(mockRepo, mockUserService, logger)

		userID := uint(1)

		// Mock expectations
		mockRepo.On("GetByUserID", userID).Return([]entity.Payment{}, nil)

		// When
		response, err := service.GetPaymentsByUser(userID)

		// Then
		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Empty(t, response)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error when repository fails", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockPaymentRepository{}
		mockUserService := &testutil.MockUserService{}
		logger := testutil.NewSilentLogger()
		service := NewPaymentService(mockRepo, mockUserService, logger)

		userID := uint(1)

		// Mock expectations
		mockRepo.On("GetByUserID", userID).Return(nil, errors.New("database error"))

		// When
		response, err := service.GetPaymentsByUser(userID)

		// Then
		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "database error")
		mockRepo.AssertExpectations(t)
	})
}

func TestPaymentService_entityToResponse(t *testing.T) {
	t.Run("should convert entity to response correctly", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockPaymentRepository{}
		mockUserService := &testutil.MockUserService{}
		logger := testutil.NewSilentLogger()
		service := NewPaymentService(mockRepo, mockUserService, logger).(*paymentService)

		payment := testutil.CreatePaymentFixture()
		payment.ID = 1
		payment.Amount = 150.75
		payment.Currency = "EUR"
		payment.Status = entity.PaymentStatusCompleted
		payment.Description = "Test payment"
		payment.UserID = 2
		payment.CreatedAt = time.Now()
		payment.UpdatedAt = time.Now()

		// When
		response := service.entityToResponse(payment)

		// Then
		assert.Equal(t, payment.ID, response.ID)
		assert.Equal(t, payment.Amount, response.Amount)
		assert.Equal(t, payment.Currency, response.Currency)
		assert.Equal(t, payment.Status.String(), response.Status)
		assert.Equal(t, payment.Description, response.Description)
		assert.Equal(t, payment.UserID, response.UserID)
		assert.Equal(t, payment.CreatedAt, response.CreatedAt)
		assert.Equal(t, payment.UpdatedAt, response.UpdatedAt)
	})
}
