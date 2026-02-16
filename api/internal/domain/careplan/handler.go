package careplan

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
	read.GET("/care-plans", h.ListCarePlans)
	read.GET("/care-plans/:id", h.GetCarePlan)
	read.GET("/care-plans/:id/activities", h.GetActivities)
	read.GET("/goals", h.ListGoals)
	read.GET("/goals/:id", h.GetGoal)

	write := api.Group("", role)
	write.POST("/care-plans", h.CreateCarePlan)
	write.PUT("/care-plans/:id", h.UpdateCarePlan)
	write.DELETE("/care-plans/:id", h.DeleteCarePlan)
	write.POST("/care-plans/:id/activities", h.AddActivity)
	write.POST("/goals", h.CreateGoal)
	write.PUT("/goals/:id", h.UpdateGoal)
	write.DELETE("/goals/:id", h.DeleteGoal)

	fhirRead := fhirGroup.Group("", role)
	fhirRead.GET("/CarePlan", h.SearchCarePlansFHIR)
	fhirRead.GET("/CarePlan/:id", h.GetCarePlanFHIR)
	fhirRead.GET("/Goal", h.SearchGoalsFHIR)
	fhirRead.GET("/Goal/:id", h.GetGoalFHIR)

	fhirWrite := fhirGroup.Group("", role)
	fhirWrite.POST("/CarePlan", h.CreateCarePlanFHIR)
	fhirWrite.PUT("/CarePlan/:id", h.UpdateCarePlanFHIR)
	fhirWrite.DELETE("/CarePlan/:id", h.DeleteCarePlanFHIR)
	fhirWrite.PATCH("/CarePlan/:id", h.PatchCarePlanFHIR)
	fhirWrite.POST("/Goal", h.CreateGoalFHIR)
	fhirWrite.PUT("/Goal/:id", h.UpdateGoalFHIR)
	fhirWrite.DELETE("/Goal/:id", h.DeleteGoalFHIR)
	fhirWrite.PATCH("/Goal/:id", h.PatchGoalFHIR)

	fhirRead.POST("/CarePlan/_search", h.SearchCarePlansFHIR)
	fhirRead.POST("/Goal/_search", h.SearchGoalsFHIR)

	fhirRead.GET("/CarePlan/:id/_history/:vid", h.VreadCarePlanFHIR)
	fhirRead.GET("/CarePlan/:id/_history", h.HistoryCarePlanFHIR)
	fhirRead.GET("/Goal/:id/_history/:vid", h.VreadGoalFHIR)
	fhirRead.GET("/Goal/:id/_history", h.HistoryGoalFHIR)
}

// -- CarePlan REST --

func (h *Handler) CreateCarePlan(c echo.Context) error {
	var cp CarePlan
	if err := c.Bind(&cp); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateCarePlan(c.Request().Context(), &cp); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, cp)
}

func (h *Handler) GetCarePlan(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	cp, err := h.svc.GetCarePlan(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "care plan not found")
	}
	return c.JSON(http.StatusOK, cp)
}

func (h *Handler) ListCarePlans(c echo.Context) error {
	pg := pagination.FromContext(c)
	if patientID := c.QueryParam("patient_id"); patientID != "" {
		pid, err := uuid.Parse(patientID)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid patient_id")
		}
		items, total, err := h.svc.ListCarePlansByPatient(c.Request().Context(), pid, pg.Limit, pg.Offset)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
	}
	items, total, err := h.svc.SearchCarePlans(c.Request().Context(), nil, pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *Handler) UpdateCarePlan(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var cp CarePlan
	if err := c.Bind(&cp); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	cp.ID = id
	if err := h.svc.UpdateCarePlan(c.Request().Context(), &cp); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, cp)
}

func (h *Handler) DeleteCarePlan(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeleteCarePlan(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) AddActivity(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var a CarePlanActivity
	if err := c.Bind(&a); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	a.CarePlanID = id
	if err := h.svc.AddActivity(c.Request().Context(), &a); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, a)
}

func (h *Handler) GetActivities(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	activities, err := h.svc.GetActivities(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, activities)
}

// -- Goal REST --

func (h *Handler) CreateGoal(c echo.Context) error {
	var g Goal
	if err := c.Bind(&g); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateGoal(c.Request().Context(), &g); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, g)
}

func (h *Handler) GetGoal(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	g, err := h.svc.GetGoal(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "goal not found")
	}
	return c.JSON(http.StatusOK, g)
}

