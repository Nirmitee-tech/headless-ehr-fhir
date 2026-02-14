package fhir

import (
	"encoding/json"
	"testing"
)

func TestValidateResource_ValidPatient(t *testing.T) {
	v := NewValidator()
	data := json.RawMessage(`{"resourceType": "Patient", "id": "123", "status": "active"}`)
	result := v.ValidateResource(data, true)

	if !result.Valid {
		t.Errorf("expected valid, got invalid with issues: %v", result.Issues)
	}
	if len(result.Issues) != 0 {
		t.Errorf("expected 0 issues, got %d", len(result.Issues))
	}
}

func TestValidateResource_MissingResourceType(t *testing.T) {
	v := NewValidator()
	data := json.RawMessage(`{"id": "123"}`)
	result := v.ValidateResource(data, false)

	if result.Valid {
		t.Error("expected invalid for missing resourceType")
	}
	if len(result.Issues) == 0 {
		t.Fatal("expected at least 1 issue")
	}
	if result.Issues[0].Code != IssueTypeRequired {
		t.Errorf("expected code 'required', got '%s'", result.Issues[0].Code)
	}
}

func TestValidateResource_UnknownResourceType(t *testing.T) {
	v := NewValidator()
	data := json.RawMessage(`{"resourceType": "FakeResource", "id": "123"}`)
	result := v.ValidateResource(data, false)

	if result.Valid {
		t.Error("expected invalid for unknown resourceType")
	}
	found := false
	for _, issue := range result.Issues {
		if issue.Code == IssueTypeValue && len(issue.Expression) > 0 && issue.Expression[0] == "resourceType" {
			found = true
		}
	}
	if !found {
		t.Error("expected a value issue for resourceType")
	}
}

func TestValidateResource_MissingID_RequiredTrue(t *testing.T) {
	v := NewValidator()
	data := json.RawMessage(`{"resourceType": "Patient"}`)
	result := v.ValidateResource(data, true)

	if result.Valid {
		t.Error("expected invalid when id is required but missing")
	}
	found := false
	for _, issue := range result.Issues {
		if issue.Code == IssueTypeRequired && len(issue.Expression) > 0 && issue.Expression[0] == "id" {
			found = true
		}
	}
	if !found {
		t.Error("expected a required issue for id")
	}
}

func TestValidateResource_MissingID_RequiredFalse(t *testing.T) {
	v := NewValidator()
	data := json.RawMessage(`{"resourceType": "Patient"}`)
	result := v.ValidateResource(data, false)

	if !result.Valid {
		t.Error("expected valid when id is not required")
	}
}

func TestValidateResource_EmptyID(t *testing.T) {
	v := NewValidator()
	data := json.RawMessage(`{"resourceType": "Patient", "id": ""}`)
	result := v.ValidateResource(data, true)

	if result.Valid {
		t.Error("expected invalid for empty id")
	}
}

func TestValidateResource_InvalidStatus(t *testing.T) {
	v := NewValidator()
	data := json.RawMessage(`{"resourceType": "Patient", "id": "123", "status": "bogus"}`)
	result := v.ValidateResource(data, false)

	if result.Valid {
		t.Error("expected invalid for bogus status")
	}
	found := false
	for _, issue := range result.Issues {
		if issue.Code == IssueTypeCodeInvalid {
			found = true
		}
	}
	if !found {
		t.Error("expected a code-invalid issue for status")
	}
}

func TestValidateResource_ValidStatus(t *testing.T) {
	tests := []struct {
		resourceType string
		status       string
	}{
		{"Patient", "active"},
		{"Patient", "inactive"},
		{"Encounter", "planned"},
		{"Encounter", "in-progress"},
		{"Encounter", "finished"},
		{"Observation", "final"},
		{"Observation", "amended"},
		{"Condition", "active"},
		{"Condition", "resolved"},
		{"MedicationRequest", "active"},
		{"MedicationRequest", "completed"},
	}

	v := NewValidator()
	for _, tt := range tests {
		t.Run(tt.resourceType+"_"+tt.status, func(t *testing.T) {
			data, _ := json.Marshal(map[string]string{
				"resourceType": tt.resourceType,
				"id":           "test-1",
				"status":       tt.status,
			})
			result := v.ValidateResource(data, false)
			if !result.Valid {
				t.Errorf("expected valid for %s status '%s', got issues: %v", tt.resourceType, tt.status, result.Issues)
			}
		})
	}
}

