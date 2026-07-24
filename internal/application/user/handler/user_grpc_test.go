package handler_test

import (
	"context"
	"os"
	"testing"
	"time"

	"vibe-ddd-golang/internal/application/user/dto"
	"vibe-ddd-golang/internal/application/user/handler"
	"vibe-ddd-golang/internal/pkg/response"
	"vibe-ddd-golang/internal/pkg/testutil"
	"vibe-ddd-golang/internal/pkg/validation"
	userv1 "vibe-ddd-golang/internal/server/grpc/proto/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestMain(m *testing.M) {
	if err := validation.Setup(); err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}

func TestUserGRPCServer_RPCs(t *testing.T) {
	ctx := context.Background()
	now := time.Unix(1_700_000_000, 0).UTC()
	user := &dto.UserResponse{
		ID:        7,
		Name:      "John Doe",
		Email:     "john@example.com",
		CreatedAt: now,
		UpdatedAt: now,
	}

	svc := new(testutil.MockUserService)
	svc.On("CreateUser", mock.Anything, mock.MatchedBy(func(req *dto.CreateUserRequest) bool {
		return req.Name == "John Doe" && req.Email == "john@example.com" && req.Password == "password123"
	})).Return(user, nil).Once()
	svc.On("GetUserByID", mock.Anything, uint(7)).Return(user, nil).Once()
	svc.On("GetUsers", mock.Anything, mock.MatchedBy(func(filter *dto.UserFilter) bool {
		return filter.Name == "John" && filter.Email == "john@example.com" && filter.Page == 2 && filter.PageSize == 25
	})).Return(&dto.UserListResponse{
		Data:       []dto.UserResponse{*user},
		TotalCount: 1,
		Page:       2,
		PageSize:   25,
	}, nil).Once()
	svc.On("UpdateUser", mock.Anything, uint(7), mock.MatchedBy(func(req *dto.UpdateUserRequest) bool {
		return req.Name == "Jane Doe" && req.Email == "jane@example.com"
	})).Return(&dto.UserResponse{
		ID:        7,
		Name:      "Jane Doe",
		Email:     "jane@example.com",
		CreatedAt: now,
		UpdatedAt: now,
	}, nil).Once()
	svc.On("DeleteUser", mock.Anything, uint(7)).Return(nil).Once()
	svc.On("UpdateUserPassword", mock.Anything, uint(7), mock.MatchedBy(func(req *dto.UpdateUserPasswordRequest) bool {
		return req.CurrentPassword == "password123" && req.NewPassword == "newpassword123"
	})).Return(nil).Once()

	server := handler.NewUserGRPCServer(svc)

	created, err := server.CreateUser(ctx, &userv1.CreateUserRequest{
		Name:     "John Doe",
		Email:    "john@example.com",
		Password: "password123",
	})
	require.NoError(t, err)
	assert.Equal(t, uint32(7), created.GetUser().GetId())
	assert.Equal(t, "john@example.com", created.GetUser().GetEmail())

	got, err := server.GetUser(ctx, &userv1.GetUserRequest{Id: 7})
	require.NoError(t, err)
	assert.Equal(t, "John Doe", got.GetUser().GetName())

	listed, err := server.ListUsers(ctx, &userv1.ListUsersRequest{
		Name:     "John",
		Email:    "john@example.com",
		Page:     2,
		PageSize: 25,
	})
	require.NoError(t, err)
	assert.Equal(t, int64(1), listed.GetTotal())
	assert.Equal(t, int32(2), listed.GetPage())
	require.Len(t, listed.GetUsers(), 1)

	updated, err := server.UpdateUser(ctx, &userv1.UpdateUserRequest{
		Id:    7,
		Name:  "Jane Doe",
		Email: "jane@example.com",
	})
	require.NoError(t, err)
	assert.Equal(t, "Jane Doe", updated.GetUser().GetName())

	deleted, err := server.DeleteUser(ctx, &userv1.DeleteUserRequest{Id: 7})
	require.NoError(t, err)
	assert.True(t, deleted.GetSuccess())

	password, err := server.UpdateUserPassword(ctx, &userv1.UpdateUserPasswordRequest{
		Id:              7,
		CurrentPassword: "password123",
		NewPassword:     "newpassword123",
	})
	require.NoError(t, err)
	assert.True(t, password.GetSuccess())

	svc.AssertExpectations(t)
}

func TestUserGRPCServer_ErrorMapping(t *testing.T) {
	svc := new(testutil.MockUserService)
	svc.On("GetUserByID", mock.Anything, uint(9)).
		Return(nil, response.NewNotFoundException("user not found")).Once()

	_, err := handler.NewUserGRPCServer(svc).GetUser(context.Background(), &userv1.GetUserRequest{Id: 9})
	require.Error(t, err)
	assert.Equal(t, codes.NotFound, status.Code(err))
	assert.Equal(t, "user not found", status.Convert(err).Message())

	svc.AssertExpectations(t)
}

func TestUserGRPCServer_InvalidPayload(t *testing.T) {
	svc := new(testutil.MockUserService)
	_, err := handler.NewUserGRPCServer(svc).CreateUser(context.Background(), &userv1.CreateUserRequest{
		Name:     "John Doe",
		Email:    "not-email",
		Password: "short",
	})
	require.Error(t, err)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
	svc.AssertNotCalled(t, "CreateUser", mock.Anything, mock.Anything)
}
