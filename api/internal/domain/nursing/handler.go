package nursing

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
	// Read endpoints – admin, physician, nurse
	readGroup := api.Group("", auth.RequireRole("admin", "physician", "nurse"))
	readGroup.GET("/flowsheet-templates", h.ListTemplates)
	readGroup.GET("/flowsheet-templates/:id", h.GetTemplate)
	readGroup.GET("/flowsheet-templates/:id/rows", h.GetTemplateRows)
	readGroup.GET("/flowsheet-entries", h.ListEntries)
	readGroup.GET("/flowsheet-entries/:id", h.GetEntry)
	readGroup.GET("/nursing-assessments", h.ListAssessments)
	readGroup.GET("/nursing-assessments/:id", h.GetAssessment)
	readGroup.GET("/fall-risk-assessments", h.ListFallRisk)
	readGroup.GET("/fall-risk-assessments/:id", h.GetFallRisk)
	readGroup.GET("/skin-assessments", h.ListSkinAssessments)
	readGroup.GET("/skin-assessments/:id", h.GetSkinAssessment)
	readGroup.GET("/pain-assessments", h.ListPainAssessments)
	readGroup.GET("/pain-assessments/:id", h.GetPainAssessment)
	readGroup.GET("/lines-drains", h.ListLinesDrains)
	readGroup.GET("/lines-drains/:id", h.GetLinesDrains)
	readGroup.GET("/restraints", h.ListRestraints)
	readGroup.GET("/restraints/:id", h.GetRestraint)
	readGroup.GET("/intake-output", h.ListIntakeOutput)
	readGroup.GET("/intake-output/:id", h.GetIntakeOutput)

	// Write endpoints – admin, nurse
	writeGroup := api.Group("", auth.RequireRole("admin", "nurse"))
	writeGroup.POST("/flowsheet-templates", h.CreateTemplate)
	writeGroup.PUT("/flowsheet-templates/:id", h.UpdateTemplate)
	writeGroup.DELETE("/flowsheet-templates/:id", h.DeleteTemplate)
	writeGroup.POST("/flowsheet-templates/:id/rows", h.AddTemplateRow)
	writeGroup.POST("/flowsheet-entries", h.CreateEntry)
	writeGroup.DELETE("/flowsheet-entries/:id", h.DeleteEntry)
	writeGroup.POST("/nursing-assessments", h.CreateAssessment)
	writeGroup.PUT("/nursing-assessments/:id", h.UpdateAssessment)
	writeGroup.DELETE("/nursing-assessments/:id", h.DeleteAssessment)
	writeGroup.POST("/fall-risk-assessments", h.CreateFallRisk)
	writeGroup.POST("/skin-assessments", h.CreateSkinAssessment)
	writeGroup.POST("/pain-assessments", h.CreatePainAssessment)
	writeGroup.POST("/lines-drains", h.CreateLinesDrains)
	writeGroup.PUT("/lines-drains/:id", h.UpdateLinesDrains)
	writeGroup.DELETE("/lines-drains/:id", h.DeleteLinesDrains)
	writeGroup.POST("/restraints", h.CreateRestraint)
	writeGroup.PUT("/restraints/:id", h.UpdateRestraint)
	writeGroup.POST("/intake-output", h.CreateIntakeOutput)
	writeGroup.DELETE("/intake-output/:id", h.DeleteIntakeOutput)
}

// -- Flowsheet Template Handlers --

func (h *Handler) CreateTemplate(c echo.Context) error {
	var t FlowsheetTemplate
	if err := c.Bind(&t); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateTemplate(c.Request().Context(), &t); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, t)
}

func (h *Handler) GetTemplate(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	t, err := h.svc.GetTemplate(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "template not found")
	}
	return c.JSON(http.StatusOK, t)
}

func (h *Handler) ListTemplates(c echo.Context) error {
	pg := pagination.FromContext(c)
	items, total, err := h.svc.ListTemplates(c.Request().Context(), pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *Handler) UpdateTemplate(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var t FlowsheetTemplate
	if err := c.Bind(&t); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	t.ID = id
	if err := h.svc.UpdateTemplate(c.Request().Context(), &t); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, t)
}

func (h *Handler) DeleteTemplate(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeleteTemplate(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) AddTemplateRow(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var r FlowsheetRow
	if err := c.Bind(&r); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	r.TemplateID = id
	if err := h.svc.AddTemplateRow(c.Request().Context(), &r); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, r)
}

func (h *Handler) GetTemplateRows(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	rows, err := h.svc.GetTemplateRows(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, rows)
}

