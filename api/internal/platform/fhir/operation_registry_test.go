package fhir

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/labstack/echo/v4"
)

// ===========================================================================
// OperationRegistry core tests
// ===========================================================================

func TestNewOperationRegistry(t *testing.T) {
	reg := NewOperationRegistry()
	if reg == nil {
		t.Fatal("expected non-nil registry")
	}
	if len(reg.List()) != 0 {
		t.Errorf("expected empty registry, got %d operations", len(reg.List()))
	}
}

func TestOperationRegistry_Register(t *testing.T) {
	reg := NewOperationRegistry()
	reg.Register(&OperationDefinitionResource{
		ResourceType: "OperationDefinition",
		Code:         "validate",
		Name:         "Validate",
		URL:          "http://hl7.org/fhir/OperationDefinition/Resource-validate",
		Status:       "active",
		Kind:         "operation",
		System:       true,
		Type:         true,
		Instance:     true,
	})

	if len(reg.List()) != 1 {
		t.Fatalf("expected 1 operation, got %d", len(reg.List()))
	}
}

func TestOperationRegistry_Get(t *testing.T) {
	reg := NewOperationRegistry()
	reg.Register(&OperationDefinitionResource{
		ResourceType: "OperationDefinition",
		Code:         "validate",
		Name:         "Validate",
		URL:          "http://hl7.org/fhir/OperationDefinition/Resource-validate",
		Status:       "active",
		Kind:         "operation",
	})

	op := reg.Get("validate")
	if op == nil {
		t.Fatal("expected to find validate operation")
	}
	if op.Name != "Validate" {
		t.Errorf("expected name Validate, got %s", op.Name)
	}
}

func TestOperationRegistry_Get_NotFound(t *testing.T) {
	reg := NewOperationRegistry()
	op := reg.Get("nonexistent")
	if op != nil {
		t.Errorf("expected nil for nonexistent operation, got %+v", op)
	}
}

func TestOperationRegistry_List_Sorted(t *testing.T) {
	reg := NewOperationRegistry()
	reg.Register(&OperationDefinitionResource{Code: "validate", Name: "Validate", Status: "active", Kind: "operation"})
	reg.Register(&OperationDefinitionResource{Code: "everything", Name: "Everything", Status: "active", Kind: "operation"})
	reg.Register(&OperationDefinitionResource{Code: "apply", Name: "Apply", Status: "active", Kind: "operation"})

	ops := reg.List()
	if len(ops) != 3 {
		t.Fatalf("expected 3 operations, got %d", len(ops))
	}

	// Should be sorted alphabetically by code: apply, everything, validate
	if ops[0].Code != "apply" {
		t.Errorf("expected first op apply, got %s", ops[0].Code)
	}
	if ops[1].Code != "everything" {
		t.Errorf("expected second op everything, got %s", ops[1].Code)
	}
	if ops[2].Code != "validate" {
		t.Errorf("expected third op validate, got %s", ops[2].Code)
	}
}

func TestOperationRegistry_Register_Overwrite(t *testing.T) {
	reg := NewOperationRegistry()
	reg.Register(&OperationDefinitionResource{Code: "validate", Name: "Validate", Status: "active", Kind: "operation"})
	reg.Register(&OperationDefinitionResource{Code: "validate", Name: "ValidateV2", Status: "active", Kind: "operation"})

	ops := reg.List()
	if len(ops) != 1 {
		t.Fatalf("expected 1 operation after overwrite, got %d", len(ops))
	}
	if ops[0].Name != "ValidateV2" {
		t.Errorf("expected overwritten name ValidateV2, got %s", ops[0].Name)
	}
}

func TestOperationRegistry_ConcurrentAccess(t *testing.T) {
	reg := NewOperationRegistry()

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			code := fmt.Sprintf("op-%d", idx)
			reg.Register(&OperationDefinitionResource{
				Code:   code,
				Name:   code,
				Status: "active",
				Kind:   "operation",
			})
			_ = reg.Get(code)
			_ = reg.List()
		}(i)
	}
	wg.Wait()

	if len(reg.List()) != 50 {
		t.Errorf("expected 50 operations after concurrent registration, got %d", len(reg.List()))
	}
}

