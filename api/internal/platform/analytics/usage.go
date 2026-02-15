package analytics

import (
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/labstack/echo/v4"
)

// ---------------------------------------------------------------------------
// Core metric type
// ---------------------------------------------------------------------------

// RequestMetric captures a single API request's metadata for analytics.
type RequestMetric struct {
	Timestamp    time.Time     `json:"timestamp"`
	Method       string        `json:"method"`
	Path         string        `json:"path"`
	StatusCode   int           `json:"status_code"`
	Duration     time.Duration `json:"duration"`
	ClientID     string        `json:"client_id"`
	TenantID     string        `json:"tenant_id"`
	ResourceType string        `json:"resource_type"`
	RequestSize  int64         `json:"request_size"`
	ResponseSize int64         `json:"response_size"`
}

// ---------------------------------------------------------------------------
// Internal counter types
// ---------------------------------------------------------------------------

type endpointStats struct {
	Path          string
	TotalRequests int64
	TotalErrors   int64
	TotalDuration int64 // nanoseconds
	StatusCounts  map[int]int64
	mu            sync.Mutex
}

type clientStats struct {
	ClientID      string
	TotalRequests int64
	TotalErrors   int64
	LastRequestAt time.Time
	BytesSent     int64
	BytesReceived int64
	mu            sync.Mutex
}

type resourceStats struct {
	ResourceType string
	ReadCount    int64
	CreateCount  int64
	UpdateCount  int64
	DeleteCount  int64
	SearchCount  int64
	mu           sync.Mutex
}

// ---------------------------------------------------------------------------
// Summary types (returned by query methods)
// ---------------------------------------------------------------------------

// EndpointSummary provides aggregated statistics for a single API endpoint.
type EndpointSummary struct {
	Path            string         `json:"path"`
	TotalRequests   int64          `json:"total_requests"`
	ErrorRate       float64        `json:"error_rate"`
	AvgLatency      time.Duration  `json:"avg_latency"`
	P95Latency      time.Duration  `json:"p95_latency"`
	StatusBreakdown map[int]int64  `json:"status_breakdown"`
}

// ClientSummary provides aggregated statistics for a single API client.
type ClientSummary struct {
	ClientID      string    `json:"client_id"`
	TotalRequests int64     `json:"total_requests"`
	ErrorRate     float64   `json:"error_rate"`
	LastSeen      time.Time `json:"last_seen"`
	BytesSent     int64     `json:"bytes_sent"`
	BytesReceived int64     `json:"bytes_received"`
}

// ResourceSummary provides CRUD+Search breakdown for a FHIR resource type.
type ResourceSummary struct {
	ResourceType string `json:"resource_type"`
	ReadCount    int64  `json:"read_count"`
	CreateCount  int64  `json:"create_count"`
	UpdateCount  int64  `json:"update_count"`
	DeleteCount  int64  `json:"delete_count"`
	SearchCount  int64  `json:"search_count"`
	Total        int64  `json:"total"`
}

// UsageOverview provides a high-level summary of API usage.
type UsageOverview struct {
	TotalRequests   int64              `json:"total_requests"`
	TotalErrors     int64              `json:"total_errors"`
	ErrorRate       float64            `json:"error_rate"`
	AvgLatency      time.Duration      `json:"avg_latency"`
	UniqueClients   int                `json:"unique_clients"`
	UniqueEndpoints int                `json:"unique_endpoints"`
	TopEndpoints    []*EndpointSummary `json:"top_endpoints"`
	TopClients      []*ClientSummary   `json:"top_clients"`
}

// TimeSeriesBucket holds aggregated metrics for a single time bucket.
type TimeSeriesBucket struct {
	Timestamp    time.Time     `json:"timestamp"`
	RequestCount int64         `json:"request_count"`
	ErrorCount   int64         `json:"error_count"`
	AvgLatency   time.Duration `json:"avg_latency"`
}

// ---------------------------------------------------------------------------
// UsageTracker — the main thread-safe analytics aggregator
// ---------------------------------------------------------------------------

