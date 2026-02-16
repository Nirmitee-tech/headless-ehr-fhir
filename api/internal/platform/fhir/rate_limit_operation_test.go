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
// classifyOperation tests
// ---------------------------------------------------------------------------

func TestClassifyOperation_SystemExport(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/fhir/$export", nil)
	c := e.NewContext(req, httptest.NewRecorder())
	if got := classifyOperation(c); got != "$export" {
		t.Errorf("classifyOperation = %q, want %q", got, "$export")
	}
}

func TestClassifyOperation_SystemImport(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/fhir/$import", nil)
	c := e.NewContext(req, httptest.NewRecorder())
	if got := classifyOperation(c); got != "$import" {
		t.Errorf("classifyOperation = %q, want %q", got, "$import")
	}
}

func TestClassifyOperation_SystemValidate(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/fhir/$validate", nil)
	c := e.NewContext(req, httptest.NewRecorder())
	if got := classifyOperation(c); got != "$validate" {
		t.Errorf("classifyOperation = %q, want %q", got, "$validate")
	}
}

func TestClassifyOperation_SystemConvert(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/fhir/$convert", nil)
	c := e.NewContext(req, httptest.NewRecorder())
	if got := classifyOperation(c); got != "$convert" {
		t.Errorf("classifyOperation = %q, want %q", got, "$convert")
	}
}

func TestClassifyOperation_ResourceOperation(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/fhir/Patient/$everything", nil)
	c := e.NewContext(req, httptest.NewRecorder())
	if got := classifyOperation(c); got != "$everything" {
		t.Errorf("classifyOperation = %q, want %q", got, "$everything")
	}
}

func TestClassifyOperation_InstanceOperation(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/fhir/Patient/123/$everything", nil)
	c := e.NewContext(req, httptest.NewRecorder())
	if got := classifyOperation(c); got != "$everything" {
		t.Errorf("classifyOperation = %q, want %q", got, "$everything")
	}
}

func TestClassifyOperation_Search(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient", nil)
	c := e.NewContext(req, httptest.NewRecorder())
	if got := classifyOperation(c); got != "search" {
		t.Errorf("classifyOperation = %q, want %q", got, "search")
	}
}

func TestClassifyOperation_SearchWithParams(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient?name=John", nil)
	c := e.NewContext(req, httptest.NewRecorder())
	if got := classifyOperation(c); got != "search" {
		t.Errorf("classifyOperation = %q, want %q", got, "search")
	}
}

func TestClassifyOperation_Read(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient/123", nil)
	c := e.NewContext(req, httptest.NewRecorder())
	if got := classifyOperation(c); got != "read" {
		t.Errorf("classifyOperation = %q, want %q", got, "read")
	}
}

func TestClassifyOperation_Create(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/fhir/Patient", nil)
	c := e.NewContext(req, httptest.NewRecorder())
	if got := classifyOperation(c); got != "create" {
		t.Errorf("classifyOperation = %q, want %q", got, "create")
	}
}

func TestClassifyOperation_Update(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPut, "/fhir/Patient/123", nil)
	c := e.NewContext(req, httptest.NewRecorder())
	if got := classifyOperation(c); got != "update" {
		t.Errorf("classifyOperation = %q, want %q", got, "update")
	}
}

func TestClassifyOperation_Patch(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPatch, "/fhir/Patient/123", nil)
	c := e.NewContext(req, httptest.NewRecorder())
	if got := classifyOperation(c); got != "update" {
		t.Errorf("classifyOperation = %q, want %q", got, "update")
	}
}

func TestClassifyOperation_Delete(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodDelete, "/fhir/Patient/123", nil)
	c := e.NewContext(req, httptest.NewRecorder())
	if got := classifyOperation(c); got != "delete" {
		t.Errorf("classifyOperation = %q, want %q", got, "delete")
	}
}

func TestClassifyOperation_HistorySearch(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient/_history", nil)
	c := e.NewContext(req, httptest.NewRecorder())
	if got := classifyOperation(c); got != "search" {
		t.Errorf("classifyOperation = %q, want %q", got, "search")
	}
}

