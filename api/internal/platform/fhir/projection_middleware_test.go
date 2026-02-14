package fhir

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
)

// =========== ProjectionMiddleware Tests ===========

func newTestEchoWithMiddleware(handler echo.HandlerFunc) (*echo.Echo, string) {
	e := echo.New()
	e.Use(ProjectionMiddleware())
	e.GET("/fhir/Patient/:id", handler)
	return e, "/fhir/Patient/123"
}

func singleResourceHandler(c echo.Context) error {
	resource := map[string]interface{}{
		"resourceType":  "Patient",
		"id":            "123",
		"meta":          map[string]interface{}{"versionId": "1"},
		"name":          []interface{}{map[string]interface{}{"family": "Smith"}},
		"gender":        "male",
		"birthDate":     "1990-01-01",
		"text":          map[string]interface{}{"div": "<div>text</div>"},
		"communication": []interface{}{map[string]interface{}{"language": "en"}},
	}
	return c.JSON(http.StatusOK, resource)
}

func bundleHandler(c echo.Context) error {
	bundle := map[string]interface{}{
		"resourceType": "Bundle",
		"type":         "searchset",
		"total":        1,
		"entry": []interface{}{
			map[string]interface{}{
				"resource": map[string]interface{}{
					"resourceType":  "Patient",
					"id":            "1",
					"meta":          map[string]interface{}{"versionId": "1"},
					"name":          []interface{}{map[string]interface{}{"family": "Doe"}},
					"gender":        "female",
					"birthDate":     "1985-05-15",
					"communication": []interface{}{map[string]interface{}{"language": "en"}},
				},
			},
		},
	}
	return c.JSON(http.StatusOK, bundle)
}

func TestProjectionMiddleware_NoParams(t *testing.T) {
	e, path := newTestEchoWithMiddleware(singleResourceHandler)

	req := httptest.NewRequest(http.MethodGet, path, nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var result map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &result)

	// All fields should be present
	if result["name"] == nil {
		t.Error("name should be present without projection")
	}
	if result["gender"] == nil {
		t.Error("gender should be present without projection")
	}
	if result["birthDate"] == nil {
		t.Error("birthDate should be present without projection")
	}
	if result["communication"] == nil {
		t.Error("communication should be present without projection")
	}
}

func TestProjectionMiddleware_SummaryTrue(t *testing.T) {
	e, path := newTestEchoWithMiddleware(singleResourceHandler)

	req := httptest.NewRequest(http.MethodGet, path+"?_summary=true", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var result map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &result)

	// Mandatory fields
	if result["resourceType"] != "Patient" {
		t.Error("resourceType should be present")
	}
	if result["id"] != "123" {
		t.Error("id should be present")
	}

	// Summary fields for Patient (name, gender, birthDate are in summary)
	if result["name"] == nil {
		t.Error("name should be in Patient summary")
	}
	if result["gender"] == nil {
		t.Error("gender should be in Patient summary")
	}

	// Non-summary fields
	if _, ok := result["communication"]; ok {
		t.Error("communication should not be in Patient summary")
	}

	// Content-Type header
	contentType := rec.Header().Get("Content-Type")
	if contentType != "application/fhir+json" {
		t.Errorf("expected Content-Type 'application/fhir+json', got %q", contentType)
	}
}

func TestProjectionMiddleware_SummaryText(t *testing.T) {
	e, path := newTestEchoWithMiddleware(singleResourceHandler)

	req := httptest.NewRequest(http.MethodGet, path+"?_summary=text", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	var result map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &result)

	if result["id"] != "123" {
		t.Error("id should be present in text mode")
	}
	if result["text"] == nil {
		t.Error("text should be present in text mode")
	}
	if _, ok := result["name"]; ok {
		t.Error("name should not be in text mode")
	}
	if _, ok := result["gender"]; ok {
		t.Error("gender should not be in text mode")
	}
}

func TestProjectionMiddleware_SummaryData(t *testing.T) {
	e, path := newTestEchoWithMiddleware(singleResourceHandler)

	req := httptest.NewRequest(http.MethodGet, path+"?_summary=data", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	var result map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &result)

	if _, ok := result["text"]; ok {
		t.Error("text should be removed in data mode")
	}
	if result["name"] == nil {
		t.Error("name should be present in data mode")
	}
	if result["gender"] == nil {
		t.Error("gender should be present in data mode")
	}
}

