package hipaa

import (
	"testing"
)

func TestDefaultPHIFields_CoversExpectedResources(t *testing.T) {
	configs := DefaultPHIFields()

	expected := map[string]bool{
		"Patient":       false,
		"Practitioner":  false,
		"RelatedPerson": false,
	}

	for _, c := range configs {
		if _, ok := expected[c.ResourceType]; ok {
			expected[c.ResourceType] = true
		}
	}

	for rt, found := range expected {
		if !found {
			t.Errorf("expected PHI config for resource type %q but it was missing", rt)
		}
	}
}

func TestDefaultPHIFields_PatientFields(t *testing.T) {
	configs := DefaultPHIFields()

	var patientCfg *PHIFieldConfig
	for i := range configs {
		if configs[i].ResourceType == "Patient" {
			patientCfg = &configs[i]
			break
		}
	}

	if patientCfg == nil {
		t.Fatal("Patient PHI config not found")
	}

	requiredFields := []string{
		"identifier.ssn",
		"address.line",
		"telecom.phone",
		"telecom.email",
	}

	fieldSet := make(map[string]bool, len(patientCfg.Fields))
	for _, f := range patientCfg.Fields {
		fieldSet[f] = true
	}

	for _, rf := range requiredFields {
		if !fieldSet[rf] {
			t.Errorf("Patient config missing required PHI field %q", rf)
		}
	}
}

func TestDefaultPHIFields_PractitionerFields(t *testing.T) {
	configs := DefaultPHIFields()

	var cfg *PHIFieldConfig
	for i := range configs {
		if configs[i].ResourceType == "Practitioner" {
			cfg = &configs[i]
			break
		}
	}

	if cfg == nil {
		t.Fatal("Practitioner PHI config not found")
	}

	requiredFields := []string{
		"address.line",
		"telecom.phone",
		"telecom.email",
	}

	fieldSet := make(map[string]bool, len(cfg.Fields))
	for _, f := range cfg.Fields {
		fieldSet[f] = true
	}

	for _, rf := range requiredFields {
		if !fieldSet[rf] {
			t.Errorf("Practitioner config missing required PHI field %q", rf)
		}
	}
}

func TestDefaultPHIFields_RelatedPersonFields(t *testing.T) {
	configs := DefaultPHIFields()

	var cfg *PHIFieldConfig
	for i := range configs {
		if configs[i].ResourceType == "RelatedPerson" {
			cfg = &configs[i]
			break
		}
	}

	if cfg == nil {
		t.Fatal("RelatedPerson PHI config not found")
	}

	requiredFields := []string{
		"address.line",
		"telecom.phone",
	}

	fieldSet := make(map[string]bool, len(cfg.Fields))
	for _, f := range cfg.Fields {
		fieldSet[f] = true
	}

	for _, rf := range requiredFields {
		if !fieldSet[rf] {
			t.Errorf("RelatedPerson config missing required PHI field %q", rf)
		}
	}
}

func TestPHIFieldPaths(t *testing.T) {
	paths := PHIFieldPaths()

	expectedPaths := []string{
		"Patient.identifier.ssn",
		"Patient.address.line",
		"Patient.telecom.phone",
		"Patient.telecom.email",
		"Practitioner.address.line",
		"Practitioner.telecom.phone",
		"Practitioner.telecom.email",
		"RelatedPerson.address.line",
		"RelatedPerson.telecom.phone",
	}

	for _, p := range expectedPaths {
		if !paths[p] {
			t.Errorf("PHIFieldPaths() missing expected path %q", p)
		}
	}

	// Verify total count matches expectations (no unexpected extras).
	if len(paths) != len(expectedPaths) {
		t.Errorf("PHIFieldPaths() has %d entries, expected %d", len(paths), len(expectedPaths))
	}
}

func TestDefaultPHIFields_AllHaveNonEmptyFields(t *testing.T) {
	for _, cfg := range DefaultPHIFields() {
		if cfg.ResourceType == "" {
			t.Error("found PHIFieldConfig with empty ResourceType")
		}
		if len(cfg.Fields) == 0 {
			t.Errorf("PHIFieldConfig for %q has no fields", cfg.ResourceType)
		}
	}
}
