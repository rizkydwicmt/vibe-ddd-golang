package redis

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDefaultConfig(t *testing.T) {
	tests := []struct {
		name     string
		input    *Config
		expected *Config
	}{
		{
			name:  "All defaults",
			input: &Config{},
			expected: &Config{
				ReconnectInterval:   1 * time.Second,
				HealthCheckInterval: 3 * time.Second,
				DialTimeout:         5 * time.Second,
			},
		},
		{
			name: "Partial override",
			input: &Config{
				ReconnectInterval: 10 * time.Second,
			},
			expected: &Config{
				ReconnectInterval:   10 * time.Second,
				HealthCheckInterval: 3 * time.Second,
				DialTimeout:         5 * time.Second,
			},
		},
		{
			name: "All custom",
			input: &Config{
				ReconnectInterval:   2 * time.Second,
				HealthCheckInterval: 4 * time.Second,
				DialTimeout:         6 * time.Second,
			},
			expected: &Config{
				ReconnectInterval:   2 * time.Second,
				HealthCheckInterval: 4 * time.Second,
				DialTimeout:         6 * time.Second,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := defaultConfig(tt.input)
			assert.Equal(t, tt.expected.ReconnectInterval, result.ReconnectInterval)
			assert.Equal(t, tt.expected.HealthCheckInterval, result.HealthCheckInterval)
			assert.Equal(t, tt.expected.DialTimeout, result.DialTimeout)
		})
	}
}

func TestIsConnectionError(t *testing.T) {
	// Create a dummy client to access the method
	client := &Client{}

	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "Nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "Connection refused",
			err:      errors.New("dial tcp 127.0.0.1:6379: connection refused"),
			expected: true,
		},
		{
			name:     "EOF error",
			err:      errors.New("unexpected EOF"),
			expected: true,
		},
		{
			name:     "Broken pipe",
			err:      errors.New("write: broken pipe"),
			expected: true,
		},
		{
			name:     "Key not found (Redis error)",
			err:      errors.New("redis: nil"),
			expected: false,
		},
		{
			name:     "Type error (Redis error)",
			err:      errors.New("WRONGTYPE Operation against a key holding the wrong kind of value"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := client.isConnectionError(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		substr   string
		expected bool
	}{
		{"Contains substring", "hello world", "world", true},
		{"Does not contain", "hello world", "mars", false},
		{"Empty string", "", "test", false},
		{"Empty substring", "test", "", true},
		{"Exact match", "test", "test", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := contains(tt.s, tt.substr)
			assert.Equal(t, tt.expected, result)
		})
	}
}
