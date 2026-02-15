package fhir

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/labstack/echo/v4"
)

// ===========================================================================
// Legacy-compatible builder tests (preserved from original)
// ===========================================================================

func TestCapabilityBuilder_AddResource(t *testing.T) {
	b := NewCapabilityBuilder("http://localhost:8000/fhir", "0.1.0")

	b.AddResource("Patient", DefaultInteractions(), []SearchParam{
		{Name: "name", Type: "string"},
		{Name: "family", Type: "string"},
		{Name: "birthdate", Type: "date"},
	})

	if b.ResourceCount() != 1 {
		t.Fatalf("expected 1 resource, got %d", b.ResourceCount())
	}

	cs := b.Build()
	if cs["resourceType"] != "CapabilityStatement" {
		t.Errorf("expected CapabilityStatement, got %v", cs["resourceType"])
	}
	if cs["fhirVersion"] != "4.0.1" {
		t.Errorf("expected fhirVersion 4.0.1, got %v", cs["fhirVersion"])
	}
	if cs["kind"] != "instance" {
		t.Errorf("expected kind instance, got %v", cs["kind"])
	}
	if cs["status"] != "active" {
		t.Errorf("expected status active, got %v", cs["status"])
	}

	// Check format
	formats := cs["format"].([]string)
	if len(formats) != 1 || formats[0] != "json" {
		t.Errorf("expected format [json], got %v", formats)
	}

	// Check software version
	software := cs["software"].(map[string]string)
	if software["version"] != "0.1.0" {
		t.Errorf("expected version 0.1.0, got %s", software["version"])
	}
}

func TestCapabilityBuilder_Build_Resources(t *testing.T) {
	b := NewCapabilityBuilder("http://localhost:8000/fhir", "0.1.0")

	b.AddResource("Patient", DefaultInteractions(), []SearchParam{
		{Name: "name", Type: "string"},
	})
	b.AddResource("Observation", DefaultInteractions(), []SearchParam{
		{Name: "patient", Type: "reference"},
		{Name: "code", Type: "token"},
	})
	b.AddResource("Encounter", []string{"read", "search-type"}, []SearchParam{
		{Name: "patient", Type: "reference"},
	})

	cs := b.Build()
	rest := cs["rest"].([]map[string]interface{})
	if len(rest) != 1 {
		t.Fatalf("expected 1 rest entry, got %d", len(rest))
	}

	resources := rest[0]["resource"].([]map[string]interface{})
	if len(resources) != 3 {
		t.Fatalf("expected 3 resources, got %d", len(resources))
	}

	// Resources should be sorted alphabetically
	if resources[0]["type"] != "Encounter" {
		t.Errorf("expected first resource Encounter, got %v", resources[0]["type"])
	}
	if resources[1]["type"] != "Observation" {
		t.Errorf("expected second resource Observation, got %v", resources[1]["type"])
	}
	if resources[2]["type"] != "Patient" {
		t.Errorf("expected third resource Patient, got %v", resources[2]["type"])
	}

	// Check Encounter has 2 interactions (read, search-type)
	encInteractions := resources[0]["interaction"].([]map[string]string)
	if len(encInteractions) != 2 {
		t.Errorf("expected 2 Encounter interactions, got %d", len(encInteractions))
	}

	// Check Patient has 9 interactions (default)
	patInteractions := resources[2]["interaction"].([]map[string]string)
	if len(patInteractions) != 9 {
		t.Errorf("expected 9 Patient interactions, got %d", len(patInteractions))
	}
}

func TestCapabilityBuilder_MergeResources(t *testing.T) {
	b := NewCapabilityBuilder("http://localhost:8000/fhir", "0.1.0")

	// First registration
	b.AddResource("Patient", []string{"read", "search-type"}, []SearchParam{
		{Name: "name", Type: "string"},
	})

	// Second registration adds more interactions and search params
	b.AddResource("Patient", []string{"read", "create", "update"}, []SearchParam{
		{Name: "name", Type: "string"},
		{Name: "birthdate", Type: "date"},
	})

	if b.ResourceCount() != 1 {
		t.Fatalf("expected 1 resource after merge, got %d", b.ResourceCount())
	}

	cs := b.Build()
	rest := cs["rest"].([]map[string]interface{})
	resources := rest[0]["resource"].([]map[string]interface{})

	// Should have deduplicated interactions: read, search-type, create, update
	interactions := resources[0]["interaction"].([]map[string]string)
	if len(interactions) != 4 {
		t.Errorf("expected 4 merged interactions, got %d", len(interactions))
	}

	// Should have deduplicated search params: name, birthdate
	params := resources[0]["searchParam"].([]map[string]string)
	if len(params) != 2 {
		t.Errorf("expected 2 merged search params, got %d", len(params))
	}
}

func TestCapabilityBuilder_SecuritySection(t *testing.T) {
	b := NewCapabilityBuilder("http://localhost:8000/fhir", "0.1.0")
	b.SetOAuthURIs(
		"http://keycloak:8080/realms/ehr/protocol/openid-connect/auth",
		"http://keycloak:8080/realms/ehr/protocol/openid-connect/token",
	)

	b.AddResource("Patient", DefaultInteractions(), nil)

	cs := b.Build()
	rest := cs["rest"].([]map[string]interface{})
	security := rest[0]["security"].(map[string]interface{})

	// Check CORS
	if security["cors"] != true {
		t.Error("expected cors to be true")
	}

	// Check service coding
	services := security["service"].([]map[string]interface{})
	if len(services) != 1 {
		t.Fatalf("expected 1 service, got %d", len(services))
	}
	codings := services[0]["coding"].([]map[string]string)
	if codings[0]["code"] != "SMART-on-FHIR" {
		t.Errorf("expected SMART-on-FHIR, got %s", codings[0]["code"])
	}
	if codings[0]["system"] != "http://hl7.org/fhir/restful-security-service" {
		t.Errorf("unexpected system: %s", codings[0]["system"])
	}

	// Check OAuth extension
	extensions := security["extension"].([]map[string]interface{})
	if len(extensions) != 1 {
		t.Fatalf("expected 1 extension, got %d", len(extensions))
	}
	if extensions[0]["url"] != "http://fhir-registry.smarthealthit.org/StructureDefinition/oauth-uris" {
		t.Errorf("unexpected extension URL: %s", extensions[0]["url"])
	}

	oauthExts := extensions[0]["extension"].([]map[string]string)
	if len(oauthExts) != 2 {
		t.Fatalf("expected 2 OAuth extension entries, got %d", len(oauthExts))
	}

	if oauthExts[0]["url"] != "authorize" {
		t.Errorf("expected authorize, got %s", oauthExts[0]["url"])
	}
	if oauthExts[0]["valueUri"] != "http://keycloak:8080/realms/ehr/protocol/openid-connect/auth" {
		t.Errorf("unexpected authorize URI: %s", oauthExts[0]["valueUri"])
	}
	if oauthExts[1]["url"] != "token" {
		t.Errorf("expected token, got %s", oauthExts[1]["url"])
	}
	if oauthExts[1]["valueUri"] != "http://keycloak:8080/realms/ehr/protocol/openid-connect/token" {
		t.Errorf("unexpected token URI: %s", oauthExts[1]["valueUri"])
	}
}

