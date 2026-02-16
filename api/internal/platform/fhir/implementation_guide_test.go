package fhir

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/labstack/echo/v4"
)

// ============================================================================
// DefaultImplementationGuide tests
// ============================================================================

func TestDefaultImplementationGuide_ResourceType(t *testing.T) {
	ig := DefaultImplementationGuide()
	if ig.ResourceType != "ImplementationGuide" {
		t.Errorf("expected resourceType ImplementationGuide, got %s", ig.ResourceType)
	}
}

func TestDefaultImplementationGuide_HasID(t *testing.T) {
	ig := DefaultImplementationGuide()
	if ig.ID == "" {
		t.Error("expected non-empty ID")
	}
}

func TestDefaultImplementationGuide_HasURL(t *testing.T) {
	ig := DefaultImplementationGuide()
	if ig.URL == "" {
		t.Error("expected non-empty URL")
	}
}

func TestDefaultImplementationGuide_StatusActive(t *testing.T) {
	ig := DefaultImplementationGuide()
	if ig.Status != "active" {
		t.Errorf("expected status active, got %s", ig.Status)
	}
}

func TestDefaultImplementationGuide_FHIRVersion(t *testing.T) {
	ig := DefaultImplementationGuide()
	if len(ig.FHIRVersion) == 0 {
		t.Fatal("expected at least one FHIR version")
	}
	if ig.FHIRVersion[0] != "4.0.1" {
		t.Errorf("expected FHIR version 4.0.1, got %s", ig.FHIRVersion[0])
	}
}

func TestDefaultImplementationGuide_HasName(t *testing.T) {
	ig := DefaultImplementationGuide()
	if ig.Name == "" {
		t.Error("expected non-empty Name")
	}
}

func TestDefaultImplementationGuide_HasTitle(t *testing.T) {
	ig := DefaultImplementationGuide()
	if ig.Title == "" {
		t.Error("expected non-empty Title")
	}
}

func TestDefaultImplementationGuide_HasDescription(t *testing.T) {
	ig := DefaultImplementationGuide()
	if ig.Description == "" {
		t.Error("expected non-empty Description")
	}
}

func TestDefaultImplementationGuide_HasPackageID(t *testing.T) {
	ig := DefaultImplementationGuide()
	if ig.PackageID == "" {
		t.Error("expected non-empty PackageID")
	}
}

func TestDefaultImplementationGuide_DependsOnUSCore(t *testing.T) {
	ig := DefaultImplementationGuide()
	if len(ig.DependsOn) < 2 {
		t.Fatalf("expected at least 2 dependencies, got %d", len(ig.DependsOn))
	}

	foundUSCore := false
	foundR4 := false
	for _, dep := range ig.DependsOn {
		if dep.Version == "6.1.0" {
			foundUSCore = true
		}
		if dep.Version == "4.0.1" {
			foundR4 = true
		}
	}
	if !foundUSCore {
		t.Error("expected US Core 6.1.0 dependency")
	}
	if !foundR4 {
		t.Error("expected FHIR R4 4.0.1 dependency")
	}
}

func TestDefaultImplementationGuide_HasGlobalProfiles(t *testing.T) {
	ig := DefaultImplementationGuide()
	if len(ig.Global) == 0 {
		t.Fatal("expected at least one global profile")
	}

	types := make(map[string]bool)
	for _, g := range ig.Global {
		types[g.Type] = true
		if g.Profile == "" {
			t.Errorf("expected non-empty profile for type %s", g.Type)
		}
	}
	if !types["Patient"] {
		t.Error("expected Patient in global profiles")
	}
}

func TestDefaultImplementationGuide_HasDefinition(t *testing.T) {
	ig := DefaultImplementationGuide()
	if ig.Definition == nil {
		t.Fatal("expected non-nil Definition")
	}
}

func TestDefaultImplementationGuide_DefinitionResources(t *testing.T) {
	ig := DefaultImplementationGuide()
	if ig.Definition == nil {
		t.Fatal("expected non-nil Definition")
	}
	if len(ig.Definition.Resource) < 3 {
		t.Errorf("expected at least 3 definition resources, got %d", len(ig.Definition.Resource))
	}

	for _, r := range ig.Definition.Resource {
		if r.Reference == nil || r.Reference["reference"] == "" {
			t.Error("expected non-empty reference in definition resource")
		}
		if r.Name == "" {
			t.Error("expected non-empty name in definition resource")
		}
	}
}

