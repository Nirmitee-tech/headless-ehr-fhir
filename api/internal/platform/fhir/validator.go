package fhir

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

// referencePattern matches FHIR references in the format "ResourceType/id".
var referencePattern = regexp.MustCompile(`^[A-Z][a-zA-Z]+/[a-zA-Z0-9\-\.]+$`)

// knownResourceTypes lists FHIR R4 resource types recognized by this server.
var knownResourceTypes = map[string]bool{
	"Patient": true, "Practitioner": true, "PractitionerRole": true,
	"Organization": true, "Location": true, "Encounter": true,
	"Condition": true, "Observation": true, "AllergyIntolerance": true,
	"Procedure": true, "Medication": true, "MedicationRequest": true,
	"MedicationAdministration": true, "MedicationDispense": true,
	"MedicationStatement": true, "ServiceRequest": true,
	"DiagnosticReport": true, "ImagingStudy": true, "Specimen": true,
	"Appointment": true, "Schedule": true, "Slot": true,
	"Coverage": true, "Claim": true, "ClaimResponse": true,
	"Consent": true, "DocumentReference": true, "Composition": true,
	"Communication": true, "ResearchStudy": true, "ResearchSubject": true,
	"Questionnaire": true, "QuestionnaireResponse": true,
	"Bundle": true, "OperationOutcome": true, "CapabilityStatement": true,
	"Invoice": true, "CareTeam": true, "CarePlan": true,
}

// statusValues maps resource types to their valid status values per FHIR R4.
var statusValues = map[string][]string{
	"Patient":                  {"active", "inactive", "entered-in-error"},
	"Practitioner":             {"active", "inactive", "entered-in-error"},
	"Organization":             {"active", "inactive", "entered-in-error"},
	"Encounter":                {"planned", "arrived", "triaged", "in-progress", "onleave", "finished", "cancelled", "entered-in-error", "unknown"},
	"Condition":                {"active", "recurrence", "relapse", "inactive", "remission", "resolved"},
	"Observation":              {"registered", "preliminary", "final", "amended", "corrected", "cancelled", "entered-in-error", "unknown"},
	"AllergyIntolerance":       {"active", "inactive", "resolved"},
	"Procedure":                {"preparation", "in-progress", "not-done", "on-hold", "stopped", "completed", "entered-in-error", "unknown"},
	"Medication":               {"active", "inactive", "entered-in-error"},
	"MedicationRequest":        {"active", "on-hold", "cancelled", "completed", "entered-in-error", "stopped", "draft", "unknown"},
	"MedicationAdministration": {"in-progress", "not-done", "on-hold", "completed", "entered-in-error", "stopped", "unknown"},
	"MedicationDispense":       {"preparation", "in-progress", "cancelled", "on-hold", "completed", "entered-in-error", "stopped", "declined", "unknown"},
	"MedicationStatement":      {"active", "completed", "entered-in-error", "intended", "stopped", "on-hold", "unknown", "not-taken"},
	"ServiceRequest":           {"draft", "active", "on-hold", "revoked", "completed", "entered-in-error", "unknown"},
	"DiagnosticReport":         {"registered", "partial", "preliminary", "final", "amended", "corrected", "appended", "cancelled", "entered-in-error", "unknown"},
	"Appointment":              {"proposed", "pending", "booked", "arrived", "fulfilled", "cancelled", "noshow", "entered-in-error", "checked-in", "waitlist"},
	"Slot":                     {"busy", "free", "busy-unavailable", "busy-tentative", "entered-in-error"},
	"Coverage":                 {"active", "cancelled", "draft", "entered-in-error"},
	"Claim":                    {"active", "cancelled", "draft", "entered-in-error"},
	"Consent":                  {"draft", "proposed", "active", "rejected", "inactive", "entered-in-error"},
	"DocumentReference":        {"current", "superseded", "entered-in-error"},
	"Composition":              {"preliminary", "final", "amended", "entered-in-error"},
	"Communication":            {"preparation", "in-progress", "not-done", "on-hold", "stopped", "completed", "entered-in-error", "unknown"},
	"ResearchStudy":            {"active", "administratively-completed", "approved", "closed-to-accrual", "closed-to-accrual-and-intervention", "completed", "disapproved", "in-review", "temporarily-closed-to-accrual", "temporarily-closed-to-accrual-and-intervention", "withdrawn"},
}

// ValidationResult holds the results of a FHIR resource validation.
type ValidationResult struct {
	Valid  bool
	Issues []OperationOutcomeIssue
}

// ToOperationOutcome converts a ValidationResult into an OperationOutcome.
func (vr *ValidationResult) ToOperationOutcome() *OperationOutcome {
	return &OperationOutcome{
		ResourceType: "OperationOutcome",
		Issue:        vr.Issues,
	}
}

// Validator provides FHIR R4 resource validation.
type Validator struct{}

// NewValidator creates a new FHIR Validator.
func NewValidator() *Validator {
	return &Validator{}
}