func TestCapabilityBuilder_NoOAuthURIs(t *testing.T) {
	b := NewCapabilityBuilder("http://localhost:8000/fhir", "0.1.0")
	// Do NOT set OAuth URIs
	b.AddResource("Patient", DefaultInteractions(), nil)

	cs := b.Build()
	rest := cs["rest"].([]map[string]interface{})
	security := rest[0]["security"].(map[string]interface{})

	// Should still have service but no OAuth extension
	if _, hasExt := security["extension"]; hasExt {
		t.Error("expected no extension when OAuth URIs are not set")
	}
}

func TestCapabilityBuilder_WithProfile(t *testing.T) {
	b := NewCapabilityBuilder("http://localhost:8000/fhir", "0.1.0")
	b.AddResourceWithProfile("Patient", DefaultInteractions(), []SearchParam{
		{Name: "name", Type: "string"},
	}, []string{
		"http://hl7.org/fhir/us/core/StructureDefinition/us-core-patient",
	})

	cs := b.Build()
	rest := cs["rest"].([]map[string]interface{})
	resources := rest[0]["resource"].([]map[string]interface{})

	profiles, ok := resources[0]["supportedProfile"].([]string)
	if !ok {
		t.Fatal("expected supportedProfile to be set")
	}
	if len(profiles) != 1 {
		t.Fatalf("expected 1 profile, got %d", len(profiles))
	}
	if profiles[0] != "http://hl7.org/fhir/us/core/StructureDefinition/us-core-patient" {
		t.Errorf("unexpected profile: %s", profiles[0])
	}
}

func TestCapabilityBuilder_JSONSerialization(t *testing.T) {
	b := NewCapabilityBuilder("http://localhost:8000/fhir", "0.1.0")
	b.SetOAuthURIs(
		"http://keycloak:8080/auth",
		"http://keycloak:8080/token",
	)
	b.AddResource("Patient", DefaultInteractions(), []SearchParam{
		{Name: "name", Type: "string"},
	})

	cs := b.Build()

	data, err := json.Marshal(cs)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if result["resourceType"] != "CapabilityStatement" {
		t.Errorf("expected CapabilityStatement, got %v", result["resourceType"])
	}
}

func TestCapabilityBuilder_EmptyBuild(t *testing.T) {
	b := NewCapabilityBuilder("http://localhost:8000/fhir", "0.1.0")

	cs := b.Build()

	rest := cs["rest"].([]map[string]interface{})
	resources := rest[0]["resource"].([]map[string]interface{})
	if len(resources) != 0 {
		t.Errorf("expected 0 resources, got %d", len(resources))
	}
}

func TestCapabilityBuilder_SearchParamDocumentation(t *testing.T) {
	b := NewCapabilityBuilder("http://localhost:8000/fhir", "0.1.0")
	b.AddResource("Patient", DefaultInteractions(), []SearchParam{
		{Name: "name", Type: "string", Documentation: "A server defined search for the patient name"},
	})

	cs := b.Build()
	rest := cs["rest"].([]map[string]interface{})
	resources := rest[0]["resource"].([]map[string]interface{})
	params := resources[0]["searchParam"].([]map[string]string)

	if params[0]["documentation"] != "A server defined search for the patient name" {
		t.Errorf("expected documentation, got %s", params[0]["documentation"])
	}
}

func TestDefaultInteractions(t *testing.T) {
	interactions := DefaultInteractions()
	if len(interactions) != 9 {
		t.Fatalf("expected 9 interactions, got %d", len(interactions))
	}

	expected := map[string]bool{
		"read": true, "vread": true, "search-type": true,
		"create": true, "update": true, "delete": true,
		"patch": true, "history-instance": true, "history-type": true,
	}
	for _, i := range interactions {
		if !expected[i] {
			t.Errorf("unexpected interaction: %s", i)
		}
	}
}

func TestReadOnlyInteractions(t *testing.T) {
	interactions := ReadOnlyInteractions()
	if len(interactions) != 3 {
		t.Fatalf("expected 3 interactions, got %d", len(interactions))
	}

	expected := map[string]bool{
		"read": true, "vread": true, "search-type": true,
	}
	for _, i := range interactions {
		if !expected[i] {
			t.Errorf("unexpected interaction: %s", i)
		}
	}
}

