package rabbitmq

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Event is the typed envelope used by the default event helpers.
type Event[T any] struct {
	ID       string
	Pattern  string
	Data     T
	Headers  amqp.Table
	Body     []byte
	Delivery *amqp.Delivery
}

type EventHandler[T any] func(ctx context.Context, event *Event[T]) error

type EventSubscribeConfig struct {
	Options      *SubscribeOptions
	Callback     CallbackHandler
	Patterns     map[string]struct{}
	Exchange     string
	ExchangeType string
}

type EventSubscribeOption func(*EventSubscribeConfig)

func DefaultEventSubscribeConfig(queueName string) *EventSubscribeConfig {
	return &EventSubscribeConfig{
		Options:      DefaultSubscribeOptions(queueName, false),
		ExchangeType: "topic",
	}
}

func WithEventSubscribeOptions(opts *SubscribeOptions) EventSubscribeOption {
	return func(cfg *EventSubscribeConfig) {
		cfg.Options = cloneSubscribeOptions(opts)
	}
}

func WithEventSubscribeCallback(callback CallbackHandler) EventSubscribeOption {
	return func(cfg *EventSubscribeConfig) {
		cfg.Callback = callback
	}
}

func WithEventSubscribeWorkers(workerCount, prefetchCount int) EventSubscribeOption {
	return func(cfg *EventSubscribeConfig) {
		if cfg.Options == nil {
			cfg.Options = DefaultSubscribeOptions("", false)
		}
		if workerCount > 0 {
			cfg.Options.WorkerCount = workerCount
		}
		if prefetchCount > 0 {
			cfg.Options.PrefetchCount = prefetchCount
		}
	}
}

func WithEventSubscribeRetry(maxAttempts int, strategy RetryStrategy, baseDelay, maxDelay time.Duration) EventSubscribeOption {
	return func(cfg *EventSubscribeConfig) {
		if cfg.Options == nil {
			cfg.Options = DefaultSubscribeOptions("", false)
		}
		if maxAttempts >= 0 {
			cfg.Options.MaxRetryAttempts = maxAttempts
		}
		if strategy != "" {
			cfg.Options.RetryStrategy = strategy
		}
		if baseDelay > 0 {
			cfg.Options.BaseRetryDelay = baseDelay
		}
		if maxDelay > 0 {
			cfg.Options.MaxRetryDelay = maxDelay
		}
	}
}

func WithEventSubscribeDeadLetter(enabled bool, deadLetterName string) EventSubscribeOption {
	return func(cfg *EventSubscribeConfig) {
		if cfg.Options == nil {
			cfg.Options = DefaultSubscribeOptions("", false)
		}
		cfg.Options.EnableDeadLetter = enabled
		if deadLetterName != "" {
			cfg.Options.DeadLetterName = deadLetterName
		}
	}
}

func WithEventSubscribeConsumerName(name string) EventSubscribeOption {
	return func(cfg *EventSubscribeConfig) {
		if cfg.Options == nil {
			cfg.Options = DefaultSubscribeOptions("", false)
		}
		if name != "" {
			cfg.Options.ConsumerName = name
		}
	}
}

func WithEventSubscribePatterns(patterns ...string) EventSubscribeOption {
	return func(cfg *EventSubscribeConfig) {
		if len(patterns) == 0 {
			return
		}
		cfg.Patterns = make(map[string]struct{}, len(patterns))
		for _, pattern := range patterns {
			if pattern != "" {
				cfg.Patterns[pattern] = struct{}{}
			}
		}
	}
}

func WithEventSubscribeExchange(exchange, exchangeType string) EventSubscribeOption {
	return func(cfg *EventSubscribeConfig) {
		cfg.Exchange = exchange
		if exchangeType != "" {
			cfg.ExchangeType = exchangeType
		}
	}
}

