package openapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/labstack/echo/v4"
)

func newTestCapabilityBuilder() *fhir.CapabilityBuilder {
	b := fhir.NewCapabilityBuilder("http://localhost:8000/fhir", "1.0.0")
	b.AddResource("Patient", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "name", Type: "string"},
		{Name: "birthdate", Type: "date"},
	})
	b.AddResource("Observation", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "patient", Type: "reference"},
		{Name: "code", Type: "token"},
	})
	b.AddResource("Encounter", []string{"read", "search-type"}, []fhir.SearchParam{
		{Name: "patient", Type: "reference"},
	})
	return b
}

func TestGenerateSpec_Structure(t *testing.T) {
	b := newTestCapabilityBuilder()
	g := NewGenerator(b, "1.0.0", "http://localhost:8000")

	spec := g.GenerateSpec()

	// Check top-level fields
	if spec["openapi"] != "3.0.3" {
		t.Errorf("expected openapi '3.0.3', got %v", spec["openapi"])
	}

	info, ok := spec["info"].(map[string]interface{})
	if !ok {
		t.Fatal("expected info object")
	}
	if info["title"] != "Headless EHR FHIR R4 API" {
		t.Errorf("expected title 'Headless EHR FHIR R4 API', got %v", info["title"])
	}
	if info["version"] != "1.0.0" {
		t.Errorf("expected version '1.0.0', got %v", info["version"])
	}
	if info["description"] != "FHIR R4 compliant EHR API" {
		t.Errorf("expected description, got %v", info["description"])
	}

	servers, ok := spec["servers"].([]map[string]string)
	if !ok {
		t.Fatal("expected servers array")
	}
	if len(servers) != 1 {
		t.Fatalf("expected 1 server, got %d", len(servers))
	}
	if servers[0]["url"] != "http://localhost:8000" {
		t.Errorf("expected server URL 'http://localhost:8000', got %v", servers[0]["url"])
	}
}

func TestGenerateSpec_Paths(t *testing.T) {
	b := newTestCapabilityBuilder()
	g := NewGenerator(b, "1.0.0", "http://localhost:8000")

	spec := g.GenerateSpec()

	paths, ok := spec["paths"].(map[string]interface{})
	if !ok {
		t.Fatal("expected paths object")
	}

	// Should have search and read paths for each resource (3 resources = 6 paths)
	expectedPaths := []string{
		"/fhir/Patient",
		"/fhir/Patient/{id}",
		"/fhir/Observation",
		"/fhir/Observation/{id}",
		"/fhir/Encounter",
		"/fhir/Encounter/{id}",
	}

	for _, p := range expectedPaths {
		if _, exists := paths[p]; !exists {
			t.Errorf("missing expected path: %s", p)
		}
	}
}

func TestGenerateSpec_SearchPath(t *testing.T) {
	b := newTestCapabilityBuilder()
	g := NewGenerator(b, "1.0.0", "http://localhost:8000")

	spec := g.GenerateSpec()
	paths := spec["paths"].(map[string]interface{})

	patientPath, ok := paths["/fhir/Patient"].(map[string]interface{})
	if !ok {
		t.Fatal("expected /fhir/Patient path")
	}

	// Check GET (search)
	get, ok := patientPath["get"].(map[string]interface{})
	if !ok {
		t.Fatal("expected GET method on /fhir/Patient")
	}
	if get["summary"] != "Search Patient" {
		t.Errorf("expected summary 'Search Patient', got %v", get["summary"])
	}
	if get["operationId"] != "searchPatient" {
		t.Errorf("expected operationId 'searchPatient', got %v", get["operationId"])
	}
	tags, ok := get["tags"].([]string)
	if !ok || len(tags) != 1 || tags[0] != "Patient" {
		t.Errorf("expected tags [Patient], got %v", get["tags"])
	}

	// Check POST (create)
	post, ok := patientPath["post"].(map[string]interface{})
	if !ok {
		t.Fatal("expected POST method on /fhir/Patient")
	}
	if post["summary"] != "Create Patient" {
		t.Errorf("expected summary 'Create Patient', got %v", post["summary"])
	}
}

func TestGenerateSpec_ReadPath(t *testing.T) {
	b := newTestCapabilityBuilder()
	g := NewGenerator(b, "1.0.0", "http://localhost:8000")

	spec := g.GenerateSpec()
	paths := spec["paths"].(map[string]interface{})

	readPath, ok := paths["/fhir/Patient/{id}"].(map[string]interface{})
	if !ok {
		t.Fatal("expected /fhir/Patient/{id} path")
	}

	// Check GET (read)
	get, ok := readPath["get"].(map[string]interface{})
	if !ok {
		t.Fatal("expected GET method on /fhir/Patient/{id}")
	}
	if get["summary"] != "Read Patient" {
		t.Errorf("expected summary 'Read Patient', got %v", get["summary"])
	}
	params, ok := get["parameters"].([]map[string]interface{})
	if !ok || len(params) != 1 {
		t.Fatal("expected 1 parameter")
	}
	if params[0]["name"] != "id" {
		t.Errorf("expected parameter name 'id', got %v", params[0]["name"])
	}
	if params[0]["in"] != "path" {
		t.Errorf("expected parameter in 'path', got %v", params[0]["in"])
	}
	if params[0]["required"] != true {
		t.Error("expected parameter to be required")
	}

	// Check PUT (update)
	put, ok := readPath["put"].(map[string]interface{})
	if !ok {
		t.Fatal("expected PUT method on /fhir/Patient/{id}")
	}
	if put["summary"] != "Update Patient" {
		t.Errorf("expected summary 'Update Patient', got %v", put["summary"])
	}

	// Check DELETE
	del, ok := readPath["delete"].(map[string]interface{})
	if !ok {
		t.Fatal("expected DELETE method on /fhir/Patient/{id}")
	}
	if del["summary"] != "Delete Patient" {
		t.Errorf("expected summary 'Delete Patient', got %v", del["summary"])
	}
}

func TestGenerateSpec_EmptyCapability(t *testing.T) {
	b := fhir.NewCapabilityBuilder("http://localhost:8000/fhir", "1.0.0")
	g := NewGenerator(b, "1.0.0", "http://localhost:8000")

	spec := g.GenerateSpec()
	paths, ok := spec["paths"].(map[string]interface{})
	if !ok {
		t.Fatal("expected paths object")
	}
	if len(paths) != 0 {
		t.Errorf("expected 0 paths for empty capability, got %d", len(paths))
	}
}

