package fhir

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
)

// =========== ValidateResourceStructure Tests ===========

func TestValidateResourceStructure_MissingResourceType(t *testing.T) {
	resource := map[string]interface{}{
		"id": "123",
	}

	issues := ValidateResourceStructure(resource)

	if len(issues) == 0 {
		t.Fatal("expected issues for missing resourceType")
	}
	if issues[0].Severity != SeverityFatal {
		t.Errorf("expected fatal severity, got %s", issues[0].Severity)
	}
	if !strings.Contains(issues[0].Diagnostics, "resourceType is required") {
		t.Errorf("expected diagnostics about missing resourceType, got %s", issues[0].Diagnostics)
	}
}

func TestValidateResourceStructure_EmptyResourceType(t *testing.T) {
	resource := map[string]interface{}{
		"resourceType": "",
	}

	issues := ValidateResourceStructure(resource)

	if len(issues) == 0 {
		t.Fatal("expected issues for empty resourceType")
	}
	if issues[0].Severity != SeverityFatal {
		t.Errorf("expected fatal severity, got %s", issues[0].Severity)
	}
	if !strings.Contains(issues[0].Diagnostics, "non-empty string") {
		t.Errorf("expected diagnostics about non-empty string, got %s", issues[0].Diagnostics)
	}
}

func TestValidateResourceStructure_NonStringResourceType(t *testing.T) {
	resource := map[string]interface{}{
		"resourceType": 42,
	}

	issues := ValidateResourceStructure(resource)

	if len(issues) == 0 {
		t.Fatal("expected issues for non-string resourceType")
	}
	if issues[0].Severity != SeverityFatal {
		t.Errorf("expected fatal severity, got %s", issues[0].Severity)
	}
}

func TestValidateResourceStructure_UnknownResourceType(t *testing.T) {
	resource := map[string]interface{}{
		"resourceType": "FakeResource",
		"id":           "123",
	}

	issues := ValidateResourceStructure(resource)

	found := false
	for _, issue := range issues {
		if issue.Code == VIssueTypeStructure && strings.Contains(issue.Diagnostics, "Unknown resource type") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected issue about unknown resource type, got: %+v", issues)
	}
}

func TestValidateResourceStructure_ValidPatient(t *testing.T) {
	resource := map[string]interface{}{
		"resourceType": "Patient",
		"id":           "patient-123",
	}

	issues := ValidateResourceStructure(resource)

	for _, issue := range issues {
		if issue.Severity == SeverityError || issue.Severity == SeverityFatal {
			t.Errorf("expected no error/fatal issues for valid Patient, got: %+v", issues)
			break
		}
	}
}

func TestValidateResourceStructure_InvalidIDFormat(t *testing.T) {
	resource := map[string]interface{}{
		"resourceType": "Patient",
		"id":           "invalid id with spaces!!",
	}

	issues := ValidateResourceStructure(resource)

	found := false
	for _, issue := range issues {
		if issue.Code == VIssueTypeValue && strings.Contains(issue.Diagnostics, "does not match FHIR id format") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected issue about invalid id format, got: %+v", issues)
	}
}

func TestValidateResourceStructure_NonStringID(t *testing.T) {
	resource := map[string]interface{}{
		"resourceType": "Patient",
		"id":           12345,
	}

	issues := ValidateResourceStructure(resource)

	found := false
	for _, issue := range issues {
		if issue.Code == VIssueTypeValue && strings.Contains(issue.Diagnostics, "id must be a string") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected issue about id being a string, got: %+v", issues)
	}
}

func TestValidateResourceStructure_ValidIDFormats(t *testing.T) {
	validIDs := []string{
		"abc",
		"patient-123",
		"obs.vital.1",
		"A1b2-C3.d4",
		"a",
	}

	for _, id := range validIDs {
		resource := map[string]interface{}{
			"resourceType": "Patient",
			"id":           id,
		}
		issues := ValidateResourceStructure(resource)
		for _, issue := range issues {
			if issue.Location == "id" {
				t.Errorf("expected no id issues for valid id %q, got: %+v", id, issue)
			}
		}
	}
}

func TestValidateResourceStructure_IDTooLong(t *testing.T) {
	longID := strings.Repeat("a", 65)
	resource := map[string]interface{}{
		"resourceType": "Patient",
		"id":           longID,
	}

	issues := ValidateResourceStructure(resource)

	found := false
	for _, issue := range issues {
		if issue.Code == VIssueTypeValue && strings.Contains(issue.Diagnostics, "does not match FHIR id format") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected issue about id format for 65-char id, got: %+v", issues)
	}
}

