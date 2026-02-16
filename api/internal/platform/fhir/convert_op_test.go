package fhir

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
)

// =========== parseConvertParams Tests ===========

func TestParseConvertParams_Defaults(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/fhir/$convert", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	params := parseConvertParams(c)

	if params.InputFormat != "json" {
		t.Errorf("expected default InputFormat 'json', got '%s'", params.InputFormat)
	}
	if params.OutputFormat != "json" {
		t.Errorf("expected default OutputFormat 'json', got '%s'", params.OutputFormat)
	}
}

func TestParseConvertParams_Explicit(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/fhir/$convert?_inputFormat=xml&_outputFormat=json", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	params := parseConvertParams(c)

	if params.InputFormat != "xml" {
		t.Errorf("expected InputFormat 'xml', got '%s'", params.InputFormat)
	}
	if params.OutputFormat != "json" {
		t.Errorf("expected OutputFormat 'json', got '%s'", params.OutputFormat)
	}
}

func TestParseConvertParams_MIMETypes(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost,
		"/fhir/$convert?_inputFormat=application/fhir+json&_outputFormat=application/fhir+xml", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	params := parseConvertParams(c)

	if params.InputFormat != "json" {
		t.Errorf("expected InputFormat 'json', got '%s'", params.InputFormat)
	}
	if params.OutputFormat != "xml" {
		t.Errorf("expected OutputFormat 'xml', got '%s'", params.OutputFormat)
	}
}

// =========== normalizeConvertFormat Tests ===========

