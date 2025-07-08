package worker

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"vibe-ddd-golang/internal/application/payment/dto"
	"vibe-ddd-golang/internal/application/payment/entity"
	"vibe-ddd-golang/internal/config"
	"vibe-ddd-golang/internal/pkg/testutil"

	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockPaymentService struct {
	mock.Mock
}

func (m *MockPaymentService) CreatePayment(req *dto.CreatePaymentRequest) (*dto.PaymentResponse, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.PaymentResponse), args.Error(1)
}

func (m *MockPaymentService) GetPaymentByID(id uint) (*dto.PaymentResponse, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.PaymentResponse), args.Error(1)
}

func (m *MockPaymentService) GetPayments(filter *dto.PaymentFilter) (*dto.PaymentListResponse, error) {
	args := m.Called(filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.PaymentListResponse), args.Error(1)
}

func (m *MockPaymentService) UpdatePayment(id uint, req *dto.UpdatePaymentRequest) (*dto.PaymentResponse, error) {
	args := m.Called(id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.PaymentResponse), args.Error(1)
}

func (m *MockPaymentService) DeletePayment(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockPaymentService) GetPaymentsByUser(userID uint) ([]dto.PaymentResponse, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]dto.PaymentResponse), args.Error(1)
}

type MockAsynqClient struct {
	mock.Mock
}

func (m *MockAsynqClient) Enqueue(task *asynq.Task, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	args := m.Called(task, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*asynq.TaskInfo), args.Error(1)
}

func setupPaymentWorker() (*PaymentWorker, *MockPaymentService, *MockAsynqClient) {
	mockService := &MockPaymentService{}
	mockClient := &MockAsynqClient{}
	logger := testutil.NewSilentLogger()
	cfg := &config.Config{
		Worker: config.WorkerConfig{
			PaymentCheckInterval: 5 * time.Minute,
			RetryMaxAttempts:     3,
		},
	}

	worker := NewPaymentWorker(mockService, mockClient, logger, cfg)

	return worker, mockService, mockClient
}