func TestDefaultImplementationGuide_DefinitionPage(t *testing.T) {
	ig := DefaultImplementationGuide()
	if ig.Definition == nil || ig.Definition.Page == nil {
		t.Fatal("expected non-nil Definition.Page")
	}

	page := ig.Definition.Page
	if page.NameURL == "" {
		t.Error("expected non-empty root page nameUrl")
	}
	if page.Title == "" {
		t.Error("expected non-empty root page title")
	}
	if page.Generation == "" {
		t.Error("expected non-empty root page generation")
	}
	if len(page.Page) == 0 {
		t.Error("expected at least one child page")
	}
}

func TestDefaultImplementationGuide_DefinitionParameters(t *testing.T) {
	ig := DefaultImplementationGuide()
	if ig.Definition == nil {
		t.Fatal("expected non-nil Definition")
	}
	if len(ig.Definition.Parameter) == 0 {
		t.Error("expected at least one parameter")
	}

	for _, p := range ig.Definition.Parameter {
		if p.Code == "" {
			t.Error("expected non-empty parameter code")
		}
		if p.Value == "" {
			t.Error("expected non-empty parameter value")
		}
	}
}

func TestDefaultImplementationGuide_JSONSerialization(t *testing.T) {
	ig := DefaultImplementationGuide()
	data, err := json.Marshal(ig)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if result["resourceType"] != "ImplementationGuide" {
		t.Errorf("expected ImplementationGuide, got %v", result["resourceType"])
	}
	if result["status"] != "active" {
		t.Errorf("expected status active, got %v", result["status"])
	}

	dependsOn, ok := result["dependsOn"].([]interface{})
	if !ok {
		t.Fatal("expected dependsOn array")
	}
	if len(dependsOn) < 2 {
		t.Errorf("expected at least 2 dependsOn entries, got %d", len(dependsOn))
	}
}

// ============================================================================
// DefaultTerminologyCapabilities tests
// ============================================================================

func TestDefaultTerminologyCapabilities_ResourceType(t *testing.T) {
	tc := DefaultTerminologyCapabilities()
	if tc.ResourceType != "TerminologyCapabilities" {
		t.Errorf("expected resourceType TerminologyCapabilities, got %s", tc.ResourceType)
	}
}

func TestDefaultTerminologyCapabilities_HasID(t *testing.T) {
	tc := DefaultTerminologyCapabilities()
	if tc.ID == "" {
		t.Error("expected non-empty ID")
	}
}

func TestDefaultTerminologyCapabilities_StatusActive(t *testing.T) {
	tc := DefaultTerminologyCapabilities()
	if tc.Status != "active" {
		t.Errorf("expected status active, got %s", tc.Status)
	}
}

func TestDefaultTerminologyCapabilities_KindInstance(t *testing.T) {
	tc := DefaultTerminologyCapabilities()
	if tc.Kind != "instance" {
		t.Errorf("expected kind instance, got %s", tc.Kind)
	}
}

func TestDefaultTerminologyCapabilities_HasDate(t *testing.T) {
	tc := DefaultTerminologyCapabilities()
	if tc.Date == "" {
		t.Error("expected non-empty Date")
	}
	// Date should be in YYYY-MM-DD format
	if len(tc.Date) != 10 || tc.Date[4] != '-' || tc.Date[7] != '-' {
		t.Errorf("expected date in YYYY-MM-DD format, got %q", tc.Date)
	}
}

func TestDefaultTerminologyCapabilities_HasDescription(t *testing.T) {
	tc := DefaultTerminologyCapabilities()
	if tc.Description == "" {
		t.Error("expected non-empty Description")
	}
}

func TestDefaultTerminologyCapabilities_CodeSystems(t *testing.T) {
	tc := DefaultTerminologyCapabilities()
	if len(tc.CodeSystem) < 4 {
		t.Fatalf("expected at least 4 code systems, got %d", len(tc.CodeSystem))
	}

	uris := make(map[string]bool)
	for _, cs := range tc.CodeSystem {
		uris[cs.URI] = true
	}
	expected := []string{
		"http://snomed.info/sct",
		"http://loinc.org",
		"http://www.nlm.nih.gov/research/umls/rxnorm",
		"http://hl7.org/fhir/sid/icd-10-cm",
	}
	for _, e := range expected {
		if !uris[e] {
			t.Errorf("missing code system: %s", e)
		}
	}
}

