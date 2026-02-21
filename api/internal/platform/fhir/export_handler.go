package fhir

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
)

// ExportStore manages export jobs in memory. It wraps ExportManager and
// provides a thin persistence layer. The store is safe for concurrent use.
type ExportStore struct {
	mu   sync.RWMutex
	jobs map[string]*ExportJob
}

// NewExportStore creates a new, empty ExportStore.
func NewExportStore() *ExportStore {
	return &ExportStore{
		jobs: make(map[string]*ExportJob),
	}
}

// Create stores a new export job.
func (s *ExportStore) Create(job *ExportJob) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.jobs[job.ID] = job
}

// Get retrieves an export job by ID. Returns the job and true if found, nil
// and false otherwise.
func (s *ExportStore) Get(id string) (*ExportJob, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	job, ok := s.jobs[id]
	if !ok {
		return nil, false
	}
	// Return a copy to avoid data races.
	snapshot := *job
	if job.OutputFiles != nil {
		snapshot.OutputFiles = make([]ExportOutputFile, len(job.OutputFiles))
		copy(snapshot.OutputFiles, job.OutputFiles)
	}
	if job.ResourceTypes != nil {
		snapshot.ResourceTypes = make([]string, len(job.ResourceTypes))
		copy(snapshot.ResourceTypes, job.ResourceTypes)
	}
	return &snapshot, true
}

// Update replaces an existing export job in the store.
func (s *ExportStore) Update(job *ExportJob) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.jobs[job.ID] = job
}

// Delete removes an export job from the store.
func (s *ExportStore) Delete(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.jobs, id)
}

// BulkExportHandler provides the FHIR Bulk Data Access (Flat FHIR)
// specification-compliant HTTP endpoints. It delegates actual export
// processing to the ExportManager and uses ExportStore for job tracking.
//
// Endpoints:
//
//	GET  /fhir/Patient/:id/$export         Patient-level export kick-off
//	GET  /fhir/$export                     System-level export kick-off
//	GET  /fhir/$export-poll-status         Export status polling
//	DELETE /fhir/$export-poll-status       Cancel/delete an export job
//	GET  /fhir/$export-output/:jobId/:fileName  Download NDJSON output file
type BulkExportHandler struct {
	manager *ExportManager
	store   *ExportStore
	baseURL string
}

// NewBulkExportHandler creates a new BulkExportHandler.
// baseURL is the public base URL of the server (e.g. "https://example.com")
// used to build absolute URLs in status responses. Pass an empty string
// to use relative URLs.
func NewBulkExportHandler(store *ExportStore, manager *ExportManager, baseURL string) *BulkExportHandler {
	return &BulkExportHandler{
		manager: manager,
		store:   store,
		baseURL: strings.TrimRight(baseURL, "/"),
	}
}

// RegisterBulkExportRoutes registers the Bulk Data Access routes on the
// given Echo group. These are GET-based endpoints per the FHIR Bulk Data
// specification and coexist with the existing POST-based export routes.
func (h *BulkExportHandler) RegisterBulkExportRoutes(fhirGroup *echo.Group) {
	// Kick-off (GET per spec)
	fhirGroup.GET("/$export", h.SystemExportKickOff)
	fhirGroup.GET("/Patient/:id/$export", h.PatientExportKickOff)

	// Status polling
	fhirGroup.GET("/$export-poll-status", h.ExportPollStatus)
	fhirGroup.DELETE("/$export-poll-status", h.DeleteExportPollStatus)

	// Output file download
	fhirGroup.GET("/$export-output/:jobId/:fileName", h.ExportOutput)
}

// SystemExportKickOff handles GET /fhir/$export (system-level export).
// This is intended for admin users to export all data.
func (h *BulkExportHandler) SystemExportKickOff(c echo.Context) error {
	return h.kickOffExport(c, "")
}

// PatientExportKickOff handles GET /fhir/Patient/:id/$export.
// Exports all data for a single patient.
func (h *BulkExportHandler) PatientExportKickOff(c echo.Context) error {
	patientID := c.Param("id")
	if patientID == "" {
		return c.JSON(http.StatusBadRequest, ErrorOutcome("patient ID is required"))
	}
	return h.kickOffExport(c, patientID)
}

func (h *BulkExportHandler) kickOffExport(c echo.Context, patientID string) error {
	// Validate _outputFormat
	outputFormat := c.QueryParam("_outputFormat")
	if outputFormat != "" && !validOutputFormats[outputFormat] {
		return c.JSON(http.StatusBadRequest, ErrorOutcome(
			fmt.Sprintf("unsupported _outputFormat: %s; only application/fhir+ndjson is supported", outputFormat)))
	}

	// Parse _type
	var resourceTypes []string
	typeParam := c.QueryParam("_type")
	if typeParam != "" {
		for _, t := range strings.Split(typeParam, ",") {
			t = strings.TrimSpace(t)
			if t != "" {
				resourceTypes = append(resourceTypes, t)
			}
		}
	}

	// Parse _since
	var since *time.Time
	sinceParam := c.QueryParam("_since")
	if sinceParam != "" {
		t, err := time.Parse(time.RFC3339, sinceParam)
		if err != nil {
			return c.JSON(http.StatusBadRequest, ErrorOutcome("invalid _since format, expected RFC3339"))
		}
		since = &t
	}

	// Delegate to ExportManager for job creation and async processing
	var job *ExportJob
	var err error
	if patientID != "" {
		job, err = h.manager.KickOffForPatient(resourceTypes, patientID, since)
	} else {
		job, err = h.manager.KickOff(resourceTypes, since)
	}
	if err != nil {
		if strings.Contains(err.Error(), "concurrent") {
			c.Response().Header().Set("Retry-After", "120")
			return c.JSON(http.StatusTooManyRequests, ErrorOutcome(err.Error()))
		}
		if strings.Contains(err.Error(), "unsupported") {
			return c.JSON(http.StatusBadRequest, ErrorOutcome(err.Error()))
		}
		return c.JSON(http.StatusInternalServerError, ErrorOutcome(err.Error()))
	}

	// Store the job for later status queries
	h.store.Create(job)

	// Return 202 Accepted with Content-Location pointing to the poll status endpoint
	statusURL := fmt.Sprintf("%s/fhir/$export-poll-status?job=%s", h.baseURL, job.ID)
	c.Response().Header().Set("Content-Location", statusURL)
	return c.NoContent(http.StatusAccepted)
}

