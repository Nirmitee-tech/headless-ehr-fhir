package webhook

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
)

// helper: create a WebhookManager with in-memory store and optional http client override.
func newTestManager(client *http.Client) *WebhookManager {
	store := NewInMemoryWebhookStore()
	opts := []ManagerOption{}
	if client != nil {
		opts = append(opts, WithHTTPClient(client))
	}
	return NewWebhookManager(store, opts...)
}

// helper: create an active endpoint in the manager.
func mustRegisterEndpoint(t *testing.T, m *WebhookManager, url, tenantID string, events []string) *WebhookEndpoint {
	t.Helper()
	ep, err := m.RegisterEndpoint(context.Background(), url, "test-secret-key", tenantID, "client-1", events)
	if err != nil {
		t.Fatalf("failed to register endpoint: %v", err)
	}
	return ep
}

// ===================== Endpoint Management =====================

func TestWebhookManager_RegisterEndpoint(t *testing.T) {
	m := newTestManager(nil)
	ep, err := m.RegisterEndpoint(context.Background(), "https://example.com/hook", "my-secret", "tenant-1", "client-1", []string{"Patient.create"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ep.ID == "" {
		t.Error("expected ID to be set")
	}
	if ep.URL != "https://example.com/hook" {
		t.Errorf("expected URL 'https://example.com/hook', got %q", ep.URL)
	}
	if ep.Secret != "my-secret" {
		t.Errorf("expected secret 'my-secret', got %q", ep.Secret)
	}
	if ep.Status != "active" {
		t.Errorf("expected status 'active', got %q", ep.Status)
	}
	if ep.TenantID != "tenant-1" {
		t.Errorf("expected tenant 'tenant-1', got %q", ep.TenantID)
	}
	if len(ep.Events) != 1 || ep.Events[0] != "Patient.create" {
		t.Errorf("unexpected events: %v", ep.Events)
	}
	if ep.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}
}

func TestWebhookManager_RegisterEndpoint_GeneratesSecret(t *testing.T) {
	m := newTestManager(nil)
	ep, err := m.RegisterEndpoint(context.Background(), "https://example.com/hook", "", "tenant-1", "client-1", []string{"Patient.create"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ep.Secret == "" {
		t.Error("expected auto-generated secret")
	}
	if len(ep.Secret) < 32 {
		t.Errorf("expected secret at least 32 chars, got %d", len(ep.Secret))
	}
}

func TestWebhookManager_RegisterEndpoint_ValidatesURL(t *testing.T) {
	m := newTestManager(nil)
	tests := []struct {
		name string
		url  string
	}{
		{"empty", ""},
		{"no scheme", "example.com/hook"},
		{"ftp scheme", "ftp://example.com/hook"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := m.RegisterEndpoint(context.Background(), tt.url, "secret", "tenant-1", "client-1", []string{"Patient.create"})
			if err == nil {
				t.Errorf("expected error for URL %q", tt.url)
			}
		})
	}
}

func TestWebhookManager_ListEndpoints(t *testing.T) {
	m := newTestManager(nil)
	mustRegisterEndpoint(t, m, "https://example.com/hook1", "tenant-1", []string{"Patient.create"})
	mustRegisterEndpoint(t, m, "https://example.com/hook2", "tenant-1", []string{"Patient.update"})
	mustRegisterEndpoint(t, m, "https://example.com/hook3", "tenant-2", []string{"Patient.delete"})

	eps, total, err := m.store.ListEndpoints(context.Background(), "tenant-1", 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 2 {
		t.Errorf("expected 2 endpoints for tenant-1, got %d", total)
	}
	if len(eps) != 2 {
		t.Errorf("expected 2 endpoints, got %d", len(eps))
	}
}

func TestWebhookManager_PauseEndpoint(t *testing.T) {
	m := newTestManager(nil)
	ep := mustRegisterEndpoint(t, m, "https://example.com/hook", "tenant-1", []string{"Patient.create"})

	if err := m.PauseEndpoint(context.Background(), ep.ID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := m.store.GetEndpoint(context.Background(), ep.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Status != "paused" {
		t.Errorf("expected status 'paused', got %q", got.Status)
	}
}

func TestWebhookManager_ResumeEndpoint(t *testing.T) {
	m := newTestManager(nil)
	ep := mustRegisterEndpoint(t, m, "https://example.com/hook", "tenant-1", []string{"Patient.create"})
	m.PauseEndpoint(context.Background(), ep.ID)

	if err := m.ResumeEndpoint(context.Background(), ep.ID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := m.store.GetEndpoint(context.Background(), ep.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Status != "active" {
		t.Errorf("expected status 'active', got %q", got.Status)
	}
}

func TestWebhookManager_DeleteEndpoint(t *testing.T) {
	m := newTestManager(nil)
	ep := mustRegisterEndpoint(t, m, "https://example.com/hook", "tenant-1", []string{"Patient.create"})

	if err := m.store.DeleteEndpoint(context.Background(), ep.ID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err := m.store.GetEndpoint(context.Background(), ep.ID)
	if err == nil {
		t.Error("expected error after delete")
	}
}

// ===================== Signature =====================

func TestSignPayload(t *testing.T) {
	payload := []byte(`{"type":"Patient.create","id":"123"}`)
	sig1 := SignPayload(payload, "secret-key")
	sig2 := SignPayload(payload, "secret-key")
	if sig1 != sig2 {
		t.Error("expected deterministic signatures")
	}
	if sig1 == "" {
		t.Error("expected non-empty signature")
	}
}

func TestVerifySignature(t *testing.T) {
	payload := []byte(`{"type":"Patient.create","id":"123"}`)
	sig := SignPayload(payload, "secret-key")
	if !VerifySignature(payload, "secret-key", sig) {
		t.Error("expected valid signature to verify")
	}
}

func TestVerifySignature_Invalid(t *testing.T) {
	payload := []byte(`{"type":"Patient.create","id":"123"}`)
	if VerifySignature(payload, "secret-key", "invalid-sig") {
		t.Error("expected invalid signature to fail verification")
	}
}

func TestVerifySignature_WrongSecret(t *testing.T) {
	payload := []byte(`{"type":"Patient.create","id":"123"}`)
	sig := SignPayload(payload, "secret-key")
	if VerifySignature(payload, "wrong-secret", sig) {
		t.Error("expected wrong secret to fail verification")
	}
}

// ===================== Delivery =====================

func TestWebhookManager_Deliver(t *testing.T) {
	var receivedBody []byte
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok":true}`))
	}))
	defer ts.Close()

	m := newTestManager(ts.Client())
	mustRegisterEndpoint(t, m, ts.URL+"/hook", "tenant-1", []string{"Patient.create"})

	event := WebhookEvent{
		ID:           "evt-1",
		Type:         "Patient.create",
		ResourceType: "Patient",
		ResourceID:   "p-123",
		TenantID:     "tenant-1",
		Payload:      json.RawMessage(`{"resourceType":"Patient","id":"p-123"}`),
		Timestamp:    time.Now(),
	}

	results := m.Deliver(context.Background(), event)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if !results[0].Success {
		t.Errorf("expected success, got error: %s", results[0].Error)
	}
	if results[0].StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", results[0].StatusCode)
	}
	if len(receivedBody) == 0 {
		t.Error("expected server to receive payload")
	}
}

func TestWebhookManager_Deliver_EventFiltering(t *testing.T) {
	callCount := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	m := newTestManager(ts.Client())
	mustRegisterEndpoint(t, m, ts.URL+"/hook", "tenant-1", []string{"Patient.create"})

	event := WebhookEvent{
		ID:           "evt-1",
		Type:         "Encounter.update",
		ResourceType: "Encounter",
		ResourceID:   "e-123",
		TenantID:     "tenant-1",
		Payload:      json.RawMessage(`{}`),
		Timestamp:    time.Now(),
	}

	results := m.Deliver(context.Background(), event)
	if len(results) != 0 {
		t.Errorf("expected 0 results (no matching endpoints), got %d", len(results))
	}
	if callCount != 0 {
		t.Errorf("expected 0 calls, got %d", callCount)
	}
}

func TestWebhookManager_Deliver_WildcardEvent(t *testing.T) {
	callCount := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	m := newTestManager(ts.Client())
	mustRegisterEndpoint(t, m, ts.URL+"/hook", "tenant-1", []string{"*.delete"})

	// Should match
	event1 := WebhookEvent{
		ID: "evt-1", Type: "Patient.delete", ResourceType: "Patient",
		ResourceID: "p-1", TenantID: "tenant-1", Payload: json.RawMessage(`{}`), Timestamp: time.Now(),
	}
	results := m.Deliver(context.Background(), event1)
	if len(results) != 1 || !results[0].Success {
		t.Error("expected wildcard to match Patient.delete")
	}

	// Should also match
	event2 := WebhookEvent{
		ID: "evt-2", Type: "Encounter.delete", ResourceType: "Encounter",
		ResourceID: "e-1", TenantID: "tenant-1", Payload: json.RawMessage(`{}`), Timestamp: time.Now(),
	}
	results = m.Deliver(context.Background(), event2)
	if len(results) != 1 || !results[0].Success {
		t.Error("expected wildcard to match Encounter.delete")
	}

	// Should NOT match
	event3 := WebhookEvent{
		ID: "evt-3", Type: "Patient.create", ResourceType: "Patient",
		ResourceID: "p-2", TenantID: "tenant-1", Payload: json.RawMessage(`{}`), Timestamp: time.Now(),
	}
	results = m.Deliver(context.Background(), event3)
	if len(results) != 0 {
		t.Error("expected wildcard *.delete NOT to match Patient.create")
	}
}

func TestWebhookManager_Deliver_PausedSkipped(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	m := newTestManager(ts.Client())
	ep := mustRegisterEndpoint(t, m, ts.URL+"/hook", "tenant-1", []string{"Patient.create"})
	m.PauseEndpoint(context.Background(), ep.ID)

	event := WebhookEvent{
		ID: "evt-1", Type: "Patient.create", ResourceType: "Patient",
		ResourceID: "p-1", TenantID: "tenant-1", Payload: json.RawMessage(`{}`), Timestamp: time.Now(),
	}

	results := m.Deliver(context.Background(), event)
	if len(results) != 0 {
		t.Errorf("expected 0 results for paused endpoint, got %d", len(results))
	}
}

func TestWebhookManager_Deliver_RecordsAttempt(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))
	defer ts.Close()

	m := newTestManager(ts.Client())
	ep := mustRegisterEndpoint(t, m, ts.URL+"/hook", "tenant-1", []string{"Patient.create"})

	event := WebhookEvent{
		ID: "evt-1", Type: "Patient.create", ResourceType: "Patient",
		ResourceID: "p-1", TenantID: "tenant-1", Payload: json.RawMessage(`{"id":"p-1"}`), Timestamp: time.Now(),
	}
	m.Deliver(context.Background(), event)

	deliveries, total, err := m.GetDeliveryLogs(context.Background(), ep.ID, 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 1 {
		t.Fatalf("expected 1 delivery, got %d", total)
	}
	if deliveries[0].Status != "success" {
		t.Errorf("expected status 'success', got %q", deliveries[0].Status)
	}
	if deliveries[0].StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", deliveries[0].StatusCode)
	}
	if deliveries[0].EventType != "Patient.create" {
		t.Errorf("expected event type 'Patient.create', got %q", deliveries[0].EventType)
	}
}

func TestWebhookManager_Deliver_SignatureHeader(t *testing.T) {
	var sigHeader string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sigHeader = r.Header.Get("X-Webhook-Signature")
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	m := newTestManager(ts.Client())
	ep := mustRegisterEndpoint(t, m, ts.URL+"/hook", "tenant-1", []string{"Patient.create"})

	event := WebhookEvent{
		ID: "evt-1", Type: "Patient.create", ResourceType: "Patient",
		ResourceID: "p-1", TenantID: "tenant-1", Payload: json.RawMessage(`{"id":"p-1"}`), Timestamp: time.Now(),
	}
	m.Deliver(context.Background(), event)

	if sigHeader == "" {
		t.Error("expected X-Webhook-Signature header to be set")
	}
	if !strings.HasPrefix(sigHeader, "sha256=") {
		t.Errorf("expected signature to start with 'sha256=', got %q", sigHeader)
	}

	// Verify signature matches
	deliveries, _, _ := m.GetDeliveryLogs(context.Background(), ep.ID, 10, 0)
	if len(deliveries) == 0 {
		t.Fatal("expected at least one delivery")
	}
	expectedSig := SignPayload(deliveries[0].Payload, ep.Secret)
	if sigHeader != "sha256="+expectedSig {
		t.Errorf("signature mismatch: header=%q, expected sha256=%s", sigHeader, expectedSig)
	}
}

func TestWebhookManager_Deliver_TimestampHeader(t *testing.T) {
	var tsHeader string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tsHeader = r.Header.Get("X-Webhook-Timestamp")
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	m := newTestManager(ts.Client())
	mustRegisterEndpoint(t, m, ts.URL+"/hook", "tenant-1", []string{"Patient.create"})

	event := WebhookEvent{
		ID: "evt-1", Type: "Patient.create", ResourceType: "Patient",
		ResourceID: "p-1", TenantID: "tenant-1", Payload: json.RawMessage(`{}`), Timestamp: time.Now(),
	}
	m.Deliver(context.Background(), event)

	if tsHeader == "" {
		t.Error("expected X-Webhook-Timestamp header to be set")
	}
	// Verify it parses as a valid RFC3339 timestamp
	if _, err := time.Parse(time.RFC3339, tsHeader); err != nil {
		t.Errorf("expected valid RFC3339 timestamp, got %q: %v", tsHeader, err)
	}
}

func TestWebhookManager_Deliver_FailedEndpoint(t *testing.T) {
	// Use a URL that will definitely fail to connect
	m := newTestManager(&http.Client{Timeout: 100 * time.Millisecond})
	ep := mustRegisterEndpoint(t, m, "http://192.0.2.1:1/hook", "tenant-1", []string{"Patient.create"})

	event := WebhookEvent{
		ID: "evt-1", Type: "Patient.create", ResourceType: "Patient",
		ResourceID: "p-1", TenantID: "tenant-1", Payload: json.RawMessage(`{}`), Timestamp: time.Now(),
	}
	results := m.Deliver(context.Background(), event)

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Success {
		t.Error("expected failure")
	}
	if results[0].Error == "" {
		t.Error("expected error message")
	}

	deliveries, _, _ := m.GetDeliveryLogs(context.Background(), ep.ID, 10, 0)
	if len(deliveries) == 0 {
		t.Fatal("expected delivery to be recorded")
	}
	if deliveries[0].Status != "failed" {
		t.Errorf("expected status 'failed', got %q", deliveries[0].Status)
	}
	if deliveries[0].StatusCode != 0 {
		t.Errorf("expected status code 0 for connection failure, got %d", deliveries[0].StatusCode)
	}
}

func TestWebhookManager_Deliver_Non2xxRecorded(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal server error"))
	}))
	defer ts.Close()

	m := newTestManager(ts.Client())
	ep := mustRegisterEndpoint(t, m, ts.URL+"/hook", "tenant-1", []string{"Patient.create"})

	event := WebhookEvent{
		ID: "evt-1", Type: "Patient.create", ResourceType: "Patient",
		ResourceID: "p-1", TenantID: "tenant-1", Payload: json.RawMessage(`{}`), Timestamp: time.Now(),
	}
	results := m.Deliver(context.Background(), event)

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Success {
		t.Error("expected failure for 500")
	}
	if results[0].StatusCode != 500 {
		t.Errorf("expected 500, got %d", results[0].StatusCode)
	}

	deliveries, _, _ := m.GetDeliveryLogs(context.Background(), ep.ID, 10, 0)
	if len(deliveries) == 0 {
		t.Fatal("expected delivery to be recorded")
	}
	if deliveries[0].Status != "failed" {
		t.Errorf("expected status 'failed', got %q", deliveries[0].Status)
	}
	if deliveries[0].ResponseBody == "" {
		t.Error("expected response body to be captured")
	}
}

