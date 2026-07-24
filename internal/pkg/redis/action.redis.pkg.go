package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	_redis "github.com/redis/go-redis/v9"
)

// Set stores a key-value pair with an expiration time.
func (r *Client) Set(key string, value any, expiration time.Duration) error {
	var data string
	switch v := value.(type) {
	case string:
		data = v
	case []byte:
		data = string(v)
	default:
		res, err := json.Marshal(value)
		if err != nil {
			return err
		}
		data = string(res)
	}
	return r.executeWithRetry(func() error {
		return r.Client.Set(r.ctx, r.config.Prefix+key, data, expiration).Err()
	}, "SET")
}

func (c *Client) SetNX(key, value string, ttl time.Duration) (bool, error) {
	var won bool
	err := c.executeWithRetry(func() error {
		v, err := c.Client.SetNX(c.ctx, c.config.Prefix+key, value, ttl).Result()
		if err != nil {
			return err
		}
		won = v
		return nil
	}, "SETNX")
	return won, err
}

// Get retrieves the value of a key.
// Example usage:
//
// value, err := redisClient.Get("myKey") #output is string
// valueParsed, err := convert.StringToStruct[SampleStruct](value) #output is struct
func (r *Client) Get(key string) (string, error) {
	var result string
	err := r.executeWithRetry(func() error {
		res, err := r.Client.Get(r.ctx, r.config.Prefix+key).Result()
		if err != nil {
			if errors.Is(err, NilType) {
				result = ""
				return nil // Key does not exist, not an error
			}
			return err
		}
		result = res
		return nil
	}, "GET")

	return result, err
}

// Get2 retrieves the value of a key.
// Example usage:
//
//	var valueParsed SampleStruct
//	err := redisClient.Get2("myKey", &valueParsed) #output is struct
func (r *Client) Get2(key string, output any) error {
	var result string
	err := r.executeWithRetry(func() error {
		res, err := r.Client.Get(r.ctx, r.config.Prefix+key).Result()
		if err != nil {
			if errors.Is(err, NilType) {
				result = ""
				return nil // Key does not exist, not an error
			}
			return err
		}
		result = res
		return nil
	}, "GET2")

	if err != nil {
		return err
	}

	if output != nil {
		err := json.Unmarshal([]byte(result), output)
		if err != nil {
			return fmt.Errorf("failed to unmarshal value for key %s: %w", key, err)
		}
	}
	return nil
}

func (r *Client) GetKeys(key string) ([]string, error) {
	var result []string
	err := r.executeWithRetry(func() error {
		res, err := r.Client.Keys(r.ctx, r.config.Prefix+key).Result()
		if err != nil {
			if errors.Is(err, NilType) {
				return nil // Key does not exist, not an error
			}
			return err
		}
		result = res
		return nil
	}, "KEYS")

	return result, err
}

// Del deletes a key from IRedis.
func (r *Client) Del(key string) error {
	err := r.executeWithRetry(func() error {
		err := r.Client.Del(r.ctx, r.config.Prefix+key).Err()
		if err != nil {
			return err
		}
		return nil
	}, "DEL")
	return err
}

// Expire sets a timeout on a key.
func (r *Client) Expire(key string, expiration time.Duration) error {
	return r.executeWithRetry(func() error {
		err := r.Client.Expire(r.ctx, r.config.Prefix+key, expiration).Err()
		if err != nil {
			return err
		}
		return nil
	}, "EXPIRE")
}

// SAdd adds a member to a set (used to track live sessions per user).
func (r *Client) SAdd(key, member string) (int64, error) {
	var n int64
	err := r.executeWithRetry(func() error {
		v, err := r.Client.SAdd(r.ctx, r.config.Prefix+key, member).Result()
		if err != nil {
			return err
		}
		n = v
		return nil
	}, "SADD")
	return n, err
}

// SRem removes a member from a set.
func (r *Client) SRem(key, member string) (int64, error) {
	var n int64
	err := r.executeWithRetry(func() error {
		v, err := r.Client.SRem(r.ctx, r.config.Prefix+key, member).Result()
		if err != nil {
			return err
		}
		n = v
		return nil
	}, "SREM")
	return n, err
}

// SCard returns set cardinality.
func (r *Client) SCard(key string) (int64, error) {
	var n int64
	err := r.executeWithRetry(func() error {
		v, err := r.Client.SCard(r.ctx, r.config.Prefix+key).Result()
		if err != nil {
			return err
		}
		n = v
		return nil
	}, "SCARD")
	return n, err
}

// SMembers returns all members of a set (used to enumerate a user's active
// session ids for revocation).
func (r *Client) SMembers(key string) ([]string, error) {
	var members []string
	err := r.executeWithRetry(func() error {
		v, err := r.Client.SMembers(r.ctx, r.config.Prefix+key).Result()
		if err != nil {
			return err
		}
		members = v
		return nil
	}, "SMEMBERS")
	return members, err
}

// ZAdd adds/updates a member in a sorted set with the given score. Re-adding an
// existing member overwrites its score (used to track live sessions per user with
// the score carrying the session timestamp).
func (r *Client) ZAdd(key string, score float64, member string) (int64, error) {
	var n int64
	err := r.executeWithRetry(func() error {
		v, err := r.Client.ZAdd(r.ctx, r.config.Prefix+key, _redis.Z{Score: score, Member: member}).Result()
		if err != nil {
			return err
		}
		n = v
		return nil
	}, "ZADD")
	return n, err
}

// ZRem removes a member from a sorted set.
func (r *Client) ZRem(key, member string) (int64, error) {
	var n int64
	err := r.executeWithRetry(func() error {
		v, err := r.Client.ZRem(r.ctx, r.config.Prefix+key, member).Result()
		if err != nil {
			return err
		}
		n = v
		return nil
	}, "ZREM")
	return n, err
}

// ZRange returns members in score order (ascending) for the given rank range;
// use start=0, stop=-1 to enumerate all members (e.g. for revocation).
func (r *Client) ZRange(key string, start, stop int64) ([]string, error) {
	var members []string
	err := r.executeWithRetry(func() error {
		v, err := r.Client.ZRange(r.ctx, r.config.Prefix+key, start, stop).Result()
		if err != nil {
			return err
		}
		members = v
		return nil
	}, "ZRANGE")
	return members, err
}

func (r *Client) GenerateKey(key string) string {
	return r.config.Prefix + key
}

// Ping checks liveness against the supplied context (used by readiness probes).
func (r *Client) Ping(ctx context.Context) error {
	return r.Client.Ping(ctx).Err()
}
