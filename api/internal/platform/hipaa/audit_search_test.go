package hipaa

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
)

// helper to create test entries
func makeTestEntries() []*AuditEntry {
	now := time.Now().UTC()
	return []*AuditEntry{
		{
			ID: "entry-1", Timestamp: now.Add(-5 * time.Hour),
			UserID: "user-1", UserName: "Dr. Smith", PatientID: "patient-1",
			Action: "read", ResourceType: "Patient", ResourceID: "res-1",
			Outcome: "success", SourceIP: "10.0.0.1", UserAgent: "Mozilla/5.0",
			Detail: "Read patient record", TenantID: "tenant-1",
		},
		{
			ID: "entry-2", Timestamp: now.Add(-4 * time.Hour),
			UserID: "user-2", UserName: "Nurse Johnson", PatientID: "patient-2",
			Action: "create", ResourceType: "Observation", ResourceID: "res-2",
			Outcome: "success", SourceIP: "10.0.0.2", UserAgent: "Chrome/120",
			Detail: "Created observation", TenantID: "tenant-1",
		},
		{
			ID: "entry-3", Timestamp: now.Add(-3 * time.Hour),
			UserID: "user-1", UserName: "Dr. Smith", PatientID: "patient-1",
			Action: "update", ResourceType: "Patient", ResourceID: "res-1",
			Outcome: "success", SourceIP: "10.0.0.1", UserAgent: "Mozilla/5.0",
			Detail: "Updated patient demographics", TenantID: "tenant-1",
		},
		{
			ID: "entry-4", Timestamp: now.Add(-2 * time.Hour),
			UserID: "user-3", UserName: "Admin User", PatientID: "patient-3",
			Action: "delete", ResourceType: "DocumentReference", ResourceID: "res-3",
			Outcome: "failure", SourceIP: "10.0.0.3", UserAgent: "Safari/17",
			Detail: "Delete denied", TenantID: "tenant-2",
		},
		{
			ID: "entry-5", Timestamp: now.Add(-1 * time.Hour),
			UserID: "user-2", UserName: "Nurse Johnson", PatientID: "patient-1",
			Action: "search", ResourceType: "MedicationRequest", ResourceID: "",
			Outcome: "success", SourceIP: "10.0.0.2", UserAgent: "Chrome/120",
			Detail: "Searched medications", TenantID: "tenant-1",
		},
	}
}

func newPopulatedSearcher() *AuditSearcher {
	s := NewAuditSearcher()
	for _, e := range makeTestEntries() {
		s.AddEntry(e)
	}
	return s
}

// --- Search tests ---

