package fhir

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
)

// =========== ParseStatsParams Tests ===========

func TestParseStatsParams_AllParams(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet,
		"/fhir/Observation/$stats?patient=Patient/123&code=http://loinc.org|8480-6&system=http://loinc.org&period=ge2024-01-01&statistic=min,max,mean",
		nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	params := ParseStatsParams(c)

	if params.Patient != "Patient/123" {
		t.Errorf("expected patient 'Patient/123', got %q", params.Patient)
	}
	if params.Code != "http://loinc.org|8480-6" {
		t.Errorf("expected code 'http://loinc.org|8480-6', got %q", params.Code)
	}
	if params.System != "http://loinc.org" {
		t.Errorf("expected system 'http://loinc.org', got %q", params.System)
	}
	if params.Period != "ge2024-01-01" {
		t.Errorf("expected period 'ge2024-01-01', got %q", params.Period)
	}
	if len(params.Statistic) != 3 {
		t.Fatalf("expected 3 statistics, got %d", len(params.Statistic))
	}
	expectedStats := []string{"min", "max", "mean"}
	for i, expected := range expectedStats {
		if params.Statistic[i] != expected {
			t.Errorf("expected statistic[%d] %q, got %q", i, expected, params.Statistic[i])
		}
	}
}

func TestParseStatsParams_EmptyParams(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/fhir/Observation/$stats", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	params := ParseStatsParams(c)

	if params.Patient != "" {
		t.Errorf("expected empty patient, got %q", params.Patient)
	}
	if params.Code != "" {
		t.Errorf("expected empty code, got %q", params.Code)
	}
	if params.System != "" {
		t.Errorf("expected empty system, got %q", params.System)
	}
	if params.Period != "" {
		t.Errorf("expected empty period, got %q", params.Period)
	}
	if len(params.Statistic) != 0 {
		t.Errorf("expected empty statistic slice, got %v", params.Statistic)
	}
}

func TestParseStatsParams_SingleStatistic(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/fhir/Observation/$stats?patient=Patient/1&code=8480-6&statistic=count", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	params := ParseStatsParams(c)

	if len(params.Statistic) != 1 {
		t.Fatalf("expected 1 statistic, got %d", len(params.Statistic))
	}
	if params.Statistic[0] != "count" {
		t.Errorf("expected statistic 'count', got %q", params.Statistic[0])
	}
}

// =========== StatsHandler Tests ===========

func mockStatsExecutor(result *StatsResult, err error) StatsExecutor {
	return func(ctx context.Context, params StatsParams) (*StatsResult, error) {
		return result, err
	}
}

func float64Ptr(v float64) *float64 {
	return &v
}

func TestStatsHandler_ReturnsParameters(t *testing.T) {
	result := &StatsResult{
		Code:    "8480-6",
		Subject: "Patient/123",
		Period:  "2024-01-01 to 2024-12-31",
		Count:   10,
		Min:     float64Ptr(90.0),
		Max:     float64Ptr(140.0),
		Mean:    float64Ptr(115.5),
		Median:  float64Ptr(116.0),
		StdDev:  float64Ptr(12.3),
		Sum:     float64Ptr(1155.0),
	}

	handler := StatsHandler(mockStatsExecutor(result, nil))
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet,
		"/fhir/Observation/$stats?patient=Patient/123&code=8480-6&statistic=min,max,mean,median,stddev,sum",
		nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var params map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &params); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if params["resourceType"] != "Parameters" {
		t.Errorf("expected resourceType 'Parameters', got %v", params["resourceType"])
	}

	paramList, ok := params["parameter"].([]interface{})
	if !ok {
		t.Fatal("expected parameter array")
	}

	// Should have: code, subject, count, period, min, max, mean, median, stddev, sum = 10 params
	if len(paramList) != 10 {
		t.Errorf("expected 10 parameters, got %d", len(paramList))
	}

	// Verify specific parameters exist.
	paramMap := make(map[string]interface{})
	for _, p := range paramList {
		pm := p.(map[string]interface{})
		name := pm["name"].(string)
		paramMap[name] = pm
	}

	expectedNames := []string{"code", "subject", "count", "period", "min", "max", "mean", "median", "stddev", "sum"}
	for _, name := range expectedNames {
		if _, ok := paramMap[name]; !ok {
			t.Errorf("expected parameter %q not found in response", name)
		}
	}

	// Verify code value.
	codeParam := paramMap["code"].(map[string]interface{})
	if codeParam["valueString"] != "8480-6" {
		t.Errorf("expected code valueString '8480-6', got %v", codeParam["valueString"])
	}

	// Verify count value.
	countParam := paramMap["count"].(map[string]interface{})
	countVal, ok := countParam["valueInteger"].(float64)
	if !ok {
		t.Fatal("expected count valueInteger to be a number")
	}
	if int(countVal) != 10 {
		t.Errorf("expected count 10, got %v", countVal)
	}

	// Verify a decimal value.
	meanParam := paramMap["mean"].(map[string]interface{})
	meanVal, ok := meanParam["valueDecimal"].(float64)
	if !ok {
		t.Fatal("expected mean valueDecimal to be a number")
	}
	if meanVal != 115.5 {
		t.Errorf("expected mean 115.5, got %v", meanVal)
	}
}

