package analytics

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
)

// ---------------------------------------------------------------------------
// Recording
// ---------------------------------------------------------------------------

func TestUsageTracker_Record(t *testing.T) {
	tracker := NewUsageTracker(1000)
	m := &RequestMetric{
		Timestamp:    time.Now(),
		Method:       "GET",
		Path:         "/fhir/Patient",
		StatusCode:   200,
		Duration:     50 * time.Millisecond,
		ClientID:     "client-1",
		TenantID:     "tenant-1",
		ResourceType: "Patient",
		RequestSize:  128,
		ResponseSize: 4096,
	}
	tracker.Record(m)

	overview := tracker.GetOverview()
	if overview.TotalRequests != 1 {
		t.Fatalf("expected TotalRequests=1, got %d", overview.TotalRequests)
	}
	if overview.TotalErrors != 0 {
		t.Fatalf("expected TotalErrors=0, got %d", overview.TotalErrors)
	}
}

func TestUsageTracker_Record_MaxMetrics(t *testing.T) {
	maxMetrics := 100
	tracker := NewUsageTracker(maxMetrics)

	for i := 0; i < 250; i++ {
		tracker.Record(&RequestMetric{
			Timestamp:  time.Now(),
			Method:     "GET",
			Path:       fmt.Sprintf("/fhir/Patient/%d", i),
			StatusCode: 200,
			Duration:   time.Millisecond,
			ClientID:   "client-1",
		})
	}

	tracker.mu.RLock()
	count := len(tracker.metrics)
	tracker.mu.RUnlock()

	if count != maxMetrics {
		t.Fatalf("expected ring buffer to cap at %d, got %d", maxMetrics, count)
	}

	overview := tracker.GetOverview()
	if overview.TotalRequests != 250 {
		t.Fatalf("expected TotalRequests=250, got %d", overview.TotalRequests)
	}
}

func TestUsageTracker_Record_ConcurrentAccess(t *testing.T) {
	tracker := NewUsageTracker(100000)
	var wg sync.WaitGroup
	goroutines := 100
	perGoroutine := 100

	for g := 0; g < goroutines; g++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for i := 0; i < perGoroutine; i++ {
				tracker.Record(&RequestMetric{
					Timestamp:    time.Now(),
					Method:       "GET",
					Path:         "/fhir/Patient",
					StatusCode:   200,
					Duration:     time.Millisecond,
					ClientID:     fmt.Sprintf("client-%d", id),
					ResourceType: "Patient",
				})
			}
		}(g)
	}
	wg.Wait()

	overview := tracker.GetOverview()
	expected := int64(goroutines * perGoroutine)
	if overview.TotalRequests != expected {
		t.Fatalf("expected TotalRequests=%d, got %d", expected, overview.TotalRequests)
	}
}

// ---------------------------------------------------------------------------
// Endpoint stats
// ---------------------------------------------------------------------------

func TestUsageTracker_GetEndpointStats(t *testing.T) {
	tracker := NewUsageTracker(1000)
	for i := 0; i < 10; i++ {
		tracker.Record(&RequestMetric{
			Timestamp:  time.Now(),
			Method:     "GET",
			Path:       "/fhir/Patient",
			StatusCode: 200,
			Duration:   10 * time.Millisecond,
		})
	}

	summary := tracker.GetEndpointStats("/fhir/Patient")
	if summary == nil {
		t.Fatal("expected endpoint stats, got nil")
	}
	if summary.TotalRequests != 10 {
		t.Fatalf("expected TotalRequests=10, got %d", summary.TotalRequests)
	}
	if summary.AvgLatency != 10*time.Millisecond {
		t.Fatalf("expected AvgLatency=10ms, got %v", summary.AvgLatency)
	}
}

func TestUsageTracker_GetEndpointStats_NotFound(t *testing.T) {
	tracker := NewUsageTracker(1000)
	summary := tracker.GetEndpointStats("/nonexistent")
	if summary != nil {
		t.Fatalf("expected nil for unknown path, got %+v", summary)
	}
}

