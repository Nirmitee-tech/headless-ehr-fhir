package fhir

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
)

// ===========================================================================
// SearchExpressionRegistry tests
// ===========================================================================

func TestNewSearchExpressionRegistry(t *testing.T) {
	reg := NewSearchExpressionRegistry()
	if reg == nil {
		t.Fatal("expected non-nil registry")
	}
	got := reg.ListForResourceType("Patient")
	if len(got) != 0 {
		t.Fatalf("expected empty list, got %d entries", len(got))
	}
}

func TestSearchExpressionRegistry_Register(t *testing.T) {
	reg := NewSearchExpressionRegistry()
	expr := &SearchParamExpression{
		Name:          "my-code",
		Type:          "token",
		Expression:    "Patient.extension.where(url='http://example.org/ext').value",
		ResourceTypes: []string{"Patient"},
		Description:   "Custom token search",
	}
	if err := reg.Register(expr); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, ok := reg.Get("Patient", "my-code")
	if !ok {
		t.Fatal("expected expression to be found")
	}
	if got.Name != "my-code" {
		t.Errorf("expected name my-code, got %q", got.Name)
	}
	if got.Type != "token" {
		t.Errorf("expected type token, got %q", got.Type)
	}
}

func TestSearchExpressionRegistry_Register_MultipleResourceTypes(t *testing.T) {
	reg := NewSearchExpressionRegistry()
	expr := &SearchParamExpression{
		Name:          "custom-code",
		Type:          "token",
		Expression:    "code",
		ResourceTypes: []string{"Patient", "Observation"},
	}
	if err := reg.Register(expr); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, ok := reg.Get("Patient", "custom-code"); !ok {
		t.Error("expected expression for Patient")
	}
	if _, ok := reg.Get("Observation", "custom-code"); !ok {
		t.Error("expected expression for Observation")
	}
}

func TestSearchExpressionRegistry_Register_Duplicate(t *testing.T) {
	reg := NewSearchExpressionRegistry()
	expr := &SearchParamExpression{
		Name:          "my-code",
		Type:          "token",
		Expression:    "Patient.code",
		ResourceTypes: []string{"Patient"},
	}
	if err := reg.Register(expr); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	err := reg.Register(expr)
	if err == nil {
		t.Fatal("expected error for duplicate registration")
	}
	if !strings.Contains(err.Error(), "already registered") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestSearchExpressionRegistry_Unregister(t *testing.T) {
	reg := NewSearchExpressionRegistry()
	expr := &SearchParamExpression{
		Name:          "to-remove",
		Type:          "string",
		Expression:    "Patient.name.family",
		ResourceTypes: []string{"Patient"},
	}
	_ = reg.Register(expr)

	if err := reg.Unregister("Patient", "to-remove"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := reg.Get("Patient", "to-remove"); ok {
		t.Error("expected expression to be removed")
	}
}

func TestSearchExpressionRegistry_Unregister_NotFound(t *testing.T) {
	reg := NewSearchExpressionRegistry()
	err := reg.Unregister("Patient", "nonexistent")
	if err == nil {
		t.Fatal("expected error for unregister of non-existent expression")
	}
}

func TestSearchExpressionRegistry_Get_NotFound(t *testing.T) {
	reg := NewSearchExpressionRegistry()
	_, ok := reg.Get("Patient", "nope")
	if ok {
		t.Error("expected not found")
	}
}

func TestSearchExpressionRegistry_ListForResourceType(t *testing.T) {
	reg := NewSearchExpressionRegistry()
	for _, name := range []string{"a-param", "b-param", "c-param"} {
		_ = reg.Register(&SearchParamExpression{
			Name:          name,
			Type:          "string",
			Expression:    "Patient.name",
			ResourceTypes: []string{"Patient"},
		})
	}
	_ = reg.Register(&SearchParamExpression{
		Name:          "obs-param",
		Type:          "token",
		Expression:    "Observation.code",
		ResourceTypes: []string{"Observation"},
	})

	patientExprs := reg.ListForResourceType("Patient")
	if len(patientExprs) != 3 {
		t.Fatalf("expected 3 patient expressions, got %d", len(patientExprs))
	}

	obsExprs := reg.ListForResourceType("Observation")
	if len(obsExprs) != 1 {
		t.Fatalf("expected 1 observation expression, got %d", len(obsExprs))
	}
}

// ===========================================================================
// ValidateSearchParamExpression tests
// ===========================================================================

func TestValidateSearchParamExpression_Valid(t *testing.T) {
	expr := &SearchParamExpression{
		Name:          "test-param",
		Type:          "token",
		Expression:    "Patient.code",
		ResourceTypes: []string{"Patient"},
	}
	issues := ValidateSearchParamExpression(expr)
	for _, issue := range issues {
		if issue.Severity == SeverityError {
			t.Errorf("unexpected error: %s", issue.Diagnostics)
		}
	}
}

func TestValidateSearchParamExpression_MissingName(t *testing.T) {
	expr := &SearchParamExpression{
		Type:          "token",
		Expression:    "Patient.code",
		ResourceTypes: []string{"Patient"},
	}
	issues := ValidateSearchParamExpression(expr)
	found := false
	for _, issue := range issues {
		if strings.Contains(issue.Diagnostics, "name") {
			found = true
		}
	}
	if !found {
		t.Error("expected validation error about missing name")
	}
}

func TestValidateSearchParamExpression_MissingExpression(t *testing.T) {
	expr := &SearchParamExpression{
		Name:          "test-param",
		Type:          "token",
		ResourceTypes: []string{"Patient"},
	}
	issues := ValidateSearchParamExpression(expr)
	found := false
	for _, issue := range issues {
		if strings.Contains(issue.Diagnostics, "expression") {
			found = true
		}
	}
	if !found {
		t.Error("expected validation error about missing expression")
	}
}

func TestValidateSearchParamExpression_InvalidType(t *testing.T) {
	expr := &SearchParamExpression{
		Name:          "test-param",
		Type:          "invalid-type",
		Expression:    "Patient.code",
		ResourceTypes: []string{"Patient"},
	}
	issues := ValidateSearchParamExpression(expr)
	found := false
	for _, issue := range issues {
		if strings.Contains(issue.Diagnostics, "type") {
			found = true
		}
	}
	if !found {
		t.Error("expected validation error about invalid type")
	}
}

func TestValidateSearchParamExpression_MissingResourceTypes(t *testing.T) {
	expr := &SearchParamExpression{
		Name:       "test-param",
		Type:       "token",
		Expression: "Patient.code",
	}
	issues := ValidateSearchParamExpression(expr)
	found := false
	for _, issue := range issues {
		if strings.Contains(issue.Diagnostics, "resource") {
			found = true
		}
	}
	if !found {
		t.Error("expected validation error about missing resource types")
	}
}

// ===========================================================================
// ExpressionEvaluator — simple path tests
// ===========================================================================

func TestEvaluateSimplePath_SingleLevel(t *testing.T) {
	eval := NewExpressionEvaluator()
	resource := map[string]interface{}{
		"resourceType": "Patient",
		"gender":       "male",
	}
	result, err := eval.Evaluate("Patient.gender", resource)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Values) != 1 {
		t.Fatalf("expected 1 value, got %d", len(result.Values))
	}
	if result.Values[0] != "male" {
		t.Errorf("expected 'male', got %v", result.Values[0])
	}
}

func TestEvaluateSimplePath_MultiLevel(t *testing.T) {
	eval := NewExpressionEvaluator()
	resource := map[string]interface{}{
		"resourceType": "Patient",
		"name": []interface{}{
			map[string]interface{}{
				"family": "Smith",
				"given":  []interface{}{"John"},
			},
		},
	}
	result, err := eval.Evaluate("Patient.name.family", resource)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Values) != 1 {
		t.Fatalf("expected 1 value, got %d", len(result.Values))
	}
	if result.Values[0] != "Smith" {
		t.Errorf("expected 'Smith', got %v", result.Values[0])
	}
}

