package admin

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
	// Read endpoints – admin, physician, nurse, registrar
	readGroup := api.Group("", auth.RequireRole("admin", "physician", "nurse", "registrar"))
	readGroup.GET("/organizations", h.ListOrganizations)
	readGroup.GET("/organizations/:id", h.GetOrganization)
	readGroup.GET("/departments", h.ListDepartments)
	readGroup.GET("/departments/:id", h.GetDepartment)
	readGroup.GET("/locations", h.ListLocations)
	readGroup.GET("/locations/:id", h.GetLocation)
	readGroup.GET("/users", h.ListSystemUsers)
	readGroup.GET("/users/:id", h.GetSystemUser)
	readGroup.GET("/users/:id/roles", h.GetUserRoles)

	// Write endpoints – admin only
	writeGroup := api.Group("", auth.RequireRole("admin"))
	writeGroup.POST("/organizations", h.CreateOrganization)
	writeGroup.PUT("/organizations/:id", h.UpdateOrganization)
	writeGroup.DELETE("/organizations/:id", h.DeleteOrganization)
	writeGroup.POST("/departments", h.CreateDepartment)
	writeGroup.PUT("/departments/:id", h.UpdateDepartment)
	writeGroup.DELETE("/departments/:id", h.DeleteDepartment)
	writeGroup.POST("/locations", h.CreateLocation)
	writeGroup.PUT("/locations/:id", h.UpdateLocation)
	writeGroup.DELETE("/locations/:id", h.DeleteLocation)
	writeGroup.POST("/users", h.CreateSystemUser)
	writeGroup.PUT("/users/:id", h.UpdateSystemUser)
	writeGroup.DELETE("/users/:id", h.DeleteSystemUser)
	writeGroup.POST("/users/:id/roles", h.AssignRole)
	writeGroup.DELETE("/users/:id/roles/:role_id", h.RemoveRole)

	// FHIR read endpoints
	fhirRead := fhirGroup.Group("", auth.RequireRole("admin", "physician", "nurse", "registrar"))
	fhirRead.GET("/Organization", h.SearchOrganizationsFHIR)
	fhirRead.GET("/Organization/:id", h.GetOrganizationFHIR)
	fhirRead.GET("/Location", h.SearchLocationsFHIR)
	fhirRead.GET("/Location/:id", h.GetLocationFHIR)

	// FHIR write endpoints
	fhirWrite := fhirGroup.Group("", auth.RequireRole("admin"))
	fhirWrite.POST("/Organization", h.CreateOrganizationFHIR)
	fhirWrite.PUT("/Organization/:id", h.UpdateOrganizationFHIR)
	fhirWrite.DELETE("/Organization/:id", h.DeleteOrganizationFHIR)
	fhirWrite.PATCH("/Organization/:id", h.PatchOrganizationFHIR)
	fhirWrite.POST("/Location", h.CreateLocationFHIR)
	fhirWrite.PUT("/Location/:id", h.UpdateLocationFHIR)
	fhirWrite.DELETE("/Location/:id", h.DeleteLocationFHIR)
	fhirWrite.PATCH("/Location/:id", h.PatchLocationFHIR)

	// FHIR POST _search endpoints
	fhirRead.POST("/Organization/_search", h.SearchOrganizationsFHIR)
	fhirRead.POST("/Location/_search", h.SearchLocationsFHIR)

	// FHIR vread and history endpoints
	fhirRead.GET("/Organization/:id/_history/:vid", h.VreadOrganizationFHIR)
	fhirRead.GET("/Organization/:id/_history", h.HistoryOrganizationFHIR)
	fhirRead.GET("/Location/:id/_history/:vid", h.VreadLocationFHIR)
	fhirRead.GET("/Location/:id/_history", h.HistoryLocationFHIR)
}

// -- Organization Handlers (Operational) --

func (h *Handler) CreateOrganization(c echo.Context) error {
	var org Organization
	if err := c.Bind(&org); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateOrganization(c.Request().Context(), &org); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, org)
}

func (h *Handler) GetOrganization(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	org, err := h.svc.GetOrganization(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "organization not found")
	}
	return c.JSON(http.StatusOK, org)
}

func (h *Handler) ListOrganizations(c echo.Context) error {
	p := pagination.FromContext(c)
	orgs, total, err := h.svc.ListOrganizations(c.Request().Context(), p.Limit, p.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(orgs, total, p.Limit, p.Offset))
}

func (h *Handler) UpdateOrganization(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var org Organization
	if err := c.Bind(&org); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	org.ID = id
	if err := h.svc.UpdateOrganization(c.Request().Context(), &org); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, org)
}

func (h *Handler) DeleteOrganization(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeleteOrganization(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

// -- Department Handlers --

func (h *Handler) CreateDepartment(c echo.Context) error {
	var dept Department
	if err := c.Bind(&dept); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateDepartment(c.Request().Context(), &dept); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, dept)
}

func (h *Handler) GetDepartment(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	dept, err := h.svc.GetDepartment(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "department not found")
	}
	return c.JSON(http.StatusOK, dept)
}

