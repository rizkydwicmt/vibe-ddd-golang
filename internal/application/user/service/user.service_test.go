package service

import (
	"errors"
	"testing"
	"time"

	"vibe-ddd-golang/internal/application/user/dto"
	"vibe-ddd-golang/internal/application/user/entity"
	"vibe-ddd-golang/internal/pkg/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func TestUserService_CreateUser(t *testing.T) {
	t.Run("should create user successfully", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockUserRepository{}
		logger := testutil.NewSilentLogger()
		service := NewUserService(mockRepo, logger)

		req := testutil.CreateUserRequestFixture()

		// Mock expectations
		mockRepo.On("EmailExists", req.Email).Return(false, nil)
		mockRepo.On("Create", mock.AnythingOfType("*entity.User")).Return(nil).Run(func(args mock.Arguments) {
			user := args.Get(0).(*entity.User)
			user.ID = 1
		})

		// When
		response, err := service.CreateUser(req)

		// Then
		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, uint(1), response.ID)
		assert.Equal(t, req.Name, response.Name)
		assert.Equal(t, req.Email, response.Email)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error when email already exists", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockUserRepository{}
		logger := testutil.NewSilentLogger()
		service := NewUserService(mockRepo, logger)

		req := testutil.CreateUserRequestFixture()

		// Mock expectations
		mockRepo.On("EmailExists", req.Email).Return(true, nil)

		// When
		response, err := service.CreateUser(req)

		// Then
		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "email already exists")
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error when email check fails", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockUserRepository{}
		logger := testutil.NewSilentLogger()
		service := NewUserService(mockRepo, logger)

		req := testutil.CreateUserRequestFixture()

		// Mock expectations
		mockRepo.On("EmailExists", req.Email).Return(false, errors.New("database error"))

		// When
		response, err := service.CreateUser(req)

		// Then
		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "database error")
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error when user creation fails", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockUserRepository{}
		logger := testutil.NewSilentLogger()
		service := NewUserService(mockRepo, logger)

		req := testutil.CreateUserRequestFixture()

		// Mock expectations
		mockRepo.On("EmailExists", req.Email).Return(false, nil)
		mockRepo.On("Create", mock.AnythingOfType("*entity.User")).Return(errors.New("create failed"))

		// When
		response, err := service.CreateUser(req)

		// Then
		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "create failed")
		mockRepo.AssertExpectations(t)
	})
}

func TestUserService_GetUserByID(t *testing.T) {
	t.Run("should get user by ID successfully", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockUserRepository{}
		logger := testutil.NewSilentLogger()
		service := NewUserService(mockRepo, logger)

		userID := uint(1)
		user := testutil.CreateUserFixture()
		user.ID = userID

		// Mock expectations
		mockRepo.On("GetByID", userID).Return(user, nil)

		// When
		response, err := service.GetUserByID(userID)

		// Then
		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, userID, response.ID)
		assert.Equal(t, user.Name, response.Name)
		assert.Equal(t, user.Email, response.Email)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error when user not found", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockUserRepository{}
		logger := testutil.NewSilentLogger()
		service := NewUserService(mockRepo, logger)

		userID := uint(999)

		// Mock expectations
		mockRepo.On("GetByID", userID).Return(nil, gorm.ErrRecordNotFound)

		// When
		response, err := service.GetUserByID(userID)

		// Then
		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "user not found")
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error when repository fails", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockUserRepository{}
		logger := testutil.NewSilentLogger()
		service := NewUserService(mockRepo, logger)

		userID := uint(1)

		// Mock expectations
		mockRepo.On("GetByID", userID).Return(nil, errors.New("database error"))

		// When
		response, err := service.GetUserByID(userID)

		// Then
		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "database error")
		mockRepo.AssertExpectations(t)
	})
}

func TestUserService_GetUserByEmail(t *testing.T) {
	t.Run("should get user by email successfully", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockUserRepository{}
		logger := testutil.NewSilentLogger()
		service := NewUserService(mockRepo, logger)

		email := "test@example.com"
		user := testutil.CreateUserFixture()
		user.Email = email

		// Mock expectations
		mockRepo.On("GetByEmail", email).Return(user, nil)

		// When
		response, err := service.GetUserByEmail(email)

		// Then
		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, email, response.Email)
		assert.Equal(t, user.Name, response.Name)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error when user not found", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockUserRepository{}
		logger := testutil.NewSilentLogger()
		service := NewUserService(mockRepo, logger)

		email := "nonexistent@example.com"

		// Mock expectations
		mockRepo.On("GetByEmail", email).Return(nil, gorm.ErrRecordNotFound)

		// When
		response, err := service.GetUserByEmail(email)

		// Then
		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "user not found")
		mockRepo.AssertExpectations(t)
	})
}

