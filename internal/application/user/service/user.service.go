package service

import (
	"errors"
	"time"

	"vibe-ddd-golang/internal/application/user/dto"
	"vibe-ddd-golang/internal/application/user/entity"
	"vibe-ddd-golang/internal/application/user/repository"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserService interface {
	CreateUser(req *dto.CreateUserRequest) (*dto.UserResponse, error)
	GetUserByID(id uint) (*dto.UserResponse, error)
	GetUserByEmail(email string) (*dto.UserResponse, error)
	GetUsers(filter *dto.UserFilter) (*dto.UserListResponse, error)
	UpdateUser(id uint, req *dto.UpdateUserRequest) (*dto.UserResponse, error)
	UpdateUserPassword(id uint, req *dto.UpdateUserPasswordRequest) error
	DeleteUser(id uint) error
}

type userService struct {
	repo   repository.UserRepository
	logger *zap.Logger
}

func NewUserService(repo repository.UserRepository, logger *zap.Logger) UserService {
	return &userService{
		repo:   repo,
		logger: logger,
	}
}

func (s *userService) CreateUser(req *dto.CreateUserRequest) (*dto.UserResponse, error) {
	exists, err := s.repo.EmailExists(req.Email)
	if err != nil {
		s.logger.Error("Failed to check email existence", zap.Error(err))
		return nil, err
	}
	if exists {
		return nil, errors.New("email already exists")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error("Failed to hash password", zap.Error(err))
		return nil, err
	}

	user := &entity.User{
		Name:      req.Name,
		Email:     req.Email,
		Password:  string(hashedPassword),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = s.repo.Create(user)
	if err != nil {
		s.logger.Error("Failed to create user", zap.Error(err))
		return nil, err
	}

	return s.entityToResponse(user), nil
}

func (s *userService) GetUserByID(id uint) (*dto.UserResponse, error) {
	user, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	return s.entityToResponse(user), nil
}

func (s *userService) GetUserByEmail(email string) (*dto.UserResponse, error) {
	user, err := s.repo.GetByEmail(email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	return s.entityToResponse(user), nil
}

func (s *userService) GetUsers(filter *dto.UserFilter) (*dto.UserListResponse, error) {
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 10
	}

	users, totalCount, err := s.repo.GetAll(filter)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.UserResponse, 0, len(users))
	for _, user := range users {
		responses = append(responses, *s.entityToResponse(&user))
	}

	return &dto.UserListResponse{
		Data:       responses,
		TotalCount: totalCount,
		Page:       filter.Page,
		PageSize:   filter.PageSize,
	}, nil
}

func (s *userService) UpdateUser(id uint, req *dto.UpdateUserRequest) (*dto.UserResponse, error) {
	user, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	if req.Email != user.Email {
		exists, err := s.repo.EmailExists(req.Email)
		if err != nil {
			s.logger.Error("Failed to check email existence", zap.Error(err))
			return nil, err
		}
		if exists {
			return nil, errors.New("email already exists")
		}
	}

	user.Name = req.Name
	user.Email = req.Email
	user.UpdatedAt = time.Now()

	err = s.repo.Update(user)
	if err != nil {
		s.logger.Error("Failed to update user", zap.Error(err))
		return nil, err
	}

	return s.entityToResponse(user), nil
}

func (s *userService) UpdateUserPassword(id uint, req *dto.UpdateUserPasswordRequest) error {
	user, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("user not found")
		}
		return err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.CurrentPassword))
	if err != nil {
		return errors.New("current password is incorrect")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error("Failed to hash new password", zap.Error(err))
		return err
	}

	user.Password = string(hashedPassword)
	user.UpdatedAt = time.Now()

	return s.repo.Update(user)
}

func (s *userService) DeleteUser(id uint) error {
	_, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("user not found")
		}
		return err
	}

	return s.repo.Delete(id)
}

func (s *userService) entityToResponse(user *entity.User) *dto.UserResponse {
	return &dto.UserResponse{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}
