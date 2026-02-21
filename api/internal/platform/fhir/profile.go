package fhir

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/labstack/echo/v4"
)

// ---------------------------------------------------------------------------
// US Core IG v6.1.0 canonical profile URLs
// ---------------------------------------------------------------------------

const (
	USCorePatientURL             = "http://hl7.org/fhir/us/core/StructureDefinition/us-core-patient"
	USCoreConditionURL                   = "http://hl7.org/fhir/us/core/StructureDefinition/us-core-condition-problems-health-concerns"
	USCoreConditionEncounterDiagnosisURL = "http://hl7.org/fhir/us/core/StructureDefinition/us-core-condition-encounter-diagnosis"
	USCoreObservationLabURL      = "http://hl7.org/fhir/us/core/StructureDefinition/us-core-observation-lab"
	USCoreVitalSignsURL          = "http://hl7.org/fhir/StructureDefinition/vitalsigns"
	USCoreSmokingStatusURL       = "http://hl7.org/fhir/us/core/StructureDefinition/us-core-smokingstatus"
	USCoreObservationSDOHURL     = "http://hl7.org/fhir/us/core/StructureDefinition/us-core-observation-sdoh-assessment"
	USCoreAllergyIntoleranceURL  = "http://hl7.org/fhir/us/core/StructureDefinition/us-core-allergyintolerance"
	USCoreMedicationRequestURL   = "http://hl7.org/fhir/us/core/StructureDefinition/us-core-medicationrequest"
	USCoreEncounterURL           = "http://hl7.org/fhir/us/core/StructureDefinition/us-core-encounter"
	USCoreProcedureURL           = "http://hl7.org/fhir/us/core/StructureDefinition/us-core-procedure"
	USCoreImmunizationURL        = "http://hl7.org/fhir/us/core/StructureDefinition/us-core-immunization"
	USCoreDiagnosticReportLabURL = "http://hl7.org/fhir/us/core/StructureDefinition/us-core-diagnosticreport-lab"
	USCoreDocumentReferenceURL   = "http://hl7.org/fhir/us/core/StructureDefinition/us-core-documentreference"
)

// ---------------------------------------------------------------------------
// Profile Definition Model
// ---------------------------------------------------------------------------

// ProfileDefinition represents a FHIR StructureDefinition-based profile.
type ProfileDefinition struct {
	URL         string              // canonical URL
	Name        string              // human-readable name
	Type        string              // base resource type (e.g., "Patient")
	Version     string              // profile version
	Status      string              // draft|active|retired
	Constraints []ProfileConstraint // element-level rules
}

// ProfileConstraint defines a validation rule for a specific element.
type ProfileConstraint struct {
	Path        string                 // FHIRPath-style path (e.g., "Patient.identifier")
	Min         int                    // minimum cardinality (0 = optional, 1+ = required)
	Max         string                 // maximum cardinality ("1", "*", "0" for prohibited)
	Types       []string               // allowed types
	MustSupport bool                   // MS flag
	Binding     *ProfileBinding        // terminology binding
	Pattern     map[string]interface{} // fixed/pattern value
	Invariants  []string               // FHIRPath invariant expressions (for future use)
}

// ProfileBinding represents a value set binding on a coded element.
type ProfileBinding struct {
	Strength string // required|extensible|preferred|example
	ValueSet string // ValueSet URL
}

// ---------------------------------------------------------------------------
// Profile Validation Issue
// ---------------------------------------------------------------------------

// ProfileValidationIssue represents a single validation finding.
type ProfileValidationIssue struct {
	Severity    string // error|warning|information
	Code        string // required|value|invariant|structure|not-found
	Path        string // element path
	Description string
	ProfileURL  string
}

// ---------------------------------------------------------------------------
// Profile Registry
// ---------------------------------------------------------------------------

// ProfileRegistry stores and looks up profile definitions.
type ProfileRegistry struct {
	mu       sync.RWMutex
	byURL    map[string]*ProfileDefinition
	byType   map[string][]*ProfileDefinition
}

// NewProfileRegistry creates a new empty ProfileRegistry.
func NewProfileRegistry() *ProfileRegistry {
	return &ProfileRegistry{
		byURL:  make(map[string]*ProfileDefinition),
		byType: make(map[string][]*ProfileDefinition),
	}
}

// Register adds or replaces a profile definition in the registry.
func (r *ProfileRegistry) Register(profile ProfileDefinition) {
	r.mu.Lock()
	defer r.mu.Unlock()

	p := profile // copy

	// If a profile with this URL already exists, remove from type index.
	if existing, ok := r.byURL[p.URL]; ok {
		r.removeFromTypeIndex(existing)
	}

	r.byURL[p.URL] = &p
	r.byType[p.Type] = append(r.byType[p.Type], &p)
}

// removeFromTypeIndex removes a profile pointer from the type index.
// Must be called with mu held.
func (r *ProfileRegistry) removeFromTypeIndex(p *ProfileDefinition) {
	profiles := r.byType[p.Type]
	for i, pp := range profiles {
		if pp.URL == p.URL {
			r.byType[p.Type] = append(profiles[:i], profiles[i+1:]...)
			break
		}
	}
}

