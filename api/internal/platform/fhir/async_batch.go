package fhir

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// AsyncBatchConfig controls when batch processing switches to async mode.
type AsyncBatchConfig struct {
	MaxSyncEntries   int           // Max entries to process synchronously (e.g., 100)
	MaxSyncTimeout   time.Duration // Max time for sync processing before switching to async
	MaxAsyncEntries  int           // Max entries allowed in async batch (e.g., 10000)
	WorkerCount      int           // Number of concurrent workers for async processing
	RetryMaxAttempts int           // Max retries per entry
	RetryDelay       time.Duration // Base delay between retries
	ProgressInterval time.Duration // How often to update job progress
}

// AsyncBatchJob represents a batch/transaction processing job.
type AsyncBatchJob struct {
	ID               string                 `json:"id"`
	Status           string                 `json:"status"` // queued, processing, completed, failed, cancelled
	BundleType       string                 `json:"bundleType"`
	TotalEntries     int                    `json:"totalEntries"`
	ProcessedEntries int                    `json:"processedEntries"`
	SuccessCount     int                    `json:"successCount"`
	ErrorCount       int                    `json:"errorCount"`
	StartTime        time.Time              `json:"startTime"`
	EndTime          *time.Time             `json:"endTime,omitempty"`
	Progress         float64                `json:"progress"`
	ResultBundle     map[string]interface{} `json:"resultBundle,omitempty"`
	Errors           []AsyncBatchError      `json:"errors,omitempty"`
	Request          string                 `json:"request"`
	CreatedAt        time.Time              `json:"createdAt"`
}

// AsyncBatchError describes an error for a specific entry.
type AsyncBatchError struct {
	EntryIndex  int    `json:"entryIndex"`
	Method      string `json:"method"`
	URL         string `json:"url"`
	StatusCode  int    `json:"statusCode"`
	Diagnostics string `json:"diagnostics"`
	Severity    string `json:"severity"`
}

// AsyncBatchProgress represents a progress update.
type AsyncBatchProgress struct {
	JobID             string        `json:"jobId"`
	Processed         int           `json:"processed"`
	Total             int           `json:"total"`
	SuccessCount      int           `json:"successCount"`
	ErrorCount        int           `json:"errorCount"`
	Progress          float64       `json:"progress"`
	EstimatedTimeLeft time.Duration `json:"estimatedTimeLeft"`
}

// AsyncBatchStore interface for job persistence.
type AsyncBatchStore interface {
	CreateJob(ctx interface{}, job *AsyncBatchJob) error
	GetJob(ctx interface{}, jobID string) (*AsyncBatchJob, error)
	UpdateJob(ctx interface{}, job *AsyncBatchJob) error
	CancelJob(ctx interface{}, jobID string) error
	ListJobs(ctx interface{}, status string, limit int) ([]*AsyncBatchJob, error)
	DeleteJob(ctx interface{}, jobID string) error
}

// InMemoryAsyncBatchStore is a test/demo implementation of AsyncBatchStore.
type InMemoryAsyncBatchStore struct {
	mu   sync.RWMutex
	jobs map[string]*AsyncBatchJob
}

// AsyncBatchProcessor handles background processing of batch/transaction bundles.
type AsyncBatchProcessor struct {
	store     AsyncBatchStore
	config    AsyncBatchConfig
	handler   func(method, url string, resource map[string]interface{}) (*BundleEntryResponse, error)
	cancelFns map[string]context.CancelFunc
	mu        sync.Mutex
}

// NewInMemoryAsyncBatchStore creates a new in-memory store.
func NewInMemoryAsyncBatchStore() *InMemoryAsyncBatchStore {
	return &InMemoryAsyncBatchStore{
		jobs: make(map[string]*AsyncBatchJob),
	}
}

// CreateJob adds a new async batch job to the store. If the job ID is empty,
// a unique identifier is generated automatically.
func (s *InMemoryAsyncBatchStore) CreateJob(_ interface{}, job *AsyncBatchJob) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if job.ID == "" {
		job.ID = uuid.New().String()
	}
	if job.Status == "" {
		job.Status = "queued"
	}
	if job.CreatedAt.IsZero() {
		job.CreatedAt = time.Now().UTC()
	}
	if job.StartTime.IsZero() {
		job.StartTime = time.Now().UTC()
	}

	cp := copyAsyncBatchJob(job)
	s.jobs[job.ID] = cp
	return nil
}

