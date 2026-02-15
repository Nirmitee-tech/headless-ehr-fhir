package clinical

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

// ClinicalSafetyHandler handles REST and FHIR endpoints for Flag, DetectedIssue,
// AdverseEvent, ClinicalImpression, and RiskAssessment resources.
type ClinicalSafetyHandler struct {
	svc *ClinicalSafetyService
}

func NewClinicalSafetyHandler(svc *ClinicalSafetyService) *ClinicalSafetyHandler {
	return &ClinicalSafetyHandler{svc: svc}
}

func (h *ClinicalSafetyHandler) RegisterRoutes(api *echo.Group, fhirGroup *echo.Group) {
	read := api.Group("", auth.RequireRole("admin", "physician", "nurse"))
	read.GET("/flags", h.ListFlags)
	read.GET("/flags/:id", h.GetFlag)
	read.GET("/detected-issues", h.ListDetectedIssues)
	read.GET("/detected-issues/:id", h.GetDetectedIssue)
	read.GET("/adverse-events", h.ListAdverseEvents)
	read.GET("/adverse-events/:id", h.GetAdverseEvent)
	read.GET("/clinical-impressions", h.ListClinicalImpressions)
	read.GET("/clinical-impressions/:id", h.GetClinicalImpression)
	read.GET("/risk-assessments", h.ListRiskAssessments)
	read.GET("/risk-assessments/:id", h.GetRiskAssessment)

	write := api.Group("", auth.RequireRole("admin", "physician", "nurse"))
	write.POST("/flags", h.CreateFlag)
	write.PUT("/flags/:id", h.UpdateFlag)
	write.DELETE("/flags/:id", h.DeleteFlag)
	write.POST("/detected-issues", h.CreateDetectedIssue)
	write.PUT("/detected-issues/:id", h.UpdateDetectedIssue)
	write.DELETE("/detected-issues/:id", h.DeleteDetectedIssue)
	write.POST("/adverse-events", h.CreateAdverseEvent)
	write.PUT("/adverse-events/:id", h.UpdateAdverseEvent)
	write.DELETE("/adverse-events/:id", h.DeleteAdverseEvent)
	write.POST("/clinical-impressions", h.CreateClinicalImpression)
	write.PUT("/clinical-impressions/:id", h.UpdateClinicalImpression)
	write.DELETE("/clinical-impressions/:id", h.DeleteClinicalImpression)
	write.POST("/risk-assessments", h.CreateRiskAssessment)
	write.PUT("/risk-assessments/:id", h.UpdateRiskAssessment)
	write.DELETE("/risk-assessments/:id", h.DeleteRiskAssessment)

	// FHIR read
	fr := fhirGroup.Group("", auth.RequireRole("admin", "physician", "nurse"))
	fr.GET("/Flag", h.SearchFlagsFHIR)
	fr.GET("/Flag/:id", h.GetFlagFHIR)
	fr.GET("/DetectedIssue", h.SearchDetectedIssuesFHIR)
	fr.GET("/DetectedIssue/:id", h.GetDetectedIssueFHIR)
	fr.GET("/AdverseEvent", h.SearchAdverseEventsFHIR)
	fr.GET("/AdverseEvent/:id", h.GetAdverseEventFHIR)
	fr.GET("/ClinicalImpression", h.SearchClinicalImpressionsFHIR)
	fr.GET("/ClinicalImpression/:id", h.GetClinicalImpressionFHIR)
	fr.GET("/RiskAssessment", h.SearchRiskAssessmentsFHIR)
	fr.GET("/RiskAssessment/:id", h.GetRiskAssessmentFHIR)

	// FHIR write
	fw := fhirGroup.Group("", auth.RequireRole("admin", "physician", "nurse"))
	fw.POST("/Flag", h.CreateFlagFHIR)
	fw.PUT("/Flag/:id", h.UpdateFlagFHIR)
	fw.DELETE("/Flag/:id", h.DeleteFlagFHIR)
	fw.PATCH("/Flag/:id", h.PatchFlagFHIR)
	fw.POST("/DetectedIssue", h.CreateDetectedIssueFHIR)
	fw.PUT("/DetectedIssue/:id", h.UpdateDetectedIssueFHIR)
	fw.DELETE("/DetectedIssue/:id", h.DeleteDetectedIssueFHIR)
	fw.PATCH("/DetectedIssue/:id", h.PatchDetectedIssueFHIR)
	fw.POST("/AdverseEvent", h.CreateAdverseEventFHIR)
	fw.PUT("/AdverseEvent/:id", h.UpdateAdverseEventFHIR)
	fw.DELETE("/AdverseEvent/:id", h.DeleteAdverseEventFHIR)
	fw.PATCH("/AdverseEvent/:id", h.PatchAdverseEventFHIR)
	fw.POST("/ClinicalImpression", h.CreateClinicalImpressionFHIR)
	fw.PUT("/ClinicalImpression/:id", h.UpdateClinicalImpressionFHIR)
	fw.DELETE("/ClinicalImpression/:id", h.DeleteClinicalImpressionFHIR)
	fw.PATCH("/ClinicalImpression/:id", h.PatchClinicalImpressionFHIR)
	fw.POST("/RiskAssessment", h.CreateRiskAssessmentFHIR)
	fw.PUT("/RiskAssessment/:id", h.UpdateRiskAssessmentFHIR)
	fw.DELETE("/RiskAssessment/:id", h.DeleteRiskAssessmentFHIR)
	fw.PATCH("/RiskAssessment/:id", h.PatchRiskAssessmentFHIR)

	// FHIR _search
	fr.POST("/Flag/_search", h.SearchFlagsFHIR)
	fr.POST("/DetectedIssue/_search", h.SearchDetectedIssuesFHIR)
	fr.POST("/AdverseEvent/_search", h.SearchAdverseEventsFHIR)
	fr.POST("/ClinicalImpression/_search", h.SearchClinicalImpressionsFHIR)
	fr.POST("/RiskAssessment/_search", h.SearchRiskAssessmentsFHIR)

	// FHIR vread and history
	fr.GET("/Flag/:id/_history/:vid", h.VreadFlagFHIR)
	fr.GET("/Flag/:id/_history", h.HistoryFlagFHIR)
	fr.GET("/DetectedIssue/:id/_history/:vid", h.VreadDetectedIssueFHIR)
	fr.GET("/DetectedIssue/:id/_history", h.HistoryDetectedIssueFHIR)
	fr.GET("/AdverseEvent/:id/_history/:vid", h.VreadAdverseEventFHIR)
	fr.GET("/AdverseEvent/:id/_history", h.HistoryAdverseEventFHIR)
	fr.GET("/ClinicalImpression/:id/_history/:vid", h.VreadClinicalImpressionFHIR)
	fr.GET("/ClinicalImpression/:id/_history", h.HistoryClinicalImpressionFHIR)
	fr.GET("/RiskAssessment/:id/_history/:vid", h.VreadRiskAssessmentFHIR)
	fr.GET("/RiskAssessment/:id/_history", h.HistoryRiskAssessmentFHIR)
}