func TestValidateResourceStructure_MetaVersionIDNotString(t *testing.T) {
	resource := map[string]interface{}{
		"resourceType": "Patient",
		"id":           "p1",
		"meta": map[string]interface{}{
			"versionId": 42,
		},
	}

	issues := ValidateResourceStructure(resource)

	found := false
	for _, issue := range issues {
		if issue.Location == "meta.versionId" && strings.Contains(issue.Diagnostics, "must be a string") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected issue about meta.versionId being a string, got: %+v", issues)
	}
}

func TestValidateResourceStructure_MetaNotObject(t *testing.T) {
	resource := map[string]interface{}{
		"resourceType": "Patient",
		"id":           "p1",
		"meta":         "not-an-object",
	}

	issues := ValidateResourceStructure(resource)

	found := false
	for _, issue := range issues {
		if issue.Location == "meta" && strings.Contains(issue.Diagnostics, "meta must be an object") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected issue about meta being an object, got: %+v", issues)
	}
}

func TestValidateResourceStructure_ValidMetaVersionID(t *testing.T) {
	resource := map[string]interface{}{
		"resourceType": "Patient",
		"id":           "p1",
		"meta": map[string]interface{}{
			"versionId": "1",
		},
	}

	issues := ValidateResourceStructure(resource)

	for _, issue := range issues {
		if issue.Location == "meta.versionId" {
			t.Errorf("expected no meta.versionId issue, got: %+v", issue)
		}
	}
}

func TestValidateResourceStructure_MissingStatusForObservation(t *testing.T) {
	resource := map[string]interface{}{
		"resourceType": "Observation",
		"id":           "obs-1",
		"code": map[string]interface{}{
			"text": "blood pressure",
		},
	}

	issues := ValidateResourceStructure(resource)

	found := false
	for _, issue := range issues {
		if issue.Code == VIssueTypeRequired && strings.Contains(issue.Diagnostics, "status") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected issue about missing status for Observation, got: %+v", issues)
	}
}

func TestValidateResourceStructure_StatusPresentForObservation(t *testing.T) {
	resource := map[string]interface{}{
		"resourceType": "Observation",
		"id":           "obs-1",
		"status":       "final",
		"code": map[string]interface{}{
			"text": "blood pressure",
		},
	}

	issues := ValidateResourceStructure(resource)

	for _, issue := range issues {
		if issue.Code == VIssueTypeRequired && strings.Contains(issue.Diagnostics, "status") {
			t.Errorf("did not expect issue about missing status when status is present, got: %+v", issue)
		}
	}
}

func TestValidateResourceStructure_GoalLifecycleStatus(t *testing.T) {
	// Goal uses lifecycleStatus instead of status.
	resource := map[string]interface{}{
		"resourceType": "Goal",
		"id":           "goal-1",
		"subject": map[string]interface{}{
			"reference": "Patient/p1",
		},
	}

	issues := ValidateResourceStructure(resource)

	found := false
	for _, issue := range issues {
		if issue.Code == VIssueTypeRequired && strings.Contains(issue.Diagnostics, "lifecycleStatus") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected issue about missing lifecycleStatus for Goal, got: %+v", issues)
	}

	// Now provide lifecycleStatus; the issue should not appear.
	resource["lifecycleStatus"] = "active"
	issues = ValidateResourceStructure(resource)
	for _, issue := range issues {
		if issue.Code == VIssueTypeRequired && strings.Contains(issue.Diagnostics, "lifecycleStatus") {
			t.Errorf("did not expect lifecycleStatus issue when present, got: %+v", issue)
		}
	}
}

func TestValidateResourceStructure_NoStatusRequiredForPatient(t *testing.T) {
	// Patient does not have status in requiredFieldsRegistry.
	resource := map[string]interface{}{
		"resourceType": "Patient",
		"id":           "p1",
	}

	issues := ValidateResourceStructure(resource)

	for _, issue := range issues {
		if issue.Code == VIssueTypeRequired && strings.Contains(issue.Diagnostics, "status") {
			t.Errorf("did not expect status requirement for Patient, got: %+v", issue)
		}
	}
}

// =========== BatchValidateHandler Tests ===========

