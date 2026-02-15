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

// GroupMemberResolver resolves the patient IDs belonging to a FHIR Group.
type GroupMemberResolver func(ctx context.Context, groupID string) ([]string, error)

// validOutputFormats lists accepted _outputFormat values that map to NDJSON.
var validOutputFormats = map[string]bool{
	"application/fhir+ndjson": true,
	"application/ndjson":      true,
	"ndjson":                  true,
}

// ExportJob represents a FHIR $export bulk data job.
type ExportJob struct {
	ID            string             `json:"id"`
	Status        string             `json:"status"` // in-progress, complete, error
	ResourceTypes []string           `json:"resource_types,omitempty"`
	PatientID     string             `json:"patient_id,omitempty"`
	GroupID       string             `json:"group_id,omitempty"`
	Since         *time.Time         `json:"since,omitempty"`
	OutputFormat  string             `json:"output_format"`
	OutputFiles   []ExportOutputFile `json:"output,omitempty"`
	CreatedAt     time.Time          `json:"created_at"`
	CompletedAt   *time.Time         `json:"completed_at,omitempty"`
	ErrorMessage  string             `json:"error,omitempty"`
	TypeFilter    []string           `json:"type_filter,omitempty"`
	RequestTime   time.Time          `json:"request_time"`

	// Progress tracking
	ProcessedTypes int `json:"processed_types"`
	TotalTypes     int `json:"total_types"`

	// patientIDs is used for group exports; each member is exported separately.
	patientIDs []string

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

// ExportOptions configures the ExportManager.
type ExportOptions struct {
	MaxConcurrentJobs int
	JobTTL            time.Duration
}

// ExportManager manages export jobs in-memory and dispatches export
// processing via registered ResourceExporter implementations.
type ExportManager struct {
	mu            sync.RWMutex
	jobs          map[string]*ExportJob
	exporters     map[string]ResourceExporter
	groupResolver GroupMemberResolver

	maxConcurrentJobs int
	jobTTL            time.Duration
}

// NewExportManager creates a new ExportManager with default settings.
func NewExportManager() *ExportManager {
	return &ExportManager{
		jobs:              make(map[string]*ExportJob),
		exporters:         make(map[string]ResourceExporter),
		maxConcurrentJobs: 10,
		jobTTL:            time.Hour,
	}
}

// NewExportManagerWithOptions creates a new ExportManager with custom options.
func NewExportManagerWithOptions(opts ExportOptions) *ExportManager {
	m := NewExportManager()
	if opts.MaxConcurrentJobs > 0 {
		m.maxConcurrentJobs = opts.MaxConcurrentJobs
	}
	if opts.JobTTL > 0 {
		m.jobTTL = opts.JobTTL
	}
	return m
}

// SetGroupResolver sets the function used to resolve group membership.
func (m *ExportManager) SetGroupResolver(resolver GroupMemberResolver) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.groupResolver = resolver
}

// RegisterExporter registers a ResourceExporter for the given FHIR resource type.
func (m *ExportManager) RegisterExporter(resourceType string, exporter ResourceExporter) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.exporters[resourceType] = exporter
}

// KickOff creates a new system-level export job and starts async processing.
// It panics if the concurrent job limit is reached (callers that need error
// handling should use KickOffWithFormat directly).
func (m *ExportManager) KickOff(resourceTypes []string, since *time.Time) *ExportJob {
	job, err := m.KickOffWithFormat(resourceTypes, "", since, "", nil)
	if err != nil {
		panic("KickOff failed: " + err.Error())
	}
	return job
}

// KickOffForPatient creates a new patient-level export job and starts async processing.
// It panics if the concurrent job limit is reached (callers that need error
// handling should use KickOffWithFormat directly).
func (m *ExportManager) KickOffForPatient(resourceTypes []string, patientID string, since *time.Time) *ExportJob {
	job, err := m.KickOffWithFormat(resourceTypes, patientID, since, "", nil)
	if err != nil {
		panic("KickOffForPatient failed: " + err.Error())
	}
	return job
}

