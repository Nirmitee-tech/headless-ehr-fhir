package fhir

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/labstack/echo/v4"
)

// ===========================================================================
// Profile Registry Tests
// ===========================================================================

func TestProfileRegistry_RegisterAndGetByURL(t *testing.T) {
	reg := NewProfileRegistry()
	p := ProfileDefinition{
		URL:     "http://example.com/StructureDefinition/test",
		Name:    "TestProfile",
		Type:    "Patient",
		Version: "1.0.0",
		Status:  "active",
	}
	reg.Register(p)

	got, ok := reg.GetByURL("http://example.com/StructureDefinition/test")
	if !ok {
		t.Fatal("expected to find registered profile")
	}
	if got.Name != "TestProfile" {
		t.Errorf("expected Name=TestProfile, got %s", got.Name)
	}
	if got.Type != "Patient" {
		t.Errorf("expected Type=Patient, got %s", got.Type)
	}
}

func TestProfileRegistry_GetByURL_NotFound(t *testing.T) {
	reg := NewProfileRegistry()
	_, ok := reg.GetByURL("http://example.com/nonexistent")
	if ok {
		t.Error("expected not found for unregistered URL")
	}
}

func TestProfileRegistry_GetByType(t *testing.T) {
	reg := NewProfileRegistry()
	reg.Register(ProfileDefinition{
		URL:  "http://example.com/SD/patient-1",
		Name: "PatientProfile1",
		Type: "Patient",
	})
	reg.Register(ProfileDefinition{
		URL:  "http://example.com/SD/patient-2",
		Name: "PatientProfile2",
		Type: "Patient",
	})
	reg.Register(ProfileDefinition{
		URL:  "http://example.com/SD/obs-1",
		Name: "ObsProfile",
		Type: "Observation",
	})

	patients := reg.GetByType("Patient")
	if len(patients) != 2 {
		t.Errorf("expected 2 Patient profiles, got %d", len(patients))
	}

	obs := reg.GetByType("Observation")
	if len(obs) != 1 {
		t.Errorf("expected 1 Observation profile, got %d", len(obs))
	}
}

func TestProfileRegistry_GetByType_Empty(t *testing.T) {
	reg := NewProfileRegistry()
	result := reg.GetByType("Unknown")
	if len(result) != 0 {
		t.Errorf("expected 0 profiles for unknown type, got %d", len(result))
	}
}

func TestProfileRegistry_ListAll(t *testing.T) {
	reg := NewProfileRegistry()
	reg.Register(ProfileDefinition{URL: "http://a.com/1", Type: "Patient"})
	reg.Register(ProfileDefinition{URL: "http://a.com/2", Type: "Observation"})

	all := reg.ListAll()
	if len(all) != 2 {
		t.Errorf("expected 2 profiles, got %d", len(all))
	}
}

func TestProfileRegistry_ListAll_Empty(t *testing.T) {
	reg := NewProfileRegistry()
	all := reg.ListAll()
	if len(all) != 0 {
		t.Errorf("expected 0 profiles, got %d", len(all))
	}
}

func TestProfileRegistry_MultipleProfilesPerType(t *testing.T) {
	reg := NewProfileRegistry()
	reg.Register(ProfileDefinition{URL: "http://a.com/cond-1", Type: "Condition", Name: "Cond1"})
	reg.Register(ProfileDefinition{URL: "http://a.com/cond-2", Type: "Condition", Name: "Cond2"})
	reg.Register(ProfileDefinition{URL: "http://a.com/cond-3", Type: "Condition", Name: "Cond3"})

	conds := reg.GetByType("Condition")
	if len(conds) != 3 {
		t.Errorf("expected 3 Condition profiles, got %d", len(conds))
	}
}

func TestProfileRegistry_RegisterCustomProfile(t *testing.T) {
	reg := NewProfileRegistry()
	custom := ProfileDefinition{
		URL:     "http://myorg.com/StructureDefinition/custom-patient",
		Name:    "CustomPatient",
		Type:    "Patient",
		Version: "1.0.0",
		Status:  "active",
		Constraints: []ProfileConstraint{
			{Path: "Patient.birthDate", Min: 1, Max: "*", MustSupport: true},
		},
	}
	reg.Register(custom)

	got, ok := reg.GetByURL("http://myorg.com/StructureDefinition/custom-patient")
	if !ok {
		t.Fatal("expected to find custom profile")
	}
	if len(got.Constraints) != 1 {
		t.Errorf("expected 1 constraint, got %d", len(got.Constraints))
	}
}

func TestProfileRegistry_DuplicateURL(t *testing.T) {
	reg := NewProfileRegistry()
	reg.Register(ProfileDefinition{URL: "http://a.com/dup", Name: "First", Type: "Patient"})
	reg.Register(ProfileDefinition{URL: "http://a.com/dup", Name: "Second", Type: "Patient"})

	got, ok := reg.GetByURL("http://a.com/dup")
	if !ok {
		t.Fatal("expected to find profile")
	}
	// The latest registration should win.
	if got.Name != "Second" {
		t.Errorf("expected Name=Second (latest registration wins), got %s", got.Name)
	}

	// Only one profile should be in the list for Patient
	all := reg.ListAll()
	count := 0
	for _, p := range all {
		if p.URL == "http://a.com/dup" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("expected 1 entry for duplicate URL, got %d", count)
	}
}

func TestProfileRegistry_ConcurrentAccess(t *testing.T) {
	reg := NewProfileRegistry()
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			reg.Register(ProfileDefinition{
				URL:  "http://a.com/conc-" + string(rune('A'+n%26)),
				Type: "Patient",
			})
			reg.GetByURL("http://a.com/conc-A")
			reg.GetByType("Patient")
			reg.ListAll()
		}(i)
	}
	wg.Wait()
}

// ===========================================================================
// US Core Patient Validation Tests
// ===========================================================================

func validUSCorePatient() map[string]interface{} {
	return map[string]interface{}{
		"resourceType": "Patient",
		"id":           "patient-1",
		"identifier": []interface{}{
			map[string]interface{}{
				"system": "http://hospital.example.org/mrn",
				"value":  "12345",
			},
		},
		"name": []interface{}{
			map[string]interface{}{
				"family": "Smith",
				"given":  []interface{}{"John"},
			},
		},
		"gender": "male",
	}
}

func TestProfileValidator_USCorePatient_Valid(t *testing.T) {
	reg := NewProfileRegistry()
	RegisterUSCoreProfiles(reg)
	v := NewProfileValidator(reg)

	issues := v.ValidateAgainstProfile(validUSCorePatient(), USCorePatientURL)
	errors := filterErrors(issues)
	if len(errors) > 0 {
		t.Errorf("expected no errors for valid US Core Patient, got: %+v", errors)
	}
}

func TestProfileValidator_USCorePatient_MissingIdentifier(t *testing.T) {
	reg := NewProfileRegistry()
	RegisterUSCoreProfiles(reg)
	v := NewProfileValidator(reg)

	resource := validUSCorePatient()
	delete(resource, "identifier")

	issues := v.ValidateAgainstProfile(resource, USCorePatientURL)
	if !hasErrorAtPath(issues, "Patient.identifier") {
		t.Errorf("expected error for missing identifier, got: %+v", issues)
	}
}

func TestProfileValidator_USCorePatient_MissingIdentifierSystem(t *testing.T) {
	reg := NewProfileRegistry()
	RegisterUSCoreProfiles(reg)
	v := NewProfileValidator(reg)

	resource := validUSCorePatient()
	resource["identifier"] = []interface{}{
		map[string]interface{}{
			"value": "12345",
			// system is missing
		},
	}

	issues := v.ValidateAgainstProfile(resource, USCorePatientURL)
	if !hasErrorAtPath(issues, "Patient.identifier.system") {
		t.Errorf("expected error for missing identifier.system, got: %+v", issues)
	}
}

func TestProfileValidator_USCorePatient_MissingIdentifierValue(t *testing.T) {
	reg := NewProfileRegistry()
	RegisterUSCoreProfiles(reg)
	v := NewProfileValidator(reg)

	resource := validUSCorePatient()
	resource["identifier"] = []interface{}{
		map[string]interface{}{
			"system": "http://hospital.example.org/mrn",
			// value is missing
		},
	}

	issues := v.ValidateAgainstProfile(resource, USCorePatientURL)
	if !hasErrorAtPath(issues, "Patient.identifier.value") {
		t.Errorf("expected error for missing identifier.value, got: %+v", issues)
	}
}

func TestProfileValidator_USCorePatient_MissingName(t *testing.T) {
	reg := NewProfileRegistry()
	RegisterUSCoreProfiles(reg)
	v := NewProfileValidator(reg)

	resource := validUSCorePatient()
	delete(resource, "name")

	issues := v.ValidateAgainstProfile(resource, USCorePatientURL)
	if !hasErrorAtPath(issues, "Patient.name") {
		t.Errorf("expected error for missing name, got: %+v", issues)
	}
}

