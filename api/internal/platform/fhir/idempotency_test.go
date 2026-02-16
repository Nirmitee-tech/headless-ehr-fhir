package fhir

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
)

// ---------------------------------------------------------------------------
// InMemoryIdempotencyStore unit tests
// ---------------------------------------------------------------------------

func TestInMemoryIdempotencyStore_SetAndGet(t *testing.T) {
	store := NewInMemoryIdempotencyStore(time.Hour)
	defer store.Stop()

	entry := &IdempotencyKey{
		Key:        "key-1",
		Method:     http.MethodPost,
		Path:       "/fhir/Patient",
		StatusCode: http.StatusCreated,
		Headers:    http.Header{"Content-Type": []string{"application/fhir+json"}},
		Body:       []byte(`{"resourceType":"Patient","id":"1"}`),
	}
	store.Set("key-1", entry)

	got, ok := store.Get("key-1")
	if !ok {
		t.Fatal("expected key-1 to be found")
	}
	if got.Key != "key-1" {
		t.Errorf("Key = %q, want %q", got.Key, "key-1")
	}
	if got.Method != http.MethodPost {
		t.Errorf("Method = %q, want %q", got.Method, http.MethodPost)
	}
	if got.Path != "/fhir/Patient" {
		t.Errorf("Path = %q, want %q", got.Path, "/fhir/Patient")
	}
	if got.StatusCode != http.StatusCreated {
		t.Errorf("StatusCode = %d, want %d", got.StatusCode, http.StatusCreated)
	}
	if string(got.Body) != `{"resourceType":"Patient","id":"1"}` {
		t.Errorf("Body = %q, want %q", string(got.Body), `{"resourceType":"Patient","id":"1"}`)
	}
	if got.Headers.Get("Content-Type") != "application/fhir+json" {
		t.Errorf("Content-Type header = %q, want %q", got.Headers.Get("Content-Type"), "application/fhir+json")
	}
}

func TestInMemoryIdempotencyStore_GetNotFound(t *testing.T) {
	store := NewInMemoryIdempotencyStore(time.Hour)
	defer store.Stop()

	_, ok := store.Get("nonexistent")
	if ok {
		t.Fatal("expected nonexistent key to return false")
	}
}

func TestInMemoryIdempotencyStore_Delete(t *testing.T) {
	store := NewInMemoryIdempotencyStore(time.Hour)
	defer store.Stop()

	entry := &IdempotencyKey{
		Key:        "key-del",
		Method:     http.MethodPost,
		Path:       "/fhir/Patient",
		StatusCode: http.StatusCreated,
		Body:       []byte("{}"),
	}
	store.Set("key-del", entry)

	store.Delete("key-del")

	_, ok := store.Get("key-del")
	if ok {
		t.Fatal("expected key to be deleted")
	}
}

func TestInMemoryIdempotencyStore_DeleteNonexistent(t *testing.T) {
	store := NewInMemoryIdempotencyStore(time.Hour)
	defer store.Stop()

	// Should not panic.
	store.Delete("does-not-exist")
}

func TestInMemoryIdempotencyStore_TTLExpiration(t *testing.T) {
	store := NewInMemoryIdempotencyStore(time.Minute)
	defer store.Stop()

	now := time.Now()
	store.nowFunc = func() time.Time { return now }

	entry := &IdempotencyKey{
		Key:        "key-ttl",
		Method:     http.MethodPost,
		Path:       "/fhir/Patient",
		StatusCode: http.StatusCreated,
		Body:       []byte("{}"),
	}
	store.Set("key-ttl", entry)

	// Within TTL.
	_, ok := store.Get("key-ttl")
	if !ok {
		t.Fatal("expected key to be found within TTL")
	}

	// After TTL.
	store.nowFunc = func() time.Time { return now.Add(2 * time.Minute) }
	_, ok = store.Get("key-ttl")
	if ok {
		t.Fatal("expected key to be expired after TTL")
	}
}

func TestInMemoryIdempotencyStore_DefaultTTL(t *testing.T) {
	store := NewInMemoryIdempotencyStore(0)
	defer store.Stop()

	if store.ttl != DefaultIdempotencyTTL {
		t.Errorf("ttl = %v, want %v", store.ttl, DefaultIdempotencyTTL)
	}
}

func TestInMemoryIdempotencyStore_NegativeTTL(t *testing.T) {
	store := NewInMemoryIdempotencyStore(-1 * time.Hour)
	defer store.Stop()

	if store.ttl != DefaultIdempotencyTTL {
		t.Errorf("ttl = %v, want %v", store.ttl, DefaultIdempotencyTTL)
	}
}

