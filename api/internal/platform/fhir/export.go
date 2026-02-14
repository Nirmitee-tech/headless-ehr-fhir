package fhir

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// ExportJob represents a FHIR $export bulk data job.
type ExportJob struct {
	ID            string             `json:"id"`
	Status        string             `json:"status"` // accepted, in-progress, complete, error
	ResourceTypes []string           `json:"resource_types,omitempty"`
	Since         *time.Time         `json:"since,omitempty"`
	OutputFormat  string             `json:"output_format"`
	OutputFiles   []ExportOutputFile `json:"output,omitempty"`
	CreatedAt     time.Time          `json:"created_at"`
	CompletedAt   *time.Time         `json:"completed_at,omitempty"`
	ErrorMessage  string             `json:"error,omitempty"`
}

// ExportOutputFile represents a single output file from a bulk export job.
type ExportOutputFile struct {
	Type  string `json:"type"`
	URL   string `json:"url"`
	Count int    `json:"count,omitempty"`
}

// ExportManager manages export jobs in-memory.
type ExportManager struct {
	mu   sync.RWMutex
	jobs map[string]*ExportJob
}

// NewExportManager creates a new ExportManager.
func NewExportManager() *ExportManager {
	return &ExportManager{
		jobs: make(map[string]*ExportJob),
	}
}

// KickOff creates a new export job with the given resource types and since filter.
// In this simplified implementation, jobs are immediately set to "complete".
func (m *ExportManager) KickOff(resourceTypes []string, since *time.Time) *ExportJob {
	m.mu.Lock()
	defer m.mu.Unlock()

	id := uuid.New().String()
	now := time.Now().UTC()

	// Default resource types if none specified
	if len(resourceTypes) == 0 {
		resourceTypes = []string{"Patient", "Observation", "Condition", "Encounter", "MedicationRequest"}
	}

	// Build output files for each resource type
	outputFiles := make([]ExportOutputFile, len(resourceTypes))
	for i, rt := range resourceTypes {
		outputFiles[i] = ExportOutputFile{
			Type:  rt,
			URL:   fmt.Sprintf("/fhir/$export-data/%s/%s", id, rt),
			Count: 0,
		}
	}

	// Simplified: immediately complete the job
	job := &ExportJob{
		ID:            id,
		Status:        "complete",
		ResourceTypes: resourceTypes,
		Since:         since,
		OutputFormat:  "application/fhir+ndjson",
		OutputFiles:   outputFiles,
		CreatedAt:     now,
		CompletedAt:   &now,
	}

	m.jobs[id] = job
	return job
}

// GetStatus retrieves the status of an export job by ID.
func (m *ExportManager) GetStatus(jobID string) (*ExportJob, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	job, ok := m.jobs[jobID]
	if !ok {
		return nil, fmt.Errorf("export job not found: %s", jobID)
	}
	return job, nil
}

// GetJobData returns placeholder NDJSON data for a specific file type in an export job.
func (m *ExportManager) GetJobData(jobID, fileType string) ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	job, ok := m.jobs[jobID]
	if !ok {
		return nil, fmt.Errorf("export job not found: %s", jobID)
	}

	if job.Status != "complete" {
		return nil, fmt.Errorf("export job is not complete: %s", job.Status)
	}

	// Verify the file type is part of this job
	found := false
	for _, f := range job.OutputFiles {
		if f.Type == fileType {
			found = true
			break
		}
	}
	if !found {
		return nil, fmt.Errorf("file type %s not found in export job %s", fileType, jobID)
	}

	// Return placeholder NDJSON data
	placeholder := map[string]interface{}{
		"resourceType": fileType,
		"id":           "placeholder-1",
	}
	data, _ := json.Marshal(placeholder)
	return append(data, '\n'), nil
}

// ExportHandler provides REST endpoints for FHIR $export operations.
type ExportHandler struct {
	manager *ExportManager
}

// NewExportHandler creates a new ExportHandler.
func NewExportHandler(manager *ExportManager) *ExportHandler {
	return &ExportHandler{manager: manager}
}

// RegisterRoutes registers the export routes on the FHIR group.
func (h *ExportHandler) RegisterRoutes(fhirGroup *echo.Group) {
	fhirGroup.POST("/$export", h.SystemExport)
	fhirGroup.POST("/Patient/$export", h.PatientExport)
	fhirGroup.GET("/$export-status/:id", h.ExportStatus)
	fhirGroup.GET("/$export-data/:id/:type", h.ExportData)
}

// SystemExport handles POST /fhir/$export (system-level export kick-off).
func (h *ExportHandler) SystemExport(c echo.Context) error {
	return h.kickOff(c)
}

// PatientExport handles POST /fhir/Patient/$export (patient-level export kick-off).
func (h *ExportHandler) PatientExport(c echo.Context) error {
	return h.kickOff(c)
}

func (h *ExportHandler) kickOff(c echo.Context) error {
	// Check Prefer header for respond-async
	prefer := c.Request().Header.Get("Prefer")
	if prefer != "" && !strings.Contains(prefer, "respond-async") {
		return c.JSON(http.StatusBadRequest, ErrorOutcome("Prefer header must include respond-async for bulk export"))
	}

	// Parse _type query param (comma-separated resource types)
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

	// Parse _since query param
	var since *time.Time
	sinceParam := c.QueryParam("_since")
	if sinceParam != "" {
		t, err := time.Parse(time.RFC3339, sinceParam)
		if err != nil {
			return c.JSON(http.StatusBadRequest, ErrorOutcome("invalid _since format, expected RFC3339"))
		}
		since = &t
	}

	job := h.manager.KickOff(resourceTypes, since)

	c.Response().Header().Set("Content-Location", fmt.Sprintf("/fhir/$export-status/%s", job.ID))
	return c.NoContent(http.StatusAccepted)
}

// ExportStatus handles GET /fhir/$export-status/:id.
func (h *ExportHandler) ExportStatus(c echo.Context) error {
	jobID := c.Param("id")

	job, err := h.manager.GetStatus(jobID)
	if err != nil {
		return c.JSON(http.StatusNotFound, ErrorOutcome(err.Error()))
	}

	switch job.Status {
	case "accepted", "in-progress":
		c.Response().Header().Set("X-Progress", job.Status)
		return c.NoContent(http.StatusAccepted)
	case "complete":
		result := map[string]interface{}{
			"transactionTime": job.CompletedAt.Format(time.RFC3339),
			"request":         fmt.Sprintf("/fhir/$export?_type=%s", strings.Join(job.ResourceTypes, ",")),
			"requiresAccessToken": false,
			"output":          job.OutputFiles,
		}
		return c.JSON(http.StatusOK, result)
	case "error":
		return c.JSON(http.StatusInternalServerError, ErrorOutcome(job.ErrorMessage))
	default:
		return c.JSON(http.StatusInternalServerError, ErrorOutcome("unknown job status"))
	}
}

// ExportData handles GET /fhir/$export-data/:id/:type.
func (h *ExportHandler) ExportData(c echo.Context) error {
	jobID := c.Param("id")
	fileType := c.Param("type")

	data, err := h.manager.GetJobData(jobID, fileType)
	if err != nil {
		return c.JSON(http.StatusNotFound, ErrorOutcome(err.Error()))
	}

	c.Response().Header().Set("Content-Type", "application/fhir+ndjson")
	return c.Blob(http.StatusOK, "application/fhir+ndjson", data)
}