// ValidateResource validates a raw JSON resource per FHIR R4 rules.
// It checks: resourceType is present and known, id is present for updates,
// status values are valid, and references are properly formatted.
func (v *Validator) ValidateResource(data json.RawMessage, requireID bool) *ValidationResult {
	result := &ValidationResult{Valid: true}

	var resource map[string]interface{}
	if err := json.Unmarshal(data, &resource); err != nil {
		result.Valid = false
		result.Issues = append(result.Issues, OperationOutcomeIssue{
			Severity:    IssueSeverityError,
			Code:        IssueTypeStructure,
			Diagnostics: "invalid JSON: " + err.Error(),
		})
		return result
	}

	v.validateResourceType(resource, result)
	if requireID {
		v.validateID(resource, result)
	}
	v.validateStatus(resource, result)
	v.validateReferences(resource, result)

	return result
}

// ValidateResourceMap validates a resource already parsed as a map.
func (v *Validator) ValidateResourceMap(resource map[string]interface{}, requireID bool) *ValidationResult {
	result := &ValidationResult{Valid: true}

	v.validateResourceType(resource, result)
	if requireID {
		v.validateID(resource, result)
	}
	v.validateStatus(resource, result)
	v.validateReferences(resource, result)

	return result
}

// validateResourceType checks that resourceType is present and recognized.
func (v *Validator) validateResourceType(resource map[string]interface{}, result *ValidationResult) {
	rt, ok := resource["resourceType"]
	if !ok {
		result.Valid = false
		result.Issues = append(result.Issues, OperationOutcomeIssue{
			Severity:    IssueSeverityError,
			Code:        IssueTypeRequired,
			Diagnostics: "resourceType is required",
			Expression:  []string{"resourceType"},
		})
		return
	}

	rtStr, ok := rt.(string)
	if !ok || rtStr == "" {
		result.Valid = false
		result.Issues = append(result.Issues, OperationOutcomeIssue{
			Severity:    IssueSeverityError,
			Code:        IssueTypeValue,
			Diagnostics: "resourceType must be a non-empty string",
			Expression:  []string{"resourceType"},
		})
		return
	}

	if !knownResourceTypes[rtStr] {
		result.Valid = false
		result.Issues = append(result.Issues, OperationOutcomeIssue{
			Severity:    IssueSeverityError,
			Code:        IssueTypeValue,
			Diagnostics: fmt.Sprintf("unknown resourceType: %s", rtStr),
			Expression:  []string{"resourceType"},
		})
	}
}

// validateID checks that id is present when required (for updates).
func (v *Validator) validateID(resource map[string]interface{}, result *ValidationResult) {
	id, ok := resource["id"]
	if !ok {
		result.Valid = false
		result.Issues = append(result.Issues, OperationOutcomeIssue{
			Severity:    IssueSeverityError,
			Code:        IssueTypeRequired,
			Diagnostics: "id is required for update operations",
			Expression:  []string{"id"},
		})
		return
	}
	idStr, ok := id.(string)
	if !ok || idStr == "" {
		result.Valid = false
		result.Issues = append(result.Issues, OperationOutcomeIssue{
			Severity:    IssueSeverityError,
			Code:        IssueTypeValue,
			Diagnostics: "id must be a non-empty string",
			Expression:  []string{"id"},
		})
	}
}

// validateStatus checks that status values match the valid set for the resource type.
func (v *Validator) validateStatus(resource map[string]interface{}, result *ValidationResult) {
	status, ok := resource["status"]
	if !ok {
		return // status is not always required at the generic validation level
	}

	statusStr, ok := status.(string)
	if !ok {
		result.Valid = false
		result.Issues = append(result.Issues, OperationOutcomeIssue{
			Severity:    IssueSeverityError,
			Code:        IssueTypeValue,
			Diagnostics: "status must be a string",
			Expression:  []string{"status"},
		})
		return
	}

	rt, _ := resource["resourceType"].(string)
	validStatuses, hasStatuses := statusValues[rt]
	if !hasStatuses {
		return // no status validation rules for this resource type
	}

	found := false
	for _, vs := range validStatuses {
		if vs == statusStr {
			found = true
			break
		}
	}

	if !found {
		result.Valid = false
		result.Issues = append(result.Issues, OperationOutcomeIssue{
			Severity:    IssueSeverityError,
			Code:        IssueTypeCodeInvalid,
			Diagnostics: fmt.Sprintf("invalid status '%s' for %s; valid values: %s", statusStr, rt, strings.Join(validStatuses, ", ")),
			Expression:  []string{"status"},
		})
	}
}

// validateReferences finds reference fields and validates their format.
func (v *Validator) validateReferences(resource map[string]interface{}, result *ValidationResult) {
	v.walkReferences(resource, "", result)
}

