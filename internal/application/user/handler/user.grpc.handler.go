package handler

import (
	"context"

	"vibe-ddd-golang/internal/application/user/dto"
	"vibe-ddd-golang/internal/application/user/service"
	"vibe-ddd-golang/internal/pkg/grpcstatus"
	"vibe-ddd-golang/internal/pkg/response"
	"vibe-ddd-golang/internal/pkg/validation"
	userv1 "vibe-ddd-golang/internal/server/grpc/proto/user"

	"google.golang.org/protobuf/types/known/timestamppb"
)

type UserGRPCServer struct {
	userv1.UnimplementedUserServiceServer
	service service.UserService
}

func NewUserGRPCServer(service service.UserService) *UserGRPCServer {
	return &UserGRPCServer{service: service}
}

func (s *UserGRPCServer) CreateUser(ctx context.Context, req *userv1.CreateUserRequest) (*userv1.CreateUserResponse, error) {
	in := &dto.CreateUserRequest{
		Name:     req.GetName(),
		Email:    req.GetEmail(),
		Password: req.GetPassword(),
	}
	if err := validateGRPC(in); err != nil {
		return nil, err
	}

	user, err := s.service.CreateUser(ctx, in)
	if err != nil {
		return nil, grpcstatus.FromError(err)
	}
	return &userv1.CreateUserResponse{User: userToProto(user)}, nil
}

func (s *UserGRPCServer) GetUser(ctx context.Context, req *userv1.GetUserRequest) (*userv1.GetUserResponse, error) {
	user, err := s.service.GetUserByID(ctx, uint(req.GetId()))
	if err != nil {
		return nil, grpcstatus.FromError(err)
	}
	return &userv1.GetUserResponse{User: userToProto(user)}, nil
}

func (s *UserGRPCServer) ListUsers(ctx context.Context, req *userv1.ListUsersRequest) (*userv1.ListUsersResponse, error) {
	users, err := s.service.GetUsers(ctx, &dto.UserFilter{
		Name:     req.GetName(),
		Email:    req.GetEmail(),
		Page:     int(req.GetPage()),
		PageSize: int(req.GetPageSize()),
	})
	if err != nil {
		return nil, grpcstatus.FromError(err)
	}

	out := make([]*userv1.User, 0, len(users.Data))
	for i := range users.Data {
		out = append(out, userToProto(&users.Data[i]))
	}

	return &userv1.ListUsersResponse{
		Users:    out,
		Total:    users.TotalCount,
		Page:     int32(users.Page),
		PageSize: int32(users.PageSize),
	}, nil
}

func (s *UserGRPCServer) UpdateUser(ctx context.Context, req *userv1.UpdateUserRequest) (*userv1.UpdateUserResponse, error) {
	in := &dto.UpdateUserRequest{
		Name:  req.GetName(),
		Email: req.GetEmail(),
	}
	if err := validateGRPC(in); err != nil {
		return nil, err
	}

	user, err := s.service.UpdateUser(ctx, uint(req.GetId()), in)
	if err != nil {
		return nil, grpcstatus.FromError(err)
	}
	return &userv1.UpdateUserResponse{User: userToProto(user)}, nil
}

func (s *UserGRPCServer) DeleteUser(ctx context.Context, req *userv1.DeleteUserRequest) (*userv1.DeleteUserResponse, error) {
	if err := s.service.DeleteUser(ctx, uint(req.GetId())); err != nil {
		return nil, grpcstatus.FromError(err)
	}
	return &userv1.DeleteUserResponse{Success: true}, nil
}

func (s *UserGRPCServer) UpdateUserPassword(
	ctx context.Context, req *userv1.UpdateUserPasswordRequest,
) (*userv1.UpdateUserPasswordResponse, error) {
	in := &dto.UpdateUserPasswordRequest{
		CurrentPassword: req.GetCurrentPassword(),
		NewPassword:     req.GetNewPassword(),
	}
	if err := validateGRPC(in); err != nil {
		return nil, err
	}

	if err := s.service.UpdateUserPassword(ctx, uint(req.GetId()), in); err != nil {
		return nil, grpcstatus.FromError(err)
	}
	return &userv1.UpdateUserPasswordResponse{Success: true}, nil
}

func validateGRPC(in any) error {
	if err := validation.Validate(in); err != nil {
		return grpcstatus.FromError(response.NewBadRequestException(err.Error()))
	}
	return nil
}

func userToProto(user *dto.UserResponse) *userv1.User {
	if user == nil {
		return nil
	}
	return &userv1.User{
		Id:        uint32(user.ID),
		Name:      user.Name,
		Email:     user.Email,
		CreatedAt: timestamppb.New(user.CreatedAt),
		UpdatedAt: timestamppb.New(user.UpdatedAt),
	}
}