// -- Flowsheet Entry Handlers --

func (h *Handler) CreateEntry(c echo.Context) error {
	var e FlowsheetEntry
	if err := c.Bind(&e); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateEntry(c.Request().Context(), &e); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, e)
}

func (h *Handler) GetEntry(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	e, err := h.svc.GetEntry(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "entry not found")
	}
	return c.JSON(http.StatusOK, e)
}

func (h *Handler) ListEntries(c echo.Context) error {
	pg := pagination.FromContext(c)
	if patientID := c.QueryParam("patient_id"); patientID != "" {
		pid, err := uuid.Parse(patientID)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid patient_id")
		}
		items, total, err := h.svc.ListEntriesByPatient(c.Request().Context(), pid, pg.Limit, pg.Offset)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
	}
	if encounterID := c.QueryParam("encounter_id"); encounterID != "" {
		eid, err := uuid.Parse(encounterID)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid encounter_id")
		}
		items, total, err := h.svc.ListEntriesByEncounter(c.Request().Context(), eid, pg.Limit, pg.Offset)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
	}
	items, total, err := h.svc.SearchEntries(c.Request().Context(), nil, pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *Handler) DeleteEntry(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeleteEntry(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

// -- Nursing Assessment Handlers --

func (h *Handler) CreateAssessment(c echo.Context) error {
	var a NursingAssessment
	if err := c.Bind(&a); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateAssessment(c.Request().Context(), &a); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, a)
}

func (h *Handler) GetAssessment(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	a, err := h.svc.GetAssessment(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "assessment not found")
	}
	return c.JSON(http.StatusOK, a)
}

func (h *Handler) ListAssessments(c echo.Context) error {
	pg := pagination.FromContext(c)
	if patientID := c.QueryParam("patient_id"); patientID != "" {
		pid, err := uuid.Parse(patientID)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid patient_id")
		}
		items, total, err := h.svc.ListAssessmentsByPatient(c.Request().Context(), pid, pg.Limit, pg.Offset)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
	}
	if encounterID := c.QueryParam("encounter_id"); encounterID != "" {
		eid, err := uuid.Parse(encounterID)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid encounter_id")
		}
		items, total, err := h.svc.ListAssessmentsByEncounter(c.Request().Context(), eid, pg.Limit, pg.Offset)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
	}
	items, total, err := h.svc.ListAssessmentsByPatient(c.Request().Context(), uuid.Nil, pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *Handler) UpdateAssessment(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var a NursingAssessment
	if err := c.Bind(&a); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	a.ID = id
	if err := h.svc.UpdateAssessment(c.Request().Context(), &a); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, a)
}

func (h *Handler) DeleteAssessment(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeleteAssessment(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

// -- Fall Risk Assessment Handlers --

func (h *Handler) CreateFallRisk(c echo.Context) error {
	var a FallRiskAssessment
	if err := c.Bind(&a); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateFallRisk(c.Request().Context(), &a); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, a)
}

func (h *Handler) GetFallRisk(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	a, err := h.svc.GetFallRisk(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "fall risk assessment not found")
	}
	return c.JSON(http.StatusOK, a)
}

func (h *Handler) ListFallRisk(c echo.Context) error {
	pg := pagination.FromContext(c)
	patientID := c.QueryParam("patient_id")
	if patientID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "patient_id is required")
	}
	pid, err := uuid.Parse(patientID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid patient_id")
	}
	items, total, err := h.svc.ListFallRiskByPatient(c.Request().Context(), pid, pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

// -- Skin Assessment Handlers --

func (h *Handler) CreateSkinAssessment(c echo.Context) error {
	var a SkinAssessment
	if err := c.Bind(&a); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateSkinAssessment(c.Request().Context(), &a); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, a)
}

func (h *Handler) GetSkinAssessment(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	a, err := h.svc.GetSkinAssessment(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "skin assessment not found")
	}
	return c.JSON(http.StatusOK, a)
}

func (h *Handler) ListSkinAssessments(c echo.Context) error {
	pg := pagination.FromContext(c)
	patientID := c.QueryParam("patient_id")
	if patientID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "patient_id is required")
	}
	pid, err := uuid.Parse(patientID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid patient_id")
	}
	items, total, err := h.svc.ListSkinAssessmentsByPatient(c.Request().Context(), pid, pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

// -- Pain Assessment Handlers --

func (h *Handler) CreatePainAssessment(c echo.Context) error {
	var a PainAssessment
	if err := c.Bind(&a); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreatePainAssessment(c.Request().Context(), &a); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, a)
}

