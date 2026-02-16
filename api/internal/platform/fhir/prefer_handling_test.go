package fhir

import (
	"net/http"
	"net/http/httptest"
	"sort"
	"testing"

	"github.com/labstack/echo/v4"
)

// ---------------------------------------------------------------------------
// ParsePreferHandling
// ---------------------------------------------------------------------------

func TestParsePreferHandling_Strict(t *testing.T) {
	got := ParsePreferHandling("handling=strict")
	if got != HandlingStrict {
		t.Errorf("expected strict, got %q", got)
	}
}

func TestParsePreferHandling_Lenient(t *testing.T) {
	got := ParsePreferHandling("handling=lenient")
	if got != HandlingLenient {
		t.Errorf("expected lenient, got %q", got)
	}
}

func TestParsePreferHandling_Empty(t *testing.T) {
	got := ParsePreferHandling("")
	if got != HandlingLenient {
		t.Errorf("expected lenient for empty input, got %q", got)
	}
}

func TestParsePreferHandling_Invalid(t *testing.T) {
	got := ParsePreferHandling("handling=unknown")
	if got != HandlingLenient {
		t.Errorf("expected lenient for invalid value, got %q", got)
	}
}

func TestParsePreferHandling_MultipleSemicolon(t *testing.T) {
	got := ParsePreferHandling("return=minimal; handling=strict")
	if got != HandlingStrict {
		t.Errorf("expected strict from semicolon-separated, got %q", got)
	}
}

func TestParsePreferHandling_MultipleComma(t *testing.T) {
	got := ParsePreferHandling("return=minimal, handling=strict")
	if got != HandlingStrict {
		t.Errorf("expected strict from comma-separated, got %q", got)
	}
}

func TestParsePreferHandling_OnlyReturn(t *testing.T) {
	got := ParsePreferHandling("return=representation")
	if got != HandlingLenient {
		t.Errorf("expected lenient when only return is present, got %q", got)
	}
}

func TestParsePreferHandling_WhitespaceAround(t *testing.T) {
	got := ParsePreferHandling("  handling=strict  ")
	if got != HandlingStrict {
		t.Errorf("expected strict with whitespace, got %q", got)
	}
}

// ---------------------------------------------------------------------------
// ParsePreferReturnExtended
// ---------------------------------------------------------------------------

func TestParsePreferReturnExtended_Minimal(t *testing.T) {
	got := ParsePreferReturnExtended("return=minimal")
	if got != ReturnMinimal {
		t.Errorf("expected minimal, got %q", got)
	}
}

func TestParsePreferReturnExtended_Representation(t *testing.T) {
	got := ParsePreferReturnExtended("return=representation")
	if got != ReturnRepresentation {
		t.Errorf("expected representation, got %q", got)
	}
}

func TestParsePreferReturnExtended_OperationOutcome(t *testing.T) {
	got := ParsePreferReturnExtended("return=OperationOutcome")
	if got != ReturnOperationOutcome {
		t.Errorf("expected OperationOutcome, got %q", got)
	}
}

func TestParsePreferReturnExtended_Empty(t *testing.T) {
	got := ParsePreferReturnExtended("")
	if got != "" {
		t.Errorf("expected empty for no header, got %q", got)
	}
}

func TestParsePreferReturnExtended_Invalid(t *testing.T) {
	got := ParsePreferReturnExtended("return=bogus")
	if got != "" {
		t.Errorf("expected empty for invalid value, got %q", got)
	}
}

// ---------------------------------------------------------------------------
// PreferRespondAsync
// ---------------------------------------------------------------------------

func TestPreferRespondAsync_True(t *testing.T) {
	if !PreferRespondAsync("respond-async") {
		t.Error("expected true for respond-async")
	}
}

func TestPreferRespondAsync_TrueWithOther(t *testing.T) {
	if !PreferRespondAsync("return=minimal; respond-async") {
		t.Error("expected true for respond-async among other directives (semicolon)")
	}
}

func TestPreferRespondAsync_TrueComma(t *testing.T) {
	if !PreferRespondAsync("handling=strict, respond-async") {
		t.Error("expected true for respond-async among other directives (comma)")
	}
}