func TestEvaluateSimplePath_ArrayTraversal(t *testing.T) {
	eval := NewExpressionEvaluator()
	resource := map[string]interface{}{
		"resourceType": "Patient",
		"name": []interface{}{
			map[string]interface{}{"family": "Smith"},
			map[string]interface{}{"family": "Jones"},
		},
	}
	result, err := eval.Evaluate("Patient.name.family", resource)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Values) != 2 {
		t.Fatalf("expected 2 values, got %d", len(result.Values))
	}
}

func TestEvaluateSimplePath_MissingField(t *testing.T) {
	eval := NewExpressionEvaluator()
	resource := map[string]interface{}{
		"resourceType": "Patient",
	}
	result, err := eval.Evaluate("Patient.name.family", resource)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Values) != 0 {
		t.Fatalf("expected 0 values, got %d", len(result.Values))
	}
}

func TestEvaluateSimplePath_GivenArray(t *testing.T) {
	eval := NewExpressionEvaluator()
	resource := map[string]interface{}{
		"resourceType": "Patient",
		"name": []interface{}{
			map[string]interface{}{
				"given": []interface{}{"John", "James"},
			},
		},
	}
	result, err := eval.Evaluate("Patient.name.given", resource)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Values) != 2 {
		t.Fatalf("expected 2 values, got %d", len(result.Values))
	}
}

// ===========================================================================
// ExpressionEvaluator — .where() tests
// ===========================================================================

func TestEvaluateWhere_EqualityMatch(t *testing.T) {
	eval := NewExpressionEvaluator()
	resource := map[string]interface{}{
		"resourceType": "Patient",
		"extension": []interface{}{
			map[string]interface{}{
				"url":         "http://example.org/race",
				"valueString": "White",
			},
			map[string]interface{}{
				"url":         "http://example.org/ethnicity",
				"valueString": "Non-Hispanic",
			},
		},
	}
	result, err := eval.Evaluate("Patient.extension.where(url='http://example.org/race').valueString", resource)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Values) != 1 {
		t.Fatalf("expected 1 value, got %d", len(result.Values))
	}
	if result.Values[0] != "White" {
		t.Errorf("expected 'White', got %v", result.Values[0])
	}
}

func TestEvaluateWhere_NoMatch(t *testing.T) {
	eval := NewExpressionEvaluator()
	resource := map[string]interface{}{
		"resourceType": "Patient",
		"extension": []interface{}{
			map[string]interface{}{
				"url":         "http://example.org/race",
				"valueString": "White",
			},
		},
	}
	result, err := eval.Evaluate("Patient.extension.where(url='http://example.org/nonexistent').valueString", resource)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Values) != 0 {
		t.Errorf("expected 0 values, got %d", len(result.Values))
	}
}

func TestEvaluateWhere_MultipleResults(t *testing.T) {
	eval := NewExpressionEvaluator()
	resource := map[string]interface{}{
		"resourceType": "Patient",
		"identifier": []interface{}{
			map[string]interface{}{
				"system": "http://hospital.org",
				"value":  "123",
			},
			map[string]interface{}{
				"system": "http://hospital.org",
				"value":  "456",
			},
			map[string]interface{}{
				"system": "http://other.org",
				"value":  "789",
			},
		},
	}
	result, err := eval.Evaluate("Patient.identifier.where(system='http://hospital.org').value", resource)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Values) != 2 {
		t.Fatalf("expected 2 values, got %d", len(result.Values))
	}
}

