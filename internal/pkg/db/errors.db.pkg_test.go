package database

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestIsNotFound(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "GORM RecordNotFound error",
			err:      gorm.ErrRecordNotFound,
			expected: true,
		},
		{
			name:     "Wrapped GORM RecordNotFound error",
			err:      errors.New("some context: " + gorm.ErrRecordNotFound.Error()),
			expected: false, // errors.Is won't work with string concatenation
		},
		{
			name:     "Wrapped with errors.Wrap",
			err:      errors.New("some context: " + gorm.ErrRecordNotFound.Error()),
			expected: false,
		},
		{
			name:     "Different error",
			err:      errors.New("some other error"),
			expected: false,
		},
		{
			name:     "Nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "GORM DuplicatedKey error",
			err:      gorm.ErrDuplicatedKey,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsNotFound(tt.err)
			assert.Equal(t, tt.expected, result, "IsNotFound() should return correct boolean")
		})
	}
}

func TestIsDuplicateKey(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "GORM DuplicatedKey error",
			err:      gorm.ErrDuplicatedKey,
			expected: true,
		},
		{
			name:     "Wrapped GORM DuplicatedKey error",
			err:      errors.New("some context: " + gorm.ErrDuplicatedKey.Error()),
			expected: false, // errors.Is won't work with string concatenation
		},
		{
			name:     "Different error",
			err:      errors.New("some other error"),
			expected: false,
		},
		{
			name:     "Nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "GORM RecordNotFound error",
			err:      gorm.ErrRecordNotFound,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsDuplicateKey(tt.err)
			assert.Equal(t, tt.expected, result, "IsDuplicateKey() should return correct boolean")
		})
	}
}

func TestErrorFunctions_WithWrappedErrors(t *testing.T) {
	t.Run("IsNotFound with wrapped error", func(t *testing.T) {
		// Create a wrapped error using fmt.Errorf
		wrappedErr := errors.New("database operation failed: " + gorm.ErrRecordNotFound.Error())

		// This should return false because errors.Is won't match string concatenation
		result := IsNotFound(wrappedErr)
		assert.False(t, result, "IsNotFound() should return false for string-wrapped error")
	})

	t.Run("IsDuplicateKey with wrapped error", func(t *testing.T) {
		// Create a wrapped error using fmt.Errorf
		wrappedErr := errors.New("database operation failed: " + gorm.ErrDuplicatedKey.Error())

		// This should return false because errors.Is won't match string concatenation
		result := IsDuplicateKey(wrappedErr)
		assert.False(t, result, "IsDuplicateKey() should return false for string-wrapped error")
	})
}

func TestErrorFunctions_EdgeCases(t *testing.T) {
	t.Run("IsNotFound with empty error string", func(t *testing.T) {
		emptyErr := errors.New("")
		result := IsNotFound(emptyErr)
		assert.False(t, result, "IsNotFound() should return false for empty error")
	})

	t.Run("IsDuplicateKey with empty error string", func(t *testing.T) {
		emptyErr := errors.New("")
		result := IsDuplicateKey(emptyErr)
		assert.False(t, result, "IsDuplicateKey() should return false for empty error")
	})

	t.Run("IsNotFound with similar error message", func(t *testing.T) {
		similarErr := errors.New("record not found")
		result := IsNotFound(similarErr)
		assert.False(t, result, "IsNotFound() should return false for similar error message")
	})

	t.Run("IsDuplicateKey with similar error message", func(t *testing.T) {
		similarErr := errors.New("duplicate key")
		result := IsDuplicateKey(similarErr)
		assert.False(t, result, "IsDuplicateKey() should return false for similar error message")
	})
}

func TestErrorFunctions_Consistency(t *testing.T) {
	t.Run("Same error should return consistent results", func(t *testing.T) {
		err := gorm.ErrRecordNotFound

		// Call multiple times
		result1 := IsNotFound(err)
		result2 := IsNotFound(err)
		result3 := IsNotFound(err)

		assert.Equal(t, result1, result2, "IsNotFound() should return consistent results")
		assert.Equal(t, result2, result3, "IsNotFound() should return consistent results")
		assert.True(t, result1, "IsNotFound() should return true for GORM RecordNotFound error")
	})

	t.Run("Same duplicate key error should return consistent results", func(t *testing.T) {
		err := gorm.ErrDuplicatedKey

		// Call multiple times
		result1 := IsDuplicateKey(err)
		result2 := IsDuplicateKey(err)
		result3 := IsDuplicateKey(err)

		assert.Equal(t, result1, result2, "IsDuplicateKey() should return consistent results")
		assert.Equal(t, result2, result3, "IsDuplicateKey() should return consistent results")
		assert.True(t, result1, "IsDuplicateKey() should return true for GORM DuplicatedKey error")
	})
}