// GetJob retrieves an async batch job by ID.
func (s *InMemoryAsyncBatchStore) GetJob(_ interface{}, jobID string) (*AsyncBatchJob, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	job, ok := s.jobs[jobID]
	if !ok {
		return nil, fmt.Errorf("async batch job %s not found", jobID)
	}
	return copyAsyncBatchJob(job), nil
}

// UpdateJob replaces an existing async batch job in the store.
func (s *InMemoryAsyncBatchStore) UpdateJob(_ interface{}, job *AsyncBatchJob) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.jobs[job.ID]; !ok {
		return fmt.Errorf("async batch job %s not found", job.ID)
	}
	s.jobs[job.ID] = copyAsyncBatchJob(job)
	return nil
}

// CancelJob marks an async batch job as cancelled.
func (s *InMemoryAsyncBatchStore) CancelJob(_ interface{}, jobID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	job, ok := s.jobs[jobID]
	if !ok {
		return fmt.Errorf("async batch job %s not found", jobID)
	}
	job.Status = "cancelled"
	now := time.Now().UTC()
	job.EndTime = &now
	return nil
}

// ListJobs returns jobs, optionally filtered by status and limited in count.
func (s *InMemoryAsyncBatchStore) ListJobs(_ interface{}, status string, limit int) ([]*AsyncBatchJob, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*AsyncBatchJob, 0, len(s.jobs))
	for _, job := range s.jobs {
		if status != "" && job.Status != status {
			continue
		}
		result = append(result, copyAsyncBatchJob(job))
		if limit > 0 && len(result) >= limit {
			break
		}
	}
	return result, nil
}

// DeleteJob removes an async batch job from the store.
func (s *InMemoryAsyncBatchStore) DeleteJob(_ interface{}, jobID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.jobs[jobID]; !ok {
		return fmt.Errorf("async batch job %s not found", jobID)
	}
	delete(s.jobs, jobID)
	return nil
}

// copyAsyncBatchJob returns a deep copy of an AsyncBatchJob to avoid data races.
func copyAsyncBatchJob(job *AsyncBatchJob) *AsyncBatchJob {
	cp := *job
	if job.Errors != nil {
		cp.Errors = make([]AsyncBatchError, len(job.Errors))
		copy(cp.Errors, job.Errors)
	}
	if job.ResultBundle != nil {
		cp.ResultBundle = make(map[string]interface{}, len(job.ResultBundle))
		for k, v := range job.ResultBundle {
			cp.ResultBundle[k] = v
		}
	}
	if job.EndTime != nil {
		t := *job.EndTime
		cp.EndTime = &t
	}
	return &cp
}

// NewAsyncBatchProcessor creates a new processor.
func NewAsyncBatchProcessor(
	store AsyncBatchStore,
	config AsyncBatchConfig,
	handler func(method, url string, resource map[string]interface{}) (*BundleEntryResponse, error),
) *AsyncBatchProcessor {
	return &AsyncBatchProcessor{
		store:     store,
		config:    config,
		handler:   handler,
		cancelFns: make(map[string]context.CancelFunc),
	}
}

// DefaultAsyncBatchConfig returns sensible defaults.
func DefaultAsyncBatchConfig() AsyncBatchConfig {
	return AsyncBatchConfig{
		MaxSyncEntries:   100,
		MaxSyncTimeout:   30 * time.Second,
		MaxAsyncEntries:  10000,
		WorkerCount:      4,
		RetryMaxAttempts: 3,
		RetryDelay:       500 * time.Millisecond,
		ProgressInterval: 5 * time.Second,
	}
}

// ShouldProcessAsync determines if a bundle should be processed asynchronously
// based on bundle size, configuration thresholds, and whether the client
// requested async processing via the Prefer header.
func ShouldProcessAsync(bundle *TransactionBundle, config *AsyncBatchConfig, preferAsync bool) bool {
	if bundle == nil {
		return false
	}
	if preferAsync {
		return true
	}
	return len(bundle.Entries) > config.MaxSyncEntries
}

