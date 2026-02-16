package fhir

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
)

// ---------------------------------------------------------------------------
// DefaultAsyncBatchConfig
// ---------------------------------------------------------------------------

func TestDefaultAsyncBatchConfig(t *testing.T) {
	cfg := DefaultAsyncBatchConfig()

	if cfg.MaxSyncEntries <= 0 {
		t.Error("MaxSyncEntries should be positive")
	}
	if cfg.MaxSyncTimeout <= 0 {
		t.Error("MaxSyncTimeout should be positive")
	}
	if cfg.MaxAsyncEntries <= 0 {
		t.Error("MaxAsyncEntries should be positive")
	}
	if cfg.WorkerCount <= 0 {
		t.Error("WorkerCount should be positive")
	}
	if cfg.RetryMaxAttempts <= 0 {
		t.Error("RetryMaxAttempts should be positive")
	}
	if cfg.RetryDelay <= 0 {
		t.Error("RetryDelay should be positive")
	}
	if cfg.ProgressInterval <= 0 {
		t.Error("ProgressInterval should be positive")
	}
	if cfg.MaxAsyncEntries <= cfg.MaxSyncEntries {
		t.Error("MaxAsyncEntries should be greater than MaxSyncEntries")
	}
}

// ---------------------------------------------------------------------------
// ValidateAsyncBatchConfig
// ---------------------------------------------------------------------------

func TestValidateAsyncBatchConfig_Valid(t *testing.T) {
	cfg := DefaultAsyncBatchConfig()
	issues := ValidateAsyncBatchConfig(&cfg)
	if len(issues) != 0 {
		t.Errorf("expected no issues for default config, got %d: %v", len(issues), issues)
	}
}

func TestValidateAsyncBatchConfig_ZeroWorkers(t *testing.T) {
	cfg := DefaultAsyncBatchConfig()
	cfg.WorkerCount = 0
	issues := ValidateAsyncBatchConfig(&cfg)
	if len(issues) == 0 {
		t.Error("expected validation issues for zero workers")
	}
	found := false
	for _, issue := range issues {
		if strings.Contains(issue.Diagnostics, "WorkerCount") {
			found = true
		}
	}
	if !found {
		t.Error("expected issue mentioning WorkerCount")
	}
}

func TestValidateAsyncBatchConfig_NegativeEntries(t *testing.T) {
	cfg := DefaultAsyncBatchConfig()
	cfg.MaxSyncEntries = -1
	issues := ValidateAsyncBatchConfig(&cfg)
	if len(issues) == 0 {
		t.Error("expected validation issues for negative MaxSyncEntries")
	}
}

func TestValidateAsyncBatchConfig_NegativeMaxAsync(t *testing.T) {
	cfg := DefaultAsyncBatchConfig()
	cfg.MaxAsyncEntries = -5
	issues := ValidateAsyncBatchConfig(&cfg)
	if len(issues) == 0 {
		t.Error("expected validation issues for negative MaxAsyncEntries")
	}
}

func TestValidateAsyncBatchConfig_ZeroRetryAttempts(t *testing.T) {
	cfg := DefaultAsyncBatchConfig()
	cfg.RetryMaxAttempts = 0
	issues := ValidateAsyncBatchConfig(&cfg)
	if len(issues) == 0 {
		t.Error("expected validation issues for zero RetryMaxAttempts")
	}
}

func TestValidateAsyncBatchConfig_ZeroRetryDelay(t *testing.T) {
	cfg := DefaultAsyncBatchConfig()
	cfg.RetryDelay = 0
	issues := ValidateAsyncBatchConfig(&cfg)
	if len(issues) == 0 {
		t.Error("expected validation issues for zero RetryDelay")
	}
}

// ---------------------------------------------------------------------------
// ShouldProcessAsync
// ---------------------------------------------------------------------------

func TestShouldProcessAsync_BelowThreshold(t *testing.T) {
	cfg := DefaultAsyncBatchConfig()
	bundle := &TransactionBundle{
		Type:    "batch",
		Entries: make([]TransactionEntry, cfg.MaxSyncEntries-1),
	}
	if ShouldProcessAsync(bundle, &cfg, false) {
		t.Error("expected sync processing for bundle below threshold")
	}
}

func TestShouldProcessAsync_AboveThreshold(t *testing.T) {
	cfg := DefaultAsyncBatchConfig()
	bundle := &TransactionBundle{
		Type:    "batch",
		Entries: make([]TransactionEntry, cfg.MaxSyncEntries+1),
	}
	if !ShouldProcessAsync(bundle, &cfg, false) {
		t.Error("expected async processing for bundle above threshold")
	}
}

func TestShouldProcessAsync_ExactThreshold(t *testing.T) {
	cfg := DefaultAsyncBatchConfig()
	bundle := &TransactionBundle{
		Type:    "batch",
		Entries: make([]TransactionEntry, cfg.MaxSyncEntries),
	}
	// At exactly the threshold, should process synchronously
	if ShouldProcessAsync(bundle, &cfg, false) {
		t.Error("expected sync processing at exact threshold")
	}
}

func TestShouldProcessAsync_PreferAsync(t *testing.T) {
	cfg := DefaultAsyncBatchConfig()
	bundle := &TransactionBundle{
		Type:    "batch",
		Entries: make([]TransactionEntry, 1),
	}
	if !ShouldProcessAsync(bundle, &cfg, true) {
		t.Error("expected async when prefer-async is set, regardless of size")
	}
}

func TestShouldProcessAsync_EmptyBundle(t *testing.T) {
	cfg := DefaultAsyncBatchConfig()
	bundle := &TransactionBundle{
		Type:    "batch",
		Entries: nil,
	}
	if ShouldProcessAsync(bundle, &cfg, false) {
		t.Error("expected sync for empty bundle")
	}
}

func TestShouldProcessAsync_NilBundle(t *testing.T) {
	cfg := DefaultAsyncBatchConfig()
	if ShouldProcessAsync(nil, &cfg, false) {
		t.Error("expected sync for nil bundle")
	}
}

// ---------------------------------------------------------------------------
// InMemoryAsyncBatchStore
// ---------------------------------------------------------------------------

func TestInMemoryAsyncBatchStore_CreateAndGet(t *testing.T) {
	store := NewInMemoryAsyncBatchStore()
	ctx := context.Background()

	job := &AsyncBatchJob{
		BundleType:   "batch",
		TotalEntries: 50,
		Request:      "POST /fhir",
	}

	if err := store.CreateJob(ctx, job); err != nil {
		t.Fatalf("CreateJob failed: %v", err)
	}
	if job.ID == "" {
		t.Fatal("expected job ID to be generated")
	}
	if job.Status != "queued" {
		t.Errorf("expected status 'queued', got %q", job.Status)
	}
	if job.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}

	got, err := store.GetJob(ctx, job.ID)
	if err != nil {
		t.Fatalf("GetJob failed: %v", err)
	}
	if got.BundleType != "batch" {
		t.Errorf("expected BundleType 'batch', got %q", got.BundleType)
	}
	if got.TotalEntries != 50 {
		t.Errorf("expected TotalEntries 50, got %d", got.TotalEntries)
	}
}

