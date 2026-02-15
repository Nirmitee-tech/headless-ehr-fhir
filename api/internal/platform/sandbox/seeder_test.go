package sandbox

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
)

// ---------------------------------------------------------------------------
// Helper utilities
// ---------------------------------------------------------------------------

func mustString(m map[string]interface{}, key string) string {
	v, ok := m[key]
	if !ok {
		return ""
	}
	s, _ := v.(string)
	return s
}

func mustSlice(m map[string]interface{}, key string) []interface{} {
	v, ok := m[key]
	if !ok {
		return nil
	}
	s, _ := v.([]interface{})
	return s
}

func mustMap(v interface{}) map[string]interface{} {
	m, _ := v.(map[string]interface{})
	return m
}

// ---------------------------------------------------------------------------
// DataGenerator — Patient tests
// ---------------------------------------------------------------------------

func TestDataGenerator_GeneratePatient(t *testing.T) {
	gen := NewDataGenerator(42)
	p := gen.GeneratePatient()

	if p["resourceType"] != "Patient" {
		t.Fatalf("expected resourceType Patient, got %v", p["resourceType"])
	}
	if mustString(p, "id") == "" {
		t.Fatal("expected non-empty id")
	}
	if _, ok := p["active"]; !ok {
		t.Fatal("expected active field")
	}
}

func TestDataGenerator_GeneratePatient_HasName(t *testing.T) {
	gen := NewDataGenerator(42)
	p := gen.GeneratePatient()

	names := mustSlice(p, "name")
	if len(names) == 0 {
		t.Fatal("expected at least one name")
	}
	name := mustMap(names[0])
	if mustString(name, "family") == "" {
		t.Fatal("expected non-empty family name")
	}
	given := mustSlice(name, "given")
	if len(given) == 0 {
		t.Fatal("expected at least one given name")
	}
}

func TestDataGenerator_GeneratePatient_HasBirthDate(t *testing.T) {
	gen := NewDataGenerator(42)
	p := gen.GeneratePatient()

	bd := mustString(p, "birthDate")
	if bd == "" {
		t.Fatal("expected non-empty birthDate")
	}
	// Validate YYYY-MM-DD format
	if len(bd) != 10 || bd[4] != '-' || bd[7] != '-' {
		t.Fatalf("birthDate not in YYYY-MM-DD format: %s", bd)
	}
}

func TestDataGenerator_GeneratePatient_HasGender(t *testing.T) {
	gen := NewDataGenerator(42)
	p := gen.GeneratePatient()

	gender := mustString(p, "gender")
	if gender != "male" && gender != "female" {
		t.Fatalf("expected gender male or female, got %s", gender)
	}
}

func TestDataGenerator_GeneratePatient_HasIdentifier(t *testing.T) {
	gen := NewDataGenerator(42)
	p := gen.GeneratePatient()

	ids := mustSlice(p, "identifier")
	if len(ids) == 0 {
		t.Fatal("expected at least one identifier")
	}
	id0 := mustMap(ids[0])
	typ := mustMap(id0["type"])
	codings := mustSlice(typ, "coding")
	if len(codings) == 0 {
		t.Fatal("expected coding in identifier type")
	}
	coding := mustMap(codings[0])
	if mustString(coding, "code") != "MR" {
		t.Fatal("expected MR code in identifier type")
	}
	if mustString(id0, "value") == "" {
		t.Fatal("expected non-empty identifier value")
	}
}

// ---------------------------------------------------------------------------
// DataGenerator — Practitioner / Organization tests
// ---------------------------------------------------------------------------

func TestDataGenerator_GeneratePractitioner(t *testing.T) {
	gen := NewDataGenerator(42)
	pr := gen.GeneratePractitioner()

	if pr["resourceType"] != "Practitioner" {
		t.Fatalf("expected resourceType Practitioner, got %v", pr["resourceType"])
	}
	if mustString(pr, "id") == "" {
		t.Fatal("expected non-empty id")
	}
	names := mustSlice(pr, "name")
	if len(names) == 0 {
		t.Fatal("expected at least one name on practitioner")
	}
}

func TestDataGenerator_GenerateOrganization(t *testing.T) {
	gen := NewDataGenerator(42)
	org := gen.GenerateOrganization()

	if org["resourceType"] != "Organization" {
		t.Fatalf("expected resourceType Organization, got %v", org["resourceType"])
	}
	if mustString(org, "id") == "" {
		t.Fatal("expected non-empty id")
	}
	if mustString(org, "name") == "" {
		t.Fatal("expected non-empty name")
	}
}

// ---------------------------------------------------------------------------
// DataGenerator — Clinical resource tests
// ---------------------------------------------------------------------------