func (h *Handler) GetPainAssessment(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	a, err := h.svc.GetPainAssessment(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "pain assessment not found")
	}
	return c.JSON(http.StatusOK, a)
}

func (h *Handler) ListPainAssessments(c echo.Context) error {
	pg := pagination.FromContext(c)
	patientID := c.QueryParam("patient_id")
	if patientID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "patient_id is required")
	}
	pid, err := uuid.Parse(patientID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid patient_id")
	}
	items, total, err := h.svc.ListPainAssessmentsByPatient(c.Request().Context(), pid, pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

// -- Lines/Drains/Airways Handlers --

func (h *Handler) CreateLinesDrains(c echo.Context) error {
	var l LinesDrainsAirways
	if err := c.Bind(&l); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateLinesDrains(c.Request().Context(), &l); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, l)
}

func (h *Handler) GetLinesDrains(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	l, err := h.svc.GetLinesDrains(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "lines/drains record not found")
	}
	return c.JSON(http.StatusOK, l)
}

func (h *Handler) ListLinesDrains(c echo.Context) error {
	pg := pagination.FromContext(c)
	if patientID := c.QueryParam("patient_id"); patientID != "" {
		pid, err := uuid.Parse(patientID)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid patient_id")
		}
		items, total, err := h.svc.ListLinesDrainsByPatient(c.Request().Context(), pid, pg.Limit, pg.Offset)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
	}
	if encounterID := c.QueryParam("encounter_id"); encounterID != "" {
		eid, err := uuid.Parse(encounterID)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid encounter_id")
		}
		items, total, err := h.svc.ListLinesDrainsByEncounter(c.Request().Context(), eid, pg.Limit, pg.Offset)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
	}
	items, total, err := h.svc.ListLinesDrainsByPatient(c.Request().Context(), uuid.Nil, pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *Handler) UpdateLinesDrains(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var l LinesDrainsAirways
	if err := c.Bind(&l); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	l.ID = id
	if err := h.svc.UpdateLinesDrains(c.Request().Context(), &l); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, l)
}

func (h *Handler) DeleteLinesDrains(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeleteLinesDrains(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

// -- Restraint Handlers --

func (h *Handler) CreateRestraint(c echo.Context) error {
	var r RestraintRecord
	if err := c.Bind(&r); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateRestraint(c.Request().Context(), &r); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, r)
}

func (h *Handler) GetRestraint(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	r, err := h.svc.GetRestraint(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "restraint record not found")
	}
	return c.JSON(http.StatusOK, r)
}

func (h *Handler) ListRestraints(c echo.Context) error {
	pg := pagination.FromContext(c)
	patientID := c.QueryParam("patient_id")
	if patientID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "patient_id is required")
	}
	pid, err := uuid.Parse(patientID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid patient_id")
	}
	items, total, err := h.svc.ListRestraintsByPatient(c.Request().Context(), pid, pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *Handler) UpdateRestraint(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var r RestraintRecord
	if err := c.Bind(&r); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	r.ID = id
	if err := h.svc.UpdateRestraint(c.Request().Context(), &r); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, r)
}

// -- Intake/Output Handlers --

func (h *Handler) CreateIntakeOutput(c echo.Context) error {
	var r IntakeOutputRecord
	if err := c.Bind(&r); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateIntakeOutput(c.Request().Context(), &r); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, r)
}

func (h *Handler) GetIntakeOutput(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	r, err := h.svc.GetIntakeOutput(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "intake/output record not found")
	}
	return c.JSON(http.StatusOK, r)
}

func (h *Handler) ListIntakeOutput(c echo.Context) error {
	pg := pagination.FromContext(c)
	if patientID := c.QueryParam("patient_id"); patientID != "" {
		pid, err := uuid.Parse(patientID)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid patient_id")
		}
		items, total, err := h.svc.ListIntakeOutputByPatient(c.Request().Context(), pid, pg.Limit, pg.Offset)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
	}
	if encounterID := c.QueryParam("encounter_id"); encounterID != "" {
		eid, err := uuid.Parse(encounterID)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid encounter_id")
		}
		items, total, err := h.svc.ListIntakeOutputByEncounter(c.Request().Context(), eid, pg.Limit, pg.Offset)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
	}
	items, total, err := h.svc.ListIntakeOutputByPatient(c.Request().Context(), uuid.Nil, pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *Handler) DeleteIntakeOutput(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeleteIntakeOutput(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}
