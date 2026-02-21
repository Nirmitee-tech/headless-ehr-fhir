package coverageeligibility

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

	// -- REST: CoverageEligibilityRequest --
	read := api.Group("", role)
	read.GET("/coverage-eligibility-requests", h.ListRequests)
	read.GET("/coverage-eligibility-requests/:id", h.GetRequest)

	write := api.Group("", role)
	write.POST("/coverage-eligibility-requests", h.CreateRequest)
	write.PUT("/coverage-eligibility-requests/:id", h.UpdateRequest)
	write.DELETE("/coverage-eligibility-requests/:id", h.DeleteRequest)

	// -- REST: CoverageEligibilityResponse --
	read.GET("/coverage-eligibility-responses", h.ListResponses)
	read.GET("/coverage-eligibility-responses/:id", h.GetResponse)

	write.POST("/coverage-eligibility-responses", h.CreateResponse)
	write.PUT("/coverage-eligibility-responses/:id", h.UpdateResponse)
	write.DELETE("/coverage-eligibility-responses/:id", h.DeleteResponse)

	// -- FHIR: CoverageEligibilityRequest --
	fhirRead := fhirGroup.Group("", role)
	fhirRead.GET("/CoverageEligibilityRequest", h.SearchRequestsFHIR)
	fhirRead.GET("/CoverageEligibilityRequest/:id", h.GetRequestFHIR)

	fhirWrite := fhirGroup.Group("", role)
	fhirWrite.POST("/CoverageEligibilityRequest", h.CreateRequestFHIR)
	fhirWrite.PUT("/CoverageEligibilityRequest/:id", h.UpdateRequestFHIR)
	fhirWrite.DELETE("/CoverageEligibilityRequest/:id", h.DeleteRequestFHIR)
	fhirWrite.PATCH("/CoverageEligibilityRequest/:id", h.PatchRequestFHIR)

	fhirRead.POST("/CoverageEligibilityRequest/_search", h.SearchRequestsFHIR)
	fhirRead.GET("/CoverageEligibilityRequest/:id/_history/:vid", h.VreadRequestFHIR)
	fhirRead.GET("/CoverageEligibilityRequest/:id/_history", h.HistoryRequestFHIR)

	// -- FHIR: CoverageEligibilityResponse --
	fhirRead.GET("/CoverageEligibilityResponse", h.SearchResponsesFHIR)
	fhirRead.GET("/CoverageEligibilityResponse/:id", h.GetResponseFHIR)

	fhirWrite.POST("/CoverageEligibilityResponse", h.CreateResponseFHIR)
	fhirWrite.PUT("/CoverageEligibilityResponse/:id", h.UpdateResponseFHIR)
	fhirWrite.DELETE("/CoverageEligibilityResponse/:id", h.DeleteResponseFHIR)
	fhirWrite.PATCH("/CoverageEligibilityResponse/:id", h.PatchResponseFHIR)

	fhirRead.POST("/CoverageEligibilityResponse/_search", h.SearchResponsesFHIR)
	fhirRead.GET("/CoverageEligibilityResponse/:id/_history/:vid", h.VreadResponseFHIR)
	fhirRead.GET("/CoverageEligibilityResponse/:id/_history", h.HistoryResponseFHIR)
}

// ============================================================
// REST: CoverageEligibilityRequest
// ============================================================

func (h *Handler) CreateRequest(c echo.Context) error {
	var r CoverageEligibilityRequest
	if err := c.Bind(&r); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateRequest(c.Request().Context(), &r); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, r)
}

func (h *Handler) GetRequest(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	r, err := h.svc.GetRequest(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "coverage eligibility request not found")
	}
	return c.JSON(http.StatusOK, r)
}

func (h *Handler) ListRequests(c echo.Context) error {
	pg := pagination.FromContext(c)
	items, total, err := h.svc.SearchRequests(c.Request().Context(), nil, pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *Handler) UpdateRequest(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var r CoverageEligibilityRequest
	if err := c.Bind(&r); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	r.ID = id
	if err := h.svc.UpdateRequest(c.Request().Context(), &r); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, r)
}

