package fhir

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

func TestParseCriteria_ResourceOnly(t *testing.T) {
	rt, params := ParseCriteria("Patient")
	if rt != "Patient" {
		t.Errorf("expected 'Patient', got %q", rt)
	}
	if len(params) != 0 {
		t.Errorf("expected 0 params, got %d", len(params))
	}
}

func TestParseCriteria_WithParams(t *testing.T) {
	rt, params := ParseCriteria("Observation?code=1234&status=final")
	if rt != "Observation" {
		t.Errorf("expected 'Observation', got %q", rt)
	}
	if len(params) != 2 {
		t.Fatalf("expected 2 params, got %d", len(params))
	}
	if params["code"] != "1234" {
		t.Errorf("expected code '1234', got %q", params["code"])
	}
	if params["status"] != "final" {
		t.Errorf("expected status 'final', got %q", params["status"])
	}
}

func TestParseCriteria_EmptyQueryString(t *testing.T) {
	rt, params := ParseCriteria("Condition?")
	if rt != "Condition" {
		t.Errorf("expected 'Condition', got %q", rt)
	}
	if len(params) != 0 {
		t.Errorf("expected 0 params, got %d", len(params))
	}
}

func TestParseCriteria_SingleParam(t *testing.T) {
	rt, params := ParseCriteria("Encounter?status=in-progress")
	if rt != "Encounter" {
		t.Errorf("expected 'Encounter', got %q", rt)
	}
	if params["status"] != "in-progress" {
		t.Errorf("expected status 'in-progress', got %q", params["status"])
	}
}

func TestMatchesCriteria_ResourceTypeMatch(t *testing.T) {
	cs := cachedSubscription{
		resourceType: "Patient",
		params:       map[string]string{},
	}
	event := ResourceEvent{
		ResourceType: "Patient",
		ResourceID:   "p-1",
		Action:       "create",
		Resource:     json.RawMessage(`{"resourceType":"Patient","id":"p-1"}`),
	}
	if !matchesCriteria(cs, event) {
		t.Error("expected match for same resource type")
	}
}

func TestMatchesCriteria_ResourceTypeMismatch(t *testing.T) {
	cs := cachedSubscription{
		resourceType: "Patient",
		params:       map[string]string{},
	}
	event := ResourceEvent{
		ResourceType: "Observation",
		ResourceID:   "o-1",
		Action:       "create",
		Resource:     json.RawMessage(`{"resourceType":"Observation","id":"o-1"}`),
	}
	if matchesCriteria(cs, event) {
		t.Error("expected no match for different resource type")
	}
}

func TestMatchesCriteria_WithParamMatch(t *testing.T) {
	cs := cachedSubscription{
		resourceType: "Observation",
		params:       map[string]string{"status": "final"},
	}
	event := ResourceEvent{
		ResourceType: "Observation",
		ResourceID:   "o-1",
		Action:       "create",
		Resource:     json.RawMessage(`{"resourceType":"Observation","id":"o-1","status":"final"}`),
	}
	if !matchesCriteria(cs, event) {
		t.Error("expected match for matching param")
	}
}

func TestMatchesCriteria_WithParamMismatch(t *testing.T) {
	cs := cachedSubscription{
		resourceType: "Observation",
		params:       map[string]string{"status": "final"},
	}
	event := ResourceEvent{
		ResourceType: "Observation",
		ResourceID:   "o-1",
		Action:       "create",
		Resource:     json.RawMessage(`{"resourceType":"Observation","id":"o-1","status":"preliminary"}`),
	}
	if matchesCriteria(cs, event) {
		t.Error("expected no match for non-matching param")
	}
}

func TestMatchesCriteria_DottedPath(t *testing.T) {
	cs := cachedSubscription{
		resourceType: "Observation",
		params:       map[string]string{"subject.reference": "Patient/p-1"},
	}
	event := ResourceEvent{
		ResourceType: "Observation",
		ResourceID:   "o-1",
		Action:       "create",
		Resource:     json.RawMessage(`{"resourceType":"Observation","id":"o-1","subject":{"reference":"Patient/p-1"}}`),
	}
	if !matchesCriteria(cs, event) {
		t.Error("expected match for dotted path")
	}
}