func NewEventSubscriberDefault[T any](
	ctx context.Context, connManager *ConnectionManager, queueName string,
	handler EventHandler[T], options ...EventSubscribeOption,
) (*Subscriber, error) {
	if connManager == nil {
		return nil, errors.New("connection manager is nil")
	}
	if queueName == "" {
		return nil, errors.New("queue name is required")
	}
	if handler == nil {
		return nil, errors.New("event handler is required")
	}
	if ctx == nil {
		ctx = context.Background()
	}

	cfg := buildEventSubscribeConfig(queueName, options...)
	opts := normalizeSubscribeOptions(queueName, cfg.Options)
	patterns := cfg.Patterns

	if cfg.Exchange != "" {
		if err := bindEventQueue(connManager, opts, cfg.Exchange, cfg.ExchangeType, patterns); err != nil {
			return nil, err
		}
	}

	messageHandler := func(msg *amqp.Delivery) (interface{}, error) {
		event, err := DecodeEvent[T](msg)
		if err != nil {
			return nil, err
		}
		if len(patterns) > 0 {
			if _, ok := patterns[event.Pattern]; !ok {
				return event, fmt.Errorf("unexpected event pattern %q", event.Pattern)
			}
		}
		if err := handler(ctx, event); err != nil {
			return event, err
		}
		return event, nil
	}

	return NewSubscriber(ctx, connManager, messageHandler, cfg.Callback, opts)
}

func SubscribeEventDefault[T any](
	ctx context.Context, connManager *ConnectionManager, queueName string,
	handler EventHandler[T], options ...EventSubscribeOption,
) (*Subscriber, error) {
	sub, err := NewEventSubscriberDefault(ctx, connManager, queueName, handler, options...)
	if err != nil {
		return nil, err
	}
	if err := sub.Start(); err != nil {
		_ = sub.Stop()
		return nil, err
	}
	return sub, nil
}

func DecodeEvent[T any](msg *amqp.Delivery) (*Event[T], error) {
	if msg == nil {
		return nil, errors.New("delivery is nil")
	}

	headers := cloneTable(msg.Headers)
	id := msg.MessageId
	if id == "" {
		id = headerString(headers, "id")
	}
	pattern := msg.Type
	if pattern == "" {
		pattern = headerString(headers, "type")
	}
	if pattern == "" {
		pattern = headerString(headers, "pattern")
	}

	var data T
	var envelope struct {
		Pattern string          `json:"type"`
		Data    json.RawMessage `json:"data"`
		ID      string          `json:"id"`
	}

	if len(msg.Body) > 0 {
		if err := json.Unmarshal(msg.Body, &envelope); err == nil && (envelope.ID != "" || envelope.Pattern != "" || envelope.Data != nil) {
			if envelope.ID != "" {
				id = envelope.ID
			}
			if envelope.Pattern != "" {
				pattern = envelope.Pattern
			}
			if envelope.Data != nil {
				decoded, err := decodeEventData[T](envelope.Data)
				if err != nil {
					return nil, fmt.Errorf("failed to decode event data: %w", err)
				}
				data = decoded
			}
		} else {
			decoded, err := decodeEventData[T](msg.Body)
			if err != nil {
				return nil, fmt.Errorf("failed to decode event body: %w", err)
			}
			data = decoded
		}
	}

	body := make([]byte, len(msg.Body))
	copy(body, msg.Body)

	return &Event[T]{
		ID:       id,
		Pattern:  pattern,
		Data:     data,
		Headers:  headers,
		Body:     body,
		Delivery: msg,
	}, nil
}

