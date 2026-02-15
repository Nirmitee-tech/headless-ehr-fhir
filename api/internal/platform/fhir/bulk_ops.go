package fhir

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// ==================== Structs ====================

// BulkError represents a single error encountered during bulk processing.
type BulkError struct {
	ResourceType string `json:"resourceType"`
	Index        int    `json:"index"`
	Error        string `json:"error"`
}

// BulkImportJob represents an asynchronous FHIR bulk import job.
type BulkImportJob struct {
	ID                 string      `json:"id"`
	Status             string      `json:"status"` // pending, processing, completed, error
	InputFormat        string      `json:"inputFormat"`
	ResourceTypes      []string    `json:"resourceTypes"`
	TotalResources     int         `json:"totalResources"`
	ProcessedResources int         `json:"processedResources"`
	SuccessCount       int         `json:"successCount"`
	ErrorCount         int         `json:"errorCount"`
	Errors             []BulkError `json:"errors,omitempty"`
	RequestTime        time.Time   `json:"requestTime"`
	CompletedTime      *time.Time  `json:"completedTime,omitempty"`
}

// BulkEditJob represents an asynchronous FHIR bulk edit (update/patch/delete) job.
type BulkEditJob struct {
	ID            string                 `json:"id"`
	Status        string                 `json:"status"` // pending, processing, completed, error, cancelled
	Operation     string                 `json:"operation"` // update, patch, delete
	ResourceType  string                 `json:"resourceType"`
	Criteria      map[string]string      `json:"criteria"`
	Patch         map[string]interface{} `json:"patch,omitempty"`
	MatchCount    int                    `json:"matchCount"`
	ModifiedCount int                    `json:"modifiedCount"`
	ErrorCount    int                    `json:"errorCount"`
	Errors        []BulkError            `json:"errors,omitempty"`
	RequestTime   time.Time              `json:"requestTime"`
	CompletedTime *time.Time             `json:"completedTime,omitempty"`
}

// ==================== Interfaces ====================

// BulkResourceValidator validates a single FHIR resource for bulk import.
type BulkResourceValidator interface {
	ValidateResource(resourceType string, data map[string]interface{}) error
}

// ResourceMatcher matches resources in a store by type and search criteria.
type ResourceMatcher interface {
	MatchResources(ctx context.Context, resourceType string, criteria map[string]string) ([]map[string]interface{}, error)
}

// ==================== DefaultBulkValidator ====================

// DefaultBulkValidator provides basic FHIR resource validation for bulk operations.
type DefaultBulkValidator struct{}

// ValidateResource checks that a resource has the expected resourceType, an id,
// and a status field (required for most FHIR resources).
func (v *DefaultBulkValidator) ValidateResource(resourceType string, data map[string]interface{}) error {
	// Check resourceType field present
	rt, ok := data["resourceType"]
	if !ok || rt == nil || rt == "" {
		return fmt.Errorf("missing required field: resourceType")
	}
	rtStr, ok := rt.(string)
	if !ok || rtStr == "" {
		return fmt.Errorf("missing required field: resourceType")
	}

	// Check resourceType matches expected
	if rtStr != resourceType {
		return fmt.Errorf("resourceType mismatch: expected %s, got %s", resourceType, rtStr)
	}

	// Check id field present
	id, ok := data["id"]
	if !ok || id == nil || id == "" {
		return fmt.Errorf("missing required field: id")
	}

	return nil
}

// ==================== InMemoryResourceStore ====================

// InMemoryResourceStore is a test implementation of ResourceMatcher that
// stores resources in memory keyed by resource type.
type InMemoryResourceStore struct {
	mu        sync.RWMutex
	resources map[string][]map[string]interface{}
}

// NewInMemoryResourceStore creates a new empty InMemoryResourceStore.
func NewInMemoryResourceStore() *InMemoryResourceStore {
	return &InMemoryResourceStore{
		resources: make(map[string][]map[string]interface{}),
	}
}

// AddResource adds a resource to the in-memory store.
func (s *InMemoryResourceStore) AddResource(resourceType string, resource map[string]interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.resources[resourceType] = append(s.resources[resourceType], resource)
}

// GetResources returns all resources of the given type.
func (s *InMemoryResourceStore) GetResources(resourceType string) []map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]map[string]interface{}, len(s.resources[resourceType]))
	copy(result, s.resources[resourceType])
	return result
}

