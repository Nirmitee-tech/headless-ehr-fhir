package fhir

import (
	"context"
	"fmt"
	"testing"
)

func TestIncludeRegistry_RegisterAndFetch(t *testing.T) {
	registry := NewIncludeRegistry()

	// Register a mock fetcher
	registry.RegisterFetcher("Patient", func(ctx context.Context, fhirID string) (map[string]interface{}, error) {
		return map[string]interface{}{
			"resourceType": "Patient",
			"id":           fhirID,
			"name":         "Test Patient",
		}, nil
	})

	// Register a reference
	registry.RegisterReference("Observation", "subject", "Patient")

	// Check include targets
	targets := registry.GetIncludeTargets("Observation")
	if len(targets) != 1 {
		t.Fatalf("expected 1 include target, got %d", len(targets))
	}
	if targets["subject"].TargetType != "Patient" {
		t.Errorf("target type = %q, want 'Patient'", targets["subject"].TargetType)
	}

	// Check reverse include
	revTargets := registry.GetRevIncludeTargets("Patient")
	if len(revTargets) != 1 {
		t.Fatalf("expected 1 rev include target, got %d", len(revTargets))
	}
	if revTargets[0].SourceType != "Observation" {
		t.Errorf("source type = %q, want 'Observation'", revTargets[0].SourceType)
	}
}

