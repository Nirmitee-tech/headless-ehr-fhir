package communicationrequest

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
	read.GET("/communication-requests", h.ListCommunicationRequests)
	read.GET("/communication-requests/:id", h.GetCommunicationRequest)

	write := api.Group("", role)
	write.POST("/communication-requests", h.CreateCommunicationRequest)
	write.PUT("/communication-requests/:id", h.UpdateCommunicationRequest)
	write.DELETE("/communication-requests/:id", h.DeleteCommunicationRequest)

	fhirRead := fhirGroup.Group("", role)
	fhirRead.GET("/CommunicationRequest", h.SearchCommunicationRequestsFHIR)
	fhirRead.GET("/CommunicationRequest/:id", h.GetCommunicationRequestFHIR)

	fhirWrite := fhirGroup.Group("", role)
	fhirWrite.POST("/CommunicationRequest", h.CreateCommunicationRequestFHIR)
	fhirWrite.PUT("/CommunicationRequest/:id", h.UpdateCommunicationRequestFHIR)
	fhirWrite.DELETE("/CommunicationRequest/:id", h.DeleteCommunicationRequestFHIR)
	fhirWrite.PATCH("/CommunicationRequest/:id", h.PatchCommunicationRequestFHIR)

	fhirRead.POST("/CommunicationRequest/_search", h.SearchCommunicationRequestsFHIR)
	fhirRead.GET("/CommunicationRequest/:id/_history/:vid", h.VreadCommunicationRequestFHIR)
	fhirRead.GET("/CommunicationRequest/:id/_history", h.HistoryCommunicationRequestFHIR)
}

// -- REST Endpoints --

func (h *Handler) CreateCommunicationRequest(c echo.Context) error {
	var cr CommunicationRequest
	if err := c.Bind(&cr); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateCommunicationRequest(c.Request().Context(), &cr); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, cr)
}

func (h *Handler) GetCommunicationRequest(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	cr, err := h.svc.GetCommunicationRequest(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "communication request not found")
	}
	return c.JSON(http.StatusOK, cr)
}

func (h *Handler) ListCommunicationRequests(c echo.Context) error {
	pg := pagination.FromContext(c)
	items, total, err := h.svc.SearchCommunicationRequests(c.Request().Context(), nil, pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *Handler) UpdateCommunicationRequest(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var cr CommunicationRequest
	if err := c.Bind(&cr); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	cr.ID = id
	if err := h.svc.UpdateCommunicationRequest(c.Request().Context(), &cr); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, cr)
}

func (h *Handler) DeleteCommunicationRequest(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeleteCommunicationRequest(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

// -- FHIR Endpoints --

func (h *Handler) SearchCommunicationRequestsFHIR(c echo.Context) error {
	pg := pagination.FromContext(c)
	params := fhir.ExtractSearchParams(c)
	items, total, err := h.svc.SearchCommunicationRequests(c.Request().Context(), params, pg.Limit, pg.Offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	resources := make([]interface{}, len(items))
	for i, item := range items {
		resources[i] = item.ToFHIR()
	}
	return c.JSON(http.StatusOK, fhir.NewSearchBundle(resources, total, "/fhir/CommunicationRequest"))
}

func (h *Handler) GetCommunicationRequestFHIR(c echo.Context) error {
	cr, err := h.svc.GetCommunicationRequestByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("CommunicationRequest", c.Param("id")))
	}
	fhir.SetVersionHeaders(c, 1, cr.UpdatedAt.Format("2006-01-02T15:04:05Z"))
	return c.JSON(http.StatusOK, cr.ToFHIR())
}

func (h *Handler) CreateCommunicationRequestFHIR(c echo.Context) error {
	var cr CommunicationRequest
	if err := c.Bind(&cr); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	if err := h.svc.CreateCommunicationRequest(c.Request().Context(), &cr); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	c.Response().Header().Set("Location", "/fhir/CommunicationRequest/"+cr.FHIRID)
	return c.JSON(http.StatusCreated, cr.ToFHIR())
}

func (h *Handler) UpdateCommunicationRequestFHIR(c echo.Context) error {
	var cr CommunicationRequest
	if err := c.Bind(&cr); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	existing, err := h.svc.GetCommunicationRequestByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("CommunicationRequest", c.Param("id")))
	}
	cr.ID = existing.ID
	cr.FHIRID = existing.FHIRID
	if err := h.svc.UpdateCommunicationRequest(c.Request().Context(), &cr); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, cr.ToFHIR())
}

func (h *Handler) DeleteCommunicationRequestFHIR(c echo.Context) error {
	existing, err := h.svc.GetCommunicationRequestByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("CommunicationRequest", c.Param("id")))
	}
	if err := h.svc.DeleteCommunicationRequest(c.Request().Context(), existing.ID); err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) PatchCommunicationRequestFHIR(c echo.Context) error {
	return h.handlePatch(c, c.Param("id"))
}

func (h *Handler) VreadCommunicationRequestFHIR(c echo.Context) error {
	cr, err := h.svc.GetCommunicationRequestByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("CommunicationRequest", c.Param("id")))
	}
	result := cr.ToFHIR()
	fhir.SetVersionHeaders(c, 1, cr.UpdatedAt.Format("2006-01-02T15:04:05Z"))
	return c.JSON(http.StatusOK, result)
}

func (h *Handler) HistoryCommunicationRequestFHIR(c echo.Context) error {
	cr, err := h.svc.GetCommunicationRequestByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("CommunicationRequest", c.Param("id")))
	}
	result := cr.ToFHIR()
	raw, _ := json.Marshal(result)
	entry := &fhir.HistoryEntry{
		ResourceType: "CommunicationRequest", ResourceID: cr.FHIRID, VersionID: 1,
		Resource: raw, Action: "create", Timestamp: cr.CreatedAt,
	}
	return c.JSON(http.StatusOK, fhir.NewHistoryBundle([]*fhir.HistoryEntry{entry}, 1, "/fhir"))
}

func (h *Handler) handlePatch(c echo.Context, fhirID string) error {
	contentType := c.Request().Header.Get("Content-Type")
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome("failed to read request body"))
	}
	existing, err := h.svc.GetCommunicationRequestByFHIRID(c.Request().Context(), fhirID)
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("CommunicationRequest", fhirID))
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
	if err := h.svc.UpdateCommunicationRequest(c.Request().Context(), existing); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, existing.ToFHIR())
}