// ===========================================================================
// DefaultOperationRegistry tests
// ===========================================================================

func TestDefaultOperationRegistry(t *testing.T) {
	reg := DefaultOperationRegistry()
	ops := reg.List()

	if len(ops) < 21 {
		t.Errorf("expected at least 21 default operations, got %d", len(ops))
	}
}

func TestDefaultOperationRegistry_AllExpectedCodes(t *testing.T) {
	reg := DefaultOperationRegistry()

	expectedCodes := []string{
		"validate", "everything", "export", "expand", "lookup",
		"validate-code", "translate", "subsumes", "match", "meta",
		"meta-add", "meta-delete", "diff", "lastn", "stats",
		"convert", "graph", "batch-validate", "document", "closure",
		"apply",
	}

	for _, code := range expectedCodes {
		op := reg.Get(code)
		if op == nil {
			t.Errorf("missing expected operation: %s", code)
			continue
		}
		if op.ResourceType != "OperationDefinition" {
			t.Errorf("%s: expected resourceType OperationDefinition, got %s", code, op.ResourceType)
		}
		if op.Status != "active" {
			t.Errorf("%s: expected status active, got %s", code, op.Status)
		}
		if op.Kind != "operation" {
			t.Errorf("%s: expected kind operation, got %s", code, op.Kind)
		}
	}
}

func TestDefaultOperationRegistry_Validate(t *testing.T) {
	reg := DefaultOperationRegistry()
	op := reg.Get("validate")

	if op == nil {
		t.Fatal("expected validate operation")
	}
	if !op.System {
		t.Error("validate should be a system operation")
	}
	if !op.Type {
		t.Error("validate should be a type operation")
	}
	if !op.Instance {
		t.Error("validate should be an instance operation")
	}
	if len(op.Parameter) == 0 {
		t.Error("validate should have parameters")
	}
}

func TestDefaultOperationRegistry_Everything(t *testing.T) {
	reg := DefaultOperationRegistry()
	op := reg.Get("everything")

	if op == nil {
		t.Fatal("expected everything operation")
	}
	if op.System {
		t.Error("everything should not be a system operation")
	}
	if op.Type {
		t.Error("everything should not be a type operation")
	}
	if !op.Instance {
		t.Error("everything should be an instance operation")
	}
	if len(op.Resource) != 1 || op.Resource[0] != "Patient" {
		t.Errorf("expected resource [Patient], got %v", op.Resource)
	}
}

func TestDefaultOperationRegistry_Export(t *testing.T) {
	reg := DefaultOperationRegistry()
	op := reg.Get("export")

	if op == nil {
		t.Fatal("expected export operation")
	}
	if !op.System {
		t.Error("export should be a system operation")
	}
	if !op.Type {
		t.Error("export should be a type operation")
	}
	if op.Instance {
		t.Error("export should not be an instance operation")
	}
	if len(op.Resource) != 2 {
		t.Errorf("expected 2 resources for export, got %d", len(op.Resource))
	}
}

func TestDefaultOperationRegistry_Expand(t *testing.T) {
	reg := DefaultOperationRegistry()
	op := reg.Get("expand")

	if op == nil {
		t.Fatal("expected expand operation")
	}
	if !op.Type {
		t.Error("expand should be a type operation")
	}
	if !op.Instance {
		t.Error("expand should be an instance operation")
	}
	if len(op.Resource) != 1 || op.Resource[0] != "ValueSet" {
		t.Errorf("expected resource [ValueSet], got %v", op.Resource)
	}
}

func TestDefaultOperationRegistry_Lookup(t *testing.T) {
	reg := DefaultOperationRegistry()
	op := reg.Get("lookup")

	if op == nil {
		t.Fatal("expected lookup operation")
	}
	if !op.Type {
		t.Error("lookup should be a type operation")
	}
	if op.Instance {
		t.Error("lookup should not be an instance operation")
	}
	if len(op.Resource) != 1 || op.Resource[0] != "CodeSystem" {
		t.Errorf("expected resource [CodeSystem], got %v", op.Resource)
	}
}

