package substancespecification

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
	read.GET("/substance-specifications", h.ListSubstanceSpecifications)
	read.GET("/substance-specifications/:id", h.GetSubstanceSpecification)

	write := api.Group("", role)
	write.POST("/substance-specifications", h.CreateSubstanceSpecification)
	write.PUT("/substance-specifications/:id", h.UpdateSubstanceSpecification)
	write.DELETE("/substance-specifications/:id", h.DeleteSubstanceSpecification)

	fhirRead := fhirGroup.Group("", role)
	fhirRead.GET("/SubstanceSpecification", h.SearchSubstanceSpecificationsFHIR)
	fhirRead.GET("/SubstanceSpecification/:id", h.GetSubstanceSpecificationFHIR)

	fhirWrite := fhirGroup.Group("", role)
	fhirWrite.POST("/SubstanceSpecification", h.CreateSubstanceSpecificationFHIR)
	fhirWrite.PUT("/SubstanceSpecification/:id", h.UpdateSubstanceSpecificationFHIR)
	fhirWrite.DELETE("/SubstanceSpecification/:id", h.DeleteSubstanceSpecificationFHIR)
	fhirWrite.PATCH("/SubstanceSpecification/:id", h.PatchSubstanceSpecificationFHIR)

	fhirRead.POST("/SubstanceSpecification/_search", h.SearchSubstanceSpecificationsFHIR)
	fhirRead.GET("/SubstanceSpecification/:id/_history/:vid", h.VreadSubstanceSpecificationFHIR)
	fhirRead.GET("/SubstanceSpecification/:id/_history", h.HistorySubstanceSpecificationFHIR)
}

// -- REST Endpoints --

func (h *Handler) CreateSubstanceSpecification(c echo.Context) error {
	var s SubstanceSpecification
	if err := c.Bind(&s); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateSubstanceSpecification(c.Request().Context(), &s); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, s)
}

func (h *Handler) GetSubstanceSpecification(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	s, err := h.svc.GetSubstanceSpecification(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "substance specification not found")
	}
	return c.JSON(http.StatusOK, s)
}

func (h *Handler) ListSubstanceSpecifications(c echo.Context) error {
	pg := pagination.FromContext(c)
	items, total, err := h.svc.SearchSubstanceSpecifications(c.Request().Context(), nil, pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *Handler) UpdateSubstanceSpecification(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var s SubstanceSpecification
	if err := c.Bind(&s); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	s.ID = id
	if err := h.svc.UpdateSubstanceSpecification(c.Request().Context(), &s); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, s)
}

func (h *Handler) DeleteSubstanceSpecification(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeleteSubstanceSpecification(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

// -- FHIR Endpoints --

func (h *Handler) SearchSubstanceSpecificationsFHIR(c echo.Context) error {
	pg := pagination.FromContext(c)
	params := fhir.ExtractSearchParams(c)
	items, total, err := h.svc.SearchSubstanceSpecifications(c.Request().Context(), params, pg.Limit, pg.Offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	resources := make([]interface{}, len(items))
	for i, item := range items {
		resources[i] = item.ToFHIR()
	}
	return c.JSON(http.StatusOK, fhir.NewSearchBundle(resources, total, "/fhir/SubstanceSpecification"))
}

func (h *Handler) GetSubstanceSpecificationFHIR(c echo.Context) error {
	s, err := h.svc.GetSubstanceSpecificationByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("SubstanceSpecification", c.Param("id")))
	}
	fhir.SetVersionHeaders(c, 1, s.UpdatedAt.Format("2006-01-02T15:04:05Z"))
	return c.JSON(http.StatusOK, s.ToFHIR())
}

func (h *Handler) CreateSubstanceSpecificationFHIR(c echo.Context) error {
	var s SubstanceSpecification
	if err := c.Bind(&s); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	if err := h.svc.CreateSubstanceSpecification(c.Request().Context(), &s); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	c.Response().Header().Set("Location", "/fhir/SubstanceSpecification/"+s.FHIRID)
	return c.JSON(http.StatusCreated, s.ToFHIR())
}

func (h *Handler) UpdateSubstanceSpecificationFHIR(c echo.Context) error {
	var s SubstanceSpecification
	if err := c.Bind(&s); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	existing, err := h.svc.GetSubstanceSpecificationByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("SubstanceSpecification", c.Param("id")))
	}
	s.ID = existing.ID
	s.FHIRID = existing.FHIRID
	if err := h.svc.UpdateSubstanceSpecification(c.Request().Context(), &s); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, s.ToFHIR())
}

func (h *Handler) DeleteSubstanceSpecificationFHIR(c echo.Context) error {
	existing, err := h.svc.GetSubstanceSpecificationByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("SubstanceSpecification", c.Param("id")))
	}
	if err := h.svc.DeleteSubstanceSpecification(c.Request().Context(), existing.ID); err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) PatchSubstanceSpecificationFHIR(c echo.Context) error {
	return h.handlePatch(c, c.Param("id"))
}

func (h *Handler) VreadSubstanceSpecificationFHIR(c echo.Context) error {
	s, err := h.svc.GetSubstanceSpecificationByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("SubstanceSpecification", c.Param("id")))
	}
	result := s.ToFHIR()
	fhir.SetVersionHeaders(c, 1, s.UpdatedAt.Format("2006-01-02T15:04:05Z"))
	return c.JSON(http.StatusOK, result)
}

func (h *Handler) HistorySubstanceSpecificationFHIR(c echo.Context) error {
	s, err := h.svc.GetSubstanceSpecificationByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("SubstanceSpecification", c.Param("id")))
	}
	result := s.ToFHIR()
	raw, _ := json.Marshal(result)
	entry := &fhir.HistoryEntry{
		ResourceType: "SubstanceSpecification", ResourceID: s.FHIRID, VersionID: 1,
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
	existing, err := h.svc.GetSubstanceSpecificationByFHIRID(c.Request().Context(), fhirID)
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("SubstanceSpecification", fhirID))
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
	if v, ok := patched["description"].(string); ok {
		existing.Description = &v
	}
	if v, ok := patched["comment"].(string); ok {
		existing.Comment = &v
	}
	if err := h.svc.UpdateSubstanceSpecification(c.Request().Context(), existing); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, existing.ToFHIR())
}
