package identity

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

func TestHandler_CreatePatient(t *testing.T) {
	h, e := newTestHandler()

	body := `{"first_name":"John","last_name":"Doe","mrn":"MRN001"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/patients", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreatePatient(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}

	var p Patient
	json.Unmarshal(rec.Body.Bytes(), &p)
	if p.FirstName != "John" {
		t.Errorf("expected John, got %s", p.FirstName)
	}
}

func TestHandler_CreatePatient_BadRequest(t *testing.T) {
	h, e := newTestHandler()

	body := `{"last_name":"Doe"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/patients", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreatePatient(c)
	if err == nil {
		t.Error("expected error for missing fields")
	}
}

func TestHandler_GetPatient(t *testing.T) {
	h, e := newTestHandler()

	p := &Patient{FirstName: "Jane", LastName: "Smith", MRN: "MRN002"}
	h.svc.CreatePatient(nil, p)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(p.ID.String())

	err := h.GetPatient(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_GetPatient_NotFound(t *testing.T) {
	h, e := newTestHandler()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(uuid.New().String())

	err := h.GetPatient(c)
	if err == nil {
		t.Error("expected error for not found")
	}
}

func TestHandler_DeletePatient(t *testing.T) {
	h, e := newTestHandler()

	p := &Patient{FirstName: "Delete", LastName: "Me", MRN: "MRN-DEL"}
	h.svc.CreatePatient(nil, p)

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(p.ID.String())

	err := h.DeletePatient(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

func TestHandler_ListPatients(t *testing.T) {
	h, e := newTestHandler()

	h.svc.CreatePatient(nil, &Patient{FirstName: "P1", LastName: "L1", MRN: "M1"})
	h.svc.CreatePatient(nil, &Patient{FirstName: "P2", LastName: "L2", MRN: "M2"})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/patients", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.ListPatients(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_CreatePractitioner(t *testing.T) {
	h, e := newTestHandler()

	body := `{"first_name":"Dr. Sarah","last_name":"Johnson"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/practitioners", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreatePractitioner(c)
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
	routePaths := make(map[string]bool)
	for _, r := range routes {
		routePaths[r.Method+":"+r.Path] = true
	}

	expected := []string{
		"POST:/api/v1/patients",
		"GET:/api/v1/patients",
		"GET:/api/v1/patients/:id",
		"POST:/api/v1/practitioners",
		"GET:/fhir/Patient",
		"GET:/fhir/Patient/:id",
		"GET:/fhir/Practitioner",
	}
	for _, path := range expected {
		if !routePaths[path] {
			t.Errorf("missing expected route: %s", path)
		}
	}
}