// SubmitAsyncBatch submits a batch/transaction for async processing and returns
// the job ID. The bundle is validated against the maximum async entry limit.
func (p *AsyncBatchProcessor) SubmitAsyncBatch(bundle *TransactionBundle, requestURI string) (string, error) {
	if len(bundle.Entries) > p.config.MaxAsyncEntries {
		return "", fmt.Errorf("bundle contains %d entries which exceeds maximum of %d",
			len(bundle.Entries), p.config.MaxAsyncEntries)
	}

	job := &AsyncBatchJob{
		BundleType:   bundle.Type,
		TotalEntries: len(bundle.Entries),
		Request:      requestURI,
	}

	ctx := context.Background()
	if err := p.store.CreateJob(ctx, job); err != nil {
		return "", fmt.Errorf("failed to create async batch job: %w", err)
	}

	return job.ID, nil
}

// ProcessBatchAsync processes a batch bundle asynchronously with progress
// tracking. Each entry is processed independently; failures are captured
// per-entry and do not affect other entries.
func (p *AsyncBatchProcessor) ProcessBatchAsync(jobID string, bundle *TransactionBundle) {
	ctx := context.Background()

	// Mark job as processing.
	job, err := p.store.GetJob(ctx, jobID)
	if err != nil {
		return
	}
	job.Status = "processing"
	job.StartTime = time.Now().UTC()
	_ = p.store.UpdateJob(ctx, job)

	entryCount := len(bundle.Entries)
	if entryCount == 0 {
		now := time.Now().UTC()
		job.Status = "completed"
		job.Progress = 1.0
		job.EndTime = &now
		job.ResultBundle = map[string]interface{}{
			"resourceType": "Bundle",
			"type":         "batch-response",
			"entry":        []interface{}{},
		}
		_ = p.store.UpdateJob(ctx, job)
		return
	}

	// Process entries with worker pool.
	type entryResult struct {
		index    int
		response *BundleEntryResponse
		err      error
	}

	results := make([]entryResult, entryCount)
	var wg sync.WaitGroup
	sem := make(chan struct{}, p.config.WorkerCount)

	var mu sync.Mutex
	processedCount := 0

	for i, entry := range bundle.Entries {
		wg.Add(1)
		sem <- struct{}{}
		go func(idx int, e TransactionEntry) {
			defer wg.Done()
			defer func() { <-sem }()

			resp, err := ProcessEntryWithRetry(&e, p.handler, &p.config)
			results[idx] = entryResult{index: idx, response: resp, err: err}

			mu.Lock()
			processedCount++
			currentCount := processedCount
			progress := float64(currentCount) / float64(entryCount)
			mu.Unlock()

			// Periodically update job progress.
			updated, getErr := p.store.GetJob(ctx, jobID)
			if getErr == nil {
				updated.ProcessedEntries = currentCount
				updated.Progress = progress
				_ = p.store.UpdateJob(ctx, updated)
			}
		}(i, entry)
	}
	wg.Wait()

	// Aggregate results.
	job, _ = p.store.GetJob(ctx, jobID)
	successCount := 0
	errorCount := 0
	var errors []AsyncBatchError
	responseEntries := make([]interface{}, entryCount)

	for _, r := range results {
		if r.err != nil {
			errorCount++
			errors = append(errors, AsyncBatchError{
				EntryIndex:  r.index,
				Method:      bundle.Entries[r.index].Request.Method,
				URL:         bundle.Entries[r.index].Request.URL,
				StatusCode:  400,
				Diagnostics: r.err.Error(),
				Severity:    "error",
			})
			responseEntries[r.index] = map[string]interface{}{
				"response": map[string]interface{}{
					"status":  "400 Bad Request",
					"outcome": map[string]interface{}{"resourceType": "OperationOutcome", "issue": []map[string]interface{}{{"severity": "error", "code": "processing", "diagnostics": r.err.Error()}}},
				},
			}
		} else {
			successCount++
			entry := map[string]interface{}{
				"response": map[string]interface{}{
					"status": r.response.Status,
				},
			}
			if r.response.Location != "" {
				entry["fullUrl"] = r.response.Location
				resp := entry["response"].(map[string]interface{})
				resp["location"] = r.response.Location
			}
			responseEntries[r.index] = entry
		}
	}

	now := time.Now().UTC()
	job.Status = "completed"
	job.ProcessedEntries = entryCount
	job.SuccessCount = successCount
	job.ErrorCount = errorCount
	job.Progress = 1.0
	job.EndTime = &now
	job.Errors = errors
	job.ResultBundle = map[string]interface{}{
		"resourceType": "Bundle",
		"type":         "batch-response",
		"entry":        responseEntries,
	}
	_ = p.store.UpdateJob(ctx, job)
}

