package auditevent

import (
	"encoding/json"
	"net/http"

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
	role := auth.RequireRole("admin")

	read := api.Group("", role)
	read.GET("/audit-events", h.ListAuditEvents)
	read.GET("/audit-events/:id", h.GetAuditEvent)

	fhirRead := fhirGroup.Group("", role)
	fhirRead.GET("/AuditEvent", h.SearchAuditEventsFHIR)
	fhirRead.GET("/AuditEvent/:id", h.GetAuditEventFHIR)
	fhirRead.POST("/AuditEvent/_search", h.SearchAuditEventsFHIR)
	fhirRead.GET("/AuditEvent/:id/_history/:vid", h.VreadAuditEventFHIR)
	fhirRead.GET("/AuditEvent/:id/_history", h.HistoryAuditEventFHIR)
}

func (h *Handler) GetAuditEvent(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	a, err := h.svc.GetAuditEvent(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "audit event not found")
	}
	return c.JSON(http.StatusOK, a)
}

func (h *Handler) ListAuditEvents(c echo.Context) error {
	pg := pagination.FromContext(c)
	items, total, err := h.svc.SearchAuditEvents(c.Request().Context(), nil, pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *Handler) SearchAuditEventsFHIR(c echo.Context) error {
	pg := pagination.FromContext(c)
	params := fhir.ExtractSearchParams(c)
	items, total, err := h.svc.SearchAuditEvents(c.Request().Context(), params, pg.Limit, pg.Offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	resources := make([]interface{}, len(items))
	for i, item := range items {
		resources[i] = item.ToFHIR()
	}
	return c.JSON(http.StatusOK, fhir.NewSearchBundle(resources, total, "/fhir/AuditEvent"))
}

func (h *Handler) GetAuditEventFHIR(c echo.Context) error {
	a, err := h.svc.GetAuditEventByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("AuditEvent", c.Param("id")))
	}
	return c.JSON(http.StatusOK, a.ToFHIR())
}

func (h *Handler) VreadAuditEventFHIR(c echo.Context) error {
	a, err := h.svc.GetAuditEventByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("AuditEvent", c.Param("id")))
	}
	result := a.ToFHIR()
	fhir.SetVersionHeaders(c, 1, a.Recorded.Format("2006-01-02T15:04:05Z"))
	return c.JSON(http.StatusOK, result)
}

func (h *Handler) HistoryAuditEventFHIR(c echo.Context) error {
	a, err := h.svc.GetAuditEventByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("AuditEvent", c.Param("id")))
	}
	result := a.ToFHIR()
	raw, _ := json.Marshal(result)
	entry := &fhir.HistoryEntry{
		ResourceType: "AuditEvent", ResourceID: a.FHIRID, VersionID: 1,
		Resource: raw, Action: "create", Timestamp: a.CreatedAt,
	}
	return c.JSON(http.StatusOK, fhir.NewHistoryBundle([]*fhir.HistoryEntry{entry}, 1, "/fhir"))
}