func TestInMemoryAsyncBatchStore_GetNotFound(t *testing.T) {
	store := NewInMemoryAsyncBatchStore()
	_, err := store.GetJob(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for missing job")
	}
}

func TestInMemoryAsyncBatchStore_UpdateJob(t *testing.T) {
	store := NewInMemoryAsyncBatchStore()
	ctx := context.Background()

	job := &AsyncBatchJob{BundleType: "batch", TotalEntries: 10}
	_ = store.CreateJob(ctx, job)

	job.Status = "processing"
	job.ProcessedEntries = 5
	job.SuccessCount = 4
	job.ErrorCount = 1
	if err := store.UpdateJob(ctx, job); err != nil {
		t.Fatalf("UpdateJob failed: %v", err)
	}

	got, _ := store.GetJob(ctx, job.ID)
	if got.Status != "processing" {
		t.Errorf("expected status 'processing', got %q", got.Status)
	}
	if got.ProcessedEntries != 5 {
		t.Errorf("expected ProcessedEntries 5, got %d", got.ProcessedEntries)
	}
}

func TestInMemoryAsyncBatchStore_UpdateNotFound(t *testing.T) {
	store := NewInMemoryAsyncBatchStore()
	err := store.UpdateJob(context.Background(), &AsyncBatchJob{ID: "missing"})
	if err == nil {
		t.Fatal("expected error updating nonexistent job")
	}
}

func TestInMemoryAsyncBatchStore_CancelJob(t *testing.T) {
	store := NewInMemoryAsyncBatchStore()
	ctx := context.Background()

	job := &AsyncBatchJob{BundleType: "batch", TotalEntries: 10}
	_ = store.CreateJob(ctx, job)

	if err := store.CancelJob(ctx, job.ID); err != nil {
		t.Fatalf("CancelJob failed: %v", err)
	}

	got, _ := store.GetJob(ctx, job.ID)
	if got.Status != "cancelled" {
		t.Errorf("expected status 'cancelled', got %q", got.Status)
	}
}

func TestInMemoryAsyncBatchStore_CancelNotFound(t *testing.T) {
	store := NewInMemoryAsyncBatchStore()
	err := store.CancelJob(context.Background(), "missing")
	if err == nil {
		t.Fatal("expected error cancelling nonexistent job")
	}
}

func TestInMemoryAsyncBatchStore_ListJobs(t *testing.T) {
	store := NewInMemoryAsyncBatchStore()
	ctx := context.Background()

	// Create jobs with different statuses
	for i := 0; i < 3; i++ {
		job := &AsyncBatchJob{BundleType: "batch", TotalEntries: 10}
		_ = store.CreateJob(ctx, job)
	}
	// Update one to processing
	jobs, _ := store.ListJobs(ctx, "", 100)
	if len(jobs) < 3 {
		t.Fatalf("expected at least 3 jobs, got %d", len(jobs))
	}
	jobs[0].Status = "processing"
	_ = store.UpdateJob(ctx, jobs[0])

	// List all
	allJobs, err := store.ListJobs(ctx, "", 100)
	if err != nil {
		t.Fatalf("ListJobs failed: %v", err)
	}
	if len(allJobs) != 3 {
		t.Errorf("expected 3 jobs, got %d", len(allJobs))
	}

	// Filter by status
	queuedJobs, err := store.ListJobs(ctx, "queued", 100)
	if err != nil {
		t.Fatalf("ListJobs with filter failed: %v", err)
	}
	if len(queuedJobs) != 2 {
		t.Errorf("expected 2 queued jobs, got %d", len(queuedJobs))
	}

	// Limit
	limited, err := store.ListJobs(ctx, "", 1)
	if err != nil {
		t.Fatalf("ListJobs with limit failed: %v", err)
	}
	if len(limited) != 1 {
		t.Errorf("expected 1 job with limit, got %d", len(limited))
	}
}

func TestInMemoryAsyncBatchStore_ListJobsEmpty(t *testing.T) {
	store := NewInMemoryAsyncBatchStore()
	jobs, err := store.ListJobs(context.Background(), "", 100)
	if err != nil {
		t.Fatalf("ListJobs failed: %v", err)
	}
	if len(jobs) != 0 {
		t.Errorf("expected 0 jobs, got %d", len(jobs))
	}
}

func TestInMemoryAsyncBatchStore_DeleteJob(t *testing.T) {
	store := NewInMemoryAsyncBatchStore()
	ctx := context.Background()

	job := &AsyncBatchJob{BundleType: "batch"}
	_ = store.CreateJob(ctx, job)

	if err := store.DeleteJob(ctx, job.ID); err != nil {
		t.Fatalf("DeleteJob failed: %v", err)
	}

	_, err := store.GetJob(ctx, job.ID)
	if err == nil {
		t.Fatal("expected error after deleting job")
	}
}

func TestInMemoryAsyncBatchStore_DeleteNotFound(t *testing.T) {
	store := NewInMemoryAsyncBatchStore()
	err := store.DeleteJob(context.Background(), "missing")
	if err == nil {
		t.Fatal("expected error deleting nonexistent job")
	}
}

// ---------------------------------------------------------------------------
// SubmitAsyncBatch
// ---------------------------------------------------------------------------

func newTestProcessor() (*AsyncBatchProcessor, *InMemoryAsyncBatchStore) {
	store := NewInMemoryAsyncBatchStore()
	cfg := DefaultAsyncBatchConfig()
	cfg.WorkerCount = 2
	cfg.RetryMaxAttempts = 1
	cfg.RetryDelay = time.Millisecond
	cfg.ProgressInterval = time.Millisecond
	handler := func(method, url string, resource map[string]interface{}) (*BundleEntryResponse, error) {
		return &BundleEntryResponse{
			Status:   "200 OK",
			Location: url,
		}, nil
	}
	proc := NewAsyncBatchProcessor(store, cfg, handler)
	return proc, store
}

func TestSubmitAsyncBatch_Valid(t *testing.T) {
	proc, store := newTestProcessor()
	bundle := &TransactionBundle{
		Type: "batch",
		Entries: []TransactionEntry{
			{Request: BundleEntryRequest{Method: "GET", URL: "Patient/1"}},
			{Request: BundleEntryRequest{Method: "GET", URL: "Patient/2"}},
		},
	}

	jobID, err := proc.SubmitAsyncBatch(bundle, "POST /fhir")
	if err != nil {
		t.Fatalf("SubmitAsyncBatch failed: %v", err)
	}
	if jobID == "" {
		t.Fatal("expected non-empty job ID")
	}

	job, err := store.GetJob(context.Background(), jobID)
	if err != nil {
		t.Fatalf("GetJob failed: %v", err)
	}
	if job.TotalEntries != 2 {
		t.Errorf("expected TotalEntries 2, got %d", job.TotalEntries)
	}
	if job.BundleType != "batch" {
		t.Errorf("expected BundleType 'batch', got %q", job.BundleType)
	}
	if job.Request != "POST /fhir" {
		t.Errorf("expected Request 'POST /fhir', got %q", job.Request)
	}
}