func TestValidateResource_InvalidJSON(t *testing.T) {
	v := NewValidator()
	data := json.RawMessage(`{not valid json}`)
	result := v.ValidateResource(data, false)

	if result.Valid {
		t.Error("expected invalid for bad JSON")
	}
	if len(result.Issues) == 0 {
		t.Fatal("expected at least 1 issue")
	}
	if result.Issues[0].Code != IssueTypeStructure {
		t.Errorf("expected code 'structure', got '%s'", result.Issues[0].Code)
	}
}

func TestValidateResource_ValidReference(t *testing.T) {
	v := NewValidator()
	data := json.RawMessage(`{
		"resourceType": "Observation",
		"id": "obs-1",
		"status": "final",
		"subject": {"reference": "Patient/abc-123"}
	}`)
	result := v.ValidateResource(data, false)

	if !result.Valid {
		t.Errorf("expected valid, got issues: %v", result.Issues)
	}
}

func TestValidateResource_InvalidReference(t *testing.T) {
	v := NewValidator()
	data := json.RawMessage(`{
		"resourceType": "Observation",
		"id": "obs-1",
		"status": "final",
		"subject": {"reference": "just-an-id"}
	}`)
	result := v.ValidateResource(data, false)

	if result.Valid {
		t.Error("expected invalid for malformed reference")
	}
	found := false
	for _, issue := range result.Issues {
		if issue.Code == IssueTypeValue {
			found = true
		}
	}
	if !found {
		t.Error("expected a value issue for invalid reference")
	}
}

func TestValidateResource_NestedReference(t *testing.T) {
	v := NewValidator()
	data := json.RawMessage(`{
		"resourceType": "MedicationRequest",
		"id": "mr-1",
		"status": "active",
		"subject": {"reference": "Patient/p1"},
		"requester": {"reference": "Practitioner/dr-1"},
		"dosageInstruction": [
			{"route": {"coding": [{"code": "oral"}]}}
		]
	}`)
	result := v.ValidateResource(data, false)

	if !result.Valid {
		t.Errorf("expected valid, got issues: %v", result.Issues)
	}
}

func TestValidateResource_InvalidNestedReference(t *testing.T) {
	v := NewValidator()
	data := json.RawMessage(`{
		"resourceType": "MedicationRequest",
		"id": "mr-1",
		"status": "active",
		"subject": {"reference": "Patient/p1"},
		"requester": {"reference": "bad-ref"}
	}`)
	result := v.ValidateResource(data, false)

	if result.Valid {
		t.Error("expected invalid for bad nested reference")
	}
}

func TestValidateResource_MultipleErrors(t *testing.T) {
	v := NewValidator()
	data := json.RawMessage(`{
		"status": "bogus",
		"subject": {"reference": "no-slash"}
	}`)
	result := v.ValidateResource(data, true)

	if result.Valid {
		t.Error("expected invalid")
	}
	// Should have issues for: missing resourceType, missing id, plus reference
	if len(result.Issues) < 2 {
		t.Errorf("expected at least 2 issues, got %d: %v", len(result.Issues), result.Issues)
	}
}

func TestValidateReferenceFormat(t *testing.T) {
	tests := []struct {
		ref   string
		valid bool
	}{
		{"Patient/123", true},
		{"Patient/abc-def", true},
		{"Patient/abc.def", true},
		{"Observation/obs-1", true},
		{"Practitioner/dr-smith", true},
		{"just-an-id", false},
		{"patient/123", false},  // lowercase resource type
		{"Patient/", false},     // no id
		{"", false},             // empty
		{"/Patient/123", false}, // leading slash
		{"Patient/123/extra", false},
	}

	for _, tt := range tests {
		t.Run(tt.ref, func(t *testing.T) {
			got := ValidateReferenceFormat(tt.ref)
			if got != tt.valid {
				t.Errorf("ValidateReferenceFormat(%q) = %v, want %v", tt.ref, got, tt.valid)
			}
		})
	}
}

