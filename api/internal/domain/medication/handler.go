package medication

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/ehr/ehr/internal/domain/diagnostics"
	"github.com/ehr/ehr/internal/platform/auth"
	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/ehr/ehr/pkg/pagination"
)

type Handler struct {
	svc     *Service
	diagSvc *diagnostics.Service
}

func NewHandler(svc *Service, diagSvc ...*diagnostics.Service) *Handler {
	h := &Handler{svc: svc}
	if len(diagSvc) > 0 {
		h.diagSvc = diagSvc[0]
	}
	return h
}

func (h *Handler) RegisterRoutes(api *echo.Group, fhirGroup *echo.Group) {
	// Read endpoints – admin, physician, nurse, pharmacist
	readGroup := api.Group("", auth.RequireRole("admin", "physician", "nurse", "pharmacist"))
	readGroup.GET("/medications", h.ListMedications)
	readGroup.GET("/medications/:id", h.GetMedication)
	readGroup.GET("/medications/:id/ingredients", h.GetIngredients)
	readGroup.GET("/medication-requests", h.ListMedicationRequests)
	readGroup.GET("/medication-requests/:id", h.GetMedicationRequest)
	readGroup.GET("/medication-requests/:id/status-history", h.GetMedicationRequestStatusHistory)
	readGroup.GET("/medication-administrations", h.ListMedicationAdministrations)
	readGroup.GET("/medication-administrations/:id", h.GetMedicationAdministration)
	readGroup.GET("/medication-dispenses", h.ListMedicationDispenses)
	readGroup.GET("/medication-dispenses/:id", h.GetMedicationDispense)
	readGroup.GET("/medication-statements", h.ListMedicationStatements)
	readGroup.GET("/medication-statements/:id", h.GetMedicationStatement)

	// Write endpoints – admin, physician, pharmacist
	writeGroup := api.Group("", auth.RequireRole("admin", "physician", "pharmacist"))
	writeGroup.POST("/medications", h.CreateMedication)
	writeGroup.PUT("/medications/:id", h.UpdateMedication)
	writeGroup.DELETE("/medications/:id", h.DeleteMedication)
	writeGroup.POST("/medications/:id/ingredients", h.AddIngredient)
	writeGroup.DELETE("/medications/:id/ingredients/:ingredientId", h.RemoveIngredient)
	writeGroup.POST("/medication-requests", h.CreateMedicationRequest)
	writeGroup.PUT("/medication-requests/:id", h.UpdateMedicationRequest)
	writeGroup.DELETE("/medication-requests/:id", h.DeleteMedicationRequest)
	writeGroup.POST("/medication-administrations", h.CreateMedicationAdministration)
	writeGroup.PUT("/medication-administrations/:id", h.UpdateMedicationAdministration)
	writeGroup.DELETE("/medication-administrations/:id", h.DeleteMedicationAdministration)
	writeGroup.POST("/medication-dispenses", h.CreateMedicationDispense)
	writeGroup.PUT("/medication-dispenses/:id", h.UpdateMedicationDispense)
	writeGroup.DELETE("/medication-dispenses/:id", h.DeleteMedicationDispense)
	writeGroup.POST("/medication-statements", h.CreateMedicationStatement)
	writeGroup.PUT("/medication-statements/:id", h.UpdateMedicationStatement)
	writeGroup.DELETE("/medication-statements/:id", h.DeleteMedicationStatement)

	// FHIR read endpoints
	fhirRead := fhirGroup.Group("", auth.RequireRole("admin", "physician", "nurse", "pharmacist"))
	fhirRead.GET("/Medication", h.SearchMedicationsFHIR)
	fhirRead.GET("/Medication/:id", h.GetMedicationFHIR)
	fhirRead.GET("/MedicationRequest", h.SearchMedicationRequestsFHIR)
	fhirRead.GET("/MedicationRequest/:id", h.GetMedicationRequestFHIR)
	fhirRead.GET("/MedicationAdministration", h.SearchMedicationAdministrationsFHIR)
	fhirRead.GET("/MedicationAdministration/:id", h.GetMedicationAdministrationFHIR)
	fhirRead.GET("/MedicationDispense", h.SearchMedicationDispensesFHIR)
	fhirRead.GET("/MedicationDispense/:id", h.GetMedicationDispenseFHIR)
	fhirRead.GET("/MedicationStatement", h.SearchMedicationStatementsFHIR)
	fhirRead.GET("/MedicationStatement/:id", h.GetMedicationStatementFHIR)

	// FHIR write endpoints
	fhirWrite := fhirGroup.Group("", auth.RequireRole("admin", "physician", "pharmacist"))
	fhirWrite.POST("/Medication", h.CreateMedicationFHIR)
	fhirWrite.PUT("/Medication/:id", h.UpdateMedicationFHIR)
	fhirWrite.DELETE("/Medication/:id", h.DeleteMedicationFHIR)
	fhirWrite.PATCH("/Medication/:id", h.PatchMedicationFHIR)
	fhirWrite.POST("/MedicationRequest", h.CreateMedicationRequestFHIR)
	fhirWrite.PUT("/MedicationRequest/:id", h.UpdateMedicationRequestFHIR)
	fhirWrite.DELETE("/MedicationRequest/:id", h.DeleteMedicationRequestFHIR)
	fhirWrite.PATCH("/MedicationRequest/:id", h.PatchMedicationRequestFHIR)
	fhirWrite.POST("/MedicationAdministration", h.CreateMedicationAdministrationFHIR)
	fhirWrite.PUT("/MedicationAdministration/:id", h.UpdateMedicationAdministrationFHIR)
	fhirWrite.DELETE("/MedicationAdministration/:id", h.DeleteMedicationAdministrationFHIR)
	fhirWrite.PATCH("/MedicationAdministration/:id", h.PatchMedicationAdministrationFHIR)
	fhirWrite.POST("/MedicationDispense", h.CreateMedicationDispenseFHIR)
	fhirWrite.PUT("/MedicationDispense/:id", h.UpdateMedicationDispenseFHIR)
	fhirWrite.DELETE("/MedicationDispense/:id", h.DeleteMedicationDispenseFHIR)
	fhirWrite.PATCH("/MedicationDispense/:id", h.PatchMedicationDispenseFHIR)
	fhirWrite.POST("/MedicationStatement", h.CreateMedicationStatementFHIR)
	fhirWrite.PUT("/MedicationStatement/:id", h.UpdateMedicationStatementFHIR)
	fhirWrite.DELETE("/MedicationStatement/:id", h.DeleteMedicationStatementFHIR)
	fhirWrite.PATCH("/MedicationStatement/:id", h.PatchMedicationStatementFHIR)

	// FHIR POST _search endpoints
	fhirRead.POST("/Medication/_search", h.SearchMedicationsFHIR)
	fhirRead.POST("/MedicationRequest/_search", h.SearchMedicationRequestsFHIR)
	fhirRead.POST("/MedicationAdministration/_search", h.SearchMedicationAdministrationsFHIR)
	fhirRead.POST("/MedicationDispense/_search", h.SearchMedicationDispensesFHIR)
	fhirRead.POST("/MedicationStatement/_search", h.SearchMedicationStatementsFHIR)

	// FHIR vread and history endpoints
	fhirRead.GET("/Medication/:id/_history/:vid", h.VreadMedicationFHIR)
	fhirRead.GET("/Medication/:id/_history", h.HistoryMedicationFHIR)
	fhirRead.GET("/MedicationRequest/:id/_history/:vid", h.VreadMedicationRequestFHIR)
	fhirRead.GET("/MedicationRequest/:id/_history", h.HistoryMedicationRequestFHIR)
	fhirRead.GET("/MedicationAdministration/:id/_history/:vid", h.VreadMedicationAdministrationFHIR)
	fhirRead.GET("/MedicationAdministration/:id/_history", h.HistoryMedicationAdministrationFHIR)
	fhirRead.GET("/MedicationDispense/:id/_history/:vid", h.VreadMedicationDispenseFHIR)
	fhirRead.GET("/MedicationDispense/:id/_history", h.HistoryMedicationDispenseFHIR)
	fhirRead.GET("/MedicationStatement/:id/_history/:vid", h.VreadMedicationStatementFHIR)
	fhirRead.GET("/MedicationStatement/:id/_history", h.HistoryMedicationStatementFHIR)
}

