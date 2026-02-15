package fhirlist

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
	read.GET("/lists", h.ListFHIRLists)
	read.GET("/lists/:id", h.GetFHIRList)
	read.GET("/lists/:id/entries", h.GetFHIRListEntries)

	write := api.Group("", role)
	write.POST("/lists", h.CreateFHIRList)
	write.PUT("/lists/:id", h.UpdateFHIRList)
	write.DELETE("/lists/:id", h.DeleteFHIRList)
	write.POST("/lists/:id/entries", h.AddFHIRListEntry)

	fhirRead := fhirGroup.Group("", role)
	fhirRead.GET("/List", h.SearchFHIRListsFHIR)
	fhirRead.GET("/List/:id", h.GetFHIRListFHIR)

	fhirWrite := fhirGroup.Group("", role)
	fhirWrite.POST("/List", h.CreateFHIRListFHIR)
	fhirWrite.PUT("/List/:id", h.UpdateFHIRListFHIR)
	fhirWrite.DELETE("/List/:id", h.DeleteFHIRListFHIR)
	fhirWrite.PATCH("/List/:id", h.PatchFHIRListFHIR)

	fhirRead.POST("/List/_search", h.SearchFHIRListsFHIR)

	fhirRead.GET("/List/:id/_history/:vid", h.VreadFHIRListFHIR)
	fhirRead.GET("/List/:id/_history", h.HistoryFHIRListFHIR)
}

// -- REST Endpoints --

func (h *Handler) CreateFHIRList(c echo.Context) error {
	var l FHIRList
	if err := c.Bind(&l); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateFHIRList(c.Request().Context(), &l); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, l)
}

func (h *Handler) GetFHIRList(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	l, err := h.svc.GetFHIRList(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "list not found")
	}
	return c.JSON(http.StatusOK, l)
}

func (h *Handler) ListFHIRLists(c echo.Context) error {
	pg := pagination.FromContext(c)
	params := map[string]string{}
	for _, k := range []string{"patient", "status"} {
		if v := c.QueryParam(k); v != "" {
			params[k] = v
		}
	}
	items, total, err := h.svc.SearchFHIRLists(c.Request().Context(), params, pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *Handler) UpdateFHIRList(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var l FHIRList
	if err := c.Bind(&l); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	l.ID = id
	if err := h.svc.UpdateFHIRList(c.Request().Context(), &l); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, l)
}

func (h *Handler) DeleteFHIRList(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeleteFHIRList(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) AddFHIRListEntry(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var entry FHIRListEntry
	if err := c.Bind(&entry); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	entry.ListID = id
	if err := h.svc.AddEntry(c.Request().Context(), &entry); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, entry)
}

func (h *Handler) GetFHIRListEntries(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	items, err := h.svc.GetEntries(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, items)
}

// -- FHIR Endpoints --

func (h *Handler) SearchFHIRListsFHIR(c echo.Context) error {
	pg := pagination.FromContext(c)
	params := map[string]string{}
	for _, k := range []string{"patient", "status"} {
		if v := c.QueryParam(k); v != "" {
			params[k] = v
		}
	}
	items, total, err := h.svc.SearchFHIRLists(c.Request().Context(), params, pg.Limit, pg.Offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	resources := make([]interface{}, len(items))
	for i, item := range items {
		resources[i] = item.ToFHIR()
	}
	return c.JSON(http.StatusOK, fhir.NewSearchBundle(resources, total, "/fhir/List"))
}

func (h *Handler) GetFHIRListFHIR(c echo.Context) error {
	l, err := h.svc.GetFHIRListByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("List", c.Param("id")))
	}
	result := l.ToFHIR()
	// Include entries in the FHIR response
	entries, _ := h.svc.GetEntries(c.Request().Context(), l.ID)
	if len(entries) > 0 {
		fhirEntries := make([]map[string]interface{}, len(entries))
		for i, e := range entries {
			fhirEntries[i] = e.ToFHIR()
		}
		result["entry"] = fhirEntries
	}
	return c.JSON(http.StatusOK, result)
}

func (h *Handler) CreateFHIRListFHIR(c echo.Context) error {
	var l FHIRList
	if err := c.Bind(&l); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	if err := h.svc.CreateFHIRList(c.Request().Context(), &l); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	c.Response().Header().Set("Location", "/fhir/List/"+l.FHIRID)
	return c.JSON(http.StatusCreated, l.ToFHIR())
}

func (h *Handler) UpdateFHIRListFHIR(c echo.Context) error {
	var l FHIRList
	if err := c.Bind(&l); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	existing, err := h.svc.GetFHIRListByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("List", c.Param("id")))
	}
	l.ID = existing.ID
	l.FHIRID = existing.FHIRID
	if err := h.svc.UpdateFHIRList(c.Request().Context(), &l); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, l.ToFHIR())
}

func (h *Handler) DeleteFHIRListFHIR(c echo.Context) error {
	existing, err := h.svc.GetFHIRListByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("List", c.Param("id")))
	}
	if err := h.svc.DeleteFHIRList(c.Request().Context(), existing.ID); err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) PatchFHIRListFHIR(c echo.Context) error {
	return h.handlePatch(c, c.Param("id"))
}

func (h *Handler) VreadFHIRListFHIR(c echo.Context) error {
	l, err := h.svc.GetFHIRListByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("List", c.Param("id")))
	}
	result := l.ToFHIR()
	fhir.SetVersionHeaders(c, 1, l.UpdatedAt.Format("2006-01-02T15:04:05Z"))
	return c.JSON(http.StatusOK, result)
}

func (h *Handler) HistoryFHIRListFHIR(c echo.Context) error {
	l, err := h.svc.GetFHIRListByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("List", c.Param("id")))
	}
	result := l.ToFHIR()
	raw, _ := json.Marshal(result)
	entry := &fhir.HistoryEntry{
		ResourceType: "List", ResourceID: l.FHIRID, VersionID: 1,
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

	existing, err := h.svc.GetFHIRListByFHIRID(c.Request().Context(), fhirID)
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("List", fhirID))
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
	if v, ok := patched["mode"].(string); ok {
		existing.Mode = v
	}
	if v, ok := patched["title"].(string); ok {
		existing.Title = &v
	}
	if err := h.svc.UpdateFHIRList(c.Request().Context(), existing); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, existing.ToFHIR())
}
