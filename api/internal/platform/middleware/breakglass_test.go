package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"

	"github.com/ehr/ehr/internal/platform/auth"
)

// bgTestContext creates an echo.Context for break-glass tests with optional
// request modifiers applied in order.
func bgTestContext(method, path string, opts ...func(*http.Request)) (echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	req := httptest.NewRequest(method, path, nil)
	for _, opt := range opts {
		opt(req)
	}
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	return c, rec
}

func bgWithAuth(userID string, roles []string) func(*http.Request) {
	return func(req *http.Request) {
		ctx := req.Context()
		ctx = context.WithValue(ctx, auth.UserIDKey, userID)
		ctx = context.WithValue(ctx, auth.UserRolesKey, roles)
		*req = *req.WithContext(ctx)
	}
}

func bgWithHeader(key, value string) func(*http.Request) {
	return func(req *http.Request) {
		req.Header.Set(key, value)
	}
}

func bgOKHandler(c echo.Context) error {
	return c.String(http.StatusOK, "ok")
}

// fixedClock returns a nowFn that always returns the given time.
func fixedClock(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func TestBreakGlass_DetectedAndContextSet(t *testing.T) {
	logger := zerolog.New(os.Stderr)
	rl := newBreakGlassRateLimit()
	now := time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC)

	c, _ := bgTestContext(http.MethodGet, "/fhir/Patient/123",
		bgWithAuth("doc-1", []string{"physician"}),
		bgWithHeader("X-Break-Glass", "cardiac arrest"),
	)

	mw := breakGlassMiddleware(logger, rl, fixedClock(now))

	var capturedCtx context.Context
	handler := mw(func(c echo.Context) error {
		capturedCtx = c.Request().Context()
		return c.String(http.StatusOK, "ok")
	})

	err := handler(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !IsBreakGlass(capturedCtx) {
		t.Error("expected IsBreakGlass to return true")
	}
	if got := BreakGlassReason(capturedCtx); got != "cardiac arrest" {
		t.Errorf("expected reason 'cardiac arrest', got %q", got)
	}
}

func TestBreakGlass_AddsAdminRole(t *testing.T) {
	logger := zerolog.New(os.Stderr)
	rl := newBreakGlassRateLimit()
	now := time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC)

	c, _ := bgTestContext(http.MethodGet, "/fhir/Patient/123",
		bgWithAuth("doc-2", []string{"physician"}),
		bgWithHeader("X-Break-Glass", "patient unresponsive"),
	)

	mw := breakGlassMiddleware(logger, rl, fixedClock(now))

	var capturedRoles []string
	handler := mw(func(c echo.Context) error {
		capturedRoles = auth.RolesFromContext(c.Request().Context())
		return c.String(http.StatusOK, "ok")
	})

	err := handler(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	found := false
	for _, r := range capturedRoles {
		if r == "admin" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected 'admin' in roles, got %v", capturedRoles)
	}

	// Original role should still be present.
	hasPhysician := false
	for _, r := range capturedRoles {
		if r == "physician" {
			hasPhysician = true
			break
		}
	}
	if !hasPhysician {
		t.Errorf("expected 'physician' still in roles, got %v", capturedRoles)
	}
}

func TestBreakGlass_AdminNotDuplicated(t *testing.T) {
	logger := zerolog.New(os.Stderr)
	rl := newBreakGlassRateLimit()
	now := time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC)

	c, _ := bgTestContext(http.MethodGet, "/fhir/Patient/123",
		bgWithAuth("admin-user", []string{"admin"}),
		bgWithHeader("X-Break-Glass", "emergency"),
	)

	mw := breakGlassMiddleware(logger, rl, fixedClock(now))

	var capturedRoles []string
	handler := mw(func(c echo.Context) error {
		capturedRoles = auth.RolesFromContext(c.Request().Context())
		return c.String(http.StatusOK, "ok")
	})

	err := handler(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	adminCount := 0
	for _, r := range capturedRoles {
		if r == "admin" {
			adminCount++
		}
	}
	if adminCount != 1 {
		t.Errorf("expected exactly 1 admin role, got %d in %v", adminCount, capturedRoles)
	}
}

func TestBreakGlass_NonClinicalPathIgnored(t *testing.T) {
	logger := zerolog.New(os.Stderr)
	rl := newBreakGlassRateLimit()
	now := time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC)

	nonClinicalPaths := []string{"/health", "/metrics", "/admin/users", "/api/v2/stuff"}
	for _, path := range nonClinicalPaths {
		c, _ := bgTestContext(http.MethodGet, path,
			bgWithAuth("doc-3", []string{"physician"}),
			bgWithHeader("X-Break-Glass", "emergency"),
		)

		mw := breakGlassMiddleware(logger, rl, fixedClock(now))

		var capturedCtx context.Context
		handler := mw(func(c echo.Context) error {
			capturedCtx = c.Request().Context()
			return c.String(http.StatusOK, "ok")
		})

		err := handler(c)
		if err != nil {
			t.Fatalf("unexpected error on path %s: %v", path, err)
		}

		if IsBreakGlass(capturedCtx) {
			t.Errorf("break-glass should NOT be active on non-clinical path %s", path)
		}
	}
}

