// Package telemetry provides OpenTelemetry-semantic observability for the
// headless EHR system using only standard library constructs. It exposes
// tracing (span-like structured records), metrics (counters, gauges,
// histograms), and a Prometheus text exposition endpoint -- all without
// importing the go.opentelemetry.io SDK.
package telemetry

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unicode"

	"github.com/labstack/echo/v4"
)

// ---------------------------------------------------------------------------
// Configuration
// ---------------------------------------------------------------------------

// TelemetryConfig holds all configuration for the telemetry provider.
type TelemetryConfig struct {
	ServiceName     string        `json:"service_name"`
	ServiceVersion  string        `json:"service_version"`
	OTLPEndpoint    string        `json:"otlp_endpoint"`    // gRPC endpoint for collector
	MetricsEnabled  *bool         `json:"metrics_enabled"`  // nil = use default (true)
	TracingEnabled  *bool         `json:"tracing_enabled"`  // nil = use default (true)
	MetricsInterval time.Duration `json:"metrics_interval"`
	Environment     string        `json:"environment"`
	SampleRate      float64       `json:"sample_rate"` // 0.0 to 1.0
}

// metricsOn returns whether metrics are enabled (defaults to true).
func (c *TelemetryConfig) metricsOn() bool {
	if c.MetricsEnabled == nil {
		return true
	}
	return *c.MetricsEnabled
}

// tracingOn returns whether tracing is enabled (defaults to true).
func (c *TelemetryConfig) tracingOn() bool {
	if c.TracingEnabled == nil {
		return true
	}
	return *c.TracingEnabled
}

func (c *TelemetryConfig) applyDefaults() {
	if c.ServiceName == "" {
		c.ServiceName = "ehr-server"
	}
	if c.ServiceVersion == "" {
		c.ServiceVersion = "0.0.0"
	}
	if c.Environment == "" {
		c.Environment = "development"
	}
	if c.SampleRate == 0 {
		c.SampleRate = 1.0
	}
	if c.MetricsInterval == 0 {
		c.MetricsInterval = 15 * time.Second
	}
}

// BoolPtr is a helper to create a *bool for TelemetryConfig fields.
func BoolPtr(b bool) *bool {
	return &b
}

// ---------------------------------------------------------------------------
// Span status codes (mirrors OTel SpanStatusCode)
// ---------------------------------------------------------------------------

// SpanStatus represents the status of a completed span.
type SpanStatus int

const (
	// SpanStatusUnset is the default status.
	SpanStatusUnset SpanStatus = iota
	// SpanStatusOK indicates the operation completed successfully.
	SpanStatusOK
	// SpanStatusError indicates the operation contained an error.
	SpanStatusError
)

// ---------------------------------------------------------------------------
// Span — a structured tracing record
// ---------------------------------------------------------------------------

// Span captures a single request's tracing information following OTel semantics.
type Span struct {
	TraceID    string            `json:"trace_id"`
	SpanID     string            `json:"span_id"`
	Name       string            `json:"name"`
	StartTime  time.Time         `json:"start_time"`
	EndTime    time.Time         `json:"end_time"`
	Duration   time.Duration     `json:"duration_ns"`
	StatusCode SpanStatus        `json:"status_code"`
	Attributes map[string]string `json:"attributes"`
}

// JSON serialises the span as a structured JSON string for logging.
func (s *Span) JSON() string {
	b, _ := json.Marshal(s)
	return string(b)
}

// ---------------------------------------------------------------------------
// Histogram — Prometheus-style histogram with buckets
// ---------------------------------------------------------------------------

// histogram is a thread-safe histogram with configurable bucket boundaries.
// Bucket counts are non-cumulative in storage; cumulative counts are computed
// at export time.
type histogram struct {
	boundaries   []float64
	bucketCounts []int64  // one per boundary, non-cumulative
	count        int64
	sum          uint64   // stored as math.Float64bits for atomic add
	mu           sync.Mutex // protects bucketCounts
}

func newHistogram(boundaries []float64) *histogram {
	return &histogram{
		boundaries:   boundaries,
		bucketCounts: make([]int64, len(boundaries)),
	}
}

// Observe records a single value.
func (h *histogram) Observe(v float64) {
	atomic.AddInt64(&h.count, 1)
	atomicAddFloat64(&h.sum, v)

	h.mu.Lock()
	for i, b := range h.boundaries {
		if v <= b {
			h.bucketCounts[i]++
			h.mu.Unlock()
			return
		}
	}
	// Value exceeds all boundaries — counted in +Inf (handled at export).
	h.mu.Unlock()
}