func TestDataGenerator_GenerateEncounter(t *testing.T) {
	gen := NewDataGenerator(42)
	enc := gen.GenerateEncounter("patient-1", "pract-1")

	if enc["resourceType"] != "Encounter" {
		t.Fatalf("expected resourceType Encounter, got %v", enc["resourceType"])
	}
	subj := mustMap(enc["subject"])
	if mustString(subj, "reference") != "Patient/patient-1" {
		t.Fatalf("expected subject reference Patient/patient-1, got %v", mustString(subj, "reference"))
	}
	participants := mustSlice(enc, "participant")
	if len(participants) == 0 {
		t.Fatal("expected at least one participant")
	}
	part := mustMap(participants[0])
	individual := mustMap(part["individual"])
	if mustString(individual, "reference") != "Practitioner/pract-1" {
		t.Fatalf("expected practitioner reference, got %v", mustString(individual, "reference"))
	}
}

func TestDataGenerator_GenerateObservation(t *testing.T) {
	gen := NewDataGenerator(42)
	obs := gen.GenerateObservation("patient-1", "enc-1")

	if obs["resourceType"] != "Observation" {
		t.Fatalf("expected resourceType Observation, got %v", obs["resourceType"])
	}
	code := mustMap(obs["code"])
	codings := mustSlice(code, "coding")
	if len(codings) == 0 {
		t.Fatal("expected at least one coding in observation code")
	}
	coding := mustMap(codings[0])
	if mustString(coding, "system") != "http://loinc.org" {
		t.Fatalf("expected LOINC system, got %s", mustString(coding, "system"))
	}
	if mustString(coding, "code") == "" {
		t.Fatal("expected non-empty LOINC code")
	}

	// Check value
	if obs["valueQuantity"] == nil {
		t.Fatal("expected valueQuantity")
	}
	vq := mustMap(obs["valueQuantity"])
	if vq["value"] == nil {
		t.Fatal("expected numeric value in valueQuantity")
	}
}

func TestDataGenerator_GenerateCondition(t *testing.T) {
	gen := NewDataGenerator(42)
	cond := gen.GenerateCondition("patient-1")

	if cond["resourceType"] != "Condition" {
		t.Fatalf("expected resourceType Condition, got %v", cond["resourceType"])
	}
	code := mustMap(cond["code"])
	codings := mustSlice(code, "coding")
	if len(codings) == 0 {
		t.Fatal("expected coding in condition code")
	}
	coding := mustMap(codings[0])
	if mustString(coding, "system") != "http://hl7.org/fhir/sid/icd-10-cm" {
		t.Fatalf("expected ICD-10 system, got %s", mustString(coding, "system"))
	}
}

func TestDataGenerator_GenerateMedicationRequest(t *testing.T) {
	gen := NewDataGenerator(42)
	med := gen.GenerateMedicationRequest("patient-1", "pract-1")

	if med["resourceType"] != "MedicationRequest" {
		t.Fatalf("expected resourceType MedicationRequest, got %v", med["resourceType"])
	}
	mc := mustMap(med["medicationCodeableConcept"])
	codings := mustSlice(mc, "coding")
	if len(codings) == 0 {
		t.Fatal("expected coding in medication")
	}
	coding := mustMap(codings[0])
	if mustString(coding, "system") != "http://www.nlm.nih.gov/research/umls/rxnorm" {
		t.Fatalf("expected RxNorm system, got %s", mustString(coding, "system"))
	}
}

func TestDataGenerator_GenerateAllergyIntolerance(t *testing.T) {
	gen := NewDataGenerator(42)
	allergy := gen.GenerateAllergyIntolerance("patient-1")

	if allergy["resourceType"] != "AllergyIntolerance" {
		t.Fatalf("expected resourceType AllergyIntolerance, got %v", allergy["resourceType"])
	}
	code := mustMap(allergy["code"])
	codings := mustSlice(code, "coding")
	if len(codings) == 0 {
		t.Fatal("expected coding in allergy code")
	}
	coding := mustMap(codings[0])
	if mustString(coding, "system") != "http://snomed.info/sct" {
		t.Fatalf("expected SNOMED system, got %s", mustString(coding, "system"))
	}
}

func TestDataGenerator_GenerateProcedure(t *testing.T) {
	gen := NewDataGenerator(42)
	proc := gen.GenerateProcedure("patient-1", "enc-1")

	if proc["resourceType"] != "Procedure" {
		t.Fatalf("expected resourceType Procedure, got %v", proc["resourceType"])
	}
	code := mustMap(proc["code"])
	codings := mustSlice(code, "coding")
	if len(codings) == 0 {
		t.Fatal("expected coding in procedure code")
	}
	coding := mustMap(codings[0])
	if mustString(coding, "system") != "http://www.ama-assn.org/go/cpt" {
		t.Fatalf("expected CPT system, got %s", mustString(coding, "system"))
	}
}

