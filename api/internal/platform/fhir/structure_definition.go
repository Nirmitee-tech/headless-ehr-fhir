package fhir

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"

	"github.com/labstack/echo/v4"
)

// ============================================================================
// StructureDefinition Models
// ============================================================================

// StructureDefinitionResource represents a FHIR R4 StructureDefinition resource.
type StructureDefinitionResource struct {
	ResourceType   string                 `json:"resourceType"`
	ID             string                 `json:"id,omitempty"`
	URL            string                 `json:"url"`
	Name           string                 `json:"name"`
	Title          string                 `json:"title,omitempty"`
	Status         string                 `json:"status"`          // draft, active, retired
	Kind           string                 `json:"kind"`            // primitive-type, complex-type, resource, logical
	Abstract       bool                   `json:"abstract"`
	Type           string                 `json:"type"`            // e.g., "Patient", "Observation"
	BaseDefinition string                 `json:"baseDefinition,omitempty"`
	Derivation     string                 `json:"derivation,omitempty"` // specialization, constraint
	Description    string                 `json:"description,omitempty"`
	FHIRVersion    string                 `json:"fhirVersion,omitempty"`
	Snapshot       *StructureSnapshot     `json:"snapshot,omitempty"`
	Differential   *StructureDifferential `json:"differential,omitempty"`
}

// StructureSnapshot contains the full set of element definitions for the structure.
type StructureSnapshot struct {
	Element []ElementDefinition `json:"element"`
}

// StructureDifferential contains the delta of element definitions relative to the base.
type StructureDifferential struct {
	Element []ElementDefinition `json:"element"`
}

// ElementDefinition describes a single element within a StructureDefinition.
type ElementDefinition struct {
	ID          string          `json:"id,omitempty"`
	Path        string          `json:"path"`
	Short       string          `json:"short,omitempty"`
	Definition  string          `json:"definition,omitempty"`
	Min         *int            `json:"min,omitempty"`
	Max         string          `json:"max,omitempty"`
	Type        []ElementType   `json:"type,omitempty"`
	Binding     *ElementBinding `json:"binding,omitempty"`
	MustSupport bool            `json:"mustSupport,omitempty"`
}

// ElementType describes a datatype for an element.
type ElementType struct {
	Code          string   `json:"code"`
	TargetProfile []string `json:"targetProfile,omitempty"`
}

// ElementBinding describes a terminology binding for an element.
type ElementBinding struct {
	Strength string `json:"strength"` // required, extensible, preferred, example
	ValueSet string `json:"valueSet,omitempty"`
}

// ============================================================================
// StructureDefinition Store
// ============================================================================

// StructureDefinitionStore is a thread-safe in-memory store for StructureDefinition resources.
type StructureDefinitionStore struct {
	mu   sync.RWMutex
	defs map[string]*StructureDefinitionResource
}

// NewStructureDefinitionStore creates a new empty store.
func NewStructureDefinitionStore() *StructureDefinitionStore {
	return &StructureDefinitionStore{
		defs: make(map[string]*StructureDefinitionResource),
	}
}

// Register adds or replaces a StructureDefinition in the store.
func (s *StructureDefinitionStore) Register(sd *StructureDefinitionResource) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.defs[sd.ID] = sd
}

// Get returns a StructureDefinition by ID, or nil if not found.
func (s *StructureDefinitionStore) Get(id string) *StructureDefinitionResource {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.defs[id]
}

// Search returns StructureDefinitions matching the provided search parameters.
// Supported parameters: name, type, url, status.
func (s *StructureDefinitionStore) Search(params map[string]string) []*StructureDefinitionResource {
	s.mu.RLock()
	defer s.mu.RUnlock()

	results := make([]*StructureDefinitionResource, 0)
	for _, sd := range s.defs {
		if !matchesSDParams(sd, params) {
			continue
		}
		results = append(results, sd)
	}
	return results
}

// matchesSDParams checks if a StructureDefinition matches all provided search parameters.
func matchesSDParams(sd *StructureDefinitionResource, params map[string]string) bool {
	if v, ok := params["name"]; ok && v != "" {
		if !strings.EqualFold(sd.Name, v) {
			return false
		}
	}
	if v, ok := params["type"]; ok && v != "" {
		if !strings.EqualFold(sd.Type, v) {
			return false
		}
	}
	if v, ok := params["url"]; ok && v != "" {
		if sd.URL != v {
			return false
		}
	}
	if v, ok := params["status"]; ok && v != "" {
		if sd.Status != v {
			return false
		}
	}
	return true
}

// ============================================================================
// Pre-registered Base FHIR R4 Resource Definitions
// ============================================================================

// intPtr is a helper that returns a pointer to an int.
func intPtr(v int) *int {
	return &v
}

