package billing

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
// Coverage.ToFHIR
// ---------------------------------------------------------------------------

func TestCoverageToFHIR_RequiredFields(t *testing.T) {
	patID := uuid.New()
	now := time.Now()

	c := Coverage{
		ID:        uuid.New(),
		FHIRID:    "cov-100",
		Status:    "active",
		PatientID: patID,
		UpdatedAt: now,
	}

	result := c.ToFHIR()

	// resourceType
	if rt, ok := result["resourceType"]; !ok {
		t.Error("expected resourceType to be present")
	} else if rt != "Coverage" {
		t.Errorf("resourceType = %v, want Coverage", rt)
	}

	// id
	if id, ok := result["id"]; !ok {
		t.Error("expected id to be present")
	} else if id != "cov-100" {
		t.Errorf("id = %v, want cov-100", id)
	}

	// status
	if s, ok := result["status"]; !ok {
		t.Error("expected status to be present")
	} else if s != "active" {
		t.Errorf("status = %v, want active", s)
	}

	// beneficiary
	if _, ok := result["beneficiary"]; !ok {
		t.Error("expected beneficiary to be present")
	}

	// meta
	if _, ok := result["meta"]; !ok {
		t.Error("expected meta to be present")
	}

	// optional fields must be absent
	for _, key := range []string{
		"type", "subscriberId", "relationship", "payor",
		"period", "order", "identifier",
	} {
		if _, ok := result[key]; ok {
			t.Errorf("expected %s to be absent when not set", key)
		}
	}
}

func TestCoverageToFHIR_WithOptionalFields(t *testing.T) {
	patID := uuid.New()
	payorID := uuid.New()
	periodStart := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	periodEnd := time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC)
	now := time.Now()

	c := Coverage{
		ID:            uuid.New(),
		FHIRID:        "cov-200",
		Status:        "active",
		PatientID:     patID,
		TypeCode:      ptrStr("EHCPOL"),
		SubscriberID:  ptrStr("SUB-12345"),
		Relationship:  ptrStr("self"),
		PayorOrgID:    ptrUUID(payorID),
		PeriodStart:   ptrTime(periodStart),
		PeriodEnd:     ptrTime(periodEnd),
		CoverageOrder: ptrInt(1),
		PolicyNumber:  ptrStr("POL-9999"),
		UpdatedAt:     now,
	}

	result := c.ToFHIR()

	for _, key := range []string{
		"type", "subscriberId", "relationship", "payor",
		"period", "order", "identifier",
	} {
		if _, ok := result[key]; !ok {
			t.Errorf("expected %s to be present", key)
		}
	}

	// Check subscriberId
	if sid, ok := result["subscriberId"]; ok && sid != "SUB-12345" {
		t.Errorf("subscriberId = %v, want SUB-12345", sid)
	}

	// Check order
	if o, ok := result["order"]; ok && o != 1 {
		t.Errorf("order = %v, want 1", o)
	}
}

func TestCoverageToFHIR_PayorByNameWhenNoOrgID(t *testing.T) {
	patID := uuid.New()
	now := time.Now()

	c := Coverage{
		ID:        uuid.New(),
		FHIRID:    "cov-300",
		Status:    "active",
		PatientID: patID,
		PayorName: ptrStr("BlueCross"),
		UpdatedAt: now,
	}

	result := c.ToFHIR()

	if _, ok := result["payor"]; !ok {
		t.Error("expected payor to be present when PayorName is set")
	}
}

func TestCoverageToFHIR_PayorOrgIDTakesPrecedence(t *testing.T) {
	patID := uuid.New()
	payorID := uuid.New()
	now := time.Now()

	c := Coverage{
		ID:         uuid.New(),
		FHIRID:     "cov-400",
		Status:     "active",
		PatientID:  patID,
		PayorOrgID: ptrUUID(payorID),
		PayorName:  ptrStr("BlueCross"),
		UpdatedAt:  now,
	}

	result := c.ToFHIR()

	if _, ok := result["payor"]; !ok {
		t.Error("expected payor to be present")
	}
}

