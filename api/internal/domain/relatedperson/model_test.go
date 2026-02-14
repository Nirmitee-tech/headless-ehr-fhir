package relatedperson

import (
	"testing"
	"time"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

func ptrStr(s string) *string        { return &s }
func ptrTime(t time.Time) *time.Time { return &t }

// ---------------------------------------------------------------------------
// RelatedPerson.ToFHIR
// ---------------------------------------------------------------------------

func TestRelatedPerson_ToFHIR_RequiredFields(t *testing.T) {
	now := time.Now()
	patID := uuid.New()

	rp := &RelatedPerson{
		ID:                  uuid.New(),
		FHIRID:              "rp-001",
		Active:              true,
		PatientID:           patID,
		RelationshipCode:    "WIFE",
		RelationshipDisplay: "Wife",
		CreatedAt:           now,
		UpdatedAt:           now,
	}

	result := rp.ToFHIR()

	if result["resourceType"] != "RelatedPerson" {
		t.Errorf("resourceType = %v, want RelatedPerson", result["resourceType"])
	}
	if result["id"] != "rp-001" {
		t.Errorf("id = %v, want rp-001", result["id"])
	}
	if result["active"] != true {
		t.Errorf("active = %v, want true", result["active"])
	}

	pat, ok := result["patient"].(fhir.Reference)
	if !ok {
		t.Fatal("patient is not fhir.Reference")
	}
	expected := "Patient/" + patID.String()
	if pat.Reference != expected {
		t.Errorf("patient.Reference = %v, want %v", pat.Reference, expected)
	}

	rels, ok := result["relationship"].([]fhir.CodeableConcept)
	if !ok || len(rels) == 0 {
		t.Fatal("relationship missing or wrong type")
	}
	if len(rels[0].Coding) == 0 || rels[0].Coding[0].Code != "WIFE" {
		t.Errorf("relationship[0].Coding[0].Code = %v, want WIFE", rels[0].Coding[0].Code)
	}

	meta, ok := result["meta"].(fhir.Meta)
	if !ok {
		t.Fatal("meta is not fhir.Meta")
	}
	if meta.LastUpdated != now {
		t.Errorf("meta.LastUpdated = %v, want %v", meta.LastUpdated, now)
	}
}

func TestRelatedPerson_ToFHIR_OptionalFields(t *testing.T) {
	now := time.Now()
	bd := time.Date(1985, 3, 20, 0, 0, 0, 0, time.UTC)
	ps := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)

	rp := &RelatedPerson{
		ID:                  uuid.New(),
		FHIRID:              "rp-opt",
		Active:              true,
		PatientID:           uuid.New(),
		RelationshipCode:    "WIFE",
		RelationshipDisplay: "Wife",
		FamilyName:          ptrStr("Doe"),
		GivenName:           ptrStr("Jane"),
		Phone:               ptrStr("555-1234"),
		Email:               ptrStr("jane@example.com"),
		Gender:              ptrStr("female"),
		BirthDate:           ptrTime(bd),
		AddressLine:         ptrStr("123 Main St"),
		AddressCity:         ptrStr("Springfield"),
		AddressState:        ptrStr("IL"),
		AddressPostalCode:   ptrStr("62701"),
		PeriodStart:         ptrTime(ps),
		CreatedAt:           now,
		UpdatedAt:           now,
	}

	result := rp.ToFHIR()

	// name
	names, ok := result["name"].([]map[string]interface{})
	if !ok || len(names) == 0 {
		t.Fatal("name missing or wrong type")
	}
	if names[0]["family"] != "Doe" {
		t.Errorf("name[0].family = %v, want Doe", names[0]["family"])
	}
	given, ok := names[0]["given"].([]string)
	if !ok || len(given) == 0 || given[0] != "Jane" {
		t.Errorf("name[0].given = %v, want [Jane]", names[0]["given"])
	}

	// gender
	if result["gender"] != "female" {
		t.Errorf("gender = %v, want female", result["gender"])
	}

	// birthDate
	if result["birthDate"] != "1985-03-20" {
		t.Errorf("birthDate = %v, want 1985-03-20", result["birthDate"])
	}

	// telecom
	telecom, ok := result["telecom"].([]map[string]string)
	if !ok || len(telecom) < 2 {
		t.Fatal("telecom missing or wrong count")
	}
	foundPhone := false
	foundEmail := false
	for _, tc := range telecom {
		if tc["system"] == "phone" && tc["value"] == "555-1234" {
			foundPhone = true
		}
		if tc["system"] == "email" && tc["value"] == "jane@example.com" {
			foundEmail = true
		}
	}
	if !foundPhone {
		t.Error("expected phone telecom entry")
	}
	if !foundEmail {
		t.Error("expected email telecom entry")
	}
}

func TestRelatedPerson_ToFHIR_OptionalFieldsNil(t *testing.T) {
	now := time.Now()

	rp := &RelatedPerson{
		ID:                  uuid.New(),
		FHIRID:              "rp-nil",
		Active:              true,
		PatientID:           uuid.New(),
		RelationshipCode:    "WIFE",
		RelationshipDisplay: "Wife",
		CreatedAt:           now,
		UpdatedAt:           now,
	}

	result := rp.ToFHIR()

	absentKeys := []string{"name", "gender", "birthDate", "telecom"}
	for _, key := range absentKeys {
		if _, ok := result[key]; ok {
			t.Errorf("expected key %q to be absent", key)
		}
	}
}