// walkReferences recursively walks through a resource to find and validate reference fields.
func (v *Validator) walkReferences(obj map[string]interface{}, path string, result *ValidationResult) {
	for key, val := range obj {
		currentPath := key
		if path != "" {
			currentPath = path + "." + key
		}

		switch typedVal := val.(type) {
		case map[string]interface{}:
			// Check if this looks like a FHIR Reference (has "reference" field).
			if ref, ok := typedVal["reference"]; ok {
				refStr, isStr := ref.(string)
				if isStr && refStr != "" {
					if !ValidateReferenceFormat(refStr) {
						result.Valid = false
						result.Issues = append(result.Issues, OperationOutcomeIssue{
							Severity:    IssueSeverityError,
							Code:        IssueTypeValue,
							Diagnostics: fmt.Sprintf("invalid reference format '%s'; expected 'ResourceType/id'", refStr),
							Expression:  []string{currentPath + ".reference"},
						})
					}
				}
			}
			// Recurse into nested objects.
			v.walkReferences(typedVal, currentPath, result)

		case []interface{}:
			for i, item := range typedVal {
				if m, ok := item.(map[string]interface{}); ok {
					itemPath := fmt.Sprintf("%s[%d]", currentPath, i)
					v.walkReferences(m, itemPath, result)
				}
			}
		}
	}
}

// ValidateReferenceFormat validates that a reference string matches "ResourceType/id".
func ValidateReferenceFormat(ref string) bool {
	return referencePattern.MatchString(ref)
}

// ValidateBundleEntry validates a single entry in a transaction/batch bundle.
func (v *Validator) ValidateBundleEntry(entry BundleEntry, index int) []OperationOutcomeIssue {
	var issues []OperationOutcomeIssue

	if entry.Request == nil {
		issues = append(issues, OperationOutcomeIssue{
			Severity:    IssueSeverityError,
			Code:        IssueTypeRequired,
			Diagnostics: fmt.Sprintf("entry[%d].request is required for transaction/batch bundles", index),
			Expression:  []string{fmt.Sprintf("entry[%d].request", index)},
		})
		return issues
	}

	method := strings.ToUpper(entry.Request.Method)
	if method != "GET" && method != "POST" && method != "PUT" && method != "DELETE" {
		issues = append(issues, OperationOutcomeIssue{
			Severity:    IssueSeverityError,
			Code:        IssueTypeValue,
			Diagnostics: fmt.Sprintf("entry[%d].request.method must be GET, POST, PUT, or DELETE; got '%s'", index, entry.Request.Method),
			Expression:  []string{fmt.Sprintf("entry[%d].request.method", index)},
		})
	}

	if entry.Request.URL == "" {
		issues = append(issues, OperationOutcomeIssue{
			Severity:    IssueSeverityError,
			Code:        IssueTypeRequired,
			Diagnostics: fmt.Sprintf("entry[%d].request.url is required", index),
			Expression:  []string{fmt.Sprintf("entry[%d].request.url", index)},
		})
	}

	// Resource is required for POST and PUT
	if (method == "POST" || method == "PUT") && len(entry.Resource) == 0 {
		issues = append(issues, OperationOutcomeIssue{
			Severity:    IssueSeverityError,
			Code:        IssueTypeRequired,
			Diagnostics: fmt.Sprintf("entry[%d].resource is required for %s requests", index, method),
			Expression:  []string{fmt.Sprintf("entry[%d].resource", index)},
		})
	}

	// Validate the resource itself for POST/PUT
	if (method == "POST" || method == "PUT") && len(entry.Resource) > 0 {
		requireID := method == "PUT"
		vResult := v.ValidateResource(entry.Resource, requireID)
		issues = append(issues, vResult.Issues...)
	}

	return issues
}

// ValidateBundle validates an entire transaction or batch bundle.
func (v *Validator) ValidateBundle(bundle *Bundle) *ValidationResult {
	result := &ValidationResult{Valid: true}

	if bundle.Type != "transaction" && bundle.Type != "batch" {
		result.Valid = false
		result.Issues = append(result.Issues, OperationOutcomeIssue{
			Severity:    IssueSeverityError,
			Code:        IssueTypeValue,
			Diagnostics: fmt.Sprintf("bundle type must be 'transaction' or 'batch' for processing; got '%s'", bundle.Type),
			Expression:  []string{"type"},
		})
		return result
	}

	if len(bundle.Entry) == 0 {
		result.Valid = false
		result.Issues = append(result.Issues, OperationOutcomeIssue{
			Severity:    IssueSeverityError,
			Code:        IssueTypeRequired,
			Diagnostics: "bundle must contain at least one entry",
			Expression:  []string{"entry"},
		})
		return result
	}

	for i, entry := range bundle.Entry {
		entryIssues := v.ValidateBundleEntry(entry, i)
		if len(entryIssues) > 0 {
			result.Valid = false
			result.Issues = append(result.Issues, entryIssues...)
		}
	}

	return result
}

// IsKnownResourceType returns true if the resource type is recognized.
func IsKnownResourceType(rt string) bool {
	return knownResourceTypes[rt]
}

// ValidStatusValues returns the valid status values for a given resource type.
// Returns nil if no status validation rules exist for the type.
func ValidStatusValues(resourceType string) []string {
	return statusValues[resourceType]
}