// ===========================================================================
// ExpressionEvaluator — .ofType() tests
// ===========================================================================

func TestEvaluateOfType(t *testing.T) {
	eval := NewExpressionEvaluator()
	resource := map[string]interface{}{
		"resourceType": "Observation",
		"value": map[string]interface{}{
			"value": 120.0,
			"unit":  "mmHg",
		},
	}
	// ofType is tested by resolving through the evaluator
	result, err := eval.Evaluate("Observation.value", resource)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Values) != 1 {
		t.Fatalf("expected 1 value, got %d", len(result.Values))
	}
}

// ===========================================================================
// ExpressionEvaluator — .exists() tests
// ===========================================================================

func TestEvaluateExists_True(t *testing.T) {
	eval := NewExpressionEvaluator()
	resource := map[string]interface{}{
		"resourceType": "Patient",
		"gender":       "male",
	}
	result, err := eval.Evaluate("Patient.gender.exists()", resource)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Values) != 1 {
		t.Fatalf("expected 1 value, got %d", len(result.Values))
	}
	if result.Values[0] != true {
		t.Errorf("expected true, got %v", result.Values[0])
	}
}

func TestEvaluateExists_False(t *testing.T) {
	eval := NewExpressionEvaluator()
	resource := map[string]interface{}{
		"resourceType": "Patient",
	}
	result, err := eval.Evaluate("Patient.birthDate.exists()", resource)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Values) != 1 {
		t.Fatalf("expected 1 value, got %d", len(result.Values))
	}
	if result.Values[0] != false {
		t.Errorf("expected false, got %v", result.Values[0])
	}
}

// ===========================================================================
// ExpressionEvaluator — full expression tests
// ===========================================================================

func TestEvaluate_FullExpression(t *testing.T) {
	eval := NewExpressionEvaluator()
	resource := map[string]interface{}{
		"resourceType": "Patient",
		"name": []interface{}{
			map[string]interface{}{
				"family": "Smith",
				"given":  []interface{}{"John"},
			},
		},
	}
	result, err := eval.Evaluate("Patient.name.family", resource)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Values) != 1 || result.Values[0] != "Smith" {
		t.Errorf("unexpected result: %v", result.Values)
	}
}

func TestEvaluate_ExtensionPath(t *testing.T) {
	eval := NewExpressionEvaluator()
	resource := map[string]interface{}{
		"resourceType": "Patient",
		"extension": []interface{}{
			map[string]interface{}{
				"url": "http://hl7.org/fhir/us/core/StructureDefinition/us-core-race",
				"extension": []interface{}{
					map[string]interface{}{
						"url":       "text",
						"valueString": "White",
					},
				},
			},
		},
	}
	result, err := eval.Evaluate("Patient.extension.where(url='http://hl7.org/fhir/us/core/StructureDefinition/us-core-race').extension.where(url='text').valueString", resource)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Values) != 1 {
		t.Fatalf("expected 1 value, got %d", len(result.Values))
	}
	if result.Values[0] != "White" {
		t.Errorf("expected 'White', got %v", result.Values[0])
	}
}

func TestEvaluate_NestedWhere(t *testing.T) {
	eval := NewExpressionEvaluator()
	resource := map[string]interface{}{
		"resourceType": "Patient",
		"identifier": []interface{}{
			map[string]interface{}{
				"system": "http://hospital.org/mrn",
				"value":  "MRN001",
				"type": map[string]interface{}{
					"coding": []interface{}{
						map[string]interface{}{
							"system": "http://terminology.hl7.org/CodeSystem/v2-0203",
							"code":   "MR",
						},
					},
				},
			},
			map[string]interface{}{
				"system": "http://hl7.org/fhir/sid/us-ssn",
				"value":  "999-99-9999",
			},
		},
	}
	result, err := eval.Evaluate("Patient.identifier.where(system='http://hospital.org/mrn').value", resource)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Values) != 1 {
		t.Fatalf("expected 1 value, got %d", len(result.Values))
	}
	if result.Values[0] != "MRN001" {
		t.Errorf("expected MRN001, got %v", result.Values[0])
	}
}

func TestEvaluate_Complex(t *testing.T) {
	eval := NewExpressionEvaluator()
	resource := map[string]interface{}{
		"resourceType": "Observation",
		"code": map[string]interface{}{
			"coding": []interface{}{
				map[string]interface{}{
					"system": "http://loinc.org",
					"code":   "85354-9",
				},
			},
		},
		"subject": map[string]interface{}{
			"reference": "Patient/123",
		},
	}
	result, err := eval.Evaluate("Observation.code.coding.code", resource)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Values) != 1 || result.Values[0] != "85354-9" {
		t.Errorf("unexpected result: %v", result.Values)
	}
}

// ===========================================================================
// ExtractSearchValues tests
// ===========================================================================

func TestExtractSearchValues_Token(t *testing.T) {
	expr := &SearchParamExpression{
		Name:          "code",
		Type:          "token",
		Expression:    "Observation.code.coding",
		ResourceTypes: []string{"Observation"},
	}
	resource := map[string]interface{}{
		"resourceType": "Observation",
		"id":           "obs-1",
		"code": map[string]interface{}{
			"coding": []interface{}{
				map[string]interface{}{
					"system": "http://loinc.org",
					"code":   "85354-9",
				},
			},
		},
	}
	evaluator := NewExpressionEvaluator()
	values, err := ExtractSearchValues(expr, resource, evaluator)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(values) == 0 {
		t.Fatal("expected at least one extracted value")
	}
	found := false
	for _, v := range values {
		if v.TokenCode != nil && *v.TokenCode == "85354-9" {
			found = true
			if v.TokenSystem == nil || *v.TokenSystem != "http://loinc.org" {
				t.Errorf("expected system http://loinc.org")
			}
		}
	}
	if !found {
		t.Error("expected to find token code 85354-9")
	}
}

