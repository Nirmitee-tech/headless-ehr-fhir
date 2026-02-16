package fhir

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/labstack/echo/v4"
)

// ===========================================================================
// Store Tests
// ===========================================================================

func TestStructureDefinitionStore_NewStore(t *testing.T) {
	store := NewStructureDefinitionStore()
	if store == nil {
		t.Fatal("expected non-nil store")
	}
}

func TestStructureDefinitionStore_RegisterAndGet(t *testing.T) {
	store := NewStructureDefinitionStore()
	sd := &StructureDefinitionResource{
		ResourceType: "StructureDefinition",
		ID:           "test-sd-1",
		URL:          "http://example.org/StructureDefinition/test-1",
		Name:         "TestDef",
		Status:       "active",
		Kind:         "resource",
		Type:         "Patient",
	}
	store.Register(sd)

	got := store.Get("test-sd-1")
	if got == nil {
		t.Fatal("expected to find registered StructureDefinition")
	}
	if got.Name != "TestDef" {
		t.Errorf("expected name TestDef, got %s", got.Name)
	}
}

func TestStructureDefinitionStore_GetNotFound(t *testing.T) {
	store := NewStructureDefinitionStore()
	got := store.Get("nonexistent")
	if got != nil {
		t.Errorf("expected nil for nonexistent ID, got %+v", got)
	}
}

func TestStructureDefinitionStore_SearchByName(t *testing.T) {
	store := NewStructureDefinitionStore()
	store.Register(&StructureDefinitionResource{
		ResourceType: "StructureDefinition", ID: "sd-a", URL: "http://example.org/a",
		Name: "AlphaResource", Status: "active", Kind: "resource", Type: "Alpha",
	})
	store.Register(&StructureDefinitionResource{
		ResourceType: "StructureDefinition", ID: "sd-b", URL: "http://example.org/b",
		Name: "BetaResource", Status: "active", Kind: "resource", Type: "Beta",
	})

	results := store.Search(map[string]string{"name": "AlphaResource"})
	if len(results) != 1 {
		t.Fatalf("expected 1 result for name search, got %d", len(results))
	}
	if results[0].ID != "sd-a" {
		t.Errorf("expected ID sd-a, got %s", results[0].ID)
	}
}

func TestStructureDefinitionStore_SearchByType(t *testing.T) {
	store := NewStructureDefinitionStore()
	store.Register(&StructureDefinitionResource{
		ResourceType: "StructureDefinition", ID: "sd-patient", URL: "http://example.org/Patient",
		Name: "Patient", Status: "active", Kind: "resource", Type: "Patient",
	})
	store.Register(&StructureDefinitionResource{
		ResourceType: "StructureDefinition", ID: "sd-obs", URL: "http://example.org/Observation",
		Name: "Observation", Status: "active", Kind: "resource", Type: "Observation",
	})

	results := store.Search(map[string]string{"type": "Patient"})
	if len(results) != 1 {
		t.Fatalf("expected 1 result for type search, got %d", len(results))
	}
	if results[0].ID != "sd-patient" {
		t.Errorf("expected ID sd-patient, got %s", results[0].ID)
	}
}

func TestStructureDefinitionStore_SearchByURL(t *testing.T) {
	store := NewStructureDefinitionStore()
	store.Register(&StructureDefinitionResource{
		ResourceType: "StructureDefinition", ID: "sd-1", URL: "http://example.org/fhir/StructureDefinition/MyProfile",
		Name: "MyProfile", Status: "active", Kind: "resource", Type: "Patient",
	})

	results := store.Search(map[string]string{"url": "http://example.org/fhir/StructureDefinition/MyProfile"})
	if len(results) != 1 {
		t.Fatalf("expected 1 result for URL search, got %d", len(results))
	}
}

func TestStructureDefinitionStore_SearchByStatus(t *testing.T) {
	store := NewStructureDefinitionStore()
	store.Register(&StructureDefinitionResource{
		ResourceType: "StructureDefinition", ID: "sd-active", URL: "http://example.org/active",
		Name: "ActiveDef", Status: "active", Kind: "resource", Type: "Patient",
	})
	store.Register(&StructureDefinitionResource{
		ResourceType: "StructureDefinition", ID: "sd-draft", URL: "http://example.org/draft",
		Name: "DraftDef", Status: "draft", Kind: "resource", Type: "Patient",
	})

	results := store.Search(map[string]string{"status": "draft"})
	if len(results) != 1 {
		t.Fatalf("expected 1 result for status search, got %d", len(results))
	}
	if results[0].ID != "sd-draft" {
		t.Errorf("expected ID sd-draft, got %s", results[0].ID)
	}
}

