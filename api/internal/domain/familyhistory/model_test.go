package familyhistory

import (
	"testing"
	"time"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

func ptrStr(s string) *string        { return &s }
func ptrTime(t time.Time) *time.Time { return &t }
func ptrBool(b bool) *bool           { return &b }
func ptrInt(i int) *int              { return &i }

// ---------------------------------------------------------------------------
// FamilyMemberHistory.ToFHIR
// ---------------------------------------------------------------------------

func TestFamilyMemberHistory_ToFHIR_RequiredFields(t *testing.T) {
	now := time.Now()
	patID := uuid.New()

	f := &FamilyMemberHistory{
		ID:                  uuid.New(),
		FHIRID:              "fmh-001",
		Status:              "completed",
		PatientID:           patID,
		RelationshipCode:    "FTH",
		RelationshipDisplay: "Father",
		CreatedAt:           now,
		UpdatedAt:           now,
	}

	result := f.ToFHIR()

	if result["resourceType"] != "FamilyMemberHistory" {
		t.Errorf("resourceType = %v, want FamilyMemberHistory", result["resourceType"])
	}
	if result["id"] != "fmh-001" {
		t.Errorf("id = %v, want fmh-001", result["id"])
	}
	if result["status"] != "completed" {
		t.Errorf("status = %v, want completed", result["status"])
	}

	pat, ok := result["patient"].(fhir.Reference)
	if !ok {
		t.Fatal("patient is not fhir.Reference")
	}
	expected := "Patient/" + patID.String()
	if pat.Reference != expected {
		t.Errorf("patient.Reference = %v, want %v", pat.Reference, expected)
	}

	rel, ok := result["relationship"].(fhir.CodeableConcept)
	if !ok {
		t.Fatal("relationship is not fhir.CodeableConcept")
	}
	if len(rel.Coding) == 0 || rel.Coding[0].Code != "FTH" {
		t.Errorf("relationship.Coding[0].Code = %v, want FTH", rel.Coding[0].Code)
	}
	if rel.Coding[0].Display != "Father" {
		t.Errorf("relationship.Coding[0].Display = %v, want Father", rel.Coding[0].Display)
	}

	meta, ok := result["meta"].(fhir.Meta)
	if !ok {
		t.Fatal("meta is not fhir.Meta")
	}
	if meta.LastUpdated != now {
		t.Errorf("meta.LastUpdated = %v, want %v", meta.LastUpdated, now)
	}
}

func TestFamilyMemberHistory_ToFHIR_OptionalFields(t *testing.T) {
	now := time.Now()
	d := time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC)
	born := time.Date(1950, 5, 20, 0, 0, 0, 0, time.UTC)

	f := &FamilyMemberHistory{
		ID:                  uuid.New(),
		FHIRID:              "fmh-opt",
		Status:              "completed",
		PatientID:           uuid.New(),
		RelationshipCode:    "MTH",
		RelationshipDisplay: "Mother",
		Date:                ptrTime(d),
		Name:                ptrStr("Jane Doe"),
		Sex:                 ptrStr("female"),
		BornDate:            ptrTime(born),
		DeceasedBoolean:     ptrBool(true),
		DeceasedAge:         ptrInt(72),
		Note:                ptrStr("Heart disease history"),
		CreatedAt:           now,
		UpdatedAt:           now,
	}

	result := f.ToFHIR()

	if result["name"] != "Jane Doe" {
		t.Errorf("name = %v, want Jane Doe", result["name"])
	}
	if _, ok := result["date"]; !ok {
		t.Error("expected date to be present")
	}

	sex, ok := result["sex"].(fhir.CodeableConcept)
	if !ok {
		t.Fatal("sex missing or wrong type")
	}
	if len(sex.Coding) == 0 || sex.Coding[0].Code != "female" {
		t.Errorf("sex.Coding[0].Code = %v, want female", sex.Coding[0].Code)
	}

	if result["deceasedBoolean"] != true {
		t.Errorf("deceasedBoolean = %v, want true", result["deceasedBoolean"])
	}

	da, ok := result["deceasedAge"].(map[string]interface{})
	if !ok {
		t.Fatal("deceasedAge missing or wrong type")
	}
	if da["value"] != 72 {
		t.Errorf("deceasedAge.value = %v, want 72", da["value"])
	}

	notes, ok := result["note"].([]map[string]string)
	if !ok || len(notes) == 0 {
		t.Fatal("note missing or wrong type")
	}
	if notes[0]["text"] != "Heart disease history" {
		t.Errorf("note[0].text = %v, want Heart disease history", notes[0]["text"])
	}
}

func TestFamilyMemberHistory_ToFHIR_OptionalFieldsNil(t *testing.T) {
	now := time.Now()

	f := &FamilyMemberHistory{
		ID:                  uuid.New(),
		FHIRID:              "fmh-nil",
		Status:              "completed",
		PatientID:           uuid.New(),
		RelationshipCode:    "FTH",
		RelationshipDisplay: "Father",
		CreatedAt:           now,
		UpdatedAt:           now,
	}

	result := f.ToFHIR()

	absentKeys := []string{"name", "date", "sex", "deceasedBoolean", "deceasedAge", "note"}
	for _, key := range absentKeys {
		if _, ok := result[key]; ok {
			t.Errorf("expected key %q to be absent", key)
		}
	}
}