func TestDefaultOperationRegistry_ValidateCode(t *testing.T) {
	reg := DefaultOperationRegistry()
	op := reg.Get("validate-code")

	if op == nil {
		t.Fatal("expected validate-code operation")
	}
	if !op.Type {
		t.Error("validate-code should be a type operation")
	}
	if !op.Instance {
		t.Error("validate-code should be an instance operation")
	}
	if len(op.Resource) != 1 || op.Resource[0] != "ValueSet" {
		t.Errorf("expected resource [ValueSet], got %v", op.Resource)
	}
}

func TestDefaultOperationRegistry_Translate(t *testing.T) {
	reg := DefaultOperationRegistry()
	op := reg.Get("translate")

	if op == nil {
		t.Fatal("expected translate operation")
	}
	if !op.Type {
		t.Error("translate should be a type operation")
	}
	if !op.Instance {
		t.Error("translate should be an instance operation")
	}
	if len(op.Resource) != 1 || op.Resource[0] != "ConceptMap" {
		t.Errorf("expected resource [ConceptMap], got %v", op.Resource)
	}
}

func TestDefaultOperationRegistry_Subsumes(t *testing.T) {
	reg := DefaultOperationRegistry()
	op := reg.Get("subsumes")

	if op == nil {
		t.Fatal("expected subsumes operation")
	}
	if !op.Type {
		t.Error("subsumes should be a type operation")
	}
	if op.Instance {
		t.Error("subsumes should not be an instance operation")
	}
	if len(op.Resource) != 1 || op.Resource[0] != "CodeSystem" {
		t.Errorf("expected resource [CodeSystem], got %v", op.Resource)
	}
}

func TestDefaultOperationRegistry_Match(t *testing.T) {
	reg := DefaultOperationRegistry()
	op := reg.Get("match")

	if op == nil {
		t.Fatal("expected match operation")
	}
	if !op.Type {
		t.Error("match should be a type operation")
	}
	if op.Instance {
		t.Error("match should not be an instance operation")
	}
	if len(op.Resource) != 1 || op.Resource[0] != "Patient" {
		t.Errorf("expected resource [Patient], got %v", op.Resource)
	}
}

func TestDefaultOperationRegistry_Meta(t *testing.T) {
	reg := DefaultOperationRegistry()
	op := reg.Get("meta")

	if op == nil {
		t.Fatal("expected meta operation")
	}
	if op.System {
		t.Error("meta should not be a system operation")
	}
	if op.Type {
		t.Error("meta should not be a type operation")
	}
	if !op.Instance {
		t.Error("meta should be an instance operation")
	}
}

func TestDefaultOperationRegistry_MetaAdd(t *testing.T) {
	reg := DefaultOperationRegistry()
	op := reg.Get("meta-add")

	if op == nil {
		t.Fatal("expected meta-add operation")
	}
	if !op.Instance {
		t.Error("meta-add should be an instance operation")
	}
}

func TestDefaultOperationRegistry_MetaDelete(t *testing.T) {
	reg := DefaultOperationRegistry()
	op := reg.Get("meta-delete")

	if op == nil {
		t.Fatal("expected meta-delete operation")
	}
	if !op.Instance {
		t.Error("meta-delete should be an instance operation")
	}
}

func TestDefaultOperationRegistry_Diff(t *testing.T) {
	reg := DefaultOperationRegistry()
	op := reg.Get("diff")

	if op == nil {
		t.Fatal("expected diff operation")
	}
	if !op.Instance {
		t.Error("diff should be an instance operation")
	}
	if op.System {
		t.Error("diff should not be a system operation")
	}
}

func TestDefaultOperationRegistry_Lastn(t *testing.T) {
	reg := DefaultOperationRegistry()
	op := reg.Get("lastn")

	if op == nil {
		t.Fatal("expected lastn operation")
	}
	if !op.Type {
		t.Error("lastn should be a type operation")
	}
	if len(op.Resource) != 1 || op.Resource[0] != "Observation" {
		t.Errorf("expected resource [Observation], got %v", op.Resource)
	}
}

func TestDefaultOperationRegistry_Stats(t *testing.T) {
	reg := DefaultOperationRegistry()
	op := reg.Get("stats")

	if op == nil {
		t.Fatal("expected stats operation")
	}
	if !op.Type {
		t.Error("stats should be a type operation")
	}
	if len(op.Resource) != 1 || op.Resource[0] != "Observation" {
		t.Errorf("expected resource [Observation], got %v", op.Resource)
	}
}

