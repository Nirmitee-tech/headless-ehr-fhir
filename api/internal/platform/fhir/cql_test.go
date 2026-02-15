package fhir

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
)

// ===========================================================================
// Test helpers
// ===========================================================================

func newCQLEngine() *CQLEngine {
	return NewCQLEngine()
}

func newMeasureEvaluator() *MeasureEvaluator {
	return NewMeasureEvaluator()
}

// diabeticPatient builds a patient with diabetes, age ~55.
func diabeticPatient() map[string]interface{} {
	return map[string]interface{}{
		"resourceType": "Patient",
		"id":           "pt-diab-1",
		"gender":       "male",
		"birthDate":    "1970-06-15",
		"active":       true,
	}
}

// diabetesCondition returns a Condition with diabetes ICD-10 code.
func diabetesCondition() map[string]interface{} {
	return map[string]interface{}{
		"resourceType": "Condition",
		"id":           "cond-diab-1",
		"subject":      map[string]interface{}{"reference": "Patient/pt-diab-1"},
		"code": map[string]interface{}{
			"coding": []interface{}{
				map[string]interface{}{
					"system": "http://hl7.org/fhir/sid/icd-10-cm",
					"code":   "E11.9",
				},
			},
		},
		"clinicalStatus": map[string]interface{}{
			"coding": []interface{}{
				map[string]interface{}{"code": "active"},
			},
		},
	}
}

// hba1cObservation returns an HbA1c Observation with the given value.
func hba1cObservation(value float64, dateStr string) map[string]interface{} {
	return map[string]interface{}{
		"resourceType": "Observation",
		"id":           "obs-hba1c-1",
		"status":       "final",
		"code": map[string]interface{}{
			"coding": []interface{}{
				map[string]interface{}{
					"system": "http://loinc.org",
					"code":   "4548-4",
				},
			},
		},
		"effectiveDateTime": dateStr,
		"valueQuantity": map[string]interface{}{
			"value": value,
			"unit":  "%",
		},
	}
}

// hypertensionPatient builds a patient with hypertension, age ~50.
func hypertensionPatient() map[string]interface{} {
	return map[string]interface{}{
		"resourceType": "Patient",
		"id":           "pt-htn-1",
		"gender":       "male",
		"birthDate":    "1975-03-20",
		"active":       true,
	}
}

// hypertensionCondition returns a Condition for hypertension (I10).
func hypertensionCondition() map[string]interface{} {
	return map[string]interface{}{
		"resourceType": "Condition",
		"id":           "cond-htn-1",
		"subject":      map[string]interface{}{"reference": "Patient/pt-htn-1"},
		"code": map[string]interface{}{
			"coding": []interface{}{
				map[string]interface{}{
					"system": "http://hl7.org/fhir/sid/icd-10-cm",
					"code":   "I10",
				},
			},
		},
		"clinicalStatus": map[string]interface{}{
			"coding": []interface{}{
				map[string]interface{}{"code": "active"},
			},
		},
	}
}

// bpObservation returns a BP observation with given sys/dia values.
func bpObservation(systolic, diastolic float64, dateStr string) map[string]interface{} {
	return map[string]interface{}{
		"resourceType": "Observation",
		"id":           "obs-bp-1",
		"status":       "final",
		"code": map[string]interface{}{
			"coding": []interface{}{
				map[string]interface{}{
					"system": "http://loinc.org",
					"code":   "85354-9",
				},
			},
		},
		"effectiveDateTime": dateStr,
		"component": []interface{}{
			map[string]interface{}{
				"code": map[string]interface{}{
					"coding": []interface{}{
						map[string]interface{}{
							"system": "http://loinc.org",
							"code":   "8480-6",
						},
					},
				},
				"valueQuantity": map[string]interface{}{
					"value": systolic,
					"unit":  "mmHg",
				},
			},
			map[string]interface{}{
				"code": map[string]interface{}{
					"coding": []interface{}{
						map[string]interface{}{
							"system": "http://loinc.org",
							"code":   "8462-4",
						},
					},
				},
				"valueQuantity": map[string]interface{}{
					"value": diastolic,
					"unit":  "mmHg",
				},
			},
		},
	}
}

// femalePatient builds a female patient for breast cancer screening tests.
func femalePatient(birthDate string) map[string]interface{} {
	return map[string]interface{}{
		"resourceType": "Patient",
		"id":           "pt-female-1",
		"gender":       "female",
		"birthDate":    birthDate,
		"active":       true,
	}
}

// mammogramReport returns a DiagnosticReport for mammography.
func mammogramReport(dateStr string) map[string]interface{} {
	return map[string]interface{}{
		"resourceType": "DiagnosticReport",
		"id":           "dr-mammo-1",
		"status":       "final",
		"code": map[string]interface{}{
			"coding": []interface{}{
				map[string]interface{}{
					"system": "http://loinc.org",
					"code":   "24606-6",
				},
			},
		},
		"effectiveDateTime": dateStr,
	}
}

// hospiceEncounter returns an Encounter with hospice type.
func hospiceEncounter() map[string]interface{} {
	return map[string]interface{}{
		"resourceType": "Encounter",
		"id":           "enc-hospice-1",
		"status":       "finished",
		"class": map[string]interface{}{
			"code": "hospice",
		},
		"type": []interface{}{
			map[string]interface{}{
				"coding": []interface{}{
					map[string]interface{}{
						"system": "http://snomed.info/sct",
						"code":   "385765002",
					},
				},
			},
		},
	}
}

// ===========================================================================
// CQL Engine Tests
// ===========================================================================

func TestCQLEngine_NewEngine(t *testing.T) {
	engine := newCQLEngine()
	if engine == nil {
		t.Fatal("NewCQLEngine returned nil")
	}
	if engine.fhirpath == nil {
		t.Fatal("CQLEngine.fhirpath is nil")
	}
}