// GetByURL returns the profile with the given canonical URL.
func (r *ProfileRegistry) GetByURL(url string) (*ProfileDefinition, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.byURL[url]
	return p, ok
}

// GetByType returns all profiles for the given resource type.
func (r *ProfileRegistry) GetByType(resourceType string) []ProfileDefinition {
	r.mu.RLock()
	defer r.mu.RUnlock()

	ptrs := r.byType[resourceType]
	result := make([]ProfileDefinition, 0, len(ptrs))
	for _, p := range ptrs {
		result = append(result, *p)
	}
	return result
}

// ListAll returns all registered profiles.
func (r *ProfileRegistry) ListAll() []ProfileDefinition {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]ProfileDefinition, 0, len(r.byURL))
	for _, p := range r.byURL {
		result = append(result, *p)
	}
	return result
}

// ---------------------------------------------------------------------------
// Profile Validator
// ---------------------------------------------------------------------------

// ProfileValidator validates FHIR resources against registered profiles.
type ProfileValidator struct {
	registry *ProfileRegistry
}

// NewProfileValidator creates a new ProfileValidator.
func NewProfileValidator(registry *ProfileRegistry) *ProfileValidator {
	return &ProfileValidator{registry: registry}
}

// ValidateAgainstProfile validates a resource against a specific profile URL.
func (v *ProfileValidator) ValidateAgainstProfile(resource map[string]interface{}, profileURL string) []ProfileValidationIssue {
	if resource == nil {
		return []ProfileValidationIssue{{
			Severity:    "error",
			Code:        "structure",
			Description: "resource is nil",
			ProfileURL:  profileURL,
		}}
	}

	profile, ok := v.registry.GetByURL(profileURL)
	if !ok {
		return []ProfileValidationIssue{{
			Severity:    "error",
			Code:        "not-found",
			Description: fmt.Sprintf("profile '%s' not found in registry", profileURL),
			ProfileURL:  profileURL,
		}}
	}

	// Check resourceType matches
	rt, _ := resource["resourceType"].(string)
	if rt == "" {
		return []ProfileValidationIssue{{
			Severity:    "error",
			Code:        "structure",
			Description: "resourceType is missing or empty",
			ProfileURL:  profileURL,
		}}
	}
	if rt != profile.Type {
		return []ProfileValidationIssue{{
			Severity:    "error",
			Code:        "structure",
			Description: fmt.Sprintf("resource type mismatch: resource is '%s' but profile '%s' is for '%s'", rt, profile.Name, profile.Type),
			ProfileURL:  profileURL,
		}}
	}

	return v.validateConstraints(resource, profile)
}

// ValidateResource validates a resource against all applicable profiles for its type.
func (v *ProfileValidator) ValidateResource(resource map[string]interface{}) []ProfileValidationIssue {
	if resource == nil {
		return nil
	}

	rt, _ := resource["resourceType"].(string)
	if rt == "" {
		return nil
	}

	profiles := v.registry.GetByType(rt)
	if len(profiles) == 0 {
		return nil
	}

	var allIssues []ProfileValidationIssue
	for _, p := range profiles {
		issues := v.validateConstraints(resource, &p)
		allIssues = append(allIssues, issues...)
	}
	return allIssues
}

// validateConstraints checks all constraints in a profile against a resource.
func (v *ProfileValidator) validateConstraints(resource map[string]interface{}, profile *ProfileDefinition) []ProfileValidationIssue {
	var issues []ProfileValidationIssue

	for _, c := range profile.Constraints {
		cIssues := v.evaluateConstraint(resource, profile, c)
		issues = append(issues, cIssues...)
	}

	// Run profile-specific business rules (e.g., category code checks, name family/given).
	specificIssues := v.validateProfileSpecificRules(resource, profile)
	issues = append(issues, specificIssues...)

	return issues
}

// evaluateConstraint evaluates a single profile constraint against a resource.
func (v *ProfileValidator) evaluateConstraint(resource map[string]interface{}, profile *ProfileDefinition, c ProfileConstraint) []ProfileValidationIssue {
	var issues []ProfileValidationIssue

	// Parse the path: "Patient.identifier.system" -> ["identifier", "system"]
	// The first segment is the resource type, skip it.
	parts := strings.Split(c.Path, ".")
	if len(parts) < 2 {
		return nil
	}

	fieldParts := parts[1:] // everything after the resource type

	// Handle choice types: e.g., "medication[x]", "effective[x]", "value[x]", "occurrence[x]", "performed[x]"
	if isChoiceType(fieldParts[0]) {
		return v.evaluateChoiceConstraint(resource, profile, c, fieldParts[0])
	}

	// Resolve the value at the path
	val, present := resolveFieldPath(resource, fieldParts)

	// Check cardinality
	if c.Min > 0 {
		if !present || isEmptyValue(val) {
			issues = append(issues, ProfileValidationIssue{
				Severity:    "error",
				Code:        "required",
				Path:        c.Path,
				Description: fmt.Sprintf("element '%s' is required (min=%d) by profile '%s'", c.Path, c.Min, profile.Name),
				ProfileURL:  profile.URL,
			})
			return issues // no point checking further
		}
	}

	// Check MustSupport: missing MS fields generate warnings
	if c.MustSupport && c.Min == 0 {
		if !present || isEmptyValue(val) {
			issues = append(issues, ProfileValidationIssue{
				Severity:    "warning",
				Code:        "invariant",
				Path:        c.Path,
				Description: fmt.Sprintf("MustSupport element '%s' is not present", c.Path),
				ProfileURL:  profile.URL,
			})
		}
	}

	// Check max="0" (prohibited element)
	if c.Max == "0" && present && !isEmptyValue(val) {
		issues = append(issues, ProfileValidationIssue{
			Severity:    "error",
			Code:        "structure",
			Path:        c.Path,
			Description: fmt.Sprintf("element '%s' is prohibited (max=0) by profile '%s'", c.Path, profile.Name),
			ProfileURL:  profile.URL,
		})
	}

	// Check required binding on coded elements
	if c.Binding != nil && c.Binding.Strength == "required" && present && !isEmptyValue(val) {
		bindingIssues := v.validateBinding(val, c, profile)
		issues = append(issues, bindingIssues...)
	}

	// Check sub-element constraints for arrays
	// For example, Patient.identifier requires items to have .system and .value
	if present && !isEmptyValue(val) {
		subIssues := v.validateSubElements(resource, profile, c, fieldParts)
		issues = append(issues, subIssues...)
	}

	return issues
}