func TestGenerateSpec_JSONSerialization(t *testing.T) {
	b := newTestCapabilityBuilder()
	g := NewGenerator(b, "1.0.0", "http://localhost:8000")

	spec := g.GenerateSpec()

	data, err := json.Marshal(spec)
	if err != nil {
		t.Fatalf("failed to marshal spec: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to unmarshal spec: %v", err)
	}

	if result["openapi"] != "3.0.3" {
		t.Errorf("expected openapi '3.0.3' after round-trip, got %v", result["openapi"])
	}
}

func TestGenerator_RegisterRoutes(t *testing.T) {
	b := newTestCapabilityBuilder()
	g := NewGenerator(b, "1.0.0", "http://localhost:8000")

	e := echo.New()
	apiGroup := e.Group("/api")
	g.RegisterRoutes(apiGroup)

	routes := e.Routes()
	routePaths := make(map[string]bool)
	for _, r := range routes {
		routePaths[r.Method+":"+r.Path] = true
	}

	if !routePaths["GET:/api/openapi.json"] {
		t.Error("missing route GET /api/openapi.json")
	}
	if !routePaths["GET:/api/docs"] {
		t.Error("missing route GET /api/docs")
	}
}

func TestGenerator_OpenAPIEndpoint(t *testing.T) {
	b := newTestCapabilityBuilder()
	g := NewGenerator(b, "1.0.0", "http://localhost:8000")

	e := echo.New()
	apiGroup := e.Group("/api")
	g.RegisterRoutes(apiGroup)

	req := httptest.NewRequest(http.MethodGet, "/api/openapi.json", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var spec map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &spec); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if spec["openapi"] != "3.0.3" {
		t.Errorf("expected openapi '3.0.3', got %v", spec["openapi"])
	}
}

func TestGenerator_DocsRedirect(t *testing.T) {
	b := newTestCapabilityBuilder()
	g := NewGenerator(b, "1.0.0", "http://localhost:8000")

	e := echo.New()
	apiGroup := e.Group("/api")
	g.RegisterRoutes(apiGroup)

	req := httptest.NewRequest(http.MethodGet, "/api/docs", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "swagger-ui") {
		t.Error("expected Swagger UI HTML page, not a redirect")
	}
	if !strings.Contains(body, "/api/openapi.json") {
		t.Error("expected docs page to reference /api/openapi.json")
	}
}

func TestGenerateSpec_ResponseStructure(t *testing.T) {
	b := fhir.NewCapabilityBuilder("http://localhost:8000/fhir", "1.0.0")
	b.AddResource("Patient", fhir.DefaultInteractions(), nil)
	g := NewGenerator(b, "1.0.0", "http://localhost:8000")

	spec := g.GenerateSpec()
	paths := spec["paths"].(map[string]interface{})

	// Check read path responses
	readPath := paths["/fhir/Patient/{id}"].(map[string]interface{})
	getOp := readPath["get"].(map[string]interface{})
	responses := getOp["responses"].(map[string]interface{})

	if _, ok := responses["200"]; !ok {
		t.Error("expected 200 response in read operation")
	}
	if _, ok := responses["404"]; !ok {
		t.Error("expected 404 response in read operation")
	}

	// Check create responses
	searchPath := paths["/fhir/Patient"].(map[string]interface{})
	postOp := searchPath["post"].(map[string]interface{})
	createResponses := postOp["responses"].(map[string]interface{})

	if _, ok := createResponses["201"]; !ok {
		t.Error("expected 201 response in create operation")
	}

	// Check delete responses
	deleteOp := readPath["delete"].(map[string]interface{})
	deleteResponses := deleteOp["responses"].(map[string]interface{})

	if _, ok := deleteResponses["204"]; !ok {
		t.Error("expected 204 response in delete operation")
	}
}

// =========================================================================
// New tests for FHIR resource schemas (TDD - written before implementation)
// =========================================================================

// TestOpenAPISpec_HasComponentSchemas verifies that spec.Components.Schemas is populated.
func TestOpenAPISpec_HasComponentSchemas(t *testing.T) {
	b := newTestCapabilityBuilder()
	g := NewGenerator(b, "1.0.0", "http://localhost:8000")

	spec := g.GenerateSpec()

	components, ok := spec["components"].(map[string]interface{})
	if !ok {
		t.Fatal("expected components object in spec")
	}

	schemas, ok := components["schemas"].(map[string]interface{})
	if !ok {
		t.Fatal("expected schemas object in components")
	}

	if len(schemas) == 0 {
		t.Error("expected schemas to be populated, got empty map")
	}

	// Core schemas that must always exist
	coreSchemas := []string{"Bundle", "BundleEntry", "OperationOutcome", "Reference", "Meta", "CodeableConcept", "HumanName", "Address", "ContactPoint", "Identifier", "Period", "Coding"}
	for _, name := range coreSchemas {
		if _, exists := schemas[name]; !exists {
			t.Errorf("missing core schema: %s", name)
		}
	}
}

// TestOpenAPISpec_PatientSchema verifies Patient schema has correct properties.
func TestOpenAPISpec_PatientSchema(t *testing.T) {
	b := fhir.NewCapabilityBuilder("http://localhost:8000/fhir", "1.0.0")
	b.AddResource("Patient", fhir.DefaultInteractions(), nil)
	g := NewGenerator(b, "1.0.0", "http://localhost:8000")

	spec := g.GenerateSpec()
	schemas := spec["components"].(map[string]interface{})["schemas"].(map[string]interface{})

	patientSchema, ok := schemas["Patient"].(map[string]interface{})
	if !ok {
		t.Fatal("expected Patient schema")
	}

	if patientSchema["type"] != "object" {
		t.Errorf("expected Patient type 'object', got %v", patientSchema["type"])
	}

	props, ok := patientSchema["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("expected Patient properties")
	}

	// Verify essential FHIR Patient properties
	requiredProps := []string{"resourceType", "id", "meta", "active", "name", "telecom", "gender", "birthDate", "address", "maritalStatus", "identifier"}
	for _, prop := range requiredProps {
		if _, exists := props[prop]; !exists {
			t.Errorf("Patient schema missing property: %s", prop)
		}
	}

	// Verify resourceType has const/enum value
	rtProp, ok := props["resourceType"].(map[string]interface{})
	if !ok {
		t.Fatal("expected resourceType property object")
	}
	if rtProp["type"] != "string" {
		t.Errorf("expected resourceType type 'string', got %v", rtProp["type"])
	}

	// Check enum contains "Patient"
	enumVal, ok := rtProp["enum"].([]string)
	if !ok || len(enumVal) != 1 || enumVal[0] != "Patient" {
		t.Errorf("expected resourceType enum ['Patient'], got %v", rtProp["enum"])
	}

	// Check name is array of HumanName refs
	nameProp, ok := props["name"].(map[string]interface{})
	if !ok {
		t.Fatal("expected name property object")
	}
	if nameProp["type"] != "array" {
		t.Errorf("expected name type 'array', got %v", nameProp["type"])
	}
	nameItems, ok := nameProp["items"].(map[string]interface{})
	if !ok {
		t.Fatal("expected name items")
	}
	if nameItems["$ref"] != "#/components/schemas/HumanName" {
		t.Errorf("expected name items $ref to HumanName, got %v", nameItems["$ref"])
	}

	// Check gender is string enum
	genderProp, ok := props["gender"].(map[string]interface{})
	if !ok {
		t.Fatal("expected gender property object")
	}
	if genderProp["type"] != "string" {
		t.Errorf("expected gender type 'string', got %v", genderProp["type"])
	}
	genderEnum, ok := genderProp["enum"].([]string)
	if !ok {
		t.Fatal("expected gender to have enum values")
	}
	expectedGenders := map[string]bool{"male": true, "female": true, "other": true, "unknown": true}
	for _, g := range genderEnum {
		if !expectedGenders[g] {
			t.Errorf("unexpected gender enum value: %s", g)
		}
	}

	// Check id has format uuid
	idProp, ok := props["id"].(map[string]interface{})
	if !ok {
		t.Fatal("expected id property object")
	}
	if idProp["type"] != "string" {
		t.Errorf("expected id type 'string', got %v", idProp["type"])
	}
	if idProp["format"] != "uuid" {
		t.Errorf("expected id format 'uuid', got %v", idProp["format"])
	}

	// Check meta references Meta schema
	metaProp, ok := props["meta"].(map[string]interface{})
	if !ok {
		t.Fatal("expected meta property object")
	}
	if metaProp["$ref"] != "#/components/schemas/Meta" {
		t.Errorf("expected meta $ref to Meta, got %v", metaProp["$ref"])
	}
}

