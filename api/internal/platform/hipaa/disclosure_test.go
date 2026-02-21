package hipaa

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// --- Disclosure purpose constants tests ---

func TestDisclosurePurposeConstants(t *testing.T) {
	purposes := ValidDisclosurePurposes()
	expected := []string{
		"public-health",
		"research",
		"law-enforcement",
		"judicial",
		"workers-comp",
		"decedent",
		"organ-donation",
		"health-oversight",
		"other",
	}

	if len(purposes) != len(expected) {
		t.Errorf("expected %d purposes, got %d", len(expected), len(purposes))
	}

	for _, e := range expected {
		if !IsValidDisclosurePurpose(e) {
			t.Errorf("expected %s to be a valid purpose", e)
		}
	}

	if IsValidDisclosurePurpose("invalid-purpose") {
		t.Error("expected 'invalid-purpose' to be invalid")
	}
	if IsValidDisclosurePurpose("") {
		t.Error("expected empty string to be invalid")
	}
}

func TestPurposeConstantValues(t *testing.T) {
	if PurposePublicHealth != "public-health" {
		t.Errorf("PurposePublicHealth = %s", PurposePublicHealth)
	}
	if PurposeResearch != "research" {
		t.Errorf("PurposeResearch = %s", PurposeResearch)
	}
	if PurposeLawEnforcement != "law-enforcement" {
		t.Errorf("PurposeLawEnforcement = %s", PurposeLawEnforcement)
	}
	if PurposeJudicial != "judicial" {
		t.Errorf("PurposeJudicial = %s", PurposeJudicial)
	}
	if PurposeWorkerComp != "workers-comp" {
		t.Errorf("PurposeWorkerComp = %s", PurposeWorkerComp)
	}
	if PurposeDecedent != "decedent" {
		t.Errorf("PurposeDecedent = %s", PurposeDecedent)
	}
	if PurposeOrganDonation != "organ-donation" {
		t.Errorf("PurposeOrganDonation = %s", PurposeOrganDonation)
	}
	if PurposeHealthOversight != "health-oversight" {
		t.Errorf("PurposeHealthOversight = %s", PurposeHealthOversight)
	}
	if PurposeOther != "other" {
		t.Errorf("PurposeOther = %s", PurposeOther)
	}
}

// --- DisclosureStore tests ---

func TestDisclosureStore_Record(t *testing.T) {
	store := NewDisclosureStore()
	patientID := uuid.New()

	d := &Disclosure{
		PatientID:       patientID,
		DisclosedTo:     "State Health Department",
		DisclosedToType: "organization",
		Purpose:         PurposePublicHealth,
		ResourceTypes:   []string{"Patient", "Condition"},
		DisclosedBy:     "dr-smith",
		Method:          "api",
		Description:     "Required public health reporting",
	}

	err := store.Record(d)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if d.ID == uuid.Nil {
		t.Error("expected ID to be assigned")
	}
	if d.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}
	if d.DateDisclosed.IsZero() {
		t.Error("expected DateDisclosed to be set")
	}
}

func TestDisclosureStore_Record_Validation(t *testing.T) {
	store := NewDisclosureStore()

	// Missing patient_id
	err := store.Record(&Disclosure{
		DisclosedTo: "Some Org",
		Purpose:     PurposeResearch,
	})
	if err == nil {
		t.Error("expected error for missing patient_id")
	}

	// Missing disclosed_to
	err = store.Record(&Disclosure{
		PatientID: uuid.New(),
		Purpose:   PurposeResearch,
	})
	if err == nil {
		t.Error("expected error for missing disclosed_to")
	}

	// Missing purpose
	err = store.Record(&Disclosure{
		PatientID:   uuid.New(),
		DisclosedTo: "Some Org",
	})
	if err == nil {
		t.Error("expected error for missing purpose")
	}
}

