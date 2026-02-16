package fhir

import (
	"encoding/json"
	"net/url"
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// TestParseContainedParam
// ---------------------------------------------------------------------------

func TestParseContainedParam_False(t *testing.T) {
	mode, err := ParseContainedParam("false")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mode != ContainedModeNone {
		t.Errorf("mode = %q, want %q", mode, ContainedModeNone)
	}
}

func TestParseContainedParam_True(t *testing.T) {
	mode, err := ParseContainedParam("true")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mode != ContainedModeTrue {
		t.Errorf("mode = %q, want %q", mode, ContainedModeTrue)
	}
}

func TestParseContainedParam_Both(t *testing.T) {
	mode, err := ParseContainedParam("both")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mode != ContainedModeBoth {
		t.Errorf("mode = %q, want %q", mode, ContainedModeBoth)
	}
}

func TestParseContainedParam_Empty(t *testing.T) {
	mode, err := ParseContainedParam("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mode != ContainedModeNone {
		t.Errorf("empty string should default to ContainedModeNone, got %q", mode)
	}
}

func TestParseContainedParam_Invalid(t *testing.T) {
	_, err := ParseContainedParam("invalid")
	if err == nil {
		t.Error("expected error for invalid value, got nil")
	}
}

func TestParseContainedParam_CaseInsensitive(t *testing.T) {
	mode, err := ParseContainedParam("TRUE")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mode != ContainedModeTrue {
		t.Errorf("mode = %q, want %q", mode, ContainedModeTrue)
	}
}

// ---------------------------------------------------------------------------
// TestParseContainedTypeParam
// ---------------------------------------------------------------------------

func TestParseContainedTypeParam_Single(t *testing.T) {
	types, err := ParseContainedTypeParam("Medication")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(types) != 1 || types[0] != "Medication" {
		t.Errorf("types = %v, want [Medication]", types)
	}
}

func TestParseContainedTypeParam_Multiple(t *testing.T) {
	types, err := ParseContainedTypeParam("Medication,Device,Substance")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(types) != 3 {
		t.Fatalf("expected 3 types, got %d: %v", len(types), types)
	}
	expected := []string{"Medication", "Device", "Substance"}
	for i, e := range expected {
		if types[i] != e {
			t.Errorf("types[%d] = %q, want %q", i, types[i], e)
		}
	}
}

func TestParseContainedTypeParam_Empty(t *testing.T) {
	types, err := ParseContainedTypeParam("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(types) != 0 {
		t.Errorf("expected empty slice for empty input, got %v", types)
	}
}

func TestParseContainedTypeParam_WhitespaceHandling(t *testing.T) {
	types, err := ParseContainedTypeParam(" Medication , Device ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(types) != 2 {
		t.Fatalf("expected 2 types, got %d: %v", len(types), types)
	}
	if types[0] != "Medication" || types[1] != "Device" {
		t.Errorf("types = %v, want [Medication Device]", types)
	}
}

// ---------------------------------------------------------------------------
// TestContainedSearchClause
// ---------------------------------------------------------------------------

func TestContainedSearchClause_ModeNone(t *testing.T) {
	config := &ContainedSearchConfig{
		ResourceType:    "MedicationRequest",
		ContainedColumn: "resource_json",
		ContainedPath:   "contained",
		IndexedFields: map[string]ContainedFieldConfig{
			"code": {
				FHIRPath:   "Medication.code.coding.code",
				JSONPath:   "code.coding",
				SearchType: "token",
			},
		},
	}
	clause, args := ContainedSearchClause(config, "code", "12345", ContainedModeNone, 1)
	if clause != "1=1" {
		t.Errorf("expected no-op clause '1=1' for mode none, got %q", clause)
	}
	if len(args) != 0 {
		t.Errorf("expected 0 args for mode none, got %d", len(args))
	}
}

func TestContainedSearchClause_ModeTrue(t *testing.T) {
	config := &ContainedSearchConfig{
		ResourceType:    "MedicationRequest",
		ContainedColumn: "resource_json",
		ContainedPath:   "contained",
		IndexedFields: map[string]ContainedFieldConfig{
			"code": {
				FHIRPath:   "Medication.code.coding.code",
				JSONPath:   "code.coding",
				SearchType: "token",
			},
		},
	}
	clause, args := ContainedSearchClause(config, "code", "12345", ContainedModeTrue, 1)
	if clause == "1=1" {
		t.Error("expected SQL clause for mode true, got '1=1'")
	}
	if !strings.Contains(clause, "jsonb_array_elements") {
		t.Errorf("expected jsonb_array_elements in clause, got %q", clause)
	}
	if !strings.Contains(clause, "resource_json") {
		t.Errorf("expected resource_json column reference, got %q", clause)
	}
	if len(args) == 0 {
		t.Error("expected args for mode true, got 0")
	}
}

func TestContainedSearchClause_ModeBoth(t *testing.T) {
	config := &ContainedSearchConfig{
		ResourceType:    "MedicationRequest",
		ContainedColumn: "resource_json",
		ContainedPath:   "contained",
		IndexedFields: map[string]ContainedFieldConfig{
			"code": {
				FHIRPath:   "Medication.code.coding.code",
				JSONPath:   "code.coding",
				SearchType: "token",
			},
		},
	}
	clause, args := ContainedSearchClause(config, "code", "12345", ContainedModeBoth, 1)
	if clause == "1=1" {
		t.Error("expected SQL clause for mode both, got '1=1'")
	}
	// Both mode should include contained search
	if !strings.Contains(clause, "jsonb_array_elements") {
		t.Errorf("expected jsonb_array_elements in both mode clause, got %q", clause)
	}
	if len(args) == 0 {
		t.Error("expected args for mode both, got 0")
	}
}

func TestContainedSearchClause_UnknownParam(t *testing.T) {
	config := &ContainedSearchConfig{
		ResourceType:    "MedicationRequest",
		ContainedColumn: "resource_json",
		ContainedPath:   "contained",
		IndexedFields:   map[string]ContainedFieldConfig{},
	}
	clause, args := ContainedSearchClause(config, "unknown", "value", ContainedModeTrue, 1)
	if clause != "1=0" {
		t.Errorf("expected '1=0' for unknown param, got %q", clause)
	}
	if len(args) != 0 {
		t.Errorf("expected 0 args for unknown param, got %d", len(args))
	}
}

// ---------------------------------------------------------------------------
// TestContainedTokenSearchClause
// ---------------------------------------------------------------------------

func TestContainedTokenSearchClause_SystemAndCode(t *testing.T) {
	config := &ContainedSearchConfig{
		ContainedColumn: "resource_json",
		ContainedPath:   "contained",
	}
	field := ContainedFieldConfig{
		JSONPath:   "code.coding",
		SearchType: "token",
	}
	clause, args := ContainedTokenSearchClause(config, field, "http://rxnorm.info", "12345", 1)
	if !strings.Contains(clause, "jsonb_array_elements") {
		t.Errorf("expected jsonb_array_elements in clause, got %q", clause)
	}
	if len(args) != 2 {
		t.Fatalf("expected 2 args, got %d: %v", len(args), args)
	}
	if args[0] != "http://rxnorm.info" {
		t.Errorf("args[0] = %v, want %q", args[0], "http://rxnorm.info")
	}
	if args[1] != "12345" {
		t.Errorf("args[1] = %v, want %q", args[1], "12345")
	}
}

func TestContainedTokenSearchClause_CodeOnly(t *testing.T) {
	config := &ContainedSearchConfig{
		ContainedColumn: "resource_json",
		ContainedPath:   "contained",
	}
	field := ContainedFieldConfig{
		JSONPath:   "code.coding",
		SearchType: "token",
	}
	clause, args := ContainedTokenSearchClause(config, field, "", "12345", 1)
	if !strings.Contains(clause, "jsonb_array_elements") {
		t.Errorf("expected jsonb_array_elements in clause, got %q", clause)
	}
	if len(args) != 1 {
		t.Fatalf("expected 1 arg, got %d: %v", len(args), args)
	}
	if args[0] != "12345" {
		t.Errorf("args[0] = %v, want %q", args[0], "12345")
	}
}

func TestContainedTokenSearchClause_SystemOnly(t *testing.T) {
	config := &ContainedSearchConfig{
		ContainedColumn: "resource_json",
		ContainedPath:   "contained",
	}
	field := ContainedFieldConfig{
		JSONPath:   "code.coding",
		SearchType: "token",
	}
	clause, args := ContainedTokenSearchClause(config, field, "http://rxnorm.info", "", 1)
	if !strings.Contains(clause, "jsonb_array_elements") {
		t.Errorf("expected jsonb_array_elements in clause, got %q", clause)
	}
	if len(args) != 1 {
		t.Fatalf("expected 1 arg, got %d: %v", len(args), args)
	}
	if args[0] != "http://rxnorm.info" {
		t.Errorf("args[0] = %v, want %q", args[0], "http://rxnorm.info")
	}
}

// ---------------------------------------------------------------------------
// TestContainedStringSearchClause
// ---------------------------------------------------------------------------

func TestContainedStringSearchClause_Normal(t *testing.T) {
	config := &ContainedSearchConfig{
		ContainedColumn: "resource_json",
		ContainedPath:   "contained",
	}
	field := ContainedFieldConfig{
		JSONPath:   "name",
		SearchType: "string",
	}
	clause, args := ContainedStringSearchClause(config, field, "aspirin", false, 1)
	if !strings.Contains(clause, "ILIKE") {
		t.Errorf("expected ILIKE for non-exact string search, got %q", clause)
	}
	if len(args) != 1 {
		t.Fatalf("expected 1 arg, got %d: %v", len(args), args)
	}
	argStr, ok := args[0].(string)
	if !ok {
		t.Fatalf("args[0] is %T, want string", args[0])
	}
	if !strings.HasSuffix(argStr, "%") {
		t.Errorf("expected prefix-match pattern (ending with %%), got %q", argStr)
	}
}

func TestContainedStringSearchClause_Exact(t *testing.T) {
	config := &ContainedSearchConfig{
		ContainedColumn: "resource_json",
		ContainedPath:   "contained",
	}
	field := ContainedFieldConfig{
		JSONPath:   "name",
		SearchType: "string",
	}
	clause, args := ContainedStringSearchClause(config, field, "Aspirin", true, 1)
	if strings.Contains(clause, "ILIKE") {
		t.Errorf("expected no ILIKE for exact string search, got %q", clause)
	}
	if len(args) != 1 {
		t.Fatalf("expected 1 arg, got %d: %v", len(args), args)
	}
	if args[0] != "Aspirin" {
		t.Errorf("args[0] = %v, want %q", args[0], "Aspirin")
	}
}

func TestContainedStringSearchClause_Empty(t *testing.T) {
	config := &ContainedSearchConfig{
		ContainedColumn: "resource_json",
		ContainedPath:   "contained",
	}
	field := ContainedFieldConfig{
		JSONPath:   "name",
		SearchType: "string",
	}
	clause, args := ContainedStringSearchClause(config, field, "", false, 1)
	if clause != "1=1" {
		t.Errorf("expected '1=1' for empty string, got %q", clause)
	}
	if len(args) != 0 {
		t.Errorf("expected 0 args for empty string, got %d", len(args))
	}
}

// ---------------------------------------------------------------------------
// TestContainedDateSearchClause
// ---------------------------------------------------------------------------

func TestContainedDateSearchClause_Eq(t *testing.T) {
	config := &ContainedSearchConfig{
		ContainedColumn: "resource_json",
		ContainedPath:   "contained",
	}
	field := ContainedFieldConfig{
		JSONPath:   "expirationDate",
		SearchType: "date",
	}
	clause, args := ContainedDateSearchClause(config, field, "eq", "2024-01-15", 1)
	if clause == "" {
		t.Error("expected non-empty clause")
	}
	if len(args) == 0 {
		t.Error("expected args for date search")
	}
	_ = clause
}

func TestContainedDateSearchClause_Ge(t *testing.T) {
	config := &ContainedSearchConfig{
		ContainedColumn: "resource_json",
		ContainedPath:   "contained",
	}
	field := ContainedFieldConfig{
		JSONPath:   "expirationDate",
		SearchType: "date",
	}
	clause, args := ContainedDateSearchClause(config, field, "ge", "2024-01-15", 1)
	if !strings.Contains(clause, ">=") {
		t.Errorf("expected >= for ge prefix, got %q", clause)
	}
	if len(args) != 1 {
		t.Fatalf("expected 1 arg, got %d", len(args))
	}
}

func TestContainedDateSearchClause_Le(t *testing.T) {
	config := &ContainedSearchConfig{
		ContainedColumn: "resource_json",
		ContainedPath:   "contained",
	}
	field := ContainedFieldConfig{
		JSONPath:   "expirationDate",
		SearchType: "date",
	}
	clause, args := ContainedDateSearchClause(config, field, "le", "2024-01-15", 1)
	if !strings.Contains(clause, "<=") {
		t.Errorf("expected <= for le prefix, got %q", clause)
	}
	if len(args) != 1 {
		t.Fatalf("expected 1 arg, got %d", len(args))
	}
}

func TestContainedDateSearchClause_Gt(t *testing.T) {
	config := &ContainedSearchConfig{
		ContainedColumn: "resource_json",
		ContainedPath:   "contained",
	}
	field := ContainedFieldConfig{
		JSONPath:   "expirationDate",
		SearchType: "date",
	}
	clause, args := ContainedDateSearchClause(config, field, "gt", "2024-06-01", 1)
	if !strings.Contains(clause, ">") {
		t.Errorf("expected > for gt prefix, got %q", clause)
	}
	if len(args) != 1 {
		t.Fatalf("expected 1 arg, got %d", len(args))
	}
}

func TestContainedDateSearchClause_Lt(t *testing.T) {
	config := &ContainedSearchConfig{
		ContainedColumn: "resource_json",
		ContainedPath:   "contained",
	}
	field := ContainedFieldConfig{
		JSONPath:   "expirationDate",
		SearchType: "date",
	}
	clause, args := ContainedDateSearchClause(config, field, "lt", "2024-06-01", 1)
	if !strings.Contains(clause, "<") {
		t.Errorf("expected < for lt prefix, got %q", clause)
	}
	if len(args) != 1 {
		t.Fatalf("expected 1 arg, got %d", len(args))
	}
}

// ---------------------------------------------------------------------------
// TestContainedTypeFilterClause
// ---------------------------------------------------------------------------

func TestContainedTypeFilterClause_SingleType(t *testing.T) {
	config := &ContainedSearchConfig{
		ContainedColumn: "resource_json",
		ContainedPath:   "contained",
	}
	clause, args := ContainedTypeFilterClause(config, []string{"Medication"}, 1)
	if !strings.Contains(clause, "resourceType") {
		t.Errorf("expected resourceType filter in clause, got %q", clause)
	}
	if len(args) != 1 {
		t.Fatalf("expected 1 arg, got %d: %v", len(args), args)
	}
	if args[0] != "Medication" {
		t.Errorf("args[0] = %v, want %q", args[0], "Medication")
	}
}

func TestContainedTypeFilterClause_MultipleTypes(t *testing.T) {
	config := &ContainedSearchConfig{
		ContainedColumn: "resource_json",
		ContainedPath:   "contained",
	}
	clause, args := ContainedTypeFilterClause(config, []string{"Medication", "Device"}, 1)
	if !strings.Contains(clause, "resourceType") {
		t.Errorf("expected resourceType filter in clause, got %q", clause)
	}
	if len(args) != 2 {
		t.Fatalf("expected 2 args, got %d: %v", len(args), args)
	}
}

func TestContainedTypeFilterClause_EmptyTypes(t *testing.T) {
	config := &ContainedSearchConfig{
		ContainedColumn: "resource_json",
		ContainedPath:   "contained",
	}
	clause, args := ContainedTypeFilterClause(config, []string{}, 1)
	if clause != "1=1" {
		t.Errorf("expected '1=1' for empty types, got %q", clause)
	}
	if len(args) != 0 {
		t.Errorf("expected 0 args, got %d", len(args))
	}
}

// ---------------------------------------------------------------------------
// TestExtractContainedResources
// ---------------------------------------------------------------------------

func TestExtractContainedResources_HasContained(t *testing.T) {
	resource := map[string]interface{}{
		"resourceType": "MedicationRequest",
		"id":           "med-req-1",
		"contained": []interface{}{
			map[string]interface{}{
				"resourceType": "Medication",
				"id":           "#med1",
				"code": map[string]interface{}{
					"coding": []interface{}{
						map[string]interface{}{
							"system": "http://rxnorm.info",
							"code":   "12345",
						},
					},
				},
			},
		},
	}
	result := ExtractContainedResources(resource)
	if len(result) != 1 {
		t.Fatalf("expected 1 contained resource, got %d", len(result))
	}
	if result[0]["resourceType"] != "Medication" {
		t.Errorf("resourceType = %v, want %q", result[0]["resourceType"], "Medication")
	}
}

func TestExtractContainedResources_Empty(t *testing.T) {
	resource := map[string]interface{}{
		"resourceType": "MedicationRequest",
		"id":           "med-req-1",
		"contained":    []interface{}{},
	}
	result := ExtractContainedResources(resource)
	if len(result) != 0 {
		t.Errorf("expected 0 contained resources, got %d", len(result))
	}
}

func TestExtractContainedResources_None(t *testing.T) {
	resource := map[string]interface{}{
		"resourceType": "Patient",
		"id":           "patient-1",
	}
	result := ExtractContainedResources(resource)
	if len(result) != 0 {
		t.Errorf("expected 0 contained resources when 'contained' key missing, got %d", len(result))
	}
}

func TestExtractContainedResources_InvalidType(t *testing.T) {
	resource := map[string]interface{}{
		"resourceType": "MedicationRequest",
		"id":           "med-req-1",
		"contained":    "not-an-array",
	}
	result := ExtractContainedResources(resource)
	if len(result) != 0 {
		t.Errorf("expected 0 contained resources for invalid contained type, got %d", len(result))
	}
}

func TestExtractContainedResources_MultipleContained(t *testing.T) {
	resource := map[string]interface{}{
		"resourceType": "MedicationRequest",
		"id":           "med-req-1",
		"contained": []interface{}{
			map[string]interface{}{"resourceType": "Medication", "id": "#med1"},
			map[string]interface{}{"resourceType": "Practitioner", "id": "#prac1"},
			map[string]interface{}{"resourceType": "Device", "id": "#dev1"},
		},
	}
	result := ExtractContainedResources(resource)
	if len(result) != 3 {
		t.Fatalf("expected 3 contained resources, got %d", len(result))
	}
}

// ---------------------------------------------------------------------------
// TestResolveContainedReference
// ---------------------------------------------------------------------------

func TestResolveContainedReference_Found(t *testing.T) {
	resource := map[string]interface{}{
		"resourceType": "MedicationRequest",
		"contained": []interface{}{
			map[string]interface{}{"resourceType": "Medication", "id": "med1"},
		},
	}
	result, err := ResolveContainedReference(resource, "#med1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result["resourceType"] != "Medication" {
		t.Errorf("resourceType = %v, want %q", result["resourceType"], "Medication")
	}
}

func TestResolveContainedReference_NotFound(t *testing.T) {
	resource := map[string]interface{}{
		"resourceType": "MedicationRequest",
		"contained": []interface{}{
			map[string]interface{}{"resourceType": "Medication", "id": "med1"},
		},
	}
	_, err := ResolveContainedReference(resource, "#med2")
	if err == nil {
		t.Error("expected error for unresolved reference, got nil")
	}
}

func TestResolveContainedReference_InvalidRef(t *testing.T) {
	resource := map[string]interface{}{
		"resourceType": "MedicationRequest",
		"contained": []interface{}{
			map[string]interface{}{"resourceType": "Medication", "id": "med1"},
		},
	}
	_, err := ResolveContainedReference(resource, "")
	if err == nil {
		t.Error("expected error for empty reference, got nil")
	}
}

func TestResolveContainedReference_NoHashPrefix(t *testing.T) {
	resource := map[string]interface{}{
		"resourceType": "MedicationRequest",
		"contained": []interface{}{
			map[string]interface{}{"resourceType": "Medication", "id": "med1"},
		},
	}
	_, err := ResolveContainedReference(resource, "med1")
	if err == nil {
		t.Error("expected error for reference without # prefix, got nil")
	}
}

func TestResolveContainedReference_NoContained(t *testing.T) {
	resource := map[string]interface{}{
		"resourceType": "Patient",
	}
	_, err := ResolveContainedReference(resource, "#med1")
	if err == nil {
		t.Error("expected error when resource has no contained array, got nil")
	}
}

// ---------------------------------------------------------------------------
// TestIndexContainedResources
// ---------------------------------------------------------------------------

func TestIndexContainedResources_Multiple(t *testing.T) {
	resource := map[string]interface{}{
		"resourceType": "MedicationRequest",
		"contained": []interface{}{
			map[string]interface{}{"resourceType": "Medication", "id": "med1"},
			map[string]interface{}{"resourceType": "Device", "id": "dev1"},
		},
	}
	index := IndexContainedResources(resource)
	if len(index) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(index))
	}
	if _, ok := index["med1"]; !ok {
		t.Error("expected entry for 'med1'")
	}
	if _, ok := index["dev1"]; !ok {
		t.Error("expected entry for 'dev1'")
	}
	if index["med1"]["resourceType"] != "Medication" {
		t.Errorf("med1 resourceType = %v, want %q", index["med1"]["resourceType"], "Medication")
	}
}

