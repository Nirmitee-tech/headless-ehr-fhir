package fhir

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
)

// ---------------------------------------------------------------------------
// TestDefaultHeadMethodConfig
// ---------------------------------------------------------------------------

func TestDefaultHeadMethodConfig(t *testing.T) {
	cfg := DefaultHeadMethodConfig()
	if !cfg.EnableContentLength {
		t.Error("EnableContentLength should default to true")
	}
	if cfg.CacheHeaders {
		t.Error("CacheHeaders should default to false")
	}
	if cfg.CacheTTL != 0 {
		t.Errorf("CacheTTL = %v, want 0", cfg.CacheTTL)
	}
	if len(cfg.AllowedPaths) != 0 {
		t.Errorf("AllowedPaths = %v, want empty", cfg.AllowedPaths)
	}
	if !cfg.IncludeResourceHeaders {
		t.Error("IncludeResourceHeaders should default to true")
	}
}

// ---------------------------------------------------------------------------
// TestIsHeadRequest
// ---------------------------------------------------------------------------

func TestIsHeadRequest_HEAD(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodHead, "/fhir/Patient/1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if !IsHeadRequest(c) {
		t.Error("expected IsHeadRequest to return true for HEAD")
	}
}

func TestIsHeadRequest_GET(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient/1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if IsHeadRequest(c) {
		t.Error("expected IsHeadRequest to return false for GET")
	}
}

func TestIsHeadRequest_POST(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/fhir/Patient", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if IsHeadRequest(c) {
		t.Error("expected IsHeadRequest to return false for POST")
	}
}

// ---------------------------------------------------------------------------
// TestHeadResponseWriter
// ---------------------------------------------------------------------------

func TestHeadResponseWriter_CapturesBody(t *testing.T) {
	rec := httptest.NewRecorder()
	w := NewHeadResponseWriter(rec)

	body := []byte(`{"resourceType":"Patient","id":"1"}`)
	n, err := w.Write(body)
	if err != nil {
		t.Fatalf("Write error: %v", err)
	}
	if n != len(body) {
		t.Errorf("Write returned %d, want %d", n, len(body))
	}
	if w.body.String() != string(body) {
		t.Errorf("captured body = %q, want %q", w.body.String(), string(body))
	}
}

func TestHeadResponseWriter_PreservesStatusCode(t *testing.T) {
	rec := httptest.NewRecorder()
	w := NewHeadResponseWriter(rec)

	w.WriteHeader(http.StatusOK)
	if w.statusCode != http.StatusOK {
		t.Errorf("statusCode = %d, want %d", w.statusCode, http.StatusOK)
	}
}

func TestHeadResponseWriter_StatusCodeNotFound(t *testing.T) {
	rec := httptest.NewRecorder()
	w := NewHeadResponseWriter(rec)

	w.WriteHeader(http.StatusNotFound)
	if w.statusCode != http.StatusNotFound {
		t.Errorf("statusCode = %d, want %d", w.statusCode, http.StatusNotFound)
	}
}

func TestHeadResponseWriter_CapturesHeaders(t *testing.T) {
	rec := httptest.NewRecorder()
	w := NewHeadResponseWriter(rec)

	w.Header().Set("ETag", `W/"5"`)
	w.Header().Set("Content-Type", "application/fhir+json")

	if got := w.Header().Get("ETag"); got != `W/"5"` {
		t.Errorf("ETag = %q, want W/\"5\"", got)
	}
	if got := w.Header().Get("Content-Type"); got != "application/fhir+json" {
		t.Errorf("Content-Type = %q, want application/fhir+json", got)
	}
}

func TestHeadResponseWriter_DefaultStatusCode(t *testing.T) {
	rec := httptest.NewRecorder()
	w := NewHeadResponseWriter(rec)

	// Before WriteHeader is called, statusCode should be 0 (unset).
	if w.statusCode != 0 {
		t.Errorf("initial statusCode = %d, want 0", w.statusCode)
	}
}

func TestHeadResponseWriter_MultipleWrites(t *testing.T) {
	rec := httptest.NewRecorder()
	w := NewHeadResponseWriter(rec)

	w.Write([]byte(`{"resourceType"`))
	w.Write([]byte(`:"Patient"}`))

	expected := `{"resourceType":"Patient"}`
	if w.body.String() != expected {
		t.Errorf("body = %q, want %q", w.body.String(), expected)
	}
}

// ---------------------------------------------------------------------------
// TestExtractResourceMetadata
// ---------------------------------------------------------------------------

func TestExtractResourceMetadata_PatientResource(t *testing.T) {
	body := []byte(`{
		"resourceType": "Patient",
		"id": "123",
		"meta": {
			"versionId": "5",
			"lastUpdated": "2024-01-15T10:30:00Z"
		}
	}`)

	meta := ExtractResourceMetadata(body)

	if meta["resourceType"] != "Patient" {
		t.Errorf("resourceType = %q, want Patient", meta["resourceType"])
	}
	if meta["id"] != "123" {
		t.Errorf("id = %q, want 123", meta["id"])
	}
	if meta["versionId"] != "5" {
		t.Errorf("versionId = %q, want 5", meta["versionId"])
	}
	if meta["lastUpdated"] != "2024-01-15T10:30:00Z" {
		t.Errorf("lastUpdated = %q, want 2024-01-15T10:30:00Z", meta["lastUpdated"])
	}
}