func (h *Handler) DeleteRequest(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeleteRequest(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

// ============================================================
// REST: CoverageEligibilityResponse
// ============================================================

func (h *Handler) CreateResponse(c echo.Context) error {
	var r CoverageEligibilityResponse
	if err := c.Bind(&r); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateResponse(c.Request().Context(), &r); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, r)
}

func (h *Handler) GetResponse(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	r, err := h.svc.GetResponse(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "coverage eligibility response not found")
	}
	return c.JSON(http.StatusOK, r)
}

func (h *Handler) ListResponses(c echo.Context) error {
	pg := pagination.FromContext(c)
	items, total, err := h.svc.SearchResponses(c.Request().Context(), nil, pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *Handler) UpdateResponse(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var r CoverageEligibilityResponse
	if err := c.Bind(&r); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	r.ID = id
	if err := h.svc.UpdateResponse(c.Request().Context(), &r); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, r)
}

func (h *Handler) DeleteResponse(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeleteResponse(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

// ============================================================
// FHIR: CoverageEligibilityRequest
// ============================================================

func (h *Handler) SearchRequestsFHIR(c echo.Context) error {
	pg := pagination.FromContext(c)
	params := fhir.ExtractSearchParams(c)
	items, total, err := h.svc.SearchRequests(c.Request().Context(), params, pg.Limit, pg.Offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	resources := make([]interface{}, len(items))
	for i, item := range items {
		resources[i] = item.ToFHIR()
	}
	return c.JSON(http.StatusOK, fhir.NewSearchBundle(resources, total, "/fhir/CoverageEligibilityRequest"))
}

func (h *Handler) GetRequestFHIR(c echo.Context) error {
	r, err := h.svc.GetRequestByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("CoverageEligibilityRequest", c.Param("id")))
	}
	fhir.SetVersionHeaders(c, 1, r.UpdatedAt.Format("2006-01-02T15:04:05Z"))
	return c.JSON(http.StatusOK, r.ToFHIR())
}

func (h *Handler) CreateRequestFHIR(c echo.Context) error {
	var r CoverageEligibilityRequest
	if err := c.Bind(&r); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	if err := h.svc.CreateRequest(c.Request().Context(), &r); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	c.Response().Header().Set("Location", "/fhir/CoverageEligibilityRequest/"+r.FHIRID)
	return c.JSON(http.StatusCreated, r.ToFHIR())
}

func (h *Handler) UpdateRequestFHIR(c echo.Context) error {
	var r CoverageEligibilityRequest
	if err := c.Bind(&r); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	existing, err := h.svc.GetRequestByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("CoverageEligibilityRequest", c.Param("id")))
	}
	r.ID = existing.ID
	r.FHIRID = existing.FHIRID
	if err := h.svc.UpdateRequest(c.Request().Context(), &r); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, r.ToFHIR())
}

func (h *Handler) DeleteRequestFHIR(c echo.Context) error {
	existing, err := h.svc.GetRequestByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("CoverageEligibilityRequest", c.Param("id")))
	}
	if err := h.svc.DeleteRequest(c.Request().Context(), existing.ID); err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) PatchRequestFHIR(c echo.Context) error {
	return h.handlePatchRequest(c, c.Param("id"))
}

func (h *Handler) VreadRequestFHIR(c echo.Context) error {
	r, err := h.svc.GetRequestByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("CoverageEligibilityRequest", c.Param("id")))
	}
	result := r.ToFHIR()
	fhir.SetVersionHeaders(c, 1, r.UpdatedAt.Format("2006-01-02T15:04:05Z"))
	return c.JSON(http.StatusOK, result)
}

func (h *Handler) HistoryRequestFHIR(c echo.Context) error {
	r, err := h.svc.GetRequestByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("CoverageEligibilityRequest", c.Param("id")))
	}
	result := r.ToFHIR()
	raw, _ := json.Marshal(result)
	entry := &fhir.HistoryEntry{
		ResourceType: "CoverageEligibilityRequest", ResourceID: r.FHIRID, VersionID: 1,
		Resource: raw, Action: "create", Timestamp: r.CreatedAt,
	}
	return c.JSON(http.StatusOK, fhir.NewHistoryBundle([]*fhir.HistoryEntry{entry}, 1, "/fhir"))
}