// ============ REST Handlers ============

// -- Flag REST --

func (h *ClinicalSafetyHandler) CreateFlag(c echo.Context) error {
	var f Flag
	if err := c.Bind(&f); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateFlag(c.Request().Context(), &f); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, f)
}

func (h *ClinicalSafetyHandler) GetFlag(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	f, err := h.svc.GetFlag(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "flag not found")
	}
	return c.JSON(http.StatusOK, f)
}

func (h *ClinicalSafetyHandler) ListFlags(c echo.Context) error {
	pg := pagination.FromContext(c)
	params := map[string]string{}
	for _, k := range []string{"patient", "status"} {
		if v := c.QueryParam(k); v != "" {
			params[k] = v
		}
	}
	items, total, err := h.svc.SearchFlags(c.Request().Context(), params, pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *ClinicalSafetyHandler) UpdateFlag(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var f Flag
	if err := c.Bind(&f); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	f.ID = id
	if err := h.svc.UpdateFlag(c.Request().Context(), &f); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, f)
}

func (h *ClinicalSafetyHandler) DeleteFlag(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeleteFlag(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

// -- DetectedIssue REST --

func (h *ClinicalSafetyHandler) CreateDetectedIssue(c echo.Context) error {
	var d DetectedIssue
	if err := c.Bind(&d); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateDetectedIssue(c.Request().Context(), &d); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, d)
}

func (h *ClinicalSafetyHandler) GetDetectedIssue(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	d, err := h.svc.GetDetectedIssue(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "detected issue not found")
	}
	return c.JSON(http.StatusOK, d)
}

