package behavioral

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

func (h *Handler) RegisterRoutes(api *echo.Group, _ *echo.Group) {
	// Read endpoints – admin, physician, nurse
	readGroup := api.Group("", auth.RequireRole("admin", "physician", "nurse"))
	readGroup.GET("/psychiatric-assessments", h.ListPsychAssessments)
	readGroup.GET("/psychiatric-assessments/:id", h.GetPsychAssessment)
	readGroup.GET("/safety-plans", h.ListSafetyPlans)
	readGroup.GET("/safety-plans/:id", h.GetSafetyPlan)
	readGroup.GET("/legal-holds", h.ListLegalHolds)
	readGroup.GET("/legal-holds/:id", h.GetLegalHold)
	readGroup.GET("/seclusion-restraints", h.ListSeclusionRestraints)
	readGroup.GET("/seclusion-restraints/:id", h.GetSeclusionRestraint)
	readGroup.GET("/group-therapy-sessions", h.ListGroupTherapySessions)
	readGroup.GET("/group-therapy-sessions/:id", h.GetGroupTherapySession)
	readGroup.GET("/group-therapy-sessions/:id/attendance", h.GetAttendance)

	// Write endpoints – admin, physician, nurse
	writeGroup := api.Group("", auth.RequireRole("admin", "physician", "nurse"))
	writeGroup.POST("/psychiatric-assessments", h.CreatePsychAssessment)
	writeGroup.PUT("/psychiatric-assessments/:id", h.UpdatePsychAssessment)
	writeGroup.DELETE("/psychiatric-assessments/:id", h.DeletePsychAssessment)
	writeGroup.POST("/safety-plans", h.CreateSafetyPlan)
	writeGroup.PUT("/safety-plans/:id", h.UpdateSafetyPlan)
	writeGroup.DELETE("/safety-plans/:id", h.DeleteSafetyPlan)
	writeGroup.POST("/legal-holds", h.CreateLegalHold)
	writeGroup.PUT("/legal-holds/:id", h.UpdateLegalHold)
	writeGroup.DELETE("/legal-holds/:id", h.DeleteLegalHold)
	writeGroup.POST("/seclusion-restraints", h.CreateSeclusionRestraint)
	writeGroup.PUT("/seclusion-restraints/:id", h.UpdateSeclusionRestraint)
	writeGroup.DELETE("/seclusion-restraints/:id", h.DeleteSeclusionRestraint)
	writeGroup.POST("/group-therapy-sessions", h.CreateGroupTherapySession)
	writeGroup.PUT("/group-therapy-sessions/:id", h.UpdateGroupTherapySession)
	writeGroup.DELETE("/group-therapy-sessions/:id", h.DeleteGroupTherapySession)
	writeGroup.POST("/group-therapy-sessions/:id/attendance", h.AddAttendance)
}

// -- Psychiatric Assessment Handlers --

func (h *Handler) CreatePsychAssessment(c echo.Context) error {
	var a PsychiatricAssessment
	if err := c.Bind(&a); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreatePsychAssessment(c.Request().Context(), &a); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, a)
}

func (h *Handler) GetPsychAssessment(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	a, err := h.svc.GetPsychAssessment(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "psychiatric assessment not found")
	}
	return c.JSON(http.StatusOK, a)
}

func (h *Handler) ListPsychAssessments(c echo.Context) error {
	pg := pagination.FromContext(c)
	if patientID := c.QueryParam("patient_id"); patientID != "" {
		pid, err := uuid.Parse(patientID)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid patient_id")
		}
		items, total, err := h.svc.ListPsychAssessmentsByPatient(c.Request().Context(), pid, pg.Limit, pg.Offset)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
	}
	items, total, err := h.svc.SearchPsychAssessments(c.Request().Context(), nil, pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *Handler) UpdatePsychAssessment(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var a PsychiatricAssessment
	if err := c.Bind(&a); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	a.ID = id
	if err := h.svc.UpdatePsychAssessment(c.Request().Context(), &a); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, a)
}

func (h *Handler) DeletePsychAssessment(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeletePsychAssessment(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

// -- Safety Plan Handlers --

func (h *Handler) CreateSafetyPlan(c echo.Context) error {
	var sp SafetyPlan
	if err := c.Bind(&sp); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateSafetyPlan(c.Request().Context(), &sp); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, sp)
}

func (h *Handler) GetSafetyPlan(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	sp, err := h.svc.GetSafetyPlan(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "safety plan not found")
	}
	return c.JSON(http.StatusOK, sp)
}

