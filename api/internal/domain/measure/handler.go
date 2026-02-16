package measure

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
	read.GET("/measures", h.ListMeasures)
	read.GET("/measures/:id", h.GetMeasure)

	write := api.Group("", role)
	write.POST("/measures", h.CreateMeasure)
	write.PUT("/measures/:id", h.UpdateMeasure)
	write.DELETE("/measures/:id", h.DeleteMeasure)

	fhirRead := fhirGroup.Group("", role)
	fhirRead.GET("/Measure", h.SearchMeasuresFHIR)
	fhirRead.GET("/Measure/:id", h.GetMeasureFHIR)

	fhirWrite := fhirGroup.Group("", role)
	fhirWrite.POST("/Measure", h.CreateMeasureFHIR)
	fhirWrite.PUT("/Measure/:id", h.UpdateMeasureFHIR)
	fhirWrite.DELETE("/Measure/:id", h.DeleteMeasureFHIR)
	fhirWrite.PATCH("/Measure/:id", h.PatchMeasureFHIR)

	fhirRead.POST("/Measure/_search", h.SearchMeasuresFHIR)
	fhirRead.GET("/Measure/:id/_history/:vid", h.VreadMeasureFHIR)
	fhirRead.GET("/Measure/:id/_history", h.HistoryMeasureFHIR)
}

// -- REST Endpoints --

func (h *Handler) CreateMeasure(c echo.Context) error {
	var m Measure
	if err := c.Bind(&m); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateMeasure(c.Request().Context(), &m); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, m)
}

func (h *Handler) GetMeasure(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	m, err := h.svc.GetMeasure(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "measure not found")
	}
	return c.JSON(http.StatusOK, m)
}

func (h *Handler) ListMeasures(c echo.Context) error {
	pg := pagination.FromContext(c)
	items, total, err := h.svc.SearchMeasures(c.Request().Context(), nil, pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *Handler) UpdateMeasure(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var m Measure
	if err := c.Bind(&m); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	m.ID = id
	if err := h.svc.UpdateMeasure(c.Request().Context(), &m); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, m)
}

func (h *Handler) DeleteMeasure(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeleteMeasure(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

// -- FHIR Endpoints --

func (h *Handler) SearchMeasuresFHIR(c echo.Context) error {
	pg := pagination.FromContext(c)
	params := map[string]string{}
	for _, k := range []string{"status", "name", "title", "url"} {
		if v := c.QueryParam(k); v != "" {
			params[k] = v
		}
	}
	items, total, err := h.svc.SearchMeasures(c.Request().Context(), params, pg.Limit, pg.Offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	resources := make([]interface{}, len(items))
	for i, item := range items {
		resources[i] = item.ToFHIR()
	}
	return c.JSON(http.StatusOK, fhir.NewSearchBundle(resources, total, "/fhir/Measure"))
}

func (h *Handler) GetMeasureFHIR(c echo.Context) error {
	m, err := h.svc.GetMeasureByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Measure", c.Param("id")))
	}
	return c.JSON(http.StatusOK, m.ToFHIR())
}

func (h *Handler) CreateMeasureFHIR(c echo.Context) error {
	var m Measure
	if err := c.Bind(&m); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	if err := h.svc.CreateMeasure(c.Request().Context(), &m); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	c.Response().Header().Set("Location", "/fhir/Measure/"+m.FHIRID)
	return c.JSON(http.StatusCreated, m.ToFHIR())
}

func (h *Handler) UpdateMeasureFHIR(c echo.Context) error {
	var m Measure
	if err := c.Bind(&m); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	existing, err := h.svc.GetMeasureByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Measure", c.Param("id")))
	}
	m.ID = existing.ID
	m.FHIRID = existing.FHIRID
	if err := h.svc.UpdateMeasure(c.Request().Context(), &m); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, m.ToFHIR())
}

func (h *Handler) DeleteMeasureFHIR(c echo.Context) error {
	existing, err := h.svc.GetMeasureByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Measure", c.Param("id")))
	}
	if err := h.svc.DeleteMeasure(c.Request().Context(), existing.ID); err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) PatchMeasureFHIR(c echo.Context) error {
	return h.handlePatch(c, c.Param("id"))
}

func (h *Handler) VreadMeasureFHIR(c echo.Context) error {
	m, err := h.svc.GetMeasureByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Measure", c.Param("id")))
	}
	result := m.ToFHIR()
	fhir.SetVersionHeaders(c, 1, m.UpdatedAt.Format("2006-01-02T15:04:05Z"))
	return c.JSON(http.StatusOK, result)
}

func (h *Handler) HistoryMeasureFHIR(c echo.Context) error {
	m, err := h.svc.GetMeasureByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Measure", c.Param("id")))
	}
	result := m.ToFHIR()
	raw, _ := json.Marshal(result)
	entry := &fhir.HistoryEntry{
		ResourceType: "Measure", ResourceID: m.FHIRID, VersionID: 1,
		Resource: raw, Action: "create", Timestamp: m.CreatedAt,
	}
	return c.JSON(http.StatusOK, fhir.NewHistoryBundle([]*fhir.HistoryEntry{entry}, 1, "/fhir"))
}

func (h *Handler) handlePatch(c echo.Context, fhirID string) error {
	contentType := c.Request().Header.Get("Content-Type")
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome("failed to read request body"))
	}
	existing, err := h.svc.GetMeasureByFHIRID(c.Request().Context(), fhirID)
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Measure", fhirID))
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
	if err := h.svc.UpdateMeasure(c.Request().Context(), existing); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, existing.ToFHIR())
}
