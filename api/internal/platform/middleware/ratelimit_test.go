package middleware

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestRateLimit_RequestsWithinLimit(t *testing.T) {
	cfg := RateLimitConfig{
		RequestsPerSecond: 10,
		BurstSize:         5,
	}

	e := echo.New()
	mw := RateLimit(cfg)
	handler := mw(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	// Send 5 requests (within burst size), all should pass
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := handler(c)
		if err != nil {
			t.Fatalf("request %d: expected no error, got %v", i+1, err)
		}
		if rec.Code != http.StatusOK {
			t.Fatalf("request %d: expected 200, got %d", i+1, rec.Code)
		}

		// Verify X-RateLimit-Limit header is set
		limitHeader := rec.Header().Get("X-RateLimit-Limit")
		if limitHeader != "10" {
			t.Errorf("request %d: expected X-RateLimit-Limit '10', got %q", i+1, limitHeader)
		}
	}
}

func TestRateLimit_ExceedsLimit(t *testing.T) {
	cfg := RateLimitConfig{
		RequestsPerSecond: 1,
		BurstSize:         2,
	}

	e := echo.New()
	mw := RateLimit(cfg)
	handler := mw(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	// First 2 requests should pass (burst size = 2)
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		err := handler(c)
		if err != nil {
			t.Fatalf("request %d: expected no error, got %v", i+1, err)
		}
	}

	// Third request should be rate limited
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := handler(c)

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
}

func TestRateLimit_RetryAfterHeader(t *testing.T) {
	cfg := RateLimitConfig{
		RequestsPerSecond: 1,
		BurstSize:         1,
	}

	e := echo.New()
	mw := RateLimit(cfg)
	handler := mw(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	// First request passes
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	_ = handler(c)

	// Second request should be rate limited and include Retry-After
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	err := handler(c)

	if err == nil {
		t.Fatal("expected error for rate-limited request")
	}

	retryAfter := rec.Header().Get("Retry-After")
	if retryAfter == "" {
		t.Error("expected Retry-After header to be set")
	}

	retryVal, parseErr := strconv.Atoi(retryAfter)
	if parseErr != nil {
		t.Fatalf("Retry-After header is not a valid integer: %q", retryAfter)
	}
	if retryVal < 1 {
		t.Errorf("expected Retry-After >= 1, got %d", retryVal)
	}

	// Check X-RateLimit-Remaining is "0" for rate-limited requests
	remaining := rec.Header().Get("X-RateLimit-Remaining")
	if remaining != "0" {
		t.Errorf("expected X-RateLimit-Remaining '0', got %q", remaining)
	}
}

func TestRateLimit_PerKeyIsolation(t *testing.T) {
	cfg := RateLimitConfig{
		RequestsPerSecond: 1,
		BurstSize:         1,
	}

	e := echo.New()
	mw := RateLimit(cfg)
	handler := mw(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	// First request from "tenant-a" - should pass
	req1 := httptest.NewRequest(http.MethodGet, "/", nil)
	rec1 := httptest.NewRecorder()
	c1 := e.NewContext(req1, rec1)
	c1.Set("jwt_tenant_id", "tenant-a")
	err := handler(c1)
	if err != nil {
		t.Fatalf("tenant-a first request: expected no error, got %v", err)
	}

	// Second request from "tenant-a" - should be rate limited
	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	rec2 := httptest.NewRecorder()
	c2 := e.NewContext(req2, rec2)
	c2.Set("jwt_tenant_id", "tenant-a")
	err = handler(c2)
	if err == nil {
		t.Fatal("tenant-a second request: expected rate limit error")
	}

	// First request from "tenant-b" - should pass (separate bucket)
	req3 := httptest.NewRequest(http.MethodGet, "/", nil)
	rec3 := httptest.NewRecorder()
	c3 := e.NewContext(req3, rec3)
	c3.Set("jwt_tenant_id", "tenant-b")
	err = handler(c3)
	if err != nil {
		t.Fatalf("tenant-b first request: expected no error, got %v", err)
	}
}

func TestRateLimit_DefaultConfig(t *testing.T) {
	cfg := DefaultRateLimitConfig()
	if cfg.RequestsPerSecond != 100 {
		t.Errorf("expected RequestsPerSecond 100, got %f", cfg.RequestsPerSecond)
	}
	if cfg.BurstSize != 200 {
		t.Errorf("expected BurstSize 200, got %d", cfg.BurstSize)
	}
}

func TestTokenBucket_RetryAfterWithZeroRate(t *testing.T) {
	b := newTokenBucket(0, 1)
	// Exhaust the single token
	b.allow()
	// With zero refill rate, retryAfter should return 1
	ra := b.retryAfter()
	if ra != 1 {
		t.Errorf("expected retryAfter 1 for zero rate, got %d", ra)
	}
}

func TestRateLimiterStore_DoubleCheck(t *testing.T) {
	cfg := RateLimitConfig{
		RequestsPerSecond: 10,
		BurstSize:         5,
	}
	store := newRateLimiterStore(cfg)

	// Get a bucket - creates it
	b1 := store.getBucket("key1")
	if b1 == nil {
		t.Fatal("expected non-nil bucket")
	}

	// Get the same bucket again - returns existing
	b2 := store.getBucket("key1")
	if b1 != b2 {
		t.Error("expected same bucket instance for same key")
	}

	// Different key gets different bucket
	b3 := store.getBucket("key2")
	if b1 == b3 {
		t.Error("expected different bucket for different key")
	}
}