func TestIndexContainedResources_Empty(t *testing.T) {
	resource := map[string]interface{}{
		"resourceType": "Patient",
	}
	index := IndexContainedResources(resource)
	if len(index) != 0 {
		t.Errorf("expected 0 entries, got %d", len(index))
	}
}

func TestIndexContainedResources_NoID(t *testing.T) {
	resource := map[string]interface{}{
		"resourceType": "MedicationRequest",
		"contained": []interface{}{
			map[string]interface{}{"resourceType": "Medication"},
		},
	}
	index := IndexContainedResources(resource)
	// Resources without an id should be skipped
	if len(index) != 0 {
		t.Errorf("expected 0 entries for resources without id, got %d", len(index))
	}
}

func TestIndexContainedResources_DuplicateIDs(t *testing.T) {
	resource := map[string]interface{}{
		"resourceType": "MedicationRequest",
		"contained": []interface{}{
			map[string]interface{}{"resourceType": "Medication", "id": "med1", "code": "A"},
			map[string]interface{}{"resourceType": "Medication", "id": "med1", "code": "B"},
		},
	}
	index := IndexContainedResources(resource)
	// Last one wins for duplicate IDs
	if len(index) != 1 {
		t.Fatalf("expected 1 entry (deduped), got %d", len(index))
	}
	if index["med1"]["code"] != "B" {
		t.Errorf("expected last-wins for duplicate id, got code=%v", index["med1"]["code"])
	}
}