func TestInMemoryIdempotencyStore_EvictExpired(t *testing.T) {
	store := NewInMemoryIdempotencyStore(time.Minute)
	defer store.Stop()

	now := time.Now()
	store.nowFunc = func() time.Time { return now }

	store.Set("fresh", &IdempotencyKey{Key: "fresh", Body: []byte("{}")})
	store.Set("stale", &IdempotencyKey{Key: "stale", Body: []byte("{}")})

	// Advance time past the TTL for both, then set a new fresh entry.
	store.nowFunc = func() time.Time { return now.Add(2 * time.Minute) }
	store.Set("new-fresh", &IdempotencyKey{Key: "new-fresh", Body: []byte("{}")})

	store.evictExpired()

	store.mu.RLock()
	defer store.mu.RUnlock()
	if _, ok := store.entries["fresh"]; ok {
		t.Error("expected 'fresh' to be evicted")
	}
	if _, ok := store.entries["stale"]; ok {
		t.Error("expected 'stale' to be evicted")
	}
	if _, ok := store.entries["new-fresh"]; !ok {
		t.Error("expected 'new-fresh' to still exist")
	}
}

func TestInMemoryIdempotencyStore_GetReturnsCopy(t *testing.T) {
	store := NewInMemoryIdempotencyStore(time.Hour)
	defer store.Stop()

	store.Set("key-copy", &IdempotencyKey{
		Key:     "key-copy",
		Method:  http.MethodPost,
		Path:    "/fhir/Patient",
		Body:    []byte(`{"id":"1"}`),
		Headers: http.Header{"X-Custom": []string{"val"}},
	})

	got1, _ := store.Get("key-copy")
	got1.Body[0] = 'X' // Mutate the copy.
	got1.Headers.Set("X-Custom", "mutated")

	got2, _ := store.Get("key-copy")
	if string(got2.Body) != `{"id":"1"}` {
		t.Errorf("mutation leaked: Body = %q", string(got2.Body))
	}
	if got2.Headers.Get("X-Custom") != "val" {
		t.Errorf("mutation leaked: X-Custom = %q", got2.Headers.Get("X-Custom"))
	}
}

func TestInMemoryIdempotencyStore_ConcurrentAccess(t *testing.T) {
	store := NewInMemoryIdempotencyStore(time.Hour)
	defer store.Stop()

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			key := fmt.Sprintf("key-%d", n)
			store.Set(key, &IdempotencyKey{
				Key:    key,
				Method: http.MethodPost,
				Path:   "/fhir/Patient",
				Body:   []byte(fmt.Sprintf(`{"id":"%d"}`, n)),
			})
			store.Get(key)
			if n%3 == 0 {
				store.Delete(key)
			}
		}(i)
	}
	wg.Wait()
}

// ---------------------------------------------------------------------------
// IdempotencyMiddleware tests
// ---------------------------------------------------------------------------

func TestIdempotencyMiddleware_NoKey_Passthrough(t *testing.T) {
	store := NewInMemoryIdempotencyStore(time.Hour)
	defer store.Stop()

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/fhir/Patient", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := IdempotencyMiddleware(store)(func(c echo.Context) error {
		return c.String(http.StatusCreated, `{"resourceType":"Patient","id":"1"}`)
	})

	if err := handler(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
	if rec.Body.String() != `{"resourceType":"Patient","id":"1"}` {
		t.Errorf("unexpected body: %s", rec.Body.String())
	}
}

func TestIdempotencyMiddleware_EmptyKey_Passthrough(t *testing.T) {
	store := NewInMemoryIdempotencyStore(time.Hour)
	defer store.Stop()

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/fhir/Patient", nil)
	req.Header.Set("Idempotency-Key", "")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	called := false
	handler := IdempotencyMiddleware(store)(func(c echo.Context) error {
		called = true
		return c.String(http.StatusCreated, `{"id":"1"}`)
	})

	if err := handler(c); err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Error("expected handler to be called for empty key")
	}
}

func TestIdempotencyMiddleware_GET_Passthrough(t *testing.T) {
	store := NewInMemoryIdempotencyStore(time.Hour)
	defer store.Stop()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient/1", nil)
	req.Header.Set("Idempotency-Key", "get-key")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	called := false
	handler := IdempotencyMiddleware(store)(func(c echo.Context) error {
		called = true
		return c.String(http.StatusOK, `{"resourceType":"Patient","id":"1"}`)
	})

	if err := handler(c); err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Error("expected handler to be called for GET")
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestIdempotencyMiddleware_DELETE_Passthrough(t *testing.T) {
	store := NewInMemoryIdempotencyStore(time.Hour)
	defer store.Stop()

	e := echo.New()
	req := httptest.NewRequest(http.MethodDelete, "/fhir/Patient/1", nil)
	req.Header.Set("Idempotency-Key", "delete-key")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	called := false
	handler := IdempotencyMiddleware(store)(func(c echo.Context) error {
		called = true
		return c.NoContent(http.StatusNoContent)
	})

	if err := handler(c); err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Error("expected handler to be called for DELETE")
	}
}

