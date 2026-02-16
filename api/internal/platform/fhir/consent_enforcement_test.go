package fhir

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
)

// ---------------------------------------------------------------------------
// Helper factories
// ---------------------------------------------------------------------------

func timePtr(t time.Time) *time.Time { return &t }

var (
	testNow   = time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)
	testPast  = time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	testFuture = time.Date(2026, 12, 31, 23, 59, 59, 0, time.UTC)
)

func baseRequest() ConsentAccessRequest {
	return ConsentAccessRequest{
		PatientID:      "patient-1",
		ActorReference: "Practitioner/dr-smith",
		ResourceType:   "Observation",
		Purpose:        "TREAT",
		SecurityLabels: []string{"N"},
		AccessTime:     testNow,
	}
}

func activePermitPolicy(id string) ConsentPolicy {
	return ConsentPolicy{
		ID:        id,
		PatientID: "patient-1",
		Scope:     ConsentScopePatientPrivacy,
		Status:    ConsentStatusActive,
		Provision: ConsentProvision{
			Type: "permit",
		},
		CreatedAt: testPast,
	}
}

func activeDenyPolicy(id string) ConsentPolicy {
	return ConsentPolicy{
		ID:        id,
		PatientID: "patient-1",
		Scope:     ConsentScopePatientPrivacy,
		Status:    ConsentStatusActive,
		Provision: ConsentProvision{
			Type: "deny",
		},
		CreatedAt: testPast,
	}
}

// =========== EvaluateConsent Tests ===========

func TestEvaluateConsent_ActivePermit(t *testing.T) {
	policies := []ConsentPolicy{activePermitPolicy("c1")}
	decision := EvaluateConsent(policies, baseRequest())
	if decision != ConsentDecisionPermit {
		t.Errorf("expected permit, got %q", decision)
	}
}

func TestEvaluateConsent_ActiveDeny(t *testing.T) {
	policies := []ConsentPolicy{activeDenyPolicy("c1")}
	decision := EvaluateConsent(policies, baseRequest())
	if decision != ConsentDecisionDeny {
		t.Errorf("expected deny, got %q", decision)
	}
}

func TestEvaluateConsent_NoPolicies(t *testing.T) {
	decision := EvaluateConsent(nil, baseRequest())
	if decision != ConsentDecisionNoConsent {
		t.Errorf("expected no-consent, got %q", decision)
	}
}

func TestEvaluateConsent_EmptyPolicies(t *testing.T) {
	decision := EvaluateConsent([]ConsentPolicy{}, baseRequest())
	if decision != ConsentDecisionNoConsent {
		t.Errorf("expected no-consent, got %q", decision)
	}
}

func TestEvaluateConsent_DenyOverridesPermit(t *testing.T) {
	policies := []ConsentPolicy{
		activePermitPolicy("c1"),
		activeDenyPolicy("c2"),
	}
	decision := EvaluateConsent(policies, baseRequest())
	if decision != ConsentDecisionDeny {
		t.Errorf("deny should override permit, got %q", decision)
	}
}

func TestEvaluateConsent_DenyOverridesPermitReverseOrder(t *testing.T) {
	policies := []ConsentPolicy{
		activeDenyPolicy("c1"),
		activePermitPolicy("c2"),
	}
	decision := EvaluateConsent(policies, baseRequest())
	if decision != ConsentDecisionDeny {
		t.Errorf("deny should override permit regardless of order, got %q", decision)
	}
}

func TestEvaluateConsent_ExpiredPolicyProvisionPeriod(t *testing.T) {
	expired := activePermitPolicy("c1")
	expEnd := testPast.Add(24 * time.Hour) // expired well before testNow
	expired.Provision.Period = &Period{
		Start: &testPast,
		End:   &expEnd,
	}
	policies := []ConsentPolicy{expired}
	decision := EvaluateConsent(policies, baseRequest())
	if decision != ConsentDecisionNoConsent {
		t.Errorf("expired provision should not match, got %q", decision)
	}
}