// RegisterBaseDefinitions populates the store with base FHIR R4 StructureDefinitions
// for the 20 most commonly used resource types.
func RegisterBaseDefinitions(store *StructureDefinitionStore) {
	baseURL := "http://hl7.org/fhir/StructureDefinition/"

	// Common base elements shared by every resource.
	baseElements := func(typeName string) []ElementDefinition {
		return []ElementDefinition{
			{ID: typeName, Path: typeName, Short: typeName + " Resource", Min: intPtr(0), Max: "*"},
			{ID: typeName + ".id", Path: typeName + ".id", Short: "Logical id of this artifact", Min: intPtr(0), Max: "1", Type: []ElementType{{Code: "id"}}},
			{ID: typeName + ".meta", Path: typeName + ".meta", Short: "Metadata about the resource", Min: intPtr(0), Max: "1", Type: []ElementType{{Code: "Meta"}}},
			{ID: typeName + ".text", Path: typeName + ".text", Short: "Text summary of the resource", Min: intPtr(0), Max: "1", Type: []ElementType{{Code: "Narrative"}}},
		}
	}

	// --- Patient ---
	patientElements := append(baseElements("Patient"),
		ElementDefinition{ID: "Patient.identifier", Path: "Patient.identifier", Short: "An identifier for this patient", Min: intPtr(0), Max: "*", Type: []ElementType{{Code: "Identifier"}}},
		ElementDefinition{ID: "Patient.active", Path: "Patient.active", Short: "Whether this patient record is active", Min: intPtr(0), Max: "1", Type: []ElementType{{Code: "boolean"}}},
		ElementDefinition{ID: "Patient.name", Path: "Patient.name", Short: "A name associated with the patient", Min: intPtr(0), Max: "*", Type: []ElementType{{Code: "HumanName"}}},
		ElementDefinition{ID: "Patient.gender", Path: "Patient.gender", Short: "male | female | other | unknown", Min: intPtr(0), Max: "1", Type: []ElementType{{Code: "code"}}, Binding: &ElementBinding{Strength: "required", ValueSet: "http://hl7.org/fhir/ValueSet/administrative-gender"}},
		ElementDefinition{ID: "Patient.birthDate", Path: "Patient.birthDate", Short: "The date of birth for the individual", Min: intPtr(0), Max: "1", Type: []ElementType{{Code: "date"}}},
	)
	store.Register(&StructureDefinitionResource{
		ResourceType: "StructureDefinition", ID: "Patient", URL: baseURL + "Patient",
		Name: "Patient", Title: "Patient", Status: "active", Kind: "resource",
		Abstract: false, Type: "Patient", FHIRVersion: "4.0.1",
		BaseDefinition: "http://hl7.org/fhir/StructureDefinition/DomainResource", Derivation: "specialization",
		Description: "Demographics and other administrative information about an individual receiving care.",
		Snapshot:    &StructureSnapshot{Element: patientElements},
	})

	// --- Observation ---
	observationElements := append(baseElements("Observation"),
		ElementDefinition{ID: "Observation.status", Path: "Observation.status", Short: "registered | preliminary | final | amended +", Min: intPtr(1), Max: "1", Type: []ElementType{{Code: "code"}}, Binding: &ElementBinding{Strength: "required", ValueSet: "http://hl7.org/fhir/ValueSet/observation-status"}},
		ElementDefinition{ID: "Observation.code", Path: "Observation.code", Short: "Type of observation (code / type)", Min: intPtr(1), Max: "1", Type: []ElementType{{Code: "CodeableConcept"}}},
		ElementDefinition{ID: "Observation.subject", Path: "Observation.subject", Short: "Who this is about", Min: intPtr(0), Max: "1", Type: []ElementType{{Code: "Reference", TargetProfile: []string{"http://hl7.org/fhir/StructureDefinition/Patient"}}}},
		ElementDefinition{ID: "Observation.value[x]", Path: "Observation.value[x]", Short: "Actual result", Min: intPtr(0), Max: "1", Type: []ElementType{{Code: "Quantity"}, {Code: "string"}, {Code: "CodeableConcept"}}},
		ElementDefinition{ID: "Observation.effective[x]", Path: "Observation.effective[x]", Short: "Clinically relevant time for observation", Min: intPtr(0), Max: "1", Type: []ElementType{{Code: "dateTime"}, {Code: "Period"}}},
	)
	store.Register(&StructureDefinitionResource{
		ResourceType: "StructureDefinition", ID: "Observation", URL: baseURL + "Observation",
		Name: "Observation", Title: "Observation", Status: "active", Kind: "resource",
		Abstract: false, Type: "Observation", FHIRVersion: "4.0.1",
		BaseDefinition: "http://hl7.org/fhir/StructureDefinition/DomainResource", Derivation: "specialization",
		Description: "Measurements and simple assertions made about a patient.",
		Snapshot:    &StructureSnapshot{Element: observationElements},
	})

	// --- Condition ---
	conditionElements := append(baseElements("Condition"),
		ElementDefinition{ID: "Condition.clinicalStatus", Path: "Condition.clinicalStatus", Short: "active | recurrence | relapse | inactive | remission | resolved", Min: intPtr(0), Max: "1", Type: []ElementType{{Code: "CodeableConcept"}}, Binding: &ElementBinding{Strength: "required", ValueSet: "http://hl7.org/fhir/ValueSet/condition-clinical"}},
		ElementDefinition{ID: "Condition.verificationStatus", Path: "Condition.verificationStatus", Short: "unconfirmed | provisional | differential | confirmed | refuted | entered-in-error", Min: intPtr(0), Max: "1", Type: []ElementType{{Code: "CodeableConcept"}}},
		ElementDefinition{ID: "Condition.code", Path: "Condition.code", Short: "Identification of the condition, problem or diagnosis", Min: intPtr(0), Max: "1", Type: []ElementType{{Code: "CodeableConcept"}}},
		ElementDefinition{ID: "Condition.subject", Path: "Condition.subject", Short: "Who has the condition", Min: intPtr(1), Max: "1", Type: []ElementType{{Code: "Reference", TargetProfile: []string{"http://hl7.org/fhir/StructureDefinition/Patient"}}}},
		ElementDefinition{ID: "Condition.onset[x]", Path: "Condition.onset[x]", Short: "Estimated or actual date/time", Min: intPtr(0), Max: "1", Type: []ElementType{{Code: "dateTime"}, {Code: "Age"}, {Code: "Period"}}},
	)
	store.Register(&StructureDefinitionResource{
		ResourceType: "StructureDefinition", ID: "Condition", URL: baseURL + "Condition",
		Name: "Condition", Title: "Condition", Status: "active", Kind: "resource",
		Abstract: false, Type: "Condition", FHIRVersion: "4.0.1",
		BaseDefinition: "http://hl7.org/fhir/StructureDefinition/DomainResource", Derivation: "specialization",
		Description: "A clinical condition, problem, diagnosis, or other event.",
		Snapshot:    &StructureSnapshot{Element: conditionElements},
	})

	// --- Encounter ---
	encounterElements := append(baseElements("Encounter"),
		ElementDefinition{ID: "Encounter.status", Path: "Encounter.status", Short: "planned | arrived | triaged | in-progress | onleave | finished | cancelled +", Min: intPtr(1), Max: "1", Type: []ElementType{{Code: "code"}}, Binding: &ElementBinding{Strength: "required", ValueSet: "http://hl7.org/fhir/ValueSet/encounter-status"}},
		ElementDefinition{ID: "Encounter.class", Path: "Encounter.class", Short: "Classification of patient encounter", Min: intPtr(1), Max: "1", Type: []ElementType{{Code: "Coding"}}},
		ElementDefinition{ID: "Encounter.subject", Path: "Encounter.subject", Short: "The patient present at the encounter", Min: intPtr(0), Max: "1", Type: []ElementType{{Code: "Reference", TargetProfile: []string{"http://hl7.org/fhir/StructureDefinition/Patient"}}}},
		ElementDefinition{ID: "Encounter.period", Path: "Encounter.period", Short: "The start and end time of the encounter", Min: intPtr(0), Max: "1", Type: []ElementType{{Code: "Period"}}},
		ElementDefinition{ID: "Encounter.reasonCode", Path: "Encounter.reasonCode", Short: "Coded reason the encounter takes place", Min: intPtr(0), Max: "*", Type: []ElementType{{Code: "CodeableConcept"}}},
	)
	store.Register(&StructureDefinitionResource{
		ResourceType: "StructureDefinition", ID: "Encounter", URL: baseURL + "Encounter",
		Name: "Encounter", Title: "Encounter", Status: "active", Kind: "resource",
		Abstract: false, Type: "Encounter", FHIRVersion: "4.0.1",
		BaseDefinition: "http://hl7.org/fhir/StructureDefinition/DomainResource", Derivation: "specialization",
		Description: "An interaction between a patient and healthcare provider(s).",
		Snapshot:    &StructureSnapshot{Element: encounterElements},
	})

	// --- MedicationRequest ---
	medicationRequestElements := append(baseElements("MedicationRequest"),
		ElementDefinition{ID: "MedicationRequest.status", Path: "MedicationRequest.status", Short: "active | on-hold | cancelled | completed | entered-in-error | stopped | draft | unknown", Min: intPtr(1), Max: "1", Type: []ElementType{{Code: "code"}}, Binding: &ElementBinding{Strength: "required", ValueSet: "http://hl7.org/fhir/ValueSet/medicationrequest-status"}},
		ElementDefinition{ID: "MedicationRequest.intent", Path: "MedicationRequest.intent", Short: "proposal | plan | order | original-order | reflex-order | filler-order | instance-order | option", Min: intPtr(1), Max: "1", Type: []ElementType{{Code: "code"}}},
		ElementDefinition{ID: "MedicationRequest.medication[x]", Path: "MedicationRequest.medication[x]", Short: "Medication to be taken", Min: intPtr(1), Max: "1", Type: []ElementType{{Code: "CodeableConcept"}, {Code: "Reference"}}},
		ElementDefinition{ID: "MedicationRequest.subject", Path: "MedicationRequest.subject", Short: "Who the medication request is for", Min: intPtr(1), Max: "1", Type: []ElementType{{Code: "Reference", TargetProfile: []string{"http://hl7.org/fhir/StructureDefinition/Patient"}}}},
		ElementDefinition{ID: "MedicationRequest.dosageInstruction", Path: "MedicationRequest.dosageInstruction", Short: "How the medication should be taken", Min: intPtr(0), Max: "*", Type: []ElementType{{Code: "Dosage"}}},
	)
	store.Register(&StructureDefinitionResource{
		ResourceType: "StructureDefinition", ID: "MedicationRequest", URL: baseURL + "MedicationRequest",
		Name: "MedicationRequest", Title: "MedicationRequest", Status: "active", Kind: "resource",
		Abstract: false, Type: "MedicationRequest", FHIRVersion: "4.0.1",
		BaseDefinition: "http://hl7.org/fhir/StructureDefinition/DomainResource", Derivation: "specialization",
		Description: "An order or request for a medication.",
		Snapshot:    &StructureSnapshot{Element: medicationRequestElements},
	})

	// --- Procedure ---
	procedureElements := append(baseElements("Procedure"),
		ElementDefinition{ID: "Procedure.status", Path: "Procedure.status", Short: "preparation | in-progress | not-done | on-hold | stopped | completed | entered-in-error | unknown", Min: intPtr(1), Max: "1", Type: []ElementType{{Code: "code"}}, Binding: &ElementBinding{Strength: "required", ValueSet: "http://hl7.org/fhir/ValueSet/event-status"}},
		ElementDefinition{ID: "Procedure.code", Path: "Procedure.code", Short: "Identification of the procedure", Min: intPtr(0), Max: "1", Type: []ElementType{{Code: "CodeableConcept"}}},
		ElementDefinition{ID: "Procedure.subject", Path: "Procedure.subject", Short: "Who the procedure was performed on", Min: intPtr(1), Max: "1", Type: []ElementType{{Code: "Reference", TargetProfile: []string{"http://hl7.org/fhir/StructureDefinition/Patient"}}}},
		ElementDefinition{ID: "Procedure.performed[x]", Path: "Procedure.performed[x]", Short: "When the procedure was performed", Min: intPtr(0), Max: "1", Type: []ElementType{{Code: "dateTime"}, {Code: "Period"}}},
	)
	store.Register(&StructureDefinitionResource{
		ResourceType: "StructureDefinition", ID: "Procedure", URL: baseURL + "Procedure",
		Name: "Procedure", Title: "Procedure", Status: "active", Kind: "resource",
		Abstract: false, Type: "Procedure", FHIRVersion: "4.0.1",
		BaseDefinition: "http://hl7.org/fhir/StructureDefinition/DomainResource", Derivation: "specialization",
		Description: "An action that is or was performed on or for a patient.",
		Snapshot:    &StructureSnapshot{Element: procedureElements},
	})

	// --- DiagnosticReport ---
	diagnosticReportElements := append(baseElements("DiagnosticReport"),
		ElementDefinition{ID: "DiagnosticReport.status", Path: "DiagnosticReport.status", Short: "registered | partial | preliminary | final +", Min: intPtr(1), Max: "1", Type: []ElementType{{Code: "code"}}, Binding: &ElementBinding{Strength: "required", ValueSet: "http://hl7.org/fhir/ValueSet/diagnostic-report-status"}},
		ElementDefinition{ID: "DiagnosticReport.code", Path: "DiagnosticReport.code", Short: "Name/Code for this diagnostic report", Min: intPtr(1), Max: "1", Type: []ElementType{{Code: "CodeableConcept"}}},
		ElementDefinition{ID: "DiagnosticReport.subject", Path: "DiagnosticReport.subject", Short: "The subject of the report", Min: intPtr(0), Max: "1", Type: []ElementType{{Code: "Reference", TargetProfile: []string{"http://hl7.org/fhir/StructureDefinition/Patient"}}}},
		ElementDefinition{ID: "DiagnosticReport.result", Path: "DiagnosticReport.result", Short: "Observations", Min: intPtr(0), Max: "*", Type: []ElementType{{Code: "Reference", TargetProfile: []string{"http://hl7.org/fhir/StructureDefinition/Observation"}}}},
	)
	store.Register(&StructureDefinitionResource{
		ResourceType: "StructureDefinition", ID: "DiagnosticReport", URL: baseURL + "DiagnosticReport",
		Name: "DiagnosticReport", Title: "DiagnosticReport", Status: "active", Kind: "resource",
		Abstract: false, Type: "DiagnosticReport", FHIRVersion: "4.0.1",
		BaseDefinition: "http://hl7.org/fhir/StructureDefinition/DomainResource", Derivation: "specialization",
		Description: "The findings and interpretation of diagnostic tests.",
		Snapshot:    &StructureSnapshot{Element: diagnosticReportElements},
	})

	// --- AllergyIntolerance ---
	allergyElements := append(baseElements("AllergyIntolerance"),
		ElementDefinition{ID: "AllergyIntolerance.clinicalStatus", Path: "AllergyIntolerance.clinicalStatus", Short: "active | inactive | resolved", Min: intPtr(0), Max: "1", Type: []ElementType{{Code: "CodeableConcept"}}},
		ElementDefinition{ID: "AllergyIntolerance.code", Path: "AllergyIntolerance.code", Short: "Code that identifies the allergy or intolerance", Min: intPtr(0), Max: "1", Type: []ElementType{{Code: "CodeableConcept"}}},
		ElementDefinition{ID: "AllergyIntolerance.patient", Path: "AllergyIntolerance.patient", Short: "Who the sensitivity is for", Min: intPtr(1), Max: "1", Type: []ElementType{{Code: "Reference", TargetProfile: []string{"http://hl7.org/fhir/StructureDefinition/Patient"}}}},
		ElementDefinition{ID: "AllergyIntolerance.criticality", Path: "AllergyIntolerance.criticality", Short: "low | high | unable-to-assess", Min: intPtr(0), Max: "1", Type: []ElementType{{Code: "code"}}},
	)
	store.Register(&StructureDefinitionResource{
		ResourceType: "StructureDefinition", ID: "AllergyIntolerance", URL: baseURL + "AllergyIntolerance",
		Name: "AllergyIntolerance", Title: "AllergyIntolerance", Status: "active", Kind: "resource",
		Abstract: false, Type: "AllergyIntolerance", FHIRVersion: "4.0.1",
		BaseDefinition: "http://hl7.org/fhir/StructureDefinition/DomainResource", Derivation: "specialization",
		Description: "Risk of harmful or undesirable reaction to a substance.",
		Snapshot:    &StructureSnapshot{Element: allergyElements},
	})

	// --- Immunization ---
	immunizationElements := append(baseElements("Immunization"),
		ElementDefinition{ID: "Immunization.status", Path: "Immunization.status", Short: "completed | entered-in-error | not-done", Min: intPtr(1), Max: "1", Type: []ElementType{{Code: "code"}}, Binding: &ElementBinding{Strength: "required", ValueSet: "http://hl7.org/fhir/ValueSet/immunization-status"}},
		ElementDefinition{ID: "Immunization.vaccineCode", Path: "Immunization.vaccineCode", Short: "Vaccine product administered", Min: intPtr(1), Max: "1", Type: []ElementType{{Code: "CodeableConcept"}}},
		ElementDefinition{ID: "Immunization.patient", Path: "Immunization.patient", Short: "Who was immunized", Min: intPtr(1), Max: "1", Type: []ElementType{{Code: "Reference", TargetProfile: []string{"http://hl7.org/fhir/StructureDefinition/Patient"}}}},
		ElementDefinition{ID: "Immunization.occurrence[x]", Path: "Immunization.occurrence[x]", Short: "Vaccine administration date", Min: intPtr(1), Max: "1", Type: []ElementType{{Code: "dateTime"}, {Code: "string"}}},
	)
	store.Register(&StructureDefinitionResource{
		ResourceType: "StructureDefinition", ID: "Immunization", URL: baseURL + "Immunization",
		Name: "Immunization", Title: "Immunization", Status: "active", Kind: "resource",
		Abstract: false, Type: "Immunization", FHIRVersion: "4.0.1",
		BaseDefinition: "http://hl7.org/fhir/StructureDefinition/DomainResource", Derivation: "specialization",
		Description: "Describes the event of a patient being administered a vaccine.",
		Snapshot:    &StructureSnapshot{Element: immunizationElements},
	})

	// --- CarePlan ---
	carePlanElements := append(baseElements("CarePlan"),
		ElementDefinition{ID: "CarePlan.status", Path: "CarePlan.status", Short: "draft | active | on-hold | revoked | completed | entered-in-error | unknown", Min: intPtr(1), Max: "1", Type: []ElementType{{Code: "code"}}, Binding: &ElementBinding{Strength: "required", ValueSet: "http://hl7.org/fhir/ValueSet/request-status"}},
		ElementDefinition{ID: "CarePlan.intent", Path: "CarePlan.intent", Short: "proposal | plan | order | option", Min: intPtr(1), Max: "1", Type: []ElementType{{Code: "code"}}},
		ElementDefinition{ID: "CarePlan.subject", Path: "CarePlan.subject", Short: "Who the care plan is for", Min: intPtr(1), Max: "1", Type: []ElementType{{Code: "Reference", TargetProfile: []string{"http://hl7.org/fhir/StructureDefinition/Patient"}}}},
		ElementDefinition{ID: "CarePlan.activity", Path: "CarePlan.activity", Short: "Action to occur as part of plan", Min: intPtr(0), Max: "*", Type: []ElementType{{Code: "BackboneElement"}}},
	)
	store.Register(&StructureDefinitionResource{
		ResourceType: "StructureDefinition", ID: "CarePlan", URL: baseURL + "CarePlan",
		Name: "CarePlan", Title: "CarePlan", Status: "active", Kind: "resource",
		Abstract: false, Type: "CarePlan", FHIRVersion: "4.0.1",
		BaseDefinition: "http://hl7.org/fhir/StructureDefinition/DomainResource", Derivation: "specialization",
		Description: "Describes the intention of how one or more practitioners deliver care for a patient.",
		Snapshot:    &StructureSnapshot{Element: carePlanElements},
	})

	// --- Medication ---
	medicationElements := append(baseElements("Medication"),
		ElementDefinition{ID: "Medication.code", Path: "Medication.code", Short: "Codes that identify this medication", Min: intPtr(0), Max: "1", Type: []ElementType{{Code: "CodeableConcept"}}},
		ElementDefinition{ID: "Medication.status", Path: "Medication.status", Short: "active | inactive | entered-in-error", Min: intPtr(0), Max: "1", Type: []ElementType{{Code: "code"}}},
		ElementDefinition{ID: "Medication.form", Path: "Medication.form", Short: "powder | tablets | capsule +", Min: intPtr(0), Max: "1", Type: []ElementType{{Code: "CodeableConcept"}}},
	)
	store.Register(&StructureDefinitionResource{
		ResourceType: "StructureDefinition", ID: "Medication", URL: baseURL + "Medication",
		Name: "Medication", Title: "Medication", Status: "active", Kind: "resource",
		Abstract: false, Type: "Medication", FHIRVersion: "4.0.1",
		BaseDefinition: "http://hl7.org/fhir/StructureDefinition/DomainResource", Derivation: "specialization",
		Description: "Definition of a medication.",
		Snapshot:    &StructureSnapshot{Element: medicationElements},
	})

	// --- Practitioner ---
	practitionerElements := append(baseElements("Practitioner"),
		ElementDefinition{ID: "Practitioner.identifier", Path: "Practitioner.identifier", Short: "An identifier for the person", Min: intPtr(0), Max: "*", Type: []ElementType{{Code: "Identifier"}}},
		ElementDefinition{ID: "Practitioner.active", Path: "Practitioner.active", Short: "Whether this practitioner record is active", Min: intPtr(0), Max: "1", Type: []ElementType{{Code: "boolean"}}},
		ElementDefinition{ID: "Practitioner.name", Path: "Practitioner.name", Short: "The name(s) associated with the practitioner", Min: intPtr(0), Max: "*", Type: []ElementType{{Code: "HumanName"}}},
		ElementDefinition{ID: "Practitioner.qualification", Path: "Practitioner.qualification", Short: "Certification, licenses, or training", Min: intPtr(0), Max: "*", Type: []ElementType{{Code: "BackboneElement"}}},
	)
	store.Register(&StructureDefinitionResource{
		ResourceType: "StructureDefinition", ID: "Practitioner", URL: baseURL + "Practitioner",
		Name: "Practitioner", Title: "Practitioner", Status: "active", Kind: "resource",
		Abstract: false, Type: "Practitioner", FHIRVersion: "4.0.1",
		BaseDefinition: "http://hl7.org/fhir/StructureDefinition/DomainResource", Derivation: "specialization",
		Description: "A person who is directly or indirectly involved in the provisioning of healthcare.",
		Snapshot:    &StructureSnapshot{Element: practitionerElements},
	})

	// --- Organization ---
	organizationElements := append(baseElements("Organization"),
		ElementDefinition{ID: "Organization.identifier", Path: "Organization.identifier", Short: "Identifies this organization across multiple systems", Min: intPtr(0), Max: "*", Type: []ElementType{{Code: "Identifier"}}},
		ElementDefinition{ID: "Organization.active", Path: "Organization.active", Short: "Whether the organization is still active", Min: intPtr(0), Max: "1", Type: []ElementType{{Code: "boolean"}}},
		ElementDefinition{ID: "Organization.name", Path: "Organization.name", Short: "Name used for the organization", Min: intPtr(0), Max: "1", Type: []ElementType{{Code: "string"}}},
		ElementDefinition{ID: "Organization.type", Path: "Organization.type", Short: "Kind of organization", Min: intPtr(0), Max: "*", Type: []ElementType{{Code: "CodeableConcept"}}},
	)
	store.Register(&StructureDefinitionResource{
		ResourceType: "StructureDefinition", ID: "Organization", URL: baseURL + "Organization",
		Name: "Organization", Title: "Organization", Status: "active", Kind: "resource",
		Abstract: false, Type: "Organization", FHIRVersion: "4.0.1",
		BaseDefinition: "http://hl7.org/fhir/StructureDefinition/DomainResource", Derivation: "specialization",
		Description: "A formally or informally recognized grouping of people or organizations.",
		Snapshot:    &StructureSnapshot{Element: organizationElements},
	})

	// --- Location ---
	locationElements := append(baseElements("Location"),
		ElementDefinition{ID: "Location.status", Path: "Location.status", Short: "active | suspended | inactive", Min: intPtr(0), Max: "1", Type: []ElementType{{Code: "code"}}},
		ElementDefinition{ID: "Location.name", Path: "Location.name", Short: "Name of the location", Min: intPtr(0), Max: "1", Type: []ElementType{{Code: "string"}}},
		ElementDefinition{ID: "Location.mode", Path: "Location.mode", Short: "instance | kind", Min: intPtr(0), Max: "1", Type: []ElementType{{Code: "code"}}},
		ElementDefinition{ID: "Location.address", Path: "Location.address", Short: "Physical location", Min: intPtr(0), Max: "1", Type: []ElementType{{Code: "Address"}}},
	)
	store.Register(&StructureDefinitionResource{
		ResourceType: "StructureDefinition", ID: "Location", URL: baseURL + "Location",
		Name: "Location", Title: "Location", Status: "active", Kind: "resource",
		Abstract: false, Type: "Location", FHIRVersion: "4.0.1",
		BaseDefinition: "http://hl7.org/fhir/StructureDefinition/DomainResource", Derivation: "specialization",
		Description: "Details and position information for a physical place.",
		Snapshot:    &StructureSnapshot{Element: locationElements},
	})

	// --- ServiceRequest ---
	serviceRequestElements := append(baseElements("ServiceRequest"),
		ElementDefinition{ID: "ServiceRequest.status", Path: "ServiceRequest.status", Short: "draft | active | on-hold | revoked | completed | entered-in-error | unknown", Min: intPtr(1), Max: "1", Type: []ElementType{{Code: "code"}}},
		ElementDefinition{ID: "ServiceRequest.intent", Path: "ServiceRequest.intent", Short: "proposal | plan | directive | order +", Min: intPtr(1), Max: "1", Type: []ElementType{{Code: "code"}}},
		ElementDefinition{ID: "ServiceRequest.code", Path: "ServiceRequest.code", Short: "What is being requested/ordered", Min: intPtr(0), Max: "1", Type: []ElementType{{Code: "CodeableConcept"}}},
		ElementDefinition{ID: "ServiceRequest.subject", Path: "ServiceRequest.subject", Short: "Individual or Entity the service is ordered for", Min: intPtr(1), Max: "1", Type: []ElementType{{Code: "Reference", TargetProfile: []string{"http://hl7.org/fhir/StructureDefinition/Patient"}}}},
	)
	store.Register(&StructureDefinitionResource{
		ResourceType: "StructureDefinition", ID: "ServiceRequest", URL: baseURL + "ServiceRequest",
		Name: "ServiceRequest", Title: "ServiceRequest", Status: "active", Kind: "resource",
		Abstract: false, Type: "ServiceRequest", FHIRVersion: "4.0.1",
		BaseDefinition: "http://hl7.org/fhir/StructureDefinition/DomainResource", Derivation: "specialization",
		Description: "A record of a request for service such as diagnostic investigations or treatments.",
		Snapshot:    &StructureSnapshot{Element: serviceRequestElements},
	})

	// --- DocumentReference ---
	documentReferenceElements := append(baseElements("DocumentReference"),
		ElementDefinition{ID: "DocumentReference.status", Path: "DocumentReference.status", Short: "current | superseded | entered-in-error", Min: intPtr(1), Max: "1", Type: []ElementType{{Code: "code"}}},
		ElementDefinition{ID: "DocumentReference.type", Path: "DocumentReference.type", Short: "Kind of document", Min: intPtr(0), Max: "1", Type: []ElementType{{Code: "CodeableConcept"}}},
		ElementDefinition{ID: "DocumentReference.subject", Path: "DocumentReference.subject", Short: "Who/what is the subject of the document", Min: intPtr(0), Max: "1", Type: []ElementType{{Code: "Reference", TargetProfile: []string{"http://hl7.org/fhir/StructureDefinition/Patient"}}}},
		ElementDefinition{ID: "DocumentReference.content", Path: "DocumentReference.content", Short: "Document referenced", Min: intPtr(1), Max: "*", Type: []ElementType{{Code: "BackboneElement"}}},
	)
	store.Register(&StructureDefinitionResource{
		ResourceType: "StructureDefinition", ID: "DocumentReference", URL: baseURL + "DocumentReference",
		Name: "DocumentReference", Title: "DocumentReference", Status: "active", Kind: "resource",
		Abstract: false, Type: "DocumentReference", FHIRVersion: "4.0.1",
		BaseDefinition: "http://hl7.org/fhir/StructureDefinition/DomainResource", Derivation: "specialization",
		Description: "A reference to a document of any kind for any purpose.",
		Snapshot:    &StructureSnapshot{Element: documentReferenceElements},
	})

	// --- MedicationAdministration ---
	medicationAdminElements := append(baseElements("MedicationAdministration"),
		ElementDefinition{ID: "MedicationAdministration.status", Path: "MedicationAdministration.status", Short: "in-progress | not-done | on-hold | completed | entered-in-error | stopped | unknown", Min: intPtr(1), Max: "1", Type: []ElementType{{Code: "code"}}},
		ElementDefinition{ID: "MedicationAdministration.medication[x]", Path: "MedicationAdministration.medication[x]", Short: "What was administered", Min: intPtr(1), Max: "1", Type: []ElementType{{Code: "CodeableConcept"}, {Code: "Reference"}}},
		ElementDefinition{ID: "MedicationAdministration.subject", Path: "MedicationAdministration.subject", Short: "Who received medication", Min: intPtr(1), Max: "1", Type: []ElementType{{Code: "Reference", TargetProfile: []string{"http://hl7.org/fhir/StructureDefinition/Patient"}}}},
		ElementDefinition{ID: "MedicationAdministration.effective[x]", Path: "MedicationAdministration.effective[x]", Short: "Start and end time of administration", Min: intPtr(1), Max: "1", Type: []ElementType{{Code: "dateTime"}, {Code: "Period"}}},
	)
	store.Register(&StructureDefinitionResource{
		ResourceType: "StructureDefinition", ID: "MedicationAdministration", URL: baseURL + "MedicationAdministration",
		Name: "MedicationAdministration", Title: "MedicationAdministration", Status: "active", Kind: "resource",
		Abstract: false, Type: "MedicationAdministration", FHIRVersion: "4.0.1",
		BaseDefinition: "http://hl7.org/fhir/StructureDefinition/DomainResource", Derivation: "specialization",
		Description: "Describes the event of a patient consuming or being administered a medication.",
		Snapshot:    &StructureSnapshot{Element: medicationAdminElements},
	})

	// --- Goal ---
	goalElements := append(baseElements("Goal"),
		ElementDefinition{ID: "Goal.lifecycleStatus", Path: "Goal.lifecycleStatus", Short: "proposed | planned | accepted | active | on-hold | completed | cancelled | entered-in-error | rejected", Min: intPtr(1), Max: "1", Type: []ElementType{{Code: "code"}}, Binding: &ElementBinding{Strength: "required", ValueSet: "http://hl7.org/fhir/ValueSet/goal-status"}},
		ElementDefinition{ID: "Goal.description", Path: "Goal.description", Short: "Code or text describing goal", Min: intPtr(1), Max: "1", Type: []ElementType{{Code: "CodeableConcept"}}},
		ElementDefinition{ID: "Goal.subject", Path: "Goal.subject", Short: "Who this goal is intended for", Min: intPtr(1), Max: "1", Type: []ElementType{{Code: "Reference", TargetProfile: []string{"http://hl7.org/fhir/StructureDefinition/Patient"}}}},
		ElementDefinition{ID: "Goal.target", Path: "Goal.target", Short: "Target outcome for the goal", Min: intPtr(0), Max: "*", Type: []ElementType{{Code: "BackboneElement"}}},
	)
	store.Register(&StructureDefinitionResource{
		ResourceType: "StructureDefinition", ID: "Goal", URL: baseURL + "Goal",
		Name: "Goal", Title: "Goal", Status: "active", Kind: "resource",
		Abstract: false, Type: "Goal", FHIRVersion: "4.0.1",
		BaseDefinition: "http://hl7.org/fhir/StructureDefinition/DomainResource", Derivation: "specialization",
		Description: "Describes the intended objective(s) for a patient.",
		Snapshot:    &StructureSnapshot{Element: goalElements},
	})

	// --- Claim ---
	claimElements := append(baseElements("Claim"),
		ElementDefinition{ID: "Claim.status", Path: "Claim.status", Short: "active | cancelled | draft | entered-in-error", Min: intPtr(1), Max: "1", Type: []ElementType{{Code: "code"}}, Binding: &ElementBinding{Strength: "required", ValueSet: "http://hl7.org/fhir/ValueSet/fm-status"}},
		ElementDefinition{ID: "Claim.type", Path: "Claim.type", Short: "Category or discipline", Min: intPtr(1), Max: "1", Type: []ElementType{{Code: "CodeableConcept"}}},
		ElementDefinition{ID: "Claim.patient", Path: "Claim.patient", Short: "The recipient of the products and services", Min: intPtr(1), Max: "1", Type: []ElementType{{Code: "Reference", TargetProfile: []string{"http://hl7.org/fhir/StructureDefinition/Patient"}}}},
		ElementDefinition{ID: "Claim.provider", Path: "Claim.provider", Short: "Party responsible for the claim", Min: intPtr(1), Max: "1", Type: []ElementType{{Code: "Reference", TargetProfile: []string{"http://hl7.org/fhir/StructureDefinition/Practitioner"}}}},
		ElementDefinition{ID: "Claim.use", Path: "Claim.use", Short: "claim | preauthorization | predetermination", Min: intPtr(1), Max: "1", Type: []ElementType{{Code: "code"}}},
	)
	store.Register(&StructureDefinitionResource{
		ResourceType: "StructureDefinition", ID: "Claim", URL: baseURL + "Claim",
		Name: "Claim", Title: "Claim", Status: "active", Kind: "resource",
		Abstract: false, Type: "Claim", FHIRVersion: "4.0.1",
		BaseDefinition: "http://hl7.org/fhir/StructureDefinition/DomainResource", Derivation: "specialization",
		Description: "A provider issued list of professional services for reimbursement.",
		Snapshot:    &StructureSnapshot{Element: claimElements},
	})
}

