package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"
	"weblineBackend/internal/model"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

var (
	cacheHits = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "cache_hits_total",
		Help: "Total number of cache hits",
	})
	cacheMisses = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "cache_misses_total",
		Help: "Total number of cache misses",
	})
	cacheErrors = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "cache_errors_total",
		Help: "Total number of cache errors",
	})
)

func init() {
	prometheus.MustRegister(cacheHits, cacheMisses, cacheErrors)
}

type CacheService interface {
	Set(ctx context.Context, key string, value interface{}) error
	Get(ctx context.Context, key string, dest interface{}) error
	Delete(ctx context.Context, key string) error
	HSet(ctx context.Context, key, field string, value interface{}) error
	HGet(ctx context.Context, key, field string) (string, error)
	Pipeline(ctx context.Context, fn func(redis.Pipeliner) error) error
	SAdd(ctx context.Context, key string, members ...interface{}) error
	SMembers(ctx context.Context, key string) ([]string, error)
	ZAdd(ctx context.Context, key string, members ...*redis.Z) error
	ZRange(ctx context.Context, key string, start, stop int64) ([]string, error)
	Incr(ctx context.Context, key string) (int64, error)
	Decr(ctx context.Context, key string) (int64, error)
	GetOrSet(ctx context.Context, key string, dest interface{}, fetchFunc func() error) error
	HealthCheck(ctx context.Context) error
	Initialize() error
}

type cacheService struct {
	redisClient *redis.Client
	logger      *zap.Logger
	ttl         time.Duration
	initialized bool
	mu sync.Mutex

	// Rate limiting
	rateLimiter chan struct{}
}

func NewCacheService(redisClient *redis.Client, logger *zap.Logger, ttl time.Duration, rateLimit int) CacheService {
	return &cacheService{
		redisClient: redisClient,
		logger:      logger,
		ttl:         ttl,
		rateLimiter: make(chan struct{}, rateLimit),
	}
}

// Initialize performs any necessary initialization tasks.
// It ensures that initialization happens only once (lazy initialization).
func (c *cacheService) Initialize() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.initialized {
		return nil
	}

	// Perform initialization tasks if needed (e.g., preloading certain cache entries)
	// For now, we just set the initialized flag.
	c.initialized = true
	c.logger.Info("CacheService initialized successfully")
	return nil
}

// HealthCheck verifies the connectivity to the Redis server.
func (c *cacheService) HealthCheck(ctx context.Context) error {
	err := c.redisClient.Ping(ctx).Err()
	if err != nil {
		c.logger.Error("Cache health check failed", zap.Error(err))
		return err
	}
	c.logger.Info("Cache health check passed")
	return nil
}

// Set sets a value in the cache with a specified key.
func (c *cacheService) Set(ctx context.Context, key string, value interface{}) error {
	// Ensure the cache is initialized
	if !c.initialized {
		if err := c.Initialize(); err != nil {
			return err
		}
	}

	data, err := json.Marshal(value)
	if err != nil {
		c.logger.Error("Failed to marshal data for caching", zap.Error(err))
		return err
	}

	err = c.redisClient.Set(ctx, key, data, c.ttl).Err()
	if err != nil {
		c.logger.Error("Failed to set data in Redis", zap.Error(err))
		return err
	}

	c.logger.Info("Data cached successfully", zap.String("key", key))
	return nil
}

// Get retrieves a value from the cache by key.
func (c *cacheService) Get(ctx context.Context, key string, dest interface{}) error {
	// Ensure the cache is initialized
	if !c.initialized {
		if err := c.Initialize(); err != nil {
			return err
		}
	}

	data, err := c.redisClient.Get(ctx, key).Result()
	if err == redis.Nil {
		cacheMisses.Inc()
		c.logger.Info("Cache miss", zap.String("key", key))
		return nil // Cache miss is not an error
	} else if err != nil {
		cacheErrors.Inc()
		c.logger.Error("Failed to get data from Redis", zap.Error(err))
		return err
	}

	err = json.Unmarshal([]byte(data), dest)
	if err != nil {
		cacheErrors.Inc()
		c.logger.Error("Failed to unmarshal cached data", zap.Error(err))
		return err
	}

	c.logger.Info("Data retrieved from cache", zap.String("key", key))
	return nil
}

