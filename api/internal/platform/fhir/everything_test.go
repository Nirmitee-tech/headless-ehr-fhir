package fhir

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
)

// newEverythingTestHandler creates an EverythingHandler with a mock patient
// fetcher that returns a patient for "patient-123" and nil for anything else.
func newEverythingTestHandler() *EverythingHandler {
	h := NewEverythingHandler()
	h.SetPatientFetcher(func(ctx context.Context, fhirID string) (map[string]interface{}, error) {
		if fhirID == "patient-123" {
			return map[string]interface{}{
				"resourceType": "Patient",
				"id":           "patient-123",
				"name":         []interface{}{map[string]interface{}{"family": "Smith"}},
			}, nil
		}
		return nil, fmt.Errorf("not found")
	})
	return h
}

func TestEverything_Success(t *testing.T) {
	h := newEverythingTestHandler()

	// Register two mock fetchers
	h.RegisterFetcher("Condition", func(ctx context.Context, patientID string) ([]map[string]interface{}, error) {
		return []map[string]interface{}{
			{"resourceType": "Condition", "id": "cond-1", "subject": map[string]interface{}{"reference": "Patient/" + patientID}},
			{"resourceType": "Condition", "id": "cond-2", "subject": map[string]interface{}{"reference": "Patient/" + patientID}},
		}, nil
	})
	h.RegisterFetcher("Observation", func(ctx context.Context, patientID string) ([]map[string]interface{}, error) {
		return []map[string]interface{}{
			{"resourceType": "Observation", "id": "obs-1", "subject": map[string]interface{}{"reference": "Patient/" + patientID}},
		}, nil
	})

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient/patient-123/$everything", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("patient-123")

	err := h.Handle(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var bundle Bundle
	if err := json.Unmarshal(rec.Body.Bytes(), &bundle); err != nil {
		t.Fatalf("failed to unmarshal bundle: %v", err)
	}

	if bundle.ResourceType != "Bundle" {
		t.Errorf("expected resourceType Bundle, got %s", bundle.ResourceType)
	}
	if bundle.Type != "searchset" {
		t.Errorf("expected type searchset, got %s", bundle.Type)
	}
	// 1 Patient + 2 Conditions + 1 Observation = 4
	if *bundle.Total != 4 {
		t.Errorf("expected total 4, got %d", *bundle.Total)
	}
	if len(bundle.Entry) != 4 {
		t.Fatalf("expected 4 entries, got %d", len(bundle.Entry))
	}

	// First entry should be Patient
	var firstResource map[string]interface{}
	if err := json.Unmarshal(bundle.Entry[0].Resource, &firstResource); err != nil {
		t.Fatalf("failed to unmarshal first entry: %v", err)
	}
	if firstResource["resourceType"] != "Patient" {
		t.Errorf("expected first entry to be Patient, got %s", firstResource["resourceType"])
	}
	if bundle.Entry[0].FullURL != "Patient/patient-123" {
		t.Errorf("expected fullUrl Patient/patient-123, got %s", bundle.Entry[0].FullURL)
	}
	if bundle.Entry[0].Search == nil || bundle.Entry[0].Search.Mode != "match" {
		t.Error("expected search mode 'match' on first entry")
	}
}

func TestEverything_TypeFilter(t *testing.T) {
	h := newEverythingTestHandler()

	h.RegisterFetcher("Condition", func(ctx context.Context, patientID string) ([]map[string]interface{}, error) {
		return []map[string]interface{}{
			{"resourceType": "Condition", "id": "cond-1"},
		}, nil
	})
	h.RegisterFetcher("Observation", func(ctx context.Context, patientID string) ([]map[string]interface{}, error) {
		return []map[string]interface{}{
			{"resourceType": "Observation", "id": "obs-1"},
		}, nil
	})

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient/patient-123/$everything?_type=Condition", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("patient-123")

	err := h.Handle(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var bundle Bundle
	if err := json.Unmarshal(rec.Body.Bytes(), &bundle); err != nil {
		t.Fatalf("failed to unmarshal bundle: %v", err)
	}

	// 1 Patient + 1 Condition = 2 (Observation filtered out)
	if *bundle.Total != 2 {
		t.Errorf("expected total 2, got %d", *bundle.Total)
	}
	if len(bundle.Entry) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(bundle.Entry))
	}

	// Verify no Observation entry
	for _, entry := range bundle.Entry {
		var r map[string]interface{}
		json.Unmarshal(entry.Resource, &r)
		if r["resourceType"] == "Observation" {
			t.Error("Observation should have been filtered out by _type=Condition")
		}
	}
}