func (h *Handler) ListDepartments(c echo.Context) error {
	orgIDStr := c.QueryParam("organization_id")
	if orgIDStr == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "organization_id is required")
	}
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid organization_id")
	}
	p := pagination.FromContext(c)
	depts, total, err := h.svc.ListDepartments(c.Request().Context(), orgID, p.Limit, p.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(depts, total, p.Limit, p.Offset))
}

func (h *Handler) UpdateDepartment(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var dept Department
	if err := c.Bind(&dept); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	dept.ID = id
	if err := h.svc.UpdateDepartment(c.Request().Context(), &dept); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, dept)
}

func (h *Handler) DeleteDepartment(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeleteDepartment(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

// -- Location Handlers --

func (h *Handler) CreateLocation(c echo.Context) error {
	var loc Location
	if err := c.Bind(&loc); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateLocation(c.Request().Context(), &loc); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, loc)
}

func (h *Handler) GetLocation(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	loc, err := h.svc.GetLocation(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "location not found")
	}
	return c.JSON(http.StatusOK, loc)
}

func (h *Handler) ListLocations(c echo.Context) error {
	p := pagination.FromContext(c)
	locs, total, err := h.svc.ListLocations(c.Request().Context(), p.Limit, p.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(locs, total, p.Limit, p.Offset))
}

func (h *Handler) UpdateLocation(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var loc Location
	if err := c.Bind(&loc); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	loc.ID = id
	if err := h.svc.UpdateLocation(c.Request().Context(), &loc); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, loc)
}

func (h *Handler) DeleteLocation(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeleteLocation(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

// -- System User Handlers --

func (h *Handler) CreateSystemUser(c echo.Context) error {
	var user SystemUser
	if err := c.Bind(&user); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateSystemUser(c.Request().Context(), &user); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, user)
}

func (h *Handler) GetSystemUser(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	user, err := h.svc.GetSystemUser(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "user not found")
	}
	return c.JSON(http.StatusOK, user)
}

func (h *Handler) ListSystemUsers(c echo.Context) error {
	p := pagination.FromContext(c)
	users, total, err := h.svc.ListSystemUsers(c.Request().Context(), p.Limit, p.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(users, total, p.Limit, p.Offset))
}

func (h *Handler) UpdateSystemUser(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var user SystemUser
	if err := c.Bind(&user); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	user.ID = id
	if err := h.svc.UpdateSystemUser(c.Request().Context(), &user); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, user)
}

func (h *Handler) DeleteSystemUser(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeleteSystemUser(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) AssignRole(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var assignment UserRoleAssignment
	if err := c.Bind(&assignment); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	assignment.UserID = id
	if err := h.svc.AssignRole(c.Request().Context(), &assignment); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, assignment)
}

func (h *Handler) GetUserRoles(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	roles, err := h.svc.GetUserRoles(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, roles)
}

func (h *Handler) RemoveRole(c echo.Context) error {
	roleID, err := uuid.Parse(c.Param("role_id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid role_id")
	}
	if err := h.svc.RemoveRole(c.Request().Context(), roleID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

// -- FHIR Organization Handlers --

func (h *Handler) SearchOrganizationsFHIR(c echo.Context) error {
	p := pagination.FromContext(c)
	params := fhir.ExtractSearchParams(c)

	orgs, total, err := h.svc.SearchOrganizations(c.Request().Context(), params, p.Limit, p.Offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}

	resources := make([]interface{}, len(orgs))
	for i, org := range orgs {
		resources[i] = org.ToFHIR()
	}
	return c.JSON(http.StatusOK, fhir.NewSearchBundleWithLinks(resources, fhir.SearchBundleParams{
		BaseURL:  "/fhir/Organization",
		QueryStr: c.QueryString(),
		Count:    p.Limit,
		Offset:   p.Offset,
		Total:    total,
	}))
}

func (h *Handler) GetOrganizationFHIR(c echo.Context) error {
	org, err := h.svc.GetOrganizationByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Organization", c.Param("id")))
	}
	return c.JSON(http.StatusOK, org.ToFHIR())
}

func (h *Handler) CreateOrganizationFHIR(c echo.Context) error {
	var org Organization
	if err := c.Bind(&org); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	if err := h.svc.CreateOrganization(c.Request().Context(), &org); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	c.Response().Header().Set("Location", "/fhir/Organization/"+org.FHIRID)
	return c.JSON(http.StatusCreated, org.ToFHIR())
}

func (h *Handler) UpdateOrganizationFHIR(c echo.Context) error {
	existing, err := h.svc.GetOrganizationByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Organization", c.Param("id")))
	}
	var org Organization
	if err := c.Bind(&org); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	org.ID = existing.ID
	org.FHIRID = existing.FHIRID
	if err := h.svc.UpdateOrganization(c.Request().Context(), &org); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, org.ToFHIR())
}

func (h *Handler) DeleteOrganizationFHIR(c echo.Context) error {
	org, err := h.svc.GetOrganizationByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Organization", c.Param("id")))
	}
	if err := h.svc.DeleteOrganization(c.Request().Context(), org.ID); err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	return c.NoContent(http.StatusNoContent)
}

