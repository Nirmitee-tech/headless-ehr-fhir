package encounter

import (
	"testing"
	"time"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

func ptrStr(s string) *string       { return &s }
func ptrUUID(u uuid.UUID) *uuid.UUID { return &u }

// ---------------------------------------------------------------------------
// Encounter.ToFHIR
// ---------------------------------------------------------------------------

func TestEncounter_ToFHIR_RequiredFields(t *testing.T) {
	now := time.Now()
	patID := uuid.New()
	start := time.Date(2024, 7, 1, 8, 0, 0, 0, time.UTC)

	e := &Encounter{
		ID:          uuid.New(),
		FHIRID:      "enc-001",
		Status:      "in-progress",
		ClassCode:   "AMB",
		PatientID:   patID,
		PeriodStart: start,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	result := e.ToFHIR()

	if result["resourceType"] != "Encounter" {
		t.Errorf("resourceType = %v, want Encounter", result["resourceType"])
	}
	if result["id"] != "enc-001" {
		t.Errorf("id = %v, want enc-001", result["id"])
	}
	if result["status"] != "in-progress" {
		t.Errorf("status = %v, want in-progress", result["status"])
	}

	// class coding
	cls, ok := result["class"].(fhir.Coding)
	if !ok {
		t.Fatal("class is not fhir.Coding")
	}
	if cls.Code != "AMB" {
		t.Errorf("class.Code = %v, want AMB", cls.Code)
	}
	if cls.System != "http://terminology.hl7.org/CodeSystem/v3-ActCode" {
		t.Errorf("class.System = %v, want v3-ActCode system", cls.System)
	}

	// subject reference
	subj, ok := result["subject"].(fhir.Reference)
	if !ok {
		t.Fatal("subject is not fhir.Reference")
	}
	expected := "Patient/" + patID.String()
	if subj.Reference != expected {
		t.Errorf("subject.Reference = %v, want %v", subj.Reference, expected)
	}

	// period
	period, ok := result["period"].(fhir.Period)
	if !ok {
		t.Fatal("period is not fhir.Period")
	}
	if period.Start == nil || !period.Start.Equal(start) {
		t.Errorf("period.Start = %v, want %v", period.Start, start)
	}

	// meta
	meta, ok := result["meta"].(fhir.Meta)
	if !ok {
		t.Fatal("meta is not fhir.Meta")
	}
	if meta.LastUpdated != now {
		t.Errorf("meta.LastUpdated = %v, want %v", meta.LastUpdated, now)
	}
}

func TestEncounter_ToFHIR_OptionalFields(t *testing.T) {
	now := time.Now()
	patID := uuid.New()
	practID := uuid.New()
	spID := uuid.New()
	locID := uuid.New()
	start := time.Date(2024, 7, 1, 8, 0, 0, 0, time.UTC)

	e := &Encounter{
		ID:                       uuid.New(),
		FHIRID:                   "enc-opt",
		Status:                   "finished",
		ClassCode:                "IMP",
		ClassDisplay:             ptrStr("inpatient encounter"),
		TypeCode:                 ptrStr("183452005"),
		ServiceTypeCode:          ptrStr("394802001"),
		PriorityCode:             ptrStr("R"),
		PatientID:                patID,
		PrimaryPractitionerID:    ptrUUID(practID),
		ServiceProviderID:        ptrUUID(spID),
		LocationID:               ptrUUID(locID),
		PeriodStart:              start,
		ReasonText:               ptrStr("Chest pain"),
		AdmitSourceCode:          ptrStr("emd"),
		DischargeDispositionCode: ptrStr("home"),
		ReAdmission:              true,
		CreatedAt:                now,
		UpdatedAt:                now,
	}

	result := e.ToFHIR()

	// class display
	cls := result["class"].(fhir.Coding)
	if cls.Display != "inpatient encounter" {
		t.Errorf("class.Display = %v, want inpatient encounter", cls.Display)
	}

	// type
	if _, ok := result["type"]; !ok {
		t.Error("expected type to be present")
	}

	// serviceType
	if _, ok := result["serviceType"]; !ok {
		t.Error("expected serviceType to be present")
	}

	// priority
	pri, ok := result["priority"].(fhir.CodeableConcept)
	if !ok {
		t.Fatal("priority missing or wrong type")
	}
	if len(pri.Coding) == 0 || pri.Coding[0].Code != "R" {
		t.Errorf("priority.Coding[0].Code = %v, want R", pri.Coding[0].Code)
	}

	// participant
	participants, ok := result["participant"].([]map[string]interface{})
	if !ok || len(participants) == 0 {
		t.Fatal("participant missing or wrong type")
	}
	individual, ok := participants[0]["individual"].(fhir.Reference)
	if !ok {
		t.Fatal("participant[0].individual is not fhir.Reference")
	}
	expectedPract := "Practitioner/" + practID.String()
	if individual.Reference != expectedPract {
		t.Errorf("participant[0].individual.Reference = %v, want %v", individual.Reference, expectedPract)
	}

	// serviceProvider
	sp, ok := result["serviceProvider"].(fhir.Reference)
	if !ok {
		t.Fatal("serviceProvider missing or wrong type")
	}
	expectedSP := "Organization/" + spID.String()
	if sp.Reference != expectedSP {
		t.Errorf("serviceProvider.Reference = %v, want %v", sp.Reference, expectedSP)
	}

	// location
	locs, ok := result["location"].([]map[string]interface{})
	if !ok || len(locs) == 0 {
		t.Fatal("location missing or wrong type")
	}
	locRef, ok := locs[0]["location"].(fhir.Reference)
	if !ok {
		t.Fatal("location[0].location is not fhir.Reference")
	}
	expectedLoc := "Location/" + locID.String()
	if locRef.Reference != expectedLoc {
		t.Errorf("location[0].location.Reference = %v, want %v", locRef.Reference, expectedLoc)
	}

	// reasonCode
	reasons, ok := result["reasonCode"].([]fhir.CodeableConcept)
	if !ok || len(reasons) == 0 {
		t.Fatal("reasonCode missing or wrong type")
	}
	if reasons[0].Text != "Chest pain" {
		t.Errorf("reasonCode[0].Text = %v, want Chest pain", reasons[0].Text)
	}

	// hospitalization
	hosp, ok := result["hospitalization"].(map[string]interface{})
	if !ok {
		t.Fatal("hospitalization missing or wrong type")
	}
	if _, ok := hosp["admitSource"]; !ok {
		t.Error("hospitalization missing admitSource")
	}
	if _, ok := hosp["dischargeDisposition"]; !ok {
		t.Error("hospitalization missing dischargeDisposition")
	}
	if _, ok := hosp["reAdmission"]; !ok {
		t.Error("hospitalization missing reAdmission")
	}
}

func TestEncounter_ToFHIR_OptionalFieldsNil(t *testing.T) {
	now := time.Now()
	start := time.Date(2024, 7, 1, 8, 0, 0, 0, time.UTC)

	e := &Encounter{
		ID:          uuid.New(),
		FHIRID:      "enc-nil",
		Status:      "in-progress",
		ClassCode:   "AMB",
		PatientID:   uuid.New(),
		PeriodStart: start,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	result := e.ToFHIR()

	absentKeys := []string{
		"type", "serviceType", "priority",
		"participant", "serviceProvider",
		"location", "reasonCode", "hospitalization",
	}
	for _, key := range absentKeys {
		if _, ok := result[key]; ok {
			t.Errorf("expected key %q to be absent", key)
		}
	}
}