// TestOpenAPISpec_BundleSchema verifies Bundle schema exists with entries array.
func TestOpenAPISpec_BundleSchema(t *testing.T) {
	b := newTestCapabilityBuilder()
	g := NewGenerator(b, "1.0.0", "http://localhost:8000")

	spec := g.GenerateSpec()
	schemas := spec["components"].(map[string]interface{})["schemas"].(map[string]interface{})

	bundleSchema, ok := schemas["Bundle"].(map[string]interface{})
	if !ok {
		t.Fatal("expected Bundle schema")
	}

	if bundleSchema["type"] != "object" {
		t.Errorf("expected Bundle type 'object', got %v", bundleSchema["type"])
	}

	props, ok := bundleSchema["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("expected Bundle properties")
	}

	// Check required properties
	requiredBundleProps := []string{"resourceType", "type", "total", "link", "entry"}
	for _, prop := range requiredBundleProps {
		if _, exists := props[prop]; !exists {
			t.Errorf("Bundle schema missing property: %s", prop)
		}
	}

	// Check entry is array of BundleEntry
	entryProp, ok := props["entry"].(map[string]interface{})
	if !ok {
		t.Fatal("expected entry property")
	}
	if entryProp["type"] != "array" {
		t.Errorf("expected entry type 'array', got %v", entryProp["type"])
	}
	entryItems, ok := entryProp["items"].(map[string]interface{})
	if !ok {
		t.Fatal("expected entry items")
	}
	if entryItems["$ref"] != "#/components/schemas/BundleEntry" {
		t.Errorf("expected entry items $ref to BundleEntry, got %v", entryItems["$ref"])
	}

	// Check total is integer
	totalProp, ok := props["total"].(map[string]interface{})
	if !ok {
		t.Fatal("expected total property")
	}
	if totalProp["type"] != "integer" {
		t.Errorf("expected total type 'integer', got %v", totalProp["type"])
	}

	// Check BundleEntry schema
	bundleEntry, ok := schemas["BundleEntry"].(map[string]interface{})
	if !ok {
		t.Fatal("expected BundleEntry schema")
	}
	beProps, ok := bundleEntry["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("expected BundleEntry properties")
	}

	// BundleEntry should have resource, search, fullUrl
	beRequiredProps := []string{"resource", "search", "fullUrl"}
	for _, prop := range beRequiredProps {
		if _, exists := beProps[prop]; !exists {
			t.Errorf("BundleEntry schema missing property: %s", prop)
		}
	}

	// fullUrl should be a string with format uri
	fullUrlProp, ok := beProps["fullUrl"].(map[string]interface{})
	if !ok {
		t.Fatal("expected fullUrl property")
	}
	if fullUrlProp["type"] != "string" {
		t.Errorf("expected fullUrl type 'string', got %v", fullUrlProp["type"])
	}
	if fullUrlProp["format"] != "uri" {
		t.Errorf("expected fullUrl format 'uri', got %v", fullUrlProp["format"])
	}
}

// TestOpenAPISpec_OperationOutcomeSchema verifies OperationOutcome schema exists.
func TestOpenAPISpec_OperationOutcomeSchema(t *testing.T) {
	b := newTestCapabilityBuilder()
	g := NewGenerator(b, "1.0.0", "http://localhost:8000")

	spec := g.GenerateSpec()
	schemas := spec["components"].(map[string]interface{})["schemas"].(map[string]interface{})

	ooSchema, ok := schemas["OperationOutcome"].(map[string]interface{})
	if !ok {
		t.Fatal("expected OperationOutcome schema")
	}

	props, ok := ooSchema["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("expected OperationOutcome properties")
	}

	// Check resourceType
	if _, exists := props["resourceType"]; !exists {
		t.Error("OperationOutcome missing resourceType property")
	}

	// Check issue array
	issueProp, ok := props["issue"].(map[string]interface{})
	if !ok {
		t.Fatal("expected issue property")
	}
	if issueProp["type"] != "array" {
		t.Errorf("expected issue type 'array', got %v", issueProp["type"])
	}

	issueItems, ok := issueProp["items"].(map[string]interface{})
	if !ok {
		t.Fatal("expected issue items")
	}

	// Issue items should have severity, code, diagnostics
	issueProps, ok := issueItems["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("expected issue item properties")
	}

	for _, prop := range []string{"severity", "code", "diagnostics"} {
		if _, exists := issueProps[prop]; !exists {
			t.Errorf("OperationOutcome issue missing property: %s", prop)
		}
	}

	// Verify severity has enum values
	sevProp, ok := issueProps["severity"].(map[string]interface{})
	if !ok {
		t.Fatal("expected severity property object")
	}
	sevEnum, ok := sevProp["enum"].([]string)
	if !ok {
		t.Fatal("expected severity to have enum values")
	}
	expectedSev := map[string]bool{"fatal": true, "error": true, "warning": true, "information": true}
	for _, s := range sevEnum {
		if !expectedSev[s] {
			t.Errorf("unexpected severity enum value: %s", s)
		}
	}
}

