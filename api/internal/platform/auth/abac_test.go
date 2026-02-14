package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestABACEngine_AdminBypass(t *testing.T) {
	engine := NewABACEngine(DefaultPolicies())

	ctx := context.WithValue(context.Background(), UserRolesKey, []string{"admin"})
	decision := engine.Evaluate(ctx, "Patient")

	if !decision.Allowed {
		t.Error("expected admin to be allowed")
	}
	if decision.Reason != "admin role" {
		t.Errorf("expected reason 'admin role', got %q", decision.Reason)
	}
}

func TestABACEngine_AdminBypassUnknownResource(t *testing.T) {
	engine := NewABACEngine(DefaultPolicies())

	ctx := context.WithValue(context.Background(), UserRolesKey, []string{"admin"})
	decision := engine.Evaluate(ctx, "UnknownResource")

	if !decision.Allowed {
		t.Error("expected admin to bypass even for unknown resource types")
	}
}

func TestABACEngine_PhysicianAccessPatient(t *testing.T) {
	engine := NewABACEngine(DefaultPolicies())

	ctx := context.WithValue(context.Background(), UserRolesKey, []string{"physician"})
	decision := engine.Evaluate(ctx, "Patient")

	if !decision.Allowed {
		t.Error("expected physician to access Patient")
	}
	if decision.Reason != "policy match" {
		t.Errorf("expected reason 'policy match', got %q", decision.Reason)
	}
}

func TestABACEngine_NurseAccessMedicationRequest(t *testing.T) {
	engine := NewABACEngine(DefaultPolicies())

	ctx := context.WithValue(context.Background(), UserRolesKey, []string{"nurse"})
	decision := engine.Evaluate(ctx, "MedicationRequest")

	if decision.Allowed {
		t.Error("expected nurse to be denied access to MedicationRequest")
	}
	if decision.Reason != "insufficient role for MedicationRequest" {
		t.Errorf("expected reason 'insufficient role for MedicationRequest', got %q", decision.Reason)
	}
}

func TestABACEngine_PhysicianAccessMedicationRequest(t *testing.T) {
	engine := NewABACEngine(DefaultPolicies())

	ctx := context.WithValue(context.Background(), UserRolesKey, []string{"physician"})
	decision := engine.Evaluate(ctx, "MedicationRequest")

	if !decision.Allowed {
		t.Error("expected physician to access MedicationRequest")
	}
}

func TestABACEngine_ReceptionistAccessPatient(t *testing.T) {
	engine := NewABACEngine(DefaultPolicies())

	ctx := context.WithValue(context.Background(), UserRolesKey, []string{"receptionist"})
	decision := engine.Evaluate(ctx, "Patient")

	if !decision.Allowed {
		t.Error("expected receptionist to access Patient")
	}
}

func TestABACEngine_ReceptionistDeniedCondition(t *testing.T) {
	engine := NewABACEngine(DefaultPolicies())

	ctx := context.WithValue(context.Background(), UserRolesKey, []string{"receptionist"})
	decision := engine.Evaluate(ctx, "Condition")

	if decision.Allowed {
		t.Error("expected receptionist to be denied access to Condition")
	}
}

func TestABACEngine_UnknownResourceDefaultDeny(t *testing.T) {
	engine := NewABACEngine(DefaultPolicies())

	ctx := context.WithValue(context.Background(), UserRolesKey, []string{"physician"})
	decision := engine.Evaluate(ctx, "CustomResource")

	if decision.Allowed {
		t.Error("expected unknown resource type to be denied")
	}
	if decision.Reason != "no policy for CustomResource" {
		t.Errorf("expected reason 'no policy for CustomResource', got %q", decision.Reason)
	}
}

func TestABACEngine_NoRoles(t *testing.T) {
	engine := NewABACEngine(DefaultPolicies())

	ctx := context.Background()
	decision := engine.Evaluate(ctx, "Patient")

	if decision.Allowed {
		t.Error("expected denial when no roles are present")
	}
}

