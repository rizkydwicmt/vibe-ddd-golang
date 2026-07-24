package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"vibe-ddd-golang/internal/application/payment/dto"
	"vibe-ddd-golang/internal/application/payment/handler"
	types "vibe-ddd-golang/internal/common/type"
	"vibe-ddd-golang/internal/pkg/response"
	"vibe-ddd-golang/internal/pkg/testutil"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func setup(svc *testutil.MockPaymentService) *gin.Engine {
	r := testutil.NewTestRouter()
	h := handler.NewPaymentHandler(svc, testutil.NewSilentLogger())
	h.RegisterRoutes(r.Group("/api/v1"))
	return r
}

func do(r *gin.Engine, method, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func decode(t *testing.T, w *httptest.ResponseRecorder) types.ResponseAPI {
	t.Helper()
	var env types.ResponseAPI
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &env))
	return env
}

func TestCreatePayment_Created(t *testing.T) {
	svc := new(testutil.MockPaymentService)
	svc.On("CreatePayment", mock.Anything, mock.AnythingOfType("*dto.CreatePaymentRequest")).
		Return(&dto.PaymentResponse{ID: 1, Status: "pending"}, nil)

	w := do(setup(svc), http.MethodPost, "/api/v1/payments",
		`{"amount":10,"currency":"USD","description":"x","user_id":1}`)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Equal(t, "CREATED", string(decode(t, w).Code))
}

func TestCreatePayment_InvalidPayload(t *testing.T) {
	svc := new(testutil.MockPaymentService) // never called
	w := do(setup(svc), http.MethodPost, "/api/v1/payments", `{"amount":0}`)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, "INVALID_PAYLOAD", string(decode(t, w).Code))
}

func TestGetPayment_NotFound(t *testing.T) {
	svc := new(testutil.MockPaymentService)
	svc.On("GetPaymentByID", mock.Anything, uint(9)).Return(nil, response.NewNotFoundException("payment not found"))

	w := do(setup(svc), http.MethodGet, "/api/v1/payments/9", "")
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Equal(t, "NOT_FOUND", string(decode(t, w).Code))
}
