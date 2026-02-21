package valueset

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
	read.GET("/value-sets", h.ListValueSets)
	read.GET("/value-sets/:id", h.GetValueSet)

	write := api.Group("", role)
	write.POST("/value-sets", h.CreateValueSet)
	write.PUT("/value-sets/:id", h.UpdateValueSet)
	write.DELETE("/value-sets/:id", h.DeleteValueSet)

	fhirRead := fhirGroup.Group("", role)
	fhirRead.GET("/ValueSet", h.SearchValueSetsFHIR)
	fhirRead.GET("/ValueSet/:id", h.GetValueSetFHIR)

	fhirWrite := fhirGroup.Group("", role)
	fhirWrite.POST("/ValueSet", h.CreateValueSetFHIR)
	fhirWrite.PUT("/ValueSet/:id", h.UpdateValueSetFHIR)
	fhirWrite.DELETE("/ValueSet/:id", h.DeleteValueSetFHIR)
	fhirWrite.PATCH("/ValueSet/:id", h.PatchValueSetFHIR)

	fhirRead.POST("/ValueSet/_search", h.SearchValueSetsFHIR)
	fhirRead.GET("/ValueSet/:id/_history/:vid", h.VreadValueSetFHIR)
	fhirRead.GET("/ValueSet/:id/_history", h.HistoryValueSetFHIR)
}

// -- REST Endpoints --

func (h *Handler) CreateValueSet(c echo.Context) error {
	var vs ValueSet
	if err := c.Bind(&vs); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateValueSet(c.Request().Context(), &vs); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, vs)
}

func (h *Handler) GetValueSet(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	vs, err := h.svc.GetValueSet(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "value set not found")
	}
	return c.JSON(http.StatusOK, vs)
}

func (h *Handler) ListValueSets(c echo.Context) error {
	pg := pagination.FromContext(c)
	items, total, err := h.svc.SearchValueSets(c.Request().Context(), nil, pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *Handler) UpdateValueSet(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var vs ValueSet
	if err := c.Bind(&vs); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	vs.ID = id
	if err := h.svc.UpdateValueSet(c.Request().Context(), &vs); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, vs)
}

func (h *Handler) DeleteValueSet(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeleteValueSet(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

// -- FHIR Endpoints --

func (h *Handler) SearchValueSetsFHIR(c echo.Context) error {
	pg := pagination.FromContext(c)
	params := fhir.ExtractSearchParams(c)
	items, total, err := h.svc.SearchValueSets(c.Request().Context(), params, pg.Limit, pg.Offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	resources := make([]interface{}, len(items))
	for i, item := range items {
		resources[i] = item.ToFHIR()
	}
	return c.JSON(http.StatusOK, fhir.NewSearchBundle(resources, total, "/fhir/ValueSet"))
}

func (h *Handler) GetValueSetFHIR(c echo.Context) error {
	vs, err := h.svc.GetValueSetByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("ValueSet", c.Param("id")))
	}
	fhir.SetVersionHeaders(c, 1, vs.UpdatedAt.Format("2006-01-02T15:04:05Z"))
	return c.JSON(http.StatusOK, vs.ToFHIR())
}

func (h *Handler) CreateValueSetFHIR(c echo.Context) error {
	var vs ValueSet
	if err := c.Bind(&vs); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	if err := h.svc.CreateValueSet(c.Request().Context(), &vs); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	c.Response().Header().Set("Location", "/fhir/ValueSet/"+vs.FHIRID)
	return c.JSON(http.StatusCreated, vs.ToFHIR())
}

func (h *Handler) UpdateValueSetFHIR(c echo.Context) error {
	var vs ValueSet
	if err := c.Bind(&vs); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	existing, err := h.svc.GetValueSetByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("ValueSet", c.Param("id")))
	}
	vs.ID = existing.ID
	vs.FHIRID = existing.FHIRID
	if err := h.svc.UpdateValueSet(c.Request().Context(), &vs); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, vs.ToFHIR())
}

func (h *Handler) DeleteValueSetFHIR(c echo.Context) error {
	existing, err := h.svc.GetValueSetByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("ValueSet", c.Param("id")))
	}
	if err := h.svc.DeleteValueSet(c.Request().Context(), existing.ID); err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) PatchValueSetFHIR(c echo.Context) error {
	return h.handlePatch(c, c.Param("id"))
}

func (h *Handler) VreadValueSetFHIR(c echo.Context) error {
	vs, err := h.svc.GetValueSetByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("ValueSet", c.Param("id")))
	}
	result := vs.ToFHIR()
	fhir.SetVersionHeaders(c, 1, vs.UpdatedAt.Format("2006-01-02T15:04:05Z"))
	return c.JSON(http.StatusOK, result)
}

func (h *Handler) HistoryValueSetFHIR(c echo.Context) error {
	vs, err := h.svc.GetValueSetByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("ValueSet", c.Param("id")))
	}
	result := vs.ToFHIR()
	raw, _ := json.Marshal(result)
	entry := &fhir.HistoryEntry{
		ResourceType: "ValueSet", ResourceID: vs.FHIRID, VersionID: 1,
		Resource: raw, Action: "create", Timestamp: vs.CreatedAt,
	}
	return c.JSON(http.StatusOK, fhir.NewHistoryBundle([]*fhir.HistoryEntry{entry}, 1, "/fhir"))
}

func (h *Handler) handlePatch(c echo.Context, fhirID string) error {
	contentType := c.Request().Header.Get("Content-Type")
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome("failed to read request body"))
	}
	existing, err := h.svc.GetValueSetByFHIRID(c.Request().Context(), fhirID)
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("ValueSet", fhirID))
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
	if err := h.svc.UpdateValueSet(c.Request().Context(), existing); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, existing.ToFHIR())
}
