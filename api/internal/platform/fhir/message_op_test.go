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

// newTestMessageBundle creates a valid FHIR Message Bundle for testing.
func newTestMessageBundle(eventCode, focusRef string, focusResource map[string]interface{}) map[string]interface{} {
	entries := []interface{}{
		map[string]interface{}{
			"fullUrl": "MessageHeader/1",
			"resource": map[string]interface{}{
				"resourceType": "MessageHeader",
				"id":           "1",
				"eventCoding": map[string]interface{}{
					"system": "http://example.org/events",
					"code":   eventCode,
				},
				"source": map[string]interface{}{"endpoint": "http://sender.example.org"},
				"focus":  []interface{}{map[string]interface{}{"reference": focusRef}},
			},
		},
	}

	if focusResource != nil {
		entries = append(entries, map[string]interface{}{
			"fullUrl":  focusRef,
			"resource": focusResource,
		})
	}

	return map[string]interface{}{
		"resourceType": "Bundle",
		"type":         "message",
		"entry":        entries,
	}
}

// =========== MessageProcessor Tests ===========

func TestProcessMessage_ValidNotification(t *testing.T) {
	p := NewMessageProcessor()

	bundle := newTestMessageBundle("notification", "Patient/123", map[string]interface{}{
		"resourceType": "Patient",
		"id":           "123",
		"name":         []interface{}{map[string]interface{}{"family": "Smith"}},
	})

	resp, err := p.ProcessMessage(context.Background(), bundle)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify response bundle type.
	if resp["type"] != "message" {
		t.Errorf("expected response type 'message', got %v", resp["type"])
	}

	// Verify response code is "ok".
	entries := resp["entry"].([]interface{})
	firstEntry := entries[0].(map[string]interface{})
	header := firstEntry["resource"].(map[string]interface{})
	response := header["response"].(map[string]interface{})

	if response["code"] != "ok" {
		t.Errorf("expected response code 'ok', got %v", response["code"])
	}
}

func TestProcessMessage_ResponseHasMessageHeader(t *testing.T) {
	p := NewMessageProcessor()

	bundle := newTestMessageBundle("notification", "Patient/123", map[string]interface{}{
		"resourceType": "Patient",
		"id":           "123",
		"name":         []interface{}{map[string]interface{}{"family": "Smith"}},
	})

	resp, err := p.ProcessMessage(context.Background(), bundle)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	entries := resp["entry"].([]interface{})
	if len(entries) == 0 {
		t.Fatal("expected at least one entry in response bundle")
	}

	firstEntry := entries[0].(map[string]interface{})
	header := firstEntry["resource"].(map[string]interface{})

	if header["resourceType"] != "MessageHeader" {
		t.Errorf("expected first entry to be MessageHeader, got %v", header["resourceType"])
	}
}

func TestProcessMessage_ResponseIdentifier(t *testing.T) {
	p := NewMessageProcessor()

	bundle := newTestMessageBundle("notification", "Patient/123", map[string]interface{}{
		"resourceType": "Patient",
		"id":           "123",
		"name":         []interface{}{map[string]interface{}{"family": "Smith"}},
	})

	resp, err := p.ProcessMessage(context.Background(), bundle)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	entries := resp["entry"].([]interface{})
	firstEntry := entries[0].(map[string]interface{})
	header := firstEntry["resource"].(map[string]interface{})
	response := header["response"].(map[string]interface{})

	if response["identifier"] != "1" {
		t.Errorf("expected response.identifier '1', got %v", response["identifier"])
	}
}

func TestProcessMessage_FocusResolution(t *testing.T) {
	p := NewMessageProcessor()

	// Register a custom handler that checks focus resources are passed correctly.
	var receivedFocus []map[string]interface{}
	p.RegisterHandler("test-event", func(_ context.Context, _ map[string]interface{}, focus []map[string]interface{}) ([]map[string]interface{}, error) {
		receivedFocus = focus
		return nil, nil
	})

	patient := map[string]interface{}{
		"resourceType": "Patient",
		"id":           "456",
		"name":         []interface{}{map[string]interface{}{"family": "Doe"}},
	}

	bundle := newTestMessageBundle("test-event", "Patient/456", patient)

	_, err := p.ProcessMessage(context.Background(), bundle)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(receivedFocus) != 1 {
		t.Fatalf("expected 1 focus resource, got %d", len(receivedFocus))
	}

	if receivedFocus[0]["resourceType"] != "Patient" {
		t.Errorf("expected focus resourceType 'Patient', got %v", receivedFocus[0]["resourceType"])
	}

	if receivedFocus[0]["id"] != "456" {
		t.Errorf("expected focus id '456', got %v", receivedFocus[0]["id"])
	}
}

