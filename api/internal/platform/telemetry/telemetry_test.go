package telemetry

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
)

// ---------------------------------------------------------------------------
// Config defaults
// ---------------------------------------------------------------------------

func TestTelemetryConfig_Defaults(t *testing.T) {
	cfg := TelemetryConfig{}
	tp := NewTelemetryProvider(cfg)
	defer tp.Shutdown(context.Background())

	if tp.cfg.ServiceName != "ehr-server" {
		t.Fatalf("expected default ServiceName='ehr-server', got %q", tp.cfg.ServiceName)
	}
	if tp.cfg.ServiceVersion != "0.0.0" {
		t.Fatalf("expected default ServiceVersion='0.0.0', got %q", tp.cfg.ServiceVersion)
	}
	if tp.cfg.Environment != "development" {
		t.Fatalf("expected default Environment='development', got %q", tp.cfg.Environment)
	}
	if tp.cfg.SampleRate != 1.0 {
		t.Fatalf("expected default SampleRate=1.0, got %f", tp.cfg.SampleRate)
	}
	if tp.cfg.MetricsInterval != 15*time.Second {
		t.Fatalf("expected default MetricsInterval=15s, got %v", tp.cfg.MetricsInterval)
	}
	if !tp.cfg.metricsOn() {
		t.Fatal("expected MetricsEnabled=true by default")
	}
	if !tp.cfg.tracingOn() {
		t.Fatal("expected TracingEnabled=true by default")
	}
}

func TestTelemetryConfig_CustomValues(t *testing.T) {
	cfg := TelemetryConfig{
		ServiceName:     "my-ehr",
		ServiceVersion:  "1.2.3",
		OTLPEndpoint:    "localhost:4317",
		MetricsEnabled:  BoolPtr(true),
		TracingEnabled:  BoolPtr(true),
		MetricsInterval: 30 * time.Second,
		Environment:     "production",
		SampleRate:      0.5,
	}
	tp := NewTelemetryProvider(cfg)
	defer tp.Shutdown(context.Background())

	if tp.cfg.ServiceName != "my-ehr" {
		t.Fatalf("expected ServiceName='my-ehr', got %q", tp.cfg.ServiceName)
	}
	if tp.cfg.ServiceVersion != "1.2.3" {
		t.Fatalf("expected ServiceVersion='1.2.3', got %q", tp.cfg.ServiceVersion)
	}
	if tp.cfg.Environment != "production" {
		t.Fatalf("expected Environment='production', got %q", tp.cfg.Environment)
	}
	if tp.cfg.SampleRate != 0.5 {
		t.Fatalf("expected SampleRate=0.5, got %f", tp.cfg.SampleRate)
	}
}

// ---------------------------------------------------------------------------
// Shutdown
// ---------------------------------------------------------------------------

