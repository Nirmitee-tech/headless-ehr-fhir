package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/labstack/echo/v4"
)

// ---------------------------------------------------------------------------
// Data structures
// ---------------------------------------------------------------------------

// RatePlan defines rate limiting parameters for a tier of service.
type RatePlan struct {
	Name               string `json:"name"`
	RequestsPerMinute  int    `json:"requests_per_minute"`
	RequestsPerHour    int    `json:"requests_per_hour"`
	RequestsPerDay     int    `json:"requests_per_day"`
	BurstSize          int    `json:"burst_size"`
	ConcurrentRequests int    `json:"concurrent_requests"`
}

// RateLimitInfo is returned by Allow to communicate the decision and metadata.
type RateLimitInfo struct {
	Allowed    bool   `json:"allowed"`
	Remaining  int    `json:"remaining"`
	Limit      int    `json:"limit"`
	RetryAfter int    `json:"retry_after"`
	Plan       string `json:"plan"`
}

// ClientUsage exposes the current usage counters for a client.
type ClientUsage struct {
	ClientID        string `json:"client_id"`
	Plan            string `json:"plan"`
	MinuteUsed      int    `json:"minute_used"`
	MinuteLimit     int    `json:"minute_limit"`
	HourUsed        int    `json:"hour_used"`
	HourLimit       int    `json:"hour_limit"`
	DayUsed         int    `json:"day_used"`
	DayLimit        int    `json:"day_limit"`
	ConcurrentUsed  int    `json:"concurrent_used"`
	ConcurrentLimit int    `json:"concurrent_limit"`
}

// clientCounter tracks per-client request counts with atomic counters and
// time-window-based resets.
type clientCounter struct {
	minuteCount int64
	hourCount   int64
	dayCount    int64
	concurrent  int64
	minuteReset time.Time
	hourReset   time.Time
	dayReset    time.Time
	mu          sync.Mutex // protects reset times
}

// ClientRateLimiter provides thread-safe per-client rate limiting with
// multiple time windows and concurrent request tracking.
type ClientRateLimiter struct {
	plans       map[string]*RatePlan
	clientPlans map[string]string
	counters    map[string]*clientCounter
	mu          sync.RWMutex
}

// ---------------------------------------------------------------------------
// Default plans
// ---------------------------------------------------------------------------

// DefaultRatePlans returns the four predefined rate plans.
func DefaultRatePlans() []RatePlan {
	return []RatePlan{
		{
			Name:               "free",
			RequestsPerMinute:  60,
			RequestsPerHour:    1000,
			RequestsPerDay:     10000,
			BurstSize:          10,
			ConcurrentRequests: 5,
		},
		{
			Name:               "starter",
			RequestsPerMinute:  300,
			RequestsPerHour:    10000,
			RequestsPerDay:     100000,
			BurstSize:          30,
			ConcurrentRequests: 20,
		},
		{
			Name:               "professional",
			RequestsPerMinute:  1000,
			RequestsPerHour:    50000,
			RequestsPerDay:     500000,
			BurstSize:          100,
			ConcurrentRequests: 50,
		},
		{
			Name:               "enterprise",
			RequestsPerMinute:  5000,
			RequestsPerHour:    200000,
			RequestsPerDay:     2000000,
			BurstSize:          500,
			ConcurrentRequests: 200,
		},
	}
}

// ---------------------------------------------------------------------------
// Constructor
// ---------------------------------------------------------------------------

// NewClientRateLimiter creates a ClientRateLimiter pre-loaded with the four
// default rate plans.
func NewClientRateLimiter() *ClientRateLimiter {
	rl := &ClientRateLimiter{
		plans:       make(map[string]*RatePlan),
		clientPlans: make(map[string]string),
		counters:    make(map[string]*clientCounter),
	}
	for _, p := range DefaultRatePlans() {
		plan := p // copy
		rl.plans[plan.Name] = &plan
	}
	return rl
}

// ---------------------------------------------------------------------------
// Plan management
// ---------------------------------------------------------------------------

// RegisterPlan adds or replaces a rate plan by name.
func (rl *ClientRateLimiter) RegisterPlan(plan RatePlan) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	p := plan // copy
	rl.plans[p.Name] = &p
}

// AssignPlan assigns clientID to the named plan. Returns an error if the plan
// does not exist.
func (rl *ClientRateLimiter) AssignPlan(clientID, planName string) error {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	if _, ok := rl.plans[planName]; !ok {
		return fmt.Errorf("rate plan %q not found", planName)
	}
	rl.clientPlans[clientID] = planName
	return nil
}

