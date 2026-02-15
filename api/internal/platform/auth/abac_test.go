package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
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

// ---------------------------------------------------------------------------
// ABACEngine.Evaluate - RequireConsent flag
// ---------------------------------------------------------------------------

func TestABACEngine_EvaluateReturnsRequireConsentFlag(t *testing.T) {
	engine := NewABACEngine(DefaultPolicies())

	ctx := context.WithValue(context.Background(), UserRolesKey, []string{"physician"})

	// Condition requires consent
	decision := engine.Evaluate(ctx, "Condition")
	if !decision.Allowed {
		t.Fatal("expected allowed for physician + Condition")
	}
	if !decision.RequireConsent {
		t.Error("expected RequireConsent=true for Condition")
	}

	// Patient does NOT require consent
	decision = engine.Evaluate(ctx, "Patient")
	if !decision.Allowed {
		t.Fatal("expected allowed for physician + Patient")
	}
	if decision.RequireConsent {
		t.Error("expected RequireConsent=false for Patient")
	}

	// Encounter does NOT require consent
	decision = engine.Evaluate(ctx, "Encounter")
	if !decision.Allowed {
		t.Fatal("expected allowed for physician + Encounter")
	}
	if decision.RequireConsent {
		t.Error("expected RequireConsent=false for Encounter")
	}
}

func TestABACEngine_AdminBypassDoesNotSetRequireConsent(t *testing.T) {
	engine := NewABACEngine(DefaultPolicies())

	ctx := context.WithValue(context.Background(), UserRolesKey, []string{"admin"})
	decision := engine.Evaluate(ctx, "Condition")

	if !decision.Allowed {
		t.Fatal("expected admin allowed")
	}
	if decision.RequireConsent {
		t.Error("admin bypass should not set RequireConsent")
	}
}

// ---------------------------------------------------------------------------
// ABACMiddleware sets context flags
// ---------------------------------------------------------------------------