// KickOffWithFormat creates a new export job with output format validation
// and optional type filters. Returns an error if the format is unsupported
// or the concurrent job limit is reached.
func (m *ExportManager) KickOffWithFormat(resourceTypes []string, patientID string, since *time.Time, outputFormat string, typeFilter []string) (*ExportJob, error) {
	// Validate output format
	if outputFormat != "" {
		if !validOutputFormats[outputFormat] {
			return nil, fmt.Errorf("unsupported _outputFormat: %s", outputFormat)
		}
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Check concurrent job limit
	activeCount := 0
	for _, j := range m.jobs {
		if j.Status == "in-progress" {
			activeCount++
		}
	}
	if activeCount >= m.maxConcurrentJobs {
		return nil, fmt.Errorf("max concurrent export jobs reached (%d)", m.maxConcurrentJobs)
	}

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
		RequestTime:   now,
		TypeFilter:    typeFilter,
		TotalTypes:    len(resourceTypes),
		ndjsonData:    make(map[string][]byte),
	}

	m.jobs[id] = job

	// Start async export processing in a goroutine
	go m.processExport(job)

	return job, nil
}

// KickOffForGroup creates a group-level export job. It resolves group
// members via the registered GroupMemberResolver and exports per-patient.
func (m *ExportManager) KickOffForGroup(resourceTypes []string, groupID string, since *time.Time, outputFormat string, typeFilter []string) (*ExportJob, error) {
	// Validate output format
	if outputFormat != "" {
		if !validOutputFormats[outputFormat] {
			return nil, fmt.Errorf("unsupported _outputFormat: %s", outputFormat)
		}
	}

	m.mu.RLock()
	resolver := m.groupResolver
	m.mu.RUnlock()

	if resolver == nil {
		return nil, fmt.Errorf("group export not configured")
	}

	// Resolve group members
	members, err := resolver(context.Background(), groupID)
	if err != nil {
		return nil, err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Check concurrent job limit
	activeCount := 0
	for _, j := range m.jobs {
		if j.Status == "in-progress" {
			activeCount++
		}
	}
	if activeCount >= m.maxConcurrentJobs {
		return nil, fmt.Errorf("max concurrent export jobs reached (%d)", m.maxConcurrentJobs)
	}

	id := uuid.New().String()
	now := time.Now().UTC()

	if len(resourceTypes) == 0 {
		resourceTypes = []string{"Patient", "Observation", "Condition", "Encounter", "MedicationRequest"}
	}

	job := &ExportJob{
		ID:            id,
		Status:        "in-progress",
		ResourceTypes: resourceTypes,
		GroupID:       groupID,
		Since:         since,
		OutputFormat:  "application/fhir+ndjson",
		CreatedAt:     now,
		RequestTime:   now,
		TypeFilter:    typeFilter,
		TotalTypes:    len(resourceTypes),
		patientIDs:    members,
		ndjsonData:    make(map[string][]byte),
	}

	m.jobs[id] = job

	go m.processGroupExport(job)

	return job, nil
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
			m.mu.Lock()
			job.ProcessedTypes++
			m.mu.Unlock()
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

		m.mu.Lock()
		job.ProcessedTypes++
		m.mu.Unlock()
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

// processGroupExport exports data for each member of a group.
func (m *ExportManager) processGroupExport(job *ExportJob) {
	ctx := context.Background()

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
			ndjsonData[rt] = nil
			outputFiles = append(outputFiles, ExportOutputFile{
				Type:  rt,
				URL:   fmt.Sprintf("/fhir/$export-data/%s/%s", job.ID, rt),
				Count: 0,
			})
			m.mu.Lock()
			job.ProcessedTypes++
			m.mu.Unlock()
			continue
		}

		var buf bytes.Buffer
		totalCount := 0

		for _, pid := range job.patientIDs {
			resources, err := exporter.ExportByPatient(ctx, pid, job.Since)
			if err != nil {
				m.mu.Lock()
				job.Status = "error"
				job.ErrorMessage = fmt.Sprintf("export failed for %s (patient %s): %s", rt, pid, err.Error())
				m.mu.Unlock()
				return
			}
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
				totalCount++
			}
		}

		ndjsonData[rt] = buf.Bytes()
		outputFiles = append(outputFiles, ExportOutputFile{
			Type:  rt,
			URL:   fmt.Sprintf("/fhir/$export-data/%s/%s", job.ID, rt),
			Count: totalCount,
		})

		m.mu.Lock()
		job.ProcessedTypes++
		m.mu.Unlock()
	}

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
	if job.TypeFilter != nil {
		snapshot.TypeFilter = make([]string, len(job.TypeFilter))
		copy(snapshot.TypeFilter, job.TypeFilter)
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

// CleanupExpiredJobs removes completed/error jobs older than TTL and marks
// in-progress jobs older than 2x TTL as timed-out errors.
func (m *ExportManager) CleanupExpiredJobs() {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now().UTC()
	for id, job := range m.jobs {
		age := now.Sub(job.CreatedAt)
		switch job.Status {
		case "complete", "error":
			if age > m.jobTTL {
				delete(m.jobs, id)
			}
		case "in-progress":
			if age > 2*m.jobTTL {
				job.Status = "error"
				job.ErrorMessage = "export job timed out"
			}
		}
	}
}

// StartCleanup runs a background goroutine that periodically removes expired jobs.
func (m *ExportManager) StartCleanup(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				m.CleanupExpiredJobs()
			}
		}
	}()
}

