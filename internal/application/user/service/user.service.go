package service

import (
	"context"
	"errors"
	"time"

	"vibe-ddd-golang/internal/application/user/dto"
	"vibe-ddd-golang/internal/application/user/entity"
	"vibe-ddd-golang/internal/application/user/repository"
	"vibe-ddd-golang/internal/pkg/response"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserService interface {
	CreateUser(ctx context.Context, req *dto.CreateUserRequest) (*dto.UserResponse, error)
	GetUserByID(ctx context.Context, id uint) (*dto.UserResponse, error)
	GetUserByEmail(ctx context.Context, email string) (*dto.UserResponse, error)
	GetUsers(ctx context.Context, filter *dto.UserFilter) (*dto.UserListResponse, error)
	UpdateUser(ctx context.Context, id uint, req *dto.UpdateUserRequest) (*dto.UserResponse, error)
	UpdateUserPassword(ctx context.Context, id uint, req *dto.UpdateUserPasswordRequest) error
	DeleteUser(ctx context.Context, id uint) error
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

func (s *userService) CreateUser(ctx context.Context, req *dto.CreateUserRequest) (*dto.UserResponse, error) {
	exists, err := s.repo.EmailExists(ctx, req.Email)
	if err != nil {
		return nil, response.NewInternalServerException("failed to check email").WithCause(err)
	}
	if exists {
		return nil, response.NewConflictException("email already exists")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, response.NewInternalServerException("failed to hash password").WithCause(err)
	}

	user := &entity.User{
		Name:      req.Name,
		Email:     req.Email,
		Password:  string(hashedPassword),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, response.NewInternalServerException("failed to create user").WithCause(err)
	}

	return s.entityToResponse(user), nil
}

func (s *userService) GetUserByID(ctx context.Context, id uint) (*dto.UserResponse, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, response.NewNotFoundException("user not found")
		}
		return nil, response.NewInternalServerException("failed to get user").WithCause(err)
	}

	return s.entityToResponse(user), nil
}

func (s *userService) GetUserByEmail(ctx context.Context, email string) (*dto.UserResponse, error) {
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, response.NewNotFoundException("user not found")
		}
		return nil, response.NewInternalServerException("failed to get user").WithCause(err)
	}

	return s.entityToResponse(user), nil
}

func (s *userService) GetUsers(ctx context.Context, filter *dto.UserFilter) (*dto.UserListResponse, error) {
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 10
	}

	users, totalCount, err := s.repo.GetAll(ctx, filter)
	if err != nil {
		return nil, response.NewInternalServerException("failed to list users").WithCause(err)
	}

	responses := make([]dto.UserResponse, 0, len(users))
	for i := range users {
		responses = append(responses, *s.entityToResponse(&users[i]))
	}

	return &dto.UserListResponse{
		Data:       responses,
		TotalCount: totalCount,
		Page:       filter.Page,
		PageSize:   filter.PageSize,
	}, nil
}

func (s *userService) UpdateUser(ctx context.Context, id uint, req *dto.UpdateUserRequest) (*dto.UserResponse, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, response.NewNotFoundException("user not found")
		}
		return nil, response.NewInternalServerException("failed to get user").WithCause(err)
	}

	if req.Email != user.Email {
		exists, err := s.repo.EmailExists(ctx, req.Email)
		if err != nil {
			return nil, response.NewInternalServerException("failed to check email").WithCause(err)
		}
		if exists {
			return nil, response.NewConflictException("email already exists")
		}
	}

	user.Name = req.Name
	user.Email = req.Email
	user.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, user); err != nil {
		return nil, response.NewInternalServerException("failed to update user").WithCause(err)
	}

	return s.entityToResponse(user), nil
}

func (s *userService) UpdateUserPassword(ctx context.Context, id uint, req *dto.UpdateUserPasswordRequest) error {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return response.NewNotFoundException("user not found")
		}
		return response.NewInternalServerException("failed to get user").WithCause(err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.CurrentPassword)); err != nil {
		return response.NewUnauthorizedException("current password is incorrect")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return response.NewInternalServerException("failed to hash new password").WithCause(err)
	}

	user.Password = string(hashedPassword)
	user.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, user); err != nil {
		return response.NewInternalServerException("failed to update password").WithCause(err)
	}
	return nil
}

func (s *userService) DeleteUser(ctx context.Context, id uint) error {
	if _, err := s.repo.GetByID(ctx, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return response.NewNotFoundException("user not found")
		}
		return response.NewInternalServerException("failed to get user").WithCause(err)
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return response.NewInternalServerException("failed to delete user").WithCause(err)
	}
	return nil
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
