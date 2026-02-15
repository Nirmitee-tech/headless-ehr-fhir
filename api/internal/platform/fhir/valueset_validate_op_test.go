package fhir

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
)

// =========== ValueSetValidator Tests ===========

func TestValidateCode_ObservationStatus_Final(t *testing.T) {
	v := NewValueSetValidator()
	result := v.ValidateCode("http://hl7.org/fhir/ValueSet/observation-status", "final", "")

	if !result.Result {
		t.Error("expected Result true for 'final' in observation-status")
	}
	if result.Display == "" {
		t.Error("expected non-empty Display for valid code")
	}
}

func TestValidateCode_ObservationStatus_Invalid(t *testing.T) {
	v := NewValueSetValidator()
	result := v.ValidateCode("http://hl7.org/fhir/ValueSet/observation-status", "bogus", "")

	if result.Result {
		t.Error("expected Result false for 'bogus' in observation-status")
	}
	if result.Message == "" {
		t.Error("expected a message for invalid code")
	}
}

func TestValidateCode_ConditionClinical_Active(t *testing.T) {
	v := NewValueSetValidator()
	result := v.ValidateCode("http://hl7.org/fhir/ValueSet/condition-clinical", "active", "")

	if !result.Result {
		t.Error("expected Result true for 'active' in condition-clinical")
	}
}

func TestValidateCode_AdminGender_Male(t *testing.T) {
	v := NewValueSetValidator()
	result := v.ValidateCode("http://hl7.org/fhir/ValueSet/administrative-gender", "male", "")

	if !result.Result {
		t.Error("expected Result true for 'male' in administrative-gender")
	}
	if result.Display != "Male" {
		t.Errorf("expected Display 'Male', got '%s'", result.Display)
	}
}

func TestValidateCode_AdminGender_Invalid(t *testing.T) {
	v := NewValueSetValidator()
	result := v.ValidateCode("http://hl7.org/fhir/ValueSet/administrative-gender", "x", "")

	if result.Result {
		t.Error("expected Result false for 'x' in administrative-gender")
	}
}

func TestValidateCode_MedRequestStatus_Completed(t *testing.T) {
	v := NewValueSetValidator()
	result := v.ValidateCode("http://hl7.org/fhir/ValueSet/medication-request-status", "completed", "")

	if !result.Result {
		t.Error("expected Result true for 'completed' in medication-request-status")
	}
}

func TestValidateCode_EncounterStatus_InProgress(t *testing.T) {
	v := NewValueSetValidator()
	result := v.ValidateCode("http://hl7.org/fhir/ValueSet/encounter-status", "in-progress", "")

	if !result.Result {
		t.Error("expected Result true for 'in-progress' in encounter-status")
	}
}

func TestValidateCode_UnknownValueSet(t *testing.T) {
	v := NewValueSetValidator()
	result := v.ValidateCode("http://hl7.org/fhir/ValueSet/does-not-exist", "active", "")

	if result.Result {
		t.Error("expected Result false for unknown ValueSet")
	}
	if !strings.Contains(result.Message, "ValueSet not found") {
		t.Errorf("expected message about ValueSet not found, got: %s", result.Message)
	}
}

func TestValidateCode_Display(t *testing.T) {
	v := NewValueSetValidator()
	result := v.ValidateCode("http://hl7.org/fhir/ValueSet/observation-status", "final", "")

	if !result.Result {
		t.Fatal("expected Result true for 'final'")
	}
	if result.Display != "Final" {
		t.Errorf("expected Display 'Final', got '%s'", result.Display)
	}
}

func TestValidateCode_AllValueSetsLoaded(t *testing.T) {
	v := NewValueSetValidator()
	sets := v.ListValueSets()

	if len(sets) < 10 {
		t.Errorf("expected at least 10 built-in value sets, got %d", len(sets))
	}

	// Verify each entry has the expected structure.
	for _, s := range sets {
		if _, ok := s["url"]; !ok {
			t.Error("expected 'url' in value set summary")
		}
		if _, ok := s["name"]; !ok {
			t.Error("expected 'name' in value set summary")
		}
		if s["resourceType"] != "ValueSet" {
			t.Errorf("expected resourceType 'ValueSet', got '%v'", s["resourceType"])
		}
	}
}

