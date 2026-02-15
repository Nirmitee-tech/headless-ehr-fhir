package subscription

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
)

// -- Mock Repository --

type mockSubRepo struct {
	store         map[uuid.UUID]*Subscription
	notifications map[uuid.UUID][]*SubscriptionNotification
}

func newMockSubRepo() *mockSubRepo {
	return &mockSubRepo{
		store:         make(map[uuid.UUID]*Subscription),
		notifications: make(map[uuid.UUID][]*SubscriptionNotification),
	}
}

func (m *mockSubRepo) Create(_ context.Context, sub *Subscription) error {
	sub.ID = uuid.New()
	if sub.FHIRID == "" {
		sub.FHIRID = sub.ID.String()
	}
	sub.CreatedAt = time.Now()
	sub.UpdatedAt = time.Now()
	m.store[sub.ID] = sub
	return nil
}

func (m *mockSubRepo) GetByID(_ context.Context, id uuid.UUID) (*Subscription, error) {
	s, ok := m.store[id]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return s, nil
}

func (m *mockSubRepo) GetByFHIRID(_ context.Context, fhirID string) (*Subscription, error) {
	for _, s := range m.store {
		if s.FHIRID == fhirID {
			return s, nil
		}
	}
	return nil, fmt.Errorf("not found")
}

func (m *mockSubRepo) Update(_ context.Context, sub *Subscription) error {
	if _, ok := m.store[sub.ID]; !ok {
		return fmt.Errorf("not found")
	}
	m.store[sub.ID] = sub
	return nil
}

func (m *mockSubRepo) Delete(_ context.Context, id uuid.UUID) error {
	delete(m.store, id)
	return nil
}

func (m *mockSubRepo) Search(_ context.Context, params map[string]string, limit, offset int) ([]*Subscription, int, error) {
	var r []*Subscription
	for _, s := range m.store {
		r = append(r, s)
	}
	return r, len(r), nil
}

func (m *mockSubRepo) ListActive(_ context.Context) ([]*Subscription, error) {
	var r []*Subscription
	for _, s := range m.store {
		if s.Status == "active" {
			r = append(r, s)
		}
	}
	return r, nil
}

func (m *mockSubRepo) ListExpired(_ context.Context) ([]*Subscription, error) {
	now := time.Now()
	var r []*Subscription
	for _, s := range m.store {
		if s.Status == "active" && s.EndTime != nil && s.EndTime.Before(now) {
			r = append(r, s)
		}
	}
	return r, nil
}

func (m *mockSubRepo) UpdateStatus(_ context.Context, id uuid.UUID, status string, errorText *string) error {
	s, ok := m.store[id]
	if !ok {
		return fmt.Errorf("not found")
	}
	s.Status = status
	s.ErrorText = errorText
	return nil
}

func (m *mockSubRepo) CreateNotification(_ context.Context, n *SubscriptionNotification) error {
	n.ID = uuid.New()
	m.notifications[n.SubscriptionID] = append(m.notifications[n.SubscriptionID], n)
	return nil
}

func (m *mockSubRepo) ListPendingNotifications(_ context.Context, limit int) ([]*SubscriptionNotification, error) {
	var r []*SubscriptionNotification
	for _, notifs := range m.notifications {
		for _, n := range notifs {
			if n.Status == "pending" {
				r = append(r, n)
			}
		}
	}
	return r, nil
}

func (m *mockSubRepo) UpdateNotification(_ context.Context, n *SubscriptionNotification) error {
	for subID, notifs := range m.notifications {
		for i, existing := range notifs {
			if existing.ID == n.ID {
				m.notifications[subID][i] = n
				return nil
			}
		}
	}
	return fmt.Errorf("notification not found")
}

func (m *mockSubRepo) ListNotificationsBySubscription(_ context.Context, subscriptionID uuid.UUID, limit, offset int) ([]*SubscriptionNotification, int, error) {
	notifs := m.notifications[subscriptionID]
	return notifs, len(notifs), nil
}

func (m *mockSubRepo) DeleteOldNotifications(_ context.Context, before time.Time, statuses []string) (int64, error) {
	return 0, nil
}

func newTestService() *Service {
	return NewService(newMockSubRepo())
}

// -- Service Tests --