// evaluateChoiceConstraint handles choice-type elements like medication[x], effective[x], value[x].
func (v *ProfileValidator) evaluateChoiceConstraint(resource map[string]interface{}, profile *ProfileDefinition, c ProfileConstraint, choiceField string) []ProfileValidationIssue {
	baseName := choiceField[:len(choiceField)-3] // strip "[x]"
	suffixes := choiceTypeSuffixes(baseName)

	present := false
	for _, suffix := range suffixes {
		if val, ok := resource[baseName+suffix]; ok && !isEmptyValue(val) {
			present = true
			break
		}
	}

	if c.Min > 0 && !present {
		return []ProfileValidationIssue{{
			Severity:    "error",
			Code:        "required",
			Path:        c.Path,
			Description: fmt.Sprintf("element '%s' is required (min=%d) by profile '%s'", c.Path, c.Min, profile.Name),
			ProfileURL:  profile.URL,
		}}
	}

	if c.MustSupport && c.Min == 0 && !present {
		return []ProfileValidationIssue{{
			Severity:    "warning",
			Code:        "invariant",
			Path:        c.Path,
			Description: fmt.Sprintf("MustSupport element '%s' is not present", c.Path),
			ProfileURL:  profile.URL,
		}}
	}

	return nil
}

// isChoiceType returns true if the field name ends with "[x]".
func isChoiceType(field string) bool {
	return strings.HasSuffix(field, "[x]")
}

// choiceTypeSuffixes returns the common FHIR type suffixes for choice types.
func choiceTypeSuffixes(baseName string) []string {
	return []string{
		"DateTime", "Period", "Timing", "Instant",
		"CodeableConcept", "Coding", "Reference",
		"Quantity", "Range", "Ratio",
		"Boolean", "Integer", "String",
		"Date", "Time",
		"Attachment", "Identifier",
		"Age", "Duration", "SampledData",
	}
}

// resolveFieldPath walks the resource map to find a value at the given path.
func resolveFieldPath(resource map[string]interface{}, parts []string) (interface{}, bool) {
	if len(parts) == 0 {
		return resource, true
	}

	current := parts[0]
	val, ok := resource[current]
	if !ok {
		return nil, false
	}

	if len(parts) == 1 {
		return val, true
	}

	// Navigate deeper
	remaining := parts[1:]
	switch v := val.(type) {
	case map[string]interface{}:
		return resolveFieldPath(v, remaining)
	case []interface{}:
		// For arrays, we check each item
		for _, item := range v {
			if m, ok := item.(map[string]interface{}); ok {
				if val, ok := resolveFieldPath(m, remaining); ok {
					return val, true
				}
			}
		}
		return nil, false
	}

	return nil, false
}

// isEmptyValue returns true if a value is nil, empty string, empty array, or empty map.
func isEmptyValue(val interface{}) bool {
	if val == nil {
		return true
	}
	switch v := val.(type) {
	case string:
		return v == ""
	case []interface{}:
		return len(v) == 0
	case map[string]interface{}:
		return len(v) == 0
	}
	return false
}

// validateBinding checks that a coded value is from the required value set.
func (v *ProfileValidator) validateBinding(val interface{}, c ProfileConstraint, profile *ProfileDefinition) []ProfileValidationIssue {
	// For gender binding validation
	if c.Binding.ValueSet == "http://hl7.org/fhir/ValueSet/administrative-gender" {
		str, ok := val.(string)
		if ok {
			validGenders := map[string]bool{"male": true, "female": true, "other": true, "unknown": true}
			if !validGenders[str] {
				return []ProfileValidationIssue{{
					Severity:    "error",
					Code:        "value",
					Path:        c.Path,
					Description: fmt.Sprintf("value '%s' is not in required binding '%s'", str, c.Binding.ValueSet),
					ProfileURL:  profile.URL,
				}}
			}
		}
	}
	return nil
}

