package fhir

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// ---------------------------------------------------------------------------
// Request / response types for the FHIR $import operation.
// ---------------------------------------------------------------------------

// ImportRequest describes the payload for a FHIR Bulk Data $import kick-off.
type ImportRequest struct {
	InputFormat   string         `json:"inputFormat"`             // e.g. "application/fhir+ndjson"
	InputSource   string         `json:"inputSource"`             // URL of the data source
	Input         []ImportInput  `json:"input"`                   // list of NDJSON files to import
	StorageDetail *StorageDetail `json:"storageDetail,omitempty"` // optional storage configuration
}

// ImportInput describes a single input file within an ImportRequest.
type ImportInput struct {
	Type string `json:"type"` // FHIR resource type (e.g. "Patient")
	URL  string `json:"url"`  // URL to the NDJSON file
}

// StorageDetail carries storage-related metadata for the import.
type StorageDetail struct {
	Type string `json:"type"` // e.g. "https"
}

// ImportResult is the response body returned when polling a completed $import job.
type ImportResult struct {
	TransactionTime time.Time      `json:"transactionTime"`
	Request         string         `json:"request"`
	Outcome         []ImportOutcome `json:"outcome"`
	Error           []ImportOutcome `json:"error,omitempty"`
}

// ImportOutcome describes the result for a single resource type within an import.
type ImportOutcome struct {
	Type  string `json:"type"`
	Count int    `json:"count"`
	URL   string `json:"url,omitempty"`
}

// ---------------------------------------------------------------------------
// Supported input formats
// ---------------------------------------------------------------------------

// validImportFormats lists the input formats accepted by the $import operation.
var validImportFormats = map[string]bool{
	"application/fhir+ndjson": true,
	"application/ndjson":      true,
	"ndjson":                  true,
}

// ---------------------------------------------------------------------------
// Validation
// ---------------------------------------------------------------------------

// ValidateImportRequest validates an ImportRequest and returns a slice of
// ValidationIssue values describing any problems found. An empty slice
// indicates the request is valid.
func ValidateImportRequest(req *ImportRequest) []ValidationIssue {
	var issues []ValidationIssue

	// inputFormat must be a supported value.
	if req.InputFormat == "" {
		issues = append(issues, ValidationIssue{
			Severity:    SeverityError,
			Code:        VIssueTypeRequired,
			Location:    "inputFormat",
			Diagnostics: "inputFormat is required",
		})
	} else if !validImportFormats[req.InputFormat] {
		issues = append(issues, ValidationIssue{
			Severity:    SeverityError,
			Code:        VIssueTypeValue,
			Location:    "inputFormat",
			Diagnostics: fmt.Sprintf("unsupported inputFormat: %s", req.InputFormat),
		})
	}

	// inputSource, when provided, must be a valid URL.
	if req.InputSource != "" {
		if _, err := url.ParseRequestURI(req.InputSource); err != nil {
			issues = append(issues, ValidationIssue{
				Severity:    SeverityError,
				Code:        VIssueTypeValue,
				Location:    "inputSource",
				Diagnostics: fmt.Sprintf("inputSource is not a valid URL: %s", err.Error()),
			})
		}
	}

	// input array must not be empty.
	if len(req.Input) == 0 {
		issues = append(issues, ValidationIssue{
			Severity:    SeverityError,
			Code:        VIssueTypeRequired,
			Location:    "input",
			Diagnostics: "input array must not be empty",
		})
	}

	// Validate each individual input entry.
	for i, inp := range req.Input {
		loc := fmt.Sprintf("input[%d]", i)

		if inp.Type == "" {
			issues = append(issues, ValidationIssue{
				Severity:    SeverityError,
				Code:        VIssueTypeRequired,
				Location:    loc + ".type",
				Diagnostics: fmt.Sprintf("%s.type is required", loc),
			})
		} else if !IsValidResourceType(inp.Type) {
			issues = append(issues, ValidationIssue{
				Severity:    SeverityError,
				Code:        VIssueTypeValue,
				Location:    loc + ".type",
				Diagnostics: fmt.Sprintf("%s.type '%s' is not a valid FHIR resource type", loc, inp.Type),
			})
		}

		if inp.URL == "" {
			issues = append(issues, ValidationIssue{
				Severity:    SeverityError,
				Code:        VIssueTypeRequired,
				Location:    loc + ".url",
				Diagnostics: fmt.Sprintf("%s.url is required", loc),
			})
		} else if _, err := url.ParseRequestURI(inp.URL); err != nil {
			issues = append(issues, ValidationIssue{
				Severity:    SeverityError,
				Code:        VIssueTypeValue,
				Location:    loc + ".url",
				Diagnostics: fmt.Sprintf("%s.url is not a valid URL: %s", loc, err.Error()),
			})
		}
	}

	return issues
}

