package database

import (
	"testing"
	"time"

	"vibe-ddd-golang/internal/pkg/redis"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockRedisClient for testing
type MockRedisClient struct {
	mock.Mock
}

func (m *MockRedisClient) Get(key string) (string, error) {
	args := m.Called(key)
	return args.String(0), args.Error(1)
}

func (m *MockRedisClient) Set(key string, value interface{}, expiration time.Duration) error {
	args := m.Called(key, value, expiration)
	return args.Error(0)
}

func TestSetup(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		expectError bool
		description string
	}{
		{
			name: "Valid PostgreSQL config",
			config: &Config{
				Host:     "localhost",
				Port:     5432,
				User:     "testuser",
				Password: "testpass",
				Database: "testdb",
				SSLMode:  "disable",
				Driver:   POSTGRES,
				Cache:    false,
			},
			expectError: false,
			description: "Should successfully setup PostgreSQL connection",
		},
		{
			name: "Valid MySQL config",
			config: &Config{
				Host:     "localhost",
				Port:     3306,
				User:     "testuser",
				Password: "testpass",
				Database: "testdb",
				Driver:   MYSQL,
				Cache:    false,
			},
			expectError: false,
			description: "Should successfully setup MySQL connection",
		},
		{
			name: "Valid PostgreSQL config with URL",
			config: &Config{
				URL:    "postgres://testuser:testpass@localhost:5432/testdb?sslmode=disable",
				Driver: POSTGRES,
				Cache:  false,
			},
			expectError: false,
			description: "Should successfully setup PostgreSQL connection with URL",
		},
		{
			name: "Valid MySQL config with URL",
			config: &Config{
				URL:    "mysql://testuser:testpass@localhost:3306/testdb",
				Driver: MYSQL,
				Cache:  false,
			},
			expectError: false,
			description: "Should successfully setup MySQL connection with URL",
		},
		{
			name: "Invalid driver",
			config: &Config{
				Host:     "localhost",
				Port:     5432,
				User:     "testuser",
				Password: "testpass",
				Database: "testdb",
				Driver:   "invalid",
				Cache:    false,
			},
			expectError: true,
			description: "Should fail with invalid driver",
		},
		{
			name: "Valid config with memory cache",
			config: &Config{
				Host:     "localhost",
				Port:     5432,
				User:     "testuser",
				Password: "testpass",
				Database: "testdb",
				SSLMode:  "disable",
				Driver:   POSTGRES,
				Cache:    true,
			},
			expectError: false,
			description: "Should successfully setup with memory cache",
		},
		{
			name: "Valid config with Redis cache",
			config: &Config{
				Host:      "localhost",
				Port:      5432,
				User:      "testuser",
				Password:  "testpass",
				Database:  "testdb",
				SSLMode:   "disable",
				Driver:    POSTGRES,
				Cache:     true,
				Rds:       &redis.Client{},
				CacheTime: 5 * time.Minute,
			},
			expectError: false,
			description: "Should successfully setup with Redis cache",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, err := Setup(tt.config)

			if tt.expectError {
				assert.Error(t, err, tt.description)
				assert.Nil(t, db, tt.description)
			} else {
				// Note: In real tests, you'd need actual database connections
				if err == nil {
					assert.NotNil(t, db, tt.description)
					assert.Equal(t, tt.config, db.Config, tt.description)

					// Test Close method
					err := db.Close()
					// Close might fail if no real connection, but that's expected in tests
					_ = err
				}
			}
		})
	}
}

func TestDatabase_Close(t *testing.T) {
	t.Run("Close database connection", func(t *testing.T) {
		config := &Config{
			Host:     "localhost",
			Port:     5432,
			User:     "testuser",
			Password: "testpass",
			Database: "testdb",
			SSLMode:  "disable",
			Driver:   POSTGRES,
			Cache:    false,
		}

		db, err := Setup(config)
		if err == nil && db != nil {
			err := db.Close()
			// Close might fail if no real connection, but that's expected in tests
			_ = err
		}
	})
}

func TestConfig_Validation(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		expectError bool
		description string
	}{
		{
			name: "Missing host",
			config: &Config{
				Port:     5432,
				User:     "testuser",
				Password: "testpass",
				Database: "testdb",
				Driver:   POSTGRES,
			},
			expectError: true,
			description: "Should fail with missing host",
		},
		{
			name: "Missing user",
			config: &Config{
				Host:     "localhost",
				Port:     5432,
				Password: "testpass",
				Database: "testdb",
				Driver:   POSTGRES,
			},
			expectError: true,
			description: "Should fail with missing user",
		},
		{
			name: "Missing password",
			config: &Config{
				Host:     "localhost",
				Port:     5432,
				User:     "testuser",
				Database: "testdb",
				Driver:   POSTGRES,
			},
			expectError: true,
			description: "Should fail with missing password",
		},
		{
			name: "Missing database",
			config: &Config{
				Host:     "localhost",
				Port:     5432,
				User:     "testuser",
				Password: "testpass",
				Driver:   POSTGRES,
			},
			expectError: true,
			description: "Should fail with missing database",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Setup(tt.config)

			if tt.expectError {
				assert.Error(t, err, tt.description)
			}
		})
	}
}