func TestStructureDefinitionStore_SearchMultipleParams(t *testing.T) {
	store := NewStructureDefinitionStore()
	store.Register(&StructureDefinitionResource{
		ResourceType: "StructureDefinition", ID: "sd-1", URL: "http://example.org/1",
		Name: "PatientProfile", Status: "active", Kind: "resource", Type: "Patient",
	})
	store.Register(&StructureDefinitionResource{
		ResourceType: "StructureDefinition", ID: "sd-2", URL: "http://example.org/2",
		Name: "PatientDraft", Status: "draft", Kind: "resource", Type: "Patient",
	})
	store.Register(&StructureDefinitionResource{
		ResourceType: "StructureDefinition", ID: "sd-3", URL: "http://example.org/3",
		Name: "ObsProfile", Status: "active", Kind: "resource", Type: "Observation",
	})

	results := store.Search(map[string]string{"type": "Patient", "status": "active"})
	if len(results) != 1 {
		t.Fatalf("expected 1 result for combined search, got %d", len(results))
	}
	if results[0].ID != "sd-1" {
		t.Errorf("expected ID sd-1, got %s", results[0].ID)
	}
}

func TestStructureDefinitionStore_SearchNoParams(t *testing.T) {
	store := NewStructureDefinitionStore()
	store.Register(&StructureDefinitionResource{
		ResourceType: "StructureDefinition", ID: "sd-a", URL: "http://example.org/a",
		Name: "A", Status: "active", Kind: "resource", Type: "Alpha",
	})
	store.Register(&StructureDefinitionResource{
		ResourceType: "StructureDefinition", ID: "sd-b", URL: "http://example.org/b",
		Name: "B", Status: "draft", Kind: "resource", Type: "Beta",
	})

	results := store.Search(map[string]string{})
	if len(results) != 2 {
		t.Errorf("expected 2 results for empty search, got %d", len(results))
	}
}

func TestStructureDefinitionStore_SearchCaseInsensitiveName(t *testing.T) {
	store := NewStructureDefinitionStore()
	store.Register(&StructureDefinitionResource{
		ResourceType: "StructureDefinition", ID: "sd-1", URL: "http://example.org/1",
		Name: "PatientProfile", Status: "active", Kind: "resource", Type: "Patient",
	})

	results := store.Search(map[string]string{"name": "patientprofile"})
	if len(results) != 1 {
		t.Errorf("expected case-insensitive name match, got %d results", len(results))
	}
}

func TestStructureDefinitionStore_Concurrent(t *testing.T) {
	store := NewStructureDefinitionStore()
	RegisterBaseDefinitions(store)

	var wg sync.WaitGroup
	errs := make(chan string, 40)

	// Concurrent reads
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sd := store.Get("Patient")
			if sd == nil {
				errs <- "Patient not found in concurrent read"
			}
		}()
	}

	// Concurrent searches
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			results := store.Search(map[string]string{"status": "active"})
			if len(results) == 0 {
				errs <- "no results in concurrent search"
			}
		}()
	}

	wg.Wait()
	close(errs)

	for e := range errs {
		t.Errorf("concurrent error: %s", e)
	}
}

// ===========================================================================
// Base Definitions Tests
// ===========================================================================

func TestRegisterBaseDefinitions_AllPresent(t *testing.T) {
	store := NewStructureDefinitionStore()
	RegisterBaseDefinitions(store)

	expectedResources := []string{
		"Patient", "Observation", "Condition", "Encounter", "MedicationRequest",
		"Procedure", "DiagnosticReport", "AllergyIntolerance", "Immunization",
		"CarePlan", "Medication", "Practitioner", "Organization", "Location",
		"ServiceRequest", "DocumentReference", "MedicationAdministration",
		"Goal", "Claim",
	}

	for _, name := range expectedResources {
		sd := store.Get(name)
		if sd == nil {
			t.Errorf("expected base definition for %s, not found", name)
			continue
		}
		if sd.ResourceType != "StructureDefinition" {
			t.Errorf("%s: expected resourceType StructureDefinition, got %s", name, sd.ResourceType)
		}
		if sd.Status != "active" {
			t.Errorf("%s: expected status active, got %s", name, sd.Status)
		}
		if sd.Kind != "resource" {
			t.Errorf("%s: expected kind resource, got %s", name, sd.Kind)
		}
		if sd.FHIRVersion != "4.0.1" {
			t.Errorf("%s: expected fhirVersion 4.0.1, got %s", name, sd.FHIRVersion)
		}
		if sd.Type != name {
			t.Errorf("%s: expected type %s, got %s", name, name, sd.Type)
		}
	}
}