func TestProcessMessage_UnknownEvent(t *testing.T) {
	p := NewMessageProcessor()

	bundle := newTestMessageBundle("unknown-event", "Patient/123", map[string]interface{}{
		"resourceType": "Patient",
		"id":           "123",
	})

	resp, err := p.ProcessMessage(context.Background(), bundle)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	entries := resp["entry"].([]interface{})
	firstEntry := entries[0].(map[string]interface{})
	header := firstEntry["resource"].(map[string]interface{})
	response := header["response"].(map[string]interface{})

	if response["code"] != "fatal-error" {
		t.Errorf("expected response code 'fatal-error' for unknown event, got %v", response["code"])
	}
}

func TestProcessMessage_NotMessageBundle(t *testing.T) {
	p := NewMessageProcessor()

	bundle := map[string]interface{}{
		"resourceType": "Bundle",
		"type":         "transaction",
		"entry":        []interface{}{},
	}

	_, err := p.ProcessMessage(context.Background(), bundle)
	if err == nil {
		t.Fatal("expected error for non-message bundle type")
	}

	if !strings.Contains(err.Error(), "message") {
		t.Errorf("expected error to mention 'message', got: %v", err)
	}
}

func TestProcessMessage_EmptyBundle(t *testing.T) {
	p := NewMessageProcessor()

	bundle := map[string]interface{}{
		"resourceType": "Bundle",
		"type":         "message",
	}

	_, err := p.ProcessMessage(context.Background(), bundle)
	if err == nil {
		t.Fatal("expected error for empty message bundle")
	}

	if !strings.Contains(err.Error(), "no entries") {
		t.Errorf("expected error to mention 'no entries', got: %v", err)
	}
}

func TestProcessMessage_FirstEntryNotMessageHeader(t *testing.T) {
	p := NewMessageProcessor()

	bundle := map[string]interface{}{
		"resourceType": "Bundle",
		"type":         "message",
		"entry": []interface{}{
			map[string]interface{}{
				"fullUrl": "Patient/123",
				"resource": map[string]interface{}{
					"resourceType": "Patient",
					"id":           "123",
				},
			},
		},
	}

	_, err := p.ProcessMessage(context.Background(), bundle)
	if err == nil {
		t.Fatal("expected error when first entry is not MessageHeader")
	}

	if !strings.Contains(err.Error(), "MessageHeader") {
		t.Errorf("expected error to mention 'MessageHeader', got: %v", err)
	}
}

func TestProcessMessage_PatientLinkEvent(t *testing.T) {
	p := NewMessageProcessor()

	// Valid patient-link with Patient focus.
	bundle := newTestMessageBundle("patient-link", "Patient/123", map[string]interface{}{
		"resourceType": "Patient",
		"id":           "123",
		"name":         []interface{}{map[string]interface{}{"family": "Smith"}},
	})

	resp, err := p.ProcessMessage(context.Background(), bundle)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	entries := resp["entry"].([]interface{})
	firstEntry := entries[0].(map[string]interface{})
	header := firstEntry["resource"].(map[string]interface{})
	response := header["response"].(map[string]interface{})

	if response["code"] != "ok" {
		t.Errorf("expected response code 'ok', got %v", response["code"])
	}

	// Invalid patient-link without Patient focus.
	bundleNoPatient := newTestMessageBundle("patient-link", "Observation/obs-1", map[string]interface{}{
		"resourceType": "Observation",
		"id":           "obs-1",
		"status":       "final",
	})

	resp2, err := p.ProcessMessage(context.Background(), bundleNoPatient)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	entries2 := resp2["entry"].([]interface{})
	firstEntry2 := entries2[0].(map[string]interface{})
	header2 := firstEntry2["resource"].(map[string]interface{})
	response2 := header2["response"].(map[string]interface{})

	if response2["code"] != "fatal-error" {
		t.Errorf("expected response code 'fatal-error' for patient-link without Patient, got %v", response2["code"])
	}
}

func TestProcessMessage_DiagnosticReportEvent(t *testing.T) {
	p := NewMessageProcessor()

	// Valid diagnostic-report with DiagnosticReport focus.
	bundle := newTestMessageBundle("diagnostic-report", "DiagnosticReport/dr-1", map[string]interface{}{
		"resourceType": "DiagnosticReport",
		"id":           "dr-1",
		"status":       "final",
		"code":         map[string]interface{}{"text": "CBC"},
	})

	resp, err := p.ProcessMessage(context.Background(), bundle)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	entries := resp["entry"].([]interface{})
	firstEntry := entries[0].(map[string]interface{})
	header := firstEntry["resource"].(map[string]interface{})
	response := header["response"].(map[string]interface{})

	if response["code"] != "ok" {
		t.Errorf("expected response code 'ok', got %v", response["code"])
	}

	// Invalid diagnostic-report without DiagnosticReport focus.
	bundleNoReport := newTestMessageBundle("diagnostic-report", "Patient/123", map[string]interface{}{
		"resourceType": "Patient",
		"id":           "123",
	})

	resp2, err := p.ProcessMessage(context.Background(), bundleNoReport)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	entries2 := resp2["entry"].([]interface{})
	firstEntry2 := entries2[0].(map[string]interface{})
	header2 := firstEntry2["resource"].(map[string]interface{})
	response2 := header2["response"].(map[string]interface{})

	if response2["code"] != "fatal-error" {
		t.Errorf("expected response code 'fatal-error' for diagnostic-report without DiagnosticReport, got %v", response2["code"])
	}
}

