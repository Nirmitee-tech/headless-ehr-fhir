package fhir

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
)

// ============================================================================
// Test Helpers
// ============================================================================

func makeApplyRequest(subject, encounter, practitioner, organization string) *ApplyRequest {
	return &ApplyRequest{
		Subject:      subject,
		Encounter:    encounter,
		Practitioner: practitioner,
		Organization: organization,
		Parameters:   make(map[string]interface{}),
	}
}

func makePlanDefinitionJSON(id, status, title string, actions []map[string]interface{}) map[string]interface{} {
	pd := map[string]interface{}{
		"resourceType": "PlanDefinition",
		"id":           id,
		"status":       status,
		"title":        title,
	}
	if actions != nil {
		pd["action"] = actions
	}
	return pd
}

func makeActivityDefinitionJSON(id, status, kind string) map[string]interface{} {
	return map[string]interface{}{
		"resourceType": "ActivityDefinition",
		"id":           id,
		"status":       status,
		"kind":         kind,
	}
}

func makeActionJSON(id, title, actionType string) map[string]interface{} {
	a := map[string]interface{}{
		"id":    id,
		"title": title,
	}
	if actionType != "" {
		a["type"] = map[string]interface{}{
			"coding": []interface{}{map[string]interface{}{"code": actionType}},
		}
	}
	return a
}

// mockApplyResolver implements ResourceResolver for apply tests.
type mockApplyResolver struct {
	resources map[string]map[string]interface{}
}

func newMockApplyResolver() *mockApplyResolver {
	return &mockApplyResolver{
		resources: make(map[string]map[string]interface{}),
	}
}

func (m *mockApplyResolver) ResolveReference(_ interface{}, reference string) (map[string]interface{}, error) {
	r, ok := m.resources[reference]
	if !ok {
		return nil, nil
	}
	return r, nil
}

// ============================================================================
// TestParsePlanDefinition
// ============================================================================

func TestParsePlanDefinition_Valid(t *testing.T) {
	data := makePlanDefinitionJSON("pd-1", "active", "Diabetes Protocol", []map[string]interface{}{
		{
			"id":          "action-1",
			"title":       "Order Lab",
			"description": "Order HbA1c lab test",
			"priority":    "routine",
			"type": map[string]interface{}{
				"coding": []interface{}{map[string]interface{}{"code": "create"}},
			},
			"definitionCanonical": "ActivityDefinition/ad-1",
		},
	})

	parsed, err := ParseApplyPlanDefinition(data)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if parsed.ID != "pd-1" {
		t.Errorf("expected ID pd-1, got %s", parsed.ID)
	}
	if parsed.Status != "active" {
		t.Errorf("expected status active, got %s", parsed.Status)
	}
	if parsed.Title != "Diabetes Protocol" {
		t.Errorf("expected title Diabetes Protocol, got %s", parsed.Title)
	}
	if len(parsed.Actions) != 1 {
		t.Fatalf("expected 1 action, got %d", len(parsed.Actions))
	}
	if parsed.Actions[0].ID != "action-1" {
		t.Errorf("expected action ID action-1, got %s", parsed.Actions[0].ID)
	}
	if parsed.Actions[0].Title != "Order Lab" {
		t.Errorf("expected action title Order Lab, got %s", parsed.Actions[0].Title)
	}
	if parsed.Actions[0].Priority != "routine" {
		t.Errorf("expected priority routine, got %s", parsed.Actions[0].Priority)
	}
	if parsed.Actions[0].Type != "create" {
		t.Errorf("expected type create, got %s", parsed.Actions[0].Type)
	}
	if parsed.Actions[0].DefinitionURI != "ActivityDefinition/ad-1" {
		t.Errorf("expected definitionURI ActivityDefinition/ad-1, got %s", parsed.Actions[0].DefinitionURI)
	}
}

func TestParsePlanDefinition_NestedActions(t *testing.T) {
	data := makePlanDefinitionJSON("pd-nested", "active", "Nested Plan", []map[string]interface{}{
		{
			"id":    "parent",
			"title": "Parent Action",
			"selectionBehavior": "exactly-one",
			"groupingBehavior":  "logical-group",
			"action": []interface{}{
				map[string]interface{}{
					"id":    "child-1",
					"title": "Child Action 1",
				},
				map[string]interface{}{
					"id":    "child-2",
					"title": "Child Action 2",
				},
			},
		},
	})

	parsed, err := ParseApplyPlanDefinition(data)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(parsed.Actions) != 1 {
		t.Fatalf("expected 1 top-level action, got %d", len(parsed.Actions))
	}
	parent := parsed.Actions[0]
	if parent.SelectionBehavior != "exactly-one" {
		t.Errorf("expected selectionBehavior exactly-one, got %s", parent.SelectionBehavior)
	}
	if parent.GroupingBehavior != "logical-group" {
		t.Errorf("expected groupingBehavior logical-group, got %s", parent.GroupingBehavior)
	}
	if len(parent.Action) != 2 {
		t.Fatalf("expected 2 child actions, got %d", len(parent.Action))
	}
	if parent.Action[0].ID != "child-1" {
		t.Errorf("expected child ID child-1, got %s", parent.Action[0].ID)
	}
	if parent.Action[1].ID != "child-2" {
		t.Errorf("expected child ID child-2, got %s", parent.Action[1].ID)
	}
}

func TestParsePlanDefinition_Empty(t *testing.T) {
	data := makePlanDefinitionJSON("pd-empty", "draft", "Empty Plan", nil)
	parsed, err := ParseApplyPlanDefinition(data)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(parsed.Actions) != 0 {
		t.Errorf("expected 0 actions, got %d", len(parsed.Actions))
	}
}

func TestParsePlanDefinition_InvalidResourceType(t *testing.T) {
	data := map[string]interface{}{
		"resourceType": "Patient",
		"id":           "pd-bad",
	}
	_, err := ParseApplyPlanDefinition(data)
	if err == nil {
		t.Fatal("expected error for invalid resourceType, got nil")
	}
}

func TestParsePlanDefinition_NilData(t *testing.T) {
	_, err := ParseApplyPlanDefinition(nil)
	if err == nil {
		t.Fatal("expected error for nil data, got nil")
	}
}

func TestParsePlanDefinition_WithConditions(t *testing.T) {
	data := makePlanDefinitionJSON("pd-cond", "active", "Conditional Plan", []map[string]interface{}{
		{
			"id":    "action-cond",
			"title": "Conditional Action",
			"condition": []interface{}{
				map[string]interface{}{
					"kind": "applicability",
					"expression": map[string]interface{}{
						"expression": "%age > 40",
					},
				},
			},
		},
	})

	parsed, err := ParseApplyPlanDefinition(data)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(parsed.Actions[0].Condition) != 1 {
		t.Fatalf("expected 1 condition, got %d", len(parsed.Actions[0].Condition))
	}
	cond := parsed.Actions[0].Condition[0]
	if cond.Kind != "applicability" {
		t.Errorf("expected kind applicability, got %s", cond.Kind)
	}
	if cond.Expression != "%age > 40" {
		t.Errorf("expected expression '%%age > 40', got %s", cond.Expression)
	}
}

