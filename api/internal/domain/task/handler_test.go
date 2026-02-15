package task

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

// -- REST Handler Tests --

func TestHandler_CreateTask(t *testing.T) {
	h, e := newTestHandler()
	body := `{"for_patient_id":"` + uuid.New().String() + `","intent":"order","status":"draft"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if err := h.CreateTask(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestHandler_GetTask(t *testing.T) {
	h, e := newTestHandler()
	tk := &Task{ForPatientID: uuid.New(), Intent: "order", Status: "draft"}
	h.svc.CreateTask(nil, tk)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(tk.ID.String())
	if err := h.GetTask(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_GetTask_NotFound(t *testing.T) {
	h, e := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(uuid.New().String())
	if err := h.GetTask(c); err == nil {
		t.Error("expected error")
	}
}

func TestHandler_ListTasks(t *testing.T) {
	h, e := newTestHandler()
	h.svc.CreateTask(nil, &Task{ForPatientID: uuid.New(), Intent: "order", Status: "draft"})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if err := h.ListTasks(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_UpdateTask(t *testing.T) {
	h, e := newTestHandler()
	tk := &Task{ForPatientID: uuid.New(), Intent: "order", Status: "draft"}
	h.svc.CreateTask(nil, tk)

	body := `{"status":"requested","intent":"order"}`
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(tk.ID.String())
	if err := h.UpdateTask(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_DeleteTask(t *testing.T) {
	h, e := newTestHandler()
	tk := &Task{ForPatientID: uuid.New(), Intent: "order", Status: "draft"}
	h.svc.CreateTask(nil, tk)

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(tk.ID.String())
	if err := h.DeleteTask(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

// -- FHIR Handler Tests --

func TestGetTaskFHIR_Success(t *testing.T) {
	h, e := newTestHandler()
	tk := &Task{ForPatientID: uuid.New(), Intent: "order", Status: "draft"}
	h.svc.CreateTask(nil, tk)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(tk.FHIRID)
	if err := h.GetTaskFHIR(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var result map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &result)
	if result["resourceType"] != "Task" {
		t.Errorf("expected resourceType 'Task', got %v", result["resourceType"])
	}
}

func TestGetTaskFHIR_NotFound(t *testing.T) {
	h, e := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("nonexistent")
	if err := h.GetTaskFHIR(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestCreateTaskFHIR_Success(t *testing.T) {
	h, e := newTestHandler()
	body := `{"for_patient_id":"` + uuid.New().String() + `","intent":"order","status":"requested"}`
	req := httptest.NewRequest(http.MethodPost, "/fhir/Task", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if err := h.CreateTaskFHIR(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}

	var result map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &result)
	if result["resourceType"] != "Task" {
		t.Errorf("expected resourceType 'Task', got %v", result["resourceType"])
	}

	// Check Location header
	location := rec.Header().Get("Location")
	if location == "" {
		t.Error("expected Location header to be set")
	}
}

func TestSearchTasksFHIR_Success(t *testing.T) {
	h, e := newTestHandler()
	h.svc.CreateTask(nil, &Task{ForPatientID: uuid.New(), Intent: "order", Status: "draft"})

	req := httptest.NewRequest(http.MethodGet, "/fhir/Task", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if err := h.SearchTasksFHIR(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var bundle map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &bundle)
	if bundle["resourceType"] != "Bundle" {
		t.Errorf("expected Bundle resourceType, got %v", bundle["resourceType"])
	}
	if bundle["type"] != "searchset" {
		t.Errorf("expected searchset type, got %v", bundle["type"])
	}
}

func TestDeleteTaskFHIR_Success(t *testing.T) {
	h, e := newTestHandler()
	tk := &Task{ForPatientID: uuid.New(), Intent: "order", Status: "draft"}
	h.svc.CreateTask(nil, tk)

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(tk.FHIRID)
	if err := h.DeleteTaskFHIR(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

func TestDeleteTaskFHIR_NotFound(t *testing.T) {
	h, e := newTestHandler()
	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("nonexistent")
	if err := h.DeleteTaskFHIR(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
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
		"POST:/api/v1/tasks",
		"GET:/api/v1/tasks",
		"GET:/api/v1/tasks/:id",
		"PUT:/api/v1/tasks/:id",
		"DELETE:/api/v1/tasks/:id",
		"GET:/fhir/Task",
		"GET:/fhir/Task/:id",
		"POST:/fhir/Task",
		"PUT:/fhir/Task/:id",
		"DELETE:/fhir/Task/:id",
	}
	for _, path := range expected {
		if !routePaths[path] {
			t.Errorf("missing route: %s", path)
		}
	}
}
