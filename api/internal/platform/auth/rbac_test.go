package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestMatchScope(t *testing.T) {
	tests := []struct {
		granted  string
		required string
		want     bool
	}{
		{"Patient.read", "Patient.read", true},
		{"Patient.write", "Patient.read", false},
		{"user/*.*", "Patient.read", true},
		{"user/*.*", "Encounter.write", true},
		{"patient/*.read", "Patient.read", true},
		{"patient/*.read", "Patient.write", false},
		{"Patient.read", "Encounter.read", false},
		{"", "Patient.read", false},
		{"Patient.read", "", false},
		{"invalid", "Patient.read", false},
	}

	for _, tt := range tests {
		got := matchScope(tt.granted, tt.required)
		if got != tt.want {
			t.Errorf("matchScope(%q, %q) = %v, want %v", tt.granted, tt.required, got, tt.want)
		}
	}
}

func TestRequireRole_Allowed(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx := context.WithValue(req.Context(), UserRolesKey, []string{"physician"})
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	}

	mw := RequireRole("physician", "nurse")
	h := mw(handler)
	err := h(c)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestRequireRole_Denied(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx := context.WithValue(req.Context(), UserRolesKey, []string{"billing"})
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	}

	mw := RequireRole("physician", "nurse")
	h := mw(handler)
	err := h(c)

	if err == nil {
		t.Error("expected error for unauthorized role")
	}
	httpErr, ok := err.(*echo.HTTPError)
	if !ok {
		t.Fatalf("expected echo.HTTPError, got %T", err)
	}
	if httpErr.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", httpErr.Code)
	}
}

func TestRequireRole_AdminBypass(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx := context.WithValue(req.Context(), UserRolesKey, []string{"admin"})
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	}

	mw := RequireRole("physician")
	h := mw(handler)
	err := h(c)

	if err != nil {
		t.Error("admin should bypass role checks")
	}
}

func TestRequireScope_Allowed(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx := context.WithValue(req.Context(), UserScopesKey, []string{"Patient.read", "Encounter.read"})
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	}

	mw := RequireScope("Patient", "read")
	h := mw(handler)
	err := h(c)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestRequireScope_Denied(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx := context.WithValue(req.Context(), UserScopesKey, []string{"Patient.read"})
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	}

	mw := RequireScope("Patient", "write")
	h := mw(handler)
	err := h(c)

	if err == nil {
		t.Error("expected error for missing scope")
	}
}

func TestDevAuthMiddleware(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := func(c echo.Context) error {
		uid := UserIDFromContext(c.Request().Context())
		roles := RolesFromContext(c.Request().Context())
		if uid != "dev-user" {
			t.Errorf("expected dev-user, got %s", uid)
		}
		if len(roles) != 1 || roles[0] != "admin" {
			t.Errorf("expected [admin] roles, got %v", roles)
		}
		return c.String(http.StatusOK, "ok")
	}

	mw := DevAuthMiddleware()
	h := mw(handler)
	err := h(c)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestUserIDFromContext(t *testing.T) {
	ctx := context.WithValue(context.Background(), UserIDKey, "user-123")
	uid := UserIDFromContext(ctx)
	if uid != "user-123" {
		t.Errorf("expected user-123, got %s", uid)
	}

	empty := UserIDFromContext(context.Background())
	if empty != "" {
		t.Errorf("expected empty string, got %s", empty)
	}
}