func (h *ClinicalSafetyHandler) ListDetectedIssues(c echo.Context) error {
	pg := pagination.FromContext(c)
	params := map[string]string{}
	for _, k := range []string{"patient", "status"} {
		if v := c.QueryParam(k); v != "" {
			params[k] = v
		}
	}
	items, total, err := h.svc.SearchDetectedIssues(c.Request().Context(), params, pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *ClinicalSafetyHandler) UpdateDetectedIssue(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var d DetectedIssue
	if err := c.Bind(&d); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	d.ID = id
	if err := h.svc.UpdateDetectedIssue(c.Request().Context(), &d); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, d)
}

func (h *ClinicalSafetyHandler) DeleteDetectedIssue(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeleteDetectedIssue(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

// -- AdverseEvent REST --

func (h *ClinicalSafetyHandler) CreateAdverseEvent(c echo.Context) error {
	var a AdverseEvent
	if err := c.Bind(&a); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateAdverseEvent(c.Request().Context(), &a); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, a)
}

func (h *ClinicalSafetyHandler) GetAdverseEvent(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	a, err := h.svc.GetAdverseEvent(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "adverse event not found")
	}
	return c.JSON(http.StatusOK, a)
}

func (h *ClinicalSafetyHandler) ListAdverseEvents(c echo.Context) error {
	pg := pagination.FromContext(c)
	params := map[string]string{}
	for _, k := range []string{"patient", "actuality"} {
		if v := c.QueryParam(k); v != "" {
			params[k] = v
		}
	}
	items, total, err := h.svc.SearchAdverseEvents(c.Request().Context(), params, pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *ClinicalSafetyHandler) UpdateAdverseEvent(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var a AdverseEvent
	if err := c.Bind(&a); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	a.ID = id
	if err := h.svc.UpdateAdverseEvent(c.Request().Context(), &a); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, a)
}

func (h *ClinicalSafetyHandler) DeleteAdverseEvent(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeleteAdverseEvent(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

// -- ClinicalImpression REST --

func (h *ClinicalSafetyHandler) CreateClinicalImpression(c echo.Context) error {
	var ci ClinicalImpression
	if err := c.Bind(&ci); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateClinicalImpression(c.Request().Context(), &ci); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, ci)
}

func (h *ClinicalSafetyHandler) GetClinicalImpression(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	ci, err := h.svc.GetClinicalImpression(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "clinical impression not found")
	}
	return c.JSON(http.StatusOK, ci)
}

func (h *ClinicalSafetyHandler) ListClinicalImpressions(c echo.Context) error {
	pg := pagination.FromContext(c)
	params := map[string]string{}
	for _, k := range []string{"patient", "status"} {
		if v := c.QueryParam(k); v != "" {
			params[k] = v
		}
	}
	items, total, err := h.svc.SearchClinicalImpressions(c.Request().Context(), params, pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *ClinicalSafetyHandler) UpdateClinicalImpression(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var ci ClinicalImpression
	if err := c.Bind(&ci); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	ci.ID = id
	if err := h.svc.UpdateClinicalImpression(c.Request().Context(), &ci); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, ci)
}

func (h *ClinicalSafetyHandler) DeleteClinicalImpression(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeleteClinicalImpression(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

// -- RiskAssessment REST --

func (h *ClinicalSafetyHandler) CreateRiskAssessment(c echo.Context) error {
	var ra RiskAssessment
	if err := c.Bind(&ra); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateRiskAssessment(c.Request().Context(), &ra); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, ra)
}

func (h *ClinicalSafetyHandler) GetRiskAssessment(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	ra, err := h.svc.GetRiskAssessment(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "risk assessment not found")
	}
	return c.JSON(http.StatusOK, ra)
}

func (h *ClinicalSafetyHandler) ListRiskAssessments(c echo.Context) error {
	pg := pagination.FromContext(c)
	params := map[string]string{}
	for _, k := range []string{"patient", "status"} {
		if v := c.QueryParam(k); v != "" {
			params[k] = v
		}
	}
	items, total, err := h.svc.SearchRiskAssessments(c.Request().Context(), params, pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *ClinicalSafetyHandler) UpdateRiskAssessment(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var ra RiskAssessment
	if err := c.Bind(&ra); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	ra.ID = id
	if err := h.svc.UpdateRiskAssessment(c.Request().Context(), &ra); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, ra)
}

