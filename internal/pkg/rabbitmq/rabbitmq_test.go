package rabbitmq

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultQueueConfig(t *testing.T) {
	config := DefaultQueueConfig()

	// Default values verification
	assert.True(t, config.Durable, "Queue should be durable by default")
	assert.False(t, config.AutoDelete, "Queue should not be auto-deleted by default")
	assert.False(t, config.Exclusive, "Queue should not be exclusive by default")
	assert.False(t, config.NoWait, "Queue should not be no-wait by default")
	assert.Nil(t, config.Args, "Queue args should be nil by default")
}

// Ensure ConnectionManager implements basic interface methods (compile-time check)
// This is a common pattern to ensure struct compatibility
type Closer interface {
	Close() error
}

func TestConnectionManager_Interface(t *testing.T) {
	var _ Closer = &ConnectionManager{}
}

// func TestNewConnectionManager(t *testing.T) {
// 	// This test requires a running RabbitMQ instance.
// 	// It is commented out to prevent failure in CI/CD environment without RabbitMQ.
//  // If you want to run this test, ensure RabbitMQ is running on localhost:5672
// 	/*
// 	ctx := context.Background()
// 	config := &Config{
// 		Host: "localhost",
// 		Port: 5672,
// 		Username: "guest",
// 		Password: "guest",
// 	}
// 	cm, err := NewConnectionManager(ctx, config)
// 	assert.NoError(t, err)
// 	defer cm.Close()
// 	assert.True(t, cm.IsHealthy())
// 	*/
// }

func TestConnectionManager_ConfigLogic(t *testing.T) {
	// We cannot test success path of NewConnectionManager without mocking amqp.Dial
	// But we can test failure path (invalid host) which confirms the connection logic attempts to run

	ctx := context.Background()
	config := &Config{
		Host:     "invalid-host-for-testing",
		Port:     1234,
		Username: "user",
		Password: "pass",
	}

	// This should fail quickly
	cm, err := NewConnectionManager(ctx, config)

	assert.Error(t, err)
	assert.Nil(t, cm)
	assert.Contains(t, err.Error(), "failed to create connection")
}
