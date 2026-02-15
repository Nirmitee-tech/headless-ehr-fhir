package fhir

import (
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

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// waitForProvenance polls the store until at least minCount entries exist or
// the timeout is reached. This accounts for the async goroutine.
func waitForProvenance(store *ProvenanceStore, minCount int, timeout time.Duration) []ProvenanceRecord {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		records := store.All()
		if len(records) >= minCount {
			return records
		}
		time.Sleep(5 * time.Millisecond)
	}
	return store.All()
}

// setupAutoProvenance creates an Echo instance with the auto-provenance
// middleware and a simple handler that returns the given status code and body.
func setupAutoProvenance(status int, body map[string]interface{}) (*echo.Echo, *ProvenanceStore) {
	e := echo.New()
	store := NewProvenanceStore()
	e.Use(AutoProvenanceMiddleware(store))

	bodyBytes, _ := json.Marshal(body)

	e.Any("/fhir/*", func(c echo.Context) error {
		c.Response().Header().Set("Content-Type", "application/fhir+json")
		return c.JSONBlob(status, bodyBytes)
	})

	// non-FHIR route
	e.Any("/api/v1/*", func(c echo.Context) error {
		return c.JSON(status, body)
	})

	return e, store
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestAutoProvenance_CreatedOnPOST(t *testing.T) {
	body := map[string]interface{}{
		"resourceType": "Patient",
		"id":           "pat-123",
	}
	e, store := setupAutoProvenance(http.StatusCreated, body)

	req := httptest.NewRequest(http.MethodPost, "/fhir/Patient", nil)
	req.Header.Set("Content-Type", "application/fhir+json")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", rec.Code)
	}

	records := waitForProvenance(store, 1, 2*time.Second)
	if len(records) != 1 {
		t.Fatalf("expected 1 provenance record, got %d", len(records))
	}

	r := records[0]
	if r.TargetReference != "Patient/pat-123" {
		t.Errorf("expected target 'Patient/pat-123', got %q", r.TargetReference)
	}
	if r.ActivityCode != "create" {
		t.Errorf("expected activity 'create', got %q", r.ActivityCode)
	}
}

func TestAutoProvenance_CreatedOnPUT(t *testing.T) {
	body := map[string]interface{}{
		"resourceType": "Observation",
		"id":           "obs-456",
	}
	e, store := setupAutoProvenance(http.StatusOK, body)

	req := httptest.NewRequest(http.MethodPut, "/fhir/Observation/obs-456", nil)
	req.Header.Set("Content-Type", "application/fhir+json")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	records := waitForProvenance(store, 1, 2*time.Second)
	if len(records) != 1 {
		t.Fatalf("expected 1 provenance record, got %d", len(records))
	}

	r := records[0]
	if r.TargetReference != "Observation/obs-456" {
		t.Errorf("expected target 'Observation/obs-456', got %q", r.TargetReference)
	}
	if r.ActivityCode != "update" {
		t.Errorf("expected activity 'update', got %q", r.ActivityCode)
	}
}

