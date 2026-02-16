package fhir

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
)

// =========== FHIRCompartmentDefinition struct tests ===========

func TestPatientCompartmentDef_Structure(t *testing.T) {
	def := PatientCompartmentDef()

	if def.ResourceType != "CompartmentDefinition" {
		t.Errorf("expected resourceType CompartmentDefinition, got %q", def.ResourceType)
	}
	if def.ID != "patient" {
		t.Errorf("expected ID patient, got %q", def.ID)
	}
	if def.URL != "http://hl7.org/fhir/CompartmentDefinition/patient" {
		t.Errorf("expected URL http://hl7.org/fhir/CompartmentDefinition/patient, got %q", def.URL)
	}
	if def.Name != "Patient" {
		t.Errorf("expected Name Patient, got %q", def.Name)
	}
	if def.Status != "active" {
		t.Errorf("expected Status active, got %q", def.Status)
	}
	if def.Code != "Patient" {
		t.Errorf("expected Code Patient, got %q", def.Code)
	}
	if !def.Search {
		t.Error("expected Search to be true")
	}
	if len(def.Resource) != 38 {
		t.Errorf("expected 38 resource entries in Patient compartment, got %d", len(def.Resource))
	}
}

func TestPatientCompartmentDef_ResourceMemberships(t *testing.T) {
	def := PatientCompartmentDef()
	resourceMap := make(map[string][]string)
	for _, r := range def.Resource {
		resourceMap[r.Code] = r.Param
	}

	tests := []struct {
		resourceType   string
		expectedParams []string
	}{
		{"Account", []string{"subject"}},
		{"AllergyIntolerance", []string{"patient", "recorder", "asserter"}},
		{"Appointment", []string{"actor"}},
		{"AuditEvent", []string{"patient"}},
		{"CarePlan", []string{"patient", "performer"}},
		{"CareTeam", []string{"patient", "participant"}},
		{"Claim", []string{"patient", "payee"}},
		{"ClinicalImpression", []string{"subject"}},
		{"Communication", []string{"subject", "sender", "recipient"}},
		{"Condition", []string{"patient", "asserter"}},
		{"Consent", []string{"patient"}},
		{"Coverage", []string{"patient", "subscriber", "beneficiary", "payor"}},
		{"DetectedIssue", []string{"patient"}},
		{"DeviceRequest", []string{"subject", "performer"}},
		{"DiagnosticReport", []string{"subject"}},
		{"DocumentReference", []string{"subject", "author"}},
		{"Encounter", []string{"patient"}},
		{"EpisodeOfCare", []string{"patient"}},
		{"ExplanationOfBenefit", []string{"patient", "payee"}},
		{"FamilyMemberHistory", []string{"patient"}},
		{"Goal", []string{"patient"}},
		{"ImagingStudy", []string{"patient"}},
		{"Immunization", []string{"patient"}},
		{"List", []string{"subject", "source"}},
		{"MedicationAdministration", []string{"patient", "performer", "subject"}},
		{"MedicationDispense", []string{"subject", "patient", "receiver"}},
		{"MedicationRequest", []string{"subject"}},
		{"MedicationStatement", []string{"subject"}},
		{"NutritionOrder", []string{"patient"}},
		{"Observation", []string{"subject", "performer"}},
		{"Procedure", []string{"patient", "performer"}},
		{"Provenance", []string{"patient"}},
		{"QuestionnaireResponse", []string{"subject", "author"}},
		{"RelatedPerson", []string{"patient"}},
		{"RiskAssessment", []string{"subject"}},
		{"Schedule", []string{"actor"}},
		{"ServiceRequest", []string{"subject", "performer"}},
		{"Specimen", []string{"subject"}},
	}

	for _, tt := range tests {
		t.Run(tt.resourceType, func(t *testing.T) {
			params, ok := resourceMap[tt.resourceType]
			if !ok {
				t.Fatalf("%s not found in Patient compartment definition", tt.resourceType)
			}
			if len(params) != len(tt.expectedParams) {
				t.Fatalf("expected %d params for %s, got %d", len(tt.expectedParams), tt.resourceType, len(params))
			}
			for i, expected := range tt.expectedParams {
				if params[i] != expected {
					t.Errorf("param[%d] for %s: expected %q, got %q", i, tt.resourceType, expected, params[i])
				}
			}
		})
	}
}

