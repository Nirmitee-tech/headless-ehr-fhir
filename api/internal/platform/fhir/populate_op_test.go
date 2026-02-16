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

func makeQuestionnaireJSON(id, status, title string, items []interface{}) map[string]interface{} {
	q := map[string]interface{}{
		"resourceType": "Questionnaire",
		"id":           id,
		"status":       status,
		"title":        title,
		"url":          "http://example.org/fhir/Questionnaire/" + id,
	}
	if items != nil {
		q["item"] = items
	}
	return q
}

func makeQuestionnaireItemJSON(linkID, text, itemType string) map[string]interface{} {
	return map[string]interface{}{
		"linkId": linkID,
		"text":   text,
		"type":   itemType,
	}
}

func makePatientJSON(id, family, given, gender, birthDate string) map[string]interface{} {
	return map[string]interface{}{
		"resourceType": "Patient",
		"id":           id,
		"name": []interface{}{
			map[string]interface{}{
				"family": family,
				"given":  []interface{}{given},
			},
		},
		"gender":    gender,
		"birthDate": birthDate,
		"telecom": []interface{}{
			map[string]interface{}{
				"system": "phone",
				"value":  "555-1234",
			},
		},
		"address": []interface{}{
			map[string]interface{}{
				"line":       []interface{}{"123 Main St"},
				"city":       "Springfield",
				"state":      "IL",
				"postalCode": "62704",
			},
		},
	}
}

func makeObservationJSON(id, system, code, display string, valueQuantity map[string]interface{}) map[string]interface{} {
	obs := map[string]interface{}{
		"resourceType": "Observation",
		"id":           id,
		"status":       "final",
		"code": map[string]interface{}{
			"coding": []interface{}{
				map[string]interface{}{
					"system":  system,
					"code":    code,
					"display": display,
				},
			},
		},
		"subject": map[string]interface{}{
			"reference": "Patient/p-1",
		},
	}
	if valueQuantity != nil {
		obs["valueQuantity"] = valueQuantity
	}
	return obs
}

func makeConditionJSON(id, system, code, display string) map[string]interface{} {
	return map[string]interface{}{
		"resourceType": "Condition",
		"id":           id,
		"clinicalStatus": map[string]interface{}{
			"coding": []interface{}{
				map[string]interface{}{
					"system": "http://terminology.hl7.org/CodeSystem/condition-clinical",
					"code":   "active",
				},
			},
		},
		"code": map[string]interface{}{
			"coding": []interface{}{
				map[string]interface{}{
					"system":  system,
					"code":    code,
					"display": display,
				},
			},
		},
		"subject": map[string]interface{}{
			"reference": "Patient/p-1",
		},
	}
}

func makeMedicationJSON(id, system, code, display string) map[string]interface{} {
	return map[string]interface{}{
		"resourceType": "MedicationRequest",
		"id":           id,
		"status":       "active",
		"intent":       "order",
		"medicationCodeableConcept": map[string]interface{}{
			"coding": []interface{}{
				map[string]interface{}{
					"system":  system,
					"code":    code,
					"display": display,
				},
			},
		},
		"subject": map[string]interface{}{
			"reference": "Patient/p-1",
		},
	}
}

func makePopulateContext() *PopulateContext {
	patient := makePatientJSON("p-1", "Smith", "John", "male", "1990-05-15")
	obs := []map[string]interface{}{
		makeObservationJSON("obs-1", "http://loinc.org", "29463-7", "Body Weight", map[string]interface{}{
			"value": 85.0,
			"unit":  "kg",
		}),
		makeObservationJSON("obs-2", "http://loinc.org", "8302-2", "Body Height", map[string]interface{}{
			"value": 180.0,
			"unit":  "cm",
		}),
	}
	conds := []map[string]interface{}{
		makeConditionJSON("cond-1", "http://snomed.info/sct", "73211009", "Diabetes mellitus"),
	}
	meds := []map[string]interface{}{
		makeMedicationJSON("med-1", "http://www.nlm.nih.gov/research/umls/rxnorm", "860975", "Metformin 500mg"),
	}
	return &PopulateContext{
		Patient:      patient,
		Observations: obs,
		Conditions:   conds,
		Medications:  meds,
		AllResources: map[string][]map[string]interface{}{
			"Observation":       obs,
			"Condition":         conds,
			"MedicationRequest": meds,
		},
	}
}

// ============================================================================
// TestParseQuestionnaire
// ============================================================================

func TestParseQuestionnaire_Valid(t *testing.T) {
	data := makeQuestionnaireJSON("q-1", "active", "Health Survey", []interface{}{
		makeQuestionnaireItemJSON("q1", "What is your name?", "string"),
	})

	parsed, err := ParseQuestionnaire(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if parsed.ID != "q-1" {
		t.Errorf("expected ID q-1, got %s", parsed.ID)
	}
	if parsed.Title != "Health Survey" {
		t.Errorf("expected title Health Survey, got %s", parsed.Title)
	}
	if parsed.Status != "active" {
		t.Errorf("expected status active, got %s", parsed.Status)
	}
	if len(parsed.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(parsed.Items))
	}
	if parsed.Items[0].LinkID != "q1" {
		t.Errorf("expected linkId q1, got %s", parsed.Items[0].LinkID)
	}
}