// UsageTracker provides thread-safe API usage tracking with an append-only
// ring buffer and per-endpoint, per-client, and per-resource counters.
type UsageTracker struct {
	metrics          []*RequestMetric
	maxMetrics       int
	writePos         int
	full             bool
	endpointCounters map[string]*endpointStats
	clientCounters   map[string]*clientStats
	resourceCounters map[string]*resourceStats
	mu               sync.RWMutex
	totalRequests    int64
	totalErrors      int64
	totalDuration    int64 // nanoseconds
}

// NewUsageTracker creates a new UsageTracker with the given ring buffer capacity.
func NewUsageTracker(maxMetrics int) *UsageTracker {
	if maxMetrics <= 0 {
		maxMetrics = 100000
	}
	return &UsageTracker{
		metrics:          make([]*RequestMetric, 0, maxMetrics),
		maxMetrics:       maxMetrics,
		endpointCounters: make(map[string]*endpointStats),
		clientCounters:   make(map[string]*clientStats),
		resourceCounters: make(map[string]*resourceStats),
	}
}

// Record appends a metric to the ring buffer and updates all counters.
func (ut *UsageTracker) Record(metric *RequestMetric) {
	isError := metric.StatusCode >= 400

	// Update atomic totals.
	atomic.AddInt64(&ut.totalRequests, 1)
	if isError {
		atomic.AddInt64(&ut.totalErrors, 1)
	}
	atomic.AddInt64(&ut.totalDuration, int64(metric.Duration))

	ut.mu.Lock()

	// Ring buffer insert.
	if ut.full {
		ut.metrics[ut.writePos] = metric
	} else if len(ut.metrics) < ut.maxMetrics {
		ut.metrics = append(ut.metrics, metric)
	}
	ut.writePos++
	if ut.writePos >= ut.maxMetrics {
		ut.writePos = 0
		ut.full = true
	}

	// Endpoint counters.
	ep, ok := ut.endpointCounters[metric.Path]
	if !ok {
		ep = &endpointStats{
			Path:         metric.Path,
			StatusCounts: make(map[int]int64),
		}
		ut.endpointCounters[metric.Path] = ep
	}

	// Client counters.
	var cs *clientStats
	if metric.ClientID != "" {
		cs, ok = ut.clientCounters[metric.ClientID]
		if !ok {
			cs = &clientStats{ClientID: metric.ClientID}
			ut.clientCounters[metric.ClientID] = cs
		}
	}

	// Resource counters.
	var rs *resourceStats
	if metric.ResourceType != "" {
		rs, ok = ut.resourceCounters[metric.ResourceType]
		if !ok {
			rs = &resourceStats{ResourceType: metric.ResourceType}
			ut.resourceCounters[metric.ResourceType] = rs
		}
	}

	ut.mu.Unlock()

	// Update endpoint stats (per-endpoint mutex to reduce contention).
	ep.mu.Lock()
	ep.TotalRequests++
	if isError {
		ep.TotalErrors++
	}
	ep.TotalDuration += int64(metric.Duration)
	ep.StatusCounts[metric.StatusCode]++
	ep.mu.Unlock()

	// Update client stats.
	if cs != nil {
		cs.mu.Lock()
		cs.TotalRequests++
		if isError {
			cs.TotalErrors++
		}
		cs.LastRequestAt = metric.Timestamp
		cs.BytesSent += metric.RequestSize
		cs.BytesReceived += metric.ResponseSize
		cs.mu.Unlock()
	}

	// Update resource stats.
	if rs != nil {
		rs.mu.Lock()
		switch metric.Method {
		case "POST":
			rs.CreateCount++
		case "PUT", "PATCH":
			rs.UpdateCount++
		case "DELETE":
			rs.DeleteCount++
		case "GET":
			if isReadByID(metric.Path, metric.ResourceType) {
				rs.ReadCount++
			} else {
				rs.SearchCount++
			}
		}
		rs.mu.Unlock()
	}
}