func TestValidateCode_ImmunizationStatus(t *testing.T) {
	v := NewValueSetValidator()

	// "completed" should be valid.
	result := v.ValidateCode("http://hl7.org/fhir/ValueSet/immunization-status", "completed", "")
	if !result.Result {
		t.Error("expected Result true for 'completed' in immunization-status")
	}

	// "active" should not be valid for immunization-status.
	result2 := v.ValidateCode("http://hl7.org/fhir/ValueSet/immunization-status", "active", "")
	if result2.Result {
		t.Error("expected Result false for 'active' in immunization-status")
	}
}

func TestValidateCode_CarePlanStatus_Draft(t *testing.T) {
	v := NewValueSetValidator()
	result := v.ValidateCode("http://hl7.org/fhir/ValueSet/care-plan-status", "draft", "")

	if !result.Result {
		t.Error("expected Result true for 'draft' in care-plan-status")
	}
}

func TestValidateCode_SystemFiltering(t *testing.T) {
	v := NewValueSetValidator()

	// With the correct system, should match.
	result := v.ValidateCode(
		"http://hl7.org/fhir/ValueSet/observation-status",
		"final",
		"http://hl7.org/fhir/observation-status",
	)
	if !result.Result {
		t.Error("expected Result true with matching system")
	}

	// With an incorrect system, should not match.
	result2 := v.ValidateCode(
		"http://hl7.org/fhir/ValueSet/observation-status",
		"final",
		"http://wrong-system.example.org",
	)
	if result2.Result {
		t.Error("expected Result false with non-matching system")
	}
}

func TestValidateCode_ProcedureStatus(t *testing.T) {
	v := NewValueSetValidator()
	result := v.ValidateCode("http://hl7.org/fhir/ValueSet/procedure-status", "completed", "")

	if !result.Result {
		t.Error("expected Result true for 'completed' in procedure-status")
	}
}

func TestValidateCode_DiagnosticReportStatus(t *testing.T) {
	v := NewValueSetValidator()
	result := v.ValidateCode("http://hl7.org/fhir/ValueSet/diagnostic-report-status", "final", "")

	if !result.Result {
		t.Error("expected Result true for 'final' in diagnostic-report-status")
	}
}

func TestValidateCode_AllergyIntoleranceClinical(t *testing.T) {
	v := NewValueSetValidator()

	result := v.ValidateCode("http://hl7.org/fhir/ValueSet/allergy-intolerance-clinical", "active", "")
	if !result.Result {
		t.Error("expected Result true for 'active' in allergy-intolerance-clinical")
	}

	result2 := v.ValidateCode("http://hl7.org/fhir/ValueSet/allergy-intolerance-clinical", "remission", "")
	if result2.Result {
		t.Error("expected Result false for 'remission' in allergy-intolerance-clinical")
	}
}

// =========== ValueSetValidateHandler Tests ===========