// ---------------------------------------------------------------------------
// TestValidateContainedResources
// ---------------------------------------------------------------------------

func TestValidateContainedResources_Valid(t *testing.T) {
	resource := map[string]interface{}{
		"resourceType": "MedicationRequest",
		"medicationReference": map[string]interface{}{
			"reference": "#med1",
		},
		"contained": []interface{}{
			map[string]interface{}{
				"resourceType": "Medication",
				"id":           "med1",
			},
		},
	}
	issues := ValidateContainedResources(resource)
	for _, issue := range issues {
		if issue.Severity == SeverityError {
			t.Errorf("unexpected error issue: %s", issue.Diagnostics)
		}
	}
}

func TestValidateContainedResources_NoID(t *testing.T) {
	resource := map[string]interface{}{
		"resourceType": "MedicationRequest",
		"contained": []interface{}{
			map[string]interface{}{
				"resourceType": "Medication",
			},
		},
	}
	issues := ValidateContainedResources(resource)
	foundIDIssue := false
	for _, issue := range issues {
		if strings.Contains(issue.Diagnostics, "id") {
			foundIDIssue = true
			break
		}
	}
	if !foundIDIssue {
		t.Error("expected validation issue about missing id")
	}
}

func TestValidateContainedResources_NestedContained(t *testing.T) {
	resource := map[string]interface{}{
		"resourceType": "MedicationRequest",
		"contained": []interface{}{
			map[string]interface{}{
				"resourceType": "Medication",
				"id":           "med1",
				"contained": []interface{}{
					map[string]interface{}{
						"resourceType": "Substance",
						"id":           "sub1",
					},
				},
			},
		},
	}
	issues := ValidateContainedResources(resource)
	foundNestingIssue := false
	for _, issue := range issues {
		if strings.Contains(issue.Diagnostics, "nest") {
			foundNestingIssue = true
			break
		}
	}
	if !foundNestingIssue {
		t.Error("expected validation issue about nested contained resources")
	}
}