func TestIdempotencyMiddleware_POST_FirstRequest_ExecuteAndCache(t *testing.T) {
	store := NewInMemoryIdempotencyStore(time.Hour)
	defer store.Stop()

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/fhir/Patient", nil)
	req.Header.Set("Idempotency-Key", "post-key-1")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	callCount := 0
	handler := IdempotencyMiddleware(store)(func(c echo.Context) error {
		callCount++
		c.Response().Header().Set("X-Custom", "test-value")
		return c.String(http.StatusCreated, `{"resourceType":"Patient","id":"new"}`)
	})

	if err := handler(c); err != nil {
		t.Fatal(err)
	}
	if callCount != 1 {
		t.Errorf("handler called %d times, want 1", callCount)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
	if rec.Body.String() != `{"resourceType":"Patient","id":"new"}` {
		t.Errorf("unexpected body: %s", rec.Body.String())
	}

	// Verify the response was cached.
	cached, ok := store.Get("post-key-1")
	if !ok {
		t.Fatal("expected response to be cached")
	}
	if cached.StatusCode != http.StatusCreated {
		t.Errorf("cached status = %d, want 201", cached.StatusCode)
	}
}

func TestIdempotencyMiddleware_POST_ReplayedResponse(t *testing.T) {
	store := NewInMemoryIdempotencyStore(time.Hour)
	defer store.Stop()

	e := echo.New()
	mw := IdempotencyMiddleware(store)

	callCount := 0
	innerHandler := func(c echo.Context) error {
		callCount++
		c.Response().Header().Set("ETag", `W/"1"`)
		return c.String(http.StatusCreated, `{"resourceType":"Patient","id":"1"}`)
	}

	// First request: execute and cache.
	req1 := httptest.NewRequest(http.MethodPost, "/fhir/Patient", nil)
	req1.Header.Set("Idempotency-Key", "replay-key")
	rec1 := httptest.NewRecorder()
	c1 := e.NewContext(req1, rec1)
	if err := mw(innerHandler)(c1); err != nil {
		t.Fatal(err)
	}

	// Second request: should replay.
	req2 := httptest.NewRequest(http.MethodPost, "/fhir/Patient", nil)
	req2.Header.Set("Idempotency-Key", "replay-key")
	rec2 := httptest.NewRecorder()
	c2 := e.NewContext(req2, rec2)
	if err := mw(innerHandler)(c2); err != nil {
		t.Fatal(err)
	}

	if callCount != 1 {
		t.Errorf("handler called %d times, want 1 (second request should be replayed)", callCount)
	}
	if rec2.Code != http.StatusCreated {
		t.Errorf("replayed status = %d, want 201", rec2.Code)
	}
	if rec2.Body.String() != `{"resourceType":"Patient","id":"1"}` {
		t.Errorf("replayed body = %q", rec2.Body.String())
	}
}

func TestIdempotencyMiddleware_ReplayedHeader(t *testing.T) {
	store := NewInMemoryIdempotencyStore(time.Hour)
	defer store.Stop()

	e := echo.New()
	mw := IdempotencyMiddleware(store)

	innerHandler := func(c echo.Context) error {
		return c.String(http.StatusCreated, `{"id":"1"}`)
	}

	// First request.
	req1 := httptest.NewRequest(http.MethodPost, "/fhir/Patient", nil)
	req1.Header.Set("Idempotency-Key", "replayed-header-key")
	rec1 := httptest.NewRecorder()
	c1 := e.NewContext(req1, rec1)
	if err := mw(innerHandler)(c1); err != nil {
		t.Fatal(err)
	}
	if rec1.Header().Get("X-Idempotency-Replayed") != "" {
		t.Error("first request should not have X-Idempotency-Replayed header")
	}

	// Second (replayed) request.
	req2 := httptest.NewRequest(http.MethodPost, "/fhir/Patient", nil)
	req2.Header.Set("Idempotency-Key", "replayed-header-key")
	rec2 := httptest.NewRecorder()
	c2 := e.NewContext(req2, rec2)
	if err := mw(innerHandler)(c2); err != nil {
		t.Fatal(err)
	}
	if rec2.Header().Get("X-Idempotency-Replayed") != "true" {
		t.Errorf("X-Idempotency-Replayed = %q, want %q", rec2.Header().Get("X-Idempotency-Replayed"), "true")
	}
}

func TestIdempotencyMiddleware_DifferentMethodSameKey_422(t *testing.T) {
	store := NewInMemoryIdempotencyStore(time.Hour)
	defer store.Stop()

	e := echo.New()
	mw := IdempotencyMiddleware(store)

	innerHandler := func(c echo.Context) error {
		return c.String(http.StatusOK, `{"id":"1"}`)
	}

	// First: POST /fhir/Patient
	req1 := httptest.NewRequest(http.MethodPost, "/fhir/Patient", nil)
	req1.Header.Set("Idempotency-Key", "conflict-key")
	rec1 := httptest.NewRecorder()
	c1 := e.NewContext(req1, rec1)
	if err := mw(innerHandler)(c1); err != nil {
		t.Fatal(err)
	}

	// Second: PUT /fhir/Patient (different method, same key)
	req2 := httptest.NewRequest(http.MethodPut, "/fhir/Patient", nil)
	req2.Header.Set("Idempotency-Key", "conflict-key")
	rec2 := httptest.NewRecorder()
	c2 := e.NewContext(req2, rec2)
	if err := mw(innerHandler)(c2); err != nil {
		t.Fatal(err)
	}

	if rec2.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected 422, got %d", rec2.Code)
	}
	body := rec2.Body.String()
	if !strings.Contains(body, "OperationOutcome") {
		t.Errorf("expected OperationOutcome in body, got: %s", body)
	}
	if !strings.Contains(body, "different operation") {
		t.Errorf("expected 'different operation' in diagnostics, got: %s", body)
	}
}

