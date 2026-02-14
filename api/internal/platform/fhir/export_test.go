package fhir

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
)

// =========== ExportManager Tests ===========

func TestExportManager_KickOff(t *testing.T) {
	mgr := NewExportManager()

	job := mgr.KickOff([]string{"Patient", "Observation"}, nil)

	if job.ID == "" {
		t.Error("expected job ID to be set")
	}
	if job.Status != "complete" {
		t.Errorf("expected status 'complete', got %q", job.Status)
	}
	if len(job.ResourceTypes) != 2 {
		t.Errorf("expected 2 resource types, got %d", len(job.ResourceTypes))
	}
	if job.ResourceTypes[0] != "Patient" {
		t.Errorf("expected first type 'Patient', got %q", job.ResourceTypes[0])
	}
	if job.OutputFormat != "application/fhir+ndjson" {
		t.Errorf("expected output format 'application/fhir+ndjson', got %q", job.OutputFormat)
	}
	if len(job.OutputFiles) != 2 {
		t.Errorf("expected 2 output files, got %d", len(job.OutputFiles))
	}
	if job.CompletedAt == nil {
		t.Error("expected CompletedAt to be set")
	}
	if job.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}
}

func TestExportManager_KickOff_DefaultTypes(t *testing.T) {
	mgr := NewExportManager()

	job := mgr.KickOff(nil, nil)

	if len(job.ResourceTypes) != 5 {
		t.Errorf("expected 5 default resource types, got %d", len(job.ResourceTypes))
	}
}

func TestExportManager_KickOff_WithSince(t *testing.T) {
	mgr := NewExportManager()

	since := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	job := mgr.KickOff([]string{"Patient"}, &since)

	if job.Since == nil {
		t.Fatal("expected Since to be set")
	}
	if !job.Since.Equal(since) {
		t.Errorf("expected Since %v, got %v", since, *job.Since)
	}
}

