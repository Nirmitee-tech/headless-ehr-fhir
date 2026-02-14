package fhir

import (
	"encoding/json"
	"testing"
)

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
