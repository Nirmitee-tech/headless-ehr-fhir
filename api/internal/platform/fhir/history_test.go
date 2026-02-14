package fhir

import (
	"encoding/json"
	"testing"
	"time"
)

func TestNewHistoryBundle(t *testing.T) {
	now := time.Now().UTC()
	entries := []*HistoryEntry{
		{
			ResourceType: "Patient",
			ResourceID:   "p1",
			VersionID:    2,
			Resource:     json.RawMessage(`{"resourceType":"Patient","id":"p1"}`),
			Action:       "update",
			Timestamp:    now,
		},
		{
			ResourceType: "Patient",
			ResourceID:   "p1",
			VersionID:    1,
			Resource:     json.RawMessage(`{"resourceType":"Patient","id":"p1"}`),
			Action:       "create",
			Timestamp:    now.Add(-time.Hour),
		},
	}

	bundle := NewHistoryBundle(entries, 2, "/fhir")

	if bundle.Type != "history" {
		t.Errorf("bundle type = %q, want 'history'", bundle.Type)
	}
	if *bundle.Total != 2 {
		t.Errorf("total = %d, want 2", *bundle.Total)
	}
	if len(bundle.Entry) != 2 {
		t.Fatalf("entries = %d, want 2", len(bundle.Entry))
	}

	// Check first entry (update)
	if bundle.Entry[0].Request.Method != "PUT" {
		t.Errorf("entry[0] method = %q, want PUT", bundle.Entry[0].Request.Method)
	}
	if bundle.Entry[0].Response.Status != "200 OK" {
		t.Errorf("entry[0] status = %q, want '200 OK'", bundle.Entry[0].Response.Status)
	}

	// Check second entry (create)
	if bundle.Entry[1].Request.Method != "POST" {
		t.Errorf("entry[1] method = %q, want POST", bundle.Entry[1].Request.Method)
	}
	if bundle.Entry[1].Response.Status != "201 Created" {
		t.Errorf("entry[1] status = %q, want '201 Created'", bundle.Entry[1].Response.Status)
	}
}

func TestNewHistoryBundle_DeleteAction(t *testing.T) {
	now := time.Now().UTC()
	entries := []*HistoryEntry{
		{
			ResourceType: "Patient",
			ResourceID:   "p1",
			VersionID:    3,
			Resource:     json.RawMessage(`{}`),
			Action:       "delete",
			Timestamp:    now,
		},
	}

	bundle := NewHistoryBundle(entries, 1, "/fhir")
	if bundle.Entry[0].Request.Method != "DELETE" {
		t.Errorf("delete entry method = %q, want DELETE", bundle.Entry[0].Request.Method)
	}
	if bundle.Entry[0].Response.Status != "204 No Content" {
		t.Errorf("delete entry status = %q", bundle.Entry[0].Response.Status)
	}
}

func TestNewHistoryBundle_Empty(t *testing.T) {
	bundle := NewHistoryBundle(nil, 0, "/fhir")
	if bundle.Type != "history" {
		t.Error("empty history should still be type 'history'")
	}
	if *bundle.Total != 0 {
		t.Error("empty history total should be 0")
	}
}

func TestNewHistoryBundle_FullURL(t *testing.T) {
	entries := []*HistoryEntry{
		{
			ResourceType: "Observation",
			ResourceID:   "obs-1",
			VersionID:    5,
			Resource:     json.RawMessage(`{}`),
			Action:       "update",
			Timestamp:    time.Now(),
		},
	}

	bundle := NewHistoryBundle(entries, 1, "/fhir")
	expected := "/fhir/Observation/obs-1/_history/5"
	if bundle.Entry[0].FullURL != expected {
		t.Errorf("fullUrl = %q, want %q", bundle.Entry[0].FullURL, expected)
	}
}

func TestNewHistoryBundle_RequestURL(t *testing.T) {
	now := time.Now().UTC()
	entries := []*HistoryEntry{
		{
			ResourceType: "Condition",
			ResourceID:   "cond-1",
			VersionID:    1,
			Resource:     json.RawMessage(`{}`),
			Action:       "create",
			Timestamp:    now,
		},
	}

	bundle := NewHistoryBundle(entries, 1, "/fhir")
	if bundle.Entry[0].Request == nil {
		t.Fatal("expected request to be set")
	}
	expectedURL := "Condition/cond-1"
	if bundle.Entry[0].Request.URL != expectedURL {
		t.Errorf("request.url = %q, want %q", bundle.Entry[0].Request.URL, expectedURL)
	}
}