func TestEverything_EmptyResult(t *testing.T) {
	h := newEverythingTestHandler()

	// Register fetchers that return empty slices
	h.RegisterFetcher("Condition", func(ctx context.Context, patientID string) ([]map[string]interface{}, error) {
		return nil, nil
	})

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient/patient-123/$everything", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("patient-123")

	err := h.Handle(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var bundle Bundle
	if err := json.Unmarshal(rec.Body.Bytes(), &bundle); err != nil {
		t.Fatalf("failed to unmarshal bundle: %v", err)
	}

	// Only the Patient resource should be present
	if *bundle.Total != 1 {
		t.Errorf("expected total 1, got %d", *bundle.Total)
	}
	if len(bundle.Entry) != 1 {
		t.Fatalf("expected 1 entry (Patient only), got %d", len(bundle.Entry))
	}
}

func TestEverything_PatientNotFound(t *testing.T) {
	h := newEverythingTestHandler()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient/nonexistent/$everything", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("nonexistent")

	err := h.Handle(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}

	var outcome OperationOutcome
	if err := json.Unmarshal(rec.Body.Bytes(), &outcome); err != nil {
		t.Fatalf("failed to unmarshal OperationOutcome: %v", err)
	}
	if outcome.ResourceType != "OperationOutcome" {
		t.Errorf("expected resourceType OperationOutcome, got %s", outcome.ResourceType)
	}
	if len(outcome.Issue) == 0 {
		t.Error("expected at least one issue")
	}
	if outcome.Issue[0].Code != "not-found" {
		t.Errorf("expected issue code not-found, got %s", outcome.Issue[0].Code)
	}
}

func TestEverything_CountLimit(t *testing.T) {
	h := newEverythingTestHandler()

	h.RegisterFetcher("Condition", func(ctx context.Context, patientID string) ([]map[string]interface{}, error) {
		return []map[string]interface{}{
			{"resourceType": "Condition", "id": "cond-1"},
			{"resourceType": "Condition", "id": "cond-2"},
			{"resourceType": "Condition", "id": "cond-3"},
			{"resourceType": "Condition", "id": "cond-4"},
			{"resourceType": "Condition", "id": "cond-5"},
		}, nil
	})
	h.RegisterFetcher("Observation", func(ctx context.Context, patientID string) ([]map[string]interface{}, error) {
		return []map[string]interface{}{
			{"resourceType": "Observation", "id": "obs-1"},
			{"resourceType": "Observation", "id": "obs-2"},
			{"resourceType": "Observation", "id": "obs-3"},
		}, nil
	})

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient/patient-123/$everything?_count=2", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("patient-123")

	err := h.Handle(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var bundle Bundle
	if err := json.Unmarshal(rec.Body.Bytes(), &bundle); err != nil {
		t.Fatalf("failed to unmarshal bundle: %v", err)
	}

	// 1 Patient + 2 Conditions (limited) + 2 Observations (limited) = 5
	if *bundle.Total != 5 {
		t.Errorf("expected total 5, got %d", *bundle.Total)
	}
	if len(bundle.Entry) != 5 {
		t.Fatalf("expected 5 entries, got %d", len(bundle.Entry))
	}

	// Verify per-type limits: count Conditions and Observations
	condCount := 0
	obsCount := 0
	for _, entry := range bundle.Entry {
		var r map[string]interface{}
		json.Unmarshal(entry.Resource, &r)
		switch r["resourceType"] {
		case "Condition":
			condCount++
		case "Observation":
			obsCount++
		}
	}
	if condCount != 2 {
		t.Errorf("expected 2 Conditions (limited by _count), got %d", condCount)
	}
	if obsCount != 2 {
		t.Errorf("expected 2 Observations (limited by _count), got %d", obsCount)
	}
}
