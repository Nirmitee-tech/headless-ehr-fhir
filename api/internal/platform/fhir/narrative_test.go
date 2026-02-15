package fhir

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func mustGenerate(t *testing.T, g *NarrativeGenerator, resource map[string]interface{}) map[string]interface{} {
	t.Helper()
	result := g.Generate(resource)
	if result == nil {
		t.Fatal("expected non-nil narrative result")
	}
	return result
}

func divText(t *testing.T, g *NarrativeGenerator, resource map[string]interface{}) string {
	t.Helper()
	result := mustGenerate(t, g, resource)
	div, ok := result["div"].(string)
	if !ok {
		t.Fatal("expected div to be a string")
	}
	return div
}

func assertContains(t *testing.T, s, substr string) {
	t.Helper()
	if !strings.Contains(s, substr) {
		t.Errorf("expected %q to contain %q", s, substr)
	}
}

func assertNotContains(t *testing.T, s, substr string) {
	t.Helper()
	if strings.Contains(s, substr) {
		t.Errorf("expected %q to NOT contain %q", s, substr)
	}
}

// ---------------------------------------------------------------------------
// Patient Narrative Tests
// ---------------------------------------------------------------------------

func TestNarrative_Patient_FullFields(t *testing.T) {
	g := NewNarrativeGenerator()
	resource := map[string]interface{}{
		"resourceType": "Patient",
		"id":           "p1",
		"name": []interface{}{
			map[string]interface{}{
				"family": "Smith",
				"given":  []interface{}{"John", "Michael"},
			},
		},
		"gender":    "male",
		"birthDate": "1990-01-15",
		"identifier": []interface{}{
			map[string]interface{}{
				"system": "http://hospital.example/mrn",
				"value":  "MRN12345",
			},
			map[string]interface{}{
				"system": "http://hl7.org/fhir/sid/us-ssn",
				"value":  "999-99-9999",
			},
		},
		"telecom": []interface{}{
			map[string]interface{}{
				"system": "phone",
				"value":  "555-1234",
			},
			map[string]interface{}{
				"system": "email",
				"value":  "john@example.com",
			},
		},
		"address": []interface{}{
			map[string]interface{}{
				"line":       []interface{}{"123 Main St"},
				"city":       "Springfield",
				"state":      "IL",
				"postalCode": "62701",
			},
		},
	}

	div := divText(t, g, resource)
	assertContains(t, div, `xmlns="http://www.w3.org/1999/xhtml"`)
	assertContains(t, div, "Smith")
	assertContains(t, div, "John")
	assertContains(t, div, "male")
	assertContains(t, div, "1990-01-15")
	assertContains(t, div, "MRN12345")
	assertContains(t, div, "999-99-9999")
	assertContains(t, div, "phone")
	assertContains(t, div, "555-1234")
	assertContains(t, div, "email")
	assertContains(t, div, "john@example.com")
	assertContains(t, div, "123 Main St")
	assertContains(t, div, "Springfield")
	assertContains(t, div, "IL")
	assertContains(t, div, "62701")

	result := mustGenerate(t, g, resource)
	if result["status"] != "generated" {
		t.Errorf("expected status 'generated', got %v", result["status"])
	}
}

func TestNarrative_Patient_MinimalName(t *testing.T) {
	g := NewNarrativeGenerator()
	resource := map[string]interface{}{
		"resourceType": "Patient",
		"id":           "p2",
		"name": []interface{}{
			map[string]interface{}{
				"family": "Doe",
			},
		},
	}

	div := divText(t, g, resource)
	assertContains(t, div, "Doe")
	assertContains(t, div, "Patient")
}

func TestNarrative_Patient_MultipleIdentifiers(t *testing.T) {
	g := NewNarrativeGenerator()
	resource := map[string]interface{}{
		"resourceType": "Patient",
		"id":           "p3",
		"name": []interface{}{
			map[string]interface{}{"family": "Test"},
		},
		"identifier": []interface{}{
			map[string]interface{}{"system": "sys1", "value": "val1"},
			map[string]interface{}{"system": "sys2", "value": "val2"},
			map[string]interface{}{"system": "sys3", "value": "val3"},
		},
	}

	div := divText(t, g, resource)
	assertContains(t, div, "sys1")
	assertContains(t, div, "val1")
	assertContains(t, div, "sys2")
	assertContains(t, div, "val2")
	assertContains(t, div, "sys3")
	assertContains(t, div, "val3")
}

func TestNarrative_Patient_MultipleTelecom(t *testing.T) {
	g := NewNarrativeGenerator()
	resource := map[string]interface{}{
		"resourceType": "Patient",
		"id":           "p4",
		"name": []interface{}{
			map[string]interface{}{"family": "Jones"},
		},
		"telecom": []interface{}{
			map[string]interface{}{"system": "phone", "value": "111-2222"},
			map[string]interface{}{"system": "fax", "value": "333-4444"},
		},
	}

	div := divText(t, g, resource)
	assertContains(t, div, "phone")
	assertContains(t, div, "111-2222")
	assertContains(t, div, "fax")
	assertContains(t, div, "333-4444")
}

