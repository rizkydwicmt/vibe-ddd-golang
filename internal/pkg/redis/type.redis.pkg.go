package redis

import (
	"context"
	"sync"
	"time"

	_redis "github.com/redis/go-redis/v9"
)

type Config struct {
	Host                string
	Username            string
	Port                int
	Password            string
	PoolSize            int
	DB                  int
	Prefix              string
	DialTimeout         time.Duration `default:"5s"`
	ReconnectInterval   time.Duration `default:"1s"`
	HealthCheckInterval time.Duration `default:"3s"`
}

type ConnectionState int

const (
	StateDisconnected ConnectionState = iota
	StateConnecting
	StateConnected
	StateReconnecting
)

func (s ConnectionState) String() string {
	switch s {
	case StateDisconnected:
		return "disconnected"
	case StateConnecting:
		return "connecting"
	case StateConnected:
		return "connected"
	case StateReconnecting:
		return "reconnecting"
	default:
		return "unknown"
	}
}

type Client struct {
	Client              *_redis.Client
	config              *Config
	cancel              context.CancelFunc
	ctx                 context.Context
	state               ConnectionState
	stateMu             sync.RWMutex
	reconnectChan       chan struct{}
	healthCheckChan     chan struct{}
	isShuttingDown      bool
	shutdownMu          sync.RWMutex
	onConnected         []func()
	onDisconnected      []func()
	onReconnecting      []func()
	callbackMu          sync.RWMutex
	lastError           error
	lastErrorMu         sync.RWMutex
	connectionAttempts  int
	connectionAttemptMu sync.RWMutex
}

type IRedis interface {
	//base
	IsHealthy() bool
	GetState() ConnectionState
	GetLastError() error
	OnConnected(callback func())
	OnDisconnected(callback func())
	OnReconnecting(callback func())
	WaitForConnection(timeout time.Duration) error
	Close() error
	//action
	Set(key string, value any, expiration time.Duration) error
	SetNX(key string, value string, ttl time.Duration) (bool, error)
	Get(key string) (string, error)
	Get2(key string, output any) error
	GetKeys(pattern string) ([]string, error)
	GenerateKey(key string) string
	Del(key string) error
	Expire(key string, expiration time.Duration) error
	Ping(ctx context.Context) error
	//paginator
	GetPageOffset(key string, page, pageSize int) (*PaginationResult, error)
	GetAll(key string) (*PaginationResult, error)
	AddData(key string, data []map[string]any, ttl *time.Duration) error
	AppendData(key string, data []map[string]any, ttl *time.Duration) error
	UpdateData(key string, data []map[string]any) error
	RemoveDataByIDs(key string, ids []string) error
	RemoveDataByScore(key string, minScore, maxScore float64) (int64, error)
	RemoveDataByRank(key string, start, stop int64) (int64, error)
	GetById(key string, id string) (map[string]any, error)
	GetByScore(key string, minScore, maxScore float64) ([]map[string]any, error)
	//user
	SAdd(key, member string) (int64, error)
	SRem(key, member string) (int64, error)
	SCard(key string) (int64, error)
	SMembers(key string) ([]string, error)
	//sorted set (member + score; score doubles as a timestamp index)
	ZAdd(key string, score float64, member string) (int64, error)
	ZRem(key, member string) (int64, error)
	ZRange(key string, start, stop int64) ([]string, error)
}

type ClientType = _redis.Client

const NilType = _redis.Nil
