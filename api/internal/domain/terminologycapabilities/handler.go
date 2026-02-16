package terminologycapabilities

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
	read.GET("/terminology-capabilities", h.ListTerminologyCapabilities)
	read.GET("/terminology-capabilities/:id", h.GetTerminologyCapabilities)

	write := api.Group("", role)
	write.POST("/terminology-capabilities", h.CreateTerminologyCapabilities)
	write.PUT("/terminology-capabilities/:id", h.UpdateTerminologyCapabilities)
	write.DELETE("/terminology-capabilities/:id", h.DeleteTerminologyCapabilities)

	fhirRead := fhirGroup.Group("", role)
	fhirRead.GET("/TerminologyCapabilities", h.SearchTerminologyCapabilitiesFHIR)
	fhirRead.GET("/TerminologyCapabilities/:id", h.GetTerminologyCapabilitiesFHIR)

	fhirWrite := fhirGroup.Group("", role)
	fhirWrite.POST("/TerminologyCapabilities", h.CreateTerminologyCapabilitiesFHIR)
	fhirWrite.PUT("/TerminologyCapabilities/:id", h.UpdateTerminologyCapabilitiesFHIR)
	fhirWrite.DELETE("/TerminologyCapabilities/:id", h.DeleteTerminologyCapabilitiesFHIR)
	fhirWrite.PATCH("/TerminologyCapabilities/:id", h.PatchTerminologyCapabilitiesFHIR)

	fhirRead.POST("/TerminologyCapabilities/_search", h.SearchTerminologyCapabilitiesFHIR)
	fhirRead.GET("/TerminologyCapabilities/:id/_history/:vid", h.VreadTerminologyCapabilitiesFHIR)
	fhirRead.GET("/TerminologyCapabilities/:id/_history", h.HistoryTerminologyCapabilitiesFHIR)
}

// -- REST Endpoints --

func (h *Handler) CreateTerminologyCapabilities(c echo.Context) error {
	var tc TerminologyCapabilities
	if err := c.Bind(&tc); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateTerminologyCapabilities(c.Request().Context(), &tc); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, tc)
}

func (h *Handler) GetTerminologyCapabilities(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	tc, err := h.svc.GetTerminologyCapabilities(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "terminology capabilities not found")
	}
	return c.JSON(http.StatusOK, tc)
}

func (h *Handler) ListTerminologyCapabilities(c echo.Context) error {
	pg := pagination.FromContext(c)
	items, total, err := h.svc.SearchTerminologyCapabilities(c.Request().Context(), nil, pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *Handler) UpdateTerminologyCapabilities(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var tc TerminologyCapabilities
	if err := c.Bind(&tc); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	tc.ID = id
	if err := h.svc.UpdateTerminologyCapabilities(c.Request().Context(), &tc); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, tc)
}

func (h *Handler) DeleteTerminologyCapabilities(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeleteTerminologyCapabilities(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

// -- FHIR Endpoints --

func (h *Handler) SearchTerminologyCapabilitiesFHIR(c echo.Context) error {
	pg := pagination.FromContext(c)
	params := map[string]string{}
	for _, k := range []string{"status", "url", "name"} {
		if v := c.QueryParam(k); v != "" {
			params[k] = v
		}
	}
	items, total, err := h.svc.SearchTerminologyCapabilities(c.Request().Context(), params, pg.Limit, pg.Offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	resources := make([]interface{}, len(items))
	for i, item := range items {
		resources[i] = item.ToFHIR()
	}
	return c.JSON(http.StatusOK, fhir.NewSearchBundle(resources, total, "/fhir/TerminologyCapabilities"))
}

func (h *Handler) GetTerminologyCapabilitiesFHIR(c echo.Context) error {
	tc, err := h.svc.GetTerminologyCapabilitiesByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("TerminologyCapabilities", c.Param("id")))
	}
	return c.JSON(http.StatusOK, tc.ToFHIR())
}

func (h *Handler) CreateTerminologyCapabilitiesFHIR(c echo.Context) error {
	var tc TerminologyCapabilities
	if err := c.Bind(&tc); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	if err := h.svc.CreateTerminologyCapabilities(c.Request().Context(), &tc); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	c.Response().Header().Set("Location", "/fhir/TerminologyCapabilities/"+tc.FHIRID)
	return c.JSON(http.StatusCreated, tc.ToFHIR())
}

func (h *Handler) UpdateTerminologyCapabilitiesFHIR(c echo.Context) error {
	var tc TerminologyCapabilities
	if err := c.Bind(&tc); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	existing, err := h.svc.GetTerminologyCapabilitiesByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("TerminologyCapabilities", c.Param("id")))
	}
	tc.ID = existing.ID
	tc.FHIRID = existing.FHIRID
	if err := h.svc.UpdateTerminologyCapabilities(c.Request().Context(), &tc); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, tc.ToFHIR())
}

func (h *Handler) DeleteTerminologyCapabilitiesFHIR(c echo.Context) error {
	existing, err := h.svc.GetTerminologyCapabilitiesByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("TerminologyCapabilities", c.Param("id")))
	}
	if err := h.svc.DeleteTerminologyCapabilities(c.Request().Context(), existing.ID); err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) PatchTerminologyCapabilitiesFHIR(c echo.Context) error {
	return h.handlePatch(c, c.Param("id"))
}

func (h *Handler) VreadTerminologyCapabilitiesFHIR(c echo.Context) error {
	tc, err := h.svc.GetTerminologyCapabilitiesByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("TerminologyCapabilities", c.Param("id")))
	}
	result := tc.ToFHIR()
	fhir.SetVersionHeaders(c, 1, tc.UpdatedAt.Format("2006-01-02T15:04:05Z"))
	return c.JSON(http.StatusOK, result)
}

func (h *Handler) HistoryTerminologyCapabilitiesFHIR(c echo.Context) error {
	tc, err := h.svc.GetTerminologyCapabilitiesByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("TerminologyCapabilities", c.Param("id")))
	}
	result := tc.ToFHIR()
	raw, _ := json.Marshal(result)
	entry := &fhir.HistoryEntry{
		ResourceType: "TerminologyCapabilities", ResourceID: tc.FHIRID, VersionID: 1,
		Resource: raw, Action: "create", Timestamp: tc.CreatedAt,
	}
	return c.JSON(http.StatusOK, fhir.NewHistoryBundle([]*fhir.HistoryEntry{entry}, 1, "/fhir"))
}

func (h *Handler) handlePatch(c echo.Context, fhirID string) error {
	contentType := c.Request().Header.Get("Content-Type")
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome("failed to read request body"))
	}
	existing, err := h.svc.GetTerminologyCapabilitiesByFHIRID(c.Request().Context(), fhirID)
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("TerminologyCapabilities", fhirID))
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
	if err := h.svc.UpdateTerminologyCapabilities(c.Request().Context(), existing); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, existing.ToFHIR())
}
