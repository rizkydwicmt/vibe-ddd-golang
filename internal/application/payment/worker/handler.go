package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"vibe-ddd-golang/internal/application/payment/dto"
	"vibe-ddd-golang/internal/application/payment/entity"
	"vibe-ddd-golang/internal/application/payment/service"
	"vibe-ddd-golang/internal/config"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"
)

type AsynqClient interface {
	Enqueue(task *asynq.Task, opts ...asynq.Option) (*asynq.TaskInfo, error)
}

type PaymentWorker struct {
	paymentService service.PaymentService
	client         AsynqClient
	logger         *zap.Logger
	cfg            *config.Config
}

type CheckPaymentStatusPayload struct {
	PaymentID uint `json:"payment_id"`
}

type ProcessPaymentPayload struct {
	PaymentID uint `json:"payment_id"`
}

func NewPaymentWorker(
	paymentService service.PaymentService,
	client AsynqClient,
	logger *zap.Logger,
	cfg *config.Config,
) *PaymentWorker {
	return &PaymentWorker{
		paymentService: paymentService,
		client:         client,
		logger:         logger,
		cfg:            cfg,
	}
}

func (w *PaymentWorker) HandleCheckPaymentStatus(ctx context.Context, task *asynq.Task) error {
	var payload CheckPaymentStatusPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		w.logger.Error("Failed to unmarshal payment status check payload",
			zap.Error(err),
			zap.ByteString("payload", task.Payload()))
		return fmt.Errorf("json.Unmarshal failed: %w", err)
	}

	w.logger.Info("Processing payment status check",
		zap.Uint("payment_id", payload.PaymentID))

	// Get payment from database
	payment, err := w.paymentService.GetPaymentByID(payload.PaymentID)
	if err != nil {
		w.logger.Error("Failed to get payment",
			zap.Uint("payment_id", payload.PaymentID),
			zap.Error(err))
		return fmt.Errorf("failed to get payment: %w", err)
	}

	// Skip if payment is already completed or failed
	if payment.Status == entity.PaymentStatusCompleted.String() ||
		payment.Status == entity.PaymentStatusFailed.String() ||
		payment.Status == entity.PaymentStatusCanceled.String() {
		w.logger.Info("Payment already in final state, skipping check",
			zap.Uint("payment_id", payload.PaymentID),
			zap.String("status", payment.Status))
		return nil
	}

	// Simulate external payment gateway status check
	// In real implementation, you would call external payment gateway API
	newStatus := w.simulatePaymentGatewayCheck(payment)

	// Update payment status if changed
	if newStatus != payment.Status {
		updateReq := &dto.UpdatePaymentRequest{
			Status:      newStatus,
			Description: fmt.Sprintf("Status updated by worker at %s", time.Now().Format(time.RFC3339)),
		}

		_, err := w.paymentService.UpdatePayment(payload.PaymentID, updateReq)
		if err != nil {
			w.logger.Error("Failed to update payment status",
				zap.Uint("payment_id", payload.PaymentID),
				zap.String("new_status", newStatus),
				zap.Error(err))
			return fmt.Errorf("failed to update payment status: %w", err)
		}

		w.logger.Info("Payment status updated",
			zap.Uint("payment_id", payload.PaymentID),
			zap.String("old_status", payment.Status),
			zap.String("new_status", newStatus))
	}

	// Schedule next check if payment is still pending
	if newStatus == entity.PaymentStatusPending.String() {
		if err := w.SchedulePaymentStatusCheck(payload.PaymentID, w.cfg.Worker.PaymentCheckInterval); err != nil {
			w.logger.Error("Failed to schedule next payment check",
				zap.Uint("payment_id", payload.PaymentID),
				zap.Error(err))
			// Don't return error as the current task was successful
		}
	}

	return nil
}