func TestAuditSearcher_SearchByUserID(t *testing.T) {
	s := newPopulatedSearcher()
	result, err := s.Search(context.Background(), AuditSearchParams{UserID: "user-1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Total != 2 {
		t.Errorf("expected 2 entries for user-1, got %d", result.Total)
	}
	for _, e := range result.Entries {
		if e.UserID != "user-1" {
			t.Errorf("expected UserID user-1, got %s", e.UserID)
		}
	}
}

func TestAuditSearcher_SearchByPatientID(t *testing.T) {
	s := newPopulatedSearcher()
	result, err := s.Search(context.Background(), AuditSearchParams{PatientID: "patient-1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Total != 3 {
		t.Errorf("expected 3 entries for patient-1, got %d", result.Total)
	}
	for _, e := range result.Entries {
		if e.PatientID != "patient-1" {
			t.Errorf("expected PatientID patient-1, got %s", e.PatientID)
		}
	}
}

func TestAuditSearcher_SearchByAction(t *testing.T) {
	s := newPopulatedSearcher()
	result, err := s.Search(context.Background(), AuditSearchParams{Action: "read"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Total != 1 {
		t.Errorf("expected 1 read entry, got %d", result.Total)
	}
	if result.Entries[0].Action != "read" {
		t.Errorf("expected action read, got %s", result.Entries[0].Action)
	}
}

func TestAuditSearcher_SearchByResourceType(t *testing.T) {
	s := newPopulatedSearcher()
	result, err := s.Search(context.Background(), AuditSearchParams{ResourceType: "Patient"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Total != 2 {
		t.Errorf("expected 2 Patient entries, got %d", result.Total)
	}
	for _, e := range result.Entries {
		if e.ResourceType != "Patient" {
			t.Errorf("expected ResourceType Patient, got %s", e.ResourceType)
		}
	}
}

func TestAuditSearcher_SearchByTimeRange(t *testing.T) {
	s := newPopulatedSearcher()
	now := time.Now().UTC()
	start := now.Add(-4*time.Hour - 30*time.Minute)
	end := now.Add(-1*time.Hour - 30*time.Minute)
	result, err := s.Search(context.Background(), AuditSearchParams{
		StartTime: &start,
		EndTime:   &end,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// entries 2,3,4 should match (-4h, -3h, -2h are within -4.5h to -1.5h)
	if result.Total != 3 {
		t.Errorf("expected 3 entries in time range, got %d", result.Total)
	}
}

func TestAuditSearcher_SearchByOutcome(t *testing.T) {
	s := newPopulatedSearcher()
	result, err := s.Search(context.Background(), AuditSearchParams{Outcome: "failure"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Total != 1 {
		t.Errorf("expected 1 failure entry, got %d", result.Total)
	}
	if result.Entries[0].Outcome != "failure" {
		t.Errorf("expected outcome failure, got %s", result.Entries[0].Outcome)
	}
}

func TestAuditSearcher_SearchBySourceIP(t *testing.T) {
	s := newPopulatedSearcher()
	result, err := s.Search(context.Background(), AuditSearchParams{SourceIP: "10.0.0.2"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Total != 2 {
		t.Errorf("expected 2 entries from 10.0.0.2, got %d", result.Total)
	}
	for _, e := range result.Entries {
		if e.SourceIP != "10.0.0.2" {
			t.Errorf("expected SourceIP 10.0.0.2, got %s", e.SourceIP)
		}
	}
}

func TestAuditSearcher_SearchCombinedFilters(t *testing.T) {
	s := newPopulatedSearcher()
	result, err := s.Search(context.Background(), AuditSearchParams{
		UserID:       "user-1",
		ResourceType: "Patient",
		Outcome:      "success",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Total != 2 {
		t.Errorf("expected 2 entries with combined filters, got %d", result.Total)
	}
	for _, e := range result.Entries {
		if e.UserID != "user-1" || e.ResourceType != "Patient" || e.Outcome != "success" {
			t.Errorf("entry did not match all filters: %+v", e)
		}
	}
}

func TestAuditSearcher_SearchPagination(t *testing.T) {
	s := newPopulatedSearcher()

	// Get first page (limit 2)
	result1, err := s.Search(context.Background(), AuditSearchParams{Limit: 2, Offset: 0})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result1.Entries) != 2 {
		t.Errorf("expected 2 entries on page 1, got %d", len(result1.Entries))
	}
	if result1.Total != 5 {
		t.Errorf("expected total 5, got %d", result1.Total)
	}
	if result1.Limit != 2 {
		t.Errorf("expected limit 2, got %d", result1.Limit)
	}
	if result1.Offset != 0 {
		t.Errorf("expected offset 0, got %d", result1.Offset)
	}

	// Get second page
	result2, err := s.Search(context.Background(), AuditSearchParams{Limit: 2, Offset: 2})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result2.Entries) != 2 {
		t.Errorf("expected 2 entries on page 2, got %d", len(result2.Entries))
	}

	// Get third page (only 1 left)
	result3, err := s.Search(context.Background(), AuditSearchParams{Limit: 2, Offset: 4})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result3.Entries) != 1 {
		t.Errorf("expected 1 entry on page 3, got %d", len(result3.Entries))
	}

	// Ensure no overlap
	ids := make(map[string]bool)
	for _, e := range result1.Entries {
		ids[e.ID] = true
	}
	for _, e := range result2.Entries {
		if ids[e.ID] {
			t.Errorf("duplicate entry %s across pages", e.ID)
		}
		ids[e.ID] = true
	}
	for _, e := range result3.Entries {
		if ids[e.ID] {
			t.Errorf("duplicate entry %s across pages", e.ID)
		}
	}
}

func TestAuditSearcher_SearchSortByTimestamp(t *testing.T) {
	s := newPopulatedSearcher()
	// Default sort: timestamp desc
	result, err := s.Search(context.Background(), AuditSearchParams{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Entries) < 2 {
		t.Fatal("need at least 2 entries")
	}
	for i := 1; i < len(result.Entries); i++ {
		if result.Entries[i].Timestamp.After(result.Entries[i-1].Timestamp) {
			t.Errorf("entries not sorted desc by timestamp: entry[%d]=%v after entry[%d]=%v",
				i, result.Entries[i].Timestamp, i-1, result.Entries[i-1].Timestamp)
		}
	}
}

func TestAuditSearcher_SearchSortAscending(t *testing.T) {
	s := newPopulatedSearcher()
	result, err := s.Search(context.Background(), AuditSearchParams{
		SortBy:    "timestamp",
		SortOrder: "asc",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Entries) < 2 {
		t.Fatal("need at least 2 entries")
	}
	for i := 1; i < len(result.Entries); i++ {
		if result.Entries[i].Timestamp.Before(result.Entries[i-1].Timestamp) {
			t.Errorf("entries not sorted asc by timestamp: entry[%d]=%v before entry[%d]=%v",
				i, result.Entries[i].Timestamp, i-1, result.Entries[i-1].Timestamp)
		}
	}
}

func TestAuditSearcher_SearchEmptyResult(t *testing.T) {
	s := newPopulatedSearcher()
	result, err := s.Search(context.Background(), AuditSearchParams{UserID: "nonexistent"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Total != 0 {
		t.Errorf("expected 0 total, got %d", result.Total)
	}
	if len(result.Entries) != 0 {
		t.Errorf("expected 0 entries, got %d", len(result.Entries))
	}
}

func TestAuditSearcher_ConcurrentAccess(t *testing.T) {
	s := NewAuditSearcher()
	var wg sync.WaitGroup

	// Concurrent writes
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			s.AddEntry(&AuditEntry{
				ID:        fmt.Sprintf("concurrent-%d", idx),
				Timestamp: time.Now().UTC(),
				UserID:    fmt.Sprintf("user-%d", idx%5),
				Action:    "read",
				Outcome:   "success",
			})
		}(i)
	}

	// Concurrent reads
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = s.Search(context.Background(), AuditSearchParams{})
		}()
	}

	wg.Wait()

	result, err := s.Search(context.Background(), AuditSearchParams{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Total != 100 {
		t.Errorf("expected 100 entries after concurrent writes, got %d", result.Total)
	}
}

// --- Export tests ---

func TestAuditSearcher_ExportCSV(t *testing.T) {
	s := newPopulatedSearcher()
	var buf bytes.Buffer
	err := s.ExportCSV(context.Background(), AuditSearchParams{}, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	reader := csv.NewReader(strings.NewReader(buf.String()))
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("invalid CSV: %v", err)
	}

	// Header + 5 data rows
	if len(records) != 6 {
		t.Errorf("expected 6 CSV rows (1 header + 5 data), got %d", len(records))
	}

	// Check header
	expectedHeaders := []string{"ID", "Timestamp", "UserID", "UserName", "PatientID",
		"Action", "ResourceType", "ResourceID", "Outcome", "SourceIP", "UserAgent", "Detail", "TenantID"}
	if len(records) > 0 {
		for i, h := range expectedHeaders {
			if i >= len(records[0]) || records[0][i] != h {
				t.Errorf("expected header[%d]=%s, got %s", i, h, records[0][i])
			}
		}
	}
}

func TestAuditSearcher_ExportCSV_EmptyResult(t *testing.T) {
	s := newPopulatedSearcher()
	var buf bytes.Buffer
	err := s.ExportCSV(context.Background(), AuditSearchParams{UserID: "nonexistent"}, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	reader := csv.NewReader(strings.NewReader(buf.String()))
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("invalid CSV: %v", err)
	}

	// Header only
	if len(records) != 1 {
		t.Errorf("expected 1 CSV row (header only), got %d", len(records))
	}
}

func TestAuditSearcher_ExportJSON(t *testing.T) {
	s := newPopulatedSearcher()
	var buf bytes.Buffer
	err := s.ExportJSON(context.Background(), AuditSearchParams{}, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var entries []*AuditEntry
	if err := json.Unmarshal(buf.Bytes(), &entries); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(entries) != 5 {
		t.Errorf("expected 5 JSON entries, got %d", len(entries))
	}
}

func TestAuditSearcher_ExportJSON_EmptyResult(t *testing.T) {
	s := newPopulatedSearcher()
	var buf bytes.Buffer
	err := s.ExportJSON(context.Background(), AuditSearchParams{UserID: "nonexistent"}, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var entries []*AuditEntry
	if err := json.Unmarshal(buf.Bytes(), &entries); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected 0 JSON entries, got %d", len(entries))
	}
}

// --- Summary tests ---

func TestAuditSearcher_Summary(t *testing.T) {
	s := newPopulatedSearcher()
	summary, err := s.Summary(context.Background(), AuditSearchParams{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary.TotalEntries != 5 {
		t.Errorf("expected 5 total entries, got %d", summary.TotalEntries)
	}
	if len(summary.ByAction) == 0 {
		t.Error("expected non-empty ByAction map")
	}
	if len(summary.ByResourceType) == 0 {
		t.Error("expected non-empty ByResourceType map")
	}
	if len(summary.ByOutcome) == 0 {
		t.Error("expected non-empty ByOutcome map")
	}
	if len(summary.ByUser) == 0 {
		t.Error("expected non-empty ByUser map")
	}
}

func TestAuditSearcher_SummaryByAction(t *testing.T) {
	s := newPopulatedSearcher()
	summary, err := s.Summary(context.Background(), AuditSearchParams{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary.ByAction["read"] != 1 {
		t.Errorf("expected 1 read action, got %d", summary.ByAction["read"])
	}
	if summary.ByAction["create"] != 1 {
		t.Errorf("expected 1 create action, got %d", summary.ByAction["create"])
	}
	if summary.ByAction["update"] != 1 {
		t.Errorf("expected 1 update action, got %d", summary.ByAction["update"])
	}
	if summary.ByAction["delete"] != 1 {
		t.Errorf("expected 1 delete action, got %d", summary.ByAction["delete"])
	}
	if summary.ByAction["search"] != 1 {
		t.Errorf("expected 1 search action, got %d", summary.ByAction["search"])
	}
}

func TestAuditSearcher_SummaryTimeRange(t *testing.T) {
	s := newPopulatedSearcher()
	summary, err := s.Summary(context.Background(), AuditSearchParams{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary.TimeRange.First.IsZero() {
		t.Error("expected non-zero First time")
	}
	if summary.TimeRange.Last.IsZero() {
		t.Error("expected non-zero Last time")
	}
	if !summary.TimeRange.First.Before(summary.TimeRange.Last) {
		t.Errorf("expected First (%v) before Last (%v)", summary.TimeRange.First, summary.TimeRange.Last)
	}
}

// --- Handler tests ---

func TestAuditSearchHandler_Search(t *testing.T) {
	s := newPopulatedSearcher()
	h := NewAuditSearchHandler(s)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/audit/search?user_id=user-1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := h.HandleSearch(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	var result AuditSearchResult
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON response: %v", err)
	}
	if result.Total != 2 {
		t.Errorf("expected total 2, got %d", result.Total)
	}
	if len(result.Entries) != 2 {
		t.Errorf("expected 2 entries, got %d", len(result.Entries))
	}
}

func TestAuditSearchHandler_ExportCSV(t *testing.T) {
	s := newPopulatedSearcher()
	h := NewAuditSearchHandler(s)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/audit/export/csv", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := h.HandleExportCSV(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	contentType := rec.Header().Get("Content-Type")
	if contentType != "text/csv" {
		t.Errorf("expected Content-Type text/csv, got %s", contentType)
	}

	disposition := rec.Header().Get("Content-Disposition")
	if !strings.Contains(disposition, "attachment") {
		t.Errorf("expected Content-Disposition with attachment, got %s", disposition)
	}
	if !strings.Contains(disposition, ".csv") {
		t.Errorf("expected Content-Disposition with .csv filename, got %s", disposition)
	}

	// Validate CSV content
	reader := csv.NewReader(strings.NewReader(rec.Body.String()))
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("invalid CSV response: %v", err)
	}
	if len(records) != 6 { // header + 5 data rows
		t.Errorf("expected 6 CSV rows, got %d", len(records))
	}
}

func TestAuditSearchHandler_ExportJSON(t *testing.T) {
	s := newPopulatedSearcher()
	h := NewAuditSearchHandler(s)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/audit/export/json", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := h.HandleExportJSON(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	contentType := rec.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("expected Content-Type application/json, got %s", contentType)
	}

	disposition := rec.Header().Get("Content-Disposition")
	if !strings.Contains(disposition, "attachment") {
		t.Errorf("expected Content-Disposition with attachment, got %s", disposition)
	}

	var entries []*AuditEntry
	if err := json.Unmarshal(rec.Body.Bytes(), &entries); err != nil {
		t.Fatalf("invalid JSON response: %v", err)
	}
	if len(entries) != 5 {
		t.Errorf("expected 5 entries, got %d", len(entries))
	}
}

func TestAuditSearchHandler_Summary(t *testing.T) {
	s := newPopulatedSearcher()
	h := NewAuditSearchHandler(s)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/audit/summary", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := h.HandleSummary(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	var summary AuditSummary
	if err := json.Unmarshal(rec.Body.Bytes(), &summary); err != nil {
		t.Fatalf("invalid JSON response: %v", err)
	}
	if summary.TotalEntries != 5 {
		t.Errorf("expected 5 total entries, got %d", summary.TotalEntries)
	}
}

func TestAuditSearchHandler_GetEntry(t *testing.T) {
	s := newPopulatedSearcher()
	h := NewAuditSearchHandler(s)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/audit/entry-1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("entry-1")

	if err := h.HandleGetEntry(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	var entry AuditEntry
	if err := json.Unmarshal(rec.Body.Bytes(), &entry); err != nil {
		t.Fatalf("invalid JSON response: %v", err)
	}
	if entry.ID != "entry-1" {
		t.Errorf("expected ID entry-1, got %s", entry.ID)
	}
	if entry.UserName != "Dr. Smith" {
		t.Errorf("expected UserName 'Dr. Smith', got %s", entry.UserName)
	}
}
