package fhir

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/labstack/echo/v4"
)

// ValidationSeverity represents the severity of a validation issue.
type ValidationSeverity string

const (
	SeverityError       ValidationSeverity = "error"
	SeverityWarning     ValidationSeverity = "warning"
	SeverityInformation ValidationSeverity = "information"
	SeverityFatal       ValidationSeverity = "fatal"
)

// ValidationIssueType represents the type of validation issue.
type ValidationIssueType string

const (
	VIssueTypeStructure    ValidationIssueType = "structure"
	VIssueTypeRequired     ValidationIssueType = "required"
	VIssueTypeValue        ValidationIssueType = "value"
	VIssueTypeInvariant    ValidationIssueType = "invariant"
	VIssueTypeBusinessRule ValidationIssueType = "business-rule"
	VIssueTypeNotFound     ValidationIssueType = "not-found"
)

// ValidationIssue represents a single validation problem.
type ValidationIssue struct {
	Severity    ValidationSeverity  `json:"severity"`
	Code        ValidationIssueType `json:"code"`
	Location    string              `json:"location,omitempty"`
	Diagnostics string              `json:"diagnostics"`
}

// ValidateOpResult holds the complete validation output for the $validate operation.
type ValidateOpResult struct {
	Valid  bool              `json:"valid"`
	Issues []ValidationIssue `json:"issues"`
}

// fhirIDPattern matches valid FHIR id values: [A-Za-z0-9\-\.]{1,64}
var fhirIDPattern = regexp.MustCompile(`^[A-Za-z0-9\-.]{1,64}$`)

// fhirDatePattern matches FHIR date/dateTime formats.
var fhirDatePattern = regexp.MustCompile(`^\d{4}(-\d{2}(-\d{2}(T\d{2}:\d{2}(:\d{2}(\.\d+)?)?(Z|[+-]\d{2}:\d{2})?)?)?)?$`)

// fhirReferenceOpPattern matches FHIR references: ResourceType/id or absolute URLs.
var fhirReferenceOpPattern = regexp.MustCompile(`^([A-Z][a-zA-Z]+/[A-Za-z0-9\-.]+|https?://.+)$`)

// requiredFieldsRegistry maps resource types to their required field names.
var requiredFieldsRegistry = map[string][]string{
	"Patient":             {"name"},
	"Observation":         {"status", "code"},
	"Condition":           {"subject"},
	"AllergyIntolerance":  {"patient"},
	"MedicationRequest":   {"status", "intent", "medication", "subject"},
	"Procedure":           {"status", "subject"},
	"Encounter":           {"status", "class"},
	"DiagnosticReport":    {"status", "code"},
	"Immunization":        {"status", "vaccineCode", "patient", "occurrenceDateTime"},
	"CarePlan":            {"status", "intent", "subject"},
	"CareTeam":            {"subject"},
	"Claim":               {"status", "type", "patient", "provider"},
	"Consent":             {"status", "scope", "category"},
	"Composition":         {"status", "type", "date", "author", "title"},
	"ServiceRequest":      {"status", "intent", "subject"},
	"Coverage":            {"status", "beneficiary"},
	"DocumentReference":   {"status", "content"},
	"Goal":                {"lifecycleStatus", "subject"},
	"Task":                {"status", "intent"},
	"Device":              {},
	"Specimen":            {},
	"FamilyMemberHistory": {"status", "patient", "relationship"},
	"RelatedPerson":       {"patient"},
	"Appointment":         {"status"},
	"Schedule":            {"actor"},
	"Slot":                {"status", "schedule", "start", "end"},
	"Organization":        {},
	"Location":            {},
	"Practitioner":        {},
	"PractitionerRole":    {},
}

// validEncounterClasses lists the valid Encounter.class codes per FHIR R4.
var validEncounterClasses = map[string]bool{
	"AMB":   true,
	"EMER":  true,
	"IMP":   true,
	"ACUTE": true,
	"NONAC": true,
	"SS":    true,
	"HH":    true,
	"FLD":   true,
	"VR":    true,
	"OBSENC": true,
	"PRENC": true,
}

// validMedicationRequestIntents lists valid MedicationRequest.intent values.
var validMedicationRequestIntents = map[string]bool{
	"proposal":       true,
	"plan":           true,
	"order":          true,
	"original-order": true,
	"reflex-order":   true,
	"filler-order":   true,
	"instance-order": true,
	"option":         true,
}

