package effectevidencesynthesis

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
	read.GET("/effect-evidence-syntheses", h.ListEffectEvidenceSyntheses)
	read.GET("/effect-evidence-syntheses/:id", h.GetEffectEvidenceSynthesis)

	write := api.Group("", role)
	write.POST("/effect-evidence-syntheses", h.CreateEffectEvidenceSynthesis)
	write.PUT("/effect-evidence-syntheses/:id", h.UpdateEffectEvidenceSynthesis)
	write.DELETE("/effect-evidence-syntheses/:id", h.DeleteEffectEvidenceSynthesis)

	fhirRead := fhirGroup.Group("", role)
	fhirRead.GET("/EffectEvidenceSynthesis", h.SearchEffectEvidenceSynthesesFHIR)
	fhirRead.GET("/EffectEvidenceSynthesis/:id", h.GetEffectEvidenceSynthesisFHIR)

	fhirWrite := fhirGroup.Group("", role)
	fhirWrite.POST("/EffectEvidenceSynthesis", h.CreateEffectEvidenceSynthesisFHIR)
	fhirWrite.PUT("/EffectEvidenceSynthesis/:id", h.UpdateEffectEvidenceSynthesisFHIR)
	fhirWrite.DELETE("/EffectEvidenceSynthesis/:id", h.DeleteEffectEvidenceSynthesisFHIR)
	fhirWrite.PATCH("/EffectEvidenceSynthesis/:id", h.PatchEffectEvidenceSynthesisFHIR)

	fhirRead.POST("/EffectEvidenceSynthesis/_search", h.SearchEffectEvidenceSynthesesFHIR)
	fhirRead.GET("/EffectEvidenceSynthesis/:id/_history/:vid", h.VreadEffectEvidenceSynthesisFHIR)
	fhirRead.GET("/EffectEvidenceSynthesis/:id/_history", h.HistoryEffectEvidenceSynthesisFHIR)
}

// -- REST Endpoints --

func (h *Handler) CreateEffectEvidenceSynthesis(c echo.Context) error {
	var e EffectEvidenceSynthesis
	if err := c.Bind(&e); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateEffectEvidenceSynthesis(c.Request().Context(), &e); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, e)
}

func (h *Handler) GetEffectEvidenceSynthesis(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	e, err := h.svc.GetEffectEvidenceSynthesis(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "effect evidence synthesis not found")
	}
	return c.JSON(http.StatusOK, e)
}

func (h *Handler) ListEffectEvidenceSyntheses(c echo.Context) error {
	pg := pagination.FromContext(c)
	items, total, err := h.svc.SearchEffectEvidenceSyntheses(c.Request().Context(), nil, pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *Handler) UpdateEffectEvidenceSynthesis(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var e EffectEvidenceSynthesis
	if err := c.Bind(&e); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	e.ID = id
	if err := h.svc.UpdateEffectEvidenceSynthesis(c.Request().Context(), &e); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, e)
}

func (h *Handler) DeleteEffectEvidenceSynthesis(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeleteEffectEvidenceSynthesis(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

// -- FHIR Endpoints --

func (h *Handler) SearchEffectEvidenceSynthesesFHIR(c echo.Context) error {
	pg := pagination.FromContext(c)
	params := fhir.ExtractSearchParams(c)
	items, total, err := h.svc.SearchEffectEvidenceSyntheses(c.Request().Context(), params, pg.Limit, pg.Offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	resources := make([]interface{}, len(items))
	for i, item := range items {
		resources[i] = item.ToFHIR()
	}
	return c.JSON(http.StatusOK, fhir.NewSearchBundle(resources, total, "/fhir/EffectEvidenceSynthesis"))
}

func (h *Handler) GetEffectEvidenceSynthesisFHIR(c echo.Context) error {
	e, err := h.svc.GetEffectEvidenceSynthesisByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("EffectEvidenceSynthesis", c.Param("id")))
	}
	fhir.SetVersionHeaders(c, 1, e.UpdatedAt.Format("2006-01-02T15:04:05Z"))
	return c.JSON(http.StatusOK, e.ToFHIR())
}

func (h *Handler) CreateEffectEvidenceSynthesisFHIR(c echo.Context) error {
	var e EffectEvidenceSynthesis
	if err := c.Bind(&e); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	if err := h.svc.CreateEffectEvidenceSynthesis(c.Request().Context(), &e); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	c.Response().Header().Set("Location", "/fhir/EffectEvidenceSynthesis/"+e.FHIRID)
	return c.JSON(http.StatusCreated, e.ToFHIR())
}

func (h *Handler) UpdateEffectEvidenceSynthesisFHIR(c echo.Context) error {
	var e EffectEvidenceSynthesis
	if err := c.Bind(&e); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	existing, err := h.svc.GetEffectEvidenceSynthesisByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("EffectEvidenceSynthesis", c.Param("id")))
	}
	e.ID = existing.ID
	e.FHIRID = existing.FHIRID
	if err := h.svc.UpdateEffectEvidenceSynthesis(c.Request().Context(), &e); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, e.ToFHIR())
}

func (h *Handler) DeleteEffectEvidenceSynthesisFHIR(c echo.Context) error {
	existing, err := h.svc.GetEffectEvidenceSynthesisByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("EffectEvidenceSynthesis", c.Param("id")))
	}
	if err := h.svc.DeleteEffectEvidenceSynthesis(c.Request().Context(), existing.ID); err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) PatchEffectEvidenceSynthesisFHIR(c echo.Context) error {
	return h.handlePatch(c, c.Param("id"))
}

func (h *Handler) VreadEffectEvidenceSynthesisFHIR(c echo.Context) error {
	e, err := h.svc.GetEffectEvidenceSynthesisByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("EffectEvidenceSynthesis", c.Param("id")))
	}
	result := e.ToFHIR()
	fhir.SetVersionHeaders(c, 1, e.UpdatedAt.Format("2006-01-02T15:04:05Z"))
	return c.JSON(http.StatusOK, result)
}

func (h *Handler) HistoryEffectEvidenceSynthesisFHIR(c echo.Context) error {
	e, err := h.svc.GetEffectEvidenceSynthesisByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("EffectEvidenceSynthesis", c.Param("id")))
	}
	result := e.ToFHIR()
	raw, _ := json.Marshal(result)
	entry := &fhir.HistoryEntry{
		ResourceType: "EffectEvidenceSynthesis", ResourceID: e.FHIRID, VersionID: 1,
		Resource: raw, Action: "create", Timestamp: e.CreatedAt,
	}
	return c.JSON(http.StatusOK, fhir.NewHistoryBundle([]*fhir.HistoryEntry{entry}, 1, "/fhir"))
}

func (h *Handler) handlePatch(c echo.Context, fhirID string) error {
	contentType := c.Request().Header.Get("Content-Type")
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome("failed to read request body"))
	}
	existing, err := h.svc.GetEffectEvidenceSynthesisByFHIRID(c.Request().Context(), fhirID)
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("EffectEvidenceSynthesis", fhirID))
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
	if err := h.svc.UpdateEffectEvidenceSynthesis(c.Request().Context(), existing); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, existing.ToFHIR())
}
