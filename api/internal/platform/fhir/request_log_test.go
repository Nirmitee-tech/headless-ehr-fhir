package fhir

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
)

func TestClassifyInteraction_Read(t *testing.T) {
	got := ClassifyInteraction(http.MethodGet, "/fhir/Patient/123")
	if got != "read" {
		t.Errorf("expected read, got %s", got)
	}
}

func TestClassifyInteraction_VRead(t *testing.T) {
	got := ClassifyInteraction(http.MethodGet, "/fhir/Patient/123/_history/1")
	if got != "vread" {
		t.Errorf("expected vread, got %s", got)
	}
}

func TestClassifyInteraction_SearchType(t *testing.T) {
	got := ClassifyInteraction(http.MethodGet, "/fhir/Patient")
	if got != "search-type" {
		t.Errorf("expected search-type, got %s", got)
	}
}

func TestClassifyInteraction_Create(t *testing.T) {
	got := ClassifyInteraction(http.MethodPost, "/fhir/Patient")
	if got != "create" {
		t.Errorf("expected create, got %s", got)
	}
}

func TestClassifyInteraction_Update(t *testing.T) {
	got := ClassifyInteraction(http.MethodPut, "/fhir/Patient/123")
	if got != "update" {
		t.Errorf("expected update, got %s", got)
	}
}

func TestClassifyInteraction_Patch(t *testing.T) {
	got := ClassifyInteraction(http.MethodPatch, "/fhir/Patient/123")
	if got != "update" {
		t.Errorf("expected update, got %s", got)
	}
}

func TestClassifyInteraction_Delete(t *testing.T) {
	got := ClassifyInteraction(http.MethodDelete, "/fhir/Patient/123")
	if got != "delete" {
		t.Errorf("expected delete, got %s", got)
	}
}

func TestClassifyInteraction_HistoryInstance(t *testing.T) {
	got := ClassifyInteraction(http.MethodGet, "/fhir/Patient/123/_history")
	if got != "history-instance" {
		t.Errorf("expected history-instance, got %s", got)
	}
}

func TestClassifyInteraction_HistoryType(t *testing.T) {
	got := ClassifyInteraction(http.MethodGet, "/fhir/Patient/_history")
	if got != "history-type" {
		t.Errorf("expected history-type, got %s", got)
	}
}

func TestClassifyInteraction_Operation(t *testing.T) {
	tests := []struct {
		name   string
		method string
		path   string
	}{
		{"validate", http.MethodPost, "/fhir/Patient/$validate"},
		{"export", http.MethodGet, "/fhir/Patient/$export"},
		{"instance-op", http.MethodPost, "/fhir/Patient/123/$validate"},
		{"everything", http.MethodGet, "/fhir/Patient/123/$everything"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ClassifyInteraction(tt.method, tt.path)
			if got != "operation" {
				t.Errorf("expected operation, got %s", got)
			}
		})
	}
}

func TestClassifyInteraction_NoBasePath(t *testing.T) {
	got := ClassifyInteraction(http.MethodGet, "/Patient/123")
	if got != "read" {
		t.Errorf("expected read, got %s", got)
	}
}

func TestExtractResourceInfo_Read(t *testing.T) {
	resType, resID, op := ExtractResourceInfo("/fhir/Patient/123")
	if resType != "Patient" {
		t.Errorf("expected Patient, got %s", resType)
	}
	if resID != "123" {
		t.Errorf("expected 123, got %s", resID)
	}
	if op != "" {
		t.Errorf("expected empty operation, got %s", op)
	}
}

func TestExtractResourceInfo_SearchType(t *testing.T) {
	resType, resID, op := ExtractResourceInfo("/fhir/Patient")
	if resType != "Patient" {
		t.Errorf("expected Patient, got %s", resType)
	}
	if resID != "" {
		t.Errorf("expected empty id, got %s", resID)
	}
	if op != "" {
		t.Errorf("expected empty operation, got %s", op)
	}
}

func TestExtractResourceInfo_Operation(t *testing.T) {
	resType, resID, op := ExtractResourceInfo("/fhir/Patient/$validate")
	if resType != "Patient" {
		t.Errorf("expected Patient, got %s", resType)
	}
	if resID != "" {
		t.Errorf("expected empty id, got %s", resID)
	}
	if op != "$validate" {
		t.Errorf("expected $validate, got %s", op)
	}
}

