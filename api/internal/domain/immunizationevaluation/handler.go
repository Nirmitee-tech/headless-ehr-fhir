package immunizationevaluation

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
	read.GET("/immunization-evaluations", h.ListImmunizationEvaluations)
	read.GET("/immunization-evaluations/:id", h.GetImmunizationEvaluation)

	write := api.Group("", role)
	write.POST("/immunization-evaluations", h.CreateImmunizationEvaluation)
	write.PUT("/immunization-evaluations/:id", h.UpdateImmunizationEvaluation)
	write.DELETE("/immunization-evaluations/:id", h.DeleteImmunizationEvaluation)

	fhirRead := fhirGroup.Group("", role)
	fhirRead.GET("/ImmunizationEvaluation", h.SearchImmunizationEvaluationsFHIR)
	fhirRead.GET("/ImmunizationEvaluation/:id", h.GetImmunizationEvaluationFHIR)

	fhirWrite := fhirGroup.Group("", role)
	fhirWrite.POST("/ImmunizationEvaluation", h.CreateImmunizationEvaluationFHIR)
	fhirWrite.PUT("/ImmunizationEvaluation/:id", h.UpdateImmunizationEvaluationFHIR)
	fhirWrite.DELETE("/ImmunizationEvaluation/:id", h.DeleteImmunizationEvaluationFHIR)
	fhirWrite.PATCH("/ImmunizationEvaluation/:id", h.PatchImmunizationEvaluationFHIR)

	fhirRead.POST("/ImmunizationEvaluation/_search", h.SearchImmunizationEvaluationsFHIR)
	fhirRead.GET("/ImmunizationEvaluation/:id/_history/:vid", h.VreadImmunizationEvaluationFHIR)
	fhirRead.GET("/ImmunizationEvaluation/:id/_history", h.HistoryImmunizationEvaluationFHIR)
}

// -- REST --

func (h *Handler) CreateImmunizationEvaluation(c echo.Context) error {
	var ie ImmunizationEvaluation
	if err := c.Bind(&ie); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateImmunizationEvaluation(c.Request().Context(), &ie); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, ie)
}

func (h *Handler) GetImmunizationEvaluation(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	ie, err := h.svc.GetImmunizationEvaluation(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "immunization evaluation not found")
	}
	return c.JSON(http.StatusOK, ie)
}

func (h *Handler) ListImmunizationEvaluations(c echo.Context) error {
	pg := pagination.FromContext(c)
	items, total, err := h.svc.SearchImmunizationEvaluations(c.Request().Context(), nil, pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *Handler) UpdateImmunizationEvaluation(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var ie ImmunizationEvaluation
	if err := c.Bind(&ie); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	ie.ID = id
	if err := h.svc.UpdateImmunizationEvaluation(c.Request().Context(), &ie); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, ie)
}

func (h *Handler) DeleteImmunizationEvaluation(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeleteImmunizationEvaluation(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

// -- FHIR --

func (h *Handler) SearchImmunizationEvaluationsFHIR(c echo.Context) error {
	pg := pagination.FromContext(c)
	params := fhir.ExtractSearchParams(c)
	items, total, err := h.svc.SearchImmunizationEvaluations(c.Request().Context(), params, pg.Limit, pg.Offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	resources := make([]interface{}, len(items))
	for i, item := range items {
		resources[i] = item.ToFHIR()
	}
	return c.JSON(http.StatusOK, fhir.NewSearchBundle(resources, total, "/fhir/ImmunizationEvaluation"))
}

func (h *Handler) GetImmunizationEvaluationFHIR(c echo.Context) error {
	ie, err := h.svc.GetImmunizationEvaluationByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("ImmunizationEvaluation", c.Param("id")))
	}
	return c.JSON(http.StatusOK, ie.ToFHIR())
}

func (h *Handler) CreateImmunizationEvaluationFHIR(c echo.Context) error {
	var ie ImmunizationEvaluation
	if err := c.Bind(&ie); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	if err := h.svc.CreateImmunizationEvaluation(c.Request().Context(), &ie); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	c.Response().Header().Set("Location", "/fhir/ImmunizationEvaluation/"+ie.FHIRID)
	return c.JSON(http.StatusCreated, ie.ToFHIR())
}

func (h *Handler) UpdateImmunizationEvaluationFHIR(c echo.Context) error {
	var ie ImmunizationEvaluation
	if err := c.Bind(&ie); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	existing, err := h.svc.GetImmunizationEvaluationByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("ImmunizationEvaluation", c.Param("id")))
	}
	ie.ID = existing.ID
	ie.FHIRID = existing.FHIRID
	if err := h.svc.UpdateImmunizationEvaluation(c.Request().Context(), &ie); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, ie.ToFHIR())
}

func (h *Handler) DeleteImmunizationEvaluationFHIR(c echo.Context) error {
	existing, err := h.svc.GetImmunizationEvaluationByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("ImmunizationEvaluation", c.Param("id")))
	}
	if err := h.svc.DeleteImmunizationEvaluation(c.Request().Context(), existing.ID); err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) PatchImmunizationEvaluationFHIR(c echo.Context) error {
	return h.handlePatch(c, c.Param("id"))
}

func (h *Handler) VreadImmunizationEvaluationFHIR(c echo.Context) error {
	ie, err := h.svc.GetImmunizationEvaluationByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("ImmunizationEvaluation", c.Param("id")))
	}
	result := ie.ToFHIR()
	fhir.SetVersionHeaders(c, 1, ie.UpdatedAt.Format("2006-01-02T15:04:05Z"))
	return c.JSON(http.StatusOK, result)
}

func (h *Handler) HistoryImmunizationEvaluationFHIR(c echo.Context) error {
	ie, err := h.svc.GetImmunizationEvaluationByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("ImmunizationEvaluation", c.Param("id")))
	}
	result := ie.ToFHIR()
	raw, _ := json.Marshal(result)
	entry := &fhir.HistoryEntry{
		ResourceType: "ImmunizationEvaluation", ResourceID: ie.FHIRID, VersionID: 1,
		Resource: raw, Action: "create", Timestamp: ie.CreatedAt,
	}
	return c.JSON(http.StatusOK, fhir.NewHistoryBundle([]*fhir.HistoryEntry{entry}, 1, "/fhir"))
}

func (h *Handler) handlePatch(c echo.Context, fhirID string) error {
	contentType := c.Request().Header.Get("Content-Type")
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome("failed to read request body"))
	}
	existing, err := h.svc.GetImmunizationEvaluationByFHIRID(c.Request().Context(), fhirID)
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("ImmunizationEvaluation", fhirID))
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
	if err := h.svc.UpdateImmunizationEvaluation(c.Request().Context(), existing); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, existing.ToFHIR())
}