func TestIdempotencyMiddleware_DifferentPathSameKey_422(t *testing.T) {
	store := NewInMemoryIdempotencyStore(time.Hour)
	defer store.Stop()

	e := echo.New()
	mw := IdempotencyMiddleware(store)

	innerHandler := func(c echo.Context) error {
		return c.String(http.StatusCreated, `{}`)
	}

	// First: POST /fhir/Patient
	req1 := httptest.NewRequest(http.MethodPost, "/fhir/Patient", nil)
	req1.Header.Set("Idempotency-Key", "path-conflict-key")
	rec1 := httptest.NewRecorder()
	c1 := e.NewContext(req1, rec1)
	if err := mw(innerHandler)(c1); err != nil {
		t.Fatal(err)
	}

	// Second: POST /fhir/Observation (same method, different path)
	req2 := httptest.NewRequest(http.MethodPost, "/fhir/Observation", nil)
	req2.Header.Set("Idempotency-Key", "path-conflict-key")
	rec2 := httptest.NewRecorder()
	c2 := e.NewContext(req2, rec2)
	if err := mw(innerHandler)(c2); err != nil {
		t.Fatal(err)
	}

	if rec2.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected 422, got %d", rec2.Code)
	}
}

func TestIdempotencyMiddleware_PUT_Supported(t *testing.T) {
	store := NewInMemoryIdempotencyStore(time.Hour)
	defer store.Stop()

	e := echo.New()
	mw := IdempotencyMiddleware(store)

	callCount := 0
	innerHandler := func(c echo.Context) error {
		callCount++
		return c.String(http.StatusOK, `{"resourceType":"Patient","id":"1"}`)
	}

	// First PUT.
	req1 := httptest.NewRequest(http.MethodPut, "/fhir/Patient/1", nil)
	req1.Header.Set("Idempotency-Key", "put-key")
	rec1 := httptest.NewRecorder()
	c1 := e.NewContext(req1, rec1)
	if err := mw(innerHandler)(c1); err != nil {
		t.Fatal(err)
	}

	// Second PUT with same key: should replay.
	req2 := httptest.NewRequest(http.MethodPut, "/fhir/Patient/1", nil)
	req2.Header.Set("Idempotency-Key", "put-key")
	rec2 := httptest.NewRecorder()
	c2 := e.NewContext(req2, rec2)
	if err := mw(innerHandler)(c2); err != nil {
		t.Fatal(err)
	}

	if callCount != 1 {
		t.Errorf("handler called %d times, want 1", callCount)
	}
	if rec2.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec2.Code)
	}
}

