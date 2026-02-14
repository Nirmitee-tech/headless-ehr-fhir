package fhir

import (
	"encoding/json"
	"testing"
)

func TestNewOperationOutcome(t *testing.T) {
	oo := NewOperationOutcome("error", "processing", "something went wrong")

	if oo.ResourceType != "OperationOutcome" {
		t.Errorf("expected resourceType OperationOutcome, got %s", oo.ResourceType)
	}
	if len(oo.Issue) != 1 {
		t.Fatalf("expected 1 issue, got %d", len(oo.Issue))
	}
	if oo.Issue[0].Severity != "error" {
		t.Errorf("expected severity error, got %s", oo.Issue[0].Severity)
	}
	if oo.Issue[0].Code != "processing" {
		t.Errorf("expected code processing, got %s", oo.Issue[0].Code)
	}
	if oo.Issue[0].Diagnostics != "something went wrong" {
		t.Errorf("expected diagnostics 'something went wrong', got %s", oo.Issue[0].Diagnostics)
	}
}

func TestErrorOutcome(t *testing.T) {
	oo := ErrorOutcome("test error")
	if oo.Issue[0].Severity != "error" {
		t.Error("expected error severity")
	}
	if oo.Issue[0].Diagnostics != "test error" {
		t.Errorf("expected diagnostics 'test error', got %s", oo.Issue[0].Diagnostics)
	}
}

func TestNotFoundOutcome(t *testing.T) {
	oo := NotFoundOutcome("Patient", "123")
	if oo.Issue[0].Code != "not-found" {
		t.Error("expected not-found code")
	}
	if oo.Issue[0].Diagnostics != "Patient/123 not found" {
		t.Errorf("unexpected diagnostics: %s", oo.Issue[0].Diagnostics)
	}
}

func TestFormatReference(t *testing.T) {
	ref := FormatReference("Patient", "abc-123")
	if ref != "Patient/abc-123" {
		t.Errorf("expected Patient/abc-123, got %s", ref)
	}
}

func TestValidationOutcome(t *testing.T) {
	oo := ValidationOutcome("name", "must not be empty")

	if oo.ResourceType != "OperationOutcome" {
		t.Errorf("expected OperationOutcome, got %s", oo.ResourceType)
	}
	if len(oo.Issue) != 1 {
		t.Fatalf("expected 1 issue, got %d", len(oo.Issue))
	}
	if oo.Issue[0].Severity != IssueSeverityError {
		t.Errorf("expected error severity, got %s", oo.Issue[0].Severity)
	}
	if oo.Issue[0].Code != IssueTypeInvalid {
		t.Errorf("expected invalid code, got %s", oo.Issue[0].Code)
	}
	if oo.Issue[0].Diagnostics != "name: must not be empty" {
		t.Errorf("unexpected diagnostics: %s", oo.Issue[0].Diagnostics)
	}
	if len(oo.Issue[0].Expression) != 1 || oo.Issue[0].Expression[0] != "name" {
		t.Errorf("expected expression ['name'], got %v", oo.Issue[0].Expression)
	}
}

func TestRequiredFieldOutcome(t *testing.T) {
	oo := RequiredFieldOutcome("resourceType")

	if oo.Issue[0].Code != IssueTypeRequired {
		t.Errorf("expected required code, got %s", oo.Issue[0].Code)
	}
	if oo.Issue[0].Diagnostics != "resourceType is required" {
		t.Errorf("unexpected diagnostics: %s", oo.Issue[0].Diagnostics)
	}
}

func TestConflictOutcome(t *testing.T) {
	oo := ConflictOutcome("version conflict")

	if oo.Issue[0].Code != IssueTypeConflict {
		t.Errorf("expected conflict code, got %s", oo.Issue[0].Code)
	}
	if oo.Issue[0].Severity != IssueSeverityError {
		t.Errorf("expected error severity, got %s", oo.Issue[0].Severity)
	}
}

func TestNotSupportedOutcome(t *testing.T) {
	oo := NotSupportedOutcome("operation not supported")

	if oo.Issue[0].Code != IssueTypeNotSupported {
		t.Errorf("expected not-supported code, got %s", oo.Issue[0].Code)
	}
}

