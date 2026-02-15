package fhir

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
)

// =========== InMemoryGraphQLResolver ===========

func newTestGraphQLEngine() (*GraphQLEngine, *InMemoryResourceResolver) {
	resolver := NewInMemoryResourceResolver()
	resolver.AddResource("Patient", map[string]interface{}{
		"resourceType": "Patient",
		"id":           "123",
		"name": []interface{}{
			map[string]interface{}{
				"family": "Smith",
				"given":  []interface{}{"John"},
			},
		},
		"birthDate": "1990-01-15",
		"gender":    "male",
	})
	resolver.AddResource("Patient", map[string]interface{}{
		"resourceType": "Patient",
		"id":           "456",
		"name": []interface{}{
			map[string]interface{}{
				"family": "Smith",
				"given":  []interface{}{"Jane"},
			},
		},
		"birthDate": "1985-06-20",
		"gender":    "female",
	})
	resolver.AddResource("Patient", map[string]interface{}{
		"resourceType": "Patient",
		"id":           "789",
		"name": []interface{}{
			map[string]interface{}{
				"family": "Jones",
				"given":  []interface{}{"Bob"},
			},
		},
		"birthDate": "2000-03-10",
		"gender":    "male",
	})
	resolver.AddResource("Observation", map[string]interface{}{
		"resourceType": "Observation",
		"id":           "obs-1",
		"status":       "final",
		"code": map[string]interface{}{
			"coding": []interface{}{
				map[string]interface{}{
					"system":  "http://loinc.org",
					"code":    "29463-7",
					"display": "Body Weight",
				},
			},
		},
		"subject": map[string]interface{}{
			"reference": "Patient/123",
		},
	})

	engine := NewGraphQLEngine()
	engine.RegisterResolver("Patient", resolver)
	engine.RegisterResolver("Observation", resolver)
	return engine, resolver
}

// =========== GraphQL Engine Tests ===========

func TestGraphQL_SingleResourceByID(t *testing.T) {
	engine, _ := newTestGraphQLEngine()
	req := GraphQLRequest{
		Query: `{ Patient(id: "123") { id, name, birthDate } }`,
	}
	resp := engine.Execute(context.Background(), req)
	if len(resp.Errors) > 0 {
		t.Fatalf("unexpected errors: %v", resp.Errors)
	}
	if resp.Data == nil {
		t.Fatal("expected data, got nil")
	}
	data, ok := resp.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map data, got %T", resp.Data)
	}
	patient, ok := data["Patient"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected Patient in data, got %T", data["Patient"])
	}
	if patient["id"] != "123" {
		t.Errorf("expected id 123, got %v", patient["id"])
	}
}

func TestGraphQL_ResourceNotFound(t *testing.T) {
	engine, _ := newTestGraphQLEngine()
	req := GraphQLRequest{
		Query: `{ Patient(id: "nonexistent") { id, name } }`,
	}
	resp := engine.Execute(context.Background(), req)
	if len(resp.Errors) == 0 {
		t.Fatal("expected errors for not found resource")
	}
	if !strings.Contains(resp.Errors[0].Message, "not found") {
		t.Errorf("expected 'not found' in error message, got %s", resp.Errors[0].Message)
	}
}

func TestGraphQL_FieldSelection(t *testing.T) {
	engine, _ := newTestGraphQLEngine()
	req := GraphQLRequest{
		Query: `{ Patient(id: "123") { id, birthDate } }`,
	}
	resp := engine.Execute(context.Background(), req)
	if len(resp.Errors) > 0 {
		t.Fatalf("unexpected errors: %v", resp.Errors)
	}
	data := resp.Data.(map[string]interface{})
	patient := data["Patient"].(map[string]interface{})
	if patient["id"] != "123" {
		t.Errorf("expected id 123, got %v", patient["id"])
	}
	if patient["birthDate"] != "1990-01-15" {
		t.Errorf("expected birthDate 1990-01-15, got %v", patient["birthDate"])
	}
	// name should NOT be present since we only asked for id and birthDate
	if _, exists := patient["name"]; exists {
		t.Error("expected 'name' field to be absent from response")
	}
	// gender should NOT be present
	if _, exists := patient["gender"]; exists {
		t.Error("expected 'gender' field to be absent from response")
	}
}