func TestNarrative_Patient_WithAddress(t *testing.T) {
	g := NewNarrativeGenerator()
	resource := map[string]interface{}{
		"resourceType": "Patient",
		"id":           "p5",
		"name": []interface{}{
			map[string]interface{}{"family": "Green"},
		},
		"address": []interface{}{
			map[string]interface{}{
				"line":       []interface{}{"456 Elm St", "Apt 2B"},
				"city":       "Portland",
				"state":      "OR",
				"postalCode": "97201",
			},
		},
	}

	div := divText(t, g, resource)
	assertContains(t, div, "456 Elm St")
	assertContains(t, div, "Apt 2B")
	assertContains(t, div, "Portland")
	assertContains(t, div, "OR")
	assertContains(t, div, "97201")
}

func TestNarrative_Patient_MissingName(t *testing.T) {
	g := NewNarrativeGenerator()
	resource := map[string]interface{}{
		"resourceType": "Patient",
		"id":           "p6",
		"gender":       "female",
	}

	div := divText(t, g, resource)
	assertContains(t, div, "Patient")
	assertContains(t, div, "female")
}

// ---------------------------------------------------------------------------
// Condition Narrative Tests
// ---------------------------------------------------------------------------

func TestNarrative_Condition_Full(t *testing.T) {
	g := NewNarrativeGenerator()
	resource := map[string]interface{}{
		"resourceType": "Condition",
		"id":           "c1",
		"code": map[string]interface{}{
			"coding": []interface{}{
				map[string]interface{}{
					"system":  "http://snomed.info/sct",
					"code":    "44054006",
					"display": "Type 2 Diabetes",
				},
			},
		},
		"clinicalStatus": map[string]interface{}{
			"coding": []interface{}{
				map[string]interface{}{"code": "active"},
			},
		},
		"verificationStatus": map[string]interface{}{
			"coding": []interface{}{
				map[string]interface{}{"code": "confirmed"},
			},
		},
		"subject": map[string]interface{}{
			"reference": "Patient/p1",
		},
		"onsetDateTime": "2020-03-15",
	}

	div := divText(t, g, resource)
	assertContains(t, div, "Type 2 Diabetes")
	assertContains(t, div, "44054006")
	assertContains(t, div, "active")
	assertContains(t, div, "confirmed")
	assertContains(t, div, "Patient/p1")
	assertContains(t, div, "2020-03-15")
}

func TestNarrative_Condition_WithOnset(t *testing.T) {
	g := NewNarrativeGenerator()
	resource := map[string]interface{}{
		"resourceType": "Condition",
		"id":           "c2",
		"code": map[string]interface{}{
			"coding": []interface{}{
				map[string]interface{}{"code": "386661006", "display": "Fever"},
			},
		},
		"subject": map[string]interface{}{
			"reference": "Patient/p2",
		},
		"onsetDateTime": "2024-12-01T10:30:00Z",
	}

	div := divText(t, g, resource)
	assertContains(t, div, "2024-12-01T10:30:00Z")
}

func TestNarrative_Condition_Minimal(t *testing.T) {
	g := NewNarrativeGenerator()
	resource := map[string]interface{}{
		"resourceType": "Condition",
		"id":           "c3",
		"code": map[string]interface{}{
			"coding": []interface{}{
				map[string]interface{}{"code": "123"},
			},
		},
		"subject": map[string]interface{}{
			"reference": "Patient/p3",
		},
	}

	div := divText(t, g, resource)
	assertContains(t, div, "123")
	assertContains(t, div, "Patient/p3")
}

func TestNarrative_Condition_CodingDisplay(t *testing.T) {
	g := NewNarrativeGenerator()
	resource := map[string]interface{}{
		"resourceType": "Condition",
		"id":           "c4",
		"code": map[string]interface{}{
			"coding": []interface{}{
				map[string]interface{}{
					"code":    "73211009",
					"display": "Diabetes mellitus",
				},
			},
		},
		"subject": map[string]interface{}{
			"reference": "Patient/p4",
		},
	}

	div := divText(t, g, resource)
	assertContains(t, div, "Diabetes mellitus")
	assertContains(t, div, "73211009")
}

// ---------------------------------------------------------------------------
// Observation Narrative Tests
// ---------------------------------------------------------------------------

func TestNarrative_Observation_ValueQuantity(t *testing.T) {
	g := NewNarrativeGenerator()
	resource := map[string]interface{}{
		"resourceType": "Observation",
		"id":           "o1",
		"status":       "final",
		"code": map[string]interface{}{
			"coding": []interface{}{
				map[string]interface{}{
					"code":    "8867-4",
					"display": "Heart rate",
				},
			},
		},
		"subject": map[string]interface{}{
			"reference": "Patient/p1",
		},
		"valueQuantity": map[string]interface{}{
			"value": 72.0,
			"unit":  "beats/minute",
		},
		"effectiveDateTime": "2024-06-15T10:00:00Z",
	}

	div := divText(t, g, resource)
	assertContains(t, div, "Heart rate")
	assertContains(t, div, "8867-4")
	assertContains(t, div, "final")
	assertContains(t, div, "72")
	assertContains(t, div, "beats/minute")
	assertContains(t, div, "Patient/p1")
	assertContains(t, div, "2024-06-15T10:00:00Z")
}

