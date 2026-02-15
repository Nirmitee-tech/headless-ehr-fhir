package hl7v2

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
)

// =========== Handler Tests ===========

func TestHandler_ParseMessage(t *testing.T) {
	h := NewHandler()
	e := echo.New()

	body := "MSH|^~\\&|SendingApp|SendingFac|ReceivingApp|ReceivingFac|20240115143025||ADT^A01|MSG00001|P|2.5.1\rPID|1||MRN12345||Doe^John||19800515|M"

	req := httptest.NewRequest(http.MethodPost, "/api/v1/hl7v2/parse", strings.NewReader(body))
	req.Header.Set("Content-Type", "text/plain")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.ParseMessage(c)
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

	var result map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse JSON response: %v", err)
	}

	if result["type"] != "ADT^A01" {
		t.Errorf("expected type 'ADT^A01', got %v", result["type"])
	}
	if result["controlId"] != "MSG00001" {
		t.Errorf("expected controlId 'MSG00001', got %v", result["controlId"])
	}
	if result["version"] != "2.5.1" {
		t.Errorf("expected version '2.5.1', got %v", result["version"])
	}

	segments, ok := result["segments"].([]interface{})
	if !ok {
		t.Fatal("expected segments array in response")
	}
	if len(segments) < 2 {
		t.Errorf("expected at least 2 segments, got %d", len(segments))
	}
}

func TestHandler_ParseMessage_Invalid(t *testing.T) {
	h := NewHandler()
	e := echo.New()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/hl7v2/parse", strings.NewReader("this is not a valid hl7 message"))
	req.Header.Set("Content-Type", "text/plain")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.ParseMessage(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestHandler_GenerateADT(t *testing.T) {
	h := NewHandler()
	e := echo.New()

	reqBody := `{
		"event": "A01",
		"patient": {
			"resourceType": "Patient",
			"name": [{"family": "Doe", "given": ["John"]}],
			"birthDate": "1980-05-15",
			"gender": "male"
		},
		"encounter": {
			"resourceType": "Encounter",
			"class": {"code": "IMP"},
			"status": "in-progress"
		}
	}`

	req := httptest.NewRequest(http.MethodPost, "/api/v1/hl7v2/generate/adt", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.GenerateADTHandler(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d; body: %s", rec.Code, rec.Body.String())
	}

	body := rec.Body.String()
	if !strings.Contains(body, "MSH|") {
		t.Error("expected MSH segment in response")
	}
	if !strings.Contains(body, "ADT^A01") {
		t.Error("expected ADT^A01 in response")
	}
	if !strings.Contains(body, "Doe^John") {
		t.Error("expected patient name in response")
	}
}

func TestHandler_GenerateADT_MissingEvent(t *testing.T) {
	h := NewHandler()
	e := echo.New()

	reqBody := `{
		"patient": {
			"name": [{"family": "Doe", "given": ["John"]}]
		}
	}`

	req := httptest.NewRequest(http.MethodPost, "/api/v1/hl7v2/generate/adt", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.GenerateADTHandler(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestHandler_GenerateORM(t *testing.T) {
	h := NewHandler()
	e := echo.New()

	reqBody := `{
		"serviceRequest": {
			"resourceType": "ServiceRequest",
			"code": {
				"coding": [{"system": "http://loinc.org", "code": "85025", "display": "CBC"}]
			},
			"authoredOn": "2024-01-15T12:00:00Z"
		},
		"patient": {
			"name": [{"family": "Doe", "given": ["John"]}],
			"birthDate": "1980-05-15",
			"gender": "male"
		}
	}`

	req := httptest.NewRequest(http.MethodPost, "/api/v1/hl7v2/generate/orm", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.GenerateORMHandler(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d; body: %s", rec.Code, rec.Body.String())
	}

	body := rec.Body.String()
	if !strings.Contains(body, "ORM^O01") {
		t.Error("expected ORM^O01 in response")
	}
	if !strings.Contains(body, "ORC|") {
		t.Error("expected ORC segment in response")
	}
}

func TestHandler_GenerateORU(t *testing.T) {
	h := NewHandler()
	e := echo.New()

	reqBody := `{
		"diagnosticReport": {
			"resourceType": "DiagnosticReport",
			"code": {
				"coding": [{"system": "http://loinc.org", "code": "85025", "display": "CBC"}]
			},
			"effectiveDateTime": "2024-01-15T14:00:00Z"
		},
		"observations": [
			{
				"resourceType": "Observation",
				"code": {"coding": [{"code": "718-7", "display": "Hemoglobin"}]},
				"valueQuantity": {"value": 13.5, "unit": "g/dL"},
				"status": "final"
			}
		],
		"patient": {
			"name": [{"family": "Doe", "given": ["John"]}],
			"birthDate": "1980-05-15",
			"gender": "male"
		}
	}`

	req := httptest.NewRequest(http.MethodPost, "/api/v1/hl7v2/generate/oru", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.GenerateORUHandler(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d; body: %s", rec.Code, rec.Body.String())
	}

	body := rec.Body.String()
	if !strings.Contains(body, "ORU^R01") {
		t.Error("expected ORU^R01 in response")
	}
	if !strings.Contains(body, "OBX|") {
		t.Error("expected OBX segment in response")
	}
}

func TestHandler_GenerateORU_NoObservations(t *testing.T) {
	h := NewHandler()
	e := echo.New()

	reqBody := `{
		"diagnosticReport": {
			"resourceType": "DiagnosticReport",
			"code": {
				"coding": [{"code": "85025", "display": "CBC"}]
			}
		},
		"observations": [],
		"patient": {
			"name": [{"family": "Doe", "given": ["John"]}]
		}
	}`

	req := httptest.NewRequest(http.MethodPost, "/api/v1/hl7v2/generate/oru", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.GenerateORUHandler(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d; body: %s", rec.Code, rec.Body.String())
	}

	body := rec.Body.String()
	if !strings.Contains(body, "ORU^R01") {
		t.Error("expected ORU^R01 in response")
	}
}

func TestHandler_ParseMessage_EmptyBody(t *testing.T) {
	h := NewHandler()
	e := echo.New()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/hl7v2/parse", strings.NewReader(""))
	req.Header.Set("Content-Type", "text/plain")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.ParseMessage(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestHandler_RegisterRoutes(t *testing.T) {
	h := NewHandler()
	e := echo.New()

	g := e.Group("/api/v1")
	h.RegisterRoutes(g)

	routes := e.Routes()
	routePaths := make(map[string]bool)
	for _, r := range routes {
		routePaths[r.Method+":"+r.Path] = true
	}

	expected := []string{
		"POST:/api/v1/hl7v2/parse",
		"POST:/api/v1/hl7v2/generate/adt",
		"POST:/api/v1/hl7v2/generate/orm",
		"POST:/api/v1/hl7v2/generate/oru",
	}
	for _, path := range expected {
		if !routePaths[path] {
			t.Errorf("missing expected route: %s", path)
		}
	}
}
