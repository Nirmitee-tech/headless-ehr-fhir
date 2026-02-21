package structuremap

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
	read.GET("/structure-maps", h.ListStructureMaps)
	read.GET("/structure-maps/:id", h.GetStructureMap)

	write := api.Group("", role)
	write.POST("/structure-maps", h.CreateStructureMap)
	write.PUT("/structure-maps/:id", h.UpdateStructureMap)
	write.DELETE("/structure-maps/:id", h.DeleteStructureMap)

	fhirRead := fhirGroup.Group("", role)
	fhirRead.GET("/StructureMap", h.SearchStructureMapsFHIR)
	fhirRead.GET("/StructureMap/:id", h.GetStructureMapFHIR)

	fhirWrite := fhirGroup.Group("", role)
	fhirWrite.POST("/StructureMap", h.CreateStructureMapFHIR)
	fhirWrite.PUT("/StructureMap/:id", h.UpdateStructureMapFHIR)
	fhirWrite.DELETE("/StructureMap/:id", h.DeleteStructureMapFHIR)
	fhirWrite.PATCH("/StructureMap/:id", h.PatchStructureMapFHIR)

	fhirRead.POST("/StructureMap/_search", h.SearchStructureMapsFHIR)
	fhirRead.GET("/StructureMap/:id/_history/:vid", h.VreadStructureMapFHIR)
	fhirRead.GET("/StructureMap/:id/_history", h.HistoryStructureMapFHIR)
}

// -- REST Endpoints --

func (h *Handler) CreateStructureMap(c echo.Context) error {
	var sm StructureMap
	if err := c.Bind(&sm); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateStructureMap(c.Request().Context(), &sm); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, sm)
}

func (h *Handler) GetStructureMap(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	sm, err := h.svc.GetStructureMap(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "structure map not found")
	}
	return c.JSON(http.StatusOK, sm)
}

func (h *Handler) ListStructureMaps(c echo.Context) error {
	pg := pagination.FromContext(c)
	items, total, err := h.svc.SearchStructureMaps(c.Request().Context(), nil, pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *Handler) UpdateStructureMap(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var sm StructureMap
	if err := c.Bind(&sm); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	sm.ID = id
	if err := h.svc.UpdateStructureMap(c.Request().Context(), &sm); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, sm)
}

func (h *Handler) DeleteStructureMap(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeleteStructureMap(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

// -- FHIR Endpoints --

func (h *Handler) SearchStructureMapsFHIR(c echo.Context) error {
	pg := pagination.FromContext(c)
	params := fhir.ExtractSearchParams(c)
	items, total, err := h.svc.SearchStructureMaps(c.Request().Context(), params, pg.Limit, pg.Offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	resources := make([]interface{}, len(items))
	for i, item := range items {
		resources[i] = item.ToFHIR()
	}
	return c.JSON(http.StatusOK, fhir.NewSearchBundle(resources, total, "/fhir/StructureMap"))
}

func (h *Handler) GetStructureMapFHIR(c echo.Context) error {
	sm, err := h.svc.GetStructureMapByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("StructureMap", c.Param("id")))
	}
	fhir.SetVersionHeaders(c, 1, sm.UpdatedAt.Format("2006-01-02T15:04:05Z"))
	return c.JSON(http.StatusOK, sm.ToFHIR())
}

func (h *Handler) CreateStructureMapFHIR(c echo.Context) error {
	var sm StructureMap
	if err := c.Bind(&sm); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	if err := h.svc.CreateStructureMap(c.Request().Context(), &sm); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	c.Response().Header().Set("Location", "/fhir/StructureMap/"+sm.FHIRID)
	return c.JSON(http.StatusCreated, sm.ToFHIR())
}

func (h *Handler) UpdateStructureMapFHIR(c echo.Context) error {
	var sm StructureMap
	if err := c.Bind(&sm); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	existing, err := h.svc.GetStructureMapByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("StructureMap", c.Param("id")))
	}
	sm.ID = existing.ID
	sm.FHIRID = existing.FHIRID
	if err := h.svc.UpdateStructureMap(c.Request().Context(), &sm); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, sm.ToFHIR())
}

func (h *Handler) DeleteStructureMapFHIR(c echo.Context) error {
	existing, err := h.svc.GetStructureMapByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("StructureMap", c.Param("id")))
	}
	if err := h.svc.DeleteStructureMap(c.Request().Context(), existing.ID); err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) PatchStructureMapFHIR(c echo.Context) error {
	return h.handlePatch(c, c.Param("id"))
}

func (h *Handler) VreadStructureMapFHIR(c echo.Context) error {
	sm, err := h.svc.GetStructureMapByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("StructureMap", c.Param("id")))
	}
	result := sm.ToFHIR()
	fhir.SetVersionHeaders(c, 1, sm.UpdatedAt.Format("2006-01-02T15:04:05Z"))
	return c.JSON(http.StatusOK, result)
}

func (h *Handler) HistoryStructureMapFHIR(c echo.Context) error {
	sm, err := h.svc.GetStructureMapByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("StructureMap", c.Param("id")))
	}
	result := sm.ToFHIR()
	raw, _ := json.Marshal(result)
	entry := &fhir.HistoryEntry{
		ResourceType: "StructureMap", ResourceID: sm.FHIRID, VersionID: 1,
		Resource: raw, Action: "create", Timestamp: sm.CreatedAt,
	}
	return c.JSON(http.StatusOK, fhir.NewHistoryBundle([]*fhir.HistoryEntry{entry}, 1, "/fhir"))
}

func (h *Handler) handlePatch(c echo.Context, fhirID string) error {
	contentType := c.Request().Header.Get("Content-Type")
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome("failed to read request body"))
	}
	existing, err := h.svc.GetStructureMapByFHIRID(c.Request().Context(), fhirID)
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("StructureMap", fhirID))
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
	if err := h.svc.UpdateStructureMap(c.Request().Context(), existing); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, existing.ToFHIR())
}