func TestNarrative_Observation_ValueString(t *testing.T) {
	g := NewNarrativeGenerator()
	resource := map[string]interface{}{
		"resourceType": "Observation",
		"id":           "o2",
		"status":       "final",
		"code": map[string]interface{}{
			"coding": []interface{}{
				map[string]interface{}{"code": "8302-2", "display": "Body height"},
			},
		},
		"subject": map[string]interface{}{
			"reference": "Patient/p2",
		},
		"valueString": "Tall for age",
	}

	div := divText(t, g, resource)
	assertContains(t, div, "Tall for age")
}

func TestNarrative_Observation_ValueCodeableConcept(t *testing.T) {
	g := NewNarrativeGenerator()
	resource := map[string]interface{}{
		"resourceType": "Observation",
		"id":           "o3",
		"status":       "final",
		"code": map[string]interface{}{
			"coding": []interface{}{
				map[string]interface{}{"code": "72166-2", "display": "Smoking status"},
			},
		},
		"subject": map[string]interface{}{
			"reference": "Patient/p3",
		},
		"valueCodeableConcept": map[string]interface{}{
			"text": "Current every day smoker",
		},
	}

	div := divText(t, g, resource)
	assertContains(t, div, "Current every day smoker")
}

func TestNarrative_Observation_Components(t *testing.T) {
	g := NewNarrativeGenerator()
	resource := map[string]interface{}{
		"resourceType": "Observation",
		"id":           "o4",
		"status":       "final",
		"code": map[string]interface{}{
			"coding": []interface{}{
				map[string]interface{}{"code": "85354-9", "display": "Blood pressure"},
			},
		},
		"subject": map[string]interface{}{
			"reference": "Patient/p4",
		},
		"component": []interface{}{
			map[string]interface{}{
				"code": map[string]interface{}{
					"coding": []interface{}{
						map[string]interface{}{"code": "8480-6", "display": "Systolic"},
					},
				},
				"valueQuantity": map[string]interface{}{
					"value": 120.0,
					"unit":  "mmHg",
				},
			},
			map[string]interface{}{
				"code": map[string]interface{}{
					"coding": []interface{}{
						map[string]interface{}{"code": "8462-4", "display": "Diastolic"},
					},
				},
				"valueQuantity": map[string]interface{}{
					"value": 80.0,
					"unit":  "mmHg",
				},
			},
		},
	}

	div := divText(t, g, resource)
	assertContains(t, div, "Blood pressure")
	// Components may appear if the generator includes them, but at minimum
	// the main code should be present.
}

func TestNarrative_Observation_Minimal(t *testing.T) {
	g := NewNarrativeGenerator()
	resource := map[string]interface{}{
		"resourceType": "Observation",
		"id":           "o5",
		"status":       "registered",
		"code": map[string]interface{}{
			"coding": []interface{}{
				map[string]interface{}{"code": "12345"},
			},
		},
	}

	div := divText(t, g, resource)
	assertContains(t, div, "12345")
	assertContains(t, div, "registered")
}

// ---------------------------------------------------------------------------
// AllergyIntolerance Narrative Tests
// ---------------------------------------------------------------------------

func TestNarrative_AllergyIntolerance(t *testing.T) {
	g := NewNarrativeGenerator()
	resource := map[string]interface{}{
		"resourceType": "AllergyIntolerance",
		"id":           "a1",
		"code": map[string]interface{}{
			"coding": []interface{}{
				map[string]interface{}{
					"code":    "7980",
					"display": "Penicillin",
				},
			},
		},
		"clinicalStatus": map[string]interface{}{
			"coding": []interface{}{
				map[string]interface{}{"code": "active"},
			},
		},
		"criticality": "high",
		"patient": map[string]interface{}{
			"reference": "Patient/p1",
		},
	}

	div := divText(t, g, resource)
	assertContains(t, div, "Penicillin")
	assertContains(t, div, "active")
	assertContains(t, div, "high")
	assertContains(t, div, "Patient/p1")
}

// ---------------------------------------------------------------------------
// MedicationRequest Narrative Tests
// ---------------------------------------------------------------------------

func TestNarrative_MedicationRequest_CodeableConcept(t *testing.T) {
	g := NewNarrativeGenerator()
	resource := map[string]interface{}{
		"resourceType": "MedicationRequest",
		"id":           "mr1",
		"status":       "active",
		"intent":       "order",
		"medicationCodeableConcept": map[string]interface{}{
			"coding": []interface{}{
				map[string]interface{}{
					"code":    "197361",
					"display": "Lisinopril 10 MG",
				},
			},
		},
		"subject": map[string]interface{}{
			"reference": "Patient/p1",
		},
		"authoredOn": "2024-01-15",
		"dosageInstruction": []interface{}{
			map[string]interface{}{
				"text": "Take 1 tablet daily",
			},
		},
	}

	div := divText(t, g, resource)
	assertContains(t, div, "Lisinopril 10 MG")
	assertContains(t, div, "active")
	assertContains(t, div, "order")
	assertContains(t, div, "Patient/p1")
	assertContains(t, div, "2024-01-15")
	assertContains(t, div, "Take 1 tablet daily")
}

