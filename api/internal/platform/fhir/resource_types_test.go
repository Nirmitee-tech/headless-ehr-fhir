package fhir

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestFHIRResourceTypes_ContainsMajorTypes(t *testing.T) {
	major := []string{
		"Patient", "Practitioner", "PractitionerRole", "Organization",
		"Location", "Encounter", "Condition", "Observation",
		"DiagnosticReport", "Procedure", "AllergyIntolerance",
		"Immunization", "MedicationRequest", "MedicationAdministration",
		"MedicationDispense", "MedicationStatement", "Medication",
		"ServiceRequest", "CarePlan", "CareTeam", "Goal",
		"NutritionOrder", "DocumentReference", "Composition",
		"Consent", "Coverage", "Claim", "ClaimResponse",
		"ExplanationOfBenefit", "Communication", "CommunicationRequest",
		"Questionnaire", "QuestionnaireResponse", "Task",
		"Appointment", "AppointmentResponse", "Schedule", "Slot",
		"Device", "DeviceRequest", "DeviceUseStatement",
		"ImagingStudy", "Specimen", "FamilyMemberHistory",
		"RelatedPerson", "Provenance", "AuditEvent",
		"Subscription", "MessageHeader", "OperationOutcome",
		"Bundle", "Binary", "List", "Flag",
		"DetectedIssue", "ClinicalImpression", "RiskAssessment",
		"AdverseEvent", "EpisodeOfCare", "HealthcareService",
		"Endpoint", "Account", "ChargeItem", "ChargeItemDefinition",
		"Contract", "InsurancePlan", "Invoice",
		"PaymentNotice", "PaymentReconciliation",
		"EnrollmentRequest", "EnrollmentResponse",
		"ActivityDefinition", "PlanDefinition", "RequestGroup",
		"GuidanceResponse", "Measure", "MeasureReport",
		"ResearchStudy", "ResearchSubject", "Group",
		"SupplyRequest", "SupplyDelivery",
		"NamingSystem", "OperationDefinition", "MessageDefinition",
		"StructureDefinition", "StructureMap",
		"ValueSet", "CodeSystem", "ConceptMap",
		"TerminologyCapabilities", "CapabilityStatement",
		"SearchParameter", "ImplementationGuide",
		"TestScript", "TestReport", "Linkage", "Basic",
		"Media", "Substance", "SubstanceSpecification",
		"MedicationKnowledge", "MedicinalProduct",
		"MedicinalProductIngredient", "MedicinalProductAuthorization",
		"MedicinalProductContraindication", "MedicinalProductIndication",
		"MedicinalProductInteraction", "MedicinalProductPharmaceutical",
		"MedicinalProductUndesirableEffect",
		"ObservationDefinition", "SpecimenDefinition",
		"Parameters", "VerificationResult", "VisionPrescription",
		"BiologicallyDerivedProduct", "BodyStructure",
		"CatalogEntry", "DeviceDefinition", "DeviceMetric",
		"EffectEvidenceSynthesis", "Evidence", "EvidenceVariable",
		"ImmunizationEvaluation", "ImmunizationRecommendation",
		"Library", "MolecularSequence", "OrganizationAffiliation",
		"Person", "ResearchDefinition", "ResearchElementDefinition",
		"RiskEvidenceSynthesis", "SubstanceNucleicAcid",
		"SubstancePolymer", "SubstanceProtein",
		"SubstanceReferenceInformation", "SubstanceSourceMaterial",
		"EventDefinition",
	}

	for _, rt := range major {
		if !FHIRResourceTypes[rt] {
			t.Errorf("expected FHIRResourceTypes to contain %q", rt)
		}
	}
}

func TestFHIRResourceTypes_Count(t *testing.T) {
	// FHIR R4 defines approximately 150 resource types. We expect at least 145.
	if count := len(FHIRResourceTypes); count < 145 {
		t.Errorf("expected at least 145 FHIR R4 resource types, got %d", count)
	}
}

func TestIsValidResourceType_Valid(t *testing.T) {
	valid := []string{
		"Patient",
		"Observation",
		"Bundle",
		"OperationOutcome",
		"CapabilityStatement",
		"MedicinalProduct",
		"StructureDefinition",
	}
	for _, rt := range valid {
		if !IsValidResourceType(rt) {
			t.Errorf("expected %q to be a valid resource type", rt)
		}
	}
}

func TestIsValidResourceType_Invalid(t *testing.T) {
	invalid := []string{
		"FakeResource",
		"patient",
		"PATIENT",
		"",
		"Patientx",
		"unknown",
	}
	for _, rt := range invalid {
		if IsValidResourceType(rt) {
			t.Errorf("expected %q to be an invalid resource type", rt)
		}
	}
}

func TestValidateResourceType_ValidType(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/Patient", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := ValidateResourceType(c, "Patient")
	if err != nil {
		t.Fatalf("expected no error for valid resource type, got %v", err)
	}
	// No response should have been written for a valid type.
	if rec.Code == http.StatusNotFound {
		t.Error("expected no 404 response for valid resource type")
	}
}

func TestValidateResourceType_InvalidType(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/FakeResource", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := ValidateResourceType(c, "FakeResource")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", rec.Code)
	}

	var outcome OperationOutcome
	if err := json.Unmarshal(rec.Body.Bytes(), &outcome); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if outcome.ResourceType != "OperationOutcome" {
		t.Errorf("expected resourceType OperationOutcome, got %s", outcome.ResourceType)
	}
	if len(outcome.Issue) != 1 {
		t.Fatalf("expected 1 issue, got %d", len(outcome.Issue))
	}
	if outcome.Issue[0].Severity != IssueSeverityError {
		t.Errorf("expected severity error, got %s", outcome.Issue[0].Severity)
	}
	if outcome.Issue[0].Code != IssueTypeNotFound {
		t.Errorf("expected code not-found, got %s", outcome.Issue[0].Code)
	}
}

func TestValidateResourceType_EmptyString(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := ValidateResourceType(c, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status 404 for empty string, got %d", rec.Code)
	}
}

func TestValidateResourceType_CaseSensitive(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/patient", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := ValidateResourceType(c, "patient")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status 404 for lowercase 'patient', got %d", rec.Code)
	}
}

func TestValidateResourceType_DiagnosticsMessage(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/Fake", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	_ = ValidateResourceType(c, "Fake")

	var outcome OperationOutcome
	if err := json.Unmarshal(rec.Body.Bytes(), &outcome); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	expected := "resource type 'Fake' is not a known FHIR R4 resource type"
	if outcome.Issue[0].Diagnostics != expected {
		t.Errorf("expected diagnostics %q, got %q", expected, outcome.Issue[0].Diagnostics)
	}
}

func TestFHIRResourceTypes_NoDuplicateKeys(t *testing.T) {
	// Go maps inherently prevent duplicate keys at the language level.
	// This test verifies the map is well-formed and non-empty.
	if len(FHIRResourceTypes) == 0 {
		t.Fatal("FHIRResourceTypes should not be empty")
	}
	// Verify every entry is true (no false values accidentally set).
	for rt, val := range FHIRResourceTypes {
		if !val {
			t.Errorf("FHIRResourceTypes[%q] is false; all values should be true", rt)
		}
	}
}