func TestValidateResourceMap(t *testing.T) {
	v := NewValidator()
	resource := map[string]interface{}{
		"resourceType": "Patient",
		"id":           "123",
		"status":       "active",
	}

	result := v.ValidateResourceMap(resource, true)
	if !result.Valid {
		t.Errorf("expected valid, got issues: %v", result.Issues)
	}
}

func TestValidateResourceMap_Invalid(t *testing.T) {
	v := NewValidator()
	resource := map[string]interface{}{
		"id": "123",
	}

	result := v.ValidateResourceMap(resource, false)
	if result.Valid {
		t.Error("expected invalid for missing resourceType")
	}
}

func TestValidateBundleEntry_MissingRequest(t *testing.T) {
	v := NewValidator()
	entry := BundleEntry{}
	issues := v.ValidateBundleEntry(entry, 0)

	if len(issues) == 0 {
		t.Error("expected issues for missing request")
	}
}

func TestValidateBundleEntry_InvalidMethod(t *testing.T) {
	v := NewValidator()
	entry := BundleEntry{
		Request: &BundleRequest{Method: "PATCH", URL: "Patient/123"},
	}
	issues := v.ValidateBundleEntry(entry, 0)

	if len(issues) == 0 {
		t.Error("expected issues for invalid method")
	}
}

func TestValidateBundleEntry_MissingURL(t *testing.T) {
	v := NewValidator()
	entry := BundleEntry{
		Request: &BundleRequest{Method: "POST", URL: ""},
	}
	issues := v.ValidateBundleEntry(entry, 0)

	found := false
	for _, issue := range issues {
		if issue.Code == IssueTypeRequired {
			found = true
		}
	}
	if !found {
		t.Error("expected required issue for missing URL")
	}
}

func TestValidateBundleEntry_POSTMissingResource(t *testing.T) {
	v := NewValidator()
	entry := BundleEntry{
		Request: &BundleRequest{Method: "POST", URL: "Patient"},
	}
	issues := v.ValidateBundleEntry(entry, 0)

	found := false
	for _, issue := range issues {
		if issue.Code == IssueTypeRequired {
			found = true
		}
	}
	if !found {
		t.Error("expected required issue for POST without resource")
	}
}

func TestValidateBundleEntry_ValidPOST(t *testing.T) {
	v := NewValidator()
	resource, _ := json.Marshal(map[string]string{"resourceType": "Patient"})
	entry := BundleEntry{
		Request:  &BundleRequest{Method: "POST", URL: "Patient"},
		Resource: resource,
	}
	issues := v.ValidateBundleEntry(entry, 0)

	if len(issues) != 0 {
		t.Errorf("expected 0 issues for valid POST, got %d: %v", len(issues), issues)
	}
}

func TestValidateBundleEntry_ValidDELETE(t *testing.T) {
	v := NewValidator()
	entry := BundleEntry{
		Request: &BundleRequest{Method: "DELETE", URL: "Patient/123"},
	}
	issues := v.ValidateBundleEntry(entry, 0)

	if len(issues) != 0 {
		t.Errorf("expected 0 issues for valid DELETE, got %d: %v", len(issues), issues)
	}
}

func TestValidateBundle_InvalidType(t *testing.T) {
	v := NewValidator()
	bundle := &Bundle{Type: "searchset"}
	result := v.ValidateBundle(bundle)

	if result.Valid {
		t.Error("expected invalid for searchset type in processing context")
	}
}

func TestValidateBundle_EmptyEntries(t *testing.T) {
	v := NewValidator()
	bundle := &Bundle{Type: "transaction", Entry: []BundleEntry{}}
	result := v.ValidateBundle(bundle)

	if result.Valid {
		t.Error("expected invalid for empty entries")
	}
}

