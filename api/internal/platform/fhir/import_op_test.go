package fhir

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
)

// ===========================================================================
// ValidateImportRequest tests
// ===========================================================================

func TestValidateImportRequest_ValidRequest(t *testing.T) {
	req := &ImportRequest{
		InputFormat: "application/fhir+ndjson",
		InputSource: "https://example.com/data",
		Input: []ImportInput{
			{Type: "Patient", URL: "https://example.com/Patient.ndjson"},
			{Type: "Observation", URL: "https://example.com/Observation.ndjson"},
		},
	}

	issues := ValidateImportRequest(req)
	if len(issues) != 0 {
		t.Errorf("expected 0 issues for a valid request, got %d: %+v", len(issues), issues)
	}
}

func TestValidateImportRequest_AlternateFormats(t *testing.T) {
	for _, format := range []string{"application/ndjson", "ndjson"} {
		req := &ImportRequest{
			InputFormat: format,
			Input: []ImportInput{
				{Type: "Patient", URL: "https://example.com/Patient.ndjson"},
			},
		}
		issues := ValidateImportRequest(req)
		if len(issues) != 0 {
			t.Errorf("expected 0 issues for format %q, got %d: %+v", format, len(issues), issues)
		}
	}
}

func TestValidateImportRequest_MissingInputFormat(t *testing.T) {
	req := &ImportRequest{
		Input: []ImportInput{
			{Type: "Patient", URL: "https://example.com/Patient.ndjson"},
		},
	}

	issues := ValidateImportRequest(req)
	if len(issues) == 0 {
		t.Fatal("expected validation issues for missing inputFormat")
	}

	found := false
	for _, issue := range issues {
		if issue.Location == "inputFormat" && issue.Code == VIssueTypeRequired {
			found = true
		}
	}
	if !found {
		t.Error("expected a 'required' issue for inputFormat")
	}
}

func TestValidateImportRequest_UnsupportedInputFormat(t *testing.T) {
	req := &ImportRequest{
		InputFormat: "text/csv",
		Input: []ImportInput{
			{Type: "Patient", URL: "https://example.com/Patient.ndjson"},
		},
	}

	issues := ValidateImportRequest(req)
	if len(issues) == 0 {
		t.Fatal("expected validation issues for unsupported inputFormat")
	}

	found := false
	for _, issue := range issues {
		if issue.Location == "inputFormat" && issue.Code == VIssueTypeValue {
			found = true
		}
	}
	if !found {
		t.Error("expected a 'value' issue for unsupported inputFormat")
	}
}

func TestValidateImportRequest_EmptyInput(t *testing.T) {
	req := &ImportRequest{
		InputFormat: "application/fhir+ndjson",
		Input:       []ImportInput{},
	}

	issues := ValidateImportRequest(req)
	if len(issues) == 0 {
		t.Fatal("expected validation issues for empty input array")
	}

	found := false
	for _, issue := range issues {
		if issue.Location == "input" && issue.Code == VIssueTypeRequired {
			found = true
		}
	}
	if !found {
		t.Error("expected a 'required' issue for input")
	}
}

func TestValidateImportRequest_MissingType(t *testing.T) {
	req := &ImportRequest{
		InputFormat: "application/fhir+ndjson",
		Input: []ImportInput{
			{Type: "", URL: "https://example.com/Patient.ndjson"},
		},
	}

	issues := ValidateImportRequest(req)
	if len(issues) == 0 {
		t.Fatal("expected validation issues for missing type")
	}

	found := false
	for _, issue := range issues {
		if issue.Location == "input[0].type" && issue.Code == VIssueTypeRequired {
			found = true
		}
	}
	if !found {
		t.Error("expected a 'required' issue for input[0].type")
	}
}

func TestValidateImportRequest_InvalidResourceType(t *testing.T) {
	req := &ImportRequest{
		InputFormat: "application/fhir+ndjson",
		Input: []ImportInput{
			{Type: "FakeResource", URL: "https://example.com/FakeResource.ndjson"},
		},
	}

	issues := ValidateImportRequest(req)
	if len(issues) == 0 {
		t.Fatal("expected validation issues for invalid resource type")
	}

	found := false
	for _, issue := range issues {
		if issue.Location == "input[0].type" && issue.Code == VIssueTypeValue {
			found = true
		}
	}
	if !found {
		t.Error("expected a 'value' issue for invalid resource type")
	}
}