func TestCreateSubscription_Success(t *testing.T) {
	svc := newTestService()
	sub := &Subscription{
		Criteria:        "Observation?code=1234",
		ChannelEndpoint: "https://example.com/webhook",
	}
	if err := svc.CreateSubscription(context.Background(), sub); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sub.ID == uuid.Nil {
		t.Error("expected ID to be set")
	}
	if sub.FHIRID == "" {
		t.Error("expected FHIRID to be set")
	}
	if sub.Status != "requested" {
		t.Errorf("expected status 'requested', got %q", sub.Status)
	}
	if sub.ChannelType != "rest-hook" {
		t.Errorf("expected channel type 'rest-hook', got %q", sub.ChannelType)
	}
	if sub.VersionID != 1 {
		t.Errorf("expected version 1, got %d", sub.VersionID)
	}
}

func TestCreateSubscription_MissingCriteria(t *testing.T) {
	svc := newTestService()
	sub := &Subscription{ChannelEndpoint: "https://example.com/webhook"}
	if err := svc.CreateSubscription(context.Background(), sub); err == nil {
		t.Fatal("expected error for missing criteria")
	}
}

func TestCreateSubscription_MissingEndpoint(t *testing.T) {
	svc := newTestService()
	sub := &Subscription{Criteria: "Observation?code=1234"}
	if err := svc.CreateSubscription(context.Background(), sub); err == nil {
		t.Fatal("expected error for missing channel endpoint")
	}
}

func TestCreateSubscription_InvalidStatus(t *testing.T) {
	svc := newTestService()
	sub := &Subscription{
		Criteria:        "Observation",
		ChannelEndpoint: "https://example.com/webhook",
		Status:          "bogus",
	}
	if err := svc.CreateSubscription(context.Background(), sub); err == nil {
		t.Fatal("expected error for invalid status")
	}
}

func TestCreateSubscription_InvalidChannelType(t *testing.T) {
	svc := newTestService()
	sub := &Subscription{
		Criteria:        "Observation",
		ChannelEndpoint: "https://example.com/webhook",
		ChannelType:     "email",
	}
	if err := svc.CreateSubscription(context.Background(), sub); err == nil {
		t.Fatal("expected error for invalid channel type")
	}
}

func TestCreateSubscription_ValidStatuses(t *testing.T) {
	for _, s := range []string{"requested", "active", "error", "off"} {
		svc := newTestService()
		sub := &Subscription{
			Criteria:        "Observation",
			ChannelEndpoint: "https://example.com/webhook",
			Status:          s,
		}
		if err := svc.CreateSubscription(context.Background(), sub); err != nil {
			t.Errorf("status %q should be valid: %v", s, err)
		}
	}
}

