package rabbitmq

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"vibe-ddd-golang/internal/pkg/logger"

	amqp "github.com/rabbitmq/amqp091-go"
)

type ChannelManager struct {
	connManager   *ConnectionManager
	channel       *amqp.Channel
	mu            sync.Mutex
	closed        bool
	maxRetries    int
	retryInterval time.Duration
	ctx           context.Context
	cancel        context.CancelFunc
}

func NewChannelManager(ctx context.Context, connManager *ConnectionManager) *ChannelManager {
	ctx, cancel := context.WithCancel(ctx)
	c := &ChannelManager{
		connManager:   connManager,
		maxRetries:    5,
		retryInterval: time.Second * 2,
		ctx:           ctx,
		cancel:        cancel,
		closed:        false,
	}

	ch, err := c.GetChannel()
	if err != nil {
		logger.Error.Printf("Failed to establish initial channel: %v\n", err)
	}

	c.channel = ch

	return c
}

func (cm *ChannelManager) GetChannel() (*amqp.Channel, error) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if cm.closed {
		return nil, errors.New("channel manager is closed")
	}

	if cm.channel != nil {
		return cm.channel, nil
	}

	return cm.setupChannelWithRetry()
}

func (cm *ChannelManager) setupChannelWithRetry() (*amqp.Channel, error) {
	var err error
	for attempt := 0; attempt < cm.maxRetries; attempt++ {
		if attempt > 0 {
			logger.Info.Printf("Retrying channel setup (attempt %d/%d)\n", attempt+1, cm.maxRetries)
			time.Sleep(cm.retryInterval)
		}

		cm.channel, err = cm.setupChannel()
		if err == nil {
			return cm.channel, nil
		}

		logger.Warning.Printf("Failed to setup channel: %v\n", err)
	}

	return nil, fmt.Errorf("failed to setup channel after %d attempts: %w", cm.maxRetries, err)
}

func (cm *ChannelManager) setupChannel() (*amqp.Channel, error) {
	conn := cm.connManager.GetConnection()
	if conn == nil {
		return nil, errors.New("no connection available")
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	if err := ch.Confirm(false); err != nil {
		ch.Close()
		return nil, fmt.Errorf("failed to enable publisher confirms: %w", err)
	}

	go cm.channelMonitor(ch)

	return ch, nil
}

func (cm *ChannelManager) channelMonitor(ch *amqp.Channel) {
	chErr := make(chan *amqp.Error)
	ch.NotifyClose(chErr)

	select {
	case err := <-chErr:
		if err != nil {
			logger.Warning.Printf("Channel closed: %v. Attempting to reconnect...\n", err)
			cm.reconnect()
		}
	case <-cm.ctx.Done():
		return
	}
}

func (cm *ChannelManager) reconnect() {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if cm.closed {
		return
	}

	cm.channel = nil
	_, err := cm.setupChannelWithRetry()
	if err != nil {
		logger.Error.Printf("Failed to reconnect: %v\n", err)
	} else {
		logger.Info.Println("Successfully reconnected and established a new channel")
	}
}

func (cm *ChannelManager) Close() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if cm.closed {
		return nil
	}

	cm.closed = true
	cm.cancel()

	if cm.channel != nil {
		err := cm.channel.Close()
		cm.channel = nil
		if err != nil {
			return fmt.Errorf("failed to close channel: %w", err)
		}
	}

	return nil
}