// ===================== Retry =====================

func TestWebhookManager_RetryDelivery(t *testing.T) {
	callCount := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount == 1 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	m := newTestManager(ts.Client())
	ep := mustRegisterEndpoint(t, m, ts.URL+"/hook", "tenant-1", []string{"Patient.create"})

	event := WebhookEvent{
		ID: "evt-1", Type: "Patient.create", ResourceType: "Patient",
		ResourceID: "p-1", TenantID: "tenant-1", Payload: json.RawMessage(`{"id":"p-1"}`), Timestamp: time.Now(),
	}
	m.Deliver(context.Background(), event)

	// Get the failed delivery
	deliveries, _, _ := m.GetDeliveryLogs(context.Background(), ep.ID, 10, 0)
	if len(deliveries) == 0 {
		t.Fatal("expected delivery to be recorded")
	}

	// Retry
	retryAttempt, err := m.RetryDelivery(context.Background(), deliveries[0].ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if retryAttempt.Status != "success" {
		t.Errorf("expected retry to succeed, got status %q", retryAttempt.Status)
	}
	if retryAttempt.Attempt != 2 {
		t.Errorf("expected attempt 2, got %d", retryAttempt.Attempt)
	}
}

func TestWebhookManager_RetryDelivery_NotFound(t *testing.T) {
	m := newTestManager(nil)
	_, err := m.RetryDelivery(context.Background(), "nonexistent-id")
	if err == nil {
		t.Error("expected error for unknown delivery ID")
	}
}

// ===================== Test Endpoint =====================

func TestWebhookManager_TestEndpoint(t *testing.T) {
	var receivedWebhookID string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedWebhookID = r.Header.Get("X-Webhook-ID")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("pong"))
	}))
	defer ts.Close()

	m := newTestManager(ts.Client())
	ep := mustRegisterEndpoint(t, m, ts.URL+"/hook", "tenant-1", []string{"Patient.create"})

	attempt, err := m.TestEndpoint(context.Background(), ep.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if attempt.Status != "success" {
		t.Errorf("expected status 'success', got %q", attempt.Status)
	}
	if attempt.EventType != "webhook.test" {
		t.Errorf("expected event type 'webhook.test', got %q", attempt.EventType)
	}
	if receivedWebhookID == "" {
		t.Error("expected X-Webhook-ID header")
	}
}

