package subscription

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestSubscription_ToFHIR(t *testing.T) {
	now := time.Now().UTC()
	sub := &Subscription{
		ID:              uuid.New(),
		FHIRID:          "sub-123",
		Status:          "active",
		Criteria:        "Observation?code=1234",
		ChannelType:     "rest-hook",
		ChannelEndpoint: "https://example.com/webhook",
		ChannelPayload:  "application/fhir+json",
		VersionID:       1,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	result := sub.ToFHIR()

	if result["resourceType"] != "Subscription" {
		t.Errorf("expected resourceType 'Subscription', got %v", result["resourceType"])
	}
	if result["id"] != "sub-123" {
		t.Errorf("expected id 'sub-123', got %v", result["id"])
	}
	if result["status"] != "active" {
		t.Errorf("expected status 'active', got %v", result["status"])
	}
	if result["criteria"] != "Observation?code=1234" {
		t.Errorf("expected criteria 'Observation?code=1234', got %v", result["criteria"])
	}
	channel, ok := result["channel"].(map[string]interface{})
	if !ok {
		t.Fatal("expected channel to be map[string]interface{}")
	}
	if channel["type"] != "rest-hook" {
		t.Errorf("expected channel type 'rest-hook', got %v", channel["type"])
	}
	if channel["endpoint"] != "https://example.com/webhook" {
		t.Errorf("expected channel endpoint, got %v", channel["endpoint"])
	}
	if channel["payload"] != "application/fhir+json" {
		t.Errorf("expected channel payload, got %v", channel["payload"])
	}
	// Verify meta block
	meta, ok := result["meta"].(map[string]interface{})
	if !ok {
		t.Fatal("expected meta to be map[string]interface{}")
	}
	if meta["versionId"] != 1 {
		t.Errorf("expected versionId 1, got %v", meta["versionId"])
	}
	// No end or error when not set
	if result["end"] != nil {
		t.Error("expected end to be nil when not set")
	}
	if result["error"] != nil {
		t.Error("expected error to be nil when not set")
	}
}

func TestSubscription_ToFHIR_WithHeaders(t *testing.T) {
	now := time.Now().UTC()
	headers, _ := json.Marshal([]string{"Authorization: Bearer token123"})
	sub := &Subscription{
		ID:              uuid.New(),
		FHIRID:          "sub-456",
		Status:          "active",
		Criteria:        "Patient",
		ChannelType:     "rest-hook",
		ChannelEndpoint: "https://example.com/webhook",
		ChannelPayload:  "application/fhir+json",
		ChannelHeaders:  headers,
		VersionID:       1,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	result := sub.ToFHIR()
	channel := result["channel"].(map[string]interface{})
	hdrs, ok := channel["header"].([]string)
	if !ok {
		t.Fatal("expected header to be []string")
	}
	if len(hdrs) != 1 || hdrs[0] != "Authorization: Bearer token123" {
		t.Errorf("unexpected headers: %v", hdrs)
	}
}

func TestSubscription_ToFHIR_WithMultipleHeaders(t *testing.T) {
	now := time.Now().UTC()
	headers, _ := json.Marshal([]string{"Authorization: Bearer token123", "X-Custom: value"})
	sub := &Subscription{
		ID:              uuid.New(),
		FHIRID:          "sub-multi",
		Status:          "active",
		Criteria:        "Patient",
		ChannelType:     "rest-hook",
		ChannelEndpoint: "https://example.com/webhook",
		ChannelPayload:  "application/fhir+json",
		ChannelHeaders:  headers,
		VersionID:       1,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	result := sub.ToFHIR()
	channel := result["channel"].(map[string]interface{})
	hdrs := channel["header"].([]string)
	if len(hdrs) != 2 {
		t.Fatalf("expected 2 headers, got %d", len(hdrs))
	}
}

func TestSubscription_ToFHIR_NullHeaders(t *testing.T) {
	now := time.Now().UTC()
	sub := &Subscription{
		ID:              uuid.New(),
		FHIRID:          "sub-null",
		Status:          "active",
		Criteria:        "Patient",
		ChannelType:     "rest-hook",
		ChannelEndpoint: "https://example.com/webhook",
		ChannelPayload:  "application/fhir+json",
		ChannelHeaders:  json.RawMessage("null"),
		VersionID:       1,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	result := sub.ToFHIR()
	channel := result["channel"].(map[string]interface{})
	if channel["header"] != nil {
		t.Error("expected no header when ChannelHeaders is null")
	}
}

func TestSubscription_ToFHIR_MalformedHeaders(t *testing.T) {
	now := time.Now().UTC()
	sub := &Subscription{
		ID:              uuid.New(),
		FHIRID:          "sub-bad",
		Status:          "active",
		Criteria:        "Patient",
		ChannelType:     "rest-hook",
		ChannelEndpoint: "https://example.com/webhook",
		ChannelPayload:  "application/fhir+json",
		ChannelHeaders:  json.RawMessage(`{"not":"an array"}`),
		VersionID:       1,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	// Should not panic on malformed headers
	result := sub.ToFHIR()
	channel := result["channel"].(map[string]interface{})
	if channel["header"] != nil {
		t.Error("expected no header when ChannelHeaders is malformed")
	}
}

func TestSubscription_ToFHIR_WithEndTime(t *testing.T) {
	now := time.Now().UTC()
	endTime := now.Add(24 * time.Hour)
	sub := &Subscription{
		ID:              uuid.New(),
		FHIRID:          "sub-789",
		Status:          "active",
		Criteria:        "Observation",
		ChannelType:     "rest-hook",
		ChannelEndpoint: "https://example.com/webhook",
		ChannelPayload:  "application/fhir+json",
		EndTime:         &endTime,
		VersionID:       1,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	result := sub.ToFHIR()
	endStr, ok := result["end"].(string)
	if !ok || endStr == "" {
		t.Error("expected end to be a non-empty string")
	}
}

func TestSubscription_ToFHIR_WithError(t *testing.T) {
	now := time.Now().UTC()
	errText := "connection refused"
	sub := &Subscription{
		ID:              uuid.New(),
		FHIRID:          "sub-err",
		Status:          "error",
		Criteria:        "Observation",
		ChannelType:     "rest-hook",
		ChannelEndpoint: "https://example.com/webhook",
		ChannelPayload:  "application/fhir+json",
		ErrorText:       &errText,
		VersionID:       1,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	result := sub.ToFHIR()
	if result["error"] != "connection refused" {
		t.Errorf("expected error 'connection refused', got %v", result["error"])
	}
}

func TestSubscription_ToFHIR_AllOptionalFields(t *testing.T) {
	now := time.Now().UTC()
	endTime := now.Add(24 * time.Hour)
	errText := "timeout"
	headers, _ := json.Marshal([]string{"Authorization: Bearer abc"})
	sub := &Subscription{
		ID:              uuid.New(),
		FHIRID:          "sub-all",
		Status:          "error",
		Criteria:        "Observation?code=1234",
		ChannelType:     "rest-hook",
		ChannelEndpoint: "https://example.com/webhook",
		ChannelPayload:  "application/fhir+json",
		ChannelHeaders:  headers,
		EndTime:         &endTime,
		ErrorText:       &errText,
		VersionID:       3,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	result := sub.ToFHIR()
	if result["end"] == nil {
		t.Error("expected end to be set")
	}
	if result["error"] != "timeout" {
		t.Errorf("expected error 'timeout', got %v", result["error"])
	}
	channel := result["channel"].(map[string]interface{})
	if channel["header"] == nil {
		t.Error("expected header to be set")
	}
}

func TestSubscription_ToFHIR_EmptyPayload(t *testing.T) {
	now := time.Now().UTC()
	sub := &Subscription{
		ID:              uuid.New(),
		FHIRID:          "sub-nop",
		Status:          "active",
		Criteria:        "Patient",
		ChannelType:     "rest-hook",
		ChannelEndpoint: "https://example.com/webhook",
		ChannelPayload:  "",
		VersionID:       1,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	result := sub.ToFHIR()
	channel := result["channel"].(map[string]interface{})
	// Empty payload should not appear
	if channel["payload"] != nil {
		t.Error("expected no payload when ChannelPayload is empty")
	}
}

func TestSubscription_GetSetVersionID(t *testing.T) {
	sub := &Subscription{VersionID: 3}
	if sub.GetVersionID() != 3 {
		t.Errorf("expected version 3, got %d", sub.GetVersionID())
	}
	sub.SetVersionID(5)
	if sub.GetVersionID() != 5 {
		t.Errorf("expected version 5, got %d", sub.GetVersionID())
	}
}

// -- SubscriptionNotification model tests --

func TestSubscriptionNotification_Fields(t *testing.T) {
	now := time.Now().UTC()
	subID := uuid.New()
	payload := json.RawMessage(`{"resourceType":"Observation","id":"obs-1"}`)
	n := &SubscriptionNotification{
		ID:             uuid.New(),
		SubscriptionID: subID,
		ResourceType:   "Observation",
		ResourceID:     "obs-1",
		EventType:      "create",
		Status:         "pending",
		Payload:        payload,
		AttemptCount:   0,
		MaxAttempts:    5,
		NextAttemptAt:  now,
		CreatedAt:      now,
	}
	if n.SubscriptionID != subID {
		t.Error("subscription ID mismatch")
	}
	if n.ResourceType != "Observation" {
		t.Errorf("expected resource type 'Observation', got %q", n.ResourceType)
	}
	if n.Status != "pending" {
		t.Errorf("expected status 'pending', got %q", n.Status)
	}
	if n.MaxAttempts != 5 {
		t.Errorf("expected max attempts 5, got %d", n.MaxAttempts)
	}
}

func TestSubscriptionNotification_DeliveredFields(t *testing.T) {
	now := time.Now().UTC()
	errMsg := "connection refused"
	delivered := now.Add(1 * time.Second)
	n := &SubscriptionNotification{
		ID:            uuid.New(),
		Status:        "delivered",
		AttemptCount:  2,
		LastError:     &errMsg,
		DeliveredAt:   &delivered,
		NextAttemptAt: now,
		CreatedAt:     now,
	}
	if n.LastError == nil || *n.LastError != "connection refused" {
		t.Error("expected last error to be set")
	}
	if n.DeliveredAt == nil {
		t.Error("expected delivered_at to be set")
	}
	if n.AttemptCount != 2 {
		t.Errorf("expected attempt count 2, got %d", n.AttemptCount)
	}
}

func TestSubscriptionNotification_JSONRoundTrip(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Millisecond)
	n := &SubscriptionNotification{
		ID:             uuid.New(),
		SubscriptionID: uuid.New(),
		ResourceType:   "Patient",
		ResourceID:     "p-1",
		EventType:      "update",
		Status:         "pending",
		Payload:        json.RawMessage(`{"id":"p-1"}`),
		AttemptCount:   1,
		MaxAttempts:    5,
		NextAttemptAt:  now,
		CreatedAt:      now,
	}
	data, err := json.Marshal(n)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}
	var decoded SubscriptionNotification
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if decoded.ResourceType != "Patient" {
		t.Errorf("expected resource type 'Patient', got %q", decoded.ResourceType)
	}
	if decoded.EventType != "update" {
		t.Errorf("expected event type 'update', got %q", decoded.EventType)
	}
}