func TestUserService_GetUsers(t *testing.T) {
	t.Run("should get users with pagination successfully", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockUserRepository{}
		logger := testutil.NewSilentLogger()
		service := NewUserService(mockRepo, logger)

		filter := &dto.UserFilter{
			Page:     1,
			PageSize: 10,
		}

		users := []entity.User{
			*testutil.CreateUserFixture(),
			*testutil.CreateUserFixture(),
		}
		users[0].ID = 1
		users[1].ID = 2
		users[1].Email = "user2@example.com"

		// Mock expectations
		mockRepo.On("GetAll", filter).Return(users, int64(2), nil)

		// When
		response, err := service.GetUsers(filter)

		// Then
		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Len(t, response.Data, 2)
		assert.Equal(t, int64(2), response.TotalCount)
		assert.Equal(t, 1, response.Page)
		assert.Equal(t, 10, response.PageSize)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should set default pagination values", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockUserRepository{}
		logger := testutil.NewSilentLogger()
		service := NewUserService(mockRepo, logger)

		filter := &dto.UserFilter{
			Page:     0,
			PageSize: 0,
		}

		expectedFilter := &dto.UserFilter{
			Page:     1,
			PageSize: 10,
		}

		// Mock expectations
		mockRepo.On("GetAll", expectedFilter).Return([]entity.User{}, int64(0), nil)

		// When
		response, err := service.GetUsers(filter)

		// Then
		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, 1, response.Page)
		assert.Equal(t, 10, response.PageSize)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error when repository fails", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockUserRepository{}
		logger := testutil.NewSilentLogger()
		service := NewUserService(mockRepo, logger)

		filter := &dto.UserFilter{
			Page:     1,
			PageSize: 10,
		}

		// Mock expectations
		mockRepo.On("GetAll", filter).Return(nil, int64(0), errors.New("database error"))

		// When
		response, err := service.GetUsers(filter)

		// Then
		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "database error")
		mockRepo.AssertExpectations(t)
	})
}

func TestUserService_UpdateUser(t *testing.T) {
	t.Run("should update user successfully", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockUserRepository{}
		logger := testutil.NewSilentLogger()
		service := NewUserService(mockRepo, logger)

		userID := uint(1)
		existingUser := testutil.CreateUserFixture()
		existingUser.ID = userID
		existingUser.Email = "old@example.com"

		req := testutil.CreateUpdateUserRequestFixture()
		req.Email = "new@example.com"

		// Mock expectations
		mockRepo.On("GetByID", userID).Return(existingUser, nil)
		mockRepo.On("EmailExists", req.Email).Return(false, nil)
		mockRepo.On("Update", mock.AnythingOfType("*entity.User")).Return(nil)

		// When
		response, err := service.UpdateUser(userID, req)

		// Then
		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, userID, response.ID)
		assert.Equal(t, req.Name, response.Name)
		assert.Equal(t, req.Email, response.Email)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error when user not found", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockUserRepository{}
		logger := testutil.NewSilentLogger()
		service := NewUserService(mockRepo, logger)

		userID := uint(999)
		req := testutil.CreateUpdateUserRequestFixture()

		// Mock expectations
		mockRepo.On("GetByID", userID).Return(nil, gorm.ErrRecordNotFound)

		// When
		response, err := service.UpdateUser(userID, req)

		// Then
		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "user not found")
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error when new email already exists", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockUserRepository{}
		logger := testutil.NewSilentLogger()
		service := NewUserService(mockRepo, logger)

		userID := uint(1)
		existingUser := testutil.CreateUserFixture()
		existingUser.ID = userID
		existingUser.Email = "old@example.com"

		req := testutil.CreateUpdateUserRequestFixture()
		req.Email = "exists@example.com"

		// Mock expectations
		mockRepo.On("GetByID", userID).Return(existingUser, nil)
		mockRepo.On("EmailExists", req.Email).Return(true, nil)

		// When
		response, err := service.UpdateUser(userID, req)

		// Then
		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "email already exists")
		mockRepo.AssertExpectations(t)
	})

	t.Run("should not check email existence when email unchanged", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockUserRepository{}
		logger := testutil.NewSilentLogger()
		service := NewUserService(mockRepo, logger)

		userID := uint(1)
		existingUser := testutil.CreateUserFixture()
		existingUser.ID = userID
		existingUser.Email = "same@example.com"

		req := testutil.CreateUpdateUserRequestFixture()
		req.Email = "same@example.com"

		// Mock expectations
		mockRepo.On("GetByID", userID).Return(existingUser, nil)
		mockRepo.On("Update", mock.AnythingOfType("*entity.User")).Return(nil)

		// When
		response, err := service.UpdateUser(userID, req)

		// Then
		assert.NoError(t, err)
		assert.NotNil(t, response)
		mockRepo.AssertExpectations(t)
		mockRepo.AssertNotCalled(t, "EmailExists")
	})
}