func TestGraphQL_SearchList(t *testing.T) {
	engine, _ := newTestGraphQLEngine()
	req := GraphQLRequest{
		Query: `{ PatientList(_count: "10") { id, name } }`,
	}
	resp := engine.Execute(context.Background(), req)
	if len(resp.Errors) > 0 {
		t.Fatalf("unexpected errors: %v", resp.Errors)
	}
	data := resp.Data.(map[string]interface{})
	list, ok := data["PatientList"].([]interface{})
	if !ok {
		t.Fatalf("expected array for PatientList, got %T", data["PatientList"])
	}
	if len(list) == 0 {
		t.Fatal("expected non-empty PatientList")
	}
}

func TestGraphQL_SearchWithParams(t *testing.T) {
	engine, _ := newTestGraphQLEngine()
	req := GraphQLRequest{
		Query: `{ PatientList(name: "Smith") { id, name } }`,
	}
	resp := engine.Execute(context.Background(), req)
	if len(resp.Errors) > 0 {
		t.Fatalf("unexpected errors: %v", resp.Errors)
	}
	data := resp.Data.(map[string]interface{})
	list, ok := data["PatientList"].([]interface{})
	if !ok {
		t.Fatalf("expected array for PatientList, got %T", data["PatientList"])
	}
	// Should only return Smiths
	for _, item := range list {
		resource := item.(map[string]interface{})
		names, ok := resource["name"].([]interface{})
		if !ok || len(names) == 0 {
			continue
		}
		nameObj := names[0].(map[string]interface{})
		if nameObj["family"] != "Smith" {
			t.Errorf("expected family Smith, got %v", nameObj["family"])
		}
	}
}

func TestGraphQL_UnknownResourceType(t *testing.T) {
	engine, _ := newTestGraphQLEngine()
	req := GraphQLRequest{
		Query: `{ UnknownResource(id: "1") { id } }`,
	}
	resp := engine.Execute(context.Background(), req)
	if len(resp.Errors) == 0 {
		t.Fatal("expected errors for unknown resource type")
	}
	if !strings.Contains(resp.Errors[0].Message, "no resolver") {
		t.Errorf("expected 'no resolver' in error, got %s", resp.Errors[0].Message)
	}
}

func TestGraphQL_InvalidQuery(t *testing.T) {
	engine, _ := newTestGraphQLEngine()
	req := GraphQLRequest{
		Query: `this is not a valid query`,
	}
	resp := engine.Execute(context.Background(), req)
	if len(resp.Errors) == 0 {
		t.Fatal("expected errors for invalid query")
	}
}

func TestGraphQL_EmptyQuery(t *testing.T) {
	engine, _ := newTestGraphQLEngine()
	req := GraphQLRequest{
		Query: "",
	}
	resp := engine.Execute(context.Background(), req)
	if len(resp.Errors) == 0 {
		t.Fatal("expected errors for empty query")
	}
}

func TestGraphQL_NestedFields(t *testing.T) {
	engine, _ := newTestGraphQLEngine()
	req := GraphQLRequest{
		Query: `{ Observation(id: "obs-1") { id, status, code, subject } }`,
	}
	resp := engine.Execute(context.Background(), req)
	if len(resp.Errors) > 0 {
		t.Fatalf("unexpected errors: %v", resp.Errors)
	}
	data := resp.Data.(map[string]interface{})
	obs := data["Observation"].(map[string]interface{})
	if obs["status"] != "final" {
		t.Errorf("expected status final, got %v", obs["status"])
	}
	// code should be the nested object
	code, ok := obs["code"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected code as map, got %T", obs["code"])
	}
	if code["coding"] == nil {
		t.Error("expected coding array in code")
	}
	// subject should be the nested reference
	subj, ok := obs["subject"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected subject as map, got %T", obs["subject"])
	}
	if subj["reference"] != "Patient/123" {
		t.Errorf("expected reference Patient/123, got %v", subj["reference"])
	}
}