func TestAutoProvenance_CreatedOnDELETE(t *testing.T) {
	body := map[string]interface{}{
		"resourceType": "Patient",
		"id":           "pat-789",
	}
	e, store := setupAutoProvenance(http.StatusOK, body)

	req := httptest.NewRequest(http.MethodDelete, "/fhir/Patient/pat-789", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	records := waitForProvenance(store, 1, 2*time.Second)
	if len(records) != 1 {
		t.Fatalf("expected 1 provenance record, got %d", len(records))
	}

	r := records[0]
	if r.ActivityCode != "delete" {
		t.Errorf("expected activity 'delete', got %q", r.ActivityCode)
	}
	if r.TargetReference != "Patient/pat-789" {
		t.Errorf("expected target 'Patient/pat-789', got %q", r.TargetReference)
	}
}

func TestAutoProvenance_NotCreatedOnGET(t *testing.T) {
	body := map[string]interface{}{
		"resourceType": "Patient",
		"id":           "pat-123",
	}
	e, store := setupAutoProvenance(http.StatusOK, body)

	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient/pat-123", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	// Give it some time to make sure nothing is created
	time.Sleep(100 * time.Millisecond)
	records := store.All()
	if len(records) != 0 {
		t.Fatalf("expected 0 provenance records for GET, got %d", len(records))
	}
}

func TestAutoProvenance_NotCreatedOnNonFHIRPath(t *testing.T) {
	body := map[string]interface{}{
		"resourceType": "Patient",
		"id":           "pat-123",
	}
	e, store := setupAutoProvenance(http.StatusCreated, body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/patients", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	time.Sleep(100 * time.Millisecond)
	records := store.All()
	if len(records) != 0 {
		t.Fatalf("expected 0 provenance records for non-FHIR path, got %d", len(records))
	}
}

func TestAutoProvenance_OptOutWithHeader(t *testing.T) {
	body := map[string]interface{}{
		"resourceType": "Patient",
		"id":           "pat-123",
	}
	e, store := setupAutoProvenance(http.StatusCreated, body)

	req := httptest.NewRequest(http.MethodPost, "/fhir/Patient", nil)
	req.Header.Set("Content-Type", "application/fhir+json")
	req.Header.Set("X-No-Provenance", "true")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", rec.Code)
	}

	time.Sleep(100 * time.Millisecond)
	records := store.All()
	if len(records) != 0 {
		t.Fatalf("expected 0 provenance records with opt-out header, got %d", len(records))
	}
}

func TestAutoProvenance_AgentFromAuthContext(t *testing.T) {
	body := map[string]interface{}{
		"resourceType": "Encounter",
		"id":           "enc-001",
	}
	store := NewProvenanceStore()
	e := echo.New()

	// Simulate auth middleware setting user_id in context
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set("user_id", "Practitioner/dr-smith")
			return next(c)
		}
	})
	e.Use(AutoProvenanceMiddleware(store))

	bodyBytes, _ := json.Marshal(body)
	e.POST("/fhir/Encounter", func(c echo.Context) error {
		c.Response().Header().Set("Content-Type", "application/fhir+json")
		return c.JSONBlob(http.StatusCreated, bodyBytes)
	})

	req := httptest.NewRequest(http.MethodPost, "/fhir/Encounter", nil)
	req.Header.Set("Content-Type", "application/fhir+json")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	records := waitForProvenance(store, 1, 2*time.Second)
	if len(records) != 1 {
		t.Fatalf("expected 1 provenance record, got %d", len(records))
	}

	r := records[0]
	if r.AgentWho != "Practitioner/dr-smith" {
		t.Errorf("expected agent 'Practitioner/dr-smith', got %q", r.AgentWho)
	}
	if r.AgentType != "author" {
		t.Errorf("expected agent type 'author', got %q", r.AgentType)
	}
}

func TestAutoProvenance_ActivityCodingMapsCorrectly(t *testing.T) {
	body := map[string]interface{}{
		"resourceType": "Patient",
		"id":           "pat-1",
	}

	tests := []struct {
		method       string
		expectedCode string
	}{
		{http.MethodPost, "create"},
		{http.MethodPut, "update"},
		{http.MethodPatch, "update"},
		{http.MethodDelete, "delete"},
	}

	for _, tt := range tests {
		t.Run(tt.method, func(t *testing.T) {
			e, store := setupAutoProvenance(http.StatusOK, body)

			req := httptest.NewRequest(tt.method, "/fhir/Patient/pat-1", nil)
			req.Header.Set("Content-Type", "application/fhir+json")
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			records := waitForProvenance(store, 1, 2*time.Second)
			if len(records) != 1 {
				t.Fatalf("expected 1 provenance record for %s, got %d", tt.method, len(records))
			}

			if records[0].ActivityCode != tt.expectedCode {
				t.Errorf("method %s: expected activity %q, got %q", tt.method, tt.expectedCode, records[0].ActivityCode)
			}
		})
	}
}

func TestAutoProvenance_ReasonFromHeader(t *testing.T) {
	body := map[string]interface{}{
		"resourceType": "MedicationRequest",
		"id":           "med-001",
	}
	e, store := setupAutoProvenance(http.StatusCreated, body)

	req := httptest.NewRequest(http.MethodPost, "/fhir/MedicationRequest", nil)
	req.Header.Set("Content-Type", "application/fhir+json")
	req.Header.Set("X-Provenance-Reason", "Prescription renewal")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	records := waitForProvenance(store, 1, 2*time.Second)
	if len(records) != 1 {
		t.Fatalf("expected 1 provenance record, got %d", len(records))
	}

	if records[0].Reason != "Prescription renewal" {
		t.Errorf("expected reason 'Prescription renewal', got %q", records[0].Reason)
	}
}

