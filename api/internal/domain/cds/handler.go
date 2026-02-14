package cds

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/ehr/ehr/internal/platform/auth"
	"github.com/ehr/ehr/pkg/pagination"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterRoutes(api *echo.Group) {
	// Read endpoints – admin, physician, pharmacist
	readGroup := api.Group("", auth.RequireRole("admin", "physician", "pharmacist"))
	readGroup.GET("/cds-rules", h.ListCDSRules)
	readGroup.GET("/cds-rules/:id", h.GetCDSRule)
	readGroup.GET("/cds-alerts", h.ListCDSAlerts)
	readGroup.GET("/cds-alerts/:id", h.GetCDSAlert)
	readGroup.GET("/cds-alerts/:id/responses", h.GetAlertResponses)
	readGroup.GET("/drug-interactions", h.ListDrugInteractions)
	readGroup.GET("/drug-interactions/:id", h.GetDrugInteraction)
	readGroup.GET("/order-sets", h.ListOrderSets)
	readGroup.GET("/order-sets/:id", h.GetOrderSet)
	readGroup.GET("/order-sets/:id/sections", h.GetOrderSetSections)
	readGroup.GET("/order-sets/:id/items", h.GetOrderSetItems)
	readGroup.GET("/clinical-pathways", h.ListClinicalPathways)
	readGroup.GET("/clinical-pathways/:id", h.GetClinicalPathway)
	readGroup.GET("/clinical-pathways/:id/phases", h.GetPathwayPhases)
	readGroup.GET("/pathway-enrollments", h.ListPathwayEnrollments)
	readGroup.GET("/pathway-enrollments/:id", h.GetPathwayEnrollment)
	readGroup.GET("/formularies", h.ListFormularies)
	readGroup.GET("/formularies/:id", h.GetFormulary)
	readGroup.GET("/formularies/:id/items", h.GetFormularyItems)
	readGroup.GET("/medication-reconciliations", h.ListMedReconciliations)
	readGroup.GET("/medication-reconciliations/:id", h.GetMedReconciliation)
	readGroup.GET("/medication-reconciliations/:id/items", h.GetMedReconciliationItems)

	// Write endpoints – admin, physician
	writeGroup := api.Group("", auth.RequireRole("admin", "physician"))
	writeGroup.POST("/cds-rules", h.CreateCDSRule)
	writeGroup.PUT("/cds-rules/:id", h.UpdateCDSRule)
	writeGroup.DELETE("/cds-rules/:id", h.DeleteCDSRule)
	writeGroup.POST("/cds-alerts", h.CreateCDSAlert)
	writeGroup.PUT("/cds-alerts/:id", h.UpdateCDSAlert)
	writeGroup.DELETE("/cds-alerts/:id", h.DeleteCDSAlert)
	writeGroup.POST("/cds-alerts/:id/responses", h.AddAlertResponse)
	writeGroup.POST("/drug-interactions", h.CreateDrugInteraction)
	writeGroup.PUT("/drug-interactions/:id", h.UpdateDrugInteraction)
	writeGroup.DELETE("/drug-interactions/:id", h.DeleteDrugInteraction)
	writeGroup.POST("/order-sets", h.CreateOrderSet)
	writeGroup.PUT("/order-sets/:id", h.UpdateOrderSet)
	writeGroup.DELETE("/order-sets/:id", h.DeleteOrderSet)
	writeGroup.POST("/order-sets/:id/sections", h.AddOrderSetSection)
	writeGroup.POST("/order-sets/:id/items", h.AddOrderSetItem)
	writeGroup.POST("/clinical-pathways", h.CreateClinicalPathway)
	writeGroup.PUT("/clinical-pathways/:id", h.UpdateClinicalPathway)
	writeGroup.DELETE("/clinical-pathways/:id", h.DeleteClinicalPathway)
	writeGroup.POST("/clinical-pathways/:id/phases", h.AddPathwayPhase)
	writeGroup.POST("/pathway-enrollments", h.CreatePathwayEnrollment)
	writeGroup.PUT("/pathway-enrollments/:id", h.UpdatePathwayEnrollment)
	writeGroup.DELETE("/pathway-enrollments/:id", h.DeletePathwayEnrollment)
	writeGroup.POST("/formularies", h.CreateFormulary)
	writeGroup.PUT("/formularies/:id", h.UpdateFormulary)
	writeGroup.DELETE("/formularies/:id", h.DeleteFormulary)
	writeGroup.POST("/formularies/:id/items", h.AddFormularyItem)
	writeGroup.POST("/medication-reconciliations", h.CreateMedReconciliation)
	writeGroup.PUT("/medication-reconciliations/:id", h.UpdateMedReconciliation)
	writeGroup.DELETE("/medication-reconciliations/:id", h.DeleteMedReconciliation)
	writeGroup.POST("/medication-reconciliations/:id/items", h.AddMedReconciliationItem)
}

// -- CDS Rule Handlers --

func (h *Handler) CreateCDSRule(c echo.Context) error {
	var r CDSRule
	if err := c.Bind(&r); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateCDSRule(c.Request().Context(), &r); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, r)
}