func TestABACMiddleware_SetsRequireConsentFlag(t *testing.T) {
	engine := NewABACEngine(DefaultPolicies())

	e := echo.New()
	patientID := uuid.New()
	req := httptest.NewRequest(http.MethodGet, "/fhir/Condition?patient="+patientID.String(), nil)
	ctx := context.WithValue(req.Context(), UserRolesKey, []string{"physician"})
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/fhir/Condition")

	var gotConsent interface{}
	handler := func(c echo.Context) error {
		gotConsent = c.Get("require_consent")
		return c.String(http.StatusOK, "ok")
	}

	mw := ABACMiddleware(engine)
	h := mw(handler)
	err := h(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotConsent != true {
		t.Errorf("expected require_consent=true, got %v", gotConsent)
	}
}

func TestABACMiddleware_DoesNotSetFlagForPatient(t *testing.T) {
	engine := NewABACEngine(DefaultPolicies())

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient/123", nil)
	ctx := context.WithValue(req.Context(), UserRolesKey, []string{"physician"})
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/fhir/Patient/:id")

	var gotConsent interface{}
	handler := func(c echo.Context) error {
		gotConsent = c.Get("require_consent")
		return c.String(http.StatusOK, "ok")
	}

	mw := ABACMiddleware(engine)
	h := mw(handler)
	err := h(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotConsent != nil {
		t.Errorf("expected require_consent to be nil for Patient, got %v", gotConsent)
	}
}

// ---------------------------------------------------------------------------
// Consent enforcement middleware helpers
// ---------------------------------------------------------------------------

// mockConsentChecker implements ConsentChecker for tests.
type mockConsentChecker struct {
	consents []*ConsentInfo
	err      error
}

func (m *mockConsentChecker) ListActiveConsentsForPatient(_ context.Context, _ uuid.UUID) ([]*ConsentInfo, error) {
	return m.consents, m.err
}

// newConsentTestContext creates an echo.Context with path, method, roles, and
// optionally sets the "require_consent" flag (simulating what ABACMiddleware would do).
func newConsentTestContext(method, path string, roles []string, requireConsent bool) (echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	req := httptest.NewRequest(method, path, nil)
	if len(roles) > 0 {
		ctx := context.WithValue(req.Context(), UserRolesKey, roles)
		req = req.WithContext(ctx)
	}
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	// Set the path template so extractABACResourceType works.
	c.SetPath(path)
	if requireConsent {
		c.Set("require_consent", true)
	}
	return c, rec
}

// ---------------------------------------------------------------------------
// ConsentEnforcementMiddleware tests
// ---------------------------------------------------------------------------

func TestConsentEnforcementMiddleware_NilChecker_PassThrough(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/fhir/Condition", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("require_consent", true)

	called := false
	handler := func(c echo.Context) error {
		called = true
		return c.String(http.StatusOK, "ok")
	}

	mw := ConsentEnforcementMiddleware(nil)
	h := mw(handler)
	err := h(c)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !called {
		t.Error("expected handler to be called when checker is nil")
	}
}

func TestConsentEnforcementMiddleware_NoConsentRequired_PassThrough(t *testing.T) {
	checker := &mockConsentChecker{}

	c, _ := newConsentTestContext(http.MethodGet, "/fhir/Patient", []string{"physician"}, false)

	called := false
	handler := func(c echo.Context) error {
		called = true
		return c.String(http.StatusOK, "ok")
	}

	mw := ConsentEnforcementMiddleware(checker)
	h := mw(handler)
	err := h(c)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("expected handler to be called when consent is not required")
	}
}

func TestConsentEnforcementMiddleware_ActiveConsentPermit_PassesThrough(t *testing.T) {
	patientID := uuid.New()
	checker := &mockConsentChecker{
		consents: []*ConsentInfo{
			{
				Status:          "active",
				ProvisionType:   "permit",
				ProvisionAction: "access",
			},
		},
	}

	c, rec := newConsentTestContext(
		http.MethodGet,
		"/fhir/Condition?patient="+patientID.String(),
		[]string{"physician"},
		true,
	)

	called := false
	handler := func(c echo.Context) error {
		called = true
		return c.String(http.StatusOK, "ok")
	}

	mw := ConsentEnforcementMiddleware(checker)
	h := mw(handler)
	err := h(c)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("expected handler to be called with active permit consent")
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestConsentEnforcementMiddleware_NoConsentsExist_Returns403(t *testing.T) {
	patientID := uuid.New()
	checker := &mockConsentChecker{
		consents: []*ConsentInfo{}, // no consents at all
	}

	c, rec := newConsentTestContext(
		http.MethodGet,
		"/fhir/Observation?patient="+patientID.String(),
		[]string{"physician"},
		true,
	)

	handler := func(c echo.Context) error {
		t.Error("handler should not be called")
		return nil
	}

	mw := ConsentEnforcementMiddleware(checker)
	h := mw(handler)
	err := h(c)

	if err != nil {
		t.Fatalf("unexpected error (should use c.JSON for 403): %v", err)
	}
	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rec.Code)
	}

	// Verify OperationOutcome body
	var outcome map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &outcome); err != nil {
		t.Fatalf("failed to parse response body: %v", err)
	}
	if outcome["resourceType"] != "OperationOutcome" {
		t.Error("expected OperationOutcome resourceType")
	}
}

func TestConsentEnforcementMiddleware_ExpiredConsent_Returns403(t *testing.T) {
	patientID := uuid.New()
	pastEnd := time.Now().Add(-24 * time.Hour)
	pastStart := time.Now().Add(-48 * time.Hour)
	checker := &mockConsentChecker{
		consents: []*ConsentInfo{
			{
				Status:          "active",
				ProvisionType:   "permit",
				ProvisionAction: "access",
				ProvisionStart:  &pastStart,
				ProvisionEnd:    &pastEnd, // ended yesterday
			},
		},
	}

	c, rec := newConsentTestContext(
		http.MethodGet,
		"/fhir/Condition?patient="+patientID.String(),
		[]string{"physician"},
		true,
	)

	handler := func(c echo.Context) error {
		t.Error("handler should not be called")
		return nil
	}

	mw := ConsentEnforcementMiddleware(checker)
	h := mw(handler)
	err := h(c)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403 for expired consent, got %d", rec.Code)
	}
}

