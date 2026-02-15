package fhir

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
)

// =========== SubsumptionChecker Tests ===========

func TestSubsumes_Equivalent(t *testing.T) {
	checker := NewSubsumptionChecker()
	result, err := checker.CheckSubsumption(systemSNOMED, "73211009", "73211009")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != Equivalent {
		t.Errorf("expected equivalent, got %s", result)
	}
}

func TestSubsumes_SNOMED_DiabetesSubsumesType2(t *testing.T) {
	checker := NewSubsumptionChecker()
	result, err := checker.CheckSubsumption(systemSNOMED, "73211009", "44054006")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != Subsumes {
		t.Errorf("expected subsumes, got %s", result)
	}
}

func TestSubsumes_SNOMED_Type2SubsumedByDiabetes(t *testing.T) {
	checker := NewSubsumptionChecker()
	result, err := checker.CheckSubsumption(systemSNOMED, "44054006", "73211009")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != SubsumedBy {
		t.Errorf("expected subsumed-by, got %s", result)
	}
}

func TestSubsumes_SNOMED_TransitiveAncestor(t *testing.T) {
	checker := NewSubsumptionChecker()
	// Diabetes mellitus (73211009) should subsume Type 2 diabetes with renal
	// complications (313436004) through the intermediate Type 2 (44054006).
	result, err := checker.CheckSubsumption(systemSNOMED, "73211009", "313436004")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != Subsumes {
		t.Errorf("expected subsumes for transitive ancestor, got %s", result)
	}
}

func TestSubsumes_SNOMED_NotSubsumed(t *testing.T) {
	checker := NewSubsumptionChecker()
	// Asthma (195967001) and Diabetes (73211009) are in different branches.
	result, err := checker.CheckSubsumption(systemSNOMED, "195967001", "73211009")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != NotSubsumed {
		t.Errorf("expected not-subsumed, got %s", result)
	}
}

func TestSubsumes_SNOMED_HeartFailureSubsumesLeft(t *testing.T) {
	checker := NewSubsumptionChecker()
	// Heart failure (84114007) subsumes Left heart failure (85232009).
	result, err := checker.CheckSubsumption(systemSNOMED, "84114007", "85232009")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != Subsumes {
		t.Errorf("expected subsumes, got %s", result)
	}
}

func TestSubsumes_SNOMED_SiblingNotSubsumed(t *testing.T) {
	checker := NewSubsumptionChecker()
	// Type 1 (46635009) and Type 2 (44054006) are siblings, not related.
	result, err := checker.CheckSubsumption(systemSNOMED, "46635009", "44054006")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != NotSubsumed {
		t.Errorf("expected not-subsumed for siblings, got %s", result)
	}
}

func TestSubsumes_ICD10_PrefixSubsumes(t *testing.T) {
	checker := NewSubsumptionChecker()
	// E11 subsumes E11.9 via prefix matching.
	result, err := checker.CheckSubsumption(systemICD10, "E11", "E11.9")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != Subsumes {
		t.Errorf("expected subsumes for E11 -> E11.9, got %s", result)
	}
}

func TestSubsumes_ICD10_LongerPrefixSubsumes(t *testing.T) {
	checker := NewSubsumptionChecker()
	// E11.6 subsumes E11.65 via prefix matching.
	result, err := checker.CheckSubsumption(systemICD10, "E11.6", "E11.65")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != Subsumes {
		t.Errorf("expected subsumes for E11.6 -> E11.65, got %s", result)
	}
}

func TestSubsumes_ICD10_NotSubsumed(t *testing.T) {
	checker := NewSubsumptionChecker()
	// E11 and I10 are in different chapters; no relationship.
	result, err := checker.CheckSubsumption(systemICD10, "E11", "I10")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != NotSubsumed {
		t.Errorf("expected not-subsumed for E11 and I10, got %s", result)
	}
}

func TestSubsumes_UnknownSystem(t *testing.T) {
	checker := NewSubsumptionChecker()
	_, err := checker.CheckSubsumption("http://unknown.system/codes", "A", "B")
	if err == nil {
		t.Error("expected error for unknown code system")
	}
}

func TestSubsumes_UnknownCode(t *testing.T) {
	checker := NewSubsumptionChecker()
	// Unknown codes within a known system should return not-subsumed, not error.
	result, err := checker.CheckSubsumption(systemSNOMED, "99999999", "88888888")
	if err != nil {
		t.Fatalf("unexpected error for unknown codes: %v", err)
	}
	if result != NotSubsumed {
		t.Errorf("expected not-subsumed for unknown codes, got %s", result)
	}
}

