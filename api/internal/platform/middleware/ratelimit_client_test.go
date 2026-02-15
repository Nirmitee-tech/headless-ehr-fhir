package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
)

// ---------------------------------------------------------------------------
// Plan tests
// ---------------------------------------------------------------------------

func TestDefaultRatePlans(t *testing.T) {
	plans := DefaultRatePlans()
	if len(plans) != 4 {
		t.Fatalf("expected 4 default plans, got %d", len(plans))
	}

	tests := []struct {
		name              string
		requestsPerMinute int
		requestsPerHour   int
		requestsPerDay    int
		burstSize         int
		concurrent        int
	}{
		{"free", 60, 1000, 10000, 10, 5},
		{"starter", 300, 10000, 100000, 30, 20},
		{"professional", 1000, 50000, 500000, 100, 50},
		{"enterprise", 5000, 200000, 2000000, 500, 200},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var found *RatePlan
			for i := range plans {
				if plans[i].Name == tt.name {
					found = &plans[i]
					break
				}
			}
			if found == nil {
				t.Fatalf("plan %q not found", tt.name)
			}
			if found.RequestsPerMinute != tt.requestsPerMinute {
				t.Errorf("RequestsPerMinute: expected %d, got %d", tt.requestsPerMinute, found.RequestsPerMinute)
			}
			if found.RequestsPerHour != tt.requestsPerHour {
				t.Errorf("RequestsPerHour: expected %d, got %d", tt.requestsPerHour, found.RequestsPerHour)
			}
			if found.RequestsPerDay != tt.requestsPerDay {
				t.Errorf("RequestsPerDay: expected %d, got %d", tt.requestsPerDay, found.RequestsPerDay)
			}
			if found.BurstSize != tt.burstSize {
				t.Errorf("BurstSize: expected %d, got %d", tt.burstSize, found.BurstSize)
			}
			if found.ConcurrentRequests != tt.concurrent {
				t.Errorf("ConcurrentRequests: expected %d, got %d", tt.concurrent, found.ConcurrentRequests)
			}
		})
	}
}

func TestClientRateLimiter_RegisterPlan(t *testing.T) {
	rl := NewClientRateLimiter()
	custom := RatePlan{
		Name:               "custom",
		RequestsPerMinute:  42,
		RequestsPerHour:    420,
		RequestsPerDay:     4200,
		BurstSize:          5,
		ConcurrentRequests: 3,
	}
	rl.RegisterPlan(custom)

	plan := rl.GetPlan("some-client")
	// Still default "free" since client not assigned
	if plan.Name != "free" {
		t.Errorf("expected free plan for unassigned client, got %s", plan.Name)
	}

	// Assign and verify
	err := rl.AssignPlan("some-client", "custom")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	plan = rl.GetPlan("some-client")
	if plan.Name != "custom" {
		t.Errorf("expected custom plan, got %s", plan.Name)
	}
	if plan.RequestsPerMinute != 42 {
		t.Errorf("expected 42 rpm, got %d", plan.RequestsPerMinute)
	}
}

func TestClientRateLimiter_AssignPlan(t *testing.T) {
	rl := NewClientRateLimiter()

	// Assign to existing plan
	err := rl.AssignPlan("client-1", "starter")
	if err != nil {
		t.Fatalf("unexpected error assigning starter: %v", err)
	}
	plan := rl.GetPlan("client-1")
	if plan.Name != "starter" {
		t.Errorf("expected starter, got %s", plan.Name)
	}

	// Assign to non-existent plan
	err = rl.AssignPlan("client-1", "nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent plan")
	}
}

func TestClientRateLimiter_GetPlan_Default(t *testing.T) {
	rl := NewClientRateLimiter()
	plan := rl.GetPlan("unknown-client")
	if plan == nil {
		t.Fatal("expected non-nil plan for unknown client")
	}
	if plan.Name != "free" {
		t.Errorf("expected free plan for unknown client, got %s", plan.Name)
	}
}