// -- Medication Handlers --

func (h *Handler) CreateMedication(c echo.Context) error {
	var m Medication
	if err := c.Bind(&m); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateMedication(c.Request().Context(), &m); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, m)
}

func (h *Handler) GetMedication(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	m, err := h.svc.GetMedication(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "medication not found")
	}
	return c.JSON(http.StatusOK, m)
}

func (h *Handler) ListMedications(c echo.Context) error {
	pg := pagination.FromContext(c)
	params := fhir.ExtractSearchParams(c)
	items, total, err := h.svc.SearchMedications(c.Request().Context(), params, pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *Handler) UpdateMedication(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var m Medication
	if err := c.Bind(&m); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	m.ID = id
	if err := h.svc.UpdateMedication(c.Request().Context(), &m); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, m)
}

func (h *Handler) DeleteMedication(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeleteMedication(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) AddIngredient(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var ing MedicationIngredient
	if err := c.Bind(&ing); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	ing.MedicationID = id
	if err := h.svc.AddIngredient(c.Request().Context(), &ing); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, ing)
}

func (h *Handler) GetIngredients(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	ingredients, err := h.svc.GetIngredients(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, ingredients)
}

func (h *Handler) RemoveIngredient(c echo.Context) error {
	ingID, err := uuid.Parse(c.Param("ingredientId"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid ingredient id")
	}
	if err := h.svc.RemoveIngredient(c.Request().Context(), ingID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

// -- MedicationRequest Handlers --

func (h *Handler) CreateMedicationRequest(c echo.Context) error {
	var mr MedicationRequest
	if err := c.Bind(&mr); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateMedicationRequest(c.Request().Context(), &mr); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, mr)
}

func (h *Handler) GetMedicationRequest(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	mr, err := h.svc.GetMedicationRequest(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "medication request not found")
	}
	return c.JSON(http.StatusOK, mr)
}

func (h *Handler) ListMedicationRequests(c echo.Context) error {
	pg := pagination.FromContext(c)
	if patientID := c.QueryParam("patient_id"); patientID != "" {
		pid, err := uuid.Parse(patientID)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid patient_id")
		}
		items, total, err := h.svc.ListMedicationRequestsByPatient(c.Request().Context(), pid, pg.Limit, pg.Offset)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
	}
	items, total, err := h.svc.SearchMedicationRequests(c.Request().Context(), nil, pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *Handler) UpdateMedicationRequest(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var mr MedicationRequest
	if err := c.Bind(&mr); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	mr.ID = id
	if err := h.svc.UpdateMedicationRequest(c.Request().Context(), &mr); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, mr)
}

func (h *Handler) DeleteMedicationRequest(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeleteMedicationRequest(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

// -- MedicationRequest Status History Handler --

func (h *Handler) GetMedicationRequestStatusHistory(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if h.diagSvc == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "status history not configured")
	}
	history, err := h.diagSvc.GetStatusHistory(c.Request().Context(), "MedicationRequest", id)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, history)
}

// -- MedicationAdministration Handlers --

func (h *Handler) CreateMedicationAdministration(c echo.Context) error {
	var ma MedicationAdministration
	if err := c.Bind(&ma); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateMedicationAdministration(c.Request().Context(), &ma); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, ma)
}

func (h *Handler) GetMedicationAdministration(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	ma, err := h.svc.GetMedicationAdministration(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "medication administration not found")
	}
	return c.JSON(http.StatusOK, ma)
}

func (h *Handler) ListMedicationAdministrations(c echo.Context) error {
	pg := pagination.FromContext(c)
	if patientID := c.QueryParam("patient_id"); patientID != "" {
		pid, err := uuid.Parse(patientID)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid patient_id")
		}
		items, total, err := h.svc.ListMedicationAdministrationsByPatient(c.Request().Context(), pid, pg.Limit, pg.Offset)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
	}
	items, total, err := h.svc.SearchMedicationAdministrations(c.Request().Context(), nil, pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *Handler) UpdateMedicationAdministration(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var ma MedicationAdministration
	if err := c.Bind(&ma); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	ma.ID = id
	if err := h.svc.UpdateMedicationAdministration(c.Request().Context(), &ma); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, ma)
}

func (h *Handler) DeleteMedicationAdministration(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeleteMedicationAdministration(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

// -- MedicationDispense Handlers --

func (h *Handler) CreateMedicationDispense(c echo.Context) error {
	var md MedicationDispense
	if err := c.Bind(&md); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateMedicationDispense(c.Request().Context(), &md); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, md)
}

func (h *Handler) GetMedicationDispense(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	md, err := h.svc.GetMedicationDispense(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "medication dispense not found")
	}
	return c.JSON(http.StatusOK, md)
}

func (h *Handler) ListMedicationDispenses(c echo.Context) error {
	pg := pagination.FromContext(c)
	if patientID := c.QueryParam("patient_id"); patientID != "" {
		pid, err := uuid.Parse(patientID)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid patient_id")
		}
		items, total, err := h.svc.ListMedicationDispensesByPatient(c.Request().Context(), pid, pg.Limit, pg.Offset)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
	}
	items, total, err := h.svc.SearchMedicationDispenses(c.Request().Context(), nil, pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *Handler) UpdateMedicationDispense(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var md MedicationDispense
	if err := c.Bind(&md); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	md.ID = id
	if err := h.svc.UpdateMedicationDispense(c.Request().Context(), &md); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, md)
}

func (h *Handler) DeleteMedicationDispense(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeleteMedicationDispense(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

// -- MedicationStatement Handlers --

func (h *Handler) CreateMedicationStatement(c echo.Context) error {
	var ms MedicationStatement
	if err := c.Bind(&ms); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateMedicationStatement(c.Request().Context(), &ms); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, ms)
}

func (h *Handler) GetMedicationStatement(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	ms, err := h.svc.GetMedicationStatement(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "medication statement not found")
	}
	return c.JSON(http.StatusOK, ms)
}

func (h *Handler) ListMedicationStatements(c echo.Context) error {
	pg := pagination.FromContext(c)
	if patientID := c.QueryParam("patient_id"); patientID != "" {
		pid, err := uuid.Parse(patientID)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid patient_id")
		}
		items, total, err := h.svc.ListMedicationStatementsByPatient(c.Request().Context(), pid, pg.Limit, pg.Offset)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
	}
	items, total, err := h.svc.SearchMedicationStatements(c.Request().Context(), nil, pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *Handler) UpdateMedicationStatement(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var ms MedicationStatement
	if err := c.Bind(&ms); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	ms.ID = id
	if err := h.svc.UpdateMedicationStatement(c.Request().Context(), &ms); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, ms)
}

