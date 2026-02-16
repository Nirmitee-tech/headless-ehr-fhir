package fhir

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
)

// =========== InMemoryMetaStore Tests ===========

func TestInMemoryMetaStore_GetMeta_Empty(t *testing.T) {
	store := NewInMemoryMetaStore()
	meta, err := store.GetMeta(context.Background(), "Patient", "nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(meta.Profile) != 0 || len(meta.Security) != 0 || len(meta.Tag) != 0 {
		t.Errorf("expected empty meta for nonexistent resource, got %+v", meta)
	}
}

func TestInMemoryMetaStore_AddMeta(t *testing.T) {
	store := NewInMemoryMetaStore()
	ctx := context.Background()

	input := &Meta{
		Profile: []string{"http://hl7.org/fhir/StructureDefinition/Patient"},
		Security: []Coding{
			{System: "http://terminology.hl7.org/CodeSystem/v3-Confidentiality", Code: "R", Display: "Restricted"},
		},
		Tag: []Coding{
			{System: "http://example.org/tags", Code: "needs-review", Display: "Needs Review"},
		},
	}

	result, err := store.AddMeta(ctx, "Patient", "123", input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Profile) != 1 || result.Profile[0] != "http://hl7.org/fhir/StructureDefinition/Patient" {
		t.Errorf("expected 1 profile, got %v", result.Profile)
	}
	if len(result.Security) != 1 || result.Security[0].Code != "R" {
		t.Errorf("expected 1 security label with code R, got %v", result.Security)
	}
	if len(result.Tag) != 1 || result.Tag[0].Code != "needs-review" {
		t.Errorf("expected 1 tag with code needs-review, got %v", result.Tag)
	}
}

func TestInMemoryMetaStore_AddMeta_MergesDuplicates(t *testing.T) {
	store := NewInMemoryMetaStore()
	ctx := context.Background()

	first := &Meta{
		Tag: []Coding{
			{System: "http://example.org/tags", Code: "tag-a"},
			{System: "http://example.org/tags", Code: "tag-b"},
		},
	}
	_, err := store.AddMeta(ctx, "Patient", "123", first)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	second := &Meta{
		Tag: []Coding{
			{System: "http://example.org/tags", Code: "tag-b"}, // duplicate
			{System: "http://example.org/tags", Code: "tag-c"}, // new
		},
	}
	result, err := store.AddMeta(ctx, "Patient", "123", second)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Tag) != 3 {
		t.Errorf("expected 3 tags after merge (a, b, c), got %d: %v", len(result.Tag), result.Tag)
	}

	codes := map[string]bool{}
	for _, tag := range result.Tag {
		codes[tag.Code] = true
	}
	for _, expected := range []string{"tag-a", "tag-b", "tag-c"} {
		if !codes[expected] {
			t.Errorf("expected tag %s to be present", expected)
		}
	}
}

func TestInMemoryMetaStore_DeleteMeta(t *testing.T) {
	store := NewInMemoryMetaStore()
	ctx := context.Background()

	// Set up initial meta.
	initial := &Meta{
		Profile: []string{
			"http://hl7.org/fhir/StructureDefinition/Patient",
			"http://example.org/StructureDefinition/custom",
		},
		Tag: []Coding{
			{System: "http://example.org/tags", Code: "tag-a"},
			{System: "http://example.org/tags", Code: "tag-b"},
			{System: "http://example.org/tags", Code: "tag-c"},
		},
	}
	_, err := store.AddMeta(ctx, "Patient", "123", initial)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Delete specific items.
	toDelete := &Meta{
		Profile: []string{"http://example.org/StructureDefinition/custom"},
		Tag: []Coding{
			{System: "http://example.org/tags", Code: "tag-b"},
		},
	}
	result, err := store.DeleteMeta(ctx, "Patient", "123", toDelete)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Profile) != 1 || result.Profile[0] != "http://hl7.org/fhir/StructureDefinition/Patient" {
		t.Errorf("expected 1 profile remaining, got %v", result.Profile)
	}
	if len(result.Tag) != 2 {
		t.Errorf("expected 2 tags remaining, got %d: %v", len(result.Tag), result.Tag)
	}
	for _, tag := range result.Tag {
		if tag.Code == "tag-b" {
			t.Error("tag-b should have been removed")
		}
	}
}

