package terminology

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
)

func newTestHandler() (*Handler, *echo.Echo) {
	svc := newTestService()
	h := NewHandler(svc)
	e := echo.New()
	return h, e
}

// =========== SearchLOINC Handler Tests ===========

func TestHandler_SearchLOINC_Success(t *testing.T) {
	h, e := newTestHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/terminology/loinc?q=heart", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.SearchLOINC(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var results []*LOINCCode
	json.Unmarshal(rec.Body.Bytes(), &results)
	if len(results) == 0 {
		t.Error("expected results")
	}
}

func TestHandler_SearchLOINC_MissingQuery(t *testing.T) {
	h, e := newTestHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/terminology/loinc", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.SearchLOINC(c)
	if err == nil {
		t.Error("expected error for missing query parameter")
	}
}

// =========== SearchICD10 Handler Tests ===========

func TestHandler_SearchICD10_Success(t *testing.T) {
	h, e := newTestHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/terminology/icd10?q=diabetes", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.SearchICD10(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_SearchICD10_MissingQuery(t *testing.T) {
	h, e := newTestHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/terminology/icd10", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.SearchICD10(c)
	if err == nil {
		t.Error("expected error for missing query parameter")
	}
}

// =========== SearchSNOMED Handler Tests ===========

func TestHandler_SearchSNOMED_Success(t *testing.T) {
	h, e := newTestHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/terminology/snomed?q=appendectomy", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.SearchSNOMED(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_SearchSNOMED_MissingQuery(t *testing.T) {
	h, e := newTestHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/terminology/snomed", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.SearchSNOMED(c)
	if err == nil {
		t.Error("expected error for missing query parameter")
	}
}

// =========== SearchRxNorm Handler Tests ===========

func TestHandler_SearchRxNorm_Success(t *testing.T) {
	h, e := newTestHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/terminology/rxnorm?q=metformin", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.SearchRxNorm(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_SearchRxNorm_MissingQuery(t *testing.T) {
	h, e := newTestHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/terminology/rxnorm", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.SearchRxNorm(c)
	if err == nil {
		t.Error("expected error for missing query parameter")
	}
}

// =========== SearchCPT Handler Tests ===========

func TestHandler_SearchCPT_Success(t *testing.T) {
	h, e := newTestHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/terminology/cpt?q=99213", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.SearchCPT(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_SearchCPT_MissingQuery(t *testing.T) {
	h, e := newTestHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/terminology/cpt", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.SearchCPT(c)
	if err == nil {
		t.Error("expected error for missing query parameter")
	}
}

// =========== FHIR $lookup Handler Tests ===========

