package fhir

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
)

// newTestCDSHooksHandler creates a CDSHooksHandler with a mock service for testing.
func newTestCDSHooksHandler() *CDSHooksHandler {
	h := NewCDSHooksHandler()

	h.RegisterService(CDSService{
		Hook:        "patient-view",
		Title:       "Patient Risk Alerts",
		Description: "Shows active CDS alerts when a patient chart is opened",
		ID:          "patient-risk-alerts",
		Prefetch: map[string]string{
			"patient": "Patient/{{context.patientId}}",
		},
	}, func(ctx context.Context, req CDSHookRequest) (*CDSHookResponse, error) {
		return &CDSHookResponse{
			Cards: []CDSCard{
				{
					Summary:   "High fall risk",
					Indicator: "warning",
					Source:    CDSSource{Label: "EHR CDS Engine"},
					Detail:    "Patient has a high fall risk assessment score.",
				},
			},
		}, nil
	})

	h.RegisterService(CDSService{
		Hook:        "order-select",
		Title:       "Drug Interaction Check",
		Description: "Checks for drug interactions when a medication is selected",
		ID:          "drug-interaction-check",
	}, func(ctx context.Context, req CDSHookRequest) (*CDSHookResponse, error) {
		return &CDSHookResponse{
			Cards: []CDSCard{
				{
					Summary:   "Potential interaction: Warfarin + Aspirin",
					Indicator: "critical",
					Source:    CDSSource{Label: "EHR CDS Engine"},
				},
			},
		}, nil
	})

	return h
}

