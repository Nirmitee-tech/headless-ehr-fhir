package immunization

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

	// REST endpoints
	read := api.Group("", role)
	read.GET("/immunizations", h.ListImmunizations)
	read.GET("/immunizations/:id", h.GetImmunization)
	read.GET("/immunization-recommendations", h.ListRecommendations)
	read.GET("/immunization-recommendations/:id", h.GetRecommendation)

	write := api.Group("", role)
	write.POST("/immunizations", h.CreateImmunization)
	write.PUT("/immunizations/:id", h.UpdateImmunization)
	write.DELETE("/immunizations/:id", h.DeleteImmunization)
	write.POST("/immunization-recommendations", h.CreateRecommendation)
	write.PUT("/immunization-recommendations/:id", h.UpdateRecommendation)
	write.DELETE("/immunization-recommendations/:id", h.DeleteRecommendation)

	// FHIR read endpoints
	fhirRead := fhirGroup.Group("", role)
	fhirRead.GET("/Immunization", h.SearchImmunizationsFHIR)
	fhirRead.GET("/Immunization/:id", h.GetImmunizationFHIR)
	fhirRead.GET("/ImmunizationRecommendation", h.SearchRecommendationsFHIR)
	fhirRead.GET("/ImmunizationRecommendation/:id", h.GetRecommendationFHIR)

	// FHIR write endpoints
	fhirWrite := fhirGroup.Group("", role)
	fhirWrite.POST("/Immunization", h.CreateImmunizationFHIR)
	fhirWrite.PUT("/Immunization/:id", h.UpdateImmunizationFHIR)
	fhirWrite.DELETE("/Immunization/:id", h.DeleteImmunizationFHIR)
	fhirWrite.PATCH("/Immunization/:id", h.PatchImmunizationFHIR)
	fhirWrite.POST("/ImmunizationRecommendation", h.CreateRecommendationFHIR)
	fhirWrite.PUT("/ImmunizationRecommendation/:id", h.UpdateRecommendationFHIR)
	fhirWrite.DELETE("/ImmunizationRecommendation/:id", h.DeleteRecommendationFHIR)
	fhirWrite.PATCH("/ImmunizationRecommendation/:id", h.PatchRecommendationFHIR)

	// FHIR POST _search
	fhirRead.POST("/Immunization/_search", h.SearchImmunizationsFHIR)
	fhirRead.POST("/ImmunizationRecommendation/_search", h.SearchRecommendationsFHIR)

	// FHIR vread and history
	fhirRead.GET("/Immunization/:id/_history/:vid", h.VreadImmunizationFHIR)
	fhirRead.GET("/Immunization/:id/_history", h.HistoryImmunizationFHIR)
	fhirRead.GET("/ImmunizationRecommendation/:id/_history/:vid", h.VreadRecommendationFHIR)
	fhirRead.GET("/ImmunizationRecommendation/:id/_history", h.HistoryRecommendationFHIR)
}

// -- Immunization REST Handlers --

func (h *Handler) CreateImmunization(c echo.Context) error {
	var im Immunization
	if err := c.Bind(&im); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateImmunization(c.Request().Context(), &im); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, im)
}

func (h *Handler) GetImmunization(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	im, err := h.svc.GetImmunization(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "immunization not found")
	}
	return c.JSON(http.StatusOK, im)
}

func (h *Handler) ListImmunizations(c echo.Context) error {
	pg := pagination.FromContext(c)
	if patientID := c.QueryParam("patient_id"); patientID != "" {
		pid, err := uuid.Parse(patientID)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid patient_id")
		}
		items, total, err := h.svc.ListImmunizationsByPatient(c.Request().Context(), pid, pg.Limit, pg.Offset)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
	}
	items, total, err := h.svc.SearchImmunizations(c.Request().Context(), nil, pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *Handler) UpdateImmunization(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var im Immunization
	if err := c.Bind(&im); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	im.ID = id
	if err := h.svc.UpdateImmunization(c.Request().Context(), &im); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, im)
}

func (h *Handler) DeleteImmunization(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeleteImmunization(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

// -- Recommendation REST Handlers --

func (h *Handler) CreateRecommendation(c echo.Context) error {
	var r ImmunizationRecommendation
	if err := c.Bind(&r); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateRecommendation(c.Request().Context(), &r); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, r)
}

func (h *Handler) GetRecommendation(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	r, err := h.svc.GetRecommendation(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "recommendation not found")
	}
	return c.JSON(http.StatusOK, r)
}