func TestClassifyOperation_UnknownMethod(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodOptions, "/fhir/Patient", nil)
	c := e.NewContext(req, httptest.NewRecorder())
	if got := classifyOperation(c); got != "other" {
		t.Errorf("classifyOperation = %q, want %q", got, "other")
	}
}

// ---------------------------------------------------------------------------
// OperationRateLimiter unit tests
// ---------------------------------------------------------------------------

func TestOperationRateLimiter_DefaultConfig(t *testing.T) {
	limiter := NewOperationRateLimiter(100, time.Minute)

	cfg := limiter.GetLimit("read")
	if cfg.MaxRequests != 100 {
		t.Errorf("default MaxRequests = %d, want 100", cfg.MaxRequests)
	}
	if cfg.Window != time.Minute {
		t.Errorf("default Window = %v, want %v", cfg.Window, time.Minute)
	}
	if cfg.Operation != "read" {
		t.Errorf("Operation = %q, want %q", cfg.Operation, "read")
	}
}

func TestOperationRateLimiter_SetOperationLimit(t *testing.T) {
	limiter := NewOperationRateLimiter(100, time.Minute)
	limiter.SetOperationLimit("$export", 5, 10*time.Minute)

	cfg := limiter.GetLimit("$export")
	if cfg.MaxRequests != 5 {
		t.Errorf("$export MaxRequests = %d, want 5", cfg.MaxRequests)
	}
	if cfg.Window != 10*time.Minute {
		t.Errorf("$export Window = %v, want %v", cfg.Window, 10*time.Minute)
	}
	if cfg.Operation != "$export" {
		t.Errorf("Operation = %q, want %q", cfg.Operation, "$export")
	}
}

func TestOperationRateLimiter_OverrideDoesNotAffectDefault(t *testing.T) {
	limiter := NewOperationRateLimiter(100, time.Minute)
	limiter.SetOperationLimit("$export", 5, 10*time.Minute)

	// Unconfigured operation should still use defaults.
	cfg := limiter.GetLimit("search")
	if cfg.MaxRequests != 100 {
		t.Errorf("search MaxRequests = %d, want 100", cfg.MaxRequests)
	}
}

func TestOperationRateLimiter_AllowUnderLimit(t *testing.T) {
	limiter := NewOperationRateLimiter(3, time.Minute)

	for i := 0; i < 3; i++ {
		if !limiter.Allow("read", "client-a") {
			t.Fatalf("request %d should be allowed", i+1)
		}
	}
}

func TestOperationRateLimiter_DenyOverLimit(t *testing.T) {
	limiter := NewOperationRateLimiter(2, time.Minute)

	limiter.Allow("read", "client-a")
	limiter.Allow("read", "client-a")

	if limiter.Allow("read", "client-a") {
		t.Error("third request should be denied")
	}
}

func TestOperationRateLimiter_DifferentOperationsIndependent(t *testing.T) {
	limiter := NewOperationRateLimiter(1, time.Minute)

	if !limiter.Allow("read", "client-a") {
		t.Error("read should be allowed")
	}
	if !limiter.Allow("search", "client-a") {
		t.Error("search should be allowed (independent from read)")
	}

	// Second read should be denied.
	if limiter.Allow("read", "client-a") {
		t.Error("second read should be denied")
	}
	// Second search should also be denied.
	if limiter.Allow("search", "client-a") {
		t.Error("second search should be denied")
	}
}

func TestOperationRateLimiter_DifferentClientsIndependent(t *testing.T) {
	limiter := NewOperationRateLimiter(1, time.Minute)

	if !limiter.Allow("read", "client-a") {
		t.Error("client-a should be allowed")
	}
	if !limiter.Allow("read", "client-b") {
		t.Error("client-b should be allowed (independent from client-a)")
	}

	// client-a second request should be denied.
	if limiter.Allow("read", "client-a") {
		t.Error("client-a second request should be denied")
	}
}