// ProcessTransactionAsync processes a transaction bundle asynchronously with
// atomic semantics. If any entry fails, the entire transaction is marked as
// failed and errors are recorded.
func (p *AsyncBatchProcessor) ProcessTransactionAsync(jobID string, bundle *TransactionBundle) {
	ctx := context.Background()

	// Mark job as processing.
	job, err := p.store.GetJob(ctx, jobID)
	if err != nil {
		return
	}
	job.Status = "processing"
	job.StartTime = time.Now().UTC()
	_ = p.store.UpdateJob(ctx, job)

	// Sort entries according to FHIR transaction processing order.
	sorted := SortTransactionEntries(bundle.Entries)
	entryCount := len(sorted)

	idMap := make(map[string]string)
	responseEntries := make([]interface{}, entryCount)
	successCount := 0

	for i, entry := range sorted {
		// Resolve urn:uuid references in resources.
		if entry.Resource != nil && len(idMap) > 0 {
			resolveRefsInResource(entry.Resource, idMap)
		}
		url := replaceURNRefs(entry.Request.URL, idMap)

		resp, entryErr := p.handler(entry.Request.Method, url, entry.Resource)
		if entryErr != nil {
			// Transaction fails atomically.
			now := time.Now().UTC()
			job, _ = p.store.GetJob(ctx, jobID)
			job.Status = "failed"
			job.ProcessedEntries = i + 1
			job.SuccessCount = successCount
			job.ErrorCount = 1
			job.Progress = float64(i+1) / float64(entryCount)
			job.EndTime = &now
			job.Errors = []AsyncBatchError{
				{
					EntryIndex:  i,
					Method:      entry.Request.Method,
					URL:         entry.Request.URL,
					StatusCode:  400,
					Diagnostics: fmt.Sprintf("transaction failed at entry %d: %s", i, entryErr.Error()),
					Severity:    "error",
				},
			}
			_ = p.store.UpdateJob(ctx, job)
			return
		}

		// Map urn:uuid references.
		if entry.FullURL != "" && strings.HasPrefix(entry.FullURL, "urn:uuid:") && resp.Location != "" {
			idMap[entry.FullURL] = resp.Location
		}

		successCount++
		respEntry := map[string]interface{}{
			"response": map[string]interface{}{
				"status": resp.Status,
			},
		}
		if resp.Location != "" {
			respEntry["fullUrl"] = resp.Location
			re := respEntry["response"].(map[string]interface{})
			re["location"] = resp.Location
		}
		responseEntries[i] = respEntry

		// Update progress.
		updated, getErr := p.store.GetJob(ctx, jobID)
		if getErr == nil {
			updated.ProcessedEntries = i + 1
			updated.SuccessCount = successCount
			updated.Progress = float64(i+1) / float64(entryCount)
			_ = p.store.UpdateJob(ctx, updated)
		}
	}

	// All entries succeeded.
	now := time.Now().UTC()
	job, _ = p.store.GetJob(ctx, jobID)
	job.Status = "completed"
	job.ProcessedEntries = entryCount
	job.SuccessCount = successCount
	job.ErrorCount = 0
	job.Progress = 1.0
	job.EndTime = &now
	job.ResultBundle = map[string]interface{}{
		"resourceType": "Bundle",
		"type":         "transaction-response",
		"entry":        responseEntries,
	}
	_ = p.store.UpdateJob(ctx, job)
}

// CancelJob cancels a running async job. Returns an error if the job is not
// found or is already in a terminal state (completed, failed).
func (p *AsyncBatchProcessor) CancelJob(jobID string) error {
	ctx := context.Background()
	job, err := p.store.GetJob(ctx, jobID)
	if err != nil {
		return err
	}

	if job.Status == "completed" || job.Status == "failed" {
		return fmt.Errorf("cannot cancel job %s in status %q", jobID, job.Status)
	}

	// Cancel the context if we have a cancel function for this job.
	p.mu.Lock()
	if cancelFn, ok := p.cancelFns[jobID]; ok {
		cancelFn()
		delete(p.cancelFns, jobID)
	}
	p.mu.Unlock()

	return p.store.CancelJob(ctx, jobID)
}

