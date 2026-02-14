package middleware

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
)

// RateLimitConfig holds rate limiting configuration.
type RateLimitConfig struct {
	RequestsPerSecond float64
	BurstSize         int
}

// DefaultRateLimitConfig returns default rate limiting settings.
func DefaultRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		RequestsPerSecond: 100,
		BurstSize:         200,
	}
}

// tokenBucket implements a token bucket rate limiter.
type tokenBucket struct {
	tokens     float64
	maxTokens  float64
	refillRate float64 // tokens per second
	lastRefill time.Time
	mu         sync.Mutex
}

func newTokenBucket(rate float64, burst int) *tokenBucket {
	return &tokenBucket{
		tokens:     float64(burst),
		maxTokens:  float64(burst),
		refillRate: rate,
		lastRefill: time.Now(),
	}
}

func (b *tokenBucket) allow() bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(b.lastRefill).Seconds()
	b.tokens += elapsed * b.refillRate
	if b.tokens > b.maxTokens {
		b.tokens = b.maxTokens
	}
	b.lastRefill = now

	if b.tokens >= 1 {
		b.tokens--
		return true
	}
	return false
}

func (b *tokenBucket) retryAfter() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.refillRate <= 0 {
		return 1
	}
	return int((1-b.tokens)/b.refillRate) + 1
}

// rateLimiterStore holds per-key token buckets.
type rateLimiterStore struct {
	buckets map[string]*tokenBucket
	mu      sync.RWMutex
	config  RateLimitConfig
}

func newRateLimiterStore(cfg RateLimitConfig) *rateLimiterStore {
	return &rateLimiterStore{
		buckets: make(map[string]*tokenBucket),
		config:  cfg,
	}
}

func (s *rateLimiterStore) getBucket(key string) *tokenBucket {
	s.mu.RLock()
	bucket, ok := s.buckets[key]
	s.mu.RUnlock()
	if ok {
		return bucket
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	// Double-check after acquiring write lock
	if bucket, ok := s.buckets[key]; ok {
		return bucket
	}
	bucket = newTokenBucket(s.config.RequestsPerSecond, s.config.BurstSize)
	s.buckets[key] = bucket
	return bucket
}

// RateLimit returns a rate limiting middleware.
func RateLimit(cfg RateLimitConfig) echo.MiddlewareFunc {
	store := newRateLimiterStore(cfg)

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Use IP as rate limit key, fall back to tenant ID
			key := c.RealIP()
			if tenantID := c.Get("jwt_tenant_id"); tenantID != nil {
				key = tenantID.(string) + ":" + key
			}

			bucket := store.getBucket(key)
			if !bucket.allow() {
				retryAfter := bucket.retryAfter()
				c.Response().Header().Set("Retry-After", strconv.Itoa(retryAfter))
				c.Response().Header().Set("X-RateLimit-Limit", strconv.FormatFloat(cfg.RequestsPerSecond, 'f', 0, 64))
				c.Response().Header().Set("X-RateLimit-Remaining", "0")
				return echo.NewHTTPError(http.StatusTooManyRequests, "rate limit exceeded")
			}

			c.Response().Header().Set("X-RateLimit-Limit", strconv.FormatFloat(cfg.RequestsPerSecond, 'f', 0, 64))
			return next(c)
		}
	}
}