func TestSubmitAsyncBatch_TooManyEntries(t *testing.T) {
	proc, _ := newTestProcessor()
	cfg := DefaultAsyncBatchConfig()
	entries := make([]TransactionEntry, cfg.MaxAsyncEntries+1)
	for i := range entries {
		entries[i] = TransactionEntry{Request: BundleEntryRequest{Method: "GET", URL: "Patient/1"}}
	}
	bundle := &TransactionBundle{
		Type:    "batch",
		Entries: entries,
	}

	_, err := proc.SubmitAsyncBatch(bundle, "POST /fhir")
	if err == nil {
		t.Fatal("expected error for too many entries")
	}
	if !strings.Contains(err.Error(), "exceeds maximum") {
		t.Errorf("expected 'exceeds maximum' in error, got %q", err.Error())
	}
}

func TestSubmitAsyncBatch_Transaction(t *testing.T) {
	proc, store := newTestProcessor()
	bundle := &TransactionBundle{
		Type: "transaction",
		Entries: []TransactionEntry{
			{FullURL: "urn:uuid:1", Request: BundleEntryRequest{Method: "POST", URL: "Patient"}, Resource: map[string]interface{}{"resourceType": "Patient"}},
		},
	}

	jobID, err := proc.SubmitAsyncBatch(bundle, "POST /fhir")
	if err != nil {
		t.Fatalf("SubmitAsyncBatch failed: %v", err)
	}

	job, _ := store.GetJob(context.Background(), jobID)
	if job.BundleType != "transaction" {
		t.Errorf("expected BundleType 'transaction', got %q", job.BundleType)
	}
}

// ---------------------------------------------------------------------------
// ProcessBatchAsync
// ---------------------------------------------------------------------------

func TestProcessBatchAsync_AllSuccess(t *testing.T) {
	store := NewInMemoryAsyncBatchStore()
	cfg := DefaultAsyncBatchConfig()
	cfg.WorkerCount = 2
	cfg.RetryMaxAttempts = 1
	cfg.RetryDelay = time.Millisecond
	cfg.ProgressInterval = time.Millisecond

	handler := func(method, url string, resource map[string]interface{}) (*BundleEntryResponse, error) {
		return &BundleEntryResponse{
			Status:   "200 OK",
			Location: url,
		}, nil
	}
	proc := NewAsyncBatchProcessor(store, cfg, handler)

	bundle := &TransactionBundle{
		Type: "batch",
		Entries: []TransactionEntry{
			{Request: BundleEntryRequest{Method: "GET", URL: "Patient/1"}},
			{Request: BundleEntryRequest{Method: "GET", URL: "Patient/2"}},
			{Request: BundleEntryRequest{Method: "GET", URL: "Patient/3"}},
		},
	}

	jobID, _ := proc.SubmitAsyncBatch(bundle, "POST /fhir")

	// Process synchronously for testing
	proc.ProcessBatchAsync(jobID, bundle)

	job, _ := store.GetJob(context.Background(), jobID)
	if job.Status != "completed" {
		t.Errorf("expected status 'completed', got %q", job.Status)
	}
	if job.SuccessCount != 3 {
		t.Errorf("expected SuccessCount 3, got %d", job.SuccessCount)
	}
	if job.ErrorCount != 0 {
		t.Errorf("expected ErrorCount 0, got %d", job.ErrorCount)
	}
	if job.ProcessedEntries != 3 {
		t.Errorf("expected ProcessedEntries 3, got %d", job.ProcessedEntries)
	}
	if job.Progress != 1.0 {
		t.Errorf("expected Progress 1.0, got %f", job.Progress)
	}
	if job.EndTime == nil {
		t.Error("expected EndTime to be set")
	}
	if job.ResultBundle == nil {
		t.Error("expected ResultBundle to be set")
	}
}

func TestProcessBatchAsync_SomeFailures(t *testing.T) {
	store := NewInMemoryAsyncBatchStore()
	cfg := DefaultAsyncBatchConfig()
	cfg.WorkerCount = 1
	cfg.RetryMaxAttempts = 1
	cfg.RetryDelay = time.Millisecond
	cfg.ProgressInterval = time.Millisecond

	callCount := 0
	handler := func(method, url string, resource map[string]interface{}) (*BundleEntryResponse, error) {
		callCount++
		if url == "Patient/2" {
			return nil, fmt.Errorf("not found")
		}
		return &BundleEntryResponse{
			Status:   "200 OK",
			Location: url,
		}, nil
	}
	proc := NewAsyncBatchProcessor(store, cfg, handler)

	bundle := &TransactionBundle{
		Type: "batch",
		Entries: []TransactionEntry{
			{Request: BundleEntryRequest{Method: "GET", URL: "Patient/1"}},
			{Request: BundleEntryRequest{Method: "GET", URL: "Patient/2"}},
			{Request: BundleEntryRequest{Method: "GET", URL: "Patient/3"}},
		},
	}

	jobID, _ := proc.SubmitAsyncBatch(bundle, "POST /fhir")
	proc.ProcessBatchAsync(jobID, bundle)

	job, _ := store.GetJob(context.Background(), jobID)
	if job.Status != "completed" {
		t.Errorf("expected status 'completed', got %q", job.Status)
	}
	if job.SuccessCount != 2 {
		t.Errorf("expected SuccessCount 2, got %d", job.SuccessCount)
	}
	if job.ErrorCount != 1 {
		t.Errorf("expected ErrorCount 1, got %d", job.ErrorCount)
	}
	if len(job.Errors) != 1 {
		t.Fatalf("expected 1 error, got %d", len(job.Errors))
	}
	if job.Errors[0].EntryIndex != 1 {
		t.Errorf("expected error at index 1, got %d", job.Errors[0].EntryIndex)
	}
	if job.Errors[0].URL != "Patient/2" {
		t.Errorf("expected error URL 'Patient/2', got %q", job.Errors[0].URL)
	}
}

func TestProcessBatchAsync_EmptyBundle(t *testing.T) {
	proc, store := newTestProcessor()
	bundle := &TransactionBundle{
		Type:    "batch",
		Entries: nil,
	}

	jobID, _ := proc.SubmitAsyncBatch(bundle, "POST /fhir")
	proc.ProcessBatchAsync(jobID, bundle)

	job, _ := store.GetJob(context.Background(), jobID)
	if job.Status != "completed" {
		t.Errorf("expected status 'completed', got %q", job.Status)
	}
	if job.ProcessedEntries != 0 {
		t.Errorf("expected ProcessedEntries 0, got %d", job.ProcessedEntries)
	}
}

