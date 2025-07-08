package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"vibe-ddd-golang/internal/application/payment/dto"
	"vibe-ddd-golang/internal/application/payment/entity"
	"vibe-ddd-golang/internal/pkg/testutil"

	"github.com/gin-gonic/gin"
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

func setupPaymentHandler() (*PaymentHandler, *MockPaymentService) {
	gin.SetMode(gin.TestMode)
	mockService := &MockPaymentService{}
	logger := testutil.NewSilentLogger()
	handler := NewPaymentHandler(mockService, logger)
	return handler, mockService
}

func TestPaymentHandler_CreatePayment(t *testing.T) {
	t.Run("should create payment successfully", func(t *testing.T) {
		// Setup
		handler, mockService := setupPaymentHandler()

		req := testutil.CreatePaymentRequestFixture()
		response := &dto.PaymentResponse{
			ID:          1,
			Amount:      req.Amount,
			Currency:    req.Currency,
			Status:      entity.PaymentStatusPending.String(),
			Description: req.Description,
			UserID:      req.UserID,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		mockService.On("CreatePayment", mock.AnythingOfType("*dto.CreatePaymentRequest")).Return(response, nil)

		reqBody, _ := json.Marshal(req)
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("POST", "/payments", bytes.NewBuffer(reqBody))
		ctx.Request.Header.Set("Content-Type", "application/json")

		// When
		handler.CreatePayment(ctx)

		// Then
		assert.Equal(t, http.StatusCreated, w.Code)
		mockService.AssertExpectations(t)

		var result map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &result)
		assert.Contains(t, result, "data")
		data := result["data"].(map[string]interface{})
		assert.Equal(t, float64(1), data["id"])
		assert.Equal(t, req.Amount, data["amount"])
		assert.Equal(t, req.Currency, data["currency"])
	})

	t.Run("should return bad request for invalid JSON", func(t *testing.T) {
		// Setup
		handler, mockService := setupPaymentHandler()

		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("POST", "/payments", bytes.NewBuffer([]byte("invalid json")))
		ctx.Request.Header.Set("Content-Type", "application/json")

		// When
		handler.CreatePayment(ctx)

		// Then
		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("should return internal api error when service fails", func(t *testing.T) {
		// Setup
		handler, mockService := setupPaymentHandler()

		req := testutil.CreatePaymentRequestFixture()
		mockService.On("CreatePayment", mock.AnythingOfType("*dto.CreatePaymentRequest")).Return(nil, errors.New("service error"))

		reqBody, _ := json.Marshal(req)
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("POST", "/payments", bytes.NewBuffer(reqBody))
		ctx.Request.Header.Set("Content-Type", "application/json")

		// When
		handler.CreatePayment(ctx)

		// Then
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockService.AssertExpectations(t)
	})
}

func TestPaymentHandler_GetPayment(t *testing.T) {
	t.Run("should get payment successfully", func(t *testing.T) {
		// Setup
		handler, mockService := setupPaymentHandler()

		paymentID := uint(1)
		response := &dto.PaymentResponse{
			ID:          paymentID,
			Amount:      100.50,
			Currency:    "USD",
			Status:      entity.PaymentStatusPending.String(),
			Description: "Test payment",
			UserID:      1,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		mockService.On("GetPaymentByID", paymentID).Return(response, nil)

		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("GET", "/payments/1", nil)
		ctx.Params = gin.Params{
			{Key: "id", Value: "1"},
		}

		// When
		handler.GetPayment(ctx)

		// Then
		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)

		var result map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &result)
		assert.Contains(t, result, "data")
		data := result["data"].(map[string]interface{})
		assert.Equal(t, float64(1), data["id"])
		assert.Equal(t, 100.50, data["amount"])
	})

	t.Run("should return bad request for invalid ID", func(t *testing.T) {
		// Setup
		handler, mockService := setupPaymentHandler()

		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("GET", "/payments/invalid", nil)
		ctx.Params = gin.Params{
			{Key: "id", Value: "invalid"},
		}

		// When
		handler.GetPayment(ctx)

		// Then
		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("should return not found when payment not found", func(t *testing.T) {
		// Setup
		handler, mockService := setupPaymentHandler()

		paymentID := uint(999)
		mockService.On("GetPaymentByID", paymentID).Return(nil, errors.New("payment not found"))

		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("GET", "/payments/999", nil)
		ctx.Params = gin.Params{
			{Key: "id", Value: "999"},
		}

		// When
		handler.GetPayment(ctx)

		// Then
		assert.Equal(t, http.StatusNotFound, w.Code)
		mockService.AssertExpectations(t)
	})
}

func TestPaymentHandler_GetPayments(t *testing.T) {
	t.Run("should get payments successfully", func(t *testing.T) {
		// Setup
		handler, mockService := setupPaymentHandler()

		response := &dto.PaymentListResponse{
			Data: []dto.PaymentResponse{
				{ID: 1, Amount: 100.50, Currency: "USD", Status: "pending"},
				{ID: 2, Amount: 200.75, Currency: "EUR", Status: "completed"},
			},
			TotalCount: 2,
			Page:       1,
			PageSize:   10,
		}

		mockService.On("GetPayments", mock.AnythingOfType("*dto.PaymentFilter")).Return(response, nil)

		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("GET", "/payments?page=1&page_size=10", nil)

		// When
		handler.GetPayments(ctx)

		// Then
		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)

		var result dto.PaymentListResponse
		json.Unmarshal(w.Body.Bytes(), &result)
		assert.Len(t, result.Data, 2)
		assert.Equal(t, int64(2), result.TotalCount)
	})

	t.Run("should return bad request for invalid query parameters", func(t *testing.T) {
		// Setup
		handler, mockService := setupPaymentHandler()

		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("GET", "/payments?page=invalid", nil)

		// When
		handler.GetPayments(ctx)

		// Then
		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("should return internal api error when service fails", func(t *testing.T) {
		// Setup
		handler, mockService := setupPaymentHandler()

		mockService.On("GetPayments", mock.AnythingOfType("*dto.PaymentFilter")).Return(nil, errors.New("database error"))

		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("GET", "/payments", nil)

		// When
		handler.GetPayments(ctx)

		// Then
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockService.AssertExpectations(t)
	})
}

func TestPaymentHandler_UpdatePayment(t *testing.T) {
	t.Run("should update payment successfully", func(t *testing.T) {
		// Setup
		handler, mockService := setupPaymentHandler()

		paymentID := uint(1)
		req := testutil.CreateUpdatePaymentRequestFixture()
		response := &dto.PaymentResponse{
			ID:          paymentID,
			Amount:      100.50,
			Currency:    "USD",
			Status:      req.Status,
			Description: req.Description,
			UserID:      1,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		mockService.On("UpdatePayment", paymentID, mock.AnythingOfType("*dto.UpdatePaymentRequest")).Return(response, nil)

		reqBody, _ := json.Marshal(req)
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("PUT", "/payments/1", bytes.NewBuffer(reqBody))
		ctx.Request.Header.Set("Content-Type", "application/json")
		ctx.Params = gin.Params{
			{Key: "id", Value: "1"},
		}

		// When
		handler.UpdatePayment(ctx)

		// Then
		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)

		var result map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &result)
		assert.Contains(t, result, "data")
		data := result["data"].(map[string]interface{})
		assert.Equal(t, float64(1), data["id"])
		assert.Equal(t, req.Status, data["status"])
	})

	t.Run("should return bad request for invalid ID", func(t *testing.T) {
		// Setup
		handler, mockService := setupPaymentHandler()

		req := testutil.CreateUpdatePaymentRequestFixture()
		reqBody, _ := json.Marshal(req)
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("PUT", "/payments/invalid", bytes.NewBuffer(reqBody))
		ctx.Request.Header.Set("Content-Type", "application/json")
		ctx.Params = gin.Params{
			{Key: "id", Value: "invalid"},
		}

		// When
		handler.UpdatePayment(ctx)

		// Then
		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("should return bad request for invalid JSON", func(t *testing.T) {
		// Setup
		handler, mockService := setupPaymentHandler()

		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("PUT", "/payments/1", bytes.NewBuffer([]byte("invalid json")))
		ctx.Request.Header.Set("Content-Type", "application/json")
		ctx.Params = gin.Params{
			{Key: "id", Value: "1"},
		}

		// When
		handler.UpdatePayment(ctx)

		// Then
		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("should return internal api error when service fails", func(t *testing.T) {
		// Setup
		handler, mockService := setupPaymentHandler()

		paymentID := uint(1)
		req := testutil.CreateUpdatePaymentRequestFixture()
		mockService.On("UpdatePayment", paymentID, mock.AnythingOfType("*dto.UpdatePaymentRequest")).Return(nil, errors.New("service error"))

		reqBody, _ := json.Marshal(req)
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("PUT", "/payments/1", bytes.NewBuffer(reqBody))
		ctx.Request.Header.Set("Content-Type", "application/json")
		ctx.Params = gin.Params{
			{Key: "id", Value: "1"},
		}

		// When
		handler.UpdatePayment(ctx)

		// Then
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockService.AssertExpectations(t)
	})
}

func TestPaymentHandler_DeletePayment(t *testing.T) {
	t.Run("should delete payment successfully", func(t *testing.T) {
		// Setup
		handler, mockService := setupPaymentHandler()

		paymentID := uint(1)
		mockService.On("DeletePayment", paymentID).Return(nil)

		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("DELETE", "/payments/1", nil)
		ctx.Params = gin.Params{
			{Key: "id", Value: "1"},
		}

		// When
		handler.DeletePayment(ctx)

		// Then
		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)

		var result map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &result)
		assert.Contains(t, result, "message")
		assert.Equal(t, "Payment deleted successfully", result["message"])
	})

	t.Run("should return bad request for invalid ID", func(t *testing.T) {
		// Setup
		handler, mockService := setupPaymentHandler()

		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("DELETE", "/payments/invalid", nil)
		ctx.Params = gin.Params{
			{Key: "id", Value: "invalid"},
		}

		// When
		handler.DeletePayment(ctx)

		// Then
		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("should return internal api error when service fails", func(t *testing.T) {
		// Setup
		handler, mockService := setupPaymentHandler()

		paymentID := uint(1)
		mockService.On("DeletePayment", paymentID).Return(errors.New("service error"))

		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("DELETE", "/payments/1", nil)
		ctx.Params = gin.Params{
			{Key: "id", Value: "1"},
		}

		// When
		handler.DeletePayment(ctx)

		// Then
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockService.AssertExpectations(t)
	})
}

func TestPaymentHandler_GetPaymentsByUser(t *testing.T) {
	t.Run("should get payments by user successfully", func(t *testing.T) {
		// Setup
		handler, mockService := setupPaymentHandler()

		userID := uint(1)
		response := []dto.PaymentResponse{
			{ID: 1, Amount: 100.50, Currency: "USD", Status: "pending", UserID: userID},
			{ID: 2, Amount: 200.75, Currency: "EUR", Status: "completed", UserID: userID},
		}

		mockService.On("GetPaymentsByUser", userID).Return(response, nil)

		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("GET", "/users/1/payments", nil)
		ctx.Params = gin.Params{
			{Key: "id", Value: "1"},
		}

		// When
		handler.GetPaymentsByUser(ctx)

		// Then
		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)

		var result map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &result)
		assert.Contains(t, result, "data")
		data := result["data"].([]interface{})
		assert.Len(t, data, 2)
	})

	t.Run("should return bad request for invalid user ID", func(t *testing.T) {
		// Setup
		handler, mockService := setupPaymentHandler()

		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("GET", "/users/invalid/payments", nil)
		ctx.Params = gin.Params{
			{Key: "id", Value: "invalid"},
		}

		// When
		handler.GetPaymentsByUser(ctx)

		// Then
		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("should return internal api error when service fails", func(t *testing.T) {
		// Setup
		handler, mockService := setupPaymentHandler()

		userID := uint(1)
		mockService.On("GetPaymentsByUser", userID).Return(nil, errors.New("service error"))

		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("GET", "/users/1/payments", nil)
		ctx.Params = gin.Params{
			{Key: "id", Value: "1"},
		}

		// When
		handler.GetPaymentsByUser(ctx)

		// Then
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockService.AssertExpectations(t)
	})
}

func TestPaymentHandler_RegisterRoutes(t *testing.T) {
	t.Run("should register all routes correctly", func(t *testing.T) {
		// Setup
		handler, _ := setupPaymentHandler()
		router := gin.New()
		api := router.Group("/api/v1")

		// When
		handler.RegisterRoutes(api)

		// Then
		routes := router.Routes()
		expectedRoutes := []string{
			"POST /api/v1/payments",
			"GET /api/v1/payments",
			"GET /api/v1/payments/:id",
			"PUT /api/v1/payments/:id",
			"DELETE /api/v1/payments/:id",
			"GET /api/v1/users/:id/payments",
		}

		assert.Len(t, routes, len(expectedRoutes))
		for _, expectedRoute := range expectedRoutes {
			found := false
			for _, route := range routes {
				if route.Method+" "+route.Path == expectedRoute {
					found = true
					break
				}
			}
			assert.True(t, found, "Route %s not found", expectedRoute)
		}
	})
}