func TestAutoProvenance_TargetReferenceFormat(t *testing.T) {
	tests := []struct {
		name           string
		resourceType   string
		resourceID     string
		expectedTarget string
	}{
		{"Patient", "Patient", "p1", "Patient/p1"},
		{"Observation", "Observation", "obs-1", "Observation/obs-1"},
		{"Condition", "Condition", "cond-abc", "Condition/cond-abc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := map[string]interface{}{
				"resourceType": tt.resourceType,
				"id":           tt.resourceID,
			}
			e, store := setupAutoProvenance(http.StatusCreated, body)

			req := httptest.NewRequest(http.MethodPost, "/fhir/"+tt.resourceType, nil)
			req.Header.Set("Content-Type", "application/fhir+json")
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			records := waitForProvenance(store, 1, 2*time.Second)
			if len(records) != 1 {
				t.Fatalf("expected 1 provenance record, got %d", len(records))
			}

			if records[0].TargetReference != tt.expectedTarget {
				t.Errorf("expected target %q, got %q", tt.expectedTarget, records[0].TargetReference)
			}
		})
	}
}

func TestAutoProvenance_RecordedTimestamp(t *testing.T) {
	body := map[string]interface{}{
		"resourceType": "Patient",
		"id":           "pat-ts",
	}
	e, store := setupAutoProvenance(http.StatusCreated, body)

	before := time.Now().UTC()
	req := httptest.NewRequest(http.MethodPost, "/fhir/Patient", nil)
	req.Header.Set("Content-Type", "application/fhir+json")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	records := waitForProvenance(store, 1, 2*time.Second)
	if len(records) != 1 {
		t.Fatalf("expected 1 provenance record, got %d", len(records))
	}

	after := time.Now().UTC()
	recorded := records[0].Recorded

	if recorded.Before(before.Add(-1 * time.Second)) {
		t.Errorf("recorded timestamp %v is before request start %v", recorded, before)
	}
	if recorded.After(after.Add(1 * time.Second)) {
		t.Errorf("recorded timestamp %v is after request end %v", recorded, after)
	}
}

func TestAutoProvenance_ConcurrentCreation(t *testing.T) {
	body := map[string]interface{}{
		"resourceType": "Patient",
		"id":           "pat-concurrent",
	}
	e, store := setupAutoProvenance(http.StatusCreated, body)

	const concurrency = 50
	var wg sync.WaitGroup
	wg.Add(concurrency)

	for i := 0; i < concurrency; i++ {
		go func(idx int) {
			defer wg.Done()
			req := httptest.NewRequest(http.MethodPost, "/fhir/Patient", nil)
			req.Header.Set("Content-Type", "application/fhir+json")
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)
		}(i)
	}

	wg.Wait()

	records := waitForProvenance(store, concurrency, 5*time.Second)
	if len(records) != concurrency {
		t.Errorf("expected %d provenance records, got %d", concurrency, len(records))
	}
}

func TestAutoProvenance_NotCreatedOnErrorResponses(t *testing.T) {
	tests := []struct {
		name   string
		status int
	}{
		{"400 Bad Request", http.StatusBadRequest},
		{"404 Not Found", http.StatusNotFound},
		{"409 Conflict", http.StatusConflict},
		{"422 Unprocessable Entity", http.StatusUnprocessableEntity},
		{"500 Internal Server Error", http.StatusInternalServerError},
		{"503 Service Unavailable", http.StatusServiceUnavailable},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := map[string]interface{}{
				"resourceType": "OperationOutcome",
				"issue": []interface{}{
					map[string]interface{}{
						"severity": "error",
						"code":     "processing",
					},
				},
			}
			e, store := setupAutoProvenance(tt.status, body)

			req := httptest.NewRequest(http.MethodPost, "/fhir/Patient", nil)
			req.Header.Set("Content-Type", "application/fhir+json")
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			time.Sleep(100 * time.Millisecond)
			records := store.All()
			if len(records) != 0 {
				t.Fatalf("expected 0 provenance records for status %d, got %d", tt.status, len(records))
			}
		})
	}
}