// validateSubElements handles sub-element constraints within arrays (e.g., identifier items).
func (v *ProfileValidator) validateSubElements(resource map[string]interface{}, profile *ProfileDefinition, c ProfileConstraint, fieldParts []string) []ProfileValidationIssue {
	// Sub-element validation is handled by constraints with deeper paths
	// (e.g., Patient.identifier.system is a separate constraint)
	return nil
}

// ---------------------------------------------------------------------------
// Built-in US Core Profile Registration
// ---------------------------------------------------------------------------

// RegisterUSCoreProfiles registers all built-in US Core IG v6.1.0 profiles.
func RegisterUSCoreProfiles(reg *ProfileRegistry) {
	registerUSCorePatient(reg)
	registerUSCoreCondition(reg)
	registerUSCoreObservationLab(reg)
	registerUSCoreAllergyIntolerance(reg)
	registerUSCoreMedicationRequest(reg)
	registerUSCoreEncounter(reg)
	registerUSCoreProcedure(reg)
	registerUSCoreImmunization(reg)
	registerUSCoreDiagnosticReportLab(reg)
	registerUSCoreDocumentReference(reg)
}

func registerUSCorePatient(reg *ProfileRegistry) {
	reg.Register(ProfileDefinition{
		URL:     USCorePatientURL,
		Name:    "USCorePatient",
		Type:    "Patient",
		Version: "6.1.0",
		Status:  "active",
		Constraints: []ProfileConstraint{
			{Path: "Patient.identifier", Min: 1, Max: "*"},
			{Path: "Patient.identifier.system", Min: 1, Max: "1"},
			{Path: "Patient.identifier.value", Min: 1, Max: "1"},
			{Path: "Patient.name", Min: 1, Max: "*"},
			{Path: "Patient.gender", Min: 1, Max: "1", Binding: &ProfileBinding{
				Strength: "required",
				ValueSet: "http://hl7.org/fhir/ValueSet/administrative-gender",
			}},
			{Path: "Patient.birthDate", Min: 0, Max: "1", MustSupport: true},
			{Path: "Patient.address", Min: 0, Max: "*", MustSupport: true},
			{Path: "Patient.telecom", Min: 0, Max: "*", MustSupport: true},
			{Path: "Patient.communication", Min: 0, Max: "*", MustSupport: true},
		},
	})
}

func registerUSCoreCondition(reg *ProfileRegistry) {
	reg.Register(ProfileDefinition{
		URL:     USCoreConditionURL,
		Name:    "USCoreConditionProblemsHealthConcerns",
		Type:    "Condition",
		Version: "6.1.0",
		Status:  "active",
		Constraints: []ProfileConstraint{
			{Path: "Condition.clinicalStatus", Min: 1, Max: "1", Binding: &ProfileBinding{
				Strength: "required",
				ValueSet: "http://hl7.org/fhir/ValueSet/condition-clinical",
			}},
			{Path: "Condition.verificationStatus", Min: 0, Max: "1", MustSupport: true},
			{Path: "Condition.category", Min: 1, Max: "*", Binding: &ProfileBinding{
				Strength: "extensible",
				ValueSet: "http://hl7.org/fhir/ValueSet/condition-category",
			}},
			{Path: "Condition.code", Min: 1, Max: "1", Binding: &ProfileBinding{
				Strength: "extensible",
				ValueSet: "http://hl7.org/fhir/us/core/ValueSet/us-core-condition-code",
			}},
			{Path: "Condition.subject", Min: 1, Max: "1"},
		},
	})
}

func registerUSCoreObservationLab(reg *ProfileRegistry) {
	reg.Register(ProfileDefinition{
		URL:     USCoreObservationLabURL,
		Name:    "USCoreObservationLab",
		Type:    "Observation",
		Version: "6.1.0",
		Status:  "active",
		Constraints: []ProfileConstraint{
			{Path: "Observation.status", Min: 1, Max: "1", Binding: &ProfileBinding{
				Strength: "required",
				ValueSet: "http://hl7.org/fhir/ValueSet/observation-status",
			}},
			{Path: "Observation.category", Min: 1, Max: "*"},
			{Path: "Observation.code", Min: 1, Max: "1", Binding: &ProfileBinding{
				Strength: "extensible",
				ValueSet: "http://hl7.org/fhir/ValueSet/observation-codes",
			}},
			{Path: "Observation.subject", Min: 1, Max: "1"},
			{Path: "Observation.effective[x]", Min: 0, Max: "1", MustSupport: true},
			{Path: "Observation.value[x]", Min: 0, Max: "1", MustSupport: true},
		},
	})
}

func registerUSCoreAllergyIntolerance(reg *ProfileRegistry) {
	reg.Register(ProfileDefinition{
		URL:     USCoreAllergyIntoleranceURL,
		Name:    "USCoreAllergyIntolerance",
		Type:    "AllergyIntolerance",
		Version: "6.1.0",
		Status:  "active",
		Constraints: []ProfileConstraint{
			{Path: "AllergyIntolerance.clinicalStatus", Min: 0, Max: "1", MustSupport: true, Binding: &ProfileBinding{
				Strength: "required",
				ValueSet: "http://hl7.org/fhir/ValueSet/allergyintolerance-clinical",
			}},
			{Path: "AllergyIntolerance.verificationStatus", Min: 0, Max: "1", MustSupport: true},
			{Path: "AllergyIntolerance.code", Min: 1, Max: "1"},
			{Path: "AllergyIntolerance.patient", Min: 1, Max: "1"},
		},
	})
}