func TestShutdown_Clean(t *testing.T) {
	tp := NewTelemetryProvider(TelemetryConfig{})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := tp.Shutdown(ctx)
	if err != nil {
		t.Fatalf("expected clean shutdown, got error: %v", err)
	}

	// Calling shutdown again should not panic.
	err = tp.Shutdown(ctx)
	if err != nil {
		t.Fatalf("second shutdown should not error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Noop behavior when disabled
// ---------------------------------------------------------------------------

func TestNoop_WhenDisabled(t *testing.T) {
	tp := NewTelemetryProvider(TelemetryConfig{
		MetricsEnabled: BoolPtr(false),
		TracingEnabled: BoolPtr(false),
	})
	defer tp.Shutdown(context.Background())

	// Tracing middleware should still work as passthrough.
	e := echo.New()
	e.Use(tp.TracingMiddleware())
	e.GET("/test", func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	// Metrics middleware should still work as passthrough.
	e2 := echo.New()
	e2.Use(tp.MetricsMiddleware())
	e2.GET("/test2", func(c echo.Context) error {
		return c.String(http.StatusOK, "ok2")
	})

	req2 := httptest.NewRequest(http.MethodGet, "/test2", nil)
	rec2 := httptest.NewRecorder()
	e2.ServeHTTP(rec2, req2)

	if rec2.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec2.Code)
	}

	// When disabled, spans should not be recorded.
	spans := tp.GetRecordedSpans()
	if len(spans) != 0 {
		t.Fatalf("expected 0 spans when tracing disabled, got %d", len(spans))
	}
}

// ---------------------------------------------------------------------------
// TracingMiddleware
// ---------------------------------------------------------------------------

func TestTracingMiddleware_CreatesSpan(t *testing.T) {
	tp := NewTelemetryProvider(TelemetryConfig{TracingEnabled: BoolPtr(true)})
	defer tp.Shutdown(context.Background())

	e := echo.New()
	e.Use(tp.TracingMiddleware())
	e.GET("/fhir/Patient/:id", func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient/123", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	spans := tp.GetRecordedSpans()
	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}

	span := spans[0]
	if span.Name != "HTTP GET /fhir/Patient/:id" {
		t.Fatalf("expected span name 'HTTP GET /fhir/Patient/:id', got %q", span.Name)
	}
}

func TestTracingMiddleware_SpanAttributes(t *testing.T) {
	tp := NewTelemetryProvider(TelemetryConfig{TracingEnabled: BoolPtr(true)})
	defer tp.Shutdown(context.Background())

	e := echo.New()
	e.Use(tp.TracingMiddleware())
	e.GET("/fhir/Patient/:id", func(c echo.Context) error {
		return c.String(http.StatusOK, "patient data")
	})

	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient/123", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	spans := tp.GetRecordedSpans()
	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}

	span := spans[0]

	assertAttribute(t, span, "http.method", "GET")
	assertAttribute(t, span, "http.route", "/fhir/Patient/:id")
	assertAttribute(t, span, "http.status_code", "200")
	assertAttribute(t, span, "fhir.resource_type", "Patient")
}

func TestTracingMiddleware_SpanAttributeURL(t *testing.T) {
	tp := NewTelemetryProvider(TelemetryConfig{TracingEnabled: BoolPtr(true)})
	defer tp.Shutdown(context.Background())

	e := echo.New()
	e.Use(tp.TracingMiddleware())
	e.GET("/fhir/Patient/:id", func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient/123?_include=Observation", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	spans := tp.GetRecordedSpans()
	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}

	url, ok := spans[0].Attributes["http.url"]
	if !ok {
		t.Fatal("expected http.url attribute")
	}
	if !strings.Contains(url, "/fhir/Patient/123") {
		t.Fatalf("expected URL to contain path, got %q", url)
	}
}

func TestTracingMiddleware_TenantID(t *testing.T) {
	tp := NewTelemetryProvider(TelemetryConfig{TracingEnabled: BoolPtr(true)})
	defer tp.Shutdown(context.Background())

	e := echo.New()
	// Inject tenant_id into context before tracing middleware.
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set("tenant_id", "tenant-abc")
			return next(c)
		}
	})
	e.Use(tp.TracingMiddleware())
	e.GET("/fhir/Patient", func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	spans := tp.GetRecordedSpans()
	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}

	assertAttribute(t, spans[0], "tenant.id", "tenant-abc")
}

func TestTracingMiddleware_SpanStatusError(t *testing.T) {
	tp := NewTelemetryProvider(TelemetryConfig{TracingEnabled: BoolPtr(true)})
	defer tp.Shutdown(context.Background())

	e := echo.New()
	e.Use(tp.TracingMiddleware())
	e.GET("/error", func(c echo.Context) error {
		return c.String(http.StatusInternalServerError, "error")
	})

	req := httptest.NewRequest(http.MethodGet, "/error", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	spans := tp.GetRecordedSpans()
	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}

	if spans[0].StatusCode != SpanStatusError {
		t.Fatalf("expected span status Error, got %v", spans[0].StatusCode)
	}
}

func TestTracingMiddleware_SpanStatusOK(t *testing.T) {
	tp := NewTelemetryProvider(TelemetryConfig{TracingEnabled: BoolPtr(true)})
	defer tp.Shutdown(context.Background())

	e := echo.New()
	e.Use(tp.TracingMiddleware())
	e.GET("/ok", func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/ok", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	spans := tp.GetRecordedSpans()
	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}

	if spans[0].StatusCode != SpanStatusOK {
		t.Fatalf("expected span status OK, got %v", spans[0].StatusCode)
	}
}

// ---------------------------------------------------------------------------
// MetricsMiddleware — request duration
// ---------------------------------------------------------------------------

