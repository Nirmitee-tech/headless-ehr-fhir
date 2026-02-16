package observationdefinition

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
	read.GET("/observation-definitions", h.ListObservationDefinitions)
	read.GET("/observation-definitions/:id", h.GetObservationDefinition)

	write := api.Group("", role)
	write.POST("/observation-definitions", h.CreateObservationDefinition)
	write.PUT("/observation-definitions/:id", h.UpdateObservationDefinition)
	write.DELETE("/observation-definitions/:id", h.DeleteObservationDefinition)

	fhirRead := fhirGroup.Group("", role)
	fhirRead.GET("/ObservationDefinition", h.SearchObservationDefinitionsFHIR)
	fhirRead.GET("/ObservationDefinition/:id", h.GetObservationDefinitionFHIR)

	fhirWrite := fhirGroup.Group("", role)
	fhirWrite.POST("/ObservationDefinition", h.CreateObservationDefinitionFHIR)
	fhirWrite.PUT("/ObservationDefinition/:id", h.UpdateObservationDefinitionFHIR)
	fhirWrite.DELETE("/ObservationDefinition/:id", h.DeleteObservationDefinitionFHIR)
	fhirWrite.PATCH("/ObservationDefinition/:id", h.PatchObservationDefinitionFHIR)

	fhirRead.POST("/ObservationDefinition/_search", h.SearchObservationDefinitionsFHIR)
	fhirRead.GET("/ObservationDefinition/:id/_history/:vid", h.VreadObservationDefinitionFHIR)
	fhirRead.GET("/ObservationDefinition/:id/_history", h.HistoryObservationDefinitionFHIR)
}

// -- REST Endpoints --

func (h *Handler) CreateObservationDefinition(c echo.Context) error {
	var od ObservationDefinition
	if err := c.Bind(&od); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateObservationDefinition(c.Request().Context(), &od); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, od)
}

func (h *Handler) GetObservationDefinition(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	od, err := h.svc.GetObservationDefinition(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "observation definition not found")
	}
	return c.JSON(http.StatusOK, od)
}

func (h *Handler) ListObservationDefinitions(c echo.Context) error {
	pg := pagination.FromContext(c)
	items, total, err := h.svc.SearchObservationDefinitions(c.Request().Context(), nil, pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *Handler) UpdateObservationDefinition(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var od ObservationDefinition
	if err := c.Bind(&od); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	od.ID = id
	if err := h.svc.UpdateObservationDefinition(c.Request().Context(), &od); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, od)
}

func (h *Handler) DeleteObservationDefinition(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeleteObservationDefinition(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

// -- FHIR Endpoints --

func (h *Handler) SearchObservationDefinitionsFHIR(c echo.Context) error {
	pg := pagination.FromContext(c)
	params := map[string]string{}
	for _, k := range []string{"status", "code", "category"} {
		if v := c.QueryParam(k); v != "" {
			params[k] = v
		}
	}
	items, total, err := h.svc.SearchObservationDefinitions(c.Request().Context(), params, pg.Limit, pg.Offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	resources := make([]interface{}, len(items))
	for i, item := range items {
		resources[i] = item.ToFHIR()
	}
	return c.JSON(http.StatusOK, fhir.NewSearchBundle(resources, total, "/fhir/ObservationDefinition"))
}

func (h *Handler) GetObservationDefinitionFHIR(c echo.Context) error {
	od, err := h.svc.GetObservationDefinitionByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("ObservationDefinition", c.Param("id")))
	}
	return c.JSON(http.StatusOK, od.ToFHIR())
}

func (h *Handler) CreateObservationDefinitionFHIR(c echo.Context) error {
	var od ObservationDefinition
	if err := c.Bind(&od); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	if err := h.svc.CreateObservationDefinition(c.Request().Context(), &od); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	c.Response().Header().Set("Location", "/fhir/ObservationDefinition/"+od.FHIRID)
	return c.JSON(http.StatusCreated, od.ToFHIR())
}

func (h *Handler) UpdateObservationDefinitionFHIR(c echo.Context) error {
	var od ObservationDefinition
	if err := c.Bind(&od); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	existing, err := h.svc.GetObservationDefinitionByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("ObservationDefinition", c.Param("id")))
	}
	od.ID = existing.ID
	od.FHIRID = existing.FHIRID
	if err := h.svc.UpdateObservationDefinition(c.Request().Context(), &od); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, od.ToFHIR())
}

func (h *Handler) DeleteObservationDefinitionFHIR(c echo.Context) error {
	existing, err := h.svc.GetObservationDefinitionByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("ObservationDefinition", c.Param("id")))
	}
	if err := h.svc.DeleteObservationDefinition(c.Request().Context(), existing.ID); err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) PatchObservationDefinitionFHIR(c echo.Context) error {
	return h.handlePatch(c, c.Param("id"))
}

func (h *Handler) VreadObservationDefinitionFHIR(c echo.Context) error {
	od, err := h.svc.GetObservationDefinitionByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("ObservationDefinition", c.Param("id")))
	}
	result := od.ToFHIR()
	fhir.SetVersionHeaders(c, 1, od.UpdatedAt.Format("2006-01-02T15:04:05Z"))
	return c.JSON(http.StatusOK, result)
}

func (h *Handler) HistoryObservationDefinitionFHIR(c echo.Context) error {
	od, err := h.svc.GetObservationDefinitionByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("ObservationDefinition", c.Param("id")))
	}
	result := od.ToFHIR()
	raw, _ := json.Marshal(result)
	entry := &fhir.HistoryEntry{
		ResourceType: "ObservationDefinition", ResourceID: od.FHIRID, VersionID: 1,
		Resource: raw, Action: "create", Timestamp: od.CreatedAt,
	}
	return c.JSON(http.StatusOK, fhir.NewHistoryBundle([]*fhir.HistoryEntry{entry}, 1, "/fhir"))
}

func (h *Handler) handlePatch(c echo.Context, fhirID string) error {
	contentType := c.Request().Header.Get("Content-Type")
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome("failed to read request body"))
	}
	existing, err := h.svc.GetObservationDefinitionByFHIRID(c.Request().Context(), fhirID)
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("ObservationDefinition", fhirID))
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
	if v, ok := patched["status"].(string); ok {
		existing.Status = v
	}
	if err := h.svc.UpdateObservationDefinition(c.Request().Context(), existing); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, existing.ToFHIR())
}
