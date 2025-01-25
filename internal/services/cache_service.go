// cache_service.go

package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"
	"weblineBackend/internal/model"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
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

// CacheService interface defines methods for cache operations.
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
	DeleteKeysByPattern(ctx context.Context, pattern string) error
	HealthCheck(ctx context.Context) error
	Initialize() error
	SetWithTTL(ctx context.Context, key string, value interface{}, ttl time.Duration) error
}

type cacheService struct {
	redisClient *redis.Client
	logger      *zap.Logger
	ttl         time.Duration
	initialized bool
	mu          sync.Mutex

	// Rate limiting
	rateLimiter chan struct{}
}

// NewCacheService creates a new instance of CacheService.
func NewCacheService(redisClient *redis.Client, logger *zap.Logger, ttl time.Duration, rateLimit int) CacheService {
	return &cacheService{
		redisClient: redisClient,
		logger:      logger,
		ttl:         ttl,
		rateLimiter: make(chan struct{}, rateLimit),
	}
}

// Initialize performs any necessary initialization tasks.
func (c *cacheService) Initialize() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.initialized {
		return nil
	}

	// Perform initialization tasks if needed.
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
		c.logger.Error("Failed to set data in Redis", zap.Error(err), zap.String("key", key))
		return err
	}

	c.logger.Debug("Data cached successfully", zap.String("key", key))
	return nil
}

// Get retrieves a value from the cache by key.
func (c *cacheService) Get(ctx context.Context, key string, dest interface{}) error {
	if !c.initialized {
		if err := c.Initialize(); err != nil {
			return err
		}
	}

	data, err := c.redisClient.Get(ctx, key).Result()
	if err == redis.Nil {
		cacheMisses.Inc()
		c.logger.Debug("Cache miss", zap.String("key", key))
		return nil // Cache miss is not an error
	} else if err != nil {
		cacheErrors.Inc()
		c.logger.Error("Failed to get data from Redis", zap.Error(err), zap.String("key", key))
		return err
	}

	err = json.Unmarshal([]byte(data), dest)
	if err != nil {
		cacheErrors.Inc()
		c.logger.Error("Failed to unmarshal cached data", zap.Error(err), zap.String("key", key))
		return err
	}

	c.logger.Debug("Data retrieved from cache", zap.String("key", key))
	cacheHits.Inc()
	return nil
}

// Delete removes a value from the cache by key.
func (c *cacheService) Delete(ctx context.Context, key string) error {
	if !c.initialized {
		if err := c.Initialize(); err != nil {
			return err
		}
	}

	err := c.redisClient.Del(ctx, key).Err()
	if err != nil {
		cacheErrors.Inc()
		c.logger.Error("Failed to delete data from Redis", zap.Error(err), zap.String("key", key))
		return err
	}

	c.logger.Debug("Data deleted from cache", zap.String("key", key))
	return nil
}

// HSet sets a field in a Redis hash.
func (c *cacheService) HSet(ctx context.Context, key, field string, value interface{}) error {
	if !c.initialized {
		if err := c.Initialize(); err != nil {
			return err
		}
	}

	data, err := json.Marshal(value)
	if err != nil {
		cacheErrors.Inc()
		c.logger.Error("Failed to marshal data for HSet", zap.Error(err), zap.String("key", key), zap.String("field", field))
		return err
	}

	err = c.redisClient.HSet(ctx, key, field, data).Err()
	if err != nil {
		cacheErrors.Inc()
		c.logger.Error("Failed to set hash field in Redis", zap.Error(err), zap.String("key", key), zap.String("field", field))
		return err
	}

	c.logger.Debug("Hash field set successfully", zap.String("key", key), zap.String("field", field))
	return nil
}

// HGet retrieves a field from a Redis hash.
func (c *cacheService) HGet(ctx context.Context, key, field string) (string, error) {
	if !c.initialized {
		if err := c.Initialize(); err != nil {
			return "", err
		}
	}

	value, err := c.redisClient.HGet(ctx, key, field).Result()
	if err == redis.Nil {
		cacheMisses.Inc()
		c.logger.Debug("Hash field not found", zap.String("key", key), zap.String("field", field))
		return "", nil
	} else if err != nil {
		cacheErrors.Inc()
		c.logger.Error("Failed to get hash field from Redis", zap.Error(err), zap.String("key", key), zap.String("field", field))
		return "", err
	}

	c.logger.Debug("Hash field retrieved successfully", zap.String("key", key), zap.String("field", field))
	cacheHits.Inc()
	return value, nil
}

// Pipeline executes multiple Redis commands in a single round-trip.
func (c *cacheService) Pipeline(ctx context.Context, fn func(redis.Pipeliner) error) error {
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

	c.logger.Debug("Pipeline executed successfully")
	return nil
}

