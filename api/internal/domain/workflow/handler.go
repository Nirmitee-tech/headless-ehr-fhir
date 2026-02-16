package workflow

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
	role := auth.RequireRole("admin", "physician")

	// REST read endpoints
	read := api.Group("", role)
	read.GET("/activity-definitions", h.ListActivityDefinitions)
	read.GET("/activity-definitions/:id", h.GetActivityDefinition)
	read.GET("/request-groups", h.ListRequestGroups)
	read.GET("/request-groups/:id", h.GetRequestGroup)
	read.GET("/request-groups/:id/actions", h.GetRequestGroupActions)
	read.GET("/guidance-responses", h.ListGuidanceResponses)
	read.GET("/guidance-responses/:id", h.GetGuidanceResponse)

	// REST write endpoints
	write := api.Group("", role)
	write.POST("/activity-definitions", h.CreateActivityDefinition)
	write.PUT("/activity-definitions/:id", h.UpdateActivityDefinition)
	write.DELETE("/activity-definitions/:id", h.DeleteActivityDefinition)
	write.POST("/request-groups", h.CreateRequestGroup)
	write.PUT("/request-groups/:id", h.UpdateRequestGroup)
	write.DELETE("/request-groups/:id", h.DeleteRequestGroup)
	write.POST("/request-groups/:id/actions", h.AddRequestGroupAction)
	write.POST("/guidance-responses", h.CreateGuidanceResponse)
	write.PUT("/guidance-responses/:id", h.UpdateGuidanceResponse)
	write.DELETE("/guidance-responses/:id", h.DeleteGuidanceResponse)

	// FHIR read endpoints
	fhirRead := fhirGroup.Group("", role)
	fhirRead.GET("/ActivityDefinition", h.SearchActivityDefinitionsFHIR)
	fhirRead.GET("/ActivityDefinition/:id", h.GetActivityDefinitionFHIR)
	fhirRead.GET("/RequestGroup", h.SearchRequestGroupsFHIR)
	fhirRead.GET("/RequestGroup/:id", h.GetRequestGroupFHIR)
	fhirRead.GET("/GuidanceResponse", h.SearchGuidanceResponsesFHIR)
	fhirRead.GET("/GuidanceResponse/:id", h.GetGuidanceResponseFHIR)

	// FHIR write endpoints
	fhirWrite := fhirGroup.Group("", role)
	fhirWrite.POST("/ActivityDefinition", h.CreateActivityDefinitionFHIR)
	fhirWrite.PUT("/ActivityDefinition/:id", h.UpdateActivityDefinitionFHIR)
	fhirWrite.DELETE("/ActivityDefinition/:id", h.DeleteActivityDefinitionFHIR)
	fhirWrite.PATCH("/ActivityDefinition/:id", h.PatchActivityDefinitionFHIR)
	fhirWrite.POST("/RequestGroup", h.CreateRequestGroupFHIR)
	fhirWrite.PUT("/RequestGroup/:id", h.UpdateRequestGroupFHIR)
	fhirWrite.DELETE("/RequestGroup/:id", h.DeleteRequestGroupFHIR)
	fhirWrite.PATCH("/RequestGroup/:id", h.PatchRequestGroupFHIR)
	fhirWrite.POST("/GuidanceResponse", h.CreateGuidanceResponseFHIR)
	fhirWrite.PUT("/GuidanceResponse/:id", h.UpdateGuidanceResponseFHIR)
	fhirWrite.DELETE("/GuidanceResponse/:id", h.DeleteGuidanceResponseFHIR)
	fhirWrite.PATCH("/GuidanceResponse/:id", h.PatchGuidanceResponseFHIR)

	// FHIR POST _search endpoints
	fhirRead.POST("/ActivityDefinition/_search", h.SearchActivityDefinitionsFHIR)
	fhirRead.POST("/RequestGroup/_search", h.SearchRequestGroupsFHIR)
	fhirRead.POST("/GuidanceResponse/_search", h.SearchGuidanceResponsesFHIR)

	// FHIR vread and history endpoints
	fhirRead.GET("/ActivityDefinition/:id/_history/:vid", h.VreadActivityDefinitionFHIR)
	fhirRead.GET("/ActivityDefinition/:id/_history", h.HistoryActivityDefinitionFHIR)
	fhirRead.GET("/RequestGroup/:id/_history/:vid", h.VreadRequestGroupFHIR)
	fhirRead.GET("/RequestGroup/:id/_history", h.HistoryRequestGroupFHIR)
	fhirRead.GET("/GuidanceResponse/:id/_history/:vid", h.VreadGuidanceResponseFHIR)
	fhirRead.GET("/GuidanceResponse/:id/_history", h.HistoryGuidanceResponseFHIR)
}