// TestOpenAPISpec_SearchParamsDocumented verifies GET operations have parameter definitions.
func TestOpenAPISpec_SearchParamsDocumented(t *testing.T) {
	b := newTestCapabilityBuilder()
	g := NewGenerator(b, "1.0.0", "http://localhost:8000")

	spec := g.GenerateSpec()
	paths := spec["paths"].(map[string]interface{})

	// Patient search should have name and birthdate params
	patientPath := paths["/fhir/Patient"].(map[string]interface{})
	getOp := patientPath["get"].(map[string]interface{})

	params, ok := getOp["parameters"].([]map[string]interface{})
	if !ok {
		t.Fatal("expected parameters array on Patient search operation")
	}

	paramNames := make(map[string]bool)
	for _, p := range params {
		name, _ := p["name"].(string)
		paramNames[name] = true

		// Verify each param has in=query
		if p["in"] != "query" {
			t.Errorf("expected search param '%s' to be in 'query', got %v", name, p["in"])
		}

		// Verify each param has a schema
		if _, hasSchema := p["schema"]; !hasSchema {
			t.Errorf("expected search param '%s' to have a schema", name)
		}
	}

	if !paramNames["name"] {
		t.Error("Patient search missing 'name' parameter")
	}
	if !paramNames["birthdate"] {
		t.Error("Patient search missing 'birthdate' parameter")
	}

	// Also check common FHIR params like _count are included
	if !paramNames["_count"] {
		t.Error("Patient search missing '_count' parameter")
	}
	if !paramNames["_offset"] {
		t.Error("Patient search missing '_offset' parameter")
	}

	// Observation search should have patient and code params
	obsPath := paths["/fhir/Observation"].(map[string]interface{})
	obsGet := obsPath["get"].(map[string]interface{})
	obsParams, ok := obsGet["parameters"].([]map[string]interface{})
	if !ok {
		t.Fatal("expected parameters array on Observation search operation")
	}

	obsParamNames := make(map[string]bool)
	for _, p := range obsParams {
		name, _ := p["name"].(string)
		obsParamNames[name] = true
	}
	if !obsParamNames["patient"] {
		t.Error("Observation search missing 'patient' parameter")
	}
	if !obsParamNames["code"] {
		t.Error("Observation search missing 'code' parameter")
	}
}

// TestOpenAPISpec_PostHasRequestBody verifies POST operations have requestBody with schema ref.
func TestOpenAPISpec_PostHasRequestBody(t *testing.T) {
	b := fhir.NewCapabilityBuilder("http://localhost:8000/fhir", "1.0.0")
	b.AddResource("Patient", fhir.DefaultInteractions(), nil)
	g := NewGenerator(b, "1.0.0", "http://localhost:8000")

	spec := g.GenerateSpec()
	paths := spec["paths"].(map[string]interface{})

	// Check POST /fhir/Patient
	patientPath := paths["/fhir/Patient"].(map[string]interface{})
	postOp := patientPath["post"].(map[string]interface{})

	requestBody, ok := postOp["requestBody"].(map[string]interface{})
	if !ok {
		t.Fatal("expected requestBody on POST operation")
	}

	if requestBody["required"] != true {
		t.Error("expected requestBody to be required")
	}

	content, ok := requestBody["content"].(map[string]interface{})
	if !ok {
		t.Fatal("expected content in requestBody")
	}

	fhirJSON, ok := content["application/fhir+json"].(map[string]interface{})
	if !ok {
		t.Fatal("expected application/fhir+json content type")
	}

	schema, ok := fhirJSON["schema"].(map[string]interface{})
	if !ok {
		t.Fatal("expected schema in content type")
	}

	if schema["$ref"] != "#/components/schemas/Patient" {
		t.Errorf("expected $ref '#/components/schemas/Patient', got %v", schema["$ref"])
	}

	// Check PUT /fhir/Patient/{id}
	readPath := paths["/fhir/Patient/{id}"].(map[string]interface{})
	putOp := readPath["put"].(map[string]interface{})

	putRequestBody, ok := putOp["requestBody"].(map[string]interface{})
	if !ok {
		t.Fatal("expected requestBody on PUT operation")
	}

	if putRequestBody["required"] != true {
		t.Error("expected PUT requestBody to be required")
	}

	putContent, ok := putRequestBody["content"].(map[string]interface{})
	if !ok {
		t.Fatal("expected content in PUT requestBody")
	}

	putFhirJSON, ok := putContent["application/fhir+json"].(map[string]interface{})
	if !ok {
		t.Fatal("expected application/fhir+json content type in PUT")
	}

	putSchema, ok := putFhirJSON["schema"].(map[string]interface{})
	if !ok {
		t.Fatal("expected schema in PUT content type")
	}

	if putSchema["$ref"] != "#/components/schemas/Patient" {
		t.Errorf("expected PUT $ref '#/components/schemas/Patient', got %v", putSchema["$ref"])
	}
}