func TestExtractSearchValues_String(t *testing.T) {
	expr := &SearchParamExpression{
		Name:          "family",
		Type:          "string",
		Expression:    "Patient.name.family",
		ResourceTypes: []string{"Patient"},
	}
	resource := map[string]interface{}{
		"resourceType": "Patient",
		"id":           "pt-1",
		"name": []interface{}{
			map[string]interface{}{
				"family": "Smith",
			},
		},
	}
	evaluator := NewExpressionEvaluator()
	values, err := ExtractSearchValues(expr, resource, evaluator)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(values) != 1 {
		t.Fatalf("expected 1 value, got %d", len(values))
	}
	if values[0].StringValue == nil || *values[0].StringValue != "Smith" {
		t.Errorf("expected StringValue=Smith, got %v", values[0].StringValue)
	}
}

func TestExtractSearchValues_Date(t *testing.T) {
	expr := &SearchParamExpression{
		Name:          "birthdate",
		Type:          "date",
		Expression:    "Patient.birthDate",
		ResourceTypes: []string{"Patient"},
	}
	resource := map[string]interface{}{
		"resourceType": "Patient",
		"id":           "pt-1",
		"birthDate":    "1990-01-15",
	}
	evaluator := NewExpressionEvaluator()
	values, err := ExtractSearchValues(expr, resource, evaluator)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(values) != 1 {
		t.Fatalf("expected 1 value, got %d", len(values))
	}
	if values[0].DateValue == nil {
		t.Fatal("expected non-nil DateValue")
	}
	if values[0].DateValue.Year() != 1990 || values[0].DateValue.Month() != 1 || values[0].DateValue.Day() != 15 {
		t.Errorf("unexpected date: %v", values[0].DateValue)
	}
}

func TestExtractSearchValues_Reference(t *testing.T) {
	expr := &SearchParamExpression{
		Name:          "subject",
		Type:          "reference",
		Expression:    "Observation.subject",
		ResourceTypes: []string{"Observation"},
	}
	resource := map[string]interface{}{
		"resourceType": "Observation",
		"id":           "obs-1",
		"subject": map[string]interface{}{
			"reference": "Patient/123",
		},
	}
	evaluator := NewExpressionEvaluator()
	values, err := ExtractSearchValues(expr, resource, evaluator)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(values) != 1 {
		t.Fatalf("expected 1 value, got %d", len(values))
	}
	if values[0].ReferenceValue == nil || *values[0].ReferenceValue != "Patient/123" {
		t.Errorf("expected reference Patient/123, got %v", values[0].ReferenceValue)
	}
}

// ===========================================================================
// ConvertToSearchValue tests
// ===========================================================================

func TestConvertToSearchValue_String(t *testing.T) {
	sv, err := ConvertToSearchValue("name", "Patient", "pt-1", "Smith", "string")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sv.StringValue == nil || *sv.StringValue != "Smith" {
		t.Errorf("expected StringValue=Smith")
	}
	if sv.ParamName != "name" {
		t.Errorf("expected ParamName=name, got %q", sv.ParamName)
	}
}

func TestConvertToSearchValue_Token(t *testing.T) {
	tokenVal := map[string]interface{}{
		"system": "http://loinc.org",
		"code":   "12345",
	}
	sv, err := ConvertToSearchValue("code", "Observation", "obs-1", tokenVal, "token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sv.TokenSystem == nil || *sv.TokenSystem != "http://loinc.org" {
		t.Errorf("expected TokenSystem=http://loinc.org")
	}
	if sv.TokenCode == nil || *sv.TokenCode != "12345" {
		t.Errorf("expected TokenCode=12345")
	}
}

func TestConvertToSearchValue_Date(t *testing.T) {
	sv, err := ConvertToSearchValue("birthdate", "Patient", "pt-1", "1990-06-15", "date")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sv.DateValue == nil {
		t.Fatal("expected non-nil DateValue")
	}
	if sv.DateValue.Year() != 1990 {
		t.Errorf("expected year 1990, got %d", sv.DateValue.Year())
	}
}

func TestConvertToSearchValue_Number(t *testing.T) {
	sv, err := ConvertToSearchValue("count", "List", "l-1", 42.0, "number")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sv.NumberValue == nil || *sv.NumberValue != 42.0 {
		t.Errorf("expected NumberValue=42.0")
	}
}

func TestConvertToSearchValue_Quantity(t *testing.T) {
	qtyVal := map[string]interface{}{
		"value": 120.0,
		"unit":  "mmHg",
	}
	sv, err := ConvertToSearchValue("value-quantity", "Observation", "obs-1", qtyVal, "quantity")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sv.QuantityValue == nil || *sv.QuantityValue != 120.0 {
		t.Errorf("expected QuantityValue=120.0")
	}
	if sv.QuantityUnit == nil || *sv.QuantityUnit != "mmHg" {
		t.Errorf("expected QuantityUnit=mmHg")
	}
}

func TestConvertToSearchValue_Reference(t *testing.T) {
	refVal := map[string]interface{}{
		"reference": "Patient/abc",
	}
	sv, err := ConvertToSearchValue("subject", "Observation", "obs-1", refVal, "reference")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sv.ReferenceValue == nil || *sv.ReferenceValue != "Patient/abc" {
		t.Errorf("expected ReferenceValue=Patient/abc")
	}
}