func TestEvaluateConsent_FutureProvisionPeriod(t *testing.T) {
	future := activePermitPolicy("c1")
	futureStart := testFuture
	future.Provision.Period = &Period{
		Start: &futureStart,
	}
	policies := []ConsentPolicy{future}
	decision := EvaluateConsent(policies, baseRequest())
	if decision != ConsentDecisionNoConsent {
		t.Errorf("future provision should not match, got %q", decision)
	}
}

func TestEvaluateConsent_ValidProvisionPeriod(t *testing.T) {
	valid := activePermitPolicy("c1")
	valid.Provision.Period = &Period{
		Start: &testPast,
		End:   &testFuture,
	}
	policies := []ConsentPolicy{valid}
	decision := EvaluateConsent(policies, baseRequest())
	if decision != ConsentDecisionPermit {
		t.Errorf("provision within valid period should match, got %q", decision)
	}
}

func TestEvaluateConsent_ActorRestriction_Match(t *testing.T) {
	policy := activePermitPolicy("c1")
	policy.Provision.Actor = []ConsentActor{
		{Role: "primary", Reference: "Practitioner/dr-smith"},
	}
	policies := []ConsentPolicy{policy}
	decision := EvaluateConsent(policies, baseRequest())
	if decision != ConsentDecisionPermit {
		t.Errorf("actor reference matches, expected permit, got %q", decision)
	}
}

func TestEvaluateConsent_ActorRestriction_NoMatch(t *testing.T) {
	policy := activePermitPolicy("c1")
	policy.Provision.Actor = []ConsentActor{
		{Role: "primary", Reference: "Practitioner/dr-jones"},
	}
	policies := []ConsentPolicy{policy}
	decision := EvaluateConsent(policies, baseRequest())
	if decision != ConsentDecisionNoConsent {
		t.Errorf("actor reference does not match, expected no-consent, got %q", decision)
	}
}

func TestEvaluateConsent_ActorRestriction_MultipleActors(t *testing.T) {
	policy := activePermitPolicy("c1")
	policy.Provision.Actor = []ConsentActor{
		{Role: "primary", Reference: "Practitioner/dr-jones"},
		{Role: "delegated", Reference: "Practitioner/dr-smith"},
	}
	policies := []ConsentPolicy{policy}
	decision := EvaluateConsent(policies, baseRequest())
	if decision != ConsentDecisionPermit {
		t.Errorf("one of the actors matches, expected permit, got %q", decision)
	}
}

func TestEvaluateConsent_ResourceTypeRestriction_Match(t *testing.T) {
	policy := activePermitPolicy("c1")
	policy.Provision.ResourceClass = []string{"Observation", "Condition"}
	policies := []ConsentPolicy{policy}
	decision := EvaluateConsent(policies, baseRequest())
	if decision != ConsentDecisionPermit {
		t.Errorf("resource type matches, expected permit, got %q", decision)
	}
}

func TestEvaluateConsent_ResourceTypeRestriction_NoMatch(t *testing.T) {
	policy := activePermitPolicy("c1")
	policy.Provision.ResourceClass = []string{"MedicationRequest", "Condition"}
	policies := []ConsentPolicy{policy}
	decision := EvaluateConsent(policies, baseRequest())
	if decision != ConsentDecisionNoConsent {
		t.Errorf("resource type does not match, expected no-consent, got %q", decision)
	}
}

func TestEvaluateConsent_PurposeRestriction_Match(t *testing.T) {
	policy := activePermitPolicy("c1")
	policy.Provision.Purpose = []string{"TREAT", "HPAYMT"}
	policies := []ConsentPolicy{policy}
	decision := EvaluateConsent(policies, baseRequest())
	if decision != ConsentDecisionPermit {
		t.Errorf("purpose matches, expected permit, got %q", decision)
	}
}

func TestEvaluateConsent_PurposeRestriction_NoMatch(t *testing.T) {
	policy := activePermitPolicy("c1")
	policy.Provision.Purpose = []string{"HPAYMT", "HOPERAT"}
	policies := []ConsentPolicy{policy}
	decision := EvaluateConsent(policies, baseRequest())
	if decision != ConsentDecisionNoConsent {
		t.Errorf("purpose does not match, expected no-consent, got %q", decision)
	}
}