func TestMatchesCriteria_InvalidJSON(t *testing.T) {
	cs := cachedSubscription{
		resourceType: "Observation",
		params:       map[string]string{"status": "final"},
	}
	event := ResourceEvent{
		ResourceType: "Observation",
		ResourceID:   "o-1",
		Action:       "create",
		Resource:     json.RawMessage(`invalid json`),
	}
	if matchesCriteria(cs, event) {
		t.Error("expected no match for invalid JSON")
	}
}

func TestRetryBackoff(t *testing.T) {
	tests := []struct {
		attempt  int
		expected time.Duration
	}{
		{1, 30 * time.Second},
		{2, 1 * time.Minute},
		{3, 5 * time.Minute},
		{4, 15 * time.Minute},
		{5, 1 * time.Hour},
		{10, 1 * time.Hour},
	}
	for _, tt := range tests {
		got := retryBackoff(tt.attempt)
		if got != tt.expected {
			t.Errorf("attempt %d: expected %v, got %v", tt.attempt, tt.expected, got)
		}
	}
}

func TestExtractField_SimpleString(t *testing.T) {
	resource := map[string]interface{}{"status": "active"}
	if got := extractField(resource, "status"); got != "active" {
		t.Errorf("expected 'active', got %q", got)
	}
}

func TestExtractField_Number(t *testing.T) {
	resource := map[string]interface{}{"count": float64(42)}
	if got := extractField(resource, "count"); got != "42" {
		t.Errorf("expected '42', got %q", got)
	}
}

func TestExtractField_Nested(t *testing.T) {
	resource := map[string]interface{}{
		"subject": map[string]interface{}{
			"reference": "Patient/p-1",
		},
	}
	if got := extractField(resource, "subject.reference"); got != "Patient/p-1" {
		t.Errorf("expected 'Patient/p-1', got %q", got)
	}
}

func TestExtractField_Missing(t *testing.T) {
	resource := map[string]interface{}{"status": "active"}
	if got := extractField(resource, "missing"); got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}

func TestExtractField_Bool(t *testing.T) {
	resource := map[string]interface{}{"active": true}
	if got := extractField(resource, "active"); got != "true" {
		t.Errorf("expected 'true', got %q", got)
	}
}

func TestBuildNotificationBundle(t *testing.T) {
	n := &NotificationRecord{
		ResourceType: "Observation",
		ResourceID:   "obs-1",
		EventType:    "create",
		Payload:      json.RawMessage(`{"resourceType":"Observation","id":"obs-1"}`),
	}
	bundle := buildNotificationBundle(n)
	if bundle["resourceType"] != "Bundle" {
		t.Errorf("expected Bundle resourceType")
	}
	if bundle["type"] != "history" {
		t.Errorf("expected history type")
	}
	entries := bundle["entry"].([]map[string]interface{})
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	req := entries[0]["request"].(map[string]string)
	if req["method"] != "POST" {
		t.Errorf("expected POST for create, got %q", req["method"])
	}
}

func TestActionToMethod(t *testing.T) {
	tests := []struct {
		action   string
		expected string
	}{
		{"create", "POST"},
		{"update", "PUT"},
		{"delete", "DELETE"},
		{"unknown", "PUT"},
	}
	for _, tt := range tests {
		if got := actionToMethod(tt.action); got != tt.expected {
			t.Errorf("action %q: expected %q, got %q", tt.action, tt.expected, got)
		}
	}
}

// -- Mock repo for notification engine tests --

type mockStatusUpdate struct {
	id        uuid.UUID
	status    string
	errorText *string
}

type mockDeleteCall struct {
	before   time.Time
	statuses []string
}

type mockNotifyRepo struct {
	created       []*NotificationRecord
	updated       []*NotificationRecord
	activeSubs    []SubscriptionInfo
	pendingNotifs []*NotificationRecord
	expiredSubs   []ExpiredSubscription
	statusUpdates []mockStatusUpdate
	deleteCalls   []mockDeleteCall
	deleteReturn  int64
}

