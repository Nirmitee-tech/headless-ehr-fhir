package substance

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

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
	read.GET("/substances", h.ListSubstances)
	read.GET("/substances/:id", h.GetSubstance)

	write := api.Group("", role)
	write.POST("/substances", h.CreateSubstance)
	write.PUT("/substances/:id", h.UpdateSubstance)
	write.DELETE("/substances/:id", h.DeleteSubstance)

	fhirRead := fhirGroup.Group("", role)
	fhirRead.GET("/Substance", h.SearchSubstancesFHIR)
	fhirRead.GET("/Substance/:id", h.GetSubstanceFHIR)

	fhirWrite := fhirGroup.Group("", role)
	fhirWrite.POST("/Substance", h.CreateSubstanceFHIR)
	fhirWrite.PUT("/Substance/:id", h.UpdateSubstanceFHIR)
	fhirWrite.DELETE("/Substance/:id", h.DeleteSubstanceFHIR)
	fhirWrite.PATCH("/Substance/:id", h.PatchSubstanceFHIR)

	fhirRead.POST("/Substance/_search", h.SearchSubstancesFHIR)
	fhirRead.GET("/Substance/:id/_history/:vid", h.VreadSubstanceFHIR)
	fhirRead.GET("/Substance/:id/_history", h.HistorySubstanceFHIR)
}

func (h *Handler) CreateSubstance(c echo.Context) error {
	var s Substance
	if err := c.Bind(&s); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateSubstance(c.Request().Context(), &s); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, s)
}

func (h *Handler) GetSubstance(c echo.Context) error {
	s, err := h.svc.GetSubstance(c.Request().Context(), c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "substance not found")
	}
	return c.JSON(http.StatusOK, s)
}

func (h *Handler) ListSubstances(c echo.Context) error {
	pg := pagination.FromContext(c)
	items, total, err := h.svc.SearchSubstances(c.Request().Context(), nil, pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *Handler) UpdateSubstance(c echo.Context) error {
	var s Substance
	if err := c.Bind(&s); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	s.ID = c.Param("id")
	if err := h.svc.UpdateSubstance(c.Request().Context(), &s); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, s)
}

func (h *Handler) DeleteSubstance(c echo.Context) error {
	if err := h.svc.DeleteSubstance(c.Request().Context(), c.Param("id")); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) SearchSubstancesFHIR(c echo.Context) error {
	pg := pagination.FromContext(c)
	params := map[string]string{}
	for _, k := range []string{"status", "code"} {
		if v := c.QueryParam(k); v != "" {
			params[k] = v
		}
	}
	items, total, err := h.svc.SearchSubstances(c.Request().Context(), params, pg.Limit, pg.Offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	resources := make([]interface{}, len(items))
	for i, item := range items {
		resources[i] = item.ToFHIR()
	}
	return c.JSON(http.StatusOK, fhir.NewSearchBundle(resources, total, "/fhir/Substance"))
}

func (h *Handler) GetSubstanceFHIR(c echo.Context) error {
	s, err := h.svc.GetSubstanceByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Substance", c.Param("id")))
	}
	return c.JSON(http.StatusOK, s.ToFHIR())
}

func (h *Handler) CreateSubstanceFHIR(c echo.Context) error {
	var s Substance
	if err := c.Bind(&s); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	if err := h.svc.CreateSubstance(c.Request().Context(), &s); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	c.Response().Header().Set("Location", "/fhir/Substance/"+s.FHIRID)
	return c.JSON(http.StatusCreated, s.ToFHIR())
}

func (h *Handler) UpdateSubstanceFHIR(c echo.Context) error {
	var s Substance
	if err := c.Bind(&s); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	existing, err := h.svc.GetSubstanceByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Substance", c.Param("id")))
	}
	s.ID = existing.ID
	s.FHIRID = existing.FHIRID
	if err := h.svc.UpdateSubstance(c.Request().Context(), &s); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, s.ToFHIR())
}

func (h *Handler) DeleteSubstanceFHIR(c echo.Context) error {
	existing, err := h.svc.GetSubstanceByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Substance", c.Param("id")))
	}
	if err := h.svc.DeleteSubstance(c.Request().Context(), existing.ID); err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) PatchSubstanceFHIR(c echo.Context) error {
	return h.handlePatch(c, c.Param("id"))
}

func (h *Handler) VreadSubstanceFHIR(c echo.Context) error {
	s, err := h.svc.GetSubstanceByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Substance", c.Param("id")))
	}
	result := s.ToFHIR()
	fhir.SetVersionHeaders(c, 1, s.UpdatedAt.Format("2006-01-02T15:04:05Z"))
	return c.JSON(http.StatusOK, result)
}

func (h *Handler) HistorySubstanceFHIR(c echo.Context) error {
	s, err := h.svc.GetSubstanceByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Substance", c.Param("id")))
	}
	result := s.ToFHIR()
	raw, _ := json.Marshal(result)
	entry := &fhir.HistoryEntry{
		ResourceType: "Substance", ResourceID: s.FHIRID, VersionID: 1,
		Resource: raw, Action: "create", Timestamp: s.CreatedAt,
	}
	return c.JSON(http.StatusOK, fhir.NewHistoryBundle([]*fhir.HistoryEntry{entry}, 1, "/fhir"))
}

func (h *Handler) handlePatch(c echo.Context, fhirID string) error {
	contentType := c.Request().Header.Get("Content-Type")
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome("failed to read request body"))
	}
	existing, err := h.svc.GetSubstanceByFHIRID(c.Request().Context(), fhirID)
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Substance", fhirID))
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
	if err := h.svc.UpdateSubstance(c.Request().Context(), existing); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, existing.ToFHIR())
}