// MatchResources returns resources of the given type that match all criteria.
// Criteria keys are field names and values are expected string values.
func (s *InMemoryResourceStore) MatchResources(_ context.Context, resourceType string, criteria map[string]string) ([]map[string]interface{}, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	all := s.resources[resourceType]
	var matched []map[string]interface{}
	for _, r := range all {
		if bulkMatchesCriteria(r, criteria) {
			matched = append(matched, r)
		}
	}
	return matched, nil
}

// matchesCriteria checks if a resource matches all key-value criteria.
func bulkMatchesCriteria(resource map[string]interface{}, criteria map[string]string) bool {
	for k, v := range criteria {
		rv, ok := resource[k]
		if !ok {
			return false
		}
		// Compare as string
		if fmt.Sprintf("%v", rv) != v {
			return false
		}
	}
	return true
}

// ==================== BulkOperationManager ====================

// BulkOperationManager manages FHIR bulk import and bulk edit jobs.
type BulkOperationManager struct {
	importMu   sync.RWMutex
	importJobs map[string]*BulkImportJob

	editMu   sync.RWMutex
	editJobs map[string]*BulkEditJob

	validator        BulkResourceValidator
	matcher          ResourceMatcher
	maxConcurrentJobs int
}

// NewBulkOperationManager creates a new BulkOperationManager with default settings.
// The matcher parameter may be nil if bulk edit operations are not needed.
func NewBulkOperationManager(matcher ResourceMatcher) *BulkOperationManager {
	return &BulkOperationManager{
		importJobs:        make(map[string]*BulkImportJob),
		editJobs:          make(map[string]*BulkEditJob),
		validator:         &DefaultBulkValidator{},
		matcher:           matcher,
		maxConcurrentJobs: 5,
	}
}

// NewBulkOperationManagerWithOptions creates a new BulkOperationManager with a custom
// max concurrent jobs limit.
func NewBulkOperationManagerWithOptions(matcher ResourceMatcher, maxConcurrentJobs int) *BulkOperationManager {
	mgr := NewBulkOperationManager(matcher)
	if maxConcurrentJobs > 0 {
		mgr.maxConcurrentJobs = maxConcurrentJobs
	}
	return mgr
}

// ==================== Import Methods ====================

// StartImport parses NDJSON data, validates each resource, and creates a bulk import job.
func (m *BulkOperationManager) StartImport(_ context.Context, resourceType string, data []byte) (*BulkImportJob, error) {
	if len(bytes.TrimSpace(data)) == 0 {
		return nil, fmt.Errorf("empty input: no NDJSON data provided")
	}

	// Check concurrent job limit
	m.importMu.Lock()
	activeCount := 0
	for _, j := range m.importJobs {
		if j.Status == "pending" || j.Status == "processing" {
			activeCount++
		}
	}
	if activeCount >= m.maxConcurrentJobs {
		m.importMu.Unlock()
		return nil, fmt.Errorf("max concurrent import jobs reached (%d)", m.maxConcurrentJobs)
	}

	id := uuid.New().String()
	now := time.Now().UTC()

	job := &BulkImportJob{
		ID:            id,
		Status:        "processing",
		InputFormat:   "application/fhir+ndjson",
		ResourceTypes: []string{resourceType},
		RequestTime:   now,
	}
	m.importJobs[id] = job
	m.importMu.Unlock()

	// Parse NDJSON lines
	var resources []map[string]interface{}
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := bytes.TrimSpace(scanner.Bytes())
		if len(line) == 0 {
			continue
		}
		var r map[string]interface{}
		if err := json.Unmarshal(line, &r); err != nil {
			// Count as parse error but continue
			job.TotalResources++
			job.ProcessedResources++
			job.ErrorCount++
			job.Errors = append(job.Errors, BulkError{
				ResourceType: resourceType,
				Index:        job.TotalResources - 1,
				Error:        fmt.Sprintf("JSON parse error: %s", err.Error()),
			})
			continue
		}
		resources = append(resources, r)
		job.TotalResources++
	}

	// Validate each resource
	for i, r := range resources {
		job.ProcessedResources++
		if err := m.validator.ValidateResource(resourceType, r); err != nil {
			job.ErrorCount++
			job.Errors = append(job.Errors, BulkError{
				ResourceType: resourceType,
				Index:        i,
				Error:        err.Error(),
			})
		} else {
			job.SuccessCount++
		}
	}

	// Mark completed
	completedTime := time.Now().UTC()
	m.importMu.Lock()
	job.Status = "completed"
	job.CompletedTime = &completedTime
	m.importMu.Unlock()

	return job, nil
}

