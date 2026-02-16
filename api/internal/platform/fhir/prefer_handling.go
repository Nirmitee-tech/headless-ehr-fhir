package fhir

import (
	"fmt"
	"strings"

	"github.com/labstack/echo/v4"
)

// HandlingPreference represents the FHIR Prefer handling directive value.
// When handling=strict, the server must reject resources with unrecognized elements.
// When handling=lenient (default), unrecognized elements are silently ignored.
type HandlingPreference string

const (
	HandlingStrict  HandlingPreference = "strict"
	HandlingLenient HandlingPreference = "lenient"
)

// contextKeyHandling is the echo.Context key for storing the handling preference.
const contextKeyHandling = "fhir.handling"

// PreferReturnPreference represents the FHIR Prefer return directive value.
type PreferReturnPreference string

const (
	ReturnMinimal          PreferReturnPreference = "minimal"
	ReturnRepresentation   PreferReturnPreference = "representation"
	ReturnOperationOutcome PreferReturnPreference = "OperationOutcome"
)

// PreferDirective holds all parsed directives from a single Prefer header value.
type PreferDirective struct {
	Return       PreferReturnPreference
	Handling     HandlingPreference
	RespondAsync bool
}

// ParsePreferHandling extracts the handling preference from a Prefer header value.
// It supports directives separated by semicolons or commas.
// Returns HandlingLenient if no valid handling directive is found.
func ParsePreferHandling(prefer string) HandlingPreference {
	prefer = strings.TrimSpace(prefer)
	if prefer == "" {
		return HandlingLenient
	}

	for _, sep := range []string{",", ";"} {
		for _, part := range strings.Split(prefer, sep) {
			part = strings.TrimSpace(part)
			if strings.HasPrefix(part, "handling=") {
				val := strings.TrimSpace(part[len("handling="):])
				switch HandlingPreference(val) {
				case HandlingStrict:
					return HandlingStrict
				case HandlingLenient:
					return HandlingLenient
				}
			}
		}
	}

	return HandlingLenient
}

// ParsePreferReturnExtended extracts the return preference from a Prefer header value.
// This provides the same logic as the unexported parsePreferReturn in prefer_middleware.go
// but returns the typed PreferReturnPreference.
func ParsePreferReturnExtended(prefer string) PreferReturnPreference {
	raw := parsePreferReturn(prefer)
	switch PreferReturnPreference(raw) {
	case ReturnMinimal:
		return ReturnMinimal
	case ReturnRepresentation:
		return ReturnRepresentation
	case ReturnOperationOutcome:
		return ReturnOperationOutcome
	default:
		return ""
	}
}

// PreferRespondAsync checks whether the Prefer header contains the respond-async directive.
func PreferRespondAsync(prefer string) bool {
	prefer = strings.TrimSpace(prefer)
	if prefer == "" {
		return false
	}

	for _, sep := range []string{",", ";"} {
		for _, part := range strings.Split(prefer, sep) {
			part = strings.TrimSpace(part)
			if strings.EqualFold(part, "respond-async") {
				return true
			}
		}
	}

	return false
}

// ParsePreferHeader comprehensively parses all directives from a Prefer header value.
// It extracts return, handling, and respond-async directives in a single pass.
func ParsePreferHeader(prefer string) PreferDirective {
	d := PreferDirective{
		Handling: HandlingLenient,
	}

	prefer = strings.TrimSpace(prefer)
	if prefer == "" {
		return d
	}

	// Normalize: replace commas with semicolons so we only split once.
	normalized := strings.ReplaceAll(prefer, ",", ";")
	parts := strings.Split(normalized, ";")

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		if strings.EqualFold(part, "respond-async") {
			d.RespondAsync = true
			continue
		}

		if strings.HasPrefix(part, "return=") {
			val := strings.TrimSpace(part[len("return="):])
			switch PreferReturnPreference(val) {
			case ReturnMinimal, ReturnRepresentation, ReturnOperationOutcome:
				d.Return = PreferReturnPreference(val)
			}
			continue
		}

		if strings.HasPrefix(part, "handling=") {
			val := strings.TrimSpace(part[len("handling="):])
			switch HandlingPreference(val) {
			case HandlingStrict:
				d.Handling = HandlingStrict
			case HandlingLenient:
				d.Handling = HandlingLenient
			}
			continue
		}
	}

	return d
}

