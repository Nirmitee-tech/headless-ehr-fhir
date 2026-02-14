package careplan

import (
	"testing"
	"time"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

func ptrStr(s string) *string        { return &s }
func ptrTime(t time.Time) *time.Time { return &t }
func ptrUUID(u uuid.UUID) *uuid.UUID { return &u }

func TestCarePlan_ToFHIR_RequiredFields(t *testing.T) {
	now := time.Now()
	patID := uuid.New()
	cp := &CarePlan{
		ID: uuid.New(), FHIRID: "cp-001", Status: "active", Intent: "plan",
		PatientID: patID, CreatedAt: now, UpdatedAt: now,
	}
	result := cp.ToFHIR()
	if result["resourceType"] != "CarePlan" {
		t.Errorf("resourceType = %v, want CarePlan", result["resourceType"])
	}
	if result["status"] != "active" {
		t.Errorf("status = %v, want active", result["status"])
	}
	if result["intent"] != "plan" {
		t.Errorf("intent = %v, want plan", result["intent"])
	}
	subj, ok := result["subject"].(fhir.Reference)
	if !ok {
		t.Fatal("subject is not fhir.Reference")
	}
	if subj.Reference != "Patient/"+patID.String() {
		t.Errorf("subject.Reference = %v", subj.Reference)
	}
}

func TestCarePlan_ToFHIR_OptionalFields(t *testing.T) {
	now := time.Now()
	encID := uuid.New()
	authID := uuid.New()
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	cp := &CarePlan{
		ID: uuid.New(), FHIRID: "cp-opt", Status: "active", Intent: "plan",
		PatientID: uuid.New(), CategoryCode: ptrStr("assess-plan"),
		Title: ptrStr("Diabetes Care Plan"), Description: ptrStr("Manage T2DM"),
		EncounterID: ptrUUID(encID), PeriodStart: ptrTime(start),
		AuthorID: ptrUUID(authID), Note: ptrStr("Annual review"),
		CreatedAt: now, UpdatedAt: now,
	}
	result := cp.ToFHIR()
	if result["title"] != "Diabetes Care Plan" {
		t.Errorf("title = %v", result["title"])
	}
	if _, ok := result["category"]; !ok {
		t.Error("expected category")
	}
	if _, ok := result["encounter"]; !ok {
		t.Error("expected encounter")
	}
	if _, ok := result["period"]; !ok {
		t.Error("expected period")
	}
	if _, ok := result["author"]; !ok {
		t.Error("expected author")
	}
	if _, ok := result["note"]; !ok {
		t.Error("expected note")
	}
}

func TestCarePlan_ToFHIR_OptionalFieldsNil(t *testing.T) {
	now := time.Now()
	cp := &CarePlan{
		ID: uuid.New(), FHIRID: "cp-nil", Status: "draft", Intent: "plan",
		PatientID: uuid.New(), CreatedAt: now, UpdatedAt: now,
	}
	result := cp.ToFHIR()
	for _, key := range []string{"category", "title", "description", "encounter", "period", "author", "note"} {
		if _, ok := result[key]; ok {
			t.Errorf("expected key %q to be absent", key)
		}
	}
}

func TestGoal_ToFHIR_RequiredFields(t *testing.T) {
	now := time.Now()
	patID := uuid.New()
	g := &Goal{
		ID: uuid.New(), FHIRID: "goal-001", LifecycleStatus: "active",
		Description: "Reduce A1C to 7%", PatientID: patID,
		CreatedAt: now, UpdatedAt: now,
	}
	result := g.ToFHIR()
	if result["resourceType"] != "Goal" {
		t.Errorf("resourceType = %v, want Goal", result["resourceType"])
	}
	if result["lifecycleStatus"] != "active" {
		t.Errorf("lifecycleStatus = %v, want active", result["lifecycleStatus"])
	}
	desc, ok := result["description"].(fhir.CodeableConcept)
	if !ok {
		t.Fatal("description is not fhir.CodeableConcept")
	}
	if desc.Text != "Reduce A1C to 7%" {
		t.Errorf("description.Text = %v", desc.Text)
	}
}

func TestGoal_ToFHIR_OptionalFields(t *testing.T) {
	now := time.Now()
	due := time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)
	exprID := uuid.New()
	g := &Goal{
		ID: uuid.New(), FHIRID: "goal-opt", LifecycleStatus: "active",
		Description: "Lose weight", PatientID: uuid.New(),
		AchievementStatus: ptrStr("in-progress"), CategoryCode: ptrStr("dietary"),
		TargetMeasure: ptrStr("29463-7"), TargetDetailString: ptrStr("< 200 lbs"),
		TargetDueDate: ptrTime(due), ExpressedByID: ptrUUID(exprID),
		Note: ptrStr("Patient motivated"), CreatedAt: now, UpdatedAt: now,
	}
	result := g.ToFHIR()
	if _, ok := result["achievementStatus"]; !ok {
		t.Error("expected achievementStatus")
	}
	if _, ok := result["category"]; !ok {
		t.Error("expected category")
	}
	if _, ok := result["target"]; !ok {
		t.Error("expected target")
	}
	if _, ok := result["expressedBy"]; !ok {
		t.Error("expected expressedBy")
	}
	if _, ok := result["note"]; !ok {
		t.Error("expected note")
	}
}

func TestGoal_ToFHIR_OptionalFieldsNil(t *testing.T) {
	now := time.Now()
	g := &Goal{
		ID: uuid.New(), FHIRID: "goal-nil", LifecycleStatus: "proposed",
		Description: "General health", PatientID: uuid.New(),
		CreatedAt: now, UpdatedAt: now,
	}
	result := g.ToFHIR()
	for _, key := range []string{"achievementStatus", "category", "target", "expressedBy", "note"} {
		if _, ok := result[key]; ok {
			t.Errorf("expected key %q to be absent", key)
		}
	}
}