func TestGetSubscription_Success(t *testing.T) {
	svc := newTestService()
	sub := &Subscription{Criteria: "Observation", ChannelEndpoint: "https://example.com/webhook"}
	svc.CreateSubscription(context.Background(), sub)
	got, err := svc.GetSubscription(context.Background(), sub.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != sub.ID {
		t.Error("ID mismatch")
	}
}

func TestGetSubscription_NotFound(t *testing.T) {
	svc := newTestService()
	if _, err := svc.GetSubscription(context.Background(), uuid.New()); err == nil {
		t.Fatal("expected error")
	}
}

func TestGetSubscriptionByFHIRID(t *testing.T) {
	svc := newTestService()
	sub := &Subscription{Criteria: "Observation", ChannelEndpoint: "https://example.com/webhook"}
	svc.CreateSubscription(context.Background(), sub)
	got, err := svc.GetSubscriptionByFHIRID(context.Background(), sub.FHIRID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != sub.ID {
		t.Error("ID mismatch")
	}
}

func TestUpdateSubscription_Success(t *testing.T) {
	svc := newTestService()
	sub := &Subscription{Criteria: "Observation", ChannelEndpoint: "https://example.com/webhook"}
	svc.CreateSubscription(context.Background(), sub)
	sub.Status = "active"
	if err := svc.UpdateSubscription(context.Background(), sub); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, _ := svc.GetSubscription(context.Background(), sub.ID)
	if got.Status != "active" {
		t.Errorf("expected status 'active', got %q", got.Status)
	}
}

func TestUpdateSubscription_InvalidStatus(t *testing.T) {
	svc := newTestService()
	sub := &Subscription{Criteria: "Observation", ChannelEndpoint: "https://example.com/webhook"}
	svc.CreateSubscription(context.Background(), sub)
	sub.Status = "invalid"
	if err := svc.UpdateSubscription(context.Background(), sub); err == nil {
		t.Fatal("expected error for invalid status")
	}
}

func TestDeleteSubscription_Success(t *testing.T) {
	svc := newTestService()
	sub := &Subscription{Criteria: "Observation", ChannelEndpoint: "https://example.com/webhook"}
	svc.CreateSubscription(context.Background(), sub)
	if err := svc.DeleteSubscription(context.Background(), sub.ID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := svc.GetSubscription(context.Background(), sub.ID); err == nil {
		t.Fatal("expected error after delete")
	}
}

func TestSearchSubscriptions(t *testing.T) {
	svc := newTestService()
	svc.CreateSubscription(context.Background(), &Subscription{Criteria: "Observation", ChannelEndpoint: "https://example.com/webhook"})
	svc.CreateSubscription(context.Background(), &Subscription{Criteria: "Patient", ChannelEndpoint: "https://example.com/webhook2"})
	items, total, err := svc.SearchSubscriptions(context.Background(), map[string]string{}, 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 2 || len(items) != 2 {
		t.Errorf("expected 2 subscriptions, got %d", total)
	}
}

func TestCreateNotification_Success(t *testing.T) {
	svc := newTestService()
	sub := &Subscription{Criteria: "Observation", ChannelEndpoint: "https://example.com/webhook"}
	svc.CreateSubscription(context.Background(), sub)

	payload, _ := json.Marshal(map[string]string{"resourceType": "Observation", "id": "obs-1"})
	n := &SubscriptionNotification{
		SubscriptionID: sub.ID,
		ResourceType:   "Observation",
		ResourceID:     "obs-1",
		EventType:      "create",
		Status:         "pending",
		Payload:        payload,
		MaxAttempts:    5,
		NextAttemptAt:  time.Now(),
	}
	if err := svc.CreateNotification(context.Background(), n); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n.ID == uuid.Nil {
		t.Error("expected notification ID to be set")
	}

	notifs, total, err := svc.ListNotificationsBySubscription(context.Background(), sub.ID, 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 1 || len(notifs) != 1 {
		t.Errorf("expected 1 notification, got %d", total)
	}
}

func TestListActive(t *testing.T) {
	svc := newTestService()
	svc.CreateSubscription(context.Background(), &Subscription{Criteria: "Observation", ChannelEndpoint: "https://example.com/webhook", Status: "active"})
	svc.CreateSubscription(context.Background(), &Subscription{Criteria: "Patient", ChannelEndpoint: "https://example.com/webhook2", Status: "off"})
	active, err := svc.ListActive(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(active) != 1 {
		t.Errorf("expected 1 active subscription, got %d", len(active))
	}
}

func TestUpdateStatus(t *testing.T) {
	svc := newTestService()
	sub := &Subscription{Criteria: "Observation", ChannelEndpoint: "https://example.com/webhook", Status: "active"}
	svc.CreateSubscription(context.Background(), sub)

	errText := "connection refused"
	if err := svc.UpdateStatus(context.Background(), sub.ID, "error", &errText); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, _ := svc.GetSubscription(context.Background(), sub.ID)
	if got.Status != "error" {
		t.Errorf("expected status 'error', got %q", got.Status)
	}
	if got.ErrorText == nil || *got.ErrorText != "connection refused" {
		t.Error("expected error text to be set")
	}
}

func TestUpdateSubscription_InvalidChannelType(t *testing.T) {
	svc := newTestService()
	sub := &Subscription{Criteria: "Observation", ChannelEndpoint: "https://example.com/webhook"}
	svc.CreateSubscription(context.Background(), sub)
	sub.ChannelType = "email"
	if err := svc.UpdateSubscription(context.Background(), sub); err == nil {
		t.Fatal("expected error for invalid channel type")
	}
}

func TestListExpired(t *testing.T) {
	svc := newTestService()
	past := time.Now().Add(-1 * time.Hour)
	svc.CreateSubscription(context.Background(), &Subscription{
		Criteria: "Observation", ChannelEndpoint: "https://example.com/webhook",
		Status: "active", EndTime: &past,
	})
	svc.CreateSubscription(context.Background(), &Subscription{
		Criteria: "Patient", ChannelEndpoint: "https://example.com/webhook2",
		Status: "active",
	})
	expired, err := svc.ListExpired(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(expired) != 1 {
		t.Errorf("expected 1 expired subscription, got %d", len(expired))
	}
}

func TestSetVersionTracker(t *testing.T) {
	svc := newTestService()
	if svc.VersionTracker() != nil {
		t.Error("expected nil VersionTracker initially")
	}
	svc.SetVersionTracker(nil)
	if svc.VersionTracker() != nil {
		t.Error("expected nil after setting nil")
	}
}

func TestListPendingNotifications(t *testing.T) {
	svc := newTestService()
	sub := &Subscription{Criteria: "Observation", ChannelEndpoint: "https://example.com/webhook"}
	svc.CreateSubscription(context.Background(), sub)
	svc.CreateNotification(context.Background(), &SubscriptionNotification{
		SubscriptionID: sub.ID, ResourceType: "Observation", ResourceID: "obs-1",
		EventType: "create", Status: "pending", Payload: json.RawMessage(`{}`), MaxAttempts: 5,
	})
	pending, err := svc.ListPendingNotifications(context.Background(), 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pending) != 1 {
		t.Errorf("expected 1 pending notification, got %d", len(pending))
	}
}

func TestUpdateNotification(t *testing.T) {
	svc := newTestService()
	sub := &Subscription{Criteria: "Observation", ChannelEndpoint: "https://example.com/webhook"}
	svc.CreateSubscription(context.Background(), sub)
	notif := &SubscriptionNotification{
		SubscriptionID: sub.ID, ResourceType: "Observation", ResourceID: "obs-1",
		EventType: "create", Status: "pending", Payload: json.RawMessage(`{}`), MaxAttempts: 5,
	}
	svc.CreateNotification(context.Background(), notif)
	notif.Status = "delivered"
	now := time.Now()
	notif.DeliveredAt = &now
	if err := svc.UpdateNotification(context.Background(), notif); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGetSubscriptionByFHIRID_NotFound(t *testing.T) {
	svc := newTestService()
	if _, err := svc.GetSubscriptionByFHIRID(context.Background(), "nonexistent"); err == nil {
		t.Fatal("expected error")
	}
}

// -- SSRF Protection Tests --

func TestCreateSubscription_PrivateIPEndpoint(t *testing.T) {
	original := resolveHost
	defer func() { resolveHost = original }()

	tests := []struct {
		name     string
		endpoint string
		ip       string
	}{
		{"10.x private", "https://internal.corp/hook", "10.0.0.1"},
		{"192.168.x private", "https://homelab.local/hook", "192.168.1.1"},
		{"loopback", "https://loop.test/hook", "127.0.0.1"},
		{"cloud metadata", "https://metadata.test/hook", "169.254.169.254"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolveHost = func(host string) ([]string, error) {
				return []string{tt.ip}, nil
			}
			svc := newTestService()
			sub := &Subscription{
				Criteria:        "Observation",
				ChannelEndpoint: tt.endpoint,
			}
			if err := svc.CreateSubscription(context.Background(), sub); err == nil {
				t.Fatalf("expected error for endpoint resolving to %s", tt.ip)
			}
		})
	}
}

func TestCreateSubscription_ValidEndpoint(t *testing.T) {
	original := resolveHost
	defer func() { resolveHost = original }()

	resolveHost = func(host string) ([]string, error) {
		return []string{"93.184.216.34"}, nil
	}
	svc := newTestService()
	sub := &Subscription{
		Criteria:        "Observation",
		ChannelEndpoint: "https://example.com/webhook",
	}
	if err := svc.CreateSubscription(context.Background(), sub); err != nil {
		t.Fatalf("unexpected error for valid endpoint: %v", err)
	}
}

func TestUpdateSubscription_SSRFProtection(t *testing.T) {
	original := resolveHost
	defer func() { resolveHost = original }()

	// First create with a valid endpoint
	resolveHost = func(host string) ([]string, error) {
		return []string{"93.184.216.34"}, nil
	}
	svc := newTestService()
	sub := &Subscription{
		Criteria:        "Observation",
		ChannelEndpoint: "https://example.com/webhook",
	}
	if err := svc.CreateSubscription(context.Background(), sub); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Now try to update with a private IP endpoint
	resolveHost = func(host string) ([]string, error) {
		return []string{"10.0.0.1"}, nil
	}
	sub.ChannelEndpoint = "https://evil.internal/steal"
	if err := svc.UpdateSubscription(context.Background(), sub); err == nil {
		t.Fatal("expected error for private IP endpoint on update")
	}
}

func TestCreateSubscription_RejectsNonHTTPScheme(t *testing.T) {
	svc := newTestService()
	sub := &Subscription{
		Criteria:        "Observation",
		ChannelEndpoint: "ftp://example.com/webhook",
	}
	if err := svc.CreateSubscription(context.Background(), sub); err == nil {
		t.Fatal("expected error for ftp scheme")
	}
}

func TestCreateSubscription_RejectsLocalhost(t *testing.T) {
	svc := newTestService()
	sub := &Subscription{
		Criteria:        "Observation",
		ChannelEndpoint: "https://localhost/webhook",
	}
	if err := svc.CreateSubscription(context.Background(), sub); err == nil {
		t.Fatal("expected error for localhost")
	}
}