// TestOpenAPISpec_ResponseSchemas verifies responses reference proper schemas.
func TestOpenAPISpec_ResponseSchemas(t *testing.T) {
	b := fhir.NewCapabilityBuilder("http://localhost:8000/fhir", "1.0.0")
	b.AddResource("Patient", fhir.DefaultInteractions(), nil)
	g := NewGenerator(b, "1.0.0", "http://localhost:8000")

	spec := g.GenerateSpec()
	paths := spec["paths"].(map[string]interface{})

	// Search response should reference Bundle
	searchPath := paths["/fhir/Patient"].(map[string]interface{})
	searchGet := searchPath["get"].(map[string]interface{})
	searchResponses := searchGet["responses"].(map[string]interface{})

	resp200, ok := searchResponses["200"].(map[string]interface{})
	if !ok {
		t.Fatal("expected 200 response object")
	}

	content, ok := resp200["content"].(map[string]interface{})
	if !ok {
		t.Fatal("expected content in 200 response")
	}

	fhirJSON, ok := content["application/fhir+json"].(map[string]interface{})
	if !ok {
		t.Fatal("expected application/fhir+json in 200 response")
	}

	schema, ok := fhirJSON["schema"].(map[string]interface{})
	if !ok {
		t.Fatal("expected schema in 200 response content")
	}

	if schema["$ref"] != "#/components/schemas/Bundle" {
		t.Errorf("expected search 200 $ref to Bundle, got %v", schema["$ref"])
	}

	// Read response should reference the resource type
	readPath := paths["/fhir/Patient/{id}"].(map[string]interface{})
	readGet := readPath["get"].(map[string]interface{})
	readResponses := readGet["responses"].(map[string]interface{})

	readResp200, ok := readResponses["200"].(map[string]interface{})
	if !ok {
		t.Fatal("expected 200 response in read operation")
	}

	readContent, ok := readResp200["content"].(map[string]interface{})
	if !ok {
		t.Fatal("expected content in read 200 response")
	}

	readFhirJSON, ok := readContent["application/fhir+json"].(map[string]interface{})
	if !ok {
		t.Fatal("expected application/fhir+json in read 200 response")
	}

	readSchema, ok := readFhirJSON["schema"].(map[string]interface{})
	if !ok {
		t.Fatal("expected schema in read 200 response content")
	}

	if readSchema["$ref"] != "#/components/schemas/Patient" {
		t.Errorf("expected read 200 $ref to Patient, got %v", readSchema["$ref"])
	}

	// 404 should reference OperationOutcome
	readResp404, ok := readResponses["404"].(map[string]interface{})
	if !ok {
		t.Fatal("expected 404 response in read operation")
	}

	resp404Content, ok := readResp404["content"].(map[string]interface{})
	if !ok {
		t.Fatal("expected content in 404 response")
	}

	resp404FHIR, ok := resp404Content["application/fhir+json"].(map[string]interface{})
	if !ok {
		t.Fatal("expected application/fhir+json in 404 response")
	}

	resp404Schema, ok := resp404FHIR["schema"].(map[string]interface{})
	if !ok {
		t.Fatal("expected schema in 404 response content")
	}

	if resp404Schema["$ref"] != "#/components/schemas/OperationOutcome" {
		t.Errorf("expected 404 $ref to OperationOutcome, got %v", resp404Schema["$ref"])
	}

	// Create 201 response should reference the resource type
	postOp := searchPath["post"].(map[string]interface{})
	postResponses := postOp["responses"].(map[string]interface{})
	resp201, ok := postResponses["201"].(map[string]interface{})
	if !ok {
		t.Fatal("expected 201 response")
	}

	resp201Content, ok := resp201["content"].(map[string]interface{})
	if !ok {
		t.Fatal("expected content in 201 response")
	}

	resp201FHIR, ok := resp201Content["application/fhir+json"].(map[string]interface{})
	if !ok {
		t.Fatal("expected application/fhir+json in 201 response")
	}

	resp201Schema, ok := resp201FHIR["schema"].(map[string]interface{})
	if !ok {
		t.Fatal("expected schema in 201 response content")
	}

	if resp201Schema["$ref"] != "#/components/schemas/Patient" {
		t.Errorf("expected 201 $ref to Patient, got %v", resp201Schema["$ref"])
	}
}

// TestOpenAPISpec_FHIRResourceSchemas verifies all registered resource types have schemas.
func TestOpenAPISpec_FHIRResourceSchemas(t *testing.T) {
	b := fhir.NewCapabilityBuilder("http://localhost:8000/fhir", "1.0.0")
	resourceTypes := []string{"Patient", "Practitioner", "Encounter", "Observation", "Condition", "MedicationRequest", "AllergyIntolerance", "DiagnosticReport"}
	for _, rt := range resourceTypes {
		b.AddResource(rt, fhir.DefaultInteractions(), nil)
	}
	g := NewGenerator(b, "1.0.0", "http://localhost:8000")

	spec := g.GenerateSpec()
	schemas := spec["components"].(map[string]interface{})["schemas"].(map[string]interface{})

	for _, rt := range resourceTypes {
		schema, ok := schemas[rt].(map[string]interface{})
		if !ok {
			t.Errorf("missing schema for registered resource type: %s", rt)
			continue
		}

		if schema["type"] != "object" {
			t.Errorf("expected %s type 'object', got %v", rt, schema["type"])
		}

		props, ok := schema["properties"].(map[string]interface{})
		if !ok {
			t.Errorf("expected properties in %s schema", rt)
			continue
		}

		// Every resource must have resourceType, id, meta
		for _, baseProp := range []string{"resourceType", "id", "meta"} {
			if _, exists := props[baseProp]; !exists {
				t.Errorf("%s schema missing base property: %s", rt, baseProp)
			}
		}

		// Verify resourceType enum is set correctly
		rtProp, ok := props["resourceType"].(map[string]interface{})
		if !ok {
			t.Errorf("expected resourceType property in %s", rt)
			continue
		}
		enumVal, ok := rtProp["enum"].([]string)
		if !ok || len(enumVal) != 1 || enumVal[0] != rt {
			t.Errorf("expected %s resourceType enum ['%s'], got %v", rt, rt, rtProp["enum"])
		}
	}
}

// TestOpenAPISpec_DetailedResourceSchemas verifies detailed schemas for major resources.
func TestOpenAPISpec_DetailedResourceSchemas(t *testing.T) {
	b := fhir.NewCapabilityBuilder("http://localhost:8000/fhir", "1.0.0")
	b.AddResource("Practitioner", fhir.DefaultInteractions(), nil)
	b.AddResource("Encounter", fhir.DefaultInteractions(), nil)
	b.AddResource("Observation", fhir.DefaultInteractions(), nil)
	b.AddResource("Condition", fhir.DefaultInteractions(), nil)
	b.AddResource("MedicationRequest", fhir.DefaultInteractions(), nil)
	g := NewGenerator(b, "1.0.0", "http://localhost:8000")

	spec := g.GenerateSpec()
	schemas := spec["components"].(map[string]interface{})["schemas"].(map[string]interface{})

	// Practitioner should have name, telecom, qualification
	practitioner := schemas["Practitioner"].(map[string]interface{})
	practProps := practitioner["properties"].(map[string]interface{})
	for _, p := range []string{"active", "name", "telecom", "gender", "birthDate", "qualification"} {
		if _, exists := practProps[p]; !exists {
			t.Errorf("Practitioner missing property: %s", p)
		}
	}

	// Encounter should have status, class, type, subject, participant, period
	encounter := schemas["Encounter"].(map[string]interface{})
	encProps := encounter["properties"].(map[string]interface{})
	for _, p := range []string{"status", "class", "type", "subject", "participant", "period"} {
		if _, exists := encProps[p]; !exists {
			t.Errorf("Encounter missing property: %s", p)
		}
	}

	// Observation should have status, category, code, subject, valueQuantity, valueString, valueCodeableConcept
	observation := schemas["Observation"].(map[string]interface{})
	obsProps := observation["properties"].(map[string]interface{})
	for _, p := range []string{"status", "category", "code", "subject", "valueQuantity", "valueString", "valueCodeableConcept"} {
		if _, exists := obsProps[p]; !exists {
			t.Errorf("Observation missing property: %s", p)
		}
	}

	// Condition should have clinicalStatus, verificationStatus, category, code, subject
	condition := schemas["Condition"].(map[string]interface{})
	condProps := condition["properties"].(map[string]interface{})
	for _, p := range []string{"clinicalStatus", "verificationStatus", "category", "code", "subject"} {
		if _, exists := condProps[p]; !exists {
			t.Errorf("Condition missing property: %s", p)
		}
	}

	// MedicationRequest should have status, intent, medicationCodeableConcept, subject, dosageInstruction
	medReq := schemas["MedicationRequest"].(map[string]interface{})
	medProps := medReq["properties"].(map[string]interface{})
	for _, p := range []string{"status", "intent", "medicationCodeableConcept", "subject", "dosageInstruction"} {
		if _, exists := medProps[p]; !exists {
			t.Errorf("MedicationRequest missing property: %s", p)
		}
	}
}

