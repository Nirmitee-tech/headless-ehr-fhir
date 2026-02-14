package careplan

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

func TestHandler_CreateCarePlan(t *testing.T) {
	h, e := newTestHandler()
	body := `{"patient_id":"` + uuid.New().String() + `","intent":"plan"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if err := h.CreateCarePlan(c); err != nil { t.Fatalf("unexpected error: %v", err) }
	if rec.Code != http.StatusCreated { t.Errorf("expected 201, got %d", rec.Code) }
}

func TestHandler_GetCarePlan(t *testing.T) {
	h, e := newTestHandler()
	cp := &CarePlan{PatientID: uuid.New(), Intent: "plan"}
	h.svc.CreateCarePlan(nil, cp)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id"); c.SetParamValues(cp.ID.String())
	if err := h.GetCarePlan(c); err != nil { t.Fatalf("unexpected error: %v", err) }
	if rec.Code != http.StatusOK { t.Errorf("expected 200, got %d", rec.Code) }
}

func TestHandler_GetCarePlan_NotFound(t *testing.T) {
	h, e := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id"); c.SetParamValues(uuid.New().String())
	if err := h.GetCarePlan(c); err == nil { t.Error("expected error") }
}

func TestHandler_ListCarePlans(t *testing.T) {
	h, e := newTestHandler()
	h.svc.CreateCarePlan(nil, &CarePlan{PatientID: uuid.New(), Intent: "plan"})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if err := h.ListCarePlans(c); err != nil { t.Fatalf("unexpected error: %v", err) }
	if rec.Code != http.StatusOK { t.Errorf("expected 200, got %d", rec.Code) }
}

func TestHandler_DeleteCarePlan(t *testing.T) {
	h, e := newTestHandler()
	cp := &CarePlan{PatientID: uuid.New(), Intent: "plan"}
	h.svc.CreateCarePlan(nil, cp)
	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id"); c.SetParamValues(cp.ID.String())
	if err := h.DeleteCarePlan(c); err != nil { t.Fatalf("unexpected error: %v", err) }
	if rec.Code != http.StatusNoContent { t.Errorf("expected 204, got %d", rec.Code) }
}

func TestHandler_CreateGoal(t *testing.T) {
	h, e := newTestHandler()
	body := `{"patient_id":"` + uuid.New().String() + `","description":"Reduce A1C"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if err := h.CreateGoal(c); err != nil { t.Fatalf("unexpected error: %v", err) }
	if rec.Code != http.StatusCreated { t.Errorf("expected 201, got %d", rec.Code) }
}

func TestHandler_GetGoal(t *testing.T) {
	h, e := newTestHandler()
	g := &Goal{PatientID: uuid.New(), Description: "test"}
	h.svc.CreateGoal(nil, g)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id"); c.SetParamValues(g.ID.String())
	if err := h.GetGoal(c); err != nil { t.Fatalf("unexpected error: %v", err) }
	if rec.Code != http.StatusOK { t.Errorf("expected 200, got %d", rec.Code) }
}

func TestHandler_SearchCarePlansFHIR(t *testing.T) {
	h, e := newTestHandler()
	h.svc.CreateCarePlan(nil, &CarePlan{PatientID: uuid.New(), Intent: "plan"})
	req := httptest.NewRequest(http.MethodGet, "/fhir/CarePlan", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if err := h.SearchCarePlansFHIR(c); err != nil { t.Fatalf("unexpected error: %v", err) }
	if rec.Code != http.StatusOK { t.Errorf("expected 200, got %d", rec.Code) }
	var bundle map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &bundle)
	if bundle["resourceType"] != "Bundle" { t.Errorf("expected Bundle") }
}

func TestHandler_GetCarePlanFHIR(t *testing.T) {
	h, e := newTestHandler()
	cp := &CarePlan{PatientID: uuid.New(), Intent: "plan"}
	h.svc.CreateCarePlan(nil, cp)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id"); c.SetParamValues(cp.FHIRID)
	if err := h.GetCarePlanFHIR(c); err != nil { t.Fatalf("unexpected error: %v", err) }
	if rec.Code != http.StatusOK { t.Errorf("expected 200, got %d", rec.Code) }
}

func TestHandler_GetCarePlanFHIR_NotFound(t *testing.T) {
	h, e := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id"); c.SetParamValues("nonexistent")
	if err := h.GetCarePlanFHIR(c); err != nil { t.Fatalf("unexpected error: %v", err) }
	if rec.Code != http.StatusNotFound { t.Errorf("expected 404, got %d", rec.Code) }
}

func TestHandler_RegisterRoutes(t *testing.T) {
	h, e := newTestHandler()
	api := e.Group("/api/v1")
	fhir := e.Group("/fhir")
	h.RegisterRoutes(api, fhir)
	routes := e.Routes()
	if len(routes) == 0 { t.Error("expected routes") }
	routePaths := make(map[string]bool)
	for _, r := range routes { routePaths[r.Method+":"+r.Path] = true }
	expected := []string{"POST:/api/v1/care-plans", "GET:/api/v1/care-plans", "GET:/api/v1/goals", "GET:/fhir/CarePlan", "GET:/fhir/Goal", "POST:/fhir/CarePlan", "POST:/fhir/Goal"}
	for _, path := range expected {
		if !routePaths[path] { t.Errorf("missing route: %s", path) }
	}
}