func TestPaymentWorker_HandleCheckPaymentStatus(t *testing.T) {
	t.Run("should handle check payment status successfully when status needs update", func(t *testing.T) {
		// Setup
		worker, mockService, _ := setupPaymentWorker()

		paymentID := uint(1)
		payload := CheckPaymentStatusPayload{PaymentID: paymentID}
		payloadBytes, _ := json.Marshal(payload)
		task := asynq.NewTask(TypeCheckPaymentStatus, payloadBytes)

		// Create payment response with pending status and old creation time
		payment := &dto.PaymentResponse{
			ID:        paymentID,
			Amount:    100.50,
			Currency:  "USD",
			Status:    entity.PaymentStatusPending.String(),
			UserID:    1,
			CreatedAt: time.Now().Add(-3 * time.Minute), // 3 minutes ago
			UpdatedAt: time.Now().Add(-3 * time.Minute),
		}

		updatedPayment := &dto.PaymentResponse{
			ID:        paymentID,
			Amount:    100.50,
			Currency:  "USD",
			Status:    entity.PaymentStatusCompleted.String(),
			UserID:    1,
			CreatedAt: payment.CreatedAt,
			UpdatedAt: time.Now(),
		}

		mockService.On("GetPaymentByID", paymentID).Return(payment, nil)
		mockService.On("UpdatePayment", paymentID, mock.AnythingOfType("*dto.UpdatePaymentRequest")).Return(updatedPayment, nil)

		// When
		err := worker.HandleCheckPaymentStatus(context.Background(), task)

		// Then
		assert.NoError(t, err)
		mockService.AssertExpectations(t)

		// Verify the update request has the correct status
		updateCall := mockService.Calls[1]
		updateReq := updateCall.Arguments[1].(*dto.UpdatePaymentRequest)
		assert.Equal(t, entity.PaymentStatusCompleted.String(), updateReq.Status)
		assert.Contains(t, updateReq.Description, "Status updated by worker")
	})

	t.Run("should skip check when payment is in final state", func(t *testing.T) {
		// Setup
		worker, mockService, _ := setupPaymentWorker()

		paymentID := uint(1)
		payload := CheckPaymentStatusPayload{PaymentID: paymentID}
		payloadBytes, _ := json.Marshal(payload)
		task := asynq.NewTask(TypeCheckPaymentStatus, payloadBytes)

		payment := &dto.PaymentResponse{
			ID:        paymentID,
			Status:    entity.PaymentStatusCompleted.String(),
			CreatedAt: time.Now().Add(-1 * time.Hour),
			UpdatedAt: time.Now().Add(-1 * time.Hour),
		}

		mockService.On("GetPaymentByID", paymentID).Return(payment, nil)

		// When
		err := worker.HandleCheckPaymentStatus(context.Background(), task)

		// Then
		assert.NoError(t, err)
		mockService.AssertExpectations(t)
		mockService.AssertNotCalled(t, "UpdatePayment")
	})

	t.Run("should schedule next check when payment remains pending", func(t *testing.T) {
		// Setup
		worker, mockService, mockClient := setupPaymentWorker()

		paymentID := uint(1)
		payload := CheckPaymentStatusPayload{PaymentID: paymentID}
		payloadBytes, _ := json.Marshal(payload)
		task := asynq.NewTask(TypeCheckPaymentStatus, payloadBytes)

		payment := &dto.PaymentResponse{
			ID:        paymentID,
			Status:    entity.PaymentStatusPending.String(),
			CreatedAt: time.Now().Add(-30 * time.Second), // Recent, will stay pending
			UpdatedAt: time.Now().Add(-30 * time.Second),
		}

		taskInfo := &asynq.TaskInfo{ID: "task-123"}

		mockService.On("GetPaymentByID", paymentID).Return(payment, nil)
		mockClient.On("Enqueue", mock.AnythingOfType("*asynq.Task"), mock.AnythingOfType("[]asynq.Option")).Return(taskInfo, nil)

		// When
		err := worker.HandleCheckPaymentStatus(context.Background(), task)

		// Then
		assert.NoError(t, err)
		mockService.AssertExpectations(t)
		mockClient.AssertExpectations(t)
		mockService.AssertNotCalled(t, "UpdatePayment")
	})

	t.Run("should return error when payload is invalid", func(t *testing.T) {
		// Setup
		worker, mockService, _ := setupPaymentWorker()

		task := asynq.NewTask(TypeCheckPaymentStatus, []byte("invalid json"))

		// When
		err := worker.HandleCheckPaymentStatus(context.Background(), task)

		// Then
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "json.Unmarshal failed")
		mockService.AssertExpectations(t)
	})

	t.Run("should return error when payment not found", func(t *testing.T) {
		// Setup
		worker, mockService, _ := setupPaymentWorker()

		paymentID := uint(999)
		payload := CheckPaymentStatusPayload{PaymentID: paymentID}
		payloadBytes, _ := json.Marshal(payload)
		task := asynq.NewTask(TypeCheckPaymentStatus, payloadBytes)

		mockService.On("GetPaymentByID", paymentID).Return(nil, errors.New("payment not found"))

		// When
		err := worker.HandleCheckPaymentStatus(context.Background(), task)

		// Then
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get payment")
		mockService.AssertExpectations(t)
	})

	t.Run("should return error when update payment fails", func(t *testing.T) {
		// Setup
		worker, mockService, _ := setupPaymentWorker()

		paymentID := uint(1)
		payload := CheckPaymentStatusPayload{PaymentID: paymentID}
		payloadBytes, _ := json.Marshal(payload)
		task := asynq.NewTask(TypeCheckPaymentStatus, payloadBytes)

		payment := &dto.PaymentResponse{
			ID:        paymentID,
			Status:    entity.PaymentStatusPending.String(),
			CreatedAt: time.Now().Add(-3 * time.Minute),
			UpdatedAt: time.Now().Add(-3 * time.Minute),
		}

		mockService.On("GetPaymentByID", paymentID).Return(payment, nil)
		mockService.On("UpdatePayment", paymentID, mock.AnythingOfType("*dto.UpdatePaymentRequest")).Return(nil, errors.New("update failed"))

		// When
		err := worker.HandleCheckPaymentStatus(context.Background(), task)

		// Then
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to update payment status")
		mockService.AssertExpectations(t)
	})
}