func TestExtractResourceMetadata_NoMetadata(t *testing.T) {
	body := []byte(`{"resourceType": "Patient", "id": "123"}`)

	meta := ExtractResourceMetadata(body)

	if meta["resourceType"] != "Patient" {
		t.Errorf("resourceType = %q, want Patient", meta["resourceType"])
	}
	if meta["id"] != "123" {
		t.Errorf("id = %q, want 123", meta["id"])
	}
	if _, ok := meta["versionId"]; ok {
		t.Error("expected no versionId key")
	}
	if _, ok := meta["lastUpdated"]; ok {
		t.Error("expected no lastUpdated key")
	}
}

func TestExtractResourceMetadata_EmptyBody(t *testing.T) {
	meta := ExtractResourceMetadata([]byte{})
	if len(meta) != 0 {
		t.Errorf("expected empty metadata, got %v", meta)
	}
}

func TestExtractResourceMetadata_InvalidJSON(t *testing.T) {
	meta := ExtractResourceMetadata([]byte(`not valid json`))
	if len(meta) != 0 {
		t.Errorf("expected empty metadata for invalid JSON, got %v", meta)
	}
}

func TestExtractResourceMetadata_BundleResource(t *testing.T) {
	body := []byte(`{
		"resourceType": "Bundle",
		"id": "search-results",
		"type": "searchset",
		"total": 5
	}`)

	meta := ExtractResourceMetadata(body)

	if meta["resourceType"] != "Bundle" {
		t.Errorf("resourceType = %q, want Bundle", meta["resourceType"])
	}
	if meta["id"] != "search-results" {
		t.Errorf("id = %q, want search-results", meta["id"])
	}
}

// ---------------------------------------------------------------------------
// TestBuildHeadHeaders
// ---------------------------------------------------------------------------

func TestBuildHeadHeaders_FullResponse(t *testing.T) {
	resp := &HeadResponse{
		StatusCode:    http.StatusOK,
		ContentLength: 1234,
		ContentType:   "application/fhir+json; charset=utf-8",
		ETag:          `W/"5"`,
		LastModified:  "2024-01-15T10:30:00Z",
		Headers: map[string][]string{
			"X-Request-ID": {"req-123"},
		},
	}

	headers := BuildHeadHeaders(resp)

	if got := headers.Get("Content-Length"); got != "1234" {
		t.Errorf("Content-Length = %q, want 1234", got)
	}
	if got := headers.Get("Content-Type"); got != "application/fhir+json; charset=utf-8" {
		t.Errorf("Content-Type = %q, want application/fhir+json; charset=utf-8", got)
	}
	if got := headers.Get("ETag"); got != `W/"5"` {
		t.Errorf("ETag = %q, want W/\"5\"", got)
	}
	if got := headers.Get("Last-Modified"); got != "2024-01-15T10:30:00Z" {
		t.Errorf("Last-Modified = %q, want 2024-01-15T10:30:00Z", got)
	}
	if got := headers.Get("X-Request-ID"); got != "req-123" {
		t.Errorf("X-Request-ID = %q, want req-123", got)
	}
}

func TestBuildHeadHeaders_MinimalResponse(t *testing.T) {
	resp := &HeadResponse{
		StatusCode:  http.StatusOK,
		ContentType: "application/fhir+json",
	}

	headers := BuildHeadHeaders(resp)

	if got := headers.Get("Content-Type"); got != "application/fhir+json" {
		t.Errorf("Content-Type = %q, want application/fhir+json", got)
	}
	// Content-Length should be "0" when not set (zero value).
	if got := headers.Get("Content-Length"); got != "0" {
		t.Errorf("Content-Length = %q, want 0", got)
	}
	if got := headers.Get("ETag"); got != "" {
		t.Errorf("ETag = %q, want empty", got)
	}
	if got := headers.Get("Last-Modified"); got != "" {
		t.Errorf("Last-Modified = %q, want empty", got)
	}
}

func TestBuildHeadHeaders_ZeroContentLength(t *testing.T) {
	resp := &HeadResponse{
		StatusCode:    http.StatusOK,
		ContentLength: 0,
		ContentType:   "application/fhir+json",
	}

	headers := BuildHeadHeaders(resp)

	if got := headers.Get("Content-Length"); got != "0" {
		t.Errorf("Content-Length = %q, want 0", got)
	}
}

// ---------------------------------------------------------------------------
// TestValidateHeadResponse
// ---------------------------------------------------------------------------

func TestValidateHeadResponse_Valid(t *testing.T) {
	headers := http.Header{}
	headers.Set("Content-Type", "application/fhir+json")

	issues := ValidateHeadResponse(http.StatusOK, headers)

	if len(issues) != 0 {
		t.Errorf("expected no issues, got %d: %v", len(issues), issues)
	}
}