func TestDefaultTerminologyCapabilities_Expansion(t *testing.T) {
	tc := DefaultTerminologyCapabilities()
	if tc.Expansion == nil {
		t.Fatal("expected non-nil Expansion")
	}
	if !tc.Expansion.Hierarchical {
		t.Error("expected Hierarchical true")
	}
	if !tc.Expansion.Paging {
		t.Error("expected Paging true")
	}
	if tc.Expansion.Incomplete {
		t.Error("expected Incomplete false")
	}
}

func TestDefaultTerminologyCapabilities_ValidateCode(t *testing.T) {
	tc := DefaultTerminologyCapabilities()
	if tc.ValidateCode == nil {
		t.Fatal("expected non-nil ValidateCode")
	}
	if !tc.ValidateCode.Translations {
		t.Error("expected Translations true")
	}
}

func TestDefaultTerminologyCapabilities_Translation(t *testing.T) {
	tc := DefaultTerminologyCapabilities()
	if tc.Translation == nil {
		t.Fatal("expected non-nil Translation")
	}
	if !tc.Translation.NeedsMap {
		t.Error("expected NeedsMap true")
	}
}

func TestDefaultTerminologyCapabilities_Closure(t *testing.T) {
	tc := DefaultTerminologyCapabilities()
	if tc.Closure == nil {
		t.Fatal("expected non-nil Closure")
	}
	if !tc.Closure.Translation {
		t.Error("expected Translation true")
	}
}

func TestDefaultTerminologyCapabilities_JSONSerialization(t *testing.T) {
	tc := DefaultTerminologyCapabilities()
	data, err := json.Marshal(tc)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if result["resourceType"] != "TerminologyCapabilities" {
		t.Errorf("expected TerminologyCapabilities, got %v", result["resourceType"])
	}
	if result["kind"] != "instance" {
		t.Errorf("expected kind instance, got %v", result["kind"])
	}
}

// ============================================================================
// ImplementationGuide Handler tests
// ============================================================================

func TestImplementationGuideHandler_List(t *testing.T) {
	h := NewImplementationGuideHandler()

	e := echo.New()
	h.RegisterRoutes(e.Group("/fhir"))

	req := httptest.NewRequest(http.MethodGet, "/fhir/ImplementationGuide", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var bundle map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &bundle); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if bundle["resourceType"] != "Bundle" {
		t.Errorf("expected Bundle, got %v", bundle["resourceType"])
	}
	if bundle["type"] != "searchset" {
		t.Errorf("expected searchset, got %v", bundle["type"])
	}

	total, ok := bundle["total"].(float64)
	if !ok {
		t.Fatal("expected total as number")
	}
	if total < 1 {
		t.Errorf("expected at least 1 entry, got %v", total)
	}

	entries, ok := bundle["entry"].([]interface{})
	if !ok {
		t.Fatal("expected entry array")
	}
	if len(entries) < 1 {
		t.Errorf("expected at least 1 entry, got %d", len(entries))
	}
}

func TestImplementationGuideHandler_Read(t *testing.T) {
	h := NewImplementationGuideHandler()

	e := echo.New()
	h.RegisterRoutes(e.Group("/fhir"))

	ig := DefaultImplementationGuide()
	req := httptest.NewRequest(http.MethodGet, "/fhir/ImplementationGuide/"+ig.ID, nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d; body: %s", rec.Code, rec.Body.String())
	}

	var result map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if result["resourceType"] != "ImplementationGuide" {
		t.Errorf("expected ImplementationGuide, got %v", result["resourceType"])
	}
	if result["id"] != ig.ID {
		t.Errorf("expected id %s, got %v", ig.ID, result["id"])
	}
}

