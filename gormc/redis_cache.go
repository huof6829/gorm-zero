package gormc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	// ErrCacheMiss indicates the key is not found in cache.
	ErrCacheMiss = errors.New("cache: key not found")
)

// RedisConfig is the redis configuration.
type RedisConfig struct {
	Addr         string        // Redis server address
	Password     string        // Redis password
	DB           int           // Redis database index
	PoolSize     int           // Connection pool size
	MinIdleConns int           // Minimum idle connections
	DialTimeout  time.Duration // Dial timeout
	ReadTimeout  time.Duration // Read timeout
	WriteTimeout time.Duration // Write timeout
}

// RedisCache is a cache implementation based on native go-redis.
type RedisCache struct {
	client        *redis.Client
	notFoundError error
	expiry        time.Duration
}

// NewRedisCache creates a new RedisCache instance.
func NewRedisCache(conf RedisConfig, expiry time.Duration) (*RedisCache, error) {
	if conf.DialTimeout == 0 {
		conf.DialTimeout = 5 * time.Second
	}
	if conf.ReadTimeout == 0 {
		conf.ReadTimeout = 3 * time.Second
	}
	if conf.WriteTimeout == 0 {
		conf.WriteTimeout = 3 * time.Second
	}
	if conf.PoolSize == 0 {
		conf.PoolSize = 10
	}
	if conf.MinIdleConns == 0 {
		conf.MinIdleConns = 2
	}

	client := redis.NewClient(&redis.Options{
		Addr:         conf.Addr,
		Password:     conf.Password,
		DB:           conf.DB,
		PoolSize:     conf.PoolSize,
		MinIdleConns: conf.MinIdleConns,
		DialTimeout:  conf.DialTimeout,
		ReadTimeout:  conf.ReadTimeout,
		WriteTimeout: conf.WriteTimeout,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return &RedisCache{
		client:        client,
		notFoundError: ErrNotFound,
		expiry:        expiry,
	}, nil
}

// DelCtx deletes cached values with keys.
func (c *RedisCache) DelCtx(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}
	return c.client.Del(ctx, keys...).Err()
}

// GetCtx unmarshals cache with given key into v.
func (c *RedisCache) GetCtx(ctx context.Context, key string, v interface{}) error {
	data, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return ErrCacheMiss
		}
		return err
	}

	if len(data) == 0 {
		return ErrCacheMiss
	}

	return json.Unmarshal(data, v)
}

// SetCtx sets cache with given key and value.
func (c *RedisCache) SetCtx(ctx context.Context, key string, v interface{}) error {
	return c.SetWithExpireCtx(ctx, key, v, c.expiry)
}

// SetWithExpireCtx sets cache with given key, value and expire time.
func (c *RedisCache) SetWithExpireCtx(ctx context.Context, key string, v interface{}, expire time.Duration) error {
	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	return c.client.Set(ctx, key, data, expire).Err()
}

// TakeCtx takes the result from cache first, if not found,
// query from the query function and set cache with the result.
func (c *RedisCache) TakeCtx(ctx context.Context, v interface{}, key string, query func(v interface{}) error) error {
	return c.TakeWithExpireCtx(ctx, v, key, query, c.expiry)
}

// TakeWithExpireCtx takes the result from cache first, if not found,
// query from the query function and set cache with the result with given expire time.
func (c *RedisCache) TakeWithExpireCtx(ctx context.Context, v interface{}, key string, query func(v interface{}) error, expire time.Duration) error {
	err := c.GetCtx(ctx, key, v)
	if err == nil {
		return nil
	}

	if !errors.Is(err, ErrCacheMiss) && !errors.Is(err, redis.Nil) {
		return err
	}

	// Query from database
	if err := query(v); err != nil {
		return err
	}

	// Set cache with the result
	if err := c.SetWithExpireCtx(ctx, key, v, expire); err != nil {
		// Log error but don't fail the request
		// You might want to add proper logging here
		_ = err
	}

	return nil
}

// Close closes the redis client.
func (c *RedisCache) Close() error {
	return c.client.Close()
}

// GetClient returns the underlying redis client.
func (c *RedisCache) GetClient() *redis.Client {
	return c.client
}