// isReadByID checks whether a GET request targets a specific resource by ID
// (e.g., /fhir/Patient/123) rather than a search/list (e.g., /fhir/Patient).
func isReadByID(path, resourceType string) bool {
	if resourceType == "" {
		return false
	}
	// After /fhir/<ResourceType> there should be another segment for an ID.
	idx := strings.Index(path, resourceType)
	if idx < 0 {
		return false
	}
	rest := path[idx+len(resourceType):]
	return len(rest) > 1 && rest[0] == '/'
}

// ---------------------------------------------------------------------------
// Query methods
// ---------------------------------------------------------------------------

// GetEndpointStats returns aggregated stats for a single endpoint path.
func (ut *UsageTracker) GetEndpointStats(path string) *EndpointSummary {
	ut.mu.RLock()
	ep, ok := ut.endpointCounters[path]
	ut.mu.RUnlock()
	if !ok {
		return nil
	}
	return ut.buildEndpointSummary(ep)
}

// GetClientStats returns aggregated stats for a single client.
func (ut *UsageTracker) GetClientStats(clientID string) *ClientSummary {
	ut.mu.RLock()
	cs, ok := ut.clientCounters[clientID]
	ut.mu.RUnlock()
	if !ok {
		return nil
	}
	return ut.buildClientSummary(cs)
}

// GetResourceStats returns CRUD+Search breakdown for a resource type.
func (ut *UsageTracker) GetResourceStats(resourceType string) *ResourceSummary {
	ut.mu.RLock()
	rs, ok := ut.resourceCounters[resourceType]
	ut.mu.RUnlock()
	if !ok {
		return nil
	}

	rs.mu.Lock()
	summary := &ResourceSummary{
		ResourceType: rs.ResourceType,
		ReadCount:    rs.ReadCount,
		CreateCount:  rs.CreateCount,
		UpdateCount:  rs.UpdateCount,
		DeleteCount:  rs.DeleteCount,
		SearchCount:  rs.SearchCount,
		Total:        rs.ReadCount + rs.CreateCount + rs.UpdateCount + rs.DeleteCount + rs.SearchCount,
	}
	rs.mu.Unlock()
	return summary
}

// GetOverview returns a high-level usage summary.
func (ut *UsageTracker) GetOverview() *UsageOverview {
	total := atomic.LoadInt64(&ut.totalRequests)
	errors := atomic.LoadInt64(&ut.totalErrors)
	dur := atomic.LoadInt64(&ut.totalDuration)

	var errorRate float64
	if total > 0 {
		errorRate = float64(errors) / float64(total)
	}

	var avgLatency time.Duration
	if total > 0 {
		avgLatency = time.Duration(dur / total)
	}

	ut.mu.RLock()
	uniqueClients := len(ut.clientCounters)
	uniqueEndpoints := len(ut.endpointCounters)
	ut.mu.RUnlock()

	return &UsageOverview{
		TotalRequests:   total,
		TotalErrors:     errors,
		ErrorRate:       errorRate,
		AvgLatency:      avgLatency,
		UniqueClients:   uniqueClients,
		UniqueEndpoints: uniqueEndpoints,
		TopEndpoints:    ut.GetTopEndpoints(5),
		TopClients:      ut.GetTopClients(5),
	}
}

// GetTopEndpoints returns the top N endpoints sorted by request count descending.
func (ut *UsageTracker) GetTopEndpoints(limit int) []*EndpointSummary {
	ut.mu.RLock()
	summaries := make([]*EndpointSummary, 0, len(ut.endpointCounters))
	for _, ep := range ut.endpointCounters {
		summaries = append(summaries, ut.buildEndpointSummary(ep))
	}
	ut.mu.RUnlock()

	sort.Slice(summaries, func(i, j int) bool {
		return summaries[i].TotalRequests > summaries[j].TotalRequests
	})

	if limit > len(summaries) {
		limit = len(summaries)
	}
	return summaries[:limit]
}