// GetJobProgress returns the current progress of a job.
func (p *AsyncBatchProcessor) GetJobProgress(jobID string) (*AsyncBatchProgress, error) {
	ctx := context.Background()
	job, err := p.store.GetJob(ctx, jobID)
	if err != nil {
		return nil, err
	}

	progress := &AsyncBatchProgress{
		JobID:        job.ID,
		Processed:    job.ProcessedEntries,
		Total:        job.TotalEntries,
		SuccessCount: job.SuccessCount,
		ErrorCount:   job.ErrorCount,
		Progress:     job.Progress,
	}

	// Estimate remaining time based on elapsed time and progress.
	if job.Progress > 0 && job.Progress < 1.0 {
		elapsed := time.Since(job.StartTime)
		totalEstimated := time.Duration(float64(elapsed) / job.Progress)
		progress.EstimatedTimeLeft = totalEstimated - elapsed
	}

	return progress, nil
}

// BuildProgressResponse builds a FHIR OperationOutcome showing progress.
func BuildProgressResponse(progress *AsyncBatchProgress) map[string]interface{} {
	progressPct := int(progress.Progress * 100)
	diagnostics := fmt.Sprintf(
		"Processing %d/%d entries (%d%% complete). Success: %d, Errors: %d.",
		progress.Processed, progress.Total, progressPct,
		progress.SuccessCount, progress.ErrorCount,
	)
	if progress.EstimatedTimeLeft > 0 {
		diagnostics += fmt.Sprintf(" Estimated time remaining: %s.", progress.EstimatedTimeLeft.Round(time.Second))
	}

	return map[string]interface{}{
		"resourceType": "OperationOutcome",
		"issue": []map[string]interface{}{
			{
				"severity":    "information",
				"code":        "informational",
				"diagnostics": diagnostics,
			},
		},
	}
}

// BuildAsyncBatchResult builds the final response bundle from a completed job.
// If the job's ResultBundle is already populated, it is returned directly.
// Otherwise, a minimal bundle is constructed.
func BuildAsyncBatchResult(job *AsyncBatchJob) map[string]interface{} {
	if job.ResultBundle != nil {
		return job.ResultBundle
	}

	responseType := "batch-response"
	if job.BundleType == "transaction" {
		responseType = "transaction-response"
	}

	return map[string]interface{}{
		"resourceType": "Bundle",
		"type":         responseType,
		"entry":        []interface{}{},
	}
}

// AsyncBatchHandler returns an echo.HandlerFunc for async batch detection.
// It reads the request body, determines whether the bundle should be processed
// asynchronously, and either submits it for background processing (returning
// 202 Accepted with Content-Location) or indicates the bundle should be
// processed synchronously.
func AsyncBatchHandler(processor *AsyncBatchProcessor, config *AsyncBatchConfig) echo.HandlerFunc {
	return func(c echo.Context) error {
		body, err := io.ReadAll(c.Request().Body)
		if err != nil {
			return c.JSON(http.StatusBadRequest, NewOperationOutcome(
				IssueSeverityError, IssueTypeStructure,
				fmt.Sprintf("failed to read request body: %s", err.Error()),
			))
		}

		bundle, err := ParseTransactionBundle(body)
		if err != nil {
			return c.JSON(http.StatusBadRequest, NewOperationOutcome(
				IssueSeverityError, IssueTypeStructure,
				fmt.Sprintf("failed to parse Bundle: %s", err.Error()),
			))
		}

		if bundle.Type != "batch" && bundle.Type != "transaction" {
			return c.JSON(http.StatusBadRequest, NewOperationOutcome(
				IssueSeverityError, IssueTypeValue,
				fmt.Sprintf("unsupported bundle type %q; expected 'transaction' or 'batch'", bundle.Type),
			))
		}

		preferAsync := ParsePreferAsync(c.Request().Header.Get("Prefer"))

		if !ShouldProcessAsync(bundle, config, preferAsync) {
			// Return a pass-through response indicating sync processing.
			// The caller (middleware chain) should handle this by delegating
			// to the synchronous TransactionHandler.
			return c.JSON(http.StatusOK, map[string]interface{}{
				"async": false,
			})
		}

		// Submit for async processing.
		requestURI := c.Request().Method + " " + c.Request().RequestURI
		jobID, err := processor.SubmitAsyncBatch(bundle, requestURI)
		if err != nil {
			return c.JSON(http.StatusBadRequest, NewOperationOutcome(
				IssueSeverityError, IssueTypeProcessing, err.Error(),
			))
		}

		// Start background processing.
		go func() {
			switch bundle.Type {
			case "batch":
				processor.ProcessBatchAsync(jobID, bundle)
			case "transaction":
				processor.ProcessTransactionAsync(jobID, bundle)
			}
		}()

		// Return 202 Accepted with Content-Location for polling.
		c.Response().Header().Set("Content-Location", fmt.Sprintf("/fhir/_async-batch/%s", jobID))
		return c.NoContent(http.StatusAccepted)
	}
}