func TestUsageTracker_GetTopEndpoints(t *testing.T) {
	tracker := NewUsageTracker(1000)

	for i := 0; i < 5; i++ {
		tracker.Record(&RequestMetric{
			Timestamp: time.Now(), Method: "GET", Path: "/fhir/Patient",
			StatusCode: 200, Duration: time.Millisecond,
		})
	}
	for i := 0; i < 10; i++ {
		tracker.Record(&RequestMetric{
			Timestamp: time.Now(), Method: "GET", Path: "/fhir/Observation",
			StatusCode: 200, Duration: time.Millisecond,
		})
	}
	for i := 0; i < 3; i++ {
		tracker.Record(&RequestMetric{
			Timestamp: time.Now(), Method: "POST", Path: "/fhir/Encounter",
			StatusCode: 201, Duration: time.Millisecond,
		})
	}

	top := tracker.GetTopEndpoints(2)
	if len(top) != 2 {
		t.Fatalf("expected 2 endpoints, got %d", len(top))
	}
	if top[0].Path != "/fhir/Observation" {
		t.Fatalf("expected top endpoint /fhir/Observation, got %s", top[0].Path)
	}
	if top[0].TotalRequests != 10 {
		t.Fatalf("expected 10, got %d", top[0].TotalRequests)
	}
	if top[1].Path != "/fhir/Patient" {
		t.Fatalf("expected second endpoint /fhir/Patient, got %s", top[1].Path)
	}
}

func TestUsageTracker_GetEndpointStats_ErrorRate(t *testing.T) {
	tracker := NewUsageTracker(1000)
	for i := 0; i < 8; i++ {
		tracker.Record(&RequestMetric{
			Timestamp: time.Now(), Method: "GET", Path: "/fhir/Patient",
			StatusCode: 200, Duration: time.Millisecond,
		})
	}
	for i := 0; i < 2; i++ {
		tracker.Record(&RequestMetric{
			Timestamp: time.Now(), Method: "GET", Path: "/fhir/Patient",
			StatusCode: 500, Duration: time.Millisecond,
		})
	}

	summary := tracker.GetEndpointStats("/fhir/Patient")
	if summary == nil {
		t.Fatal("expected endpoint stats, got nil")
	}
	// 2 errors out of 10 = 0.2
	if summary.ErrorRate < 0.19 || summary.ErrorRate > 0.21 {
		t.Fatalf("expected ErrorRate ~0.2, got %f", summary.ErrorRate)
	}
}

// ---------------------------------------------------------------------------
// Client stats
// ---------------------------------------------------------------------------

func TestUsageTracker_GetClientStats(t *testing.T) {
	tracker := NewUsageTracker(1000)
	for i := 0; i < 5; i++ {
		tracker.Record(&RequestMetric{
			Timestamp: time.Now(), Method: "GET", Path: "/fhir/Patient",
			StatusCode: 200, Duration: time.Millisecond, ClientID: "app-1",
		})
	}

	summary := tracker.GetClientStats("app-1")
	if summary == nil {
		t.Fatal("expected client stats, got nil")
	}
	if summary.TotalRequests != 5 {
		t.Fatalf("expected 5 requests, got %d", summary.TotalRequests)
	}
	if summary.ClientID != "app-1" {
		t.Fatalf("expected ClientID=app-1, got %s", summary.ClientID)
	}
}

func TestUsageTracker_GetTopClients(t *testing.T) {
	tracker := NewUsageTracker(1000)
	for i := 0; i < 10; i++ {
		tracker.Record(&RequestMetric{
			Timestamp: time.Now(), Method: "GET", Path: "/fhir/Patient",
			StatusCode: 200, Duration: time.Millisecond, ClientID: "heavy-client",
		})
	}
	for i := 0; i < 3; i++ {
		tracker.Record(&RequestMetric{
			Timestamp: time.Now(), Method: "GET", Path: "/fhir/Patient",
			StatusCode: 200, Duration: time.Millisecond, ClientID: "light-client",
		})
	}

	top := tracker.GetTopClients(2)
	if len(top) != 2 {
		t.Fatalf("expected 2 clients, got %d", len(top))
	}
	if top[0].ClientID != "heavy-client" {
		t.Fatalf("expected heavy-client first, got %s", top[0].ClientID)
	}
	if top[0].TotalRequests != 10 {
		t.Fatalf("expected 10, got %d", top[0].TotalRequests)
	}
}

