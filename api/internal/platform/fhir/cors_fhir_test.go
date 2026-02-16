package fhir

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
)

// ---------------------------------------------------------------------------
// DefaultFHIRCORSConfig
// ---------------------------------------------------------------------------

func TestDefaultFHIRCORSConfig(t *testing.T) {
	cfg := DefaultFHIRCORSConfig()

	if len(cfg.AllowOrigins) != 1 || cfg.AllowOrigins[0] != "*" {
		t.Errorf("expected AllowOrigins [\"*\"], got %v", cfg.AllowOrigins)
	}
	if cfg.AllowCredentials {
		t.Error("expected AllowCredentials false")
	}
	if cfg.MaxAge != 3600 {
		t.Errorf("expected MaxAge 3600, got %d", cfg.MaxAge)
	}
}

// ---------------------------------------------------------------------------
// Preflight (OPTIONS) requests
// ---------------------------------------------------------------------------

func TestFHIRCORS_PreflightReturns204(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodOptions, "/fhir/Patient", nil)
	req.Header.Set("Origin", "https://app.example.com")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := FHIRCORSMiddleware()(func(c echo.Context) error {
		t.Fatal("next handler should not be called for preflight")
		return nil
	})

	if err := handler(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

func TestFHIRCORS_PreflightAllowMethods(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodOptions, "/fhir/Patient", nil)
	req.Header.Set("Origin", "https://app.example.com")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := FHIRCORSMiddleware()(func(c echo.Context) error {
		return nil
	})

	if err := handler(c); err != nil {
		t.Fatal(err)
	}

	methods := rec.Header().Get("Access-Control-Allow-Methods")
	for _, m := range []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"} {
		if !strings.Contains(methods, m) {
			t.Errorf("expected Allow-Methods to contain %s, got: %s", m, methods)
		}
	}
}

func TestFHIRCORS_PreflightAllowHeaders(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodOptions, "/fhir/Patient", nil)
	req.Header.Set("Origin", "https://app.example.com")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := FHIRCORSMiddleware()(func(c echo.Context) error {
		return nil
	})

	if err := handler(c); err != nil {
		t.Fatal(err)
	}

	headers := rec.Header().Get("Access-Control-Allow-Headers")
	expected := []string{
		// Standard
		"Content-Type", "Authorization", "Accept", "Cache-Control",
		// FHIR-specific
		"Prefer", "If-Match", "If-None-Match", "If-Modified-Since", "If-None-Exist",
		// Custom
		"X-Tenant-ID", "X-Request-ID", "X-Break-Glass", "X-Security-Labels", "X-Purpose-Of-Use",
	}
	for _, h := range expected {
		if !strings.Contains(headers, h) {
			t.Errorf("expected Allow-Headers to contain %s, got: %s", h, headers)
		}
	}
}

func TestFHIRCORS_PreflightMaxAge(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodOptions, "/fhir/Patient", nil)
	req.Header.Set("Origin", "https://app.example.com")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := FHIRCORSMiddleware()(func(c echo.Context) error {
		return nil
	})

	if err := handler(c); err != nil {
		t.Fatal(err)
	}

	maxAge := rec.Header().Get("Access-Control-Max-Age")
	if maxAge != "3600" {
		t.Errorf("expected Max-Age 3600, got %s", maxAge)
	}
}

func TestFHIRCORS_PreflightCustomMaxAge(t *testing.T) {
	cfg := FHIRCORSConfig{
		AllowOrigins: []string{"*"},
		MaxAge:       7200,
	}

	e := echo.New()
	req := httptest.NewRequest(http.MethodOptions, "/fhir/Patient", nil)
	req.Header.Set("Origin", "https://app.example.com")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := FHIRCORSMiddleware(cfg)(func(c echo.Context) error {
		return nil
	})

	if err := handler(c); err != nil {
		t.Fatal(err)
	}

	maxAge := rec.Header().Get("Access-Control-Max-Age")
	if maxAge != "7200" {
		t.Errorf("expected Max-Age 7200, got %s", maxAge)
	}
}

// ---------------------------------------------------------------------------
// Normal (non-preflight) requests
// ---------------------------------------------------------------------------

