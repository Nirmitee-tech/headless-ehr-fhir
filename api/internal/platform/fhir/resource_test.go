package fhir

import (
	"encoding/json"
	"testing"
	"time"
)

func TestResource_JSONSerialization(t *testing.T) {
	r := Resource{
		ResourceType: "Patient",
		ID:           "test-123",
		Meta: &Meta{
			VersionID: "1",
			Profile:   []string{"http://hl7.org/fhir/StructureDefinition/Patient"},
		},
	}

	data, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if parsed["resourceType"] != "Patient" {
		t.Errorf("expected Patient, got %v", parsed["resourceType"])
	}
	if parsed["id"] != "test-123" {
		t.Errorf("expected test-123, got %v", parsed["id"])
	}
}

func TestCoding_JSON(t *testing.T) {
	c := Coding{
		System:  "http://loinc.org",
		Code:    "8480-6",
		Display: "Systolic blood pressure",
	}

	data, err := json.Marshal(c)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var parsed Coding
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if parsed.System != c.System {
		t.Errorf("expected system %s, got %s", c.System, parsed.System)
	}
	if parsed.Code != c.Code {
		t.Errorf("expected code %s, got %s", c.Code, parsed.Code)
	}
}

func TestReference_JSON(t *testing.T) {
	ref := Reference{
		Reference: "Patient/123",
		Type:      "Patient",
		Display:   "John Smith",
	}

	data, err := json.Marshal(ref)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var parsed Reference
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if parsed.Reference != ref.Reference {
		t.Errorf("expected reference %s, got %s", ref.Reference, parsed.Reference)
	}
}

func TestPeriod_JSON(t *testing.T) {
	start := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	p := Period{
		Start: &start,
	}

	data, err := json.Marshal(p)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var parsed Period
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if parsed.Start == nil {
		t.Fatal("expected start to be set")
	}
	if !parsed.Start.Equal(start) {
		t.Errorf("expected start %v, got %v", start, *parsed.Start)
	}
	if parsed.End != nil {
		t.Error("expected end to be nil")
	}
}

func TestHumanName_JSON(t *testing.T) {
	name := HumanName{
		Use:    "official",
		Family: "Smith",
		Given:  []string{"John", "Michael"},
		Prefix: []string{"Dr"},
	}

	data, err := json.Marshal(name)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var parsed HumanName
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if parsed.Family != "Smith" {
		t.Errorf("expected family Smith, got %s", parsed.Family)
	}
	if len(parsed.Given) != 2 {
		t.Fatalf("expected 2 given names, got %d", len(parsed.Given))
	}
}

func TestIdentifier_JSON(t *testing.T) {
	ident := Identifier{
		Use:    "official",
		System: "http://hospital.org/mrn",
		Value:  "MRN-12345",
	}

	data, err := json.Marshal(ident)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var parsed Identifier
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if parsed.System != ident.System {
		t.Errorf("expected system %s, got %s", ident.System, parsed.System)
	}
	if parsed.Value != ident.Value {
		t.Errorf("expected value %s, got %s", ident.Value, parsed.Value)
	}
}