func TestEncounterCompartmentDef_Structure(t *testing.T) {
	def := EncounterCompartmentDef()

	if def.Code != "Encounter" {
		t.Errorf("expected Code Encounter, got %q", def.Code)
	}
	if def.ID != "encounter" {
		t.Errorf("expected ID encounter, got %q", def.ID)
	}
	if len(def.Resource) == 0 {
		t.Error("expected non-empty resource list for Encounter compartment")
	}

	// Verify Encounter itself is a member via _id
	found := false
	for _, r := range def.Resource {
		if r.Code == "Encounter" {
			found = true
			if len(r.Param) == 0 || r.Param[0] != "_id" {
				t.Errorf("expected Encounter to link via _id, got %v", r.Param)
			}
			break
		}
	}
	if !found {
		t.Error("expected Encounter resource type in Encounter compartment")
	}
}

func TestPractitionerCompartmentDef_Structure(t *testing.T) {
	def := PractitionerCompartmentDef()

	if def.Code != "Practitioner" {
		t.Errorf("expected Code Practitioner, got %q", def.Code)
	}
	if def.ID != "practitioner" {
		t.Errorf("expected ID practitioner, got %q", def.ID)
	}
	if len(def.Resource) == 0 {
		t.Error("expected non-empty resource list for Practitioner compartment")
	}

	// Verify Encounter is a member
	found := false
	for _, r := range def.Resource {
		if r.Code == "Encounter" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected Encounter resource type in Practitioner compartment")
	}
}

func TestRelatedPersonCompartmentDef_Structure(t *testing.T) {
	def := RelatedPersonCompartmentDef()

	if def.Code != "RelatedPerson" {
		t.Errorf("expected Code RelatedPerson, got %q", def.Code)
	}
	if def.ID != "relatedperson" {
		t.Errorf("expected ID relatedperson, got %q", def.ID)
	}
	if len(def.Resource) == 0 {
		t.Error("expected non-empty resource list for RelatedPerson compartment")
	}
}

func TestDeviceCompartmentDef_Structure(t *testing.T) {
	def := DeviceCompartmentDef()

	if def.Code != "Device" {
		t.Errorf("expected Code Device, got %q", def.Code)
	}
	if def.ID != "device" {
		t.Errorf("expected ID device, got %q", def.ID)
	}
	if len(def.Resource) == 0 {
		t.Error("expected non-empty resource list for Device compartment")
	}

	// Verify Observation is a member
	found := false
	for _, r := range def.Resource {
		if r.Code == "Observation" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected Observation resource type in Device compartment")
	}
}

// =========== JSON serialization tests ===========

func TestFHIRCompartmentDefinition_JSONRoundtrip(t *testing.T) {
	def := PatientCompartmentDef()

	data, err := json.Marshal(def)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded FHIRCompartmentDefinition
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.ResourceType != "CompartmentDefinition" {
		t.Errorf("resourceType mismatch after roundtrip: %q", decoded.ResourceType)
	}
	if decoded.Code != "Patient" {
		t.Errorf("code mismatch after roundtrip: %q", decoded.Code)
	}
	if len(decoded.Resource) != len(def.Resource) {
		t.Errorf("resource count mismatch: expected %d, got %d", len(def.Resource), len(decoded.Resource))
	}
}

func TestFHIRCompartmentDefinition_JSONFields(t *testing.T) {
	def := EncounterCompartmentDef()

	data, err := json.Marshal(def)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("failed to unmarshal to map: %v", err)
	}

	// Verify JSON field names match FHIR spec
	if raw["resourceType"] != "CompartmentDefinition" {
		t.Errorf("expected resourceType field, got %v", raw["resourceType"])
	}
	if raw["code"] != "Encounter" {
		t.Errorf("expected code Encounter, got %v", raw["code"])
	}
	if raw["search"] != true {
		t.Errorf("expected search true, got %v", raw["search"])
	}
	if _, ok := raw["resource"]; !ok {
		t.Error("expected resource field in JSON output")
	}
}