func registerUSCoreMedicationRequest(reg *ProfileRegistry) {
	reg.Register(ProfileDefinition{
		URL:     USCoreMedicationRequestURL,
		Name:    "USCoreMedicationRequest",
		Type:    "MedicationRequest",
		Version: "6.1.0",
		Status:  "active",
		Constraints: []ProfileConstraint{
			{Path: "MedicationRequest.status", Min: 1, Max: "1", Binding: &ProfileBinding{
				Strength: "required",
				ValueSet: "http://hl7.org/fhir/ValueSet/medicationrequest-status",
			}},
			{Path: "MedicationRequest.intent", Min: 1, Max: "1", Binding: &ProfileBinding{
				Strength: "required",
				ValueSet: "http://hl7.org/fhir/ValueSet/medicationrequest-intent",
			}},
			{Path: "MedicationRequest.medication[x]", Min: 1, Max: "1"},
			{Path: "MedicationRequest.subject", Min: 1, Max: "1"},
			{Path: "MedicationRequest.authoredOn", Min: 0, Max: "1", MustSupport: true},
			{Path: "MedicationRequest.requester", Min: 0, Max: "1", MustSupport: true},
			{Path: "MedicationRequest.dosageInstruction", Min: 0, Max: "*", MustSupport: true},
		},
	})
}

func registerUSCoreEncounter(reg *ProfileRegistry) {
	reg.Register(ProfileDefinition{
		URL:     USCoreEncounterURL,
		Name:    "USCoreEncounter",
		Type:    "Encounter",
		Version: "6.1.0",
		Status:  "active",
		Constraints: []ProfileConstraint{
			{Path: "Encounter.status", Min: 1, Max: "1"},
			{Path: "Encounter.class", Min: 1, Max: "1"},
			{Path: "Encounter.type", Min: 1, Max: "*"},
			{Path: "Encounter.subject", Min: 1, Max: "1"},
			{Path: "Encounter.period", Min: 0, Max: "1", MustSupport: true},
		},
	})
}

func registerUSCoreProcedure(reg *ProfileRegistry) {
	reg.Register(ProfileDefinition{
		URL:     USCoreProcedureURL,
		Name:    "USCoreProcedure",
		Type:    "Procedure",
		Version: "6.1.0",
		Status:  "active",
		Constraints: []ProfileConstraint{
			{Path: "Procedure.status", Min: 1, Max: "1"},
			{Path: "Procedure.code", Min: 1, Max: "1"},
			{Path: "Procedure.subject", Min: 1, Max: "1"},
			{Path: "Procedure.performed[x]", Min: 0, Max: "1", MustSupport: true},
		},
	})
}

func registerUSCoreImmunization(reg *ProfileRegistry) {
	reg.Register(ProfileDefinition{
		URL:     USCoreImmunizationURL,
		Name:    "USCoreImmunization",
		Type:    "Immunization",
		Version: "6.1.0",
		Status:  "active",
		Constraints: []ProfileConstraint{
			{Path: "Immunization.status", Min: 1, Max: "1"},
			{Path: "Immunization.vaccineCode", Min: 1, Max: "1"},
			{Path: "Immunization.patient", Min: 1, Max: "1"},
			{Path: "Immunization.occurrence[x]", Min: 1, Max: "1"},
			{Path: "Immunization.primarySource", Min: 0, Max: "1", MustSupport: true},
		},
	})
}

func registerUSCoreDiagnosticReportLab(reg *ProfileRegistry) {
	reg.Register(ProfileDefinition{
		URL:     USCoreDiagnosticReportLabURL,
		Name:    "USCoreDiagnosticReportLab",
		Type:    "DiagnosticReport",
		Version: "6.1.0",
		Status:  "active",
		Constraints: []ProfileConstraint{
			{Path: "DiagnosticReport.status", Min: 1, Max: "1"},
			{Path: "DiagnosticReport.category", Min: 1, Max: "*"},
			{Path: "DiagnosticReport.code", Min: 1, Max: "1"},
			{Path: "DiagnosticReport.subject", Min: 1, Max: "1"},
			{Path: "DiagnosticReport.effective[x]", Min: 0, Max: "1", MustSupport: true},
			{Path: "DiagnosticReport.result", Min: 0, Max: "*", MustSupport: true},
		},
	})
}

func registerUSCoreDocumentReference(reg *ProfileRegistry) {
	reg.Register(ProfileDefinition{
		URL:     USCoreDocumentReferenceURL,
		Name:    "USCoreDocumentReference",
		Type:    "DocumentReference",
		Version: "6.1.0",
		Status:  "active",
		Constraints: []ProfileConstraint{
			{Path: "DocumentReference.status", Min: 1, Max: "1"},
			{Path: "DocumentReference.type", Min: 1, Max: "1"},
			{Path: "DocumentReference.category", Min: 1, Max: "*"},
			{Path: "DocumentReference.subject", Min: 1, Max: "1"},
			{Path: "DocumentReference.date", Min: 0, Max: "1", MustSupport: true},
			{Path: "DocumentReference.content", Min: 1, Max: "*"},
			{Path: "DocumentReference.content.attachment", Min: 1, Max: "1"},
			{Path: "DocumentReference.content.attachment.contentType", Min: 1, Max: "1"},
		},
	})
}