// AsyncBatchStatusHandler returns a handler for polling job status.
// GET /fhir/_async-batch/:jobId
//
// Behaviour by job status:
//   - "queued"/"processing": 202 Accepted with X-Progress header.
//   - "completed":           200 OK with the result bundle.
//   - "failed":              500 Internal Server Error with OperationOutcome.
//   - "cancelled":           410 Gone with OperationOutcome.
//   - not found:             404 Not Found with OperationOutcome.
func AsyncBatchStatusHandler(store AsyncBatchStore) echo.HandlerFunc {
	return func(c echo.Context) error {
		jobID := c.Param("jobId")
		if jobID == "" {
			return c.JSON(http.StatusBadRequest, NewOperationOutcome(
				IssueSeverityError, IssueTypeRequired, "jobId parameter is required",
			))
		}

		job, err := store.GetJob(c.Request().Context(), jobID)
		if err != nil {
			return c.JSON(http.StatusNotFound, NewOperationOutcome(
				IssueSeverityError, IssueTypeNotFound,
				fmt.Sprintf("async batch job %s not found", jobID),
			))
		}

		switch job.Status {
		case "queued", "processing":
			progressPct := int(job.Progress * 100)
			c.Response().Header().Set("X-Progress",
				fmt.Sprintf("%d/%d entries processed (%d%%)",
					job.ProcessedEntries, job.TotalEntries, progressPct))
			c.Response().Header().Set("Retry-After", "5")
			return c.NoContent(http.StatusAccepted)

		case "completed":
			result := BuildAsyncBatchResult(job)
			return c.JSON(http.StatusOK, result)

		case "failed":
			diagnostics := "async batch job failed"
			if len(job.Errors) > 0 {
				diagnostics = job.Errors[0].Diagnostics
			}
			return c.JSON(http.StatusInternalServerError, NewOperationOutcome(
				IssueSeverityError, IssueTypeException, diagnostics,
			))

		case "cancelled":
			return c.JSON(http.StatusGone, NewOperationOutcome(
				IssueSeverityError, IssueTypeProcessing,
				fmt.Sprintf("async batch job %s has been cancelled", jobID),
			))

		default:
			return c.JSON(http.StatusInternalServerError, NewOperationOutcome(
				IssueSeverityError, IssueTypeException, "unknown job status",
			))
		}
	}
}

// AsyncBatchCancelHandler returns a handler for cancelling async jobs.
// DELETE /fhir/_async-batch/:jobId
func AsyncBatchCancelHandler(processor *AsyncBatchProcessor) echo.HandlerFunc {
	return func(c echo.Context) error {
		jobID := c.Param("jobId")
		if jobID == "" {
			return c.JSON(http.StatusBadRequest, NewOperationOutcome(
				IssueSeverityError, IssueTypeRequired, "jobId parameter is required",
			))
		}

		err := processor.CancelJob(jobID)
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				return c.JSON(http.StatusNotFound, NewOperationOutcome(
					IssueSeverityError, IssueTypeNotFound,
					fmt.Sprintf("async batch job %s not found", jobID),
				))
			}
			return c.JSON(http.StatusConflict, NewOperationOutcome(
				IssueSeverityError, IssueTypeConflict, err.Error(),
			))
		}

		return c.NoContent(http.StatusAccepted)
	}
}

