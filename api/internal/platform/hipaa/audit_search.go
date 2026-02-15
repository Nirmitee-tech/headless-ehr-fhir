package hipaa

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
)

// AuditSearchParams holds filter, pagination, and sort parameters for audit trail search.
type AuditSearchParams struct {
	UserID       string     `json:"user_id" query:"user_id"`
	PatientID    string     `json:"patient_id" query:"patient_id"`
	Action       string     `json:"action" query:"action"`
	ResourceType string     `json:"resource_type" query:"resource_type"`
	StartTime    *time.Time `json:"start_time" query:"start_time"`
	EndTime      *time.Time `json:"end_time" query:"end_time"`
	Outcome      string     `json:"outcome" query:"outcome"`
	SourceIP     string     `json:"source_ip" query:"source_ip"`
	Limit        int        `json:"limit" query:"limit"`
	Offset       int        `json:"offset" query:"offset"`
	SortBy       string     `json:"sort_by" query:"sort_by"`
	SortOrder    string     `json:"sort_order" query:"sort_order"`
}

// AuditSearchResult contains paginated search results.
type AuditSearchResult struct {
	Entries []*AuditEntry `json:"entries"`
	Total   int           `json:"total"`
	Limit   int           `json:"limit"`
	Offset  int           `json:"offset"`
}

// AuditEntry represents a single audit log entry for the search/export layer.
type AuditEntry struct {
	ID           string    `json:"id"`
	Timestamp    time.Time `json:"timestamp"`
	UserID       string    `json:"user_id"`
	UserName     string    `json:"user_name"`
	PatientID    string    `json:"patient_id"`
	Action       string    `json:"action"`
	ResourceType string    `json:"resource_type"`
	ResourceID   string    `json:"resource_id"`
	Outcome      string    `json:"outcome"`
	SourceIP     string    `json:"source_ip"`
	UserAgent    string    `json:"user_agent"`
	Detail       string    `json:"detail"`
	TenantID     string    `json:"tenant_id"`
}

// AuditSummary contains aggregated statistics for audit entries.
type AuditSummary struct {
	TotalEntries   int            `json:"total_entries"`
	ByAction       map[string]int `json:"by_action"`
	ByResourceType map[string]int `json:"by_resource_type"`
	ByOutcome      map[string]int `json:"by_outcome"`
	ByUser         map[string]int `json:"by_user"`
	TimeRange      struct {
		First time.Time `json:"first"`
		Last  time.Time `json:"last"`
	} `json:"time_range"`
}

// AuditSearcher provides in-memory audit entry storage and search for dev/test use.
type AuditSearcher struct {
	mu      sync.RWMutex
	entries []*AuditEntry
}

// NewAuditSearcher creates a new empty AuditSearcher.
func NewAuditSearcher() *AuditSearcher {
	return &AuditSearcher{
		entries: make([]*AuditEntry, 0),
	}
}

// AddEntry appends a new audit entry to the in-memory store. Thread-safe.
func (s *AuditSearcher) AddEntry(entry *AuditEntry) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries = append(s.entries, entry)
}

// applyDefaults normalizes search params, applying defaults for limit, sort, etc.
func applyDefaults(params *AuditSearchParams) {
	if params.Limit <= 0 {
		params.Limit = 100
	}
	if params.Limit > 1000 {
		params.Limit = 1000
	}
	if params.Offset < 0 {
		params.Offset = 0
	}
	if params.SortBy == "" {
		params.SortBy = "timestamp"
	}
	if params.SortOrder == "" {
		params.SortOrder = "desc"
	}
}

// matchEntry returns true if the entry matches all non-zero filter criteria.
func matchEntry(entry *AuditEntry, params AuditSearchParams) bool {
	if params.UserID != "" && entry.UserID != params.UserID {
		return false
	}
	if params.PatientID != "" && entry.PatientID != params.PatientID {
		return false
	}
	if params.Action != "" && entry.Action != params.Action {
		return false
	}
	if params.ResourceType != "" && entry.ResourceType != params.ResourceType {
		return false
	}
	if params.Outcome != "" && entry.Outcome != params.Outcome {
		return false
	}
	if params.SourceIP != "" && entry.SourceIP != params.SourceIP {
		return false
	}
	if params.StartTime != nil && entry.Timestamp.Before(*params.StartTime) {
		return false
	}
	if params.EndTime != nil && entry.Timestamp.After(*params.EndTime) {
		return false
	}
	return true
}

