package fhir

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
)

// =========== ParseGraphDefinition Tests ===========

func TestParseGraphDefinition_Valid(t *testing.T) {
	data := `{
		"resourceType": "GraphDefinition",
		"id": "gd-1",
		"name": "PatientGraph",
		"start": "Patient",
		"link": [
			{
				"path": "subject",
				"target": [
					{
						"type": "Observation",
						"params": "subject={ref}"
					}
				]
			}
		]
	}`

	gd, err := ParseGraphDefinition([]byte(data))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gd.ResourceType != "GraphDefinition" {
		t.Errorf("expected resourceType 'GraphDefinition', got '%s'", gd.ResourceType)
	}
	if gd.Name != "PatientGraph" {
		t.Errorf("expected name 'PatientGraph', got '%s'", gd.Name)
	}
	if gd.Start != "Patient" {
		t.Errorf("expected start 'Patient', got '%s'", gd.Start)
	}
	if len(gd.Link) != 1 {
		t.Fatalf("expected 1 link, got %d", len(gd.Link))
	}
	if gd.Link[0].Path != "subject" {
		t.Errorf("expected link path 'subject', got '%s'", gd.Link[0].Path)
	}
	if len(gd.Link[0].Target) != 1 {
		t.Fatalf("expected 1 target, got %d", len(gd.Link[0].Target))
	}
	if gd.Link[0].Target[0].Type != "Observation" {
		t.Errorf("expected target type 'Observation', got '%s'", gd.Link[0].Target[0].Type)
	}
}