func TestUsageTracker_GetClientStats_ByteTracking(t *testing.T) {
	tracker := NewUsageTracker(1000)
	tracker.Record(&RequestMetric{
		Timestamp: time.Now(), Method: "POST", Path: "/fhir/Patient",
		StatusCode: 201, Duration: time.Millisecond, ClientID: "app-1",
		RequestSize: 512, ResponseSize: 1024,
	})
	tracker.Record(&RequestMetric{
		Timestamp: time.Now(), Method: "GET", Path: "/fhir/Patient",
		StatusCode: 200, Duration: time.Millisecond, ClientID: "app-1",
		RequestSize: 64, ResponseSize: 2048,
	})

	summary := tracker.GetClientStats("app-1")
	if summary == nil {
		t.Fatal("expected client stats, got nil")
	}
	if summary.BytesSent != 576 {
		t.Fatalf("expected BytesSent=576, got %d", summary.BytesSent)
	}
	if summary.BytesReceived != 3072 {
		t.Fatalf("expected BytesReceived=3072, got %d", summary.BytesReceived)
	}
}

// ---------------------------------------------------------------------------
// Resource stats
// ---------------------------------------------------------------------------

func TestUsageTracker_GetResourceStats(t *testing.T) {
	tracker := NewUsageTracker(1000)

	// CREATE
	tracker.Record(&RequestMetric{
		Timestamp: time.Now(), Method: "POST", Path: "/fhir/Patient",
		StatusCode: 201, Duration: time.Millisecond, ResourceType: "Patient",
	})
	// READ
	tracker.Record(&RequestMetric{
		Timestamp: time.Now(), Method: "GET", Path: "/fhir/Patient/123",
		StatusCode: 200, Duration: time.Millisecond, ResourceType: "Patient",
	})
	// UPDATE
	tracker.Record(&RequestMetric{
		Timestamp: time.Now(), Method: "PUT", Path: "/fhir/Patient/123",
		StatusCode: 200, Duration: time.Millisecond, ResourceType: "Patient",
	})
	// DELETE
	tracker.Record(&RequestMetric{
		Timestamp: time.Now(), Method: "DELETE", Path: "/fhir/Patient/123",
		StatusCode: 204, Duration: time.Millisecond, ResourceType: "Patient",
	})

	summary := tracker.GetResourceStats("Patient")
	if summary == nil {
		t.Fatal("expected resource stats, got nil")
	}
	if summary.CreateCount != 1 {
		t.Fatalf("expected CreateCount=1, got %d", summary.CreateCount)
	}
	if summary.ReadCount != 1 {
		t.Fatalf("expected ReadCount=1, got %d", summary.ReadCount)
	}
	if summary.UpdateCount != 1 {
		t.Fatalf("expected UpdateCount=1, got %d", summary.UpdateCount)
	}
	if summary.DeleteCount != 1 {
		t.Fatalf("expected DeleteCount=1, got %d", summary.DeleteCount)
	}
	if summary.Total != 4 {
		t.Fatalf("expected Total=4, got %d", summary.Total)
	}
}

func TestUsageTracker_GetResourceStats_ReadVsSearch(t *testing.T) {
	tracker := NewUsageTracker(1000)

	// READ by ID
	tracker.Record(&RequestMetric{
		Timestamp: time.Now(), Method: "GET", Path: "/fhir/Patient/123",
		StatusCode: 200, Duration: time.Millisecond, ResourceType: "Patient",
	})
	// SEARCH (no ID, treated as search)
	tracker.Record(&RequestMetric{
		Timestamp: time.Now(), Method: "GET", Path: "/fhir/Patient",
		StatusCode: 200, Duration: time.Millisecond, ResourceType: "Patient",
	})

	summary := tracker.GetResourceStats("Patient")
	if summary == nil {
		t.Fatal("expected resource stats, got nil")
	}
	if summary.ReadCount != 1 {
		t.Fatalf("expected ReadCount=1 (by-ID), got %d", summary.ReadCount)
	}
	if summary.SearchCount != 1 {
		t.Fatalf("expected SearchCount=1 (list), got %d", summary.SearchCount)
	}
}

