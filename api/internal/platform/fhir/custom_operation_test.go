package fhir

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
)

// ===========================================================================
// Helper: create a valid CustomOperationDef for testing
// ===========================================================================

func sampleOperationDef() *CustomOperationDef {
	return &CustomOperationDef{
		Name:        "$calculate-risk",
		Code:        "calculate-risk",
		Title:       "Calculate Risk Score",
		Description: "Calculates a risk score for a patient based on clinical data",
		Scope:       OperationScopeType | OperationScopeInstance,
		ResourceTypes: []string{"Patient"},
		Parameters: []OperationParamDef{
			{Name: "subject", Use: "in", Min: 1, Max: "1", Type: "Reference", Required: true, Documentation: "Patient reference"},
			{Name: "period", Use: "in", Min: 0, Max: "1", Type: "Period", Documentation: "Assessment period"},
			{Name: "score", Use: "out", Min: 1, Max: "1", Type: "decimal", Documentation: "Risk score"},
			{Name: "assessment", Use: "out", Min: 0, Max: "*", Type: "string", Documentation: "Assessment details"},
		},
		AffectsState: false,
		Idempotent:   true,
		System:       false,
		Type:         true,
		Instance:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
}

func sampleHandler() OperationHandler {
	return func(ctx *OperationContext) (*OperationResponse, error) {
		return &OperationResponse{
			StatusCode: http.StatusOK,
			Resource: map[string]interface{}{
				"resourceType": "Parameters",
				"parameter": []interface{}{
					map[string]interface{}{
						"name":         "score",
						"valueDecimal": 0.75,
					},
				},
			},
			ContentType: "application/fhir+json",
		}, nil
	}
}

// ===========================================================================
// TestNewCustomOperationRegistry
// ===========================================================================

func TestNewCustomOperationRegistry(t *testing.T) {
	reg := NewCustomOperationRegistry()
	if reg == nil {
		t.Fatal("expected non-nil registry")
	}
	ops := reg.List()
	if len(ops) != 0 {
		t.Errorf("expected empty registry, got %d operations", len(ops))
	}
}

// ===========================================================================
// TestRegister
// ===========================================================================

func TestRegister_Valid(t *testing.T) {
	reg := NewCustomOperationRegistry()
	def := sampleOperationDef()
	err := reg.Register(def, sampleHandler())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ops := reg.List()
	if len(ops) != 1 {
		t.Fatalf("expected 1 operation, got %d", len(ops))
	}
	if ops[0].Code != "calculate-risk" {
		t.Errorf("expected code 'calculate-risk', got %q", ops[0].Code)
	}
}

func TestRegister_Duplicate(t *testing.T) {
	reg := NewCustomOperationRegistry()
	def := sampleOperationDef()
	err := reg.Register(def, sampleHandler())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = reg.Register(def, sampleHandler())
	if err == nil {
		t.Fatal("expected error for duplicate registration")
	}
	if !strings.Contains(err.Error(), "already registered") {
		t.Errorf("expected 'already registered' in error, got: %v", err)
	}
}

func TestRegister_InvalidDef(t *testing.T) {
	reg := NewCustomOperationRegistry()
	// Missing name and code
	def := &CustomOperationDef{}
	err := reg.Register(def, sampleHandler())
	if err == nil {
		t.Fatal("expected error for invalid definition")
	}
}

func TestRegister_NilHandler(t *testing.T) {
	reg := NewCustomOperationRegistry()
	def := sampleOperationDef()
	err := reg.Register(def, nil)
	if err == nil {
		t.Fatal("expected error for nil handler")
	}
	if !strings.Contains(err.Error(), "handler") {
		t.Errorf("expected 'handler' in error, got: %v", err)
	}
}

func TestRegister_NilDef(t *testing.T) {
	reg := NewCustomOperationRegistry()
	err := reg.Register(nil, sampleHandler())
	if err == nil {
		t.Fatal("expected error for nil definition")
	}
}

// ===========================================================================
// TestUnregister
// ===========================================================================

func TestUnregister_Exists(t *testing.T) {
	reg := NewCustomOperationRegistry()
	def := sampleOperationDef()
	_ = reg.Register(def, sampleHandler())

	err := reg.Unregister("calculate-risk")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ops := reg.List()
	if len(ops) != 0 {
		t.Errorf("expected 0 operations after unregister, got %d", len(ops))
	}
}

func TestUnregister_NotFound(t *testing.T) {
	reg := NewCustomOperationRegistry()
	err := reg.Unregister("nonexistent")
	if err == nil {
		t.Fatal("expected error for unregistering nonexistent operation")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' in error, got: %v", err)
	}
}

// ===========================================================================
// TestGet
// ===========================================================================

func TestGet_Exists(t *testing.T) {
	reg := NewCustomOperationRegistry()
	def := sampleOperationDef()
	_ = reg.Register(def, sampleHandler())

	got, ok := reg.Get("calculate-risk")
	if !ok {
		t.Fatal("expected operation to be found")
	}
	if got.Code != "calculate-risk" {
		t.Errorf("expected code 'calculate-risk', got %q", got.Code)
	}
}

func TestGet_NotFound(t *testing.T) {
	reg := NewCustomOperationRegistry()
	_, ok := reg.Get("nonexistent")
	if ok {
		t.Fatal("expected operation to not be found")
	}
}

// ===========================================================================
// TestList
// ===========================================================================

func TestList_Empty(t *testing.T) {
	reg := NewCustomOperationRegistry()
	ops := reg.List()
	if len(ops) != 0 {
		t.Errorf("expected 0 operations, got %d", len(ops))
	}
}

func TestList_Multiple(t *testing.T) {
	reg := NewCustomOperationRegistry()

	def1 := sampleOperationDef()
	def1.Name = "$alpha"
	def1.Code = "alpha"
	_ = reg.Register(def1, sampleHandler())

	def2 := sampleOperationDef()
	def2.Name = "$beta"
	def2.Code = "beta"
	_ = reg.Register(def2, sampleHandler())

	def3 := sampleOperationDef()
	def3.Name = "$gamma"
	def3.Code = "gamma"
	_ = reg.Register(def3, sampleHandler())

	ops := reg.List()
	if len(ops) != 3 {
		t.Fatalf("expected 3 operations, got %d", len(ops))
	}

	// Should be sorted alphabetically by code
	if ops[0].Code != "alpha" || ops[1].Code != "beta" || ops[2].Code != "gamma" {
		t.Errorf("expected alphabetical order, got %q, %q, %q", ops[0].Code, ops[1].Code, ops[2].Code)
	}
}

// ===========================================================================
// TestListForResourceType
// ===========================================================================

func TestListForResourceType_Matching(t *testing.T) {
	reg := NewCustomOperationRegistry()

	def1 := sampleOperationDef()
	def1.Code = "patient-op"
	def1.Name = "$patient-op"
	def1.ResourceTypes = []string{"Patient"}
	_ = reg.Register(def1, sampleHandler())

	def2 := sampleOperationDef()
	def2.Code = "obs-op"
	def2.Name = "$obs-op"
	def2.ResourceTypes = []string{"Observation"}
	_ = reg.Register(def2, sampleHandler())

	def3 := sampleOperationDef()
	def3.Code = "both-op"
	def3.Name = "$both-op"
	def3.ResourceTypes = []string{"Patient", "Observation"}
	_ = reg.Register(def3, sampleHandler())

	// All-types operation (empty ResourceTypes)
	def4 := sampleOperationDef()
	def4.Code = "all-op"
	def4.Name = "$all-op"
	def4.ResourceTypes = nil
	_ = reg.Register(def4, sampleHandler())

	patientOps := reg.ListForResourceType("Patient")
	// Should match: patient-op, both-op, all-op
	if len(patientOps) != 3 {
		t.Errorf("expected 3 Patient operations, got %d", len(patientOps))
	}

	obsOps := reg.ListForResourceType("Observation")
	// Should match: obs-op, both-op, all-op
	if len(obsOps) != 3 {
		t.Errorf("expected 3 Observation operations, got %d", len(obsOps))
	}
}

func TestListForResourceType_None(t *testing.T) {
	reg := NewCustomOperationRegistry()

	def := sampleOperationDef()
	def.ResourceTypes = []string{"Patient"}
	_ = reg.Register(def, sampleHandler())

	ops := reg.ListForResourceType("Encounter")
	if len(ops) != 0 {
		t.Errorf("expected 0 operations for Encounter, got %d", len(ops))
	}
}

// ===========================================================================
// TestListByScope
// ===========================================================================

func TestListByScope_System(t *testing.T) {
	reg := NewCustomOperationRegistry()

	def := sampleOperationDef()
	def.Code = "sys-op"
	def.Name = "$sys-op"
	def.Scope = OperationScopeSystem
	_ = reg.Register(def, sampleHandler())

	def2 := sampleOperationDef()
	def2.Code = "type-op"
	def2.Name = "$type-op"
	def2.Scope = OperationScopeType
	_ = reg.Register(def2, sampleHandler())

	ops := reg.ListByScope(OperationScopeSystem)
	if len(ops) != 1 {
		t.Fatalf("expected 1 system operation, got %d", len(ops))
	}
	if ops[0].Code != "sys-op" {
		t.Errorf("expected sys-op, got %q", ops[0].Code)
	}
}

func TestListByScope_Type(t *testing.T) {
	reg := NewCustomOperationRegistry()

	def := sampleOperationDef()
	def.Code = "type-op"
	def.Name = "$type-op"
	def.Scope = OperationScopeType
	_ = reg.Register(def, sampleHandler())

	ops := reg.ListByScope(OperationScopeType)
	if len(ops) != 1 {
		t.Fatalf("expected 1 type operation, got %d", len(ops))
	}
}

func TestListByScope_Instance(t *testing.T) {
	reg := NewCustomOperationRegistry()

	def := sampleOperationDef()
	def.Code = "inst-op"
	def.Name = "$inst-op"
	def.Scope = OperationScopeInstance
	_ = reg.Register(def, sampleHandler())

	ops := reg.ListByScope(OperationScopeInstance)
	if len(ops) != 1 {
		t.Fatalf("expected 1 instance operation, got %d", len(ops))
	}
}

func TestListByScope_Combined(t *testing.T) {
	reg := NewCustomOperationRegistry()

	def := sampleOperationDef()
	def.Code = "multi-op"
	def.Name = "$multi-op"
	def.Scope = OperationScopeSystem | OperationScopeType | OperationScopeInstance
	_ = reg.Register(def, sampleHandler())

	sysOps := reg.ListByScope(OperationScopeSystem)
	if len(sysOps) != 1 {
		t.Errorf("expected 1 system operation, got %d", len(sysOps))
	}

	typeOps := reg.ListByScope(OperationScopeType)
	if len(typeOps) != 1 {
		t.Errorf("expected 1 type operation, got %d", len(typeOps))
	}

	instOps := reg.ListByScope(OperationScopeInstance)
	if len(instOps) != 1 {
		t.Errorf("expected 1 instance operation, got %d", len(instOps))
	}
}

// ===========================================================================
// TestValidateOperationDef
// ===========================================================================

func TestValidateOperationDef_Valid(t *testing.T) {
	def := sampleOperationDef()
	issues := ValidateOperationDef(def)
	for _, issue := range issues {
		if issue.Severity == SeverityError || issue.Severity == SeverityFatal {
			t.Errorf("unexpected error issue: %s", issue.Diagnostics)
		}
	}
}

func TestValidateOperationDef_MissingName(t *testing.T) {
	def := sampleOperationDef()
	def.Name = ""
	issues := ValidateOperationDef(def)
	found := false
	for _, issue := range issues {
		if strings.Contains(issue.Diagnostics, "name") || strings.Contains(issue.Diagnostics, "Name") {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected validation issue for missing name")
	}
}

func TestValidateOperationDef_MissingCode(t *testing.T) {
	def := sampleOperationDef()
	def.Code = ""
	issues := ValidateOperationDef(def)
	found := false
	for _, issue := range issues {
		if strings.Contains(issue.Diagnostics, "code") || strings.Contains(issue.Diagnostics, "Code") {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected validation issue for missing code")
	}
}

func TestValidateOperationDef_InvalidParamUse(t *testing.T) {
	def := sampleOperationDef()
	def.Parameters = []OperationParamDef{
		{Name: "badparam", Use: "invalid", Min: 0, Max: "1", Type: "string"},
	}
	issues := ValidateOperationDef(def)
	found := false
	for _, issue := range issues {
		if strings.Contains(issue.Diagnostics, "use") || strings.Contains(issue.Diagnostics, "Use") {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected validation issue for invalid param use")
	}
}

func TestValidateOperationDef_InvalidParamMax(t *testing.T) {
	def := sampleOperationDef()
	def.Parameters = []OperationParamDef{
		{Name: "badparam", Use: "in", Min: 0, Max: "abc", Type: "string"},
	}
	issues := ValidateOperationDef(def)
	found := false
	for _, issue := range issues {
		if strings.Contains(issue.Diagnostics, "max") || strings.Contains(issue.Diagnostics, "Max") {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected validation issue for invalid param max")
	}
}

func TestValidateOperationDef_MissingParamName(t *testing.T) {
	def := sampleOperationDef()
	def.Parameters = []OperationParamDef{
		{Name: "", Use: "in", Min: 0, Max: "1", Type: "string"},
	}
	issues := ValidateOperationDef(def)
	found := false
	for _, issue := range issues {
		if strings.Contains(issue.Diagnostics, "parameter name") || strings.Contains(issue.Diagnostics, "Parameter name") {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected validation issue for missing param name")
	}
}

func TestValidateOperationDef_MissingParamType(t *testing.T) {
	def := sampleOperationDef()
	def.Parameters = []OperationParamDef{
		{Name: "notype", Use: "in", Min: 0, Max: "1", Type: ""},
	}
	issues := ValidateOperationDef(def)
	found := false
	for _, issue := range issues {
		if strings.Contains(issue.Diagnostics, "type") || strings.Contains(issue.Diagnostics, "Type") {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected validation issue for missing param type")
	}
}

func TestValidateOperationDef_NilDef(t *testing.T) {
	issues := ValidateOperationDef(nil)
	if len(issues) == 0 {
		t.Error("expected at least one issue for nil definition")
	}
}

func TestValidateOperationDef_CodeWithDollarSign(t *testing.T) {
	def := sampleOperationDef()
	def.Code = "$calculate-risk" // Code should not have $
	issues := ValidateOperationDef(def)
	found := false
	for _, issue := range issues {
		if strings.Contains(issue.Diagnostics, "$") {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected validation issue for code containing $")
	}
}

func TestValidateOperationDef_NameWithoutDollar(t *testing.T) {
	def := sampleOperationDef()
	def.Name = "calculate-risk" // Name should have $
	issues := ValidateOperationDef(def)
	found := false
	for _, issue := range issues {
		if strings.Contains(issue.Diagnostics, "$") {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected validation issue for name missing $")
	}
}

// ===========================================================================
// TestParseOperationParameters
// ===========================================================================

func TestParseOperationParameters_Valid(t *testing.T) {
	body := map[string]interface{}{
		"resourceType": "Parameters",
		"parameter": []interface{}{
			map[string]interface{}{
				"name":        "subject",
				"valueString": "Patient/123",
			},
			map[string]interface{}{
				"name":         "score",
				"valueDecimal": 0.5,
			},
		},
	}

	params, err := ParseOperationParameters(body)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if params["subject"] != "Patient/123" {
		t.Errorf("expected subject=Patient/123, got %v", params["subject"])
	}
	if params["score"] != 0.5 {
		t.Errorf("expected score=0.5, got %v", params["score"])
	}
}

func TestParseOperationParameters_NestedResource(t *testing.T) {
	body := map[string]interface{}{
		"resourceType": "Parameters",
		"parameter": []interface{}{
			map[string]interface{}{
				"name": "resource",
				"resource": map[string]interface{}{
					"resourceType": "Patient",
					"id":           "123",
				},
			},
		},
	}

	params, err := ParseOperationParameters(body)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	res, ok := params["resource"].(map[string]interface{})
	if !ok {
		t.Fatal("expected resource param to be a map")
	}
	if res["resourceType"] != "Patient" {
		t.Errorf("expected Patient resource type, got %v", res["resourceType"])
	}
}

func TestParseOperationParameters_Empty(t *testing.T) {
	body := map[string]interface{}{
		"resourceType": "Parameters",
		"parameter":    []interface{}{},
	}

	params, err := ParseOperationParameters(body)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(params) != 0 {
		t.Errorf("expected 0 params, got %d", len(params))
	}
}

func TestParseOperationParameters_InvalidResourceType(t *testing.T) {
	body := map[string]interface{}{
		"resourceType": "Patient",
		"parameter":    []interface{}{},
	}

	_, err := ParseOperationParameters(body)
	if err == nil {
		t.Fatal("expected error for non-Parameters resource type")
	}
}

func TestParseOperationParameters_NilBody(t *testing.T) {
	_, err := ParseOperationParameters(nil)
	if err == nil {
		t.Fatal("expected error for nil body")
	}
}

func TestParseOperationParameters_MissingParameterField(t *testing.T) {
	body := map[string]interface{}{
		"resourceType": "Parameters",
	}

	params, err := ParseOperationParameters(body)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(params) != 0 {
		t.Errorf("expected 0 params, got %d", len(params))
	}
}

func TestParseOperationParameters_BooleanValue(t *testing.T) {
	body := map[string]interface{}{
		"resourceType": "Parameters",
		"parameter": []interface{}{
			map[string]interface{}{
				"name":         "persist",
				"valueBoolean": true,
			},
		},
	}

	params, err := ParseOperationParameters(body)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if params["persist"] != true {
		t.Errorf("expected persist=true, got %v", params["persist"])
	}
}

func TestParseOperationParameters_IntegerValue(t *testing.T) {
	body := map[string]interface{}{
		"resourceType": "Parameters",
		"parameter": []interface{}{
			map[string]interface{}{
				"name":         "count",
				"valueInteger": float64(10),
			},
		},
	}

	params, err := ParseOperationParameters(body)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if params["count"] != float64(10) {
		t.Errorf("expected count=10, got %v", params["count"])
	}
}

// ===========================================================================
// TestBuildParametersResource
// ===========================================================================

func TestBuildParametersResource_VariousTypes(t *testing.T) {
	params := map[string]interface{}{
		"name":    "test",
		"score":   0.75,
		"active":  true,
		"count":   float64(5),
	}

	result := BuildParametersResource(params)
	if result["resourceType"] != "Parameters" {
		t.Errorf("expected resourceType=Parameters, got %v", result["resourceType"])
	}

	paramList, ok := result["parameter"].([]interface{})
	if !ok {
		t.Fatal("expected parameter to be a slice")
	}
	if len(paramList) != 4 {
		t.Errorf("expected 4 parameters, got %d", len(paramList))
	}
}

func TestBuildParametersResource_EmptyMap(t *testing.T) {
	result := BuildParametersResource(map[string]interface{}{})
	if result["resourceType"] != "Parameters" {
		t.Error("expected resourceType=Parameters")
	}
	paramList, ok := result["parameter"].([]interface{})
	if !ok {
		t.Fatal("expected parameter to be a slice")
	}
	if len(paramList) != 0 {
		t.Errorf("expected 0 parameters, got %d", len(paramList))
	}
}

func TestBuildParametersResource_NilMap(t *testing.T) {
	result := BuildParametersResource(nil)
	if result["resourceType"] != "Parameters" {
		t.Error("expected resourceType=Parameters")
	}
}

func TestBuildParametersResource_ResourceValue(t *testing.T) {
	params := map[string]interface{}{
		"return": map[string]interface{}{
			"resourceType": "Bundle",
			"type":         "searchset",
		},
	}

	result := BuildParametersResource(params)
	paramList := result["parameter"].([]interface{})
	if len(paramList) != 1 {
		t.Fatalf("expected 1 parameter, got %d", len(paramList))
	}
	p := paramList[0].(map[string]interface{})
	if p["name"] != "return" {
		t.Errorf("expected name=return, got %v", p["name"])
	}
	// Resource-typed values should be stored under "resource" key
	if _, ok := p["resource"]; !ok {
		t.Error("expected 'resource' key for map value")
	}
}

// ===========================================================================
// TestValidateOperationInput
// ===========================================================================

func TestValidateOperationInput_Valid(t *testing.T) {
	def := sampleOperationDef()
	params := map[string]interface{}{
		"subject": "Patient/123",
	}
	issues := ValidateOperationInput(def, params)
	for _, issue := range issues {
		if issue.Severity == SeverityError || issue.Severity == SeverityFatal {
			t.Errorf("unexpected error: %s", issue.Diagnostics)
		}
	}
}

func TestValidateOperationInput_MissingRequired(t *testing.T) {
	def := sampleOperationDef()
	params := map[string]interface{}{
		"period": "2024-01-01",
	}
	issues := ValidateOperationInput(def, params)
	found := false
	for _, issue := range issues {
		if strings.Contains(issue.Diagnostics, "subject") && issue.Severity == SeverityError {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected error for missing required 'subject' parameter")
	}
}

func TestValidateOperationInput_ExtraParams(t *testing.T) {
	def := sampleOperationDef()
	params := map[string]interface{}{
		"subject":   "Patient/123",
		"unknown":   "value",
	}
	issues := ValidateOperationInput(def, params)
	found := false
	for _, issue := range issues {
		if strings.Contains(issue.Diagnostics, "unknown") {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected warning for unknown parameter")
	}
}

func TestValidateOperationInput_EmptyParams(t *testing.T) {
	def := sampleOperationDef()
	issues := ValidateOperationInput(def, map[string]interface{}{})
	// Should flag required 'subject' as missing
	found := false
	for _, issue := range issues {
		if strings.Contains(issue.Diagnostics, "subject") {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected error for missing required parameter")
	}
}

func TestValidateOperationInput_NilParams(t *testing.T) {
	def := sampleOperationDef()
	issues := ValidateOperationInput(def, nil)
	found := false
	for _, issue := range issues {
		if issue.Severity == SeverityError {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected errors for nil params with required fields")
	}
}

// ===========================================================================
// TestValidateOperationOutput
// ===========================================================================

func TestValidateOperationOutput_Valid(t *testing.T) {
	def := sampleOperationDef()
	result := map[string]interface{}{
		"score": 0.75,
	}
	issues := ValidateOperationOutput(def, result)
	for _, issue := range issues {
		if issue.Severity == SeverityError {
			t.Errorf("unexpected error: %s", issue.Diagnostics)
		}
	}
}

func TestValidateOperationOutput_MissingRequired(t *testing.T) {
	def := sampleOperationDef()
	result := map[string]interface{}{
		"assessment": "low risk",
	}
	issues := ValidateOperationOutput(def, result)
	found := false
	for _, issue := range issues {
		if strings.Contains(issue.Diagnostics, "score") && issue.Severity == SeverityError {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected error for missing required 'score' output parameter")
	}
}

// ===========================================================================
// TestRouteOperation
// ===========================================================================

func TestRouteOperation_System(t *testing.T) {
	reg := NewCustomOperationRegistry()
	def := sampleOperationDef()
	def.Code = "sys-op"
	def.Name = "$sys-op"
	def.Scope = OperationScopeSystem
	def.System = true
	def.Type = false
	def.Instance = false
	_ = reg.Register(def, sampleHandler())

	opDef, inv, err := reg.RouteOperation("POST", "/fhir/$sys-op")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if opDef.Code != "sys-op" {
		t.Errorf("expected sys-op, got %q", opDef.Code)
	}
	if inv.Scope != OperationScopeSystem {
		t.Errorf("expected system scope, got %d", inv.Scope)
	}
	if inv.ResourceType != "" {
		t.Errorf("expected empty resource type, got %q", inv.ResourceType)
	}
}

func TestRouteOperation_Type(t *testing.T) {
	reg := NewCustomOperationRegistry()
	def := sampleOperationDef()
	def.Scope = OperationScopeType
	def.Type = true
	_ = reg.Register(def, sampleHandler())

	opDef, inv, err := reg.RouteOperation("POST", "/fhir/Patient/$calculate-risk")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if opDef.Code != "calculate-risk" {
		t.Errorf("expected calculate-risk, got %q", opDef.Code)
	}
	if inv.ResourceType != "Patient" {
		t.Errorf("expected Patient, got %q", inv.ResourceType)
	}
	if inv.Scope != OperationScopeType {
		t.Errorf("expected type scope, got %d", inv.Scope)
	}
}

func TestRouteOperation_Instance(t *testing.T) {
	reg := NewCustomOperationRegistry()
	def := sampleOperationDef()
	def.Scope = OperationScopeInstance
	def.Instance = true
	_ = reg.Register(def, sampleHandler())

	opDef, inv, err := reg.RouteOperation("POST", "/fhir/Patient/123/$calculate-risk")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if opDef.Code != "calculate-risk" {
		t.Errorf("expected calculate-risk, got %q", opDef.Code)
	}
	if inv.ResourceType != "Patient" {
		t.Errorf("expected Patient, got %q", inv.ResourceType)
	}
	if inv.ResourceID != "123" {
		t.Errorf("expected resource ID 123, got %q", inv.ResourceID)
	}
	if inv.Scope != OperationScopeInstance {
		t.Errorf("expected instance scope, got %d", inv.Scope)
	}
}

func TestRouteOperation_Unknown(t *testing.T) {
	reg := NewCustomOperationRegistry()
	_, _, err := reg.RouteOperation("POST", "/fhir/$nonexistent")
	if err == nil {
		t.Fatal("expected error for unknown operation")
	}
}

func TestRouteOperation_WrongMethod(t *testing.T) {
	reg := NewCustomOperationRegistry()
	def := sampleOperationDef()
	def.AffectsState = true
	_ = reg.Register(def, sampleHandler())

	_, _, err := reg.RouteOperation("GET", "/fhir/Patient/$calculate-risk")
	if err == nil {
		t.Fatal("expected error for GET on affects-state operation")
	}
	if !strings.Contains(err.Error(), "POST") {
		t.Errorf("expected error to mention POST, got: %v", err)
	}
}

func TestRouteOperation_GetAllowed(t *testing.T) {
	reg := NewCustomOperationRegistry()
	def := sampleOperationDef()
	def.AffectsState = false
	def.Scope = OperationScopeSystem
	def.System = true
	_ = reg.Register(def, sampleHandler())

	opDef, _, err := reg.RouteOperation("GET", "/fhir/$calculate-risk")
	if err != nil {
		t.Fatalf("unexpected error for GET on non-affects-state operation: %v", err)
	}
	if opDef.Code != "calculate-risk" {
		t.Errorf("expected calculate-risk, got %q", opDef.Code)
	}
}

func TestRouteOperation_NonOperationPath(t *testing.T) {
	reg := NewCustomOperationRegistry()
	_, _, err := reg.RouteOperation("GET", "/fhir/Patient/123")
	if err == nil {
		t.Fatal("expected error for non-operation path")
	}
}

// ===========================================================================
// TestCustomOperationHandler (HTTP handler)
// ===========================================================================

func TestCustomOperationHandler_Success(t *testing.T) {
	reg := NewCustomOperationRegistry()
	def := sampleOperationDef()
	def.Scope = OperationScopeSystem
	def.System = true
	def.AffectsState = false
	_ = reg.Register(def, sampleHandler())

	e := echo.New()
	reqBody := `{"resourceType":"Parameters","parameter":[{"name":"subject","valueString":"Patient/123"}]}`
	req := httptest.NewRequest(http.MethodPost, "/fhir/$calculate-risk", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/fhir/$calculate-risk")

	handler := CustomOperationHandler(reg)
	err := handler(c)
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp["resourceType"] != "Parameters" {
		t.Errorf("expected Parameters response, got %v", resp["resourceType"])
	}
}

func TestCustomOperationHandler_NotFound(t *testing.T) {
	reg := NewCustomOperationRegistry()

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/fhir/$nonexistent", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/fhir/$nonexistent")

	handler := CustomOperationHandler(reg)
	err := handler(c)
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestCustomOperationHandler_MethodNotAllowed(t *testing.T) {
	reg := NewCustomOperationRegistry()
	def := sampleOperationDef()
	def.AffectsState = true
	def.Scope = OperationScopeSystem
	def.System = true
	_ = reg.Register(def, sampleHandler())

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/fhir/$calculate-risk", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/fhir/$calculate-risk")

	handler := CustomOperationHandler(reg)
	err := handler(c)
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rec.Code)
	}
}

func TestCustomOperationHandler_HandlerError(t *testing.T) {
	reg := NewCustomOperationRegistry()
	def := sampleOperationDef()
	def.Scope = OperationScopeSystem
	def.System = true
	def.AffectsState = false

	errHandler := func(ctx *OperationContext) (*OperationResponse, error) {
		return nil, fmt.Errorf("internal computation failed")
	}
	_ = reg.Register(def, errHandler)

	e := echo.New()
	reqBody := `{"resourceType":"Parameters","parameter":[{"name":"subject","valueString":"Patient/123"}]}`
	req := httptest.NewRequest(http.MethodPost, "/fhir/$calculate-risk", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/fhir/$calculate-risk")

	handler := CustomOperationHandler(reg)
	err := handler(c)
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rec.Code)
	}
}

func TestCustomOperationHandler_GetWithQueryParams(t *testing.T) {
	reg := NewCustomOperationRegistry()
	def := sampleOperationDef()
	def.AffectsState = false
	def.Scope = OperationScopeSystem
	def.System = true

	qHandler := func(ctx *OperationContext) (*OperationResponse, error) {
		subj := ctx.Invocation.Parameters["subject"]
		return &OperationResponse{
			StatusCode: http.StatusOK,
			Resource: map[string]interface{}{
				"resourceType": "Parameters",
				"parameter": []interface{}{
					map[string]interface{}{
						"name":        "result",
						"valueString": fmt.Sprintf("received:%v", subj),
					},
				},
			},
		}, nil
	}
	_ = reg.Register(def, qHandler)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/fhir/$calculate-risk?subject=Patient/123", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/fhir/$calculate-risk")

	handler := CustomOperationHandler(reg)
	err := handler(c)
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

// ===========================================================================
// TestCustomOperationMiddleware
// ===========================================================================

func TestCustomOperationMiddleware_RoutesCorrectly(t *testing.T) {
	reg := NewCustomOperationRegistry()
	def := sampleOperationDef()
	def.Scope = OperationScopeSystem
	def.System = true
	def.AffectsState = false
	_ = reg.Register(def, sampleHandler())

	e := echo.New()
	e.Use(CustomOperationMiddleware(reg))
	e.POST("/fhir/*", func(c echo.Context) error {
		return c.String(http.StatusOK, "fallthrough")
	})
	e.GET("/fhir/*", func(c echo.Context) error {
		return c.String(http.StatusOK, "fallthrough")
	})

	reqBody := `{"resourceType":"Parameters","parameter":[{"name":"subject","valueString":"Patient/123"}]}`
	req := httptest.NewRequest(http.MethodPost, "/fhir/$calculate-risk", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp["resourceType"] != "Parameters" {
		t.Errorf("expected Parameters in response, got %v", resp["resourceType"])
	}
}

func TestCustomOperationMiddleware_PassesThroughNonOperation(t *testing.T) {
	reg := NewCustomOperationRegistry()

	e := echo.New()
	e.Use(CustomOperationMiddleware(reg))
	e.GET("/fhir/Patient/:id", func(c echo.Context) error {
		return c.String(http.StatusOK, "patient-resource")
	})

	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient/123", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	if rec.Body.String() != "patient-resource" {
		t.Errorf("expected 'patient-resource', got %q", rec.Body.String())
	}
}

// ===========================================================================
// TestToOperationDefinition
// ===========================================================================

func TestToOperationDefinition(t *testing.T) {
	def := sampleOperationDef()
	opDef := def.ToOperationDefinition()

	if opDef["resourceType"] != "OperationDefinition" {
		t.Errorf("expected resourceType=OperationDefinition, got %v", opDef["resourceType"])
	}
	if opDef["code"] != "calculate-risk" {
		t.Errorf("expected code=calculate-risk, got %v", opDef["code"])
	}
	if opDef["name"] != "$calculate-risk" {
		t.Errorf("expected name=$calculate-risk, got %v", opDef["name"])
	}
	if opDef["kind"] != "operation" {
		t.Errorf("expected kind=operation, got %v", opDef["kind"])
	}
	if opDef["system"] != false {
		t.Errorf("expected system=false, got %v", opDef["system"])
	}
	if opDef["type"] != true {
		t.Errorf("expected type=true, got %v", opDef["type"])
	}
	if opDef["instance"] != true {
		t.Errorf("expected instance=true, got %v", opDef["instance"])
	}
	if opDef["affectsState"] != false {
		t.Errorf("expected affectsState=false, got %v", opDef["affectsState"])
	}

	params, ok := opDef["parameter"].([]map[string]interface{})
	if !ok {
		t.Fatal("expected parameter to be a slice of maps")
	}
	if len(params) != 4 {
		t.Errorf("expected 4 parameters, got %d", len(params))
	}

	resources, ok := opDef["resource"].([]string)
	if !ok {
		t.Fatal("expected resource to be a string slice")
	}
	if len(resources) != 1 || resources[0] != "Patient" {
		t.Errorf("expected [Patient], got %v", resources)
	}
}

func TestToOperationDefinition_NoResourceTypes(t *testing.T) {
	def := sampleOperationDef()
	def.ResourceTypes = nil
	opDef := def.ToOperationDefinition()

	if _, ok := opDef["resource"]; ok {
		t.Error("expected no 'resource' key when ResourceTypes is nil")
	}
}

func TestToOperationDefinition_RoundTrip(t *testing.T) {
	def := sampleOperationDef()
	opDef := def.ToOperationDefinition()

	// Marshal to JSON and back to verify it's valid JSON
	data, err := json.Marshal(opDef)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if parsed["resourceType"] != "OperationDefinition" {
		t.Errorf("round-trip failed: resourceType=%v", parsed["resourceType"])
	}
	if parsed["code"] != "calculate-risk" {
		t.Errorf("round-trip failed: code=%v", parsed["code"])
	}
}

// ===========================================================================
// TestConcurrentRegistration
// ===========================================================================

func TestConcurrentRegistration(t *testing.T) {
	reg := NewCustomOperationRegistry()
	var wg sync.WaitGroup

	// Register 50 operations concurrently
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			def := &CustomOperationDef{
				Name:       fmt.Sprintf("$op-%d", idx),
				Code:       fmt.Sprintf("op-%d", idx),
				Title:      fmt.Sprintf("Operation %d", idx),
				Scope:      OperationScopeSystem,
				System:     true,
				Parameters: []OperationParamDef{},
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
			}
			_ = reg.Register(def, sampleHandler())
		}(i)
	}
	wg.Wait()

	ops := reg.List()
	if len(ops) != 50 {
		t.Errorf("expected 50 operations, got %d", len(ops))
	}
}

func TestConcurrentReadWrite(t *testing.T) {
	reg := NewCustomOperationRegistry()

	// Pre-register some operations
	for i := 0; i < 10; i++ {
		def := &CustomOperationDef{
			Name:       fmt.Sprintf("$existing-%d", i),
			Code:       fmt.Sprintf("existing-%d", i),
			Title:      fmt.Sprintf("Existing %d", i),
			Scope:      OperationScopeSystem,
			System:     true,
			Parameters: []OperationParamDef{},
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}
		_ = reg.Register(def, sampleHandler())
	}

	var wg sync.WaitGroup

	// Concurrent reads
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			_ = reg.List()
			_, _ = reg.Get(fmt.Sprintf("existing-%d", idx%10))
			_ = reg.ListByScope(OperationScopeSystem)
		}(i)
	}

	// Concurrent writes
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			def := &CustomOperationDef{
				Name:       fmt.Sprintf("$new-%d", idx),
				Code:       fmt.Sprintf("new-%d", idx),
				Title:      fmt.Sprintf("New %d", idx),
				Scope:      OperationScopeType,
				Type:       true,
				Parameters: []OperationParamDef{},
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
			}
			_ = reg.Register(def, sampleHandler())
		}(i)
	}

	wg.Wait()

	ops := reg.List()
	if len(ops) != 20 {
		t.Errorf("expected 20 operations, got %d", len(ops))
	}
}

// ===========================================================================
// TestAffectsStateEnforcement
// ===========================================================================

func TestAffectsState_PostRequired(t *testing.T) {
	reg := NewCustomOperationRegistry()
	def := sampleOperationDef()
	def.AffectsState = true
	def.Scope = OperationScopeSystem
	def.System = true
	_ = reg.Register(def, sampleHandler())

	// GET should fail
	_, _, err := reg.RouteOperation("GET", "/fhir/$calculate-risk")
	if err == nil {
		t.Fatal("expected error for GET on affects-state operation")
	}

	// POST should succeed
	_, _, err = reg.RouteOperation("POST", "/fhir/$calculate-risk")
	if err != nil {
		t.Fatalf("unexpected error for POST on affects-state operation: %v", err)
	}
}

func TestAffectsState_GetAllowedWhenFalse(t *testing.T) {
	reg := NewCustomOperationRegistry()
	def := sampleOperationDef()
	def.AffectsState = false
	def.Scope = OperationScopeSystem
	def.System = true
	_ = reg.Register(def, sampleHandler())

	_, _, err := reg.RouteOperation("GET", "/fhir/$calculate-risk")
	if err != nil {
		t.Fatalf("unexpected error for GET on non-affects-state: %v", err)
	}
}

// ===========================================================================
// Edge cases
// ===========================================================================

func TestRegister_EmptyParams(t *testing.T) {
	reg := NewCustomOperationRegistry()
	def := sampleOperationDef()
	def.Parameters = []OperationParamDef{}
	err := reg.Register(def, sampleHandler())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRouteOperation_EmptyPath(t *testing.T) {
	reg := NewCustomOperationRegistry()
	_, _, err := reg.RouteOperation("GET", "")
	if err == nil {
		t.Fatal("expected error for empty path")
	}
}

func TestRouteOperation_PathWithoutDollar(t *testing.T) {
	reg := NewCustomOperationRegistry()
	_, _, err := reg.RouteOperation("GET", "/fhir/Patient")
	if err == nil {
		t.Fatal("expected error for path without $")
	}
}

func TestValidateOperationInput_NoRequiredParams(t *testing.T) {
	def := &CustomOperationDef{
		Name: "$simple",
		Code: "simple",
		Parameters: []OperationParamDef{
			{Name: "optional", Use: "in", Min: 0, Max: "1", Type: "string", Required: false},
		},
	}
	issues := ValidateOperationInput(def, map[string]interface{}{})
	for _, issue := range issues {
		if issue.Severity == SeverityError {
			t.Errorf("unexpected error: %s", issue.Diagnostics)
		}
	}
}

func TestRouteOperation_ScopeMismatch_TypeOnSystemOnly(t *testing.T) {
	reg := NewCustomOperationRegistry()
	def := sampleOperationDef()
	def.Scope = OperationScopeSystem
	def.System = true
	def.Type = false
	def.Instance = false
	_ = reg.Register(def, sampleHandler())

	// System scope should work
	_, _, err := reg.RouteOperation("POST", "/fhir/$calculate-risk")
	if err != nil {
		t.Fatalf("unexpected error for system scope: %v", err)
	}

	// Type scope should fail
	_, _, err = reg.RouteOperation("POST", "/fhir/Patient/$calculate-risk")
	if err == nil {
		t.Fatal("expected error for type scope on system-only operation")
	}
}

func TestRouteOperation_ScopeMismatch_InstanceOnTypeOnly(t *testing.T) {
	reg := NewCustomOperationRegistry()
	def := sampleOperationDef()
	def.Scope = OperationScopeType
	def.System = false
	def.Type = true
	def.Instance = false
	_ = reg.Register(def, sampleHandler())

	// Instance scope should fail
	_, _, err := reg.RouteOperation("POST", "/fhir/Patient/123/$calculate-risk")
	if err == nil {
		t.Fatal("expected error for instance scope on type-only operation")
	}
}

func TestOperationContext_Fields(t *testing.T) {
	ctx := &OperationContext{
		Invocation: &OperationInvocation{
			OperationCode: "test-op",
			Scope:         OperationScopeInstance,
			ResourceType:  "Patient",
			ResourceID:    "456",
			Parameters:    map[string]interface{}{"key": "value"},
		},
		RequestID: "req-abc",
		TenantID:  "tenant-1",
		UserID:    "user-1",
	}

	if ctx.Invocation.OperationCode != "test-op" {
		t.Error("unexpected operation code")
	}
	if ctx.RequestID != "req-abc" {
		t.Error("unexpected request ID")
	}
	if ctx.TenantID != "tenant-1" {
		t.Error("unexpected tenant ID")
	}
	if ctx.UserID != "user-1" {
		t.Error("unexpected user ID")
	}
}

func TestOperationResponse_StatusCode(t *testing.T) {
	resp := &OperationResponse{
		StatusCode: http.StatusCreated,
		Resource: map[string]interface{}{
			"resourceType": "Patient",
		},
		ContentType: "application/fhir+json",
	}
	if resp.StatusCode != 201 {
		t.Errorf("expected 201, got %d", resp.StatusCode)
	}
}

func TestCustomOperationHandler_EmptyBody_POST(t *testing.T) {
	reg := NewCustomOperationRegistry()
	def := sampleOperationDef()
	def.Scope = OperationScopeSystem
	def.System = true
	def.Parameters = []OperationParamDef{} // No required params

	handler := func(ctx *OperationContext) (*OperationResponse, error) {
		return &OperationResponse{
			StatusCode: http.StatusOK,
			Resource:   map[string]interface{}{"resourceType": "Parameters"},
		}, nil
	}
	_ = reg.Register(def, handler)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/fhir/$calculate-risk", strings.NewReader(""))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/fhir/$calculate-risk")

	h := CustomOperationHandler(reg)
	err := h(c)
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}
	// Should still succeed since no required params
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestCustomOperationHandler_TypeLevel(t *testing.T) {
	reg := NewCustomOperationRegistry()
	def := sampleOperationDef()
	def.Scope = OperationScopeType
	def.Type = true
	def.AffectsState = false

	typeHandler := func(ctx *OperationContext) (*OperationResponse, error) {
		return &OperationResponse{
			StatusCode: http.StatusOK,
			Resource: map[string]interface{}{
				"resourceType": "Parameters",
				"parameter": []interface{}{
					map[string]interface{}{
						"name":        "resourceType",
						"valueString": ctx.Invocation.ResourceType,
					},
				},
			},
		}, nil
	}
	_ = reg.Register(def, typeHandler)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/fhir/Patient/$calculate-risk",
		strings.NewReader(`{"resourceType":"Parameters","parameter":[{"name":"subject","valueString":"Patient/123"}]}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/fhir/Patient/$calculate-risk")

	h := CustomOperationHandler(reg)
	err := h(c)
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestCustomOperationHandler_InstanceLevel(t *testing.T) {
	reg := NewCustomOperationRegistry()
	def := sampleOperationDef()
	def.Scope = OperationScopeInstance
	def.Instance = true
	def.AffectsState = false

	instHandler := func(ctx *OperationContext) (*OperationResponse, error) {
		return &OperationResponse{
			StatusCode: http.StatusOK,
			Resource: map[string]interface{}{
				"resourceType": "Parameters",
				"parameter": []interface{}{
					map[string]interface{}{
						"name":        "id",
						"valueString": ctx.Invocation.ResourceID,
					},
				},
			},
		}, nil
	}
	_ = reg.Register(def, instHandler)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/fhir/Patient/abc/$calculate-risk",
		strings.NewReader(`{"resourceType":"Parameters","parameter":[{"name":"subject","valueString":"Patient/abc"}]}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/fhir/Patient/abc/$calculate-risk")

	h := CustomOperationHandler(reg)
	err := h(c)
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestCustomOperationHandler_CustomStatusCode(t *testing.T) {
	reg := NewCustomOperationRegistry()
	def := sampleOperationDef()
	def.Scope = OperationScopeSystem
	def.System = true

	customHandler := func(ctx *OperationContext) (*OperationResponse, error) {
		return &OperationResponse{
			StatusCode: http.StatusAccepted,
			Resource: map[string]interface{}{
				"resourceType": "OperationOutcome",
				"issue": []interface{}{
					map[string]interface{}{
						"severity":    "information",
						"code":        "informational",
						"diagnostics": "Operation accepted",
					},
				},
			},
		}, nil
	}
	_ = reg.Register(def, customHandler)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/fhir/$calculate-risk",
		strings.NewReader(`{"resourceType":"Parameters","parameter":[{"name":"subject","valueString":"Patient/1"}]}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/fhir/$calculate-risk")

	h := CustomOperationHandler(reg)
	_ = h(c)
	if rec.Code != http.StatusAccepted {
		t.Errorf("expected 202, got %d", rec.Code)
	}
}

func TestParseOperationParameters_MultipleValueTypes(t *testing.T) {
	body := map[string]interface{}{
		"resourceType": "Parameters",
		"parameter": []interface{}{
			map[string]interface{}{
				"name":           "uri-val",
				"valueUri":       "http://example.com",
			},
			map[string]interface{}{
				"name":           "code-val",
				"valueCode":      "active",
			},
			map[string]interface{}{
				"name":           "date-val",
				"valueDate":      "2024-01-01",
			},
		},
	}

	params, err := ParseOperationParameters(body)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if params["uri-val"] != "http://example.com" {
		t.Errorf("expected uri value, got %v", params["uri-val"])
	}
	if params["code-val"] != "active" {
		t.Errorf("expected code value, got %v", params["code-val"])
	}
	if params["date-val"] != "2024-01-01" {
		t.Errorf("expected date value, got %v", params["date-val"])
	}
}

func TestVeryLongOperationName(t *testing.T) {
	longName := "$" + strings.Repeat("a", 500)
	longCode := strings.Repeat("a", 500)

	def := sampleOperationDef()
	def.Name = longName
	def.Code = longCode
	issues := ValidateOperationDef(def)
	found := false
	for _, issue := range issues {
		if strings.Contains(issue.Diagnostics, "long") || strings.Contains(issue.Diagnostics, "length") || strings.Contains(issue.Diagnostics, "exceed") {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected validation warning or error for very long name/code")
	}
}

func TestOperationScope_Flags(t *testing.T) {
	if OperationScopeSystem&OperationScopeType != 0 {
		t.Error("system and type scopes should be distinct bits")
	}
	if OperationScopeType&OperationScopeInstance != 0 {
		t.Error("type and instance scopes should be distinct bits")
	}

	combined := OperationScopeSystem | OperationScopeType
	if combined&OperationScopeSystem == 0 {
		t.Error("combined should include system")
	}
	if combined&OperationScopeType == 0 {
		t.Error("combined should include type")
	}
	if combined&OperationScopeInstance != 0 {
		t.Error("combined should not include instance")
	}
}

func TestIdempotentCacheHint(t *testing.T) {
	def := sampleOperationDef()
	def.Idempotent = true
	opDefResource := def.ToOperationDefinition()

	// When idempotent, should have a hint in extension or as a field
	// For now check the field itself is present
	if opDefResource["idempotent"] != true {
		t.Error("expected idempotent=true in OperationDefinition resource")
	}
}

func TestIdempotentCacheHint_NotIdempotent(t *testing.T) {
	def := sampleOperationDef()
	def.Idempotent = false
	opDefResource := def.ToOperationDefinition()

	if opDefResource["idempotent"] != false {
		t.Error("expected idempotent=false in OperationDefinition resource")
	}
}

func TestListForResourceType_EmptyResourceType(t *testing.T) {
	reg := NewCustomOperationRegistry()
	def := sampleOperationDef()
	def.ResourceTypes = nil // all types
	_ = reg.Register(def, sampleHandler())

	// Empty string should still match all-types operations
	ops := reg.ListForResourceType("")
	if len(ops) != 1 {
		t.Errorf("expected 1 operation for empty resource type, got %d", len(ops))
	}
}

func TestRouteOperation_ResourceTypeMismatch(t *testing.T) {
	reg := NewCustomOperationRegistry()
	def := sampleOperationDef()
	def.ResourceTypes = []string{"Patient"}
	def.Scope = OperationScopeType
	def.Type = true
	_ = reg.Register(def, sampleHandler())

	// Should fail for Observation
	_, _, err := reg.RouteOperation("POST", "/fhir/Observation/$calculate-risk")
	if err == nil {
		t.Fatal("expected error for resource type mismatch")
	}
}