func (h *Handler) DeleteMedicationStatement(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeleteMedicationStatement(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

// -- FHIR Endpoints --

func (h *Handler) SearchMedicationsFHIR(c echo.Context) error {
	pg := pagination.FromContext(c)
	params := fhir.ExtractSearchParams(c)
	items, total, err := h.svc.SearchMedications(c.Request().Context(), params, pg.Limit, pg.Offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	resources := make([]interface{}, len(items))
	for i, item := range items {
		resources[i] = map[string]interface{}{
			"resourceType": "Medication",
			"id":           item.FHIRID,
			"status":       item.Status,
			"code": fhir.CodeableConcept{
				Coding: []fhir.Coding{{System: strVal(item.CodeSystem), Code: item.CodeValue, Display: item.CodeDisplay}},
			},
			"meta": fhir.Meta{
				LastUpdated: item.UpdatedAt,
				Profile:     []string{"http://hl7.org/fhir/us/core/StructureDefinition/us-core-medication"},
			},
		}
	}
	return c.JSON(http.StatusOK, fhir.NewSearchBundleWithLinks(resources, fhir.SearchBundleParams{
		BaseURL:  "/fhir/Medication",
		QueryStr: c.QueryString(),
		Count:    pg.Limit,
		Offset:   pg.Offset,
		Total:    total,
	}))
}

func (h *Handler) GetMedicationFHIR(c echo.Context) error {
	m, err := h.svc.GetMedicationByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Medication", c.Param("id")))
	}
	result := map[string]interface{}{
		"resourceType": "Medication",
		"id":           m.FHIRID,
		"status":       m.Status,
		"code": fhir.CodeableConcept{
			Coding: []fhir.Coding{{System: strVal(m.CodeSystem), Code: m.CodeValue, Display: m.CodeDisplay}},
		},
		"meta": fhir.Meta{
			LastUpdated: m.UpdatedAt,
			Profile:     []string{"http://hl7.org/fhir/us/core/StructureDefinition/us-core-medication"},
		},
	}
	if m.FormCode != nil {
		result["form"] = fhir.CodeableConcept{
			Coding: []fhir.Coding{{Code: *m.FormCode, Display: strVal(m.FormDisplay)}},
		}
	}
	return c.JSON(http.StatusOK, result)
}

func (h *Handler) CreateMedicationFHIR(c echo.Context) error {
	var m Medication
	if err := c.Bind(&m); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	if err := h.svc.CreateMedication(c.Request().Context(), &m); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	c.Response().Header().Set("Location", "/fhir/Medication/"+m.FHIRID)
	return c.JSON(http.StatusCreated, map[string]interface{}{
		"resourceType": "Medication",
		"id":           m.FHIRID,
		"status":       m.Status,
		"code": fhir.CodeableConcept{
			Coding: []fhir.Coding{{System: strVal(m.CodeSystem), Code: m.CodeValue, Display: m.CodeDisplay}},
		},
		"meta": fhir.Meta{
			LastUpdated: m.UpdatedAt,
			Profile:     []string{"http://hl7.org/fhir/us/core/StructureDefinition/us-core-medication"},
		},
	})
}

func (h *Handler) SearchMedicationRequestsFHIR(c echo.Context) error {
	pg := pagination.FromContext(c)
	params := fhir.ExtractSearchParams(c)
	items, total, err := h.svc.SearchMedicationRequests(c.Request().Context(), params, pg.Limit, pg.Offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	resources := make([]interface{}, len(items))
	for i, item := range items {
		resources[i] = item.ToFHIR()
	}
	return c.JSON(http.StatusOK, fhir.NewSearchBundleWithLinks(resources, fhir.SearchBundleParams{
		BaseURL:  "/fhir/MedicationRequest",
		QueryStr: c.QueryString(),
		Count:    pg.Limit,
		Offset:   pg.Offset,
		Total:    total,
	}))
}

func (h *Handler) GetMedicationRequestFHIR(c echo.Context) error {
	mr, err := h.svc.GetMedicationRequestByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("MedicationRequest", c.Param("id")))
	}
	return c.JSON(http.StatusOK, mr.ToFHIR())
}

func (h *Handler) CreateMedicationRequestFHIR(c echo.Context) error {
	var mr MedicationRequest
	if err := c.Bind(&mr); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	if err := h.svc.CreateMedicationRequest(c.Request().Context(), &mr); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	c.Response().Header().Set("Location", "/fhir/MedicationRequest/"+mr.FHIRID)
	return c.JSON(http.StatusCreated, mr.ToFHIR())
}

func (h *Handler) SearchMedicationAdministrationsFHIR(c echo.Context) error {
	pg := pagination.FromContext(c)
	params := fhir.ExtractSearchParams(c)
	items, total, err := h.svc.SearchMedicationAdministrations(c.Request().Context(), params, pg.Limit, pg.Offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	resources := make([]interface{}, len(items))
	for i, item := range items {
		resources[i] = item.ToFHIR()
	}
	return c.JSON(http.StatusOK, fhir.NewSearchBundleWithLinks(resources, fhir.SearchBundleParams{
		BaseURL:  "/fhir/MedicationAdministration",
		QueryStr: c.QueryString(),
		Count:    pg.Limit,
		Offset:   pg.Offset,
		Total:    total,
	}))
}

func (h *Handler) GetMedicationAdministrationFHIR(c echo.Context) error {
	ma, err := h.svc.GetMedicationAdministrationByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("MedicationAdministration", c.Param("id")))
	}
	return c.JSON(http.StatusOK, ma.ToFHIR())
}

func (h *Handler) CreateMedicationAdministrationFHIR(c echo.Context) error {
	var ma MedicationAdministration
	if err := c.Bind(&ma); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	if err := h.svc.CreateMedicationAdministration(c.Request().Context(), &ma); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	c.Response().Header().Set("Location", "/fhir/MedicationAdministration/"+ma.FHIRID)
	return c.JSON(http.StatusCreated, ma.ToFHIR())
}

func (h *Handler) SearchMedicationDispensesFHIR(c echo.Context) error {
	pg := pagination.FromContext(c)
	params := fhir.ExtractSearchParams(c)
	items, total, err := h.svc.SearchMedicationDispenses(c.Request().Context(), params, pg.Limit, pg.Offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	resources := make([]interface{}, len(items))
	for i, item := range items {
		resources[i] = item.ToFHIR()
	}
	return c.JSON(http.StatusOK, fhir.NewSearchBundleWithLinks(resources, fhir.SearchBundleParams{
		BaseURL:  "/fhir/MedicationDispense",
		QueryStr: c.QueryString(),
		Count:    pg.Limit,
		Offset:   pg.Offset,
		Total:    total,
	}))
}

func (h *Handler) GetMedicationDispenseFHIR(c echo.Context) error {
	md, err := h.svc.GetMedicationDispenseByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("MedicationDispense", c.Param("id")))
	}
	return c.JSON(http.StatusOK, md.ToFHIR())
}

func (h *Handler) CreateMedicationDispenseFHIR(c echo.Context) error {
	var md MedicationDispense
	if err := c.Bind(&md); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	if err := h.svc.CreateMedicationDispense(c.Request().Context(), &md); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	c.Response().Header().Set("Location", "/fhir/MedicationDispense/"+md.FHIRID)
	return c.JSON(http.StatusCreated, md.ToFHIR())
}

