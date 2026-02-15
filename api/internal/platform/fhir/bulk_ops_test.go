package fhir

import (
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

// ==================== Validator Tests ====================

func TestDefaultBulkValidator_ValidResource(t *testing.T) {
	v := &DefaultBulkValidator{}
	resource := map[string]interface{}{
		"resourceType": "Patient",
		"id":           "p1",
		"status":       "active",
	}
	err := v.ValidateResource("Patient", resource)
	if err != nil {
		t.Fatalf("expected no error for valid resource, got: %v", err)
	}
}

func TestDefaultBulkValidator_MissingResourceType(t *testing.T) {
	v := &DefaultBulkValidator{}
	resource := map[string]interface{}{
		"id":     "p1",
		"status": "active",
	}
	err := v.ValidateResource("Patient", resource)
	if err == nil {
		t.Fatal("expected error for missing resourceType, got nil")
	}
	if !strings.Contains(err.Error(), "resourceType") {
		t.Fatalf("expected error about resourceType, got: %v", err)
	}
}

func TestDefaultBulkValidator_WrongResourceType(t *testing.T) {
	v := &DefaultBulkValidator{}
	resource := map[string]interface{}{
		"resourceType": "Observation",
		"id":           "o1",
		"status":       "final",
	}
	err := v.ValidateResource("Patient", resource)
	if err == nil {
		t.Fatal("expected error for wrong resourceType, got nil")
	}
	if !strings.Contains(err.Error(), "mismatch") {
		t.Fatalf("expected error about type mismatch, got: %v", err)
	}
}

func TestDefaultBulkValidator_MissingID(t *testing.T) {
	v := &DefaultBulkValidator{}
	resource := map[string]interface{}{
		"resourceType": "Patient",
		"status":       "active",
	}
	err := v.ValidateResource("Patient", resource)
	if err == nil {
		t.Fatal("expected error for missing id, got nil")
	}
	if !strings.Contains(err.Error(), "id") {
		t.Fatalf("expected error about missing id, got: %v", err)
	}
}

// ==================== Import Tests ====================

func TestBulkImport_StartImport(t *testing.T) {
	mgr := NewBulkOperationManager(nil)
	ndjson := "{\"resourceType\":\"Patient\",\"id\":\"p1\",\"status\":\"active\"}\n{\"resourceType\":\"Patient\",\"id\":\"p2\",\"status\":\"active\"}\n"
	job, err := mgr.StartImport(context.Background(), "Patient", []byte(ndjson))
	if err != nil {
		t.Fatalf("StartImport failed: %v", err)
	}
	if job.ID == "" {
		t.Fatal("expected job ID to be set")
	}
	if job.Status != "completed" {
		t.Fatalf("expected status completed, got %s", job.Status)
	}
	if job.TotalResources != 2 {
		t.Fatalf("expected 2 total resources, got %d", job.TotalResources)
	}
	if job.SuccessCount != 2 {
		t.Fatalf("expected 2 successes, got %d", job.SuccessCount)
	}
	if job.ErrorCount != 0 {
		t.Fatalf("expected 0 errors, got %d", job.ErrorCount)
	}
}

func TestBulkImport_ParseNDJSON(t *testing.T) {
	mgr := NewBulkOperationManager(nil)
	ndjson := "{\"resourceType\":\"Observation\",\"id\":\"o1\",\"status\":\"final\"}\n{\"resourceType\":\"Observation\",\"id\":\"o2\",\"status\":\"final\"}\n{\"resourceType\":\"Observation\",\"id\":\"o3\",\"status\":\"final\"}\n"
	job, err := mgr.StartImport(context.Background(), "Observation", []byte(ndjson))
	if err != nil {
		t.Fatalf("StartImport failed: %v", err)
	}
	if job.TotalResources != 3 {
		t.Fatalf("expected 3 resources parsed, got %d", job.TotalResources)
	}
	if job.SuccessCount != 3 {
		t.Fatalf("expected 3 successes, got %d", job.SuccessCount)
	}
}

func TestBulkImport_ValidationErrors(t *testing.T) {
	mgr := NewBulkOperationManager(nil)
	ndjson := "{\"resourceType\":\"Patient\",\"id\":\"p1\",\"status\":\"active\"}\n{\"resourceType\":\"Patient\",\"status\":\"active\"}\n{\"resourceType\":\"Observation\",\"id\":\"o1\",\"status\":\"final\"}\n"
	job, err := mgr.StartImport(context.Background(), "Patient", []byte(ndjson))
	if err != nil {
		t.Fatalf("StartImport failed: %v", err)
	}
	if job.TotalResources != 3 {
		t.Fatalf("expected 3 total resources, got %d", job.TotalResources)
	}
	if job.SuccessCount != 1 {
		t.Fatalf("expected 1 success, got %d", job.SuccessCount)
	}
	if job.ErrorCount != 2 {
		t.Fatalf("expected 2 errors, got %d", job.ErrorCount)
	}
	if len(job.Errors) != 2 {
		t.Fatalf("expected 2 error entries, got %d", len(job.Errors))
	}
}

func TestBulkImport_EmptyInput(t *testing.T) {
	mgr := NewBulkOperationManager(nil)
	_, err := mgr.StartImport(context.Background(), "Patient", []byte(""))
	if err == nil {
		t.Fatal("expected error for empty input, got nil")
	}
	if !strings.Contains(err.Error(), "empty") {
		t.Fatalf("expected error about empty input, got: %v", err)
	}
}

func TestBulkImport_GetStatus(t *testing.T) {
	mgr := NewBulkOperationManager(nil)
	ndjson := "{\"resourceType\":\"Patient\",\"id\":\"p1\",\"status\":\"active\"}\n{\"resourceType\":\"Patient\",\"id\":\"p2\",\"status\":\"active\"}\n"
	job, err := mgr.StartImport(context.Background(), "Patient", []byte(ndjson))
	if err != nil {
		t.Fatalf("StartImport failed: %v", err)
	}

	status, err := mgr.GetImportStatus(context.Background(), job.ID)
	if err != nil {
		t.Fatalf("GetImportStatus failed: %v", err)
	}
	if status.ID != job.ID {
		t.Fatalf("expected job ID %s, got %s", job.ID, status.ID)
	}
	if status.Status != "completed" {
		t.Fatalf("expected status completed, got %s", status.Status)
	}
	if status.SuccessCount != 2 {
		t.Fatalf("expected 2 successes, got %d", status.SuccessCount)
	}
}

func TestBulkImport_ListJobs(t *testing.T) {
	mgr := NewBulkOperationManager(nil)
	ndjson := "{\"resourceType\":\"Patient\",\"id\":\"p1\",\"status\":\"active\"}\n"
	_, err := mgr.StartImport(context.Background(), "Patient", []byte(ndjson))
	if err != nil {
		t.Fatalf("StartImport 1 failed: %v", err)
	}
	_, err = mgr.StartImport(context.Background(), "Patient", []byte(ndjson))
	if err != nil {
		t.Fatalf("StartImport 2 failed: %v", err)
	}

	jobs, err := mgr.ListImportJobs(context.Background(), 10)
	if err != nil {
		t.Fatalf("ListImportJobs failed: %v", err)
	}
	if len(jobs) != 2 {
		t.Fatalf("expected 2 jobs, got %d", len(jobs))
	}
}

func TestBulkImport_ConcurrentJobs(t *testing.T) {
	mgr := NewBulkOperationManagerWithOptions(nil, 2)
	ndjson := "{\"resourceType\":\"Patient\",\"id\":\"p1\",\"status\":\"active\"}\n"

	// Create 2 jobs (which is the max)
	_, err := mgr.StartImport(context.Background(), "Patient", []byte(ndjson))
	if err != nil {
		t.Fatalf("StartImport 1 failed: %v", err)
	}
	_, err = mgr.StartImport(context.Background(), "Patient", []byte(ndjson))
	if err != nil {
		t.Fatalf("StartImport 2 failed: %v", err)
	}

	// These completed synchronously, so the pending count is 0 and we can
	// create more. To test the limit, manually set jobs to processing status.
	mgr.importMu.Lock()
	for _, j := range mgr.importJobs {
		j.Status = "processing"
	}
	mgr.importMu.Unlock()

	// Now the third should fail
	_, err = mgr.StartImport(context.Background(), "Patient", []byte(ndjson))
	if err == nil {
		t.Fatal("expected error when exceeding concurrent job limit, got nil")
	}
	if !strings.Contains(err.Error(), "concurrent") {
		t.Fatalf("expected concurrent job limit error, got: %v", err)
	}
}

// ==================== Edit Tests ====================

func TestBulkEdit_StartUpdate(t *testing.T) {
	store := NewInMemoryResourceStore()
	store.AddResource("Patient", map[string]interface{}{
		"resourceType": "Patient",
		"id":           "p1",
		"status":       "active",
		"name":         "John",
	})
	store.AddResource("Patient", map[string]interface{}{
		"resourceType": "Patient",
		"id":           "p2",
		"status":       "active",
		"name":         "Jane",
	})

	mgr := NewBulkOperationManager(store)
	criteria := map[string]string{"status": "active"}
	patch := map[string]interface{}{"status": "inactive"}

	job, err := mgr.StartBulkUpdate(context.Background(), "Patient", criteria, patch)
	if err != nil {
		t.Fatalf("StartBulkUpdate failed: %v", err)
	}
	if job.ID == "" {
		t.Fatal("expected job ID to be set")
	}
	if job.Status != "completed" {
		t.Fatalf("expected status completed, got %s", job.Status)
	}
	if job.MatchCount != 2 {
		t.Fatalf("expected 2 matches, got %d", job.MatchCount)
	}
	if job.ModifiedCount != 2 {
		t.Fatalf("expected 2 modified, got %d", job.ModifiedCount)
	}
	if job.Operation != "update" {
		t.Fatalf("expected operation update, got %s", job.Operation)
	}
}

func TestBulkEdit_StartUpdate_NoCriteria(t *testing.T) {
	store := NewInMemoryResourceStore()
	mgr := NewBulkOperationManager(store)
	_, err := mgr.StartBulkUpdate(context.Background(), "Patient", nil, map[string]interface{}{"status": "inactive"})
	if err == nil {
		t.Fatal("expected error for empty criteria, got nil")
	}
	if !strings.Contains(err.Error(), "criteria") {
		t.Fatalf("expected error about criteria, got: %v", err)
	}
}

func TestBulkEdit_StartDelete(t *testing.T) {
	store := NewInMemoryResourceStore()
	store.AddResource("Patient", map[string]interface{}{
		"resourceType": "Patient",
		"id":           "p1",
		"status":       "active",
	})
	store.AddResource("Patient", map[string]interface{}{
		"resourceType": "Patient",
		"id":           "p2",
		"status":       "active",
	})

	mgr := NewBulkOperationManager(store)
	criteria := map[string]string{"status": "active"}

	job, err := mgr.StartBulkDelete(context.Background(), "Patient", criteria)
	if err != nil {
		t.Fatalf("StartBulkDelete failed: %v", err)
	}
	if job.Operation != "delete" {
		t.Fatalf("expected operation delete, got %s", job.Operation)
	}
	if job.MatchCount != 2 {
		t.Fatalf("expected 2 matches, got %d", job.MatchCount)
	}
	if job.ModifiedCount != 2 {
		t.Fatalf("expected 2 modified (deleted), got %d", job.ModifiedCount)
	}
	if job.Status != "completed" {
		t.Fatalf("expected status completed, got %s", job.Status)
	}
}

func TestBulkEdit_GetStatus(t *testing.T) {
	store := NewInMemoryResourceStore()
	store.AddResource("Patient", map[string]interface{}{
		"resourceType": "Patient",
		"id":           "p1",
		"status":       "active",
	})

	mgr := NewBulkOperationManager(store)
	criteria := map[string]string{"status": "active"}
	patch := map[string]interface{}{"status": "inactive"}

	job, err := mgr.StartBulkUpdate(context.Background(), "Patient", criteria, patch)
	if err != nil {
		t.Fatalf("StartBulkUpdate failed: %v", err)
	}

	status, err := mgr.GetEditStatus(context.Background(), job.ID)
	if err != nil {
		t.Fatalf("GetEditStatus failed: %v", err)
	}
	if status.ID != job.ID {
		t.Fatalf("expected job ID %s, got %s", job.ID, status.ID)
	}
	if status.Status != "completed" {
		t.Fatalf("expected status completed, got %s", status.Status)
	}
}

func TestBulkEdit_ListJobs(t *testing.T) {
	store := NewInMemoryResourceStore()
	store.AddResource("Patient", map[string]interface{}{
		"resourceType": "Patient",
		"id":           "p1",
		"status":       "active",
	})

	mgr := NewBulkOperationManager(store)
	criteria := map[string]string{"status": "active"}
	patch := map[string]interface{}{"status": "inactive"}

	_, err := mgr.StartBulkUpdate(context.Background(), "Patient", criteria, patch)
	if err != nil {
		t.Fatalf("StartBulkUpdate failed: %v", err)
	}
	_, err = mgr.StartBulkDelete(context.Background(), "Patient", criteria)
	if err != nil {
		t.Fatalf("StartBulkDelete failed: %v", err)
	}

	jobs, err := mgr.ListEditJobs(context.Background(), 10)
	if err != nil {
		t.Fatalf("ListEditJobs failed: %v", err)
	}
	if len(jobs) != 2 {
		t.Fatalf("expected 2 edit jobs, got %d", len(jobs))
	}
}

func TestBulkEdit_CancelJob(t *testing.T) {
	store := NewInMemoryResourceStore()
	mgr := NewBulkOperationManager(store)

	// Manually insert a processing job to cancel
	mgr.editMu.Lock()
	jobID := "cancel-test-id"
	mgr.editJobs[jobID] = &BulkEditJob{
		ID:           jobID,
		Status:       "processing",
		Operation:    "update",
		ResourceType: "Patient",
		RequestTime:  time.Now().UTC(),
	}
	mgr.editMu.Unlock()

	err := mgr.CancelJob(context.Background(), jobID)
	if err != nil {
		t.Fatalf("CancelJob failed: %v", err)
	}

	job, err := mgr.GetEditStatus(context.Background(), jobID)
	if err != nil {
		t.Fatalf("GetEditStatus failed: %v", err)
	}
	if job.Status != "cancelled" {
		t.Fatalf("expected status cancelled, got %s", job.Status)
	}
}

func TestBulkEdit_CancelCompletedJob(t *testing.T) {
	store := NewInMemoryResourceStore()
	store.AddResource("Patient", map[string]interface{}{
		"resourceType": "Patient",
		"id":           "p1",
		"status":       "active",
	})

	mgr := NewBulkOperationManager(store)
	criteria := map[string]string{"status": "active"}
	patch := map[string]interface{}{"status": "inactive"}

	job, err := mgr.StartBulkUpdate(context.Background(), "Patient", criteria, patch)
	if err != nil {
		t.Fatalf("StartBulkUpdate failed: %v", err)
	}

	err = mgr.CancelJob(context.Background(), job.ID)
	if err == nil {
		t.Fatal("expected error cancelling completed job, got nil")
	}
	if !strings.Contains(err.Error(), "cannot cancel") {
		t.Fatalf("expected cannot cancel error, got: %v", err)
	}
}

func TestBulkEdit_NoMatches(t *testing.T) {
	store := NewInMemoryResourceStore()
	store.AddResource("Patient", map[string]interface{}{
		"resourceType": "Patient",
		"id":           "p1",
		"status":       "active",
	})

	mgr := NewBulkOperationManager(store)
	criteria := map[string]string{"status": "inactive"}
	patch := map[string]interface{}{"status": "active"}

	job, err := mgr.StartBulkUpdate(context.Background(), "Patient", criteria, patch)
	if err != nil {
		t.Fatalf("StartBulkUpdate failed: %v", err)
	}
	if job.MatchCount != 0 {
		t.Fatalf("expected 0 matches, got %d", job.MatchCount)
	}
	if job.ModifiedCount != 0 {
		t.Fatalf("expected 0 modified, got %d", job.ModifiedCount)
	}
	if job.Status != "completed" {
		t.Fatalf("expected status completed, got %s", job.Status)
	}
}

// ==================== Handler Tests ====================

func setupBulkOpsHandler() (*echo.Echo, *BulkOpsHandler) {
	e := echo.New()
	store := NewInMemoryResourceStore()
	store.AddResource("Patient", map[string]interface{}{
		"resourceType": "Patient",
		"id":           "p1",
		"status":       "active",
		"name":         "Test",
	})
	mgr := NewBulkOperationManager(store)
	handler := NewBulkOpsHandler(mgr)
	g := e.Group("/fhir")
	handler.RegisterRoutes(g)
	return e, handler
}

func TestBulkOpsHandler_StartImport(t *testing.T) {
	e, _ := setupBulkOpsHandler()
	ndjson := "{\"resourceType\":\"Patient\",\"id\":\"p1\",\"status\":\"active\"}\n{\"resourceType\":\"Patient\",\"id\":\"p2\",\"status\":\"active\"}\n"
	req := httptest.NewRequest(http.MethodPost, "/fhir/$import?resourceType=Patient", strings.NewReader(ndjson))
	req.Header.Set("Content-Type", "application/fhir+ndjson")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected 202 Accepted, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp["id"] == nil || resp["id"] == "" {
		t.Fatal("expected job id in response")
	}
	if resp["status"] == nil {
		t.Fatal("expected status in response")
	}
}

func TestBulkOpsHandler_GetImportStatus(t *testing.T) {
	e, handler := setupBulkOpsHandler()

	// First, start an import to get a job ID
	ndjson := "{\"resourceType\":\"Patient\",\"id\":\"p1\",\"status\":\"active\"}\n"
	req := httptest.NewRequest(http.MethodPost, "/fhir/$import?resourceType=Patient", strings.NewReader(ndjson))
	req.Header.Set("Content-Type", "application/fhir+ndjson")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	var createResp map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &createResp); err != nil {
		t.Fatalf("failed to parse create response: %v", err)
	}
	jobID := fmt.Sprintf("%v", createResp["id"])

	// Now get the status
	_ = handler // suppress unused
	req2 := httptest.NewRequest(http.MethodGet, "/fhir/$import/"+jobID, nil)
	rec2 := httptest.NewRecorder()
	e.ServeHTTP(rec2, req2)

	if rec2.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec2.Code, rec2.Body.String())
	}

	var statusResp map[string]interface{}
	if err := json.Unmarshal(rec2.Body.Bytes(), &statusResp); err != nil {
		t.Fatalf("failed to parse status response: %v", err)
	}
	if statusResp["id"] != jobID {
		t.Fatalf("expected job ID %s, got %v", jobID, statusResp["id"])
	}
}

