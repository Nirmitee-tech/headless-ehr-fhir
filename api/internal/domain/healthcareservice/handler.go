package healthcareservice

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
	role := auth.RequireRole("admin")

	read := api.Group("", role)
	read.GET("/healthcare-services", h.ListHealthcareServices)
	read.GET("/healthcare-services/:id", h.GetHealthcareService)

	write := api.Group("", role)
	write.POST("/healthcare-services", h.CreateHealthcareService)
	write.PUT("/healthcare-services/:id", h.UpdateHealthcareService)
	write.DELETE("/healthcare-services/:id", h.DeleteHealthcareService)

	fhirRead := fhirGroup.Group("", role)
	fhirRead.GET("/HealthcareService", h.SearchHealthcareServicesFHIR)
	fhirRead.GET("/HealthcareService/:id", h.GetHealthcareServiceFHIR)

	fhirWrite := fhirGroup.Group("", role)
	fhirWrite.POST("/HealthcareService", h.CreateHealthcareServiceFHIR)
	fhirWrite.PUT("/HealthcareService/:id", h.UpdateHealthcareServiceFHIR)
	fhirWrite.DELETE("/HealthcareService/:id", h.DeleteHealthcareServiceFHIR)
	fhirWrite.PATCH("/HealthcareService/:id", h.PatchHealthcareServiceFHIR)

	fhirRead.POST("/HealthcareService/_search", h.SearchHealthcareServicesFHIR)

	fhirRead.GET("/HealthcareService/:id/_history/:vid", h.VreadHealthcareServiceFHIR)
	fhirRead.GET("/HealthcareService/:id/_history", h.HistoryHealthcareServiceFHIR)
}

// -- REST Endpoints --

func (h *Handler) CreateHealthcareService(c echo.Context) error {
	var hs HealthcareService
	if err := c.Bind(&hs); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateHealthcareService(c.Request().Context(), &hs); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, hs)
}

func (h *Handler) GetHealthcareService(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	hs, err := h.svc.GetHealthcareService(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "healthcare service not found")
	}
	return c.JSON(http.StatusOK, hs)
}

func (h *Handler) ListHealthcareServices(c echo.Context) error {
	pg := pagination.FromContext(c)
	params := fhir.ExtractSearchParams(c)
	items, total, err := h.svc.SearchHealthcareServices(c.Request().Context(), params, pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *Handler) UpdateHealthcareService(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var hs HealthcareService
	if err := c.Bind(&hs); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	hs.ID = id
	if err := h.svc.UpdateHealthcareService(c.Request().Context(), &hs); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, hs)
}

func (h *Handler) DeleteHealthcareService(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeleteHealthcareService(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

// -- FHIR Endpoints --

func (h *Handler) SearchHealthcareServicesFHIR(c echo.Context) error {
	pg := pagination.FromContext(c)
	params := fhir.ExtractSearchParams(c)
	items, total, err := h.svc.SearchHealthcareServices(c.Request().Context(), params, pg.Limit, pg.Offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	resources := make([]interface{}, len(items))
	for i, item := range items {
		resources[i] = item.ToFHIR()
	}
	return c.JSON(http.StatusOK, fhir.NewSearchBundle(resources, total, "/fhir/HealthcareService"))
}

func (h *Handler) GetHealthcareServiceFHIR(c echo.Context) error {
	hs, err := h.svc.GetHealthcareServiceByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("HealthcareService", c.Param("id")))
	}
	return c.JSON(http.StatusOK, hs.ToFHIR())
}

func (h *Handler) CreateHealthcareServiceFHIR(c echo.Context) error {
	var hs HealthcareService
	if err := c.Bind(&hs); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	if err := h.svc.CreateHealthcareService(c.Request().Context(), &hs); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	c.Response().Header().Set("Location", "/fhir/HealthcareService/"+hs.FHIRID)
	return c.JSON(http.StatusCreated, hs.ToFHIR())
}

func (h *Handler) UpdateHealthcareServiceFHIR(c echo.Context) error {
	var hs HealthcareService
	if err := c.Bind(&hs); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	existing, err := h.svc.GetHealthcareServiceByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("HealthcareService", c.Param("id")))
	}
	hs.ID = existing.ID
	hs.FHIRID = existing.FHIRID
	if err := h.svc.UpdateHealthcareService(c.Request().Context(), &hs); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, hs.ToFHIR())
}

func (h *Handler) DeleteHealthcareServiceFHIR(c echo.Context) error {
	existing, err := h.svc.GetHealthcareServiceByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("HealthcareService", c.Param("id")))
	}
	if err := h.svc.DeleteHealthcareService(c.Request().Context(), existing.ID); err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) PatchHealthcareServiceFHIR(c echo.Context) error {
	return h.handlePatch(c, c.Param("id"))
}

func (h *Handler) VreadHealthcareServiceFHIR(c echo.Context) error {
	hs, err := h.svc.GetHealthcareServiceByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("HealthcareService", c.Param("id")))
	}
	result := hs.ToFHIR()
	fhir.SetVersionHeaders(c, 1, hs.UpdatedAt.Format("2006-01-02T15:04:05Z"))
	return c.JSON(http.StatusOK, result)
}

func (h *Handler) HistoryHealthcareServiceFHIR(c echo.Context) error {
	hs, err := h.svc.GetHealthcareServiceByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("HealthcareService", c.Param("id")))
	}
	result := hs.ToFHIR()
	raw, _ := json.Marshal(result)
	entry := &fhir.HistoryEntry{
		ResourceType: "HealthcareService", ResourceID: hs.FHIRID, VersionID: 1,
		Resource: raw, Action: "create", Timestamp: hs.CreatedAt,
	}
	return c.JSON(http.StatusOK, fhir.NewHistoryBundle([]*fhir.HistoryEntry{entry}, 1, "/fhir"))
}

func (h *Handler) handlePatch(c echo.Context, fhirID string) error {
	contentType := c.Request().Header.Get("Content-Type")
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome("failed to read request body"))
	}

	existing, err := h.svc.GetHealthcareServiceByFHIRID(c.Request().Context(), fhirID)
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("HealthcareService", fhirID))
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

	if v, ok := patched["active"].(bool); ok {
		existing.Active = v
	}
	if v, ok := patched["name"].(string); ok {
		existing.Name = v
	}
	if err := h.svc.UpdateHealthcareService(c.Request().Context(), existing); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, existing.ToFHIR())
}