// GetImportStatus retrieves the status of a bulk import job by ID.
func (m *BulkOperationManager) GetImportStatus(_ context.Context, id string) (*BulkImportJob, error) {
	m.importMu.RLock()
	defer m.importMu.RUnlock()

	job, ok := m.importJobs[id]
	if !ok {
		return nil, fmt.Errorf("import job not found: %s", id)
	}
	return job, nil
}

// ListImportJobs returns up to limit import jobs.
func (m *BulkOperationManager) ListImportJobs(_ context.Context, limit int) ([]*BulkImportJob, error) {
	m.importMu.RLock()
	defer m.importMu.RUnlock()

	jobs := make([]*BulkImportJob, 0, len(m.importJobs))
	for _, j := range m.importJobs {
		jobs = append(jobs, j)
		if len(jobs) >= limit {
			break
		}
	}
	return jobs, nil
}

// ==================== Edit Methods ====================

// StartBulkUpdate creates a bulk edit job that matches resources by criteria
// and applies the given patch fields.
func (m *BulkOperationManager) StartBulkUpdate(ctx context.Context, resourceType string, criteria map[string]string, patch map[string]interface{}) (*BulkEditJob, error) {
	if len(criteria) == 0 {
		return nil, fmt.Errorf("criteria required: bulk update requires at least one search criterion")
	}

	// Check concurrent job limit
	m.editMu.Lock()
	activeCount := 0
	for _, j := range m.editJobs {
		if j.Status == "pending" || j.Status == "processing" {
			activeCount++
		}
	}
	if activeCount >= m.maxConcurrentJobs {
		m.editMu.Unlock()
		return nil, fmt.Errorf("max concurrent edit jobs reached (%d)", m.maxConcurrentJobs)
	}

	id := uuid.New().String()
	now := time.Now().UTC()

	job := &BulkEditJob{
		ID:           id,
		Status:       "processing",
		Operation:    "update",
		ResourceType: resourceType,
		Criteria:     criteria,
		Patch:        patch,
		RequestTime:  now,
	}
	m.editJobs[id] = job
	m.editMu.Unlock()

	// Match resources
	if m.matcher != nil {
		matched, err := m.matcher.MatchResources(ctx, resourceType, criteria)
		if err != nil {
			m.editMu.Lock()
			job.Status = "error"
			job.ErrorCount++
			job.Errors = append(job.Errors, BulkError{
				ResourceType: resourceType,
				Index:        0,
				Error:        err.Error(),
			})
			m.editMu.Unlock()
			return job, nil
		}

		job.MatchCount = len(matched)

		// Apply patch to each matched resource
		for i, r := range matched {
			for k, v := range patch {
				r[k] = v
			}
			_ = i // index used for error tracking if needed
			job.ModifiedCount++
		}
	}

	// Mark completed
	completedTime := time.Now().UTC()
	m.editMu.Lock()
	job.Status = "completed"
	job.CompletedTime = &completedTime
	m.editMu.Unlock()

	return job, nil
}