func TestDataGenerator_GenerateImmunization(t *testing.T) {
	gen := NewDataGenerator(42)
	imm := gen.GenerateImmunization("patient-1")

	if imm["resourceType"] != "Immunization" {
		t.Fatalf("expected resourceType Immunization, got %v", imm["resourceType"])
	}
	vc := mustMap(imm["vaccineCode"])
	codings := mustSlice(vc, "coding")
	if len(codings) == 0 {
		t.Fatal("expected coding in vaccineCode")
	}
	coding := mustMap(codings[0])
	if mustString(coding, "system") != "http://hl7.org/fhir/sid/cvx" {
		t.Fatalf("expected CVX system, got %s", mustString(coding, "system"))
	}
}

// ---------------------------------------------------------------------------
// DataGenerator — Reproducibility & Uniqueness
// ---------------------------------------------------------------------------

func TestDataGenerator_Reproducible(t *testing.T) {
	gen1 := NewDataGenerator(99)
	gen2 := NewDataGenerator(99)

	p1 := gen1.GeneratePatient()
	p2 := gen2.GeneratePatient()

	if mustString(p1, "id") != mustString(p2, "id") {
		t.Fatal("same seed should produce same patient id")
	}
	names1 := mustSlice(p1, "name")
	names2 := mustSlice(p2, "name")
	n1 := mustMap(names1[0])
	n2 := mustMap(names2[0])
	if mustString(n1, "family") != mustString(n2, "family") {
		t.Fatal("same seed should produce same family name")
	}
}

func TestDataGenerator_DifferentSeeds(t *testing.T) {
	gen1 := NewDataGenerator(1)
	gen2 := NewDataGenerator(2)

	p1 := gen1.GeneratePatient()
	p2 := gen2.GeneratePatient()

	if mustString(p1, "id") == mustString(p2, "id") {
		t.Fatal("different seeds should produce different patient ids")
	}
}

func TestDataGenerator_UniqueIDs(t *testing.T) {
	gen := NewDataGenerator(42)
	ids := make(map[string]bool)

	for i := 0; i < 50; i++ {
		p := gen.GeneratePatient()
		id := mustString(p, "id")
		if ids[id] {
			t.Fatalf("duplicate id found: %s", id)
		}
		ids[id] = true
	}

	for i := 0; i < 20; i++ {
		pr := gen.GeneratePractitioner()
		id := mustString(pr, "id")
		if ids[id] {
			t.Fatalf("duplicate id found: %s", id)
		}
		ids[id] = true
	}
}

// ---------------------------------------------------------------------------
// Seeder tests
// ---------------------------------------------------------------------------

func TestSeeder_Generate_DefaultConfig(t *testing.T) {
	cfg := DefaultSeedConfig()
	cfg.PatientCount = 5
	cfg.PractitionerCount = 2
	s := NewSeeder(cfg)

	result, err := s.Generate()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Patients != 5 {
		t.Fatalf("expected 5 patients, got %d", result.Patients)
	}
	if result.Practitioners != 2 {
		t.Fatalf("expected 2 practitioners, got %d", result.Practitioners)
	}
	if result.Organizations < 1 {
		t.Fatal("expected at least 1 organization")
	}
}

func TestSeeder_Generate_CustomConfig(t *testing.T) {
	cfg := SeedConfig{
		PatientCount:             3,
		EncountersPerPatient:     2,
		ObservationsPerEncounter: 1,
		ConditionsPerPatient:     1,
		MedicationsPerPatient:    1,
		AllergiesPerPatient:      1,
		ProceduresPerPatient:     1,
		ImmunizationsPerPatient:  1,
		IncludePractitioners:     true,
		PractitionerCount:        2,
		IncludeOrganization:      true,
		Seed:                     42,
	}
	s := NewSeeder(cfg)

	result, err := s.Generate()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Patients != 3 {
		t.Fatalf("expected 3 patients, got %d", result.Patients)
	}
	if result.Encounters != 6 { // 3 patients * 2 encounters
		t.Fatalf("expected 6 encounters, got %d", result.Encounters)
	}
	if result.Observations != 6 { // 6 encounters * 1 observation
		t.Fatalf("expected 6 observations, got %d", result.Observations)
	}
}