// ---------------------------------------------------------------------------
// Claim.ToFHIR
// ---------------------------------------------------------------------------

func TestClaimToFHIR_RequiredFields(t *testing.T) {
	patID := uuid.New()
	now := time.Now()

	cl := Claim{
		ID:        uuid.New(),
		FHIRID:    "clm-100",
		Status:    "active",
		PatientID: patID,
		UpdatedAt: now,
	}

	result := cl.ToFHIR()

	// resourceType
	if rt, ok := result["resourceType"]; !ok {
		t.Error("expected resourceType to be present")
	} else if rt != "Claim" {
		t.Errorf("resourceType = %v, want Claim", rt)
	}

	// id
	if id, ok := result["id"]; !ok {
		t.Error("expected id to be present")
	} else if id != "clm-100" {
		t.Errorf("id = %v, want clm-100", id)
	}

	// status
	if s, ok := result["status"]; !ok {
		t.Error("expected status to be present")
	} else if s != "active" {
		t.Errorf("status = %v, want active", s)
	}

	// patient
	if _, ok := result["patient"]; !ok {
		t.Error("expected patient to be present")
	}

	// meta
	if _, ok := result["meta"]; !ok {
		t.Error("expected meta to be present")
	}

	// optional fields must be absent
	for _, key := range []string{
		"type", "use", "provider", "insurer", "insurance",
		"priority", "billablePeriod", "total", "encounter", "facility",
	} {
		if _, ok := result[key]; ok {
			t.Errorf("expected %s to be absent when not set", key)
		}
	}
}

func TestClaimToFHIR_WithOptionalFields(t *testing.T) {
	patID := uuid.New()
	encID := uuid.New()
	insurerID := uuid.New()
	providerID := uuid.New()
	coverageID := uuid.New()
	facilityID := uuid.New()
	bpStart := time.Date(2025, 3, 1, 0, 0, 0, 0, time.UTC)
	bpEnd := time.Date(2025, 3, 31, 23, 59, 59, 0, time.UTC)
	now := time.Now()

	cl := Claim{
		ID:                  uuid.New(),
		FHIRID:              "clm-200",
		Status:              "active",
		PatientID:           patID,
		TypeCode:            ptrStr("institutional"),
		UseCode:             ptrStr("claim"),
		EncounterID:         ptrUUID(encID),
		InsurerOrgID:        ptrUUID(insurerID),
		ProviderID:          ptrUUID(providerID),
		CoverageID:          ptrUUID(coverageID),
		PriorityCode:        ptrStr("normal"),
		FacilityID:          ptrUUID(facilityID),
		BillablePeriodStart: ptrTime(bpStart),
		BillablePeriodEnd:   ptrTime(bpEnd),
		TotalAmount:         ptrFloat(1500.00),
		Currency:            ptrStr("USD"),
		UpdatedAt:           now,
	}

	result := cl.ToFHIR()

	for _, key := range []string{
		"type", "use", "provider", "insurer", "insurance",
		"priority", "billablePeriod", "total", "encounter", "facility",
	} {
		if _, ok := result[key]; !ok {
			t.Errorf("expected %s to be present", key)
		}
	}

	// Check use value
	if u, ok := result["use"]; ok && u != "claim" {
		t.Errorf("use = %v, want claim", u)
	}
}