func TestDefaultOperationRegistry_Convert(t *testing.T) {
	reg := DefaultOperationRegistry()
	op := reg.Get("convert")

	if op == nil {
		t.Fatal("expected convert operation")
	}
	if !op.System {
		t.Error("convert should be a system operation")
	}
	if op.Type {
		t.Error("convert should not be a type operation")
	}
	if op.Instance {
		t.Error("convert should not be an instance operation")
	}
}

func TestDefaultOperationRegistry_Graph(t *testing.T) {
	reg := DefaultOperationRegistry()
	op := reg.Get("graph")

	if op == nil {
		t.Fatal("expected graph operation")
	}
	if !op.System {
		t.Error("graph should be a system operation")
	}
}

func TestDefaultOperationRegistry_BatchValidate(t *testing.T) {
	reg := DefaultOperationRegistry()
	op := reg.Get("batch-validate")

	if op == nil {
		t.Fatal("expected batch-validate operation")
	}
	if !op.System {
		t.Error("batch-validate should be a system operation")
	}
}

func TestDefaultOperationRegistry_Document(t *testing.T) {
	reg := DefaultOperationRegistry()
	op := reg.Get("document")

	if op == nil {
		t.Fatal("expected document operation")
	}
	if !op.Instance {
		t.Error("document should be an instance operation")
	}
	if len(op.Resource) != 1 || op.Resource[0] != "Composition" {
		t.Errorf("expected resource [Composition], got %v", op.Resource)
	}
}

func TestDefaultOperationRegistry_Closure(t *testing.T) {
	reg := DefaultOperationRegistry()
	op := reg.Get("closure")

	if op == nil {
		t.Fatal("expected closure operation")
	}
	if !op.System {
		t.Error("closure should be a system operation")
	}
}

func TestDefaultOperationRegistry_Apply(t *testing.T) {
	reg := DefaultOperationRegistry()
	op := reg.Get("apply")

	if op == nil {
		t.Fatal("expected apply operation")
	}
	if !op.Instance {
		t.Error("apply should be an instance operation")
	}
	if len(op.Resource) != 1 || op.Resource[0] != "PlanDefinition" {
		t.Errorf("expected resource [PlanDefinition], got %v", op.Resource)
	}
}

func TestDefaultOperationRegistry_AllHaveParameters(t *testing.T) {
	reg := DefaultOperationRegistry()
	for _, op := range reg.List() {
		if len(op.Parameter) == 0 {
			t.Errorf("operation %s has no parameters defined", op.Code)
		}
	}
}

func TestDefaultOperationRegistry_AllHaveURLs(t *testing.T) {
	reg := DefaultOperationRegistry()
	for _, op := range reg.List() {
		if op.URL == "" {
			t.Errorf("operation %s has no URL", op.Code)
		}
	}
}

func TestDefaultOperationRegistry_AllHaveIDs(t *testing.T) {
	reg := DefaultOperationRegistry()
	for _, op := range reg.List() {
		if op.ID == "" {
			t.Errorf("operation %s has no ID", op.Code)
		}
	}
}

func TestDefaultOperationRegistry_AllHaveDescriptions(t *testing.T) {
	reg := DefaultOperationRegistry()
	for _, op := range reg.List() {
		if op.Description == "" {
			t.Errorf("operation %s has no description", op.Code)
		}
	}
}

// ===========================================================================
// OperationDefinitionResource serialization tests
// ===========================================================================

func TestOperationDefinitionResource_JSONSerialization(t *testing.T) {
	op := &OperationDefinitionResource{
		ResourceType: "OperationDefinition",
		ID:           "validate",
		URL:          "http://hl7.org/fhir/OperationDefinition/Resource-validate",
		Name:         "Validate",
		Title:        "Validate a resource",
		Status:       "active",
		Kind:         "operation",
		Code:         "validate",
		System:       true,
		Type:         true,
		Instance:     true,
		Description:  "Validate a resource against its structure definition",
		Parameter: []OperationParam{
			{Name: "resource", Use: "in", Min: 1, Max: "1", Type: "Resource"},
			{Name: "return", Use: "out", Min: 1, Max: "1", Type: "OperationOutcome"},
		},
	}

	data, err := json.Marshal(op)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded OperationDefinitionResource
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.ResourceType != "OperationDefinition" {
		t.Errorf("expected resourceType OperationDefinition, got %s", decoded.ResourceType)
	}
	if decoded.Code != "validate" {
		t.Errorf("expected code validate, got %s", decoded.Code)
	}
	if !decoded.System {
		t.Error("expected system true")
	}
	if !decoded.Type {
		t.Error("expected type true")
	}
	if !decoded.Instance {
		t.Error("expected instance true")
	}
	if len(decoded.Parameter) != 2 {
		t.Errorf("expected 2 parameters, got %d", len(decoded.Parameter))
	}
}