// -- ActivityDefinition REST --

func (h *Handler) CreateActivityDefinition(c echo.Context) error {
	var a ActivityDefinition
	if err := c.Bind(&a); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateActivityDefinition(c.Request().Context(), &a); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, a)
}

func (h *Handler) GetActivityDefinition(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	a, err := h.svc.GetActivityDefinition(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "activity definition not found")
	}
	return c.JSON(http.StatusOK, a)
}

func (h *Handler) ListActivityDefinitions(c echo.Context) error {
	pg := pagination.FromContext(c)
	items, total, err := h.svc.SearchActivityDefinitions(c.Request().Context(), nil, pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *Handler) UpdateActivityDefinition(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var a ActivityDefinition
	if err := c.Bind(&a); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	a.ID = id
	if err := h.svc.UpdateActivityDefinition(c.Request().Context(), &a); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, a)
}

func (h *Handler) DeleteActivityDefinition(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeleteActivityDefinition(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

// -- RequestGroup REST --

func (h *Handler) CreateRequestGroup(c echo.Context) error {
	var rg RequestGroup
	if err := c.Bind(&rg); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateRequestGroup(c.Request().Context(), &rg); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, rg)
}

func (h *Handler) GetRequestGroup(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	rg, err := h.svc.GetRequestGroup(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "request group not found")
	}
	return c.JSON(http.StatusOK, rg)
}

func (h *Handler) ListRequestGroups(c echo.Context) error {
	pg := pagination.FromContext(c)
	items, total, err := h.svc.SearchRequestGroups(c.Request().Context(), nil, pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *Handler) UpdateRequestGroup(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var rg RequestGroup
	if err := c.Bind(&rg); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	rg.ID = id
	if err := h.svc.UpdateRequestGroup(c.Request().Context(), &rg); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, rg)
}

func (h *Handler) DeleteRequestGroup(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeleteRequestGroup(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) AddRequestGroupAction(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var a RequestGroupAction
	if err := c.Bind(&a); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	a.RequestGroupID = id
	if err := h.svc.AddRequestGroupAction(c.Request().Context(), &a); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, a)
}

func (h *Handler) GetRequestGroupActions(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	items, err := h.svc.GetRequestGroupActions(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, items)
}

// -- GuidanceResponse REST --

func (h *Handler) CreateGuidanceResponse(c echo.Context) error {
	var gr GuidanceResponse
	if err := c.Bind(&gr); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateGuidanceResponse(c.Request().Context(), &gr); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, gr)
}

func (h *Handler) GetGuidanceResponse(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	gr, err := h.svc.GetGuidanceResponse(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "guidance response not found")
	}
	return c.JSON(http.StatusOK, gr)
}

func (h *Handler) ListGuidanceResponses(c echo.Context) error {
	pg := pagination.FromContext(c)
	items, total, err := h.svc.SearchGuidanceResponses(c.Request().Context(), nil, pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *Handler) UpdateGuidanceResponse(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var gr GuidanceResponse
	if err := c.Bind(&gr); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	gr.ID = id
	if err := h.svc.UpdateGuidanceResponse(c.Request().Context(), &gr); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, gr)
}

