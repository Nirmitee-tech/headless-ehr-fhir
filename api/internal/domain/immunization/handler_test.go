package immunization

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func newTestHandler() (*Handler, *echo.Echo) {
	svc := newTestService()
	h := NewHandler(svc, nil)
	e := echo.New()
	return h, e
}

// ── Immunization Handlers ──

func TestHandler_CreateImmunization(t *testing.T) {
	h, e := newTestHandler()
	body := `{"patient_id":"` + uuid.New().String() + `","vaccine_code":"08","vaccine_display":"Hep B"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := h.CreateImmunization(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestHandler_CreateImmunization_BadRequest(t *testing.T) {
	h, e := newTestHandler()
	body := `{"vaccine_display":"Hep B"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := h.CreateImmunization(c)
	if err == nil {
		t.Error("expected error for missing required fields")
	}
}

func TestHandler_GetImmunization(t *testing.T) {
	h, e := newTestHandler()
	im := &Immunization{PatientID: uuid.New(), VaccineCode: "08", VaccineDisplay: "Hep B"}
	h.svc.CreateImmunization(nil, im)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(im.ID.String())
	err := h.GetImmunization(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_GetImmunization_NotFound(t *testing.T) {
	h, e := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(uuid.New().String())
	err := h.GetImmunization(c)
	if err == nil {
		t.Error("expected error for not found")
	}
}

func TestHandler_GetImmunization_InvalidID(t *testing.T) {
	h, e := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("not-a-uuid")
	err := h.GetImmunization(c)
	if err == nil {
		t.Error("expected error for invalid id")
	}
}

func TestHandler_ListImmunizations(t *testing.T) {
	h, e := newTestHandler()
	h.svc.CreateImmunization(nil, &Immunization{PatientID: uuid.New(), VaccineCode: "08", VaccineDisplay: "Hep B"})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := h.ListImmunizations(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_UpdateImmunization(t *testing.T) {
	h, e := newTestHandler()
	im := &Immunization{PatientID: uuid.New(), VaccineCode: "08", VaccineDisplay: "Hep B"}
	h.svc.CreateImmunization(nil, im)

	body := `{"status":"completed","vaccine_code":"08","vaccine_display":"Hep B"}`
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(im.ID.String())
	err := h.UpdateImmunization(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_DeleteImmunization(t *testing.T) {
	h, e := newTestHandler()
	im := &Immunization{PatientID: uuid.New(), VaccineCode: "08", VaccineDisplay: "Hep B"}
	h.svc.CreateImmunization(nil, im)

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(im.ID.String())
	err := h.DeleteImmunization(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

// ── Recommendation Handlers ──

func TestHandler_CreateRecommendation(t *testing.T) {
	h, e := newTestHandler()
	body := `{"patient_id":"` + uuid.New().String() + `","vaccine_code":"08","vaccine_display":"Hep B","forecast_status":"due","date":"2024-06-01T00:00:00Z"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := h.CreateRecommendation(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestHandler_GetRecommendation(t *testing.T) {
	h, e := newTestHandler()
	r := &ImmunizationRecommendation{PatientID: uuid.New(), VaccineCode: "08", VaccineDisplay: "Hep B", ForecastStatus: "due", Date: time.Now()}
	h.svc.CreateRecommendation(nil, r)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(r.ID.String())
	err := h.GetRecommendation(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_DeleteRecommendation(t *testing.T) {
	h, e := newTestHandler()
	r := &ImmunizationRecommendation{PatientID: uuid.New(), VaccineCode: "08", VaccineDisplay: "Hep B", ForecastStatus: "due", Date: time.Now()}
	h.svc.CreateRecommendation(nil, r)

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(r.ID.String())
	err := h.DeleteRecommendation(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

// ── FHIR Endpoints ──

func TestHandler_SearchImmunizationsFHIR(t *testing.T) {
	h, e := newTestHandler()
	h.svc.CreateImmunization(nil, &Immunization{PatientID: uuid.New(), VaccineCode: "08", VaccineDisplay: "Hep B"})

	req := httptest.NewRequest(http.MethodGet, "/fhir/Immunization", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := h.SearchImmunizationsFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	var bundle map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &bundle)
	if bundle["resourceType"] != "Bundle" {
		t.Errorf("expected Bundle, got %v", bundle["resourceType"])
	}
}

func TestHandler_GetImmunizationFHIR(t *testing.T) {
	h, e := newTestHandler()
	im := &Immunization{PatientID: uuid.New(), VaccineCode: "08", VaccineDisplay: "Hep B"}
	h.svc.CreateImmunization(nil, im)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(im.FHIRID)
	err := h.GetImmunizationFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_GetImmunizationFHIR_NotFound(t *testing.T) {
	h, e := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("nonexistent")
	err := h.GetImmunizationFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestHandler_CreateImmunizationFHIR(t *testing.T) {
	h, e := newTestHandler()
	body := `{"patient_id":"` + uuid.New().String() + `","vaccine_code":"08","vaccine_display":"Hep B"}`
	req := httptest.NewRequest(http.MethodPost, "/fhir/Immunization", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := h.CreateImmunizationFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
	loc := rec.Header().Get("Location")
	if loc == "" {
		t.Error("expected Location header")
	}
}

func TestHandler_VreadImmunizationFHIR(t *testing.T) {
	h, e := newTestHandler()
	im := &Immunization{PatientID: uuid.New(), VaccineCode: "08", VaccineDisplay: "Hep B"}
	h.svc.CreateImmunization(nil, im)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id", "vid")
	c.SetParamValues(im.FHIRID, "1")
	err := h.VreadImmunizationFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_HistoryImmunizationFHIR(t *testing.T) {
	h, e := newTestHandler()
	im := &Immunization{PatientID: uuid.New(), VaccineCode: "08", VaccineDisplay: "Hep B"}
	h.svc.CreateImmunization(nil, im)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(im.FHIRID)
	err := h.HistoryImmunizationFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	var bundle map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &bundle)
	if bundle["type"] != "history" {
		t.Errorf("expected history bundle type, got %v", bundle["type"])
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
		"POST:/api/v1/immunizations",
		"GET:/api/v1/immunizations/:id",
		"GET:/api/v1/immunizations",
		"PUT:/api/v1/immunizations/:id",
		"DELETE:/api/v1/immunizations/:id",
		"GET:/fhir/Immunization",
		"GET:/fhir/Immunization/:id",
		"POST:/fhir/Immunization",
		"PUT:/fhir/Immunization/:id",
		"DELETE:/fhir/Immunization/:id",
		"PATCH:/fhir/Immunization/:id",
	}
	for _, path := range expected {
		if !routePaths[path] {
			t.Errorf("missing expected route: %s", path)
		}
	}
}
