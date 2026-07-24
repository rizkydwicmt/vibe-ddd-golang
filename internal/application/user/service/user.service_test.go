package service_test

import (
	"context"
	"testing"

	"vibe-ddd-golang/internal/application/user/dto"
	"vibe-ddd-golang/internal/application/user/entity"
	"vibe-ddd-golang/internal/application/user/service"
	"vibe-ddd-golang/internal/common/enum"
	"vibe-ddd-golang/internal/pkg/response"
	"vibe-ddd-golang/internal/pkg/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func newService(repo *testutil.MockUserRepository) service.UserService {
	return service.NewUserService(repo, testutil.NewSilentLogger())
}

// assertCode unwraps an AppError and asserts its result code.
func assertCode(t *testing.T, err error, code enum.ResultCode) {
	t.Helper()
	require.Error(t, err)
	ae, ok := response.As(err)
	require.True(t, ok)
	assert.Equal(t, code, ae.Code)
}

func TestCreateUser_Success(t *testing.T) {
	repo := new(testutil.MockUserRepository)
	repo.On("EmailExists", mock.Anything, "john@example.com").Return(false, nil)
	repo.On("Create", mock.Anything, mock.AnythingOfType("*entity.User")).Return(nil)

	res, err := newService(repo).CreateUser(context.Background(), &dto.CreateUserRequest{
		Name: "John", Email: "john@example.com", Password: "password123",
	})
	require.NoError(t, err)
	assert.Equal(t, "john@example.com", res.Email)
	repo.AssertExpectations(t)
}

func TestCreateUser_EmailConflict(t *testing.T) {
	repo := new(testutil.MockUserRepository)
	repo.On("EmailExists", mock.Anything, "dup@example.com").Return(true, nil)

	_, err := newService(repo).CreateUser(context.Background(), &dto.CreateUserRequest{
		Name: "Dup", Email: "dup@example.com", Password: "password123",
	})
	assertCode(t, err, enum.CodeConflict)
}

func TestGetUserByID_NotFound(t *testing.T) {
	repo := new(testutil.MockUserRepository)
	repo.On("GetByID", mock.Anything, uint(9)).Return(nil, gorm.ErrRecordNotFound)

	_, err := newService(repo).GetUserByID(context.Background(), 9)
	assertCode(t, err, enum.CodeNotFound)
}

func TestGetUserByID_Success(t *testing.T) {
	repo := new(testutil.MockUserRepository)
	repo.On("GetByID", mock.Anything, uint(1)).Return(&entity.User{ID: 1, Email: "a@x.com"}, nil)

	res, err := newService(repo).GetUserByID(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, uint(1), res.ID)
}

func TestUpdateUserPassword_WrongCurrent(t *testing.T) {
	repo := new(testutil.MockUserRepository)
	// stored password hash of "correct"
	repo.On("GetByID", mock.Anything, uint(1)).Return(&entity.User{
		ID: 1, Password: "$2a$10$7EqJtq98hPqEX7fNZaFWoOhi5.raNQ1i7uJ0J2u9vXOa2A0Y0m9lS",
	}, nil)

	err := newService(repo).UpdateUserPassword(context.Background(), 1, &dto.UpdateUserPasswordRequest{
		CurrentPassword: "wrong", NewPassword: "newpassword123",
	})
	assertCode(t, err, enum.CodeUnauthorized)
}

func TestGetUsers_DefaultsPagination(t *testing.T) {
	repo := new(testutil.MockUserRepository)
	repo.On("GetAll", mock.Anything, mock.AnythingOfType("*dto.UserFilter")).
		Return([]entity.User{{ID: 1, Email: "a@x.com"}}, int64(1), nil)

	res, err := newService(repo).GetUsers(context.Background(), &dto.UserFilter{})
	require.NoError(t, err)
	assert.Equal(t, 1, res.Page)
	assert.Equal(t, 10, res.PageSize)
	assert.Len(t, res.Data, 1)
}
