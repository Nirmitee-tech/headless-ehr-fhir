package medication

import (
	"encoding/json"
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
// MedicationRequest.ToFHIR
// ---------------------------------------------------------------------------

func TestMedicationRequestToFHIR_RequiredFields(t *testing.T) {
	medID := uuid.New()
	patID := uuid.New()
	reqID := uuid.New()
	now := time.Now()

	mr := MedicationRequest{
		ID:           uuid.New(),
		FHIRID:       "mr-123",
		Status:       "active",
		Intent:       "order",
		MedicationID: medID,
		PatientID:    patID,
		RequesterID:  reqID,
		UpdatedAt:    now,
	}

	result := mr.ToFHIR()

	// resourceType
	if rt, ok := result["resourceType"]; !ok {
		t.Error("expected resourceType to be present")
	} else if rt != "MedicationRequest" {
		t.Errorf("resourceType = %v, want MedicationRequest", rt)
	}

	// id
	if id, ok := result["id"]; !ok {
		t.Error("expected id to be present")
	} else if id != "mr-123" {
		t.Errorf("id = %v, want mr-123", id)
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

	// medicationReference
	if _, ok := result["medicationReference"]; !ok {
		t.Error("expected medicationReference to be present")
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
		"category", "priority", "encounter", "authoredOn",
		"reasonCode", "dosageInstruction", "dispenseRequest",
		"substitution", "note",
	} {
		if _, ok := result[key]; ok {
			t.Errorf("expected %s to be absent when not set", key)
		}
	}
}

func TestMedicationRequestToFHIR_WithOptionalFields(t *testing.T) {
	medID := uuid.New()
	patID := uuid.New()
	reqID := uuid.New()
	encID := uuid.New()
	authored := time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC)
	now := time.Now()

	mr := MedicationRequest{
		ID:                  uuid.New(),
		FHIRID:              "mr-456",
		Status:              "active",
		Intent:              "order",
		MedicationID:        medID,
		PatientID:           patID,
		RequesterID:         reqID,
		CategoryCode:        ptrStr("inpatient"),
		CategoryDisplay:     ptrStr("Inpatient"),
		Priority:            ptrStr("urgent"),
		EncounterID:         ptrUUID(encID),
		AuthoredOn:          ptrTime(authored),
		ReasonCode:          ptrStr("pain"),
		ReasonDisplay:       ptrStr("Pain"),
		DosageText:          ptrStr("Take 1 tablet daily"),
		DosageTimingCode:    ptrStr("QD"),
		DosageTimingDisplay: ptrStr("Once daily"),
		DosageRouteCode:     ptrStr("PO"),
		DosageRouteDisplay:  ptrStr("Oral"),
		DoseQuantity:        ptrFloat(500),
		DoseUnit:            ptrStr("mg"),
		AsNeeded:            ptrBool(false),
		QuantityValue:       ptrFloat(30),
		QuantityUnit:        ptrStr("tablets"),
		DaysSupply:          ptrInt(30),
		RefillsAllowed:      ptrInt(3),
		ValidityStart:       ptrTime(authored),
		SubstitutionAllowed: ptrBool(true),
		Note:                ptrStr("Monitor side effects"),
		UpdatedAt:           now,
	}

	result := mr.ToFHIR()

	// category
	if _, ok := result["category"]; !ok {
		t.Error("expected category to be present")
	}

	// priority
	if p, ok := result["priority"]; !ok {
		t.Error("expected priority to be present")
	} else if p != "urgent" {
		t.Errorf("priority = %v, want urgent", p)
	}

	// encounter
	if _, ok := result["encounter"]; !ok {
		t.Error("expected encounter to be present")
	}

	// authoredOn
	if ao, ok := result["authoredOn"]; !ok {
		t.Error("expected authoredOn to be present")
	} else if ao != authored.Format(time.RFC3339) {
		t.Errorf("authoredOn = %v, want %v", ao, authored.Format(time.RFC3339))
	}

	// reasonCode
	if _, ok := result["reasonCode"]; !ok {
		t.Error("expected reasonCode to be present")
	}

	// dosageInstruction
	if _, ok := result["dosageInstruction"]; !ok {
		t.Error("expected dosageInstruction to be present")
	}

	// dispenseRequest
	if _, ok := result["dispenseRequest"]; !ok {
		t.Error("expected dispenseRequest to be present")
	}

	// substitution
	if _, ok := result["substitution"]; !ok {
		t.Error("expected substitution to be present")
	}

	// note
	if _, ok := result["note"]; !ok {
		t.Error("expected note to be present")
	}
}

// ---------------------------------------------------------------------------
// MedicationAdministration.ToFHIR
// ---------------------------------------------------------------------------

func TestMedicationAdministrationToFHIR_RequiredFields(t *testing.T) {
	medID := uuid.New()
	patID := uuid.New()
	now := time.Now()

	ma := MedicationAdministration{
		ID:           uuid.New(),
		FHIRID:       "ma-100",
		Status:       "completed",
		MedicationID: medID,
		PatientID:    patID,
		UpdatedAt:    now,
	}

	result := ma.ToFHIR()

	// resourceType
	if rt, ok := result["resourceType"]; !ok {
		t.Error("expected resourceType to be present")
	} else if rt != "MedicationAdministration" {
		t.Errorf("resourceType = %v, want MedicationAdministration", rt)
	}

	// id
	if id, ok := result["id"]; !ok {
		t.Error("expected id to be present")
	} else if id != "ma-100" {
		t.Errorf("id = %v, want ma-100", id)
	}

	// status
	if s, ok := result["status"]; !ok {
		t.Error("expected status to be present")
	} else if s != "completed" {
		t.Errorf("status = %v, want completed", s)
	}

	// medicationReference
	if _, ok := result["medicationReference"]; !ok {
		t.Error("expected medicationReference to be present")
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
		"category", "context", "request", "performer",
		"effectiveDateTime", "effectivePeriod",
		"reasonCode", "dosage", "note",
	} {
		if _, ok := result[key]; ok {
			t.Errorf("expected %s to be absent when not set", key)
		}
	}
}

func TestMedicationAdministrationToFHIR_WithOptionalFields(t *testing.T) {
	medID := uuid.New()
	patID := uuid.New()
	encID := uuid.New()
	mrID := uuid.New()
	perfID := uuid.New()
	effectiveTime := time.Date(2025, 3, 10, 14, 30, 0, 0, time.UTC)
	now := time.Now()

	ma := MedicationAdministration{
		ID:                   uuid.New(),
		FHIRID:               "ma-200",
		Status:               "completed",
		MedicationID:         medID,
		PatientID:            patID,
		CategoryCode:         ptrStr("inpatient"),
		CategoryDisplay:      ptrStr("Inpatient"),
		EncounterID:          ptrUUID(encID),
		MedicationRequestID:  ptrUUID(mrID),
		PerformerID:          ptrUUID(perfID),
		PerformerRoleCode:    ptrStr("nurse"),
		PerformerRoleDisplay: ptrStr("Nurse"),
		EffectiveDatetime:    ptrTime(effectiveTime),
		ReasonCode:           ptrStr("pain"),
		ReasonDisplay:        ptrStr("Pain"),
		DosageText:           ptrStr("500mg IV"),
		DosageRouteCode:      ptrStr("IV"),
		DosageRouteDisplay:   ptrStr("Intravenous"),
		DosageSiteCode:       ptrStr("left-arm"),
		DosageSiteDisplay:    ptrStr("Left Arm"),
		DoseQuantity:         ptrFloat(500),
		DoseUnit:             ptrStr("mg"),
		RateQuantity:         ptrFloat(100),
		RateUnit:             ptrStr("mL/hr"),
		Note:                 ptrStr("Administered without issues"),
		UpdatedAt:            now,
	}

	result := ma.ToFHIR()

	// category
	if _, ok := result["category"]; !ok {
		t.Error("expected category to be present")
	}

	// context (encounter)
	if _, ok := result["context"]; !ok {
		t.Error("expected context to be present")
	}

	// request
	if _, ok := result["request"]; !ok {
		t.Error("expected request to be present")
	}

	// performer
	if _, ok := result["performer"]; !ok {
		t.Error("expected performer to be present")
	}

	// effectiveDateTime
	if edt, ok := result["effectiveDateTime"]; !ok {
		t.Error("expected effectiveDateTime to be present")
	} else if edt != effectiveTime.Format(time.RFC3339) {
		t.Errorf("effectiveDateTime = %v, want %v", edt, effectiveTime.Format(time.RFC3339))
	}

	// effectivePeriod should NOT be set when effectiveDateTime is set
	if _, ok := result["effectivePeriod"]; ok {
		t.Error("expected effectivePeriod to be absent when effectiveDateTime is set")
	}

	// reasonCode
	if _, ok := result["reasonCode"]; !ok {
		t.Error("expected reasonCode to be present")
	}

	// dosage
	if _, ok := result["dosage"]; !ok {
		t.Error("expected dosage to be present")
	}

	// note
	if _, ok := result["note"]; !ok {
		t.Error("expected note to be present")
	}
}

func TestMedicationAdministrationToFHIR_EffectivePeriod(t *testing.T) {
	medID := uuid.New()
	patID := uuid.New()
	start := time.Date(2025, 3, 10, 14, 0, 0, 0, time.UTC)
	end := time.Date(2025, 3, 10, 15, 0, 0, 0, time.UTC)
	now := time.Now()

	ma := MedicationAdministration{
		ID:             uuid.New(),
		FHIRID:         "ma-300",
		Status:         "completed",
		MedicationID:   medID,
		PatientID:      patID,
		EffectiveStart: ptrTime(start),
		EffectiveEnd:   ptrTime(end),
		UpdatedAt:      now,
	}

	result := ma.ToFHIR()

	if _, ok := result["effectivePeriod"]; !ok {
		t.Error("expected effectivePeriod to be present")
	}
	if _, ok := result["effectiveDateTime"]; ok {
		t.Error("expected effectiveDateTime to be absent when using period")
	}
}

// ---------------------------------------------------------------------------
// MedicationDispense.ToFHIR
// ---------------------------------------------------------------------------

func TestMedicationDispenseToFHIR_RequiredFields(t *testing.T) {
	medID := uuid.New()
	patID := uuid.New()
	now := time.Now()

	md := MedicationDispense{
		ID:           uuid.New(),
		FHIRID:       "md-100",
		Status:       "completed",
		MedicationID: medID,
		PatientID:    patID,
		UpdatedAt:    now,
	}

	result := md.ToFHIR()

	// resourceType
	if rt, ok := result["resourceType"]; !ok {
		t.Error("expected resourceType to be present")
	} else if rt != "MedicationDispense" {
		t.Errorf("resourceType = %v, want MedicationDispense", rt)
	}

	// id
	if id, ok := result["id"]; !ok {
		t.Error("expected id to be present")
	} else if id != "md-100" {
		t.Errorf("id = %v, want md-100", id)
	}

	// status
	if s, ok := result["status"]; !ok {
		t.Error("expected status to be present")
	} else if s != "completed" {
		t.Errorf("status = %v, want completed", s)
	}

	// medicationReference
	if _, ok := result["medicationReference"]; !ok {
		t.Error("expected medicationReference to be present")
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
		"category", "context", "authorizingPrescription", "performer",
		"location", "quantity", "daysSupply", "whenPrepared",
		"whenHandedOver", "destination", "substitution", "note",
	} {
		if _, ok := result[key]; ok {
			t.Errorf("expected %s to be absent when not set", key)
		}
	}
}

func TestMedicationDispenseToFHIR_WithOptionalFields(t *testing.T) {
	medID := uuid.New()
	patID := uuid.New()
	encID := uuid.New()
	mrID := uuid.New()
	perfID := uuid.New()
	locID := uuid.New()
	destID := uuid.New()
	prepared := time.Date(2025, 4, 1, 9, 0, 0, 0, time.UTC)
	handedOver := time.Date(2025, 4, 1, 10, 0, 0, 0, time.UTC)
	now := time.Now()

	md := MedicationDispense{
		ID:                   uuid.New(),
		FHIRID:               "md-200",
		Status:               "completed",
		MedicationID:         medID,
		PatientID:            patID,
		CategoryCode:         ptrStr("outpatient"),
		CategoryDisplay:      ptrStr("Outpatient"),
		EncounterID:          ptrUUID(encID),
		MedicationRequestID:  ptrUUID(mrID),
		PerformerID:          ptrUUID(perfID),
		LocationID:           ptrUUID(locID),
		QuantityValue:        ptrFloat(30),
		QuantityUnit:         ptrStr("tablets"),
		DaysSupply:           ptrInt(30),
		WhenPrepared:         ptrTime(prepared),
		WhenHandedOver:       ptrTime(handedOver),
		DestinationID:        ptrUUID(destID),
		WasSubstituted:       ptrBool(true),
		SubstitutionTypeCode: ptrStr("G"),
		SubstitutionReason:   ptrStr("Cost savings"),
		Note:                 ptrStr("Dispensed generic"),
		UpdatedAt:            now,
	}

	result := md.ToFHIR()

	for _, key := range []string{
		"category", "context", "authorizingPrescription", "performer",
		"location", "quantity", "daysSupply", "whenPrepared",
		"whenHandedOver", "destination", "substitution", "note",
	} {
		if _, ok := result[key]; !ok {
			t.Errorf("expected %s to be present", key)
		}
	}

	// Verify whenPrepared format
	if wp, ok := result["whenPrepared"]; ok {
		if wp != prepared.Format(time.RFC3339) {
			t.Errorf("whenPrepared = %v, want %v", wp, prepared.Format(time.RFC3339))
		}
	}

	// Verify whenHandedOver format
	if who, ok := result["whenHandedOver"]; ok {
		if who != handedOver.Format(time.RFC3339) {
			t.Errorf("whenHandedOver = %v, want %v", who, handedOver.Format(time.RFC3339))
		}
	}
}

// ---------------------------------------------------------------------------
// Medication struct JSON marshal roundtrip
// ---------------------------------------------------------------------------

func TestMedicationJSONMarshalRoundtrip(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	expDate := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)
	mfgID := uuid.New()

	med := Medication{
		ID:                 uuid.New(),
		FHIRID:             "med-001",
		CodeValue:          "1234",
		CodeDisplay:        "Amoxicillin",
		CodeSystem:         ptrStr("http://www.nlm.nih.gov/research/umls/rxnorm"),
		Status:             "active",
		FormCode:           ptrStr("TAB"),
		FormDisplay:        ptrStr("Tablet"),
		AmountNumerator:    ptrFloat(500),
		AmountNumeratorUnit: ptrStr("mg"),
		IsBrand:            ptrBool(false),
		IsOverTheCounter:   ptrBool(true),
		ManufacturerID:     ptrUUID(mfgID),
		ManufacturerName:   ptrStr("Pharma Corp"),
		LotNumber:          ptrStr("LOT-9876"),
		ExpirationDate:     ptrTime(expDate),
		NDCCode:            ptrStr("12345-6789"),
		IsHighAlert:        ptrBool(false),
		Description:        ptrStr("Amoxicillin 500mg tablet"),
		Note:               ptrStr("Keep refrigerated"),
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	data, err := json.Marshal(med)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var decoded Medication
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	if decoded.FHIRID != med.FHIRID {
		t.Errorf("FHIRID = %v, want %v", decoded.FHIRID, med.FHIRID)
	}
	if decoded.CodeValue != med.CodeValue {
		t.Errorf("CodeValue = %v, want %v", decoded.CodeValue, med.CodeValue)
	}
	if decoded.CodeDisplay != med.CodeDisplay {
		t.Errorf("CodeDisplay = %v, want %v", decoded.CodeDisplay, med.CodeDisplay)
	}
	if decoded.Status != med.Status {
		t.Errorf("Status = %v, want %v", decoded.Status, med.Status)
	}
	if decoded.CodeSystem == nil || *decoded.CodeSystem != *med.CodeSystem {
		t.Error("CodeSystem mismatch after roundtrip")
	}
	if decoded.FormCode == nil || *decoded.FormCode != "TAB" {
		t.Error("FormCode mismatch after roundtrip")
	}
	if decoded.IsBrand == nil || *decoded.IsBrand != false {
		t.Error("IsBrand mismatch after roundtrip")
	}
	if decoded.IsOverTheCounter == nil || *decoded.IsOverTheCounter != true {
		t.Error("IsOverTheCounter mismatch after roundtrip")
	}
	if decoded.NDCCode == nil || *decoded.NDCCode != "12345-6789" {
		t.Error("NDCCode mismatch after roundtrip")
	}
	if decoded.LotNumber == nil || *decoded.LotNumber != "LOT-9876" {
		t.Error("LotNumber mismatch after roundtrip")
	}
	if decoded.Note == nil || *decoded.Note != "Keep refrigerated" {
		t.Error("Note mismatch after roundtrip")
	}
}

func TestMedicationJSONMarshal_OmitsNilOptionalFields(t *testing.T) {
	now := time.Now().Truncate(time.Second)

	med := Medication{
		ID:          uuid.New(),
		FHIRID:      "med-002",
		CodeValue:   "5678",
		CodeDisplay: "Ibuprofen",
		Status:      "active",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	data, err := json.Marshal(med)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("json.Unmarshal into map failed: %v", err)
	}

	// Required fields must exist
	for _, key := range []string{"id", "fhir_id", "code_value", "code_display", "status", "created_at", "updated_at"} {
		if _, ok := raw[key]; !ok {
			t.Errorf("expected required field %s to be present in JSON", key)
		}
	}

	// Optional fields should be omitted
	for _, key := range []string{
		"code_system", "form_code", "form_display",
		"amount_numerator", "amount_numerator_unit",
		"is_brand", "is_over_the_counter", "manufacturer_id",
		"lot_number", "expiration_date", "ndc_code", "description", "note",
	} {
		if _, ok := raw[key]; ok {
			t.Errorf("expected optional field %s to be omitted when nil", key)
		}
	}
}