func (m *mockNotifyRepo) ListActive(_ context.Context) ([]SubscriptionInfo, error) {
	return m.activeSubs, nil
}
func (m *mockNotifyRepo) CreateNotification(_ context.Context, n *NotificationRecord) error {
	n.ID = uuid.New()
	m.created = append(m.created, n)
	return nil
}
func (m *mockNotifyRepo) ListPendingNotifications(_ context.Context, limit int) ([]*NotificationRecord, error) {
	return m.pendingNotifs, nil
}
func (m *mockNotifyRepo) UpdateNotification(_ context.Context, n *NotificationRecord) error {
	m.updated = append(m.updated, n)
	return nil
}
func (m *mockNotifyRepo) UpdateSubscriptionStatus(_ context.Context, id uuid.UUID, status string, errorText *string) error {
	m.statusUpdates = append(m.statusUpdates, mockStatusUpdate{id, status, errorText})
	return nil
}
func (m *mockNotifyRepo) ListExpiredSubscriptions(_ context.Context) ([]ExpiredSubscription, error) {
	return m.expiredSubs, nil
}
func (m *mockNotifyRepo) DeleteOldNotifications(_ context.Context, before time.Time, statuses []string) (int64, error) {
	m.deleteCalls = append(m.deleteCalls, mockDeleteCall{before: before, statuses: statuses})
	return m.deleteReturn, nil
}

// -- OnResourceEvent tests --

func TestOnResourceEvent_MatchCreatesNotification(t *testing.T) {
	repo := &mockNotifyRepo{}
	logger := zerolog.Nop()
	engine := NewNotificationEngine(repo, logger)

	// Manually set cache
	engine.mu.Lock()
	engine.cache = []cachedSubscription{
		{
			info: SubscriptionInfo{
				ID:     uuid.New(),
				FHIRID: "sub-1",
			},
			resourceType: "Observation",
			params:       map[string]string{},
		},
	}
	engine.mu.Unlock()

	event := ResourceEvent{
		ResourceType: "Observation",
		ResourceID:   "obs-1",
		VersionID:    1,
		Action:       "create",
		Resource:     json.RawMessage(`{"resourceType":"Observation","id":"obs-1"}`),
	}
	engine.OnResourceEvent(context.Background(), event)

	if len(repo.created) != 1 {
		t.Fatalf("expected 1 notification, got %d", len(repo.created))
	}
	n := repo.created[0]
	if n.ResourceType != "Observation" {
		t.Errorf("expected resource type 'Observation', got %q", n.ResourceType)
	}
	if n.EventType != "create" {
		t.Errorf("expected event type 'create', got %q", n.EventType)
	}
	if n.Status != "pending" {
		t.Errorf("expected status 'pending', got %q", n.Status)
	}
}

func TestOnResourceEvent_NoMatchSkips(t *testing.T) {
	repo := &mockNotifyRepo{}
	logger := zerolog.Nop()
	engine := NewNotificationEngine(repo, logger)

	engine.mu.Lock()
	engine.cache = []cachedSubscription{
		{
			info: SubscriptionInfo{
				ID:     uuid.New(),
				FHIRID: "sub-1",
			},
			resourceType: "Patient",
			params:       map[string]string{},
		},
	}
	engine.mu.Unlock()

	event := ResourceEvent{
		ResourceType: "Observation",
		ResourceID:   "obs-1",
		Action:       "create",
		Resource:     json.RawMessage(`{"resourceType":"Observation","id":"obs-1"}`),
	}
	engine.OnResourceEvent(context.Background(), event)

	if len(repo.created) != 0 {
		t.Errorf("expected 0 notifications, got %d", len(repo.created))
	}
}

func TestOnResourceEvent_MultipleMatches(t *testing.T) {
	repo := &mockNotifyRepo{}
	logger := zerolog.Nop()
	engine := NewNotificationEngine(repo, logger)

	engine.mu.Lock()
	engine.cache = []cachedSubscription{
		{
			info:         SubscriptionInfo{ID: uuid.New(), FHIRID: "sub-1"},
			resourceType: "Observation",
			params:       map[string]string{},
		},
		{
			info:         SubscriptionInfo{ID: uuid.New(), FHIRID: "sub-2"},
			resourceType: "Observation",
			params:       map[string]string{},
		},
	}
	engine.mu.Unlock()

	event := ResourceEvent{
		ResourceType: "Observation",
		ResourceID:   "obs-1",
		Action:       "create",
		Resource:     json.RawMessage(`{"resourceType":"Observation","id":"obs-1"}`),
	}
	engine.OnResourceEvent(context.Background(), event)

	if len(repo.created) != 2 {
		t.Fatalf("expected 2 notifications, got %d", len(repo.created))
	}
}

// -- Delivery tests --

