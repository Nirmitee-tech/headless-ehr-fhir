package fhir

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
)

// ---------------------------------------------------------------------------
// SlidingWindowLimiter unit tests
// ---------------------------------------------------------------------------

func TestSlidingWindowLimiter_AllowWithinLimit(t *testing.T) {
	limiter := NewSlidingWindowLimiter(3, time.Minute)

	for i := 0; i < 3; i++ {
		allowed, remaining, _ := limiter.Allow("client-a")
		if !allowed {
			t.Fatalf("request %d: expected allowed", i+1)
		}
		want := 3 - (i + 1)
		if remaining != want {
			t.Errorf("request %d: remaining = %d, want %d", i+1, remaining, want)
		}
	}
}

func TestSlidingWindowLimiter_DenyWhenExceeded(t *testing.T) {
	limiter := NewSlidingWindowLimiter(2, time.Minute)

	limiter.Allow("client-a")
	limiter.Allow("client-a")

	allowed, remaining, _ := limiter.Allow("client-a")
	if allowed {
		t.Fatal("expected request to be denied")
	}
	if remaining != 0 {
		t.Errorf("remaining = %d, want 0", remaining)
	}
}

func TestSlidingWindowLimiter_SeparateClients(t *testing.T) {
	limiter := NewSlidingWindowLimiter(1, time.Minute)

	allowed1, _, _ := limiter.Allow("client-a")
	allowed2, _, _ := limiter.Allow("client-b")

	if !allowed1 {
		t.Error("client-a should be allowed")
	}
	if !allowed2 {
		t.Error("client-b should be allowed")
	}

	// client-a is now exhausted
	allowed3, _, _ := limiter.Allow("client-a")
	if allowed3 {
		t.Error("client-a second request should be denied")
	}
}

func TestSlidingWindowLimiter_WindowExpiry(t *testing.T) {
	limiter := NewSlidingWindowLimiter(1, time.Minute)

	now := time.Now()
	limiter.nowFunc = func() time.Time { return now }

	allowed, _, _ := limiter.Allow("client-a")
	if !allowed {
		t.Fatal("first request should be allowed")
	}

	// Still within window.
	limiter.nowFunc = func() time.Time { return now.Add(30 * time.Second) }
	allowed, _, _ = limiter.Allow("client-a")
	if allowed {
		t.Error("request within window should be denied")
	}

	// After window expires.
	limiter.nowFunc = func() time.Time { return now.Add(61 * time.Second) }
	allowed, remaining, _ := limiter.Allow("client-a")
	if !allowed {
		t.Error("request after window expiry should be allowed")
	}
	if remaining != 0 {
		t.Errorf("remaining = %d, want 0", remaining)
	}
}

func TestSlidingWindowLimiter_ResetAt(t *testing.T) {
	limiter := NewSlidingWindowLimiter(5, time.Minute)

	now := time.Now()
	limiter.nowFunc = func() time.Time { return now }

	_, _, resetAt := limiter.Allow("client-a")

	// The oldest request is the one we just made, so reset should be that time + window.
	expected := now.Add(time.Minute)
	if resetAt.Unix() != expected.Unix() {
		t.Errorf("resetAt = %v, want %v", resetAt.Unix(), expected.Unix())
	}
}

func TestSlidingWindowLimiter_Limit(t *testing.T) {
	limiter := NewSlidingWindowLimiter(42, time.Hour)
	if limiter.Limit() != 42 {
		t.Errorf("Limit() = %d, want 42", limiter.Limit())
	}
}

// ---------------------------------------------------------------------------
// clientKey tests
// ---------------------------------------------------------------------------

func TestClientKey_XForwardedFor(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient", nil)
	req.Header.Set("X-Forwarded-For", "203.0.113.50, 70.41.3.18")

	key := clientKey(req)
	if key != "203.0.113.50" {
		t.Errorf("clientKey = %q, want %q", key, "203.0.113.50")
	}
}

func TestClientKey_FallbackToRemoteAddr(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient", nil)
	req.RemoteAddr = "192.168.1.1:54321"

	key := clientKey(req)
	if key != "192.168.1.1" {
		t.Errorf("clientKey = %q, want %q", key, "192.168.1.1")
	}
}

func TestClientKey_RemoteAddrNoPort(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient", nil)
	req.RemoteAddr = "192.168.1.1"

	key := clientKey(req)
	if key != "192.168.1.1" {
		t.Errorf("clientKey = %q, want %q", key, "192.168.1.1")
	}
}