func TestCQLEngine_EvaluateSimpleFHIRPath(t *testing.T) {
	engine := newCQLEngine()
	patient := diabeticPatient()
	resources := map[string][]map[string]interface{}{}

	result, err := engine.EvaluateExpression(context.Background(), "Patient.gender", patient, resources)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "male" {
		t.Fatalf("expected 'male', got %v", result)
	}
}

func TestCQLEngine_EvaluateWithPatientContext(t *testing.T) {
	engine := newCQLEngine()
	patient := diabeticPatient()
	resources := map[string][]map[string]interface{}{}

	result, err := engine.EvaluateExpression(context.Background(), "Patient.id", patient, resources)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "pt-diab-1" {
		t.Fatalf("expected 'pt-diab-1', got %v", result)
	}
}

func TestCQLEngine_EvaluateWithResourcesConditions(t *testing.T) {
	engine := newCQLEngine()
	patient := diabeticPatient()
	resources := map[string][]map[string]interface{}{
		"Condition": {diabetesCondition()},
	}

	result, err := engine.EvaluateExpression(context.Background(), "Condition.exists()", patient, resources)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != true {
		t.Fatalf("expected true, got %v", result)
	}
}

func TestCQLEngine_EvaluateWithResourcesObservations(t *testing.T) {
	engine := newCQLEngine()
	patient := diabeticPatient()
	resources := map[string][]map[string]interface{}{
		"Observation": {hba1cObservation(8.5, "2025-06-01")},
	}

	result, err := engine.EvaluateExpression(context.Background(), "Observation.count()", patient, resources)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	count, ok := result.(int)
	if !ok {
		t.Fatalf("expected int, got %T: %v", result, result)
	}
	if count != 1 {
		t.Fatalf("expected 1, got %d", count)
	}
}

func TestCQLEngine_EvaluateAgeCalculation(t *testing.T) {
	engine := newCQLEngine()
	patient := map[string]interface{}{
		"resourceType": "Patient",
		"id":           "pt-age",
		"birthDate":    "2000-01-01",
		"gender":       "male",
	}
	resources := map[string][]map[string]interface{}{}

	result, err := engine.EvaluateExpression(context.Background(), "AgeInYears()", patient, resources)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	age, ok := result.(int)
	if !ok {
		t.Fatalf("expected int, got %T: %v", result, result)
	}
	if age < 25 || age > 27 {
		t.Fatalf("expected age around 26, got %d", age)
	}
}

func TestCQLEngine_EvaluateBooleanExpression(t *testing.T) {
	engine := newCQLEngine()
	patient := diabeticPatient()
	resources := map[string][]map[string]interface{}{}

	result, err := engine.EvaluateExpression(context.Background(), "Patient.gender = 'male'", patient, resources)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != true {
		t.Fatalf("expected true, got %v", result)
	}
}

func TestCQLEngine_EvaluateFilterByCode(t *testing.T) {
	engine := newCQLEngine()
	patient := diabeticPatient()
	resources := map[string][]map[string]interface{}{
		"Condition": {
			diabetesCondition(),
			{
				"resourceType": "Condition",
				"id":           "cond-other",
				"code": map[string]interface{}{
					"coding": []interface{}{
						map[string]interface{}{
							"system": "http://hl7.org/fhir/sid/icd-10-cm",
							"code":   "J06.9",
						},
					},
				},
			},
		},
	}

	result, err := engine.EvaluateExpression(context.Background(), "HasConditionCode('E11')", patient, resources)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != true {
		t.Fatalf("expected true for E11 prefix match, got %v", result)
	}
}

func TestCQLEngine_EvaluateLibrary(t *testing.T) {
	engine := newCQLEngine()
	patient := diabeticPatient()
	resources := map[string][]map[string]interface{}{
		"Condition":   {diabetesCondition()},
		"Observation": {hba1cObservation(7.5, "2025-06-01")},
	}

	library := &CQLLibrary{
		Name:    "TestLibrary",
		Version: "1.0.0",
		Status:  "active",
		Definitions: map[string]CQLDefinition{
			"IsMale": {
				Name:       "IsMale",
				Expression: "Patient.gender = 'male'",
				Context:    "Patient",
				Type:       "Boolean",
			},
			"HasDiabetes": {
				Name:       "HasDiabetes",
				Expression: "HasConditionCode('E11')",
				Context:    "Patient",
				Type:       "Boolean",
			},
		},
	}

	results, err := engine.EvaluateLibrary(context.Background(), library, patient, resources)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if results["IsMale"] != true {
		t.Fatalf("expected IsMale=true, got %v", results["IsMale"])
	}
	if results["HasDiabetes"] != true {
		t.Fatalf("expected HasDiabetes=true, got %v", results["HasDiabetes"])
	}
}

func TestCQLEngine_ParameterHandling(t *testing.T) {
	engine := newCQLEngine()
	library := &CQLLibrary{
		Name:    "ParamLib",
		Version: "1.0.0",
		Status:  "active",
		Parameters: []CQLParameter{
			{Name: "MeasurePeriod", Type: "Period"},
		},
		Definitions: map[string]CQLDefinition{
			"PatientGender": {
				Name:       "PatientGender",
				Expression: "Patient.gender",
				Context:    "Patient",
				Type:       "String",
			},
		},
	}

	patient := diabeticPatient()
	resources := map[string][]map[string]interface{}{}

	results, err := engine.EvaluateLibrary(context.Background(), library, patient, resources)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if results["PatientGender"] != "male" {
		t.Fatalf("expected 'male', got %v", results["PatientGender"])
	}
}

func TestCQLEngine_ErrorInvalidExpression(t *testing.T) {
	engine := newCQLEngine()
	patient := diabeticPatient()
	resources := map[string][]map[string]interface{}{}

	_, err := engine.EvaluateExpression(context.Background(), "!!!invalid!!!", patient, resources)
	if err == nil {
		t.Fatal("expected error for invalid expression, got nil")
	}
}

func TestCQLEngine_ErrorNilPatient(t *testing.T) {
	engine := newCQLEngine()
	resources := map[string][]map[string]interface{}{}

	_, err := engine.EvaluateExpression(context.Background(), "Patient.gender", nil, resources)
	if err == nil {
		t.Fatal("expected error for nil patient, got nil")
	}
}

