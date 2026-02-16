package supply

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
	read.GET("/supply-requests", h.ListSupplyRequests)
	read.GET("/supply-requests/:id", h.GetSupplyRequest)
	read.GET("/supply-deliveries", h.ListSupplyDeliveries)
	read.GET("/supply-deliveries/:id", h.GetSupplyDelivery)

	write := api.Group("", role)
	write.POST("/supply-requests", h.CreateSupplyRequest)
	write.PUT("/supply-requests/:id", h.UpdateSupplyRequest)
	write.DELETE("/supply-requests/:id", h.DeleteSupplyRequest)
	write.POST("/supply-deliveries", h.CreateSupplyDelivery)
	write.PUT("/supply-deliveries/:id", h.UpdateSupplyDelivery)
	write.DELETE("/supply-deliveries/:id", h.DeleteSupplyDelivery)

	fhirRead := fhirGroup.Group("", role)
	fhirRead.GET("/SupplyRequest", h.SearchSupplyRequestsFHIR)
	fhirRead.GET("/SupplyRequest/:id", h.GetSupplyRequestFHIR)
	fhirRead.POST("/SupplyRequest/_search", h.SearchSupplyRequestsFHIR)
	fhirRead.GET("/SupplyRequest/:id/_history/:vid", h.VreadSupplyRequestFHIR)
	fhirRead.GET("/SupplyRequest/:id/_history", h.HistorySupplyRequestFHIR)

	fhirWrite := fhirGroup.Group("", role)
	fhirWrite.POST("/SupplyRequest", h.CreateSupplyRequestFHIR)
	fhirWrite.PUT("/SupplyRequest/:id", h.UpdateSupplyRequestFHIR)
	fhirWrite.DELETE("/SupplyRequest/:id", h.DeleteSupplyRequestFHIR)
	fhirWrite.PATCH("/SupplyRequest/:id", h.PatchSupplyRequestFHIR)

	fhirRead.GET("/SupplyDelivery", h.SearchSupplyDeliveriesFHIR)
	fhirRead.GET("/SupplyDelivery/:id", h.GetSupplyDeliveryFHIR)
	fhirRead.POST("/SupplyDelivery/_search", h.SearchSupplyDeliveriesFHIR)
	fhirRead.GET("/SupplyDelivery/:id/_history/:vid", h.VreadSupplyDeliveryFHIR)
	fhirRead.GET("/SupplyDelivery/:id/_history", h.HistorySupplyDeliveryFHIR)

	fhirWrite.POST("/SupplyDelivery", h.CreateSupplyDeliveryFHIR)
	fhirWrite.PUT("/SupplyDelivery/:id", h.UpdateSupplyDeliveryFHIR)
	fhirWrite.DELETE("/SupplyDelivery/:id", h.DeleteSupplyDeliveryFHIR)
	fhirWrite.PATCH("/SupplyDelivery/:id", h.PatchSupplyDeliveryFHIR)
}

// ---- SupplyRequest REST ----

func (h *Handler) CreateSupplyRequest(c echo.Context) error {
	var sr SupplyRequest
	if err := c.Bind(&sr); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateSupplyRequest(c.Request().Context(), &sr); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, sr)
}

func (h *Handler) GetSupplyRequest(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	sr, err := h.svc.GetSupplyRequest(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "supply request not found")
	}
	return c.JSON(http.StatusOK, sr)
}

func (h *Handler) ListSupplyRequests(c echo.Context) error {
	pg := pagination.FromContext(c)
	params := fhir.ExtractSearchParams(c)
	items, total, err := h.svc.SearchSupplyRequests(c.Request().Context(), params, pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *Handler) UpdateSupplyRequest(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var sr SupplyRequest
	if err := c.Bind(&sr); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	sr.ID = id
	if err := h.svc.UpdateSupplyRequest(c.Request().Context(), &sr); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, sr)
}

func (h *Handler) DeleteSupplyRequest(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeleteSupplyRequest(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

// ---- SupplyDelivery REST ----

func (h *Handler) CreateSupplyDelivery(c echo.Context) error {
	var sd SupplyDelivery
	if err := c.Bind(&sd); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateSupplyDelivery(c.Request().Context(), &sd); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, sd)
}

func (h *Handler) GetSupplyDelivery(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	sd, err := h.svc.GetSupplyDelivery(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "supply delivery not found")
	}
	return c.JSON(http.StatusOK, sd)
}