// dateTimeFields lists field names that should contain FHIR date/dateTime values.
var dateTimeFields = map[string]bool{
	"date":               true,
	"birthDate":          true,
	"deceasedDateTime":   true,
	"onsetDateTime":      true,
	"abatementDateTime":  true,
	"recordedDate":       true,
	"effectiveDateTime":  true,
	"issued":             true,
	"occurrenceDateTime": true,
	"authoredOn":         true,
	"start":              true,
	"end":                true,
	"created":            true,
	"sent":               true,
	"received":           true,
}

// ResourceValidator validates FHIR resources against structure rules,
// required fields, value sets, and business rules.
type ResourceValidator struct {
	knownTypes     map[string]bool
	requiredFields map[string][]string
	validStatuses  map[string][]string
}

// additionalResourceTypes lists FHIR R4 resource types that are used in the
// required fields registry but may not be present in the base knownResourceTypes
// map (which is maintained in validator.go for the basic validator).
var additionalResourceTypes = []string{
	"Task",
	"FamilyMemberHistory",
	"RelatedPerson",
	"Device",
	"Goal",
	"Immunization",
}

// NewResourceValidator creates a validator with built-in FHIR R4 rules.
func NewResourceValidator() *ResourceValidator {
	// Build a merged known-types map that includes both the base set
	// from validator.go and the additional types needed by this operation.
	merged := make(map[string]bool, len(knownResourceTypes)+len(additionalResourceTypes))
	for k, v := range knownResourceTypes {
		merged[k] = v
	}
	for _, rt := range additionalResourceTypes {
		merged[rt] = true
	}

	return &ResourceValidator{
		knownTypes:     merged,
		requiredFields: requiredFieldsRegistry,
		validStatuses:  statusValues,
	}
}

// Validate checks a resource against all validation rules.
func (v *ResourceValidator) Validate(resource map[string]interface{}) *ValidateOpResult {
	return v.ValidateWithMode(resource, "")
}

// ValidateWithMode supports different validation modes.
// mode can be: "create", "update", "delete" (affects which rules apply).
// An empty string means all rules apply.
func (v *ResourceValidator) ValidateWithMode(resource map[string]interface{}, mode string) *ValidateOpResult {
	result := &ValidateOpResult{Valid: true}

	if resource == nil {
		result.Valid = false
		result.Issues = append(result.Issues, ValidationIssue{
			Severity:    SeverityFatal,
			Code:        VIssueTypeStructure,
			Diagnostics: "Resource is nil",
		})
		return result
	}

	// Step 1: Validate resourceType
	rt := v.validateResourceType(resource, result)

	// Step 2: Validate id format (skip for create mode)
	if mode != "create" {
		v.validateIDFormat(resource, result, mode)
	}

	// Step 3: Validate meta structure
	v.validateMeta(resource, result)

	// Step 4: Validate required fields
	if rt != "" {
		v.validateRequiredFields(resource, rt, result, mode)
	}

	// Step 5: Validate status values
	if rt != "" {
		v.validateStatus(resource, rt, result)
	}

	// Step 6: Validate references
	v.validateReferences(resource, "", result)

	// Step 7: Validate date/dateTime fields
	v.validateDateFields(resource, rt, result)

	// Step 8: Validate field types (booleans, integers)
	v.validateFieldTypes(resource, rt, result)

	// Step 9: Validate business rules
	if rt != "" {
		v.validateBusinessRules(resource, rt, result)
	}

	return result
}

// validateResourceType checks that resourceType is present and known.
// Returns the resourceType string or empty if invalid.
func (v *ResourceValidator) validateResourceType(resource map[string]interface{}, result *ValidateOpResult) string {
	rtVal, ok := resource["resourceType"]
	if !ok {
		result.Valid = false
		result.Issues = append(result.Issues, ValidationIssue{
			Severity:    SeverityFatal,
			Code:        VIssueTypeStructure,
			Location:    "resourceType",
			Diagnostics: "resourceType is required",
		})
		return ""
	}

	rt, ok := rtVal.(string)
	if !ok || rt == "" {
		result.Valid = false
		result.Issues = append(result.Issues, ValidationIssue{
			Severity:    SeverityFatal,
			Code:        VIssueTypeStructure,
			Location:    "resourceType",
			Diagnostics: "resourceType must be a non-empty string",
		})
		return ""
	}

	if !v.knownTypes[rt] {
		result.Valid = false
		result.Issues = append(result.Issues, ValidationIssue{
			Severity:    SeverityError,
			Code:        VIssueTypeStructure,
			Location:    "resourceType",
			Diagnostics: fmt.Sprintf("Unknown resource type '%s'", rt),
		})
		return ""
	}

	return rt
}