func TestDisclosureStore_ListByPatient(t *testing.T) {
	store := NewDisclosureStore()
	patientA := uuid.New()
	patientB := uuid.New()

	now := time.Now().UTC()

	// Patient A disclosures
	_ = store.Record(&Disclosure{
		PatientID:     patientA,
		DisclosedTo:   "Org A",
		Purpose:       PurposePublicHealth,
		DateDisclosed: now.Add(-1 * time.Hour),
	})
	_ = store.Record(&Disclosure{
		PatientID:     patientA,
		DisclosedTo:   "Org B",
		Purpose:       PurposeResearch,
		DateDisclosed: now.Add(-2 * time.Hour),
	})

	// Patient B disclosure
	_ = store.Record(&Disclosure{
		PatientID:     patientB,
		DisclosedTo:   "Org C",
		Purpose:       PurposeLawEnforcement,
		DateDisclosed: now.Add(-30 * time.Minute),
	})

	// List patient A
	results, err := store.ListByPatient(patientA, time.Time{}, time.Time{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 disclosures for patient A, got %d", len(results))
	}

	// Verify all belong to patient A
	for _, d := range results {
		if d.PatientID != patientA {
			t.Errorf("expected patient_id %s, got %s", patientA, d.PatientID)
		}
	}

	// Verify sorted by date descending (most recent first)
	if len(results) >= 2 && results[0].DateDisclosed.Before(results[1].DateDisclosed) {
		t.Error("results should be sorted by date descending")
	}
}

func TestDisclosureStore_ListByPatient_FiltersByDateRange(t *testing.T) {
	store := NewDisclosureStore()
	patientID := uuid.New()

	now := time.Now().UTC()

	_ = store.Record(&Disclosure{
		PatientID:     patientID,
		DisclosedTo:   "Org 1",
		Purpose:       PurposePublicHealth,
		DateDisclosed: now.Add(-48 * time.Hour),
	})
	_ = store.Record(&Disclosure{
		PatientID:     patientID,
		DisclosedTo:   "Org 2",
		Purpose:       PurposeResearch,
		DateDisclosed: now.Add(-24 * time.Hour),
	})
	_ = store.Record(&Disclosure{
		PatientID:     patientID,
		DisclosedTo:   "Org 3",
		Purpose:       PurposeJudicial,
		DateDisclosed: now.Add(-1 * time.Hour),
	})

	// Filter: only disclosures in the last 25 hours
	from := now.Add(-25 * time.Hour)
	to := now

	results, err := store.ListByPatient(patientID, from, to)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("expected 2 disclosures within time range, got %d", len(results))
	}
}

func TestDisclosureStore_ListByPatient_ReturnsOnlyThatPatient(t *testing.T) {
	store := NewDisclosureStore()
	patientA := uuid.New()
	patientB := uuid.New()

	for i := 0; i < 5; i++ {
		_ = store.Record(&Disclosure{
			PatientID:   patientA,
			DisclosedTo: "Org",
			Purpose:     PurposeResearch,
		})
		_ = store.Record(&Disclosure{
			PatientID:   patientB,
			DisclosedTo: "Org",
			Purpose:     PurposeResearch,
		})
	}

	resultsA, _ := store.ListByPatient(patientA, time.Time{}, time.Time{})
	if len(resultsA) != 5 {
		t.Errorf("expected 5 disclosures for patient A, got %d", len(resultsA))
	}
	for _, d := range resultsA {
		if d.PatientID != patientA {
			t.Errorf("got disclosure for wrong patient: %s", d.PatientID)
		}
	}

	resultsB, _ := store.ListByPatient(patientB, time.Time{}, time.Time{})
	if len(resultsB) != 5 {
		t.Errorf("expected 5 disclosures for patient B, got %d", len(resultsB))
	}
}