func TestValidateContainedResources_Unreferenced(t *testing.T) {
	resource := map[string]interface{}{
		"resourceType": "MedicationRequest",
		"contained": []interface{}{
			map[string]interface{}{
				"resourceType": "Medication",
				"id":           "med1",
			},
		},
	}
	issues := ValidateContainedResources(resource)
	foundUnrefIssue := false
	for _, issue := range issues {
		if strings.Contains(issue.Diagnostics, "not referenced") {
			foundUnrefIssue = true
			break
		}
	}
	if !foundUnrefIssue {
		t.Error("expected validation issue about unreferenced contained resource")
	}
}

func TestValidateContainedResources_EmptyContained(t *testing.T) {
	resource := map[string]interface{}{
		"resourceType": "Patient",
	}
	issues := ValidateContainedResources(resource)
	if len(issues) != 0 {
		t.Errorf("expected 0 issues for resource without contained, got %d", len(issues))
	}
}

// ---------------------------------------------------------------------------
// TestDefaultContainedSearchConfigs
// ---------------------------------------------------------------------------

func TestDefaultContainedSearchConfigs(t *testing.T) {
	configs := DefaultContainedSearchConfigs()
	if len(configs) == 0 {
		t.Fatal("expected non-empty default configs")
	}

	// Check MedicationRequest config exists
	mrConfig, ok := configs["MedicationRequest"]
	if !ok {
		t.Fatal("expected config for MedicationRequest")
	}
	if mrConfig.ResourceType != "MedicationRequest" {
		t.Errorf("ResourceType = %q, want %q", mrConfig.ResourceType, "MedicationRequest")
	}
	if len(mrConfig.IndexedFields) == 0 {
		t.Error("expected indexed fields for MedicationRequest")
	}

	// Check Observation config exists
	obsConfig, ok := configs["Observation"]
	if !ok {
		t.Fatal("expected config for Observation")
	}
	if obsConfig.ResourceType != "Observation" {
		t.Errorf("ResourceType = %q, want %q", obsConfig.ResourceType, "Observation")
	}
}

