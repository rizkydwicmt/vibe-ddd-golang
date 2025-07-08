package repository

import (
	"vibe-ddd-golang/internal/application/user/dto"
	"vibe-ddd-golang/internal/application/user/entity"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type UserRepository interface {
	Create(user *entity.User) error
	GetByID(id uint) (*entity.User, error)
	GetByEmail(email string) (*entity.User, error)
	GetAll(filter *dto.UserFilter) ([]entity.User, int64, error)
	Update(user *entity.User) error
	Delete(id uint) error
	EmailExists(email string) (bool, error)
}

type userRepository struct {
	db     *gorm.DB
	logger *zap.Logger
}

func NewUserRepository(db *gorm.DB, logger *zap.Logger) UserRepository {
	return &userRepository{
		db:     db,
		logger: logger,
	}
}

func (r *userRepository) Create(user *entity.User) error {
	r.logger.Info("Creating user", zap.String("email", user.Email))
	return r.db.Create(user).Error
}

func (r *userRepository) GetByID(id uint) (*entity.User, error) {
	var user entity.User
	err := r.db.First(&user, id).Error
	if err != nil {
		r.logger.Error("Failed to get user by ID", zap.Uint("id", id), zap.Error(err))
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetByEmail(email string) (*entity.User, error) {
	var user entity.User
	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		r.logger.Error("Failed to get user by email", zap.String("email", email), zap.Error(err))
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetAll(filter *dto.UserFilter) ([]entity.User, int64, error) {
	var users []entity.User
	var totalCount int64

	query := r.db.Model(&entity.User{})

	if filter.Name != "" {
		query = query.Where("name LIKE ?", "%"+filter.Name+"%")
	}
	if filter.Email != "" {
		query = query.Where("email LIKE ?", "%"+filter.Email+"%")
	}

	query.Count(&totalCount)

	if filter.Page > 0 && filter.PageSize > 0 {
		offset := (filter.Page - 1) * filter.PageSize
		query = query.Offset(offset).Limit(filter.PageSize)
	}

	err := query.Find(&users).Error
	if err != nil {
		r.logger.Error("Failed to get users", zap.Error(err))
		return nil, 0, err
	}

	return users, totalCount, nil
}

func (r *userRepository) Update(user *entity.User) error {
	r.logger.Info("Updating user", zap.Uint("id", user.ID))
	return r.db.Save(user).Error
}

func (r *userRepository) Delete(id uint) error {
	r.logger.Info("Deleting user", zap.Uint("id", id))
	return r.db.Delete(&entity.User{}, id).Error
}

func (r *userRepository) EmailExists(email string) (bool, error) {
	var count int64
	err := r.db.Model(&entity.User{}).Where("email = ?", email).Count(&count).Error
	return count > 0, err
}
