package repository

import (
	"testing"

	"vibe-ddd-golang/internal/application/payment/dto"
	"vibe-ddd-golang/internal/application/payment/entity"
	"vibe-ddd-golang/internal/pkg/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestPaymentRepository_Create(t *testing.T) {
	// Setup
	db, err := testutil.SetupTestDB()
	require.NoError(t, err)
	logger := testutil.NewTestLogger(t)
	repo := NewPaymentRepository(db, logger)

	t.Run("should create payment successfully", func(t *testing.T) {
		// Given
		payment := testutil.CreatePaymentFixture()
		payment.ID = 0 // Reset ID for creation

		// When
		err := repo.Create(payment)

		// Then
		assert.NoError(t, err)
		assert.NotZero(t, payment.ID)

		// Verify payment was created in database
		var dbPayment entity.Payment
		err = db.First(&dbPayment, payment.ID).Error
		assert.NoError(t, err)
		assert.Equal(t, payment.Amount, dbPayment.Amount)
		assert.Equal(t, payment.Currency, dbPayment.Currency)
		assert.Equal(t, payment.UserID, dbPayment.UserID)
	})

	// Cleanup
	testutil.CleanDB(db)
}

func TestPaymentRepository_GetByID(t *testing.T) {
	// Setup
	db, err := testutil.SetupTestDB()
	require.NoError(t, err)
	logger := testutil.NewTestLogger(t)
	repo := NewPaymentRepository(db, logger)

	t.Run("should get payment by ID successfully", func(t *testing.T) {
		// Given
		payment := testutil.CreatePaymentFixture()
		payment.ID = 0
		err := repo.Create(payment)
		require.NoError(t, err)

		// When
		foundPayment, err := repo.GetByID(payment.ID)

		// Then
		assert.NoError(t, err)
		assert.Equal(t, payment.ID, foundPayment.ID)
		assert.Equal(t, payment.Amount, foundPayment.Amount)
		assert.Equal(t, payment.Currency, foundPayment.Currency)
		assert.Equal(t, payment.UserID, foundPayment.UserID)
	})

	t.Run("should return error when payment not found", func(t *testing.T) {
		// When
		_, err := repo.GetByID(999)

		// Then
		assert.Error(t, err)
		assert.Equal(t, gorm.ErrRecordNotFound, err)
	})

	// Cleanup
	testutil.CleanDB(db)
}

func TestPaymentRepository_GetAll(t *testing.T) {
	// Setup
	db, err := testutil.SetupTestDB()
	require.NoError(t, err)
	logger := testutil.NewTestLogger(t)
	repo := NewPaymentRepository(db, logger)
	
	// Clean up function
	cleanup := func() {
		db.Exec("DELETE FROM payments")
	}

	t.Run("should get all payments with pagination", func(t *testing.T) {
		cleanup() // Clean before test
		// Given - Create multiple payments
		for i := 0; i < 5; i++ {
			payment := testutil.CreatePaymentFixture()
			payment.ID = 0
			payment.Amount = float64(100 + i)
			payment.UserID = uint(i + 1)
			err := repo.Create(payment)
			require.NoError(t, err)
		}

		filter := &dto.PaymentFilter{
			Page:     1,
			PageSize: 3,
		}

		// When
		payments, totalCount, err := repo.GetAll(filter)

		// Then
		assert.NoError(t, err)
		assert.Len(t, payments, 3)            // Should return 3 payments due to page size
		assert.Equal(t, int64(5), totalCount) // Total count should be 5
	})

	t.Run("should filter payments by status", func(t *testing.T) {
		cleanup() // Clean before test
		// Given
		payment1 := testutil.CreatePaymentFixture()
		payment1.ID = 0
		payment1.Status = entity.PaymentStatusPending
		payment1.UserID = 1
		err := repo.Create(payment1)
		require.NoError(t, err)

		payment2 := testutil.CreatePaymentFixture()
		payment2.ID = 0
		payment2.Status = entity.PaymentStatusCompleted
		payment2.UserID = 2
		err = repo.Create(payment2)
		require.NoError(t, err)

		filter := &dto.PaymentFilter{
			Status: entity.PaymentStatusPending.String(),
		}

		// When
		payments, totalCount, err := repo.GetAll(filter)

		// Then
		assert.NoError(t, err)
		assert.Len(t, payments, 1)
		assert.Equal(t, int64(1), totalCount)
		assert.Equal(t, entity.PaymentStatusPending, payments[0].Status)
	})

	t.Run("should filter payments by currency", func(t *testing.T) {
		cleanup() // Clean before test
		// Given
		payment1 := testutil.CreatePaymentFixture()
		payment1.ID = 0
		payment1.Currency = "USD"
		payment1.UserID = 1
		err := repo.Create(payment1)
		require.NoError(t, err)

		payment2 := testutil.CreatePaymentFixture()
		payment2.ID = 0
		payment2.Currency = "EUR"
		payment2.UserID = 2
		err = repo.Create(payment2)
		require.NoError(t, err)

		filter := &dto.PaymentFilter{
			Currency: "USD",
		}

		// When
		payments, totalCount, err := repo.GetAll(filter)

		// Then
		assert.NoError(t, err)
		assert.Len(t, payments, 1)
		assert.Equal(t, int64(1), totalCount)
		assert.Equal(t, "USD", payments[0].Currency)
	})

	t.Run("should filter payments by user ID", func(t *testing.T) {
		cleanup() // Clean before test
		// Given
		payment1 := testutil.CreatePaymentFixture()
		payment1.ID = 0
		payment1.UserID = 1
		err := repo.Create(payment1)
		require.NoError(t, err)

		payment2 := testutil.CreatePaymentFixture()
		payment2.ID = 0
		payment2.UserID = 2
		err = repo.Create(payment2)
		require.NoError(t, err)

		filter := &dto.PaymentFilter{
			UserID: 1,
		}

		// When
		payments, totalCount, err := repo.GetAll(filter)

		// Then
		assert.NoError(t, err)
		assert.Len(t, payments, 1)
		assert.Equal(t, int64(1), totalCount)
		assert.Equal(t, uint(1), payments[0].UserID)
	})

	// Cleanup
	testutil.CleanDB(db)
}