// ============================================================================
// Snapshot Generation
// ============================================================================

// GenerateSnapshot produces a snapshot for a StructureDefinition that has a
// differential and a base definition. If the definition already has a snapshot,
// it is returned as-is. If no differential exists, the base definition's
// snapshot is used. Otherwise, the differential elements are merged on top of
// the base snapshot.
func GenerateSnapshot(store *StructureDefinitionStore, sd *StructureDefinitionResource) *StructureDefinitionResource {
	if sd.Snapshot != nil {
		return sd
	}

	// Look up the base definition.
	baseSD := resolveBase(store, sd.BaseDefinition)

	if sd.Differential == nil {
		if baseSD != nil && baseSD.Snapshot != nil {
			result := *sd
			result.Snapshot = baseSD.Snapshot
			return &result
		}
		return sd
	}

	// Build snapshot by merging differential onto the base snapshot.
	var baseElements []ElementDefinition
	if baseSD != nil && baseSD.Snapshot != nil {
		baseElements = baseSD.Snapshot.Element
	}

	merged := mergeElements(baseElements, sd.Differential.Element)
	result := *sd
	result.Snapshot = &StructureSnapshot{Element: merged}
	return &result
}

// resolveBase looks up a base StructureDefinition by URL from the store.
func resolveBase(store *StructureDefinitionStore, baseURL string) *StructureDefinitionResource {
	if baseURL == "" || store == nil {
		return nil
	}
	// Search by URL.
	results := store.Search(map[string]string{"url": baseURL})
	if len(results) > 0 {
		return results[0]
	}
	return nil
}