func TestFHIRCORS_NormalRequestSetsOrigin(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient", nil)
	req.Header.Set("Origin", "https://app.example.com")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := FHIRCORSMiddleware()(func(c echo.Context) error {
		return c.String(http.StatusOK, `{"resourceType":"Patient"}`)
	})

	if err := handler(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	origin := rec.Header().Get("Access-Control-Allow-Origin")
	if origin != "*" {
		t.Errorf("expected Allow-Origin *, got %s", origin)
	}
}

func TestFHIRCORS_NormalRequestExposeHeaders(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient", nil)
	req.Header.Set("Origin", "https://app.example.com")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := FHIRCORSMiddleware()(func(c echo.Context) error {
		return c.String(http.StatusOK, `{"resourceType":"Patient"}`)
	})

	if err := handler(c); err != nil {
		t.Fatal(err)
	}

	exposed := rec.Header().Get("Access-Control-Expose-Headers")
	expected := []string{
		"ETag", "Last-Modified", "Location", "Content-Location", "X-Request-ID",
		"X-RateLimit-Limit", "X-RateLimit-Remaining", "X-RateLimit-Reset", "Retry-After",
	}
	for _, h := range expected {
		if !strings.Contains(exposed, h) {
			t.Errorf("expected Expose-Headers to contain %s, got: %s", h, exposed)
		}
	}
}

func TestFHIRCORS_NormalRequestDoesNotSetPreflightHeaders(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient", nil)
	req.Header.Set("Origin", "https://app.example.com")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := FHIRCORSMiddleware()(func(c echo.Context) error {
		return c.String(http.StatusOK, `{"resourceType":"Patient"}`)
	})

	if err := handler(c); err != nil {
		t.Fatal(err)
	}

	// These headers should only appear on preflight responses.
	if v := rec.Header().Get("Access-Control-Allow-Methods"); v != "" {
		t.Errorf("expected no Allow-Methods on normal request, got: %s", v)
	}
	if v := rec.Header().Get("Access-Control-Allow-Headers"); v != "" {
		t.Errorf("expected no Allow-Headers on normal request, got: %s", v)
	}
	if v := rec.Header().Get("Access-Control-Max-Age"); v != "" {
		t.Errorf("expected no Max-Age on normal request, got: %s", v)
	}
}

// ---------------------------------------------------------------------------
// No Origin header (non-CORS request)
// ---------------------------------------------------------------------------

func TestFHIRCORS_NoOriginSkipsCORSHeaders(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := FHIRCORSMiddleware()(func(c echo.Context) error {
		return c.String(http.StatusOK, `{"resourceType":"Patient"}`)
	})

	if err := handler(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	if v := rec.Header().Get("Access-Control-Allow-Origin"); v != "" {
		t.Errorf("expected no CORS headers without Origin, got Allow-Origin: %s", v)
	}
}

// ---------------------------------------------------------------------------
// Credentials support
// ---------------------------------------------------------------------------

func TestFHIRCORS_CredentialsEnabled(t *testing.T) {
	cfg := FHIRCORSConfig{
		AllowOrigins:     []string{"https://trusted.example.com"},
		AllowCredentials: true,
		MaxAge:           3600,
	}

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient", nil)
	req.Header.Set("Origin", "https://trusted.example.com")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := FHIRCORSMiddleware(cfg)(func(c echo.Context) error {
		return c.String(http.StatusOK, `{"resourceType":"Patient"}`)
	})

	if err := handler(c); err != nil {
		t.Fatal(err)
	}

	creds := rec.Header().Get("Access-Control-Allow-Credentials")
	if creds != "true" {
		t.Errorf("expected Allow-Credentials true, got %s", creds)
	}

	origin := rec.Header().Get("Access-Control-Allow-Origin")
	if origin != "https://trusted.example.com" {
		t.Errorf("expected specific origin, got %s", origin)
	}
}

func TestFHIRCORS_CredentialsDisabledByDefault(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient", nil)
	req.Header.Set("Origin", "https://app.example.com")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := FHIRCORSMiddleware()(func(c echo.Context) error {
		return c.String(http.StatusOK, `{"resourceType":"Patient"}`)
	})

	if err := handler(c); err != nil {
		t.Fatal(err)
	}

	creds := rec.Header().Get("Access-Control-Allow-Credentials")
	if creds != "" {
		t.Errorf("expected no Allow-Credentials header, got %s", creds)
	}
}

// ---------------------------------------------------------------------------
// Origin restriction
// ---------------------------------------------------------------------------

func TestFHIRCORS_SpecificOriginsAllowed(t *testing.T) {
	cfg := FHIRCORSConfig{
		AllowOrigins: []string{"https://app.example.com", "https://admin.example.com"},
		MaxAge:       3600,
	}

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient", nil)
	req.Header.Set("Origin", "https://admin.example.com")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := FHIRCORSMiddleware(cfg)(func(c echo.Context) error {
		return c.String(http.StatusOK, `{"resourceType":"Patient"}`)
	})

	if err := handler(c); err != nil {
		t.Fatal(err)
	}

	origin := rec.Header().Get("Access-Control-Allow-Origin")
	if origin != "https://admin.example.com" {
		t.Errorf("expected https://admin.example.com, got %s", origin)
	}
}

func TestFHIRCORS_DisallowedOriginNoCORSHeaders(t *testing.T) {
	cfg := FHIRCORSConfig{
		AllowOrigins: []string{"https://app.example.com"},
		MaxAge:       3600,
	}

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient", nil)
	req.Header.Set("Origin", "https://evil.example.com")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := FHIRCORSMiddleware(cfg)(func(c echo.Context) error {
		return c.String(http.StatusOK, `{"resourceType":"Patient"}`)
	})

	if err := handler(c); err != nil {
		t.Fatal(err)
	}

	// Response should still succeed but without CORS headers.
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	if v := rec.Header().Get("Access-Control-Allow-Origin"); v != "" {
		t.Errorf("expected no Allow-Origin for disallowed origin, got %s", v)
	}
}

// ---------------------------------------------------------------------------
// Vary header
// ---------------------------------------------------------------------------

func TestFHIRCORS_VaryHeaderSetForSpecificOrigin(t *testing.T) {
	cfg := FHIRCORSConfig{
		AllowOrigins: []string{"https://app.example.com"},
		MaxAge:       3600,
	}

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient", nil)
	req.Header.Set("Origin", "https://app.example.com")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := FHIRCORSMiddleware(cfg)(func(c echo.Context) error {
		return c.String(http.StatusOK, `{"resourceType":"Patient"}`)
	})

	if err := handler(c); err != nil {
		t.Fatal(err)
	}

	vary := rec.Header().Get("Vary")
	if !strings.Contains(vary, "Origin") {
		t.Errorf("expected Vary to contain Origin for specific origin config, got: %s", vary)
	}
}

func TestFHIRCORS_VaryHeaderNotSetForWildcard(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient", nil)
	req.Header.Set("Origin", "https://app.example.com")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := FHIRCORSMiddleware()(func(c echo.Context) error {
		return c.String(http.StatusOK, `{"resourceType":"Patient"}`)
	})

	if err := handler(c); err != nil {
		t.Fatal(err)
	}

	vary := rec.Header().Get("Vary")
	if strings.Contains(vary, "Origin") {
		t.Errorf("expected no Vary: Origin for wildcard config, got: %s", vary)
	}
}

// ---------------------------------------------------------------------------
// Preflight with credentials and specific origin
// ---------------------------------------------------------------------------

func TestFHIRCORS_PreflightWithCredentials(t *testing.T) {
	cfg := FHIRCORSConfig{
		AllowOrigins:     []string{"https://portal.example.com"},
		AllowCredentials: true,
		MaxAge:           1800,
	}

	e := echo.New()
	req := httptest.NewRequest(http.MethodOptions, "/fhir/Patient", nil)
	req.Header.Set("Origin", "https://portal.example.com")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := FHIRCORSMiddleware(cfg)(func(c echo.Context) error {
		return nil
	})

	if err := handler(c); err != nil {
		t.Fatal(err)
	}

	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
	if v := rec.Header().Get("Access-Control-Allow-Origin"); v != "https://portal.example.com" {
		t.Errorf("expected specific origin, got %s", v)
	}
	if v := rec.Header().Get("Access-Control-Allow-Credentials"); v != "true" {
		t.Errorf("expected Allow-Credentials true, got %s", v)
	}
	if v := rec.Header().Get("Access-Control-Max-Age"); v != "1800" {
		t.Errorf("expected Max-Age 1800, got %s", v)
	}
}

// ---------------------------------------------------------------------------
// resolveAllowOrigin helper
// ---------------------------------------------------------------------------

func TestResolveAllowOrigin(t *testing.T) {
	tests := []struct {
		name     string
		allowed  []string
		origin   string
		expected string
	}{
		{"wildcard", []string{"*"}, "https://any.example.com", "*"},
		{"exact match", []string{"https://app.example.com"}, "https://app.example.com", "https://app.example.com"},
		{"no match", []string{"https://app.example.com"}, "https://evil.example.com", ""},
		{"multiple allowed", []string{"https://a.example.com", "https://b.example.com"}, "https://b.example.com", "https://b.example.com"},
		{"empty allowed", []string{}, "https://app.example.com", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resolveAllowOrigin(tt.allowed, tt.origin)
			if got != tt.expected {
				t.Errorf("resolveAllowOrigin(%v, %q) = %q, want %q", tt.allowed, tt.origin, got, tt.expected)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Multiple HTTP methods with CORS
// ---------------------------------------------------------------------------

func TestFHIRCORS_PostRequestWithOrigin(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/fhir/Patient", nil)
	req.Header.Set("Origin", "https://app.example.com")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	called := false
	handler := FHIRCORSMiddleware()(func(c echo.Context) error {
		called = true
		return c.String(http.StatusCreated, `{"resourceType":"Patient","id":"1"}`)
	})

	if err := handler(c); err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Error("expected next handler to be called for POST")
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}

	origin := rec.Header().Get("Access-Control-Allow-Origin")
	if origin != "*" {
		t.Errorf("expected Allow-Origin *, got %s", origin)
	}
}