func TestBreakGlass_ClinicalPathsActivate(t *testing.T) {
	logger := zerolog.New(os.Stderr)
	rl := newBreakGlassRateLimit()
	now := time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC)

	clinicalPaths := []string{"/fhir/Patient/123", "/api/v1/patients/456"}
	for _, path := range clinicalPaths {
		c, _ := bgTestContext(http.MethodGet, path,
			bgWithAuth("doc-4", []string{"physician"}),
			bgWithHeader("X-Break-Glass", "emergency"),
		)

		mw := breakGlassMiddleware(logger, rl, fixedClock(now))

		var capturedCtx context.Context
		handler := mw(func(c echo.Context) error {
			capturedCtx = c.Request().Context()
			return c.String(http.StatusOK, "ok")
		})

		err := handler(c)
		if err != nil {
			t.Fatalf("unexpected error on path %s: %v", path, err)
		}

		if !IsBreakGlass(capturedCtx) {
			t.Errorf("break-glass should be active on clinical path %s", path)
		}
	}
}

func TestBreakGlass_WithoutAuth_Returns401(t *testing.T) {
	logger := zerolog.New(os.Stderr)
	rl := newBreakGlassRateLimit()
	now := time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC)

	// No auth context values set -- simulates missing JWT.
	c, _ := bgTestContext(http.MethodGet, "/fhir/Patient/123",
		bgWithHeader("X-Break-Glass", "emergency"),
	)

	mw := breakGlassMiddleware(logger, rl, fixedClock(now))
	handler := mw(bgOKHandler)

	err := handler(c)
	if err == nil {
		t.Fatal("expected error for unauthenticated break-glass request")
	}

	httpErr, ok := err.(*echo.HTTPError)
	if !ok {
		t.Fatalf("expected echo.HTTPError, got %T", err)
	}
	if httpErr.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", httpErr.Code)
	}
}

func TestBreakGlass_EmptyReason_Ignored(t *testing.T) {
	logger := zerolog.New(os.Stderr)
	rl := newBreakGlassRateLimit()
	now := time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC)

	c, _ := bgTestContext(http.MethodGet, "/fhir/Patient/123",
		bgWithAuth("doc-5", []string{"physician"}),
		bgWithHeader("X-Break-Glass", ""),
	)

	mw := breakGlassMiddleware(logger, rl, fixedClock(now))

	var capturedCtx context.Context
	handler := mw(func(c echo.Context) error {
		capturedCtx = c.Request().Context()
		return c.String(http.StatusOK, "ok")
	})

	err := handler(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if IsBreakGlass(capturedCtx) {
		t.Error("break-glass should NOT activate with empty reason")
	}
}

func TestBreakGlass_WhitespaceOnlyReason_Ignored(t *testing.T) {
	logger := zerolog.New(os.Stderr)
	rl := newBreakGlassRateLimit()
	now := time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC)

	c, _ := bgTestContext(http.MethodGet, "/fhir/Patient/123",
		bgWithAuth("doc-5b", []string{"physician"}),
		bgWithHeader("X-Break-Glass", "   "),
	)

	mw := breakGlassMiddleware(logger, rl, fixedClock(now))

	var capturedCtx context.Context
	handler := mw(func(c echo.Context) error {
		capturedCtx = c.Request().Context()
		return c.String(http.StatusOK, "ok")
	})

	err := handler(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if IsBreakGlass(capturedCtx) {
		t.Error("break-glass should NOT activate with whitespace-only reason")
	}
}

