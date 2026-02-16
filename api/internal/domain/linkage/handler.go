package linkage

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
	read.GET("/linkages", h.ListLinkages)
	read.GET("/linkages/:id", h.GetLinkage)

	write := api.Group("", role)
	write.POST("/linkages", h.CreateLinkage)
	write.PUT("/linkages/:id", h.UpdateLinkage)
	write.DELETE("/linkages/:id", h.DeleteLinkage)

	fhirRead := fhirGroup.Group("", role)
	fhirRead.GET("/Linkage", h.SearchLinkagesFHIR)
	fhirRead.GET("/Linkage/:id", h.GetLinkageFHIR)

	fhirWrite := fhirGroup.Group("", role)
	fhirWrite.POST("/Linkage", h.CreateLinkageFHIR)
	fhirWrite.PUT("/Linkage/:id", h.UpdateLinkageFHIR)
	fhirWrite.DELETE("/Linkage/:id", h.DeleteLinkageFHIR)
	fhirWrite.PATCH("/Linkage/:id", h.PatchLinkageFHIR)

	fhirRead.POST("/Linkage/_search", h.SearchLinkagesFHIR)
	fhirRead.GET("/Linkage/:id/_history/:vid", h.VreadLinkageFHIR)
	fhirRead.GET("/Linkage/:id/_history", h.HistoryLinkageFHIR)
}

// -- REST Endpoints --

func (h *Handler) CreateLinkage(c echo.Context) error {
	var l Linkage
	if err := c.Bind(&l); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateLinkage(c.Request().Context(), &l); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, l)
}

func (h *Handler) GetLinkage(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	l, err := h.svc.GetLinkage(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "linkage not found")
	}
	return c.JSON(http.StatusOK, l)
}

func (h *Handler) ListLinkages(c echo.Context) error {
	pg := pagination.FromContext(c)
	items, total, err := h.svc.SearchLinkages(c.Request().Context(), nil, pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *Handler) UpdateLinkage(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var l Linkage
	if err := c.Bind(&l); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	l.ID = id
	if err := h.svc.UpdateLinkage(c.Request().Context(), &l); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, l)
}

func (h *Handler) DeleteLinkage(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeleteLinkage(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

// -- FHIR Endpoints --

func (h *Handler) SearchLinkagesFHIR(c echo.Context) error {
	pg := pagination.FromContext(c)
	params := fhir.ExtractSearchParams(c)
	items, total, err := h.svc.SearchLinkages(c.Request().Context(), params, pg.Limit, pg.Offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	resources := make([]interface{}, len(items))
	for i, item := range items {
		resources[i] = item.ToFHIR()
	}
	return c.JSON(http.StatusOK, fhir.NewSearchBundle(resources, total, "/fhir/Linkage"))
}

func (h *Handler) GetLinkageFHIR(c echo.Context) error {
	l, err := h.svc.GetLinkageByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Linkage", c.Param("id")))
	}
	return c.JSON(http.StatusOK, l.ToFHIR())
}

func (h *Handler) CreateLinkageFHIR(c echo.Context) error {
	var l Linkage
	if err := c.Bind(&l); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	if err := h.svc.CreateLinkage(c.Request().Context(), &l); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	c.Response().Header().Set("Location", "/fhir/Linkage/"+l.FHIRID)
	return c.JSON(http.StatusCreated, l.ToFHIR())
}

func (h *Handler) UpdateLinkageFHIR(c echo.Context) error {
	var l Linkage
	if err := c.Bind(&l); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	existing, err := h.svc.GetLinkageByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Linkage", c.Param("id")))
	}
	l.ID = existing.ID
	l.FHIRID = existing.FHIRID
	if err := h.svc.UpdateLinkage(c.Request().Context(), &l); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, l.ToFHIR())
}

func (h *Handler) DeleteLinkageFHIR(c echo.Context) error {
	existing, err := h.svc.GetLinkageByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Linkage", c.Param("id")))
	}
	if err := h.svc.DeleteLinkage(c.Request().Context(), existing.ID); err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) PatchLinkageFHIR(c echo.Context) error {
	return h.handlePatch(c, c.Param("id"))
}

func (h *Handler) VreadLinkageFHIR(c echo.Context) error {
	l, err := h.svc.GetLinkageByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Linkage", c.Param("id")))
	}
	result := l.ToFHIR()
	fhir.SetVersionHeaders(c, 1, l.UpdatedAt.Format("2006-01-02T15:04:05Z"))
	return c.JSON(http.StatusOK, result)
}

func (h *Handler) HistoryLinkageFHIR(c echo.Context) error {
	l, err := h.svc.GetLinkageByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Linkage", c.Param("id")))
	}
	result := l.ToFHIR()
	raw, _ := json.Marshal(result)
	entry := &fhir.HistoryEntry{
		ResourceType: "Linkage", ResourceID: l.FHIRID, VersionID: 1,
		Resource: raw, Action: "create", Timestamp: l.CreatedAt,
	}
	return c.JSON(http.StatusOK, fhir.NewHistoryBundle([]*fhir.HistoryEntry{entry}, 1, "/fhir"))
}

func (h *Handler) handlePatch(c echo.Context, fhirID string) error {
	contentType := c.Request().Header.Get("Content-Type")
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome("failed to read request body"))
	}
	existing, err := h.svc.GetLinkageByFHIRID(c.Request().Context(), fhirID)
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Linkage", fhirID))
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
	if v, ok := patched["active"].(bool); ok {
		existing.Active = v
	}
	if err := h.svc.UpdateLinkage(c.Request().Context(), existing); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, existing.ToFHIR())
}
