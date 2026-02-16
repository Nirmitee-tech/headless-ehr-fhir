package fhir

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestSuccessOutcome(t *testing.T) {
	oo := SuccessOutcome("operation completed successfully")

	if oo.ResourceType != "OperationOutcome" {
		t.Errorf("expected resourceType OperationOutcome, got %s", oo.ResourceType)
	}
	if len(oo.Issue) != 1 {
		t.Fatalf("expected 1 issue, got %d", len(oo.Issue))
	}
	if oo.Issue[0].Severity != IssueSeverityInformation {
		t.Errorf("expected severity %s, got %s", IssueSeverityInformation, oo.Issue[0].Severity)
	}
	if oo.Issue[0].Code != IssueTypeProcessing {
		t.Errorf("expected code %s, got %s", IssueTypeProcessing, oo.Issue[0].Code)
	}
	if oo.Issue[0].Diagnostics != "operation completed successfully" {
		t.Errorf("expected diagnostics 'operation completed successfully', got %s", oo.Issue[0].Diagnostics)
	}
}

func TestSuccessOutcome_HasNoErrors(t *testing.T) {
	oo := SuccessOutcome("all good")
	if oo.HasErrors() {
		t.Error("SuccessOutcome should not have errors")
	}
}

func TestSuccessOutcome_JSON(t *testing.T) {
	oo := SuccessOutcome("resource validated")
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
	issues := parsed["issue"].([]interface{})
	issue := issues[0].(map[string]interface{})
	if issue["severity"] != "information" {
		t.Errorf("expected severity 'information' in JSON, got %v", issue["severity"])
	}
}

func TestWarningOutcome(t *testing.T) {
	oo := WarningOutcome("deprecated parameter used")

	if oo.ResourceType != "OperationOutcome" {
		t.Errorf("expected resourceType OperationOutcome, got %s", oo.ResourceType)
	}
	if len(oo.Issue) != 1 {
		t.Fatalf("expected 1 issue, got %d", len(oo.Issue))
	}
	if oo.Issue[0].Severity != IssueSeverityWarning {
		t.Errorf("expected severity %s, got %s", IssueSeverityWarning, oo.Issue[0].Severity)
	}
	if oo.Issue[0].Code != IssueTypeProcessing {
		t.Errorf("expected code %s, got %s", IssueTypeProcessing, oo.Issue[0].Code)
	}
	if oo.Issue[0].Diagnostics != "deprecated parameter used" {
		t.Errorf("expected diagnostics 'deprecated parameter used', got %s", oo.Issue[0].Diagnostics)
	}
}

func TestWarningOutcome_HasNoErrors(t *testing.T) {
	oo := WarningOutcome("something unusual")
	if oo.HasErrors() {
		t.Error("WarningOutcome should not report HasErrors as true")
	}
}

func TestMultiValidationOutcome_Empty(t *testing.T) {
	oo := MultiValidationOutcome(nil)

	if oo.ResourceType != "OperationOutcome" {
		t.Errorf("expected resourceType OperationOutcome, got %s", oo.ResourceType)
	}
	if len(oo.Issue) != 0 {
		t.Errorf("expected 0 issues for nil input, got %d", len(oo.Issue))
	}
}

func TestMultiValidationOutcome_SingleIssue(t *testing.T) {
	issues := []ValidationIssue{
		{
			Severity:    SeverityError,
			Code:        VIssueTypeRequired,
			Diagnostics: "Patient.name is required",
			Location:    "Patient.name",
		},
	}

	oo := MultiValidationOutcome(issues)

	if len(oo.Issue) != 1 {
		t.Fatalf("expected 1 issue, got %d", len(oo.Issue))
	}
	if oo.Issue[0].Severity != IssueSeverityError {
		t.Errorf("expected severity %s, got %s", IssueSeverityError, oo.Issue[0].Severity)
	}
	if oo.Issue[0].Code != "required" {
		t.Errorf("expected code 'required', got %s", oo.Issue[0].Code)
	}
	if oo.Issue[0].Diagnostics != "Patient.name is required" {
		t.Errorf("unexpected diagnostics: %s", oo.Issue[0].Diagnostics)
	}
	if len(oo.Issue[0].Expression) != 1 || oo.Issue[0].Expression[0] != "Patient.name" {
		t.Errorf("expected expression ['Patient.name'], got %v", oo.Issue[0].Expression)
	}
}