func TestProjectionMiddleware_Elements(t *testing.T) {
	e, path := newTestEchoWithMiddleware(singleResourceHandler)

	req := httptest.NewRequest(http.MethodGet, path+"?_elements=name,gender", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	var result map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &result)

	// Mandatory fields always present
	if result["resourceType"] != "Patient" {
		t.Error("resourceType should always be present")
	}
	if result["id"] != "123" {
		t.Error("id should always be present")
	}

	// Requested fields
	if result["name"] == nil {
		t.Error("name should be present")
	}
	if result["gender"] == nil {
		t.Error("gender should be present")
	}

	// Non-requested fields
	if _, ok := result["birthDate"]; ok {
		t.Error("birthDate should not be present")
	}
	if _, ok := result["communication"]; ok {
		t.Error("communication should not be present")
	}
}

func TestProjectionMiddleware_ElementsPrecedence(t *testing.T) {
	e, path := newTestEchoWithMiddleware(singleResourceHandler)

	// Both _elements and _summary specified; _elements should take precedence
	req := httptest.NewRequest(http.MethodGet, path+"?_elements=name&_summary=true", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	var result map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &result)

	if result["name"] == nil {
		t.Error("name should be present (from _elements)")
	}
	if _, ok := result["gender"]; ok {
		t.Error("gender should not be present (not in _elements)")
	}
}

func TestProjectionMiddleware_BundleSummaryTrue(t *testing.T) {
	e := echo.New()
	e.Use(ProjectionMiddleware())
	e.GET("/fhir/Patient", bundleHandler)

	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient?_summary=true", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var result map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &result)

	// Bundle itself should be intact
	if result["resourceType"] != "Bundle" {
		t.Error("expected Bundle resourceType")
	}

	entries, ok := result["entry"].([]interface{})
	if !ok || len(entries) != 1 {
		t.Fatal("expected 1 entry in bundle")
	}

	entryMap := entries[0].(map[string]interface{})
	resource := entryMap["resource"].(map[string]interface{})

	// Summary fields should be present
	if resource["name"] == nil {
		t.Error("name should be in Patient summary")
	}
	if resource["gender"] == nil {
		t.Error("gender should be in Patient summary")
	}

	// Non-summary fields should be filtered
	if _, ok := resource["communication"]; ok {
		t.Error("communication should not be in Patient summary")
	}
}

func TestProjectionMiddleware_BundleElements(t *testing.T) {
	e := echo.New()
	e.Use(ProjectionMiddleware())
	e.GET("/fhir/Patient", bundleHandler)

	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient?_elements=name", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	var result map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &result)

	entries := result["entry"].([]interface{})
	entryMap := entries[0].(map[string]interface{})
	resource := entryMap["resource"].(map[string]interface{})

	if resource["name"] == nil {
		t.Error("name should be present")
	}
	if _, ok := resource["gender"]; ok {
		t.Error("gender should not be present")
	}
	if _, ok := resource["birthDate"]; ok {
		t.Error("birthDate should not be present")
	}
}

func TestProjectionMiddleware_NonJSONResponse(t *testing.T) {
	e := echo.New()
	e.Use(ProjectionMiddleware())
	e.GET("/fhir/test", func(c echo.Context) error {
		return c.String(http.StatusOK, "not json")
	})

	req := httptest.NewRequest(http.MethodGet, "/fhir/test?_summary=true", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	if rec.Body.String() != "not json" {
		t.Errorf("expected original body, got %q", rec.Body.String())
	}
}

func TestProjectionMiddleware_HandlerError(t *testing.T) {
	e := echo.New()
	e.Use(ProjectionMiddleware())
	e.GET("/fhir/error", func(c echo.Context) error {
		return echo.NewHTTPError(http.StatusInternalServerError, "test error")
	})

	req := httptest.NewRequest(http.MethodGet, "/fhir/error?_summary=true", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	// Error should propagate
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rec.Code)
	}
}

func TestProjectionMiddleware_ElementsIDStatus(t *testing.T) {
	e := echo.New()
	e.Use(ProjectionMiddleware())
	e.GET("/fhir/Patient/:id", func(c echo.Context) error {
		resource := map[string]interface{}{
			"resourceType": "Patient",
			"id":           "abc",
			"meta":         map[string]interface{}{"versionId": "1"},
			"active":       true,
			"name":         "John",
			"gender":       "male",
		}
		return c.JSON(http.StatusOK, resource)
	})

	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient/abc?_elements=id,active", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	var result map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &result)

	if result["id"] != "abc" {
		t.Error("id should be present")
	}
	if result["active"] != true {
		t.Error("active should be present")
	}
	if _, ok := result["name"]; ok {
		t.Error("name should not be present")
	}
	if _, ok := result["gender"]; ok {
		t.Error("gender should not be present")
	}
	// resourceType and meta are mandatory
	if result["resourceType"] != "Patient" {
		t.Error("resourceType should always be present")
	}
}