func TestWebhookManager_TestEndpoint_NotFound(t *testing.T) {
	m := newTestManager(nil)
	_, err := m.TestEndpoint(context.Background(), "nonexistent-id")
	if err == nil {
		t.Error("expected error for unknown endpoint ID")
	}
}

// ===================== Delivery Logs =====================

func TestWebhookManager_GetDeliveryLogs(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	m := newTestManager(ts.Client())
	ep := mustRegisterEndpoint(t, m, ts.URL+"/hook", "tenant-1", []string{"Patient.create"})

	// Create multiple deliveries
	for i := 0; i < 5; i++ {
		event := WebhookEvent{
			ID: fmt.Sprintf("evt-%d", i), Type: "Patient.create", ResourceType: "Patient",
			ResourceID: fmt.Sprintf("p-%d", i), TenantID: "tenant-1",
			Payload: json.RawMessage(`{}`), Timestamp: time.Now(),
		}
		m.Deliver(context.Background(), event)
	}

	logs, total, err := m.GetDeliveryLogs(context.Background(), ep.ID, 3, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 5 {
		t.Errorf("expected total 5, got %d", total)
	}
	if len(logs) != 3 {
		t.Errorf("expected 3 logs (limit), got %d", len(logs))
	}
}

func TestWebhookManager_GetDeliveryLogs_Empty(t *testing.T) {
	m := newTestManager(nil)
	ep := mustRegisterEndpoint(t, m, "https://example.com/hook", "tenant-1", []string{"Patient.create"})

	logs, total, err := m.GetDeliveryLogs(context.Background(), ep.ID, 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 0 {
		t.Errorf("expected 0, got %d", total)
	}
	if len(logs) != 0 {
		t.Errorf("expected empty logs, got %d", len(logs))
	}
}

// ===================== Concurrent =====================

func TestWebhookManager_ConcurrentDelivery(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	m := newTestManager(ts.Client())
	mustRegisterEndpoint(t, m, ts.URL+"/hook", "tenant-1", []string{"Patient.create"})

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			event := WebhookEvent{
				ID: fmt.Sprintf("evt-%d", idx), Type: "Patient.create", ResourceType: "Patient",
				ResourceID: fmt.Sprintf("p-%d", idx), TenantID: "tenant-1",
				Payload: json.RawMessage(`{}`), Timestamp: time.Now(),
			}
			results := m.Deliver(context.Background(), event)
			if len(results) != 1 {
				t.Errorf("goroutine %d: expected 1 result, got %d", idx, len(results))
			}
		}(i)
	}
	wg.Wait()
}

