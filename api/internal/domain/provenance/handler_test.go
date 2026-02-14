package provenance

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
	h := NewHandler(svc)
	e := echo.New()
	return h, e
}

// ── REST Handlers ──

func TestHandler_CreateProvenance(t *testing.T) {
	h, e := newTestHandler()
	body := `{"target_type":"Patient","target_id":"pat-123","recorded":"2024-06-01T10:30:00Z"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := h.CreateProvenance(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestHandler_CreateProvenance_BadRequest(t *testing.T) {
	h, e := newTestHandler()
	body := `{"target_id":"pat-123"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := h.CreateProvenance(c)
	if err == nil {
		t.Error("expected error for missing required fields")
	}
}

func TestHandler_GetProvenance(t *testing.T) {
	h, e := newTestHandler()
	p := &Provenance{TargetType: "Patient", TargetID: "pat-123", Recorded: time.Now()}
	h.svc.CreateProvenance(nil, p)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(p.ID.String())
	err := h.GetProvenance(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_GetProvenance_NotFound(t *testing.T) {
	h, e := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(uuid.New().String())
	err := h.GetProvenance(c)
	if err == nil {
		t.Error("expected error for not found")
	}
}

func TestHandler_GetProvenance_InvalidID(t *testing.T) {
	h, e := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("not-a-uuid")
	err := h.GetProvenance(c)
	if err == nil {
		t.Error("expected error for invalid id")
	}
}

func TestHandler_ListProvenances(t *testing.T) {
	h, e := newTestHandler()
	h.svc.CreateProvenance(nil, &Provenance{TargetType: "Patient", TargetID: "pat-123", Recorded: time.Now()})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := h.ListProvenances(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_UpdateProvenance(t *testing.T) {
	h, e := newTestHandler()
	p := &Provenance{TargetType: "Patient", TargetID: "pat-123", Recorded: time.Now()}
	h.svc.CreateProvenance(nil, p)

	body := `{"target_type":"Patient","target_id":"pat-123"}`
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(p.ID.String())
	err := h.UpdateProvenance(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_DeleteProvenance(t *testing.T) {
	h, e := newTestHandler()
	p := &Provenance{TargetType: "Patient", TargetID: "pat-123", Recorded: time.Now()}
	h.svc.CreateProvenance(nil, p)

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(p.ID.String())
	err := h.DeleteProvenance(c)
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
	h.svc.CreateProvenance(nil, &Provenance{TargetType: "Patient", TargetID: "pat-123", Recorded: time.Now()})

	req := httptest.NewRequest(http.MethodGet, "/fhir/Provenance", nil)
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
	p := &Provenance{TargetType: "Patient", TargetID: "pat-123", Recorded: time.Now()}
	h.svc.CreateProvenance(nil, p)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(p.FHIRID)
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
	body := `{"target_type":"Patient","target_id":"pat-123","recorded":"2024-06-01T10:30:00Z"}`
	req := httptest.NewRequest(http.MethodPost, "/fhir/Provenance", strings.NewReader(body))
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
	p := &Provenance{TargetType: "Patient", TargetID: "pat-123", Recorded: time.Now()}
	h.svc.CreateProvenance(nil, p)

	body := `{"target_type":"Patient","target_id":"pat-123"}`
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(p.FHIRID)
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
	p := &Provenance{TargetType: "Patient", TargetID: "pat-123", Recorded: time.Now()}
	h.svc.CreateProvenance(nil, p)

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(p.FHIRID)
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
	p := &Provenance{TargetType: "Patient", TargetID: "pat-123", Recorded: time.Now()}
	h.svc.CreateProvenance(nil, p)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id", "vid")
	c.SetParamValues(p.FHIRID, "1")
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
	p := &Provenance{TargetType: "Patient", TargetID: "pat-123", Recorded: time.Now()}
	h.svc.CreateProvenance(nil, p)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(p.FHIRID)
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
		"POST:/api/v1/provenances",
		"GET:/api/v1/provenances",
		"GET:/api/v1/provenances/:id",
		"PUT:/api/v1/provenances/:id",
		"DELETE:/api/v1/provenances/:id",
		"GET:/fhir/Provenance",
		"GET:/fhir/Provenance/:id",
		"POST:/fhir/Provenance",
		"PUT:/fhir/Provenance/:id",
		"DELETE:/fhir/Provenance/:id",
	}
	for _, path := range expected {
		if !routePaths[path] {
			t.Errorf("missing expected route: %s", path)
		}
	}
}
