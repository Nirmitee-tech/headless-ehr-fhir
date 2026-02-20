package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sort"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

// ctxWithRoles creates a context.Context carrying the given roles.
func ctxWithRoles(roles ...string) context.Context {
	return context.WithValue(context.Background(), UserRolesKey, roles)
}

// allPolicyResourceTypes returns a sorted slice of every ResourceType covered
// by DefaultPolicies().
func allPolicyResourceTypes() []string {
	seen := map[string]bool{}
	for _, p := range DefaultPolicies() {
		seen[p.ResourceType] = true
	}
	out := make([]string, 0, len(seen))
	for rt := range seen {
		out = append(out, rt)
	}
	sort.Strings(out)
	return out
}

// policyForResource returns the first policy matching resourceType or nil.
func policyForResource(resourceType string) *ABACPolicy {
	for _, p := range DefaultPolicies() {
		if p.ResourceType == resourceType {
			cp := p
			return &cp
		}
	}
	return nil
}

// ---------------------------------------------------------------------------
// TestABAC_DefaultPoliciesCoverage
// ---------------------------------------------------------------------------

// TestABAC_DefaultPoliciesCoverage verifies that every expected resource type
// has an explicit policy entry.
func TestABAC_DefaultPoliciesCoverage(t *testing.T) {
	expectedResources := []string{
		// Clinical PHI
		"Condition", "Observation", "AllergyIntolerance", "Procedure", "NutritionOrder",
		"DiagnosticReport", "ServiceRequest", "ImagingStudy", "Specimen",
		"MedicationRequest", "MedicationAdministration", "MedicationDispense", "MedicationStatement",
		"DocumentReference", "Composition",
		"FamilyMemberHistory", "ClinicalImpression", "RiskAssessment",
		"Flag", "DetectedIssue", "AdverseEvent",

		// Patient-context
		"Patient", "Encounter", "EpisodeOfCare", "Consent", "Communication",
		"CarePlan", "Goal", "CareTeam", "Task",
		"RelatedPerson",
		"Immunization", "ImmunizationRecommendation", "ImmunizationEvaluation",
		"Coverage", "Claim", "ClaimResponse", "ExplanationOfBenefit",
		"QuestionnaireResponse",
		"Device", "DeviceRequest", "DeviceUseStatement",
		"MolecularSequence", "BodyStructure", "Media",

		// Administrative
		"Practitioner", "PractitionerRole", "Organization", "Location",
		"HealthcareService", "Endpoint", "Group",
		"Questionnaire",
		"ResearchStudy", "ResearchSubject",
		"Provenance",
		"Subscription",
		"Account", "InsurancePlan", "Invoice",
		"StructureDefinition", "SearchParameter", "CodeSystem", "ValueSet",
		"ConceptMap", "NamingSystem", "CapabilityStatement",
		"OperationDefinition", "CompartmentDefinition", "ImplementationGuide",
		"Medication", "MedicationKnowledge", "Substance",
		"Schedule", "Slot", "Appointment", "AppointmentResponse",
	}

	policyMap := map[string]bool{}
	for _, p := range DefaultPolicies() {
		policyMap[p.ResourceType] = true
	}

	for _, rt := range expectedResources {
		if !policyMap[rt] {
			t.Errorf("missing policy for expected resource type %q", rt)
		}
	}

	// Also confirm no duplicate resource types.
	seen := map[string]int{}
	for _, p := range DefaultPolicies() {
		seen[p.ResourceType]++
	}
	for rt, count := range seen {
		if count > 1 {
			t.Errorf("duplicate policy for resource type %q (appeared %d times)", rt, count)
		}
	}
}

// ---------------------------------------------------------------------------
// TestABAC_ConsentRequirements
// ---------------------------------------------------------------------------

// TestABAC_ConsentRequirements verifies that exactly the clinical PHI
// resources have RequireConsent: true and all others have false.
func TestABAC_ConsentRequirements(t *testing.T) {
	phiResources := map[string]bool{
		"Condition": true, "Observation": true, "AllergyIntolerance": true,
		"Procedure": true, "NutritionOrder": true,
		"DiagnosticReport": true, "ServiceRequest": true, "ImagingStudy": true, "Specimen": true,
		"MedicationRequest": true, "MedicationAdministration": true,
		"MedicationDispense": true, "MedicationStatement": true,
		"DocumentReference": true, "Composition": true,
		"FamilyMemberHistory": true, "ClinicalImpression": true, "RiskAssessment": true,
		"Flag": true, "DetectedIssue": true, "AdverseEvent": true,
	}

	for _, p := range DefaultPolicies() {
		expectConsent := phiResources[p.ResourceType]
		if p.RequireConsent != expectConsent {
			t.Errorf("resource %q: RequireConsent=%v, want %v",
				p.ResourceType, p.RequireConsent, expectConsent)
		}
	}
}

