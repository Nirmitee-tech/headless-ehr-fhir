package ccda

import (
	"encoding/xml"
	"strings"
	"testing"
)

// =========== Test Data Helpers ===========

func testPatient() map[string]interface{} {
	return map[string]interface{}{
		"resourceType": "Patient",
		"id":           "patient-123",
		"name": []interface{}{
			map[string]interface{}{
				"given":  []interface{}{"John"},
				"family": "Doe",
			},
		},
		"gender":    "male",
		"birthDate": "1980-01-15",
		"address": []interface{}{
			map[string]interface{}{
				"use":        "home",
				"line":       []interface{}{"123 Main St"},
				"city":       "Springfield",
				"state":      "IL",
				"postalCode": "62704",
				"country":    "US",
			},
		},
		"telecom": []interface{}{
			map[string]interface{}{
				"use":   "home",
				"value": "tel:+1-555-555-1234",
			},
		},
	}
}

func testAllergies() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"resourceType": "AllergyIntolerance",
			"id":           "allergy-1",
			"code": map[string]interface{}{
				"coding": []interface{}{
					map[string]interface{}{
						"system":  "http://snomed.info/sct",
						"code":    "387517004",
						"display": "Penicillin",
					},
				},
			},
			"clinicalStatus": map[string]interface{}{
				"coding": []interface{}{
					map[string]interface{}{
						"code": "active",
					},
				},
			},
			"reaction": []interface{}{
				map[string]interface{}{
					"manifestation": []interface{}{
						map[string]interface{}{
							"coding": []interface{}{
								map[string]interface{}{
									"display": "Hives",
								},
							},
						},
					},
				},
			},
		},
	}
}

func testMedications() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"resourceType": "MedicationRequest",
			"id":           "med-1",
			"status":       "active",
			"medicationCodeableConcept": map[string]interface{}{
				"coding": []interface{}{
					map[string]interface{}{
						"system":  "http://www.nlm.nih.gov/research/umls/rxnorm",
						"code":    "197361",
						"display": "Lisinopril 10 MG",
					},
				},
			},
			"dosageInstruction": []interface{}{
				map[string]interface{}{
					"text": "Take once daily",
				},
			},
		},
	}
}

func testConditions() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"resourceType": "Condition",
			"id":           "cond-1",
			"code": map[string]interface{}{
				"coding": []interface{}{
					map[string]interface{}{
						"system":  "http://snomed.info/sct",
						"code":    "38341003",
						"display": "Essential hypertension",
					},
				},
			},
			"clinicalStatus": map[string]interface{}{
				"coding": []interface{}{
					map[string]interface{}{
						"code": "active",
					},
				},
			},
			"onsetDateTime": "2020-03-15",
		},
	}
}

func testProcedures() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"resourceType":     "Procedure",
			"id":               "proc-1",
			"status":           "completed",
			"performedDateTime": "2023-06-20",
			"code": map[string]interface{}{
				"coding": []interface{}{
					map[string]interface{}{
						"system":  "http://snomed.info/sct",
						"code":    "80146002",
						"display": "Appendectomy",
					},
				},
			},
		},
	}
}

func testResults() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"resourceType":     "Observation",
			"id":               "obs-1",
			"status":           "final",
			"effectiveDateTime": "2024-01-10",
			"code": map[string]interface{}{
				"coding": []interface{}{
					map[string]interface{}{
						"system":  "http://loinc.org",
						"code":    "2093-3",
						"display": "Total Cholesterol",
					},
				},
			},
			"valueQuantity": map[string]interface{}{
				"value": 195.0,
				"unit":  "mg/dL",
			},
		},
	}
}

func testVitalSigns() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"resourceType":     "Observation",
			"id":               "vital-1",
			"status":           "final",
			"effectiveDateTime": "2024-01-10",
			"code": map[string]interface{}{
				"coding": []interface{}{
					map[string]interface{}{
						"system":  "http://loinc.org",
						"code":    "8480-6",
						"display": "Systolic Blood Pressure",
					},
				},
			},
			"valueQuantity": map[string]interface{}{
				"value": 120.0,
				"unit":  "mmHg",
			},
		},
	}
}

func testImmunizations() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"resourceType":       "Immunization",
			"id":                 "imm-1",
			"status":             "completed",
			"occurrenceDateTime": "2023-10-01",
			"vaccineCode": map[string]interface{}{
				"coding": []interface{}{
					map[string]interface{}{
						"system":  "http://hl7.org/fhir/sid/cvx",
						"code":    "141",
						"display": "Influenza Vaccine",
					},
				},
			},
		},
	}
}