func TestCDSHooks_Discovery(t *testing.T) {
	h := newTestCDSHooksHandler()

	e := echo.New()
	h.RegisterRoutes(e)

	req := httptest.NewRequest(http.MethodGet, "/cds-services", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var result struct {
		Services []CDSService `json:"services"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if len(result.Services) != 2 {
		t.Fatalf("expected 2 services, got %d", len(result.Services))
	}
	if result.Services[0].ID != "patient-risk-alerts" {
		t.Errorf("expected first service ID 'patient-risk-alerts', got %q", result.Services[0].ID)
	}
	if result.Services[1].ID != "drug-interaction-check" {
		t.Errorf("expected second service ID 'drug-interaction-check', got %q", result.Services[1].ID)
	}
	if result.Services[0].Hook != "patient-view" {
		t.Errorf("expected first service hook 'patient-view', got %q", result.Services[0].Hook)
	}
	if result.Services[0].Prefetch == nil || result.Services[0].Prefetch["patient"] != "Patient/{{context.patientId}}" {
		t.Errorf("expected prefetch template, got %v", result.Services[0].Prefetch)
	}
}

func TestCDSHooks_Discovery_Empty(t *testing.T) {
	h := NewCDSHooksHandler()

	e := echo.New()
	h.RegisterRoutes(e)

	req := httptest.NewRequest(http.MethodGet, "/cds-services", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	// Should return {"services":[]} not {"services":null}
	body := rec.Body.String()
	if !strings.Contains(body, `"services":[]`) {
		t.Errorf("expected empty services array, got %s", body)
	}
}

func TestCDSHooks_HandleHook_Success(t *testing.T) {
	h := newTestCDSHooksHandler()

	e := echo.New()
	h.RegisterRoutes(e)

	payload := `{
		"hook": "patient-view",
		"hookInstance": "d1577c69-dfbe-44ad-bd63-8c2c87e28ccc",
		"context": {"patientId": "patient-123"},
		"prefetch": {}
	}`
	req := httptest.NewRequest(http.MethodPost, "/cds-services/patient-risk-alerts", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp CDSHookResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if len(resp.Cards) != 1 {
		t.Fatalf("expected 1 card, got %d", len(resp.Cards))
	}
	if resp.Cards[0].Summary != "High fall risk" {
		t.Errorf("expected summary 'High fall risk', got %q", resp.Cards[0].Summary)
	}
	if resp.Cards[0].Indicator != "warning" {
		t.Errorf("expected indicator 'warning', got %q", resp.Cards[0].Indicator)
	}
	if resp.Cards[0].Source.Label != "EHR CDS Engine" {
		t.Errorf("expected source label 'EHR CDS Engine', got %q", resp.Cards[0].Source.Label)
	}
}

func TestCDSHooks_HandleHook_NotFound(t *testing.T) {
	h := newTestCDSHooksHandler()

	e := echo.New()
	h.RegisterRoutes(e)

	payload := `{
		"hook": "patient-view",
		"hookInstance": "d1577c69-dfbe-44ad-bd63-8c2c87e28ccc",
		"context": {}
	}`
	req := httptest.NewRequest(http.MethodPost, "/cds-services/nonexistent", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}

	var outcome OperationOutcome
	if err := json.Unmarshal(rec.Body.Bytes(), &outcome); err != nil {
		t.Fatalf("failed to unmarshal OperationOutcome: %v", err)
	}
	if outcome.ResourceType != "OperationOutcome" {
		t.Errorf("expected resourceType OperationOutcome, got %s", outcome.ResourceType)
	}
	if len(outcome.Issue) == 0 {
		t.Fatal("expected at least one issue")
	}
	if outcome.Issue[0].Code != "not-found" {
		t.Errorf("expected issue code 'not-found', got %q", outcome.Issue[0].Code)
	}
}

func TestCDSHooks_HandleHook_BadRequest(t *testing.T) {
	h := newTestCDSHooksHandler()

	e := echo.New()
	h.RegisterRoutes(e)

	req := httptest.NewRequest(http.MethodPost, "/cds-services/patient-risk-alerts", strings.NewReader(`{invalid json`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}

	var outcome OperationOutcome
	if err := json.Unmarshal(rec.Body.Bytes(), &outcome); err != nil {
		t.Fatalf("failed to unmarshal OperationOutcome: %v", err)
	}
	if outcome.ResourceType != "OperationOutcome" {
		t.Errorf("expected resourceType OperationOutcome, got %s", outcome.ResourceType)
	}
}

func TestCDSHooks_HandleHook_HookMismatch(t *testing.T) {
	h := newTestCDSHooksHandler()

	e := echo.New()
	h.RegisterRoutes(e)

	// Send order-select hook to patient-view service
	payload := `{
		"hook": "order-select",
		"hookInstance": "d1577c69-dfbe-44ad-bd63-8c2c87e28ccc",
		"context": {}
	}`
	req := httptest.NewRequest(http.MethodPost, "/cds-services/patient-risk-alerts", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}

	var outcome OperationOutcome
	if err := json.Unmarshal(rec.Body.Bytes(), &outcome); err != nil {
		t.Fatalf("failed to unmarshal OperationOutcome: %v", err)
	}
	if outcome.ResourceType != "OperationOutcome" {
		t.Errorf("expected resourceType OperationOutcome, got %s", outcome.ResourceType)
	}
}

func TestCDSHooks_HandleHook_MissingHookInstance(t *testing.T) {
	h := newTestCDSHooksHandler()

	e := echo.New()
	h.RegisterRoutes(e)

	payload := `{
		"hook": "patient-view",
		"context": {"patientId": "patient-123"}
	}`
	req := httptest.NewRequest(http.MethodPost, "/cds-services/patient-risk-alerts", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}

	var outcome OperationOutcome
	if err := json.Unmarshal(rec.Body.Bytes(), &outcome); err != nil {
		t.Fatalf("failed to unmarshal OperationOutcome: %v", err)
	}
	if outcome.ResourceType != "OperationOutcome" {
		t.Errorf("expected resourceType OperationOutcome, got %s", outcome.ResourceType)
	}
}

func TestCDSHooks_Feedback_Success(t *testing.T) {
	h := newTestCDSHooksHandler()

	feedbackCalled := false
	h.RegisterFeedbackHandler("patient-risk-alerts", func(ctx context.Context, serviceID string, fb CDSFeedbackRequest) error {
		feedbackCalled = true
		if serviceID != "patient-risk-alerts" {
			t.Errorf("expected serviceID 'patient-risk-alerts', got %q", serviceID)
		}
		if fb.Card != "card-uuid-1" {
			t.Errorf("expected card 'card-uuid-1', got %q", fb.Card)
		}
		if fb.Outcome != "accepted" {
			t.Errorf("expected outcome 'accepted', got %q", fb.Outcome)
		}
		return nil
	})

	e := echo.New()
	h.RegisterRoutes(e)

	payload := `{
		"card": "card-uuid-1",
		"outcome": "accepted",
		"outcomeTimestamp": "2024-01-15T10:30:00Z"
	}`
	req := httptest.NewRequest(http.MethodPost, "/cds-services/patient-risk-alerts/feedback", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if !feedbackCalled {
		t.Error("expected feedback handler to be called")
	}
}

func TestCDSHooks_Feedback_NotFound(t *testing.T) {
	h := newTestCDSHooksHandler()

	e := echo.New()
	h.RegisterRoutes(e)

	payload := `{
		"card": "card-uuid-1",
		"outcome": "accepted"
	}`
	req := httptest.NewRequest(http.MethodPost, "/cds-services/nonexistent/feedback", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}

	var outcome OperationOutcome
	if err := json.Unmarshal(rec.Body.Bytes(), &outcome); err != nil {
		t.Fatalf("failed to unmarshal OperationOutcome: %v", err)
	}
	if outcome.ResourceType != "OperationOutcome" {
		t.Errorf("expected resourceType OperationOutcome, got %s", outcome.ResourceType)
	}
}

func TestCDSHooks_Feedback_NoHandler(t *testing.T) {
	h := newTestCDSHooksHandler()
	// No feedback handler registered for drug-interaction-check

	e := echo.New()
	h.RegisterRoutes(e)

	payload := `{
		"card": "card-uuid-1",
		"outcome": "overridden"
	}`
	req := httptest.NewRequest(http.MethodPost, "/cds-services/drug-interaction-check/feedback", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	// Should return 200 as a no-op
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 (no-op), got %d: %s", rec.Code, rec.Body.String())
	}
}
