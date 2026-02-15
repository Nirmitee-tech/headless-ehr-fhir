package measurereport

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
	role := auth.RequireRole("admin", "physician")

	read := api.Group("", role)
	read.GET("/measure-reports", h.ListMeasureReports)
	read.GET("/measure-reports/:id", h.GetMeasureReport)

	write := api.Group("", role)
	write.POST("/measure-reports", h.CreateMeasureReport)
	write.PUT("/measure-reports/:id", h.UpdateMeasureReport)
	write.DELETE("/measure-reports/:id", h.DeleteMeasureReport)

	fhirRead := fhirGroup.Group("", role)
	fhirRead.GET("/MeasureReport", h.SearchMeasureReportsFHIR)
	fhirRead.GET("/MeasureReport/:id", h.GetMeasureReportFHIR)

	fhirWrite := fhirGroup.Group("", role)
	fhirWrite.POST("/MeasureReport", h.CreateMeasureReportFHIR)
	fhirWrite.PUT("/MeasureReport/:id", h.UpdateMeasureReportFHIR)
	fhirWrite.DELETE("/MeasureReport/:id", h.DeleteMeasureReportFHIR)
	fhirWrite.PATCH("/MeasureReport/:id", h.PatchMeasureReportFHIR)

	fhirRead.POST("/MeasureReport/_search", h.SearchMeasureReportsFHIR)

	fhirRead.GET("/MeasureReport/:id/_history/:vid", h.VreadMeasureReportFHIR)
	fhirRead.GET("/MeasureReport/:id/_history", h.HistoryMeasureReportFHIR)
}

// -- REST Endpoints --

func (h *Handler) CreateMeasureReport(c echo.Context) error {
	var mr MeasureReport
	if err := c.Bind(&mr); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateMeasureReport(c.Request().Context(), &mr); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, mr)
}

func (h *Handler) GetMeasureReport(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	mr, err := h.svc.GetMeasureReport(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "measure report not found")
	}
	return c.JSON(http.StatusOK, mr)
}

func (h *Handler) ListMeasureReports(c echo.Context) error {
	pg := pagination.FromContext(c)
	params := map[string]string{}
	for _, k := range []string{"patient", "status"} {
		if v := c.QueryParam(k); v != "" {
			params[k] = v
		}
	}
	items, total, err := h.svc.SearchMeasureReports(c.Request().Context(), params, pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *Handler) UpdateMeasureReport(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var mr MeasureReport
	if err := c.Bind(&mr); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	mr.ID = id
	if err := h.svc.UpdateMeasureReport(c.Request().Context(), &mr); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, mr)
}

func (h *Handler) DeleteMeasureReport(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeleteMeasureReport(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

// -- FHIR Endpoints --

func (h *Handler) SearchMeasureReportsFHIR(c echo.Context) error {
	pg := pagination.FromContext(c)
	params := map[string]string{}
	for _, k := range []string{"patient", "status"} {
		if v := c.QueryParam(k); v != "" {
			params[k] = v
		}
	}
	items, total, err := h.svc.SearchMeasureReports(c.Request().Context(), params, pg.Limit, pg.Offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	resources := make([]interface{}, len(items))
	for i, item := range items {
		resources[i] = item.ToFHIR()
	}
	return c.JSON(http.StatusOK, fhir.NewSearchBundle(resources, total, "/fhir/MeasureReport"))
}

func (h *Handler) GetMeasureReportFHIR(c echo.Context) error {
	mr, err := h.svc.GetMeasureReportByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("MeasureReport", c.Param("id")))
	}
	return c.JSON(http.StatusOK, mr.ToFHIR())
}

func (h *Handler) CreateMeasureReportFHIR(c echo.Context) error {
	var mr MeasureReport
	if err := c.Bind(&mr); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	if err := h.svc.CreateMeasureReport(c.Request().Context(), &mr); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	c.Response().Header().Set("Location", "/fhir/MeasureReport/"+mr.FHIRID)
	return c.JSON(http.StatusCreated, mr.ToFHIR())
}

func (h *Handler) UpdateMeasureReportFHIR(c echo.Context) error {
	var mr MeasureReport
	if err := c.Bind(&mr); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	existing, err := h.svc.GetMeasureReportByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("MeasureReport", c.Param("id")))
	}
	mr.ID = existing.ID
	mr.FHIRID = existing.FHIRID
	if err := h.svc.UpdateMeasureReport(c.Request().Context(), &mr); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, mr.ToFHIR())
}

func (h *Handler) DeleteMeasureReportFHIR(c echo.Context) error {
	existing, err := h.svc.GetMeasureReportByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("MeasureReport", c.Param("id")))
	}
	if err := h.svc.DeleteMeasureReport(c.Request().Context(), existing.ID); err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) PatchMeasureReportFHIR(c echo.Context) error {
	return h.handlePatch(c, c.Param("id"))
}

func (h *Handler) VreadMeasureReportFHIR(c echo.Context) error {
	mr, err := h.svc.GetMeasureReportByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("MeasureReport", c.Param("id")))
	}
	result := mr.ToFHIR()
	fhir.SetVersionHeaders(c, 1, mr.UpdatedAt.Format("2006-01-02T15:04:05Z"))
	return c.JSON(http.StatusOK, result)
}

func (h *Handler) HistoryMeasureReportFHIR(c echo.Context) error {
	mr, err := h.svc.GetMeasureReportByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("MeasureReport", c.Param("id")))
	}
	result := mr.ToFHIR()
	raw, _ := json.Marshal(result)
	entry := &fhir.HistoryEntry{
		ResourceType: "MeasureReport", ResourceID: mr.FHIRID, VersionID: 1,
		Resource: raw, Action: "create", Timestamp: mr.CreatedAt,
	}
	return c.JSON(http.StatusOK, fhir.NewHistoryBundle([]*fhir.HistoryEntry{entry}, 1, "/fhir"))
}

func (h *Handler) handlePatch(c echo.Context, fhirID string) error {
	contentType := c.Request().Header.Get("Content-Type")
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome("failed to read request body"))
	}

	existing, err := h.svc.GetMeasureReportByFHIRID(c.Request().Context(), fhirID)
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("MeasureReport", fhirID))
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
	if err := h.svc.UpdateMeasureReport(c.Request().Context(), existing); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, existing.ToFHIR())
}