// SAdd adds members to a Redis set.
func (c *cacheService) SAdd(ctx context.Context, key string, members ...interface{}) error {
	if !c.initialized {
		if err := c.Initialize(); err != nil {
			return err
		}
	}

	err := c.redisClient.SAdd(ctx, key, members...).Err()
	if err != nil {
		cacheErrors.Inc()
		c.logger.Error("Failed to add members to set", zap.Error(err), zap.String("key", key))
		return err
	}

	c.logger.Debug("Members added to set successfully", zap.String("key", key))
	return nil
}

// SMembers returns all members of a Redis set.
func (c *cacheService) SMembers(ctx context.Context, key string) ([]string, error) {
	if !c.initialized {
		if err := c.Initialize(); err != nil {
			return nil, err
		}
	}

	members, err := c.redisClient.SMembers(ctx, key).Result()
	if err != nil {
		cacheErrors.Inc()
		c.logger.Error("Failed to get set members", zap.Error(err), zap.String("key", key))
		return nil, err
	}

	c.logger.Debug("Set members retrieved successfully", zap.String("key", key))
	cacheHits.Inc()
	return members, nil
}

// ZAdd adds members to a Redis sorted set.
func (c *cacheService) ZAdd(ctx context.Context, key string, members ...*redis.Z) error {
	if !c.initialized {
		if err := c.Initialize(); err != nil {
			return err
		}
	}

	err := c.redisClient.ZAdd(ctx, key, members...).Err()
	if err != nil {
		cacheErrors.Inc()
		c.logger.Error("Failed to add members to sorted set", zap.Error(err), zap.String("key", key))
		return err
	}

	c.logger.Debug("Members added to sorted set successfully", zap.String("key", key))
	return nil
}

// ZRange returns a range of members from a Redis sorted set.
func (c *cacheService) ZRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	if !c.initialized {
		if err := c.Initialize(); err != nil {
			return nil, err
		}
	}

	members, err := c.redisClient.ZRange(ctx, key, start, stop).Result()
	if err != nil {
		cacheErrors.Inc()
		c.logger.Error("Failed to get range from sorted set", zap.Error(err), zap.String("key", key))
		return nil, err
	}

	c.logger.Debug("Range retrieved from sorted set successfully", zap.String("key", key))
	cacheHits.Inc()
	return members, nil
}

// Incr atomically increments a key's value.
func (c *cacheService) Incr(ctx context.Context, key string) (int64, error) {
	if !c.initialized {
		if err := c.Initialize(); err != nil {
			return 0, err
		}
	}

	value, err := c.redisClient.Incr(ctx, key).Result()
	if err != nil {
		cacheErrors.Inc()
		c.logger.Error("Failed to increment key", zap.Error(err), zap.String("key", key))
		return 0, err
	}

	c.logger.Debug("Key incremented successfully", zap.String("key", key), zap.Int64("value", value))
	cacheHits.Inc()
	return value, nil
}

// Decr atomically decrements a key's value.
func (c *cacheService) Decr(ctx context.Context, key string) (int64, error) {
	if !c.initialized {
		if err := c.Initialize(); err != nil {
			return 0, err
		}
	}

	value, err := c.redisClient.Decr(ctx, key).Result()
	if err != nil {
		cacheErrors.Inc()
		c.logger.Error("Failed to decrement key", zap.Error(err), zap.String("key", key))
		return 0, err
	}

	c.logger.Debug("Key decremented successfully", zap.String("key", key), zap.Int64("value", value))
	cacheHits.Inc()
	return value, nil
}

// DeleteKeysByPattern deletes keys matching a specific pattern.
func (c *cacheService) DeleteKeysByPattern(ctx context.Context, pattern string) error {
	if !c.initialized {
		if err := c.Initialize(); err != nil {
			return err
		}
	}

	var cursor uint64
	batchSize := int64(100) // Number of keys to delete per batch

	for {
		fetchedKeys, nextCursor, err := c.redisClient.Scan(ctx, cursor, pattern, batchSize).Result()
		if err != nil {
			cacheErrors.Inc()
			c.logger.Error("Failed to scan keys for deletion", zap.String("pattern", pattern), zap.Error(err))
			return fmt.Errorf("failed to scan keys for deletion: %w", err)
		}

		if len(fetchedKeys) > 0 {
			// Use UNLINK for non-blocking deletion
			if err := c.redisClient.Unlink(ctx, fetchedKeys...).Err(); err != nil {
				cacheErrors.Inc()
				c.logger.Error("Failed to unlink keys", zap.Strings("keys", fetchedKeys), zap.Error(err))
			} else {
				c.logger.Debug("Unlinked keys successfully", zap.Int("count", len(fetchedKeys)), zap.String("pattern", pattern))
			}
		}

		if nextCursor == 0 {
			break
		}
		cursor = nextCursor
	}

	c.logger.Info("Completed deletion of keys by pattern", zap.String("pattern", pattern))
	return nil
}

