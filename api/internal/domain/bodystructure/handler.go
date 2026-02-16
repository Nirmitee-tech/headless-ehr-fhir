package bodystructure

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
	read.GET("/body-structures", h.ListBodyStructures)
	read.GET("/body-structures/:id", h.GetBodyStructure)

	write := api.Group("", role)
	write.POST("/body-structures", h.CreateBodyStructure)
	write.PUT("/body-structures/:id", h.UpdateBodyStructure)
	write.DELETE("/body-structures/:id", h.DeleteBodyStructure)

	fhirRead := fhirGroup.Group("", role)
	fhirRead.GET("/BodyStructure", h.SearchBodyStructuresFHIR)
	fhirRead.GET("/BodyStructure/:id", h.GetBodyStructureFHIR)

	fhirWrite := fhirGroup.Group("", role)
	fhirWrite.POST("/BodyStructure", h.CreateBodyStructureFHIR)
	fhirWrite.PUT("/BodyStructure/:id", h.UpdateBodyStructureFHIR)
	fhirWrite.DELETE("/BodyStructure/:id", h.DeleteBodyStructureFHIR)
	fhirWrite.PATCH("/BodyStructure/:id", h.PatchBodyStructureFHIR)

	fhirRead.POST("/BodyStructure/_search", h.SearchBodyStructuresFHIR)
	fhirRead.GET("/BodyStructure/:id/_history/:vid", h.VreadBodyStructureFHIR)
	fhirRead.GET("/BodyStructure/:id/_history", h.HistoryBodyStructureFHIR)
}

func (h *Handler) CreateBodyStructure(c echo.Context) error {
	var b BodyStructure
	if err := c.Bind(&b); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateBodyStructure(c.Request().Context(), &b); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, b)
}

func (h *Handler) GetBodyStructure(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	b, err := h.svc.GetBodyStructure(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "body structure not found")
	}
	return c.JSON(http.StatusOK, b)
}

func (h *Handler) ListBodyStructures(c echo.Context) error {
	pg := pagination.FromContext(c)
	items, total, err := h.svc.SearchBodyStructures(c.Request().Context(), nil, pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *Handler) UpdateBodyStructure(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var b BodyStructure
	if err := c.Bind(&b); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	b.ID = id
	if err := h.svc.UpdateBodyStructure(c.Request().Context(), &b); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, b)
}

func (h *Handler) DeleteBodyStructure(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeleteBodyStructure(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) SearchBodyStructuresFHIR(c echo.Context) error {
	pg := pagination.FromContext(c)
	params := map[string]string{}
	for _, k := range []string{"patient"} {
		if v := c.QueryParam(k); v != "" {
			params[k] = v
		}
	}
	items, total, err := h.svc.SearchBodyStructures(c.Request().Context(), params, pg.Limit, pg.Offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	resources := make([]interface{}, len(items))
	for i, item := range items {
		resources[i] = item.ToFHIR()
	}
	return c.JSON(http.StatusOK, fhir.NewSearchBundle(resources, total, "/fhir/BodyStructure"))
}

func (h *Handler) GetBodyStructureFHIR(c echo.Context) error {
	b, err := h.svc.GetBodyStructureByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("BodyStructure", c.Param("id")))
	}
	return c.JSON(http.StatusOK, b.ToFHIR())
}

func (h *Handler) CreateBodyStructureFHIR(c echo.Context) error {
	var b BodyStructure
	if err := c.Bind(&b); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	if err := h.svc.CreateBodyStructure(c.Request().Context(), &b); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	c.Response().Header().Set("Location", "/fhir/BodyStructure/"+b.FHIRID)
	return c.JSON(http.StatusCreated, b.ToFHIR())
}

func (h *Handler) UpdateBodyStructureFHIR(c echo.Context) error {
	var b BodyStructure
	if err := c.Bind(&b); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	existing, err := h.svc.GetBodyStructureByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("BodyStructure", c.Param("id")))
	}
	b.ID = existing.ID
	b.FHIRID = existing.FHIRID
	if err := h.svc.UpdateBodyStructure(c.Request().Context(), &b); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, b.ToFHIR())
}

func (h *Handler) DeleteBodyStructureFHIR(c echo.Context) error {
	existing, err := h.svc.GetBodyStructureByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("BodyStructure", c.Param("id")))
	}
	if err := h.svc.DeleteBodyStructure(c.Request().Context(), existing.ID); err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) PatchBodyStructureFHIR(c echo.Context) error {
	return h.handlePatch(c, c.Param("id"))
}

func (h *Handler) VreadBodyStructureFHIR(c echo.Context) error {
	b, err := h.svc.GetBodyStructureByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("BodyStructure", c.Param("id")))
	}
	result := b.ToFHIR()
	fhir.SetVersionHeaders(c, 1, b.UpdatedAt.Format("2006-01-02T15:04:05Z"))
	return c.JSON(http.StatusOK, result)
}

func (h *Handler) HistoryBodyStructureFHIR(c echo.Context) error {
	b, err := h.svc.GetBodyStructureByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("BodyStructure", c.Param("id")))
	}
	result := b.ToFHIR()
	raw, _ := json.Marshal(result)
	entry := &fhir.HistoryEntry{
		ResourceType: "BodyStructure", ResourceID: b.FHIRID, VersionID: 1,
		Resource: raw, Action: "create", Timestamp: b.CreatedAt,
	}
	return c.JSON(http.StatusOK, fhir.NewHistoryBundle([]*fhir.HistoryEntry{entry}, 1, "/fhir"))
}

func (h *Handler) handlePatch(c echo.Context, fhirID string) error {
	contentType := c.Request().Header.Get("Content-Type")
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome("failed to read request body"))
	}
	existing, err := h.svc.GetBodyStructureByFHIRID(c.Request().Context(), fhirID)
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("BodyStructure", fhirID))
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
	_ = patched
	if err := h.svc.UpdateBodyStructure(c.Request().Context(), existing); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, existing.ToFHIR())
}
