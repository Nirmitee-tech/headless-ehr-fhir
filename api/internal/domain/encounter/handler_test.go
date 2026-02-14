package encounter

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

func TestHandler_CreateEncounter(t *testing.T) {
	h, e := newTestHandler()

	patientID := uuid.New()
	body := `{"patient_id":"` + patientID.String() + `","class_code":"AMB"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/encounters", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreateEncounter(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}

	var enc Encounter
	json.Unmarshal(rec.Body.Bytes(), &enc)
	if enc.ClassCode != "AMB" {
		t.Errorf("expected AMB, got %s", enc.ClassCode)
	}
}

func TestHandler_CreateEncounter_BadRequest(t *testing.T) {
	h, e := newTestHandler()

	body := `{"class_code":"AMB"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/encounters", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreateEncounter(c)
	if err == nil {
		t.Error("expected error for missing patient_id")
	}
}

func TestHandler_GetEncounter(t *testing.T) {
	h, e := newTestHandler()

	enc := &Encounter{PatientID: uuid.New(), ClassCode: "AMB"}
	h.svc.CreateEncounter(nil, enc)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(enc.ID.String())

	err := h.GetEncounter(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_GetEncounter_NotFound(t *testing.T) {
	h, e := newTestHandler()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(uuid.New().String())

	err := h.GetEncounter(c)
	if err == nil {
		t.Error("expected error for not found")
	}
}

func TestHandler_DeleteEncounter(t *testing.T) {
	h, e := newTestHandler()

	enc := &Encounter{PatientID: uuid.New(), ClassCode: "AMB"}
	h.svc.CreateEncounter(nil, enc)

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(enc.ID.String())

	err := h.DeleteEncounter(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

func TestHandler_UpdateStatus(t *testing.T) {
	h, e := newTestHandler()

	enc := &Encounter{PatientID: uuid.New(), ClassCode: "AMB"}
	h.svc.CreateEncounter(nil, enc)

	body := `{"status":"in-progress"}`
	req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(enc.ID.String())

	err := h.UpdateEncounterStatus(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_ListEncounters(t *testing.T) {
	h, e := newTestHandler()

	h.svc.CreateEncounter(nil, &Encounter{PatientID: uuid.New(), ClassCode: "AMB"})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/encounters", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.ListEncounters(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_RegisterRoutes(t *testing.T) {
	h, e := newTestHandler()
	api := e.Group("/api/v1")
	fhir := e.Group("/fhir")

	h.RegisterRoutes(api, fhir)

	routes := e.Routes()
	routePaths := make(map[string]bool)
	for _, r := range routes {
		routePaths[r.Method+":"+r.Path] = true
	}

	expected := []string{
		"POST:/api/v1/encounters",
		"GET:/api/v1/encounters",
		"GET:/api/v1/encounters/:id",
		"GET:/fhir/Encounter",
		"GET:/fhir/Encounter/:id",
	}
	for _, path := range expected {
		if !routePaths[path] {
			t.Errorf("missing expected route: %s", path)
		}
	}
}