func (h *ClinicalSafetyHandler) DeleteRiskAssessment(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeleteRiskAssessment(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

// ============ FHIR Handlers ============

// -- Flag FHIR --

func (h *ClinicalSafetyHandler) SearchFlagsFHIR(c echo.Context) error {
	pg := pagination.FromContext(c)
	params := map[string]string{}
	for _, k := range []string{"patient", "status"} {
		if v := c.QueryParam(k); v != "" {
			params[k] = v
		}
	}
	items, total, err := h.svc.SearchFlags(c.Request().Context(), params, pg.Limit, pg.Offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	resources := make([]interface{}, len(items))
	for i, item := range items {
		resources[i] = item.ToFHIR()
	}
	return c.JSON(http.StatusOK, fhir.NewSearchBundle(resources, total, "/fhir/Flag"))
}

func (h *ClinicalSafetyHandler) GetFlagFHIR(c echo.Context) error {
	f, err := h.svc.GetFlagByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Flag", c.Param("id")))
	}
	return c.JSON(http.StatusOK, f.ToFHIR())
}

func (h *ClinicalSafetyHandler) CreateFlagFHIR(c echo.Context) error {
	var f Flag
	if err := c.Bind(&f); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	if err := h.svc.CreateFlag(c.Request().Context(), &f); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	c.Response().Header().Set("Location", "/fhir/Flag/"+f.FHIRID)
	return c.JSON(http.StatusCreated, f.ToFHIR())
}

func (h *ClinicalSafetyHandler) UpdateFlagFHIR(c echo.Context) error {
	var f Flag
	if err := c.Bind(&f); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	existing, err := h.svc.GetFlagByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Flag", c.Param("id")))
	}
	f.ID = existing.ID
	f.FHIRID = existing.FHIRID
	if err := h.svc.UpdateFlag(c.Request().Context(), &f); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, f.ToFHIR())
}

func (h *ClinicalSafetyHandler) DeleteFlagFHIR(c echo.Context) error {
	existing, err := h.svc.GetFlagByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Flag", c.Param("id")))
	}
	if err := h.svc.DeleteFlag(c.Request().Context(), existing.ID); err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	return c.NoContent(http.StatusNoContent)
}

// -- DetectedIssue FHIR --

func (h *ClinicalSafetyHandler) SearchDetectedIssuesFHIR(c echo.Context) error {
	pg := pagination.FromContext(c)
	params := map[string]string{}
	for _, k := range []string{"patient", "status"} {
		if v := c.QueryParam(k); v != "" {
			params[k] = v
		}
	}
	items, total, err := h.svc.SearchDetectedIssues(c.Request().Context(), params, pg.Limit, pg.Offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	resources := make([]interface{}, len(items))
	for i, item := range items {
		resources[i] = item.ToFHIR()
	}
	return c.JSON(http.StatusOK, fhir.NewSearchBundle(resources, total, "/fhir/DetectedIssue"))
}

func (h *ClinicalSafetyHandler) GetDetectedIssueFHIR(c echo.Context) error {
	d, err := h.svc.GetDetectedIssueByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("DetectedIssue", c.Param("id")))
	}
	return c.JSON(http.StatusOK, d.ToFHIR())
}

func (h *ClinicalSafetyHandler) CreateDetectedIssueFHIR(c echo.Context) error {
	var d DetectedIssue
	if err := c.Bind(&d); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	if err := h.svc.CreateDetectedIssue(c.Request().Context(), &d); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	c.Response().Header().Set("Location", "/fhir/DetectedIssue/"+d.FHIRID)
	return c.JSON(http.StatusCreated, d.ToFHIR())
}

func (h *ClinicalSafetyHandler) UpdateDetectedIssueFHIR(c echo.Context) error {
	var d DetectedIssue
	if err := c.Bind(&d); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	existing, err := h.svc.GetDetectedIssueByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("DetectedIssue", c.Param("id")))
	}
	d.ID = existing.ID
	d.FHIRID = existing.FHIRID
	if err := h.svc.UpdateDetectedIssue(c.Request().Context(), &d); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, d.ToFHIR())
}

func (h *ClinicalSafetyHandler) DeleteDetectedIssueFHIR(c echo.Context) error {
	existing, err := h.svc.GetDetectedIssueByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("DetectedIssue", c.Param("id")))
	}
	if err := h.svc.DeleteDetectedIssue(c.Request().Context(), existing.ID); err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	return c.NoContent(http.StatusNoContent)
}

// -- AdverseEvent FHIR --