func TestIncludeRegistry_ResolveIncludes(t *testing.T) {
	registry := NewIncludeRegistry()

	// Register a mock patient fetcher
	registry.RegisterFetcher("Patient", func(ctx context.Context, fhirID string) (map[string]interface{}, error) {
		return map[string]interface{}{
			"resourceType": "Patient",
			"id":           fhirID,
		}, nil
	})
	registry.RegisterReference("Observation", "subject", "Patient")

	// Create test resources with references
	resources := []interface{}{
		map[string]interface{}{
			"resourceType": "Observation",
			"id":           "obs-1",
			"subject": map[string]interface{}{
				"reference": "Patient/pat-1",
			},
		},
		map[string]interface{}{
			"resourceType": "Observation",
			"id":           "obs-2",
			"subject": map[string]interface{}{
				"reference": "Patient/pat-1", // duplicate reference
			},
		},
	}

	entries, err := registry.ResolveIncludes(context.Background(), resources, []string{"Observation:subject"})
	if err != nil {
		t.Fatalf("ResolveIncludes failed: %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("expected 1 include entry (deduplicated), got %d", len(entries))
	}
	if entries[0].Search == nil || entries[0].Search.Mode != "include" {
		t.Error("include entry should have search.mode='include'")
	}
}

func TestIncludeRegistry_ResolveIncludes_Empty(t *testing.T) {
	registry := NewIncludeRegistry()

	entries, err := registry.ResolveIncludes(context.Background(), nil, []string{"Obs:subject"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if entries != nil {
		t.Error("empty resources should return nil entries")
	}
}

func TestSearchParamToFields(t *testing.T) {
	tests := []struct {
		param    string
		expected []string
	}{
		{"subject", []string{"subject"}},
		{"patient", []string{"subject", "patient"}},
		{"encounter", []string{"encounter"}},
		{"unknown", []string{"unknown"}},
	}

	for _, tt := range tests {
		t.Run(tt.param, func(t *testing.T) {
			fields := searchParamToFields(tt.param)
			if len(fields) != len(tt.expected) {
				t.Fatalf("fields for %q: got %d, want %d", tt.param, len(fields), len(tt.expected))
			}
			for i, f := range fields {
				if f != tt.expected[i] {
					t.Errorf("field[%d] = %q, want %q", i, f, tt.expected[i])
				}
			}
		})
	}
}

func TestExtractReferenceID(t *testing.T) {
	resource := map[string]interface{}{
		"subject": map[string]interface{}{
			"reference": "Patient/123",
		},
	}

	id := extractReferenceID(resource, "subject", "Patient")
	if id != "123" {
		t.Errorf("extractReferenceID = %q, want '123'", id)
	}

	id = extractReferenceID(resource, "subject", "Organization")
	if id != "" {
		t.Errorf("extractReferenceID for wrong type = %q, want ''", id)
	}

	id = extractReferenceID(resource, "encounter", "Encounter")
	if id != "" {
		t.Errorf("missing field should return empty string")
	}
}

func TestResolveIncludes_InvalidParamFormat(t *testing.T) {
	registry := NewIncludeRegistry()
	registry.RegisterFetcher("Patient", func(ctx context.Context, fhirID string) (map[string]interface{}, error) {
		return map[string]interface{}{"resourceType": "Patient", "id": fhirID}, nil
	})
	registry.RegisterReference("Observation", "subject", "Patient")

	resources := []interface{}{
		map[string]interface{}{
			"resourceType": "Observation",
			"id":           "obs-1",
			"subject":      map[string]interface{}{"reference": "Patient/pat-1"},
		},
	}

	// Include param with less than 2 parts (no colon) should be skipped
	entries, err := registry.ResolveIncludes(context.Background(), resources, []string{"Observation"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected 0 entries for invalid param format, got %d", len(entries))
	}
}

func TestResolveIncludes_UnknownSourceType(t *testing.T) {
	registry := NewIncludeRegistry()
	registry.RegisterFetcher("Patient", func(ctx context.Context, fhirID string) (map[string]interface{}, error) {
		return map[string]interface{}{"resourceType": "Patient", "id": fhirID}, nil
	})
	registry.RegisterReference("Observation", "subject", "Patient")

	resources := []interface{}{
		map[string]interface{}{
			"resourceType": "Observation",
			"id":           "obs-1",
			"subject":      map[string]interface{}{"reference": "Patient/pat-1"},
		},
	}

	// "UnknownType" is not registered as a source type
	entries, err := registry.ResolveIncludes(context.Background(), resources, []string{"UnknownType:subject"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected 0 entries for unknown source type, got %d", len(entries))
	}
}

func TestResolveIncludes_UnknownSearchParam(t *testing.T) {
	registry := NewIncludeRegistry()
	registry.RegisterFetcher("Patient", func(ctx context.Context, fhirID string) (map[string]interface{}, error) {
		return map[string]interface{}{"resourceType": "Patient", "id": fhirID}, nil
	})
	registry.RegisterReference("Observation", "subject", "Patient")

	resources := []interface{}{
		map[string]interface{}{
			"resourceType": "Observation",
			"id":           "obs-1",
			"subject":      map[string]interface{}{"reference": "Patient/pat-1"},
		},
	}

	// "nonexistent" search param is not registered for Observation
	entries, err := registry.ResolveIncludes(context.Background(), resources, []string{"Observation:nonexistent"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected 0 entries for unknown search param, got %d", len(entries))
	}
}

func TestResolveIncludes_ThreePartParamTargetTypeOverride(t *testing.T) {
	registry := NewIncludeRegistry()

	// Register fetchers for both Patient and Practitioner
	registry.RegisterFetcher("Patient", func(ctx context.Context, fhirID string) (map[string]interface{}, error) {
		return map[string]interface{}{"resourceType": "Patient", "id": fhirID}, nil
	})
	registry.RegisterFetcher("Practitioner", func(ctx context.Context, fhirID string) (map[string]interface{}, error) {
		return map[string]interface{}{"resourceType": "Practitioner", "id": fhirID}, nil
	})
	// Register reference from Observation:subject -> Patient
	registry.RegisterReference("Observation", "subject", "Patient")

	resources := []interface{}{
		map[string]interface{}{
			"resourceType": "Observation",
			"id":           "obs-1",
			"subject":      map[string]interface{}{"reference": "Practitioner/pract-1"},
		},
	}

	// Use 3-part include to override target type to Practitioner
	entries, err := registry.ResolveIncludes(context.Background(), resources, []string{"Observation:subject:Practitioner"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry with target type override, got %d", len(entries))
	}
	if entries[0].FullURL != "Practitioner/pract-1" {
		t.Errorf("expected fullUrl 'Practitioner/pract-1', got %q", entries[0].FullURL)
	}
}

func TestResolveIncludes_FetcherReturnsError(t *testing.T) {
	registry := NewIncludeRegistry()

	// Register a fetcher that always returns an error
	registry.RegisterFetcher("Patient", func(ctx context.Context, fhirID string) (map[string]interface{}, error) {
		return nil, fmt.Errorf("database connection failed")
	})
	registry.RegisterReference("Observation", "subject", "Patient")

	resources := []interface{}{
		map[string]interface{}{
			"resourceType": "Observation",
			"id":           "obs-1",
			"subject":      map[string]interface{}{"reference": "Patient/pat-1"},
		},
	}

	// Fetcher error should be skipped, not returned
	entries, err := registry.ResolveIncludes(context.Background(), resources, []string{"Observation:subject"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected 0 entries when fetcher errors, got %d", len(entries))
	}
}

func TestExtractReferenceID_ResourceNotAMap(t *testing.T) {
	// Pass a non-map, non-struct type that cannot be converted
	id := extractReferenceID(make(chan int), "subject", "Patient")
	if id != "" {
		t.Errorf("expected empty string for non-map resource, got %q", id)
	}
}

func TestExtractReferenceID_ReferenceFieldNotAMap(t *testing.T) {
	// The "subject" field is a string instead of a map with a "reference" key
	resource := map[string]interface{}{
		"subject": "Patient/123",
	}
	id := extractReferenceID(resource, "subject", "Patient")
	if id != "" {
		t.Errorf("expected empty string when reference field is not a map, got %q", id)
	}
}

func TestExtractReferenceID_ReferenceStringWithoutSlash(t *testing.T) {
	// The reference string does not contain a "/" separator
	resource := map[string]interface{}{
		"subject": map[string]interface{}{
			"reference": "just-an-id-no-slash",
		},
	}
	id := extractReferenceID(resource, "subject", "Patient")
	if id != "" {
		t.Errorf("expected empty string when reference has no slash, got %q", id)
	}
}

func TestExtractReferenceID_EmptyTargetType(t *testing.T) {
	// When targetType is empty, it should match any type
	resource := map[string]interface{}{
		"subject": map[string]interface{}{
			"reference": "Patient/pat-99",
		},
	}
	id := extractReferenceID(resource, "subject", "")
	if id != "pat-99" {
		t.Errorf("expected 'pat-99' with empty targetType, got %q", id)
	}
}

func TestIncludeRegistry_GetIncludeTargets_Empty(t *testing.T) {
	registry := NewIncludeRegistry()
	targets := registry.GetIncludeTargets("NonExistent")
	if targets != nil {
		t.Errorf("expected nil for unregistered resource type, got %v", targets)
	}
}

func TestIncludeRegistry_GetRevIncludeTargets_Empty(t *testing.T) {
	registry := NewIncludeRegistry()
	targets := registry.GetRevIncludeTargets("NonExistent")
	if targets != nil {
		t.Errorf("expected nil for unregistered resource type, got %v", targets)
	}
}

func TestIncludeRegistry_MultipleReferences(t *testing.T) {
	registry := NewIncludeRegistry()

	registry.RegisterReference("Observation", "subject", "Patient")
	registry.RegisterReference("Observation", "performer", "Practitioner")

	targets := registry.GetIncludeTargets("Observation")
	if len(targets) != 2 {
		t.Fatalf("expected 2 include targets, got %d", len(targets))
	}
	if targets["subject"].TargetType != "Patient" {
		t.Errorf("subject target = %q, want Patient", targets["subject"].TargetType)
	}
	if targets["performer"].TargetType != "Practitioner" {
		t.Errorf("performer target = %q, want Practitioner", targets["performer"].TargetType)
	}
}

func TestIncludeRegistry_MultipleRevIncludes(t *testing.T) {
	registry := NewIncludeRegistry()

	registry.RegisterReference("Observation", "subject", "Patient")
	registry.RegisterReference("Condition", "subject", "Patient")

	revTargets := registry.GetRevIncludeTargets("Patient")
	if len(revTargets) != 2 {
		t.Fatalf("expected 2 reverse include targets, got %d", len(revTargets))
	}

	// Check that both source types are present
	sources := make(map[string]bool)
	for _, rt := range revTargets {
		sources[rt.SourceType] = true
	}
	if !sources["Observation"] {
		t.Error("expected Observation in rev include sources")
	}
	if !sources["Condition"] {
		t.Error("expected Condition in rev include sources")
	}
}

func TestResolveIncludes_NoFetcherForTargetType(t *testing.T) {
	registry := NewIncludeRegistry()

	// Register reference but NOT the fetcher for Patient
	registry.RegisterReference("Observation", "subject", "Patient")

	resources := []interface{}{
		map[string]interface{}{
			"resourceType": "Observation",
			"id":           "obs-1",
			"subject":      map[string]interface{}{"reference": "Patient/pat-1"},
		},
	}

	entries, err := registry.ResolveIncludes(context.Background(), resources, []string{"Observation:subject"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected 0 entries when no fetcher is registered, got %d", len(entries))
	}
}

func TestResolveIncludes_EmptyIncludeParams(t *testing.T) {
	registry := NewIncludeRegistry()
	registry.RegisterFetcher("Patient", func(ctx context.Context, fhirID string) (map[string]interface{}, error) {
		return map[string]interface{}{"id": fhirID}, nil
	})

	resources := []interface{}{
		map[string]interface{}{"id": "obs-1"},
	}

	entries, err := registry.ResolveIncludes(context.Background(), resources, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if entries != nil {
		t.Errorf("expected nil entries for empty include params, got %v", entries)
	}
}

func TestSearchParamToFields_AllMappings(t *testing.T) {
	knownParams := []string{"subject", "patient", "encounter", "performer", "requester",
		"recorder", "asserter", "author", "practitioner", "organization", "location"}
	for _, param := range knownParams {
		fields := searchParamToFields(param)
		if len(fields) == 0 {
			t.Errorf("expected non-empty fields for param %q", param)
		}
	}

	// Organization should map to two fields
	orgFields := searchParamToFields("organization")
	if len(orgFields) != 2 {
		t.Errorf("expected 2 fields for 'organization', got %d", len(orgFields))
	}
	if orgFields[0] != "managingOrganization" {
		t.Errorf("orgFields[0] = %q, want 'managingOrganization'", orgFields[0])
	}
}

func TestExtractReferenceID_ReferenceValueNotString(t *testing.T) {
	// The reference field value inside the map is not a string
	resource := map[string]interface{}{
		"subject": map[string]interface{}{
			"reference": 12345,
		},
	}
	id := extractReferenceID(resource, "subject", "Patient")
	if id != "" {
		t.Errorf("expected empty string when reference value is not string, got %q", id)
	}
}
