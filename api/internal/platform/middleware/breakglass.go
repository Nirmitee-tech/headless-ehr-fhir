package middleware

import (
	"context"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"

	"github.com/ehr/ehr/internal/platform/auth"
)

// breakGlassContextKey is the unexported type used for break-glass context values
// to avoid collisions with other packages.
type breakGlassContextKey string

const (
	breakGlassKey       breakGlassContextKey = "break_glass"
	breakGlassReasonKey breakGlassContextKey = "break_glass_reason"
)

// breakGlassRateLimit tracks per-user request counts within a rolling window.
type breakGlassRateLimit struct {
	mu      sync.Mutex
	entries map[string][]time.Time // userID -> list of request timestamps
}

// newBreakGlassRateLimit creates a new rate limiter for break-glass requests.
func newBreakGlassRateLimit() *breakGlassRateLimit {
	return &breakGlassRateLimit{
		entries: make(map[string][]time.Time),
	}
}

// allow checks whether the user is under the break-glass rate limit.
// It keeps only timestamps within the last hour and enforces a maximum of
// maxPerHour requests. If the request is allowed, the current timestamp is
// recorded. The caller must supply the current time so that tests can inject
// a deterministic clock.
func (rl *breakGlassRateLimit) allow(userID string, now time.Time, maxPerHour int) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	cutoff := now.Add(-1 * time.Hour)

	// Prune expired entries for this user.
	existing := rl.entries[userID]
	pruned := existing[:0]
	for _, ts := range existing {
		if ts.After(cutoff) {
			pruned = append(pruned, ts)
		}
	}

	if len(pruned) >= maxPerHour {
		rl.entries[userID] = pruned
		return false
	}

	rl.entries[userID] = append(pruned, now)
	return true
}

// cleanup removes all entries older than one hour. This can be called
// periodically from a background goroutine to prevent unbounded memory growth.
func (rl *breakGlassRateLimit) cleanup(now time.Time) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	cutoff := now.Add(-1 * time.Hour)
	for userID, timestamps := range rl.entries {
		pruned := timestamps[:0]
		for _, ts := range timestamps {
			if ts.After(cutoff) {
				pruned = append(pruned, ts)
			}
		}
		if len(pruned) == 0 {
			delete(rl.entries, userID)
		} else {
			rl.entries[userID] = pruned
		}
	}
}

const (
	breakGlassMaxPerHour    = 10
	breakGlassCleanupPeriod = 5 * time.Minute
)

// isClinicalPath returns true if the request path is under /fhir/ or /api/v1/.
func isClinicalPath(path string) bool {
	return strings.HasPrefix(path, "/fhir/") || strings.HasPrefix(path, "/api/v1/")
}

// BreakGlass returns Echo middleware that implements the emergency break-glass
// override for clinical data access. When a request includes the X-Break-Glass
// header with a non-empty reason string, the middleware:
//
//  1. Verifies the user is authenticated (user_id present in context).
//  2. Enforces a per-user rate limit (10 requests per hour).
//  3. Injects "admin" into the user's roles so that downstream RBAC, ABAC, and
//     consent enforcement middleware pass.
//  4. Sets require_consent = false on the echo context so that
//     ConsentEnforcementMiddleware skips.
//  5. Stores break_glass and break_glass_reason in the request context for
//     downstream handlers and audit logging.
//  6. Emits a WARN-level structured log entry.
//
// The middleware only activates on clinical paths (/fhir/* and /api/v1/*).
// It must be placed AFTER authentication middleware and BEFORE ABAC/consent
// middleware in the middleware chain.
func BreakGlass(logger zerolog.Logger) echo.MiddlewareFunc {
	rl := newBreakGlassRateLimit()

	// Background cleanup goroutine to prevent unbounded memory growth.
	go func() {
		ticker := time.NewTicker(breakGlassCleanupPeriod)
		defer ticker.Stop()
		for range ticker.C {
			rl.cleanup(time.Now())
		}
	}()

	return breakGlassMiddleware(logger, rl, time.Now)
}

// breakGlassMiddleware is the internal constructor that accepts a clock function
// for testing determinism and a pre-built rate limiter.
func breakGlassMiddleware(logger zerolog.Logger, rl *breakGlassRateLimit, nowFn func() time.Time) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			path := req.URL.Path

			// Only activate on clinical paths.
			if !isClinicalPath(path) {
				return next(c)
			}

			reason := strings.TrimSpace(req.Header.Get("X-Break-Glass"))
			if reason == "" {
				return next(c)
			}

			// The user must be authenticated. The JWT middleware runs before
			// this middleware and sets user_id in the request context. If
			// user_id is absent, the user is not authenticated.
			ctx := req.Context()
			userID := auth.UserIDFromContext(ctx)
			if userID == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "break-glass requires authentication")
			}

			// Rate limiting.
			now := nowFn()
			if !rl.allow(userID, now, breakGlassMaxPerHour) {
				return echo.NewHTTPError(http.StatusTooManyRequests,
					"break-glass rate limit exceeded: maximum 10 requests per user per hour")
			}

			// Retrieve current roles and append "admin" if not already present.
			roles := auth.RolesFromContext(ctx)
			hasAdmin := false
			for _, r := range roles {
				if r == "admin" {
					hasAdmin = true
					break
				}
			}
			if !hasAdmin {
				roles = append(roles, "admin")
			}

			// Update the request context with break-glass flags and elevated roles.
			ctx = context.WithValue(ctx, breakGlassKey, true)
			ctx = context.WithValue(ctx, breakGlassReasonKey, reason)
			ctx = context.WithValue(ctx, auth.UserRolesKey, roles)
			c.SetRequest(req.WithContext(ctx))

			// Disable consent enforcement for this request.
			c.Set("require_consent", false)

			// Structured WARN log for audit trail.
			logger.Warn().
				Str("type", "break_glass").
				Str("user_id", userID).
				Strs("original_roles", auth.RolesFromContext(req.Context())).
				Str("break_glass_reason", reason).
				Str("path", path).
				Str("method", req.Method).
				Str("remote_ip", c.RealIP()).
				Time("timestamp", now).
				Msg("break_glass_override")

			return next(c)
		}
	}
}

// ---------------------------------------------------------------------------
// Context helpers
// ---------------------------------------------------------------------------

// IsBreakGlass returns true if the request is a break-glass override.
func IsBreakGlass(ctx context.Context) bool {
	v, _ := ctx.Value(breakGlassKey).(bool)
	return v
}

// BreakGlassReason returns the reason string provided in the X-Break-Glass
// header, or an empty string if break-glass was not invoked.
func BreakGlassReason(ctx context.Context) string {
	v, _ := ctx.Value(breakGlassReasonKey).(string)
	return v
}
