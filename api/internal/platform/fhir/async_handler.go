package fhir

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
)

// AsyncJob represents the state of an asynchronous FHIR request.
type AsyncJob struct {
	ID            string           `json:"id"`
	Status        string           `json:"status"` // "in-progress", "completed", "error"
	ResourceType  string           `json:"resourceType,omitempty"`
	Request       string           `json:"request"`
	TransactionTS time.Time        `json:"transactionTime"`
	Output        []AsyncJobOutput `json:"output,omitempty"`
	Error         string           `json:"error,omitempty"`
}

// AsyncJobOutput describes one output file produced by a completed async job.
type AsyncJobOutput struct {
	Type string `json:"type"`
	URL  string `json:"url"`
}

// Async job status constants.
const (
	AsyncStatusInProgress = "in-progress"
	AsyncStatusCompleted  = "completed"
	AsyncStatusError      = "error"
)

// AsyncJobStore defines the persistence interface for async job tracking.
type AsyncJobStore interface {
	Create(ctx context.Context, job *AsyncJob) error
	Get(ctx context.Context, jobID string) (*AsyncJob, error)
	Update(ctx context.Context, job *AsyncJob) error
	Delete(ctx context.Context, jobID string) error
}

// InMemoryAsyncJobStore is a concurrency-safe, in-memory implementation of AsyncJobStore.
type InMemoryAsyncJobStore struct {
	mu   sync.RWMutex
	jobs map[string]*AsyncJob
}

// NewInMemoryAsyncJobStore creates an empty InMemoryAsyncJobStore.
func NewInMemoryAsyncJobStore() *InMemoryAsyncJobStore {
	return &InMemoryAsyncJobStore{
		jobs: make(map[string]*AsyncJob),
	}
}

// Create adds a new async job to the store. If the job ID is empty a unique
// identifier is generated automatically.
func (s *InMemoryAsyncJobStore) Create(_ context.Context, job *AsyncJob) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if job.ID == "" {
		job.ID = fmt.Sprintf("job-%d", time.Now().UnixNano())
	}
	if job.TransactionTS.IsZero() {
		job.TransactionTS = time.Now().UTC()
	}
	if job.Status == "" {
		job.Status = AsyncStatusInProgress
	}

	// Store a copy so callers cannot mutate the map entry.
	cp := *job
	s.jobs[job.ID] = &cp
	return nil
}

// Get retrieves an async job by ID. Returns an error when the job is not found.
func (s *InMemoryAsyncJobStore) Get(_ context.Context, jobID string) (*AsyncJob, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	job, ok := s.jobs[jobID]
	if !ok {
		return nil, fmt.Errorf("async job %s not found", jobID)
	}
	cp := *job
	return &cp, nil
}

// Update replaces an existing async job in the store.
func (s *InMemoryAsyncJobStore) Update(_ context.Context, job *AsyncJob) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.jobs[job.ID]; !ok {
		return fmt.Errorf("async job %s not found", job.ID)
	}
	cp := *job
	s.jobs[job.ID] = &cp
	return nil
}

// Delete removes an async job from the store. It is not an error to delete a
// job that does not exist.
func (s *InMemoryAsyncJobStore) Delete(_ context.Context, jobID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.jobs, jobID)
	return nil
}

// AsyncStatusHandler returns an echo.HandlerFunc that serves
// GET /_async/:jobId and reports the current status of an async job.
//
// Behaviour by job status:
//   - "in-progress": 202 Accepted with an X-Progress header.
//   - "completed":   200 OK with the job output as JSON.
//   - "error":       500 Internal Server Error with an OperationOutcome.
//   - not found:     404 Not Found with an OperationOutcome.
func AsyncStatusHandler(store AsyncJobStore) echo.HandlerFunc {
	return func(c echo.Context) error {
		jobID := c.Param("jobId")
		if jobID == "" {
			return c.JSON(http.StatusBadRequest, NewOperationOutcome(
				IssueSeverityError, IssueTypeRequired, "jobId parameter is required",
			))
		}

		job, err := store.Get(c.Request().Context(), jobID)
		if err != nil {
			return c.JSON(http.StatusNotFound, NewOperationOutcome(
				IssueSeverityError, IssueTypeNotFound, fmt.Sprintf("async job %s not found", jobID),
			))
		}

		switch job.Status {
		case AsyncStatusInProgress:
			c.Response().Header().Set("X-Progress", "in-progress")
			return c.NoContent(http.StatusAccepted)

		case AsyncStatusCompleted:
			body := map[string]interface{}{
				"transactionTime": job.TransactionTS.Format(time.RFC3339),
				"request":         job.Request,
				"output":          job.Output,
			}
			return c.JSON(http.StatusOK, body)

		case AsyncStatusError:
			return c.JSON(http.StatusInternalServerError, NewOperationOutcome(
				IssueSeverityError, IssueTypeException, job.Error,
			))

		default:
			return c.JSON(http.StatusInternalServerError, NewOperationOutcome(
				IssueSeverityError, IssueTypeException, "unknown job status",
			))
		}
	}
}

// AsyncDeleteHandler returns an echo.HandlerFunc that serves
// DELETE /_async/:jobId to cancel or remove a completed async job.
func AsyncDeleteHandler(store AsyncJobStore) echo.HandlerFunc {
	return func(c echo.Context) error {
		jobID := c.Param("jobId")
		if jobID == "" {
			return c.JSON(http.StatusBadRequest, NewOperationOutcome(
				IssueSeverityError, IssueTypeRequired, "jobId parameter is required",
			))
		}

		if err := store.Delete(c.Request().Context(), jobID); err != nil {
			return c.JSON(http.StatusInternalServerError, NewOperationOutcome(
				IssueSeverityError, IssueTypeException, err.Error(),
			))
		}

		return c.NoContent(http.StatusAccepted)
	}
}

// RespondAsync writes a 202 Accepted response with the Content-Location header
// pointing to the async polling endpoint for the given job ID.
func RespondAsync(c echo.Context, store AsyncJobStore, jobID string) error {
	_ = store // store kept in signature for future use (e.g. inline creation)
	c.Response().Header().Set("Content-Location", fmt.Sprintf("/_async/%s", jobID))
	return c.NoContent(http.StatusAccepted)
}

// ParsePreferAsync checks whether the Prefer header value contains the
// "respond-async" preference token as defined in RFC 7240.
func ParsePreferAsync(prefer string) bool {
	for _, sep := range []string{",", ";"} {
		for _, part := range strings.Split(prefer, sep) {
			if strings.TrimSpace(part) == "respond-async" {
				return true
			}
		}
	}
	return false
}

// operationOutcomeJSON is a small helper to marshal an OperationOutcome for
// use in non-echo contexts. Callers in this file use echo's c.JSON instead.
func operationOutcomeJSON(severity, code, diagnostics string) json.RawMessage {
	out := NewOperationOutcome(severity, code, diagnostics)
	b, _ := json.Marshal(out)
	return b
}