// validateIDFormat checks that if an id is present it matches the FHIR id format.
// In update mode, id is required.
func (v *ResourceValidator) validateIDFormat(resource map[string]interface{}, result *ValidateOpResult, mode string) {
	idVal, hasID := resource["id"]

	if mode == "update" && !hasID {
		result.Valid = false
		result.Issues = append(result.Issues, ValidationIssue{
			Severity:    SeverityError,
			Code:        VIssueTypeRequired,
			Location:    "id",
			Diagnostics: "id is required for update operations",
		})
		return
	}

	if !hasID {
		return
	}

	idStr, ok := idVal.(string)
	if !ok {
		result.Valid = false
		result.Issues = append(result.Issues, ValidationIssue{
			Severity:    SeverityError,
			Code:        VIssueTypeValue,
			Location:    "id",
			Diagnostics: "id must be a string",
		})
		return
	}

	if idStr == "" {
		// Empty id is only an error if we are in update mode
		if mode == "update" {
			result.Valid = false
			result.Issues = append(result.Issues, ValidationIssue{
				Severity:    SeverityError,
				Code:        VIssueTypeValue,
				Location:    "id",
				Diagnostics: "id must not be empty for update operations",
			})
		}
		return
	}

	if !fhirIDPattern.MatchString(idStr) {
		result.Valid = false
		result.Issues = append(result.Issues, ValidationIssue{
			Severity:    SeverityError,
			Code:        VIssueTypeValue,
			Location:    "id",
			Diagnostics: fmt.Sprintf("id '%s' does not match FHIR id format (alphanumeric, hyphens, dots, up to 64 chars)", idStr),
		})
	}
}

// validateMeta checks that the meta field, if present, has a valid structure.
func (v *ResourceValidator) validateMeta(resource map[string]interface{}, result *ValidateOpResult) {
	metaVal, ok := resource["meta"]
	if !ok {
		return
	}

	metaMap, ok := metaVal.(map[string]interface{})
	if !ok {
		result.Valid = false
		result.Issues = append(result.Issues, ValidationIssue{
			Severity:    SeverityError,
			Code:        VIssueTypeStructure,
			Location:    "meta",
			Diagnostics: "meta must be an object",
		})
		return
	}

	// Validate versionId if present
	if vid, ok := metaMap["versionId"]; ok {
		if _, ok := vid.(string); !ok {
			result.Valid = false
			result.Issues = append(result.Issues, ValidationIssue{
				Severity:    SeverityError,
				Code:        VIssueTypeValue,
				Location:    "meta.versionId",
				Diagnostics: "meta.versionId must be a string",
			})
		}
	}

	// Validate profile if present
	if profileVal, ok := metaMap["profile"]; ok {
		profiles, ok := profileVal.([]interface{})
		if !ok {
			result.Valid = false
			result.Issues = append(result.Issues, ValidationIssue{
				Severity:    SeverityError,
				Code:        VIssueTypeStructure,
				Location:    "meta.profile",
				Diagnostics: "meta.profile must be an array",
			})
		} else {
			for i, p := range profiles {
				if _, ok := p.(string); !ok {
					result.Valid = false
					result.Issues = append(result.Issues, ValidationIssue{
						Severity:    SeverityError,
						Code:        VIssueTypeValue,
						Location:    fmt.Sprintf("meta.profile[%d]", i),
						Diagnostics: "meta.profile entries must be strings (canonical URLs)",
					})
				}
			}
		}
	}
}

// validateRequiredFields checks that all required fields for the resource type are present.
func (v *ResourceValidator) validateRequiredFields(resource map[string]interface{}, rt string, result *ValidateOpResult, mode string) {
	fields, ok := v.requiredFields[rt]
	if !ok {
		return
	}

	for _, field := range fields {
		// For MedicationRequest, the "medication" field can be satisfied by
		// either medicationCodeableConcept or medicationReference.
		if rt == "MedicationRequest" && field == "medication" {
			_, hasConcept := resource["medicationCodeableConcept"]
			_, hasRef := resource["medicationReference"]
			if !hasConcept && !hasRef {
				result.Valid = false
				result.Issues = append(result.Issues, ValidationIssue{
					Severity:    SeverityError,
					Code:        VIssueTypeRequired,
					Location:    fmt.Sprintf("%s.medication[x]", rt),
					Diagnostics: fmt.Sprintf("MedicationRequest must have either medicationCodeableConcept or medicationReference"),
				})
			}
			continue
		}

		if _, ok := resource[field]; !ok {
			result.Valid = false
			result.Issues = append(result.Issues, ValidationIssue{
				Severity:    SeverityError,
				Code:        VIssueTypeRequired,
				Location:    fmt.Sprintf("%s.%s", rt, field),
				Diagnostics: fmt.Sprintf("Required field '%s' is missing", field),
			})
		}
	}
}

