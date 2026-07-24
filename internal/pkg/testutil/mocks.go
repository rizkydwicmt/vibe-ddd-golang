package testutil

import (
	"context"

	"vibe-ddd-golang/internal/application/payment/dto"
	"vibe-ddd-golang/internal/application/payment/entity"
	userDto "vibe-ddd-golang/internal/application/user/dto"
	userEntity "vibe-ddd-golang/internal/application/user/entity"

	"github.com/stretchr/testify/mock"
)

// MockUserRepository is a mock implementation of UserRepository.
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *userEntity.User) error {
	return m.Called(ctx, user).Error(0)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id uint) (*userEntity.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userEntity.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*userEntity.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userEntity.User), args.Error(1)
}

func (m *MockUserRepository) GetAll(ctx context.Context, filter *userDto.UserFilter) ([]userEntity.User, int64, error) {
	args := m.Called(ctx, filter)
	var users []userEntity.User
	if args.Get(0) != nil {
		users = args.Get(0).([]userEntity.User)
	}
	var count int64
	if args.Get(1) != nil {
		count = args.Get(1).(int64)
	}
	return users, count, args.Error(2)
}

func (m *MockUserRepository) Update(ctx context.Context, user *userEntity.User) error {
	return m.Called(ctx, user).Error(0)
}

func (m *MockUserRepository) Delete(ctx context.Context, id uint) error {
	return m.Called(ctx, id).Error(0)
}

func (m *MockUserRepository) EmailExists(ctx context.Context, email string) (bool, error) {
	args := m.Called(ctx, email)
	return args.Bool(0), args.Error(1)
}

// MockPaymentRepository is a mock implementation of PaymentRepository.
type MockPaymentRepository struct {
	mock.Mock
}

func (m *MockPaymentRepository) Create(ctx context.Context, payment *entity.Payment) error {
	return m.Called(ctx, payment).Error(0)
}

func (m *MockPaymentRepository) GetByID(ctx context.Context, id uint) (*entity.Payment, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Payment), args.Error(1)
}

func (m *MockPaymentRepository) GetAll(ctx context.Context, filter *dto.PaymentFilter) ([]entity.Payment, int64, error) {
	args := m.Called(ctx, filter)
	var payments []entity.Payment
	if args.Get(0) != nil {
		payments = args.Get(0).([]entity.Payment)
	}
	var count int64
	if args.Get(1) != nil {
		count = args.Get(1).(int64)
	}
	return payments, count, args.Error(2)
}

func (m *MockPaymentRepository) Update(ctx context.Context, payment *entity.Payment) error {
	return m.Called(ctx, payment).Error(0)
}

func (m *MockPaymentRepository) Delete(ctx context.Context, id uint) error {
	return m.Called(ctx, id).Error(0)
}

func (m *MockPaymentRepository) GetByUserID(ctx context.Context, userID uint) ([]entity.Payment, error) {
	args := m.Called(ctx, userID)
	var payments []entity.Payment
	if args.Get(0) != nil {
		payments = args.Get(0).([]entity.Payment)
	}
	return payments, args.Error(1)
}

// MockUserService is a mock implementation of UserService.
type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) CreateUser(ctx context.Context, req *userDto.CreateUserRequest) (*userDto.UserResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userDto.UserResponse), args.Error(1)
}

func (m *MockUserService) GetUserByID(ctx context.Context, id uint) (*userDto.UserResponse, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userDto.UserResponse), args.Error(1)
}

func (m *MockUserService) GetUserByEmail(ctx context.Context, email string) (*userDto.UserResponse, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userDto.UserResponse), args.Error(1)
}

func (m *MockUserService) GetUsers(ctx context.Context, filter *userDto.UserFilter) (*userDto.UserListResponse, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userDto.UserListResponse), args.Error(1)
}

func (m *MockUserService) UpdateUser(ctx context.Context, id uint, req *userDto.UpdateUserRequest) (*userDto.UserResponse, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userDto.UserResponse), args.Error(1)
}

func (m *MockUserService) UpdateUserPassword(ctx context.Context, id uint, req *userDto.UpdateUserPasswordRequest) error {
	return m.Called(ctx, id, req).Error(0)
}

func (m *MockUserService) DeleteUser(ctx context.Context, id uint) error {
	return m.Called(ctx, id).Error(0)
}

// MockPaymentService is a mock implementation of PaymentService.
type MockPaymentService struct {
	mock.Mock
}

func (m *MockPaymentService) CreatePayment(ctx context.Context, req *dto.CreatePaymentRequest) (*dto.PaymentResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.PaymentResponse), args.Error(1)
}

func (m *MockPaymentService) GetPaymentByID(ctx context.Context, id uint) (*dto.PaymentResponse, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.PaymentResponse), args.Error(1)
}

func (m *MockPaymentService) GetPayments(ctx context.Context, filter *dto.PaymentFilter) (*dto.PaymentListResponse, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.PaymentListResponse), args.Error(1)
}

func (m *MockPaymentService) UpdatePayment(ctx context.Context, id uint, req *dto.UpdatePaymentRequest) (*dto.PaymentResponse, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.PaymentResponse), args.Error(1)
}

func (m *MockPaymentService) DeletePayment(ctx context.Context, id uint) error {
	return m.Called(ctx, id).Error(0)
}

func (m *MockPaymentService) GetPaymentsByUser(ctx context.Context, userID uint) ([]dto.PaymentResponse, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]dto.PaymentResponse), args.Error(1)
}