func TestOperationRateLimiter_PerOperationOverrideLimits(t *testing.T) {
	limiter := NewOperationRateLimiter(100, time.Minute)
	limiter.SetOperationLimit("$export", 2, time.Minute)

	// $export should be limited to 2.
	if !limiter.Allow("$export", "client-a") {
		t.Error("first $export should be allowed")
	}
	if !limiter.Allow("$export", "client-a") {
		t.Error("second $export should be allowed")
	}
	if limiter.Allow("$export", "client-a") {
		t.Error("third $export should be denied")
	}

	// read should still use default limit of 100.
	for i := 0; i < 100; i++ {
		if !limiter.Allow("read", "client-a") {
			t.Fatalf("read request %d should be allowed", i+1)
		}
	}
}

func TestOperationRateLimiter_SetOperationLimitResetsLimiter(t *testing.T) {
	limiter := NewOperationRateLimiter(1, time.Minute)

	// Exhaust the limit.
	limiter.Allow("read", "client-a")
	if limiter.Allow("read", "client-a") {
		t.Error("should be denied before reconfiguration")
	}

	// Reconfigure with a higher limit; the old limiter should be removed.
	limiter.SetOperationLimit("read", 10, time.Minute)

	// Now should be allowed again since the limiter was reset.
	if !limiter.Allow("read", "client-a") {
		t.Error("should be allowed after reconfiguration")
	}
}

func TestOperationRateLimiter_GetLimitReturnsOperationName(t *testing.T) {
	limiter := NewOperationRateLimiter(50, time.Minute)

	cfg := limiter.GetLimit("$validate")
	if cfg.Operation != "$validate" {
		t.Errorf("Operation = %q, want %q", cfg.Operation, "$validate")
	}
}

// ---------------------------------------------------------------------------
// DefaultOperationRateLimits tests
// ---------------------------------------------------------------------------

func TestDefaultOperationRateLimits_ReadConfig(t *testing.T) {
	limiter := DefaultOperationRateLimits()
	cfg := limiter.GetLimit("read")
	if cfg.MaxRequests != 1000 {
		t.Errorf("read MaxRequests = %d, want 1000", cfg.MaxRequests)
	}
	if cfg.Window != time.Minute {
		t.Errorf("read Window = %v, want %v", cfg.Window, time.Minute)
	}
}

func TestDefaultOperationRateLimits_SearchConfig(t *testing.T) {
	limiter := DefaultOperationRateLimits()
	cfg := limiter.GetLimit("search")
	if cfg.MaxRequests != 500 {
		t.Errorf("search MaxRequests = %d, want 500", cfg.MaxRequests)
	}
}

func TestDefaultOperationRateLimits_CreateConfig(t *testing.T) {
	limiter := DefaultOperationRateLimits()
	cfg := limiter.GetLimit("create")
	if cfg.MaxRequests != 200 {
		t.Errorf("create MaxRequests = %d, want 200", cfg.MaxRequests)
	}
}

func TestDefaultOperationRateLimits_UpdateConfig(t *testing.T) {
	limiter := DefaultOperationRateLimits()
	cfg := limiter.GetLimit("update")
	if cfg.MaxRequests != 200 {
		t.Errorf("update MaxRequests = %d, want 200", cfg.MaxRequests)
	}
}

func TestDefaultOperationRateLimits_DeleteConfig(t *testing.T) {
	limiter := DefaultOperationRateLimits()
	cfg := limiter.GetLimit("delete")
	if cfg.MaxRequests != 200 {
		t.Errorf("delete MaxRequests = %d, want 200", cfg.MaxRequests)
	}
}

func TestDefaultOperationRateLimits_ExportConfig(t *testing.T) {
	limiter := DefaultOperationRateLimits()
	cfg := limiter.GetLimit("$export")
	if cfg.MaxRequests != 10 {
		t.Errorf("$export MaxRequests = %d, want 10", cfg.MaxRequests)
	}
}

