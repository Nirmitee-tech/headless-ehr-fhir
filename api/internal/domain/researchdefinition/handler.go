package researchdefinition

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
	read.GET("/research-definitions", h.ListResearchDefinitions)
	read.GET("/research-definitions/:id", h.GetResearchDefinition)

	write := api.Group("", role)
	write.POST("/research-definitions", h.CreateResearchDefinition)
	write.PUT("/research-definitions/:id", h.UpdateResearchDefinition)
	write.DELETE("/research-definitions/:id", h.DeleteResearchDefinition)

	fhirRead := fhirGroup.Group("", role)
	fhirRead.GET("/ResearchDefinition", h.SearchResearchDefinitionsFHIR)
	fhirRead.GET("/ResearchDefinition/:id", h.GetResearchDefinitionFHIR)

	fhirWrite := fhirGroup.Group("", role)
	fhirWrite.POST("/ResearchDefinition", h.CreateResearchDefinitionFHIR)
	fhirWrite.PUT("/ResearchDefinition/:id", h.UpdateResearchDefinitionFHIR)
	fhirWrite.DELETE("/ResearchDefinition/:id", h.DeleteResearchDefinitionFHIR)
	fhirWrite.PATCH("/ResearchDefinition/:id", h.PatchResearchDefinitionFHIR)

	fhirRead.POST("/ResearchDefinition/_search", h.SearchResearchDefinitionsFHIR)
	fhirRead.GET("/ResearchDefinition/:id/_history/:vid", h.VreadResearchDefinitionFHIR)
	fhirRead.GET("/ResearchDefinition/:id/_history", h.HistoryResearchDefinitionFHIR)
}

// -- REST Endpoints --

func (h *Handler) CreateResearchDefinition(c echo.Context) error {
	var e ResearchDefinition
	if err := c.Bind(&e); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateResearchDefinition(c.Request().Context(), &e); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, e)
}

func (h *Handler) GetResearchDefinition(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	e, err := h.svc.GetResearchDefinition(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "research definition not found")
	}
	return c.JSON(http.StatusOK, e)
}

func (h *Handler) ListResearchDefinitions(c echo.Context) error {
	pg := pagination.FromContext(c)
	items, total, err := h.svc.SearchResearchDefinitions(c.Request().Context(), nil, pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *Handler) UpdateResearchDefinition(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var e ResearchDefinition
	if err := c.Bind(&e); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	e.ID = id
	if err := h.svc.UpdateResearchDefinition(c.Request().Context(), &e); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, e)
}

func (h *Handler) DeleteResearchDefinition(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeleteResearchDefinition(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

// -- FHIR Endpoints --

func (h *Handler) SearchResearchDefinitionsFHIR(c echo.Context) error {
	pg := pagination.FromContext(c)
	params := fhir.ExtractSearchParams(c)
	items, total, err := h.svc.SearchResearchDefinitions(c.Request().Context(), params, pg.Limit, pg.Offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	resources := make([]interface{}, len(items))
	for i, item := range items {
		resources[i] = item.ToFHIR()
	}
	return c.JSON(http.StatusOK, fhir.NewSearchBundle(resources, total, "/fhir/ResearchDefinition"))
}

func (h *Handler) GetResearchDefinitionFHIR(c echo.Context) error {
	e, err := h.svc.GetResearchDefinitionByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("ResearchDefinition", c.Param("id")))
	}
	return c.JSON(http.StatusOK, e.ToFHIR())
}

func (h *Handler) CreateResearchDefinitionFHIR(c echo.Context) error {
	var e ResearchDefinition
	if err := c.Bind(&e); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	if err := h.svc.CreateResearchDefinition(c.Request().Context(), &e); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	c.Response().Header().Set("Location", "/fhir/ResearchDefinition/"+e.FHIRID)
	return c.JSON(http.StatusCreated, e.ToFHIR())
}

func (h *Handler) UpdateResearchDefinitionFHIR(c echo.Context) error {
	var e ResearchDefinition
	if err := c.Bind(&e); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	existing, err := h.svc.GetResearchDefinitionByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("ResearchDefinition", c.Param("id")))
	}
	e.ID = existing.ID
	e.FHIRID = existing.FHIRID
	if err := h.svc.UpdateResearchDefinition(c.Request().Context(), &e); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, e.ToFHIR())
}

func (h *Handler) DeleteResearchDefinitionFHIR(c echo.Context) error {
	existing, err := h.svc.GetResearchDefinitionByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("ResearchDefinition", c.Param("id")))
	}
	if err := h.svc.DeleteResearchDefinition(c.Request().Context(), existing.ID); err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) PatchResearchDefinitionFHIR(c echo.Context) error {
	return h.handlePatch(c, c.Param("id"))
}

func (h *Handler) VreadResearchDefinitionFHIR(c echo.Context) error {
	e, err := h.svc.GetResearchDefinitionByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("ResearchDefinition", c.Param("id")))
	}
	result := e.ToFHIR()
	fhir.SetVersionHeaders(c, 1, e.UpdatedAt.Format("2006-01-02T15:04:05Z"))
	return c.JSON(http.StatusOK, result)
}

func (h *Handler) HistoryResearchDefinitionFHIR(c echo.Context) error {
	e, err := h.svc.GetResearchDefinitionByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("ResearchDefinition", c.Param("id")))
	}
	result := e.ToFHIR()
	raw, _ := json.Marshal(result)
	entry := &fhir.HistoryEntry{
		ResourceType: "ResearchDefinition", ResourceID: e.FHIRID, VersionID: 1,
		Resource: raw, Action: "create", Timestamp: e.CreatedAt,
	}
	return c.JSON(http.StatusOK, fhir.NewHistoryBundle([]*fhir.HistoryEntry{entry}, 1, "/fhir"))
}

func (h *Handler) handlePatch(c echo.Context, fhirID string) error {
	contentType := c.Request().Header.Get("Content-Type")
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome("failed to read request body"))
	}
	existing, err := h.svc.GetResearchDefinitionByFHIRID(c.Request().Context(), fhirID)
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("ResearchDefinition", fhirID))
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
	if err := h.svc.UpdateResearchDefinition(c.Request().Context(), existing); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, existing.ToFHIR())
}