// Count returns the total number of observations.
func (h *histogram) Count() int64 {
	return atomic.LoadInt64(&h.count)
}

// Sum returns the total sum of all observations.
func (h *histogram) Sum() float64 {
	return math.Float64frombits(atomic.LoadUint64(&h.sum))
}

// cumulativeBuckets returns cumulative bucket counts for Prometheus export.
func (h *histogram) cumulativeBuckets() []int64 {
	h.mu.Lock()
	raw := make([]int64, len(h.bucketCounts))
	copy(raw, h.bucketCounts)
	h.mu.Unlock()

	cum := make([]int64, len(raw))
	var running int64
	for i, c := range raw {
		running += c
		cum[i] = running
	}
	return cum
}

// atomicAddFloat64 performs an atomic add on a uint64 that stores a float64
// using CAS.
func atomicAddFloat64(addr *uint64, delta float64) {
	for {
		old := atomic.LoadUint64(addr)
		newVal := math.Float64frombits(old) + delta
		if atomic.CompareAndSwapUint64(addr, old, math.Float64bits(newVal)) {
			return
		}
	}
}

// ---------------------------------------------------------------------------
// Labeled histogram — keyed by (method, route, status_code)
// ---------------------------------------------------------------------------

type labeledHistogramStore struct {
	mu    sync.RWMutex
	items map[string]*histogram
}

func newLabeledHistogramStore() *labeledHistogramStore {
	return &labeledHistogramStore{items: make(map[string]*histogram)}
}

func (s *labeledHistogramStore) getOrCreate(key string, boundaries []float64) *histogram {
	s.mu.RLock()
	h, ok := s.items[key]
	s.mu.RUnlock()
	if ok {
		return h
	}
	s.mu.Lock()
	h, ok = s.items[key]
	if !ok {
		h = newHistogram(boundaries)
		s.items[key] = h
	}
	s.mu.Unlock()
	return h
}

func (s *labeledHistogramStore) snapshot() map[string]*histogram {
	s.mu.RLock()
	defer s.mu.RUnlock()
	cp := make(map[string]*histogram, len(s.items))
	for k, v := range s.items {
		cp[k] = v
	}
	return cp
}

// LabelsKey builds the map key for a labeled histogram. Exported so tests
// can construct the same key.
func LabelsKey(method, route, statusCode string) string {
	return method + "|" + route + "|" + statusCode
}

// ---------------------------------------------------------------------------
// Counter store — keyed by (metricName, label1, label2, ...)
// ---------------------------------------------------------------------------

type counterStore struct {
	mu    sync.RWMutex
	items map[string]*int64
}

func newCounterStore() *counterStore {
	return &counterStore{items: make(map[string]*int64)}
}

func (s *counterStore) inc(key string) {
	s.mu.RLock()
	p, ok := s.items[key]
	s.mu.RUnlock()
	if ok {
		atomic.AddInt64(p, 1)
		return
	}
	s.mu.Lock()
	p, ok = s.items[key]
	if !ok {
		v := int64(1)
		s.items[key] = &v
		s.mu.Unlock()
		return
	}
	s.mu.Unlock()
	atomic.AddInt64(p, 1)
}

func (s *counterStore) get(key string) int64 {
	s.mu.RLock()
	p, ok := s.items[key]
	s.mu.RUnlock()
	if !ok {
		return 0
	}
	return atomic.LoadInt64(p)
}

func (s *counterStore) snapshot() map[string]int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	cp := make(map[string]int64, len(s.items))
	for k, p := range s.items {
		cp[k] = atomic.LoadInt64(p)
	}
	return cp
}

// ---------------------------------------------------------------------------
// Gauge store — keyed by name
// ---------------------------------------------------------------------------

type gaugeStore struct {
	mu    sync.RWMutex
	items map[string]*int64
}

func newGaugeStore() *gaugeStore {
	return &gaugeStore{items: make(map[string]*int64)}
}

func (s *gaugeStore) set(name string, val int64) {
	s.mu.RLock()
	p, ok := s.items[name]
	s.mu.RUnlock()
	if ok {
		atomic.StoreInt64(p, val)
		return
	}
	s.mu.Lock()
	p, ok = s.items[name]
	if !ok {
		v := val
		s.items[name] = &v
		s.mu.Unlock()
		return
	}
	s.mu.Unlock()
	atomic.StoreInt64(p, val)
}