func setupBatchValidateHandler() (*BatchValidateHandler, *echo.Echo) {
	e := echo.New()
	validator := NewResourceValidator()
	handler := NewBatchValidateHandler(validator)
	return handler, e
}

func TestBatchValidateHandler_EmptyBody(t *testing.T) {
	handler, e := setupBatchValidateHandler()

	req := httptest.NewRequest(http.MethodPost, "/fhir/$batch-validate", strings.NewReader(""))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.Handle(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestBatchValidateHandler_InvalidJSON(t *testing.T) {
	handler, e := setupBatchValidateHandler()

	req := httptest.NewRequest(http.MethodPost, "/fhir/$batch-validate", strings.NewReader("{not json"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.Handle(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestBatchValidateHandler_NoResources(t *testing.T) {
	handler, e := setupBatchValidateHandler()

	body := `{"resources": []}`
	req := httptest.NewRequest(http.MethodPost, "/fhir/$batch-validate", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.Handle(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestBatchValidateHandler_SingleValidResource(t *testing.T) {
	handler, e := setupBatchValidateHandler()

	body := `{
		"resources": [
			{
				"resourceType": "Patient",
				"id": "patient-1",
				"name": [{"family": "Smith", "given": ["John"]}]
			}
		]
	}`
	req := httptest.NewRequest(http.MethodPost, "/fhir/$batch-validate", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.Handle(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var result BatchValidateResult
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if result.TotalCount != 1 {
		t.Errorf("expected totalCount 1, got %d", result.TotalCount)
	}
	if result.ValidCount != 1 {
		t.Errorf("expected validCount 1, got %d", result.ValidCount)
	}
	if result.InvalidCount != 0 {
		t.Errorf("expected invalidCount 0, got %d", result.InvalidCount)
	}
	if len(result.Results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result.Results))
	}
	if !result.Results[0].Valid {
		t.Errorf("expected resource to be valid, got invalid with issues: %+v", result.Results[0].Issues)
	}
	if result.Results[0].ResourceType != "Patient" {
		t.Errorf("expected resourceType Patient, got %s", result.Results[0].ResourceType)
	}
	if result.Results[0].ResourceID != "patient-1" {
		t.Errorf("expected resourceId patient-1, got %s", result.Results[0].ResourceID)
	}
}

func TestBatchValidateHandler_SingleInvalidResource(t *testing.T) {
	handler, e := setupBatchValidateHandler()

	body := `{
		"resources": [
			{
				"id": "no-type"
			}
		]
	}`
	req := httptest.NewRequest(http.MethodPost, "/fhir/$batch-validate", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.Handle(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var result BatchValidateResult
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if result.TotalCount != 1 {
		t.Errorf("expected totalCount 1, got %d", result.TotalCount)
	}
	if result.ValidCount != 0 {
		t.Errorf("expected validCount 0, got %d", result.ValidCount)
	}
	if result.InvalidCount != 1 {
		t.Errorf("expected invalidCount 1, got %d", result.InvalidCount)
	}
	if result.Results[0].Valid {
		t.Error("expected resource to be invalid")
	}
	if len(result.Results[0].Issues) == 0 {
		t.Error("expected issues to be reported")
	}
}

func TestBatchValidateHandler_MixedResources(t *testing.T) {
	handler, e := setupBatchValidateHandler()

	body := `{
		"resources": [
			{
				"resourceType": "Patient",
				"id": "patient-1",
				"name": [{"family": "Smith"}]
			},
			{
				"id": "missing-type"
			},
			{
				"resourceType": "Observation",
				"id": "obs-1",
				"status": "final",
				"code": {"text": "BP"}
			}
		]
	}`
	req := httptest.NewRequest(http.MethodPost, "/fhir/$batch-validate", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.Handle(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var result BatchValidateResult
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if result.TotalCount != 3 {
		t.Errorf("expected totalCount 3, got %d", result.TotalCount)
	}

	// First resource (Patient) should be valid.
	if !result.Results[0].Valid {
		t.Errorf("expected first resource (Patient) to be valid, issues: %+v", result.Results[0].Issues)
	}

	// Second resource (missing resourceType) should be invalid.
	if result.Results[1].Valid {
		t.Error("expected second resource to be invalid (missing resourceType)")
	}

	// Third resource (Observation) should be valid.
	if !result.Results[2].Valid {
		t.Errorf("expected third resource (Observation) to be valid, issues: %+v", result.Results[2].Issues)
	}

	// Counts should reflect mixed results.
	if result.ValidCount != 2 {
		t.Errorf("expected validCount 2, got %d", result.ValidCount)
	}
	if result.InvalidCount != 1 {
		t.Errorf("expected invalidCount 1, got %d", result.InvalidCount)
	}
}

func TestBatchValidateHandler_InvalidIDInResource(t *testing.T) {
	handler, e := setupBatchValidateHandler()

	body := `{
		"resources": [
			{
				"resourceType": "Patient",
				"id": "invalid id!!",
				"name": [{"family": "Doe"}]
			}
		]
	}`
	req := httptest.NewRequest(http.MethodPost, "/fhir/$batch-validate", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.Handle(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result BatchValidateResult
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if result.Results[0].Valid {
		t.Error("expected resource with invalid id to be invalid")
	}

	foundIDIssue := false
	for _, issue := range result.Results[0].Issues {
		diag, _ := issue["diagnostics"].(string)
		if strings.Contains(diag, "id") && strings.Contains(diag, "does not match") {
			foundIDIssue = true
			break
		}
	}
	if !foundIDIssue {
		t.Errorf("expected id format issue in results, got: %+v", result.Results[0].Issues)
	}
}

func TestBatchValidateHandler_InvalidResourceJSON(t *testing.T) {
	handler, e := setupBatchValidateHandler()

	// Include a resource that is not valid JSON (number instead of object).
	body := `{
		"resources": [
			42
		]
	}`
	req := httptest.NewRequest(http.MethodPost, "/fhir/$batch-validate", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.Handle(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result BatchValidateResult
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if result.TotalCount != 1 {
		t.Errorf("expected totalCount 1, got %d", result.TotalCount)
	}
	if result.Results[0].Valid {
		t.Error("expected non-object resource to be invalid")
	}
}

func TestBatchValidateHandler_IndexOrdering(t *testing.T) {
	handler, e := setupBatchValidateHandler()

	body := `{
		"resources": [
			{"resourceType": "Patient", "id": "p1", "name": [{"family": "A"}]},
			{"resourceType": "Patient", "id": "p2", "name": [{"family": "B"}]},
			{"resourceType": "Patient", "id": "p3", "name": [{"family": "C"}]}
		]
	}`
	req := httptest.NewRequest(http.MethodPost, "/fhir/$batch-validate", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.Handle(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result BatchValidateResult
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	for i, r := range result.Results {
		if r.Index != i {
			t.Errorf("expected index %d, got %d", i, r.Index)
		}
	}
}

// =========== Bundle Input Tests ===========

func TestBatchValidateHandler_BundleInput(t *testing.T) {
	handler, e := setupBatchValidateHandler()

	body := `{
		"resourceType": "Bundle",
		"type": "collection",
		"entry": [
			{
				"resource": {
					"resourceType": "Patient",
					"id": "p1",
					"name": [{"family": "Smith"}]
				}
			},
			{
				"resource": {
					"resourceType": "Observation",
					"id": "obs-1",
					"status": "final",
					"code": {"text": "Glucose"}
				}
			}
		]
	}`
	req := httptest.NewRequest(http.MethodPost, "/fhir/$batch-validate", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.Handle(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var result BatchValidateResult
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if result.TotalCount != 2 {
		t.Errorf("expected totalCount 2, got %d", result.TotalCount)
	}
	if result.Results[0].ResourceType != "Patient" {
		t.Errorf("expected first resource type Patient, got %s", result.Results[0].ResourceType)
	}
	if result.Results[1].ResourceType != "Observation" {
		t.Errorf("expected second resource type Observation, got %s", result.Results[1].ResourceType)
	}
}

func TestBatchValidateHandler_BundleNoEntries(t *testing.T) {
	handler, e := setupBatchValidateHandler()

	body := `{
		"resourceType": "Bundle",
		"type": "collection",
		"entry": []
	}`
	req := httptest.NewRequest(http.MethodPost, "/fhir/$batch-validate", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.Handle(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for empty Bundle entries, got %d", rec.Code)
	}
}

func TestBatchValidateHandler_BundleEntryMissingResource(t *testing.T) {
	handler, e := setupBatchValidateHandler()

	body := `{
		"resourceType": "Bundle",
		"type": "collection",
		"entry": [
			{
				"fullUrl": "Patient/p1"
			}
		]
	}`
	req := httptest.NewRequest(http.MethodPost, "/fhir/$batch-validate", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.Handle(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for entry missing resource, got %d", rec.Code)
	}
}

func TestBatchValidateHandler_BundleWithInvalidResources(t *testing.T) {
	handler, e := setupBatchValidateHandler()

	body := `{
		"resourceType": "Bundle",
		"type": "collection",
		"entry": [
			{
				"resource": {
					"resourceType": "Patient",
					"id": "p1",
					"name": [{"family": "Valid"}]
				}
			},
			{
				"resource": {
					"id": "no-type"
				}
			}
		]
	}`
	req := httptest.NewRequest(http.MethodPost, "/fhir/$batch-validate", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.Handle(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var result BatchValidateResult
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if result.ValidCount != 1 {
		t.Errorf("expected validCount 1, got %d", result.ValidCount)
	}
	if result.InvalidCount != 1 {
		t.Errorf("expected invalidCount 1, got %d", result.InvalidCount)
	}
}

// =========== RegisterRoutes Test ===========

func TestBatchValidateHandler_RegisterRoutes(t *testing.T) {
	e := echo.New()
	g := e.Group("/fhir")
	validator := NewResourceValidator()
	handler := NewBatchValidateHandler(validator)
	handler.RegisterRoutes(g)

	body := `{
		"resources": [
			{"resourceType": "Patient", "id": "p1", "name": [{"family": "Test"}]}
		]
	}`
	req := httptest.NewRequest(http.MethodPost, "/fhir/$batch-validate", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 via registered route, got %d", rec.Code)
	}
}

// =========== ResourceMetadata Extraction Tests ===========

func TestBatchValidateHandler_ResourceTypeAndIDExtracted(t *testing.T) {
	handler, e := setupBatchValidateHandler()

	body := `{
		"resources": [
			{
				"resourceType": "Encounter",
				"id": "enc-42",
				"status": "finished",
				"class": {"code": "AMB"}
			}
		]
	}`
	req := httptest.NewRequest(http.MethodPost, "/fhir/$batch-validate", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.Handle(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result BatchValidateResult
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if result.Results[0].ResourceType != "Encounter" {
		t.Errorf("expected resourceType Encounter, got %s", result.Results[0].ResourceType)
	}
	if result.Results[0].ResourceID != "enc-42" {
		t.Errorf("expected resourceId enc-42, got %s", result.Results[0].ResourceID)
	}
}

func TestBatchValidateHandler_NoResourceID(t *testing.T) {
	handler, e := setupBatchValidateHandler()

	body := `{
		"resources": [
			{
				"resourceType": "Patient",
				"name": [{"family": "NoID"}]
			}
		]
	}`
	req := httptest.NewRequest(http.MethodPost, "/fhir/$batch-validate", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.Handle(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result BatchValidateResult
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if result.Results[0].ResourceID != "" {
		t.Errorf("expected empty resourceId, got %s", result.Results[0].ResourceID)
	}
}

// =========== Edge Cases ===========

func TestBatchValidateHandler_UnknownResourceType(t *testing.T) {
	handler, e := setupBatchValidateHandler()

	body := `{
		"resources": [
			{
				"resourceType": "FakeResource",
				"id": "fake-1"
			}
		]
	}`
	req := httptest.NewRequest(http.MethodPost, "/fhir/$batch-validate", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.Handle(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result BatchValidateResult
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if result.Results[0].Valid {
		t.Error("expected unknown resourceType to be invalid")
	}
	if result.Results[0].ResourceType != "FakeResource" {
		t.Errorf("expected resourceType FakeResource, got %s", result.Results[0].ResourceType)
	}
}

func TestBatchValidateHandler_LargerBatch(t *testing.T) {
	handler, e := setupBatchValidateHandler()

	resources := make([]string, 10)
	for i := 0; i < 10; i++ {
		resources[i] = `{"resourceType": "Patient", "id": "p` + strings.Repeat("x", i+1) + `", "name": [{"family": "Test"}]}`
	}
	body := `{"resources": [` + strings.Join(resources, ",") + `]}`

	req := httptest.NewRequest(http.MethodPost, "/fhir/$batch-validate", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.Handle(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result BatchValidateResult
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if result.TotalCount != 10 {
		t.Errorf("expected totalCount 10, got %d", result.TotalCount)
	}
	if result.ValidCount != 10 {
		t.Errorf("expected validCount 10, got %d", result.ValidCount)
	}
}
