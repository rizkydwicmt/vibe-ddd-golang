package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"vibe-ddd-golang/internal/pkg/helper/convert"
	"vibe-ddd-golang/internal/pkg/logger"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Publisher struct {
	connManager       *ConnectionManager
	channelManager    *ChannelManager
	mu                sync.Mutex
	wg                sync.WaitGroup
	maxRetries        int
	retryInterval     time.Duration
	ctx               context.Context
	cancel            context.CancelFunc
	rpcResponses      map[string]chan *RPCResponse
	rpcResponsesMu    sync.RWMutex
	rpcConsumerMu     sync.Mutex
	rpcConsumerActive bool
}

type PublishOptions struct {
	QueueOpts    *QueueConfig
	QueueName    string
	Pattern      string
	Exchange     string
	Mandatory    bool
	Immediate    bool
	MaxRetries   int
	RetryBackoff time.Duration
	IsRPC        bool
	Timeout      time.Duration
}

func DefaultPublishOptions(queueName, pattern string, isRPC bool) *PublishOptions {
	opts := &PublishOptions{
		QueueOpts:    DefaultQueueConfig(),
		Exchange:     "",
		QueueName:    queueName,
		Pattern:      pattern,
		Mandatory:    false,
		Immediate:    false,
		MaxRetries:   3,
		RetryBackoff: time.Second * 10,
		Timeout:      time.Minute,
		IsRPC:        isRPC,
	}

	if isRPC {
		opts.QueueOpts.Durable = true
		if pattern == "" {
			pattern = "general"
		}
	}

	return opts
}

func NewPublisher(ctx context.Context, connManager *ConnectionManager) (*Publisher, error) {
	ctx, cancel := context.WithCancel(ctx)

	pub := &Publisher{
		connManager:    connManager,
		maxRetries:     3,
		retryInterval:  time.Second * 2,
		ctx:            ctx,
		cancel:         cancel,
		channelManager: NewChannelManager(ctx, connManager),
		rpcResponses:   make(map[string]chan *RPCResponse),
	}

	return pub, nil
}