func (h *Handler) SearchMedicationStatementsFHIR(c echo.Context) error {
	pg := pagination.FromContext(c)
	params := fhir.ExtractSearchParams(c)
	items, total, err := h.svc.SearchMedicationStatements(c.Request().Context(), params, pg.Limit, pg.Offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	resources := make([]interface{}, len(items))
	for i, item := range items {
		resources[i] = item.ToFHIR()
	}
	return c.JSON(http.StatusOK, fhir.NewSearchBundleWithLinks(resources, fhir.SearchBundleParams{
		BaseURL:  "/fhir/MedicationStatement",
		QueryStr: c.QueryString(),
		Count:    pg.Limit,
		Offset:   pg.Offset,
		Total:    total,
	}))
}

func (h *Handler) GetMedicationStatementFHIR(c echo.Context) error {
	ms, err := h.svc.GetMedicationStatementByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("MedicationStatement", c.Param("id")))
	}
	return c.JSON(http.StatusOK, ms.ToFHIR())
}

func (h *Handler) CreateMedicationStatementFHIR(c echo.Context) error {
	var ms MedicationStatement
	if err := c.Bind(&ms); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	if err := h.svc.CreateMedicationStatement(c.Request().Context(), &ms); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	c.Response().Header().Set("Location", "/fhir/MedicationStatement/"+ms.FHIRID)
	return c.JSON(http.StatusCreated, ms.ToFHIR())
}

// -- FHIR Update Endpoints --

func (h *Handler) UpdateMedicationFHIR(c echo.Context) error {
	var m Medication
	if err := c.Bind(&m); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	existing, err := h.svc.GetMedicationByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Medication", c.Param("id")))
	}
	m.ID = existing.ID
	m.FHIRID = existing.FHIRID
	if err := h.svc.UpdateMedication(c.Request().Context(), &m); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, medicationToFHIR(&m))
}

func (h *Handler) UpdateMedicationRequestFHIR(c echo.Context) error {
	var mr MedicationRequest
	if err := c.Bind(&mr); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	existing, err := h.svc.GetMedicationRequestByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("MedicationRequest", c.Param("id")))
	}
	mr.ID = existing.ID
	mr.FHIRID = existing.FHIRID
	if err := h.svc.UpdateMedicationRequest(c.Request().Context(), &mr); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, mr.ToFHIR())
}

func (h *Handler) UpdateMedicationAdministrationFHIR(c echo.Context) error {
	var ma MedicationAdministration
	if err := c.Bind(&ma); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	existing, err := h.svc.GetMedicationAdministrationByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("MedicationAdministration", c.Param("id")))
	}
	ma.ID = existing.ID
	ma.FHIRID = existing.FHIRID
	if err := h.svc.UpdateMedicationAdministration(c.Request().Context(), &ma); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, ma.ToFHIR())
}

func (h *Handler) UpdateMedicationDispenseFHIR(c echo.Context) error {
	var md MedicationDispense
	if err := c.Bind(&md); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	existing, err := h.svc.GetMedicationDispenseByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("MedicationDispense", c.Param("id")))
	}
	md.ID = existing.ID
	md.FHIRID = existing.FHIRID
	if err := h.svc.UpdateMedicationDispense(c.Request().Context(), &md); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, md.ToFHIR())
}

func (h *Handler) UpdateMedicationStatementFHIR(c echo.Context) error {
	var ms MedicationStatement
	if err := c.Bind(&ms); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	existing, err := h.svc.GetMedicationStatementByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("MedicationStatement", c.Param("id")))
	}
	ms.ID = existing.ID
	ms.FHIRID = existing.FHIRID
	if err := h.svc.UpdateMedicationStatement(c.Request().Context(), &ms); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, ms.ToFHIR())
}

// -- FHIR Delete Endpoints --

func (h *Handler) DeleteMedicationFHIR(c echo.Context) error {
	existing, err := h.svc.GetMedicationByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Medication", c.Param("id")))
	}
	if err := h.svc.DeleteMedication(c.Request().Context(), existing.ID); err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) DeleteMedicationRequestFHIR(c echo.Context) error {
	existing, err := h.svc.GetMedicationRequestByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("MedicationRequest", c.Param("id")))
	}
	if err := h.svc.DeleteMedicationRequest(c.Request().Context(), existing.ID); err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) DeleteMedicationAdministrationFHIR(c echo.Context) error {
	existing, err := h.svc.GetMedicationAdministrationByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("MedicationAdministration", c.Param("id")))
	}
	if err := h.svc.DeleteMedicationAdministration(c.Request().Context(), existing.ID); err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) DeleteMedicationDispenseFHIR(c echo.Context) error {
	existing, err := h.svc.GetMedicationDispenseByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("MedicationDispense", c.Param("id")))
	}
	if err := h.svc.DeleteMedicationDispense(c.Request().Context(), existing.ID); err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) DeleteMedicationStatementFHIR(c echo.Context) error {
	existing, err := h.svc.GetMedicationStatementByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("MedicationStatement", c.Param("id")))
	}
	if err := h.svc.DeleteMedicationStatement(c.Request().Context(), existing.ID); err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	return c.NoContent(http.StatusNoContent)
}

// -- FHIR PATCH Endpoints --

func (h *Handler) PatchMedicationFHIR(c echo.Context) error {
	contentType := c.Request().Header.Get("Content-Type")
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome("failed to read request body"))
	}
	existing, err := h.svc.GetMedicationByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Medication", c.Param("id")))
	}
	currentResource := medicationToFHIR(existing)
	var patched map[string]interface{}
	if strings.Contains(contentType, "json-patch+json") {
		ops, err := fhir.ParseJSONPatch(body)
		if err != nil {
			return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
		}
		patched, err = fhir.ApplyJSONPatch(currentResource, ops)
		if err != nil {
			return c.JSON(http.StatusUnprocessableEntity, fhir.ErrorOutcome(err.Error()))
		}
	} else if strings.Contains(contentType, "merge-patch+json") {
		var mp map[string]interface{}
		if err := json.Unmarshal(body, &mp); err != nil {
			return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome("invalid merge patch JSON: "+err.Error()))
		}
		patched, err = fhir.ApplyMergePatch(currentResource, mp)
		if err != nil {
			return c.JSON(http.StatusUnprocessableEntity, fhir.ErrorOutcome(err.Error()))
		}
	} else {
		return c.JSON(http.StatusUnsupportedMediaType, fhir.ErrorOutcome(
			"PATCH requires Content-Type: application/json-patch+json or application/merge-patch+json"))
	}
	applyMedicationPatch(existing, patched)
	if err := h.svc.UpdateMedication(c.Request().Context(), existing); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, medicationToFHIR(existing))
}

func (h *Handler) PatchMedicationRequestFHIR(c echo.Context) error {
	contentType := c.Request().Header.Get("Content-Type")
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome("failed to read request body"))
	}
	existing, err := h.svc.GetMedicationRequestByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("MedicationRequest", c.Param("id")))
	}
	currentResource := existing.ToFHIR()
	var patched map[string]interface{}
	if strings.Contains(contentType, "json-patch+json") {
		ops, err := fhir.ParseJSONPatch(body)
		if err != nil {
			return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
		}
		patched, err = fhir.ApplyJSONPatch(currentResource, ops)
		if err != nil {
			return c.JSON(http.StatusUnprocessableEntity, fhir.ErrorOutcome(err.Error()))
		}
	} else if strings.Contains(contentType, "merge-patch+json") {
		var mp map[string]interface{}
		if err := json.Unmarshal(body, &mp); err != nil {
			return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome("invalid merge patch JSON: "+err.Error()))
		}
		patched, err = fhir.ApplyMergePatch(currentResource, mp)
		if err != nil {
			return c.JSON(http.StatusUnprocessableEntity, fhir.ErrorOutcome(err.Error()))
		}
	} else {
		return c.JSON(http.StatusUnsupportedMediaType, fhir.ErrorOutcome(
			"PATCH requires Content-Type: application/json-patch+json or application/merge-patch+json"))
	}
	applyMedicationRequestPatch(existing, patched)
	if err := h.svc.UpdateMedicationRequest(c.Request().Context(), existing); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, existing.ToFHIR())
}

