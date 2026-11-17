// Package redis implements the Redis driver for DataPulse.
package redis

import (
	"context"
	"fmt"
	"strconv"

	"github.com/collins/DataPulse/internal/driver"
	"github.com/collins/DataPulse/internal/models"
	"github.com/redis/go-redis/v9"
)

const (
	defaultHost = "localhost"
	defaultPort = 6379
)

// Driver opens Redis connections.
type Driver struct{}

// NewDriver returns a Redis driver.
func NewDriver() *Driver {
	return &Driver{}
}

// Type returns the Redis driver type.
func (d *Driver) Type() driver.Type {
	return driver.Redis
}

// Open opens a Redis connection from a profile.
func (d *Driver) Open(ctx context.Context, profile *models.Profile) (driver.Conn, error) {
	host := profile.Host
	if host == "" {
		host = defaultHost
	}
	port := profile.Port
	if port == 0 {
		port = defaultPort
	}
	client := redis.NewClient(&redis.Options{Addr: host + ":" + strconv.Itoa(port)})
	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("redis: ping connection: %w", err)
	}
	return &redisConn{client: client}, nil
}

// redisConn adapts a Redis client to the driver connection port.
type redisConn struct {
	client *redis.Client
}

// Close closes the Redis connection.
func (c *redisConn) Close() error {
	return c.client.Close()
}

// Catalog returns nil because Redis is not catalog driven.
func (c *redisConn) Catalog() driver.Catalog {
	return nil
}

// KV returns the key-value access port for Redis.
func (c *redisConn) KV() driver.KV {
	return &redisKV{client: c.client}
}

// redisKV implements driver.KV for Redis.
type redisKV struct {
	client *redis.Client
}

// ScanKeys lists keys matching a pattern using cursor-based scan.
func (k *redisKV) ScanKeys(ctx context.Context, pattern string, cursor uint64, count int64) ([]string, uint64, error) {
	cmd := k.client.Scan(ctx, cursor, pattern, count)
	keys, next, err := cmd.Result()
	if err != nil {
		return nil, 0, err
	}
	return keys, next, nil
}

// KeyInfo returns metadata for a single key.
func (k *redisKV) KeyInfo(ctx context.Context, key string) (models.KeyInfo, error) {
	keyType, err := k.client.Type(ctx, key).Result()
	if err != nil {
		return models.KeyInfo{}, err
	}
	ttl, err := k.client.TTL(ctx, key).Result()
	if err != nil {
		return models.KeyInfo{}, err
	}
	length, err := k.client.MemoryUsage(ctx, key).Result()
	if err != nil {
		return models.KeyInfo{}, err
	}
	return models.KeyInfo{Key: key, Type: keyType, TTL: int64(ttl.Seconds()), Length: length}, nil
}

// GetValue returns a single key value.
func (k *redisKV) GetValue(ctx context.Context, key string) (models.ValueView, error) {
	value, err := k.client.Get(ctx, key).Result()
	if err != nil {
		return models.ValueView{}, err
	}
	keyType, err := k.client.Type(ctx, key).Result()
	if err != nil {
		return models.ValueView{}, err
	}
	return models.ValueView{Key: key, Value: value, Kind: keyType}, nil
}

// Exec runs a raw Redis command and returns its array reply.
func (k *redisKV) Exec(ctx context.Context, args []string) ([]string, error) {
	argv := make([]interface{}, len(args))
	for i, arg := range args {
		argv[i] = arg
	}
	cmd := k.client.Do(ctx, argv...)
	result, err := cmd.Result()
	if err != nil {
		return nil, err
	}
	arr, ok := result.([]interface{})
	if !ok {
		return nil, fmt.Errorf("redis: unexpected exec result type %T", result)
	}
	out := make([]string, 0, len(arr))
	for _, v := range arr {
		s, ok := v.(string)
		if !ok {
			return nil, fmt.Errorf("redis: unexpected exec result element type %T", v)
		}
		out = append(out, s)
	}
	return out, nil
}
