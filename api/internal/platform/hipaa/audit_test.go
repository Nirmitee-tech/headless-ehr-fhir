package hipaa

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestNewReadEvent(t *testing.T) {
	agentID := uuid.New()
	entityID := uuid.New()

	event := NewReadEvent(agentID, "Dr. Smith", "Patient", entityID)

	if event.FHIRId == "" {
		t.Error("expected non-empty FHIR ID")
	}
	if event.TypeCode != "rest" {
		t.Errorf("expected type_code 'rest', got %q", event.TypeCode)
	}
	if event.TypeDisplay != "RESTful Operation" {
		t.Errorf("expected type_display 'RESTful Operation', got %q", event.TypeDisplay)
	}
	if event.SubtypeCode != "read" {
		t.Errorf("expected subtype_code 'read', got %q", event.SubtypeCode)
	}
	if event.SubtypeDisplay != "Read" {
		t.Errorf("expected subtype_display 'Read', got %q", event.SubtypeDisplay)
	}
	if event.Action != "R" {
		t.Errorf("expected action 'R', got %q", event.Action)
	}
	if event.Outcome != "0" {
		t.Errorf("expected outcome '0', got %q", event.Outcome)
	}
	if event.AgentWhoID == nil || *event.AgentWhoID != agentID {
		t.Errorf("expected agent_who_id %s", agentID)
	}
	if event.AgentWhoDisplay != "Dr. Smith" {
		t.Errorf("expected agent_who_display 'Dr. Smith', got %q", event.AgentWhoDisplay)
	}
	if !event.AgentRequestor {
		t.Error("expected agent_requestor to be true")
	}
	if event.EntityWhatType != "Patient" {
		t.Errorf("expected entity_what_type 'Patient', got %q", event.EntityWhatType)
	}
	if event.EntityWhatID == nil || *event.EntityWhatID != entityID {
		t.Errorf("expected entity_what_id %s", entityID)
	}
	if event.PurposeCode != "TREAT" {
		t.Errorf("expected purpose_of_use_code 'TREAT', got %q", event.PurposeCode)
	}
	if event.PurposeDisplay != "Treatment" {
		t.Errorf("expected purpose_of_use_display 'Treatment', got %q", event.PurposeDisplay)
	}
	if event.Recorded.IsZero() {
		t.Error("expected recorded timestamp to be set")
	}
}

func TestNewWriteEvent(t *testing.T) {
	agentID := uuid.New()
	entityID := uuid.New()

	event := NewWriteEvent(agentID, "Nurse Johnson", "Observation", entityID)

	if event.FHIRId == "" {
		t.Error("expected non-empty FHIR ID")
	}
	if event.TypeCode != "rest" {
		t.Errorf("expected type_code 'rest', got %q", event.TypeCode)
	}
	if event.TypeDisplay != "RESTful Operation" {
		t.Errorf("expected type_display 'RESTful Operation', got %q", event.TypeDisplay)
	}
	if event.SubtypeCode != "create" {
		t.Errorf("expected subtype_code 'create', got %q", event.SubtypeCode)
	}
	if event.SubtypeDisplay != "Create" {
		t.Errorf("expected subtype_display 'Create', got %q", event.SubtypeDisplay)
	}
	if event.Action != "C" {
		t.Errorf("expected action 'C', got %q", event.Action)
	}
	if event.Outcome != "0" {
		t.Errorf("expected outcome '0', got %q", event.Outcome)
	}
	if event.AgentWhoID == nil || *event.AgentWhoID != agentID {
		t.Errorf("expected agent_who_id %s", agentID)
	}
	if event.AgentWhoDisplay != "Nurse Johnson" {
		t.Errorf("expected agent_who_display 'Nurse Johnson', got %q", event.AgentWhoDisplay)
	}
	if !event.AgentRequestor {
		t.Error("expected agent_requestor to be true")
	}
	if event.EntityWhatType != "Observation" {
		t.Errorf("expected entity_what_type 'Observation', got %q", event.EntityWhatType)
	}
	if event.EntityWhatID == nil || *event.EntityWhatID != entityID {
		t.Errorf("expected entity_what_id %s", entityID)
	}
	if event.PurposeCode != "TREAT" {
		t.Errorf("expected purpose_of_use_code 'TREAT', got %q", event.PurposeCode)
	}
	if event.Recorded.IsZero() {
		t.Error("expected recorded timestamp to be set")
	}
}