func TestOperationParam_JSONSerialization(t *testing.T) {
	param := OperationParam{
		Name:          "resource",
		Use:           "in",
		Min:           1,
		Max:           "1",
		Type:          "Resource",
		Documentation: "The resource to validate",
	}

	data, err := json.Marshal(param)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded OperationParam
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.Name != "resource" {
		t.Errorf("expected name resource, got %s", decoded.Name)
	}
	if decoded.Use != "in" {
		t.Errorf("expected use in, got %s", decoded.Use)
	}
	if decoded.Min != 1 {
		t.Errorf("expected min 1, got %d", decoded.Min)
	}
	if decoded.Max != "1" {
		t.Errorf("expected max 1, got %s", decoded.Max)
	}
	if decoded.Type != "Resource" {
		t.Errorf("expected type Resource, got %s", decoded.Type)
	}
	if decoded.Documentation != "The resource to validate" {
		t.Errorf("unexpected documentation: %s", decoded.Documentation)
	}
}

func TestOperationDefinitionResource_JSONOmitsEmpty(t *testing.T) {
	op := &OperationDefinitionResource{
		ResourceType: "OperationDefinition",
		URL:          "http://example.com/op",
		Name:         "Test",
		Status:       "active",
		Kind:         "operation",
		Code:         "test",
		System:       true,
		Type:         false,
		Instance:     false,
	}

	data, err := json.Marshal(op)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("failed to unmarshal to map: %v", err)
	}

	// id, title, resource, parameter, description should be omitted
	if _, ok := raw["id"]; ok {
		t.Error("expected id to be omitted when empty")
	}
	if _, ok := raw["title"]; ok {
		t.Error("expected title to be omitted when empty")
	}
	if _, ok := raw["resource"]; ok {
		t.Error("expected resource to be omitted when empty")
	}
	if _, ok := raw["parameter"]; ok {
		t.Error("expected parameter to be omitted when empty")
	}
	if _, ok := raw["description"]; ok {
		t.Error("expected description to be omitted when empty")
	}
}

// ===========================================================================
// OperationRegistryHandler tests
// ===========================================================================

func newTestOperationServer() (*echo.Echo, *OperationRegistry) {
	reg := DefaultOperationRegistry()
	h := NewOperationRegistryHandler(reg)

	e := echo.New()
	h.RegisterRoutes(e.Group("/fhir"))
	return e, reg
}

func TestOperationRegistryHandler_Search(t *testing.T) {
	e, _ := newTestOperationServer()

	req := httptest.NewRequest(http.MethodGet, "/fhir/OperationDefinition", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var bundle map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &bundle); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if bundle["resourceType"] != "Bundle" {
		t.Errorf("expected Bundle, got %v", bundle["resourceType"])
	}
	if bundle["type"] != "searchset" {
		t.Errorf("expected searchset, got %v", bundle["type"])
	}

	total, ok := bundle["total"].(float64)
	if !ok {
		t.Fatal("expected total to be a number")
	}
	if int(total) < 21 {
		t.Errorf("expected at least 21 entries, got %d", int(total))
	}
}

func TestOperationRegistryHandler_Search_FilterByCode(t *testing.T) {
	e, _ := newTestOperationServer()

	req := httptest.NewRequest(http.MethodGet, "/fhir/OperationDefinition?code=validate", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var bundle map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &bundle); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	total := int(bundle["total"].(float64))
	if total != 1 {
		t.Errorf("expected 1 result for code=validate, got %d", total)
	}

	entries := bundle["entry"].([]interface{})
	entry := entries[0].(map[string]interface{})
	resource := entry["resource"].(map[string]interface{})
	if resource["code"] != "validate" {
		t.Errorf("expected code validate, got %v", resource["code"])
	}
}