// GetTopClients returns the top N clients sorted by request count descending.
func (ut *UsageTracker) GetTopClients(limit int) []*ClientSummary {
	ut.mu.RLock()
	summaries := make([]*ClientSummary, 0, len(ut.clientCounters))
	for _, cs := range ut.clientCounters {
		summaries = append(summaries, ut.buildClientSummary(cs))
	}
	ut.mu.RUnlock()

	sort.Slice(summaries, func(i, j int) bool {
		return summaries[i].TotalRequests > summaries[j].TotalRequests
	})

	if limit > len(summaries) {
		limit = len(summaries)
	}
	return summaries[:limit]
}

// GetTimeSeries returns request counts bucketed by the given interval over the
// specified lookback duration.
func (ut *UsageTracker) GetTimeSeries(interval, duration time.Duration) []*TimeSeriesBucket {
	now := time.Now()
	start := now.Add(-duration).Truncate(interval)
	numBuckets := int(duration/interval) + 1

	buckets := make([]*TimeSeriesBucket, numBuckets)
	for i := 0; i < numBuckets; i++ {
		buckets[i] = &TimeSeriesBucket{
			Timestamp: start.Add(time.Duration(i) * interval),
		}
	}

	ut.mu.RLock()
	metricsCopy := make([]*RequestMetric, len(ut.metrics))
	copy(metricsCopy, ut.metrics)
	ut.mu.RUnlock()

	for _, m := range metricsCopy {
		if m == nil {
			continue
		}
		if m.Timestamp.Before(start) || m.Timestamp.After(now) {
			continue
		}
		idx := int(m.Timestamp.Sub(start) / interval)
		if idx < 0 || idx >= numBuckets {
			continue
		}
		buckets[idx].RequestCount++
		if m.StatusCode >= 400 {
			buckets[idx].ErrorCount++
		}
		buckets[idx].AvgLatency += m.Duration // accumulate, we'll average below
	}

	for _, b := range buckets {
		if b.RequestCount > 0 {
			b.AvgLatency = time.Duration(int64(b.AvgLatency) / b.RequestCount)
		}
	}

	return buckets
}

// GetErrorRate returns the overall error rate as a float between 0 and 1.
func (ut *UsageTracker) GetErrorRate() float64 {
	total := atomic.LoadInt64(&ut.totalRequests)
	errors := atomic.LoadInt64(&ut.totalErrors)
	if total == 0 {
		return 0
	}
	return float64(errors) / float64(total)
}

// GetAverageLatency returns the average request duration.
func (ut *UsageTracker) GetAverageLatency() time.Duration {
	total := atomic.LoadInt64(&ut.totalRequests)
	dur := atomic.LoadInt64(&ut.totalDuration)
	if total == 0 {
		return 0
	}
	return time.Duration(dur / total)
}

// ---------------------------------------------------------------------------
// Internal helpers
// ---------------------------------------------------------------------------

func (ut *UsageTracker) buildEndpointSummary(ep *endpointStats) *EndpointSummary {
	ep.mu.Lock()
	defer ep.mu.Unlock()

	var errorRate float64
	if ep.TotalRequests > 0 {
		errorRate = float64(ep.TotalErrors) / float64(ep.TotalRequests)
	}

	var avgLatency time.Duration
	if ep.TotalRequests > 0 {
		avgLatency = time.Duration(ep.TotalDuration / ep.TotalRequests)
	}

	statusBreakdown := make(map[int]int64, len(ep.StatusCounts))
	for code, count := range ep.StatusCounts {
		statusBreakdown[code] = count
	}

	// P95 requires the stored metrics; we compute it from the ring buffer.
	p95 := ut.computeP95ForPath(ep.Path)

	return &EndpointSummary{
		Path:            ep.Path,
		TotalRequests:   ep.TotalRequests,
		ErrorRate:       errorRate,
		AvgLatency:      avgLatency,
		P95Latency:      p95,
		StatusBreakdown: statusBreakdown,
	}
}