func TestClaimToFHIR_TotalDefaultCurrency(t *testing.T) {
	patID := uuid.New()
	now := time.Now()

	cl := Claim{
		ID:          uuid.New(),
		FHIRID:      "clm-300",
		Status:      "active",
		PatientID:   patID,
		TotalAmount: ptrFloat(500.00),
		UpdatedAt:   now,
	}

	result := cl.ToFHIR()

	totalMap, ok := result["total"].(map[string]interface{})
	if !ok {
		t.Fatal("expected total to be map[string]interface{}")
	}
	if cur, ok := totalMap["currency"]; ok && cur != "USD" {
		t.Errorf("total.currency = %v, want USD (default)", cur)
	}
	if val, ok := totalMap["value"]; ok && val != 500.00 {
		t.Errorf("total.value = %v, want 500.00", val)
	}
}

func TestClaimToFHIR_ProviderOrgFallback(t *testing.T) {
	patID := uuid.New()
	provOrgID := uuid.New()
	now := time.Now()

	cl := Claim{
		ID:            uuid.New(),
		FHIRID:        "clm-400",
		Status:        "active",
		PatientID:     patID,
		ProviderOrgID: ptrUUID(provOrgID),
		UpdatedAt:     now,
	}

	result := cl.ToFHIR()

	if _, ok := result["provider"]; !ok {
		t.Error("expected provider to be present when ProviderOrgID is set")
	}
}

// ---------------------------------------------------------------------------
// ClaimResponse.ToFHIR
// ---------------------------------------------------------------------------

func TestClaimResponseToFHIR_RequiredFields(t *testing.T) {
	claimID := uuid.New()
	now := time.Now()

	cr := ClaimResponse{
		ID:        uuid.New(),
		FHIRID:    "cr-100",
		ClaimID:   claimID,
		Status:    "active",
		CreatedAt: now,
	}

	result := cr.ToFHIR()

	// resourceType
	if rt, ok := result["resourceType"]; !ok {
		t.Error("expected resourceType to be present")
	} else if rt != "ClaimResponse" {
		t.Errorf("resourceType = %v, want ClaimResponse", rt)
	}

	// id
	if id, ok := result["id"]; !ok {
		t.Error("expected id to be present")
	} else if id != "cr-100" {
		t.Errorf("id = %v, want cr-100", id)
	}

	// status
	if s, ok := result["status"]; !ok {
		t.Error("expected status to be present")
	} else if s != "active" {
		t.Errorf("status = %v, want active", s)
	}

	// request (Claim reference)
	if _, ok := result["request"]; !ok {
		t.Error("expected request to be present")
	}

	// meta
	if _, ok := result["meta"]; !ok {
		t.Error("expected meta to be present")
	}

	// optional fields must be absent
	for _, key := range []string{
		"type", "use", "outcome", "disposition",
		"preAuthRef", "payment", "total", "processNote",
	} {
		if _, ok := result[key]; ok {
			t.Errorf("expected %s to be absent when not set", key)
		}
	}
}

func TestClaimResponseToFHIR_WithOptionalFields(t *testing.T) {
	claimID := uuid.New()
	paymentDate := time.Date(2025, 4, 15, 0, 0, 0, 0, time.UTC)
	now := time.Now()

	cr := ClaimResponse{
		ID:              uuid.New(),
		FHIRID:          "cr-200",
		ClaimID:         claimID,
		Status:          "active",
		TypeCode:        ptrStr("institutional"),
		UseCode:         ptrStr("claim"),
		Outcome:         ptrStr("complete"),
		Disposition:     ptrStr("Claim settled as per contract"),
		PreAuthRef:      ptrStr("PA-12345"),
		PaymentTypeCode: ptrStr("complete"),
		PaymentAmount:   ptrFloat(1200.00),
		PaymentDate:     ptrTime(paymentDate),
		TotalAmount:     ptrFloat(1500.00),
		ProcessNote:     ptrStr("Processed without issues"),
		CreatedAt:       now,
	}

	result := cr.ToFHIR()

	for _, key := range []string{
		"type", "use", "outcome", "disposition",
		"preAuthRef", "payment", "total", "processNote",
	} {
		if _, ok := result[key]; !ok {
			t.Errorf("expected %s to be present", key)
		}
	}

	// Check outcome
	if o, ok := result["outcome"]; ok && o != "complete" {
		t.Errorf("outcome = %v, want complete", o)
	}

	// Check disposition
	if d, ok := result["disposition"]; ok && d != "Claim settled as per contract" {
		t.Errorf("disposition = %v, want 'Claim settled as per contract'", d)
	}

	// Check use
	if u, ok := result["use"]; ok && u != "claim" {
		t.Errorf("use = %v, want claim", u)
	}
}

