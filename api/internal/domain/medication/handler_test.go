package medication

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func newTestHandler() (*Handler, *echo.Echo) {
	svc := newTestService()
	h := NewHandler(svc)
	e := echo.New()
	return h, e
}

func TestHandler_CreateMedication(t *testing.T) {
	h, e := newTestHandler()

	body := `{"code_value":"12345","code_display":"Aspirin"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/medications", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreateMedication(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}

	var m Medication
	json.Unmarshal(rec.Body.Bytes(), &m)
	if m.CodeDisplay != "Aspirin" {
		t.Errorf("expected 'Aspirin', got %s", m.CodeDisplay)
	}
}

func TestHandler_CreateMedication_BadRequest(t *testing.T) {
	h, e := newTestHandler()

	body := `{"code_display":"Aspirin"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/medications", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreateMedication(c)
	if err == nil {
		t.Error("expected error for missing code_value")
	}
}

func TestHandler_GetMedication(t *testing.T) {
	h, e := newTestHandler()

	m := &Medication{CodeValue: "12345", CodeDisplay: "Aspirin"}
	h.svc.CreateMedication(nil, m)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(m.ID.String())

	err := h.GetMedication(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_GetMedication_NotFound(t *testing.T) {
	h, e := newTestHandler()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(uuid.New().String())

	err := h.GetMedication(c)
	if err == nil {
		t.Error("expected error for not found")
	}
}

func TestHandler_GetMedication_InvalidID(t *testing.T) {
	h, e := newTestHandler()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("not-a-uuid")

	err := h.GetMedication(c)
	if err == nil {
		t.Error("expected error for invalid id")
	}
}

func TestHandler_DeleteMedication(t *testing.T) {
	h, e := newTestHandler()

	m := &Medication{CodeValue: "12345", CodeDisplay: "Aspirin"}
	h.svc.CreateMedication(nil, m)

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(m.ID.String())

	err := h.DeleteMedication(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

func TestHandler_CreateMedicationRequest(t *testing.T) {
	h, e := newTestHandler()

	patientID := uuid.New()
	medID := uuid.New()
	requesterID := uuid.New()
	body := `{"patient_id":"` + patientID.String() + `","medication_id":"` + medID.String() + `","requester_id":"` + requesterID.String() + `"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/medication-requests", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreateMedicationRequest(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestHandler_CreateMedicationRequest_BadRequest(t *testing.T) {
	h, e := newTestHandler()

	body := `{"medication_id":"` + uuid.New().String() + `"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/medication-requests", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreateMedicationRequest(c)
	if err == nil {
		t.Error("expected error for missing patient_id")
	}
}

func TestHandler_CreateMedicationAdministration(t *testing.T) {
	h, e := newTestHandler()

	body := `{"patient_id":"` + uuid.New().String() + `","medication_id":"` + uuid.New().String() + `"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/medication-administrations", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreateMedicationAdministration(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestHandler_CreateMedicationDispense(t *testing.T) {
	h, e := newTestHandler()

	body := `{"patient_id":"` + uuid.New().String() + `","medication_id":"` + uuid.New().String() + `"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/medication-dispenses", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreateMedicationDispense(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestHandler_CreateMedicationStatement(t *testing.T) {
	h, e := newTestHandler()

	body := `{"patient_id":"` + uuid.New().String() + `"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/medication-statements", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreateMedicationStatement(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestHandler_RegisterRoutes(t *testing.T) {
	h, e := newTestHandler()
	api := e.Group("/api/v1")
	fhir := e.Group("/fhir")

	h.RegisterRoutes(api, fhir)

	routes := e.Routes()
	if len(routes) == 0 {
		t.Error("expected routes to be registered")
	}

	routePaths := make(map[string]bool)
	for _, r := range routes {
		routePaths[r.Method+":"+r.Path] = true
	}

	expected := []string{
		"POST:/api/v1/medications",
		"GET:/api/v1/medications/:id",
		"POST:/api/v1/medication-requests",
		"GET:/fhir/Medication",
		"GET:/fhir/MedicationRequest",
	}
	for _, path := range expected {
		if !routePaths[path] {
			t.Errorf("missing expected route: %s", path)
		}
	}
}