func (h *Handler) ListRecommendations(c echo.Context) error {
	pg := pagination.FromContext(c)
	if patientID := c.QueryParam("patient_id"); patientID != "" {
		pid, err := uuid.Parse(patientID)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid patient_id")
		}
		items, total, err := h.svc.ListRecommendationsByPatient(c.Request().Context(), pid, pg.Limit, pg.Offset)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
	}
	items, total, err := h.svc.SearchRecommendations(c.Request().Context(), nil, pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *Handler) UpdateRecommendation(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var r ImmunizationRecommendation
	if err := c.Bind(&r); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	r.ID = id
	if err := h.svc.UpdateRecommendation(c.Request().Context(), &r); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, r)
}

func (h *Handler) DeleteRecommendation(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeleteRecommendation(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

// -- FHIR Immunization Endpoints --

func (h *Handler) SearchImmunizationsFHIR(c echo.Context) error {
	pg := pagination.FromContext(c)
	params := fhir.ExtractSearchParams(c)
	items, total, err := h.svc.SearchImmunizations(c.Request().Context(), params, pg.Limit, pg.Offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	resources := make([]interface{}, len(items))
	for i, item := range items {
		resources[i] = item.ToFHIR()
	}
	return c.JSON(http.StatusOK, fhir.NewSearchBundleWithLinks(resources, fhir.SearchBundleParams{
		BaseURL:  "/fhir/Immunization",
		QueryStr: c.QueryString(),
		Count:    pg.Limit,
		Offset:   pg.Offset,
		Total:    total,
	}))
}

func (h *Handler) GetImmunizationFHIR(c echo.Context) error {
	im, err := h.svc.GetImmunizationByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Immunization", c.Param("id")))
	}
	return c.JSON(http.StatusOK, im.ToFHIR())
}

func (h *Handler) CreateImmunizationFHIR(c echo.Context) error {
	var im Immunization
	if err := c.Bind(&im); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	if err := h.svc.CreateImmunization(c.Request().Context(), &im); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	c.Response().Header().Set("Location", "/fhir/Immunization/"+im.FHIRID)
	return c.JSON(http.StatusCreated, im.ToFHIR())
}

func (h *Handler) UpdateImmunizationFHIR(c echo.Context) error {
	var im Immunization
	if err := c.Bind(&im); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	existing, err := h.svc.GetImmunizationByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Immunization", c.Param("id")))
	}
	im.ID = existing.ID
	im.FHIRID = existing.FHIRID
	if err := h.svc.UpdateImmunization(c.Request().Context(), &im); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, im.ToFHIR())
}

func (h *Handler) DeleteImmunizationFHIR(c echo.Context) error {
	existing, err := h.svc.GetImmunizationByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Immunization", c.Param("id")))
	}
	if err := h.svc.DeleteImmunization(c.Request().Context(), existing.ID); err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) PatchImmunizationFHIR(c echo.Context) error {
	return h.handlePatch(c, "Immunization", c.Param("id"), func(ctx echo.Context, resource map[string]interface{}) error {
		existing, err := h.svc.GetImmunizationByFHIRID(ctx.Request().Context(), ctx.Param("id"))
		if err != nil {
			return ctx.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Immunization", ctx.Param("id")))
		}
		if v, ok := resource["status"].(string); ok {
			existing.Status = v
		}
		if err := h.svc.UpdateImmunization(ctx.Request().Context(), existing); err != nil {
			return ctx.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
		}
		return ctx.JSON(http.StatusOK, existing.ToFHIR())
	})
}

func (h *Handler) VreadImmunizationFHIR(c echo.Context) error {
	im, err := h.svc.GetImmunizationByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Immunization", c.Param("id")))
	}
	result := im.ToFHIR()
	fhir.SetVersionHeaders(c, 1, im.UpdatedAt.Format("2006-01-02T15:04:05Z"))
	return c.JSON(http.StatusOK, result)
}

func (h *Handler) HistoryImmunizationFHIR(c echo.Context) error {
	im, err := h.svc.GetImmunizationByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Immunization", c.Param("id")))
	}
	result := im.ToFHIR()
	raw, _ := json.Marshal(result)
	entry := &fhir.HistoryEntry{
		ResourceType: "Immunization", ResourceID: im.FHIRID, VersionID: 1,
		Resource: raw, Action: "create", Timestamp: im.CreatedAt,
	}
	return c.JSON(http.StatusOK, fhir.NewHistoryBundle([]*fhir.HistoryEntry{entry}, 1, "/fhir"))
}

// -- FHIR Recommendation Endpoints --