func TestUserService_UpdateUserPassword(t *testing.T) {
	t.Run("should update user password successfully", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockUserRepository{}
		logger := testutil.NewSilentLogger()
		service := NewUserService(mockRepo, logger)

		userID := uint(1)
		currentPassword := "currentpassword"
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(currentPassword), bcrypt.DefaultCost)

		existingUser := testutil.CreateUserFixture()
		existingUser.ID = userID
		existingUser.Password = string(hashedPassword)

		req := &dto.UpdateUserPasswordRequest{
			CurrentPassword: currentPassword,
			NewPassword:     "newpassword123",
		}

		// Mock expectations
		mockRepo.On("GetByID", userID).Return(existingUser, nil)
		mockRepo.On("Update", mock.AnythingOfType("*entity.User")).Return(nil)

		// When
		err := service.UpdateUserPassword(userID, req)

		// Then
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error when user not found", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockUserRepository{}
		logger := testutil.NewSilentLogger()
		service := NewUserService(mockRepo, logger)

		userID := uint(999)
		req := &dto.UpdateUserPasswordRequest{
			CurrentPassword: "password",
			NewPassword:     "newpassword",
		}

		// Mock expectations
		mockRepo.On("GetByID", userID).Return(nil, gorm.ErrRecordNotFound)

		// When
		err := service.UpdateUserPassword(userID, req)

		// Then
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user not found")
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error when current password is incorrect", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockUserRepository{}
		logger := testutil.NewSilentLogger()
		service := NewUserService(mockRepo, logger)

		userID := uint(1)
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("correctpassword"), bcrypt.DefaultCost)

		existingUser := testutil.CreateUserFixture()
		existingUser.ID = userID
		existingUser.Password = string(hashedPassword)

		req := &dto.UpdateUserPasswordRequest{
			CurrentPassword: "wrongpassword",
			NewPassword:     "newpassword123",
		}

		// Mock expectations
		mockRepo.On("GetByID", userID).Return(existingUser, nil)

		// When
		err := service.UpdateUserPassword(userID, req)

		// Then
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "current password is incorrect")
		mockRepo.AssertExpectations(t)
	})
}

func TestUserService_DeleteUser(t *testing.T) {
	t.Run("should delete user successfully", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockUserRepository{}
		logger := testutil.NewSilentLogger()
		service := NewUserService(mockRepo, logger)

		userID := uint(1)
		user := testutil.CreateUserFixture()
		user.ID = userID

		// Mock expectations
		mockRepo.On("GetByID", userID).Return(user, nil)
		mockRepo.On("Delete", userID).Return(nil)

		// When
		err := service.DeleteUser(userID)

		// Then
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error when user not found", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockUserRepository{}
		logger := testutil.NewSilentLogger()
		service := NewUserService(mockRepo, logger)

		userID := uint(999)

		// Mock expectations
		mockRepo.On("GetByID", userID).Return(nil, gorm.ErrRecordNotFound)

		// When
		err := service.DeleteUser(userID)

		// Then
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user not found")
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error when delete fails", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockUserRepository{}
		logger := testutil.NewSilentLogger()
		service := NewUserService(mockRepo, logger)

		userID := uint(1)
		user := testutil.CreateUserFixture()
		user.ID = userID

		// Mock expectations
		mockRepo.On("GetByID", userID).Return(user, nil)
		mockRepo.On("Delete", userID).Return(errors.New("delete failed"))

		// When
		err := service.DeleteUser(userID)

		// Then
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "delete failed")
		mockRepo.AssertExpectations(t)
	})
}

func TestUserService_entityToResponse(t *testing.T) {
	t.Run("should convert entity to response correctly", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockUserRepository{}
		logger := testutil.NewSilentLogger()
		service := NewUserService(mockRepo, logger).(*userService)

		user := testutil.CreateUserFixture()
		user.ID = 1
		user.Name = "Test User"
		user.Email = "test@example.com"
		user.CreatedAt = time.Now()
		user.UpdatedAt = time.Now()

		// When
		response := service.entityToResponse(user)

		// Then
		assert.Equal(t, user.ID, response.ID)
		assert.Equal(t, user.Name, response.Name)
		assert.Equal(t, user.Email, response.Email)
		assert.Equal(t, user.CreatedAt, response.CreatedAt)
		assert.Equal(t, user.UpdatedAt, response.UpdatedAt)
		// Password should not be included in response (UserResponse doesn't have Password field)
	})
}