func testEncounters() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"resourceType": "Encounter",
			"id":           "enc-1",
			"status":       "finished",
			"type": []interface{}{
				map[string]interface{}{
					"coding": []interface{}{
						map[string]interface{}{
							"system":  "http://snomed.info/sct",
							"code":    "185349003",
							"display": "Office Visit",
						},
					},
				},
			},
			"period": map[string]interface{}{
				"start": "2024-01-10",
			},
		},
	}
}

func testSocialHistory() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"resourceType":     "Observation",
			"id":               "social-1",
			"effectiveDateTime": "2024-01-10",
			"code": map[string]interface{}{
				"coding": []interface{}{
					map[string]interface{}{
						"system":  "http://loinc.org",
						"code":    "72166-2",
						"display": "Tobacco smoking status",
					},
				},
			},
			"valueCodeableConcept": map[string]interface{}{
				"coding": []interface{}{
					map[string]interface{}{
						"display": "Never smoker",
					},
				},
			},
		},
	}
}

func testCarePlans() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"resourceType": "CarePlan",
			"id":           "cp-1",
			"status":       "active",
			"title":        "Hypertension Management Plan",
			"period": map[string]interface{}{
				"start": "2024-01-10",
			},
		},
	}
}

func fullPatientData() *PatientData {
	return &PatientData{
		Patient:       testPatient(),
		Allergies:     testAllergies(),
		Medications:   testMedications(),
		Conditions:    testConditions(),
		Procedures:    testProcedures(),
		Results:       testResults(),
		VitalSigns:    testVitalSigns(),
		Immunizations: testImmunizations(),
		Encounters:    testEncounters(),
		SocialHistory: testSocialHistory(),
		CarePlans:     testCarePlans(),
	}
}

// =========== Generator Tests ===========