// GetPlan returns the plan assigned to clientID, falling back to "free".
func (rl *ClientRateLimiter) GetPlan(clientID string) *RatePlan {
	rl.mu.RLock()
	defer rl.mu.RUnlock()
	planName, ok := rl.clientPlans[clientID]
	if !ok {
		planName = "free"
	}
	plan, ok := rl.plans[planName]
	if !ok {
		plan = rl.plans["free"]
	}
	return plan
}

// ---------------------------------------------------------------------------
// Counter helpers
// ---------------------------------------------------------------------------

// getOrCreateCounter returns the counter for clientID, creating one if needed.
// Caller must NOT hold rl.mu.
func (rl *ClientRateLimiter) getOrCreateCounter(clientID string) *clientCounter {
	rl.mu.RLock()
	c, ok := rl.counters[clientID]
	rl.mu.RUnlock()
	if ok {
		return c
	}

	rl.mu.Lock()
	defer rl.mu.Unlock()
	// Double-check
	if c, ok := rl.counters[clientID]; ok {
		return c
	}
	now := time.Now()
	c = &clientCounter{
		minuteReset: now.Add(time.Minute),
		hourReset:   now.Add(time.Hour),
		dayReset:    now.Add(24 * time.Hour),
	}
	rl.counters[clientID] = c
	return c
}

// maybeResetWindows checks and resets expired time windows. Must be called
// with counter.mu held.
func maybeResetWindows(c *clientCounter) {
	now := time.Now()
	if now.After(c.minuteReset) {
		atomic.StoreInt64(&c.minuteCount, 0)
		c.minuteReset = now.Add(time.Minute)
	}
	if now.After(c.hourReset) {
		atomic.StoreInt64(&c.hourCount, 0)
		c.hourReset = now.Add(time.Hour)
	}
	if now.After(c.dayReset) {
		atomic.StoreInt64(&c.dayCount, 0)
		c.dayReset = now.Add(24 * time.Hour)
	}
}

// ---------------------------------------------------------------------------
// Allow / Release
// ---------------------------------------------------------------------------

// Allow checks whether clientID may issue a new request. It atomically
// increments all counters and the concurrent gauge. The caller MUST call
// Release after the request completes to free the concurrent slot.
//
// The effective per-minute limit is RequestsPerMinute + BurstSize.
func (rl *ClientRateLimiter) Allow(clientID string) (bool, *RateLimitInfo) {
	plan := rl.GetPlan(clientID)
	counter := rl.getOrCreateCounter(clientID)

	// Reset expired windows under lock
	counter.mu.Lock()
	maybeResetWindows(counter)
	resetTime := counter.minuteReset
	counter.mu.Unlock()

	effectiveMinuteLimit := int64(plan.RequestsPerMinute + plan.BurstSize)
	info := &RateLimitInfo{
		Plan:  plan.Name,
		Limit: plan.RequestsPerMinute + plan.BurstSize,
	}

	// Check concurrent limit first (if configured)
	if plan.ConcurrentRequests > 0 {
		cur := atomic.LoadInt64(&counter.concurrent)
		if cur >= int64(plan.ConcurrentRequests) {
			info.Allowed = false
			info.Remaining = 0
			info.RetryAfter = 1 // retry quickly for concurrent
			return false, info
		}
	}

	// Check minute limit
	minuteVal := atomic.LoadInt64(&counter.minuteCount)
	if minuteVal >= effectiveMinuteLimit {
		info.Allowed = false
		info.Remaining = 0
		info.RetryAfter = secondsUntil(resetTime)
		return false, info
	}

	// Check hour limit
	hourVal := atomic.LoadInt64(&counter.hourCount)
	if hourVal >= int64(plan.RequestsPerHour) {
		info.Allowed = false
		info.Remaining = 0
		counter.mu.Lock()
		info.RetryAfter = secondsUntil(counter.hourReset)
		counter.mu.Unlock()
		return false, info
	}

	// Check day limit
	dayVal := atomic.LoadInt64(&counter.dayCount)
	if dayVal >= int64(plan.RequestsPerDay) {
		info.Allowed = false
		info.Remaining = 0
		counter.mu.Lock()
		info.RetryAfter = secondsUntil(counter.dayReset)
		counter.mu.Unlock()
		return false, info
	}

	// All checks passed — increment counters
	newMinute := atomic.AddInt64(&counter.minuteCount, 1)
	atomic.AddInt64(&counter.hourCount, 1)
	atomic.AddInt64(&counter.dayCount, 1)
	atomic.AddInt64(&counter.concurrent, 1)

	remaining := int(effectiveMinuteLimit - newMinute)
	if remaining < 0 {
		remaining = 0
	}

	info.Allowed = true
	info.Remaining = remaining
	return true, info
}