func (h *ClinicalSafetyHandler) SearchAdverseEventsFHIR(c echo.Context) error {
	pg := pagination.FromContext(c)
	params := map[string]string{}
	for _, k := range []string{"patient", "actuality"} {
		if v := c.QueryParam(k); v != "" {
			params[k] = v
		}
	}
	items, total, err := h.svc.SearchAdverseEvents(c.Request().Context(), params, pg.Limit, pg.Offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	resources := make([]interface{}, len(items))
	for i, item := range items {
		resources[i] = item.ToFHIR()
	}
	return c.JSON(http.StatusOK, fhir.NewSearchBundle(resources, total, "/fhir/AdverseEvent"))
}

func (h *ClinicalSafetyHandler) GetAdverseEventFHIR(c echo.Context) error {
	a, err := h.svc.GetAdverseEventByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("AdverseEvent", c.Param("id")))
	}
	return c.JSON(http.StatusOK, a.ToFHIR())
}

func (h *ClinicalSafetyHandler) CreateAdverseEventFHIR(c echo.Context) error {
	var a AdverseEvent
	if err := c.Bind(&a); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	if err := h.svc.CreateAdverseEvent(c.Request().Context(), &a); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	c.Response().Header().Set("Location", "/fhir/AdverseEvent/"+a.FHIRID)
	return c.JSON(http.StatusCreated, a.ToFHIR())
}

func (h *ClinicalSafetyHandler) UpdateAdverseEventFHIR(c echo.Context) error {
	var a AdverseEvent
	if err := c.Bind(&a); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	existing, err := h.svc.GetAdverseEventByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("AdverseEvent", c.Param("id")))
	}
	a.ID = existing.ID
	a.FHIRID = existing.FHIRID
	if err := h.svc.UpdateAdverseEvent(c.Request().Context(), &a); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, a.ToFHIR())
}

func (h *ClinicalSafetyHandler) DeleteAdverseEventFHIR(c echo.Context) error {
	existing, err := h.svc.GetAdverseEventByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("AdverseEvent", c.Param("id")))
	}
	if err := h.svc.DeleteAdverseEvent(c.Request().Context(), existing.ID); err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	return c.NoContent(http.StatusNoContent)
}

// -- ClinicalImpression FHIR --

func (h *ClinicalSafetyHandler) SearchClinicalImpressionsFHIR(c echo.Context) error {
	pg := pagination.FromContext(c)
	params := map[string]string{}
	for _, k := range []string{"patient", "status"} {
		if v := c.QueryParam(k); v != "" {
			params[k] = v
		}
	}
	items, total, err := h.svc.SearchClinicalImpressions(c.Request().Context(), params, pg.Limit, pg.Offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	resources := make([]interface{}, len(items))
	for i, item := range items {
		resources[i] = item.ToFHIR()
	}
	return c.JSON(http.StatusOK, fhir.NewSearchBundle(resources, total, "/fhir/ClinicalImpression"))
}

func (h *ClinicalSafetyHandler) GetClinicalImpressionFHIR(c echo.Context) error {
	ci, err := h.svc.GetClinicalImpressionByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("ClinicalImpression", c.Param("id")))
	}
	return c.JSON(http.StatusOK, ci.ToFHIR())
}

func (h *ClinicalSafetyHandler) CreateClinicalImpressionFHIR(c echo.Context) error {
	var ci ClinicalImpression
	if err := c.Bind(&ci); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	if err := h.svc.CreateClinicalImpression(c.Request().Context(), &ci); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	c.Response().Header().Set("Location", "/fhir/ClinicalImpression/"+ci.FHIRID)
	return c.JSON(http.StatusCreated, ci.ToFHIR())
}

func (h *ClinicalSafetyHandler) UpdateClinicalImpressionFHIR(c echo.Context) error {
	var ci ClinicalImpression
	if err := c.Bind(&ci); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	existing, err := h.svc.GetClinicalImpressionByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("ClinicalImpression", c.Param("id")))
	}
	ci.ID = existing.ID
	ci.FHIRID = existing.FHIRID
	if err := h.svc.UpdateClinicalImpression(c.Request().Context(), &ci); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, ci.ToFHIR())
}

func (h *ClinicalSafetyHandler) DeleteClinicalImpressionFHIR(c echo.Context) error {
	existing, err := h.svc.GetClinicalImpressionByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("ClinicalImpression", c.Param("id")))
	}
	if err := h.svc.DeleteClinicalImpression(c.Request().Context(), existing.ID); err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	return c.NoContent(http.StatusNoContent)
}

