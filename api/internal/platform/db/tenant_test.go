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