func (ut *UsageTracker) buildClientSummary(cs *clientStats) *ClientSummary {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	var errorRate float64
	if cs.TotalRequests > 0 {
		errorRate = float64(cs.TotalErrors) / float64(cs.TotalRequests)
	}

	return &ClientSummary{
		ClientID:      cs.ClientID,
		TotalRequests: cs.TotalRequests,
		ErrorRate:     errorRate,
		LastSeen:      cs.LastRequestAt,
		BytesSent:     cs.BytesSent,
		BytesReceived: cs.BytesReceived,
	}
}

func (ut *UsageTracker) computeP95ForPath(path string) time.Duration {
	ut.mu.RLock()
	var durations []time.Duration
	for _, m := range ut.metrics {
		if m != nil && m.Path == path {
			durations = append(durations, m.Duration)
		}
	}
	ut.mu.RUnlock()

	if len(durations) == 0 {
		return 0
	}
	sort.Slice(durations, func(i, j int) bool { return durations[i] < durations[j] })
	idx := int(float64(len(durations)) * 0.95)
	if idx >= len(durations) {
		idx = len(durations) - 1
	}
	return durations[idx]
}

// ---------------------------------------------------------------------------
// FHIR resource type extraction
// ---------------------------------------------------------------------------

// extractResourceType parses a FHIR resource type from a URL path.
// Examples:
//   - "/fhir/Patient/123"  → "Patient"
//   - "/fhir/Patient"      → "Patient"
//   - "/fhir/$export"      → "$export"
//   - "/api/v1/users"      → ""
func extractResourceType(path string) string {
	// Only extract from FHIR paths.
	const fhirPrefix = "/fhir/"
	idx := strings.Index(path, fhirPrefix)
	if idx < 0 {
		return ""
	}

	rest := path[idx+len(fhirPrefix):]
	if rest == "" {
		return ""
	}

	// Take everything up to the next slash (or end of string).
	if slashIdx := strings.Index(rest, "/"); slashIdx >= 0 {
		return rest[:slashIdx]
	}
	return rest
}

// ---------------------------------------------------------------------------
// Echo middleware
// ---------------------------------------------------------------------------

// responseCapture wraps an echo.Response to capture the status code that
// was actually written by the handler.
type responseCapture struct {
	statusCode   int
	responseSize int64
	written      bool
}

// UsageMiddleware returns Echo middleware that records every request into the
// provided UsageTracker.
func UsageMiddleware(tracker *UsageTracker) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			req := c.Request()
			path := req.URL.Path

			// Execute handler.
			err := next(c)

			duration := time.Since(start)
			resp := c.Response()
			statusCode := resp.Status
			responseSize := resp.Size

			// Extract client ID from context.
			clientID := ""
			if v := c.Get("api_key_id"); v != nil {
				if s, ok := v.(string); ok {
					clientID = s
				}
			}
			if clientID == "" {
				if v := c.Get("sub"); v != nil {
					if s, ok := v.(string); ok {
						clientID = s
					}
				}
			}

			// Extract tenant ID.
			tenantID := ""
			if v := c.Get("tenant_id"); v != nil {
				if s, ok := v.(string); ok {
					tenantID = s
				}
			}

			resourceType := extractResourceType(path)

			var requestSize int64
			if req.ContentLength > 0 {
				requestSize = req.ContentLength
			}

			tracker.Record(&RequestMetric{
				Timestamp:    start,
				Method:       req.Method,
				Path:         path,
				StatusCode:   statusCode,
				Duration:     duration,
				ClientID:     clientID,
				TenantID:     tenantID,
				ResourceType: resourceType,
				RequestSize:  requestSize,
				ResponseSize: responseSize,
			})

			return err
		}
	}
}

// ---------------------------------------------------------------------------
// Echo HTTP handler
// ---------------------------------------------------------------------------

// UsageHandler provides HTTP endpoints for querying API usage analytics.
type UsageHandler struct {
	tracker *UsageTracker
}

// NewUsageHandler creates a new handler backed by the given tracker.
func NewUsageHandler(tracker *UsageTracker) *UsageHandler {
	return &UsageHandler{tracker: tracker}
}

