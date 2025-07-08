package repository

import (
	"vibe-ddd-golang/internal/application/payment/dto"
	"vibe-ddd-golang/internal/application/payment/entity"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type PaymentRepository interface {
	Create(payment *entity.Payment) error
	GetByID(id uint) (*entity.Payment, error)
	GetAll(filter *dto.PaymentFilter) ([]entity.Payment, int64, error)
	Update(payment *entity.Payment) error
	Delete(id uint) error
	GetByUserID(userID uint) ([]entity.Payment, error)
}

type paymentRepository struct {
	db     *gorm.DB
	logger *zap.Logger
}

func NewPaymentRepository(db *gorm.DB, logger *zap.Logger) PaymentRepository {
	return &paymentRepository{
		db:     db,
		logger: logger,
	}
}

func (r *paymentRepository) Create(payment *entity.Payment) error {
	r.logger.Info("Creating payment", zap.Uint("user_id", payment.UserID))
	return r.db.Create(payment).Error
}

func (r *paymentRepository) GetByID(id uint) (*entity.Payment, error) {
	var payment entity.Payment
	err := r.db.First(&payment, id).Error
	if err != nil {
		r.logger.Error("Failed to get payment by ID", zap.Uint("id", id), zap.Error(err))
		return nil, err
	}
	return &payment, nil
}

func (r *paymentRepository) GetAll(filter *dto.PaymentFilter) ([]entity.Payment, int64, error) {
	var payments []entity.Payment
	var totalCount int64

	query := r.db.Model(&entity.Payment{})

	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.Currency != "" {
		query = query.Where("currency = ?", filter.Currency)
	}
	if filter.UserID != 0 {
		query = query.Where("user_id = ?", filter.UserID)
	}

	query.Count(&totalCount)

	if filter.Page > 0 && filter.PageSize > 0 {
		offset := (filter.Page - 1) * filter.PageSize
		query = query.Offset(offset).Limit(filter.PageSize)
	}

	err := query.Find(&payments).Error
	if err != nil {
		r.logger.Error("Failed to get payments", zap.Error(err))
		return nil, 0, err
	}

	return payments, totalCount, nil
}

func (r *paymentRepository) Update(payment *entity.Payment) error {
	r.logger.Info("Updating payment", zap.Uint("id", payment.ID))
	return r.db.Save(payment).Error
}

func (r *paymentRepository) Delete(id uint) error {
	r.logger.Info("Deleting payment", zap.Uint("id", id))
	return r.db.Delete(&entity.Payment{}, id).Error
}

func (r *paymentRepository) GetByUserID(userID uint) ([]entity.Payment, error) {
	var payments []entity.Payment
	err := r.db.Where("user_id = ?", userID).Find(&payments).Error
	if err != nil {
		r.logger.Error("Failed to get payments by user ID", zap.Uint("user_id", userID), zap.Error(err))
		return nil, err
	}
	return payments, nil
}