func TestValidateHeadResponse_MissingContentType(t *testing.T) {
	headers := http.Header{}

	issues := ValidateHeadResponse(http.StatusOK, headers)

	found := false
	for _, issue := range issues {
		if issue.Code == VIssueTypeRequired {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected a 'required' validation issue for missing Content-Type")
	}
}

func TestValidateHeadResponse_InvalidStatusCode(t *testing.T) {
	headers := http.Header{}
	headers.Set("Content-Type", "application/fhir+json")

	issues := ValidateHeadResponse(0, headers)

	found := false
	for _, issue := range issues {
		if issue.Code == VIssueTypeValue {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected a 'value' validation issue for invalid status code")
	}
}

func TestValidateHeadResponse_Non2xxStatus(t *testing.T) {
	headers := http.Header{}
	headers.Set("Content-Type", "application/fhir+json")

	// 404 is valid -- it should pass with just a warning or no issue.
	issues := ValidateHeadResponse(http.StatusNotFound, headers)

	// Error-level issues should not include the status code since 404 is valid HTTP.
	for _, issue := range issues {
		if issue.Severity == SeverityError && issue.Code == VIssueTypeValue {
			t.Errorf("unexpected error-level issue for 404: %v", issue)
		}
	}
}

func TestValidateHeadResponse_WithETagHeader(t *testing.T) {
	headers := http.Header{}
	headers.Set("Content-Type", "application/fhir+json")
	headers.Set("ETag", `W/"3"`)

	issues := ValidateHeadResponse(http.StatusOK, headers)
	if len(issues) != 0 {
		t.Errorf("expected no issues for valid response with ETag, got %d", len(issues))
	}
}

// ---------------------------------------------------------------------------
// TestHeadCache
// ---------------------------------------------------------------------------

func TestHeadCache_GetSet(t *testing.T) {
	cache := NewHeadCache(5 * time.Minute)

	resp := &HeadResponse{
		StatusCode:    http.StatusOK,
		ContentLength: 100,
		ContentType:   "application/fhir+json",
		ETag:          `W/"1"`,
	}

	cache.Set("Patient/1", resp)

	got, ok := cache.Get("Patient/1")
	if !ok {
		t.Fatal("expected cache hit")
	}
	if got.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", got.StatusCode, http.StatusOK)
	}
	if got.ContentLength != 100 {
		t.Errorf("ContentLength = %d, want 100", got.ContentLength)
	}
	if got.ETag != `W/"1"` {
		t.Errorf("ETag = %q, want W/\"1\"", got.ETag)
	}
}

func TestHeadCache_Miss(t *testing.T) {
	cache := NewHeadCache(5 * time.Minute)

	_, ok := cache.Get("Patient/nonexistent")
	if ok {
		t.Error("expected cache miss for nonexistent key")
	}
}

func TestHeadCache_Expired(t *testing.T) {
	cache := NewHeadCache(1 * time.Millisecond)

	resp := &HeadResponse{
		StatusCode: http.StatusOK,
	}

	cache.Set("Patient/1", resp)

	// Wait for TTL to expire.
	time.Sleep(5 * time.Millisecond)

	_, ok := cache.Get("Patient/1")
	if ok {
		t.Error("expected cache miss for expired entry")
	}
}

func TestHeadCache_Invalidate(t *testing.T) {
	cache := NewHeadCache(5 * time.Minute)

	resp := &HeadResponse{StatusCode: http.StatusOK}
	cache.Set("Patient/1", resp)

	cache.Invalidate("Patient/1")

	_, ok := cache.Get("Patient/1")
	if ok {
		t.Error("expected cache miss after invalidation")
	}
}

func TestHeadCache_Clear(t *testing.T) {
	cache := NewHeadCache(5 * time.Minute)

	cache.Set("Patient/1", &HeadResponse{StatusCode: http.StatusOK})
	cache.Set("Patient/2", &HeadResponse{StatusCode: http.StatusOK})
	cache.Set("Observation/1", &HeadResponse{StatusCode: http.StatusOK})

	cache.Clear()

	if _, ok := cache.Get("Patient/1"); ok {
		t.Error("expected cache miss for Patient/1 after clear")
	}
	if _, ok := cache.Get("Patient/2"); ok {
		t.Error("expected cache miss for Patient/2 after clear")
	}
	if _, ok := cache.Get("Observation/1"); ok {
		t.Error("expected cache miss for Observation/1 after clear")
	}
}

func TestHeadCache_ConcurrentAccess(t *testing.T) {
	cache := NewHeadCache(5 * time.Minute)

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			key := fmt.Sprintf("Patient/%d", i)
			cache.Set(key, &HeadResponse{
				StatusCode:    http.StatusOK,
				ContentLength: int64(i),
			})
			cache.Get(key)
		}(i)
	}
	wg.Wait()

	// Verify at least some entries are present.
	hits := 0
	for i := 0; i < 100; i++ {
		key := fmt.Sprintf("Patient/%d", i)
		if _, ok := cache.Get(key); ok {
			hits++
		}
	}
	if hits != 100 {
		t.Errorf("expected 100 cache hits after concurrent writes, got %d", hits)
	}
}

func TestHeadCache_OverwriteExisting(t *testing.T) {
	cache := NewHeadCache(5 * time.Minute)

	cache.Set("Patient/1", &HeadResponse{StatusCode: http.StatusOK, ETag: `W/"1"`})
	cache.Set("Patient/1", &HeadResponse{StatusCode: http.StatusOK, ETag: `W/"2"`})

	got, ok := cache.Get("Patient/1")
	if !ok {
		t.Fatal("expected cache hit")
	}
	if got.ETag != `W/"2"` {
		t.Errorf("ETag = %q, want W/\"2\"", got.ETag)
	}
}

// ---------------------------------------------------------------------------
// TestGenerateCacheKey
// ---------------------------------------------------------------------------

func TestGenerateCacheKey_SameRequest(t *testing.T) {
	headers := http.Header{}
	headers.Set("Accept", "application/fhir+json")

	key1 := GenerateCacheKey("HEAD", "/fhir/Patient/1", headers)
	key2 := GenerateCacheKey("HEAD", "/fhir/Patient/1", headers)

	if key1 != key2 {
		t.Errorf("same request produced different keys: %q vs %q", key1, key2)
	}
}

func TestGenerateCacheKey_DifferentPaths(t *testing.T) {
	headers := http.Header{}

	key1 := GenerateCacheKey("HEAD", "/fhir/Patient/1", headers)
	key2 := GenerateCacheKey("HEAD", "/fhir/Patient/2", headers)

	if key1 == key2 {
		t.Error("different paths should produce different cache keys")
	}
}