func TestMultiValidationOutcome_MultipleIssues(t *testing.T) {
	issues := []ValidationIssue{
		{
			Severity:    SeverityError,
			Code:        VIssueTypeRequired,
			Diagnostics: "name is required",
			Location:    "Patient.name",
		},
		{
			Severity:    SeverityWarning,
			Code:        VIssueTypeValue,
			Diagnostics: "birthDate is in the future",
			Location:    "Patient.birthDate",
		},
		{
			Severity:    SeverityInformation,
			Code:        VIssueTypeStructure,
			Diagnostics: "extension is non-standard",
			Location:    "Patient.extension",
		},
	}

	oo := MultiValidationOutcome(issues)

	if oo.ResourceType != "OperationOutcome" {
		t.Errorf("expected OperationOutcome, got %s", oo.ResourceType)
	}
	if len(oo.Issue) != 3 {
		t.Fatalf("expected 3 issues, got %d", len(oo.Issue))
	}

	// Verify first issue: error
	if oo.Issue[0].Severity != "error" {
		t.Errorf("issue 0: expected severity 'error', got %s", oo.Issue[0].Severity)
	}
	if oo.Issue[0].Code != "required" {
		t.Errorf("issue 0: expected code 'required', got %s", oo.Issue[0].Code)
	}

	// Verify second issue: warning
	if oo.Issue[1].Severity != "warning" {
		t.Errorf("issue 1: expected severity 'warning', got %s", oo.Issue[1].Severity)
	}
	if oo.Issue[1].Code != "value" {
		t.Errorf("issue 1: expected code 'value', got %s", oo.Issue[1].Code)
	}

	// Verify third issue: information
	if oo.Issue[2].Severity != "information" {
		t.Errorf("issue 2: expected severity 'information', got %s", oo.Issue[2].Severity)
	}
}

func TestMultiValidationOutcome_NoLocation(t *testing.T) {
	issues := []ValidationIssue{
		{
			Severity:    SeverityError,
			Code:        VIssueTypeStructure,
			Diagnostics: "invalid JSON structure",
		},
	}

	oo := MultiValidationOutcome(issues)

	if len(oo.Issue) != 1 {
		t.Fatalf("expected 1 issue, got %d", len(oo.Issue))
	}
	if len(oo.Issue[0].Expression) != 0 {
		t.Errorf("expected no expressions when location is empty, got %v", oo.Issue[0].Expression)
	}
}

func TestMultiValidationOutcome_HasErrors(t *testing.T) {
	issues := []ValidationIssue{
		{Severity: SeverityError, Code: VIssueTypeRequired, Diagnostics: "missing field"},
	}
	oo := MultiValidationOutcome(issues)
	if !oo.HasErrors() {
		t.Error("MultiValidationOutcome with error issue should report HasErrors")
	}
}

func TestMultiValidationOutcome_WarningsOnly(t *testing.T) {
	issues := []ValidationIssue{
		{Severity: SeverityWarning, Code: VIssueTypeValue, Diagnostics: "unusual value"},
	}
	oo := MultiValidationOutcome(issues)
	if oo.HasErrors() {
		t.Error("MultiValidationOutcome with only warnings should not report HasErrors")
	}
}

func TestGoneOutcome(t *testing.T) {
	oo := GoneOutcome("Patient", "456")

	if oo.ResourceType != "OperationOutcome" {
		t.Errorf("expected resourceType OperationOutcome, got %s", oo.ResourceType)
	}
	if len(oo.Issue) != 1 {
		t.Fatalf("expected 1 issue, got %d", len(oo.Issue))
	}
	if oo.Issue[0].Severity != IssueSeverityError {
		t.Errorf("expected severity %s, got %s", IssueSeverityError, oo.Issue[0].Severity)
	}
	if oo.Issue[0].Code != IssueTypeDeleted {
		t.Errorf("expected code %s, got %s", IssueTypeDeleted, oo.Issue[0].Code)
	}
	if oo.Issue[0].Diagnostics != "Patient/456 has been deleted" {
		t.Errorf("unexpected diagnostics: %s", oo.Issue[0].Diagnostics)
	}
}

func TestGoneOutcome_DifferentResource(t *testing.T) {
	oo := GoneOutcome("Observation", "obs-789")

	if oo.Issue[0].Diagnostics != "Observation/obs-789 has been deleted" {
		t.Errorf("unexpected diagnostics: %s", oo.Issue[0].Diagnostics)
	}
}

