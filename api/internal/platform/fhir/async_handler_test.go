package fhir

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
)

// ---------------------------------------------------------------------------
// InMemoryAsyncJobStore CRUD
// ---------------------------------------------------------------------------

func TestInMemoryAsyncJobStore_CreateAndGet(t *testing.T) {
	store := NewInMemoryAsyncJobStore()
	ctx := context.Background()

	job := &AsyncJob{
		Request: "GET /fhir/Patient/$export",
	}

	if err := store.Create(ctx, job); err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if job.ID == "" {
		t.Fatal("expected job ID to be generated")
	}
	if job.Status != AsyncStatusInProgress {
		t.Errorf("expected default status %q, got %q", AsyncStatusInProgress, job.Status)
	}
	if job.TransactionTS.IsZero() {
		t.Error("expected TransactionTS to be set")
	}

	got, err := store.Get(ctx, job.ID)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if got.Request != job.Request {
		t.Errorf("expected request %q, got %q", job.Request, got.Request)
	}
}

func TestInMemoryAsyncJobStore_GetNotFound(t *testing.T) {
	store := NewInMemoryAsyncJobStore()
	_, err := store.Get(context.Background(), "does-not-exist")
	if err == nil {
		t.Fatal("expected error for missing job")
	}
}

func TestInMemoryAsyncJobStore_Update(t *testing.T) {
	store := NewInMemoryAsyncJobStore()
	ctx := context.Background()

	job := &AsyncJob{Request: "POST /fhir/$export"}
	if err := store.Create(ctx, job); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	job.Status = AsyncStatusCompleted
	job.Output = []AsyncJobOutput{
		{Type: "Patient", URL: "http://example.com/output/1"},
	}
	if err := store.Update(ctx, job); err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	got, _ := store.Get(ctx, job.ID)
	if got.Status != AsyncStatusCompleted {
		t.Errorf("expected status %q, got %q", AsyncStatusCompleted, got.Status)
	}
	if len(got.Output) != 1 {
		t.Errorf("expected 1 output, got %d", len(got.Output))
	}
}

func TestInMemoryAsyncJobStore_UpdateNotFound(t *testing.T) {
	store := NewInMemoryAsyncJobStore()
	err := store.Update(context.Background(), &AsyncJob{ID: "missing"})
	if err == nil {
		t.Fatal("expected error updating nonexistent job")
	}
}

func TestInMemoryAsyncJobStore_Delete(t *testing.T) {
	store := NewInMemoryAsyncJobStore()
	ctx := context.Background()

	job := &AsyncJob{Request: "GET /fhir/$export"}
	_ = store.Create(ctx, job)

	if err := store.Delete(ctx, job.ID); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	_, err := store.Get(ctx, job.ID)
	if err == nil {
		t.Fatal("expected error after deleting job")
	}
}

