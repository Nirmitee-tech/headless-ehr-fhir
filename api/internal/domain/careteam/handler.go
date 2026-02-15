package careteam

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
	role := auth.RequireRole("admin", "physician", "nurse")

	read := api.Group("", role)
	read.GET("/care-teams", h.ListCareTeams)
	read.GET("/care-teams/:id", h.GetCareTeam)
	read.GET("/care-teams/:id/participants", h.GetParticipants)

	write := api.Group("", role)
	write.POST("/care-teams", h.CreateCareTeam)
	write.PUT("/care-teams/:id", h.UpdateCareTeam)
	write.DELETE("/care-teams/:id", h.DeleteCareTeam)
	write.POST("/care-teams/:id/participants", h.AddParticipant)
	write.DELETE("/care-teams/:id/participants/:pid", h.RemoveParticipant)

	fhirRead := fhirGroup.Group("", role)
	fhirRead.GET("/CareTeam", h.SearchCareTeamsFHIR)
	fhirRead.GET("/CareTeam/:id", h.GetCareTeamFHIR)

	fhirWrite := fhirGroup.Group("", role)
	fhirWrite.POST("/CareTeam", h.CreateCareTeamFHIR)
	fhirWrite.PUT("/CareTeam/:id", h.UpdateCareTeamFHIR)
	fhirWrite.DELETE("/CareTeam/:id", h.DeleteCareTeamFHIR)
	fhirWrite.PATCH("/CareTeam/:id", h.PatchCareTeamFHIR)

	fhirRead.POST("/CareTeam/_search", h.SearchCareTeamsFHIR)

	fhirRead.GET("/CareTeam/:id/_history/:vid", h.VreadCareTeamFHIR)
	fhirRead.GET("/CareTeam/:id/_history", h.HistoryCareTeamFHIR)
}

// -- REST handlers --

func (h *Handler) CreateCareTeam(c echo.Context) error {
	var ct CareTeam
	if err := c.Bind(&ct); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateCareTeam(c.Request().Context(), &ct); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, ct)
}

func (h *Handler) GetCareTeam(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	ct, err := h.svc.GetCareTeam(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "care team not found")
	}
	return c.JSON(http.StatusOK, ct)
}

func (h *Handler) ListCareTeams(c echo.Context) error {
	pg := pagination.FromContext(c)
	if patientID := c.QueryParam("patient_id"); patientID != "" {
		pid, err := uuid.Parse(patientID)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid patient_id")
		}
		items, total, err := h.svc.ListCareTeamsByPatient(c.Request().Context(), pid, pg.Limit, pg.Offset)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
	}
	items, total, err := h.svc.SearchCareTeams(c.Request().Context(), nil, pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *Handler) UpdateCareTeam(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var ct CareTeam
	if err := c.Bind(&ct); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	ct.ID = id
	if err := h.svc.UpdateCareTeam(c.Request().Context(), &ct); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, ct)
}

func (h *Handler) DeleteCareTeam(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeleteCareTeam(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) AddParticipant(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var p CareTeamParticipant
	if err := c.Bind(&p); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.AddParticipant(c.Request().Context(), id, &p); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, p)
}

func (h *Handler) RemoveParticipant(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	pid, err := uuid.Parse(c.Param("pid"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid participant id")
	}
	if err := h.svc.RemoveParticipant(c.Request().Context(), id, pid); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) GetParticipants(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	participants, err := h.svc.GetParticipants(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, participants)
}

// -- FHIR handlers --

func (h *Handler) SearchCareTeamsFHIR(c echo.Context) error {
	pg := pagination.FromContext(c)
	params := map[string]string{}
	for _, k := range []string{"patient", "status", "category"} {
		if v := c.QueryParam(k); v != "" {
			params[k] = v
		}
	}
	items, total, err := h.svc.SearchCareTeams(c.Request().Context(), params, pg.Limit, pg.Offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	resources := make([]interface{}, len(items))
	for i, item := range items {
		resources[i] = item.ToFHIR()
	}
	return c.JSON(http.StatusOK, fhir.NewSearchBundle(resources, total, "/fhir/CareTeam"))
}

func (h *Handler) GetCareTeamFHIR(c echo.Context) error {
	ct, err := h.svc.GetCareTeamByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("CareTeam", c.Param("id")))
	}
	return c.JSON(http.StatusOK, ct.ToFHIR())
}