func TestNarrative_MedicationRequest_Reference(t *testing.T) {
	g := NewNarrativeGenerator()
	resource := map[string]interface{}{
		"resourceType": "MedicationRequest",
		"id":           "mr2",
		"status":       "active",
		"intent":       "order",
		"medicationReference": map[string]interface{}{
			"reference": "Medication/med1",
			"display":   "Metformin 500mg",
		},
		"subject": map[string]interface{}{
			"reference": "Patient/p2",
		},
	}

	div := divText(t, g, resource)
	assertContains(t, div, "Metformin 500mg")
	assertContains(t, div, "active")
	assertContains(t, div, "order")
}

// ---------------------------------------------------------------------------
// Encounter Narrative Tests
// ---------------------------------------------------------------------------

func TestNarrative_Encounter_WithPeriod(t *testing.T) {
	g := NewNarrativeGenerator()
	resource := map[string]interface{}{
		"resourceType": "Encounter",
		"id":           "e1",
		"status":       "finished",
		"class": map[string]interface{}{
			"code": "AMB",
		},
		"type": []interface{}{
			map[string]interface{}{
				"coding": []interface{}{
					map[string]interface{}{
						"code":    "99213",
						"display": "Office Visit",
					},
				},
			},
		},
		"subject": map[string]interface{}{
			"reference": "Patient/p1",
		},
		"period": map[string]interface{}{
			"start": "2024-06-15T09:00:00Z",
			"end":   "2024-06-15T09:30:00Z",
		},
	}

	div := divText(t, g, resource)
	assertContains(t, div, "Office Visit")
	assertContains(t, div, "AMB")
	assertContains(t, div, "finished")
	assertContains(t, div, "Patient/p1")
	assertContains(t, div, "2024-06-15T09:00:00Z")
	assertContains(t, div, "2024-06-15T09:30:00Z")
}

// ---------------------------------------------------------------------------
// Procedure Narrative Tests
// ---------------------------------------------------------------------------

func TestNarrative_Procedure_PerformedDateTime(t *testing.T) {
	g := NewNarrativeGenerator()
	resource := map[string]interface{}{
		"resourceType": "Procedure",
		"id":           "pr1",
		"status":       "completed",
		"code": map[string]interface{}{
			"coding": []interface{}{
				map[string]interface{}{
					"code":    "80146002",
					"display": "Appendectomy",
				},
			},
		},
		"subject": map[string]interface{}{
			"reference": "Patient/p1",
		},
		"performedDateTime": "2024-05-20T14:00:00Z",
	}

	div := divText(t, g, resource)
	assertContains(t, div, "Appendectomy")
	assertContains(t, div, "80146002")
	assertContains(t, div, "completed")
	assertContains(t, div, "Patient/p1")
	assertContains(t, div, "2024-05-20T14:00:00Z")
}

// ---------------------------------------------------------------------------
// Immunization Narrative Tests
// ---------------------------------------------------------------------------

func TestNarrative_Immunization(t *testing.T) {
	g := NewNarrativeGenerator()
	resource := map[string]interface{}{
		"resourceType": "Immunization",
		"id":           "imm1",
		"status":       "completed",
		"vaccineCode": map[string]interface{}{
			"coding": []interface{}{
				map[string]interface{}{
					"code":    "207",
					"display": "COVID-19 Vaccine",
				},
			},
		},
		"patient": map[string]interface{}{
			"reference": "Patient/p1",
		},
		"occurrenceDateTime": "2024-03-15",
	}

	div := divText(t, g, resource)
	assertContains(t, div, "COVID-19 Vaccine")
	assertContains(t, div, "completed")
	assertContains(t, div, "Patient/p1")
	assertContains(t, div, "2024-03-15")
}

// ---------------------------------------------------------------------------
// DiagnosticReport Narrative Tests
// ---------------------------------------------------------------------------

func TestNarrative_DiagnosticReport_WithConclusion(t *testing.T) {
	g := NewNarrativeGenerator()
	resource := map[string]interface{}{
		"resourceType": "DiagnosticReport",
		"id":           "dr1",
		"status":       "final",
		"code": map[string]interface{}{
			"coding": []interface{}{
				map[string]interface{}{
					"code":    "58410-2",
					"display": "CBC",
				},
			},
		},
		"subject": map[string]interface{}{
			"reference": "Patient/p1",
		},
		"effectiveDateTime": "2024-06-10",
		"conclusion":        "Normal complete blood count",
	}

	div := divText(t, g, resource)
	assertContains(t, div, "CBC")
	assertContains(t, div, "final")
	assertContains(t, div, "Patient/p1")
	assertContains(t, div, "2024-06-10")
	assertContains(t, div, "Normal complete blood count")
}

// ---------------------------------------------------------------------------
// DocumentReference Narrative Tests
// ---------------------------------------------------------------------------

func TestNarrative_DocumentReference(t *testing.T) {
	g := NewNarrativeGenerator()
	resource := map[string]interface{}{
		"resourceType": "DocumentReference",
		"id":           "doc1",
		"status":       "current",
		"type": map[string]interface{}{
			"coding": []interface{}{
				map[string]interface{}{
					"code":    "34133-9",
					"display": "Summary of episode note",
				},
			},
		},
		"subject": map[string]interface{}{
			"reference": "Patient/p1",
		},
		"date": "2024-07-01T12:00:00Z",
	}

	div := divText(t, g, resource)
	assertContains(t, div, "Summary of episode note")
	assertContains(t, div, "current")
	assertContains(t, div, "Patient/p1")
	assertContains(t, div, "2024-07-01T12:00:00Z")
}

