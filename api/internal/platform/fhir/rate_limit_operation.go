package fhir

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
)

// OperationRateConfig defines the rate limit configuration for a specific FHIR
// operation category. Each operation (e.g. "$export", "read", "search") can
// have its own maximum request count and sliding window duration.
type OperationRateConfig struct {
	Operation   string
	MaxRequests int
	Window      time.Duration
}

// OperationRateLimiter enforces per-operation rate limits. Different FHIR
// operations can have independent rate limit configurations, allowing heavy
// operations like $export to have lower limits than simple reads. Each
// (operation, clientID) pair is tracked independently using a
// SlidingWindowLimiter.
type OperationRateLimiter struct {
	mu            sync.RWMutex
	configs       map[string]OperationRateConfig
	limiters      map[string]*SlidingWindowLimiter
	defaultConfig OperationRateConfig
}

// NewOperationRateLimiter creates an OperationRateLimiter with the given
// default maximum requests and window duration. Operations without an explicit
// configuration use these defaults.
func NewOperationRateLimiter(defaultMax int, defaultWindow time.Duration) *OperationRateLimiter {
	return &OperationRateLimiter{
		configs:  make(map[string]OperationRateConfig),
		limiters: make(map[string]*SlidingWindowLimiter),
		defaultConfig: OperationRateConfig{
			Operation:   "default",
			MaxRequests: defaultMax,
			Window:      defaultWindow,
		},
	}
}

// SetOperationLimit configures a per-operation rate limit. If a limiter already
// exists for the operation it is replaced.
func (o *OperationRateLimiter) SetOperationLimit(operation string, maxRequests int, window time.Duration) {
	o.mu.Lock()
	defer o.mu.Unlock()

	o.configs[operation] = OperationRateConfig{
		Operation:   operation,
		MaxRequests: maxRequests,
		Window:      window,
	}
	// Remove any existing limiter so it is re-created with the new config on
	// the next call to Allow.
	delete(o.limiters, operation)
}

// Allow checks whether the request for the given operation and client is
// permitted under the configured rate limit. It returns true if the request is
// allowed and false otherwise.
func (o *OperationRateLimiter) Allow(operation, clientID string) bool {
	o.mu.Lock()
	defer o.mu.Unlock()

	limiter, exists := o.limiters[operation]
	if !exists {
		cfg := o.configForLocked(operation)
		limiter = NewSlidingWindowLimiter(cfg.MaxRequests, cfg.Window)
		o.limiters[operation] = limiter
	}

	allowed, _, _ := limiter.Allow(clientID)
	return allowed
}

// GetLimit returns the OperationRateConfig for the given operation. If no
// explicit configuration has been set the default config is returned with the
// Operation field set to the requested operation name.
func (o *OperationRateLimiter) GetLimit(operation string) OperationRateConfig {
	o.mu.RLock()
	defer o.mu.RUnlock()

	if cfg, ok := o.configs[operation]; ok {
		return cfg
	}
	return OperationRateConfig{
		Operation:   operation,
		MaxRequests: o.defaultConfig.MaxRequests,
		Window:      o.defaultConfig.Window,
	}
}

// configForLocked returns the config for an operation. Caller must hold o.mu.
func (o *OperationRateLimiter) configForLocked(operation string) OperationRateConfig {
	if cfg, ok := o.configs[operation]; ok {
		return cfg
	}
	return o.defaultConfig
}

// limiterForOperation returns the SlidingWindowLimiter for a given operation,
// along with the applicable config. The returned limiter can be used to obtain
// remaining-count and reset-time information.
func (o *OperationRateLimiter) limiterForOperation(operation string) (*SlidingWindowLimiter, OperationRateConfig) {
	o.mu.Lock()
	defer o.mu.Unlock()

	cfg := o.configForLocked(operation)
	limiter, exists := o.limiters[operation]
	if !exists {
		limiter = NewSlidingWindowLimiter(cfg.MaxRequests, cfg.Window)
		o.limiters[operation] = limiter
	}
	return limiter, cfg
}

// classifyOperation determines the FHIR operation category from the request
// method and URL path. The returned string is used as the key for per-operation
// rate limiting.
func classifyOperation(c echo.Context) string {
	method := c.Request().Method
	path := c.Request().URL.Path

	segments := trimmedPathSegments(path)
	resSegs := resourceSegments(segments)

	// Check for system-level operations first (e.g. POST /$export).
	for _, seg := range resSegs {
		if strings.HasPrefix(seg, "$") {
			return seg
		}
	}

	// Classify based on HTTP method.
	switch method {
	case http.MethodGet:
		if len(resSegs) >= 2 && resSegs[1] != "_history" {
			return "read"
		}
		return "search"
	case http.MethodPost:
		return "create"
	case http.MethodPut:
		return "update"
	case http.MethodPatch:
		return "update"
	case http.MethodDelete:
		return "delete"
	default:
		return "other"
	}
}

// OperationRateLimitMiddleware returns an Echo middleware that enforces
// per-operation rate limits. It classifies each request into an operation
// category, identifies the client by IP or X-API-Key header, and checks the
// corresponding limiter. Rejected requests receive a 429 response with
// appropriate rate-limit headers.
func OperationRateLimitMiddleware(limiter *OperationRateLimiter) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			operation := classifyOperation(c)

			// Prefer API key for client identification; fall back to IP.
			client := c.Request().Header.Get("X-API-Key")
			if client == "" {
				client = clientKey(c.Request())
			}

			swLimiter, cfg := limiter.limiterForOperation(operation)
			allowed, remaining, resetAt := swLimiter.Allow(client)

			now := time.Now()

			// Set rate-limit headers on every response.
			h := c.Response().Header()
			h.Set("X-RateLimit-Limit", fmt.Sprintf("%d", cfg.MaxRequests))
			h.Set("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
			h.Set("X-RateLimit-Reset", fmt.Sprintf("%d", resetAt.Unix()))
			h.Set("X-RateLimit-Operation", operation)

			if !allowed {
				retryAfter := resetAt.Sub(now)
				if retryAfter < 0 {
					retryAfter = 0
				}
				h.Set("Retry-After", fmt.Sprintf("%d", int(retryAfter.Seconds())+1))
				return c.JSON(http.StatusTooManyRequests, NewOperationOutcome(
					"error",
					"throttled",
					fmt.Sprintf("Rate limit exceeded for operation %q. Please retry after the period indicated in the Retry-After header.", operation),
				))
			}

			return next(c)
		}
	}
}

// DefaultOperationRateLimits returns an OperationRateLimiter pre-configured
// with sensible defaults for common FHIR operations:
//
//   - read:     1000 requests per minute
//   - search:    500 requests per minute
//   - create:    200 requests per minute
//   - update:    200 requests per minute
//   - delete:    200 requests per minute
//   - $export:    10 requests per minute
//   - $import:    10 requests per minute
//   - $validate: 100 requests per minute
//   - other:      50 requests per minute (default)
func DefaultOperationRateLimits() *OperationRateLimiter {
	limiter := NewOperationRateLimiter(50, time.Minute)

	limiter.SetOperationLimit("read", 1000, time.Minute)
	limiter.SetOperationLimit("search", 500, time.Minute)
	limiter.SetOperationLimit("create", 200, time.Minute)
	limiter.SetOperationLimit("update", 200, time.Minute)
	limiter.SetOperationLimit("delete", 200, time.Minute)
	limiter.SetOperationLimit("$export", 10, time.Minute)
	limiter.SetOperationLimit("$import", 10, time.Minute)
	limiter.SetOperationLimit("$validate", 100, time.Minute)

	return limiter
}