// ---------------------------------------------------------------------------
// Special validation: US Core Patient name requires family OR given
// ---------------------------------------------------------------------------

// validateConstraints is augmented with profile-specific business rules.
func (v *ProfileValidator) validateProfileSpecificRules(resource map[string]interface{}, profile *ProfileDefinition) []ProfileValidationIssue {
	var issues []ProfileValidationIssue

	switch profile.URL {
	case USCorePatientURL:
		issues = append(issues, v.validateUSCorePatientName(resource, profile)...)
		issues = append(issues, v.validateUSCorePatientIdentifierSubElements(resource, profile)...)
	case USCoreObservationLabURL:
		issues = append(issues, v.validateUSCoreObservationLabCategory(resource, profile)...)
	case USCoreDiagnosticReportLabURL:
		issues = append(issues, v.validateUSCoreDiagnosticReportLabCategory(resource, profile)...)
	case USCoreDocumentReferenceURL:
		issues = append(issues, v.validateUSCoreDocumentReferenceContent(resource, profile)...)
	}

	return issues
}

func (v *ProfileValidator) validateUSCorePatientName(resource map[string]interface{}, profile *ProfileDefinition) []ProfileValidationIssue {
	nameVal, ok := resource["name"]
	if !ok {
		return nil // already caught by cardinality check
	}

	names, ok := nameVal.([]interface{})
	if !ok || len(names) == 0 {
		return nil
	}

	// Check at least one name has family or given
	for _, nameRaw := range names {
		nameMap, ok := nameRaw.(map[string]interface{})
		if !ok {
			continue
		}

		hasFamily := false
		hasGiven := false

		if f, ok := nameMap["family"]; ok {
			if s, ok := f.(string); ok && s != "" {
				hasFamily = true
			}
		}
		if g, ok := nameMap["given"]; ok {
			if arr, ok := g.([]interface{}); ok && len(arr) > 0 {
				hasGiven = true
			}
		}

		if hasFamily || hasGiven {
			return nil // at least one name has family or given
		}
	}

	return []ProfileValidationIssue{{
		Severity:    "error",
		Code:        "required",
		Path:        "Patient.name",
		Description: "US Core Patient requires name to have at least family or given",
		ProfileURL:  profile.URL,
	}}
}

func (v *ProfileValidator) validateUSCorePatientIdentifierSubElements(resource map[string]interface{}, profile *ProfileDefinition) []ProfileValidationIssue {
	idVal, ok := resource["identifier"]
	if !ok {
		return nil // already caught by cardinality
	}

	ids, ok := idVal.([]interface{})
	if !ok || len(ids) == 0 {
		return nil
	}

	var issues []ProfileValidationIssue
	for _, idRaw := range ids {
		idMap, ok := idRaw.(map[string]interface{})
		if !ok {
			continue
		}

		if sys, ok := idMap["system"]; !ok || isEmptyValue(sys) {
			issues = append(issues, ProfileValidationIssue{
				Severity:    "error",
				Code:        "required",
				Path:        "Patient.identifier.system",
				Description: "US Core Patient identifier.system is required",
				ProfileURL:  profile.URL,
			})
		}
		if val, ok := idMap["value"]; !ok || isEmptyValue(val) {
			issues = append(issues, ProfileValidationIssue{
				Severity:    "error",
				Code:        "required",
				Path:        "Patient.identifier.value",
				Description: "US Core Patient identifier.value is required",
				ProfileURL:  profile.URL,
			})
		}
	}

	return issues
}

func (v *ProfileValidator) validateUSCoreObservationLabCategory(resource map[string]interface{}, profile *ProfileDefinition) []ProfileValidationIssue {
	catVal, ok := resource["category"]
	if !ok {
		return nil
	}

	cats, ok := catVal.([]interface{})
	if !ok || len(cats) == 0 {
		return nil
	}

	if !hasCategoryCode(cats, "laboratory") {
		return []ProfileValidationIssue{{
			Severity:    "error",
			Code:        "value",
			Path:        "Observation.category",
			Description: "US Core Observation Lab requires category to include 'laboratory'",
			ProfileURL:  profile.URL,
		}}
	}

	return nil
}

func (v *ProfileValidator) validateUSCoreDiagnosticReportLabCategory(resource map[string]interface{}, profile *ProfileDefinition) []ProfileValidationIssue {
	catVal, ok := resource["category"]
	if !ok {
		return nil
	}

	cats, ok := catVal.([]interface{})
	if !ok || len(cats) == 0 {
		return nil
	}

	if !hasCategoryCode(cats, "LAB") {
		return []ProfileValidationIssue{{
			Severity:    "error",
			Code:        "value",
			Path:        "DiagnosticReport.category",
			Description: "US Core DiagnosticReport Lab requires category to include 'LAB'",
			ProfileURL:  profile.URL,
		}}
	}

	return nil
}

