package graphdefinition

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
	read.GET("/graph-definitions", h.ListGraphDefinitions)
	read.GET("/graph-definitions/:id", h.GetGraphDefinition)

	write := api.Group("", role)
	write.POST("/graph-definitions", h.CreateGraphDefinition)
	write.PUT("/graph-definitions/:id", h.UpdateGraphDefinition)
	write.DELETE("/graph-definitions/:id", h.DeleteGraphDefinition)

	fhirRead := fhirGroup.Group("", role)
	fhirRead.GET("/GraphDefinition", h.SearchGraphDefinitionsFHIR)
	fhirRead.GET("/GraphDefinition/:id", h.GetGraphDefinitionFHIR)

	fhirWrite := fhirGroup.Group("", role)
	fhirWrite.POST("/GraphDefinition", h.CreateGraphDefinitionFHIR)
	fhirWrite.PUT("/GraphDefinition/:id", h.UpdateGraphDefinitionFHIR)
	fhirWrite.DELETE("/GraphDefinition/:id", h.DeleteGraphDefinitionFHIR)
	fhirWrite.PATCH("/GraphDefinition/:id", h.PatchGraphDefinitionFHIR)

	fhirRead.POST("/GraphDefinition/_search", h.SearchGraphDefinitionsFHIR)
	fhirRead.GET("/GraphDefinition/:id/_history/:vid", h.VreadGraphDefinitionFHIR)
	fhirRead.GET("/GraphDefinition/:id/_history", h.HistoryGraphDefinitionFHIR)
}

// -- REST Endpoints --

func (h *Handler) CreateGraphDefinition(c echo.Context) error {
	var g GraphDefinition
	if err := c.Bind(&g); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateGraphDefinition(c.Request().Context(), &g); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, g)
}

func (h *Handler) GetGraphDefinition(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	g, err := h.svc.GetGraphDefinition(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "graph definition not found")
	}
	return c.JSON(http.StatusOK, g)
}

func (h *Handler) ListGraphDefinitions(c echo.Context) error {
	pg := pagination.FromContext(c)
	items, total, err := h.svc.SearchGraphDefinitions(c.Request().Context(), nil, pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *Handler) UpdateGraphDefinition(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var g GraphDefinition
	if err := c.Bind(&g); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	g.ID = id
	if err := h.svc.UpdateGraphDefinition(c.Request().Context(), &g); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, g)
}

func (h *Handler) DeleteGraphDefinition(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeleteGraphDefinition(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

// -- FHIR Endpoints --

func (h *Handler) SearchGraphDefinitionsFHIR(c echo.Context) error {
	pg := pagination.FromContext(c)
	params := map[string]string{}
	for _, k := range []string{"status", "name", "url", "start"} {
		if v := c.QueryParam(k); v != "" {
			params[k] = v
		}
	}
	items, total, err := h.svc.SearchGraphDefinitions(c.Request().Context(), params, pg.Limit, pg.Offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	resources := make([]interface{}, len(items))
	for i, item := range items {
		resources[i] = item.ToFHIR()
	}
	return c.JSON(http.StatusOK, fhir.NewSearchBundle(resources, total, "/fhir/GraphDefinition"))
}

func (h *Handler) GetGraphDefinitionFHIR(c echo.Context) error {
	g, err := h.svc.GetGraphDefinitionByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("GraphDefinition", c.Param("id")))
	}
	return c.JSON(http.StatusOK, g.ToFHIR())
}

func (h *Handler) CreateGraphDefinitionFHIR(c echo.Context) error {
	var g GraphDefinition
	if err := c.Bind(&g); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	if err := h.svc.CreateGraphDefinition(c.Request().Context(), &g); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	c.Response().Header().Set("Location", "/fhir/GraphDefinition/"+g.FHIRID)
	return c.JSON(http.StatusCreated, g.ToFHIR())
}

func (h *Handler) UpdateGraphDefinitionFHIR(c echo.Context) error {
	var g GraphDefinition
	if err := c.Bind(&g); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	existing, err := h.svc.GetGraphDefinitionByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("GraphDefinition", c.Param("id")))
	}
	g.ID = existing.ID
	g.FHIRID = existing.FHIRID
	if err := h.svc.UpdateGraphDefinition(c.Request().Context(), &g); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, g.ToFHIR())
}

func (h *Handler) DeleteGraphDefinitionFHIR(c echo.Context) error {
	existing, err := h.svc.GetGraphDefinitionByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("GraphDefinition", c.Param("id")))
	}
	if err := h.svc.DeleteGraphDefinition(c.Request().Context(), existing.ID); err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) PatchGraphDefinitionFHIR(c echo.Context) error {
	return h.handlePatch(c, c.Param("id"))
}

func (h *Handler) VreadGraphDefinitionFHIR(c echo.Context) error {
	g, err := h.svc.GetGraphDefinitionByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("GraphDefinition", c.Param("id")))
	}
	result := g.ToFHIR()
	fhir.SetVersionHeaders(c, 1, g.UpdatedAt.Format("2006-01-02T15:04:05Z"))
	return c.JSON(http.StatusOK, result)
}

func (h *Handler) HistoryGraphDefinitionFHIR(c echo.Context) error {
	g, err := h.svc.GetGraphDefinitionByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("GraphDefinition", c.Param("id")))
	}
	result := g.ToFHIR()
	raw, _ := json.Marshal(result)
	entry := &fhir.HistoryEntry{
		ResourceType: "GraphDefinition", ResourceID: g.FHIRID, VersionID: 1,
		Resource: raw, Action: "create", Timestamp: g.CreatedAt,
	}
	return c.JSON(http.StatusOK, fhir.NewHistoryBundle([]*fhir.HistoryEntry{entry}, 1, "/fhir"))
}

func (h *Handler) handlePatch(c echo.Context, fhirID string) error {
	contentType := c.Request().Header.Get("Content-Type")
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome("failed to read request body"))
	}
	existing, err := h.svc.GetGraphDefinitionByFHIRID(c.Request().Context(), fhirID)
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("GraphDefinition", fhirID))
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
	if err := h.svc.UpdateGraphDefinition(c.Request().Context(), existing); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, existing.ToFHIR())
}