func (h *Handler) DeleteGuidanceResponse(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeleteGuidanceResponse(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

// -- FHIR ActivityDefinition --

func (h *Handler) SearchActivityDefinitionsFHIR(c echo.Context) error {
	pg := pagination.FromContext(c)
	params := fhir.ExtractSearchParams(c)
	items, total, err := h.svc.SearchActivityDefinitions(c.Request().Context(), params, pg.Limit, pg.Offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	resources := make([]interface{}, len(items))
	for i, item := range items {
		resources[i] = item.ToFHIR()
	}
	return c.JSON(http.StatusOK, fhir.NewSearchBundle(resources, total, "/fhir/ActivityDefinition"))
}

func (h *Handler) GetActivityDefinitionFHIR(c echo.Context) error {
	a, err := h.svc.GetActivityDefinitionByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("ActivityDefinition", c.Param("id")))
	}
	return c.JSON(http.StatusOK, a.ToFHIR())
}

func (h *Handler) CreateActivityDefinitionFHIR(c echo.Context) error {
	var a ActivityDefinition
	if err := c.Bind(&a); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	if err := h.svc.CreateActivityDefinition(c.Request().Context(), &a); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	c.Response().Header().Set("Location", "/fhir/ActivityDefinition/"+a.FHIRID)
	return c.JSON(http.StatusCreated, a.ToFHIR())
}

func (h *Handler) UpdateActivityDefinitionFHIR(c echo.Context) error {
	var a ActivityDefinition
	if err := c.Bind(&a); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	existing, err := h.svc.GetActivityDefinitionByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("ActivityDefinition", c.Param("id")))
	}
	a.ID = existing.ID
	a.FHIRID = existing.FHIRID
	if err := h.svc.UpdateActivityDefinition(c.Request().Context(), &a); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, a.ToFHIR())
}

func (h *Handler) DeleteActivityDefinitionFHIR(c echo.Context) error {
	existing, err := h.svc.GetActivityDefinitionByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("ActivityDefinition", c.Param("id")))
	}
	if err := h.svc.DeleteActivityDefinition(c.Request().Context(), existing.ID); err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) PatchActivityDefinitionFHIR(c echo.Context) error {
	return h.handlePatch(c, "ActivityDefinition", c.Param("id"))
}

func (h *Handler) VreadActivityDefinitionFHIR(c echo.Context) error {
	a, err := h.svc.GetActivityDefinitionByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("ActivityDefinition", c.Param("id")))
	}
	result := a.ToFHIR()
	fhir.SetVersionHeaders(c, 1, a.UpdatedAt.Format("2006-01-02T15:04:05Z"))
	return c.JSON(http.StatusOK, result)
}

func (h *Handler) HistoryActivityDefinitionFHIR(c echo.Context) error {
	a, err := h.svc.GetActivityDefinitionByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("ActivityDefinition", c.Param("id")))
	}
	result := a.ToFHIR()
	raw, _ := json.Marshal(result)
	entry := &fhir.HistoryEntry{
		ResourceType: "ActivityDefinition", ResourceID: a.FHIRID, VersionID: 1,
		Resource: raw, Action: "create", Timestamp: a.CreatedAt,
	}
	return c.JSON(http.StatusOK, fhir.NewHistoryBundle([]*fhir.HistoryEntry{entry}, 1, "/fhir"))
}

// -- FHIR RequestGroup --

func (h *Handler) SearchRequestGroupsFHIR(c echo.Context) error {
	pg := pagination.FromContext(c)
	params := fhir.ExtractSearchParams(c)
	items, total, err := h.svc.SearchRequestGroups(c.Request().Context(), params, pg.Limit, pg.Offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	resources := make([]interface{}, len(items))
	for i, item := range items {
		resources[i] = item.ToFHIR()
	}
	return c.JSON(http.StatusOK, fhir.NewSearchBundle(resources, total, "/fhir/RequestGroup"))
}

func (h *Handler) GetRequestGroupFHIR(c echo.Context) error {
	rg, err := h.svc.GetRequestGroupByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("RequestGroup", c.Param("id")))
	}
	return c.JSON(http.StatusOK, rg.ToFHIR())
}

