package db

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestExtractTenantID_FromHeader(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Tenant-ID", "hospital_abc")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	tid := extractTenantID(c, "default")
	if tid != "hospital_abc" {
		t.Errorf("expected hospital_abc, got %s", tid)
	}
}

func TestExtractTenantID_FromQuery(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/?tenant_id=clinic_xyz", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	tid := extractTenantID(c, "default")
	if tid != "clinic_xyz" {
		t.Errorf("expected clinic_xyz, got %s", tid)
	}
}

func TestExtractTenantID_FromJWT(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("jwt_tenant_id", "jwt_tenant")

	tid := extractTenantID(c, "default")
	if tid != "jwt_tenant" {
		t.Errorf("expected jwt_tenant, got %s", tid)
	}
}

func TestExtractTenantID_Default(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	tid := extractTenantID(c, "default")
	if tid != "default" {
		t.Errorf("expected default, got %s", tid)
	}
}

func TestExtractTenantID_Priority(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/?tenant_id=query", nil)
	req.Header.Set("X-Tenant-ID", "header")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("jwt_tenant_id", "jwt")

	// JWT takes highest priority
	tid := extractTenantID(c, "default")
	if tid != "jwt" {
		t.Errorf("expected jwt (highest priority), got %s", tid)
	}
}

func TestTenantIDPattern(t *testing.T) {
	valid := []string{"abc", "hospital_1", "tenant_abc_123", "A1B2"}
	for _, v := range valid {
		if !tenantIDPattern.MatchString(v) {
			t.Errorf("expected %s to be valid", v)
		}
	}

	invalid := []string{"a-b", "a.b", "a b", "'; DROP TABLE", "a/b", ""}
	for _, v := range invalid {
		if tenantIDPattern.MatchString(v) {
			t.Errorf("expected %s to be invalid", v)
		}
	}
}

func TestConnFromContext_Nil(t *testing.T) {
	conn := ConnFromContext(context.Background())
	if conn != nil {
		t.Error("expected nil conn from empty context")
	}
}

func TestTenantFromContext(t *testing.T) {
	ctx := context.WithValue(context.Background(), TenantIDKey, "test_tenant")
	tid := TenantFromContext(ctx)
	if tid != "test_tenant" {
		t.Errorf("expected test_tenant, got %s", tid)
	}

	empty := TenantFromContext(context.Background())
	if empty != "" {
		t.Errorf("expected empty string, got %s", empty)
	}
}

func TestCreateTenantSchema_InvalidID(t *testing.T) {
	err := CreateTenantSchema(context.Background(), nil, "invalid-id!", "")
	if err == nil {
		t.Error("expected error for invalid tenant ID")
	}
}

func TestTxFromContext_Nil(t *testing.T) {
	tx := TxFromContext(context.Background())
	if tx != nil {
		t.Error("expected nil tx from empty context")
	}
}

func TestWithTx_NoConnection(t *testing.T) {
	ctx := context.Background()
	_, _, err := WithTx(ctx)
	if err == nil {
		t.Error("expected error when no connection in context")
	}
	if err.Error() != "no database connection in context" {
		t.Errorf("unexpected error message: %s", err.Error())
	}
}

func TestExtractTenantID_HeaderPriorityOverQuery(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/?tenant_id=query_tenant", nil)
	req.Header.Set("X-Tenant-ID", "header_tenant")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	tid := extractTenantID(c, "default")
	if tid != "header_tenant" {
		t.Errorf("expected header_tenant (header has priority over query), got %s", tid)
	}
}

func TestExtractTenantID_EmptyJWT(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Tenant-ID", "header_tenant")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	// Set jwt_tenant_id to empty string -- should fall through
	c.Set("jwt_tenant_id", "")

	tid := extractTenantID(c, "default")
	if tid != "header_tenant" {
		t.Errorf("expected header_tenant when JWT is empty, got %s", tid)
	}
}