func TestGenerateCacheKey_DifferentHeaders(t *testing.T) {
	h1 := http.Header{}
	h1.Set("Accept", "application/fhir+json")

	h2 := http.Header{}
	h2.Set("Accept", "application/json")

	key1 := GenerateCacheKey("HEAD", "/fhir/Patient/1", h1)
	key2 := GenerateCacheKey("HEAD", "/fhir/Patient/1", h2)

	if key1 == key2 {
		t.Error("different Accept headers should produce different cache keys")
	}
}

func TestGenerateCacheKey_DifferentMethods(t *testing.T) {
	headers := http.Header{}

	key1 := GenerateCacheKey("HEAD", "/fhir/Patient/1", headers)
	key2 := GenerateCacheKey("GET", "/fhir/Patient/1", headers)

	if key1 == key2 {
		t.Error("different methods should produce different cache keys")
	}
}

func TestGenerateCacheKey_EmptyHeaders(t *testing.T) {
	key := GenerateCacheKey("HEAD", "/fhir/Patient/1", http.Header{})
	if key == "" {
		t.Error("cache key should not be empty")
	}
}

// ---------------------------------------------------------------------------
// TestHeadMethodMiddleware
// ---------------------------------------------------------------------------

func TestHeadMethodMiddleware_BasicHEAD(t *testing.T) {
	e := echo.New()
	cfg := DefaultHeadMethodConfig()
	req := httptest.NewRequest(http.MethodHead, "/fhir/Patient/1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := HeadMethodMiddleware(&cfg)(func(c echo.Context) error {
		c.Response().Header().Set("ETag", `W/"5"`)
		c.Response().Header().Set("Last-Modified", "2024-01-15T10:30:00Z")
		return c.String(http.StatusOK, `{"resourceType":"Patient","id":"1"}`)
	})

	if err := handler(c); err != nil {
		t.Fatalf("handler error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	// HEAD response MUST NOT have a body.
	if rec.Body.Len() != 0 {
		t.Errorf("expected empty body for HEAD, got %d bytes: %s", rec.Body.Len(), rec.Body.String())
	}
}

func TestHeadMethodMiddleware_PassesGETThrough(t *testing.T) {
	e := echo.New()
	cfg := DefaultHeadMethodConfig()
	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient/1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	body := `{"resourceType":"Patient","id":"1"}`
	handler := HeadMethodMiddleware(&cfg)(func(c echo.Context) error {
		return c.String(http.StatusOK, body)
	})

	if err := handler(c); err != nil {
		t.Fatalf("handler error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if rec.Body.String() != body {
		t.Errorf("body = %q, want %q", rec.Body.String(), body)
	}
}

func TestHeadMethodMiddleware_PassesPOSTThrough(t *testing.T) {
	e := echo.New()
	cfg := DefaultHeadMethodConfig()
	req := httptest.NewRequest(http.MethodPost, "/fhir/Patient", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	body := `{"resourceType":"Patient","id":"1"}`
	handler := HeadMethodMiddleware(&cfg)(func(c echo.Context) error {
		return c.String(http.StatusCreated, body)
	})

	if err := handler(c); err != nil {
		t.Fatalf("handler error: %v", err)
	}

	if rec.Code != http.StatusCreated {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusCreated)
	}
	if rec.Body.String() != body {
		t.Errorf("body = %q, want %q", rec.Body.String(), body)
	}
}

func TestHeadMethodMiddleware_PreservesHeaders(t *testing.T) {
	e := echo.New()
	cfg := DefaultHeadMethodConfig()
	req := httptest.NewRequest(http.MethodHead, "/fhir/Patient/1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := HeadMethodMiddleware(&cfg)(func(c echo.Context) error {
		c.Response().Header().Set("ETag", `W/"7"`)
		c.Response().Header().Set("Last-Modified", "2024-06-01T12:00:00Z")
		c.Response().Header().Set("Content-Type", "application/fhir+json; charset=utf-8")
		c.Response().Header().Set("X-Request-ID", "req-456")
		return c.String(http.StatusOK, `{"resourceType":"Patient","id":"1"}`)
	})

	if err := handler(c); err != nil {
		t.Fatalf("handler error: %v", err)
	}

	if got := rec.Header().Get("ETag"); got != `W/"7"` {
		t.Errorf("ETag = %q, want W/\"7\"", got)
	}
	if got := rec.Header().Get("Last-Modified"); got != "2024-06-01T12:00:00Z" {
		t.Errorf("Last-Modified = %q, want 2024-06-01T12:00:00Z", got)
	}
	if got := rec.Header().Get("X-Request-ID"); got != "req-456" {
		t.Errorf("X-Request-ID = %q, want req-456", got)
	}
}

func TestHeadMethodMiddleware_StripsBody(t *testing.T) {
	e := echo.New()
	cfg := DefaultHeadMethodConfig()
	req := httptest.NewRequest(http.MethodHead, "/fhir/Patient/1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	largeBody := `{"resourceType":"Patient","id":"1","name":[{"family":"Smith","given":["John"]}]}`
	handler := HeadMethodMiddleware(&cfg)(func(c echo.Context) error {
		return c.String(http.StatusOK, largeBody)
	})

	if err := handler(c); err != nil {
		t.Fatalf("handler error: %v", err)
	}

	if rec.Body.Len() != 0 {
		t.Errorf("expected empty body, got %d bytes", rec.Body.Len())
	}
}

func TestHeadMethodMiddleware_ContentLength(t *testing.T) {
	e := echo.New()
	cfg := DefaultHeadMethodConfig()
	cfg.EnableContentLength = true
	req := httptest.NewRequest(http.MethodHead, "/fhir/Patient/1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	body := `{"resourceType":"Patient","id":"1"}`
	handler := HeadMethodMiddleware(&cfg)(func(c echo.Context) error {
		return c.String(http.StatusOK, body)
	})

	if err := handler(c); err != nil {
		t.Fatalf("handler error: %v", err)
	}

	expected := fmt.Sprintf("%d", len(body))
	if got := rec.Header().Get("Content-Length"); got != expected {
		t.Errorf("Content-Length = %q, want %q", got, expected)
	}
}

func TestHeadMethodMiddleware_DisableContentLength(t *testing.T) {
	e := echo.New()
	cfg := DefaultHeadMethodConfig()
	cfg.EnableContentLength = false
	req := httptest.NewRequest(http.MethodHead, "/fhir/Patient/1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := HeadMethodMiddleware(&cfg)(func(c echo.Context) error {
		return c.String(http.StatusOK, `{"resourceType":"Patient","id":"1"}`)
	})

	if err := handler(c); err != nil {
		t.Fatalf("handler error: %v", err)
	}

	// When disabled, Content-Length should not be set by the middleware.
	if got := rec.Header().Get("Content-Length"); got != "" {
		t.Errorf("Content-Length = %q, want empty (disabled)", got)
	}
}

func TestHeadMethodMiddleware_ResourceHeaders(t *testing.T) {
	e := echo.New()
	cfg := DefaultHeadMethodConfig()
	cfg.IncludeResourceHeaders = true
	req := httptest.NewRequest(http.MethodHead, "/fhir/Patient/1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := HeadMethodMiddleware(&cfg)(func(c echo.Context) error {
		return c.String(http.StatusOK, `{"resourceType":"Patient","id":"123","meta":{"versionId":"5","lastUpdated":"2024-01-15T10:30:00Z"}}`)
	})

	if err := handler(c); err != nil {
		t.Fatalf("handler error: %v", err)
	}

	if got := rec.Header().Get("X-FHIR-ResourceType"); got != "Patient" {
		t.Errorf("X-FHIR-ResourceType = %q, want Patient", got)
	}
	if got := rec.Header().Get("X-FHIR-ResourceId"); got != "123" {
		t.Errorf("X-FHIR-ResourceId = %q, want 123", got)
	}
}

func TestHeadMethodMiddleware_DisableResourceHeaders(t *testing.T) {
	e := echo.New()
	cfg := DefaultHeadMethodConfig()
	cfg.IncludeResourceHeaders = false
	req := httptest.NewRequest(http.MethodHead, "/fhir/Patient/1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := HeadMethodMiddleware(&cfg)(func(c echo.Context) error {
		return c.String(http.StatusOK, `{"resourceType":"Patient","id":"123"}`)
	})

	if err := handler(c); err != nil {
		t.Fatalf("handler error: %v", err)
	}

	if got := rec.Header().Get("X-FHIR-ResourceType"); got != "" {
		t.Errorf("X-FHIR-ResourceType = %q, want empty (disabled)", got)
	}
	if got := rec.Header().Get("X-FHIR-ResourceId"); got != "" {
		t.Errorf("X-FHIR-ResourceId = %q, want empty (disabled)", got)
	}
}

