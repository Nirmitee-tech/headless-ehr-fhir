package devicerequest

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
	read.GET("/device-requests", h.ListDeviceRequests)
	read.GET("/device-requests/:id", h.GetDeviceRequest)

	write := api.Group("", role)
	write.POST("/device-requests", h.CreateDeviceRequest)
	write.PUT("/device-requests/:id", h.UpdateDeviceRequest)
	write.DELETE("/device-requests/:id", h.DeleteDeviceRequest)

	fhirRead := fhirGroup.Group("", role)
	fhirRead.GET("/DeviceRequest", h.SearchDeviceRequestsFHIR)
	fhirRead.GET("/DeviceRequest/:id", h.GetDeviceRequestFHIR)

	fhirWrite := fhirGroup.Group("", role)
	fhirWrite.POST("/DeviceRequest", h.CreateDeviceRequestFHIR)
	fhirWrite.PUT("/DeviceRequest/:id", h.UpdateDeviceRequestFHIR)
	fhirWrite.DELETE("/DeviceRequest/:id", h.DeleteDeviceRequestFHIR)
	fhirWrite.PATCH("/DeviceRequest/:id", h.PatchDeviceRequestFHIR)

	fhirRead.POST("/DeviceRequest/_search", h.SearchDeviceRequestsFHIR)
	fhirRead.GET("/DeviceRequest/:id/_history/:vid", h.VreadDeviceRequestFHIR)
	fhirRead.GET("/DeviceRequest/:id/_history", h.HistoryDeviceRequestFHIR)
}

func (h *Handler) CreateDeviceRequest(c echo.Context) error {
	var d DeviceRequest
	if err := c.Bind(&d); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateDeviceRequest(c.Request().Context(), &d); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, d)
}

func (h *Handler) GetDeviceRequest(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	d, err := h.svc.GetDeviceRequest(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "device request not found")
	}
	return c.JSON(http.StatusOK, d)
}

func (h *Handler) ListDeviceRequests(c echo.Context) error {
	pg := pagination.FromContext(c)
	items, total, err := h.svc.SearchDeviceRequests(c.Request().Context(), nil, pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *Handler) UpdateDeviceRequest(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var d DeviceRequest
	if err := c.Bind(&d); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	d.ID = id
	if err := h.svc.UpdateDeviceRequest(c.Request().Context(), &d); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, d)
}

func (h *Handler) DeleteDeviceRequest(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeleteDeviceRequest(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) SearchDeviceRequestsFHIR(c echo.Context) error {
	pg := pagination.FromContext(c)
	params := map[string]string{}
	for _, k := range []string{"patient", "status", "intent"} {
		if v := c.QueryParam(k); v != "" {
			params[k] = v
		}
	}
	items, total, err := h.svc.SearchDeviceRequests(c.Request().Context(), params, pg.Limit, pg.Offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	resources := make([]interface{}, len(items))
	for i, item := range items {
		resources[i] = item.ToFHIR()
	}
	return c.JSON(http.StatusOK, fhir.NewSearchBundle(resources, total, "/fhir/DeviceRequest"))
}

func (h *Handler) GetDeviceRequestFHIR(c echo.Context) error {
	d, err := h.svc.GetDeviceRequestByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("DeviceRequest", c.Param("id")))
	}
	return c.JSON(http.StatusOK, d.ToFHIR())
}

func (h *Handler) CreateDeviceRequestFHIR(c echo.Context) error {
	var d DeviceRequest
	if err := c.Bind(&d); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	if err := h.svc.CreateDeviceRequest(c.Request().Context(), &d); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	c.Response().Header().Set("Location", "/fhir/DeviceRequest/"+d.FHIRID)
	return c.JSON(http.StatusCreated, d.ToFHIR())
}

func (h *Handler) UpdateDeviceRequestFHIR(c echo.Context) error {
	var d DeviceRequest
	if err := c.Bind(&d); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	existing, err := h.svc.GetDeviceRequestByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("DeviceRequest", c.Param("id")))
	}
	d.ID = existing.ID
	d.FHIRID = existing.FHIRID
	if err := h.svc.UpdateDeviceRequest(c.Request().Context(), &d); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, d.ToFHIR())
}

func (h *Handler) DeleteDeviceRequestFHIR(c echo.Context) error {
	existing, err := h.svc.GetDeviceRequestByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("DeviceRequest", c.Param("id")))
	}
	if err := h.svc.DeleteDeviceRequest(c.Request().Context(), existing.ID); err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) PatchDeviceRequestFHIR(c echo.Context) error {
	return h.handlePatch(c, c.Param("id"))
}

func (h *Handler) VreadDeviceRequestFHIR(c echo.Context) error {
	d, err := h.svc.GetDeviceRequestByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("DeviceRequest", c.Param("id")))
	}
	result := d.ToFHIR()
	fhir.SetVersionHeaders(c, 1, d.UpdatedAt.Format("2006-01-02T15:04:05Z"))
	return c.JSON(http.StatusOK, result)
}

func (h *Handler) HistoryDeviceRequestFHIR(c echo.Context) error {
	d, err := h.svc.GetDeviceRequestByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("DeviceRequest", c.Param("id")))
	}
	result := d.ToFHIR()
	raw, _ := json.Marshal(result)
	entry := &fhir.HistoryEntry{
		ResourceType: "DeviceRequest", ResourceID: d.FHIRID, VersionID: 1,
		Resource: raw, Action: "create", Timestamp: d.CreatedAt,
	}
	return c.JSON(http.StatusOK, fhir.NewHistoryBundle([]*fhir.HistoryEntry{entry}, 1, "/fhir"))
}

func (h *Handler) handlePatch(c echo.Context, fhirID string) error {
	contentType := c.Request().Header.Get("Content-Type")
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome("failed to read request body"))
	}
	existing, err := h.svc.GetDeviceRequestByFHIRID(c.Request().Context(), fhirID)
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("DeviceRequest", fhirID))
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
	if err := h.svc.UpdateDeviceRequest(c.Request().Context(), existing); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, existing.ToFHIR())
}