func TestCQLEngine_EmptyResources(t *testing.T) {
	engine := newCQLEngine()
	patient := diabeticPatient()
	resources := map[string][]map[string]interface{}{}

	result, err := engine.EvaluateExpression(context.Background(), "Condition.exists()", patient, resources)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != false {
		t.Fatalf("expected false for no conditions, got %v", result)
	}
}

func TestCQLEngine_AgeInRangeCheck(t *testing.T) {
	engine := newCQLEngine()
	patient := diabeticPatient()
	resources := map[string][]map[string]interface{}{}

	result, err := engine.EvaluateExpression(context.Background(), "AgeInYearsAt('2025-06-15')", patient, resources)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	age, ok := result.(int)
	if !ok {
		t.Fatalf("expected int, got %T: %v", result, result)
	}
	if age != 55 {
		t.Fatalf("expected age 55, got %d", age)
	}
}

// ===========================================================================
// Measure Model Tests
// ===========================================================================

func TestMeasure_Create(t *testing.T) {
	m := &Measure{
		ID:      "cms122",
		URL:     "http://example.org/fhir/Measure/cms122",
		Name:    "DiabetesHbA1cControl",
		Title:   "Diabetes: Hemoglobin A1c (HbA1c) Poor Control (> 9%)",
		Status:  "active",
		Scoring: "proportion",
		Type:    []string{"process"},
		Group: []MeasureGroup{
			{
				ID: "group-1",
				Population: []MeasurePopulation{
					{Code: "initial-population", Expression: "InitialPopulation"},
					{Code: "denominator", Expression: "Denominator"},
					{Code: "numerator", Expression: "Numerator"},
				},
			},
		},
	}
	if m.ID != "cms122" {
		t.Fatalf("expected ID 'cms122', got '%s'", m.ID)
	}
	if len(m.Group) != 1 {
		t.Fatalf("expected 1 group, got %d", len(m.Group))
	}
	if len(m.Group[0].Population) != 3 {
		t.Fatalf("expected 3 populations, got %d", len(m.Group[0].Population))
	}
}

func TestMeasure_FHIRJSONSerialization(t *testing.T) {
	m := &Measure{
		ID:      "cms122",
		URL:     "http://example.org/fhir/Measure/cms122",
		Name:    "DiabetesHbA1cControl",
		Title:   "Diabetes: HbA1c Poor Control",
		Status:  "active",
		Scoring: "proportion",
		Type:    []string{"process"},
		Library: []string{"http://example.org/fhir/Library/diabetes-hba1c"},
	}

	fhirJSON := m.ToFHIR()
	data, err := json.Marshal(fhirJSON)
	if err != nil {
		t.Fatalf("json marshal error: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("json unmarshal error: %v", err)
	}
	if parsed["resourceType"] != "Measure" {
		t.Fatalf("expected resourceType 'Measure', got %v", parsed["resourceType"])
	}
	if parsed["id"] != "cms122" {
		t.Fatalf("expected id 'cms122', got %v", parsed["id"])
	}
	if parsed["status"] != "active" {
		t.Fatalf("expected status 'active', got %v", parsed["status"])
	}
}

func TestMeasure_ScoringTypeValidation(t *testing.T) {
	validTypes := []string{"proportion", "ratio", "continuous-variable", "cohort"}
	for _, st := range validTypes {
		if !isValidScoringType(st) {
			t.Fatalf("expected %q to be valid scoring type", st)
		}
	}
	if isValidScoringType("invalid") {
		t.Fatal("expected 'invalid' to not be a valid scoring type")
	}
}

func TestMeasure_PopulationCodeValidation(t *testing.T) {
	validCodes := []string{
		"initial-population", "numerator", "denominator",
		"denominator-exclusion", "denominator-exception",
		"numerator-exclusion", "measure-population",
		"measure-observation",
	}
	for _, code := range validCodes {
		if !isValidPopulationCode(code) {
			t.Fatalf("expected %q to be valid population code", code)
		}
	}
	if isValidPopulationCode("invalid") {
		t.Fatal("expected 'invalid' to not be a valid population code")
	}
}

func TestMeasure_WithStratifier(t *testing.T) {
	m := &Measure{
		ID:     "test-strat",
		Status: "active",
		Group: []MeasureGroup{
			{
				ID: "group-1",
				Population: []MeasurePopulation{
					{Code: "initial-population", Expression: "IP"},
				},
				Stratifier: []MeasureStratifier{
					{Code: "gender", Expression: "Patient.gender"},
				},
			},
		},
	}
	if len(m.Group[0].Stratifier) != 1 {
		t.Fatalf("expected 1 stratifier, got %d", len(m.Group[0].Stratifier))
	}
	if m.Group[0].Stratifier[0].Code != "gender" {
		t.Fatalf("expected stratifier code 'gender', got %s", m.Group[0].Stratifier[0].Code)
	}
}

// ===========================================================================
// Individual Evaluation Tests — Diabetes HbA1c (CMS122)
// ===========================================================================

func TestMeasureEvaluator_DiabetesInNumerator(t *testing.T) {
	eval := newMeasureEvaluator()
	patient := diabeticPatient()
	resources := map[string][]map[string]interface{}{
		"Condition":   {diabetesCondition()},
		"Observation": {hba1cObservation(7.5, "2025-06-01")},
	}
	period := MeasurePeriod{
		Start: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		End:   time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC),
	}

	report, err := eval.EvaluateIndividual(context.Background(), CMS122URL, patient, resources, period)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if report.Status != "complete" {
		t.Fatalf("expected status 'complete', got %s", report.Status)
	}
	if report.Type != "individual" {
		t.Fatalf("expected type 'individual', got %s", report.Type)
	}

	grp := report.Group[0]
	assertPopulationCount(t, grp, "initial-population", 1)
	assertPopulationCount(t, grp, "denominator", 1)
	assertPopulationCount(t, grp, "numerator", 1)
}

