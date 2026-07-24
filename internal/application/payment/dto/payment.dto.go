package dto

import (
	"time"
)

type CreatePaymentRequest struct {
	Amount      float64 `json:"amount" validate:"required,gt=0"`
	Currency    string  `json:"currency" validate:"required,len=3"`
	Description string  `json:"description" validate:"required"`
	UserID      uint    `json:"user_id" validate:"required"`
}

type UpdatePaymentRequest struct {
	Status      string `json:"status" validate:"required,oneof=pending completed failed canceled"`
	Description string `json:"description"`
}

type PaymentResponse struct {
	ID          uint      `json:"id"`
	Amount      float64   `json:"amount"`
	Currency    string    `json:"currency"`
	Status      string    `json:"status"`
	Description string    `json:"description"`
	UserID      uint      `json:"user_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type PaymentListResponse struct {
	Data       []PaymentResponse `json:"data"`
	TotalCount int64             `json:"total_count"`
	Page       int               `json:"page"`
	PageSize   int               `json:"page_size"`
}

type PaymentFilter struct {
	Status   string `form:"status"`
	Currency string `form:"currency"`
	UserID   uint   `form:"user_id"`
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
}
