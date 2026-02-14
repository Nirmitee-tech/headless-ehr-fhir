package medication

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/ehr/ehr/internal/platform/auth"
	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/ehr/ehr/pkg/pagination"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterRoutes(api *echo.Group, fhirGroup *echo.Group) {
	// Read endpoints – admin, physician, nurse, pharmacist
	readGroup := api.Group("", auth.RequireRole("admin", "physician", "nurse", "pharmacist"))
	readGroup.GET("/medications", h.ListMedications)
	readGroup.GET("/medications/:id", h.GetMedication)
	readGroup.GET("/medications/:id/ingredients", h.GetIngredients)
	readGroup.GET("/medication-requests", h.ListMedicationRequests)
	readGroup.GET("/medication-requests/:id", h.GetMedicationRequest)
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

	// FHIR POST _search endpoints
	fhirRead.POST("/Medication/_search", h.SearchMedicationsFHIR)
	fhirRead.POST("/MedicationRequest/_search", h.SearchMedicationRequestsFHIR)
	fhirRead.POST("/MedicationAdministration/_search", h.SearchMedicationAdministrationsFHIR)
	fhirRead.POST("/MedicationDispense/_search", h.SearchMedicationDispensesFHIR)

	// FHIR vread and history endpoints
	fhirRead.GET("/Medication/:id/_history/:vid", h.VreadMedicationFHIR)
	fhirRead.GET("/Medication/:id/_history", h.HistoryMedicationFHIR)
	fhirRead.GET("/MedicationRequest/:id/_history/:vid", h.VreadMedicationRequestFHIR)
	fhirRead.GET("/MedicationRequest/:id/_history", h.HistoryMedicationRequestFHIR)
	fhirRead.GET("/MedicationAdministration/:id/_history/:vid", h.VreadMedicationAdministrationFHIR)
	fhirRead.GET("/MedicationAdministration/:id/_history", h.HistoryMedicationAdministrationFHIR)
	fhirRead.GET("/MedicationDispense/:id/_history/:vid", h.VreadMedicationDispenseFHIR)
	fhirRead.GET("/MedicationDispense/:id/_history", h.HistoryMedicationDispenseFHIR)
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
	params := map[string]string{}
	for _, k := range []string{"code", "status", "form"} {
		if v := c.QueryParam(k); v != "" {
			params[k] = v
		}
	}
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
	params := map[string]string{}
	for _, k := range []string{"code", "status", "form"} {
		if v := c.QueryParam(k); v != "" {
			params[k] = v
		}
	}
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
			"meta": fhir.Meta{LastUpdated: item.UpdatedAt},
		}
	}
	return c.JSON(http.StatusOK, fhir.NewSearchBundle(resources, total, "/fhir/Medication"))
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
		"meta": fhir.Meta{LastUpdated: m.UpdatedAt},
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
		"meta": fhir.Meta{LastUpdated: m.UpdatedAt},
	})
}

func (h *Handler) SearchMedicationRequestsFHIR(c echo.Context) error {
	pg := pagination.FromContext(c)
	params := map[string]string{}
	for _, k := range []string{"patient", "status", "intent", "medication"} {
		if v := c.QueryParam(k); v != "" {
			params[k] = v
		}
	}
	items, total, err := h.svc.SearchMedicationRequests(c.Request().Context(), params, pg.Limit, pg.Offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	resources := make([]interface{}, len(items))
	for i, item := range items {
		resources[i] = item.ToFHIR()
	}
	return c.JSON(http.StatusOK, fhir.NewSearchBundle(resources, total, "/fhir/MedicationRequest"))
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
	params := map[string]string{}
	for _, k := range []string{"patient", "status", "medication"} {
		if v := c.QueryParam(k); v != "" {
			params[k] = v
		}
	}
	items, total, err := h.svc.SearchMedicationAdministrations(c.Request().Context(), params, pg.Limit, pg.Offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	resources := make([]interface{}, len(items))
	for i, item := range items {
		resources[i] = item.ToFHIR()
	}
	return c.JSON(http.StatusOK, fhir.NewSearchBundle(resources, total, "/fhir/MedicationAdministration"))
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
	params := map[string]string{}
	for _, k := range []string{"patient", "status", "medication"} {
		if v := c.QueryParam(k); v != "" {
			params[k] = v
		}
	}
	items, total, err := h.svc.SearchMedicationDispenses(c.Request().Context(), params, pg.Limit, pg.Offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	resources := make([]interface{}, len(items))
	for i, item := range items {
		resources[i] = item.ToFHIR()
	}
	return c.JSON(http.StatusOK, fhir.NewSearchBundle(resources, total, "/fhir/MedicationDispense"))
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
	_ = patched
	if v, ok := patched["status"].(string); ok {
		existing.Status = v
	}
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
	_ = patched
	if v, ok := patched["status"].(string); ok {
		existing.Status = v
	}
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
	_ = patched
	if v, ok := patched["status"].(string); ok {
		existing.Status = v
	}
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
	_ = patched
	if v, ok := patched["status"].(string); ok {
		existing.Status = v
	}
	if err := h.svc.UpdateMedicationDispense(c.Request().Context(), existing); err != nil {
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

// medicationToFHIR builds a FHIR resource map for Medication.
func medicationToFHIR(m *Medication) map[string]interface{} {
	result := map[string]interface{}{
		"resourceType": "Medication",
		"id":           m.FHIRID,
		"status":       m.Status,
		"code": fhir.CodeableConcept{
			Coding: []fhir.Coding{{System: strVal(m.CodeSystem), Code: m.CodeValue, Display: m.CodeDisplay}},
		},
		"meta": fhir.Meta{LastUpdated: m.UpdatedAt},
	}
	if m.FormCode != nil {
		result["form"] = fhir.CodeableConcept{
			Coding: []fhir.Coding{{Code: *m.FormCode, Display: strVal(m.FormDisplay)}},
		}
	}
	return result
}
