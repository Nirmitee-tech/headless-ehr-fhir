package diagnostics

import (
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

// ---------------------------------------------------------------------------
// ServiceRequest.ToFHIR
// ---------------------------------------------------------------------------

func TestServiceRequestToFHIR_RequiredFields(t *testing.T) {
	patID := uuid.New()
	reqID := uuid.New()
	now := time.Now()

	sr := ServiceRequest{
		ID:          uuid.New(),
		FHIRID:      "sr-100",
		PatientID:   patID,
		RequesterID: reqID,
		Status:      "active",
		Intent:      "order",
		CodeValue:   "CBC",
		CodeDisplay: "Complete Blood Count",
		UpdatedAt:   now,
	}

	result := sr.ToFHIR()

	// resourceType
	if rt, ok := result["resourceType"]; !ok {
		t.Error("expected resourceType to be present")
	} else if rt != "ServiceRequest" {
		t.Errorf("resourceType = %v, want ServiceRequest", rt)
	}

	// id
	if id, ok := result["id"]; !ok {
		t.Error("expected id to be present")
	} else if id != "sr-100" {
		t.Errorf("id = %v, want sr-100", id)
	}

	// status
	if s, ok := result["status"]; !ok {
		t.Error("expected status to be present")
	} else if s != "active" {
		t.Errorf("status = %v, want active", s)
	}

	// intent
	if i, ok := result["intent"]; !ok {
		t.Error("expected intent to be present")
	} else if i != "order" {
		t.Errorf("intent = %v, want order", i)
	}

	// code
	if _, ok := result["code"]; !ok {
		t.Error("expected code to be present")
	}

	// subject
	if _, ok := result["subject"]; !ok {
		t.Error("expected subject to be present")
	}

	// requester
	if _, ok := result["requester"]; !ok {
		t.Error("expected requester to be present")
	}

	// meta
	if _, ok := result["meta"]; !ok {
		t.Error("expected meta to be present")
	}

	// optional fields must be absent
	for _, key := range []string{
		"priority", "category", "encounter", "performer",
		"occurrenceDateTime", "occurrencePeriod", "authoredOn",
		"reasonCode", "bodySite", "note", "patientInstruction",
	} {
		if _, ok := result[key]; ok {
			t.Errorf("expected %s to be absent when not set", key)
		}
	}
}

func TestServiceRequestToFHIR_WithOptionalFields(t *testing.T) {
	patID := uuid.New()
	reqID := uuid.New()
	encID := uuid.New()
	perfID := uuid.New()
	occDT := time.Date(2025, 6, 1, 8, 0, 0, 0, time.UTC)
	authored := time.Date(2025, 5, 30, 14, 0, 0, 0, time.UTC)
	now := time.Now()

	sr := ServiceRequest{
		ID:                 uuid.New(),
		FHIRID:             "sr-200",
		PatientID:          patID,
		RequesterID:        reqID,
		Status:             "active",
		Intent:             "order",
		CodeValue:          "CBC",
		CodeDisplay:        "Complete Blood Count",
		Priority:           ptrStr("urgent"),
		CategoryCode:       ptrStr("laboratory"),
		CategoryDisplay:    ptrStr("Laboratory"),
		EncounterID:        ptrUUID(encID),
		PerformerID:        ptrUUID(perfID),
		OccurrenceDatetime: ptrTime(occDT),
		AuthoredOn:         ptrTime(authored),
		ReasonCode:         ptrStr("anemia"),
		ReasonDisplay:      ptrStr("Anemia"),
		BodySiteCode:       ptrStr("left-arm"),
		BodySiteDisplay:    ptrStr("Left Arm"),
		Note:               ptrStr("Fasting required"),
		PatientInstruction: ptrStr("Do not eat for 12 hours"),
		UpdatedAt:          now,
	}

	result := sr.ToFHIR()

	for _, key := range []string{
		"priority", "category", "encounter", "performer",
		"occurrenceDateTime", "authoredOn",
		"reasonCode", "bodySite", "note", "patientInstruction",
	} {
		if _, ok := result[key]; !ok {
			t.Errorf("expected %s to be present", key)
		}
	}

	// occurrencePeriod should NOT be set when occurrenceDateTime is set
	if _, ok := result["occurrencePeriod"]; ok {
		t.Error("expected occurrencePeriod to be absent when occurrenceDateTime is set")
	}

	// Check specific values
	if p, ok := result["priority"]; ok && p != "urgent" {
		t.Errorf("priority = %v, want urgent", p)
	}
	if pi, ok := result["patientInstruction"]; ok && pi != "Do not eat for 12 hours" {
		t.Errorf("patientInstruction = %v, want 'Do not eat for 12 hours'", pi)
	}
}

func TestServiceRequestToFHIR_OccurrencePeriod(t *testing.T) {
	patID := uuid.New()
	reqID := uuid.New()
	start := time.Date(2025, 6, 1, 8, 0, 0, 0, time.UTC)
	end := time.Date(2025, 6, 1, 10, 0, 0, 0, time.UTC)
	now := time.Now()

	sr := ServiceRequest{
		ID:              uuid.New(),
		FHIRID:          "sr-300",
		PatientID:       patID,
		RequesterID:     reqID,
		Status:          "active",
		Intent:          "order",
		CodeValue:       "CBC",
		CodeDisplay:     "Complete Blood Count",
		OccurrenceStart: ptrTime(start),
		OccurrenceEnd:   ptrTime(end),
		UpdatedAt:       now,
	}

	result := sr.ToFHIR()

	if _, ok := result["occurrencePeriod"]; !ok {
		t.Error("expected occurrencePeriod to be present")
	}
	if _, ok := result["occurrenceDateTime"]; ok {
		t.Error("expected occurrenceDateTime to be absent when using period")
	}
}

// ---------------------------------------------------------------------------
// Specimen.ToFHIR
// ---------------------------------------------------------------------------

func TestSpecimenToFHIR_RequiredFields(t *testing.T) {
	patID := uuid.New()
	now := time.Now()

	sp := Specimen{
		ID:        uuid.New(),
		FHIRID:    "sp-100",
		PatientID: patID,
		Status:    "available",
		UpdatedAt: now,
	}

	result := sp.ToFHIR()

	// resourceType
	if rt, ok := result["resourceType"]; !ok {
		t.Error("expected resourceType to be present")
	} else if rt != "Specimen" {
		t.Errorf("resourceType = %v, want Specimen", rt)
	}

	// id
	if id, ok := result["id"]; !ok {
		t.Error("expected id to be present")
	} else if id != "sp-100" {
		t.Errorf("id = %v, want sp-100", id)
	}

	// status
	if s, ok := result["status"]; !ok {
		t.Error("expected status to be present")
	} else if s != "available" {
		t.Errorf("status = %v, want available", s)
	}

	// subject
	if _, ok := result["subject"]; !ok {
		t.Error("expected subject to be present")
	}

	// meta
	if _, ok := result["meta"]; !ok {
		t.Error("expected meta to be present")
	}

	// optional fields must be absent
	for _, key := range []string{
		"accessionIdentifier", "type", "receivedTime",
		"collection", "condition", "note",
	} {
		if _, ok := result[key]; ok {
			t.Errorf("expected %s to be absent when not set", key)
		}
	}
}

func TestSpecimenToFHIR_WithOptionalFields(t *testing.T) {
	patID := uuid.New()
	collectorID := uuid.New()
	collDT := time.Date(2025, 7, 1, 9, 0, 0, 0, time.UTC)
	recvTime := time.Date(2025, 7, 1, 10, 0, 0, 0, time.UTC)
	now := time.Now()

	sp := Specimen{
		ID:                  uuid.New(),
		FHIRID:              "sp-200",
		PatientID:           patID,
		Status:              "available",
		AccessionID:         ptrStr("ACC-001"),
		TypeCode:            ptrStr("BLD"),
		TypeDisplay:         ptrStr("Blood"),
		ReceivedTime:        ptrTime(recvTime),
		CollectionCollector: ptrUUID(collectorID),
		CollectionDatetime:  ptrTime(collDT),
		CollectionQuantity:  ptrFloat(10),
		CollectionUnit:      ptrStr("mL"),
		CollectionMethod:    ptrStr("venipuncture"),
		CollectionBodySite:  ptrStr("left-arm"),
		ConditionCode:       ptrStr("satisfactory"),
		ConditionDisplay:    ptrStr("Satisfactory"),
		Note:                ptrStr("Collected fasting"),
		UpdatedAt:           now,
	}

	result := sp.ToFHIR()

	for _, key := range []string{
		"accessionIdentifier", "type", "receivedTime",
		"collection", "condition", "note",
	} {
		if _, ok := result[key]; !ok {
			t.Errorf("expected %s to be present", key)
		}
	}

	// Verify receivedTime format
	if rt, ok := result["receivedTime"]; ok {
		if rt != recvTime.Format(time.RFC3339) {
			t.Errorf("receivedTime = %v, want %v", rt, recvTime.Format(time.RFC3339))
		}
	}
}

// ---------------------------------------------------------------------------
// DiagnosticReport.ToFHIR
// ---------------------------------------------------------------------------

func TestDiagnosticReportToFHIR_RequiredFields(t *testing.T) {
	patID := uuid.New()
	now := time.Now()

	dr := DiagnosticReport{
		ID:          uuid.New(),
		FHIRID:      "dr-100",
		PatientID:   patID,
		Status:      "final",
		CodeValue:   "58410-2",
		CodeDisplay: "CBC Panel",
		UpdatedAt:   now,
	}

	result := dr.ToFHIR()

	// resourceType
	if rt, ok := result["resourceType"]; !ok {
		t.Error("expected resourceType to be present")
	} else if rt != "DiagnosticReport" {
		t.Errorf("resourceType = %v, want DiagnosticReport", rt)
	}

	// id
	if id, ok := result["id"]; !ok {
		t.Error("expected id to be present")
	} else if id != "dr-100" {
		t.Errorf("id = %v, want dr-100", id)
	}

	// status
	if s, ok := result["status"]; !ok {
		t.Error("expected status to be present")
	} else if s != "final" {
		t.Errorf("status = %v, want final", s)
	}

	// code
	if _, ok := result["code"]; !ok {
		t.Error("expected code to be present")
	}

	// subject
	if _, ok := result["subject"]; !ok {
		t.Error("expected subject to be present")
	}

	// meta
	if _, ok := result["meta"]; !ok {
		t.Error("expected meta to be present")
	}

	// optional fields must be absent
	for _, key := range []string{
		"category", "encounter", "performer",
		"effectiveDateTime", "effectivePeriod", "issued",
		"specimen", "conclusion", "conclusionCode", "presentedForm",
	} {
		if _, ok := result[key]; ok {
			t.Errorf("expected %s to be absent when not set", key)
		}
	}
}

func TestDiagnosticReportToFHIR_WithOptionalFields(t *testing.T) {
	patID := uuid.New()
	encID := uuid.New()
	perfID := uuid.New()
	specID := uuid.New()
	effectiveDT := time.Date(2025, 8, 1, 14, 0, 0, 0, time.UTC)
	issued := time.Date(2025, 8, 2, 10, 0, 0, 0, time.UTC)
	now := time.Now()

	dr := DiagnosticReport{
		ID:                uuid.New(),
		FHIRID:            "dr-200",
		PatientID:         patID,
		EncounterID:       ptrUUID(encID),
		PerformerID:       ptrUUID(perfID),
		Status:            "final",
		CategoryCode:      ptrStr("LAB"),
		CategoryDisplay:   ptrStr("Laboratory"),
		CodeValue:         "58410-2",
		CodeDisplay:       "CBC Panel",
		EffectiveDatetime: ptrTime(effectiveDT),
		Issued:            ptrTime(issued),
		SpecimenID:        ptrUUID(specID),
		Conclusion:        ptrStr("All values within normal range"),
		ConclusionCode:    ptrStr("normal"),
		ConclusionDisplay: ptrStr("Normal"),
		PresentedFormURL:  ptrStr("https://example.com/report.pdf"),
		PresentedFormType: ptrStr("application/pdf"),
		UpdatedAt:         now,
	}

	result := dr.ToFHIR()

	for _, key := range []string{
		"category", "encounter", "performer",
		"effectiveDateTime", "issued",
		"specimen", "conclusion", "conclusionCode", "presentedForm",
	} {
		if _, ok := result[key]; !ok {
			t.Errorf("expected %s to be present", key)
		}
	}

	// effectivePeriod should NOT be present
	if _, ok := result["effectivePeriod"]; ok {
		t.Error("expected effectivePeriod to be absent when effectiveDateTime is set")
	}

	// Check conclusion value
	if c, ok := result["conclusion"]; ok && c != "All values within normal range" {
		t.Errorf("conclusion = %v, want 'All values within normal range'", c)
	}
}

func TestDiagnosticReportToFHIR_EffectivePeriod(t *testing.T) {
	patID := uuid.New()
	start := time.Date(2025, 8, 1, 8, 0, 0, 0, time.UTC)
	end := time.Date(2025, 8, 1, 12, 0, 0, 0, time.UTC)
	now := time.Now()

	dr := DiagnosticReport{
		ID:             uuid.New(),
		FHIRID:         "dr-300",
		PatientID:      patID,
		Status:         "final",
		CodeValue:      "CBC",
		CodeDisplay:    "CBC",
		EffectiveStart: ptrTime(start),
		EffectiveEnd:   ptrTime(end),
		UpdatedAt:      now,
	}

	result := dr.ToFHIR()

	if _, ok := result["effectivePeriod"]; !ok {
		t.Error("expected effectivePeriod to be present")
	}
	if _, ok := result["effectiveDateTime"]; ok {
		t.Error("expected effectiveDateTime to be absent when using period")
	}
}

// ---------------------------------------------------------------------------
// ImagingStudy.ToFHIR
// ---------------------------------------------------------------------------

func TestImagingStudyToFHIR_RequiredFields(t *testing.T) {
	patID := uuid.New()
	now := time.Now()

	is := ImagingStudy{
		ID:        uuid.New(),
		FHIRID:    "is-100",
		PatientID: patID,
		Status:    "available",
		UpdatedAt: now,
	}

	result := is.ToFHIR()

	// resourceType
	if rt, ok := result["resourceType"]; !ok {
		t.Error("expected resourceType to be present")
	} else if rt != "ImagingStudy" {
		t.Errorf("resourceType = %v, want ImagingStudy", rt)
	}

	// id
	if id, ok := result["id"]; !ok {
		t.Error("expected id to be present")
	} else if id != "is-100" {
		t.Errorf("id = %v, want is-100", id)
	}

	// status
	if s, ok := result["status"]; !ok {
		t.Error("expected status to be present")
	} else if s != "available" {
		t.Errorf("status = %v, want available", s)
	}

	// subject
	if _, ok := result["subject"]; !ok {
		t.Error("expected subject to be present")
	}

	// meta
	if _, ok := result["meta"]; !ok {
		t.Error("expected meta to be present")
	}

	// optional fields must be absent
	for _, key := range []string{
		"modality", "encounter", "referrer", "identifier",
		"numberOfSeries", "numberOfInstances", "description",
		"started", "endpoint", "reasonCode", "note",
	} {
		if _, ok := result[key]; ok {
			t.Errorf("expected %s to be absent when not set", key)
		}
	}
}

func TestImagingStudyToFHIR_WithOptionalFields(t *testing.T) {
	patID := uuid.New()
	encID := uuid.New()
	refID := uuid.New()
	started := time.Date(2025, 9, 15, 11, 0, 0, 0, time.UTC)
	now := time.Now()

	is := ImagingStudy{
		ID:                uuid.New(),
		FHIRID:            "is-200",
		PatientID:         patID,
		EncounterID:       ptrUUID(encID),
		ReferrerID:        ptrUUID(refID),
		Status:            "available",
		ModalityCode:      ptrStr("CT"),
		ModalityDisplay:   ptrStr("Computed Tomography"),
		StudyUID:          ptrStr("1.2.3.4.5"),
		NumberOfSeries:    ptrInt(3),
		NumberOfInstances: ptrInt(120),
		Description:       ptrStr("CT Abdomen with contrast"),
		Started:           ptrTime(started),
		Endpoint:          ptrStr("https://pacs.example.com/wado"),
		ReasonCode:        ptrStr("abdominal-pain"),
		ReasonDisplay:     ptrStr("Abdominal Pain"),
		Note:              ptrStr("Contrast dye used"),
		UpdatedAt:         now,
	}

	result := is.ToFHIR()

	for _, key := range []string{
		"modality", "encounter", "referrer", "identifier",
		"numberOfSeries", "numberOfInstances", "description",
		"started", "endpoint", "reasonCode", "note",
	} {
		if _, ok := result[key]; !ok {
			t.Errorf("expected %s to be present", key)
		}
	}

	// Check specific values
	if ns, ok := result["numberOfSeries"]; ok && ns != 3 {
		t.Errorf("numberOfSeries = %v, want 3", ns)
	}
	if ni, ok := result["numberOfInstances"]; ok && ni != 120 {
		t.Errorf("numberOfInstances = %v, want 120", ni)
	}
	if d, ok := result["description"]; ok && d != "CT Abdomen with contrast" {
		t.Errorf("description = %v, want 'CT Abdomen with contrast'", d)
	}
	if st, ok := result["started"]; ok {
		if st != started.Format(time.RFC3339) {
			t.Errorf("started = %v, want %v", st, started.Format(time.RFC3339))
		}
	}
}