func TestDefaultOperationRateLimits_ImportConfig(t *testing.T) {
	limiter := DefaultOperationRateLimits()
	cfg := limiter.GetLimit("$import")
	if cfg.MaxRequests != 10 {
		t.Errorf("$import MaxRequests = %d, want 10", cfg.MaxRequests)
	}
}

func TestDefaultOperationRateLimits_ValidateConfig(t *testing.T) {
	limiter := DefaultOperationRateLimits()
	cfg := limiter.GetLimit("$validate")
	if cfg.MaxRequests != 100 {
		t.Errorf("$validate MaxRequests = %d, want 100", cfg.MaxRequests)
	}
}

func TestDefaultOperationRateLimits_OtherUsesDefault(t *testing.T) {
	limiter := DefaultOperationRateLimits()
	cfg := limiter.GetLimit("$some-custom-op")
	if cfg.MaxRequests != 50 {
		t.Errorf("other MaxRequests = %d, want 50", cfg.MaxRequests)
	}
}

// ---------------------------------------------------------------------------
// OperationRateLimitMiddleware integration tests
// ---------------------------------------------------------------------------

func TestOperationRateLimitMiddleware_PassesThrough(t *testing.T) {
	limiter := NewOperationRateLimiter(10, time.Minute)
	e := echo.New()

	handler := OperationRateLimitMiddleware(limiter)(func(c echo.Context) error {
		return c.String(http.StatusOK, `{"resourceType":"Patient"}`)
	})

	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient/123", nil)
	req.RemoteAddr = "10.0.0.1:1234"
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := handler(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestOperationRateLimitMiddleware_SetsHeaders(t *testing.T) {
	limiter := NewOperationRateLimiter(10, time.Minute)
	e := echo.New()

	handler := OperationRateLimitMiddleware(limiter)(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient/123", nil)
	req.RemoteAddr = "10.0.0.1:1234"
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := handler(c); err != nil {
		t.Fatal(err)
	}

	if got := rec.Header().Get("X-RateLimit-Limit"); got != "10" {
		t.Errorf("X-RateLimit-Limit = %q, want %q", got, "10")
	}
	if got := rec.Header().Get("X-RateLimit-Remaining"); got != "9" {
		t.Errorf("X-RateLimit-Remaining = %q, want %q", got, "9")
	}
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
	if got := rec.Header().Get("X-RateLimit-Operation"); got != "read" {
		t.Errorf("X-RateLimit-Operation = %q, want %q", got, "read")
	}
}

func TestOperationRateLimitMiddleware_429WhenExceeded(t *testing.T) {
	limiter := NewOperationRateLimiter(2, time.Minute)
	e := echo.New()

	okHandler := func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	}
	mw := OperationRateLimitMiddleware(limiter)

	// Exhaust the read quota.
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodGet, "/fhir/Patient/123", nil)
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

	// Third read request should be rejected.
	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient/123", nil)
	req.RemoteAddr = "10.0.0.2:1234"
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if err := mw(okHandler)(c); err != nil {
		t.Fatal(err)
	}

	if rec.Code != http.StatusTooManyRequests {
		t.Errorf("expected 429, got %d", rec.Code)
	}

	// Verify Retry-After header.
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

	// Verify remaining is 0.
	if rem := rec.Header().Get("X-RateLimit-Remaining"); rem != "0" {
		t.Errorf("X-RateLimit-Remaining = %q, want %q", rem, "0")
	}

	// Verify body contains OperationOutcome with operation name.
	body := rec.Body.String()
	if !strings.Contains(body, "OperationOutcome") {
		t.Errorf("expected OperationOutcome in body, got: %s", body)
	}
	if !strings.Contains(body, "throttled") {
		t.Errorf("expected 'throttled' code in body, got: %s", body)
	}
	if !strings.Contains(body, "read") {
		t.Errorf("expected operation name 'read' in body, got: %s", body)
	}
}

func TestOperationRateLimitMiddleware_HandlerNotCalledWhenDenied(t *testing.T) {
	limiter := NewOperationRateLimiter(1, time.Minute)
	e := echo.New()
	mw := OperationRateLimitMiddleware(limiter)

	called := 0
	handler := func(c echo.Context) error {
		called++
		return c.String(http.StatusOK, "ok")
	}

	// First request - should call handler.
	req1 := httptest.NewRequest(http.MethodGet, "/fhir/Patient/123", nil)
	req1.RemoteAddr = "10.0.0.1:1111"
	rec1 := httptest.NewRecorder()
	c1 := e.NewContext(req1, rec1)
	if err := mw(handler)(c1); err != nil {
		t.Fatal(err)
	}

	// Second request - should NOT call handler.
	req2 := httptest.NewRequest(http.MethodGet, "/fhir/Patient/123", nil)
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

func TestOperationRateLimitMiddleware_DifferentOperationsIndependent(t *testing.T) {
	limiter := NewOperationRateLimiter(1, time.Minute)
	e := echo.New()

	okHandler := func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	}
	mw := OperationRateLimitMiddleware(limiter)

	// Exhaust read quota.
	req1 := httptest.NewRequest(http.MethodGet, "/fhir/Patient/123", nil)
	req1.RemoteAddr = "10.0.0.1:1111"
	rec1 := httptest.NewRecorder()
	c1 := e.NewContext(req1, rec1)
	if err := mw(okHandler)(c1); err != nil {
		t.Fatal(err)
	}
	if rec1.Code != http.StatusOK {
		t.Fatalf("read: expected 200, got %d", rec1.Code)
	}

	// Search should still be allowed (different operation).
	req2 := httptest.NewRequest(http.MethodGet, "/fhir/Patient", nil)
	req2.RemoteAddr = "10.0.0.1:1111"
	rec2 := httptest.NewRecorder()
	c2 := e.NewContext(req2, rec2)
	if err := mw(okHandler)(c2); err != nil {
		t.Fatal(err)
	}
	if rec2.Code != http.StatusOK {
		t.Fatalf("search: expected 200, got %d", rec2.Code)
	}

	// Second read should be denied.
	req3 := httptest.NewRequest(http.MethodGet, "/fhir/Patient/456", nil)
	req3.RemoteAddr = "10.0.0.1:1111"
	rec3 := httptest.NewRecorder()
	c3 := e.NewContext(req3, rec3)
	if err := mw(okHandler)(c3); err != nil {
		t.Fatal(err)
	}
	if rec3.Code != http.StatusTooManyRequests {
		t.Errorf("second read: expected 429, got %d", rec3.Code)
	}
}

func TestOperationRateLimitMiddleware_UsesAPIKeyAsClient(t *testing.T) {
	limiter := NewOperationRateLimiter(1, time.Minute)
	e := echo.New()

	okHandler := func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	}
	mw := OperationRateLimitMiddleware(limiter)

	// First request with API key A.
	req1 := httptest.NewRequest(http.MethodGet, "/fhir/Patient/123", nil)
	req1.Header.Set("X-API-Key", "key-a")
	req1.RemoteAddr = "10.0.0.1:1111"
	rec1 := httptest.NewRecorder()
	c1 := e.NewContext(req1, rec1)
	if err := mw(okHandler)(c1); err != nil {
		t.Fatal(err)
	}
	if rec1.Code != http.StatusOK {
		t.Fatalf("key-a first request: expected 200, got %d", rec1.Code)
	}

	// Second request with same API key should be denied.
	req2 := httptest.NewRequest(http.MethodGet, "/fhir/Patient/123", nil)
	req2.Header.Set("X-API-Key", "key-a")
	req2.RemoteAddr = "10.0.0.2:2222" // different IP, same key
	rec2 := httptest.NewRecorder()
	c2 := e.NewContext(req2, rec2)
	if err := mw(okHandler)(c2); err != nil {
		t.Fatal(err)
	}
	if rec2.Code != http.StatusTooManyRequests {
		t.Errorf("key-a second request: expected 429, got %d", rec2.Code)
	}

	// Request with different API key should be allowed.
	req3 := httptest.NewRequest(http.MethodGet, "/fhir/Patient/123", nil)
	req3.Header.Set("X-API-Key", "key-b")
	req3.RemoteAddr = "10.0.0.1:1111"
	rec3 := httptest.NewRecorder()
	c3 := e.NewContext(req3, rec3)
	if err := mw(okHandler)(c3); err != nil {
		t.Fatal(err)
	}
	if rec3.Code != http.StatusOK {
		t.Errorf("key-b first request: expected 200, got %d", rec3.Code)
	}
}