func TestInternalErrorOutcome(t *testing.T) {
	oo := InternalErrorOutcome("database error")

	if oo.Issue[0].Code != IssueTypeException {
		t.Errorf("expected exception code, got %s", oo.Issue[0].Code)
	}
	if oo.Issue[0].Severity != IssueSeverityFatal {
		t.Errorf("expected fatal severity, got %s", oo.Issue[0].Severity)
	}
}

func TestOutcomeBuilder(t *testing.T) {
	outcome := NewOutcomeBuilder().
		AddIssue(IssueSeverityError, IssueTypeRequired, "field A is required").
		AddIssue(IssueSeverityWarning, IssueTypeValue, "field B has unusual value").
		Build()

	if outcome.ResourceType != "OperationOutcome" {
		t.Errorf("expected OperationOutcome, got %s", outcome.ResourceType)
	}
	if len(outcome.Issue) != 2 {
		t.Fatalf("expected 2 issues, got %d", len(outcome.Issue))
	}
	if outcome.Issue[0].Severity != IssueSeverityError {
		t.Errorf("expected first issue severity error, got %s", outcome.Issue[0].Severity)
	}
	if outcome.Issue[1].Severity != IssueSeverityWarning {
		t.Errorf("expected second issue severity warning, got %s", outcome.Issue[1].Severity)
	}
}

func TestOutcomeBuilder_AddIssueWithDetails(t *testing.T) {
	details := &CodeableConcept{
		Text: "Validation failed",
		Coding: []Coding{
			{System: "http://example.com", Code: "val-001"},
		},
	}

	outcome := NewOutcomeBuilder().
		AddIssueWithDetails(IssueSeverityError, IssueTypeInvalid, "invalid value", details).
		Build()

	if len(outcome.Issue) != 1 {
		t.Fatalf("expected 1 issue, got %d", len(outcome.Issue))
	}
	if outcome.Issue[0].Details == nil {
		t.Fatal("expected details to be set")
	}
	if outcome.Issue[0].Details.Text != "Validation failed" {
		t.Errorf("expected details text 'Validation failed', got '%s'", outcome.Issue[0].Details.Text)
	}
}

func TestOutcomeBuilder_AddIssueWithLocation(t *testing.T) {
	outcome := NewOutcomeBuilder().
		AddIssueWithLocation(IssueSeverityError, IssueTypeRequired, "name is required", "Patient.name").
		Build()

	if len(outcome.Issue) != 1 {
		t.Fatalf("expected 1 issue, got %d", len(outcome.Issue))
	}
	if len(outcome.Issue[0].Expression) != 1 {
		t.Fatalf("expected 1 expression, got %d", len(outcome.Issue[0].Expression))
	}
	if outcome.Issue[0].Expression[0] != "Patient.name" {
		t.Errorf("expected expression 'Patient.name', got '%s'", outcome.Issue[0].Expression[0])
	}
}