// ---------------------------------------------------------------------------
// Allow tests
// ---------------------------------------------------------------------------

func TestClientRateLimiter_Allow_UnderLimit(t *testing.T) {
	rl := NewClientRateLimiter()
	// Free plan: 60/min
	allowed, info := rl.Allow("client-a")
	if !allowed {
		t.Fatal("expected request to be allowed")
	}
	if !info.Allowed {
		t.Fatal("expected info.Allowed to be true")
	}
	if info.Plan != "free" {
		t.Errorf("expected plan 'free', got %s", info.Plan)
	}
	if info.Limit != 70 {
		t.Errorf("expected limit 70 (60 RPM + 10 burst), got %d", info.Limit)
	}
	// After 1 request, remaining should be 60+10(burst)-1 = 69
	if info.Remaining < 0 {
		t.Errorf("expected non-negative remaining, got %d", info.Remaining)
	}
	rl.Release("client-a")
}

func TestClientRateLimiter_Allow_AtMinuteLimit(t *testing.T) {
	rl := NewClientRateLimiter()
	// Create a tiny plan for testing
	rl.RegisterPlan(RatePlan{
		Name:               "tiny",
		RequestsPerMinute:  3,
		RequestsPerHour:    1000,
		RequestsPerDay:     10000,
		BurstSize:          0,
		ConcurrentRequests: 0,
	})
	rl.AssignPlan("client-b", "tiny")

	// Use all 3 requests
	for i := 0; i < 3; i++ {
		allowed, _ := rl.Allow("client-b")
		if !allowed {
			t.Fatalf("request %d should be allowed", i+1)
		}
		rl.Release("client-b")
	}

	// 4th should fail
	allowed, info := rl.Allow("client-b")
	if allowed {
		t.Fatal("expected request to be blocked at minute limit")
	}
	if info.Allowed {
		t.Fatal("expected info.Allowed to be false")
	}
	if info.Remaining != 0 {
		t.Errorf("expected remaining 0, got %d", info.Remaining)
	}
}

func TestClientRateLimiter_Allow_AtHourLimit(t *testing.T) {
	rl := NewClientRateLimiter()
	rl.RegisterPlan(RatePlan{
		Name:               "hour-test",
		RequestsPerMinute:  1000,
		RequestsPerHour:    5,
		RequestsPerDay:     10000,
		BurstSize:          0,
		ConcurrentRequests: 0,
	})
	rl.AssignPlan("client-c", "hour-test")

	for i := 0; i < 5; i++ {
		allowed, _ := rl.Allow("client-c")
		if !allowed {
			t.Fatalf("request %d should be allowed", i+1)
		}
		rl.Release("client-c")
	}

	allowed, _ := rl.Allow("client-c")
	if allowed {
		t.Fatal("expected request to be blocked at hour limit")
	}
}

func TestClientRateLimiter_Allow_AtDayLimit(t *testing.T) {
	rl := NewClientRateLimiter()
	rl.RegisterPlan(RatePlan{
		Name:               "day-test",
		RequestsPerMinute:  1000,
		RequestsPerHour:    1000,
		RequestsPerDay:     4,
		BurstSize:          0,
		ConcurrentRequests: 0,
	})
	rl.AssignPlan("client-d", "day-test")

	for i := 0; i < 4; i++ {
		allowed, _ := rl.Allow("client-d")
		if !allowed {
			t.Fatalf("request %d should be allowed", i+1)
		}
		rl.Release("client-d")
	}

	allowed, _ := rl.Allow("client-d")
	if allowed {
		t.Fatal("expected request to be blocked at day limit")
	}
}