func TestDeliverOne_Success(t *testing.T) {
	var receivedContentType string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedContentType = r.Header.Get("Content-Type")
		io.Copy(io.Discard, r.Body)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	repo := &mockNotifyRepo{}
	logger := zerolog.Nop()
	engine := NewNotificationEngine(repo, logger)

	n := &NotificationRecord{
		ID:              uuid.New(),
		SubscriptionID:  uuid.New(),
		ResourceType:    "Observation",
		ResourceID:      "obs-1",
		EventType:       "create",
		Status:          "pending",
		Payload:         json.RawMessage(`{"resourceType":"Observation","id":"obs-1"}`),
		ChannelEndpoint: srv.URL,
		ChannelPayload:  "application/fhir+json",
		MaxAttempts:     5,
	}

	engine.deliverOne(context.Background(), n)

	if len(repo.updated) != 1 {
		t.Fatalf("expected 1 update, got %d", len(repo.updated))
	}
	if repo.updated[0].Status != "delivered" {
		t.Errorf("expected status 'delivered', got %q", repo.updated[0].Status)
	}
	if repo.updated[0].DeliveredAt == nil {
		t.Error("expected DeliveredAt to be set")
	}
	if repo.updated[0].AttemptCount != 1 {
		t.Errorf("expected attempt count 1, got %d", repo.updated[0].AttemptCount)
	}
	if receivedContentType != "application/fhir+json" {
		t.Errorf("expected Content-Type 'application/fhir+json', got %q", receivedContentType)
	}
}

func TestDeliverOne_HTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.Copy(io.Discard, r.Body)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	repo := &mockNotifyRepo{}
	logger := zerolog.Nop()
	engine := NewNotificationEngine(repo, logger)

	n := &NotificationRecord{
		ID:              uuid.New(),
		SubscriptionID:  uuid.New(),
		ResourceType:    "Observation",
		ResourceID:      "obs-1",
		EventType:       "create",
		Status:          "pending",
		Payload:         json.RawMessage(`{"resourceType":"Observation","id":"obs-1"}`),
		ChannelEndpoint: srv.URL,
		ChannelPayload:  "application/fhir+json",
		MaxAttempts:     5,
	}

	engine.deliverOne(context.Background(), n)

	if len(repo.updated) != 1 {
		t.Fatalf("expected 1 update, got %d", len(repo.updated))
	}
	if repo.updated[0].AttemptCount != 1 {
		t.Errorf("expected attempt count 1, got %d", repo.updated[0].AttemptCount)
	}
	if repo.updated[0].LastError == nil {
		t.Error("expected last error to be set")
	}
}

func TestDeliverOne_WithHeaders(t *testing.T) {
	var receivedAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedAuth = r.Header.Get("Authorization")
		io.Copy(io.Discard, r.Body)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	repo := &mockNotifyRepo{}
	logger := zerolog.Nop()
	engine := NewNotificationEngine(repo, logger)

	n := &NotificationRecord{
		ID:              uuid.New(),
		SubscriptionID:  uuid.New(),
		ResourceType:    "Observation",
		ResourceID:      "obs-1",
		EventType:       "create",
		Status:          "pending",
		Payload:         json.RawMessage(`{"resourceType":"Observation","id":"obs-1"}`),
		ChannelEndpoint: srv.URL,
		ChannelPayload:  "application/fhir+json",
		ChannelHeaders:  []string{"Authorization: Bearer test123"},
		MaxAttempts:     5,
	}

	engine.deliverOne(context.Background(), n)

	if receivedAuth != "Bearer test123" {
		t.Errorf("expected Authorization header 'Bearer test123', got %q", receivedAuth)
	}
	if len(repo.updated) != 1 || repo.updated[0].Status != "delivered" {
		t.Error("expected notification to be marked as delivered")
	}
}

func TestDeliverPending(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.Copy(io.Discard, r.Body)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	n := &NotificationRecord{
		ID:              uuid.New(),
		SubscriptionID:  uuid.New(),
		ResourceType:    "Observation",
		ResourceID:      "obs-1",
		EventType:       "create",
		Status:          "pending",
		Payload:         json.RawMessage(`{"resourceType":"Observation","id":"obs-1"}`),
		ChannelEndpoint: srv.URL,
		ChannelPayload:  "application/fhir+json",
		MaxAttempts:     5,
	}
	repo := &mockNotifyRepo{
		pendingNotifs: []*NotificationRecord{n},
	}
	logger := zerolog.Nop()
	engine := NewNotificationEngine(repo, logger)

	engine.deliverPending(context.Background())

	if len(repo.updated) != 1 {
		t.Fatalf("expected 1 update, got %d", len(repo.updated))
	}
	if repo.updated[0].Status != "delivered" {
		t.Errorf("expected status 'delivered', got %q", repo.updated[0].Status)
	}
}

