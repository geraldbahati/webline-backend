// internal/middleware/rate_limit.go

package middleware

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

// Client holds the rate limiter and the last seen time.
type Client struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// RateLimiter manages rate limiters for multiple clients.
type RateLimiter struct {
	clients         map[string]*Client
	mu              sync.RWMutex
	rps             rate.Limit
	burst           int
	cleanupInterval time.Duration
	logger          *zap.Logger // Added logger
}

// NewRateLimiter creates a new RateLimiter.
func NewRateLimiter(rps float64, burst int, cleanupInterval time.Duration, logger *zap.Logger) *RateLimiter {
	rl := &RateLimiter{
		clients:         make(map[string]*Client),
		rps:             rate.Limit(rps),
		burst:           burst,
		cleanupInterval: cleanupInterval,
		logger:          logger, // Initialize logger
	}

	// Start the cleanup goroutine
	go rl.cleanupStaleClients()

	return rl
}

// getLimiter retrieves the rate limiter for the given IP, creating one if it doesn't exist.
func (rl *RateLimiter) getLimiter(ip string) *rate.Limiter {
	// First, try to read with read lock
	rl.mu.RLock()
	client, exists := rl.clients[ip]
	rl.mu.RUnlock()

	if exists {
		// Update lastSeen with write lock
		rl.mu.Lock()
		client.lastSeen = time.Now()
		rl.mu.Unlock()
		return client.limiter
	}

	// If not exists, acquire write lock to create a new limiter
	rl.mu.Lock()
	defer rl.mu.Unlock()

	// Double-check if another goroutine has created the limiter
	if client, exists = rl.clients[ip]; exists {
		client.lastSeen = time.Now()
		return client.limiter
	}

	limiter := rate.NewLimiter(rl.rps, rl.burst)
	rl.clients[ip] = &Client{
		limiter:  limiter,
		lastSeen: time.Now(),
	}

	// Log the creation of a new limiter
	if rl.logger != nil {
		rl.logger.Info("Created new rate limiter for IP", zap.String("ip", ip))
	}

	return limiter
}

// cleanupStaleClients periodically removes clients that haven't been seen for longer than cleanupInterval.
func (rl *RateLimiter) cleanupStaleClients() {
	ticker := time.NewTicker(rl.cleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		for ip, client := range rl.clients {
			if time.Since(client.lastSeen) > rl.cleanupInterval {
				delete(rl.clients, ip)
				// Log the cleanup action
				if rl.logger != nil {
					rl.logger.Info("Cleaned up stale rate limiter", zap.String("ip", ip))
				}
			}
		}
		rl.mu.Unlock()
	}
}

// RateLimitMiddleware returns a middleware that limits requests per client IP.
func RateLimitMiddleware(rps float64, burst int, cleanupInterval time.Duration, logger *zap.Logger) func(http.Handler) http.Handler {
	rl := NewRateLimiter(rps, burst, cleanupInterval, logger)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := getClientIP(r)
			if ip == "" {
				// Unable to determine IP, treat as internal server error
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			limiter := rl.getLimiter(ip)
			if !limiter.Allow() {
				// Rate limit exceeded
				w.Header().Set("Retry-After", "60") // Example: 60 seconds
				http.Error(w, "Too Many Requests", http.StatusTooManyRequests)

				// Log the rate limit event
				if rl.logger != nil {
					rl.logger.Warn("Rate limit exceeded", zap.String("ip", ip), zap.String("path", r.URL.Path))
				}
				return
			}

			// Set rate limit headers
			w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%.2f", rl.rps))
			// X-RateLimit-Remaining is omitted due to limitations in rate.Limiter

			next.ServeHTTP(w, r)
		})
	}
}
