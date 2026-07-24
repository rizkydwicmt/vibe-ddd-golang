package testutil

import (
	"time"

	"vibe-ddd-golang/internal/application/payment/dto"
	"vibe-ddd-golang/internal/application/payment/entity"
	userDto "vibe-ddd-golang/internal/application/user/dto"
	userEntity "vibe-ddd-golang/internal/application/user/entity"
)

// User fixtures
func CreateUserFixture() *userEntity.User {
	return &userEntity.User{
		ID:        1,
		Name:      "John Doe",
		Email:     "john@example.com",
		Password:  "$2a$10$example.hashed.password",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func CreateUserRequestFixture() *userDto.CreateUserRequest {
	return &userDto.CreateUserRequest{
		Name:     "John Doe",
		Email:    "john@example.com",
		Password: "password123",
	}
}

func CreateUpdateUserRequestFixture() *userDto.UpdateUserRequest {
	return &userDto.UpdateUserRequest{
		Name:  "John Updated",
		Email: "john.updated@example.com",
	}
}

// Payment fixtures
func CreatePaymentFixture() *entity.Payment {
	return &entity.Payment{
		ID:          1,
		Amount:      100.50,
		Currency:    "USD",
		Status:      entity.PaymentStatusPending,
		Description: "Test payment",
		UserID:      1,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

func CreatePaymentRequestFixture() *dto.CreatePaymentRequest {
	return &dto.CreatePaymentRequest{
		Amount:      100.50,
		Currency:    "USD",
		Description: "Test payment",
		UserID:      1,
	}
}

func CreateUpdatePaymentRequestFixture() *dto.UpdatePaymentRequest {
	return &dto.UpdatePaymentRequest{
		Status:      entity.PaymentStatusCompleted.String(),
		Description: "Payment completed",
	}
}

func CreatePaymentFilterFixture() *dto.PaymentFilter {
	return &dto.PaymentFilter{
		Status:   "pending",
		Currency: "USD",
		UserID:   1,
		Page:     1,
		PageSize: 10,
	}
}