func (v *ProfileValidator) validateUSCoreDocumentReferenceContent(resource map[string]interface{}, profile *ProfileDefinition) []ProfileValidationIssue {
	contentVal, ok := resource["content"]
	if !ok {
		return nil // caught by cardinality
	}

	contents, ok := contentVal.([]interface{})
	if !ok || len(contents) == 0 {
		return nil
	}

	var issues []ProfileValidationIssue
	for _, contentRaw := range contents {
		contentMap, ok := contentRaw.(map[string]interface{})
		if !ok {
			continue
		}

		attachVal, ok := contentMap["attachment"]
		if !ok || isEmptyValue(attachVal) {
			issues = append(issues, ProfileValidationIssue{
				Severity:    "error",
				Code:        "required",
				Path:        "DocumentReference.content.attachment",
				Description: "US Core DocumentReference content.attachment is required",
				ProfileURL:  profile.URL,
			})
			continue
		}

		attachMap, ok := attachVal.(map[string]interface{})
		if !ok {
			continue
		}

		if ct, ok := attachMap["contentType"]; !ok || isEmptyValue(ct) {
			issues = append(issues, ProfileValidationIssue{
				Severity:    "error",
				Code:        "required",
				Path:        "DocumentReference.content.attachment.contentType",
				Description: "US Core DocumentReference content.attachment.contentType is required",
				ProfileURL:  profile.URL,
			})
		}
	}

	return issues
}

// hasCategoryCode checks if any category in the array has a coding with the given code.
func hasCategoryCode(categories []interface{}, code string) bool {
	for _, catRaw := range categories {
		catMap, ok := catRaw.(map[string]interface{})
		if !ok {
			continue
		}
		codingsVal, ok := catMap["coding"]
		if !ok {
			continue
		}
		codings, ok := codingsVal.([]interface{})
		if !ok {
			continue
		}
		for _, codingRaw := range codings {
			codingMap, ok := codingRaw.(map[string]interface{})
			if !ok {
				continue
			}
			if c, ok := codingMap["code"].(string); ok && c == code {
				return true
			}
		}
	}
	return false
}

// ---------------------------------------------------------------------------
// HTTP Handler
// ---------------------------------------------------------------------------

// ProfileHandler provides HTTP endpoints for profile management and validation.
type ProfileHandler struct {
	validator *ProfileValidator
	registry  *ProfileRegistry
}

// NewProfileHandler creates a new ProfileHandler.
func NewProfileHandler(validator *ProfileValidator, registry *ProfileRegistry) *ProfileHandler {
	return &ProfileHandler{
		validator: validator,
		registry:  registry,
	}
}

// RegisterRoutes adds profile routes to the given FHIR group.
func (h *ProfileHandler) RegisterRoutes(g *echo.Group) {
	g.GET("/StructureDefinition", h.ListProfiles)
	g.GET("/StructureDefinition/:id", h.GetProfile)
	g.GET("/metadata/profiles", h.ListProfilesByType)
	g.POST("/metadata/profiles", h.RegisterCustomProfile)
}

// ListProfiles handles GET /fhir/StructureDefinition — returns all profiles as a Bundle.
func (h *ProfileHandler) ListProfiles(c echo.Context) error {
	profiles := h.registry.ListAll()

	entries := make([]map[string]interface{}, 0, len(profiles))
	for _, p := range profiles {
		entries = append(entries, profileToStructureDefinition(p))
	}

	bundle := map[string]interface{}{
		"resourceType": "Bundle",
		"type":         "searchset",
		"total":        len(entries),
		"entry":        entries,
	}

	return c.JSON(http.StatusOK, bundle)
}

// GetProfile handles GET /fhir/StructureDefinition/:id
func (h *ProfileHandler) GetProfile(c echo.Context) error {
	id := c.Param("id")

	// Try as canonical URL first
	if p, ok := h.registry.GetByURL(id); ok {
		return c.JSON(http.StatusOK, profileToStructureDefinition(*p))
	}

	// Try matching by name or by URL suffix
	for _, p := range h.registry.ListAll() {
		if p.Name == id {
			return c.JSON(http.StatusOK, profileToStructureDefinition(p))
		}
		// Match last segment of URL (e.g., "us-core-patient")
		parts := strings.Split(p.URL, "/")
		if len(parts) > 0 && parts[len(parts)-1] == id {
			return c.JSON(http.StatusOK, profileToStructureDefinition(p))
		}
	}

	return c.JSON(http.StatusNotFound, map[string]interface{}{
		"resourceType": "OperationOutcome",
		"issue": []map[string]interface{}{
			{
				"severity":    "error",
				"code":        "not-found",
				"diagnostics": fmt.Sprintf("StructureDefinition '%s' not found", id),
			},
		},
	})
}

// ValidateWithProfile handles POST /fhir/$validate with optional profile parameter.
func (h *ProfileHandler) ValidateWithProfile(c echo.Context) error {
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, errorOutcomeMap("Failed to read request body"))
	}

	if len(body) == 0 {
		return c.JSON(http.StatusBadRequest, errorOutcomeMap("Request body is empty"))
	}

	var resource map[string]interface{}
	if err := json.Unmarshal(body, &resource); err != nil {
		return c.JSON(http.StatusBadRequest, errorOutcomeMap("Invalid JSON: "+err.Error()))
	}

	profileURL := c.QueryParam("profile")

	var issues []ProfileValidationIssue
	if profileURL != "" {
		issues = h.validator.ValidateAgainstProfile(resource, profileURL)
	} else {
		issues = h.validator.ValidateResource(resource)
	}

	return c.JSON(http.StatusOK, profilesToOperationOutcome(issues))
}