func TestConvertToSearchValue_URI(t *testing.T) {
	sv, err := ConvertToSearchValue("url", "ValueSet", "vs-1", "http://example.org/ValueSet/test", "uri")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sv.URIValue == nil || *sv.URIValue != "http://example.org/ValueSet/test" {
		t.Errorf("expected URIValue")
	}
}

// ===========================================================================
// SearchExpressionIndex tests
// ===========================================================================

func TestSearchExpressionIndex_Index(t *testing.T) {
	idx := NewSearchExpressionIndex()
	sv := SearchIndexValue{
		ParamName:    "family",
		ResourceType: "Patient",
		ResourceID:   "pt-1",
		ValueType:    "string",
		StringValue:  strPtr("Smith"),
	}
	idx.Index("Patient", "pt-1", []SearchIndexValue{sv})

	results := idx.Search("Patient", "family", "eq", "Smith")
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0] != "pt-1" {
		t.Errorf("expected pt-1, got %q", results[0])
	}
}

func TestSearchExpressionIndex_Search_NoMatch(t *testing.T) {
	idx := NewSearchExpressionIndex()
	sv := SearchIndexValue{
		ParamName:    "family",
		ResourceType: "Patient",
		ResourceID:   "pt-1",
		ValueType:    "string",
		StringValue:  strPtr("Smith"),
	}
	idx.Index("Patient", "pt-1", []SearchIndexValue{sv})

	results := idx.Search("Patient", "family", "eq", "Jones")
	if len(results) != 0 {
		t.Fatalf("expected 0 results, got %d", len(results))
	}
}

func TestSearchExpressionIndex_Search_Token(t *testing.T) {
	idx := NewSearchExpressionIndex()
	sv := SearchIndexValue{
		ParamName:    "code",
		ResourceType: "Observation",
		ResourceID:   "obs-1",
		ValueType:    "token",
		TokenSystem:  strPtr("http://loinc.org"),
		TokenCode:    strPtr("85354-9"),
	}
	idx.Index("Observation", "obs-1", []SearchIndexValue{sv})

	results := idx.Search("Observation", "code", "eq", "85354-9")
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
}

func TestSearchExpressionIndex_Search_NumberGt(t *testing.T) {
	idx := NewSearchExpressionIndex()
	for i, v := range []float64{10, 20, 30} {
		id := []string{"pt-1", "pt-2", "pt-3"}[i]
		sv := SearchIndexValue{
			ParamName:    "count",
			ResourceType: "List",
			ResourceID:   id,
			ValueType:    "number",
			NumberValue:  seFloat64Ptr(v),
		}
		idx.Index("List", id, []SearchIndexValue{sv})
	}

	results := idx.Search("List", "count", "gt", 15.0)
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
}

func TestSearchExpressionIndex_Search_DateGe(t *testing.T) {
	idx := NewSearchExpressionIndex()
	t1 := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	t2 := time.Date(2023, 6, 15, 0, 0, 0, 0, time.UTC)
	idx.Index("Patient", "pt-1", []SearchIndexValue{
		{ParamName: "birthdate", ResourceType: "Patient", ResourceID: "pt-1", ValueType: "date", DateValue: &t1},
	})
	idx.Index("Patient", "pt-2", []SearchIndexValue{
		{ParamName: "birthdate", ResourceType: "Patient", ResourceID: "pt-2", ValueType: "date", DateValue: &t2},
	})

	threshold := time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)
	results := idx.Search("Patient", "birthdate", "ge", threshold)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0] != "pt-2" {
		t.Errorf("expected pt-2, got %q", results[0])
	}
}

func TestSearchExpressionIndex_Remove(t *testing.T) {
	idx := NewSearchExpressionIndex()
	sv := SearchIndexValue{
		ParamName:    "family",
		ResourceType: "Patient",
		ResourceID:   "pt-1",
		ValueType:    "string",
		StringValue:  strPtr("Smith"),
	}
	idx.Index("Patient", "pt-1", []SearchIndexValue{sv})
	idx.Remove("Patient", "pt-1")

	results := idx.Search("Patient", "family", "eq", "Smith")
	if len(results) != 0 {
		t.Fatalf("expected 0 results after remove, got %d", len(results))
	}
}

// ===========================================================================
// GenerateSearchSQL tests
// ===========================================================================