func TestIdempotencyMiddleware_PATCH_Supported(t *testing.T) {
	store := NewInMemoryIdempotencyStore(time.Hour)
	defer store.Stop()

	e := echo.New()
	mw := IdempotencyMiddleware(store)

	callCount := 0
	innerHandler := func(c echo.Context) error {
		callCount++
		return c.String(http.StatusOK, `{"patched":true}`)
	}

	// First PATCH.
	req1 := httptest.NewRequest(http.MethodPatch, "/fhir/Patient/1", nil)
	req1.Header.Set("Idempotency-Key", "patch-key")
	rec1 := httptest.NewRecorder()
	c1 := e.NewContext(req1, rec1)
	if err := mw(innerHandler)(c1); err != nil {
		t.Fatal(err)
	}

	// Second PATCH with same key: should replay.
	req2 := httptest.NewRequest(http.MethodPatch, "/fhir/Patient/1", nil)
	req2.Header.Set("Idempotency-Key", "patch-key")
	rec2 := httptest.NewRecorder()
	c2 := e.NewContext(req2, rec2)
	if err := mw(innerHandler)(c2); err != nil {
		t.Fatal(err)
	}

	if callCount != 1 {
		t.Errorf("handler called %d times, want 1", callCount)
	}
	if rec2.Body.String() != `{"patched":true}` {
		t.Errorf("unexpected body: %s", rec2.Body.String())
	}
}

func TestIdempotencyMiddleware_LegacyHeader(t *testing.T) {
	store := NewInMemoryIdempotencyStore(time.Hour)
	defer store.Stop()

	e := echo.New()
	mw := IdempotencyMiddleware(store)

	callCount := 0
	innerHandler := func(c echo.Context) error {
		callCount++
		return c.String(http.StatusCreated, `{"id":"1"}`)
	}

	// First request with legacy X-Idempotency-Key header.
	req1 := httptest.NewRequest(http.MethodPost, "/fhir/Patient", nil)
	req1.Header.Set("X-Idempotency-Key", "legacy-key")
	rec1 := httptest.NewRecorder()
	c1 := e.NewContext(req1, rec1)
	if err := mw(innerHandler)(c1); err != nil {
		t.Fatal(err)
	}

	// Second request with same legacy header: should replay.
	req2 := httptest.NewRequest(http.MethodPost, "/fhir/Patient", nil)
	req2.Header.Set("X-Idempotency-Key", "legacy-key")
	rec2 := httptest.NewRecorder()
	c2 := e.NewContext(req2, rec2)
	if err := mw(innerHandler)(c2); err != nil {
		t.Fatal(err)
	}

	if callCount != 1 {
		t.Errorf("handler called %d times, want 1", callCount)
	}
	if rec2.Header().Get("X-Idempotency-Replayed") != "true" {
		t.Error("expected X-Idempotency-Replayed header on replayed response")
	}
}

func TestIdempotencyMiddleware_StandardHeaderTakesPrecedence(t *testing.T) {
	store := NewInMemoryIdempotencyStore(time.Hour)
	defer store.Stop()

	e := echo.New()
	mw := IdempotencyMiddleware(store)

	callCount := 0
	innerHandler := func(c echo.Context) error {
		callCount++
		return c.String(http.StatusCreated, `{"id":"1"}`)
	}

	// Send both headers; standard should take precedence.
	req1 := httptest.NewRequest(http.MethodPost, "/fhir/Patient", nil)
	req1.Header.Set("Idempotency-Key", "standard-key")
	req1.Header.Set("X-Idempotency-Key", "legacy-key")
	rec1 := httptest.NewRecorder()
	c1 := e.NewContext(req1, rec1)
	if err := mw(innerHandler)(c1); err != nil {
		t.Fatal(err)
	}

	// Retry with same standard key: should replay.
	req2 := httptest.NewRequest(http.MethodPost, "/fhir/Patient", nil)
	req2.Header.Set("Idempotency-Key", "standard-key")
	rec2 := httptest.NewRecorder()
	c2 := e.NewContext(req2, rec2)
	if err := mw(innerHandler)(c2); err != nil {
		t.Fatal(err)
	}

	if callCount != 1 {
		t.Errorf("handler called %d times, want 1", callCount)
	}

	// Using the legacy key should NOT find the cache (it was stored under standard-key).
	req3 := httptest.NewRequest(http.MethodPost, "/fhir/Patient", nil)
	req3.Header.Set("X-Idempotency-Key", "legacy-key")
	rec3 := httptest.NewRecorder()
	c3 := e.NewContext(req3, rec3)
	if err := mw(innerHandler)(c3); err != nil {
		t.Fatal(err)
	}

	if callCount != 2 {
		t.Errorf("handler called %d times, want 2 (legacy key should not match standard key)", callCount)
	}
}