func (h *Handler) GetCDSRule(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	r, err := h.svc.GetCDSRule(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "cds rule not found")
	}
	return c.JSON(http.StatusOK, r)
}

func (h *Handler) ListCDSRules(c echo.Context) error {
	pg := pagination.FromContext(c)
	items, total, err := h.svc.ListCDSRules(c.Request().Context(), pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *Handler) UpdateCDSRule(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var r CDSRule
	if err := c.Bind(&r); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	r.ID = id
	if err := h.svc.UpdateCDSRule(c.Request().Context(), &r); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, r)
}

func (h *Handler) DeleteCDSRule(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeleteCDSRule(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

// -- CDS Alert Handlers --

func (h *Handler) CreateCDSAlert(c echo.Context) error {
	var a CDSAlert
	if err := c.Bind(&a); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateCDSAlert(c.Request().Context(), &a); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, a)
}

func (h *Handler) GetCDSAlert(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	a, err := h.svc.GetCDSAlert(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "cds alert not found")
	}
	return c.JSON(http.StatusOK, a)
}

func (h *Handler) ListCDSAlerts(c echo.Context) error {
	pg := pagination.FromContext(c)
	if patientID := c.QueryParam("patient_id"); patientID != "" {
		pid, err := uuid.Parse(patientID)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid patient_id")
		}
		items, total, err := h.svc.ListCDSAlertsByPatient(c.Request().Context(), pid, pg.Limit, pg.Offset)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
	}
	items, total, err := h.svc.ListCDSAlerts(c.Request().Context(), pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *Handler) UpdateCDSAlert(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var a CDSAlert
	if err := c.Bind(&a); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	a.ID = id
	if err := h.svc.UpdateCDSAlert(c.Request().Context(), &a); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, a)
}

func (h *Handler) DeleteCDSAlert(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeleteCDSAlert(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) AddAlertResponse(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var resp CDSAlertResponse
	if err := c.Bind(&resp); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	resp.AlertID = id
	if err := h.svc.AddAlertResponse(c.Request().Context(), &resp); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, resp)
}

func (h *Handler) GetAlertResponses(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	responses, err := h.svc.GetAlertResponses(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, responses)
}

// -- Drug Interaction Handlers --

func (h *Handler) CreateDrugInteraction(c echo.Context) error {
	var d DrugInteraction
	if err := c.Bind(&d); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateDrugInteraction(c.Request().Context(), &d); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, d)
}

func (h *Handler) GetDrugInteraction(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	d, err := h.svc.GetDrugInteraction(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "drug interaction not found")
	}
	return c.JSON(http.StatusOK, d)
}

func (h *Handler) ListDrugInteractions(c echo.Context) error {
	pg := pagination.FromContext(c)
	items, total, err := h.svc.ListDrugInteractions(c.Request().Context(), pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *Handler) UpdateDrugInteraction(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var d DrugInteraction
	if err := c.Bind(&d); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	d.ID = id
	if err := h.svc.UpdateDrugInteraction(c.Request().Context(), &d); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, d)
}

func (h *Handler) DeleteDrugInteraction(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeleteDrugInteraction(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

// -- Order Set Handlers --

func (h *Handler) CreateOrderSet(c echo.Context) error {
	var o OrderSet
	if err := c.Bind(&o); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateOrderSet(c.Request().Context(), &o); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, o)
}

func (h *Handler) GetOrderSet(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	o, err := h.svc.GetOrderSet(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "order set not found")
	}
	return c.JSON(http.StatusOK, o)
}

func (h *Handler) ListOrderSets(c echo.Context) error {
	pg := pagination.FromContext(c)
	items, total, err := h.svc.ListOrderSets(c.Request().Context(), pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *Handler) UpdateOrderSet(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var o OrderSet
	if err := c.Bind(&o); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	o.ID = id
	if err := h.svc.UpdateOrderSet(c.Request().Context(), &o); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, o)
}

func (h *Handler) DeleteOrderSet(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeleteOrderSet(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) AddOrderSetSection(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var sec OrderSetSection
	if err := c.Bind(&sec); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	sec.OrderSetID = id
	if err := h.svc.AddOrderSetSection(c.Request().Context(), &sec); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, sec)
}

func (h *Handler) GetOrderSetSections(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	sections, err := h.svc.GetOrderSetSections(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, sections)
}

func (h *Handler) AddOrderSetItem(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var item OrderSetItem
	if err := c.Bind(&item); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	item.SectionID = id
	if err := h.svc.AddOrderSetItem(c.Request().Context(), &item); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, item)
}

func (h *Handler) GetOrderSetItems(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	items, err := h.svc.GetOrderSetItems(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, items)
}

// -- Clinical Pathway Handlers --

func (h *Handler) CreateClinicalPathway(c echo.Context) error {
	var p ClinicalPathway
	if err := c.Bind(&p); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateClinicalPathway(c.Request().Context(), &p); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, p)
}