func TestPreferRespondAsync_False(t *testing.T) {
	if PreferRespondAsync("return=minimal") {
		t.Error("expected false when respond-async not present")
	}
}

func TestPreferRespondAsync_Empty(t *testing.T) {
	if PreferRespondAsync("") {
		t.Error("expected false for empty header")
	}
}

func TestPreferRespondAsync_CaseInsensitive(t *testing.T) {
	if !PreferRespondAsync("Respond-Async") {
		t.Error("expected true for case-insensitive respond-async")
	}
}

// ---------------------------------------------------------------------------
// ParsePreferHeader (comprehensive)
// ---------------------------------------------------------------------------

func TestParsePreferHeader_AllDirectives(t *testing.T) {
	d := ParsePreferHeader("return=minimal; handling=strict; respond-async")
	if d.Return != ReturnMinimal {
		t.Errorf("Return: expected minimal, got %q", d.Return)
	}
	if d.Handling != HandlingStrict {
		t.Errorf("Handling: expected strict, got %q", d.Handling)
	}
	if !d.RespondAsync {
		t.Error("RespondAsync: expected true")
	}
}

func TestParsePreferHeader_CommaSeparated(t *testing.T) {
	d := ParsePreferHeader("handling=lenient, return=OperationOutcome, respond-async")
	if d.Return != ReturnOperationOutcome {
		t.Errorf("Return: expected OperationOutcome, got %q", d.Return)
	}
	if d.Handling != HandlingLenient {
		t.Errorf("Handling: expected lenient, got %q", d.Handling)
	}
	if !d.RespondAsync {
		t.Error("RespondAsync: expected true")
	}
}

func TestParsePreferHeader_Empty(t *testing.T) {
	d := ParsePreferHeader("")
	if d.Return != "" {
		t.Errorf("Return: expected empty, got %q", d.Return)
	}
	if d.Handling != HandlingLenient {
		t.Errorf("Handling: expected lenient default, got %q", d.Handling)
	}
	if d.RespondAsync {
		t.Error("RespondAsync: expected false")
	}
}

func TestParsePreferHeader_OnlyReturn(t *testing.T) {
	d := ParsePreferHeader("return=representation")
	if d.Return != ReturnRepresentation {
		t.Errorf("Return: expected representation, got %q", d.Return)
	}
	if d.Handling != HandlingLenient {
		t.Errorf("Handling: expected lenient default, got %q", d.Handling)
	}
	if d.RespondAsync {
		t.Error("RespondAsync: expected false")
	}
}

func TestParsePreferHeader_OnlyHandling(t *testing.T) {
	d := ParsePreferHeader("handling=strict")
	if d.Return != "" {
		t.Errorf("Return: expected empty, got %q", d.Return)
	}
	if d.Handling != HandlingStrict {
		t.Errorf("Handling: expected strict, got %q", d.Handling)
	}
}

func TestParsePreferHeader_OnlyRespondAsync(t *testing.T) {
	d := ParsePreferHeader("respond-async")
	if d.Return != "" {
		t.Errorf("Return: expected empty, got %q", d.Return)
	}
	if d.Handling != HandlingLenient {
		t.Errorf("Handling: expected lenient default, got %q", d.Handling)
	}
	if !d.RespondAsync {
		t.Error("RespondAsync: expected true")
	}
}

func TestParsePreferHeader_MixedSeparators(t *testing.T) {
	// Commas and semicolons mixed: commas replaced by semicolons first
	d := ParsePreferHeader("return=minimal, handling=strict; respond-async")
	if d.Return != ReturnMinimal {
		t.Errorf("Return: expected minimal, got %q", d.Return)
	}
	if d.Handling != HandlingStrict {
		t.Errorf("Handling: expected strict, got %q", d.Handling)
	}
	if !d.RespondAsync {
		t.Error("RespondAsync: expected true")
	}
}

func TestParsePreferHeader_InvalidDirectivesIgnored(t *testing.T) {
	d := ParsePreferHeader("return=minimal; foo=bar; handling=strict; baz")
	if d.Return != ReturnMinimal {
		t.Errorf("Return: expected minimal, got %q", d.Return)
	}
	if d.Handling != HandlingStrict {
		t.Errorf("Handling: expected strict, got %q", d.Handling)
	}
	if d.RespondAsync {
		t.Error("RespondAsync: expected false with unrelated extra directives")
	}
}

