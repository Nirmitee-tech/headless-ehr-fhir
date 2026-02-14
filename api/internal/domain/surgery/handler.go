package surgery

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
	// Read endpoints – admin, physician, surgeon, nurse
	readGroup := api.Group("", auth.RequireRole("admin", "physician", "surgeon", "nurse"))
	readGroup.GET("/or-rooms", h.ListORRooms)
	readGroup.GET("/or-rooms/:id", h.GetORRoom)
	readGroup.GET("/surgical-cases", h.ListSurgicalCases)
	readGroup.GET("/surgical-cases/:id", h.GetSurgicalCase)
	readGroup.GET("/surgical-cases/:id/procedures", h.GetCaseProcedures)
	readGroup.GET("/surgical-cases/:id/team", h.GetCaseTeamMembers)
	readGroup.GET("/surgical-cases/:id/time-events", h.GetCaseTimeEvents)
	readGroup.GET("/surgical-cases/:id/counts", h.GetCaseCounts)
	readGroup.GET("/surgical-cases/:id/supplies", h.GetCaseSupplies)
	readGroup.GET("/preference-cards", h.ListPreferenceCards)
	readGroup.GET("/preference-cards/:id", h.GetPreferenceCard)
	readGroup.GET("/implant-logs", h.ListImplantLogs)
	readGroup.GET("/implant-logs/:id", h.GetImplantLog)

	// Write endpoints – admin, physician, surgeon
	writeGroup := api.Group("", auth.RequireRole("admin", "physician", "surgeon"))
	writeGroup.POST("/or-rooms", h.CreateORRoom)
	writeGroup.PUT("/or-rooms/:id", h.UpdateORRoom)
	writeGroup.DELETE("/or-rooms/:id", h.DeleteORRoom)
	writeGroup.POST("/surgical-cases", h.CreateSurgicalCase)
	writeGroup.PUT("/surgical-cases/:id", h.UpdateSurgicalCase)
	writeGroup.DELETE("/surgical-cases/:id", h.DeleteSurgicalCase)
	writeGroup.POST("/surgical-cases/:id/procedures", h.AddCaseProcedure)
	writeGroup.POST("/surgical-cases/:id/team", h.AddCaseTeamMember)
	writeGroup.POST("/surgical-cases/:id/time-events", h.AddCaseTimeEvent)
	writeGroup.POST("/surgical-cases/:id/counts", h.AddCaseCount)
	writeGroup.POST("/surgical-cases/:id/supplies", h.AddCaseSupply)
	writeGroup.POST("/preference-cards", h.CreatePreferenceCard)
	writeGroup.PUT("/preference-cards/:id", h.UpdatePreferenceCard)
	writeGroup.DELETE("/preference-cards/:id", h.DeletePreferenceCard)
	writeGroup.POST("/implant-logs", h.CreateImplantLog)
	writeGroup.PUT("/implant-logs/:id", h.UpdateImplantLog)
	writeGroup.DELETE("/implant-logs/:id", h.DeleteImplantLog)
}

// -- OR Room Handlers --

func (h *Handler) CreateORRoom(c echo.Context) error {
	var r ORRoom
	if err := c.Bind(&r); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateORRoom(c.Request().Context(), &r); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, r)
}

func (h *Handler) GetORRoom(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	r, err := h.svc.GetORRoom(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "or room not found")
	}
	return c.JSON(http.StatusOK, r)
}

func (h *Handler) ListORRooms(c echo.Context) error {
	pg := pagination.FromContext(c)
	items, total, err := h.svc.SearchORRooms(c.Request().Context(), nil, pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *Handler) UpdateORRoom(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var r ORRoom
	if err := c.Bind(&r); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	r.ID = id
	if err := h.svc.UpdateORRoom(c.Request().Context(), &r); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, r)
}

func (h *Handler) DeleteORRoom(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeleteORRoom(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

// -- Surgical Case Handlers --

func (h *Handler) CreateSurgicalCase(c echo.Context) error {
	var sc SurgicalCase
	if err := c.Bind(&sc); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateSurgicalCase(c.Request().Context(), &sc); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, sc)
}

func (h *Handler) GetSurgicalCase(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	sc, err := h.svc.GetSurgicalCase(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "surgical case not found")
	}
	return c.JSON(http.StatusOK, sc)
}

func (h *Handler) ListSurgicalCases(c echo.Context) error {
	pg := pagination.FromContext(c)
	if patientID := c.QueryParam("patient_id"); patientID != "" {
		pid, err := uuid.Parse(patientID)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid patient_id")
		}
		items, total, err := h.svc.ListSurgicalCasesByPatient(c.Request().Context(), pid, pg.Limit, pg.Offset)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
	}
	items, total, err := h.svc.SearchSurgicalCases(c.Request().Context(), nil, pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *Handler) UpdateSurgicalCase(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var sc SurgicalCase
	if err := c.Bind(&sc); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	sc.ID = id
	if err := h.svc.UpdateSurgicalCase(c.Request().Context(), &sc); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, sc)
}