func TestParsePlanDefinition_WithRelatedActions(t *testing.T) {
	data := makePlanDefinitionJSON("pd-related", "active", "Related Plan", []map[string]interface{}{
		{
			"id":    "action-a",
			"title": "Action A",
			"relatedAction": []interface{}{
				map[string]interface{}{
					"actionId":     "action-b",
					"relationship": "before-start",
					"offsetDuration": map[string]interface{}{
						"value": 30,
						"unit":  "min",
					},
				},
			},
		},
		{
			"id":    "action-b",
			"title": "Action B",
		},
	})

	parsed, err := ParseApplyPlanDefinition(data)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(parsed.Actions[0].RelatedAction) != 1 {
		t.Fatalf("expected 1 related action, got %d", len(parsed.Actions[0].RelatedAction))
	}
	rel := parsed.Actions[0].RelatedAction[0]
	if rel.ActionID != "action-b" {
		t.Errorf("expected actionId action-b, got %s", rel.ActionID)
	}
	if rel.Relationship != "before-start" {
		t.Errorf("expected relationship before-start, got %s", rel.Relationship)
	}
	if rel.OffsetDuration == nil {
		t.Fatal("expected offsetDuration, got nil")
	}
}

func TestParsePlanDefinition_WithTiming(t *testing.T) {
	data := makePlanDefinitionJSON("pd-timing", "active", "Timed Plan", []map[string]interface{}{
		{
			"id":    "action-timed",
			"title": "Timed Action",
			"timingTiming": map[string]interface{}{
				"repeat": map[string]interface{}{
					"frequency": 1,
					"period":    1,
					"periodUnit": "d",
				},
			},
		},
	})

	parsed, err := ParseApplyPlanDefinition(data)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if parsed.Actions[0].Timing == nil {
		t.Error("expected timing to be set")
	}
}

// ============================================================================
// TestParseActivityDefinition
// ============================================================================

func TestParseActivityDefinition_MedicationRequest(t *testing.T) {
	data := map[string]interface{}{
		"resourceType": "ActivityDefinition",
		"id":           "ad-med",
		"status":       "active",
		"kind":         "MedicationRequest",
		"code": map[string]interface{}{
			"coding": []interface{}{
				map[string]interface{}{
					"system":  "http://rxnorm",
					"code":    "860975",
					"display": "Metformin 500mg",
				},
			},
		},
		"dosage": []interface{}{
			map[string]interface{}{"text": "500mg twice daily"},
		},
	}

	parsed, err := ParseApplyActivityDefinition(data)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if parsed.ID != "ad-med" {
		t.Errorf("expected ID ad-med, got %s", parsed.ID)
	}
	if parsed.Kind != "MedicationRequest" {
		t.Errorf("expected kind MedicationRequest, got %s", parsed.Kind)
	}
	if parsed.Status != "active" {
		t.Errorf("expected status active, got %s", parsed.Status)
	}
}

func TestParseActivityDefinition_ServiceRequest(t *testing.T) {
	data := makeActivityDefinitionJSON("ad-svc", "active", "ServiceRequest")
	data["code"] = map[string]interface{}{
		"coding": []interface{}{
			map[string]interface{}{"code": "4548-4", "display": "HbA1c"},
		},
	}

	parsed, err := ParseApplyActivityDefinition(data)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if parsed.Kind != "ServiceRequest" {
		t.Errorf("expected kind ServiceRequest, got %s", parsed.Kind)
	}
}

func TestParseActivityDefinition_Task(t *testing.T) {
	data := makeActivityDefinitionJSON("ad-task", "active", "Task")
	parsed, err := ParseApplyActivityDefinition(data)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if parsed.Kind != "Task" {
		t.Errorf("expected kind Task, got %s", parsed.Kind)
	}
}

func TestParseActivityDefinition_CommunicationRequest(t *testing.T) {
	data := makeActivityDefinitionJSON("ad-comm", "active", "CommunicationRequest")
	parsed, err := ParseApplyActivityDefinition(data)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if parsed.Kind != "CommunicationRequest" {
		t.Errorf("expected kind CommunicationRequest, got %s", parsed.Kind)
	}
}

func TestParseActivityDefinition_InvalidResourceType(t *testing.T) {
	data := map[string]interface{}{
		"resourceType": "Patient",
		"id":           "ad-bad",
	}
	_, err := ParseApplyActivityDefinition(data)
	if err == nil {
		t.Fatal("expected error for invalid resourceType, got nil")
	}
}

func TestParseActivityDefinition_NilData(t *testing.T) {
	_, err := ParseApplyActivityDefinition(nil)
	if err == nil {
		t.Fatal("expected error for nil data, got nil")
	}
}

func TestParseActivityDefinition_WithDynamicValue(t *testing.T) {
	data := makeActivityDefinitionJSON("ad-dv", "active", "ServiceRequest")
	data["dynamicValue"] = []interface{}{
		map[string]interface{}{
			"path": "status",
			"expression": map[string]interface{}{
				"expression": "'active'",
			},
		},
	}

	parsed, err := ParseApplyActivityDefinition(data)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(parsed.DynamicValues) != 1 {
		t.Fatalf("expected 1 dynamic value, got %d", len(parsed.DynamicValues))
	}
	if parsed.DynamicValues[0].Path != "status" {
		t.Errorf("expected path status, got %s", parsed.DynamicValues[0].Path)
	}
}

// ============================================================================
// TestValidateApplyRequest
// ============================================================================

func TestValidateApplyRequest_Valid(t *testing.T) {
	req := &ApplyRequest{
		PlanDefinitionID: "pd-1",
		Subject:          "Patient/123",
	}
	issues := ValidateApplyRequest(req)
	if len(issues) != 0 {
		t.Errorf("expected 0 issues, got %d: %v", len(issues), issues)
	}
}