func TestNormalizeConvertFormat(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"json", "json"},
		{"xml", "xml"},
		{"application/json", "json"},
		{"application/xml", "xml"},
		{"application/fhir+json", "json"},
		{"application/fhir+xml", "xml"},
		{"unknown", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := normalizeConvertFormat(tt.input)
			if result != tt.expected {
				t.Errorf("normalizeConvertFormat(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// =========== ConvertHandler Tests ===========

func TestConvertHandler_JSONtoJSON(t *testing.T) {
	e := echo.New()
	body := `{
		"resourceType": "Patient",
		"id": "p-1",
		"name": [{"family": "Smith"}]
	}`

	req := httptest.NewRequest(http.MethodPost, "/fhir/$convert", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/fhir+json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := ConvertHandler()
	err := handler(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse response JSON: %v", err)
	}

	if result["resourceType"] != "Patient" {
		t.Errorf("expected resourceType 'Patient', got %v", result["resourceType"])
	}
	if result["id"] != "p-1" {
		t.Errorf("expected id 'p-1', got %v", result["id"])
	}
}

func TestConvertHandler_JSONtoJSON_ExplicitParams(t *testing.T) {
	e := echo.New()
	body := `{"resourceType": "Observation", "id": "obs-1", "status": "final"}`

	req := httptest.NewRequest(http.MethodPost,
		"/fhir/$convert?_inputFormat=json&_outputFormat=json", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/fhir+json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := ConvertHandler()
	err := handler(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestConvertHandler_XMLOutput_NotAcceptable(t *testing.T) {
	e := echo.New()
	body := `{"resourceType": "Patient", "id": "p-1"}`

	req := httptest.NewRequest(http.MethodPost,
		"/fhir/$convert?_outputFormat=xml", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/fhir+json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := ConvertHandler()
	err := handler(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusNotAcceptable {
		t.Errorf("expected 406, got %d", rec.Code)
	}

	var outcome map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &outcome); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if outcome["resourceType"] != "OperationOutcome" {
		t.Errorf("expected OperationOutcome, got %v", outcome["resourceType"])
	}
}

func TestConvertHandler_XMLOutput_FHIRMime(t *testing.T) {
	e := echo.New()
	body := `{"resourceType": "Patient", "id": "p-1"}`

	req := httptest.NewRequest(http.MethodPost,
		"/fhir/$convert?_outputFormat=application/fhir+xml", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/fhir+json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := ConvertHandler()
	err := handler(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusNotAcceptable {
		t.Errorf("expected 406 for FHIR XML MIME type, got %d", rec.Code)
	}
}

func TestConvertHandler_XMLInput_UnsupportedMedia(t *testing.T) {
	e := echo.New()
	body := `<Patient xmlns="http://hl7.org/fhir"><id value="p-1"/></Patient>`

	req := httptest.NewRequest(http.MethodPost,
		"/fhir/$convert?_inputFormat=xml", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/fhir+xml")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := ConvertHandler()
	err := handler(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusUnsupportedMediaType {
		t.Errorf("expected 415, got %d", rec.Code)
	}

	var outcome map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &outcome); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if outcome["resourceType"] != "OperationOutcome" {
		t.Errorf("expected OperationOutcome, got %v", outcome["resourceType"])
	}
}

func TestConvertHandler_EmptyBody(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/fhir/$convert", strings.NewReader(""))
	req.Header.Set("Content-Type", "application/fhir+json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := ConvertHandler()
	err := handler(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestConvertHandler_InvalidJSON(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/fhir/$convert", strings.NewReader("{bad json"))
	req.Header.Set("Content-Type", "application/fhir+json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := ConvertHandler()
	err := handler(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}

	var outcome map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &outcome); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	issues, ok := outcome["issue"].([]interface{})
	if !ok || len(issues) == 0 {
		t.Fatal("expected issues in response")
	}

	issue := issues[0].(map[string]interface{})
	diag, _ := issue["diagnostics"].(string)
	if !strings.Contains(diag, "Invalid JSON") {
		t.Errorf("expected 'Invalid JSON' in diagnostics, got: %s", diag)
	}
}

func TestConvertHandler_MissingResourceType(t *testing.T) {
	e := echo.New()
	body := `{"id": "123", "name": "test"}`
	req := httptest.NewRequest(http.MethodPost, "/fhir/$convert", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/fhir+json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := ConvertHandler()
	err := handler(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for missing resourceType, got %d", rec.Code)
	}

	var outcome map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &outcome); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	issues := outcome["issue"].([]interface{})
	issue := issues[0].(map[string]interface{})
	diag := issue["diagnostics"].(string)
	if !strings.Contains(diag, "resourceType") {
		t.Errorf("expected diagnostics to mention resourceType, got: %s", diag)
	}
}

func TestConvertHandler_UnknownInputFormat(t *testing.T) {
	e := echo.New()
	body := `{"resourceType": "Patient", "id": "p-1"}`
	req := httptest.NewRequest(http.MethodPost,
		"/fhir/$convert?_inputFormat=yaml", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/fhir+json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := ConvertHandler()
	err := handler(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusUnsupportedMediaType {
		t.Errorf("expected 415 for unknown input format, got %d", rec.Code)
	}
}

func TestConvertHandler_UnknownOutputFormat(t *testing.T) {
	e := echo.New()
	body := `{"resourceType": "Patient", "id": "p-1"}`
	req := httptest.NewRequest(http.MethodPost,
		"/fhir/$convert?_outputFormat=yaml", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/fhir+json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := ConvertHandler()
	err := handler(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusNotAcceptable {
		t.Errorf("expected 406 for unknown output format, got %d", rec.Code)
	}
}

func TestConvertHandler_ContentTypeHeader(t *testing.T) {
	e := echo.New()
	body := `{"resourceType": "Patient", "id": "p-1", "name": [{"family": "Test"}]}`
	req := httptest.NewRequest(http.MethodPost, "/fhir/$convert", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/fhir+json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := ConvertHandler()
	err := handler(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ct := rec.Header().Get("Content-Type")
	if !strings.Contains(ct, "json") {
		t.Errorf("expected JSON content type in response, got: %s", ct)
	}
}

func TestConvertHandler_PreservesAllFields(t *testing.T) {
	e := echo.New()
	body := `{
		"resourceType": "Patient",
		"id": "p-1",
		"name": [{"family": "Smith", "given": ["John"]}],
		"birthDate": "1990-01-15",
		"gender": "male",
		"active": true
	}`

	req := httptest.NewRequest(http.MethodPost, "/fhir/$convert", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/fhir+json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := ConvertHandler()
	err := handler(c)
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

	expectedFields := []string{"resourceType", "id", "name", "birthDate", "gender", "active"}
	for _, field := range expectedFields {
		if _, ok := result[field]; !ok {
			t.Errorf("expected field '%s' in response, but it was missing", field)
		}
	}
}
