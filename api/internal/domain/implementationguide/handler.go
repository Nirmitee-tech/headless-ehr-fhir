package implementationguide

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
	read.GET("/implementation-guides", h.ListImplementationGuides)
	read.GET("/implementation-guides/:id", h.GetImplementationGuide)

	write := api.Group("", role)
	write.POST("/implementation-guides", h.CreateImplementationGuide)
	write.PUT("/implementation-guides/:id", h.UpdateImplementationGuide)
	write.DELETE("/implementation-guides/:id", h.DeleteImplementationGuide)

	fhirRead := fhirGroup.Group("", role)
	fhirRead.GET("/ImplementationGuide", h.SearchImplementationGuidesFHIR)
	fhirRead.GET("/ImplementationGuide/:id", h.GetImplementationGuideFHIR)

	fhirWrite := fhirGroup.Group("", role)
	fhirWrite.POST("/ImplementationGuide", h.CreateImplementationGuideFHIR)
	fhirWrite.PUT("/ImplementationGuide/:id", h.UpdateImplementationGuideFHIR)
	fhirWrite.DELETE("/ImplementationGuide/:id", h.DeleteImplementationGuideFHIR)
	fhirWrite.PATCH("/ImplementationGuide/:id", h.PatchImplementationGuideFHIR)

	fhirRead.POST("/ImplementationGuide/_search", h.SearchImplementationGuidesFHIR)
	fhirRead.GET("/ImplementationGuide/:id/_history/:vid", h.VreadImplementationGuideFHIR)
	fhirRead.GET("/ImplementationGuide/:id/_history", h.HistoryImplementationGuideFHIR)
}

// -- REST Endpoints --

func (h *Handler) CreateImplementationGuide(c echo.Context) error {
	var ig ImplementationGuide
	if err := c.Bind(&ig); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateImplementationGuide(c.Request().Context(), &ig); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, ig)
}

func (h *Handler) GetImplementationGuide(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	ig, err := h.svc.GetImplementationGuide(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "implementation guide not found")
	}
	return c.JSON(http.StatusOK, ig)
}

func (h *Handler) ListImplementationGuides(c echo.Context) error {
	pg := pagination.FromContext(c)
	items, total, err := h.svc.SearchImplementationGuides(c.Request().Context(), nil, pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *Handler) UpdateImplementationGuide(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var ig ImplementationGuide
	if err := c.Bind(&ig); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	ig.ID = id
	if err := h.svc.UpdateImplementationGuide(c.Request().Context(), &ig); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, ig)
}

func (h *Handler) DeleteImplementationGuide(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeleteImplementationGuide(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

// -- FHIR Endpoints --

func (h *Handler) SearchImplementationGuidesFHIR(c echo.Context) error {
	pg := pagination.FromContext(c)
	params := map[string]string{}
	for _, k := range []string{"status", "url", "name"} {
		if v := c.QueryParam(k); v != "" {
			params[k] = v
		}
	}
	items, total, err := h.svc.SearchImplementationGuides(c.Request().Context(), params, pg.Limit, pg.Offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	resources := make([]interface{}, len(items))
	for i, item := range items {
		resources[i] = item.ToFHIR()
	}
	return c.JSON(http.StatusOK, fhir.NewSearchBundle(resources, total, "/fhir/ImplementationGuide"))
}

func (h *Handler) GetImplementationGuideFHIR(c echo.Context) error {
	ig, err := h.svc.GetImplementationGuideByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("ImplementationGuide", c.Param("id")))
	}
	return c.JSON(http.StatusOK, ig.ToFHIR())
}

func (h *Handler) CreateImplementationGuideFHIR(c echo.Context) error {
	var ig ImplementationGuide
	if err := c.Bind(&ig); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	if err := h.svc.CreateImplementationGuide(c.Request().Context(), &ig); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	c.Response().Header().Set("Location", "/fhir/ImplementationGuide/"+ig.FHIRID)
	return c.JSON(http.StatusCreated, ig.ToFHIR())
}

func (h *Handler) UpdateImplementationGuideFHIR(c echo.Context) error {
	var ig ImplementationGuide
	if err := c.Bind(&ig); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	existing, err := h.svc.GetImplementationGuideByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("ImplementationGuide", c.Param("id")))
	}
	ig.ID = existing.ID
	ig.FHIRID = existing.FHIRID
	if err := h.svc.UpdateImplementationGuide(c.Request().Context(), &ig); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, ig.ToFHIR())
}

func (h *Handler) DeleteImplementationGuideFHIR(c echo.Context) error {
	existing, err := h.svc.GetImplementationGuideByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("ImplementationGuide", c.Param("id")))
	}
	if err := h.svc.DeleteImplementationGuide(c.Request().Context(), existing.ID); err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) PatchImplementationGuideFHIR(c echo.Context) error {
	return h.handlePatch(c, c.Param("id"))
}

func (h *Handler) VreadImplementationGuideFHIR(c echo.Context) error {
	ig, err := h.svc.GetImplementationGuideByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("ImplementationGuide", c.Param("id")))
	}
	result := ig.ToFHIR()
	fhir.SetVersionHeaders(c, 1, ig.UpdatedAt.Format("2006-01-02T15:04:05Z"))
	return c.JSON(http.StatusOK, result)
}

func (h *Handler) HistoryImplementationGuideFHIR(c echo.Context) error {
	ig, err := h.svc.GetImplementationGuideByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("ImplementationGuide", c.Param("id")))
	}
	result := ig.ToFHIR()
	raw, _ := json.Marshal(result)
	entry := &fhir.HistoryEntry{
		ResourceType: "ImplementationGuide", ResourceID: ig.FHIRID, VersionID: 1,
		Resource: raw, Action: "create", Timestamp: ig.CreatedAt,
	}
	return c.JSON(http.StatusOK, fhir.NewHistoryBundle([]*fhir.HistoryEntry{entry}, 1, "/fhir"))
}

func (h *Handler) handlePatch(c echo.Context, fhirID string) error {
	contentType := c.Request().Header.Get("Content-Type")
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome("failed to read request body"))
	}
	existing, err := h.svc.GetImplementationGuideByFHIRID(c.Request().Context(), fhirID)
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("ImplementationGuide", fhirID))
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
	if err := h.svc.UpdateImplementationGuide(c.Request().Context(), existing); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, existing.ToFHIR())
}