func (s *gaugeStore) add(name string, delta int64) {
	s.mu.RLock()
	p, ok := s.items[name]
	s.mu.RUnlock()
	if ok {
		atomic.AddInt64(p, delta)
		return
	}
	s.mu.Lock()
	p, ok = s.items[name]
	if !ok {
		v := delta
		s.items[name] = &v
		s.mu.Unlock()
		return
	}
	s.mu.Unlock()
	atomic.AddInt64(p, delta)
}

func (s *gaugeStore) get(name string) int64 {
	s.mu.RLock()
	p, ok := s.items[name]
	s.mu.RUnlock()
	if !ok {
		return 0
	}
	return atomic.LoadInt64(p)
}

func (s *gaugeStore) snapshot() map[string]int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	cp := make(map[string]int64, len(s.items))
	for k, p := range s.items {
		cp[k] = atomic.LoadInt64(p)
	}
	return cp
}

// ---------------------------------------------------------------------------
// TelemetryProvider — the main entry point
// ---------------------------------------------------------------------------

// defaultDurationBuckets are the histogram bucket boundaries (in seconds)
// used for HTTP request duration, following OTel HTTP semantic conventions.
var defaultDurationBuckets = []float64{
	0.010, 0.025, 0.050, 0.100, 0.250, 0.500, 1.0, 2.5, 5.0, 10.0,
}

// defaultSizeBuckets are the histogram bucket boundaries (in bytes)
// used for HTTP request/response size.
var defaultSizeBuckets = []float64{
	100, 1_000, 10_000, 100_000, 1_000_000, 10_000_000,
}

// TelemetryProvider manages all observability state.
type TelemetryProvider struct {
	cfg TelemetryConfig

	// Tracing
	spans   []*Span
	spansMu sync.Mutex

	// Metrics — histograms
	histograms        map[string]*histogram
	labeledHistograms map[string]*labeledHistogramStore
	histMu            sync.RWMutex

	// Metrics — counters
	counters *counterStore

	// Metrics — gauges
	gauges *gaugeStore

	// Shutdown
	shutdownOnce sync.Once
	done         chan struct{}
}

// NewTelemetryProvider creates and initialises the telemetry provider.
func NewTelemetryProvider(cfg TelemetryConfig) *TelemetryProvider {
	cfg.applyDefaults()

	tp := &TelemetryProvider{
		cfg:               cfg,
		histograms:        make(map[string]*histogram),
		labeledHistograms: make(map[string]*labeledHistogramStore),
		counters:          newCounterStore(),
		gauges:            newGaugeStore(),
		done:              make(chan struct{}),
	}

	return tp
}

// Shutdown gracefully shuts down the telemetry provider.
func (tp *TelemetryProvider) Shutdown(_ context.Context) error {
	tp.shutdownOnce.Do(func() {
		close(tp.done)
	})
	return nil
}

// Resource returns the OTel resource attributes.
func (tp *TelemetryProvider) Resource() map[string]string {
	return map[string]string{
		"service.name":           tp.cfg.ServiceName,
		"service.version":        tp.cfg.ServiceVersion,
		"deployment.environment": tp.cfg.Environment,
	}
}

// ---------------------------------------------------------------------------
// Span recording (for tracing)
// ---------------------------------------------------------------------------

// GetRecordedSpans returns a copy of all recorded spans.
func (tp *TelemetryProvider) GetRecordedSpans() []*Span {
	tp.spansMu.Lock()
	defer tp.spansMu.Unlock()
	cp := make([]*Span, len(tp.spans))
	copy(cp, tp.spans)
	return cp
}

func (tp *TelemetryProvider) recordSpan(s *Span) {
	tp.spansMu.Lock()
	tp.spans = append(tp.spans, s)
	tp.spansMu.Unlock()
}

// ---------------------------------------------------------------------------
// Metrics accessors (for tests and introspection)
// ---------------------------------------------------------------------------

// GetHistogram returns the named histogram, or nil if it does not exist.
func (tp *TelemetryProvider) GetHistogram(name string) *histogram {
	tp.histMu.RLock()
	defer tp.histMu.RUnlock()
	return tp.histograms[name]
}

func (tp *TelemetryProvider) getOrCreateHistogram(name string, boundaries []float64) *histogram {
	tp.histMu.RLock()
	h, ok := tp.histograms[name]
	tp.histMu.RUnlock()
	if ok {
		return h
	}
	tp.histMu.Lock()
	h, ok = tp.histograms[name]
	if !ok {
		h = newHistogram(boundaries)
		tp.histograms[name] = h
	}
	tp.histMu.Unlock()
	return h
}

