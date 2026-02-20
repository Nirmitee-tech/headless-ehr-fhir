package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

func TestAuthSkipper_PublicPaths(t *testing.T) {
	publicPaths := []string{
		"/health",
		"/health/db",
		"/metrics",
		"/.well-known/smart-configuration",
		"/fhir/metadata",
	}

	for _, path := range publicPaths {
		t.Run(path, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, path, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetPath(path)

			if !AuthSkipper(c) {
				t.Errorf("expected AuthSkipper to return true for %s", path)
			}
		})
	}
}

func TestAuthSkipper_ProtectedPaths(t *testing.T) {
	protectedPaths := []string{
		"/api/v1/patients",
		"/fhir/Patient",
		"/fhir/Observation",
		"/api/v1/admin",
		"/",
		"/health/extra",
	}

	for _, path := range protectedPaths {
		t.Run(path, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, path, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetPath(path)

			if AuthSkipper(c) {
				t.Errorf("expected AuthSkipper to return false for %s", path)
			}
		})
	}
}

func TestIsPublicPath(t *testing.T) {
	if !IsPublicPath("/health") {
		t.Error("expected /health to be public")
	}
	if !IsPublicPath("/metrics") {
		t.Error("expected /metrics to be public")
	}
	if IsPublicPath("/api/v1/patients") {
		t.Error("expected /api/v1/patients to NOT be public")
	}
}

func TestJWTMiddleware_SkipsPublicPaths(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/health")
	// No Authorization header set — normally this would fail

	var handlerCalled bool
	handler := func(c echo.Context) error {
		handlerCalled = true
		return c.String(http.StatusOK, "ok")
	}

	cfg := JWTConfig{
		SigningKey: testSigningKey,
		Skipper:   AuthSkipper,
	}
	mw := JWTMiddleware(cfg)
	h := mw(handler)
	err := h(c)

	if err != nil {
		t.Fatalf("expected no error for skipped path, got: %v", err)
	}
	if !handlerCalled {
		t.Error("expected handler to be called for skipped path")
	}
}

func TestJWTMiddleware_DoesNotSkipProtectedPaths(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/patients", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/api/v1/patients")
	// No Authorization header — should fail

	handler := func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	}

	cfg := JWTConfig{
		SigningKey: testSigningKey,
		Skipper:   AuthSkipper,
	}
	mw := JWTMiddleware(cfg)
	h := mw(handler)
	err := h(c)

	if err == nil {
		t.Fatal("expected error for protected path without auth")
	}
	httpErr, ok := err.(*echo.HTTPError)
	if !ok {
		t.Fatalf("expected echo.HTTPError, got %T", err)
	}
	if httpErr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", httpErr.Code)
	}
}

func TestJWTMiddleware_NilSkipperDoesNotSkip(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/health")
	// No Skipper set, no auth header — should fail

	handler := func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	}

	cfg := JWTConfig{
		SigningKey: testSigningKey,
		// Skipper is nil — no skipping
	}
	mw := JWTMiddleware(cfg)
	h := mw(handler)
	err := h(c)

	if err == nil {
		t.Fatal("expected error when skipper is nil and no auth header")
	}
}

func TestDevAuthMiddleware_SkipsPublicPaths(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/health")

	var handlerCalled bool
	handler := func(c echo.Context) error {
		handlerCalled = true
		// Verify that dev defaults are NOT set when the path is skipped
		ctx := c.Request().Context()
		uid := UserIDFromContext(ctx)
		if uid != "" {
			t.Errorf("expected empty user_id on skipped path, got %s", uid)
		}
		return c.String(http.StatusOK, "ok")
	}

	mw := DevAuthMiddleware(AuthSkipper)
	h := mw(handler)
	err := h(c)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !handlerCalled {
		t.Error("handler was not called for skipped path")
	}
}

func TestDevAuthMiddleware_NoSkipper_StillWorks(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/patients", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	var handlerCalled bool
	handler := func(c echo.Context) error {
		handlerCalled = true
		ctx := c.Request().Context()
		uid := UserIDFromContext(ctx)
		if uid != "dev-user" {
			t.Errorf("expected dev-user, got %s", uid)
		}
		return c.String(http.StatusOK, "ok")
	}

	// No skipper argument — backwards compatible
	mw := DevAuthMiddleware()
	h := mw(handler)
	err := h(c)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !handlerCalled {
		t.Error("handler was not called")
	}
}

func TestJWTMiddleware_SkipsMetricsPath(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/metrics")

	var handlerCalled bool
	handler := func(c echo.Context) error {
		handlerCalled = true
		return c.String(http.StatusOK, "metrics")
	}

	cfg := JWTConfig{
		SigningKey: testSigningKey,
		Skipper:   AuthSkipper,
	}
	mw := JWTMiddleware(cfg)
	h := mw(handler)
	err := h(c)

	if err != nil {
		t.Fatalf("expected no error for /metrics, got: %v", err)
	}
	if !handlerCalled {
		t.Error("handler was not called for /metrics")
	}
}

func TestJWTMiddleware_SkipsFhirMetadata(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/fhir/metadata", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/fhir/metadata")

	var handlerCalled bool
	handler := func(c echo.Context) error {
		handlerCalled = true
		return c.String(http.StatusOK, "capability")
	}

	cfg := JWTConfig{
		SigningKey: testSigningKey,
		Skipper:   AuthSkipper,
	}
	mw := JWTMiddleware(cfg)
	h := mw(handler)
	err := h(c)

	if err != nil {
		t.Fatalf("expected no error for /fhir/metadata, got: %v", err)
	}
	if !handlerCalled {
		t.Error("handler was not called for /fhir/metadata")
	}
}

func TestJWTMiddleware_AuthStillWorksWithSkipper(t *testing.T) {
	// Ensure that protected paths still enforce auth when skipper is configured
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "user-789",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		TenantID:   "tenant-1",
		Roles:      []string{"physician"},
		FHIRScopes: []string{"patient/*.read"},
	}
	tokenStr := createTestToken(t, claims, testSigningKey)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/patients", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/api/v1/patients")

	var handlerCalled bool
	handler := func(c echo.Context) error {
		handlerCalled = true
		ctx := c.Request().Context()
		uid := UserIDFromContext(ctx)
		if uid != "user-789" {
			t.Errorf("expected user-789, got %s", uid)
		}
		return c.String(http.StatusOK, "ok")
	}

	cfg := JWTConfig{
		SigningKey: testSigningKey,
		Skipper:   AuthSkipper,
	}
	mw := JWTMiddleware(cfg)
	h := mw(handler)
	err := h(c)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !handlerCalled {
		t.Error("handler was not called")
	}
}
