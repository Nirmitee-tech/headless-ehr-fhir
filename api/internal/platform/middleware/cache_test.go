package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
)

// ---------------------------------------------------------------------------
// ETag tests
// ---------------------------------------------------------------------------

func TestETagMiddleware_SetsETagHeader(t *testing.T) {
	e := echo.New()
	cfg := CacheConfig{
		MaxAge:      300,
		Private:     true,
		ETagEnabled: true,
		VaryHeaders: []string{"Accept", "Authorization"},
	}
	handler := ETagMiddleware(cfg)(func(c echo.Context) error {
		return c.String(http.StatusOK, "hello world")
	})

	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := handler(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	etag := rec.Header().Get("ETag")
	if etag == "" {
		t.Fatal("expected ETag header to be set")
	}
	// Weak validator format: W/"..."
	if len(etag) < 4 || etag[:3] != `W/"` || etag[len(etag)-1] != '"' {
		t.Errorf("expected weak ETag format W/\"...\", got %q", etag)
	}
}

func TestETagMiddleware_304OnMatch(t *testing.T) {
	e := echo.New()
	cfg := CacheConfig{
		MaxAge:             300,
		Private:            true,
		ETagEnabled:        true,
		ConditionalEnabled: true,
		VaryHeaders:        []string{"Accept"},
	}
	body := "hello world"

	// First request to get the ETag.
	handler := ETagMiddleware(cfg)(func(c echo.Context) error {
		return c.String(http.StatusOK, body)
	})
	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	_ = handler(c)
	etag := rec.Header().Get("ETag")
	if etag == "" {
		t.Fatal("expected ETag from first request")
	}

	// Second request with If-None-Match.
	req2 := httptest.NewRequest(http.MethodGet, "/fhir/Patient", nil)
	req2.Header.Set("If-None-Match", etag)
	rec2 := httptest.NewRecorder()
	c2 := e.NewContext(req2, rec2)
	err := handler(c2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec2.Code != http.StatusNotModified {
		t.Errorf("expected 304, got %d", rec2.Code)
	}
	if rec2.Body.Len() != 0 {
		t.Errorf("expected empty body for 304, got %d bytes", rec2.Body.Len())
	}
}

func TestETagMiddleware_200OnMismatch(t *testing.T) {
	e := echo.New()
	cfg := CacheConfig{
		MaxAge:             300,
		Private:            true,
		ETagEnabled:        true,
		ConditionalEnabled: true,
		VaryHeaders:        []string{"Accept"},
	}
	handler := ETagMiddleware(cfg)(func(c echo.Context) error {
		return c.String(http.StatusOK, "hello world")
	})

	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient", nil)
	req.Header.Set("If-None-Match", `W/"does-not-match"`)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := handler(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestETagMiddleware_SkipsPOST(t *testing.T) {
	e := echo.New()
	cfg := CacheConfig{
		MaxAge:      300,
		Private:     true,
		ETagEnabled: true,
		VaryHeaders: []string{"Accept"},
	}
	handler := ETagMiddleware(cfg)(func(c echo.Context) error {
		return c.String(http.StatusOK, "created")
	})

	req := httptest.NewRequest(http.MethodPost, "/fhir/Patient", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := handler(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Header().Get("ETag") != "" {
		t.Error("expected no ETag on POST request")
	}
}

func TestETagMiddleware_SkipsErrorResponses(t *testing.T) {
	e := echo.New()
	cfg := CacheConfig{
		MaxAge:      300,
		Private:     true,
		ETagEnabled: true,
		VaryHeaders: []string{"Accept"},
	}
	handler := ETagMiddleware(cfg)(func(c echo.Context) error {
		return c.String(http.StatusNotFound, "not found")
	})

	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient/123", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := handler(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Header().Get("ETag") != "" {
		t.Error("expected no ETag for 404 response")
	}
}

func TestETagMiddleware_SetsCacheControl(t *testing.T) {
	e := echo.New()
	cfg := CacheConfig{
		MaxAge:      600,
		Private:     false,
		ETagEnabled: true,
		VaryHeaders: []string{"Accept"},
	}
	handler := ETagMiddleware(cfg)(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	_ = handler(c)

	cc := rec.Header().Get("Cache-Control")
	if cc == "" {
		t.Fatal("expected Cache-Control header")
	}
	// Should contain public and max-age=600
	if !containsSubstring(cc, "public") {
		t.Errorf("expected 'public' in Cache-Control, got %q", cc)
	}
	if !containsSubstring(cc, "max-age=600") {
		t.Errorf("expected 'max-age=600' in Cache-Control, got %q", cc)
	}
}

func TestETagMiddleware_PrivateCacheControl(t *testing.T) {
	e := echo.New()
	cfg := CacheConfig{
		MaxAge:      300,
		Private:     true,
		ETagEnabled: true,
		VaryHeaders: []string{"Accept"},
	}
	handler := ETagMiddleware(cfg)(func(c echo.Context) error {
		return c.String(http.StatusOK, "phi data")
	})

	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	_ = handler(c)

	cc := rec.Header().Get("Cache-Control")
	if !containsSubstring(cc, "private") {
		t.Errorf("expected 'private' in Cache-Control for PHI, got %q", cc)
	}
}

func TestETagMiddleware_NoStoreCacheControl(t *testing.T) {
	e := echo.New()
	cfg := CacheConfig{
		MaxAge:      300,
		NoStore:     true,
		ETagEnabled: true,
		VaryHeaders: []string{"Accept"},
	}
	handler := ETagMiddleware(cfg)(func(c echo.Context) error {
		return c.String(http.StatusOK, "sensitive")
	})

	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	_ = handler(c)

	cc := rec.Header().Get("Cache-Control")
	if !containsSubstring(cc, "no-store") {
		t.Errorf("expected 'no-store' in Cache-Control, got %q", cc)
	}
}

func TestETagMiddleware_SetsVaryHeader(t *testing.T) {
	e := echo.New()
	cfg := CacheConfig{
		MaxAge:      300,
		Private:     true,
		ETagEnabled: true,
		VaryHeaders: []string{"Accept", "Authorization", "Accept-Encoding"},
	}
	handler := ETagMiddleware(cfg)(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	_ = handler(c)

	vary := rec.Header().Get("Vary")
	if vary == "" {
		t.Fatal("expected Vary header")
	}
	for _, h := range []string{"Accept", "Authorization", "Accept-Encoding"} {
		if !containsSubstring(vary, h) {
			t.Errorf("expected %q in Vary header, got %q", h, vary)
		}
	}
}

func TestETagMiddleware_SkipsExcludedPaths(t *testing.T) {
	e := echo.New()
	cfg := CacheConfig{
		MaxAge:       300,
		Private:      true,
		ETagEnabled:  true,
		VaryHeaders:  []string{"Accept"},
		ExcludePaths: []string{"/fhir/$export", "/health"},
	}
	handler := ETagMiddleware(cfg)(func(c echo.Context) error {
		return c.String(http.StatusOK, "exporting")
	})

	req := httptest.NewRequest(http.MethodGet, "/fhir/$export", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := handler(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Header().Get("ETag") != "" {
		t.Error("expected no ETag for excluded path")
	}
	if rec.Header().Get("Cache-Control") != "" {
		t.Error("expected no Cache-Control for excluded path")
	}
}

// ---------------------------------------------------------------------------
// Conditional request tests
// ---------------------------------------------------------------------------

func TestConditionalRequest_IfModifiedSince(t *testing.T) {
	e := echo.New()
	handler := ConditionalRequestMiddleware()(func(c echo.Context) error {
		// Simulate a Last-Modified header set by the handler.
		c.Response().Header().Set("Last-Modified", time.Now().Add(-1*time.Hour).UTC().Format(http.TimeFormat))
		return c.String(http.StatusOK, "data")
	})

	// Request with If-Modified-Since in the future (meaning: "I already have a recent version").
	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient", nil)
	req.Header.Set("If-Modified-Since", time.Now().Add(1*time.Hour).UTC().Format(http.TimeFormat))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := handler(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNotModified {
		t.Errorf("expected 304, got %d", rec.Code)
	}
}

func TestConditionalRequest_IfMatch_Precondition(t *testing.T) {
	e := echo.New()
	handler := ConditionalRequestMiddleware()(func(c echo.Context) error {
		c.Response().Header().Set("ETag", `W/"abc123"`)
		return c.String(http.StatusOK, "data")
	})

	req := httptest.NewRequest(http.MethodPut, "/fhir/Patient/1", nil)
	req.Header.Set("If-Match", `W/"wrong-etag"`)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := handler(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusPreconditionFailed {
		t.Errorf("expected 412, got %d", rec.Code)
	}
}

// ---------------------------------------------------------------------------
// CacheStore tests
// ---------------------------------------------------------------------------

func TestInMemoryCacheStore_SetAndGet(t *testing.T) {
	store := NewInMemoryCacheStore()
	store.Set("key1", []byte("value1"), 5*time.Minute)

	data, ok := store.Get("key1")
	if !ok {
		t.Fatal("expected cache hit")
	}
	if string(data) != "value1" {
		t.Errorf("expected 'value1', got %q", string(data))
	}
}

func TestInMemoryCacheStore_Expiration(t *testing.T) {
	store := NewInMemoryCacheStore()
	store.Set("key1", []byte("value1"), 1*time.Millisecond)

	// Wait for expiration.
	time.Sleep(10 * time.Millisecond)

	_, ok := store.Get("key1")
	if ok {
		t.Error("expected cache miss for expired entry")
	}
}

func TestInMemoryCacheStore_Delete(t *testing.T) {
	store := NewInMemoryCacheStore()
	store.Set("key1", []byte("value1"), 5*time.Minute)
	store.Delete("key1")

	_, ok := store.Get("key1")
	if ok {
		t.Error("expected cache miss after delete")
	}
}

func TestInMemoryCacheStore_Clear(t *testing.T) {
	store := NewInMemoryCacheStore()
	store.Set("key1", []byte("value1"), 5*time.Minute)
	store.Set("key2", []byte("value2"), 5*time.Minute)
	store.Clear()

	_, ok1 := store.Get("key1")
	_, ok2 := store.Get("key2")
	if ok1 || ok2 {
		t.Error("expected cache to be empty after clear")
	}
}

func TestInMemoryCacheStore_ConcurrentAccess(t *testing.T) {
	store := NewInMemoryCacheStore()
	var wg sync.WaitGroup
	iterations := 100

	// Concurrent writes.
	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			key := "key"
			store.Set(key, []byte("value"), 1*time.Minute)
		}(i)
	}

	// Concurrent reads.
	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			store.Get("key")
		}()
	}

	// Concurrent deletes.
	for i := 0; i < iterations/2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			store.Delete("key")
		}()
	}

	wg.Wait()
}

// ---------------------------------------------------------------------------
// Response cache tests
// ---------------------------------------------------------------------------

func TestResponseCache_CacheMiss(t *testing.T) {
	e := echo.New()
	store := NewInMemoryCacheStore()
	handler := ResponseCacheMiddleware(store, 5*time.Minute)(func(c echo.Context) error {
		return c.String(http.StatusOK, "fresh data")
	})

	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient", nil)
	req.Header.Set("Accept", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := handler(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Header().Get("X-Cache") != "MISS" {
		t.Errorf("expected X-Cache: MISS, got %q", rec.Header().Get("X-Cache"))
	}
}

func TestResponseCache_CacheHit(t *testing.T) {
	e := echo.New()
	store := NewInMemoryCacheStore()
	callCount := 0
	handler := ResponseCacheMiddleware(store, 5*time.Minute)(func(c echo.Context) error {
		callCount++
		return c.String(http.StatusOK, "fresh data")
	})

	// First request: MISS
	req1 := httptest.NewRequest(http.MethodGet, "/fhir/Patient", nil)
	req1.Header.Set("Accept", "application/json")
	rec1 := httptest.NewRecorder()
	c1 := e.NewContext(req1, rec1)
	_ = handler(c1)

	// Second request: HIT
	req2 := httptest.NewRequest(http.MethodGet, "/fhir/Patient", nil)
	req2.Header.Set("Accept", "application/json")
	rec2 := httptest.NewRecorder()
	c2 := e.NewContext(req2, rec2)
	err := handler(c2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec2.Header().Get("X-Cache") != "HIT" {
		t.Errorf("expected X-Cache: HIT, got %q", rec2.Header().Get("X-Cache"))
	}
	if callCount != 1 {
		t.Errorf("expected handler called once, called %d times", callCount)
	}
}

func TestResponseCache_SkipsAuthorized(t *testing.T) {
	e := echo.New()
	store := NewInMemoryCacheStore()
	handler := ResponseCacheMiddleware(store, 5*time.Minute)(func(c echo.Context) error {
		return c.String(http.StatusOK, "private data")
	})

	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient", nil)
	req.Header.Set("Authorization", "Bearer token123")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := handler(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should not cache - no X-Cache header or MISS without storing
	xcache := rec.Header().Get("X-Cache")
	if xcache == "HIT" {
		t.Error("expected authorized request to not be served from cache")
	}
}

func TestResponseCache_XCacheHeader(t *testing.T) {
	e := echo.New()
	store := NewInMemoryCacheStore()
	handler := ResponseCacheMiddleware(store, 5*time.Minute)(func(c echo.Context) error {
		return c.String(http.StatusOK, "data")
	})

	// MISS
	req1 := httptest.NewRequest(http.MethodGet, "/fhir/Patient", nil)
	rec1 := httptest.NewRecorder()
	c1 := e.NewContext(req1, rec1)
	_ = handler(c1)
	if rec1.Header().Get("X-Cache") != "MISS" {
		t.Errorf("first request: expected MISS, got %q", rec1.Header().Get("X-Cache"))
	}

	// HIT
	req2 := httptest.NewRequest(http.MethodGet, "/fhir/Patient", nil)
	rec2 := httptest.NewRecorder()
	c2 := e.NewContext(req2, rec2)
	_ = handler(c2)
	if rec2.Header().Get("X-Cache") != "HIT" {
		t.Errorf("second request: expected HIT, got %q", rec2.Header().Get("X-Cache"))
	}
}

func TestResponseCache_Expiration(t *testing.T) {
	e := echo.New()
	store := NewInMemoryCacheStore()
	callCount := 0
	handler := ResponseCacheMiddleware(store, 1*time.Millisecond)(func(c echo.Context) error {
		callCount++
		return c.String(http.StatusOK, "data")
	})

	// First request
	req1 := httptest.NewRequest(http.MethodGet, "/fhir/Patient", nil)
	rec1 := httptest.NewRecorder()
	c1 := e.NewContext(req1, rec1)
	_ = handler(c1)

	// Wait for expiration
	time.Sleep(10 * time.Millisecond)

	// Second request should be a miss
	req2 := httptest.NewRequest(http.MethodGet, "/fhir/Patient", nil)
	rec2 := httptest.NewRecorder()
	c2 := e.NewContext(req2, rec2)
	_ = handler(c2)

	if rec2.Header().Get("X-Cache") != "MISS" {
		t.Errorf("expected MISS after expiry, got %q", rec2.Header().Get("X-Cache"))
	}
	if callCount != 2 {
		t.Errorf("expected handler called twice, called %d times", callCount)
	}
}

// ---------------------------------------------------------------------------
// Cleanup goroutine test
// ---------------------------------------------------------------------------

func TestInMemoryCacheStore_StartCleanup(t *testing.T) {
	store := NewInMemoryCacheStore()
	store.Set("key1", []byte("value1"), 1*time.Millisecond)

	ctx, cancel := context.WithCancel(context.Background())
	store.StartCleanup(ctx, 5*time.Millisecond)

	// Wait for cleanup to run at least once.
	time.Sleep(50 * time.Millisecond)
	cancel()

	_, ok := store.Get("key1")
	if ok {
		t.Error("expected expired entry to be cleaned up")
	}
}

// ---------------------------------------------------------------------------
// Helper function tests
// ---------------------------------------------------------------------------

func TestComputeETag(t *testing.T) {
	etag := computeETag([]byte("hello world"))
	if etag == "" {
		t.Fatal("expected non-empty ETag")
	}
	if etag[:3] != `W/"` {
		t.Errorf("expected weak validator prefix, got %q", etag)
	}
	// Same input should produce same ETag.
	etag2 := computeETag([]byte("hello world"))
	if etag != etag2 {
		t.Errorf("expected deterministic ETag: %q != %q", etag, etag2)
	}
	// Different input should produce different ETag.
	etag3 := computeETag([]byte("different"))
	if etag == etag3 {
		t.Error("expected different ETag for different input")
	}
}

func TestCacheKey(t *testing.T) {
	key := cacheKey("GET", "/fhir/Patient", "application/json")
	if key == "" {
		t.Fatal("expected non-empty cache key")
	}
	// Same inputs same key.
	key2 := cacheKey("GET", "/fhir/Patient", "application/json")
	if key != key2 {
		t.Error("expected same cache key for same inputs")
	}
	// Different inputs different key.
	key3 := cacheKey("GET", "/fhir/Patient", "application/xml")
	if key == key3 {
		t.Error("expected different cache key for different Accept")
	}
}

func TestShouldSkip(t *testing.T) {
	excludes := []string{"/fhir/$export", "/health"}
	if !shouldSkip("/fhir/$export", excludes) {
		t.Error("expected /fhir/$export to be skipped")
	}
	if !shouldSkip("/health", excludes) {
		t.Error("expected /health to be skipped")
	}
	if shouldSkip("/fhir/Patient", excludes) {
		t.Error("expected /fhir/Patient to not be skipped")
	}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsStr(s, substr))
}

func containsStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
