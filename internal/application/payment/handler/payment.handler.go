package handler

import (
	"net/http"
	"strconv"

	"vibe-ddd-golang/internal/application/payment/dto"
	"vibe-ddd-golang/internal/application/payment/service"
	"vibe-ddd-golang/internal/common/enum"
	types "vibe-ddd-golang/internal/common/type"
	"vibe-ddd-golang/internal/pkg/reqbind"
	"vibe-ddd-golang/internal/pkg/response"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type PaymentHandler struct {
	service service.PaymentService
	logger  *zap.Logger
}

func NewPaymentHandler(service service.PaymentService, logger *zap.Logger) *PaymentHandler {
	return &PaymentHandler{
		service: service,
		logger:  logger,
	}
}

func paramUint(c *gin.Context, key string) (uint, bool) {
	id, err := strconv.ParseUint(c.Param(key), 10, 32)
	if err != nil {
		response.Render(c, response.NewBadRequestException("invalid "+key))
		return 0, false
	}
	return uint(id), true
}

// CreatePayment godoc
// @Summary Create a new payment
// @Tags payments
// @Accept json
// @Produce json
// @Param payment body dto.CreatePaymentRequest true "Payment creation request"
// @Success 201 {object} types.ResponseAPI "Created payment"
// @Failure 400 {object} types.ResponseAPI "Invalid request body"
// @Router /payments [post]
func (h *PaymentHandler) CreatePayment(c *gin.Context) {
	var req dto.CreatePaymentRequest
	if err := reqbind.Bind(c, &req); err != nil {
		response.Render(c, response.NewBadRequestException("Invalid request body").WithCause(err))
		return
	}
	if err := response.Validate(c, &req); err != nil {
		return
	}

	payment, err := h.service.CreatePayment(c.Request.Context(), &req)
	if err != nil {
		response.Render(c, err)
		return
	}

	response.Send(c, &types.Response{HTTPStatus: http.StatusCreated, Code: enum.CodeCreated, Message: "Payment created.", Data: payment})
}

// GetPayment godoc
// @Summary Get a payment by ID
// @Tags payments
// @Produce json
// @Param id path int true "Payment ID"
// @Success 200 {object} types.ResponseAPI "Payment details"
// @Failure 404 {object} types.ResponseAPI "Payment not found"
// @Router /payments/{id} [get]
func (h *PaymentHandler) GetPayment(c *gin.Context) {
	id, ok := paramUint(c, "id")
	if !ok {
		return
	}

	payment, err := h.service.GetPaymentByID(c.Request.Context(), id)
	if err != nil {
		response.Render(c, err)
		return
	}

	response.Send(c, &types.Response{HTTPStatus: http.StatusOK, Code: enum.CodeOK, Message: "OK", Data: payment})
}

// GetPayments godoc
// @Summary Get all payments
// @Tags payments
// @Produce json
// @Param status query string false "Filter by status" Enums(pending, completed, failed, canceled)
// @Param currency query string false "Filter by currency (3-letter code)"
// @Param user_id query int false "Filter by user ID"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Number of items per page" default(10)
// @Success 200 {object} types.ResponseAPI "List of payments"
// @Router /payments [get]
func (h *PaymentHandler) GetPayments(c *gin.Context) {
	var filter dto.PaymentFilter
	if err := reqbind.BindQuery(c, &filter); err != nil {
		response.Render(c, response.NewBadRequestException("Invalid query parameters").WithCause(err))
		return
	}

	payments, err := h.service.GetPayments(c.Request.Context(), &filter)
	if err != nil {
		response.Render(c, err)
		return
	}

	response.Send(c, &types.Response{HTTPStatus: http.StatusOK, Code: enum.CodeOK, Message: "OK", Data: payments})
}

// UpdatePayment godoc
// @Summary Update a payment
// @Tags payments
// @Accept json
// @Produce json
// @Param id path int true "Payment ID"
// @Param payment body dto.UpdatePaymentRequest true "Payment update request"
// @Success 200 {object} types.ResponseAPI "Updated payment"
// @Failure 404 {object} types.ResponseAPI "Payment not found"
// @Router /payments/{id} [put]
func (h *PaymentHandler) UpdatePayment(c *gin.Context) {
	id, ok := paramUint(c, "id")
	if !ok {
		return
	}

	var req dto.UpdatePaymentRequest
	if err := reqbind.Bind(c, &req); err != nil {
		response.Render(c, response.NewBadRequestException("Invalid request body").WithCause(err))
		return
	}
	if err := response.Validate(c, &req); err != nil {
		return
	}

	payment, err := h.service.UpdatePayment(c.Request.Context(), id, &req)
	if err != nil {
		response.Render(c, err)
		return
	}

	response.Send(c, &types.Response{HTTPStatus: http.StatusOK, Code: enum.CodeOK, Message: "Payment updated.", Data: payment})
}

// DeletePayment godoc
// @Summary Delete a payment
// @Tags payments
// @Produce json
// @Param id path int true "Payment ID"
// @Success 200 {object} types.ResponseAPI "Payment deleted"
// @Failure 404 {object} types.ResponseAPI "Payment not found"
// @Router /payments/{id} [delete]
func (h *PaymentHandler) DeletePayment(c *gin.Context) {
	id, ok := paramUint(c, "id")
	if !ok {
		return
	}

	if err := h.service.DeletePayment(c.Request.Context(), id); err != nil {
		response.Render(c, err)
		return
	}

	response.Send(c, &types.Response{HTTPStatus: http.StatusOK, Code: enum.CodeOK, Message: "Payment deleted.", Data: nil})
}

// GetPaymentsByUser godoc
// @Summary Get payments by user ID
// @Tags payments
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} types.ResponseAPI "List of payments for the user"
// @Router /users/{id}/payments [get]
func (h *PaymentHandler) GetPaymentsByUser(c *gin.Context) {
	userID, ok := paramUint(c, "id")
	if !ok {
		return
	}

	payments, err := h.service.GetPaymentsByUser(c.Request.Context(), userID)
	if err != nil {
		response.Render(c, err)
		return
	}

	response.Send(c, &types.Response{HTTPStatus: http.StatusOK, Code: enum.CodeOK, Message: "OK", Data: payments})
}

func (h *PaymentHandler) RegisterRoutes(api *gin.RouterGroup) {
	payments := api.Group("/payments")
	{
		payments.POST("", h.CreatePayment)
		payments.GET("", h.GetPayments)
		payments.GET("/:id", h.GetPayment)
		payments.PUT("/:id", h.UpdatePayment)
		payments.DELETE("/:id", h.DeletePayment)
	}

	users := api.Group("/users")
	{
		users.GET("/:id/payments", h.GetPaymentsByUser)
	}
}