func TestValueSetValidateHandler_GET_Valid(t *testing.T) {
	v := NewValueSetValidator()
	h := NewValueSetValidateHandler(v)
	e := echo.New()

	req := httptest.NewRequest(http.MethodGet,
		"/fhir/ValueSet/$validate-code?url=http://hl7.org/fhir/ValueSet/observation-status&code=final",
		nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.ValidateCode(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if result["resourceType"] != "Parameters" {
		t.Errorf("expected resourceType Parameters, got %v", result["resourceType"])
	}

	params, ok := result["parameter"].([]interface{})
	if !ok {
		t.Fatal("expected parameter array in response")
	}

	foundResult := false
	foundDisplay := false
	for _, p := range params {
		param := p.(map[string]interface{})
		if param["name"] == "result" {
			if param["valueBoolean"] != true {
				t.Error("expected result valueBoolean to be true")
			}
			foundResult = true
		}
		if param["name"] == "display" {
			if param["valueString"] != "Final" {
				t.Errorf("expected display 'Final', got '%v'", param["valueString"])
			}
			foundDisplay = true
		}
	}
	if !foundResult {
		t.Error("expected 'result' parameter in response")
	}
	if !foundDisplay {
		t.Error("expected 'display' parameter in response")
	}
}

func TestValueSetValidateHandler_GET_Invalid(t *testing.T) {
	v := NewValueSetValidator()
	h := NewValueSetValidateHandler(v)
	e := echo.New()

	req := httptest.NewRequest(http.MethodGet,
		"/fhir/ValueSet/$validate-code?url=http://hl7.org/fhir/ValueSet/observation-status&code=bogus",
		nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.ValidateCode(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	params := result["parameter"].([]interface{})
	for _, p := range params {
		param := p.(map[string]interface{})
		if param["name"] == "result" {
			if param["valueBoolean"] != false {
				t.Error("expected result valueBoolean to be false for invalid code")
			}
		}
	}
}

func TestValueSetValidateHandler_GET_MissingURL(t *testing.T) {
	v := NewValueSetValidator()
	h := NewValueSetValidateHandler(v)
	e := echo.New()

	req := httptest.NewRequest(http.MethodGet,
		"/fhir/ValueSet/$validate-code?code=final",
		nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.ValidateCode(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for missing url, got %d", rec.Code)
	}
}

func TestValueSetValidateHandler_GET_MissingCode(t *testing.T) {
	v := NewValueSetValidator()
	h := NewValueSetValidateHandler(v)
	e := echo.New()

	req := httptest.NewRequest(http.MethodGet,
		"/fhir/ValueSet/$validate-code?url=http://hl7.org/fhir/ValueSet/observation-status",
		nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.ValidateCode(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for missing code, got %d", rec.Code)
	}
}

func TestValueSetValidateHandler_POST_Valid(t *testing.T) {
	v := NewValueSetValidator()
	h := NewValueSetValidateHandler(v)
	e := echo.New()

	body := `{
		"resourceType": "Parameters",
		"parameter": [
			{"name": "url", "valueUri": "http://hl7.org/fhir/ValueSet/administrative-gender"},
			{"name": "code", "valueCode": "female"},
			{"name": "system", "valueUri": "http://hl7.org/fhir/administrative-gender"}
		]
	}`

	req := httptest.NewRequest(http.MethodPost, "/fhir/ValueSet/$validate-code",
		strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.ValidateCodePost(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if result["resourceType"] != "Parameters" {
		t.Errorf("expected resourceType Parameters, got %v", result["resourceType"])
	}

	params := result["parameter"].([]interface{})
	foundResult := false
	for _, p := range params {
		param := p.(map[string]interface{})
		if param["name"] == "result" {
			if param["valueBoolean"] != true {
				t.Error("expected result valueBoolean to be true for valid POST")
			}
			foundResult = true
		}
	}
	if !foundResult {
		t.Error("expected 'result' parameter in POST response")
	}
}

func TestValueSetValidateHandler_POST_InvalidJSON(t *testing.T) {
	v := NewValueSetValidator()
	h := NewValueSetValidateHandler(v)
	e := echo.New()

	req := httptest.NewRequest(http.MethodPost, "/fhir/ValueSet/$validate-code",
		strings.NewReader("{not valid json"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.ValidateCodePost(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid JSON, got %d", rec.Code)
	}
}

func TestValueSetValidateHandler_RegisterRoutes(t *testing.T) {
	v := NewValueSetValidator()
	h := NewValueSetValidateHandler(v)
	e := echo.New()
	fhirGroup := e.Group("/fhir")

	h.RegisterRoutes(fhirGroup)

	routes := e.Routes()
	routePaths := make(map[string]bool)
	for _, r := range routes {
		routePaths[r.Method+":"+r.Path] = true
	}

	expected := []string{
		"GET:/fhir/ValueSet/$validate-code",
		"POST:/fhir/ValueSet/$validate-code",
	}
	for _, path := range expected {
		if !routePaths[path] {
			t.Errorf("missing expected route: %s (registered: %v)", path, routePaths)
		}
	}
}