func TestConsentEnforcementMiddleware_FutureConsent_Returns403(t *testing.T) {
	patientID := uuid.New()
	futureStart := time.Now().Add(24 * time.Hour)
	futureEnd := time.Now().Add(48 * time.Hour)
	checker := &mockConsentChecker{
		consents: []*ConsentInfo{
			{
				Status:          "active",
				ProvisionType:   "permit",
				ProvisionAction: "access",
				ProvisionStart:  &futureStart, // starts tomorrow
				ProvisionEnd:    &futureEnd,
			},
		},
	}

	c, rec := newConsentTestContext(
		http.MethodGet,
		"/fhir/Condition?patient="+patientID.String(),
		[]string{"physician"},
		true,
	)

	handler := func(c echo.Context) error {
		t.Error("handler should not be called")
		return nil
	}

	mw := ConsentEnforcementMiddleware(checker)
	h := mw(handler)
	err := h(c)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403 for future consent, got %d", rec.Code)
	}
}

func TestConsentEnforcementMiddleware_DenyProvision_Returns403(t *testing.T) {
	patientID := uuid.New()
	checker := &mockConsentChecker{
		consents: []*ConsentInfo{
			{
				Status:          "active",
				ProvisionType:   "deny",
				ProvisionAction: "access",
			},
		},
	}

	c, rec := newConsentTestContext(
		http.MethodGet,
		"/fhir/Condition?patient="+patientID.String(),
		[]string{"physician"},
		true,
	)

	handler := func(c echo.Context) error {
		t.Error("handler should not be called")
		return nil
	}

	mw := ConsentEnforcementMiddleware(checker)
	h := mw(handler)
	err := h(c)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403 for deny consent, got %d", rec.Code)
	}

	var outcome map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &outcome); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	issues := outcome["issue"].([]interface{})
	firstIssue := issues[0].(map[string]interface{})
	if firstIssue["diagnostics"] != "access denied by patient consent directive" {
		t.Errorf("unexpected diagnostics: %v", firstIssue["diagnostics"])
	}
}

func TestConsentEnforcementMiddleware_DenyTakesPrecedenceOverPermit(t *testing.T) {
	patientID := uuid.New()
	checker := &mockConsentChecker{
		consents: []*ConsentInfo{
			{
				Status:          "active",
				ProvisionType:   "permit",
				ProvisionAction: "access",
			},
			{
				Status:          "active",
				ProvisionType:   "deny",
				ProvisionAction: "access",
			},
		},
	}

	c, rec := newConsentTestContext(
		http.MethodGet,
		"/fhir/Condition?patient="+patientID.String(),
		[]string{"physician"},
		true,
	)

	handler := func(c echo.Context) error {
		t.Error("handler should not be called")
		return nil
	}

	mw := ConsentEnforcementMiddleware(checker)
	h := mw(handler)
	err := h(c)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403 when deny takes precedence, got %d", rec.Code)
	}
}

func TestConsentEnforcementMiddleware_AdminBypass(t *testing.T) {
	checker := &mockConsentChecker{
		consents: []*ConsentInfo{}, // no consents - would normally deny
	}

	patientID := uuid.New()
	c, _ := newConsentTestContext(
		http.MethodGet,
		"/fhir/Condition?patient="+patientID.String(),
		[]string{"admin"},
		true,
	)

	called := false
	handler := func(c echo.Context) error {
		called = true
		return c.String(http.StatusOK, "ok")
	}

	mw := ConsentEnforcementMiddleware(checker)
	h := mw(handler)
	err := h(c)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("expected handler to be called for admin bypass")
	}
}

