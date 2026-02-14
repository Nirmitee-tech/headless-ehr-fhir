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

	// Check Patient has 6 interactions (default)
	patInteractions := resources[2]["interaction"].([]map[string]string)
	if len(patInteractions) != 6 {
		t.Errorf("expected 6 Patient interactions, got %d", len(patInteractions))
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
	if len(interactions) != 6 {
		t.Fatalf("expected 6 interactions, got %d", len(interactions))
	}

	expected := map[string]bool{
		"read": true, "vread": true, "search-type": true,
		"create": true, "update": true, "delete": true,
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