func TestEvaluateConsent_SecurityLabelRestriction_Match(t *testing.T) {
	policy := activePermitPolicy("c1")
	policy.Provision.SecurityLabel = []string{"N", "R"}
	policies := []ConsentPolicy{policy}
	decision := EvaluateConsent(policies, baseRequest())
	if decision != ConsentDecisionPermit {
		t.Errorf("security label matches, expected permit, got %q", decision)
	}
}

func TestEvaluateConsent_SecurityLabelRestriction_NoMatch(t *testing.T) {
	policy := activePermitPolicy("c1")
	policy.Provision.SecurityLabel = []string{"V", "R"}
	policies := []ConsentPolicy{policy}
	decision := EvaluateConsent(policies, baseRequest())
	if decision != ConsentDecisionNoConsent {
		t.Errorf("security label does not match, expected no-consent, got %q", decision)
	}
}

func TestEvaluateConsent_InactivePolicy_Ignored(t *testing.T) {
	policy := activePermitPolicy("c1")
	policy.Status = ConsentStatusInactive
	policies := []ConsentPolicy{policy}
	decision := EvaluateConsent(policies, baseRequest())
	if decision != ConsentDecisionNoConsent {
		t.Errorf("inactive policy should be ignored, got %q", decision)
	}
}

func TestEvaluateConsent_DraftPolicy_Ignored(t *testing.T) {
	policy := activePermitPolicy("c1")
	policy.Status = ConsentStatusDraft
	policies := []ConsentPolicy{policy}
	decision := EvaluateConsent(policies, baseRequest())
	if decision != ConsentDecisionNoConsent {
		t.Errorf("draft policy should be ignored, got %q", decision)
	}
}

func TestEvaluateConsent_RejectedPolicy_Ignored(t *testing.T) {
	policy := activePermitPolicy("c1")
	policy.Status = ConsentStatusRejected
	policies := []ConsentPolicy{policy}
	decision := EvaluateConsent(policies, baseRequest())
	if decision != ConsentDecisionNoConsent {
		t.Errorf("rejected policy should be ignored, got %q", decision)
	}
}

func TestEvaluateConsent_DataPeriod_Match(t *testing.T) {
	policy := activePermitPolicy("c1")
	policy.Provision.DataPeriod = &Period{
		Start: &testPast,
		End:   &testFuture,
	}
	policies := []ConsentPolicy{policy}
	decision := EvaluateConsent(policies, baseRequest())
	if decision != ConsentDecisionPermit {
		t.Errorf("access time within data period, expected permit, got %q", decision)
	}
}

func TestEvaluateConsent_DataPeriod_NoMatch(t *testing.T) {
	policy := activePermitPolicy("c1")
	dpEnd := testPast.Add(24 * time.Hour) // data period ends well before testNow
	policy.Provision.DataPeriod = &Period{
		Start: &testPast,
		End:   &dpEnd,
	}
	policies := []ConsentPolicy{policy}
	decision := EvaluateConsent(policies, baseRequest())
	if decision != ConsentDecisionNoConsent {
		t.Errorf("access time outside data period, expected no-consent, got %q", decision)
	}
}

func TestEvaluateConsent_MultiplePolicies_OnlyOneMatches(t *testing.T) {
	// First policy restricts to a different resource type.
	p1 := activePermitPolicy("c1")
	p1.Provision.ResourceClass = []string{"MedicationRequest"}

	// Second policy matches.
	p2 := activePermitPolicy("c2")

	policies := []ConsentPolicy{p1, p2}
	decision := EvaluateConsent(policies, baseRequest())
	if decision != ConsentDecisionPermit {
		t.Errorf("second policy matches, expected permit, got %q", decision)
	}
}

func TestEvaluateConsent_MultiplePolicies_DenyOnSpecificResource(t *testing.T) {
	// Broad permit policy.
	p1 := activePermitPolicy("c1")

	// Specific deny for Observation.
	p2 := activeDenyPolicy("c2")
	p2.Provision.ResourceClass = []string{"Observation"}

	policies := []ConsentPolicy{p1, p2}
	decision := EvaluateConsent(policies, baseRequest())
	if decision != ConsentDecisionDeny {
		t.Errorf("deny on Observation should override broad permit, got %q", decision)
	}
}