func TestExportManager_GetStatus(t *testing.T) {
	mgr := NewExportManager()

	job := mgr.KickOff([]string{"Patient"}, nil)

	retrieved, err := mgr.GetStatus(job.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if retrieved.ID != job.ID {
		t.Errorf("expected job ID %q, got %q", job.ID, retrieved.ID)
	}
	if retrieved.Status != "complete" {
		t.Errorf("expected status 'complete', got %q", retrieved.Status)
	}
}

func TestExportManager_GetStatus_NotFound(t *testing.T) {
	mgr := NewExportManager()

	_, err := mgr.GetStatus("nonexistent-id")
	if err == nil {
		t.Error("expected error for nonexistent job")
	}
}

func TestExportManager_GetJobData(t *testing.T) {
	mgr := NewExportManager()

	job := mgr.KickOff([]string{"Patient"}, nil)

	data, err := mgr.GetJobData(job.ID, "Patient")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(data) == 0 {
		t.Error("expected non-empty data")
	}

	// Parse the NDJSON line
	var resource map[string]interface{}
	if err := json.Unmarshal(data[:len(data)-1], &resource); err != nil {
		t.Fatalf("failed to parse NDJSON: %v", err)
	}
	if resource["resourceType"] != "Patient" {
		t.Errorf("expected resourceType 'Patient', got %v", resource["resourceType"])
	}
}

func TestExportManager_GetJobData_NotFound(t *testing.T) {
	mgr := NewExportManager()

	_, err := mgr.GetJobData("nonexistent-id", "Patient")
	if err == nil {
		t.Error("expected error for nonexistent job")
	}
}

func TestExportManager_GetJobData_WrongType(t *testing.T) {
	mgr := NewExportManager()

	job := mgr.KickOff([]string{"Patient"}, nil)

	_, err := mgr.GetJobData(job.ID, "Observation")
	if err == nil {
		t.Error("expected error for wrong file type")
	}
}

// =========== ExportHandler Tests ===========

func TestExportHandler_SystemExport(t *testing.T) {
	mgr := NewExportManager()
	h := NewExportHandler(mgr)
	e := echo.New()

	req := httptest.NewRequest(http.MethodPost, "/fhir/$export?_type=Patient,Observation", nil)
	req.Header.Set("Prefer", "respond-async")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.SystemExport(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusAccepted {
		t.Errorf("expected 202, got %d", rec.Code)
	}
	contentLocation := rec.Header().Get("Content-Location")
	if contentLocation == "" {
		t.Error("expected Content-Location header")
	}
}

func TestExportHandler_SystemExport_NoTypes(t *testing.T) {
	mgr := NewExportManager()
	h := NewExportHandler(mgr)
	e := echo.New()

	req := httptest.NewRequest(http.MethodPost, "/fhir/$export", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.SystemExport(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusAccepted {
		t.Errorf("expected 202, got %d", rec.Code)
	}
}

func TestExportHandler_SystemExport_WithSince(t *testing.T) {
	mgr := NewExportManager()
	h := NewExportHandler(mgr)
	e := echo.New()

	req := httptest.NewRequest(http.MethodPost, "/fhir/$export?_since=2024-01-01T00:00:00Z", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.SystemExport(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusAccepted {
		t.Errorf("expected 202, got %d", rec.Code)
	}
}

func TestExportHandler_SystemExport_InvalidSince(t *testing.T) {
	mgr := NewExportManager()
	h := NewExportHandler(mgr)
	e := echo.New()

	req := httptest.NewRequest(http.MethodPost, "/fhir/$export?_since=not-a-date", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.SystemExport(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestExportHandler_SystemExport_BadPrefer(t *testing.T) {
	mgr := NewExportManager()
	h := NewExportHandler(mgr)
	e := echo.New()

	req := httptest.NewRequest(http.MethodPost, "/fhir/$export", nil)
	req.Header.Set("Prefer", "return=representation")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.SystemExport(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for bad Prefer header, got %d", rec.Code)
	}
}

func TestExportHandler_PatientExport(t *testing.T) {
	mgr := NewExportManager()
	h := NewExportHandler(mgr)
	e := echo.New()

	req := httptest.NewRequest(http.MethodPost, "/fhir/Patient/$export?_type=Patient", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.PatientExport(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusAccepted {
		t.Errorf("expected 202, got %d", rec.Code)
	}
}

func TestExportHandler_ExportStatus_Complete(t *testing.T) {
	mgr := NewExportManager()
	h := NewExportHandler(mgr)
	e := echo.New()

	job := mgr.KickOff([]string{"Patient"}, nil)

	req := httptest.NewRequest(http.MethodGet, "/fhir/$export-status/"+job.ID, nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(job.ID)

	err := h.ExportStatus(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var result map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &result)
	if result["transactionTime"] == nil {
		t.Error("expected transactionTime in response")
	}
	output, ok := result["output"].([]interface{})
	if !ok {
		t.Fatal("expected output array in response")
	}
	if len(output) != 1 {
		t.Errorf("expected 1 output file, got %d", len(output))
	}
}

func TestExportHandler_ExportStatus_NotFound(t *testing.T) {
	mgr := NewExportManager()
	h := NewExportHandler(mgr)
	e := echo.New()

	req := httptest.NewRequest(http.MethodGet, "/fhir/$export-status/nonexistent", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("nonexistent")

	err := h.ExportStatus(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestExportHandler_ExportStatus_InProgress(t *testing.T) {
	mgr := NewExportManager()
	h := NewExportHandler(mgr)
	e := echo.New()

	// Manually create an in-progress job
	mgr.mu.Lock()
	mgr.jobs["in-progress-1"] = &ExportJob{
		ID:            "in-progress-1",
		Status:        "in-progress",
		ResourceTypes: []string{"Patient"},
		OutputFormat:  "application/fhir+ndjson",
		CreatedAt:     time.Now().UTC(),
	}
	mgr.mu.Unlock()

	req := httptest.NewRequest(http.MethodGet, "/fhir/$export-status/in-progress-1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("in-progress-1")

	err := h.ExportStatus(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusAccepted {
		t.Errorf("expected 202, got %d", rec.Code)
	}
	progress := rec.Header().Get("X-Progress")
	if progress != "in-progress" {
		t.Errorf("expected X-Progress 'in-progress', got %q", progress)
	}
}

func TestExportHandler_ExportStatus_Error(t *testing.T) {
	mgr := NewExportManager()
	h := NewExportHandler(mgr)
	e := echo.New()

	// Manually create an error job
	mgr.mu.Lock()
	mgr.jobs["error-1"] = &ExportJob{
		ID:           "error-1",
		Status:       "error",
		ErrorMessage: "something went wrong",
		CreatedAt:    time.Now().UTC(),
	}
	mgr.mu.Unlock()

	req := httptest.NewRequest(http.MethodGet, "/fhir/$export-status/error-1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("error-1")

	err := h.ExportStatus(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rec.Code)
	}
}

func TestExportHandler_ExportData(t *testing.T) {
	mgr := NewExportManager()
	h := NewExportHandler(mgr)
	e := echo.New()

	job := mgr.KickOff([]string{"Patient"}, nil)

	req := httptest.NewRequest(http.MethodGet, "/fhir/$export-data/"+job.ID+"/Patient", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id", "type")
	c.SetParamValues(job.ID, "Patient")

	err := h.ExportData(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	contentType := rec.Header().Get("Content-Type")
	if contentType != "application/fhir+ndjson" {
		t.Errorf("expected content-type 'application/fhir+ndjson', got %q", contentType)
	}
	if rec.Body.Len() == 0 {
		t.Error("expected non-empty body")
	}
}

func TestExportHandler_ExportData_NotFound(t *testing.T) {
	mgr := NewExportManager()
	h := NewExportHandler(mgr)
	e := echo.New()

	req := httptest.NewRequest(http.MethodGet, "/fhir/$export-data/nonexistent/Patient", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id", "type")
	c.SetParamValues("nonexistent", "Patient")

	err := h.ExportData(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestExportHandler_RegisterRoutes(t *testing.T) {
	mgr := NewExportManager()
	h := NewExportHandler(mgr)
	e := echo.New()
	fhirGroup := e.Group("/fhir")

	h.RegisterRoutes(fhirGroup)

	routes := e.Routes()
	routePaths := make(map[string]bool)
	for _, r := range routes {
		routePaths[r.Method+":"+r.Path] = true
	}

	expected := []string{
		"POST:/fhir/$export",
		"POST:/fhir/Patient/$export",
		"GET:/fhir/$export-status/:id",
		"GET:/fhir/$export-data/:id/:type",
	}
	for _, path := range expected {
		if !routePaths[path] {
			t.Errorf("missing expected route: %s", path)
		}
	}
}