// ---------------------------------------------------------------------------
// Overview
// ---------------------------------------------------------------------------

func TestUsageTracker_GetOverview(t *testing.T) {
	tracker := NewUsageTracker(1000)
	tracker.Record(&RequestMetric{
		Timestamp: time.Now(), Method: "GET", Path: "/fhir/Patient",
		StatusCode: 200, Duration: 10 * time.Millisecond, ClientID: "a",
	})
	tracker.Record(&RequestMetric{
		Timestamp: time.Now(), Method: "POST", Path: "/fhir/Observation",
		StatusCode: 500, Duration: 20 * time.Millisecond, ClientID: "b",
	})

	overview := tracker.GetOverview()
	if overview.TotalRequests != 2 {
		t.Fatalf("expected TotalRequests=2, got %d", overview.TotalRequests)
	}
	if overview.TotalErrors != 1 {
		t.Fatalf("expected TotalErrors=1, got %d", overview.TotalErrors)
	}
	if overview.UniqueClients != 2 {
		t.Fatalf("expected UniqueClients=2, got %d", overview.UniqueClients)
	}
	if overview.UniqueEndpoints != 2 {
		t.Fatalf("expected UniqueEndpoints=2, got %d", overview.UniqueEndpoints)
	}
}

func TestUsageTracker_GetErrorRate(t *testing.T) {
	tracker := NewUsageTracker(1000)
	for i := 0; i < 7; i++ {
		tracker.Record(&RequestMetric{
			Timestamp: time.Now(), Method: "GET", Path: "/fhir/Patient",
			StatusCode: 200, Duration: time.Millisecond,
		})
	}
	for i := 0; i < 3; i++ {
		tracker.Record(&RequestMetric{
			Timestamp: time.Now(), Method: "GET", Path: "/fhir/Patient",
			StatusCode: 500, Duration: time.Millisecond,
		})
	}

	rate := tracker.GetErrorRate()
	if rate < 0.29 || rate > 0.31 {
		t.Fatalf("expected error rate ~0.3, got %f", rate)
	}
}

func TestUsageTracker_GetAverageLatency(t *testing.T) {
	tracker := NewUsageTracker(1000)
	tracker.Record(&RequestMetric{
		Timestamp: time.Now(), Method: "GET", Path: "/fhir/Patient",
		StatusCode: 200, Duration: 10 * time.Millisecond,
	})
	tracker.Record(&RequestMetric{
		Timestamp: time.Now(), Method: "GET", Path: "/fhir/Patient",
		StatusCode: 200, Duration: 30 * time.Millisecond,
	})

	avg := tracker.GetAverageLatency()
	if avg != 20*time.Millisecond {
		t.Fatalf("expected avg latency 20ms, got %v", avg)
	}
}

// ---------------------------------------------------------------------------
// Time series
// ---------------------------------------------------------------------------

func TestUsageTracker_GetTimeSeries_1MinBuckets(t *testing.T) {
	tracker := NewUsageTracker(10000)
	now := time.Now().Truncate(time.Minute)

	// Add metrics in two different minutes.
	for i := 0; i < 5; i++ {
		tracker.Record(&RequestMetric{
			Timestamp: now.Add(-2 * time.Minute), Method: "GET", Path: "/fhir/Patient",
			StatusCode: 200, Duration: time.Millisecond,
		})
	}
	for i := 0; i < 3; i++ {
		tracker.Record(&RequestMetric{
			Timestamp: now.Add(-1 * time.Minute), Method: "GET", Path: "/fhir/Patient",
			StatusCode: 200, Duration: time.Millisecond,
		})
	}

	buckets := tracker.GetTimeSeries(time.Minute, 5*time.Minute)
	if len(buckets) == 0 {
		t.Fatal("expected non-empty time series")
	}

	totalCount := int64(0)
	for _, b := range buckets {
		totalCount += b.RequestCount
	}
	if totalCount != 8 {
		t.Fatalf("expected total 8 requests across buckets, got %d", totalCount)
	}
}