func TestIdempotencyMiddleware_DifferentKeys_DifferentResponses(t *testing.T) {
	store := NewInMemoryIdempotencyStore(time.Hour)
	defer store.Stop()

	e := echo.New()
	mw := IdempotencyMiddleware(store)

	callCount := 0
	innerHandler := func(c echo.Context) error {
		callCount++
		return c.String(http.StatusCreated, fmt.Sprintf(`{"id":"%d"}`, callCount))
	}

	// First key.
	req1 := httptest.NewRequest(http.MethodPost, "/fhir/Patient", nil)
	req1.Header.Set("Idempotency-Key", "key-A")
	rec1 := httptest.NewRecorder()
	c1 := e.NewContext(req1, rec1)
	if err := mw(innerHandler)(c1); err != nil {
		t.Fatal(err)
	}

	// Second key.
	req2 := httptest.NewRequest(http.MethodPost, "/fhir/Patient", nil)
	req2.Header.Set("Idempotency-Key", "key-B")
	rec2 := httptest.NewRecorder()
	c2 := e.NewContext(req2, rec2)
	if err := mw(innerHandler)(c2); err != nil {
		t.Fatal(err)
	}

	if callCount != 2 {
		t.Errorf("handler called %d times, want 2", callCount)
	}
	if rec1.Body.String() == rec2.Body.String() {
		t.Error("different keys should produce different cached responses")
	}
}

func TestIdempotencyMiddleware_TTLExpiration_ReExecute(t *testing.T) {
	store := NewInMemoryIdempotencyStore(time.Minute)
	defer store.Stop()

	now := time.Now()
	store.nowFunc = func() time.Time { return now }

	e := echo.New()
	mw := IdempotencyMiddleware(store)

	callCount := 0
	innerHandler := func(c echo.Context) error {
		callCount++
		return c.String(http.StatusCreated, fmt.Sprintf(`{"call":%d}`, callCount))
	}

	// First request.
	req1 := httptest.NewRequest(http.MethodPost, "/fhir/Patient", nil)
	req1.Header.Set("Idempotency-Key", "ttl-key")
	rec1 := httptest.NewRecorder()
	c1 := e.NewContext(req1, rec1)
	if err := mw(innerHandler)(c1); err != nil {
		t.Fatal(err)
	}
	if callCount != 1 {
		t.Fatalf("handler called %d times, want 1", callCount)
	}

	// Advance past TTL.
	store.nowFunc = func() time.Time { return now.Add(2 * time.Minute) }

	// Second request: should re-execute because cache expired.
	req2 := httptest.NewRequest(http.MethodPost, "/fhir/Patient", nil)
	req2.Header.Set("Idempotency-Key", "ttl-key")
	rec2 := httptest.NewRecorder()
	c2 := e.NewContext(req2, rec2)
	if err := mw(innerHandler)(c2); err != nil {
		t.Fatal(err)
	}

	if callCount != 2 {
		t.Errorf("handler called %d times, want 2 (cache should have expired)", callCount)
	}
}

func TestIdempotencyMiddleware_HeadersReplayedCorrectly(t *testing.T) {
	store := NewInMemoryIdempotencyStore(time.Hour)
	defer store.Stop()

	e := echo.New()
	mw := IdempotencyMiddleware(store)

	innerHandler := func(c echo.Context) error {
		c.Response().Header().Set("ETag", `W/"abc"`)
		c.Response().Header().Set("Last-Modified", "Thu, 01 Jan 2026 00:00:00 GMT")
		c.Response().Header().Set("Content-Location", "/fhir/Patient/1/_history/1")
		return c.String(http.StatusCreated, `{"id":"1"}`)
	}

	// First request.
	req1 := httptest.NewRequest(http.MethodPost, "/fhir/Patient", nil)
	req1.Header.Set("Idempotency-Key", "headers-key")
	rec1 := httptest.NewRecorder()
	c1 := e.NewContext(req1, rec1)
	if err := mw(innerHandler)(c1); err != nil {
		t.Fatal(err)
	}

	// Second request (replayed).
	req2 := httptest.NewRequest(http.MethodPost, "/fhir/Patient", nil)
	req2.Header.Set("Idempotency-Key", "headers-key")
	rec2 := httptest.NewRecorder()
	c2 := e.NewContext(req2, rec2)
	if err := mw(innerHandler)(c2); err != nil {
		t.Fatal(err)
	}

	if rec2.Header().Get("ETag") != `W/"abc"` {
		t.Errorf("ETag = %q, want %q", rec2.Header().Get("ETag"), `W/"abc"`)
	}
	if rec2.Header().Get("Last-Modified") != "Thu, 01 Jan 2026 00:00:00 GMT" {
		t.Errorf("Last-Modified = %q", rec2.Header().Get("Last-Modified"))
	}
	if rec2.Header().Get("Content-Location") != "/fhir/Patient/1/_history/1" {
		t.Errorf("Content-Location = %q", rec2.Header().Get("Content-Location"))
	}
}