// TestOpenAPISpec_MetaSchema verifies the Meta schema structure.
func TestOpenAPISpec_MetaSchema(t *testing.T) {
	b := newTestCapabilityBuilder()
	g := NewGenerator(b, "1.0.0", "http://localhost:8000")

	spec := g.GenerateSpec()
	schemas := spec["components"].(map[string]interface{})["schemas"].(map[string]interface{})

	metaSchema, ok := schemas["Meta"].(map[string]interface{})
	if !ok {
		t.Fatal("expected Meta schema")
	}

	props, ok := metaSchema["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("expected Meta properties")
	}

	for _, prop := range []string{"versionId", "lastUpdated"} {
		if _, exists := props[prop]; !exists {
			t.Errorf("Meta missing property: %s", prop)
		}
	}

	// lastUpdated should have format date-time
	lastUpdated, ok := props["lastUpdated"].(map[string]interface{})
	if !ok {
		t.Fatal("expected lastUpdated property object")
	}
	if lastUpdated["type"] != "string" {
		t.Errorf("expected lastUpdated type 'string', got %v", lastUpdated["type"])
	}
	if lastUpdated["format"] != "date-time" {
		t.Errorf("expected lastUpdated format 'date-time', got %v", lastUpdated["format"])
	}
}

// TestOpenAPISpec_ReferenceSchema verifies the Reference schema structure.
func TestOpenAPISpec_ReferenceSchema(t *testing.T) {
	b := newTestCapabilityBuilder()
	g := NewGenerator(b, "1.0.0", "http://localhost:8000")

	spec := g.GenerateSpec()
	schemas := spec["components"].(map[string]interface{})["schemas"].(map[string]interface{})

	refSchema, ok := schemas["Reference"].(map[string]interface{})
	if !ok {
		t.Fatal("expected Reference schema")
	}

	props, ok := refSchema["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("expected Reference properties")
	}

	for _, prop := range []string{"reference", "type", "display"} {
		if _, exists := props[prop]; !exists {
			t.Errorf("Reference missing property: %s", prop)
		}
	}
}

// TestOpenAPISpec_DocsServesSwaggerUI verifies the docs endpoint serves a Swagger UI HTML page.
func TestOpenAPISpec_DocsServesSwaggerUI(t *testing.T) {
	b := newTestCapabilityBuilder()
	g := NewGenerator(b, "1.0.0", "http://localhost:8000")

	e := echo.New()
	apiGroup := e.Group("/api")
	g.RegisterRoutes(apiGroup)

	req := httptest.NewRequest(http.MethodGet, "/api/docs", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	contentType := rec.Header().Get("Content-Type")
	if !strings.Contains(contentType, "text/html") {
		t.Errorf("expected text/html content type, got %s", contentType)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "swagger-ui") {
		t.Error("expected body to contain swagger-ui reference")
	}
	if !strings.Contains(body, "/api/openapi.json") {
		t.Error("expected body to reference /api/openapi.json spec URL")
	}
}

// TestOpenAPISpec_SearchParamSchemaTypes verifies search param types map to correct OpenAPI types.
func TestOpenAPISpec_SearchParamSchemaTypes(t *testing.T) {
	b := fhir.NewCapabilityBuilder("http://localhost:8000/fhir", "1.0.0")
	b.AddResource("Patient", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "name", Type: "string"},
		{Name: "birthdate", Type: "date"},
		{Name: "active", Type: "token"},
		{Name: "general-practitioner", Type: "reference"},
		{Name: "address", Type: "string"},
	})
	g := NewGenerator(b, "1.0.0", "http://localhost:8000")

	spec := g.GenerateSpec()
	paths := spec["paths"].(map[string]interface{})
	patientPath := paths["/fhir/Patient"].(map[string]interface{})
	getOp := patientPath["get"].(map[string]interface{})
	params := getOp["parameters"].([]map[string]interface{})

	paramMap := make(map[string]map[string]interface{})
	for _, p := range params {
		name := p["name"].(string)
		paramMap[name] = p
	}

	// String type should map to string
	if nameParam, ok := paramMap["name"]; ok {
		schema := nameParam["schema"].(map[string]interface{})
		if schema["type"] != "string" {
			t.Errorf("expected name param schema type 'string', got %v", schema["type"])
		}
	}

	// Date type should map to string with format date
	if bdParam, ok := paramMap["birthdate"]; ok {
		schema := bdParam["schema"].(map[string]interface{})
		if schema["type"] != "string" {
			t.Errorf("expected birthdate param schema type 'string', got %v", schema["type"])
		}
		if schema["format"] != "date" {
			t.Errorf("expected birthdate param schema format 'date', got %v", schema["format"])
		}
	}
}

// TestOpenAPISpec_EmptyCapabilityStillHasCoreSchemas verifies core schemas exist even with no resources.
func TestOpenAPISpec_EmptyCapabilityStillHasCoreSchemas(t *testing.T) {
	b := fhir.NewCapabilityBuilder("http://localhost:8000/fhir", "1.0.0")
	g := NewGenerator(b, "1.0.0", "http://localhost:8000")

	spec := g.GenerateSpec()

	components, ok := spec["components"].(map[string]interface{})
	if !ok {
		t.Fatal("expected components even with empty capability")
	}

	schemas, ok := components["schemas"].(map[string]interface{})
	if !ok {
		t.Fatal("expected schemas even with empty capability")
	}

	// Core schemas should always be present
	for _, name := range []string{"Bundle", "BundleEntry", "OperationOutcome", "Reference", "Meta"} {
		if _, exists := schemas[name]; !exists {
			t.Errorf("missing core schema %s with empty capability", name)
		}
	}
}

