package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
)

// helper creates an echo context with the given roles set on the request context.
func newContextWithRoles(method, path string, roles []string) (echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	req := httptest.NewRequest(method, path, nil)
	ctx := context.WithValue(req.Context(), UserRolesKey, roles)
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	return c, rec
}

// helper creates an echo context with the given scopes set on the request context.
func newContextWithScopes(method, path string, scopes []string) (echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	req := httptest.NewRequest(method, path, nil)
	ctx := context.WithValue(req.Context(), UserScopesKey, scopes)
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	return c, rec
}

var okHandler = func(c echo.Context) error {
	return c.String(http.StatusOK, "ok")
}

// TestRequireRole_AdminAccessesAll verifies that the admin role can access any
// role-protected endpoint regardless of which roles are listed.
func TestRequireRole_AdminAccessesAll(t *testing.T) {
	domainRoles := [][]string{
		{"physician", "nurse"},
		{"billing"},
		{"lab_tech", "radiologist"},
		{"pharmacist"},
		{"surgeon"},
		{"registrar"},
		{"patient"},
	}

	for _, roles := range domainRoles {
		c, _ := newContextWithRoles(http.MethodGet, "/", []string{"admin"})
		mw := RequireRole(roles...)
		err := mw(okHandler)(c)
		if err != nil {
			t.Errorf("admin should access endpoint requiring %v, got error: %v", roles, err)
		}
	}
}

// TestRequireRole_PhysicianAccessesClinical verifies that a physician can access
// clinical domain endpoints which list "physician" as a permitted role.
func TestRequireRole_PhysicianAccessesClinical(t *testing.T) {
	clinicalRoles := []string{"admin", "physician", "nurse"}

	c, _ := newContextWithRoles(http.MethodGet, "/conditions", []string{"physician"})
	mw := RequireRole(clinicalRoles...)
	err := mw(okHandler)(c)
	if err != nil {
		t.Errorf("physician should access clinical endpoints, got error: %v", err)
	}

	// Also verify write access
	c, _ = newContextWithRoles(http.MethodPost, "/conditions", []string{"physician"})
	mw = RequireRole(clinicalRoles...)
	err = mw(okHandler)(c)
	if err != nil {
		t.Errorf("physician should write to clinical endpoints, got error: %v", err)
	}
}

// TestRequireRole_NurseAccessesClinical verifies that a nurse can access
// clinical domain endpoints which list "nurse" as a permitted role.
func TestRequireRole_NurseAccessesClinical(t *testing.T) {
	// Clinical read: admin, physician, nurse
	c, _ := newContextWithRoles(http.MethodGet, "/conditions", []string{"nurse"})
	mw := RequireRole("admin", "physician", "nurse")
	err := mw(okHandler)(c)
	if err != nil {
		t.Errorf("nurse should read clinical endpoints, got error: %v", err)
	}

	// Nursing write: admin, nurse (physician NOT included for write)
	c, _ = newContextWithRoles(http.MethodPost, "/nursing-assessments", []string{"nurse"})
	mw = RequireRole("admin", "nurse")
	err = mw(okHandler)(c)
	if err != nil {
		t.Errorf("nurse should write to nursing endpoints, got error: %v", err)
	}
}

// TestRequireRole_PharmacistAccessesMedication verifies that a pharmacist can
// access medication domain endpoints.
func TestRequireRole_PharmacistAccessesMedication(t *testing.T) {
	// Medication read: admin, physician, nurse, pharmacist
	c, _ := newContextWithRoles(http.MethodGet, "/medications", []string{"pharmacist"})
	mw := RequireRole("admin", "physician", "nurse", "pharmacist")
	err := mw(okHandler)(c)
	if err != nil {
		t.Errorf("pharmacist should read medication endpoints, got error: %v", err)
	}

	// Medication write: admin, physician, pharmacist
	c, _ = newContextWithRoles(http.MethodPost, "/medications", []string{"pharmacist"})
	mw = RequireRole("admin", "physician", "pharmacist")
	err = mw(okHandler)(c)
	if err != nil {
		t.Errorf("pharmacist should write to medication endpoints, got error: %v", err)
	}
}

// TestRequireRole_BillingAccessesBilling verifies that a billing role can
// access billing domain endpoints.
func TestRequireRole_BillingAccessesBilling(t *testing.T) {
	// Billing read: admin, billing
	c, _ := newContextWithRoles(http.MethodGet, "/claims", []string{"billing"})
	mw := RequireRole("admin", "billing")
	err := mw(okHandler)(c)
	if err != nil {
		t.Errorf("billing role should read billing endpoints, got error: %v", err)
	}

	// Billing write: admin, billing
	c, _ = newContextWithRoles(http.MethodPost, "/claims", []string{"billing"})
	mw = RequireRole("admin", "billing")
	err = mw(okHandler)(c)
	if err != nil {
		t.Errorf("billing role should write to billing endpoints, got error: %v", err)
	}
}

