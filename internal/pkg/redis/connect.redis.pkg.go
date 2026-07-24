package redis

import (
	"context"
	"fmt"
	"time"

	"vibe-ddd-golang/internal/pkg/logger"

	_redis "github.com/redis/go-redis/v9"
)

func Setup(ctx context.Context, config *Config) (*Client, error) {
	clientCtx, cancel := context.WithCancel(ctx)

	r := &Client{
		cancel:          cancel,
		ctx:             clientCtx,
		config:          defaultConfig(config),
		state:           StateDisconnected,
		reconnectChan:   make(chan struct{}, 1),
		healthCheckChan: make(chan struct{}, 1),
	}

	// Initial connection
	if err := r.connect(); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to initial connect to redis: %w", err)
	}

	// Start monitoring goroutines
	go r.healthCheckLoop()
	go r.reconnectionLoop()

	return r, nil
}

func defaultConfig(cfg *Config) *Config {
	if cfg.ReconnectInterval == 0 {
		cfg.ReconnectInterval = 1 * time.Second
	}
	if cfg.HealthCheckInterval == 0 {
		cfg.HealthCheckInterval = 3 * time.Second
	}
	if cfg.DialTimeout == 0 {
		cfg.DialTimeout = 5 * time.Second
	}
	return cfg
}

func (r *Client) setState(state ConnectionState) {
	r.stateMu.Lock()
	oldState := r.state
	r.state = state
	r.stateMu.Unlock()

	if oldState != state {
		r.triggerCallbacks(state)
	}
}

func (r *Client) getState() ConnectionState {
	r.stateMu.RLock()
	defer r.stateMu.RUnlock()
	return r.state
}

func (r *Client) setLastError(err error) {
	r.lastErrorMu.Lock()
	r.lastError = err
	r.lastErrorMu.Unlock()
}

func (r *Client) incrementConnectionAttempts() int {
	r.connectionAttemptMu.Lock()
	defer r.connectionAttemptMu.Unlock()
	r.connectionAttempts++
	return r.connectionAttempts
}

func (r *Client) resetConnectionAttempts() {
	r.connectionAttemptMu.Lock()
	r.connectionAttempts = 0
	r.connectionAttemptMu.Unlock()
}

func (r *Client) connect() error {
	r.setState(StateConnecting)

	if r.Client != nil {
		_ = r.Client.Close()
	}

	r.Client = _redis.NewClient(&_redis.Options{
		Addr:        fmt.Sprintf("%s:%d", r.config.Host, r.config.Port),
		Username:    r.config.Username,
		Password:    r.config.Password,
		PoolSize:    r.config.PoolSize,
		DB:          r.config.DB,
		DialTimeout: r.config.DialTimeout,
	})

	ctx, cancel := context.WithTimeout(r.ctx, r.config.DialTimeout)
	defer cancel()

	if err := r.Client.Ping(ctx).Err(); err != nil {
		r.setState(StateDisconnected)
		r.setLastError(err)
		return fmt.Errorf("failed to ping redis: %w", err)
	}

	r.setState(StateConnected)
	r.setLastError(nil)
	r.resetConnectionAttempts()

	if err := r.WaitForConnection(5 * time.Second); err != nil {
		return fmt.Errorf("redis connection not stable: %w", err)
	}

	return nil
}

func (r *Client) isShuttingDownFlag() bool {
	r.shutdownMu.RLock()
	defer r.shutdownMu.RUnlock()
	return r.isShuttingDown
}

func (r *Client) healthCheckLoop() {
	ticker := time.NewTicker(r.config.HealthCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-r.ctx.Done():
			logger.Debug.Println("Health check loop shutting down...")
			return
		case <-ticker.C:
			if r.isShuttingDownFlag() {
				return
			}

			if r.getState() == StateConnected {
				ctx, cancel := context.WithTimeout(r.ctx, 3*time.Second)
				err := r.Client.Ping(ctx).Err()
				cancel()

				if err != nil {
					logger.Warning.Printf("Redis health check failed: %v", err)
					r.setState(StateDisconnected)
					r.setLastError(err)
					r.triggerReconnect()
				}
			}
		}
	}
}

func (r *Client) reconnectionLoop() {
	for {
		select {
		case <-r.ctx.Done():
			logger.Debug.Println("Reconnection loop shutting down...")
			return
		case <-r.reconnectChan:
			if r.isShuttingDownFlag() {
				return
			}

			r.performReconnection()
		}
	}
}