// ---------------------------------------------------------------------------
// RateLimitMiddleware integration tests
// ---------------------------------------------------------------------------

func TestRateLimitMiddleware_HeadersOnAllowedRequest(t *testing.T) {
	limiter := NewSlidingWindowLimiter(10, time.Minute)
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient", nil)
	req.RemoteAddr = "10.0.0.1:1234"
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := RateLimitMiddleware(limiter)(func(c echo.Context) error {
		return c.String(http.StatusOK, `{"resourceType":"Patient"}`)
	})

	if err := handler(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	// Verify X-RateLimit-Limit header.
	limitHeader := rec.Header().Get("X-RateLimit-Limit")
	if limitHeader != "10" {
		t.Errorf("X-RateLimit-Limit = %q, want %q", limitHeader, "10")
	}

	// Verify X-RateLimit-Remaining header.
	remainingHeader := rec.Header().Get("X-RateLimit-Remaining")
	if remainingHeader != "9" {
		t.Errorf("X-RateLimit-Remaining = %q, want %q", remainingHeader, "9")
	}

	// Verify X-RateLimit-Reset header is present and parseable.
	resetHeader := rec.Header().Get("X-RateLimit-Reset")
	if resetHeader == "" {
		t.Fatal("X-RateLimit-Reset header is missing")
	}
	resetUnix, err := strconv.ParseInt(resetHeader, 10, 64)
	if err != nil {
		t.Fatalf("X-RateLimit-Reset is not a valid integer: %v", err)
	}
	if resetUnix <= time.Now().Unix() {
		t.Errorf("X-RateLimit-Reset (%d) should be in the future", resetUnix)
	}

	// Verify Retry-After is NOT present on allowed requests.
	if ra := rec.Header().Get("Retry-After"); ra != "" {
		t.Errorf("Retry-After should not be set on allowed request, got %q", ra)
	}
}

func TestRateLimitMiddleware_429WhenExceeded(t *testing.T) {
	limiter := NewSlidingWindowLimiter(2, time.Minute)
	e := echo.New()

	okHandler := func(c echo.Context) error {
		return c.String(http.StatusOK, `{"resourceType":"Patient"}`)
	}
	mw := RateLimitMiddleware(limiter)

	// Exhaust the quota.
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodGet, "/fhir/Patient", nil)
		req.RemoteAddr = "10.0.0.2:1234"
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		if err := mw(okHandler)(c); err != nil {
			t.Fatal(err)
		}
		if rec.Code != http.StatusOK {
			t.Fatalf("request %d: expected 200, got %d", i+1, rec.Code)
		}
	}

	// Third request should be rejected.
	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient", nil)
	req.RemoteAddr = "10.0.0.2:1234"
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if err := mw(okHandler)(c); err != nil {
		t.Fatal(err)
	}

	if rec.Code != http.StatusTooManyRequests {
		t.Errorf("expected 429, got %d", rec.Code)
	}

	// Verify Retry-After header is set.
	ra := rec.Header().Get("Retry-After")
	if ra == "" {
		t.Fatal("Retry-After header missing on 429 response")
	}
	raSeconds, err := strconv.Atoi(ra)
	if err != nil {
		t.Fatalf("Retry-After is not a valid integer: %v", err)
	}
	if raSeconds <= 0 {
		t.Errorf("Retry-After = %d, want > 0", raSeconds)
	}

	// Verify X-RateLimit-Remaining is 0.
	if rem := rec.Header().Get("X-RateLimit-Remaining"); rem != "0" {
		t.Errorf("X-RateLimit-Remaining = %q, want %q", rem, "0")
	}

	// Verify body contains OperationOutcome.
	body := rec.Body.String()
	if !strings.Contains(body, "OperationOutcome") {
		t.Errorf("expected OperationOutcome in body, got: %s", body)
	}
	if !strings.Contains(body, "throttled") {
		t.Errorf("expected 'throttled' code in body, got: %s", body)
	}
}