func TestIdempotencyMiddleware_StatusCodeReplayedCorrectly(t *testing.T) {
	store := NewInMemoryIdempotencyStore(time.Hour)
	defer store.Stop()

	e := echo.New()
	mw := IdempotencyMiddleware(store)

	statusCodes := []int{http.StatusCreated, http.StatusOK, http.StatusAccepted}
	for _, code := range statusCodes {
		key := fmt.Sprintf("status-%d", code)
		expectedCode := code

		innerHandler := func(c echo.Context) error {
			return c.String(expectedCode, `{"status":"ok"}`)
		}

		// First request.
		req1 := httptest.NewRequest(http.MethodPost, "/fhir/Patient", nil)
		req1.Header.Set("Idempotency-Key", key)
		rec1 := httptest.NewRecorder()
		c1 := e.NewContext(req1, rec1)
		if err := mw(innerHandler)(c1); err != nil {
			t.Fatal(err)
		}

		// Replayed request.
		req2 := httptest.NewRequest(http.MethodPost, "/fhir/Patient", nil)
		req2.Header.Set("Idempotency-Key", key)
		rec2 := httptest.NewRecorder()
		c2 := e.NewContext(req2, rec2)
		if err := mw(innerHandler)(c2); err != nil {
			t.Fatal(err)
		}

		if rec2.Code != expectedCode {
			t.Errorf("key %s: replayed status = %d, want %d", key, rec2.Code, expectedCode)
		}
	}
}

func TestIdempotencyMiddleware_HandlerError_NotCached(t *testing.T) {
	store := NewInMemoryIdempotencyStore(time.Hour)
	defer store.Stop()

	e := echo.New()
	mw := IdempotencyMiddleware(store)

	callCount := 0
	innerHandler := func(c echo.Context) error {
		callCount++
		return echo.NewHTTPError(http.StatusInternalServerError, "server error")
	}

	// First request: handler returns error.
	req1 := httptest.NewRequest(http.MethodPost, "/fhir/Patient", nil)
	req1.Header.Set("Idempotency-Key", "error-key")
	rec1 := httptest.NewRecorder()
	c1 := e.NewContext(req1, rec1)
	err := mw(innerHandler)(c1)
	if err == nil {
		t.Fatal("expected error from handler")
	}

	// Second request: should re-execute because error was not cached.
	req2 := httptest.NewRequest(http.MethodPost, "/fhir/Patient", nil)
	req2.Header.Set("Idempotency-Key", "error-key")
	rec2 := httptest.NewRecorder()
	c2 := e.NewContext(req2, rec2)
	err = mw(innerHandler)(c2)
	if err == nil {
		t.Fatal("expected error from handler")
	}

	if callCount != 2 {
		t.Errorf("handler called %d times, want 2 (errors should not be cached)", callCount)
	}
}

func TestIdempotencyMiddleware_ConcurrentRequests_SameKey(t *testing.T) {
	store := NewInMemoryIdempotencyStore(time.Hour)
	defer store.Stop()

	e := echo.New()
	mw := IdempotencyMiddleware(store)

	var mu sync.Mutex
	callCount := 0
	innerHandler := func(c echo.Context) error {
		mu.Lock()
		callCount++
		mu.Unlock()
		return c.String(http.StatusCreated, `{"id":"1"}`)
	}

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			req := httptest.NewRequest(http.MethodPost, "/fhir/Patient", nil)
			req.Header.Set("Idempotency-Key", "concurrent-key")
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			_ = mw(innerHandler)(c)
		}()
	}
	wg.Wait()

	// Due to race conditions, more than one goroutine may execute the handler
	// before the first one caches the result. The critical thing is that the
	// store is not corrupted and does not panic.
	mu.Lock()
	defer mu.Unlock()
	if callCount < 1 {
		t.Error("expected at least one handler invocation")
	}
}

func TestIdempotencyMiddleware_MultipleReplaysSameResponse(t *testing.T) {
	store := NewInMemoryIdempotencyStore(time.Hour)
	defer store.Stop()

	e := echo.New()
	mw := IdempotencyMiddleware(store)

	callCount := 0
	innerHandler := func(c echo.Context) error {
		callCount++
		return c.String(http.StatusCreated, `{"resourceType":"Patient","id":"repeat"}`)
	}

	// First request.
	req1 := httptest.NewRequest(http.MethodPost, "/fhir/Patient", nil)
	req1.Header.Set("Idempotency-Key", "multi-replay")
	rec1 := httptest.NewRecorder()
	c1 := e.NewContext(req1, rec1)
	if err := mw(innerHandler)(c1); err != nil {
		t.Fatal(err)
	}

	// Replay 5 more times.
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodPost, "/fhir/Patient", nil)
		req.Header.Set("Idempotency-Key", "multi-replay")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		if err := mw(innerHandler)(c); err != nil {
			t.Fatalf("replay %d: %v", i+1, err)
		}
		if rec.Code != http.StatusCreated {
			t.Errorf("replay %d: status = %d, want 201", i+1, rec.Code)
		}
		if rec.Body.String() != `{"resourceType":"Patient","id":"repeat"}` {
			t.Errorf("replay %d: unexpected body: %s", i+1, rec.Body.String())
		}
		if rec.Header().Get("X-Idempotency-Replayed") != "true" {
			t.Errorf("replay %d: missing X-Idempotency-Replayed header", i+1)
		}
	}

	if callCount != 1 {
		t.Errorf("handler called %d times, want 1", callCount)
	}
}