// PreferHandlingMiddleware returns Echo middleware that parses the Prefer handling directive
// and stores it in the request context. It also sets the X-FHIR-Handling response header
// to indicate the applied handling mode.
func PreferHandlingMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			prefer := c.Request().Header.Get("Prefer")
			handling := ParsePreferHandling(prefer)
			c.Set(contextKeyHandling, handling)
			c.Response().Header().Set("X-FHIR-Handling", string(handling))
			return next(c)
		}
	}
}

// GetHandlingPreference retrieves the handling preference from the echo.Context.
// Returns HandlingLenient if no preference has been set.
func GetHandlingPreference(c echo.Context) HandlingPreference {
	val := c.Get(contextKeyHandling)
	if val == nil {
		return HandlingLenient
	}
	if h, ok := val.(HandlingPreference); ok {
		return h
	}
	return HandlingLenient
}

// ValidateUnknownElements checks a resource (as a generic map) against a set of
// known element names. It returns the names of any elements not present in knownElements.
// The special keys "resourceType", "id", and "meta" are always considered known.
func ValidateUnknownElements(resource map[string]interface{}, knownElements map[string]bool) []string {
	if len(resource) == 0 {
		return nil
	}

	// These elements are part of every FHIR resource and are always valid.
	always := map[string]bool{
		"resourceType": true,
		"id":           true,
		"meta":         true,
	}

	var unknown []string
	for key := range resource {
		if always[key] {
			continue
		}
		if !knownElements[key] {
			unknown = append(unknown, key)
		}
	}

	return unknown
}

// StrictModeResponse creates an OperationOutcome map suitable for JSON serialization
// when strict handling mode encounters unknown elements in a resource.
func StrictModeResponse(unknownElements []string) map[string]interface{} {
	issues := make([]interface{}, 0, len(unknownElements))
	for _, elem := range unknownElements {
		issues = append(issues, map[string]interface{}{
			"severity":    IssueSeverityError,
			"code":        IssueTypeStructure,
			"diagnostics": fmt.Sprintf("Unknown element '%s' found in resource", elem),
			"expression":  []string{elem},
		})
	}

	return map[string]interface{}{
		"resourceType": "OperationOutcome",
		"issue":        issues,
	}
}

