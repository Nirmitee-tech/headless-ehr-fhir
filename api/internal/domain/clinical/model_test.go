package clinical

import (
	"testing"
	"time"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

func ptrStr(s string) *string         { return &s }
func ptrFloat(f float64) *float64     { return &f }
func ptrTime(t time.Time) *time.Time  { return &t }
func ptrUUID(u uuid.UUID) *uuid.UUID  { return &u }

// ---------------------------------------------------------------------------
// Condition.ToFHIR
// ---------------------------------------------------------------------------

func TestCondition_ToFHIR_RequiredFields(t *testing.T) {
	now := time.Now()
	patID := uuid.New()

	c := &Condition{
		ID:             uuid.New(),
		FHIRID:         "cond-001",
		ClinicalStatus: "active",
		CodeValue:      "J06.9",
		CodeDisplay:    "Acute upper respiratory infection",
		PatientID:      patID,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	result := c.ToFHIR()

	if result["resourceType"] != "Condition" {
		t.Errorf("resourceType = %v, want Condition", result["resourceType"])
	}
	if result["id"] != "cond-001" {
		t.Errorf("id = %v, want cond-001", result["id"])
	}

	// clinicalStatus coding
	cs, ok := result["clinicalStatus"].(fhir.CodeableConcept)
	if !ok {
		t.Fatal("clinicalStatus is not fhir.CodeableConcept")
	}
	if len(cs.Coding) == 0 || cs.Coding[0].Code != "active" {
		t.Errorf("clinicalStatus.Coding[0].Code = %v, want active", cs.Coding[0].Code)
	}
	if cs.Coding[0].System != "http://terminology.hl7.org/CodeSystem/condition-clinical" {
		t.Errorf("clinicalStatus.Coding[0].System = %v, want condition-clinical system", cs.Coding[0].System)
	}

	// code coding
	code, ok := result["code"].(fhir.CodeableConcept)
	if !ok {
		t.Fatal("code is not fhir.CodeableConcept")
	}
	if len(code.Coding) == 0 || code.Coding[0].Code != "J06.9" {
		t.Errorf("code.Coding[0].Code = %v, want J06.9", code.Coding[0].Code)
	}
	if code.Coding[0].Display != "Acute upper respiratory infection" {
		t.Errorf("code.Coding[0].Display = %v, want Acute upper respiratory infection", code.Coding[0].Display)
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

	// meta
	meta, ok := result["meta"].(fhir.Meta)
	if !ok {
		t.Fatal("meta is not fhir.Meta")
	}
	if meta.LastUpdated != now {
		t.Errorf("meta.LastUpdated = %v, want %v", meta.LastUpdated, now)
	}
}

func TestCondition_ToFHIR_OptionalFields(t *testing.T) {
	now := time.Now()
	patID := uuid.New()
	encID := uuid.New()
	onset := time.Date(2024, 6, 1, 10, 0, 0, 0, time.UTC)
	recorded := time.Date(2024, 6, 2, 0, 0, 0, 0, time.UTC)

	c := &Condition{
		ID:                 uuid.New(),
		FHIRID:             "cond-opt",
		ClinicalStatus:     "active",
		VerificationStatus: ptrStr("confirmed"),
		CategoryCode:       ptrStr("encounter-diagnosis"),
		SeverityCode:       ptrStr("moderate"),
		CodeValue:          "I10",
		CodeDisplay:        "Hypertension",
		PatientID:          patID,
		EncounterID:        ptrUUID(encID),
		OnsetDatetime:      ptrTime(onset),
		BodySiteCode:       ptrStr("368209003"),
		RecordedDate:       ptrTime(recorded),
		Note:               ptrStr("Patient history note"),
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	result := c.ToFHIR()

	// verificationStatus
	vs, ok := result["verificationStatus"].(fhir.CodeableConcept)
	if !ok {
		t.Fatal("verificationStatus missing or wrong type")
	}
	if len(vs.Coding) == 0 || vs.Coding[0].Code != "confirmed" {
		t.Errorf("verificationStatus.Coding[0].Code = %v, want confirmed", vs.Coding[0].Code)
	}

	// category
	if _, ok := result["category"]; !ok {
		t.Error("expected category to be present")
	}

	// severity
	sev, ok := result["severity"].(fhir.CodeableConcept)
	if !ok {
		t.Fatal("severity missing or wrong type")
	}
	if len(sev.Coding) == 0 || sev.Coding[0].Code != "moderate" {
		t.Errorf("severity.Coding[0].Code = %v, want moderate", sev.Coding[0].Code)
	}

	// encounter
	enc, ok := result["encounter"].(fhir.Reference)
	if !ok {
		t.Fatal("encounter missing or wrong type")
	}
	expectedEnc := "Encounter/" + encID.String()
	if enc.Reference != expectedEnc {
		t.Errorf("encounter.Reference = %v, want %v", enc.Reference, expectedEnc)
	}

	// onsetDateTime
	if _, ok := result["onsetDateTime"]; !ok {
		t.Error("expected onsetDateTime to be present")
	}

	// bodySite
	if _, ok := result["bodySite"]; !ok {
		t.Error("expected bodySite to be present")
	}

	// recordedDate
	if result["recordedDate"] != "2024-06-02" {
		t.Errorf("recordedDate = %v, want 2024-06-02", result["recordedDate"])
	}

	// note
	notes, ok := result["note"].([]map[string]string)
	if !ok || len(notes) == 0 {
		t.Fatal("note missing or wrong type")
	}
	if notes[0]["text"] != "Patient history note" {
		t.Errorf("note[0].text = %v, want Patient history note", notes[0]["text"])
	}
}

func TestCondition_ToFHIR_OptionalFieldsNil(t *testing.T) {
	now := time.Now()
	c := &Condition{
		ID:             uuid.New(),
		FHIRID:         "cond-nil",
		ClinicalStatus: "active",
		CodeValue:      "Z00",
		CodeDisplay:    "General exam",
		PatientID:      uuid.New(),
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	result := c.ToFHIR()

	absentKeys := []string{
		"verificationStatus", "category", "severity", "encounter",
		"onsetDateTime", "bodySite", "recordedDate", "note",
	}
	for _, key := range absentKeys {
		if _, ok := result[key]; ok {
			t.Errorf("expected key %q to be absent", key)
		}
	}
}

// ---------------------------------------------------------------------------
// Observation.ToFHIR
// ---------------------------------------------------------------------------

func TestObservation_ToFHIR_RequiredFields(t *testing.T) {
	now := time.Now()
	patID := uuid.New()

	o := &Observation{
		ID:          uuid.New(),
		FHIRID:      "obs-001",
		Status:      "final",
		CodeValue:   "8480-6",
		CodeDisplay: "Systolic blood pressure",
		PatientID:   patID,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	result := o.ToFHIR()

	if result["resourceType"] != "Observation" {
		t.Errorf("resourceType = %v, want Observation", result["resourceType"])
	}
	if result["id"] != "obs-001" {
		t.Errorf("id = %v, want obs-001", result["id"])
	}
	if result["status"] != "final" {
		t.Errorf("status = %v, want final", result["status"])
	}

	code, ok := result["code"].(fhir.CodeableConcept)
	if !ok {
		t.Fatal("code is not fhir.CodeableConcept")
	}
	if len(code.Coding) == 0 || code.Coding[0].Code != "8480-6" {
		t.Errorf("code.Coding[0].Code = %v, want 8480-6", code.Coding[0].Code)
	}

	subj, ok := result["subject"].(fhir.Reference)
	if !ok {
		t.Fatal("subject is not fhir.Reference")
	}
	expected := "Patient/" + patID.String()
	if subj.Reference != expected {
		t.Errorf("subject.Reference = %v, want %v", subj.Reference, expected)
	}
}

func TestObservation_ToFHIR_ValueQuantity(t *testing.T) {
	now := time.Now()
	o := &Observation{
		ID:            uuid.New(),
		FHIRID:        "obs-vq",
		Status:        "final",
		CodeValue:     "8480-6",
		CodeDisplay:   "Systolic BP",
		PatientID:     uuid.New(),
		ValueQuantity: ptrFloat(120.0),
		ValueUnit:     ptrStr("mmHg"),
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	result := o.ToFHIR()

	vq, ok := result["valueQuantity"].(map[string]interface{})
	if !ok {
		t.Fatal("valueQuantity missing or wrong type")
	}
	if vq["value"] != 120.0 {
		t.Errorf("valueQuantity.value = %v, want 120.0", vq["value"])
	}
	if vq["unit"] != "mmHg" {
		t.Errorf("valueQuantity.unit = %v, want mmHg", vq["unit"])
	}

	// valueString should not be present
	if _, ok := result["valueString"]; ok {
		t.Error("expected valueString to be absent when valueQuantity is set")
	}
}

func TestObservation_ToFHIR_ValueString(t *testing.T) {
	now := time.Now()
	o := &Observation{
		ID:          uuid.New(),
		FHIRID:      "obs-vs",
		Status:      "final",
		CodeValue:   "11506-3",
		CodeDisplay: "Provider notes",
		PatientID:   uuid.New(),
		ValueString: ptrStr("Normal findings"),
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	result := o.ToFHIR()

	if result["valueString"] != "Normal findings" {
		t.Errorf("valueString = %v, want Normal findings", result["valueString"])
	}

	// valueQuantity should not be present
	if _, ok := result["valueQuantity"]; ok {
		t.Error("expected valueQuantity to be absent when valueString is set")
	}
}

func TestObservation_ToFHIR_OptionalFields(t *testing.T) {
	now := time.Now()
	encID := uuid.New()
	effDt := time.Date(2024, 7, 1, 14, 30, 0, 0, time.UTC)

	o := &Observation{
		ID:                 uuid.New(),
		FHIRID:             "obs-opt",
		Status:             "final",
		CategoryCode:       ptrStr("vital-signs"),
		CodeValue:          "8480-6",
		CodeDisplay:        "Systolic BP",
		PatientID:          uuid.New(),
		EncounterID:        ptrUUID(encID),
		EffectiveDatetime:  ptrTime(effDt),
		ValueQuantity:      ptrFloat(130.0),
		ValueUnit:          ptrStr("mmHg"),
		ReferenceRangeLow:  ptrFloat(90.0),
		ReferenceRangeHigh: ptrFloat(140.0),
		ReferenceRangeUnit: ptrStr("mmHg"),
		InterpretationCode: ptrStr("N"),
		Note:               ptrStr("Within normal limits"),
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	result := o.ToFHIR()

	// category
	if _, ok := result["category"]; !ok {
		t.Error("expected category to be present")
	}

	// effectiveDateTime
	if _, ok := result["effectiveDateTime"]; !ok {
		t.Error("expected effectiveDateTime to be present")
	}

	// encounter
	enc, ok := result["encounter"].(fhir.Reference)
	if !ok {
		t.Fatal("encounter missing or wrong type")
	}
	expectedEnc := "Encounter/" + encID.String()
	if enc.Reference != expectedEnc {
		t.Errorf("encounter.Reference = %v, want %v", enc.Reference, expectedEnc)
	}

	// referenceRange
	rr, ok := result["referenceRange"].([]interface{})
	if !ok || len(rr) == 0 {
		t.Fatal("referenceRange missing or wrong type")
	}
	rrMap, ok := rr[0].(map[string]interface{})
	if !ok {
		t.Fatal("referenceRange[0] is not a map")
	}
	if _, ok := rrMap["low"]; !ok {
		t.Error("referenceRange[0] missing low")
	}
	if _, ok := rrMap["high"]; !ok {
		t.Error("referenceRange[0] missing high")
	}

	// interpretation
	if _, ok := result["interpretation"]; !ok {
		t.Error("expected interpretation to be present")
	}

	// note
	notes, ok := result["note"].([]map[string]string)
	if !ok || len(notes) == 0 {
		t.Fatal("note missing or wrong type")
	}
	if notes[0]["text"] != "Within normal limits" {
		t.Errorf("note[0].text = %v, want Within normal limits", notes[0]["text"])
	}
}

func TestObservation_ToFHIR_OptionalFieldsNil(t *testing.T) {
	now := time.Now()
	o := &Observation{
		ID:          uuid.New(),
		FHIRID:      "obs-nil",
		Status:      "final",
		CodeValue:   "8480-6",
		CodeDisplay: "Systolic BP",
		PatientID:   uuid.New(),
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	result := o.ToFHIR()

	absentKeys := []string{
		"category", "effectiveDateTime", "encounter",
		"valueQuantity", "valueString", "referenceRange",
		"interpretation", "note",
	}
	for _, key := range absentKeys {
		if _, ok := result[key]; ok {
			t.Errorf("expected key %q to be absent", key)
		}
	}
}

// ---------------------------------------------------------------------------
// AllergyIntolerance.ToFHIR
// ---------------------------------------------------------------------------

func TestAllergyIntolerance_ToFHIR_RequiredFields(t *testing.T) {
	now := time.Now()
	patID := uuid.New()

	a := &AllergyIntolerance{
		ID:        uuid.New(),
		FHIRID:    "allergy-001",
		PatientID: patID,
		CreatedAt: now,
		UpdatedAt: now,
	}

	result := a.ToFHIR()

	if result["resourceType"] != "AllergyIntolerance" {
		t.Errorf("resourceType = %v, want AllergyIntolerance", result["resourceType"])
	}
	if result["id"] != "allergy-001" {
		t.Errorf("id = %v, want allergy-001", result["id"])
	}

	// patient reference
	pat, ok := result["patient"].(fhir.Reference)
	if !ok {
		t.Fatal("patient is not fhir.Reference")
	}
	expected := "Patient/" + patID.String()
	if pat.Reference != expected {
		t.Errorf("patient.Reference = %v, want %v", pat.Reference, expected)
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

func TestAllergyIntolerance_ToFHIR_OptionalFields(t *testing.T) {
	now := time.Now()
	onset := time.Date(2023, 3, 15, 0, 0, 0, 0, time.UTC)
	recorded := time.Date(2023, 3, 16, 0, 0, 0, 0, time.UTC)

	a := &AllergyIntolerance{
		ID:                 uuid.New(),
		FHIRID:             "allergy-opt",
		PatientID:          uuid.New(),
		ClinicalStatus:     ptrStr("active"),
		VerificationStatus: ptrStr("confirmed"),
		Type:               ptrStr("allergy"),
		Category:           []string{"food"},
		Criticality:        ptrStr("high"),
		CodeValue:          ptrStr("227493005"),
		OnsetDatetime:      ptrTime(onset),
		RecordedDate:       ptrTime(recorded),
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	result := a.ToFHIR()

	// clinicalStatus
	cs, ok := result["clinicalStatus"].(fhir.CodeableConcept)
	if !ok {
		t.Fatal("clinicalStatus missing or wrong type")
	}
	if len(cs.Coding) == 0 || cs.Coding[0].Code != "active" {
		t.Errorf("clinicalStatus.Coding[0].Code = %v, want active", cs.Coding[0].Code)
	}

	// verificationStatus
	vs, ok := result["verificationStatus"].(fhir.CodeableConcept)
	if !ok {
		t.Fatal("verificationStatus missing or wrong type")
	}
	if len(vs.Coding) == 0 || vs.Coding[0].Code != "confirmed" {
		t.Errorf("verificationStatus.Coding[0].Code = %v, want confirmed", vs.Coding[0].Code)
	}

	// type
	if result["type"] != "allergy" {
		t.Errorf("type = %v, want allergy", result["type"])
	}

	// category
	cat, ok := result["category"].([]string)
	if !ok || len(cat) == 0 {
		t.Fatal("category missing or wrong type")
	}
	if cat[0] != "food" {
		t.Errorf("category[0] = %v, want food", cat[0])
	}

	// criticality
	if result["criticality"] != "high" {
		t.Errorf("criticality = %v, want high", result["criticality"])
	}

	// code
	code, ok := result["code"].(fhir.CodeableConcept)
	if !ok {
		t.Fatal("code missing or wrong type")
	}
	if len(code.Coding) == 0 || code.Coding[0].Code != "227493005" {
		t.Errorf("code.Coding[0].Code = %v, want 227493005", code.Coding[0].Code)
	}

	// onsetDateTime
	if _, ok := result["onsetDateTime"]; !ok {
		t.Error("expected onsetDateTime to be present")
	}

	// recordedDate
	if result["recordedDate"] != "2023-03-16" {
		t.Errorf("recordedDate = %v, want 2023-03-16", result["recordedDate"])
	}
}

func TestAllergyIntolerance_ToFHIR_OptionalFieldsNil(t *testing.T) {
	now := time.Now()
	a := &AllergyIntolerance{
		ID:        uuid.New(),
		FHIRID:    "allergy-nil",
		PatientID: uuid.New(),
		CreatedAt: now,
		UpdatedAt: now,
	}

	result := a.ToFHIR()

	absentKeys := []string{
		"clinicalStatus", "verificationStatus", "type",
		"category", "criticality", "code",
		"onsetDateTime", "recordedDate",
	}
	for _, key := range absentKeys {
		if _, ok := result[key]; ok {
			t.Errorf("expected key %q to be absent", key)
		}
	}
}

// ---------------------------------------------------------------------------
// ProcedureRecord.ToFHIR
// ---------------------------------------------------------------------------

func TestProcedureRecord_ToFHIR_RequiredFields(t *testing.T) {
	now := time.Now()
	patID := uuid.New()

	p := &ProcedureRecord{
		ID:          uuid.New(),
		FHIRID:      "proc-001",
		Status:      "completed",
		CodeValue:   "80146002",
		CodeDisplay: "Appendectomy",
		PatientID:   patID,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	result := p.ToFHIR()

	if result["resourceType"] != "Procedure" {
		t.Errorf("resourceType = %v, want Procedure", result["resourceType"])
	}
	if result["id"] != "proc-001" {
		t.Errorf("id = %v, want proc-001", result["id"])
	}
	if result["status"] != "completed" {
		t.Errorf("status = %v, want completed", result["status"])
	}

	code, ok := result["code"].(fhir.CodeableConcept)
	if !ok {
		t.Fatal("code is not fhir.CodeableConcept")
	}
	if len(code.Coding) == 0 || code.Coding[0].Code != "80146002" {
		t.Errorf("code.Coding[0].Code = %v, want 80146002", code.Coding[0].Code)
	}
	if code.Coding[0].Display != "Appendectomy" {
		t.Errorf("code.Coding[0].Display = %v, want Appendectomy", code.Coding[0].Display)
	}

	subj, ok := result["subject"].(fhir.Reference)
	if !ok {
		t.Fatal("subject is not fhir.Reference")
	}
	expected := "Patient/" + patID.String()
	if subj.Reference != expected {
		t.Errorf("subject.Reference = %v, want %v", subj.Reference, expected)
	}
}

func TestProcedureRecord_ToFHIR_OptionalFields(t *testing.T) {
	now := time.Now()
	encID := uuid.New()
	performed := time.Date(2024, 8, 10, 9, 0, 0, 0, time.UTC)
	locID := uuid.New()

	p := &ProcedureRecord{
		ID:                uuid.New(),
		FHIRID:            "proc-opt",
		Status:            "completed",
		CodeValue:         "80146002",
		CodeDisplay:       "Appendectomy",
		PatientID:         uuid.New(),
		CategoryCode:      ptrStr("24642003"),
		EncounterID:       ptrUUID(encID),
		PerformedDatetime: ptrTime(performed),
		BodySiteCode:      ptrStr("66754008"),
		OutcomeCode:       ptrStr("385669000"),
		ReasonCode:        ptrStr("36048009"),
		LocationID:        ptrUUID(locID),
		Note:              ptrStr("Uncomplicated procedure"),
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	result := p.ToFHIR()

	// category
	cat, ok := result["category"].(fhir.CodeableConcept)
	if !ok {
		t.Fatal("category missing or wrong type")
	}
	if len(cat.Coding) == 0 || cat.Coding[0].Code != "24642003" {
		t.Errorf("category.Coding[0].Code = %v, want 24642003", cat.Coding[0].Code)
	}

	// encounter
	enc, ok := result["encounter"].(fhir.Reference)
	if !ok {
		t.Fatal("encounter missing or wrong type")
	}
	expectedEnc := "Encounter/" + encID.String()
	if enc.Reference != expectedEnc {
		t.Errorf("encounter.Reference = %v, want %v", enc.Reference, expectedEnc)
	}

	// performedDateTime
	if _, ok := result["performedDateTime"]; !ok {
		t.Error("expected performedDateTime to be present")
	}

	// bodySite
	if _, ok := result["bodySite"]; !ok {
		t.Error("expected bodySite to be present")
	}

	// outcome
	outcome, ok := result["outcome"].(fhir.CodeableConcept)
	if !ok {
		t.Fatal("outcome missing or wrong type")
	}
	if len(outcome.Coding) == 0 || outcome.Coding[0].Code != "385669000" {
		t.Errorf("outcome.Coding[0].Code = %v, want 385669000", outcome.Coding[0].Code)
	}

	// reasonCode
	if _, ok := result["reasonCode"]; !ok {
		t.Error("expected reasonCode to be present")
	}

	// location
	loc, ok := result["location"].(fhir.Reference)
	if !ok {
		t.Fatal("location missing or wrong type")
	}
	expectedLoc := "Location/" + locID.String()
	if loc.Reference != expectedLoc {
		t.Errorf("location.Reference = %v, want %v", loc.Reference, expectedLoc)
	}

	// note
	notes, ok := result["note"].([]map[string]string)
	if !ok || len(notes) == 0 {
		t.Fatal("note missing or wrong type")
	}
	if notes[0]["text"] != "Uncomplicated procedure" {
		t.Errorf("note[0].text = %v, want Uncomplicated procedure", notes[0]["text"])
	}
}

func TestProcedureRecord_ToFHIR_OptionalFieldsNil(t *testing.T) {
	now := time.Now()
	p := &ProcedureRecord{
		ID:          uuid.New(),
		FHIRID:      "proc-nil",
		Status:      "completed",
		CodeValue:   "80146002",
		CodeDisplay: "Appendectomy",
		PatientID:   uuid.New(),
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	result := p.ToFHIR()

	absentKeys := []string{
		"category", "encounter", "performedDateTime",
		"bodySite", "outcome", "reasonCode", "location", "note",
	}
	for _, key := range absentKeys {
		if _, ok := result[key]; ok {
			t.Errorf("expected key %q to be absent", key)
		}
	}
}
