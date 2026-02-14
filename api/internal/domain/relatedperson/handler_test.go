package relatedperson

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

// ── REST Handlers ──

func TestHandler_CreateRelatedPerson(t *testing.T) {
	h, e := newTestHandler()
	body := `{"patient_id":"` + uuid.New().String() + `","relationship_code":"WIFE","relationship_display":"Wife"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := h.CreateRelatedPerson(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestHandler_CreateRelatedPerson_BadRequest(t *testing.T) {
	h, e := newTestHandler()
	body := `{"relationship_display":"Wife"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := h.CreateRelatedPerson(c)
	if err == nil {
		t.Error("expected error for missing required fields")
	}
}

func TestHandler_GetRelatedPerson(t *testing.T) {
	h, e := newTestHandler()
	rp := &RelatedPerson{PatientID: uuid.New(), RelationshipCode: "WIFE", RelationshipDisplay: "Wife"}
	h.svc.CreateRelatedPerson(nil, rp)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(rp.ID.String())
	err := h.GetRelatedPerson(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_GetRelatedPerson_NotFound(t *testing.T) {
	h, e := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(uuid.New().String())
	err := h.GetRelatedPerson(c)
	if err == nil {
		t.Error("expected error for not found")
	}
}

func TestHandler_GetRelatedPerson_InvalidID(t *testing.T) {
	h, e := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("not-a-uuid")
	err := h.GetRelatedPerson(c)
	if err == nil {
		t.Error("expected error for invalid id")
	}
}

func TestHandler_ListRelatedPersons(t *testing.T) {
	h, e := newTestHandler()
	h.svc.CreateRelatedPerson(nil, &RelatedPerson{PatientID: uuid.New(), RelationshipCode: "WIFE", RelationshipDisplay: "Wife"})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := h.ListRelatedPersons(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_UpdateRelatedPerson(t *testing.T) {
	h, e := newTestHandler()
	rp := &RelatedPerson{PatientID: uuid.New(), RelationshipCode: "WIFE", RelationshipDisplay: "Wife"}
	h.svc.CreateRelatedPerson(nil, rp)

	body := `{"active":true,"relationship_code":"WIFE","relationship_display":"Wife"}`
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(rp.ID.String())
	err := h.UpdateRelatedPerson(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_DeleteRelatedPerson(t *testing.T) {
	h, e := newTestHandler()
	rp := &RelatedPerson{PatientID: uuid.New(), RelationshipCode: "WIFE", RelationshipDisplay: "Wife"}
	h.svc.CreateRelatedPerson(nil, rp)

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(rp.ID.String())
	err := h.DeleteRelatedPerson(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

// ── FHIR Endpoints ──

func TestHandler_SearchFHIR(t *testing.T) {
	h, e := newTestHandler()
	h.svc.CreateRelatedPerson(nil, &RelatedPerson{PatientID: uuid.New(), RelationshipCode: "WIFE", RelationshipDisplay: "Wife"})

	req := httptest.NewRequest(http.MethodGet, "/fhir/RelatedPerson", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := h.SearchFHIR(c)
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

func TestHandler_GetFHIR(t *testing.T) {
	h, e := newTestHandler()
	rp := &RelatedPerson{PatientID: uuid.New(), RelationshipCode: "WIFE", RelationshipDisplay: "Wife"}
	h.svc.CreateRelatedPerson(nil, rp)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(rp.FHIRID)
	err := h.GetFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_GetFHIR_NotFound(t *testing.T) {
	h, e := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("nonexistent")
	err := h.GetFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestHandler_CreateFHIR(t *testing.T) {
	h, e := newTestHandler()
	body := `{"patient_id":"` + uuid.New().String() + `","relationship_code":"WIFE","relationship_display":"Wife"}`
	req := httptest.NewRequest(http.MethodPost, "/fhir/RelatedPerson", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := h.CreateFHIR(c)
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

func TestHandler_UpdateFHIR(t *testing.T) {
	h, e := newTestHandler()
	rp := &RelatedPerson{PatientID: uuid.New(), RelationshipCode: "WIFE", RelationshipDisplay: "Wife"}
	h.svc.CreateRelatedPerson(nil, rp)

	body := `{"active":true,"relationship_code":"WIFE","relationship_display":"Wife"}`
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(rp.FHIRID)
	err := h.UpdateFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_DeleteFHIR(t *testing.T) {
	h, e := newTestHandler()
	rp := &RelatedPerson{PatientID: uuid.New(), RelationshipCode: "WIFE", RelationshipDisplay: "Wife"}
	h.svc.CreateRelatedPerson(nil, rp)

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(rp.FHIRID)
	err := h.DeleteFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

func TestHandler_VreadFHIR(t *testing.T) {
	h, e := newTestHandler()
	rp := &RelatedPerson{PatientID: uuid.New(), RelationshipCode: "WIFE", RelationshipDisplay: "Wife"}
	h.svc.CreateRelatedPerson(nil, rp)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id", "vid")
	c.SetParamValues(rp.FHIRID, "1")
	err := h.VreadFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_HistoryFHIR(t *testing.T) {
	h, e := newTestHandler()
	rp := &RelatedPerson{PatientID: uuid.New(), RelationshipCode: "WIFE", RelationshipDisplay: "Wife"}
	h.svc.CreateRelatedPerson(nil, rp)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(rp.FHIRID)
	err := h.HistoryFHIR(c)
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
		"POST:/api/v1/related-persons",
		"GET:/api/v1/related-persons",
		"GET:/api/v1/related-persons/:id",
		"PUT:/api/v1/related-persons/:id",
		"DELETE:/api/v1/related-persons/:id",
		"GET:/fhir/RelatedPerson",
		"GET:/fhir/RelatedPerson/:id",
		"POST:/fhir/RelatedPerson",
		"PUT:/fhir/RelatedPerson/:id",
		"DELETE:/fhir/RelatedPerson/:id",
	}
	for _, path := range expected {
		if !routePaths[path] {
			t.Errorf("missing expected route: %s", path)
		}
	}
}