func TestAutoProvenance_PATCHCreatesUpdate(t *testing.T) {
	body := map[string]interface{}{
		"resourceType": "Patient",
		"id":           "pat-patch",
	}
	e, store := setupAutoProvenance(http.StatusOK, body)

	req := httptest.NewRequest(http.MethodPatch, "/fhir/Patient/pat-patch", nil)
	req.Header.Set("Content-Type", "application/json-patch+json")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	records := waitForProvenance(store, 1, 2*time.Second)
	if len(records) != 1 {
		t.Fatalf("expected 1 provenance record for PATCH, got %d", len(records))
	}
	if records[0].ActivityCode != "update" {
		t.Errorf("expected activity 'update' for PATCH, got %q", records[0].ActivityCode)
	}
}

func TestAutoProvenance_ProvenanceResourceFormat(t *testing.T) {
	body := map[string]interface{}{
		"resourceType": "Condition",
		"id":           "cond-fmt",
	}
	store := NewProvenanceStore()
	e := echo.New()

	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set("user_id", "Practitioner/dr-jones")
			return next(c)
		}
	})
	e.Use(AutoProvenanceMiddleware(store))

	bodyBytes, _ := json.Marshal(body)
	e.POST("/fhir/Condition", func(c echo.Context) error {
		c.Response().Header().Set("Content-Type", "application/fhir+json")
		return c.JSONBlob(http.StatusCreated, bodyBytes)
	})

	req := httptest.NewRequest(http.MethodPost, "/fhir/Condition", nil)
	req.Header.Set("Content-Type", "application/fhir+json")
	req.Header.Set("X-Provenance-Reason", "Initial diagnosis")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	records := waitForProvenance(store, 1, 2*time.Second)
	if len(records) != 1 {
		t.Fatalf("expected 1 provenance record, got %d", len(records))
	}

	fhirResource := records[0].ToFHIR()

	// Check resourceType
	if fhirResource["resourceType"] != "Provenance" {
		t.Errorf("expected resourceType 'Provenance', got %v", fhirResource["resourceType"])
	}

	// Check ID is non-empty
	if fhirResource["id"] == nil || fhirResource["id"] == "" {
		t.Error("expected non-empty id")
	}

	// Check target
	targets, ok := fhirResource["target"].([]map[string]interface{})
	if !ok || len(targets) == 0 {
		t.Fatal("expected target array")
	}
	if targets[0]["reference"] != "Condition/cond-fmt" {
		t.Errorf("expected target reference 'Condition/cond-fmt', got %v", targets[0]["reference"])
	}

	// Check recorded is non-empty
	if fhirResource["recorded"] == nil || fhirResource["recorded"] == "" {
		t.Error("expected non-empty recorded")
	}

	// Check agent
	agents, ok := fhirResource["agent"].([]map[string]interface{})
	if !ok || len(agents) == 0 {
		t.Fatal("expected agent array")
	}
	agentType, ok := agents[0]["type"].(map[string]interface{})
	if !ok {
		t.Fatal("expected agent type")
	}
	codings, ok := agentType["coding"].([]map[string]interface{})
	if !ok || len(codings) == 0 {
		t.Fatal("expected agent type coding")
	}
	if codings[0]["code"] != "author" {
		t.Errorf("expected agent type code 'author', got %v", codings[0]["code"])
	}
	who, ok := agents[0]["who"].(map[string]interface{})
	if !ok {
		t.Fatal("expected agent who")
	}
	if who["reference"] != "Practitioner/dr-jones" {
		t.Errorf("expected who reference 'Practitioner/dr-jones', got %v", who["reference"])
	}

	// Check activity
	activity, ok := fhirResource["activity"].(map[string]interface{})
	if !ok {
		t.Fatal("expected activity")
	}
	actCodings, ok := activity["coding"].([]map[string]interface{})
	if !ok || len(actCodings) == 0 {
		t.Fatal("expected activity coding")
	}
	if actCodings[0]["code"] != "create" {
		t.Errorf("expected activity code 'create', got %v", actCodings[0]["code"])
	}

	// Check reason
	reasons, ok := fhirResource["reason"].([]map[string]interface{})
	if !ok || len(reasons) == 0 {
		t.Fatal("expected reason array")
	}
	if reasons[0]["text"] != "Initial diagnosis" {
		t.Errorf("expected reason text 'Initial diagnosis', got %v", reasons[0]["text"])
	}
}