func TestCapabilityBuilder_ConcurrentAccess(t *testing.T) {
	b := NewCapabilityBuilder("http://localhost:8000/fhir", "0.1.0")

	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(idx int) {
			resources := []string{"Patient", "Observation", "Encounter", "Condition", "Procedure"}
			rt := resources[idx%len(resources)]
			b.AddResource(rt, DefaultInteractions(), []SearchParam{
				{Name: "test", Type: "string"},
			})
			_ = b.Build()
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	if b.ResourceCount() > 5 {
		t.Errorf("expected at most 5 resources, got %d", b.ResourceCount())
	}
}

func TestDefaultCapabilityOptions(t *testing.T) {
	opts := DefaultCapabilityOptions()
	if !opts.ConditionalCreate {
		t.Error("expected ConditionalCreate true")
	}
	if !opts.ConditionalUpdate {
		t.Error("expected ConditionalUpdate true")
	}
	if opts.ConditionalDelete != "single" {
		t.Errorf("expected ConditionalDelete 'single', got %q", opts.ConditionalDelete)
	}
	if !opts.ReadHistory {
		t.Error("expected ReadHistory true")
	}
	if opts.UpdateCreate {
		t.Error("expected UpdateCreate false")
	}
	if len(opts.PatchFormats) != 2 {
		t.Fatalf("expected 2 PatchFormats, got %d", len(opts.PatchFormats))
	}
	if opts.PatchFormats[0] != "application/json-patch+json" {
		t.Errorf("expected first patch format 'application/json-patch+json', got %q", opts.PatchFormats[0])
	}
	if opts.PatchFormats[1] != "application/merge-patch+json" {
		t.Errorf("expected second patch format 'application/merge-patch+json', got %q", opts.PatchFormats[1])
	}
	if opts.SearchInclude != nil {
		t.Errorf("expected nil SearchInclude, got %v", opts.SearchInclude)
	}
	if opts.SearchRevInclude != nil {
		t.Errorf("expected nil SearchRevInclude, got %v", opts.SearchRevInclude)
	}
}

func TestSetResourceCapabilities_WithDefaults(t *testing.T) {
	b := NewCapabilityBuilder("http://localhost:8000/fhir", "0.1.0")
	b.AddResource("Patient", DefaultInteractions(), []SearchParam{
		{Name: "name", Type: "string"},
	})

	opts := DefaultCapabilityOptions()
	opts.SearchInclude = []string{"Patient:organization"}
	opts.SearchRevInclude = []string{"Observation:patient"}
	b.SetResourceCapabilities("Patient", opts)

	cs := b.Build()
	rest := cs["rest"].([]map[string]interface{})
	resources := rest[0]["resource"].([]map[string]interface{})
	res := resources[0]

	// conditionalCreate
	if res["conditionalCreate"] != true {
		t.Error("expected conditionalCreate true")
	}
	// conditionalUpdate
	if res["conditionalUpdate"] != true {
		t.Error("expected conditionalUpdate true")
	}
	// conditionalDelete
	if res["conditionalDelete"] != "single" {
		t.Errorf("expected conditionalDelete 'single', got %v", res["conditionalDelete"])
	}
	// readHistory
	if res["readHistory"] != true {
		t.Error("expected readHistory true")
	}
	// updateCreate
	if res["updateCreate"] != false {
		t.Error("expected updateCreate false")
	}
	// patchFormats
	pf, ok := res["patchFormats"].([]string)
	if !ok {
		t.Fatal("expected patchFormats to be set")
	}
	if len(pf) != 2 {
		t.Fatalf("expected 2 patchFormats, got %d", len(pf))
	}
	if pf[0] != "application/json-patch+json" {
		t.Errorf("unexpected first patchFormat: %s", pf[0])
	}
	// searchInclude
	si, ok := res["searchInclude"].([]string)
	if !ok {
		t.Fatal("expected searchInclude to be set")
	}
	if len(si) != 1 || si[0] != "Patient:organization" {
		t.Errorf("unexpected searchInclude: %v", si)
	}
	// searchRevInclude
	sri, ok := res["searchRevInclude"].([]string)
	if !ok {
		t.Fatal("expected searchRevInclude to be set")
	}
	if len(sri) != 1 || sri[0] != "Observation:patient" {
		t.Errorf("unexpected searchRevInclude: %v", sri)
	}
}

func TestSetResourceCapabilities_NonExistentResource(t *testing.T) {
	b := NewCapabilityBuilder("http://localhost:8000/fhir", "0.1.0")
	// Do NOT register "Ghost" resource first
	b.SetResourceCapabilities("Ghost", DefaultCapabilityOptions())

	// Should be no-op; builder has no resources
	if b.ResourceCount() != 0 {
		t.Errorf("expected 0 resources, got %d", b.ResourceCount())
	}
}

func TestBuild_ConditionalDeleteOnly(t *testing.T) {
	b := NewCapabilityBuilder("http://localhost:8000/fhir", "0.1.0")
	b.AddResource("Observation", []string{"read"}, nil)
	b.SetResourceCapabilities("Observation", ResourceCapabilityOptions{
		ConditionalDelete: "multiple",
	})

	cs := b.Build()
	rest := cs["rest"].([]map[string]interface{})
	resources := rest[0]["resource"].([]map[string]interface{})
	res := resources[0]

	if res["conditionalDelete"] != "multiple" {
		t.Errorf("expected conditionalDelete 'multiple', got %v", res["conditionalDelete"])
	}
	// conditionalCreate and conditionalUpdate should NOT appear when false
	if _, ok := res["conditionalCreate"]; ok {
		t.Error("conditionalCreate should not be in output when false")
	}
	if _, ok := res["conditionalUpdate"]; ok {
		t.Error("conditionalUpdate should not be in output when false")
	}
}

func TestBuild_PatchFormatsOnly(t *testing.T) {
	b := NewCapabilityBuilder("http://localhost:8000/fhir", "0.1.0")
	b.AddResource("Condition", []string{"read", "patch"}, nil)
	b.SetResourceCapabilities("Condition", ResourceCapabilityOptions{
		PatchFormats: []string{"application/json-patch+json"},
	})

	cs := b.Build()
	rest := cs["rest"].([]map[string]interface{})
	resources := rest[0]["resource"].([]map[string]interface{})
	res := resources[0]

	pf, ok := res["patchFormats"].([]string)
	if !ok {
		t.Fatal("expected patchFormats to be present")
	}
	if len(pf) != 1 || pf[0] != "application/json-patch+json" {
		t.Errorf("unexpected patchFormats: %v", pf)
	}
}

func TestCapabilityBuilder_SetOAuthURIs_OnlyAuthorize(t *testing.T) {
	b := NewCapabilityBuilder("http://localhost:8000/fhir", "0.1.0")
	b.SetOAuthURIs("http://auth.example.com/auth", "")
	b.AddResource("Patient", []string{"read"}, nil)

	cs := b.Build()
	rest := cs["rest"].([]map[string]interface{})
	security := rest[0]["security"].(map[string]interface{})

	extensions := security["extension"].([]map[string]interface{})
	if len(extensions) != 1 {
		t.Fatalf("expected 1 extension, got %d", len(extensions))
	}
	oauthExts := extensions[0]["extension"].([]map[string]string)
	if len(oauthExts) != 1 {
		t.Fatalf("expected 1 OAuth extension entry (authorize only), got %d", len(oauthExts))
	}
	if oauthExts[0]["url"] != "authorize" {
		t.Errorf("expected authorize, got %s", oauthExts[0]["url"])
	}
}

func TestCapabilityBuilder_SetOAuthURIs_OnlyToken(t *testing.T) {
	b := NewCapabilityBuilder("http://localhost:8000/fhir", "0.1.0")
	b.SetOAuthURIs("", "http://auth.example.com/token")
	b.AddResource("Patient", []string{"read"}, nil)

	cs := b.Build()
	rest := cs["rest"].([]map[string]interface{})
	security := rest[0]["security"].(map[string]interface{})

	extensions := security["extension"].([]map[string]interface{})
	if len(extensions) != 1 {
		t.Fatalf("expected 1 extension, got %d", len(extensions))
	}
	oauthExts := extensions[0]["extension"].([]map[string]string)
	if len(oauthExts) != 1 {
		t.Fatalf("expected 1 OAuth extension entry (token only), got %d", len(oauthExts))
	}
	if oauthExts[0]["url"] != "token" {
		t.Errorf("expected token, got %s", oauthExts[0]["url"])
	}
}

func TestCapabilityBuilder_AddResourceWithProfile_DuplicateProfiles(t *testing.T) {
	b := NewCapabilityBuilder("http://localhost:8000/fhir", "0.1.0")

	profile := "http://hl7.org/fhir/us/core/StructureDefinition/us-core-patient"
	b.AddResourceWithProfile("Patient", DefaultInteractions(), nil, []string{profile})
	b.AddResourceWithProfile("Patient", nil, nil, []string{profile, "http://example.com/profile"})

	cs := b.Build()
	rest := cs["rest"].([]map[string]interface{})
	resources := rest[0]["resource"].([]map[string]interface{})
	profiles := resources[0]["supportedProfile"].([]string)

	if len(profiles) != 2 {
		t.Fatalf("expected 2 profiles (deduplicated), got %d: %v", len(profiles), profiles)
	}
}

func TestCapabilityBuilder_Build_ImplementationSection(t *testing.T) {
	b := NewCapabilityBuilder("http://example.com/fhir", "2.0.0")
	b.AddResource("Patient", []string{"read"}, nil)

	cs := b.Build()
	impl := cs["implementation"].(map[string]string)
	if impl["description"] != "Headless EHR FHIR R4 Server" {
		t.Errorf("unexpected description: %s", impl["description"])
	}
	if impl["url"] != "http://example.com/fhir" {
		t.Errorf("unexpected url: %s", impl["url"])
	}
}

func TestCapabilityBuilder_Build_DateFormat(t *testing.T) {
	b := NewCapabilityBuilder("http://localhost:8000/fhir", "0.1.0")
	cs := b.Build()
	date := cs["date"].(string)
	// Date should be in YYYY-MM-DD format
	if len(date) != 10 || date[4] != '-' || date[7] != '-' {
		t.Errorf("date should be in YYYY-MM-DD format, got %q", date)
	}
}

func TestCapabilityBuilder_AddResource_NoSearchParams(t *testing.T) {
	b := NewCapabilityBuilder("http://localhost:8000/fhir", "0.1.0")
	b.AddResource("Patient", []string{"read"}, nil)

	cs := b.Build()
	rest := cs["rest"].([]map[string]interface{})
	resources := rest[0]["resource"].([]map[string]interface{})
	if _, ok := resources[0]["searchParam"]; ok {
		t.Error("searchParam should not be present when no search params are registered")
	}
}

func TestCapabilityBuilder_AddResource_NoInteractions(t *testing.T) {
	b := NewCapabilityBuilder("http://localhost:8000/fhir", "0.1.0")
	b.AddResource("Patient", nil, nil)

	cs := b.Build()
	rest := cs["rest"].([]map[string]interface{})
	resources := rest[0]["resource"].([]map[string]interface{})
	if _, ok := resources[0]["interaction"]; ok {
		t.Error("interaction should not be present when no interactions are registered")
	}
}

func TestBuild_SearchIncludeRevIncludeOnly(t *testing.T) {
	b := NewCapabilityBuilder("http://localhost:8000/fhir", "0.1.0")
	b.AddResource("Encounter", []string{"read", "search-type"}, nil)
	b.SetResourceCapabilities("Encounter", ResourceCapabilityOptions{
		SearchInclude:    []string{"Encounter:patient", "Encounter:practitioner"},
		SearchRevInclude: []string{"Observation:encounter", "Condition:encounter"},
	})

	cs := b.Build()
	rest := cs["rest"].([]map[string]interface{})
	resources := rest[0]["resource"].([]map[string]interface{})
	res := resources[0]

	si := res["searchInclude"].([]string)
	if len(si) != 2 {
		t.Fatalf("expected 2 searchInclude, got %d", len(si))
	}
	if si[0] != "Encounter:patient" || si[1] != "Encounter:practitioner" {
		t.Errorf("unexpected searchInclude: %v", si)
	}

	sri := res["searchRevInclude"].([]string)
	if len(sri) != 2 {
		t.Fatalf("expected 2 searchRevInclude, got %d", len(sri))
	}
	if sri[0] != "Observation:encounter" || sri[1] != "Condition:encounter" {
		t.Errorf("unexpected searchRevInclude: %v", sri)
	}
}

// ===========================================================================
// NEW: CapabilityConfig / ResourceCapability / enhanced builder tests
// ===========================================================================

// 1. TestCapabilityBuilder_Build — returns valid CapabilityStatement
func TestCapabilityBuilder_Build(t *testing.T) {
	cfg := CapabilityConfig{
		ServerName:    "Test Server",
		ServerVersion: "1.0.0",
		FHIRVersion:   "4.0.1",
		Publisher:     "Test Publisher",
		Description:   "Test Description",
		BaseURL:       "http://localhost:8000/fhir",
	}
	b := NewCapabilityBuilderFromConfig(cfg)
	b.AddResourceCapability(ResourceCapabilityDef{
		Type:         "Patient",
		Profile:      "http://hl7.org/fhir/StructureDefinition/Patient",
		Interactions: DefaultInteractions(),
		Versioning:   "versioned",
	})

	cs := b.Build()

	if cs["resourceType"] != "CapabilityStatement" {
		t.Errorf("expected resourceType CapabilityStatement, got %v", cs["resourceType"])
	}
	if cs["status"] != "active" {
		t.Errorf("expected status active, got %v", cs["status"])
	}
	if cs["kind"] != "instance" {
		t.Errorf("expected kind instance, got %v", cs["kind"])
	}

	software := cs["software"].(map[string]string)
	if software["name"] != "Test Server" {
		t.Errorf("expected server name 'Test Server', got %s", software["name"])
	}
	if software["version"] != "1.0.0" {
		t.Errorf("expected version 1.0.0, got %s", software["version"])
	}
}

// 2. TestCapabilityBuilder_Build_HasResourceType
func TestCapabilityBuilder_Build_HasResourceType(t *testing.T) {
	b := NewCapabilityBuilderFromConfig(CapabilityConfig{BaseURL: "http://localhost/fhir"})
	cs := b.Build()
	if cs["resourceType"] != "CapabilityStatement" {
		t.Errorf("expected CapabilityStatement, got %v", cs["resourceType"])
	}
}

// 3. TestCapabilityBuilder_Build_HasFHIRVersion
func TestCapabilityBuilder_Build_HasFHIRVersion(t *testing.T) {
	b := NewCapabilityBuilderFromConfig(CapabilityConfig{FHIRVersion: "4.0.1", BaseURL: "http://localhost/fhir"})
	cs := b.Build()
	if cs["fhirVersion"] != "4.0.1" {
		t.Errorf("expected fhirVersion 4.0.1, got %v", cs["fhirVersion"])
	}
}

// 4. TestCapabilityBuilder_Build_HasRest
func TestCapabilityBuilder_Build_HasRest(t *testing.T) {
	b := NewCapabilityBuilderFromConfig(CapabilityConfig{BaseURL: "http://localhost/fhir"})
	cs := b.Build()
	rest := cs["rest"].([]map[string]interface{})
	if len(rest) != 1 {
		t.Fatalf("expected 1 rest entry, got %d", len(rest))
	}
	if rest[0]["mode"] != "server" {
		t.Errorf("expected rest mode server, got %v", rest[0]["mode"])
	}
}

// 5. TestCapabilityBuilder_AddResourceCapability
func TestCapabilityBuilder_AddResourceCapability(t *testing.T) {
	b := NewCapabilityBuilderFromConfig(CapabilityConfig{BaseURL: "http://localhost/fhir"})
	b.AddResourceCapability(ResourceCapabilityDef{
		Type:         "Observation",
		Profile:      "http://hl7.org/fhir/StructureDefinition/Observation",
		Interactions: []string{"read", "search-type"},
		Versioning:   "versioned",
	})

	cs := b.Build()
	rest := cs["rest"].([]map[string]interface{})
	resources := rest[0]["resource"].([]map[string]interface{})

	if len(resources) != 1 {
		t.Fatalf("expected 1 resource, got %d", len(resources))
	}
	if resources[0]["type"] != "Observation" {
		t.Errorf("expected Observation, got %v", resources[0]["type"])
	}
}

// 6. TestCapabilityBuilder_AddResourceCapability_Interactions
func TestCapabilityBuilder_AddResourceCapability_Interactions(t *testing.T) {
	b := NewCapabilityBuilderFromConfig(CapabilityConfig{BaseURL: "http://localhost/fhir"})
	b.AddResourceCapability(ResourceCapabilityDef{
		Type:         "Patient",
		Interactions: []string{"read", "create", "update", "delete"},
		Versioning:   "versioned",
	})

	cs := b.Build()
	rest := cs["rest"].([]map[string]interface{})
	resources := rest[0]["resource"].([]map[string]interface{})
	interactions := resources[0]["interaction"].([]map[string]string)

	if len(interactions) != 4 {
		t.Fatalf("expected 4 interactions, got %d", len(interactions))
	}

	codes := make(map[string]bool)
	for _, ia := range interactions {
		codes[ia["code"]] = true
	}
	for _, expected := range []string{"read", "create", "update", "delete"} {
		if !codes[expected] {
			t.Errorf("missing interaction: %s", expected)
		}
	}
}

// 7. TestCapabilityBuilder_AddResourceCapability_SearchParams
func TestCapabilityBuilder_AddResourceCapability_SearchParams(t *testing.T) {
	b := NewCapabilityBuilderFromConfig(CapabilityConfig{BaseURL: "http://localhost/fhir"})
	b.AddResourceCapability(ResourceCapabilityDef{
		Type:         "Patient",
		Interactions: []string{"search-type"},
		SearchParams: []SearchParamCapability{
			{Name: "name", Type: "string", Documentation: "Patient name"},
			{Name: "birthdate", Type: "date", Documentation: "Date of birth"},
			{Name: "identifier", Type: "token", Documentation: "Patient identifier"},
		},
		Versioning: "versioned",
	})

	cs := b.Build()
	rest := cs["rest"].([]map[string]interface{})
	resources := rest[0]["resource"].([]map[string]interface{})
	params := resources[0]["searchParam"].([]map[string]string)

	if len(params) != 3 {
		t.Fatalf("expected 3 search params, got %d", len(params))
	}

	paramMap := make(map[string]string)
	for _, p := range params {
		paramMap[p["name"]] = p["type"]
	}
	if paramMap["name"] != "string" {
		t.Errorf("expected name type string, got %s", paramMap["name"])
	}
	if paramMap["birthdate"] != "date" {
		t.Errorf("expected birthdate type date, got %s", paramMap["birthdate"])
	}
	if paramMap["identifier"] != "token" {
		t.Errorf("expected identifier type token, got %s", paramMap["identifier"])
	}
}

// 8. TestCapabilityBuilder_AddServerOperation
func TestCapabilityBuilder_AddServerOperation(t *testing.T) {
	b := NewCapabilityBuilderFromConfig(CapabilityConfig{BaseURL: "http://localhost/fhir"})
	b.AddServerOperation(OperationCapability{
		Name:          "$export",
		Definition:    "http://hl7.org/fhir/uv/bulkdata/OperationDefinition/export",
		Documentation: "Bulk data export",
	})

	cs := b.Build()
	rest := cs["rest"].([]map[string]interface{})
	operations := rest[0]["operation"].([]map[string]interface{})

	if len(operations) != 1 {
		t.Fatalf("expected 1 operation, got %d", len(operations))
	}
	if operations[0]["name"] != "$export" {
		t.Errorf("expected $export, got %v", operations[0]["name"])
	}
	if operations[0]["definition"] != "http://hl7.org/fhir/uv/bulkdata/OperationDefinition/export" {
		t.Errorf("unexpected definition: %v", operations[0]["definition"])
	}
}

// 9. TestCapabilityBuilder_Build_HasSecurity
func TestCapabilityBuilder_Build_HasSecurity(t *testing.T) {
	b := NewCapabilityBuilderFromConfig(CapabilityConfig{BaseURL: "http://localhost/fhir"})
	b.SetOAuthURIs("http://auth/authorize", "http://auth/token")

	cs := b.Build()
	rest := cs["rest"].([]map[string]interface{})
	security, ok := rest[0]["security"].(map[string]interface{})
	if !ok {
		t.Fatal("expected security section")
	}

	services := security["service"].([]map[string]interface{})
	codings := services[0]["coding"].([]map[string]string)
	if codings[0]["code"] != "SMART-on-FHIR" {
		t.Errorf("expected SMART-on-FHIR, got %s", codings[0]["code"])
	}
}

// 10. TestCapabilityBuilder_Build_HasFormats
func TestCapabilityBuilder_Build_HasFormats(t *testing.T) {
	cfg := CapabilityConfig{
		BaseURL:          "http://localhost/fhir",
		SupportedFormats: []string{"application/fhir+json"},
	}
	b := NewCapabilityBuilderFromConfig(cfg)
	cs := b.Build()
	formats := cs["format"].([]string)
	found := false
	for _, f := range formats {
		if f == "application/fhir+json" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected application/fhir+json in formats, got %v", formats)
	}
}

// ===========================================================================
// Default builder tests
// ===========================================================================

// 11. TestDefaultCapabilityBuilder_HasPatient
func TestDefaultCapabilityBuilder_HasPatient(t *testing.T) {
	b := DefaultCapabilityBuilder()
	cs := b.Build()
	rest := cs["rest"].([]map[string]interface{})
	resources := rest[0]["resource"].([]map[string]interface{})

	var patient map[string]interface{}
	for _, r := range resources {
		if r["type"] == "Patient" {
			patient = r
			break
		}
	}
	if patient == nil {
		t.Fatal("Patient resource not found")
	}

	params := patient["searchParam"].([]map[string]string)
	if len(params) < 20 {
		t.Errorf("expected at least 20 Patient search params, got %d", len(params))
	}
}

// 12. TestDefaultCapabilityBuilder_HasObservation
func TestDefaultCapabilityBuilder_HasObservation(t *testing.T) {
	b := DefaultCapabilityBuilder()
	cs := b.Build()
	rest := cs["rest"].([]map[string]interface{})
	resources := rest[0]["resource"].([]map[string]interface{})

	var obs map[string]interface{}
	for _, r := range resources {
		if r["type"] == "Observation" {
			obs = r
			break
		}
	}
	if obs == nil {
		t.Fatal("Observation resource not found")
	}

	params := obs["searchParam"].([]map[string]string)
	if len(params) < 5 {
		t.Errorf("expected at least 5 Observation search params, got %d", len(params))
	}
}

// 13. TestDefaultCapabilityBuilder_HasEncounter
func TestDefaultCapabilityBuilder_HasEncounter(t *testing.T) {
	b := DefaultCapabilityBuilder()
	cs := b.Build()
	rest := cs["rest"].([]map[string]interface{})
	resources := rest[0]["resource"].([]map[string]interface{})

	var enc map[string]interface{}
	for _, r := range resources {
		if r["type"] == "Encounter" {
			enc = r
			break
		}
	}
	if enc == nil {
		t.Fatal("Encounter resource not found")
	}

	params := enc["searchParam"].([]map[string]string)
	if len(params) < 5 {
		t.Errorf("expected at least 5 Encounter search params, got %d", len(params))
	}
}

// 14. TestDefaultCapabilityBuilder_Has20PlusResources
func TestDefaultCapabilityBuilder_Has20PlusResources(t *testing.T) {
	b := DefaultCapabilityBuilder()
	cs := b.Build()
	rest := cs["rest"].([]map[string]interface{})
	resources := rest[0]["resource"].([]map[string]interface{})

	if len(resources) < 20 {
		t.Errorf("expected at least 20 resource types, got %d", len(resources))
	}
}

// 15. TestDefaultCapabilityBuilder_HasOperations
func TestDefaultCapabilityBuilder_HasOperations(t *testing.T) {
	b := DefaultCapabilityBuilder()
	cs := b.Build()
	rest := cs["rest"].([]map[string]interface{})
	operations := rest[0]["operation"].([]map[string]interface{})

	if len(operations) < 12 {
		t.Errorf("expected at least 12 server operations, got %d", len(operations))
	}

	opNames := make(map[string]bool)
	for _, op := range operations {
		opNames[op["name"].(string)] = true
	}
	for _, expected := range []string{"$export", "$everything", "$validate", "$match", "$graphql"} {
		if !opNames[expected] {
			t.Errorf("missing server operation: %s", expected)
		}
	}
}

// 16. TestDefaultCapabilityBuilder_PatientSearchParams
func TestDefaultCapabilityBuilder_PatientSearchParams(t *testing.T) {
	b := DefaultCapabilityBuilder()
	cs := b.Build()
	rest := cs["rest"].([]map[string]interface{})
	resources := rest[0]["resource"].([]map[string]interface{})

	var patient map[string]interface{}
	for _, r := range resources {
		if r["type"] == "Patient" {
			patient = r
			break
		}
	}
	if patient == nil {
		t.Fatal("Patient resource not found")
	}

	params := patient["searchParam"].([]map[string]string)
	paramNames := make(map[string]bool)
	for _, p := range params {
		paramNames[p["name"]] = true
	}

	expected := []string{"name", "family", "given", "birthdate", "gender", "identifier",
		"address", "phone", "email", "_id", "_lastUpdated", "general-practitioner",
		"organization", "active", "deceased", "death-date", "language",
		"address-city", "address-state", "address-postalcode"}
	for _, e := range expected {
		if !paramNames[e] {
			t.Errorf("missing Patient search param: %s", e)
		}
	}
}

// 17. TestDefaultCapabilityBuilder_ObservationSearchParams
func TestDefaultCapabilityBuilder_ObservationSearchParams(t *testing.T) {
	b := DefaultCapabilityBuilder()
	cs := b.Build()
	rest := cs["rest"].([]map[string]interface{})
	resources := rest[0]["resource"].([]map[string]interface{})

	var obs map[string]interface{}
	for _, r := range resources {
		if r["type"] == "Observation" {
			obs = r
			break
		}
	}
	if obs == nil {
		t.Fatal("Observation resource not found")
	}

	params := obs["searchParam"].([]map[string]string)
	paramNames := make(map[string]bool)
	for _, p := range params {
		paramNames[p["name"]] = true
	}

	for _, e := range []string{"category", "code", "date"} {
		if !paramNames[e] {
			t.Errorf("missing Observation search param: %s", e)
		}
	}
}

// 18. TestDefaultCapabilityBuilder_SystemInteractions
func TestDefaultCapabilityBuilder_SystemInteractions(t *testing.T) {
	b := DefaultCapabilityBuilder()
	cs := b.Build()
	rest := cs["rest"].([]map[string]interface{})
	interactions := rest[0]["interaction"].([]map[string]string)

	codes := make(map[string]bool)
	for _, ia := range interactions {
		codes[ia["code"]] = true
	}
	for _, expected := range []string{"transaction", "batch", "search-system", "history-system"} {
		if !codes[expected] {
			t.Errorf("missing system interaction: %s", expected)
		}
	}
}

// ===========================================================================
// Custom search params
// ===========================================================================

// 19. TestCapabilityBuilder_AddCustomSearchParam
func TestCapabilityBuilder_AddCustomSearchParam(t *testing.T) {
	b := NewCapabilityBuilderFromConfig(CapabilityConfig{BaseURL: "http://localhost/fhir"})
	b.AddResourceCapability(ResourceCapabilityDef{
		Type:         "Patient",
		Interactions: []string{"read", "search-type"},
		Versioning:   "versioned",
	})

	b.AddCustomSearchParam(CustomSearchParam{
		Name:         "my-custom-param",
		Type:         "string",
		ResourceType: "Patient",
		Expression:   "Patient.extension.where(url='http://example.com/custom').value",
		Description:  "Custom search parameter",
	})

	cs := b.Build()
	rest := cs["rest"].([]map[string]interface{})
	resources := rest[0]["resource"].([]map[string]interface{})

	var patient map[string]interface{}
	for _, r := range resources {
		if r["type"] == "Patient" {
			patient = r
			break
		}
	}
	if patient == nil {
		t.Fatal("Patient not found")
	}

	params := patient["searchParam"].([]map[string]string)
	found := false
	for _, p := range params {
		if p["name"] == "my-custom-param" {
			found = true
			if p["type"] != "string" {
				t.Errorf("expected type string, got %s", p["type"])
			}
		}
	}
	if !found {
		t.Error("custom search param not found in build output")
	}
}

// 20. TestCapabilityBuilder_ListCustomSearchParams
func TestCapabilityBuilder_ListCustomSearchParams(t *testing.T) {
	b := NewCapabilityBuilderFromConfig(CapabilityConfig{BaseURL: "http://localhost/fhir"})
	b.AddCustomSearchParam(CustomSearchParam{
		Name:         "param-a",
		Type:         "string",
		ResourceType: "Patient",
		Expression:   "Patient.name.text",
		Description:  "Custom A",
	})
	b.AddCustomSearchParam(CustomSearchParam{
		Name:         "param-b",
		Type:         "token",
		ResourceType: "Patient",
		Expression:   "Patient.identifier",
		Description:  "Custom B",
	})
	b.AddCustomSearchParam(CustomSearchParam{
		Name:         "param-c",
		Type:         "reference",
		ResourceType: "Observation",
		Expression:   "Observation.subject",
		Description:  "Custom C",
	})

	patientParams := b.ListCustomSearchParams("Patient")
	if len(patientParams) != 2 {
		t.Fatalf("expected 2 custom params for Patient, got %d", len(patientParams))
	}

	obsParams := b.ListCustomSearchParams("Observation")
	if len(obsParams) != 1 {
		t.Fatalf("expected 1 custom param for Observation, got %d", len(obsParams))
	}
}

// 21. TestCapabilityBuilder_DeleteCustomSearchParam
func TestCapabilityBuilder_DeleteCustomSearchParam(t *testing.T) {
	b := NewCapabilityBuilderFromConfig(CapabilityConfig{BaseURL: "http://localhost/fhir"})
	b.AddCustomSearchParam(CustomSearchParam{
		Name:         "deleteme",
		Type:         "string",
		ResourceType: "Patient",
		Expression:   "Patient.name",
		Description:  "To be deleted",
	})

	if len(b.ListCustomSearchParams("Patient")) != 1 {
		t.Fatal("expected 1 custom param before delete")
	}

	err := b.DeleteCustomSearchParam("Patient", "deleteme")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(b.ListCustomSearchParams("Patient")) != 0 {
		t.Error("expected 0 custom params after delete")
	}
}

// 22. TestCapabilityBuilder_DeleteCustomSearchParam_NotFound
func TestCapabilityBuilder_DeleteCustomSearchParam_NotFound(t *testing.T) {
	b := NewCapabilityBuilderFromConfig(CapabilityConfig{BaseURL: "http://localhost/fhir"})
	err := b.DeleteCustomSearchParam("Patient", "nonexistent")
	if err == nil {
		t.Error("expected error for deleting nonexistent param")
	}
}

// ===========================================================================
// Handler tests
// ===========================================================================

// 23. TestCapabilityHandler_Metadata
func TestCapabilityHandler_Metadata(t *testing.T) {
	b := DefaultCapabilityBuilder()
	h := NewCapabilityHandler(b)

	e := echo.New()
	h.RegisterRoutes(e.Group("/fhir"))

	req := httptest.NewRequest(http.MethodGet, "/fhir/metadata", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var cs map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &cs); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if cs["resourceType"] != "CapabilityStatement" {
		t.Errorf("expected CapabilityStatement, got %v", cs["resourceType"])
	}
}

// 24. TestCapabilityHandler_ListResources
func TestCapabilityHandler_ListResources(t *testing.T) {
	b := DefaultCapabilityBuilder()
	h := NewCapabilityHandler(b)

	e := echo.New()
	h.RegisterRoutes(e.Group("/fhir"))

	req := httptest.NewRequest(http.MethodGet, "/fhir/metadata/resources", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	resourceTypes, ok := result["resourceTypes"].([]interface{})
	if !ok {
		t.Fatal("expected resourceTypes array in response")
	}
	if len(resourceTypes) < 20 {
		t.Errorf("expected at least 20 resource types, got %d", len(resourceTypes))
	}
}

// 25. TestCapabilityHandler_GetResourceCapability
func TestCapabilityHandler_GetResourceCapability(t *testing.T) {
	b := DefaultCapabilityBuilder()
	h := NewCapabilityHandler(b)

	e := echo.New()
	h.RegisterRoutes(e.Group("/fhir"))

	req := httptest.NewRequest(http.MethodGet, "/fhir/metadata/resources/Patient", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if result["type"] != "Patient" {
		t.Errorf("expected Patient, got %v", result["type"])
	}
}

// TestCapabilityHandler_GetResourceCapability_NotFound verifies 404 for unknown type.
func TestCapabilityHandler_GetResourceCapability_NotFound(t *testing.T) {
	b := DefaultCapabilityBuilder()
	h := NewCapabilityHandler(b)

	e := echo.New()
	h.RegisterRoutes(e.Group("/fhir"))

	req := httptest.NewRequest(http.MethodGet, "/fhir/metadata/resources/NonExistent", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

// 26. TestCapabilityHandler_ListOperations
func TestCapabilityHandler_ListOperations(t *testing.T) {
	b := DefaultCapabilityBuilder()
	h := NewCapabilityHandler(b)

	e := echo.New()
	h.RegisterRoutes(e.Group("/fhir"))

	req := httptest.NewRequest(http.MethodGet, "/fhir/metadata/operations", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	operations, ok := result["operations"].([]interface{})
	if !ok {
		t.Fatal("expected operations array in response")
	}
	if len(operations) < 12 {
		t.Errorf("expected at least 12 operations, got %d", len(operations))
	}
}

// 27. TestCapabilityHandler_RegisterCustomParam
func TestCapabilityHandler_RegisterCustomParam(t *testing.T) {
	b := DefaultCapabilityBuilder()
	h := NewCapabilityHandler(b)

	e := echo.New()
	h.RegisterRoutes(e.Group("/fhir"))

	body := `{"name":"custom-mrn","type":"token","resourceType":"Patient","expression":"Patient.identifier.where(system='http://example.com/mrn')","description":"MRN search"}`
	req := httptest.NewRequest(http.MethodPost, "/fhir/metadata/search-params", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d; body: %s", rec.Code, rec.Body.String())
	}

	// Verify param was registered
	params := b.ListCustomSearchParams("Patient")
	found := false
	for _, p := range params {
		if p.Name == "custom-mrn" {
			found = true
		}
	}
	if !found {
		t.Error("custom param not registered after POST")
	}
}

// TestCapabilityHandler_ListCustomSearchParams verifies GET /fhir/metadata/search-params.
func TestCapabilityHandler_ListCustomSearchParams(t *testing.T) {
	b := DefaultCapabilityBuilder()
	h := NewCapabilityHandler(b)

	b.AddCustomSearchParam(CustomSearchParam{
		Name:         "custom-list-test",
		Type:         "string",
		ResourceType: "Observation",
		Expression:   "Observation.code.text",
		Description:  "list test param",
	})

	e := echo.New()
	h.RegisterRoutes(e.Group("/fhir"))

	req := httptest.NewRequest(http.MethodGet, "/fhir/metadata/search-params", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	customParams, ok := result["customSearchParams"].([]interface{})
	if !ok {
		t.Fatal("expected customSearchParams array")
	}
	if len(customParams) < 1 {
		t.Error("expected at least 1 custom search param")
	}
}

// TestCapabilityHandler_DeleteCustomParam verifies DELETE /fhir/metadata/search-params/:type/:name.
func TestCapabilityHandler_DeleteCustomParam(t *testing.T) {
	b := DefaultCapabilityBuilder()
	h := NewCapabilityHandler(b)

	b.AddCustomSearchParam(CustomSearchParam{
		Name:         "to-delete",
		Type:         "string",
		ResourceType: "Patient",
		Expression:   "Patient.name.text",
		Description:  "delete test",
	})

	e := echo.New()
	h.RegisterRoutes(e.Group("/fhir"))

	req := httptest.NewRequest(http.MethodDelete, "/fhir/metadata/search-params/Patient/to-delete", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d; body: %s", rec.Code, rec.Body.String())
	}

	params := b.ListCustomSearchParams("Patient")
	for _, p := range params {
		if p.Name == "to-delete" {
			t.Error("param should have been deleted")
		}
	}
}

// 28. TestCapabilityHandler_ConcurrentAccess
func TestCapabilityHandler_ConcurrentAccess(t *testing.T) {
	b := DefaultCapabilityBuilder()
	h := NewCapabilityHandler(b)

	e := echo.New()
	h.RegisterRoutes(e.Group("/fhir"))

	var wg sync.WaitGroup
	errs := make(chan error, 50)

	// Concurrent metadata reads
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			req := httptest.NewRequest(http.MethodGet, "/fhir/metadata", nil)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)
			if rec.Code != http.StatusOK {
				errs <- fmt.Errorf("metadata returned %d", rec.Code)
			}
		}()
	}

	// Concurrent custom param registrations
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			body := fmt.Sprintf(`{"name":"concurrent-%d","type":"string","resourceType":"Patient","expression":"Patient.name","description":"concurrent test"}`, idx)
			req := httptest.NewRequest(http.MethodPost, "/fhir/metadata/search-params", strings.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)
			if rec.Code != http.StatusCreated {
				errs <- fmt.Errorf("register param returned %d", rec.Code)
			}
		}(i)
	}

	// Concurrent builds
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = b.Build()
		}()
	}

	wg.Wait()
	close(errs)

	for err := range errs {
		t.Errorf("concurrent error: %v", err)
	}
}