func (h *Handler) CreateRequestGroupFHIR(c echo.Context) error {
	var rg RequestGroup
	if err := c.Bind(&rg); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	if err := h.svc.CreateRequestGroup(c.Request().Context(), &rg); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	c.Response().Header().Set("Location", "/fhir/RequestGroup/"+rg.FHIRID)
	return c.JSON(http.StatusCreated, rg.ToFHIR())
}

func (h *Handler) UpdateRequestGroupFHIR(c echo.Context) error {
	var rg RequestGroup
	if err := c.Bind(&rg); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	existing, err := h.svc.GetRequestGroupByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("RequestGroup", c.Param("id")))
	}
	rg.ID = existing.ID
	rg.FHIRID = existing.FHIRID
	if err := h.svc.UpdateRequestGroup(c.Request().Context(), &rg); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, rg.ToFHIR())
}

func (h *Handler) DeleteRequestGroupFHIR(c echo.Context) error {
	existing, err := h.svc.GetRequestGroupByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("RequestGroup", c.Param("id")))
	}
	if err := h.svc.DeleteRequestGroup(c.Request().Context(), existing.ID); err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) PatchRequestGroupFHIR(c echo.Context) error {
	return h.handlePatch(c, "RequestGroup", c.Param("id"))
}

func (h *Handler) VreadRequestGroupFHIR(c echo.Context) error {
	rg, err := h.svc.GetRequestGroupByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("RequestGroup", c.Param("id")))
	}
	result := rg.ToFHIR()
	fhir.SetVersionHeaders(c, 1, rg.UpdatedAt.Format("2006-01-02T15:04:05Z"))
	return c.JSON(http.StatusOK, result)
}

func (h *Handler) HistoryRequestGroupFHIR(c echo.Context) error {
	rg, err := h.svc.GetRequestGroupByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("RequestGroup", c.Param("id")))
	}
	result := rg.ToFHIR()
	raw, _ := json.Marshal(result)
	entry := &fhir.HistoryEntry{
		ResourceType: "RequestGroup", ResourceID: rg.FHIRID, VersionID: 1,
		Resource: raw, Action: "create", Timestamp: rg.CreatedAt,
	}
	return c.JSON(http.StatusOK, fhir.NewHistoryBundle([]*fhir.HistoryEntry{entry}, 1, "/fhir"))
}

// -- FHIR GuidanceResponse --

func (h *Handler) SearchGuidanceResponsesFHIR(c echo.Context) error {
	pg := pagination.FromContext(c)
	params := fhir.ExtractSearchParams(c)
	items, total, err := h.svc.SearchGuidanceResponses(c.Request().Context(), params, pg.Limit, pg.Offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	resources := make([]interface{}, len(items))
	for i, item := range items {
		resources[i] = item.ToFHIR()
	}
	return c.JSON(http.StatusOK, fhir.NewSearchBundle(resources, total, "/fhir/GuidanceResponse"))
}

func (h *Handler) GetGuidanceResponseFHIR(c echo.Context) error {
	gr, err := h.svc.GetGuidanceResponseByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("GuidanceResponse", c.Param("id")))
	}
	return c.JSON(http.StatusOK, gr.ToFHIR())
}

func (h *Handler) CreateGuidanceResponseFHIR(c echo.Context) error {
	var gr GuidanceResponse
	if err := c.Bind(&gr); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	if err := h.svc.CreateGuidanceResponse(c.Request().Context(), &gr); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	c.Response().Header().Set("Location", "/fhir/GuidanceResponse/"+gr.FHIRID)
	return c.JSON(http.StatusCreated, gr.ToFHIR())
}

func (h *Handler) UpdateGuidanceResponseFHIR(c echo.Context) error {
	var gr GuidanceResponse
	if err := c.Bind(&gr); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	existing, err := h.svc.GetGuidanceResponseByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("GuidanceResponse", c.Param("id")))
	}
	gr.ID = existing.ID
	gr.FHIRID = existing.FHIRID
	if err := h.svc.UpdateGuidanceResponse(c.Request().Context(), &gr); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, gr.ToFHIR())
}