func (w *PaymentWorker) HandleProcessPayment(ctx context.Context, task *asynq.Task) error {
	var payload ProcessPaymentPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		w.logger.Error("Failed to unmarshal process payment payload",
			zap.Error(err),
			zap.ByteString("payload", task.Payload()))
		return fmt.Errorf("json.Unmarshal failed: %w", err)
	}

	w.logger.Info("Processing payment",
		zap.Uint("payment_id", payload.PaymentID))

	// Get payment from database
	payment, err := w.paymentService.GetPaymentByID(payload.PaymentID)
	if err != nil {
		w.logger.Error("Failed to get payment for processing",
			zap.Uint("payment_id", payload.PaymentID),
			zap.Error(err))
		return fmt.Errorf("failed to get payment: %w", err)
	}

	// Simulate payment processing
	// In real implementation, you would call external payment gateway
	success := w.simulatePaymentProcessing(payment)

	var newStatus string
	if success {
		newStatus = entity.PaymentStatusCompleted.String()
	} else {
		newStatus = entity.PaymentStatusFailed.String()
	}

	updateReq := &dto.UpdatePaymentRequest{
		Status:      newStatus,
		Description: fmt.Sprintf("Payment processed by worker at %s", time.Now().Format(time.RFC3339)),
	}

	_, err = w.paymentService.UpdatePayment(payload.PaymentID, updateReq)
	if err != nil {
		w.logger.Error("Failed to update payment after processing",
			zap.Uint("payment_id", payload.PaymentID),
			zap.String("new_status", newStatus),
			zap.Error(err))
		return fmt.Errorf("failed to update payment: %w", err)
	}

	w.logger.Info("Payment processing completed",
		zap.Uint("payment_id", payload.PaymentID),
		zap.String("final_status", newStatus),
		zap.Bool("success", success))

	return nil
}

func (w *PaymentWorker) SchedulePaymentStatusCheck(paymentID uint, delay time.Duration) error {
	payload := CheckPaymentStatusPayload{PaymentID: paymentID}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	task := asynq.NewTask(TypeCheckPaymentStatus, payloadBytes)
	opts := []asynq.Option{
		asynq.ProcessIn(delay),
		asynq.Queue("default"),
		asynq.MaxRetry(w.cfg.Worker.RetryMaxAttempts),
	}

	info, err := w.client.Enqueue(task, opts...)
	if err != nil {
		return fmt.Errorf("failed to enqueue task: %w", err)
	}

	w.logger.Info("Scheduled payment status check",
		zap.Uint("payment_id", paymentID),
		zap.Duration("delay", delay),
		zap.String("task_id", info.ID))

	return nil
}

func (w *PaymentWorker) SchedulePaymentProcessing(paymentID uint) error {
	payload := ProcessPaymentPayload{PaymentID: paymentID}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	task := asynq.NewTask(TypeProcessPayment, payloadBytes)
	opts := []asynq.Option{
		asynq.Queue("critical"),
		asynq.MaxRetry(w.cfg.Worker.RetryMaxAttempts),
	}

	info, err := w.client.Enqueue(task, opts...)
	if err != nil {
		return fmt.Errorf("failed to enqueue task: %w", err)
	}

	w.logger.Info("Scheduled payment processing",
		zap.Uint("payment_id", paymentID),
		zap.String("task_id", info.ID))

	return nil
}

// simulatePaymentGatewayCheck simulates checking payment status with external gateway
func (w *PaymentWorker) simulatePaymentGatewayCheck(payment *dto.PaymentResponse) string {
	// Simulate random status changes for demo purposes
	// In real implementation, this would call actual payment gateway API

	elapsed := time.Since(payment.CreatedAt)

	// After 2 minutes, 80% chance to complete, 10% to fail, 10% stay pending
	if elapsed > 2*time.Minute {
		rand := time.Now().UnixNano() % 10
		if rand < 8 {
			return entity.PaymentStatusCompleted.String()
		} else if rand < 9 {
			return entity.PaymentStatusFailed.String()
		}
	}

	return entity.PaymentStatusPending.String()
}

// simulatePaymentProcessing simulates processing payment with external gateway
func (w *PaymentWorker) simulatePaymentProcessing(payment *dto.PaymentResponse) bool {
	// Simulate 90% success rate for demo purposes
	rand := time.Now().UnixNano() % 10
	return rand < 9
}