func TestInMemoryMetaStore_DeleteMeta_Nonexistent(t *testing.T) {
	store := NewInMemoryMetaStore()
	ctx := context.Background()

	toDelete := &Meta{
		Tag: []Coding{
			{System: "http://example.org/tags", Code: "does-not-exist"},
		},
	}
	result, err := store.DeleteMeta(ctx, "Patient", "nonexistent", toDelete)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Tag) != 0 {
		t.Errorf("expected empty meta, got %+v", result)
	}
}

func TestInMemoryMetaStore_GetMeta_AfterAdd(t *testing.T) {
	store := NewInMemoryMetaStore()
	ctx := context.Background()

	input := &Meta{
		Tag: []Coding{
			{System: "http://example.org/tags", Code: "test-tag", Display: "Test Tag"},
		},
	}
	_, err := store.AddMeta(ctx, "Observation", "obs-1", input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	meta, err := store.GetMeta(ctx, "Observation", "obs-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(meta.Tag) != 1 || meta.Tag[0].Code != "test-tag" {
		t.Errorf("expected tag test-tag, got %v", meta.Tag)
	}
}

func TestInMemoryMetaStore_IsolatesResources(t *testing.T) {
	store := NewInMemoryMetaStore()
	ctx := context.Background()

	_, _ = store.AddMeta(ctx, "Patient", "1", &Meta{
		Tag: []Coding{{System: "http://example.org/tags", Code: "patient-tag"}},
	})
	_, _ = store.AddMeta(ctx, "Observation", "1", &Meta{
		Tag: []Coding{{System: "http://example.org/tags", Code: "obs-tag"}},
	})

	patientMeta, _ := store.GetMeta(ctx, "Patient", "1")
	obsMeta, _ := store.GetMeta(ctx, "Observation", "1")

	if len(patientMeta.Tag) != 1 || patientMeta.Tag[0].Code != "patient-tag" {
		t.Errorf("patient meta contaminated: %v", patientMeta.Tag)
	}
	if len(obsMeta.Tag) != 1 || obsMeta.Tag[0].Code != "obs-tag" {
		t.Errorf("observation meta contaminated: %v", obsMeta.Tag)
	}
}

// =========== MetaHandler Tests ===========

func TestMetaHandler_GetMeta_ReturnsParameters(t *testing.T) {
	store := NewInMemoryMetaStore()
	ctx := context.Background()
	_, _ = store.AddMeta(ctx, "Patient", "123", &Meta{
		Profile: []string{"http://hl7.org/fhir/StructureDefinition/Patient"},
		Tag: []Coding{
			{System: "http://example.org/tags", Code: "test", Display: "Test"},
		},
	})

	handler := NewMetaHandler(store)
	e := echo.New()

	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient/123/$meta", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("resourceType", "id")
	c.SetParamValues("Patient", "123")

	err := handler.GetMeta(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if result["resourceType"] != "Parameters" {
		t.Errorf("expected resourceType Parameters, got %v", result["resourceType"])
	}

	params, ok := result["parameter"].([]interface{})
	if !ok || len(params) == 0 {
		t.Fatal("expected non-empty parameter array")
	}

	param := params[0].(map[string]interface{})
	if param["name"] != "return" {
		t.Errorf("expected parameter name 'return', got %v", param["name"])
	}

	valueMeta, ok := param["valueMeta"].(map[string]interface{})
	if !ok {
		t.Fatal("expected valueMeta in parameter")
	}

	profiles, ok := valueMeta["profile"].([]interface{})
	if !ok || len(profiles) != 1 {
		t.Errorf("expected 1 profile in response, got %v", valueMeta["profile"])
	}

	tags, ok := valueMeta["tag"].([]interface{})
	if !ok || len(tags) != 1 {
		t.Errorf("expected 1 tag in response, got %v", valueMeta["tag"])
	}
}

func TestMetaHandler_GetMeta_Nonexistent(t *testing.T) {
	store := NewInMemoryMetaStore()
	handler := NewMetaHandler(store)
	e := echo.New()

	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient/nonexistent/$meta", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("resourceType", "id")
	c.SetParamValues("Patient", "nonexistent")

	err := handler.GetMeta(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if result["resourceType"] != "Parameters" {
		t.Errorf("expected resourceType Parameters, got %v", result["resourceType"])
	}

	params := result["parameter"].([]interface{})
	param := params[0].(map[string]interface{})
	valueMeta := param["valueMeta"].(map[string]interface{})

	// Empty meta should have no profiles, security, or tags.
	if _, ok := valueMeta["profile"]; ok {
		t.Error("expected no profile key in empty meta")
	}
	if _, ok := valueMeta["security"]; ok {
		t.Error("expected no security key in empty meta")
	}
	if _, ok := valueMeta["tag"]; ok {
		t.Error("expected no tag key in empty meta")
	}
}

func TestMetaHandler_AddMeta_MergesTags(t *testing.T) {
	store := NewInMemoryMetaStore()
	ctx := context.Background()
	_, _ = store.AddMeta(ctx, "Patient", "123", &Meta{
		Tag: []Coding{
			{System: "http://example.org/tags", Code: "existing-tag"},
		},
	})

	handler := NewMetaHandler(store)
	e := echo.New()

	body := `{
		"resourceType": "Parameters",
		"parameter": [{
			"name": "meta",
			"valueMeta": {
				"tag": [
					{"system": "http://example.org/tags", "code": "new-tag", "display": "New Tag"}
				]
			}
		}]
	}`

	req := httptest.NewRequest(http.MethodPost, "/fhir/Patient/123/$meta-add", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("resourceType", "id")
	c.SetParamValues("Patient", "123")

	err := handler.AddMeta(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	params := result["parameter"].([]interface{})
	param := params[0].(map[string]interface{})
	valueMeta := param["valueMeta"].(map[string]interface{})
	tags := valueMeta["tag"].([]interface{})

	if len(tags) != 2 {
		t.Errorf("expected 2 tags after merge, got %d", len(tags))
	}

	codes := map[string]bool{}
	for _, tagRaw := range tags {
		tag := tagRaw.(map[string]interface{})
		codes[tag["code"].(string)] = true
	}
	if !codes["existing-tag"] {
		t.Error("expected existing-tag to be present")
	}
	if !codes["new-tag"] {
		t.Error("expected new-tag to be present")
	}
}

func TestMetaHandler_DeleteMeta_RemovesSpecificTags(t *testing.T) {
	store := NewInMemoryMetaStore()
	ctx := context.Background()
	_, _ = store.AddMeta(ctx, "Patient", "123", &Meta{
		Tag: []Coding{
			{System: "http://example.org/tags", Code: "keep-me"},
			{System: "http://example.org/tags", Code: "remove-me"},
			{System: "http://example.org/tags", Code: "also-keep"},
		},
	})

	handler := NewMetaHandler(store)
	e := echo.New()

	body := `{
		"resourceType": "Parameters",
		"parameter": [{
			"name": "meta",
			"valueMeta": {
				"tag": [
					{"system": "http://example.org/tags", "code": "remove-me"}
				]
			}
		}]
	}`

	req := httptest.NewRequest(http.MethodPost, "/fhir/Patient/123/$meta-delete", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("resourceType", "id")
	c.SetParamValues("Patient", "123")

	err := handler.DeleteMeta(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	params := result["parameter"].([]interface{})
	param := params[0].(map[string]interface{})
	valueMeta := param["valueMeta"].(map[string]interface{})
	tags := valueMeta["tag"].([]interface{})

	if len(tags) != 2 {
		t.Errorf("expected 2 tags remaining, got %d", len(tags))
	}

	for _, tagRaw := range tags {
		tag := tagRaw.(map[string]interface{})
		if tag["code"] == "remove-me" {
			t.Error("expected remove-me tag to be deleted")
		}
	}
}

func TestMetaHandler_AddMeta_InvalidBody(t *testing.T) {
	store := NewInMemoryMetaStore()
	handler := NewMetaHandler(store)
	e := echo.New()

	req := httptest.NewRequest(http.MethodPost, "/fhir/Patient/123/$meta-add", strings.NewReader("{invalid json"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("resourceType", "id")
	c.SetParamValues("Patient", "123")

	err := handler.AddMeta(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid JSON, got %d", rec.Code)
	}
}

func TestMetaHandler_DeleteMeta_InvalidBody(t *testing.T) {
	store := NewInMemoryMetaStore()
	handler := NewMetaHandler(store)
	e := echo.New()

	req := httptest.NewRequest(http.MethodPost, "/fhir/Patient/123/$meta-delete", strings.NewReader("{bad"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("resourceType", "id")
	c.SetParamValues("Patient", "123")

	err := handler.DeleteMeta(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid JSON, got %d", rec.Code)
	}
}

func TestMetaHandler_RegisterRoutes(t *testing.T) {
	store := NewInMemoryMetaStore()
	handler := NewMetaHandler(store)
	e := echo.New()
	fhirGroup := e.Group("/fhir")

	handler.RegisterRoutes(fhirGroup)

	routes := e.Routes()
	routePaths := make(map[string]bool)
	for _, r := range routes {
		routePaths[r.Method+":"+r.Path] = true
	}

	expected := []string{
		"GET:/fhir/:resourceType/:id/$meta",
		"POST:/fhir/:resourceType/:id/$meta",
		"POST:/fhir/:resourceType/:id/$meta-add",
		"POST:/fhir/:resourceType/:id/$meta-delete",
	}
	for _, path := range expected {
		if !routePaths[path] {
			t.Errorf("missing expected route: %s (registered: %v)", path, routePaths)
		}
	}
}

func TestMetaHandler_AddMeta_SecurityLabels(t *testing.T) {
	store := NewInMemoryMetaStore()
	handler := NewMetaHandler(store)
	e := echo.New()

	body := `{
		"resourceType": "Parameters",
		"parameter": [{
			"name": "meta",
			"valueMeta": {
				"security": [
					{"system": "http://terminology.hl7.org/CodeSystem/v3-Confidentiality", "code": "R", "display": "Restricted"}
				],
				"profile": ["http://hl7.org/fhir/StructureDefinition/Patient"]
			}
		}]
	}`

	req := httptest.NewRequest(http.MethodPost, "/fhir/Patient/456/$meta-add", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("resourceType", "id")
	c.SetParamValues("Patient", "456")

	err := handler.AddMeta(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	params := result["parameter"].([]interface{})
	param := params[0].(map[string]interface{})
	valueMeta := param["valueMeta"].(map[string]interface{})

	security, ok := valueMeta["security"].([]interface{})
	if !ok || len(security) != 1 {
		t.Fatalf("expected 1 security label, got %v", valueMeta["security"])
	}
	sec := security[0].(map[string]interface{})
	if sec["code"] != "R" {
		t.Errorf("expected security code R, got %v", sec["code"])
	}

	profiles, ok := valueMeta["profile"].([]interface{})
	if !ok || len(profiles) != 1 {
		t.Fatalf("expected 1 profile, got %v", valueMeta["profile"])
	}
}
