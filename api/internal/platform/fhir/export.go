package fhir

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// ResourceExporter defines the interface for exporting FHIR resources.
// Domain services implement this (via the ServiceExporter adapter) to
// provide real data to the bulk export engine.
type ResourceExporter interface {
	// ExportAll returns all resources of this type as FHIR JSON maps.
	ExportAll(ctx context.Context, since *time.Time) ([]map[string]interface{}, error)
	// ExportByPatient returns resources for a specific patient.
	ExportByPatient(ctx context.Context, patientID string, since *time.Time) ([]map[string]interface{}, error)
}

// ServiceExporter is a generic adapter that wraps domain service list
// functions to implement the ResourceExporter interface. Callers supply
// function values that delegate to the appropriate domain service methods.
type ServiceExporter struct {
	ResourceType    string
	ListFn          func(ctx context.Context, since *time.Time) ([]map[string]interface{}, error)
	ListByPatientFn func(ctx context.Context, patientID string, since *time.Time) ([]map[string]interface{}, error)
}

// ExportAll delegates to ListFn if set, otherwise returns an empty slice.
func (s *ServiceExporter) ExportAll(ctx context.Context, since *time.Time) ([]map[string]interface{}, error) {
	if s.ListFn == nil {
		return nil, nil
	}
	return s.ListFn(ctx, since)
}

// ExportByPatient delegates to ListByPatientFn if set, otherwise returns an empty slice.
func (s *ServiceExporter) ExportByPatient(ctx context.Context, patientID string, since *time.Time) ([]map[string]interface{}, error) {
	if s.ListByPatientFn == nil {
		return nil, nil
	}
	return s.ListByPatientFn(ctx, patientID, since)
}

// ExportJob represents a FHIR $export bulk data job.
type ExportJob struct {
	ID            string             `json:"id"`
	Status        string             `json:"status"` // in-progress, complete, error
	ResourceTypes []string           `json:"resource_types,omitempty"`
	PatientID     string             `json:"patient_id,omitempty"`
	Since         *time.Time         `json:"since,omitempty"`
	OutputFormat  string             `json:"output_format"`
	OutputFiles   []ExportOutputFile `json:"output,omitempty"`
	CreatedAt     time.Time          `json:"created_at"`
	CompletedAt   *time.Time         `json:"completed_at,omitempty"`
	ErrorMessage  string             `json:"error,omitempty"`

	// ndjsonData stores the exported NDJSON bytes keyed by resource type.
	// This field is not serialised to JSON; it is internal storage.
	ndjsonData map[string][]byte
}

// ExportOutputFile represents a single output file from a bulk export job.
type ExportOutputFile struct {
	Type  string `json:"type"`
	URL   string `json:"url"`
	Count int    `json:"count,omitempty"`
}

// ExportManager manages export jobs in-memory and dispatches export
// processing via registered ResourceExporter implementations.
type ExportManager struct {
	mu        sync.RWMutex
	jobs      map[string]*ExportJob
	exporters map[string]ResourceExporter
}

// NewExportManager creates a new ExportManager.
func NewExportManager() *ExportManager {
	return &ExportManager{
		jobs:      make(map[string]*ExportJob),
		exporters: make(map[string]ResourceExporter),
	}
}

// RegisterExporter registers a ResourceExporter for the given FHIR resource type.
func (m *ExportManager) RegisterExporter(resourceType string, exporter ResourceExporter) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.exporters[resourceType] = exporter
}

// KickOff creates a new system-level export job and starts async processing.
func (m *ExportManager) KickOff(resourceTypes []string, since *time.Time) *ExportJob {
	return m.kickOff(resourceTypes, "", since)
}

// KickOffForPatient creates a new patient-level export job and starts async processing.
func (m *ExportManager) KickOffForPatient(resourceTypes []string, patientID string, since *time.Time) *ExportJob {
	return m.kickOff(resourceTypes, patientID, since)
}

func (m *ExportManager) kickOff(resourceTypes []string, patientID string, since *time.Time) *ExportJob {
	m.mu.Lock()
	defer m.mu.Unlock()

	id := uuid.New().String()
	now := time.Now().UTC()

	// Default resource types if none specified
	if len(resourceTypes) == 0 {
		resourceTypes = []string{"Patient", "Observation", "Condition", "Encounter", "MedicationRequest"}
	}

	job := &ExportJob{
		ID:            id,
		Status:        "in-progress",
		ResourceTypes: resourceTypes,
		PatientID:     patientID,
		Since:         since,
		OutputFormat:  "application/fhir+ndjson",
		CreatedAt:     now,
		ndjsonData:    make(map[string][]byte),
	}

	m.jobs[id] = job

	// Start async export processing in a goroutine
	go m.processExport(job)

	return job
}