func TestNewDeleteEvent(t *testing.T) {
	agentID := uuid.New()
	entityID := uuid.New()

	event := NewDeleteEvent(agentID, "Admin User", "DocumentReference", entityID)

	if event.FHIRId == "" {
		t.Error("expected non-empty FHIR ID")
	}
	if event.TypeCode != "rest" {
		t.Errorf("expected type_code 'rest', got %q", event.TypeCode)
	}
	if event.TypeDisplay != "RESTful Operation" {
		t.Errorf("expected type_display 'RESTful Operation', got %q", event.TypeDisplay)
	}
	if event.SubtypeCode != "delete" {
		t.Errorf("expected subtype_code 'delete', got %q", event.SubtypeCode)
	}
	if event.SubtypeDisplay != "Delete" {
		t.Errorf("expected subtype_display 'Delete', got %q", event.SubtypeDisplay)
	}
	if event.Action != "D" {
		t.Errorf("expected action 'D', got %q", event.Action)
	}
	if event.Outcome != "0" {
		t.Errorf("expected outcome '0', got %q", event.Outcome)
	}
	if event.AgentWhoID == nil || *event.AgentWhoID != agentID {
		t.Errorf("expected agent_who_id %s", agentID)
	}
	if event.AgentWhoDisplay != "Admin User" {
		t.Errorf("expected agent_who_display 'Admin User', got %q", event.AgentWhoDisplay)
	}
	if !event.AgentRequestor {
		t.Error("expected agent_requestor to be true")
	}
	if event.EntityWhatType != "DocumentReference" {
		t.Errorf("expected entity_what_type 'DocumentReference', got %q", event.EntityWhatType)
	}
	if event.EntityWhatID == nil || *event.EntityWhatID != entityID {
		t.Errorf("expected entity_what_id %s", entityID)
	}
	if event.PurposeCode != "TREAT" {
		t.Errorf("expected purpose_of_use_code 'TREAT', got %q", event.PurposeCode)
	}
	if event.PurposeDisplay != "Treatment" {
		t.Errorf("expected purpose_of_use_display 'Treatment', got %q", event.PurposeDisplay)
	}
	if event.Recorded.IsZero() {
		t.Error("expected recorded timestamp to be set")
	}
}

func TestNewAuditLogger(t *testing.T) {
	logger := NewAuditLogger(nil)
	if logger == nil {
		t.Fatal("expected non-nil AuditLogger")
	}
	if logger.pool != nil {
		t.Error("expected nil pool when passed nil")
	}
}

func TestNewReadEvent_UniqueIDs(t *testing.T) {
	agentID := uuid.New()
	entityID := uuid.New()

	event1 := NewReadEvent(agentID, "Dr. Smith", "Patient", entityID)
	event2 := NewReadEvent(agentID, "Dr. Smith", "Patient", entityID)

	if event1.FHIRId == event2.FHIRId {
		t.Error("expected unique FHIR IDs for each event")
	}
}

func TestNewWriteEvent_Fields(t *testing.T) {
	agentID := uuid.New()
	entityID := uuid.New()

	event := NewWriteEvent(agentID, "Nurse", "Observation", entityID)

	// Verify it doesn't share the same FHIRId with any other event factory
	if event.SubtypeCode != "create" {
		t.Errorf("expected SubtypeCode 'create', got %q", event.SubtypeCode)
	}
	if event.SubtypeDisplay != "Create" {
		t.Errorf("expected SubtypeDisplay 'Create', got %q", event.SubtypeDisplay)
	}
	if event.Action != "C" {
		t.Errorf("expected Action 'C', got %q", event.Action)
	}
}

func TestNewDeleteEvent_Fields(t *testing.T) {
	agentID := uuid.New()
	entityID := uuid.New()

	event := NewDeleteEvent(agentID, "Admin", "MedicationRequest", entityID)

	if event.SubtypeCode != "delete" {
		t.Errorf("expected SubtypeCode 'delete', got %q", event.SubtypeCode)
	}
	if event.Action != "D" {
		t.Errorf("expected Action 'D', got %q", event.Action)
	}
	if event.EntityWhatType != "MedicationRequest" {
		t.Errorf("expected EntityWhatType 'MedicationRequest', got %q", event.EntityWhatType)
	}
}

func TestAuditEvent_DefaultTimestamp(t *testing.T) {
	agentID := uuid.New()
	entityID := uuid.New()

	event := NewReadEvent(agentID, "Dr. Smith", "Patient", entityID)

	if event.Recorded.IsZero() {
		t.Error("expected Recorded to be set automatically")
	}
	// Recorded should be recent
	if time.Since(event.Recorded) > time.Second {
		t.Errorf("Recorded timestamp too old: %v", event.Recorded)
	}
}

func TestAuditEvent_FieldDefaults(t *testing.T) {
	agentID := uuid.New()
	entityID := uuid.New()

	event := NewReadEvent(agentID, "Dr. Smith", "Patient", entityID)

	// Check fields common to all event factory functions
	if event.TypeCode != "rest" {
		t.Errorf("TypeCode = %q, want 'rest'", event.TypeCode)
	}
	if event.TypeDisplay != "RESTful Operation" {
		t.Errorf("TypeDisplay = %q, want 'RESTful Operation'", event.TypeDisplay)
	}
	if event.Outcome != "0" {
		t.Errorf("Outcome = %q, want '0'", event.Outcome)
	}
	if !event.AgentRequestor {
		t.Error("AgentRequestor should be true")
	}
	if event.PurposeCode != "TREAT" {
		t.Errorf("PurposeCode = %q, want 'TREAT'", event.PurposeCode)
	}
	if event.PurposeDisplay != "Treatment" {
		t.Errorf("PurposeDisplay = %q, want 'Treatment'", event.PurposeDisplay)
	}
}

func TestPHIAccessLog_Struct(t *testing.T) {
	// Verify the PHIAccessLog struct can be instantiated and has expected zero values
	log := PHIAccessLog{}
	if log.IsBreakGlass {
		t.Error("expected IsBreakGlass to default to false")
	}
	if !log.AccessedAt.IsZero() {
		t.Error("expected AccessedAt to be zero")
	}
	if log.Action != "" {
		t.Error("expected Action to be empty")
	}
}