func TestProfileValidator_USCorePatient_MissingGender(t *testing.T) {
	reg := NewProfileRegistry()
	RegisterUSCoreProfiles(reg)
	v := NewProfileValidator(reg)

	resource := validUSCorePatient()
	delete(resource, "gender")

	issues := v.ValidateAgainstProfile(resource, USCorePatientURL)
	if !hasErrorAtPath(issues, "Patient.gender") {
		t.Errorf("expected error for missing gender, got: %+v", issues)
	}
}

func TestProfileValidator_USCorePatient_NameWithOnlyFamily(t *testing.T) {
	reg := NewProfileRegistry()
	RegisterUSCoreProfiles(reg)
	v := NewProfileValidator(reg)

	resource := validUSCorePatient()
	resource["name"] = []interface{}{
		map[string]interface{}{
			"family": "Smith",
		},
	}

	issues := v.ValidateAgainstProfile(resource, USCorePatientURL)
	errors := filterErrors(issues)
	for _, e := range errors {
		if strings.Contains(e.Path, "name") && strings.Contains(e.Description, "family or given") {
			t.Errorf("name with family only should pass, but got: %+v", e)
		}
	}
}

func TestProfileValidator_USCorePatient_NameWithOnlyGiven(t *testing.T) {
	reg := NewProfileRegistry()
	RegisterUSCoreProfiles(reg)
	v := NewProfileValidator(reg)

	resource := validUSCorePatient()
	resource["name"] = []interface{}{
		map[string]interface{}{
			"given": []interface{}{"John"},
		},
	}

	issues := v.ValidateAgainstProfile(resource, USCorePatientURL)
	errors := filterErrors(issues)
	for _, e := range errors {
		if strings.Contains(e.Path, "name") && strings.Contains(e.Description, "family or given") {
			t.Errorf("name with given only should pass, but got: %+v", e)
		}
	}
}

func TestProfileValidator_USCorePatient_NameWithNeitherFamilyNorGiven(t *testing.T) {
	reg := NewProfileRegistry()
	RegisterUSCoreProfiles(reg)
	v := NewProfileValidator(reg)

	resource := validUSCorePatient()
	resource["name"] = []interface{}{
		map[string]interface{}{
			"use": "official",
			// no family, no given
		},
	}

	issues := v.ValidateAgainstProfile(resource, USCorePatientURL)
	found := false
	for _, issue := range issues {
		if issue.Severity == "error" && strings.Contains(issue.Description, "family or given") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected error when name has neither family nor given, got: %+v", issues)
	}
}

func TestProfileValidator_USCorePatient_ExtraFieldsDontFail(t *testing.T) {
	reg := NewProfileRegistry()
	RegisterUSCoreProfiles(reg)
	v := NewProfileValidator(reg)

	resource := validUSCorePatient()
	resource["maritalStatus"] = map[string]interface{}{"text": "Married"}
	resource["contact"] = []interface{}{map[string]interface{}{"name": map[string]interface{}{"family": "Doe"}}}
	resource["generalPractitioner"] = []interface{}{map[string]interface{}{"reference": "Practitioner/p1"}}

	issues := v.ValidateAgainstProfile(resource, USCorePatientURL)
	errors := filterErrors(issues)
	if len(errors) > 0 {
		t.Errorf("extra fields should not cause errors, got: %+v", errors)
	}
}

func TestProfileValidator_USCorePatient_MustSupportWarnings(t *testing.T) {
	reg := NewProfileRegistry()
	RegisterUSCoreProfiles(reg)
	v := NewProfileValidator(reg)

	// Valid patient but missing MustSupport fields: birthDate, address, telecom, communication
	resource := validUSCorePatient()

	issues := v.ValidateAgainstProfile(resource, USCorePatientURL)
	warnings := filterWarnings(issues)

	// Should have warnings for missing MustSupport fields
	mustSupportPaths := []string{"Patient.birthDate", "Patient.address", "Patient.telecom", "Patient.communication"}
	for _, path := range mustSupportPaths {
		found := false
		for _, w := range warnings {
			if w.Path == path {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected MustSupport warning for %s, got warnings: %+v", path, warnings)
		}
	}
}

func TestProfileValidator_USCorePatient_EmptyIdentifierArray(t *testing.T) {
	reg := NewProfileRegistry()
	RegisterUSCoreProfiles(reg)
	v := NewProfileValidator(reg)

	resource := validUSCorePatient()
	resource["identifier"] = []interface{}{}

	issues := v.ValidateAgainstProfile(resource, USCorePatientURL)
	if !hasErrorAtPath(issues, "Patient.identifier") {
		t.Errorf("expected error for empty identifier array, got: %+v", issues)
	}
}

func TestProfileValidator_USCorePatient_InvalidGender(t *testing.T) {
	reg := NewProfileRegistry()
	RegisterUSCoreProfiles(reg)
	v := NewProfileValidator(reg)

	resource := validUSCorePatient()
	resource["gender"] = "invalid-gender"

	issues := v.ValidateAgainstProfile(resource, USCorePatientURL)
	found := false
	for _, issue := range issues {
		if issue.Severity == "error" && strings.Contains(issue.Path, "gender") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected error for invalid gender value, got: %+v", issues)
	}
}

// ===========================================================================
// US Core Condition Validation Tests
// ===========================================================================

func validUSCoreCondition() map[string]interface{} {
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
		"category": []interface{}{
			map[string]interface{}{
				"coding": []interface{}{
					map[string]interface{}{
						"system": "http://terminology.hl7.org/CodeSystem/condition-category",
						"code":   "problem-list-item",
					},
				},
			},
		},
		"code": map[string]interface{}{
			"coding": []interface{}{
				map[string]interface{}{
					"system": "http://snomed.info/sct",
					"code":   "44054006",
				},
			},
		},
		"subject": map[string]interface{}{
			"reference": "Patient/patient-1",
		},
	}
}

func TestProfileValidator_USCoreCondition_Valid(t *testing.T) {
	reg := NewProfileRegistry()
	RegisterUSCoreProfiles(reg)
	v := NewProfileValidator(reg)

	issues := v.ValidateAgainstProfile(validUSCoreCondition(), USCoreConditionURL)
	errors := filterErrors(issues)
	if len(errors) > 0 {
		t.Errorf("expected no errors for valid US Core Condition, got: %+v", errors)
	}
}

func TestProfileValidator_USCoreCondition_MissingClinicalStatus(t *testing.T) {
	reg := NewProfileRegistry()
	RegisterUSCoreProfiles(reg)
	v := NewProfileValidator(reg)

	resource := validUSCoreCondition()
	delete(resource, "clinicalStatus")

	issues := v.ValidateAgainstProfile(resource, USCoreConditionURL)
	if !hasErrorAtPath(issues, "Condition.clinicalStatus") {
		t.Errorf("expected error for missing clinicalStatus, got: %+v", issues)
	}
}

func TestProfileValidator_USCoreCondition_MissingCategory(t *testing.T) {
	reg := NewProfileRegistry()
	RegisterUSCoreProfiles(reg)
	v := NewProfileValidator(reg)

	resource := validUSCoreCondition()
	delete(resource, "category")

	issues := v.ValidateAgainstProfile(resource, USCoreConditionURL)
	if !hasErrorAtPath(issues, "Condition.category") {
		t.Errorf("expected error for missing category, got: %+v", issues)
	}
}

func TestProfileValidator_USCoreCondition_MissingCode(t *testing.T) {
	reg := NewProfileRegistry()
	RegisterUSCoreProfiles(reg)
	v := NewProfileValidator(reg)

	resource := validUSCoreCondition()
	delete(resource, "code")

	issues := v.ValidateAgainstProfile(resource, USCoreConditionURL)
	if !hasErrorAtPath(issues, "Condition.code") {
		t.Errorf("expected error for missing code, got: %+v", issues)
	}
}

func TestProfileValidator_USCoreCondition_MissingSubject(t *testing.T) {
	reg := NewProfileRegistry()
	RegisterUSCoreProfiles(reg)
	v := NewProfileValidator(reg)

	resource := validUSCoreCondition()
	delete(resource, "subject")

	issues := v.ValidateAgainstProfile(resource, USCoreConditionURL)
	if !hasErrorAtPath(issues, "Condition.subject") {
		t.Errorf("expected error for missing subject, got: %+v", issues)
	}
}

func TestProfileValidator_USCoreCondition_EmptyCategory(t *testing.T) {
	reg := NewProfileRegistry()
	RegisterUSCoreProfiles(reg)
	v := NewProfileValidator(reg)

	resource := validUSCoreCondition()
	resource["category"] = []interface{}{}

	issues := v.ValidateAgainstProfile(resource, USCoreConditionURL)
	if !hasErrorAtPath(issues, "Condition.category") {
		t.Errorf("expected error for empty category, got: %+v", issues)
	}
}

