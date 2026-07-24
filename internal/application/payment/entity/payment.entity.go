package entity

import (
	"time"

	"gorm.io/gorm"
)

type Payment struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Amount      float64        `json:"amount" gorm:"not null"`
	Currency    string         `json:"currency" gorm:"size:3;not null"`
	Status      PaymentStatus  `json:"status" gorm:"default:pending"`
	Description string         `json:"description" gorm:"size:500"`
	UserID      uint           `json:"user_id" gorm:"not null"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

type PaymentStatus string

const (
	PaymentStatusPending   PaymentStatus = "pending"
	PaymentStatusCompleted PaymentStatus = "completed"
	PaymentStatusFailed    PaymentStatus = "failed"
	PaymentStatusCanceled  PaymentStatus = "canceled"
)

func (p Payment) TableName() string {
	return "payments"
}

func (ps PaymentStatus) String() string {
	return string(ps)
}

func (ps PaymentStatus) IsValid() bool {
	switch ps {
	case PaymentStatusPending, PaymentStatusCompleted, PaymentStatusFailed, PaymentStatusCanceled:
		return true
	default:
		return false
	}
}