// GetOrSet retrieves a value from the cache or sets it using the provided fetch function.
func (c *cacheService) GetOrSet(ctx context.Context, key string, dest interface{}, fetchFunc func() error) error {
	// Acquire a slot in the rate limiter
	select {
	case c.rateLimiter <- struct{}{}:
		// Proceed
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
		c.logger.Error("Failed to get data from cache", zap.Error(err), zap.String("key", key))
		return err
	}

	// Check if cache miss
	isEmpty, err := isEmpty(dest)
	if err != nil {
		cacheErrors.Inc()
		c.logger.Error("Failed to check if destination is empty", zap.Error(err), zap.String("key", key))
		return err
	}
	if !isEmpty {
		// Cache hit
		cacheHits.Inc()
		return nil
	}

	// Cache miss, fetch data
	if fetchFunc == nil {
		c.logger.Warn("Fetch function is nil for cache miss", zap.String("key", key))
		cacheMisses.Inc()
		return errors.New("fetch function is nil")
	}

	// Fetch data
	err = fetchFunc()
	if err != nil {
		cacheMisses.Inc()
		c.logger.Error("Fetch function failed", zap.Error(err), zap.String("key", key))
		return err
	}

	// Set to cache
	err = c.Set(ctx, key, dest)
	if err != nil {
		cacheErrors.Inc()
		c.logger.Error("Failed to set data in cache after fetch", zap.Error(err), zap.String("key", key))
		return err
	}

	cacheMisses.Inc()
	return nil
}

// isEmpty checks if the destination is empty (cache miss).
func isEmpty(dest interface{}) (bool, error) {
	switch v := dest.(type) {
	case *model.V2ProductDetail:
		return v == nil || v.Slug == "", nil
	case *model.ProductSEO:
		return v == nil || v.ID == uuid.Nil, nil
	case *[]*model.Product:
		return v == nil || len(*v) == 0, nil
	case *[]*model.ProductSitemap:
		return v == nil || len(*v) == 0, nil
	case *model.ProductSchema:
		return v == nil || v.Slug == "", nil
	case *float64:
		// For exchange rate
		return v == nil || *v == 0, nil
	case *model.ProductPricing:
		return v == nil || v.ID == uuid.Nil, nil
	case *model.ProductSpecs:
		return v == nil || v.Description == "" || len(v.Specs) == 0, nil
	case *[]string: // Handle []string
		return v == nil || len(*v) == 0, nil
	case *model.ProductCart:
		return v == nil || v.ID == uuid.Nil, nil
	case *model.Session:
		return v == nil || v.ID == uuid.Nil, nil
	case *model.AdminRequest:
		return v == nil || v.ID == uuid.Nil, nil
	case *model.ApprovalToken:
		return v == nil || v.ID == uuid.Nil, nil
	case *model.User:
		return v == nil || v.ID == uuid.Nil, nil
	case *model.ShoppingCart:
		return v == nil || v.ID == uuid.Nil, nil
	case *model.CartItem:
		return v == nil || v.ID == uuid.Nil, nil
	case *[]model.CartItem:
		return v == nil || len(*v) == 0, nil
	case *[]*model.CartItem:
		return v == nil || len(*v) == 0, nil
	case *[]model.ProductImage:
		return v == nil || len(*v) == 0, nil
	// Add more cases based on your models
	default:
		return false, fmt.Errorf("unsupported type for isEmpty check: %T", dest)
	}
}

// SetWithTTL sets a value in the cache with a specified key and TTL.
func (c *cacheService) SetWithTTL(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	if !c.initialized {
		if err := c.Initialize(); err != nil {
			return err
		}
	}

	data, err := json.Marshal(value)
	if err != nil {
		c.logger.Error("Failed to marshal data for caching", zap.Error(err))
		return fmt.Errorf("marshal data: %w", err)
	}

	err = c.redisClient.Set(ctx, key, data, ttl).Err()
	if err != nil {
		c.logger.Error("Failed to set data in Redis with TTL",
			zap.Error(err),
			zap.String("key", key),
			zap.Duration("ttl", ttl))
		return fmt.Errorf("set with TTL: %w", err)
	}

	c.logger.Debug("Data cached successfully with TTL",
		zap.String("key", key),
		zap.Duration("ttl", ttl))
	return nil
}