func TestSeeder_Generate_ResultCounts(t *testing.T) {
	cfg := SeedConfig{
		PatientCount:             2,
		EncountersPerPatient:     2,
		ObservationsPerEncounter: 2,
		ConditionsPerPatient:     1,
		MedicationsPerPatient:    1,
		AllergiesPerPatient:      1,
		ProceduresPerPatient:     1,
		ImmunizationsPerPatient:  1,
		IncludePractitioners:     true,
		PractitionerCount:        3,
		IncludeOrganization:      true,
		Seed:                     42,
	}
	s := NewSeeder(cfg)

	result, err := s.Generate()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := result.Patients + result.Practitioners + result.Organizations +
		result.Encounters + result.Observations + result.Conditions +
		result.Medications + result.Allergies + result.Procedures + result.Immunizations
	if result.TotalResources != expected {
		t.Fatalf("TotalResources %d != sum of components %d", result.TotalResources, expected)
	}
	if result.Duration <= 0 {
		t.Fatal("expected positive duration")
	}
}

func TestSeeder_GetResources(t *testing.T) {
	cfg := DefaultSeedConfig()
	cfg.PatientCount = 3
	cfg.PractitionerCount = 2
	s := NewSeeder(cfg)

	_, err := s.Generate()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	patients := s.GetResources("Patient")
	if len(patients) != 3 {
		t.Fatalf("expected 3 patients, got %d", len(patients))
	}
	practitioners := s.GetResources("Practitioner")
	if len(practitioners) != 2 {
		t.Fatalf("expected 2 practitioners, got %d", len(practitioners))
	}
}

func TestSeeder_ExportNDJSON(t *testing.T) {
	cfg := DefaultSeedConfig()
	cfg.PatientCount = 3
	cfg.PractitionerCount = 1
	s := NewSeeder(cfg)

	_, err := s.Generate()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var buf bytes.Buffer
	if err := s.ExportNDJSON(&buf, "Patient"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 NDJSON lines, got %d", len(lines))
	}

	// Each line must be valid JSON
	for i, line := range lines {
		var m map[string]interface{}
		if err := json.Unmarshal([]byte(line), &m); err != nil {
			t.Fatalf("line %d is not valid JSON: %v", i, err)
		}
		if m["resourceType"] != "Patient" {
			t.Fatalf("line %d: expected resourceType Patient", i)
		}
	}
}

func TestSeeder_ExportBundle(t *testing.T) {
	cfg := DefaultSeedConfig()
	cfg.PatientCount = 2
	cfg.PractitionerCount = 1
	cfg.EncountersPerPatient = 1
	cfg.ObservationsPerEncounter = 1
	cfg.ConditionsPerPatient = 1
	cfg.MedicationsPerPatient = 0
	cfg.AllergiesPerPatient = 0
	cfg.ProceduresPerPatient = 0
	cfg.ImmunizationsPerPatient = 0
	s := NewSeeder(cfg)

	_, err := s.Generate()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var buf bytes.Buffer
	if err := s.ExportBundle(&buf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var bundle map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &bundle); err != nil {
		t.Fatalf("bundle is not valid JSON: %v", err)
	}
	if bundle["resourceType"] != "Bundle" {
		t.Fatal("expected resourceType Bundle")
	}
	if bundle["type"] != "transaction" {
		t.Fatal("expected type transaction")
	}
	entries := mustSlice(bundle, "entry")
	if len(entries) == 0 {
		t.Fatal("expected entries in bundle")
	}
	// Check first entry has resource and request
	entry0 := mustMap(entries[0])
	if entry0["resource"] == nil {
		t.Fatal("expected resource in entry")
	}
	if entry0["request"] == nil {
		t.Fatal("expected request in entry")
	}
}

func TestSeeder_Generate_ReferencesValid(t *testing.T) {
	cfg := DefaultSeedConfig()
	cfg.PatientCount = 3
	cfg.EncountersPerPatient = 2
	cfg.PractitionerCount = 2
	s := NewSeeder(cfg)

	_, err := s.Generate()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Collect patient IDs
	patientIDs := make(map[string]bool)
	for _, p := range s.GetResources("Patient") {
		patientIDs["Patient/"+mustString(p, "id")] = true
	}

	// Check encounter references
	for _, enc := range s.GetResources("Encounter") {
		subj := mustMap(enc["subject"])
		ref := mustString(subj, "reference")
		if !patientIDs[ref] {
			t.Fatalf("encounter references non-existent patient: %s", ref)
		}
	}
}