// ActiveJobCount returns the number of in-progress jobs (for 429 responses).
func (m *ExportManager) ActiveJobCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	count := 0
	for _, j := range m.jobs {
		if j.Status == "in-progress" {
			count++
		}
	}
	return count
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
	fhirGroup.POST("/Patient/:id/$export", h.PatientExportByID)
	fhirGroup.POST("/Group/:id/$export", h.GroupExport)
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

// PatientExportByID handles POST /fhir/Patient/:id/$export (patient-level export by FHIR ID).
func (h *ExportHandler) PatientExportByID(c echo.Context) error {
	patientID := c.Param("id")
	return h.kickOff(c, patientID)
}

// GroupExport handles POST /fhir/Group/:id/$export (group-level export).
func (h *ExportHandler) GroupExport(c echo.Context) error {
	groupID := c.Param("id")

	// Check Prefer header for respond-async
	prefer := c.Request().Header.Get("Prefer")
	if prefer != "" && !strings.Contains(prefer, "respond-async") {
		return c.JSON(http.StatusBadRequest, ErrorOutcome("Prefer header must include respond-async for bulk export"))
	}

	// Parse _outputFormat
	outputFormat := c.QueryParam("_outputFormat")

	// Parse _type query param
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

	// Parse _typeFilter
	var typeFilter []string
	typeFilterParam := c.QueryParam("_typeFilter")
	if typeFilterParam != "" {
		for _, tf := range strings.Split(typeFilterParam, ",") {
			tf = strings.TrimSpace(tf)
			if tf != "" {
				typeFilter = append(typeFilter, tf)
			}
		}
	}

	job, err := h.manager.KickOffForGroup(resourceTypes, groupID, since, outputFormat, typeFilter)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return c.JSON(http.StatusNotFound, ErrorOutcome(err.Error()))
		}
		if strings.Contains(err.Error(), "concurrent") {
			c.Response().Header().Set("Retry-After", "120")
			return c.JSON(http.StatusTooManyRequests, ErrorOutcome(err.Error()))
		}
		if strings.Contains(err.Error(), "unsupported") {
			return c.JSON(http.StatusBadRequest, ErrorOutcome(err.Error()))
		}
		return c.JSON(http.StatusInternalServerError, ErrorOutcome(err.Error()))
	}

	c.Response().Header().Set("Content-Location", fmt.Sprintf("/fhir/$export-status/%s", job.ID))
	return c.NoContent(http.StatusAccepted)
}

func (h *ExportHandler) kickOff(c echo.Context, patientID string) error {
	// Check Prefer header for respond-async
	prefer := c.Request().Header.Get("Prefer")
	if prefer != "" && !strings.Contains(prefer, "respond-async") {
		return c.JSON(http.StatusBadRequest, ErrorOutcome("Prefer header must include respond-async for bulk export"))
	}

	// Parse _outputFormat
	outputFormat := c.QueryParam("_outputFormat")

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

	// Parse _typeFilter
	var typeFilter []string
	typeFilterParam := c.QueryParam("_typeFilter")
	if typeFilterParam != "" {
		for _, tf := range strings.Split(typeFilterParam, ",") {
			tf = strings.TrimSpace(tf)
			if tf != "" {
				typeFilter = append(typeFilter, tf)
			}
		}
	}

	job, err := h.manager.KickOffWithFormat(resourceTypes, patientID, since, outputFormat, typeFilter)
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
		c.Response().Header().Set("Retry-After", "120")
		if job.TotalTypes > 0 {
			c.Response().Header().Set("X-Progress", fmt.Sprintf("%d/%d resource types processed", job.ProcessedTypes, job.TotalTypes))
		} else {
			c.Response().Header().Set("X-Progress", job.Status)
		}
		return c.NoContent(http.StatusAccepted)
	case "complete":
		result := map[string]interface{}{
			"transactionTime":     job.RequestTime.Format(time.RFC3339),
			"request":             fmt.Sprintf("/fhir/$export?_type=%s", strings.Join(job.ResourceTypes, ",")),
			"requiresAccessToken": true,
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