// =========== Lookup helper tests ===========

func TestGetCompartmentDefinitionByCode(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		wantName string
		wantNil  bool
	}{
		{"Patient exact case", "Patient", "Patient", false},
		{"Patient lowercase", "patient", "Patient", false},
		{"Encounter", "Encounter", "Encounter", false},
		{"Practitioner", "Practitioner", "Practitioner", false},
		{"RelatedPerson", "RelatedPerson", "RelatedPerson", false},
		{"Device", "Device", "Device", false},
		{"Unknown code", "Organization", "", true},
		{"Empty code", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			def := GetCompartmentDefinitionByCode(tt.code)
			if tt.wantNil {
				if def != nil {
					t.Errorf("expected nil for code %q, got %+v", tt.code, def)
				}
				return
			}
			if def == nil {
				t.Fatalf("expected non-nil for code %q", tt.code)
			}
			if def.Name != tt.wantName {
				t.Errorf("expected Name %q, got %q", tt.wantName, def.Name)
			}
		})
	}
}

func TestCompartmentResourceParams(t *testing.T) {
	def := PatientCompartmentDef()

	tests := []struct {
		name         string
		resourceType string
		wantLen      int
		wantFirst    string
	}{
		{"Observation has 2 params", "Observation", 2, "subject"},
		{"Encounter has 1 param", "Encounter", 1, "patient"},
		{"Coverage has 4 params", "Coverage", 4, "patient"},
		{"Unknown resource", "Organization", 0, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := CompartmentResourceParams(def, tt.resourceType)
			if tt.wantLen == 0 {
				if params != nil {
					t.Errorf("expected nil params for %s, got %v", tt.resourceType, params)
				}
				return
			}
			if len(params) != tt.wantLen {
				t.Fatalf("expected %d params for %s, got %d", tt.wantLen, tt.resourceType, len(params))
			}
			if params[0] != tt.wantFirst {
				t.Errorf("expected first param %q, got %q", tt.wantFirst, params[0])
			}
		})
	}
}

func TestCompartmentResourceParams_AllCompartments(t *testing.T) {
	// Ensure each compartment definition has at least one resource with params
	defs := []struct {
		name string
		def  *FHIRCompartmentDefinition
	}{
		{"Patient", PatientCompartmentDef()},
		{"Encounter", EncounterCompartmentDef()},
		{"Practitioner", PractitionerCompartmentDef()},
		{"RelatedPerson", RelatedPersonCompartmentDef()},
		{"Device", DeviceCompartmentDef()},
	}

	for _, d := range defs {
		t.Run(d.name, func(t *testing.T) {
			if len(d.def.Resource) == 0 {
				t.Errorf("%s compartment has no resources", d.name)
			}
			hasParams := false
			for _, r := range d.def.Resource {
				if len(r.Param) > 0 {
					hasParams = true
					break
				}
			}
			if !hasParams {
				t.Errorf("%s compartment has no resources with linking params", d.name)
			}
		})
	}
}

// =========== CompartmentDefinitionHandler tests ===========