func (h *Handler) PatchMedicationAdministrationFHIR(c echo.Context) error {
	contentType := c.Request().Header.Get("Content-Type")
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome("failed to read request body"))
	}
	existing, err := h.svc.GetMedicationAdministrationByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("MedicationAdministration", c.Param("id")))
	}
	currentResource := existing.ToFHIR()
	var patched map[string]interface{}
	if strings.Contains(contentType, "json-patch+json") {
		ops, err := fhir.ParseJSONPatch(body)
		if err != nil {
			return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
		}
		patched, err = fhir.ApplyJSONPatch(currentResource, ops)
		if err != nil {
			return c.JSON(http.StatusUnprocessableEntity, fhir.ErrorOutcome(err.Error()))
		}
	} else if strings.Contains(contentType, "merge-patch+json") {
		var mp map[string]interface{}
		if err := json.Unmarshal(body, &mp); err != nil {
			return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome("invalid merge patch JSON: "+err.Error()))
		}
		patched, err = fhir.ApplyMergePatch(currentResource, mp)
		if err != nil {
			return c.JSON(http.StatusUnprocessableEntity, fhir.ErrorOutcome(err.Error()))
		}
	} else {
		return c.JSON(http.StatusUnsupportedMediaType, fhir.ErrorOutcome(
			"PATCH requires Content-Type: application/json-patch+json or application/merge-patch+json"))
	}
	applyMedicationAdministrationPatch(existing, patched)
	if err := h.svc.UpdateMedicationAdministration(c.Request().Context(), existing); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, existing.ToFHIR())
}

func (h *Handler) PatchMedicationDispenseFHIR(c echo.Context) error {
	contentType := c.Request().Header.Get("Content-Type")
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome("failed to read request body"))
	}
	existing, err := h.svc.GetMedicationDispenseByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("MedicationDispense", c.Param("id")))
	}
	currentResource := existing.ToFHIR()
	var patched map[string]interface{}
	if strings.Contains(contentType, "json-patch+json") {
		ops, err := fhir.ParseJSONPatch(body)
		if err != nil {
			return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
		}
		patched, err = fhir.ApplyJSONPatch(currentResource, ops)
		if err != nil {
			return c.JSON(http.StatusUnprocessableEntity, fhir.ErrorOutcome(err.Error()))
		}
	} else if strings.Contains(contentType, "merge-patch+json") {
		var mp map[string]interface{}
		if err := json.Unmarshal(body, &mp); err != nil {
			return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome("invalid merge patch JSON: "+err.Error()))
		}
		patched, err = fhir.ApplyMergePatch(currentResource, mp)
		if err != nil {
			return c.JSON(http.StatusUnprocessableEntity, fhir.ErrorOutcome(err.Error()))
		}
	} else {
		return c.JSON(http.StatusUnsupportedMediaType, fhir.ErrorOutcome(
			"PATCH requires Content-Type: application/json-patch+json or application/merge-patch+json"))
	}
	applyMedicationDispensePatch(existing, patched)
	if err := h.svc.UpdateMedicationDispense(c.Request().Context(), existing); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, existing.ToFHIR())
}

func (h *Handler) PatchMedicationStatementFHIR(c echo.Context) error {
	contentType := c.Request().Header.Get("Content-Type")
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome("failed to read request body"))
	}
	existing, err := h.svc.GetMedicationStatementByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("MedicationStatement", c.Param("id")))
	}
	currentResource := existing.ToFHIR()
	var patched map[string]interface{}
	if strings.Contains(contentType, "json-patch+json") {
		ops, err := fhir.ParseJSONPatch(body)
		if err != nil {
			return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
		}
		patched, err = fhir.ApplyJSONPatch(currentResource, ops)
		if err != nil {
			return c.JSON(http.StatusUnprocessableEntity, fhir.ErrorOutcome(err.Error()))
		}
	} else if strings.Contains(contentType, "merge-patch+json") {
		var mp map[string]interface{}
		if err := json.Unmarshal(body, &mp); err != nil {
			return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome("invalid merge patch JSON: "+err.Error()))
		}
		patched, err = fhir.ApplyMergePatch(currentResource, mp)
		if err != nil {
			return c.JSON(http.StatusUnprocessableEntity, fhir.ErrorOutcome(err.Error()))
		}
	} else {
		return c.JSON(http.StatusUnsupportedMediaType, fhir.ErrorOutcome(
			"PATCH requires Content-Type: application/json-patch+json or application/merge-patch+json"))
	}
	applyMedicationStatementPatch(existing, patched)
	if err := h.svc.UpdateMedicationStatement(c.Request().Context(), existing); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, existing.ToFHIR())
}

// -- FHIR vread and history endpoints --

func (h *Handler) VreadMedicationFHIR(c echo.Context) error {
	m, err := h.svc.GetMedicationByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Medication", c.Param("id")))
	}
	result := medicationToFHIR(m)
	fhir.SetVersionHeaders(c, 1, m.UpdatedAt.Format("2006-01-02T15:04:05Z"))
	return c.JSON(http.StatusOK, result)
}

func (h *Handler) HistoryMedicationFHIR(c echo.Context) error {
	m, err := h.svc.GetMedicationByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Medication", c.Param("id")))
	}
	result := medicationToFHIR(m)
	raw, _ := json.Marshal(result)
	entry := &fhir.HistoryEntry{
		ResourceType: "Medication", ResourceID: m.FHIRID, VersionID: 1,
		Resource: raw, Action: "create", Timestamp: m.CreatedAt,
	}
	return c.JSON(http.StatusOK, fhir.NewHistoryBundle([]*fhir.HistoryEntry{entry}, 1, "/fhir"))
}

func (h *Handler) VreadMedicationRequestFHIR(c echo.Context) error {
	mr, err := h.svc.GetMedicationRequestByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("MedicationRequest", c.Param("id")))
	}
	result := mr.ToFHIR()
	fhir.SetVersionHeaders(c, 1, mr.UpdatedAt.Format("2006-01-02T15:04:05Z"))
	return c.JSON(http.StatusOK, result)
}

func (h *Handler) HistoryMedicationRequestFHIR(c echo.Context) error {
	mr, err := h.svc.GetMedicationRequestByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("MedicationRequest", c.Param("id")))
	}
	result := mr.ToFHIR()
	raw, _ := json.Marshal(result)
	entry := &fhir.HistoryEntry{
		ResourceType: "MedicationRequest", ResourceID: mr.FHIRID, VersionID: 1,
		Resource: raw, Action: "create", Timestamp: mr.CreatedAt,
	}
	return c.JSON(http.StatusOK, fhir.NewHistoryBundle([]*fhir.HistoryEntry{entry}, 1, "/fhir"))
}