func (h *Handler) ListSafetyPlans(c echo.Context) error {
	pg := pagination.FromContext(c)
	if patientID := c.QueryParam("patient_id"); patientID != "" {
		pid, err := uuid.Parse(patientID)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid patient_id")
		}
		items, total, err := h.svc.ListSafetyPlansByPatient(c.Request().Context(), pid, pg.Limit, pg.Offset)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
	}
	items, total, err := h.svc.SearchSafetyPlans(c.Request().Context(), nil, pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *Handler) UpdateSafetyPlan(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var sp SafetyPlan
	if err := c.Bind(&sp); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	sp.ID = id
	if err := h.svc.UpdateSafetyPlan(c.Request().Context(), &sp); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, sp)
}

func (h *Handler) DeleteSafetyPlan(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeleteSafetyPlan(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

// -- Legal Hold Handlers --

func (h *Handler) CreateLegalHold(c echo.Context) error {
	var lh LegalHold
	if err := c.Bind(&lh); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateLegalHold(c.Request().Context(), &lh); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, lh)
}

func (h *Handler) GetLegalHold(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	lh, err := h.svc.GetLegalHold(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "legal hold not found")
	}
	return c.JSON(http.StatusOK, lh)
}

func (h *Handler) ListLegalHolds(c echo.Context) error {
	pg := pagination.FromContext(c)
	if patientID := c.QueryParam("patient_id"); patientID != "" {
		pid, err := uuid.Parse(patientID)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid patient_id")
		}
		items, total, err := h.svc.ListLegalHoldsByPatient(c.Request().Context(), pid, pg.Limit, pg.Offset)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
	}
	items, total, err := h.svc.SearchLegalHolds(c.Request().Context(), nil, pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *Handler) UpdateLegalHold(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var lh LegalHold
	if err := c.Bind(&lh); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	lh.ID = id
	if err := h.svc.UpdateLegalHold(c.Request().Context(), &lh); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, lh)
}

func (h *Handler) DeleteLegalHold(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeleteLegalHold(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

// -- Seclusion/Restraint Handlers --

func (h *Handler) CreateSeclusionRestraint(c echo.Context) error {
	var e SeclusionRestraintEvent
	if err := c.Bind(&e); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateSeclusionRestraint(c.Request().Context(), &e); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, e)
}

func (h *Handler) GetSeclusionRestraint(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	e, err := h.svc.GetSeclusionRestraint(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "seclusion/restraint event not found")
	}
	return c.JSON(http.StatusOK, e)
}

func (h *Handler) ListSeclusionRestraints(c echo.Context) error {
	pg := pagination.FromContext(c)
	if patientID := c.QueryParam("patient_id"); patientID != "" {
		pid, err := uuid.Parse(patientID)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid patient_id")
		}
		items, total, err := h.svc.ListSeclusionRestraintsByPatient(c.Request().Context(), pid, pg.Limit, pg.Offset)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
	}
	items, total, err := h.svc.SearchSeclusionRestraints(c.Request().Context(), nil, pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *Handler) UpdateSeclusionRestraint(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var e SeclusionRestraintEvent
	if err := c.Bind(&e); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	e.ID = id
	if err := h.svc.UpdateSeclusionRestraint(c.Request().Context(), &e); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, e)
}

func (h *Handler) DeleteSeclusionRestraint(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeleteSeclusionRestraint(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

// -- Group Therapy Session Handlers --

func (h *Handler) CreateGroupTherapySession(c echo.Context) error {
	var gs GroupTherapySession
	if err := c.Bind(&gs); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateGroupTherapySession(c.Request().Context(), &gs); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, gs)
}

func (h *Handler) GetGroupTherapySession(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	gs, err := h.svc.GetGroupTherapySession(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "group therapy session not found")
	}
	return c.JSON(http.StatusOK, gs)
}

func (h *Handler) ListGroupTherapySessions(c echo.Context) error {
	pg := pagination.FromContext(c)
	items, total, err := h.svc.ListGroupTherapySessions(c.Request().Context(), pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *Handler) UpdateGroupTherapySession(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var gs GroupTherapySession
	if err := c.Bind(&gs); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	gs.ID = id
	if err := h.svc.UpdateGroupTherapySession(c.Request().Context(), &gs); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, gs)
}

func (h *Handler) DeleteGroupTherapySession(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeleteGroupTherapySession(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) AddAttendance(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var a GroupTherapyAttendance
	if err := c.Bind(&a); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	a.SessionID = id
	if err := h.svc.AddGroupTherapyAttendance(c.Request().Context(), &a); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, a)
}

func (h *Handler) GetAttendance(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	items, err := h.svc.GetGroupTherapyAttendance(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, items)
}