// Delete removes a value from the cache by key.
func (c *cacheService) Delete(ctx context.Context, key string) error {
	// Ensure the cache is initialized
	if !c.initialized {
		if err := c.Initialize(); err != nil {
			return err
		}
	}

	err := c.redisClient.Del(ctx, key).Err()
	if err != nil {
		cacheErrors.Inc()
		c.logger.Error("Failed to delete data from Redis", zap.Error(err))
		return err
	}

	c.logger.Info("Data deleted from cache", zap.String("key", key))
	return nil
}

// HSet sets a field in a Redis hash.
func (c *cacheService) HSet(ctx context.Context, key, field string, value interface{}) error {
	// Ensure the cache is initialized
	if !c.initialized {
		if err := c.Initialize(); err != nil {
			return err
		}
	}

	data, err := json.Marshal(value)
	if err != nil {
		cacheErrors.Inc()
		c.logger.Error("Failed to marshal data for HSet", zap.Error(err))
		return err
	}

	err = c.redisClient.HSet(ctx, key, field, data).Err()
	if err != nil {
		cacheErrors.Inc()
		c.logger.Error("Failed to set hash field in Redis", zap.Error(err))
		return err
	}

	c.logger.Info("Hash field set successfully", zap.String("key", key), zap.String("field", field))
	return nil
}

// HGet retrieves a field from a Redis hash.
func (c *cacheService) HGet(ctx context.Context, key, field string) (string, error) {
	// Ensure the cache is initialized
	if !c.initialized {
		if err := c.Initialize(); err != nil {
			return "", err
		}
	}

	value, err := c.redisClient.HGet(ctx, key, field).Result()
	if err == redis.Nil {
		cacheMisses.Inc()
		c.logger.Info("Hash field not found", zap.String("key", key), zap.String("field", field))
		return "", nil
	} else if err != nil {
		cacheErrors.Inc()
		c.logger.Error("Failed to get hash field from Redis", zap.Error(err))
		return "", err
	}

	c.logger.Info("Hash field retrieved successfully", zap.String("key", key), zap.String("field", field))
	return value, nil
}

// Pipeline executes multiple Redis commands in a single round-trip.
func (c *cacheService) Pipeline(ctx context.Context, fn func(redis.Pipeliner) error) error {
	// Ensure the cache is initialized
	if !c.initialized {
		if err := c.Initialize(); err != nil {
			return err
		}
	}

	pipe := c.redisClient.Pipeline()
	err := fn(pipe)
	if err != nil {
		c.logger.Error("Pipeline function execution failed", zap.Error(err))
		cacheErrors.Inc()
		return err
	}

	_, err = pipe.Exec(ctx)
	if err != nil {
		c.logger.Error("Pipeline execution failed", zap.Error(err))
		cacheErrors.Inc()
		return err
	}

	c.logger.Info("Pipeline executed successfully")
	return nil
}

// SAdd adds members to a Redis set.
func (c *cacheService) SAdd(ctx context.Context, key string, members ...interface{}) error {
	// Ensure the cache is initialized
	if !c.initialized {
		if err := c.Initialize(); err != nil {
			return err
		}
	}

	err := c.redisClient.SAdd(ctx, key, members...).Err()
	if err != nil {
		cacheErrors.Inc()
		c.logger.Error("Failed to add members to set", zap.Error(err))
		return err
	}

	c.logger.Info("Members added to set successfully", zap.String("key", key))
	return nil
}

// SMembers returns all members of a Redis set.
func (c *cacheService) SMembers(ctx context.Context, key string) ([]string, error) {
	// Ensure the cache is initialized
	if !c.initialized {
		if err := c.Initialize(); err != nil {
			return nil, err
		}
	}

	members, err := c.redisClient.SMembers(ctx, key).Result()
	if err != nil {
		cacheErrors.Inc()
		c.logger.Error("Failed to get set members", zap.Error(err))
		return nil, err
	}

	c.logger.Info("Set members retrieved successfully", zap.String("key", key))
	return members, nil
}