// DefaultKnownElements returns a mapping of FHIR resource type names to their known
// element names. This covers the most commonly used FHIR R4 resources. The returned
// sets do not include the universal elements (resourceType, id, meta) which are
// handled separately by ValidateUnknownElements.
func DefaultKnownElements() map[string]map[string]bool {
	return map[string]map[string]bool{
		"Patient": {
			"identifier":            true,
			"active":                true,
			"name":                  true,
			"telecom":               true,
			"gender":                true,
			"birthDate":             true,
			"deceasedBoolean":       true,
			"deceasedDateTime":      true,
			"address":               true,
			"maritalStatus":         true,
			"multipleBirthBoolean":  true,
			"multipleBirthInteger":  true,
			"photo":                 true,
			"contact":               true,
			"communication":         true,
			"generalPractitioner":   true,
			"managingOrganization":  true,
			"link":                  true,
			"text":                  true,
			"contained":             true,
			"extension":             true,
			"modifierExtension":     true,
			"implicitRules":         true,
			"language":              true,
		},
		"Observation": {
			"identifier":         true,
			"basedOn":            true,
			"partOf":             true,
			"status":             true,
			"category":           true,
			"code":               true,
			"subject":            true,
			"focus":              true,
			"encounter":          true,
			"effectiveDateTime":  true,
			"effectivePeriod":    true,
			"effectiveTiming":    true,
			"effectiveInstant":   true,
			"issued":             true,
			"performer":          true,
			"valueQuantity":      true,
			"valueCodeableConcept": true,
			"valueString":        true,
			"valueBoolean":       true,
			"valueInteger":       true,
			"valueRange":         true,
			"valueRatio":         true,
			"valueSampledData":   true,
			"valueTime":          true,
			"valueDateTime":      true,
			"valuePeriod":        true,
			"dataAbsentReason":   true,
			"interpretation":     true,
			"note":               true,
			"bodySite":           true,
			"method":             true,
			"specimen":           true,
			"device":             true,
			"referenceRange":     true,
			"hasMember":          true,
			"derivedFrom":        true,
			"component":          true,
			"text":               true,
			"contained":          true,
			"extension":          true,
			"modifierExtension":  true,
			"implicitRules":      true,
			"language":           true,
		},
		"Encounter": {
			"identifier":       true,
			"status":           true,
			"statusHistory":    true,
			"class":            true,
			"classHistory":     true,
			"type":             true,
			"serviceType":      true,
			"priority":         true,
			"subject":          true,
			"episodeOfCare":    true,
			"basedOn":          true,
			"participant":      true,
			"appointment":      true,
			"period":           true,
			"length":           true,
			"reasonCode":       true,
			"reasonReference":  true,
			"diagnosis":        true,
			"account":          true,
			"hospitalization":  true,
			"location":         true,
			"serviceProvider":  true,
			"partOf":           true,
			"text":             true,
			"contained":        true,
			"extension":        true,
			"modifierExtension": true,
			"implicitRules":    true,
			"language":         true,
		},
		"Condition": {
			"identifier":          true,
			"clinicalStatus":      true,
			"verificationStatus":  true,
			"category":            true,
			"severity":            true,
			"code":                true,
			"bodySite":            true,
			"subject":             true,
			"encounter":           true,
			"onsetDateTime":       true,
			"onsetAge":            true,
			"onsetPeriod":         true,
			"onsetRange":          true,
			"onsetString":         true,
			"abatementDateTime":   true,
			"abatementAge":        true,
			"abatementPeriod":     true,
			"abatementRange":      true,
			"abatementString":     true,
			"recordedDate":        true,
			"recorder":            true,
			"asserter":            true,
			"stage":               true,
			"evidence":            true,
			"note":                true,
			"text":                true,
			"contained":           true,
			"extension":           true,
			"modifierExtension":   true,
			"implicitRules":       true,
			"language":            true,
		},
		"Procedure": {
			"identifier":          true,
			"instantiatesCanonical": true,
			"instantiatesUri":     true,
			"basedOn":             true,
			"partOf":              true,
			"status":              true,
			"statusReason":        true,
			"category":            true,
			"code":                true,
			"subject":             true,
			"encounter":           true,
			"performedDateTime":   true,
			"performedPeriod":     true,
			"performedString":     true,
			"performedAge":        true,
			"performedRange":      true,
			"recorder":            true,
			"asserter":            true,
			"performer":           true,
			"location":            true,
			"reasonCode":          true,
			"reasonReference":     true,
			"bodySite":            true,
			"outcome":             true,
			"report":              true,
			"complication":        true,
			"complicationDetail":  true,
			"followUp":            true,
			"note":                true,
			"focalDevice":         true,
			"usedReference":       true,
			"usedCode":            true,
			"text":                true,
			"contained":           true,
			"extension":           true,
			"modifierExtension":   true,
			"implicitRules":       true,
			"language":            true,
		},
		"MedicationRequest": {
			"identifier":                 true,
			"status":                     true,
			"statusReason":               true,
			"intent":                     true,
			"category":                   true,
			"priority":                   true,
			"doNotPerform":               true,
			"reportedBoolean":            true,
			"reportedReference":          true,
			"medicationCodeableConcept":  true,
			"medicationReference":        true,
			"subject":                    true,
			"encounter":                  true,
			"supportingInformation":      true,
			"authoredOn":                 true,
			"requester":                  true,
			"performer":                  true,
			"performerType":              true,
			"recorder":                   true,
			"reasonCode":                 true,
			"reasonReference":            true,
			"instantiatesCanonical":      true,
			"instantiatesUri":            true,
			"basedOn":                    true,
			"groupIdentifier":            true,
			"courseOfTherapyType":         true,
			"insurance":                  true,
			"note":                       true,
			"dosageInstruction":          true,
			"dispenseRequest":            true,
			"substitution":              true,
			"priorPrescription":          true,
			"detectedIssue":             true,
			"eventHistory":              true,
			"text":                       true,
			"contained":                  true,
			"extension":                  true,
			"modifierExtension":          true,
			"implicitRules":              true,
			"language":                   true,
		},
		"DiagnosticReport": {
			"identifier":        true,
			"basedOn":           true,
			"status":            true,
			"category":          true,
			"code":              true,
			"subject":           true,
			"encounter":         true,
			"effectiveDateTime": true,
			"effectivePeriod":   true,
			"issued":            true,
			"performer":         true,
			"resultsInterpreter": true,
			"specimen":          true,
			"result":            true,
			"imagingStudy":      true,
			"media":             true,
			"conclusion":        true,
			"conclusionCode":    true,
			"presentedForm":     true,
			"text":              true,
			"contained":         true,
			"extension":         true,
			"modifierExtension": true,
			"implicitRules":     true,
			"language":          true,
		},
		"AllergyIntolerance": {
			"identifier":        true,
			"clinicalStatus":    true,
			"verificationStatus": true,
			"type":              true,
			"category":          true,
			"criticality":       true,
			"code":              true,
			"patient":           true,
			"encounter":         true,
			"onsetDateTime":     true,
			"onsetAge":          true,
			"onsetPeriod":       true,
			"onsetRange":        true,
			"onsetString":       true,
			"recordedDate":      true,
			"recorder":          true,
			"asserter":          true,
			"lastOccurrence":    true,
			"note":              true,
			"reaction":          true,
			"text":              true,
			"contained":         true,
			"extension":         true,
			"modifierExtension": true,
			"implicitRules":     true,
			"language":          true,
		},
		"Immunization": {
			"identifier":         true,
			"status":             true,
			"statusReason":       true,
			"vaccineCode":        true,
			"patient":            true,
			"encounter":          true,
			"occurrenceDateTime": true,
			"occurrenceString":   true,
			"recorded":           true,
			"primarySource":      true,
			"reportOrigin":       true,
			"location":           true,
			"manufacturer":       true,
			"lotNumber":          true,
			"expirationDate":     true,
			"site":               true,
			"route":              true,
			"doseQuantity":       true,
			"performer":          true,
			"note":               true,
			"reasonCode":         true,
			"reasonReference":    true,
			"isSubpotent":        true,
			"subpotentReason":    true,
			"education":          true,
			"programEligibility": true,
			"fundingSource":      true,
			"reaction":           true,
			"protocolApplied":    true,
			"text":               true,
			"contained":          true,
			"extension":          true,
			"modifierExtension":  true,
			"implicitRules":      true,
			"language":           true,
		},
		"CarePlan": {
			"identifier":         true,
			"instantiatesCanonical": true,
			"instantiatesUri":    true,
			"basedOn":            true,
			"replaces":           true,
			"partOf":             true,
			"status":             true,
			"intent":             true,
			"category":           true,
			"title":              true,
			"description":        true,
			"subject":            true,
			"encounter":          true,
			"period":             true,
			"created":            true,
			"author":             true,
			"contributor":        true,
			"careTeam":           true,
			"addresses":          true,
			"supportingInfo":     true,
			"goal":               true,
			"activity":           true,
			"note":               true,
			"text":               true,
			"contained":          true,
			"extension":          true,
			"modifierExtension":  true,
			"implicitRules":      true,
			"language":           true,
		},
		"Organization": {
			"identifier":        true,
			"active":            true,
			"type":              true,
			"name":              true,
			"alias":             true,
			"telecom":           true,
			"address":           true,
			"partOf":            true,
			"contact":           true,
			"endpoint":          true,
			"text":              true,
			"contained":         true,
			"extension":         true,
			"modifierExtension": true,
			"implicitRules":     true,
			"language":          true,
		},
		"Practitioner": {
			"identifier":        true,
			"active":            true,
			"name":              true,
			"telecom":           true,
			"address":           true,
			"gender":            true,
			"birthDate":         true,
			"photo":             true,
			"qualification":     true,
			"communication":     true,
			"text":              true,
			"contained":         true,
			"extension":         true,
			"modifierExtension": true,
			"implicitRules":     true,
			"language":          true,
		},
	}
}