func TestRegisterBaseDefinitions_PatientSnapshot(t *testing.T) {
	store := NewStructureDefinitionStore()
	RegisterBaseDefinitions(store)

	sd := store.Get("Patient")
	if sd == nil {
		t.Fatal("expected Patient definition")
	}
	if sd.Snapshot == nil {
		t.Fatal("expected Patient to have snapshot")
	}
	if len(sd.Snapshot.Element) < 5 {
		t.Errorf("expected at least 5 elements in Patient snapshot, got %d", len(sd.Snapshot.Element))
	}

	// Verify base elements present
	paths := make(map[string]bool)
	for _, e := range sd.Snapshot.Element {
		paths[e.Path] = true
	}

	requiredPaths := []string{"Patient", "Patient.id", "Patient.meta", "Patient.text", "Patient.name", "Patient.gender", "Patient.birthDate"}
	for _, p := range requiredPaths {
		if !paths[p] {
			t.Errorf("expected path %s in Patient snapshot", p)
		}
	}
}

func TestRegisterBaseDefinitions_ObservationRequiredElements(t *testing.T) {
	store := NewStructureDefinitionStore()
	RegisterBaseDefinitions(store)

	sd := store.Get("Observation")
	if sd == nil {
		t.Fatal("expected Observation definition")
	}
	if sd.Snapshot == nil {
		t.Fatal("expected Observation to have snapshot")
	}

	// Observation.status and Observation.code should have min=1
	for _, e := range sd.Snapshot.Element {
		switch e.Path {
		case "Observation.status":
			if e.Min == nil || *e.Min != 1 {
				t.Errorf("expected Observation.status min=1")
			}
		case "Observation.code":
			if e.Min == nil || *e.Min != 1 {
				t.Errorf("expected Observation.code min=1")
			}
		}
	}
}

func TestRegisterBaseDefinitions_EncounterBinding(t *testing.T) {
	store := NewStructureDefinitionStore()
	RegisterBaseDefinitions(store)

	sd := store.Get("Encounter")
	if sd == nil {
		t.Fatal("expected Encounter definition")
	}

	for _, e := range sd.Snapshot.Element {
		if e.Path == "Encounter.status" {
			if e.Binding == nil {
				t.Fatal("expected binding on Encounter.status")
			}
			if e.Binding.Strength != "required" {
				t.Errorf("expected required binding strength, got %s", e.Binding.Strength)
			}
			if e.Binding.ValueSet != "http://hl7.org/fhir/ValueSet/encounter-status" {
				t.Errorf("unexpected valueSet: %s", e.Binding.ValueSet)
			}
			return
		}
	}
	t.Error("Encounter.status not found in snapshot")
}

func TestRegisterBaseDefinitions_BaseDefinitionAndDerivation(t *testing.T) {
	store := NewStructureDefinitionStore()
	RegisterBaseDefinitions(store)

	sd := store.Get("Patient")
	if sd == nil {
		t.Fatal("expected Patient definition")
	}
	if sd.BaseDefinition != "http://hl7.org/fhir/StructureDefinition/DomainResource" {
		t.Errorf("expected DomainResource base, got %s", sd.BaseDefinition)
	}
	if sd.Derivation != "specialization" {
		t.Errorf("expected specialization derivation, got %s", sd.Derivation)
	}
}