func TestClientRateLimiter_Allow_BurstAllowed(t *testing.T) {
	rl := NewClientRateLimiter()
	rl.RegisterPlan(RatePlan{
		Name:               "burst-test",
		RequestsPerMinute:  5,
		RequestsPerHour:    10000,
		RequestsPerDay:     100000,
		BurstSize:          3,
		ConcurrentRequests: 0,
	})
	rl.AssignPlan("client-e", "burst-test")

	// Should allow 5+3 = 8 requests (sustained + burst)
	for i := 0; i < 8; i++ {
		allowed, _ := rl.Allow("client-e")
		if !allowed {
			t.Fatalf("request %d should be allowed (within burst)", i+1)
		}
		rl.Release("client-e")
	}

	// 9th should be blocked
	allowed, _ := rl.Allow("client-e")
	if allowed {
		t.Fatal("expected request to be blocked after burst exhausted")
	}
}

func TestClientRateLimiter_Allow_ConcurrentLimit(t *testing.T) {
	rl := NewClientRateLimiter()
	rl.RegisterPlan(RatePlan{
		Name:               "conc-test",
		RequestsPerMinute:  1000,
		RequestsPerHour:    100000,
		RequestsPerDay:     1000000,
		BurstSize:          0,
		ConcurrentRequests: 2,
	})
	rl.AssignPlan("client-f", "conc-test")

	// Take 2 concurrent slots (don't release)
	allowed1, _ := rl.Allow("client-f")
	if !allowed1 {
		t.Fatal("first concurrent request should be allowed")
	}
	allowed2, _ := rl.Allow("client-f")
	if !allowed2 {
		t.Fatal("second concurrent request should be allowed")
	}

	// 3rd should be blocked (concurrent limit reached)
	allowed3, _ := rl.Allow("client-f")
	if allowed3 {
		t.Fatal("expected request to be blocked at concurrent limit")
	}

	// Release one, then try again
	rl.Release("client-f")
	allowed4, _ := rl.Allow("client-f")
	if !allowed4 {
		t.Fatal("expected request to be allowed after release")
	}

	// Clean up
	rl.Release("client-f")
	rl.Release("client-f")
}

func TestClientRateLimiter_Allow_WindowReset(t *testing.T) {
	rl := NewClientRateLimiter()
	rl.RegisterPlan(RatePlan{
		Name:               "reset-test",
		RequestsPerMinute:  2,
		RequestsPerHour:    10000,
		RequestsPerDay:     100000,
		BurstSize:          0,
		ConcurrentRequests: 0,
	})
	rl.AssignPlan("client-g", "reset-test")

	// Use up quota
	for i := 0; i < 2; i++ {
		allowed, _ := rl.Allow("client-g")
		if !allowed {
			t.Fatalf("request %d should be allowed", i+1)
		}
		rl.Release("client-g")
	}

	// Should be blocked
	allowed, _ := rl.Allow("client-g")
	if allowed {
		t.Fatal("expected block at minute limit")
	}

	// Manually set the minute reset to the past to simulate window expiry
	rl.mu.RLock()
	counter := rl.counters["client-g"]
	rl.mu.RUnlock()

	counter.mu.Lock()
	counter.minuteReset = time.Now().Add(-1 * time.Second)
	counter.mu.Unlock()

	// Now should be allowed again
	allowed, _ = rl.Allow("client-g")
	if !allowed {
		t.Fatal("expected request to be allowed after window reset")
	}
	rl.Release("client-g")
}

func TestClientRateLimiter_Allow_DifferentClients(t *testing.T) {
	rl := NewClientRateLimiter()
	rl.RegisterPlan(RatePlan{
		Name:               "diff-test",
		RequestsPerMinute:  1,
		RequestsPerHour:    10000,
		RequestsPerDay:     100000,
		BurstSize:          0,
		ConcurrentRequests: 0,
	})
	rl.AssignPlan("client-x", "diff-test")
	rl.AssignPlan("client-y", "diff-test")

	// client-x uses their 1 request
	allowed, _ := rl.Allow("client-x")
	if !allowed {
		t.Fatal("client-x first request should be allowed")
	}
	rl.Release("client-x")

	// client-x is now blocked
	allowed, _ = rl.Allow("client-x")
	if allowed {
		t.Fatal("client-x second request should be blocked")
	}

	// client-y should still be allowed (separate counter)
	allowed, _ = rl.Allow("client-y")
	if !allowed {
		t.Fatal("client-y first request should be allowed")
	}
	rl.Release("client-y")
}