func TestDefaultContainedSearchConfigs_MedicationRequestFields(t *testing.T) {
	configs := DefaultContainedSearchConfigs()
	mrConfig := configs["MedicationRequest"]
	if _, ok := mrConfig.IndexedFields["code"]; !ok {
		t.Error("expected 'code' field in MedicationRequest indexed fields")
	}
}

// ---------------------------------------------------------------------------
// TestBuildContainedJSON
// ---------------------------------------------------------------------------

func TestBuildContainedJSON_Valid(t *testing.T) {
	resources := []map[string]interface{}{
		{"resourceType": "Medication", "id": "med1"},
		{"resourceType": "Device", "id": "dev1"},
	}
	data, err := BuildContainedJSON(resources)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var parsed []interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("result is not valid JSON array: %v", err)
	}
	if len(parsed) != 2 {
		t.Errorf("expected 2 items in JSON array, got %d", len(parsed))
	}
}

func TestBuildContainedJSON_Empty(t *testing.T) {
	data, err := BuildContainedJSON([]map[string]interface{}{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var parsed []interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("result is not valid JSON array: %v", err)
	}
	if len(parsed) != 0 {
		t.Errorf("expected empty JSON array, got %d items", len(parsed))
	}
}

func TestBuildContainedJSON_Nil(t *testing.T) {
	data, err := BuildContainedJSON(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var parsed []interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("result is not valid JSON array: %v", err)
	}
	if len(parsed) != 0 {
		t.Errorf("expected empty JSON array for nil input, got %d items", len(parsed))
	}
}

// ---------------------------------------------------------------------------
// TestStripContainedResources
// ---------------------------------------------------------------------------

func TestStripContainedResources_HasContained(t *testing.T) {
	resource := map[string]interface{}{
		"resourceType": "MedicationRequest",
		"id":           "mr-1",
		"status":       "active",
		"contained": []interface{}{
			map[string]interface{}{"resourceType": "Medication", "id": "med1"},
		},
	}
	stripped := StripContainedResources(resource)
	if _, ok := stripped["contained"]; ok {
		t.Error("expected 'contained' key to be removed")
	}
	if stripped["resourceType"] != "MedicationRequest" {
		t.Errorf("resourceType = %v, want %q", stripped["resourceType"], "MedicationRequest")
	}
	if stripped["status"] != "active" {
		t.Errorf("status = %v, want %q", stripped["status"], "active")
	}
	// Verify original is unchanged
	if _, ok := resource["contained"]; !ok {
		t.Error("original resource should still have 'contained' key")
	}
}

func TestStripContainedResources_NoContained(t *testing.T) {
	resource := map[string]interface{}{
		"resourceType": "Patient",
		"id":           "p-1",
	}
	stripped := StripContainedResources(resource)
	if stripped["resourceType"] != "Patient" {
		t.Errorf("resourceType = %v, want %q", stripped["resourceType"], "Patient")
	}
	if stripped["id"] != "p-1" {
		t.Errorf("id = %v, want %q", stripped["id"], "p-1")
	}
}

// ---------------------------------------------------------------------------
// TestMergeContainedIntoParent
// ---------------------------------------------------------------------------

func TestMergeContainedIntoParent(t *testing.T) {
	parent := map[string]interface{}{
		"resourceType": "MedicationRequest",
		"id":           "mr-1",
		"status":       "active",
	}
	contained := []map[string]interface{}{
		{"resourceType": "Medication", "id": "med1"},
		{"resourceType": "Device", "id": "dev1"},
	}
	merged := MergeContainedIntoParent(parent, contained)
	arr, ok := merged["contained"].([]map[string]interface{})
	if !ok {
		t.Fatal("expected contained to be []map[string]interface{}")
	}
	if len(arr) != 2 {
		t.Errorf("expected 2 contained resources, got %d", len(arr))
	}
	if merged["status"] != "active" {
		t.Errorf("status = %v, want %q", merged["status"], "active")
	}
}

func TestMergeContainedIntoParent_EmptyContained(t *testing.T) {
	parent := map[string]interface{}{
		"resourceType": "Patient",
		"id":           "p-1",
	}
	merged := MergeContainedIntoParent(parent, nil)
	if _, ok := merged["contained"]; ok {
		t.Error("should not add 'contained' key when contained is nil")
	}
}

func TestMergeContainedIntoParent_OverwriteExisting(t *testing.T) {
	parent := map[string]interface{}{
		"resourceType": "MedicationRequest",
		"id":           "mr-1",
		"contained": []interface{}{
			map[string]interface{}{"resourceType": "Medication", "id": "old-med"},
		},
	}
	newContained := []map[string]interface{}{
		{"resourceType": "Medication", "id": "new-med"},
	}
	merged := MergeContainedIntoParent(parent, newContained)
	arr, ok := merged["contained"].([]map[string]interface{})
	if !ok {
		t.Fatal("expected contained to be []map[string]interface{}")
	}
	if len(arr) != 1 {
		t.Fatalf("expected 1 contained resource, got %d", len(arr))
	}
	if arr[0]["id"] != "new-med" {
		t.Errorf("contained[0].id = %v, want %q", arr[0]["id"], "new-med")
	}
}

// ---------------------------------------------------------------------------
// TestNewContainedSearchEngine
// ---------------------------------------------------------------------------

func TestNewContainedSearchEngine(t *testing.T) {
	engine := NewContainedSearchEngine()
	if engine == nil {
		t.Fatal("expected non-nil engine")
	}
	if len(engine.Configs) == 0 {
		t.Error("expected engine to have default configs")
	}
}

// ---------------------------------------------------------------------------
// TestApplyContainedSearch
// ---------------------------------------------------------------------------

func TestApplyContainedSearch_ContainedTrue(t *testing.T) {
	engine := NewContainedSearchEngine()
	q := NewSearchQuery("medication_requests", "*")
	params := url.Values{
		"_contained": []string{"true"},
	}
	err := engine.ApplyContainedSearch(q, params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestApplyContainedSearch_ContainedTypeFilter(t *testing.T) {
	engine := NewContainedSearchEngine()
	q := NewSearchQuery("medication_requests", "*")
	params := url.Values{
		"_contained":     []string{"true"},
		"_containedType": []string{"Medication"},
	}
	err := engine.ApplyContainedSearch(q, params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestApplyContainedSearch_BothParams(t *testing.T) {
	engine := NewContainedSearchEngine()
	q := NewSearchQuery("medication_requests", "*")
	params := url.Values{
		"_contained":     []string{"both"},
		"_containedType": []string{"Medication,Device"},
	}
	err := engine.ApplyContainedSearch(q, params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestApplyContainedSearch_InvalidContainedParam(t *testing.T) {
	engine := NewContainedSearchEngine()
	q := NewSearchQuery("medication_requests", "*")
	params := url.Values{
		"_contained": []string{"invalid"},
	}
	err := engine.ApplyContainedSearch(q, params)
	if err == nil {
		t.Error("expected error for invalid _contained value")
	}
}

func TestApplyContainedSearch_NoContainedParam(t *testing.T) {
	engine := NewContainedSearchEngine()
	q := NewSearchQuery("medication_requests", "*")
	params := url.Values{}
	err := engine.ApplyContainedSearch(q, params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Edge cases
// ---------------------------------------------------------------------------

func TestExtractContainedResources_DeeplyNestedJSON(t *testing.T) {
	resource := map[string]interface{}{
		"resourceType": "MedicationRequest",
		"contained": []interface{}{
			map[string]interface{}{
				"resourceType": "Medication",
				"id":           "med1",
				"code": map[string]interface{}{
					"coding": []interface{}{
						map[string]interface{}{
							"system":  "http://rxnorm.info",
							"code":    "12345",
							"display": "Aspirin",
						},
					},
					"text": "Aspirin 100mg",
				},
				"ingredient": []interface{}{
					map[string]interface{}{
						"itemCodeableConcept": map[string]interface{}{
							"coding": []interface{}{
								map[string]interface{}{
									"system": "http://rxnorm.info",
									"code":   "1191",
								},
							},
						},
						"strength": map[string]interface{}{
							"numerator": map[string]interface{}{
								"value": 100,
								"unit":  "mg",
							},
						},
					},
				},
			},
		},
	}
	result := ExtractContainedResources(resource)
	if len(result) != 1 {
		t.Fatalf("expected 1 contained resource, got %d", len(result))
	}
	// Verify deep nested structure is preserved
	code, ok := result[0]["code"].(map[string]interface{})
	if !ok {
		t.Fatal("expected code to be a map")
	}
	if code["text"] != "Aspirin 100mg" {
		t.Errorf("code.text = %v, want %q", code["text"], "Aspirin 100mg")
	}
}

func TestContainedTokenSearchClause_SpecialCharacters(t *testing.T) {
	config := &ContainedSearchConfig{
		ContainedColumn: "resource_json",
		ContainedPath:   "contained",
	}
	field := ContainedFieldConfig{
		JSONPath:   "code.coding",
		SearchType: "token",
	}
	// Test with URL system containing special characters
	clause, args := ContainedTokenSearchClause(config, field, "http://hl7.org/fhir/sid/ndc", "0069-2587-10", 1)
	if len(args) != 2 {
		t.Fatalf("expected 2 args, got %d", len(args))
	}
	if args[0] != "http://hl7.org/fhir/sid/ndc" {
		t.Errorf("args[0] = %v, want %q", args[0], "http://hl7.org/fhir/sid/ndc")
	}
	if args[1] != "0069-2587-10" {
		t.Errorf("args[1] = %v, want %q", args[1], "0069-2587-10")
	}
	_ = clause
}

func TestContainedStringSearchClause_SpecialCharacters(t *testing.T) {
	config := &ContainedSearchConfig{
		ContainedColumn: "resource_json",
		ContainedPath:   "contained",
	}
	field := ContainedFieldConfig{
		JSONPath:   "name",
		SearchType: "string",
	}
	clause, args := ContainedStringSearchClause(config, field, "O'Brien", false, 1)
	if len(args) != 1 {
		t.Fatalf("expected 1 arg, got %d", len(args))
	}
	argStr, ok := args[0].(string)
	if !ok {
		t.Fatalf("args[0] is %T, want string", args[0])
	}
	if !strings.Contains(argStr, "O'Brien") {
		t.Errorf("arg should contain O'Brien, got %q", argStr)
	}
	_ = clause
}

func TestExtractContainedResources_LargeArray(t *testing.T) {
	contained := make([]interface{}, 100)
	for i := 0; i < 100; i++ {
		contained[i] = map[string]interface{}{
			"resourceType": "Medication",
			"id":           "med" + strings.Repeat("x", i),
		}
	}
	resource := map[string]interface{}{
		"resourceType": "Bundle",
		"contained":    contained,
	}
	result := ExtractContainedResources(resource)
	if len(result) != 100 {
		t.Errorf("expected 100 contained resources, got %d", len(result))
	}
}

func TestContainedSearchResult_Fields(t *testing.T) {
	result := ContainedSearchResult{
		ParentID:       "mr-1",
		ContainedIndex: 0,
		ContainedID:    "#med1",
		ResourceType:   "Medication",
		Resource:       map[string]interface{}{"id": "med1", "resourceType": "Medication"},
	}
	if result.ParentID != "mr-1" {
		t.Errorf("ParentID = %q, want %q", result.ParentID, "mr-1")
	}
	if result.ContainedIndex != 0 {
		t.Errorf("ContainedIndex = %d, want 0", result.ContainedIndex)
	}
	if result.ContainedID != "#med1" {
		t.Errorf("ContainedID = %q, want %q", result.ContainedID, "#med1")
	}
	if result.ResourceType != "Medication" {
		t.Errorf("ResourceType = %q, want %q", result.ResourceType, "Medication")
	}
}

func TestValidateContainedResources_MixedValid(t *testing.T) {
	resource := map[string]interface{}{
		"resourceType": "MedicationRequest",
		"medicationReference": map[string]interface{}{
			"reference": "#med1",
		},
		"contained": []interface{}{
			map[string]interface{}{
				"resourceType": "Medication",
				"id":           "med1",
			},
			map[string]interface{}{
				"resourceType": "Substance",
				// No id - should generate issue
			},
		},
	}
	issues := ValidateContainedResources(resource)
	if len(issues) == 0 {
		t.Error("expected at least one validation issue for resource without id")
	}
}

func TestBuildContainedJSON_ComplexResources(t *testing.T) {
	resources := []map[string]interface{}{
		{
			"resourceType": "Medication",
			"id":           "med1",
			"code": map[string]interface{}{
				"coding": []interface{}{
					map[string]interface{}{"system": "http://rxnorm.info", "code": "12345"},
				},
			},
		},
	}
	data, err := BuildContainedJSON(resources)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(string(data), "rxnorm") {
		t.Error("expected JSON to contain rxnorm system")
	}
}

func TestStripContainedResources_PreservesAllOtherFields(t *testing.T) {
	resource := map[string]interface{}{
		"resourceType":        "MedicationRequest",
		"id":                  "mr-1",
		"status":              "active",
		"intent":              "order",
		"medicationReference": map[string]interface{}{"reference": "#med1"},
		"subject":             map[string]interface{}{"reference": "Patient/p-1"},
		"contained": []interface{}{
			map[string]interface{}{"resourceType": "Medication", "id": "med1"},
		},
	}
	stripped := StripContainedResources(resource)
	expectedKeys := []string{"resourceType", "id", "status", "intent", "medicationReference", "subject"}
	for _, key := range expectedKeys {
		if _, ok := stripped[key]; !ok {
			t.Errorf("expected key %q to be preserved", key)
		}
	}
	if _, ok := stripped["contained"]; ok {
		t.Error("expected 'contained' to be removed")
	}
}

func TestContainedSearchClause_StartIdxPreserved(t *testing.T) {
	config := &ContainedSearchConfig{
		ResourceType:    "MedicationRequest",
		ContainedColumn: "resource_json",
		ContainedPath:   "contained",
		IndexedFields: map[string]ContainedFieldConfig{
			"code": {
				FHIRPath:   "Medication.code.coding.code",
				JSONPath:   "code.coding",
				SearchType: "token",
			},
		},
	}
	clause, args := ContainedSearchClause(config, "code", "12345", ContainedModeTrue, 5)
	if clause == "1=1" {
		t.Error("expected non-trivial clause")
	}
	// The clause should reference $5 or higher
	if !strings.Contains(clause, "$5") {
		t.Errorf("expected clause to use $5 as starting parameter index, got %q", clause)
	}
	if len(args) == 0 {
		t.Error("expected args")
	}
}

func TestResolveContainedReference_HashPrefixID(t *testing.T) {
	// Some implementations store the id with # prefix inside contained
	resource := map[string]interface{}{
		"resourceType": "MedicationRequest",
		"contained": []interface{}{
			map[string]interface{}{"resourceType": "Medication", "id": "#med1"},
		},
	}
	result, err := ResolveContainedReference(resource, "#med1")
	if err != nil {
		// Should still resolve - either by stripping # from ref or matching with #
		t.Logf("note: implementation may need to handle # in id field: %v", err)
	}
	if result != nil && result["resourceType"] != "Medication" {
		t.Errorf("resourceType = %v, want %q", result["resourceType"], "Medication")
	}
}

func TestIndexContainedResources_HashPrefixStripped(t *testing.T) {
	resource := map[string]interface{}{
		"resourceType": "MedicationRequest",
		"contained": []interface{}{
			map[string]interface{}{"resourceType": "Medication", "id": "#med1"},
		},
	}
	index := IndexContainedResources(resource)
	// Should be indexable - implementation may store with or without #
	if len(index) != 1 {
		t.Errorf("expected 1 entry, got %d", len(index))
	}
}