// Release decrements the concurrent request counter for clientID. It is safe
// to call even if Allow was never called (the counter will not go below zero).
func (rl *ClientRateLimiter) Release(clientID string) {
	counter := rl.getOrCreateCounter(clientID)
	for {
		cur := atomic.LoadInt64(&counter.concurrent)
		if cur <= 0 {
			return
		}
		if atomic.CompareAndSwapInt64(&counter.concurrent, cur, cur-1) {
			return
		}
	}
}

// ---------------------------------------------------------------------------
// Usage / Reset
// ---------------------------------------------------------------------------

// GetUsage returns a snapshot of the current counters for clientID.
func (rl *ClientRateLimiter) GetUsage(clientID string) *ClientUsage {
	plan := rl.GetPlan(clientID)
	counter := rl.getOrCreateCounter(clientID)

	counter.mu.Lock()
	maybeResetWindows(counter)
	counter.mu.Unlock()

	return &ClientUsage{
		ClientID:        clientID,
		Plan:            plan.Name,
		MinuteUsed:      int(atomic.LoadInt64(&counter.minuteCount)),
		MinuteLimit:     plan.RequestsPerMinute + plan.BurstSize,
		HourUsed:        int(atomic.LoadInt64(&counter.hourCount)),
		HourLimit:       plan.RequestsPerHour,
		DayUsed:         int(atomic.LoadInt64(&counter.dayCount)),
		DayLimit:        plan.RequestsPerDay,
		ConcurrentUsed:  int(atomic.LoadInt64(&counter.concurrent)),
		ConcurrentLimit: plan.ConcurrentRequests,
	}
}

// ResetCounters zeroes all rate-limit counters for clientID and resets the
// time windows.
func (rl *ClientRateLimiter) ResetCounters(clientID string) {
	counter := rl.getOrCreateCounter(clientID)
	counter.mu.Lock()
	defer counter.mu.Unlock()

	atomic.StoreInt64(&counter.minuteCount, 0)
	atomic.StoreInt64(&counter.hourCount, 0)
	atomic.StoreInt64(&counter.dayCount, 0)
	atomic.StoreInt64(&counter.concurrent, 0)

	now := time.Now()
	counter.minuteReset = now.Add(time.Minute)
	counter.hourReset = now.Add(time.Hour)
	counter.dayReset = now.Add(24 * time.Hour)
}

// StartCleanup removes stale counters (no requests in 24 h) on a periodic
// interval. It blocks until ctx is cancelled, so call it in a goroutine.
func (rl *ClientRateLimiter) StartCleanup(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			rl.mu.Lock()
			now := time.Now()
			for id, c := range rl.counters {
				c.mu.Lock()
				// If all windows have expired and no concurrent requests, remove.
				if now.After(c.dayReset) &&
					atomic.LoadInt64(&c.minuteCount) == 0 &&
					atomic.LoadInt64(&c.hourCount) == 0 &&
					atomic.LoadInt64(&c.dayCount) == 0 &&
					atomic.LoadInt64(&c.concurrent) == 0 {
					delete(rl.counters, id)
				}
				c.mu.Unlock()
			}
			rl.mu.Unlock()
		}
	}
}

// ---------------------------------------------------------------------------
// Echo middleware
// ---------------------------------------------------------------------------

