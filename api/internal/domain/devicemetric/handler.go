package devicemetric

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
	read.GET("/device-metrics", h.ListDeviceMetrics)
	read.GET("/device-metrics/:id", h.GetDeviceMetric)

	write := api.Group("", role)
	write.POST("/device-metrics", h.CreateDeviceMetric)
	write.PUT("/device-metrics/:id", h.UpdateDeviceMetric)
	write.DELETE("/device-metrics/:id", h.DeleteDeviceMetric)

	fhirRead := fhirGroup.Group("", role)
	fhirRead.GET("/DeviceMetric", h.SearchDeviceMetricsFHIR)
	fhirRead.GET("/DeviceMetric/:id", h.GetDeviceMetricFHIR)

	fhirWrite := fhirGroup.Group("", role)
	fhirWrite.POST("/DeviceMetric", h.CreateDeviceMetricFHIR)
	fhirWrite.PUT("/DeviceMetric/:id", h.UpdateDeviceMetricFHIR)
	fhirWrite.DELETE("/DeviceMetric/:id", h.DeleteDeviceMetricFHIR)
	fhirWrite.PATCH("/DeviceMetric/:id", h.PatchDeviceMetricFHIR)

	fhirRead.POST("/DeviceMetric/_search", h.SearchDeviceMetricsFHIR)
	fhirRead.GET("/DeviceMetric/:id/_history/:vid", h.VreadDeviceMetricFHIR)
	fhirRead.GET("/DeviceMetric/:id/_history", h.HistoryDeviceMetricFHIR)
}

// -- REST Endpoints --

func (h *Handler) CreateDeviceMetric(c echo.Context) error {
	var m DeviceMetric
	if err := c.Bind(&m); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateDeviceMetric(c.Request().Context(), &m); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, m)
}

func (h *Handler) GetDeviceMetric(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	m, err := h.svc.GetDeviceMetric(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "device metric not found")
	}
	return c.JSON(http.StatusOK, m)
}

func (h *Handler) ListDeviceMetrics(c echo.Context) error {
	pg := pagination.FromContext(c)
	items, total, err := h.svc.SearchDeviceMetrics(c.Request().Context(), nil, pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *Handler) UpdateDeviceMetric(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var m DeviceMetric
	if err := c.Bind(&m); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	m.ID = id
	if err := h.svc.UpdateDeviceMetric(c.Request().Context(), &m); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, m)
}

func (h *Handler) DeleteDeviceMetric(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeleteDeviceMetric(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

// -- FHIR Endpoints --

func (h *Handler) SearchDeviceMetricsFHIR(c echo.Context) error {
	pg := pagination.FromContext(c)
	params := map[string]string{}
	for _, k := range []string{"source", "type", "category"} {
		if v := c.QueryParam(k); v != "" {
			params[k] = v
		}
	}
	items, total, err := h.svc.SearchDeviceMetrics(c.Request().Context(), params, pg.Limit, pg.Offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	resources := make([]interface{}, len(items))
	for i, item := range items {
		resources[i] = item.ToFHIR()
	}
	return c.JSON(http.StatusOK, fhir.NewSearchBundle(resources, total, "/fhir/DeviceMetric"))
}

func (h *Handler) GetDeviceMetricFHIR(c echo.Context) error {
	m, err := h.svc.GetDeviceMetricByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("DeviceMetric", c.Param("id")))
	}
	return c.JSON(http.StatusOK, m.ToFHIR())
}

func (h *Handler) CreateDeviceMetricFHIR(c echo.Context) error {
	var m DeviceMetric
	if err := c.Bind(&m); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	if err := h.svc.CreateDeviceMetric(c.Request().Context(), &m); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	c.Response().Header().Set("Location", "/fhir/DeviceMetric/"+m.FHIRID)
	return c.JSON(http.StatusCreated, m.ToFHIR())
}

func (h *Handler) UpdateDeviceMetricFHIR(c echo.Context) error {
	var m DeviceMetric
	if err := c.Bind(&m); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	existing, err := h.svc.GetDeviceMetricByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("DeviceMetric", c.Param("id")))
	}
	m.ID = existing.ID
	m.FHIRID = existing.FHIRID
	if err := h.svc.UpdateDeviceMetric(c.Request().Context(), &m); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, m.ToFHIR())
}

func (h *Handler) DeleteDeviceMetricFHIR(c echo.Context) error {
	existing, err := h.svc.GetDeviceMetricByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("DeviceMetric", c.Param("id")))
	}
	if err := h.svc.DeleteDeviceMetric(c.Request().Context(), existing.ID); err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) PatchDeviceMetricFHIR(c echo.Context) error {
	return h.handlePatch(c, c.Param("id"))
}

func (h *Handler) VreadDeviceMetricFHIR(c echo.Context) error {
	m, err := h.svc.GetDeviceMetricByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("DeviceMetric", c.Param("id")))
	}
	result := m.ToFHIR()
	fhir.SetVersionHeaders(c, 1, m.UpdatedAt.Format("2006-01-02T15:04:05Z"))
	return c.JSON(http.StatusOK, result)
}

func (h *Handler) HistoryDeviceMetricFHIR(c echo.Context) error {
	m, err := h.svc.GetDeviceMetricByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("DeviceMetric", c.Param("id")))
	}
	result := m.ToFHIR()
	raw, _ := json.Marshal(result)
	entry := &fhir.HistoryEntry{
		ResourceType: "DeviceMetric", ResourceID: m.FHIRID, VersionID: 1,
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
	existing, err := h.svc.GetDeviceMetricByFHIRID(c.Request().Context(), fhirID)
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("DeviceMetric", fhirID))
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
	if v, ok := patched["category"].(string); ok {
		existing.Category = v
	}
	if err := h.svc.UpdateDeviceMetric(c.Request().Context(), existing); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, existing.ToFHIR())
}
