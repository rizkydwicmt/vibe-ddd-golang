package service_test

import (
	"context"
	"testing"

	"vibe-ddd-golang/internal/application/payment/dto"
	"vibe-ddd-golang/internal/application/payment/entity"
	"vibe-ddd-golang/internal/application/payment/service"
	userDto "vibe-ddd-golang/internal/application/user/dto"
	"vibe-ddd-golang/internal/common/enum"
	"vibe-ddd-golang/internal/pkg/response"
	"vibe-ddd-golang/internal/pkg/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func newService(repo *testutil.MockPaymentRepository, users *testutil.MockUserService) service.PaymentService {
	return service.NewPaymentService(repo, users, testutil.NewSilentLogger())
}

func assertCode(t *testing.T, err error, code enum.ResultCode) {
	t.Helper()
	require.Error(t, err)
	ae, ok := response.As(err)
	require.True(t, ok)
	assert.Equal(t, code, ae.Code)
}

func TestCreatePayment_Success(t *testing.T) {
	repo := new(testutil.MockPaymentRepository)
	users := new(testutil.MockUserService)
	users.On("GetUserByID", mock.Anything, uint(1)).Return(&userDto.UserResponse{ID: 1}, nil)
	repo.On("Create", mock.Anything, mock.AnythingOfType("*entity.Payment")).Return(nil)

	res, err := newService(repo, users).CreatePayment(context.Background(), &dto.CreatePaymentRequest{
		Amount: 10, Currency: "USD", Description: "x", UserID: 1,
	})
	require.NoError(t, err)
	assert.Equal(t, "pending", res.Status)
}

func TestCreatePayment_UserMissing(t *testing.T) {
	repo := new(testutil.MockPaymentRepository)
	users := new(testutil.MockUserService)
	users.On("GetUserByID", mock.Anything, uint(9)).Return(nil, response.NewNotFoundException("user not found"))

	_, err := newService(repo, users).CreatePayment(context.Background(), &dto.CreatePaymentRequest{
		Amount: 10, Currency: "USD", Description: "x", UserID: 9,
	})
	assertCode(t, err, enum.CodeInvalidPayload)
}

func TestGetPaymentByID_NotFound(t *testing.T) {
	repo := new(testutil.MockPaymentRepository)
	repo.On("GetByID", mock.Anything, uint(9)).Return(nil, gorm.ErrRecordNotFound)

	_, err := newService(repo, new(testutil.MockUserService)).GetPaymentByID(context.Background(), 9)
	assertCode(t, err, enum.CodeNotFound)
}

func TestUpdatePayment_InvalidStatus(t *testing.T) {
	repo := new(testutil.MockPaymentRepository)
	repo.On("GetByID", mock.Anything, uint(1)).Return(&entity.Payment{ID: 1}, nil)

	_, err := newService(repo, new(testutil.MockUserService)).UpdatePayment(context.Background(), 1, &dto.UpdatePaymentRequest{
		Status: "bogus",
	})
	assertCode(t, err, enum.CodeInvalidPayload)
}
