package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

var testSigningKey = []byte("test-secret-key-for-unit-tests-only")

func createTestToken(t *testing.T, claims Claims, key []byte) string {
	t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString(key)
	if err != nil {
		t.Fatalf("failed to sign test token: %v", err)
	}
	return tokenStr
}

func TestJWTMiddleware_MissingHeader(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	}

	cfg := JWTConfig{SigningKey: testSigningKey}
	mw := JWTMiddleware(cfg)
	h := mw(handler)
	err := h(c)

	if err == nil {
		t.Fatal("expected error for missing header")
	}
	httpErr, ok := err.(*echo.HTTPError)
	if !ok {
		t.Fatalf("expected echo.HTTPError, got %T", err)
	}
	if httpErr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", httpErr.Code)
	}
}

func TestJWTMiddleware_InvalidFormat(t *testing.T) {
	tests := []struct {
		name   string
		header string
	}{
		{"no bearer prefix", "Token abc123"},
		{"missing token", "Bearer"},
		{"empty value", "Bearer "},
		{"basic auth", "Basic dXNlcjpwYXNz"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set("Authorization", tt.header)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			handler := func(c echo.Context) error {
				return c.String(http.StatusOK, "ok")
			}

			cfg := JWTConfig{SigningKey: testSigningKey}
			mw := JWTMiddleware(cfg)
			h := mw(handler)
			err := h(c)

			if err == nil {
				t.Fatal("expected error for invalid format")
			}
			httpErr, ok := err.(*echo.HTTPError)
			if !ok {
				t.Fatalf("expected echo.HTTPError, got %T", err)
			}
			if httpErr.Code != http.StatusUnauthorized {
				t.Errorf("expected 401, got %d", httpErr.Code)
			}
		})
	}
}

func TestJWTMiddleware_ValidToken(t *testing.T) {
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "user-123",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		TenantID:   "tenant-1",
		Roles:      []string{"physician"},
		FHIRScopes: []string{"patient/*.read"},
	}

	tokenStr := createTestToken(t, claims, testSigningKey)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	var handlerCalled bool
	handler := func(c echo.Context) error {
		handlerCalled = true
		return c.String(http.StatusOK, "ok")
	}

	cfg := JWTConfig{SigningKey: testSigningKey}
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

func TestJWTMiddleware_ExpiredToken(t *testing.T) {
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "user-123",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
		},
		TenantID: "tenant-1",
	}

	tokenStr := createTestToken(t, claims, testSigningKey)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	}

	cfg := JWTConfig{SigningKey: testSigningKey}
	mw := JWTMiddleware(cfg)
	h := mw(handler)
	err := h(c)

	if err == nil {
		t.Fatal("expected error for expired token")
	}
	httpErr, ok := err.(*echo.HTTPError)
	if !ok {
		t.Fatalf("expected echo.HTTPError, got %T", err)
	}
	if httpErr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", httpErr.Code)
	}
}

func TestJWTMiddleware_ClaimsExtraction(t *testing.T) {
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "user-456",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		TenantID:   "tenant-abc",
		Roles:      []string{"physician", "surgeon"},
		FHIRScopes: []string{"patient/*.read", "patient/*.write"},
	}

	tokenStr := createTestToken(t, claims, testSigningKey)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := func(c echo.Context) error {
		ctx := c.Request().Context()

		uid := UserIDFromContext(ctx)
		if uid != "user-456" {
			t.Errorf("expected user_id=user-456, got %s", uid)
		}

		roles := RolesFromContext(ctx)
		if len(roles) != 2 || roles[0] != "physician" || roles[1] != "surgeon" {
			t.Errorf("expected roles=[physician surgeon], got %v", roles)
		}

		scopes := ScopesFromContext(ctx)
		if len(scopes) != 2 || scopes[0] != "patient/*.read" || scopes[1] != "patient/*.write" {
			t.Errorf("expected scopes=[patient/*.read patient/*.write], got %v", scopes)
		}

		tenantID, _ := c.Get("jwt_tenant_id").(string)
		if tenantID != "tenant-abc" {
			t.Errorf("expected tenant_id=tenant-abc, got %s", tenantID)
		}

		return c.String(http.StatusOK, "ok")
	}

	cfg := JWTConfig{SigningKey: testSigningKey}
	mw := JWTMiddleware(cfg)
	h := mw(handler)
	err := h(c)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDevAuthMiddleware_NoToken(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	var handlerCalled bool
	handler := func(c echo.Context) error {
		handlerCalled = true
		return c.String(http.StatusOK, "ok")
	}

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

func TestDevAuthMiddleware_WithDefaults(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := func(c echo.Context) error {
		ctx := c.Request().Context()

		uid := UserIDFromContext(ctx)
		if uid != "dev-user" {
			t.Errorf("expected user_id=dev-user, got %s", uid)
		}

		roles := RolesFromContext(ctx)
		if len(roles) != 1 || roles[0] != "admin" {
			t.Errorf("expected roles=[admin], got %v", roles)
		}

		scopes := ScopesFromContext(ctx)
		if len(scopes) != 1 || scopes[0] != "user/*.*" {
			t.Errorf("expected scopes=[user/*.*], got %v", scopes)
		}

		tenantID, _ := c.Get("jwt_tenant_id").(string)
		if tenantID != "default" {
			t.Errorf("expected tenant_id=default, got %s", tenantID)
		}

		return c.String(http.StatusOK, "ok")
	}

	mw := DevAuthMiddleware()
	h := mw(handler)
	err := h(c)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