func (h *Handler) SearchRecommendationsFHIR(c echo.Context) error {
	pg := pagination.FromContext(c)
	params := fhir.ExtractSearchParams(c)
	items, total, err := h.svc.SearchRecommendations(c.Request().Context(), params, pg.Limit, pg.Offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	resources := make([]interface{}, len(items))
	for i, item := range items {
		resources[i] = item.ToFHIR()
	}
	return c.JSON(http.StatusOK, fhir.NewSearchBundleWithLinks(resources, fhir.SearchBundleParams{
		BaseURL:  "/fhir/ImmunizationRecommendation",
		QueryStr: c.QueryString(),
		Count:    pg.Limit,
		Offset:   pg.Offset,
		Total:    total,
	}))
}

func (h *Handler) GetRecommendationFHIR(c echo.Context) error {
	r, err := h.svc.GetRecommendationByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("ImmunizationRecommendation", c.Param("id")))
	}
	return c.JSON(http.StatusOK, r.ToFHIR())
}

func (h *Handler) CreateRecommendationFHIR(c echo.Context) error {
	var r ImmunizationRecommendation
	if err := c.Bind(&r); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	if err := h.svc.CreateRecommendation(c.Request().Context(), &r); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	c.Response().Header().Set("Location", "/fhir/ImmunizationRecommendation/"+r.FHIRID)
	return c.JSON(http.StatusCreated, r.ToFHIR())
}

func (h *Handler) UpdateRecommendationFHIR(c echo.Context) error {
	var r ImmunizationRecommendation
	if err := c.Bind(&r); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	existing, err := h.svc.GetRecommendationByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("ImmunizationRecommendation", c.Param("id")))
	}
	r.ID = existing.ID
	r.FHIRID = existing.FHIRID
	if err := h.svc.UpdateRecommendation(c.Request().Context(), &r); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, r.ToFHIR())
}

func (h *Handler) DeleteRecommendationFHIR(c echo.Context) error {
	existing, err := h.svc.GetRecommendationByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("ImmunizationRecommendation", c.Param("id")))
	}
	if err := h.svc.DeleteRecommendation(c.Request().Context(), existing.ID); err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) PatchRecommendationFHIR(c echo.Context) error {
	return h.handlePatch(c, "ImmunizationRecommendation", c.Param("id"), func(ctx echo.Context, resource map[string]interface{}) error {
		existing, err := h.svc.GetRecommendationByFHIRID(ctx.Request().Context(), ctx.Param("id"))
		if err != nil {
			return ctx.JSON(http.StatusNotFound, fhir.NotFoundOutcome("ImmunizationRecommendation", ctx.Param("id")))
		}
		if err := h.svc.UpdateRecommendation(ctx.Request().Context(), existing); err != nil {
			return ctx.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
		}
		return ctx.JSON(http.StatusOK, existing.ToFHIR())
	})
}

func (h *Handler) VreadRecommendationFHIR(c echo.Context) error {
	r, err := h.svc.GetRecommendationByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("ImmunizationRecommendation", c.Param("id")))
	}
	result := r.ToFHIR()
	fhir.SetVersionHeaders(c, 1, r.UpdatedAt.Format("2006-01-02T15:04:05Z"))
	return c.JSON(http.StatusOK, result)
}

func (h *Handler) HistoryRecommendationFHIR(c echo.Context) error {
	r, err := h.svc.GetRecommendationByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("ImmunizationRecommendation", c.Param("id")))
	}
	result := r.ToFHIR()
	raw, _ := json.Marshal(result)
	entry := &fhir.HistoryEntry{
		ResourceType: "ImmunizationRecommendation", ResourceID: r.FHIRID, VersionID: 1,
		Resource: raw, Action: "create", Timestamp: r.CreatedAt,
	}
	return c.JSON(http.StatusOK, fhir.NewHistoryBundle([]*fhir.HistoryEntry{entry}, 1, "/fhir"))
}

// handlePatch dispatches to JSON Patch or Merge Patch based on Content-Type.
func (h *Handler) handlePatch(c echo.Context, resourceType, fhirID string, applyFn func(echo.Context, map[string]interface{}) error) error {
	contentType := c.Request().Header.Get("Content-Type")
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome("failed to read request body"))
	}

	var currentResource map[string]interface{}
	switch resourceType {
	case "Immunization":
		existing, err := h.svc.GetImmunizationByFHIRID(c.Request().Context(), fhirID)
		if err != nil {
			return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome(resourceType, fhirID))
		}
		currentResource = existing.ToFHIR()
	case "ImmunizationRecommendation":
		existing, err := h.svc.GetRecommendationByFHIRID(c.Request().Context(), fhirID)
		if err != nil {
			return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome(resourceType, fhirID))
		}
		currentResource = existing.ToFHIR()
	default:
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome("unsupported resource type for PATCH"))
	}

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

	return applyFn(c, patched)
}