// ---------------------------------------------------------------------------
// ProcessTransactionAsync
// ---------------------------------------------------------------------------

func TestProcessTransactionAsync_AllSuccess(t *testing.T) {
	store := NewInMemoryAsyncBatchStore()
	cfg := DefaultAsyncBatchConfig()
	cfg.WorkerCount = 1
	cfg.RetryMaxAttempts = 1
	cfg.RetryDelay = time.Millisecond
	cfg.ProgressInterval = time.Millisecond

	handler := func(method, url string, resource map[string]interface{}) (*BundleEntryResponse, error) {
		return &BundleEntryResponse{
			Status:   "201 Created",
			Location: url + "/new-id",
		}, nil
	}
	proc := NewAsyncBatchProcessor(store, cfg, handler)

	bundle := &TransactionBundle{
		Type: "transaction",
		Entries: []TransactionEntry{
			{FullURL: "urn:uuid:1", Request: BundleEntryRequest{Method: "POST", URL: "Patient"}, Resource: map[string]interface{}{"resourceType": "Patient"}},
			{FullURL: "urn:uuid:2", Request: BundleEntryRequest{Method: "POST", URL: "Observation"}, Resource: map[string]interface{}{"resourceType": "Observation"}},
		},
	}

	jobID, _ := proc.SubmitAsyncBatch(bundle, "POST /fhir")
	proc.ProcessTransactionAsync(jobID, bundle)

	job, _ := store.GetJob(context.Background(), jobID)
	if job.Status != "completed" {
		t.Errorf("expected status 'completed', got %q", job.Status)
	}
	if job.SuccessCount != 2 {
		t.Errorf("expected SuccessCount 2, got %d", job.SuccessCount)
	}
	if job.ErrorCount != 0 {
		t.Errorf("expected ErrorCount 0, got %d", job.ErrorCount)
	}
}

func TestProcessTransactionAsync_FailureCausesRollback(t *testing.T) {
	store := NewInMemoryAsyncBatchStore()
	cfg := DefaultAsyncBatchConfig()
	cfg.WorkerCount = 1
	cfg.RetryMaxAttempts = 1
	cfg.RetryDelay = time.Millisecond
	cfg.ProgressInterval = time.Millisecond

	callCount := 0
	handler := func(method, url string, resource map[string]interface{}) (*BundleEntryResponse, error) {
		callCount++
		if callCount == 2 {
			return nil, fmt.Errorf("constraint violation")
		}
		return &BundleEntryResponse{
			Status:   "201 Created",
			Location: url + "/new-id",
		}, nil
	}
	proc := NewAsyncBatchProcessor(store, cfg, handler)

	bundle := &TransactionBundle{
		Type: "transaction",
		Entries: []TransactionEntry{
			{FullURL: "urn:uuid:1", Request: BundleEntryRequest{Method: "POST", URL: "Patient"}, Resource: map[string]interface{}{"resourceType": "Patient"}},
			{FullURL: "urn:uuid:2", Request: BundleEntryRequest{Method: "POST", URL: "Observation"}, Resource: map[string]interface{}{"resourceType": "Observation"}},
			{FullURL: "urn:uuid:3", Request: BundleEntryRequest{Method: "POST", URL: "Condition"}, Resource: map[string]interface{}{"resourceType": "Condition"}},
		},
	}

	jobID, _ := proc.SubmitAsyncBatch(bundle, "POST /fhir")
	proc.ProcessTransactionAsync(jobID, bundle)

	job, _ := store.GetJob(context.Background(), jobID)
	if job.Status != "failed" {
		t.Errorf("expected status 'failed', got %q", job.Status)
	}
	if len(job.Errors) == 0 {
		t.Error("expected errors to be recorded")
	}
}

// ---------------------------------------------------------------------------
// CancelJob
// ---------------------------------------------------------------------------

func TestCancelJob_Running(t *testing.T) {
	proc, store := newTestProcessor()
	ctx := context.Background()

	job := &AsyncBatchJob{BundleType: "batch", TotalEntries: 100, Status: "processing"}
	_ = store.CreateJob(ctx, job)

	err := proc.CancelJob(job.ID)
	if err != nil {
		t.Fatalf("CancelJob failed: %v", err)
	}

	got, _ := store.GetJob(ctx, job.ID)
	if got.Status != "cancelled" {
		t.Errorf("expected status 'cancelled', got %q", got.Status)
	}
}

func TestCancelJob_Completed(t *testing.T) {
	proc, store := newTestProcessor()
	ctx := context.Background()

	job := &AsyncBatchJob{BundleType: "batch", TotalEntries: 10, Status: "completed"}
	_ = store.CreateJob(ctx, job)
	job.Status = "completed"
	_ = store.UpdateJob(ctx, job)

	err := proc.CancelJob(job.ID)
	if err == nil {
		t.Fatal("expected error cancelling completed job")
	}
}

func TestCancelJob_NotFound(t *testing.T) {
	proc, _ := newTestProcessor()
	err := proc.CancelJob("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent job")
	}
}

// ---------------------------------------------------------------------------
// GetJobProgress
// ---------------------------------------------------------------------------

func TestGetJobProgress_InProgress(t *testing.T) {
	proc, store := newTestProcessor()
	ctx := context.Background()

	job := &AsyncBatchJob{
		BundleType:       "batch",
		TotalEntries:     100,
		ProcessedEntries: 50,
		SuccessCount:     45,
		ErrorCount:       5,
		Progress:         0.5,
	}
	_ = store.CreateJob(ctx, job)
	job.Status = "processing"
	job.ProcessedEntries = 50
	job.SuccessCount = 45
	job.ErrorCount = 5
	job.Progress = 0.5
	_ = store.UpdateJob(ctx, job)

	progress, err := proc.GetJobProgress(job.ID)
	if err != nil {
		t.Fatalf("GetJobProgress failed: %v", err)
	}
	if progress.JobID != job.ID {
		t.Errorf("expected JobID %q, got %q", job.ID, progress.JobID)
	}
	if progress.Processed != 50 {
		t.Errorf("expected Processed 50, got %d", progress.Processed)
	}
	if progress.Total != 100 {
		t.Errorf("expected Total 100, got %d", progress.Total)
	}
	if progress.SuccessCount != 45 {
		t.Errorf("expected SuccessCount 45, got %d", progress.SuccessCount)
	}
	if progress.ErrorCount != 5 {
		t.Errorf("expected ErrorCount 5, got %d", progress.ErrorCount)
	}
	if progress.Progress != 0.5 {
		t.Errorf("expected Progress 0.5, got %f", progress.Progress)
	}
}