func TestEvaluateConsent_MultiplePolicies_DenyDoesNotMatchResourceType(t *testing.T) {
	// Broad permit.
	p1 := activePermitPolicy("c1")

	// Deny for MedicationRequest only — should not affect Observation access.
	p2 := activeDenyPolicy("c2")
	p2.Provision.ResourceClass = []string{"MedicationRequest"}

	policies := []ConsentPolicy{p1, p2}
	req := baseRequest() // ResourceType = "Observation"
	decision := EvaluateConsent(policies, req)
	if decision != ConsentDecisionPermit {
		t.Errorf("deny on MedicationRequest should not affect Observation, got %q", decision)
	}
}

func TestEvaluateConsent_CombinedRestrictions(t *testing.T) {
	policy := activePermitPolicy("c1")
	policy.Provision.Actor = []ConsentActor{
		{Role: "primary", Reference: "Practitioner/dr-smith"},
	}
	policy.Provision.ResourceClass = []string{"Observation"}
	policy.Provision.Purpose = []string{"TREAT"}
	policy.Provision.Period = &Period{Start: &testPast, End: &testFuture}

	policies := []ConsentPolicy{policy}
	decision := EvaluateConsent(policies, baseRequest())
	if decision != ConsentDecisionPermit {
		t.Errorf("all restrictions satisfied, expected permit, got %q", decision)
	}
}

func TestEvaluateConsent_CombinedRestrictions_OneFailsActor(t *testing.T) {
	policy := activePermitPolicy("c1")
	policy.Provision.Actor = []ConsentActor{
		{Role: "primary", Reference: "Practitioner/dr-jones"},
	}
	policy.Provision.ResourceClass = []string{"Observation"}
	policy.Provision.Purpose = []string{"TREAT"}

	policies := []ConsentPolicy{policy}
	decision := EvaluateConsent(policies, baseRequest())
	if decision != ConsentDecisionNoConsent {
		t.Errorf("actor mismatch should prevent match, got %q", decision)
	}
}

func TestEvaluateConsent_OpenEndedPeriod_NoStart(t *testing.T) {
	policy := activePermitPolicy("c1")
	policy.Provision.Period = &Period{
		End: &testFuture,
	}
	policies := []ConsentPolicy{policy}
	decision := EvaluateConsent(policies, baseRequest())
	if decision != ConsentDecisionPermit {
		t.Errorf("open start period should match, got %q", decision)
	}
}

func TestEvaluateConsent_OpenEndedPeriod_NoEnd(t *testing.T) {
	policy := activePermitPolicy("c1")
	policy.Provision.Period = &Period{
		Start: &testPast,
	}
	policies := []ConsentPolicy{policy}
	decision := EvaluateConsent(policies, baseRequest())
	if decision != ConsentDecisionPermit {
		t.Errorf("open end period should match, got %q", decision)
	}
}

// =========== Period Tests ===========

func TestPeriod_Contains_NilPeriod(t *testing.T) {
	var p *Period
	if !p.Contains(testNow) {
		t.Error("nil period should contain any time")
	}
}

func TestPeriod_Contains_BothBounds(t *testing.T) {
	p := &Period{Start: &testPast, End: &testFuture}
	if !p.Contains(testNow) {
		t.Error("testNow should be within [testPast, testFuture]")
	}
}

func TestPeriod_Contains_BeforeStart(t *testing.T) {
	earlyTime := testPast.Add(-24 * time.Hour)
	p := &Period{Start: &testPast, End: &testFuture}
	if p.Contains(earlyTime) {
		t.Error("time before start should not be contained")
	}
}

func TestPeriod_Contains_AfterEnd(t *testing.T) {
	lateTime := testFuture.Add(24 * time.Hour)
	p := &Period{Start: &testPast, End: &testFuture}
	if p.Contains(lateTime) {
		t.Error("time after end should not be contained")
	}
}

// =========== InMemoryConsentStore Tests ===========