// =========== MessageHandler Tests ===========

func TestMessageHandler_POST_Valid(t *testing.T) {
	p := NewMessageProcessor()
	h := NewMessageHandler(p)
	e := echo.New()

	body := `{
		"resourceType": "Bundle",
		"type": "message",
		"entry": [
			{
				"fullUrl": "MessageHeader/1",
				"resource": {
					"resourceType": "MessageHeader",
					"id": "1",
					"eventCoding": {
						"system": "http://example.org/events",
						"code": "notification"
					},
					"source": {"endpoint": "http://sender.example.org"},
					"focus": [{"reference": "Patient/123"}]
				}
			},
			{
				"fullUrl": "Patient/123",
				"resource": {
					"resourceType": "Patient",
					"id": "123",
					"name": [{"family": "Smith"}]
				}
			}
		]
	}`

	req := httptest.NewRequest(http.MethodPost, "/fhir/$process-message", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.ProcessMessage(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if resp["resourceType"] != "Bundle" {
		t.Errorf("expected resourceType Bundle, got %v", resp["resourceType"])
	}
	if resp["type"] != "message" {
		t.Errorf("expected type message, got %v", resp["type"])
	}

	entries, ok := resp["entry"].([]interface{})
	if !ok || len(entries) == 0 {
		t.Fatal("expected at least one entry in response")
	}

	firstEntry := entries[0].(map[string]interface{})
	header := firstEntry["resource"].(map[string]interface{})

	if header["resourceType"] != "MessageHeader" {
		t.Errorf("expected first entry resource to be MessageHeader, got %v", header["resourceType"])
	}

	response := header["response"].(map[string]interface{})
	if response["code"] != "ok" {
		t.Errorf("expected response code 'ok', got %v", response["code"])
	}
}

func TestMessageHandler_POST_InvalidJSON(t *testing.T) {
	p := NewMessageProcessor()
	h := NewMessageHandler(p)
	e := echo.New()

	req := httptest.NewRequest(http.MethodPost, "/fhir/$process-message", strings.NewReader("{not valid json}"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.ProcessMessage(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}

	var outcome map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &outcome); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if outcome["resourceType"] != "OperationOutcome" {
		t.Errorf("expected OperationOutcome, got %v", outcome["resourceType"])
	}
}

func TestMessageHandler_POST_EmptyBody(t *testing.T) {
	p := NewMessageProcessor()
	h := NewMessageHandler(p)
	e := echo.New()

	req := httptest.NewRequest(http.MethodPost, "/fhir/$process-message", strings.NewReader(""))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.ProcessMessage(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}

	var outcome map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &outcome); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if outcome["resourceType"] != "OperationOutcome" {
		t.Errorf("expected OperationOutcome, got %v", outcome["resourceType"])
	}
}

func TestMessageHandler_POST_WrongBundleType(t *testing.T) {
	p := NewMessageProcessor()
	h := NewMessageHandler(p)
	e := echo.New()

	body := `{
		"resourceType": "Bundle",
		"type": "transaction",
		"entry": []
	}`

	req := httptest.NewRequest(http.MethodPost, "/fhir/$process-message", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.ProcessMessage(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}

	var outcome map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &outcome); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if outcome["resourceType"] != "OperationOutcome" {
		t.Errorf("expected OperationOutcome, got %v", outcome["resourceType"])
	}

	// Verify the issue mentions 'message'.
	issues, ok := outcome["issue"].([]interface{})
	if !ok || len(issues) == 0 {
		t.Fatal("expected issues in OperationOutcome")
	}
	issue := issues[0].(map[string]interface{})
	diag, _ := issue["diagnostics"].(string)
	if !strings.Contains(diag, "message") {
		t.Errorf("expected diagnostics to mention 'message', got: %s", diag)
	}
}

func TestMessageHandler_RegisterRoutes(t *testing.T) {
	p := NewMessageProcessor()
	h := NewMessageHandler(p)
	e := echo.New()
	fhirGroup := e.Group("/fhir")

	h.RegisterRoutes(fhirGroup)

	routes := e.Routes()
	routePaths := make(map[string]bool)
	for _, r := range routes {
		routePaths[r.Method+":"+r.Path] = true
	}

	expected := "POST:/fhir/$process-message"
	if !routePaths[expected] {
		t.Errorf("missing expected route: %s (registered: %v)", expected, routePaths)
	}
}
