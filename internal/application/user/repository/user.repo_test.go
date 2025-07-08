package repository

import (
	"fmt"
	"testing"

	"vibe-ddd-golang/internal/application/user/dto"
	"vibe-ddd-golang/internal/application/user/entity"
	"vibe-ddd-golang/internal/pkg/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestUserRepository_Create(t *testing.T) {
	// Setup
	db, err := testutil.SetupTestDB()
	require.NoError(t, err)
	logger := testutil.NewTestLogger(t)
	repo := NewUserRepository(db, logger)

	t.Run("should create user successfully", func(t *testing.T) {
		// Given
		user := testutil.CreateUserFixture()
		user.ID = 0 // Reset ID for creation

		// When
		err := repo.Create(user)

		// Then
		assert.NoError(t, err)
		assert.NotZero(t, user.ID)

		// Verify user was created in database
		var dbUser entity.User
		err = db.First(&dbUser, user.ID).Error
		assert.NoError(t, err)
		assert.Equal(t, user.Email, dbUser.Email)
		assert.Equal(t, user.Name, dbUser.Name)
	})

	t.Run("should fail to create user with duplicate email", func(t *testing.T) {
		// Given
		user1 := testutil.CreateUserFixture()
		user1.ID = 0
		user1.Email = "duplicate@example.com"

		user2 := testutil.CreateUserFixture()
		user2.ID = 0
		user2.Email = "duplicate@example.com"

		// When
		err1 := repo.Create(user1)
		err2 := repo.Create(user2)

		// Then
		assert.NoError(t, err1)
		assert.Error(t, err2) // Should fail due to unique constraint
	})

	// Cleanup
	testutil.CleanDB(db)
}

func TestUserRepository_GetByID(t *testing.T) {
	// Setup
	db, err := testutil.SetupTestDB()
	require.NoError(t, err)
	logger := testutil.NewTestLogger(t)
	repo := NewUserRepository(db, logger)

	t.Run("should get user by ID successfully", func(t *testing.T) {
		// Given
		user := testutil.CreateUserFixture()
		user.ID = 0
		err := repo.Create(user)
		require.NoError(t, err)

		// When
		foundUser, err := repo.GetByID(user.ID)

		// Then
		assert.NoError(t, err)
		assert.Equal(t, user.ID, foundUser.ID)
		assert.Equal(t, user.Email, foundUser.Email)
		assert.Equal(t, user.Name, foundUser.Name)
	})

	t.Run("should return error when user not found", func(t *testing.T) {
		// When
		_, err := repo.GetByID(999)

		// Then
		assert.Error(t, err)
		assert.Equal(t, gorm.ErrRecordNotFound, err)
	})

	// Cleanup
	testutil.CleanDB(db)
}

func TestUserRepository_GetByEmail(t *testing.T) {
	// Setup
	db, err := testutil.SetupTestDB()
	require.NoError(t, err)
	logger := testutil.NewTestLogger(t)
	repo := NewUserRepository(db, logger)

	t.Run("should get user by email successfully", func(t *testing.T) {
		// Given
		user := testutil.CreateUserFixture()
		user.ID = 0
		err := repo.Create(user)
		require.NoError(t, err)

		// When
		foundUser, err := repo.GetByEmail(user.Email)

		// Then
		assert.NoError(t, err)
		assert.Equal(t, user.ID, foundUser.ID)
		assert.Equal(t, user.Email, foundUser.Email)
		assert.Equal(t, user.Name, foundUser.Name)
	})

	t.Run("should return error when user email not found", func(t *testing.T) {
		// When
		_, err := repo.GetByEmail("nonexistent@example.com")

		// Then
		assert.Error(t, err)
		assert.Equal(t, gorm.ErrRecordNotFound, err)
	})

	// Cleanup
	testutil.CleanDB(db)
}