func (h *Handler) DeleteSurgicalCase(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeleteSurgicalCase(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

// -- Surgical Case Sub-Resource Handlers --

func (h *Handler) AddCaseProcedure(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var p SurgicalCaseProcedure
	if err := c.Bind(&p); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	p.SurgicalCaseID = id
	if err := h.svc.AddCaseProcedure(c.Request().Context(), &p); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, p)
}

func (h *Handler) GetCaseProcedures(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	items, err := h.svc.GetCaseProcedures(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, items)
}

func (h *Handler) AddCaseTeamMember(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var t SurgicalCaseTeam
	if err := c.Bind(&t); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	t.SurgicalCaseID = id
	if err := h.svc.AddCaseTeamMember(c.Request().Context(), &t); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, t)
}

func (h *Handler) GetCaseTeamMembers(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	items, err := h.svc.GetCaseTeamMembers(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, items)
}

func (h *Handler) AddCaseTimeEvent(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var e SurgicalTimeEvent
	if err := c.Bind(&e); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	e.SurgicalCaseID = id
	if err := h.svc.AddCaseTimeEvent(c.Request().Context(), &e); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, e)
}

func (h *Handler) GetCaseTimeEvents(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	items, err := h.svc.GetCaseTimeEvents(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, items)
}

func (h *Handler) AddCaseCount(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var cnt SurgicalCount
	if err := c.Bind(&cnt); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	cnt.SurgicalCaseID = id
	if err := h.svc.AddCaseCount(c.Request().Context(), &cnt); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, cnt)
}

func (h *Handler) GetCaseCounts(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	items, err := h.svc.GetCaseCounts(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, items)
}

func (h *Handler) AddCaseSupply(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var su SurgicalSupplyUsed
	if err := c.Bind(&su); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	su.SurgicalCaseID = id
	if err := h.svc.AddCaseSupply(c.Request().Context(), &su); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, su)
}

func (h *Handler) GetCaseSupplies(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	items, err := h.svc.GetCaseSupplies(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, items)
}

// -- Preference Card Handlers --

func (h *Handler) CreatePreferenceCard(c echo.Context) error {
	var pc SurgicalPreferenceCard
	if err := c.Bind(&pc); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreatePreferenceCard(c.Request().Context(), &pc); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, pc)
}

func (h *Handler) GetPreferenceCard(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	pc, err := h.svc.GetPreferenceCard(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "preference card not found")
	}
	return c.JSON(http.StatusOK, pc)
}

func (h *Handler) ListPreferenceCards(c echo.Context) error {
	pg := pagination.FromContext(c)
	if surgeonID := c.QueryParam("surgeon_id"); surgeonID != "" {
		sid, err := uuid.Parse(surgeonID)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid surgeon_id")
		}
		items, total, err := h.svc.ListPreferenceCardsBySurgeon(c.Request().Context(), sid, pg.Limit, pg.Offset)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
	}
	items, total, err := h.svc.SearchPreferenceCards(c.Request().Context(), nil, pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *Handler) UpdatePreferenceCard(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var pc SurgicalPreferenceCard
	if err := c.Bind(&pc); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	pc.ID = id
	if err := h.svc.UpdatePreferenceCard(c.Request().Context(), &pc); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, pc)
}

func (h *Handler) DeletePreferenceCard(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeletePreferenceCard(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

// -- Implant Log Handlers --

func (h *Handler) CreateImplantLog(c echo.Context) error {
	var il ImplantLog
	if err := c.Bind(&il); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateImplantLog(c.Request().Context(), &il); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, il)
}

func (h *Handler) GetImplantLog(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	il, err := h.svc.GetImplantLog(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "implant log not found")
	}
	return c.JSON(http.StatusOK, il)
}

func (h *Handler) ListImplantLogs(c echo.Context) error {
	pg := pagination.FromContext(c)
	if patientID := c.QueryParam("patient_id"); patientID != "" {
		pid, err := uuid.Parse(patientID)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid patient_id")
		}
		items, total, err := h.svc.ListImplantLogsByPatient(c.Request().Context(), pid, pg.Limit, pg.Offset)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
	}
	items, total, err := h.svc.SearchImplantLogs(c.Request().Context(), nil, pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *Handler) UpdateImplantLog(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var il ImplantLog
	if err := c.Bind(&il); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	il.ID = id
	if err := h.svc.UpdateImplantLog(c.Request().Context(), &il); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, il)
}

func (h *Handler) DeleteImplantLog(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeleteImplantLog(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}