func TestMetricsMiddleware_RecordsDuration(t *testing.T) {
	tp := NewTelemetryProvider(TelemetryConfig{MetricsEnabled: BoolPtr(true)})
	defer tp.Shutdown(context.Background())

	e := echo.New()
	e.Use(tp.MetricsMiddleware())
	e.GET("/fhir/Patient", func(c echo.Context) error {
		time.Sleep(5 * time.Millisecond) // ensure measurable duration
		return c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	hist := tp.GetHistogram("http.server.request.duration")
	if hist == nil {
		t.Fatal("expected http.server.request.duration histogram to exist")
	}

	if hist.Count() == 0 {
		t.Fatal("expected at least 1 observation in duration histogram")
	}

	if hist.Sum() <= 0 {
		t.Fatal("expected positive sum in duration histogram")
	}
}

// ---------------------------------------------------------------------------
// MetricsMiddleware — active requests
// ---------------------------------------------------------------------------

func TestMetricsMiddleware_ActiveRequests(t *testing.T) {
	tp := NewTelemetryProvider(TelemetryConfig{MetricsEnabled: BoolPtr(true)})
	defer tp.Shutdown(context.Background())

	activeObserved := make(chan int64, 1)

	e := echo.New()
	e.Use(tp.MetricsMiddleware())
	e.GET("/slow", func(c echo.Context) error {
		// Capture active requests while handling.
		val := tp.GetGauge("http.server.active_requests")
		activeObserved <- val
		return c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/slow", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	active := <-activeObserved
	if active != 1 {
		t.Fatalf("expected active_requests=1 during handling, got %d", active)
	}

	// After request completes, gauge should be back to 0.
	val := tp.GetGauge("http.server.active_requests")
	if val != 0 {
		t.Fatalf("expected active_requests=0 after request, got %d", val)
	}
}

// ---------------------------------------------------------------------------
// MetricsMiddleware — labels include method, route, status
// ---------------------------------------------------------------------------

func TestMetricsMiddleware_Labels(t *testing.T) {
	tp := NewTelemetryProvider(TelemetryConfig{MetricsEnabled: BoolPtr(true)})
	defer tp.Shutdown(context.Background())

	e := echo.New()
	e.Use(tp.MetricsMiddleware())
	e.POST("/fhir/Patient", func(c echo.Context) error {
		return c.String(http.StatusCreated, "created")
	})

	req := httptest.NewRequest(http.MethodPost, "/fhir/Patient", strings.NewReader(`{"resourceType":"Patient"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	// Check labeled histogram exists.
	key := LabelsKey("POST", "/fhir/Patient", "201")
	hist := tp.GetLabeledHistogram("http.server.request.duration", key)
	if hist == nil {
		t.Fatalf("expected labeled histogram for key %q", key)
	}
	if hist.Count() != 1 {
		t.Fatalf("expected count=1, got %d", hist.Count())
	}
}

// ---------------------------------------------------------------------------
// MetricsMiddleware — request/response size
// ---------------------------------------------------------------------------

func TestMetricsMiddleware_RequestSize(t *testing.T) {
	tp := NewTelemetryProvider(TelemetryConfig{MetricsEnabled: BoolPtr(true)})
	defer tp.Shutdown(context.Background())

	e := echo.New()
	e.Use(tp.MetricsMiddleware())
	e.POST("/fhir/Patient", func(c echo.Context) error {
		return c.String(http.StatusCreated, "created")
	})

	body := `{"resourceType":"Patient","name":[{"family":"Smith"}]}`
	req := httptest.NewRequest(http.MethodPost, "/fhir/Patient", strings.NewReader(body))
	req.Header.Set("Content-Length", fmt.Sprintf("%d", len(body)))
	req.ContentLength = int64(len(body))
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	hist := tp.GetHistogram("http.server.request.size")
	if hist == nil {
		t.Fatal("expected http.server.request.size histogram to exist")
	}
	if hist.Count() != 1 {
		t.Fatalf("expected count=1 for request size, got %d", hist.Count())
	}
	if hist.Sum() != float64(len(body)) {
		t.Fatalf("expected request size sum=%d, got %f", len(body), hist.Sum())
	}
}

func TestMetricsMiddleware_ResponseSize(t *testing.T) {
	tp := NewTelemetryProvider(TelemetryConfig{MetricsEnabled: BoolPtr(true)})
	defer tp.Shutdown(context.Background())

	e := echo.New()
	e.Use(tp.MetricsMiddleware())
	e.GET("/test", func(c echo.Context) error {
		return c.String(http.StatusOK, "hello world response")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	hist := tp.GetHistogram("http.server.response.size")
	if hist == nil {
		t.Fatal("expected http.server.response.size histogram to exist")
	}
	if hist.Count() != 1 {
		t.Fatalf("expected count=1 for response size, got %d", hist.Count())
	}
	if hist.Sum() <= 0 {
		t.Fatal("expected positive response size sum")
	}
}

// ---------------------------------------------------------------------------
// FHIROperationCounter
// ---------------------------------------------------------------------------

func TestFHIROperationCounter_Increments(t *testing.T) {
	tp := NewTelemetryProvider(TelemetryConfig{MetricsEnabled: BoolPtr(true)})
	defer tp.Shutdown(context.Background())

	tp.FHIROperationCounter("Patient", "read")
	tp.FHIROperationCounter("Patient", "read")
	tp.FHIROperationCounter("Patient", "create")
	tp.FHIROperationCounter("Observation", "search")

	count := tp.GetCounter("fhir.operation.count", "Patient", "read")
	if count != 2 {
		t.Fatalf("expected Patient/read count=2, got %d", count)
	}

	count = tp.GetCounter("fhir.operation.count", "Patient", "create")
	if count != 1 {
		t.Fatalf("expected Patient/create count=1, got %d", count)
	}

	count = tp.GetCounter("fhir.operation.count", "Observation", "search")
	if count != 1 {
		t.Fatalf("expected Observation/search count=1, got %d", count)
	}
}

// ---------------------------------------------------------------------------
// PrometheusHandler — valid text format
// ---------------------------------------------------------------------------

func TestPrometheusHandler_ValidFormat(t *testing.T) {
	tp := NewTelemetryProvider(TelemetryConfig{MetricsEnabled: BoolPtr(true)})
	defer tp.Shutdown(context.Background())

	// Record some metrics.
	e := echo.New()
	e.Use(tp.MetricsMiddleware())
	e.GET("/fhir/Patient", func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})
	e.GET("/metrics", tp.PrometheusHandler())

	// Generate some traffic.
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodGet, "/fhir/Patient", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
	}

	tp.FHIROperationCounter("Patient", "read")

	// Now request metrics.
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	body := rec.Body.String()

	// Check that required metrics are present.
	requiredMetrics := []string{
		"http_server_request_duration_seconds",
		"http_server_active_requests",
		"http_server_request_size_bytes",
		"http_server_response_size_bytes",
		"fhir_operation_count",
	}

	for _, m := range requiredMetrics {
		if !strings.Contains(body, m) {
			t.Errorf("expected metrics output to contain %q, body:\n%s", m, body)
		}
	}

	// Prometheus format uses # HELP and # TYPE lines.
	if !strings.Contains(body, "# HELP") {
		t.Error("expected Prometheus HELP comments in output")
	}
	if !strings.Contains(body, "# TYPE") {
		t.Error("expected Prometheus TYPE comments in output")
	}
}

// ---------------------------------------------------------------------------
// Histogram buckets
// ---------------------------------------------------------------------------

func TestHistogramBuckets(t *testing.T) {
	expectedBuckets := []float64{
		0.010, 0.025, 0.050, 0.100, 0.250, 0.500, 1.0, 2.5, 5.0, 10.0,
	}

	h := newHistogram(expectedBuckets)

	if len(h.boundaries) != len(expectedBuckets) {
		t.Fatalf("expected %d bucket boundaries, got %d", len(expectedBuckets), len(h.boundaries))
	}

	for i, b := range expectedBuckets {
		if h.boundaries[i] != b {
			t.Fatalf("expected bucket[%d]=%f, got %f", i, b, h.boundaries[i])
		}
	}
}

func TestHistogramBuckets_Observation(t *testing.T) {
	buckets := []float64{0.010, 0.025, 0.050, 0.100, 0.250, 0.500, 1.0, 2.5, 5.0, 10.0}
	h := newHistogram(buckets)

	// 5ms = 0.005s -> falls into the first bucket (le=0.010)
	h.Observe(0.005)
	// 15ms = 0.015s -> falls into the second bucket (le=0.025)
	h.Observe(0.015)
	// 3s -> falls into the 9th bucket (le=5.0)
	h.Observe(3.0)

	if h.Count() != 3 {
		t.Fatalf("expected count=3, got %d", h.Count())
	}

	// Check bucket counts.
	// le=0.010: 1 observation (0.005)
	if h.bucketCounts[0] != 1 {
		t.Fatalf("expected bucket[0.010]=1, got %d", h.bucketCounts[0])
	}
	// le=0.025: 1 observation (0.015) -- not cumulative in storage, cumulative in export
	if h.bucketCounts[1] != 1 {
		t.Fatalf("expected bucket[0.025]=1, got %d", h.bucketCounts[1])
	}
	// le=5.0: 1 observation (3.0)
	if h.bucketCounts[8] != 1 {
		t.Fatalf("expected bucket[5.0]=1, got %d", h.bucketCounts[8])
	}
}

// ---------------------------------------------------------------------------
// FHIR resource type extraction
// ---------------------------------------------------------------------------

func TestExtractFHIRResourceType(t *testing.T) {
	tests := []struct {
		path     string
		expected string
	}{
		{"/fhir/Patient/123", "Patient"},
		{"/fhir/Patient", "Patient"},
		{"/fhir/Observation/abc", "Observation"},
		{"/fhir/MedicationRequest", "MedicationRequest"},
		{"/fhir/$export", ""},
		{"/api/v1/users", ""},
		{"/health", ""},
		{"", ""},
		{"/fhir/", ""},
		{"/fhir/Patient/123/_history/1", "Patient"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := extractFHIRResourceType(tt.path)
			if got != tt.expected {
				t.Fatalf("extractFHIRResourceType(%q) = %q, want %q", tt.path, got, tt.expected)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Tenant ID extraction
// ---------------------------------------------------------------------------

func TestExtractTenantID(t *testing.T) {
	tp := NewTelemetryProvider(TelemetryConfig{TracingEnabled: BoolPtr(true)})
	defer tp.Shutdown(context.Background())

	e := echo.New()
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set("tenant_id", "tenant-xyz")
			return next(c)
		}
	})
	e.Use(tp.TracingMiddleware())
	e.GET("/fhir/Patient", func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	spans := tp.GetRecordedSpans()
	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}

	val, ok := spans[0].Attributes["tenant.id"]
	if !ok || val != "tenant-xyz" {
		t.Fatalf("expected tenant.id='tenant-xyz', got %q (ok=%v)", val, ok)
	}
}

// ---------------------------------------------------------------------------
// Concurrent safety (race detector test)
// ---------------------------------------------------------------------------

func TestMetrics_ConcurrentSafe(t *testing.T) {
	tp := NewTelemetryProvider(TelemetryConfig{
		MetricsEnabled: BoolPtr(true),
		TracingEnabled: BoolPtr(true),
	})
	defer tp.Shutdown(context.Background())

	e := echo.New()
	e.Use(tp.TracingMiddleware())
	e.Use(tp.MetricsMiddleware())
	e.GET("/fhir/Patient/:id", func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})
	e.POST("/fhir/Patient", func(c echo.Context) error {
		return c.String(http.StatusCreated, "created")
	})

	var wg sync.WaitGroup
	goroutines := 50
	requestsPerGoroutine := 20

	for g := 0; g < goroutines; g++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for i := 0; i < requestsPerGoroutine; i++ {
				var req *http.Request
				if i%2 == 0 {
					req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/fhir/Patient/%d", i), nil)
				} else {
					req = httptest.NewRequest(http.MethodPost, "/fhir/Patient", strings.NewReader(`{}`))
				}
				rec := httptest.NewRecorder()
				e.ServeHTTP(rec, req)
			}
		}(g)
	}

	// Concurrently read metrics while writing.
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			tp.FHIROperationCounter("Patient", "read")
			tp.GetGauge("http.server.active_requests")
			tp.GetHistogram("http.server.request.duration")
			time.Sleep(time.Millisecond)
		}
	}()

	wg.Wait()

	totalExpected := int64(goroutines * requestsPerGoroutine)
	hist := tp.GetHistogram("http.server.request.duration")
	if hist == nil {
		t.Fatal("expected duration histogram to exist after concurrent test")
	}
	if hist.Count() != totalExpected {
		t.Fatalf("expected count=%d, got %d", totalExpected, hist.Count())
	}
}

// ---------------------------------------------------------------------------
// HealthMetrics
// ---------------------------------------------------------------------------

func TestHealthMetrics_DBPool(t *testing.T) {
	tp := NewTelemetryProvider(TelemetryConfig{MetricsEnabled: BoolPtr(true)})
	defer tp.Shutdown(context.Background())

	hm := tp.HealthMetrics()

	hm.SetDBPoolActive(5)
	hm.SetDBPoolIdle(10)

	if tp.GetGauge("db.pool.active_connections") != 5 {
		t.Fatalf("expected db.pool.active_connections=5, got %d", tp.GetGauge("db.pool.active_connections"))
	}
	if tp.GetGauge("db.pool.idle_connections") != 10 {
		t.Fatalf("expected db.pool.idle_connections=10, got %d", tp.GetGauge("db.pool.idle_connections"))
	}
}

func TestHealthMetrics_FHIRResourcesTotal(t *testing.T) {
	tp := NewTelemetryProvider(TelemetryConfig{MetricsEnabled: BoolPtr(true)})
	defer tp.Shutdown(context.Background())

	hm := tp.HealthMetrics()
	hm.SetFHIRResourcesTotal(42000)

	if tp.GetGauge("fhir.resources.total") != 42000 {
		t.Fatalf("expected fhir.resources.total=42000, got %d", tp.GetGauge("fhir.resources.total"))
	}
}

func TestHealthMetrics_InPrometheusOutput(t *testing.T) {
	tp := NewTelemetryProvider(TelemetryConfig{MetricsEnabled: BoolPtr(true)})
	defer tp.Shutdown(context.Background())

	hm := tp.HealthMetrics()
	hm.SetDBPoolActive(3)
	hm.SetDBPoolIdle(7)
	hm.SetFHIRResourcesTotal(1000)

	e := echo.New()
	e.GET("/metrics", tp.PrometheusHandler())

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	body := rec.Body.String()

	if !strings.Contains(body, "db_pool_active_connections 3") {
		t.Errorf("expected db_pool_active_connections in output, got:\n%s", body)
	}
	if !strings.Contains(body, "db_pool_idle_connections 7") {
		t.Errorf("expected db_pool_idle_connections in output, got:\n%s", body)
	}
	if !strings.Contains(body, "fhir_resources_total 1000") {
		t.Errorf("expected fhir_resources_total in output, got:\n%s", body)
	}
}

// ---------------------------------------------------------------------------
// Span JSON serialization
// ---------------------------------------------------------------------------

func TestSpan_JSONSerialization(t *testing.T) {
	tp := NewTelemetryProvider(TelemetryConfig{TracingEnabled: BoolPtr(true)})
	defer tp.Shutdown(context.Background())

	e := echo.New()
	e.Use(tp.TracingMiddleware())
	e.GET("/fhir/Patient/:id", func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient/123", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	spans := tp.GetRecordedSpans()
	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}

	jsonStr := spans[0].JSON()
	if jsonStr == "" {
		t.Fatal("expected non-empty JSON")
	}
	if !strings.Contains(jsonStr, "HTTP GET /fhir/Patient/:id") {
		t.Fatalf("expected span name in JSON, got %s", jsonStr)
	}
	if !strings.Contains(jsonStr, "trace_id") {
		t.Fatalf("expected trace_id in JSON, got %s", jsonStr)
	}
	if !strings.Contains(jsonStr, "span_id") {
		t.Fatalf("expected span_id in JSON, got %s", jsonStr)
	}
}

// ---------------------------------------------------------------------------
// Resource info in provider
// ---------------------------------------------------------------------------

func TestProvider_Resource(t *testing.T) {
	tp := NewTelemetryProvider(TelemetryConfig{
		ServiceName:    "test-ehr",
		ServiceVersion: "2.0.0",
		Environment:    "staging",
	})
	defer tp.Shutdown(context.Background())

	res := tp.Resource()
	if res["service.name"] != "test-ehr" {
		t.Fatalf("expected service.name='test-ehr', got %q", res["service.name"])
	}
	if res["service.version"] != "2.0.0" {
		t.Fatalf("expected service.version='2.0.0', got %q", res["service.version"])
	}
	if res["deployment.environment"] != "staging" {
		t.Fatalf("expected deployment.environment='staging', got %q", res["deployment.environment"])
	}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func assertAttribute(t *testing.T, span *Span, key, expected string) {
	t.Helper()
	val, ok := span.Attributes[key]
	if !ok {
		t.Fatalf("expected attribute %q to exist in span", key)
	}
	if val != expected {
		t.Fatalf("expected attribute %q=%q, got %q", key, expected, val)
	}
}