func TestUsageTracker_GetTimeSeries_1HourBuckets(t *testing.T) {
	tracker := NewUsageTracker(10000)
	now := time.Now().Truncate(time.Hour)

	for i := 0; i < 10; i++ {
		tracker.Record(&RequestMetric{
			Timestamp: now.Add(-30 * time.Minute), Method: "GET", Path: "/fhir/Patient",
			StatusCode: 200, Duration: time.Millisecond,
		})
	}

	buckets := tracker.GetTimeSeries(time.Hour, 2*time.Hour)
	totalCount := int64(0)
	for _, b := range buckets {
		totalCount += b.RequestCount
	}
	if totalCount != 10 {
		t.Fatalf("expected 10 requests, got %d", totalCount)
	}
}

func TestUsageTracker_GetTimeSeries_EmptyRange(t *testing.T) {
	tracker := NewUsageTracker(1000)
	buckets := tracker.GetTimeSeries(time.Minute, time.Hour)
	// Should return buckets (empty ones) even with no data
	for _, b := range buckets {
		if b.RequestCount != 0 {
			t.Fatalf("expected 0 requests in empty bucket, got %d", b.RequestCount)
		}
	}
}

// ---------------------------------------------------------------------------
// Resource type extraction
// ---------------------------------------------------------------------------

func TestExtractResourceType_PatientByID(t *testing.T) {
	result := extractResourceType("/fhir/Patient/123")
	if result != "Patient" {
		t.Fatalf("expected 'Patient', got %q", result)
	}
}

func TestExtractResourceType_PatientSearch(t *testing.T) {
	result := extractResourceType("/fhir/Patient")
	if result != "Patient" {
		t.Fatalf("expected 'Patient', got %q", result)
	}
}

func TestExtractResourceType_Operation(t *testing.T) {
	result := extractResourceType("/fhir/$export")
	if result != "$export" {
		t.Fatalf("expected '$export', got %q", result)
	}
}

func TestExtractResourceType_NonFHIR(t *testing.T) {
	result := extractResourceType("/api/v1/users")
	if result != "" {
		t.Fatalf("expected empty string for non-FHIR path, got %q", result)
	}
}

// ---------------------------------------------------------------------------
// Middleware
// ---------------------------------------------------------------------------

