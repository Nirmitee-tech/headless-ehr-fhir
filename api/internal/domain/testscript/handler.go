package testscript

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
	read.GET("/test-scripts", h.ListTestScripts)
	read.GET("/test-scripts/:id", h.GetTestScript)

	write := api.Group("", role)
	write.POST("/test-scripts", h.CreateTestScript)
	write.PUT("/test-scripts/:id", h.UpdateTestScript)
	write.DELETE("/test-scripts/:id", h.DeleteTestScript)

	fhirRead := fhirGroup.Group("", role)
	fhirRead.GET("/TestScript", h.SearchTestScriptsFHIR)
	fhirRead.GET("/TestScript/:id", h.GetTestScriptFHIR)

	fhirWrite := fhirGroup.Group("", role)
	fhirWrite.POST("/TestScript", h.CreateTestScriptFHIR)
	fhirWrite.PUT("/TestScript/:id", h.UpdateTestScriptFHIR)
	fhirWrite.DELETE("/TestScript/:id", h.DeleteTestScriptFHIR)
	fhirWrite.PATCH("/TestScript/:id", h.PatchTestScriptFHIR)

	fhirRead.POST("/TestScript/_search", h.SearchTestScriptsFHIR)
	fhirRead.GET("/TestScript/:id/_history/:vid", h.VreadTestScriptFHIR)
	fhirRead.GET("/TestScript/:id/_history", h.HistoryTestScriptFHIR)
}

// -- REST Endpoints --

func (h *Handler) CreateTestScript(c echo.Context) error {
	var ts TestScript
	if err := c.Bind(&ts); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateTestScript(c.Request().Context(), &ts); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, ts)
}

func (h *Handler) GetTestScript(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	ts, err := h.svc.GetTestScript(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "test script not found")
	}
	return c.JSON(http.StatusOK, ts)
}

func (h *Handler) ListTestScripts(c echo.Context) error {
	pg := pagination.FromContext(c)
	items, total, err := h.svc.SearchTestScripts(c.Request().Context(), nil, pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *Handler) UpdateTestScript(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var ts TestScript
	if err := c.Bind(&ts); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	ts.ID = id
	if err := h.svc.UpdateTestScript(c.Request().Context(), &ts); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, ts)
}

func (h *Handler) DeleteTestScript(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeleteTestScript(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

// -- FHIR Endpoints --

func (h *Handler) SearchTestScriptsFHIR(c echo.Context) error {
	pg := pagination.FromContext(c)
	params := fhir.ExtractSearchParams(c)
	items, total, err := h.svc.SearchTestScripts(c.Request().Context(), params, pg.Limit, pg.Offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	resources := make([]interface{}, len(items))
	for i, item := range items {
		resources[i] = item.ToFHIR()
	}
	return c.JSON(http.StatusOK, fhir.NewSearchBundle(resources, total, "/fhir/TestScript"))
}

func (h *Handler) GetTestScriptFHIR(c echo.Context) error {
	ts, err := h.svc.GetTestScriptByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("TestScript", c.Param("id")))
	}
	return c.JSON(http.StatusOK, ts.ToFHIR())
}

func (h *Handler) CreateTestScriptFHIR(c echo.Context) error {
	var ts TestScript
	if err := c.Bind(&ts); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	if err := h.svc.CreateTestScript(c.Request().Context(), &ts); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	c.Response().Header().Set("Location", "/fhir/TestScript/"+ts.FHIRID)
	return c.JSON(http.StatusCreated, ts.ToFHIR())
}

func (h *Handler) UpdateTestScriptFHIR(c echo.Context) error {
	var ts TestScript
	if err := c.Bind(&ts); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	existing, err := h.svc.GetTestScriptByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("TestScript", c.Param("id")))
	}
	ts.ID = existing.ID
	ts.FHIRID = existing.FHIRID
	if err := h.svc.UpdateTestScript(c.Request().Context(), &ts); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, ts.ToFHIR())
}

func (h *Handler) DeleteTestScriptFHIR(c echo.Context) error {
	existing, err := h.svc.GetTestScriptByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("TestScript", c.Param("id")))
	}
	if err := h.svc.DeleteTestScript(c.Request().Context(), existing.ID); err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) PatchTestScriptFHIR(c echo.Context) error {
	return h.handlePatch(c, c.Param("id"))
}

func (h *Handler) VreadTestScriptFHIR(c echo.Context) error {
	ts, err := h.svc.GetTestScriptByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("TestScript", c.Param("id")))
	}
	result := ts.ToFHIR()
	fhir.SetVersionHeaders(c, 1, ts.UpdatedAt.Format("2006-01-02T15:04:05Z"))
	return c.JSON(http.StatusOK, result)
}

func (h *Handler) HistoryTestScriptFHIR(c echo.Context) error {
	ts, err := h.svc.GetTestScriptByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("TestScript", c.Param("id")))
	}
	result := ts.ToFHIR()
	raw, _ := json.Marshal(result)
	entry := &fhir.HistoryEntry{
		ResourceType: "TestScript", ResourceID: ts.FHIRID, VersionID: 1,
		Resource: raw, Action: "create", Timestamp: ts.CreatedAt,
	}
	return c.JSON(http.StatusOK, fhir.NewHistoryBundle([]*fhir.HistoryEntry{entry}, 1, "/fhir"))
}

func (h *Handler) handlePatch(c echo.Context, fhirID string) error {
	contentType := c.Request().Header.Get("Content-Type")
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome("failed to read request body"))
	}
	existing, err := h.svc.GetTestScriptByFHIRID(c.Request().Context(), fhirID)
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("TestScript", fhirID))
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
	if err := h.svc.UpdateTestScript(c.Request().Context(), existing); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, existing.ToFHIR())
}