func TestInMemoryAsyncJobStore_DeleteNonexistent(t *testing.T) {
	store := NewInMemoryAsyncJobStore()
	// Deleting a nonexistent job should not error.
	if err := store.Delete(context.Background(), "nope"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

// ---------------------------------------------------------------------------
// AsyncStatusHandler
// ---------------------------------------------------------------------------

func newAsyncContext(method, path, jobID string) (echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	req := httptest.NewRequest(method, path, nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("jobId")
	c.SetParamValues(jobID)
	return c, rec
}

func TestAsyncStatusHandler_InProgress(t *testing.T) {
	store := NewInMemoryAsyncJobStore()
	ctx := context.Background()

	job := &AsyncJob{Request: "GET /fhir/$export", Status: AsyncStatusInProgress}
	_ = store.Create(ctx, job)

	c, rec := newAsyncContext(http.MethodGet, "/_async/"+job.ID, job.ID)
	handler := AsyncStatusHandler(store)

	if err := handler(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusAccepted {
		t.Errorf("expected 202, got %d", rec.Code)
	}
	if progress := rec.Header().Get("X-Progress"); progress != "in-progress" {
		t.Errorf("expected X-Progress header 'in-progress', got %q", progress)
	}
}

func TestAsyncStatusHandler_Completed(t *testing.T) {
	store := NewInMemoryAsyncJobStore()
	ctx := context.Background()

	job := &AsyncJob{
		Request: "GET /fhir/$export",
		Status:  AsyncStatusCompleted,
		Output: []AsyncJobOutput{
			{Type: "Patient", URL: "http://example.com/out/1"},
		},
	}
	_ = store.Create(ctx, job)

	c, rec := newAsyncContext(http.MethodGet, "/_async/"+job.ID, job.ID)
	handler := AsyncStatusHandler(store)

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
	if body["request"] != "GET /fhir/$export" {
		t.Errorf("unexpected request in body: %v", body["request"])
	}
	output, ok := body["output"].([]interface{})
	if !ok || len(output) != 1 {
		t.Errorf("expected 1 output entry, got %v", body["output"])
	}
}

func TestAsyncStatusHandler_Error(t *testing.T) {
	store := NewInMemoryAsyncJobStore()
	ctx := context.Background()

	job := &AsyncJob{
		Request: "GET /fhir/$export",
		Status:  AsyncStatusError,
		Error:   "internal processing failure",
	}
	_ = store.Create(ctx, job)

	c, rec := newAsyncContext(http.MethodGet, "/_async/"+job.ID, job.ID)
	handler := AsyncStatusHandler(store)

	if err := handler(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "OperationOutcome") {
		t.Errorf("expected OperationOutcome in body, got %s", body)
	}
	if !strings.Contains(body, "internal processing failure") {
		t.Errorf("expected error message in body, got %s", body)
	}
}

func TestAsyncStatusHandler_NotFound(t *testing.T) {
	store := NewInMemoryAsyncJobStore()

	c, rec := newAsyncContext(http.MethodGet, "/_async/missing", "missing")
	handler := AsyncStatusHandler(store)

	if err := handler(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "OperationOutcome") {
		t.Errorf("expected OperationOutcome in body, got %s", body)
	}
}

// ---------------------------------------------------------------------------
// AsyncDeleteHandler
// ---------------------------------------------------------------------------

func TestAsyncDeleteHandler_RemovesJob(t *testing.T) {
	store := NewInMemoryAsyncJobStore()
	ctx := context.Background()

	job := &AsyncJob{Request: "GET /fhir/$export"}
	_ = store.Create(ctx, job)

	c, rec := newAsyncContext(http.MethodDelete, "/_async/"+job.ID, job.ID)
	handler := AsyncDeleteHandler(store)

	if err := handler(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusAccepted {
		t.Errorf("expected 202, got %d", rec.Code)
	}

	// Verify the job is gone.
	_, err := store.Get(ctx, job.ID)
	if err == nil {
		t.Error("expected job to be deleted")
	}
}

func TestAsyncDeleteHandler_MissingParam(t *testing.T) {
	store := NewInMemoryAsyncJobStore()

	// Provide empty jobId param.
	c, rec := newAsyncContext(http.MethodDelete, "/_async/", "")
	handler := AsyncDeleteHandler(store)

	if err := handler(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

// ---------------------------------------------------------------------------
// RespondAsync
// ---------------------------------------------------------------------------

func TestRespondAsync(t *testing.T) {
	store := NewInMemoryAsyncJobStore()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := RespondAsync(c, store, "job-123"); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusAccepted {
		t.Errorf("expected 202, got %d", rec.Code)
	}
	loc := rec.Header().Get("Content-Location")
	if loc != "/_async/job-123" {
		t.Errorf("expected Content-Location /_async/job-123, got %q", loc)
	}
}

// ---------------------------------------------------------------------------
// ParsePreferAsync
// ---------------------------------------------------------------------------

func TestParsePreferAsync(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"respond-async", true},
		{"respond-async; return=representation", true},
		{"return=minimal, respond-async", true},
		{"respond-async, return=minimal", true},
		{"return=minimal; respond-async", true},
		{"return=minimal", false},
		{"", false},
		{"handling=strict", false},
		{"async", false},
	}
	for _, tt := range tests {
		got := ParsePreferAsync(tt.input)
		if got != tt.want {
			t.Errorf("ParsePreferAsync(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}