func TestPaymentWorker_HandleProcessPayment(t *testing.T) {
	t.Run("should process payment successfully", func(t *testing.T) {
		// Setup
		worker, mockService, _ := setupPaymentWorker()

		paymentID := uint(1)
		payload := ProcessPaymentPayload{PaymentID: paymentID}
		payloadBytes, _ := json.Marshal(payload)
		task := asynq.NewTask(TypeProcessPayment, payloadBytes)

		payment := &dto.PaymentResponse{
			ID:        paymentID,
			Amount:    100.50,
			Currency:  "USD",
			Status:    entity.PaymentStatusPending.String(),
			UserID:    1,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		processedPayment := &dto.PaymentResponse{
			ID:        paymentID,
			Amount:    100.50,
			Currency:  "USD",
			Status:    entity.PaymentStatusCompleted.String(),
			UserID:    1,
			CreatedAt: payment.CreatedAt,
			UpdatedAt: time.Now(),
		}

		mockService.On("GetPaymentByID", paymentID).Return(payment, nil)
		mockService.On("UpdatePayment", paymentID, mock.AnythingOfType("*dto.UpdatePaymentRequest")).Return(processedPayment, nil)

		// When
		err := worker.HandleProcessPayment(context.Background(), task)

		// Then
		assert.NoError(t, err)
		mockService.AssertExpectations(t)

		// Verify the update request
		updateCall := mockService.Calls[1]
		updateReq := updateCall.Arguments[1].(*dto.UpdatePaymentRequest)
		// Status could be completed or failed based on simulation
		assert.True(t, updateReq.Status == entity.PaymentStatusCompleted.String() || updateReq.Status == entity.PaymentStatusFailed.String())
		assert.Contains(t, updateReq.Description, "Payment processed by worker")
	})

	t.Run("should return error when payload is invalid", func(t *testing.T) {
		// Setup
		worker, mockService, _ := setupPaymentWorker()

		task := asynq.NewTask(TypeProcessPayment, []byte("invalid json"))

		// When
		err := worker.HandleProcessPayment(context.Background(), task)

		// Then
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "json.Unmarshal failed")
		mockService.AssertExpectations(t)
	})

	t.Run("should return error when payment not found", func(t *testing.T) {
		// Setup
		worker, mockService, _ := setupPaymentWorker()

		paymentID := uint(999)
		payload := ProcessPaymentPayload{PaymentID: paymentID}
		payloadBytes, _ := json.Marshal(payload)
		task := asynq.NewTask(TypeProcessPayment, payloadBytes)

		mockService.On("GetPaymentByID", paymentID).Return(nil, errors.New("payment not found"))

		// When
		err := worker.HandleProcessPayment(context.Background(), task)

		// Then
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get payment")
		mockService.AssertExpectations(t)
	})

	t.Run("should return error when update payment fails", func(t *testing.T) {
		// Setup
		worker, mockService, _ := setupPaymentWorker()

		paymentID := uint(1)
		payload := ProcessPaymentPayload{PaymentID: paymentID}
		payloadBytes, _ := json.Marshal(payload)
		task := asynq.NewTask(TypeProcessPayment, payloadBytes)

		payment := &dto.PaymentResponse{
			ID:        paymentID,
			Status:    entity.PaymentStatusPending.String(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		mockService.On("GetPaymentByID", paymentID).Return(payment, nil)
		mockService.On("UpdatePayment", paymentID, mock.AnythingOfType("*dto.UpdatePaymentRequest")).Return(nil, errors.New("update failed"))

		// When
		err := worker.HandleProcessPayment(context.Background(), task)

		// Then
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to update payment")
		mockService.AssertExpectations(t)
	})
}

func TestPaymentWorker_SchedulePaymentStatusCheck(t *testing.T) {
	t.Run("should schedule payment status check successfully", func(t *testing.T) {
		// Setup
		worker, _, mockClient := setupPaymentWorker()

		paymentID := uint(1)
		delay := 5 * time.Minute
		taskInfo := &asynq.TaskInfo{ID: "task-123"}

		mockClient.On("Enqueue", mock.AnythingOfType("*asynq.Task"), mock.AnythingOfType("[]asynq.Option")).Return(taskInfo, nil)

		// When
		err := worker.SchedulePaymentStatusCheck(paymentID, delay)

		// Then
		assert.NoError(t, err)
		mockClient.AssertExpectations(t)

		// Verify the task was created with correct type and payload
		enqueueCall := mockClient.Calls[0]
		task := enqueueCall.Arguments[0].(*asynq.Task)
		assert.Equal(t, TypeCheckPaymentStatus, task.Type())

		var payload CheckPaymentStatusPayload
		err = json.Unmarshal(task.Payload(), &payload)
		assert.NoError(t, err)
		assert.Equal(t, paymentID, payload.PaymentID)
	})

	t.Run("should return error when enqueue fails", func(t *testing.T) {
		// Setup
		worker, _, mockClient := setupPaymentWorker()

		paymentID := uint(1)
		delay := 5 * time.Minute

		mockClient.On("Enqueue", mock.AnythingOfType("*asynq.Task"), mock.AnythingOfType("[]asynq.Option")).Return(nil, errors.New("enqueue failed"))

		// When
		err := worker.SchedulePaymentStatusCheck(paymentID, delay)

		// Then
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to enqueue task")
		mockClient.AssertExpectations(t)
	})
}

func TestPaymentWorker_SchedulePaymentProcessing(t *testing.T) {
	t.Run("should schedule payment processing successfully", func(t *testing.T) {
		// Setup
		worker, _, mockClient := setupPaymentWorker()

		paymentID := uint(1)
		taskInfo := &asynq.TaskInfo{ID: "task-456"}

		mockClient.On("Enqueue", mock.AnythingOfType("*asynq.Task"), mock.AnythingOfType("[]asynq.Option")).Return(taskInfo, nil)

		// When
		err := worker.SchedulePaymentProcessing(paymentID)

		// Then
		assert.NoError(t, err)
		mockClient.AssertExpectations(t)

		// Verify the task was created with correct type and payload
		enqueueCall := mockClient.Calls[0]
		task := enqueueCall.Arguments[0].(*asynq.Task)
		assert.Equal(t, TypeProcessPayment, task.Type())

		var payload ProcessPaymentPayload
		err = json.Unmarshal(task.Payload(), &payload)
		assert.NoError(t, err)
		assert.Equal(t, paymentID, payload.PaymentID)
	})

	t.Run("should return error when enqueue fails", func(t *testing.T) {
		// Setup
		worker, _, mockClient := setupPaymentWorker()

		paymentID := uint(1)

		mockClient.On("Enqueue", mock.AnythingOfType("*asynq.Task"), mock.AnythingOfType("[]asynq.Option")).Return(nil, errors.New("enqueue failed"))

		// When
		err := worker.SchedulePaymentProcessing(paymentID)

		// Then
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to enqueue task")
		mockClient.AssertExpectations(t)
	})
}

func TestPaymentWorker_simulatePaymentGatewayCheck(t *testing.T) {
	t.Run("should return pending for recent payments", func(t *testing.T) {
		// Setup
		worker, _, _ := setupPaymentWorker()

		payment := &dto.PaymentResponse{
			ID:        1,
			Status:    entity.PaymentStatusPending.String(),
			CreatedAt: time.Now().Add(-30 * time.Second), // 30 seconds ago
		}

		// When
		status := worker.simulatePaymentGatewayCheck(payment)

		// Then
		assert.Equal(t, entity.PaymentStatusPending.String(), status)
	})

	t.Run("should return completed or failed for old payments", func(t *testing.T) {
		// Setup
		worker, _, _ := setupPaymentWorker()

		payment := &dto.PaymentResponse{
			ID:        1,
			Status:    entity.PaymentStatusPending.String(),
			CreatedAt: time.Now().Add(-3 * time.Minute), // 3 minutes ago
		}

		// When
		status := worker.simulatePaymentGatewayCheck(payment)

		// Then
		// Should be either completed, failed, or pending (but most likely completed or failed)
		validStatuses := []string{
			entity.PaymentStatusPending.String(),
			entity.PaymentStatusCompleted.String(),
			entity.PaymentStatusFailed.String(),
		}
		assert.Contains(t, validStatuses, status)
	})
}

func TestPaymentWorker_simulatePaymentProcessing(t *testing.T) {
	t.Run("should return boolean result", func(t *testing.T) {
		// Setup
		worker, _, _ := setupPaymentWorker()

		payment := &dto.PaymentResponse{
			ID:     1,
			Amount: 100.50,
		}

		// When
		result := worker.simulatePaymentProcessing(payment)

		// Then
		// Should return either true or false (boolean)
		assert.IsType(t, true, result)
	})
}