func TestInMemoryConsentStore_AddAndGet(t *testing.T) {
	store := NewInMemoryConsentStore()
	policy := activePermitPolicy("c1")

	if err := store.AddConsent(policy); err != nil {
		t.Fatalf("AddConsent failed: %v", err)
	}

	policies, err := store.GetActiveConsents("patient-1")
	if err != nil {
		t.Fatalf("GetActiveConsents failed: %v", err)
	}
	if len(policies) != 1 {
		t.Fatalf("expected 1 policy, got %d", len(policies))
	}
	if policies[0].ID != "c1" {
		t.Errorf("expected policy ID c1, got %q", policies[0].ID)
	}
}

func TestInMemoryConsentStore_OnlyReturnsActiveConsents(t *testing.T) {
	store := NewInMemoryConsentStore()

	active := activePermitPolicy("c1")
	draft := ConsentPolicy{
		ID:        "c2",
		PatientID: "patient-1",
		Status:    ConsentStatusDraft,
		Provision: ConsentProvision{Type: "permit"},
	}
	inactive := ConsentPolicy{
		ID:        "c3",
		PatientID: "patient-1",
		Status:    ConsentStatusInactive,
		Provision: ConsentProvision{Type: "permit"},
	}

	store.AddConsent(active)
	store.AddConsent(draft)
	store.AddConsent(inactive)

	policies, err := store.GetActiveConsents("patient-1")
	if err != nil {
		t.Fatalf("GetActiveConsents failed: %v", err)
	}
	if len(policies) != 1 {
		t.Fatalf("expected 1 active policy, got %d", len(policies))
	}
	if policies[0].ID != "c1" {
		t.Errorf("expected policy c1, got %q", policies[0].ID)
	}
}

func TestInMemoryConsentStore_GetByPatientID(t *testing.T) {
	store := NewInMemoryConsentStore()

	p1 := activePermitPolicy("c1")
	p1.PatientID = "patient-1"

	p2 := activePermitPolicy("c2")
	p2.PatientID = "patient-2"

	store.AddConsent(p1)
	store.AddConsent(p2)

	policies, _ := store.GetActiveConsents("patient-1")
	if len(policies) != 1 {
		t.Fatalf("expected 1 policy for patient-1, got %d", len(policies))
	}

	policies, _ = store.GetActiveConsents("patient-2")
	if len(policies) != 1 {
		t.Fatalf("expected 1 policy for patient-2, got %d", len(policies))
	}

	policies, _ = store.GetActiveConsents("patient-3")
	if len(policies) != 0 {
		t.Fatalf("expected 0 policies for patient-3, got %d", len(policies))
	}
}

func TestInMemoryConsentStore_RevokeConsent(t *testing.T) {
	store := NewInMemoryConsentStore()
	policy := activePermitPolicy("c1")
	store.AddConsent(policy)

	if err := store.RevokeConsent("c1"); err != nil {
		t.Fatalf("RevokeConsent failed: %v", err)
	}

	policies, _ := store.GetActiveConsents("patient-1")
	if len(policies) != 0 {
		t.Errorf("revoked policy should not appear in active consents, got %d", len(policies))
	}
}

func TestInMemoryConsentStore_RevokeConsent_NotFound(t *testing.T) {
	store := NewInMemoryConsentStore()
	err := store.RevokeConsent("nonexistent")
	if err == nil {
		t.Error("expected error when revoking nonexistent consent")
	}
}

func TestInMemoryConsentStore_OverwriteExisting(t *testing.T) {
	store := NewInMemoryConsentStore()

	p1 := activePermitPolicy("c1")
	store.AddConsent(p1)

	p1Updated := activeDenyPolicy("c1")
	store.AddConsent(p1Updated)

	policies, _ := store.GetActiveConsents("patient-1")
	if len(policies) != 1 {
		t.Fatalf("expected 1 policy after overwrite, got %d", len(policies))
	}
	if policies[0].Provision.Type != "deny" {
		t.Errorf("expected overwritten policy to be deny, got %q", policies[0].Provision.Type)
	}
}

// =========== Middleware Tests ===========

func consentTestHandler() echo.HandlerFunc {
	return func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"resourceType": "Observation",
			"id":           "obs-1",
		})
	}
}