// ===================== Handler Tests =====================

func newTestEchoHandler(client *http.Client) (*WebhookHandler, *echo.Echo) {
	m := newTestManager(client)
	h := NewWebhookHandler(m)
	e := echo.New()
	return h, e
}

func TestWebhookHandler_RegisterEndpoint(t *testing.T) {
	h, e := newTestEchoHandler(nil)
	body := `{"url":"https://example.com/hook","secret":"my-secret","tenant_id":"tenant-1","client_id":"client-1","events":["Patient.create"]}`
	req := httptest.NewRequest(http.MethodPost, "/webhooks", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := h.RegisterEndpoint(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
	var result map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &result)
	if result["id"] == nil || result["id"] == "" {
		t.Error("expected 'id' in response")
	}
	if result["url"] != "https://example.com/hook" {
		t.Errorf("unexpected URL: %v", result["url"])
	}
}

func TestWebhookHandler_ListEndpoints(t *testing.T) {
	h, e := newTestEchoHandler(nil)

	// Create two endpoints first
	ctx := context.Background()
	h.manager.RegisterEndpoint(ctx, "https://example.com/hook1", "s1", "tenant-1", "c1", []string{"Patient.create"})
	h.manager.RegisterEndpoint(ctx, "https://example.com/hook2", "s2", "tenant-1", "c1", []string{"Patient.update"})

	req := httptest.NewRequest(http.MethodGet, "/webhooks?tenant_id=tenant-1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := h.ListEndpoints(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	var result map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &result)
	data, ok := result["data"].([]interface{})
	if !ok {
		t.Fatal("expected 'data' array in response")
	}
	if len(data) != 2 {
		t.Errorf("expected 2 endpoints, got %d", len(data))
	}
}

func TestWebhookHandler_TestEndpoint(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	h, e := newTestEchoHandler(ts.Client())
	ep, _ := h.manager.RegisterEndpoint(context.Background(), ts.URL+"/hook", "s1", "tenant-1", "c1", []string{"Patient.create"})

	req := httptest.NewRequest(http.MethodPost, "/webhooks/"+ep.ID+"/test", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(ep.ID)

	if err := h.TestEndpointHandler(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestWebhookHandler_GetDeliveryLogs(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	h, e := newTestEchoHandler(ts.Client())
	ep, _ := h.manager.RegisterEndpoint(context.Background(), ts.URL+"/hook", "s1", "tenant-1", "c1", []string{"Patient.create"})

	event := WebhookEvent{
		ID: "evt-1", Type: "Patient.create", ResourceType: "Patient",
		ResourceID: "p-1", TenantID: "tenant-1", Payload: json.RawMessage(`{}`), Timestamp: time.Now(),
	}
	h.manager.Deliver(context.Background(), event)

	req := httptest.NewRequest(http.MethodGet, "/webhooks/"+ep.ID+"/deliveries", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(ep.ID)

	if err := h.GetDeliveryLogs(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestWebhookHandler_RetryDelivery(t *testing.T) {
	callCount := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount == 1 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	h, e := newTestEchoHandler(ts.Client())
	ep, _ := h.manager.RegisterEndpoint(context.Background(), ts.URL+"/hook", "s1", "tenant-1", "c1", []string{"Patient.create"})

	event := WebhookEvent{
		ID: "evt-1", Type: "Patient.create", ResourceType: "Patient",
		ResourceID: "p-1", TenantID: "tenant-1", Payload: json.RawMessage(`{"id":"p-1"}`), Timestamp: time.Now(),
	}
	h.manager.Deliver(context.Background(), event)

	// Get the failed delivery ID
	deliveries, _, _ := h.manager.GetDeliveryLogs(context.Background(), ep.ID, 10, 0)
	if len(deliveries) == 0 {
		t.Fatal("expected at least one delivery")
	}

	req := httptest.NewRequest(http.MethodPost, "/webhooks/deliveries/"+deliveries[0].ID+"/retry", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(deliveries[0].ID)

	if err := h.RetryDeliveryHandler(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}