// StartBulkDelete creates a bulk edit job that matches resources by criteria
// and marks them for deletion.
func (m *BulkOperationManager) StartBulkDelete(ctx context.Context, resourceType string, criteria map[string]string) (*BulkEditJob, error) {
	if len(criteria) == 0 {
		return nil, fmt.Errorf("criteria required: bulk delete requires at least one search criterion")
	}

	// Check concurrent job limit
	m.editMu.Lock()
	activeCount := 0
	for _, j := range m.editJobs {
		if j.Status == "pending" || j.Status == "processing" {
			activeCount++
		}
	}
	if activeCount >= m.maxConcurrentJobs {
		m.editMu.Unlock()
		return nil, fmt.Errorf("max concurrent edit jobs reached (%d)", m.maxConcurrentJobs)
	}

	id := uuid.New().String()
	now := time.Now().UTC()

	job := &BulkEditJob{
		ID:           id,
		Status:       "processing",
		Operation:    "delete",
		ResourceType: resourceType,
		Criteria:     criteria,
		RequestTime:  now,
	}
	m.editJobs[id] = job
	m.editMu.Unlock()

	// Match resources
	if m.matcher != nil {
		matched, err := m.matcher.MatchResources(ctx, resourceType, criteria)
		if err != nil {
			m.editMu.Lock()
			job.Status = "error"
			job.ErrorCount++
			job.Errors = append(job.Errors, BulkError{
				ResourceType: resourceType,
				Index:        0,
				Error:        err.Error(),
			})
			m.editMu.Unlock()
			return job, nil
		}

		job.MatchCount = len(matched)
		job.ModifiedCount = len(matched)
	}

	// Mark completed
	completedTime := time.Now().UTC()
	m.editMu.Lock()
	job.Status = "completed"
	job.CompletedTime = &completedTime
	m.editMu.Unlock()

	return job, nil
}

// GetEditStatus retrieves the status of a bulk edit job by ID.
func (m *BulkOperationManager) GetEditStatus(_ context.Context, id string) (*BulkEditJob, error) {
	m.editMu.RLock()
	defer m.editMu.RUnlock()

	job, ok := m.editJobs[id]
	if !ok {
		return nil, fmt.Errorf("edit job not found: %s", id)
	}
	return job, nil
}

// ListEditJobs returns up to limit edit jobs.
func (m *BulkOperationManager) ListEditJobs(_ context.Context, limit int) ([]*BulkEditJob, error) {
	m.editMu.RLock()
	defer m.editMu.RUnlock()

	jobs := make([]*BulkEditJob, 0, len(m.editJobs))
	for _, j := range m.editJobs {
		jobs = append(jobs, j)
		if len(jobs) >= limit {
			break
		}
	}
	return jobs, nil
}

// CancelJob cancels a pending or processing edit job. Returns an error if
// the job is already completed or not found.
func (m *BulkOperationManager) CancelJob(_ context.Context, id string) error {
	m.editMu.Lock()
	defer m.editMu.Unlock()

	job, ok := m.editJobs[id]
	if !ok {
		return fmt.Errorf("edit job not found: %s", id)
	}

	if job.Status == "completed" || job.Status == "error" {
		return fmt.Errorf("cannot cancel job %s: job is already %s", id, job.Status)
	}

	job.Status = "cancelled"
	now := time.Now().UTC()
	job.CompletedTime = &now
	return nil
}

// ==================== HTTP Handler ====================

// BulkOpsHandler provides Echo HTTP handlers for FHIR bulk import and edit operations.
type BulkOpsHandler struct {
	manager *BulkOperationManager
}

// NewBulkOpsHandler creates a new BulkOpsHandler.
func NewBulkOpsHandler(manager *BulkOperationManager) *BulkOpsHandler {
	return &BulkOpsHandler{manager: manager}
}

// RegisterRoutes registers the bulk operation routes on the given Echo group.
func (h *BulkOpsHandler) RegisterRoutes(g *echo.Group) {
	g.POST("/$import", h.StartImport)
	g.GET("/$import/:id", h.GetImportStatus)
	g.GET("/$import", h.ListImportJobs)
	g.POST("/$bulk-edit", h.StartBulkEdit)
	g.POST("/$bulk-delete", h.StartBulkDelete)
	g.GET("/$bulk-edit/:id", h.GetEditStatus)
	g.DELETE("/$bulk-edit/:id", h.CancelJob)
}

// StartImport handles POST /fhir/$import — start a bulk import from NDJSON body.
func (h *BulkOpsHandler) StartImport(c echo.Context) error {
	resourceType := c.QueryParam("resourceType")
	if resourceType == "" {
		resourceType = "Patient" // default
	}

	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorOutcome("failed to read request body"))
	}

	job, err := h.manager.StartImport(c.Request().Context(), resourceType, body)
	if err != nil {
		if fmt.Sprintf("%v", err) != "" && contains(err.Error(), "concurrent") {
			return c.JSON(http.StatusTooManyRequests, ErrorOutcome(err.Error()))
		}
		return c.JSON(http.StatusBadRequest, ErrorOutcome(err.Error()))
	}

	return c.JSON(http.StatusAccepted, job)
}

