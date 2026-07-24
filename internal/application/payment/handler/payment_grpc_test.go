package handler_test

import (
	"context"
	"os"
	"testing"
	"time"

	"vibe-ddd-golang/internal/application/payment/dto"
	"vibe-ddd-golang/internal/application/payment/handler"
	"vibe-ddd-golang/internal/pkg/response"
	"vibe-ddd-golang/internal/pkg/testutil"
	"vibe-ddd-golang/internal/pkg/validation"
	paymentv1 "vibe-ddd-golang/internal/server/grpc/proto/payment"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestMain(m *testing.M) {
	if err := validation.Setup(); err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}

func TestPaymentGRPCServer_RPCs(t *testing.T) {
	ctx := context.Background()
	now := time.Unix(1_700_000_000, 0).UTC()
	payment := &dto.PaymentResponse{
		ID:          12,
		Amount:      100.5,
		Currency:    "USD",
		Status:      "pending",
		Description: "invoice",
		UserID:      7,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	completed := *payment
	completed.Status = "completed"
	completed.Description = "paid"

	svc := new(testutil.MockPaymentService)
	svc.On("CreatePayment", mock.Anything, mock.MatchedBy(func(req *dto.CreatePaymentRequest) bool {
		return req.Amount == 100.5 && req.Currency == "USD" && req.Description == "invoice" && req.UserID == 7
	})).Return(payment, nil).Once()
	svc.On("GetPaymentByID", mock.Anything, uint(12)).Return(payment, nil).Once()
	svc.On("GetPayments", mock.Anything, mock.MatchedBy(func(filter *dto.PaymentFilter) bool {
		return filter.Status == "pending" && filter.Currency == "USD" && filter.UserID == 7 && filter.Page == 1 && filter.PageSize == 10
	})).Return(&dto.PaymentListResponse{
		Data:       []dto.PaymentResponse{*payment},
		TotalCount: 1,
		Page:       1,
		PageSize:   10,
	}, nil).Once()
	svc.On("UpdatePayment", mock.Anything, uint(12), mock.MatchedBy(func(req *dto.UpdatePaymentRequest) bool {
		return req.Status == "completed" && req.Description == "paid"
	})).Return(&completed, nil).Once()
	svc.On("DeletePayment", mock.Anything, uint(12)).Return(nil).Once()
	svc.On("GetPaymentsByUser", mock.Anything, uint(7)).
		Return([]dto.PaymentResponse{*payment, completed}, nil).Once()

	server := handler.NewPaymentGRPCServer(svc)

	created, err := server.CreatePayment(ctx, &paymentv1.CreatePaymentRequest{
		Amount:      100.5,
		Currency:    "USD",
		Description: "invoice",
		UserId:      7,
	})
	require.NoError(t, err)
	assert.Equal(t, uint32(12), created.GetPayment().GetId())
	assert.Equal(t, paymentv1.PaymentStatus_PAYMENT_STATUS_PENDING, created.GetPayment().GetStatus())

	got, err := server.GetPayment(ctx, &paymentv1.GetPaymentRequest{Id: 12})
	require.NoError(t, err)
	assert.Equal(t, "invoice", got.GetPayment().GetDescription())

	listed, err := server.ListPayments(ctx, &paymentv1.ListPaymentsRequest{
		Status:   paymentv1.PaymentStatus_PAYMENT_STATUS_PENDING,
		Currency: "USD",
		UserId:   7,
		Page:     1,
		PageSize: 10,
	})
	require.NoError(t, err)
	assert.Equal(t, int64(1), listed.GetTotal())
	require.Len(t, listed.GetPayments(), 1)

	updated, err := server.UpdatePayment(ctx, &paymentv1.UpdatePaymentRequest{
		Id:          12,
		Status:      paymentv1.PaymentStatus_PAYMENT_STATUS_COMPLETED,
		Description: "paid",
	})
	require.NoError(t, err)
	assert.Equal(t, paymentv1.PaymentStatus_PAYMENT_STATUS_COMPLETED, updated.GetPayment().GetStatus())

	deleted, err := server.DeletePayment(ctx, &paymentv1.DeletePaymentRequest{Id: 12})
	require.NoError(t, err)
	assert.True(t, deleted.GetSuccess())

	userPayments, err := server.GetUserPayments(ctx, &paymentv1.GetUserPaymentsRequest{UserId: 7})
	require.NoError(t, err)
	require.Len(t, userPayments.GetPayments(), 2)

	svc.AssertExpectations(t)
}

func TestPaymentGRPCServer_ErrorMapping(t *testing.T) {
	svc := new(testutil.MockPaymentService)
	svc.On("GetPaymentByID", mock.Anything, uint(9)).
		Return(nil, response.NewNotFoundException("payment not found")).Once()

	_, err := handler.NewPaymentGRPCServer(svc).GetPayment(context.Background(), &paymentv1.GetPaymentRequest{Id: 9})
	require.Error(t, err)
	assert.Equal(t, codes.NotFound, status.Code(err))
	assert.Equal(t, "payment not found", status.Convert(err).Message())

	svc.AssertExpectations(t)
}

func TestPaymentGRPCServer_InvalidPayload(t *testing.T) {
	svc := new(testutil.MockPaymentService)
	_, err := handler.NewPaymentGRPCServer(svc).UpdatePayment(context.Background(), &paymentv1.UpdatePaymentRequest{
		Id:          12,
		Status:      paymentv1.PaymentStatus_PAYMENT_STATUS_UNSPECIFIED,
		Description: "paid",
	})
	require.Error(t, err)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
	svc.AssertNotCalled(t, "UpdatePayment", mock.Anything, mock.Anything, mock.Anything)
}