func TestOperationRateLimitMiddleware_PerOperationLimitsInHeaders(t *testing.T) {
	limiter := NewOperationRateLimiter(50, time.Minute)
	limiter.SetOperationLimit("$export", 5, time.Minute)
	e := echo.New()

	okHandler := func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	}
	mw := OperationRateLimitMiddleware(limiter)

	// $export request should show limit of 5.
	req := httptest.NewRequest(http.MethodPost, "/fhir/$export", nil)
	req.RemoteAddr = "10.0.0.1:1234"
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if err := mw(okHandler)(c); err != nil {
		t.Fatal(err)
	}

	if got := rec.Header().Get("X-RateLimit-Limit"); got != "5" {
		t.Errorf("X-RateLimit-Limit = %q, want %q", got, "5")
	}
	if got := rec.Header().Get("X-RateLimit-Operation"); got != "$export" {
		t.Errorf("X-RateLimit-Operation = %q, want %q", got, "$export")
	}
}

func TestOperationRateLimitMiddleware_RemainingDecrementsCorrectly(t *testing.T) {
	limiter := NewOperationRateLimiter(3, time.Minute)
	e := echo.New()
	mw := OperationRateLimitMiddleware(limiter)

	okHandler := func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	}

	expected := []string{"2", "1", "0"}
	for i, want := range expected {
		req := httptest.NewRequest(http.MethodGet, "/fhir/Patient/123", nil)
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

func TestOperationRateLimitMiddleware_FallsBackToIPWithoutAPIKey(t *testing.T) {
	limiter := NewOperationRateLimiter(1, time.Minute)
	e := echo.New()

	okHandler := func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	}
	mw := OperationRateLimitMiddleware(limiter)

	// Request without API key from IP A.
	req1 := httptest.NewRequest(http.MethodGet, "/fhir/Patient/123", nil)
	req1.RemoteAddr = "10.0.0.1:1111"
	rec1 := httptest.NewRecorder()
	c1 := e.NewContext(req1, rec1)
	if err := mw(okHandler)(c1); err != nil {
		t.Fatal(err)
	}
	if rec1.Code != http.StatusOK {
		t.Fatalf("first request: expected 200, got %d", rec1.Code)
	}

	// Second request from same IP should be denied.
	req2 := httptest.NewRequest(http.MethodGet, "/fhir/Patient/123", nil)
	req2.RemoteAddr = "10.0.0.1:2222"
	rec2 := httptest.NewRecorder()
	c2 := e.NewContext(req2, rec2)
	if err := mw(okHandler)(c2); err != nil {
		t.Fatal(err)
	}
	if rec2.Code != http.StatusTooManyRequests {
		t.Errorf("second request: expected 429, got %d", rec2.Code)
	}
}

func TestOperationRateLimitMiddleware_NoRetryAfterOnAllowed(t *testing.T) {
	limiter := NewOperationRateLimiter(10, time.Minute)
	e := echo.New()

	handler := OperationRateLimitMiddleware(limiter)(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient/123", nil)
	req.RemoteAddr = "10.0.0.1:1234"
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := handler(c); err != nil {
		t.Fatal(err)
	}
	if ra := rec.Header().Get("Retry-After"); ra != "" {
		t.Errorf("Retry-After should not be set on allowed request, got %q", ra)
	}
}
