package fhir

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
)

// =========== Mock ResourceExporter ===========

// mockExporter is a test double for the ResourceExporter interface.
type mockExporter struct {
	resources []map[string]interface{}
	err       error
	// track calls
	exportAllCalled      bool
	exportByPatientCalls []string
	sincePassed          *time.Time
}

func (m *mockExporter) ExportAll(ctx context.Context, since *time.Time) ([]map[string]interface{}, error) {
	m.exportAllCalled = true
	m.sincePassed = since
	if m.err != nil {
		return nil, m.err
	}
	return m.resources, nil
}

func (m *mockExporter) ExportByPatient(ctx context.Context, patientID string, since *time.Time) ([]map[string]interface{}, error) {
	m.exportByPatientCalls = append(m.exportByPatientCalls, patientID)
	m.sincePassed = since
	if m.err != nil {
		return nil, m.err
	}
	return m.resources, nil
}

// waitForComplete polls the manager until the job reaches "complete" or "error" status, or times out.
func waitForComplete(t *testing.T, mgr *ExportManager, jobID string, timeout time.Duration) *ExportJob {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		job, err := mgr.GetStatus(jobID)
		if err != nil {
			t.Fatalf("GetStatus failed: %v", err)
		}
		if job.Status == "complete" || job.Status == "error" {
			return job
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("job %s did not complete within %v", jobID, timeout)
	return nil
}

// mustKickOff calls KickOff and fails the test on error.
func mustKickOff(t *testing.T, mgr *ExportManager, resourceTypes []string, since *time.Time) *ExportJob {
	t.Helper()
	job, err := mgr.KickOff(resourceTypes, since)
	if err != nil {
		t.Fatalf("KickOff failed: %v", err)
	}
	return job
}

// mustKickOffForPatient calls KickOffForPatient and fails the test on error.
func mustKickOffForPatient(t *testing.T, mgr *ExportManager, resourceTypes []string, patientID string, since *time.Time) *ExportJob {
	t.Helper()
	job, err := mgr.KickOffForPatient(resourceTypes, patientID, since)
	if err != nil {
		t.Fatalf("KickOffForPatient failed: %v", err)
	}
	return job
}

// =========== ExportManager Tests ===========

func TestExportManager_KickOff_CreatesJob(t *testing.T) {
	mgr := NewExportManager()

	// With no exporters registered, kick off should still create a job
	job := mustKickOff(t, mgr,[]string{"Patient", "Observation"}, nil)

	if job.ID == "" {
		t.Error("expected job ID to be set")
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
	if job.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}
}

func TestExportManager_KickOff_AsyncProcessing(t *testing.T) {
	mgr := NewExportManager()
	exporter := &mockExporter{
		resources: []map[string]interface{}{
			{"resourceType": "Patient", "id": "p1", "name": "Alice"},
		},
	}
	mgr.RegisterExporter("Patient", exporter)

	job := mustKickOff(t, mgr,[]string{"Patient"}, nil)

	// Job should start in a non-complete state (in-progress or accepted)
	// It transitions to complete asynchronously
	if job.Status != "in-progress" {
		t.Errorf("expected initial status 'in-progress', got %q", job.Status)
	}

	// Wait for completion
	completed := waitForComplete(t, mgr, job.ID, 5*time.Second)
	if completed.Status != "complete" {
		t.Errorf("expected final status 'complete', got %q", completed.Status)
	}
	if completed.CompletedAt == nil {
		t.Error("expected CompletedAt to be set after completion")
	}
}

func TestExportManager_KickOff_DefaultTypes(t *testing.T) {
	mgr := NewExportManager()

	job := mustKickOff(t, mgr,nil, nil)

	if len(job.ResourceTypes) != 5 {
		t.Errorf("expected 5 default resource types, got %d", len(job.ResourceTypes))
	}
}

func TestExportManager_KickOff_WithSince(t *testing.T) {
	mgr := NewExportManager()

	since := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	job := mustKickOff(t, mgr,[]string{"Patient"}, &since)

	if job.Since == nil {
		t.Fatal("expected Since to be set")
	}
	if !job.Since.Equal(since) {
		t.Errorf("expected Since %v, got %v", since, *job.Since)
	}
}

func TestExportManager_Status_Pending(t *testing.T) {
	// Create a manager with a slow exporter to observe in-progress state
	mgr := NewExportManager()
	slowExporter := &blockingExporter{
		started: make(chan struct{}),
		release: make(chan struct{}),
		resources: []map[string]interface{}{
			{"resourceType": "Patient", "id": "p1"},
		},
	}
	mgr.RegisterExporter("Patient", slowExporter)

	job := mustKickOff(t, mgr,[]string{"Patient"}, nil)

	// Wait for the exporter goroutine to start
	<-slowExporter.started

	// While the exporter is blocked, status should be in-progress
	status, err := mgr.GetStatus(job.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status.Status != "in-progress" {
		t.Errorf("expected status 'in-progress' while processing, got %q", status.Status)
	}

	// Release the exporter
	close(slowExporter.release)

	// Wait for completion
	completed := waitForComplete(t, mgr, job.ID, 5*time.Second)
	if completed.Status != "complete" {
		t.Errorf("expected 'complete', got %q", completed.Status)
	}
}

func TestExportManager_Status_Complete(t *testing.T) {
	mgr := NewExportManager()
	exporter := &mockExporter{
		resources: []map[string]interface{}{
			{"resourceType": "Patient", "id": "p1"},
			{"resourceType": "Patient", "id": "p2"},
		},
	}
	mgr.RegisterExporter("Patient", exporter)

	job := mustKickOff(t, mgr,[]string{"Patient"}, nil)
	completed := waitForComplete(t, mgr, job.ID, 5*time.Second)

	if completed.Status != "complete" {
		t.Errorf("expected 'complete', got %q", completed.Status)
	}
	if len(completed.OutputFiles) != 1 {
		t.Fatalf("expected 1 output file, got %d", len(completed.OutputFiles))
	}
	if completed.OutputFiles[0].Count != 2 {
		t.Errorf("expected count 2, got %d", completed.OutputFiles[0].Count)
	}
	expectedURL := fmt.Sprintf("/fhir/$export-data/%s/Patient", job.ID)
	if completed.OutputFiles[0].URL != expectedURL {
		t.Errorf("expected URL %q, got %q", expectedURL, completed.OutputFiles[0].URL)
	}
}

func TestExportManager_GetJobData_ReturnsNDJSON(t *testing.T) {
	mgr := NewExportManager()
	exporter := &mockExporter{
		resources: []map[string]interface{}{
			{"resourceType": "Patient", "id": "p1", "name": "Alice"},
			{"resourceType": "Patient", "id": "p2", "name": "Bob"},
		},
	}
	mgr.RegisterExporter("Patient", exporter)

	job := mustKickOff(t, mgr,[]string{"Patient"}, nil)
	waitForComplete(t, mgr, job.ID, 5*time.Second)

	data, err := mgr.GetJobData(job.ID, "Patient")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify NDJSON format: each line is a valid JSON object
	scanner := bufio.NewScanner(bytes.NewReader(data))
	var resources []map[string]interface{}
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		var resource map[string]interface{}
		if err := json.Unmarshal([]byte(line), &resource); err != nil {
			t.Fatalf("invalid NDJSON line: %v\nline: %s", err, line)
		}
		resources = append(resources, resource)
	}

	if len(resources) != 2 {
		t.Fatalf("expected 2 NDJSON lines, got %d", len(resources))
	}
	if resources[0]["id"] != "p1" {
		t.Errorf("expected first resource id 'p1', got %v", resources[0]["id"])
	}
	if resources[1]["id"] != "p2" {
		t.Errorf("expected second resource id 'p2', got %v", resources[1]["id"])
	}
}

func TestExportManager_GetJobData_RealData(t *testing.T) {
	// Verify data comes from the registered exporter, not placeholders
	mgr := NewExportManager()
	exporter := &mockExporter{
		resources: []map[string]interface{}{
			{"resourceType": "Observation", "id": "obs-42", "status": "final", "code": "blood-pressure"},
		},
	}
	mgr.RegisterExporter("Observation", exporter)

	job := mustKickOff(t, mgr,[]string{"Observation"}, nil)
	waitForComplete(t, mgr, job.ID, 5*time.Second)

	data, err := mgr.GetJobData(job.ID, "Observation")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var resource map[string]interface{}
	if err := json.Unmarshal(bytes.TrimSpace(data), &resource); err != nil {
		t.Fatalf("failed to parse NDJSON: %v", err)
	}

	// The data should be the REAL resource from the exporter, not a placeholder
	if resource["id"] != "obs-42" {
		t.Errorf("expected id 'obs-42' from exporter, got %v", resource["id"])
	}
	if resource["status"] != "final" {
		t.Errorf("expected status 'final', got %v", resource["status"])
	}
	if resource["code"] != "blood-pressure" {
		t.Errorf("expected code 'blood-pressure', got %v", resource["code"])
	}

	// Verify the exporter was actually called
	if !exporter.exportAllCalled {
		t.Error("expected ExportAll to be called on the exporter")
	}
}

func TestExportManager_PatientExport(t *testing.T) {
	mgr := NewExportManager()
	exporter := &mockExporter{
		resources: []map[string]interface{}{
			{"resourceType": "Condition", "id": "c1", "subject": "Patient/p123"},
		},
	}
	mgr.RegisterExporter("Condition", exporter)

	job := mustKickOffForPatient(t, mgr,[]string{"Condition"}, "p123", nil)

	if job.PatientID != "p123" {
		t.Errorf("expected PatientID 'p123', got %q", job.PatientID)
	}

	waitForComplete(t, mgr, job.ID, 5*time.Second)

	// Verify ExportByPatient was called with the correct patient ID
	if len(exporter.exportByPatientCalls) == 0 {
		t.Fatal("expected ExportByPatient to be called")
	}
	if exporter.exportByPatientCalls[0] != "p123" {
		t.Errorf("expected patient ID 'p123', got %q", exporter.exportByPatientCalls[0])
	}

	data, err := mgr.GetJobData(job.ID, "Condition")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var resource map[string]interface{}
	if err := json.Unmarshal(bytes.TrimSpace(data), &resource); err != nil {
		t.Fatalf("failed to parse NDJSON: %v", err)
	}
	if resource["id"] != "c1" {
		t.Errorf("expected id 'c1', got %v", resource["id"])
	}
}

func TestExportManager_Delete(t *testing.T) {
	mgr := NewExportManager()
	exporter := &mockExporter{
		resources: []map[string]interface{}{
			{"resourceType": "Patient", "id": "p1"},
		},
	}
	mgr.RegisterExporter("Patient", exporter)

	job := mustKickOff(t, mgr,[]string{"Patient"}, nil)
	waitForComplete(t, mgr, job.ID, 5*time.Second)

	// Delete the job
	err := mgr.DeleteJob(job.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify it's gone
	_, err = mgr.GetStatus(job.ID)
	if err == nil {
		t.Error("expected error after deleting job")
	}
}

func TestExportManager_Delete_NotFound(t *testing.T) {
	mgr := NewExportManager()

	err := mgr.DeleteJob("nonexistent")
	if err == nil {
		t.Error("expected error for deleting nonexistent job")
	}
}

func TestExportManager_GetStatus_NotFound(t *testing.T) {
	mgr := NewExportManager()

	_, err := mgr.GetStatus("nonexistent-id")
	if err == nil {
		t.Error("expected error for nonexistent job")
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

	job := mustKickOff(t, mgr,[]string{"Patient"}, nil)
	waitForComplete(t, mgr, job.ID, 5*time.Second)

	_, err := mgr.GetJobData(job.ID, "Observation")
	if err == nil {
		t.Error("expected error for wrong file type")
	}
}

func TestExportManager_GetJobData_NotComplete(t *testing.T) {
	mgr := NewExportManager()
	slowExporter := &blockingExporter{
		started:   make(chan struct{}),
		release:   make(chan struct{}),
		resources: []map[string]interface{}{{"resourceType": "Patient", "id": "p1"}},
	}
	mgr.RegisterExporter("Patient", slowExporter)

	job := mustKickOff(t, mgr,[]string{"Patient"}, nil)
	<-slowExporter.started

	_, err := mgr.GetJobData(job.ID, "Patient")
	if err == nil {
		t.Error("expected error when job is not complete")
	}

	close(slowExporter.release)
	waitForComplete(t, mgr, job.ID, 5*time.Second)
}

func TestExportManager_SinceParameter(t *testing.T) {
	mgr := NewExportManager()
	exporter := &mockExporter{
		resources: []map[string]interface{}{
			{"resourceType": "Patient", "id": "recent-patient"},
		},
	}
	mgr.RegisterExporter("Patient", exporter)

	since := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	job := mustKickOff(t, mgr,[]string{"Patient"}, &since)
	waitForComplete(t, mgr, job.ID, 5*time.Second)

	// Verify the _since parameter was passed to the exporter
	if exporter.sincePassed == nil {
		t.Fatal("expected _since to be passed to exporter")
	}
	if !exporter.sincePassed.Equal(since) {
		t.Errorf("expected since %v, got %v", since, *exporter.sincePassed)
	}
}

func TestExportManager_MultipleResourceTypes(t *testing.T) {
	mgr := NewExportManager()

	patientExporter := &mockExporter{
		resources: []map[string]interface{}{
			{"resourceType": "Patient", "id": "p1"},
			{"resourceType": "Patient", "id": "p2"},
		},
	}
	obsExporter := &mockExporter{
		resources: []map[string]interface{}{
			{"resourceType": "Observation", "id": "obs1"},
		},
	}
	mgr.RegisterExporter("Patient", patientExporter)
	mgr.RegisterExporter("Observation", obsExporter)

	job := mustKickOff(t, mgr,[]string{"Patient", "Observation"}, nil)
	completed := waitForComplete(t, mgr, job.ID, 5*time.Second)

	if len(completed.OutputFiles) != 2 {
		t.Fatalf("expected 2 output files, got %d", len(completed.OutputFiles))
	}

	// Verify Patient data
	patientData, err := mgr.GetJobData(job.ID, "Patient")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	patientLines := countNDJSONLines(t, patientData)
	if patientLines != 2 {
		t.Errorf("expected 2 Patient lines, got %d", patientLines)
	}

	// Verify Observation data
	obsData, err := mgr.GetJobData(job.ID, "Observation")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	obsLines := countNDJSONLines(t, obsData)
	if obsLines != 1 {
		t.Errorf("expected 1 Observation line, got %d", obsLines)
	}
}

func TestExportManager_NoExporterForType(t *testing.T) {
	// When no exporter is registered for a type, the job should still complete
	// but the data for that type should be empty
	mgr := NewExportManager()

	job := mustKickOff(t, mgr,[]string{"UnknownType"}, nil)
	completed := waitForComplete(t, mgr, job.ID, 5*time.Second)

	if completed.Status != "complete" {
		t.Errorf("expected 'complete', got %q", completed.Status)
	}

	data, err := mgr.GetJobData(job.ID, "UnknownType")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Empty data is fine for types without exporters
	lines := countNDJSONLines(t, data)
	if lines != 0 {
		t.Errorf("expected 0 lines for unregistered type, got %d", lines)
	}
}

func TestExportManager_ExporterError(t *testing.T) {
	mgr := NewExportManager()
	exporter := &mockExporter{
		err: fmt.Errorf("database connection lost"),
	}
	mgr.RegisterExporter("Patient", exporter)

	job := mustKickOff(t, mgr,[]string{"Patient"}, nil)
	completed := waitForComplete(t, mgr, job.ID, 5*time.Second)

	if completed.Status != "error" {
		t.Errorf("expected status 'error', got %q", completed.Status)
	}
	if completed.ErrorMessage == "" {
		t.Error("expected error message to be set")
	}
	if !strings.Contains(completed.ErrorMessage, "database connection lost") {
		t.Errorf("expected error message to contain 'database connection lost', got %q", completed.ErrorMessage)
	}
}

func TestExportManager_ResourceFetcher_Integration(t *testing.T) {
	// Test that ServiceExporter adapter produces real FHIR resources
	called := false
	adapter := &ServiceExporter{
		ResourceType: "Patient",
		ListFn: func(ctx context.Context, since *time.Time) ([]map[string]interface{}, error) {
			called = true
			return []map[string]interface{}{
				{"resourceType": "Patient", "id": "real-1", "active": true},
				{"resourceType": "Patient", "id": "real-2", "active": false},
			}, nil
		},
		ListByPatientFn: func(ctx context.Context, patientID string, since *time.Time) ([]map[string]interface{}, error) {
			return []map[string]interface{}{
				{"resourceType": "Patient", "id": patientID, "active": true},
			}, nil
		},
	}

	mgr := NewExportManager()
	mgr.RegisterExporter("Patient", adapter)

	job := mustKickOff(t, mgr,[]string{"Patient"}, nil)
	waitForComplete(t, mgr, job.ID, 5*time.Second)

	if !called {
		t.Error("expected ListFn to be called")
	}

	data, err := mgr.GetJobData(job.ID, "Patient")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	lines := parseNDJSONLines(t, data)
	if len(lines) != 2 {
		t.Fatalf("expected 2 resources, got %d", len(lines))
	}
	if lines[0]["id"] != "real-1" {
		t.Errorf("expected id 'real-1', got %v", lines[0]["id"])
	}
	if lines[1]["id"] != "real-2" {
		t.Errorf("expected id 'real-2', got %v", lines[1]["id"])
	}
}

func TestServiceExporter_ExportAll(t *testing.T) {
	adapter := &ServiceExporter{
		ResourceType: "Condition",
		ListFn: func(ctx context.Context, since *time.Time) ([]map[string]interface{}, error) {
			return []map[string]interface{}{
				{"resourceType": "Condition", "id": "cond-1"},
			}, nil
		},
	}

	results, err := adapter.ExportAll(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0]["id"] != "cond-1" {
		t.Errorf("expected id 'cond-1', got %v", results[0]["id"])
	}
}

func TestServiceExporter_ExportByPatient(t *testing.T) {
	adapter := &ServiceExporter{
		ResourceType: "Condition",
		ListByPatientFn: func(ctx context.Context, patientID string, since *time.Time) ([]map[string]interface{}, error) {
			if patientID != "p1" {
				t.Errorf("expected patientID 'p1', got %q", patientID)
			}
			return []map[string]interface{}{
				{"resourceType": "Condition", "id": "cond-for-p1"},
			}, nil
		},
	}

	results, err := adapter.ExportByPatient(context.Background(), "p1", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0]["id"] != "cond-for-p1" {
		t.Errorf("expected id 'cond-for-p1', got %v", results[0]["id"])
	}
}

func TestServiceExporter_ExportAll_NoListFn(t *testing.T) {
	adapter := &ServiceExporter{
		ResourceType: "Patient",
	}

	results, err := adapter.ExportAll(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results when ListFn is nil, got %d", len(results))
	}
}

func TestServiceExporter_ExportByPatient_NoListByPatientFn(t *testing.T) {
	adapter := &ServiceExporter{
		ResourceType: "Patient",
	}

	results, err := adapter.ExportByPatient(context.Background(), "p1", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results when ListByPatientFn is nil, got %d", len(results))
	}
}

func TestServiceExporter_PassesSince(t *testing.T) {
	var capturedSince *time.Time
	adapter := &ServiceExporter{
		ResourceType: "Patient",
		ListFn: func(ctx context.Context, since *time.Time) ([]map[string]interface{}, error) {
			capturedSince = since
			return nil, nil
		},
	}

	since := time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC)
	_, _ = adapter.ExportAll(context.Background(), &since)

	if capturedSince == nil {
		t.Fatal("expected since to be passed through")
	}
	if !capturedSince.Equal(since) {
		t.Errorf("expected since %v, got %v", since, *capturedSince)
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

func TestExportHandler_PatientExport_WithPatientID(t *testing.T) {
	mgr := NewExportManager()
	exporter := &mockExporter{
		resources: []map[string]interface{}{
			{"resourceType": "Patient", "id": "abc-123"},
		},
	}
	mgr.RegisterExporter("Patient", exporter)
	h := NewExportHandler(mgr)
	e := echo.New()

	req := httptest.NewRequest(http.MethodPost, "/fhir/Patient/abc-123/$export?_type=Patient", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("patient_id")
	c.SetParamValues("abc-123")

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
	exporter := &mockExporter{
		resources: []map[string]interface{}{
			{"resourceType": "Patient", "id": "p1"},
		},
	}
	mgr.RegisterExporter("Patient", exporter)
	h := NewExportHandler(mgr)
	e := echo.New()

	job := mustKickOff(t, mgr,[]string{"Patient"}, nil)
	waitForComplete(t, mgr, job.ID, 5*time.Second)

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
		RequestTime:   time.Now().UTC(),
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
	exporter := &mockExporter{
		resources: []map[string]interface{}{
			{"resourceType": "Patient", "id": "p1"},
		},
	}
	mgr.RegisterExporter("Patient", exporter)
	h := NewExportHandler(mgr)
	e := echo.New()

	job := mustKickOff(t, mgr,[]string{"Patient"}, nil)
	waitForComplete(t, mgr, job.ID, 5*time.Second)

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

func TestExportHandler_DeleteExport(t *testing.T) {
	mgr := NewExportManager()
	exporter := &mockExporter{
		resources: []map[string]interface{}{
			{"resourceType": "Patient", "id": "p1"},
		},
	}
	mgr.RegisterExporter("Patient", exporter)
	h := NewExportHandler(mgr)
	e := echo.New()

	job := mustKickOff(t, mgr,[]string{"Patient"}, nil)
	waitForComplete(t, mgr, job.ID, 5*time.Second)

	req := httptest.NewRequest(http.MethodDelete, "/fhir/$export-status/"+job.ID, nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(job.ID)

	err := h.DeleteExport(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}

	// Verify the job is gone
	_, err = mgr.GetStatus(job.ID)
	if err == nil {
		t.Error("expected error after job deletion")
	}
}

func TestExportHandler_DeleteExport_NotFound(t *testing.T) {
	mgr := NewExportManager()
	h := NewExportHandler(mgr)
	e := echo.New()

	req := httptest.NewRequest(http.MethodDelete, "/fhir/$export-status/nonexistent", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("nonexistent")

	err := h.DeleteExport(c)
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
		"POST:/fhir/Patient/:id/$export",
		"POST:/fhir/Group/:id/$export",
		"GET:/fhir/$export-status/:id",
		"GET:/fhir/$export-data/:id/:type",
		"DELETE:/fhir/$export-status/:id",
	}
	for _, path := range expected {
		if !routePaths[path] {
			t.Errorf("missing expected route: %s", path)
		}
	}
}

// =========== New TDD Tests (Spec Compliance + Production Hardening) ===========

func TestExportManager_OutputFormat_Valid(t *testing.T) {
	mgr := NewExportManagerWithOptions(ExportOptions{MaxConcurrentJobs: 10, JobTTL: time.Hour})

	job, err := mgr.KickOffWithFormat([]string{"Patient"}, "", nil, "application/fhir+ndjson", nil)
	if err != nil {
		t.Fatalf("expected no error for valid format, got %v", err)
	}
	if job.OutputFormat != "application/fhir+ndjson" {
		t.Errorf("expected format 'application/fhir+ndjson', got %q", job.OutputFormat)
	}
}

func TestExportManager_OutputFormat_Invalid(t *testing.T) {
	mgr := NewExportManagerWithOptions(ExportOptions{MaxConcurrentJobs: 10, JobTTL: time.Hour})

	_, err := mgr.KickOffWithFormat([]string{"Patient"}, "", nil, "text/csv", nil)
	if err == nil {
		t.Fatal("expected error for invalid format text/csv")
	}
	if !strings.Contains(err.Error(), "unsupported") {
		t.Errorf("expected error containing 'unsupported', got %q", err.Error())
	}
}

func TestExportHandler_OutputFormat_Aliases(t *testing.T) {
	mgr := NewExportManagerWithOptions(ExportOptions{MaxConcurrentJobs: 10, JobTTL: time.Hour})

	// "application/ndjson" should be accepted
	job1, err := mgr.KickOffWithFormat([]string{"Patient"}, "", nil, "application/ndjson", nil)
	if err != nil {
		t.Fatalf("expected no error for application/ndjson, got %v", err)
	}
	if job1.OutputFormat != "application/fhir+ndjson" {
		t.Errorf("expected canonical format, got %q", job1.OutputFormat)
	}

	// "ndjson" should be accepted
	job2, err := mgr.KickOffWithFormat([]string{"Patient"}, "", nil, "ndjson", nil)
	if err != nil {
		t.Fatalf("expected no error for ndjson, got %v", err)
	}
	if job2.OutputFormat != "application/fhir+ndjson" {
		t.Errorf("expected canonical format, got %q", job2.OutputFormat)
	}
}

func TestExportHandler_RetryAfterHeader(t *testing.T) {
	mgr := NewExportManagerWithOptions(ExportOptions{MaxConcurrentJobs: 10, JobTTL: time.Hour})
	h := NewExportHandler(mgr)
	e := echo.New()

	// Manually create an in-progress job
	mgr.mu.Lock()
	mgr.jobs["retry-test"] = &ExportJob{
		ID:            "retry-test",
		Status:        "in-progress",
		ResourceTypes: []string{"Patient"},
		OutputFormat:  "application/fhir+ndjson",
		CreatedAt:     time.Now().UTC(),
		RequestTime:   time.Now().UTC(),
		TotalTypes:    1,
	}
	mgr.mu.Unlock()

	req := httptest.NewRequest(http.MethodGet, "/fhir/$export-status/retry-test", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("retry-test")

	err := h.ExportStatus(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusAccepted {
		t.Errorf("expected 202, got %d", rec.Code)
	}
	retryAfter := rec.Header().Get("Retry-After")
	if retryAfter != "120" {
		t.Errorf("expected Retry-After '120', got %q", retryAfter)
	}
}

func TestExportHandler_RequiresAccessToken(t *testing.T) {
	mgr := NewExportManagerWithOptions(ExportOptions{MaxConcurrentJobs: 10, JobTTL: time.Hour})
	exporter := &mockExporter{
		resources: []map[string]interface{}{
			{"resourceType": "Patient", "id": "p1"},
		},
	}
	mgr.RegisterExporter("Patient", exporter)
	h := NewExportHandler(mgr)
	e := echo.New()

	job := mustKickOff(t, mgr,[]string{"Patient"}, nil)
	waitForComplete(t, mgr, job.ID, 5*time.Second)

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
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var result map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &result)
	rat, ok := result["requiresAccessToken"].(bool)
	if !ok || !rat {
		t.Errorf("expected requiresAccessToken: true, got %v", result["requiresAccessToken"])
	}
}

func TestExportHandler_PatientExportByID(t *testing.T) {
	mgr := NewExportManagerWithOptions(ExportOptions{MaxConcurrentJobs: 10, JobTTL: time.Hour})
	exporter := &mockExporter{
		resources: []map[string]interface{}{
			{"resourceType": "Patient", "id": "patient-42"},
		},
	}
	mgr.RegisterExporter("Patient", exporter)
	h := NewExportHandler(mgr)
	e := echo.New()

	req := httptest.NewRequest(http.MethodPost, "/fhir/Patient/patient-42/$export?_type=Patient", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("patient-42")

	err := h.PatientExportByID(c)
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

	// Extract job ID and wait for completion
	parts := strings.Split(contentLocation, "/")
	jobID := parts[len(parts)-1]
	completed := waitForComplete(t, mgr, jobID, 5*time.Second)

	if len(exporter.exportByPatientCalls) == 0 {
		t.Fatal("expected ExportByPatient to be called")
	}
	if exporter.exportByPatientCalls[0] != "patient-42" {
		t.Errorf("expected patient ID 'patient-42', got %q", exporter.exportByPatientCalls[0])
	}
	if completed.Status != "complete" {
		t.Errorf("expected 'complete', got %q", completed.Status)
	}
}

func TestExportHandler_GroupExport(t *testing.T) {
	mgr := NewExportManagerWithOptions(ExportOptions{MaxConcurrentJobs: 10, JobTTL: time.Hour})
	exporter := &mockExporter{
		resources: []map[string]interface{}{
			{"resourceType": "Condition", "id": "c1"},
		},
	}
	mgr.RegisterExporter("Condition", exporter)

	// Register a group resolver that returns two patients
	mgr.SetGroupResolver(func(ctx context.Context, groupID string) ([]string, error) {
		if groupID == "grp-1" {
			return []string{"p1", "p2"}, nil
		}
		return nil, fmt.Errorf("group not found: %s", groupID)
	})

	h := NewExportHandler(mgr)
	e := echo.New()

	req := httptest.NewRequest(http.MethodPost, "/fhir/Group/grp-1/$export?_type=Condition", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("grp-1")

	err := h.GroupExport(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusAccepted {
		t.Errorf("expected 202, got %d", rec.Code)
	}

	// Extract job ID and wait
	contentLocation := rec.Header().Get("Content-Location")
	parts := strings.Split(contentLocation, "/")
	jobID := parts[len(parts)-1]
	completed := waitForComplete(t, mgr, jobID, 5*time.Second)

	if completed.Status != "complete" {
		t.Errorf("expected 'complete', got %q", completed.Status)
	}
	// ExportByPatient should have been called for both p1 and p2
	if len(exporter.exportByPatientCalls) < 2 {
		t.Errorf("expected at least 2 ExportByPatient calls, got %d", len(exporter.exportByPatientCalls))
	}
}

func TestExportHandler_GroupExport_NotFound(t *testing.T) {
	mgr := NewExportManagerWithOptions(ExportOptions{MaxConcurrentJobs: 10, JobTTL: time.Hour})

	mgr.SetGroupResolver(func(ctx context.Context, groupID string) ([]string, error) {
		return nil, fmt.Errorf("group not found: %s", groupID)
	})

	h := NewExportHandler(mgr)
	e := echo.New()

	req := httptest.NewRequest(http.MethodPost, "/fhir/Group/nonexistent/$export?_type=Patient", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("nonexistent")

	err := h.GroupExport(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestExportManager_MaxConcurrentJobs(t *testing.T) {
	mgr := NewExportManagerWithOptions(ExportOptions{MaxConcurrentJobs: 2, JobTTL: time.Hour})

	// Register a blocking exporter so jobs stay in-progress
	blocker1 := &blockingExporter{
		started:   make(chan struct{}),
		release:   make(chan struct{}),
		resources: []map[string]interface{}{{"resourceType": "Patient", "id": "p1"}},
	}
	blocker2 := &blockingExporter{
		started:   make(chan struct{}),
		release:   make(chan struct{}),
		resources: []map[string]interface{}{{"resourceType": "Patient", "id": "p2"}},
	}
	mgr.RegisterExporter("Patient", blocker1)

	// Job 1
	mustKickOff(t, mgr, []string{"Patient"}, nil)
	<-blocker1.started

	// Swap to second blocker for job 2
	mgr.RegisterExporter("Patient", blocker2)
	mustKickOff(t, mgr, []string{"Patient"}, nil)
	<-blocker2.started

	// Job 3 should fail because we're at max
	_, err := mgr.KickOffWithFormat([]string{"Patient"}, "", nil, "", nil)
	if err == nil {
		t.Fatal("expected error when exceeding max concurrent jobs")
	}
	if !strings.Contains(err.Error(), "concurrent") {
		t.Errorf("expected error about concurrent jobs, got %q", err.Error())
	}

	// Release blockers
	close(blocker1.release)
	close(blocker2.release)
}

func TestExportManager_JobExpiration(t *testing.T) {
	// Use a very short TTL for testing
	mgr := NewExportManagerWithOptions(ExportOptions{MaxConcurrentJobs: 10, JobTTL: 50 * time.Millisecond})
	exporter := &mockExporter{
		resources: []map[string]interface{}{
			{"resourceType": "Patient", "id": "p1"},
		},
	}
	mgr.RegisterExporter("Patient", exporter)

	job := mustKickOff(t, mgr,[]string{"Patient"}, nil)
	waitForComplete(t, mgr, job.ID, 5*time.Second)

	// Verify job exists
	_, err := mgr.GetStatus(job.ID)
	if err != nil {
		t.Fatalf("job should exist before cleanup: %v", err)
	}

	// Wait for TTL to expire
	time.Sleep(100 * time.Millisecond)

	// Run cleanup
	mgr.CleanupExpiredJobs()

	// Job should be gone
	_, err = mgr.GetStatus(job.ID)
	if err == nil {
		t.Error("expected job to be cleaned up after TTL")
	}
}

func TestExportManager_ProgressTracking(t *testing.T) {
	mgr := NewExportManagerWithOptions(ExportOptions{MaxConcurrentJobs: 10, JobTTL: time.Hour})
	h := NewExportHandler(mgr)
	e := echo.New()

	// Create a job with known progress
	mgr.mu.Lock()
	mgr.jobs["progress-test"] = &ExportJob{
		ID:             "progress-test",
		Status:         "in-progress",
		ResourceTypes:  []string{"Patient", "Observation", "Condition"},
		OutputFormat:   "application/fhir+ndjson",
		CreatedAt:      time.Now().UTC(),
		RequestTime:    time.Now().UTC(),
		ProcessedTypes: 2,
		TotalTypes:     3,
	}
	mgr.mu.Unlock()

	req := httptest.NewRequest(http.MethodGet, "/fhir/$export-status/progress-test", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("progress-test")

	err := h.ExportStatus(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	progress := rec.Header().Get("X-Progress")
	if progress != "2/3 resource types processed" {
		t.Errorf("expected X-Progress '2/3 resource types processed', got %q", progress)
	}
}

func TestExportHandler_TypeFilter(t *testing.T) {
	mgr := NewExportManagerWithOptions(ExportOptions{MaxConcurrentJobs: 10, JobTTL: time.Hour})
	exporter := &mockExporter{
		resources: []map[string]interface{}{
			{"resourceType": "Observation", "id": "obs1"},
		},
	}
	mgr.RegisterExporter("Observation", exporter)
	h := NewExportHandler(mgr)
	e := echo.New()

	req := httptest.NewRequest(http.MethodPost, "/fhir/$export?_type=Observation&_typeFilter=Observation%3Fcategory%3Dlaboratory", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.SystemExport(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusAccepted {
		t.Errorf("expected 202, got %d", rec.Code)
	}

	// Verify the type filter was stored on the job
	contentLocation := rec.Header().Get("Content-Location")
	parts := strings.Split(contentLocation, "/")
	jobID := parts[len(parts)-1]

	job, err := mgr.GetStatus(jobID)
	if err != nil {
		t.Fatalf("failed to get job: %v", err)
	}
	if len(job.TypeFilter) == 0 {
		t.Error("expected TypeFilter to be stored on job")
	}
	if job.TypeFilter[0] != "Observation?category=laboratory" {
		t.Errorf("expected type filter 'Observation?category=laboratory', got %q", job.TypeFilter[0])
	}
}

func TestExportManager_TransactionTime(t *testing.T) {
	mgr := NewExportManagerWithOptions(ExportOptions{MaxConcurrentJobs: 10, JobTTL: time.Hour})
	exporter := &mockExporter{
		resources: []map[string]interface{}{
			{"resourceType": "Patient", "id": "p1"},
		},
	}
	mgr.RegisterExporter("Patient", exporter)
	h := NewExportHandler(mgr)
	e := echo.New()

	beforeKickOff := time.Now().UTC()
	job := mustKickOff(t, mgr,[]string{"Patient"}, nil)
	waitForComplete(t, mgr, job.ID, 5*time.Second)

	req := httptest.NewRequest(http.MethodGet, "/fhir/$export-status/"+job.ID, nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(job.ID)

	err := h.ExportStatus(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &result)

	txTimeStr, ok := result["transactionTime"].(string)
	if !ok {
		t.Fatal("expected transactionTime in response")
	}

	txTime, err := time.Parse(time.RFC3339, txTimeStr)
	if err != nil {
		t.Fatalf("invalid transactionTime format: %v", err)
	}

	// transactionTime should be the request time (before or close to kickoff), not the completion time
	if txTime.Before(beforeKickOff.Add(-time.Second)) {
		t.Errorf("transactionTime %v is before kickoff %v", txTime, beforeKickOff)
	}
	// It should be within a second of kickoff, not at completion time
	if txTime.After(beforeKickOff.Add(2 * time.Second)) {
		t.Errorf("transactionTime %v is too far after kickoff %v; should be request time, not completion", txTime, beforeKickOff)
	}
}

// =========== Test Helpers ===========

// blockingExporter is a test helper that blocks until released.
type blockingExporter struct {
	started   chan struct{}
	release   chan struct{}
	resources []map[string]interface{}
}

func (b *blockingExporter) ExportAll(ctx context.Context, since *time.Time) ([]map[string]interface{}, error) {
	close(b.started)
	<-b.release
	return b.resources, nil
}

func (b *blockingExporter) ExportByPatient(ctx context.Context, patientID string, since *time.Time) ([]map[string]interface{}, error) {
	close(b.started)
	<-b.release
	return b.resources, nil
}

func countNDJSONLines(t *testing.T, data []byte) int {
	t.Helper()
	return len(parseNDJSONLines(t, data))
}

func parseNDJSONLines(t *testing.T, data []byte) []map[string]interface{} {
	t.Helper()
	var results []map[string]interface{}
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		var resource map[string]interface{}
		if err := json.Unmarshal([]byte(line), &resource); err != nil {
			t.Fatalf("invalid NDJSON line: %v\nline: %s", err, line)
		}
		results = append(results, resource)
	}
	return results
}
