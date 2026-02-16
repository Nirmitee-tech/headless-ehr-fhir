package compartmentdefinition

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
	read.GET("/compartment-definitions", h.ListCompartmentDefinitions)
	read.GET("/compartment-definitions/:id", h.GetCompartmentDefinition)

	write := api.Group("", role)
	write.POST("/compartment-definitions", h.CreateCompartmentDefinition)
	write.PUT("/compartment-definitions/:id", h.UpdateCompartmentDefinition)
	write.DELETE("/compartment-definitions/:id", h.DeleteCompartmentDefinition)

	fhirRead := fhirGroup.Group("", role)
	fhirRead.GET("/CompartmentDefinition", h.SearchCompartmentDefinitionsFHIR)
	fhirRead.GET("/CompartmentDefinition/:id", h.GetCompartmentDefinitionFHIR)

	fhirWrite := fhirGroup.Group("", role)
	fhirWrite.POST("/CompartmentDefinition", h.CreateCompartmentDefinitionFHIR)
	fhirWrite.PUT("/CompartmentDefinition/:id", h.UpdateCompartmentDefinitionFHIR)
	fhirWrite.DELETE("/CompartmentDefinition/:id", h.DeleteCompartmentDefinitionFHIR)
	fhirWrite.PATCH("/CompartmentDefinition/:id", h.PatchCompartmentDefinitionFHIR)

	fhirRead.POST("/CompartmentDefinition/_search", h.SearchCompartmentDefinitionsFHIR)
	fhirRead.GET("/CompartmentDefinition/:id/_history/:vid", h.VreadCompartmentDefinitionFHIR)
	fhirRead.GET("/CompartmentDefinition/:id/_history", h.HistoryCompartmentDefinitionFHIR)
}

// -- REST Endpoints --

func (h *Handler) CreateCompartmentDefinition(c echo.Context) error {
	var cd CompartmentDefinition
	if err := c.Bind(&cd); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateCompartmentDefinition(c.Request().Context(), &cd); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, cd)
}

func (h *Handler) GetCompartmentDefinition(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	cd, err := h.svc.GetCompartmentDefinition(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "compartment definition not found")
	}
	return c.JSON(http.StatusOK, cd)
}

func (h *Handler) ListCompartmentDefinitions(c echo.Context) error {
	pg := pagination.FromContext(c)
	items, total, err := h.svc.SearchCompartmentDefinitions(c.Request().Context(), nil, pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *Handler) UpdateCompartmentDefinition(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var cd CompartmentDefinition
	if err := c.Bind(&cd); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	cd.ID = id
	if err := h.svc.UpdateCompartmentDefinition(c.Request().Context(), &cd); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, cd)
}

func (h *Handler) DeleteCompartmentDefinition(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeleteCompartmentDefinition(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

// -- FHIR Endpoints --

func (h *Handler) SearchCompartmentDefinitionsFHIR(c echo.Context) error {
	pg := pagination.FromContext(c)
	params := map[string]string{}
	for _, k := range []string{"status", "url", "name", "code"} {
		if v := c.QueryParam(k); v != "" {
			params[k] = v
		}
	}
	items, total, err := h.svc.SearchCompartmentDefinitions(c.Request().Context(), params, pg.Limit, pg.Offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	resources := make([]interface{}, len(items))
	for i, item := range items {
		resources[i] = item.ToFHIR()
	}
	return c.JSON(http.StatusOK, fhir.NewSearchBundle(resources, total, "/fhir/CompartmentDefinition"))
}

func (h *Handler) GetCompartmentDefinitionFHIR(c echo.Context) error {
	cd, err := h.svc.GetCompartmentDefinitionByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("CompartmentDefinition", c.Param("id")))
	}
	return c.JSON(http.StatusOK, cd.ToFHIR())
}

func (h *Handler) CreateCompartmentDefinitionFHIR(c echo.Context) error {
	var cd CompartmentDefinition
	if err := c.Bind(&cd); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	if err := h.svc.CreateCompartmentDefinition(c.Request().Context(), &cd); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	c.Response().Header().Set("Location", "/fhir/CompartmentDefinition/"+cd.FHIRID)
	return c.JSON(http.StatusCreated, cd.ToFHIR())
}

func (h *Handler) UpdateCompartmentDefinitionFHIR(c echo.Context) error {
	var cd CompartmentDefinition
	if err := c.Bind(&cd); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	existing, err := h.svc.GetCompartmentDefinitionByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("CompartmentDefinition", c.Param("id")))
	}
	cd.ID = existing.ID
	cd.FHIRID = existing.FHIRID
	if err := h.svc.UpdateCompartmentDefinition(c.Request().Context(), &cd); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, cd.ToFHIR())
}

func (h *Handler) DeleteCompartmentDefinitionFHIR(c echo.Context) error {
	existing, err := h.svc.GetCompartmentDefinitionByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("CompartmentDefinition", c.Param("id")))
	}
	if err := h.svc.DeleteCompartmentDefinition(c.Request().Context(), existing.ID); err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) PatchCompartmentDefinitionFHIR(c echo.Context) error {
	return h.handlePatch(c, c.Param("id"))
}

func (h *Handler) VreadCompartmentDefinitionFHIR(c echo.Context) error {
	cd, err := h.svc.GetCompartmentDefinitionByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("CompartmentDefinition", c.Param("id")))
	}
	result := cd.ToFHIR()
	fhir.SetVersionHeaders(c, 1, cd.UpdatedAt.Format("2006-01-02T15:04:05Z"))
	return c.JSON(http.StatusOK, result)
}

func (h *Handler) HistoryCompartmentDefinitionFHIR(c echo.Context) error {
	cd, err := h.svc.GetCompartmentDefinitionByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("CompartmentDefinition", c.Param("id")))
	}
	result := cd.ToFHIR()
	raw, _ := json.Marshal(result)
	entry := &fhir.HistoryEntry{
		ResourceType: "CompartmentDefinition", ResourceID: cd.FHIRID, VersionID: 1,
		Resource: raw, Action: "create", Timestamp: cd.CreatedAt,
	}
	return c.JSON(http.StatusOK, fhir.NewHistoryBundle([]*fhir.HistoryEntry{entry}, 1, "/fhir"))
}

func (h *Handler) handlePatch(c echo.Context, fhirID string) error {
	contentType := c.Request().Header.Get("Content-Type")
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome("failed to read request body"))
	}
	existing, err := h.svc.GetCompartmentDefinitionByFHIRID(c.Request().Context(), fhirID)
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("CompartmentDefinition", fhirID))
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
	if err := h.svc.UpdateCompartmentDefinition(c.Request().Context(), existing); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, existing.ToFHIR())
}
