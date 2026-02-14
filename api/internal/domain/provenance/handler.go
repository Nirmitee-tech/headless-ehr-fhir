package provenance

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/ehr/ehr/internal/platform/auth"
	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/ehr/ehr/pkg/pagination"
)

// Handler provides HTTP handlers for the Provenance domain.
type Handler struct {
	svc *Service
}

// NewHandler creates a new Provenance domain handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// RegisterRoutes registers all Provenance domain routes.
func (h *Handler) RegisterRoutes(api *echo.Group, fhirGroup *echo.Group) {
	role := auth.RequireRole("admin", "physician", "nurse")

	read := api.Group("", role)
	read.GET("/provenances", h.ListProvenances)
	read.GET("/provenances/:id", h.GetProvenance)

	write := api.Group("", role)
	write.POST("/provenances", h.CreateProvenance)
	write.PUT("/provenances/:id", h.UpdateProvenance)
	write.DELETE("/provenances/:id", h.DeleteProvenance)

	fhirRead := fhirGroup.Group("", role)
	fhirRead.GET("/Provenance", h.SearchFHIR)
	fhirRead.GET("/Provenance/:id", h.GetFHIR)
	fhirRead.POST("/Provenance/_search", h.SearchFHIR)
	fhirRead.GET("/Provenance/:id/_history/:vid", h.VreadFHIR)
	fhirRead.GET("/Provenance/:id/_history", h.HistoryFHIR)

	fhirWrite := fhirGroup.Group("", role)
	fhirWrite.POST("/Provenance", h.CreateFHIR)
	fhirWrite.PUT("/Provenance/:id", h.UpdateFHIR)
	fhirWrite.DELETE("/Provenance/:id", h.DeleteFHIR)
}

// -- REST Handlers --

func (h *Handler) CreateProvenance(c echo.Context) error {
	var p Provenance
	if err := c.Bind(&p); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateProvenance(c.Request().Context(), &p); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, p)
}

func (h *Handler) GetProvenance(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	p, err := h.svc.GetProvenance(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "provenance not found")
	}
	return c.JSON(http.StatusOK, p)
}

func (h *Handler) ListProvenances(c echo.Context) error {
	pg := pagination.FromContext(c)
	items, total, err := h.svc.SearchProvenances(c.Request().Context(), nil, pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *Handler) UpdateProvenance(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var p Provenance
	if err := c.Bind(&p); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	p.ID = id
	if err := h.svc.UpdateProvenance(c.Request().Context(), &p); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, p)
}

func (h *Handler) DeleteProvenance(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeleteProvenance(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

// -- FHIR Handlers --

func (h *Handler) SearchFHIR(c echo.Context) error {
	pg := pagination.FromContext(c)
	params := map[string]string{}
	for _, k := range []string{"target", "agent"} {
		if v := c.QueryParam(k); v != "" {
			params[k] = v
		}
	}
	items, total, err := h.svc.SearchProvenances(c.Request().Context(), params, pg.Limit, pg.Offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	resources := make([]interface{}, len(items))
	for i, item := range items {
		resources[i] = item.ToFHIR()
	}
	return c.JSON(http.StatusOK, fhir.NewSearchBundle(resources, total, "/fhir/Provenance"))
}

func (h *Handler) GetFHIR(c echo.Context) error {
	p, err := h.svc.GetProvenanceByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Provenance", c.Param("id")))
	}
	return c.JSON(http.StatusOK, p.ToFHIR())
}

func (h *Handler) CreateFHIR(c echo.Context) error {
	var p Provenance
	if err := c.Bind(&p); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	if err := h.svc.CreateProvenance(c.Request().Context(), &p); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	c.Response().Header().Set("Location", "/fhir/Provenance/"+p.FHIRID)
	return c.JSON(http.StatusCreated, p.ToFHIR())
}

func (h *Handler) UpdateFHIR(c echo.Context) error {
	var p Provenance
	if err := c.Bind(&p); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	existing, err := h.svc.GetProvenanceByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Provenance", c.Param("id")))
	}
	p.ID = existing.ID
	p.FHIRID = existing.FHIRID
	if err := h.svc.UpdateProvenance(c.Request().Context(), &p); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, p.ToFHIR())
}

func (h *Handler) DeleteFHIR(c echo.Context) error {
	existing, err := h.svc.GetProvenanceByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Provenance", c.Param("id")))
	}
	if err := h.svc.DeleteProvenance(c.Request().Context(), existing.ID); err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) VreadFHIR(c echo.Context) error {
	p, err := h.svc.GetProvenanceByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Provenance", c.Param("id")))
	}
	result := p.ToFHIR()
	fhir.SetVersionHeaders(c, 1, p.UpdatedAt.Format("2006-01-02T15:04:05Z"))
	return c.JSON(http.StatusOK, result)
}

func (h *Handler) HistoryFHIR(c echo.Context) error {
	p, err := h.svc.GetProvenanceByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Provenance", c.Param("id")))
	}
	result := p.ToFHIR()
	raw, _ := json.Marshal(result)
	entry := &fhir.HistoryEntry{
		ResourceType: "Provenance", ResourceID: p.FHIRID, VersionID: 1,
		Resource: raw, Action: "create", Timestamp: p.CreatedAt,
	}
	return c.JSON(http.StatusOK, fhir.NewHistoryBundle([]*fhir.HistoryEntry{entry}, 1, "/fhir"))
}