// mergeElements merges differential elements onto base snapshot elements.
// Elements with matching paths are overridden; new elements from the
// differential are appended.
func mergeElements(base, differential []ElementDefinition) []ElementDefinition {
	result := make([]ElementDefinition, len(base))
	copy(result, base)

	baseIndex := make(map[string]int, len(base))
	for i, e := range result {
		baseIndex[e.Path] = i
	}

	for _, de := range differential {
		if idx, ok := baseIndex[de.Path]; ok {
			result[idx] = de
		} else {
			result = append(result, de)
		}
	}
	return result
}

// ============================================================================
// HTTP Handler
// ============================================================================

// StructureDefinitionHandler provides HTTP handlers for the StructureDefinition
// resource endpoint.
type StructureDefinitionHandler struct {
	store *StructureDefinitionStore
}

// NewStructureDefinitionHandler creates a new handler with base definitions pre-registered.
func NewStructureDefinitionHandler() *StructureDefinitionHandler {
	store := NewStructureDefinitionStore()
	RegisterBaseDefinitions(store)
	return &StructureDefinitionHandler{store: store}
}

// RegisterRoutes registers StructureDefinition routes on the provided Echo group.
// Expects the group to be the FHIR base (e.g., /fhir).
func (h *StructureDefinitionHandler) RegisterRoutes(fhirGroup *echo.Group) {
	fhirGroup.GET("/StructureDefinition", h.SearchStructureDefinitions)
	fhirGroup.GET("/StructureDefinition/:id", h.GetStructureDefinition)
	fhirGroup.GET("/StructureDefinition/$snapshot", h.GenerateSnapshotOp)
}

