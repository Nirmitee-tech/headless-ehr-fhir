package fhir

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"

	"github.com/labstack/echo/v4"
)

// ===========================================================================
// ValidationProfile Type Tests
// ===========================================================================

func TestNewValidationProfileRegistry(t *testing.T) {
	reg := NewValidationProfileRegistry()
	if reg == nil {
		t.Fatal("expected non-nil registry")
	}
	profiles := reg.ListValidationProfiles()
	if len(profiles) != 0 {
		t.Errorf("expected 0 profiles in new registry, got %d", len(profiles))
	}
}

func TestRegisterValidationProfile(t *testing.T) {
	reg := NewValidationProfileRegistry()
	profile := &ValidationProfile{
		URL:          "http://example.com/SD/test",
		Name:         "TestProfile",
		ResourceType: "Patient",
		Required:     true,
		Version:      "1.0.0",
	}
	err := reg.RegisterValidationProfile(profile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, ok := reg.GetValidationProfile("http://example.com/SD/test")
	if !ok {
		t.Fatal("expected to find registered profile")
	}
	if got.Name != "TestProfile" {
		t.Errorf("expected Name=TestProfile, got %s", got.Name)
	}
	if got.ResourceType != "Patient" {
		t.Errorf("expected ResourceType=Patient, got %s", got.ResourceType)
	}
}

func TestRegisterValidationProfile_MissingURL(t *testing.T) {
	reg := NewValidationProfileRegistry()
	err := reg.RegisterValidationProfile(&ValidationProfile{
		Name:         "NoURL",
		ResourceType: "Patient",
	})
	if err == nil {
		t.Fatal("expected error for missing URL")
	}
}

func TestRegisterValidationProfile_MissingResourceType(t *testing.T) {
	reg := NewValidationProfileRegistry()
	err := reg.RegisterValidationProfile(&ValidationProfile{
		URL:  "http://example.com/SD/test",
		Name: "NoType",
	})
	if err == nil {
		t.Fatal("expected error for missing ResourceType")
	}
}

func TestRegisterValidationProfile_Overwrite(t *testing.T) {
	reg := NewValidationProfileRegistry()
	p1 := &ValidationProfile{
		URL:          "http://example.com/SD/test",
		Name:         "First",
		ResourceType: "Patient",
	}
	p2 := &ValidationProfile{
		URL:          "http://example.com/SD/test",
		Name:         "Second",
		ResourceType: "Patient",
	}
	_ = reg.RegisterValidationProfile(p1)
	_ = reg.RegisterValidationProfile(p2)

	got, ok := reg.GetValidationProfile("http://example.com/SD/test")
	if !ok {
		t.Fatal("expected to find profile")
	}
	if got.Name != "Second" {
		t.Errorf("expected Second, got %s", got.Name)
	}
}

func TestUnregisterValidationProfile(t *testing.T) {
	reg := NewValidationProfileRegistry()
	_ = reg.RegisterValidationProfile(&ValidationProfile{
		URL:          "http://example.com/SD/test",
		Name:         "ToRemove",
		ResourceType: "Patient",
	})

	err := reg.UnregisterValidationProfile("http://example.com/SD/test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, ok := reg.GetValidationProfile("http://example.com/SD/test")
	if ok {
		t.Error("expected profile to be removed")
	}
}

func TestUnregisterValidationProfile_NotFound(t *testing.T) {
	reg := NewValidationProfileRegistry()
	err := reg.UnregisterValidationProfile("http://example.com/nonexistent")
	if err == nil {
		t.Fatal("expected error for unregistering nonexistent profile")
	}
}

func TestGetValidationProfile(t *testing.T) {
	reg := NewValidationProfileRegistry()
	_ = reg.RegisterValidationProfile(&ValidationProfile{
		URL:          "http://example.com/SD/test",
		Name:         "Test",
		ResourceType: "Patient",
	})

	got, ok := reg.GetValidationProfile("http://example.com/SD/test")
	if !ok {
		t.Fatal("expected to find profile")
	}
	if got.URL != "http://example.com/SD/test" {
		t.Errorf("unexpected URL: %s", got.URL)
	}
}

func TestGetValidationProfile_NotFound(t *testing.T) {
	reg := NewValidationProfileRegistry()
	_, ok := reg.GetValidationProfile("http://example.com/nonexistent")
	if ok {
		t.Error("expected not found")
	}
}

func TestListValidationProfiles(t *testing.T) {
	reg := NewValidationProfileRegistry()
	_ = reg.RegisterValidationProfile(&ValidationProfile{URL: "http://a.com/1", Name: "A", ResourceType: "Patient"})
	_ = reg.RegisterValidationProfile(&ValidationProfile{URL: "http://a.com/2", Name: "B", ResourceType: "Observation"})
	_ = reg.RegisterValidationProfile(&ValidationProfile{URL: "http://a.com/3", Name: "C", ResourceType: "Condition"})

	all := reg.ListValidationProfiles()
	if len(all) != 3 {
		t.Errorf("expected 3 profiles, got %d", len(all))
	}
}

func TestListValidationProfilesForResourceType(t *testing.T) {
	reg := NewValidationProfileRegistry()
	_ = reg.RegisterValidationProfile(&ValidationProfile{URL: "http://a.com/1", Name: "P1", ResourceType: "Patient"})
	_ = reg.RegisterValidationProfile(&ValidationProfile{URL: "http://a.com/2", Name: "P2", ResourceType: "Patient"})
	_ = reg.RegisterValidationProfile(&ValidationProfile{URL: "http://a.com/3", Name: "O1", ResourceType: "Observation"})

	patients := reg.ListValidationProfilesForResourceType("Patient")
	if len(patients) != 2 {
		t.Errorf("expected 2 Patient profiles, got %d", len(patients))
	}

	obs := reg.ListValidationProfilesForResourceType("Observation")
	if len(obs) != 1 {
		t.Errorf("expected 1 Observation profile, got %d", len(obs))
	}

	empty := reg.ListValidationProfilesForResourceType("Unknown")
	if len(empty) != 0 {
		t.Errorf("expected 0 profiles for unknown type, got %d", len(empty))
	}
}

// ===========================================================================
// Default Profile Tests
// ===========================================================================

func TestSetDefaultValidationProfiles(t *testing.T) {
	reg := NewValidationProfileRegistry()
	urls := []string{"http://a.com/1", "http://a.com/2"}
	reg.SetDefaultValidationProfiles("Patient", urls)

	got := reg.GetDefaultValidationProfiles("Patient")
	if len(got) != 2 {
		t.Errorf("expected 2 defaults, got %d", len(got))
	}
	if got[0] != "http://a.com/1" {
		t.Errorf("expected first default http://a.com/1, got %s", got[0])
	}
}

func TestGetDefaultValidationProfiles_Empty(t *testing.T) {
	reg := NewValidationProfileRegistry()
	got := reg.GetDefaultValidationProfiles("Patient")
	if got != nil {
		t.Errorf("expected nil for no defaults, got %v", got)
	}
}

func TestSetDefaultValidationProfiles_Overwrite(t *testing.T) {
	reg := NewValidationProfileRegistry()
	reg.SetDefaultValidationProfiles("Patient", []string{"http://a.com/1"})
	reg.SetDefaultValidationProfiles("Patient", []string{"http://b.com/2"})

	got := reg.GetDefaultValidationProfiles("Patient")
	if len(got) != 1 || got[0] != "http://b.com/2" {
		t.Errorf("expected overwritten defaults, got %v", got)
	}
}

// ===========================================================================
// Cardinality Validation Tests
// ===========================================================================

func TestValidateCardinality_MinSatisfied(t *testing.T) {
	resource := map[string]interface{}{
		"name": []interface{}{map[string]interface{}{"family": "Smith"}},
	}
	constraint := &ElementConstraint{
		Path: "Patient.name",
		Min:  1,
		Max:  "*",
	}
	issues := ValidateCardinality(resource, "name", constraint)
	if len(issues) != 0 {
		t.Errorf("expected no issues, got %d: %v", len(issues), issues)
	}
}

func TestValidateCardinality_MinViolated(t *testing.T) {
	resource := map[string]interface{}{}
	constraint := &ElementConstraint{
		Path: "Patient.name",
		Min:  1,
		Max:  "*",
	}
	issues := ValidateCardinality(resource, "name", constraint)
	if len(issues) != 1 {
		t.Fatalf("expected 1 issue, got %d", len(issues))
	}
	if issues[0].Severity != SeverityError {
		t.Errorf("expected error severity, got %s", issues[0].Severity)
	}
}

func TestValidateCardinality_MinViolated_EmptyArray(t *testing.T) {
	resource := map[string]interface{}{
		"name": []interface{}{},
	}
	constraint := &ElementConstraint{
		Path: "Patient.name",
		Min:  1,
		Max:  "*",
	}
	issues := ValidateCardinality(resource, "name", constraint)
	if len(issues) != 1 {
		t.Fatalf("expected 1 issue, got %d", len(issues))
	}
}

func TestValidateCardinality_MaxSatisfied(t *testing.T) {
	resource := map[string]interface{}{
		"gender": "male",
	}
	constraint := &ElementConstraint{
		Path: "Patient.gender",
		Min:  0,
		Max:  "1",
	}
	issues := ValidateCardinality(resource, "gender", constraint)
	if len(issues) != 0 {
		t.Errorf("expected no issues, got %d", len(issues))
	}
}

func TestValidateCardinality_MaxViolated(t *testing.T) {
	resource := map[string]interface{}{
		"name": []interface{}{
			map[string]interface{}{"family": "A"},
			map[string]interface{}{"family": "B"},
			map[string]interface{}{"family": "C"},
		},
	}
	constraint := &ElementConstraint{
		Path: "Patient.name",
		Min:  0,
		Max:  "2",
	}
	issues := ValidateCardinality(resource, "name", constraint)
	if len(issues) != 1 {
		t.Fatalf("expected 1 issue for max violation, got %d", len(issues))
	}
}

func TestValidateCardinality_Unbounded(t *testing.T) {
	items := make([]interface{}, 100)
	for i := range items {
		items[i] = map[string]interface{}{"value": i}
	}
	resource := map[string]interface{}{
		"identifier": items,
	}
	constraint := &ElementConstraint{
		Path: "Patient.identifier",
		Min:  0,
		Max:  "*",
	}
	issues := ValidateCardinality(resource, "identifier", constraint)
	if len(issues) != 0 {
		t.Errorf("expected no issues for unbounded max, got %d", len(issues))
	}
}

func TestValidateCardinality_ProhibitedPresent(t *testing.T) {
	resource := map[string]interface{}{
		"modifierExtension": []interface{}{map[string]interface{}{"url": "http://x"}},
	}
	constraint := &ElementConstraint{
		Path: "Patient.modifierExtension",
		Min:  0,
		Max:  "0",
	}
	issues := ValidateCardinality(resource, "modifierExtension", constraint)
	if len(issues) != 1 {
		t.Fatalf("expected 1 issue for prohibited element, got %d", len(issues))
	}
}

func TestValidateCardinality_ProhibitedAbsent(t *testing.T) {
	resource := map[string]interface{}{}
	constraint := &ElementConstraint{
		Path: "Patient.modifierExtension",
		Min:  0,
		Max:  "0",
	}
	issues := ValidateCardinality(resource, "modifierExtension", constraint)
	if len(issues) != 0 {
		t.Errorf("expected no issues for prohibited absent element, got %d", len(issues))
	}
}

// ===========================================================================
// Value Set Binding Validation Tests
// ===========================================================================

func TestValidateBinding_RequiredSatisfied(t *testing.T) {
	binding := &ValueSetBinding{
		Strength: "required",
		ValueSet: "http://hl7.org/fhir/ValueSet/administrative-gender",
	}
	issues := ValidateBinding("male", binding)
	if len(issues) != 0 {
		t.Errorf("expected no issues, got %d", len(issues))
	}
}

func TestValidateBinding_RequiredViolated(t *testing.T) {
	binding := &ValueSetBinding{
		Strength: "required",
		ValueSet: "http://hl7.org/fhir/ValueSet/administrative-gender",
	}
	issues := ValidateBinding("invalid-gender", binding)
	if len(issues) != 1 {
		t.Fatalf("expected 1 issue, got %d", len(issues))
	}
	if issues[0].Severity != SeverityError {
		t.Errorf("expected error severity, got %s", issues[0].Severity)
	}
}

func TestValidateBinding_RequiredCodeableConcept(t *testing.T) {
	binding := &ValueSetBinding{
		Strength: "required",
		ValueSet: "http://hl7.org/fhir/ValueSet/observation-status",
	}
	value := map[string]interface{}{
		"coding": []interface{}{
			map[string]interface{}{
				"system": "http://hl7.org/fhir/observation-status",
				"code":   "final",
			},
		},
	}
	issues := ValidateBinding(value, binding)
	if len(issues) != 0 {
		t.Errorf("expected no issues for valid codeable concept binding, got %d", len(issues))
	}
}

func TestValidateBinding_Extensible(t *testing.T) {
	binding := &ValueSetBinding{
		Strength: "extensible",
		ValueSet: "http://hl7.org/fhir/ValueSet/some-set",
	}
	// Extensible bindings should produce warnings, not errors
	issues := ValidateBinding("unknown-code", binding)
	for _, issue := range issues {
		if issue.Severity == SeverityError {
			t.Errorf("extensible binding should not produce errors, got error: %s", issue.Diagnostics)
		}
	}
}

func TestValidateBinding_Preferred(t *testing.T) {
	binding := &ValueSetBinding{
		Strength: "preferred",
		ValueSet: "http://hl7.org/fhir/ValueSet/some-set",
	}
	// Preferred bindings should not produce errors
	issues := ValidateBinding("any-value", binding)
	for _, issue := range issues {
		if issue.Severity == SeverityError {
			t.Errorf("preferred binding should not produce errors")
		}
	}
}

func TestValidateBinding_Example(t *testing.T) {
	binding := &ValueSetBinding{
		Strength: "example",
		ValueSet: "http://hl7.org/fhir/ValueSet/some-set",
	}
	issues := ValidateBinding("any-value", binding)
	if len(issues) != 0 {
		t.Errorf("example binding should produce no issues, got %d", len(issues))
	}
}

func TestValidateBinding_NilBinding(t *testing.T) {
	issues := ValidateBinding("value", nil)
	if len(issues) != 0 {
		t.Errorf("nil binding should produce no issues, got %d", len(issues))
	}
}

// ===========================================================================
// Fixed Value Validation Tests
// ===========================================================================

func TestValidateFixed_Match(t *testing.T) {
	if !ValidateFixed("hello", "hello") {
		t.Error("expected match for identical strings")
	}
}

func TestValidateFixed_Mismatch(t *testing.T) {
	if ValidateFixed("hello", "world") {
		t.Error("expected mismatch for different strings")
	}
}

func TestValidateFixed_NumericMatch(t *testing.T) {
	if !ValidateFixed(42.0, 42.0) {
		t.Error("expected match for identical numbers")
	}
}

func TestValidateFixed_NumericMismatch(t *testing.T) {
	if ValidateFixed(42.0, 43.0) {
		t.Error("expected mismatch for different numbers")
	}
}

func TestValidateFixed_BoolMatch(t *testing.T) {
	if !ValidateFixed(true, true) {
		t.Error("expected match for identical booleans")
	}
}

func TestValidateFixed_BoolMismatch(t *testing.T) {
	if ValidateFixed(true, false) {
		t.Error("expected mismatch for different booleans")
	}
}

func TestValidateFixed_NilMatch(t *testing.T) {
	if !ValidateFixed(nil, nil) {
		t.Error("expected match for nil/nil")
	}
}

func TestValidateFixed_NilMismatch(t *testing.T) {
	if ValidateFixed(nil, "hello") {
		t.Error("expected mismatch for nil vs string")
	}
}

// ===========================================================================
// Pattern Validation Tests
// ===========================================================================

func TestValidatePattern_MapMatch(t *testing.T) {
	value := map[string]interface{}{
		"system": "http://loinc.org",
		"code":   "8480-6",
		"display": "Systolic blood pressure",
	}
	pattern := map[string]interface{}{
		"system": "http://loinc.org",
		"code":   "8480-6",
	}
	if !ValidatePattern(value, pattern) {
		t.Error("expected pattern match: value contains all pattern keys")
	}
}

func TestValidatePattern_MapMismatch(t *testing.T) {
	value := map[string]interface{}{
		"system": "http://loinc.org",
		"code":   "8480-6",
	}
	pattern := map[string]interface{}{
		"system": "http://snomed.info/sct",
		"code":   "8480-6",
	}
	if ValidatePattern(value, pattern) {
		t.Error("expected pattern mismatch: system differs")
	}
}

func TestValidatePattern_SimpleMatch(t *testing.T) {
	if !ValidatePattern("hello", "hello") {
		t.Error("expected pattern match for identical strings")
	}
}

func TestValidatePattern_SimpleMismatch(t *testing.T) {
	if ValidatePattern("hello", "world") {
		t.Error("expected pattern mismatch for different strings")
	}
}

func TestValidatePattern_ArrayContainsPattern(t *testing.T) {
	value := []interface{}{
		map[string]interface{}{"system": "http://loinc.org", "code": "8480-6"},
		map[string]interface{}{"system": "http://snomed.info/sct", "code": "271649006"},
	}
	pattern := map[string]interface{}{
		"system": "http://loinc.org",
		"code":   "8480-6",
	}
	if !ValidatePattern(value, pattern) {
		t.Error("expected pattern match: array contains matching element")
	}
}

func TestValidatePattern_ArrayMissingPattern(t *testing.T) {
	value := []interface{}{
		map[string]interface{}{"system": "http://snomed.info/sct", "code": "271649006"},
	}
	pattern := map[string]interface{}{
		"system": "http://loinc.org",
		"code":   "8480-6",
	}
	if ValidatePattern(value, pattern) {
		t.Error("expected pattern mismatch: array does not contain matching element")
	}
}

func TestValidatePattern_NestedMapMatch(t *testing.T) {
	value := map[string]interface{}{
		"coding": []interface{}{
			map[string]interface{}{"system": "http://loinc.org", "code": "8480-6"},
		},
	}
	pattern := map[string]interface{}{
		"coding": []interface{}{
			map[string]interface{}{"system": "http://loinc.org"},
		},
	}
	if !ValidatePattern(value, pattern) {
		t.Error("expected nested pattern match")
	}
}

// ===========================================================================
// MustSupport Validation Tests
// ===========================================================================

func TestValidateMustSupport_AllPresent(t *testing.T) {
	resource := map[string]interface{}{
		"birthDate": "1990-01-01",
		"gender":    "male",
		"address":   []interface{}{map[string]interface{}{"city": "Boston"}},
	}
	elements := map[string]*ElementConstraint{
		"birthDate": {Path: "Patient.birthDate", MustSupport: true},
		"gender":    {Path: "Patient.gender", MustSupport: true},
		"address":   {Path: "Patient.address", MustSupport: true},
	}
	issues := ValidateMustSupport(resource, elements)
	if len(issues) != 0 {
		t.Errorf("expected no issues, got %d: %v", len(issues), issues)
	}
}

func TestValidateMustSupport_SomeMissing(t *testing.T) {
	resource := map[string]interface{}{
		"birthDate": "1990-01-01",
	}
	elements := map[string]*ElementConstraint{
		"birthDate": {Path: "Patient.birthDate", MustSupport: true},
		"gender":    {Path: "Patient.gender", MustSupport: true},
		"address":   {Path: "Patient.address", MustSupport: true},
	}
	issues := ValidateMustSupport(resource, elements)
	if len(issues) != 2 {
		t.Errorf("expected 2 issues for 2 missing must-support elements, got %d", len(issues))
	}
	for _, issue := range issues {
		if issue.Severity != SeverityWarning {
			t.Errorf("must-support issues should be warnings, got %s", issue.Severity)
		}
	}
}

func TestValidateMustSupport_NonMustSupportIgnored(t *testing.T) {
	resource := map[string]interface{}{}
	elements := map[string]*ElementConstraint{
		"birthDate": {Path: "Patient.birthDate", MustSupport: false},
	}
	issues := ValidateMustSupport(resource, elements)
	if len(issues) != 0 {
		t.Errorf("expected no issues for non-must-support elements, got %d", len(issues))
	}
}

// ===========================================================================
// Extension Validation Tests
// ===========================================================================

func TestValidateExtensions_RequiredPresent(t *testing.T) {
	resource := map[string]interface{}{
		"extension": []interface{}{
			map[string]interface{}{
				"url":         "http://example.com/ext/race",
				"valueString": "Caucasian",
			},
		},
	}
	extensions := []ExtensionDefinition{
		{URL: "http://example.com/ext/race", Required: true, ValueType: "valueString"},
	}
	issues := ValidateExtensions(resource, extensions)
	if len(issues) != 0 {
		t.Errorf("expected no issues, got %d: %v", len(issues), issues)
	}
}

func TestValidateExtensions_RequiredMissing(t *testing.T) {
	resource := map[string]interface{}{
		"extension": []interface{}{},
	}
	extensions := []ExtensionDefinition{
		{URL: "http://example.com/ext/race", Required: true, ValueType: "valueString"},
	}
	issues := ValidateExtensions(resource, extensions)
	if len(issues) != 1 {
		t.Fatalf("expected 1 issue, got %d", len(issues))
	}
	if issues[0].Severity != SeverityError {
		t.Errorf("expected error severity, got %s", issues[0].Severity)
	}
}

func TestValidateExtensions_RequiredMissing_NoExtensionField(t *testing.T) {
	resource := map[string]interface{}{}
	extensions := []ExtensionDefinition{
		{URL: "http://example.com/ext/race", Required: true, ValueType: "valueString"},
	}
	issues := ValidateExtensions(resource, extensions)
	if len(issues) != 1 {
		t.Fatalf("expected 1 issue, got %d", len(issues))
	}
}

func TestValidateExtensions_ValidType(t *testing.T) {
	resource := map[string]interface{}{
		"extension": []interface{}{
			map[string]interface{}{
				"url":         "http://example.com/ext/active",
				"valueBoolean": true,
			},
		},
	}
	extensions := []ExtensionDefinition{
		{URL: "http://example.com/ext/active", Required: false, ValueType: "valueBoolean"},
	}
	issues := ValidateExtensions(resource, extensions)
	if len(issues) != 0 {
		t.Errorf("expected no issues, got %d", len(issues))
	}
}

func TestValidateExtensions_InvalidType(t *testing.T) {
	resource := map[string]interface{}{
		"extension": []interface{}{
			map[string]interface{}{
				"url":         "http://example.com/ext/active",
				"valueString": "true", // wrong type, should be valueBoolean
			},
		},
	}
	extensions := []ExtensionDefinition{
		{URL: "http://example.com/ext/active", Required: false, ValueType: "valueBoolean"},
	}
	issues := ValidateExtensions(resource, extensions)
	if len(issues) != 1 {
		t.Fatalf("expected 1 issue for wrong value type, got %d", len(issues))
	}
}

func TestValidateExtensions_OptionalMissing(t *testing.T) {
	resource := map[string]interface{}{}
	extensions := []ExtensionDefinition{
		{URL: "http://example.com/ext/optional", Required: false, ValueType: "valueString"},
	}
	issues := ValidateExtensions(resource, extensions)
	if len(issues) != 0 {
		t.Errorf("expected no issues for missing optional extension, got %d", len(issues))
	}
}

// ===========================================================================
// Slicing Validation Tests
// ===========================================================================

func TestValidateSlicing_ClosedValid(t *testing.T) {
	values := []interface{}{
		map[string]interface{}{"system": "http://loinc.org", "code": "8480-6"},
		map[string]interface{}{"system": "http://snomed.info/sct", "code": "271649006"},
	}
	rules := &SlicingRules{
		Discriminator: []SlicingDiscriminator{
			{Type: "value", Path: "system"},
		},
		Rules:   "closed",
		Ordered: false,
	}
	issues := ValidateSlicing(values, rules)
	if len(issues) != 0 {
		t.Errorf("expected no issues for valid closed slicing, got %d: %v", len(issues), issues)
	}
}

func TestValidateSlicing_ClosedDuplicateDiscriminator(t *testing.T) {
	values := []interface{}{
		map[string]interface{}{"system": "http://loinc.org", "code": "8480-6"},
		map[string]interface{}{"system": "http://loinc.org", "code": "8462-4"},
	}
	rules := &SlicingRules{
		Discriminator: []SlicingDiscriminator{
			{Type: "value", Path: "system"},
		},
		Rules:   "closed",
		Ordered: false,
	}
	issues := ValidateSlicing(values, rules)
	if len(issues) != 1 {
		t.Fatalf("expected 1 issue for duplicate discriminator in closed slicing, got %d", len(issues))
	}
}

func TestValidateSlicing_Open(t *testing.T) {
	values := []interface{}{
		map[string]interface{}{"system": "http://loinc.org", "code": "8480-6"},
		map[string]interface{}{"system": "http://loinc.org", "code": "8462-4"},
	}
	rules := &SlicingRules{
		Discriminator: []SlicingDiscriminator{
			{Type: "value", Path: "system"},
		},
		Rules:   "open",
		Ordered: false,
	}
	// Open slicing allows duplicates
	issues := ValidateSlicing(values, rules)
	if len(issues) != 0 {
		t.Errorf("expected no issues for open slicing, got %d", len(issues))
	}
}

func TestValidateSlicing_OrderedViolation(t *testing.T) {
	values := []interface{}{
		map[string]interface{}{"system": "http://snomed.info/sct", "code": "271649006"},
		map[string]interface{}{"system": "http://loinc.org", "code": "8480-6"},
	}
	rules := &SlicingRules{
		Discriminator: []SlicingDiscriminator{
			{Type: "value", Path: "system"},
		},
		Rules:   "closed",
		Ordered: true,
	}
	// Ordered slicing: discriminator values must be sorted
	issues := ValidateSlicing(values, rules)
	// At minimum we should get the order violation
	hasOrderIssue := false
	for _, issue := range issues {
		if strings.Contains(issue.Diagnostics, "order") {
			hasOrderIssue = true
		}
	}
	if !hasOrderIssue {
		t.Error("expected order violation issue for unordered slicing")
	}
}

func TestValidateSlicing_NilRules(t *testing.T) {
	values := []interface{}{
		map[string]interface{}{"system": "http://loinc.org"},
	}
	issues := ValidateSlicing(values, nil)
	if len(issues) != 0 {
		t.Errorf("expected no issues for nil rules, got %d", len(issues))
	}
}

// ===========================================================================
// ExtractProfiles Tests
// ===========================================================================

func TestExtractProfiles_Single(t *testing.T) {
	resource := map[string]interface{}{
		"resourceType": "Patient",
		"meta": map[string]interface{}{
			"profile": []interface{}{"http://hl7.org/fhir/us/core/StructureDefinition/us-core-patient"},
		},
	}
	profiles := ExtractProfiles(resource)
	if len(profiles) != 1 {
		t.Fatalf("expected 1 profile, got %d", len(profiles))
	}
	if profiles[0] != "http://hl7.org/fhir/us/core/StructureDefinition/us-core-patient" {
		t.Errorf("unexpected profile URL: %s", profiles[0])
	}
}

func TestExtractProfiles_Multiple(t *testing.T) {
	resource := map[string]interface{}{
		"resourceType": "Patient",
		"meta": map[string]interface{}{
			"profile": []interface{}{
				"http://example.com/SD/profile1",
				"http://example.com/SD/profile2",
			},
		},
	}
	profiles := ExtractProfiles(resource)
	if len(profiles) != 2 {
		t.Errorf("expected 2 profiles, got %d", len(profiles))
	}
}

func TestExtractProfiles_None(t *testing.T) {
	resource := map[string]interface{}{
		"resourceType": "Patient",
	}
	profiles := ExtractProfiles(resource)
	if len(profiles) != 0 {
		t.Errorf("expected 0 profiles, got %d", len(profiles))
	}
}

func TestExtractProfiles_NoMeta(t *testing.T) {
	resource := map[string]interface{}{
		"resourceType": "Patient",
		"name":         []interface{}{map[string]interface{}{"family": "Smith"}},
	}
	profiles := ExtractProfiles(resource)
	if len(profiles) != 0 {
		t.Errorf("expected 0 profiles, got %d", len(profiles))
	}
}

func TestExtractProfiles_InvalidMetaType(t *testing.T) {
	resource := map[string]interface{}{
		"resourceType": "Patient",
		"meta":         "invalid",
	}
	profiles := ExtractProfiles(resource)
	if len(profiles) != 0 {
		t.Errorf("expected 0 profiles for invalid meta type, got %d", len(profiles))
	}
}

func TestExtractProfiles_InvalidProfileType(t *testing.T) {
	resource := map[string]interface{}{
		"resourceType": "Patient",
		"meta": map[string]interface{}{
			"profile": "not-an-array",
		},
	}
	profiles := ExtractProfiles(resource)
	if len(profiles) != 0 {
		t.Errorf("expected 0 profiles for non-array profile, got %d", len(profiles))
	}
}

func TestExtractProfiles_MixedTypes(t *testing.T) {
	resource := map[string]interface{}{
		"resourceType": "Patient",
		"meta": map[string]interface{}{
			"profile": []interface{}{
				"http://example.com/SD/profile1",
				42, // invalid, should be skipped
				"http://example.com/SD/profile2",
			},
		},
	}
	profiles := ExtractProfiles(resource)
	if len(profiles) != 2 {
		t.Errorf("expected 2 valid profiles (skipping non-string), got %d", len(profiles))
	}
}

// ===========================================================================
// ValidateAgainstProfile Tests
// ===========================================================================

func TestValidateAgainstProfile_Valid(t *testing.T) {
	profile := &ValidationProfile{
		URL:          "http://example.com/SD/patient",
		Name:         "TestPatient",
		ResourceType: "Patient",
		Elements: map[string]*ElementConstraint{
			"name": {Path: "Patient.name", Min: 1, Max: "*"},
		},
	}
	resource := map[string]interface{}{
		"resourceType": "Patient",
		"name":         []interface{}{map[string]interface{}{"family": "Smith"}},
	}
	result := ValidateAgainstProfile(resource, profile)
	if !result.Valid {
		t.Errorf("expected valid result, got invalid: %v", result.Issues)
	}
	if result.ProfileURL != profile.URL {
		t.Errorf("expected ProfileURL=%s, got %s", profile.URL, result.ProfileURL)
	}
}

func TestValidateAgainstProfile_MissingRequired(t *testing.T) {
	profile := &ValidationProfile{
		URL:          "http://example.com/SD/patient",
		Name:         "TestPatient",
		ResourceType: "Patient",
		Elements: map[string]*ElementConstraint{
			"name":   {Path: "Patient.name", Min: 1, Max: "*"},
			"gender": {Path: "Patient.gender", Min: 1, Max: "1"},
		},
	}
	resource := map[string]interface{}{
		"resourceType": "Patient",
		"name":         []interface{}{map[string]interface{}{"family": "Smith"}},
		// gender is missing
	}
	result := ValidateAgainstProfile(resource, profile)
	if result.Valid {
		t.Error("expected invalid result for missing required element")
	}
	if len(result.Issues) == 0 {
		t.Error("expected at least one issue")
	}
}

func TestValidateAgainstProfile_InvalidBinding(t *testing.T) {
	profile := &ValidationProfile{
		URL:          "http://example.com/SD/patient",
		Name:         "TestPatient",
		ResourceType: "Patient",
		Elements: map[string]*ElementConstraint{
			"gender": {
				Path: "Patient.gender", Min: 1, Max: "1",
				Binding: &ValueSetBinding{
					Strength: "required",
					ValueSet: "http://hl7.org/fhir/ValueSet/administrative-gender",
				},
			},
		},
	}
	resource := map[string]interface{}{
		"resourceType": "Patient",
		"gender":       "invalid-gender",
	}
	result := ValidateAgainstProfile(resource, profile)
	if result.Valid {
		t.Error("expected invalid result for binding violation")
	}
}

func TestValidateAgainstProfile_NilResource(t *testing.T) {
	profile := &ValidationProfile{
		URL:          "http://example.com/SD/patient",
		Name:         "TestPatient",
		ResourceType: "Patient",
	}
	result := ValidateAgainstProfile(nil, profile)
	if result.Valid {
		t.Error("expected invalid result for nil resource")
	}
}

func TestValidateAgainstProfile_NilProfile(t *testing.T) {
	resource := map[string]interface{}{
		"resourceType": "Patient",
	}
	result := ValidateAgainstProfile(resource, nil)
	if result.Valid {
		t.Error("expected invalid result for nil profile")
	}
}

func TestValidateAgainstProfile_ResourceTypeMismatch(t *testing.T) {
	profile := &ValidationProfile{
		URL:          "http://example.com/SD/patient",
		Name:         "TestPatient",
		ResourceType: "Patient",
	}
	resource := map[string]interface{}{
		"resourceType": "Observation",
	}
	result := ValidateAgainstProfile(resource, profile)
	if result.Valid {
		t.Error("expected invalid result for resource type mismatch")
	}
}

func TestValidateAgainstProfile_FixedValue(t *testing.T) {
	profile := &ValidationProfile{
		URL:          "http://example.com/SD/obs",
		Name:         "TestObs",
		ResourceType: "Observation",
		Elements: map[string]*ElementConstraint{
			"status": {Path: "Observation.status", Min: 1, Max: "1", Fixed: "final"},
		},
	}
	resource := map[string]interface{}{
		"resourceType": "Observation",
		"status":       "final",
	}
	result := ValidateAgainstProfile(resource, profile)
	if !result.Valid {
		t.Errorf("expected valid for matching fixed value: %v", result.Issues)
	}
}

func TestValidateAgainstProfile_FixedValueMismatch(t *testing.T) {
	profile := &ValidationProfile{
		URL:          "http://example.com/SD/obs",
		Name:         "TestObs",
		ResourceType: "Observation",
		Elements: map[string]*ElementConstraint{
			"status": {Path: "Observation.status", Min: 1, Max: "1", Fixed: "final"},
		},
	}
	resource := map[string]interface{}{
		"resourceType": "Observation",
		"status":       "preliminary",
	}
	result := ValidateAgainstProfile(resource, profile)
	if result.Valid {
		t.Error("expected invalid for fixed value mismatch")
	}
}

func TestValidateAgainstProfile_PatternValue(t *testing.T) {
	profile := &ValidationProfile{
		URL:          "http://example.com/SD/obs",
		Name:         "TestObs",
		ResourceType: "Observation",
		Elements: map[string]*ElementConstraint{
			"code": {
				Path: "Observation.code", Min: 1, Max: "1",
				Pattern: map[string]interface{}{
					"coding": []interface{}{
						map[string]interface{}{"system": "http://loinc.org"},
					},
				},
			},
		},
	}
	resource := map[string]interface{}{
		"resourceType": "Observation",
		"code": map[string]interface{}{
			"coding": []interface{}{
				map[string]interface{}{"system": "http://loinc.org", "code": "8480-6"},
			},
		},
	}
	result := ValidateAgainstProfile(resource, profile)
	if !result.Valid {
		t.Errorf("expected valid for pattern match: %v", result.Issues)
	}
}

func TestValidateAgainstProfile_Extensions(t *testing.T) {
	profile := &ValidationProfile{
		URL:          "http://example.com/SD/patient",
		Name:         "TestPatient",
		ResourceType: "Patient",
		Extensions: []ExtensionDefinition{
			{URL: "http://example.com/ext/race", Required: true, ValueType: "valueString"},
		},
	}
	resource := map[string]interface{}{
		"resourceType": "Patient",
		"extension": []interface{}{
			map[string]interface{}{
				"url":         "http://example.com/ext/race",
				"valueString": "test",
			},
		},
	}
	result := ValidateAgainstProfile(resource, profile)
	if !result.Valid {
		t.Errorf("expected valid for present required extension: %v", result.Issues)
	}
}

func TestValidateAgainstProfile_MustSupportWarnings(t *testing.T) {
	profile := &ValidationProfile{
		URL:          "http://example.com/SD/patient",
		Name:         "TestPatient",
		ResourceType: "Patient",
		Elements: map[string]*ElementConstraint{
			"birthDate": {Path: "Patient.birthDate", Min: 0, Max: "1", MustSupport: true},
		},
	}
	resource := map[string]interface{}{
		"resourceType": "Patient",
		// birthDate missing but not required, should get warning
	}
	result := ValidateAgainstProfile(resource, profile)
	// Should be valid (must-support generates warnings, not errors)
	if !result.Valid {
		t.Error("expected valid result (must-support only generates warnings)")
	}
	if len(result.Issues) == 0 {
		t.Error("expected warning issues for missing must-support element")
	}
}

func TestValidateAgainstProfile_EmptyResource(t *testing.T) {
	profile := &ValidationProfile{
		URL:          "http://example.com/SD/patient",
		Name:         "TestPatient",
		ResourceType: "Patient",
		Elements: map[string]*ElementConstraint{
			"name": {Path: "Patient.name", Min: 1, Max: "*"},
		},
	}
	resource := map[string]interface{}{
		"resourceType": "Patient",
	}
	result := ValidateAgainstProfile(resource, profile)
	if result.Valid {
		t.Error("expected invalid for empty resource missing required elements")
	}
}

// ===========================================================================
// Middleware Tests
// ===========================================================================

func TestProfileValidationMiddleware_CreateValid(t *testing.T) {
	reg := NewValidationProfileRegistry()
	_ = reg.RegisterValidationProfile(&ValidationProfile{
		URL:          "http://example.com/SD/patient",
		Name:         "TestPatient",
		ResourceType: "Patient",
		Required:     true,
		Elements: map[string]*ElementConstraint{
			"name": {Path: "Patient.name", Min: 1, Max: "*"},
		},
	})
	reg.SetDefaultValidationProfiles("Patient", []string{"http://example.com/SD/patient"})

	config := &ProfileValidationConfig{
		ValidateOnCreate: true,
		RequiredProfiles: []string{"http://example.com/SD/patient"},
	}

	e := echo.New()
	body := `{"resourceType":"Patient","name":[{"family":"Smith"}]}`
	req := httptest.NewRequest(http.MethodPost, "/fhir/Patient", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/fhir+json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/fhir/:resourceType")
	c.SetParamNames("resourceType")
	c.SetParamValues("Patient")

	handler := ProfileValidationMiddleware(reg, config)(func(c echo.Context) error {
		return c.JSON(http.StatusCreated, map[string]interface{}{"status": "created"})
	})

	err := handler(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestProfileValidationMiddleware_CreateInvalid(t *testing.T) {
	reg := NewValidationProfileRegistry()
	_ = reg.RegisterValidationProfile(&ValidationProfile{
		URL:          "http://example.com/SD/patient",
		Name:         "TestPatient",
		ResourceType: "Patient",
		Required:     true,
		Elements: map[string]*ElementConstraint{
			"name":   {Path: "Patient.name", Min: 1, Max: "*"},
			"gender": {Path: "Patient.gender", Min: 1, Max: "1"},
		},
	})
	reg.SetDefaultValidationProfiles("Patient", []string{"http://example.com/SD/patient"})

	config := &ProfileValidationConfig{
		ValidateOnCreate: true,
		RequiredProfiles: []string{"http://example.com/SD/patient"},
	}

	e := echo.New()
	body := `{"resourceType":"Patient","name":[{"family":"Smith"}]}`
	req := httptest.NewRequest(http.MethodPost, "/fhir/Patient", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/fhir+json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/fhir/:resourceType")
	c.SetParamNames("resourceType")
	c.SetParamValues("Patient")

	handler := ProfileValidationMiddleware(reg, config)(func(c echo.Context) error {
		return c.JSON(http.StatusCreated, map[string]interface{}{"status": "created"})
	})

	_ = handler(c)
	// Should return 422 Unprocessable Entity for invalid resource
	if rec.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected 422, got %d", rec.Code)
	}
}

func TestProfileValidationMiddleware_UpdateValid(t *testing.T) {
	reg := NewValidationProfileRegistry()
	_ = reg.RegisterValidationProfile(&ValidationProfile{
		URL:          "http://example.com/SD/patient",
		Name:         "TestPatient",
		ResourceType: "Patient",
		Required:     true,
		Elements: map[string]*ElementConstraint{
			"name": {Path: "Patient.name", Min: 1, Max: "*"},
		},
	})
	reg.SetDefaultValidationProfiles("Patient", []string{"http://example.com/SD/patient"})

	config := &ProfileValidationConfig{
		ValidateOnUpdate: true,
		RequiredProfiles: []string{"http://example.com/SD/patient"},
	}

	e := echo.New()
	body := `{"resourceType":"Patient","id":"123","name":[{"family":"Smith"}]}`
	req := httptest.NewRequest(http.MethodPut, "/fhir/Patient/123", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/fhir+json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/fhir/:resourceType/:id")
	c.SetParamNames("resourceType", "id")
	c.SetParamValues("Patient", "123")

	handler := ProfileValidationMiddleware(reg, config)(func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]interface{}{"status": "updated"})
	})

	err := handler(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestProfileValidationMiddleware_GETSkipped(t *testing.T) {
	reg := NewValidationProfileRegistry()
	config := &ProfileValidationConfig{
		ValidateOnCreate: true,
		ValidateOnUpdate: true,
	}

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient/123", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := ProfileValidationMiddleware(reg, config)(func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]interface{}{"status": "ok"})
	})

	err := handler(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestProfileValidationMiddleware_IgnoredProfiles(t *testing.T) {
	reg := NewValidationProfileRegistry()
	_ = reg.RegisterValidationProfile(&ValidationProfile{
		URL:          "http://example.com/SD/strict",
		Name:         "Strict",
		ResourceType: "Patient",
		Required:     true,
		Elements: map[string]*ElementConstraint{
			"gender": {Path: "Patient.gender", Min: 1, Max: "1"},
		},
	})
	reg.SetDefaultValidationProfiles("Patient", []string{"http://example.com/SD/strict"})

	config := &ProfileValidationConfig{
		ValidateOnCreate: true,
		IgnoreProfiles:   []string{"http://example.com/SD/strict"},
	}

	e := echo.New()
	body := `{"resourceType":"Patient","name":[{"family":"Smith"}]}`
	req := httptest.NewRequest(http.MethodPost, "/fhir/Patient", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/fhir+json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/fhir/:resourceType")
	c.SetParamNames("resourceType")
	c.SetParamValues("Patient")

	handler := ProfileValidationMiddleware(reg, config)(func(c echo.Context) error {
		return c.JSON(http.StatusCreated, map[string]interface{}{"status": "created"})
	})

	err := handler(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201 (ignored profile), got %d", rec.Code)
	}
}

func TestProfileValidationMiddleware_StrictMode(t *testing.T) {
	reg := NewValidationProfileRegistry()
	_ = reg.RegisterValidationProfile(&ValidationProfile{
		URL:          "http://example.com/SD/patient",
		Name:         "TestPatient",
		ResourceType: "Patient",
		Required:     true,
		Elements: map[string]*ElementConstraint{
			"birthDate": {Path: "Patient.birthDate", Min: 0, Max: "1", MustSupport: true},
		},
	})
	reg.SetDefaultValidationProfiles("Patient", []string{"http://example.com/SD/patient"})

	config := &ProfileValidationConfig{
		ValidateOnCreate: true,
		StrictMode:       true,
		RequiredProfiles: []string{"http://example.com/SD/patient"},
	}

	e := echo.New()
	body := `{"resourceType":"Patient"}`
	req := httptest.NewRequest(http.MethodPost, "/fhir/Patient", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/fhir+json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/fhir/:resourceType")
	c.SetParamNames("resourceType")
	c.SetParamValues("Patient")

	handler := ProfileValidationMiddleware(reg, config)(func(c echo.Context) error {
		return c.JSON(http.StatusCreated, map[string]interface{}{"status": "created"})
	})

	_ = handler(c)
	// Strict mode: warnings become errors, so missing must-support should fail
	if rec.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected 422 in strict mode for missing must-support, got %d", rec.Code)
	}
}

// ===========================================================================
// Handler Tests
// ===========================================================================

func TestProfileValidationHandler_WithProfileParam(t *testing.T) {
	reg := NewValidationProfileRegistry()
	_ = reg.RegisterValidationProfile(&ValidationProfile{
		URL:          "http://example.com/SD/patient",
		Name:         "TestPatient",
		ResourceType: "Patient",
		Elements: map[string]*ElementConstraint{
			"name": {Path: "Patient.name", Min: 1, Max: "*"},
		},
	})

	e := echo.New()
	body := `{"resourceType":"Patient","name":[{"family":"Smith"}]}`
	req := httptest.NewRequest(http.MethodPost, "/fhir/Patient/$validate?profile=http://example.com/SD/patient", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/fhir+json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := ProfileValidationHandler(reg)
	err := handler(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var outcome map[string]interface{}
	_ = json.Unmarshal(rec.Body.Bytes(), &outcome)
	if outcome["resourceType"] != "OperationOutcome" {
		t.Error("expected OperationOutcome response")
	}
}

func TestProfileValidationHandler_ProfileNotFound(t *testing.T) {
	reg := NewValidationProfileRegistry()

	e := echo.New()
	body := `{"resourceType":"Patient","name":[{"family":"Smith"}]}`
	req := httptest.NewRequest(http.MethodPost, "/fhir/Patient/$validate?profile=http://example.com/SD/nonexistent", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/fhir+json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := ProfileValidationHandler(reg)
	err := handler(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should return an OperationOutcome with not-found
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404 for unknown profile, got %d", rec.Code)
	}
}

func TestProfileValidationHandler_NoProfileParam(t *testing.T) {
	reg := NewValidationProfileRegistry()

	e := echo.New()
	body := `{"resourceType":"Patient","name":[{"family":"Smith"}]}`
	req := httptest.NewRequest(http.MethodPost, "/fhir/Patient/$validate", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/fhir+json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := ProfileValidationHandler(reg)
	err := handler(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestProfileValidationHandler_EmptyBody(t *testing.T) {
	reg := NewValidationProfileRegistry()

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/fhir/Patient/$validate", strings.NewReader(""))
	req.Header.Set("Content-Type", "application/fhir+json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := ProfileValidationHandler(reg)
	err := handler(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for empty body, got %d", rec.Code)
	}
}

func TestProfileValidationHandler_InvalidJSON(t *testing.T) {
	reg := NewValidationProfileRegistry()

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/fhir/Patient/$validate", strings.NewReader("{invalid"))
	req.Header.Set("Content-Type", "application/fhir+json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := ProfileValidationHandler(reg)
	err := handler(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid JSON, got %d", rec.Code)
	}
}

func TestProfileValidationHandler_MultipleProfiles(t *testing.T) {
	reg := NewValidationProfileRegistry()
	_ = reg.RegisterValidationProfile(&ValidationProfile{
		URL:          "http://example.com/SD/patient1",
		Name:         "TestPatient1",
		ResourceType: "Patient",
		Elements: map[string]*ElementConstraint{
			"name": {Path: "Patient.name", Min: 1, Max: "*"},
		},
	})
	_ = reg.RegisterValidationProfile(&ValidationProfile{
		URL:          "http://example.com/SD/patient2",
		Name:         "TestPatient2",
		ResourceType: "Patient",
		Elements: map[string]*ElementConstraint{
			"gender": {Path: "Patient.gender", Min: 1, Max: "1"},
		},
	})

	e := echo.New()
	body := `{"resourceType":"Patient","name":[{"family":"Smith"}]}`
	req := httptest.NewRequest(http.MethodPost,
		"/fhir/Patient/$validate?profile=http://example.com/SD/patient1&profile=http://example.com/SD/patient2",
		strings.NewReader(body))
	req.Header.Set("Content-Type", "application/fhir+json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := ProfileValidationHandler(reg)
	err := handler(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Patient satisfies patient1 but not patient2, so outcome should show issues
	var outcome map[string]interface{}
	_ = json.Unmarshal(rec.Body.Bytes(), &outcome)
	issues, _ := outcome["issue"].([]interface{})
	hasError := false
	for _, i := range issues {
		issueMap, _ := i.(map[string]interface{})
		if issueMap["severity"] == "error" {
			hasError = true
		}
	}
	if !hasError {
		t.Error("expected error issues for profile2 validation failure")
	}
}

// ===========================================================================
// ParseValidateProfileParams Tests
// ===========================================================================

func TestParseValidateProfileParams_Single(t *testing.T) {
	params := url.Values{}
	params.Set("profile", "http://example.com/SD/test")
	result := ParseValidateProfileParams(params)
	if len(result) != 1 {
		t.Fatalf("expected 1 profile, got %d", len(result))
	}
	if result[0] != "http://example.com/SD/test" {
		t.Errorf("unexpected profile: %s", result[0])
	}
}

func TestParseValidateProfileParams_Multiple(t *testing.T) {
	params := url.Values{}
	params["profile"] = []string{"http://a.com/1", "http://a.com/2"}
	result := ParseValidateProfileParams(params)
	if len(result) != 2 {
		t.Errorf("expected 2 profiles, got %d", len(result))
	}
}

func TestParseValidateProfileParams_None(t *testing.T) {
	params := url.Values{}
	result := ParseValidateProfileParams(params)
	if len(result) != 0 {
		t.Errorf("expected 0 profiles, got %d", len(result))
	}
}

func TestParseValidateProfileParams_Empty(t *testing.T) {
	params := url.Values{}
	params.Set("profile", "")
	result := ParseValidateProfileParams(params)
	if len(result) != 0 {
		t.Errorf("expected 0 profiles for empty string, got %d", len(result))
	}
}

// ===========================================================================
// Default US Core Profiles Tests
// ===========================================================================

func TestDefaultUSCoreValidationProfiles(t *testing.T) {
	profiles := DefaultUSCoreValidationProfiles()
	if len(profiles) < 3 {
		t.Errorf("expected at least 3 US Core profiles (Patient, Observation, Condition), got %d", len(profiles))
	}

	typeMap := make(map[string]bool)
	for _, p := range profiles {
		typeMap[p.ResourceType] = true
	}

	for _, expected := range []string{"Patient", "Observation", "Condition"} {
		if !typeMap[expected] {
			t.Errorf("expected US Core profile for %s", expected)
		}
	}
}

func TestDefaultUSCoreValidationProfiles_Patient(t *testing.T) {
	profiles := DefaultUSCoreValidationProfiles()
	var patient *ValidationProfile
	for _, p := range profiles {
		if p.ResourceType == "Patient" {
			patient = p
			break
		}
	}
	if patient == nil {
		t.Fatal("expected US Core Patient profile")
	}

	if patient.URL == "" {
		t.Error("expected non-empty URL for US Core Patient")
	}
	if len(patient.Elements) == 0 {
		t.Error("expected elements in US Core Patient profile")
	}

	// US Core Patient requires identifier
	identConst, ok := patient.Elements["identifier"]
	if !ok {
		t.Fatal("expected identifier constraint")
	}
	if identConst.Min != 1 {
		t.Errorf("expected identifier min=1, got %d", identConst.Min)
	}
}

func TestDefaultUSCoreValidationProfiles_Observation(t *testing.T) {
	profiles := DefaultUSCoreValidationProfiles()
	var obs *ValidationProfile
	for _, p := range profiles {
		if p.ResourceType == "Observation" {
			obs = p
			break
		}
	}
	if obs == nil {
		t.Fatal("expected US Core Observation profile")
	}
	if len(obs.Elements) == 0 {
		t.Error("expected elements in US Core Observation profile")
	}

	// Observation requires status, code, subject
	for _, required := range []string{"status", "code", "subject"} {
		ec, ok := obs.Elements[required]
		if !ok {
			t.Errorf("expected %s constraint", required)
			continue
		}
		if ec.Min < 1 {
			t.Errorf("expected %s min >= 1, got %d", required, ec.Min)
		}
	}
}

func TestDefaultUSCoreValidationProfiles_Condition(t *testing.T) {
	profiles := DefaultUSCoreValidationProfiles()
	var cond *ValidationProfile
	for _, p := range profiles {
		if p.ResourceType == "Condition" {
			cond = p
			break
		}
	}
	if cond == nil {
		t.Fatal("expected US Core Condition profile")
	}
	if len(cond.Elements) == 0 {
		t.Error("expected elements in US Core Condition profile")
	}
}

// ===========================================================================
// BuildProfileOperationOutcome Tests
// ===========================================================================

func TestBuildProfileOperationOutcome_Valid(t *testing.T) {
	results := []*ProfileValidationResult{
		{Valid: true, ProfileURL: "http://example.com/SD/test", ProfileName: "Test"},
	}
	outcome := BuildProfileOperationOutcome(results)
	if outcome["resourceType"] != "OperationOutcome" {
		t.Error("expected OperationOutcome resourceType")
	}
	issues, ok := outcome["issue"].([]map[string]interface{})
	if !ok {
		t.Fatal("expected issue array")
	}
	if len(issues) == 0 {
		t.Fatal("expected at least one issue")
	}
	if issues[0]["severity"] != "information" {
		t.Errorf("expected information severity for valid result, got %s", issues[0]["severity"])
	}
}

func TestBuildProfileOperationOutcome_Invalid(t *testing.T) {
	results := []*ProfileValidationResult{
		{
			Valid:       false,
			ProfileURL:  "http://example.com/SD/test",
			ProfileName: "Test",
			Issues: []ValidationIssue{
				{Severity: SeverityError, Code: VIssueTypeRequired, Location: "Patient.name", Diagnostics: "name is required"},
			},
		},
	}
	outcome := BuildProfileOperationOutcome(results)
	issues, ok := outcome["issue"].([]map[string]interface{})
	if !ok {
		t.Fatal("expected issue array")
	}
	hasError := false
	for _, issue := range issues {
		if issue["severity"] == "error" {
			hasError = true
		}
	}
	if !hasError {
		t.Error("expected error issues")
	}
}

func TestBuildProfileOperationOutcome_MultipleResults(t *testing.T) {
	results := []*ProfileValidationResult{
		{Valid: true, ProfileURL: "http://a.com/1", ProfileName: "A"},
		{
			Valid:       false,
			ProfileURL:  "http://a.com/2",
			ProfileName: "B",
			Issues: []ValidationIssue{
				{Severity: SeverityError, Code: VIssueTypeRequired, Diagnostics: "missing field"},
				{Severity: SeverityWarning, Code: VIssueTypeInvariant, Diagnostics: "warning msg"},
			},
		},
	}
	outcome := BuildProfileOperationOutcome(results)
	issues, ok := outcome["issue"].([]map[string]interface{})
	if !ok {
		t.Fatal("expected issue array")
	}
	if len(issues) < 2 {
		t.Errorf("expected at least 2 issues, got %d", len(issues))
	}
}

func TestBuildProfileOperationOutcome_EmptyResults(t *testing.T) {
	results := []*ProfileValidationResult{}
	outcome := BuildProfileOperationOutcome(results)
	if outcome["resourceType"] != "OperationOutcome" {
		t.Error("expected OperationOutcome")
	}
	issues, ok := outcome["issue"].([]map[string]interface{})
	if !ok {
		t.Fatal("expected issue array")
	}
	if len(issues) != 1 || issues[0]["severity"] != "information" {
		t.Error("expected single information issue for empty results")
	}
}

// ===========================================================================
// Thread Safety Test
// ===========================================================================

func TestValidationProfileRegistry_ConcurrentAccess(t *testing.T) {
	reg := NewValidationProfileRegistry()
	var wg sync.WaitGroup

	// Concurrent writes
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			url := "http://example.com/SD/" + strings.Repeat("a", n%10+1)
			_ = reg.RegisterValidationProfile(&ValidationProfile{
				URL:          url,
				Name:         "Test",
				ResourceType: "Patient",
			})
		}(i)
	}

	// Concurrent reads
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = reg.ListValidationProfiles()
			_ = reg.ListValidationProfilesForResourceType("Patient")
			_, _ = reg.GetValidationProfile("http://example.com/SD/aaa")
			_ = reg.GetDefaultValidationProfiles("Patient")
		}()
	}

	wg.Wait()
}

// ===========================================================================
// Edge Case Tests
// ===========================================================================

func TestValidateAgainstProfile_NestedElements(t *testing.T) {
	profile := &ValidationProfile{
		URL:          "http://example.com/SD/patient",
		Name:         "TestPatient",
		ResourceType: "Patient",
		Elements: map[string]*ElementConstraint{
			"name": {Path: "Patient.name", Min: 1, Max: "*"},
		},
	}
	resource := map[string]interface{}{
		"resourceType": "Patient",
		"name": []interface{}{
			map[string]interface{}{
				"family": "Smith",
				"given":  []interface{}{"John"},
			},
		},
	}
	result := ValidateAgainstProfile(resource, profile)
	if !result.Valid {
		t.Errorf("expected valid result for nested resource: %v", result.Issues)
	}
}

func TestValidateAgainstProfile_ConflictingConstraints(t *testing.T) {
	// A profile with a required field that also has a fixed value
	profile := &ValidationProfile{
		URL:          "http://example.com/SD/obs",
		Name:         "TestObs",
		ResourceType: "Observation",
		Elements: map[string]*ElementConstraint{
			"status": {Path: "Observation.status", Min: 1, Max: "1", Fixed: "final"},
		},
	}
	resource := map[string]interface{}{
		"resourceType": "Observation",
		// status is missing - should fail on required
	}
	result := ValidateAgainstProfile(resource, profile)
	if result.Valid {
		t.Error("expected invalid for missing required element with fixed value")
	}
}

func TestValidateCardinality_ScalarMax1_WithScalar(t *testing.T) {
	resource := map[string]interface{}{
		"gender": "male",
	}
	constraint := &ElementConstraint{
		Path: "Patient.gender",
		Min:  1,
		Max:  "1",
	}
	issues := ValidateCardinality(resource, "gender", constraint)
	if len(issues) != 0 {
		t.Errorf("expected no issues for scalar value with max=1, got %d", len(issues))
	}
}

func TestValidateCardinality_MinZero_Absent(t *testing.T) {
	resource := map[string]interface{}{}
	constraint := &ElementConstraint{
		Path: "Patient.birthDate",
		Min:  0,
		Max:  "1",
	}
	issues := ValidateCardinality(resource, "birthDate", constraint)
	if len(issues) != 0 {
		t.Errorf("expected no issues for optional absent element, got %d", len(issues))
	}
}

func TestValidateBinding_AllGenderValues(t *testing.T) {
	binding := &ValueSetBinding{
		Strength: "required",
		ValueSet: "http://hl7.org/fhir/ValueSet/administrative-gender",
	}
	validGenders := []string{"male", "female", "other", "unknown"}
	for _, gender := range validGenders {
		issues := ValidateBinding(gender, binding)
		if len(issues) != 0 {
			t.Errorf("expected %s to be a valid gender, got issues: %v", gender, issues)
		}
	}
}

func TestValidateExtensions_MultipleExtensions(t *testing.T) {
	resource := map[string]interface{}{
		"extension": []interface{}{
			map[string]interface{}{
				"url":         "http://example.com/ext/race",
				"valueString": "test",
			},
			map[string]interface{}{
				"url":          "http://example.com/ext/active",
				"valueBoolean": true,
			},
		},
	}
	extensions := []ExtensionDefinition{
		{URL: "http://example.com/ext/race", Required: true, ValueType: "valueString"},
		{URL: "http://example.com/ext/active", Required: true, ValueType: "valueBoolean"},
	}
	issues := ValidateExtensions(resource, extensions)
	if len(issues) != 0 {
		t.Errorf("expected no issues for all required extensions present, got %d: %v", len(issues), issues)
	}
}

func TestValidateSlicing_EmptyValues(t *testing.T) {
	values := []interface{}{}
	rules := &SlicingRules{
		Discriminator: []SlicingDiscriminator{
			{Type: "value", Path: "system"},
		},
		Rules: "closed",
	}
	issues := ValidateSlicing(values, rules)
	if len(issues) != 0 {
		t.Errorf("expected no issues for empty values, got %d", len(issues))
	}
}

func TestValidateSlicing_ExistsDiscriminator(t *testing.T) {
	values := []interface{}{
		map[string]interface{}{"system": "http://loinc.org", "code": "8480-6"},
		map[string]interface{}{"code": "other"}, // no system field
	}
	rules := &SlicingRules{
		Discriminator: []SlicingDiscriminator{
			{Type: "exists", Path: "system"},
		},
		Rules: "closed",
	}
	// Each item has distinct "exists" for system (true vs false), so should be valid
	issues := ValidateSlicing(values, rules)
	if len(issues) != 0 {
		t.Errorf("expected no issues for valid exists-based slicing, got %d: %v", len(issues), issues)
	}
}

func TestValidateAgainstProfile_SlicingElement(t *testing.T) {
	profile := &ValidationProfile{
		URL:          "http://example.com/SD/obs",
		Name:         "TestObs",
		ResourceType: "Observation",
		Elements: map[string]*ElementConstraint{
			"category": {
				Path: "Observation.category",
				Min:  1,
				Max:  "*",
				Slicing: &SlicingRules{
					Discriminator: []SlicingDiscriminator{
						{Type: "value", Path: "coding.system"},
					},
					Rules: "open",
				},
			},
		},
	}
	resource := map[string]interface{}{
		"resourceType": "Observation",
		"category": []interface{}{
			map[string]interface{}{
				"coding": []interface{}{
					map[string]interface{}{"system": "http://terminology.hl7.org/CodeSystem/observation-category", "code": "laboratory"},
				},
			},
		},
	}
	result := ValidateAgainstProfile(resource, profile)
	if !result.Valid {
		t.Errorf("expected valid result for sliced element: %v", result.Issues)
	}
}