func TestClientRateLimiter_Allow_RetryAfterCalculation(t *testing.T) {
	rl := NewClientRateLimiter()
	rl.RegisterPlan(RatePlan{
		Name:               "retry-test",
		RequestsPerMinute:  1,
		RequestsPerHour:    10000,
		RequestsPerDay:     100000,
		BurstSize:          0,
		ConcurrentRequests: 0,
	})
	rl.AssignPlan("client-r", "retry-test")

	// Use up quota
	rl.Allow("client-r")
	rl.Release("client-r")

	// Should get retry-after info
	allowed, info := rl.Allow("client-r")
	if allowed {
		t.Fatal("expected block")
	}
	if info.RetryAfter <= 0 {
		t.Errorf("expected positive RetryAfter, got %d", info.RetryAfter)
	}
	if info.RetryAfter > 60 {
		t.Errorf("expected RetryAfter <= 60s, got %d", info.RetryAfter)
	}
}

// ---------------------------------------------------------------------------
// Release tests
// ---------------------------------------------------------------------------

func TestClientRateLimiter_Release(t *testing.T) {
	rl := NewClientRateLimiter()
	rl.RegisterPlan(RatePlan{
		Name:               "release-test",
		RequestsPerMinute:  1000,
		RequestsPerHour:    100000,
		RequestsPerDay:     1000000,
		BurstSize:          0,
		ConcurrentRequests: 2,
	})
	rl.AssignPlan("client-rel", "release-test")

	rl.Allow("client-rel")
	rl.Allow("client-rel")

	usage := rl.GetUsage("client-rel")
	if usage.ConcurrentUsed != 2 {
		t.Errorf("expected 2 concurrent, got %d", usage.ConcurrentUsed)
	}

	rl.Release("client-rel")
	usage = rl.GetUsage("client-rel")
	if usage.ConcurrentUsed != 1 {
		t.Errorf("expected 1 concurrent after release, got %d", usage.ConcurrentUsed)
	}

	rl.Release("client-rel")
	usage = rl.GetUsage("client-rel")
	if usage.ConcurrentUsed != 0 {
		t.Errorf("expected 0 concurrent after 2 releases, got %d", usage.ConcurrentUsed)
	}
}

func TestClientRateLimiter_Release_NeverNegative(t *testing.T) {
	rl := NewClientRateLimiter()
	// Release without any Allow calls
	rl.Release("phantom-client")
	rl.Release("phantom-client")

	usage := rl.GetUsage("phantom-client")
	if usage.ConcurrentUsed < 0 {
		t.Errorf("concurrent should never be negative, got %d", usage.ConcurrentUsed)
	}
}

// ---------------------------------------------------------------------------
// Usage tests
// ---------------------------------------------------------------------------