func TestExtractResourceInfo_InstanceOperation(t *testing.T) {
	resType, resID, op := ExtractResourceInfo("/fhir/Patient/123/$validate")
	if resType != "Patient" {
		t.Errorf("expected Patient, got %s", resType)
	}
	if resID != "123" {
		t.Errorf("expected 123, got %s", resID)
	}
	if op != "$validate" {
		t.Errorf("expected $validate, got %s", op)
	}
}

func TestExtractResourceInfo_History(t *testing.T) {
	resType, resID, op := ExtractResourceInfo("/fhir/Patient/123/_history/1")
	if resType != "Patient" {
		t.Errorf("expected Patient, got %s", resType)
	}
	if resID != "123" {
		t.Errorf("expected 123, got %s", resID)
	}
	if op != "" {
		t.Errorf("expected empty operation, got %s", op)
	}
}

func TestExtractResourceInfo_HistoryType(t *testing.T) {
	resType, resID, op := ExtractResourceInfo("/fhir/Patient/_history")
	if resType != "Patient" {
		t.Errorf("expected Patient, got %s", resType)
	}
	if resID != "" {
		t.Errorf("expected empty id, got %s", resID)
	}
	if op != "" {
		t.Errorf("expected empty operation, got %s", op)
	}
}

func TestExtractResourceInfo_NoBasePath(t *testing.T) {
	resType, resID, op := ExtractResourceInfo("/Observation/obs-1")
	if resType != "Observation" {
		t.Errorf("expected Observation, got %s", resType)
	}
	if resID != "obs-1" {
		t.Errorf("expected obs-1, got %s", resID)
	}
	if op != "" {
		t.Errorf("expected empty operation, got %s", op)
	}
}

func TestChannelLogSink_Buffering(t *testing.T) {
	sink := NewChannelLogSink(10)

	entry := FHIRRequestLog{
		Timestamp:    time.Now(),
		Method:       http.MethodGet,
		Path:         "/fhir/Patient/123",
		ResourceType: "Patient",
		ResourceID:   "123",
		Interaction:  "read",
		StatusCode:   http.StatusOK,
		Duration:     42,
		ClientIP:     "127.0.0.1",
	}

	sink.Log(entry)

	select {
	case got := <-sink.Entries():
		if got.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", got.Method)
		}
		if got.ResourceType != "Patient" {
			t.Errorf("expected Patient, got %s", got.ResourceType)
		}
		if got.ResourceID != "123" {
			t.Errorf("expected 123, got %s", got.ResourceID)
		}
		if got.Interaction != "read" {
			t.Errorf("expected read, got %s", got.Interaction)
		}
		if got.StatusCode != http.StatusOK {
			t.Errorf("expected 200, got %d", got.StatusCode)
		}
		if got.Duration != 42 {
			t.Errorf("expected 42ms, got %d", got.Duration)
		}
	default:
		t.Fatal("expected entry in channel but got none")
	}
}

func TestChannelLogSink_DropsWhenFull(t *testing.T) {
	sink := NewChannelLogSink(1)

	entry := FHIRRequestLog{Method: http.MethodGet}
	sink.Log(entry)
	// Second log should be silently dropped (buffer size = 1).
	sink.Log(entry)

	// Drain the single buffered entry.
	<-sink.Entries()

	select {
	case <-sink.Entries():
		t.Fatal("expected channel to be empty after draining one entry")
	default:
		// Expected: channel is empty.
	}
}

func TestChannelLogSink_MultipleEntries(t *testing.T) {
	sink := NewChannelLogSink(5)

	for i := 0; i < 5; i++ {
		sink.Log(FHIRRequestLog{StatusCode: 200 + i})
	}

	for i := 0; i < 5; i++ {
		select {
		case got := <-sink.Entries():
			if got.StatusCode != 200+i {
				t.Errorf("entry %d: expected status %d, got %d", i, 200+i, got.StatusCode)
			}
		default:
			t.Fatalf("expected entry %d but channel was empty", i)
		}
	}
}