func TestRegisterBaseDefinitions_TargetProfile(t *testing.T) {
	store := NewStructureDefinitionStore()
	RegisterBaseDefinitions(store)

	sd := store.Get("Observation")
	if sd == nil {
		t.Fatal("expected Observation definition")
	}

	for _, e := range sd.Snapshot.Element {
		if e.Path == "Observation.subject" {
			if len(e.Type) == 0 {
				t.Fatal("expected type on Observation.subject")
			}
			if e.Type[0].Code != "Reference" {
				t.Errorf("expected Reference type, got %s", e.Type[0].Code)
			}
			if len(e.Type[0].TargetProfile) == 0 {
				t.Fatal("expected targetProfile on Observation.subject")
			}
			if e.Type[0].TargetProfile[0] != "http://hl7.org/fhir/StructureDefinition/Patient" {
				t.Errorf("unexpected targetProfile: %s", e.Type[0].TargetProfile[0])
			}
			return
		}
	}
	t.Error("Observation.subject not found in snapshot")
}

// ===========================================================================
// Snapshot Generation Tests
// ===========================================================================

func TestGenerateSnapshot_AlreadyHasSnapshot(t *testing.T) {
	store := NewStructureDefinitionStore()
	sd := &StructureDefinitionResource{
		ResourceType: "StructureDefinition", ID: "test",
		URL: "http://example.org/test", Name: "Test", Status: "active",
		Kind: "resource", Type: "Patient",
		Snapshot: &StructureSnapshot{
			Element: []ElementDefinition{
				{Path: "Patient", Short: "Existing snapshot"},
			},
		},
	}

	result := GenerateSnapshot(store, sd)
	if result.Snapshot == nil {
		t.Fatal("expected snapshot")
	}
	if len(result.Snapshot.Element) != 1 {
		t.Errorf("expected 1 element (unchanged), got %d", len(result.Snapshot.Element))
	}
	if result.Snapshot.Element[0].Short != "Existing snapshot" {
		t.Error("expected existing snapshot to remain unchanged")
	}
}

func TestGenerateSnapshot_FromBaseDefinition(t *testing.T) {
	store := NewStructureDefinitionStore()
	RegisterBaseDefinitions(store)

	// A profile that constrains Patient but has no snapshot or differential
	sd := &StructureDefinitionResource{
		ResourceType:   "StructureDefinition", ID: "my-patient",
		URL:            "http://example.org/StructureDefinition/MyPatient",
		Name:           "MyPatient", Status: "active",
		Kind:           "resource", Type: "Patient",
		BaseDefinition: "http://hl7.org/fhir/StructureDefinition/Patient",
		Derivation:     "constraint",
	}

	result := GenerateSnapshot(store, sd)
	if result.Snapshot == nil {
		t.Fatal("expected snapshot from base definition")
	}
	if len(result.Snapshot.Element) < 5 {
		t.Errorf("expected at least 5 elements from Patient base, got %d", len(result.Snapshot.Element))
	}
}

func TestGenerateSnapshot_MergesDifferential(t *testing.T) {
	store := NewStructureDefinitionStore()
	RegisterBaseDefinitions(store)

	// A profile that overrides Patient.gender to be required
	sd := &StructureDefinitionResource{
		ResourceType:   "StructureDefinition", ID: "strict-patient",
		URL:            "http://example.org/StructureDefinition/StrictPatient",
		Name:           "StrictPatient", Status: "active",
		Kind:           "resource", Type: "Patient",
		BaseDefinition: "http://hl7.org/fhir/StructureDefinition/Patient",
		Derivation:     "constraint",
		Differential: &StructureDifferential{
			Element: []ElementDefinition{
				{
					ID:          "Patient.gender",
					Path:        "Patient.gender",
					Short:       "Required gender",
					Min:         intPtr(1),
					Max:         "1",
					MustSupport: true,
				},
			},
		},
	}

	result := GenerateSnapshot(store, sd)
	if result.Snapshot == nil {
		t.Fatal("expected snapshot after merge")
	}

	// Find Patient.gender in the result
	for _, e := range result.Snapshot.Element {
		if e.Path == "Patient.gender" {
			if e.Min == nil || *e.Min != 1 {
				t.Error("expected Patient.gender min=1 from differential")
			}
			if !e.MustSupport {
				t.Error("expected Patient.gender mustSupport from differential")
			}
			if e.Short != "Required gender" {
				t.Errorf("expected overridden short text, got %s", e.Short)
			}
			return
		}
	}
	t.Error("Patient.gender not found in merged snapshot")
}