func TestGetJobProgress_Completed(t *testing.T) {
	proc, store := newTestProcessor()
	ctx := context.Background()

	job := &AsyncBatchJob{
		BundleType:       "batch",
		TotalEntries:     10,
		ProcessedEntries: 10,
		SuccessCount:     10,
		Progress:         1.0,
	}
	_ = store.CreateJob(ctx, job)
	job.Status = "completed"
	job.ProcessedEntries = 10
	job.SuccessCount = 10
	job.Progress = 1.0
	_ = store.UpdateJob(ctx, job)

	progress, err := proc.GetJobProgress(job.ID)
	if err != nil {
		t.Fatalf("GetJobProgress failed: %v", err)
	}
	if progress.Progress != 1.0 {
		t.Errorf("expected Progress 1.0, got %f", progress.Progress)
	}
}

func TestGetJobProgress_NotFound(t *testing.T) {
	proc, _ := newTestProcessor()
	_, err := proc.GetJobProgress("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent job")
	}
}

// ---------------------------------------------------------------------------
// BuildProgressResponse
// ---------------------------------------------------------------------------

func TestBuildProgressResponse(t *testing.T) {
	progress := &AsyncBatchProgress{
		JobID:             "job-123",
		Processed:         50,
		Total:             100,
		SuccessCount:      45,
		ErrorCount:        5,
		Progress:          0.5,
		EstimatedTimeLeft: 30 * time.Second,
	}

	result := BuildProgressResponse(progress)

	rt, ok := result["resourceType"].(string)
	if !ok || rt != "OperationOutcome" {
		t.Errorf("expected resourceType OperationOutcome, got %v", result["resourceType"])
	}

	issues, ok := result["issue"].([]map[string]interface{})
	if !ok || len(issues) == 0 {
		t.Fatal("expected at least one issue in progress response")
	}

	// Check severity is information
	if issues[0]["severity"] != "information" {
		t.Errorf("expected severity 'information', got %v", issues[0]["severity"])
	}
}

// ---------------------------------------------------------------------------
// BuildAsyncBatchResult
// ---------------------------------------------------------------------------

func TestBuildAsyncBatchResult(t *testing.T) {
	now := time.Now()
	job := &AsyncBatchJob{
		ID:               "job-456",
		Status:           "completed",
		BundleType:       "batch",
		TotalEntries:     3,
		ProcessedEntries: 3,
		SuccessCount:     2,
		ErrorCount:       1,
		EndTime:          &now,
		ResultBundle: map[string]interface{}{
			"resourceType": "Bundle",
			"type":         "batch-response",
		},
	}

	result := BuildAsyncBatchResult(job)
	if result == nil {
		t.Fatal("expected non-nil result")
	}

	resType, _ := result["resourceType"].(string)
	if resType != "Bundle" {
		t.Errorf("expected resourceType 'Bundle', got %q", resType)
	}
}

func TestBuildAsyncBatchResult_NilResultBundle(t *testing.T) {
	job := &AsyncBatchJob{
		ID:     "job-789",
		Status: "completed",
	}

	result := BuildAsyncBatchResult(job)
	if result == nil {
		t.Fatal("expected non-nil result even for nil ResultBundle")
	}
}

// ---------------------------------------------------------------------------
// RetryWithBackoff
// ---------------------------------------------------------------------------

