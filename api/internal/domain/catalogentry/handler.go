package catalogentry

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
	read.GET("/catalog-entries", h.ListCatalogEntries)
	read.GET("/catalog-entries/:id", h.GetCatalogEntry)

	write := api.Group("", role)
	write.POST("/catalog-entries", h.CreateCatalogEntry)
	write.PUT("/catalog-entries/:id", h.UpdateCatalogEntry)
	write.DELETE("/catalog-entries/:id", h.DeleteCatalogEntry)

	fhirRead := fhirGroup.Group("", role)
	fhirRead.GET("/CatalogEntry", h.SearchCatalogEntriesFHIR)
	fhirRead.GET("/CatalogEntry/:id", h.GetCatalogEntryFHIR)

	fhirWrite := fhirGroup.Group("", role)
	fhirWrite.POST("/CatalogEntry", h.CreateCatalogEntryFHIR)
	fhirWrite.PUT("/CatalogEntry/:id", h.UpdateCatalogEntryFHIR)
	fhirWrite.DELETE("/CatalogEntry/:id", h.DeleteCatalogEntryFHIR)
	fhirWrite.PATCH("/CatalogEntry/:id", h.PatchCatalogEntryFHIR)

	fhirRead.POST("/CatalogEntry/_search", h.SearchCatalogEntriesFHIR)
	fhirRead.GET("/CatalogEntry/:id/_history/:vid", h.VreadCatalogEntryFHIR)
	fhirRead.GET("/CatalogEntry/:id/_history", h.HistoryCatalogEntryFHIR)
}

// -- REST Endpoints --

func (h *Handler) CreateCatalogEntry(c echo.Context) error {
	var ce CatalogEntry
	if err := c.Bind(&ce); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateCatalogEntry(c.Request().Context(), &ce); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, ce)
}

func (h *Handler) GetCatalogEntry(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	ce, err := h.svc.GetCatalogEntry(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "catalog entry not found")
	}
	return c.JSON(http.StatusOK, ce)
}

func (h *Handler) ListCatalogEntries(c echo.Context) error {
	pg := pagination.FromContext(c)
	items, total, err := h.svc.SearchCatalogEntries(c.Request().Context(), nil, pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *Handler) UpdateCatalogEntry(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var ce CatalogEntry
	if err := c.Bind(&ce); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	ce.ID = id
	if err := h.svc.UpdateCatalogEntry(c.Request().Context(), &ce); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, ce)
}

func (h *Handler) DeleteCatalogEntry(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeleteCatalogEntry(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

// -- FHIR Endpoints --

func (h *Handler) SearchCatalogEntriesFHIR(c echo.Context) error {
	pg := pagination.FromContext(c)
	params := map[string]string{}
	for _, k := range []string{"status", "type", "orderable"} {
		if v := c.QueryParam(k); v != "" {
			params[k] = v
		}
	}
	items, total, err := h.svc.SearchCatalogEntries(c.Request().Context(), params, pg.Limit, pg.Offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	resources := make([]interface{}, len(items))
	for i, item := range items {
		resources[i] = item.ToFHIR()
	}
	return c.JSON(http.StatusOK, fhir.NewSearchBundle(resources, total, "/fhir/CatalogEntry"))
}

func (h *Handler) GetCatalogEntryFHIR(c echo.Context) error {
	ce, err := h.svc.GetCatalogEntryByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("CatalogEntry", c.Param("id")))
	}
	return c.JSON(http.StatusOK, ce.ToFHIR())
}

func (h *Handler) CreateCatalogEntryFHIR(c echo.Context) error {
	var ce CatalogEntry
	if err := c.Bind(&ce); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	if err := h.svc.CreateCatalogEntry(c.Request().Context(), &ce); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	c.Response().Header().Set("Location", "/fhir/CatalogEntry/"+ce.FHIRID)
	return c.JSON(http.StatusCreated, ce.ToFHIR())
}

func (h *Handler) UpdateCatalogEntryFHIR(c echo.Context) error {
	var ce CatalogEntry
	if err := c.Bind(&ce); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	existing, err := h.svc.GetCatalogEntryByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("CatalogEntry", c.Param("id")))
	}
	ce.ID = existing.ID
	ce.FHIRID = existing.FHIRID
	if err := h.svc.UpdateCatalogEntry(c.Request().Context(), &ce); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, ce.ToFHIR())
}

func (h *Handler) DeleteCatalogEntryFHIR(c echo.Context) error {
	existing, err := h.svc.GetCatalogEntryByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("CatalogEntry", c.Param("id")))
	}
	if err := h.svc.DeleteCatalogEntry(c.Request().Context(), existing.ID); err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) PatchCatalogEntryFHIR(c echo.Context) error {
	return h.handlePatch(c, c.Param("id"))
}

func (h *Handler) VreadCatalogEntryFHIR(c echo.Context) error {
	ce, err := h.svc.GetCatalogEntryByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("CatalogEntry", c.Param("id")))
	}
	result := ce.ToFHIR()
	fhir.SetVersionHeaders(c, 1, ce.UpdatedAt.Format("2006-01-02T15:04:05Z"))
	return c.JSON(http.StatusOK, result)
}

func (h *Handler) HistoryCatalogEntryFHIR(c echo.Context) error {
	ce, err := h.svc.GetCatalogEntryByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("CatalogEntry", c.Param("id")))
	}
	result := ce.ToFHIR()
	raw, _ := json.Marshal(result)
	entry := &fhir.HistoryEntry{
		ResourceType: "CatalogEntry", ResourceID: ce.FHIRID, VersionID: 1,
		Resource: raw, Action: "create", Timestamp: ce.CreatedAt,
	}
	return c.JSON(http.StatusOK, fhir.NewHistoryBundle([]*fhir.HistoryEntry{entry}, 1, "/fhir"))
}

func (h *Handler) handlePatch(c echo.Context, fhirID string) error {
	contentType := c.Request().Header.Get("Content-Type")
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome("failed to read request body"))
	}
	existing, err := h.svc.GetCatalogEntryByFHIRID(c.Request().Context(), fhirID)
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("CatalogEntry", fhirID))
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
	if err := h.svc.UpdateCatalogEntry(c.Request().Context(), existing); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, existing.ToFHIR())
}