// ClientRateLimitMiddleware returns an Echo middleware that enforces per-client
// rate limits. Client identity is resolved in priority order:
//  1. "api_key_id" context value (set by APIKeyMiddleware)
//  2. "client_id" context value
//  3. X-Client-ID request header
//  4. Client IP address (fallback)
func ClientRateLimitMiddleware(limiter *ClientRateLimiter) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			clientID := extractClientID(c)

			allowed, info := limiter.Allow(clientID)

			// Always set rate limit headers
			c.Response().Header().Set("X-RateLimit-Limit", strconv.Itoa(info.Limit))
			c.Response().Header().Set("X-RateLimit-Remaining", strconv.Itoa(info.Remaining))

			// Compute reset epoch
			limiter.mu.RLock()
			counter, ok := limiter.counters[clientID]
			limiter.mu.RUnlock()
			if ok {
				counter.mu.Lock()
				resetEpoch := counter.minuteReset.Unix()
				counter.mu.Unlock()
				c.Response().Header().Set("X-RateLimit-Reset", strconv.FormatInt(resetEpoch, 10))
			}

			if !allowed {
				c.Response().Header().Set("Retry-After", strconv.Itoa(info.RetryAfter))
				// Release the concurrent slot that was NOT acquired
				// (Allow does not increment concurrent on denial)
				return echo.NewHTTPError(http.StatusTooManyRequests, "rate limit exceeded")
			}

			// Execute handler, then release concurrent slot
			err := next(c)
			limiter.Release(clientID)
			return err
		}
	}
}

// extractClientID determines the client identifier from the echo context.
func extractClientID(c echo.Context) string {
	if v := c.Get("api_key_id"); v != nil {
		if s, ok := v.(string); ok && s != "" {
			return s
		}
	}
	if v := c.Get("client_id"); v != nil {
		if s, ok := v.(string); ok && s != "" {
			return s
		}
	}
	if h := c.Request().Header.Get("X-Client-ID"); h != "" {
		return h
	}
	return c.RealIP()
}

// ---------------------------------------------------------------------------
// Admin API handler
// ---------------------------------------------------------------------------

// RateLimitHandler exposes admin endpoints for managing rate limits.
type RateLimitHandler struct {
	limiter *ClientRateLimiter
}

// NewRateLimitHandler creates a handler backed by the given limiter.
func NewRateLimitHandler(limiter *ClientRateLimiter) *RateLimitHandler {
	return &RateLimitHandler{limiter: limiter}
}

// RegisterRoutes mounts the admin rate-limit endpoints on the given group.
func (h *RateLimitHandler) RegisterRoutes(g *echo.Group) {
	g.GET("/rate-limits/plans", h.ListPlans)
	g.POST("/rate-limits/plans", h.CreateOrUpdatePlan)
	g.GET("/rate-limits/clients/:id", h.GetClientUsage)
	g.PUT("/rate-limits/clients/:id/plan", h.AssignClientPlan)
	g.POST("/rate-limits/clients/:id/reset", h.ResetClientCounters)
}

// ListPlans returns all registered rate plans.
func (h *RateLimitHandler) ListPlans(c echo.Context) error {
	h.limiter.mu.RLock()
	plans := make([]RatePlan, 0, len(h.limiter.plans))
	for _, p := range h.limiter.plans {
		plans = append(plans, *p)
	}
	h.limiter.mu.RUnlock()
	return c.JSON(http.StatusOK, plans)
}

// CreateOrUpdatePlan creates or replaces a rate plan from the request body.
func (h *RateLimitHandler) CreateOrUpdatePlan(c echo.Context) error {
	var plan RatePlan
	if err := c.Bind(&plan); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid plan: "+err.Error())
	}
	if plan.Name == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "plan name is required")
	}
	h.limiter.RegisterPlan(plan)
	return c.JSON(http.StatusOK, plan)
}

// GetClientUsage returns current usage stats for a client.
func (h *RateLimitHandler) GetClientUsage(c echo.Context) error {
	clientID := c.Param("id")
	usage := h.limiter.GetUsage(clientID)
	return c.JSON(http.StatusOK, usage)
}

// AssignClientPlan assigns a rate plan to a client.
func (h *RateLimitHandler) AssignClientPlan(c echo.Context) error {
	clientID := c.Param("id")
	var body struct {
		Plan string `json:"plan"`
	}
	if err := c.Bind(&body); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body: "+err.Error())
	}
	if err := h.limiter.AssignPlan(clientID, body.Plan); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]string{
		"client_id": clientID,
		"plan":      body.Plan,
	})
}

// ResetClientCounters zeroes all counters for a client.
func (h *RateLimitHandler) ResetClientCounters(c echo.Context) error {
	clientID := c.Param("id")
	h.limiter.ResetCounters(clientID)
	return c.JSON(http.StatusOK, map[string]string{
		"client_id": clientID,
		"status":    "reset",
	})
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// secondsUntil returns the number of seconds from now until t, minimum 1.
func secondsUntil(t time.Time) int {
	d := time.Until(t)
	s := int(d.Seconds())
	if s < 1 {
		return 1
	}
	return s
}