func TestMeasureEvaluator_DiabetesNotInNumerator(t *testing.T) {
	eval := newMeasureEvaluator()
	patient := diabeticPatient()
	resources := map[string][]map[string]interface{}{
		"Condition":   {diabetesCondition()},
		"Observation": {hba1cObservation(10.5, "2025-06-01")},
	}
	period := MeasurePeriod{
		Start: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		End:   time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC),
	}

	report, err := eval.EvaluateIndividual(context.Background(), CMS122URL, patient, resources, period)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	grp := report.Group[0]
	assertPopulationCount(t, grp, "initial-population", 1)
	assertPopulationCount(t, grp, "denominator", 1)
	assertPopulationCount(t, grp, "numerator", 0)
}

func TestMeasureEvaluator_DiabetesExcludedNoDiabetes(t *testing.T) {
	eval := newMeasureEvaluator()
	patient := diabeticPatient()
	resources := map[string][]map[string]interface{}{
		"Condition":   {},
		"Observation": {hba1cObservation(7.5, "2025-06-01")},
	}
	period := MeasurePeriod{
		Start: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		End:   time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC),
	}

	report, err := eval.EvaluateIndividual(context.Background(), CMS122URL, patient, resources, period)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	grp := report.Group[0]
	assertPopulationCount(t, grp, "initial-population", 0)
	assertPopulationCount(t, grp, "denominator", 0)
	assertPopulationCount(t, grp, "numerator", 0)
}

// ===========================================================================
// Individual Evaluation Tests — Controlling High BP (CMS165)
// ===========================================================================

func TestMeasureEvaluator_BPControlled(t *testing.T) {
	eval := newMeasureEvaluator()
	patient := hypertensionPatient()
	resources := map[string][]map[string]interface{}{
		"Condition":   {hypertensionCondition()},
		"Observation": {bpObservation(120, 80, "2025-06-01")},
	}
	period := MeasurePeriod{
		Start: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		End:   time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC),
	}

	report, err := eval.EvaluateIndividual(context.Background(), CMS165URL, patient, resources, period)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	grp := report.Group[0]
	assertPopulationCount(t, grp, "initial-population", 1)
	assertPopulationCount(t, grp, "denominator", 1)
	assertPopulationCount(t, grp, "numerator", 1)
}

func TestMeasureEvaluator_BPUncontrolled(t *testing.T) {
	eval := newMeasureEvaluator()
	patient := hypertensionPatient()
	resources := map[string][]map[string]interface{}{
		"Condition":   {hypertensionCondition()},
		"Observation": {bpObservation(155, 95, "2025-06-01")},
	}
	period := MeasurePeriod{
		Start: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		End:   time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC),
	}

	report, err := eval.EvaluateIndividual(context.Background(), CMS165URL, patient, resources, period)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	grp := report.Group[0]
	assertPopulationCount(t, grp, "initial-population", 1)
	assertPopulationCount(t, grp, "numerator", 0)
}

// ===========================================================================
// Individual Evaluation Tests — Breast Cancer Screening (CMS125)
// ===========================================================================

func TestMeasureEvaluator_BreastCancerScreened(t *testing.T) {
	eval := newMeasureEvaluator()
	patient := femalePatient("1965-04-10")
	resources := map[string][]map[string]interface{}{
		"DiagnosticReport": {mammogramReport("2025-03-15")},
	}
	period := MeasurePeriod{
		Start: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		End:   time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC),
	}

	report, err := eval.EvaluateIndividual(context.Background(), CMS125URL, patient, resources, period)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	grp := report.Group[0]
	assertPopulationCount(t, grp, "initial-population", 1)
	assertPopulationCount(t, grp, "numerator", 1)
}

func TestMeasureEvaluator_BreastCancerNotScreened(t *testing.T) {
	eval := newMeasureEvaluator()
	patient := femalePatient("1965-04-10")
	resources := map[string][]map[string]interface{}{
		"DiagnosticReport": {},
	}
	period := MeasurePeriod{
		Start: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		End:   time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC),
	}

	report, err := eval.EvaluateIndividual(context.Background(), CMS125URL, patient, resources, period)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	grp := report.Group[0]
	assertPopulationCount(t, grp, "initial-population", 1)
	assertPopulationCount(t, grp, "numerator", 0)
}

func TestMeasureEvaluator_BreastCancerMaleExcluded(t *testing.T) {
	eval := newMeasureEvaluator()
	patient := map[string]interface{}{
		"resourceType": "Patient",
		"id":           "pt-male",
		"gender":       "male",
		"birthDate":    "1965-04-10",
	}
	resources := map[string][]map[string]interface{}{
		"DiagnosticReport": {mammogramReport("2025-03-15")},
	}
	period := MeasurePeriod{
		Start: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		End:   time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC),
	}

	report, err := eval.EvaluateIndividual(context.Background(), CMS125URL, patient, resources, period)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	grp := report.Group[0]
	assertPopulationCount(t, grp, "initial-population", 0)
}

func TestMeasureEvaluator_IndividualReportStructure(t *testing.T) {
	eval := newMeasureEvaluator()
	patient := diabeticPatient()
	resources := map[string][]map[string]interface{}{
		"Condition":   {diabetesCondition()},
		"Observation": {hba1cObservation(7.5, "2025-06-01")},
	}
	period := MeasurePeriod{
		Start: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		End:   time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC),
	}

	report, err := eval.EvaluateIndividual(context.Background(), CMS122URL, patient, resources, period)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if report.ID == "" {
		t.Fatal("expected non-empty report ID")
	}
	if report.Status != "complete" {
		t.Fatalf("expected status 'complete', got %s", report.Status)
	}
	if report.Type != "individual" {
		t.Fatalf("expected type 'individual', got %s", report.Type)
	}
	if report.Measure != CMS122URL {
		t.Fatalf("expected measure URL %s, got %s", CMS122URL, report.Measure)
	}
	if report.Subject == nil || *report.Subject != "Patient/pt-diab-1" {
		subj := "<nil>"
		if report.Subject != nil {
			subj = *report.Subject
		}
		t.Fatalf("expected subject 'Patient/pt-diab-1', got %s", subj)
	}
	if len(report.Group) == 0 {
		t.Fatal("expected at least one group in report")
	}
}