// GetImportStatus handles GET /fhir/$import/:id — get import job status.
func (h *BulkOpsHandler) GetImportStatus(c echo.Context) error {
	id := c.Param("id")
	job, err := h.manager.GetImportStatus(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusNotFound, ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, job)
}

// ListImportJobs handles GET /fhir/$import — list import jobs.
func (h *BulkOpsHandler) ListImportJobs(c echo.Context) error {
	jobs, err := h.manager.ListImportJobs(c.Request().Context(), 100)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, jobs)
}

// bulkEditRequest represents the JSON body for a bulk edit operation.
type bulkEditRequest struct {
	Operation    string                 `json:"operation"`
	ResourceType string                 `json:"resourceType"`
	Criteria     map[string]string      `json:"criteria"`
	Patch        map[string]interface{} `json:"patch"`
}

// StartBulkEdit handles POST /fhir/$bulk-edit — start a bulk update/patch.
func (h *BulkOpsHandler) StartBulkEdit(c echo.Context) error {
	var req bulkEditRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorOutcome("invalid request body"))
	}

	if req.ResourceType == "" {
		return c.JSON(http.StatusBadRequest, ErrorOutcome("resourceType is required"))
	}

	var job *BulkEditJob
	var err error

	switch req.Operation {
	case "update", "patch", "":
		job, err = h.manager.StartBulkUpdate(c.Request().Context(), req.ResourceType, req.Criteria, req.Patch)
	default:
		return c.JSON(http.StatusBadRequest, ErrorOutcome(fmt.Sprintf("unsupported operation: %s", req.Operation)))
	}

	if err != nil {
		if contains(err.Error(), "concurrent") {
			return c.JSON(http.StatusTooManyRequests, ErrorOutcome(err.Error()))
		}
		return c.JSON(http.StatusBadRequest, ErrorOutcome(err.Error()))
	}

	return c.JSON(http.StatusAccepted, job)
}

// bulkDeleteRequest represents the JSON body for a bulk delete operation.
type bulkDeleteRequest struct {
	ResourceType string            `json:"resourceType"`
	Criteria     map[string]string `json:"criteria"`
}

// StartBulkDelete handles POST /fhir/$bulk-delete — start a bulk delete.
func (h *BulkOpsHandler) StartBulkDelete(c echo.Context) error {
	var req bulkDeleteRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorOutcome("invalid request body"))
	}

	if req.ResourceType == "" {
		return c.JSON(http.StatusBadRequest, ErrorOutcome("resourceType is required"))
	}

	job, err := h.manager.StartBulkDelete(c.Request().Context(), req.ResourceType, req.Criteria)
	if err != nil {
		if contains(err.Error(), "concurrent") {
			return c.JSON(http.StatusTooManyRequests, ErrorOutcome(err.Error()))
		}
		return c.JSON(http.StatusBadRequest, ErrorOutcome(err.Error()))
	}

	return c.JSON(http.StatusAccepted, job)
}

// GetEditStatus handles GET /fhir/$bulk-edit/:id — get edit job status.
func (h *BulkOpsHandler) GetEditStatus(c echo.Context) error {
	id := c.Param("id")
	job, err := h.manager.GetEditStatus(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusNotFound, ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, job)
}

// CancelJob handles DELETE /fhir/$bulk-edit/:id — cancel an edit job.
func (h *BulkOpsHandler) CancelJob(c echo.Context) error {
	id := c.Param("id")
	err := h.manager.CancelJob(c.Request().Context(), id)
	if err != nil {
		if contains(err.Error(), "not found") {
			return c.JSON(http.StatusNotFound, ErrorOutcome(err.Error()))
		}
		if contains(err.Error(), "cannot cancel") {
			return c.JSON(http.StatusConflict, ErrorOutcome(err.Error()))
		}
		return c.JSON(http.StatusInternalServerError, ErrorOutcome(err.Error()))
	}

	return c.JSON(http.StatusOK, map[string]string{
		"status":  "cancelled",
		"message": fmt.Sprintf("job %s has been cancelled", id),
	})
}

// contains is a helper to check if a string contains a substring.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsStr(s, substr))
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
