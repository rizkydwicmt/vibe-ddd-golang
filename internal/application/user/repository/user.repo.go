package repository

import (
	"context"

	"vibe-ddd-golang/internal/application/user/dto"
	"vibe-ddd-golang/internal/application/user/entity"
	"vibe-ddd-golang/internal/common/params"
	database "vibe-ddd-golang/internal/pkg/db"

	"go.uber.org/zap"
)

type UserRepository interface {
	Create(ctx context.Context, user *entity.User) error
	GetByID(ctx context.Context, id uint) (*entity.User, error)
	GetByEmail(ctx context.Context, email string) (*entity.User, error)
	GetAll(ctx context.Context, filter *dto.UserFilter) ([]entity.User, int64, error)
	Update(ctx context.Context, user *entity.User) error
	Delete(ctx context.Context, id uint) error
	EmailExists(ctx context.Context, email string) (bool, error)
}

type userRepository struct {
	db     *database.Database
	logger *zap.Logger
}

func NewUserRepository(p params.Params, logger *zap.Logger) UserRepository {
	return &userRepository{
		db:     p.MainDB,
		logger: logger,
	}
}

func (r *userRepository) Create(ctx context.Context, user *entity.User) error {
	r.logger.Info("Creating user", zap.String("email", user.Email))
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *userRepository) GetByID(ctx context.Context, id uint) (*entity.User, error) {
	var user entity.User
	if err := r.db.WithContext(ctx).First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	var user entity.User
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetAll(ctx context.Context, filter *dto.UserFilter) ([]entity.User, int64, error) {
	var users []entity.User
	var totalCount int64

	query := r.db.WithContext(ctx).Model(&entity.User{})

	if filter.Name != "" {
		query = query.Where("name LIKE ?", "%"+filter.Name+"%")
	}
	if filter.Email != "" {
		query = query.Where("email LIKE ?", "%"+filter.Email+"%")
	}

	if err := query.Count(&totalCount).Error; err != nil {
		return nil, 0, err
	}

	if filter.Page > 0 && filter.PageSize > 0 {
		offset := (filter.Page - 1) * filter.PageSize
		query = query.Offset(offset).Limit(filter.PageSize)
	}

	if err := query.Find(&users).Error; err != nil {
		r.logger.Error("Failed to get users", zap.Error(err))
		return nil, 0, err
	}

	return users, totalCount, nil
}

func (r *userRepository) Update(ctx context.Context, user *entity.User) error {
	r.logger.Info("Updating user", zap.Uint("id", user.ID))
	return r.db.WithContext(ctx).Save(user).Error
}

func (r *userRepository) Delete(ctx context.Context, id uint) error {
	r.logger.Info("Deleting user", zap.Uint("id", id))
	return r.db.WithContext(ctx).Delete(&entity.User{}, id).Error
}

func (r *userRepository) EmailExists(ctx context.Context, email string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entity.User{}).Where("email = ?", email).Count(&count).Error
	return count > 0, err
}
