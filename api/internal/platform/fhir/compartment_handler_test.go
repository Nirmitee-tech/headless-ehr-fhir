package fhir

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
)

// =========== CompartmentHandler Tests ===========

func TestCompartmentHandler_PatientCompartmentSearch_Success(t *testing.T) {
	h := NewCompartmentHandler()
	e := echo.New()

	// Register a mock search handler for Observation
	var capturedPatientParam string
	h.RegisterSearchHandler("Observation", func(c echo.Context) error {
		capturedPatientParam = c.QueryParam("patient")
		return c.JSON(http.StatusOK, map[string]interface{}{
			"resourceType": "Bundle",
			"type":         "searchset",
			"total":        0,
		})
	})

	h.RegisterRoutes(e.Group("/fhir"))

	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient/patient-123/Observation", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("pid", "resourceType")
	c.SetParamValues("patient-123", "Observation")

	err := h.PatientCompartmentSearch(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	if capturedPatientParam != "patient-123" {
		t.Errorf("expected patient param 'patient-123', got %q", capturedPatientParam)
	}
}

func TestCompartmentHandler_PatientCompartmentSearch_Condition(t *testing.T) {
	h := NewCompartmentHandler()
	e := echo.New()

	var capturedPatientParam string
	h.RegisterSearchHandler("Condition", func(c echo.Context) error {
		capturedPatientParam = c.QueryParam("patient")
		return c.JSON(http.StatusOK, map[string]interface{}{
			"resourceType": "Bundle",
			"type":         "searchset",
		})
	})

	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient/p1/Condition", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("pid", "resourceType")
	c.SetParamValues("p1", "Condition")

	err := h.PatientCompartmentSearch(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	if capturedPatientParam != "p1" {
		t.Errorf("expected patient param 'p1', got %q", capturedPatientParam)
	}
}

func TestCompartmentHandler_PatientCompartmentSearch_NotInCompartment(t *testing.T) {
	h := NewCompartmentHandler()
	e := echo.New()

	// "Organization" is not in the Patient compartment
	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient/p1/Organization", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("pid", "resourceType")
	c.SetParamValues("p1", "Organization")

	err := h.PatientCompartmentSearch(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}

	var outcome OperationOutcome
	json.Unmarshal(rec.Body.Bytes(), &outcome)
	if len(outcome.Issue) == 0 {
		t.Error("expected OperationOutcome with issues")
	}
}

func TestCompartmentHandler_PatientCompartmentSearch_NoHandler(t *testing.T) {
	h := NewCompartmentHandler()
	e := echo.New()

	// Encounter is in the compartment, but no handler registered
	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient/p1/Encounter", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("pid", "resourceType")
	c.SetParamValues("p1", "Encounter")

	err := h.PatientCompartmentSearch(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestCompartmentHandler_PatientCompartmentSearch_EmptyPID(t *testing.T) {
	h := NewCompartmentHandler()
	e := echo.New()

	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient//Observation", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("pid", "resourceType")
	c.SetParamValues("", "Observation")

	err := h.PatientCompartmentSearch(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestCompartmentHandler_PatientCompartmentSearch_MedicationNoParam(t *testing.T) {
	h := NewCompartmentHandler()
	e := echo.New()

	// Medication is in the compartment definition but has empty linking params
	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient/p1/Medication", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("pid", "resourceType")
	c.SetParamValues("p1", "Medication")

	err := h.PatientCompartmentSearch(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Medication has empty params list, so GetCompartmentParam returns ""
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for resource with no linking param, got %d", rec.Code)
	}
}

func TestCompartmentHandler_RegisterSearchHandler(t *testing.T) {
	h := NewCompartmentHandler()

	handlerCalled := false
	h.RegisterSearchHandler("Observation", func(c echo.Context) error {
		handlerCalled = true
		return c.NoContent(http.StatusOK)
	})

	if _, ok := h.searchHandlers["Observation"]; !ok {
		t.Error("expected Observation handler to be registered")
	}

	// Verify handler can be called
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient/p1/Observation", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("pid", "resourceType")
	c.SetParamValues("p1", "Observation")

	h.PatientCompartmentSearch(c)
	if !handlerCalled {
		t.Error("expected registered handler to be called")
	}
}

func TestCompartmentHandler_RegisterRoutes(t *testing.T) {
	h := NewCompartmentHandler()
	e := echo.New()
	fhirGroup := e.Group("/fhir")

	h.RegisterRoutes(fhirGroup)

	routes := e.Routes()
	found := false
	for _, r := range routes {
		if r.Method == "GET" && r.Path == "/fhir/Patient/:pid/:resourceType" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected route GET /fhir/Patient/:pid/:resourceType")
	}
}

func TestCompartmentHandler_PreservesExistingQueryParams(t *testing.T) {
	h := NewCompartmentHandler()
	e := echo.New()

	var capturedStatus string
	var capturedPatient string
	h.RegisterSearchHandler("Observation", func(c echo.Context) error {
		capturedStatus = c.QueryParam("status")
		capturedPatient = c.QueryParam("patient")
		return c.NoContent(http.StatusOK)
	})

	// Request with existing query params
	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient/p1/Observation?status=final", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("pid", "resourceType")
	c.SetParamValues("p1", "Observation")

	h.PatientCompartmentSearch(c)

	if capturedPatient != "p1" {
		t.Errorf("expected patient 'p1', got %q", capturedPatient)
	}
	if capturedStatus != "final" {
		t.Errorf("expected status 'final', got %q", capturedStatus)
	}
}