// ===========================================================================
// Additional structural tests for full coverage
// ===========================================================================

// TestDefaultCapabilityBuilder_HasCondition verifies Condition resource in defaults.
func TestDefaultCapabilityBuilder_HasCondition(t *testing.T) {
	b := DefaultCapabilityBuilder()
	cs := b.Build()
	rest := cs["rest"].([]map[string]interface{})
	resources := rest[0]["resource"].([]map[string]interface{})

	found := false
	for _, r := range resources {
		if r["type"] == "Condition" {
			found = true
			params := r["searchParam"].([]map[string]string)
			paramNames := make(map[string]bool)
			for _, p := range params {
				paramNames[p["name"]] = true
			}
			for _, expected := range []string{"patient", "code", "clinical-status", "category"} {
				if !paramNames[expected] {
					t.Errorf("Condition missing search param: %s", expected)
				}
			}
			break
		}
	}
	if !found {
		t.Error("Condition resource not found in default builder")
	}
}

// TestDefaultCapabilityBuilder_HasMedicationRequest verifies MedicationRequest in defaults.
func TestDefaultCapabilityBuilder_HasMedicationRequest(t *testing.T) {
	b := DefaultCapabilityBuilder()
	cs := b.Build()
	rest := cs["rest"].([]map[string]interface{})
	resources := rest[0]["resource"].([]map[string]interface{})

	found := false
	for _, r := range resources {
		if r["type"] == "MedicationRequest" {
			found = true
			params := r["searchParam"].([]map[string]string)
			paramNames := make(map[string]bool)
			for _, p := range params {
				paramNames[p["name"]] = true
			}
			for _, expected := range []string{"patient", "status", "intent"} {
				if !paramNames[expected] {
					t.Errorf("MedicationRequest missing search param: %s", expected)
				}
			}
			break
		}
	}
	if !found {
		t.Error("MedicationRequest resource not found in default builder")
	}
}