func TestParseQuestionnaire_WithItems(t *testing.T) {
	data := makeQuestionnaireJSON("q-2", "active", "Multi Item", []interface{}{
		makeQuestionnaireItemJSON("q1", "Name", "string"),
		makeQuestionnaireItemJSON("q2", "DOB", "date"),
		makeQuestionnaireItemJSON("q3", "Active?", "boolean"),
	})

	parsed, err := ParseQuestionnaire(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(parsed.Items) != 3 {
		t.Fatalf("expected 3 items, got %d", len(parsed.Items))
	}
	if parsed.Items[0].Type != "string" {
		t.Errorf("expected type string, got %s", parsed.Items[0].Type)
	}
	if parsed.Items[1].Type != "date" {
		t.Errorf("expected type date, got %s", parsed.Items[1].Type)
	}
	if parsed.Items[2].Type != "boolean" {
		t.Errorf("expected type boolean, got %s", parsed.Items[2].Type)
	}
}

func TestParseQuestionnaire_NestedGroups(t *testing.T) {
	data := makeQuestionnaireJSON("q-3", "active", "Nested", []interface{}{
		map[string]interface{}{
			"linkId": "g1",
			"text":   "Demographics",
			"type":   "group",
			"item": []interface{}{
				makeQuestionnaireItemJSON("g1.1", "Name", "string"),
				makeQuestionnaireItemJSON("g1.2", "DOB", "date"),
			},
		},
	})

	parsed, err := ParseQuestionnaire(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(parsed.Items) != 1 {
		t.Fatalf("expected 1 top-level item, got %d", len(parsed.Items))
	}
	if parsed.Items[0].Type != "group" {
		t.Errorf("expected type group, got %s", parsed.Items[0].Type)
	}
	if len(parsed.Items[0].Item) != 2 {
		t.Fatalf("expected 2 nested items, got %d", len(parsed.Items[0].Item))
	}
}

func TestParseQuestionnaire_Empty(t *testing.T) {
	data := makeQuestionnaireJSON("q-4", "active", "Empty", nil)

	parsed, err := ParseQuestionnaire(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(parsed.Items) != 0 {
		t.Errorf("expected 0 items, got %d", len(parsed.Items))
	}
}

func TestParseQuestionnaire_Invalid_Nil(t *testing.T) {
	_, err := ParseQuestionnaire(nil)
	if err == nil {
		t.Error("expected error for nil data")
	}
}

func TestParseQuestionnaire_Invalid_WrongResourceType(t *testing.T) {
	data := map[string]interface{}{
		"resourceType": "Patient",
		"id":           "wrong",
	}
	_, err := ParseQuestionnaire(data)
	if err == nil {
		t.Error("expected error for wrong resourceType")
	}
}

func TestParseQuestionnaire_WithEnableWhen(t *testing.T) {
	data := makeQuestionnaireJSON("q-5", "active", "Conditional", []interface{}{
		map[string]interface{}{
			"linkId": "q1",
			"text":   "Do you smoke?",
			"type":   "boolean",
		},
		map[string]interface{}{
			"linkId": "q2",
			"text":   "How many packs per day?",
			"type":   "integer",
			"enableWhen": []interface{}{
				map[string]interface{}{
					"question":     "q1",
					"operator":     "=",
					"answerBoolean": true,
				},
			},
			"enableBehavior": "all",
		},
	})

	parsed, err := ParseQuestionnaire(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(parsed.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(parsed.Items))
	}
	if len(parsed.Items[1].EnableWhen) != 1 {
		t.Fatalf("expected 1 enableWhen, got %d", len(parsed.Items[1].EnableWhen))
	}
	if parsed.Items[1].EnableWhen[0].Question != "q1" {
		t.Errorf("expected question q1, got %s", parsed.Items[1].EnableWhen[0].Question)
	}
	if parsed.Items[1].EnableBehavior != "all" {
		t.Errorf("expected enableBehavior all, got %s", parsed.Items[1].EnableBehavior)
	}
}

func TestParseQuestionnaire_WithAnswerOptions(t *testing.T) {
	data := makeQuestionnaireJSON("q-6", "active", "Choices", []interface{}{
		map[string]interface{}{
			"linkId": "q1",
			"text":   "Blood Type",
			"type":   "choice",
			"answerOption": []interface{}{
				map[string]interface{}{
					"valueCoding": map[string]interface{}{
						"system":  "http://example.org",
						"code":    "A",
						"display": "Type A",
					},
				},
				map[string]interface{}{
					"valueCoding": map[string]interface{}{
						"system":  "http://example.org",
						"code":    "B",
						"display": "Type B",
					},
				},
			},
		},
	})

	parsed, err := ParseQuestionnaire(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(parsed.Items[0].AnswerOption) != 2 {
		t.Fatalf("expected 2 answer options, got %d", len(parsed.Items[0].AnswerOption))
	}
}

func TestParseQuestionnaire_WithInitialValues(t *testing.T) {
	data := makeQuestionnaireJSON("q-7", "active", "Defaults", []interface{}{
		map[string]interface{}{
			"linkId": "q1",
			"text":   "Default Name",
			"type":   "string",
			"initial": []interface{}{
				map[string]interface{}{
					"valueString": "Unknown",
				},
			},
		},
	})

	parsed, err := ParseQuestionnaire(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(parsed.Items[0].Initial) != 1 {
		t.Fatalf("expected 1 initial value, got %d", len(parsed.Items[0].Initial))
	}
	if parsed.Items[0].Initial[0].ValueString != "Unknown" {
		t.Errorf("expected initial valueString Unknown, got %s", parsed.Items[0].Initial[0].ValueString)
	}
}

func TestParseQuestionnaire_ItemAttributes(t *testing.T) {
	data := makeQuestionnaireJSON("q-8", "active", "Attributes", []interface{}{
		map[string]interface{}{
			"linkId":    "q1",
			"text":      "Required String",
			"type":      "string",
			"required":  true,
			"repeats":   false,
			"readOnly":  true,
			"maxLength": float64(100),
			"definition": "http://hl7.org/fhir/StructureDefinition/Patient#Patient.name",
			"code": []interface{}{
				map[string]interface{}{
					"system":  "http://loinc.org",
					"code":    "54125-0",
					"display": "Patient name",
				},
			},
		},
	})

	parsed, err := ParseQuestionnaire(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	item := parsed.Items[0]
	if !item.Required {
		t.Error("expected required=true")
	}
	if item.Repeats {
		t.Error("expected repeats=false")
	}
	if !item.ReadOnly {
		t.Error("expected readOnly=true")
	}
	if item.MaxLength != 100 {
		t.Errorf("expected maxLength=100, got %d", item.MaxLength)
	}
	if item.Definition != "http://hl7.org/fhir/StructureDefinition/Patient#Patient.name" {
		t.Errorf("unexpected definition: %s", item.Definition)
	}
	if len(item.Code) != 1 {
		t.Fatalf("expected 1 code, got %d", len(item.Code))
	}
	if item.Code[0].Code != "54125-0" {
		t.Errorf("expected code 54125-0, got %s", item.Code[0].Code)
	}
}

func TestParseQuestionnaire_SubjectType(t *testing.T) {
	data := makeQuestionnaireJSON("q-9", "active", "Typed", nil)
	data["subjectType"] = []interface{}{"Patient", "Encounter"}

	parsed, err := ParseQuestionnaire(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(parsed.SubjectType) != 2 {
		t.Fatalf("expected 2 subject types, got %d", len(parsed.SubjectType))
	}
	if parsed.SubjectType[0] != "Patient" {
		t.Errorf("expected Patient, got %s", parsed.SubjectType[0])
	}
}

// ============================================================================
// TestValidatePopulateRequest
// ============================================================================

func TestValidatePopulateRequest_Valid(t *testing.T) {
	req := &PopulateRequest{
		QuestionnaireID: "q-1",
		Subject:         "Patient/p-1",
	}
	issues := ValidatePopulateRequest(req)
	if len(issues) != 0 {
		t.Errorf("expected 0 issues, got %d: %v", len(issues), issues)
	}
}

func TestValidatePopulateRequest_MissingSubject(t *testing.T) {
	req := &PopulateRequest{
		QuestionnaireID: "q-1",
	}
	issues := ValidatePopulateRequest(req)
	if len(issues) == 0 {
		t.Error("expected validation issues for missing subject")
	}
	found := false
	for _, iss := range issues {
		if strings.Contains(iss.Diagnostics, "subject") {
			found = true
		}
	}
	if !found {
		t.Error("expected issue about missing subject")
	}
}

func TestValidatePopulateRequest_MissingQuestionnaire(t *testing.T) {
	req := &PopulateRequest{
		Subject: "Patient/p-1",
	}
	issues := ValidatePopulateRequest(req)
	if len(issues) == 0 {
		t.Error("expected validation issues for missing questionnaire")
	}
}

func TestValidatePopulateRequest_BothMissing(t *testing.T) {
	req := &PopulateRequest{}
	issues := ValidatePopulateRequest(req)
	if len(issues) < 2 {
		t.Errorf("expected at least 2 issues, got %d", len(issues))
	}
}

func TestValidatePopulateRequest_Nil(t *testing.T) {
	issues := ValidatePopulateRequest(nil)
	if len(issues) == 0 {
		t.Error("expected validation issues for nil request")
	}
}

func TestValidatePopulateRequest_InvalidSubjectFormat(t *testing.T) {
	req := &PopulateRequest{
		QuestionnaireID: "q-1",
		Subject:         "invalid-no-slash",
	}
	issues := ValidatePopulateRequest(req)
	if len(issues) == 0 {
		t.Error("expected validation issue for invalid subject format")
	}
}

func TestValidatePopulateRequest_WithInlineQuestionnaire(t *testing.T) {
	req := &PopulateRequest{
		Questionnaire: map[string]interface{}{
			"resourceType": "Questionnaire",
			"id":           "inline-q",
		},
		Subject: "Patient/p-1",
	}
	issues := ValidatePopulateRequest(req)
	if len(issues) != 0 {
		t.Errorf("expected 0 issues with inline questionnaire, got %d", len(issues))
	}
}

// ============================================================================
// TestPopulateQuestionnaire
// ============================================================================

func TestPopulateQuestionnaire_Basic(t *testing.T) {
	q := &ParsedQuestionnaire{
		ID:     "q-1",
		URL:    "http://example.org/fhir/Questionnaire/q-1",
		Title:  "Basic Survey",
		Status: "active",
		Items: []QuestionnaireItem{
			{LinkID: "q1", Text: "Name", Type: "string"},
		},
	}
	ctx := makePopulateContext()
	req := &PopulateRequest{
		QuestionnaireID: "q-1",
		Subject:         "Patient/p-1",
	}

	result, err := PopulateQuestionnaire(q, ctx, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.QuestionnaireResponse == nil {
		t.Fatal("expected non-nil QuestionnaireResponse")
	}
	rt, _ := result.QuestionnaireResponse["resourceType"].(string)
	if rt != "QuestionnaireResponse" {
		t.Errorf("expected resourceType QuestionnaireResponse, got %s", rt)
	}
	if result.TotalItems != 1 {
		t.Errorf("expected 1 total item, got %d", result.TotalItems)
	}
}

func TestPopulateQuestionnaire_WithDemographics(t *testing.T) {
	q := &ParsedQuestionnaire{
		ID:     "q-demo",
		Title:  "Demographics Form",
		Status: "active",
		Items: []QuestionnaireItem{
			{
				LinkID:     "name",
				Text:       "Patient Name",
				Type:       "string",
				Definition: "http://hl7.org/fhir/StructureDefinition/Patient#Patient.name.family",
			},
			{
				LinkID:     "gender",
				Text:       "Gender",
				Type:       "string",
				Definition: "http://hl7.org/fhir/StructureDefinition/Patient#Patient.gender",
			},
			{
				LinkID:     "dob",
				Text:       "Date of Birth",
				Type:       "date",
				Definition: "http://hl7.org/fhir/StructureDefinition/Patient#Patient.birthDate",
			},
		},
	}
	ctx := makePopulateContext()
	req := &PopulateRequest{
		QuestionnaireID: "q-demo",
		Subject:         "Patient/p-1",
	}

	result, err := PopulateQuestionnaire(q, ctx, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.PopulatedCount == 0 {
		t.Error("expected at least some populated items")
	}

	// Check that QR items exist
	items, ok := result.QuestionnaireResponse["item"].([]interface{})
	if !ok || len(items) == 0 {
		t.Fatal("expected items in QuestionnaireResponse")
	}
}

func TestPopulateQuestionnaire_WithObservations(t *testing.T) {
	q := &ParsedQuestionnaire{
		ID:     "q-obs",
		Title:  "Vitals Form",
		Status: "active",
		Items: []QuestionnaireItem{
			{
				LinkID: "weight",
				Text:   "Body Weight",
				Type:   "quantity",
				Code: []QuestionnaireCode{
					{System: "http://loinc.org", Code: "29463-7", Display: "Body Weight"},
				},
			},
		},
	}
	ctx := makePopulateContext()
	req := &PopulateRequest{
		QuestionnaireID: "q-obs",
		Subject:         "Patient/p-1",
	}

	result, err := PopulateQuestionnaire(q, ctx, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.PopulatedCount == 0 {
		t.Error("expected populated count > 0 for matching observation code")
	}
}

func TestPopulateQuestionnaire_NestedGroups(t *testing.T) {
	q := &ParsedQuestionnaire{
		ID:     "q-nested",
		Title:  "Nested Form",
		Status: "active",
		Items: []QuestionnaireItem{
			{
				LinkID: "g1",
				Text:   "Demographics",
				Type:   "group",
				Item: []QuestionnaireItem{
					{LinkID: "g1.1", Text: "Name", Type: "string",
						Definition: "http://hl7.org/fhir/StructureDefinition/Patient#Patient.name.family"},
					{LinkID: "g1.2", Text: "Gender", Type: "string",
						Definition: "http://hl7.org/fhir/StructureDefinition/Patient#Patient.gender"},
				},
			},
		},
	}
	ctx := makePopulateContext()
	req := &PopulateRequest{
		QuestionnaireID: "q-nested",
		Subject:         "Patient/p-1",
	}

	result, err := PopulateQuestionnaire(q, ctx, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Check that the group item is present with nested items
	items, ok := result.QuestionnaireResponse["item"].([]interface{})
	if !ok || len(items) == 0 {
		t.Fatal("expected items in QuestionnaireResponse")
	}
	groupItem, ok := items[0].(map[string]interface{})
	if !ok {
		t.Fatal("expected first item to be a map")
	}
	if groupItem["linkId"] != "g1" {
		t.Errorf("expected linkId g1, got %v", groupItem["linkId"])
	}
	subItems, ok := groupItem["item"].([]interface{})
	if !ok || len(subItems) == 0 {
		t.Fatal("expected nested items in group")
	}
}

func TestPopulateQuestionnaire_AllItemTypes(t *testing.T) {
	q := &ParsedQuestionnaire{
		ID:     "q-types",
		Title:  "All Types",
		Status: "active",
		Items: []QuestionnaireItem{
			{LinkID: "q-bool", Text: "Boolean", Type: "boolean"},
			{LinkID: "q-dec", Text: "Decimal", Type: "decimal"},
			{LinkID: "q-int", Text: "Integer", Type: "integer"},
			{LinkID: "q-date", Text: "Date", Type: "date"},
			{LinkID: "q-dt", Text: "DateTime", Type: "dateTime"},
			{LinkID: "q-time", Text: "Time", Type: "time"},
			{LinkID: "q-str", Text: "String", Type: "string"},
			{LinkID: "q-text", Text: "Text", Type: "text"},
			{LinkID: "q-url", Text: "URL", Type: "url"},
			{LinkID: "q-choice", Text: "Choice", Type: "choice"},
			{LinkID: "q-open", Text: "Open Choice", Type: "open-choice"},
			{LinkID: "q-attach", Text: "Attachment", Type: "attachment"},
			{LinkID: "q-ref", Text: "Reference", Type: "reference"},
			{LinkID: "q-qty", Text: "Quantity", Type: "quantity"},
			{LinkID: "q-disp", Text: "Display", Type: "display"},
			{LinkID: "q-grp", Text: "Group", Type: "group"},
		},
	}
	ctx := makePopulateContext()
	req := &PopulateRequest{
		QuestionnaireID: "q-types",
		Subject:         "Patient/p-1",
	}

	result, err := PopulateQuestionnaire(q, ctx, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TotalItems != 16 {
		t.Errorf("expected 16 total items, got %d", result.TotalItems)
	}
}

func TestPopulateQuestionnaire_NilQuestionnaire(t *testing.T) {
	ctx := makePopulateContext()
	req := &PopulateRequest{Subject: "Patient/p-1"}

	_, err := PopulateQuestionnaire(nil, ctx, req)
	if err == nil {
		t.Error("expected error for nil questionnaire")
	}
}

func TestPopulateQuestionnaire_NilContext(t *testing.T) {
	q := &ParsedQuestionnaire{
		ID:     "q-1",
		Status: "active",
		Items: []QuestionnaireItem{
			{LinkID: "q1", Text: "Name", Type: "string"},
		},
	}
	req := &PopulateRequest{Subject: "Patient/p-1"}

	result, err := PopulateQuestionnaire(q, nil, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should still produce an empty QR
	if result.QuestionnaireResponse == nil {
		t.Fatal("expected non-nil QR even with nil context")
	}
}

func TestPopulateQuestionnaire_WithInitialValues(t *testing.T) {
	boolVal := true
	q := &ParsedQuestionnaire{
		ID:     "q-init",
		Status: "active",
		Items: []QuestionnaireItem{
			{
				LinkID: "q1",
				Text:   "Default answer",
				Type:   "boolean",
				Initial: []InitialValue{
					{ValueBoolean: &boolVal},
				},
			},
		},
	}
	ctx := &PopulateContext{}
	req := &PopulateRequest{Subject: "Patient/p-1"}

	result, err := PopulateQuestionnaire(q, ctx, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.PopulatedCount == 0 {
		t.Error("expected populated count > 0 for initial value item")
	}
}

// ============================================================================
// TestPopulateItem
// ============================================================================

func TestPopulateItem_String(t *testing.T) {
	item := &QuestionnaireItem{
		LinkID:     "name",
		Text:       "Name",
		Type:       "string",
		Definition: "http://hl7.org/fhir/StructureDefinition/Patient#Patient.name.family",
	}
	ctx := makePopulateContext()

	result, populated := PopulateItem(item, ctx)
	if !populated {
		t.Error("expected item to be populated")
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result["linkId"] != "name" {
		t.Errorf("expected linkId name, got %v", result["linkId"])
	}
}

func TestPopulateItem_Boolean(t *testing.T) {
	boolVal := true
	item := &QuestionnaireItem{
		LinkID: "active",
		Text:   "Active?",
		Type:   "boolean",
		Initial: []InitialValue{
			{ValueBoolean: &boolVal},
		},
	}
	ctx := &PopulateContext{}

	result, populated := PopulateItem(item, ctx)
	if !populated {
		t.Error("expected item to be populated from initial value")
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestPopulateItem_Date(t *testing.T) {
	item := &QuestionnaireItem{
		LinkID:     "dob",
		Text:       "Date of Birth",
		Type:       "date",
		Definition: "http://hl7.org/fhir/StructureDefinition/Patient#Patient.birthDate",
	}
	ctx := makePopulateContext()

	result, populated := PopulateItem(item, ctx)
	if !populated {
		t.Error("expected item to be populated for birthDate")
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestPopulateItem_Choice(t *testing.T) {
	item := &QuestionnaireItem{
		LinkID: "bloodtype",
		Text:   "Blood Type",
		Type:   "choice",
		AnswerOption: []AnswerOption{
			{ValueCoding: map[string]interface{}{"code": "A"}},
			{ValueCoding: map[string]interface{}{"code": "B"}},
		},
	}
	ctx := &PopulateContext{}

	result, populated := PopulateItem(item, ctx)
	// No matching data, but should still create an item shell
	if populated {
		t.Error("expected item not to be populated without matching data")
	}
	_ = result
}

func TestPopulateItem_Quantity(t *testing.T) {
	item := &QuestionnaireItem{
		LinkID: "weight",
		Text:   "Body Weight",
		Type:   "quantity",
		Code: []QuestionnaireCode{
			{System: "http://loinc.org", Code: "29463-7"},
		},
	}
	ctx := makePopulateContext()

	result, populated := PopulateItem(item, ctx)
	if !populated {
		t.Error("expected item to be populated from observation")
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestPopulateItem_Reference(t *testing.T) {
	item := &QuestionnaireItem{
		LinkID: "ref1",
		Text:   "Reference field",
		Type:   "reference",
	}
	ctx := &PopulateContext{}

	_, populated := PopulateItem(item, ctx)
	if populated {
		t.Error("expected reference item not populated without matching data")
	}
}

func TestPopulateItem_Display(t *testing.T) {
	item := &QuestionnaireItem{
		LinkID: "info",
		Text:   "This is informational only",
		Type:   "display",
	}
	ctx := &PopulateContext{}

	result, populated := PopulateItem(item, ctx)
	// Display items are never populated with answers
	if populated {
		t.Error("display items should not be populated")
	}
	_ = result
}

func TestPopulateItem_Group(t *testing.T) {
	item := &QuestionnaireItem{
		LinkID: "g1",
		Text:   "Group",
		Type:   "group",
		Item: []QuestionnaireItem{
			{
				LinkID:     "g1.1",
				Text:       "Name",
				Type:       "string",
				Definition: "http://hl7.org/fhir/StructureDefinition/Patient#Patient.name.family",
			},
		},
	}
	ctx := makePopulateContext()

	result, populated := PopulateItem(item, ctx)
	if !populated {
		t.Error("expected group to be populated when children have data")
	}
	if result == nil {
		t.Fatal("expected non-nil result for group")
	}
	subItems, ok := result["item"].([]interface{})
	if !ok {
		t.Fatal("expected nested items in group result")
	}
	if len(subItems) == 0 {
		t.Error("expected at least one nested item")
	}
}

// ============================================================================
// TestBuildQuestionnaireResponseItem
// ============================================================================

func TestBuildQuestionnaireResponseItem_String(t *testing.T) {
	item := &QuestionnaireItem{LinkID: "q1", Text: "Name", Type: "string"}
	result := BuildQuestionnaireResponseItem(item, "John Smith")
	if result["linkId"] != "q1" {
		t.Errorf("expected linkId q1, got %v", result["linkId"])
	}
	answers, ok := result["answer"].([]interface{})
	if !ok || len(answers) == 0 {
		t.Fatal("expected answer array")
	}
	ans, ok := answers[0].(map[string]interface{})
	if !ok {
		t.Fatal("expected answer map")
	}
	if ans["valueString"] != "John Smith" {
		t.Errorf("expected valueString John Smith, got %v", ans["valueString"])
	}
}

func TestBuildQuestionnaireResponseItem_Boolean(t *testing.T) {
	item := &QuestionnaireItem{LinkID: "q1", Text: "Active?", Type: "boolean"}
	result := BuildQuestionnaireResponseItem(item, true)
	answers, ok := result["answer"].([]interface{})
	if !ok || len(answers) == 0 {
		t.Fatal("expected answer array")
	}
	ans := answers[0].(map[string]interface{})
	if ans["valueBoolean"] != true {
		t.Errorf("expected valueBoolean true, got %v", ans["valueBoolean"])
	}
}

func TestBuildQuestionnaireResponseItem_Integer(t *testing.T) {
	item := &QuestionnaireItem{LinkID: "q1", Text: "Age", Type: "integer"}
	result := BuildQuestionnaireResponseItem(item, 35)
	answers := result["answer"].([]interface{})
	ans := answers[0].(map[string]interface{})
	if ans["valueInteger"] != 35 {
		t.Errorf("expected valueInteger 35, got %v", ans["valueInteger"])
	}
}

func TestBuildQuestionnaireResponseItem_Decimal(t *testing.T) {
	item := &QuestionnaireItem{LinkID: "q1", Text: "Weight", Type: "decimal"}
	result := BuildQuestionnaireResponseItem(item, 85.5)
	answers := result["answer"].([]interface{})
	ans := answers[0].(map[string]interface{})
	if ans["valueDecimal"] != 85.5 {
		t.Errorf("expected valueDecimal 85.5, got %v", ans["valueDecimal"])
	}
}

func TestBuildQuestionnaireResponseItem_Date(t *testing.T) {
	item := &QuestionnaireItem{LinkID: "q1", Text: "DOB", Type: "date"}
	result := BuildQuestionnaireResponseItem(item, "1990-05-15")
	answers := result["answer"].([]interface{})
	ans := answers[0].(map[string]interface{})
	if ans["valueDate"] != "1990-05-15" {
		t.Errorf("expected valueDate 1990-05-15, got %v", ans["valueDate"])
	}
}

func TestBuildQuestionnaireResponseItem_Coding(t *testing.T) {
	item := &QuestionnaireItem{LinkID: "q1", Text: "Code", Type: "choice"}
	coding := map[string]interface{}{"system": "http://example.org", "code": "A"}
	result := BuildQuestionnaireResponseItem(item, coding)
	answers := result["answer"].([]interface{})
	ans := answers[0].(map[string]interface{})
	if ans["valueCoding"] == nil {
		t.Error("expected valueCoding in answer")
	}
}

func TestBuildQuestionnaireResponseItem_Quantity(t *testing.T) {
	item := &QuestionnaireItem{LinkID: "q1", Text: "Weight", Type: "quantity"}
	qty := map[string]interface{}{"value": 85.0, "unit": "kg"}
	result := BuildQuestionnaireResponseItem(item, qty)
	answers := result["answer"].([]interface{})
	ans := answers[0].(map[string]interface{})
	if ans["valueQuantity"] == nil {
		t.Error("expected valueQuantity in answer")
	}
}

// ============================================================================
// TestBuildEmptyQuestionnaireResponse
// ============================================================================

func TestBuildEmptyQuestionnaireResponse(t *testing.T) {
	q := &ParsedQuestionnaire{
		ID:    "q-1",
		URL:   "http://example.org/fhir/Questionnaire/q-1",
		Title: "Health Survey",
		Items: []QuestionnaireItem{
			{LinkID: "q1", Text: "Name", Type: "string"},
		},
	}

	qr := BuildEmptyQuestionnaireResponse(q, "Patient/p-1")
	if qr == nil {
		t.Fatal("expected non-nil QR")
	}
	if qr["resourceType"] != "QuestionnaireResponse" {
		t.Errorf("expected resourceType QuestionnaireResponse, got %v", qr["resourceType"])
	}
	if qr["status"] != "in-progress" {
		t.Errorf("expected status in-progress, got %v", qr["status"])
	}
	if qr["questionnaire"] != "http://example.org/fhir/Questionnaire/q-1" {
		t.Errorf("unexpected questionnaire: %v", qr["questionnaire"])
	}
	subj, ok := qr["subject"].(map[string]interface{})
	if !ok || subj["reference"] != "Patient/p-1" {
		t.Error("expected subject reference Patient/p-1")
	}
}

func TestBuildEmptyQuestionnaireResponse_NoURL(t *testing.T) {
	q := &ParsedQuestionnaire{
		ID:    "q-2",
		Title: "No URL",
	}

	qr := BuildEmptyQuestionnaireResponse(q, "Patient/p-1")
	if qr["questionnaire"] != "Questionnaire/q-2" {
		t.Errorf("expected fallback to Questionnaire/q-2, got %v", qr["questionnaire"])
	}
}

// ============================================================================
// TestEvaluateEnableWhen
// ============================================================================

func TestEvaluateEnableWhen_Exists_True(t *testing.T) {
	conditions := []EnableWhenCondition{
		{Question: "q1", Operator: "exists", Answer: true},
	}
	answers := map[string]interface{}{"q1": "some value"}

	if !EvaluateEnableWhen(conditions, "all", answers) {
		t.Error("expected enable when exists=true to pass when answer exists")
	}
}

func TestEvaluateEnableWhen_Exists_False(t *testing.T) {
	conditions := []EnableWhenCondition{
		{Question: "q1", Operator: "exists", Answer: true},
	}
	answers := map[string]interface{}{}

	if EvaluateEnableWhen(conditions, "all", answers) {
		t.Error("expected enable when exists=true to fail when answer missing")
	}
}

func TestEvaluateEnableWhen_Equals(t *testing.T) {
	conditions := []EnableWhenCondition{
		{Question: "q1", Operator: "=", Answer: "yes"},
	}
	answers := map[string]interface{}{"q1": "yes"}

	if !EvaluateEnableWhen(conditions, "all", answers) {
		t.Error("expected enable when = to pass")
	}
}

func TestEvaluateEnableWhen_NotEquals(t *testing.T) {
	conditions := []EnableWhenCondition{
		{Question: "q1", Operator: "!=", Answer: "no"},
	}
	answers := map[string]interface{}{"q1": "yes"}

	if !EvaluateEnableWhen(conditions, "all", answers) {
		t.Error("expected enable when != to pass")
	}
}

func TestEvaluateEnableWhen_GreaterThan(t *testing.T) {
	conditions := []EnableWhenCondition{
		{Question: "q1", Operator: ">", Answer: float64(18)},
	}
	answers := map[string]interface{}{"q1": float64(25)}

	if !EvaluateEnableWhen(conditions, "all", answers) {
		t.Error("expected enable when > to pass")
	}
}

func TestEvaluateEnableWhen_LessThan(t *testing.T) {
	conditions := []EnableWhenCondition{
		{Question: "q1", Operator: "<", Answer: float64(100)},
	}
	answers := map[string]interface{}{"q1": float64(50)}

	if !EvaluateEnableWhen(conditions, "all", answers) {
		t.Error("expected enable when < to pass")
	}
}

func TestEvaluateEnableWhen_GreaterOrEqual(t *testing.T) {
	conditions := []EnableWhenCondition{
		{Question: "q1", Operator: ">=", Answer: float64(18)},
	}
	answers := map[string]interface{}{"q1": float64(18)}

	if !EvaluateEnableWhen(conditions, "all", answers) {
		t.Error("expected enable when >= to pass for equal value")
	}
}

func TestEvaluateEnableWhen_LessOrEqual(t *testing.T) {
	conditions := []EnableWhenCondition{
		{Question: "q1", Operator: "<=", Answer: float64(100)},
	}
	answers := map[string]interface{}{"q1": float64(100)}

	if !EvaluateEnableWhen(conditions, "all", answers) {
		t.Error("expected enable when <= to pass for equal value")
	}
}

func TestEvaluateEnableWhen_AnyBehavior(t *testing.T) {
	conditions := []EnableWhenCondition{
		{Question: "q1", Operator: "=", Answer: "yes"},
		{Question: "q2", Operator: "=", Answer: "no"},
	}
	// Only q1 matches
	answers := map[string]interface{}{"q1": "yes", "q2": "yes"}

	if !EvaluateEnableWhen(conditions, "any", answers) {
		t.Error("expected any behavior to pass when at least one condition matches")
	}
}

func TestEvaluateEnableWhen_AllBehavior_Fail(t *testing.T) {
	conditions := []EnableWhenCondition{
		{Question: "q1", Operator: "=", Answer: "yes"},
		{Question: "q2", Operator: "=", Answer: "no"},
	}
	// Only q1 matches
	answers := map[string]interface{}{"q1": "yes", "q2": "yes"}

	if EvaluateEnableWhen(conditions, "all", answers) {
		t.Error("expected all behavior to fail when not all conditions match")
	}
}

func TestEvaluateEnableWhen_EmptyConditions(t *testing.T) {
	if !EvaluateEnableWhen(nil, "all", nil) {
		t.Error("expected true for empty conditions")
	}
}

func TestEvaluateEnableWhen_BooleanAnswer(t *testing.T) {
	conditions := []EnableWhenCondition{
		{Question: "q1", Operator: "=", Answer: true},
	}
	answers := map[string]interface{}{"q1": true}

	if !EvaluateEnableWhen(conditions, "all", answers) {
		t.Error("expected enable when = true to pass")
	}
}

// ============================================================================
// TestExtractPopulationValue
// ============================================================================

func TestExtractPopulationValue_FromPatientName(t *testing.T) {
	item := &QuestionnaireItem{
		LinkID:     "name",
		Type:       "string",
		Definition: "http://hl7.org/fhir/StructureDefinition/Patient#Patient.name.family",
	}
	ctx := makePopulateContext()

	val, found := ExtractPopulationValue(item, ctx)
	if !found {
		t.Error("expected value to be found for patient name")
	}
	if val != "Smith" {
		t.Errorf("expected Smith, got %v", val)
	}
}

func TestExtractPopulationValue_FromPatientGender(t *testing.T) {
	item := &QuestionnaireItem{
		LinkID:     "gender",
		Type:       "string",
		Definition: "http://hl7.org/fhir/StructureDefinition/Patient#Patient.gender",
	}
	ctx := makePopulateContext()

	val, found := ExtractPopulationValue(item, ctx)
	if !found {
		t.Error("expected value to be found for patient gender")
	}
	if val != "male" {
		t.Errorf("expected male, got %v", val)
	}
}

func TestExtractPopulationValue_FromPatientBirthDate(t *testing.T) {
	item := &QuestionnaireItem{
		LinkID:     "dob",
		Type:       "date",
		Definition: "http://hl7.org/fhir/StructureDefinition/Patient#Patient.birthDate",
	}
	ctx := makePopulateContext()

	val, found := ExtractPopulationValue(item, ctx)
	if !found {
		t.Error("expected value to be found for patient birthDate")
	}
	if val != "1990-05-15" {
		t.Errorf("expected 1990-05-15, got %v", val)
	}
}

func TestExtractPopulationValue_FromObservation(t *testing.T) {
	item := &QuestionnaireItem{
		LinkID: "weight",
		Type:   "quantity",
		Code: []QuestionnaireCode{
			{System: "http://loinc.org", Code: "29463-7"},
		},
	}
	ctx := makePopulateContext()

	val, found := ExtractPopulationValue(item, ctx)
	if !found {
		t.Error("expected value to be found for observation weight")
	}
	valMap, ok := val.(map[string]interface{})
	if !ok {
		t.Fatal("expected quantity map value")
	}
	if valMap["value"] != 85.0 {
		t.Errorf("expected value 85.0, got %v", valMap["value"])
	}
}

func TestExtractPopulationValue_FromCondition(t *testing.T) {
	item := &QuestionnaireItem{
		LinkID: "condition",
		Type:   "choice",
		Code: []QuestionnaireCode{
			{System: "http://snomed.info/sct", Code: "73211009"},
		},
	}
	ctx := makePopulateContext()

	val, found := ExtractPopulationValue(item, ctx)
	if !found {
		t.Error("expected value to be found for condition code")
	}
	if val == nil {
		t.Error("expected non-nil value")
	}
}

func TestExtractPopulationValue_NotFound(t *testing.T) {
	item := &QuestionnaireItem{
		LinkID: "unknown",
		Type:   "string",
		Code: []QuestionnaireCode{
			{System: "http://loinc.org", Code: "99999-9"},
		},
	}
	ctx := makePopulateContext()

	_, found := ExtractPopulationValue(item, ctx)
	if found {
		t.Error("expected value not to be found for unknown code")
	}
}

func TestExtractPopulationValue_NilContext(t *testing.T) {
	item := &QuestionnaireItem{
		LinkID: "q1",
		Type:   "string",
	}

	_, found := ExtractPopulationValue(item, nil)
	if found {
		t.Error("expected value not to be found with nil context")
	}
}

// ============================================================================
// TestMatchResourceByCode
// ============================================================================

func TestMatchResourceByCode_Matching(t *testing.T) {
	code := QuestionnaireCode{System: "http://loinc.org", Code: "29463-7"}
	resources := []map[string]interface{}{
		makeObservationJSON("obs-1", "http://loinc.org", "29463-7", "Weight", nil),
		makeObservationJSON("obs-2", "http://loinc.org", "8302-2", "Height", nil),
	}

	matches := MatchResourceByCode(code, resources)
	if len(matches) != 1 {
		t.Fatalf("expected 1 match, got %d", len(matches))
	}
}

func TestMatchResourceByCode_NoMatch(t *testing.T) {
	code := QuestionnaireCode{System: "http://loinc.org", Code: "99999-9"}
	resources := []map[string]interface{}{
		makeObservationJSON("obs-1", "http://loinc.org", "29463-7", "Weight", nil),
	}

	matches := MatchResourceByCode(code, resources)
	if len(matches) != 0 {
		t.Fatalf("expected 0 matches, got %d", len(matches))
	}
}

func TestMatchResourceByCode_MultipleMatches(t *testing.T) {
	code := QuestionnaireCode{System: "http://loinc.org", Code: "29463-7"}
	resources := []map[string]interface{}{
		makeObservationJSON("obs-1", "http://loinc.org", "29463-7", "Weight 1", nil),
		makeObservationJSON("obs-2", "http://loinc.org", "29463-7", "Weight 2", nil),
	}

	matches := MatchResourceByCode(code, resources)
	if len(matches) != 2 {
		t.Fatalf("expected 2 matches, got %d", len(matches))
	}
}

func TestMatchResourceByCode_EmptyResources(t *testing.T) {
	code := QuestionnaireCode{System: "http://loinc.org", Code: "29463-7"}
	matches := MatchResourceByCode(code, nil)
	if len(matches) != 0 {
		t.Fatalf("expected 0 matches for nil resources, got %d", len(matches))
	}
}

// ============================================================================
// TestConvertToAnswerValue
// ============================================================================

func TestConvertToAnswerValue_String(t *testing.T) {
	result := ConvertToAnswerValue("hello", "string")
	if result != "hello" {
		t.Errorf("expected hello, got %v", result)
	}
}

func TestConvertToAnswerValue_Boolean(t *testing.T) {
	result := ConvertToAnswerValue(true, "boolean")
	if result != true {
		t.Errorf("expected true, got %v", result)
	}
}

func TestConvertToAnswerValue_Integer(t *testing.T) {
	result := ConvertToAnswerValue(float64(42), "integer")
	if result != 42 {
		t.Errorf("expected 42, got %v", result)
	}
}

func TestConvertToAnswerValue_Decimal(t *testing.T) {
	result := ConvertToAnswerValue(3.14, "decimal")
	if result != 3.14 {
		t.Errorf("expected 3.14, got %v", result)
	}
}

func TestConvertToAnswerValue_Date(t *testing.T) {
	result := ConvertToAnswerValue("2024-01-15", "date")
	if result != "2024-01-15" {
		t.Errorf("expected 2024-01-15, got %v", result)
	}
}

func TestConvertToAnswerValue_Coding(t *testing.T) {
	coding := map[string]interface{}{"system": "http://example.org", "code": "A"}
	result := ConvertToAnswerValue(coding, "choice")
	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatal("expected map result for coding")
	}
	if resultMap["code"] != "A" {
		t.Errorf("expected code A, got %v", resultMap["code"])
	}
}

func TestConvertToAnswerValue_Quantity(t *testing.T) {
	qty := map[string]interface{}{"value": 85.0, "unit": "kg"}
	result := ConvertToAnswerValue(qty, "quantity")
	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatal("expected map result for quantity")
	}
	if resultMap["value"] != 85.0 {
		t.Errorf("expected value 85.0, got %v", resultMap["value"])
	}
}

func TestConvertToAnswerValue_IntegerFromFloat(t *testing.T) {
	result := ConvertToAnswerValue(float64(100), "integer")
	if result != 100 {
		t.Errorf("expected int 100, got %v (type %T)", result, result)
	}
}

// ============================================================================
// TestPopulateHandler
// ============================================================================

func TestPopulateHandler_Success(t *testing.T) {
	resolver := NewInMemoryPopulateResolver()
	resolver.AddPatient("p-1", makePatientJSON("p-1", "Smith", "John", "male", "1990-05-15"))
	resolver.AddQuestionnaire("q-1", makeQuestionnaireJSON("q-1", "active", "Survey", []interface{}{
		makeQuestionnaireItemJSON("q1", "Name", "string"),
	}))

	e := echo.New()
	body := `{"subject":"Patient/p-1"}`
	req := httptest.NewRequest(http.MethodPost, "/fhir/Questionnaire/q-1/$populate", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/fhir+json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("q-1")

	handler := PopulateHandler(resolver)
	err := handler(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if result["resourceType"] != "QuestionnaireResponse" {
		t.Errorf("expected QuestionnaireResponse, got %v", result["resourceType"])
	}
}

func TestPopulateHandler_NotFound(t *testing.T) {
	resolver := NewInMemoryPopulateResolver()

	e := echo.New()
	body := `{"subject":"Patient/p-1"}`
	req := httptest.NewRequest(http.MethodPost, "/fhir/Questionnaire/missing/$populate", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/fhir+json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("missing")

	handler := PopulateHandler(resolver)
	err := handler(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", rec.Code)
	}
}

func TestPopulateHandler_InvalidRequest(t *testing.T) {
	resolver := NewInMemoryPopulateResolver()
	resolver.AddQuestionnaire("q-1", makeQuestionnaireJSON("q-1", "active", "Survey", nil))

	e := echo.New()
	body := `{"subject":""}` // empty subject
	req := httptest.NewRequest(http.MethodPost, "/fhir/Questionnaire/q-1/$populate", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/fhir+json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("q-1")

	handler := PopulateHandler(resolver)
	err := handler(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestPopulateHandler_EmptyBody(t *testing.T) {
	resolver := NewInMemoryPopulateResolver()
	resolver.AddQuestionnaire("q-1", makeQuestionnaireJSON("q-1", "active", "Survey", nil))

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/fhir/Questionnaire/q-1/$populate", strings.NewReader(""))
	req.Header.Set("Content-Type", "application/fhir+json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("q-1")

	handler := PopulateHandler(resolver)
	err := handler(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestPopulateHandler_InvalidJSON(t *testing.T) {
	resolver := NewInMemoryPopulateResolver()
	resolver.AddQuestionnaire("q-1", makeQuestionnaireJSON("q-1", "active", "Survey", nil))

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/fhir/Questionnaire/q-1/$populate", strings.NewReader("{invalid"))
	req.Header.Set("Content-Type", "application/fhir+json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("q-1")

	handler := PopulateHandler(resolver)
	err := handler(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestPopulateHandler_FHIRParametersFormat(t *testing.T) {
	resolver := NewInMemoryPopulateResolver()
	resolver.AddPatient("p-1", makePatientJSON("p-1", "Smith", "John", "male", "1990-05-15"))
	resolver.AddQuestionnaire("q-1", makeQuestionnaireJSON("q-1", "active", "Survey", []interface{}{
		makeQuestionnaireItemJSON("q1", "Name", "string"),
	}))

	e := echo.New()
	body := `{
		"resourceType": "Parameters",
		"parameter": [
			{"name": "subject", "valueString": "Patient/p-1"}
		]
	}`
	req := httptest.NewRequest(http.MethodPost, "/fhir/Questionnaire/q-1/$populate", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/fhir+json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("q-1")

	handler := PopulateHandler(resolver)
	err := handler(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d. Body: %s", rec.Code, rec.Body.String())
	}
}

func TestPopulateHandler_MissingID(t *testing.T) {
	resolver := NewInMemoryPopulateResolver()

	e := echo.New()
	body := `{"subject":"Patient/p-1"}`
	req := httptest.NewRequest(http.MethodPost, "/fhir/Questionnaire//$populate", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/fhir+json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("")

	handler := PopulateHandler(resolver)
	err := handler(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

// ============================================================================
// TestInMemoryPopulateResolver
// ============================================================================

func TestInMemoryPopulateResolver_AddAndResolvePatient(t *testing.T) {
	resolver := NewInMemoryPopulateResolver()
	patient := makePatientJSON("p-1", "Smith", "John", "male", "1990-05-15")
	resolver.AddPatient("p-1", patient)

	resolved, err := resolver.ResolvePatient(nil, "Patient/p-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resolved == nil {
		t.Fatal("expected non-nil patient")
	}
	if resolved["id"] != "p-1" {
		t.Errorf("expected id p-1, got %v", resolved["id"])
	}
}

func TestInMemoryPopulateResolver_ResolvePatient_NotFound(t *testing.T) {
	resolver := NewInMemoryPopulateResolver()

	_, err := resolver.ResolvePatient(nil, "Patient/missing")
	if err == nil {
		t.Error("expected error for missing patient")
	}
}

func TestInMemoryPopulateResolver_AddAndResolveResources(t *testing.T) {
	resolver := NewInMemoryPopulateResolver()
	obs := makeObservationJSON("obs-1", "http://loinc.org", "29463-7", "Weight", nil)
	resolver.AddResource("p-1", "Observation", obs)

	resources, err := resolver.ResolveResources(nil, "Patient/p-1", "Observation")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resources) != 1 {
		t.Fatalf("expected 1 resource, got %d", len(resources))
	}
}

func TestInMemoryPopulateResolver_ResolveResources_Empty(t *testing.T) {
	resolver := NewInMemoryPopulateResolver()

	resources, err := resolver.ResolveResources(nil, "Patient/p-1", "Observation")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resources) != 0 {
		t.Errorf("expected 0 resources, got %d", len(resources))
	}
}

func TestInMemoryPopulateResolver_MultipleResources(t *testing.T) {
	resolver := NewInMemoryPopulateResolver()
	resolver.AddResource("p-1", "Observation", makeObservationJSON("obs-1", "http://loinc.org", "29463-7", "Weight", nil))
	resolver.AddResource("p-1", "Observation", makeObservationJSON("obs-2", "http://loinc.org", "8302-2", "Height", nil))
	resolver.AddResource("p-1", "Condition", makeConditionJSON("cond-1", "http://snomed.info/sct", "73211009", "Diabetes"))

	obs, err := resolver.ResolveResources(nil, "Patient/p-1", "Observation")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(obs) != 2 {
		t.Errorf("expected 2 observations, got %d", len(obs))
	}

	conds, err := resolver.ResolveResources(nil, "Patient/p-1", "Condition")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(conds) != 1 {
		t.Errorf("expected 1 condition, got %d", len(conds))
	}
}

// ============================================================================
// TestBuildPopulateContext
// ============================================================================

func TestBuildPopulateContext_Basic(t *testing.T) {
	patient := makePatientJSON("p-1", "Smith", "John", "male", "1990-05-15")
	resources := map[string][]map[string]interface{}{
		"Observation": {
			makeObservationJSON("obs-1", "http://loinc.org", "29463-7", "Weight", nil),
		},
		"Condition": {
			makeConditionJSON("cond-1", "http://snomed.info/sct", "73211009", "Diabetes"),
		},
		"MedicationRequest": {
			makeMedicationJSON("med-1", "http://www.nlm.nih.gov/research/umls/rxnorm", "860975", "Metformin"),
		},
	}

	ctx := BuildPopulateContext(patient, resources)
	if ctx == nil {
		t.Fatal("expected non-nil context")
	}
	if ctx.Patient == nil {
		t.Fatal("expected non-nil patient")
	}
	if len(ctx.Observations) != 1 {
		t.Errorf("expected 1 observation, got %d", len(ctx.Observations))
	}
	if len(ctx.Conditions) != 1 {
		t.Errorf("expected 1 condition, got %d", len(ctx.Conditions))
	}
	if len(ctx.Medications) != 1 {
		t.Errorf("expected 1 medication, got %d", len(ctx.Medications))
	}
	if len(ctx.AllResources) != 3 {
		t.Errorf("expected 3 resource types, got %d", len(ctx.AllResources))
	}
}

func TestBuildPopulateContext_NilPatient(t *testing.T) {
	ctx := BuildPopulateContext(nil, nil)
	if ctx == nil {
		t.Fatal("expected non-nil context even with nil inputs")
	}
	if ctx.Patient != nil {
		t.Error("expected nil patient")
	}
}

func TestBuildPopulateContext_EmptyResources(t *testing.T) {
	patient := makePatientJSON("p-1", "Smith", "John", "male", "1990-05-15")
	ctx := BuildPopulateContext(patient, map[string][]map[string]interface{}{})
	if len(ctx.Observations) != 0 {
		t.Errorf("expected 0 observations, got %d", len(ctx.Observations))
	}
}

// ============================================================================
// Edge Cases
// ============================================================================

func TestPopulateQuestionnaire_DeeplyNestedItems(t *testing.T) {
	q := &ParsedQuestionnaire{
		ID:     "q-deep",
		Status: "active",
		Items: []QuestionnaireItem{
			{
				LinkID: "g1",
				Type:   "group",
				Text:   "Level 1",
				Item: []QuestionnaireItem{
					{
						LinkID: "g1.1",
						Type:   "group",
						Text:   "Level 2",
						Item: []QuestionnaireItem{
							{
								LinkID: "g1.1.1",
								Type:   "group",
								Text:   "Level 3",
								Item: []QuestionnaireItem{
									{
										LinkID:     "g1.1.1.1",
										Type:       "string",
										Text:       "Deep Item",
										Definition: "http://hl7.org/fhir/StructureDefinition/Patient#Patient.name.family",
									},
								},
							},
						},
					},
				},
			},
		},
	}
	ctx := makePopulateContext()
	req := &PopulateRequest{Subject: "Patient/p-1"}

	result, err := PopulateQuestionnaire(q, ctx, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TotalItems < 4 {
		t.Errorf("expected at least 4 total items, got %d", result.TotalItems)
	}
}

func TestPopulateQuestionnaire_EmptyContext(t *testing.T) {
	q := &ParsedQuestionnaire{
		ID:     "q-empty-ctx",
		Status: "active",
		Items: []QuestionnaireItem{
			{LinkID: "q1", Text: "Name", Type: "string"},
		},
	}
	ctx := &PopulateContext{}
	req := &PopulateRequest{Subject: "Patient/p-1"}

	result, err := PopulateQuestionnaire(q, ctx, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.PopulatedCount != 0 {
		t.Errorf("expected 0 populated with empty context, got %d", result.PopulatedCount)
	}
}

func TestExtractPopulationValue_FromPatientPhone(t *testing.T) {
	item := &QuestionnaireItem{
		LinkID:     "phone",
		Type:       "string",
		Definition: "http://hl7.org/fhir/StructureDefinition/Patient#Patient.telecom.value",
	}
	ctx := makePopulateContext()

	val, found := ExtractPopulationValue(item, ctx)
	if !found {
		t.Error("expected value to be found for patient phone")
	}
	if val != "555-1234" {
		t.Errorf("expected 555-1234, got %v", val)
	}
}

func TestExtractPopulationValue_FromPatientAddress(t *testing.T) {
	item := &QuestionnaireItem{
		LinkID:     "city",
		Type:       "string",
		Definition: "http://hl7.org/fhir/StructureDefinition/Patient#Patient.address.city",
	}
	ctx := makePopulateContext()

	val, found := ExtractPopulationValue(item, ctx)
	if !found {
		t.Error("expected value to be found for patient city")
	}
	if val != "Springfield" {
		t.Errorf("expected Springfield, got %v", val)
	}
}

func TestExtractPopulationValue_FromMedication(t *testing.T) {
	item := &QuestionnaireItem{
		LinkID: "med",
		Type:   "choice",
		Code: []QuestionnaireCode{
			{System: "http://www.nlm.nih.gov/research/umls/rxnorm", Code: "860975"},
		},
	}
	ctx := makePopulateContext()

	val, found := ExtractPopulationValue(item, ctx)
	if !found {
		t.Error("expected value to be found for medication")
	}
	if val == nil {
		t.Error("expected non-nil value for medication")
	}
}

func TestPopulateQuestionnaire_WithWarnings(t *testing.T) {
	q := &ParsedQuestionnaire{
		ID:     "q-warn",
		Status: "active",
		Items: []QuestionnaireItem{
			{LinkID: "q1", Text: "Unknown Code", Type: "string",
				Code: []QuestionnaireCode{
					{System: "http://unknown.org", Code: "UNKNOWN"},
				},
			},
		},
	}
	ctx := makePopulateContext()
	req := &PopulateRequest{Subject: "Patient/p-1"}

	result, err := PopulateQuestionnaire(q, ctx, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should complete without error even if some items cannot be populated
	if result.QuestionnaireResponse == nil {
		t.Fatal("expected non-nil QR")
	}
}

func TestPopulateHandler_WithObservationData(t *testing.T) {
	resolver := NewInMemoryPopulateResolver()
	resolver.AddPatient("p-1", makePatientJSON("p-1", "Smith", "John", "male", "1990-05-15"))
	resolver.AddResource("p-1", "Observation", makeObservationJSON("obs-1", "http://loinc.org", "29463-7", "Weight", map[string]interface{}{
		"value": 85.0,
		"unit":  "kg",
	}))
	resolver.AddQuestionnaire("q-vitals", makeQuestionnaireJSON("q-vitals", "active", "Vitals", []interface{}{
		map[string]interface{}{
			"linkId": "weight",
			"text":   "Body Weight",
			"type":   "quantity",
			"code": []interface{}{
				map[string]interface{}{
					"system":  "http://loinc.org",
					"code":    "29463-7",
					"display": "Body Weight",
				},
			},
		},
	}))

	e := echo.New()
	body := `{"subject":"Patient/p-1"}`
	req := httptest.NewRequest(http.MethodPost, "/fhir/Questionnaire/q-vitals/$populate", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/fhir+json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("q-vitals")

	handler := PopulateHandler(resolver)
	err := handler(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d. Body: %s", rec.Code, rec.Body.String())
	}
}
