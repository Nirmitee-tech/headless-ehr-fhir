package conceptmap

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
	read.GET("/concept-maps", h.ListConceptMaps)
	read.GET("/concept-maps/:id", h.GetConceptMap)

	write := api.Group("", role)
	write.POST("/concept-maps", h.CreateConceptMap)
	write.PUT("/concept-maps/:id", h.UpdateConceptMap)
	write.DELETE("/concept-maps/:id", h.DeleteConceptMap)

	fhirRead := fhirGroup.Group("", role)
	fhirRead.GET("/ConceptMap", h.SearchConceptMapsFHIR)
	fhirRead.GET("/ConceptMap/:id", h.GetConceptMapFHIR)

	fhirWrite := fhirGroup.Group("", role)
	fhirWrite.POST("/ConceptMap", h.CreateConceptMapFHIR)
	fhirWrite.PUT("/ConceptMap/:id", h.UpdateConceptMapFHIR)
	fhirWrite.DELETE("/ConceptMap/:id", h.DeleteConceptMapFHIR)
	fhirWrite.PATCH("/ConceptMap/:id", h.PatchConceptMapFHIR)

	fhirRead.POST("/ConceptMap/_search", h.SearchConceptMapsFHIR)
	fhirRead.GET("/ConceptMap/:id/_history/:vid", h.VreadConceptMapFHIR)
	fhirRead.GET("/ConceptMap/:id/_history", h.HistoryConceptMapFHIR)
}

// -- REST Endpoints --

func (h *Handler) CreateConceptMap(c echo.Context) error {
	var cm ConceptMap
	if err := c.Bind(&cm); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateConceptMap(c.Request().Context(), &cm); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, cm)
}

func (h *Handler) GetConceptMap(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	cm, err := h.svc.GetConceptMap(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "concept map not found")
	}
	return c.JSON(http.StatusOK, cm)
}

func (h *Handler) ListConceptMaps(c echo.Context) error {
	pg := pagination.FromContext(c)
	items, total, err := h.svc.SearchConceptMaps(c.Request().Context(), nil, pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *Handler) UpdateConceptMap(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var cm ConceptMap
	if err := c.Bind(&cm); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	cm.ID = id
	if err := h.svc.UpdateConceptMap(c.Request().Context(), &cm); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, cm)
}

func (h *Handler) DeleteConceptMap(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeleteConceptMap(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

// -- FHIR Endpoints --

func (h *Handler) SearchConceptMapsFHIR(c echo.Context) error {
	pg := pagination.FromContext(c)
	params := map[string]string{}
	for _, k := range []string{"status", "url", "name", "source", "target"} {
		if v := c.QueryParam(k); v != "" {
			params[k] = v
		}
	}
	items, total, err := h.svc.SearchConceptMaps(c.Request().Context(), params, pg.Limit, pg.Offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	resources := make([]interface{}, len(items))
	for i, item := range items {
		resources[i] = item.ToFHIR()
	}
	return c.JSON(http.StatusOK, fhir.NewSearchBundle(resources, total, "/fhir/ConceptMap"))
}

func (h *Handler) GetConceptMapFHIR(c echo.Context) error {
	cm, err := h.svc.GetConceptMapByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("ConceptMap", c.Param("id")))
	}
	return c.JSON(http.StatusOK, cm.ToFHIR())
}

func (h *Handler) CreateConceptMapFHIR(c echo.Context) error {
	var cm ConceptMap
	if err := c.Bind(&cm); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	if err := h.svc.CreateConceptMap(c.Request().Context(), &cm); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	c.Response().Header().Set("Location", "/fhir/ConceptMap/"+cm.FHIRID)
	return c.JSON(http.StatusCreated, cm.ToFHIR())
}

func (h *Handler) UpdateConceptMapFHIR(c echo.Context) error {
	var cm ConceptMap
	if err := c.Bind(&cm); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	existing, err := h.svc.GetConceptMapByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("ConceptMap", c.Param("id")))
	}
	cm.ID = existing.ID
	cm.FHIRID = existing.FHIRID
	if err := h.svc.UpdateConceptMap(c.Request().Context(), &cm); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, cm.ToFHIR())
}

func (h *Handler) DeleteConceptMapFHIR(c echo.Context) error {
	existing, err := h.svc.GetConceptMapByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("ConceptMap", c.Param("id")))
	}
	if err := h.svc.DeleteConceptMap(c.Request().Context(), existing.ID); err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) PatchConceptMapFHIR(c echo.Context) error {
	return h.handlePatch(c, c.Param("id"))
}

func (h *Handler) VreadConceptMapFHIR(c echo.Context) error {
	cm, err := h.svc.GetConceptMapByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("ConceptMap", c.Param("id")))
	}
	result := cm.ToFHIR()
	fhir.SetVersionHeaders(c, 1, cm.UpdatedAt.Format("2006-01-02T15:04:05Z"))
	return c.JSON(http.StatusOK, result)
}

func (h *Handler) HistoryConceptMapFHIR(c echo.Context) error {
	cm, err := h.svc.GetConceptMapByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("ConceptMap", c.Param("id")))
	}
	result := cm.ToFHIR()
	raw, _ := json.Marshal(result)
	entry := &fhir.HistoryEntry{
		ResourceType: "ConceptMap", ResourceID: cm.FHIRID, VersionID: 1,
		Resource: raw, Action: "create", Timestamp: cm.CreatedAt,
	}
	return c.JSON(http.StatusOK, fhir.NewHistoryBundle([]*fhir.HistoryEntry{entry}, 1, "/fhir"))
}

func (h *Handler) handlePatch(c echo.Context, fhirID string) error {
	contentType := c.Request().Header.Get("Content-Type")
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome("failed to read request body"))
	}
	existing, err := h.svc.GetConceptMapByFHIRID(c.Request().Context(), fhirID)
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("ConceptMap", fhirID))
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
	if err := h.svc.UpdateConceptMap(c.Request().Context(), existing); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, existing.ToFHIR())
}
