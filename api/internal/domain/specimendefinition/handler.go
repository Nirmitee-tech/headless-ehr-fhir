package specimendefinition

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
	read.GET("/specimen-definitions", h.ListSpecimenDefinitions)
	read.GET("/specimen-definitions/:id", h.GetSpecimenDefinition)

	write := api.Group("", role)
	write.POST("/specimen-definitions", h.CreateSpecimenDefinition)
	write.PUT("/specimen-definitions/:id", h.UpdateSpecimenDefinition)
	write.DELETE("/specimen-definitions/:id", h.DeleteSpecimenDefinition)

	fhirRead := fhirGroup.Group("", role)
	fhirRead.GET("/SpecimenDefinition", h.SearchSpecimenDefinitionsFHIR)
	fhirRead.GET("/SpecimenDefinition/:id", h.GetSpecimenDefinitionFHIR)

	fhirWrite := fhirGroup.Group("", role)
	fhirWrite.POST("/SpecimenDefinition", h.CreateSpecimenDefinitionFHIR)
	fhirWrite.PUT("/SpecimenDefinition/:id", h.UpdateSpecimenDefinitionFHIR)
	fhirWrite.DELETE("/SpecimenDefinition/:id", h.DeleteSpecimenDefinitionFHIR)
	fhirWrite.PATCH("/SpecimenDefinition/:id", h.PatchSpecimenDefinitionFHIR)

	fhirRead.POST("/SpecimenDefinition/_search", h.SearchSpecimenDefinitionsFHIR)
	fhirRead.GET("/SpecimenDefinition/:id/_history/:vid", h.VreadSpecimenDefinitionFHIR)
	fhirRead.GET("/SpecimenDefinition/:id/_history", h.HistorySpecimenDefinitionFHIR)
}

// -- REST Endpoints --

func (h *Handler) CreateSpecimenDefinition(c echo.Context) error {
	var s SpecimenDefinition
	if err := c.Bind(&s); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateSpecimenDefinition(c.Request().Context(), &s); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, s)
}

func (h *Handler) GetSpecimenDefinition(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	s, err := h.svc.GetSpecimenDefinition(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "specimen definition not found")
	}
	return c.JSON(http.StatusOK, s)
}

func (h *Handler) ListSpecimenDefinitions(c echo.Context) error {
	pg := pagination.FromContext(c)
	items, total, err := h.svc.SearchSpecimenDefinitions(c.Request().Context(), nil, pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *Handler) UpdateSpecimenDefinition(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var s SpecimenDefinition
	if err := c.Bind(&s); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	s.ID = id
	if err := h.svc.UpdateSpecimenDefinition(c.Request().Context(), &s); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, s)
}

func (h *Handler) DeleteSpecimenDefinition(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeleteSpecimenDefinition(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

// -- FHIR Endpoints --

func (h *Handler) SearchSpecimenDefinitionsFHIR(c echo.Context) error {
	pg := pagination.FromContext(c)
	params := fhir.ExtractSearchParams(c)
	items, total, err := h.svc.SearchSpecimenDefinitions(c.Request().Context(), params, pg.Limit, pg.Offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	resources := make([]interface{}, len(items))
	for i, item := range items {
		resources[i] = item.ToFHIR()
	}
	return c.JSON(http.StatusOK, fhir.NewSearchBundle(resources, total, "/fhir/SpecimenDefinition"))
}

func (h *Handler) GetSpecimenDefinitionFHIR(c echo.Context) error {
	s, err := h.svc.GetSpecimenDefinitionByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("SpecimenDefinition", c.Param("id")))
	}
	fhir.SetVersionHeaders(c, 1, s.UpdatedAt.Format("2006-01-02T15:04:05Z"))
	return c.JSON(http.StatusOK, s.ToFHIR())
}

func (h *Handler) CreateSpecimenDefinitionFHIR(c echo.Context) error {
	var s SpecimenDefinition
	if err := c.Bind(&s); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	if err := h.svc.CreateSpecimenDefinition(c.Request().Context(), &s); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	c.Response().Header().Set("Location", "/fhir/SpecimenDefinition/"+s.FHIRID)
	return c.JSON(http.StatusCreated, s.ToFHIR())
}

func (h *Handler) UpdateSpecimenDefinitionFHIR(c echo.Context) error {
	var s SpecimenDefinition
	if err := c.Bind(&s); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	existing, err := h.svc.GetSpecimenDefinitionByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("SpecimenDefinition", c.Param("id")))
	}
	s.ID = existing.ID
	s.FHIRID = existing.FHIRID
	if err := h.svc.UpdateSpecimenDefinition(c.Request().Context(), &s); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, s.ToFHIR())
}

func (h *Handler) DeleteSpecimenDefinitionFHIR(c echo.Context) error {
	existing, err := h.svc.GetSpecimenDefinitionByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("SpecimenDefinition", c.Param("id")))
	}
	if err := h.svc.DeleteSpecimenDefinition(c.Request().Context(), existing.ID); err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) PatchSpecimenDefinitionFHIR(c echo.Context) error {
	return h.handlePatch(c, c.Param("id"))
}

func (h *Handler) VreadSpecimenDefinitionFHIR(c echo.Context) error {
	s, err := h.svc.GetSpecimenDefinitionByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("SpecimenDefinition", c.Param("id")))
	}
	result := s.ToFHIR()
	fhir.SetVersionHeaders(c, 1, s.UpdatedAt.Format("2006-01-02T15:04:05Z"))
	return c.JSON(http.StatusOK, result)
}

func (h *Handler) HistorySpecimenDefinitionFHIR(c echo.Context) error {
	s, err := h.svc.GetSpecimenDefinitionByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("SpecimenDefinition", c.Param("id")))
	}
	result := s.ToFHIR()
	raw, _ := json.Marshal(result)
	entry := &fhir.HistoryEntry{
		ResourceType: "SpecimenDefinition", ResourceID: s.FHIRID, VersionID: 1,
		Resource: raw, Action: "create", Timestamp: s.CreatedAt,
	}
	return c.JSON(http.StatusOK, fhir.NewHistoryBundle([]*fhir.HistoryEntry{entry}, 1, "/fhir"))
}

func (h *Handler) handlePatch(c echo.Context, fhirID string) error {
	contentType := c.Request().Header.Get("Content-Type")
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome("failed to read request body"))
	}
	existing, err := h.svc.GetSpecimenDefinitionByFHIRID(c.Request().Context(), fhirID)
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("SpecimenDefinition", fhirID))
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
	if v, ok := patched["timeAspect"].(string); ok {
		existing.TimeAspect = &v
	}
	if err := h.svc.UpdateSpecimenDefinition(c.Request().Context(), existing); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, existing.ToFHIR())
}
