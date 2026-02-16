package fhir

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
)

// =========== ParseLastNParams Tests ===========

func TestParseLastNParams_AllParams(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/fhir/Observation/$lastn?patient=Patient/123&category=vital-signs&code=8480-6&max=3", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	params := ParseLastNParams(c)

	if params.Patient != "Patient/123" {
		t.Errorf("expected patient 'Patient/123', got %q", params.Patient)
	}
	if params.Category != "vital-signs" {
		t.Errorf("expected category 'vital-signs', got %q", params.Category)
	}
	if params.Code != "8480-6" {
		t.Errorf("expected code '8480-6', got %q", params.Code)
	}
	if params.Max != 3 {
		t.Errorf("expected max 3, got %d", params.Max)
	}
}

func TestParseLastNParams_DefaultMax(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/fhir/Observation/$lastn?patient=Patient/123", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	params := ParseLastNParams(c)

	if params.Max != 1 {
		t.Errorf("expected default max 1, got %d", params.Max)
	}
}

func TestParseLastNParams_InvalidMaxFallsBackToDefault(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/fhir/Observation/$lastn?patient=Patient/123&max=abc", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	params := ParseLastNParams(c)

	if params.Max != 1 {
		t.Errorf("expected max to fall back to 1 for invalid value, got %d", params.Max)
	}
}

func TestParseLastNParams_ZeroMaxFallsBackToDefault(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/fhir/Observation/$lastn?patient=Patient/123&max=0", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	params := ParseLastNParams(c)

	if params.Max != 1 {
		t.Errorf("expected max to fall back to 1 for zero value, got %d", params.Max)
	}
}

func TestParseLastNParams_EmptyParams(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/fhir/Observation/$lastn", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	params := ParseLastNParams(c)

	if params.Patient != "" {
		t.Errorf("expected empty patient, got %q", params.Patient)
	}
	if params.Category != "" {
		t.Errorf("expected empty category, got %q", params.Category)
	}
	if params.Code != "" {
		t.Errorf("expected empty code, got %q", params.Code)
	}
	if params.Max != 1 {
		t.Errorf("expected default max 1, got %d", params.Max)
	}
}

// =========== LastNHandler Tests ===========

func mockLastNExecutor(results []map[string]interface{}, err error) LastNExecutor {
	return func(ctx context.Context, params LastNParams) ([]map[string]interface{}, error) {
		return results, err
	}
}

func TestLastNHandler_ReturnsBundle(t *testing.T) {
	observations := []map[string]interface{}{
		{
			"resourceType": "Observation",
			"id":           "obs-1",
			"status":       "final",
			"code": map[string]interface{}{
				"coding": []interface{}{
					map[string]interface{}{"system": "http://loinc.org", "code": "8480-6"},
				},
			},
			"valueQuantity": map[string]interface{}{
				"value": 120.0,
				"unit":  "mmHg",
			},
		},
		{
			"resourceType": "Observation",
			"id":           "obs-2",
			"status":       "final",
			"code": map[string]interface{}{
				"coding": []interface{}{
					map[string]interface{}{"system": "http://loinc.org", "code": "8462-4"},
				},
			},
			"valueQuantity": map[string]interface{}{
				"value": 80.0,
				"unit":  "mmHg",
			},
		},
	}

	handler := LastNHandler(mockLastNExecutor(observations, nil))
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/fhir/Observation/$lastn?patient=Patient/123&max=1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler(c)
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
		t.Errorf("expected resourceType 'Bundle', got %v", bundle["resourceType"])
	}
	if bundle["type"] != "searchset" {
		t.Errorf("expected type 'searchset', got %v", bundle["type"])
	}

	// total should be 2
	total, ok := bundle["total"].(float64)
	if !ok {
		t.Fatal("expected total to be a number")
	}
	if int(total) != 2 {
		t.Errorf("expected total 2, got %v", total)
	}

	entries, ok := bundle["entry"].([]interface{})
	if !ok {
		t.Fatal("expected entry array in bundle")
	}
	if len(entries) != 2 {
		t.Errorf("expected 2 entries, got %d", len(entries))
	}

	// Verify first entry has fullUrl and resource
	firstEntry := entries[0].(map[string]interface{})
	if firstEntry["fullUrl"] != "Observation/obs-1" {
		t.Errorf("expected first entry fullUrl 'Observation/obs-1', got %v", firstEntry["fullUrl"])
	}

	search, ok := firstEntry["search"].(map[string]interface{})
	if !ok {
		t.Fatal("expected search object in entry")
	}
	if search["mode"] != "match" {
		t.Errorf("expected search mode 'match', got %v", search["mode"])
	}
}

func TestLastNHandler_RequiresPatient(t *testing.T) {
	handler := LastNHandler(mockLastNExecutor(nil, nil))
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/fhir/Observation/$lastn", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for missing patient, got %d", rec.Code)
	}

	var outcome map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &outcome); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if outcome["resourceType"] != "OperationOutcome" {
		t.Errorf("expected OperationOutcome, got %v", outcome["resourceType"])
	}

	issues, ok := outcome["issue"].([]interface{})
	if !ok || len(issues) == 0 {
		t.Fatal("expected issues in OperationOutcome")
	}

	issue := issues[0].(map[string]interface{})
	if issue["severity"] != "error" {
		t.Errorf("expected severity 'error', got %v", issue["severity"])
	}
	if issue["code"] != "required" {
		t.Errorf("expected code 'required', got %v", issue["code"])
	}
}

func TestLastNHandler_EmptyResults(t *testing.T) {
	handler := LastNHandler(mockLastNExecutor([]map[string]interface{}{}, nil))
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/fhir/Observation/$lastn?patient=Patient/123", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler(c)
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

	total, ok := bundle["total"].(float64)
	if !ok {
		t.Fatal("expected total to be a number")
	}
	if int(total) != 0 {
		t.Errorf("expected total 0, got %v", total)
	}

	entries, ok := bundle["entry"].([]interface{})
	if !ok {
		t.Fatal("expected entry array in bundle")
	}
	if len(entries) != 0 {
		t.Errorf("expected 0 entries, got %d", len(entries))
	}
}

func TestLastNHandler_ExecutorError(t *testing.T) {
	handler := LastNHandler(mockLastNExecutor(nil, fmt.Errorf("database error")))
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/fhir/Observation/$lastn?patient=Patient/123", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rec.Code)
	}

	var outcome map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &outcome); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if outcome["resourceType"] != "OperationOutcome" {
		t.Errorf("expected OperationOutcome, got %v", outcome["resourceType"])
	}
}

func TestLastNHandler_BundleHasTimestamp(t *testing.T) {
	observations := []map[string]interface{}{
		{
			"resourceType": "Observation",
			"id":           "obs-1",
			"status":       "final",
			"code":         map[string]interface{}{"text": "BP"},
		},
	}

	handler := LastNHandler(mockLastNExecutor(observations, nil))
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/fhir/Observation/$lastn?patient=Patient/123", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var bundle map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &bundle); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	ts, ok := bundle["timestamp"].(string)
	if !ok || ts == "" {
		t.Error("expected Bundle to have a non-empty timestamp")
	}
}