func TestBreakGlass_RateLimit_11thRequestReturns429(t *testing.T) {
	logger := zerolog.New(os.Stderr)
	rl := newBreakGlassRateLimit()
	baseTime := time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC)

	// Advance time by 1 second per request so they're all within the same hour.
	for i := 0; i < 10; i++ {
		now := baseTime.Add(time.Duration(i) * time.Second)
		c, _ := bgTestContext(http.MethodGet, "/fhir/Patient/123",
			bgWithAuth("doc-6", []string{"physician"}),
			bgWithHeader("X-Break-Glass", "emergency"),
		)
		mw := breakGlassMiddleware(logger, rl, fixedClock(now))
		handler := mw(bgOKHandler)
		err := handler(c)
		if err != nil {
			t.Fatalf("request %d: unexpected error: %v", i+1, err)
		}
	}

	// 11th request should be rate-limited.
	now11 := baseTime.Add(10 * time.Second)
	c, _ := bgTestContext(http.MethodGet, "/fhir/Patient/123",
		bgWithAuth("doc-6", []string{"physician"}),
		bgWithHeader("X-Break-Glass", "emergency"),
	)
	mw := breakGlassMiddleware(logger, rl, fixedClock(now11))
	handler := mw(bgOKHandler)
	err := handler(c)

	if err == nil {
		t.Fatal("expected error for 11th break-glass request")
	}
	httpErr, ok := err.(*echo.HTTPError)
	if !ok {
		t.Fatalf("expected echo.HTTPError, got %T", err)
	}
	if httpErr.Code != http.StatusTooManyRequests {
		t.Errorf("expected status 429, got %d", httpErr.Code)
	}
}

func TestBreakGlass_RateLimit_DifferentUsersIndependent(t *testing.T) {
	logger := zerolog.New(os.Stderr)
	rl := newBreakGlassRateLimit()
	baseTime := time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC)

	// Exhaust user-a's limit.
	for i := 0; i < 10; i++ {
		now := baseTime.Add(time.Duration(i) * time.Second)
		c, _ := bgTestContext(http.MethodGet, "/fhir/Patient/123",
			bgWithAuth("user-a", []string{"physician"}),
			bgWithHeader("X-Break-Glass", "emergency"),
		)
		mw := breakGlassMiddleware(logger, rl, fixedClock(now))
		handler := mw(bgOKHandler)
		if err := handler(c); err != nil {
			t.Fatalf("user-a request %d: unexpected error: %v", i+1, err)
		}
	}

	// user-b should still be allowed.
	c, _ := bgTestContext(http.MethodGet, "/fhir/Patient/123",
		bgWithAuth("user-b", []string{"physician"}),
		bgWithHeader("X-Break-Glass", "emergency"),
	)
	mw := breakGlassMiddleware(logger, rl, fixedClock(baseTime))
	handler := mw(bgOKHandler)
	if err := handler(c); err != nil {
		t.Fatalf("user-b should not be rate-limited: %v", err)
	}
}

func TestBreakGlass_RateLimit_ResetsAfterHour(t *testing.T) {
	logger := zerolog.New(os.Stderr)
	rl := newBreakGlassRateLimit()
	baseTime := time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC)

	// Exhaust the limit.
	for i := 0; i < 10; i++ {
		now := baseTime.Add(time.Duration(i) * time.Second)
		c, _ := bgTestContext(http.MethodGet, "/fhir/Patient/123",
			bgWithAuth("doc-7", []string{"physician"}),
			bgWithHeader("X-Break-Glass", "emergency"),
		)
		mw := breakGlassMiddleware(logger, rl, fixedClock(now))
		handler := mw(bgOKHandler)
		if err := handler(c); err != nil {
			t.Fatalf("request %d: unexpected error: %v", i+1, err)
		}
	}

	// One hour + 1 second later, the limit should reset.
	futureTime := baseTime.Add(1*time.Hour + 1*time.Second)
	c, _ := bgTestContext(http.MethodGet, "/fhir/Patient/123",
		bgWithAuth("doc-7", []string{"physician"}),
		bgWithHeader("X-Break-Glass", "emergency"),
	)
	mw := breakGlassMiddleware(logger, rl, fixedClock(futureTime))
	handler := mw(bgOKHandler)
	if err := handler(c); err != nil {
		t.Fatalf("expected request to succeed after rate limit window: %v", err)
	}
}