func TestRetryWithBackoff_SuccessFirstTry(t *testing.T) {
	attempts := 0
	err := RetryWithBackoff(3, time.Millisecond, func() error {
		attempts++
		return nil
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if attempts != 1 {
		t.Errorf("expected 1 attempt, got %d", attempts)
	}
}

func TestRetryWithBackoff_SuccessOnRetry(t *testing.T) {
	attempts := 0
	err := RetryWithBackoff(3, time.Millisecond, func() error {
		attempts++
		if attempts < 3 {
			return fmt.Errorf("transient error")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if attempts != 3 {
		t.Errorf("expected 3 attempts, got %d", attempts)
	}
}

func TestRetryWithBackoff_AllFail(t *testing.T) {
	attempts := 0
	err := RetryWithBackoff(3, time.Millisecond, func() error {
		attempts++
		return fmt.Errorf("persistent error")
	})
	if err == nil {
		t.Fatal("expected error after all retries exhausted")
	}
	if attempts != 3 {
		t.Errorf("expected 3 attempts, got %d", attempts)
	}
	if !strings.Contains(err.Error(), "persistent error") {
		t.Errorf("expected error to contain 'persistent error', got %q", err.Error())
	}
}

func TestRetryWithBackoff_SingleAttempt(t *testing.T) {
	err := RetryWithBackoff(1, time.Millisecond, func() error {
		return fmt.Errorf("fail")
	})
	if err == nil {
		t.Fatal("expected error with single attempt failure")
	}
}

// ---------------------------------------------------------------------------
// ProcessEntryWithRetry
// ---------------------------------------------------------------------------

func TestProcessEntryWithRetry_Success(t *testing.T) {
	cfg := DefaultAsyncBatchConfig()
	cfg.RetryMaxAttempts = 3
	cfg.RetryDelay = time.Millisecond

	entry := &TransactionEntry{
		Request:  BundleEntryRequest{Method: "GET", URL: "Patient/1"},
		Resource: map[string]interface{}{"resourceType": "Patient"},
	}

	handler := func(method, url string, resource map[string]interface{}) (*BundleEntryResponse, error) {
		return &BundleEntryResponse{
			Status:   "200 OK",
			Location: url,
		}, nil
	}

	resp, err := ProcessEntryWithRetry(entry, handler, &cfg)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.Status != "200 OK" {
		t.Errorf("expected status '200 OK', got %q", resp.Status)
	}
}

func TestProcessEntryWithRetry_FailThenSucceed(t *testing.T) {
	cfg := DefaultAsyncBatchConfig()
	cfg.RetryMaxAttempts = 3
	cfg.RetryDelay = time.Millisecond

	entry := &TransactionEntry{
		Request: BundleEntryRequest{Method: "POST", URL: "Patient"},
	}

	var attempts int32
	handler := func(method, url string, resource map[string]interface{}) (*BundleEntryResponse, error) {
		count := atomic.AddInt32(&attempts, 1)
		if count < 2 {
			return nil, fmt.Errorf("temporary failure")
		}
		return &BundleEntryResponse{
			Status:   "201 Created",
			Location: "Patient/new",
		}, nil
	}

	resp, err := ProcessEntryWithRetry(entry, handler, &cfg)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.Status != "201 Created" {
		t.Errorf("expected status '201 Created', got %q", resp.Status)
	}
}

func TestProcessEntryWithRetry_AllFail(t *testing.T) {
	cfg := DefaultAsyncBatchConfig()
	cfg.RetryMaxAttempts = 2
	cfg.RetryDelay = time.Millisecond

	entry := &TransactionEntry{
		Request: BundleEntryRequest{Method: "PUT", URL: "Patient/1"},
	}

	handler := func(method, url string, resource map[string]interface{}) (*BundleEntryResponse, error) {
		return nil, fmt.Errorf("persistent failure")
	}

	_, err := ProcessEntryWithRetry(entry, handler, &cfg)
	if err == nil {
		t.Fatal("expected error after all retries exhausted")
	}
}

// ---------------------------------------------------------------------------
// HTTP Handler Tests
// ---------------------------------------------------------------------------

func newAsyncBatchEchoContext(method, path string, body string, headers map[string]string) (echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	var req *http.Request
	if body != "" {
		req = httptest.NewRequest(method, path, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/fhir+json")
	} else {
		req = httptest.NewRequest(method, path, nil)
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	return c, rec
}

// ---------------------------------------------------------------------------
// AsyncBatchHandler
// ---------------------------------------------------------------------------

func TestAsyncBatchHandler_SyncPath(t *testing.T) {
	proc, _ := newTestProcessor()
	cfg := DefaultAsyncBatchConfig()

	// Small bundle without prefer-async should be processed synchronously (passed through)
	bundle := map[string]interface{}{
		"resourceType": "Bundle",
		"type":         "batch",
		"entry": []map[string]interface{}{
			{
				"request": map[string]interface{}{
					"method": "GET",
					"url":    "Patient/1",
				},
			},
		},
	}
	body, _ := json.Marshal(bundle)

	c, rec := newAsyncBatchEchoContext(http.MethodPost, "/fhir", string(body), nil)
	handler := AsyncBatchHandler(proc, &cfg)
	err := handler(c)

	// Sync path should return nil (pass through) or a non-202 response
	// since the bundle is small enough for sync processing.
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// The sync path handler should NOT return 202 Accepted
	if rec.Code == http.StatusAccepted {
		// Check if Content-Location header was set; if not, this is the sync path
		if rec.Header().Get("Content-Location") != "" {
			t.Error("small bundle without prefer-async should not be processed asynchronously")
		}
	}
}

func TestAsyncBatchHandler_AsyncPath(t *testing.T) {
	proc, _ := newTestProcessor()
	cfg := DefaultAsyncBatchConfig()
	cfg.MaxSyncEntries = 1 // Force async for 2+ entries

	bundle := map[string]interface{}{
		"resourceType": "Bundle",
		"type":         "batch",
		"entry": []map[string]interface{}{
			{"request": map[string]interface{}{"method": "GET", "url": "Patient/1"}},
			{"request": map[string]interface{}{"method": "GET", "url": "Patient/2"}},
		},
	}
	body, _ := json.Marshal(bundle)

	c, rec := newAsyncBatchEchoContext(http.MethodPost, "/fhir", string(body), map[string]string{
		"Prefer": "respond-async",
	})
	handler := AsyncBatchHandler(proc, &cfg)
	err := handler(c)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusAccepted {
		t.Errorf("expected 202 Accepted, got %d", rec.Code)
	}
	loc := rec.Header().Get("Content-Location")
	if loc == "" {
		t.Error("expected Content-Location header")
	}
}

func TestAsyncBatchHandler_InvalidBundle(t *testing.T) {
	proc, _ := newTestProcessor()
	cfg := DefaultAsyncBatchConfig()

	c, rec := newAsyncBatchEchoContext(http.MethodPost, "/fhir", "not json", map[string]string{
		"Prefer": "respond-async",
	})
	handler := AsyncBatchHandler(proc, &cfg)
	err := handler(c)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestAsyncBatchHandler_PreferAsyncSmallBundle(t *testing.T) {
	proc, _ := newTestProcessor()
	cfg := DefaultAsyncBatchConfig()

	bundle := map[string]interface{}{
		"resourceType": "Bundle",
		"type":         "batch",
		"entry": []map[string]interface{}{
			{"request": map[string]interface{}{"method": "GET", "url": "Patient/1"}},
		},
	}
	body, _ := json.Marshal(bundle)

	c, rec := newAsyncBatchEchoContext(http.MethodPost, "/fhir", string(body), map[string]string{
		"Prefer": "respond-async",
	})
	handler := AsyncBatchHandler(proc, &cfg)
	err := handler(c)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// With prefer-async, even small bundles should go async
	if rec.Code != http.StatusAccepted {
		t.Errorf("expected 202 Accepted for prefer-async, got %d", rec.Code)
	}
}

// ---------------------------------------------------------------------------
// AsyncBatchStatusHandler
// ---------------------------------------------------------------------------

func TestAsyncBatchStatusHandler_Queued(t *testing.T) {
	store := NewInMemoryAsyncBatchStore()
	ctx := context.Background()

	job := &AsyncBatchJob{BundleType: "batch", TotalEntries: 100}
	_ = store.CreateJob(ctx, job)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/fhir/_async-batch/"+job.ID, nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("jobId")
	c.SetParamValues(job.ID)

	handler := AsyncBatchStatusHandler(store)
	if err := handler(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusAccepted {
		t.Errorf("expected 202, got %d", rec.Code)
	}
	xProgress := rec.Header().Get("X-Progress")
	if xProgress == "" {
		t.Error("expected X-Progress header")
	}
}

func TestAsyncBatchStatusHandler_InProgress(t *testing.T) {
	store := NewInMemoryAsyncBatchStore()
	ctx := context.Background()

	job := &AsyncBatchJob{
		BundleType:       "batch",
		TotalEntries:     100,
		ProcessedEntries: 50,
		Progress:         0.5,
	}
	_ = store.CreateJob(ctx, job)
	job.Status = "processing"
	job.ProcessedEntries = 50
	job.Progress = 0.5
	_ = store.UpdateJob(ctx, job)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/fhir/_async-batch/"+job.ID, nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("jobId")
	c.SetParamValues(job.ID)

	handler := AsyncBatchStatusHandler(store)
	if err := handler(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusAccepted {
		t.Errorf("expected 202, got %d", rec.Code)
	}
	xProgress := rec.Header().Get("X-Progress")
	if xProgress == "" {
		t.Error("expected X-Progress header")
	}
}

func TestAsyncBatchStatusHandler_Completed(t *testing.T) {
	store := NewInMemoryAsyncBatchStore()
	ctx := context.Background()

	now := time.Now()
	job := &AsyncBatchJob{
		BundleType:       "batch",
		TotalEntries:     10,
		ProcessedEntries: 10,
		SuccessCount:     10,
		Progress:         1.0,
		ResultBundle: map[string]interface{}{
			"resourceType": "Bundle",
			"type":         "batch-response",
		},
	}
	_ = store.CreateJob(ctx, job)
	job.Status = "completed"
	job.EndTime = &now
	job.ProcessedEntries = 10
	job.SuccessCount = 10
	job.Progress = 1.0
	_ = store.UpdateJob(ctx, job)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/fhir/_async-batch/"+job.ID, nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("jobId")
	c.SetParamValues(job.ID)

	handler := AsyncBatchStatusHandler(store)
	if err := handler(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var body map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to parse body: %v", err)
	}
	if body["resourceType"] != "Bundle" {
		t.Errorf("expected resourceType Bundle, got %v", body["resourceType"])
	}
}

func TestAsyncBatchStatusHandler_Failed(t *testing.T) {
	store := NewInMemoryAsyncBatchStore()
	ctx := context.Background()

	job := &AsyncBatchJob{BundleType: "transaction", TotalEntries: 10}
	_ = store.CreateJob(ctx, job)
	job.Status = "failed"
	job.Errors = []AsyncBatchError{{EntryIndex: 0, Diagnostics: "constraint violation"}}
	_ = store.UpdateJob(ctx, job)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/fhir/_async-batch/"+job.ID, nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("jobId")
	c.SetParamValues(job.ID)

	handler := AsyncBatchStatusHandler(store)
	if err := handler(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "OperationOutcome") {
		t.Error("expected OperationOutcome in error response")
	}
}

func TestAsyncBatchStatusHandler_NotFound(t *testing.T) {
	store := NewInMemoryAsyncBatchStore()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/fhir/_async-batch/missing", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("jobId")
	c.SetParamValues("missing")

	handler := AsyncBatchStatusHandler(store)
	if err := handler(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

// ---------------------------------------------------------------------------
// AsyncBatchCancelHandler
// ---------------------------------------------------------------------------

func TestAsyncBatchCancelHandler_Success(t *testing.T) {
	proc, store := newTestProcessor()
	ctx := context.Background()

	job := &AsyncBatchJob{BundleType: "batch", TotalEntries: 100}
	_ = store.CreateJob(ctx, job)
	job.Status = "processing"
	_ = store.UpdateJob(ctx, job)

	e := echo.New()
	req := httptest.NewRequest(http.MethodDelete, "/fhir/_async-batch/"+job.ID, nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("jobId")
	c.SetParamValues(job.ID)

	handler := AsyncBatchCancelHandler(proc)
	if err := handler(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusAccepted {
		t.Errorf("expected 202, got %d", rec.Code)
	}

	got, _ := store.GetJob(ctx, job.ID)
	if got.Status != "cancelled" {
		t.Errorf("expected status 'cancelled', got %q", got.Status)
	}
}

func TestAsyncBatchCancelHandler_NotFound(t *testing.T) {
	proc, _ := newTestProcessor()

	e := echo.New()
	req := httptest.NewRequest(http.MethodDelete, "/fhir/_async-batch/missing", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("jobId")
	c.SetParamValues("missing")

	handler := AsyncBatchCancelHandler(proc)
	if err := handler(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestAsyncBatchCancelHandler_MissingParam(t *testing.T) {
	proc, _ := newTestProcessor()

	e := echo.New()
	req := httptest.NewRequest(http.MethodDelete, "/fhir/_async-batch/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("jobId")
	c.SetParamValues("")

	handler := AsyncBatchCancelHandler(proc)
	if err := handler(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

// ---------------------------------------------------------------------------
// AsyncBatchListHandler
// ---------------------------------------------------------------------------

func TestAsyncBatchListHandler_Empty(t *testing.T) {
	store := NewInMemoryAsyncBatchStore()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/fhir/_async-batch", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := AsyncBatchListHandler(store)
	if err := handler(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var body map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to parse body: %v", err)
	}
	jobs, ok := body["jobs"].([]interface{})
	if !ok {
		t.Fatal("expected 'jobs' array in response")
	}
	if len(jobs) != 0 {
		t.Errorf("expected 0 jobs, got %d", len(jobs))
	}
}

func TestAsyncBatchListHandler_WithJobs(t *testing.T) {
	store := NewInMemoryAsyncBatchStore()
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		job := &AsyncBatchJob{BundleType: "batch", TotalEntries: 10}
		_ = store.CreateJob(ctx, job)
	}

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/fhir/_async-batch", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := AsyncBatchListHandler(store)
	if err := handler(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var body map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to parse body: %v", err)
	}
	jobs, ok := body["jobs"].([]interface{})
	if !ok {
		t.Fatal("expected 'jobs' array in response")
	}
	if len(jobs) != 3 {
		t.Errorf("expected 3 jobs, got %d", len(jobs))
	}
}

func TestAsyncBatchListHandler_FilteredByStatus(t *testing.T) {
	store := NewInMemoryAsyncBatchStore()
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		job := &AsyncBatchJob{BundleType: "batch", TotalEntries: 10}
		_ = store.CreateJob(ctx, job)
	}
	// Update one to processing
	jobs, _ := store.ListJobs(ctx, "", 100)
	jobs[0].Status = "processing"
	_ = store.UpdateJob(ctx, jobs[0])

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/fhir/_async-batch?status=processing", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := AsyncBatchListHandler(store)
	if err := handler(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var body map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to parse body: %v", err)
	}
	jobsList, ok := body["jobs"].([]interface{})
	if !ok {
		t.Fatal("expected 'jobs' array in response")
	}
	if len(jobsList) != 1 {
		t.Errorf("expected 1 processing job, got %d", len(jobsList))
	}
}

// ---------------------------------------------------------------------------
// Edge Cases
// ---------------------------------------------------------------------------

func TestAsyncBatchProcessor_ConcurrentSubmissions(t *testing.T) {
	proc, store := newTestProcessor()

	var wg sync.WaitGroup
	jobIDs := make([]string, 10)
	errors := make([]error, 10)

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			bundle := &TransactionBundle{
				Type: "batch",
				Entries: []TransactionEntry{
					{Request: BundleEntryRequest{Method: "GET", URL: fmt.Sprintf("Patient/%d", idx)}},
				},
			}
			id, err := proc.SubmitAsyncBatch(bundle, "POST /fhir")
			jobIDs[idx] = id
			errors[idx] = err
		}(i)
	}
	wg.Wait()

	for i, err := range errors {
		if err != nil {
			t.Errorf("submission %d failed: %v", i, err)
		}
	}

	// Verify all jobs were created
	allJobs, _ := store.ListJobs(context.Background(), "", 100)
	if len(allJobs) != 10 {
		t.Errorf("expected 10 jobs, got %d", len(allJobs))
	}

	// Verify all IDs are unique
	seen := make(map[string]bool)
	for _, id := range jobIDs {
		if seen[id] {
			t.Errorf("duplicate job ID: %s", id)
		}
		seen[id] = true
	}
}

func TestAsyncBatchProcessor_LargeBatch(t *testing.T) {
	store := NewInMemoryAsyncBatchStore()
	cfg := DefaultAsyncBatchConfig()
	cfg.WorkerCount = 4
	cfg.RetryMaxAttempts = 1
	cfg.RetryDelay = time.Millisecond
	cfg.ProgressInterval = time.Millisecond
	cfg.MaxAsyncEntries = 500

	var processCount int32
	handler := func(method, url string, resource map[string]interface{}) (*BundleEntryResponse, error) {
		atomic.AddInt32(&processCount, 1)
		return &BundleEntryResponse{Status: "200 OK", Location: url}, nil
	}
	proc := NewAsyncBatchProcessor(store, cfg, handler)

	entries := make([]TransactionEntry, 200)
	for i := range entries {
		entries[i] = TransactionEntry{
			Request: BundleEntryRequest{Method: "GET", URL: fmt.Sprintf("Patient/%d", i)},
		}
	}
	bundle := &TransactionBundle{Type: "batch", Entries: entries}

	jobID, err := proc.SubmitAsyncBatch(bundle, "POST /fhir")
	if err != nil {
		t.Fatalf("SubmitAsyncBatch failed: %v", err)
	}

	proc.ProcessBatchAsync(jobID, bundle)

	job, _ := store.GetJob(context.Background(), jobID)
	if job.Status != "completed" {
		t.Errorf("expected status 'completed', got %q", job.Status)
	}
	if job.ProcessedEntries != 200 {
		t.Errorf("expected 200 processed entries, got %d", job.ProcessedEntries)
	}
	if job.SuccessCount != 200 {
		t.Errorf("expected 200 successes, got %d", job.SuccessCount)
	}
	if atomic.LoadInt32(&processCount) != 200 {
		t.Errorf("expected handler called 200 times, got %d", processCount)
	}
}

func TestAsyncBatchJob_ProgressCalculation(t *testing.T) {
	store := NewInMemoryAsyncBatchStore()
	cfg := DefaultAsyncBatchConfig()
	cfg.WorkerCount = 1
	cfg.RetryMaxAttempts = 1
	cfg.RetryDelay = time.Millisecond
	cfg.ProgressInterval = time.Millisecond

	var mu sync.Mutex
	progressValues := []float64{}

	handler := func(method, url string, resource map[string]interface{}) (*BundleEntryResponse, error) {
		return &BundleEntryResponse{Status: "200 OK", Location: url}, nil
	}
	proc := NewAsyncBatchProcessor(store, cfg, handler)

	entries := make([]TransactionEntry, 10)
	for i := range entries {
		entries[i] = TransactionEntry{
			Request: BundleEntryRequest{Method: "GET", URL: fmt.Sprintf("Patient/%d", i)},
		}
	}
	bundle := &TransactionBundle{Type: "batch", Entries: entries}

	jobID, _ := proc.SubmitAsyncBatch(bundle, "POST /fhir")
	proc.ProcessBatchAsync(jobID, bundle)

	job, _ := store.GetJob(context.Background(), jobID)
	mu.Lock()
	progressValues = append(progressValues, job.Progress)
	mu.Unlock()

	if job.Progress != 1.0 {
		t.Errorf("expected final progress 1.0, got %f", job.Progress)
	}
}

func TestAsyncBatchStore_ConcurrentReadWrite(t *testing.T) {
	store := NewInMemoryAsyncBatchStore()
	ctx := context.Background()

	job := &AsyncBatchJob{BundleType: "batch", TotalEntries: 100}
	_ = store.CreateJob(ctx, job)

	var wg sync.WaitGroup

	// Concurrent readers
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				_, _ = store.GetJob(ctx, job.ID)
			}
		}()
	}

	// Concurrent writers
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			for j := 0; j < 20; j++ {
				updated := &AsyncBatchJob{
					ID:               job.ID,
					BundleType:       "batch",
					TotalEntries:     100,
					ProcessedEntries: idx*20 + j,
					Status:           "processing",
				}
				_ = store.UpdateJob(ctx, updated)
			}
		}(i)
	}

	wg.Wait()

	// Should not panic or deadlock
	got, err := store.GetJob(ctx, job.ID)
	if err != nil {
		t.Fatalf("final GetJob failed: %v", err)
	}
	if got.Status != "processing" {
		t.Errorf("expected status 'processing', got %q", got.Status)
	}
}

func TestAsyncBatchHandler_NonBundleResourceType(t *testing.T) {
	proc, _ := newTestProcessor()
	cfg := DefaultAsyncBatchConfig()

	body := `{"resourceType":"Patient","name":[{"family":"Smith"}]}`

	c, rec := newAsyncBatchEchoContext(http.MethodPost, "/fhir", body, map[string]string{
		"Prefer": "respond-async",
	})
	handler := AsyncBatchHandler(proc, &cfg)
	err := handler(c)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for non-Bundle resource, got %d", rec.Code)
	}
}

func TestAsyncBatchHandler_TransactionBundle(t *testing.T) {
	proc, _ := newTestProcessor()
	cfg := DefaultAsyncBatchConfig()
	cfg.MaxSyncEntries = 1

	bundle := map[string]interface{}{
		"resourceType": "Bundle",
		"type":         "transaction",
		"entry": []map[string]interface{}{
			{
				"fullUrl":  "urn:uuid:1",
				"request":  map[string]interface{}{"method": "POST", "url": "Patient"},
				"resource": map[string]interface{}{"resourceType": "Patient"},
			},
			{
				"fullUrl":  "urn:uuid:2",
				"request":  map[string]interface{}{"method": "POST", "url": "Observation"},
				"resource": map[string]interface{}{"resourceType": "Observation"},
			},
		},
	}
	body, _ := json.Marshal(bundle)

	c, rec := newAsyncBatchEchoContext(http.MethodPost, "/fhir", string(body), map[string]string{
		"Prefer": "respond-async",
	})
	handler := AsyncBatchHandler(proc, &cfg)
	err := handler(c)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusAccepted {
		t.Errorf("expected 202 Accepted, got %d", rec.Code)
	}
	loc := rec.Header().Get("Content-Location")
	if loc == "" {
		t.Error("expected Content-Location header for async transaction")
	}
}

func TestAsyncBatchStatusHandler_CancelledJob(t *testing.T) {
	store := NewInMemoryAsyncBatchStore()
	ctx := context.Background()

	job := &AsyncBatchJob{BundleType: "batch", TotalEntries: 50}
	_ = store.CreateJob(ctx, job)
	_ = store.CancelJob(ctx, job.ID)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/fhir/_async-batch/"+job.ID, nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("jobId")
	c.SetParamValues(job.ID)

	handler := AsyncBatchStatusHandler(store)
	if err := handler(c); err != nil {
		t.Fatal(err)
	}
	// Cancelled jobs should return an error status
	body := rec.Body.String()
	if !strings.Contains(body, "cancelled") && rec.Code == http.StatusOK {
		t.Error("expected cancelled status in response")
	}
}

func TestAsyncBatchListHandler_WithLimit(t *testing.T) {
	store := NewInMemoryAsyncBatchStore()
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		job := &AsyncBatchJob{BundleType: "batch", TotalEntries: 10}
		_ = store.CreateJob(ctx, job)
	}

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/fhir/_async-batch?_count=2", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := AsyncBatchListHandler(store)
	if err := handler(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var body map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to parse body: %v", err)
	}
	jobs, ok := body["jobs"].([]interface{})
	if !ok {
		t.Fatal("expected 'jobs' array in response")
	}
	if len(jobs) != 2 {
		t.Errorf("expected 2 jobs with limit, got %d", len(jobs))
	}
}
