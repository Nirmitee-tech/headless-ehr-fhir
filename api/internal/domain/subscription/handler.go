package subscription

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

// Handler provides HTTP endpoints for Subscription management.
type Handler struct {
	svc *Service
}

// NewHandler creates a new subscription handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// RegisterRoutes registers REST and FHIR endpoints.
func (h *Handler) RegisterRoutes(api *echo.Group, fhirGroup *echo.Group) {
	role := auth.RequireRole("admin")

	read := api.Group("", role)
	read.GET("/subscriptions", h.ListSubscriptions)
	read.GET("/subscriptions/:id", h.GetSubscription)
	read.GET("/subscriptions/:id/notifications", h.ListNotifications)

	write := api.Group("", role)
	write.POST("/subscriptions", h.CreateSubscription)
	write.PUT("/subscriptions/:id", h.UpdateSubscription)
	write.DELETE("/subscriptions/:id", h.DeleteSubscription)

	fhirRead := fhirGroup.Group("", role)
	fhirRead.GET("/Subscription", h.SearchSubscriptionsFHIR)
	fhirRead.GET("/Subscription/:id", h.GetSubscriptionFHIR)

	fhirWrite := fhirGroup.Group("", role)
	fhirWrite.POST("/Subscription", h.CreateSubscriptionFHIR)
	fhirWrite.PUT("/Subscription/:id", h.UpdateSubscriptionFHIR)
	fhirWrite.DELETE("/Subscription/:id", h.DeleteSubscriptionFHIR)
	fhirWrite.PATCH("/Subscription/:id", h.PatchSubscriptionFHIR)

	fhirRead.POST("/Subscription/_search", h.SearchSubscriptionsFHIR)

	fhirRead.GET("/Subscription/:id/_history/:vid", h.VreadSubscriptionFHIR)
	fhirRead.GET("/Subscription/:id/_history", h.HistorySubscriptionFHIR)
}

// -- REST handlers --

func (h *Handler) CreateSubscription(c echo.Context) error {
	var sub Subscription
	if err := c.Bind(&sub); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateSubscription(c.Request().Context(), &sub); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, sub)
}

func (h *Handler) GetSubscription(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	sub, err := h.svc.GetSubscription(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "subscription not found")
	}
	return c.JSON(http.StatusOK, sub)
}

func (h *Handler) ListSubscriptions(c echo.Context) error {
	pg := pagination.FromContext(c)
	params := fhir.ExtractSearchParams(c)
	items, total, err := h.svc.SearchSubscriptions(c.Request().Context(), params, pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *Handler) UpdateSubscription(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var sub Subscription
	if err := c.Bind(&sub); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	sub.ID = id
	if err := h.svc.UpdateSubscription(c.Request().Context(), &sub); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, sub)
}

func (h *Handler) DeleteSubscription(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeleteSubscription(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) ListNotifications(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	pg := pagination.FromContext(c)
	items, total, err := h.svc.ListNotificationsBySubscription(c.Request().Context(), id, pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

// -- FHIR handlers --

func (h *Handler) SearchSubscriptionsFHIR(c echo.Context) error {
	pg := pagination.FromContext(c)
	params := fhir.ExtractSearchParams(c)
	items, total, err := h.svc.SearchSubscriptions(c.Request().Context(), params, pg.Limit, pg.Offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	resources := make([]interface{}, len(items))
	for i, item := range items {
		resources[i] = item.ToFHIR()
	}
	return c.JSON(http.StatusOK, fhir.NewSearchBundle(resources, total, "/fhir/Subscription"))
}

func (h *Handler) GetSubscriptionFHIR(c echo.Context) error {
	sub, err := h.svc.GetSubscriptionByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Subscription", c.Param("id")))
	}
	return c.JSON(http.StatusOK, sub.ToFHIR())
}

func (h *Handler) CreateSubscriptionFHIR(c echo.Context) error {
	var sub Subscription
	if err := c.Bind(&sub); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	if err := h.svc.CreateSubscription(c.Request().Context(), &sub); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	c.Response().Header().Set("Location", "/fhir/Subscription/"+sub.FHIRID)
	return c.JSON(http.StatusCreated, sub.ToFHIR())
}

func (h *Handler) UpdateSubscriptionFHIR(c echo.Context) error {
	var sub Subscription
	if err := c.Bind(&sub); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	existing, err := h.svc.GetSubscriptionByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Subscription", c.Param("id")))
	}
	sub.ID = existing.ID
	sub.FHIRID = existing.FHIRID
	sub.VersionID = existing.VersionID
	if err := h.svc.UpdateSubscription(c.Request().Context(), &sub); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, sub.ToFHIR())
}

func (h *Handler) DeleteSubscriptionFHIR(c echo.Context) error {
	existing, err := h.svc.GetSubscriptionByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Subscription", c.Param("id")))
	}
	if err := h.svc.DeleteSubscription(c.Request().Context(), existing.ID); err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) PatchSubscriptionFHIR(c echo.Context) error {
	contentType := c.Request().Header.Get("Content-Type")
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome("failed to read request body"))
	}

	existing, err := h.svc.GetSubscriptionByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Subscription", c.Param("id")))
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
	if err := h.svc.UpdateSubscription(c.Request().Context(), existing); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, existing.ToFHIR())
}

func (h *Handler) VreadSubscriptionFHIR(c echo.Context) error {
	sub, err := h.svc.GetSubscriptionByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Subscription", c.Param("id")))
	}
	result := sub.ToFHIR()
	fhir.SetVersionHeaders(c, 1, sub.UpdatedAt.Format("2006-01-02T15:04:05Z"))
	return c.JSON(http.StatusOK, result)
}

func (h *Handler) HistorySubscriptionFHIR(c echo.Context) error {
	sub, err := h.svc.GetSubscriptionByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Subscription", c.Param("id")))
	}
	result := sub.ToFHIR()
	raw, _ := json.Marshal(result)
	entry := &fhir.HistoryEntry{
		ResourceType: "Subscription", ResourceID: sub.FHIRID, VersionID: 1,
		Resource: raw, Action: "create", Timestamp: sub.CreatedAt,
	}
	return c.JSON(http.StatusOK, fhir.NewHistoryBundle([]*fhir.HistoryEntry{entry}, 1, "/fhir"))
}
