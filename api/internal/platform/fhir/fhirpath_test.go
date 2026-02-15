package fhir

import (
	"math"
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

func newEngine() *FHIRPathEngine {
	return NewFHIRPathEngine()
}

func mustEval(t *testing.T, engine *FHIRPathEngine, resource map[string]interface{}, expr string) []interface{} {
	t.Helper()
	result, err := engine.Evaluate(resource, expr)
	if err != nil {
		t.Fatalf("Evaluate(%q) unexpected error: %v", expr, err)
	}
	return result
}

func mustEvalBool(t *testing.T, engine *FHIRPathEngine, resource map[string]interface{}, expr string) bool {
	t.Helper()
	result, err := engine.EvaluateBool(resource, expr)
	if err != nil {
		t.Fatalf("EvaluateBool(%q) unexpected error: %v", expr, err)
	}
	return result
}

func mustEvalString(t *testing.T, engine *FHIRPathEngine, resource map[string]interface{}, expr string) string {
	t.Helper()
	result, err := engine.EvaluateString(resource, expr)
	if err != nil {
		t.Fatalf("EvaluateString(%q) unexpected error: %v", expr, err)
	}
	return result
}

// ---------------------------------------------------------------------------
// Sample resources
// ---------------------------------------------------------------------------

func samplePatient() map[string]interface{} {
	return map[string]interface{}{
		"resourceType": "Patient",
		"id":           "pt-123",
		"active":       true,
		"birthDate":    "1990-03-15",
		"gender":       "male",
		"deceasedBoolean": false,
		"name": []interface{}{
			map[string]interface{}{
				"use":    "official",
				"family": "Smith",
				"given":  []interface{}{"John", "Michael"},
			},
			map[string]interface{}{
				"use":    "nickname",
				"family": "Smith",
				"given":  []interface{}{"Johnny"},
			},
		},
		"telecom": []interface{}{
			map[string]interface{}{
				"system": "phone",
				"value":  "555-0100",
				"use":    "home",
			},
			map[string]interface{}{
				"system": "email",
				"value":  "john@example.com",
				"use":    "work",
			},
			map[string]interface{}{
				"system": "phone",
				"value":  "555-0200",
				"use":    "work",
			},
		},
		"address": []interface{}{
			map[string]interface{}{
				"use":  "home",
				"city": "Springfield",
				"state": "IL",
				"line": []interface{}{"123 Main St"},
			},
		},
		"multipleBirthInteger": 2,
	}
}

func sampleObservation() map[string]interface{} {
	return map[string]interface{}{
		"resourceType": "Observation",
		"id":           "obs-bp-1",
		"status":       "final",
		"code": map[string]interface{}{
			"coding": []interface{}{
				map[string]interface{}{
					"system":  "http://loinc.org",
					"code":    "85354-9",
					"display": "Blood pressure panel",
				},
			},
		},
		"effectiveDateTime": "2024-06-15T10:30:00Z",
		"component": []interface{}{
			map[string]interface{}{
				"code": map[string]interface{}{
					"coding": []interface{}{
						map[string]interface{}{
							"system":  "http://loinc.org",
							"code":    "8480-6",
							"display": "Systolic blood pressure",
						},
					},
				},
				"valueQuantity": map[string]interface{}{
					"value":  float64(120),
					"unit":   "mmHg",
					"system": "http://unitsofmeasure.org",
					"code":   "mm[Hg]",
				},
			},
			map[string]interface{}{
				"code": map[string]interface{}{
					"coding": []interface{}{
						map[string]interface{}{
							"system":  "http://loinc.org",
							"code":    "8462-4",
							"display": "Diastolic blood pressure",
						},
					},
				},
				"valueQuantity": map[string]interface{}{
					"value":  float64(80),
					"unit":   "mmHg",
					"system": "http://unitsofmeasure.org",
					"code":   "mm[Hg]",
				},
			},
		},
	}
}

func sampleCondition() map[string]interface{} {
	return map[string]interface{}{
		"resourceType": "Condition",
		"id":           "cond-1",
		"clinicalStatus": map[string]interface{}{
			"coding": []interface{}{
				map[string]interface{}{
					"system": "http://terminology.hl7.org/CodeSystem/condition-clinical",
					"code":   "active",
				},
			},
		},
		"code": map[string]interface{}{
			"coding": []interface{}{
				map[string]interface{}{
					"system":  "http://snomed.info/sct",
					"code":    "73211009",
					"display": "Diabetes mellitus",
				},
				map[string]interface{}{
					"system":  "http://hl7.org/fhir/sid/icd-10-cm",
					"code":    "E11.9",
					"display": "Type 2 diabetes mellitus without complications",
				},
			},
		},
		"onsetDateTime": "2020-01-15",
	}
}

func sampleMedicationRequest() map[string]interface{} {
	return map[string]interface{}{
		"resourceType": "MedicationRequest",
		"id":           "medrx-1",
		"status":       "active",
		"intent":       "order",
		"dosageInstruction": []interface{}{
			map[string]interface{}{
				"text": "Take 1 tablet daily",
				"timing": map[string]interface{}{
					"repeat": map[string]interface{}{
						"frequency": float64(1),
						"period":    float64(1),
						"periodUnit": "d",
					},
				},
			},
			map[string]interface{}{
				"text": "Take 2 tablets twice daily",
				"timing": map[string]interface{}{
					"repeat": map[string]interface{}{
						"frequency": float64(2),
						"period":    float64(1),
						"periodUnit": "d",
					},
				},
			},
		},
	}
}

// ===========================================================================
// Path Navigation Tests
// ===========================================================================

func TestFHIRPath_Nav_SimpleField(t *testing.T) {
	e := newEngine()
	res := mustEval(t, e, samplePatient(), "Patient.id")
	if len(res) != 1 || res[0] != "pt-123" {
		t.Errorf("expected [pt-123], got %v", res)
	}
}

func TestFHIRPath_Nav_BooleanField(t *testing.T) {
	e := newEngine()
	res := mustEval(t, e, samplePatient(), "Patient.active")
	if len(res) != 1 || res[0] != true {
		t.Errorf("expected [true], got %v", res)
	}
}

func TestFHIRPath_Nav_NestedField(t *testing.T) {
	e := newEngine()
	res := mustEval(t, e, samplePatient(), "Patient.name.family")
	if len(res) != 2 {
		t.Fatalf("expected 2 family names, got %d: %v", len(res), res)
	}
	if res[0] != "Smith" || res[1] != "Smith" {
		t.Errorf("expected [Smith, Smith], got %v", res)
	}
}

func TestFHIRPath_Nav_DeepNested(t *testing.T) {
	e := newEngine()
	res := mustEval(t, e, sampleObservation(), "Observation.component.code.coding.code")
	if len(res) != 2 {
		t.Fatalf("expected 2 codes, got %d: %v", len(res), res)
	}
	if res[0] != "8480-6" || res[1] != "8462-4" {
		t.Errorf("expected [8480-6, 8462-4], got %v", res)
	}
}

func TestFHIRPath_Nav_ArrayTraversal(t *testing.T) {
	e := newEngine()
	res := mustEval(t, e, samplePatient(), "Patient.name")
	if len(res) != 2 {
		t.Errorf("expected 2 name entries, got %d", len(res))
	}
}

func TestFHIRPath_Nav_GivenNames(t *testing.T) {
	e := newEngine()
	res := mustEval(t, e, samplePatient(), "Patient.name.given")
	// Should flatten: John, Michael, Johnny
	if len(res) != 3 {
		t.Fatalf("expected 3 given names, got %d: %v", len(res), res)
	}
}

func TestFHIRPath_Nav_MissingField(t *testing.T) {
	e := newEngine()
	res := mustEval(t, e, samplePatient(), "Patient.maritalStatus")
	if len(res) != 0 {
		t.Errorf("expected empty result for missing field, got %v", res)
	}
}

func TestFHIRPath_Nav_ResourceTypeMismatch(t *testing.T) {
	e := newEngine()
	res := mustEval(t, e, samplePatient(), "Observation.code")
	if len(res) != 0 {
		t.Errorf("expected empty result for type mismatch, got %v", res)
	}
}

func TestFHIRPath_Nav_WithoutResourceType(t *testing.T) {
	e := newEngine()
	res := mustEval(t, e, samplePatient(), "name.given")
	if len(res) != 3 {
		t.Fatalf("expected 3 given names without resource prefix, got %d: %v", len(res), res)
	}
}

func TestFHIRPath_Nav_BirthDate(t *testing.T) {
	e := newEngine()
	res := mustEval(t, e, samplePatient(), "Patient.birthDate")
	if len(res) != 1 || res[0] != "1990-03-15" {
		t.Errorf("expected [1990-03-15], got %v", res)
	}
}

func TestFHIRPath_Nav_ObservationStatus(t *testing.T) {
	e := newEngine()
	s := mustEvalString(t, e, sampleObservation(), "Observation.status")
	if s != "final" {
		t.Errorf("expected 'final', got %q", s)
	}
}

func TestFHIRPath_Nav_AddressLine(t *testing.T) {
	e := newEngine()
	res := mustEval(t, e, samplePatient(), "Patient.address.line")
	if len(res) != 1 || res[0] != "123 Main St" {
		t.Errorf("expected [123 Main St], got %v", res)
	}
}

// ===========================================================================
// Literal Tests
// ===========================================================================

func TestFHIRPath_Literal_String(t *testing.T) {
	e := newEngine()
	res := mustEval(t, e, samplePatient(), "'hello world'")
	if len(res) != 1 || res[0] != "hello world" {
		t.Errorf("expected ['hello world'], got %v", res)
	}
}

func TestFHIRPath_Literal_Integer(t *testing.T) {
	e := newEngine()
	res := mustEval(t, e, samplePatient(), "42")
	if len(res) != 1 || res[0] != int64(42) {
		t.Errorf("expected [42], got %v", res)
	}
}

func TestFHIRPath_Literal_Decimal(t *testing.T) {
	e := newEngine()
	res := mustEval(t, e, samplePatient(), "3.14")
	if len(res) != 1 {
		t.Fatalf("expected 1 result, got %d", len(res))
	}
	f, ok := res[0].(float64)
	if !ok || math.Abs(f-3.14) > 0.001 {
		t.Errorf("expected [3.14], got %v", res)
	}
}

func TestFHIRPath_Literal_BoolTrue(t *testing.T) {
	e := newEngine()
	res := mustEval(t, e, samplePatient(), "true")
	if len(res) != 1 || res[0] != true {
		t.Errorf("expected [true], got %v", res)
	}
}

func TestFHIRPath_Literal_BoolFalse(t *testing.T) {
	e := newEngine()
	res := mustEval(t, e, samplePatient(), "false")
	if len(res) != 1 || res[0] != false {
		t.Errorf("expected [false], got %v", res)
	}
}

func TestFHIRPath_Literal_DateTime(t *testing.T) {
	e := newEngine()
	res := mustEval(t, e, samplePatient(), "@2024-01-01")
	if len(res) != 1 {
		t.Fatalf("expected 1 result, got %d", len(res))
	}
	dt, ok := res[0].(time.Time)
	if !ok {
		t.Fatalf("expected time.Time, got %T", res[0])
	}
	if dt.Year() != 2024 || dt.Month() != 1 || dt.Day() != 1 {
		t.Errorf("expected 2024-01-01, got %v", dt)
	}
}

// ===========================================================================
// Comparison Operator Tests
// ===========================================================================

func TestFHIRPath_Cmp_StringEqual(t *testing.T) {
	e := newEngine()
	b := mustEvalBool(t, e, samplePatient(), "Patient.gender = 'male'")
	if !b {
		t.Error("expected true for gender = 'male'")
	}
}

func TestFHIRPath_Cmp_StringNotEqual(t *testing.T) {
	e := newEngine()
	b := mustEvalBool(t, e, samplePatient(), "Patient.gender != 'female'")
	if !b {
		t.Error("expected true for gender != 'female'")
	}
}

func TestFHIRPath_Cmp_NumberEqual(t *testing.T) {
	e := newEngine()
	b := mustEvalBool(t, e, samplePatient(), "Patient.multipleBirthInteger = 2")
	if !b {
		t.Error("expected true for multipleBirthInteger = 2")
	}
}

func TestFHIRPath_Cmp_NumberLessThan(t *testing.T) {
	e := newEngine()
	b := mustEvalBool(t, e, samplePatient(), "Patient.multipleBirthInteger < 5")
	if !b {
		t.Error("expected true for multipleBirthInteger < 5")
	}
}

func TestFHIRPath_Cmp_NumberGreaterThan(t *testing.T) {
	e := newEngine()
	b := mustEvalBool(t, e, samplePatient(), "Patient.multipleBirthInteger > 1")
	if !b {
		t.Error("expected true for multipleBirthInteger > 1")
	}
}

func TestFHIRPath_Cmp_NumberLessEqual(t *testing.T) {
	e := newEngine()
	b := mustEvalBool(t, e, samplePatient(), "Patient.multipleBirthInteger <= 2")
	if !b {
		t.Error("expected true for multipleBirthInteger <= 2")
	}
}

func TestFHIRPath_Cmp_NumberGreaterEqual(t *testing.T) {
	e := newEngine()
	b := mustEvalBool(t, e, samplePatient(), "Patient.multipleBirthInteger >= 2")
	if !b {
		t.Error("expected true for multipleBirthInteger >= 2")
	}
}

func TestFHIRPath_Cmp_DateEqual(t *testing.T) {
	e := newEngine()
	b := mustEvalBool(t, e, samplePatient(), "Patient.birthDate = '1990-03-15'")
	if !b {
		t.Error("expected true for birthDate = '1990-03-15'")
	}
}

func TestFHIRPath_Cmp_BoolEqual(t *testing.T) {
	e := newEngine()
	b := mustEvalBool(t, e, samplePatient(), "Patient.active = true")
	if !b {
		t.Error("expected true for active = true")
	}
}

func TestFHIRPath_Cmp_FalseLiteral(t *testing.T) {
	e := newEngine()
	b := mustEvalBool(t, e, samplePatient(), "Patient.deceasedBoolean = false")
	if !b {
		t.Error("expected true for deceasedBoolean = false")
	}
}

// ===========================================================================
// Logical Operator Tests
// ===========================================================================

func TestFHIRPath_Logic_And(t *testing.T) {
	e := newEngine()
	b := mustEvalBool(t, e, samplePatient(), "Patient.active = true and Patient.gender = 'male'")
	if !b {
		t.Error("expected true for active and male")
	}
}

func TestFHIRPath_Logic_Or(t *testing.T) {
	e := newEngine()
	b := mustEvalBool(t, e, samplePatient(), "Patient.gender = 'female' or Patient.gender = 'male'")
	if !b {
		t.Error("expected true for female or male")
	}
}

func TestFHIRPath_Logic_Not(t *testing.T) {
	e := newEngine()
	b := mustEvalBool(t, e, samplePatient(), "Patient.active.not() = false")
	if !b {
		t.Error("expected true for active.not() = false")
	}
}

func TestFHIRPath_Logic_Implies_TrueTrue(t *testing.T) {
	e := newEngine()
	b := mustEvalBool(t, e, samplePatient(), "Patient.active implies Patient.gender = 'male'")
	if !b {
		t.Error("expected true for true implies true")
	}
}

func TestFHIRPath_Logic_Implies_FalseAnything(t *testing.T) {
	e := newEngine()
	b := mustEvalBool(t, e, samplePatient(), "Patient.deceasedBoolean implies Patient.gender = 'female'")
	if !b {
		t.Error("expected true for false implies anything")
	}
}

func TestFHIRPath_Logic_Combined(t *testing.T) {
	e := newEngine()
	b := mustEvalBool(t, e, samplePatient(), "(Patient.active = true and Patient.gender = 'male') or Patient.birthDate = '2000-01-01'")
	if !b {
		t.Error("expected true for combined logical")
	}
}

// ===========================================================================
// Collection Function Tests
// ===========================================================================

func TestFHIRPath_Fn_Where(t *testing.T) {
	e := newEngine()
	res := mustEval(t, e, samplePatient(), "Patient.name.where(use = 'official')")
	if len(res) != 1 {
		t.Fatalf("expected 1 official name, got %d: %v", len(res), res)
	}
}

func TestFHIRPath_Fn_WhereGiven(t *testing.T) {
	e := newEngine()
	res := mustEval(t, e, samplePatient(), "Patient.name.where(use = 'official').given")
	if len(res) != 2 {
		t.Fatalf("expected 2 given names of official, got %d: %v", len(res), res)
	}
	if res[0] != "John" || res[1] != "Michael" {
		t.Errorf("expected [John, Michael], got %v", res)
	}
}

func TestFHIRPath_Fn_Exists(t *testing.T) {
	e := newEngine()
	b := mustEvalBool(t, e, samplePatient(), "Patient.name.exists()")
	if !b {
		t.Error("expected true for name.exists()")
	}
}

func TestFHIRPath_Fn_ExistsWithExpr(t *testing.T) {
	e := newEngine()
	b := mustEvalBool(t, e, samplePatient(), "Patient.name.exists(use = 'official')")
	if !b {
		t.Error("expected true for name.exists(use = 'official')")
	}
}

func TestFHIRPath_Fn_ExistsFalse(t *testing.T) {
	e := newEngine()
	b := mustEvalBool(t, e, samplePatient(), "Patient.maritalStatus.exists()")
	if b {
		t.Error("expected false for maritalStatus.exists()")
	}
}

func TestFHIRPath_Fn_All(t *testing.T) {
	e := newEngine()
	b := mustEvalBool(t, e, samplePatient(), "Patient.name.all(family = 'Smith')")
	if !b {
		t.Error("expected true: all names have family Smith")
	}
}

func TestFHIRPath_Fn_AllFalse(t *testing.T) {
	e := newEngine()
	b := mustEvalBool(t, e, samplePatient(), "Patient.name.all(use = 'official')")
	if b {
		t.Error("expected false: not all names are official")
	}
}

func TestFHIRPath_Fn_Count(t *testing.T) {
	e := newEngine()
	res := mustEval(t, e, samplePatient(), "Patient.name.count()")
	if len(res) != 1 || res[0] != int64(2) {
		t.Errorf("expected [2], got %v", res)
	}
}

func TestFHIRPath_Fn_CountTelecom(t *testing.T) {
	e := newEngine()
	res := mustEval(t, e, samplePatient(), "Patient.telecom.count()")
	if len(res) != 1 || res[0] != int64(3) {
		t.Errorf("expected [3], got %v", res)
	}
}

func TestFHIRPath_Fn_First(t *testing.T) {
	e := newEngine()
	res := mustEval(t, e, samplePatient(), "Patient.name.first().family")
	if len(res) != 1 || res[0] != "Smith" {
		t.Errorf("expected [Smith], got %v", res)
	}
}

func TestFHIRPath_Fn_Last(t *testing.T) {
	e := newEngine()
	res := mustEval(t, e, samplePatient(), "Patient.name.last().use")
	if len(res) != 1 || res[0] != "nickname" {
		t.Errorf("expected [nickname], got %v", res)
	}
}

func TestFHIRPath_Fn_Tail(t *testing.T) {
	e := newEngine()
	res := mustEval(t, e, samplePatient(), "Patient.name.tail()")
	if len(res) != 1 {
		t.Fatalf("expected 1 element in tail, got %d", len(res))
	}
}

func TestFHIRPath_Fn_Empty(t *testing.T) {
	e := newEngine()
	b := mustEvalBool(t, e, samplePatient(), "Patient.maritalStatus.empty()")
	if !b {
		t.Error("expected true for maritalStatus.empty()")
	}
}

func TestFHIRPath_Fn_EmptyFalse(t *testing.T) {
	e := newEngine()
	b := mustEvalBool(t, e, samplePatient(), "Patient.name.empty()")
	if b {
		t.Error("expected false for name.empty()")
	}
}

func TestFHIRPath_Fn_Distinct(t *testing.T) {
	e := newEngine()
	res := mustEval(t, e, samplePatient(), "Patient.name.family.distinct()")
	if len(res) != 1 {
		t.Errorf("expected 1 distinct family name, got %d: %v", len(res), res)
	}
}

func TestFHIRPath_Fn_Select(t *testing.T) {
	e := newEngine()
	res := mustEval(t, e, samplePatient(), "Patient.name.select(family)")
	if len(res) != 2 {
		t.Fatalf("expected 2 families via select, got %d: %v", len(res), res)
	}
}

func TestFHIRPath_Fn_OfType(t *testing.T) {
	e := newEngine()
	res := mustEval(t, e, samplePatient(), "Patient.name.given.ofType(string)")
	// All given names are strings
	if len(res) != 3 {
		t.Errorf("expected 3 string given names, got %d: %v", len(res), res)
	}
}

func TestFHIRPath_Fn_WhereMultipleConditions(t *testing.T) {
	e := newEngine()
	res := mustEval(t, e, samplePatient(), "Patient.telecom.where(system = 'phone' and use = 'home').value")
	if len(res) != 1 || res[0] != "555-0100" {
		t.Errorf("expected [555-0100], got %v", res)
	}
}

// ===========================================================================
// String Function Tests
// ===========================================================================

func TestFHIRPath_Str_StartsWith(t *testing.T) {
	e := newEngine()
	b := mustEvalBool(t, e, samplePatient(), "Patient.id.startsWith('pt')")
	if !b {
		t.Error("expected true for id.startsWith('pt')")
	}
}

func TestFHIRPath_Str_EndsWith(t *testing.T) {
	e := newEngine()
	b := mustEvalBool(t, e, samplePatient(), "Patient.id.endsWith('123')")
	if !b {
		t.Error("expected true for id.endsWith('123')")
	}
}

func TestFHIRPath_Str_Contains(t *testing.T) {
	e := newEngine()
	b := mustEvalBool(t, e, samplePatient(), "Patient.id.contains('-')")
	if !b {
		t.Error("expected true for id.contains('-')")
	}
}

func TestFHIRPath_Str_Matches(t *testing.T) {
	e := newEngine()
	b := mustEvalBool(t, e, samplePatient(), "Patient.id.matches('^pt-[0-9]+$')")
	if !b {
		t.Error("expected true for id.matches regex")
	}
}

func TestFHIRPath_Str_Length(t *testing.T) {
	e := newEngine()
	res := mustEval(t, e, samplePatient(), "Patient.id.length()")
	if len(res) != 1 || res[0] != int64(6) {
		t.Errorf("expected [6], got %v", res)
	}
}

func TestFHIRPath_Str_Upper(t *testing.T) {
	e := newEngine()
	res := mustEval(t, e, samplePatient(), "Patient.gender.upper()")
	if len(res) != 1 || res[0] != "MALE" {
		t.Errorf("expected [MALE], got %v", res)
	}
}

func TestFHIRPath_Str_Lower(t *testing.T) {
	e := newEngine()
	res := mustEval(t, e, samplePatient(), "Patient.name.first().family.lower()")
	if len(res) != 1 || res[0] != "smith" {
		t.Errorf("expected [smith], got %v", res)
	}
}

func TestFHIRPath_Str_Replace(t *testing.T) {
	e := newEngine()
	res := mustEval(t, e, samplePatient(), "Patient.id.replace('pt', 'patient')")
	if len(res) != 1 || res[0] != "patient-123" {
		t.Errorf("expected [patient-123], got %v", res)
	}
}

func TestFHIRPath_Str_Substring(t *testing.T) {
	e := newEngine()
	res := mustEval(t, e, samplePatient(), "Patient.id.substring(3)")
	if len(res) != 1 || res[0] != "123" {
		t.Errorf("expected [123], got %v", res)
	}
}

func TestFHIRPath_Str_SubstringWithLength(t *testing.T) {
	e := newEngine()
	res := mustEval(t, e, samplePatient(), "Patient.id.substring(0, 2)")
	if len(res) != 1 || res[0] != "pt" {
		t.Errorf("expected [pt], got %v", res)
	}
}

// ===========================================================================
// Type Function Tests
// ===========================================================================

func TestFHIRPath_Type_Is(t *testing.T) {
	e := newEngine()
	b := mustEvalBool(t, e, samplePatient(), "Patient.id.is(string)")
	if !b {
		t.Error("expected true for id.is(string)")
	}
}

func TestFHIRPath_Type_As(t *testing.T) {
	e := newEngine()
	res := mustEval(t, e, samplePatient(), "Patient.id.as(string)")
	if len(res) != 1 || res[0] != "pt-123" {
		t.Errorf("expected [pt-123], got %v", res)
	}
}

func TestFHIRPath_Type_OfTypeString(t *testing.T) {
	e := newEngine()
	res := mustEval(t, e, samplePatient(), "Patient.name.given.ofType(string)")
	if len(res) != 3 {
		t.Errorf("expected 3 results, got %d", len(res))
	}
}

// ===========================================================================
// Math Function Tests
// ===========================================================================

func TestFHIRPath_Math_Abs(t *testing.T) {
	e := newEngine()
	res := mustEval(t, e, samplePatient(), "(-5).abs()")
	if len(res) != 1 {
		t.Fatalf("expected 1 result, got %d: %v", len(res), res)
	}
	v := toFloat(res[0])
	if v != 5.0 {
		t.Errorf("expected 5, got %v", res[0])
	}
}

func TestFHIRPath_Math_Ceiling(t *testing.T) {
	e := newEngine()
	res := mustEval(t, e, samplePatient(), "(3.2).ceiling()")
	if len(res) != 1 {
		t.Fatalf("expected 1 result, got %d", len(res))
	}
	v := toFloat(res[0])
	if v != 4.0 {
		t.Errorf("expected 4, got %v", res[0])
	}
}

func TestFHIRPath_Math_Floor(t *testing.T) {
	e := newEngine()
	res := mustEval(t, e, samplePatient(), "(3.8).floor()")
	if len(res) != 1 {
		t.Fatalf("expected 1 result, got %d", len(res))
	}
	v := toFloat(res[0])
	if v != 3.0 {
		t.Errorf("expected 3, got %v", res[0])
	}
}

func TestFHIRPath_Math_Round(t *testing.T) {
	e := newEngine()
	res := mustEval(t, e, samplePatient(), "(3.567).round()")
	if len(res) != 1 {
		t.Fatalf("expected 1 result, got %d", len(res))
	}
	v := toFloat(res[0])
	if v != 4.0 {
		t.Errorf("expected 4, got %v", res[0])
	}
}

// ===========================================================================
// Date/Time Function Tests
// ===========================================================================

func TestFHIRPath_DateTime_Now(t *testing.T) {
	e := newEngine()
	res := mustEval(t, e, samplePatient(), "now()")
	if len(res) != 1 {
		t.Fatalf("expected 1 result, got %d", len(res))
	}
	dt, ok := res[0].(time.Time)
	if !ok {
		t.Fatalf("expected time.Time, got %T", res[0])
	}
	if time.Since(dt) > 5*time.Second {
		t.Error("now() should return current time")
	}
}

func TestFHIRPath_DateTime_Today(t *testing.T) {
	e := newEngine()
	res := mustEval(t, e, samplePatient(), "today()")
	if len(res) != 1 {
		t.Fatalf("expected 1 result, got %d", len(res))
	}
	dt, ok := res[0].(time.Time)
	if !ok {
		t.Fatalf("expected time.Time, got %T", res[0])
	}
	now := time.Now()
	if dt.Year() != now.Year() || dt.Month() != now.Month() || dt.Day() != now.Day() {
		t.Errorf("today() should return today's date, got %v", dt)
	}
}

func TestFHIRPath_DateTime_ToDate(t *testing.T) {
	e := newEngine()
	res := mustEval(t, e, samplePatient(), "Patient.birthDate.toDate()")
	if len(res) != 1 {
		t.Fatalf("expected 1 result, got %d", len(res))
	}
	dt, ok := res[0].(time.Time)
	if !ok {
		t.Fatalf("expected time.Time, got %T", res[0])
	}
	if dt.Year() != 1990 || dt.Month() != 3 || dt.Day() != 15 {
		t.Errorf("expected 1990-03-15, got %v", dt)
	}
}

func TestFHIRPath_DateTime_ToDateTime(t *testing.T) {
	e := newEngine()
	res := mustEval(t, e, sampleObservation(), "Observation.effectiveDateTime.toDateTime()")
	if len(res) != 1 {
		t.Fatalf("expected 1 result, got %d", len(res))
	}
	dt, ok := res[0].(time.Time)
	if !ok {
		t.Fatalf("expected time.Time, got %T", res[0])
	}
	if dt.Year() != 2024 || dt.Month() != 6 || dt.Day() != 15 {
		t.Errorf("expected 2024-06-15, got %v", dt)
	}
}

// ===========================================================================
// Aggregate Function Tests
// ===========================================================================

func TestFHIRPath_Fn_Aggregate(t *testing.T) {
	e := newEngine()
	res := mustEval(t, e, sampleObservation(), "Observation.component.count()")
	if len(res) != 1 || res[0] != int64(2) {
		t.Errorf("expected [2], got %v", res)
	}
}

// ===========================================================================
// Boolean Checking Tests
// ===========================================================================

func TestFHIRPath_Fn_HasValue(t *testing.T) {
	e := newEngine()
	b := mustEvalBool(t, e, samplePatient(), "Patient.id.hasValue()")
	if !b {
		t.Error("expected true for id.hasValue()")
	}
}

func TestFHIRPath_Fn_HasValueFalse(t *testing.T) {
	e := newEngine()
	b := mustEvalBool(t, e, samplePatient(), "Patient.maritalStatus.hasValue()")
	if b {
		t.Error("expected false for maritalStatus.hasValue()")
	}
}

func TestFHIRPath_Fn_Iif(t *testing.T) {
	e := newEngine()
	res := mustEval(t, e, samplePatient(), "iif(Patient.active, 'yes', 'no')")
	if len(res) != 1 || res[0] != "yes" {
		t.Errorf("expected [yes], got %v", res)
	}
}

func TestFHIRPath_Fn_IifFalse(t *testing.T) {
	e := newEngine()
	res := mustEval(t, e, samplePatient(), "iif(Patient.deceasedBoolean, 'dead', 'alive')")
	if len(res) != 1 || res[0] != "alive" {
		t.Errorf("expected [alive], got %v", res)
	}
}

// ===========================================================================
// Indexing Tests
// ===========================================================================

func TestFHIRPath_Index_Zero(t *testing.T) {
	e := newEngine()
	res := mustEval(t, e, samplePatient(), "Patient.name[0].family")
	if len(res) != 1 || res[0] != "Smith" {
		t.Errorf("expected [Smith], got %v", res)
	}
}

func TestFHIRPath_Index_One(t *testing.T) {
	e := newEngine()
	res := mustEval(t, e, samplePatient(), "Patient.name[1].use")
	if len(res) != 1 || res[0] != "nickname" {
		t.Errorf("expected [nickname], got %v", res)
	}
}

func TestFHIRPath_Index_OutOfBounds(t *testing.T) {
	e := newEngine()
	res := mustEval(t, e, samplePatient(), "Patient.name[99].family")
	if len(res) != 0 {
		t.Errorf("expected empty for out of bounds, got %v", res)
	}
}

func TestFHIRPath_Index_GivenName(t *testing.T) {
	e := newEngine()
	res := mustEval(t, e, samplePatient(), "Patient.name[0].given[1]")
	if len(res) != 1 || res[0] != "Michael" {
		t.Errorf("expected [Michael], got %v", res)
	}
}

// ===========================================================================
// Union Tests
// ===========================================================================

func TestFHIRPath_Union_Basic(t *testing.T) {
	e := newEngine()
	res := mustEval(t, e, samplePatient(), "Patient.name[0].given | Patient.name[1].given")
	if len(res) != 3 {
		t.Fatalf("expected 3 names in union, got %d: %v", len(res), res)
	}
}

func TestFHIRPath_Union_DuplicatesRemoved(t *testing.T) {
	e := newEngine()
	res := mustEval(t, e, samplePatient(), "Patient.name[0].family | Patient.name[1].family")
	if len(res) != 1 {
		t.Errorf("expected 1 (deduplicated), got %d: %v", len(res), res)
	}
}

// ===========================================================================
// Complex Expression Tests
// ===========================================================================

func TestFHIRPath_Complex_OfficialGivenFirst(t *testing.T) {
	e := newEngine()
	res := mustEval(t, e, samplePatient(), "Patient.name.where(use = 'official').given.first()")
	if len(res) != 1 || res[0] != "John" {
		t.Errorf("expected [John], got %v", res)
	}
}

func TestFHIRPath_Complex_SystolicComponent(t *testing.T) {
	e := newEngine()
	res := mustEval(t, e, sampleObservation(), "Observation.component.where(code.coding.code = '8480-6').valueQuantity.value")
	if len(res) != 1 {
		t.Fatalf("expected 1 result, got %d: %v", len(res), res)
	}
	v := toFloat(res[0])
	if v != 120.0 {
		t.Errorf("expected 120, got %v", res[0])
	}
}

func TestFHIRPath_Complex_HomePhone(t *testing.T) {
	e := newEngine()
	res := mustEval(t, e, samplePatient(), "Patient.telecom.where(system = 'phone' and use = 'home').value")
	if len(res) != 1 || res[0] != "555-0100" {
		t.Errorf("expected [555-0100], got %v", res)
	}
}

func TestFHIRPath_Complex_SnomedCode(t *testing.T) {
	e := newEngine()
	res := mustEval(t, e, sampleCondition(), "Condition.code.coding.where(system = 'http://snomed.info/sct').code")
	if len(res) != 1 || res[0] != "73211009" {
		t.Errorf("expected [73211009], got %v", res)
	}
}

func TestFHIRPath_Complex_FreqGreaterThanOne(t *testing.T) {
	e := newEngine()
	b := mustEvalBool(t, e, sampleMedicationRequest(), "MedicationRequest.dosageInstruction.timing.repeat.where(frequency > 1).exists()")
	if !b {
		t.Error("expected true for dosage with frequency > 1")
	}
}

func TestFHIRPath_Complex_ChainedWhere(t *testing.T) {
	e := newEngine()
	res := mustEval(t, e, samplePatient(), "Patient.telecom.where(system = 'phone').where(use = 'work').value")
	if len(res) != 1 || res[0] != "555-0200" {
		t.Errorf("expected [555-0200], got %v", res)
	}
}

func TestFHIRPath_Complex_CountFiltered(t *testing.T) {
	e := newEngine()
	res := mustEval(t, e, samplePatient(), "Patient.telecom.where(system = 'phone').count()")
	if len(res) != 1 || res[0] != int64(2) {
		t.Errorf("expected [2], got %v", res)
	}
}

func TestFHIRPath_Complex_ExistsEmail(t *testing.T) {
	e := newEngine()
	b := mustEvalBool(t, e, samplePatient(), "Patient.telecom.where(system = 'email').exists()")
	if !b {
		t.Error("expected true for email exists")
	}
}

func TestFHIRPath_Complex_AllTelecomHaveValue(t *testing.T) {
	e := newEngine()
	b := mustEvalBool(t, e, samplePatient(), "Patient.telecom.all(value.exists())")
	if !b {
		t.Error("expected true: all telecom entries have a value")
	}
}

func TestFHIRPath_Complex_NestedSelectCount(t *testing.T) {
	e := newEngine()
	res := mustEval(t, e, sampleObservation(), "Observation.component.select(code.coding.code).count()")
	if len(res) != 1 || res[0] != int64(2) {
		t.Errorf("expected [2], got %v", res)
	}
}

// ===========================================================================
// Edge Case Tests
// ===========================================================================

func TestFHIRPath_Edge_EmptyExpression(t *testing.T) {
	e := newEngine()
	_, err := e.Evaluate(samplePatient(), "")
	if err == nil {
		t.Error("expected error for empty expression")
	}
}

func TestFHIRPath_Edge_InvalidExpression(t *testing.T) {
	e := newEngine()
	_, err := e.Evaluate(samplePatient(), "Patient.name.where(")
	if err == nil {
		t.Error("expected error for unclosed paren")
	}
}

func TestFHIRPath_Edge_NilResource(t *testing.T) {
	e := newEngine()
	res, err := e.Evaluate(nil, "Patient.id")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res) != 0 {
		t.Errorf("expected empty result for nil resource, got %v", res)
	}
}