func TestFHIRRequestLoggerMiddleware_RecordsEntry(t *testing.T) {
	sink := NewChannelLogSink(10)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient/456", nil)
	req.Header.Set("User-Agent", "TestAgent/1.0")
	req.Header.Set("X-Tenant-ID", "tenant-abc")
	req.Header.Set("X-Request-ID", "req-789")
	req.Header.Set("Prefer", "return=minimal")
	req.Header.Set("If-None-Match", `W/"1"`)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := FHIRRequestLoggerMiddleware(sink)(func(c echo.Context) error {
		return c.String(http.StatusOK, `{"resourceType":"Patient","id":"456"}`)
	})

	if err := handler(c); err != nil {
		t.Fatal(err)
	}

	// The response should still be written through.
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	select {
	case entry := <-sink.Entries():
		if entry.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", entry.Method)
		}
		if entry.Path != "/fhir/Patient/456" {
			t.Errorf("expected /fhir/Patient/456, got %s", entry.Path)
		}
		if entry.ResourceType != "Patient" {
			t.Errorf("expected Patient, got %s", entry.ResourceType)
		}
		if entry.ResourceID != "456" {
			t.Errorf("expected 456, got %s", entry.ResourceID)
		}
		if entry.Interaction != "read" {
			t.Errorf("expected read, got %s", entry.Interaction)
		}
		if entry.StatusCode != http.StatusOK {
			t.Errorf("expected 200, got %d", entry.StatusCode)
		}
		if entry.Duration < 0 {
			t.Errorf("expected non-negative duration, got %d", entry.Duration)
		}
		if entry.ResponseSize == 0 {
			t.Error("expected non-zero response size")
		}
		if entry.UserAgent != "TestAgent/1.0" {
			t.Errorf("expected TestAgent/1.0, got %s", entry.UserAgent)
		}
		if entry.TenantID != "tenant-abc" {
			t.Errorf("expected tenant-abc, got %s", entry.TenantID)
		}
		if entry.RequestID != "req-789" {
			t.Errorf("expected req-789, got %s", entry.RequestID)
		}
		if entry.PreferHeader != "return=minimal" {
			t.Errorf("expected return=minimal, got %s", entry.PreferHeader)
		}
		if entry.IfNoneMatch != `W/"1"` {
			t.Errorf("expected W/\"1\", got %s", entry.IfNoneMatch)
		}
	default:
		t.Fatal("expected log entry in channel but got none")
	}
}

func TestFHIRRequestLoggerMiddleware_ErrorResponse(t *testing.T) {
	sink := NewChannelLogSink(10)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient/999", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := FHIRRequestLoggerMiddleware(sink)(func(c echo.Context) error {
		return c.String(http.StatusNotFound, `{"resourceType":"OperationOutcome"}`)
	})

	if err := handler(c); err != nil {
		t.Fatal(err)
	}

	select {
	case entry := <-sink.Entries():
		if entry.StatusCode != http.StatusNotFound {
			t.Errorf("expected 404, got %d", entry.StatusCode)
		}
		if entry.Interaction != "read" {
			t.Errorf("expected read, got %s", entry.Interaction)
		}
	default:
		t.Fatal("expected log entry in channel but got none")
	}
}

func TestFHIRRequestLoggerMiddleware_PostCreate(t *testing.T) {
	sink := NewChannelLogSink(10)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/fhir/Observation", nil)
	req.Header.Set("If-Match", `W/"2"`)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := FHIRRequestLoggerMiddleware(sink)(func(c echo.Context) error {
		return c.String(http.StatusCreated, `{"resourceType":"Observation","id":"new-1"}`)
	})

	if err := handler(c); err != nil {
		t.Fatal(err)
	}

	select {
	case entry := <-sink.Entries():
		if entry.Interaction != "create" {
			t.Errorf("expected create, got %s", entry.Interaction)
		}
		if entry.ResourceType != "Observation" {
			t.Errorf("expected Observation, got %s", entry.ResourceType)
		}
		if entry.StatusCode != http.StatusCreated {
			t.Errorf("expected 201, got %d", entry.StatusCode)
		}
		if entry.IfMatch != `W/"2"` {
			t.Errorf("expected W/\"2\", got %s", entry.IfMatch)
		}
	default:
		t.Fatal("expected log entry in channel but got none")
	}
}

func TestFHIRRequestLoggerMiddleware_OperationPath(t *testing.T) {
	sink := NewChannelLogSink(10)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/fhir/Patient/$validate", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := FHIRRequestLoggerMiddleware(sink)(func(c echo.Context) error {
		return c.String(http.StatusOK, `{"resourceType":"OperationOutcome"}`)
	})

	if err := handler(c); err != nil {
		t.Fatal(err)
	}

	select {
	case entry := <-sink.Entries():
		if entry.Interaction != "operation" {
			t.Errorf("expected operation, got %s", entry.Interaction)
		}
		if entry.Operation != "$validate" {
			t.Errorf("expected $validate, got %s", entry.Operation)
		}
		if entry.ResourceType != "Patient" {
			t.Errorf("expected Patient, got %s", entry.ResourceType)
		}
	default:
		t.Fatal("expected log entry in channel but got none")
	}
}
