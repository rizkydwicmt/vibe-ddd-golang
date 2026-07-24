package repository_test

import (
	"context"
	"testing"

	"vibe-ddd-golang/internal/application/user/dto"
	"vibe-ddd-golang/internal/application/user/entity"
	"vibe-ddd-golang/internal/application/user/repository"
	"vibe-ddd-golang/internal/common/params"
	"vibe-ddd-golang/internal/pkg/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newRepo(t *testing.T) repository.UserRepository {
	t.Helper()
	db, err := testutil.SetupTestDatabase()
	require.NoError(t, err)
	return repository.NewUserRepository(params.Params{MainDB: db}, testutil.NewSilentLogger())
}

func TestUserRepository_CreateAndGet(t *testing.T) {
	ctx := context.Background()
	repo := newRepo(t)

	u := &entity.User{Name: "John", Email: "john@example.com", Password: "hash"}
	require.NoError(t, repo.Create(ctx, u))
	assert.NotZero(t, u.ID)

	got, err := repo.GetByID(ctx, u.ID)
	require.NoError(t, err)
	assert.Equal(t, "john@example.com", got.Email)

	byEmail, err := repo.GetByEmail(ctx, "john@example.com")
	require.NoError(t, err)
	assert.Equal(t, u.ID, byEmail.ID)
}

func TestUserRepository_EmailExists(t *testing.T) {
	ctx := context.Background()
	repo := newRepo(t)

	exists, err := repo.EmailExists(ctx, "none@example.com")
	require.NoError(t, err)
	assert.False(t, exists)

	require.NoError(t, repo.Create(ctx, &entity.User{Name: "A", Email: "a@example.com", Password: "h"}))
	exists, err = repo.EmailExists(ctx, "a@example.com")
	require.NoError(t, err)
	assert.True(t, exists)
}

func TestUserRepository_GetAll_FilterAndPaginate(t *testing.T) {
	ctx := context.Background()
	repo := newRepo(t)

	for _, e := range []string{"alice@x.com", "bob@x.com", "carol@x.com"} {
		require.NoError(t, repo.Create(ctx, &entity.User{Name: e, Email: e, Password: "h"}))
	}

	users, total, err := repo.GetAll(ctx, &dto.UserFilter{Page: 1, PageSize: 2})
	require.NoError(t, err)
	assert.Equal(t, int64(3), total)
	assert.Len(t, users, 2)

	users, total, err = repo.GetAll(ctx, &dto.UserFilter{Email: "bob"})
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, users, 1)
}

func TestUserRepository_UpdateAndDelete(t *testing.T) {
	ctx := context.Background()
	repo := newRepo(t)

	u := &entity.User{Name: "Old", Email: "old@x.com", Password: "h"}
	require.NoError(t, repo.Create(ctx, u))

	u.Name = "New"
	require.NoError(t, repo.Update(ctx, u))
	got, err := repo.GetByID(ctx, u.ID)
	require.NoError(t, err)
	assert.Equal(t, "New", got.Name)

	require.NoError(t, repo.Delete(ctx, u.ID))
	_, err = repo.GetByID(ctx, u.ID)
	assert.Error(t, err)
}