func TestProfileValidator_USCoreCondition_MustSupportVerificationStatus(t *testing.T) {
	reg := NewProfileRegistry()
	RegisterUSCoreProfiles(reg)
	v := NewProfileValidator(reg)

	resource := validUSCoreCondition()
	// verificationStatus is MustSupport, missing should generate warning
	issues := v.ValidateAgainstProfile(resource, USCoreConditionURL)
	warnings := filterWarnings(issues)
	found := false
	for _, w := range warnings {
		if w.Path == "Condition.verificationStatus" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected warning for missing MustSupport verificationStatus, got: %+v", warnings)
	}
}

// ===========================================================================
// US Core Observation Lab Tests
// ===========================================================================

func validUSCoreObservationLab() map[string]interface{} {
	return map[string]interface{}{
		"resourceType": "Observation",
		"id":           "obs-lab-1",
		"status":       "final",
		"category": []interface{}{
			map[string]interface{}{
				"coding": []interface{}{
					map[string]interface{}{
						"system": "http://terminology.hl7.org/CodeSystem/observation-category",
						"code":   "laboratory",
					},
				},
			},
		},
		"code": map[string]interface{}{
			"coding": []interface{}{
				map[string]interface{}{
					"system": "http://loinc.org",
					"code":   "2339-0",
				},
			},
		},
		"subject": map[string]interface{}{
			"reference": "Patient/patient-1",
		},
		"valueQuantity": map[string]interface{}{
			"value":  120.0,
			"unit":   "mg/dL",
			"system": "http://unitsofmeasure.org",
		},
	}
}

func TestProfileValidator_USCoreObservationLab_Valid(t *testing.T) {
	reg := NewProfileRegistry()
	RegisterUSCoreProfiles(reg)
	v := NewProfileValidator(reg)

	issues := v.ValidateAgainstProfile(validUSCoreObservationLab(), USCoreObservationLabURL)
	errors := filterErrors(issues)
	if len(errors) > 0 {
		t.Errorf("expected no errors for valid US Core Observation Lab, got: %+v", errors)
	}
}

func TestProfileValidator_USCoreObservationLab_MissingStatus(t *testing.T) {
	reg := NewProfileRegistry()
	RegisterUSCoreProfiles(reg)
	v := NewProfileValidator(reg)

	resource := validUSCoreObservationLab()
	delete(resource, "status")

	issues := v.ValidateAgainstProfile(resource, USCoreObservationLabURL)
	if !hasErrorAtPath(issues, "Observation.status") {
		t.Errorf("expected error for missing status, got: %+v", issues)
	}
}

func TestProfileValidator_USCoreObservationLab_MissingCategory(t *testing.T) {
	reg := NewProfileRegistry()
	RegisterUSCoreProfiles(reg)
	v := NewProfileValidator(reg)

	resource := validUSCoreObservationLab()
	delete(resource, "category")

	issues := v.ValidateAgainstProfile(resource, USCoreObservationLabURL)
	if !hasErrorAtPath(issues, "Observation.category") {
		t.Errorf("expected error for missing category, got: %+v", issues)
	}
}

func TestProfileValidator_USCoreObservationLab_CategoryWithoutLaboratory(t *testing.T) {
	reg := NewProfileRegistry()
	RegisterUSCoreProfiles(reg)
	v := NewProfileValidator(reg)

	resource := validUSCoreObservationLab()
	resource["category"] = []interface{}{
		map[string]interface{}{
			"coding": []interface{}{
				map[string]interface{}{
					"system": "http://terminology.hl7.org/CodeSystem/observation-category",
					"code":   "vital-signs", // not "laboratory"
				},
			},
		},
	}

	issues := v.ValidateAgainstProfile(resource, USCoreObservationLabURL)
	found := false
	for _, issue := range issues {
		if issue.Severity == "error" && strings.Contains(issue.Path, "category") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected error for category without 'laboratory' code, got: %+v", issues)
	}
}

func TestProfileValidator_USCoreObservationLab_MissingCode(t *testing.T) {
	reg := NewProfileRegistry()
	RegisterUSCoreProfiles(reg)
	v := NewProfileValidator(reg)

	resource := validUSCoreObservationLab()
	delete(resource, "code")

	issues := v.ValidateAgainstProfile(resource, USCoreObservationLabURL)
	if !hasErrorAtPath(issues, "Observation.code") {
		t.Errorf("expected error for missing code, got: %+v", issues)
	}
}

func TestProfileValidator_USCoreObservationLab_MissingSubject(t *testing.T) {
	reg := NewProfileRegistry()
	RegisterUSCoreProfiles(reg)
	v := NewProfileValidator(reg)

	resource := validUSCoreObservationLab()
	delete(resource, "subject")

	issues := v.ValidateAgainstProfile(resource, USCoreObservationLabURL)
	if !hasErrorAtPath(issues, "Observation.subject") {
		t.Errorf("expected error for missing subject, got: %+v", issues)
	}
}

func TestProfileValidator_USCoreObservationLab_MissingValueWarning(t *testing.T) {
	reg := NewProfileRegistry()
	RegisterUSCoreProfiles(reg)
	v := NewProfileValidator(reg)

	resource := validUSCoreObservationLab()
	delete(resource, "valueQuantity")

	issues := v.ValidateAgainstProfile(resource, USCoreObservationLabURL)
	// Should have a warning for missing value[x] (MustSupport)
	warnings := filterWarnings(issues)
	found := false
	for _, w := range warnings {
		if strings.Contains(w.Path, "value") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected warning for missing MustSupport value[x], got warnings: %+v", warnings)
	}
}

func TestProfileValidator_USCoreObservationLab_MissingEffectiveWarning(t *testing.T) {
	reg := NewProfileRegistry()
	RegisterUSCoreProfiles(reg)
	v := NewProfileValidator(reg)

	resource := validUSCoreObservationLab()
	// effective[x] is not present — should get a warning

	issues := v.ValidateAgainstProfile(resource, USCoreObservationLabURL)
	warnings := filterWarnings(issues)
	found := false
	for _, w := range warnings {
		if strings.Contains(w.Path, "effective") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected warning for missing MustSupport effective[x], got warnings: %+v", warnings)
	}
}

// ===========================================================================
// US Core MedicationRequest Tests
// ===========================================================================

func validUSCoreMedicationRequest() map[string]interface{} {
	return map[string]interface{}{
		"resourceType": "MedicationRequest",
		"id":           "mr-1",
		"status":       "active",
		"intent":       "order",
		"medicationCodeableConcept": map[string]interface{}{
			"coding": []interface{}{
				map[string]interface{}{
					"system": "http://www.nlm.nih.gov/research/umls/rxnorm",
					"code":   "1049502",
				},
			},
		},
		"subject": map[string]interface{}{
			"reference": "Patient/patient-1",
		},
	}
}

func TestProfileValidator_USCoreMedicationRequest_Valid(t *testing.T) {
	reg := NewProfileRegistry()
	RegisterUSCoreProfiles(reg)
	v := NewProfileValidator(reg)

	issues := v.ValidateAgainstProfile(validUSCoreMedicationRequest(), USCoreMedicationRequestURL)
	errors := filterErrors(issues)
	if len(errors) > 0 {
		t.Errorf("expected no errors, got: %+v", errors)
	}
}

func TestProfileValidator_USCoreMedicationRequest_MissingStatus(t *testing.T) {
	reg := NewProfileRegistry()
	RegisterUSCoreProfiles(reg)
	v := NewProfileValidator(reg)

	resource := validUSCoreMedicationRequest()
	delete(resource, "status")

	issues := v.ValidateAgainstProfile(resource, USCoreMedicationRequestURL)
	if !hasErrorAtPath(issues, "MedicationRequest.status") {
		t.Errorf("expected error for missing status, got: %+v", issues)
	}
}

func TestProfileValidator_USCoreMedicationRequest_MissingIntent(t *testing.T) {
	reg := NewProfileRegistry()
	RegisterUSCoreProfiles(reg)
	v := NewProfileValidator(reg)

	resource := validUSCoreMedicationRequest()
	delete(resource, "intent")

	issues := v.ValidateAgainstProfile(resource, USCoreMedicationRequestURL)
	if !hasErrorAtPath(issues, "MedicationRequest.intent") {
		t.Errorf("expected error for missing intent, got: %+v", issues)
	}
}

func TestProfileValidator_USCoreMedicationRequest_MissingMedication(t *testing.T) {
	reg := NewProfileRegistry()
	RegisterUSCoreProfiles(reg)
	v := NewProfileValidator(reg)

	resource := validUSCoreMedicationRequest()
	delete(resource, "medicationCodeableConcept")

	issues := v.ValidateAgainstProfile(resource, USCoreMedicationRequestURL)
	if !hasErrorAtPath(issues, "MedicationRequest.medication[x]") {
		t.Errorf("expected error for missing medication[x], got: %+v", issues)
	}
}