// ---------------------------------------------------------------------------
// TestABAC_RoleAccessByCategory
// ---------------------------------------------------------------------------

func TestABAC_RoleAccessByCategory(t *testing.T) {
	engine := NewABACEngine(DefaultPolicies())

	type testCase struct {
		resource string
		role     string
		allowed  bool
	}

	tests := []testCase{
		// --- Clinical PHI: general (admin, physician, nurse) ---
		{"Condition", "physician", true},
		{"Condition", "nurse", true},
		{"Condition", "receptionist", false},
		{"Observation", "nurse", true},
		{"AllergyIntolerance", "nurse", true},
		{"AllergyIntolerance", "pharmacist", false},
		{"Procedure", "physician", true},
		{"NutritionOrder", "nurse", true},

		// --- Clinical PHI: diagnostics (admin, physician, nurse, lab_tech, radiologist) ---
		{"DiagnosticReport", "lab_tech", true},
		{"DiagnosticReport", "radiologist", true},
		{"DiagnosticReport", "pharmacist", false},
		{"ServiceRequest", "nurse", true},
		{"ImagingStudy", "radiologist", true},
		{"Specimen", "lab_tech", true},

		// --- Clinical PHI: medications (admin, physician, pharmacist) ---
		{"MedicationRequest", "physician", true},
		{"MedicationRequest", "pharmacist", true},
		{"MedicationRequest", "nurse", false},
		{"MedicationAdministration", "pharmacist", true},
		{"MedicationDispense", "pharmacist", true},
		{"MedicationStatement", "physician", true},
		{"MedicationStatement", "nurse", false},

		// --- Clinical PHI: documents (admin, physician, nurse) ---
		{"DocumentReference", "nurse", true},
		{"Composition", "physician", true},

		// --- Clinical PHI: assessments (admin, physician, nurse) ---
		{"FamilyMemberHistory", "nurse", true},
		{"ClinicalImpression", "physician", true},
		{"RiskAssessment", "nurse", true},

		// --- Clinical PHI: safety (admin, physician, nurse) ---
		{"Flag", "nurse", true},
		{"DetectedIssue", "physician", true},
		{"AdverseEvent", "nurse", true},

		// --- Patient-context ---
		{"Patient", "physician", true},
		{"Patient", "nurse", true},
		{"Patient", "registrar", true},
		{"Patient", "receptionist", true},
		{"Patient", "pharmacist", false},
		{"Encounter", "nurse", true},
		{"EpisodeOfCare", "nurse", true},
		{"Consent", "nurse", true},
		{"Communication", "physician", true},
		{"CarePlan", "nurse", true},
		{"Goal", "physician", true},
		{"CareTeam", "nurse", true},
		{"Task", "nurse", true},
		{"RelatedPerson", "registrar", true},
		{"RelatedPerson", "pharmacist", false},
		{"Immunization", "nurse", true},
		{"ImmunizationRecommendation", "physician", true},
		{"ImmunizationEvaluation", "nurse", true},
		{"Coverage", "billing", true},
		{"Coverage", "physician", false},
		{"Claim", "billing", true},
		{"ClaimResponse", "billing", true},
		{"ExplanationOfBenefit", "billing", true},
		{"QuestionnaireResponse", "patient", true},
		{"QuestionnaireResponse", "nurse", true},
		{"QuestionnaireResponse", "billing", false},
		{"Device", "nurse", true},
		{"DeviceRequest", "physician", true},
		{"DeviceUseStatement", "nurse", true},
		{"MolecularSequence", "lab_tech", true},
		{"MolecularSequence", "nurse", false},
		{"BodyStructure", "nurse", true},
		{"Media", "nurse", true},

		// --- Administrative ---
		{"Practitioner", "registrar", true},
		{"Practitioner", "pharmacist", true},
		{"Practitioner", "lab_tech", true},
		{"PractitionerRole", "nurse", true},
		{"Organization", "registrar", true},
		{"Location", "registrar", true},
		{"HealthcareService", "nurse", true},
		{"HealthcareService", "billing", false},
		{"Endpoint", "physician", true},
		{"Endpoint", "nurse", false},
		{"Group", "nurse", true},
		{"Questionnaire", "nurse", true},
		{"ResearchStudy", "physician", true},
		{"ResearchStudy", "nurse", false},
		{"ResearchSubject", "physician", true},
		{"Provenance", "nurse", true},
		{"Subscription", "physician", false},
		{"Subscription", "nurse", false},
		{"Account", "billing", true},
		{"Account", "nurse", false},
		{"InsurancePlan", "billing", true},
		{"Invoice", "billing", true},
		{"StructureDefinition", "physician", true},
		{"StructureDefinition", "nurse", false},
		{"CodeSystem", "physician", true},
		{"ValueSet", "physician", true},
		{"ConceptMap", "physician", true},
		{"Medication", "pharmacist", true},
		{"MedicationKnowledge", "pharmacist", true},
		{"Substance", "pharmacist", true},
		{"Schedule", "registrar", true},
		{"Slot", "registrar", true},
		{"Appointment", "registrar", true},
		{"AppointmentResponse", "registrar", true},
		{"Appointment", "billing", false},
	}

	for _, tc := range tests {
		t.Run(fmt.Sprintf("%s/%s", tc.resource, tc.role), func(t *testing.T) {
			ctx := ctxWithRoles(tc.role)
			decision := engine.Evaluate(ctx, tc.resource)
			if decision.Allowed != tc.allowed {
				t.Errorf("Evaluate(%s, %s) Allowed=%v, want %v (reason: %s)",
					tc.role, tc.resource, decision.Allowed, tc.allowed, decision.Reason)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// TestABAC_AdminBypass
// ---------------------------------------------------------------------------

func TestABAC_AdminBypass(t *testing.T) {
	engine := NewABACEngine(DefaultPolicies())

	// Admin should be allowed on every resource, including unknown ones.
	resources := append(allPolicyResourceTypes(), "TotallyUnknownResource")
	for _, rt := range resources {
		ctx := ctxWithRoles("admin")
		decision := engine.Evaluate(ctx, rt)
		if !decision.Allowed {
			t.Errorf("admin should be allowed for %q but got denied (reason: %s)", rt, decision.Reason)
		}
		if decision.Reason != "admin role" {
			t.Errorf("admin reason for %q = %q, want %q", rt, decision.Reason, "admin role")
		}
		// Admin bypass should NOT set RequireConsent.
		if decision.RequireConsent {
			t.Errorf("admin bypass should not set RequireConsent for %q", rt)
		}
	}
}

// ---------------------------------------------------------------------------
// TestABAC_UnknownResource
// ---------------------------------------------------------------------------

func TestABAC_UnknownResourcePhysicianAllowed(t *testing.T) {
	engine := NewABACEngine(DefaultPolicies())

	ctx := ctxWithRoles("physician")
	decision := engine.Evaluate(ctx, "BrandNewResource")

	if !decision.Allowed {
		t.Error("expected physician to be allowed for unlisted resource via default policy")
	}
	if decision.Reason != "default policy for unlisted resource" {
		t.Errorf("unexpected reason: %q", decision.Reason)
	}
}

func TestABAC_UnknownResourceNurseDenied(t *testing.T) {
	engine := NewABACEngine(DefaultPolicies())

	ctx := ctxWithRoles("nurse")
	decision := engine.Evaluate(ctx, "BrandNewResource")

	if decision.Allowed {
		t.Error("expected nurse to be denied for unlisted resource at Evaluate level")
	}
}

func TestABAC_UnknownResourceNoRoles(t *testing.T) {
	engine := NewABACEngine(DefaultPolicies())

	ctx := context.Background()
	decision := engine.Evaluate(ctx, "UnknownResource")

	if decision.Allowed {
		t.Error("expected denial when no roles are present")
	}
}

// ---------------------------------------------------------------------------
// TestABAC_EmptyPolicies
// ---------------------------------------------------------------------------

func TestABAC_EmptyPolicies(t *testing.T) {
	engine := NewABACEngine([]ABACPolicy{})

	ctx := ctxWithRoles("physician")
	decision := engine.Evaluate(ctx, "Patient")

	// With empty policies physician still hits the "default policy for unlisted
	// resource" fallback.
	if !decision.Allowed {
		t.Error("expected physician to be allowed via default fallback even with empty policies")
	}
}

func TestABAC_EmptyPoliciesNurseDenied(t *testing.T) {
	engine := NewABACEngine([]ABACPolicy{})

	ctx := ctxWithRoles("nurse")
	decision := engine.Evaluate(ctx, "Patient")

	if decision.Allowed {
		t.Error("expected nurse to be denied with empty policies (no explicit Patient policy)")
	}
}

// ---------------------------------------------------------------------------
// TestABAC_EvaluateReturnsConsentFlag
// ---------------------------------------------------------------------------

func TestABAC_EvaluateReturnsConsentFlag(t *testing.T) {
	engine := NewABACEngine(DefaultPolicies())

	ctx := ctxWithRoles("physician")

	// PHI resource should set RequireConsent.
	decision := engine.Evaluate(ctx, "Condition")
	if !decision.Allowed {
		t.Fatal("expected allowed for physician + Condition")
	}
	if !decision.RequireConsent {
		t.Error("expected RequireConsent=true for Condition")
	}

	// Non-PHI resource should not set RequireConsent.
	decision = engine.Evaluate(ctx, "Patient")
	if !decision.Allowed {
		t.Fatal("expected allowed for physician + Patient")
	}
	if decision.RequireConsent {
		t.Error("expected RequireConsent=false for Patient")
	}

	// Encounter is not PHI.
	decision = engine.Evaluate(ctx, "Encounter")
	if !decision.Allowed {
		t.Fatal("expected allowed for physician + Encounter")
	}
	if decision.RequireConsent {
		t.Error("expected RequireConsent=false for Encounter")
	}
}

// ---------------------------------------------------------------------------
// TestABAC_isClinicalRole
// ---------------------------------------------------------------------------

func TestABAC_IsClinicalRole(t *testing.T) {
	if !isClinicalRole("physician") {
		t.Error("physician should be clinical")
	}
	if !isClinicalRole("nurse") {
		t.Error("nurse should be clinical")
	}
	if isClinicalRole("admin") {
		t.Error("admin should not be clinical")
	}
	if isClinicalRole("receptionist") {
		t.Error("receptionist should not be clinical")
	}
	if isClinicalRole("billing") {
		t.Error("billing should not be clinical")
	}
}

// ---------------------------------------------------------------------------
// TestABAC_IsReadOnly
// ---------------------------------------------------------------------------

func TestABAC_IsReadOnly(t *testing.T) {
	if !isReadOnly(http.MethodGet) {
		t.Error("GET should be read-only")
	}
	if !isReadOnly(http.MethodHead) {
		t.Error("HEAD should be read-only")
	}
	if isReadOnly(http.MethodPost) {
		t.Error("POST should not be read-only")
	}
	if isReadOnly(http.MethodPut) {
		t.Error("PUT should not be read-only")
	}
	if isReadOnly(http.MethodPatch) {
		t.Error("PATCH should not be read-only")
	}
	if isReadOnly(http.MethodDelete) {
		t.Error("DELETE should not be read-only")
	}
}

// ---------------------------------------------------------------------------
// TestABAC_HasExplicitPolicy
// ---------------------------------------------------------------------------

func TestABAC_HasExplicitPolicy(t *testing.T) {
	engine := NewABACEngine(DefaultPolicies())

	if !engine.hasExplicitPolicy("Patient") {
		t.Error("expected explicit policy for Patient")
	}
	if !engine.hasExplicitPolicy("Condition") {
		t.Error("expected explicit policy for Condition")
	}
	if engine.hasExplicitPolicy("TotallyFakeResource") {
		t.Error("expected no explicit policy for TotallyFakeResource")
	}
}

// ---------------------------------------------------------------------------
// extractABACResourceType
// ---------------------------------------------------------------------------

func TestABAC_ExtractResourceType(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{"/fhir/Patient", "Patient"},
		{"/fhir/Patient/123", "Patient"},
		{"/fhir/Observation", "Observation"},
		{"/fhir/Condition/abc-123", "Condition"},
		{"/fhir/AllergyIntolerance/xyz", "AllergyIntolerance"},
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

// ---------------------------------------------------------------------------
// ABACMiddleware tests
// ---------------------------------------------------------------------------

func TestABAC_Middleware_Allowed(t *testing.T) {
	engine := NewABACEngine(DefaultPolicies())

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient/123", nil)
	req = req.WithContext(ctxWithRoles("physician"))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/fhir/Patient/:id")

	handler := func(c echo.Context) error { return c.String(http.StatusOK, "ok") }
	mw := ABACMiddleware(engine)
	err := mw(handler)(c)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestABAC_Middleware_Denied(t *testing.T) {
	engine := NewABACEngine(DefaultPolicies())

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/fhir/MedicationRequest/123", nil)
	req = req.WithContext(ctxWithRoles("nurse"))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/fhir/MedicationRequest/:id")

	handler := func(c echo.Context) error { return c.String(http.StatusOK, "ok") }
	mw := ABACMiddleware(engine)
	err := mw(handler)(c)

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

func TestABAC_Middleware_NonFHIRPath(t *testing.T) {
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
	err := mw(handler)(c)

	if err != nil {
		t.Fatalf("expected no error for non-FHIR path, got %v", err)
	}
	if !called {
		t.Error("expected handler to be called for non-FHIR path")
	}
}

func TestABAC_Middleware_SetsConsentFlag(t *testing.T) {
	engine := NewABACEngine(DefaultPolicies())

	e := echo.New()
	patientID := uuid.New()
	req := httptest.NewRequest(http.MethodGet, "/fhir/Condition?patient="+patientID.String(), nil)
	req = req.WithContext(ctxWithRoles("physician"))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/fhir/Condition")

	var gotConsent interface{}
	handler := func(c echo.Context) error {
		gotConsent = c.Get("require_consent")
		return c.String(http.StatusOK, "ok")
	}
	mw := ABACMiddleware(engine)
	err := mw(handler)(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotConsent != true {
		t.Errorf("expected require_consent=true, got %v", gotConsent)
	}
}

func TestABAC_Middleware_DoesNotSetConsentForPatient(t *testing.T) {
	engine := NewABACEngine(DefaultPolicies())

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient/123", nil)
	req = req.WithContext(ctxWithRoles("physician"))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/fhir/Patient/:id")

	var gotConsent interface{}
	handler := func(c echo.Context) error {
		gotConsent = c.Get("require_consent")
		return c.String(http.StatusOK, "ok")
	}
	mw := ABACMiddleware(engine)
	err := mw(handler)(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotConsent != nil {
		t.Errorf("expected require_consent to be nil for Patient, got %v", gotConsent)
	}
}

// ---------------------------------------------------------------------------
// Middleware: unknown resource graceful handling
// ---------------------------------------------------------------------------

func TestABAC_Middleware_UnknownResource_NurseGETAllowed(t *testing.T) {
	engine := NewABACEngine(DefaultPolicies())

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/fhir/BrandNewResource/123", nil)
	req = req.WithContext(ctxWithRoles("nurse"))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/fhir/BrandNewResource/:id")

	called := false
	handler := func(c echo.Context) error {
		called = true
		return c.String(http.StatusOK, "ok")
	}
	mw := ABACMiddleware(engine)
	err := mw(handler)(c)

	if err != nil {
		t.Fatalf("expected nurse GET on unknown resource to pass, got %v", err)
	}
	if !called {
		t.Error("expected handler to be called (nurse GET on unknown resource)")
	}
}

func TestABAC_Middleware_UnknownResource_NurseHEADAllowed(t *testing.T) {
	engine := NewABACEngine(DefaultPolicies())

	e := echo.New()
	req := httptest.NewRequest(http.MethodHead, "/fhir/BrandNewResource/123", nil)
	req = req.WithContext(ctxWithRoles("nurse"))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/fhir/BrandNewResource/:id")

	called := false
	handler := func(c echo.Context) error {
		called = true
		return c.String(http.StatusOK, "ok")
	}
	mw := ABACMiddleware(engine)
	err := mw(handler)(c)

	if err != nil {
		t.Fatalf("expected nurse HEAD on unknown resource to pass, got %v", err)
	}
	if !called {
		t.Error("expected handler to be called (nurse HEAD on unknown resource)")
	}
}

func TestABAC_Middleware_UnknownResource_NursePOSTDenied(t *testing.T) {
	engine := NewABACEngine(DefaultPolicies())

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/fhir/BrandNewResource", nil)
	req = req.WithContext(ctxWithRoles("nurse"))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/fhir/BrandNewResource")

	handler := func(c echo.Context) error { return c.String(http.StatusOK, "ok") }
	mw := ABACMiddleware(engine)
	err := mw(handler)(c)

	if err == nil {
		t.Fatal("expected nurse POST on unknown resource to be denied")
	}
	httpErr, ok := err.(*echo.HTTPError)
	if !ok {
		t.Fatalf("expected echo.HTTPError, got %T", err)
	}
	if httpErr.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", httpErr.Code)
	}
}

func TestABAC_Middleware_UnknownResource_NursePUTDenied(t *testing.T) {
	engine := NewABACEngine(DefaultPolicies())

	e := echo.New()
	req := httptest.NewRequest(http.MethodPut, "/fhir/BrandNewResource/123", nil)
	req = req.WithContext(ctxWithRoles("nurse"))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/fhir/BrandNewResource/:id")

	handler := func(c echo.Context) error { return c.String(http.StatusOK, "ok") }
	mw := ABACMiddleware(engine)
	err := mw(handler)(c)

	if err == nil {
		t.Fatal("expected nurse PUT on unknown resource to be denied")
	}
}

func TestABAC_Middleware_UnknownResource_NurseDELETEDenied(t *testing.T) {
	engine := NewABACEngine(DefaultPolicies())

	e := echo.New()
	req := httptest.NewRequest(http.MethodDelete, "/fhir/BrandNewResource/123", nil)
	req = req.WithContext(ctxWithRoles("nurse"))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/fhir/BrandNewResource/:id")

	handler := func(c echo.Context) error { return c.String(http.StatusOK, "ok") }
	mw := ABACMiddleware(engine)
	err := mw(handler)(c)

	if err == nil {
		t.Fatal("expected nurse DELETE on unknown resource to be denied")
	}
}

func TestABAC_Middleware_UnknownResource_PhysicianGETAllowed(t *testing.T) {
	engine := NewABACEngine(DefaultPolicies())

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/fhir/BrandNewResource/123", nil)
	req = req.WithContext(ctxWithRoles("physician"))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/fhir/BrandNewResource/:id")

	called := false
	handler := func(c echo.Context) error {
		called = true
		return c.String(http.StatusOK, "ok")
	}
	mw := ABACMiddleware(engine)
	err := mw(handler)(c)

	if err != nil {
		t.Fatalf("expected physician GET on unknown resource to pass, got %v", err)
	}
	if !called {
		t.Error("expected handler to be called")
	}
}

func TestABAC_Middleware_UnknownResource_PhysicianPOSTAllowed(t *testing.T) {
	// Physician is allowed via the Evaluate default policy fallback for all
	// methods, since Evaluate grants physician access on unlisted resources.
	engine := NewABACEngine(DefaultPolicies())

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/fhir/BrandNewResource", nil)
	req = req.WithContext(ctxWithRoles("physician"))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/fhir/BrandNewResource")

	called := false
	handler := func(c echo.Context) error {
		called = true
		return c.String(http.StatusOK, "ok")
	}
	mw := ABACMiddleware(engine)
	err := mw(handler)(c)

	if err != nil {
		t.Fatalf("expected physician POST on unknown resource to pass, got %v", err)
	}
	if !called {
		t.Error("expected handler to be called")
	}
}

func TestABAC_Middleware_UnknownResource_ReceptionistGETDenied(t *testing.T) {
	// receptionist is not a clinical role, so no fallback.
	engine := NewABACEngine(DefaultPolicies())

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/fhir/BrandNewResource/123", nil)
	req = req.WithContext(ctxWithRoles("receptionist"))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/fhir/BrandNewResource/:id")

	handler := func(c echo.Context) error { return c.String(http.StatusOK, "ok") }
	mw := ABACMiddleware(engine)
	err := mw(handler)(c)

	if err == nil {
		t.Fatal("expected receptionist GET on unknown resource to be denied")
	}
}

func TestABAC_Middleware_UnknownResource_AdminAllowed(t *testing.T) {
	engine := NewABACEngine(DefaultPolicies())

	e := echo.New()
	req := httptest.NewRequest(http.MethodDelete, "/fhir/BrandNewResource/123", nil)
	req = req.WithContext(ctxWithRoles("admin"))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/fhir/BrandNewResource/:id")

	called := false
	handler := func(c echo.Context) error {
		called = true
		return c.String(http.StatusOK, "ok")
	}
	mw := ABACMiddleware(engine)
	err := mw(handler)(c)

	if err != nil {
		t.Fatalf("expected admin to bypass unknown resource, got %v", err)
	}
	if !called {
		t.Error("expected handler to be called")
	}
}

// ---------------------------------------------------------------------------
// Middleware: various FHIR resource paths
// ---------------------------------------------------------------------------

func TestABAC_Middleware_VariousResourcePaths(t *testing.T) {
	engine := NewABACEngine(DefaultPolicies())

	tests := []struct {
		name    string
		method  string
		path    string
		tmpl    string
		roles   []string
		allowed bool
	}{
		{"pharmacist reads MedicationRequest", http.MethodGet, "/fhir/MedicationRequest/123", "/fhir/MedicationRequest/:id", []string{"pharmacist"}, true},
		{"pharmacist reads Practitioner", http.MethodGet, "/fhir/Practitioner/123", "/fhir/Practitioner/:id", []string{"pharmacist"}, true},
		{"lab_tech reads DiagnosticReport", http.MethodGet, "/fhir/DiagnosticReport/123", "/fhir/DiagnosticReport/:id", []string{"lab_tech"}, true},
		{"radiologist reads ImagingStudy", http.MethodGet, "/fhir/ImagingStudy/123", "/fhir/ImagingStudy/:id", []string{"radiologist"}, true},
		{"billing reads Coverage", http.MethodGet, "/fhir/Coverage/123", "/fhir/Coverage/:id", []string{"billing"}, true},
		{"billing denied Patient", http.MethodGet, "/fhir/Patient/123", "/fhir/Patient/:id", []string{"billing"}, false},
		{"registrar reads Schedule", http.MethodGet, "/fhir/Schedule/123", "/fhir/Schedule/:id", []string{"registrar"}, true},
		{"registrar reads Appointment", http.MethodGet, "/fhir/Appointment/123", "/fhir/Appointment/:id", []string{"registrar"}, true},
		{"patient reads QuestionnaireResponse", http.MethodGet, "/fhir/QuestionnaireResponse/123", "/fhir/QuestionnaireResponse/:id", []string{"patient"}, true},
		{"patient denied Observation", http.MethodGet, "/fhir/Observation/123", "/fhir/Observation/:id", []string{"patient"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(tt.method, tt.path, nil)
			req = req.WithContext(ctxWithRoles(tt.roles...))
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetPath(tt.tmpl)

			called := false
			handler := func(c echo.Context) error {
				called = true
				return c.String(http.StatusOK, "ok")
			}
			mw := ABACMiddleware(engine)
			err := mw(handler)(c)

			if tt.allowed {
				if err != nil {
					t.Fatalf("expected allowed, got error: %v", err)
				}
				if !called {
					t.Error("expected handler to be called")
				}
			} else {
				if err == nil {
					t.Fatal("expected denied, but handler was allowed")
				}
				httpErr, ok := err.(*echo.HTTPError)
				if !ok {
					t.Fatalf("expected echo.HTTPError, got %T", err)
				}
				if httpErr.Code != http.StatusForbidden {
					t.Errorf("expected 403, got %d", httpErr.Code)
				}
			}
		})
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
	c.SetPath(path)
	if requireConsent {
		c.Set("require_consent", true)
	}
	return c, rec
}

// ---------------------------------------------------------------------------
// ConsentEnforcementMiddleware tests
// ---------------------------------------------------------------------------

func TestABAC_ConsentEnforcement_NilChecker_PassThrough(t *testing.T) {
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
	err := mw(handler)(c)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !called {
		t.Error("expected handler to be called when checker is nil")
	}
}

func TestABAC_ConsentEnforcement_NoConsentRequired_PassThrough(t *testing.T) {
	checker := &mockConsentChecker{}
	c, _ := newConsentTestContext(http.MethodGet, "/fhir/Patient", []string{"physician"}, false)

	called := false
	handler := func(c echo.Context) error {
		called = true
		return c.String(http.StatusOK, "ok")
	}
	mw := ConsentEnforcementMiddleware(checker)
	err := mw(handler)(c)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("expected handler to be called when consent is not required")
	}
}

func TestABAC_ConsentEnforcement_ActivePermit_PassesThrough(t *testing.T) {
	patientID := uuid.New()
	checker := &mockConsentChecker{
		consents: []*ConsentInfo{
			{Status: "active", ProvisionType: "permit", ProvisionAction: "access"},
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
	err := mw(handler)(c)

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

func TestABAC_ConsentEnforcement_NoConsentsExist_Returns403(t *testing.T) {
	patientID := uuid.New()
	checker := &mockConsentChecker{consents: []*ConsentInfo{}}

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
	err := mw(handler)(c)

	if err != nil {
		t.Fatalf("unexpected error (should use c.JSON for 403): %v", err)
	}
	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rec.Code)
	}

	var outcome map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &outcome); err != nil {
		t.Fatalf("failed to parse response body: %v", err)
	}
	if outcome["resourceType"] != "OperationOutcome" {
		t.Error("expected OperationOutcome resourceType")
	}
}

func TestABAC_ConsentEnforcement_ExpiredConsent_Returns403(t *testing.T) {
	patientID := uuid.New()
	pastEnd := time.Now().Add(-24 * time.Hour)
	pastStart := time.Now().Add(-48 * time.Hour)
	checker := &mockConsentChecker{
		consents: []*ConsentInfo{
			{Status: "active", ProvisionType: "permit", ProvisionAction: "access",
				ProvisionStart: &pastStart, ProvisionEnd: &pastEnd},
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
	err := mw(handler)(c)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403 for expired consent, got %d", rec.Code)
	}
}

func TestABAC_ConsentEnforcement_FutureConsent_Returns403(t *testing.T) {
	patientID := uuid.New()
	futureStart := time.Now().Add(24 * time.Hour)
	futureEnd := time.Now().Add(48 * time.Hour)
	checker := &mockConsentChecker{
		consents: []*ConsentInfo{
			{Status: "active", ProvisionType: "permit", ProvisionAction: "access",
				ProvisionStart: &futureStart, ProvisionEnd: &futureEnd},
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
	err := mw(handler)(c)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403 for future consent, got %d", rec.Code)
	}
}

func TestABAC_ConsentEnforcement_DenyProvision_Returns403(t *testing.T) {
	patientID := uuid.New()
	checker := &mockConsentChecker{
		consents: []*ConsentInfo{
			{Status: "active", ProvisionType: "deny", ProvisionAction: "access"},
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
	err := mw(handler)(c)

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

func TestABAC_ConsentEnforcement_DenyTakesPrecedenceOverPermit(t *testing.T) {
	patientID := uuid.New()
	checker := &mockConsentChecker{
		consents: []*ConsentInfo{
			{Status: "active", ProvisionType: "permit", ProvisionAction: "access"},
			{Status: "active", ProvisionType: "deny", ProvisionAction: "access"},
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
	err := mw(handler)(c)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403 when deny takes precedence, got %d", rec.Code)
	}
}

func TestABAC_ConsentEnforcement_AdminBypass(t *testing.T) {
	checker := &mockConsentChecker{consents: []*ConsentInfo{}}

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
	err := mw(handler)(c)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("expected handler to be called for admin bypass")
	}
}

func TestABAC_ConsentEnforcement_InactiveConsentIgnored(t *testing.T) {
	patientID := uuid.New()
	checker := &mockConsentChecker{
		consents: []*ConsentInfo{
			{Status: "inactive", ProvisionType: "permit", ProvisionAction: "access"},
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
	err := mw(handler)(c)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403 for inactive consent, got %d", rec.Code)
	}
}

func TestABAC_ConsentEnforcement_WriteAction_RequiresCorrectProvisionAction(t *testing.T) {
	patientID := uuid.New()
	checker := &mockConsentChecker{
		consents: []*ConsentInfo{
			{Status: "active", ProvisionType: "permit", ProvisionAction: "access"},
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
	err := mw(handler)(c)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403 for PUT with access-only consent, got %d", rec.Code)
	}
}

func TestABAC_ConsentEnforcement_EmptyProvisionAction_MatchesAny(t *testing.T) {
	patientID := uuid.New()
	checker := &mockConsentChecker{
		consents: []*ConsentInfo{
			{Status: "active", ProvisionType: "permit", ProvisionAction: ""},
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
	err := mw(handler)(c)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("expected handler to be called when provision action is empty (match-all)")
	}
}

func TestABAC_ConsentEnforcement_PatientIDFromQuerySubject(t *testing.T) {
	patientID := uuid.New()
	checker := &mockConsentChecker{
		consents: []*ConsentInfo{
			{Status: "active", ProvisionType: "permit"},
		},
	}

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
	err := mw(handler)(c)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("expected handler to be called with patient ID from subject query param")
	}
}

func TestABAC_ConsentEnforcement_NoPatientID_Returns403(t *testing.T) {
	checker := &mockConsentChecker{
		consents: []*ConsentInfo{
			{Status: "active", ProvisionType: "permit"},
		},
	}

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
	err := mw(handler)(c)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403 when patient ID cannot be determined, got %d", rec.Code)
	}
}

func TestABAC_ConsentEnforcement_CheckerError_Returns500(t *testing.T) {
	patientID := uuid.New()
	checker := &mockConsentChecker{err: fmt.Errorf("database connection failed")}

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
	err := mw(handler)(c)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rec.Code)
	}
}

// ---------------------------------------------------------------------------
// Consent enforcement: expanded PHI resources
// ---------------------------------------------------------------------------

func TestABAC_ConsentEnforcement_AllPHIResources(t *testing.T) {
	// Every PHI resource should trigger consent enforcement when going through
	// both ABACMiddleware and ConsentEnforcementMiddleware.
	phiResources := []string{
		"Condition", "Observation", "AllergyIntolerance", "Procedure", "NutritionOrder",
		"DiagnosticReport", "ServiceRequest", "ImagingStudy", "Specimen",
		"MedicationRequest", "MedicationAdministration", "MedicationDispense", "MedicationStatement",
		"DocumentReference", "Composition",
		"FamilyMemberHistory", "ClinicalImpression", "RiskAssessment",
		"Flag", "DetectedIssue", "AdverseEvent",
	}

	engine := NewABACEngine(DefaultPolicies())
	checker := &mockConsentChecker{
		consents: []*ConsentInfo{}, // no consents => should deny
	}

	for _, rt := range phiResources {
		t.Run(rt, func(t *testing.T) {
			patientID := uuid.New()
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/fhir/"+rt+"?patient="+patientID.String(), nil)
			req = req.WithContext(ctxWithRoles("physician"))
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetPath("/fhir/" + rt)

			handler := func(c echo.Context) error {
				return c.String(http.StatusOK, "ok")
			}

			// Chain: ABAC -> Consent -> handler
			abacMW := ABACMiddleware(engine)
			consentMW := ConsentEnforcementMiddleware(checker)
			h := abacMW(consentMW(handler))
			err := h(c)

			// ABAC should pass (physician allowed), then consent enforcement
			// should deny (no consents).
			if err != nil {
				t.Fatalf("unexpected echo error: %v", err)
			}
			if rec.Code != http.StatusForbidden {
				t.Errorf("expected 403 (consent denied) for %s, got %d", rt, rec.Code)
			}
		})
	}
}

func TestABAC_ConsentEnforcement_NonPHIResources_SkipConsent(t *testing.T) {
	// Non-PHI resources should not trigger consent enforcement even when
	// there are no consents.
	nonPHI := []string{
		"Patient", "Encounter", "CarePlan", "Practitioner",
		"Organization", "Schedule", "Appointment",
	}

	engine := NewABACEngine(DefaultPolicies())
	checker := &mockConsentChecker{
		consents: []*ConsentInfo{}, // no consents, but consent should not be checked
	}

	for _, rt := range nonPHI {
		t.Run(rt, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/fhir/"+rt+"/123", nil)
			req = req.WithContext(ctxWithRoles("physician"))
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetPath("/fhir/" + rt + "/:id")

			called := false
			handler := func(c echo.Context) error {
				called = true
				return c.String(http.StatusOK, "ok")
			}

			abacMW := ABACMiddleware(engine)
			consentMW := ConsentEnforcementMiddleware(checker)
			h := abacMW(consentMW(handler))
			err := h(c)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !called {
				t.Errorf("expected handler to be called for non-PHI resource %s", rt)
			}
			if rec.Code != http.StatusOK {
				t.Errorf("expected 200 for non-PHI %s, got %d", rt, rec.Code)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// httpMethodToFHIRAction
// ---------------------------------------------------------------------------

func TestABAC_HttpMethodToFHIRAction(t *testing.T) {
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

func TestABAC_ExtractPatientID(t *testing.T) {
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

func TestABAC_ConsentOperationOutcome(t *testing.T) {
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