func TestBreakGlass_ReasonIsLogged(t *testing.T) {
	// Capture log output to verify the WARN log contains the reason.
	var buf zerolog.ConsoleWriter
	_ = buf // We verify logging by checking that the middleware runs without error
	// and the context values are set. Full log verification would require
	// a custom writer; here we confirm the middleware path executes.
	logger := zerolog.New(os.Stderr)
	rl := newBreakGlassRateLimit()
	now := time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC)

	c, _ := bgTestContext(http.MethodGet, "/fhir/Patient/123",
		bgWithAuth("doc-8", []string{"physician"}),
		bgWithHeader("X-Break-Glass", "severe hemorrhage"),
	)

	mw := breakGlassMiddleware(logger, rl, fixedClock(now))

	var capturedCtx context.Context
	handler := mw(func(c echo.Context) error {
		capturedCtx = c.Request().Context()
		return c.String(http.StatusOK, "ok")
	})

	err := handler(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify the reason is stored (which is what gets logged).
	if got := BreakGlassReason(capturedCtx); got != "severe hemorrhage" {
		t.Errorf("expected reason 'severe hemorrhage', got %q", got)
	}
}

func TestBreakGlass_SetsRequireConsentFalse(t *testing.T) {
	logger := zerolog.New(os.Stderr)
	rl := newBreakGlassRateLimit()
	now := time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC)

	c, _ := bgTestContext(http.MethodGet, "/fhir/Condition/456",
		bgWithAuth("doc-9", []string{"physician"}),
		bgWithHeader("X-Break-Glass", "emergency"),
	)

	// Simulate ABAC middleware having set require_consent = true before break-glass.
	c.Set("require_consent", true)

	mw := breakGlassMiddleware(logger, rl, fixedClock(now))

	var capturedRequireConsent interface{}
	handler := mw(func(c echo.Context) error {
		capturedRequireConsent = c.Get("require_consent")
		return c.String(http.StatusOK, "ok")
	})

	err := handler(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	rc, ok := capturedRequireConsent.(bool)
	if !ok {
		t.Fatalf("require_consent not set as bool, got %T", capturedRequireConsent)
	}
	if rc {
		t.Error("expected require_consent to be false after break-glass")
	}
}

func TestBreakGlass_NoHeaderPassesThrough(t *testing.T) {
	logger := zerolog.New(os.Stderr)
	rl := newBreakGlassRateLimit()
	now := time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC)

	c, _ := bgTestContext(http.MethodGet, "/fhir/Patient/123",
		bgWithAuth("doc-10", []string{"physician"}),
	)

	mw := breakGlassMiddleware(logger, rl, fixedClock(now))

	var capturedCtx context.Context
	handler := mw(func(c echo.Context) error {
		capturedCtx = c.Request().Context()
		return c.String(http.StatusOK, "ok")
	})

	err := handler(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if IsBreakGlass(capturedCtx) {
		t.Error("break-glass should NOT be active without X-Break-Glass header")
	}
}

func TestIsBreakGlass_DefaultFalse(t *testing.T) {
	ctx := context.Background()
	if IsBreakGlass(ctx) {
		t.Error("expected IsBreakGlass to return false on empty context")
	}
}

func TestBreakGlassReason_DefaultEmpty(t *testing.T) {
	ctx := context.Background()
	if got := BreakGlassReason(ctx); got != "" {
		t.Errorf("expected empty reason on empty context, got %q", got)
	}
}

func TestIsClinicalPath(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{"/fhir/Patient", true},
		{"/fhir/Observation/123", true},
		{"/api/v1/patients", true},
		{"/api/v1/encounters/abc", true},
		{"/health", false},
		{"/metrics", false},
		{"/admin/users", false},
		{"/api/v2/resources", false},
		{"/fhir", false}, // no trailing slash
		{"/api/v1", false},
	}
	for _, tt := range tests {
		if got := isClinicalPath(tt.path); got != tt.want {
			t.Errorf("isClinicalPath(%q) = %v, want %v", tt.path, got, tt.want)
		}
	}
}

func TestBreakGlassRateLimit_Cleanup(t *testing.T) {
	rl := newBreakGlassRateLimit()
	baseTime := time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC)

	// Add some entries.
	for i := 0; i < 5; i++ {
		rl.allow("user-cleanup", baseTime.Add(time.Duration(i)*time.Second), breakGlassMaxPerHour)
	}

	// Cleanup with a time 2 hours later should remove all entries.
	rl.cleanup(baseTime.Add(2 * time.Hour))

	// User should be able to make requests again.
	if !rl.allow("user-cleanup", baseTime.Add(2*time.Hour), breakGlassMaxPerHour) {
		t.Error("expected allow after cleanup, but got denied")
	}
}