func TestImplementationGuideHandler_Read_NotFound(t *testing.T) {
	h := NewImplementationGuideHandler()

	e := echo.New()
	h.RegisterRoutes(e.Group("/fhir"))

	req := httptest.NewRequest(http.MethodGet, "/fhir/ImplementationGuide/nonexistent", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestImplementationGuideHandler_AddGuide(t *testing.T) {
	h := NewImplementationGuideHandler()

	customIG := &ImplementationGuideResource{
		ResourceType: "ImplementationGuide",
		ID:           "custom-ig",
		URL:          "http://example.org/ig/custom",
		Name:         "CustomIG",
		Status:       "draft",
	}
	h.AddGuide(customIG)

	e := echo.New()
	h.RegisterRoutes(e.Group("/fhir"))

	// Verify list now returns 2
	req := httptest.NewRequest(http.MethodGet, "/fhir/ImplementationGuide", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var bundle map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &bundle); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	total := bundle["total"].(float64)
	if total != 2 {
		t.Errorf("expected 2 entries after adding custom IG, got %v", total)
	}

	// Verify read works for the custom IG
	req = httptest.NewRequest(http.MethodGet, "/fhir/ImplementationGuide/custom-ig", nil)
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if result["id"] != "custom-ig" {
		t.Errorf("expected id custom-ig, got %v", result["id"])
	}
	if result["status"] != "draft" {
		t.Errorf("expected status draft, got %v", result["status"])
	}
}

func TestImplementationGuideHandler_ConcurrentAccess(t *testing.T) {
	h := NewImplementationGuideHandler()

	e := echo.New()
	h.RegisterRoutes(e.Group("/fhir"))

	var wg sync.WaitGroup
	errs := make(chan error, 40)

	// Concurrent reads
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			req := httptest.NewRequest(http.MethodGet, "/fhir/ImplementationGuide", nil)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)
			if rec.Code != http.StatusOK {
				errs <- &testError{msg: "list returned non-200"}
			}
		}()
	}

	// Concurrent reads by ID
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ig := DefaultImplementationGuide()
			req := httptest.NewRequest(http.MethodGet, "/fhir/ImplementationGuide/"+ig.ID, nil)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)
			if rec.Code != http.StatusOK {
				errs <- &testError{msg: "read returned non-200"}
			}
		}()
	}

	wg.Wait()
	close(errs)

	for err := range errs {
		t.Errorf("concurrent error: %v", err)
	}
}

// ============================================================================
// TerminologyCapabilities Handler tests
// ============================================================================