func TestClientRateLimiter_GetUsage(t *testing.T) {
	rl := NewClientRateLimiter()
	rl.RegisterPlan(RatePlan{
		Name:               "usage-test",
		RequestsPerMinute:  100,
		RequestsPerHour:    1000,
		RequestsPerDay:     10000,
		BurstSize:          10,
		ConcurrentRequests: 5,
	})
	rl.AssignPlan("client-u", "usage-test")

	// Make 3 requests, release 1
	rl.Allow("client-u")
	rl.Allow("client-u")
	rl.Allow("client-u")
	rl.Release("client-u")

	usage := rl.GetUsage("client-u")
	if usage.ClientID != "client-u" {
		t.Errorf("expected client-u, got %s", usage.ClientID)
	}
	if usage.Plan != "usage-test" {
		t.Errorf("expected usage-test plan, got %s", usage.Plan)
	}
	if usage.MinuteUsed != 3 {
		t.Errorf("expected 3 minute used, got %d", usage.MinuteUsed)
	}
	if usage.MinuteLimit != 110 { // 100 + 10 burst
		t.Errorf("expected 110 minute limit (100+10 burst), got %d", usage.MinuteLimit)
	}
	if usage.HourUsed != 3 {
		t.Errorf("expected 3 hour used, got %d", usage.HourUsed)
	}
	if usage.HourLimit != 1000 {
		t.Errorf("expected 1000 hour limit, got %d", usage.HourLimit)
	}
	if usage.DayUsed != 3 {
		t.Errorf("expected 3 day used, got %d", usage.DayUsed)
	}
	if usage.DayLimit != 10000 {
		t.Errorf("expected 10000 day limit, got %d", usage.DayLimit)
	}
	if usage.ConcurrentUsed != 2 {
		t.Errorf("expected 2 concurrent used, got %d", usage.ConcurrentUsed)
	}
	if usage.ConcurrentLimit != 5 {
		t.Errorf("expected 5 concurrent limit, got %d", usage.ConcurrentLimit)
	}
}

func TestClientRateLimiter_ResetCounters(t *testing.T) {
	rl := NewClientRateLimiter()
	rl.RegisterPlan(RatePlan{
		Name:               "reset-ctr",
		RequestsPerMinute:  10,
		RequestsPerHour:    100,
		RequestsPerDay:     1000,
		BurstSize:          0,
		ConcurrentRequests: 0,
	})
	rl.AssignPlan("client-reset", "reset-ctr")

	// Make some requests
	for i := 0; i < 5; i++ {
		rl.Allow("client-reset")
		rl.Release("client-reset")
	}

	usage := rl.GetUsage("client-reset")
	if usage.MinuteUsed != 5 {
		t.Fatalf("expected 5 minute used before reset, got %d", usage.MinuteUsed)
	}

	rl.ResetCounters("client-reset")

	usage = rl.GetUsage("client-reset")
	if usage.MinuteUsed != 0 {
		t.Errorf("expected 0 minute used after reset, got %d", usage.MinuteUsed)
	}
	if usage.HourUsed != 0 {
		t.Errorf("expected 0 hour used after reset, got %d", usage.HourUsed)
	}
	if usage.DayUsed != 0 {
		t.Errorf("expected 0 day used after reset, got %d", usage.DayUsed)
	}
}

func TestClientRateLimiter_ConcurrentAccess(t *testing.T) {
	rl := NewClientRateLimiter()
	rl.RegisterPlan(RatePlan{
		Name:               "conc-safe",
		RequestsPerMinute:  10000,
		RequestsPerHour:    100000,
		RequestsPerDay:     1000000,
		BurstSize:          0,
		ConcurrentRequests: 0, // unlimited concurrent
	})
	rl.AssignPlan("client-conc", "conc-safe")

	var wg sync.WaitGroup
	var allowedCount int64

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			allowed, _ := rl.Allow("client-conc")
			if allowed {
				atomic.AddInt64(&allowedCount, 1)
				rl.Release("client-conc")
			}
		}()
	}

	wg.Wait()

	// All 100 should be allowed (limit is 10000/min)
	if allowedCount != 100 {
		t.Errorf("expected 100 allowed, got %d", allowedCount)
	}

	usage := rl.GetUsage("client-conc")
	if usage.MinuteUsed != 100 {
		t.Errorf("expected 100 minute used, got %d", usage.MinuteUsed)
	}
	if usage.ConcurrentUsed != 0 {
		t.Errorf("expected 0 concurrent after all released, got %d", usage.ConcurrentUsed)
	}
}