func TestRateLimitMiddleware_DifferentClientsIndependent(t *testing.T) {
	limiter := NewSlidingWindowLimiter(1, time.Minute)
	e := echo.New()

	okHandler := func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	}
	mw := RateLimitMiddleware(limiter)

	// Client A uses its quota.
	reqA := httptest.NewRequest(http.MethodGet, "/fhir/Patient", nil)
	reqA.RemoteAddr = "10.0.0.1:1111"
	recA := httptest.NewRecorder()
	cA := e.NewContext(reqA, recA)
	if err := mw(okHandler)(cA); err != nil {
		t.Fatal(err)
	}
	if recA.Code != http.StatusOK {
		t.Errorf("client A: expected 200, got %d", recA.Code)
	}

	// Client B should still be allowed.
	reqB := httptest.NewRequest(http.MethodGet, "/fhir/Patient", nil)
	reqB.RemoteAddr = "10.0.0.2:2222"
	recB := httptest.NewRecorder()
	cB := e.NewContext(reqB, recB)
	if err := mw(okHandler)(cB); err != nil {
		t.Fatal(err)
	}
	if recB.Code != http.StatusOK {
		t.Errorf("client B: expected 200, got %d", recB.Code)
	}
}

func TestRateLimitMiddleware_UsesXForwardedFor(t *testing.T) {
	limiter := NewSlidingWindowLimiter(1, time.Minute)
	e := echo.New()

	okHandler := func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	}
	mw := RateLimitMiddleware(limiter)

	// First request with X-Forwarded-For.
	req1 := httptest.NewRequest(http.MethodGet, "/fhir/Patient", nil)
	req1.Header.Set("X-Forwarded-For", "203.0.113.50")
	req1.RemoteAddr = "10.0.0.99:9999"
	rec1 := httptest.NewRecorder()
	c1 := e.NewContext(req1, rec1)
	if err := mw(okHandler)(c1); err != nil {
		t.Fatal(err)
	}
	if rec1.Code != http.StatusOK {
		t.Fatalf("first request: expected 200, got %d", rec1.Code)
	}

	// Second request with same X-Forwarded-For should be denied.
	req2 := httptest.NewRequest(http.MethodGet, "/fhir/Patient", nil)
	req2.Header.Set("X-Forwarded-For", "203.0.113.50")
	req2.RemoteAddr = "10.0.0.100:8888"
	rec2 := httptest.NewRecorder()
	c2 := e.NewContext(req2, rec2)
	if err := mw(okHandler)(c2); err != nil {
		t.Fatal(err)
	}
	if rec2.Code != http.StatusTooManyRequests {
		t.Errorf("second request: expected 429, got %d", rec2.Code)
	}
}

func TestRateLimitMiddleware_HandlerNotCalledWhenDenied(t *testing.T) {
	limiter := NewSlidingWindowLimiter(1, time.Minute)
	e := echo.New()
	mw := RateLimitMiddleware(limiter)

	called := 0
	handler := func(c echo.Context) error {
		called++
		return c.String(http.StatusOK, "ok")
	}

	// First request — should call handler.
	req1 := httptest.NewRequest(http.MethodGet, "/fhir/Patient", nil)
	req1.RemoteAddr = "10.0.0.1:1111"
	rec1 := httptest.NewRecorder()
	c1 := e.NewContext(req1, rec1)
	if err := mw(handler)(c1); err != nil {
		t.Fatal(err)
	}

	// Second request — should NOT call handler.
	req2 := httptest.NewRequest(http.MethodGet, "/fhir/Patient", nil)
	req2.RemoteAddr = "10.0.0.1:1111"
	rec2 := httptest.NewRecorder()
	c2 := e.NewContext(req2, rec2)
	if err := mw(handler)(c2); err != nil {
		t.Fatal(err)
	}

	if called != 1 {
		t.Errorf("handler called %d times, want 1", called)
	}
}

func TestRateLimitMiddleware_RemainingDecrementsCorrectly(t *testing.T) {
	limiter := NewSlidingWindowLimiter(3, time.Minute)
	e := echo.New()
	mw := RateLimitMiddleware(limiter)

	okHandler := func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	}

	expected := []string{"2", "1", "0"}
	for i, want := range expected {
		req := httptest.NewRequest(http.MethodGet, "/fhir/Patient", nil)
		req.RemoteAddr = "10.0.0.5:5555"
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		if err := mw(okHandler)(c); err != nil {
			t.Fatal(err)
		}
		got := rec.Header().Get("X-RateLimit-Remaining")
		if got != want {
			t.Errorf("request %d: X-RateLimit-Remaining = %q, want %q", i+1, got, want)
		}
	}
}
