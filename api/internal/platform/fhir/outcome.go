package fhir

import "fmt"

// SuccessOutcome creates a success OperationOutcome with severity=information.
// It is suitable for returning an affirmative result with an informational message,
// for example after a successful $validate or $process-message operation.
func SuccessOutcome(message string) *OperationOutcome {
	return NewOperationOutcome(IssueSeverityInformation, IssueTypeProcessing, message)
}

// WarningOutcome creates a warning OperationOutcome.
// Use this when an operation succeeded but produced non-fatal warnings that
// the client should be aware of.
func WarningOutcome(message string) *OperationOutcome {
	return NewOperationOutcome(IssueSeverityWarning, IssueTypeProcessing, message)
}

// MultiValidationOutcome creates an OperationOutcome containing multiple
// validation issues. Each ValidationIssue is mapped to an OperationOutcomeIssue
// with the appropriate severity, code, diagnostics, and FHIRPath expression.
//
// This complements the existing ValidationOutcome helper (which handles a single
// field/message pair) by supporting batch validation results such as those
// produced by the $validate operation.
func MultiValidationOutcome(issues []ValidationIssue) *OperationOutcome {
	ooIssues := make([]OperationOutcomeIssue, 0, len(issues))
	for _, vi := range issues {
		issue := OperationOutcomeIssue{
			Severity:    string(vi.Severity),
			Code:        string(vi.Code),
			Diagnostics: vi.Diagnostics,
		}
		if vi.Location != "" {
			issue.Expression = []string{vi.Location}
		}
		ooIssues = append(ooIssues, issue)
	}
	return &OperationOutcome{
		ResourceType: "OperationOutcome",
		Issue:        ooIssues,
	}
}

// GoneOutcome creates a 410-style OperationOutcome for a resource that has been
// deleted. The FHIR spec uses issue type "deleted" for this scenario.
func GoneOutcome(resourceType, id string) *OperationOutcome {
	return NewOperationOutcome(
		IssueSeverityError,
		IssueTypeDeleted,
		fmt.Sprintf("%s/%s has been deleted", resourceType, id),
	)
}

// ThrottleOutcome creates a 429-style OperationOutcome indicating the server is
// rate-limiting the client. The FHIR spec uses issue type "throttled".
func ThrottleOutcome() *OperationOutcome {
	return NewOperationOutcome(
		IssueSeverityError,
		IssueTypeThrottled,
		"Rate limit exceeded. Please retry after a delay.",
	)
}

// MethodNotAllowedOutcome creates a 405-style OperationOutcome for an HTTP
// method that is not permitted on the target resource endpoint.
func MethodNotAllowedOutcome(method string) *OperationOutcome {
	return NewOperationOutcome(
		IssueSeverityError,
		IssueTypeNotSupported,
		fmt.Sprintf("HTTP method %s is not allowed on this resource", method),
	)
}
