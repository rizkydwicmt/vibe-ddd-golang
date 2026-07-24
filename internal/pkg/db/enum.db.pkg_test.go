package database

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDirectionEnum(t *testing.T) {
	tests := []struct {
		name     string
		enum     DirectionEnum
		expected string
		isValid  bool
	}{
		{
			name:     "ASC enum",
			enum:     ASC,
			expected: "asc",
			isValid:  true,
		},
		{
			name:     "DESC enum",
			enum:     DESC,
			expected: "desc",
			isValid:  true,
		},
		{
			name:     "Invalid enum",
			enum:     "invalid",
			expected: "",
			isValid:  false,
		},
		{
			name:     "Empty enum",
			enum:     "",
			expected: "",
			isValid:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test ToString method
			result := tt.enum.ToString()
			assert.Equal(t, tt.expected, result, "ToString() should return correct string")

			// Test IsValid method
			isValid := tt.enum.IsValid()
			assert.Equal(t, tt.isValid, isValid, "IsValid() should return correct boolean")
		})
	}
}

func TestDriverEnum(t *testing.T) {
	tests := []struct {
		name           string
		enum           DriverEnum
		expectedString string
		expectedShort  string
		isValid        bool
	}{
		{
			name:           "POSTGRES enum",
			enum:           POSTGRES,
			expectedString: "postgres",
			expectedShort:  "pg",
			isValid:        true,
		},
		{
			name:           "MYSQL enum",
			enum:           MYSQL,
			expectedString: "mysql",
			expectedShort:  "sql",
			isValid:        true,
		},
		{
			name:           "Invalid enum",
			enum:           "invalid",
			expectedString: "",
			expectedShort:  "",
			isValid:        false,
		},
		{
			name:           "Empty enum",
			enum:           "",
			expectedString: "",
			expectedShort:  "",
			isValid:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test ToString method
			result := tt.enum.ToString()
			assert.Equal(t, tt.expectedString, result, "ToString() should return correct string")

			// Test ToShortString method
			shortResult := tt.enum.ToShortString()
			assert.Equal(t, tt.expectedShort, shortResult, "ToShortString() should return correct short string")

			// Test IsValid method
			isValid := tt.enum.IsValid()
			assert.Equal(t, tt.isValid, isValid, "IsValid() should return correct boolean")
		})
	}
}

func TestDirectionEnum_Constants(t *testing.T) {
	t.Run("ASC constant value", func(t *testing.T) {
		assert.Equal(t, DirectionEnum("asc"), ASC, "ASC constant should have correct value")
	})

	t.Run("DESC constant value", func(t *testing.T) {
		assert.Equal(t, DirectionEnum("desc"), DESC, "DESC constant should have correct value")
	})
}

func TestDriverEnum_Constants(t *testing.T) {
	t.Run("POSTGRES constant value", func(t *testing.T) {
		assert.Equal(t, DriverEnum("postgres"), POSTGRES, "POSTGRES constant should have correct value")
	})

	t.Run("MYSQL constant value", func(t *testing.T) {
		assert.Equal(t, DriverEnum("mysql"), MYSQL, "MYSQL constant should have correct value")
	})
}

func TestDirectionEnum_EdgeCases(t *testing.T) {
	t.Run("Case sensitivity", func(t *testing.T) {
		// Test that the enum is case-sensitive
		upperAsc := DirectionEnum("ASC")
		lowerAsc := DirectionEnum("asc")

		assert.False(t, upperAsc.IsValid(), "Uppercase ASC should not be valid")
		assert.True(t, lowerAsc.IsValid(), "Lowercase asc should be valid")
	})

	t.Run("Whitespace handling", func(t *testing.T) {
		// Test that whitespace is not ignored
		spaceAsc := DirectionEnum(" asc ")

		assert.False(t, spaceAsc.IsValid(), "ASC with spaces should not be valid")
	})
}

func TestDriverEnum_EdgeCases(t *testing.T) {
	t.Run("Case sensitivity", func(t *testing.T) {
		// Test that the enum is case-sensitive
		upperPostgres := DriverEnum("POSTGRES")
		lowerPostgres := DriverEnum("postgres")

		assert.False(t, upperPostgres.IsValid(), "Uppercase POSTGRES should not be valid")
		assert.True(t, lowerPostgres.IsValid(), "Lowercase postgres should be valid")
	})

	t.Run("Whitespace handling", func(t *testing.T) {
		// Test that whitespace is not ignored
		spacePostgres := DriverEnum(" postgres ")

		assert.False(t, spacePostgres.IsValid(), "POSTGRES with spaces should not be valid")
	})
}