// ExportPollStatus handles GET /fhir/$export-poll-status?job=<id>.
//
// Response when in progress: 202 Accepted with X-Progress and Retry-After: 10.
// Response when complete: 200 OK with the FHIR Bulk Data status response body.
// Response when failed: 500 with FHIR OperationOutcome.
func (h *BulkExportHandler) ExportPollStatus(c echo.Context) error {
	jobID := c.QueryParam("job")
	if jobID == "" {
		return c.JSON(http.StatusBadRequest, ErrorOutcome("job query parameter is required"))
	}

	// Fetch the latest status from the manager (which owns the async state)
	job, err := h.manager.GetStatus(jobID)
	if err != nil {
		return c.JSON(http.StatusNotFound, ErrorOutcome(err.Error()))
	}

	switch job.Status {
	case "in-progress", "accepted":
		c.Response().Header().Set("X-Progress", formatProgress(job))
		c.Response().Header().Set("Retry-After", "10")
		return c.NoContent(http.StatusAccepted)

	case "complete":
		// Build the output array with absolute URLs
		output := make([]map[string]interface{}, 0, len(job.OutputFiles))
		for _, f := range job.OutputFiles {
			output = append(output, map[string]interface{}{
				"type":  f.Type,
				"url":   h.buildOutputURL(job.ID, f.Type),
				"count": f.Count,
			})
		}

		result := map[string]interface{}{
			"transactionTime":     job.RequestTime.Format(time.RFC3339),
			"request":             h.buildRequestURL(job),
			"requiresAccessToken": true,
			"output":              output,
			"error":               []interface{}{},
		}
		return c.JSON(http.StatusOK, result)

	case "error":
		return c.JSON(http.StatusInternalServerError, ErrorOutcome(job.ErrorMessage))

	default:
		return c.JSON(http.StatusInternalServerError, ErrorOutcome("unknown job status"))
	}
}

// DeleteExportPollStatus handles DELETE /fhir/$export-poll-status?job=<id>.
// Cancels and deletes an export job.
func (h *BulkExportHandler) DeleteExportPollStatus(c echo.Context) error {
	jobID := c.QueryParam("job")
	if jobID == "" {
		return c.JSON(http.StatusBadRequest, ErrorOutcome("job query parameter is required"))
	}

	err := h.manager.DeleteJob(jobID)
	if err != nil {
		return c.JSON(http.StatusNotFound, ErrorOutcome(err.Error()))
	}

	// Also remove from the store
	h.store.Delete(jobID)

	return c.NoContent(http.StatusAccepted)
}

// ExportOutput handles GET /fhir/$export-output/:jobId/:fileName.
// Downloads the NDJSON file for the specified resource type.
// The fileName is expected to be in the form "<ResourceType>.ndjson".
func (h *BulkExportHandler) ExportOutput(c echo.Context) error {
	jobID := c.Param("jobId")
	fileName := c.Param("fileName")

	// Strip .ndjson suffix to get the resource type
	resourceType := strings.TrimSuffix(fileName, ".ndjson")
	if resourceType == fileName {
		// No .ndjson suffix; try using fileName as-is (backward compat)
		resourceType = fileName
	}

	data, err := h.manager.GetJobData(jobID, resourceType)
	if err != nil {
		return c.JSON(http.StatusNotFound, ErrorOutcome(err.Error()))
	}

	c.Response().Header().Set("Content-Type", "application/fhir+ndjson")
	return c.Blob(http.StatusOK, "application/fhir+ndjson", data)
}

// buildOutputURL constructs the download URL for an export output file.
func (h *BulkExportHandler) buildOutputURL(jobID, resourceType string) string {
	return fmt.Sprintf("%s/fhir/$export-output/%s/%s.ndjson", h.baseURL, jobID, resourceType)
}

// buildRequestURL reconstructs the original kick-off request URL.
func (h *BulkExportHandler) buildRequestURL(job *ExportJob) string {
	if job.PatientID != "" {
		return fmt.Sprintf("%s/fhir/Patient/%s/$export", h.baseURL, job.PatientID)
	}
	return fmt.Sprintf("%s/fhir/$export", h.baseURL)
}

// formatProgress returns a human-readable progress string for the X-Progress header.
func formatProgress(job *ExportJob) string {
	if job.TotalTypes > 0 {
		return fmt.Sprintf("%d/%d resource types processed", job.ProcessedTypes, job.TotalTypes)
	}
	return job.Status
}
