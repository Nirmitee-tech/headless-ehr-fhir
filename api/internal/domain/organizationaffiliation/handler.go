package organizationaffiliation

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
	read.GET("/organization-affiliations", h.ListOrganizationAffiliations)
	read.GET("/organization-affiliations/:id", h.GetOrganizationAffiliation)

	write := api.Group("", role)
	write.POST("/organization-affiliations", h.CreateOrganizationAffiliation)
	write.PUT("/organization-affiliations/:id", h.UpdateOrganizationAffiliation)
	write.DELETE("/organization-affiliations/:id", h.DeleteOrganizationAffiliation)

	fhirRead := fhirGroup.Group("", role)
	fhirRead.GET("/OrganizationAffiliation", h.SearchOrganizationAffiliationsFHIR)
	fhirRead.GET("/OrganizationAffiliation/:id", h.GetOrganizationAffiliationFHIR)

	fhirWrite := fhirGroup.Group("", role)
	fhirWrite.POST("/OrganizationAffiliation", h.CreateOrganizationAffiliationFHIR)
	fhirWrite.PUT("/OrganizationAffiliation/:id", h.UpdateOrganizationAffiliationFHIR)
	fhirWrite.DELETE("/OrganizationAffiliation/:id", h.DeleteOrganizationAffiliationFHIR)
	fhirWrite.PATCH("/OrganizationAffiliation/:id", h.PatchOrganizationAffiliationFHIR)

	fhirRead.POST("/OrganizationAffiliation/_search", h.SearchOrganizationAffiliationsFHIR)
	fhirRead.GET("/OrganizationAffiliation/:id/_history/:vid", h.VreadOrganizationAffiliationFHIR)
	fhirRead.GET("/OrganizationAffiliation/:id/_history", h.HistoryOrganizationAffiliationFHIR)
}

// -- REST Endpoints --

func (h *Handler) CreateOrganizationAffiliation(c echo.Context) error {
	var o OrganizationAffiliation
	if err := c.Bind(&o); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateOrganizationAffiliation(c.Request().Context(), &o); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, o)
}

func (h *Handler) GetOrganizationAffiliation(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	o, err := h.svc.GetOrganizationAffiliation(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "organization affiliation not found")
	}
	return c.JSON(http.StatusOK, o)
}

func (h *Handler) ListOrganizationAffiliations(c echo.Context) error {
	pg := pagination.FromContext(c)
	items, total, err := h.svc.SearchOrganizationAffiliations(c.Request().Context(), nil, pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *Handler) UpdateOrganizationAffiliation(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var o OrganizationAffiliation
	if err := c.Bind(&o); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	o.ID = id
	if err := h.svc.UpdateOrganizationAffiliation(c.Request().Context(), &o); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, o)
}

func (h *Handler) DeleteOrganizationAffiliation(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeleteOrganizationAffiliation(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

// -- FHIR Endpoints --

func (h *Handler) SearchOrganizationAffiliationsFHIR(c echo.Context) error {
	pg := pagination.FromContext(c)
	params := map[string]string{}
	for _, k := range []string{"active", "organization", "participating-organization", "specialty"} {
		if v := c.QueryParam(k); v != "" {
			params[k] = v
		}
	}
	items, total, err := h.svc.SearchOrganizationAffiliations(c.Request().Context(), params, pg.Limit, pg.Offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	resources := make([]interface{}, len(items))
	for i, item := range items {
		resources[i] = item.ToFHIR()
	}
	return c.JSON(http.StatusOK, fhir.NewSearchBundle(resources, total, "/fhir/OrganizationAffiliation"))
}

func (h *Handler) GetOrganizationAffiliationFHIR(c echo.Context) error {
	o, err := h.svc.GetOrganizationAffiliationByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("OrganizationAffiliation", c.Param("id")))
	}
	return c.JSON(http.StatusOK, o.ToFHIR())
}

func (h *Handler) CreateOrganizationAffiliationFHIR(c echo.Context) error {
	var o OrganizationAffiliation
	if err := c.Bind(&o); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	if err := h.svc.CreateOrganizationAffiliation(c.Request().Context(), &o); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	c.Response().Header().Set("Location", "/fhir/OrganizationAffiliation/"+o.FHIRID)
	return c.JSON(http.StatusCreated, o.ToFHIR())
}

func (h *Handler) UpdateOrganizationAffiliationFHIR(c echo.Context) error {
	var o OrganizationAffiliation
	if err := c.Bind(&o); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	existing, err := h.svc.GetOrganizationAffiliationByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("OrganizationAffiliation", c.Param("id")))
	}
	o.ID = existing.ID
	o.FHIRID = existing.FHIRID
	if err := h.svc.UpdateOrganizationAffiliation(c.Request().Context(), &o); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, o.ToFHIR())
}

func (h *Handler) DeleteOrganizationAffiliationFHIR(c echo.Context) error {
	existing, err := h.svc.GetOrganizationAffiliationByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("OrganizationAffiliation", c.Param("id")))
	}
	if err := h.svc.DeleteOrganizationAffiliation(c.Request().Context(), existing.ID); err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) PatchOrganizationAffiliationFHIR(c echo.Context) error {
	return h.handlePatch(c, c.Param("id"))
}

func (h *Handler) VreadOrganizationAffiliationFHIR(c echo.Context) error {
	o, err := h.svc.GetOrganizationAffiliationByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("OrganizationAffiliation", c.Param("id")))
	}
	result := o.ToFHIR()
	fhir.SetVersionHeaders(c, 1, o.UpdatedAt.Format("2006-01-02T15:04:05Z"))
	return c.JSON(http.StatusOK, result)
}

func (h *Handler) HistoryOrganizationAffiliationFHIR(c echo.Context) error {
	o, err := h.svc.GetOrganizationAffiliationByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("OrganizationAffiliation", c.Param("id")))
	}
	result := o.ToFHIR()
	raw, _ := json.Marshal(result)
	entry := &fhir.HistoryEntry{
		ResourceType: "OrganizationAffiliation", ResourceID: o.FHIRID, VersionID: 1,
		Resource: raw, Action: "create", Timestamp: o.CreatedAt,
	}
	return c.JSON(http.StatusOK, fhir.NewHistoryBundle([]*fhir.HistoryEntry{entry}, 1, "/fhir"))
}

func (h *Handler) handlePatch(c echo.Context, fhirID string) error {
	contentType := c.Request().Header.Get("Content-Type")
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome("failed to read request body"))
	}
	existing, err := h.svc.GetOrganizationAffiliationByFHIRID(c.Request().Context(), fhirID)
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("OrganizationAffiliation", fhirID))
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
	if v, ok := patched["active"].(bool); ok {
		existing.Active = v
	}
	if err := h.svc.UpdateOrganizationAffiliation(c.Request().Context(), existing); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, existing.ToFHIR())
}