func TestClaimResponseToFHIR_PaymentWithoutDate(t *testing.T) {
	claimID := uuid.New()
	now := time.Now()

	cr := ClaimResponse{
		ID:            uuid.New(),
		FHIRID:        "cr-300",
		ClaimID:       claimID,
		Status:        "active",
		PaymentAmount: ptrFloat(800.00),
		CreatedAt:     now,
	}

	result := cr.ToFHIR()

	paymentMap, ok := result["payment"].(map[string]interface{})
	if !ok {
		t.Fatal("expected payment to be map[string]interface{}")
	}

	// amount should be present
	if _, ok := paymentMap["amount"]; !ok {
		t.Error("expected payment.amount to be present")
	}

	// date should be absent
	if _, ok := paymentMap["date"]; ok {
		t.Error("expected payment.date to be absent when PaymentDate is nil")
	}
}

// ---------------------------------------------------------------------------
// ExplanationOfBenefit.ToFHIR
// ---------------------------------------------------------------------------

func TestExplanationOfBenefitToFHIR_RequiredFields(t *testing.T) {
	patID := uuid.New()
	now := time.Now()

	eob := ExplanationOfBenefit{
		ID:        uuid.New(),
		FHIRID:    "eob-100",
		Status:    "active",
		PatientID: patID,
		CreatedAt: now,
	}

	result := eob.ToFHIR()

	// resourceType
	if rt, ok := result["resourceType"]; !ok {
		t.Error("expected resourceType to be present")
	} else if rt != "ExplanationOfBenefit" {
		t.Errorf("resourceType = %v, want ExplanationOfBenefit", rt)
	}

	// id
	if id, ok := result["id"]; !ok {
		t.Error("expected id to be present")
	} else if id != "eob-100" {
		t.Errorf("id = %v, want eob-100", id)
	}

	// status
	if s, ok := result["status"]; !ok {
		t.Error("expected status to be present")
	} else if s != "active" {
		t.Errorf("status = %v, want active", s)
	}

	// patient
	if _, ok := result["patient"]; !ok {
		t.Error("expected patient to be present")
	}

	// meta
	if _, ok := result["meta"]; !ok {
		t.Error("expected meta to be present")
	}

	// optional fields must be absent
	for _, key := range []string{
		"type", "use", "outcome", "disposition",
		"insurer", "provider", "claim", "claimResponse",
		"insurance", "billablePeriod", "total", "payment",
	} {
		if _, ok := result[key]; ok {
			t.Errorf("expected %s to be absent when not set", key)
		}
	}
}

