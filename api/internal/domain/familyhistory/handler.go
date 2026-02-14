package familyhistory

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/ehr/ehr/internal/platform/auth"
	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/ehr/ehr/pkg/pagination"
)

// Handler provides HTTP handlers for the FamilyHistory domain.
type Handler struct {
	svc *Service
}

// NewHandler creates a new FamilyHistory domain handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// RegisterRoutes registers all FamilyHistory domain routes.
func (h *Handler) RegisterRoutes(api *echo.Group, fhirGroup *echo.Group) {
	role := auth.RequireRole("admin", "physician", "nurse")

	read := api.Group("", role)
	read.GET("/family-member-histories", h.ListFamilyMemberHistories)
	read.GET("/family-member-histories/:id", h.GetFamilyMemberHistory)

	write := api.Group("", role)
	write.POST("/family-member-histories", h.CreateFamilyMemberHistory)
	write.PUT("/family-member-histories/:id", h.UpdateFamilyMemberHistory)
	write.DELETE("/family-member-histories/:id", h.DeleteFamilyMemberHistory)

	fhirRead := fhirGroup.Group("", role)
	fhirRead.GET("/FamilyMemberHistory", h.SearchFHIR)
	fhirRead.GET("/FamilyMemberHistory/:id", h.GetFHIR)
	fhirRead.POST("/FamilyMemberHistory/_search", h.SearchFHIR)
	fhirRead.GET("/FamilyMemberHistory/:id/_history/:vid", h.VreadFHIR)
	fhirRead.GET("/FamilyMemberHistory/:id/_history", h.HistoryFHIR)

	fhirWrite := fhirGroup.Group("", role)
	fhirWrite.POST("/FamilyMemberHistory", h.CreateFHIR)
	fhirWrite.PUT("/FamilyMemberHistory/:id", h.UpdateFHIR)
	fhirWrite.DELETE("/FamilyMemberHistory/:id", h.DeleteFHIR)
}

// -- REST Handlers --

func (h *Handler) CreateFamilyMemberHistory(c echo.Context) error {
	var f FamilyMemberHistory
	if err := c.Bind(&f); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateFamilyMemberHistory(c.Request().Context(), &f); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, f)
}

func (h *Handler) GetFamilyMemberHistory(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	f, err := h.svc.GetFamilyMemberHistory(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "family member history not found")
	}
	return c.JSON(http.StatusOK, f)
}

func (h *Handler) ListFamilyMemberHistories(c echo.Context) error {
	pg := pagination.FromContext(c)
	if patientID := c.QueryParam("patient_id"); patientID != "" {
		pid, err := uuid.Parse(patientID)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid patient_id")
		}
		items, total, err := h.svc.ListFamilyMemberHistoriesByPatient(c.Request().Context(), pid, pg.Limit, pg.Offset)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
	}
	items, total, err := h.svc.SearchFamilyMemberHistories(c.Request().Context(), nil, pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *Handler) UpdateFamilyMemberHistory(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var f FamilyMemberHistory
	if err := c.Bind(&f); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	f.ID = id
	if err := h.svc.UpdateFamilyMemberHistory(c.Request().Context(), &f); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, f)
}

func (h *Handler) DeleteFamilyMemberHistory(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeleteFamilyMemberHistory(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

// -- FHIR Handlers --

func (h *Handler) SearchFHIR(c echo.Context) error {
	pg := pagination.FromContext(c)
	params := map[string]string{}
	for _, k := range []string{"patient", "status", "relationship"} {
		if v := c.QueryParam(k); v != "" {
			params[k] = v
		}
	}
	items, total, err := h.svc.SearchFamilyMemberHistories(c.Request().Context(), params, pg.Limit, pg.Offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	resources := make([]interface{}, len(items))
	for i, item := range items {
		resources[i] = item.ToFHIR()
	}
	return c.JSON(http.StatusOK, fhir.NewSearchBundle(resources, total, "/fhir/FamilyMemberHistory"))
}

func (h *Handler) GetFHIR(c echo.Context) error {
	f, err := h.svc.GetFamilyMemberHistoryByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("FamilyMemberHistory", c.Param("id")))
	}
	return c.JSON(http.StatusOK, f.ToFHIR())
}

func (h *Handler) CreateFHIR(c echo.Context) error {
	var f FamilyMemberHistory
	if err := c.Bind(&f); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	if err := h.svc.CreateFamilyMemberHistory(c.Request().Context(), &f); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	c.Response().Header().Set("Location", "/fhir/FamilyMemberHistory/"+f.FHIRID)
	return c.JSON(http.StatusCreated, f.ToFHIR())
}

func (h *Handler) UpdateFHIR(c echo.Context) error {
	var f FamilyMemberHistory
	if err := c.Bind(&f); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	existing, err := h.svc.GetFamilyMemberHistoryByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("FamilyMemberHistory", c.Param("id")))
	}
	f.ID = existing.ID
	f.FHIRID = existing.FHIRID
	if err := h.svc.UpdateFamilyMemberHistory(c.Request().Context(), &f); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, f.ToFHIR())
}

func (h *Handler) DeleteFHIR(c echo.Context) error {
	existing, err := h.svc.GetFamilyMemberHistoryByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("FamilyMemberHistory", c.Param("id")))
	}
	if err := h.svc.DeleteFamilyMemberHistory(c.Request().Context(), existing.ID); err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) VreadFHIR(c echo.Context) error {
	f, err := h.svc.GetFamilyMemberHistoryByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("FamilyMemberHistory", c.Param("id")))
	}
	result := f.ToFHIR()
	fhir.SetVersionHeaders(c, 1, f.UpdatedAt.Format("2006-01-02T15:04:05Z"))
	return c.JSON(http.StatusOK, result)
}

func (h *Handler) HistoryFHIR(c echo.Context) error {
	f, err := h.svc.GetFamilyMemberHistoryByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("FamilyMemberHistory", c.Param("id")))
	}
	result := f.ToFHIR()
	raw, _ := json.Marshal(result)
	entry := &fhir.HistoryEntry{
		ResourceType: "FamilyMemberHistory", ResourceID: f.FHIRID, VersionID: 1,
		Resource: raw, Action: "create", Timestamp: f.CreatedAt,
	}
	return c.JSON(http.StatusOK, fhir.NewHistoryBundle([]*fhir.HistoryEntry{entry}, 1, "/fhir"))
}