func (h *Handler) ListGoals(c echo.Context) error {
	pg := pagination.FromContext(c)
	if patientID := c.QueryParam("patient_id"); patientID != "" {
		pid, err := uuid.Parse(patientID)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid patient_id")
		}
		items, total, err := h.svc.ListGoalsByPatient(c.Request().Context(), pid, pg.Limit, pg.Offset)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
	}
	items, total, err := h.svc.SearchGoals(c.Request().Context(), nil, pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *Handler) UpdateGoal(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var g Goal
	if err := c.Bind(&g); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	g.ID = id
	if err := h.svc.UpdateGoal(c.Request().Context(), &g); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, g)
}

func (h *Handler) DeleteGoal(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeleteGoal(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

// -- FHIR CarePlan --

func (h *Handler) SearchCarePlansFHIR(c echo.Context) error {
	pg := pagination.FromContext(c)
	params := fhir.ExtractSearchParams(c)
	items, total, err := h.svc.SearchCarePlans(c.Request().Context(), params, pg.Limit, pg.Offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	resources := make([]interface{}, len(items))
	for i, item := range items {
		resources[i] = item.ToFHIR()
	}
	return c.JSON(http.StatusOK, fhir.NewSearchBundle(resources, total, "/fhir/CarePlan"))
}

func (h *Handler) GetCarePlanFHIR(c echo.Context) error {
	cp, err := h.svc.GetCarePlanByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("CarePlan", c.Param("id")))
	}
	return c.JSON(http.StatusOK, cp.ToFHIR())
}

func (h *Handler) CreateCarePlanFHIR(c echo.Context) error {
	var cp CarePlan
	if err := c.Bind(&cp); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	if err := h.svc.CreateCarePlan(c.Request().Context(), &cp); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	c.Response().Header().Set("Location", "/fhir/CarePlan/"+cp.FHIRID)
	return c.JSON(http.StatusCreated, cp.ToFHIR())
}

func (h *Handler) UpdateCarePlanFHIR(c echo.Context) error {
	var cp CarePlan
	if err := c.Bind(&cp); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	existing, err := h.svc.GetCarePlanByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("CarePlan", c.Param("id")))
	}
	cp.ID = existing.ID
	cp.FHIRID = existing.FHIRID
	if err := h.svc.UpdateCarePlan(c.Request().Context(), &cp); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, cp.ToFHIR())
}

func (h *Handler) DeleteCarePlanFHIR(c echo.Context) error {
	existing, err := h.svc.GetCarePlanByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("CarePlan", c.Param("id")))
	}
	if err := h.svc.DeleteCarePlan(c.Request().Context(), existing.ID); err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) PatchCarePlanFHIR(c echo.Context) error {
	return h.handlePatch(c, "CarePlan", c.Param("id"), func(ctx echo.Context, resource map[string]interface{}) error {
		existing, err := h.svc.GetCarePlanByFHIRID(ctx.Request().Context(), ctx.Param("id"))
		if err != nil {
			return ctx.JSON(http.StatusNotFound, fhir.NotFoundOutcome("CarePlan", ctx.Param("id")))
		}
		if v, ok := resource["status"].(string); ok {
			existing.Status = v
		}
		if err := h.svc.UpdateCarePlan(ctx.Request().Context(), existing); err != nil {
			return ctx.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
		}
		return ctx.JSON(http.StatusOK, existing.ToFHIR())
	})
}

func (h *Handler) VreadCarePlanFHIR(c echo.Context) error {
	cp, err := h.svc.GetCarePlanByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("CarePlan", c.Param("id")))
	}
	result := cp.ToFHIR()
	fhir.SetVersionHeaders(c, 1, cp.UpdatedAt.Format("2006-01-02T15:04:05Z"))
	return c.JSON(http.StatusOK, result)
}

func (h *Handler) HistoryCarePlanFHIR(c echo.Context) error {
	cp, err := h.svc.GetCarePlanByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("CarePlan", c.Param("id")))
	}
	result := cp.ToFHIR()
	raw, _ := json.Marshal(result)
	entry := &fhir.HistoryEntry{
		ResourceType: "CarePlan", ResourceID: cp.FHIRID, VersionID: 1,
		Resource: raw, Action: "create", Timestamp: cp.CreatedAt,
	}
	return c.JSON(http.StatusOK, fhir.NewHistoryBundle([]*fhir.HistoryEntry{entry}, 1, "/fhir"))
}