func TestHeadMethodMiddleware_AllowedPaths(t *testing.T) {
	e := echo.New()
	cfg := DefaultHeadMethodConfig()
	cfg.AllowedPaths = []string{"/fhir/Patient"}

	// Allowed path.
	req := httptest.NewRequest(http.MethodHead, "/fhir/Patient/1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := HeadMethodMiddleware(&cfg)(func(c echo.Context) error {
		return c.String(http.StatusOK, `{"resourceType":"Patient","id":"1"}`)
	})

	if err := handler(c); err != nil {
		t.Fatalf("handler error: %v", err)
	}

	if rec.Body.Len() != 0 {
		t.Error("expected empty body for allowed HEAD path")
	}
}

func TestHeadMethodMiddleware_DisallowedPath(t *testing.T) {
	e := echo.New()
	cfg := DefaultHeadMethodConfig()
	cfg.AllowedPaths = []string{"/fhir/Patient"}

	// Disallowed path -- should return 405 Method Not Allowed.
	req := httptest.NewRequest(http.MethodHead, "/fhir/Observation/1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := HeadMethodMiddleware(&cfg)(func(c echo.Context) error {
		return c.String(http.StatusOK, `{"resourceType":"Observation","id":"1"}`)
	})

	err := handler(c)
	if err == nil {
		t.Fatal("expected error for disallowed HEAD path")
	}
	httpErr, ok := err.(*echo.HTTPError)
	if !ok {
		t.Fatalf("expected *echo.HTTPError, got %T", err)
	}
	if httpErr.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want %d", httpErr.Code, http.StatusMethodNotAllowed)
	}
}

func TestHeadMethodMiddleware_HandlerError(t *testing.T) {
	e := echo.New()
	cfg := DefaultHeadMethodConfig()
	req := httptest.NewRequest(http.MethodHead, "/fhir/Patient/1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := HeadMethodMiddleware(&cfg)(func(c echo.Context) error {
		return echo.NewHTTPError(http.StatusInternalServerError, "database error")
	})

	err := handler(c)
	if err == nil {
		t.Fatal("expected error from handler")
	}
	httpErr, ok := err.(*echo.HTTPError)
	if !ok {
		t.Fatalf("expected *echo.HTTPError, got %T", err)
	}
	if httpErr.Code != http.StatusInternalServerError {
		t.Errorf("error code = %d, want %d", httpErr.Code, http.StatusInternalServerError)
	}
}

