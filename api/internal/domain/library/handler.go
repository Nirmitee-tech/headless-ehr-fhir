package library

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
	read.GET("/libraries", h.ListLibraries)
	read.GET("/libraries/:id", h.GetLibrary)

	write := api.Group("", role)
	write.POST("/libraries", h.CreateLibrary)
	write.PUT("/libraries/:id", h.UpdateLibrary)
	write.DELETE("/libraries/:id", h.DeleteLibrary)

	fhirRead := fhirGroup.Group("", role)
	fhirRead.GET("/Library", h.SearchLibrariesFHIR)
	fhirRead.GET("/Library/:id", h.GetLibraryFHIR)

	fhirWrite := fhirGroup.Group("", role)
	fhirWrite.POST("/Library", h.CreateLibraryFHIR)
	fhirWrite.PUT("/Library/:id", h.UpdateLibraryFHIR)
	fhirWrite.DELETE("/Library/:id", h.DeleteLibraryFHIR)
	fhirWrite.PATCH("/Library/:id", h.PatchLibraryFHIR)

	fhirRead.POST("/Library/_search", h.SearchLibrariesFHIR)
	fhirRead.GET("/Library/:id/_history/:vid", h.VreadLibraryFHIR)
	fhirRead.GET("/Library/:id/_history", h.HistoryLibraryFHIR)
}

// -- REST Endpoints --

func (h *Handler) CreateLibrary(c echo.Context) error {
	var l Library
	if err := c.Bind(&l); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateLibrary(c.Request().Context(), &l); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, l)
}

func (h *Handler) GetLibrary(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	l, err := h.svc.GetLibrary(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "library not found")
	}
	return c.JSON(http.StatusOK, l)
}

func (h *Handler) ListLibraries(c echo.Context) error {
	pg := pagination.FromContext(c)
	items, total, err := h.svc.SearchLibraries(c.Request().Context(), nil, pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *Handler) UpdateLibrary(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var l Library
	if err := c.Bind(&l); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	l.ID = id
	if err := h.svc.UpdateLibrary(c.Request().Context(), &l); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, l)
}

func (h *Handler) DeleteLibrary(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeleteLibrary(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

// -- FHIR Endpoints --

func (h *Handler) SearchLibrariesFHIR(c echo.Context) error {
	pg := pagination.FromContext(c)
	params := fhir.ExtractSearchParams(c)
	items, total, err := h.svc.SearchLibraries(c.Request().Context(), params, pg.Limit, pg.Offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	resources := make([]interface{}, len(items))
	for i, item := range items {
		resources[i] = item.ToFHIR()
	}
	return c.JSON(http.StatusOK, fhir.NewSearchBundle(resources, total, "/fhir/Library"))
}

func (h *Handler) GetLibraryFHIR(c echo.Context) error {
	l, err := h.svc.GetLibraryByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Library", c.Param("id")))
	}
	fhir.SetVersionHeaders(c, 1, l.UpdatedAt.Format("2006-01-02T15:04:05Z"))
	return c.JSON(http.StatusOK, l.ToFHIR())
}

func (h *Handler) CreateLibraryFHIR(c echo.Context) error {
	var l Library
	if err := c.Bind(&l); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	if err := h.svc.CreateLibrary(c.Request().Context(), &l); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	c.Response().Header().Set("Location", "/fhir/Library/"+l.FHIRID)
	return c.JSON(http.StatusCreated, l.ToFHIR())
}

func (h *Handler) UpdateLibraryFHIR(c echo.Context) error {
	var l Library
	if err := c.Bind(&l); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	existing, err := h.svc.GetLibraryByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Library", c.Param("id")))
	}
	l.ID = existing.ID
	l.FHIRID = existing.FHIRID
	if err := h.svc.UpdateLibrary(c.Request().Context(), &l); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, l.ToFHIR())
}

func (h *Handler) DeleteLibraryFHIR(c echo.Context) error {
	existing, err := h.svc.GetLibraryByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Library", c.Param("id")))
	}
	if err := h.svc.DeleteLibrary(c.Request().Context(), existing.ID); err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) PatchLibraryFHIR(c echo.Context) error {
	return h.handlePatch(c, c.Param("id"))
}

func (h *Handler) VreadLibraryFHIR(c echo.Context) error {
	l, err := h.svc.GetLibraryByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Library", c.Param("id")))
	}
	result := l.ToFHIR()
	fhir.SetVersionHeaders(c, 1, l.UpdatedAt.Format("2006-01-02T15:04:05Z"))
	return c.JSON(http.StatusOK, result)
}

func (h *Handler) HistoryLibraryFHIR(c echo.Context) error {
	l, err := h.svc.GetLibraryByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Library", c.Param("id")))
	}
	result := l.ToFHIR()
	raw, _ := json.Marshal(result)
	entry := &fhir.HistoryEntry{
		ResourceType: "Library", ResourceID: l.FHIRID, VersionID: 1,
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
	existing, err := h.svc.GetLibraryByFHIRID(c.Request().Context(), fhirID)
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Library", fhirID))
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
	if err := h.svc.UpdateLibrary(c.Request().Context(), existing); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, existing.ToFHIR())
}