// ---------------------------------------------------------------------------
// PreferHandlingMiddleware
// ---------------------------------------------------------------------------

func TestPreferHandlingMiddleware_SetsContextStrict(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient", nil)
	req.Header.Set("Prefer", "handling=strict")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	var captured HandlingPreference
	handler := PreferHandlingMiddleware()(func(c echo.Context) error {
		captured = GetHandlingPreference(c)
		return c.NoContent(http.StatusOK)
	})

	if err := handler(c); err != nil {
		t.Fatal(err)
	}
	if captured != HandlingStrict {
		t.Errorf("context value: expected strict, got %q", captured)
	}
}

func TestPreferHandlingMiddleware_SetsContextLenient(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient", nil)
	req.Header.Set("Prefer", "handling=lenient")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	var captured HandlingPreference
	handler := PreferHandlingMiddleware()(func(c echo.Context) error {
		captured = GetHandlingPreference(c)
		return c.NoContent(http.StatusOK)
	})

	if err := handler(c); err != nil {
		t.Fatal(err)
	}
	if captured != HandlingLenient {
		t.Errorf("context value: expected lenient, got %q", captured)
	}
}

func TestPreferHandlingMiddleware_DefaultsLenient(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	var captured HandlingPreference
	handler := PreferHandlingMiddleware()(func(c echo.Context) error {
		captured = GetHandlingPreference(c)
		return c.NoContent(http.StatusOK)
	})

	if err := handler(c); err != nil {
		t.Fatal(err)
	}
	if captured != HandlingLenient {
		t.Errorf("context value: expected lenient default, got %q", captured)
	}
}