func TestFHIRPath_Edge_NonExistentDeepPath(t *testing.T) {
	e := newEngine()
	res := mustEval(t, e, samplePatient(), "Patient.contact.name.family")
	if len(res) != 0 {
		t.Errorf("expected empty for non-existent deep path, got %v", res)
	}
}

func TestFHIRPath_Edge_DeeplyNestedNulls(t *testing.T) {
	e := newEngine()
	resource := map[string]interface{}{
		"resourceType": "Patient",
		"id":           "p1",
		"name":         []interface{}{},
	}
	res := mustEval(t, e, resource, "Patient.name.given.first()")
	if len(res) != 0 {
		t.Errorf("expected empty for empty array traversal, got %v", res)
	}
}

func TestFHIRPath_Edge_EmptyResource(t *testing.T) {
	e := newEngine()
	resource := map[string]interface{}{}
	res := mustEval(t, e, resource, "name")
	if len(res) != 0 {
		t.Errorf("expected empty for empty resource, got %v", res)
	}
}

func TestFHIRPath_Edge_OnlyResourceType(t *testing.T) {
	e := newEngine()
	res := mustEval(t, e, samplePatient(), "Patient")
	if len(res) != 1 {
		t.Errorf("expected 1 result for bare resource type, got %d", len(res))
	}
}

