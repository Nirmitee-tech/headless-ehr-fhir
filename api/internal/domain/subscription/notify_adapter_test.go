package subscription

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/ehr/ehr/internal/platform/fhir"
)

func TestNotifyRepoAdapter_ListActive(t *testing.T) {
	repo := newMockSubRepo()
	adapter := NewNotifyRepoAdapter(repo)

	headers, _ := json.Marshal([]string{"Authorization: Bearer abc"})
	sub := &Subscription{
		Criteria:        "Observation",
		ChannelEndpoint: "https://example.com/webhook",
		ChannelPayload:  "application/fhir+json",
		Status:          "active",
		ChannelHeaders:  headers,
	}
	repo.Create(context.Background(), sub)

	active, err := adapter.ListActive(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(active) != 1 {
		t.Fatalf("expected 1 active subscription, got %d", len(active))
	}
	if active[0].FHIRID != sub.FHIRID {
		t.Errorf("expected FHIRID %q, got %q", sub.FHIRID, active[0].FHIRID)
	}
	if active[0].Criteria != "Observation" {
		t.Errorf("expected criteria 'Observation', got %q", active[0].Criteria)
	}
	if active[0].ChannelEndpoint != "https://example.com/webhook" {
		t.Errorf("expected channel endpoint, got %q", active[0].ChannelEndpoint)
	}
	if len(active[0].ChannelHeaders) != 1 {
		t.Errorf("expected 1 header, got %d", len(active[0].ChannelHeaders))
	}
}

func TestNotifyRepoAdapter_ListActive_NoHeaders(t *testing.T) {
	repo := newMockSubRepo()
	adapter := NewNotifyRepoAdapter(repo)

	sub := &Subscription{
		Criteria:        "Patient",
		ChannelEndpoint: "https://example.com/webhook",
		ChannelPayload:  "application/fhir+json",
		Status:          "active",
	}
	repo.Create(context.Background(), sub)

	active, err := adapter.ListActive(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(active) != 1 {
		t.Fatalf("expected 1 active subscription, got %d", len(active))
	}
	if len(active[0].ChannelHeaders) != 0 {
		t.Errorf("expected 0 headers, got %d", len(active[0].ChannelHeaders))
	}
}

func TestNotifyRepoAdapter_CreateNotification(t *testing.T) {
	repo := newMockSubRepo()
	adapter := NewNotifyRepoAdapter(repo)

	sub := &Subscription{Criteria: "Observation", ChannelEndpoint: "https://example.com/webhook", Status: "active"}
	repo.Create(context.Background(), sub)

	n := &fhir.NotificationRecord{
		SubscriptionID: sub.ID,
		ResourceType:   "Observation",
		ResourceID:     "obs-1",
		EventType:      "create",
		Status:         "pending",
		Payload:        json.RawMessage(`{}`),
		MaxAttempts:    5,
		NextAttemptAt:  time.Now(),
	}
	if err := adapter.CreateNotification(context.Background(), n); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n.ID == uuid.Nil {
		t.Error("expected notification ID to be set")
	}
}

func TestNotifyRepoAdapter_UpdateNotification(t *testing.T) {
	repo := newMockSubRepo()
	adapter := NewNotifyRepoAdapter(repo)

	sub := &Subscription{Criteria: "Observation", ChannelEndpoint: "https://example.com/webhook", Status: "active"}
	repo.Create(context.Background(), sub)

	// Create domain notification
	dn := &SubscriptionNotification{
		SubscriptionID: sub.ID,
		ResourceType:   "Observation",
		ResourceID:     "obs-1",
		EventType:      "create",
		Status:         "pending",
		Payload:        json.RawMessage(`{}`),
		MaxAttempts:    5,
	}
	repo.CreateNotification(context.Background(), dn)

	// Update via adapter
	now := time.Now()
	n := &fhir.NotificationRecord{
		ID:           dn.ID,
		Status:       "delivered",
		AttemptCount: 1,
		DeliveredAt:  &now,
	}
	if err := adapter.UpdateNotification(context.Background(), n); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNotifyRepoAdapter_UpdateSubscriptionStatus(t *testing.T) {
	repo := newMockSubRepo()
	adapter := NewNotifyRepoAdapter(repo)

	sub := &Subscription{Criteria: "Observation", ChannelEndpoint: "https://example.com/webhook", Status: "active"}
	repo.Create(context.Background(), sub)

	errText := "connection refused"
	if err := adapter.UpdateSubscriptionStatus(context.Background(), sub.ID, "error", &errText); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, _ := repo.GetByID(context.Background(), sub.ID)
	if got.Status != "error" {
		t.Errorf("expected status 'error', got %q", got.Status)
	}
	if got.ErrorText == nil || *got.ErrorText != "connection refused" {
		t.Error("expected error text to be set")
	}
}

func TestNotifyRepoAdapter_ListExpiredSubscriptions(t *testing.T) {
	repo := newMockSubRepo()
	adapter := NewNotifyRepoAdapter(repo)

	past := time.Now().Add(-1 * time.Hour)
	sub := &Subscription{
		Criteria:        "Observation",
		ChannelEndpoint: "https://example.com/webhook",
		Status:          "active",
		EndTime:         &past,
	}
	repo.Create(context.Background(), sub)

	expired, err := adapter.ListExpiredSubscriptions(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(expired) != 1 {
		t.Fatalf("expected 1 expired subscription, got %d", len(expired))
	}
	if expired[0].FHIRID != sub.FHIRID {
		t.Errorf("expected FHIRID %q, got %q", sub.FHIRID, expired[0].FHIRID)
	}
}

func TestNotifyRepoAdapter_ListPendingNotifications(t *testing.T) {
	repo := newMockSubRepo()
	adapter := NewNotifyRepoAdapter(repo)

	headers, _ := json.Marshal([]string{"Authorization: Bearer token"})
	sub := &Subscription{
		Criteria:        "Observation",
		ChannelEndpoint: "https://example.com/webhook",
		ChannelPayload:  "application/fhir+json",
		Status:          "active",
		ChannelHeaders:  headers,
	}
	repo.Create(context.Background(), sub)

	notif := &SubscriptionNotification{
		SubscriptionID: sub.ID,
		ResourceType:   "Observation",
		ResourceID:     "obs-1",
		EventType:      "create",
		Status:         "pending",
		Payload:        json.RawMessage(`{}`),
		MaxAttempts:    5,
	}
	repo.CreateNotification(context.Background(), notif)

	pending, err := adapter.ListPendingNotifications(context.Background(), 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pending) != 1 {
		t.Fatalf("expected 1 pending notification, got %d", len(pending))
	}
	if pending[0].ChannelEndpoint != "https://example.com/webhook" {
		t.Errorf("expected channel endpoint, got %q", pending[0].ChannelEndpoint)
	}
	if pending[0].ChannelPayload != "application/fhir+json" {
		t.Errorf("expected channel payload, got %q", pending[0].ChannelPayload)
	}
	if pending[0].SubFHIRID != sub.FHIRID {
		t.Errorf("expected sub FHIR ID %q, got %q", sub.FHIRID, pending[0].SubFHIRID)
	}
	if len(pending[0].ChannelHeaders) != 1 {
		t.Errorf("expected 1 header, got %d", len(pending[0].ChannelHeaders))
	}
}

func TestNotifyRepoAdapter_InterfaceCompliance(t *testing.T) {
	var _ fhir.NotificationRepo = (*NotifyRepoAdapter)(nil)
}