// filterEntries returns a slice of entries matching the given params (no copy of pointers needed for read-only use).
func (s *AuditSearcher) filterEntries(params AuditSearchParams) []*AuditEntry {
	var filtered []*AuditEntry
	for _, e := range s.entries {
		if matchEntry(e, params) {
			filtered = append(filtered, e)
		}
	}
	return filtered
}

// sortEntries sorts entries in place by the given sort parameters.
func sortEntries(entries []*AuditEntry, sortBy, sortOrder string) {
	sort.SliceStable(entries, func(i, j int) bool {
		var less bool
		switch sortBy {
		case "user":
			less = entries[i].UserID < entries[j].UserID
		case "action":
			less = entries[i].Action < entries[j].Action
		default: // "timestamp"
			less = entries[i].Timestamp.Before(entries[j].Timestamp)
		}
		if sortOrder == "desc" {
			return !less
		}
		return less
	})
}

// Search filters, sorts, and paginates audit entries.
func (s *AuditSearcher) Search(_ context.Context, params AuditSearchParams) (*AuditSearchResult, error) {
	applyDefaults(&params)

	s.mu.RLock()
	filtered := s.filterEntries(params)
	s.mu.RUnlock()

	sortEntries(filtered, params.SortBy, params.SortOrder)

	total := len(filtered)

	// Paginate
	start := params.Offset
	if start > total {
		start = total
	}
	end := start + params.Limit
	if end > total {
		end = total
	}

	page := filtered[start:end]

	return &AuditSearchResult{
		Entries: page,
		Total:   total,
		Limit:   params.Limit,
		Offset:  params.Offset,
	}, nil
}

// ExportCSV writes matching audit entries as CSV to the provided writer.
func (s *AuditSearcher) ExportCSV(_ context.Context, params AuditSearchParams, w io.Writer) error {
	// For export, get all results (no pagination limit)
	s.mu.RLock()
	filtered := s.filterEntries(params)
	s.mu.RUnlock()

	sortEntries(filtered, params.SortBy, params.SortOrder)

	cw := csv.NewWriter(w)
	defer cw.Flush()

	// Write header
	header := []string{"ID", "Timestamp", "UserID", "UserName", "PatientID",
		"Action", "ResourceType", "ResourceID", "Outcome", "SourceIP", "UserAgent", "Detail", "TenantID"}
	if err := cw.Write(header); err != nil {
		return fmt.Errorf("audit export csv: write header: %w", err)
	}

	for _, e := range filtered {
		record := []string{
			e.ID,
			e.Timestamp.Format(time.RFC3339),
			e.UserID,
			e.UserName,
			e.PatientID,
			e.Action,
			e.ResourceType,
			e.ResourceID,
			e.Outcome,
			e.SourceIP,
			e.UserAgent,
			e.Detail,
			e.TenantID,
		}
		if err := cw.Write(record); err != nil {
			return fmt.Errorf("audit export csv: write record: %w", err)
		}
	}

	return nil
}

// ExportJSON writes matching audit entries as a JSON array to the provided writer.
func (s *AuditSearcher) ExportJSON(_ context.Context, params AuditSearchParams, w io.Writer) error {
	s.mu.RLock()
	filtered := s.filterEntries(params)
	s.mu.RUnlock()

	sortEntries(filtered, params.SortBy, params.SortOrder)

	// Ensure empty slice serializes as [] not null
	if filtered == nil {
		filtered = make([]*AuditEntry, 0)
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(filtered); err != nil {
		return fmt.Errorf("audit export json: %w", err)
	}
	return nil
}

// Summary computes aggregate statistics for matching entries.
func (s *AuditSearcher) Summary(_ context.Context, params AuditSearchParams) (*AuditSummary, error) {
	s.mu.RLock()
	filtered := s.filterEntries(params)
	s.mu.RUnlock()

	summary := &AuditSummary{
		TotalEntries:   len(filtered),
		ByAction:       make(map[string]int),
		ByResourceType: make(map[string]int),
		ByOutcome:      make(map[string]int),
		ByUser:         make(map[string]int),
	}

	for i, e := range filtered {
		summary.ByAction[e.Action]++
		summary.ByResourceType[e.ResourceType]++
		summary.ByOutcome[e.Outcome]++
		summary.ByUser[e.UserID]++

		if i == 0 {
			summary.TimeRange.First = e.Timestamp
			summary.TimeRange.Last = e.Timestamp
		} else {
			if e.Timestamp.Before(summary.TimeRange.First) {
				summary.TimeRange.First = e.Timestamp
			}
			if e.Timestamp.After(summary.TimeRange.Last) {
				summary.TimeRange.Last = e.Timestamp
			}
		}
	}

	return summary, nil
}

// GetEntry returns a single audit entry by ID, or nil if not found.
func (s *AuditSearcher) GetEntry(id string) *AuditEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, e := range s.entries {
		if e.ID == id {
			return e
		}
	}
	return nil
}

