package searchparameter

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
	read.GET("/search-parameters", h.ListSearchParameters)
	read.GET("/search-parameters/:id", h.GetSearchParameter)

	write := api.Group("", role)
	write.POST("/search-parameters", h.CreateSearchParameter)
	write.PUT("/search-parameters/:id", h.UpdateSearchParameter)
	write.DELETE("/search-parameters/:id", h.DeleteSearchParameter)

	fhirRead := fhirGroup.Group("", role)
	fhirRead.GET("/SearchParameter", h.SearchSearchParametersFHIR)
	fhirRead.GET("/SearchParameter/:id", h.GetSearchParameterFHIR)

	fhirWrite := fhirGroup.Group("", role)
	fhirWrite.POST("/SearchParameter", h.CreateSearchParameterFHIR)
	fhirWrite.PUT("/SearchParameter/:id", h.UpdateSearchParameterFHIR)
	fhirWrite.DELETE("/SearchParameter/:id", h.DeleteSearchParameterFHIR)
	fhirWrite.PATCH("/SearchParameter/:id", h.PatchSearchParameterFHIR)

	fhirRead.POST("/SearchParameter/_search", h.SearchSearchParametersFHIR)
	fhirRead.GET("/SearchParameter/:id/_history/:vid", h.VreadSearchParameterFHIR)
	fhirRead.GET("/SearchParameter/:id/_history", h.HistorySearchParameterFHIR)
}

// -- REST Endpoints --

func (h *Handler) CreateSearchParameter(c echo.Context) error {
	var sp SearchParameter
	if err := c.Bind(&sp); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateSearchParameter(c.Request().Context(), &sp); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, sp)
}

func (h *Handler) GetSearchParameter(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	sp, err := h.svc.GetSearchParameter(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "search parameter not found")
	}
	return c.JSON(http.StatusOK, sp)
}

func (h *Handler) ListSearchParameters(c echo.Context) error {
	pg := pagination.FromContext(c)
	items, total, err := h.svc.SearchSearchParameters(c.Request().Context(), nil, pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *Handler) UpdateSearchParameter(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var sp SearchParameter
	if err := c.Bind(&sp); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	sp.ID = id
	if err := h.svc.UpdateSearchParameter(c.Request().Context(), &sp); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, sp)
}

func (h *Handler) DeleteSearchParameter(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeleteSearchParameter(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

// -- FHIR Endpoints --

func (h *Handler) SearchSearchParametersFHIR(c echo.Context) error {
	pg := pagination.FromContext(c)
	params := fhir.ExtractSearchParams(c)
	items, total, err := h.svc.SearchSearchParameters(c.Request().Context(), params, pg.Limit, pg.Offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	resources := make([]interface{}, len(items))
	for i, item := range items {
		resources[i] = item.ToFHIR()
	}
	return c.JSON(http.StatusOK, fhir.NewSearchBundle(resources, total, "/fhir/SearchParameter"))
}

func (h *Handler) GetSearchParameterFHIR(c echo.Context) error {
	sp, err := h.svc.GetSearchParameterByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("SearchParameter", c.Param("id")))
	}
	return c.JSON(http.StatusOK, sp.ToFHIR())
}

func (h *Handler) CreateSearchParameterFHIR(c echo.Context) error {
	var sp SearchParameter
	if err := c.Bind(&sp); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	if err := h.svc.CreateSearchParameter(c.Request().Context(), &sp); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	c.Response().Header().Set("Location", "/fhir/SearchParameter/"+sp.FHIRID)
	return c.JSON(http.StatusCreated, sp.ToFHIR())
}

func (h *Handler) UpdateSearchParameterFHIR(c echo.Context) error {
	var sp SearchParameter
	if err := c.Bind(&sp); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	existing, err := h.svc.GetSearchParameterByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("SearchParameter", c.Param("id")))
	}
	sp.ID = existing.ID
	sp.FHIRID = existing.FHIRID
	if err := h.svc.UpdateSearchParameter(c.Request().Context(), &sp); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, sp.ToFHIR())
}

func (h *Handler) DeleteSearchParameterFHIR(c echo.Context) error {
	existing, err := h.svc.GetSearchParameterByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("SearchParameter", c.Param("id")))
	}
	if err := h.svc.DeleteSearchParameter(c.Request().Context(), existing.ID); err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) PatchSearchParameterFHIR(c echo.Context) error {
	return h.handlePatch(c, c.Param("id"))
}

func (h *Handler) VreadSearchParameterFHIR(c echo.Context) error {
	sp, err := h.svc.GetSearchParameterByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("SearchParameter", c.Param("id")))
	}
	result := sp.ToFHIR()
	fhir.SetVersionHeaders(c, 1, sp.UpdatedAt.Format("2006-01-02T15:04:05Z"))
	return c.JSON(http.StatusOK, result)
}

func (h *Handler) HistorySearchParameterFHIR(c echo.Context) error {
	sp, err := h.svc.GetSearchParameterByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("SearchParameter", c.Param("id")))
	}
	result := sp.ToFHIR()
	raw, _ := json.Marshal(result)
	entry := &fhir.HistoryEntry{
		ResourceType: "SearchParameter", ResourceID: sp.FHIRID, VersionID: 1,
		Resource: raw, Action: "create", Timestamp: sp.CreatedAt,
	}
	return c.JSON(http.StatusOK, fhir.NewHistoryBundle([]*fhir.HistoryEntry{entry}, 1, "/fhir"))
}

func (h *Handler) handlePatch(c echo.Context, fhirID string) error {
	contentType := c.Request().Header.Get("Content-Type")
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome("failed to read request body"))
	}
	existing, err := h.svc.GetSearchParameterByFHIRID(c.Request().Context(), fhirID)
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("SearchParameter", fhirID))
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
	if err := h.svc.UpdateSearchParameter(c.Request().Context(), existing); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, existing.ToFHIR())
}