func TestMeasureEvaluator_PeriodFiltering(t *testing.T) {
	eval := newMeasureEvaluator()
	patient := diabeticPatient()
	resources := map[string][]map[string]interface{}{
		"Condition":   {diabetesCondition()},
		"Observation": {hba1cObservation(7.5, "2023-06-01")},
	}
	period := MeasurePeriod{
		Start: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		End:   time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC),
	}

	report, err := eval.EvaluateIndividual(context.Background(), CMS122URL, patient, resources, period)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	grp := report.Group[0]
	assertPopulationCount(t, grp, "initial-population", 1)
	assertPopulationCount(t, grp, "numerator", 0)
}

// ===========================================================================
// Population Evaluation Tests
// ===========================================================================

func TestMeasureEvaluator_PopulationThreePatients(t *testing.T) {
	eval := newMeasureEvaluator()
	patients := []PatientBundle{
		{
			Patient: diabeticPatient(),
			Resources: map[string][]map[string]interface{}{
				"Condition":   {diabetesCondition()},
				"Observation": {hba1cObservation(7.5, "2025-06-01")},
			},
		},
		{
			Patient: func() map[string]interface{} {
				p := diabeticPatient()
				p["id"] = "pt-diab-2"
				return p
			}(),
			Resources: map[string][]map[string]interface{}{
				"Condition":   {diabetesCondition()},
				"Observation": {hba1cObservation(10.5, "2025-06-01")},
			},
		},
		{
			Patient: func() map[string]interface{} {
				p := diabeticPatient()
				p["id"] = "pt-diab-3"
				return p
			}(),
			Resources: map[string][]map[string]interface{}{
				"Condition":   {diabetesCondition()},
				"Observation": {hba1cObservation(8.0, "2025-06-01")},
			},
		},
	}
	period := MeasurePeriod{
		Start: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		End:   time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC),
	}

	report, err := eval.EvaluatePopulation(context.Background(), CMS122URL, patients, period)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if report.Type != "summary" {
		t.Fatalf("expected type 'summary', got %s", report.Type)
	}
	grp := report.Group[0]
	assertPopulationCount(t, grp, "initial-population", 3)
	assertPopulationCount(t, grp, "denominator", 3)
	assertPopulationCount(t, grp, "numerator", 2)
}

func TestMeasureEvaluator_PopulationSummaryMeasureScore(t *testing.T) {
	eval := newMeasureEvaluator()
	patients := []PatientBundle{
		{
			Patient: diabeticPatient(),
			Resources: map[string][]map[string]interface{}{
				"Condition":   {diabetesCondition()},
				"Observation": {hba1cObservation(7.5, "2025-06-01")},
			},
		},
		{
			Patient: func() map[string]interface{} {
				p := diabeticPatient()
				p["id"] = "pt-diab-2"
				return p
			}(),
			Resources: map[string][]map[string]interface{}{
				"Condition":   {diabetesCondition()},
				"Observation": {hba1cObservation(10.5, "2025-06-01")},
			},
		},
	}
	period := MeasurePeriod{
		Start: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		End:   time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC),
	}

	report, err := eval.EvaluatePopulation(context.Background(), CMS122URL, patients, period)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	grp := report.Group[0]
	if grp.MeasureScore == nil {
		t.Fatal("expected MeasureScore, got nil")
	}
	if *grp.MeasureScore != 0.5 {
		t.Fatalf("expected MeasureScore 0.5, got %f", *grp.MeasureScore)
	}
}

func TestMeasureEvaluator_PopulationSubjectListReport(t *testing.T) {
	eval := newMeasureEvaluator()
	patients := []PatientBundle{
		{
			Patient: diabeticPatient(),
			Resources: map[string][]map[string]interface{}{
				"Condition":   {diabetesCondition()},
				"Observation": {hba1cObservation(7.5, "2025-06-01")},
			},
		},
		{
			Patient: func() map[string]interface{} {
				p := diabeticPatient()
				p["id"] = "pt-diab-2"
				return p
			}(),
			Resources: map[string][]map[string]interface{}{
				"Condition":   {diabetesCondition()},
				"Observation": {hba1cObservation(10.5, "2025-06-01")},
			},
		},
	}
	period := MeasurePeriod{
		Start: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		End:   time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC),
	}

	report, err := eval.EvaluatePopulation(context.Background(), CMS122URL, patients, period)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	grp := report.Group[0]
	for _, pop := range grp.Population {
		if pop.Code == "initial-population" {
			if len(pop.SubjectResults) != 2 {
				t.Fatalf("expected 2 subject results for IP, got %d", len(pop.SubjectResults))
			}
		}
	}
}

func TestMeasureEvaluator_PopulationEmpty(t *testing.T) {
	eval := newMeasureEvaluator()
	patients := []PatientBundle{}
	period := MeasurePeriod{
		Start: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		End:   time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC),
	}

	report, err := eval.EvaluatePopulation(context.Background(), CMS122URL, patients, period)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	grp := report.Group[0]
	assertPopulationCount(t, grp, "initial-population", 0)
	assertPopulationCount(t, grp, "denominator", 0)
	assertPopulationCount(t, grp, "numerator", 0)
}

// ===========================================================================
// Handler Tests
// ===========================================================================

func newTestMeasureHandler() (*MeasureHandler, *echo.Echo) {
	eval := newMeasureEvaluator()
	h := NewMeasureHandler(eval)
	e := echo.New()
	g := e.Group("/fhir")
	h.RegisterRoutes(g)
	return h, e
}