func TestBulkOpsHandler_StartBulkEdit(t *testing.T) {
	e, _ := setupBulkOpsHandler()

	body := "{\"operation\":\"update\",\"resourceType\":\"Patient\",\"criteria\":{\"status\":\"active\"},\"patch\":{\"status\":\"inactive\"}}"
	req := httptest.NewRequest(http.MethodPost, "/fhir/$bulk-edit", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected 202 Accepted, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp["id"] == nil || resp["id"] == "" {
		t.Fatal("expected job id in response")
	}
}

func TestBulkOpsHandler_StartBulkDelete(t *testing.T) {
	e, _ := setupBulkOpsHandler()

	body := "{\"resourceType\":\"Patient\",\"criteria\":{\"status\":\"active\"}}"
	req := httptest.NewRequest(http.MethodPost, "/fhir/$bulk-delete", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected 202 Accepted, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp["id"] == nil || resp["id"] == "" {
		t.Fatal("expected job id in response")
	}
}

func TestBulkOpsHandler_CancelJob(t *testing.T) {
	e, handler := setupBulkOpsHandler()

	// Manually insert a processing edit job
	handler.manager.editMu.Lock()
	jobID := "handler-cancel-test"
	handler.manager.editJobs[jobID] = &BulkEditJob{
		ID:           jobID,
		Status:       "processing",
		Operation:    "update",
		ResourceType: "Patient",
		RequestTime:  time.Now().UTC(),
	}
	handler.manager.editMu.Unlock()

	req := httptest.NewRequest(http.MethodDelete, "/fhir/$bulk-edit/"+jobID, nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}