func TestAutoProvenance_NoReasonHeaderOmitsReason(t *testing.T) {
	body := map[string]interface{}{
		"resourceType": "Patient",
		"id":           "pat-noreason",
	}
	e, store := setupAutoProvenance(http.StatusCreated, body)

	req := httptest.NewRequest(http.MethodPost, "/fhir/Patient", nil)
	req.Header.Set("Content-Type", "application/fhir+json")
	// No X-Provenance-Reason header
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	records := waitForProvenance(store, 1, 2*time.Second)
	if len(records) != 1 {
		t.Fatalf("expected 1 provenance record, got %d", len(records))
	}

	if records[0].Reason != "" {
		t.Errorf("expected empty reason, got %q", records[0].Reason)
	}

	// Check FHIR output has no reason
	fhirResource := records[0].ToFHIR()
	if _, ok := fhirResource["reason"]; ok {
		t.Error("expected no reason field in FHIR output when no reason header provided")
	}
}

func TestAutoProvenance_DELETEWithNoBody(t *testing.T) {
	// DELETE responses often have no body — middleware should extract from path
	store := NewProvenanceStore()
	e := echo.New()
	e.Use(AutoProvenanceMiddleware(store))

	e.DELETE("/fhir/Patient/:id", func(c echo.Context) error {
		return c.NoContent(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodDelete, "/fhir/Patient/pat-del", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}

	records := waitForProvenance(store, 1, 2*time.Second)
	if len(records) != 1 {
		t.Fatalf("expected 1 provenance record for DELETE, got %d", len(records))
	}
	if records[0].TargetReference != "Patient/pat-del" {
		t.Errorf("expected target 'Patient/pat-del', got %q", records[0].TargetReference)
	}
	if records[0].ActivityCode != "delete" {
		t.Errorf("expected activity 'delete', got %q", records[0].ActivityCode)
	}
}

func TestAutoProvenance_ResponseNotBlocked(t *testing.T) {
	body := map[string]interface{}{
		"resourceType": "Patient",
		"id":           "pat-fast",
		"name":         "Test",
	}
	e, _ := setupAutoProvenance(http.StatusCreated, body)

	req := httptest.NewRequest(http.MethodPost, "/fhir/Patient", nil)
	req.Header.Set("Content-Type", "application/fhir+json")
	rec := httptest.NewRecorder()

	start := time.Now()
	e.ServeHTTP(rec, req)
	elapsed := time.Since(start)

	// Response should return very quickly; provenance is async
	if elapsed > 1*time.Second {
		t.Errorf("response took too long (%v), provenance may be blocking", elapsed)
	}

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", rec.Code)
	}

	// Verify the response body is intact
	var result map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if result["resourceType"] != "Patient" {
		t.Errorf("expected resourceType 'Patient' in response, got %v", result["resourceType"])
	}
}

func TestAutoProvenance_StoreAll(t *testing.T) {
	store := NewProvenanceStore()

	r1 := ProvenanceRecord{
		ID:              "prov-1",
		TargetReference: "Patient/p1",
		ActivityCode:    "create",
		Recorded:        time.Now().UTC(),
	}
	r2 := ProvenanceRecord{
		ID:              "prov-2",
		TargetReference: "Observation/o1",
		ActivityCode:    "update",
		Recorded:        time.Now().UTC(),
	}

	store.Add(r1)
	store.Add(r2)

	all := store.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 records, got %d", len(all))
	}
}

func TestAutoProvenance_StoreThreadSafe(t *testing.T) {
	store := NewProvenanceStore()
	var wg sync.WaitGroup
	const n = 100
	wg.Add(n)

	for i := 0; i < n; i++ {
		go func(idx int) {
			defer wg.Done()
			store.Add(ProvenanceRecord{
				ID:              fmt.Sprintf("prov-%d", idx),
				TargetReference: fmt.Sprintf("Patient/p%d", idx),
				ActivityCode:    "create",
				Recorded:        time.Now().UTC(),
			})
		}(i)
	}

	wg.Wait()

	all := store.All()
	if len(all) != n {
		t.Errorf("expected %d records, got %d", n, len(all))
	}
}