func TestUserRepository_GetAll(t *testing.T) {
	// Setup
	db, err := testutil.SetupTestDB()
	require.NoError(t, err)
	logger := testutil.NewTestLogger(t)
	repo := NewUserRepository(db, logger)

	t.Run("should get all users with pagination", func(t *testing.T) {
		// Given - Create multiple users
		for i := 0; i < 5; i++ {
			user := testutil.CreateUserFixture()
			user.ID = 0
			user.Email = fmt.Sprintf("user%d@example.com", i)
			user.Name = fmt.Sprintf("User %d", i)
			err := repo.Create(user)
			require.NoError(t, err)
		}

		filter := &dto.UserFilter{
			Page:     1,
			PageSize: 3,
		}

		// When
		users, totalCount, err := repo.GetAll(filter)

		// Then
		assert.NoError(t, err)
		assert.Len(t, users, 3)               // Should return 3 users due to page size
		assert.Equal(t, int64(5), totalCount) // Total count should be 5
	})

	t.Run("should filter users by name", func(t *testing.T) {
		// Given
		user1 := testutil.CreateUserFixture()
		user1.ID = 0
		user1.Email = "alice@example.com"
		user1.Name = "Alice Smith"
		err := repo.Create(user1)
		require.NoError(t, err)

		user2 := testutil.CreateUserFixture()
		user2.ID = 0
		user2.Email = "bob@example.com"
		user2.Name = "Bob Johnson"
		err = repo.Create(user2)
		require.NoError(t, err)

		filter := &dto.UserFilter{
			Name: "Alice",
		}

		// When
		users, totalCount, err := repo.GetAll(filter)

		// Then
		assert.NoError(t, err)
		assert.Len(t, users, 1)
		assert.Equal(t, int64(1), totalCount)
		assert.Equal(t, "Alice Smith", users[0].Name)
	})

	// Cleanup
	testutil.CleanDB(db)
}

func TestUserRepository_Update(t *testing.T) {
	// Setup
	db, err := testutil.SetupTestDB()
	require.NoError(t, err)
	logger := testutil.NewTestLogger(t)
	repo := NewUserRepository(db, logger)

	t.Run("should update user successfully", func(t *testing.T) {
		// Given
		user := testutil.CreateUserFixture()
		user.ID = 0
		err := repo.Create(user)
		require.NoError(t, err)

		// When
		user.Name = "Updated Name"
		user.Email = "updated@example.com"
		err = repo.Update(user)

		// Then
		assert.NoError(t, err)

		// Verify update in database
		var dbUser entity.User
		err = db.First(&dbUser, user.ID).Error
		assert.NoError(t, err)
		assert.Equal(t, "Updated Name", dbUser.Name)
		assert.Equal(t, "updated@example.com", dbUser.Email)
	})

	// Cleanup
	testutil.CleanDB(db)
}

func TestUserRepository_Delete(t *testing.T) {
	// Setup
	db, err := testutil.SetupTestDB()
	require.NoError(t, err)
	logger := testutil.NewTestLogger(t)
	repo := NewUserRepository(db, logger)

	t.Run("should delete user successfully", func(t *testing.T) {
		// Given
		user := testutil.CreateUserFixture()
		user.ID = 0
		err := repo.Create(user)
		require.NoError(t, err)

		// When
		err = repo.Delete(user.ID)

		// Then
		assert.NoError(t, err)

		// Verify user is deleted
		var dbUser entity.User
		err = db.First(&dbUser, user.ID).Error
		assert.Error(t, err)
		assert.Equal(t, gorm.ErrRecordNotFound, err)
	})

	// Cleanup
	testutil.CleanDB(db)
}

func TestUserRepository_EmailExists(t *testing.T) {
	// Setup
	db, err := testutil.SetupTestDB()
	require.NoError(t, err)
	logger := testutil.NewTestLogger(t)
	repo := NewUserRepository(db, logger)

	t.Run("should return true for existing email", func(t *testing.T) {
		// Given
		user := testutil.CreateUserFixture()
		user.ID = 0
		err := repo.Create(user)
		require.NoError(t, err)

		// When
		exists, err := repo.EmailExists(user.Email)

		// Then
		assert.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("should return false for non-existing email", func(t *testing.T) {
		// When
		exists, err := repo.EmailExists("nonexistent@example.com")

		// Then
		assert.NoError(t, err)
		assert.False(t, exists)
	})

	// Cleanup
	testutil.CleanDB(db)
}