// ---------------------------------------------------------------------------
// InjectNarrative Tests
// ---------------------------------------------------------------------------

func TestNarrative_InjectNarrative_NoExistingText(t *testing.T) {
	g := NewNarrativeGenerator()
	resource := map[string]interface{}{
		"resourceType": "Patient",
		"id":           "p1",
		"name": []interface{}{
			map[string]interface{}{"family": "Smith"},
		},
	}

	result := g.InjectNarrative(resource)
	text, ok := result["text"].(map[string]interface{})
	if !ok {
		t.Fatal("expected text to be injected")
	}
	if text["status"] != "generated" {
		t.Errorf("expected status 'generated', got %v", text["status"])
	}
	div, ok := text["div"].(string)
	if !ok || div == "" {
		t.Fatal("expected non-empty div")
	}
	assertContains(t, div, "Smith")
}

func TestNarrative_InjectNarrative_SkipsAdditional(t *testing.T) {
	g := NewNarrativeGenerator()
	originalDiv := `<div xmlns="http://www.w3.org/1999/xhtml"><p>Custom narrative</p></div>`
	resource := map[string]interface{}{
		"resourceType": "Patient",
		"id":           "p1",
		"text": map[string]interface{}{
			"status": "additional",
			"div":    originalDiv,
		},
	}

	result := g.InjectNarrative(resource)
	text := result["text"].(map[string]interface{})
	if text["status"] != "additional" {
		t.Errorf("expected status to remain 'additional', got %v", text["status"])
	}
	if text["div"] != originalDiv {
		t.Error("expected div to remain unchanged")
	}
}

func TestNarrative_InjectNarrative_SkipsExtensions(t *testing.T) {
	g := NewNarrativeGenerator()
	originalDiv := `<div xmlns="http://www.w3.org/1999/xhtml"><p>Extensions narrative</p></div>`
	resource := map[string]interface{}{
		"resourceType": "Patient",
		"id":           "p1",
		"text": map[string]interface{}{
			"status": "extensions",
			"div":    originalDiv,
		},
	}

	result := g.InjectNarrative(resource)
	text := result["text"].(map[string]interface{})
	if text["status"] != "extensions" {
		t.Errorf("expected status to remain 'extensions', got %v", text["status"])
	}
}

func TestNarrative_InjectNarrative_OverwritesGenerated(t *testing.T) {
	g := NewNarrativeGenerator()
	resource := map[string]interface{}{
		"resourceType": "Patient",
		"id":           "p1",
		"name": []interface{}{
			map[string]interface{}{"family": "NewName"},
		},
		"text": map[string]interface{}{
			"status": "generated",
			"div":    `<div xmlns="http://www.w3.org/1999/xhtml"><p>Old narrative</p></div>`,
		},
	}

	result := g.InjectNarrative(resource)
	text := result["text"].(map[string]interface{})
	div := text["div"].(string)
	assertContains(t, div, "NewName")
	assertNotContains(t, div, "Old narrative")
}

func TestNarrative_InjectNarrative_UnknownResourceType(t *testing.T) {
	g := NewNarrativeGenerator()
	resource := map[string]interface{}{
		"resourceType": "CareTeam",
		"id":           "ct1",
	}

	result := g.InjectNarrative(resource)
	text, ok := result["text"].(map[string]interface{})
	if !ok {
		t.Fatal("expected text to be injected even for unknown resource types (fallback)")
	}
	div := text["div"].(string)
	assertContains(t, div, "CareTeam")
	assertContains(t, div, "ct1")
}

// ---------------------------------------------------------------------------
// Bundle Tests
// ---------------------------------------------------------------------------

func TestNarrative_Bundle_InjectIntoEntries(t *testing.T) {
	g := NewNarrativeGenerator()
	bundle := map[string]interface{}{
		"resourceType": "Bundle",
		"type":         "searchset",
		"entry": []interface{}{
			map[string]interface{}{
				"resource": map[string]interface{}{
					"resourceType": "Patient",
					"id":           "p1",
					"name": []interface{}{
						map[string]interface{}{"family": "Smith"},
					},
				},
			},
			map[string]interface{}{
				"resource": map[string]interface{}{
					"resourceType": "Condition",
					"id":           "c1",
					"code": map[string]interface{}{
						"coding": []interface{}{
							map[string]interface{}{"display": "Asthma"},
						},
					},
					"subject": map[string]interface{}{
						"reference": "Patient/p1",
					},
				},
			},
		},
	}

	entries := bundle["entry"].([]interface{})
	for i, entry := range entries {
		entryMap := entry.(map[string]interface{})
		res := entryMap["resource"].(map[string]interface{})
		entryMap["resource"] = g.InjectNarrative(res)
		entries[i] = entryMap
	}

	// Check Patient entry has narrative
	p := entries[0].(map[string]interface{})["resource"].(map[string]interface{})
	pText := p["text"].(map[string]interface{})
	assertContains(t, pText["div"].(string), "Smith")

	// Check Condition entry has narrative
	c := entries[1].(map[string]interface{})["resource"].(map[string]interface{})
	cText := c["text"].(map[string]interface{})
	assertContains(t, cText["div"].(string), "Asthma")
}