// RegisterRoutes registers the analytics admin endpoints on the provided group.
func (h *UsageHandler) RegisterRoutes(g *echo.Group) {
	g.GET("/analytics/overview", h.HandleOverview)
	g.GET("/analytics/endpoints", h.HandleTopEndpoints)
	g.GET("/analytics/endpoints/:path", h.HandleEndpointStats)
	g.GET("/analytics/clients", h.HandleTopClients)
	g.GET("/analytics/clients/:id", h.HandleClientStats)
	g.GET("/analytics/resources", h.HandleResources)
	g.GET("/analytics/timeseries", h.HandleTimeSeries)
}

// HandleOverview returns overall API usage statistics.
func (h *UsageHandler) HandleOverview(c echo.Context) error {
	return c.JSON(http.StatusOK, h.tracker.GetOverview())
}

// HandleTopEndpoints returns the top endpoints sorted by request count.
func (h *UsageHandler) HandleTopEndpoints(c echo.Context) error {
	limit := 20
	if l := c.QueryParam("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	return c.JSON(http.StatusOK, h.tracker.GetTopEndpoints(limit))
}

// HandleEndpointStats returns stats for a specific endpoint path.
func (h *UsageHandler) HandleEndpointStats(c echo.Context) error {
	path := "/" + c.Param("path")
	summary := h.tracker.GetEndpointStats(path)
	if summary == nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "endpoint not found"})
	}
	return c.JSON(http.StatusOK, summary)
}

// HandleTopClients returns the top clients sorted by request count.
func (h *UsageHandler) HandleTopClients(c echo.Context) error {
	limit := 20
	if l := c.QueryParam("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	return c.JSON(http.StatusOK, h.tracker.GetTopClients(limit))
}

// HandleClientStats returns stats for a specific client.
func (h *UsageHandler) HandleClientStats(c echo.Context) error {
	id := c.Param("id")
	summary := h.tracker.GetClientStats(id)
	if summary == nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "client not found"})
	}
	return c.JSON(http.StatusOK, summary)
}

// HandleResources returns CRUD+Search breakdown for all resource types.
func (h *UsageHandler) HandleResources(c echo.Context) error {
	h.tracker.mu.RLock()
	summaries := make([]*ResourceSummary, 0, len(h.tracker.resourceCounters))
	for _, rs := range h.tracker.resourceCounters {
		rs.mu.Lock()
		summaries = append(summaries, &ResourceSummary{
			ResourceType: rs.ResourceType,
			ReadCount:    rs.ReadCount,
			CreateCount:  rs.CreateCount,
			UpdateCount:  rs.UpdateCount,
			DeleteCount:  rs.DeleteCount,
			SearchCount:  rs.SearchCount,
			Total:        rs.ReadCount + rs.CreateCount + rs.UpdateCount + rs.DeleteCount + rs.SearchCount,
		})
		rs.mu.Unlock()
	}
	h.tracker.mu.RUnlock()

	return c.JSON(http.StatusOK, summaries)
}

// HandleTimeSeries returns time-bucketed request counts.
func (h *UsageHandler) HandleTimeSeries(c echo.Context) error {
	interval := parseDurationParam(c.QueryParam("interval"), time.Minute)
	duration := parseDurationParam(c.QueryParam("duration"), time.Hour)

	return c.JSON(http.StatusOK, h.tracker.GetTimeSeries(interval, duration))
}

// parseDurationParam parses a human-friendly duration string like "1m", "5m",
// "1h", "24h", "7d" into a time.Duration.
func parseDurationParam(s string, defaultVal time.Duration) time.Duration {
	if s == "" {
		return defaultVal
	}

	// Handle "d" suffix for days.
	if strings.HasSuffix(s, "d") {
		numStr := strings.TrimSuffix(s, "d")
		if n, err := strconv.Atoi(numStr); err == nil {
			return time.Duration(n) * 24 * time.Hour
		}
		return defaultVal
	}

	d, err := time.ParseDuration(s)
	if err != nil {
		return defaultVal
	}
	return d
}
