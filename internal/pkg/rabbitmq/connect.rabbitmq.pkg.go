package rabbitmq

import (
	"context"
	"fmt"
	"sync"
	"time"

	"vibe-ddd-golang/internal/pkg/logger"

	amqp "github.com/rabbitmq/amqp091-go"
)

type ConnectionManager struct {
	conn          *amqp.Connection
	mu            sync.Mutex
	url           string
	isConnected   bool
	retryInterval time.Duration
	ctx           context.Context
	cancel        context.CancelFunc
}

type QueueConfig struct {
	Durable    bool
	AutoDelete bool
	Exclusive  bool
	NoWait     bool
	Args       amqp.Table
}

func DefaultQueueConfig() *QueueConfig {
	config := &QueueConfig{
		Durable:    true,
		AutoDelete: false,
		Exclusive:  false,
		NoWait:     false,
		Args:       nil,
	}

	return config
}

type Config struct {
	Username string
	Password string
	Host     string
	Port     int
	URI      string
}

func NewConnectionManager(ctx context.Context, config *Config) (*ConnectionManager, error) {
	ctx, cancel := context.WithCancel(ctx)

	var url string
	if config.URI != "" {
		url = config.URI
	} else {
		url = fmt.Sprintf("amqp://%s:%s@%s:%d/", config.Username, config.Password, config.Host, config.Port)
	}

	cm := &ConnectionManager{
		url:           url,
		retryInterval: time.Second * 2,
		ctx:           ctx,
		cancel:        cancel,
	}

	if err := cm.connect(); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create connection: %w", err)
	}

	if err := cm.WaitForConnection(5 * time.Second); err != nil {
		return nil, fmt.Errorf("failed to wait for connection: %w", err)
	}

	return cm, nil
}

func (cm *ConnectionManager) connect() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if cm.isConnected {
		return nil
	}

	if err := cm.ctx.Err(); err != nil {
		return fmt.Errorf("context canceled: %w", err)
	}

	conn, err := amqp.Dial(cm.url)
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	cm.conn = conn
	cm.isConnected = true

	go cm.connectionMonitor()

	return nil
}

func (cm *ConnectionManager) connectionMonitor() {
	connErr := make(chan *amqp.Error)
	cm.conn.NotifyClose(connErr)

	for {
		select {
		case err := <-connErr:
			if err != nil {
				cm.isConnected = false
				logger.Warning.Printf("Connection lost: %v. Attempting to reconnect...\n", err)

				for !cm.isConnected {
					select {
					case <-cm.ctx.Done():
						return
					default:
						if err := cm.connect(); err != nil {
							logger.Warning.Printf("Failed to reconnect: %v. Retrying in %v...\n", err, cm.retryInterval)
							time.Sleep(cm.retryInterval)
							continue
						}
						fmt.Println("Reconnected successfully")
					}
				}
			}
		case <-cm.ctx.Done():
			return
		}
	}
}

func (cm *ConnectionManager) GetConnection() *amqp.Connection {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if cm.ctx.Err() != nil {
		return nil
	}

	return cm.conn
}

func (cm *ConnectionManager) Close() error {
	cm.cancel()

	cm.mu.Lock()
	defer cm.mu.Unlock()

	if cm.conn != nil {
		if err := cm.conn.Close(); err != nil {
			return fmt.Errorf("failed to close connection: %w", err)
		}
		cm.conn = nil
	}

	cm.isConnected = false
	return nil
}

func (cm *ConnectionManager) IsClosed() bool {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	return cm.ctx.Err() != nil || !cm.isConnected
}

func (cm *ConnectionManager) IsHealthy() bool {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if cm.ctx.Err() != nil || !cm.isConnected || cm.conn == nil {
		return false
	}

	return !cm.conn.IsClosed()
}

func (cm *ConnectionManager) WaitForConnection(timeout time.Duration) error {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		if cm.IsHealthy() {
			return nil
		}

		select {
		case <-cm.ctx.Done():
			return fmt.Errorf("context canceled while waiting for connection")
		case <-time.After(100 * time.Millisecond):
			continue
		}
	}

	return fmt.Errorf("connection not ready within timeout: %v", timeout)
}