func TestParseGraphDefinition_InvalidJSON(t *testing.T) {
	_, err := ParseGraphDefinition([]byte("{bad json"))
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestParseGraphDefinition_EmptyJSON(t *testing.T) {
	gd, err := ParseGraphDefinition([]byte("{}"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gd.Name != "" {
		t.Errorf("expected empty name, got '%s'", gd.Name)
	}
}

func TestParseGraphDefinition_WithCompartments(t *testing.T) {
	data := `{
		"resourceType": "GraphDefinition",
		"name": "EncounterGraph",
		"start": "Encounter",
		"link": [
			{
				"path": "subject",
				"target": [
					{
						"type": "Patient",
						"compartment": [
							{
								"use": "requirement",
								"code": "Patient",
								"rule": "identical"
							}
						]
					}
				]
			}
		]
	}`

	gd, err := ParseGraphDefinition([]byte(data))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(gd.Link[0].Target[0].Compartment) != 1 {
		t.Fatalf("expected 1 compartment, got %d", len(gd.Link[0].Target[0].Compartment))
	}

	comp := gd.Link[0].Target[0].Compartment[0]
	if comp.Use != "requirement" {
		t.Errorf("expected use 'requirement', got '%s'", comp.Use)
	}
	if comp.Code != "Patient" {
		t.Errorf("expected code 'Patient', got '%s'", comp.Code)
	}
	if comp.Rule != "identical" {
		t.Errorf("expected rule 'identical', got '%s'", comp.Rule)
	}
}

func TestParseGraphDefinition_NestedLinks(t *testing.T) {
	data := `{
		"resourceType": "GraphDefinition",
		"name": "DeepGraph",
		"start": "Patient",
		"link": [
			{
				"path": "subject",
				"target": [
					{
						"type": "Encounter",
						"link": [
							{
								"path": "diagnosis.condition",
								"target": [
									{"type": "Condition"}
								]
							}
						]
					}
				]
			}
		]
	}`

	gd, err := ParseGraphDefinition([]byte(data))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(gd.Link[0].Target[0].Link) != 1 {
		t.Fatalf("expected 1 nested link, got %d", len(gd.Link[0].Target[0].Link))
	}

	nestedLink := gd.Link[0].Target[0].Link[0]
	if nestedLink.Path != "diagnosis.condition" {
		t.Errorf("expected nested path 'diagnosis.condition', got '%s'", nestedLink.Path)
	}
	if nestedLink.Target[0].Type != "Condition" {
		t.Errorf("expected nested target type 'Condition', got '%s'", nestedLink.Target[0].Type)
	}
}

// =========== ValidateGraphDefinition Tests ===========

func TestValidateGraphDefinition_Valid(t *testing.T) {
	gd := &GraphDefinition{
		ResourceType: "GraphDefinition",
		Name:         "TestGraph",
		Start:        "Patient",
		Link: []GraphLink{
			{
				Path: "subject",
				Target: []GraphTarget{
					{Type: "Observation"},
				},
			},
		},
	}

	issues := ValidateGraphDefinition(gd)
	for _, issue := range issues {
		if issue.Severity == SeverityError || issue.Severity == SeverityFatal {
			t.Errorf("unexpected error issue: %+v", issue)
		}
	}
}

func TestValidateGraphDefinition_Nil(t *testing.T) {
	issues := ValidateGraphDefinition(nil)
	if len(issues) == 0 {
		t.Error("expected issues for nil GraphDefinition")
	}

	found := false
	for _, issue := range issues {
		if issue.Severity == SeverityFatal {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected fatal issue for nil GraphDefinition")
	}
}

func TestValidateGraphDefinition_MissingName(t *testing.T) {
	gd := &GraphDefinition{
		Start: "Patient",
	}

	issues := ValidateGraphDefinition(gd)
	found := false
	for _, issue := range issues {
		if issue.Code == VIssueTypeRequired && strings.Contains(issue.Diagnostics, "name") {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected required issue for missing name")
	}
}

func TestValidateGraphDefinition_MissingStart(t *testing.T) {
	gd := &GraphDefinition{
		Name: "TestGraph",
	}

	issues := ValidateGraphDefinition(gd)
	found := false
	for _, issue := range issues {
		if issue.Code == VIssueTypeRequired && strings.Contains(issue.Diagnostics, "start") {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected required issue for missing start")
	}
}

func TestValidateGraphDefinition_WrongResourceType(t *testing.T) {
	gd := &GraphDefinition{
		ResourceType: "Patient",
		Name:         "TestGraph",
		Start:        "Patient",
	}

	issues := ValidateGraphDefinition(gd)
	found := false
	for _, issue := range issues {
		if issue.Code == VIssueTypeStructure && strings.Contains(issue.Diagnostics, "GraphDefinition") {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected structure issue for wrong resourceType")
	}
}

func TestValidateGraphDefinition_MissingTargetType(t *testing.T) {
	gd := &GraphDefinition{
		Name:  "TestGraph",
		Start: "Patient",
		Link: []GraphLink{
			{
				Path: "subject",
				Target: []GraphTarget{
					{Type: ""},
				},
			},
		},
	}

	issues := ValidateGraphDefinition(gd)
	found := false
	for _, issue := range issues {
		if issue.Code == VIssueTypeRequired && strings.Contains(issue.Diagnostics, "target type") {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected required issue for missing target type")
	}
}

func TestValidateGraphDefinition_EmptyPath(t *testing.T) {
	gd := &GraphDefinition{
		Name:  "TestGraph",
		Start: "Patient",
		Link: []GraphLink{
			{
				Path:   "",
				Target: []GraphTarget{{Type: "Observation"}},
			},
		},
	}

	issues := ValidateGraphDefinition(gd)
	found := false
	for _, issue := range issues {
		if issue.Severity == SeverityWarning && strings.Contains(issue.Diagnostics, "path is empty") {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected warning for empty link path")
	}
}

func TestValidateGraphDefinition_InvalidCompartment(t *testing.T) {
	gd := &GraphDefinition{
		Name:  "TestGraph",
		Start: "Patient",
		Link: []GraphLink{
			{
				Path: "subject",
				Target: []GraphTarget{
					{
						Type: "Observation",
						Compartment: []GraphCompartment{
							{
								Use:  "invalid-use",
								Code: "",
								Rule: "invalid-rule",
							},
						},
					},
				},
			},
		},
	}

	issues := ValidateGraphDefinition(gd)

	errorCount := 0
	for _, issue := range issues {
		if issue.Severity == SeverityError {
			errorCount++
		}
	}

	// Should have errors for: invalid use, empty code, invalid rule
	if errorCount < 3 {
		t.Errorf("expected at least 3 error issues for invalid compartment, got %d: %+v", errorCount, issues)
	}
}

func TestValidateGraphDefinition_ValidCompartment(t *testing.T) {
	gd := &GraphDefinition{
		Name:  "TestGraph",
		Start: "Patient",
		Link: []GraphLink{
			{
				Path: "subject",
				Target: []GraphTarget{
					{
						Type: "Observation",
						Compartment: []GraphCompartment{
							{
								Use:  "condition",
								Code: "Patient",
								Rule: "matching",
							},
						},
					},
				},
			},
		},
	}

	issues := ValidateGraphDefinition(gd)
	for _, issue := range issues {
		if issue.Severity == SeverityError || issue.Severity == SeverityFatal {
			t.Errorf("unexpected error for valid compartment: %+v", issue)
		}
	}
}

// =========== extractRefsFromPath Tests ===========

func TestExtractRefsFromPath_SimpleField(t *testing.T) {
	resource := map[string]interface{}{
		"subject": map[string]interface{}{
			"reference": "Patient/123",
		},
	}

	refs := extractRefsFromPath(resource, "subject")
	if len(refs) != 1 {
		t.Fatalf("expected 1 ref, got %d", len(refs))
	}
	if refs[0] != "Patient/123" {
		t.Errorf("expected 'Patient/123', got '%s'", refs[0])
	}
}

func TestExtractRefsFromPath_NestedField(t *testing.T) {
	resource := map[string]interface{}{
		"participant": []interface{}{
			map[string]interface{}{
				"individual": map[string]interface{}{
					"reference": "Practitioner/pr-1",
				},
			},
			map[string]interface{}{
				"individual": map[string]interface{}{
					"reference": "Practitioner/pr-2",
				},
			},
		},
	}

	refs := extractRefsFromPath(resource, "participant.individual")
	if len(refs) != 2 {
		t.Fatalf("expected 2 refs, got %d: %v", len(refs), refs)
	}
	if refs[0] != "Practitioner/pr-1" {
		t.Errorf("expected 'Practitioner/pr-1', got '%s'", refs[0])
	}
	if refs[1] != "Practitioner/pr-2" {
		t.Errorf("expected 'Practitioner/pr-2', got '%s'", refs[1])
	}
}

func TestExtractRefsFromPath_ArrayOfReferences(t *testing.T) {
	resource := map[string]interface{}{
		"author": []interface{}{
			map[string]interface{}{"reference": "Practitioner/pr-1"},
			map[string]interface{}{"reference": "Practitioner/pr-2"},
		},
	}

	refs := extractRefsFromPath(resource, "author")
	if len(refs) != 2 {
		t.Fatalf("expected 2 refs, got %d", len(refs))
	}
}

func TestExtractRefsFromPath_EmptyPath(t *testing.T) {
	resource := map[string]interface{}{
		"subject": map[string]interface{}{
			"reference": "Patient/123",
		},
	}

	refs := extractRefsFromPath(resource, "")
	if len(refs) != 0 {
		t.Errorf("expected 0 refs for empty path, got %d", len(refs))
	}
}

func TestExtractRefsFromPath_MissingField(t *testing.T) {
	resource := map[string]interface{}{
		"status": "active",
	}

	refs := extractRefsFromPath(resource, "subject")
	if len(refs) != 0 {
		t.Errorf("expected 0 refs for missing field, got %d", len(refs))
	}
}

// =========== parseReference Tests ===========

func TestParseReference(t *testing.T) {
	tests := []struct {
		ref      string
		wantType string
		wantID   string
	}{
		{"Patient/123", "Patient", "123"},
		{"Observation/obs-1", "Observation", "obs-1"},
		{"invalid", "", ""},
		{"", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.ref, func(t *testing.T) {
			gotType, gotID := parseReference(tt.ref)
			if gotType != tt.wantType {
				t.Errorf("parseReference(%q) type = %q, want %q", tt.ref, gotType, tt.wantType)
			}
			if gotID != tt.wantID {
				t.Errorf("parseReference(%q) id = %q, want %q", tt.ref, gotID, tt.wantID)
			}
		})
	}
}

// =========== graphTraverser Tests ===========

func newTestIncludeRegistry() *IncludeRegistry {
	registry := NewIncludeRegistry()

	patients := map[string]map[string]interface{}{
		"p-1": {
			"resourceType": "Patient",
			"id":           "p-1",
			"name":         []interface{}{map[string]interface{}{"family": "Smith"}},
		},
		"p-2": {
			"resourceType": "Patient",
			"id":           "p-2",
			"name":         []interface{}{map[string]interface{}{"family": "Jones"}},
		},
	}
	observations := map[string]map[string]interface{}{
		"obs-1": {
			"resourceType": "Observation",
			"id":           "obs-1",
			"status":       "final",
			"subject":      map[string]interface{}{"reference": "Patient/p-1"},
		},
		"obs-2": {
			"resourceType": "Observation",
			"id":           "obs-2",
			"status":       "final",
			"subject":      map[string]interface{}{"reference": "Patient/p-1"},
		},
	}
	encounters := map[string]map[string]interface{}{
		"enc-1": {
			"resourceType": "Encounter",
			"id":           "enc-1",
			"status":       "finished",
			"subject":      map[string]interface{}{"reference": "Patient/p-1"},
			"participant": []interface{}{
				map[string]interface{}{
					"individual": map[string]interface{}{"reference": "Practitioner/prac-1"},
				},
			},
		},
	}
	practitioners := map[string]map[string]interface{}{
		"prac-1": {
			"resourceType": "Practitioner",
			"id":           "prac-1",
			"name":         []interface{}{map[string]interface{}{"family": "Dr. Smith"}},
		},
	}

	registry.RegisterFetcher("Patient", func(_ context.Context, id string) (map[string]interface{}, error) {
		if r, ok := patients[id]; ok {
			return r, nil
		}
		return nil, fmt.Errorf("not found")
	})
	registry.RegisterFetcher("Observation", func(_ context.Context, id string) (map[string]interface{}, error) {
		if r, ok := observations[id]; ok {
			return r, nil
		}
		return nil, fmt.Errorf("not found")
	})
	registry.RegisterFetcher("Encounter", func(_ context.Context, id string) (map[string]interface{}, error) {
		if r, ok := encounters[id]; ok {
			return r, nil
		}
		return nil, fmt.Errorf("not found")
	})
	registry.RegisterFetcher("Practitioner", func(_ context.Context, id string) (map[string]interface{}, error) {
		if r, ok := practitioners[id]; ok {
			return r, nil
		}
		return nil, fmt.Errorf("not found")
	})

	return registry
}

func TestGraphTraverser_SimpleTraversal(t *testing.T) {
	registry := newTestIncludeRegistry()
	traverser := newGraphTraverser(registry)

	startResource := map[string]interface{}{
		"resourceType": "Observation",
		"id":           "obs-1",
		"status":       "final",
		"subject":      map[string]interface{}{"reference": "Patient/p-1"},
	}
	traverser.addResource(startResource)

	links := []GraphLink{
		{
			Path: "subject",
			Target: []GraphTarget{
				{Type: "Patient"},
			},
		},
	}

	traverser.traverse(context.Background(), startResource, links)

	if len(traverser.entries) != 2 {
		t.Errorf("expected 2 entries (Observation + Patient), got %d", len(traverser.entries))
	}

	// Check that both Observation and Patient are present.
	keys := make(map[string]bool)
	for _, entry := range traverser.entries {
		keys[entry.FullURL] = true
	}
	if !keys["Observation/obs-1"] {
		t.Error("expected Observation/obs-1 in results")
	}
	if !keys["Patient/p-1"] {
		t.Error("expected Patient/p-1 in results")
	}
}

func TestGraphTraverser_NestedTraversal(t *testing.T) {
	registry := newTestIncludeRegistry()
	traverser := newGraphTraverser(registry)

	startResource := map[string]interface{}{
		"resourceType": "Encounter",
		"id":           "enc-1",
		"status":       "finished",
		"subject":      map[string]interface{}{"reference": "Patient/p-1"},
		"participant": []interface{}{
			map[string]interface{}{
				"individual": map[string]interface{}{"reference": "Practitioner/prac-1"},
			},
		},
	}
	traverser.addResource(startResource)

	links := []GraphLink{
		{
			Path: "subject",
			Target: []GraphTarget{
				{Type: "Patient"},
			},
		},
		{
			Path: "participant.individual",
			Target: []GraphTarget{
				{Type: "Practitioner"},
			},
		},
	}

	traverser.traverse(context.Background(), startResource, links)

	if len(traverser.entries) != 3 {
		t.Errorf("expected 3 entries (Encounter + Patient + Practitioner), got %d", len(traverser.entries))
	}
}

func TestGraphTraverser_DeduplicatesResources(t *testing.T) {
	registry := newTestIncludeRegistry()
	traverser := newGraphTraverser(registry)

	startResource := map[string]interface{}{
		"resourceType": "Observation",
		"id":           "obs-1",
		"status":       "final",
		"subject":      map[string]interface{}{"reference": "Patient/p-1"},
	}
	traverser.addResource(startResource)

	// Two links pointing to the same patient.
	links := []GraphLink{
		{
			Path:   "subject",
			Target: []GraphTarget{{Type: "Patient"}},
		},
		{
			Path:   "subject",
			Target: []GraphTarget{{Type: "Patient"}},
		},
	}

	traverser.traverse(context.Background(), startResource, links)

	if len(traverser.entries) != 2 {
		t.Errorf("expected 2 entries (deduplicated), got %d", len(traverser.entries))
	}
}

func TestGraphTraverser_MissingResource(t *testing.T) {
	registry := newTestIncludeRegistry()
	traverser := newGraphTraverser(registry)

	startResource := map[string]interface{}{
		"resourceType": "Observation",
		"id":           "obs-1",
		"subject":      map[string]interface{}{"reference": "Patient/nonexistent"},
	}
	traverser.addResource(startResource)

	links := []GraphLink{
		{
			Path:   "subject",
			Target: []GraphTarget{{Type: "Patient"}},
		},
	}

	traverser.traverse(context.Background(), startResource, links)

	// Only the start resource should be present.
	if len(traverser.entries) != 1 {
		t.Errorf("expected 1 entry for missing referenced resource, got %d", len(traverser.entries))
	}
}

func TestGraphTraverser_TypeMismatch(t *testing.T) {
	registry := newTestIncludeRegistry()
	traverser := newGraphTraverser(registry)

	startResource := map[string]interface{}{
		"resourceType": "Observation",
		"id":           "obs-1",
		"subject":      map[string]interface{}{"reference": "Patient/p-1"},
	}
	traverser.addResource(startResource)

	// Target type doesn't match the reference.
	links := []GraphLink{
		{
			Path:   "subject",
			Target: []GraphTarget{{Type: "Encounter"}},
		},
	}

	traverser.traverse(context.Background(), startResource, links)

	// Only the start resource; reference type doesn't match target.
	if len(traverser.entries) != 1 {
		t.Errorf("expected 1 entry for type mismatch, got %d", len(traverser.entries))
	}
}

// =========== GraphApplyHandler Tests ===========

func TestGraphApplyHandler_Success(t *testing.T) {
	registry := newTestIncludeRegistry()
	handler := GraphApplyHandler(registry)

	body := `{
		"graphDefinition": {
			"resourceType": "GraphDefinition",
			"name": "ObservationPatient",
			"start": "Observation",
			"link": [
				{
					"path": "subject",
					"target": [{"type": "Patient"}]
				}
			]
		},
		"resourceId": "obs-1"
	}`

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/fhir/$graph", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/fhir+json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var bundle Bundle
	if err := json.Unmarshal(rec.Body.Bytes(), &bundle); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if bundle.ResourceType != "Bundle" {
		t.Errorf("expected resourceType 'Bundle', got '%s'", bundle.ResourceType)
	}
	if bundle.Type != "collection" {
		t.Errorf("expected type 'collection', got '%s'", bundle.Type)
	}
	if len(bundle.Entry) != 2 {
		t.Errorf("expected 2 entries (Observation + Patient), got %d", len(bundle.Entry))
	}
}

func TestGraphApplyHandler_ResourceTypeOverride(t *testing.T) {
	registry := newTestIncludeRegistry()
	handler := GraphApplyHandler(registry)

	body := `{
		"graphDefinition": {
			"resourceType": "GraphDefinition",
			"name": "PatientGraph",
			"start": "ShouldBeOverridden",
			"link": []
		},
		"resourceId": "p-1",
		"resourceType": "Patient"
	}`

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/fhir/$graph", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/fhir+json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var bundle Bundle
	if err := json.Unmarshal(rec.Body.Bytes(), &bundle); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if len(bundle.Entry) != 1 {
		t.Errorf("expected 1 entry (just the Patient), got %d", len(bundle.Entry))
	}
}

func TestGraphApplyHandler_EmptyBody(t *testing.T) {
	registry := newTestIncludeRegistry()
	handler := GraphApplyHandler(registry)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/fhir/$graph", strings.NewReader(""))
	req.Header.Set("Content-Type", "application/fhir+json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestGraphApplyHandler_InvalidJSON(t *testing.T) {
	registry := newTestIncludeRegistry()
	handler := GraphApplyHandler(registry)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/fhir/$graph", strings.NewReader("{bad json"))
	req.Header.Set("Content-Type", "application/fhir+json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestGraphApplyHandler_MissingGraphDefinition(t *testing.T) {
	registry := newTestIncludeRegistry()
	handler := GraphApplyHandler(registry)

	body := `{"resourceId": "obs-1"}`

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/fhir/$graph", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/fhir+json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestGraphApplyHandler_MissingResourceID(t *testing.T) {
	registry := newTestIncludeRegistry()
	handler := GraphApplyHandler(registry)

	body := `{
		"graphDefinition": {
			"resourceType": "GraphDefinition",
			"name": "TestGraph",
			"start": "Patient"
		}
	}`

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/fhir/$graph", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/fhir+json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestGraphApplyHandler_InvalidGraphDefinition(t *testing.T) {
	registry := newTestIncludeRegistry()
	handler := GraphApplyHandler(registry)

	// GraphDefinition missing required name and start.
	body := `{
		"graphDefinition": {
			"resourceType": "GraphDefinition"
		},
		"resourceId": "obs-1"
	}`

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/fhir/$graph", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/fhir+json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestGraphApplyHandler_ResourceNotFound(t *testing.T) {
	registry := newTestIncludeRegistry()
	handler := GraphApplyHandler(registry)

	body := `{
		"graphDefinition": {
			"resourceType": "GraphDefinition",
			"name": "TestGraph",
			"start": "Patient"
		},
		"resourceId": "nonexistent"
	}`

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/fhir/$graph", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/fhir+json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestGraphApplyHandler_UnknownResourceType(t *testing.T) {
	registry := newTestIncludeRegistry()
	handler := GraphApplyHandler(registry)

	body := `{
		"graphDefinition": {
			"resourceType": "GraphDefinition",
			"name": "TestGraph",
			"start": "UnknownType"
		},
		"resourceId": "some-id"
	}`

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/fhir/$graph", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/fhir+json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestGraphApplyHandler_RecursiveLinks(t *testing.T) {
	registry := newTestIncludeRegistry()
	handler := GraphApplyHandler(registry)

	// Encounter -> Patient (via subject) and Encounter -> Practitioner (via participant.individual)
	body := `{
		"graphDefinition": {
			"resourceType": "GraphDefinition",
			"name": "EncounterFullGraph",
			"start": "Encounter",
			"link": [
				{
					"path": "subject",
					"target": [{"type": "Patient"}]
				},
				{
					"path": "participant.individual",
					"target": [{"type": "Practitioner"}]
				}
			]
		},
		"resourceId": "enc-1"
	}`

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/fhir/$graph", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/fhir+json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var bundle Bundle
	if err := json.Unmarshal(rec.Body.Bytes(), &bundle); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	// Encounter + Patient + Practitioner = 3 entries
	if len(bundle.Entry) != 3 {
		t.Errorf("expected 3 entries (Encounter + Patient + Practitioner), got %d", len(bundle.Entry))
	}

	// Verify the total matches.
	if bundle.Total == nil || *bundle.Total != 3 {
		t.Errorf("expected total 3, got %v", bundle.Total)
	}
}