func TestIdempotencyMiddleware_EmptyBody(t *testing.T) {
	store := NewInMemoryIdempotencyStore(time.Hour)
	defer store.Stop()

	e := echo.New()
	mw := IdempotencyMiddleware(store)

	innerHandler := func(c echo.Context) error {
		return c.NoContent(http.StatusNoContent)
	}

	// First request.
	req1 := httptest.NewRequest(http.MethodPost, "/fhir/Patient", nil)
	req1.Header.Set("Idempotency-Key", "empty-body-key")
	rec1 := httptest.NewRecorder()
	c1 := e.NewContext(req1, rec1)
	if err := mw(innerHandler)(c1); err != nil {
		t.Fatal(err)
	}

	// Replay.
	req2 := httptest.NewRequest(http.MethodPost, "/fhir/Patient", nil)
	req2.Header.Set("Idempotency-Key", "empty-body-key")
	rec2 := httptest.NewRecorder()
	c2 := e.NewContext(req2, rec2)
	if err := mw(innerHandler)(c2); err != nil {
		t.Fatal(err)
	}

	if rec2.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec2.Code)
	}
	if rec2.Body.Len() != 0 {
		t.Errorf("expected empty body, got: %s", rec2.Body.String())
	}
}

// ---------------------------------------------------------------------------
// idempotencyRecorder unit tests
// ---------------------------------------------------------------------------

func TestIdempotencyRecorder_WriteHeader(t *testing.T) {
	rec := &idempotencyRecorder{
		ResponseWriter: httptest.NewRecorder(),
		body:           &bytes.Buffer{},
		statusCode:     http.StatusOK,
		headers:        make(http.Header),
	}

	rec.WriteHeader(http.StatusCreated)
	if rec.statusCode != http.StatusCreated {
		t.Errorf("statusCode = %d, want 201", rec.statusCode)
	}
	if !rec.wroteHead {
		t.Error("wroteHead should be true after WriteHeader")
	}
}

func TestIdempotencyRecorder_Write(t *testing.T) {
	rec := &idempotencyRecorder{
		ResponseWriter: httptest.NewRecorder(),
		body:           &bytes.Buffer{},
		statusCode:     http.StatusOK,
		headers:        make(http.Header),
	}

	n, err := rec.Write([]byte("hello"))
	if err != nil {
		t.Fatal(err)
	}
	if n != 5 {
		t.Errorf("Write returned %d, want 5", n)
	}
	if rec.body.String() != "hello" {
		t.Errorf("body = %q, want %q", rec.body.String(), "hello")
	}
	if !rec.wroteHead {
		t.Error("wroteHead should be true after Write")
	}
	if rec.statusCode != http.StatusOK {
		t.Errorf("statusCode = %d, want 200 (implicit)", rec.statusCode)
	}
}

func TestIdempotencyRecorder_Header(t *testing.T) {
	rec := &idempotencyRecorder{
		ResponseWriter: httptest.NewRecorder(),
		body:           &bytes.Buffer{},
		statusCode:     http.StatusOK,
		headers:        make(http.Header),
	}

	rec.Header().Set("X-Test", "value")
	if rec.Header().Get("X-Test") != "value" {
		t.Errorf("header X-Test = %q, want %q", rec.Header().Get("X-Test"), "value")
	}
}

func TestIdempotencyRecorder_MultipleWrites(t *testing.T) {
	rec := &idempotencyRecorder{
		ResponseWriter: httptest.NewRecorder(),
		body:           &bytes.Buffer{},
		statusCode:     http.StatusOK,
		headers:        make(http.Header),
	}

	rec.WriteHeader(http.StatusCreated)
	rec.Write([]byte("part1"))
	rec.Write([]byte("part2"))

	if rec.body.String() != "part1part2" {
		t.Errorf("body = %q, want %q", rec.body.String(), "part1part2")
	}
	if rec.statusCode != http.StatusCreated {
		t.Errorf("statusCode = %d, want 201", rec.statusCode)
	}
}
