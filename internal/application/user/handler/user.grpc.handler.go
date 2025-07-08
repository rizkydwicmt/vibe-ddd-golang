package handler

import (
	"context"

	"vibe-ddd-golang/api/proto/user"
	"vibe-ddd-golang/internal/application/user/dto"
	"vibe-ddd-golang/internal/application/user/service"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type UserGrpcHandler struct {
	user.UnimplementedUserServiceServer
	userService service.UserService
	logger      *zap.Logger
}

func NewUserGrpcHandler(userService service.UserService, logger *zap.Logger) *UserGrpcHandler {
	return &UserGrpcHandler{
		userService: userService,
		logger:      logger,
	}
}

func (h *UserGrpcHandler) CreateUser(
	ctx context.Context,
	req *user.CreateUserRequest,
) (*user.CreateUserResponse, error) {
	createReq := &dto.CreateUserRequest{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
	}

	userResponse, err := h.userService.CreateUser(createReq)
	if err != nil {
		h.logger.Error("Failed to create user via gRPC", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to create user: %v", err)
	}

	return &user.CreateUserResponse{
		User: h.toProtoUser(userResponse),
	}, nil
}

func (h *UserGrpcHandler) GetUser(ctx context.Context, req *user.GetUserRequest) (*user.GetUserResponse, error) {
	userResponse, err := h.userService.GetUserByID(uint(req.Id))
	if err != nil {
		h.logger.Error("Failed to get user via gRPC", zap.Uint32("id", req.Id), zap.Error(err))
		return nil, status.Errorf(codes.NotFound, "user not found: %v", err)
	}

	return &user.GetUserResponse{
		User: h.toProtoUser(userResponse),
	}, nil
}

func (h *UserGrpcHandler) ListUsers(ctx context.Context, req *user.ListUsersRequest) (*user.ListUsersResponse, error) {
	page := int(req.Page)
	pageSize := int(req.PageSize)

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	filter := &dto.UserFilter{
		Page:     page,
		PageSize: pageSize,
	}

	listResponse, err := h.userService.GetUsers(filter)
	if err != nil {
		h.logger.Error("Failed to list users via gRPC", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to list users: %v", err)
	}

	protoUsers := make([]*user.User, len(listResponse.Data))
	for i, u := range listResponse.Data {
		protoUsers[i] = h.toProtoUser(&u)
	}

	return &user.ListUsersResponse{
		Users:    protoUsers,
		Total:    listResponse.TotalCount,
		Page:     int32(listResponse.Page),
		PageSize: int32(listResponse.PageSize),
	}, nil
}

func (h *UserGrpcHandler) UpdateUser(
	ctx context.Context,
	req *user.UpdateUserRequest,
) (*user.UpdateUserResponse, error) {
	updateReq := &dto.UpdateUserRequest{
		Name:  req.Name,
		Email: req.Email,
	}

	userResponse, err := h.userService.UpdateUser(uint(req.Id), updateReq)
	if err != nil {
		h.logger.Error("Failed to update user via gRPC", zap.Uint32("id", req.Id), zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to update user: %v", err)
	}

	return &user.UpdateUserResponse{
		User: h.toProtoUser(userResponse),
	}, nil
}

func (h *UserGrpcHandler) DeleteUser(
	ctx context.Context,
	req *user.DeleteUserRequest,
) (*user.DeleteUserResponse, error) {
	err := h.userService.DeleteUser(uint(req.Id))
	if err != nil {
		h.logger.Error("Failed to delete user via gRPC", zap.Uint32("id", req.Id), zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to delete user: %v", err)
	}

	return &user.DeleteUserResponse{
		Success: true,
	}, nil
}

func (h *UserGrpcHandler) UpdateUserPassword(
	ctx context.Context,
	req *user.UpdateUserPasswordRequest,
) (*user.UpdateUserPasswordResponse, error) {
	updateReq := &dto.UpdateUserPasswordRequest{
		CurrentPassword: req.OldPassword,
		NewPassword:     req.NewPassword,
	}

	err := h.userService.UpdateUserPassword(uint(req.Id), updateReq)
	if err != nil {
		h.logger.Error("Failed to update user password via gRPC", zap.Uint32("id", req.Id), zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to update password: %v", err)
	}

	return &user.UpdateUserPasswordResponse{
		Success: true,
	}, nil
}

func (h *UserGrpcHandler) toProtoUser(u *dto.UserResponse) *user.User {
	return &user.User{
		Id:        uint32(u.ID),
		Name:      u.Name,
		Email:     u.Email,
		CreatedAt: timestamppb.New(u.CreatedAt),
		UpdatedAt: timestamppb.New(u.UpdatedAt),
	}
}