// TestOpenAPISpec_FullJSONRoundTrip verifies the enriched spec can be serialized and deserialized.
func TestOpenAPISpec_FullJSONRoundTrip(t *testing.T) {
	b := fhir.NewCapabilityBuilder("http://localhost:8000/fhir", "1.0.0")
	b.AddResource("Patient", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "name", Type: "string"},
	})
	b.AddResource("Observation", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "patient", Type: "reference"},
	})
	g := NewGenerator(b, "1.0.0", "http://localhost:8000")

	spec := g.GenerateSpec()

	data, err := json.MarshalIndent(spec, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal enriched spec: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to unmarshal enriched spec: %v", err)
	}

	// Verify components survived round-trip
	components, ok := result["components"].(map[string]interface{})
	if !ok {
		t.Fatal("expected components after round-trip")
	}
	schemas, ok := components["schemas"].(map[string]interface{})
	if !ok {
		t.Fatal("expected schemas after round-trip")
	}

	if _, exists := schemas["Patient"]; !exists {
		t.Error("Patient schema missing after round-trip")
	}
	if _, exists := schemas["Bundle"]; !exists {
		t.Error("Bundle schema missing after round-trip")
	}
}

// TestOpenAPISpec_AllRegisteredResourceTypesHaveDetailedSchemas verifies all
// resource types registered in main.go have detailed schema definitions (not
// just base properties).
func TestOpenAPISpec_AllRegisteredResourceTypesHaveDetailedSchemas(t *testing.T) {
	b := fhir.NewCapabilityBuilder("http://localhost:8000/fhir", "1.0.0")

	// All resource types registered in main.go
	allTypes := []string{
		"Patient", "Practitioner",
		"Organization", "Location",
		"Encounter",
		"Condition", "Observation", "AllergyIntolerance", "Procedure",
		"Medication", "MedicationRequest", "MedicationAdministration", "MedicationDispense",
		"ServiceRequest", "DiagnosticReport", "ImagingStudy", "Specimen",
		"Appointment", "Schedule", "Slot",
		"Coverage", "Claim",
		"Consent", "DocumentReference", "Composition",
		"Communication",
		"ResearchStudy",
		"Questionnaire", "QuestionnaireResponse",
		"Immunization", "ImmunizationRecommendation",
		"CarePlan", "Goal",
		"FamilyMemberHistory",
		"RelatedPerson",
		"Provenance",
	}

	for _, rt := range allTypes {
		b.AddResource(rt, fhir.DefaultInteractions(), nil)
	}
	g := NewGenerator(b, "1.0.0", "http://localhost:8000")

	spec := g.GenerateSpec()
	schemas := spec["components"].(map[string]interface{})["schemas"].(map[string]interface{})

	for _, rt := range allTypes {
		schema, ok := schemas[rt].(map[string]interface{})
		if !ok {
			t.Errorf("missing schema for resource type: %s", rt)
			continue
		}

		props, ok := schema["properties"].(map[string]interface{})
		if !ok {
			t.Errorf("expected properties in %s schema", rt)
			continue
		}

		// Every resource must have more than just the 3 base properties (resourceType, id, meta)
		if len(props) <= 3 {
			t.Errorf("%s schema has only %d properties; expected detailed schema with more than base properties", rt, len(props))
		}
	}
}

// TestOpenAPISpec_NewResourceSchemaProperties verifies specific properties of
// newly added resource schemas.
func TestOpenAPISpec_NewResourceSchemaProperties(t *testing.T) {
	b := fhir.NewCapabilityBuilder("http://localhost:8000/fhir", "1.0.0")
	newTypes := []string{
		"Organization", "Location", "AllergyIntolerance", "Procedure",
		"Medication", "MedicationAdministration", "MedicationDispense",
		"ServiceRequest", "DiagnosticReport", "ImagingStudy", "Specimen",
		"Appointment", "Schedule", "Slot",
		"Coverage", "Claim", "Consent", "DocumentReference", "Composition",
		"Communication", "ResearchStudy", "Questionnaire", "QuestionnaireResponse",
		"Immunization", "ImmunizationRecommendation",
		"CarePlan", "Goal", "FamilyMemberHistory", "RelatedPerson", "Provenance",
	}
	for _, rt := range newTypes {
		b.AddResource(rt, fhir.DefaultInteractions(), nil)
	}
	g := NewGenerator(b, "1.0.0", "http://localhost:8000")

	spec := g.GenerateSpec()
	schemas := spec["components"].(map[string]interface{})["schemas"].(map[string]interface{})

	tests := []struct {
		resourceType string
		wantProps    []string
	}{
		{"Organization", []string{"active", "name", "type", "telecom", "address", "identifier"}},
		{"Location", []string{"status", "name", "type", "address", "identifier"}},
		{"AllergyIntolerance", []string{"clinicalStatus", "verificationStatus", "type", "code", "patient", "reaction", "identifier"}},
		{"Procedure", []string{"status", "code", "subject", "encounter", "performer", "identifier"}},
		{"Medication", []string{"code", "status", "form", "identifier"}},
		{"MedicationAdministration", []string{"status", "subject", "effectiveDateTime", "dosage", "identifier"}},
		{"MedicationDispense", []string{"status", "subject", "quantity", "identifier"}},
		{"ServiceRequest", []string{"status", "intent", "code", "subject", "requester", "identifier"}},
		{"DiagnosticReport", []string{"status", "code", "subject", "result", "conclusion", "identifier"}},
		{"ImagingStudy", []string{"status", "subject", "started", "modality", "identifier"}},
		{"Specimen", []string{"status", "type", "subject", "collection", "identifier"}},
		{"Appointment", []string{"status", "start", "end", "participant", "identifier"}},
		{"Schedule", []string{"active", "actor", "planningHorizon", "identifier"}},
		{"Slot", []string{"schedule", "status", "start", "end", "identifier"}},
		{"Coverage", []string{"status", "beneficiary", "payor", "period", "identifier"}},
		{"Claim", []string{"status", "type", "patient", "provider", "identifier"}},
		{"Consent", []string{"status", "scope", "category", "patient", "provision", "identifier"}},
		{"DocumentReference", []string{"status", "type", "subject", "content", "identifier"}},
		{"Composition", []string{"status", "type", "subject", "date", "author", "title", "section"}},
		{"Communication", []string{"status", "subject", "sender", "recipient", "payload", "identifier"}},
		{"ResearchStudy", []string{"status", "title", "description", "period", "identifier"}},
		{"Questionnaire", []string{"status", "name", "title", "item", "identifier"}},
		{"QuestionnaireResponse", []string{"status", "questionnaire", "subject", "authored", "item"}},
		{"Immunization", []string{"status", "vaccineCode", "patient", "lotNumber", "doseQuantity", "identifier"}},
		{"ImmunizationRecommendation", []string{"patient", "date", "recommendation", "identifier"}},
		{"CarePlan", []string{"status", "intent", "subject", "period", "activity", "identifier"}},
		{"Goal", []string{"lifecycleStatus", "description", "subject", "target", "identifier"}},
		{"FamilyMemberHistory", []string{"status", "patient", "relationship", "condition", "identifier"}},
		{"RelatedPerson", []string{"active", "patient", "relationship", "name", "identifier"}},
		{"Provenance", []string{"target", "recorded", "agent"}},
	}

	for _, tt := range tests {
		t.Run(tt.resourceType, func(t *testing.T) {
			schema, ok := schemas[tt.resourceType].(map[string]interface{})
			if !ok {
				t.Fatalf("missing schema for %s", tt.resourceType)
			}

			props, ok := schema["properties"].(map[string]interface{})
			if !ok {
				t.Fatalf("missing properties in %s schema", tt.resourceType)
			}

			for _, p := range tt.wantProps {
				if _, exists := props[p]; !exists {
					t.Errorf("%s schema missing expected property: %s", tt.resourceType, p)
				}
			}

			// Verify base properties exist on all resource types
			for _, baseProp := range []string{"resourceType", "id", "meta"} {
				if _, exists := props[baseProp]; !exists {
					t.Errorf("%s schema missing base property: %s", tt.resourceType, baseProp)
				}
			}
		})
	}
}

