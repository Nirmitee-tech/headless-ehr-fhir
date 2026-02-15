package fhir

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
)

// =========== ResourceValidator Tests ===========

func TestResourceValidator_ValidPatient(t *testing.T) {
	v := NewResourceValidator()
	resource := map[string]interface{}{
		"resourceType": "Patient",
		"id":           "patient-123",
		"name": []interface{}{
			map[string]interface{}{"family": "Smith", "given": []interface{}{"John"}},
		},
	}

	result := v.Validate(resource)

	if !result.Valid {
		t.Errorf("expected valid Patient, got invalid with issues: %+v", result.Issues)
	}

	hasErrors := false
	for _, issue := range result.Issues {
		if issue.Severity == SeverityError || issue.Severity == SeverityFatal {
			hasErrors = true
			break
		}
	}
	if hasErrors {
		t.Errorf("expected no error/fatal issues for valid Patient, got: %+v", result.Issues)
	}
}

func TestResourceValidator_MissingResourceType(t *testing.T) {
	v := NewResourceValidator()
	resource := map[string]interface{}{
		"id": "123",
	}

	result := v.Validate(resource)

	if result.Valid {
		t.Error("expected invalid for missing resourceType")
	}

	found := false
	for _, issue := range result.Issues {
		if issue.Severity == SeverityFatal && issue.Code == VIssueTypeStructure &&
			strings.Contains(issue.Diagnostics, "resourceType") {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected fatal structure issue for missing resourceType")
	}
}

func TestResourceValidator_UnknownResourceType(t *testing.T) {
	v := NewResourceValidator()
	resource := map[string]interface{}{
		"resourceType": "FakeResource",
		"id":           "123",
	}

	result := v.Validate(resource)

	if result.Valid {
		t.Error("expected invalid for unknown resourceType")
	}

	found := false
	for _, issue := range result.Issues {
		if issue.Code == VIssueTypeStructure && strings.Contains(issue.Diagnostics, "Unknown resource type") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected structure issue for unknown resourceType, got: %+v", result.Issues)
	}
}

func TestResourceValidator_RequiredFields_Observation(t *testing.T) {
	v := NewResourceValidator()
	resource := map[string]interface{}{
		"resourceType": "Observation",
	}

	result := v.Validate(resource)

	if result.Valid {
		t.Error("expected invalid for Observation missing status and code")
	}

	requiredMissing := map[string]bool{"status": false, "code": false}
	for _, issue := range result.Issues {
		if issue.Code == VIssueTypeRequired {
			if strings.Contains(issue.Diagnostics, "'status'") {
				requiredMissing["status"] = true
			}
			if strings.Contains(issue.Diagnostics, "'code'") {
				requiredMissing["code"] = true
			}
		}
	}

	for field, found := range requiredMissing {
		if !found {
			t.Errorf("expected required field error for '%s'", field)
		}
	}
}

func TestResourceValidator_RequiredFields_MedicationRequest(t *testing.T) {
	v := NewResourceValidator()
	resource := map[string]interface{}{
		"resourceType": "MedicationRequest",
	}

	result := v.Validate(resource)

	if result.Valid {
		t.Error("expected invalid for MedicationRequest missing required fields")
	}

	// Should report missing: status, intent, medication[x], subject
	expectedFields := []string{"status", "intent", "medication", "subject"}
	for _, field := range expectedFields {
		found := false
		for _, issue := range result.Issues {
			if issue.Code == VIssueTypeRequired && strings.Contains(issue.Diagnostics, field) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected required field error for '%s', issues: %+v", field, result.Issues)
		}
	}
}

func TestResourceValidator_RequiredFields_Encounter(t *testing.T) {
	v := NewResourceValidator()
	resource := map[string]interface{}{
		"resourceType": "Encounter",
	}

	result := v.Validate(resource)

	if result.Valid {
		t.Error("expected invalid for Encounter missing status and class")
	}

	expectedFields := []string{"status", "class"}
	for _, field := range expectedFields {
		found := false
		for _, issue := range result.Issues {
			if issue.Code == VIssueTypeRequired && strings.Contains(issue.Diagnostics, field) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected required field error for '%s'", field)
		}
	}
}

func TestResourceValidator_InvalidID(t *testing.T) {
	v := NewResourceValidator()
	resource := map[string]interface{}{
		"resourceType": "Practitioner",
		"id":           "bad id!@#$",
	}

	result := v.Validate(resource)

	if result.Valid {
		t.Error("expected invalid for id with special characters")
	}

	found := false
	for _, issue := range result.Issues {
		if issue.Code == VIssueTypeValue && issue.Location == "id" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected value issue for invalid id, got: %+v", result.Issues)
	}
}

func TestResourceValidator_ValidID(t *testing.T) {
	v := NewResourceValidator()
	resource := map[string]interface{}{
		"resourceType": "Practitioner",
		"id":           "valid-id.123",
	}

	result := v.Validate(resource)

	// Check that there are no id-related issues
	for _, issue := range result.Issues {
		if issue.Location == "id" && (issue.Severity == SeverityError || issue.Severity == SeverityFatal) {
			t.Errorf("expected no id issues for valid id, got: %+v", issue)
		}
	}
}

func TestResourceValidator_InvalidReference(t *testing.T) {
	v := NewResourceValidator()
	resource := map[string]interface{}{
		"resourceType": "Observation",
		"id":           "obs-1",
		"status":       "final",
		"code":         map[string]interface{}{"text": "BP"},
		"subject":      map[string]interface{}{"reference": "just-an-id"},
		"valueQuantity": map[string]interface{}{
			"value": 120,
			"unit":  "mmHg",
		},
	}

	result := v.Validate(resource)

	found := false
	for _, issue := range result.Issues {
		if issue.Code == VIssueTypeValue && strings.Contains(issue.Diagnostics, "Reference") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected warning for invalid reference format, got: %+v", result.Issues)
	}
}

func TestResourceValidator_ValidReference(t *testing.T) {
	v := NewResourceValidator()
	resource := map[string]interface{}{
		"resourceType": "Observation",
		"id":           "obs-1",
		"status":       "final",
		"code":         map[string]interface{}{"text": "BP"},
		"subject":      map[string]interface{}{"reference": "Patient/123"},
		"valueQuantity": map[string]interface{}{
			"value": 120,
			"unit":  "mmHg",
		},
	}

	result := v.Validate(resource)

	for _, issue := range result.Issues {
		if issue.Code == VIssueTypeValue && strings.Contains(issue.Diagnostics, "Reference") {
			t.Errorf("expected no reference issues for valid reference, got: %+v", issue)
		}
	}
}

func TestResourceValidator_InvalidStatus(t *testing.T) {
	v := NewResourceValidator()
	resource := map[string]interface{}{
		"resourceType": "Observation",
		"id":           "obs-1",
		"status":       "bogus-status",
		"code":         map[string]interface{}{"text": "BP"},
		"valueQuantity": map[string]interface{}{
			"value": 120,
			"unit":  "mmHg",
		},
	}

	result := v.Validate(resource)

	if result.Valid {
		t.Error("expected invalid for unknown status")
	}

	found := false
	for _, issue := range result.Issues {
		if issue.Code == VIssueTypeValue && strings.Contains(issue.Diagnostics, "Invalid status") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected value issue for invalid status, got: %+v", result.Issues)
	}
}

func TestResourceValidator_ValidStatus(t *testing.T) {
	v := NewResourceValidator()
	resource := map[string]interface{}{
		"resourceType": "Observation",
		"id":           "obs-1",
		"status":       "final",
		"code":         map[string]interface{}{"text": "BP"},
		"valueQuantity": map[string]interface{}{
			"value": 120,
			"unit":  "mmHg",
		},
	}

	result := v.Validate(resource)

	for _, issue := range result.Issues {
		if issue.Code == VIssueTypeValue && strings.Contains(issue.Diagnostics, "status") {
			t.Errorf("expected no status issues for valid status, got: %+v", issue)
		}
	}
}

func TestResourceValidator_InvalidDate(t *testing.T) {
	v := NewResourceValidator()
	resource := map[string]interface{}{
		"resourceType": "Patient",
		"id":           "p-1",
		"name": []interface{}{
			map[string]interface{}{"family": "Smith"},
		},
		"birthDate": "not-a-date",
	}

	result := v.Validate(resource)

	if result.Valid {
		t.Error("expected invalid for bad date format")
	}

	found := false
	for _, issue := range result.Issues {
		if issue.Code == VIssueTypeValue && strings.Contains(issue.Diagnostics, "date") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected value issue for invalid date, got: %+v", result.Issues)
	}
}

func TestResourceValidator_BusinessRule_PatientNameOrIdentifier(t *testing.T) {
	v := NewResourceValidator()
	resource := map[string]interface{}{
		"resourceType": "Patient",
		"id":           "p-1",
	}

	result := v.Validate(resource)

	if result.Valid {
		t.Error("expected invalid for Patient with no name and no identifier")
	}

	found := false
	for _, issue := range result.Issues {
		if issue.Code == VIssueTypeBusinessRule && strings.Contains(issue.Diagnostics, "name or identifier") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected business rule issue for Patient without name or identifier, got: %+v", result.Issues)
	}

	// Now test with identifier instead of name
	resourceWithIdentifier := map[string]interface{}{
		"resourceType": "Patient",
		"id":           "p-2",
		"identifier": []interface{}{
			map[string]interface{}{"system": "urn:mrn", "value": "12345"},
		},
	}

	result2 := v.Validate(resourceWithIdentifier)
	for _, issue := range result2.Issues {
		if issue.Code == VIssueTypeBusinessRule && strings.Contains(issue.Diagnostics, "name or identifier") {
			t.Errorf("Patient with identifier should not fail name/identifier rule, got: %+v", issue)
		}
	}
}

func TestResourceValidator_BusinessRule_ObservationFinalValue(t *testing.T) {
	v := NewResourceValidator()
	resource := map[string]interface{}{
		"resourceType": "Observation",
		"id":           "obs-1",
		"status":       "final",
		"code":         map[string]interface{}{"text": "test"},
	}

	result := v.Validate(resource)

	found := false
	for _, issue := range result.Issues {
		if issue.Code == VIssueTypeBusinessRule && issue.Severity == SeverityWarning &&
			strings.Contains(issue.Diagnostics, "final") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected warning for final Observation without value, got: %+v", result.Issues)
	}

	// With value, the warning should not appear.
	resourceWithValue := map[string]interface{}{
		"resourceType": "Observation",
		"id":           "obs-2",
		"status":       "final",
		"code":         map[string]interface{}{"text": "test"},
		"valueQuantity": map[string]interface{}{
			"value": 98.6,
			"unit":  "F",
		},
	}

	result2 := v.Validate(resourceWithValue)
	for _, issue := range result2.Issues {
		if issue.Code == VIssueTypeBusinessRule && strings.Contains(issue.Diagnostics, "final") {
			t.Errorf("Observation with value should not trigger final-without-value warning, got: %+v", issue)
		}
	}
}

func TestResourceValidator_BusinessRule_MedicationChoice(t *testing.T) {
	v := NewResourceValidator()
	// MedicationRequest with neither medicationCodeableConcept nor medicationReference.
	resource := map[string]interface{}{
		"resourceType": "MedicationRequest",
		"id":           "mr-1",
		"status":       "active",
		"intent":       "order",
		"subject":      map[string]interface{}{"reference": "Patient/p1"},
	}

	result := v.Validate(resource)

	if result.Valid {
		t.Error("expected invalid for MedicationRequest without medication[x]")
	}

	found := false
	for _, issue := range result.Issues {
		if strings.Contains(issue.Diagnostics, "medicationCodeableConcept") ||
			strings.Contains(issue.Diagnostics, "medicationReference") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected issue about medication choice, got: %+v", result.Issues)
	}

	// With medicationCodeableConcept, should pass.
	resourceWithMed := map[string]interface{}{
		"resourceType": "MedicationRequest",
		"id":           "mr-2",
		"status":       "active",
		"intent":       "order",
		"subject":      map[string]interface{}{"reference": "Patient/p1"},
		"medicationCodeableConcept": map[string]interface{}{
			"coding": []interface{}{
				map[string]interface{}{"code": "317896006"},
			},
		},
	}

	result2 := v.Validate(resourceWithMed)
	for _, issue := range result2.Issues {
		if issue.Code == VIssueTypeRequired && strings.Contains(issue.Diagnostics, "medication") {
			t.Errorf("MedicationRequest with medicationCodeableConcept should not fail medication check, got: %+v", issue)
		}
	}
}

func TestResourceValidator_MultipleIssues(t *testing.T) {
	v := NewResourceValidator()
	// Resource with several problems: missing required fields, bad id, bad status.
	resource := map[string]interface{}{
		"resourceType": "Observation",
		"id":           "bad id!@#$",
		"status":       "bogus",
		"birthDate":    "not-a-date",
	}

	result := v.Validate(resource)

	if result.Valid {
		t.Error("expected invalid")
	}

	// Should have at least: invalid id, invalid status, missing code, invalid date
	if len(result.Issues) < 3 {
		t.Errorf("expected at least 3 issues for resource with multiple problems, got %d: %+v",
			len(result.Issues), result.Issues)
	}
}

func TestResourceValidator_ValidateWithMode_Create(t *testing.T) {
	v := NewResourceValidator()
	// In create mode, id check should be skipped.
	resource := map[string]interface{}{
		"resourceType": "Practitioner",
		// No id -- this is fine in create mode.
	}

	result := v.ValidateWithMode(resource, "create")

	for _, issue := range result.Issues {
		if issue.Location == "id" {
			t.Errorf("in create mode, id should not be validated, but got issue: %+v", issue)
		}
	}
}

func TestResourceValidator_ValidateWithMode_Update(t *testing.T) {
	v := NewResourceValidator()
	// In update mode, id is required.
	resource := map[string]interface{}{
		"resourceType": "Practitioner",
		// No id -- this should fail in update mode.
	}

	result := v.ValidateWithMode(resource, "update")

	if result.Valid {
		t.Error("expected invalid for update mode without id")
	}

	found := false
	for _, issue := range result.Issues {
		if issue.Code == VIssueTypeRequired && issue.Location == "id" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected required id issue for update mode, got: %+v", result.Issues)
	}
}

func TestResourceValidator_AllResourceTypes(t *testing.T) {
	v := NewResourceValidator()

	// For each registered resource type, create a minimal valid resource and ensure no errors.
	for rt, requiredFields := range requiredFieldsRegistry {
		t.Run(rt, func(t *testing.T) {
			resource := map[string]interface{}{
				"resourceType": rt,
				"id":           "test-1",
			}

			// Add required fields with minimal values.
			for _, field := range requiredFields {
				switch field {
				case "name":
					resource["name"] = []interface{}{
						map[string]interface{}{"family": "Test"},
					}
				case "status":
					// Use the first valid status for this resource type.
					if statuses, ok := statusValues[rt]; ok && len(statuses) > 0 {
						resource["status"] = statuses[0]
					} else {
						resource["status"] = "active"
					}
				case "lifecycleStatus":
					resource["lifecycleStatus"] = "active"
				case "code", "type", "scope", "vaccineCode":
					resource[field] = map[string]interface{}{"text": "test-code"}
				case "subject", "patient", "beneficiary":
					resource[field] = map[string]interface{}{"reference": "Patient/p1"}
				case "class":
					resource["class"] = map[string]interface{}{"code": "AMB"}
				case "intent":
					resource["intent"] = "order"
				case "medication":
					resource["medicationCodeableConcept"] = map[string]interface{}{
						"text": "test medication",
					}
				case "content":
					resource["content"] = []interface{}{
						map[string]interface{}{
							"attachment": map[string]interface{}{"url": "http://example.com/doc.pdf"},
						},
					}
				case "category":
					resource["category"] = []interface{}{
						map[string]interface{}{"text": "test-category"},
					}
				case "date":
					resource["date"] = "2024-01-01"
				case "author":
					resource["author"] = []interface{}{
						map[string]interface{}{"reference": "Practitioner/pr1"},
					}
				case "title":
					resource["title"] = "Test Title"
				case "provider":
					resource["provider"] = map[string]interface{}{"reference": "Practitioner/pr1"}
				case "relationship":
					resource["relationship"] = map[string]interface{}{"text": "mother"}
				case "occurrenceDateTime":
					resource["occurrenceDateTime"] = "2024-01-15"
				case "actor":
					resource["actor"] = []interface{}{
						map[string]interface{}{"reference": "Practitioner/pr1"},
					}
				case "schedule":
					resource["schedule"] = map[string]interface{}{"reference": "Schedule/sch1"}
				case "start":
					resource["start"] = "2024-01-15T09:00:00Z"
				case "end":
					resource["end"] = "2024-01-15T10:00:00Z"
				default:
					resource[field] = "test-value"
				}
			}

			// Patient also needs name or identifier for business rule.
			if rt == "Patient" {
				if _, ok := resource["name"]; !ok {
					resource["name"] = []interface{}{
						map[string]interface{}{"family": "Test"},
					}
				}
			}

			result := v.Validate(resource)

			for _, issue := range result.Issues {
				if issue.Severity == SeverityError || issue.Severity == SeverityFatal {
					t.Errorf("resource type %s: unexpected error issue: %+v", rt, issue)
				}
			}
		})
	}
}

// =========== ValidateHandler Tests ===========

func TestValidateHandler_Success(t *testing.T) {
	v := NewResourceValidator()
	h := NewValidateHandler(v)
	e := echo.New()

	body := `{
		"resourceType": "Patient",
		"id": "p-1",
		"name": [{"family": "Smith"}]
	}`

	req := httptest.NewRequest(http.MethodPost, "/fhir/Patient/$validate", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("resourceType")
	c.SetParamValues("Patient")

	err := h.Validate(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var outcome map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &outcome); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if outcome["resourceType"] != "OperationOutcome" {
		t.Errorf("expected resourceType OperationOutcome, got %v", outcome["resourceType"])
	}

	issues, ok := outcome["issue"].([]interface{})
	if !ok {
		t.Fatal("expected issue array in response")
	}

	// Valid resource should have an information-level success message.
	for _, issueRaw := range issues {
		issue := issueRaw.(map[string]interface{})
		if issue["severity"] == "error" || issue["severity"] == "fatal" {
			t.Errorf("expected no error/fatal issues for valid resource, got: %v", issue)
		}
	}
}

func TestValidateHandler_InvalidResource(t *testing.T) {
	v := NewResourceValidator()
	h := NewValidateHandler(v)
	e := echo.New()

	// Observation missing required fields.
	body := `{
		"resourceType": "Observation",
		"id": "obs-1"
	}`

	req := httptest.NewRequest(http.MethodPost, "/fhir/Observation/$validate", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("resourceType")
	c.SetParamValues("Observation")

	err := h.Validate(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Even invalid resources should return 200 with OperationOutcome.
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 (even for invalid resource), got %d", rec.Code)
	}

	var outcome map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &outcome); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	issues, ok := outcome["issue"].([]interface{})
	if !ok {
		t.Fatal("expected issue array in response")
	}

	hasError := false
	for _, issueRaw := range issues {
		issue := issueRaw.(map[string]interface{})
		if issue["severity"] == "error" {
			hasError = true
			break
		}
	}
	if !hasError {
		t.Error("expected at least one error issue for invalid resource")
	}
}

func TestValidateHandler_TypeMismatch(t *testing.T) {
	v := NewResourceValidator()
	h := NewValidateHandler(v)
	e := echo.New()

	// URL says Patient but body says Observation.
	body := `{
		"resourceType": "Observation",
		"id": "obs-1",
		"status": "final",
		"code": {"text": "BP"}
	}`

	req := httptest.NewRequest(http.MethodPost, "/fhir/Patient/$validate", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("resourceType")
	c.SetParamValues("Patient")

	err := h.Validate(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for type mismatch, got %d", rec.Code)
	}

	var outcome map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &outcome); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	issues, ok := outcome["issue"].([]interface{})
	if !ok || len(issues) == 0 {
		t.Fatal("expected issues in response")
	}

	issue := issues[0].(map[string]interface{})
	if !strings.Contains(issue["diagnostics"].(string), "does not match") {
		t.Errorf("expected diagnostics about type mismatch, got: %v", issue["diagnostics"])
	}
}

func TestValidateHandler_EmptyBody(t *testing.T) {
	v := NewResourceValidator()
	h := NewValidateHandler(v)
	e := echo.New()

	req := httptest.NewRequest(http.MethodPost, "/fhir/$validate", strings.NewReader(""))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.Validate(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for empty body, got %d", rec.Code)
	}
}

func TestValidateHandler_InvalidJSON(t *testing.T) {
	v := NewResourceValidator()
	h := NewValidateHandler(v)
	e := echo.New()

	req := httptest.NewRequest(http.MethodPost, "/fhir/$validate", strings.NewReader("{not valid json}"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.Validate(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid JSON, got %d", rec.Code)
	}

	var outcome map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &outcome); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	issues, ok := outcome["issue"].([]interface{})
	if !ok || len(issues) == 0 {
		t.Fatal("expected issues in response")
	}

	issue := issues[0].(map[string]interface{})
	if !strings.Contains(issue["diagnostics"].(string), "Invalid JSON") {
		t.Errorf("expected diagnostics about invalid JSON, got: %v", issue["diagnostics"])
	}
}

func TestValidateHandler_ModeParam(t *testing.T) {
	v := NewResourceValidator()
	h := NewValidateHandler(v)
	e := echo.New()

	// Create mode: no id should be fine.
	body := `{
		"resourceType": "Practitioner"
	}`

	req := httptest.NewRequest(http.MethodPost, "/fhir/Practitioner/$validate?mode=create", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("resourceType")
	c.SetParamValues("Practitioner")

	err := h.Validate(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var outcome map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &outcome); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	issues := outcome["issue"].([]interface{})
	for _, issueRaw := range issues {
		issue := issueRaw.(map[string]interface{})
		if issue["severity"] == "error" || issue["severity"] == "fatal" {
			locs, _ := issue["location"].([]interface{})
			if len(locs) > 0 && locs[0] == "id" {
				t.Errorf("create mode should not produce id errors, got: %v", issue)
			}
		}
	}
}

func TestValidateHandler_GeneralEndpoint(t *testing.T) {
	v := NewResourceValidator()
	h := NewValidateHandler(v)
	e := echo.New()

	// POST to /fhir/$validate without resource type in URL.
	body := `{
		"resourceType": "Patient",
		"id": "p-1",
		"name": [{"family": "Doe"}]
	}`

	req := httptest.NewRequest(http.MethodPost, "/fhir/$validate", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.Validate(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var outcome map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &outcome); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if outcome["resourceType"] != "OperationOutcome" {
		t.Errorf("expected OperationOutcome, got %v", outcome["resourceType"])
	}
}

func TestValidateHandler_ProfileParamWarning(t *testing.T) {
	v := NewResourceValidator()
	h := NewValidateHandler(v)
	e := echo.New()

	body := `{
		"resourceType": "Patient",
		"id": "p-1",
		"name": [{"family": "Smith"}]
	}`

	profileURL := "http://hl7.org/fhir/StructureDefinition/Patient"
	req := httptest.NewRequest(http.MethodPost, "/fhir/$validate?profile="+profileURL, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.Validate(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var outcome map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &outcome); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if outcome["resourceType"] != "OperationOutcome" {
		t.Errorf("expected resourceType OperationOutcome, got %v", outcome["resourceType"])
	}

	issues, ok := outcome["issue"].([]interface{})
	if !ok || len(issues) == 0 {
		t.Fatal("expected issues in response")
	}

	// The profile warning should be the first issue.
	firstIssue := issues[0].(map[string]interface{})
	if firstIssue["severity"] != "warning" {
		t.Errorf("expected first issue severity 'warning', got %v", firstIssue["severity"])
	}
	if firstIssue["code"] != "invariant" {
		t.Errorf("expected first issue code 'invariant', got %v", firstIssue["code"])
	}

	diag, _ := firstIssue["diagnostics"].(string)
	if !strings.Contains(diag, "Profile validation") {
		t.Errorf("expected diagnostics to mention profile validation, got: %s", diag)
	}
	if !strings.Contains(diag, profileURL) {
		t.Errorf("expected diagnostics to contain the profile URL '%s', got: %s", profileURL, diag)
	}
}

func TestValidateHandler_RegisterRoutes(t *testing.T) {
	v := NewResourceValidator()
	h := NewValidateHandler(v)
	e := echo.New()
	fhirGroup := e.Group("/fhir")

	h.RegisterRoutes(fhirGroup)

	routes := e.Routes()
	routePaths := make(map[string]bool)
	for _, r := range routes {
		routePaths[r.Method+":"+r.Path] = true
	}

	expected := []string{
		"POST:/fhir/$validate",
		"POST:/fhir/:resourceType/$validate",
	}
	for _, path := range expected {
		if !routePaths[path] {
			t.Errorf("missing expected route: %s (registered: %v)", path, routePaths)
		}
	}
}