func TestTenantIDPattern_Comprehensive(t *testing.T) {
	tests := []struct {
		input string
		valid bool
	}{
		{"abc", true},
		{"ABC", true},
		{"abc123", true},
		{"tenant_1", true},
		{"a", true},
		{"A1B2C3", true},
		{"a-b", false},
		{"a.b", false},
		{"a b", false},
		{"a/b", false},
		{"", false},
		{"$pecial", false},
		{"tenant@1", false},
	}

	for _, tt := range tests {
		got := tenantIDPattern.MatchString(tt.input)
		if got != tt.valid {
			t.Errorf("tenantIDPattern.MatchString(%q) = %v, want %v", tt.input, got, tt.valid)
		}
	}
}

func TestCreateTenantSchema_VariousInvalidIDs(t *testing.T) {
	invalidIDs := []string{"tenant-with-dash", "tenant.with.dot", "ten ant", "drop;table"}
	for _, id := range invalidIDs {
		err := CreateTenantSchema(context.Background(), nil, id, "")
		if err == nil {
			t.Errorf("expected error for invalid tenant ID %q", id)
		}
	}
}

func TestConnFromContext_WithValue(t *testing.T) {
	// Verify ConnFromContext returns nil for wrong type in context
	ctx := context.WithValue(context.Background(), DBConnKey, "not-a-conn")
	conn := ConnFromContext(ctx)
	if conn != nil {
		t.Error("expected nil when context value is wrong type")
	}
}

func TestTxFromContext_WithWrongType(t *testing.T) {
	ctx := context.WithValue(context.Background(), DBTxKey, "not-a-tx")
	tx := TxFromContext(ctx)
	if tx != nil {
		t.Error("expected nil when context value is wrong type")
	}
}

func TestTenantFromContext_WithWrongType(t *testing.T) {
	ctx := context.WithValue(context.Background(), TenantIDKey, 12345)
	tid := TenantFromContext(ctx)
	if tid != "" {
		t.Errorf("expected empty string when context value is wrong type, got %q", tid)
	}
}

func TestTenantMiddleware_SkipsPublicPaths(t *testing.T) {
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

			var handlerCalled bool
			handler := func(c echo.Context) error {
				handlerCalled = true
				// Verify no DB connection was set in context
				ctx := c.Request().Context()
				conn := ConnFromContext(ctx)
				if conn != nil {
					t.Error("expected no DB connection for public path")
				}
				tid := TenantFromContext(ctx)
				if tid != "" {
					t.Errorf("expected no tenant ID for public path, got %s", tid)
				}
				return c.String(http.StatusOK, "ok")
			}

			// Pass nil pool — if the middleware does NOT skip, it will panic
			// or return an error when trying to acquire from nil pool.
			mw := TenantMiddleware(nil, "default")
			h := mw(handler)
			err := h(c)

			if err != nil {
				t.Fatalf("expected no error for public path %s, got: %v", path, err)
			}
			if !handlerCalled {
				t.Errorf("handler was not called for public path %s", path)
			}
		})
	}
}

func TestTenantMiddleware_DoesNotSkipProtectedPaths(t *testing.T) {
	// For protected paths, the middleware should NOT skip and should try to
	// acquire a connection. With a nil pool this will fail, proving the
	// middleware did not skip.
	protectedPaths := []string{
		"/api/v1/patients",
		"/fhir/Patient",
		"/",
	}

	for _, path := range protectedPaths {
		t.Run(path, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, path, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetPath(path)

			handler := func(c echo.Context) error {
				return c.String(http.StatusOK, "ok")
			}

			// nil pool: the middleware will attempt pool.Acquire and panic/error.
			// We wrap in a recover to detect it was not skipped.
			mw := TenantMiddleware(nil, "default")
			h := mw(handler)

			func() {
				defer func() {
					if r := recover(); r != nil {
						// Expected — middleware tried to use nil pool
						return
					}
				}()
				err := h(c)
				if err == nil {
					t.Errorf("expected error or panic for protected path %s with nil pool", path)
				}
			}()
		})
	}
}
