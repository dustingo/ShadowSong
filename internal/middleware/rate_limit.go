package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimiter interface for rate limiting implementations
type RateLimiter interface {
	Allow(key string) bool
}

// InMemoryRateLimiter implements a simple in-memory sliding window rate limiter
type InMemoryRateLimiter struct {
	mu       sync.Mutex
	requests map[string]*counter
	limit    int
	window   time.Duration
}

type counter struct {
	count     int
	expiresAt time.Time
}

// NewInMemoryRateLimiter creates a new in-memory rate limiter
func NewInMemoryRateLimiter(limit int, window time.Duration) *InMemoryRateLimiter {
	return &InMemoryRateLimiter{
		requests: make(map[string]*counter),
		limit:    limit,
		window:   window,
	}
}

// Allow checks if a request with the given key is allowed.
// Also performs periodic cleanup of expired entries to prevent memory leaks.
func (l *InMemoryRateLimiter) Allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()

	// Periodic cleanup: remove expired entries every 100 requests or so
	// This prevents unbounded memory growth with many unique keys
	if len(l.requests) > 100 {
		for k, c := range l.requests {
			if now.After(c.expiresAt) {
				delete(l.requests, k)
			}
		}
	}

	if c, exists := l.requests[key]; exists {
		if now.After(c.expiresAt) {
			// Window expired, reset counter
			l.requests[key] = &counter{count: 1, expiresAt: now.Add(l.window)}
			return true
		}
		if c.count >= l.limit {
			return false
		}
		c.count++
		return true
	}

	l.requests[key] = &counter{count: 1, expiresAt: now.Add(l.window)}
	return true
}

// RateLimit returns a middleware that rate limits requests based on a key function
func RateLimit(limiter RateLimiter, keyFunc func(*gin.Context) string) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := keyFunc(c)
		if !limiter.Allow(key) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "rate limit exceeded",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}