func TestValidateImportRequest_MissingURL(t *testing.T) {
	req := &ImportRequest{
		InputFormat: "application/fhir+ndjson",
		Input: []ImportInput{
			{Type: "Patient", URL: ""},
		},
	}

	issues := ValidateImportRequest(req)
	if len(issues) == 0 {
		t.Fatal("expected validation issues for missing URL")
	}

	found := false
	for _, issue := range issues {
		if issue.Location == "input[0].url" && issue.Code == VIssueTypeRequired {
			found = true
		}
	}
	if !found {
		t.Error("expected a 'required' issue for input[0].url")
	}
}

func TestValidateImportRequest_MultipleErrors(t *testing.T) {
	req := &ImportRequest{
		InputFormat: "",
		Input: []ImportInput{
			{Type: "", URL: ""},
			{Type: "InvalidType", URL: "https://example.com/data.ndjson"},
		},
	}

	issues := ValidateImportRequest(req)
	// Expect: inputFormat required, input[0].type required, input[0].url required, input[1].type invalid
	if len(issues) < 4 {
		t.Errorf("expected at least 4 issues, got %d: %+v", len(issues), issues)
	}
}

// ===========================================================================
// ParseNDJSON tests
// ===========================================================================

func TestParseNDJSON_ValidInput(t *testing.T) {
	data := []byte(`{"resourceType":"Patient","id":"1"}
{"resourceType":"Patient","id":"2"}
{"resourceType":"Patient","id":"3"}`)

	results, err := ParseNDJSON(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}

	// Verify each result is valid JSON with the expected id.
	for i, raw := range results {
		var m map[string]interface{}
		if err := json.Unmarshal(raw, &m); err != nil {
			t.Errorf("result[%d] is not valid JSON: %v", i, err)
		}
	}
}

func TestParseNDJSON_EmptyInput(t *testing.T) {
	results, err := ParseNDJSON([]byte{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if results != nil {
		t.Errorf("expected nil for empty input, got %d results", len(results))
	}
}

func TestParseNDJSON_BlankLines(t *testing.T) {
	data := []byte(`{"id":"1"}

{"id":"2"}

`)

	results, err := ParseNDJSON(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 results (blank lines skipped), got %d", len(results))
	}
}

func TestParseNDJSON_InvalidJSON(t *testing.T) {
	data := []byte(`{"id":"1"}
{not valid json}
{"id":"3"}`)

	_, err := ParseNDJSON(data)
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
	if !strings.Contains(err.Error(), "line 2") {
		t.Errorf("expected error to mention line 2, got: %s", err.Error())
	}
}

func TestParseNDJSON_SingleLine(t *testing.T) {
	data := []byte(`{"resourceType":"Observation","id":"obs-1"}`)

	results, err := ParseNDJSON(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}
}

// ===========================================================================
// ImportHandler tests
// ===========================================================================

// newImportEcho creates an echo instance with the $import route registered.
func newImportEcho(store AsyncJobStore) *echo.Echo {
	e := echo.New()
	e.POST("/fhir/$import", ImportHandler(store))
	return e
}

func TestImportHandler_Returns202WithContentLocation(t *testing.T) {
	store := NewInMemoryAsyncJobStore()
	e := newImportEcho(store)

	reqBody := ImportRequest{
		InputFormat: "application/fhir+ndjson",
		InputSource: "https://example.com",
		Input: []ImportInput{
			{Type: "Patient", URL: "https://example.com/Patient.ndjson"},
		},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/fhir/$import", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected status 202, got %d; body: %s", rec.Code, rec.Body.String())
	}

	contentLocation := rec.Header().Get("Content-Location")
	if contentLocation == "" {
		t.Fatal("expected Content-Location header to be set")
	}
	if !strings.HasPrefix(contentLocation, "/_async/") {
		t.Errorf("expected Content-Location to start with '/_async/', got %q", contentLocation)
	}
}

func TestImportHandler_RejectsInvalidRequest_EmptyInput(t *testing.T) {
	store := NewInMemoryAsyncJobStore()
	e := newImportEcho(store)

	reqBody := ImportRequest{
		InputFormat: "application/fhir+ndjson",
		Input:       []ImportInput{},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/fhir/$import", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400 for empty input, got %d", rec.Code)
	}

	var outcome OperationOutcome
	if err := json.Unmarshal(rec.Body.Bytes(), &outcome); err != nil {
		t.Fatalf("failed to parse response body: %v", err)
	}
	if len(outcome.Issue) == 0 {
		t.Error("expected at least one issue in OperationOutcome")
	}
}