func TestMiddleware_DeniedAccess_Returns403(t *testing.T) {
	store := NewInMemoryConsentStore()
	store.AddConsent(activeDenyPolicy("c1"))

	e := echo.New()
	e.Use(ConsentEnforcementMiddleware(store))
	e.GET("/fhir/Observation/:id", consentTestHandler())

	req := httptest.NewRequest(http.MethodGet, "/fhir/Observation/obs-1", nil)
	req.Header.Set("X-Patient-ID", "patient-1")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rec.Code)
	}

	var result map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &result)
	if result["resourceType"] != "OperationOutcome" {
		t.Errorf("expected OperationOutcome, got %v", result["resourceType"])
	}

	if rec.Header().Get("X-Consent-Decision") != "deny" {
		t.Errorf("expected X-Consent-Decision=deny, got %q", rec.Header().Get("X-Consent-Decision"))
	}
}

func TestMiddleware_PermittedAccess_PassesThrough(t *testing.T) {
	store := NewInMemoryConsentStore()
	store.AddConsent(activePermitPolicy("c1"))

	e := echo.New()
	e.Use(ConsentEnforcementMiddleware(store))
	e.GET("/fhir/Observation/:id", consentTestHandler())

	req := httptest.NewRequest(http.MethodGet, "/fhir/Observation/obs-1", nil)
	req.Header.Set("X-Patient-ID", "patient-1")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	if rec.Header().Get("X-Consent-Decision") != "permit" {
		t.Errorf("expected X-Consent-Decision=permit, got %q", rec.Header().Get("X-Consent-Decision"))
	}
}

func TestMiddleware_NoConsent_DefaultPermit(t *testing.T) {
	store := NewInMemoryConsentStore()

	e := echo.New()
	e.Use(ConsentEnforcementMiddleware(store))
	e.GET("/fhir/Observation/:id", consentTestHandler())

	req := httptest.NewRequest(http.MethodGet, "/fhir/Observation/obs-1", nil)
	req.Header.Set("X-Patient-ID", "patient-1")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 (default permit), got %d", rec.Code)
	}

	if rec.Header().Get("X-Consent-Decision") != "permit" {
		t.Errorf("expected X-Consent-Decision=permit, got %q", rec.Header().Get("X-Consent-Decision"))
	}
}

func TestMiddleware_OptIn_NoConsent_Denies(t *testing.T) {
	store := NewInMemoryConsentStore()

	config := ConsentEnforcementConfig{
		DefaultDecision: ConsentDecisionDeny,
		RequireConsent:  true,
	}

	e := echo.New()
	e.Use(NewConsentEnforcementMiddleware(store, config))
	e.GET("/fhir/Observation/:id", consentTestHandler())

	req := httptest.NewRequest(http.MethodGet, "/fhir/Observation/obs-1", nil)
	req.Header.Set("X-Patient-ID", "patient-1")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403 (opt-in, no consent), got %d", rec.Code)
	}

	if rec.Header().Get("X-Consent-Decision") != "deny" {
		t.Errorf("expected X-Consent-Decision=deny, got %q", rec.Header().Get("X-Consent-Decision"))
	}
}

func TestMiddleware_OptOut_NoConsent_Permits(t *testing.T) {
	store := NewInMemoryConsentStore()

	config := ConsentEnforcementConfig{
		DefaultDecision: ConsentDecisionPermit,
		RequireConsent:  false,
	}

	e := echo.New()
	e.Use(NewConsentEnforcementMiddleware(store, config))
	e.GET("/fhir/Observation/:id", consentTestHandler())

	req := httptest.NewRequest(http.MethodGet, "/fhir/Observation/obs-1", nil)
	req.Header.Set("X-Patient-ID", "patient-1")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 (opt-out, no consent), got %d", rec.Code)
	}

	if rec.Header().Get("X-Consent-Decision") != "permit" {
		t.Errorf("expected X-Consent-Decision=permit, got %q", rec.Header().Get("X-Consent-Decision"))
	}
}