// TestRequireRole_BillingDeniedClinical verifies that a billing role cannot
// access clinical domain endpoints.
func TestRequireRole_BillingDeniedClinical(t *testing.T) {
	// Clinical read: admin, physician, nurse -- billing NOT included
	c, _ := newContextWithRoles(http.MethodGet, "/conditions", []string{"billing"})
	mw := RequireRole("admin", "physician", "nurse")
	err := mw(okHandler)(c)
	if err == nil {
		t.Error("billing role should NOT access clinical endpoints")
	}
	httpErr, ok := err.(*echo.HTTPError)
	if !ok {
		t.Fatalf("expected echo.HTTPError, got %T", err)
	}
	if httpErr.Code != http.StatusForbidden {
		t.Errorf("expected 403 Forbidden, got %d", httpErr.Code)
	}
}

// TestRequireRole_PatientDeniedAdmin verifies that a patient role cannot
// access admin domain endpoints.
func TestRequireRole_PatientDeniedAdmin(t *testing.T) {
	// Admin read: admin, physician, nurse, registrar -- patient NOT included
	c, _ := newContextWithRoles(http.MethodGet, "/organizations", []string{"patient"})
	mw := RequireRole("admin", "physician", "nurse", "registrar")
	err := mw(okHandler)(c)
	if err == nil {
		t.Error("patient role should NOT access admin endpoints")
	}
	httpErr, ok := err.(*echo.HTTPError)
	if !ok {
		t.Fatalf("expected echo.HTTPError, got %T", err)
	}
	if httpErr.Code != http.StatusForbidden {
		t.Errorf("expected 403 Forbidden, got %d", httpErr.Code)
	}

	// Admin write: admin only
	c, _ = newContextWithRoles(http.MethodPost, "/organizations", []string{"patient"})
	mw = RequireRole("admin")
	err = mw(okHandler)(c)
	if err == nil {
		t.Error("patient role should NOT write to admin endpoints")
	}
}

// TestRequireRole_NoRoleDenied verifies that a request with no roles is denied
// access to any role-protected endpoint.
func TestRequireRole_NoRoleDenied(t *testing.T) {
	// Empty roles slice
	c, _ := newContextWithRoles(http.MethodGet, "/conditions", []string{})
	mw := RequireRole("admin", "physician", "nurse")
	err := mw(okHandler)(c)
	if err == nil {
		t.Error("empty roles should be denied")
	}

	// Nil roles (no context value)
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/conditions", nil)
	rec := httptest.NewRecorder()
	c = e.NewContext(req, rec)
	err = mw(okHandler)(c)
	if err == nil {
		t.Error("nil roles should be denied")
	}
	httpErr, ok := err.(*echo.HTTPError)
	if !ok {
		t.Fatalf("expected echo.HTTPError, got %T", err)
	}
	if httpErr.Code != http.StatusForbidden {
		t.Errorf("expected 403 Forbidden, got %d", httpErr.Code)
	}
}

// TestRequireScope_MatchesExact verifies that an exact scope grant matches
// the required scope.
func TestRequireScope_MatchesExact(t *testing.T) {
	tests := []struct {
		name     string
		scopes   []string
		resource string
		op       string
		wantErr  bool
	}{
		{"exact match read", []string{"Patient.read"}, "Patient", "read", false},
		{"exact match write", []string{"Patient.write"}, "Patient", "write", false},
		{"mismatch operation", []string{"Patient.read"}, "Patient", "write", true},
		{"mismatch resource", []string{"Patient.read"}, "Encounter", "read", true},
		{"multiple scopes hit", []string{"Encounter.read", "Patient.read"}, "Patient", "read", false},
		{"multiple scopes miss", []string{"Encounter.read", "Observation.read"}, "Patient", "read", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, _ := newContextWithScopes(http.MethodGet, "/", tt.scopes)
			mw := RequireScope(tt.resource, tt.op)
			err := mw(okHandler)(c)
			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("expected no error, got %v", err)
			}
		})
	}
}

// TestRequireScope_WildcardGrant verifies that wildcard scope grants cover
// specific scope requirements.
func TestRequireScope_WildcardGrant(t *testing.T) {
	tests := []struct {
		name     string
		scopes   []string
		resource string
		op       string
		wantErr  bool
	}{
		{"user wildcard covers read", []string{"user/*.*"}, "Patient", "read", false},
		{"user wildcard covers write", []string{"user/*.*"}, "Encounter", "write", false},
		{"patient wildcard read covers Patient", []string{"patient/*.read"}, "Patient", "read", false},
		{"patient wildcard read blocks write", []string{"patient/*.read"}, "Patient", "write", true},
		{"resource wildcard op", []string{"Patient.*"}, "Patient", "read", false},
		{"resource wildcard op write", []string{"Patient.*"}, "Patient", "write", false},
		{"resource wildcard wrong resource", []string{"Patient.*"}, "Encounter", "read", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, _ := newContextWithScopes(http.MethodGet, "/", tt.scopes)
			mw := RequireScope(tt.resource, tt.op)
			err := mw(okHandler)(c)
			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("expected no error, got %v", err)
			}
		})
	}
}