// TestCapabilityConfig_Defaults verifies default config values are set.
func TestCapabilityConfig_Defaults(t *testing.T) {
	cfg := CapabilityConfig{}
	b := NewCapabilityBuilderFromConfig(cfg)
	cs := b.Build()

	software := cs["software"].(map[string]string)
	if software["name"] != "Headless EHR FHIR Server" {
		t.Errorf("expected default server name, got %s", software["name"])
	}
	if cs["fhirVersion"] != "4.0.1" {
		t.Errorf("expected default FHIR version 4.0.1, got %v", cs["fhirVersion"])
	}
}

// TestResourceCapability_WithOperations verifies resource-level operations in build.
func TestResourceCapability_WithOperations(t *testing.T) {
	b := NewCapabilityBuilderFromConfig(CapabilityConfig{BaseURL: "http://localhost/fhir"})
	b.AddResourceCapability(ResourceCapabilityDef{
		Type:         "Patient",
		Interactions: []string{"read"},
		Operations: []OperationCapability{
			{Name: "$everything", Definition: "http://hl7.org/fhir/OperationDefinition/Patient-everything"},
		},
		Versioning: "versioned",
	})

	cs := b.Build()
	rest := cs["rest"].([]map[string]interface{})
	resources := rest[0]["resource"].([]map[string]interface{})

	ops, ok := resources[0]["operation"].([]map[string]interface{})
	if !ok {
		t.Fatal("expected operations on resource")
	}
	if len(ops) != 1 {
		t.Fatalf("expected 1 operation, got %d", len(ops))
	}
	if ops[0]["name"] != "$everything" {
		t.Errorf("expected $everything, got %v", ops[0]["name"])
	}
}

