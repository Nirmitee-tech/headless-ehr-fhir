package relatedperson

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/ehr/ehr/internal/platform/auth"
	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/ehr/ehr/pkg/pagination"
)

// Handler provides HTTP handlers for the RelatedPerson domain.
type Handler struct {
	svc *Service
}

// NewHandler creates a new RelatedPerson domain handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// RegisterRoutes registers all RelatedPerson domain routes.
func (h *Handler) RegisterRoutes(api *echo.Group, fhirGroup *echo.Group) {
	role := auth.RequireRole("admin", "physician", "nurse")

	read := api.Group("", role)
	read.GET("/related-persons", h.ListRelatedPersons)
	read.GET("/related-persons/:id", h.GetRelatedPerson)

	write := api.Group("", role)
	write.POST("/related-persons", h.CreateRelatedPerson)
	write.PUT("/related-persons/:id", h.UpdateRelatedPerson)
	write.DELETE("/related-persons/:id", h.DeleteRelatedPerson)

	fhirRead := fhirGroup.Group("", role)
	fhirRead.GET("/RelatedPerson", h.SearchFHIR)
	fhirRead.GET("/RelatedPerson/:id", h.GetFHIR)
	fhirRead.POST("/RelatedPerson/_search", h.SearchFHIR)
	fhirRead.GET("/RelatedPerson/:id/_history/:vid", h.VreadFHIR)
	fhirRead.GET("/RelatedPerson/:id/_history", h.HistoryFHIR)

	fhirWrite := fhirGroup.Group("", role)
	fhirWrite.POST("/RelatedPerson", h.CreateFHIR)
	fhirWrite.PUT("/RelatedPerson/:id", h.UpdateFHIR)
	fhirWrite.DELETE("/RelatedPerson/:id", h.DeleteFHIR)
}

// -- REST Handlers --

func (h *Handler) CreateRelatedPerson(c echo.Context) error {
	var rp RelatedPerson
	if err := c.Bind(&rp); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateRelatedPerson(c.Request().Context(), &rp); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, rp)
}

func (h *Handler) GetRelatedPerson(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	rp, err := h.svc.GetRelatedPerson(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "related person not found")
	}
	return c.JSON(http.StatusOK, rp)
}

func (h *Handler) ListRelatedPersons(c echo.Context) error {
	pg := pagination.FromContext(c)
	if patientID := c.QueryParam("patient_id"); patientID != "" {
		pid, err := uuid.Parse(patientID)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid patient_id")
		}
		items, total, err := h.svc.ListRelatedPersonsByPatient(c.Request().Context(), pid, pg.Limit, pg.Offset)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
	}
	items, total, err := h.svc.SearchRelatedPersons(c.Request().Context(), nil, pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *Handler) UpdateRelatedPerson(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var rp RelatedPerson
	if err := c.Bind(&rp); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	rp.ID = id
	if err := h.svc.UpdateRelatedPerson(c.Request().Context(), &rp); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, rp)
}

func (h *Handler) DeleteRelatedPerson(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeleteRelatedPerson(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

// -- FHIR Handlers --

func (h *Handler) SearchFHIR(c echo.Context) error {
	pg := pagination.FromContext(c)
	params := map[string]string{}
	for _, k := range []string{"patient", "relationship"} {
		if v := c.QueryParam(k); v != "" {
			params[k] = v
		}
	}
	items, total, err := h.svc.SearchRelatedPersons(c.Request().Context(), params, pg.Limit, pg.Offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	resources := make([]interface{}, len(items))
	for i, item := range items {
		resources[i] = item.ToFHIR()
	}
	return c.JSON(http.StatusOK, fhir.NewSearchBundle(resources, total, "/fhir/RelatedPerson"))
}

func (h *Handler) GetFHIR(c echo.Context) error {
	rp, err := h.svc.GetRelatedPersonByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("RelatedPerson", c.Param("id")))
	}
	return c.JSON(http.StatusOK, rp.ToFHIR())
}

func (h *Handler) CreateFHIR(c echo.Context) error {
	var rp RelatedPerson
	if err := c.Bind(&rp); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	if err := h.svc.CreateRelatedPerson(c.Request().Context(), &rp); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	c.Response().Header().Set("Location", "/fhir/RelatedPerson/"+rp.FHIRID)
	return c.JSON(http.StatusCreated, rp.ToFHIR())
}

func (h *Handler) UpdateFHIR(c echo.Context) error {
	var rp RelatedPerson
	if err := c.Bind(&rp); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	existing, err := h.svc.GetRelatedPersonByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("RelatedPerson", c.Param("id")))
	}
	rp.ID = existing.ID
	rp.FHIRID = existing.FHIRID
	if err := h.svc.UpdateRelatedPerson(c.Request().Context(), &rp); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, rp.ToFHIR())
}

func (h *Handler) DeleteFHIR(c echo.Context) error {
	existing, err := h.svc.GetRelatedPersonByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("RelatedPerson", c.Param("id")))
	}
	if err := h.svc.DeleteRelatedPerson(c.Request().Context(), existing.ID); err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) VreadFHIR(c echo.Context) error {
	rp, err := h.svc.GetRelatedPersonByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("RelatedPerson", c.Param("id")))
	}
	result := rp.ToFHIR()
	fhir.SetVersionHeaders(c, 1, rp.UpdatedAt.Format("2006-01-02T15:04:05Z"))
	return c.JSON(http.StatusOK, result)
}

func (h *Handler) HistoryFHIR(c echo.Context) error {
	rp, err := h.svc.GetRelatedPersonByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("RelatedPerson", c.Param("id")))
	}
	result := rp.ToFHIR()
	raw, _ := json.Marshal(result)
	entry := &fhir.HistoryEntry{
		ResourceType: "RelatedPerson", ResourceID: rp.FHIRID, VersionID: 1,
		Resource: raw, Action: "create", Timestamp: rp.CreatedAt,
	}
	return c.JSON(http.StatusOK, fhir.NewHistoryBundle([]*fhir.HistoryEntry{entry}, 1, "/fhir"))
}