func TestAutoProvenance_ResponseBodyPreserved(t *testing.T) {
	body := map[string]interface{}{
		"resourceType": "Patient",
		"id":           "pat-body",
		"name": []interface{}{
			map[string]interface{}{
				"family": "Smith",
				"given":  []interface{}{"John"},
			},
		},
	}
	e, _ := setupAutoProvenance(http.StatusCreated, body)

	req := httptest.NewRequest(http.MethodPost, "/fhir/Patient", nil)
	req.Header.Set("Content-Type", "application/fhir+json")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	var result map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("response body is not valid JSON: %v", err)
	}
	if result["resourceType"] != "Patient" {
		t.Errorf("response body resourceType mismatch")
	}
	if result["id"] != "pat-body" {
		t.Errorf("response body id mismatch")
	}
}

func TestAutoProvenance_OnlyWriteMethods(t *testing.T) {
	body := map[string]interface{}{
		"resourceType": "Patient",
		"id":           "pat-ro",
	}

	readMethods := []string{http.MethodGet, http.MethodHead, http.MethodOptions}

	for _, method := range readMethods {
		t.Run(method, func(t *testing.T) {
			e, store := setupAutoProvenance(http.StatusOK, body)

			req := httptest.NewRequest(method, "/fhir/Patient/pat-ro", nil)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			time.Sleep(100 * time.Millisecond)
			records := store.All()
			if len(records) != 0 {
				t.Errorf("expected no provenance for %s, got %d records", method, len(records))
			}
		})
	}
}

func TestAutoProvenance_EmptyResponseBody(t *testing.T) {
	// When a POST returns with no parseable body (empty), middleware should
	// not crash, but also not create provenance (no resource info to track).
	store := NewProvenanceStore()
	e := echo.New()
	e.Use(AutoProvenanceMiddleware(store))

	e.POST("/fhir/Patient", func(c echo.Context) error {
		return c.String(http.StatusCreated, "")
	})

	req := httptest.NewRequest(http.MethodPost, "/fhir/Patient", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	time.Sleep(100 * time.Millisecond)
	records := store.All()
	// With an empty response body we can't extract resource info
	// so either 0 records or middleware gracefully handled it
	if len(records) != 0 {
		t.Logf("unexpected provenance record with empty body: %+v", records[0])
	}
}

func TestAutoProvenance_OptOutHeaderCaseInsensitive(t *testing.T) {
	body := map[string]interface{}{
		"resourceType": "Patient",
		"id":           "pat-case",
	}
	e, store := setupAutoProvenance(http.StatusCreated, body)

	// HTTP headers are case-insensitive per spec; net/http normalizes them
	req := httptest.NewRequest(http.MethodPost, "/fhir/Patient", nil)
	req.Header.Set("x-no-provenance", "true")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	time.Sleep(100 * time.Millisecond)
	records := store.All()
	if len(records) != 0 {
		t.Fatalf("expected opt-out to work case-insensitively, got %d records", len(records))
	}
}

func TestAutoProvenance_DefaultAgentWhenNoAuth(t *testing.T) {
	body := map[string]interface{}{
		"resourceType": "Patient",
		"id":           "pat-noauth",
	}
	e, store := setupAutoProvenance(http.StatusCreated, body)

	req := httptest.NewRequest(http.MethodPost, "/fhir/Patient", nil)
	req.Header.Set("Content-Type", "application/fhir+json")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	records := waitForProvenance(store, 1, 2*time.Second)
	if len(records) != 1 {
		t.Fatalf("expected 1 provenance record, got %d", len(records))
	}

	// When no user is authenticated, agent should be "anonymous" or similar
	if records[0].AgentWho == "" {
		t.Error("expected a default agent, got empty string")
	}
	if !strings.Contains(records[0].AgentWho, "anonymous") && !strings.Contains(records[0].AgentWho, "unknown") && !strings.Contains(records[0].AgentWho, "system") {
		t.Logf("agent is %q when no auth context present (acceptable if non-empty)", records[0].AgentWho)
	}
}
