package repository_test

import (
	"context"
	"testing"

	"vibe-ddd-golang/internal/application/payment/dto"
	"vibe-ddd-golang/internal/application/payment/entity"
	"vibe-ddd-golang/internal/application/payment/repository"
	"vibe-ddd-golang/internal/common/params"
	"vibe-ddd-golang/internal/pkg/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newRepo(t *testing.T) repository.PaymentRepository {
	t.Helper()
	db, err := testutil.SetupTestDatabase()
	require.NoError(t, err)
	return repository.NewPaymentRepository(params.Params{MainDB: db}, testutil.NewSilentLogger())
}

func seed(t *testing.T, ctx context.Context, repo repository.PaymentRepository, userID uint, status entity.PaymentStatus) *entity.Payment {
	t.Helper()
	p := &entity.Payment{Amount: 10, Currency: "USD", Status: status, Description: "x", UserID: userID}
	require.NoError(t, repo.Create(ctx, p))
	return p
}

func TestPaymentRepository_CreateAndGet(t *testing.T) {
	ctx := context.Background()
	repo := newRepo(t)

	p := seed(t, ctx, repo, 1, entity.PaymentStatusPending)
	got, err := repo.GetByID(ctx, p.ID)
	require.NoError(t, err)
	assert.Equal(t, entity.PaymentStatusPending, got.Status)
}

func TestPaymentRepository_GetAll_Filter(t *testing.T) {
	ctx := context.Background()
	repo := newRepo(t)

	seed(t, ctx, repo, 1, entity.PaymentStatusPending)
	seed(t, ctx, repo, 1, entity.PaymentStatusCompleted)
	seed(t, ctx, repo, 2, entity.PaymentStatusPending)

	all, total, err := repo.GetAll(ctx, &dto.PaymentFilter{UserID: 1})
	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, all, 2)

	byUser, err := repo.GetByUserID(ctx, 2)
	require.NoError(t, err)
	assert.Len(t, byUser, 1)
}

func TestPaymentRepository_UpdateAndDelete(t *testing.T) {
	ctx := context.Background()
	repo := newRepo(t)

	p := seed(t, ctx, repo, 1, entity.PaymentStatusPending)
	p.Status = entity.PaymentStatusCompleted
	require.NoError(t, repo.Update(ctx, p))

	got, err := repo.GetByID(ctx, p.ID)
	require.NoError(t, err)
	assert.Equal(t, entity.PaymentStatusCompleted, got.Status)

	require.NoError(t, repo.Delete(ctx, p.ID))
	_, err = repo.GetByID(ctx, p.ID)
	assert.Error(t, err)
}