// -- markFailed tests --

func TestMarkFailed_Retry(t *testing.T) {
	repo := &mockNotifyRepo{}
	logger := zerolog.Nop()
	engine := NewNotificationEngine(repo, logger)

	n := &NotificationRecord{
		ID:             uuid.New(),
		SubscriptionID: uuid.New(),
		AttemptCount:   0,
		MaxAttempts:    5,
	}

	engine.markFailed(context.Background(), n, "connection refused")

	if len(repo.updated) != 1 {
		t.Fatalf("expected 1 update, got %d", len(repo.updated))
	}
	if n.AttemptCount != 1 {
		t.Errorf("expected attempt count 1, got %d", n.AttemptCount)
	}
	if n.Status == "abandoned" {
		t.Error("should not be abandoned after 1 attempt")
	}
	if n.LastError == nil || *n.LastError != "connection refused" {
		t.Error("expected last error to be set")
	}
	if len(repo.statusUpdates) != 0 {
		t.Error("should not update subscription status on retry")
	}
}

func TestMarkFailed_MaxAttempts(t *testing.T) {
	repo := &mockNotifyRepo{}
	logger := zerolog.Nop()
	engine := NewNotificationEngine(repo, logger)

	n := &NotificationRecord{
		ID:             uuid.New(),
		SubscriptionID: uuid.New(),
		AttemptCount:   4, // will become 5 which == MaxAttempts
		MaxAttempts:    5,
	}

	engine.markFailed(context.Background(), n, "timeout")

	if n.Status != "abandoned" {
		t.Errorf("expected status 'abandoned', got %q", n.Status)
	}
	if n.AttemptCount != 5 {
		t.Errorf("expected attempt count 5, got %d", n.AttemptCount)
	}
	if len(repo.statusUpdates) != 1 {
		t.Fatalf("expected 1 status update, got %d", len(repo.statusUpdates))
	}
	if repo.statusUpdates[0].status != "error" {
		t.Errorf("expected subscription status 'error', got %q", repo.statusUpdates[0].status)
	}
	if repo.statusUpdates[0].errorText == nil {
		t.Error("expected error text in status update")
	}
}

// -- Handshake tests --