// ListProfilesByType handles GET /fhir/metadata/profiles
func (h *ProfileHandler) ListProfilesByType(c echo.Context) error {
	resourceType := c.QueryParam("type")

	var profiles []ProfileDefinition
	if resourceType != "" {
		profiles = h.registry.GetByType(resourceType)
	} else {
		profiles = h.registry.ListAll()
	}

	result := make([]map[string]interface{}, 0, len(profiles))
	for _, p := range profiles {
		result = append(result, map[string]interface{}{
			"url":     p.URL,
			"name":    p.Name,
			"type":    p.Type,
			"version": p.Version,
			"status":  p.Status,
		})
	}

	return c.JSON(http.StatusOK, result)
}

// RegisterCustomProfile handles POST /fhir/metadata/profiles
func (h *ProfileHandler) RegisterCustomProfile(c echo.Context) error {
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, errorOutcomeMap("Failed to read request body"))
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(body, &raw); err != nil {
		return c.JSON(http.StatusBadRequest, errorOutcomeMap("Invalid JSON: "+err.Error()))
	}

	url, _ := raw["url"].(string)
	name, _ := raw["name"].(string)
	typ, _ := raw["type"].(string)
	version, _ := raw["version"].(string)
	status, _ := raw["status"].(string)

	if url == "" || typ == "" {
		return c.JSON(http.StatusBadRequest, errorOutcomeMap("url and type are required"))
	}

	profile := ProfileDefinition{
		URL:     url,
		Name:    name,
		Type:    typ,
		Version: version,
		Status:  status,
	}

	// Parse constraints
	if constraintsRaw, ok := raw["constraints"].([]interface{}); ok {
		for _, cRaw := range constraintsRaw {
			cMap, ok := cRaw.(map[string]interface{})
			if !ok {
				continue
			}
			constraint := ProfileConstraint{}
			if p, ok := cMap["path"].(string); ok {
				constraint.Path = p
			}
			if m, ok := cMap["min"].(float64); ok {
				constraint.Min = int(m)
			}
			if m, ok := cMap["max"].(string); ok {
				constraint.Max = m
			}
			if ms, ok := cMap["mustSupport"].(bool); ok {
				constraint.MustSupport = ms
			}
			profile.Constraints = append(profile.Constraints, constraint)
		}
	}

	h.registry.Register(profile)

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"url":     profile.URL,
		"name":    profile.Name,
		"type":    profile.Type,
		"version": profile.Version,
		"status":  profile.Status,
		"message": "Profile registered successfully",
	})
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func profileToStructureDefinition(p ProfileDefinition) map[string]interface{} {
	sd := map[string]interface{}{
		"resourceType": "StructureDefinition",
		"url":          p.URL,
		"name":         p.Name,
		"type":         p.Type,
		"version":      p.Version,
		"status":       p.Status,
		"kind":         "resource",
		"abstract":     false,
		"derivation":   "constraint",
		"baseDefinition": fmt.Sprintf("http://hl7.org/fhir/StructureDefinition/%s", p.Type),
	}

	if len(p.Constraints) > 0 {
		elements := make([]map[string]interface{}, 0, len(p.Constraints))
		for _, c := range p.Constraints {
			elem := map[string]interface{}{
				"path": c.Path,
				"min":  c.Min,
				"max":  c.Max,
			}
			if c.MustSupport {
				elem["mustSupport"] = true
			}
			if c.Binding != nil {
				elem["binding"] = map[string]interface{}{
					"strength": c.Binding.Strength,
					"valueSet": c.Binding.ValueSet,
				}
			}
			elements = append(elements, elem)
		}
		sd["snapshot"] = map[string]interface{}{
			"element": elements,
		}
	}

	return sd
}

func profilesToOperationOutcome(issues []ProfileValidationIssue) map[string]interface{} {
	if len(issues) == 0 {
		return map[string]interface{}{
			"resourceType": "OperationOutcome",
			"issue": []map[string]interface{}{
				{
					"severity":    "information",
					"code":        "informational",
					"diagnostics": "Validation successful - no issues found",
				},
			},
		}
	}

	outcomeIssues := make([]map[string]interface{}, 0, len(issues))
	for _, issue := range issues {
		entry := map[string]interface{}{
			"severity":    issue.Severity,
			"code":        issue.Code,
			"diagnostics": issue.Description,
		}
		if issue.Path != "" {
			entry["location"] = []string{issue.Path}
		}
		if issue.ProfileURL != "" {
			if entry["details"] == nil {
				entry["details"] = map[string]interface{}{
					"text": fmt.Sprintf("Profile: %s", issue.ProfileURL),
				}
			}
		}
		outcomeIssues = append(outcomeIssues, entry)
	}

	return map[string]interface{}{
		"resourceType": "OperationOutcome",
		"issue":        outcomeIssues,
	}
}

func errorOutcomeMap(message string) map[string]interface{} {
	return map[string]interface{}{
		"resourceType": "OperationOutcome",
		"issue": []map[string]interface{}{
			{
				"severity":    "error",
				"code":        "structure",
				"diagnostics": message,
			},
		},
	}
}