func (h *Handler) DeleteGuidanceResponseFHIR(c echo.Context) error {
	existing, err := h.svc.GetGuidanceResponseByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("GuidanceResponse", c.Param("id")))
	}
	if err := h.svc.DeleteGuidanceResponse(c.Request().Context(), existing.ID); err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) PatchGuidanceResponseFHIR(c echo.Context) error {
	return h.handlePatch(c, "GuidanceResponse", c.Param("id"))
}

func (h *Handler) VreadGuidanceResponseFHIR(c echo.Context) error {
	gr, err := h.svc.GetGuidanceResponseByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("GuidanceResponse", c.Param("id")))
	}
	result := gr.ToFHIR()
	fhir.SetVersionHeaders(c, 1, gr.UpdatedAt.Format("2006-01-02T15:04:05Z"))
	return c.JSON(http.StatusOK, result)
}

func (h *Handler) HistoryGuidanceResponseFHIR(c echo.Context) error {
	gr, err := h.svc.GetGuidanceResponseByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("GuidanceResponse", c.Param("id")))
	}
	result := gr.ToFHIR()
	raw, _ := json.Marshal(result)
	entry := &fhir.HistoryEntry{
		ResourceType: "GuidanceResponse", ResourceID: gr.FHIRID, VersionID: 1,
		Resource: raw, Action: "create", Timestamp: gr.CreatedAt,
	}
	return c.JSON(http.StatusOK, fhir.NewHistoryBundle([]*fhir.HistoryEntry{entry}, 1, "/fhir"))
}

// -- FHIR PATCH --

func (h *Handler) handlePatch(c echo.Context, resourceType, fhirID string) error {
	contentType := c.Request().Header.Get("Content-Type")
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome("failed to read request body"))
	}

	var currentResource map[string]interface{}
	switch resourceType {
	case "ActivityDefinition":
		existing, err := h.svc.GetActivityDefinitionByFHIRID(c.Request().Context(), fhirID)
		if err != nil {
			return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome(resourceType, fhirID))
		}
		currentResource = existing.ToFHIR()
	case "RequestGroup":
		existing, err := h.svc.GetRequestGroupByFHIRID(c.Request().Context(), fhirID)
		if err != nil {
			return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome(resourceType, fhirID))
		}
		currentResource = existing.ToFHIR()
	case "GuidanceResponse":
		existing, err := h.svc.GetGuidanceResponseByFHIRID(c.Request().Context(), fhirID)
		if err != nil {
			return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome(resourceType, fhirID))
		}
		currentResource = existing.ToFHIR()
	default:
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome("unsupported resource type for PATCH"))
	}

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

	switch resourceType {
	case "ActivityDefinition":
		existing, _ := h.svc.GetActivityDefinitionByFHIRID(c.Request().Context(), fhirID)
		if v, ok := patched["status"].(string); ok {
			existing.Status = v
		}
		if err := h.svc.UpdateActivityDefinition(c.Request().Context(), existing); err != nil {
			return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
		}
		return c.JSON(http.StatusOK, existing.ToFHIR())
	case "RequestGroup":
		existing, _ := h.svc.GetRequestGroupByFHIRID(c.Request().Context(), fhirID)
		if v, ok := patched["status"].(string); ok {
			existing.Status = v
		}
		if err := h.svc.UpdateRequestGroup(c.Request().Context(), existing); err != nil {
			return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
		}
		return c.JSON(http.StatusOK, existing.ToFHIR())
	case "GuidanceResponse":
		existing, _ := h.svc.GetGuidanceResponseByFHIRID(c.Request().Context(), fhirID)
		if v, ok := patched["status"].(string); ok {
			existing.Status = v
		}
		if err := h.svc.UpdateGuidanceResponse(c.Request().Context(), existing); err != nil {
			return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
		}
		return c.JSON(http.StatusOK, existing.ToFHIR())
	}
	return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome("unexpected error"))
}