func TestOperationOutcome_HasErrors(t *testing.T) {
	tests := []struct {
		name     string
		outcome  *OperationOutcome
		expected bool
	}{
		{
			name: "with error",
			outcome: NewOutcomeBuilder().
				AddIssue(IssueSeverityError, IssueTypeProcessing, "fail").
				Build(),
			expected: true,
		},
		{
			name: "with fatal",
			outcome: NewOutcomeBuilder().
				AddIssue(IssueSeverityFatal, IssueTypeException, "crash").
				Build(),
			expected: true,
		},
		{
			name: "warning only",
			outcome: NewOutcomeBuilder().
				AddIssue(IssueSeverityWarning, IssueTypeValue, "odd value").
				Build(),
			expected: false,
		},
		{
			name: "information only",
			outcome: NewOutcomeBuilder().
				AddIssue(IssueSeverityInformation, IssueTypeProcessing, "fyi").
				Build(),
			expected: false,
		},
		{
			name: "mixed with error",
			outcome: NewOutcomeBuilder().
				AddIssue(IssueSeverityWarning, IssueTypeValue, "odd").
				AddIssue(IssueSeverityError, IssueTypeRequired, "missing").
				Build(),
			expected: true,
		},
		{
			name:     "empty",
			outcome:  NewOutcomeBuilder().Build(),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.outcome.HasErrors(); got != tt.expected {
				t.Errorf("HasErrors() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestMultipleIssuesOutcome(t *testing.T) {
	issues := []OperationOutcomeIssue{
		{Severity: IssueSeverityError, Code: IssueTypeRequired, Diagnostics: "field A"},
		{Severity: IssueSeverityError, Code: IssueTypeValue, Diagnostics: "field B"},
		{Severity: IssueSeverityWarning, Code: IssueTypeInvalid, Diagnostics: "field C"},
	}

	oo := MultipleIssuesOutcome(issues)

	if oo.ResourceType != "OperationOutcome" {
		t.Errorf("expected OperationOutcome, got %s", oo.ResourceType)
	}
	if len(oo.Issue) != 3 {
		t.Fatalf("expected 3 issues, got %d", len(oo.Issue))
	}
}

func TestIsValidSeverity(t *testing.T) {
	valid := []string{"fatal", "error", "warning", "information"}
	for _, s := range valid {
		if !IsValidSeverity(s) {
			t.Errorf("expected %q to be valid severity", s)
		}
	}
	if IsValidSeverity("critical") {
		t.Error("expected 'critical' to be invalid severity")
	}
	if IsValidSeverity("") {
		t.Error("expected empty string to be invalid severity")
	}
}

func TestIsValidIssueType(t *testing.T) {
	valid := []string{"invalid", "structure", "required", "value", "not-found", "conflict", "processing"}
	for _, c := range valid {
		if !IsValidIssueType(c) {
			t.Errorf("expected %q to be valid issue type", c)
		}
	}
	if IsValidIssueType("custom-error") {
		t.Error("expected 'custom-error' to be invalid issue type")
	}
}

func TestOperationOutcome_JSON(t *testing.T) {
	oo := NewOutcomeBuilder().
		AddIssueWithLocation(IssueSeverityError, IssueTypeRequired, "name is required", "Patient.name").
		Build()

	data, err := json.Marshal(oo)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if parsed["resourceType"] != "OperationOutcome" {
		t.Error("expected resourceType OperationOutcome in JSON")
	}

	issues, ok := parsed["issue"].([]interface{})
	if !ok || len(issues) != 1 {
		t.Fatal("expected 1 issue in JSON")
	}

	issue := issues[0].(map[string]interface{})
	if issue["severity"] != "error" {
		t.Errorf("expected severity 'error' in JSON, got %v", issue["severity"])
	}
	if issue["code"] != "required" {
		t.Errorf("expected code 'required' in JSON, got %v", issue["code"])
	}

	expressions, ok := issue["expression"].([]interface{})
	if !ok || len(expressions) != 1 {
		t.Fatal("expected 1 expression in JSON")
	}
	if expressions[0] != "Patient.name" {
		t.Errorf("expected expression 'Patient.name', got %v", expressions[0])
	}
}

func TestSeverityConstants(t *testing.T) {
	if IssueSeverityFatal != "fatal" {
		t.Errorf("expected 'fatal', got %s", IssueSeverityFatal)
	}
	if IssueSeverityError != "error" {
		t.Errorf("expected 'error', got %s", IssueSeverityError)
	}
	if IssueSeverityWarning != "warning" {
		t.Errorf("expected 'warning', got %s", IssueSeverityWarning)
	}
	if IssueSeverityInformation != "information" {
		t.Errorf("expected 'information', got %s", IssueSeverityInformation)
	}
}

func TestIssueTypeConstants(t *testing.T) {
	types := map[string]string{
		"invalid":       IssueTypeInvalid,
		"structure":     IssueTypeStructure,
		"required":      IssueTypeRequired,
		"value":         IssueTypeValue,
		"not-found":     IssueTypeNotFound,
		"conflict":      IssueTypeConflict,
		"processing":    IssueTypeProcessing,
		"security":      IssueTypeSecurity,
		"not-supported": IssueTypeNotSupported,
		"business-rule": IssueTypeBusinessRule,
		"exception":     IssueTypeException,
		"timeout":       IssueTypeTimeout,
		"duplicate":     IssueTypeDuplicate,
		"deleted":       IssueTypeDeleted,
		"code-invalid":  IssueTypeCodeInvalid,
	}

	for expected, constant := range types {
		if constant != expected {
			t.Errorf("expected %q, got %q", expected, constant)
		}
	}
}