// processExport runs the export for each resource type in the job.
func (m *ExportManager) processExport(job *ExportJob) {
	ctx := context.Background()

	// Collect exporters under read lock
	m.mu.RLock()
	exportersCopy := make(map[string]ResourceExporter, len(m.exporters))
	for k, v := range m.exporters {
		exportersCopy[k] = v
	}
	m.mu.RUnlock()

	outputFiles := make([]ExportOutputFile, 0, len(job.ResourceTypes))
	ndjsonData := make(map[string][]byte, len(job.ResourceTypes))

	for _, rt := range job.ResourceTypes {
		exporter, ok := exportersCopy[rt]
		if !ok {
			// No exporter registered for this type; produce empty data
			ndjsonData[rt] = nil
			outputFiles = append(outputFiles, ExportOutputFile{
				Type:  rt,
				URL:   fmt.Sprintf("/fhir/$export-data/%s/%s", job.ID, rt),
				Count: 0,
			})
			continue
		}

		var resources []map[string]interface{}
		var err error

		if job.PatientID != "" {
			resources, err = exporter.ExportByPatient(ctx, job.PatientID, job.Since)
		} else {
			resources, err = exporter.ExportAll(ctx, job.Since)
		}

		if err != nil {
			// Mark job as error
			m.mu.Lock()
			job.Status = "error"
			job.ErrorMessage = fmt.Sprintf("export failed for %s: %s", rt, err.Error())
			m.mu.Unlock()
			return
		}

		// Convert resources to NDJSON
		var buf bytes.Buffer
		for _, r := range resources {
			line, err := json.Marshal(r)
			if err != nil {
				m.mu.Lock()
				job.Status = "error"
				job.ErrorMessage = fmt.Sprintf("json marshal failed for %s: %s", rt, err.Error())
				m.mu.Unlock()
				return
			}
			buf.Write(line)
			buf.WriteByte('\n')
		}

		ndjsonData[rt] = buf.Bytes()
		outputFiles = append(outputFiles, ExportOutputFile{
			Type:  rt,
			URL:   fmt.Sprintf("/fhir/$export-data/%s/%s", job.ID, rt),
			Count: len(resources),
		})
	}

	// Mark job as complete
	now := time.Now().UTC()
	m.mu.Lock()
	job.Status = "complete"
	job.CompletedAt = &now
	job.OutputFiles = outputFiles
	job.ndjsonData = ndjsonData
	m.mu.Unlock()
}

// GetStatus retrieves a snapshot of an export job by ID. The returned
// ExportJob is a copy so callers can inspect it without holding locks.
func (m *ExportManager) GetStatus(jobID string) (*ExportJob, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	job, ok := m.jobs[jobID]
	if !ok {
		return nil, fmt.Errorf("export job not found: %s", jobID)
	}

	// Return a shallow copy to avoid data races on fields mutated by
	// the background goroutine after the lock is released.
	snapshot := *job
	// Copy slices to avoid aliasing.
	if job.OutputFiles != nil {
		snapshot.OutputFiles = make([]ExportOutputFile, len(job.OutputFiles))
		copy(snapshot.OutputFiles, job.OutputFiles)
	}
	if job.ResourceTypes != nil {
		snapshot.ResourceTypes = make([]string, len(job.ResourceTypes))
		copy(snapshot.ResourceTypes, job.ResourceTypes)
	}
	return &snapshot, nil
}

// GetJobData returns NDJSON data for a specific resource type in an export job.
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

	data := job.ndjsonData[fileType]
	if data == nil {
		// Return empty bytes (not nil) so callers get a valid but empty response
		return []byte{}, nil
	}
	return data, nil
}

// DeleteJob removes an export job and its data.
func (m *ExportManager) DeleteJob(jobID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.jobs[jobID]; !ok {
		return fmt.Errorf("export job not found: %s", jobID)
	}

	delete(m.jobs, jobID)
	return nil
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
	fhirGroup.DELETE("/$export-status/:id", h.DeleteExport)
}

// SystemExport handles POST /fhir/$export (system-level export kick-off).
func (h *ExportHandler) SystemExport(c echo.Context) error {
	return h.kickOff(c, "")
}

// PatientExport handles POST /fhir/Patient/$export (patient-level export kick-off).
func (h *ExportHandler) PatientExport(c echo.Context) error {
	patientID := c.Param("patient_id")
	return h.kickOff(c, patientID)
}

func (h *ExportHandler) kickOff(c echo.Context, patientID string) error {
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

	var job *ExportJob
	if patientID != "" {
		job = h.manager.KickOffForPatient(resourceTypes, patientID, since)
	} else {
		job = h.manager.KickOff(resourceTypes, since)
	}

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
			"transactionTime":     job.CompletedAt.Format(time.RFC3339),
			"request":             fmt.Sprintf("/fhir/$export?_type=%s", strings.Join(job.ResourceTypes, ",")),
			"requiresAccessToken": false,
			"output":              job.OutputFiles,
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

// DeleteExport handles DELETE /fhir/$export-status/:id.
func (h *ExportHandler) DeleteExport(c echo.Context) error {
	jobID := c.Param("id")

	err := h.manager.DeleteJob(jobID)
	if err != nil {
		return c.JSON(http.StatusNotFound, ErrorOutcome(err.Error()))
	}

	return c.NoContent(http.StatusNoContent)
}