func TestDisclosureStore_ListAll(t *testing.T) {
	store := NewDisclosureStore()

	for i := 0; i < 10; i++ {
		_ = store.Record(&Disclosure{
			PatientID:   uuid.New(),
			DisclosedTo: "Org",
			Purpose:     PurposeOther,
		})
	}

	// Get first page
	page1, total, err := store.ListAll(5, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 10 {
		t.Errorf("expected total 10, got %d", total)
	}
	if len(page1) != 5 {
		t.Errorf("expected 5 items on page 1, got %d", len(page1))
	}

	// Get second page
	page2, total, err := store.ListAll(5, 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 10 {
		t.Errorf("expected total 10, got %d", total)
	}
	if len(page2) != 5 {
		t.Errorf("expected 5 items on page 2, got %d", len(page2))
	}

	// No overlap
	ids := make(map[uuid.UUID]bool)
	for _, d := range page1 {
		ids[d.ID] = true
	}
	for _, d := range page2 {
		if ids[d.ID] {
			t.Errorf("duplicate ID %s across pages", d.ID)
		}
	}
}

func TestDisclosureStore_ListAll_OffsetBeyondTotal(t *testing.T) {
	store := NewDisclosureStore()
	_ = store.Record(&Disclosure{
		PatientID:   uuid.New(),
		DisclosedTo: "Org",
		Purpose:     PurposeOther,
	})

	results, total, err := store.ListAll(10, 100)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 1 {
		t.Errorf("expected total 1, got %d", total)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results when offset > total, got %d", len(results))
	}
}

func TestDisclosureStore_GetByID(t *testing.T) {
	store := NewDisclosureStore()
	d := &Disclosure{
		PatientID:   uuid.New(),
		DisclosedTo: "Org",
		Purpose:     PurposeResearch,
	}
	_ = store.Record(d)

	found := store.GetByID(d.ID)
	if found == nil {
		t.Fatal("expected to find disclosure by ID")
	}
	if found.ID != d.ID {
		t.Errorf("expected ID %s, got %s", d.ID, found.ID)
	}

	notFound := store.GetByID(uuid.New())
	if notFound != nil {
		t.Error("expected nil for non-existent ID")
	}
}

// --- Handler tests ---

func TestDisclosureHandler_RecordDisclosure(t *testing.T) {
	store := NewDisclosureStore()
	h := NewDisclosureHandler(store)

	patientID := uuid.New()
	body := `{
		"patient_id": "` + patientID.String() + `",
		"disclosed_to": "State Health Department",
		"disclosed_to_type": "organization",
		"purpose": "public-health",
		"resource_types": ["Patient", "Condition"],
		"disclosed_by": "dr-smith",
		"method": "api",
		"description": "Required public health reporting"
	}`

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/disclosures", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := h.HandleRecordDisclosure(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d", rec.Code)
	}

	var disclosure Disclosure
	if err := json.Unmarshal(rec.Body.Bytes(), &disclosure); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if disclosure.ID == uuid.Nil {
		t.Error("expected ID to be assigned")
	}
	if disclosure.PatientID != patientID {
		t.Errorf("expected patient_id %s, got %s", patientID, disclosure.PatientID)
	}
	if disclosure.Purpose != PurposePublicHealth {
		t.Errorf("expected purpose public-health, got %s", disclosure.Purpose)
	}
}