func (h *Handler) CreateCareTeamFHIR(c echo.Context) error {
	var ct CareTeam
	if err := c.Bind(&ct); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	if err := h.svc.CreateCareTeam(c.Request().Context(), &ct); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	c.Response().Header().Set("Location", "/fhir/CareTeam/"+ct.FHIRID)
	return c.JSON(http.StatusCreated, ct.ToFHIR())
}

func (h *Handler) UpdateCareTeamFHIR(c echo.Context) error {
	var ct CareTeam
	if err := c.Bind(&ct); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	existing, err := h.svc.GetCareTeamByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("CareTeam", c.Param("id")))
	}
	ct.ID = existing.ID
	ct.FHIRID = existing.FHIRID
	if err := h.svc.UpdateCareTeam(c.Request().Context(), &ct); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, ct.ToFHIR())
}

func (h *Handler) DeleteCareTeamFHIR(c echo.Context) error {
	existing, err := h.svc.GetCareTeamByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("CareTeam", c.Param("id")))
	}
	if err := h.svc.DeleteCareTeam(c.Request().Context(), existing.ID); err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) PatchCareTeamFHIR(c echo.Context) error {
	return h.handlePatch(c, "CareTeam", c.Param("id"), func(ctx echo.Context, resource map[string]interface{}) error {
		existing, err := h.svc.GetCareTeamByFHIRID(ctx.Request().Context(), ctx.Param("id"))
		if err != nil {
			return ctx.JSON(http.StatusNotFound, fhir.NotFoundOutcome("CareTeam", ctx.Param("id")))
		}
		if v, ok := resource["status"].(string); ok {
			existing.Status = v
		}
		if err := h.svc.UpdateCareTeam(ctx.Request().Context(), existing); err != nil {
			return ctx.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
		}
		return ctx.JSON(http.StatusOK, existing.ToFHIR())
	})
}

func (h *Handler) VreadCareTeamFHIR(c echo.Context) error {
	ct, err := h.svc.GetCareTeamByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("CareTeam", c.Param("id")))
	}
	result := ct.ToFHIR()
	fhir.SetVersionHeaders(c, 1, ct.UpdatedAt.Format("2006-01-02T15:04:05Z"))
	return c.JSON(http.StatusOK, result)
}

func (h *Handler) HistoryCareTeamFHIR(c echo.Context) error {
	ct, err := h.svc.GetCareTeamByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("CareTeam", c.Param("id")))
	}
	result := ct.ToFHIR()
	raw, _ := json.Marshal(result)
	entry := &fhir.HistoryEntry{
		ResourceType: "CareTeam", ResourceID: ct.FHIRID, VersionID: 1,
		Resource: raw, Action: "create", Timestamp: ct.CreatedAt,
	}
	return c.JSON(http.StatusOK, fhir.NewHistoryBundle([]*fhir.HistoryEntry{entry}, 1, "/fhir"))
}

func (h *Handler) handlePatch(c echo.Context, resourceType, fhirID string, applyFn func(echo.Context, map[string]interface{}) error) error {
	contentType := c.Request().Header.Get("Content-Type")
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome("failed to read request body"))
	}

	existing, err := h.svc.GetCareTeamByFHIRID(c.Request().Context(), fhirID)
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome(resourceType, fhirID))
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
		var mergePatch map[string]interface{}
		if err := json.Unmarshal(body, &mergePatch); err != nil {
			return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome("invalid merge patch JSON: "+err.Error()))
		}
		patched, err = fhir.ApplyMergePatch(currentResource, mergePatch)
		if err != nil {
			return c.JSON(http.StatusUnprocessableEntity, fhir.ErrorOutcome(err.Error()))
		}
	} else {
		return c.JSON(http.StatusUnsupportedMediaType, fhir.ErrorOutcome(
			"PATCH requires Content-Type: application/json-patch+json or application/merge-patch+json"))
	}

	return applyFn(c, patched)
}