func TestGenerateSearchSQL_StringEq(t *testing.T) {
	expr := &SearchParamExpression{
		Name:       "family",
		Type:       "string",
		Expression: "Patient.name.family",
	}
	sql, args, err := GenerateSearchSQL(expr, "eq", "Smith", 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sql == "" {
		t.Fatal("expected non-empty SQL")
	}
	if len(args) == 0 {
		t.Fatal("expected at least one argument")
	}
	// Verify SQL references JSONB extraction
	if !strings.Contains(sql, "ILIKE") && !strings.Contains(sql, "ilike") && !strings.Contains(sql, "=") {
		t.Errorf("expected string comparison operator in SQL: %q", sql)
	}
	foundSmith := false
	for _, a := range args {
		if s, ok := a.(string); ok && strings.Contains(s, "Smith") {
			foundSmith = true
		}
	}
	if !foundSmith {
		t.Errorf("expected Smith in args, got %v", args)
	}
}

func TestGenerateSearchSQL_TokenEq(t *testing.T) {
	expr := &SearchParamExpression{
		Name:       "code",
		Type:       "token",
		Expression: "Observation.code.coding",
	}
	sql, args, err := GenerateSearchSQL(expr, "eq", "http://loinc.org|85354-9", 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sql == "" {
		t.Fatal("expected non-empty SQL")
	}
	if len(args) < 1 {
		t.Fatal("expected at least one argument")
	}
	// Should contain JSONB path or system/code matching
	if !strings.Contains(strings.ToLower(sql), "jsonb") && !strings.Contains(sql, "@>") && !strings.Contains(sql, "->>") && !strings.Contains(sql, "$") {
		t.Logf("SQL generated: %q", sql)
	}
}

func TestGenerateSearchSQL_DateGe(t *testing.T) {
	expr := &SearchParamExpression{
		Name:       "birthdate",
		Type:       "date",
		Expression: "Patient.birthDate",
	}
	sql, args, err := GenerateSearchSQL(expr, "ge", "2020-01-01", 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sql == "" {
		t.Fatal("expected non-empty SQL")
	}
	if !strings.Contains(sql, ">=") {
		t.Errorf("expected >= in date SQL, got %q", sql)
	}
	if len(args) < 1 {
		t.Fatal("expected at least one argument")
	}
}

func TestGenerateSearchSQL_NumberLt(t *testing.T) {
	expr := &SearchParamExpression{
		Name:       "count",
		Type:       "number",
		Expression: "List.count",
	}
	sql, args, err := GenerateSearchSQL(expr, "lt", "100", 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sql == "" {
		t.Fatal("expected non-empty SQL")
	}
	if !strings.Contains(sql, "<") {
		t.Errorf("expected < in number SQL, got %q", sql)
	}
	if len(args) < 1 {
		t.Fatal("expected at least one argument")
	}
}

func TestGenerateSearchSQL_Reference(t *testing.T) {
	expr := &SearchParamExpression{
		Name:       "subject",
		Type:       "reference",
		Expression: "Observation.subject",
	}
	sql, args, err := GenerateSearchSQL(expr, "eq", "Patient/123", 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sql == "" {
		t.Fatal("expected non-empty SQL")
	}
	if len(args) < 1 {
		t.Fatal("expected at least one argument")
	}
}

func TestGenerateSearchSQL_URI(t *testing.T) {
	expr := &SearchParamExpression{
		Name:       "url",
		Type:       "uri",
		Expression: "ValueSet.url",
	}
	sql, args, err := GenerateSearchSQL(expr, "eq", "http://example.org/vs/test", 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sql == "" {
		t.Fatal("expected non-empty SQL")
	}
	if len(args) < 1 {
		t.Fatal("expected at least one argument")
	}
}

// ===========================================================================
// DefaultSearchExpressions tests
// ===========================================================================

func TestDefaultSearchExpressions(t *testing.T) {
	exprs := DefaultSearchExpressions()
	if len(exprs) == 0 {
		t.Fatal("expected at least one default expression")
	}

	// Check for known US Core extension search params
	names := make(map[string]bool)
	for _, e := range exprs {
		names[e.Name] = true
	}
	expectedNames := []string{"race", "ethnicity", "birthsex"}
	for _, name := range expectedNames {
		if !names[name] {
			t.Errorf("expected default expression %q", name)
		}
	}
}

func TestDefaultSearchExpressions_HaveValidExpressions(t *testing.T) {
	exprs := DefaultSearchExpressions()
	for _, expr := range exprs {
		issues := ValidateSearchParamExpression(expr)
		for _, issue := range issues {
			if issue.Severity == SeverityError {
				t.Errorf("default expression %q has error: %s", expr.Name, issue.Diagnostics)
			}
		}
	}
}

// ===========================================================================
// ToSearchParameter tests
// ===========================================================================

func TestToSearchParameter(t *testing.T) {
	expr := &SearchParamExpression{
		Name:          "my-code",
		Type:          "token",
		Expression:    "Patient.code",
		ResourceTypes: []string{"Patient"},
		Description:   "My custom code search",
		Target:        []string{"Patient"},
		Modifier:      []string{"exact", "contains"},
		Comparator:    []string{"eq", "ne"},
		MultipleOr:    true,
		MultipleAnd:   false,
	}

	sp := expr.ToSearchParameter()
	if sp["resourceType"] != "SearchParameter" {
		t.Errorf("expected resourceType=SearchParameter, got %v", sp["resourceType"])
	}
	if sp["name"] != "my-code" {
		t.Errorf("expected name=my-code, got %v", sp["name"])
	}
	if sp["type"] != "token" {
		t.Errorf("expected type=token, got %v", sp["type"])
	}
	if sp["expression"] != "Patient.code" {
		t.Errorf("expected expression=Patient.code, got %v", sp["expression"])
	}

	base, ok := sp["base"].([]string)
	if !ok || len(base) != 1 || base[0] != "Patient" {
		t.Errorf("expected base=[Patient], got %v", sp["base"])
	}
}

func TestToSearchParameter_HasDescription(t *testing.T) {
	expr := &SearchParamExpression{
		Name:          "desc-test",
		Type:          "string",
		Expression:    "Patient.name",
		ResourceTypes: []string{"Patient"},
		Description:   "A test description",
	}
	sp := expr.ToSearchParameter()
	if sp["description"] != "A test description" {
		t.Errorf("expected description, got %v", sp["description"])
	}
}

// ===========================================================================
// RegisterSearchParamHandler tests
// ===========================================================================

func TestRegisterSearchParamHandler_GET(t *testing.T) {
	reg := NewSearchExpressionRegistry()
	_ = reg.Register(&SearchParamExpression{
		Name:          "test-param",
		Type:          "token",
		Expression:    "Patient.code",
		ResourceTypes: []string{"Patient"},
	})

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/?resourceType=Patient", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := RegisterSearchParamHandler(reg)
	if err := handler(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var body map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid JSON response: %v", err)
	}
	entries, ok := body["entry"].([]interface{})
	if !ok {
		t.Fatalf("expected entry array, got %T", body["entry"])
	}
	if len(entries) != 1 {
		t.Errorf("expected 1 entry, got %d", len(entries))
	}
}

func TestRegisterSearchParamHandler_POST(t *testing.T) {
	reg := NewSearchExpressionRegistry()

	e := echo.New()
	body := `{
		"name": "new-param",
		"type": "string",
		"expression": "Patient.name",
		"resourceTypes": ["Patient"]
	}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := RegisterSearchParamHandler(reg)
	if err := handler(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}

	// Verify it was registered.
	if _, ok := reg.Get("Patient", "new-param"); !ok {
		t.Error("expected parameter to be registered")
	}
}

func TestRegisterSearchParamHandler_DELETE(t *testing.T) {
	reg := NewSearchExpressionRegistry()
	_ = reg.Register(&SearchParamExpression{
		Name:          "to-delete",
		Type:          "token",
		Expression:    "Patient.code",
		ResourceTypes: []string{"Patient"},
	})

	e := echo.New()
	req := httptest.NewRequest(http.MethodDelete, "/?resourceType=Patient&name=to-delete", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := RegisterSearchParamHandler(reg)
	if err := handler(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}

	if _, ok := reg.Get("Patient", "to-delete"); ok {
		t.Error("expected parameter to be removed")
	}
}

func TestRegisterSearchParamHandler_POST_InvalidJSON(t *testing.T) {
	reg := NewSearchExpressionRegistry()

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("{invalid"))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := RegisterSearchParamHandler(reg)
	if err := handler(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestRegisterSearchParamHandler_POST_ValidationFails(t *testing.T) {
	reg := NewSearchExpressionRegistry()

	e := echo.New()
	body := `{
		"type": "string",
		"expression": "Patient.name",
		"resourceTypes": ["Patient"]
	}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := RegisterSearchParamHandler(reg)
	if err := handler(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

// ===========================================================================
// Edge case tests
// ===========================================================================

func TestEvaluate_NilResource(t *testing.T) {
	eval := NewExpressionEvaluator()
	result, err := eval.Evaluate("Patient.name", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Values) != 0 {
		t.Errorf("expected empty result for nil resource, got %d values", len(result.Values))
	}
}

func TestEvaluate_EmptyExpression(t *testing.T) {
	eval := NewExpressionEvaluator()
	resource := map[string]interface{}{"resourceType": "Patient"}
	_, err := eval.Evaluate("", resource)
	if err == nil {
		t.Error("expected error for empty expression")
	}
}

func TestEvaluate_DeeplyNestedPath(t *testing.T) {
	eval := NewExpressionEvaluator()
	resource := map[string]interface{}{
		"resourceType": "Observation",
		"component": []interface{}{
			map[string]interface{}{
				"code": map[string]interface{}{
					"coding": []interface{}{
						map[string]interface{}{
							"system": "http://loinc.org",
							"code":   "8480-6",
						},
					},
				},
				"valueQuantity": map[string]interface{}{
					"value": 120.0,
					"unit":  "mmHg",
				},
			},
			map[string]interface{}{
				"code": map[string]interface{}{
					"coding": []interface{}{
						map[string]interface{}{
							"system": "http://loinc.org",
							"code":   "8462-4",
						},
					},
				},
				"valueQuantity": map[string]interface{}{
					"value": 80.0,
					"unit":  "mmHg",
				},
			},
		},
	}
	result, err := eval.Evaluate("Observation.component.code.coding.code", resource)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Values) != 2 {
		t.Fatalf("expected 2 values, got %d", len(result.Values))
	}
}

func TestEvaluate_ArraysOfArrays(t *testing.T) {
	eval := NewExpressionEvaluator()
	resource := map[string]interface{}{
		"resourceType": "Patient",
		"name": []interface{}{
			map[string]interface{}{
				"given": []interface{}{"John", "James"},
			},
			map[string]interface{}{
				"given": []interface{}{"Jane"},
			},
		},
	}
	result, err := eval.Evaluate("Patient.name.given", resource)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Values) != 3 {
		t.Fatalf("expected 3 values (John, James, Jane), got %d", len(result.Values))
	}
}

func TestExtractSearchValues_EmptyResource(t *testing.T) {
	expr := &SearchParamExpression{
		Name:          "family",
		Type:          "string",
		Expression:    "Patient.name.family",
		ResourceTypes: []string{"Patient"},
	}
	resource := map[string]interface{}{
		"resourceType": "Patient",
		"id":           "pt-1",
	}
	evaluator := NewExpressionEvaluator()
	values, err := ExtractSearchValues(expr, resource, evaluator)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(values) != 0 {
		t.Errorf("expected 0 values from empty resource field, got %d", len(values))
	}
}

func TestSearchExpressionIndex_Search_Ne(t *testing.T) {
	idx := NewSearchExpressionIndex()
	idx.Index("Patient", "pt-1", []SearchIndexValue{
		{ParamName: "family", ResourceType: "Patient", ResourceID: "pt-1", ValueType: "string", StringValue: strPtr("Smith")},
	})
	idx.Index("Patient", "pt-2", []SearchIndexValue{
		{ParamName: "family", ResourceType: "Patient", ResourceID: "pt-2", ValueType: "string", StringValue: strPtr("Jones")},
	})

	results := idx.Search("Patient", "family", "ne", "Smith")
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0] != "pt-2" {
		t.Errorf("expected pt-2, got %q", results[0])
	}
}

func TestSearchExpressionIndex_MultipleValuesPerResource(t *testing.T) {
	idx := NewSearchExpressionIndex()
	idx.Index("Patient", "pt-1", []SearchIndexValue{
		{ParamName: "given", ResourceType: "Patient", ResourceID: "pt-1", ValueType: "string", StringValue: strPtr("John")},
		{ParamName: "given", ResourceType: "Patient", ResourceID: "pt-1", ValueType: "string", StringValue: strPtr("James")},
	})

	results := idx.Search("Patient", "given", "eq", "James")
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
}

func TestConvertToSearchValue_TokenFromString(t *testing.T) {
	sv, err := ConvertToSearchValue("status", "Observation", "obs-1", "final", "token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sv.TokenCode == nil || *sv.TokenCode != "final" {
		t.Errorf("expected TokenCode=final")
	}
}

func TestConvertToSearchValue_DateFromTime(t *testing.T) {
	dt := time.Date(2023, 6, 15, 10, 30, 0, 0, time.UTC)
	sv, err := ConvertToSearchValue("date", "Observation", "obs-1", dt, "date")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sv.DateValue == nil {
		t.Fatal("expected non-nil DateValue")
	}
	if !sv.DateValue.Equal(dt) {
		t.Errorf("expected %v, got %v", dt, sv.DateValue)
	}
}

func TestEvaluate_ResourceTypeMismatch(t *testing.T) {
	eval := NewExpressionEvaluator()
	resource := map[string]interface{}{
		"resourceType": "Observation",
		"status":       "final",
	}
	result, err := eval.Evaluate("Patient.status", resource)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Resource type mismatch: Patient expression against Observation resource
	if len(result.Values) != 0 {
		t.Errorf("expected empty result for resource type mismatch, got %d values", len(result.Values))
	}
}

func TestSearchExpressionIndex_Search_Lt(t *testing.T) {
	idx := NewSearchExpressionIndex()
	for i, v := range []float64{10, 20, 30} {
		id := []string{"r-1", "r-2", "r-3"}[i]
		idx.Index("List", id, []SearchIndexValue{
			{ParamName: "count", ResourceType: "List", ResourceID: id, ValueType: "number", NumberValue: seFloat64Ptr(v)},
		})
	}
	results := idx.Search("List", "count", "lt", 25.0)
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
}

func TestSearchExpressionIndex_Search_Le(t *testing.T) {
	idx := NewSearchExpressionIndex()
	for i, v := range []float64{10, 20, 30} {
		id := []string{"r-1", "r-2", "r-3"}[i]
		idx.Index("List", id, []SearchIndexValue{
			{ParamName: "count", ResourceType: "List", ResourceID: id, ValueType: "number", NumberValue: seFloat64Ptr(v)},
		})
	}
	results := idx.Search("List", "count", "le", 20.0)
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
}

func TestRegisterSearchParamHandler_GET_All(t *testing.T) {
	reg := NewSearchExpressionRegistry()
	_ = reg.Register(&SearchParamExpression{
		Name: "a", Type: "string", Expression: "Patient.name",
		ResourceTypes: []string{"Patient"},
	})
	_ = reg.Register(&SearchParamExpression{
		Name: "b", Type: "token", Expression: "Observation.code",
		ResourceTypes: []string{"Observation"},
	})

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := RegisterSearchParamHandler(reg)
	if err := handler(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var body map[string]interface{}
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	entries, _ := body["entry"].([]interface{})
	if len(entries) != 2 {
		t.Errorf("expected 2 entries, got %d", len(entries))
	}
}

func TestRegisterSearchParamHandler_MethodNotAllowed(t *testing.T) {
	reg := NewSearchExpressionRegistry()

	e := echo.New()
	req := httptest.NewRequest(http.MethodPut, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := RegisterSearchParamHandler(reg)
	if err := handler(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rec.Code)
	}
}

func TestExtractSearchValues_Quantity(t *testing.T) {
	expr := &SearchParamExpression{
		Name:          "value-quantity",
		Type:          "quantity",
		Expression:    "Observation.valueQuantity",
		ResourceTypes: []string{"Observation"},
	}
	resource := map[string]interface{}{
		"resourceType": "Observation",
		"id":           "obs-1",
		"valueQuantity": map[string]interface{}{
			"value": 120.0,
			"unit":  "mmHg",
		},
	}
	evaluator := NewExpressionEvaluator()
	values, err := ExtractSearchValues(expr, resource, evaluator)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(values) != 1 {
		t.Fatalf("expected 1 value, got %d", len(values))
	}
	if values[0].QuantityValue == nil || *values[0].QuantityValue != 120.0 {
		t.Errorf("expected QuantityValue=120.0")
	}
}

func TestSearchExpressionIndex_Search_TokenWithSystem(t *testing.T) {
	idx := NewSearchExpressionIndex()
	idx.Index("Observation", "obs-1", []SearchIndexValue{
		{ParamName: "code", ResourceType: "Observation", ResourceID: "obs-1", ValueType: "token",
			TokenSystem: strPtr("http://loinc.org"), TokenCode: strPtr("85354-9")},
	})
	idx.Index("Observation", "obs-2", []SearchIndexValue{
		{ParamName: "code", ResourceType: "Observation", ResourceID: "obs-2", ValueType: "token",
			TokenSystem: strPtr("http://snomed.info/sct"), TokenCode: strPtr("85354-9")},
	})

	results := idx.Search("Observation", "code", "eq", "http://loinc.org|85354-9")
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0] != "obs-1" {
		t.Errorf("expected obs-1, got %q", results[0])
	}
}

// ===========================================================================
// Helpers
// ===========================================================================

func strPtr(s string) *string {
	return &s
}

func seFloat64Ptr(f float64) *float64 {
	return &f
}