func TestValidateBundle_ValidTransaction(t *testing.T) {
	v := NewValidator()
	resource, _ := json.Marshal(map[string]string{"resourceType": "Patient"})
	bundle := &Bundle{
		Type: "transaction",
		Entry: []BundleEntry{
			{
				Request:  &BundleRequest{Method: "POST", URL: "Patient"},
				Resource: resource,
			},
			{
				Request: &BundleRequest{Method: "DELETE", URL: "Patient/old-1"},
			},
		},
	}
	result := v.ValidateBundle(bundle)

	if !result.Valid {
		t.Errorf("expected valid transaction bundle, got issues: %v", result.Issues)
	}
}

func TestValidateBundle_ValidBatch(t *testing.T) {
	v := NewValidator()
	resource, _ := json.Marshal(map[string]string{"resourceType": "Observation", "id": "obs-1"})
	bundle := &Bundle{
		Type: "batch",
		Entry: []BundleEntry{
			{
				Request:  &BundleRequest{Method: "PUT", URL: "Observation/obs-1"},
				Resource: resource,
			},
		},
	}
	result := v.ValidateBundle(bundle)

	if !result.Valid {
		t.Errorf("expected valid batch bundle, got issues: %v", result.Issues)
	}
}

func TestValidateBundle_InvalidEntry(t *testing.T) {
	v := NewValidator()
	bundle := &Bundle{
		Type: "transaction",
		Entry: []BundleEntry{
			{
				Request: &BundleRequest{Method: "POST", URL: "Patient"},
				// Missing resource for POST
			},
		},
	}
	result := v.ValidateBundle(bundle)

	if result.Valid {
		t.Error("expected invalid for entry missing resource")
	}
}

func TestIsKnownResourceType(t *testing.T) {
	if !IsKnownResourceType("Patient") {
		t.Error("expected Patient to be known")
	}
	if !IsKnownResourceType("Observation") {
		t.Error("expected Observation to be known")
	}
	if IsKnownResourceType("FakeResource") {
		t.Error("expected FakeResource to be unknown")
	}
}

func TestValidStatusValues(t *testing.T) {
	statuses := ValidStatusValues("Patient")
	if statuses == nil {
		t.Fatal("expected non-nil status values for Patient")
	}
	found := false
	for _, s := range statuses {
		if s == "active" {
			found = true
		}
	}
	if !found {
		t.Error("expected 'active' to be a valid Patient status")
	}

	unknown := ValidStatusValues("FakeResource")
	if unknown != nil {
		t.Error("expected nil status values for unknown resource type")
	}
}

func TestValidationResult_ToOperationOutcome(t *testing.T) {
	result := &ValidationResult{
		Valid: false,
		Issues: []OperationOutcomeIssue{
			{Severity: IssueSeverityError, Code: IssueTypeRequired, Diagnostics: "id is required"},
		},
	}

	outcome := result.ToOperationOutcome()
	if outcome.ResourceType != "OperationOutcome" {
		t.Errorf("expected OperationOutcome, got %s", outcome.ResourceType)
	}
	if len(outcome.Issue) != 1 {
		t.Fatalf("expected 1 issue, got %d", len(outcome.Issue))
	}
	if outcome.Issue[0].Diagnostics != "id is required" {
		t.Errorf("unexpected diagnostics: %s", outcome.Issue[0].Diagnostics)
	}
}

func TestValidateResource_NoStatusField(t *testing.T) {
	v := NewValidator()
	// A resource without a status field should be valid
	data := json.RawMessage(`{"resourceType": "Patient", "id": "123"}`)
	result := v.ValidateResource(data, false)

	if !result.Valid {
		t.Errorf("expected valid for resource without status, got issues: %v", result.Issues)
	}
}

func TestValidateResource_EmptyResourceType(t *testing.T) {
	v := NewValidator()
	data := json.RawMessage(`{"resourceType": "", "id": "123"}`)
	result := v.ValidateResource(data, false)

	if result.Valid {
		t.Error("expected invalid for empty resourceType")
	}
}

func TestValidateResource_StatusNotString(t *testing.T) {
	v := NewValidator()
	data := json.RawMessage(`{"resourceType": "Patient", "id": "123", "status": 42}`)
	result := v.ValidateResource(data, false)

	if result.Valid {
		t.Error("expected invalid for non-string status")
	}
}
