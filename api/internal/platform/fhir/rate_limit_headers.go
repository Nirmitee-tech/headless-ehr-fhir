package fhir

import (
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
)

// RateLimiter defines the interface for FHIR-aware rate limiting.
// Implementations track request counts per client key and report whether a
// request should be allowed, how many requests remain, and when the current
// window resets.
type RateLimiter interface {
	// Allow checks whether the request identified by key is allowed.
	// It returns whether the request is permitted, the number of remaining
	// requests in the current window, and the time at which the window resets.
	Allow(key string) (allowed bool, remaining int, resetAt time.Time)
	// Limit returns the maximum number of requests permitted per window.
	Limit() int
}

// clientWindow tracks the request timestamps for a single client.
type clientWindow struct {
	requests []time.Time
}

// SlidingWindowLimiter implements RateLimiter using a per-client sliding window
// algorithm. Each client is identified by a key (typically derived from the
// request IP address). The limiter tracks individual request timestamps and
// evicts entries that fall outside the current window on every call to Allow.
type SlidingWindowLimiter struct {
	mu      sync.Mutex
	limit   int
	window  time.Duration
	clients map[string]*clientWindow
	nowFunc func() time.Time // for testing; defaults to time.Now
}

// NewSlidingWindowLimiter creates a SlidingWindowLimiter that permits at most
// limit requests per client within the given window duration.
func NewSlidingWindowLimiter(limit int, window time.Duration) *SlidingWindowLimiter {
	return &SlidingWindowLimiter{
		limit:   limit,
		window:  window,
		clients: make(map[string]*clientWindow),
		nowFunc: time.Now,
	}
}

// Limit returns the maximum number of requests per window.
func (s *SlidingWindowLimiter) Limit() int {
	return s.limit
}

// Allow determines whether the request identified by key is permitted. It
// evicts expired entries, records the current request if allowed, and returns
// the remaining quota and window reset time.
func (s *SlidingWindowLimiter) Allow(key string) (allowed bool, remaining int, resetAt time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := s.nowFunc()
	windowStart := now.Add(-s.window)

	cw, exists := s.clients[key]
	if !exists {
		cw = &clientWindow{}
		s.clients[key] = cw
	}

	// Evict requests outside the current window.
	valid := cw.requests[:0]
	for _, t := range cw.requests {
		if t.After(windowStart) {
			valid = append(valid, t)
		}
	}
	cw.requests = valid

	// Compute reset time: if there are existing requests, the window resets
	// when the oldest request in the window expires; otherwise it resets at the
	// end of a full window from now.
	if len(cw.requests) > 0 {
		resetAt = cw.requests[0].Add(s.window)
	} else {
		resetAt = now.Add(s.window)
	}

	if len(cw.requests) >= s.limit {
		// Denied — do not record this request.
		return false, 0, resetAt
	}

	// Record this request.
	cw.requests = append(cw.requests, now)
	remaining = s.limit - len(cw.requests)
	return true, remaining, resetAt
}

// clientKey extracts the rate-limit key from the request. It prefers the
// X-Forwarded-For header (first address) and falls back to the remote address.
func clientKey(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.SplitN(xff, ",", 2)
		if ip := strings.TrimSpace(parts[0]); ip != "" {
			return ip
		}
	}
	// RemoteAddr may include a port; strip it.
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// RateLimitMiddleware returns an Echo middleware that adds FHIR-recommended rate
// limit response headers to every response:
//
//   - X-RateLimit-Limit:     maximum requests per window
//   - X-RateLimit-Remaining: remaining requests in the current window
//   - X-RateLimit-Reset:     Unix timestamp when the current window resets
//
// When the rate limit is exceeded the middleware responds with HTTP 429 Too Many
// Requests, a Retry-After header (in seconds), and a FHIR OperationOutcome
// body. The downstream handler is not invoked for rejected requests.
func RateLimitMiddleware(limiter RateLimiter) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			key := clientKey(c.Request())
			allowed, remaining, resetAt := limiter.Allow(key)

			now := time.Now()

			// Set rate-limit headers on every response.
			h := c.Response().Header()
			h.Set("X-RateLimit-Limit", fmt.Sprintf("%d", limiter.Limit()))
			h.Set("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
			h.Set("X-RateLimit-Reset", fmt.Sprintf("%d", resetAt.Unix()))

			if !allowed {
				retryAfter := resetAt.Sub(now)
				if retryAfter < 0 {
					retryAfter = 0
				}
				h.Set("Retry-After", fmt.Sprintf("%d", int(retryAfter.Seconds())+1))
				return c.JSON(http.StatusTooManyRequests, NewOperationOutcome(
					"error",
					"throttled",
					"Rate limit exceeded. Please retry after the period indicated in the Retry-After header.",
				))
			}

			return next(c)
		}
	}
}