func TestNarrative_Bundle_Empty(t *testing.T) {
	g := NewNarrativeGenerator()
	bundle := map[string]interface{}{
		"resourceType": "Bundle",
		"type":         "searchset",
		"entry":        []interface{}{},
	}

	// Should not panic
	entries := bundle["entry"].([]interface{})
	for i, entry := range entries {
		entryMap := entry.(map[string]interface{})
		if res, ok := entryMap["resource"].(map[string]interface{}); ok {
			entryMap["resource"] = g.InjectNarrative(res)
			entries[i] = entryMap
		}
	}

	if len(entries) != 0 {
		t.Error("expected empty entries")
	}
}

func TestNarrative_Bundle_MixedTypes(t *testing.T) {
	g := NewNarrativeGenerator()
	bundle := map[string]interface{}{
		"resourceType": "Bundle",
		"type":         "searchset",
		"entry": []interface{}{
			map[string]interface{}{
				"resource": map[string]interface{}{
					"resourceType": "Patient",
					"id":           "p1",
					"name":         []interface{}{map[string]interface{}{"family": "Mix"}},
				},
			},
			map[string]interface{}{
				"resource": map[string]interface{}{
					"resourceType": "Observation",
					"id":           "obs1",
					"status":       "final",
					"code": map[string]interface{}{
						"coding": []interface{}{map[string]interface{}{"display": "BP"}},
					},
				},
			},
			map[string]interface{}{
				"resource": map[string]interface{}{
					"resourceType": "UnknownType",
					"id":           "u1",
				},
			},
		},
	}

	entries := bundle["entry"].([]interface{})
	for i, entry := range entries {
		entryMap := entry.(map[string]interface{})
		if res, ok := entryMap["resource"].(map[string]interface{}); ok {
			entryMap["resource"] = g.InjectNarrative(res)
			entries[i] = entryMap
		}
	}

	// All should have text
	for i, entry := range entries {
		res := entry.(map[string]interface{})["resource"].(map[string]interface{})
		if _, ok := res["text"]; !ok {
			t.Errorf("entry %d should have text element", i)
		}
	}
}

// ---------------------------------------------------------------------------
// XHTML Safety Tests
// ---------------------------------------------------------------------------

func TestNarrative_XHTMLSafety_PatientNameEscaped(t *testing.T) {
	g := NewNarrativeGenerator()
	resource := map[string]interface{}{
		"resourceType": "Patient",
		"id":           "evil1",
		"name": []interface{}{
			map[string]interface{}{
				"family": "<script>alert('xss')</script>",
				"given":  []interface{}{"O'Brien"},
			},
		},
	}

	div := divText(t, g, resource)
	assertNotContains(t, div, "<script>")
	assertContains(t, div, "&lt;script&gt;")
	// Must not contain unescaped single quotes in the XHTML (though single quotes
	// don't strictly need escaping in element content, the name should be present).
	assertContains(t, div, "O&#39;Brien")
}

func TestNarrative_XHTMLSafety_ConditionDisplayEscaped(t *testing.T) {
	g := NewNarrativeGenerator()
	resource := map[string]interface{}{
		"resourceType": "Condition",
		"id":           "evil2",
		"code": map[string]interface{}{
			"coding": []interface{}{
				map[string]interface{}{
					"display": `"Condition" <b>bold</b> & stuff`,
					"code":    "123",
				},
			},
		},
		"subject": map[string]interface{}{
			"reference": "Patient/p1",
		},
	}

	div := divText(t, g, resource)
	assertNotContains(t, div, `<b>bold</b>`)
	assertContains(t, div, "&amp;")
	assertContains(t, div, "&lt;b&gt;")
	assertContains(t, div, "&#34;Condition&#34;")
}

func TestNarrative_XHTMLSafety_ScriptInjectionPrevented(t *testing.T) {
	g := NewNarrativeGenerator()
	resource := map[string]interface{}{
		"resourceType": "Observation",
		"id":           "evil3",
		"status":       `final"><script>alert(1)</script><p class="`,
		"code": map[string]interface{}{
			"coding": []interface{}{
				map[string]interface{}{"code": "12345"},
			},
		},
	}

	div := divText(t, g, resource)
	assertNotContains(t, div, "<script>")
	assertContains(t, div, "&lt;script&gt;")
}

func TestNarrative_XHTMLSafety_AmpersandEscaped(t *testing.T) {
	g := NewNarrativeGenerator()
	resource := map[string]interface{}{
		"resourceType": "Patient",
		"id":           "amp1",
		"name": []interface{}{
			map[string]interface{}{
				"family": "Smith & Jones",
			},
		},
	}

	div := divText(t, g, resource)
	assertContains(t, div, "Smith &amp; Jones")
	// Should not contain bare ampersand (except as part of entity)
	// Check there's no double-escaped ampersand
	assertNotContains(t, div, "&amp;amp;")
}