func (tp *TelemetryProvider) getOrCreateLabeledStore(name string) *labeledHistogramStore {
	tp.histMu.RLock()
	s, ok := tp.labeledHistograms[name]
	tp.histMu.RUnlock()
	if ok {
		return s
	}
	tp.histMu.Lock()
	s, ok = tp.labeledHistograms[name]
	if !ok {
		s = newLabeledHistogramStore()
		tp.labeledHistograms[name] = s
	}
	tp.histMu.Unlock()
	return s
}

// GetLabeledHistogram returns a specific labeled histogram, or nil.
func (tp *TelemetryProvider) GetLabeledHistogram(name, key string) *histogram {
	tp.histMu.RLock()
	s, ok := tp.labeledHistograms[name]
	tp.histMu.RUnlock()
	if !ok {
		return nil
	}
	s.mu.RLock()
	h := s.items[key]
	s.mu.RUnlock()
	return h
}

// GetGauge returns the current value of the named gauge.
func (tp *TelemetryProvider) GetGauge(name string) int64 {
	return tp.gauges.get(name)
}

// GetCounter returns the current value of a counter with the given name and
// label values (resource_type, operation).
func (tp *TelemetryProvider) GetCounter(name, resourceType, operation string) int64 {
	key := name + "|" + resourceType + "|" + operation
	return tp.counters.get(key)
}

// ---------------------------------------------------------------------------
// FHIROperationCounter
// ---------------------------------------------------------------------------

// FHIROperationCounter increments the fhir.operation.count metric.
func (tp *TelemetryProvider) FHIROperationCounter(resourceType, operation string) {
	key := "fhir.operation.count|" + resourceType + "|" + operation
	tp.counters.inc(key)
}

// ---------------------------------------------------------------------------
// HealthMetrics
// ---------------------------------------------------------------------------

// HealthMetricsRecorder provides methods to update health-related gauges.
type HealthMetricsRecorder struct {
	tp *TelemetryProvider
}

// HealthMetrics returns a recorder for health-related metrics.
func (tp *TelemetryProvider) HealthMetrics() *HealthMetricsRecorder {
	return &HealthMetricsRecorder{tp: tp}
}

// SetDBPoolActive sets the db.pool.active_connections gauge.
func (h *HealthMetricsRecorder) SetDBPoolActive(n int64) {
	h.tp.gauges.set("db.pool.active_connections", n)
}

// SetDBPoolIdle sets the db.pool.idle_connections gauge.
func (h *HealthMetricsRecorder) SetDBPoolIdle(n int64) {
	h.tp.gauges.set("db.pool.idle_connections", n)
}

// SetFHIRResourcesTotal sets the fhir.resources.total gauge.
func (h *HealthMetricsRecorder) SetFHIRResourcesTotal(n int64) {
	h.tp.gauges.set("fhir.resources.total", n)
}

// ---------------------------------------------------------------------------
// TracingMiddleware
// ---------------------------------------------------------------------------

// TracingMiddleware returns an Echo middleware that creates span-like records
// for every HTTP request.
func (tp *TelemetryProvider) TracingMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if !tp.cfg.tracingOn() {
				return next(c)
			}

			start := time.Now()
			req := c.Request()

			// Execute handler.
			err := next(c)

			end := time.Now()
			resp := c.Response()
			statusCode := resp.Status

			// Use route pattern, not actual path.
			route := c.Path()
			if route == "" {
				route = req.URL.Path
			}

			// Build span name: HTTP {method} {route_pattern}
			spanName := "HTTP " + req.Method + " " + route

			// Determine span status.
			var status SpanStatus
			if statusCode >= 500 {
				status = SpanStatusError
			} else {
				status = SpanStatusOK
			}

			// Extract FHIR resource type from the actual path.
			resourceType := extractFHIRResourceType(req.URL.Path)

			// Extract tenant ID from Echo context.
			tenantID := ""
			if v := c.Get("tenant_id"); v != nil {
				if s, ok := v.(string); ok {
					tenantID = s
				}
			}

			attrs := map[string]string{
				"http.method":      req.Method,
				"http.route":       route,
				"http.status_code": fmt.Sprintf("%d", statusCode),
				"http.url":         req.URL.String(),
			}
			if resourceType != "" {
				attrs["fhir.resource_type"] = resourceType
			}
			if tenantID != "" {
				attrs["tenant.id"] = tenantID
			}

			span := &Span{
				TraceID:    generateID(16),
				SpanID:     generateID(8),
				Name:       spanName,
				StartTime:  start,
				EndTime:    end,
				Duration:   end.Sub(start),
				StatusCode: status,
				Attributes: attrs,
			}

			tp.recordSpan(span)

			return err
		}
	}
}

