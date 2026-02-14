package cds

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

func TestCDSRule_JSONRoundTrip(t *testing.T) {
	now := time.Now().Truncate(time.Second)

	original := &CDSRule{
		ID:             uuid.New(),
		RuleName:       "Drug Allergy Check",
		RuleType:       "drug_allergy",
		Description:    ptrStr("Check for known drug allergies before ordering"),
		Severity:       ptrStr("high"),
		Category:       ptrStr("medication_safety"),
		TriggerEvent:   ptrStr("medication_order"),
		ConditionExpr:  ptrStr("patient.allergies CONTAINS order.medication"),
		ActionType:     ptrStr("alert"),
		ActionDetail:   ptrStr("Display allergy warning to prescriber"),
		EvidenceSource: ptrStr("FDA Drug Safety"),
		EvidenceURL:    ptrStr("https://www.fda.gov/drugs/drug-safety"),
		Active:         true,
		Version:        ptrStr("1.0"),
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded CDSRule
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if decoded.ID != original.ID {
		t.Errorf("ID mismatch")
	}
	if decoded.RuleName != original.RuleName {
		t.Errorf("RuleName mismatch: got %q, want %q", decoded.RuleName, original.RuleName)
	}
	if decoded.RuleType != original.RuleType {
		t.Errorf("RuleType mismatch: got %q, want %q", decoded.RuleType, original.RuleType)
	}
	if decoded.Active != original.Active {
		t.Errorf("Active mismatch: got %v, want %v", decoded.Active, original.Active)
	}
	if *decoded.Description != *original.Description {
		t.Errorf("Description mismatch")
	}
	if *decoded.Severity != *original.Severity {
		t.Errorf("Severity mismatch")
	}
	if *decoded.TriggerEvent != *original.TriggerEvent {
		t.Errorf("TriggerEvent mismatch")
	}
	if *decoded.ActionType != *original.ActionType {
		t.Errorf("ActionType mismatch")
	}
	if *decoded.Version != *original.Version {
		t.Errorf("Version mismatch")
	}
}

func TestCDSRule_OptionalFieldsNil(t *testing.T) {
	m := &CDSRule{
		ID:        uuid.New(),
		RuleName:  "Simple Rule",
		RuleType:  "reminder",
		Active:    true,
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
	if strings.Contains(s, `"severity"`) {
		t.Error("nil Severity should be omitted")
	}
	if strings.Contains(s, `"trigger_event"`) {
		t.Error("nil TriggerEvent should be omitted")
	}
	if strings.Contains(s, `"condition_expr"`) {
		t.Error("nil ConditionExpr should be omitted")
	}
	if strings.Contains(s, `"evidence_url"`) {
		t.Error("nil EvidenceURL should be omitted")
	}
	if strings.Contains(s, `"version"`) {
		t.Error("nil Version should be omitted")
	}
}

func TestCDSAlert_JSONRoundTrip(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	encounterID := uuid.New()
	practitionerID := uuid.New()
	expiresAt := now.Add(24 * time.Hour)
	resolvedAt := now.Add(1 * time.Hour)

	original := &CDSAlert{
		ID:              uuid.New(),
		RuleID:          uuid.New(),
		PatientID:       uuid.New(),
		EncounterID:     ptrUUID(encounterID),
		PractitionerID:  ptrUUID(practitionerID),
		Status:          "active",
		Severity:        ptrStr("high"),
		Summary:         "Patient has documented allergy to Penicillin",
		Detail:          ptrStr("Anaphylactic reaction reported 2019"),
		SuggestedAction: ptrStr("Consider alternative antibiotic"),
		Source:          ptrStr("allergy_check_rule_v1"),
		ExpiresAt:       ptrTime(expiresAt),
		FiredAt:         now,
		ResolvedAt:      ptrTime(resolvedAt),
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded CDSAlert
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if decoded.ID != original.ID {
		t.Errorf("ID mismatch")
	}
	if decoded.RuleID != original.RuleID {
		t.Errorf("RuleID mismatch")
	}
	if decoded.PatientID != original.PatientID {
		t.Errorf("PatientID mismatch")
	}
	if decoded.Status != original.Status {
		t.Errorf("Status mismatch: got %q, want %q", decoded.Status, original.Status)
	}
	if decoded.Summary != original.Summary {
		t.Errorf("Summary mismatch: got %q, want %q", decoded.Summary, original.Summary)
	}
	if *decoded.Severity != *original.Severity {
		t.Errorf("Severity mismatch")
	}
	if *decoded.SuggestedAction != *original.SuggestedAction {
		t.Errorf("SuggestedAction mismatch")
	}
	if *decoded.EncounterID != *original.EncounterID {
		t.Errorf("EncounterID mismatch")
	}
}

func TestCDSAlert_OptionalFieldsNil(t *testing.T) {
	m := &CDSAlert{
		ID:        uuid.New(),
		RuleID:    uuid.New(),
		PatientID: uuid.New(),
		Status:    "active",
		Summary:   "Test alert",
		FiredAt:   time.Now(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	data, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}

	s := string(data)
	if strings.Contains(s, `"encounter_id"`) {
		t.Error("nil EncounterID should be omitted")
	}
	if strings.Contains(s, `"severity"`) {
		t.Error("nil Severity should be omitted")
	}
	if strings.Contains(s, `"detail"`) {
		t.Error("nil Detail should be omitted")
	}
	if strings.Contains(s, `"suggested_action"`) {
		t.Error("nil SuggestedAction should be omitted")
	}
	if strings.Contains(s, `"expires_at"`) {
		t.Error("nil ExpiresAt should be omitted")
	}
	if strings.Contains(s, `"resolved_at"`) {
		t.Error("nil ResolvedAt should be omitted")
	}
}