// ===========================================================================
// EvaluateBool / EvaluateString convenience tests
// ===========================================================================

func TestFHIRPath_EvaluateBool_True(t *testing.T) {
	e := newEngine()
	b := mustEvalBool(t, e, samplePatient(), "Patient.active")
	if !b {
		t.Error("expected true")
	}
}

func TestFHIRPath_EvaluateBool_EmptyIsFalse(t *testing.T) {
	e := newEngine()
	b := mustEvalBool(t, e, samplePatient(), "Patient.maritalStatus")
	if b {
		t.Error("expected false for empty collection")
	}
}

func TestFHIRPath_EvaluateString_Simple(t *testing.T) {
	e := newEngine()
	s := mustEvalString(t, e, samplePatient(), "Patient.gender")
	if s != "male" {
		t.Errorf("expected 'male', got %q", s)
	}
}

func TestFHIRPath_EvaluateString_Empty(t *testing.T) {
	e := newEngine()
	s := mustEvalString(t, e, samplePatient(), "Patient.maritalStatus")
	if s != "" {
		t.Errorf("expected empty string, got %q", s)
	}
}

// ===========================================================================
// Parenthesized Expression Tests
// ===========================================================================

func TestFHIRPath_Paren_GroupedComparison(t *testing.T) {
	e := newEngine()
	b := mustEvalBool(t, e, samplePatient(), "(Patient.multipleBirthInteger > 0) and (Patient.active = true)")
	if !b {
		t.Error("expected true")
	}
}

func TestFHIRPath_Paren_Nested(t *testing.T) {
	e := newEngine()
	b := mustEvalBool(t, e, samplePatient(), "((Patient.gender = 'male'))")
	if !b {
		t.Error("expected true for nested parens")
	}
}

// ===========================================================================
// Helper
// ===========================================================================

func toFloat(v interface{}) float64 {
	switch n := v.(type) {
	case float64:
		return n
	case int64:
		return float64(n)
	case int:
		return float64(n)
	default:
		return math.NaN()
	}
}