func TestValidateApplyRequest_MissingSubject(t *testing.T) {
	req := &ApplyRequest{
		PlanDefinitionID: "pd-1",
	}
	issues := ValidateApplyRequest(req)
	if len(issues) == 0 {
		t.Fatal("expected validation issues for missing subject")
	}
	found := false
	for _, issue := range issues {
		if strings.Contains(issue.Diagnostics, "subject") {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected issue about missing subject")
	}
}

func TestValidateApplyRequest_MissingPlanDefinitionID(t *testing.T) {
	req := &ApplyRequest{
		Subject: "Patient/123",
	}
	issues := ValidateApplyRequest(req)
	if len(issues) == 0 {
		t.Fatal("expected validation issues for missing PlanDefinitionID")
	}
	found := false
	for _, issue := range issues {
		if strings.Contains(issue.Diagnostics, "PlanDefinition") || strings.Contains(issue.Diagnostics, "planDefinition") {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected issue about missing PlanDefinitionID")
	}
}

func TestValidateApplyRequest_MissingBothFields(t *testing.T) {
	req := &ApplyRequest{}
	issues := ValidateApplyRequest(req)
	if len(issues) < 2 {
		t.Errorf("expected at least 2 issues, got %d", len(issues))
	}
}

func TestValidateApplyRequest_InvalidSubjectFormat(t *testing.T) {
	req := &ApplyRequest{
		PlanDefinitionID: "pd-1",
		Subject:          "invalid-ref",
	}
	issues := ValidateApplyRequest(req)
	if len(issues) == 0 {
		t.Fatal("expected validation issues for invalid subject format")
	}
}

func TestValidateApplyRequest_NilRequest(t *testing.T) {
	issues := ValidateApplyRequest(nil)
	if len(issues) == 0 {
		t.Fatal("expected validation issues for nil request")
	}
}

// ============================================================================
// TestApplyPlanDefinition
// ============================================================================

func TestApplyPlanDefinition_Basic(t *testing.T) {
	planDef := &ParsedPlanDefinition{
		ID:     "pd-1",
		Status: "active",
		Title:  "Test Protocol",
		Actions: []ApplyPlanDefinitionAction{
			{
				ID:            "action-1",
				Title:         "Order Lab",
				Type:          "create",
				DefinitionURI: "ActivityDefinition/ad-lab",
			},
		},
	}

	req := &ApplyRequest{
		PlanDefinitionID: "pd-1",
		Subject:          "Patient/123",
	}

	result, err := ApplyPlanDefinition(planDef, req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result == nil {
		t.Fatal("expected result, got nil")
	}
	if result.CarePlan == nil {
		t.Fatal("expected CarePlan, got nil")
	}
	cp := result.CarePlan
	if cp["resourceType"] != "CarePlan" {
		t.Errorf("expected resourceType CarePlan, got %v", cp["resourceType"])
	}
	if cp["status"] != "active" {
		t.Errorf("expected status active, got %v", cp["status"])
	}
	if cp["intent"] != "plan" {
		t.Errorf("expected intent plan, got %v", cp["intent"])
	}

	// Check subject reference
	subj, ok := cp["subject"].(map[string]interface{})
	if !ok {
		t.Fatal("expected subject to be a map")
	}
	if subj["reference"] != "Patient/123" {
		t.Errorf("expected subject reference Patient/123, got %v", subj["reference"])
	}
}

func TestApplyPlanDefinition_GeneratesRequestGroup(t *testing.T) {
	planDef := &ParsedPlanDefinition{
		ID:     "pd-rg",
		Status: "active",
		Title:  "RG Protocol",
		Actions: []ApplyPlanDefinitionAction{
			{
				ID:    "action-1",
				Title: "First Action",
				Type:  "create",
			},
			{
				ID:    "action-2",
				Title: "Second Action",
				Type:  "create",
			},
		},
	}

	req := makeApplyRequest("Patient/456", "", "", "")
	req.PlanDefinitionID = "pd-rg"
	result, err := ApplyPlanDefinition(planDef, req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.RequestGroup == nil {
		t.Fatal("expected RequestGroup, got nil")
	}
	rg := result.RequestGroup
	if rg["resourceType"] != "RequestGroup" {
		t.Errorf("expected resourceType RequestGroup, got %v", rg["resourceType"])
	}

	// Check actions in request group
	actions, ok := rg["action"].([]interface{})
	if !ok {
		t.Fatal("expected action array in RequestGroup")
	}
	if len(actions) != 2 {
		t.Errorf("expected 2 actions, got %d", len(actions))
	}
}

func TestApplyPlanDefinition_WithConditions(t *testing.T) {
	planDef := &ParsedPlanDefinition{
		ID:     "pd-cond",
		Status: "active",
		Title:  "Conditional Protocol",
		Actions: []ApplyPlanDefinitionAction{
			{
				ID:    "action-applicable",
				Title: "Applicable Action",
				Type:  "create",
				Condition: []ApplyActionCondition{
					{Kind: "applicability", Expression: "true"},
				},
			},
			{
				ID:    "action-not-applicable",
				Title: "Not Applicable Action",
				Type:  "create",
				Condition: []ApplyActionCondition{
					{Kind: "applicability", Expression: "false"},
				},
			},
		},
	}

	req := makeApplyRequest("Patient/789", "", "", "")
	req.PlanDefinitionID = "pd-cond"
	result, err := ApplyPlanDefinition(planDef, req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Only the applicable action should appear in the request group
	rg := result.RequestGroup
	actions, ok := rg["action"].([]interface{})
	if !ok {
		t.Fatal("expected action array in RequestGroup")
	}
	if len(actions) != 1 {
		t.Errorf("expected 1 applicable action, got %d", len(actions))
	}
	if len(actions) > 0 {
		firstAction, _ := actions[0].(map[string]interface{})
		if firstAction["id"] != "action-applicable" {
			t.Errorf("expected action-applicable, got %v", firstAction["id"])
		}
	}
}

func TestApplyPlanDefinition_WithNestedActions(t *testing.T) {
	planDef := &ParsedPlanDefinition{
		ID:     "pd-nested",
		Status: "active",
		Title:  "Nested Protocol",
		Actions: []ApplyPlanDefinitionAction{
			{
				ID:                "parent",
				Title:             "Parent",
				SelectionBehavior: "any",
				Action: []ApplyPlanDefinitionAction{
					{ID: "child-1", Title: "Child 1", Type: "create"},
					{ID: "child-2", Title: "Child 2", Type: "create"},
				},
			},
		},
	}

	req := makeApplyRequest("Patient/100", "", "", "")
	req.PlanDefinitionID = "pd-nested"
	result, err := ApplyPlanDefinition(planDef, req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	rg := result.RequestGroup
	actions, _ := rg["action"].([]interface{})
	if len(actions) != 1 {
		t.Fatalf("expected 1 top-level action, got %d", len(actions))
	}

	parentAction, _ := actions[0].(map[string]interface{})
	childActions, ok := parentAction["action"].([]interface{})
	if !ok {
		t.Fatal("expected nested action array")
	}
	if len(childActions) != 2 {
		t.Errorf("expected 2 child actions, got %d", len(childActions))
	}
}

func TestApplyPlanDefinition_WithRelatedActions(t *testing.T) {
	planDef := &ParsedPlanDefinition{
		ID:     "pd-related",
		Status: "active",
		Title:  "Related Protocol",
		Actions: []ApplyPlanDefinitionAction{
			{
				ID:    "action-a",
				Title: "Action A",
				Type:  "create",
			},
			{
				ID:    "action-b",
				Title: "Action B",
				Type:  "create",
				RelatedAction: []ApplyRelatedAction{
					{
						ActionID:     "action-a",
						Relationship: "after-start",
					},
				},
			},
		},
	}

	req := makeApplyRequest("Patient/200", "", "", "")
	req.PlanDefinitionID = "pd-related"
	result, err := ApplyPlanDefinition(planDef, req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	rg := result.RequestGroup
	actions, _ := rg["action"].([]interface{})
	if len(actions) != 2 {
		t.Fatalf("expected 2 actions, got %d", len(actions))
	}

	// Check second action has relatedAction
	secondAction, _ := actions[1].(map[string]interface{})
	relatedActions, ok := secondAction["relatedAction"].([]interface{})
	if !ok {
		t.Fatal("expected relatedAction in second action")
	}
	if len(relatedActions) != 1 {
		t.Fatalf("expected 1 related action, got %d", len(relatedActions))
	}
	relAction, _ := relatedActions[0].(map[string]interface{})
	if relAction["actionId"] != "action-a" {
		t.Errorf("expected actionId action-a, got %v", relAction["actionId"])
	}
	if relAction["relationship"] != "after-start" {
		t.Errorf("expected relationship after-start, got %v", relAction["relationship"])
	}
}

func TestApplyPlanDefinition_WithTiming(t *testing.T) {
	timing := map[string]interface{}{
		"repeat": map[string]interface{}{
			"frequency":  1,
			"period":     1,
			"periodUnit": "d",
		},
	}
	planDef := &ParsedPlanDefinition{
		ID:     "pd-timed",
		Status: "active",
		Title:  "Timed Protocol",
		Actions: []ApplyPlanDefinitionAction{
			{
				ID:     "action-timed",
				Title:  "Timed Action",
				Type:   "create",
				Timing: timing,
			},
		},
	}

	req := makeApplyRequest("Patient/300", "", "", "")
	req.PlanDefinitionID = "pd-timed"
	result, err := ApplyPlanDefinition(planDef, req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	rg := result.RequestGroup
	actions, _ := rg["action"].([]interface{})
	if len(actions) != 1 {
		t.Fatalf("expected 1 action, got %d", len(actions))
	}
	action, _ := actions[0].(map[string]interface{})
	if action["timingTiming"] == nil {
		t.Error("expected timing in action")
	}
}

func TestApplyPlanDefinition_RetiredPlan(t *testing.T) {
	planDef := &ParsedPlanDefinition{
		ID:     "pd-retired",
		Status: "retired",
		Title:  "Retired Protocol",
	}

	req := makeApplyRequest("Patient/400", "", "", "")
	req.PlanDefinitionID = "pd-retired"
	_, err := ApplyPlanDefinition(planDef, req)
	if err == nil {
		t.Fatal("expected error for retired plan, got nil")
	}
}

func TestApplyPlanDefinition_NilPlan(t *testing.T) {
	req := makeApplyRequest("Patient/400", "", "", "")
	_, err := ApplyPlanDefinition(nil, req)
	if err == nil {
		t.Fatal("expected error for nil plan, got nil")
	}
}

func TestApplyPlanDefinition_NilRequest(t *testing.T) {
	planDef := &ParsedPlanDefinition{
		ID:     "pd-1",
		Status: "active",
	}
	_, err := ApplyPlanDefinition(planDef, nil)
	if err == nil {
		t.Fatal("expected error for nil request, got nil")
	}
}

func TestApplyPlanDefinition_WithEncounterContext(t *testing.T) {
	planDef := &ParsedPlanDefinition{
		ID:     "pd-enc",
		Status: "active",
		Title:  "Encounter Protocol",
		Actions: []ApplyPlanDefinitionAction{
			{ID: "a1", Title: "Action 1", Type: "create"},
		},
	}

	req := &ApplyRequest{
		PlanDefinitionID: "pd-enc",
		Subject:          "Patient/500",
		Encounter:        "Encounter/enc-99",
	}
	result, err := ApplyPlanDefinition(planDef, req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	cp := result.CarePlan
	enc, ok := cp["encounter"].(map[string]interface{})
	if !ok {
		t.Fatal("expected encounter in CarePlan")
	}
	if enc["reference"] != "Encounter/enc-99" {
		t.Errorf("expected encounter reference Encounter/enc-99, got %v", enc["reference"])
	}
}

func TestApplyPlanDefinition_WithPractitionerContext(t *testing.T) {
	planDef := &ParsedPlanDefinition{
		ID:     "pd-pract",
		Status: "active",
		Title:  "Practitioner Protocol",
		Actions: []ApplyPlanDefinitionAction{
			{ID: "a1", Title: "Action 1", Type: "create"},
		},
	}

	req := &ApplyRequest{
		PlanDefinitionID: "pd-pract",
		Subject:          "Patient/500",
		Practitioner:     "Practitioner/pract-55",
	}
	result, err := ApplyPlanDefinition(planDef, req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	cp := result.CarePlan
	author, ok := cp["author"].(map[string]interface{})
	if !ok {
		t.Fatal("expected author in CarePlan")
	}
	if author["reference"] != "Practitioner/pract-55" {
		t.Errorf("expected author reference Practitioner/pract-55, got %v", author["reference"])
	}
}

// ============================================================================
// TestApplyActivityDefinition
// ============================================================================

func TestApplyActivityDefinition_MedicationRequest(t *testing.T) {
	actDef := &ParsedActivityDefinition{
		ID:     "ad-med",
		Status: "active",
		Kind:   "MedicationRequest",
		Code: map[string]interface{}{
			"coding": []interface{}{
				map[string]interface{}{"code": "860975", "display": "Metformin"},
			},
		},
		Dosage: []interface{}{
			map[string]interface{}{"text": "500mg twice daily"},
		},
	}

	req := makeApplyRequest("Patient/123", "Encounter/enc-1", "Practitioner/pract-1", "")
	activity, err := ApplyActivityDefinition(actDef, req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if activity.ResourceType != "MedicationRequest" {
		t.Errorf("expected resourceType MedicationRequest, got %s", activity.ResourceType)
	}
	resource := activity.Resource
	if resource["resourceType"] != "MedicationRequest" {
		t.Errorf("expected resourceType MedicationRequest, got %v", resource["resourceType"])
	}
	if resource["intent"] != "order" {
		t.Errorf("expected intent order, got %v", resource["intent"])
	}

	subj, _ := resource["subject"].(map[string]interface{})
	if subj["reference"] != "Patient/123" {
		t.Errorf("expected subject Patient/123, got %v", subj["reference"])
	}
}

func TestApplyActivityDefinition_ServiceRequest(t *testing.T) {
	actDef := &ParsedActivityDefinition{
		ID:     "ad-svc",
		Status: "active",
		Kind:   "ServiceRequest",
		Code: map[string]interface{}{
			"coding": []interface{}{
				map[string]interface{}{"code": "4548-4", "display": "HbA1c"},
			},
		},
	}

	req := makeApplyRequest("Patient/456", "", "", "")
	activity, err := ApplyActivityDefinition(actDef, req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if activity.ResourceType != "ServiceRequest" {
		t.Errorf("expected resourceType ServiceRequest, got %s", activity.ResourceType)
	}
	resource := activity.Resource
	if resource["intent"] != "order" {
		t.Errorf("expected intent order, got %v", resource["intent"])
	}
}

func TestApplyActivityDefinition_Task(t *testing.T) {
	actDef := &ParsedActivityDefinition{
		ID:     "ad-task",
		Status: "active",
		Kind:   "Task",
	}

	req := makeApplyRequest("Patient/789", "", "", "")
	activity, err := ApplyActivityDefinition(actDef, req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if activity.ResourceType != "Task" {
		t.Errorf("expected resourceType Task, got %s", activity.ResourceType)
	}
	resource := activity.Resource
	if resource["intent"] != "order" {
		t.Errorf("expected intent order, got %v", resource["intent"])
	}
	// Task uses "for" for subject reference
	forRef, ok := resource["for"].(map[string]interface{})
	if !ok {
		t.Fatal("expected 'for' reference in Task")
	}
	if forRef["reference"] != "Patient/789" {
		t.Errorf("expected for reference Patient/789, got %v", forRef["reference"])
	}
}

func TestApplyActivityDefinition_CommunicationRequest(t *testing.T) {
	actDef := &ParsedActivityDefinition{
		ID:     "ad-comm",
		Status: "active",
		Kind:   "CommunicationRequest",
	}

	req := makeApplyRequest("Patient/100", "", "", "")
	activity, err := ApplyActivityDefinition(actDef, req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if activity.ResourceType != "CommunicationRequest" {
		t.Errorf("expected resourceType CommunicationRequest, got %s", activity.ResourceType)
	}
}

func TestApplyActivityDefinition_NilDefinition(t *testing.T) {
	req := makeApplyRequest("Patient/100", "", "", "")
	_, err := ApplyActivityDefinition(nil, req)
	if err == nil {
		t.Fatal("expected error for nil definition, got nil")
	}
}

func TestApplyActivityDefinition_NilRequest(t *testing.T) {
	actDef := &ParsedActivityDefinition{
		ID:     "ad-1",
		Status: "active",
		Kind:   "ServiceRequest",
	}
	_, err := ApplyActivityDefinition(actDef, nil)
	if err == nil {
		t.Fatal("expected error for nil request, got nil")
	}
}

func TestApplyActivityDefinition_RetiredDefinition(t *testing.T) {
	actDef := &ParsedActivityDefinition{
		ID:     "ad-retired",
		Status: "retired",
		Kind:   "ServiceRequest",
	}

	req := makeApplyRequest("Patient/100", "", "", "")
	_, err := ApplyActivityDefinition(actDef, req)
	if err == nil {
		t.Fatal("expected error for retired activity definition, got nil")
	}
}

func TestApplyActivityDefinition_WithEncounter(t *testing.T) {
	actDef := &ParsedActivityDefinition{
		ID:     "ad-enc",
		Status: "active",
		Kind:   "ServiceRequest",
	}

	req := makeApplyRequest("Patient/100", "Encounter/enc-55", "", "")
	activity, err := ApplyActivityDefinition(actDef, req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	resource := activity.Resource
	enc, ok := resource["encounter"].(map[string]interface{})
	if !ok {
		t.Fatal("expected encounter reference")
	}
	if enc["reference"] != "Encounter/enc-55" {
		t.Errorf("expected encounter reference Encounter/enc-55, got %v", enc["reference"])
	}
}

func TestApplyActivityDefinition_WithPractitioner(t *testing.T) {
	actDef := &ParsedActivityDefinition{
		ID:     "ad-pract",
		Status: "active",
		Kind:   "MedicationRequest",
	}

	req := makeApplyRequest("Patient/100", "", "Practitioner/pract-77", "")
	activity, err := ApplyActivityDefinition(actDef, req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	resource := activity.Resource
	requester, ok := resource["requester"].(map[string]interface{})
	if !ok {
		t.Fatal("expected requester reference in MedicationRequest")
	}
	if requester["reference"] != "Practitioner/pract-77" {
		t.Errorf("expected requester reference Practitioner/pract-77, got %v", requester["reference"])
	}
}

func TestApplyActivityDefinition_GeneratesUniqueID(t *testing.T) {
	actDef := &ParsedActivityDefinition{
		ID:     "ad-id",
		Status: "active",
		Kind:   "ServiceRequest",
	}

	req := makeApplyRequest("Patient/100", "", "", "")
	activity1, _ := ApplyActivityDefinition(actDef, req)
	activity2, _ := ApplyActivityDefinition(actDef, req)

	id1 := activity1.Resource["id"].(string)
	id2 := activity2.Resource["id"].(string)
	if id1 == "" {
		t.Error("expected non-empty resource ID")
	}
	if id1 == id2 {
		t.Error("expected unique IDs for each application")
	}
}

// ============================================================================
// TestActionConditionEvaluation
// ============================================================================

func TestEvaluateApplyConditions_AllApplicable(t *testing.T) {
	conditions := []ApplyActionCondition{
		{Kind: "applicability", Expression: "true"},
	}
	if !evaluateApplyConditions(conditions) {
		t.Error("expected conditions to be applicable")
	}
}

func TestEvaluateApplyConditions_NotApplicable(t *testing.T) {
	conditions := []ApplyActionCondition{
		{Kind: "applicability", Expression: "false"},
	}
	if evaluateApplyConditions(conditions) {
		t.Error("expected conditions to NOT be applicable")
	}
}

func TestEvaluateApplyConditions_Empty(t *testing.T) {
	if !evaluateApplyConditions(nil) {
		t.Error("expected empty conditions to be applicable")
	}
}

func TestEvaluateApplyConditions_StartStopIgnored(t *testing.T) {
	conditions := []ApplyActionCondition{
		{Kind: "start", Expression: "false"},
		{Kind: "stop", Expression: "false"},
	}
	// start and stop conditions should not block applicability
	if !evaluateApplyConditions(conditions) {
		t.Error("expected start/stop conditions to be ignored for applicability")
	}
}

func TestEvaluateApplyConditions_MultipleApplicability(t *testing.T) {
	conditions := []ApplyActionCondition{
		{Kind: "applicability", Expression: "true"},
		{Kind: "applicability", Expression: "false"},
	}
	// All applicability conditions must be true
	if evaluateApplyConditions(conditions) {
		t.Error("expected multiple conditions with one false to be not applicable")
	}
}

// ============================================================================
// TestSelectionBehavior
// ============================================================================

func TestSelectionBehavior_Any(t *testing.T) {
	planDef := &ParsedPlanDefinition{
		ID:     "pd-sel-any",
		Status: "active",
		Actions: []ApplyPlanDefinitionAction{
			{
				ID:                "group",
				Title:             "Selection Group",
				SelectionBehavior: "any",
				Action: []ApplyPlanDefinitionAction{
					{ID: "opt-1", Title: "Option 1", Type: "create"},
					{ID: "opt-2", Title: "Option 2", Type: "create"},
				},
			},
		},
	}

	req := makeApplyRequest("Patient/sel", "", "", "")
	req.PlanDefinitionID = "pd-sel-any"
	result, err := ApplyPlanDefinition(planDef, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	rg := result.RequestGroup
	actions, _ := rg["action"].([]interface{})
	if len(actions) != 1 {
		t.Fatalf("expected 1 group action, got %d", len(actions))
	}
	groupAction, _ := actions[0].(map[string]interface{})
	if groupAction["selectionBehavior"] != "any" {
		t.Errorf("expected selectionBehavior any, got %v", groupAction["selectionBehavior"])
	}
}

func TestSelectionBehavior_All(t *testing.T) {
	planDef := &ParsedPlanDefinition{
		ID:     "pd-sel-all",
		Status: "active",
		Actions: []ApplyPlanDefinitionAction{
			{
				ID:                "group",
				SelectionBehavior: "all",
				Action: []ApplyPlanDefinitionAction{
					{ID: "opt-1", Title: "Option 1"},
					{ID: "opt-2", Title: "Option 2"},
				},
			},
		},
	}

	req := makeApplyRequest("Patient/sel", "", "", "")
	req.PlanDefinitionID = "pd-sel-all"
	result, _ := ApplyPlanDefinition(planDef, req)
	rg := result.RequestGroup
	actions, _ := rg["action"].([]interface{})
	groupAction, _ := actions[0].(map[string]interface{})
	if groupAction["selectionBehavior"] != "all" {
		t.Errorf("expected selectionBehavior all, got %v", groupAction["selectionBehavior"])
	}
}

func TestSelectionBehavior_ExactlyOne(t *testing.T) {
	planDef := &ParsedPlanDefinition{
		ID:     "pd-sel-one",
		Status: "active",
		Actions: []ApplyPlanDefinitionAction{
			{
				ID:                "group",
				SelectionBehavior: "exactly-one",
				Action: []ApplyPlanDefinitionAction{
					{ID: "opt-1", Title: "Option 1"},
					{ID: "opt-2", Title: "Option 2"},
				},
			},
		},
	}

	req := makeApplyRequest("Patient/sel", "", "", "")
	req.PlanDefinitionID = "pd-sel-one"
	result, _ := ApplyPlanDefinition(planDef, req)
	rg := result.RequestGroup
	actions, _ := rg["action"].([]interface{})
	groupAction, _ := actions[0].(map[string]interface{})
	if groupAction["selectionBehavior"] != "exactly-one" {
		t.Errorf("expected selectionBehavior exactly-one, got %v", groupAction["selectionBehavior"])
	}
}

func TestSelectionBehavior_AtMostOne(t *testing.T) {
	planDef := &ParsedPlanDefinition{
		ID:     "pd-sel-atmost",
		Status: "active",
		Actions: []ApplyPlanDefinitionAction{
			{
				ID:                "group",
				SelectionBehavior: "at-most-one",
				Action: []ApplyPlanDefinitionAction{
					{ID: "opt-1", Title: "Option 1"},
				},
			},
		},
	}

	req := makeApplyRequest("Patient/sel", "", "", "")
	req.PlanDefinitionID = "pd-sel-atmost"
	result, _ := ApplyPlanDefinition(planDef, req)
	rg := result.RequestGroup
	actions, _ := rg["action"].([]interface{})
	groupAction, _ := actions[0].(map[string]interface{})
	if groupAction["selectionBehavior"] != "at-most-one" {
		t.Errorf("expected selectionBehavior at-most-one, got %v", groupAction["selectionBehavior"])
	}
}

func TestSelectionBehavior_OneOrMore(t *testing.T) {
	planDef := &ParsedPlanDefinition{
		ID:     "pd-sel-oneormore",
		Status: "active",
		Actions: []ApplyPlanDefinitionAction{
			{
				ID:                "group",
				SelectionBehavior: "one-or-more",
				Action: []ApplyPlanDefinitionAction{
					{ID: "opt-1", Title: "Option 1"},
				},
			},
		},
	}

	req := makeApplyRequest("Patient/sel", "", "", "")
	req.PlanDefinitionID = "pd-sel-oneormore"
	result, _ := ApplyPlanDefinition(planDef, req)
	rg := result.RequestGroup
	actions, _ := rg["action"].([]interface{})
	groupAction, _ := actions[0].(map[string]interface{})
	if groupAction["selectionBehavior"] != "one-or-more" {
		t.Errorf("expected selectionBehavior one-or-more, got %v", groupAction["selectionBehavior"])
	}
}

func TestSelectionBehavior_AllOrNone(t *testing.T) {
	planDef := &ParsedPlanDefinition{
		ID:     "pd-sel-allornone",
		Status: "active",
		Actions: []ApplyPlanDefinitionAction{
			{
				ID:                "group",
				SelectionBehavior: "all-or-none",
				Action: []ApplyPlanDefinitionAction{
					{ID: "opt-1", Title: "Option 1"},
				},
			},
		},
	}

	req := makeApplyRequest("Patient/sel", "", "", "")
	req.PlanDefinitionID = "pd-sel-allornone"
	result, _ := ApplyPlanDefinition(planDef, req)
	rg := result.RequestGroup
	actions, _ := rg["action"].([]interface{})
	groupAction, _ := actions[0].(map[string]interface{})
	if groupAction["selectionBehavior"] != "all-or-none" {
		t.Errorf("expected selectionBehavior all-or-none, got %v", groupAction["selectionBehavior"])
	}
}

// ============================================================================
// TestRelatedActionResolution
// ============================================================================

func TestRelatedActionResolution_BeforeStart(t *testing.T) {
	planDef := &ParsedPlanDefinition{
		ID:     "pd-rel",
		Status: "active",
		Actions: []ApplyPlanDefinitionAction{
			{ID: "first", Title: "First"},
			{
				ID: "second", Title: "Second",
				RelatedAction: []ApplyRelatedAction{
					{ActionID: "first", Relationship: "before-start"},
				},
			},
		},
	}

	req := makeApplyRequest("Patient/rel", "", "", "")
	req.PlanDefinitionID = "pd-rel"
	result, err := ApplyPlanDefinition(planDef, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	rg := result.RequestGroup
	actions, _ := rg["action"].([]interface{})
	if len(actions) != 2 {
		t.Fatalf("expected 2 actions, got %d", len(actions))
	}
	secondAction, _ := actions[1].(map[string]interface{})
	relatedActions, _ := secondAction["relatedAction"].([]interface{})
	if len(relatedActions) != 1 {
		t.Fatalf("expected 1 related action, got %d", len(relatedActions))
	}
	rel, _ := relatedActions[0].(map[string]interface{})
	if rel["relationship"] != "before-start" {
		t.Errorf("expected before-start, got %v", rel["relationship"])
	}
}

func TestRelatedActionResolution_WithOffset(t *testing.T) {
	offset := map[string]interface{}{
		"value": 30,
		"unit":  "min",
	}
	planDef := &ParsedPlanDefinition{
		ID:     "pd-rel-off",
		Status: "active",
		Actions: []ApplyPlanDefinitionAction{
			{ID: "first", Title: "First"},
			{
				ID: "second", Title: "Second",
				RelatedAction: []ApplyRelatedAction{
					{ActionID: "first", Relationship: "after", OffsetDuration: offset},
				},
			},
		},
	}

	req := makeApplyRequest("Patient/rel", "", "", "")
	req.PlanDefinitionID = "pd-rel-off"
	result, _ := ApplyPlanDefinition(planDef, req)
	rg := result.RequestGroup
	actions, _ := rg["action"].([]interface{})
	secondAction, _ := actions[1].(map[string]interface{})
	relatedActions, _ := secondAction["relatedAction"].([]interface{})
	rel, _ := relatedActions[0].(map[string]interface{})
	if rel["offsetDuration"] == nil {
		t.Error("expected offsetDuration in related action")
	}
}

// ============================================================================
// TestReferenceResolution
// ============================================================================

func TestReferenceResolution_PatientReplacement(t *testing.T) {
	actDef := &ParsedActivityDefinition{
		ID: "ad-reftest", Status: "active", Kind: "ServiceRequest",
	}
	req := makeApplyRequest("Patient/ref-pat", "", "", "")
	activity, _ := ApplyActivityDefinition(actDef, req)
	subj, _ := activity.Resource["subject"].(map[string]interface{})
	if subj["reference"] != "Patient/ref-pat" {
		t.Errorf("expected Patient/ref-pat, got %v", subj["reference"])
	}
}

func TestReferenceResolution_EncounterReplacement(t *testing.T) {
	actDef := &ParsedActivityDefinition{
		ID: "ad-reftest", Status: "active", Kind: "ServiceRequest",
	}
	req := makeApplyRequest("Patient/ref-pat", "Encounter/ref-enc", "", "")
	activity, _ := ApplyActivityDefinition(actDef, req)
	enc, _ := activity.Resource["encounter"].(map[string]interface{})
	if enc["reference"] != "Encounter/ref-enc" {
		t.Errorf("expected Encounter/ref-enc, got %v", enc["reference"])
	}
}

func TestReferenceResolution_PractitionerReplacement(t *testing.T) {
	actDef := &ParsedActivityDefinition{
		ID: "ad-reftest", Status: "active", Kind: "MedicationRequest",
	}
	req := makeApplyRequest("Patient/ref-pat", "", "Practitioner/ref-pract", "")
	activity, _ := ApplyActivityDefinition(actDef, req)
	requester, _ := activity.Resource["requester"].(map[string]interface{})
	if requester["reference"] != "Practitioner/ref-pract" {
		t.Errorf("expected Practitioner/ref-pract, got %v", requester["reference"])
	}
}

func TestReferenceResolution_OrganizationReplacement(t *testing.T) {
	actDef := &ParsedActivityDefinition{
		ID: "ad-reftest", Status: "active", Kind: "ServiceRequest",
	}
	req := makeApplyRequest("Patient/ref-pat", "", "", "Organization/ref-org")
	activity, _ := ApplyActivityDefinition(actDef, req)
	performer, _ := activity.Resource["performer"].(map[string]interface{})
	if performer == nil {
		t.Fatal("expected performer reference for organization")
	}
	if performer["reference"] != "Organization/ref-org" {
		t.Errorf("expected Organization/ref-org, got %v", performer["reference"])
	}
}

// ============================================================================
// TestApplyHandler (HTTP)
// ============================================================================

func TestApplyHandler_Success(t *testing.T) {
	resolver := newMockApplyResolver()
	resolver.resources["PlanDefinition/pd-1"] = makePlanDefinitionJSON("pd-1", "active", "Test Plan", []map[string]interface{}{
		{"id": "a1", "title": "Action 1"},
	})

	e := echo.New()
	body := `{"subject": "Patient/123"}`
	req := httptest.NewRequest(http.MethodPost, "/fhir/PlanDefinition/pd-1/$apply", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("pd-1")

	handler := ApplyHandler(resolver)
	err := handler(c)
	if err != nil {
		t.Fatalf("handler returned error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if result["resourceType"] != "Bundle" {
		t.Errorf("expected Bundle, got %v", result["resourceType"])
	}
}

func TestApplyHandler_NotFound(t *testing.T) {
	resolver := newMockApplyResolver()

	e := echo.New()
	body := `{"subject": "Patient/123"}`
	req := httptest.NewRequest(http.MethodPost, "/fhir/PlanDefinition/nonexistent/$apply", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("nonexistent")

	handler := ApplyHandler(resolver)
	err := handler(c)
	if err != nil {
		t.Fatalf("handler returned error: %v", err)
	}
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestApplyHandler_InvalidRequest(t *testing.T) {
	resolver := newMockApplyResolver()
	resolver.resources["PlanDefinition/pd-1"] = makePlanDefinitionJSON("pd-1", "active", "Test Plan", nil)

	e := echo.New()
	// Empty body - missing subject
	body := `{}`
	req := httptest.NewRequest(http.MethodPost, "/fhir/PlanDefinition/pd-1/$apply", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("pd-1")

	handler := ApplyHandler(resolver)
	err := handler(c)
	if err != nil {
		t.Fatalf("handler returned error: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestApplyHandler_InvalidJSON(t *testing.T) {
	resolver := newMockApplyResolver()
	resolver.resources["PlanDefinition/pd-1"] = makePlanDefinitionJSON("pd-1", "active", "Test Plan", nil)

	e := echo.New()
	body := `{invalid json`
	req := httptest.NewRequest(http.MethodPost, "/fhir/PlanDefinition/pd-1/$apply", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("pd-1")

	handler := ApplyHandler(resolver)
	err := handler(c)
	if err != nil {
		t.Fatalf("handler returned error: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestApplyHandler_EmptyBody(t *testing.T) {
	resolver := newMockApplyResolver()
	resolver.resources["PlanDefinition/pd-1"] = makePlanDefinitionJSON("pd-1", "active", "Test Plan", nil)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/fhir/PlanDefinition/pd-1/$apply", strings.NewReader(""))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("pd-1")

	handler := ApplyHandler(resolver)
	err := handler(c)
	if err != nil {
		t.Fatalf("handler returned error: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestApplyHandler_WithParametersResource(t *testing.T) {
	resolver := newMockApplyResolver()
	resolver.resources["PlanDefinition/pd-1"] = makePlanDefinitionJSON("pd-1", "active", "Test Plan", []map[string]interface{}{
		{"id": "a1", "title": "Action 1"},
	})

	e := echo.New()
	body := `{
		"resourceType": "Parameters",
		"parameter": [
			{"name": "subject", "valueString": "Patient/123"},
			{"name": "encounter", "valueString": "Encounter/enc-1"},
			{"name": "practitioner", "valueString": "Practitioner/pract-1"}
		]
	}`
	req := httptest.NewRequest(http.MethodPost, "/fhir/PlanDefinition/pd-1/$apply", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("pd-1")

	handler := ApplyHandler(resolver)
	err := handler(c)
	if err != nil {
		t.Fatalf("handler returned error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

// ============================================================================
// TestActivityDefinitionApplyHandler (HTTP)
// ============================================================================

func TestActivityDefinitionApplyHandler_Success(t *testing.T) {
	resolver := newMockApplyResolver()
	resolver.resources["ActivityDefinition/ad-1"] = makeActivityDefinitionJSON("ad-1", "active", "ServiceRequest")

	e := echo.New()
	body := `{"subject": "Patient/123"}`
	req := httptest.NewRequest(http.MethodPost, "/fhir/ActivityDefinition/ad-1/$apply", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("ad-1")

	handler := ActivityDefinitionApplyHandler(resolver)
	err := handler(c)
	if err != nil {
		t.Fatalf("handler returned error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var result map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if result["resourceType"] != "ServiceRequest" {
		t.Errorf("expected ServiceRequest, got %v", result["resourceType"])
	}
}

func TestActivityDefinitionApplyHandler_NotFound(t *testing.T) {
	resolver := newMockApplyResolver()

	e := echo.New()
	body := `{"subject": "Patient/123"}`
	req := httptest.NewRequest(http.MethodPost, "/fhir/ActivityDefinition/nonexistent/$apply", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("nonexistent")

	handler := ActivityDefinitionApplyHandler(resolver)
	err := handler(c)
	if err != nil {
		t.Fatalf("handler returned error: %v", err)
	}
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestActivityDefinitionApplyHandler_InvalidJSON(t *testing.T) {
	resolver := newMockApplyResolver()
	resolver.resources["ActivityDefinition/ad-1"] = makeActivityDefinitionJSON("ad-1", "active", "ServiceRequest")

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/fhir/ActivityDefinition/ad-1/$apply", strings.NewReader("{bad"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("ad-1")

	handler := ActivityDefinitionApplyHandler(resolver)
	err := handler(c)
	if err != nil {
		t.Fatalf("handler returned error: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestActivityDefinitionApplyHandler_MissingSubject(t *testing.T) {
	resolver := newMockApplyResolver()
	resolver.resources["ActivityDefinition/ad-1"] = makeActivityDefinitionJSON("ad-1", "active", "ServiceRequest")

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/fhir/ActivityDefinition/ad-1/$apply", strings.NewReader("{}"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("ad-1")

	handler := ActivityDefinitionApplyHandler(resolver)
	err := handler(c)
	if err != nil {
		t.Fatalf("handler returned error: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

// ============================================================================
// Edge Cases
// ============================================================================

func TestApplyPlanDefinition_DeeplyNestedActions(t *testing.T) {
	planDef := &ParsedPlanDefinition{
		ID:     "pd-deep",
		Status: "active",
		Actions: []ApplyPlanDefinitionAction{
			{
				ID: "level-1",
				Action: []ApplyPlanDefinitionAction{
					{
						ID: "level-2",
						Action: []ApplyPlanDefinitionAction{
							{
								ID: "level-3",
								Action: []ApplyPlanDefinitionAction{
									{ID: "level-4", Title: "Deepest Action"},
								},
							},
						},
					},
				},
			},
		},
	}

	req := makeApplyRequest("Patient/deep", "", "", "")
	req.PlanDefinitionID = "pd-deep"
	result, err := ApplyPlanDefinition(planDef, req)
	if err != nil {
		t.Fatalf("expected no error for deeply nested actions, got %v", err)
	}
	if result.RequestGroup == nil {
		t.Fatal("expected RequestGroup")
	}
}

func TestApplyPlanDefinition_EmptyActions(t *testing.T) {
	planDef := &ParsedPlanDefinition{
		ID:      "pd-empty",
		Status:  "active",
		Actions: []ApplyPlanDefinitionAction{},
	}

	req := makeApplyRequest("Patient/empty", "", "", "")
	req.PlanDefinitionID = "pd-empty"
	result, err := ApplyPlanDefinition(planDef, req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.CarePlan == nil {
		t.Error("expected CarePlan even with empty actions")
	}
}

func TestApplyPlanDefinition_ActionWithDescription(t *testing.T) {
	planDef := &ParsedPlanDefinition{
		ID:     "pd-desc",
		Status: "active",
		Actions: []ApplyPlanDefinitionAction{
			{
				ID:          "a1",
				Title:       "Action Title",
				Description: "Detailed description of the action",
				Type:        "create",
			},
		},
	}

	req := makeApplyRequest("Patient/desc", "", "", "")
	req.PlanDefinitionID = "pd-desc"
	result, _ := ApplyPlanDefinition(planDef, req)
	rg := result.RequestGroup
	actions, _ := rg["action"].([]interface{})
	action, _ := actions[0].(map[string]interface{})
	if action["description"] != "Detailed description of the action" {
		t.Errorf("expected description, got %v", action["description"])
	}
}

func TestApplyPlanDefinition_ActionPriority(t *testing.T) {
	priorities := []string{"routine", "urgent", "asap", "stat"}
	for _, priority := range priorities {
		t.Run(priority, func(t *testing.T) {
			planDef := &ParsedPlanDefinition{
				ID:     "pd-pri-" + priority,
				Status: "active",
				Actions: []ApplyPlanDefinitionAction{
					{
						ID:       "a1",
						Title:    "Priority Action",
						Priority: priority,
					},
				},
			}

			req := makeApplyRequest("Patient/pri", "", "", "")
			req.PlanDefinitionID = planDef.ID
			result, _ := ApplyPlanDefinition(planDef, req)
			rg := result.RequestGroup
			actions, _ := rg["action"].([]interface{})
			action, _ := actions[0].(map[string]interface{})
			if action["priority"] != priority {
				t.Errorf("expected priority %s, got %v", priority, action["priority"])
			}
		})
	}
}

func TestApplyPlanDefinition_GroupingBehavior(t *testing.T) {
	behaviors := []string{"visual-group", "logical-group", "sentence-group"}
	for _, behavior := range behaviors {
		t.Run(behavior, func(t *testing.T) {
			planDef := &ParsedPlanDefinition{
				ID:     "pd-grp-" + behavior,
				Status: "active",
				Actions: []ApplyPlanDefinitionAction{
					{
						ID:               "a1",
						GroupingBehavior: behavior,
					},
				},
			}

			req := makeApplyRequest("Patient/grp", "", "", "")
			req.PlanDefinitionID = planDef.ID
			result, _ := ApplyPlanDefinition(planDef, req)
			rg := result.RequestGroup
			actions, _ := rg["action"].([]interface{})
			action, _ := actions[0].(map[string]interface{})
			if action["groupingBehavior"] != behavior {
				t.Errorf("expected groupingBehavior %s, got %v", behavior, action["groupingBehavior"])
			}
		})
	}
}

func TestApplyPlanDefinition_CarePlanHasID(t *testing.T) {
	planDef := &ParsedPlanDefinition{
		ID:     "pd-cpid",
		Status: "active",
	}

	req := makeApplyRequest("Patient/cpid", "", "", "")
	req.PlanDefinitionID = "pd-cpid"
	result, _ := ApplyPlanDefinition(planDef, req)
	cpID, ok := result.CarePlan["id"].(string)
	if !ok || cpID == "" {
		t.Error("expected CarePlan to have a UUID id")
	}
}

func TestApplyPlanDefinition_RequestGroupHasID(t *testing.T) {
	planDef := &ParsedPlanDefinition{
		ID:     "pd-rgid",
		Status: "active",
	}

	req := makeApplyRequest("Patient/rgid", "", "", "")
	req.PlanDefinitionID = "pd-rgid"
	result, _ := ApplyPlanDefinition(planDef, req)
	rgID, ok := result.RequestGroup["id"].(string)
	if !ok || rgID == "" {
		t.Error("expected RequestGroup to have a UUID id")
	}
}

func TestApplyPlanDefinition_ActivitiesGenerated(t *testing.T) {
	planDef := &ParsedPlanDefinition{
		ID:     "pd-activities",
		Status: "active",
		Actions: []ApplyPlanDefinitionAction{
			{
				ID:            "a1",
				Title:         "Order Lab",
				Type:          "create",
				DefinitionURI: "ActivityDefinition/ad-lab",
			},
			{
				ID:            "a2",
				Title:         "Order Med",
				Type:          "create",
				DefinitionURI: "ActivityDefinition/ad-med",
			},
		},
	}

	req := makeApplyRequest("Patient/act", "", "", "")
	req.PlanDefinitionID = "pd-activities"
	result, _ := ApplyPlanDefinition(planDef, req)
	if len(result.Activities) != 2 {
		t.Errorf("expected 2 activities, got %d", len(result.Activities))
	}
	for _, act := range result.Activities {
		if act.ActionID == "" {
			t.Error("expected ActionID to be set")
		}
	}
}

func TestApplyPlanDefinition_CarePlanTitle(t *testing.T) {
	planDef := &ParsedPlanDefinition{
		ID:     "pd-title",
		Status: "active",
		Title:  "My Protocol Title",
	}

	req := makeApplyRequest("Patient/title", "", "", "")
	req.PlanDefinitionID = "pd-title"
	result, _ := ApplyPlanDefinition(planDef, req)
	if result.CarePlan["title"] != "My Protocol Title" {
		t.Errorf("expected title My Protocol Title, got %v", result.CarePlan["title"])
	}
}

func TestApplyActivityDefinition_DefaultResourceType(t *testing.T) {
	// Test a non-standard kind falls back to using kind as resourceType
	actDef := &ParsedActivityDefinition{
		ID:     "ad-custom",
		Status: "active",
		Kind:   "Appointment",
	}

	req := makeApplyRequest("Patient/custom", "", "", "")
	activity, err := ApplyActivityDefinition(actDef, req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if activity.ResourceType != "Appointment" {
		t.Errorf("expected resourceType Appointment, got %s", activity.ResourceType)
	}
}

func TestApplyHandler_RetiredPlanDefinition(t *testing.T) {
	resolver := newMockApplyResolver()
	resolver.resources["PlanDefinition/pd-ret"] = makePlanDefinitionJSON("pd-ret", "retired", "Retired Plan", nil)

	e := echo.New()
	body := `{"subject": "Patient/123"}`
	req := httptest.NewRequest(http.MethodPost, "/fhir/PlanDefinition/pd-ret/$apply", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("pd-ret")

	handler := ApplyHandler(resolver)
	err := handler(c)
	if err != nil {
		t.Fatalf("handler returned error: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for retired plan, got %d", rec.Code)
	}
}

func TestActivityDefinitionApplyHandler_WithParametersResource(t *testing.T) {
	resolver := newMockApplyResolver()
	resolver.resources["ActivityDefinition/ad-1"] = makeActivityDefinitionJSON("ad-1", "active", "MedicationRequest")

	e := echo.New()
	body := `{
		"resourceType": "Parameters",
		"parameter": [
			{"name": "subject", "valueString": "Patient/123"},
			{"name": "encounter", "valueString": "Encounter/enc-1"}
		]
	}`
	req := httptest.NewRequest(http.MethodPost, "/fhir/ActivityDefinition/ad-1/$apply", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("ad-1")

	handler := ActivityDefinitionApplyHandler(resolver)
	err := handler(c)
	if err != nil {
		t.Fatalf("handler returned error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestApplyPlanDefinition_ResultResourceType(t *testing.T) {
	planDef := &ParsedPlanDefinition{
		ID:     "pd-rt",
		Status: "active",
		Title:  "Test",
	}

	req := makeApplyRequest("Patient/rt", "", "", "")
	req.PlanDefinitionID = "pd-rt"
	result, _ := ApplyPlanDefinition(planDef, req)
	if result.ResourceType != "CarePlan" {
		t.Errorf("expected ResourceType CarePlan, got %s", result.ResourceType)
	}
}

func TestApplyActivityDefinition_ActionID(t *testing.T) {
	actDef := &ParsedActivityDefinition{
		ID: "ad-aid", Status: "active", Kind: "ServiceRequest",
	}
	req := makeApplyRequest("Patient/aid", "", "", "")
	activity, _ := ApplyActivityDefinition(actDef, req)
	if activity.ActionID != "ad-aid" {
		t.Errorf("expected ActionID ad-aid, got %s", activity.ActionID)
	}
}
