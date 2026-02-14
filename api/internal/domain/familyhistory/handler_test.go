package familyhistory

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

func TestHandler_CreateFamilyMemberHistory(t *testing.T) {
	h, e := newTestHandler()
	body := `{"patient_id":"` + uuid.New().String() + `","relationship_code":"FTH","relationship_display":"Father"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := h.CreateFamilyMemberHistory(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestHandler_CreateFamilyMemberHistory_BadRequest(t *testing.T) {
	h, e := newTestHandler()
	body := `{"relationship_display":"Father"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := h.CreateFamilyMemberHistory(c)
	if err == nil {
		t.Error("expected error for missing required fields")
	}
}

func TestHandler_GetFamilyMemberHistory(t *testing.T) {
	h, e := newTestHandler()
	f := &FamilyMemberHistory{PatientID: uuid.New(), RelationshipCode: "FTH", RelationshipDisplay: "Father"}
	h.svc.CreateFamilyMemberHistory(nil, f)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(f.ID.String())
	err := h.GetFamilyMemberHistory(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_GetFamilyMemberHistory_NotFound(t *testing.T) {
	h, e := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(uuid.New().String())
	err := h.GetFamilyMemberHistory(c)
	if err == nil {
		t.Error("expected error for not found")
	}
}

func TestHandler_GetFamilyMemberHistory_InvalidID(t *testing.T) {
	h, e := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("not-a-uuid")
	err := h.GetFamilyMemberHistory(c)
	if err == nil {
		t.Error("expected error for invalid id")
	}
}

func TestHandler_ListFamilyMemberHistories(t *testing.T) {
	h, e := newTestHandler()
	h.svc.CreateFamilyMemberHistory(nil, &FamilyMemberHistory{PatientID: uuid.New(), RelationshipCode: "FTH", RelationshipDisplay: "Father"})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := h.ListFamilyMemberHistories(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_UpdateFamilyMemberHistory(t *testing.T) {
	h, e := newTestHandler()
	f := &FamilyMemberHistory{PatientID: uuid.New(), RelationshipCode: "FTH", RelationshipDisplay: "Father"}
	h.svc.CreateFamilyMemberHistory(nil, f)

	body := `{"status":"partial","relationship_code":"FTH","relationship_display":"Father"}`
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(f.ID.String())
	err := h.UpdateFamilyMemberHistory(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_DeleteFamilyMemberHistory(t *testing.T) {
	h, e := newTestHandler()
	f := &FamilyMemberHistory{PatientID: uuid.New(), RelationshipCode: "FTH", RelationshipDisplay: "Father"}
	h.svc.CreateFamilyMemberHistory(nil, f)

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(f.ID.String())
	err := h.DeleteFamilyMemberHistory(c)
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
	h.svc.CreateFamilyMemberHistory(nil, &FamilyMemberHistory{PatientID: uuid.New(), RelationshipCode: "FTH", RelationshipDisplay: "Father"})

	req := httptest.NewRequest(http.MethodGet, "/fhir/FamilyMemberHistory", nil)
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
	f := &FamilyMemberHistory{PatientID: uuid.New(), RelationshipCode: "FTH", RelationshipDisplay: "Father"}
	h.svc.CreateFamilyMemberHistory(nil, f)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(f.FHIRID)
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
	body := `{"patient_id":"` + uuid.New().String() + `","relationship_code":"FTH","relationship_display":"Father"}`
	req := httptest.NewRequest(http.MethodPost, "/fhir/FamilyMemberHistory", strings.NewReader(body))
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
	f := &FamilyMemberHistory{PatientID: uuid.New(), RelationshipCode: "FTH", RelationshipDisplay: "Father"}
	h.svc.CreateFamilyMemberHistory(nil, f)

	body := `{"status":"partial","relationship_code":"FTH","relationship_display":"Father"}`
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(f.FHIRID)
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
	f := &FamilyMemberHistory{PatientID: uuid.New(), RelationshipCode: "FTH", RelationshipDisplay: "Father"}
	h.svc.CreateFamilyMemberHistory(nil, f)

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(f.FHIRID)
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
	f := &FamilyMemberHistory{PatientID: uuid.New(), RelationshipCode: "FTH", RelationshipDisplay: "Father"}
	h.svc.CreateFamilyMemberHistory(nil, f)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id", "vid")
	c.SetParamValues(f.FHIRID, "1")
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
	f := &FamilyMemberHistory{PatientID: uuid.New(), RelationshipCode: "FTH", RelationshipDisplay: "Father"}
	h.svc.CreateFamilyMemberHistory(nil, f)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(f.FHIRID)
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
		"POST:/api/v1/family-member-histories",
		"GET:/api/v1/family-member-histories",
		"GET:/api/v1/family-member-histories/:id",
		"PUT:/api/v1/family-member-histories/:id",
		"DELETE:/api/v1/family-member-histories/:id",
		"GET:/fhir/FamilyMemberHistory",
		"GET:/fhir/FamilyMemberHistory/:id",
		"POST:/fhir/FamilyMemberHistory",
		"PUT:/fhir/FamilyMemberHistory/:id",
		"DELETE:/fhir/FamilyMemberHistory/:id",
	}
	for _, path := range expected {
		if !routePaths[path] {
			t.Errorf("missing expected route: %s", path)
		}
	}
}