func TestPreferHandlingMiddleware_SetsResponseHeader(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient", nil)
	req.Header.Set("Prefer", "handling=strict")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := PreferHandlingMiddleware()(func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	if err := handler(c); err != nil {
		t.Fatal(err)
	}
	hdr := rec.Header().Get("X-FHIR-Handling")
	if hdr != "strict" {
		t.Errorf("X-FHIR-Handling header: expected strict, got %q", hdr)
	}
}

func TestPreferHandlingMiddleware_ResponseHeaderLenient(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := PreferHandlingMiddleware()(func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	if err := handler(c); err != nil {
		t.Fatal(err)
	}
	hdr := rec.Header().Get("X-FHIR-Handling")
	if hdr != "lenient" {
		t.Errorf("X-FHIR-Handling header: expected lenient, got %q", hdr)
	}
}

// ---------------------------------------------------------------------------
// GetHandlingPreference
// ---------------------------------------------------------------------------

func TestGetHandlingPreference_NoValueSet(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	got := GetHandlingPreference(c)
	if got != HandlingLenient {
		t.Errorf("expected lenient when no value set, got %q", got)
	}
}

func TestGetHandlingPreference_WrongType(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(contextKeyHandling, "not-a-handling-preference")

	got := GetHandlingPreference(c)
	if got != HandlingLenient {
		t.Errorf("expected lenient for wrong type, got %q", got)
	}
}

// ---------------------------------------------------------------------------
// ValidateUnknownElements
// ---------------------------------------------------------------------------

func TestValidateUnknownElements_AllKnown(t *testing.T) {
	resource := map[string]interface{}{
		"resourceType": "Patient",
		"id":           "123",
		"name":         "John",
		"gender":       "male",
	}
	known := map[string]bool{"name": true, "gender": true}
	unknowns := ValidateUnknownElements(resource, known)
	if len(unknowns) != 0 {
		t.Errorf("expected no unknowns, got %v", unknowns)
	}
}

func TestValidateUnknownElements_SomeUnknown(t *testing.T) {
	resource := map[string]interface{}{
		"resourceType": "Patient",
		"id":           "123",
		"name":         "John",
		"foobar":       "baz",
		"customField":  true,
	}
	known := map[string]bool{"name": true}
	unknowns := ValidateUnknownElements(resource, known)
	sort.Strings(unknowns)
	if len(unknowns) != 2 {
		t.Fatalf("expected 2 unknowns, got %d: %v", len(unknowns), unknowns)
	}
	if unknowns[0] != "customField" || unknowns[1] != "foobar" {
		t.Errorf("unexpected unknown elements: %v", unknowns)
	}
}

func TestValidateUnknownElements_EmptyResource(t *testing.T) {
	unknowns := ValidateUnknownElements(map[string]interface{}{}, map[string]bool{"name": true})
	if len(unknowns) != 0 {
		t.Errorf("expected no unknowns for empty resource, got %v", unknowns)
	}
}

func TestValidateUnknownElements_NilResource(t *testing.T) {
	unknowns := ValidateUnknownElements(nil, map[string]bool{"name": true})
	if unknowns != nil {
		t.Errorf("expected nil for nil resource, got %v", unknowns)
	}
}

func TestValidateUnknownElements_MetaAlwaysKnown(t *testing.T) {
	resource := map[string]interface{}{
		"resourceType": "Patient",
		"id":           "1",
		"meta":         map[string]interface{}{"versionId": "1"},
	}
	unknowns := ValidateUnknownElements(resource, map[string]bool{})
	if len(unknowns) != 0 {
		t.Errorf("resourceType/id/meta should always be known, got unknowns: %v", unknowns)
	}
}

func TestValidateUnknownElements_EmptyKnownElements(t *testing.T) {
	resource := map[string]interface{}{
		"resourceType": "Patient",
		"id":           "1",
		"name":         "test",
	}
	unknowns := ValidateUnknownElements(resource, map[string]bool{})
	if len(unknowns) != 1 || unknowns[0] != "name" {
		t.Errorf("expected [name], got %v", unknowns)
	}
}

// ---------------------------------------------------------------------------
// StrictModeResponse
// ---------------------------------------------------------------------------

func TestStrictModeResponse_SingleElement(t *testing.T) {
	resp := StrictModeResponse([]string{"foobar"})
	rt, ok := resp["resourceType"].(string)
	if !ok || rt != "OperationOutcome" {
		t.Errorf("expected resourceType=OperationOutcome, got %v", resp["resourceType"])
	}
	issues, ok := resp["issue"].([]interface{})
	if !ok || len(issues) != 1 {
		t.Fatalf("expected 1 issue, got %v", resp["issue"])
	}
	issue := issues[0].(map[string]interface{})
	if issue["severity"] != IssueSeverityError {
		t.Errorf("severity: expected error, got %v", issue["severity"])
	}
	if issue["code"] != IssueTypeStructure {
		t.Errorf("code: expected structure, got %v", issue["code"])
	}
	diag := issue["diagnostics"].(string)
	if diag != "Unknown element 'foobar' found in resource" {
		t.Errorf("unexpected diagnostics: %s", diag)
	}
	expr := issue["expression"].([]string)
	if len(expr) != 1 || expr[0] != "foobar" {
		t.Errorf("expression: expected [foobar], got %v", expr)
	}
}

func TestStrictModeResponse_MultipleElements(t *testing.T) {
	resp := StrictModeResponse([]string{"alpha", "beta", "gamma"})
	issues, ok := resp["issue"].([]interface{})
	if !ok || len(issues) != 3 {
		t.Fatalf("expected 3 issues, got %v", resp["issue"])
	}
}

func TestStrictModeResponse_EmptySlice(t *testing.T) {
	resp := StrictModeResponse([]string{})
	issues, ok := resp["issue"].([]interface{})
	if !ok || len(issues) != 0 {
		t.Errorf("expected 0 issues for empty input, got %v", resp["issue"])
	}
}

// ---------------------------------------------------------------------------
// DefaultKnownElements
// ---------------------------------------------------------------------------

func TestDefaultKnownElements_HasRequiredResources(t *testing.T) {
	dke := DefaultKnownElements()
	required := []string{
		"Patient", "Observation", "Encounter", "Condition", "Procedure",
		"MedicationRequest", "DiagnosticReport", "AllergyIntolerance",
		"Immunization", "CarePlan", "Organization", "Practitioner",
	}
	for _, rt := range required {
		if _, ok := dke[rt]; !ok {
			t.Errorf("DefaultKnownElements missing resource type %s", rt)
		}
	}
}

func TestDefaultKnownElements_PatientHasGender(t *testing.T) {
	dke := DefaultKnownElements()
	patient := dke["Patient"]
	if !patient["gender"] {
		t.Error("Patient should have gender as a known element")
	}
}

func TestDefaultKnownElements_ObservationHasStatus(t *testing.T) {
	dke := DefaultKnownElements()
	obs := dke["Observation"]
	if !obs["status"] {
		t.Error("Observation should have status as a known element")
	}
}

func TestDefaultKnownElements_EncounterHasClass(t *testing.T) {
	dke := DefaultKnownElements()
	enc := dke["Encounter"]
	if !enc["class"] {
		t.Error("Encounter should have class as a known element")
	}
}

func TestDefaultKnownElements_AllHaveExtension(t *testing.T) {
	dke := DefaultKnownElements()
	for rt, elems := range dke {
		if !elems["extension"] {
			t.Errorf("%s should have extension as a known element", rt)
		}
	}
}

// ---------------------------------------------------------------------------
// Integration: strict mode with unknown elements
// ---------------------------------------------------------------------------

func TestIntegration_StrictModeUnknownElements(t *testing.T) {
	dke := DefaultKnownElements()
	resource := map[string]interface{}{
		"resourceType":  "Patient",
		"id":            "1",
		"gender":        "male",
		"unknownField1": "value1",
		"unknownField2": "value2",
	}

	unknowns := ValidateUnknownElements(resource, dke["Patient"])
	sort.Strings(unknowns)
	if len(unknowns) != 2 {
		t.Fatalf("expected 2 unknowns, got %d: %v", len(unknowns), unknowns)
	}
	if unknowns[0] != "unknownField1" || unknowns[1] != "unknownField2" {
		t.Errorf("unexpected unknowns: %v", unknowns)
	}

	resp := StrictModeResponse(unknowns)
	issues := resp["issue"].([]interface{})
	if len(issues) != 2 {
		t.Errorf("expected 2 issues in response, got %d", len(issues))
	}
}

// ---------------------------------------------------------------------------
// Integration: lenient mode ignores unknown elements
// ---------------------------------------------------------------------------

func TestIntegration_LenientModeIgnoresUnknowns(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/fhir/Patient", nil)
	req.Header.Set("Prefer", "handling=lenient")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := PreferHandlingMiddleware()(func(c echo.Context) error {
		handling := GetHandlingPreference(c)
		if handling != HandlingLenient {
			t.Errorf("expected lenient handling in middleware, got %q", handling)
		}
		// In lenient mode, unknown elements are silently ignored; no validation error.
		return c.String(http.StatusCreated, `{"resourceType":"Patient","id":"1"}`)
	})

	if err := handler(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestIntegration_StrictModeMiddlewareFlow(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/fhir/Patient", nil)
	req.Header.Set("Prefer", "handling=strict")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	dke := DefaultKnownElements()

	handler := PreferHandlingMiddleware()(func(c echo.Context) error {
		handling := GetHandlingPreference(c)
		if handling != HandlingStrict {
			t.Errorf("expected strict handling, got %q", handling)
		}

		// Simulate parsing a resource with unknown elements
		resource := map[string]interface{}{
			"resourceType": "Patient",
			"id":           "1",
			"gender":       "male",
			"bogusElement": "should fail",
		}
		unknowns := ValidateUnknownElements(resource, dke["Patient"])
		if len(unknowns) > 0 {
			return c.JSON(http.StatusBadRequest, StrictModeResponse(unknowns))
		}
		return c.String(http.StatusCreated, `{"resourceType":"Patient","id":"1"}`)
	})

	if err := handler(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for strict mode with unknown element, got %d", rec.Code)
	}
	body := rec.Body.String()
	if body == "" {
		t.Fatal("expected non-empty OperationOutcome body")
	}
	if !contains(body, "OperationOutcome") {
		t.Errorf("response should contain OperationOutcome, got: %s", body)
	}
	if !contains(body, "bogusElement") {
		t.Errorf("response should mention bogusElement, got: %s", body)
	}
}

// testContains is a test helper to check substring presence.
func testContains(s, substr string) bool {
	return len(s) >= len(substr) && searchSubstring(s, substr)
}

func searchSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