func TestMeasureHandler_ListMeasures(t *testing.T) {
	_, e := newTestMeasureHandler()
	req := httptest.NewRequest(http.MethodGet, "/fhir/Measure", nil)
	req.Header.Set("Content-Type", "application/fhir+json")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var result map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("json unmarshal error: %v", err)
	}
	if result["resourceType"] != "Bundle" {
		t.Fatalf("expected Bundle, got %v", result["resourceType"])
	}
}

func TestMeasureHandler_GetMeasure(t *testing.T) {
	_, e := newTestMeasureHandler()
	req := httptest.NewRequest(http.MethodGet, "/fhir/Measure/cms122", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var result map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("json unmarshal error: %v", err)
	}
	if result["resourceType"] != "Measure" {
		t.Fatalf("expected Measure, got %v", result["resourceType"])
	}
	if result["id"] != "cms122" {
		t.Fatalf("expected id 'cms122', got %v", result["id"])
	}
}

func TestMeasureHandler_RegisterCustomMeasure(t *testing.T) {
	_, e := newTestMeasureHandler()

	body := `{
		"resourceType": "Measure",
		"id": "custom-1",
		"url": "http://example.org/fhir/Measure/custom-1",
		"name": "CustomMeasure",
		"title": "Custom Test Measure",
		"status": "draft",
		"scoring": {"coding": [{"code": "proportion"}]},
		"group": [{
			"population": [
				{"code": {"coding": [{"code": "initial-population"}]}, "criteria": {"expression": "InitialPopulation"}},
				{"code": {"coding": [{"code": "denominator"}]}, "criteria": {"expression": "Denominator"}},
				{"code": {"coding": [{"code": "numerator"}]}, "criteria": {"expression": "Numerator"}}
			]
		}]
	}`
	req := httptest.NewRequest(http.MethodPost, "/fhir/Measure", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/fhir+json")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestMeasureHandler_EvaluateMeasureIndividual(t *testing.T) {
	_, e := newTestMeasureHandler()

	body := `{
		"resourceType": "Bundle",
		"type": "collection",
		"entry": [
			{"resource": {"resourceType": "Patient", "id": "pt-diab-1", "gender": "male", "birthDate": "1970-06-15"}},
			{"resource": {"resourceType": "Condition", "id": "cond-1", "code": {"coding": [{"system": "http://hl7.org/fhir/sid/icd-10-cm", "code": "E11.9"}]}, "clinicalStatus": {"coding": [{"code": "active"}]}}},
			{"resource": {"resourceType": "Observation", "id": "obs-1", "status": "final", "code": {"coding": [{"system": "http://loinc.org", "code": "4548-4"}]}, "effectiveDateTime": "2025-06-01", "valueQuantity": {"value": 7.5, "unit": "%"}}}
		]
	}`

	req := httptest.NewRequest(http.MethodPost, "/fhir/Measure/cms122/$evaluate-measure?periodStart=2025-01-01&periodEnd=2025-12-31&reportType=individual", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/fhir+json")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var result map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("json unmarshal error: %v", err)
	}
	if result["resourceType"] != "MeasureReport" {
		t.Fatalf("expected MeasureReport, got %v", result["resourceType"])
	}
}

func TestMeasureHandler_EvaluateMeasurePopulation(t *testing.T) {
	_, e := newTestMeasureHandler()

	body := `{
		"resourceType": "Bundle",
		"type": "collection",
		"entry": [
			{"resource": {"resourceType": "Patient", "id": "pt-1", "gender": "male", "birthDate": "1970-06-15"}},
			{"resource": {"resourceType": "Condition", "id": "cond-1", "subject": {"reference": "Patient/pt-1"}, "code": {"coding": [{"system": "http://hl7.org/fhir/sid/icd-10-cm", "code": "E11.9"}]}, "clinicalStatus": {"coding": [{"code": "active"}]}}},
			{"resource": {"resourceType": "Observation", "id": "obs-1", "subject": {"reference": "Patient/pt-1"}, "status": "final", "code": {"coding": [{"system": "http://loinc.org", "code": "4548-4"}]}, "effectiveDateTime": "2025-06-01", "valueQuantity": {"value": 7.5, "unit": "%"}}},
			{"resource": {"resourceType": "Patient", "id": "pt-2", "gender": "male", "birthDate": "1968-03-20"}},
			{"resource": {"resourceType": "Condition", "id": "cond-2", "subject": {"reference": "Patient/pt-2"}, "code": {"coding": [{"system": "http://hl7.org/fhir/sid/icd-10-cm", "code": "E11.9"}]}, "clinicalStatus": {"coding": [{"code": "active"}]}}},
			{"resource": {"resourceType": "Observation", "id": "obs-2", "subject": {"reference": "Patient/pt-2"}, "status": "final", "code": {"coding": [{"system": "http://loinc.org", "code": "4548-4"}]}, "effectiveDateTime": "2025-06-01", "valueQuantity": {"value": 10.5, "unit": "%"}}}
		]
	}`

	req := httptest.NewRequest(http.MethodPost, "/fhir/Measure/cms122/$evaluate-measure?periodStart=2025-01-01&periodEnd=2025-12-31&reportType=summary", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/fhir+json")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var result map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("json unmarshal error: %v", err)
	}
	if result["resourceType"] != "MeasureReport" {
		t.Fatalf("expected MeasureReport, got %v", result["resourceType"])
	}
	if result["type"] != "summary" {
		t.Fatalf("expected type 'summary', got %v", result["type"])
	}
}

func TestMeasureHandler_ListReports(t *testing.T) {
	h, e := newTestMeasureHandler()
	patient := diabeticPatient()
	resources := map[string][]map[string]interface{}{
		"Condition":   {diabetesCondition()},
		"Observation": {hba1cObservation(7.5, "2025-06-01")},
	}
	period := MeasurePeriod{
		Start: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		End:   time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC),
	}
	_, err := h.evaluator.EvaluateIndividual(context.Background(), CMS122URL, patient, resources, period)
	if err != nil {
		t.Fatalf("evaluation failed: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/fhir/MeasureReport", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var bundle map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &bundle); err != nil {
		t.Fatalf("json unmarshal error: %v", err)
	}
	if bundle["resourceType"] != "Bundle" {
		t.Fatalf("expected Bundle, got %v", bundle["resourceType"])
	}
}

func TestMeasureHandler_GetReport(t *testing.T) {
	h, e := newTestMeasureHandler()
	patient := diabeticPatient()
	resources := map[string][]map[string]interface{}{
		"Condition":   {diabetesCondition()},
		"Observation": {hba1cObservation(7.5, "2025-06-01")},
	}
	period := MeasurePeriod{
		Start: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		End:   time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC),
	}
	report, err := h.evaluator.EvaluateIndividual(context.Background(), CMS122URL, patient, resources, period)
	if err != nil {
		t.Fatalf("evaluation failed: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/fhir/MeasureReport/"+report.ID, nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestMeasureHandler_RegisterLibrary(t *testing.T) {
	_, e := newTestMeasureHandler()

	body := `{
		"resourceType": "Library",
		"url": "http://example.org/fhir/Library/test-lib",
		"name": "TestLib",
		"version": "1.0.0",
		"status": "active",
		"content": [{
			"contentType": "text/cql",
			"data": "library TestLib version '1.0.0'\ndefine IsMale: Patient.gender = 'male'"
		}]
	}`
	req := httptest.NewRequest(http.MethodPost, "/fhir/Library", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/fhir+json")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestMeasureHandler_InvalidMeasureID(t *testing.T) {
	_, e := newTestMeasureHandler()
	req := httptest.NewRequest(http.MethodGet, "/fhir/Measure/nonexistent", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestMeasureHandler_MissingPeriodParams(t *testing.T) {
	_, e := newTestMeasureHandler()

	body := `{"resourceType": "Bundle", "type": "collection", "entry": []}`
	req := httptest.NewRequest(http.MethodPost, "/fhir/Measure/cms122/$evaluate-measure", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/fhir+json")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing period params, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestMeasureHandler_ListLibraries(t *testing.T) {
	_, e := newTestMeasureHandler()
	req := httptest.NewRequest(http.MethodGet, "/fhir/Library", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var bundle map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &bundle); err != nil {
		t.Fatalf("json unmarshal error: %v", err)
	}
	if bundle["resourceType"] != "Bundle" {
		t.Fatalf("expected Bundle, got %v", bundle["resourceType"])
	}
}

// ===========================================================================
// Edge Case Tests
// ===========================================================================

func TestMeasureEvaluator_MissingLibraryForMeasure(t *testing.T) {
	eval := NewMeasureEvaluator()
	m := &Measure{
		ID:      "test-missing-lib",
		URL:     "http://example.org/fhir/Measure/test-missing-lib",
		Status:  "active",
		Scoring: "proportion",
		Library: []string{"http://nonexistent/Library/nothing"},
		Group: []MeasureGroup{
			{
				Population: []MeasurePopulation{
					{Code: "initial-population", Expression: "InitialPopulation"},
				},
			},
		},
	}
	eval.RegisterMeasure(m)

	patient := diabeticPatient()
	resources := map[string][]map[string]interface{}{}
	period := MeasurePeriod{
		Start: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		End:   time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC),
	}

	report, err := eval.EvaluateIndividual(context.Background(), "http://example.org/fhir/Measure/test-missing-lib", patient, resources, period)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if report.Status != "complete" {
		t.Fatalf("expected complete, got %s", report.Status)
	}
}

func TestMeasureEvaluator_ExpressionError(t *testing.T) {
	eval := NewMeasureEvaluator()
	m := &Measure{
		ID:      "test-expr-err",
		URL:     "http://example.org/fhir/Measure/test-expr-err",
		Status:  "active",
		Scoring: "proportion",
		Group: []MeasureGroup{
			{
				Population: []MeasurePopulation{
					{Code: "initial-population", Expression: "!!!bad-expr!!!"},
				},
			},
		},
	}
	eval.RegisterMeasure(m)

	patient := diabeticPatient()
	resources := map[string][]map[string]interface{}{}
	period := MeasurePeriod{
		Start: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		End:   time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC),
	}

	report, err := eval.EvaluateIndividual(context.Background(), "http://example.org/fhir/Measure/test-expr-err", patient, resources, period)
	if err != nil {
		t.Fatalf("unexpected error (should handle gracefully): %v", err)
	}
	if report.Status != "complete" {
		t.Fatalf("expected complete (with zero counts), got %s", report.Status)
	}
}

func TestMeasureEvaluator_PatientNoMatchingData(t *testing.T) {
	eval := newMeasureEvaluator()
	patient := map[string]interface{}{
		"resourceType": "Patient",
		"id":           "pt-empty",
		"gender":       "male",
		"birthDate":    "1990-01-01",
	}
	resources := map[string][]map[string]interface{}{}
	period := MeasurePeriod{
		Start: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		End:   time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC),
	}

	report, err := eval.EvaluateIndividual(context.Background(), CMS122URL, patient, resources, period)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	grp := report.Group[0]
	assertPopulationCount(t, grp, "initial-population", 0)
	assertPopulationCount(t, grp, "numerator", 0)
}

func TestMeasureEvaluator_ZeroDenominator(t *testing.T) {
	eval := newMeasureEvaluator()
	patients := []PatientBundle{
		{
			Patient: map[string]interface{}{
				"resourceType": "Patient",
				"id":           "pt-no-match",
				"gender":       "male",
				"birthDate":    "2020-01-01",
			},
			Resources: map[string][]map[string]interface{}{},
		},
	}
	period := MeasurePeriod{
		Start: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		End:   time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC),
	}

	report, err := eval.EvaluatePopulation(context.Background(), CMS122URL, patients, period)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	grp := report.Group[0]
	if grp.MeasureScore != nil {
		t.Fatalf("expected nil MeasureScore for zero denominator, got %v", *grp.MeasureScore)
	}
}

func TestMeasureEvaluator_ConcurrentEvaluationSafety(t *testing.T) {
	eval := newMeasureEvaluator()
	period := MeasurePeriod{
		Start: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		End:   time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC),
	}

	var wg sync.WaitGroup
	errs := make(chan error, 20)

	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			patient := diabeticPatient()
			patient["id"] = "pt-concurrent-" + string(rune('A'+idx))
			resources := map[string][]map[string]interface{}{
				"Condition":   {diabetesCondition()},
				"Observation": {hba1cObservation(7.5, "2025-06-01")},
			}
			_, err := eval.EvaluateIndividual(context.Background(), CMS122URL, patient, resources, period)
			if err != nil {
				errs <- err
			}
		}(i)
	}
	wg.Wait()
	close(errs)

	for err := range errs {
		t.Fatalf("concurrent evaluation error: %v", err)
	}
}

