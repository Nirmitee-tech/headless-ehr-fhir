package nursing

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func ptrStr(s string) *string       { return &s }
func ptrInt(i int) *int             { return &i }
func ptrFloat(f float64) *float64   { return &f }
func ptrBool(b bool) *bool          { return &b }
func ptrTime(t time.Time) *time.Time { return &t }
func ptrUUID(u uuid.UUID) *uuid.UUID { return &u }

func TestFlowsheetTemplate_JSONRoundTrip(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	createdBy := uuid.New()

	original := &FlowsheetTemplate{
		ID:          uuid.New(),
		Name:        "Vital Signs",
		Description: ptrStr("Standard vital signs flowsheet"),
		Category:    ptrStr("vitals"),
		IsActive:    true,
		CreatedBy:   ptrUUID(createdBy),
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded FlowsheetTemplate
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if decoded.ID != original.ID {
		t.Errorf("ID mismatch")
	}
	if decoded.Name != original.Name {
		t.Errorf("Name mismatch: got %q, want %q", decoded.Name, original.Name)
	}
	if *decoded.Description != *original.Description {
		t.Errorf("Description mismatch")
	}
	if *decoded.Category != *original.Category {
		t.Errorf("Category mismatch")
	}
	if decoded.IsActive != original.IsActive {
		t.Errorf("IsActive mismatch: got %v, want %v", decoded.IsActive, original.IsActive)
	}
	if *decoded.CreatedBy != *original.CreatedBy {
		t.Errorf("CreatedBy mismatch")
	}
}

func TestFlowsheetTemplate_OptionalFieldsNil(t *testing.T) {
	m := &FlowsheetTemplate{
		ID:        uuid.New(),
		Name:      "Basic Template",
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	data, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}

	s := string(data)
	if strings.Contains(s, `"description"`) {
		t.Error("nil Description should be omitted")
	}
	if strings.Contains(s, `"category"`) {
		t.Error("nil Category should be omitted")
	}
	if strings.Contains(s, `"created_by"`) {
		t.Error("nil CreatedBy should be omitted")
	}
}

func TestNursingAssessment_JSONRoundTrip(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	completedAt := now.Add(30 * time.Minute)

	original := &NursingAssessment{
		ID:             uuid.New(),
		PatientID:      uuid.New(),
		EncounterID:    uuid.New(),
		NurseID:        uuid.New(),
		AssessmentType: "admission",
		AssessmentData: ptrStr(`{"neuro":"alert","cardiac":"regular"}`),
		Status:         "completed",
		CompletedAt:    ptrTime(completedAt),
		Note:           ptrStr("thorough assessment completed"),
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded NursingAssessment
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if decoded.ID != original.ID {
		t.Errorf("ID mismatch")
	}
	if decoded.PatientID != original.PatientID {
		t.Errorf("PatientID mismatch")
	}
	if decoded.EncounterID != original.EncounterID {
		t.Errorf("EncounterID mismatch")
	}
	if decoded.NurseID != original.NurseID {
		t.Errorf("NurseID mismatch")
	}
	if decoded.AssessmentType != original.AssessmentType {
		t.Errorf("AssessmentType mismatch: got %q, want %q", decoded.AssessmentType, original.AssessmentType)
	}
	if decoded.Status != original.Status {
		t.Errorf("Status mismatch: got %q, want %q", decoded.Status, original.Status)
	}
	if *decoded.AssessmentData != *original.AssessmentData {
		t.Errorf("AssessmentData mismatch")
	}
	if *decoded.Note != *original.Note {
		t.Errorf("Note mismatch")
	}
}

func TestNursingAssessment_OptionalFieldsNil(t *testing.T) {
	m := &NursingAssessment{
		ID:             uuid.New(),
		PatientID:      uuid.New(),
		EncounterID:    uuid.New(),
		NurseID:        uuid.New(),
		AssessmentType: "shift",
		Status:         "in_progress",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	data, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}

	s := string(data)
	if strings.Contains(s, `"assessment_data"`) {
		t.Error("nil AssessmentData should be omitted")
	}
	if strings.Contains(s, `"completed_at"`) {
		t.Error("nil CompletedAt should be omitted")
	}
	if strings.Contains(s, `"note"`) {
		t.Error("nil Note should be omitted")
	}
}

func TestFlowsheetEntry_JSONRoundTrip(t *testing.T) {
	now := time.Now().Truncate(time.Second)

	original := &FlowsheetEntry{
		ID:           uuid.New(),
		TemplateID:   uuid.New(),
		RowID:        uuid.New(),
		PatientID:    uuid.New(),
		EncounterID:  uuid.New(),
		ValueText:    ptrStr("120/80"),
		ValueNumeric: ptrFloat(120.0),
		RecordedAt:   now,
		RecordedByID: uuid.New(),
		Note:         ptrStr("taken sitting"),
		CreatedAt:    now,
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded FlowsheetEntry
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if decoded.ID != original.ID {
		t.Errorf("ID mismatch")
	}
	if decoded.TemplateID != original.TemplateID {
		t.Errorf("TemplateID mismatch")
	}
	if decoded.RowID != original.RowID {
		t.Errorf("RowID mismatch")
	}
	if decoded.PatientID != original.PatientID {
		t.Errorf("PatientID mismatch")
	}
	if decoded.EncounterID != original.EncounterID {
		t.Errorf("EncounterID mismatch")
	}
	if decoded.RecordedByID != original.RecordedByID {
		t.Errorf("RecordedByID mismatch")
	}
	if *decoded.ValueText != *original.ValueText {
		t.Errorf("ValueText mismatch")
	}
	if *decoded.ValueNumeric != *original.ValueNumeric {
		t.Errorf("ValueNumeric mismatch")
	}
	if *decoded.Note != *original.Note {
		t.Errorf("Note mismatch")
	}
}

func TestFlowsheetEntry_OptionalFieldsNil(t *testing.T) {
	m := &FlowsheetEntry{
		ID:           uuid.New(),
		TemplateID:   uuid.New(),
		RowID:        uuid.New(),
		PatientID:    uuid.New(),
		EncounterID:  uuid.New(),
		RecordedAt:   time.Now(),
		RecordedByID: uuid.New(),
		CreatedAt:    time.Now(),
	}

	data, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}

	s := string(data)
	if strings.Contains(s, `"value_text"`) {
		t.Error("nil ValueText should be omitted")
	}
	if strings.Contains(s, `"value_numeric"`) {
		t.Error("nil ValueNumeric should be omitted")
	}
	if strings.Contains(s, `"note"`) {
		t.Error("nil Note should be omitted")
	}
}