// AsyncBatchListHandler returns a handler for listing async batch jobs.
// GET /fhir/_async-batch?status=...&_count=...
func AsyncBatchListHandler(store AsyncBatchStore) echo.HandlerFunc {
	return func(c echo.Context) error {
		status := c.QueryParam("status")
		limitStr := c.QueryParam("_count")
		limit := 100
		if limitStr != "" {
			if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 {
				limit = parsed
			}
		}

		jobs, err := store.ListJobs(c.Request().Context(), status, limit)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, NewOperationOutcome(
				IssueSeverityError, IssueTypeException, err.Error(),
			))
		}

		jobSummaries := make([]map[string]interface{}, 0, len(jobs))
		for _, job := range jobs {
			summary := map[string]interface{}{
				"id":               job.ID,
				"status":           job.Status,
				"bundleType":       job.BundleType,
				"totalEntries":     job.TotalEntries,
				"processedEntries": job.ProcessedEntries,
				"successCount":     job.SuccessCount,
				"errorCount":       job.ErrorCount,
				"progress":         job.Progress,
				"createdAt":        job.CreatedAt.Format(time.RFC3339),
				"request":          job.Request,
			}
			if job.EndTime != nil {
				summary["endTime"] = job.EndTime.Format(time.RFC3339)
			}
			jobSummaries = append(jobSummaries, summary)
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"jobs":  jobSummaries,
			"total": len(jobSummaries),
		})
	}
}

// RetryWithBackoff retries an operation with exponential backoff.
// The function is called up to maxAttempts times. If the function returns nil,
// retry stops immediately. Otherwise, the last error is returned.
func RetryWithBackoff(maxAttempts int, baseDelay time.Duration, fn func() error) error {
	var lastErr error
	for attempt := 0; attempt < maxAttempts; attempt++ {
		lastErr = fn()
		if lastErr == nil {
			return nil
		}
		if attempt < maxAttempts-1 {
			delay := baseDelay * time.Duration(1<<uint(attempt))
			time.Sleep(delay)
		}
	}
	return lastErr
}

// ValidateAsyncBatchConfig validates the configuration and returns any issues.
func ValidateAsyncBatchConfig(config *AsyncBatchConfig) []ValidationIssue {
	var issues []ValidationIssue

	if config.MaxSyncEntries < 0 {
		issues = append(issues, ValidationIssue{
			Severity:    SeverityError,
			Code:        VIssueTypeValue,
			Diagnostics: "MaxSyncEntries must be non-negative",
			Location:    "AsyncBatchConfig.MaxSyncEntries",
		})
	}

	if config.MaxAsyncEntries <= 0 {
		issues = append(issues, ValidationIssue{
			Severity:    SeverityError,
			Code:        VIssueTypeValue,
			Diagnostics: "MaxAsyncEntries must be positive",
			Location:    "AsyncBatchConfig.MaxAsyncEntries",
		})
	}

	if config.WorkerCount <= 0 {
		issues = append(issues, ValidationIssue{
			Severity:    SeverityError,
			Code:        VIssueTypeValue,
			Diagnostics: "WorkerCount must be positive",
			Location:    "AsyncBatchConfig.WorkerCount",
		})
	}

	if config.RetryMaxAttempts <= 0 {
		issues = append(issues, ValidationIssue{
			Severity:    SeverityError,
			Code:        VIssueTypeValue,
			Diagnostics: "RetryMaxAttempts must be positive",
			Location:    "AsyncBatchConfig.RetryMaxAttempts",
		})
	}

	if config.RetryDelay <= 0 {
		issues = append(issues, ValidationIssue{
			Severity:    SeverityError,
			Code:        VIssueTypeValue,
			Diagnostics: "RetryDelay must be positive",
			Location:    "AsyncBatchConfig.RetryDelay",
		})
	}

	if config.MaxSyncTimeout <= 0 {
		issues = append(issues, ValidationIssue{
			Severity:    SeverityWarning,
			Code:        VIssueTypeValue,
			Diagnostics: "MaxSyncTimeout should be positive",
			Location:    "AsyncBatchConfig.MaxSyncTimeout",
		})
	}

	if config.ProgressInterval <= 0 {
		issues = append(issues, ValidationIssue{
			Severity:    SeverityWarning,
			Code:        VIssueTypeValue,
			Diagnostics: "ProgressInterval should be positive",
			Location:    "AsyncBatchConfig.ProgressInterval",
		})
	}

	return issues
}

// ProcessEntryWithRetry processes a single entry with retry logic.
func ProcessEntryWithRetry(
	entry *TransactionEntry,
	handler func(method, url string, resource map[string]interface{}) (*BundleEntryResponse, error),
	config *AsyncBatchConfig,
) (*BundleEntryResponse, error) {
	var resp *BundleEntryResponse
	err := RetryWithBackoff(config.RetryMaxAttempts, config.RetryDelay, func() error {
		var handlerErr error
		resp, handlerErr = handler(entry.Request.Method, entry.Request.URL, entry.Resource)
		return handlerErr
	})
	return resp, err
}