// validateStatus checks that the status value is valid for the resource type.
func (v *ResourceValidator) validateStatus(resource map[string]interface{}, rt string, result *ValidateOpResult) {
	statusVal, ok := resource["status"]
	if !ok {
		return
	}

	statusStr, ok := statusVal.(string)
	if !ok {
		result.Valid = false
		result.Issues = append(result.Issues, ValidationIssue{
			Severity:    SeverityError,
			Code:        VIssueTypeValue,
			Location:    fmt.Sprintf("%s.status", rt),
			Diagnostics: "status must be a string",
		})
		return
	}

	validStatuses, hasStatuses := v.validStatuses[rt]
	if !hasStatuses {
		return
	}

	for _, vs := range validStatuses {
		if vs == statusStr {
			return
		}
	}

	result.Valid = false
	result.Issues = append(result.Issues, ValidationIssue{
		Severity:    SeverityError,
		Code:        VIssueTypeValue,
		Location:    fmt.Sprintf("%s.status", rt),
		Diagnostics: fmt.Sprintf("Invalid status '%s' for %s; valid values: %s", statusStr, rt, strings.Join(validStatuses, ", ")),
	})
}

// validateReferences recursively finds and validates reference fields.
func (v *ResourceValidator) validateReferences(obj map[string]interface{}, path string, result *ValidateOpResult) {
	rt, _ := obj["resourceType"].(string)

	for key, val := range obj {
		currentPath := key
		if path != "" {
			currentPath = path + "." + key
		} else if rt != "" {
			currentPath = rt + "." + key
		}

		switch typedVal := val.(type) {
		case map[string]interface{}:
			// Check if this is a FHIR Reference.
			if ref, ok := typedVal["reference"]; ok {
				refStr, isStr := ref.(string)
				if isStr && refStr != "" {
					if !fhirReferenceOpPattern.MatchString(refStr) {
						result.Issues = append(result.Issues, ValidationIssue{
							Severity:    SeverityWarning,
							Code:        VIssueTypeValue,
							Location:    currentPath + ".reference",
							Diagnostics: fmt.Sprintf("Reference '%s' does not match expected format 'ResourceType/id' or absolute URL", refStr),
						})
					}
				}
			}
			v.validateReferences(typedVal, currentPath, result)

		case []interface{}:
			for i, item := range typedVal {
				if m, ok := item.(map[string]interface{}); ok {
					itemPath := fmt.Sprintf("%s[%d]", currentPath, i)
					v.validateReferences(m, itemPath, result)
				}
			}
		}
	}
}

// validateDateFields checks that date/dateTime fields have valid formats.
func (v *ResourceValidator) validateDateFields(resource map[string]interface{}, rt string, result *ValidateOpResult) {
	for field, val := range resource {
		if !dateTimeFields[field] {
			continue
		}

		dateStr, ok := val.(string)
		if !ok {
			continue
		}

		if !fhirDatePattern.MatchString(dateStr) {
			location := field
			if rt != "" {
				location = rt + "." + field
			}
			result.Valid = false
			result.Issues = append(result.Issues, ValidationIssue{
				Severity:    SeverityError,
				Code:        VIssueTypeValue,
				Location:    location,
				Diagnostics: fmt.Sprintf("Invalid date/dateTime format '%s' for field '%s'", dateStr, field),
			})
		}
	}
}

// validateFieldTypes checks that known typed fields have correct Go types.
func (v *ResourceValidator) validateFieldTypes(resource map[string]interface{}, rt string, result *ValidateOpResult) {
	// Check boolean fields.
	booleanFields := []string{"active", "deceasedBoolean", "multipleBirthBoolean"}
	for _, field := range booleanFields {
		val, ok := resource[field]
		if !ok {
			continue
		}
		if _, isBool := val.(bool); !isBool {
			location := field
			if rt != "" {
				location = rt + "." + field
			}
			result.Valid = false
			result.Issues = append(result.Issues, ValidationIssue{
				Severity:    SeverityError,
				Code:        VIssueTypeValue,
				Location:    location,
				Diagnostics: fmt.Sprintf("Field '%s' must be a boolean", field),
			})
		}
	}
}