func TestHandler_FHIRLookup_Success(t *testing.T) {
	h, e := newTestHandler()

	body := `{"system":"http://loinc.org","code":"8310-5"}`
	req := httptest.NewRequest(http.MethodPost, "/fhir/CodeSystem/$lookup", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.FHIRLookup(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var resp LookupResponse
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp.ResourceType != "Parameters" {
		t.Errorf("expected resourceType 'Parameters', got %q", resp.ResourceType)
	}
}

func TestHandler_FHIRLookup_NotFound(t *testing.T) {
	h, e := newTestHandler()

	body := `{"system":"http://loinc.org","code":"99999-9"}`
	req := httptest.NewRequest(http.MethodPost, "/fhir/CodeSystem/$lookup", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.FHIRLookup(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestHandler_FHIRLookup_BadRequest(t *testing.T) {
	h, e := newTestHandler()

	body := `{invalid json`
	req := httptest.NewRequest(http.MethodPost, "/fhir/CodeSystem/$lookup", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.FHIRLookup(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

// =========== FHIR $validate-code Handler Tests ===========

func TestHandler_FHIRValidateCode_Valid(t *testing.T) {
	h, e := newTestHandler()

	body := `{"system":"http://loinc.org","code":"8310-5"}`
	req := httptest.NewRequest(http.MethodPost, "/fhir/CodeSystem/$validate-code", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.FHIRValidateCode(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var resp ValidateCodeResponse
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp.ResourceType != "Parameters" {
		t.Errorf("expected resourceType 'Parameters', got %q", resp.ResourceType)
	}
}

func TestHandler_FHIRValidateCode_Invalid(t *testing.T) {
	h, e := newTestHandler()

	body := `{"system":"http://loinc.org","code":"99999-9"}`
	req := httptest.NewRequest(http.MethodPost, "/fhir/CodeSystem/$validate-code", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.FHIRValidateCode(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d (invalid code still returns 200 with result=false)", rec.Code)
	}
}

func TestHandler_FHIRValidateCode_BadRequest(t *testing.T) {
	h, e := newTestHandler()

	body := `{invalid json`
	req := httptest.NewRequest(http.MethodPost, "/fhir/CodeSystem/$validate-code", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.FHIRValidateCode(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

// =========== ExpandValueSet Handler Tests ===========

func TestHandler_ExpandValueSet_LOINC(t *testing.T) {
	h, e := newTestHandler()

	req := httptest.NewRequest(http.MethodGet, "/fhir/ValueSet/$expand?url=http://loinc.org/vs&filter=heart&count=10", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.ExpandValueSet(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var result map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &result)
	if result["resourceType"] != "ValueSet" {
		t.Errorf("expected resourceType 'ValueSet', got %v", result["resourceType"])
	}
	expansion, ok := result["expansion"].(map[string]interface{})
	if !ok {
		t.Fatal("expected expansion object")
	}
	if expansion["identifier"] == nil {
		t.Error("expected expansion identifier")
	}
	contains, ok := expansion["contains"].([]interface{})
	if !ok {
		t.Fatal("expected contains array")
	}
	if len(contains) == 0 {
		t.Error("expected results for LOINC 'heart' filter")
	}
}

func TestHandler_ExpandValueSet_ICD10(t *testing.T) {
	h, e := newTestHandler()

	req := httptest.NewRequest(http.MethodGet, "/fhir/ValueSet/$expand?url=http://hl7.org/fhir/ValueSet/icd10&filter=diabetes", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.ExpandValueSet(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var result map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &result)
	expansion := result["expansion"].(map[string]interface{})
	contains := expansion["contains"].([]interface{})
	if len(contains) == 0 {
		t.Error("expected results for ICD-10 'diabetes' filter")
	}
}

func TestHandler_ExpandValueSet_SNOMED(t *testing.T) {
	h, e := newTestHandler()

	req := httptest.NewRequest(http.MethodGet, "/fhir/ValueSet/$expand?url=http://snomed.info/sct/vs&filter=appendectomy", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.ExpandValueSet(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var result map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &result)
	expansion := result["expansion"].(map[string]interface{})
	contains := expansion["contains"].([]interface{})
	if len(contains) == 0 {
		t.Error("expected results for SNOMED 'appendectomy' filter")
	}
}

func TestHandler_ExpandValueSet_EmptyURL(t *testing.T) {
	h, e := newTestHandler()

	req := httptest.NewRequest(http.MethodGet, "/fhir/ValueSet/$expand?filter=test", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.ExpandValueSet(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var result map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &result)
	expansion := result["expansion"].(map[string]interface{})
	contains := expansion["contains"].([]interface{})
	if len(contains) != 0 {
		t.Errorf("expected empty contains for unknown URL, got %d", len(contains))
	}
}

func TestHandler_ExpandValueSet_NoFilter(t *testing.T) {
	h, e := newTestHandler()

	req := httptest.NewRequest(http.MethodGet, "/fhir/ValueSet/$expand?url=http://loinc.org/vs", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.ExpandValueSet(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var result map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &result)
	expansion := result["expansion"].(map[string]interface{})
	contains := expansion["contains"].([]interface{})
	// No filter means empty results (filter is required for search)
	if len(contains) != 0 {
		t.Errorf("expected empty contains without filter, got %d", len(contains))
	}
}

func TestHandler_ExpandValueSet_WithOffset(t *testing.T) {
	h, e := newTestHandler()

	req := httptest.NewRequest(http.MethodGet, "/fhir/ValueSet/$expand?url=http://loinc.org/vs&filter=hemoglobin&offset=0&count=10", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.ExpandValueSet(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var result map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &result)
	expansion := result["expansion"].(map[string]interface{})
	offset := expansion["offset"].(float64)
	if offset != 0 {
		t.Errorf("expected offset 0, got %v", offset)
	}
}

// =========== Route Registration Tests ===========

func TestHandler_RegisterRoutes(t *testing.T) {
	h, e := newTestHandler()
	api := e.Group("/api/v1")
	fhirGroup := e.Group("/fhir")

	h.RegisterRoutes(api, fhirGroup)

	routes := e.Routes()
	if len(routes) == 0 {
		t.Error("expected routes to be registered")
	}

	routePaths := make(map[string]bool)
	for _, r := range routes {
		routePaths[r.Method+":"+r.Path] = true
	}

	expected := []string{
		"GET:/api/v1/terminology/loinc",
		"GET:/api/v1/terminology/icd10",
		"GET:/api/v1/terminology/snomed",
		"GET:/api/v1/terminology/rxnorm",
		"GET:/api/v1/terminology/cpt",
		"POST:/fhir/CodeSystem/$lookup",
		"POST:/fhir/CodeSystem/$validate-code",
		"GET:/fhir/ValueSet/$expand",
		"POST:/fhir/ValueSet/$expand",
	}
	for _, path := range expected {
		if !routePaths[path] {
			t.Errorf("missing expected route: %s", path)
		}
	}
}

// =========== getLimit Tests ===========

func TestGetLimit_Default(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	limit := getLimit(c)
	if limit != 20 {
		t.Errorf("expected default limit 20, got %d", limit)
	}
}

func TestGetLimit_CountParam(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/?_count=50", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	limit := getLimit(c)
	if limit != 50 {
		t.Errorf("expected limit 50, got %d", limit)
	}
}

func TestGetLimit_LimitParam(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/?limit=30", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	limit := getLimit(c)
	if limit != 30 {
		t.Errorf("expected limit 30, got %d", limit)
	}
}

func TestGetLimit_MaxCapped(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/?_count=500", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	limit := getLimit(c)
	if limit != 100 {
		t.Errorf("expected limit capped at 100, got %d", limit)
	}
}
