package fhir

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
)

// =========== ConceptMapTranslator Tests ===========

func TestTranslator_SNOMED_to_ICD10(t *testing.T) {
	tr := NewConceptMapTranslator()
	resp, err := tr.Translate(&TranslateRequest{
		Code:         "73211009",
		System:       "http://snomed.info/sct",
		TargetSystem: "http://hl7.org/fhir/sid/icd-10-cm",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Result {
		t.Fatal("expected Result true for diabetes SNOMED to ICD-10")
	}
	if len(resp.Matches) == 0 {
		t.Fatal("expected at least one match")
	}
	m := resp.Matches[0]
	if m.Code != "E11.9" {
		t.Errorf("expected target code E11.9, got %s", m.Code)
	}
	if m.System != "http://hl7.org/fhir/sid/icd-10-cm" {
		t.Errorf("expected target system ICD-10, got %s", m.System)
	}
	if m.Display == "" {
		t.Error("expected non-empty display")
	}
}

func TestTranslator_ICD10_to_SNOMED(t *testing.T) {
	tr := NewConceptMapTranslator()
	resp, err := tr.Translate(&TranslateRequest{
		Code:         "E11.9",
		System:       "http://hl7.org/fhir/sid/icd-10-cm",
		TargetSystem: "http://snomed.info/sct",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Result {
		t.Fatal("expected Result true for E11.9 reverse mapping")
	}
	if len(resp.Matches) == 0 {
		t.Fatal("expected at least one match")
	}
	// E11.9 maps back to both 73211009 and 44054006
	found := false
	for _, m := range resp.Matches {
		if m.Code == "73211009" || m.Code == "44054006" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected reverse mapping to contain 73211009 or 44054006, got %+v", resp.Matches)
	}
}

func TestTranslator_LOINC_to_SNOMED(t *testing.T) {
	tr := NewConceptMapTranslator()
	resp, err := tr.Translate(&TranslateRequest{
		Code:         "2339-0",
		System:       "http://loinc.org",
		TargetSystem: "http://snomed.info/sct",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Result {
		t.Fatal("expected Result true for Glucose LOINC to SNOMED")
	}
	if len(resp.Matches) == 0 {
		t.Fatal("expected at least one match")
	}
	if resp.Matches[0].Code != "33747003" {
		t.Errorf("expected target code 33747003, got %s", resp.Matches[0].Code)
	}
}

func TestTranslator_UnknownCode(t *testing.T) {
	tr := NewConceptMapTranslator()
	resp, err := tr.Translate(&TranslateRequest{
		Code:         "99999999",
		System:       "http://snomed.info/sct",
		TargetSystem: "http://hl7.org/fhir/sid/icd-10-cm",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Result {
		t.Error("expected Result false for unknown code")
	}
	if resp.Message == "" {
		t.Error("expected a message for unknown code")
	}
}

func TestTranslator_UnknownSystem(t *testing.T) {
	tr := NewConceptMapTranslator()
	_, err := tr.Translate(&TranslateRequest{
		Code:         "12345",
		System:       "http://unknown.system/codes",
		TargetSystem: "http://snomed.info/sct",
	})
	if err == nil {
		t.Error("expected error for unknown source system")
	}
}

func TestTranslator_Hypertension(t *testing.T) {
	tr := NewConceptMapTranslator()
	resp, err := tr.Translate(&TranslateRequest{
		Code:         "38341003",
		System:       "http://snomed.info/sct",
		TargetSystem: "http://hl7.org/fhir/sid/icd-10-cm",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Result {
		t.Fatal("expected Result true for hypertension")
	}
	if resp.Matches[0].Code != "I10" {
		t.Errorf("expected I10, got %s", resp.Matches[0].Code)
	}
}

func TestTranslator_Asthma(t *testing.T) {
	tr := NewConceptMapTranslator()
	resp, err := tr.Translate(&TranslateRequest{
		Code:         "195967001",
		System:       "http://snomed.info/sct",
		TargetSystem: "http://hl7.org/fhir/sid/icd-10-cm",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Result {
		t.Fatal("expected Result true for asthma")
	}
	if resp.Matches[0].Code != "J45.909" {
		t.Errorf("expected J45.909, got %s", resp.Matches[0].Code)
	}
}

func TestTranslator_HeartFailure(t *testing.T) {
	tr := NewConceptMapTranslator()
	resp, err := tr.Translate(&TranslateRequest{
		Code:         "84114007",
		System:       "http://snomed.info/sct",
		TargetSystem: "http://hl7.org/fhir/sid/icd-10-cm",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Result {
		t.Fatal("expected Result true for heart failure")
	}
	if resp.Matches[0].Code != "I50.9" {
		t.Errorf("expected I50.9, got %s", resp.Matches[0].Code)
	}
}

func TestTranslator_HbA1c(t *testing.T) {
	tr := NewConceptMapTranslator()
	resp, err := tr.Translate(&TranslateRequest{
		Code:         "4548-4",
		System:       "http://loinc.org",
		TargetSystem: "http://snomed.info/sct",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Result {
		t.Fatal("expected Result true for HbA1c")
	}
	if resp.Matches[0].Code != "43396009" {
		t.Errorf("expected 43396009, got %s", resp.Matches[0].Code)
	}
}

func TestTranslator_MultipleBuiltinMaps(t *testing.T) {
	tr := NewConceptMapTranslator()
	maps := tr.ListConceptMaps()
	if len(maps) < 3 {
		t.Errorf("expected at least 3 built-in concept maps, got %d", len(maps))
	}
}

func TestTranslator_ListConceptMaps(t *testing.T) {
	tr := NewConceptMapTranslator()
	maps := tr.ListConceptMaps()
	if len(maps) == 0 {
		t.Fatal("expected non-empty list of concept maps")
	}

	// Each entry should have the expected structure.
	for _, m := range maps {
		if _, ok := m["id"]; !ok {
			t.Error("expected 'id' in concept map summary")
		}
		if _, ok := m["url"]; !ok {
			t.Error("expected 'url' in concept map summary")
		}
		if _, ok := m["name"]; !ok {
			t.Error("expected 'name' in concept map summary")
		}
		if _, ok := m["sourceUri"]; !ok {
			t.Error("expected 'sourceUri' in concept map summary")
		}
		if _, ok := m["targetUri"]; !ok {
			t.Error("expected 'targetUri' in concept map summary")
		}
	}
}

func TestTranslator_ByConceptMapURL(t *testing.T) {
	tr := NewConceptMapTranslator()
	resp, err := tr.Translate(&TranslateRequest{
		Code:          "73211009",
		System:        "http://snomed.info/sct",
		ConceptMapURL: "http://ehr.example.org/fhir/ConceptMap/snomed-to-icd10",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Result {
		t.Fatal("expected Result true when using ConceptMapURL")
	}
	if resp.Matches[0].Code != "E11.9" {
		t.Errorf("expected E11.9, got %s", resp.Matches[0].Code)
	}
}

func TestTranslator_Equivalence(t *testing.T) {
	tr := NewConceptMapTranslator()
	resp, err := tr.Translate(&TranslateRequest{
		Code:         "73211009",
		System:       "http://snomed.info/sct",
		TargetSystem: "http://hl7.org/fhir/sid/icd-10-cm",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Result || len(resp.Matches) == 0 {
		t.Fatal("expected a match")
	}
	if resp.Matches[0].Equivalence != "equivalent" {
		t.Errorf("expected equivalence 'equivalent', got '%s'", resp.Matches[0].Equivalence)
	}
}

func TestTranslator_CaseSensitiveCode(t *testing.T) {
	tr := NewConceptMapTranslator()
	// SNOMED codes are numeric, but checking that "e11.9" (lowercase) does not match "E11.9".
	resp, err := tr.Translate(&TranslateRequest{
		Code:         "e11.9",
		System:       "http://hl7.org/fhir/sid/icd-10-cm",
		TargetSystem: "http://snomed.info/sct",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Result {
		t.Error("expected Result false for case-mismatched code 'e11.9'")
	}
}

// =========== TranslateHandler Tests ===========

func TestTranslateHandler_GET_ValidTranslation(t *testing.T) {
	tr := NewConceptMapTranslator()
	h := NewTranslateHandler(tr)
	e := echo.New()

	req := httptest.NewRequest(http.MethodGet,
		"/fhir/ConceptMap/$translate?code=73211009&system=http://snomed.info/sct&targetsystem=http://hl7.org/fhir/sid/icd-10-cm",
		nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.Translate(c)
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

	// Check for result=true and a match.
	foundResult := false
	foundMatch := false
	for _, p := range params {
		param := p.(map[string]interface{})
		if param["name"] == "result" {
			if param["valueBoolean"] != true {
				t.Error("expected result valueBoolean to be true")
			}
			foundResult = true
		}
		if param["name"] == "match" {
			foundMatch = true
		}
	}
	if !foundResult {
		t.Error("expected 'result' parameter in response")
	}
	if !foundMatch {
		t.Error("expected 'match' parameter in response")
	}
}

func TestTranslateHandler_GET_MissingCode(t *testing.T) {
	tr := NewConceptMapTranslator()
	h := NewTranslateHandler(tr)
	e := echo.New()

	req := httptest.NewRequest(http.MethodGet,
		"/fhir/ConceptMap/$translate?system=http://snomed.info/sct&targetsystem=http://hl7.org/fhir/sid/icd-10-cm",
		nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.Translate(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for missing code, got %d", rec.Code)
	}
}

func TestTranslateHandler_GET_MissingSystem(t *testing.T) {
	tr := NewConceptMapTranslator()
	h := NewTranslateHandler(tr)
	e := echo.New()

	req := httptest.NewRequest(http.MethodGet,
		"/fhir/ConceptMap/$translate?code=73211009&targetsystem=http://hl7.org/fhir/sid/icd-10-cm",
		nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.Translate(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for missing system, got %d", rec.Code)
	}
}

func TestTranslateHandler_GET_NoMapping(t *testing.T) {
	tr := NewConceptMapTranslator()
	h := NewTranslateHandler(tr)
	e := echo.New()

	req := httptest.NewRequest(http.MethodGet,
		"/fhir/ConceptMap/$translate?code=99999999&system=http://snomed.info/sct&targetsystem=http://hl7.org/fhir/sid/icd-10-cm",
		nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.Translate(c)
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
				t.Error("expected result valueBoolean to be false for unknown code")
			}
		}
	}
}

func TestTranslateHandler_POST_ValidTranslation(t *testing.T) {
	tr := NewConceptMapTranslator()
	h := NewTranslateHandler(tr)
	e := echo.New()

	body := `{
		"resourceType": "Parameters",
		"parameter": [
			{"name": "code", "valueCode": "73211009"},
			{"name": "system", "valueUri": "http://snomed.info/sct"},
			{"name": "targetsystem", "valueUri": "http://hl7.org/fhir/sid/icd-10-cm"}
		]
	}`

	req := httptest.NewRequest(http.MethodPost, "/fhir/ConceptMap/$translate",
		strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.TranslatePost(c)
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
	foundMatch := false
	for _, p := range params {
		param := p.(map[string]interface{})
		if param["name"] == "match" {
			foundMatch = true
		}
	}
	if !foundMatch {
		t.Error("expected 'match' parameter in POST response")
	}
}

func TestTranslateHandler_POST_InvalidJSON(t *testing.T) {
	tr := NewConceptMapTranslator()
	h := NewTranslateHandler(tr)
	e := echo.New()

	req := httptest.NewRequest(http.MethodPost, "/fhir/ConceptMap/$translate",
		strings.NewReader("{not valid json"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.TranslatePost(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid JSON, got %d", rec.Code)
	}
}

func TestTranslateHandler_ListConceptMaps(t *testing.T) {
	tr := NewConceptMapTranslator()
	h := NewTranslateHandler(tr)
	e := echo.New()

	req := httptest.NewRequest(http.MethodGet, "/fhir/ConceptMap", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.ListConceptMaps(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var bundle map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &bundle); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if bundle["resourceType"] != "Bundle" {
		t.Errorf("expected resourceType Bundle, got %v", bundle["resourceType"])
	}
	if bundle["type"] != "searchset" {
		t.Errorf("expected type searchset, got %v", bundle["type"])
	}

	entries, ok := bundle["entry"].([]interface{})
	if !ok {
		t.Fatal("expected entry array in bundle")
	}
	if len(entries) < 3 {
		t.Errorf("expected at least 3 entries, got %d", len(entries))
	}
}

func TestTranslateHandler_TranslateByMapID(t *testing.T) {
	tr := NewConceptMapTranslator()
	h := NewTranslateHandler(tr)
	e := echo.New()

	req := httptest.NewRequest(http.MethodGet,
		"/fhir/ConceptMap/snomed-to-icd10/$translate?code=38341003&system=http://snomed.info/sct",
		nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("snomed-to-icd10")

	err := h.TranslateByMap(c)
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
	foundMatch := false
	for _, p := range params {
		param := p.(map[string]interface{})
		if param["name"] == "result" {
			if param["valueBoolean"] != true {
				t.Error("expected result true for known mapping by map ID")
			}
		}
		if param["name"] == "match" {
			foundMatch = true
		}
	}
	if !foundMatch {
		t.Error("expected match in response when translating by map ID")
	}
}