func TestGenerateSnapshot_DifferentialAddsElements(t *testing.T) {
	store := NewStructureDefinitionStore()
	RegisterBaseDefinitions(store)

	sd := &StructureDefinitionResource{
		ResourceType:   "StructureDefinition", ID: "ext-patient",
		URL:            "http://example.org/StructureDefinition/ExtPatient",
		Name:           "ExtPatient", Status: "active",
		Kind:           "resource", Type: "Patient",
		BaseDefinition: "http://hl7.org/fhir/StructureDefinition/Patient",
		Derivation:     "constraint",
		Differential: &StructureDifferential{
			Element: []ElementDefinition{
				{
					ID:   "Patient.extension:race",
					Path: "Patient.extension",
					Short: "US Core Race Extension",
					MustSupport: true,
				},
			},
		},
	}

	result := GenerateSnapshot(store, sd)
	if result.Snapshot == nil {
		t.Fatal("expected snapshot")
	}

	// The result should have the base elements plus the new extension element
	patientBase := store.Get("Patient")
	baseCount := len(patientBase.Snapshot.Element)
	if len(result.Snapshot.Element) != baseCount+1 {
		t.Errorf("expected %d elements (base %d + 1 new), got %d",
			baseCount+1, baseCount, len(result.Snapshot.Element))
	}
}

func TestGenerateSnapshot_NoDifferentialNoBase(t *testing.T) {
	store := NewStructureDefinitionStore()
	sd := &StructureDefinitionResource{
		ResourceType: "StructureDefinition", ID: "orphan",
		URL: "http://example.org/orphan", Name: "Orphan", Status: "active",
		Kind: "logical", Type: "Orphan",
	}

	result := GenerateSnapshot(store, sd)
	if result.Snapshot != nil {
		t.Error("expected nil snapshot for orphan definition with no base and no differential")
	}
}

// ===========================================================================
// Handler Tests
// ===========================================================================

