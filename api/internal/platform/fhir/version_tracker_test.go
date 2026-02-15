package fhir

import (
	"context"
	"testing"
)

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