// -- RiskAssessment FHIR --

func (h *ClinicalSafetyHandler) SearchRiskAssessmentsFHIR(c echo.Context) error {
	pg := pagination.FromContext(c)
	params := map[string]string{}
	for _, k := range []string{"patient", "status"} {
		if v := c.QueryParam(k); v != "" {
			params[k] = v
		}
	}
	items, total, err := h.svc.SearchRiskAssessments(c.Request().Context(), params, pg.Limit, pg.Offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	resources := make([]interface{}, len(items))
	for i, item := range items {
		resources[i] = item.ToFHIR()
	}
	return c.JSON(http.StatusOK, fhir.NewSearchBundle(resources, total, "/fhir/RiskAssessment"))
}

func (h *ClinicalSafetyHandler) GetRiskAssessmentFHIR(c echo.Context) error {
	ra, err := h.svc.GetRiskAssessmentByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("RiskAssessment", c.Param("id")))
	}
	return c.JSON(http.StatusOK, ra.ToFHIR())
}

func (h *ClinicalSafetyHandler) CreateRiskAssessmentFHIR(c echo.Context) error {
	var ra RiskAssessment
	if err := c.Bind(&ra); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	if err := h.svc.CreateRiskAssessment(c.Request().Context(), &ra); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	c.Response().Header().Set("Location", "/fhir/RiskAssessment/"+ra.FHIRID)
	return c.JSON(http.StatusCreated, ra.ToFHIR())
}

func (h *ClinicalSafetyHandler) UpdateRiskAssessmentFHIR(c echo.Context) error {
	var ra RiskAssessment
	if err := c.Bind(&ra); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	existing, err := h.svc.GetRiskAssessmentByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("RiskAssessment", c.Param("id")))
	}
	ra.ID = existing.ID
	ra.FHIRID = existing.FHIRID
	if err := h.svc.UpdateRiskAssessment(c.Request().Context(), &ra); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, ra.ToFHIR())
}

func (h *ClinicalSafetyHandler) DeleteRiskAssessmentFHIR(c echo.Context) error {
	existing, err := h.svc.GetRiskAssessmentByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("RiskAssessment", c.Param("id")))
	}
	if err := h.svc.DeleteRiskAssessment(c.Request().Context(), existing.ID); err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	return c.NoContent(http.StatusNoContent)
}

// ============ FHIR PATCH ============

func (h *ClinicalSafetyHandler) PatchFlagFHIR(c echo.Context) error {
	return h.handleSafetyPatch(c, "Flag", c.Param("id"), func(ctx echo.Context, resource map[string]interface{}) error {
		existing, err := h.svc.GetFlagByFHIRID(ctx.Request().Context(), ctx.Param("id"))
		if err != nil {
			return ctx.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Flag", ctx.Param("id")))
		}
		if v, ok := resource["status"].(string); ok {
			existing.Status = v
		}
		if err := h.svc.UpdateFlag(ctx.Request().Context(), existing); err != nil {
			return ctx.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
		}
		return ctx.JSON(http.StatusOK, existing.ToFHIR())
	})
}

func (h *ClinicalSafetyHandler) PatchDetectedIssueFHIR(c echo.Context) error {
	return h.handleSafetyPatch(c, "DetectedIssue", c.Param("id"), func(ctx echo.Context, resource map[string]interface{}) error {
		existing, err := h.svc.GetDetectedIssueByFHIRID(ctx.Request().Context(), ctx.Param("id"))
		if err != nil {
			return ctx.JSON(http.StatusNotFound, fhir.NotFoundOutcome("DetectedIssue", ctx.Param("id")))
		}
		if v, ok := resource["status"].(string); ok {
			existing.Status = v
		}
		if err := h.svc.UpdateDetectedIssue(ctx.Request().Context(), existing); err != nil {
			return ctx.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
		}
		return ctx.JSON(http.StatusOK, existing.ToFHIR())
	})
}