func TestGraphQL_Variables(t *testing.T) {
	engine, _ := newTestGraphQLEngine()
	req := GraphQLRequest{
		Query: `{ Patient(id: $patientId) { id, name } }`,
		Variables: map[string]interface{}{
			"patientId": "123",
		},
	}
	resp := engine.Execute(context.Background(), req)
	if len(resp.Errors) > 0 {
		t.Fatalf("unexpected errors: %v", resp.Errors)
	}
	data := resp.Data.(map[string]interface{})
	patient := data["Patient"].(map[string]interface{})
	if patient["id"] != "123" {
		t.Errorf("expected id 123, got %v", patient["id"])
	}
}

// =========== GraphQL Handler Tests ===========

func TestGraphQLHandler_POST(t *testing.T) {
	engine, _ := newTestGraphQLEngine()
	handler := NewGraphQLHandler(engine)
	e := echo.New()

	body := `{"query": "{ Patient(id: \"123\") { id, name } }"}`
	req := httptest.NewRequest(http.MethodPost, "/fhir/$graphql", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.HandlePost(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var resp GraphQLResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp.Data == nil {
		t.Error("expected data in response")
	}
}

func TestGraphQLHandler_GET(t *testing.T) {
	engine, _ := newTestGraphQLEngine()
	handler := NewGraphQLHandler(engine)
	e := echo.New()

	q := url.Values{}
	q.Set("query", `{ Patient(id: "123") { id, name } }`)
	req := httptest.NewRequest(http.MethodGet, "/fhir/$graphql?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.HandleGet(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var resp GraphQLResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp.Data == nil {
		t.Error("expected data in response")
	}
}

func TestGraphQLHandler_EmptyBody(t *testing.T) {
	engine, _ := newTestGraphQLEngine()
	handler := NewGraphQLHandler(engine)
	e := echo.New()

	req := httptest.NewRequest(http.MethodPost, "/fhir/$graphql", strings.NewReader(""))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.HandlePost(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

// =========== Query Parser Tests ===========

func TestParseQuery_Simple(t *testing.T) {
	parsed, err := parseGraphQLQuery(`{ Patient(id: "123") { id, name } }`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if parsed.ResourceType != "Patient" {
		t.Errorf("expected Patient, got %s", parsed.ResourceType)
	}
	if parsed.IsList {
		t.Error("expected IsList to be false")
	}
	if parsed.ID != "123" {
		t.Errorf("expected ID 123, got %s", parsed.ID)
	}
}

func TestParseQuery_WithArgs(t *testing.T) {
	parsed, err := parseGraphQLQuery(`{ Patient(id: "abc", name: "Smith") { id, name, birthDate } }`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if parsed.ResourceType != "Patient" {
		t.Errorf("expected Patient, got %s", parsed.ResourceType)
	}
	if parsed.ID != "abc" {
		t.Errorf("expected ID abc, got %s", parsed.ID)
	}
	if parsed.Params["name"] != "Smith" {
		t.Errorf("expected param name=Smith, got %s", parsed.Params["name"])
	}
	if len(parsed.Fields) != 3 {
		t.Errorf("expected 3 fields, got %d", len(parsed.Fields))
	}
}

func TestParseQuery_List(t *testing.T) {
	parsed, err := parseGraphQLQuery(`{ PatientList(name: "Smith", _count: "10") { id, name } }`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if parsed.ResourceType != "Patient" {
		t.Errorf("expected Patient, got %s", parsed.ResourceType)
	}
	if !parsed.IsList {
		t.Error("expected IsList to be true")
	}
	if parsed.Params["name"] != "Smith" {
		t.Errorf("expected param name=Smith, got %s", parsed.Params["name"])
	}
	if parsed.Params["_count"] != "10" {
		t.Errorf("expected param _count=10, got %s", parsed.Params["_count"])
	}
}
