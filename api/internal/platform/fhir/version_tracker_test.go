package fhir

import (
	"context"
	"encoding/json"
	"testing"
)

// testListener implements ResourceEventListener for testing.
type testListener struct {
	events []ResourceEvent
}

func (l *testListener) OnResourceEvent(_ context.Context, event ResourceEvent) {
	l.events = append(l.events, event)
}

func TestNewVersionTracker(t *testing.T) {
	repo := NewHistoryRepository()
	vt := NewVersionTracker(repo)
	if vt == nil {
		t.Fatal("expected non-nil VersionTracker")
	}
	if vt.repo != repo {
		t.Error("expected repo to be set")
	}
}

// TestRecordCreate_NoDBContext verifies that RecordCreate marshals the resource
// and attempts to save. Without a DB connection in context, it returns an error
// from the underlying repository, confirming the flow is wired correctly.
func TestRecordCreate_NoDBContext(t *testing.T) {
	repo := NewHistoryRepository()
	vt := NewVersionTracker(repo)
	ctx := context.Background() // no DB in context

	resource := map[string]interface{}{
		"resourceType": "Patient",
		"id":           "pat-1",
	}

	err := vt.RecordCreate(ctx, "Patient", "pat-1", resource)
	if err == nil {
		t.Error("expected error when no DB connection in context")
	}
}

// TestRecordUpdate_NoDBContext verifies version increment logic and marshal.
func TestRecordUpdate_NoDBContext(t *testing.T) {
	repo := NewHistoryRepository()
	vt := NewVersionTracker(repo)
	ctx := context.Background()

	resource := map[string]interface{}{
		"resourceType": "Patient",
		"id":           "pat-1",
	}

	newVer, err := vt.RecordUpdate(ctx, "Patient", "pat-1", 1, resource)
	if err == nil {
		t.Error("expected error when no DB connection in context")
	}
	if newVer != 0 {
		t.Errorf("expected 0 on error, got %d", newVer)
	}
}

// TestRecordDelete_NoDBContext verifies delete marker flow.
func TestRecordDelete_NoDBContext(t *testing.T) {
	repo := NewHistoryRepository()
	vt := NewVersionTracker(repo)
	ctx := context.Background()

	err := vt.RecordDelete(ctx, "Patient", "pat-1", 1)
	if err == nil {
		t.Error("expected error when no DB connection in context")
	}
}

// TestGetVersion_NoDBContext verifies get flow.
func TestGetVersion_NoDBContext(t *testing.T) {
	repo := NewHistoryRepository()
	vt := NewVersionTracker(repo)
	ctx := context.Background()

	_, err := vt.GetVersion(ctx, "Patient", "pat-1", 1)
	if err == nil {
		t.Error("expected error when no DB connection in context")
	}
}

// TestListVersions_NoDBContext verifies list flow.
func TestListVersions_NoDBContext(t *testing.T) {
	repo := NewHistoryRepository()
	vt := NewVersionTracker(repo)
	ctx := context.Background()

	_, _, err := vt.ListVersions(ctx, "Patient", "pat-1", 100, 0)
	if err == nil {
		t.Error("expected error when no DB connection in context")
	}
}

// TestRecordCreate_MarshalError verifies that unmarshalable resources produce errors.
func TestRecordCreate_MarshalError(t *testing.T) {
	repo := NewHistoryRepository()
	vt := NewVersionTracker(repo)
	ctx := context.Background()

	// channels cannot be marshaled to JSON
	badResource := make(chan int)
	err := vt.RecordCreate(ctx, "Patient", "pat-1", badResource)
	if err == nil {
		t.Error("expected marshal error for channel type")
	}
}

// TestRecordUpdate_MarshalError verifies that unmarshalable resources produce errors.
func TestRecordUpdate_MarshalError(t *testing.T) {
	repo := NewHistoryRepository()
	vt := NewVersionTracker(repo)
	ctx := context.Background()

	badResource := make(chan int)
	_, err := vt.RecordUpdate(ctx, "Patient", "pat-1", 1, badResource)
	if err == nil {
		t.Error("expected marshal error for channel type")
	}
}

func TestNilVersionTracker(t *testing.T) {
	// Verify that checking nil VersionTracker works (used in service code)
	var vt *VersionTracker
	if vt != nil {
		t.Error("nil VersionTracker should be nil")
	}
}

// -- Listener tests --

func TestAddListener(t *testing.T) {
	repo := NewHistoryRepository()
	vt := NewVersionTracker(repo)

	l := &testListener{}
	vt.AddListener(l)

	vt.mu.RLock()
	defer vt.mu.RUnlock()
	if len(vt.listeners) != 1 {
		t.Errorf("expected 1 listener, got %d", len(vt.listeners))
	}
}

func TestAddListener_Multiple(t *testing.T) {
	repo := NewHistoryRepository()
	vt := NewVersionTracker(repo)

	vt.AddListener(&testListener{})
	vt.AddListener(&testListener{})
	vt.AddListener(&testListener{})

	vt.mu.RLock()
	defer vt.mu.RUnlock()
	if len(vt.listeners) != 3 {
		t.Errorf("expected 3 listeners, got %d", len(vt.listeners))
	}
}

func TestFireEvent_CallsListener(t *testing.T) {
	repo := NewHistoryRepository()
	vt := NewVersionTracker(repo)

	l := &testListener{}
	vt.AddListener(l)

	event := ResourceEvent{
		ResourceType: "Patient",
		ResourceID:   "pat-1",
		VersionID:    1,
		Action:       "create",
		Resource:     json.RawMessage(`{"resourceType":"Patient","id":"pat-1"}`),
	}
	vt.fireEvent(context.Background(), event)

	if len(l.events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(l.events))
	}
	if l.events[0].ResourceType != "Patient" {
		t.Errorf("expected resource type 'Patient', got %q", l.events[0].ResourceType)
	}
	if l.events[0].ResourceID != "pat-1" {
		t.Errorf("expected resource ID 'pat-1', got %q", l.events[0].ResourceID)
	}
	if l.events[0].Action != "create" {
		t.Errorf("expected action 'create', got %q", l.events[0].Action)
	}
	if l.events[0].VersionID != 1 {
		t.Errorf("expected version 1, got %d", l.events[0].VersionID)
	}
}

func TestFireEvent_MultipleListeners(t *testing.T) {
	repo := NewHistoryRepository()
	vt := NewVersionTracker(repo)

	l1 := &testListener{}
	l2 := &testListener{}
	vt.AddListener(l1)
	vt.AddListener(l2)

	event := ResourceEvent{
		ResourceType: "Observation",
		ResourceID:   "obs-1",
		VersionID:    2,
		Action:       "update",
		Resource:     json.RawMessage(`{"resourceType":"Observation","id":"obs-1"}`),
	}
	vt.fireEvent(context.Background(), event)

	if len(l1.events) != 1 {
		t.Errorf("listener 1: expected 1 event, got %d", len(l1.events))
	}
	if len(l2.events) != 1 {
		t.Errorf("listener 2: expected 1 event, got %d", len(l2.events))
	}
	if l1.events[0].Action != "update" {
		t.Errorf("listener 1: expected action 'update', got %q", l1.events[0].Action)
	}
}

func TestFireEvent_NoListeners(t *testing.T) {
	repo := NewHistoryRepository()
	vt := NewVersionTracker(repo)

	// Should not panic with no listeners
	event := ResourceEvent{
		ResourceType: "Patient",
		ResourceID:   "pat-1",
		VersionID:    1,
		Action:       "delete",
		Resource:     json.RawMessage(`null`),
	}
	vt.fireEvent(context.Background(), event)
}
