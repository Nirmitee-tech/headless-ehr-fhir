package basic

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
	read.GET("/basics", h.ListBasics)
	read.GET("/basics/:id", h.GetBasic)

	write := api.Group("", role)
	write.POST("/basics", h.CreateBasic)
	write.PUT("/basics/:id", h.UpdateBasic)
	write.DELETE("/basics/:id", h.DeleteBasic)

	fhirRead := fhirGroup.Group("", role)
	fhirRead.GET("/Basic", h.SearchBasicsFHIR)
	fhirRead.GET("/Basic/:id", h.GetBasicFHIR)

	fhirWrite := fhirGroup.Group("", role)
	fhirWrite.POST("/Basic", h.CreateBasicFHIR)
	fhirWrite.PUT("/Basic/:id", h.UpdateBasicFHIR)
	fhirWrite.DELETE("/Basic/:id", h.DeleteBasicFHIR)
	fhirWrite.PATCH("/Basic/:id", h.PatchBasicFHIR)

	fhirRead.POST("/Basic/_search", h.SearchBasicsFHIR)
	fhirRead.GET("/Basic/:id/_history/:vid", h.VreadBasicFHIR)
	fhirRead.GET("/Basic/:id/_history", h.HistoryBasicFHIR)
}

// -- REST Endpoints --

func (h *Handler) CreateBasic(c echo.Context) error {
	var b Basic
	if err := c.Bind(&b); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateBasic(c.Request().Context(), &b); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, b)
}

func (h *Handler) GetBasic(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	b, err := h.svc.GetBasic(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "basic not found")
	}
	return c.JSON(http.StatusOK, b)
}

func (h *Handler) ListBasics(c echo.Context) error {
	pg := pagination.FromContext(c)
	items, total, err := h.svc.SearchBasics(c.Request().Context(), nil, pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *Handler) UpdateBasic(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var b Basic
	if err := c.Bind(&b); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	b.ID = id
	if err := h.svc.UpdateBasic(c.Request().Context(), &b); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, b)
}

func (h *Handler) DeleteBasic(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeleteBasic(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

// -- FHIR Endpoints --

func (h *Handler) SearchBasicsFHIR(c echo.Context) error {
	pg := pagination.FromContext(c)
	params := fhir.ExtractSearchParams(c)
	items, total, err := h.svc.SearchBasics(c.Request().Context(), params, pg.Limit, pg.Offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	resources := make([]interface{}, len(items))
	for i, item := range items {
		resources[i] = item.ToFHIR()
	}
	return c.JSON(http.StatusOK, fhir.NewSearchBundle(resources, total, "/fhir/Basic"))
}

func (h *Handler) GetBasicFHIR(c echo.Context) error {
	b, err := h.svc.GetBasicByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Basic", c.Param("id")))
	}
	fhir.SetVersionHeaders(c, 1, b.UpdatedAt.Format("2006-01-02T15:04:05Z"))
	return c.JSON(http.StatusOK, b.ToFHIR())
}

func (h *Handler) CreateBasicFHIR(c echo.Context) error {
	var b Basic
	if err := c.Bind(&b); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	if err := h.svc.CreateBasic(c.Request().Context(), &b); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	c.Response().Header().Set("Location", "/fhir/Basic/"+b.FHIRID)
	return c.JSON(http.StatusCreated, b.ToFHIR())
}

func (h *Handler) UpdateBasicFHIR(c echo.Context) error {
	var b Basic
	if err := c.Bind(&b); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	existing, err := h.svc.GetBasicByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Basic", c.Param("id")))
	}
	b.ID = existing.ID
	b.FHIRID = existing.FHIRID
	if err := h.svc.UpdateBasic(c.Request().Context(), &b); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, b.ToFHIR())
}

func (h *Handler) DeleteBasicFHIR(c echo.Context) error {
	existing, err := h.svc.GetBasicByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Basic", c.Param("id")))
	}
	if err := h.svc.DeleteBasic(c.Request().Context(), existing.ID); err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) PatchBasicFHIR(c echo.Context) error {
	return h.handlePatch(c, c.Param("id"))
}

func (h *Handler) VreadBasicFHIR(c echo.Context) error {
	b, err := h.svc.GetBasicByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Basic", c.Param("id")))
	}
	result := b.ToFHIR()
	fhir.SetVersionHeaders(c, 1, b.UpdatedAt.Format("2006-01-02T15:04:05Z"))
	return c.JSON(http.StatusOK, result)
}

func (h *Handler) HistoryBasicFHIR(c echo.Context) error {
	b, err := h.svc.GetBasicByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Basic", c.Param("id")))
	}
	result := b.ToFHIR()
	raw, _ := json.Marshal(result)
	entry := &fhir.HistoryEntry{
		ResourceType: "Basic", ResourceID: b.FHIRID, VersionID: 1,
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
	existing, err := h.svc.GetBasicByFHIRID(c.Request().Context(), fhirID)
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Basic", fhirID))
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
	if code, ok := patched["code"].(map[string]interface{}); ok {
		if coding, ok := code["coding"].([]interface{}); ok && len(coding) > 0 {
			if c0, ok := coding[0].(map[string]interface{}); ok {
				if v, ok := c0["code"].(string); ok {
					existing.CodeCode = v
				}
			}
		}
	}
	if err := h.svc.UpdateBasic(c.Request().Context(), existing); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, existing.ToFHIR())
}
