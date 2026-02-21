package careteam

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
	h := NewHandler(svc, nil)
	e := echo.New()
	return h, e
}

func TestGetCareTeamFHIR_Success(t *testing.T) {
	h, e := newTestHandler()
	ct := &CareTeam{PatientID: uuid.New(), Status: "active"}
	h.svc.CreateCareTeam(nil, ct)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(ct.FHIRID)
	if err := h.GetCareTeamFHIR(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	var result map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &result)
	if result["resourceType"] != "CareTeam" {
		t.Errorf("expected resourceType 'CareTeam', got %v", result["resourceType"])
	}
}

func TestGetCareTeamFHIR_NotFound(t *testing.T) {
	h, e := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("nonexistent")
	if err := h.GetCareTeamFHIR(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestCreateCareTeamFHIR_Success(t *testing.T) {
	h, e := newTestHandler()
	body := `{"patient_id":"` + uuid.New().String() + `","status":"active"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if err := h.CreateCareTeamFHIR(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
	location := rec.Header().Get("Location")
	if location == "" {
		t.Error("expected Location header to be set")
	}
}

func TestSearchCareTeamsFHIR_Success(t *testing.T) {
	h, e := newTestHandler()
	h.svc.CreateCareTeam(nil, &CareTeam{PatientID: uuid.New(), Status: "active"})
	req := httptest.NewRequest(http.MethodGet, "/fhir/CareTeam", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if err := h.SearchCareTeamsFHIR(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	var bundle map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &bundle)
	if bundle["resourceType"] != "Bundle" {
		t.Errorf("expected Bundle")
	}
}

func TestDeleteCareTeamFHIR_Success(t *testing.T) {
	h, e := newTestHandler()
	ct := &CareTeam{PatientID: uuid.New(), Status: "active"}
	h.svc.CreateCareTeam(nil, ct)
	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(ct.FHIRID)
	if err := h.DeleteCareTeamFHIR(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

func TestDeleteCareTeamFHIR_NotFound(t *testing.T) {
	h, e := newTestHandler()
	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("nonexistent")
	if err := h.DeleteCareTeamFHIR(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestHandler_CreateCareTeam(t *testing.T) {
	h, e := newTestHandler()
	body := `{"patient_id":"` + uuid.New().String() + `","status":"active"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if err := h.CreateCareTeam(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestHandler_GetCareTeam(t *testing.T) {
	h, e := newTestHandler()
	ct := &CareTeam{PatientID: uuid.New(), Status: "active"}
	h.svc.CreateCareTeam(nil, ct)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(ct.ID.String())
	if err := h.GetCareTeam(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_GetCareTeam_NotFound(t *testing.T) {
	h, e := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(uuid.New().String())
	if err := h.GetCareTeam(c); err == nil {
		t.Error("expected error")
	}
}

func TestHandler_DeleteCareTeam(t *testing.T) {
	h, e := newTestHandler()
	ct := &CareTeam{PatientID: uuid.New(), Status: "active"}
	h.svc.CreateCareTeam(nil, ct)
	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(ct.ID.String())
	if err := h.DeleteCareTeam(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

func TestHandler_ListCareTeams(t *testing.T) {
	h, e := newTestHandler()
	h.svc.CreateCareTeam(nil, &CareTeam{PatientID: uuid.New(), Status: "active"})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if err := h.ListCareTeams(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_RegisterRoutes(t *testing.T) {
	h, e := newTestHandler()
	api := e.Group("/api/v1")
	fhirGroup := e.Group("/fhir")
	h.RegisterRoutes(api, fhirGroup)
	routes := e.Routes()
	if len(routes) == 0 {
		t.Error("expected routes")
	}
	routePaths := make(map[string]bool)
	for _, r := range routes {
		routePaths[r.Method+":"+r.Path] = true
	}
	expected := []string{
		"POST:/api/v1/care-teams",
		"GET:/api/v1/care-teams",
		"GET:/api/v1/care-teams/:id",
		"PUT:/api/v1/care-teams/:id",
		"DELETE:/api/v1/care-teams/:id",
		"GET:/fhir/CareTeam",
		"GET:/fhir/CareTeam/:id",
		"POST:/fhir/CareTeam",
		"PUT:/fhir/CareTeam/:id",
		"DELETE:/fhir/CareTeam/:id",
	}
	for _, path := range expected {
		if !routePaths[path] {
			t.Errorf("missing route: %s", path)
		}
	}
}