// SearchStructureDefinitions handles GET /StructureDefinition with optional query filters.
func (h *StructureDefinitionHandler) SearchStructureDefinitions(c echo.Context) error {
	params := make(map[string]string)
	for _, key := range []string{"name", "type", "url", "status"} {
		if v := c.QueryParam(key); v != "" {
			params[key] = v
		}
	}

	results := h.store.Search(params)

	resources := make([]interface{}, 0, len(results))
	for _, sd := range results {
		resources = append(resources, sd)
	}
	return c.JSON(http.StatusOK, NewSearchBundle(resources, len(resources), "/fhir/StructureDefinition"))
}

// GetStructureDefinition handles GET /StructureDefinition/:id.
func (h *StructureDefinitionHandler) GetStructureDefinition(c echo.Context) error {
	id := c.Param("id")
	sd := h.store.Get(id)
	if sd == nil {
		return c.JSON(http.StatusNotFound, NotFoundOutcome("StructureDefinition", id))
	}
	return c.JSON(http.StatusOK, sd)
}

// GenerateSnapshotOp handles GET /StructureDefinition/$snapshot.
// It accepts a "url" query parameter identifying the StructureDefinition to
// generate a snapshot for.
func (h *StructureDefinitionHandler) GenerateSnapshotOp(c echo.Context) error {
	url := c.QueryParam("url")
	if url == "" {
		return c.JSON(http.StatusBadRequest, ErrorOutcome("url parameter is required for $snapshot"))
	}

	results := h.store.Search(map[string]string{"url": url})
	if len(results) == 0 {
		return c.JSON(http.StatusNotFound, ErrorOutcome("StructureDefinition not found for url: "+url))
	}

	sd := results[0]
	expanded := GenerateSnapshot(h.store, sd)
	return c.JSON(http.StatusOK, expanded)
}

// snapshotRequestBody is used to parse POST $snapshot requests.
type snapshotRequestBody struct {
	ResourceType string          `json:"resourceType"`
	Parameter    []snapshotParam `json:"parameter,omitempty"`
}

type snapshotParam struct {
	Name     string                       `json:"name"`
	Resource *StructureDefinitionResource `json:"resource,omitempty"`
}

// GenerateSnapshotFromBody handles a POST $snapshot operation that receives a
// StructureDefinition in the request body as a Parameters resource.
func (h *StructureDefinitionHandler) GenerateSnapshotFromBody(c echo.Context) error {
	var body snapshotRequestBody
	if err := json.NewDecoder(c.Request().Body).Decode(&body); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorOutcome("invalid JSON: "+err.Error()))
	}

	var sd *StructureDefinitionResource
	for _, p := range body.Parameter {
		if p.Name == "definition" && p.Resource != nil {
			sd = p.Resource
			break
		}
	}

	if sd == nil {
		return c.JSON(http.StatusBadRequest, ErrorOutcome("parameter 'definition' with a StructureDefinition resource is required"))
	}

	expanded := GenerateSnapshot(h.store, sd)
	return c.JSON(http.StatusOK, expanded)
}