// validateBusinessRules applies resource-type-specific clinical rules.
func (v *ResourceValidator) validateBusinessRules(resource map[string]interface{}, rt string, result *ValidateOpResult) {
	switch rt {
	case "Patient":
		v.validatePatientRules(resource, result)
	case "Observation":
		v.validateObservationRules(resource, result)
	case "MedicationRequest":
		v.validateMedicationRequestRules(resource, result)
	case "Encounter":
		v.validateEncounterRules(resource, result)
	}
}

// validatePatientRules checks that Patient has at least one name or identifier.
func (v *ResourceValidator) validatePatientRules(resource map[string]interface{}, result *ValidateOpResult) {
	hasName := false
	hasIdentifier := false

	if nameVal, ok := resource["name"]; ok {
		switch n := nameVal.(type) {
		case []interface{}:
			hasName = len(n) > 0
		case map[string]interface{}:
			hasName = true
		}
	}

	if idVal, ok := resource["identifier"]; ok {
		switch id := idVal.(type) {
		case []interface{}:
			hasIdentifier = len(id) > 0
		case map[string]interface{}:
			hasIdentifier = true
		}
	}

	if !hasName && !hasIdentifier {
		result.Valid = false
		result.Issues = append(result.Issues, ValidationIssue{
			Severity:    SeverityError,
			Code:        VIssueTypeBusinessRule,
			Location:    "Patient",
			Diagnostics: "Patient must have at least one name or identifier",
		})
	}
}

// validateObservationRules checks Observation-specific business rules.
func (v *ResourceValidator) validateObservationRules(resource map[string]interface{}, result *ValidateOpResult) {
	statusVal, ok := resource["status"]
	if !ok {
		return
	}

	statusStr, ok := statusVal.(string)
	if !ok {
		return
	}

	if statusStr == "final" {
		_, hasValue := resource["valueQuantity"]
		_, hasValueCC := resource["valueCodeableConcept"]
		_, hasValueStr := resource["valueString"]
		_, hasValueBool := resource["valueBoolean"]
		_, hasValueInt := resource["valueInteger"]
		_, hasValueRange := resource["valueRange"]
		_, hasValueRatio := resource["valueRatio"]
		_, hasValueTime := resource["valueTime"]
		_, hasValueDateTime := resource["valueDateTime"]
		_, hasValuePeriod := resource["valuePeriod"]
		_, hasComponent := resource["component"]
		_, hasDataAbsent := resource["dataAbsentReason"]

		hasAnyValue := hasValue || hasValueCC || hasValueStr || hasValueBool ||
			hasValueInt || hasValueRange || hasValueRatio || hasValueTime ||
			hasValueDateTime || hasValuePeriod || hasComponent || hasDataAbsent

		if !hasAnyValue {
			result.Issues = append(result.Issues, ValidationIssue{
				Severity:    SeverityWarning,
				Code:        VIssueTypeBusinessRule,
				Location:    "Observation",
				Diagnostics: "Observation with status 'final' should have a value or dataAbsentReason",
			})
		}
	}
}

// validateMedicationRequestRules checks MedicationRequest-specific rules.
func (v *ResourceValidator) validateMedicationRequestRules(resource map[string]interface{}, result *ValidateOpResult) {
	// Validate intent if present.
	if intentVal, ok := resource["intent"]; ok {
		if intentStr, ok := intentVal.(string); ok {
			if !validMedicationRequestIntents[intentStr] {
				result.Valid = false
				result.Issues = append(result.Issues, ValidationIssue{
					Severity:    SeverityError,
					Code:        VIssueTypeValue,
					Location:    "MedicationRequest.intent",
					Diagnostics: fmt.Sprintf("Invalid intent '%s' for MedicationRequest", intentStr),
				})
			}
		}
	}
}

// validateEncounterRules checks Encounter-specific rules.
func (v *ResourceValidator) validateEncounterRules(resource map[string]interface{}, result *ValidateOpResult) {
	classVal, ok := resource["class"]
	if !ok {
		return
	}

	classMap, ok := classVal.(map[string]interface{})
	if !ok {
		return
	}

	codeVal, ok := classMap["code"]
	if !ok {
		return
	}

	codeStr, ok := codeVal.(string)
	if !ok {
		return
	}

	if !validEncounterClasses[codeStr] {
		result.Issues = append(result.Issues, ValidationIssue{
			Severity:    SeverityWarning,
			Code:        VIssueTypeValue,
			Location:    "Encounter.class.code",
			Diagnostics: fmt.Sprintf("Encounter class code '%s' is not a standard v3 ActEncounterCode (AMB, EMER, IMP, etc.)", codeStr),
		})
	}
}