func TestSeeder_Generate_ClinicalDataRealistic(t *testing.T) {
	cfg := DefaultSeedConfig()
	cfg.PatientCount = 5
	cfg.EncountersPerPatient = 2
	cfg.ObservationsPerEncounter = 3
	s := NewSeeder(cfg)

	_, err := s.Generate()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, obs := range s.GetResources("Observation") {
		vq := mustMap(obs["valueQuantity"])
		val, ok := vq["value"].(float64)
		if !ok {
			t.Fatal("expected float64 value in observation valueQuantity")
		}
		if val <= 0 || val > 1000 {
			t.Fatalf("observation value out of realistic range: %f", val)
		}
	}
}

// ---------------------------------------------------------------------------
// Handler tests
// ---------------------------------------------------------------------------

func setupTestEcho() (*echo.Echo, *SeedHandler) {
	e := echo.New()
	h := NewSeedHandler()
	g := e.Group("/admin/sandbox")
	h.RegisterRoutes(g)
	return e, h
}

func TestSeedHandler_Seed(t *testing.T) {
	e, _ := setupTestEcho()

	body := `{"patientCount":3,"encountersPerPatient":1,"observationsPerEncounter":1,"conditionsPerPatient":1,"medicationsPerPatient":0,"allergiesPerPatient":0,"proceduresPerPatient":0,"immunizationsPerPatient":0,"includePractitioners":true,"practitionerCount":2,"includeOrganization":true,"seed":42}`
	req := httptest.NewRequest(http.MethodPost, "/admin/sandbox/seed", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var result SeedResult
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON response: %v", err)
	}
	if result.Patients != 3 {
		t.Fatalf("expected 3 patients, got %d", result.Patients)
	}
	if result.TotalResources == 0 {
		t.Fatal("expected non-zero total resources")
	}
}

func TestSeedHandler_ListResources(t *testing.T) {
	e, h := setupTestEcho()

	// First seed data
	cfg := DefaultSeedConfig()
	cfg.PatientCount = 2
	cfg.PractitionerCount = 1
	h.seeder = NewSeeder(cfg)
	_, _ = h.seeder.Generate()

	req := httptest.NewRequest(http.MethodGet, "/admin/sandbox/resources/Patient", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resources []map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &resources); err != nil {
		t.Fatalf("invalid JSON response: %v", err)
	}
	if len(resources) != 2 {
		t.Fatalf("expected 2 patients, got %d", len(resources))
	}
}

func TestSeedHandler_Reset(t *testing.T) {
	e, h := setupTestEcho()

	// First seed data
	cfg := DefaultSeedConfig()
	cfg.PatientCount = 2
	cfg.PractitionerCount = 1
	h.seeder = NewSeeder(cfg)
	_, _ = h.seeder.Generate()

	req := httptest.NewRequest(http.MethodPost, "/admin/sandbox/reset", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	// Verify data is cleared
	patients := h.seeder.GetResources("Patient")
	if len(patients) != 0 {
		t.Fatalf("expected 0 patients after reset, got %d", len(patients))
	}
}

func TestSeedHandler_ExportNDJSON(t *testing.T) {
	e, h := setupTestEcho()

	// First seed data
	cfg := DefaultSeedConfig()
	cfg.PatientCount = 3
	cfg.PractitionerCount = 1
	h.seeder = NewSeeder(cfg)
	_, _ = h.seeder.Generate()

	req := httptest.NewRequest(http.MethodGet, "/admin/sandbox/export/ndjson/Patient", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	ct := rec.Header().Get(echo.HeaderContentType)
	if !strings.Contains(ct, "application/x-ndjson") {
		t.Fatalf("expected application/x-ndjson content type, got %s", ct)
	}

	lines := strings.Split(strings.TrimSpace(rec.Body.String()), "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 NDJSON lines, got %d", len(lines))
	}
}

func TestSeedHandler_ExportBundle(t *testing.T) {
	e, h := setupTestEcho()

	cfg := DefaultSeedConfig()
	cfg.PatientCount = 2
	cfg.PractitionerCount = 1
	cfg.EncountersPerPatient = 1
	cfg.ObservationsPerEncounter = 0
	cfg.ConditionsPerPatient = 0
	cfg.MedicationsPerPatient = 0
	cfg.AllergiesPerPatient = 0
	cfg.ProceduresPerPatient = 0
	cfg.ImmunizationsPerPatient = 0
	h.seeder = NewSeeder(cfg)
	_, _ = h.seeder.Generate()

	req := httptest.NewRequest(http.MethodGet, "/admin/sandbox/export/bundle", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var bundle map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &bundle); err != nil {
		t.Fatalf("invalid JSON response: %v", err)
	}
	if bundle["resourceType"] != "Bundle" {
		t.Fatal("expected resourceType Bundle")
	}
}