func TestPaymentRepository_Update(t *testing.T) {
	// Setup
	db, err := testutil.SetupTestDB()
	require.NoError(t, err)
	logger := testutil.NewTestLogger(t)
	repo := NewPaymentRepository(db, logger)

	t.Run("should update payment successfully", func(t *testing.T) {
		// Given
		payment := testutil.CreatePaymentFixture()
		payment.ID = 0
		err := repo.Create(payment)
		require.NoError(t, err)

		// When
		payment.Status = entity.PaymentStatusCompleted
		payment.Description = "Updated description"
		err = repo.Update(payment)

		// Then
		assert.NoError(t, err)

		// Verify update in database
		var dbPayment entity.Payment
		err = db.First(&dbPayment, payment.ID).Error
		assert.NoError(t, err)
		assert.Equal(t, entity.PaymentStatusCompleted, dbPayment.Status)
		assert.Equal(t, "Updated description", dbPayment.Description)
	})

	// Cleanup
	testutil.CleanDB(db)
}

func TestPaymentRepository_Delete(t *testing.T) {
	// Setup
	db, err := testutil.SetupTestDB()
	require.NoError(t, err)
	logger := testutil.NewTestLogger(t)
	repo := NewPaymentRepository(db, logger)

	t.Run("should delete payment successfully", func(t *testing.T) {
		// Given
		payment := testutil.CreatePaymentFixture()
		payment.ID = 0
		err := repo.Create(payment)
		require.NoError(t, err)

		// When
		err = repo.Delete(payment.ID)

		// Then
		assert.NoError(t, err)

		// Verify payment is deleted (soft delete with GORM)
		var dbPayment entity.Payment
		err = db.First(&dbPayment, payment.ID).Error
		assert.Error(t, err)
		assert.Equal(t, gorm.ErrRecordNotFound, err)
	})

	// Cleanup
	testutil.CleanDB(db)
}

func TestPaymentRepository_GetByUserID(t *testing.T) {
	// Setup
	db, err := testutil.SetupTestDB()
	require.NoError(t, err)
	logger := testutil.NewTestLogger(t)
	repo := NewPaymentRepository(db, logger)

	t.Run("should get payments by user ID successfully", func(t *testing.T) {
		// Given
		userID := uint(1)
		for i := 0; i < 3; i++ {
			payment := testutil.CreatePaymentFixture()
			payment.ID = 0
			payment.UserID = userID
			payment.Amount = float64(100 + i)
			err := repo.Create(payment)
			require.NoError(t, err)
		}

		// Create payment for different user
		payment := testutil.CreatePaymentFixture()
		payment.ID = 0
		payment.UserID = 2
		err = repo.Create(payment)
		require.NoError(t, err)

		// When
		payments, err := repo.GetByUserID(userID)

		// Then
		assert.NoError(t, err)
		assert.Len(t, payments, 3) // Should return only payments for user 1
		for _, p := range payments {
			assert.Equal(t, userID, p.UserID)
		}
	})

	t.Run("should return empty slice for user with no payments", func(t *testing.T) {
		// When
		payments, err := repo.GetByUserID(999)

		// Then
		assert.NoError(t, err)
		assert.Empty(t, payments)
	})

	// Cleanup
	testutil.CleanDB(db)
}