// ---------------------------------------------------------------------------
// NDJSON parsing
// ---------------------------------------------------------------------------

// ParseNDJSON splits newline-delimited JSON data into individual JSON messages.
// Blank lines are silently skipped. An error is returned if any non-blank line
// is not valid JSON.
func ParseNDJSON(data []byte) ([]json.RawMessage, error) {
	if len(data) == 0 {
		return nil, nil
	}

	var results []json.RawMessage
	lines := bytes.Split(data, []byte("\n"))

	for i, line := range lines {
		trimmed := bytes.TrimSpace(line)
		if len(trimmed) == 0 {
			continue
		}

		if !json.Valid(trimmed) {
			return nil, fmt.Errorf("invalid JSON on line %d", i+1)
		}

		results = append(results, json.RawMessage(trimmed))
	}

	return results, nil
}

// ---------------------------------------------------------------------------
// HTTP handler
// ---------------------------------------------------------------------------

// ImportHandler returns an echo.HandlerFunc that handles POST /fhir/$import.
//
// The handler validates the import request, creates an async job via the
// provided AsyncJobStore, kicks off a background goroutine to process the
// import, and returns 202 Accepted with a Content-Location header pointing
// to the async status endpoint.
func ImportHandler(store AsyncJobStore) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Only accept POST.
		if c.Request().Method != http.MethodPost {
			return c.JSON(http.StatusMethodNotAllowed, NewOperationOutcome(
				IssueSeverityError, IssueTypeNotSupported,
				fmt.Sprintf("HTTP method %s is not allowed; use POST", c.Request().Method),
			))
		}

		// Parse the request body.
		var req ImportRequest
		if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
			return c.JSON(http.StatusBadRequest, NewOperationOutcome(
				IssueSeverityError, IssueTypeStructure,
				fmt.Sprintf("invalid request body: %s", err.Error()),
			))
		}

		// Validate the request.
		issues := ValidateImportRequest(&req)
		if len(issues) > 0 {
			return c.JSON(http.StatusBadRequest, MultiValidationOutcome(issues))
		}

		// Build the async job.
		jobID := uuid.New().String()
		now := time.Now().UTC()

		job := &AsyncJob{
			ID:            jobID,
			Status:        AsyncStatusInProgress,
			Request:       c.Request().RequestURI,
			TransactionTS: now,
		}

		if err := store.Create(c.Request().Context(), job); err != nil {
			return c.JSON(http.StatusInternalServerError, NewOperationOutcome(
				IssueSeverityError, IssueTypeException,
				fmt.Sprintf("failed to create import job: %s", err.Error()),
			))
		}

		// Process the import asynchronously. This is a mock implementation
		// that immediately marks the job as completed.
		go processImport(store, job.ID, &req)

		// Return 202 Accepted with Content-Location.
		return RespondAsync(c, store, jobID)
	}
}

// processImport is the background worker for an $import job.
// This is a mock implementation that simulates processing the import
// inputs and marks the job as completed.
func processImport(store AsyncJobStore, jobID string, req *ImportRequest) {
	// Build mock outcome: one entry per input with a count of 0 (no real data loaded).
	var outputs []AsyncJobOutput
	for _, inp := range req.Input {
		outputs = append(outputs, AsyncJobOutput{
			Type: inp.Type,
			URL:  inp.URL,
		})
	}

	// Simulate a small processing delay.
	time.Sleep(10 * time.Millisecond)

	job, err := store.Get(nil, jobID)
	if err != nil {
		return
	}

	job.Status = AsyncStatusCompleted
	job.Output = outputs
	_ = store.Update(nil, job)
}

// ---------------------------------------------------------------------------
// Helper: build Content-Type for import responses
// ---------------------------------------------------------------------------

// importContentType returns the FHIR-compliant content type for import
// responses.
func importContentType() string {
	return strings.Join([]string{"application", "fhir+json"}, "/")
}