func TestMiddleware_ExemptResourceTypes(t *testing.T) {
	store := NewInMemoryConsentStore()
	store.AddConsent(activeDenyPolicy("c1"))

	config := ConsentEnforcementConfig{
		DefaultDecision:     ConsentDecisionDeny,
		RequireConsent:      true,
		ExemptResourceTypes: []string{"CapabilityStatement", "OperationDefinition"},
	}

	e := echo.New()
	e.Use(NewConsentEnforcementMiddleware(store, config))
	e.GET("/fhir/CapabilityStatement", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"resourceType": "CapabilityStatement",
		})
	})

	req := httptest.NewRequest(http.MethodGet, "/fhir/CapabilityStatement", nil)
	req.Header.Set("X-Patient-ID", "patient-1")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("exempt resource should pass through, got %d", rec.Code)
	}

	if rec.Header().Get("X-Consent-Decision") != "permit" {
		t.Errorf("exempt resource should get permit decision, got %q", rec.Header().Get("X-Consent-Decision"))
	}
}

func TestMiddleware_NoPatientID_DefaultPermit(t *testing.T) {
	store := NewInMemoryConsentStore()

	e := echo.New()
	e.Use(ConsentEnforcementMiddleware(store))
	e.GET("/fhir/Observation", consentTestHandler())

	req := httptest.NewRequest(http.MethodGet, "/fhir/Observation", nil)
	// No X-Patient-ID header.
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("no patient ID with default permit should pass through, got %d", rec.Code)
	}
}

func TestMiddleware_NoPatientID_RequireConsent_Denies(t *testing.T) {
	store := NewInMemoryConsentStore()

	config := ConsentEnforcementConfig{
		DefaultDecision: ConsentDecisionDeny,
		RequireConsent:  true,
	}

	e := echo.New()
	e.Use(NewConsentEnforcementMiddleware(store, config))
	e.GET("/fhir/Observation", consentTestHandler())

	req := httptest.NewRequest(http.MethodGet, "/fhir/Observation", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("no patient ID with deny default should return 403, got %d", rec.Code)
	}
}

func TestMiddleware_PatientIDFromPathParam(t *testing.T) {
	store := NewInMemoryConsentStore()
	store.AddConsent(activeDenyPolicy("c1"))

	e := echo.New()
	e.Use(ConsentEnforcementMiddleware(store))
	e.GET("/fhir/Patient/:patientId/Observation", consentTestHandler())

	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient/patient-1/Observation", nil)
	// No X-Patient-ID header — should pick up from path param.
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("deny policy should apply via path param patient ID, got %d", rec.Code)
	}
}

func TestMiddleware_SetsConsentDecisionHeader(t *testing.T) {
	store := NewInMemoryConsentStore()
	store.AddConsent(activePermitPolicy("c1"))

	e := echo.New()
	e.Use(ConsentEnforcementMiddleware(store))
	e.GET("/fhir/Observation/:id", consentTestHandler())

	req := httptest.NewRequest(http.MethodGet, "/fhir/Observation/obs-1", nil)
	req.Header.Set("X-Patient-ID", "patient-1")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	header := rec.Header().Get("X-Consent-Decision")
	if header != "permit" {
		t.Errorf("expected X-Consent-Decision=permit, got %q", header)
	}
}

func TestMiddleware_ActorFromHeader(t *testing.T) {
	store := NewInMemoryConsentStore()

	// Permit only for dr-smith.
	policy := activePermitPolicy("c1")
	policy.Provision.Actor = []ConsentActor{
		{Role: "primary", Reference: "Practitioner/dr-smith"},
	}
	store.AddConsent(policy)

	config := ConsentEnforcementConfig{
		DefaultDecision: ConsentDecisionDeny,
		RequireConsent:  true,
	}

	e := echo.New()
	e.Use(NewConsentEnforcementMiddleware(store, config))
	e.GET("/fhir/Observation/:id", consentTestHandler())

	// Request from dr-smith (should be permitted).
	req := httptest.NewRequest(http.MethodGet, "/fhir/Observation/obs-1", nil)
	req.Header.Set("X-Patient-ID", "patient-1")
	req.Header.Set("X-Actor-Reference", "Practitioner/dr-smith")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("dr-smith should be permitted, got %d", rec.Code)
	}

	// Request from dr-jones (should be denied).
	req2 := httptest.NewRequest(http.MethodGet, "/fhir/Observation/obs-1", nil)
	req2.Header.Set("X-Patient-ID", "patient-1")
	req2.Header.Set("X-Actor-Reference", "Practitioner/dr-jones")
	rec2 := httptest.NewRecorder()
	e.ServeHTTP(rec2, req2)

	if rec2.Code != http.StatusForbidden {
		t.Errorf("dr-jones should be denied, got %d", rec2.Code)
	}
}