// -- FHIR Goal --

func (h *Handler) SearchGoalsFHIR(c echo.Context) error {
	pg := pagination.FromContext(c)
	params := fhir.ExtractSearchParams(c)
	items, total, err := h.svc.SearchGoals(c.Request().Context(), params, pg.Limit, pg.Offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	resources := make([]interface{}, len(items))
	for i, item := range items {
		resources[i] = item.ToFHIR()
	}
	return c.JSON(http.StatusOK, fhir.NewSearchBundle(resources, total, "/fhir/Goal"))
}

func (h *Handler) GetGoalFHIR(c echo.Context) error {
	g, err := h.svc.GetGoalByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Goal", c.Param("id")))
	}
	return c.JSON(http.StatusOK, g.ToFHIR())
}

func (h *Handler) CreateGoalFHIR(c echo.Context) error {
	var g Goal
	if err := c.Bind(&g); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	if err := h.svc.CreateGoal(c.Request().Context(), &g); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	c.Response().Header().Set("Location", "/fhir/Goal/"+g.FHIRID)
	return c.JSON(http.StatusCreated, g.ToFHIR())
}

func (h *Handler) UpdateGoalFHIR(c echo.Context) error {
	var g Goal
	if err := c.Bind(&g); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	existing, err := h.svc.GetGoalByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Goal", c.Param("id")))
	}
	g.ID = existing.ID
	g.FHIRID = existing.FHIRID
	if err := h.svc.UpdateGoal(c.Request().Context(), &g); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, g.ToFHIR())
}

func (h *Handler) DeleteGoalFHIR(c echo.Context) error {
	existing, err := h.svc.GetGoalByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Goal", c.Param("id")))
	}
	if err := h.svc.DeleteGoal(c.Request().Context(), existing.ID); err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) PatchGoalFHIR(c echo.Context) error {
	return h.handlePatch(c, "Goal", c.Param("id"), func(ctx echo.Context, resource map[string]interface{}) error {
		existing, err := h.svc.GetGoalByFHIRID(ctx.Request().Context(), ctx.Param("id"))
		if err != nil {
			return ctx.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Goal", ctx.Param("id")))
		}
		if v, ok := resource["lifecycleStatus"].(string); ok {
			existing.LifecycleStatus = v
		}
		if err := h.svc.UpdateGoal(ctx.Request().Context(), existing); err != nil {
			return ctx.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
		}
		return ctx.JSON(http.StatusOK, existing.ToFHIR())
	})
}

func (h *Handler) VreadGoalFHIR(c echo.Context) error {
	g, err := h.svc.GetGoalByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Goal", c.Param("id")))
	}
	result := g.ToFHIR()
	fhir.SetVersionHeaders(c, 1, g.UpdatedAt.Format("2006-01-02T15:04:05Z"))
	return c.JSON(http.StatusOK, result)
}

func (h *Handler) HistoryGoalFHIR(c echo.Context) error {
	g, err := h.svc.GetGoalByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Goal", c.Param("id")))
	}
	result := g.ToFHIR()
	raw, _ := json.Marshal(result)
	entry := &fhir.HistoryEntry{
		ResourceType: "Goal", ResourceID: g.FHIRID, VersionID: 1,
		Resource: raw, Action: "create", Timestamp: g.CreatedAt,
	}
	return c.JSON(http.StatusOK, fhir.NewHistoryBundle([]*fhir.HistoryEntry{entry}, 1, "/fhir"))
}

func (h *Handler) handlePatch(c echo.Context, resourceType, fhirID string, applyFn func(echo.Context, map[string]interface{}) error) error {
	contentType := c.Request().Header.Get("Content-Type")
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome("failed to read request body"))
	}

	var currentResource map[string]interface{}
	switch resourceType {
	case "CarePlan":
		existing, err := h.svc.GetCarePlanByFHIRID(c.Request().Context(), fhirID)
		if err != nil {
			return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome(resourceType, fhirID))
		}
		currentResource = existing.ToFHIR()
	case "Goal":
		existing, err := h.svc.GetGoalByFHIRID(c.Request().Context(), fhirID)
		if err != nil {
			return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome(resourceType, fhirID))
		}
		currentResource = existing.ToFHIR()
	default:
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome("unsupported resource type for PATCH"))
	}

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