func TestGenerator_GenerateCCD_BasicPatient(t *testing.T) {
	gen := NewGenerator("Test Hospital", "2.16.840.1.113883.3.1234")

	data := &PatientData{
		Patient: testPatient(),
	}

	xmlData, err := gen.GenerateCCD(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	xmlStr := string(xmlData)

	// Should have XML declaration
	if !strings.HasPrefix(xmlStr, "<?xml") {
		t.Error("expected XML declaration at the start")
	}

	// Should have ClinicalDocument root
	if !strings.Contains(xmlStr, "ClinicalDocument") {
		t.Error("expected ClinicalDocument root element")
	}

	// Should have patient name
	if !strings.Contains(xmlStr, "John") {
		t.Error("expected patient given name 'John' in output")
	}
	if !strings.Contains(xmlStr, "Doe") {
		t.Error("expected patient family name 'Doe' in output")
	}

	// Should have custodian info
	if !strings.Contains(xmlStr, "Test Hospital") {
		t.Error("expected custodian organization name")
	}
}

func TestGenerator_GenerateCCD_AllSections(t *testing.T) {
	gen := NewGenerator("Test Hospital", "2.16.840.1.113883.3.1234")
	data := fullPatientData()

	xmlData, err := gen.GenerateCCD(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	xmlStr := string(xmlData)

	expectedSections := []struct {
		title string
		loinc string
	}{
		{"Allergies and Adverse Reactions", LOINCAllergies},
		{"Medications", LOINCMedications},
		{"Problems", LOINCProblems},
		{"Procedures", LOINCProcedures},
		{"Results", LOINCResults},
		{"Vital Signs", LOINCVitalSigns},
		{"Immunizations", LOINCImmunizations},
		{"Social History", LOINCSocialHistory},
		{"Plan of Care", LOINCPlanOfCare},
		{"Encounters", LOINCEncounters},
	}

	for _, es := range expectedSections {
		if !strings.Contains(xmlStr, es.title) {
			t.Errorf("expected section title %q in output", es.title)
		}
		if !strings.Contains(xmlStr, es.loinc) {
			t.Errorf("expected LOINC code %q in output", es.loinc)
		}
	}
}

func TestGenerator_GenerateCCD_EmptySections(t *testing.T) {
	gen := NewGenerator("Test Hospital", "2.16.840.1.113883.3.1234")

	data := &PatientData{
		Patient:   testPatient(),
		Allergies: []map[string]interface{}{},
	}

	xmlData, err := gen.GenerateCCD(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	xmlStr := string(xmlData)

	// Should have a header but no body sections
	if !strings.Contains(xmlStr, "ClinicalDocument") {
		t.Error("expected ClinicalDocument root element")
	}

	// No section-specific content should be present
	sectionTitles := []string{
		"Allergies and Adverse Reactions",
		"Medications",
		"Problems",
		"Procedures",
		"Results",
		"Vital Signs",
		"Immunizations",
		"Social History",
		"Plan of Care",
		"Encounters",
	}
	for _, title := range sectionTitles {
		if strings.Contains(xmlStr, "<title>"+title+"</title>") {
			t.Errorf("did not expect section %q for empty data", title)
		}
	}
}

func TestGenerator_AllergiesSection(t *testing.T) {
	gen := NewGenerator("Test Hospital", "2.16.840.1.113883.3.1234")

	data := &PatientData{
		Patient:   testPatient(),
		Allergies: testAllergies(),
	}

	xmlData, err := gen.GenerateCCD(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	xmlStr := string(xmlData)

	if !strings.Contains(xmlStr, "Penicillin") {
		t.Error("expected allergy substance 'Penicillin' in output")
	}
	if !strings.Contains(xmlStr, "Hives") {
		t.Error("expected allergy reaction 'Hives' in narrative")
	}
	if !strings.Contains(xmlStr, OIDAllergyEntry) {
		t.Error("expected allergy entry template ID")
	}
}

func TestGenerator_MedicationsSection(t *testing.T) {
	gen := NewGenerator("Test Hospital", "2.16.840.1.113883.3.1234")

	data := &PatientData{
		Patient:     testPatient(),
		Medications: testMedications(),
	}

	xmlData, err := gen.GenerateCCD(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	xmlStr := string(xmlData)

	if !strings.Contains(xmlStr, "Lisinopril") {
		t.Error("expected medication 'Lisinopril' in output")
	}
	if !strings.Contains(xmlStr, "197361") {
		t.Error("expected RxNorm code '197361' in output")
	}
	if !strings.Contains(xmlStr, OIDMedicationEntry) {
		t.Error("expected medication entry template ID")
	}
}

func TestGenerator_ProblemsSection(t *testing.T) {
	gen := NewGenerator("Test Hospital", "2.16.840.1.113883.3.1234")

	data := &PatientData{
		Patient:    testPatient(),
		Conditions: testConditions(),
	}

	xmlData, err := gen.GenerateCCD(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	xmlStr := string(xmlData)

	if !strings.Contains(xmlStr, "Essential hypertension") {
		t.Error("expected condition 'Essential hypertension' in output")
	}
	if !strings.Contains(xmlStr, "38341003") {
		t.Error("expected SNOMED code '38341003' in output")
	}
	if !strings.Contains(xmlStr, OIDProblemEntry) {
		t.Error("expected problem entry template ID")
	}
}

func TestGenerator_ProceduresSection(t *testing.T) {
	gen := NewGenerator("Test Hospital", "2.16.840.1.113883.3.1234")

	data := &PatientData{
		Patient:    testPatient(),
		Procedures: testProcedures(),
	}

	xmlData, err := gen.GenerateCCD(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	xmlStr := string(xmlData)

	if !strings.Contains(xmlStr, "Appendectomy") {
		t.Error("expected procedure 'Appendectomy' in output")
	}
	if !strings.Contains(xmlStr, "80146002") {
		t.Error("expected SNOMED code '80146002' in output")
	}
	if !strings.Contains(xmlStr, OIDProcedureEntry) {
		t.Error("expected procedure entry template ID")
	}
}

func TestGenerator_ResultsSection(t *testing.T) {
	gen := NewGenerator("Test Hospital", "2.16.840.1.113883.3.1234")

	data := &PatientData{
		Patient: testPatient(),
		Results: testResults(),
	}

	xmlData, err := gen.GenerateCCD(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	xmlStr := string(xmlData)

	if !strings.Contains(xmlStr, "Total Cholesterol") {
		t.Error("expected result 'Total Cholesterol' in output")
	}
	if !strings.Contains(xmlStr, "195") {
		t.Error("expected result value '195' in output")
	}
	if !strings.Contains(xmlStr, "mg/dL") {
		t.Error("expected unit 'mg/dL' in output")
	}
	if !strings.Contains(xmlStr, OIDResultEntry) {
		t.Error("expected result entry template ID")
	}
}

func TestGenerator_VitalSignsSection(t *testing.T) {
	gen := NewGenerator("Test Hospital", "2.16.840.1.113883.3.1234")

	data := &PatientData{
		Patient:    testPatient(),
		VitalSigns: testVitalSigns(),
	}

	xmlData, err := gen.GenerateCCD(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	xmlStr := string(xmlData)

	if !strings.Contains(xmlStr, "Systolic Blood Pressure") {
		t.Error("expected vital sign 'Systolic Blood Pressure' in output")
	}
	if !strings.Contains(xmlStr, "120") {
		t.Error("expected value '120' in output")
	}
	if !strings.Contains(xmlStr, "mmHg") {
		t.Error("expected unit 'mmHg' in output")
	}
	if !strings.Contains(xmlStr, OIDVitalSignEntry) {
		t.Error("expected vital sign entry template ID")
	}
}

func TestGenerator_ImmunizationsSection(t *testing.T) {
	gen := NewGenerator("Test Hospital", "2.16.840.1.113883.3.1234")

	data := &PatientData{
		Patient:       testPatient(),
		Immunizations: testImmunizations(),
	}

	xmlData, err := gen.GenerateCCD(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	xmlStr := string(xmlData)

	if !strings.Contains(xmlStr, "Influenza Vaccine") {
		t.Error("expected immunization 'Influenza Vaccine' in output")
	}
	if !strings.Contains(xmlStr, "141") {
		t.Error("expected CVX code '141' in output")
	}
	if !strings.Contains(xmlStr, OIDImmunizationEntry) {
		t.Error("expected immunization entry template ID")
	}
}

func TestGenerator_EncountersSection(t *testing.T) {
	gen := NewGenerator("Test Hospital", "2.16.840.1.113883.3.1234")

	data := &PatientData{
		Patient:    testPatient(),
		Encounters: testEncounters(),
	}

	xmlData, err := gen.GenerateCCD(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	xmlStr := string(xmlData)

	if !strings.Contains(xmlStr, "Office Visit") {
		t.Error("expected encounter type 'Office Visit' in output")
	}
	if !strings.Contains(xmlStr, OIDEncounterEntry) {
		t.Error("expected encounter entry template ID")
	}
}

func TestGenerator_GenerateCCD_ValidXML(t *testing.T) {
	gen := NewGenerator("Test Hospital", "2.16.840.1.113883.3.1234")
	data := fullPatientData()

	xmlData, err := gen.GenerateCCD(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify it's parseable XML
	var doc ClinicalDocument
	if err := xml.Unmarshal(xmlData, &doc); err != nil {
		t.Fatalf("generated XML is not valid: %v", err)
	}
}

func TestGenerator_GenerateCCD_TemplateIDs(t *testing.T) {
	gen := NewGenerator("Test Hospital", "2.16.840.1.113883.3.1234")
	data := fullPatientData()

	xmlData, err := gen.GenerateCCD(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	xmlStr := string(xmlData)

	// Document-level template IDs
	if !strings.Contains(xmlStr, OIDUSRealmHeader) {
		t.Error("expected US Realm Header template ID")
	}
	if !strings.Contains(xmlStr, OIDCCDDocument) {
		t.Error("expected CCD Document template ID")
	}

	// Section-level template IDs
	sectionOIDs := []string{
		OIDAllergiesSection,
		OIDMedicationsSection,
		OIDProblemsSection,
		OIDProceduresSection,
		OIDResultsSection,
		OIDVitalSignsSection,
		OIDImmunizationsSection,
		OIDSocialHistorySection,
		OIDPlanOfCareSection,
		OIDEncountersSection,
	}

	for _, oid := range sectionOIDs {
		if !strings.Contains(xmlStr, oid) {
			t.Errorf("expected section template ID %q in output", oid)
		}
	}
}

func TestGenerator_GenerateCCD_PatientDemographics(t *testing.T) {
	gen := NewGenerator("Test Hospital", "2.16.840.1.113883.3.1234")

	data := &PatientData{
		Patient: testPatient(),
	}

	xmlData, err := gen.GenerateCCD(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	xmlStr := string(xmlData)

	// Name
	if !strings.Contains(xmlStr, "John") {
		t.Error("expected given name 'John'")
	}
	if !strings.Contains(xmlStr, "Doe") {
		t.Error("expected family name 'Doe'")
	}

	// Gender code
	if !strings.Contains(xmlStr, "administrativeGenderCode") {
		t.Error("expected administrativeGenderCode element")
	}
	if !strings.Contains(xmlStr, `code="M"`) {
		t.Error("expected gender code 'M' for male")
	}

	// Birth date (YYYYMMDD format)
	if !strings.Contains(xmlStr, "19800115") {
		t.Error("expected birth date '19800115'")
	}

	// Address
	if !strings.Contains(xmlStr, "123 Main St") {
		t.Error("expected street address")
	}
	if !strings.Contains(xmlStr, "Springfield") {
		t.Error("expected city 'Springfield'")
	}
	if !strings.Contains(xmlStr, "IL") {
		t.Error("expected state 'IL'")
	}
	if !strings.Contains(xmlStr, "62704") {
		t.Error("expected postal code '62704'")
	}

	// Patient ID
	if !strings.Contains(xmlStr, "patient-123") {
		t.Error("expected patient ID 'patient-123'")
	}
}

func TestGenerator_GenerateCCD_NilData(t *testing.T) {
	gen := NewGenerator("Test Hospital", "2.16.840.1.113883.3.1234")

	_, err := gen.GenerateCCD(nil)
	if err == nil {
		t.Error("expected error for nil data")
	}
}

func TestGenerator_GenerateCCD_NilPatient(t *testing.T) {
	gen := NewGenerator("Test Hospital", "2.16.840.1.113883.3.1234")

	_, err := gen.GenerateCCD(&PatientData{})
	if err == nil {
		t.Error("expected error for nil patient")
	}
}