func (h *Handler) ListSupplyDeliveries(c echo.Context) error {
	pg := pagination.FromContext(c)
	params := fhir.ExtractSearchParams(c)
	items, total, err := h.svc.SearchSupplyDeliveries(c.Request().Context(), params, pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *Handler) UpdateSupplyDelivery(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var sd SupplyDelivery
	if err := c.Bind(&sd); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	sd.ID = id
	if err := h.svc.UpdateSupplyDelivery(c.Request().Context(), &sd); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, sd)
}

func (h *Handler) DeleteSupplyDelivery(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeleteSupplyDelivery(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

// ---- FHIR SupplyRequest ----

func (h *Handler) SearchSupplyRequestsFHIR(c echo.Context) error {
	pg := pagination.FromContext(c)
	params := fhir.ExtractSearchParams(c)
	items, total, err := h.svc.SearchSupplyRequests(c.Request().Context(), params, pg.Limit, pg.Offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	resources := make([]interface{}, len(items))
	for i, item := range items {
		resources[i] = item.ToFHIR()
	}
	return c.JSON(http.StatusOK, fhir.NewSearchBundle(resources, total, "/fhir/SupplyRequest"))
}

func (h *Handler) GetSupplyRequestFHIR(c echo.Context) error {
	sr, err := h.svc.GetSupplyRequestByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("SupplyRequest", c.Param("id")))
	}
	return c.JSON(http.StatusOK, sr.ToFHIR())
}

func (h *Handler) CreateSupplyRequestFHIR(c echo.Context) error {
	var sr SupplyRequest
	if err := c.Bind(&sr); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	if err := h.svc.CreateSupplyRequest(c.Request().Context(), &sr); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	c.Response().Header().Set("Location", "/fhir/SupplyRequest/"+sr.FHIRID)
	return c.JSON(http.StatusCreated, sr.ToFHIR())
}

func (h *Handler) UpdateSupplyRequestFHIR(c echo.Context) error {
	var sr SupplyRequest
	if err := c.Bind(&sr); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	existing, err := h.svc.GetSupplyRequestByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("SupplyRequest", c.Param("id")))
	}
	sr.ID = existing.ID
	sr.FHIRID = existing.FHIRID
	if err := h.svc.UpdateSupplyRequest(c.Request().Context(), &sr); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, sr.ToFHIR())
}

func (h *Handler) DeleteSupplyRequestFHIR(c echo.Context) error {
	existing, err := h.svc.GetSupplyRequestByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("SupplyRequest", c.Param("id")))
	}
	if err := h.svc.DeleteSupplyRequest(c.Request().Context(), existing.ID); err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) PatchSupplyRequestFHIR(c echo.Context) error {
	return h.handlePatchSupplyRequest(c, c.Param("id"))
}

func (h *Handler) VreadSupplyRequestFHIR(c echo.Context) error {
	sr, err := h.svc.GetSupplyRequestByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("SupplyRequest", c.Param("id")))
	}
	result := sr.ToFHIR()
	fhir.SetVersionHeaders(c, 1, sr.UpdatedAt.Format("2006-01-02T15:04:05Z"))
	return c.JSON(http.StatusOK, result)
}

func (h *Handler) HistorySupplyRequestFHIR(c echo.Context) error {
	sr, err := h.svc.GetSupplyRequestByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("SupplyRequest", c.Param("id")))
	}
	result := sr.ToFHIR()
	raw, _ := json.Marshal(result)
	entry := &fhir.HistoryEntry{
		ResourceType: "SupplyRequest", ResourceID: sr.FHIRID, VersionID: 1,
		Resource: raw, Action: "create", Timestamp: sr.CreatedAt,
	}
	return c.JSON(http.StatusOK, fhir.NewHistoryBundle([]*fhir.HistoryEntry{entry}, 1, "/fhir"))
}

