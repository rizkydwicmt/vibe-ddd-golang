package repository

import (
	"context"

	"vibe-ddd-golang/internal/application/payment/dto"
	"vibe-ddd-golang/internal/application/payment/entity"
	"vibe-ddd-golang/internal/common/params"
	database "vibe-ddd-golang/internal/pkg/db"

	"go.uber.org/zap"
)

type PaymentRepository interface {
	Create(ctx context.Context, payment *entity.Payment) error
	GetByID(ctx context.Context, id uint) (*entity.Payment, error)
	GetAll(ctx context.Context, filter *dto.PaymentFilter) ([]entity.Payment, int64, error)
	Update(ctx context.Context, payment *entity.Payment) error
	Delete(ctx context.Context, id uint) error
	GetByUserID(ctx context.Context, userID uint) ([]entity.Payment, error)
}

type paymentRepository struct {
	db     *database.Database
	logger *zap.Logger
}

func NewPaymentRepository(p params.Params, logger *zap.Logger) PaymentRepository {
	return &paymentRepository{
		db:     p.MainDB,
		logger: logger,
	}
}

func (r *paymentRepository) Create(ctx context.Context, payment *entity.Payment) error {
	r.logger.Info("Creating payment", zap.Uint("user_id", payment.UserID))
	return r.db.WithContext(ctx).Create(payment).Error
}

func (r *paymentRepository) GetByID(ctx context.Context, id uint) (*entity.Payment, error) {
	var payment entity.Payment
	if err := r.db.WithContext(ctx).First(&payment, id).Error; err != nil {
		return nil, err
	}
	return &payment, nil
}

func (r *paymentRepository) GetAll(ctx context.Context, filter *dto.PaymentFilter) ([]entity.Payment, int64, error) {
	var payments []entity.Payment
	var totalCount int64

	query := r.db.WithContext(ctx).Model(&entity.Payment{})

	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.Currency != "" {
		query = query.Where("currency = ?", filter.Currency)
	}
	if filter.UserID != 0 {
		query = query.Where("user_id = ?", filter.UserID)
	}

	if err := query.Count(&totalCount).Error; err != nil {
		return nil, 0, err
	}

	if filter.Page > 0 && filter.PageSize > 0 {
		offset := (filter.Page - 1) * filter.PageSize
		query = query.Offset(offset).Limit(filter.PageSize)
	}

	if err := query.Find(&payments).Error; err != nil {
		r.logger.Error("Failed to get payments", zap.Error(err))
		return nil, 0, err
	}

	return payments, totalCount, nil
}

func (r *paymentRepository) Update(ctx context.Context, payment *entity.Payment) error {
	r.logger.Info("Updating payment", zap.Uint("id", payment.ID))
	return r.db.WithContext(ctx).Save(payment).Error
}

func (r *paymentRepository) Delete(ctx context.Context, id uint) error {
	r.logger.Info("Deleting payment", zap.Uint("id", id))
	return r.db.WithContext(ctx).Delete(&entity.Payment{}, id).Error
}

func (r *paymentRepository) GetByUserID(ctx context.Context, userID uint) ([]entity.Payment, error) {
	var payments []entity.Payment
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&payments).Error; err != nil {
		r.logger.Error("Failed to get payments by user ID", zap.Uint("user_id", userID), zap.Error(err))
		return nil, err
	}
	return payments, nil
}