func TestMeasureEvaluator_InvalidMeasureURL(t *testing.T) {
	eval := newMeasureEvaluator()
	patient := diabeticPatient()
	resources := map[string][]map[string]interface{}{}
	period := MeasurePeriod{
		Start: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		End:   time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC),
	}

	_, err := eval.EvaluateIndividual(context.Background(), "http://nonexistent/Measure/fake", patient, resources, period)
	if err == nil {
		t.Fatal("expected error for invalid measure URL")
	}
}

func TestMeasureReport_ToFHIR(t *testing.T) {
	subj := "Patient/pt-1"
	score := 0.75
	report := &MeasureReport{
		ID:      "mr-1",
		Status:  "complete",
		Type:    "summary",
		Measure: CMS122URL,
		Subject: &subj,
		Period: MeasurePeriod{
			Start: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			End:   time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC),
		},
		Group: []MeasureReportGroup{
			{
				Population: []MeasureReportPopulation{
					{Code: "initial-population", Count: 10},
					{Code: "denominator", Count: 10},
					{Code: "numerator", Count: 7},
				},
				MeasureScore: &score,
			},
		},
	}

	fhirJSON := report.ToFHIR()
	data, err := json.Marshal(fhirJSON)
	if err != nil {
		t.Fatalf("json marshal error: %v", err)
	}
	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("json unmarshal error: %v", err)
	}
	if parsed["resourceType"] != "MeasureReport" {
		t.Fatalf("expected MeasureReport, got %v", parsed["resourceType"])
	}
	if parsed["status"] != "complete" {
		t.Fatalf("expected complete, got %v", parsed["status"])
	}
}