func TestABACEngine_EmptyPolicies(t *testing.T) {
	engine := NewABACEngine([]ABACPolicy{})

	ctx := context.WithValue(context.Background(), UserRolesKey, []string{"physician"})
	decision := engine.Evaluate(ctx, "Patient")

	if decision.Allowed {
		t.Error("expected denial with empty policies")
	}
	if decision.Reason != "no policy for Patient" {
		t.Errorf("expected reason 'no policy for Patient', got %q", decision.Reason)
	}
}

func TestABACEngine_NurseAccessObservation(t *testing.T) {
	engine := NewABACEngine(DefaultPolicies())

	ctx := context.WithValue(context.Background(), UserRolesKey, []string{"nurse"})
	decision := engine.Evaluate(ctx, "Observation")

	if !decision.Allowed {
		t.Error("expected nurse to access Observation")
	}
}

func TestABACEngine_NurseAccessEncounter(t *testing.T) {
	engine := NewABACEngine(DefaultPolicies())

	ctx := context.WithValue(context.Background(), UserRolesKey, []string{"nurse"})
	decision := engine.Evaluate(ctx, "Encounter")

	if !decision.Allowed {
		t.Error("expected nurse to access Encounter")
	}
}

func TestExtractABACResourceType(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{"/fhir/Patient", "Patient"},
		{"/fhir/Patient/123", "Patient"},
		{"/fhir/Observation", "Observation"},
		{"/fhir/Condition/abc-123", "Condition"},
		{"/other/path", ""},
		{"/", ""},
		{"/fhir", ""},
		{"/api/v1/patients", ""},
	}

	for _, tt := range tests {
		got := extractABACResourceType(tt.path)
		if got != tt.want {
			t.Errorf("extractABACResourceType(%q) = %q, want %q", tt.path, got, tt.want)
		}
	}
}

func TestABACMiddleware_Allowed(t *testing.T) {
	engine := NewABACEngine(DefaultPolicies())

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient/123", nil)
	ctx := context.WithValue(req.Context(), UserRolesKey, []string{"physician"})
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/fhir/Patient/:id")

	handler := func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	}

	mw := ABACMiddleware(engine)
	h := mw(handler)
	err := h(c)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestABACMiddleware_Denied(t *testing.T) {
	engine := NewABACEngine(DefaultPolicies())

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/fhir/MedicationRequest/123", nil)
	ctx := context.WithValue(req.Context(), UserRolesKey, []string{"nurse"})
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/fhir/MedicationRequest/:id")

	handler := func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	}

	mw := ABACMiddleware(engine)
	h := mw(handler)
	err := h(c)

	if err == nil {
		t.Fatal("expected error for denied access")
	}
	httpErr, ok := err.(*echo.HTTPError)
	if !ok {
		t.Fatalf("expected echo.HTTPError, got %T", err)
	}
	if httpErr.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", httpErr.Code)
	}
}

func TestABACMiddleware_NonFHIRPath(t *testing.T) {
	engine := NewABACEngine(DefaultPolicies())

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/health")

	called := false
	handler := func(c echo.Context) error {
		called = true
		return c.String(http.StatusOK, "ok")
	}

	mw := ABACMiddleware(engine)
	h := mw(handler)
	err := h(c)

	if err != nil {
		t.Fatalf("expected no error for non-FHIR path, got %v", err)
	}
	if !called {
		t.Error("expected handler to be called for non-FHIR path")
	}
}

func TestConsentEnforcementMiddleware_PassThrough(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	called := false
	handler := func(c echo.Context) error {
		called = true
		return c.String(http.StatusOK, "ok")
	}

	mw := ConsentEnforcementMiddleware()
	h := mw(handler)
	err := h(c)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !called {
		t.Error("expected handler to be called")
	}
}

func TestDefaultPolicies(t *testing.T) {
	policies := DefaultPolicies()
	if len(policies) != 5 {
		t.Errorf("expected 5 default policies, got %d", len(policies))
	}

	// Verify resource types are present
	expectedTypes := map[string]bool{
		"Patient":           false,
		"Condition":         false,
		"Observation":       false,
		"MedicationRequest": false,
		"Encounter":         false,
	}
	for _, p := range policies {
		if _, ok := expectedTypes[p.ResourceType]; ok {
			expectedTypes[p.ResourceType] = true
		}
	}
	for rt, found := range expectedTypes {
		if !found {
			t.Errorf("expected default policy for %q", rt)
		}
	}
}
