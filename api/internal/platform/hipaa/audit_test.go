package hipaa

import (
	"testing"

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