func TestOperationRegistryHandler_Search_FilterByName(t *testing.T) {
	e, _ := newTestOperationServer()

	req := httptest.NewRequest(http.MethodGet, "/fhir/OperationDefinition?name=Validate", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var bundle map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &bundle); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	total := int(bundle["total"].(float64))
	if total != 1 {
		t.Errorf("expected 1 result for name=Validate, got %d", total)
	}
}

func TestOperationRegistryHandler_Search_FilterBySystem(t *testing.T) {
	e, _ := newTestOperationServer()

	req := httptest.NewRequest(http.MethodGet, "/fhir/OperationDefinition?system=true", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var bundle map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &bundle); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	total := int(bundle["total"].(float64))
	if total < 5 {
		t.Errorf("expected at least 5 system operations, got %d", total)
	}

	// Verify all returned entries have system=true
	entries := bundle["entry"].([]interface{})
	for _, e := range entries {
		entry := e.(map[string]interface{})
		resource := entry["resource"].(map[string]interface{})
		if resource["system"] != true {
			t.Errorf("expected system=true for %v", resource["code"])
		}
	}
}

func TestOperationRegistryHandler_Search_FilterByType(t *testing.T) {
	e, _ := newTestOperationServer()

	req := httptest.NewRequest(http.MethodGet, "/fhir/OperationDefinition?type=true", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var bundle map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &bundle); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	total := int(bundle["total"].(float64))
	if total < 5 {
		t.Errorf("expected at least 5 type operations, got %d", total)
	}
}

func TestOperationRegistryHandler_Search_FilterByInstance(t *testing.T) {
	e, _ := newTestOperationServer()

	req := httptest.NewRequest(http.MethodGet, "/fhir/OperationDefinition?instance=true", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var bundle map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &bundle); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	total := int(bundle["total"].(float64))
	if total < 5 {
		t.Errorf("expected at least 5 instance operations, got %d", total)
	}
}

func TestOperationRegistryHandler_Search_FilterByResource(t *testing.T) {
	e, _ := newTestOperationServer()

	req := httptest.NewRequest(http.MethodGet, "/fhir/OperationDefinition?resource=Patient", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var bundle map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &bundle); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	total := int(bundle["total"].(float64))
	// At minimum: everything, export, match
	if total < 3 {
		t.Errorf("expected at least 3 Patient operations, got %d", total)
	}
}

func TestOperationRegistryHandler_Search_FilterBySystemFalse(t *testing.T) {
	e, _ := newTestOperationServer()

	req := httptest.NewRequest(http.MethodGet, "/fhir/OperationDefinition?system=false", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var bundle map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &bundle); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	// Verify no entries have system=true
	entries := bundle["entry"].([]interface{})
	for _, e := range entries {
		entry := e.(map[string]interface{})
		resource := entry["resource"].(map[string]interface{})
		if resource["system"] == true {
			t.Errorf("expected system=false for %v, but got system=true", resource["code"])
		}
	}
}

func TestOperationRegistryHandler_Search_NoResults(t *testing.T) {
	e, _ := newTestOperationServer()

	req := httptest.NewRequest(http.MethodGet, "/fhir/OperationDefinition?code=nonexistent", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var bundle map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &bundle); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	total := int(bundle["total"].(float64))
	if total != 0 {
		t.Errorf("expected 0 results for nonexistent code, got %d", total)
	}
}

func TestOperationRegistryHandler_Search_MultipleFilters(t *testing.T) {
	e, _ := newTestOperationServer()

	req := httptest.NewRequest(http.MethodGet, "/fhir/OperationDefinition?system=true&instance=false", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var bundle map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &bundle); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	// Verify all returned entries have system=true and instance=false
	entries := bundle["entry"].([]interface{})
	for _, e := range entries {
		entry := e.(map[string]interface{})
		resource := entry["resource"].(map[string]interface{})
		if resource["system"] != true {
			t.Errorf("expected system=true for %v", resource["code"])
		}
		if resource["instance"] == true {
			t.Errorf("expected instance=false for %v", resource["code"])
		}
	}
}