// ZAdd adds members to a Redis sorted set.
func (c *cacheService) ZAdd(ctx context.Context, key string, members ...*redis.Z) error {
	// Ensure the cache is initialized
	if !c.initialized {
		if err := c.Initialize(); err != nil {
			return err
		}
	}

	err := c.redisClient.ZAdd(ctx, key, members...).Err()
	if err != nil {
		cacheErrors.Inc()
		c.logger.Error("Failed to add members to sorted set", zap.Error(err))
		return err
	}

	c.logger.Info("Members added to sorted set successfully", zap.String("key", key))
	return nil
}

// ZRange returns a range of members from a Redis sorted set.
func (c *cacheService) ZRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	// Ensure the cache is initialized
	if !c.initialized {
		if err := c.Initialize(); err != nil {
			return nil, err
		}
	}

	members, err := c.redisClient.ZRange(ctx, key, start, stop).Result()
	if err != nil {
		cacheErrors.Inc()
		c.logger.Error("Failed to get range from sorted set", zap.Error(err))
		return nil, err
	}

	c.logger.Info("Range retrieved from sorted set successfully", zap.String("key", key))
	return members, nil
}

// Incr atomically increments a key's value.
func (c *cacheService) Incr(ctx context.Context, key string) (int64, error) {
	// Ensure the cache is initialized
	if !c.initialized {
		if err := c.Initialize(); err != nil {
			return 0, err
		}
	}

	value, err := c.redisClient.Incr(ctx, key).Result()
	if err != nil {
		cacheErrors.Inc()
		c.logger.Error("Failed to increment key", zap.Error(err))
		return 0, err
	}

	c.logger.Info("Key incremented successfully", zap.String("key", key), zap.Int64("value", value))
	return value, nil
}

// Decr atomically decrements a key's value.
func (c *cacheService) Decr(ctx context.Context, key string) (int64, error) {
	// Ensure the cache is initialized
	if !c.initialized {
		if err := c.Initialize(); err != nil {
			return 0, err
		}
	}

	value, err := c.redisClient.Decr(ctx, key).Result()
	if err != nil {
		cacheErrors.Inc()
		c.logger.Error("Failed to decrement key", zap.Error(err))
		return 0, err
	}

	c.logger.Info("Key decremented successfully", zap.String("key", key), zap.Int64("value", value))
	return value, nil
}

// GetOrSet retrieves a value from the cache or sets it using the provided fetch function.
// It implements rate limiting to prevent cache stampedes.
func (c *cacheService) GetOrSet(ctx context.Context, key string, dest interface{}, fetchFunc func() error) error {
	// Acquire a slot in the rate limiter
	select {
	case c.rateLimiter <- struct{}{}:
		// Acquired, proceed
	default:
		// Rate limit exceeded
		c.logger.Warn("Rate limit exceeded for cache GetOrSet", zap.String("key", key))
		return errors.New("rate limit exceeded")
	}
	// Ensure the slot is released
	defer func() {
		<-c.rateLimiter
	}()

	// Attempt to get from cache
	err := c.Get(ctx, key, dest)
	if err != nil {
		cacheErrors.Inc()
		return err
	}

	// Check if cache miss
	isEmpty, err := isEmpty(dest)
	if err != nil {
		cacheErrors.Inc()
		return err
	}
	if !isEmpty {
		// Cache hit
		return nil
	}

	// Cache miss, fetch data
	if fetchFunc == nil {
		c.logger.Warn("Fetch function is nil for cache miss", zap.String("key", key))
		return errors.New("fetch function is nil")
	}

	// Fetch data
	err = fetchFunc()
	if err != nil {
		cacheErrors.Inc()
		c.logger.Error("Fetch function failed", zap.Error(err))
		return err
	}

	// Set to cache
	err = c.Set(ctx, key, dest)
	if err != nil {
		cacheErrors.Inc()
		return err
	}

	return nil
}

// isEmpty checks if the destination is empty (cache miss).
func isEmpty(dest interface{}) (bool, error) {
	switch v := dest.(type) {
	case *[]*model.Product:
		return v == nil || len(*v) == 0, nil

	// Add other cases based on your models
	default:
		return false, fmt.Errorf("unsupported type for isEmpty check: %T", dest)
	}
}