func TestStatsHandler_RequiresPatient(t *testing.T) {
	handler := StatsHandler(mockStatsExecutor(nil, nil))
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/fhir/Observation/$stats?code=8480-6", nil)
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

	issues := outcome["issue"].([]interface{})
	issue := issues[0].(map[string]interface{})
	diag := issue["diagnostics"].(string)
	if !strings.Contains(diag, "patient") {
		t.Errorf("expected diagnostics to mention 'patient', got %q", diag)
	}
}

func TestStatsHandler_RequiresCode(t *testing.T) {
	handler := StatsHandler(mockStatsExecutor(nil, nil))
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/fhir/Observation/$stats?patient=Patient/123", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for missing code, got %d", rec.Code)
	}

	var outcome map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &outcome); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if outcome["resourceType"] != "OperationOutcome" {
		t.Errorf("expected OperationOutcome, got %v", outcome["resourceType"])
	}

	issues := outcome["issue"].([]interface{})
	issue := issues[0].(map[string]interface{})
	diag := issue["diagnostics"].(string)
	if !strings.Contains(diag, "code") {
		t.Errorf("expected diagnostics to mention 'code', got %q", diag)
	}
}

func TestStatsHandler_RequiresBothPatientAndCode(t *testing.T) {
	handler := StatsHandler(mockStatsExecutor(nil, nil))
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/fhir/Observation/$stats", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for missing patient and code, got %d", rec.Code)
	}
}

func TestStatsHandler_ExecutorError(t *testing.T) {
	handler := StatsHandler(mockStatsExecutor(nil, fmt.Errorf("computation error")))
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/fhir/Observation/$stats?patient=Patient/123&code=8480-6", nil)
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

func TestStatsHandler_PartialStats(t *testing.T) {
	// Only count and mean, no other stats.
	result := &StatsResult{
		Code:    "8480-6",
		Subject: "Patient/123",
		Count:   5,
		Mean:    float64Ptr(110.0),
	}

	handler := StatsHandler(mockStatsExecutor(result, nil))
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet,
		"/fhir/Observation/$stats?patient=Patient/123&code=8480-6&statistic=count,mean",
		nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var params map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &params); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	paramList := params["parameter"].([]interface{})

	// Should have: code, subject, count, mean = 4 params (no period, no min/max/median/stddev/sum)
	if len(paramList) != 4 {
		t.Errorf("expected 4 parameters for partial stats, got %d", len(paramList))
	}

	// Verify omitted stats are not present.
	paramNames := make(map[string]bool)
	for _, p := range paramList {
		pm := p.(map[string]interface{})
		paramNames[pm["name"].(string)] = true
	}

	omittedStats := []string{"min", "max", "median", "stddev", "sum", "period"}
	for _, name := range omittedStats {
		if paramNames[name] {
			t.Errorf("expected parameter %q to be omitted for partial stats", name)
		}
	}
}