func TestConsentEnforcementMiddleware_InactiveConsentIgnored(t *testing.T) {
	patientID := uuid.New()
	checker := &mockConsentChecker{
		consents: []*ConsentInfo{
			{
				Status:          "inactive",
				ProvisionType:   "permit",
				ProvisionAction: "access",
			},
		},
	}

	c, rec := newConsentTestContext(
		http.MethodGet,
		"/fhir/Condition?patient="+patientID.String(),
		[]string{"physician"},
		true,
	)

	handler := func(c echo.Context) error {
		t.Error("handler should not be called")
		return nil
	}

	mw := ConsentEnforcementMiddleware(checker)
	h := mw(handler)
	err := h(c)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403 for inactive consent, got %d", rec.Code)
	}
}

func TestConsentEnforcementMiddleware_WriteAction_RequiresCorrectProvisionAction(t *testing.T) {
	patientID := uuid.New()

	// Consent only permits "access" (read), not "correct" (write).
	checker := &mockConsentChecker{
		consents: []*ConsentInfo{
			{
				Status:          "active",
				ProvisionType:   "permit",
				ProvisionAction: "access", // read only
			},
		},
	}

	c, rec := newConsentTestContext(
		http.MethodPut,
		"/fhir/Condition?patient="+patientID.String(),
		[]string{"physician"},
		true,
	)

	handler := func(c echo.Context) error {
		t.Error("handler should not be called")
		return nil
	}

	mw := ConsentEnforcementMiddleware(checker)
	h := mw(handler)
	err := h(c)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403 for PUT with access-only consent, got %d", rec.Code)
	}
}

func TestConsentEnforcementMiddleware_EmptyProvisionAction_MatchesAny(t *testing.T) {
	patientID := uuid.New()

	// Consent has no provision action set - should match any action.
	checker := &mockConsentChecker{
		consents: []*ConsentInfo{
			{
				Status:          "active",
				ProvisionType:   "permit",
				ProvisionAction: "", // matches all actions
			},
		},
	}

	c, _ := newConsentTestContext(
		http.MethodPut,
		"/fhir/Condition?patient="+patientID.String(),
		[]string{"physician"},
		true,
	)

	called := false
	handler := func(c echo.Context) error {
		called = true
		return c.String(http.StatusOK, "ok")
	}

	mw := ConsentEnforcementMiddleware(checker)
	h := mw(handler)
	err := h(c)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("expected handler to be called when provision action is empty (match-all)")
	}
}

func TestConsentEnforcementMiddleware_PatientIDFromQuerySubject(t *testing.T) {
	patientID := uuid.New()
	checker := &mockConsentChecker{
		consents: []*ConsentInfo{
			{Status: "active", ProvisionType: "permit"},
		},
	}

	// Use "subject" query param with FHIR reference prefix
	c, _ := newConsentTestContext(
		http.MethodGet,
		"/fhir/Observation?subject=Patient/"+patientID.String(),
		[]string{"physician"},
		true,
	)

	called := false
	handler := func(c echo.Context) error {
		called = true
		return c.String(http.StatusOK, "ok")
	}

	mw := ConsentEnforcementMiddleware(checker)
	h := mw(handler)
	err := h(c)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("expected handler to be called with patient ID from subject query param")
	}
}

func TestConsentEnforcementMiddleware_NoPatientID_Returns403(t *testing.T) {
	checker := &mockConsentChecker{
		consents: []*ConsentInfo{
			{Status: "active", ProvisionType: "permit"},
		},
	}

	// No patient ID in path or query - consent required but cannot determine patient
	c, rec := newConsentTestContext(
		http.MethodGet,
		"/fhir/Condition",
		[]string{"physician"},
		true,
	)

	handler := func(c echo.Context) error {
		t.Error("handler should not be called")
		return nil
	}

	mw := ConsentEnforcementMiddleware(checker)
	h := mw(handler)
	err := h(c)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403 when patient ID cannot be determined, got %d", rec.Code)
	}
}