func sdSetupRequest(method, path, query, body string) (echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	target := path
	if query != "" {
		target = path + "?" + query
	}
	var req *http.Request
	if body != "" {
		req = httptest.NewRequest(method, target, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(method, target, nil)
	}
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	return c, rec
}

func TestStructureDefinitionHandler_SearchAll(t *testing.T) {
	h := NewStructureDefinitionHandler()
	c, rec := sdSetupRequest(http.MethodGet, "/fhir/StructureDefinition", "", "")

	err := h.SearchStructureDefinitions(c)
	if err != nil {
		t.Fatalf("SearchStructureDefinitions error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var bundle map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &bundle)
	if bundle["resourceType"] != "Bundle" {
		t.Errorf("expected Bundle, got %v", bundle["resourceType"])
	}
	total := bundle["total"]
	if total == nil {
		t.Fatal("expected total in bundle")
	}
	// We registered at least 19 base definitions
	totalFloat, ok := total.(float64)
	if !ok {
		t.Fatalf("expected total as number, got %T", total)
	}
	if totalFloat < 19 {
		t.Errorf("expected at least 19 base definitions, got %v", totalFloat)
	}
}

func TestStructureDefinitionHandler_SearchByType(t *testing.T) {
	h := NewStructureDefinitionHandler()
	c, rec := sdSetupRequest(http.MethodGet, "/fhir/StructureDefinition", "type=Patient", "")

	err := h.SearchStructureDefinitions(c)
	if err != nil {
		t.Fatalf("SearchStructureDefinitions error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var bundle map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &bundle)
	totalFloat, _ := bundle["total"].(float64)
	if totalFloat != 1 {
		t.Errorf("expected 1 result for type=Patient, got %v", totalFloat)
	}
}

func TestStructureDefinitionHandler_SearchByStatus(t *testing.T) {
	h := NewStructureDefinitionHandler()
	c, rec := sdSetupRequest(http.MethodGet, "/fhir/StructureDefinition", "status=active", "")

	err := h.SearchStructureDefinitions(c)
	if err != nil {
		t.Fatalf("SearchStructureDefinitions error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var bundle map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &bundle)
	totalFloat, _ := bundle["total"].(float64)
	if totalFloat < 19 {
		t.Errorf("expected at least 19 active definitions, got %v", totalFloat)
	}
}

func TestStructureDefinitionHandler_SearchByName(t *testing.T) {
	h := NewStructureDefinitionHandler()
	c, rec := sdSetupRequest(http.MethodGet, "/fhir/StructureDefinition", "name=Observation", "")

	err := h.SearchStructureDefinitions(c)
	if err != nil {
		t.Fatalf("SearchStructureDefinitions error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var bundle map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &bundle)
	totalFloat, _ := bundle["total"].(float64)
	if totalFloat != 1 {
		t.Errorf("expected 1 result for name=Observation, got %v", totalFloat)
	}
}

func TestStructureDefinitionHandler_SearchByURL(t *testing.T) {
	h := NewStructureDefinitionHandler()
	c, rec := sdSetupRequest(http.MethodGet, "/fhir/StructureDefinition", "url=http://hl7.org/fhir/StructureDefinition/Condition", "")

	err := h.SearchStructureDefinitions(c)
	if err != nil {
		t.Fatalf("SearchStructureDefinitions error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var bundle map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &bundle)
	totalFloat, _ := bundle["total"].(float64)
	if totalFloat != 1 {
		t.Errorf("expected 1 result for Condition URL, got %v", totalFloat)
	}
}

func TestStructureDefinitionHandler_SearchNoResults(t *testing.T) {
	h := NewStructureDefinitionHandler()
	c, rec := sdSetupRequest(http.MethodGet, "/fhir/StructureDefinition", "type=NonExistent", "")

	err := h.SearchStructureDefinitions(c)
	if err != nil {
		t.Fatalf("SearchStructureDefinitions error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var bundle map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &bundle)
	totalFloat, _ := bundle["total"].(float64)
	if totalFloat != 0 {
		t.Errorf("expected 0 results for nonexistent type, got %v", totalFloat)
	}
}

func TestStructureDefinitionHandler_GetByID(t *testing.T) {
	h := NewStructureDefinitionHandler()
	c, rec := sdSetupRequest(http.MethodGet, "/fhir/StructureDefinition/Patient", "", "")
	c.SetParamNames("id")
	c.SetParamValues("Patient")

	err := h.GetStructureDefinition(c)
	if err != nil {
		t.Fatalf("GetStructureDefinition error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var result map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &result)
	if result["resourceType"] != "StructureDefinition" {
		t.Errorf("expected StructureDefinition, got %v", result["resourceType"])
	}
	if result["id"] != "Patient" {
		t.Errorf("expected id Patient, got %v", result["id"])
	}
	if result["name"] != "Patient" {
		t.Errorf("expected name Patient, got %v", result["name"])
	}
}

func TestStructureDefinitionHandler_GetByID_NotFound(t *testing.T) {
	h := NewStructureDefinitionHandler()
	c, rec := sdSetupRequest(http.MethodGet, "/fhir/StructureDefinition/NonExistent", "", "")
	c.SetParamNames("id")
	c.SetParamValues("NonExistent")

	err := h.GetStructureDefinition(c)
	if err != nil {
		t.Fatalf("GetStructureDefinition error: %v", err)
	}
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestStructureDefinitionHandler_GetByID_HasSnapshot(t *testing.T) {
	h := NewStructureDefinitionHandler()
	c, rec := sdSetupRequest(http.MethodGet, "/fhir/StructureDefinition/Observation", "", "")
	c.SetParamNames("id")
	c.SetParamValues("Observation")

	err := h.GetStructureDefinition(c)
	if err != nil {
		t.Fatalf("GetStructureDefinition error: %v", err)
	}

	var result map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &result)

	snapshot, ok := result["snapshot"].(map[string]interface{})
	if !ok {
		t.Fatal("expected snapshot in response")
	}
	elements, ok := snapshot["element"].([]interface{})
	if !ok {
		t.Fatal("expected elements array in snapshot")
	}
	if len(elements) < 5 {
		t.Errorf("expected at least 5 elements in Observation snapshot, got %d", len(elements))
	}
}

func TestStructureDefinitionHandler_SnapshotOp(t *testing.T) {
	h := NewStructureDefinitionHandler()
	c, rec := sdSetupRequest(http.MethodGet, "/fhir/StructureDefinition/$snapshot", "url=http://hl7.org/fhir/StructureDefinition/Patient", "")

	err := h.GenerateSnapshotOp(c)
	if err != nil {
		t.Fatalf("GenerateSnapshotOp error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var result map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &result)
	if result["resourceType"] != "StructureDefinition" {
		t.Errorf("expected StructureDefinition, got %v", result["resourceType"])
	}
	snapshot, ok := result["snapshot"].(map[string]interface{})
	if !ok {
		t.Fatal("expected snapshot in $snapshot response")
	}
	elements, _ := snapshot["element"].([]interface{})
	if len(elements) < 5 {
		t.Errorf("expected at least 5 elements in Patient snapshot, got %d", len(elements))
	}
}

func TestStructureDefinitionHandler_SnapshotOp_MissingURL(t *testing.T) {
	h := NewStructureDefinitionHandler()
	c, rec := sdSetupRequest(http.MethodGet, "/fhir/StructureDefinition/$snapshot", "", "")

	err := h.GenerateSnapshotOp(c)
	if err != nil {
		t.Fatalf("GenerateSnapshotOp error: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for missing url, got %d", rec.Code)
	}
}

func TestStructureDefinitionHandler_SnapshotOp_NotFound(t *testing.T) {
	h := NewStructureDefinitionHandler()
	c, rec := sdSetupRequest(http.MethodGet, "/fhir/StructureDefinition/$snapshot", "url=http://example.org/nonexistent", "")

	err := h.GenerateSnapshotOp(c)
	if err != nil {
		t.Fatalf("GenerateSnapshotOp error: %v", err)
	}
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404 for unknown URL, got %d", rec.Code)
	}
}

func TestStructureDefinitionHandler_SnapshotFromBody(t *testing.T) {
	h := NewStructureDefinitionHandler()

	bodyJSON := `{
		"resourceType": "Parameters",
		"parameter": [{
			"name": "definition",
			"resource": {
				"resourceType": "StructureDefinition",
				"id": "custom-patient",
				"url": "http://example.org/StructureDefinition/CustomPatient",
				"name": "CustomPatient",
				"status": "active",
				"kind": "resource",
				"type": "Patient",
				"baseDefinition": "http://hl7.org/fhir/StructureDefinition/Patient",
				"derivation": "constraint",
				"differential": {
					"element": [{
						"id": "Patient.name",
						"path": "Patient.name",
						"min": 1,
						"max": "1",
						"mustSupport": true
					}]
				}
			}
		}]
	}`

	c, rec := sdSetupRequest(http.MethodPost, "/fhir/StructureDefinition/$snapshot", "", bodyJSON)

	err := h.GenerateSnapshotFromBody(c)
	if err != nil {
		t.Fatalf("GenerateSnapshotFromBody error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var result map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &result)
	snapshot, ok := result["snapshot"].(map[string]interface{})
	if !ok {
		t.Fatal("expected snapshot in response")
	}
	elements, _ := snapshot["element"].([]interface{})
	if len(elements) == 0 {
		t.Fatal("expected elements in snapshot")
	}

	// Verify Patient.name was overridden
	for _, elem := range elements {
		e, ok := elem.(map[string]interface{})
		if !ok {
			continue
		}
		if e["path"] == "Patient.name" {
			if ms, ok := e["mustSupport"].(bool); !ok || !ms {
				t.Error("expected Patient.name mustSupport=true from differential")
			}
			return
		}
	}
	t.Error("Patient.name not found in merged snapshot")
}

func TestStructureDefinitionHandler_SnapshotFromBody_MissingParameter(t *testing.T) {
	h := NewStructureDefinitionHandler()

	bodyJSON := `{
		"resourceType": "Parameters",
		"parameter": [{
			"name": "other",
			"resource": null
		}]
	}`

	c, rec := sdSetupRequest(http.MethodPost, "/fhir/StructureDefinition/$snapshot", "", bodyJSON)

	err := h.GenerateSnapshotFromBody(c)
	if err != nil {
		t.Fatalf("GenerateSnapshotFromBody error: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for missing definition parameter, got %d", rec.Code)
	}
}

// ===========================================================================
// JSON Serialization Tests
// ===========================================================================

func TestStructureDefinitionResource_JSONRoundTrip(t *testing.T) {
	sd := &StructureDefinitionResource{
		ResourceType:   "StructureDefinition",
		ID:             "test-json",
		URL:            "http://example.org/test",
		Name:           "TestJSON",
		Title:          "Test JSON Roundtrip",
		Status:         "active",
		Kind:           "resource",
		Abstract:       false,
		Type:           "Patient",
		BaseDefinition: "http://hl7.org/fhir/StructureDefinition/DomainResource",
		Derivation:     "specialization",
		Description:    "A test definition",
		FHIRVersion:    "4.0.1",
		Snapshot: &StructureSnapshot{
			Element: []ElementDefinition{
				{
					ID:    "Patient",
					Path:  "Patient",
					Short: "Patient resource",
					Min:   intPtr(0),
					Max:   "*",
				},
				{
					ID:    "Patient.id",
					Path:  "Patient.id",
					Short: "Logical id",
					Min:   intPtr(0),
					Max:   "1",
					Type:  []ElementType{{Code: "id"}},
				},
				{
					ID:          "Patient.gender",
					Path:        "Patient.gender",
					Short:       "Gender",
					Min:         intPtr(0),
					Max:         "1",
					Type:        []ElementType{{Code: "code"}},
					Binding:     &ElementBinding{Strength: "required", ValueSet: "http://hl7.org/fhir/ValueSet/administrative-gender"},
					MustSupport: true,
				},
			},
		},
	}

	data, err := json.Marshal(sd)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var decoded StructureDefinitionResource
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if decoded.ResourceType != "StructureDefinition" {
		t.Errorf("expected StructureDefinition, got %s", decoded.ResourceType)
	}
	if decoded.ID != "test-json" {
		t.Errorf("expected test-json, got %s", decoded.ID)
	}
	if decoded.Snapshot == nil {
		t.Fatal("expected snapshot after roundtrip")
	}
	if len(decoded.Snapshot.Element) != 3 {
		t.Errorf("expected 3 elements, got %d", len(decoded.Snapshot.Element))
	}

	// Verify binding
	genderElem := decoded.Snapshot.Element[2]
	if genderElem.Binding == nil {
		t.Fatal("expected binding on gender element")
	}
	if genderElem.Binding.Strength != "required" {
		t.Errorf("expected required strength, got %s", genderElem.Binding.Strength)
	}
	if !genderElem.MustSupport {
		t.Error("expected mustSupport true")
	}
}

func TestElementDefinition_JSONOmitsEmpty(t *testing.T) {
	ed := ElementDefinition{
		Path: "Patient.id",
	}

	data, err := json.Marshal(ed)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var m map[string]interface{}
	json.Unmarshal(data, &m)

	if _, ok := m["id"]; ok {
		t.Error("expected omitted empty id field")
	}
	if _, ok := m["min"]; ok {
		t.Error("expected omitted nil min field")
	}
	if _, ok := m["type"]; ok {
		t.Error("expected omitted nil type field")
	}
	if _, ok := m["binding"]; ok {
		t.Error("expected omitted nil binding field")
	}
	if _, ok := m["mustSupport"]; ok {
		t.Error("expected omitted false mustSupport field")
	}
	if _, ok := m["path"]; !ok {
		t.Error("expected path field to be present")
	}
}

// ===========================================================================
// Integration-style Tests
// ===========================================================================

func TestStructureDefinitionHandler_RegisterRoutes(t *testing.T) {
	e := echo.New()
	g := e.Group("/fhir")
	h := NewStructureDefinitionHandler()
	h.RegisterRoutes(g)

	// Verify routes by making requests to the registered paths
	req := httptest.NewRequest(http.MethodGet, "/fhir/StructureDefinition", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code == http.StatusNotFound {
		t.Error("expected /fhir/StructureDefinition route to be registered")
	}

	req2 := httptest.NewRequest(http.MethodGet, "/fhir/StructureDefinition/Patient", nil)
	rec2 := httptest.NewRecorder()
	e.ServeHTTP(rec2, req2)

	if rec2.Code == http.StatusNotFound {
		t.Error("expected /fhir/StructureDefinition/:id route to be registered")
	}
}

func TestStructureDefinitionHandler_SearchMultipleFilters(t *testing.T) {
	h := NewStructureDefinitionHandler()
	c, rec := sdSetupRequest(http.MethodGet, "/fhir/StructureDefinition", "type=Patient&status=active", "")

	err := h.SearchStructureDefinitions(c)
	if err != nil {
		t.Fatalf("SearchStructureDefinitions error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var bundle map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &bundle)
	totalFloat, _ := bundle["total"].(float64)
	if totalFloat != 1 {
		t.Errorf("expected 1 result for type=Patient&status=active, got %v", totalFloat)
	}
}
