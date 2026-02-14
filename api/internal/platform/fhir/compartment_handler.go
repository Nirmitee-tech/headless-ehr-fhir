package fhir

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

// CompartmentHandler implements Patient compartment search.
// It delegates to registered search handlers after injecting the patient filter parameter.
type CompartmentHandler struct {
	searchHandlers map[string]echo.HandlerFunc
}

// NewCompartmentHandler creates a new CompartmentHandler.
func NewCompartmentHandler() *CompartmentHandler {
	return &CompartmentHandler{
		searchHandlers: make(map[string]echo.HandlerFunc),
	}
}

// RegisterSearchHandler registers a FHIR search handler for a resource type.
func (h *CompartmentHandler) RegisterSearchHandler(resourceType string, handler echo.HandlerFunc) {
	h.searchHandlers[resourceType] = handler
}

// RegisterRoutes registers compartment search routes.
func (h *CompartmentHandler) RegisterRoutes(fhirGroup *echo.Group) {
	// Patient compartment: GET /fhir/Patient/:pid/:resourceType
	fhirGroup.GET("/Patient/:pid/:resourceType", h.PatientCompartmentSearch)
}

// PatientCompartmentSearch handles GET /fhir/Patient/:pid/:resourceType.
// It validates that the resource type belongs to the Patient compartment,
// injects the patient search parameter, and delegates to the registered search handler.
func (h *CompartmentHandler) PatientCompartmentSearch(c echo.Context) error {
	patientID := c.Param("pid")
	resourceType := c.Param("resourceType")

	if patientID == "" {
		return c.JSON(http.StatusBadRequest, ErrorOutcome("patient ID is required"))
	}

	// Check if resource type is in patient compartment
	param := GetCompartmentParam(&PatientCompartment, resourceType)
	if param == "" {
		return c.JSON(http.StatusBadRequest, ErrorOutcome(
			fmt.Sprintf("%s is not in the Patient compartment", resourceType)))
	}

	// Set the patient search parameter on the query
	q := c.QueryParams()
	q.Set(param, patientID)

	handler, ok := h.searchHandlers[resourceType]
	if !ok {
		return c.JSON(http.StatusNotFound, ErrorOutcome(
			fmt.Sprintf("no search handler registered for %s", resourceType)))
	}

	return handler(c)
}