func (h *Handler) VreadMedicationAdministrationFHIR(c echo.Context) error {
	ma, err := h.svc.GetMedicationAdministrationByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("MedicationAdministration", c.Param("id")))
	}
	result := ma.ToFHIR()
	fhir.SetVersionHeaders(c, 1, ma.UpdatedAt.Format("2006-01-02T15:04:05Z"))
	return c.JSON(http.StatusOK, result)
}

func (h *Handler) HistoryMedicationAdministrationFHIR(c echo.Context) error {
	ma, err := h.svc.GetMedicationAdministrationByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("MedicationAdministration", c.Param("id")))
	}
	result := ma.ToFHIR()
	raw, _ := json.Marshal(result)
	entry := &fhir.HistoryEntry{
		ResourceType: "MedicationAdministration", ResourceID: ma.FHIRID, VersionID: 1,
		Resource: raw, Action: "create", Timestamp: ma.CreatedAt,
	}
	return c.JSON(http.StatusOK, fhir.NewHistoryBundle([]*fhir.HistoryEntry{entry}, 1, "/fhir"))
}

func (h *Handler) VreadMedicationDispenseFHIR(c echo.Context) error {
	md, err := h.svc.GetMedicationDispenseByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("MedicationDispense", c.Param("id")))
	}
	result := md.ToFHIR()
	fhir.SetVersionHeaders(c, 1, md.UpdatedAt.Format("2006-01-02T15:04:05Z"))
	return c.JSON(http.StatusOK, result)
}

func (h *Handler) HistoryMedicationDispenseFHIR(c echo.Context) error {
	md, err := h.svc.GetMedicationDispenseByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("MedicationDispense", c.Param("id")))
	}
	result := md.ToFHIR()
	raw, _ := json.Marshal(result)
	entry := &fhir.HistoryEntry{
		ResourceType: "MedicationDispense", ResourceID: md.FHIRID, VersionID: 1,
		Resource: raw, Action: "create", Timestamp: md.CreatedAt,
	}
	return c.JSON(http.StatusOK, fhir.NewHistoryBundle([]*fhir.HistoryEntry{entry}, 1, "/fhir"))
}

func (h *Handler) VreadMedicationStatementFHIR(c echo.Context) error {
	ms, err := h.svc.GetMedicationStatementByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("MedicationStatement", c.Param("id")))
	}
	result := ms.ToFHIR()
	fhir.SetVersionHeaders(c, 1, ms.UpdatedAt.Format("2006-01-02T15:04:05Z"))
	return c.JSON(http.StatusOK, result)
}

func (h *Handler) HistoryMedicationStatementFHIR(c echo.Context) error {
	ms, err := h.svc.GetMedicationStatementByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("MedicationStatement", c.Param("id")))
	}
	result := ms.ToFHIR()
	raw, _ := json.Marshal(result)
	entry := &fhir.HistoryEntry{
		ResourceType: "MedicationStatement", ResourceID: ms.FHIRID, VersionID: 1,
		Resource: raw, Action: "create", Timestamp: ms.CreatedAt,
	}
	return c.JSON(http.StatusOK, fhir.NewHistoryBundle([]*fhir.HistoryEntry{entry}, 1, "/fhir"))
}

// medicationToFHIR builds a FHIR resource map for Medication.
func medicationToFHIR(m *Medication) map[string]interface{} {
	result := map[string]interface{}{
		"resourceType": "Medication",
		"id":           m.FHIRID,
		"status":       m.Status,
		"code": fhir.CodeableConcept{
			Coding: []fhir.Coding{{System: strVal(m.CodeSystem), Code: m.CodeValue, Display: m.CodeDisplay}},
		},
		"meta": fhir.Meta{
			LastUpdated: m.UpdatedAt,
			Profile:     []string{"http://hl7.org/fhir/us/core/StructureDefinition/us-core-medication"},
		},
	}
	if m.FormCode != nil {
		result["form"] = fhir.CodeableConcept{
			Coding: []fhir.Coding{{Code: *m.FormCode, Display: strVal(m.FormDisplay)}},
		}
	}
	return result
}

// -- FHIR PATCH helpers --

func patchCodeableConcept(patched map[string]interface{}, key string) (code, display string, ok bool) {
	v, exists := patched[key]
	if !exists {
		return "", "", false
	}
	switch val := v.(type) {
	case map[string]interface{}:
		if coding, ok := val["coding"].([]interface{}); ok && len(coding) > 0 {
			if c, ok := coding[0].(map[string]interface{}); ok {
				code, _ = c["code"].(string)
				display, _ = c["display"].(string)
				return code, display, true
			}
		}
	}
	return "", "", false
}

func patchStringPtr(patched map[string]interface{}, key string, target **string) {
	if v, ok := patched[key].(string); ok {
		*target = &v
	}
}

func patchFloat64Ptr(patched map[string]interface{}, key string, target **float64) {
	if v, ok := patched[key].(float64); ok {
		*target = &v
	}
}

func patchBoolPtr(patched map[string]interface{}, key string, target **bool) {
	if v, ok := patched[key].(bool); ok {
		*target = &v
	}
}

func patchTimePtr(patched map[string]interface{}, key string, target **time.Time) {
	if v, ok := patched[key].(string); ok {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			*target = &t
		}
	}
}

func applyMedicationPatch(m *Medication, patched map[string]interface{}) {
	if v, ok := patched["status"].(string); ok {
		m.Status = v
	}
	// code
	if v, ok := patched["code"]; ok {
		if cc, ok := v.(map[string]interface{}); ok {
			if coding, ok := cc["coding"].([]interface{}); ok && len(coding) > 0 {
				if c, ok := coding[0].(map[string]interface{}); ok {
					if code, ok := c["code"].(string); ok {
						m.CodeValue = code
					}
					if display, ok := c["display"].(string); ok {
						m.CodeDisplay = display
					}
					if system, ok := c["system"].(string); ok {
						m.CodeSystem = &system
					}
				}
			}
		}
	}
	// form
	if code, display, ok := patchCodeableConcept(patched, "form"); ok {
		m.FormCode = &code
		m.FormDisplay = &display
	}
	// manufacturer
	if v, ok := patched["manufacturer"]; ok {
		if mfr, ok := v.(map[string]interface{}); ok {
			if display, ok := mfr["display"].(string); ok {
				m.ManufacturerName = &display
			}
		}
	}
	// batch
	if v, ok := patched["batch"]; ok {
		if batch, ok := v.(map[string]interface{}); ok {
			patchStringPtr(batch, "lotNumber", &m.LotNumber)
			if exp, ok := batch["expirationDate"].(string); ok {
				if t, err := time.Parse("2006-01-02", exp); err == nil {
					m.ExpirationDate = &t
				}
			}
		}
	}
}