func (h *Handler) handlePatchSupplyRequest(c echo.Context, fhirID string) error {
	contentType := c.Request().Header.Get("Content-Type")
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome("failed to read request body"))
	}

	existing, err := h.svc.GetSupplyRequestByFHIRID(c.Request().Context(), fhirID)
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("SupplyRequest", fhirID))
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
	if err := h.svc.UpdateSupplyRequest(c.Request().Context(), existing); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, existing.ToFHIR())
}

// ---- FHIR SupplyDelivery ----

func (h *Handler) SearchSupplyDeliveriesFHIR(c echo.Context) error {
	pg := pagination.FromContext(c)
	params := fhir.ExtractSearchParams(c)
	items, total, err := h.svc.SearchSupplyDeliveries(c.Request().Context(), params, pg.Limit, pg.Offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	resources := make([]interface{}, len(items))
	for i, item := range items {
		resources[i] = item.ToFHIR()
	}
	return c.JSON(http.StatusOK, fhir.NewSearchBundle(resources, total, "/fhir/SupplyDelivery"))
}

func (h *Handler) GetSupplyDeliveryFHIR(c echo.Context) error {
	sd, err := h.svc.GetSupplyDeliveryByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("SupplyDelivery", c.Param("id")))
	}
	return c.JSON(http.StatusOK, sd.ToFHIR())
}

func (h *Handler) CreateSupplyDeliveryFHIR(c echo.Context) error {
	var sd SupplyDelivery
	if err := c.Bind(&sd); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	if err := h.svc.CreateSupplyDelivery(c.Request().Context(), &sd); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	c.Response().Header().Set("Location", "/fhir/SupplyDelivery/"+sd.FHIRID)
	return c.JSON(http.StatusCreated, sd.ToFHIR())
}

func (h *Handler) UpdateSupplyDeliveryFHIR(c echo.Context) error {
	var sd SupplyDelivery
	if err := c.Bind(&sd); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	existing, err := h.svc.GetSupplyDeliveryByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("SupplyDelivery", c.Param("id")))
	}
	sd.ID = existing.ID
	sd.FHIRID = existing.FHIRID
	if err := h.svc.UpdateSupplyDelivery(c.Request().Context(), &sd); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, sd.ToFHIR())
}

func (h *Handler) DeleteSupplyDeliveryFHIR(c echo.Context) error {
	existing, err := h.svc.GetSupplyDeliveryByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("SupplyDelivery", c.Param("id")))
	}
	if err := h.svc.DeleteSupplyDelivery(c.Request().Context(), existing.ID); err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) PatchSupplyDeliveryFHIR(c echo.Context) error {
	return h.handlePatchSupplyDelivery(c, c.Param("id"))
}

func (h *Handler) VreadSupplyDeliveryFHIR(c echo.Context) error {
	sd, err := h.svc.GetSupplyDeliveryByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("SupplyDelivery", c.Param("id")))
	}
	result := sd.ToFHIR()
	fhir.SetVersionHeaders(c, 1, sd.UpdatedAt.Format("2006-01-02T15:04:05Z"))
	return c.JSON(http.StatusOK, result)
}

func (h *Handler) HistorySupplyDeliveryFHIR(c echo.Context) error {
	sd, err := h.svc.GetSupplyDeliveryByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("SupplyDelivery", c.Param("id")))
	}
	result := sd.ToFHIR()
	raw, _ := json.Marshal(result)
	entry := &fhir.HistoryEntry{
		ResourceType: "SupplyDelivery", ResourceID: sd.FHIRID, VersionID: 1,
		Resource: raw, Action: "create", Timestamp: sd.CreatedAt,
	}
	return c.JSON(http.StatusOK, fhir.NewHistoryBundle([]*fhir.HistoryEntry{entry}, 1, "/fhir"))
}

func (h *Handler) handlePatchSupplyDelivery(c echo.Context, fhirID string) error {
	contentType := c.Request().Header.Get("Content-Type")
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome("failed to read request body"))
	}

	existing, err := h.svc.GetSupplyDeliveryByFHIRID(c.Request().Context(), fhirID)
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("SupplyDelivery", fhirID))
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
	if err := h.svc.UpdateSupplyDelivery(c.Request().Context(), existing); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, existing.ToFHIR())
}