// TestOpenAPISpec_ResourceSchemaEnumValues verifies enum values on key resource fields.
func TestOpenAPISpec_ResourceSchemaEnumValues(t *testing.T) {
	b := fhir.NewCapabilityBuilder("http://localhost:8000/fhir", "1.0.0")
	b.AddResource("AllergyIntolerance", fhir.DefaultInteractions(), nil)
	b.AddResource("Immunization", fhir.DefaultInteractions(), nil)
	b.AddResource("CarePlan", fhir.DefaultInteractions(), nil)
	g := NewGenerator(b, "1.0.0", "http://localhost:8000")

	spec := g.GenerateSpec()
	schemas := spec["components"].(map[string]interface{})["schemas"].(map[string]interface{})

	// AllergyIntolerance.type enum
	aiSchema := schemas["AllergyIntolerance"].(map[string]interface{})
	aiProps := aiSchema["properties"].(map[string]interface{})
	aiType := aiProps["type"].(map[string]interface{})
	aiTypeEnum, ok := aiType["enum"].([]string)
	if !ok {
		t.Fatal("expected AllergyIntolerance.type to have enum")
	}
	aiTypeExpected := map[string]bool{"allergy": true, "intolerance": true}
	for _, v := range aiTypeEnum {
		if !aiTypeExpected[v] {
			t.Errorf("unexpected AllergyIntolerance.type enum value: %s", v)
		}
	}

	// Immunization.status enum
	immSchema := schemas["Immunization"].(map[string]interface{})
	immProps := immSchema["properties"].(map[string]interface{})
	immStatus := immProps["status"].(map[string]interface{})
	immStatusEnum, ok := immStatus["enum"].([]string)
	if !ok {
		t.Fatal("expected Immunization.status to have enum")
	}
	immStatusExpected := map[string]bool{"completed": true, "entered-in-error": true, "not-done": true}
	for _, v := range immStatusEnum {
		if !immStatusExpected[v] {
			t.Errorf("unexpected Immunization.status enum value: %s", v)
		}
	}

	// CarePlan.intent enum
	cpSchema := schemas["CarePlan"].(map[string]interface{})
	cpProps := cpSchema["properties"].(map[string]interface{})
	cpIntent := cpProps["intent"].(map[string]interface{})
	cpIntentEnum, ok := cpIntent["enum"].([]string)
	if !ok {
		t.Fatal("expected CarePlan.intent to have enum")
	}
	cpIntentExpected := map[string]bool{"proposal": true, "plan": true, "order": true, "option": true}
	for _, v := range cpIntentEnum {
		if !cpIntentExpected[v] {
			t.Errorf("unexpected CarePlan.intent enum value: %s", v)
		}
	}
}

// TestOpenAPISpec_PathsForNewResourceTypes verifies paths exist for newly schematized resources.
func TestOpenAPISpec_PathsForNewResourceTypes(t *testing.T) {
	b := fhir.NewCapabilityBuilder("http://localhost:8000/fhir", "1.0.0")
	newTypes := []string{
		"Organization", "Location", "AllergyIntolerance", "Procedure",
		"Immunization", "CarePlan", "Goal", "FamilyMemberHistory",
		"RelatedPerson", "Provenance",
	}
	for _, rt := range newTypes {
		b.AddResource(rt, fhir.DefaultInteractions(), nil)
	}
	g := NewGenerator(b, "1.0.0", "http://localhost:8000")

	spec := g.GenerateSpec()
	paths := spec["paths"].(map[string]interface{})

	for _, rt := range newTypes {
		searchPath := "/fhir/" + rt
		readPath := "/fhir/" + rt + "/{id}"

		if _, exists := paths[searchPath]; !exists {
			t.Errorf("missing search path: %s", searchPath)
		}
		if _, exists := paths[readPath]; !exists {
			t.Errorf("missing read path: %s", readPath)
		}

		// Verify POST has requestBody referencing the correct schema
		if pathObj, ok := paths[searchPath].(map[string]interface{}); ok {
			postOp := pathObj["post"].(map[string]interface{})
			rb := postOp["requestBody"].(map[string]interface{})
			content := rb["content"].(map[string]interface{})
			fhirJSON := content["application/fhir+json"].(map[string]interface{})
			schema := fhirJSON["schema"].(map[string]interface{})
			expectedRef := "#/components/schemas/" + rt
			if schema["$ref"] != expectedRef {
				t.Errorf("POST %s requestBody $ref = %v, want %s", searchPath, schema["$ref"], expectedRef)
			}
		}

		// Verify GET read response references the correct schema
		if pathObj, ok := paths[readPath].(map[string]interface{}); ok {
			getOp := pathObj["get"].(map[string]interface{})
			responses := getOp["responses"].(map[string]interface{})
			resp200 := responses["200"].(map[string]interface{})
			content := resp200["content"].(map[string]interface{})
			fhirJSON := content["application/fhir+json"].(map[string]interface{})
			schema := fhirJSON["schema"].(map[string]interface{})
			expectedRef := "#/components/schemas/" + rt
			if schema["$ref"] != expectedRef {
				t.Errorf("GET %s 200 $ref = %v, want %s", readPath, schema["$ref"], expectedRef)
			}

			// 404 should reference OperationOutcome
			resp404 := responses["404"].(map[string]interface{})
			content404 := resp404["content"].(map[string]interface{})
			fhir404 := content404["application/fhir+json"].(map[string]interface{})
			schema404 := fhir404["schema"].(map[string]interface{})
			if schema404["$ref"] != "#/components/schemas/OperationOutcome" {
				t.Errorf("GET %s 404 $ref = %v, want OperationOutcome", readPath, schema404["$ref"])
			}
		}
	}
}