func applyMedicationRequestPatch(mr *MedicationRequest, patched map[string]interface{}) {
	if v, ok := patched["status"].(string); ok {
		mr.Status = v
	}
	if v, ok := patched["intent"].(string); ok {
		mr.Intent = v
	}
	patchStringPtr(patched, "priority", &mr.Priority)
	// category
	if v, ok := patched["category"]; ok {
		if cats, ok := v.([]interface{}); ok && len(cats) > 0 {
			if cat, ok := cats[0].(map[string]interface{}); ok {
				if coding, ok := cat["coding"].([]interface{}); ok && len(coding) > 0 {
					if c, ok := coding[0].(map[string]interface{}); ok {
						if code, ok := c["code"].(string); ok {
							mr.CategoryCode = &code
						}
						if display, ok := c["display"].(string); ok {
							mr.CategoryDisplay = &display
						}
					}
				}
			}
		}
	}
	// authoredOn
	patchTimePtr(patched, "authoredOn", &mr.AuthoredOn)
	// reasonCode
	if v, ok := patched["reasonCode"]; ok {
		if reasons, ok := v.([]interface{}); ok && len(reasons) > 0 {
			if reason, ok := reasons[0].(map[string]interface{}); ok {
				if coding, ok := reason["coding"].([]interface{}); ok && len(coding) > 0 {
					if c, ok := coding[0].(map[string]interface{}); ok {
						if code, ok := c["code"].(string); ok {
							mr.ReasonCode = &code
						}
						if display, ok := c["display"].(string); ok {
							mr.ReasonDisplay = &display
						}
					}
				}
			}
		}
	}
	// dosageInstruction
	if v, ok := patched["dosageInstruction"]; ok {
		if dosages, ok := v.([]interface{}); ok && len(dosages) > 0 {
			if dosage, ok := dosages[0].(map[string]interface{}); ok {
				patchStringPtr(dosage, "text", &mr.DosageText)
				if v, ok := dosage["asNeededBoolean"].(bool); ok {
					mr.AsNeeded = &v
				}
				// timing
				if timing, ok := dosage["timing"].(map[string]interface{}); ok {
					if timingCode, ok := timing["code"].(map[string]interface{}); ok {
						if coding, ok := timingCode["coding"].([]interface{}); ok && len(coding) > 0 {
							if c, ok := coding[0].(map[string]interface{}); ok {
								if code, ok := c["code"].(string); ok {
									mr.DosageTimingCode = &code
								}
								if display, ok := c["display"].(string); ok {
									mr.DosageTimingDisplay = &display
								}
							}
						}
					}
				}
				// route
				if code, display, ok := patchCodeableConcept(dosage, "route"); ok {
					mr.DosageRouteCode = &code
					mr.DosageRouteDisplay = &display
				}
				// site
				if code, display, ok := patchCodeableConcept(dosage, "site"); ok {
					mr.DosageSiteCode = &code
					mr.DosageSiteDisplay = &display
				}
				// method
				if code, display, ok := patchCodeableConcept(dosage, "method"); ok {
					mr.DosageMethodCode = &code
					mr.DosageMethodDisplay = &display
				}
				// doseAndRate
				if dar, ok := dosage["doseAndRate"].([]interface{}); ok && len(dar) > 0 {
					if d, ok := dar[0].(map[string]interface{}); ok {
						if dq, ok := d["doseQuantity"].(map[string]interface{}); ok {
							patchFloat64Ptr(dq, "value", &mr.DoseQuantity)
							patchStringPtr(dq, "unit", &mr.DoseUnit)
						}
					}
				}
			}
		}
	}
	// dispenseRequest
	if v, ok := patched["dispenseRequest"]; ok {
		if dr, ok := v.(map[string]interface{}); ok {
			if qty, ok := dr["quantity"].(map[string]interface{}); ok {
				patchFloat64Ptr(qty, "value", &mr.QuantityValue)
				patchStringPtr(qty, "unit", &mr.QuantityUnit)
			}
			if supply, ok := dr["expectedSupplyDuration"].(map[string]interface{}); ok {
				if v, ok := supply["value"].(float64); ok {
					iv := int(v)
					mr.DaysSupply = &iv
				}
			}
			if nr, ok := dr["numberOfRepeatsAllowed"].(float64); ok {
				iv := int(nr)
				mr.RefillsAllowed = &iv
			}
			if vp, ok := dr["validityPeriod"].(map[string]interface{}); ok {
				patchTimePtr(vp, "start", &mr.ValidityStart)
				patchTimePtr(vp, "end", &mr.ValidityEnd)
			}
		}
	}
	// substitution
	if v, ok := patched["substitution"].(map[string]interface{}); ok {
		patchBoolPtr(v, "allowedBoolean", &mr.SubstitutionAllowed)
	}
	// note
	if v, ok := patched["note"]; ok {
		if notes, ok := v.([]interface{}); ok && len(notes) > 0 {
			if note, ok := notes[0].(map[string]interface{}); ok {
				patchStringPtr(note, "text", &mr.Note)
			}
		}
	}
}

func applyMedicationAdministrationPatch(ma *MedicationAdministration, patched map[string]interface{}) {
	if v, ok := patched["status"].(string); ok {
		ma.Status = v
	}
	// statusReason
	if code, display, ok := patchCodeableConcept(patched, "statusReason"); ok {
		ma.StatusReasonCode = &code
		ma.StatusReasonDisplay = &display
	}
	// category
	if code, display, ok := patchCodeableConcept(patched, "category"); ok {
		ma.CategoryCode = &code
		ma.CategoryDisplay = &display
	}
	// effectiveDateTime
	patchTimePtr(patched, "effectiveDateTime", &ma.EffectiveDatetime)
	// effectivePeriod
	if v, ok := patched["effectivePeriod"].(map[string]interface{}); ok {
		patchTimePtr(v, "start", &ma.EffectiveStart)
		patchTimePtr(v, "end", &ma.EffectiveEnd)
	}
	// reasonCode
	if v, ok := patched["reasonCode"]; ok {
		if reasons, ok := v.([]interface{}); ok && len(reasons) > 0 {
			if reason, ok := reasons[0].(map[string]interface{}); ok {
				if coding, ok := reason["coding"].([]interface{}); ok && len(coding) > 0 {
					if c, ok := coding[0].(map[string]interface{}); ok {
						if code, ok := c["code"].(string); ok {
							ma.ReasonCode = &code
						}
						if display, ok := c["display"].(string); ok {
							ma.ReasonDisplay = &display
						}
					}
				}
			}
		}
	}
	// dosage
	if v, ok := patched["dosage"].(map[string]interface{}); ok {
		patchStringPtr(v, "text", &ma.DosageText)
		// route
		if code, display, ok := patchCodeableConcept(v, "route"); ok {
			ma.DosageRouteCode = &code
			ma.DosageRouteDisplay = &display
		}
		// site
		if code, display, ok := patchCodeableConcept(v, "site"); ok {
			ma.DosageSiteCode = &code
			ma.DosageSiteDisplay = &display
		}
		// method
		if code, display, ok := patchCodeableConcept(v, "method"); ok {
			ma.DosageMethodCode = &code
			ma.DosageMethodDisplay = &display
		}
		// dose
		if dose, ok := v["dose"].(map[string]interface{}); ok {
			patchFloat64Ptr(dose, "value", &ma.DoseQuantity)
			patchStringPtr(dose, "unit", &ma.DoseUnit)
		}
		// rateQuantity
		if rate, ok := v["rateQuantity"].(map[string]interface{}); ok {
			patchFloat64Ptr(rate, "value", &ma.RateQuantity)
			patchStringPtr(rate, "unit", &ma.RateUnit)
		}
	}
	// performer
	if v, ok := patched["performer"]; ok {
		if performers, ok := v.([]interface{}); ok && len(performers) > 0 {
			if perf, ok := performers[0].(map[string]interface{}); ok {
				if actor, ok := perf["actor"].(map[string]interface{}); ok {
					if ref, ok := actor["reference"].(string); ok {
						parts := strings.Split(ref, "/")
						if len(parts) >= 2 {
							if id, err := uuid.Parse(parts[len(parts)-1]); err == nil {
								ma.PerformerID = &id
							}
						}
					}
				}
				if fn, ok := perf["function"]; ok {
					if fc, ok := fn.(map[string]interface{}); ok {
						if coding, ok := fc["coding"].([]interface{}); ok && len(coding) > 0 {
							if c, ok := coding[0].(map[string]interface{}); ok {
								if code, ok := c["code"].(string); ok {
									ma.PerformerRoleCode = &code
								}
								if display, ok := c["display"].(string); ok {
									ma.PerformerRoleDisplay = &display
								}
							}
						}
					}
				}
			}
		}
	}
	// note
	if v, ok := patched["note"]; ok {
		if notes, ok := v.([]interface{}); ok && len(notes) > 0 {
			if note, ok := notes[0].(map[string]interface{}); ok {
				patchStringPtr(note, "text", &ma.Note)
			}
		}
	}
}