func (h *Handler) handlePatchRequest(c echo.Context, fhirID string) error {
	contentType := c.Request().Header.Get("Content-Type")
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome("failed to read request body"))
	}
	existing, err := h.svc.GetRequestByFHIRID(c.Request().Context(), fhirID)
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("CoverageEligibilityRequest", fhirID))
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
	if err := h.svc.UpdateRequest(c.Request().Context(), existing); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, existing.ToFHIR())
}

// ============================================================
// FHIR: CoverageEligibilityResponse
// ============================================================

func (h *Handler) SearchResponsesFHIR(c echo.Context) error {
	pg := pagination.FromContext(c)
	params := fhir.ExtractSearchParams(c)
	items, total, err := h.svc.SearchResponses(c.Request().Context(), params, pg.Limit, pg.Offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	resources := make([]interface{}, len(items))
	for i, item := range items {
		resources[i] = item.ToFHIR()
	}
	return c.JSON(http.StatusOK, fhir.NewSearchBundle(resources, total, "/fhir/CoverageEligibilityResponse"))
}

func (h *Handler) GetResponseFHIR(c echo.Context) error {
	r, err := h.svc.GetResponseByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("CoverageEligibilityResponse", c.Param("id")))
	}
	fhir.SetVersionHeaders(c, 1, r.UpdatedAt.Format("2006-01-02T15:04:05Z"))
	return c.JSON(http.StatusOK, r.ToFHIR())
}

func (h *Handler) CreateResponseFHIR(c echo.Context) error {
	var r CoverageEligibilityResponse
	if err := c.Bind(&r); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	if err := h.svc.CreateResponse(c.Request().Context(), &r); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	c.Response().Header().Set("Location", "/fhir/CoverageEligibilityResponse/"+r.FHIRID)
	return c.JSON(http.StatusCreated, r.ToFHIR())
}

func (h *Handler) UpdateResponseFHIR(c echo.Context) error {
	var r CoverageEligibilityResponse
	if err := c.Bind(&r); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	existing, err := h.svc.GetResponseByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("CoverageEligibilityResponse", c.Param("id")))
	}
	r.ID = existing.ID
	r.FHIRID = existing.FHIRID
	if err := h.svc.UpdateResponse(c.Request().Context(), &r); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, r.ToFHIR())
}

func (h *Handler) DeleteResponseFHIR(c echo.Context) error {
	existing, err := h.svc.GetResponseByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("CoverageEligibilityResponse", c.Param("id")))
	}
	if err := h.svc.DeleteResponse(c.Request().Context(), existing.ID); err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) PatchResponseFHIR(c echo.Context) error {
	return h.handlePatchResponse(c, c.Param("id"))
}

func (h *Handler) VreadResponseFHIR(c echo.Context) error {
	r, err := h.svc.GetResponseByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("CoverageEligibilityResponse", c.Param("id")))
	}
	result := r.ToFHIR()
	fhir.SetVersionHeaders(c, 1, r.UpdatedAt.Format("2006-01-02T15:04:05Z"))
	return c.JSON(http.StatusOK, result)
}

func (h *Handler) HistoryResponseFHIR(c echo.Context) error {
	r, err := h.svc.GetResponseByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("CoverageEligibilityResponse", c.Param("id")))
	}
	result := r.ToFHIR()
	raw, _ := json.Marshal(result)
	entry := &fhir.HistoryEntry{
		ResourceType: "CoverageEligibilityResponse", ResourceID: r.FHIRID, VersionID: 1,
		Resource: raw, Action: "create", Timestamp: r.CreatedAt,
	}
	return c.JSON(http.StatusOK, fhir.NewHistoryBundle([]*fhir.HistoryEntry{entry}, 1, "/fhir"))
}

func (h *Handler) handlePatchResponse(c echo.Context, fhirID string) error {
	contentType := c.Request().Header.Get("Content-Type")
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome("failed to read request body"))
	}
	existing, err := h.svc.GetResponseByFHIRID(c.Request().Context(), fhirID)
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("CoverageEligibilityResponse", fhirID))
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
	if v, ok := patched["outcome"].(string); ok {
		existing.Outcome = v
	}
	if err := h.svc.UpdateResponse(c.Request().Context(), existing); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, existing.ToFHIR())
}