func (p *Publisher) ensureQueue(name string, cfg *QueueConfig) error {
	ch, err := p.channelManager.GetChannel()
	if err != nil || ch == nil {
		return fmt.Errorf("failed to get healthy channel: %w", err)
	}

	if ch.IsClosed() {
		return fmt.Errorf("channel is closed")
	}

	_, err = ch.QueueDeclare(
		name,
		cfg.Durable,
		cfg.AutoDelete,
		cfg.Exclusive,
		cfg.NoWait,
		cfg.Args,
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	return nil
}

func (p *Publisher) Publish(msg *Message, opts *PublishOptions) (interface{}, error) {
	to, cancel := context.WithTimeout(p.ctx, opts.Timeout)
	defer cancel()
	return p.PublishWithContext(to, msg, opts)
}

func (p *Publisher) PublishDefault(queueName, pattern string, payload any) error {
	opts := DefaultPublishOptions(queueName, pattern, false)
	msg, err := NewMessage(payload, nil)
	if err != nil {
		return err
	}
	_, err = p.Publish(msg, opts)
	return err
}

func (p *Publisher) PublishRPCDefault(queueName, pattern string, payload any) (*RPCResponse, error) {
	opts := DefaultPublishOptions(queueName, pattern, true)
	msg, err := NewMessage(payload, nil)
	if err != nil {
		return nil, err
	}
	resp, err := p.Publish(msg, opts)
	if err != nil {
		return nil, err
	}

	response, err := convert.JSONToStruct[RPCResponse](resp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse RPC response: %w", err)
	}
	return response, err
}

func (p *Publisher) PublishWithContext(ctx context.Context, msg *Message, opts *PublishOptions) (interface{}, error) {
	if opts.MaxRetries == 0 {
		opts.MaxRetries = p.maxRetries
	}
	if opts.RetryBackoff == 0 {
		opts.RetryBackoff = p.retryInterval
	}

	var lastErr error

	for attempt := 0; attempt <= opts.MaxRetries; attempt++ {
		if attempt > 0 {
			if err := p.waitForRetry(ctx, opts, attempt); err != nil {
				return nil, err
			}
		}

		if opts.QueueName != "" {
			if err := p.ensureQueue(opts.QueueName, opts.QueueOpts); err != nil {
				lastErr = err
				logger.Warning.Printf("Failed to declare queue on attempt %d: %v\n", attempt, err)
				continue
			}
		}

		if opts.IsRPC {
			if result, err := p.handleDirectReplyToRPC(ctx, msg, opts); err != nil {
				lastErr = err
				logger.Warning.Printf("Failed to publish RPC message on attempt %d: %v\n", attempt, err)
				continue
			} else {
				return result, nil
			}
		} else {
			payload := msg.GeneratePayload(opts.Pattern)
			if err := p.publishMessage(ctx, opts, payload); err != nil {
				lastErr = err
				logger.Warning.Printf("Failed to publish message on attempt %d: %v\n", attempt, err)
				continue
			}
			return nil, nil //nolint:nilnil // fire-and-forget publish has no confirmation to return
		}
	}

	return nil, fmt.Errorf("failed to publish message after %d attempts: %w", opts.MaxRetries, lastErr)
}

func (p *Publisher) waitForRetry(ctx context.Context, opts *PublishOptions, attempt int) error {
	select {
	case <-ctx.Done():
		return fmt.Errorf("context canceled during retry: %w", ctx.Err())
	case <-time.After(opts.RetryBackoff * time.Duration(attempt)):
	}
	return nil
}

func (p *Publisher) publishMessage(ctx context.Context, opts *PublishOptions, payload *amqp.Publishing) error {
	ch, err := p.channelManager.GetChannel()
	if err != nil || ch == nil {
		return fmt.Errorf("failed to get healthy channel: %w", err)
	}

	// Check channel health before publishing
	if ch.IsClosed() {
		return fmt.Errorf("channel is closed")
	}

	err = ch.PublishWithContext(
		ctx,
		opts.Exchange,
		opts.QueueName,
		opts.Mandatory,
		opts.Immediate,
		*payload,
	)
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	return nil
}

func (p *Publisher) handleDirectReplyToRPC(ctx context.Context, msg *Message, opts *PublishOptions) (interface{}, error) {
	responseCh := make(chan *RPCResponse, 1)
	p.rpcResponsesMu.Lock()
	p.rpcResponses[msg.ID] = responseCh
	p.rpcResponsesMu.Unlock()

	defer func() {
		p.rpcResponsesMu.Lock()
		delete(p.rpcResponses, msg.ID)
		close(responseCh)
		p.rpcResponsesMu.Unlock()
	}()

	if err := p.ensureRPCConsumer(); err != nil {
		return nil, fmt.Errorf("failed to ensure RPC consumer: %w", err)
	}

	payload := msg.GenerateRPCPayload(opts.Pattern)

	if err := p.publishMessage(ctx, opts, payload); err != nil {
		return nil, fmt.Errorf("failed to publish RPC message: %w", err)
	}

	select {
	case response := <-responseCh:
		return response, nil
	case <-ctx.Done():
		return nil, fmt.Errorf("RPC request timeout: %w", ctx.Err())
	}
}

func (p *Publisher) ensureRPCConsumer() error {
	p.rpcConsumerMu.Lock()
	defer p.rpcConsumerMu.Unlock()

	if p.rpcConsumerActive {
		return nil
	}

	return p.startRPCConsumerUnsafe()
}

func (p *Publisher) startRPCConsumer() error {
	p.rpcConsumerMu.Lock()
	defer p.rpcConsumerMu.Unlock()

	return p.startRPCConsumerUnsafe()
}

func (p *Publisher) startRPCConsumerUnsafe() error {
	ch, err := p.channelManager.GetChannel()
	if err != nil {
		return fmt.Errorf("failed to get healthy channel for RPC consumer: %w", err)
	}

	if ch.IsClosed() {
		return fmt.Errorf("channel is closed, cannot start RPC consumer")
	}

	if err := ch.Qos(10, 0, false); err != nil {
		return fmt.Errorf("failed to set QoS: %w", err)
	}

	msgs, err := ch.ConsumeWithContext(
		p.ctx,
		"amq.rabbitmq.reply-to",
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to start RPC reply consumer: %w", err)
	}

	p.rpcConsumerActive = true
	go p.handleRPCReplies(msgs)
	return nil
}

func (p *Publisher) handleRPCReplies(msgs <-chan amqp.Delivery) {
	defer func() {
		p.rpcConsumerMu.Lock()
		p.rpcConsumerActive = false
		p.rpcConsumerMu.Unlock()
		logger.Info.Println("RPC consumer handler exited, consumer marked as inactive")
	}()

	for {
		select {
		case msg, ok := <-msgs:
			if !ok {
				logger.Warning.Println("RPC reply consumer channel closed")
				return
			}
			p.processRPCReply(msg)
		case <-p.ctx.Done():
			return
		}
	}
}

func (p *Publisher) processRPCReply(msg amqp.Delivery) {
	correlationID := msg.CorrelationId
	if correlationID == "" {
		logger.Warning.Println("Received RPC reply without correlation ID")
		return
	}

	var response RPCResponse
	if err := json.Unmarshal(msg.Body, &response); err != nil {
		logger.Error.Printf("Failed to parse RPC response: %v", err)
		response = RPCResponse{
			Status: "error",
			Err:    fmt.Sprintf("Failed to parse response: %v", err),
		}
	}

	p.rpcResponsesMu.RLock()
	ch, exists := p.rpcResponses[correlationID]
	p.rpcResponsesMu.RUnlock()

	if exists {
		select {
		case ch <- &response:
		default:
			logger.Warning.Printf("RPC response channel blocked for correlation ID: %s", correlationID)
		}
	} else {
		logger.Warning.Printf("No pending RPC request for correlation ID: %s", correlationID)
	}
}

func (p *Publisher) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if err := p.channelManager.Close(); err != nil {
		return fmt.Errorf("failed to close channel: %w", err)
	}

	p.rpcResponsesMu.Lock()
	for correlationID, ch := range p.rpcResponses {
		close(ch)
		delete(p.rpcResponses, correlationID)
	}
	p.rpcResponsesMu.Unlock()

	p.cancel()

	return nil
}