// ---------------------------------------------------------------------------
// MetricsMiddleware
// ---------------------------------------------------------------------------

// MetricsMiddleware returns an Echo middleware that records HTTP server metrics.
func (tp *TelemetryProvider) MetricsMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if !tp.cfg.metricsOn() {
				return next(c)
			}

			// Increment active requests.
			tp.gauges.add("http.server.active_requests", 1)

			start := time.Now()
			req := c.Request()

			// Execute handler.
			err := next(c)

			duration := time.Since(start).Seconds()
			resp := c.Response()
			statusCode := resp.Status

			// Decrement active requests.
			tp.gauges.add("http.server.active_requests", -1)

			// Route pattern.
			route := c.Path()
			if route == "" {
				route = req.URL.Path
			}
			statusStr := fmt.Sprintf("%d", statusCode)

			// Record duration in the global histogram.
			durationHist := tp.getOrCreateHistogram("http.server.request.duration", defaultDurationBuckets)
			durationHist.Observe(duration)

			// Record duration in labeled histogram.
			store := tp.getOrCreateLabeledStore("http.server.request.duration")
			key := LabelsKey(req.Method, route, statusStr)
			labeled := store.getOrCreate(key, defaultDurationBuckets)
			labeled.Observe(duration)

			// Record request size (from Content-Length).
			if req.ContentLength > 0 {
				reqSizeHist := tp.getOrCreateHistogram("http.server.request.size", defaultSizeBuckets)
				reqSizeHist.Observe(float64(req.ContentLength))
			}

			// Record response size.
			respSize := resp.Size
			if respSize > 0 {
				respSizeHist := tp.getOrCreateHistogram("http.server.response.size", defaultSizeBuckets)
				respSizeHist.Observe(float64(respSize))
			}

			return err
		}
	}
}

// ---------------------------------------------------------------------------
// PrometheusHandler
// ---------------------------------------------------------------------------

// PrometheusHandler returns an Echo handler that serves metrics in Prometheus
// text exposition format at /metrics.
func (tp *TelemetryProvider) PrometheusHandler() echo.HandlerFunc {
	return func(c echo.Context) error {
		var b strings.Builder

		// --- http_server_request_duration_seconds (histogram) ---
		tp.histMu.RLock()
		durationHist := tp.histograms["http.server.request.duration"]
		durationStore := tp.labeledHistograms["http.server.request.duration"]
		reqSizeHist := tp.histograms["http.server.request.size"]
		respSizeHist := tp.histograms["http.server.response.size"]
		tp.histMu.RUnlock()

		writeHistogramMetric(&b, "http_server_request_duration_seconds",
			"Duration of HTTP requests in seconds.", "histogram",
			durationHist, durationStore, defaultDurationBuckets)

		// --- http_server_active_requests (gauge) ---
		b.WriteString("# HELP http_server_active_requests Number of active HTTP requests.\n")
		b.WriteString("# TYPE http_server_active_requests gauge\n")
		fmt.Fprintf(&b, "http_server_active_requests %d\n", tp.gauges.get("http.server.active_requests"))
		b.WriteByte('\n')

		// --- http_server_request_size_bytes (histogram) ---
		writeSimpleHistogram(&b, "http_server_request_size_bytes",
			"Size of HTTP request bodies in bytes.", reqSizeHist, defaultSizeBuckets)

		// --- http_server_response_size_bytes (histogram) ---
		writeSimpleHistogram(&b, "http_server_response_size_bytes",
			"Size of HTTP response bodies in bytes.", respSizeHist, defaultSizeBuckets)

		// --- fhir_operation_count (counter) ---
		counters := tp.counters.snapshot()
		b.WriteString("# HELP fhir_operation_count Total FHIR operations by resource type and operation.\n")
		b.WriteString("# TYPE fhir_operation_count counter\n")
		for key, val := range counters {
			parts := strings.SplitN(key, "|", 3)
			if len(parts) == 3 && parts[0] == "fhir.operation.count" {
				fmt.Fprintf(&b, "fhir_operation_count{resource_type=%q,operation=%q} %d\n",
					parts[1], parts[2], val)
			}
		}
		b.WriteByte('\n')

		// --- Health gauges ---
		healthGauges := []struct {
			promName string
			otelName string
			help     string
		}{
			{"db_pool_active_connections", "db.pool.active_connections", "Number of active database pool connections."},
			{"db_pool_idle_connections", "db.pool.idle_connections", "Number of idle database pool connections."},
			{"fhir_resources_total", "fhir.resources.total", "Total number of FHIR resources."},
		}
		for _, g := range healthGauges {
			val := tp.gauges.get(g.otelName)
			fmt.Fprintf(&b, "# HELP %s %s\n", g.promName, g.help)
			fmt.Fprintf(&b, "# TYPE %s gauge\n", g.promName)
			fmt.Fprintf(&b, "%s %d\n", g.promName, val)
			b.WriteByte('\n')
		}

		return c.String(http.StatusOK, b.String())
	}
}