// TestDefaultCapabilityBuilder_ResourceProfile verifies profiles are set on default resources.
func TestDefaultCapabilityBuilder_ResourceProfile(t *testing.T) {
	b := DefaultCapabilityBuilder()
	cs := b.Build()
	rest := cs["rest"].([]map[string]interface{})
	resources := rest[0]["resource"].([]map[string]interface{})

	for _, r := range resources {
		if r["type"] == "Patient" {
			profile, ok := r["profile"].(string)
			if !ok {
				t.Fatal("expected profile on Patient resource")
			}
			if profile != "http://hl7.org/fhir/StructureDefinition/Patient" {
				t.Errorf("unexpected profile: %s", profile)
			}
			return
		}
	}
	t.Fatal("Patient resource not found")
}

// TestDefaultCapabilityBuilder_Versioning verifies versioning is set on default resources.
func TestDefaultCapabilityBuilder_Versioning(t *testing.T) {
	b := DefaultCapabilityBuilder()
	cs := b.Build()
	rest := cs["rest"].([]map[string]interface{})
	resources := rest[0]["resource"].([]map[string]interface{})

	for _, r := range resources {
		if r["versioning"] != "versioned" {
			t.Errorf("expected versioning=versioned on %s, got %v", r["type"], r["versioning"])
		}
	}
}