// =========== SubsumesHandler Tests ===========

func TestSubsumesHandler_GET_Subsumes(t *testing.T) {
	checker := NewSubsumptionChecker()
	h := NewSubsumesHandler(checker)
	e := echo.New()

	req := httptest.NewRequest(http.MethodGet,
		"/fhir/CodeSystem/$subsumes?system=http://snomed.info/sct&codeA=73211009&codeB=44054006",
		nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.HandleSubsumes(c)
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

	foundOutcome := false
	for _, p := range params {
		param := p.(map[string]interface{})
		if param["name"] == "outcome" {
			foundOutcome = true
			if param["valueCode"] != "subsumes" {
				t.Errorf("expected outcome 'subsumes', got '%v'", param["valueCode"])
			}
		}
	}
	if !foundOutcome {
		t.Error("expected 'outcome' parameter in response")
	}
}

func TestSubsumesHandler_GET_MissingParams(t *testing.T) {
	checker := NewSubsumptionChecker()
	h := NewSubsumesHandler(checker)
	e := echo.New()

	tests := []struct {
		name string
		url  string
	}{
		{"missing system", "/fhir/CodeSystem/$subsumes?codeA=1&codeB=2"},
		{"missing codeA", "/fhir/CodeSystem/$subsumes?system=http://snomed.info/sct&codeB=2"},
		{"missing codeB", "/fhir/CodeSystem/$subsumes?system=http://snomed.info/sct&codeA=1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err := h.HandleSubsumes(c)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if rec.Code != http.StatusBadRequest {
				t.Errorf("expected 400 for %s, got %d", tt.name, rec.Code)
			}
		})
	}
}

func TestSubsumesHandler_POST_Valid(t *testing.T) {
	checker := NewSubsumptionChecker()
	h := NewSubsumesHandler(checker)
	e := echo.New()

	body := `{
		"resourceType": "Parameters",
		"parameter": [
			{"name": "system", "valueUri": "http://snomed.info/sct"},
			{"name": "codeA", "valueCode": "73211009"},
			{"name": "codeB", "valueCode": "44054006"}
		]
	}`

	req := httptest.NewRequest(http.MethodPost, "/fhir/CodeSystem/$subsumes",
		strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.HandleSubsumesPost(c)
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
	foundOutcome := false
	for _, p := range params {
		param := p.(map[string]interface{})
		if param["name"] == "outcome" {
			foundOutcome = true
			if param["valueCode"] != "subsumes" {
				t.Errorf("expected outcome 'subsumes', got '%v'", param["valueCode"])
			}
		}
	}
	if !foundOutcome {
		t.Error("expected 'outcome' parameter in POST response")
	}
}

func TestSubsumesHandler_POST_InvalidJSON(t *testing.T) {
	checker := NewSubsumptionChecker()
	h := NewSubsumesHandler(checker)
	e := echo.New()

	req := httptest.NewRequest(http.MethodPost, "/fhir/CodeSystem/$subsumes",
		strings.NewReader("{not valid json"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.HandleSubsumesPost(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid JSON, got %d", rec.Code)
	}
}

func TestSubsumesHandler_GET_Equivalent(t *testing.T) {
	checker := NewSubsumptionChecker()
	h := NewSubsumesHandler(checker)
	e := echo.New()

	req := httptest.NewRequest(http.MethodGet,
		"/fhir/CodeSystem/$subsumes?system=http://snomed.info/sct&codeA=73211009&codeB=73211009",
		nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.HandleSubsumes(c)
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
		if param["name"] == "outcome" {
			if param["valueCode"] != "equivalent" {
				t.Errorf("expected outcome 'equivalent', got '%v'", param["valueCode"])
			}
		}
	}
}

func TestSubsumesHandler_GET_UnknownSystem(t *testing.T) {
	checker := NewSubsumptionChecker()
	h := NewSubsumesHandler(checker)
	e := echo.New()

	req := httptest.NewRequest(http.MethodGet,
		"/fhir/CodeSystem/$subsumes?system=http://unknown.system&codeA=A&codeB=B",
		nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.HandleSubsumes(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for unknown system, got %d", rec.Code)
	}
}