func TestProfileValidator_USCoreMedicationRequest_MissingSubject(t *testing.T) {
	reg := NewProfileRegistry()
	RegisterUSCoreProfiles(reg)
	v := NewProfileValidator(reg)

	resource := validUSCoreMedicationRequest()
	delete(resource, "subject")

	issues := v.ValidateAgainstProfile(resource, USCoreMedicationRequestURL)
	if !hasErrorAtPath(issues, "MedicationRequest.subject") {
		t.Errorf("expected error for missing subject, got: %+v", issues)
	}
}

func TestProfileValidator_USCoreMedicationRequest_MedicationReference(t *testing.T) {
	reg := NewProfileRegistry()
	RegisterUSCoreProfiles(reg)
	v := NewProfileValidator(reg)

	resource := validUSCoreMedicationRequest()
	delete(resource, "medicationCodeableConcept")
	resource["medicationReference"] = map[string]interface{}{
		"reference": "Medication/med-1",
	}

	issues := v.ValidateAgainstProfile(resource, USCoreMedicationRequestURL)
	errors := filterErrors(issues)
	for _, e := range errors {
		if strings.Contains(e.Path, "medication") {
			t.Errorf("medicationReference should satisfy medication[x], got error: %+v", e)
		}
	}
}

func TestProfileValidator_USCoreMedicationRequest_MustSupportWarnings(t *testing.T) {
	reg := NewProfileRegistry()
	RegisterUSCoreProfiles(reg)
	v := NewProfileValidator(reg)

	resource := validUSCoreMedicationRequest()
	// authoredOn, requester, dosageInstruction are MustSupport

	issues := v.ValidateAgainstProfile(resource, USCoreMedicationRequestURL)
	warnings := filterWarnings(issues)

	mustSupportPaths := []string{
		"MedicationRequest.authoredOn",
		"MedicationRequest.requester",
		"MedicationRequest.dosageInstruction",
	}
	for _, path := range mustSupportPaths {
		found := false
		for _, w := range warnings {
			if w.Path == path {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected MustSupport warning for %s, got warnings: %+v", path, warnings)
		}
	}
}

// ===========================================================================
// US Core Encounter Tests
// ===========================================================================

func validUSCoreEncounter() map[string]interface{} {
	return map[string]interface{}{
		"resourceType": "Encounter",
		"id":           "enc-1",
		"status":       "finished",
		"class": map[string]interface{}{
			"system": "http://terminology.hl7.org/CodeSystem/v3-ActCode",
			"code":   "AMB",
		},
		"type": []interface{}{
			map[string]interface{}{
				"coding": []interface{}{
					map[string]interface{}{
						"system": "http://www.ama-assn.org/go/cpt",
						"code":   "99213",
					},
				},
			},
		},
		"subject": map[string]interface{}{
			"reference": "Patient/patient-1",
		},
	}
}

func TestProfileValidator_USCoreEncounter_Valid(t *testing.T) {
	reg := NewProfileRegistry()
	RegisterUSCoreProfiles(reg)
	v := NewProfileValidator(reg)

	issues := v.ValidateAgainstProfile(validUSCoreEncounter(), USCoreEncounterURL)
	errors := filterErrors(issues)
	if len(errors) > 0 {
		t.Errorf("expected no errors, got: %+v", errors)
	}
}

func TestProfileValidator_USCoreEncounter_MissingStatus(t *testing.T) {
	reg := NewProfileRegistry()
	RegisterUSCoreProfiles(reg)
	v := NewProfileValidator(reg)

	resource := validUSCoreEncounter()
	delete(resource, "status")

	issues := v.ValidateAgainstProfile(resource, USCoreEncounterURL)
	if !hasErrorAtPath(issues, "Encounter.status") {
		t.Errorf("expected error for missing status, got: %+v", issues)
	}
}

func TestProfileValidator_USCoreEncounter_MissingClass(t *testing.T) {
	reg := NewProfileRegistry()
	RegisterUSCoreProfiles(reg)
	v := NewProfileValidator(reg)

	resource := validUSCoreEncounter()
	delete(resource, "class")

	issues := v.ValidateAgainstProfile(resource, USCoreEncounterURL)
	if !hasErrorAtPath(issues, "Encounter.class") {
		t.Errorf("expected error for missing class, got: %+v", issues)
	}
}

func TestProfileValidator_USCoreEncounter_MissingType(t *testing.T) {
	reg := NewProfileRegistry()
	RegisterUSCoreProfiles(reg)
	v := NewProfileValidator(reg)

	resource := validUSCoreEncounter()
	delete(resource, "type")

	issues := v.ValidateAgainstProfile(resource, USCoreEncounterURL)
	if !hasErrorAtPath(issues, "Encounter.type") {
		t.Errorf("expected error for missing type, got: %+v", issues)
	}
}

func TestProfileValidator_USCoreEncounter_MissingSubject(t *testing.T) {
	reg := NewProfileRegistry()
	RegisterUSCoreProfiles(reg)
	v := NewProfileValidator(reg)

	resource := validUSCoreEncounter()
	delete(resource, "subject")

	issues := v.ValidateAgainstProfile(resource, USCoreEncounterURL)
	if !hasErrorAtPath(issues, "Encounter.subject") {
		t.Errorf("expected error for missing subject, got: %+v", issues)
	}
}

func TestProfileValidator_USCoreEncounter_MustSupportPeriodWarning(t *testing.T) {
	reg := NewProfileRegistry()
	RegisterUSCoreProfiles(reg)
	v := NewProfileValidator(reg)

	resource := validUSCoreEncounter()
	// period is MustSupport, missing should generate warning

	issues := v.ValidateAgainstProfile(resource, USCoreEncounterURL)
	warnings := filterWarnings(issues)
	found := false
	for _, w := range warnings {
		if w.Path == "Encounter.period" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected MustSupport warning for period, got warnings: %+v", warnings)
	}
}

// ===========================================================================
// US Core Procedure Tests
// ===========================================================================

func validUSCoreProcedure() map[string]interface{} {
	return map[string]interface{}{
		"resourceType": "Procedure",
		"id":           "proc-1",
		"status":       "completed",
		"code": map[string]interface{}{
			"coding": []interface{}{
				map[string]interface{}{
					"system": "http://snomed.info/sct",
					"code":   "80146002",
				},
			},
		},
		"subject": map[string]interface{}{
			"reference": "Patient/patient-1",
		},
		"performedDateTime": "2024-01-15",
	}
}

func TestProfileValidator_USCoreProcedure_Valid(t *testing.T) {
	reg := NewProfileRegistry()
	RegisterUSCoreProfiles(reg)
	v := NewProfileValidator(reg)

	issues := v.ValidateAgainstProfile(validUSCoreProcedure(), USCoreProcedureURL)
	errors := filterErrors(issues)
	if len(errors) > 0 {
		t.Errorf("expected no errors, got: %+v", errors)
	}
}

func TestProfileValidator_USCoreProcedure_MissingStatus(t *testing.T) {
	reg := NewProfileRegistry()
	RegisterUSCoreProfiles(reg)
	v := NewProfileValidator(reg)

	resource := validUSCoreProcedure()
	delete(resource, "status")

	issues := v.ValidateAgainstProfile(resource, USCoreProcedureURL)
	if !hasErrorAtPath(issues, "Procedure.status") {
		t.Errorf("expected error for missing status, got: %+v", issues)
	}
}

func TestProfileValidator_USCoreProcedure_MissingCode(t *testing.T) {
	reg := NewProfileRegistry()
	RegisterUSCoreProfiles(reg)
	v := NewProfileValidator(reg)

	resource := validUSCoreProcedure()
	delete(resource, "code")

	issues := v.ValidateAgainstProfile(resource, USCoreProcedureURL)
	if !hasErrorAtPath(issues, "Procedure.code") {
		t.Errorf("expected error for missing code, got: %+v", issues)
	}
}

func TestProfileValidator_USCoreProcedure_MissingSubject(t *testing.T) {
	reg := NewProfileRegistry()
	RegisterUSCoreProfiles(reg)
	v := NewProfileValidator(reg)

	resource := validUSCoreProcedure()
	delete(resource, "subject")

	issues := v.ValidateAgainstProfile(resource, USCoreProcedureURL)
	if !hasErrorAtPath(issues, "Procedure.subject") {
		t.Errorf("expected error for missing subject, got: %+v", issues)
	}
}

func TestProfileValidator_USCoreProcedure_MustSupportPerformedWarning(t *testing.T) {
	reg := NewProfileRegistry()
	RegisterUSCoreProfiles(reg)
	v := NewProfileValidator(reg)

	resource := validUSCoreProcedure()
	delete(resource, "performedDateTime")

	issues := v.ValidateAgainstProfile(resource, USCoreProcedureURL)
	warnings := filterWarnings(issues)
	found := false
	for _, w := range warnings {
		if strings.Contains(w.Path, "performed") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected MustSupport warning for performed[x], got warnings: %+v", warnings)
	}
}

func TestProfileValidator_USCoreProcedure_AllRequiredMissing(t *testing.T) {
	reg := NewProfileRegistry()
	RegisterUSCoreProfiles(reg)
	v := NewProfileValidator(reg)

	resource := map[string]interface{}{
		"resourceType": "Procedure",
		"id":           "proc-1",
	}

	issues := v.ValidateAgainstProfile(resource, USCoreProcedureURL)
	errors := filterErrors(issues)
	if len(errors) < 3 {
		t.Errorf("expected at least 3 errors (status, code, subject), got %d: %+v", len(errors), errors)
	}
}

// ===========================================================================
// US Core Immunization Tests
// ===========================================================================

func validUSCoreImmunization() map[string]interface{} {
	return map[string]interface{}{
		"resourceType": "Immunization",
		"id":           "imm-1",
		"status":       "completed",
		"vaccineCode": map[string]interface{}{
			"coding": []interface{}{
				map[string]interface{}{
					"system": "http://hl7.org/fhir/sid/cvx",
					"code":   "158",
				},
			},
		},
		"patient": map[string]interface{}{
			"reference": "Patient/patient-1",
		},
		"occurrenceDateTime": "2024-06-15",
	}
}

func TestProfileValidator_USCoreImmunization_Valid(t *testing.T) {
	reg := NewProfileRegistry()
	RegisterUSCoreProfiles(reg)
	v := NewProfileValidator(reg)

	issues := v.ValidateAgainstProfile(validUSCoreImmunization(), USCoreImmunizationURL)
	errors := filterErrors(issues)
	if len(errors) > 0 {
		t.Errorf("expected no errors, got: %+v", errors)
	}
}

func TestProfileValidator_USCoreImmunization_MissingStatus(t *testing.T) {
	reg := NewProfileRegistry()
	RegisterUSCoreProfiles(reg)
	v := NewProfileValidator(reg)

	resource := validUSCoreImmunization()
	delete(resource, "status")

	issues := v.ValidateAgainstProfile(resource, USCoreImmunizationURL)
	if !hasErrorAtPath(issues, "Immunization.status") {
		t.Errorf("expected error for missing status, got: %+v", issues)
	}
}

func TestProfileValidator_USCoreImmunization_MissingVaccineCode(t *testing.T) {
	reg := NewProfileRegistry()
	RegisterUSCoreProfiles(reg)
	v := NewProfileValidator(reg)

	resource := validUSCoreImmunization()
	delete(resource, "vaccineCode")

	issues := v.ValidateAgainstProfile(resource, USCoreImmunizationURL)
	if !hasErrorAtPath(issues, "Immunization.vaccineCode") {
		t.Errorf("expected error for missing vaccineCode, got: %+v", issues)
	}
}

func TestProfileValidator_USCoreImmunization_MissingPatient(t *testing.T) {
	reg := NewProfileRegistry()
	RegisterUSCoreProfiles(reg)
	v := NewProfileValidator(reg)

	resource := validUSCoreImmunization()
	delete(resource, "patient")

	issues := v.ValidateAgainstProfile(resource, USCoreImmunizationURL)
	if !hasErrorAtPath(issues, "Immunization.patient") {
		t.Errorf("expected error for missing patient, got: %+v", issues)
	}
}

func TestProfileValidator_USCoreImmunization_MissingOccurrence(t *testing.T) {
	reg := NewProfileRegistry()
	RegisterUSCoreProfiles(reg)
	v := NewProfileValidator(reg)

	resource := validUSCoreImmunization()
	delete(resource, "occurrenceDateTime")

	issues := v.ValidateAgainstProfile(resource, USCoreImmunizationURL)
	if !hasErrorAtPath(issues, "Immunization.occurrence[x]") {
		t.Errorf("expected error for missing occurrence[x], got: %+v", issues)
	}
}

func TestProfileValidator_USCoreImmunization_MustSupportPrimarySource(t *testing.T) {
	reg := NewProfileRegistry()
	RegisterUSCoreProfiles(reg)
	v := NewProfileValidator(reg)

	resource := validUSCoreImmunization()
	// primarySource is MustSupport

	issues := v.ValidateAgainstProfile(resource, USCoreImmunizationURL)
	warnings := filterWarnings(issues)
	found := false
	for _, w := range warnings {
		if w.Path == "Immunization.primarySource" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected MustSupport warning for primarySource, got warnings: %+v", warnings)
	}
}

// ===========================================================================
// US Core DiagnosticReport Lab Tests
// ===========================================================================

func validUSCoreDiagnosticReport() map[string]interface{} {
	return map[string]interface{}{
		"resourceType": "DiagnosticReport",
		"id":           "dr-1",
		"status":       "final",
		"category": []interface{}{
			map[string]interface{}{
				"coding": []interface{}{
					map[string]interface{}{
						"system": "http://terminology.hl7.org/CodeSystem/v2-0074",
						"code":   "LAB",
					},
				},
			},
		},
		"code": map[string]interface{}{
			"coding": []interface{}{
				map[string]interface{}{
					"system": "http://loinc.org",
					"code":   "58410-2",
				},
			},
		},
		"subject": map[string]interface{}{
			"reference": "Patient/patient-1",
		},
	}
}

func TestProfileValidator_USCoreDiagnosticReport_Valid(t *testing.T) {
	reg := NewProfileRegistry()
	RegisterUSCoreProfiles(reg)
	v := NewProfileValidator(reg)

	issues := v.ValidateAgainstProfile(validUSCoreDiagnosticReport(), USCoreDiagnosticReportLabURL)
	errors := filterErrors(issues)
	if len(errors) > 0 {
		t.Errorf("expected no errors, got: %+v", errors)
	}
}

func TestProfileValidator_USCoreDiagnosticReport_MissingStatus(t *testing.T) {
	reg := NewProfileRegistry()
	RegisterUSCoreProfiles(reg)
	v := NewProfileValidator(reg)

	resource := validUSCoreDiagnosticReport()
	delete(resource, "status")

	issues := v.ValidateAgainstProfile(resource, USCoreDiagnosticReportLabURL)
	if !hasErrorAtPath(issues, "DiagnosticReport.status") {
		t.Errorf("expected error for missing status, got: %+v", issues)
	}
}

func TestProfileValidator_USCoreDiagnosticReport_MissingCategory(t *testing.T) {
	reg := NewProfileRegistry()
	RegisterUSCoreProfiles(reg)
	v := NewProfileValidator(reg)

	resource := validUSCoreDiagnosticReport()
	delete(resource, "category")

	issues := v.ValidateAgainstProfile(resource, USCoreDiagnosticReportLabURL)
	if !hasErrorAtPath(issues, "DiagnosticReport.category") {
		t.Errorf("expected error for missing category, got: %+v", issues)
	}
}

func TestProfileValidator_USCoreDiagnosticReport_CategoryWithoutLAB(t *testing.T) {
	reg := NewProfileRegistry()
	RegisterUSCoreProfiles(reg)
	v := NewProfileValidator(reg)

	resource := validUSCoreDiagnosticReport()
	resource["category"] = []interface{}{
		map[string]interface{}{
			"coding": []interface{}{
				map[string]interface{}{
					"system": "http://terminology.hl7.org/CodeSystem/v2-0074",
					"code":   "RAD", // not LAB
				},
			},
		},
	}

	issues := v.ValidateAgainstProfile(resource, USCoreDiagnosticReportLabURL)
	found := false
	for _, issue := range issues {
		if issue.Severity == "error" && strings.Contains(issue.Path, "category") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected error for category without LAB, got: %+v", issues)
	}
}

func TestProfileValidator_USCoreDiagnosticReport_MissingCode(t *testing.T) {
	reg := NewProfileRegistry()
	RegisterUSCoreProfiles(reg)
	v := NewProfileValidator(reg)

	resource := validUSCoreDiagnosticReport()
	delete(resource, "code")

	issues := v.ValidateAgainstProfile(resource, USCoreDiagnosticReportLabURL)
	if !hasErrorAtPath(issues, "DiagnosticReport.code") {
		t.Errorf("expected error for missing code, got: %+v", issues)
	}
}

func TestProfileValidator_USCoreDiagnosticReport_MissingSubject(t *testing.T) {
	reg := NewProfileRegistry()
	RegisterUSCoreProfiles(reg)
	v := NewProfileValidator(reg)

	resource := validUSCoreDiagnosticReport()
	delete(resource, "subject")

	issues := v.ValidateAgainstProfile(resource, USCoreDiagnosticReportLabURL)
	if !hasErrorAtPath(issues, "DiagnosticReport.subject") {
		t.Errorf("expected error for missing subject, got: %+v", issues)
	}
}

func TestProfileValidator_USCoreDiagnosticReport_MustSupportWarnings(t *testing.T) {
	reg := NewProfileRegistry()
	RegisterUSCoreProfiles(reg)
	v := NewProfileValidator(reg)

	resource := validUSCoreDiagnosticReport()
	// effective[x] and result are MustSupport

	issues := v.ValidateAgainstProfile(resource, USCoreDiagnosticReportLabURL)
	warnings := filterWarnings(issues)

	msFields := []string{"effective", "result"}
	for _, field := range msFields {
		found := false
		for _, w := range warnings {
			if strings.Contains(w.Path, field) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected MustSupport warning for %s, got warnings: %+v", field, warnings)
		}
	}
}

// ===========================================================================
// US Core DocumentReference Tests
// ===========================================================================

func validUSCoreDocumentReference() map[string]interface{} {
	return map[string]interface{}{
		"resourceType": "DocumentReference",
		"id":           "docref-1",
		"status":       "current",
		"type": map[string]interface{}{
			"coding": []interface{}{
				map[string]interface{}{
					"system": "http://loinc.org",
					"code":   "34133-9",
				},
			},
		},
		"category": []interface{}{
			map[string]interface{}{
				"coding": []interface{}{
					map[string]interface{}{
						"system": "http://hl7.org/fhir/us/core/CodeSystem/us-core-documentreference-category",
						"code":   "clinical-note",
					},
				},
			},
		},
		"subject": map[string]interface{}{
			"reference": "Patient/patient-1",
		},
		"content": []interface{}{
			map[string]interface{}{
				"attachment": map[string]interface{}{
					"contentType": "application/pdf",
					"url":         "http://example.com/doc.pdf",
				},
			},
		},
	}
}

func TestProfileValidator_USCoreDocumentReference_Valid(t *testing.T) {
	reg := NewProfileRegistry()
	RegisterUSCoreProfiles(reg)
	v := NewProfileValidator(reg)

	issues := v.ValidateAgainstProfile(validUSCoreDocumentReference(), USCoreDocumentReferenceURL)
	errors := filterErrors(issues)
	if len(errors) > 0 {
		t.Errorf("expected no errors, got: %+v", errors)
	}
}

func TestProfileValidator_USCoreDocumentReference_MissingStatus(t *testing.T) {
	reg := NewProfileRegistry()
	RegisterUSCoreProfiles(reg)
	v := NewProfileValidator(reg)

	resource := validUSCoreDocumentReference()
	delete(resource, "status")

	issues := v.ValidateAgainstProfile(resource, USCoreDocumentReferenceURL)
	if !hasErrorAtPath(issues, "DocumentReference.status") {
		t.Errorf("expected error for missing status, got: %+v", issues)
	}
}

func TestProfileValidator_USCoreDocumentReference_MissingType(t *testing.T) {
	reg := NewProfileRegistry()
	RegisterUSCoreProfiles(reg)
	v := NewProfileValidator(reg)

	resource := validUSCoreDocumentReference()
	delete(resource, "type")

	issues := v.ValidateAgainstProfile(resource, USCoreDocumentReferenceURL)
	if !hasErrorAtPath(issues, "DocumentReference.type") {
		t.Errorf("expected error for missing type, got: %+v", issues)
	}
}

func TestProfileValidator_USCoreDocumentReference_MissingCategory(t *testing.T) {
	reg := NewProfileRegistry()
	RegisterUSCoreProfiles(reg)
	v := NewProfileValidator(reg)

	resource := validUSCoreDocumentReference()
	delete(resource, "category")

	issues := v.ValidateAgainstProfile(resource, USCoreDocumentReferenceURL)
	if !hasErrorAtPath(issues, "DocumentReference.category") {
		t.Errorf("expected error for missing category, got: %+v", issues)
	}
}

func TestProfileValidator_USCoreDocumentReference_MissingSubject(t *testing.T) {
	reg := NewProfileRegistry()
	RegisterUSCoreProfiles(reg)
	v := NewProfileValidator(reg)

	resource := validUSCoreDocumentReference()
	delete(resource, "subject")

	issues := v.ValidateAgainstProfile(resource, USCoreDocumentReferenceURL)
	if !hasErrorAtPath(issues, "DocumentReference.subject") {
		t.Errorf("expected error for missing subject, got: %+v", issues)
	}
}

func TestProfileValidator_USCoreDocumentReference_MissingContent(t *testing.T) {
	reg := NewProfileRegistry()
	RegisterUSCoreProfiles(reg)
	v := NewProfileValidator(reg)

	resource := validUSCoreDocumentReference()
	delete(resource, "content")

	issues := v.ValidateAgainstProfile(resource, USCoreDocumentReferenceURL)
	if !hasErrorAtPath(issues, "DocumentReference.content") {
		t.Errorf("expected error for missing content, got: %+v", issues)
	}
}

func TestProfileValidator_USCoreDocumentReference_MissingContentAttachment(t *testing.T) {
	reg := NewProfileRegistry()
	RegisterUSCoreProfiles(reg)
	v := NewProfileValidator(reg)

	resource := validUSCoreDocumentReference()
	resource["content"] = []interface{}{
		map[string]interface{}{
			// attachment is missing
		},
	}

	issues := v.ValidateAgainstProfile(resource, USCoreDocumentReferenceURL)
	if !hasErrorAtPath(issues, "DocumentReference.content.attachment") {
		t.Errorf("expected error for missing content.attachment, got: %+v", issues)
	}
}

func TestProfileValidator_USCoreDocumentReference_MissingContentType(t *testing.T) {
	reg := NewProfileRegistry()
	RegisterUSCoreProfiles(reg)
	v := NewProfileValidator(reg)

	resource := validUSCoreDocumentReference()
	resource["content"] = []interface{}{
		map[string]interface{}{
			"attachment": map[string]interface{}{
				"url": "http://example.com/doc.pdf",
				// contentType is missing
			},
		},
	}

	issues := v.ValidateAgainstProfile(resource, USCoreDocumentReferenceURL)
	if !hasErrorAtPath(issues, "DocumentReference.content.attachment.contentType") {
		t.Errorf("expected error for missing content.attachment.contentType, got: %+v", issues)
	}
}

// ===========================================================================
// US Core AllergyIntolerance Tests
// ===========================================================================

func validUSCoreAllergyIntolerance() map[string]interface{} {
	return map[string]interface{}{
		"resourceType": "AllergyIntolerance",
		"id":           "ai-1",
		"clinicalStatus": map[string]interface{}{
			"coding": []interface{}{
				map[string]interface{}{
					"system": "http://terminology.hl7.org/CodeSystem/allergyintolerance-clinical",
					"code":   "active",
				},
			},
		},
		"verificationStatus": map[string]interface{}{
			"coding": []interface{}{
				map[string]interface{}{
					"system": "http://terminology.hl7.org/CodeSystem/allergyintolerance-verification",
					"code":   "confirmed",
				},
			},
		},
		"code": map[string]interface{}{
			"coding": []interface{}{
				map[string]interface{}{
					"system": "http://snomed.info/sct",
					"code":   "762952008",
				},
			},
		},
		"patient": map[string]interface{}{
			"reference": "Patient/patient-1",
		},
	}
}

func TestProfileValidator_USCoreAllergyIntolerance_Valid(t *testing.T) {
	reg := NewProfileRegistry()
	RegisterUSCoreProfiles(reg)
	v := NewProfileValidator(reg)

	issues := v.ValidateAgainstProfile(validUSCoreAllergyIntolerance(), USCoreAllergyIntoleranceURL)
	errors := filterErrors(issues)
	if len(errors) > 0 {
		t.Errorf("expected no errors, got: %+v", errors)
	}
}

func TestProfileValidator_USCoreAllergyIntolerance_MissingCode(t *testing.T) {
	reg := NewProfileRegistry()
	RegisterUSCoreProfiles(reg)
	v := NewProfileValidator(reg)

	resource := validUSCoreAllergyIntolerance()
	delete(resource, "code")

	issues := v.ValidateAgainstProfile(resource, USCoreAllergyIntoleranceURL)
	if !hasErrorAtPath(issues, "AllergyIntolerance.code") {
		t.Errorf("expected error for missing code, got: %+v", issues)
	}
}

func TestProfileValidator_USCoreAllergyIntolerance_MissingPatient(t *testing.T) {
	reg := NewProfileRegistry()
	RegisterUSCoreProfiles(reg)
	v := NewProfileValidator(reg)

	resource := validUSCoreAllergyIntolerance()
	delete(resource, "patient")

	issues := v.ValidateAgainstProfile(resource, USCoreAllergyIntoleranceURL)
	if !hasErrorAtPath(issues, "AllergyIntolerance.patient") {
		t.Errorf("expected error for missing patient, got: %+v", issues)
	}
}

func TestProfileValidator_USCoreAllergyIntolerance_MustSupportClinicalStatus(t *testing.T) {
	reg := NewProfileRegistry()
	RegisterUSCoreProfiles(reg)
	v := NewProfileValidator(reg)

	resource := validUSCoreAllergyIntolerance()
	delete(resource, "clinicalStatus")

	issues := v.ValidateAgainstProfile(resource, USCoreAllergyIntoleranceURL)
	warnings := filterWarnings(issues)
	found := false
	for _, w := range warnings {
		if w.Path == "AllergyIntolerance.clinicalStatus" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected MustSupport warning for clinicalStatus, got warnings: %+v", warnings)
	}
}

func TestProfileValidator_USCoreAllergyIntolerance_MustSupportVerificationStatus(t *testing.T) {
	reg := NewProfileRegistry()
	RegisterUSCoreProfiles(reg)
	v := NewProfileValidator(reg)

	resource := validUSCoreAllergyIntolerance()
	delete(resource, "verificationStatus")

	issues := v.ValidateAgainstProfile(resource, USCoreAllergyIntoleranceURL)
	warnings := filterWarnings(issues)
	found := false
	for _, w := range warnings {
		if w.Path == "AllergyIntolerance.verificationStatus" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected MustSupport warning for verificationStatus, got warnings: %+v", warnings)
	}
}

// ===========================================================================
// ValidateResource (auto-detect profiles by type) Tests
// ===========================================================================

func TestProfileValidator_ValidateResource_AutoDetect(t *testing.T) {
	reg := NewProfileRegistry()
	RegisterUSCoreProfiles(reg)
	v := NewProfileValidator(reg)

	resource := validUSCorePatient()
	issues := v.ValidateResource(resource)

	// Should auto-detect Patient profiles and validate
	// A valid patient should have no errors from profile validation
	errors := filterErrors(issues)
	if len(errors) > 0 {
		t.Errorf("expected no errors for valid auto-detected Patient, got: %+v", errors)
	}
}

func TestProfileValidator_ValidateResource_InvalidAutoDetect(t *testing.T) {
	reg := NewProfileRegistry()
	RegisterUSCoreProfiles(reg)
	v := NewProfileValidator(reg)

	resource := map[string]interface{}{
		"resourceType": "Patient",
		"id":           "p-1",
		// Missing identifier, name, gender
	}

	issues := v.ValidateResource(resource)
	errors := filterErrors(issues)
	if len(errors) == 0 {
		t.Error("expected errors for Patient missing required US Core fields")
	}
}

func TestProfileValidator_ValidateResource_NoProfilesForType(t *testing.T) {
	reg := NewProfileRegistry()
	// Don't register any profiles
	v := NewProfileValidator(reg)

	resource := map[string]interface{}{
		"resourceType": "Patient",
		"id":           "p-1",
	}

	issues := v.ValidateResource(resource)
	if len(issues) != 0 {
		t.Errorf("expected no issues when no profiles registered, got: %+v", issues)
	}
}

func TestProfileValidator_ValidateAgainstProfile_UnknownProfile(t *testing.T) {
	reg := NewProfileRegistry()
	v := NewProfileValidator(reg)

	resource := validUSCorePatient()
	issues := v.ValidateAgainstProfile(resource, "http://unknown/profile")
	if len(issues) != 1 {
		t.Fatalf("expected 1 issue for unknown profile, got: %+v", issues)
	}
	if issues[0].Severity != "error" {
		t.Errorf("expected error severity, got %s", issues[0].Severity)
	}
	if issues[0].Code != "not-found" {
		t.Errorf("expected not-found code, got %s", issues[0].Code)
	}
}

// ===========================================================================
// Edge Case Tests
// ===========================================================================

func TestProfileValidator_NonFHIRObject(t *testing.T) {
	reg := NewProfileRegistry()
	RegisterUSCoreProfiles(reg)
	v := NewProfileValidator(reg)

	// No resourceType at all
	resource := map[string]interface{}{
		"name": "not a FHIR resource",
	}

	issues := v.ValidateResource(resource)
	// Should return empty (no resourceType means no profiles match)
	if len(issues) != 0 {
		t.Errorf("expected no issues for non-FHIR object, got: %+v", issues)
	}
}

func TestProfileValidator_MissingResourceType(t *testing.T) {
	reg := NewProfileRegistry()
	RegisterUSCoreProfiles(reg)
	v := NewProfileValidator(reg)

	resource := map[string]interface{}{
		"id": "123",
	}

	issues := v.ValidateAgainstProfile(resource, USCorePatientURL)
	found := false
	for _, issue := range issues {
		if issue.Severity == "error" && strings.Contains(issue.Description, "resourceType") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected error about missing resourceType, got: %+v", issues)
	}
}

func TestProfileValidator_NullValuesInRequired(t *testing.T) {
	reg := NewProfileRegistry()
	RegisterUSCoreProfiles(reg)
	v := NewProfileValidator(reg)

	resource := map[string]interface{}{
		"resourceType": "Patient",
		"id":           "p-1",
		"identifier":   nil,
		"name":         nil,
		"gender":       nil,
	}

	issues := v.ValidateAgainstProfile(resource, USCorePatientURL)
	errors := filterErrors(issues)
	if len(errors) < 3 {
		t.Errorf("expected at least 3 errors for nil required fields, got %d: %+v", len(errors), errors)
	}
}

func TestProfileValidator_EmptyArraysForRequired(t *testing.T) {
	reg := NewProfileRegistry()
	RegisterUSCoreProfiles(reg)
	v := NewProfileValidator(reg)

	resource := map[string]interface{}{
		"resourceType": "Patient",
		"id":           "p-1",
		"identifier":   []interface{}{},
		"name":         []interface{}{},
		"gender":       "male",
	}

	issues := v.ValidateAgainstProfile(resource, USCorePatientURL)
	if !hasErrorAtPath(issues, "Patient.identifier") {
		t.Errorf("expected error for empty identifier array, got: %+v", issues)
	}
	if !hasErrorAtPath(issues, "Patient.name") {
		t.Errorf("expected error for empty name array, got: %+v", issues)
	}
}

func TestProfileValidator_ValidateAgainstMultipleProfiles(t *testing.T) {
	reg := NewProfileRegistry()
	RegisterUSCoreProfiles(reg)
	// Register a second custom Patient profile
	reg.Register(ProfileDefinition{
		URL:     "http://custom.org/StructureDefinition/custom-patient",
		Name:    "CustomPatient",
		Type:    "Patient",
		Version: "1.0.0",
		Status:  "active",
		Constraints: []ProfileConstraint{
			{Path: "Patient.birthDate", Min: 1, Max: "*"},
		},
	})
	v := NewProfileValidator(reg)

	// Patient without birthDate — should fail custom profile
	resource := validUSCorePatient()

	issues := v.ValidateResource(resource)
	// Should have an error from the custom profile for missing birthDate
	found := false
	for _, issue := range issues {
		if issue.Severity == "error" && strings.Contains(issue.Path, "birthDate") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected error from custom profile for missing birthDate, got: %+v", issues)
	}
}

func TestProfileValidator_ResourceTypeMismatch(t *testing.T) {
	reg := NewProfileRegistry()
	RegisterUSCoreProfiles(reg)
	v := NewProfileValidator(reg)

	// Try to validate Observation against Patient profile
	resource := validUSCoreObservationLab()
	issues := v.ValidateAgainstProfile(resource, USCorePatientURL)
	found := false
	for _, issue := range issues {
		if issue.Severity == "error" && strings.Contains(issue.Description, "type mismatch") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected error for resource type mismatch, got: %+v", issues)
	}
}

func TestProfileValidator_NilResource(t *testing.T) {
	reg := NewProfileRegistry()
	RegisterUSCoreProfiles(reg)
	v := NewProfileValidator(reg)

	issues := v.ValidateAgainstProfile(nil, USCorePatientURL)
	if len(issues) != 1 || issues[0].Severity != "error" {
		t.Errorf("expected single error for nil resource, got: %+v", issues)
	}
}

// ===========================================================================
// Handler Tests
// ===========================================================================

func newProfileTestSetup() (*ProfileHandler, *echo.Echo) {
	reg := NewProfileRegistry()
	RegisterUSCoreProfiles(reg)
	v := NewProfileValidator(reg)
	h := NewProfileHandler(v, reg)
	e := echo.New()
	return h, e
}

func TestProfileHandler_ListProfiles(t *testing.T) {
	h, e := newProfileTestSetup()

	req := httptest.NewRequest(http.MethodGet, "/fhir/StructureDefinition", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := h.ListProfiles(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if result["resourceType"] != "Bundle" {
		t.Errorf("expected Bundle resourceType, got %v", result["resourceType"])
	}

	entries, ok := result["entry"].([]interface{})
	if !ok {
		t.Fatal("expected entry array")
	}
	if len(entries) < 10 {
		t.Errorf("expected at least 10 profiles, got %d", len(entries))
	}
}

func TestProfileHandler_GetProfile(t *testing.T) {
	h, e := newProfileTestSetup()

	req := httptest.NewRequest(http.MethodGet, "/fhir/StructureDefinition/us-core-patient", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("us-core-patient")

	if err := h.GetProfile(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if result["resourceType"] != "StructureDefinition" {
		t.Errorf("expected StructureDefinition resourceType, got %v", result["resourceType"])
	}
	if result["name"] != "USCorePatient" {
		t.Errorf("expected name USCorePatient, got %v", result["name"])
	}
}

func TestProfileHandler_GetProfile_ByCanonicalURL(t *testing.T) {
	h, e := newProfileTestSetup()

	req := httptest.NewRequest(http.MethodGet, "/fhir/StructureDefinition/http%3A%2F%2Fhl7.org%2Ffhir%2Fus%2Fcore%2FStructureDefinition%2Fus-core-patient", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("http://hl7.org/fhir/us/core/StructureDefinition/us-core-patient")

	if err := h.GetProfile(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestProfileHandler_GetProfile_NotFound(t *testing.T) {
	h, e := newProfileTestSetup()

	req := httptest.NewRequest(http.MethodGet, "/fhir/StructureDefinition/nonexistent", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("nonexistent")

	if err := h.GetProfile(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestProfileHandler_ValidateWithProfile(t *testing.T) {
	h, e := newProfileTestSetup()

	body, _ := json.Marshal(validUSCorePatient())
	req := httptest.NewRequest(http.MethodPost, "/fhir/$validate?profile="+USCorePatientURL, strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := h.ValidateWithProfile(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if result["resourceType"] != "OperationOutcome" {
		t.Errorf("expected OperationOutcome, got %v", result["resourceType"])
	}

	issues := result["issue"].([]interface{})
	for _, issueRaw := range issues {
		issue := issueRaw.(map[string]interface{})
		if issue["severity"] == "error" || issue["severity"] == "fatal" {
			t.Errorf("expected no errors for valid Patient, got: %v", issue)
		}
	}
}

func TestProfileHandler_ValidateWithoutProfile_AutoDetect(t *testing.T) {
	h, e := newProfileTestSetup()

	body, _ := json.Marshal(validUSCorePatient())
	req := httptest.NewRequest(http.MethodPost, "/fhir/$validate", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := h.ValidateWithProfile(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestProfileHandler_ValidateInvalidResource(t *testing.T) {
	h, e := newProfileTestSetup()

	invalid := map[string]interface{}{
		"resourceType": "Patient",
		"id":           "p-1",
		// Missing required fields
	}
	body, _ := json.Marshal(invalid)
	req := httptest.NewRequest(http.MethodPost, "/fhir/$validate?profile="+USCorePatientURL, strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := h.ValidateWithProfile(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 (OperationOutcome with errors), got %d", rec.Code)
	}

	var result map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &result)
	issues := result["issue"].([]interface{})

	hasError := false
	for _, issueRaw := range issues {
		issue := issueRaw.(map[string]interface{})
		if issue["severity"] == "error" {
			hasError = true
			break
		}
	}
	if !hasError {
		t.Error("expected at least one error for invalid resource")
	}
}

func TestProfileHandler_ValidateValidResource_NoErrors(t *testing.T) {
	h, e := newProfileTestSetup()

	body, _ := json.Marshal(validUSCoreCondition())
	req := httptest.NewRequest(http.MethodPost, "/fhir/$validate?profile="+USCoreConditionURL, strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := h.ValidateWithProfile(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &result)
	issues := result["issue"].([]interface{})

	for _, issueRaw := range issues {
		issue := issueRaw.(map[string]interface{})
		if issue["severity"] == "error" || issue["severity"] == "fatal" {
			t.Errorf("expected no errors for valid Condition, got: %v", issue)
		}
	}
}

func TestProfileHandler_RegisterCustomProfile(t *testing.T) {
	h, e := newProfileTestSetup()

	custom := map[string]interface{}{
		"url":     "http://custom.org/SD/my-patient",
		"name":    "MyPatient",
		"type":    "Patient",
		"version": "1.0.0",
		"status":  "active",
		"constraints": []interface{}{
			map[string]interface{}{
				"path":        "Patient.birthDate",
				"min":         float64(1),
				"max":         "*",
				"mustSupport": true,
			},
		},
	}

	body, _ := json.Marshal(custom)
	req := httptest.NewRequest(http.MethodPost, "/fhir/metadata/profiles", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := h.RegisterCustomProfile(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}

	// Verify the profile was registered
	_, ok := h.registry.GetByURL("http://custom.org/SD/my-patient")
	if !ok {
		t.Error("expected custom profile to be registered")
	}
}

func TestProfileHandler_ListProfilesByType(t *testing.T) {
	h, e := newProfileTestSetup()

	req := httptest.NewRequest(http.MethodGet, "/fhir/metadata/profiles?type=Patient", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := h.ListProfilesByType(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var result []map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if len(result) == 0 {
		t.Error("expected at least one Patient profile")
	}

	for _, p := range result {
		if p["type"] != "Patient" {
			t.Errorf("expected all profiles to be Patient type, got %v", p["type"])
		}
	}
}

func TestProfileHandler_RegisterRoutes(t *testing.T) {
	h, e := newProfileTestSetup()
	g := e.Group("/fhir")
	h.RegisterRoutes(g)

	routes := e.Routes()
	routePaths := make(map[string]bool)
	for _, r := range routes {
		routePaths[r.Method+":"+r.Path] = true
	}

	expected := []string{
		"GET:/fhir/StructureDefinition",
		"GET:/fhir/StructureDefinition/:id",
		"GET:/fhir/metadata/profiles",
		"POST:/fhir/metadata/profiles",
	}
	for _, path := range expected {
		if !routePaths[path] {
			t.Errorf("missing expected route: %s", path)
		}
	}
}

func TestProfileHandler_ValidateEmptyBody(t *testing.T) {
	h, e := newProfileTestSetup()

	req := httptest.NewRequest(http.MethodPost, "/fhir/$validate?profile="+USCorePatientURL, strings.NewReader(""))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := h.ValidateWithProfile(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for empty body, got %d", rec.Code)
	}
}

func TestProfileHandler_ValidateInvalidJSON(t *testing.T) {
	h, e := newProfileTestSetup()

	req := httptest.NewRequest(http.MethodPost, "/fhir/$validate?profile="+USCorePatientURL, strings.NewReader("{not valid"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := h.ValidateWithProfile(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid JSON, got %d", rec.Code)
	}
}

// ===========================================================================
// Integration: $validate with profile parameter
// ===========================================================================

func TestProfileValidation_IntegrationWithExistingValidate(t *testing.T) {
	reg := NewProfileRegistry()
	RegisterUSCoreProfiles(reg)
	pv := NewProfileValidator(reg)

	// Simulate what the existing $validate handler does: run base validation
	// then run profile validation when profile param is set
	rv := NewResourceValidator()

	resource := map[string]interface{}{
		"resourceType": "Patient",
		"id":           "p-1",
		"name": []interface{}{
			map[string]interface{}{"family": "Smith"},
		},
		// Missing identifier and gender for US Core
	}

	// Base validation should pass
	baseResult := rv.Validate(resource)
	if !baseResult.Valid {
		t.Errorf("expected base validation to pass, got: %+v", baseResult.Issues)
	}

	// Profile validation should find issues
	profileIssues := pv.ValidateAgainstProfile(resource, USCorePatientURL)
	errors := filterErrors(profileIssues)
	if len(errors) < 2 {
		t.Errorf("expected at least 2 profile errors (identifier, gender), got %d: %+v", len(errors), errors)
	}
}

// ===========================================================================
// Test Helpers
// ===========================================================================

func filterErrors(issues []ProfileValidationIssue) []ProfileValidationIssue {
	var result []ProfileValidationIssue
	for _, i := range issues {
		if i.Severity == "error" {
			result = append(result, i)
		}
	}
	return result
}

func filterWarnings(issues []ProfileValidationIssue) []ProfileValidationIssue {
	var result []ProfileValidationIssue
	for _, i := range issues {
		if i.Severity == "warning" {
			result = append(result, i)
		}
	}
	return result
}

func hasErrorAtPath(issues []ProfileValidationIssue, path string) bool {
	for _, issue := range issues {
		if issue.Severity == "error" && issue.Path == path {
			return true
		}
	}
	return false
}