func applyMedicationDispensePatch(md *MedicationDispense, patched map[string]interface{}) {
	if v, ok := patched["status"].(string); ok {
		md.Status = v
	}
	// statusReason
	if code, display, ok := patchCodeableConcept(patched, "statusReasonCodeableConcept"); ok {
		md.StatusReasonCode = &code
		md.StatusReasonDisplay = &display
	}
	// category
	if code, display, ok := patchCodeableConcept(patched, "category"); ok {
		md.CategoryCode = &code
		md.CategoryDisplay = &display
	}
	// quantity
	if v, ok := patched["quantity"].(map[string]interface{}); ok {
		patchFloat64Ptr(v, "value", &md.QuantityValue)
		patchStringPtr(v, "unit", &md.QuantityUnit)
	}
	// daysSupply
	if v, ok := patched["daysSupply"].(map[string]interface{}); ok {
		if val, ok := v["value"].(float64); ok {
			iv := int(val)
			md.DaysSupply = &iv
		}
	}
	// whenPrepared
	patchTimePtr(patched, "whenPrepared", &md.WhenPrepared)
	// whenHandedOver
	patchTimePtr(patched, "whenHandedOver", &md.WhenHandedOver)
	// performer
	if v, ok := patched["performer"]; ok {
		if performers, ok := v.([]interface{}); ok && len(performers) > 0 {
			if perf, ok := performers[0].(map[string]interface{}); ok {
				if actor, ok := perf["actor"].(map[string]interface{}); ok {
					if ref, ok := actor["reference"].(string); ok {
						parts := strings.Split(ref, "/")
						if len(parts) >= 2 {
							if id, err := uuid.Parse(parts[len(parts)-1]); err == nil {
								md.PerformerID = &id
							}
						}
					}
				}
			}
		}
	}
	// substitution
	if v, ok := patched["substitution"].(map[string]interface{}); ok {
		patchBoolPtr(v, "wasSubstituted", &md.WasSubstituted)
		if typeCC, ok := v["type"].(map[string]interface{}); ok {
			if coding, ok := typeCC["coding"].([]interface{}); ok && len(coding) > 0 {
				if c, ok := coding[0].(map[string]interface{}); ok {
					if code, ok := c["code"].(string); ok {
						md.SubstitutionTypeCode = &code
					}
				}
			}
		}
		if reasons, ok := v["reason"].([]interface{}); ok && len(reasons) > 0 {
			if reason, ok := reasons[0].(map[string]interface{}); ok {
				if text, ok := reason["text"].(string); ok {
					md.SubstitutionReason = &text
				}
			}
		}
	}
	// note
	if v, ok := patched["note"]; ok {
		if notes, ok := v.([]interface{}); ok && len(notes) > 0 {
			if note, ok := notes[0].(map[string]interface{}); ok {
				patchStringPtr(note, "text", &md.Note)
			}
		}
	}
}

func applyMedicationStatementPatch(ms *MedicationStatement, patched map[string]interface{}) {
	if v, ok := patched["status"].(string); ok {
		ms.Status = v
	}
	// statusReason
	if v, ok := patched["statusReason"]; ok {
		if reasons, ok := v.([]interface{}); ok && len(reasons) > 0 {
			if reason, ok := reasons[0].(map[string]interface{}); ok {
				if coding, ok := reason["coding"].([]interface{}); ok && len(coding) > 0 {
					if c, ok := coding[0].(map[string]interface{}); ok {
						if code, ok := c["code"].(string); ok {
							ms.StatusReasonCode = &code
						}
						if display, ok := c["display"].(string); ok {
							ms.StatusReasonDisplay = &display
						}
					}
				}
			}
		}
	}
	// category
	if code, display, ok := patchCodeableConcept(patched, "category"); ok {
		ms.CategoryCode = &code
		ms.CategoryDisplay = &display
	}
	// effectiveDateTime
	patchTimePtr(patched, "effectiveDateTime", &ms.EffectiveDatetime)
	// effectivePeriod
	if v, ok := patched["effectivePeriod"].(map[string]interface{}); ok {
		patchTimePtr(v, "start", &ms.EffectiveStart)
		patchTimePtr(v, "end", &ms.EffectiveEnd)
	}
	// dateAsserted
	patchTimePtr(patched, "dateAsserted", &ms.DateAsserted)
	// reasonCode
	if v, ok := patched["reasonCode"]; ok {
		if reasons, ok := v.([]interface{}); ok && len(reasons) > 0 {
			if reason, ok := reasons[0].(map[string]interface{}); ok {
				if coding, ok := reason["coding"].([]interface{}); ok && len(coding) > 0 {
					if c, ok := coding[0].(map[string]interface{}); ok {
						if code, ok := c["code"].(string); ok {
							ms.ReasonCode = &code
						}
						if display, ok := c["display"].(string); ok {
							ms.ReasonDisplay = &display
						}
					}
				}
			}
		}
	}
	// dosage
	if v, ok := patched["dosage"]; ok {
		if dosages, ok := v.([]interface{}); ok && len(dosages) > 0 {
			if dosage, ok := dosages[0].(map[string]interface{}); ok {
				patchStringPtr(dosage, "text", &ms.DosageText)
				// route
				if code, display, ok := patchCodeableConcept(dosage, "route"); ok {
					ms.DosageRouteCode = &code
					ms.DosageRouteDisplay = &display
				}
				// timing
				if timing, ok := dosage["timing"].(map[string]interface{}); ok {
					if timingCode, ok := timing["code"].(map[string]interface{}); ok {
						if coding, ok := timingCode["coding"].([]interface{}); ok && len(coding) > 0 {
							if c, ok := coding[0].(map[string]interface{}); ok {
								if code, ok := c["code"].(string); ok {
									ms.DosageTimingCode = &code
								}
								if display, ok := c["display"].(string); ok {
									ms.DosageTimingDisplay = &display
								}
							}
						}
					}
				}
				// doseAndRate
				if dar, ok := dosage["doseAndRate"].([]interface{}); ok && len(dar) > 0 {
					if d, ok := dar[0].(map[string]interface{}); ok {
						if dq, ok := d["doseQuantity"].(map[string]interface{}); ok {
							patchFloat64Ptr(dq, "value", &ms.DoseQuantity)
							patchStringPtr(dq, "unit", &ms.DoseUnit)
						}
					}
				}
			}
		}
	}
	// note
	if v, ok := patched["note"]; ok {
		if notes, ok := v.([]interface{}); ok && len(notes) > 0 {
			if note, ok := notes[0].(map[string]interface{}); ok {
				patchStringPtr(note, "text", &ms.Note)
			}
		}
	}
}
