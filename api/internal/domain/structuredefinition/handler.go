package structuredefinition

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
	read.GET("/structure-definitions", h.ListStructureDefinitions)
	read.GET("/structure-definitions/:id", h.GetStructureDefinition)

	write := api.Group("", role)
	write.POST("/structure-definitions", h.CreateStructureDefinition)
	write.PUT("/structure-definitions/:id", h.UpdateStructureDefinition)
	write.DELETE("/structure-definitions/:id", h.DeleteStructureDefinition)

	fhirRead := fhirGroup.Group("", role)
	fhirRead.GET("/StructureDefinition", h.SearchStructureDefinitionsFHIR)
	fhirRead.GET("/StructureDefinition/:id", h.GetStructureDefinitionFHIR)

	fhirWrite := fhirGroup.Group("", role)
	fhirWrite.POST("/StructureDefinition", h.CreateStructureDefinitionFHIR)
	fhirWrite.PUT("/StructureDefinition/:id", h.UpdateStructureDefinitionFHIR)
	fhirWrite.DELETE("/StructureDefinition/:id", h.DeleteStructureDefinitionFHIR)
	fhirWrite.PATCH("/StructureDefinition/:id", h.PatchStructureDefinitionFHIR)

	fhirRead.POST("/StructureDefinition/_search", h.SearchStructureDefinitionsFHIR)
	fhirRead.GET("/StructureDefinition/:id/_history/:vid", h.VreadStructureDefinitionFHIR)
	fhirRead.GET("/StructureDefinition/:id/_history", h.HistoryStructureDefinitionFHIR)
}

// -- REST Endpoints --

func (h *Handler) CreateStructureDefinition(c echo.Context) error {
	var sd StructureDefinition
	if err := c.Bind(&sd); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateStructureDefinition(c.Request().Context(), &sd); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, sd)
}

func (h *Handler) GetStructureDefinition(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	sd, err := h.svc.GetStructureDefinition(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "structure definition not found")
	}
	return c.JSON(http.StatusOK, sd)
}

func (h *Handler) ListStructureDefinitions(c echo.Context) error {
	pg := pagination.FromContext(c)
	items, total, err := h.svc.SearchStructureDefinitions(c.Request().Context(), nil, pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *Handler) UpdateStructureDefinition(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var sd StructureDefinition
	if err := c.Bind(&sd); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	sd.ID = id
	if err := h.svc.UpdateStructureDefinition(c.Request().Context(), &sd); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, sd)
}

func (h *Handler) DeleteStructureDefinition(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeleteStructureDefinition(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

// -- FHIR Endpoints --

func (h *Handler) SearchStructureDefinitionsFHIR(c echo.Context) error {
	pg := pagination.FromContext(c)
	params := fhir.ExtractSearchParams(c)
	items, total, err := h.svc.SearchStructureDefinitions(c.Request().Context(), params, pg.Limit, pg.Offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	resources := make([]interface{}, len(items))
	for i, item := range items {
		resources[i] = item.ToFHIR()
	}
	return c.JSON(http.StatusOK, fhir.NewSearchBundle(resources, total, "/fhir/StructureDefinition"))
}

func (h *Handler) GetStructureDefinitionFHIR(c echo.Context) error {
	sd, err := h.svc.GetStructureDefinitionByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("StructureDefinition", c.Param("id")))
	}
	return c.JSON(http.StatusOK, sd.ToFHIR())
}

func (h *Handler) CreateStructureDefinitionFHIR(c echo.Context) error {
	var sd StructureDefinition
	if err := c.Bind(&sd); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	if err := h.svc.CreateStructureDefinition(c.Request().Context(), &sd); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	c.Response().Header().Set("Location", "/fhir/StructureDefinition/"+sd.FHIRID)
	return c.JSON(http.StatusCreated, sd.ToFHIR())
}

func (h *Handler) UpdateStructureDefinitionFHIR(c echo.Context) error {
	var sd StructureDefinition
	if err := c.Bind(&sd); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	existing, err := h.svc.GetStructureDefinitionByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("StructureDefinition", c.Param("id")))
	}
	sd.ID = existing.ID
	sd.FHIRID = existing.FHIRID
	if err := h.svc.UpdateStructureDefinition(c.Request().Context(), &sd); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, sd.ToFHIR())
}

func (h *Handler) DeleteStructureDefinitionFHIR(c echo.Context) error {
	existing, err := h.svc.GetStructureDefinitionByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("StructureDefinition", c.Param("id")))
	}
	if err := h.svc.DeleteStructureDefinition(c.Request().Context(), existing.ID); err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) PatchStructureDefinitionFHIR(c echo.Context) error {
	return h.handlePatch(c, c.Param("id"))
}

func (h *Handler) VreadStructureDefinitionFHIR(c echo.Context) error {
	sd, err := h.svc.GetStructureDefinitionByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("StructureDefinition", c.Param("id")))
	}
	result := sd.ToFHIR()
	fhir.SetVersionHeaders(c, 1, sd.UpdatedAt.Format("2006-01-02T15:04:05Z"))
	return c.JSON(http.StatusOK, result)
}

func (h *Handler) HistoryStructureDefinitionFHIR(c echo.Context) error {
	sd, err := h.svc.GetStructureDefinitionByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("StructureDefinition", c.Param("id")))
	}
	result := sd.ToFHIR()
	raw, _ := json.Marshal(result)
	entry := &fhir.HistoryEntry{
		ResourceType: "StructureDefinition", ResourceID: sd.FHIRID, VersionID: 1,
		Resource: raw, Action: "create", Timestamp: sd.CreatedAt,
	}
	return c.JSON(http.StatusOK, fhir.NewHistoryBundle([]*fhir.HistoryEntry{entry}, 1, "/fhir"))
}

func (h *Handler) handlePatch(c echo.Context, fhirID string) error {
	contentType := c.Request().Header.Get("Content-Type")
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome("failed to read request body"))
	}
	existing, err := h.svc.GetStructureDefinitionByFHIRID(c.Request().Context(), fhirID)
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("StructureDefinition", fhirID))
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
	if err := h.svc.UpdateStructureDefinition(c.Request().Context(), existing); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, existing.ToFHIR())
}
