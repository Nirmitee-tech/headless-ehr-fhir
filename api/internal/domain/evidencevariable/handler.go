package evidencevariable

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
	read.GET("/evidence-variables", h.ListEvidenceVariables)
	read.GET("/evidence-variables/:id", h.GetEvidenceVariable)

	write := api.Group("", role)
	write.POST("/evidence-variables", h.CreateEvidenceVariable)
	write.PUT("/evidence-variables/:id", h.UpdateEvidenceVariable)
	write.DELETE("/evidence-variables/:id", h.DeleteEvidenceVariable)

	fhirRead := fhirGroup.Group("", role)
	fhirRead.GET("/EvidenceVariable", h.SearchEvidenceVariablesFHIR)
	fhirRead.GET("/EvidenceVariable/:id", h.GetEvidenceVariableFHIR)

	fhirWrite := fhirGroup.Group("", role)
	fhirWrite.POST("/EvidenceVariable", h.CreateEvidenceVariableFHIR)
	fhirWrite.PUT("/EvidenceVariable/:id", h.UpdateEvidenceVariableFHIR)
	fhirWrite.DELETE("/EvidenceVariable/:id", h.DeleteEvidenceVariableFHIR)
	fhirWrite.PATCH("/EvidenceVariable/:id", h.PatchEvidenceVariableFHIR)

	fhirRead.POST("/EvidenceVariable/_search", h.SearchEvidenceVariablesFHIR)
	fhirRead.GET("/EvidenceVariable/:id/_history/:vid", h.VreadEvidenceVariableFHIR)
	fhirRead.GET("/EvidenceVariable/:id/_history", h.HistoryEvidenceVariableFHIR)
}

// -- REST Endpoints --

func (h *Handler) CreateEvidenceVariable(c echo.Context) error {
	var e EvidenceVariable
	if err := c.Bind(&e); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateEvidenceVariable(c.Request().Context(), &e); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, e)
}

func (h *Handler) GetEvidenceVariable(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	e, err := h.svc.GetEvidenceVariable(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "evidence variable not found")
	}
	return c.JSON(http.StatusOK, e)
}

func (h *Handler) ListEvidenceVariables(c echo.Context) error {
	pg := pagination.FromContext(c)
	items, total, err := h.svc.SearchEvidenceVariables(c.Request().Context(), nil, pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *Handler) UpdateEvidenceVariable(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var e EvidenceVariable
	if err := c.Bind(&e); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	e.ID = id
	if err := h.svc.UpdateEvidenceVariable(c.Request().Context(), &e); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, e)
}

func (h *Handler) DeleteEvidenceVariable(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeleteEvidenceVariable(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

// -- FHIR Endpoints --

func (h *Handler) SearchEvidenceVariablesFHIR(c echo.Context) error {
	pg := pagination.FromContext(c)
	params := fhir.ExtractSearchParams(c)
	items, total, err := h.svc.SearchEvidenceVariables(c.Request().Context(), params, pg.Limit, pg.Offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	resources := make([]interface{}, len(items))
	for i, item := range items {
		resources[i] = item.ToFHIR()
	}
	return c.JSON(http.StatusOK, fhir.NewSearchBundle(resources, total, "/fhir/EvidenceVariable"))
}

func (h *Handler) GetEvidenceVariableFHIR(c echo.Context) error {
	e, err := h.svc.GetEvidenceVariableByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("EvidenceVariable", c.Param("id")))
	}
	return c.JSON(http.StatusOK, e.ToFHIR())
}

func (h *Handler) CreateEvidenceVariableFHIR(c echo.Context) error {
	var e EvidenceVariable
	if err := c.Bind(&e); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	if err := h.svc.CreateEvidenceVariable(c.Request().Context(), &e); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	c.Response().Header().Set("Location", "/fhir/EvidenceVariable/"+e.FHIRID)
	return c.JSON(http.StatusCreated, e.ToFHIR())
}

func (h *Handler) UpdateEvidenceVariableFHIR(c echo.Context) error {
	var e EvidenceVariable
	if err := c.Bind(&e); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	existing, err := h.svc.GetEvidenceVariableByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("EvidenceVariable", c.Param("id")))
	}
	e.ID = existing.ID
	e.FHIRID = existing.FHIRID
	if err := h.svc.UpdateEvidenceVariable(c.Request().Context(), &e); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, e.ToFHIR())
}

func (h *Handler) DeleteEvidenceVariableFHIR(c echo.Context) error {
	existing, err := h.svc.GetEvidenceVariableByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("EvidenceVariable", c.Param("id")))
	}
	if err := h.svc.DeleteEvidenceVariable(c.Request().Context(), existing.ID); err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) PatchEvidenceVariableFHIR(c echo.Context) error {
	return h.handlePatch(c, c.Param("id"))
}

func (h *Handler) VreadEvidenceVariableFHIR(c echo.Context) error {
	e, err := h.svc.GetEvidenceVariableByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("EvidenceVariable", c.Param("id")))
	}
	result := e.ToFHIR()
	fhir.SetVersionHeaders(c, 1, e.UpdatedAt.Format("2006-01-02T15:04:05Z"))
	return c.JSON(http.StatusOK, result)
}

func (h *Handler) HistoryEvidenceVariableFHIR(c echo.Context) error {
	e, err := h.svc.GetEvidenceVariableByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("EvidenceVariable", c.Param("id")))
	}
	result := e.ToFHIR()
	raw, _ := json.Marshal(result)
	entry := &fhir.HistoryEntry{
		ResourceType: "EvidenceVariable", ResourceID: e.FHIRID, VersionID: 1,
		Resource: raw, Action: "create", Timestamp: e.CreatedAt,
	}
	return c.JSON(http.StatusOK, fhir.NewHistoryBundle([]*fhir.HistoryEntry{entry}, 1, "/fhir"))
}

func (h *Handler) handlePatch(c echo.Context, fhirID string) error {
	contentType := c.Request().Header.Get("Content-Type")
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome("failed to read request body"))
	}
	existing, err := h.svc.GetEvidenceVariableByFHIRID(c.Request().Context(), fhirID)
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("EvidenceVariable", fhirID))
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
	if err := h.svc.UpdateEvidenceVariable(c.Request().Context(), existing); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, existing.ToFHIR())
}