func TestPerformHandshake_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		ct := r.Header.Get("Content-Type")
		if ct != "application/fhir+json" {
			t.Errorf("expected Content-Type 'application/fhir+json', got %q", ct)
		}
		io.Copy(io.Discard, r.Body)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	logger := zerolog.Nop()
	engine := NewNotificationEngine(&mockNotifyRepo{}, logger)
	err := engine.PerformHandshake(context.Background(), srv.URL, []string{"Authorization: Bearer test"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPerformHandshake_Failure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.Copy(io.Discard, r.Body)
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer srv.Close()

	logger := zerolog.Nop()
	engine := NewNotificationEngine(&mockNotifyRepo{}, logger)
	err := engine.PerformHandshake(context.Background(), srv.URL, nil)
	if err == nil {
		t.Fatal("expected error for non-2xx response")
	}
}

func TestPerformHandshake_WithHeaders(t *testing.T) {
	var receivedAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedAuth = r.Header.Get("Authorization")
		io.Copy(io.Discard, r.Body)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	logger := zerolog.Nop()
	engine := NewNotificationEngine(&mockNotifyRepo{}, logger)
	err := engine.PerformHandshake(context.Background(), srv.URL, []string{"Authorization: Bearer mytoken"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if receivedAuth != "Bearer mytoken" {
		t.Errorf("expected Authorization header 'Bearer mytoken', got %q", receivedAuth)
	}
}

// -- Expiry tests --

func TestExpireSubscriptions(t *testing.T) {
	subID := uuid.New()
	repo := &mockNotifyRepo{
		expiredSubs: []ExpiredSubscription{
			{ID: subID, FHIRID: "sub-expired"},
		},
	}
	logger := zerolog.Nop()
	engine := NewNotificationEngine(repo, logger)

	engine.expireSubscriptions(context.Background())

	if len(repo.statusUpdates) != 1 {
		t.Fatalf("expected 1 status update, got %d", len(repo.statusUpdates))
	}
	if repo.statusUpdates[0].status != "off" {
		t.Errorf("expected status 'off', got %q", repo.statusUpdates[0].status)
	}
	if repo.statusUpdates[0].id != subID {
		t.Error("status update for wrong subscription")
	}
}

func TestExpireSubscriptions_None(t *testing.T) {
	repo := &mockNotifyRepo{}
	logger := zerolog.Nop()
	engine := NewNotificationEngine(repo, logger)

	engine.expireSubscriptions(context.Background())

	if len(repo.statusUpdates) != 0 {
		t.Errorf("expected 0 status updates, got %d", len(repo.statusUpdates))
	}
}

// -- Cache tests --

func TestRefreshCache(t *testing.T) {
	repo := &mockNotifyRepo{
		activeSubs: []SubscriptionInfo{
			{
				ID:       uuid.New(),
				FHIRID:   "sub-1",
				Criteria: "Observation?code=1234",
			},
			{
				ID:       uuid.New(),
				FHIRID:   "sub-2",
				Criteria: "Patient",
			},
		},
	}
	logger := zerolog.Nop()
	engine := NewNotificationEngine(repo, logger)

	engine.RefreshCache(context.Background())

	engine.mu.RLock()
	defer engine.mu.RUnlock()
	if len(engine.cache) != 2 {
		t.Fatalf("expected 2 cached subscriptions, got %d", len(engine.cache))
	}
	if engine.cache[0].resourceType != "Observation" {
		t.Errorf("expected resource type 'Observation', got %q", engine.cache[0].resourceType)
	}
	if engine.cache[0].params["code"] != "1234" {
		t.Errorf("expected param code '1234', got %q", engine.cache[0].params["code"])
	}
	if engine.cache[1].resourceType != "Patient" {
		t.Errorf("expected resource type 'Patient', got %q", engine.cache[1].resourceType)
	}
	if len(engine.cache[1].params) != 0 {
		t.Errorf("expected 0 params for Patient, got %d", len(engine.cache[1].params))
	}
}

// -- Start loop test --

func TestStart_ContextCancel(t *testing.T) {
	repo := &mockNotifyRepo{}
	logger := zerolog.Nop()
	engine := NewNotificationEngine(repo, logger)
	engine.CacheRefreshInterval = 100 * time.Millisecond
	engine.DeliveryInterval = 100 * time.Millisecond

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		engine.Start(ctx)
		close(done)
	}()

	// Let it run briefly
	time.Sleep(250 * time.Millisecond)
	cancel()

	select {
	case <-done:
		// expected
	case <-time.After(2 * time.Second):
		t.Fatal("Start did not return after context cancellation")
	}
}

func TestCleanupOldNotifications(t *testing.T) {
	repo := &mockNotifyRepo{deleteReturn: 5}
	logger := zerolog.Nop()
	engine := NewNotificationEngine(repo, logger)

	engine.cleanupOldNotifications(context.Background())

	if len(repo.deleteCalls) != 2 {
		t.Fatalf("expected 2 delete calls (delivered + abandoned), got %d", len(repo.deleteCalls))
	}

	// First call: delivered notifications older than 30 days
	if len(repo.deleteCalls[0].statuses) != 1 || repo.deleteCalls[0].statuses[0] != "delivered" {
		t.Errorf("first call should delete 'delivered', got %v", repo.deleteCalls[0].statuses)
	}
	deliveredAge := time.Since(repo.deleteCalls[0].before)
	if deliveredAge < 29*24*time.Hour || deliveredAge > 31*24*time.Hour {
		t.Errorf("delivered cutoff should be ~30 days ago, got %v", deliveredAge)
	}

	// Second call: abandoned notifications older than 90 days
	if len(repo.deleteCalls[1].statuses) != 1 || repo.deleteCalls[1].statuses[0] != "abandoned" {
		t.Errorf("second call should delete 'abandoned', got %v", repo.deleteCalls[1].statuses)
	}
	abandonedAge := time.Since(repo.deleteCalls[1].before)
	if abandonedAge < 89*24*time.Hour || abandonedAge > 91*24*time.Hour {
		t.Errorf("abandoned cutoff should be ~90 days ago, got %v", abandonedAge)
	}
}