func (h *Handler) GetClinicalPathway(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	p, err := h.svc.GetClinicalPathway(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "clinical pathway not found")
	}
	return c.JSON(http.StatusOK, p)
}

func (h *Handler) ListClinicalPathways(c echo.Context) error {
	pg := pagination.FromContext(c)
	items, total, err := h.svc.ListClinicalPathways(c.Request().Context(), pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *Handler) UpdateClinicalPathway(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var p ClinicalPathway
	if err := c.Bind(&p); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	p.ID = id
	if err := h.svc.UpdateClinicalPathway(c.Request().Context(), &p); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, p)
}

func (h *Handler) DeleteClinicalPathway(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeleteClinicalPathway(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) AddPathwayPhase(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var phase ClinicalPathwayPhase
	if err := c.Bind(&phase); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	phase.PathwayID = id
	if err := h.svc.AddPathwayPhase(c.Request().Context(), &phase); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, phase)
}

func (h *Handler) GetPathwayPhases(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	phases, err := h.svc.GetPathwayPhases(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, phases)
}

// -- Pathway Enrollment Handlers --

func (h *Handler) CreatePathwayEnrollment(c echo.Context) error {
	var e PatientPathwayEnrollment
	if err := c.Bind(&e); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreatePathwayEnrollment(c.Request().Context(), &e); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, e)
}

func (h *Handler) GetPathwayEnrollment(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	e, err := h.svc.GetPathwayEnrollment(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "pathway enrollment not found")
	}
	return c.JSON(http.StatusOK, e)
}

func (h *Handler) ListPathwayEnrollments(c echo.Context) error {
	pg := pagination.FromContext(c)
	if patientID := c.QueryParam("patient_id"); patientID != "" {
		pid, err := uuid.Parse(patientID)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid patient_id")
		}
		items, total, err := h.svc.ListPathwayEnrollmentsByPatient(c.Request().Context(), pid, pg.Limit, pg.Offset)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
	}
	items, total, err := h.svc.ListPathwayEnrollments(c.Request().Context(), pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *Handler) UpdatePathwayEnrollment(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var e PatientPathwayEnrollment
	if err := c.Bind(&e); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	e.ID = id
	if err := h.svc.UpdatePathwayEnrollment(c.Request().Context(), &e); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, e)
}

func (h *Handler) DeletePathwayEnrollment(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeletePathwayEnrollment(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

// -- Formulary Handlers --

func (h *Handler) CreateFormulary(c echo.Context) error {
	var f Formulary
	if err := c.Bind(&f); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateFormulary(c.Request().Context(), &f); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, f)
}

func (h *Handler) GetFormulary(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	f, err := h.svc.GetFormulary(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "formulary not found")
	}
	return c.JSON(http.StatusOK, f)
}

func (h *Handler) ListFormularies(c echo.Context) error {
	pg := pagination.FromContext(c)
	items, total, err := h.svc.ListFormularies(c.Request().Context(), pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *Handler) UpdateFormulary(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var f Formulary
	if err := c.Bind(&f); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	f.ID = id
	if err := h.svc.UpdateFormulary(c.Request().Context(), &f); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, f)
}

func (h *Handler) DeleteFormulary(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeleteFormulary(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) AddFormularyItem(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var item FormularyItem
	if err := c.Bind(&item); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	item.FormularyID = id
	if err := h.svc.AddFormularyItem(c.Request().Context(), &item); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, item)
}

func (h *Handler) GetFormularyItems(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	items, err := h.svc.GetFormularyItems(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, items)
}

// -- Medication Reconciliation Handlers --

func (h *Handler) CreateMedReconciliation(c echo.Context) error {
	var mr MedicationReconciliation
	if err := c.Bind(&mr); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateMedReconciliation(c.Request().Context(), &mr); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, mr)
}

func (h *Handler) GetMedReconciliation(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	mr, err := h.svc.GetMedReconciliation(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "medication reconciliation not found")
	}
	return c.JSON(http.StatusOK, mr)
}

func (h *Handler) ListMedReconciliations(c echo.Context) error {
	pg := pagination.FromContext(c)
	if patientID := c.QueryParam("patient_id"); patientID != "" {
		pid, err := uuid.Parse(patientID)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid patient_id")
		}
		items, total, err := h.svc.ListMedReconciliationsByPatient(c.Request().Context(), pid, pg.Limit, pg.Offset)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
	}
	items, total, err := h.svc.ListMedReconciliations(c.Request().Context(), pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *Handler) UpdateMedReconciliation(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var mr MedicationReconciliation
	if err := c.Bind(&mr); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	mr.ID = id
	if err := h.svc.UpdateMedReconciliation(c.Request().Context(), &mr); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, mr)
}

func (h *Handler) DeleteMedReconciliation(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeleteMedReconciliation(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) AddMedReconciliationItem(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var item MedicationReconciliationItem
	if err := c.Bind(&item); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	item.ReconciliationID = id
	if err := h.svc.AddMedReconciliationItem(c.Request().Context(), &item); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, item)
}

func (h *Handler) GetMedReconciliationItems(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	items, err := h.svc.GetMedReconciliationItems(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, items)
}