// ---------- HTTP Handler ----------

// AuditSearchHandler provides Echo HTTP handlers for audit trail search and export.
type AuditSearchHandler struct {
	searcher *AuditSearcher
}

// NewAuditSearchHandler creates a new handler backed by the given searcher.
func NewAuditSearchHandler(searcher *AuditSearcher) *AuditSearchHandler {
	return &AuditSearchHandler{searcher: searcher}
}

// RegisterRoutes registers all audit search routes on the provided Echo group.
func (h *AuditSearchHandler) RegisterRoutes(g *echo.Group) {
	g.GET("/audit/search", h.HandleSearch)
	g.GET("/audit/export/csv", h.HandleExportCSV)
	g.GET("/audit/export/json", h.HandleExportJSON)
	g.GET("/audit/summary", h.HandleSummary)
	g.GET("/audit/:id", h.HandleGetEntry)
}

// parseSearchParams extracts AuditSearchParams from Echo query parameters.
func parseSearchParams(c echo.Context) AuditSearchParams {
	params := AuditSearchParams{
		UserID:       c.QueryParam("user_id"),
		PatientID:    c.QueryParam("patient_id"),
		Action:       c.QueryParam("action"),
		ResourceType: c.QueryParam("resource_type"),
		Outcome:      c.QueryParam("outcome"),
		SourceIP:     c.QueryParam("source_ip"),
		SortBy:       c.QueryParam("sort_by"),
		SortOrder:    c.QueryParam("sort_order"),
	}

	if v := c.QueryParam("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			params.Limit = n
		}
	}
	if v := c.QueryParam("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			params.Offset = n
		}
	}
	if v := c.QueryParam("start_time"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			params.StartTime = &t
		}
	}
	if v := c.QueryParam("end_time"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			params.EndTime = &t
		}
	}

	return params
}

// HandleSearch handles GET /audit/search.
func (h *AuditSearchHandler) HandleSearch(c echo.Context) error {
	params := parseSearchParams(c)
	result, err := h.searcher.Search(c.Request().Context(), params)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, result)
}

// HandleExportCSV handles GET /audit/export/csv.
func (h *AuditSearchHandler) HandleExportCSV(c echo.Context) error {
	params := parseSearchParams(c)

	c.Response().Header().Set("Content-Type", "text/csv")
	c.Response().Header().Set("Content-Disposition",
		fmt.Sprintf("attachment; filename=\"audit_export_%s.csv\"", time.Now().UTC().Format("20060102_150405")))
	c.Response().WriteHeader(http.StatusOK)

	if err := h.searcher.ExportCSV(c.Request().Context(), params, c.Response()); err != nil {
		return err
	}
	return nil
}

// HandleExportJSON handles GET /audit/export/json.
func (h *AuditSearchHandler) HandleExportJSON(c echo.Context) error {
	params := parseSearchParams(c)

	c.Response().Header().Set("Content-Type", "application/json")
	c.Response().Header().Set("Content-Disposition",
		fmt.Sprintf("attachment; filename=\"audit_export_%s.json\"", time.Now().UTC().Format("20060102_150405")))
	c.Response().WriteHeader(http.StatusOK)

	if err := h.searcher.ExportJSON(c.Request().Context(), params, c.Response()); err != nil {
		return err
	}
	return nil
}

// HandleSummary handles GET /audit/summary.
func (h *AuditSearchHandler) HandleSummary(c echo.Context) error {
	params := parseSearchParams(c)
	summary, err := h.searcher.Summary(c.Request().Context(), params)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, summary)
}

// HandleGetEntry handles GET /audit/:id.
func (h *AuditSearchHandler) HandleGetEntry(c echo.Context) error {
	id := c.Param("id")
	entry := h.searcher.GetEntry(id)
	if entry == nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "entry not found"})
	}
	return c.JSON(http.StatusOK, entry)
}