func TestDisclosureHandler_RecordDisclosure_InvalidPurpose(t *testing.T) {
	store := NewDisclosureStore()
	h := NewDisclosureHandler(store)

	body := `{
		"patient_id": "` + uuid.New().String() + `",
		"disclosed_to": "Some Org",
		"purpose": "invalid-purpose"
	}`

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/disclosures", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := h.HandleRecordDisclosure(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestDisclosureHandler_RecordDisclosure_MissingFields(t *testing.T) {
	store := NewDisclosureStore()
	h := NewDisclosureHandler(store)

	tests := []struct {
		name string
		body string
	}{
		{"missing patient_id", `{"disclosed_to": "Org", "purpose": "research"}`},
		{"missing disclosed_to", `{"patient_id": "` + uuid.New().String() + `", "purpose": "research"}`},
		{"missing purpose", `{"patient_id": "` + uuid.New().String() + `", "disclosed_to": "Org"}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/api/v1/disclosures", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			if err := h.HandleRecordDisclosure(c); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if rec.Code != http.StatusBadRequest {
				t.Errorf("expected status 400, got %d", rec.Code)
			}
		})
	}
}

func TestDisclosureHandler_ListDisclosures(t *testing.T) {
	store := NewDisclosureStore()
	h := NewDisclosureHandler(store)

	// Add some disclosures
	for i := 0; i < 5; i++ {
		_ = store.Record(&Disclosure{
			PatientID:   uuid.New(),
			DisclosedTo: "Org",
			Purpose:     PurposeResearch,
		})
	}

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/disclosures?limit=3&offset=0", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := h.HandleListDisclosures(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	total := int(resp["total"].(float64))
	if total != 5 {
		t.Errorf("expected total 5, got %d", total)
	}

	data := resp["data"].([]interface{})
	if len(data) != 3 {
		t.Errorf("expected 3 items, got %d", len(data))
	}

	hasMore := resp["has_more"].(bool)
	if !hasMore {
		t.Error("expected has_more to be true")
	}
}

func TestDisclosureHandler_ListPatientDisclosures(t *testing.T) {
	store := NewDisclosureStore()
	h := NewDisclosureHandler(store)

	patientID := uuid.New()

	_ = store.Record(&Disclosure{
		PatientID:   patientID,
		DisclosedTo: "Org A",
		Purpose:     PurposePublicHealth,
	})
	_ = store.Record(&Disclosure{
		PatientID:   patientID,
		DisclosedTo: "Org B",
		Purpose:     PurposeResearch,
	})
	_ = store.Record(&Disclosure{
		PatientID:   uuid.New(), // different patient
		DisclosedTo: "Org C",
		Purpose:     PurposeLawEnforcement,
	})

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/patients/"+patientID.String()+"/disclosures", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("patientId")
	c.SetParamValues(patientID.String())

	if err := h.HandleListPatientDisclosures(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	total := int(resp["total"].(float64))
	if total != 2 {
		t.Errorf("expected 2 disclosures for patient, got %d", total)
	}

	returnedPatientID := resp["patient_id"].(string)
	if returnedPatientID != patientID.String() {
		t.Errorf("expected patient_id %s, got %s", patientID, returnedPatientID)
	}
}

func TestDisclosureHandler_ListPatientDisclosures_InvalidID(t *testing.T) {
	store := NewDisclosureStore()
	h := NewDisclosureHandler(store)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/patients/invalid/disclosures", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("patientId")
	c.SetParamValues("invalid")

	if err := h.HandleListPatientDisclosures(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestDisclosureHandler_FHIRAccountingOfDisclosures(t *testing.T) {
	store := NewDisclosureStore()
	h := NewDisclosureHandler(store)

	patientID := uuid.New()

	_ = store.Record(&Disclosure{
		PatientID:     patientID,
		DisclosedTo:   "Research Institute",
		DisclosedToType: "organization",
		Purpose:       PurposeResearch,
		ResourceTypes: []string{"Patient", "Observation"},
		DisclosedBy:   "dr-jones",
		Method:        "export",
		Description:   "De-identified data for clinical trial",
	})

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient/"+patientID.String()+"/$accounting-of-disclosures", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(patientID.String())

	if err := h.HandleFHIRAccountingOfDisclosures(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	var bundle map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &bundle); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if bundle["resourceType"] != "Bundle" {
		t.Errorf("expected resourceType Bundle, got %v", bundle["resourceType"])
	}
	if bundle["type"] != "searchset" {
		t.Errorf("expected type searchset, got %v", bundle["type"])
	}

	total := int(bundle["total"].(float64))
	if total != 1 {
		t.Errorf("expected 1 entry, got %d", total)
	}

	entries := bundle["entry"].([]interface{})
	if len(entries) != 1 {
		t.Errorf("expected 1 entry, got %d", len(entries))
	}

	entry := entries[0].(map[string]interface{})
	resource := entry["resource"].(map[string]interface{})
	if resource["resourceType"] != "AuditEvent" {
		t.Errorf("expected AuditEvent resource, got %v", resource["resourceType"])
	}
}

func TestDisclosureHandler_FHIRAccountingOfDisclosures_InvalidID(t *testing.T) {
	store := NewDisclosureStore()
	h := NewDisclosureHandler(store)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient/invalid/$accounting-of-disclosures", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("invalid")

	if err := h.HandleFHIRAccountingOfDisclosures(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if resp["resourceType"] != "OperationOutcome" {
		t.Errorf("expected OperationOutcome, got %v", resp["resourceType"])
	}
}

func TestDisclosureHandler_FHIRAccountingOfDisclosures_EmptyResult(t *testing.T) {
	store := NewDisclosureStore()
	h := NewDisclosureHandler(store)

	patientID := uuid.New()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient/"+patientID.String()+"/$accounting-of-disclosures", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(patientID.String())

	if err := h.HandleFHIRAccountingOfDisclosures(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	var bundle map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &bundle); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	total := int(bundle["total"].(float64))
	if total != 0 {
		t.Errorf("expected 0 entries, got %d", total)
	}
}