func TestHeadMethodMiddleware_NilConfig(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodHead, "/fhir/Patient/1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Passing nil should use defaults.
	handler := HeadMethodMiddleware(nil)(func(c echo.Context) error {
		return c.String(http.StatusOK, `{"resourceType":"Patient","id":"1"}`)
	})

	if err := handler(c); err != nil {
		t.Fatalf("handler error: %v", err)
	}

	if rec.Body.Len() != 0 {
		t.Error("expected empty body for HEAD with nil config")
	}
}

// ---------------------------------------------------------------------------
// TestHeadMethodMiddleware Integration Tests
// ---------------------------------------------------------------------------

func TestHeadMethodMiddleware_FHIRReadEndpoint(t *testing.T) {
	e := echo.New()
	cfg := DefaultHeadMethodConfig()
	req := httptest.NewRequest(http.MethodHead, "/fhir/Patient/123", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	patient := `{"resourceType":"Patient","id":"123","meta":{"versionId":"3","lastUpdated":"2024-06-15T08:00:00Z"},"name":[{"family":"Doe","given":["Jane"]}]}`
	handler := HeadMethodMiddleware(&cfg)(func(c echo.Context) error {
		c.Response().Header().Set("ETag", `W/"3"`)
		c.Response().Header().Set("Last-Modified", "2024-06-15T08:00:00Z")
		c.Response().Header().Set("Content-Type", "application/fhir+json; charset=utf-8")
		return c.String(http.StatusOK, patient)
	})

	if err := handler(c); err != nil {
		t.Fatalf("handler error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if rec.Body.Len() != 0 {
		t.Errorf("expected empty body, got %d bytes", rec.Body.Len())
	}
	if got := rec.Header().Get("ETag"); got != `W/"3"` {
		t.Errorf("ETag = %q, want W/\"3\"", got)
	}
	if got := rec.Header().Get("Last-Modified"); got != "2024-06-15T08:00:00Z" {
		t.Errorf("Last-Modified = %q, want 2024-06-15T08:00:00Z", got)
	}
	if got := rec.Header().Get("X-FHIR-ResourceType"); got != "Patient" {
		t.Errorf("X-FHIR-ResourceType = %q, want Patient", got)
	}
	if got := rec.Header().Get("X-FHIR-ResourceId"); got != "123" {
		t.Errorf("X-FHIR-ResourceId = %q, want 123", got)
	}
	expected := fmt.Sprintf("%d", len(patient))
	if got := rec.Header().Get("Content-Length"); got != expected {
		t.Errorf("Content-Length = %q, want %q", got, expected)
	}
}

func TestHeadMethodMiddleware_SearchBundle(t *testing.T) {
	e := echo.New()
	cfg := DefaultHeadMethodConfig()
	req := httptest.NewRequest(http.MethodHead, "/fhir/Patient?name=Smith", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	bundle := `{"resourceType":"Bundle","type":"searchset","total":2,"entry":[{"resource":{"resourceType":"Patient","id":"1"}},{"resource":{"resourceType":"Patient","id":"2"}}]}`
	handler := HeadMethodMiddleware(&cfg)(func(c echo.Context) error {
		c.Response().Header().Set("Content-Type", "application/fhir+json; charset=utf-8")
		return c.String(http.StatusOK, bundle)
	})

	if err := handler(c); err != nil {
		t.Fatalf("handler error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if rec.Body.Len() != 0 {
		t.Errorf("expected empty body for HEAD, got %d bytes", rec.Body.Len())
	}
	expected := fmt.Sprintf("%d", len(bundle))
	if got := rec.Header().Get("Content-Length"); got != expected {
		t.Errorf("Content-Length = %q, want %q", got, expected)
	}
}

func TestHeadMethodMiddleware_ErrorResponse(t *testing.T) {
	e := echo.New()
	cfg := DefaultHeadMethodConfig()
	req := httptest.NewRequest(http.MethodHead, "/fhir/Patient/1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := HeadMethodMiddleware(&cfg)(func(c echo.Context) error {
		c.Response().Header().Set("Content-Type", "application/fhir+json; charset=utf-8")
		return c.String(http.StatusBadRequest, `{"resourceType":"OperationOutcome","issue":[{"severity":"error"}]}`)
	})

	if err := handler(c); err != nil {
		t.Fatalf("handler error: %v", err)
	}

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
	if rec.Body.Len() != 0 {
		t.Errorf("expected empty body even for error, got %d bytes", rec.Body.Len())
	}
}

func TestHeadMethodMiddleware_404Response(t *testing.T) {
	e := echo.New()
	cfg := DefaultHeadMethodConfig()
	req := httptest.NewRequest(http.MethodHead, "/fhir/Patient/nonexistent", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := HeadMethodMiddleware(&cfg)(func(c echo.Context) error {
		c.Response().Header().Set("Content-Type", "application/fhir+json; charset=utf-8")
		return c.String(http.StatusNotFound, `{"resourceType":"OperationOutcome","issue":[{"severity":"error","code":"not-found"}]}`)
	})

	if err := handler(c); err != nil {
		t.Fatalf("handler error: %v", err)
	}

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
	if rec.Body.Len() != 0 {
		t.Errorf("expected empty body for 404 HEAD, got %d bytes", rec.Body.Len())
	}
}

// ---------------------------------------------------------------------------
// Edge Cases
// ---------------------------------------------------------------------------

func TestHeadMethodMiddleware_EmptyResponse(t *testing.T) {
	e := echo.New()
	cfg := DefaultHeadMethodConfig()
	req := httptest.NewRequest(http.MethodHead, "/fhir/Patient/1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := HeadMethodMiddleware(&cfg)(func(c echo.Context) error {
		return c.String(http.StatusOK, "")
	})

	if err := handler(c); err != nil {
		t.Fatalf("handler error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if rec.Body.Len() != 0 {
		t.Errorf("expected empty body, got %d bytes", rec.Body.Len())
	}
}

func TestHeadMethodMiddleware_VeryLargeBody(t *testing.T) {
	e := echo.New()
	cfg := DefaultHeadMethodConfig()
	req := httptest.NewRequest(http.MethodHead, "/fhir/Patient", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Build a large body (~100KB).
	largeBody := `{"resourceType":"Bundle","type":"searchset","entry":[`
	for i := 0; i < 1000; i++ {
		if i > 0 {
			largeBody += ","
		}
		largeBody += fmt.Sprintf(`{"resource":{"resourceType":"Patient","id":"%d"}}`, i)
	}
	largeBody += "]}"

	handler := HeadMethodMiddleware(&cfg)(func(c echo.Context) error {
		return c.String(http.StatusOK, largeBody)
	})

	if err := handler(c); err != nil {
		t.Fatalf("handler error: %v", err)
	}

	if rec.Body.Len() != 0 {
		t.Errorf("expected empty body for HEAD, got %d bytes", rec.Body.Len())
	}

	expected := fmt.Sprintf("%d", len(largeBody))
	if got := rec.Header().Get("Content-Length"); got != expected {
		t.Errorf("Content-Length = %q, want %q", got, expected)
	}
}

func TestHeadMethodMiddleware_MultipleContentTypes(t *testing.T) {
	e := echo.New()
	cfg := DefaultHeadMethodConfig()
	req := httptest.NewRequest(http.MethodHead, "/fhir/Patient/1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := HeadMethodMiddleware(&cfg)(func(c echo.Context) error {
		c.Response().Header().Set("Content-Type", "application/fhir+json; charset=utf-8")
		c.Response().Header().Add("Content-Type", "application/json")
		return c.String(http.StatusOK, `{"resourceType":"Patient","id":"1"}`)
	})

	if err := handler(c); err != nil {
		t.Fatalf("handler error: %v", err)
	}

	if rec.Body.Len() != 0 {
		t.Errorf("expected empty body, got %d bytes", rec.Body.Len())
	}

	// Content-Type header should be preserved.
	ct := rec.Header().Get("Content-Type")
	if ct == "" {
		t.Error("expected Content-Type header to be present")
	}
}

func TestHeadMethodMiddleware_NoContentType(t *testing.T) {
	e := echo.New()
	cfg := DefaultHeadMethodConfig()
	req := httptest.NewRequest(http.MethodHead, "/fhir/Patient/1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := HeadMethodMiddleware(&cfg)(func(c echo.Context) error {
		c.Response().WriteHeader(http.StatusNoContent)
		return nil
	})

	if err := handler(c); err != nil {
		t.Fatalf("handler error: %v", err)
	}

	if rec.Code != http.StatusNoContent {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
	if rec.Body.Len() != 0 {
		t.Errorf("expected empty body, got %d bytes", rec.Body.Len())
	}
}

func TestHeadMethodMiddleware_ConvertsToGETInternally(t *testing.T) {
	e := echo.New()
	cfg := DefaultHeadMethodConfig()
	req := httptest.NewRequest(http.MethodHead, "/fhir/Patient/1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	var capturedMethod string
	handler := HeadMethodMiddleware(&cfg)(func(c echo.Context) error {
		capturedMethod = c.Request().Method
		return c.String(http.StatusOK, `{"resourceType":"Patient","id":"1"}`)
	})

	if err := handler(c); err != nil {
		t.Fatalf("handler error: %v", err)
	}

	if capturedMethod != http.MethodGet {
		t.Errorf("handler saw method %q, want GET (HEAD should be converted)", capturedMethod)
	}
}

func TestHeadMethodMiddleware_CacheHeaders(t *testing.T) {
	e := echo.New()
	cfg := DefaultHeadMethodConfig()
	cfg.CacheHeaders = true
	cfg.CacheTTL = 5 * time.Minute

	body := `{"resourceType":"Patient","id":"1"}`
	callCount := 0

	mw := HeadMethodMiddleware(&cfg)

	// First request -- populates cache.
	req1 := httptest.NewRequest(http.MethodHead, "/fhir/Patient/1", nil)
	rec1 := httptest.NewRecorder()
	c1 := e.NewContext(req1, rec1)

	handler1 := mw(func(c echo.Context) error {
		callCount++
		c.Response().Header().Set("ETag", `W/"1"`)
		return c.String(http.StatusOK, body)
	})

	if err := handler1(c1); err != nil {
		t.Fatalf("first handler error: %v", err)
	}

	if callCount != 1 {
		t.Errorf("expected handler called once, got %d", callCount)
	}

	// Second request -- should use cache.
	req2 := httptest.NewRequest(http.MethodHead, "/fhir/Patient/1", nil)
	rec2 := httptest.NewRecorder()
	c2 := e.NewContext(req2, rec2)

	handler2 := mw(func(c echo.Context) error {
		callCount++
		c.Response().Header().Set("ETag", `W/"2"`)
		return c.String(http.StatusOK, body)
	})

	if err := handler2(c2); err != nil {
		t.Fatalf("second handler error: %v", err)
	}

	if callCount != 1 {
		t.Errorf("expected handler called only once (cached), got %d", callCount)
	}

	// Verify cached headers are correct.
	if got := rec2.Header().Get("ETag"); got != `W/"1"` {
		t.Errorf("cached ETag = %q, want W/\"1\"", got)
	}
}

func TestHeadMethodMiddleware_WriteMethods_InvalidateCache(t *testing.T) {
	e := echo.New()
	cfg := DefaultHeadMethodConfig()
	cfg.CacheHeaders = true
	cfg.CacheTTL = 5 * time.Minute

	// PUT/POST/DELETE should not be intercepted but we verify
	// that the middleware does not affect non-HEAD methods.
	for _, method := range []string{http.MethodPut, http.MethodPost, http.MethodDelete, http.MethodPatch} {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/fhir/Patient/1", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			body := `{"resourceType":"Patient","id":"1"}`
			handler := HeadMethodMiddleware(&cfg)(func(c echo.Context) error {
				return c.String(http.StatusOK, body)
			})

			if err := handler(c); err != nil {
				t.Fatalf("handler error: %v", err)
			}

			// Non-HEAD methods should pass through with body intact.
			if rec.Body.String() != body {
				t.Errorf("body = %q, want %q", rec.Body.String(), body)
			}
		})
	}
}

func TestHeadMethodMiddleware_PreservesStatusCode(t *testing.T) {
	codes := []int{
		http.StatusOK,
		http.StatusCreated,
		http.StatusNoContent,
		http.StatusNotFound,
		http.StatusGone,
		http.StatusBadRequest,
	}

	for _, code := range codes {
		t.Run(fmt.Sprintf("status_%d", code), func(t *testing.T) {
			e := echo.New()
			cfg := DefaultHeadMethodConfig()
			req := httptest.NewRequest(http.MethodHead, "/fhir/Patient/1", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			handler := HeadMethodMiddleware(&cfg)(func(c echo.Context) error {
				return c.String(code, `{"resourceType":"OperationOutcome"}`)
			})

			if err := handler(c); err != nil {
				t.Fatalf("handler error: %v", err)
			}

			if rec.Code != code {
				t.Errorf("status = %d, want %d", rec.Code, code)
			}
		})
	}
}

func TestExtractResourceMetadata_ObservationWithMeta(t *testing.T) {
	body := []byte(`{
		"resourceType": "Observation",
		"id": "obs-42",
		"meta": {
			"versionId": "12",
			"lastUpdated": "2025-02-10T15:00:00Z"
		},
		"status": "final"
	}`)

	meta := ExtractResourceMetadata(body)

	if meta["resourceType"] != "Observation" {
		t.Errorf("resourceType = %q, want Observation", meta["resourceType"])
	}
	if meta["id"] != "obs-42" {
		t.Errorf("id = %q, want obs-42", meta["id"])
	}
	if meta["versionId"] != "12" {
		t.Errorf("versionId = %q, want 12", meta["versionId"])
	}
	if meta["lastUpdated"] != "2025-02-10T15:00:00Z" {
		t.Errorf("lastUpdated = %q, want 2025-02-10T15:00:00Z", meta["lastUpdated"])
	}
}

func TestHeadCache_InvalidateNonExistentKey(t *testing.T) {
	cache := NewHeadCache(5 * time.Minute)

	// Should not panic.
	cache.Invalidate("nonexistent-key")

	_, ok := cache.Get("nonexistent-key")
	if ok {
		t.Error("expected cache miss for nonexistent key after invalidate")
	}
}

func TestHeadCache_ZeroTTL(t *testing.T) {
	cache := NewHeadCache(0)

	cache.Set("Patient/1", &HeadResponse{StatusCode: http.StatusOK})

	// With zero TTL, entries should expire immediately.
	_, ok := cache.Get("Patient/1")
	if ok {
		t.Error("expected cache miss with zero TTL")
	}
}

func TestBuildHeadHeaders_PreservesCustomHeaders(t *testing.T) {
	resp := &HeadResponse{
		StatusCode:  http.StatusOK,
		ContentType: "application/fhir+json",
		Headers: map[string][]string{
			"X-Custom-Header":   {"custom-value"},
			"X-Another-Header":  {"another-value"},
			"X-Multi-Value":     {"val1", "val2"},
		},
	}

	headers := BuildHeadHeaders(resp)

	if got := headers.Get("X-Custom-Header"); got != "custom-value" {
		t.Errorf("X-Custom-Header = %q, want custom-value", got)
	}
	if got := headers.Get("X-Another-Header"); got != "another-value" {
		t.Errorf("X-Another-Header = %q, want another-value", got)
	}
	vals := headers.Values("X-Multi-Value")
	if len(vals) != 2 {
		t.Errorf("X-Multi-Value has %d values, want 2", len(vals))
	}
}

func TestGenerateCacheKey_ConsistentForSameInput(t *testing.T) {
	h := http.Header{}
	h.Set("Accept", "application/fhir+json")
	h.Set("Authorization", "Bearer token123")

	keys := make([]string, 10)
	for i := 0; i < 10; i++ {
		keys[i] = GenerateCacheKey("HEAD", "/fhir/Patient/1", h)
	}

	for i := 1; i < 10; i++ {
		if keys[i] != keys[0] {
			t.Errorf("key[%d] = %q, differs from key[0] = %q", i, keys[i], keys[0])
		}
	}
}