// -- FHIR Location Handlers --

func (h *Handler) SearchLocationsFHIR(c echo.Context) error {
	p := pagination.FromContext(c)
	locs, total, err := h.svc.ListLocations(c.Request().Context(), p.Limit, p.Offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	resources := make([]interface{}, len(locs))
	for i, loc := range locs {
		resources[i] = loc.ToFHIR()
	}
	return c.JSON(http.StatusOK, fhir.NewSearchBundleWithLinks(resources, fhir.SearchBundleParams{
		BaseURL:  "/fhir/Location",
		QueryStr: c.QueryString(),
		Count:    p.Limit,
		Offset:   p.Offset,
		Total:    total,
	}))
}

func (h *Handler) GetLocationFHIR(c echo.Context) error {
	loc, err := h.svc.GetLocationByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Location", c.Param("id")))
	}
	return c.JSON(http.StatusOK, loc.ToFHIR())
}

func (h *Handler) CreateLocationFHIR(c echo.Context) error {
	var loc Location
	if err := c.Bind(&loc); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	if err := h.svc.CreateLocation(c.Request().Context(), &loc); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	c.Response().Header().Set("Location", "/fhir/Location/"+loc.FHIRID)
	return c.JSON(http.StatusCreated, loc.ToFHIR())
}

func (h *Handler) UpdateLocationFHIR(c echo.Context) error {
	existing, err := h.svc.GetLocationByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Location", c.Param("id")))
	}
	var loc Location
	if err := c.Bind(&loc); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	loc.ID = existing.ID
	loc.FHIRID = existing.FHIRID
	if err := h.svc.UpdateLocation(c.Request().Context(), &loc); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, loc.ToFHIR())
}

func (h *Handler) DeleteLocationFHIR(c echo.Context) error {
	loc, err := h.svc.GetLocationByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Location", c.Param("id")))
	}
	if err := h.svc.DeleteLocation(c.Request().Context(), loc.ID); err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	return c.NoContent(http.StatusNoContent)
}

// -- FHIR Organization PATCH/vread/history --

func (h *Handler) PatchOrganizationFHIR(c echo.Context) error {
	contentType := c.Request().Header.Get("Content-Type")
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome("failed to read request body"))
	}

	existing, err := h.svc.GetOrganizationByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Organization", c.Param("id")))
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

	if v, ok := patched["name"].(string); ok {
		existing.Name = v
	}
	if v, ok := patched["active"].(bool); ok {
		existing.Active = v
	}
	if err := h.svc.UpdateOrganization(c.Request().Context(), existing); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, existing.ToFHIR())
}

func (h *Handler) VreadOrganizationFHIR(c echo.Context) error {
	org, err := h.svc.GetOrganizationByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Organization", c.Param("id")))
	}
	result := org.ToFHIR()
	fhir.SetVersionHeaders(c, 1, org.UpdatedAt.Format("2006-01-02T15:04:05Z"))
	return c.JSON(http.StatusOK, result)
}

func (h *Handler) HistoryOrganizationFHIR(c echo.Context) error {
	org, err := h.svc.GetOrganizationByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Organization", c.Param("id")))
	}
	result := org.ToFHIR()
	raw, _ := json.Marshal(result)
	entry := &fhir.HistoryEntry{
		ResourceType: "Organization", ResourceID: org.FHIRID, VersionID: 1,
		Resource: raw, Action: "create", Timestamp: org.CreatedAt,
	}
	return c.JSON(http.StatusOK, fhir.NewHistoryBundle([]*fhir.HistoryEntry{entry}, 1, "/fhir"))
}

// -- FHIR Location PATCH/vread/history --

func (h *Handler) PatchLocationFHIR(c echo.Context) error {
	contentType := c.Request().Header.Get("Content-Type")
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome("failed to read request body"))
	}

	existing, err := h.svc.GetLocationByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Location", c.Param("id")))
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
	if v, ok := patched["name"].(string); ok {
		existing.Name = v
	}
	if err := h.svc.UpdateLocation(c.Request().Context(), existing); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, existing.ToFHIR())
}

func (h *Handler) VreadLocationFHIR(c echo.Context) error {
	loc, err := h.svc.GetLocationByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Location", c.Param("id")))
	}
	result := loc.ToFHIR()
	fhir.SetVersionHeaders(c, 1, loc.CreatedAt.Format("2006-01-02T15:04:05Z"))
	return c.JSON(http.StatusOK, result)
}

func (h *Handler) HistoryLocationFHIR(c echo.Context) error {
	loc, err := h.svc.GetLocationByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Location", c.Param("id")))
	}
	result := loc.ToFHIR()
	raw, _ := json.Marshal(result)
	entry := &fhir.HistoryEntry{
		ResourceType: "Location", ResourceID: loc.FHIRID, VersionID: 1,
		Resource: raw, Action: "create", Timestamp: loc.CreatedAt,
	}
	return c.JSON(http.StatusOK, fhir.NewHistoryBundle([]*fhir.HistoryEntry{entry}, 1, "/fhir"))
}
