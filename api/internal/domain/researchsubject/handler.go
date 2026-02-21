package researchsubject

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
	read.GET("/research-subjects", h.ListResearchSubjects)
	read.GET("/research-subjects/:id", h.GetResearchSubject)

	write := api.Group("", role)
	write.POST("/research-subjects", h.CreateResearchSubject)
	write.PUT("/research-subjects/:id", h.UpdateResearchSubject)
	write.DELETE("/research-subjects/:id", h.DeleteResearchSubject)

	fhirRead := fhirGroup.Group("", role)
	fhirRead.GET("/ResearchSubject", h.SearchResearchSubjectsFHIR)
	fhirRead.GET("/ResearchSubject/:id", h.GetResearchSubjectFHIR)

	fhirWrite := fhirGroup.Group("", role)
	fhirWrite.POST("/ResearchSubject", h.CreateResearchSubjectFHIR)
	fhirWrite.PUT("/ResearchSubject/:id", h.UpdateResearchSubjectFHIR)
	fhirWrite.DELETE("/ResearchSubject/:id", h.DeleteResearchSubjectFHIR)
	fhirWrite.PATCH("/ResearchSubject/:id", h.PatchResearchSubjectFHIR)

	fhirRead.POST("/ResearchSubject/_search", h.SearchResearchSubjectsFHIR)
	fhirRead.GET("/ResearchSubject/:id/_history/:vid", h.VreadResearchSubjectFHIR)
	fhirRead.GET("/ResearchSubject/:id/_history", h.HistoryResearchSubjectFHIR)
}

// -- REST Endpoints --

func (h *Handler) CreateResearchSubject(c echo.Context) error {
	var r ResearchSubject
	if err := c.Bind(&r); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateResearchSubject(c.Request().Context(), &r); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, r)
}

func (h *Handler) GetResearchSubject(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	r, err := h.svc.GetResearchSubject(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "research subject not found")
	}
	return c.JSON(http.StatusOK, r)
}

func (h *Handler) ListResearchSubjects(c echo.Context) error {
	pg := pagination.FromContext(c)
	items, total, err := h.svc.SearchResearchSubjects(c.Request().Context(), nil, pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *Handler) UpdateResearchSubject(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var r ResearchSubject
	if err := c.Bind(&r); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	r.ID = id
	if err := h.svc.UpdateResearchSubject(c.Request().Context(), &r); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, r)
}

func (h *Handler) DeleteResearchSubject(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeleteResearchSubject(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

// -- FHIR Endpoints --

func (h *Handler) SearchResearchSubjectsFHIR(c echo.Context) error {
	pg := pagination.FromContext(c)
	params := fhir.ExtractSearchParams(c)
	items, total, err := h.svc.SearchResearchSubjects(c.Request().Context(), params, pg.Limit, pg.Offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	resources := make([]interface{}, len(items))
	for i, item := range items {
		resources[i] = item.ToFHIR()
	}
	return c.JSON(http.StatusOK, fhir.NewSearchBundle(resources, total, "/fhir/ResearchSubject"))
}

func (h *Handler) GetResearchSubjectFHIR(c echo.Context) error {
	r, err := h.svc.GetResearchSubjectByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("ResearchSubject", c.Param("id")))
	}
	fhir.SetVersionHeaders(c, 1, r.UpdatedAt.Format("2006-01-02T15:04:05Z"))
	return c.JSON(http.StatusOK, r.ToFHIR())
}

func (h *Handler) CreateResearchSubjectFHIR(c echo.Context) error {
	var r ResearchSubject
	if err := c.Bind(&r); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	if err := h.svc.CreateResearchSubject(c.Request().Context(), &r); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	c.Response().Header().Set("Location", "/fhir/ResearchSubject/"+r.FHIRID)
	return c.JSON(http.StatusCreated, r.ToFHIR())
}

func (h *Handler) UpdateResearchSubjectFHIR(c echo.Context) error {
	var r ResearchSubject
	if err := c.Bind(&r); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	existing, err := h.svc.GetResearchSubjectByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("ResearchSubject", c.Param("id")))
	}
	r.ID = existing.ID
	r.FHIRID = existing.FHIRID
	if err := h.svc.UpdateResearchSubject(c.Request().Context(), &r); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, r.ToFHIR())
}

func (h *Handler) DeleteResearchSubjectFHIR(c echo.Context) error {
	existing, err := h.svc.GetResearchSubjectByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("ResearchSubject", c.Param("id")))
	}
	if err := h.svc.DeleteResearchSubject(c.Request().Context(), existing.ID); err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) PatchResearchSubjectFHIR(c echo.Context) error {
	return h.handlePatch(c, c.Param("id"))
}

func (h *Handler) VreadResearchSubjectFHIR(c echo.Context) error {
	r, err := h.svc.GetResearchSubjectByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("ResearchSubject", c.Param("id")))
	}
	result := r.ToFHIR()
	fhir.SetVersionHeaders(c, 1, r.UpdatedAt.Format("2006-01-02T15:04:05Z"))
	return c.JSON(http.StatusOK, result)
}

func (h *Handler) HistoryResearchSubjectFHIR(c echo.Context) error {
	r, err := h.svc.GetResearchSubjectByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("ResearchSubject", c.Param("id")))
	}
	result := r.ToFHIR()
	raw, _ := json.Marshal(result)
	entry := &fhir.HistoryEntry{
		ResourceType: "ResearchSubject", ResourceID: r.FHIRID, VersionID: 1,
		Resource: raw, Action: "create", Timestamp: r.CreatedAt,
	}
	return c.JSON(http.StatusOK, fhir.NewHistoryBundle([]*fhir.HistoryEntry{entry}, 1, "/fhir"))
}

func (h *Handler) handlePatch(c echo.Context, fhirID string) error {
	contentType := c.Request().Header.Get("Content-Type")
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome("failed to read request body"))
	}
	existing, err := h.svc.GetResearchSubjectByFHIRID(c.Request().Context(), fhirID)
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("ResearchSubject", fhirID))
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
	if err := h.svc.UpdateResearchSubject(c.Request().Context(), existing); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, existing.ToFHIR())
}