func (h *ClinicalSafetyHandler) PatchAdverseEventFHIR(c echo.Context) error {
	return h.handleSafetyPatch(c, "AdverseEvent", c.Param("id"), func(ctx echo.Context, resource map[string]interface{}) error {
		existing, err := h.svc.GetAdverseEventByFHIRID(ctx.Request().Context(), ctx.Param("id"))
		if err != nil {
			return ctx.JSON(http.StatusNotFound, fhir.NotFoundOutcome("AdverseEvent", ctx.Param("id")))
		}
		if v, ok := resource["actuality"].(string); ok {
			existing.Actuality = v
		}
		if err := h.svc.UpdateAdverseEvent(ctx.Request().Context(), existing); err != nil {
			return ctx.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
		}
		return ctx.JSON(http.StatusOK, existing.ToFHIR())
	})
}

func (h *ClinicalSafetyHandler) PatchClinicalImpressionFHIR(c echo.Context) error {
	return h.handleSafetyPatch(c, "ClinicalImpression", c.Param("id"), func(ctx echo.Context, resource map[string]interface{}) error {
		existing, err := h.svc.GetClinicalImpressionByFHIRID(ctx.Request().Context(), ctx.Param("id"))
		if err != nil {
			return ctx.JSON(http.StatusNotFound, fhir.NotFoundOutcome("ClinicalImpression", ctx.Param("id")))
		}
		if v, ok := resource["status"].(string); ok {
			existing.Status = v
		}
		if err := h.svc.UpdateClinicalImpression(ctx.Request().Context(), existing); err != nil {
			return ctx.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
		}
		return ctx.JSON(http.StatusOK, existing.ToFHIR())
	})
}

func (h *ClinicalSafetyHandler) PatchRiskAssessmentFHIR(c echo.Context) error {
	return h.handleSafetyPatch(c, "RiskAssessment", c.Param("id"), func(ctx echo.Context, resource map[string]interface{}) error {
		existing, err := h.svc.GetRiskAssessmentByFHIRID(ctx.Request().Context(), ctx.Param("id"))
		if err != nil {
			return ctx.JSON(http.StatusNotFound, fhir.NotFoundOutcome("RiskAssessment", ctx.Param("id")))
		}
		if v, ok := resource["status"].(string); ok {
			existing.Status = v
		}
		if err := h.svc.UpdateRiskAssessment(ctx.Request().Context(), existing); err != nil {
			return ctx.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
		}
		return ctx.JSON(http.StatusOK, existing.ToFHIR())
	})
}

func (h *ClinicalSafetyHandler) handleSafetyPatch(c echo.Context, resourceType, fhirID string, applyFn func(echo.Context, map[string]interface{}) error) error {
	contentType := c.Request().Header.Get("Content-Type")
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome("failed to read request body"))
	}

	var currentResource map[string]interface{}
	switch resourceType {
	case "Flag":
		existing, err := h.svc.GetFlagByFHIRID(c.Request().Context(), fhirID)
		if err != nil {
			return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome(resourceType, fhirID))
		}
		currentResource = existing.ToFHIR()
	case "DetectedIssue":
		existing, err := h.svc.GetDetectedIssueByFHIRID(c.Request().Context(), fhirID)
		if err != nil {
			return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome(resourceType, fhirID))
		}
		currentResource = existing.ToFHIR()
	case "AdverseEvent":
		existing, err := h.svc.GetAdverseEventByFHIRID(c.Request().Context(), fhirID)
		if err != nil {
			return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome(resourceType, fhirID))
		}
		currentResource = existing.ToFHIR()
	case "ClinicalImpression":
		existing, err := h.svc.GetClinicalImpressionByFHIRID(c.Request().Context(), fhirID)
		if err != nil {
			return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome(resourceType, fhirID))
		}
		currentResource = existing.ToFHIR()
	case "RiskAssessment":
		existing, err := h.svc.GetRiskAssessmentByFHIRID(c.Request().Context(), fhirID)
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

// ============ FHIR Vread & History ============

func (h *ClinicalSafetyHandler) VreadFlagFHIR(c echo.Context) error {
	f, err := h.svc.GetFlagByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Flag", c.Param("id")))
	}
	fhir.SetVersionHeaders(c, 1, f.UpdatedAt.Format("2006-01-02T15:04:05Z"))
	return c.JSON(http.StatusOK, f.ToFHIR())
}

func (h *ClinicalSafetyHandler) HistoryFlagFHIR(c echo.Context) error {
	f, err := h.svc.GetFlagByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Flag", c.Param("id")))
	}
	raw, _ := json.Marshal(f.ToFHIR())
	entry := &fhir.HistoryEntry{ResourceType: "Flag", ResourceID: f.FHIRID, VersionID: 1, Resource: raw, Action: "create", Timestamp: f.CreatedAt}
	return c.JSON(http.StatusOK, fhir.NewHistoryBundle([]*fhir.HistoryEntry{entry}, 1, "/fhir"))
}