func (r *Client) performReconnection() {
	currentState := r.getState()
	if currentState == StateConnected || currentState == StateConnecting {
		return
	}

	r.setState(StateReconnecting)

	for {
		if r.isShuttingDownFlag() {
			return
		}

		attempts := r.incrementConnectionAttempts()
		logger.Warning.Printf("Redis reconnection attempt #%d", attempts)

		if err := r.connect(); err != nil {
			logger.Warning.Printf("Reconnection attempt #%d #%s failed: %v", attempts, r.getState(), err)

			select {
			case <-r.ctx.Done():
				return
			case <-time.After(r.config.ReconnectInterval):
				continue
			}
		} else {
			logger.Info.Printf("Redis reconnection successful after %d attempts", attempts)
			return
		}
	}
}

func (r *Client) triggerReconnect() {
	currentState := r.getState()
	if currentState == StateReconnecting || currentState == StateConnecting {
		return
	}

	if currentState == StateConnected {
		ctx, cancel := context.WithTimeout(r.ctx, r.config.DialTimeout)
		defer cancel()
		if pingErr := r.Client.Ping(ctx).Err(); pingErr != nil {
			r.setState(StateDisconnected)
		} else {
			return
		}
	}

	select {
	case r.reconnectChan <- struct{}{}:
	default:
	}
}

func (r *Client) triggerCallbacks(state ConnectionState) {
	r.callbackMu.RLock()
	defer r.callbackMu.RUnlock()

	switch state {
	case StateConnected:
		for _, cb := range r.onConnected {
			go cb()
		}
	case StateDisconnected:
		for _, cb := range r.onDisconnected {
			go cb()
		}
	case StateReconnecting:
		for _, cb := range r.onReconnecting {
			go cb()
		}
	}
}

func (r *Client) executeWithRetry(operation func() error, operationName string) error {
	maxAttempts := 3
	var lastErr error

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		if r.getState() != StateConnected {
			if attempt == 1 {
				r.triggerReconnect()
			}
		}

		if err := operation(); err != nil {
			lastErr = err

			if r.isConnectionError(err) {
				logger.Debug.Printf("Detected potential connection error for %s: %v", operationName, err)
				r.setState(StateDisconnected)
				r.setLastError(err)
				r.triggerReconnect()

				if attempt < maxAttempts {
					time.Sleep(time.Duration(attempt) * 100 * time.Millisecond)
				}
				continue
			}

			return fmt.Errorf("%s failed: %w", operationName, err)
		}

		return nil
	}

	return fmt.Errorf("%s failed after %d attempts: %w", operationName, maxAttempts, lastErr)
}

func (r *Client) isConnectionError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()
	connectionErrors := []string{
		"connection refused",
		"no such host",
		"broken pipe",
		"connection reset",
		"network is unreachable",
		"connection closed",
		"EOF",
	}

	for _, connErr := range connectionErrors {
		if contains(errStr, connErr) {
			return true
		}
	}

	if r.Client != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		if pingErr := r.Client.Ping(ctx).Err(); pingErr != nil {
			return true
		}
	}

	return false
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			indexOf(s, substr) >= 0))
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func (r *Client) Close() error {
	r.shutdownMu.Lock()
	r.isShuttingDown = true
	r.shutdownMu.Unlock()
	r.cancel()

	if r.Client != nil {
		return r.Client.Close()
	}
	return nil
}

func (r *Client) IsHealthy() bool {
	if r.getState() != StateConnected || r.Client == nil {
		return false
	}

	ctx, cancel := context.WithTimeout(r.ctx, 2*time.Second)
	defer cancel()

	if err := r.Client.Ping(ctx).Err(); err != nil {
		return false
	}

	return true
}

func (r *Client) GetState() ConnectionState {
	return r.getState()
}

func (r *Client) GetLastError() error {
	r.lastErrorMu.RLock()
	defer r.lastErrorMu.RUnlock()
	return r.lastError
}

func (r *Client) OnConnected(callback func()) {
	r.callbackMu.Lock()
	r.onConnected = append(r.onConnected, callback)
	r.callbackMu.Unlock()
}

func (r *Client) OnDisconnected(callback func()) {
	r.callbackMu.Lock()
	r.onDisconnected = append(r.onDisconnected, callback)
	r.callbackMu.Unlock()
}

func (r *Client) OnReconnecting(callback func()) {
	r.callbackMu.Lock()
	r.onReconnecting = append(r.onReconnecting, callback)
	r.callbackMu.Unlock()
}

func (r *Client) WaitForConnection(timeout time.Duration) error {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		if r.IsHealthy() {
			return nil
		}

		select {
		case <-r.ctx.Done():
			return fmt.Errorf("context canceled while waiting for connection")
		case <-time.After(100 * time.Millisecond):
			continue
		}
	}

	return fmt.Errorf("connection not ready within timeout: %v", timeout)
}