func TestThrottleOutcome(t *testing.T) {
	oo := ThrottleOutcome()

	if oo.ResourceType != "OperationOutcome" {
		t.Errorf("expected resourceType OperationOutcome, got %s", oo.ResourceType)
	}
	if len(oo.Issue) != 1 {
		t.Fatalf("expected 1 issue, got %d", len(oo.Issue))
	}
	if oo.Issue[0].Severity != IssueSeverityError {
		t.Errorf("expected severity %s, got %s", IssueSeverityError, oo.Issue[0].Severity)
	}
	if oo.Issue[0].Code != IssueTypeThrottled {
		t.Errorf("expected code %s, got %s", IssueTypeThrottled, oo.Issue[0].Code)
	}
	if !strings.Contains(oo.Issue[0].Diagnostics, "Rate limit") {
		t.Errorf("expected diagnostics to mention rate limit, got: %s", oo.Issue[0].Diagnostics)
	}
}

func TestThrottleOutcome_HasErrors(t *testing.T) {
	oo := ThrottleOutcome()
	if !oo.HasErrors() {
		t.Error("ThrottleOutcome should report HasErrors")
	}
}

func TestMethodNotAllowedOutcome(t *testing.T) {
	oo := MethodNotAllowedOutcome("DELETE")

	if oo.ResourceType != "OperationOutcome" {
		t.Errorf("expected resourceType OperationOutcome, got %s", oo.ResourceType)
	}
	if len(oo.Issue) != 1 {
		t.Fatalf("expected 1 issue, got %d", len(oo.Issue))
	}
	if oo.Issue[0].Severity != IssueSeverityError {
		t.Errorf("expected severity %s, got %s", IssueSeverityError, oo.Issue[0].Severity)
	}
	if oo.Issue[0].Code != IssueTypeNotSupported {
		t.Errorf("expected code %s, got %s", IssueTypeNotSupported, oo.Issue[0].Code)
	}
	if !strings.Contains(oo.Issue[0].Diagnostics, "DELETE") {
		t.Errorf("expected diagnostics to contain method name, got: %s", oo.Issue[0].Diagnostics)
	}
	if !strings.Contains(oo.Issue[0].Diagnostics, "not allowed") {
		t.Errorf("expected diagnostics to mention 'not allowed', got: %s", oo.Issue[0].Diagnostics)
	}
}

func TestMethodNotAllowedOutcome_DifferentMethods(t *testing.T) {
	methods := []string{"PATCH", "PUT", "POST", "DELETE"}
	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			oo := MethodNotAllowedOutcome(method)
			if !strings.Contains(oo.Issue[0].Diagnostics, method) {
				t.Errorf("expected diagnostics to contain %s, got: %s", method, oo.Issue[0].Diagnostics)
			}
		})
	}
}

func TestMethodNotAllowedOutcome_HasErrors(t *testing.T) {
	oo := MethodNotAllowedOutcome("PATCH")
	if !oo.HasErrors() {
		t.Error("MethodNotAllowedOutcome should report HasErrors")
	}
}

func TestGoneOutcome_JSON(t *testing.T) {
	oo := GoneOutcome("MedicationRequest", "med-001")
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

	issues := parsed["issue"].([]interface{})
	if len(issues) != 1 {
		t.Fatalf("expected 1 issue in JSON, got %d", len(issues))
	}

	issue := issues[0].(map[string]interface{})
	if issue["severity"] != "error" {
		t.Errorf("expected severity 'error' in JSON, got %v", issue["severity"])
	}
	if issue["code"] != "deleted" {
		t.Errorf("expected code 'deleted' in JSON, got %v", issue["code"])
	}
}

func TestMultiValidationOutcome_JSON(t *testing.T) {
	issues := []ValidationIssue{
		{
			Severity:    SeverityError,
			Code:        VIssueTypeRequired,
			Diagnostics: "name is required",
			Location:    "Patient.name",
		},
		{
			Severity:    SeverityWarning,
			Code:        VIssueTypeValue,
			Diagnostics: "unusual birthDate",
			Location:    "Patient.birthDate",
		},
	}

	oo := MultiValidationOutcome(issues)
	data, err := json.Marshal(oo)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	jsonIssues := parsed["issue"].([]interface{})
	if len(jsonIssues) != 2 {
		t.Fatalf("expected 2 issues in JSON, got %d", len(jsonIssues))
	}

	first := jsonIssues[0].(map[string]interface{})
	if first["severity"] != "error" {
		t.Errorf("issue 0: expected severity 'error', got %v", first["severity"])
	}
	exprs := first["expression"].([]interface{})
	if len(exprs) != 1 || exprs[0] != "Patient.name" {
		t.Errorf("issue 0: unexpected expression %v", exprs)
	}

	second := jsonIssues[1].(map[string]interface{})
	if second["severity"] != "warning" {
		t.Errorf("issue 1: expected severity 'warning', got %v", second["severity"])
	}
}