func TestUsageMiddleware_RecordsMetric(t *testing.T) {
	tracker := NewUsageTracker(1000)
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := UsageMiddleware(tracker)(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	if err := handler(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	overview := tracker.GetOverview()
	if overview.TotalRequests != 1 {
		t.Fatalf("expected 1 recorded metric, got %d", overview.TotalRequests)
	}
}

func TestUsageMiddleware_CapturesStatusCode(t *testing.T) {
	tracker := NewUsageTracker(1000)
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := UsageMiddleware(tracker)(func(c echo.Context) error {
		return c.String(http.StatusNotFound, "not found")
	})

	if err := handler(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	stats := tracker.GetEndpointStats("/fhir/Patient")
	if stats == nil {
		t.Fatal("expected endpoint stats")
	}
	if _, ok := stats.StatusBreakdown[404]; !ok {
		t.Fatalf("expected status 404 in breakdown, got %v", stats.StatusBreakdown)
	}
}

func TestUsageMiddleware_CapturesDuration(t *testing.T) {
	tracker := NewUsageTracker(1000)
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := UsageMiddleware(tracker)(func(c echo.Context) error {
		time.Sleep(5 * time.Millisecond)
		return c.String(http.StatusOK, "ok")
	})

	if err := handler(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	avg := tracker.GetAverageLatency()
	if avg < 5*time.Millisecond {
		t.Fatalf("expected duration >= 5ms, got %v", avg)
	}
}

func TestUsageMiddleware_ExtractsClientID(t *testing.T) {
	tracker := NewUsageTracker(1000)
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("api_key_id", "my-api-key")

	handler := UsageMiddleware(tracker)(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	if err := handler(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	summary := tracker.GetClientStats("my-api-key")
	if summary == nil {
		t.Fatal("expected client stats for api key")
	}
	if summary.TotalRequests != 1 {
		t.Fatalf("expected 1 request, got %d", summary.TotalRequests)
	}
}

// ---------------------------------------------------------------------------
// Handler
// ---------------------------------------------------------------------------

func TestUsageHandler_Overview(t *testing.T) {
	tracker := NewUsageTracker(1000)
	tracker.Record(&RequestMetric{
		Timestamp: time.Now(), Method: "GET", Path: "/fhir/Patient",
		StatusCode: 200, Duration: time.Millisecond, ClientID: "c1",
	})

	e := echo.New()
	h := NewUsageHandler(tracker)
	req := httptest.NewRequest(http.MethodGet, "/admin/analytics/overview", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := h.HandleOverview(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var result UsageOverview
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if result.TotalRequests != 1 {
		t.Fatalf("expected TotalRequests=1, got %d", result.TotalRequests)
	}
}

func TestUsageHandler_TopEndpoints(t *testing.T) {
	tracker := NewUsageTracker(1000)
	for i := 0; i < 5; i++ {
		tracker.Record(&RequestMetric{
			Timestamp: time.Now(), Method: "GET", Path: "/fhir/Patient",
			StatusCode: 200, Duration: time.Millisecond,
		})
	}
	for i := 0; i < 10; i++ {
		tracker.Record(&RequestMetric{
			Timestamp: time.Now(), Method: "GET", Path: "/fhir/Observation",
			StatusCode: 200, Duration: time.Millisecond,
		})
	}

	e := echo.New()
	h := NewUsageHandler(tracker)
	req := httptest.NewRequest(http.MethodGet, "/admin/analytics/endpoints?limit=10", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := h.HandleTopEndpoints(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result []*EndpointSummary
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 endpoints, got %d", len(result))
	}
	if result[0].Path != "/fhir/Observation" {
		t.Fatalf("expected top endpoint /fhir/Observation, got %s", result[0].Path)
	}
}

func TestUsageHandler_TimeSeries(t *testing.T) {
	tracker := NewUsageTracker(10000)
	now := time.Now()
	for i := 0; i < 5; i++ {
		tracker.Record(&RequestMetric{
			Timestamp: now.Add(-30 * time.Second), Method: "GET", Path: "/fhir/Patient",
			StatusCode: 200, Duration: time.Millisecond,
		})
	}

	e := echo.New()
	h := NewUsageHandler(tracker)
	req := httptest.NewRequest(http.MethodGet, "/admin/analytics/timeseries?interval=1m&duration=5m", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := h.HandleTimeSeries(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result []*TimeSeriesBucket
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if len(result) == 0 {
		t.Fatal("expected non-empty time series")
	}

	total := int64(0)
	for _, b := range result {
		total += b.RequestCount
	}
	if total != 5 {
		t.Fatalf("expected 5 total requests, got %d", total)
	}
}

func TestUsageHandler_ClientStats(t *testing.T) {
	tracker := NewUsageTracker(1000)
	for i := 0; i < 7; i++ {
		tracker.Record(&RequestMetric{
			Timestamp: time.Now(), Method: "GET", Path: "/fhir/Patient",
			StatusCode: 200, Duration: time.Millisecond, ClientID: "app-x",
			RequestSize: 100, ResponseSize: 500,
		})
	}

	e := echo.New()
	h := NewUsageHandler(tracker)
	req := httptest.NewRequest(http.MethodGet, "/admin/analytics/clients/app-x", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("app-x")

	if err := h.HandleClientStats(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result ClientSummary
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if result.TotalRequests != 7 {
		t.Fatalf("expected 7 requests, got %d", result.TotalRequests)
	}
	if result.BytesSent != 700 {
		t.Fatalf("expected BytesSent=700, got %d", result.BytesSent)
	}
}