func TestConsentEnforcementMiddleware_CheckerError_Returns500(t *testing.T) {
	patientID := uuid.New()
	checker := &mockConsentChecker{
		err: fmt.Errorf("database connection failed"),
	}

	c, rec := newConsentTestContext(
		http.MethodGet,
		"/fhir/Condition?patient="+patientID.String(),
		[]string{"physician"},
		true,
	)

	handler := func(c echo.Context) error {
		t.Error("handler should not be called")
		return nil
	}

	mw := ConsentEnforcementMiddleware(checker)
	h := mw(handler)
	err := h(c)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rec.Code)
	}
}

// ---------------------------------------------------------------------------
// httpMethodToFHIRAction
// ---------------------------------------------------------------------------

func TestHttpMethodToFHIRAction(t *testing.T) {
	tests := []struct {
		method string
		want   string
	}{
		{http.MethodGet, "access"},
		{http.MethodHead, "access"},
		{http.MethodPost, "access"},
		{http.MethodPut, "correct"},
		{http.MethodPatch, "correct"},
		{http.MethodDelete, "correct"},
		{"OPTIONS", "access"},
	}
	for _, tt := range tests {
		got := httpMethodToFHIRAction(tt.method)
		if got != tt.want {
			t.Errorf("httpMethodToFHIRAction(%q) = %q, want %q", tt.method, got, tt.want)
		}
	}
}

// ---------------------------------------------------------------------------
// extractPatientID
// ---------------------------------------------------------------------------

func TestExtractPatientID(t *testing.T) {
	patientID := uuid.New()

	t.Run("from query param patient", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/fhir/Condition?patient="+patientID.String(), nil)
		c := e.NewContext(req, httptest.NewRecorder())
		c.SetPath("/fhir/Condition")

		got, ok := extractPatientID(c)
		if !ok {
			t.Fatal("expected patient ID extraction to succeed")
		}
		if got != patientID {
			t.Errorf("got %v, want %v", got, patientID)
		}
	})

	t.Run("from query param subject with prefix", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/fhir/Observation?subject=Patient/"+patientID.String(), nil)
		c := e.NewContext(req, httptest.NewRecorder())
		c.SetPath("/fhir/Observation")

		got, ok := extractPatientID(c)
		if !ok {
			t.Fatal("expected patient ID extraction to succeed")
		}
		if got != patientID {
			t.Errorf("got %v, want %v", got, patientID)
		}
	})

	t.Run("no patient ID available", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/fhir/Condition", nil)
		c := e.NewContext(req, httptest.NewRecorder())
		c.SetPath("/fhir/Condition")

		_, ok := extractPatientID(c)
		if ok {
			t.Error("expected patient ID extraction to fail")
		}
	})

	t.Run("invalid UUID in query param", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/fhir/Condition?patient=not-a-uuid", nil)
		c := e.NewContext(req, httptest.NewRecorder())
		c.SetPath("/fhir/Condition")

		_, ok := extractPatientID(c)
		if ok {
			t.Error("expected patient ID extraction to fail for invalid UUID")
		}
	})
}

// ---------------------------------------------------------------------------
// consentOperationOutcome
// ---------------------------------------------------------------------------

func TestConsentOperationOutcome(t *testing.T) {
	outcome := consentOperationOutcome("test message")
	if outcome["resourceType"] != "OperationOutcome" {
		t.Error("expected resourceType OperationOutcome")
	}
	issues, ok := outcome["issue"].([]map[string]interface{})
	if !ok || len(issues) != 1 {
		t.Fatal("expected exactly one issue")
	}
	if issues[0]["severity"] != "error" {
		t.Error("expected severity error")
	}
	if issues[0]["code"] != "forbidden" {
		t.Error("expected code forbidden")
	}
	if issues[0]["diagnostics"] != "test message" {
		t.Errorf("expected diagnostics 'test message', got %v", issues[0]["diagnostics"])
	}
}