func TestImportHandler_RejectsInvalidRequest_BadFormat(t *testing.T) {
	store := NewInMemoryAsyncJobStore()
	e := newImportEcho(store)

	reqBody := ImportRequest{
		InputFormat: "text/csv",
		Input: []ImportInput{
			{Type: "Patient", URL: "https://example.com/Patient.ndjson"},
		},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/fhir/$import", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400 for unsupported format, got %d", rec.Code)
	}
}

func TestImportHandler_RejectsInvalidRequest_InvalidResourceType(t *testing.T) {
	store := NewInMemoryAsyncJobStore()
	e := newImportEcho(store)

	reqBody := ImportRequest{
		InputFormat: "application/fhir+ndjson",
		Input: []ImportInput{
			{Type: "NotARealType", URL: "https://example.com/data.ndjson"},
		},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/fhir/$import", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400 for invalid resource type, got %d", rec.Code)
	}
}

func TestImportHandler_RejectsMalformedJSON(t *testing.T) {
	store := NewInMemoryAsyncJobStore()
	e := newImportEcho(store)

	req := httptest.NewRequest(http.MethodPost, "/fhir/$import", strings.NewReader("{invalid"))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400 for malformed JSON, got %d", rec.Code)
	}
}

func TestImportHandler_CreatesAsyncJob(t *testing.T) {
	store := NewInMemoryAsyncJobStore()
	e := newImportEcho(store)

	reqBody := ImportRequest{
		InputFormat: "application/fhir+ndjson",
		Input: []ImportInput{
			{Type: "Patient", URL: "https://example.com/Patient.ndjson"},
			{Type: "Observation", URL: "https://example.com/Observation.ndjson"},
		},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/fhir/$import", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected status 202, got %d", rec.Code)
	}

	// Extract job ID from Content-Location.
	contentLocation := rec.Header().Get("Content-Location")
	jobID := strings.TrimPrefix(contentLocation, "/_async/")

	// Wait briefly for the mock import goroutine to complete.
	time.Sleep(50 * time.Millisecond)

	// Verify the job exists in the store and has been completed.
	job, err := store.Get(nil, jobID)
	if err != nil {
		t.Fatalf("failed to get job from store: %v", err)
	}
	if job.Status != AsyncStatusCompleted {
		t.Errorf("expected job status %q, got %q", AsyncStatusCompleted, job.Status)
	}
	if len(job.Output) != 2 {
		t.Errorf("expected 2 output entries, got %d", len(job.Output))
	}
}

func TestImportHandler_MultipleInputTypes(t *testing.T) {
	store := NewInMemoryAsyncJobStore()
	e := newImportEcho(store)

	reqBody := ImportRequest{
		InputFormat: "application/fhir+ndjson",
		InputSource: "https://source.example.com",
		Input: []ImportInput{
			{Type: "Patient", URL: "https://example.com/Patient.ndjson"},
			{Type: "Condition", URL: "https://example.com/Condition.ndjson"},
			{Type: "Encounter", URL: "https://example.com/Encounter.ndjson"},
		},
		StorageDetail: &StorageDetail{Type: "https"},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/fhir/$import", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected status 202, got %d", rec.Code)
	}

	// Wait for async processing.
	time.Sleep(50 * time.Millisecond)

	contentLocation := rec.Header().Get("Content-Location")
	jobID := strings.TrimPrefix(contentLocation, "/_async/")

	job, err := store.Get(nil, jobID)
	if err != nil {
		t.Fatalf("failed to get job: %v", err)
	}
	if len(job.Output) != 3 {
		t.Errorf("expected 3 output entries, got %d", len(job.Output))
	}
}
