package device

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

func TestGetDeviceFHIR_Success(t *testing.T) {
	h, e := newTestHandler()
	d := &Device{DeviceName: "Pulse Oximeter", Status: "active"}
	h.svc.CreateDevice(nil, d)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(d.FHIRID)
	if err := h.GetDeviceFHIR(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	var result map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &result)
	if result["resourceType"] != "Device" {
		t.Errorf("resourceType = %v, want Device", result["resourceType"])
	}
}

func TestGetDeviceFHIR_NotFound(t *testing.T) {
	h, e := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("nonexistent")
	if err := h.GetDeviceFHIR(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestCreateDeviceFHIR_Success(t *testing.T) {
	h, e := newTestHandler()
	body := `{"device_name":"New Device","status":"active"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if err := h.CreateDeviceFHIR(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
	var result map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &result)
	if result["resourceType"] != "Device" {
		t.Errorf("resourceType = %v, want Device", result["resourceType"])
	}
	// Verify Location header is set
	loc := rec.Header().Get("Location")
	if loc == "" {
		t.Error("expected Location header to be set")
	}
}

func TestCreateDeviceFHIR_InvalidBody(t *testing.T) {
	h, e := newTestHandler()
	body := `{"status":"active"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if err := h.CreateDeviceFHIR(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestSearchDevicesFHIR_Success(t *testing.T) {
	h, e := newTestHandler()
	h.svc.CreateDevice(nil, &Device{DeviceName: "Device A", Status: "active"})
	h.svc.CreateDevice(nil, &Device{DeviceName: "Device B", Status: "active"})
	req := httptest.NewRequest(http.MethodGet, "/fhir/Device", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if err := h.SearchDevicesFHIR(c); err != nil {
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
	if bundle["type"] != "searchset" {
		t.Errorf("expected searchset, got %v", bundle["type"])
	}
}

func TestDeleteDeviceFHIR_Success(t *testing.T) {
	h, e := newTestHandler()
	d := &Device{DeviceName: "Pulse Oximeter", Status: "active"}
	h.svc.CreateDevice(nil, d)
	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(d.FHIRID)
	if err := h.DeleteDeviceFHIR(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

func TestDeleteDeviceFHIR_NotFound(t *testing.T) {
	h, e := newTestHandler()
	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("nonexistent")
	if err := h.DeleteDeviceFHIR(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestHandler_CreateDevice(t *testing.T) {
	h, e := newTestHandler()
	body := `{"device_name":"Pulse Oximeter","status":"active"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if err := h.CreateDevice(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestHandler_GetDevice(t *testing.T) {
	h, e := newTestHandler()
	d := &Device{DeviceName: "Pulse Oximeter", Status: "active"}
	h.svc.CreateDevice(nil, d)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(d.ID.String())
	if err := h.GetDevice(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_GetDevice_NotFound(t *testing.T) {
	h, e := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(uuid.New().String())
	if err := h.GetDevice(c); err == nil {
		t.Error("expected error")
	}
}

func TestHandler_ListDevices(t *testing.T) {
	h, e := newTestHandler()
	h.svc.CreateDevice(nil, &Device{DeviceName: "Device A", Status: "active"})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if err := h.ListDevices(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_DeleteDevice(t *testing.T) {
	h, e := newTestHandler()
	d := &Device{DeviceName: "Pulse Oximeter", Status: "active"}
	h.svc.CreateDevice(nil, d)
	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(d.ID.String())
	if err := h.DeleteDevice(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
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
		"POST:/api/v1/devices",
		"GET:/api/v1/devices",
		"GET:/api/v1/devices/:id",
		"PUT:/api/v1/devices/:id",
		"DELETE:/api/v1/devices/:id",
		"GET:/fhir/Device",
		"GET:/fhir/Device/:id",
		"POST:/fhir/Device",
		"PUT:/fhir/Device/:id",
		"DELETE:/fhir/Device/:id",
	}
	for _, path := range expected {
		if !routePaths[path] {
			t.Errorf("missing route: %s", path)
		}
	}
}