func bindEventQueue(
	connManager *ConnectionManager, opts *SubscribeOptions,
	exchange, exchangeType string, patterns map[string]struct{},
) error {
	conn := connManager.GetConnection()
	if conn == nil {
		return errors.New("rabbitmq connection is not available")
	}
	ch, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open channel: %w", err)
	}
	defer ch.Close()

	if exchangeType == "" {
		exchangeType = "topic"
	}
	if err := ch.ExchangeDeclare(exchange, exchangeType, true, false, false, false, nil); err != nil {
		return fmt.Errorf("failed to declare exchange: %w", err)
	}

	queueOpts := opts.QueueOpts
	if queueOpts == nil {
		queueOpts = DefaultQueueConfig()
	}
	if _, err := ch.QueueDeclare(opts.QueueName, queueOpts.Durable, queueOpts.AutoDelete,
		queueOpts.Exclusive, queueOpts.NoWait, queueOpts.Args); err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	if len(patterns) == 0 {
		patterns = map[string]struct{}{"#": {}}
	}
	for pattern := range patterns {
		if err := ch.QueueBind(opts.QueueName, pattern, exchange, false, nil); err != nil {
			return fmt.Errorf("failed to bind queue %q with pattern %q: %w", opts.QueueName, pattern, err)
		}
	}
	return nil
}

func buildEventSubscribeConfig(queueName string, options ...EventSubscribeOption) *EventSubscribeConfig {
	cfg := DefaultEventSubscribeConfig(queueName)
	for _, option := range options {
		if option != nil {
			option(cfg)
		}
	}
	return cfg
}

func normalizeSubscribeOptions(queueName string, opts *SubscribeOptions) *SubscribeOptions {
	if opts == nil {
		opts = DefaultSubscribeOptions(queueName, false)
	} else {
		opts = cloneSubscribeOptions(opts)
	}

	if opts.QueueName == "" {
		opts.QueueName = queueName
	}
	if opts.ConsumerName == "" {
		opts.ConsumerName = opts.QueueName
	}
	if opts.WorkerCount <= 0 {
		opts.WorkerCount = 3
	}
	if opts.PrefetchCount <= 0 {
		opts.PrefetchCount = 10
	}
	if opts.MessageBuffer <= 0 {
		opts.MessageBuffer = 100
	}
	if opts.DeadLetterName == "" {
		opts.DeadLetterName = "fail:" + opts.QueueName
	}
	if opts.BaseRetryDelay <= 0 {
		opts.BaseRetryDelay = 5 * time.Second
	}
	if opts.MaxRetryDelay <= 0 {
		opts.MaxRetryDelay = 10 * time.Minute
	}
	if opts.RetryStrategy == "" {
		opts.RetryStrategy = FixedRetry
	}

	return opts
}

func decodeEventData[T any](body []byte) (T, error) {
	var data T
	if len(body) == 0 {
		return data, nil
	}

	if err := json.Unmarshal(body, &data); err == nil {
		return data, nil
	} else {
		switch target := any(&data).(type) {
		case *string:
			*target = string(body)
			return data, nil
		case *[]byte:
			*target = append((*target)[:0], body...)
			return data, nil
		default:
			return data, err
		}
	}
}

func cloneSubscribeOptions(opts *SubscribeOptions) *SubscribeOptions {
	if opts == nil {
		return nil
	}
	cloned := *opts
	cloned.QueueOpts = cloneQueueConfig(opts.QueueOpts)
	return &cloned
}

func cloneQueueConfig(cfg *QueueConfig) *QueueConfig {
	if cfg == nil {
		return DefaultQueueConfig()
	}
	return &QueueConfig{
		Durable:    cfg.Durable,
		AutoDelete: cfg.AutoDelete,
		Exclusive:  cfg.Exclusive,
		NoWait:     cfg.NoWait,
		Args:       cloneTableOrNil(cfg.Args),
	}
}

func cloneTable(table amqp.Table) amqp.Table {
	if table == nil {
		return amqp.Table{}
	}
	cloned := make(amqp.Table, len(table))
	for key, value := range table {
		cloned[key] = value
	}
	return cloned
}

func cloneTableOrNil(table amqp.Table) amqp.Table {
	if table == nil {
		return nil
	}
	return cloneTable(table)
}

func headerString(headers amqp.Table, key string) string {
	value, ok := headers[key]
	if !ok {
		return ""
	}
	switch v := value.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	default:
		return fmt.Sprint(v)
	}
}