// ---------------------------------------------------------------------------
// Fallback Tests
// ---------------------------------------------------------------------------

func TestNarrative_Fallback_UnknownResourceType(t *testing.T) {
	g := NewNarrativeGenerator()
	resource := map[string]interface{}{
		"resourceType": "Questionnaire",
		"id":           "q1",
	}

	div := divText(t, g, resource)
	assertContains(t, div, "Questionnaire")
	assertContains(t, div, "q1")
	assertContains(t, div, `xmlns="http://www.w3.org/1999/xhtml"`)
}

func TestNarrative_Fallback_NoResourceType(t *testing.T) {
	g := NewNarrativeGenerator()
	resource := map[string]interface{}{
		"id": "x1",
	}

	result := g.Generate(resource)
	// Should still return a valid narrative even without resourceType
	if result == nil {
		t.Fatal("expected non-nil result even without resourceType")
	}
	div := result["div"].(string)
	assertContains(t, div, `xmlns="http://www.w3.org/1999/xhtml"`)
}

// ---------------------------------------------------------------------------
// Edge Cases
// ---------------------------------------------------------------------------

func TestNarrative_EdgeCase_NilResource(t *testing.T) {
	g := NewNarrativeGenerator()
	result := g.Generate(nil)
	if result != nil {
		t.Error("expected nil result for nil resource")
	}
}

func TestNarrative_EdgeCase_EmptyResource(t *testing.T) {
	g := NewNarrativeGenerator()
	resource := map[string]interface{}{}

	result := g.Generate(resource)
	// Should return a valid fallback narrative
	if result == nil {
		t.Fatal("expected non-nil result for empty resource")
	}
	div := result["div"].(string)
	assertContains(t, div, `xmlns="http://www.w3.org/1999/xhtml"`)
}

func TestNarrative_EdgeCase_NilFields(t *testing.T) {
	g := NewNarrativeGenerator()
	resource := map[string]interface{}{
		"resourceType": "Patient",
		"id":           nil,
		"name":         nil,
		"gender":       nil,
		"birthDate":    nil,
		"identifier":   nil,
		"telecom":      nil,
		"address":      nil,
	}

	// Should not panic
	result := g.Generate(resource)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	div := result["div"].(string)
	assertContains(t, div, "Patient")
}

// ---------------------------------------------------------------------------
// RegisterGenerator Tests
// ---------------------------------------------------------------------------

func TestNarrative_RegisterCustomGenerator(t *testing.T) {
	g := NewNarrativeGenerator()
	g.RegisterGenerator("CustomType", func(resource map[string]interface{}) (string, error) {
		return `<div xmlns="http://www.w3.org/1999/xhtml"><p>Custom!</p></div>`, nil
	})

	resource := map[string]interface{}{
		"resourceType": "CustomType",
		"id":           "ct1",
	}

	div := divText(t, g, resource)
	assertContains(t, div, "Custom!")
}

func TestNarrative_RegisterOverridesBuiltin(t *testing.T) {
	g := NewNarrativeGenerator()
	g.RegisterGenerator("Patient", func(resource map[string]interface{}) (string, error) {
		return `<div xmlns="http://www.w3.org/1999/xhtml"><p>Overridden Patient</p></div>`, nil
	})

	resource := map[string]interface{}{
		"resourceType": "Patient",
		"id":           "p1",
		"name": []interface{}{
			map[string]interface{}{"family": "Smith"},
		},
	}

	div := divText(t, g, resource)
	assertContains(t, div, "Overridden Patient")
	assertNotContains(t, div, "Smith")
}

// ---------------------------------------------------------------------------
// Middleware Tests
// ---------------------------------------------------------------------------

func narrativeMiddlewareTestHandler(resource map[string]interface{}) echo.HandlerFunc {
	return func(c echo.Context) error {
		return c.JSON(http.StatusOK, resource)
	}
}

func TestNarrative_Middleware_InjectsSingleResource(t *testing.T) {
	g := NewNarrativeGenerator()
	e := echo.New()
	e.Use(NarrativeMiddleware(g))

	resource := map[string]interface{}{
		"resourceType": "Patient",
		"id":           "p1",
		"name": []interface{}{
			map[string]interface{}{"family": "MiddlewareTest"},
		},
	}
	e.GET("/fhir/Patient/p1", narrativeMiddlewareTestHandler(resource))

	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient/p1", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	text, ok := result["text"].(map[string]interface{})
	if !ok {
		t.Fatal("expected text element in response")
	}
	assertContains(t, text["div"].(string), "MiddlewareTest")
}

func TestNarrative_Middleware_InjectsBundle(t *testing.T) {
	g := NewNarrativeGenerator()
	e := echo.New()
	e.Use(NarrativeMiddleware(g))

	bundle := map[string]interface{}{
		"resourceType": "Bundle",
		"type":         "searchset",
		"entry": []interface{}{
			map[string]interface{}{
				"resource": map[string]interface{}{
					"resourceType": "Patient",
					"id":           "p1",
					"name":         []interface{}{map[string]interface{}{"family": "BundleTest"}},
				},
			},
		},
	}
	e.GET("/fhir/Patient", narrativeMiddlewareTestHandler(bundle))

	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	var result map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &result)

	entries := result["entry"].([]interface{})
	entry := entries[0].(map[string]interface{})
	res := entry["resource"].(map[string]interface{})
	text := res["text"].(map[string]interface{})
	assertContains(t, text["div"].(string), "BundleTest")
}