func TestMeasureEvaluator_DiabetesHospiceExclusion(t *testing.T) {
	eval := newMeasureEvaluator()
	patient := diabeticPatient()
	resources := map[string][]map[string]interface{}{
		"Condition":   {diabetesCondition()},
		"Observation": {hba1cObservation(7.5, "2025-06-01")},
		"Encounter":   {hospiceEncounter()},
	}
	period := MeasurePeriod{
		Start: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		End:   time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC),
	}

	report, err := eval.EvaluateIndividual(context.Background(), CMS122URL, patient, resources, period)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	grp := report.Group[0]
	assertPopulationCount(t, grp, "denominator-exclusion", 1)
}

func TestMeasureEvaluator_BreastCancerTooYoung(t *testing.T) {
	eval := newMeasureEvaluator()
	patient := femalePatient("2000-01-01")
	resources := map[string][]map[string]interface{}{
		"DiagnosticReport": {mammogramReport("2025-03-15")},
	}
	period := MeasurePeriod{
		Start: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		End:   time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC),
	}

	report, err := eval.EvaluateIndividual(context.Background(), CMS125URL, patient, resources, period)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	grp := report.Group[0]
	assertPopulationCount(t, grp, "initial-population", 0)
}

func TestMeasureEvaluator_BPNoHypertension(t *testing.T) {
	eval := newMeasureEvaluator()
	patient := hypertensionPatient()
	resources := map[string][]map[string]interface{}{
		"Condition":   {},
		"Observation": {bpObservation(120, 80, "2025-06-01")},
	}
	period := MeasurePeriod{
		Start: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		End:   time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC),
	}

	report, err := eval.EvaluateIndividual(context.Background(), CMS165URL, patient, resources, period)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	grp := report.Group[0]
	assertPopulationCount(t, grp, "initial-population", 0)
}

func TestCQLLibrary_Creation(t *testing.T) {
	lib := &CQLLibrary{
		Name:      "TestLib",
		Version:   "1.0.0",
		URL:       "http://example.org/fhir/Library/test",
		Status:    "active",
		CreatedAt: time.Now(),
		Parameters: []CQLParameter{
			{Name: "MeasurePeriod", Type: "Period"},
		},
		Definitions: map[string]CQLDefinition{
			"IP": {
				Name:       "IP",
				Expression: "Patient.active = true",
				Context:    "Patient",
				Type:       "Boolean",
			},
		},
	}
	if lib.Name != "TestLib" {
		t.Fatalf("expected name TestLib, got %s", lib.Name)
	}
	if len(lib.Parameters) != 1 {
		t.Fatalf("expected 1 parameter, got %d", len(lib.Parameters))
	}
	if len(lib.Definitions) != 1 {
		t.Fatalf("expected 1 definition, got %d", len(lib.Definitions))
	}
}

func TestMeasureHandler_EvaluateInvalidMeasure(t *testing.T) {
	_, e := newTestMeasureHandler()

	body := `{"resourceType": "Bundle", "type": "collection", "entry": []}`
	req := httptest.NewRequest(http.MethodPost, "/fhir/Measure/nonexistent/$evaluate-measure?periodStart=2025-01-01&periodEnd=2025-12-31", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/fhir+json")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

// ===========================================================================
// Test helpers
// ===========================================================================

func assertPopulationCount(t *testing.T, grp MeasureReportGroup, code string, expected int) {
	t.Helper()
	for _, pop := range grp.Population {
		if pop.Code == code {
			if pop.Count != expected {
				t.Fatalf("expected %s count %d, got %d", code, expected, pop.Count)
			}
			return
		}
	}
	if expected != 0 {
		t.Fatalf("population code %q not found in report group", code)
	}
}