func TestMiddleware_PurposeFromHeader(t *testing.T) {
	store := NewInMemoryConsentStore()

	// Permit only for TREAT purpose.
	policy := activePermitPolicy("c1")
	policy.Provision.Purpose = []string{"TREAT"}
	store.AddConsent(policy)

	config := ConsentEnforcementConfig{
		DefaultDecision: ConsentDecisionDeny,
		RequireConsent:  true,
	}

	e := echo.New()
	e.Use(NewConsentEnforcementMiddleware(store, config))
	e.GET("/fhir/Observation/:id", consentTestHandler())

	// TREAT purpose should be permitted.
	req := httptest.NewRequest(http.MethodGet, "/fhir/Observation/obs-1", nil)
	req.Header.Set("X-Patient-ID", "patient-1")
	req.Header.Set("X-Purpose-Of-Use", "TREAT")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("TREAT purpose should be permitted, got %d", rec.Code)
	}

	// HPAYMT purpose should be denied.
	req2 := httptest.NewRequest(http.MethodGet, "/fhir/Observation/obs-1", nil)
	req2.Header.Set("X-Patient-ID", "patient-1")
	req2.Header.Set("X-Purpose-Of-Use", "HPAYMT")
	rec2 := httptest.NewRecorder()
	e.ServeHTTP(rec2, req2)

	if rec2.Code != http.StatusForbidden {
		t.Errorf("HPAYMT purpose should be denied, got %d", rec2.Code)
	}
}

// =========== extractResourceTypeFromPath Tests ===========

func TestExtractResourceTypeFromPath(t *testing.T) {
	tests := []struct {
		path     string
		expected string
	}{
		{"/fhir/Patient/123", "Patient"},
		{"/fhir/Observation", "Observation"},
		{"/fhir/MedicationRequest/456", "MedicationRequest"},
		{"/fhir/CapabilityStatement", "CapabilityStatement"},
		{"/api/v1/fhir/Condition/789", "Condition"},
		{"/lowercase/path/only", ""},
		{"", ""},
	}

	for _, tt := range tests {
		got := extractResourceTypeFromPath(tt.path)
		if got != tt.expected {
			t.Errorf("extractResourceTypeFromPath(%q) = %q, want %q", tt.path, got, tt.expected)
		}
	}
}

// =========== ConsentScope / ConsentStatus / ConsentDecision constant tests ===========

func TestConsentScopeConstants(t *testing.T) {
	scopes := []ConsentScope{
		ConsentScopePatientPrivacy,
		ConsentScopeResearch,
		ConsentScopeADR,
		ConsentScopeTreatment,
	}
	expected := []string{"patient-privacy", "research", "adr", "treatment"}
	for i, s := range scopes {
		if string(s) != expected[i] {
			t.Errorf("ConsentScope constant %d = %q, want %q", i, s, expected[i])
		}
	}
}

func TestConsentStatusConstants(t *testing.T) {
	statuses := []ConsentStatus{
		ConsentStatusDraft,
		ConsentStatusProposed,
		ConsentStatusActive,
		ConsentStatusRejected,
		ConsentStatusInactive,
		ConsentStatusEnteredInError,
	}
	expected := []string{"draft", "proposed", "active", "rejected", "inactive", "entered-in-error"}
	for i, s := range statuses {
		if string(s) != expected[i] {
			t.Errorf("ConsentStatus constant %d = %q, want %q", i, s, expected[i])
		}
	}
}

func TestConsentDecisionConstants(t *testing.T) {
	decisions := []ConsentDecision{
		ConsentDecisionPermit,
		ConsentDecisionDeny,
		ConsentDecisionNoConsent,
	}
	expected := []string{"permit", "deny", "no-consent"}
	for i, d := range decisions {
		if string(d) != expected[i] {
			t.Errorf("ConsentDecision constant %d = %q, want %q", i, d, expected[i])
		}
	}
}