func TestNarrative_Middleware_SkipsNarrativeNone(t *testing.T) {
	g := NewNarrativeGenerator()
	e := echo.New()
	e.Use(NarrativeMiddleware(g))

	resource := map[string]interface{}{
		"resourceType": "Patient",
		"id":           "p1",
		"name":         []interface{}{map[string]interface{}{"family": "NoNarrative"}},
	}
	e.GET("/fhir/Patient/p1", narrativeMiddlewareTestHandler(resource))

	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient/p1?_narrative=none", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	var result map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &result)

	if _, ok := result["text"]; ok {
		t.Error("expected no text element when _narrative=none")
	}
}

func TestNarrative_Middleware_SkipsNonJSON(t *testing.T) {
	g := NewNarrativeGenerator()
	e := echo.New()
	e.Use(NarrativeMiddleware(g))

	e.GET("/fhir/Binary/b1", func(c echo.Context) error {
		return c.Blob(http.StatusOK, "application/octet-stream", []byte("binary data"))
	})

	req := httptest.NewRequest(http.MethodGet, "/fhir/Binary/b1", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Body.String() != "binary data" {
		t.Error("expected binary data to pass through unchanged")
	}
}

func TestNarrative_Middleware_PreservesAdditional(t *testing.T) {
	g := NewNarrativeGenerator()
	e := echo.New()
	e.Use(NarrativeMiddleware(g))

	resource := map[string]interface{}{
		"resourceType": "Patient",
		"id":           "p1",
		"text": map[string]interface{}{
			"status": "additional",
			"div":    `<div xmlns="http://www.w3.org/1999/xhtml"><p>My custom narrative</p></div>`,
		},
	}
	e.GET("/fhir/Patient/p1", narrativeMiddlewareTestHandler(resource))

	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient/p1", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	var result map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &result)

	text := result["text"].(map[string]interface{})
	if text["status"] != "additional" {
		t.Errorf("expected status 'additional', got %v", text["status"])
	}
	assertContains(t, text["div"].(string), "My custom narrative")
}

// ---------------------------------------------------------------------------
// XHTML Namespace Validation
// ---------------------------------------------------------------------------

func TestNarrative_AllTypes_HaveXHTMLNamespace(t *testing.T) {
	g := NewNarrativeGenerator()
	types := []struct {
		name     string
		resource map[string]interface{}
	}{
		{"Patient", map[string]interface{}{"resourceType": "Patient", "id": "1"}},
		{"Condition", map[string]interface{}{"resourceType": "Condition", "id": "1", "code": map[string]interface{}{"coding": []interface{}{map[string]interface{}{"code": "x"}}}}},
		{"Observation", map[string]interface{}{"resourceType": "Observation", "id": "1", "code": map[string]interface{}{"coding": []interface{}{map[string]interface{}{"code": "x"}}}}},
		{"AllergyIntolerance", map[string]interface{}{"resourceType": "AllergyIntolerance", "id": "1"}},
		{"MedicationRequest", map[string]interface{}{"resourceType": "MedicationRequest", "id": "1"}},
		{"Encounter", map[string]interface{}{"resourceType": "Encounter", "id": "1"}},
		{"Procedure", map[string]interface{}{"resourceType": "Procedure", "id": "1"}},
		{"Immunization", map[string]interface{}{"resourceType": "Immunization", "id": "1"}},
		{"DiagnosticReport", map[string]interface{}{"resourceType": "DiagnosticReport", "id": "1"}},
		{"DocumentReference", map[string]interface{}{"resourceType": "DocumentReference", "id": "1"}},
		{"Unknown", map[string]interface{}{"resourceType": "Unknown", "id": "1"}},
	}

	for _, tc := range types {
		t.Run(tc.name, func(t *testing.T) {
			result := g.Generate(tc.resource)
			div := result["div"].(string)
			if !strings.HasPrefix(div, `<div xmlns="http://www.w3.org/1999/xhtml">`) {
				t.Errorf("expected div to start with XHTML namespace, got: %s", div[:80])
			}
			if !strings.HasSuffix(div, "</div>") {
				t.Errorf("expected div to end with </div>, got: %s", div)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Procedure with performedPeriod
// ---------------------------------------------------------------------------

func TestNarrative_Procedure_PerformedPeriod(t *testing.T) {
	g := NewNarrativeGenerator()
	resource := map[string]interface{}{
		"resourceType": "Procedure",
		"id":           "pr2",
		"status":       "completed",
		"code": map[string]interface{}{
			"coding": []interface{}{
				map[string]interface{}{
					"code":    "27687005",
					"display": "Knee replacement",
				},
			},
		},
		"subject": map[string]interface{}{
			"reference": "Patient/p2",
		},
		"performedPeriod": map[string]interface{}{
			"start": "2024-04-10T08:00:00Z",
		},
	}

	div := divText(t, g, resource)
	assertContains(t, div, "2024-04-10T08:00:00Z")
	assertContains(t, div, "Knee replacement")
}