func TestOperationRegistryHandler_Read(t *testing.T) {
	e, _ := newTestOperationServer()

	req := httptest.NewRequest(http.MethodGet, "/fhir/OperationDefinition/validate", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var op map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &op); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if op["resourceType"] != "OperationDefinition" {
		t.Errorf("expected OperationDefinition, got %v", op["resourceType"])
	}
	if op["code"] != "validate" {
		t.Errorf("expected code validate, got %v", op["code"])
	}
}

func TestOperationRegistryHandler_Read_NotFound(t *testing.T) {
	e, _ := newTestOperationServer()

	req := httptest.NewRequest(http.MethodGet, "/fhir/OperationDefinition/nonexistent", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}

	var outcome map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &outcome); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if outcome["resourceType"] != "OperationOutcome" {
		t.Errorf("expected OperationOutcome, got %v", outcome["resourceType"])
	}
}

func TestOperationRegistryHandler_Read_AllDefaults(t *testing.T) {
	e, _ := newTestOperationServer()

	codes := []string{
		"validate", "everything", "export", "expand", "lookup",
		"validate-code", "translate", "subsumes", "match", "meta",
		"meta-add", "meta-delete", "diff", "lastn", "stats",
		"convert", "graph", "batch-validate", "document", "closure",
		"apply",
	}

	for _, code := range codes {
		req := httptest.NewRequest(http.MethodGet, "/fhir/OperationDefinition/"+code, nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected 200 for %s, got %d", code, rec.Code)
		}
	}
}

func TestOperationRegistryHandler_ConcurrentAccess(t *testing.T) {
	e, _ := newTestOperationServer()

	var wg sync.WaitGroup
	errs := make(chan error, 100)

	// Concurrent search requests
	for i := 0; i < 30; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			req := httptest.NewRequest(http.MethodGet, "/fhir/OperationDefinition", nil)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)
			if rec.Code != http.StatusOK {
				errs <- fmt.Errorf("search returned %d", rec.Code)
			}
		}()
	}

	// Concurrent read requests
	for i := 0; i < 30; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			req := httptest.NewRequest(http.MethodGet, "/fhir/OperationDefinition/validate", nil)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)
			if rec.Code != http.StatusOK {
				errs <- fmt.Errorf("read returned %d", rec.Code)
			}
		}()
	}

	wg.Wait()
	close(errs)

	for err := range errs {
		t.Errorf("concurrent error: %v", err)
	}
}

// ===========================================================================
// NewOperationRegistryHandler tests
// ===========================================================================

func TestNewOperationRegistryHandler(t *testing.T) {
	reg := NewOperationRegistry()
	h := NewOperationRegistryHandler(reg)
	if h == nil {
		t.Fatal("expected non-nil handler")
	}
}

func TestNewOperationRegistryHandler_RegisterRoutes(t *testing.T) {
	reg := NewOperationRegistry()
	h := NewOperationRegistryHandler(reg)

	e := echo.New()
	h.RegisterRoutes(e.Group("/fhir"))

	// Verify routes exist by making requests
	req := httptest.NewRequest(http.MethodGet, "/fhir/OperationDefinition", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 for search endpoint, got %d", rec.Code)
	}
}

// ===========================================================================
// Search entry format tests
// ===========================================================================

func TestOperationRegistryHandler_Search_EntryFormat(t *testing.T) {
	e, _ := newTestOperationServer()

	req := httptest.NewRequest(http.MethodGet, "/fhir/OperationDefinition?code=validate", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	var bundle map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &bundle); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	entries := bundle["entry"].([]interface{})
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}

	entry := entries[0].(map[string]interface{})

	// Verify search mode
	search, ok := entry["search"].(map[string]interface{})
	if !ok {
		t.Fatal("expected search element in entry")
	}
	if search["mode"] != "match" {
		t.Errorf("expected search mode match, got %v", search["mode"])
	}

	// Verify resource is present
	resource, ok := entry["resource"].(map[string]interface{})
	if !ok {
		t.Fatal("expected resource element in entry")
	}
	if resource["resourceType"] != "OperationDefinition" {
		t.Errorf("expected OperationDefinition, got %v", resource["resourceType"])
	}
}