// ValidateHandler provides the $validate HTTP endpoint.
type ValidateHandler struct {
	validator *ResourceValidator
}

// NewValidateHandler creates a new ValidateHandler.
func NewValidateHandler(validator *ResourceValidator) *ValidateHandler {
	return &ValidateHandler{validator: validator}
}

// RegisterRoutes adds $validate routes to the given FHIR group.
func (h *ValidateHandler) RegisterRoutes(g *echo.Group) {
	g.POST("/$validate", h.Validate)
	g.POST("/:resourceType/$validate", h.Validate)
}

// Validate handles POST /fhir/$validate and POST /fhir/{ResourceType}/$validate.
// It accepts a FHIR resource in the body and returns an OperationOutcome.
func (h *ValidateHandler) Validate(c echo.Context) error {
	// Read the request body.
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, buildValidateOutcome([]ValidationIssue{
			{
				Severity:    SeverityFatal,
				Code:        VIssueTypeStructure,
				Diagnostics: "Failed to read request body",
			},
		}))
	}

	if len(body) == 0 {
		return c.JSON(http.StatusBadRequest, buildValidateOutcome([]ValidationIssue{
			{
				Severity:    SeverityFatal,
				Code:        VIssueTypeStructure,
				Diagnostics: "Request body is empty",
			},
		}))
	}

	// Parse the JSON.
	var resource map[string]interface{}
	if err := json.Unmarshal(body, &resource); err != nil {
		return c.JSON(http.StatusBadRequest, buildValidateOutcome([]ValidationIssue{
			{
				Severity:    SeverityFatal,
				Code:        VIssueTypeStructure,
				Diagnostics: "Invalid JSON: " + err.Error(),
			},
		}))
	}

	// Check for resource type mismatch between URL and body.
	urlType := c.Param("resourceType")
	if urlType != "" {
		bodyType, _ := resource["resourceType"].(string)
		if bodyType != "" && bodyType != urlType {
			return c.JSON(http.StatusBadRequest, buildValidateOutcome([]ValidationIssue{
				{
					Severity:    SeverityError,
					Code:        VIssueTypeStructure,
					Diagnostics: fmt.Sprintf("Resource type in URL '%s' does not match resource type in body '%s'", urlType, bodyType),
				},
			}))
		}
		// If body doesn't have resourceType, set it from URL.
		if bodyType == "" {
			resource["resourceType"] = urlType
		}
	}

	// Check profile parameter (log warning, not yet supported).
	if profile := c.QueryParam("profile"); profile != "" {
		log.Printf("WARN: $validate profile parameter '%s' requested but profile validation is not yet supported", profile)
	}

	// Determine mode.
	mode := c.QueryParam("mode")

	// Run validation.
	vResult := h.validator.ValidateWithMode(resource, mode)

	// Build the OperationOutcome.
	outcome := buildValidateOperationOutcome(vResult)

	return c.JSON(http.StatusOK, outcome)
}

// buildValidateOperationOutcome converts a ValidateOpResult to a FHIR OperationOutcome.
func buildValidateOperationOutcome(result *ValidateOpResult) map[string]interface{} {
	if len(result.Issues) == 0 {
		return buildValidateOutcome([]ValidationIssue{
			{
				Severity:    SeverityInformation,
				Code:        VIssueTypeInvariant,
				Diagnostics: "Validation successful",
			},
		})
	}

	return buildValidateOutcome(result.Issues)
}

// buildValidateOutcome builds a raw OperationOutcome map from validation issues.
func buildValidateOutcome(issues []ValidationIssue) map[string]interface{} {
	issueList := make([]map[string]interface{}, 0, len(issues))
	for _, issue := range issues {
		entry := map[string]interface{}{
			"severity":    string(issue.Severity),
			"code":        string(issue.Code),
			"diagnostics": issue.Diagnostics,
		}
		if issue.Location != "" {
			entry["location"] = []string{issue.Location}
		}
		issueList = append(issueList, entry)
	}

	return map[string]interface{}{
		"resourceType": "OperationOutcome",
		"issue":        issueList,
	}
}
