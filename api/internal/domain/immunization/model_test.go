package immunization

import (
	"testing"
	"time"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

func ptrStr(s string) *string        { return &s }
func ptrFloat(f float64) *float64    { return &f }
func ptrTime(t time.Time) *time.Time { return &t }
func ptrUUID(u uuid.UUID) *uuid.UUID { return &u }
func ptrInt(i int) *int              { return &i }

// ---------------------------------------------------------------------------
// Immunization.ToFHIR
// ---------------------------------------------------------------------------

func TestImmunization_ToFHIR_RequiredFields(t *testing.T) {
	now := time.Now()
	patID := uuid.New()

	im := &Immunization{
		ID:             uuid.New(),
		FHIRID:         "imm-001",
		Status:         "completed",
		VaccineCode:    "08",
		VaccineDisplay: "Hep B",
		PatientID:      patID,
		PrimarySource:  true,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	result := im.ToFHIR()

	if result["resourceType"] != "Immunization" {
		t.Errorf("resourceType = %v, want Immunization", result["resourceType"])
	}
	if result["id"] != "imm-001" {
		t.Errorf("id = %v, want imm-001", result["id"])
	}
	if result["status"] != "completed" {
		t.Errorf("status = %v, want completed", result["status"])
	}

	vc, ok := result["vaccineCode"].(fhir.CodeableConcept)
	if !ok {
		t.Fatal("vaccineCode is not fhir.CodeableConcept")
	}
	if len(vc.Coding) == 0 || vc.Coding[0].Code != "08" {
		t.Errorf("vaccineCode.Coding[0].Code = %v, want 08", vc.Coding[0].Code)
	}
	if vc.Coding[0].System != "http://hl7.org/fhir/sid/cvx" {
		t.Errorf("vaccineCode.Coding[0].System = %v, want CVX system", vc.Coding[0].System)
	}

	subj, ok := result["patient"].(fhir.Reference)
	if !ok {
		t.Fatal("patient is not fhir.Reference")
	}
	expected := "Patient/" + patID.String()
	if subj.Reference != expected {
		t.Errorf("patient.Reference = %v, want %v", subj.Reference, expected)
	}

	if result["primarySource"] != true {
		t.Errorf("primarySource = %v, want true", result["primarySource"])
	}

	meta, ok := result["meta"].(fhir.Meta)
	if !ok {
		t.Fatal("meta is not fhir.Meta")
	}
	if meta.LastUpdated != now {
		t.Errorf("meta.LastUpdated = %v, want %v", meta.LastUpdated, now)
	}
}

func TestImmunization_ToFHIR_OptionalFields(t *testing.T) {
	now := time.Now()
	encID := uuid.New()
	perfID := uuid.New()
	occ := time.Date(2024, 3, 15, 10, 0, 0, 0, time.UTC)
	exp := time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC)

	im := &Immunization{
		ID:                 uuid.New(),
		FHIRID:             "imm-opt",
		Status:             "completed",
		VaccineCode:        "08",
		VaccineDisplay:     "Hep B",
		PatientID:          uuid.New(),
		EncounterID:        ptrUUID(encID),
		OccurrenceDateTime: ptrTime(occ),
		PrimarySource:      true,
		LotNumber:          ptrStr("LOT123"),
		ExpirationDate:     ptrTime(exp),
		SiteCode:           ptrStr("LA"),
		SiteDisplay:        ptrStr("Left arm"),
		RouteCode:          ptrStr("IM"),
		RouteDisplay:       ptrStr("Intramuscular"),
		DoseQuantity:       ptrFloat(0.5),
		DoseUnit:           ptrStr("mL"),
		PerformerID:        ptrUUID(perfID),
		Note:               ptrStr("No adverse reactions"),
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	result := im.ToFHIR()

	// encounter
	enc, ok := result["encounter"].(fhir.Reference)
	if !ok {
		t.Fatal("encounter missing or wrong type")
	}
	expectedEnc := "Encounter/" + encID.String()
	if enc.Reference != expectedEnc {
		t.Errorf("encounter.Reference = %v, want %v", enc.Reference, expectedEnc)
	}

	// occurrenceDateTime
	if _, ok := result["occurrenceDateTime"]; !ok {
		t.Error("expected occurrenceDateTime to be present")
	}

	// lotNumber
	if result["lotNumber"] != "LOT123" {
		t.Errorf("lotNumber = %v, want LOT123", result["lotNumber"])
	}

	// expirationDate
	if result["expirationDate"] != "2025-12-31" {
		t.Errorf("expirationDate = %v, want 2025-12-31", result["expirationDate"])
	}

	// site
	site, ok := result["site"].(fhir.CodeableConcept)
	if !ok {
		t.Fatal("site missing or wrong type")
	}
	if len(site.Coding) == 0 || site.Coding[0].Code != "LA" {
		t.Errorf("site.Coding[0].Code = %v, want LA", site.Coding[0].Code)
	}

	// route
	route, ok := result["route"].(fhir.CodeableConcept)
	if !ok {
		t.Fatal("route missing or wrong type")
	}
	if len(route.Coding) == 0 || route.Coding[0].Code != "IM" {
		t.Errorf("route.Coding[0].Code = %v, want IM", route.Coding[0].Code)
	}

	// doseQuantity
	dq, ok := result["doseQuantity"].(map[string]interface{})
	if !ok {
		t.Fatal("doseQuantity missing or wrong type")
	}
	if dq["value"] != 0.5 {
		t.Errorf("doseQuantity.value = %v, want 0.5", dq["value"])
	}

	// performer
	if _, ok := result["performer"]; !ok {
		t.Error("expected performer to be present")
	}

	// note
	notes, ok := result["note"].([]map[string]string)
	if !ok || len(notes) == 0 {
		t.Fatal("note missing or wrong type")
	}
	if notes[0]["text"] != "No adverse reactions" {
		t.Errorf("note[0].text = %v, want No adverse reactions", notes[0]["text"])
	}
}

func TestImmunization_ToFHIR_OptionalFieldsNil(t *testing.T) {
	now := time.Now()
	im := &Immunization{
		ID:             uuid.New(),
		FHIRID:         "imm-nil",
		Status:         "completed",
		VaccineCode:    "08",
		VaccineDisplay: "Hep B",
		PatientID:      uuid.New(),
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	result := im.ToFHIR()

	absentKeys := []string{
		"encounter", "occurrenceDateTime", "occurrenceString",
		"lotNumber", "expirationDate", "site", "route",
		"doseQuantity", "performer", "note",
	}
	for _, key := range absentKeys {
		if _, ok := result[key]; ok {
			t.Errorf("expected key %q to be absent", key)
		}
	}
}

// ---------------------------------------------------------------------------
// ImmunizationRecommendation.ToFHIR
// ---------------------------------------------------------------------------

func TestImmunizationRecommendation_ToFHIR_RequiredFields(t *testing.T) {
	now := time.Now()
	patID := uuid.New()
	date := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)

	r := &ImmunizationRecommendation{
		ID:             uuid.New(),
		FHIRID:         "rec-001",
		PatientID:      patID,
		Date:           date,
		VaccineCode:    "08",
		VaccineDisplay: "Hep B",
		ForecastStatus: "due",
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	result := r.ToFHIR()

	if result["resourceType"] != "ImmunizationRecommendation" {
		t.Errorf("resourceType = %v, want ImmunizationRecommendation", result["resourceType"])
	}
	if result["id"] != "rec-001" {
		t.Errorf("id = %v, want rec-001", result["id"])
	}

	pat, ok := result["patient"].(fhir.Reference)
	if !ok {
		t.Fatal("patient is not fhir.Reference")
	}
	expected := "Patient/" + patID.String()
	if pat.Reference != expected {
		t.Errorf("patient.Reference = %v, want %v", pat.Reference, expected)
	}

	if _, ok := result["recommendation"]; !ok {
		t.Error("expected recommendation to be present")
	}
}

func TestImmunizationRecommendation_ToFHIR_OptionalFields(t *testing.T) {
	now := time.Now()
	date := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	dateCrit := time.Date(2024, 9, 1, 0, 0, 0, 0, time.UTC)

	r := &ImmunizationRecommendation{
		ID:              uuid.New(),
		FHIRID:          "rec-opt",
		PatientID:       uuid.New(),
		Date:            date,
		VaccineCode:     "08",
		VaccineDisplay:  "Hep B",
		ForecastStatus:  "due",
		ForecastDisplay: ptrStr("Due"),
		DateCriterion:   ptrTime(dateCrit),
		SeriesDoses:     ptrInt(3),
		DoseNumber:      ptrInt(2),
		Description:     ptrStr("Second dose recommended"),
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	result := r.ToFHIR()

	recs, ok := result["recommendation"].([]map[string]interface{})
	if !ok || len(recs) == 0 {
		t.Fatal("recommendation missing or wrong type")
	}
	rec := recs[0]

	if _, ok := rec["dateCriterion"]; !ok {
		t.Error("expected dateCriterion to be present")
	}
	if rec["seriesDosesPositiveInt"] != 3 {
		t.Errorf("seriesDosesPositiveInt = %v, want 3", rec["seriesDosesPositiveInt"])
	}
	if rec["doseNumberPositiveInt"] != 2 {
		t.Errorf("doseNumberPositiveInt = %v, want 2", rec["doseNumberPositiveInt"])
	}
	if rec["description"] != "Second dose recommended" {
		t.Errorf("description = %v, want Second dose recommended", rec["description"])
	}
}

func TestImmunizationRecommendation_ToFHIR_OptionalFieldsNil(t *testing.T) {
	now := time.Now()
	date := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)

	r := &ImmunizationRecommendation{
		ID:             uuid.New(),
		FHIRID:         "rec-nil",
		PatientID:      uuid.New(),
		Date:           date,
		VaccineCode:    "08",
		VaccineDisplay: "Hep B",
		ForecastStatus: "due",
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	result := r.ToFHIR()

	recs, ok := result["recommendation"].([]map[string]interface{})
	if !ok || len(recs) == 0 {
		t.Fatal("recommendation missing")
	}
	rec := recs[0]

	absentKeys := []string{"dateCriterion", "seriesDosesPositiveInt", "doseNumberPositiveInt", "description"}
	for _, key := range absentKeys {
		if _, ok := rec[key]; ok {
			t.Errorf("expected key %q to be absent in recommendation", key)
		}
	}
}
