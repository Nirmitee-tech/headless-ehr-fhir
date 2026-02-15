package ccda

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
)

// =========== Mock DataFetcher ===========

type mockFetcher struct {
	data *PatientData
	err  error
}

func (m *mockFetcher) FetchPatientData(ctx context.Context, patientID string) (*PatientData, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.data, nil
}

// =========== Handler Tests ===========

func TestHandler_GenerateCCD_Success(t *testing.T) {
	gen := NewGenerator("Test Hospital", "2.16.840.1.113883.3.1234")
	parser := NewParser()
	fetcher := &mockFetcher{
		data: fullPatientData(),
	}
	h := NewHandler(gen, parser, fetcher)
	e := echo.New()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/patients/patient-123/ccd", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("patient-123")

	err := h.GenerateCCD(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	contentType := rec.Header().Get("Content-Type")
	if !strings.Contains(contentType, "application/xml") {
		t.Errorf("expected Content-Type containing 'application/xml', got %q", contentType)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "ClinicalDocument") {
		t.Error("expected ClinicalDocument in response body")
	}
	if !strings.Contains(body, "John") {
		t.Error("expected patient name in response body")
	}
}

func TestHandler_GenerateCCD_PatientNotFound(t *testing.T) {
	gen := NewGenerator("Test Hospital", "2.16.840.1.113883.3.1234")
	parser := NewParser()
	fetcher := &mockFetcher{
		err: fmt.Errorf("patient not found"),
	}
	h := NewHandler(gen, parser, fetcher)
	e := echo.New()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/patients/nonexistent/ccd", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("nonexistent")

	err := h.GenerateCCD(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestHandler_ParseCCDA_Success(t *testing.T) {
	gen := NewGenerator("Test Hospital", "2.16.840.1.113883.3.1234")
	parser := NewParser()
	fetcher := &mockFetcher{}
	h := NewHandler(gen, parser, fetcher)
	e := echo.New()

	// First generate a valid CCD document
	data := fullPatientData()
	xmlData, err := gen.GenerateCCD(data)
	if err != nil {
		t.Fatalf("failed to generate test CCD: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/ccda/parse", strings.NewReader(string(xmlData)))
	req.Header.Set("Content-Type", "application/xml")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err = h.ParseCCDA(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	contentType := rec.Header().Get("Content-Type")
	if !strings.Contains(contentType, "application/json") {
		t.Errorf("expected Content-Type containing 'application/json', got %q", contentType)
	}

	// Verify the JSON response has expected structure
	var result map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse JSON response: %v", err)
	}

	if result["Title"] != "Continuity of Care Document" {
		t.Errorf("expected Title 'Continuity of Care Document', got %v", result["Title"])
	}

	sections, ok := result["Sections"].([]interface{})
	if !ok {
		t.Fatal("expected Sections array in response")
	}
	if len(sections) == 0 {
		t.Error("expected at least one section in parsed output")
	}
}

func TestHandler_ParseCCDA_InvalidXML(t *testing.T) {
	gen := NewGenerator("Test Hospital", "2.16.840.1.113883.3.1234")
	parser := NewParser()
	fetcher := &mockFetcher{}
	h := NewHandler(gen, parser, fetcher)
	e := echo.New()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/ccda/parse", strings.NewReader("this is not valid xml"))
	req.Header.Set("Content-Type", "application/xml")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.ParseCCDA(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestHandler_RegisterRoutes(t *testing.T) {
	gen := NewGenerator("Test Hospital", "2.16.840.1.113883.3.1234")
	parser := NewParser()
	fetcher := &mockFetcher{}
	h := NewHandler(gen, parser, fetcher)
	e := echo.New()

	g := e.Group("/api/v1")
	h.RegisterRoutes(g)

	routes := e.Routes()
	routePaths := make(map[string]bool)
	for _, r := range routes {
		routePaths[r.Method+":"+r.Path] = true
	}

	expected := []string{
		"GET:/api/v1/patients/:id/ccd",
		"POST:/api/v1/ccda/parse",
	}
	for _, path := range expected {
		if !routePaths[path] {
			t.Errorf("missing expected route: %s", path)
		}
	}
}
