package task

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestTask_ToFHIR(t *testing.T) {
	patientID := uuid.New()
	now := time.Now().UTC()
	tk := &Task{
		ID:            uuid.New(),
		FHIRID:        "task-123",
		Status:        "requested",
		Intent:        "order",
		ForPatientID:  patientID,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	result := tk.ToFHIR()

	if result["resourceType"] != "Task" {
		t.Errorf("expected resourceType 'Task', got %v", result["resourceType"])
	}
	if result["id"] != "task-123" {
		t.Errorf("expected id 'task-123', got %v", result["id"])
	}
	if result["status"] != "requested" {
		t.Errorf("expected status 'requested', got %v", result["status"])
	}
	if result["intent"] != "order" {
		t.Errorf("expected intent 'order', got %v", result["intent"])
	}

	// Verify subject/for reference is populated
	raw, _ := json.Marshal(result["for"])
	subjectJSON := string(raw)
	if subjectJSON == "" {
		t.Error("expected 'for' reference to be populated")
	}
}

func TestTask_ToFHIR_AllFields(t *testing.T) {
	patientID := uuid.New()
	encounterID := uuid.New()
	requesterID := uuid.New()
	ownerID := uuid.New()
	now := time.Now().UTC()
	later := now.Add(24 * time.Hour)
	reps := 3

	statusReason := "Patient unavailable"
	priority := "urgent"
	codeValue := "approve"
	codeDisplay := "Approve the action"
	codeSystem := "http://hl7.org/fhir/CodeSystem/task-code"
	description := "Approve the medication request"
	focusType := "MedicationRequest"
	focusID := uuid.New().String()
	reasonCode := "workflow"
	reasonDisplay := "Workflow requirement"
	note := "Please review urgently"
	inputJSON := json.RawMessage(`[{"type":{"text":"input1"},"valueString":"val1"}]`)
	outputJSON := json.RawMessage(`[{"type":{"text":"output1"},"valueString":"result1"}]`)

	tk := &Task{
		ID:                     uuid.New(),
		FHIRID:                 "task-full",
		Status:                 "in-progress",
		StatusReason:           &statusReason,
		Intent:                 "order",
		Priority:               &priority,
		CodeValue:              &codeValue,
		CodeDisplay:            &codeDisplay,
		CodeSystem:             &codeSystem,
		Description:            &description,
		FocusResourceType:      &focusType,
		FocusResourceID:        &focusID,
		ForPatientID:           patientID,
		EncounterID:            &encounterID,
		AuthoredOn:             &now,
		LastModified:           &later,
		RequesterID:            &requesterID,
		OwnerID:                &ownerID,
		ReasonCode:             &reasonCode,
		ReasonDisplay:          &reasonDisplay,
		Note:                   &note,
		RestrictionRepetitions: &reps,
		RestrictionPeriodStart: &now,
		RestrictionPeriodEnd:   &later,
		InputJSON:              &inputJSON,
		OutputJSON:             &outputJSON,
		CreatedAt:              now,
		UpdatedAt:              later,
	}

	result := tk.ToFHIR()

	// Check all required fields
	if result["resourceType"] != "Task" {
		t.Errorf("expected resourceType 'Task', got %v", result["resourceType"])
	}
	if result["status"] != "in-progress" {
		t.Errorf("expected status 'in-progress', got %v", result["status"])
	}
	if result["intent"] != "order" {
		t.Errorf("expected intent 'order', got %v", result["intent"])
	}
	if result["priority"] != "urgent" {
		t.Errorf("expected priority 'urgent', got %v", result["priority"])
	}
	if result["description"] != "Approve the medication request" {
		t.Errorf("expected description, got %v", result["description"])
	}

	// Check statusReason
	if result["statusReason"] == nil {
		t.Error("expected statusReason to be populated")
	}

	// Check code
	if result["code"] == nil {
		t.Error("expected code to be populated")
	}

	// Check focus
	if result["focus"] == nil {
		t.Error("expected focus to be populated")
	}

	// Check encounter
	if result["encounter"] == nil {
		t.Error("expected encounter to be populated")
	}

	// Check authoredOn
	if result["authoredOn"] == nil {
		t.Error("expected authoredOn to be populated")
	}

	// Check lastModified
	if result["lastModified"] == nil {
		t.Error("expected lastModified to be populated")
	}

	// Check requester
	if result["requester"] == nil {
		t.Error("expected requester to be populated")
	}

	// Check owner
	if result["owner"] == nil {
		t.Error("expected owner to be populated")
	}

	// Check reasonCode
	if result["reasonCode"] == nil {
		t.Error("expected reasonCode to be populated")
	}

	// Check note
	if result["note"] == nil {
		t.Error("expected note to be populated")
	}

	// Check restriction
	if result["restriction"] == nil {
		t.Error("expected restriction to be populated")
	}

	// Check input
	if result["input"] == nil {
		t.Error("expected input to be populated")
	}

	// Check output
	if result["output"] == nil {
		t.Error("expected output to be populated")
	}

	// Verify the FHIR resource is valid JSON
	raw, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("failed to marshal FHIR resource: %v", err)
	}
	var parsed map[string]interface{}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		t.Fatalf("FHIR resource is not valid JSON: %v", err)
	}
}

func TestTask_ToFHIR_MinimalFields(t *testing.T) {
	patientID := uuid.New()
	now := time.Now().UTC()

	tk := &Task{
		ID:           uuid.New(),
		FHIRID:       "task-minimal",
		Status:       "draft",
		Intent:       "proposal",
		ForPatientID: patientID,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	result := tk.ToFHIR()

	// Required fields must be present
	if result["resourceType"] != "Task" {
		t.Errorf("expected resourceType 'Task', got %v", result["resourceType"])
	}
	if result["id"] != "task-minimal" {
		t.Errorf("expected id 'task-minimal', got %v", result["id"])
	}
	if result["status"] != "draft" {
		t.Errorf("expected status 'draft', got %v", result["status"])
	}
	if result["intent"] != "proposal" {
		t.Errorf("expected intent 'proposal', got %v", result["intent"])
	}

	// Optional fields must be absent
	optionalFields := []string{
		"statusReason", "priority", "code", "description", "focus",
		"encounter", "authoredOn", "lastModified", "requester",
		"owner", "reasonCode", "note", "restriction", "input", "output",
	}
	for _, f := range optionalFields {
		if result[f] != nil {
			t.Errorf("expected %s to be nil for minimal task, got %v", f, result[f])
		}
	}
}