func TestTerminologyCapabilitiesHandler_Get(t *testing.T) {
	h := NewTerminologyCapabilitiesHandler()

	e := echo.New()
	h.RegisterRoutes(e.Group("/fhir"))

	req := httptest.NewRequest(http.MethodGet, "/fhir/TerminologyCapabilities", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if result["resourceType"] != "TerminologyCapabilities" {
		t.Errorf("expected TerminologyCapabilities, got %v", result["resourceType"])
	}
	if result["kind"] != "instance" {
		t.Errorf("expected kind instance, got %v", result["kind"])
	}
	if result["status"] != "active" {
		t.Errorf("expected status active, got %v", result["status"])
	}
}

func TestTerminologyCapabilitiesHandler_Get_HasCodeSystems(t *testing.T) {
	h := NewTerminologyCapabilitiesHandler()

	e := echo.New()
	h.RegisterRoutes(e.Group("/fhir"))

	req := httptest.NewRequest(http.MethodGet, "/fhir/TerminologyCapabilities", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	codeSystems, ok := result["codeSystem"].([]interface{})
	if !ok {
		t.Fatal("expected codeSystem array")
	}
	if len(codeSystems) < 4 {
		t.Errorf("expected at least 4 code systems, got %d", len(codeSystems))
	}
}

func TestTerminologyCapabilitiesHandler_Get_HasExpansion(t *testing.T) {
	h := NewTerminologyCapabilitiesHandler()

	e := echo.New()
	h.RegisterRoutes(e.Group("/fhir"))

	req := httptest.NewRequest(http.MethodGet, "/fhir/TerminologyCapabilities", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	expansion, ok := result["expansion"].(map[string]interface{})
	if !ok {
		t.Fatal("expected expansion object")
	}
	if expansion["hierarchical"] != true {
		t.Error("expected hierarchical true")
	}
	if expansion["paging"] != true {
		t.Error("expected paging true")
	}
}

func TestTerminologyCapabilitiesHandler_Get_HasClosure(t *testing.T) {
	h := NewTerminologyCapabilitiesHandler()

	e := echo.New()
	h.RegisterRoutes(e.Group("/fhir"))

	req := httptest.NewRequest(http.MethodGet, "/fhir/TerminologyCapabilities", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	closure, ok := result["closure"].(map[string]interface{})
	if !ok {
		t.Fatal("expected closure object")
	}
	if closure["translation"] != true {
		t.Error("expected translation true")
	}
}

// ============================================================================
// Struct-level tests
// ============================================================================

func TestIGDependency_JSON(t *testing.T) {
	dep := IGDependency{
		URI:     "http://hl7.org/fhir/us/core/ImplementationGuide/hl7.fhir.us.core",
		Version: "6.1.0",
	}
	data, err := json.Marshal(dep)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if result["uri"] != dep.URI {
		t.Errorf("expected URI %s, got %v", dep.URI, result["uri"])
	}
	if result["version"] != "6.1.0" {
		t.Errorf("expected version 6.1.0, got %v", result["version"])
	}
}

func TestIGGlobal_JSON(t *testing.T) {
	global := IGGlobal{
		Type:    "Patient",
		Profile: "http://hl7.org/fhir/us/core/StructureDefinition/us-core-patient",
	}
	data, err := json.Marshal(global)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if result["type"] != "Patient" {
		t.Errorf("expected type Patient, got %v", result["type"])
	}
	if result["profile"] != global.Profile {
		t.Errorf("expected profile %s, got %v", global.Profile, result["profile"])
	}
}

func TestIGResource_JSON(t *testing.T) {
	res := IGResource{
		Reference:        map[string]string{"reference": "StructureDefinition/us-core-patient"},
		Name:             "US Core Patient",
		Description:      "Patient profile",
		ExampleCanonical: "http://example.org/StructureDefinition/example",
	}
	data, err := json.Marshal(res)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	ref, ok := result["reference"].(map[string]interface{})
	if !ok {
		t.Fatal("expected reference object")
	}
	if ref["reference"] != "StructureDefinition/us-core-patient" {
		t.Errorf("unexpected reference: %v", ref["reference"])
	}
	if result["exampleCanonical"] != res.ExampleCanonical {
		t.Errorf("expected exampleCanonical %s, got %v", res.ExampleCanonical, result["exampleCanonical"])
	}
}

func TestIGPage_RecursiveNesting(t *testing.T) {
	page := IGPage{
		NameURL:    "index.html",
		Title:      "Home",
		Generation: "html",
		Page: []IGPage{
			{
				NameURL:    "profiles.html",
				Title:      "Profiles",
				Generation: "markdown",
				Page: []IGPage{
					{
						NameURL:    "patient.html",
						Title:      "Patient Profile",
						Generation: "generated",
					},
				},
			},
		},
	}
	data, err := json.Marshal(page)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	pages, ok := result["page"].([]interface{})
	if !ok {
		t.Fatal("expected page array")
	}
	if len(pages) != 1 {
		t.Fatalf("expected 1 child page, got %d", len(pages))
	}

	child := pages[0].(map[string]interface{})
	if child["title"] != "Profiles" {
		t.Errorf("expected child title Profiles, got %v", child["title"])
	}

	grandchildren, ok := child["page"].([]interface{})
	if !ok {
		t.Fatal("expected nested page array")
	}
	if len(grandchildren) != 1 {
		t.Fatalf("expected 1 grandchild page, got %d", len(grandchildren))
	}
}

func TestIGParameter_JSON(t *testing.T) {
	param := IGParameter{
		Code:  "releaselabel",
		Value: "CI Build",
	}
	data, err := json.Marshal(param)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if result["code"] != "releaselabel" {
		t.Errorf("expected code releaselabel, got %v", result["code"])
	}
	if result["value"] != "CI Build" {
		t.Errorf("expected value 'CI Build', got %v", result["value"])
	}
}

func TestTCCodeSystem_JSON(t *testing.T) {
	cs := TCCodeSystem{
		URI: "http://snomed.info/sct",
	}
	data, err := json.Marshal(cs)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if result["uri"] != "http://snomed.info/sct" {
		t.Errorf("expected SNOMED URI, got %v", result["uri"])
	}
}

func TestTCExpansion_JSON(t *testing.T) {
	exp := TCExpansion{
		Hierarchical: true,
		Paging:       false,
		Incomplete:   true,
	}
	data, err := json.Marshal(exp)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if result["hierarchical"] != true {
		t.Error("expected hierarchical true")
	}
	if result["paging"] != false {
		t.Error("expected paging false")
	}
	if result["incomplete"] != true {
		t.Error("expected incomplete true")
	}
}

// ============================================================================
// testError helper
// ============================================================================

type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}
