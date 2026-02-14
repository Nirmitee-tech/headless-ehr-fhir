package openapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
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

	if rec.Code != http.StatusFound {
		t.Errorf("expected 302, got %d", rec.Code)
	}
	location := rec.Header().Get("Location")
	if location == "" {
		t.Error("expected Location header for redirect")
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