func TestNewHistoryBundle_ResponseLastModified(t *testing.T) {
	now := time.Now().UTC()
	entries := []*HistoryEntry{
		{
			ResourceType: "Patient",
			ResourceID:   "p1",
			VersionID:    1,
			Resource:     json.RawMessage(`{}`),
			Action:       "create",
			Timestamp:    now,
		},
	}

	bundle := NewHistoryBundle(entries, 1, "/fhir")
	if bundle.Entry[0].Response == nil {
		t.Fatal("expected response to be set")
	}
	if bundle.Entry[0].Response.LastModified == nil {
		t.Fatal("expected lastModified to be set")
	}
	if !bundle.Entry[0].Response.LastModified.Equal(now) {
		t.Errorf("lastModified = %v, want %v", bundle.Entry[0].Response.LastModified, now)
	}
}

func TestNewHistoryBundle_Timestamp(t *testing.T) {
	bundle := NewHistoryBundle(nil, 0, "/fhir")
	if bundle.Timestamp == nil {
		t.Fatal("expected bundle timestamp to be set")
	}
	// Timestamp should be recent (within last second)
	if time.Since(*bundle.Timestamp) > time.Second {
		t.Errorf("timestamp too old: %v", bundle.Timestamp)
	}
}

func TestNewHistoryBundle_ResourceType(t *testing.T) {
	bundle := NewHistoryBundle(nil, 0, "/fhir")
	if bundle.ResourceType != "Bundle" {
		t.Errorf("resourceType = %q, want %q", bundle.ResourceType, "Bundle")
	}
}

func TestNewHistoryRepository(t *testing.T) {
	repo := NewHistoryRepository()
	if repo == nil {
		t.Fatal("expected non-nil HistoryRepository")
	}
}

func TestNewHistoryBundle_MultipleActions(t *testing.T) {
	now := time.Now().UTC()
	entries := []*HistoryEntry{
		{ResourceType: "Patient", ResourceID: "p1", VersionID: 3, Resource: json.RawMessage(`{}`), Action: "delete", Timestamp: now},
		{ResourceType: "Patient", ResourceID: "p1", VersionID: 2, Resource: json.RawMessage(`{}`), Action: "update", Timestamp: now.Add(-time.Hour)},
		{ResourceType: "Patient", ResourceID: "p1", VersionID: 1, Resource: json.RawMessage(`{}`), Action: "create", Timestamp: now.Add(-2 * time.Hour)},
	}

	bundle := NewHistoryBundle(entries, 3, "/fhir")
	if len(bundle.Entry) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(bundle.Entry))
	}

	// Verify each entry has correct method/status mapping
	expectedMethods := []string{"DELETE", "PUT", "POST"}
	expectedStatuses := []string{"204 No Content", "200 OK", "201 Created"}

	for i, entry := range bundle.Entry {
		if entry.Request.Method != expectedMethods[i] {
			t.Errorf("entry[%d] method = %q, want %q", i, entry.Request.Method, expectedMethods[i])
		}
		if entry.Response.Status != expectedStatuses[i] {
			t.Errorf("entry[%d] status = %q, want %q", i, entry.Response.Status, expectedStatuses[i])
		}
	}
}

func TestNewHistoryBundle_UnknownAction(t *testing.T) {
	now := time.Now().UTC()
	entries := []*HistoryEntry{
		{ResourceType: "Patient", ResourceID: "p1", VersionID: 1, Resource: json.RawMessage(`{}`), Action: "unknown_action", Timestamp: now},
	}

	bundle := NewHistoryBundle(entries, 1, "/fhir")
	// Unknown action defaults to PUT / 200 OK
	if bundle.Entry[0].Request.Method != "PUT" {
		t.Errorf("unknown action method = %q, want PUT", bundle.Entry[0].Request.Method)
	}
	if bundle.Entry[0].Response.Status != "200 OK" {
		t.Errorf("unknown action status = %q, want %q", bundle.Entry[0].Response.Status, "200 OK")
	}
}
