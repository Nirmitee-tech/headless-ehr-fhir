package fhir

import "fmt"

// OperationOutcome severity levels per FHIR R4 spec.
const (
	IssueSeverityFatal       = "fatal"
	IssueSeverityError       = "error"
	IssueSeverityWarning     = "warning"
	IssueSeverityInformation = "information"
)

// OperationOutcome issue type codes per FHIR R4 spec.
const (
	IssueTypeInvalid     = "invalid"
	IssueTypeStructure   = "structure"
	IssueTypeRequired    = "required"
	IssueTypeValue       = "value"
	IssueTypeNotFound    = "not-found"
	IssueTypeConflict    = "conflict"
	IssueTypeProcessing  = "processing"
	IssueTypeSecurity    = "security"
	IssueTypeLogin       = "login"
	IssueTypeThrottled   = "throttled"
	IssueTypeNotSupported = "not-supported"
	IssueTypeBusinessRule = "business-rule"
	IssueTypeException   = "exception"
	IssueTypeTimeout     = "timeout"
	IssueTypeDuplicate   = "duplicate"
	IssueTypeDeleted     = "deleted"
	IssueTypeCodeInvalid = "code-invalid"
)

// validSeverities is the set of valid FHIR issue severity values.
var validSeverities = map[string]bool{
	IssueSeverityFatal:       true,
	IssueSeverityError:       true,
	IssueSeverityWarning:     true,
	IssueSeverityInformation: true,
}

// validIssueTypes is the set of valid FHIR issue type codes.
var validIssueTypes = map[string]bool{
	IssueTypeInvalid:      true,
	IssueTypeStructure:    true,
	IssueTypeRequired:     true,
	IssueTypeValue:        true,
	IssueTypeNotFound:     true,
	IssueTypeConflict:     true,
	IssueTypeProcessing:   true,
	IssueTypeSecurity:     true,
	IssueTypeLogin:        true,
	IssueTypeThrottled:    true,
	IssueTypeNotSupported: true,
	IssueTypeBusinessRule: true,
	IssueTypeException:    true,
	IssueTypeTimeout:      true,
	IssueTypeDuplicate:    true,
	IssueTypeDeleted:      true,
	IssueTypeCodeInvalid:  true,
}

// IsValidSeverity checks whether a severity string is a valid FHIR issue severity.
func IsValidSeverity(s string) bool {
	return validSeverities[s]
}

// IsValidIssueType checks whether a code string is a valid FHIR issue type.
func IsValidIssueType(code string) bool {
	return validIssueTypes[code]
}

// OutcomeBuilder provides a fluent API for constructing OperationOutcome resources.
type OutcomeBuilder struct {
	outcome *OperationOutcome
}

// NewOutcomeBuilder creates a new OutcomeBuilder.
func NewOutcomeBuilder() *OutcomeBuilder {
	return &OutcomeBuilder{
		outcome: &OperationOutcome{
			ResourceType: "OperationOutcome",
		},
	}
}

// AddIssue adds a single issue to the OperationOutcome.
func (b *OutcomeBuilder) AddIssue(severity, code, diagnostics string) *OutcomeBuilder {
	b.outcome.Issue = append(b.outcome.Issue, OperationOutcomeIssue{
		Severity:    severity,
		Code:        code,
		Diagnostics: diagnostics,
	})
	return b
}

// AddIssueWithDetails adds an issue with a CodeableConcept details field.
func (b *OutcomeBuilder) AddIssueWithDetails(severity, code, diagnostics string, details *CodeableConcept) *OutcomeBuilder {
	b.outcome.Issue = append(b.outcome.Issue, OperationOutcomeIssue{
		Severity:    severity,
		Code:        code,
		Diagnostics: diagnostics,
		Details:     details,
	})
	return b
}

// AddIssueWithLocation adds an issue including an expression/location path.
func (b *OutcomeBuilder) AddIssueWithLocation(severity, code, diagnostics, location string) *OutcomeBuilder {
	b.outcome.Issue = append(b.outcome.Issue, OperationOutcomeIssue{
		Severity:    severity,
		Code:        code,
		Diagnostics: diagnostics,
		Expression:  []string{location},
	})
	return b
}

// Build returns the constructed OperationOutcome.
func (b *OutcomeBuilder) Build() *OperationOutcome {
	return b.outcome
}

// HasErrors returns true if the outcome contains any error or fatal issues.
func (o *OperationOutcome) HasErrors() bool {
	for _, issue := range o.Issue {
		if issue.Severity == IssueSeverityError || issue.Severity == IssueSeverityFatal {
			return true
		}
	}
	return false
}

// ValidationOutcome creates an OperationOutcome for validation errors.
func ValidationOutcome(field, message string) *OperationOutcome {
	return &OperationOutcome{
		ResourceType: "OperationOutcome",
		Issue: []OperationOutcomeIssue{
			{
				Severity:    IssueSeverityError,
				Code:        IssueTypeInvalid,
				Diagnostics: fmt.Sprintf("%s: %s", field, message),
				Expression:  []string{field},
			},
		},
	}
}

// RequiredFieldOutcome creates an OperationOutcome for a missing required field.
func RequiredFieldOutcome(field string) *OperationOutcome {
	return &OperationOutcome{
		ResourceType: "OperationOutcome",
		Issue: []OperationOutcomeIssue{
			{
				Severity:    IssueSeverityError,
				Code:        IssueTypeRequired,
				Diagnostics: fmt.Sprintf("%s is required", field),
				Expression:  []string{field},
			},
		},
	}
}

// ConflictOutcome creates an OperationOutcome for a conflict error.
func ConflictOutcome(diagnostics string) *OperationOutcome {
	return NewOperationOutcome(IssueSeverityError, IssueTypeConflict, diagnostics)
}

// NotSupportedOutcome creates an OperationOutcome for unsupported operations.
func NotSupportedOutcome(diagnostics string) *OperationOutcome {
	return NewOperationOutcome(IssueSeverityError, IssueTypeNotSupported, diagnostics)
}

// InternalErrorOutcome creates an OperationOutcome for internal server errors.
func InternalErrorOutcome(diagnostics string) *OperationOutcome {
	return NewOperationOutcome(IssueSeverityFatal, IssueTypeException, diagnostics)
}

// MultipleIssuesOutcome creates an OperationOutcome with multiple issues from validation.
func MultipleIssuesOutcome(issues []OperationOutcomeIssue) *OperationOutcome {
	return &OperationOutcome{
		ResourceType: "OperationOutcome",
		Issue:        issues,
	}
}
