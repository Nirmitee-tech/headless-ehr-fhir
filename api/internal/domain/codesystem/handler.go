package codesystem

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
	read.GET("/code-systems", h.ListCodeSystems)
	read.GET("/code-systems/:id", h.GetCodeSystem)

	write := api.Group("", role)
	write.POST("/code-systems", h.CreateCodeSystem)
	write.PUT("/code-systems/:id", h.UpdateCodeSystem)
	write.DELETE("/code-systems/:id", h.DeleteCodeSystem)

	fhirRead := fhirGroup.Group("", role)
	fhirRead.GET("/CodeSystem", h.SearchCodeSystemsFHIR)
	fhirRead.GET("/CodeSystem/:id", h.GetCodeSystemFHIR)

	fhirWrite := fhirGroup.Group("", role)
	fhirWrite.POST("/CodeSystem", h.CreateCodeSystemFHIR)
	fhirWrite.PUT("/CodeSystem/:id", h.UpdateCodeSystemFHIR)
	fhirWrite.DELETE("/CodeSystem/:id", h.DeleteCodeSystemFHIR)
	fhirWrite.PATCH("/CodeSystem/:id", h.PatchCodeSystemFHIR)

	fhirRead.POST("/CodeSystem/_search", h.SearchCodeSystemsFHIR)
	fhirRead.GET("/CodeSystem/:id/_history/:vid", h.VreadCodeSystemFHIR)
	fhirRead.GET("/CodeSystem/:id/_history", h.HistoryCodeSystemFHIR)
}

// -- REST Endpoints --

func (h *Handler) CreateCodeSystem(c echo.Context) error {
	var cs CodeSystem
	if err := c.Bind(&cs); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateCodeSystem(c.Request().Context(), &cs); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, cs)
}

func (h *Handler) GetCodeSystem(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	cs, err := h.svc.GetCodeSystem(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "code system not found")
	}
	return c.JSON(http.StatusOK, cs)
}

func (h *Handler) ListCodeSystems(c echo.Context) error {
	pg := pagination.FromContext(c)
	items, total, err := h.svc.SearchCodeSystems(c.Request().Context(), nil, pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *Handler) UpdateCodeSystem(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var cs CodeSystem
	if err := c.Bind(&cs); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	cs.ID = id
	if err := h.svc.UpdateCodeSystem(c.Request().Context(), &cs); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, cs)
}

func (h *Handler) DeleteCodeSystem(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeleteCodeSystem(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

// -- FHIR Endpoints --

func (h *Handler) SearchCodeSystemsFHIR(c echo.Context) error {
	pg := pagination.FromContext(c)
	params := fhir.ExtractSearchParams(c)
	items, total, err := h.svc.SearchCodeSystems(c.Request().Context(), params, pg.Limit, pg.Offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	resources := make([]interface{}, len(items))
	for i, item := range items {
		resources[i] = item.ToFHIR()
	}
	return c.JSON(http.StatusOK, fhir.NewSearchBundle(resources, total, "/fhir/CodeSystem"))
}

func (h *Handler) GetCodeSystemFHIR(c echo.Context) error {
	cs, err := h.svc.GetCodeSystemByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("CodeSystem", c.Param("id")))
	}
	return c.JSON(http.StatusOK, cs.ToFHIR())
}

func (h *Handler) CreateCodeSystemFHIR(c echo.Context) error {
	var cs CodeSystem
	if err := c.Bind(&cs); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	if err := h.svc.CreateCodeSystem(c.Request().Context(), &cs); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	c.Response().Header().Set("Location", "/fhir/CodeSystem/"+cs.FHIRID)
	return c.JSON(http.StatusCreated, cs.ToFHIR())
}

func (h *Handler) UpdateCodeSystemFHIR(c echo.Context) error {
	var cs CodeSystem
	if err := c.Bind(&cs); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	existing, err := h.svc.GetCodeSystemByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("CodeSystem", c.Param("id")))
	}
	cs.ID = existing.ID
	cs.FHIRID = existing.FHIRID
	if err := h.svc.UpdateCodeSystem(c.Request().Context(), &cs); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, cs.ToFHIR())
}

func (h *Handler) DeleteCodeSystemFHIR(c echo.Context) error {
	existing, err := h.svc.GetCodeSystemByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("CodeSystem", c.Param("id")))
	}
	if err := h.svc.DeleteCodeSystem(c.Request().Context(), existing.ID); err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) PatchCodeSystemFHIR(c echo.Context) error {
	return h.handlePatch(c, c.Param("id"))
}

func (h *Handler) VreadCodeSystemFHIR(c echo.Context) error {
	cs, err := h.svc.GetCodeSystemByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("CodeSystem", c.Param("id")))
	}
	result := cs.ToFHIR()
	fhir.SetVersionHeaders(c, 1, cs.UpdatedAt.Format("2006-01-02T15:04:05Z"))
	return c.JSON(http.StatusOK, result)
}

func (h *Handler) HistoryCodeSystemFHIR(c echo.Context) error {
	cs, err := h.svc.GetCodeSystemByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("CodeSystem", c.Param("id")))
	}
	result := cs.ToFHIR()
	raw, _ := json.Marshal(result)
	entry := &fhir.HistoryEntry{
		ResourceType: "CodeSystem", ResourceID: cs.FHIRID, VersionID: 1,
		Resource: raw, Action: "create", Timestamp: cs.CreatedAt,
	}
	return c.JSON(http.StatusOK, fhir.NewHistoryBundle([]*fhir.HistoryEntry{entry}, 1, "/fhir"))
}

func (h *Handler) handlePatch(c echo.Context, fhirID string) error {
	contentType := c.Request().Header.Get("Content-Type")
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome("failed to read request body"))
	}
	existing, err := h.svc.GetCodeSystemByFHIRID(c.Request().Context(), fhirID)
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("CodeSystem", fhirID))
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
	if err := h.svc.UpdateCodeSystem(c.Request().Context(), existing); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, existing.ToFHIR())
}