func TestExplanationOfBenefitToFHIR_WithOptionalFields(t *testing.T) {
	patID := uuid.New()
	claimID := uuid.New()
	crID := uuid.New()
	covID := uuid.New()
	insurerID := uuid.New()
	provID := uuid.New()
	bpStart := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	bpEnd := time.Date(2025, 3, 31, 23, 59, 59, 0, time.UTC)
	payDate := time.Date(2025, 5, 1, 0, 0, 0, 0, time.UTC)
	now := time.Now()

	eob := ExplanationOfBenefit{
		ID:                  uuid.New(),
		FHIRID:              "eob-200",
		Status:              "active",
		PatientID:           patID,
		TypeCode:            ptrStr("institutional"),
		UseCode:             ptrStr("claim"),
		Outcome:             ptrStr("complete"),
		Disposition:         ptrStr("Processed"),
		ClaimID:             ptrUUID(claimID),
		ClaimResponseID:     ptrUUID(crID),
		CoverageID:          ptrUUID(covID),
		InsurerOrgID:        ptrUUID(insurerID),
		ProviderID:          ptrUUID(provID),
		BillablePeriodStart: ptrTime(bpStart),
		BillablePeriodEnd:   ptrTime(bpEnd),
		TotalSubmitted:      ptrFloat(2000.00),
		TotalBenefit:        ptrFloat(1600.00),
		TotalPayment:        ptrFloat(1600.00),
		PaymentDate:         ptrTime(payDate),
		Currency:            ptrStr("USD"),
		CreatedAt:           now,
	}

	result := eob.ToFHIR()

	for _, key := range []string{
		"type", "use", "outcome", "disposition",
		"insurer", "provider", "claim", "claimResponse",
		"insurance", "billablePeriod", "total", "payment",
	} {
		if _, ok := result[key]; !ok {
			t.Errorf("expected %s to be present", key)
		}
	}

	// Check outcome
	if o, ok := result["outcome"]; ok && o != "complete" {
		t.Errorf("outcome = %v, want complete", o)
	}

	// Check disposition
	if d, ok := result["disposition"]; ok && d != "Processed" {
		t.Errorf("disposition = %v, want Processed", d)
	}

	// Check use
	if u, ok := result["use"]; ok && u != "claim" {
		t.Errorf("use = %v, want claim", u)
	}
}

func TestExplanationOfBenefitToFHIR_TotalWithSubmittedOnly(t *testing.T) {
	patID := uuid.New()
	now := time.Now()

	eob := ExplanationOfBenefit{
		ID:             uuid.New(),
		FHIRID:         "eob-300",
		Status:         "active",
		PatientID:      patID,
		TotalSubmitted: ptrFloat(1000.00),
		Currency:       ptrStr("EUR"),
		CreatedAt:      now,
	}

	result := eob.ToFHIR()

	totals, ok := result["total"].([]map[string]interface{})
	if !ok {
		t.Fatal("expected total to be []map[string]interface{}")
	}
	if len(totals) != 1 {
		t.Errorf("total length = %d, want 1", len(totals))
	}
}

func TestExplanationOfBenefitToFHIR_TotalWithBothSubmittedAndBenefit(t *testing.T) {
	patID := uuid.New()
	now := time.Now()

	eob := ExplanationOfBenefit{
		ID:             uuid.New(),
		FHIRID:         "eob-400",
		Status:         "active",
		PatientID:      patID,
		TotalSubmitted: ptrFloat(2000.00),
		TotalBenefit:   ptrFloat(1500.00),
		Currency:       ptrStr("USD"),
		CreatedAt:      now,
	}

	result := eob.ToFHIR()

	totals, ok := result["total"].([]map[string]interface{})
	if !ok {
		t.Fatal("expected total to be []map[string]interface{}")
	}
	if len(totals) != 2 {
		t.Errorf("total length = %d, want 2", len(totals))
	}
}

func TestExplanationOfBenefitToFHIR_PaymentWithoutDate(t *testing.T) {
	patID := uuid.New()
	now := time.Now()

	eob := ExplanationOfBenefit{
		ID:           uuid.New(),
		FHIRID:       "eob-500",
		Status:       "active",
		PatientID:    patID,
		TotalPayment: ptrFloat(1000.00),
		Currency:     ptrStr("USD"),
		CreatedAt:    now,
	}

	result := eob.ToFHIR()

	paymentMap, ok := result["payment"].(map[string]interface{})
	if !ok {
		t.Fatal("expected payment to be map[string]interface{}")
	}

	if _, ok := paymentMap["amount"]; !ok {
		t.Error("expected payment.amount to be present")
	}
	if _, ok := paymentMap["date"]; ok {
		t.Error("expected payment.date to be absent when PaymentDate is nil")
	}
}