// ---------------------------------------------------------------------------
// Prometheus format helpers
// ---------------------------------------------------------------------------

func writeHistogramMetric(b *strings.Builder, name, help, typ string,
	global *histogram, labeled *labeledHistogramStore, boundaries []float64) {

	fmt.Fprintf(b, "# HELP %s %s\n", name, help)
	fmt.Fprintf(b, "# TYPE %s %s\n", name, typ)

	if labeled != nil {
		snap := labeled.snapshot()
		for key, h := range snap {
			parts := strings.SplitN(key, "|", 3)
			if len(parts) != 3 {
				continue
			}
			method, route, status := parts[0], parts[1], parts[2]
			labels := fmt.Sprintf("method=%q,route=%q,status_code=%q", method, route, status)
			writeSingleHistogram(b, name, labels, h, boundaries)
		}
	} else if global != nil {
		writeSingleHistogram(b, name, "", global, boundaries)
	}
	b.WriteByte('\n')
}

func writeSimpleHistogram(b *strings.Builder, name, help string,
	h *histogram, boundaries []float64) {

	fmt.Fprintf(b, "# HELP %s %s\n", name, help)
	fmt.Fprintf(b, "# TYPE %s histogram\n", name)
	if h != nil {
		writeSingleHistogram(b, name, "", h, boundaries)
	}
	b.WriteByte('\n')
}

func writeSingleHistogram(b *strings.Builder, name, labels string,
	h *histogram, boundaries []float64) {

	cum := h.cumulativeBuckets()
	total := h.Count()

	labelsPrefix := ""
	labelsSuffix := ""
	if labels != "" {
		labelsPrefix = labels + ","
		labelsSuffix = "{" + labels + "}"
	}

	for i, boundary := range boundaries {
		if labels != "" {
			fmt.Fprintf(b, "%s_bucket{%sle=\"%g\"} %d\n", name, labelsPrefix, boundary, cum[i])
		} else {
			fmt.Fprintf(b, "%s_bucket{le=\"%g\"} %d\n", name, boundary, cum[i])
		}
	}

	// +Inf bucket.
	if labels != "" {
		fmt.Fprintf(b, "%s_bucket{%sle=\"+Inf\"} %d\n", name, labelsPrefix, total)
	} else {
		fmt.Fprintf(b, "%s_bucket{le=\"+Inf\"} %d\n", name, total)
	}

	fmt.Fprintf(b, "%s_sum%s %g\n", name, labelsSuffix, h.Sum())
	fmt.Fprintf(b, "%s_count%s %d\n", name, labelsSuffix, total)
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// extractFHIRResourceType parses a FHIR resource type from a URL path.
// It returns "" for non-FHIR paths, operation paths ($export), or empty
// segments.
func extractFHIRResourceType(path string) string {
	const prefix = "/fhir/"
	idx := strings.Index(path, prefix)
	if idx < 0 {
		return ""
	}

	rest := path[idx+len(prefix):]
	if rest == "" {
		return ""
	}

	// Take up to next slash.
	if slashIdx := strings.IndexByte(rest, '/'); slashIdx >= 0 {
		rest = rest[:slashIdx]
	}

	// Must start with uppercase letter (FHIR resource types are PascalCase).
	if len(rest) == 0 || !unicode.IsUpper(rune(rest[0])) {
		return ""
	}

	return rest
}

// generateID produces a random hex string of n bytes (2n hex chars).
func generateID(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