func TestCompartmentDefinitionHandler_GetDefinition_Patient(t *testing.T) {
	h := NewCompartmentDefinitionHandler()
	e := echo.New()

	req := httptest.NewRequest(http.MethodGet, "/fhir/CompartmentDefinition/patient", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("patient")

	err := h.GetDefinition(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var def FHIRCompartmentDefinition
	if err := json.Unmarshal(rec.Body.Bytes(), &def); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if def.ResourceType != "CompartmentDefinition" {
		t.Errorf("expected resourceType CompartmentDefinition, got %q", def.ResourceType)
	}
	if def.Code != "Patient" {
		t.Errorf("expected code Patient, got %q", def.Code)
	}
}

func TestCompartmentDefinitionHandler_GetDefinition_Encounter(t *testing.T) {
	h := NewCompartmentDefinitionHandler()
	e := echo.New()

	req := httptest.NewRequest(http.MethodGet, "/fhir/CompartmentDefinition/encounter", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("encounter")

	err := h.GetDefinition(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var def FHIRCompartmentDefinition
	json.Unmarshal(rec.Body.Bytes(), &def)
	if def.Code != "Encounter" {
		t.Errorf("expected code Encounter, got %q", def.Code)
	}
}

func TestCompartmentDefinitionHandler_GetDefinition_NotFound(t *testing.T) {
	h := NewCompartmentDefinitionHandler()
	e := echo.New()

	req := httptest.NewRequest(http.MethodGet, "/fhir/CompartmentDefinition/unknown", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("unknown")

	err := h.GetDefinition(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}

	var outcome OperationOutcome
	json.Unmarshal(rec.Body.Bytes(), &outcome)
	if len(outcome.Issue) == 0 {
		t.Error("expected OperationOutcome with issues")
	}
}

func TestCompartmentDefinitionHandler_GetDefinition_EmptyID(t *testing.T) {
	h := NewCompartmentDefinitionHandler()
	e := echo.New()

	req := httptest.NewRequest(http.MethodGet, "/fhir/CompartmentDefinition/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("")

	err := h.GetDefinition(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestCompartmentDefinitionHandler_GetDefinition_AllFiveCompartments(t *testing.T) {
	h := NewCompartmentDefinitionHandler()
	e := echo.New()

	ids := []string{"patient", "encounter", "practitioner", "relatedperson", "device"}
	expectedCodes := []string{"Patient", "Encounter", "Practitioner", "RelatedPerson", "Device"}

	for i, id := range ids {
		t.Run(id, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/fhir/CompartmentDefinition/"+id, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("id")
			c.SetParamValues(id)

			err := h.GetDefinition(c)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if rec.Code != http.StatusOK {
				t.Errorf("expected 200 for %s, got %d", id, rec.Code)
			}

			var def FHIRCompartmentDefinition
			json.Unmarshal(rec.Body.Bytes(), &def)
			if def.Code != expectedCodes[i] {
				t.Errorf("expected code %q, got %q", expectedCodes[i], def.Code)
			}
		})
	}
}

func TestCompartmentDefinitionHandler_SearchDefinitions_NoFilter(t *testing.T) {
	h := NewCompartmentDefinitionHandler()
	e := echo.New()

	req := httptest.NewRequest(http.MethodGet, "/fhir/CompartmentDefinition", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.SearchDefinitions(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var bundle Bundle
	if err := json.Unmarshal(rec.Body.Bytes(), &bundle); err != nil {
		t.Fatalf("failed to unmarshal bundle: %v", err)
	}
	if bundle.ResourceType != "Bundle" {
		t.Errorf("expected resourceType Bundle, got %q", bundle.ResourceType)
	}
	if bundle.Type != "searchset" {
		t.Errorf("expected type searchset, got %q", bundle.Type)
	}
	if bundle.Total == nil || *bundle.Total != 5 {
		t.Errorf("expected total 5, got %v", bundle.Total)
	}
	if len(bundle.Entry) != 5 {
		t.Errorf("expected 5 entries, got %d", len(bundle.Entry))
	}
}

func TestCompartmentDefinitionHandler_SearchDefinitions_FilterByCode(t *testing.T) {
	h := NewCompartmentDefinitionHandler()
	e := echo.New()

	req := httptest.NewRequest(http.MethodGet, "/fhir/CompartmentDefinition?code=Patient", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.SearchDefinitions(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var bundle Bundle
	json.Unmarshal(rec.Body.Bytes(), &bundle)
	if bundle.Total == nil || *bundle.Total != 1 {
		t.Errorf("expected total 1 for code=Patient filter, got %v", bundle.Total)
	}
	if len(bundle.Entry) != 1 {
		t.Errorf("expected 1 entry, got %d", len(bundle.Entry))
	}

	// Verify the returned entry is the Patient compartment
	var def FHIRCompartmentDefinition
	json.Unmarshal(bundle.Entry[0].Resource, &def)
	if def.Code != "Patient" {
		t.Errorf("expected Patient compartment, got %q", def.Code)
	}
}

func TestCompartmentDefinitionHandler_SearchDefinitions_FilterByCodeCaseInsensitive(t *testing.T) {
	h := NewCompartmentDefinitionHandler()
	e := echo.New()

	req := httptest.NewRequest(http.MethodGet, "/fhir/CompartmentDefinition?code=patient", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.SearchDefinitions(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var bundle Bundle
	json.Unmarshal(rec.Body.Bytes(), &bundle)
	if bundle.Total == nil || *bundle.Total != 1 {
		t.Errorf("expected total 1 for code=patient (case-insensitive), got %v", bundle.Total)
	}
}

func TestCompartmentDefinitionHandler_SearchDefinitions_FilterByURL(t *testing.T) {
	h := NewCompartmentDefinitionHandler()
	e := echo.New()

	req := httptest.NewRequest(http.MethodGet, "/fhir/CompartmentDefinition?url=http://hl7.org/fhir/CompartmentDefinition/device", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.SearchDefinitions(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var bundle Bundle
	json.Unmarshal(rec.Body.Bytes(), &bundle)
	if bundle.Total == nil || *bundle.Total != 1 {
		t.Errorf("expected total 1 for url filter, got %v", bundle.Total)
	}
}

func TestCompartmentDefinitionHandler_SearchDefinitions_FilterByName(t *testing.T) {
	h := NewCompartmentDefinitionHandler()
	e := echo.New()

	req := httptest.NewRequest(http.MethodGet, "/fhir/CompartmentDefinition?name=Practitioner", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.SearchDefinitions(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var bundle Bundle
	json.Unmarshal(rec.Body.Bytes(), &bundle)
	if bundle.Total == nil || *bundle.Total != 1 {
		t.Errorf("expected total 1 for name=Practitioner, got %v", bundle.Total)
	}
}

func TestCompartmentDefinitionHandler_SearchDefinitions_NoMatch(t *testing.T) {
	h := NewCompartmentDefinitionHandler()
	e := echo.New()

	req := httptest.NewRequest(http.MethodGet, "/fhir/CompartmentDefinition?code=NonExistent", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.SearchDefinitions(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 for empty search, got %d", rec.Code)
	}

	var bundle Bundle
	json.Unmarshal(rec.Body.Bytes(), &bundle)
	if bundle.Total == nil || *bundle.Total != 0 {
		t.Errorf("expected total 0 for no-match filter, got %v", bundle.Total)
	}
	if len(bundle.Entry) != 0 {
		t.Errorf("expected 0 entries, got %d", len(bundle.Entry))
	}
}

func TestCompartmentDefinitionHandler_SearchDefinitions_BundleEntryFullURL(t *testing.T) {
	h := NewCompartmentDefinitionHandler()
	e := echo.New()

	req := httptest.NewRequest(http.MethodGet, "/fhir/CompartmentDefinition?code=Encounter", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h.SearchDefinitions(c)

	var bundle Bundle
	json.Unmarshal(rec.Body.Bytes(), &bundle)
	if len(bundle.Entry) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(bundle.Entry))
	}
	if bundle.Entry[0].FullURL != "CompartmentDefinition/encounter" {
		t.Errorf("expected fullUrl CompartmentDefinition/encounter, got %q", bundle.Entry[0].FullURL)
	}
	if bundle.Entry[0].Search == nil || bundle.Entry[0].Search.Mode != "match" {
		t.Error("expected search mode 'match' on entry")
	}
}

func TestCompartmentDefinitionHandler_RegisterRoutes(t *testing.T) {
	h := NewCompartmentDefinitionHandler()
	e := echo.New()
	fhirGroup := e.Group("/fhir")

	h.RegisterRoutes(fhirGroup)

	routes := e.Routes()
	expectedRoutes := map[string]bool{
		"GET /fhir/CompartmentDefinition/:id": false,
		"GET /fhir/CompartmentDefinition":     false,
	}

	for _, r := range routes {
		key := r.Method + " " + r.Path
		if _, ok := expectedRoutes[key]; ok {
			expectedRoutes[key] = true
		}
	}

	for route, found := range expectedRoutes {
		if !found {
			t.Errorf("expected route %s to be registered", route)
		}
	}
}

// =========== allCompartmentDefinitions tests ===========

func TestAllCompartmentDefinitions_ReturnsFive(t *testing.T) {
	defs := allCompartmentDefinitions()
	if len(defs) != 5 {
		t.Errorf("expected 5 compartment definitions, got %d", len(defs))
	}

	expectedIDs := []string{"patient", "encounter", "practitioner", "relatedperson", "device"}
	for _, id := range expectedIDs {
		if _, ok := defs[id]; !ok {
			t.Errorf("expected compartment definition with ID %q", id)
		}
	}
}

func TestAllCompartmentDefinitions_UniqueURLs(t *testing.T) {
	defs := allCompartmentDefinitions()
	urls := make(map[string]string)
	for id, def := range defs {
		if existing, ok := urls[def.URL]; ok {
			t.Errorf("duplicate URL %q found in compartments %q and %q", def.URL, existing, id)
		}
		urls[def.URL] = id
	}
}

func TestAllCompartmentDefinitions_UniqueCodes(t *testing.T) {
	defs := allCompartmentDefinitions()
	codes := make(map[string]string)
	for id, def := range defs {
		if existing, ok := codes[def.Code]; ok {
			t.Errorf("duplicate Code %q found in compartments %q and %q", def.Code, existing, id)
		}
		codes[def.Code] = id
	}
}

// =========== NewCompartmentDefinitionHandler tests ===========

func TestNewCompartmentDefinitionHandler_PreloadsDefinitions(t *testing.T) {
	h := NewCompartmentDefinitionHandler()
	if len(h.definitions) != 5 {
		t.Errorf("expected 5 pre-loaded definitions, got %d", len(h.definitions))
	}
}

// =========== Edge case tests ===========

func TestCompartmentResourceParams_NilDefinition(t *testing.T) {
	def := &FHIRCompartmentDefinition{
		ResourceType: "CompartmentDefinition",
		Code:         "Test",
		Resource:     nil,
	}
	params := CompartmentResourceParams(def, "Observation")
	if params != nil {
		t.Errorf("expected nil params for empty definition, got %v", params)
	}
}

func TestCompartmentResourceParams_EmptyResourceList(t *testing.T) {
	def := &FHIRCompartmentDefinition{
		ResourceType: "CompartmentDefinition",
		Code:         "Test",
		Resource:     []CompartmentResource{},
	}
	params := CompartmentResourceParams(def, "Observation")
	if params != nil {
		t.Errorf("expected nil params for empty resource list, got %v", params)
	}
}

func TestPatientCompartmentDef_NoDuplicateResourceTypes(t *testing.T) {
	def := PatientCompartmentDef()
	seen := make(map[string]bool)
	for _, r := range def.Resource {
		if seen[r.Code] {
			t.Errorf("duplicate resource type %q in Patient compartment definition", r.Code)
		}
		seen[r.Code] = true
	}
}

func TestEncounterCompartmentDef_NoDuplicateResourceTypes(t *testing.T) {
	def := EncounterCompartmentDef()
	seen := make(map[string]bool)
	for _, r := range def.Resource {
		if seen[r.Code] {
			t.Errorf("duplicate resource type %q in Encounter compartment definition", r.Code)
		}
		seen[r.Code] = true
	}
}

func TestAllCompartmentDefs_ResourcesHaveNonEmptyParams(t *testing.T) {
	defs := []*FHIRCompartmentDefinition{
		PatientCompartmentDef(),
		EncounterCompartmentDef(),
		PractitionerCompartmentDef(),
		RelatedPersonCompartmentDef(),
		DeviceCompartmentDef(),
	}

	for _, def := range defs {
		t.Run(def.Code, func(t *testing.T) {
			for _, r := range def.Resource {
				if len(r.Param) == 0 {
					t.Errorf("resource %s in %s compartment has empty params", r.Code, def.Code)
				}
				if r.Code == "" {
					t.Errorf("found empty resource code in %s compartment", def.Code)
				}
				for _, p := range r.Param {
					if p == "" {
						t.Errorf("empty param string for resource %s in %s compartment", r.Code, def.Code)
					}
				}
			}
		})
	}
}