func (h *ClinicalSafetyHandler) VreadDetectedIssueFHIR(c echo.Context) error {
	d, err := h.svc.GetDetectedIssueByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("DetectedIssue", c.Param("id")))
	}
	fhir.SetVersionHeaders(c, 1, d.UpdatedAt.Format("2006-01-02T15:04:05Z"))
	return c.JSON(http.StatusOK, d.ToFHIR())
}

func (h *ClinicalSafetyHandler) HistoryDetectedIssueFHIR(c echo.Context) error {
	d, err := h.svc.GetDetectedIssueByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("DetectedIssue", c.Param("id")))
	}
	raw, _ := json.Marshal(d.ToFHIR())
	entry := &fhir.HistoryEntry{ResourceType: "DetectedIssue", ResourceID: d.FHIRID, VersionID: 1, Resource: raw, Action: "create", Timestamp: d.CreatedAt}
	return c.JSON(http.StatusOK, fhir.NewHistoryBundle([]*fhir.HistoryEntry{entry}, 1, "/fhir"))
}

func (h *ClinicalSafetyHandler) VreadAdverseEventFHIR(c echo.Context) error {
	a, err := h.svc.GetAdverseEventByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("AdverseEvent", c.Param("id")))
	}
	fhir.SetVersionHeaders(c, 1, a.UpdatedAt.Format("2006-01-02T15:04:05Z"))
	return c.JSON(http.StatusOK, a.ToFHIR())
}

func (h *ClinicalSafetyHandler) HistoryAdverseEventFHIR(c echo.Context) error {
	a, err := h.svc.GetAdverseEventByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("AdverseEvent", c.Param("id")))
	}
	raw, _ := json.Marshal(a.ToFHIR())
	entry := &fhir.HistoryEntry{ResourceType: "AdverseEvent", ResourceID: a.FHIRID, VersionID: 1, Resource: raw, Action: "create", Timestamp: a.CreatedAt}
	return c.JSON(http.StatusOK, fhir.NewHistoryBundle([]*fhir.HistoryEntry{entry}, 1, "/fhir"))
}

func (h *ClinicalSafetyHandler) VreadClinicalImpressionFHIR(c echo.Context) error {
	ci, err := h.svc.GetClinicalImpressionByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("ClinicalImpression", c.Param("id")))
	}
	fhir.SetVersionHeaders(c, 1, ci.UpdatedAt.Format("2006-01-02T15:04:05Z"))
	return c.JSON(http.StatusOK, ci.ToFHIR())
}

func (h *ClinicalSafetyHandler) HistoryClinicalImpressionFHIR(c echo.Context) error {
	ci, err := h.svc.GetClinicalImpressionByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("ClinicalImpression", c.Param("id")))
	}
	raw, _ := json.Marshal(ci.ToFHIR())
	entry := &fhir.HistoryEntry{ResourceType: "ClinicalImpression", ResourceID: ci.FHIRID, VersionID: 1, Resource: raw, Action: "create", Timestamp: ci.CreatedAt}
	return c.JSON(http.StatusOK, fhir.NewHistoryBundle([]*fhir.HistoryEntry{entry}, 1, "/fhir"))
}

func (h *ClinicalSafetyHandler) VreadRiskAssessmentFHIR(c echo.Context) error {
	ra, err := h.svc.GetRiskAssessmentByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("RiskAssessment", c.Param("id")))
	}
	fhir.SetVersionHeaders(c, 1, ra.UpdatedAt.Format("2006-01-02T15:04:05Z"))
	return c.JSON(http.StatusOK, ra.ToFHIR())
}

func (h *ClinicalSafetyHandler) HistoryRiskAssessmentFHIR(c echo.Context) error {
	ra, err := h.svc.GetRiskAssessmentByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("RiskAssessment", c.Param("id")))
	}
	raw, _ := json.Marshal(ra.ToFHIR())
	entry := &fhir.HistoryEntry{ResourceType: "RiskAssessment", ResourceID: ra.FHIRID, VersionID: 1, Resource: raw, Action: "create", Timestamp: ra.CreatedAt}
	return c.JSON(http.StatusOK, fhir.NewHistoryBundle([]*fhir.HistoryEntry{entry}, 1, "/fhir"))
}