// ---------------------------------------------------------------------------
// Middleware tests
// ---------------------------------------------------------------------------

func TestClientRateLimitMiddleware_Allowed(t *testing.T) {
	rl := NewClientRateLimiter()
	e := echo.New()
	mw := ClientRateLimitMiddleware(rl)
	handler := mw(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Client-ID", "mw-client-1")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	// Check rate limit headers
	limitHeader := rec.Header().Get("X-RateLimit-Limit")
	if limitHeader == "" {
		t.Error("expected X-RateLimit-Limit header")
	}
	remainingHeader := rec.Header().Get("X-RateLimit-Remaining")
	if remainingHeader == "" {
		t.Error("expected X-RateLimit-Remaining header")
	}
}

func TestClientRateLimitMiddleware_Blocked(t *testing.T) {
	rl := NewClientRateLimiter()
	rl.RegisterPlan(RatePlan{
		Name:               "mw-tiny",
		RequestsPerMinute:  1,
		RequestsPerHour:    1000,
		RequestsPerDay:     10000,
		BurstSize:          0,
		ConcurrentRequests: 0,
	})
	rl.AssignPlan("mw-client-2", "mw-tiny")

	e := echo.New()
	mw := ClientRateLimitMiddleware(rl)
	handler := mw(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	// First request should pass
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Client-ID", "mw-client-2")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := handler(c)
	if err != nil {
		t.Fatalf("first request: unexpected error: %v", err)
	}

	// Second request should be blocked
	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	req2.Header.Set("X-Client-ID", "mw-client-2")
	rec2 := httptest.NewRecorder()
	c2 := e.NewContext(req2, rec2)
	err = handler(c2)

	if err == nil {
		t.Fatal("expected error for rate-limited request")
	}
	httpErr, ok := err.(*echo.HTTPError)
	if !ok {
		t.Fatalf("expected echo.HTTPError, got %T", err)
	}
	if httpErr.Code != http.StatusTooManyRequests {
		t.Errorf("expected 429, got %d", httpErr.Code)
	}

	retryAfter := rec2.Header().Get("Retry-After")
	if retryAfter == "" {
		t.Error("expected Retry-After header on blocked response")
	}
}

func TestClientRateLimitMiddleware_ExtractsAPIKeyID(t *testing.T) {
	rl := NewClientRateLimiter()
	rl.RegisterPlan(RatePlan{
		Name:               "apikey-test",
		RequestsPerMinute:  2,
		RequestsPerHour:    10000,
		RequestsPerDay:     100000,
		BurstSize:          0,
		ConcurrentRequests: 0,
	})
	rl.AssignPlan("api-key-123", "apikey-test")

	e := echo.New()
	mw := ClientRateLimitMiddleware(rl)
	handler := mw(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	// Request with api_key_id in context
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("api_key_id", "api-key-123")

	err := handler(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify usage is on the api-key-123 client
	usage := rl.GetUsage("api-key-123")
	if usage.MinuteUsed != 1 {
		t.Errorf("expected 1 minute used for api-key-123, got %d", usage.MinuteUsed)
	}
}

func TestClientRateLimitMiddleware_FallsBackToIP(t *testing.T) {
	rl := NewClientRateLimiter()
	e := echo.New()
	mw := ClientRateLimitMiddleware(rl)
	handler := mw(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	// No client ID headers or context values — should use IP
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// RealIP returns the remote addr for test requests
	clientIP := c.RealIP()
	usage := rl.GetUsage(clientIP)
	if usage.MinuteUsed != 1 {
		t.Errorf("expected 1 minute used for IP %s, got %d", clientIP, usage.MinuteUsed)
	}
}

func TestClientRateLimitMiddleware_SetsHeaders(t *testing.T) {
	rl := NewClientRateLimiter()
	e := echo.New()
	mw := ClientRateLimitMiddleware(rl)
	handler := mw(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Client-ID", "header-client")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	headers := []string{"X-RateLimit-Limit", "X-RateLimit-Remaining", "X-RateLimit-Reset"}
	for _, h := range headers {
		if rec.Header().Get(h) == "" {
			t.Errorf("expected %s header to be set", h)
		}
	}
}

func TestClientRateLimitMiddleware_ReleasesOnComplete(t *testing.T) {
	rl := NewClientRateLimiter()
	rl.RegisterPlan(RatePlan{
		Name:               "release-mw",
		RequestsPerMinute:  1000,
		RequestsPerHour:    100000,
		RequestsPerDay:     1000000,
		BurstSize:          0,
		ConcurrentRequests: 5,
	})
	rl.AssignPlan("release-client", "release-mw")

	e := echo.New()
	mw := ClientRateLimitMiddleware(rl)
	handler := mw(func(c echo.Context) error {
		// During handler execution, concurrent should be 1
		usage := rl.GetUsage("release-client")
		if usage.ConcurrentUsed != 1 {
			t.Errorf("expected 1 concurrent during handler, got %d", usage.ConcurrentUsed)
		}
		return c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Client-ID", "release-client")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// After handler completes, concurrent should be back to 0
	usage := rl.GetUsage("release-client")
	if usage.ConcurrentUsed != 0 {
		t.Errorf("expected 0 concurrent after handler, got %d", usage.ConcurrentUsed)
	}
}

// ---------------------------------------------------------------------------
// Handler tests
// ---------------------------------------------------------------------------

func TestRateLimitHandler_ListPlans(t *testing.T) {
	rl := NewClientRateLimiter()
	h := NewRateLimitHandler(rl)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/admin/rate-limits/plans", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.ListPlans(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var plans []RatePlan
	if err := json.Unmarshal(rec.Body.Bytes(), &plans); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if len(plans) < 4 {
		t.Errorf("expected at least 4 plans, got %d", len(plans))
	}
}

func TestRateLimitHandler_GetClientUsage(t *testing.T) {
	rl := NewClientRateLimiter()
	h := NewRateLimitHandler(rl)

	// Make a request first
	rl.Allow("usage-client")
	rl.Release("usage-client")

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/admin/rate-limits/clients/usage-client", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("usage-client")

	err := h.GetClientUsage(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var usage ClientUsage
	if err := json.Unmarshal(rec.Body.Bytes(), &usage); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if usage.ClientID != "usage-client" {
		t.Errorf("expected client ID 'usage-client', got %s", usage.ClientID)
	}
	if usage.MinuteUsed != 1 {
		t.Errorf("expected 1 minute used, got %d", usage.MinuteUsed)
	}
}

func TestRateLimitHandler_AssignPlan(t *testing.T) {
	rl := NewClientRateLimiter()
	h := NewRateLimitHandler(rl)

	e := echo.New()
	body := `{"plan":"starter"}`
	req := httptest.NewRequest(http.MethodPut, "/admin/rate-limits/clients/assign-client/plan", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("assign-client")

	err := h.AssignClientPlan(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	plan := rl.GetPlan("assign-client")
	if plan.Name != "starter" {
		t.Errorf("expected starter plan, got %s", plan.Name)
	}
}

func TestRateLimitHandler_ResetCounters(t *testing.T) {
	rl := NewClientRateLimiter()
	h := NewRateLimitHandler(rl)

	// Make some requests
	rl.Allow("reset-handler-client")
	rl.Allow("reset-handler-client")
	rl.Release("reset-handler-client")
	rl.Release("reset-handler-client")

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/admin/rate-limits/clients/reset-handler-client/reset", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("reset-handler-client")

	err := h.ResetClientCounters(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	usage := rl.GetUsage("reset-handler-client")
	if usage.MinuteUsed != 0 {
		t.Errorf("expected 0 minute used after reset, got %d", usage.MinuteUsed)
	}
}
